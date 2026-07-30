package guardian

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/evanstern/promptworld/internal/bundle"
	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/persona"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/skin"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/tool"
	"github.com/evanstern/promptworld/internal/toolloop"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// nudgeTextMax is the nudge rendering cap, read from the tool registry (spec
// 014 T021; re-pointed at send_vision when spec 029 retired nudge_dream):
// tool.Lookup("send_vision").Cost.TextCapBytes (400). It matches the sim
// reducer's NudgeTextMax enforcer — both derive from the same registry entry, so
// the guardian-side truncation and the door-side enforcement can never diverge.
var nudgeTextMax = func() int {
	t, _ := tool.Lookup("send_vision")
	return t.Cost.TextCapBytes
}()

// One console turn: player text in, charter-voiced reply out, at most one
// mediated nudge. The player's words have exactly one sink — the user turn
// of this prompt; villagers can only ever receive the model's `nudge.text`
// rendering, validated and landed through the InjectSocial door.

const playerTextMax = 2000

// ErrTurnBusy is returned while another console turn is in flight.
var ErrTurnBusy = errors.New("the guardian is attending another matter")

// TurnResult is the console-facing outcome of one turn.
type TurnResult struct {
	Reply     string        `json:"reply"`
	Nudge     *Nudge        `json:"nudge,omitempty"`
	Miracle   *Miracle      `json:"miracle,omitempty"`   // FROZEN JSON tag (spec 052 ruling 2): IPC clients decode it
	Order     *OrderReport  `json:"order,omitempty"`     // a placed standing order (spec 029 US2)
	Plan      *PlanReport   `json:"plan,omitempty"`      // a landed plan act (spec 084): placed/cancelled designation, issued/cancelled directive
	Region    *RegionReport `json:"region,omitempty"`    // a landed canonization (spec 101): christened region + optional feature
	Cancelled []string      `json:"cancelled,omitempty"` // released order ids (cancel_order)
	Clock     string        `json:"clock,omitempty"`     // a landed meta act's human line (spec 029 US5)
	Charges   int           `json:"charges"`
	Moments   []string      `json:"moments,omitempty"`
}

// OrderReport is the console-facing summary of a placed standing order (spec 029
// US2) — the id the player can name to cancel it, and its condition. Additive and
// omitempty; existing IPC clients ignore it.
type OrderReport struct {
	ID        string `json:"id"`
	Condition string `json:"condition"`
}

// Nudge reports a landed mediation.
type Nudge struct {
	Form    string   `json:"form"`
	Targets []string `json:"targets"`
	Text    string   `json:"text"`
}

// Miracle reports a landed miracle (spec 016) — the kind and a one-line human
// rendering. Never carries gratis: the guardian cannot work a free miracle.
type Miracle struct {
	Kind    string `json:"kind"`
	Summary string `json:"summary"`
}

