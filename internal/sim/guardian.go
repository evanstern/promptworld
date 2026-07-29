package sim

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/tool"
)

// Guardian's world-visible surface (TASK-12): the charge economy and the
// nudge event. Everything else about the guardian (charter, soul, console)
// lives outside deterministic space; only recorded events reach here.

const (
	// chargeRegenTicks: one charge per 6 game hours, at absolute boundaries
	// (multiples of 21600 ticks). Since spec 085 this is the STEADY BAND's
	// cadence in the faith regen curve (faith.go FaithRegenCadenceTicks) —
	// the genesis band, so a world that has never folded a faith event keeps
	// exactly this schedule; the executor's boundary check now reads the
	// band function, never this constant directly.
	chargeRegenTicks = 6 * 3600
	// GuardianChargeCap bounds the bank; GuardianGenesisCharges is day-1
	// grace (a reign begins with one favor).
	GuardianChargeCap      = 3
	GuardianGenesisCharges = 1
	// GuardianChargeRegenTicks is chargeRegenTicks exported for internal/tui,
	// mirroring the BulkCap/GuardianChargeCap export pattern (agents.go) —
	// sim stays the single source of truth for the doctrine constant; the TUI
	// never carries its own copy of "6 game hours". Since spec 085 the strip's
	// forecast reads the EFFECTIVE cadence off the wire
	// (ClockStatus.FaithRegenTicks, served from FaithRegenCadenceTicks); this
	// export survives as the steady-band value and the TUI's honest fallback
	// against a pre-085 daemon that serves no faith fields.
	GuardianChargeRegenTicks = chargeRegenTicks
)

// NudgeTextMax caps the villager-bound rendering — read from the tool registry
// (spec 014 T021/R7; re-pointed at send_vision when spec 029 retired the nudges):
// the influence tools' TextCapBytes (400). The reducer dry-run stays the
// enforcer; the registry is the single source of the cap, so the enforcer and
// the guardian-side truncation can never carry divergent literals.
var NudgeTextMax = func() int {
	t, _ := tool.Lookup("send_vision")
	return t.Cost.TextCapBytes
}()

type (
	// GuardianNudgedPayload is the injected spend + record: form "dream"
	// (exactly one living target) or "omen" (every villager alive at
	// landing, recorded explicitly).
	GuardianNudgedPayload struct {
		Form    string     `json:"form"`
		Targets []AgentRef `json:"targets"`
		Text    string     `json:"text"`
	}
	// ChargeRegeneratedPayload is empty — the event row's tick is the
	// boundary crossed.
	ChargeRegeneratedPayload struct{}
)

const (
	// ticksPerGameDay is a game day in ticks (1 tick = 1 game second, so
	// clock.secondsPerDay = 24×3600). Standing-order TTLs are game days (spec 029).
	ticksPerGameDay = 24 * 3600
	// GuardianOrderTTLMinDays / MaxDays bound a standing order's lifetime
	// (spec 029 FR-007): player-specifiable, default 3, capped 1..7 game days.
	GuardianOrderTTLMinDays = 1
	GuardianOrderTTLMaxDays = 7
	// GuardianPlayerOrderCap is the concurrent ACTIVE player-placed order cap
	// (FR-007); system-origin deferral orders are exempt (FR-012).
	GuardianPlayerOrderCap = 3
	// GuardianOriginPlayer / GuardianOriginSystem are the two GuardianOrder.Origin
	// values (spec 029): a player console monitor_and_act places "player" orders;
	// the daytime-omen deferral (spec 029 T016) and the survival watches (spec 059)
	// place "system" orders. Named here so the origin-keyed exemptions (cap/TTL/
	// cancel) have a single home rather than scattered string literals.
	GuardianOriginPlayer = "player"
	GuardianOriginSystem = "system"
)

