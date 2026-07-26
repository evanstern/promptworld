package tui

// First-occurrence lessons projection tests (spec 055, TASK-117): the
// catalog/skin-resolution seam (T003), the trigger/row state machine (T004,
// T008 — one-active/dwell/queue/decay/spacing, done-signal clearing), the
// prompting-tier discrimination sweep (T011), the help-overlay pull-half
// equality (T012, SC-002), and the render-boundary sweep (T015, SC-003).
// Wiring-level tests (applyEvent persistence, the `x` key) live alongside
// the projection tests below rather than in tui_test.go, following
// decisions_test.go's precedent of co-locating a feature's tests in its own
// file.

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/evanstern/promptworld/internal/ipc"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worlds"
)

// mkEvent builds a store.Event of the given type from any JSON-serializable
// payload — the decisions_test.go fixture-building convention, generalized
// (this feature's fixtures span six different payload types across the
// taxonomy, one shared helper instead of six near-identical ones).
func mkEvent(typ string, payload any) store.Event {
	b, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	return store.Event{Type: typ, Payload: b}
}

// lessonFixtureEvent returns a real event that fires id's trigger — the
// contracts/lessons-catalog.md taxonomy table, executable. Seq is always 1
// (nonzero): applyEvent's "already reflected" guard is `e.Seq <= lastSeq`,
// and a fresh Model's lastSeq starts at its zero value (0) — a fixture with
// the also-zero-valued default Seq would be silently dropped by every
// applyEvent-driven test below, never reaching the projection at all.
func lessonFixtureEvent(t *testing.T, id string) store.Event {
	t.Helper()
	e := lessonFixtureEventPayload(t, id)
	e.Seq = 1
	return e
}

func lessonFixtureEventPayload(t *testing.T, id string) store.Event {
	t.Helper()
	switch id {
	case "first-suppression":
		return mkEvent("cog.outcome", sim.CogOutcomePayload{Job: "j1", Outcome: sim.OutcomeSuppressed})
	case "first-gru-attack":
		return mkEvent("gru.attacked", sim.GruAttackedPayload{Agent: 0, Health: 5})
	case "first-charge-regen":
		return mkEvent("metatron.charge_regenerated", sim.ChargeRegeneratedPayload{})
	case "first-order-expired":
		return mkEvent("metatron.order_expired", sim.OrderIDPayload{ID: "ord-1"})
	case "first-death":
		return mkEvent("agent.died", sim.DiedPayload{Agent: 0, Cause: "starvation"})
	case "first-rejected-tool-call":
		return mkEvent("cog.tool_call", sim.CogToolCallPayload{Job: "j1", Tool: "gather", Verdict: "rejected_gate", Reason: "stale"})
	case "first-custom-charter":
		return mkEvent("metatron.charter_observed", sim.CharterObservedPayload{Fingerprint: "abcd1234", Default: false})
	case "first-fuzzy-order":
		return mkEvent("metatron.order_placed", sim.GuardianOrder{ID: "ord-2", Confirm: true})
	// --- tranche 2 (spec 077 US3) ---
	case "first-explain-answer":
		return mkEvent("cog.tool_call", sim.CogToolCallPayload{Job: "j1", Tool: "explain", Verdict: "read_ok"})
	case "first-report-card":
		return mkEvent("guardian.report_card", sim.GuardianReportCardPayload{Fingerprint: "a1b2c3", Note: "steady work"})
	case "first-skill-file":
		return mkEvent("metatron.skills_observed", sim.SkillsObservedPayload{Fingerprint: "ab12cd34ef56", Names: []string{"10-watch.md"}})
	case "same-refusal-pattern":
		// One same-reason rejection — the FOLD entry needs three of these
		// (the sweep primes two; the fold tests below drive the arithmetic).
		return mkEvent("cog.tool_call", sim.CogToolCallPayload{Job: "j1", Tool: "gather", Verdict: "rejected_gate", Reason: "outside the charter"})
	}
	t.Fatalf("lessonFixtureEvent: no fixture wired for id %q", id)
	return store.Event{}
}

// --- T003: catalog shape + skin-token resolution ---

