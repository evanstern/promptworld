package mind

import (
	"fmt"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/sim"
)

// legacyUserPrompt is a FROZEN byte-for-byte copy of userPrompt as it stood
// before spec 043 wrapped its content into the block assembler (context.go).
// The golden-equality test below asserts the refactor changed no legacy bytes:
// the assembled prompt, with only the NEW self_history block removed, must
// equal this reference exactly. Keeping the reference here (not deleting the
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

// TestContextGoldenIdentity (T003, FR-009): wrapping userPrompt content into the
// block assembler changed NO legacy bytes. The production prompt, with only the
// new self_history block removed, is byte-identical to the frozen pre-043
// rendering — for both the first-thought (empty IntentLog) and the populated
// cases, and across memory modes.
func TestContextGoldenIdentity(t *testing.T) {
	for _, mode := range []string{"", "shadow", "on"} {
		t.Run("mode="+mode, func(t *testing.T) {
			s := richContextState(t)

			// First-thought agent: self_history is the "no prior activity" line.
			got := userPrompt(s, 0, sim.WindowK, mode)
			sh := renderSelfHistory(s, 0)
			if sh == "" {
				t.Fatal("self_history rendered empty — it must always render (first-thought line)")
			}
			if stripped := strings.Replace(got, sh, "", 1); stripped != legacyUserPrompt(s, 0, sim.WindowK, mode) {
				t.Errorf("assembled prompt (minus self_history) diverged from legacy bytes:\ngot:    %q\nlegacy: %q", stripped, legacyUserPrompt(s, 0, sim.WindowK, mode))
			}

			// Populated IntentLog: same invariant, self_history now a real log.
			s.Agents[0].IntentLog = append(s.Agents[0].IntentLog, sim.IntentRecord{Goal: "forage", Source: "reflex", Tick: 90})
			got = userPrompt(s, 0, sim.WindowK, mode)
			sh = renderSelfHistory(s, 0)
			if stripped := strings.Replace(got, sh, "", 1); stripped != legacyUserPrompt(s, 0, sim.WindowK, mode) {
				t.Errorf("with populated IntentLog, prompt (minus self_history) diverged from legacy:\ngot:    %q\nlegacy: %q", stripped, legacyUserPrompt(s, 0, sim.WindowK, mode))
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

// TestContextDropOrder (T005, FR-008, contract §Budget): under budget pressure
// whole blocks drop lowest-priority-first — journal/serendipity/plan_echo are
// absent this slice, so the droppable order is memories → social_law →
// known_places. A budget of 0 sheds all three, in that order; survival blocks
// (frame, needs, self_history, inventory) are never dropped, at any budget.
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

	// Zero budget: every droppable block sheds, in ascending-priority order;
	// survival blocks remain.
	starved := assembleBudget(s, 0, sim.WindowK, "", "", 0)
	wantDropped := []string{"memories", "social_law", "known_places"}
	if strings.Join(starved.droppedBlocks, ",") != strings.Join(wantDropped, ",") {
		t.Errorf("drop order = %v, want %v", starved.droppedBlocks, wantDropped)
	}
	for _, survive := range []string{"frame", "needs", "self_history", "inventory"} {
		if _, ok := starved.blockBytes[survive]; !ok {
			t.Errorf("survival block %q dropped under zero budget: blocks=%v dropped=%v", survive, starved.blockBytes, starved.droppedBlocks)
		}
	}
	for _, d := range starved.droppedBlocks {
		if d == "frame" || d == "needs" || d == "self_history" || d == "inventory" {
			t.Errorf("survival block %q appears in dropped set", d)
		}
	}

	// A budget just below the full size drops ONLY the lowest-priority block
	// first (memories), never a higher-priority one before it.
	fullTokens := approxTokens(full.promptBytes)
	oneDrop := assembleBudget(s, 0, sim.WindowK, "", "", fullTokens-1)
	if len(oneDrop.droppedBlocks) == 0 || oneDrop.droppedBlocks[0] != "memories" {
		t.Errorf("first block dropped = %v, want memories first", oneDrop.droppedBlocks)
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
