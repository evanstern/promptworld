package mind

import (
	"fmt"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// legacyUserPrompt is a FROZEN byte-for-byte copy of userPrompt as it stood
// before spec 043 wrapped its content into the block assembler (context.go).
// The golden-equality test below asserts the refactor changed no legacy bytes:
// the assembled prompt, with the NEW self_history block removed and the US2
// needs block normalized back to this frozen line (see TestContextGoldenIdentity),
// must equal this reference exactly. Keeping the reference here (not deleting the
// old code from history) is what makes "byte-identical when no new blocks
// render" (T003) a live gate rather than a claim.
func legacyUserPrompt(s *sim.State, idx int, k int, mode string) string {
	a := s.Agents[idx]
	var b strings.Builder

	phase := "daytime"
	if s.Night {
		phase = "night"
	}
	fmt.Fprintf(&b, "It is %s (%s). You are at (%d, %d).\n", clock.Format(s.Tick), phase, a.X, a.Y)
	fmt.Fprintf(&b, "Needs (0-100): health %d, food %d, rest %d, warmth %d, morale %d.\n",
		a.Needs.Health/10, a.Needs.Food/10, a.Needs.Rest/10, a.Needs.Warmth/10, a.Needs.Morale/10)
	fmt.Fprintf(&b, "Carrying: %d wood, %d stone, %d water, %d planks, %d refined stone, food (%d raw, %d cooked, %d meals)",
		a.Inv.Wood, a.Inv.Stone, a.Inv.Water, a.Inv.Planks, a.Inv.RefinedStone,
		a.Inv.FoodRaw, a.Inv.FoodCooked, a.Inv.Meals)
	if n := len(a.Inv.Spears); n > 0 {
		fmt.Fprintf(&b, ", %d spear(s) (%d uses left on the most-worn)", n, a.Inv.Spears[0])
	}
	b.WriteString(".\n")
	b.WriteString(knownPlaces(s, idx))
	var nearby []string
	if a.Map != nil {
		for _, p := range a.Map.Peers {
			if p.Agent == idx || p.Agent < 0 || p.Agent >= len(s.Agents) {
				continue
			}
			o := s.Agents[p.Agent]
			if o.Dead {
				continue
			}
			d := absInt(p.X-a.X) + absInt(p.Y-a.Y)
			if d > 10 {
				continue
			}
			state := ""
			if o.Asleep && o.X == p.X && o.Y == p.Y {
				state = ", asleep"
			}
			nearby = append(nearby, fmt.Sprintf("%s (%d tiles away%s)", o.Name, d, state))
		}
	}
	if len(nearby) > 0 {
		fmt.Fprintf(&b, "Nearby: %s.\n", strings.Join(nearby, ", "))
	}
	if social := socialContext(s, idx); social != "" {
		b.WriteString(social)
	}
	if law := villageLaw(s, idx); law != "" {
		b.WriteString(law)
	}
	window := selectWindow(s, idx, k, s.Tick, mode)
	if len(window) > 0 {
		b.WriteString("\nYou remember:\n")
		for _, m := range window {
			fmt.Fprintf(&b, "- %s\n", sim.FormatMemory(m))
		}
	}
	b.WriteString("\nWhat do you do next?")
	return b.String()
}

// richContextState builds a mapped state whose agent 0 exercises every block
// present in this slice: a known landmark + a peer sighting (known_places /
// nearby), an open debt (social_law), and several memories (memories). Needs
// and inventory are set to non-trivial values. IntentLog is left empty by
// default so self_history renders its first-thought line; callers wanting the
// populated case append records themselves.
func richContextState(t *testing.T) *sim.State {
	t.Helper()
	s := knownPlacesState(t)
	a := &s.Agents[0]
	a.X, a.Y = 20, 20
	a.Needs = sim.Needs{Health: 800, Food: 430, Rest: 620, Warmth: 450, Morale: 700}
	a.Inv = sim.Inventory{Wood: 3, Stone: 1, Water: 2}
	addFact(a.Map, fact("fire", 12, 34, "witnessed", 0, 900000))
	addSighting(a.Map, 1, a.X+3, a.Y, 100)
	s.Debts = []sim.Debt{{ID: 1, Debtor: 0, Creditor: 1, Kind: "food", Due: s.Tick + 10000, Status: "open"}}
	for i := int64(0); i < 4; i++ {
		a.Memories = append(a.Memories, sim.Memory{
			Text:     fmt.Sprintf("Memory number %d about the settlement.", i),
			Salience: 5, Tick: 100 + i, Subject: -1, Seq: 1 + i,
		})
	}
	return s
}

// legacyNeedsLine is the FROZEN pre-043 needs line (no trajectory clause). The
// golden test below swaps the live needs block back to this before comparing, so
// the byte-identity gate still bites on every OTHER legacy block while allowing
// the one deliberate US2 change (trajectories) through.
func legacyNeedsLine(s *sim.State, idx int) string {
	a := s.Agents[idx]
	return fmt.Sprintf("Needs (0-100): health %d, food %d, rest %d, warmth %d, morale %d.\n",
		a.Needs.Health/10, a.Needs.Food/10, a.Needs.Rest/10, a.Needs.Warmth/10, a.Needs.Morale/10)
}

// TestContextGoldenIdentity (T003, FR-009): wrapping userPrompt content into the
// block assembler changed NO legacy bytes. The production prompt, with the new
// self_history block removed AND the needs block normalized back to its frozen
// pre-043 form, is byte-identical to the frozen pre-043 rendering — for both the
// first-thought (empty IntentLog) and the populated cases, and across memory
// modes.
//
// The needs block is DELIBERATELY excepted: US2 (T016) added a trajectory clause
// ("... and rising/falling/steady") to every need, so the block's bytes now
// differ from the frozen legacy string BY DESIGN — this is a feature, not the
// refactor drift the golden gate guards against. Swapping renderNeeds' output for
// legacyNeedsLine before the compare proves the trajectory suffix is the ONLY
// change to the needs line and that no other legacy block moved.
func TestContextGoldenIdentity(t *testing.T) {
	for _, mode := range []string{"", "shadow", "on"} {
		t.Run("mode="+mode, func(t *testing.T) {
			s := richContextState(t)

			// Normalize away the two 043-modified/added blocks: strip the new
			// self_history block and swap the US2 needs block back to its frozen
			// legacy line, then require byte-identity with the pre-043 reference.
			normalize := func(got string) string {
				got = strings.Replace(got, renderSelfHistory(s, 0), "", 1)
				return strings.Replace(got, renderNeeds(s, 0), legacyNeedsLine(s, 0), 1)
			}

			// First-thought agent: self_history is the "no prior activity" line.
			if renderSelfHistory(s, 0) == "" {
				t.Fatal("self_history rendered empty — it must always render (first-thought line)")
			}
			if stripped := normalize(userPrompt(s, 0, sim.WindowK, mode)); stripped != legacyUserPrompt(s, 0, sim.WindowK, mode) {
				t.Errorf("assembled prompt (minus self_history, needs normalized) diverged from legacy bytes:\ngot:    %q\nlegacy: %q", stripped, legacyUserPrompt(s, 0, sim.WindowK, mode))
			}

			// Populated IntentLog: same invariant, self_history now a real log.
			s.Agents[0].IntentLog = append(s.Agents[0].IntentLog, sim.IntentRecord{Goal: "forage", Source: "reflex", Tick: 90})
			if stripped := normalize(userPrompt(s, 0, sim.WindowK, mode)); stripped != legacyUserPrompt(s, 0, sim.WindowK, mode) {
				t.Errorf("with populated IntentLog, prompt (minus self_history, needs normalized) diverged from legacy:\ngot:    %q\nlegacy: %q", stripped, legacyUserPrompt(s, 0, sim.WindowK, mode))
			}
		})
	}
}

// TestContextDeterministic (T005, FR-009): identical world state ⇒ identical
// assembled bytes, identical per-block sizes, identical drop set.
func TestContextDeterministic(t *testing.T) {
	s := richContextState(t)
	a := assembleContext(s, 0, sim.WindowK, "", "")
	b := assembleContext(s, 0, sim.WindowK, "", "")
	if a.text != b.text {
		t.Errorf("assembly not deterministic:\nfirst:  %q\nsecond: %q", a.text, b.text)
	}
	if a.promptBytes != b.promptBytes || len(a.blockBytes) != len(b.blockBytes) {
		t.Errorf("sizes not deterministic: %+v vs %+v", a, b)
	}
	for name, sz := range a.blockBytes {
		if b.blockBytes[name] != sz {
			t.Errorf("block %q size not deterministic: %d vs %d", name, sz, b.blockBytes[name])
		}
	}
}

// TestContextTelemetrySizes (T005, FR-010): promptBytes is the true assembled
// length, and the per-block bytes plus the fixed closer account for every byte.
func TestContextTelemetrySizes(t *testing.T) {
	s := richContextState(t)
	asm := assembleContext(s, 0, sim.WindowK, "", "")
	if asm.promptBytes != len(asm.text) {
		t.Errorf("promptBytes %d != len(text) %d", asm.promptBytes, len(asm.text))
	}
	sum := len(promptCloser)
	for _, sz := range asm.blockBytes {
		sum += sz
	}
	if sum != asm.promptBytes {
		t.Errorf("block bytes + closer = %d, want promptBytes %d", sum, asm.promptBytes)
	}
	// Every present block is a known contract block; survival blocks are always
	// present in this fixture.
	for _, must := range []string{"frame", "needs", "self_history", "inventory"} {
		if _, ok := asm.blockBytes[must]; !ok {
			t.Errorf("survival block %q missing from blockBytes: %v", must, asm.blockBytes)
		}
	}
}

// TestContextDropOrder (T005 + T021, FR-008, contract §Budget): under budget
// pressure whole blocks drop lowest-priority-first. In this fixture (4 memories
// = exactly the protected floor, no serendipity, no journal, no plan) the only
// droppable blocks are social_law (4) and known_places (5); the memories block's
// floor is NEVER dropped, so "memories" survives at every budget. Survival
// blocks (frame, needs, self_history, inventory) are never dropped either.
func TestContextDropOrder(t *testing.T) {
	s := richContextState(t)

	// Ample budget: nothing dropped, everything present.
	full := assembleContext(s, 0, sim.WindowK, "", "")
	if len(full.droppedBlocks) != 0 {
		t.Fatalf("ample budget dropped blocks: %v", full.droppedBlocks)
	}
	for _, name := range []string{"known_places", "social_law", "memories"} {
		if _, ok := full.blockBytes[name]; !ok {
			t.Fatalf("fixture did not populate droppable block %q — the drop test cannot bite: %v", name, full.blockBytes)
		}
	}

	// Zero budget: the droppable blocks shed in ascending-priority order;
	// survival blocks AND the memories floor remain.
	starved := assembleBudget(s, 0, sim.WindowK, "", "", 0)
	wantDropped := []string{"social_law", "known_places"}
	if strings.Join(starved.droppedBlocks, ",") != strings.Join(wantDropped, ",") {
		t.Errorf("drop order = %v, want %v", starved.droppedBlocks, wantDropped)
	}
	for _, survive := range []string{"frame", "needs", "self_history", "inventory", "memories"} {
		if _, ok := starved.blockBytes[survive]; !ok {
			t.Errorf("protected block %q dropped under zero budget: blocks=%v dropped=%v", survive, starved.blockBytes, starved.droppedBlocks)
		}
	}
	for _, d := range starved.droppedBlocks {
		switch d {
		case "frame", "needs", "self_history", "inventory", "memories":
			t.Errorf("protected block %q appears in dropped set", d)
		}
	}

	// A budget just below the full size drops ONLY the lowest-priority present
	// block first (social_law), never a higher-priority one before it.
	fullTokens := approxTokens(full.promptBytes)
	oneDrop := assembleBudget(s, 0, sim.WindowK, "", "", fullTokens-1)
	if len(oneDrop.droppedBlocks) == 0 || oneDrop.droppedBlocks[0] != "social_law" {
		t.Errorf("first block dropped = %v, want social_law first", oneDrop.droppedBlocks)
	}
}

// TestNeedsTrajectoryDirections (T016, SC-003, US2 AS-1/2/3): each need renders
// its level plus a direction derived from current − anchor with the deadband —
// rising when it climbed past the band, falling when it dropped past it, steady
// otherwise. A rising warmth and a falling warmth at the SAME level render
// differently (the whole point of FR-004).
func TestNeedsTrajectoryDirections(t *testing.T) {
	s := knownPlacesState(t)
	a := &s.Agents[0]
	// Current levels; anchor set so each need moved a distinct, unambiguous way.
	a.Needs = sim.Needs{Health: 800, Food: 430, Rest: 620, Warmth: 450, Morale: 700}
	a.NeedsAnchor = &sim.Needs{
		Health: 800, // unchanged → steady
		Food:   600, // fell 170 → falling
		Rest:   400, // rose 220 → rising
		Warmth: 600, // fell 150 → falling
		Morale: 700, // unchanged → steady
	}
	a.NeedsAnchorTick = 1

	got := renderNeeds(s, 0)
	for _, want := range []string{
		"health 80 and steady",
		"food 43 and falling",
		"rest 62 and rising",
		"warmth 45 and falling",
		"morale 70 and steady",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("needs render missing %q:\n%s", want, got)
		}
	}

	// Same warmth LEVEL, rising anchor: the direction flips to rising — proof the
	// trajectory, not the level, carries the new signal.
	a.NeedsAnchor.Warmth = 300 // rose 150 → rising
	if got := renderNeeds(s, 0); !strings.Contains(got, "warmth 45 and rising") {
		t.Errorf("warmth at the same level should read rising off a lower anchor:\n%s", got)
	}
}

