package mind

// Spec 106 (sleep-gated planning) tests: the two mind-side, off-log layers
// that stop planner throughput being spent on sleeping (or dead) villagers.
// Layer 1 — the dequeue gate at the top of runPlan against the absorb-
// maintained unavailability mirror (SC-001, FR-001/FR-002/FR-003). Layer 2 —
// the per-agent in-flight cancel slot absorb fires on agent.slept /
// agent.died (SC-002, FR-004). Plus the wake-resumption regression (SC-005)
// and consolidation-untouched coverage (US2 AC-2). The landing ladder
// (internal/sim/landing.go rungUnavailable) is byte-unchanged — these tests
// exercise only the mind side.

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/toolloop"
)

// sleptAt / diedAt / wokeAt build the absorb-side events the reducer and the
// mirror both consume (payload shapes: state.go Apply arms).
func sleptAt(t *testing.T, tick int64, agent int) store.Event {
	t.Helper()
	b, err := json.Marshal(sim.AgentPayload{Agent: sim.Ref(agent)})
	if err != nil {
		t.Fatal(err)
	}
	return store.Event{Tick: tick, Type: "agent.slept", Payload: b}
}

func wokeAt(t *testing.T, tick int64, agent int) store.Event {
	t.Helper()
	b, err := json.Marshal(sim.AgentPayload{Agent: sim.Ref(agent)})
	if err != nil {
		t.Fatal(err)
	}
	return store.Event{Tick: tick, Type: "agent.woke", Payload: b}
}

func diedAt(t *testing.T, tick int64, agent int) store.Event {
	t.Helper()
	b, err := json.Marshal(sim.DiedPayload{Agent: sim.Ref(agent), Cause: "starvation"})
	if err != nil {
		t.Fatal(err)
	}
	return store.Event{Tick: tick, Type: "agent.died", Payload: b}
}

// --- Layer 1: the dequeue gate (SC-001, FR-002/FR-003) ---

// TestRunPlanSkipsAsleepAtDequeue is SC-001: a job whose agent slept while it
// sat in the queue produces ZERO runLoop invocations (no model call, no
// cog.thought) and exactly one terminal cog.outcome{suppressed} with a sleep
// reason; planInFlight is released and nothing re-arms (the wake trigger owns
// resumption).
func TestRunPlanSkipsAsleepAtDequeue(t *testing.T) {
	lm := newLoopMind(t)
	job := lm.newJob(0) // enqueued while awake
	lm.md.replica.Agents[0].Asleep = true
	lm.md.storeUnavail() // the agent.slept batch absorbed before dequeue

	invoked := false
	lm.md.runLoop = func(ctx context.Context, j toolloop.Job) (toolloop.Result, error) {
		invoked = true
		return toolloop.Result{Term: toolloop.TermModelDone}, nil
	}
	lm.md.planInFlight[0].Store(true) // as plan() sets before enqueue
	lm.md.runPlan(job)

	if invoked {
		t.Error("runLoop invoked for an asleep agent — the model call was spent (SC-001)")
	}
	if n := lm.countByType(t)["cog.thought"]; n != 0 {
		t.Errorf("cog.thought = %d, want 0 (skip precedes the thought record)", n)
	}
	outs := lm.outcomesFor(t, job.meta.job)
	if len(outs) != 1 || outs[0].Outcome != sim.OutcomeSuppressed {
		t.Fatalf("outcomes = %+v, want exactly one suppressed (FR-015)", outs)
	}
	if outs[0].Reason != "asleep at dequeue" {
		t.Errorf("reason = %q, want %q", outs[0].Reason, "asleep at dequeue")
	}
	if lm.md.planInFlight[0].Load() {
		t.Error("planInFlight not released by the skip path")
	}
	if len(lm.md.rearm) != 0 {
		t.Error("a sleep-skip must not rearm (the wake trigger owns resumption)")
	}
}

// TestRunPlanSkipsDeadAtDequeue: unavailability parity (spec US1 AC-3) — a
// dead agent's queued job skips through the same gate with a death reason,
// and dead wins over asleep (rungUnavailable's frozen ordering).
func TestRunPlanSkipsDeadAtDequeue(t *testing.T) {
	lm := newLoopMind(t)
	job := lm.newJob(0)
	lm.md.replica.Agents[0].Dead = true
	lm.md.replica.Agents[0].Asleep = true // dead takes precedence
	lm.md.storeUnavail()

	lm.md.runLoop = func(ctx context.Context, j toolloop.Job) (toolloop.Result, error) {
		t.Error("runLoop invoked for a dead agent")
		return toolloop.Result{Term: toolloop.TermModelDone}, nil
	}
	lm.md.runPlan(job)

	outs := lm.outcomesFor(t, job.meta.job)
	if len(outs) != 1 || outs[0].Outcome != sim.OutcomeSuppressed {
		t.Fatalf("outcomes = %+v, want exactly one suppressed", outs)
	}
	if outs[0].Reason != "dead at dequeue" {
		t.Errorf("reason = %q, want %q (dead before asleep)", outs[0].Reason, "dead at dequeue")
	}
}