// Survival watch kinds (spec 059 US1): the three canonical system-origin survival
// watches. A GuardianOrder whose Survival is one of these is matched by the
// survival-band predicate (not the structural orderMatches), is cap/TTL/cancel
// exempt (origin-keyed), and drives a survival-authority turn (US2). Empty
// Survival ("") is every pre-059 order — the ordinary structural watch.
const (
	SurvivalNearDeath  = "near_death"
	SurvivalStarvation = "starvation"
	SurvivalExposure   = "exposure"
)

// Survival watch bands (spec 059 FR-008): the danger thresholds the three
// survival watches match on, each REUSING an existing sim doctrine constant as
// its single home rather than a new magic number — promoted-dial-ready (named,
// one place) but deliberately NOT added to tuning.json (dials are earned by
// evidence). near_death reuses the near-death latch band (nearDeathBelow) and
// its hysteresis reset (nearDeathResetAt); starvation/exposure reuse decayNeeds'
// OWN health-loss predicate — Food==0 / Warmth==0, the exact conditions that
// drain health and stamp the "starvation"/"exposure" death causes — with the
// reflex hunger band (hungryAt) and the freezing-night band (coldNightBelow) as
// the recovery (re-arm) thresholds so a watch cannot flap on a one-tick wobble.
const (
	SurvivalNearDeathBelow = nearDeathBelow   // 200: health danger band (agents.go)
	SurvivalNearDeathRearm = nearDeathResetAt // 400: recovered — re-arm
	SurvivalStarvingAt     = 0                // Food==0: out of food, losing health
	SurvivalStarvingRearm  = hungryAt         // 350: fed enough to re-arm
	SurvivalFreezingAt     = 0                // Warmth==0: freezing, losing health
	SurvivalFreezingRearm  = coldNightBelow   // 350: warm enough to re-arm
)

// IsSurvivalKind reports whether s names a canonical survival watch kind
// (spec 059): the reducer uses it to keep the order_placed door authoritative on
// the Survival discriminator (an unknown non-empty value is refused).
func IsSurvivalKind(s string) bool {
	switch s {
	case SurvivalNearDeath, SurvivalStarvation, SurvivalExposure:
		return true
	}
	return false
}

// SurvivalWatchDefs returns the three canonical system-origin survival watches
// (spec 059 US1/FR-001) as ready-to-land GuardianOrders at the given tick — the
// SINGLE home the genesis/boot seeder and tests both build from. Ids are fixed
// and human-readable ("sys-watch-<kind>") so re-seeding is idempotent (the
// reducer rejects a duplicate id; the boot guard skips re-injection). Each
// watches agent.needs_changed (the per-game-minute heartbeat carrying the danger
// band, matched by the survival-band predicate, not the structural filter),
// pins no agent (Agent -1 = all villagers), and is non-expiring (ExpiresTick is
// ignored for a survival watch — set to PlacedTick as an honest placeholder).
// Condition/Action are the in-fiction standing duty the survival turn narrates
// under; the specific endangered villager is supplied at trigger time.
//
// RECORDED-AT-EMISSION text (spec 052 ruling 1; spec 094 doctrine): these
// strings land verbatim in recorded guardian.order_placed payloads —
// persisted vocabulary, never skinned. Rewording them is a payload-semantics
// change: it requires a store.LogFormatVersion bump + migration (and a
// wording change is NOT a pure rename, so it would snapshot-cut — see
// world/migrate.go's decision rule). The fiction-denylist sweep allowlists
// this function by name for the survival prose.
func SurvivalWatchDefs(tick int64) []GuardianOrder {
	def := func(kind, cond, act string) GuardianOrder {
		return GuardianOrder{
			ID:          "sys-watch-" + kind,
			Origin:      GuardianOriginSystem,
			Survival:    kind,
			Condition:   cond,
			Action:      act,
			EventTypes:  []string{"agent.needs_changed"},
			Agent:       -1,
			PlacedTick:  tick,
			ExpiresTick: tick, // ignored — a survival watch never expires
			Status:      "active",
		}
	}
	return []GuardianOrder{
		def(SurvivalNearDeath, "a villager is near death",
			"A villager stands at the brink of death. Act on your own authority to save a life if you can — send a vision or work a miracle — or, if nothing can be done, keep the watch and record it."),
		def(SurvivalStarvation, "a villager is starving",
			"A villager has run out of food and is starving. Act on your own authority to save a life if you can — send a vision toward food, or work a miracle — or, if nothing can be done, keep the watch and record it."),
		def(SurvivalExposure, "a villager is freezing",
			"A villager is freezing in the cold and is losing their life to exposure. Act on your own authority to save a life if you can — send a vision toward warmth, or work a miracle — or, if nothing can be done, keep the watch and record it."),
	}
}

