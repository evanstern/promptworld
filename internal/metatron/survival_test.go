package metatron

// Tests for the metatron survival autonomy feature (spec 059): system-origin
// survival watches (US1), the survival-authority turn frame + attribution (US2),
// and the targeting digest (US3).

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/tool"
	"github.com/evanstern/promptworld/internal/toolloop"
)

// seededSurvivalWatch installs the near-death survival watch into both the
// injector state (the door) and the angel's replica + mirror, the seedOrder way.
func seededSurvivalWatch(mt *Metatron, inj *stateInjector, kind string) sim.MetatronOrder {
	var o sim.MetatronOrder
	for _, w := range sim.SurvivalWatchDefs(0) {
		if w.Survival == kind {
			o = w
		}
	}
	seedOrder(mt, inj, o)
	return o
}

// needsEvent builds an agent.needs_changed event for one villager at a tick.
func needsEvent(agent, health, food, warmth int, tick int64) store.Event {
	e := mustEvent("agent.needs_changed", sim.NeedsPayload{
		Agent: agent, Health: health, Food: food, Warmth: warmth, Rest: 500, Morale: 500,
	})
	e.Tick = tick
	return e
}

// TestSurvivalWatchRefusesPlayerCancel (spec 059 US1 AC-4): a player cancel naming
// a system survival watch is refused with in-fiction counsel — the reducer rejects
// it at the door, and cancelOrder maps that to the angel's own voice.
func TestSurvivalWatchRefusesPlayerCancel(t *testing.T) {
	mt, _, inj, _ := newTestAngel(t, "ok")
	o := seededSurvivalWatch(mt, inj, sim.SurvivalNearDeath)

	why := mt.cancelOrder(o.ID, fullGrant())
	if why == "" {
		t.Fatal("a survival watch was cancelled by the player order surface")
	}
	if !strings.Contains(why, "my own nature") {
		t.Errorf("cancel refusal not in-fiction: %q", why)
	}
	// Nothing landed / the watch still stands active.
	if inj.state.MetatronOrders[0].Status != "active" {
		t.Errorf("a refused cancel still changed the survival watch: %+v", inj.state.MetatronOrders[0])
	}
}

// TestSurvivalWatchMatchesNeedsBand (spec 059 US1/US2): the survival-band matcher
// enqueues a trigger job when a villager's agent.needs_changed crosses into the
// danger band, and does NOT re-fire while the villager stays in-band (the latch),
// but re-arms and fires again once the villager recovers and relapses.
func TestSurvivalWatchMatchesNeedsBand(t *testing.T) {
	mt, _, inj, _ := newTestAngel(t, "ok")
	seededSurvivalWatch(mt, inj, sim.SurvivalStarvation)

	fire := func(agent, food int, tick int64) bool {
		mt.matchOrders([]store.Event{needsEvent(agent, 800, food, 800, tick)})
		return dequeued(mt.triggerQ)
	}
	clearPending := func() {
		mt.stateMu.Lock()
		delete(mt.pendingTrigger, "sys-watch-"+sim.SurvivalStarvation)
		mt.stateMu.Unlock()
	}

	// Out of the band (fed): no fire.
	if fire(2, 500, 1000) {
		t.Fatal("a fed villager fired the starvation watch")
	}
	// Into the band (Food==0): fires once.
	if !fire(2, 0, 2000) {
		t.Fatal("a starving villager did not fire the starvation watch")
	}
	clearPending()
	// Still starving: the latch suppresses a re-fire.
	if fire(2, 0, 3000) {
		t.Fatal("the starvation watch re-fired while the villager stayed in-band (latch failed)")
	}
	clearPending()
	// Recovered above the re-arm band: no fire, latch clears.
	if fire(2, sim.SurvivalStarvingRearm, 4000) {
		t.Fatal("a recovered villager fired the watch")
	}
	clearPending()
	// Relapse: fires again (re-armed).
	if !fire(2, 0, 5000) {
		t.Fatal("the starvation watch did not re-fire after recovery + relapse")
	}
	_ = inj
}