func TestLessonCatalogMinimumTaxonomy(t *testing.T) {
	if len(lessonCatalog) != 12 {
		t.Fatalf("lessonCatalog has %d entries, want exactly 12 (contracts minimum 8 + spec 077 tranche 2)", len(lessonCatalog))
	}
	seen := map[string]bool{}
	mechanics, prompting := 0, 0
	for _, e := range lessonCatalog {
		if e.ID == "" || e.Title == "" || e.Body == "" || e.Text == "" || e.Pointer == "" {
			t.Errorf("entry %+v has an empty required field", e)
		}
		if seen[e.ID] {
			t.Errorf("duplicate catalog id %q", e.ID)
		}
		seen[e.ID] = true
		switch e.Tier {
		case lessonTierMechanics:
			mechanics++
		case lessonTierPrompting:
			prompting++
		default:
			t.Errorf("entry %q has an unrecognized tier %q", e.ID, e.Tier)
		}
		// Exactly one trigger seam per entry (spec 077 FR-019): a pure
		// per-event predicate OR the one fold-trigger — never both, never
		// neither.
		if (e.Trigger == nil) == (e.FoldTrigger == nil) {
			t.Errorf("entry %q must carry exactly one of Trigger/FoldTrigger", e.ID)
		}
	}
	if mechanics != 5 || prompting != 7 {
		t.Errorf("got %d mechanics + %d prompting entries, want 5 + 7 (spec 077 FR-018)", mechanics, prompting)
	}
	// first-faith-event is OUT (spec 077 FR-020): TASK-118 unrun — the
	// catalog must not stub it.
	if seen["first-faith-event"] {
		t.Error("first-faith-event is stubbed — it rides TASK-118, never this catalog (FR-020)")
	}
}

// TestLessonSkinResolveNoRawTokens (FR-008/SC-005): every catalog string,
// resolved, must never contain a raw "{{" literal — the default skin's
// values (research.md R1's bounded fallback) must fully cover every token
// the catalog actually uses.
func TestLessonSkinResolveNoRawTokens(t *testing.T) {
	for _, e := range lessonCatalog {
		for field, s := range map[string]string{"Title": e.Title, "Body": e.Body, "Text": e.Text, "Pointer": e.Pointer} {
			if resolved := lessonSkinResolve(s, nil); strings.Contains(resolved, "{{") {
				t.Errorf("%s.%s resolves with a raw token literal: %q", e.ID, field, resolved)
			}
		}
	}
}

// TestLessonPointerSuffixFitsNarrowWidth guards against authoring a pointer
// string so long that its own pull-path suffix gets clipped off the visible
// line at a realistic narrow terminal width (80 cols) — clipLine's crop is
// correct overflow behavior, but a lesson whose own suffix is invisible by
// construction would defeat FR-001's "always knows how to find it again and
// how to clear it now" guarantee. Not a hard spec ceiling (no FR pins 80
// columns specifically), just a regression pin on the authored copy.
func TestLessonPointerSuffixFitsNarrowWidth(t *testing.T) {
	const narrowWidth = 80
	for _, e := range lessonCatalog {
		line := lessonSkinResolve(e.Pointer, nil) + "  " + lessonPullSuffix
		if w := len([]rune(line)); w > narrowWidth {
			t.Errorf("%s: pointer+suffix is %d runes, wider than a plausible narrow terminal (%d): %q", e.ID, w, narrowWidth, line)
		}
	}
}

// --- T004/T008: one-active, dwell, queue, decay, spacing, done-signal ---