// TestNeedsTrajectorySteadyNoFlicker (T016, SC-003): a need that moved less than
// the deadband — noise, not a trend — never renders rising or falling, in either
// direction, right up to the band edge.
func TestNeedsTrajectorySteadyNoFlicker(t *testing.T) {
	s := knownPlacesState(t)
	a := &s.Agents[0]
	a.Needs = sim.Needs{Health: 500, Food: 500, Rest: 500, Warmth: 500, Morale: 500}
	a.NeedsAnchorTick = 1

	// Movements at exactly ±deadband and below stay steady; only strictly beyond
	// the band tips the direction.
	for _, tc := range []struct {
		anchorWarmth int
		wantSteady   bool
	}{
		{500 - trajectoryDeadband, true},      // +10, at the band → steady
		{500 + trajectoryDeadband, true},      // -10, at the band → steady
		{500 - trajectoryDeadband - 1, false}, // +11, past the band → rising
		{500 + trajectoryDeadband + 1, false}, // -11, past the band → falling
	} {
		a.NeedsAnchor = &sim.Needs{Health: 500, Food: 500, Rest: 500, Warmth: tc.anchorWarmth, Morale: 500}
		got := renderNeeds(s, 0)
		steady := strings.Contains(got, "warmth 50 and steady")
		if steady != tc.wantSteady {
			t.Errorf("anchor warmth %d: steady=%v, want %v:\n%s", tc.anchorWarmth, steady, tc.wantSteady, got)
		}
	}
}

