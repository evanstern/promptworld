package mind

import (
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// Spec 110: the harvest ledger and its classifier (Phase 1, T005/T006), and
// the coalesced narration they feed (Phase 2/3, T007–T014).
//
// Phase 1's TestChronicleNoteCorrectionLineInertPhase1 asserted the ledger was
// inert — an attributed correction still emitted its own per-event line. Phase
// 2 is precisely the inversion of that assertion, so that test is superseded
// by TestChronicleNoteAttributedCorrectionEmitsNoLine below; its second half
// (an unexplained correction keeps its line, byte-identical) carries over
// unchanged, because FR-004 says it must.

// TestHarvestLedgerPopulatedFromAbsorb (T003/T005, FR-001): the existing
// agent.chopped / agent.quarried absorb arm feeds the ledger, and the spec-081
// witness re-arm that shares the arm still fires.
func TestHarvestLedgerPopulatedFromAbsorb(t *testing.T) {
	state := sim.NewState(42, worldmap.Generate(42, 64, 64))
	state.Paused = false
	state.Tick = 5000
	state.Agents[0].X, state.Agents[0].Y = 20, 20
	state.Agents[1].X, state.Agents[1].Y = 21, 20
	state.Agents[1].Intent = &sim.Intent{Goal: "chop", TargetX: 20, TargetY: 20}
	md := &Mind{replica: state}

	md.absorb([]store.Event{
		{Tick: 5000, Seq: 7, Type: "agent.chopped",
			Payload: mustJSON(t, sim.HarvestPayload{Agent: sim.Ref(0), X: 20, Y: 20})},
		{Tick: 5100, Seq: 8, Type: "agent.quarried",
			Payload: mustJSON(t, sim.HarvestPayload{Agent: sim.Ref(3), X: 40, Y: 41})},
	})

	if who, ok := md.attributedHarvest(20, 20, 5000); !ok || who != 0 {
		t.Errorf("chop at (20,20): got (%d,%v), want (0,true)", who, ok)
	}
	if who, ok := md.attributedHarvest(40, 41, 5100); !ok || who != 3 {
		t.Errorf("quarry at (40,41): got (%d,%v), want (3,true)", who, ok)
	}
	// The arm's pre-existing behaviour is undisturbed.
	if !md.pending[0] {
		t.Error("harvester was not re-armed by its own chop")
	}
	if !md.pending[1] {
		t.Error("spec-081 witness re-arm no longer fires")
	}
}

// TestHarvestLedgerAgeEvictionAtWindowEdge (T002/T005, FR-001): the window is
// inclusive at its edge and evicts oldest-first past it — both on lookup (a
// stale entry never matches) and on the ledger's own eviction pass.
func TestHarvestLedgerAgeEvictionAtWindowEdge(t *testing.T) {
	var l harvestLedger
	l.record(3, 4, 2, 1000)

	if _, ok := l.lookup(3, 4, 1000+harvestLedgerWindow); !ok {
		t.Error("harvest exactly harvestLedgerWindow old must still explain a correction")
	}
	if _, ok := l.lookup(3, 4, 1000+harvestLedgerWindow+1); ok {
		t.Error("harvest one tick past the window must not explain a correction")
	}

	// A harvest at the window edge leaves the older entry in place...
	l.record(9, 9, 5, 1000+harvestLedgerWindow)
	if len(l.at) != 2 {
		t.Fatalf("entries = %d, want 2 (edge is inclusive)", len(l.at))
	}
	// ...and one tick past it evicts the older entry outright.
	l.record(9, 8, 5, 1000+harvestLedgerWindow+1)
	if _, present := l.at[harvestKey{3, 4}]; present {
		t.Error("entry older than the window was not evicted")
	}
	if _, present := l.at[harvestKey{9, 9}]; !present {
		t.Error("in-window entry was wrongly evicted")
	}
}

// TestHarvestLedgerCapEviction (T002/T005, FR-001): the hard entry cap bounds
// the ledger on a long run, dropping oldest-first.
func TestHarvestLedgerCapEviction(t *testing.T) {
	var l harvestLedger
	const extra = 5
	for i := 0; i < harvestLedgerCap+extra; i++ {
		// Distinct coordinates, strictly increasing ticks, all inside the
		// window so only the cap can evict.
		l.record(i%1000, i/1000, i%sim.AgentCount, int64(i))
	}
	if len(l.at) != harvestLedgerCap {
		t.Fatalf("entries = %d, want the cap %d", len(l.at), harvestLedgerCap)
	}
	// The `extra` oldest records are gone; the newest survives.
	for i := 0; i < extra; i++ {
		if _, present := l.at[harvestKey{i % 1000, i / 1000}]; present {
			t.Errorf("oldest entry %d survived cap eviction", i)
		}
	}
	last := harvestLedgerCap + extra - 1
	if _, present := l.at[harvestKey{last % 1000, last / 1000}]; !present {
		t.Error("newest entry was evicted by the cap")
	}
}

// TestAttributedHarvestExactCoordinate (T004/T005, FR-002): attribution is by
// exact tile. A neighbouring harvest explains nothing — the failure direction
// D4 chose deliberately, since a false attribution would silence a real
// mystery.
func TestAttributedHarvestExactCoordinate(t *testing.T) {
	md := &Mind{}
	md.harvests.record(12, 7, 4, 2000)

	if who, ok := md.attributedHarvest(12, 7, 2500); !ok || who != 4 {
		t.Errorf("exact hit: got (%d,%v), want (4,true)", who, ok)
	}
	for _, c := range [][2]int{{11, 7}, {13, 7}, {12, 6}, {12, 8}} {
		if _, ok := md.attributedHarvest(c[0], c[1], 2500); ok {
			t.Errorf("(%d,%d) must not be attributed to a harvest at (12,7)", c[0], c[1])
		}
	}
}

// TestAttributedHarvestMiss (T004/T005, FR-002 / SC-003): coordinates with no
// harvest in the ledger classify as unexplained — the genuine-anomaly path.
func TestAttributedHarvestMiss(t *testing.T) {
	md := &Mind{}
	if who, ok := md.attributedHarvest(1, 1, 100); ok || who != 0 {
		t.Errorf("empty ledger: got (%d,%v), want (0,false)", who, ok)
	}
	md.harvests.record(1, 1, 6, 100)
	if _, ok := md.attributedHarvest(1, 1, 100+harvestLedgerWindow+1); ok {
		t.Error("a harvest outside the window must classify the correction as unexplained")
	}
}

// TestChronicleNoteAttributedCorrectionEmitsNoLine (T008, FR-003/FR-004):
// supersedes Phase 1's inert test. A correction the ledger explains
// contributes NO line of its own and folds into the chapter tally; an
// unexplained one keeps its existing per-event line, byte-identical.
func TestChronicleNoteAttributedCorrectionEmitsNoLine(t *testing.T) {
	md, _, _ := narrMind(t)
	// A harvest at the very coordinates the correction is about.
	md.harvests.record(12, 7, 0, 1000)
	if _, ok := md.attributedHarvest(12, 7, 2000); !ok {
		t.Fatal("test setup: the correction's coordinates should be attributed")
	}

	md.chronicleNote(mustEvent(t, 2000, "agent.map_corrected", sim.MapCorrectedPayload{
		Agent: sim.Ref(1),
		Gone:  []sim.PlaceFact{{Kind: "pine", X: 12, Y: 7}},
	}))
	if len(md.narrLines) != 0 {
		t.Fatalf("lines = %d, want 0 — an attributed correction earns no line of its own: %v",
			len(md.narrLines), md.narrLines)
	}
	if md.corrTally.attributed != 1 || len(md.corrTally.places) != 1 || !md.corrTally.harvesters[0] {
		t.Errorf("tally = %+v, want 1 attributed at 1 place by agent 0", md.corrTally)
	}
	// The chapter's window still opens at the correction, exactly as it did
	// when the correction carried its own line.
	if md.narrFrom != 2000 {
		t.Errorf("narrFrom = %d, want 2000 (an attributed correction still opens the chapter)", md.narrFrom)
	}

	// FR-004: the unexplained case is byte-identical to today.
	md.chronicleNote(mustEvent(t, 2100, "agent.map_corrected", sim.MapCorrectedPayload{
		Agent: sim.Ref(1),
		Gone:  []sim.PlaceFact{{Kind: "pine", X: 60, Y: 60}},
	}))
	if len(md.narrLines) != 1 {
		t.Fatalf("lines = %d, want 1 (unexplained correction keeps its line)", len(md.narrLines))
	}
	const want = "Birch went looking for the pine at (60,60) and found it gone."
	if !strings.HasSuffix(md.narrLines[0], want) {
		t.Errorf("unexplained line = %q, want it to end with %q", md.narrLines[0], want)
	}
	if md.corrTally.unexplained != 1 {
		t.Errorf("unexplained tally = %d, want 1", md.corrTally.unexplained)
	}
}

// TestChapterCoalescesAttributedCorrections (T009/T011, User Story 1 and 2,
// SC-002): a chapter of 40 harvest-explained corrections across 12 places plus
// one genuine anomaly narrates as exactly two lines — ONE coalesced summary
// naming count, places and harvesters, and the anomaly's own untouched line.
func TestChapterCoalescesAttributedCorrections(t *testing.T) {
	md, _, _ := narrMind(t)

	// Twelve harvested tiles, felled/quarried by Ash (0) and Rowan (3).
	places := make([][2]int, 0, 12)
	for i := 0; i < 12; i++ {
		x, y := 10+i, 20+i
		places = append(places, [2]int{x, y})
		md.harvests.record(x, y, []int{0, 3}[i%2], 1000)
	}
	// Forty corrections over those twelve tiles.
	for i := 0; i < 40; i++ {
		p := places[i%len(places)]
		md.chronicleNote(mustEvent(t, int64(2000+i), "agent.map_corrected", sim.MapCorrectedPayload{
			Agent: sim.Ref(i % sim.AgentCount),
			Gone:  []sim.PlaceFact{{Kind: "pine", X: p[0], Y: p[1]}},
		}))
	}
	// One anomaly: nothing was ever harvested at (60,60).
	md.chronicleNote(mustEvent(t, 3000, "agent.map_corrected", sim.MapCorrectedPayload{
		Agent: sim.Ref(1),
		Gone:  []sim.PlaceFact{{Kind: "pine", X: 60, Y: 60}},
	}))

	md.chronicleNote(mustEvent(t, 57600, "sim.night_started", sim.DayPayload{Day: 1}))
	var job narrJob
	select {
	case job = <-md.narrQ:
	default:
		t.Fatal("night boundary did not close a chapter")
	}
	if len(job.lines) != 2 {
		t.Fatalf("lines = %d, want 2 (one coalesced summary + one anomaly): %v", len(job.lines), job.lines)
	}
	if !strings.HasSuffix(job.lines[0], "Birch went looking for the pine at (60,60) and found it gone.") {
		t.Errorf("anomaly line = %q, want it unchanged and first", job.lines[0])
	}
	const want = "Ordinary harvesting: 40 remembered things the villagers went for had already " +
		"been felled or quarried, at 12 locations, by Ash and Rowan. Routine village business."
	if !strings.HasSuffix(job.lines[1], want) {
		t.Errorf("summary line = %q, want it to end with %q", job.lines[1], want)
	}
	// SC-002: corrections contribute at most one line per chapter beyond the
	// genuine anomalies — no per-correction "found it gone" survives.
	gone := 0
	for _, l := range job.lines {
		if strings.Contains(l, "found it gone") {
			gone++
		}
	}
	if gone != 1 {
		t.Errorf("per-correction lines = %d, want 1 (the anomaly alone): %v", gone, job.lines)
	}
	if job.fromTick != 2000 {
		t.Errorf("fromTick = %d, want 2000 (the chapter's window is unchanged)", job.fromTick)
	}
	// T007: the tally is per-chapter state, reset with narrLines.
	if md.corrTally.attributed != 0 || md.corrTally.unexplained != 0 || len(md.corrTally.places) != 0 {
		t.Errorf("tally survived the chapter close: %+v", md.corrTally)
	}
}

// TestChapterWithoutAttributedCorrectionsUnchanged (T010, FR-008): a chapter
// that attributed nothing emits no summary line and reads exactly as it does
// today — and a chapter with nothing in it at all still spends no call.
func TestChapterWithoutAttributedCorrectionsUnchanged(t *testing.T) {
	md, _, _ := narrMind(t)
	md.chronicleNote(mustEvent(t, 1000, "agent.died", sim.DiedPayload{Agent: sim.Ref(0), Cause: "starvation"}))
	md.chronicleNote(mustEvent(t, 2000, "agent.map_corrected", sim.MapCorrectedPayload{
		Agent: sim.Ref(1),
		Gone:  []sim.PlaceFact{{Kind: "pine", X: 60, Y: 60}},
	}))
	md.chronicleNote(mustEvent(t, 57600, "sim.night_started", sim.DayPayload{Day: 1}))

	var job narrJob
	select {
	case job = <-md.narrQ:
	default:
		t.Fatal("night boundary did not close a chapter")
	}
	if len(job.lines) != 2 {
		t.Fatalf("lines = %d, want 2 (death + the unexplained correction): %v", len(job.lines), job.lines)
	}
	for _, l := range job.lines {
		if strings.Contains(l, corrSummaryMarker) {
			t.Errorf("a chapter with no attributed correction emitted a summary line: %q", l)
		}
	}
	// A wholly quiet chapter still spends nothing.
	md.chronicleNote(mustEvent(t, 86400, "sim.day_started", sim.DayPayload{Day: 2}))
	select {
	case <-md.narrQ:
		t.Fatal("quiet chapter should spend nothing")
	default:
	}
}

// TestCorrectionSummarySingular (T009): the coalesced line agrees with itself
// at one correction, one place, one harvester — the narrator reads prose.
func TestCorrectionSummarySingular(t *testing.T) {
	var tally correctionTally
	if s := tally.summary(func(int) string { return "Ash" }); s != "" {
		t.Fatalf("empty tally summary = %q, want \"\" (FR-008)", s)
	}
	tally.attribute(3, 4, 0)
	const want = "Ordinary harvesting: 1 remembered thing the villagers went for had already " +
		"been felled or quarried, at 1 location, by Ash. Routine village business."
	if got := tally.summary(func(int) string { return "Ash" }); got != want {
		t.Errorf("summary = %q, want %q", got, want)
	}
}

// TestNarrateUserPromptMarksSummaryAsBackground (T012, FR-005): the prompt
// names the coalesced line as ordinary background WITHOUT disturbing the
// "group by storyline, not by hour" instruction that governs everything else —
// and a chapter without such a line carries today's prompt unchanged (FR-008).
func TestNarrateUserPromptMarksSummaryAsBackground(t *testing.T) {
	plain := narrateUserPrompt(narrJob{
		day: 1, label: "day 1, dawn to nightfall",
		lines: []string{"[d1 08:00] Ash died of starvation."},
	})
	if strings.Contains(plain, "ordinary background") {
		t.Error("a chapter with no coalesced line must not carry the background instruction")
	}
	if !strings.Contains(plain, "Group by storyline, not by hour.") {
		t.Fatal("the storyline instruction went missing")
	}

	withSummary := narrateUserPrompt(narrJob{
		day: 1, label: "day 1, dawn to nightfall",
		lines: []string{
			"[d1 08:00] Ash died of starvation.",
			"[d1 20:00] " + corrSummaryMarker + " 40 remembered things the villagers went for had already been felled or quarried, at 12 locations, by Ash and Rowan. Routine village business.",
		},
	})
	if !strings.Contains(withSummary, "ordinary background, not storyline material") {
		t.Errorf("prompt does not mark the coalesced line as background:\n%s", withSummary)
	}
	if !strings.Contains(withSummary, corrSummaryMarker) {
		t.Error("the background instruction does not name the line it governs")
	}
	if !strings.Contains(withSummary, "Group by storyline, not by hour.") {
		t.Error("FR-005 disturbed the storyline instruction it must leave alone")
	}
}

// TestCorrectionTallyReport (T013, FR-007): per-chapter attributed vs
// unexplained counts land on the Mind's summary-log telemetry path, so the
// soak reads the outcome instead of re-deriving it from the event log. A
// chapter with no corrections at all reports nothing (FR-008).
func TestCorrectionTallyReport(t *testing.T) {
	var tally correctionTally
	if r := tally.report("day 1, dawn to nightfall"); r != "" {
		t.Fatalf("empty tally report = %q, want \"\"", r)
	}
	tally.attribute(3, 4, 0)
	tally.attribute(3, 4, 0)
	tally.attribute(9, 9, 3)
	tally.unexplained++
	const want = `mind: chronicle "day 1, dawn to nightfall" corrections: 3 attributed ` +
		`(2 locations, 2 harvesters), 1 unexplained`
	if got := tally.report("day 1, dawn to nightfall"); got != want {
		t.Errorf("report = %q, want %q", got, want)
	}
}