// TestSleepSkipDoesNotBumpSuppressionCounter is FR-003: the dequeue skip must
// NOT count as a router suppression — the horizon surface's SuppressedCount
// keeps meaning "router suppressed". It drives the skip against the spec-037
// countingOrch seam fake (telemetry_test.go), whose live wiring is positively
// pinned by TestEmitSuppressedReachesSeam — so this non-bump cannot rot
// silently if the seam moves.
func TestSleepSkipDoesNotBumpSuppressionCounter(t *testing.T) {
	lm := newLoopMind(t)
	gate := &countingOrch{}
	lm.md.orch = gate
	job := lm.newJob(0)
	lm.md.replica.Agents[0].Asleep = true
	lm.md.storeUnavail()
	lm.md.runPlan(job)

	gate.mu.Lock()
	n := len(gate.counts)
	gate.mu.Unlock()
	if n != 0 {
		t.Errorf("sleep-skip bumped RecordSuppression (%v), want no bumps (FR-003)", gate.counts)
	}
	if outs := lm.outcomesFor(t, job.meta.job); len(outs) != 1 || outs[0].Outcome != sim.OutcomeSuppressed {
		t.Fatalf("outcomes = %+v, want the suppressed skip (the counter stayed put for the right run)", outs)
	}
}

// TestUnavailMirrorTracksBatches is the FR-001 mirror coverage: absorb
// refreshes the mirror at batch end, so slept/woke/died flip it — and a wake
// is visible from the SAME batch that woke the agent (US2 AC-1's premise).
func TestUnavailMirrorTracksBatches(t *testing.T) {
	lm := newLoopMind(t)
	md := lm.md
	if r := md.unavailReason(0); r != "" {
		t.Fatalf("agent 0 unavailable at genesis: %q", r)
	}

	md.absorb([]store.Event{sleptAt(t, 0, 0)})
	if r := md.unavailReason(0); r != "asleep at dequeue" {
		t.Errorf("after agent.slept: reason = %q, want asleep", r)
	}

	md.absorb([]store.Event{wokeAt(t, 0, 0)})
	if r := md.unavailReason(0); r != "" {
		t.Errorf("after agent.woke: reason = %q, want available", r)
	}

	md.absorb([]store.Event{diedAt(t, 0, 1)})
	if r := md.unavailReason(1); r != "dead at dequeue" {
		t.Errorf("after agent.died: reason = %q, want dead", r)
	}
	if r := md.unavailReason(0); r != "" {
		t.Errorf("agent 0 caught agent 1's death: %q", r)
	}
}

// --- Layer 2: the in-flight cancel (SC-002, FR-004) ---

// cancelFixture runs runPlan on its own goroutine with a runLoop that blocks
// until its context is cancelled, absorbs the given batch mid-call, then
// waits for runPlan to finish before returning (fully serialized: the absorb
// completes while the loop is still blocked, and runPlan's post-loop emission
// races nothing).
func cancelFixture(t *testing.T, lm *loopMind, job planJob, batch []store.Event) {
	t.Helper()
	running := make(chan struct{})
	proceed := make(chan struct{})
	lm.md.runLoop = func(ctx context.Context, j toolloop.Job) (toolloop.Result, error) {
		close(running)
		select {
		case <-ctx.Done():
		case <-time.After(10 * time.Second):
			t.Error("in-flight call never cancelled")
		}
		<-proceed // hold until the absorb (and its assertions' setup) completes
		return toolloop.Result{Term: toolloop.TermCtxDone}, ctx.Err()
	}
	done := make(chan struct{})
	go func() { lm.md.runPlan(job); close(done) }()
	<-running
	lm.md.absorb(batch) // fires the cancel slot mid-batch
	close(proceed)
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("runPlan did not return after the cancel")
	}
}