// TestNeedsTrajectoryFirstWindowSteady (T016, edge case 1): with no anchor yet
// (nil / first window) every need renders steady — never a spurious direction
// off a missing history window.
func TestNeedsTrajectoryFirstWindowSteady(t *testing.T) {
	s := knownPlacesState(t)
	a := &s.Agents[0]
	a.Needs = sim.Needs{Health: 900, Food: 100, Rest: 800, Warmth: 200, Morale: 600}
	a.NeedsAnchor = nil
	a.NeedsAnchorTick = 0

	got := renderNeeds(s, 0)
	for _, need := range []string{"health", "food", "rest", "warmth", "morale"} {
		if strings.Contains(got, need+" ") && (strings.Contains(got, need+" 90 and rising") ||
			strings.Contains(got, need+" 10 and falling")) {
			t.Errorf("first window (nil anchor) rendered a direction for %s:\n%s", need, got)
		}
	}
	if strings.Count(got, "and steady") != 5 {
		t.Errorf("first window must render all five needs steady:\n%s", got)
	}
}

// TestSelfHistoryEmptyState (T012, US1 AS-3, edge case 1): a villager with no
// recorded intents renders an explicit "no prior activity" line — never a
// missing or malformed block.
func TestSelfHistoryEmptyState(t *testing.T) {
	s := knownPlacesState(t)
	got := renderSelfHistory(s, 0)
	if !strings.Contains(got, "no prior activity") {
		t.Errorf("empty self_history must be explicit, got %q", got)
	}
	// And it reaches the prompt.
	if !strings.Contains(userPrompt(s, 0, sim.WindowK, ""), "no prior activity") {
		t.Error("first-thought self_history line missing from assembled prompt")
	}
}