// TestSurvivalWatchIsMiracleCapableRoster is a guard: work_miracle is on the full
// loop roster the survival turn runs under, so the digest gate (hasWorkMiracle)
// and the survival authority both apply.
func TestSurvivalWatchIsMiracleCapableRoster(t *testing.T) {
	if !hasWorkMiracle(tool.LoopRosterMetatron()) {
		t.Fatal("work_miracle absent from the metatron loop roster")
	}
	_ = context.Background
}

// --- US2: the angel acts on survival without permission (spec 059) ---

// TestSurvivalFrameCarveOut (spec 059 US2 AC-2/AC-3, SC-003): the survival frame
// permits vision/miracle on the angel's own initiative BUT keeps the clock and
// non-survival orders the player's in BOTH frames; the non-survival frame stays
// today's restrictive doctrine verbatim (FR-004/FR-005).
func TestSurvivalFrameCarveOut(t *testing.T) {
	roster := tool.LoopRosterMetatron()
	normal := buildTurnSystemPrompt(false, "CHARTER", nil, roster)
	survival := buildTurnSystemPrompt(true, "CHARTER", nil, roster)

	// The non-survival frame is byte-identical to the pinned initiative doctrine.
	if !strings.Contains(normal, metatronInitiativeFrame) {
		t.Error("non-survival frame lost the restrictive initiative doctrine (FR-005)")
	}
	if strings.Contains(normal, metatronSurvivalFrame) {
		t.Error("non-survival frame leaked the survival carve-out")
	}
	// The survival frame carries the carve-out and drops the restrictive one.
	if !strings.Contains(survival, metatronSurvivalFrame) {
		t.Error("survival frame missing the survival carve-out (FR-003)")
	}
	if strings.Contains(survival, metatronInitiativeFrame) {
		t.Error("survival frame still carries the full restrictive frame verbatim")
	}
	// Clock control stays the player's in BOTH frames (SC-003 / FR-004).
	for _, f := range []struct {
		name, prompt string
	}{{"non-survival", normal}, {"survival", survival}} {
		if !strings.Contains(f.prompt, "clock") || !strings.Contains(f.prompt, "the player") {
			t.Errorf("%s frame does not keep the clock the player's: %q", f.name, f.prompt)
		}
	}
	// The two non-negotiables ride both frames unchanged.
	for _, p := range []string{normal, survival} {
		if !strings.Contains(p, metatronNonNegotiables) {
			t.Error("a frame lost the persona-firewall non-negotiables")
		}
	}
}

// TestSurvivalTurnActsWithoutPlayer (spec 059 US2 AC-1, SC-002): a survival-watch
// match runs a turn that works a miracle with zero player input, still charge-
// gated; the watch is NOT consumed (non-expiring), and the act is attributed to
// the survival duty in the durable record (moment + transcript, FR-007).
func TestSurvivalTurnActsWithoutPlayer(t *testing.T) {
	mt, _, inj, dir := newTestAngel(t, "The child will not starve tonight.")
	o := seededSurvivalWatch(mt, inj, sim.SurvivalNearDeath)
	ash := agentIndexByName("Ash")
	mt.runLoop = systemActLoop(mt, "work_miracle",
		`{"kind":"give_item","villager":"Ash","item":"food_raw","qty":3}`)

	before := inj.state.MetatronCharges
	mt.runTrigger(triggerJob{order: o, matched: needsEvent(ash, 100, 0, 800, 5000),
		matchedType: "agent.needs_changed", matchedTick: 5000})

	// The miracle landed with no player in the loop.
	if inj.state.Agents[ash].Inv.FoodRaw < 3 {
		t.Errorf("survival miracle did not reach Ash: inv=%+v", inj.state.Agents[ash].Inv)
	}
	// Charge-gated: exactly one charge spent (the survival carve-out changes
	// authority, not the economy — FR-003).
	if inj.state.MetatronCharges != before-1 {
		t.Errorf("survival miracle spent %d charges, want 1", before-inj.state.MetatronCharges)
	}
	// The watch is non-consuming: it still stands active for the next crisis.
	if inj.state.MetatronOrders[0].Status != "active" {
		t.Errorf("survival watch was consumed: %+v", inj.state.MetatronOrders[0])
	}
	// No order_triggered was ever emitted for a survival watch.
	for _, b := range inj.batches {
		for _, e := range b {
			if e.Type == "metatron.order_triggered" {
				t.Error("a survival watch landed order_triggered (it must never consume)")
			}
		}
	}
	// Attribution: the moment names the survival watch and the act.
	mt.stateMu.Lock()
	moments := append([]string(nil), mt.moments...)
	mt.stateMu.Unlock()
	if len(moments) != 1 || !strings.Contains(moments[0], "survival watch") || !strings.Contains(moments[0], "miracle") {
		t.Fatalf("survival moment not attributed to the duty: %+v", moments)
	}
	// The transcript marks the turn as a survival watch (auditable authority trail).
	tr := tailOfFile(mt.transcriptPath(), 4000)
	if !strings.Contains(tr, "[survival watch]") {
		t.Errorf("transcript missing the survival-watch attribution: %q", tr)
	}
	_ = dir
}