const (
	// guardianOrderRetain bounds retained NON-ACTIVE orders (data-model §1): the
	// slice keeps every active order plus the most recent 32 consumed ones, so
	// the status/trail shows recent history without unbounded growth.
	guardianOrderRetain = 32
)

// GuardianOrder is one event-sourced standing order (spec 029, data-model §1): a
// pre-authorized watch-and-act instruction placed via monitor_and_act. Its
// lifecycle (active → triggered | cancelled | expired) is driven entirely by
// recorded events, so it reconstructs identically through snapshots, restart,
// and from-genesis replay; replay only reconstructs state — it never triggers.
type GuardianOrder struct {
	ID          string   `json:"id"`                 // "ord-<placedTick>-<seq>" (research R7)
	Origin      string   `json:"origin"`             // "player" | "system"
	Condition   string   `json:"condition"`          // original NL, ≤300 chars
	Action      string   `json:"action"`             // NL action instruction, ≤400 chars
	EventTypes  []string `json:"event_types"`        // structural predicate: non-empty
	Agent       int      `json:"agent"`              // villager index, -1 = any
	Keywords    []string `json:"keywords,omitempty"` // coarse text filter, lowercase
	Confirm     bool     `json:"confirm,omitempty"`  // fuzzy: needs the watch confirm
	PlacedTick  int64    `json:"placed_tick"`
	ExpiresTick int64    `json:"expires_tick"`       // placed + ttl_days game days (IGNORED for a survival watch — non-expiring)
	Status      string   `json:"status"`             // "active" | "triggered" | "cancelled" | "expired"
	Survival    string   `json:"survival,omitempty"` // spec 059: "" = ordinary structural order; else a survival watch kind (near_death|starvation|exposure)
	// PlacedSeq is the placement event's store seq, stamped by the reducer at
	// apply time from the event envelope (the Memory.Seq precedent, spec 054):
	// it lets the scenario rubric's OrderPlacedEvidence re-locate the recorded
	// guardian.order_placed without a log scan. Ignored on the wire payload
	// (like Status); omitempty keeps pre-054 snapshots and every injected
	// payload byte-identical.
	PlacedSeq int64 `json:"placed_seq,omitempty"`
}

// OrderPlacedPayload is guardian.order_placed's wire mirror (spec 086
// FR-003, data-model §4): GuardianOrder's fields with a named agent ref on
// the wire while the state entity above keeps the bare int (−1 = any) — the
// R2 invariant. Same json tags; legacy rows decode dual-shape and the arm
// folds .ID only.
type OrderPlacedPayload struct {
	ID          string   `json:"id"`
	Origin      string   `json:"origin"`
	Condition   string   `json:"condition"`
	Action      string   `json:"action"`
	EventTypes  []string `json:"event_types"`
	Agent       AgentRef `json:"agent"` // −1 = any (sentinel, empty name)
	Keywords    []string `json:"keywords,omitempty"`
	Confirm     bool     `json:"confirm,omitempty"`
	PlacedTick  int64    `json:"placed_tick"`
	ExpiresTick int64    `json:"expires_tick"`
	Status      string   `json:"status"`
	Survival    string   `json:"survival,omitempty"`
	PlacedSeq   int64    `json:"placed_seq,omitempty"`
}