// TestLessonTriggersSweepEachTaxonomyEntryOnce is the SC-001 fixture sweep:
// each of the 8 catalog entries fires exactly once across a simulated
// two-world + client-restart sequence, using the real per-user seen-state
// file (internal/worlds), never the mock in-memory map alone — the
// persistence half of the contract, not just the in-process state machine.
func TestLessonTriggersSweepEachTaxonomyEntryOnce(t *testing.T) {
	for _, entry := range lessonCatalog {
		entry := entry
		t.Run(entry.ID, func(t *testing.T) {
			t.Setenv("PROMPTWORLD_HOME", t.TempDir())
			ev := lessonFixtureEvent(t, entry.ID)
			now := time.Now()

			// The fold entry (spec 077): its fixture event is ALSO
			// first-rejected-tool-call's trigger — mark that one seen so the
			// row is free, and prime two same-reason strikes; the third (the
			// sweep's own ingest below) crosses the threshold.
			if entry.FoldTrigger != nil {
				worlds.MarkLessonSeen("first-rejected-tool-call", "world-a")
			}

			// World A / this client's first boot: a fresh seen-record.
			ltA := newLessonTriggers(lessonSeenIDs(worlds.LoadLessonsSeen()))
			if entry.FoldTrigger != nil {
				for i := 0; i < sameRefusalThreshold-1; i++ {
					if got := ltA.ingest(lessonFixtureEvent(t, entry.ID), now); got != nil {
						t.Fatalf("priming strike %d surfaced %v, want nothing before the threshold", i+1, got)
					}
				}
			}
			surfaced := ltA.ingest(ev, now)
			if surfaced == nil || surfaced.ID != entry.ID {
				t.Fatalf("world A: expected %q to surface, got %v", entry.ID, surfaced)
			}
			worlds.MarkLessonSeen(surfaced.ID, "world-a")

			// The same trigger re-arriving in the SAME session must not
			// re-queue or re-surface (already in the in-memory seen set).
			if got := ltA.ingest(ev, now.Add(time.Second)); got != nil {
				t.Errorf("world A re-trigger: expected no re-surface, got %v", got)
			}

			// World B — a second world for the same player, reading the
			// persisted record fresh: never fires.
			ltB := newLessonTriggers(lessonSeenIDs(worlds.LoadLessonsSeen()))
			if got := ltB.ingest(ev, now); got != nil {
				t.Errorf("world B: lesson %q fired again, want suppressed by the persisted record", entry.ID)
			}

			// A client restart for either world loads the identical
			// persisted record: still suppressed.
			ltRestart := newLessonTriggers(lessonSeenIDs(worlds.LoadLessonsSeen()))
			if got := ltRestart.ingest(ev, now); got != nil {
				t.Errorf("restart: lesson %q fired again after restart", entry.ID)
			}
		})
	}
}

// TestLessonPromptingTriggersDiscriminate (T011, US2): the payload-field
// discrimination the prompting tier depends on — a non-fuzzy order, a
// default charter, and a landed tool call must each fire NOTHING.
func TestLessonPromptingTriggersDiscriminate(t *testing.T) {
	cases := []struct {
		name string
		ev   store.Event
	}{
		{"non-fuzzy order does not trigger the fuzzy-order lesson",
			mkEvent("metatron.order_placed", sim.GuardianOrder{ID: "o1", Confirm: false})},
		{"default charter does not trigger the custom-charter lesson",
			mkEvent("metatron.charter_observed", sim.CharterObservedPayload{Fingerprint: "x", Default: true})},
		{"a landed tool call does not trigger the rejected-tool-call lesson",
			mkEvent("cog.tool_call", sim.CogToolCallPayload{Job: "j1", Tool: "gather", Verdict: "landed"})},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			lt := newLessonTriggers(nil)
			if got := lt.ingest(c.ev, time.Now()); got != nil {
				t.Errorf("expected no lesson to surface, got %v", got)
			}
		})
	}
}

// rejectionEvent builds one rejected cog.tool_call with the given reason.
func rejectionEvent(reason string) store.Event {
	e := mkEvent("cog.tool_call", sim.CogToolCallPayload{Job: "j1", Tool: "gather",
		Verdict: "rejected_gate", Reason: reason})
	e.Seq = 1
	return e
}

// TestSameRefusalFoldThreeStrikes (spec 077 US3 AS-4, SC-004): two
// same-reason rejections surface nothing; the third surfaces
// same-refusal-pattern — once per player, never again.
func TestSameRefusalFoldThreeStrikes(t *testing.T) {
	// first-rejected-tool-call pre-seen so the row is free for the detector.
	lt := newLessonTriggers(map[string]bool{"first-rejected-tool-call": true})
	now := time.Now()
	for i := 0; i < 2; i++ {
		if got := lt.ingest(rejectionEvent("outside the charter"), now); got != nil {
			t.Fatalf("strike %d surfaced %v, want nothing", i+1, got)
		}
	}
	got := lt.ingest(rejectionEvent("outside the charter"), now)
	if got == nil || got.ID != "same-refusal-pattern" {
		t.Fatalf("third strike surfaced %v, want same-refusal-pattern", got)
	}
	// A fourth same-reason rejection never re-surfaces it (seen).
	if again := lt.ingest(rejectionEvent("outside the charter"), now); again != nil {
		t.Errorf("fourth strike re-surfaced %v", again)
	}
}