// miracleArgs is the parsed work_miracle tool-call surface — the same flat field
// set the retired turnReply.Miracle struct carried (spec 016 turn contract). It
// has NO gratis field by design (FR-007/SC-005): the guardian can NEVER work a free
// miracle, and structural absence is the guarantee — landMiracle passes gratis
// false unconditionally, and there is nothing to forget to sanitize.
type miracleArgs struct {
	Kind     string `json:"kind"`
	Day      int    `json:"day"`
	Time     string `json:"time"`
	Villager string `json:"villager"`
	Item     string `json:"item"`
	Qty      int    `json:"qty"`
	Class    string `json:"class"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	ToX      int    `json:"to_x"`
	ToY      int    `json:"to_y"`
}

// turnOrigin distinguishes a console turn from a system-authored (triggered) turn
// (spec 029 US3, R6). Both run the SAME body (runTurn): same single-flight guard,
// same roster/handler/gate composition, same telemetry, same transcript append —
// only the framing differs. seed is the trailing directive: the player's words on
// the console path, the order's pre-authorized action instruction on the system
// path (which has NO player-text sink — the seed is the guardian's own recorded
// instruction). jobPrefix threads the correlation id ("turn" | "watch"); system
// marks the transcript with a [watch] origin and suppresses moment consumption.
type turnOrigin struct {
	system    bool
	jobPrefix string
	seed      string
	// survival marks a turn triggered by a system SURVIVAL watch (spec 059 US2):
	// its frame carries the survival-authority carve-out (visions/miracles on the
	// guardian's own initiative to save a life). Only a survival-watch trigger sets
	// it — a console turn and an ordinary (deferral) system turn both leave it
	// false and keep today's restrictive initiative frame verbatim (FR-004/FR-005).
	survival bool
	// angel marks a SCHEDULED cadence turn (spec 102, angel.go): system-framed
	// (no player-text sink), the incompetence-ceiling roster intersection (D3,
	// ceiling.go) and the angel initiative frame apply, and the directive is
	// the guardian's own standing cadence instruction. Only runAngel sets it.
	angel bool
	// jobID, when non-empty, overrides the minted correlation id so the
	// cog.thought the caller already opened (runAngel) and the loop's
	// cog.tool_call records share one chain (FR-008). Empty = mint as today.
	jobID string
}

// Turn runs one mediated CONSOLE turn through the bounded tool-use loop (spec 017
// T020, spec 029 T012). The model may reply with words (converse — the
// transcript-only final-answer channel, Result.Final) or call exactly one acting
// tool (send_vision / send_omen / monitor_and_act / cancel_order / work_miracle),
// which lands through its door. The driver's cardinality enforces "at most one
// mediated act per turn". Serialized: a second concurrent call fails fast with
// ErrTurnBusy (the console never waits — triggered system turns do, via
// runSystemTurn's bounded acquisition, T013).
func (mt *Guardian) Turn(ctx context.Context, playerText string) (TurnResult, error) {
	playerText = strings.TrimSpace(playerText)
	if playerText == "" {
		return TurnResult{}, errors.New("empty message")
	}
	if len(playerText) > playerTextMax {
		return TurnResult{}, fmt.Errorf("message exceeds %d characters", playerTextMax)
	}
	if !mt.turnBusy.CompareAndSwap(false, true) {
		return TurnResult{}, ErrTurnBusy
	}
	defer mt.turnBusy.Store(false)
	return mt.runTurn(ctx, turnOrigin{jobPrefix: "turn", seed: playerText})
}

// runTurn is the shared turn body for the console and system-authored paths (spec
// 029 T012/R6). The CALLER owns turnBusy (Turn CAS-fails fast; runSystemTurn waits
// bounded) — runTurn assumes it is held. It composes the instruction surface, the
// user prompt (framing the directive per origin), drives the loop, lands the
// cog.tool_call + retry telemetry on every termination path, and records the
// transcript. Moment consumption is CONSOLE-ONLY: a system turn never drains the
// player-facing moment queue (those await the next console open); the trigger
// worker queues the system turn's OWN moment from the returned outcome (T013).
func (mt *Guardian) runTurn(ctx context.Context, o turnOrigin) (TurnResult, error) {
	// The player-editable instruction surface, all read fresh this turn (FR-001):
	// the charter, the skill files composed beneath it, and (US2) the capability
	// manifest — each forked by the world's curriculum stage (spec 046 US2):
	// stage-1 locks the charter to the preset constant, skill files bind from
	// stage-3, and present-but-unbound files get the honest lock notice. Every
	// fallback/truncation/skip becomes a notice prefixed to the reply, one
	// combined line, exactly like the charter's today.
	charter, charterNotice := stageCharter(mt.worldDir, mt.stage, mt.charterPreset, mt.sk())
	// Charter-revision observation (spec 044 US2, T014): stamped at load,
	// before anything consumes the text, so the evidence timeline records
	// the revision this turn actually runs under — the stage-EFFECTIVE text
	// (at stage-1, the preset constant the lock serves), never the raw file.
	mt.observeCharter(charter)
	skills, skillNotices := stageSkills(mt.worldDir, mt.stage, mt.sk())
	// Skills observation (spec 077 FR-006): stamped at bind, before anything
	// consumes the set — the observeCharter twin, recording the stage-
	// EFFECTIVE bound set (empty at stages 1–2 by stageSkills' construction,
	// which never emits: absence is not an observation).
	mt.observeSkills(skills)
	grant, manifestNotices := loadManifest(mt.worldDir, bundleToolNames(mt.bundles)...)
	// The stage ceiling (spec 046 FR-004) intersects immediately after
	// loadManifest, BEFORE grantedRoster, so declaration/prose/door all inherit
	// it: a manifest may narrow within the ceiling, never exceed the stage.
	grant = applyStageCeiling(grant, mt.stage)
	// Persona bundles narrow the world-level grant (spec 036 US4 T030): a
	// bundle's own capabilities.json can only exclude tools/kinds the world
	// already grants, never add ones it excludes. Applied BEFORE grantedRoster
	// so the narrowing reaches every gating layer alike — the declared roster
	// (work_miracle's kind enum), the derived guidance, and (via `grant` on
	// turnDispatch below) the door itself.
	grant = narrowGrantForBundles(grant, mt.bundles)
	// The deliberate-incompetence ceiling (spec 102 D3, ceiling.go): a
	// SCHEDULED turn's grant narrows by the ceiling compiled from the
	// EFFECTIVE charter this turn runs under — default text caps initiative
	// to the modest read/counsel set; an authored charter lifts it (minus
	// the clock triple). Console and triggered turns skip this entirely:
	// compliance and tutor quality are full at any ceiling.
	angelLifted := false
	if o.angel {
		angelLifted = angelCharterLifted(charter, mt.charterPreset)
		grant = applyAngelCeiling(grant, angelLifted)
	}
	var notices []string
	if charterNotice != "" {
		notices = append(notices, charterNotice)
	}
	notices = append(notices, skillNotices...)
	notices = append(notices, manifestNotices...)
	// The ONE granted roster for this turn — the manifest-filtered guardian loop
	// roster (work_miracle's kind enum narrowed when restricted). It feeds all
	// three gating layers alike: Job.Roster (declaration), the derived guidance
	// (prose), and the handler set built from it (door), so an ungranted tool or
	// kind is structurally absent from every one of them (FR-005).
	roster := grantedRoster(grant)
	mt.stateMu.Lock()
	charges := mt.charges
	tick := mt.clockAt
	night := mt.night
	alive := make(map[int]bool, len(mt.alive))
	for k, v := range mt.alive {
		alive[k] = v
	}
	agentXY := make([][2]int, len(mt.agentXY))
	copy(agentXY, mt.agentXY)
	structXY := append([][2]int(nil), mt.structXY...)
	pileXY := append([][2]int(nil), mt.pileXY...)
	agentNeeds := append([]needMirror(nil), mt.agentNeeds...)
	moments := append([]string(nil), mt.moments...)
	story := append([]string(nil), mt.story...)
	orders := append([]sim.GuardianOrder(nil), mt.orders...)
	designations := append([]sim.Designation(nil), mt.designations...)
	directives := append([]sim.Directive(nil), mt.directives...)
	prophecies := append([]sim.Prophecy(nil), mt.prophecies...)
	faith := mt.faith
	// The guardian's working-memory window (spec 102 D1): mirrored per batch
	// from the shared selector; empty (and prompt-inert) on a non-agentized
	// world, so pre-102 prompts stay byte-identical (FR-007).
	memories := append([]string(nil), mt.memWin...)
	mt.stateMu.Unlock()

	// Bundle tools (spec 036 T014): merge the granted bundle surface into all
	// three gating layers — roster (declaration), handlers (door), and the
	// derived guidance (via turnSystemPrompt → tool.GuardianToolGuidance's
	// PromptGloss fallback, T012). Order is deterministic: built-ins first
	// (grantedRoster), then bundle tools in BundleSet load order, grant-filtered
	// the same way the world-level capabilities.json gates built-ins. The handler
	// factory needs a read snapshot of sim state (villager names/positions/
	// liveness) for target and recipient resolution — built from the same
	// absorb-mirrored positions landMiracle uses, so the turn worker never races
	// the replica the absorb goroutine owns.
	bundleHandlers := map[string]toolloop.Handler{}
	if mt.bundles != nil {
		probe := &sim.State{Agents: make([]sim.Agent, len(agentXY))}
		for i := range agentXY {
			probe.Agents[i] = sim.Agent{Name: sim.AgentNames[i], X: agentXY[i][0], Y: agentXY[i][1], Dead: !alive[i]}
		}
		// Spec 082: class+tile effect targets (structure@X,Y / pile@X,Y /
		// terrain@X,Y) resolve against this probe too — structure/pile tile
		// mirrors feed the compiler's one-per-tile presence probes and the
		// static map feeds its bounds check. Position-only; the reducer's dry
		// run stays the semantic authority for everything else.
		probe.Structures = make([]sim.Structure, len(structXY))
		for i, xy := range structXY {
			probe.Structures[i] = sim.Structure{X: xy[0], Y: xy[1]}
		}
		probe.Piles = make([]sim.Pile, len(pileXY))
		for i, xy := range pileXY {
			probe.Piles[i] = sim.Pile{X: xy[0], Y: xy[1]}
		}
		if mt.m != nil {
			probe.SetMap(mt.m)
		}
		// Invoker resolves into bundle effect TEMPLATES, which can land in
		// recorded payloads (memory text) — so it is fixed mechanics
		// vocabulary, deliberately NOT skin-resolved (spec 052 ruling 1:
		// the event log is skin-free), de-themed once at this version.
		ic := bundle.InvocationContext{State: probe, Tick: tick, Invoker: "the guardian", Inject: mt.social.InjectSocial, Seed: mt.seed}
		if mt.m != nil {
			ic.MapWidth, ic.MapHeight = mt.m.W, mt.m.H
		}
		for name, h := range mt.bundles.Handlers(ic) {
			if grant.allowsBundle(name) {
				bundleHandlers[name] = h
			}
		}
		for _, bt := range mt.bundles.Roster() {
			if grant.allowsBundle(bt.Name) {
				roster = append(roster, bt)
			}
		}
	}

	// Persona SOUL.md fragments (spec 036 US4 T029): each installed bundle's
	// identity text, in load order, appended after the charter section of the
	// system prompt (turnSystemPrompt below).
	var souls []string
	if mt.bundles != nil {
		souls = mt.bundles.SoulFragments()
	}
	// The skin's persona voice (spec 052 FR-004): one more editable-zone
	// fragment at the SAME seam, after the bundle SOULs — already validated
	// and capped at load (skin.Load, the bundle-SOUL discipline). The fixed
	// frame still lands LAST and unconditionally in turnSystemPrompt, so no
	// skin byte can displace it (spec 021 INV-1; the hostile-skin battery in
	// guardian_test.go proves it).
	if v := mt.sk().Voice(); v != "" {
		souls = append(souls, v)
	}

	// One correlation id per turn, mirroring mind's "<class>-<agent>-<tick>"
	// convention (telemetry.go newMeta): the console turn's class is "turn"; a
	// triggered system turn's is "watch" (R6). Threads every cog.tool_call.
	// The "-metatron-" correlation infix is FROZEN (spec 052 ruling 2): it
	// rides recorded cog.tool_call payloads and the TUI's decision-trace
	// attribution (tui/decisions.go) matches it verbatim.
	jobID := fmt.Sprintf("%s-metatron-%d", o.jobPrefix, tick)
	if o.jobID != "" {
		jobID = o.jobID // the caller opened the chain (runAngel) — share it
	}

	// The trailing directive: the player's words (console), the order's
	// pre-authorized action (system), or the guardian's own cadence
	// instruction (angel — spec 102, no player-text sink either way).
	directive := "The player says:\n" + o.seed
	switch {
	case o.angel:
		directive = o.seed
	case o.system:
		directive = "A standing order you placed has come due. Carry out its " +
			"pre-authorized action now, in a single act if it calls for one:\n" + o.seed
	}

	// The miracle targeting digest (spec 059 US3, FR-006): built only when this
	// world's granted roster offers work_miracle, so a dreams-only world's prompt
	// is byte-unchanged. It carries live villager positions/conditions + adjacent
	// passability so a coordinate-bearing miracle aims at a tile the door accepts.
	var digest string
	if hasWorkMiracle(roster) {
		digest = buildTargetingDigest(alive, agentNeeds, agentXY, mt.m)
	}

	result := TurnResult{}
	d := &turnDispatch{mt: mt, charges: charges, alive: alive, night: night, tick: tick, result: &result, grant: grant,
		// The explain scope (spec 063 US1): the SAME final roster this turn
		// declares (granted subset, bundle tools included, work_miracle's
		// kind enum already narrowed) plus the grant-independent catalog and
		// the stage id — fact sheets describe exactly this turn's world.
		scope: tool.ExplainScope{Granted: roster, Catalog: tool.LoopRosterGuardian(), Stage: mt.stage}}

	// Built-in handlers first, then the grant-filtered bundle handlers layered on
	// top (no name overlap survives boot — C1 skips a bundle tool that collides
	// with a built-in), so the merged map is the union both the roster and the
	// guidance already reflect.
	handlers := mt.turnHandlers(d)
	for name, h := range bundleHandlers {
		handlers[name] = h
	}

	// The tutor guide (spec 063 US3, standing resolution 2): compiled-in game
	// substrate, composed ONLY on tutor-preset worlds — keyed on the world's
	// charter preset, not its stage, so a tutor world keeps its guide as it
	// climbs. Every other world passes "" and composes byte-identically to
	// pre-063 (FR-004).
	guide := ""
	if mt.charterPreset == "tutor" {
		guide = persona.TutorGuide
	}

	// The initiative frame per origin (INV-1: always a compile-time constant
	// appended last): survival carve-out, the two angel frames (D3), or the
	// restrictive default.
	frame := guardianInitiativeFrame
	switch {
	case o.angel && angelLifted:
		frame = guardianAngelLiftedFrame
	case o.angel:
		frame = guardianAngelModestFrame
	case o.survival:
		frame = guardianSurvivalFrame
	}

	// A scheduled turn rides its OWN llm kind (spec 102 D2): the angel route,
	// estimator, and governor debt attribute to the cadence lane, never the
	// premium console kind.
	kind := llm.KindGuardian
	if o.angel {
		kind = llm.KindAngel
	}

	callCtx, cancel := context.WithTimeout(ctx, turnTimeout)
	res, err := mt.runLoop(callCtx, toolloop.Job{
		JobID:     jobID,
		Kind:      kind,
		System:    composeTurnSystemPrompt(frame, charter, guide, skills, roster, souls...),
		Seed:      turnUserPrompt(tick, charges, faith, alive, orders, designations, directives, prophecies, moments, story, memories, mt.soulTail(), mt.transcriptTail(), digest, directive),
		Roster:    roster,
		Handlers:  handlers,
		MaxRounds: mt.loopRounds,
		MaxTokens: mt.turnTokens, // llm.json max_tokens.metatron_turn (spec 025 US2), default 1024
		Record:    d.record,
	})
	cancel()

	// Land every buffered CallRecord as cog.tool_call telemetry (spec 017 FR-007,
	// T018), on EVERY termination path — a rejected / never-grounded call is
	// recorded even when nothing landed. A dedicated batch through the same
	// InjectSocial door as the nudge/miracle grounding events, so it neither
	// reorders nor entangles with them.
	mt.emitToolCalls(d.records, tick)

	// Transport retry visibility (spec 025 FR-004/SC-003): when the loop consumed
	// its one in-loop retry (recovered OR twice-failed), surface it as a
	// non-terminal cog.outcome through the same door — emitted BEFORE the
	// error-return below so a twice-failed turn's retry is still countable.
	if res.Retried {
		mt.emitRetried(jobID, tick, res.RetryReason)
	}

	if err != nil {
		// Honest unavailability; nothing landed (a landing returns a nil error),
		// nothing consumed, moments stay queued — exactly today's degraded path.
		// A system turn's caller (the trigger worker) maps this to an honest
		// model-free moment (T014).
		return TurnResult{}, err
	}

	// The reply is the model's closing/converse text (Result.Final). When the
	// model landed an act with no accompanying prose, Final may be empty — the
	// ⚡/✨/👁 report lines carry the turn. When NOTHING landed and nothing was
	// said, the loop ran dry (model_done with no text, cap exhaustion, or a soft
	// error) — the old scattered-thoughts fallback maps onto exactly these.
	reply := strings.TrimSpace(res.Final)
	if reply == "" && result.Nudge == nil && result.Miracle == nil && result.Order == nil &&
		result.Plan == nil && len(result.Cancelled) == 0 && result.Clock == "" {
		reply = "Forgive me — my thoughts scattered and I could not complete that. " +
			"Nothing was done and nothing was spent. Ask again."
	}
	result.Reply = reply
	if len(notices) > 0 {
		result.Reply = "(" + strings.Join(notices, "; ") + ")\n\n" + result.Reply
	}

	// Surfaced moments are consumed only on a completed CONSOLE turn — a system
	// (triggered) turn leaves the player-facing queue intact for the next console
	// open (R6), and reports no moments of its own here.
	if !o.system {
		mt.stateMu.Lock()
		result.Moments = moments
		mt.moments = mt.moments[len(moments):]
		result.Charges = mt.charges
		mt.stateMu.Unlock()
	} else {
		mt.stateMu.Lock()
		result.Charges = mt.charges
		mt.stateMu.Unlock()
	}

	mt.recordTurn(tick, o, result)
	return result, nil
}

// landVision validates a vision and lands it as one atomic batch (spec 029 US1,
// T005). A vision reaches exactly one living villager at ANY hour and costs one
// charge; target is the villager's name, text the model's rendering. The
// validation/batch/soul-append tail is UNCHANGED from the pre-029 landNudge (spec
// 029 T005: wrap, don't rewrite) — only the target resolution and the send_vision
// grant gate are vision-specific. Returns the landed nudge, or (nil, in-fiction
// refusal) which the handler maps to a rejected_gate the model may repair within
// the loop's round cap.
func (mt *Guardian) landVision(target, text string, reveal *placeReveal, charges int, alive map[int]bool, grant grantSet) (*Nudge, string) {
	if charges <= 0 {
		return nil, "no charges are banked"
	}
	// Capability gate (spec 021 R5.3, door layer): defense-in-depth behind the
	// handler-absence gate — a tool whose handler was never installed cannot reach
	// here, but the check keeps the door authoritative on its own.
	if !grant.allows("send_vision") {
		return nil, "that power is not granted in this world"
	}
	idx := agentIndexByName(target)
	if idx < 0 {
		return nil, fmt.Sprintf("no villager named %q", target)
	}
	if !alive[idx] {
		return nil, fmt.Sprintf("%s is beyond reach now", sim.AgentNames[idx])
	}
	// The optional place grant (spec 041 FR-014): the vision batch gains one
	// guardian.place_revealed plus its companion Origin-omen memory, riding
	// the SAME atomic batch — the grant lands with the vision or not at all.
	// Composition only, the BuildMiracleBatch contract: the sim reducer arm
	// (dry-run enforced) is the authority that the place is real and the
	// target alive; Seen/Detail are the arm's normative stamps, so the
	// emitter bakes just the place identity.
	var extra []store.Event
	if reveal != nil {
		extra = []store.Event{
			{Type: "guardian.place_revealed", Payload: mustJSON(sim.PlaceRevealedPayload{
				Agent: sim.Ref(idx), Facts: []sim.PlaceFact{{Kind: reveal.Kind, X: reveal.X, Y: reveal.Y,
					Provenance: sim.ProvenanceRevealed}}})},
			{Type: "agent.memory_added", Payload: mustJSON(sim.MemoryAddedPayload{
				Agent: sim.Ref(idx), Text: revealMemoryText(reveal), Salience: sim.SalDream,
				Subject: sim.Ref(-1), Origin: sim.OriginOmen})},
		}
	}
	return mt.landNudgeBatch("vision", []int{idx}, text, extra...)
}

// placeReveal is send_vision's optional divine place grant (spec 041 FR-014):
// one real place to upsert into the target's mental map, provenance revealed.
// Parsed from the place_kind / place_x / place_y argument triple (all or
// none — parseReveal refuses a partial triple).
type placeReveal struct {
	Kind string
	X, Y int
}

// revealMemoryText renders the place grant's fixed companion memory —
// deterministic, written for the villager's world (the miracle
// perception-memory shape; no player, no game, no outside voice).
func revealMemoryText(r *placeReveal) string {
	return fmt.Sprintf("The vision showed you the %s at (%d,%d).",
		strings.ReplaceAll(r.Kind, "_", " "), r.X, r.Y)
}

// landOmen validates an omen and either lands it now or defers it to nightfall
// (spec 029 US1/US4, T005/T016). An omen reaches one villager, a named group, or
// everyone living — at NIGHT only — for one charge regardless of recipient count.
// targetsArg is send_omen's comma-separated living-villager name list or the word
// "everyone".
//
// Night path: land immediately, spending a charge (the reducer's night gate is
// the door authority; the mirrored night flag is the turn-side pre-check).
//
// Day path (T016/R11): an omen belongs to the dark, so a daytime call does NOT
// refuse — it places a system-origin standing order that re-sends the omen the
// instant night falls (event_types ["sim.night_started"], TTL 1 game day,
// cap-exempt). Placement is FREE: the charge is spent at trigger-time landing,
// never here (FR-012/SC-004). Returns one of: (nudge, nil, "") landed at night;
// (nil, order, "") deferred to nightfall; (nil, nil, why) an in-fiction refusal.
func (mt *Guardian) landOmen(targetsArg, text string, charges int, night bool, tick int64, alive map[int]bool, grant grantSet) (*Nudge, *sim.GuardianOrder, string) {
	if !grant.allows("send_omen") {
		return nil, nil, "that power is not granted in this world"
	}
	targets, why := resolveOmenTargets(targetsArg, alive)
	if why != "" {
		return nil, nil, why
	}
	if strings.TrimSpace(text) == "" {
		return nil, nil, "the rendering was empty"
	}
	if !night {
		order, why := mt.deferOmen(targetsArg, targets, strings.TrimSpace(text), tick, grant)
		return nil, order, why
	}
	if charges <= 0 {
		return nil, nil, "no charges are banked"
	}
	nudge, why := mt.landNudgeBatch("omen", targets, text)
	return nudge, nil, why
}

// deferOmen places the daytime omen's nightfall deferral order (spec 029 T016/
// R11): a system-origin standing order whose one-shot trigger re-runs send_omen
// at night. The action is the seed the night SYSTEM turn reads (runTurn frames it
// as a due standing order), so it must lead the guardian to send_omen with these
// targets and this text; terse framing keeps it within the reducer's 400-rune
// action cap for all but the very longest renderings. "everyone" is preserved as
// the target word so the night turn re-resolves against whoever lives THEN; a
// named list re-sends to those still living. The charge is spent when the night
// turn lands, not here — placement is free and cap-exempt (origin "system"). A
// rejected placement maps to omen-appropriate counsel the model may repair.
func (mt *Guardian) deferOmen(targetsArg string, targets []int, text string, tick int64, grant grantSet) (*sim.GuardianOrder, string) {
	who := "everyone"
	if !strings.EqualFold(strings.TrimSpace(targetsArg), "everyone") {
		names := make([]string, len(targets))
		for i, t := range targets {
			names[i] = sim.AgentNames[t]
		}
		who = strings.Join(names, ", ")
	}
	a := orderArgs{
		Condition:  fmt.Sprintf("nightfall — an omen awaits %s", who),
		Action:     fmt.Sprintf("Night has fallen. Send the omen you promised to %s: %s", who, text),
		EventTypes: []string{"sim.night_started"},
		TTLDays:    1,
	}
	order, why := mt.placeOrder("system", a, tick, grant)
	if why != "" {
		return nil, "an omen belongs to the night, and I could not set it aside — " + why
	}
	return order, ""
}

// resolveOmenTargets parses send_omen's `targets` argument (spec 029 R3): a
// comma-separated list of living villager names, or the single word "everyone",
// into a deduplicated set of living villager indices. Every named villager must
// be alive — an unknown or dead name refuses the WHOLE act with counsel (one act,
// one charge, one atomic batch; never a partial omen). "everyone" resolves to the
// living set in index order.
func resolveOmenTargets(arg string, alive map[int]bool) ([]int, string) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return nil, `name the villagers the omen should reach, or say "everyone"`
	}
	if strings.EqualFold(arg, "everyone") {
		var targets []int
		for i := range sim.AgentNames {
			if alive[i] {
				targets = append(targets, i)
			}
		}
		if len(targets) == 0 {
			return nil, "no living villager remains to witness it"
		}
		return targets, ""
	}
	seen := map[int]bool{}
	var targets []int
	for _, part := range strings.Split(arg, ",") {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		idx := agentIndexByName(name)
		if idx < 0 {
			return nil, fmt.Sprintf("no villager named %q", name)
		}
		if !alive[idx] {
			return nil, fmt.Sprintf("%s is beyond reach now", sim.AgentNames[idx])
		}
		if !seen[idx] {
			seen[idx] = true
			targets = append(targets, idx)
		}
	}
	if len(targets) == 0 {
		return nil, "name at least one living villager for the omen"
	}
	return targets, ""
}

// landNudgeBatch is the shared landing tail for landVision / landOmen (spec 029
// T005): the text cap, the ONE atomic InjectSocial batch (guardian.nudged + one
// prefixed agent.memory_added per target at SalDream), and the soul append —
// VERBATIM the pre-029 landNudge body (wrap, don't rewrite). form fixes the memory
// prefix and the recorded form; the reducer dry-run is the door authority (charge
// spend, night gate for omen, living targets). extra events (spec 041: a vision's
// place grant + companion memory) ride the same atomic batch after the nudge
// memories. Returns the landed nudge, or (nil, refusal) the handler maps to a
// rejected_gate.
func (mt *Guardian) landNudgeBatch(form string, targets []int, text string, extra ...store.Event) (*Nudge, string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, "the rendering was empty"
	}
	if len(text) > nudgeTextMax {
		text = text[:nudgeTextMax]
	}
	// FROZEN recorded-at-emission text (spec 052 ruling 1 / FR-005): these
	// memory prefixes land in agent.memory_added payloads — the event log is
	// skin-free, so they use fixed mechanics vocabulary regardless of skin
	// and never change.
	prefix := "You saw a vision: "
	if form == "omen" {
		prefix = "You witnessed an omen: "
	}
	batch := []store.Event{{Type: "guardian.nudged", Payload: mustJSON(sim.GuardianNudgedPayload{
		Form: form, Targets: sim.Refs(targets), Text: text})}}
	for _, t := range targets {
		batch = append(batch, store.Event{Type: "agent.memory_added", Payload: mustJSON(sim.MemoryAddedPayload{
			Agent: sim.Ref(t), Text: prefix + text, Salience: sim.SalDream, Subject: sim.Ref(-1), Origin: sim.OriginOmen})})
	}
	batch = append(batch, extra...)
	if err := mt.social.InjectSocial(batch); err != nil {
		log.Printf("guardian: nudge rejected at the door: %v", err)
		return nil, "the world refused it (" + err.Error() + ")"
	}
	names := make([]string, len(targets))
	for i, t := range targets {
		names[i] = sim.AgentNames[t]
	}
	mt.appendFile(mt.soulPath(), fmt.Sprintf("\n- %s — I sent a %s to %s: %q\n",
		clock.Format(mt.replicaTickSafe()), form, strings.Join(names, ", "), text))
	// Agentized memory (spec 102 D5): the act enters the guardian's own store
	// too — fixed mechanics vocabulary (the event log is skin-free, ruling 1).
	mt.recordMemory(fmt.Sprintf("I sent a %s to %s: %q", form, strings.Join(names, ", "), text), salGuardianAct)
	return &Nudge{Form: form, Targets: names, Text: text}, ""
}

// landMiracle validates the model's miracle and lands it as one atomic batch
// through the same door and shared builder the operator console uses (spec 016
// R6), so the two channels cannot drift. The guardian can NEVER waive a charge:
// gratis is passed false unconditionally and does not exist on the turn contract
// (SC-005). Returns the landed miracle, or ("" is a silent skip) an in-fiction
// refusal reason the handler maps to a rejected_gate, exactly like landNudge.
// The validation, the atomic InjectSocial batch through the shared builder, and
// the soul append are UNCHANGED from the pre-loop turnReply path (spec 017 T020:
// wrap, don't rewrite) — only the input moved from a parsed JSON struct to the
// tool-call arguments (miracleArgs).
func (mt *Guardian) landMiracle(mm miracleArgs, charges int, grant grantSet) (*Miracle, string) {
	if charges <= 0 {
		return nil, "no charges are banked"
	}
	kind := strings.ToLower(strings.TrimSpace(mm.Kind))
	// Capability gate (spec 021 R5.3, door layer): work_miracle must be granted
	// and this kind offered by the world. Defense-in-depth behind handler-absence
	// (ungranted work_miracle installs no handler) and the declared kind enum
	// (ungranted kinds are never declared) — the door refuses in-fiction even if
	// a prompt-injected model conjures a call for an ungranted kind.
	if !grant.allows("work_miracle") || !grant.allowsKind(kind) {
		return nil, fmt.Sprintf("that %s is not granted in this world", mt.sk().WorkingNoun())
	}

	// Resolve the perception-memory recipients (which villager stands on a move's
	// source tile) from the absorb-mirrored positions, so the turn worker never
	// races the replica the absorb goroutine owns; the shared builder reads only
	// agent positions/liveness. Read BEFORE the kind switch (spec 091 FR-001): a
	// villager move naming a villager re-resolves its LIVE position from this
	// SAME mirror, so the door-side resolution below and the recipient lookup
	// BuildMiracleBatch performs later can never disagree.
	mt.stateMu.Lock()
	probe := &sim.State{Agents: make([]sim.Agent, len(mt.agentXY))}
	for i := range mt.agentXY {
		probe.Agents[i] = sim.Agent{X: mt.agentXY[i][0], Y: mt.agentXY[i][1], Dead: !mt.alive[i]}
	}
	mt.stateMu.Unlock()

	var params MiracleParams
	var summary string
	// nameResolved marks a villager move whose source came from live-position
	// resolution rather than the model's surveyed x/y (spec 091 decision (a)) —
	// used below to scope the raced-refusal message's name-preference suggestion
	// to the coordinate-only path it actually targets (FR-004).
	nameResolved := false
	switch kind {
	case "move":
		class := strings.ToLower(strings.TrimSpace(mm.Class))
		x, y := mm.X, mm.Y
		// Door-side name re-resolution (spec 091 decision (a), FR-001): a move
		// naming a living villager resolves that villager's LIVE position at
		// validation/emission time and moves it — the surveyed x/y become
		// advisory. Unknown or dead name refuses BEFORE the charge, mirroring
		// landVision's target resolution (agentIndexByName + alive check).
		// Coordinate-only villager moves (no name) and every structure/pile move
		// (class != "villager") take the untouched x/y path, byte-identical to
		// today (FR-003).
		if class == "villager" && strings.TrimSpace(mm.Villager) != "" {
			idx := agentIndexByName(mm.Villager)
			if idx < 0 {
				return nil, fmt.Sprintf("no villager named %q", mm.Villager)
			}
			if probe.Agents[idx].Dead {
				return nil, fmt.Sprintf("%s is beyond reach now", sim.AgentNames[idx])
			}
			x, y = probe.Agents[idx].X, probe.Agents[idx].Y
			nameResolved = true
		}
		params = MiracleParams{Class: class, X: x, Y: y, ToX: mm.ToX, ToY: mm.ToY}
		summary = fmt.Sprintf("moved the %s at (%d,%d) to (%d,%d)", params.Class, x, y, mm.ToX, mm.ToY)
	case "remove":
		params = MiracleParams{Class: strings.ToLower(strings.TrimSpace(mm.Class)), X: mm.X, Y: mm.Y}
		summary = fmt.Sprintf("removed the %s at (%d,%d)", params.Class, mm.X, mm.Y)
	case "time_snap":
		hour, min, perr := clock.ParseTimeOfDay(mm.Time)
		if perr != nil {
			return nil, perr.Error()
		}
		params = MiracleParams{ToTick: clock.TickAt(int64(mm.Day), hour, min, 0)}
		summary = fmt.Sprintf("snapped time forward to day %d %02d:%02d", mm.Day, hour, min)
	case "give_item":
		idx := agentIndexByName(mm.Villager)
		if idx < 0 {
			return nil, fmt.Sprintf("no villager named %q", mm.Villager)
		}
		item := strings.ToLower(strings.TrimSpace(mm.Item))
		params = MiracleParams{Agent: idx, Item: item, Qty: mm.Qty}
		summary = fmt.Sprintf("granted %d %s to %s", mm.Qty, item, sim.AgentNames[idx])
	default:
		return nil, fmt.Sprintf("unknown %s %q", mt.sk().WorkingNoun(), mm.Kind)
	}

	batch, err := BuildMiracleBatch(probe, kind, params, false)
	if err != nil {
		return nil, err.Error()
	}
	if err := mt.social.InjectSocial(batch); err != nil {
		log.Printf("guardian: working rejected at the door: %v", err)
		msg := "the world refused it (" + err.Error() + ")"
		// FR-004: a coordinate-only villager move that races (the villager left
		// (x,y) between survey and this call — the reducer's "no living
		// villager at" refusal) suggests the name-addressed form, which the
		// door resolves live and so cannot race. Scoped to that exact refusal
		// so a bad destination or an already name-resolved move never gets the
		// suggestion appended.
		if kind == "move" && params.Class == "villager" && !nameResolved &&
			strings.Contains(err.Error(), "no living villager at") {
			msg += " — name the villager instead of coordinates and the door will resolve their live position"
		}
		return nil, msg
	}
	// Fresh soul appends use the skin vocabulary (spec 052 edge case: old
	// transcripts/soul files are history, never rewritten; only fresh
	// appends re-voice).
	mt.appendFile(mt.soulPath(), fmt.Sprintf("\n- %s — I worked a %s: %s\n",
		clock.Format(mt.replicaTickSafe()), mt.sk().WorkingNoun(), summary))
	// Agentized memory (spec 102 D5): fixed vocabulary — a recorded payload
	// never carries skin bytes (ruling 1), so the default "working" stands.
	mt.recordMemory("I worked a working: "+summary, salGuardianAct)
	return &Miracle{Kind: kind, Summary: summary}, ""
}

// recordTurn appends the exchange to the transcript. A console turn opens with
// the player's line ("> …"); a system-authored turn opens with a "[watch]" origin
// marker over the order's pre-authorized action (spec 029 T012/R6) — never a
// player-text line, because a triggered turn has no player text.
func (mt *Guardian) recordTurn(tick int64, o turnOrigin, r TurnResult) {
	// Fresh transcript appends use the skin vocabulary (spec 052 FR-007);
	// already-written lines are history and never rewritten.
	var b strings.Builder
	if o.system {
		// A survival-watch turn is marked distinctly (spec 059 FR-007): the
		// durable transcript attributes the acting authority to the survival duty,
		// not an ordinary player-placed watch. A scheduled cadence turn (spec
		// 102) marks its own origin the same way.
		marker := "[watch]"
		if o.survival {
			marker = "[survival watch]"
		}
		if o.angel {
			marker = "[cadence]"
		}
		fmt.Fprintf(&b, "\n[%s] %s\n%s\n\n%s: %s\n", clock.Format(tick), marker, o.seed, mt.sk().Epithet(), r.Reply)
	} else {
		fmt.Fprintf(&b, "\n[%s]\n> %s\n\n%s: %s\n", clock.Format(tick), o.seed, mt.sk().Epithet(), r.Reply)
	}
	if r.Nudge != nil {
		fmt.Fprintf(&b, "⚡ %s → %s: %q\n", mt.sk().FormNoun(r.Nudge.Form), strings.Join(r.Nudge.Targets, ", "), r.Nudge.Text)
	}
	if r.Miracle != nil {
		fmt.Fprintf(&b, "✨ %s: %s\n", mt.sk().WorkingNoun(), r.Miracle.Summary)
	}
	if r.Order != nil {
		fmt.Fprintf(&b, "👁 watch set (%s): %q\n", r.Order.ID, r.Order.Condition)
	}
	for _, id := range r.Cancelled {
		fmt.Fprintf(&b, "👁 watch released: %s\n", id)
	}
	if r.Clock != "" {
		fmt.Fprintf(&b, "⏲ %s\n", r.Clock)
	}
	mt.appendFile(mt.transcriptPath(), b.String())
	// Agentized memory (spec 102 D5, "player exchanges"): a CONSOLE exchange
	// enters the store as the guardian's OWN words — never the player's text,
	// whose one sink stays the prompt (the persona-firewall non-negotiable).
	if !o.system && r.Reply != "" {
		reply := r.Reply
		if ridx := strings.IndexByte(reply, '\n'); ridx > 0 {
			reply = reply[:ridx] // first line — the memory is a note, not a transcript
		}
		mt.recordMemory("The player sought me; I said: "+reply, salGuardianTalk)
	}
}

// Status is the model-free peek: charges, charter provenance, soul tail, and
// (spec 021 R8, US3) the instruction-file + capability provenance a player reads
// to answer "what is my guardian running on, and what can it do". The new fields
// are additive and omitempty where sensible, so existing IPC clients ignore them
// (encoding/json) — no protocol version bump (contracts/status.md).
type Status struct {
	Charges         int           `json:"charges"`
	CharterDefault  bool          `json:"charter_default"`
	SoulTail        string        `json:"soul_tail"`
	Skills          []string      `json:"skills,omitempty"`        // effective skill filenames, composition order
	GrantedTools    []string      `json:"granted_tools,omitempty"` // granted roster, registry order; work_miracle(kind,…) when restricted
	ManifestDefault bool          `json:"manifest_default"`        // true ⇒ no capabilities.json (full default grant)
	Orders          []OrderStatus `json:"orders,omitempty"`        // standing orders (spec 029 US2, FR-016) — active + recent
	// Curriculum-ladder provenance (spec 046 US2) — additive omitempty, the
	// spec-021 precedent. Stage is the world's immutable stage id; absent = a
	// pre-ladder, ungated world. CharterLocked reports the stage-1 instruction
	// lock: the effective charter is the CharterPreset constant and charter.md
	// does not bind. SkillsLocked reports that skill files do not compose at
	// this stage (stage-1/-2 — they bind from stage-3).
	Stage         string `json:"stage,omitempty"`
	CharterLocked bool   `json:"charter_locked,omitempty"`
	CharterPreset string `json:"charter_preset,omitempty"` // the binding preset name when locked ("default" | "tutor")
	SkillsLocked  bool   `json:"skills_locked,omitempty"`
	// Resolved skin display facts (spec 052 FR-012, contract §7) — additive
	// omitempty, the spec-021 precedent, so clients render skin vocabulary
	// without reading world files. The identity fields are always sent
	// (resolved against the default table); SkinStrings/SkinStages carry only
	// the world's overrides. Absent fields (a pre-052 daemon) mean the
	// default Guardian skin.
	SkinName        string                        `json:"skin_name,omitempty"`
	SkinEpithet     string                        `json:"skin_epithet,omitempty"`
	SkinTabLabel    string                        `json:"skin_tab_label,omitempty"`
	SkinFamilyLabel string                        `json:"skin_family_label,omitempty"`
	SkinStrings     map[string]string             `json:"skin_strings,omitempty"`
	SkinStages      map[string]skin.StageIdentity `json:"skin_stages,omitempty"`
}

// OrderStatus is the model-free peek at one standing order (spec 029 US2/US3,
// data-model §6): what the player reads to answer "what watches stand, and how
// long". Additive and omitempty — existing IPC clients ignore it (the spec-021
// precedent).
type OrderStatus struct {
	ID         string `json:"id"`
	Condition  string `json:"condition"`
	Origin     string `json:"origin"`
	Fuzzy      bool   `json:"fuzzy,omitempty"`
	ExpiresDay int64  `json:"expires_day"`
	Status     string `json:"status"`
}

// Status is computed fresh per call from disk (same per-read discipline as the
// turn, FR-001): a skill added or a manifest edited between peeks shows on the
// next read with no cache to go stale.
func (mt *Guardian) Status() Status {
	mt.stateMu.Lock()
	c := mt.charges
	orders := mt.orderStatuses()
	mt.stateMu.Unlock()
	// The status twin of the turn's load site (spec 046 US2): the stage ceiling
	// intersects here too, so the peeked grant can never disagree with the
	// roster the next turn will run under.
	grant, _ := loadManifest(mt.worldDir)
	grant = applyStageCeiling(grant, mt.stage)
	skills := skillNames(mt.worldDir)
	if !stageBindsSkills(mt.stage) {
		// Skills is the EFFECTIVE composition list; below stage-3 nothing
		// composes — SkillsLocked carries the provenance instead.
		skills = nil
	}
	s := Status{
		Charges:         c,
		CharterDefault:  charterIsDefault(mt.worldDir, mt.charterPreset),
		SoulTail:        mt.soulTail(),
		Skills:          skills,
		GrantedTools:    grantedToolLabels(grant),
		ManifestDefault: grant.manifestDefault,
		Orders:          orders,
		Stage:           mt.stage,
		CharterLocked:   stageLocksCharter(mt.stage),
		SkillsLocked:    mt.stage != "" && !stageBindsSkills(mt.stage),
		// Skin display facts (spec 052 contract §7): identity fields resolved
		// (never empty), override maps only when a world skin overrides.
		SkinName:        mt.sk().Name(),
		SkinEpithet:     mt.sk().Epithet(),
		SkinTabLabel:    mt.sk().TabLabel(),
		SkinFamilyLabel: mt.sk().FamilyLabel(),
		SkinStrings:     mt.sk().StringOverrides(),
		SkinStages:      mt.sk().StageOverrides(),
	}
	if s.CharterLocked {
		s.CharterPreset = mt.charterPreset
		if s.CharterPreset == "" {
			s.CharterPreset = "default"
		}
	}
	return s
}

// orderStatuses projects the mirrored standing orders into the model-free status
// surface (spec 029 FR-016). Caller holds stateMu. nil when no orders stand (the
// field omits under omitempty — byte-compatible with pre-029 status).
func (mt *Guardian) orderStatuses() []OrderStatus {
	if len(mt.orders) == 0 {
		return nil
	}
	out := make([]OrderStatus, 0, len(mt.orders))
	for i := range mt.orders {
		o := mt.orders[i]
		out = append(out, OrderStatus{
			ID:         o.ID,
			Condition:  o.Condition,
			Origin:     o.Origin,
			Fuzzy:      o.Confirm,
			ExpiresDay: o.ExpiresTick / ticksPerGameDay,
			Status:     o.Status,
		})
	}
	return out
}

func (mt *Guardian) soulTail() string { return tailOfFile(mt.soulPath(), soulTailBytes) }
func (mt *Guardian) transcriptTail() string {
	t := tailOfFile(mt.transcriptPath(), 3000)
	// Trim to whole turns, newest-last.
	turns := strings.Split(t, "\n[")
	if len(turns) > transcriptTailTurns {
		turns = turns[len(turns)-transcriptTailTurns:]
	}
	return strings.Join(turns, "\n[")
}

func (mt *Guardian) replicaTickSafe() int64 {
	mt.stateMu.Lock()
	defer mt.stateMu.Unlock()
	return mt.clockAt
}

// bundleToolNames returns the boot-frozen bundle roster's tool names (nil for
// a nil/empty BundleSet) — the "known bundle tool" set loadManifest uses to
// suppress a cosmetic "unknown tool … ignored" notice for a world-level
// capabilities.json that correctly names a real bundle tool (spec 036 T030
// handoff fix): allowsBundle already grants such a name; the notice was noise.
func bundleToolNames(bs *bundle.BundleSet) []string {
	if bs == nil {
		return nil
	}
	roster := bs.Roster()
	if len(roster) == 0 {
		return nil
	}
	names := make([]string, len(roster))
	for i, t := range roster {
		names[i] = t.Name
	}
	return names
}

func agentIndexByName(name string) int {
	name = strings.ToLower(strings.TrimSpace(name))
	for i, n := range sim.AgentNames {
		if strings.ToLower(n) == name {
			return i
		}
	}
	return -1
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// guardianNonNegotiables carries the two persona-firewall invariants VERBATIM
// (spec 021 FR-003): the guardian never invents unobserved events, and the
// player's literal words never pass to a villager. It is a compile-time
// constant appended after ALL editable content on every path (INV-1), so no
// charter or skill byte can displace, truncate, or override it — the wording is
// unchanged from the pre-021 fixed frame, and door-side enforcement backs it
// independently of this prompt text.
const guardianNonNegotiables = `Whatever voice or policy the charter above gives you, two things are fixed:
you never invent events, actions, or words that are not in your notes or the
status you are given — when you have not observed something, say so in your
own way — and you never let the player's literal words pass to a villager.`

// guardianInitiativeFrame pins meta control and standing orders to player
// authority (spec 029 T019, contracts/tools.md): the clock-control tools and any
// standing order may be used ONLY when the player asks for them, or when a
// standing order the player already placed authorizes the act — never on the
// guardian's own initiative. Like guardianNonNegotiables it is a compile-time
// constant appended last (INV-1), so no charter or skill byte can override it,
// and it appears on every path (the door-side grant gate backs it independently).
const guardianInitiativeFrame = `Two more powers are the player's to command, never yours to take up alone: the ` +
	`world's clock (pausing, starting, changing its pace) and standing orders (watches you keep and act on). ` +
	`Use them only when the player asks, or when a standing order the player placed tells you to act — never on your own initiative.`

// guardianSurvivalFrame is the initiative frame for a SURVIVAL-watch turn (spec
// 059 US2, FR-003/FR-004): the ONE carve-out. A villager is at the brink of
// death, so for THIS peril the guardian may send a vision or work a miracle on its
// own initiative — no player authorization needed (charges unchanged). Everything
// else stays exactly as the restrictive frame: the clock and any other standing
// order remain the player's alone (FR-004), so the carve-out cannot leak into
// non-survival powers. Like guardianInitiativeFrame it is a compile-time constant
// appended last (INV-1) beneath all editable content.
const guardianSurvivalFrame = `A villager stands at the brink of death, and you keep the survival watch — your ` +
	`own nature, not the player's command. For THIS peril alone you may act on your own initiative: send a vision ` +
	`or work a working (the work_miracle tool) to save a life, without waiting for the player to ask (a life-saving ` +
	`act still spends a banked charge — if none is banked you can only counsel and keep the watch). This authority ` +
	`is survival's alone. The world's clock (pausing, starting, changing its pace) and every other standing order ` +
	`remain the player's to command, never yours to take up alone — use them only when the player asks.`

// hasWorkMiracle reports whether the granted roster offers the miracle tool —
// gates the miracle-specific doctrine line so a dreams-only world never mentions
// miracles (FR-005).
func hasWorkMiracle(roster []tool.Tool) bool {
	for _, t := range roster {
		if t.Name == "work_miracle" {
			return true
		}
	}
	return false
}

// turnSystemPrompt composes the guardian turn's system prompt (spec 021 R3,
// data-model.md §2): the editable charter, then (spec 036 US4 T029) each
// installed persona bundle's SOUL.md fragment under a `--- persona ---`
// separator in bundle load order, then each skill file under a
// `--- skill: <name> ---` separator in composition order, then the fixed frame
// appended LAST and unconditionally. The frame carries the two non-negotiables
// verbatim, the tool-agnostic acting doctrine, and — for THIS world's granted
// roster — the registry-derived tool guidance (tool.GuardianToolGuidance),
// which replaces the old hand-written tool list so the described surface can
// never diverge from the declared one (FR-008) and automatically reflects the
// granted subset (a conversation-only world names no acting tools at all).
//
// souls is variadic so every pre-036 call site keeps compiling unchanged
// (extend, don't break) — an absent argument composes byte-identically to the
// pre-036 prompt.
func turnSystemPrompt(charter string, skills []skillFile, roster []tool.Tool, souls ...string) string {
	return buildTurnSystemPrompt(false, charter, "", skills, roster, souls...)
}

// buildTurnSystemPrompt is turnSystemPrompt with the survival-frame selector
// (spec 059 US2, T005) and the tutor-guide slot (spec 063 US3). survival=false
// composes byte-identically to the pre-059 prompt (guardianInitiativeFrame
// verbatim, FR-005); survival=true swaps ONLY the initiative frame for
// guardianSurvivalFrame — the non-negotiables, the roster guidance, and every
// other byte are unchanged, so the carve-out is exactly the initiative
// doctrine and nothing more. runTurn passes the turn's survival origin; the
// pre-059 wrapper above pins false for every existing call site (extend,
// don't break).
//
// guide is the compiled-in tutor guide (persona.TutorGuide) on tutor-preset
// worlds, "" everywhere else (spec 063 FR-004): it composes in the EDITABLE
// zone — after the charter, the persona SOULs, and the skin voice, before the
// skill files — so the fixed frame still lands LAST and unconditionally on
// every path (spec 021 INV-1); an empty guide composes byte-identically to
// pre-063 (the non-tutor byte-identity guarantee, SC-003).
func buildTurnSystemPrompt(survival bool, charter, guide string, skills []skillFile, roster []tool.Tool, souls ...string) string {
	initiative := guardianInitiativeFrame
	if survival {
		initiative = guardianSurvivalFrame
	}
	return composeTurnSystemPrompt(initiative, charter, guide, skills, roster, souls...)
}

// composeTurnSystemPrompt is the frame-parametric composer beneath
// buildTurnSystemPrompt (spec 102): the initiative frame became a parameter
// so the two angel frames (ceiling.go) compose through the SAME body — every
// other byte, and the frame-lands-LAST invariant (spec 021 INV-1), unchanged.
func composeTurnSystemPrompt(initiative, charter, guide string, skills []skillFile, roster []tool.Tool, souls ...string) string {
	var b strings.Builder
	b.WriteString(charter)
	for _, s := range souls {
		fmt.Fprintf(&b, "\n\n--- persona ---\n%s", s)
	}
	if guide != "" {
		fmt.Fprintf(&b, "\n\n--- guide (game-authored) ---\n%s", guide)
	}
	for _, s := range skills {
		fmt.Fprintf(&b, "\n\n--- skill: %s ---\n%s", s.name, s.text)
	}
	fmt.Fprintf(&b, "\n\n--- (fixed frame, beneath the charter and skills) ---\n"+
		"You are the intermediary between the player and the village of eight: %s.\n%s\n%s\n\n",
		strings.Join(sim.AgentNames[:], ", "), guardianNonNegotiables, initiative)

	guidance := tool.GuardianToolGuidance(roster)
	// The read-tool paragraph (spec 063 US1/US2): explain's free-to-read
	// doctrine, rendered from the granted roster like the acting guidance —
	// empty (and byte-inert) when no read tool is granted.
	readGuidance := tool.GuardianReadGuidance(roster)
	if guidance == "" {
		// Conversation-only world: no acting tools granted (FR-006). The guardian
		// still converses — speech is never gateable.
		b.WriteString("This world grants you no acting tools — you may only counsel the " +
			"player in words. To SPEAK to the player, simply reply with your words; " +
			"that reply is what the player reads, and speaking costs nothing.")
		if readGuidance != "" {
			b.WriteString("\n\n" + readGuidance)
		}
		return b.String()
	}

	b.WriteString("When you choose to act on the player's behalf, judge first: the target's " +
		"persuadability, the impact on the village, and the right method. Acting spends one of " +
		"your banked charges — if none are banked, or the request is unwise, refuse and counsel " +
		"instead (refusal is free). Act at most ONCE per turn: the first act you land is the whole " +
		"of this turn. Any text a villager receives must be written for the villager's world: no " +
		"player, no game, no outside voice.\n\n")
	if hasWorkMiracle(roster) {
		// Fixed-frame doctrine (never skinnable — the working noun here is
		// the compile-time default, not a token: no skin byte reaches the
		// frame). work_miracle is the frozen tool id the model calls.
		b.WriteString("You cannot work a working (the work_miracle tool) for free, and you can never remove a villager.\n\n")
	}
	// The read paragraph composes BEFORE the acting block, so the acting
	// doctrine's closing sentence stays the prompt's last byte — the
	// adversarial battery pins that tail (fixedFrameLast), and the frame
	// staying last is the invariant, not just the wording.
	if readGuidance != "" {
		b.WriteString(readGuidance + "\n")
	}
	b.WriteString("To SPEAK to the player, simply reply with your words — that reply is what the " +
		"player reads, and speaking costs nothing. To ACT on the player's behalf, call exactly ONE " +
		"of these tools (and only one — one mediated act per turn):\n")
	b.WriteString(guidance)
	b.WriteString("If none are banked, or the request is unwise, do NOT call a tool — counsel the " +
		"player in words instead (refusal is free). Never call more than one tool: the first act " +
		"you land is the whole of this turn.")
	return b.String()
}

// targetingDigestMaxBytes hard-bounds the miracle targeting digest (spec 059 US3,
// FR-006): the digest rides the turn prompt's budget discipline, so it is capped
// and truncated rather than allowed to grow with a large or chatty village. Eight
// villagers × one terse line each sits well under this; the cap is the guard, not
// the common case.
const targetingDigestMaxBytes = 1600

// buildTargetingDigest assembles the miracle targeting digest (spec 059 US3): each
// LIVING villager's name, tile, and survival-relevant condition, plus the passable
// tiles adjacent to them, so a coordinate-bearing miracle can aim at a tile the
// landing door will accept (the world-01 "3 of 4 door-rejected" fix). Positions
// and conditions come from the absorb-mirrored snapshots (never the replica the
// absorb goroutine owns); passability is the static map's own Passable — the door
// stays the authority, this is guidance. Token-bounded: villagers are naturally
// capped at the roster size, adjacency at the four cardinals, and the whole block
// is truncated at targetingDigestMaxBytes. Empty when no one lives.
//
// Carry headroom (spec 095 FR-001) rides the SAME per-villager line, from the
// SAME needMirror snapshot as health/food/warmth: give_item's carry-cap door
// (internal/sim/miracles.go applyItemGranted, FR-011) rejects an over-cap grant
// WHOLE, never clamps — so naming each living villager's free bulk against
// sim.BulkCap here lets the model pick a quantity that fits BEFORE it calls the
// tool, instead of bouncing off the door and retrying. Dead villagers carry no
// line at all (the existing skip above), so this adds nothing for them.
func buildTargetingDigest(alive map[int]bool, needs []needMirror, xy [][2]int, m *worldmap.Map) string {
	var b strings.Builder
	b.WriteString(tool.GuardianTargetingGuidance())
	b.WriteString("\n")
	any := false
	for i, name := range sim.AgentNames {
		if !alive[i] || i >= len(xy) || i >= len(needs) {
			continue
		}
		any = true
		x, y := xy[i][0], xy[i][1]
		n := needs[i]
		fmt.Fprintf(&b, "- %s at (%d,%d) — health %d/1000, food %d/1000, warmth %d/1000%s",
			name, x, y, n.Health, n.Food, n.Warmth, survivalFlag(n))
		free := sim.BulkCap - n.Bulk
		if free < 0 {
			free = 0 // defensive floor (freeBulk's own doctrine): never a negative
		}
		fmt.Fprintf(&b, "; carrying %d/%d, %d free to receive", n.Bulk, sim.BulkCap, free)
		if m != nil {
			var tiles []string
			for _, d := range [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
				nx, ny := x+d[0], y+d[1]
				if m.Passable(nx, ny) {
					tiles = append(tiles, fmt.Sprintf("(%d,%d)", nx, ny))
				}
			}
			if len(tiles) > 0 {
				fmt.Fprintf(&b, "; passable next tiles: %s", strings.Join(tiles, " "))
			}
		}
		b.WriteString("\n")
	}
	if !any {
		return ""
	}
	out := b.String()
	if len(out) > targetingDigestMaxBytes {
		out = out[:targetingDigestMaxBytes]
	}
	return out
}

// survivalFlag annotates a villager in the targeting digest with the danger band
// it currently sits in (spec 059 US3), keyed on the SAME survival bands the
// watches match — so the digest the guardian reads names exactly the peril that
// woke the survival watch, no drift.
func survivalFlag(n needMirror) string {
	var flags []string
	if inBand, _ := survivalBand(sim.SurvivalNearDeath, n); inBand {
		flags = append(flags, "near death")
	}
	if inBand, _ := survivalBand(sim.SurvivalStarvation, n); inBand {
		flags = append(flags, "starving")
	}
	if inBand, _ := survivalBand(sim.SurvivalExposure, n); inBand {
		flags = append(flags, "freezing")
	}
	if len(flags) == 0 {
		return ""
	}
	return " (" + strings.Join(flags, ", ") + ")"
}

// turnUserPrompt composes the turn's user prompt. The trailing `directive` is the
// ALREADY-FRAMED directive block runTurn authored per origin — the console's "The
// player says:\n…" or the system turn's standing-order framing — and is appended
// verbatim. runTurn is the sole author of the origin-appropriate label: the label
// lives in exactly one place, so a console turn carries it once and a system turn
// never pretends its directive came from the player this turn (spec 029 R6).
func turnUserPrompt(tick int64, charges, faith int, alive map[int]bool, orders []sim.GuardianOrder, designations []sim.Designation, directives []sim.Directive, prophecies []sim.Prophecy, moments, story, memories []string, soulTail, transcriptTail, digest, directive string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "World clock: %s. Charges banked: %d of %d.\n", clock.Format(tick), charges, sim.GuardianChargeCap)
	// The village's faith (spec 085 FR-013): in-fiction wording, always
	// present — the guardian's counsel must know how the flock stands, and
	// the score steers its own charge regeneration.
	fmt.Fprintf(&b, "The village's faith in you stands at %d of 100 — %s. Faith rises when your directives are fulfilled and your prophecies come true; it falls with deaths, lapsed directives, and false prophecies, and it sets how quickly your charges return.\n",
		faith, faithBandWord(faith))
	var dead []string
	for i, n := range sim.AgentNames {
		if !alive[i] {
			dead = append(dead, n)
		}
	}
	if len(dead) > 0 {
		fmt.Fprintf(&b, "Departed: %s.\n", strings.Join(dead, ", "))
	}
	// Standing orders (spec 029 FR-017): the guardian's counsel and confirmations
	// stay truthful to live state only if it carries its active watches — id (so
	// the player can name one to cancel), condition, remaining game-days, and
	// whether the order is fuzzy (needs a confirm) or purely structural.
	writeStandingOrders(&b, tick, orders)
	// The plan layer (spec 084 FR-015): active designations and directives —
	// id, kind, site, days-left — so the guardian's counsel stays truthful to
	// live state (the writeStandingOrders discipline). Empty when none are
	// active, so a plan-free world's prompt is byte-unchanged.
	writeDesignations(&b, designations)
	writeDirectives(&b, tick, directives)
	// Active prophecies (spec 085 FR-013): id, claim, deadline — the
	// writeStandingOrders shape; empty when none stand, so a prophecy-free
	// world's prompt gains only the faith line above.
	writeProphecies(&b, tick, prophecies)
	if len(moments) > 0 {
		b.WriteString("\nMoments you have not yet reported (lead with these):\n")
		for _, m := range moments {
			b.WriteString("- " + m + "\n")
		}
	}
	if len(story) > 0 {
		b.WriteString("\nThe village chronicle (recent entries):\n")
		for _, s := range story {
			b.WriteString("- " + s + "\n")
		}
	}
	// The working-memory window (spec 102 D1): the agentized guardian's own
	// remembered moments, selected by the shared deterministic top-K window.
	// Empty on a non-agentized world, so pre-102 prompts are byte-unchanged.
	if len(memories) > 0 {
		b.WriteString("\nYour remembered moments (your own store, most recent first):\n")
		for _, m := range memories {
			b.WriteString("- " + m + "\n")
		}
	}
	if soulTail != "" {
		b.WriteString("\nYour recent notes:\n" + soulTail + "\n")
	}
	if transcriptTail != "" {
		b.WriteString("\nRecent conversation:\n" + transcriptTail + "\n")
	}
	// The miracle targeting digest (spec 059 US3): only present on miracle-capable
	// turns (runTurn passes "" otherwise), so a dreams-only world's prompt is
	// byte-unchanged.
	if digest != "" {
		b.WriteString("\n" + digest)
	}
	fmt.Fprintf(&b, "\n%s\n", directive)
	return b.String()
}