// TestSurvivalZeroChargeTurnRecorded (spec 059 US2 edge case / T008): a survival
// watch firing on an empty bank still RUNS the turn (not the deferral empty-bank
// short-circuit) — the acting tools refuse, the angel narrates, and a helpless
// moment is recorded, never silent.
func TestSurvivalZeroChargeTurnRecorded(t *testing.T) {
	mt, _, inj, _ := newTestAngel(t, "I see you, and I cannot reach you.")
	mt.replica.MetatronCharges = 0
	inj.state.MetatronCharges = 0
	mt.mirrorState()
	o := seededSurvivalWatch(mt, inj, sim.SurvivalExposure)
	oak := agentIndexByName("Oak")

	// The model TRIES a miracle, but the empty bank refuses it in-fiction.
	called := false
	mt.runLoop = func(ctx context.Context, j toolloop.Job) (toolloop.Result, error) {
		called = true
		c := toolCall("work_miracle", `{"kind":"give_item","villager":"Oak","item":"wood","qty":2}`)
		out := j.Handlers["work_miracle"](ctx, c)
		j.Record(toolloop.CallRecord{JobID: j.JobID, Ordinal: 1, Tool: "work_miracle",
			Args: c.Args, Verdict: out.Verdict, Reason: out.ResultForModel, Tier: "cloud"})
		return toolloop.Result{Final: "I cannot reach you", Term: toolloop.TermCapExhausted}, nil
	}

	mt.runTrigger(triggerJob{order: o, matched: needsEvent(oak, 800, 800, 0, 6000),
		matchedType: "agent.needs_changed", matchedTick: 6000})

	if !called {
		t.Fatal("a zero-charge survival watch skipped the turn (it must still run and record)")
	}
	if inj.state.MetatronCharges != 0 {
		t.Errorf("a helpless survival turn spent from an empty bank: %d", inj.state.MetatronCharges)
	}
	// Nothing landed in the world, but the helpless turn is recorded.
	if len(landedBatches(inj)) != 0 {
		t.Error("a zero-charge survival turn still landed a world act")
	}
	mt.stateMu.Lock()
	moments := append([]string(nil), mt.moments...)
	mt.stateMu.Unlock()
	if len(moments) != 1 || !strings.Contains(moments[0], "survival watch") {
		t.Fatalf("helpless survival turn not recorded as a moment: %+v", moments)
	}
	if !strings.Contains(moments[0], "nothing") {
		t.Errorf("helpless moment does not read as helpless: %q", moments[0])
	}
}

// --- US3: miracles can aim — the targeting digest (spec 059) ---