// TestSameRefusalFoldMixedReasonsNeverTrigger (AS-4's negative): three
// DIFFERENT-reason rejections never trigger the detector, and reasonless /
// non-rejection verdicts never count at all.
func TestSameRefusalFoldMixedReasonsNeverTrigger(t *testing.T) {
	lt := newLessonTriggers(map[string]bool{"first-rejected-tool-call": true})
	now := time.Now()
	for _, reason := range []string{"outside the charter", "too many targets", "stale snapshot"} {
		if got := lt.ingest(rejectionEvent(reason), now); got != nil {
			t.Fatalf("mixed reasons surfaced %v, want nothing", got)
		}
	}
	// Reasonless rejections and read verdicts are not strikes.
	empty := mkEvent("cog.tool_call", sim.CogToolCallPayload{Job: "j1", Tool: "gather", Verdict: "rejected_gate"})
	read := mkEvent("cog.tool_call", sim.CogToolCallPayload{Job: "j1", Tool: "explain", Verdict: "read_error", Reason: "outside the charter"})
	for i := 0; i < 5; i++ {
		if got := lt.ingest(empty, now); got != nil {
			t.Fatalf("reasonless rejection surfaced %v", got)
		}
		if got := lt.ingest(read, now); got != nil {
			t.Fatalf("read_error surfaced %v (only rejected_* verdicts count)", got)
		}
	}
}

// TestLessonFoldReasonCap: past the cap, NEW reasons stop being tracked
// (bounded fold, spec 077 FR-019) — already-tracked reasons keep counting.
func TestLessonFoldReasonCap(t *testing.T) {
	var f lessonFold
	for i := 0; i < lessonFoldReasonCap; i++ {
		f.note(rejectionEvent(strings.Repeat("r", i+1)))
	}
	f.note(rejectionEvent("one past the cap"))
	if len(f.rejections) != lessonFoldReasonCap {
		t.Fatalf("fold tracks %d reasons, want capped at %d", len(f.rejections), lessonFoldReasonCap)
	}
	if f.rejections["one past the cap"] != 0 {
		t.Error("a reason past the cap was tracked")
	}
	f.note(rejectionEvent("r")) // an already-tracked reason still counts
	if f.rejections["r"] != 2 {
		t.Errorf("tracked reason count = %d, want 2", f.rejections["r"])
	}
}

// TestLessonTriggersOneActiveQueuesInArrivalOrder (FR-004): a second,
// different trigger arriving while one lesson is active queues rather than
// replacing it, and promotes only once the active one clears AND the
// spacing gap elapses.
func TestLessonTriggersOneActiveQueuesInArrivalOrder(t *testing.T) {
	lt := newLessonTriggers(nil)
	now := time.Now()

	first := lt.ingest(lessonFixtureEvent(t, "first-death"), now)
	if first == nil || first.ID != "first-death" {
		t.Fatalf("expected first-death to surface immediately, got %v", first)
	}

	second := lt.ingest(lessonFixtureEvent(t, "first-gru-attack"), now.Add(time.Second))
	if second != nil {
		t.Errorf("a second trigger while one is active must queue, not surface: %v", second)
	}
	if got := lt.ActiveEntry(); got == nil || got.ID != "first-death" {
		t.Fatalf("active lesson must remain first-death, got %v", got)
	}
	if len(lt.queue) != 1 || lt.queue[0].entry.ID != "first-gru-attack" {
		t.Fatalf("expected first-gru-attack queued, got %+v", lt.queue)
	}

	clearAt := now.Add(2 * time.Second)
	if !lt.Dismiss(clearAt) {
		t.Fatal("expected Dismiss to report it cleared the active lesson")
	}
	if lt.ActiveEntry() != nil {
		t.Error("expected no active lesson immediately after dismissal")
	}

	// Too soon after the clear: spacing still gates the promotion.
	if got := lt.Advance(clearAt.Add(time.Millisecond)); got != nil {
		t.Errorf("expected spacing to block an immediate promotion, got %v", got)
	}

	// Spacing elapsed: the queued entry promotes and is marked seen.
	promoted := lt.Advance(clearAt.Add(lessonSpacing + time.Millisecond))
	if promoted == nil || promoted.ID != "first-gru-attack" {
		t.Fatalf("expected first-gru-attack to promote once spacing elapsed, got %v", promoted)
	}
	if !lt.seen["first-gru-attack"] {
		t.Error("expected the promoted lesson marked seen at surface time")
	}
}