// writeStandingOrders renders the active-order block of the turn user prompt (spec
// 029 T010, FR-017). Only ACTIVE orders show (consumed ones are history the status
// surface carries); remaining days floor at 0. A fuzzy order (Confirm) is marked
// so the guardian sets honest expectations about the confirm step.
// writeDesignations renders the active-designation block of the turn user
// prompt (spec 084 FR-015): id (so the player can name one to cancel or bind a
// directive to), kind, site, and the guardian's own label. Only ACTIVE
// designations show — consumed ones are history the status trail carries.
func writeDesignations(b *strings.Builder, designations []sim.Designation) {
	var active []sim.Designation
	for _, d := range designations {
		if d.Status == "active" {
			active = append(active, d)
		}
	}
	if len(active) == 0 {
		return
	}
	b.WriteString("\nDesignations you have marked on the world:\n")
	for i := range active {
		d := &active[i]
		line := fmt.Sprintf("- %s: %s at %s", d.ID, d.Kind, describeDesignationSite(d))
		if d.StructureKind != "" {
			line += " (" + d.StructureKind + ")"
		}
		if d.Kind == sim.DesignationSettlementZone {
			line += fmt.Sprintf(" (%d structures to settle it)", d.MinStructures)
		}
		if d.Label != "" {
			line += fmt.Sprintf(" — %q", d.Label)
		}
		b.WriteString(line + "\n")
	}
}