// TestSelfHistorySources (T012, SC-002, contract §Sources named honestly): each
// source renders in its own plain words; a reflex record invents no reason even
// when the record carries none; an unknown source renders "unknown", not a
// guess.
func TestSelfHistorySources(t *testing.T) {
	s := knownPlacesState(t)
	a := &s.Agents[0]
	a.IntentLog = append(a.IntentLog,
		sim.IntentRecord{Goal: "chop", Source: "planner", Reason: "need wood for a fire", Tick: 10, Outcome: "done", OutcomeTick: 20},
		sim.IntentRecord{Goal: "forage", Source: "reflex", Tick: 30},
		sim.IntentRecord{Goal: "attend_meeting", Source: "meeting", Tick: 40})
	got := renderSelfHistory(s, 0)

	// Newest-first ordering: the meeting (unknown) record leads.
	if !strings.HasPrefix(got, "Recently you:\n- attend_meeting") {
		t.Errorf("records not newest-first:\n%s", got)
	}
	// Planner → "you decided", with its reason quoted verbatim, outcome shown.
	if !strings.Contains(got, "chop — you decided this (\"need wood for a fire\"); completed") {
		t.Errorf("planner record wording wrong:\n%s", got)
	}
	// Reflex → "instinct", executing, and NO fabricated reason.
	if !strings.Contains(got, "forage — instinct drove this; still underway") {
		t.Errorf("reflex record wording wrong:\n%s", got)
	}
	if strings.Contains(got, "forage — instinct drove this (") {
		t.Errorf("reflex record fabricated a reason:\n%s", got)
	}
	// Unknown source → "unknown", never guessed at.
	if !strings.Contains(got, "attend_meeting — source unknown") {
		t.Errorf("unknown-source record wording wrong:\n%s", got)
	}
}

// TestSelfHistoryAlternation (T012, SC-002, FR-003): with ≥3 records the
// alternation between two goals is itself visible — order and per-record source
// preserved, newest first, capped at four shown.
func TestSelfHistoryAlternation(t *testing.T) {
	s := knownPlacesState(t)
	a := &s.Agents[0]
	// forage → goto_warmth → forage → goto_warmth → forage (oldest first).
	seq := []struct {
		goal, source string
	}{
		{"forage", "planner"},
		{"goto_warmth", "reflex"},
		{"forage", "planner"},
		{"goto_warmth", "reflex"},
		{"forage", "planner"},
	}
	for i, r := range seq {
		a.IntentLog = append(a.IntentLog, sim.IntentRecord{Goal: r.goal, Source: r.source, Tick: int64(10 * i), Outcome: "done", OutcomeTick: int64(10*i + 5)})
	}
	got := renderSelfHistory(s, 0)

	// Cap at four shown (contract block 3), newest first.
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 1+4 { // header + four records
		t.Fatalf("self_history should show a header + 4 records, got %d lines:\n%s", len(lines), got)
	}
	// The alternation is perceivable: within the shown window both goals appear,
	// interleaved (newest is forage, then goto_warmth, then forage, then
	// goto_warmth — the fifth/oldest forage is off the window).
	body := lines[1:]
	wantGoals := []string{"forage", "goto_warmth", "forage", "goto_warmth"}
	for i, w := range wantGoals {
		if !strings.HasPrefix(body[i], "- "+w+" ") {
			t.Errorf("record %d = %q, want goal %q (newest-first alternation)", i, body[i], w)
		}
	}
}

// TestSelfHistoryOutcomes (T012): every terminal outcome renders a distinct,
// honest clause; the open (executing) record reads as still underway.
func TestSelfHistoryOutcomes(t *testing.T) {
	cases := []struct {
		outcome, want string
	}{
		{"", "still underway"},
		{"done", "completed"},
		{"failed", "it failed"},
		{"rejected", "rejected before it began"},
		{"expired", "the plan expired"},
	}
	for _, c := range cases {
		line := selfHistoryLine(sim.IntentRecord{Goal: "chop", Source: "planner", Outcome: c.outcome})
		if !strings.Contains(line, c.want) {
			t.Errorf("outcome %q rendered %q, want it to contain %q", c.outcome, line, c.want)
		}
	}
}

// TestPlanEchoContent (T018, US3 AS-1, FR-005): an active multi-step plan echoes
// its remaining steps in order — the head marked "next", the rest "then" — with
// each step's guard and validity deadline in plain words (never the raw guard
// predicate name), and reaches the assembled prompt.
func TestPlanEchoContent(t *testing.T) {
	s := knownPlacesState(t)
	a := &s.Agents[0]
	until := clock.TickAt(2, 8, 30, 0)
	a.Plan = []sim.PlanStep{
		{Job: "warm-up", Goal: "goto_warmth", Until: until,
			When: &sim.Guard{Type: sim.GuardTargetPresent, Target: 1}},
		{Job: "warm-up", Goal: "rest"},
	}

	got := renderPlanEcho(s, 0)
	if !strings.Contains(got, "- next: goto_warmth") {
		t.Errorf("head step not marked next:\n%s", got)
	}
	if !strings.Contains(got, "- then: rest") {
		t.Errorf("second step not marked then:\n%s", got)
	}
	// Guard rendered in plain words, naming the target villager (agent 1 = Birch),
	// not the raw "target_present" predicate.
	if !strings.Contains(got, "while "+s.Agents[1].Name+" is still nearby") {
		t.Errorf("guard not rendered in plain words:\n%s", got)
	}
	if strings.Contains(got, sim.GuardTargetPresent) {
		t.Errorf("raw guard predicate leaked into the echo:\n%s", got)
	}
	// Validity deadline on the game clock.
	if !strings.Contains(got, "valid until "+clock.Format(until)) {
		t.Errorf("until deadline missing:\n%s", got)
	}
	// And it reaches the prompt.
	if !strings.Contains(userPrompt(s, 0, sim.WindowK, ""), "Your plan (remaining steps") {
		t.Error("plan echo missing from the assembled prompt")
	}
}

// TestPlanEchoAllGuardsPlainWords (T018): every guard in the closed vocabulary
// renders human-readably — no raw predicate name ever reaches the prompt.
func TestPlanEchoAllGuardsPlainWords(t *testing.T) {
	s := knownPlacesState(t)
	for _, g := range []struct {
		typ, want string
	}{
		{sim.GuardTargetAlive, "is alive"},
		{sim.GuardTargetPresent, "is still nearby"},
		{sim.GuardNotSuperseded, "unless something more urgent"},
		{sim.GuardAfterTick, "not before"},
		{sim.GuardBeforeTick, "only before"},
	} {
		s.Agents[0].Plan = []sim.PlanStep{{Goal: "chop", When: &sim.Guard{Type: g.typ, Target: 1, Tick: 500}}}
		got := renderPlanEcho(s, 0)
		if !strings.Contains(got, g.want) {
			t.Errorf("guard %q rendered %q, want it to contain %q", g.typ, got, g.want)
		}
		if strings.Contains(got, g.typ) {
			t.Errorf("raw guard predicate %q leaked into the echo:\n%s", g.typ, got)
		}
	}
}