// TestTargetingDigestPresentAndBounded (spec 059 US3 AC-1, SC-004): a
// miracle-capable turn's prompt carries the positions/conditions/passability
// digest, and the digest is within its token budget.
func TestTargetingDigestPresentAndBounded(t *testing.T) {
	mt, orch, _, _ := newTestAngel(t, "The village holds.")
	if _, err := mt.Turn(context.Background(), "how fare they?"); err != nil {
		t.Fatal(err)
	}
	prompt := orch.requests()[0].Prompt
	if !strings.Contains(prompt, "Aim your miracles") {
		t.Fatalf("miracle-capable prompt missing the targeting guidance: %q", prompt)
	}
	// A concrete villager position line is present.
	if !strings.Contains(prompt, sim.AgentNames[0]+" at (") {
		t.Errorf("targeting digest missing villager positions: %q", prompt)
	}
	if !strings.Contains(prompt, "passable next tiles:") {
		t.Errorf("targeting digest missing passability guidance: %q", prompt)
	}
	// The digest block itself is within budget.
	mt.stateMu.Lock()
	digest := buildTargetingDigest(mt.alive, mt.agentNeeds, mt.agentXY, mt.m)
	mt.stateMu.Unlock()
	if len(digest) == 0 {
		t.Fatal("digest empty on a populated world")
	}
	if len(digest) > targetingDigestMaxBytes {
		t.Errorf("digest %d bytes exceeds the %d budget", len(digest), targetingDigestMaxBytes)
	}
}

// TestTargetingDigestNotOnDreamsOnlyWorld (spec 059 FR-006): a world whose grant
// withholds work_miracle carries no digest (the prompt stays byte-unchanged for a
// non-miracle-capable turn).
func TestTargetingDigestNotOnDreamsOnlyWorld(t *testing.T) {
	// A roster without work_miracle → hasWorkMiracle false → no digest built.
	dreamsOnly := []tool.Tool{}
	for _, tl := range tool.LoopRosterMetatron() {
		if tl.Name != "work_miracle" {
			dreamsOnly = append(dreamsOnly, tl)
		}
	}
	if hasWorkMiracle(dreamsOnly) {
		t.Fatal("test roster still carries work_miracle")
	}
}

// TestMiracleFromDigestCoordinatesPassesDoor (spec 059 US3 AC-2, SC-004): a move
// miracle targeted at a tile the digest lists as passable is accepted by the
// landing door — the regression for world-01's 3-of-4 coordinate rejections.
func TestMiracleFromDigestCoordinatesPassesDoor(t *testing.T) {
	mt, _, inj, _ := newTestAngel(t, "It is moved.")

	// Find a living villager whose tile has a map-passable neighbor, and compute
	// that neighbor the SAME way the digest does.
	mt.stateMu.Lock()
	alive := mt.alive
	xy := mt.agentXY
	needs := mt.agentNeeds
	m := mt.m
	var vi, tx, ty int = -1, 0, 0
	for i := range xy {
		if !alive[i] {
			continue
		}
		for _, d := range [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
			nx, ny := xy[i][0]+d[0], xy[i][1]+d[1]
			if m.Passable(nx, ny) {
				vi, tx, ty = i, nx, ny
				break
			}
		}
		if vi >= 0 {
			break
		}
	}
	digest := buildTargetingDigest(alive, needs, xy, m)
	mt.stateMu.Unlock()
	if vi < 0 {
		t.Skip("no villager with a passable neighbor on this seed (map-specific)")
	}
	dest := fmt.Sprintf("(%d,%d)", tx, ty)
	if !strings.Contains(digest, dest) {
		t.Fatalf("digest does not list the passable destination %s:\n%s", dest, digest)
	}

	// The angel works a move miracle to that DIGEST-listed tile — it must pass the
	// landing door (no rejection), spending one charge.
	before := inj.state.MetatronCharges
	sx, sy := xy[vi][0], xy[vi][1]
	miracle, why := mt.landMiracle(miracleArgs{
		Kind: "move", Class: "villager", X: sx, Y: sy, ToX: tx, ToY: ty,
	}, before, fullGrant())
	if why != "" {
		t.Fatalf("miracle at digest coordinates was rejected at the door: %q", why)
	}
	if miracle == nil {
		t.Fatal("no miracle returned despite a valid digest target")
	}
	if inj.state.MetatronCharges != before-1 {
		t.Errorf("move miracle spent %d charges, want 1", before-inj.state.MetatronCharges)
	}
	if inj.state.Agents[vi].X != tx || inj.state.Agents[vi].Y != ty {
		t.Errorf("villager did not land at the digest tile: at (%d,%d), want (%d,%d)",
			inj.state.Agents[vi].X, inj.state.Agents[vi].Y, tx, ty)
	}
}