// PlacedPayload builds the wire mirror from the entity (emission side).
func (o GuardianOrder) PlacedPayload() OrderPlacedPayload {
	return OrderPlacedPayload{
		ID: o.ID, Origin: o.Origin, Condition: o.Condition, Action: o.Action,
		EventTypes: o.EventTypes, Agent: Ref(o.Agent), Keywords: o.Keywords,
		Confirm: o.Confirm, PlacedTick: o.PlacedTick, ExpiresTick: o.ExpiresTick,
		Status: o.Status, Survival: o.Survival, PlacedSeq: o.PlacedSeq,
	}
}

// order folds the mirror back into the int-typed state entity (arm side).
func (p OrderPlacedPayload) order() GuardianOrder {
	return GuardianOrder{
		ID: p.ID, Origin: p.Origin, Condition: p.Condition, Action: p.Action,
		EventTypes: p.EventTypes, Agent: p.Agent.ID, Keywords: p.Keywords,
		Confirm: p.Confirm, PlacedTick: p.PlacedTick, ExpiresTick: p.ExpiresTick,
		Status: p.Status, Survival: p.Survival, PlacedSeq: p.PlacedSeq,
	}
}

// OrderTriggeredPayload records a matched order's one-shot consumption (spec
// 029): the matched event's type + tick ride along for the trail. Injected by
// the trigger worker (Batch B), NEVER emitted during replay.
type OrderTriggeredPayload struct {
	ID          string `json:"id"`
	MatchedType string `json:"matched_type"`
	MatchedTick int64  `json:"matched_tick"`
}

// OrderIDPayload is the bare-id payload shared by guardian.order_cancelled
// (injected — cancel_order) and guardian.order_expired (executor-emitted, a pure
// function of state + tick, like charge_regenerated).
type OrderIDPayload struct {
	ID string `json:"id"`
}

// CharterObservedPayload — guardian.charter_observed (spec 044 US2, FR-008):
// the charter revision a Guardian turn actually ran under, identified by a
// short content hash of the EFFECTIVE charter text (post-fallback,
// post-truncation — what loadCharter returned), plus whether that text is
// the authored default. Emitted by the turn pipeline through the
// inject_social door only when the fingerprint differs from
// State.CharterFingerprint (the first turn always emits), giving deaths an
// event-sourced charter-revision timeline to align against. Evidence only:
// no scoring fields, by contract (contracts/events.md).
type CharterObservedPayload struct {
	Fingerprint string `json:"fingerprint"`
	Default     bool   `json:"default"`
}

// SkillsObservedPayload — guardian.skills_observed (spec 077 FR-006): the
// bound skill-file set a Guardian turn actually ran under — charter_observed's
// twin, emitted by the same turn pipeline through the same inject_social door,
// only when the fingerprint differs from State.SkillsFingerprint. Names is
// the bound file list in composition order (provenance, like Status's skills
// list). An EMPTY bound set never emits — absence is not an observation, and
// stages 1–2 (where skill files don't bind) are structurally silent. Skill
// files are inherently player-authored (no game-shipped ones exist), so no
// default flag rides here: SkillsObservedEvidence derives Custom: true by
// construction (curriculum.go). Evidence only: no scoring fields.
type SkillsObservedPayload struct {
	Fingerprint string   `json:"fingerprint"`
	Names       []string `json:"names"`
}