// TestPlanEchoOmittedNoPlan (T018, FR-005 "no stale echo"): a villager with no
// active plan renders no plan_echo block — omitted entirely, not an empty header.
func TestPlanEchoOmittedNoPlan(t *testing.T) {
	s := knownPlacesState(t)
	if got := renderPlanEcho(s, 0); got != "" {
		t.Errorf("no-plan villager rendered a plan echo: %q", got)
	}
	if strings.Contains(userPrompt(s, 0, sim.WindowK, ""), "Your plan") {
		t.Error("plan echo present in the prompt of a planless villager")
	}
}

// TestPlanEchoClearedAfterExpiry (T018, US3 AS-2, FR-005): once the plan is
// cleared by the agent.plan_expired reducer arm the echo disappears (no stale
// plan), AND the plan's end is visible in self_history at the next thought — the
// US1 ring stamped the expired step. Driven through the real reducer.
func TestPlanEchoClearedAfterExpiry(t *testing.T) {
	s := knownPlacesState(t)

	// A plan is set and its head step fires (source "plan"), then the step
	// expires — exactly the lifecycle the reducer maintains.
	apply := func(tick int64, typ string, p any) {
		t.Helper()
		if err := s.Apply(store.Event{Tick: tick, Type: typ, Payload: mustJSON(t, p)}); err != nil {
			t.Fatalf("apply %s: %v", typ, err)
		}
	}
	apply(1000, "agent.plan_set", sim.PlanSetPayload{Agent: 0, Job: "warm-up",
		Steps: []sim.PlanStep{{Job: "warm-up", Goal: "goto_warmth", Until: 2000}}})
	apply(1000, "agent.intent_set", sim.IntentSetPayload{Agent: sim.Ref(0), Goal: "goto_warmth", Source: "plan"})

	// While the plan stands the echo is present.
	if got := renderPlanEcho(s, 0); !strings.Contains(got, "next: goto_warmth") {
		t.Fatalf("active plan not echoed before expiry:\n%s", got)
	}

	apply(2000, "agent.plan_expired", sim.PlanStepPayload{Agent: 0, Job: "warm-up", Step: "goto_warmth", Reason: "window closed"})

	// No stale echo after the plan cleared.
	if got := renderPlanEcho(s, 0); got != "" {
		t.Errorf("stale plan echo after expiry: %q", got)
	}
	// The plan's end is visible in self_history at the next thought (US1 ring).
	sh := renderSelfHistory(s, 0)
	if !strings.Contains(sh, "goto_warmth") || !strings.Contains(sh, "the plan expired") {
		t.Errorf("plan end not visible in self_history:\n%s", sh)
	}
}

// TestPlanEchoClearedAfterCompletion (T018, FR-005): a plan whose last step has
// been consumed (agent.plan_step_started popped it) is empty, so no stale echo
// remains after the plan runs to completion.
func TestPlanEchoClearedAfterCompletion(t *testing.T) {
	s := knownPlacesState(t)
	apply := func(tick int64, typ string, p any) {
		t.Helper()
		if err := s.Apply(store.Event{Tick: tick, Type: typ, Payload: mustJSON(t, p)}); err != nil {
			t.Fatalf("apply %s: %v", typ, err)
		}
	}
	apply(1000, "agent.plan_set", sim.PlanSetPayload{Agent: 0, Job: "j",
		Steps: []sim.PlanStep{{Job: "j", Goal: "forage"}}})
	if renderPlanEcho(s, 0) == "" {
		t.Fatal("one-step plan not echoed before it ran")
	}
	// The single step starts and is popped — the plan is now complete/empty.
	apply(1010, "agent.plan_step_started", sim.PlanStepPayload{Agent: 0, Job: "j", Step: "forage"})
	if got := renderPlanEcho(s, 0); got != "" {
		t.Errorf("stale plan echo after the plan completed: %q", got)
	}
}

// --- spec 043 US4 (T019-T023): journal block, memory floor, budget ---

// ladderState extends richContextState so every droppable tier is present: >K
// memories (⇒ a serendipity tail + above-floor entries), a journal entry that
// matches a worst-need term ("warmth"), and a standing plan (plan_echo). The
// full drop ladder can then be exercised end to end.
func ladderState(t *testing.T) *sim.State {
	t.Helper()
	s := richContextState(t) // 4 memories, known_places, social_law (debt)
	a := &s.Agents[0]
	for i := int64(0); i < 20; i++ {
		a.Memories = append(a.Memories, sim.Memory{
			Text:     fmt.Sprintf("Older memory %d about the woods.", i),
			Salience: 3, Tick: 200 + i, Subject: -1, Seq: 100 + i,
		})
	}
	a.Journal = &sim.Journal{NextID: 1, Entries: []sim.JournalEntry{
		{ID: 0, Tick: 50, Text: "I banked the fire to keep my warmth up through the cold night."},
	}}
	a.Plan = []sim.PlanStep{{Job: "j", Goal: "rest"}}
	return s
}