// TestLessonTriggersQueueDecayDropsWithoutMarkingSeen (FR-004 "opportunity
// decay"): a queued trigger that never gets its turn within the decay
// window is dropped silently — never surfaced stale, and never recorded
// seen (a decayed opportunity may still fire on a later first occurrence).
func TestLessonTriggersQueueDecayDropsWithoutMarkingSeen(t *testing.T) {
	lt := newLessonTriggers(nil)
	now := time.Now()
	lt.ingest(lessonFixtureEvent(t, "first-death"), now) // occupies the row indefinitely (no done-signal)
	lt.ingest(lessonFixtureEvent(t, "first-gru-attack"), now)

	later := now.Add(lessonQueueDecay + time.Second)
	if got := lt.Advance(later); got != nil {
		t.Errorf("the row is still occupied; nothing should promote: %v", got)
	}
	if len(lt.queue) != 0 {
		t.Errorf("expected the decayed entry dropped from the queue, got %+v", lt.queue)
	}
	if lt.seen["first-gru-attack"] {
		t.Error("a decayed (never-shown) queue entry must not be marked seen")
	}
}

// TestLessonTriggersSpacingGatesFreshTriggerToo: the anti-spam spacing gap
// applies uniformly — a brand-new trigger arriving right after a clear
// queues just like a promoted one would, rather than bypassing the gap.
func TestLessonTriggersSpacingGatesFreshTriggerToo(t *testing.T) {
	lt := newLessonTriggers(nil)
	now := time.Now()
	lt.ingest(lessonFixtureEvent(t, "first-death"), now)
	lt.Dismiss(now.Add(time.Second))

	fresh := lt.ingest(lessonFixtureEvent(t, "first-gru-attack"), now.Add(time.Second+time.Millisecond))
	if fresh != nil {
		t.Errorf("expected the fresh trigger to queue (spacing not yet elapsed), got %v", fresh)
	}
	if len(lt.queue) != 1 {
		t.Fatalf("expected the fresh trigger queued, got %+v", lt.queue)
	}

	promoted := lt.Advance(now.Add(time.Second + lessonSpacing + time.Millisecond))
	if promoted == nil || promoted.ID != "first-gru-attack" {
		t.Errorf("expected the queued fresh trigger to promote once spacing elapsed, got %v", promoted)
	}
}

// TestLessonTriggersFirstEverLessonBypassesSpacing: a fresh lessonTriggers
// (clearedAt's zero value — nothing has ever cleared) must not make the
// very first lesson wait out the spacing gap.
func TestLessonTriggersFirstEverLessonBypassesSpacing(t *testing.T) {
	lt := newLessonTriggers(nil)
	surfaced := lt.ingest(lessonFixtureEvent(t, "first-death"), time.Now())
	if surfaced == nil || surfaced.ID != "first-death" {
		t.Fatalf("expected the very first lesson to surface immediately, got %v", surfaced)
	}
}

// TestLessonTriggersDoneSignalClearsAndRecordsSeen (FR-005): the pointed-at
// action (here, placing a new standing order) clears first-order-expired's
// dwell without `x` — marking seen already happened at surface time, so the
// done-signal itself performs no additional persistence.
func TestLessonTriggersDoneSignalClearsAndRecordsSeen(t *testing.T) {
	lt := newLessonTriggers(nil)
	now := time.Now()
	surfaced := lt.ingest(lessonFixtureEvent(t, "first-order-expired"), now)
	if surfaced == nil || surfaced.ID != "first-order-expired" {
		t.Fatalf("expected first-order-expired to surface, got %v", surfaced)
	}
	if !lt.seen["first-order-expired"] {
		t.Fatal("expected the lesson marked seen at surface time")
	}

	// A non-fuzzy order placed is the done-signal here — clears the active
	// lesson but is not itself a new trigger (Confirm is false).
	placed := mkEvent("metatron.order_placed", sim.GuardianOrder{ID: "o2", Confirm: false})
	got := lt.ingest(placed, now.Add(time.Millisecond))
	if got != nil {
		t.Errorf("the done-signal event is not itself a new lesson trigger here, got %v", got)
	}
	if lt.ActiveEntry() != nil {
		t.Error("expected the done-signal to clear the active lesson")
	}
}

// TestLessonTriggersDoneSignalAndNewTriggerSameEvent: a single event can
// BOTH clear one lesson's dwell (its done-signal) AND separately trigger
// another — metatron.order_placed is first-order-expired's done-signal AND
// (when fuzzy) first-fuzzy-order's trigger. The freshly-triggered lesson
// still respects the spacing gap the same-event clear just started.
func TestLessonTriggersDoneSignalAndNewTriggerSameEvent(t *testing.T) {
	lt := newLessonTriggers(nil)
	now := time.Now()
	lt.ingest(lessonFixtureEvent(t, "first-order-expired"), now)

	fuzzyPlaced := mkEvent("metatron.order_placed", sim.GuardianOrder{ID: "o2", Confirm: true})
	surfaced := lt.ingest(fuzzyPlaced, now.Add(time.Millisecond))
	if surfaced != nil {
		t.Errorf("expected the fuzzy-order lesson to queue (spacing just started by the same-event clear), got %v", surfaced)
	}
	if lt.ActiveEntry() != nil {
		t.Error("expected the row empty immediately after the done-signal clear")
	}
	if len(lt.queue) != 1 || lt.queue[0].entry.ID != "first-fuzzy-order" {
		t.Fatalf("expected first-fuzzy-order queued from the same event, got %+v", lt.queue)
	}
}