// applyGuardian is the reducer arm for guardian.* events. The nudged arm
// validates rather than clamps: the InjectSocial dry-run runs this on a
// state copy, so invalid spends are rejected at the door and recorded
// events always re-apply cleanly at the same position in replay.
func (s *State) applyGuardian(e store.Event) error {
	switch e.Type {
	case "guardian.charge_regenerated":
		if s.GuardianCharges < GuardianChargeCap {
			s.GuardianCharges++
		}
	case "guardian.nudged":
		var p GuardianNudgedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		if s.GuardianCharges <= 0 {
			return fmt.Errorf("apply %s: no charges banked", e.Type)
		}
		// Form validation (spec 029): the guardian.nudged form domain is
		// {vision, omen, dream}. A vision reaches exactly one living villager at
		// any hour; an omen reaches ≥1 living villagers and lands ONLY at night
		// (State.Night); dream is the RETIRED legacy form, grandfathered here
		// (exactly one target) so pre-029 histories replay to identical state —
		// but no tool, handler, or roster entry can produce a NEW one, so
		// structural absence is the guarantee dreams cannot land afresh. This
		// explicit form switch REPLACES the spec-014 OnRoster(RosterGuardian,
		// "nudge_"+form) check, which could no longer hold once nudge_dream/
		// nudge_omen left the registry (contracts/events.md).
		switch p.Form {
		case "vision":
			if len(p.Targets) != 1 {
				return fmt.Errorf("apply %s: vision needs exactly one target, got %d", e.Type, len(p.Targets))
			}
		case "omen":
			if len(p.Targets) == 0 {
				return fmt.Errorf("apply %s: omen needs targets", e.Type)
			}
			if !s.Night {
				return fmt.Errorf("apply %s: an omen may land only at night", e.Type)
			}
		case "dream":
			// Legacy (pre-029), replay-only — grandfathered so recorded histories
			// reproduce identically; unreachable from any live tool.
			if len(p.Targets) != 1 {
				return fmt.Errorf("apply %s: dream needs exactly one target, got %d", e.Type, len(p.Targets))
			}
		default:
			return fmt.Errorf("apply %s: unknown form %q", e.Type, p.Form)
		}
		for _, t := range p.Targets {
			if t.ID < 0 || t.ID >= len(s.Agents) {
				return fmt.Errorf("apply %s: unknown target %d", e.Type, t.ID)
			}
			if s.Agents[t.ID].Dead {
				return fmt.Errorf("apply %s: target %s is dead", e.Type, s.Agents[t.ID].Name)
			}
		}
		if p.Text == "" || len(p.Text) > NudgeTextMax {
			return fmt.Errorf("apply %s: text length %d outside 1..%d", e.Type, len(p.Text), NudgeTextMax)
		}
		s.GuardianCharges--
	case "guardian.place_revealed":
		// Spec 041 (FR-014, T032): a vision's divine place grant. Validates
		// rather than clamps (the nudged arm's contract — the InjectSocial
		// dry-run runs this on a state copy, so an invalid reveal is rejected
		// at the door and a recorded one always re-applies): the target must
		// live, and every fact must name a REAL place (groundFactPresent —
		// the god reveals what is; a false vision is not a channel). Seen,
		// Provenance, and Detail are stamped here, normatively (the
		// order-Status-ignored shape): Seen = the landing tick, Detail =
		// ground truth at landing (a fire's FuelUntil), a pure function of
		// (state, event) so live and replay agree byte-for-byte. A map-less
		// agent skips the upsert — the reducer stays total (agent.saw shape).
		var p PlaceRevealedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		if p.Agent.ID < 0 || p.Agent.ID >= len(s.Agents) {
			return fmt.Errorf("apply %s: unknown target %d", e.Type, p.Agent.ID)
		}
		a := &s.Agents[p.Agent.ID]
		if a.Dead {
			return fmt.Errorf("apply %s: target %s is dead", e.Type, a.Name)
		}
		if len(p.Facts) == 0 {
			return fmt.Errorf("apply %s: no facts to reveal", e.Type)
		}
		for _, f := range p.Facts {
			if s.m != nil && !groundFactPresent(s, s.m, f) {
				return fmt.Errorf("apply %s: no %s at (%d,%d)", e.Type, f.Kind, f.X, f.Y)
			}
			if a.Map != nil {
				a.Map.upsertFact(PlaceFact{Kind: f.Kind, X: f.X, Y: f.Y, Seen: e.Tick,
					Provenance: ProvenanceRevealed, Detail: groundFactDetail(s, f)})
			}
		}
	case "guardian.order_placed":
		// Spec 086: decode the OrderPlacedPayload mirror (named ref,
		// dual-shape for legacy rows) and fold .ID into the int-typed entity
		// — names never enter state (R2), no name validation here (R3).
		var mp OrderPlacedPayload
		if err := json.Unmarshal(e.Payload, &mp); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		o := mp.order()
		if o.ID == "" {
			return fmt.Errorf("apply %s: empty order id", e.Type)
		}
		// Duplicate id in ANY status is rejected — ids are assigned once and
		// consumed orders are retained, so a reused id would corrupt the trail.
		for i := range s.GuardianOrders {
			if s.GuardianOrders[i].ID == o.ID {
				return fmt.Errorf("apply %s: duplicate order id %q", e.Type, o.ID)
			}
		}
		switch o.Origin {
		case GuardianOriginPlayer, GuardianOriginSystem:
		default:
			return fmt.Errorf("apply %s: unknown origin %q", e.Type, o.Origin)
		}
		// A survival watch (spec 059) is a system-origin order with a known kind;
		// it is exempt from the TTL bounds below (non-expiring). A non-empty
		// Survival on a player order, or an unknown kind, is refused — the door
		// stays authoritative on the discriminator.
		if o.Survival != "" {
			if o.Origin != GuardianOriginSystem {
				return fmt.Errorf("apply %s: survival watch must be system origin", e.Type)
			}
			if !IsSurvivalKind(o.Survival) {
				return fmt.Errorf("apply %s: unknown survival kind %q", e.Type, o.Survival)
			}
		}
		if len(o.EventTypes) == 0 {
			return fmt.Errorf("apply %s: order has no event_types (uncompilable condition)", e.Type)
		}
		// TTL bounds hold for every order EXCEPT a survival watch, which is
		// non-expiring by nature (spec 059 FR-002) — its ExpiresTick is ignored
		// here and by the executor's expiry sweep alike (origin-keyed exemption,
		// not a giant TTL).
		if o.Survival == "" {
			if ttl := o.ExpiresTick - o.PlacedTick; ttl < GuardianOrderTTLMinDays*ticksPerGameDay || ttl > GuardianOrderTTLMaxDays*ticksPerGameDay {
				return fmt.Errorf("apply %s: ttl %d ticks outside %d..%d game days", e.Type, ttl, GuardianOrderTTLMinDays, GuardianOrderTTLMaxDays)
			}
		}
		if o.Agent < -1 || o.Agent >= len(s.Agents) {
			return fmt.Errorf("apply %s: agent index %d out of range", e.Type, o.Agent)
		}
		if utf8.RuneCountInString(o.Condition) > 300 {
			return fmt.Errorf("apply %s: condition over 300 chars", e.Type)
		}
		if utf8.RuneCountInString(o.Action) > 400 {
			return fmt.Errorf("apply %s: action over 400 chars", e.Type)
		}
		// Concurrent cap: at most 3 ACTIVE player-origin orders; system-origin
		// deferral orders are exempt (already-authorized acts, FR-012).
		if o.Origin == "player" {
			active := 0
			for i := range s.GuardianOrders {
				if s.GuardianOrders[i].Origin == "player" && s.GuardianOrders[i].Status == "active" {
					active++
				}
			}
			if active >= GuardianPlayerOrderCap {
				return fmt.Errorf("apply %s: %d player orders already active (cap %d)", e.Type, active, GuardianPlayerOrderCap)
			}
		}
		// The status field is IGNORED on the payload — an order always lands
		// active (data-model §2), then the retention prune runs. PlacedSeq is
		// likewise reducer-stamped, never payload-trusted (spec 054): the
		// event's own store seq, identical live and in replay (stampSeqs
		// pre-assigns the same last+i+1 AppendEvents records; the dry-run
		// probe applies with Seq 0 and is discarded).
		o.Status = "active"
		o.PlacedSeq = e.Seq
		s.GuardianOrders = pruneGuardianOrders(append(s.GuardianOrders, o))
	case "guardian.order_triggered":
		var p OrderTriggeredPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		return s.transitionGuardianOrder(e.Type, p.ID, "triggered")
	case "guardian.order_cancelled":
		var p OrderIDPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		// A survival watch is the guardian's nature, not a player configuration
		// (spec 059 FR-002): the player-order surface cannot cancel it. Refuse at
		// the door, keyed on the Survival discriminator, so the in-fiction refusal
		// is authoritative rather than advisory.
		for i := range s.GuardianOrders {
			if s.GuardianOrders[i].ID == p.ID && s.GuardianOrders[i].Survival != "" {
				return fmt.Errorf("apply %s: order %q is a survival watch and cannot be cancelled", e.Type, p.ID)
			}
		}
		return s.transitionGuardianOrder(e.Type, p.ID, "cancelled")
	case "guardian.order_expired":
		var p OrderIDPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		return s.transitionGuardianOrder(e.Type, p.ID, "expired")
	case "guardian.charter_observed":
		// Spec 044 US2 (FR-008): the charter-revision timeline. State keeps only
		// the CURRENT fingerprint and its authorship (spec 072 FR-006 — the-law's
		// rubric charter conjunct); the full timeline lives in the event log,
		// where the morgue's render scan aligns each death against the most
		// recent observation at or before it.
		var p CharterObservedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		if p.Fingerprint == "" {
			return fmt.Errorf("apply %s: empty fingerprint", e.Type)
		}
		s.CharterFingerprint = p.Fingerprint
		s.CharterCustom = !p.Default
		// Spec 077 (FR-005): persist the observation's envelope coordinates
		// (the PlacedSeq apply-time-stamp precedent) so the-law's pass
		// evidence is state-derivable (CharterEvidenceFromState) — the
		// spec-072 FR-009 blocker removed. The dry-run probe applies with
		// Seq 0 and is discarded; the recorded event stamps the real seq,
		// identical live and in replay.
		s.CharterObservedSeq = e.Seq
		s.CharterObservedTick = e.Tick
	case "guardian.skills_observed":
		// Spec 077 (FR-006): the skills-observation twin of the charter arm
		// above — latest observation wins, exactly how the charter
		// fingerprint is kept. The door requires a non-empty fingerprint AND
		// a non-empty name list: an empty bound set is never an observation.
		var p SkillsObservedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		if p.Fingerprint == "" {
			return fmt.Errorf("apply %s: empty fingerprint", e.Type)
		}
		if len(p.Names) == 0 {
			return fmt.Errorf("apply %s: empty skill-file set (absence is not an observation)", e.Type)
		}
		s.SkillsFingerprint = p.Fingerprint
		s.SkillsObservedSeq = e.Seq
		s.SkillsObservedTick = e.Tick
	}
	return nil
}