// TestContextDropOrderFullLadder (T021, FR-008, contract §Budget + research R7):
// with every tier present, a zero budget sheds them in the exact documented
// order — journal (1), memories_serendipity (2), memories above-floor (3),
// social_law (4), known_places (5), plan_echo (6) — while the survival blocks
// AND the memories floor never drop.
func TestContextDropOrderFullLadder(t *testing.T) {
	s := ladderState(t)

	// Sanity: every droppable tier is actually present at an ample budget.
	full := assembleContext(s, 0, sim.WindowK, "", "")
	for _, name := range []string{"memories", "memories_serendipity", "journal", "social_law", "known_places", "plan_echo"} {
		if _, ok := full.blockBytes[name]; !ok {
			t.Fatalf("ladder fixture missing droppable tier %q — the ladder cannot bite: %v", name, full.blockBytes)
		}
	}

	starved := assembleBudget(s, 0, sim.WindowK, "", "", 0)
	want := []string{"journal", "memories_serendipity", "memories", "social_law", "known_places", "plan_echo"}
	if strings.Join(starved.droppedBlocks, ",") != strings.Join(want, ",") {
		t.Errorf("drop ladder = %v\nwant                %v", starved.droppedBlocks, want)
	}
	// The memories floor survives (block still present) and the serendipity /
	// journal tiers are gone from the accounting.
	for _, survive := range []string{"frame", "needs", "self_history", "inventory", "memories"} {
		if _, ok := starved.blockBytes[survive]; !ok {
			t.Errorf("protected block %q dropped under zero budget: %v", survive, starved.blockBytes)
		}
	}
	for _, gone := range []string{"memories_serendipity", "journal", "social_law", "known_places", "plan_echo"} {
		if _, ok := starved.blockBytes[gone]; ok {
			t.Errorf("block %q should be dropped but is still accounted: %v", gone, starved.blockBytes)
		}
	}
	// The surviving memories block is exactly the floor (memoryFloor lines).
	if lines := strings.Count(starved.text, "\n- "); lines < 1 {
		t.Errorf("memories floor did not render any lines:\n%s", starved.text)
	}
}

// TestMemoriesFloorProtectedByteAccounting (T021, FR-010): the per-block bytes
// plus the fixed closer account for every assembled byte even after the
// serendipity + above-floor tiers are partially dropped — the split is honest
// accounting, not a byte leak.
func TestMemoriesFloorProtectedByteAccounting(t *testing.T) {
	s := ladderState(t)
	for _, budget := range []int{0, 50, 200, contextBudgetTokens} {
		asm := assembleBudget(s, 0, sim.WindowK, "", "", budget)
		sum := len(promptCloser)
		for _, sz := range asm.blockBytes {
			sum += sz
		}
		if sum != asm.promptBytes {
			t.Errorf("budget %d: block bytes + closer = %d, want promptBytes %d (blocks=%v)", budget, sum, asm.promptBytes, asm.blockBytes)
		}
	}
}

// TestContextByteIdenticalMemorySplit (T021, contract §Determinism / task
// constraint): with nothing dropped, splitting the memory window into
// `memories` + `memories_serendipity` is an accounting change only — the
// assembled text is byte-identical to a single interleaved "You remember:" list
// (the pre-043 rendering), for a >K soul that actually carries a serendipity
// tail.
func TestContextByteIdenticalMemorySplit(t *testing.T) {
	s := ladderState(t)
	s.Agents[0].Journal = nil // isolate the memory chunk
	s.Agents[0].Plan = nil

	asm := assembleContext(s, 0, sim.WindowK, "", "")
	// Reconstruct the pre-043 single-list rendering from the same selection.
	window := selectWindow(s, 0, sim.WindowK, s.Tick, "")
	var want strings.Builder
	want.WriteString("\nYou remember:\n")
	for _, m := range window {
		fmt.Fprintf(&want, "- %s\n", sim.FormatMemory(m))
	}
	if !strings.Contains(asm.text, want.String()) {
		t.Errorf("memory chunk is not byte-identical to the interleaved single list:\nassembled: %q\nwant sub:  %q", asm.text, want.String())
	}
}

// TestMemoriesDegradedModePassthrough (T021, US4 AS-4, FR-006): "on" mode with
// NO recorded situation vector falls back to legacy selection — the memories
// block still renders, and the assembled prompt is byte-identical to "" mode.
// Nothing crashes.
func TestMemoriesDegradedModePassthrough(t *testing.T) {
	s := richContextState(t) // no SitVec recorded
	on := assembleContext(s, 0, sim.WindowK, "on", "")
	if _, ok := on.blockBytes["memories"]; !ok {
		t.Error("degraded 'on' mode dropped the memories block")
	}
	if !strings.Contains(on.text, "You remember:") {
		t.Error("memories block did not render in degraded mode")
	}
	legacy := assembleContext(s, 0, sim.WindowK, "", "")
	if on.text != legacy.text {
		t.Errorf("degraded 'on' diverged from legacy selection:\non:     %q\nlegacy: %q", on.text, legacy.text)
	}
}

// TestJournalBlockRendersAndOmits (T020, FR-007): the journal block includes a
// bounded excerpt of a situationally-relevant entry (matched on a worst-need
// term), omits irrelevant entries, and is absent entirely when nothing matches
// or there is no journal. It reaches the assembled prompt.
func TestJournalBlockRendersAndOmits(t *testing.T) {
	s := knownPlacesState(t)
	a := &s.Agents[0]
	a.Needs = sim.Needs{Health: 800, Food: 200, Rest: 700, Warmth: 300, Morale: 900} // worst: food, warmth

	// No journal → omitted.
	if got := renderJournal(s, 0); got != "" {
		t.Errorf("no-journal villager rendered a journal block: %q", got)
	}

	a.Journal = &sim.Journal{NextID: 2, Entries: []sim.JournalEntry{
		{ID: 0, Tick: 10, Text: "Foraged berries near the river; my food stores are low."},
		{ID: 1, Tick: 20, Text: "A quiet day mending my spear."},
	}}
	got := renderJournal(s, 0)
	if !strings.Contains(got, "From your journal:") {
		t.Errorf("journal header missing:\n%s", got)
	}
	if !strings.Contains(got, "(#0)") || !strings.Contains(got, "food stores are low") {
		t.Errorf("relevant journal entry (food) missing:\n%s", got)
	}
	if strings.Contains(got, "mending my spear") {
		t.Errorf("irrelevant journal entry included:\n%s", got)
	}
	if !strings.Contains(userPrompt(s, 0, sim.WindowK, ""), "From your journal:") {
		t.Error("journal block missing from assembled prompt")
	}

	// No matching entry → omitted.
	a.Journal = &sim.Journal{NextID: 1, Entries: []sim.JournalEntry{{ID: 0, Tick: 10, Text: "A calm, unremarkable morning."}}}
	if got := renderJournal(s, 0); got != "" {
		t.Errorf("no-match journal rendered a block: %q", got)
	}
}