// TestLessonTriggersDismissNoOpWhenNothingActive (FR-010): `x` with no
// active lesson is a strict no-op — false, no state change.
func TestLessonTriggersDismissNoOpWhenNothingActive(t *testing.T) {
	lt := newLessonTriggers(nil)
	if lt.Dismiss(time.Now()) {
		t.Error("Dismiss on an empty row must report false (strict no-op)")
	}
}

// --- tui.go wiring: applyEvent persistence + the `x` key dispatch ---

// TestApplyEventSurfacesLessonAndPersists: the real ingest seam
// (applyEvent) surfaces a lesson and persists it via worlds.MarkLessonSeen
// — the same file LoadLessonsSeen reads back, proving the wiring (not just
// the projection in isolation).
func TestApplyEventSurfacesLessonAndPersists(t *testing.T) {
	m := testModel(t)
	m.applyEvent(lessonFixtureEvent(t, "first-gru-attack"))
	if got := m.lessons.ActiveEntry(); got == nil || got.ID != "first-gru-attack" {
		t.Fatalf("expected first-gru-attack active after applyEvent, got %v", got)
	}
	if seen := worlds.LoadLessonsSeen(); !seen.Seen("first-gru-attack") {
		t.Error("expected the surfaced lesson persisted to the per-user seen-state file")
	}
}

// TestXKeyDismissesActiveLessonAndNoOpsOtherwise (T006): global dispatch,
// via the real Update() path.
func TestXKeyDismissesActiveLessonAndNoOpsOtherwise(t *testing.T) {
	m := testModel(t)
	noop, _ := m.Update(key("x"))
	if noop.(Model).lessons.ActiveEntry() != nil {
		t.Fatal("fixture setup: unexpected active lesson before the no-op check")
	}

	m.applyEvent(lessonFixtureEvent(t, "first-death"))
	if m.lessons.ActiveEntry() == nil {
		t.Fatal("fixture setup: expected an active lesson")
	}
	dismissed, _ := m.Update(key("x"))
	if got := dismissed.(Model).lessons.ActiveEntry(); got != nil {
		t.Errorf("expected x to dismiss the active lesson, still active: %v", got)
	}
}

// --- T012/SC-002: the help overlay's pull half, one catalog two surfaces ---

func TestPopulateHelpLessonsCatalogEquality(t *testing.T) {
	orig := helpLessons
	t.Cleanup(func() { helpLessons = orig })
	populateHelpLessons(nil)

	if len(helpLessons) != len(lessonCatalog) {
		t.Fatalf("helpLessons has %d entries, catalog has %d", len(helpLessons), len(lessonCatalog))
	}
	for i, e := range lessonCatalog {
		if helpLessons[i].ID != e.ID {
			t.Errorf("helpLessons[%d].ID = %q, want %q (id-for-id order)", i, helpLessons[i].ID, e.ID)
		}
		if strings.Contains(helpLessons[i].Title, "{{") || strings.Contains(helpLessons[i].Body, "{{") {
			t.Errorf("helpLessons[%d] contains an unresolved skin token: %+v", i, helpLessons[i])
		}
	}
	if strings.Contains(strings.Join(helpLessonsLines(76), "\n"), "lessons appear here") {
		t.Error("populated helpLessons should no longer render the placeholder line")
	}
}

// --- T007/T013/T014/T015: rendering, stage defaults, fold, narrow carry ---

func withStage(m Model, stage string) Model {
	m.connected = true // headerView() short-circuits to the disconnected banner otherwise, before any badge
	m.status = &ipc.StatusData{World: ipc.WorldStatus{Stage: stage}}
	return m
}