// writeDirectives renders the active-directive block (spec 084 FR-015): id,
// the bound designation, who is charged, and remaining game-days (floored at
// 0, the writeStandingOrders shape).
func writeDirectives(b *strings.Builder, tick int64, directives []sim.Directive) {
	var active []sim.Directive
	for _, d := range directives {
		if d.Status == "active" {
			active = append(active, d)
		}
	}
	if len(active) == 0 {
		return
	}
	b.WriteString("\nDirectives you have laid on the village:\n")
	for i := range active {
		d := &active[i]
		days := (d.ExpiresTick - tick) / ticksPerGameDay
		if days < 0 {
			days = 0
		}
		who := make([]string, 0, len(d.Targets))
		for _, t := range d.Targets {
			if t >= 0 && t < len(sim.AgentNames) {
				who = append(who, sim.AgentNames[t])
			}
		}
		fmt.Fprintf(b, "- %s: %s bound to %s — %q (%d day(s) left)\n",
			d.ID, strings.Join(who, ", "), d.DesignationID, d.Text, days)
	}
}

// faithBandWord renders the faith score's band in fiction (the data-model §6
// vocabulary): prompt-side prose only — never a recorded payload.
func faithBandWord(score int) string {
	switch {
	case score >= 75:
		return "the village believes; power comes easily"
	case score >= 40:
		return "the old covenant pace holds"
	case score >= 15:
		return "doubt slows the flow"
	default:
		return "the well is nearly dry"
	}
}