// TestPlantedMemoryRelevance (T022, SC-006, US4 AS-1): an agent carrying a
// situationally-relevant memory (vector matching its situation vector) buried
// under louder, irrelevant high-salience memories has the relevant memory in the
// assembled window in at least 80% of thoughts across seeds — within budget —
// while the legacy window would exclude it. Vectors are constructed directly, no
// live embedder.
func TestPlantedMemoryRelevance(t *testing.T) {
	const marker = "Robbed at the river."
	build := func(seed uint64) *sim.State {
		s := sim.NewState(seed, worldmap.Generate(seed, 64, 64))
		s.Tick = 130_000
		a := &s.Agents[0]
		// One old, low-salience, situation-matching memory…
		a.Memories = append(a.Memories, sim.Memory{Text: marker, Salience: 2, Tick: 100, Subject: -1, Seq: 1, Vec: []float32{1, 0}, VecModel: "all-minilm"})
		// …buried under many newer, unrelated ones. Salience 5 is the reference
		// gap the equal-weight relevance term (+0.5 max) is tuned to overcome
		// (contracts/relevance-scoring.md §1): a perfect match beats it, a louder
		// salience-8 crowd would not — the point is relevance, not salience.
		for i := int64(1); i <= 40; i++ {
			a.Memories = append(a.Memories, sim.Memory{Text: "Routine day.", Salience: 5, Tick: 100_000 + i*600, Subject: -1, Seq: 1 + i, Vec: []float32{0, 1}, VecModel: "all-minilm"})
		}
		a.SitVec = []float32{1, 0}
		a.SitVecModel = "all-minilm"
		a.SitVecTick = 129_000
		return s
	}

	seeds := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	hits := 0
	for _, seed := range seeds {
		s := build(seed)
		asm := assembleContext(s, 0, sim.WindowK, "on", "")
		if tok := approxTokens(asm.promptBytes); tok > contextBudgetTokens {
			t.Errorf("seed %d: assembled context %d approx-tokens exceeds budget %d", seed, tok, contextBudgetTokens)
		}
		if strings.Contains(asm.text, marker) {
			hits++
		}
	}
	if frac := float64(hits) / float64(len(seeds)); frac < 0.8 {
		t.Errorf("relevant memory present in %d/%d assembled contexts (%.0f%%), want ≥80%%", hits, len(seeds), frac*100)
	}
}

// syntheticState builds a deterministic, varied villager world for the budget
// aggregate: per-agent randomized needs, memory counts, an optional journal, an
// optional plan, and a recent intent — a hermetic stand-in for the diverse
// states a running world produces.
func syntheticState(t *testing.T, seed uint64) *sim.State {
	t.Helper()
	s := sim.NewState(seed, worldmap.Generate(seed, 64, 64))
	s.Tick = int64(500_000 + seed*137)
	r := seed + 1
	next := func(mod uint64) int {
		r = r*6364136223846793005 + 1442695040888963407
		return int((r >> 33) % mod)
	}
	for i := range s.Agents {
		a := &s.Agents[i]
		a.Needs = sim.Needs{Health: next(1000), Food: next(1000), Rest: next(1000), Warmth: next(1000), Morale: next(1000)}
		for m := 0; m < next(30); m++ {
			a.Memories = append(a.Memories, sim.Memory{
				Text:     fmt.Sprintf("Memory %d for agent %d in a long-remembered place.", m, i),
				Salience: 1 + next(10), Tick: s.Tick - int64(next(400_000)), Subject: -1, Seq: int64(m + 1),
			})
		}
		if next(2) == 0 {
			a.Journal = &sim.Journal{NextID: 1, Entries: []sim.JournalEntry{
				{ID: 0, Tick: 10, Text: "Kept my warmth and food steady through a long, restful day."},
			}}
		}
		if next(3) == 0 {
			a.Plan = []sim.PlanStep{{Job: "j", Goal: "forage"}, {Job: "j", Goal: "rest"}}
		}
		a.IntentLog = append(a.IntentLog, sim.IntentRecord{Goal: "forage", Source: "reflex", Tick: s.Tick - 100})
	}
	return s
}

// TestBudgetFitAggregate (T023, SC-005, FR-010): aggregating PromptBytes /
// DroppedBlocks across many assembled thoughts over varied synthetic states, at
// least 99% land within the configured budget, and every overflow records its
// dropped blocks. This is the hermetic half of SC-005; the live multi-day
// scratch-world measurement remains for the orchestrator (see the T023 report).
func TestBudgetFitAggregate(t *testing.T) {
	total, within := 0, 0
	for seed := uint64(0); seed < 200; seed++ {
		s := syntheticState(t, seed)
		for i := range s.Agents {
			asm := assembleContext(s, i, sim.WindowK, "on", "")
			total++
			if approxTokens(asm.promptBytes) <= contextBudgetTokens {
				within++
			} else if len(asm.droppedBlocks) == 0 {
				t.Errorf("seed %d agent %d over budget (%d approx-tokens) but recorded no drops", seed, i, approxTokens(asm.promptBytes))
			}
		}
	}
	if frac := float64(within) / float64(total); frac < 0.99 {
		t.Errorf("within-budget fraction %.4f (%d/%d), want ≥0.99", frac, within, total)
	}
	if total == 0 {
		t.Fatal("no thoughts assembled")
	}
}

// --- spec 084: the directive block (FR-011) ---