// TestLessonRowRendersActiveLessonWithSuffix (FR-001): the two-line row,
// its pull-path suffix, and the no-raw-token invariant (SC-005), inside a
// real View() render.
func TestLessonRowRendersActiveLessonWithSuffix(t *testing.T) {
	m := withStage(widescreenModel(t), "stage-1")
	m.applyEvent(lessonFixtureEvent(t, "first-death"))
	view := m.View()
	if len(strings.Split(view, "\n")) != m.height {
		t.Fatalf("View() must render exactly %d lines", m.height)
	}
	if !strings.Contains(view, "A villager has died") {
		t.Errorf("expected the active lesson's text line, got:\n%s", view)
	}
	if !strings.Contains(view, lessonPullSuffix) {
		t.Errorf("expected the pull-path suffix %q, got:\n%s", lessonPullSuffix, view)
	}
	if strings.Contains(view, "{{") {
		t.Errorf("no raw skin-token literal may render: %s", view)
	}
}

// TestLessonRowAbsentWhenNoneActive (data-model.md "none" state): stage
// 1-2 eligible but nothing active renders 0 rows, not a blank block.
func TestLessonRowAbsentWhenNoneActive(t *testing.T) {
	m := withStage(widescreenModel(t), "stage-1")
	view := m.View()
	if strings.Contains(view, lessonPullSuffix) {
		t.Errorf("no active lesson: row must be absent, got:\n%s", view)
	}
	if rows := computeRows(m.height, m.wantsLessonRow()); rows.Lesson != 0 {
		t.Errorf("expected 0 lesson rows with nothing active, got %d", rows.Lesson)
	}
}

// TestLessonRowShownAtStage1And2 (patterns/stage-defaults.md): the row's
// default is on, full form, at stages 1 and 2.
func TestLessonRowShownAtStage1And2(t *testing.T) {
	for _, stage := range []string{"stage-1", "stage-2"} {
		t.Run(stage, func(t *testing.T) {
			m := withStage(widescreenModel(t), stage)
			m.applyEvent(lessonFixtureEvent(t, "first-death"))
			view := m.View()
			if strings.Contains(view, "[lesson]") {
				t.Errorf("stage %q with an active lesson should show the full row, not the badge: %s", stage, view)
			}
			if !strings.Contains(view, lessonPullSuffix) {
				t.Errorf("expected the full row at stage %q, got:\n%s", stage, view)
			}
		})
	}
}

// TestLessonBadgeAtStage3AndPreLadder (patterns/stage-defaults.md): stage
// 3+ and pre-ladder ("") default to the header badge, unconditionally —
// even with nothing active, and never the two-line row.
func TestLessonBadgeAtStage3AndPreLadder(t *testing.T) {
	for _, stage := range []string{"stage-3", "stage-4", ""} {
		t.Run("stage_"+stage, func(t *testing.T) {
			m := withStage(widescreenModel(t), stage)
			view := m.View()
			if !strings.Contains(view, "[lesson]") {
				t.Errorf("expected the [lesson] badge at stage %q, got:\n%s", stage, view)
			}
			if strings.Contains(view, lessonPullSuffix) {
				t.Errorf("stage %q must not show the full row", stage)
			}
		})
	}
}

// TestLessonRowFoldsBeforeGuardianStripUnderHeightPressure (patterns/
// layout.md ruling a step 3, spec 055): at a stage where the row would
// otherwise show, insufficient height folds the LESSON ROW first — the
// guardian strip stays on as long as folding the lesson row alone already
// bought back enough body.
func TestLessonRowFoldsBeforeGuardianStripUnderHeightPressure(t *testing.T) {
	m := withStage(widescreenModel(t), "stage-1")
	m.status.Clock = ipc.ClockStatus{GuardianCharges: 2, Tick: 100}
	m.applyEvent(lessonFixtureEvent(t, "first-death"))

	m.height = 40
	tallRows := computeRows(m.height, m.wantsLessonRow())
	if tallRows.Lesson != 2 || tallRows.Strip != 1 {
		t.Fatalf("fixture setup: expected both the row and the strip at height 40, got %+v", tallRows)
	}
	if tallView := m.View(); !strings.Contains(tallView, lessonPullSuffix) {
		t.Errorf("expected the full row at height 40, got:\n%s", tallView)
	}

	// 17: tight enough that folding ONLY the lesson row is required (and
	// sufficient) to keep the guardian strip's own row — see layout.go's
	// computeRows doc comment for the arithmetic this height was chosen to
	// exercise (folds the lesson row at step 3, never reaches step 4).
	m.height = 17
	shortRows := computeRows(m.height, m.wantsLessonRow())
	if shortRows.Lesson != 0 || shortRows.Strip != 1 {
		t.Fatalf("expected the lesson row folded but the strip still on at height 17, got %+v", shortRows)
	}
	shortView := m.View()
	if strings.Contains(shortView, lessonPullSuffix) {
		t.Errorf("expected the row folded to the badge at height 17, got:\n%s", shortView)
	}
	if !strings.Contains(shortView, "[lesson]") {
		t.Errorf("expected the [lesson] badge once folded, got:\n%s", shortView)
	}
	if lines := strings.Split(shortView, "\n"); len(lines) != m.height {
		t.Errorf("View() must render exactly %d lines, got %d", m.height, len(lines))
	}
}