// writeProphecies renders the active-prophecy block of the turn user prompt
// (spec 085 FR-013): id (the word cannot be cancelled, but the guardian must
// speak of it truthfully), the machine-checkable claim, and remaining
// game-days (floored at 0, the writeStandingOrders shape). Only ACTIVE
// prophecies show — settled ones are history the chronicle carries.
func writeProphecies(b *strings.Builder, tick int64, prophecies []sim.Prophecy) {
	var active []sim.Prophecy
	for _, p := range prophecies {
		if p.Status == "active" {
			active = append(active, p)
		}
	}
	if len(active) == 0 {
		return
	}
	b.WriteString("\nProphecies you have staked (the word, once given, stands):\n")
	for i := range active {
		p := &active[i]
		days := (p.DeadlineTick - tick) / ticksPerGameDay
		if days < 0 {
			days = 0
		}
		fmt.Fprintf(b, "- %s: %q — judged by %s (%d day(s) left)\n",
			p.ID, p.Text, describeClaim(&p.Claim), days)
	}
}

func writeStandingOrders(b *strings.Builder, tick int64, orders []sim.GuardianOrder) {
	var active []sim.GuardianOrder
	for _, o := range orders {
		if o.Status == "active" {
			active = append(active, o)
		}
	}
	if len(active) == 0 {
		return
	}
	b.WriteString("\nStanding orders you keep watch over:\n")
	for _, o := range active {
		days := (o.ExpiresTick - tick) / ticksPerGameDay
		if days < 0 {
			days = 0
		}
		kind := "structural"
		if o.Confirm {
			kind = "fuzzy"
		}
		fmt.Fprintf(b, "- %s: %q (%d day(s) left, %s)\n", o.ID, o.Condition, days, kind)
	}
}