// directiveState builds a fixture with an active designation + directive
// addressing agent 0 (and one addressing only agent 1, to prove scoping).
func directiveState(t *testing.T) *sim.State {
	t.Helper()
	s := richContextState(t)
	s.Designations = []sim.Designation{
		{ID: "dsg-100-0", Kind: sim.DesignationStructureSite, X: 4, Y: 5, X2: 4, Y2: 5,
			StructureKind: "shelter", PlacedTick: 100, Status: "active"},
		{ID: "dsg-100-1", Kind: sim.DesignationSettlementZone, X: 1, Y: 1, X2: 6, Y2: 6,
			MinStructures: 3, PlacedTick: 100, Status: "active"},
	}
	s.Directives = []sim.Directive{
		{ID: "dir-200-0", DesignationID: "dsg-100-0", Targets: []int{0, 1},
			Text: "Raise the shelter I have marked.", IssuedTick: 200,
			ExpiresTick: s.Tick + 3*24*3600, Status: "active"},
		{ID: "dir-200-1", DesignationID: "dsg-100-1", Targets: []int{1},
			Text: "Settle the valley.", IssuedTick: 200,
			ExpiresTick: s.Tick + 2*24*3600, Status: "active"},
	}
	return s
}

// TestDirectiveBlockRenders (US4 AS4): the block carries the guardian's text
// verbatim, the designation's kind + site, the fulfillment requirement, and
// plain-words time remaining — for the ADDRESSED agent only, in contract
// position between plan_echo and known_places, at neverDrop priority.
func TestDirectiveBlockRenders(t *testing.T) {
	s := directiveState(t)
	got := renderDirective(s, 0)
	for _, want := range []string{
		"The Guardian has charged you",
		`"Raise the shelter I have marked."`,
		"a shelter must stand at (4,5)",
		"about 3 days remain",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("block missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "Settle the valley") {
		t.Error("agent 0's block renders a directive addressing only agent 1")
	}
	// Agent 1 sees both, oldest-first by (IssuedTick, id).
	got1 := renderDirective(s, 1)
	if !strings.Contains(got1, "Raise the shelter") || !strings.Contains(got1, "Settle the valley") {
		t.Errorf("agent 1's block missing a directive:\n%s", got1)
	}
	if strings.Index(got1, "Raise the shelter") > strings.Index(got1, "Settle the valley") {
		t.Errorf("directives not oldest-first:\n%s", got1)
	}

	// Contract position + neverDrop: assembled under a zero budget, the block
	// survives while every droppable neighbor sheds.
	asm := assembleBudget(s, 0, sim.WindowK, "", "", 0)
	if _, ok := asm.blockBytes["directive"]; !ok {
		t.Errorf("directive block dropped under budget pressure (must be neverDrop): %v", asm.blockBytes)
	}
	full := assembleContext(s, 0, sim.WindowK, "", "")
	planIdx := strings.Index(full.text, "Your plan")
	dirIdx := strings.Index(full.text, "The Guardian has charged you")
	placesIdx := strings.Index(full.text, "Places you know")
	if planIdx >= 0 && !(planIdx < dirIdx) || placesIdx >= 0 && !(dirIdx < placesIdx) {
		t.Errorf("directive block out of contract position (plan %d, directive %d, places %d)", planIdx, dirIdx, placesIdx)
	}
}

// TestDirectiveBlockScopes: expired/cancelled directives and orphans (bound
// designation non-active) render nothing; the cap holds at 2 oldest.
func TestDirectiveBlockScopes(t *testing.T) {
	s := directiveState(t)
	// Orphan: cancel the designation — the block (like the rung) skips it.
	s.Designations[0].Status = "cancelled"
	if got := renderDirective(s, 0); got != "" {
		t.Errorf("orphaned directive rendered: %q", got)
	}
	s.Designations[0].Status = "active"
	// Non-active directive renders nothing.
	s.Directives[0].Status = "expired"
	if got := renderDirective(s, 0); got != "" {
		t.Errorf("expired directive rendered: %q", got)
	}
	s.Directives[0].Status = "active"
	// Cap 2: a third directive addressing agent 1 stays unrendered.
	s.Directives = append(s.Directives, sim.Directive{
		ID: "dir-300-0", DesignationID: "dsg-100-0", Targets: []int{1},
		Text: "A third charge.", IssuedTick: 300, ExpiresTick: s.Tick + 24*3600, Status: "active"})
	if got := renderDirective(s, 1); strings.Contains(got, "A third charge") {
		t.Errorf("cap 2 not applied:\n%s", got)
	}
}

// TestDirectiveFreeByteIdentity (SC-006): a directive-free world's assembled
// prompt is byte-identical to the same world with the block registry entry
// present — the empty block renders "" and is omitted entirely, and the
// blockBytes accounting never mentions it.
func TestDirectiveFreeByteIdentity(t *testing.T) {
	s := richContextState(t)
	asm := assembleContext(s, 0, sim.WindowK, "", "")
	if _, ok := asm.blockBytes["directive"]; ok {
		t.Error("directive-free world accounts a directive block")
	}
	if strings.Contains(asm.text, "The Guardian has charged you") {
		t.Error("directive-free prompt carries directive text")
	}
	// And the known-places designation landmark line is likewise absent.
	if strings.Contains(asm.text, "Guardian marked") {
		t.Error("directive-free prompt carries a designation landmark")
	}
}

// TestDesignationLandmarkInKnownPlaces (FR-006): a granted designation fact
// renders as an individually-named landmark, enriched from the entity in
// state, with the honest generic fallback once the entity is gone.
func TestDesignationLandmarkInKnownPlaces(t *testing.T) {
	s := directiveState(t)
	a := &s.Agents[0]
	addFact(a.Map, sim.PlaceFact{Kind: "designation", X: 4, Y: 5, Seen: s.Tick,
		Provenance: sim.ProvenanceRevealed})
	got := knownPlaces(s, 0)
	if !strings.Contains(got, `a shelter site the Guardian marked at (4,5)`) {
		t.Errorf("landmark missing/wrong:\n%s", got)
	}
	// Entity pruned away: the generic remembered-history line.
	s.Designations = nil
	got = knownPlaces(s, 0)
	if !strings.Contains(got, "a place the Guardian once marked at (4,5)") {
		t.Errorf("fallback landmark missing:\n%s", got)
	}
}