// TestLessonRowNarrowCarry (patterns/layout.md ruling b): the row is
// carried in the narrow fallback with the same stage defaults. Uses
// first-charge-regen (a short pointer) rather than first-death: at 80
// cols several catalog entries' pointer+suffix legitimately exceed the
// width and clipLine crops the tail — correct overflow behavior (the same
// discipline every other panel uses), not something this test is about.
func TestLessonRowNarrowCarry(t *testing.T) {
	m := withStage(testModel(t), "stage-1") // testModel defaults to narrow (80 cols)
	m.applyEvent(lessonFixtureEvent(t, "first-charge-regen"))
	view := m.View()
	if !strings.Contains(view, lessonPullSuffix) {
		t.Errorf("expected the lesson row carried in narrow, got:\n%s", view)
	}
}

// TestLessonBadgeNarrowAtStage3: the badge form carries into narrow too
// (the shared headerView renders it in both layouts).
func TestLessonBadgeNarrowAtStage3(t *testing.T) {
	m := withStage(testModel(t), "stage-3")
	view := m.View()
	if !strings.Contains(view, "[lesson]") {
		t.Errorf("expected the [lesson] badge in narrow at stage 3, got:\n%s", view)
	}
}

// TestLessonRowNeverExceedsTwoRows (SC-003): sweeps stage x height and
// asserts computeRows never allocates the lesson row anything but 0 or 2.
func TestLessonRowNeverExceedsTwoRows(t *testing.T) {
	stages := []string{"", "stage-1", "stage-2", "stage-3", "stage-4"}
	heights := []int{9, 10, 14, 15, 16, 17, 18, 20, 30, 40, 60}
	for _, stage := range stages {
		for _, h := range heights {
			m := withStage(widescreenModel(t), stage)
			m.height = h
			m.applyEvent(lessonFixtureEvent(t, "first-death"))
			rows := computeRows(m.height, m.wantsLessonRow())
			if rows.Lesson != 0 && rows.Lesson != 2 {
				t.Errorf("stage=%q height=%d: Lesson rows = %d, want 0 or 2", stage, h, rows.Lesson)
			}
		}
	}
}

// --- T013 (spec 066, TASK-128, SC-005): stage-driven surface arrival
// routed through this same first-occurrence machinery ---

// TestAnnounceSurfaceArrivalExactlyOnce proves stagedefaults.go's
// announceSurfaceArrival dispatch against a synthetic catalog entry (no
// real governed surface maps to one today — see surfaceArrivalLessonID's
// doc comment, stagedefaults.go: under the CURRENT authority table no
// numbered-stage transition ever widens a row): a mapped surface's arrival
// surfaces its lesson exactly once even if announced twice, using this
// package's own seen-map/dedupe machinery — no second mechanism.
func TestAnnounceSurfaceArrivalExactlyOnce(t *testing.T) {
	const testSurfaceID = "test-only-surface"
	const testLessonID = "test-only-lesson"

	origCatalog := lessonCatalog
	origMap := surfaceArrivalLessonID
	t.Cleanup(func() {
		lessonCatalog = origCatalog
		surfaceArrivalLessonID = origMap
	})
	lessonCatalog = append(append([]lessonEntry{}, origCatalog...), lessonEntry{
		ID: testLessonID, Title: "Test", Body: "test", Text: "test", Pointer: "test",
	})
	surfaceArrivalLessonID = map[string]string{testSurfaceID: testLessonID}

	lt := newLessonTriggers(nil)
	now := time.Now()
	first := announceSurfaceArrival(&lt, testSurfaceID, now)
	if first == nil || first.ID != testLessonID {
		t.Fatalf("first announcement should surface %q, got %+v", testLessonID, first)
	}
	second := announceSurfaceArrival(&lt, testSurfaceID, now)
	if second != nil {
		t.Errorf("second announcement of the same surface must not re-surface (exactly-once, SC-005), got %+v", second)
	}
	if !lt.seen[testLessonID] {
		t.Error("the announced lesson must be marked seen")
	}
}
