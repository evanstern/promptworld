package mind

// Spec 105 (TASK-172) T007: the per-night acceptance summary and the
// consecutive-failure WARNING escalation (FR-006, FR-007, SC-002).

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

func marker(night int64, outcome, reason string) sim.ConsolidatedPayload {
	return sim.ConsolidatedPayload{Agent: sim.Ref(0), Night: night, Outcome: outcome, Reason: reason}
}

// TestNightReportSummaryLine: one closed night flushes exactly one line with
// the accepted / rejected-by-reason / skipped-empty split, and the counters
// are consumed by the flush.
func TestNightReportSummaryLine(t *testing.T) {
	var r nightReport
	r.record(marker(5, sim.ConsolidationAccepted, ""))
	r.record(marker(5, sim.ConsolidationAccepted, ""))
	r.record(marker(5, sim.ConsolidationRejected, sim.ConsolidationReasonTruncated))
	r.record(marker(5, sim.ConsolidationRejected, sim.ConsolidationReasonTruncated))
	r.record(marker(5, sim.ConsolidationRejected, "unparseable"))
	r.record(marker(5, sim.ConsolidationSkippedEmpty, ""))

	// The night is still open: nothing flushes at its own index.
	if lines := r.flushBefore(5); lines != nil {
		t.Fatalf("open night flushed: %v", lines)
	}
	lines := r.flushBefore(6)
	if len(lines) != 1 {
		t.Fatalf("lines = %d, want 1", len(lines))
	}
	want := "mind: consolidation night 5: 2 accepted, 3 rejected (truncated 2, unparseable 1), 1 skipped-empty"
	if lines[0] != want {
		t.Errorf("line = %q\nwant   %q", lines[0], want)
	}
	if lines = r.flushBefore(7); lines != nil {
		t.Errorf("second flush re-reported the night: %v", lines)
	}
	// A partially-accepted night never arms the streak.
	if r.streak != 0 {
		t.Errorf("streak = %d, want 0", r.streak)
	}
}

// TestNightReportWarningStreak is SC-002: two consecutive fully-failed
// attempted nights escalate to WARNING:, it repeats each further failed night,
// and an accepting night resets it. An all-empty night neither extends nor
// resets the streak (FR-007 counts attempted nights).
func TestNightReportWarningStreak(t *testing.T) {
	var r nightReport
	fail := func(night int64) {
		r.record(marker(night, sim.ConsolidationRejected, sim.ConsolidationReasonTruncated))
		r.record(marker(night, sim.ConsolidationRejected, "unparseable"))
	}

	// Night 1 fails: a plain summary (streak 1, below threshold).
	fail(1)
	lines := r.flushBefore(2)
	if len(lines) != 1 || strings.Contains(lines[0], "WARNING") {
		t.Fatalf("first failed night must not warn: %v", lines)
	}

	// Night 2 fails: streak 2 → WARNING with streak length and remedy.
	fail(2)
	lines = r.flushBefore(3)
	if len(lines) != 1 || !strings.HasPrefix(lines[0], "mind: WARNING:") {
		t.Fatalf("second failed night must warn: %v", lines)
	}
	if !strings.Contains(lines[0], "2 consecutive nights") ||
		!strings.Contains(lines[0], "max_tokens.consolidation") {
		t.Errorf("warning lacks streak/remedy: %q", lines[0])
	}

	// Night 3: all skipped-empty — no attempt; plain line, streak untouched.
	r.record(marker(3, sim.ConsolidationSkippedEmpty, ""))
	lines = r.flushBefore(4)
	if len(lines) != 1 || strings.Contains(lines[0], "WARNING") {
		t.Fatalf("empty night must not warn: %v", lines)
	}
	if r.streak != 2 {
		t.Errorf("empty night moved the streak: %d, want 2", r.streak)
	}

	// Night 4 fails again: the warning REPEATS, streak now 3.
	fail(4)
	lines = r.flushBefore(5)
	if len(lines) != 1 || !strings.Contains(lines[0], "3 consecutive nights") {
		t.Fatalf("warning must repeat with the grown streak: %v", lines)
	}

	// Night 5 accepts one: plain summary, streak reset.
	r.record(marker(5, sim.ConsolidationAccepted, ""))
	r.record(marker(5, sim.ConsolidationRejected, sim.ConsolidationReasonTruncated))
	lines = r.flushBefore(6)
	if len(lines) != 1 || strings.Contains(lines[0], "WARNING") {
		t.Fatalf("accepting night must not warn: %v", lines)
	}
	if r.streak != 0 {
		t.Errorf("streak after acceptance = %d, want 0", r.streak)
	}

	// Night 6 fails: back to a plain line (streak restarts at 1).
	fail(6)
	lines = r.flushBefore(7)
	if len(lines) != 1 || strings.Contains(lines[0], "WARNING") {
		t.Fatalf("streak must restart after a reset: %v", lines)
	}
}

// TestNightReportMultiNightFlush: a jump across several tracked nights
// flushes each in order, once.
func TestNightReportMultiNightFlush(t *testing.T) {
	var r nightReport
	r.record(marker(2, sim.ConsolidationAccepted, ""))
	r.record(marker(1, sim.ConsolidationRejected, "unparseable"))
	lines := r.flushBefore(9)
	if len(lines) != 2 ||
		!strings.Contains(lines[0], "night 1") || !strings.Contains(lines[1], "night 2") {
		t.Fatalf("lines = %v, want night 1 then night 2", lines)
	}
}

// TestNightReportSilentOnBoot: an empty report flushes nothing — the replica
// is snapshot-seeded so absorb never replays history, and a fresh mind must
// not spew summaries for nights it never watched (FR-006).
func TestNightReportSilentOnBoot(t *testing.T) {
	var r nightReport
	if lines := r.flushBefore(30); lines != nil {
		t.Fatalf("fresh report flushed: %v", lines)
	}
}

// TestNightReportAbsorbHook: the mind's absorb path feeds the counters from
// live markers and closes nights when a later night's tick is observed.
func TestNightReportAbsorbHook(t *testing.T) {
	m := worldmap.Generate(42, 64, 64)
	md := &Mind{
		social:    &fakeSocial{},
		replica:   sim.NewState(42, m),
		m:         m,
		narrQ:     make(chan narrJob, 8),
		narrRetry: make(chan narrCarry, 1),
	}

	b, _ := json.Marshal(marker(1, sim.ConsolidationRejected, sim.ConsolidationReasonTruncated))
	md.absorb([]store.Event{{Tick: 80000, Type: "agent.consolidated", Payload: b}})
	if md.nightRep.nights[1] == nil || md.nightRep.nights[1].attempted() != 1 {
		t.Fatalf("marker not tallied: %+v", md.nightRep.nights)
	}

	// A tick from night 2 (day 2) closes night 1.
	wb, _ := json.Marshal(sim.AgentPayload{Agent: sim.Ref(0)})
	md.absorb([]store.Event{{Tick: 86400 + 100, Type: "agent.woke", Payload: wb}})
	if md.nightRep.nights[1] != nil {
		t.Error("later-night tick did not flush night 1")
	}
	if md.nightRep.streak != 1 {
		t.Errorf("streak = %d, want 1 (one fully-failed attempted night)", md.nightRep.streak)
	}
}