// transitionGuardianOrder moves the order named id from active to a terminal
// status (spec 029, one-way). An unknown id or an order not currently active is
// rejected at the door — this is where the cancel/expiry/trigger races resolve:
// exactly one terminal lands, and the loser hits a non-active order and refuses
// (contracts/events.md edge cases).
func (s *State) transitionGuardianOrder(eventType, id, to string) error {
	for i := range s.GuardianOrders {
		if s.GuardianOrders[i].ID != id {
			continue
		}
		if s.GuardianOrders[i].Status != "active" {
			return fmt.Errorf("apply %s: order %q is not active (status %q)", eventType, id, s.GuardianOrders[i].Status)
		}
		s.GuardianOrders[i].Status = to
		return nil
	}
	return fmt.Errorf("apply %s: unknown order %q", eventType, id)
}

// pruneGuardianOrders retains every active order plus the most recent
// guardianOrderRetain (32) non-active ones, dropping the oldest consumed orders
// first while preserving slice order (data-model §1). Deterministic — a pure
// function of the append-ordered slice, so replay prunes identically.
func pruneGuardianOrders(orders []GuardianOrder) []GuardianOrder {
	nonActive := 0
	for i := range orders {
		if orders[i].Status != "active" {
			nonActive++
		}
	}
	drop := nonActive - guardianOrderRetain
	if drop <= 0 {
		return orders
	}
	out := make([]GuardianOrder, 0, len(orders)-drop)
	for i := range orders {
		if orders[i].Status != "active" && drop > 0 {
			drop--
			continue
		}
		out = append(out, orders[i])
	}
	return out
}