// TestRunPlanCancelledOnSleepMidCall is SC-002: agent.slept absorbed while
// the agent's planner call is in flight cancels the call's context; nothing
// lands through InjectIntent, and the terminal outcome's reason is
// sleep-attributable — distinct from a plain callTimeout ("loop: context
// ended"). The same absorb ALSO triggers nightly consolidation (US2 AC-2:
// its own queue, untouched by both layers — sleeping villagers still dream).
func TestRunPlanCancelledOnSleepMidCall(t *testing.T) {
	lm := newLoopMind(t)
	job := lm.newJob(0)
	// A night-1 sleep tick, so ConsolidationDue holds (NightIndex > 0) and
	// the empty-buffer marker proves the consolidation path still ran.
	cancelFixture(t, lm, job, []store.Event{sleptAt(t, 0, 0)})

	if n := lm.plannerIntents(t); n != 0 {
		t.Errorf("cancelled cognition landed %d planner intent(s), want 0", n)
	}
	outs := lm.outcomesFor(t, job.meta.job)
	if len(outs) != 1 || outs[0].Outcome != sim.OutcomeUnusable {
		t.Fatalf("outcomes = %+v, want exactly one unusable (FR-015)", outs)
	}
	if outs[0].Reason != "cancelled in flight: agent slept" {
		t.Errorf("reason = %q, want the sleep-cancel attribution (distinct from callTimeout)", outs[0].Reason)
	}
	// FR-005 at the mind level: the cancel surfaced no transport-retry marker
	// (cog.outcome{retried}) — toolloop's TestRetryContextDoneNeverRetried
	// pins the loop side (ctx_done never retries, never feeds the estimator).
	if n := lm.countByType(t)["cog.thought"]; n != 1 {
		t.Errorf("cog.thought = %d, want 1 (the call was attempted, then cancelled)", n)
	}
	// US2 AC-2: the same agent.slept still drove consolidation (empty buffer
	// at genesis ⇒ the skipped_empty marker closes the night).
	consol := lm.events(t, "agent.consolidated")
	if len(consol) != 1 {
		t.Fatalf("agent.consolidated events = %d, want 1 (consolidation untouched by the cancel)", len(consol))
	}
	var p sim.ConsolidatedPayload
	if err := json.Unmarshal(consol[0].Payload, &p); err != nil {
		t.Fatal(err)
	}
	if p.Outcome != sim.ConsolidationSkippedEmpty {
		t.Errorf("consolidation outcome = %q, want skipped_empty (the empty genesis buffer)", p.Outcome)
	}
	// The slot is cleared: a later sleep fires nothing.
	if lm.md.planCancel[0].Load() != nil {
		t.Error("cancel slot not cleared after the loop returned")
	}
}

// TestRunPlanCancelledOnDeathMidCall: FR-004 parity — agent.died absorbed
// mid-call cancels with a death-attributable reason, and only the victim's
// slot fires (planner slot only; per-agent).
func TestRunPlanCancelledOnDeathMidCall(t *testing.T) {
	lm := newLoopMind(t)
	job := lm.newJob(2)
	cancelFixture(t, lm, job, []store.Event{diedAt(t, 0, 2)})

	if n := lm.plannerIntents(t); n != 0 {
		t.Errorf("cancelled cognition landed %d planner intent(s), want 0", n)
	}
	outs := lm.outcomesFor(t, job.meta.job)
	if len(outs) != 1 || outs[0].Outcome != sim.OutcomeUnusable {
		t.Fatalf("outcomes = %+v, want exactly one unusable", outs)
	}
	if outs[0].Reason != "cancelled in flight: agent died" {
		t.Errorf("reason = %q, want the death-cancel attribution", outs[0].Reason)
	}
}

// TestCancelSlotIsPerAgent: absorbing another agent's sleep must NOT cancel
// this agent's in-flight call — the slot is per-agent (FR-004's race-safety
// is scoped, not global).
func TestCancelSlotIsPerAgent(t *testing.T) {
	lm := newLoopMind(t)
	job := lm.newJob(0)
	running := make(chan struct{})
	lm.md.runLoop = func(ctx context.Context, j toolloop.Job) (toolloop.Result, error) {
		close(running)
		select {
		case <-ctx.Done():
			t.Error("agent 0's call cancelled by agent 3's sleep")
			return toolloop.Result{Term: toolloop.TermCtxDone}, ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
		return toolloop.Result{Term: toolloop.TermModelDone}, nil
	}
	done := make(chan struct{})
	go func() { lm.md.runPlan(job); close(done) }()
	<-running
	lm.md.cancelInFlightPlan(3, "cancelled in flight: agent slept") // no slot: no-op
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("runPlan did not return")
	}
}

// TestPlainTimeoutKeepsGenericReason: a ctx_done WITHOUT an absorb-side cancel
// (the callTimeout path) keeps loopFailReason's generic text — the sleep
// attribution never claims a timeout it didn't cause.
func TestPlainTimeoutKeepsGenericReason(t *testing.T) {
	lm := newLoopMind(t)
	job := lm.newJob(0)
	lm.md.runLoop = func(ctx context.Context, j toolloop.Job) (toolloop.Result, error) {
		return toolloop.Result{Term: toolloop.TermCtxDone}, context.DeadlineExceeded
	}
	lm.md.runPlan(job)

	outs := lm.outcomesFor(t, job.meta.job)
	if len(outs) != 1 || outs[0].Outcome != sim.OutcomeUnusable {
		t.Fatalf("outcomes = %+v, want exactly one unusable", outs)
	}
	if outs[0].Reason != "loop: context ended" {
		t.Errorf("reason = %q, want the generic %q (no cancel slot fired)", outs[0].Reason, "loop: context ended")
	}
}

// --- Wake resumption (SC-005) + the enqueue-time gate regression ---

// TestWakeRearmsPastGate is SC-005: after a skip, absorbing agent.woke arms
// the planner AND flips the mirror in the same batch, so the next plan() pass
// enqueues a fresh job (debounce permitting) and that job proceeds past the
// dequeue gate to a real cognition.
func TestWakeRearmsPastGate(t *testing.T) {
	lm := newLoopMind(t)
	md := lm.md
	md.planQ = make(chan planJob, sim.AgentCount)
	// The bare fixture leaves nextDue zeroed (everyone cadence-due); push the
	// others out so the enqueue below isolates the WAKE trigger.
	for i := range md.nextDue {
		md.nextDue[i] = int64(1 << 40)
	}

	// Sleep agent 0 and skip its queued thought (layer 1).
	md.absorb([]store.Event{sleptAt(t, 0, 0)})
	skipped := lm.newJob(0)
	md.planInFlight[0].Store(true)
	md.runPlan(skipped)
	if outs := lm.outcomesFor(t, skipped.meta.job); len(outs) != 1 || outs[0].Outcome != sim.OutcomeSuppressed {
		t.Fatalf("setup: skip outcomes = %+v", outs)
	}

	// Wake at a tick past the debounce floor: the SAME batch arms pending and
	// refreshes the mirror.
	wakeTick := int64(planDebounceTicks + 100)
	md.absorb([]store.Event{wokeAt(t, wakeTick, 0)})
	if !md.pending[0] {
		t.Fatal("agent.woke did not arm the planner (wake trigger regression)")
	}
	if r := md.unavailReason(0); r != "" {
		t.Fatalf("mirror still unavailable after the waking batch: %q", r)
	}

	// The next plan() pass enqueues a fresh job…
	md.plan()
	if len(md.planQ) != 1 {
		t.Fatalf("plan() enqueued %d job(s) after wake, want 1", len(md.planQ))
	}
	fresh := <-md.planQ

	// …and that job proceeds past the dequeue gate into a real cognition.
	ran := false
	md.runLoop = func(ctx context.Context, j toolloop.Job) (toolloop.Result, error) {
		ran = true
		return toolloop.Result{Term: toolloop.TermModelDone}, nil
	}
	md.runPlan(fresh)
	if !ran {
		t.Error("post-wake job did not reach the loop — the gate starved wake resumption")
	}
	if n := lm.countByType(t)["cog.thought"]; n != 1 {
		t.Errorf("cog.thought = %d, want 1 (the fresh post-wake cognition)", n)
	}
}

// TestPlanEnqueueGateUnchanged pins the pre-existing enqueue-time gate
// (FR-006): plan() still skips an asleep agent and clears its pending
// trigger — layer 1 is ADDITIVE, not a replacement.
func TestPlanEnqueueGateUnchanged(t *testing.T) {
	lm := newLoopMind(t)
	md := lm.md
	md.planQ = make(chan planJob, sim.AgentCount)
	md.replica.Agents[0].Asleep = true
	md.pending[0] = true
	md.pendingSeq[0] = 7

	md.plan()

	if len(md.planQ) != 0 {
		t.Errorf("plan() enqueued %d job(s) for an asleep agent, want 0", len(md.planQ))
	}
	if md.pending[0] {
		t.Error("plan() left the asleep agent's pending trigger armed")
	}
}
