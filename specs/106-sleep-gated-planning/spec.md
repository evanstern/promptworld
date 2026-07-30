# Feature Specification: Sleep-gated planning — stop scheduling sleeping villagers

**Feature Branch**: `task-175-sleep-gated-planning`

**Created**: 2026-07-30

**Status**: Draft

**Input**: TASK-175 — playtest-1 evidence: 905 of 1,486 `agent.intent_rejected`
events were "X is asleep" (Ash 131, Birch 127, Cedar 124, Oak 110, …), each a
planner round-trip (~37 s avg wall on local gemma) whose output was dead on
arrival at the landing ladder. Card ACs: (1) sleeping villagers do not consume
planner calls; (2) a soak shows "is asleep" rejections near zero from the
905/29-game-days baseline (~31/game-day).

## Diagnosis (grounded)

The arm/enqueue-time gate ALREADY exists and has since TASK-7:
`mind.plan()` skips `a.Dead || a.Asleep` agents and clears their pending
trigger (internal/mind/mind.go, plan()). The 905 rejections leak through the
two windows AFTER enqueue, which no gate covers today:

1. **Queue-wait staleness** — a job enqueued while the agent was awake sits in
   `planQ` behind the single-flight serialized planner worker (~37 s per call,
   up to 8 jobs deep after a shared trigger like `sim.night_started` arming
   every agent at once). By the time the worker dequeues it, the agent has
   reflex-slept. The full model call is then spent on a dead thought.
2. **In-flight staleness** — the agent reflex-sleeps DURING its own ~37 s call
   (systematic at night: an idle tired agent sleeps precisely while its
   nightfall-armed thought is being generated). The call lands via
   `Loop.InjectIntent` and `rungUnavailable` (internal/sim/landing.go:200)
   rejects it: "X is asleep".

Wake re-arm already exists and is verified: absorb's `agent.woke` case arms the
planner (mind.go absorb), and every executor wake path — dawn rest, hunger
emergency, cold emergency (spec 064 US4, `wakeReason`) — emits `agent.woke`.
The one wake with no `agent.woke` event is `gru.attacked` (reducer arm sets
`Asleep = false` directly); the replica applies the same arm, so any mirror of
replica state stays correct there too.

## Decision — gate placement

**Chosen: mind-side, two layers, off-log; the landing ladder stays the
authoritative, byte-unchanged backstop.**

- **Layer 1 — pre-submit (dequeue-time) gate** in `runPlan`: the last moment
  before the model call is spent. Closes window 1 completely.
- **Layer 2 — in-flight cancel**: absorbing `agent.slept` (or `agent.died`)
  for an agent with a planner call in flight cancels that call's context.
  Closes window 2 (minus sub-second absorb lag).

Rejected alternatives:

- **Cadence arm-time only** — already implemented (TASK-7); the evidence shows
  it insufficient: both leak windows open after arming.
- **Pre-landing check in the tool handlers** (handlers.go, before
  `InjectIntent`) — saves only the rejection event, not the call (the wall
  time is already spent by landing time); duplicates a sim gate mind-side
  against the "door decides" doctrine (spec 017); and carries a reverse race
  (mirror momentarily stale ⇒ a valid landing dropped). Cancel dominates it:
  it saves the remaining wall time AND prevents the landing.
- **Sim/reducer-side scheduler change** — the mind is the scheduler; the sim's
  ladder rung is correct behavior and replay-load-bearing. No reducer change
  is needed or wanted (determinism doctrine, spec 092): mind-side gating only
  changes WHICH recorded events get injected, and cog.* rows are reducer
  no-ops.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Planner throughput goes to the awake (Priority: P1)

As a player on a slow local model, I want my limited LLM throughput spent on
villagers who are awake and can act on the plan, so the village thinks instead
of burning 37-second calls on sleepers.

**Independent Test**: unit-level — enqueue a plan job, put the agent to sleep
via the replica, drive the worker: zero model calls.

**Acceptance Scenarios**:

1. **Given** a plan job enqueued while its agent was awake, **When** the agent
   is asleep (per the mind's mirror) at the moment the worker dequeues it,
   **Then** no model call is made, no tool-use loop runs, and one terminal
   `cog.outcome` (`suppressed`, reason naming sleep) records the skip.
2. **Given** a planner call in flight, **When** the absorb goroutine applies
   `agent.slept` for that agent, **Then** the call's context is cancelled, no
   intent reaches `InjectIntent`, and the terminal outcome's reason is
   attributable to the sleep-cancel (distinct from a plain timeout).
3. **Given** a dead agent's job in either window, **Then** the same gate
   applies (unavailability = asleep OR dead, mirroring `rungUnavailable`).

---

### User Story 2 - Waking re-arms planning promptly (Priority: P2)

As a villager woken mid-night by a cold emergency (spec 064 US4), I want
planner attention promptly after waking — the gate must never starve post-wake
planning.

**Independent Test**: sleep → skip → wake → next plan() enqueues (debounce
permitting).

**Acceptance Scenarios**:

1. **Given** an agent whose queued thought was skipped asleep, **When**
   `agent.woke` is absorbed, **Then** the existing wake trigger arms the
   planner and the next `plan()` pass enqueues a fresh job (the mirror shows
   awake from the same batch that woke it).
2. **Given** nightly consolidation (triggered BY `agent.slept`, its own queue),
   **Then** it is untouched by both layers — sleeping villagers still dream.

---

### User Story 3 - Auditable trail (Priority: P3)

As an operator auditing the event trail, I want every skipped or cancelled
thought to terminate in exactly one recorded, countable outcome (the FR-015
no-silent-failure doctrine), so the soak's before/after arithmetic is derivable
from the log alone.

**Acceptance Scenarios**:

1. **Given** a soak's event log, **Then** sleep-skips are countable as
   `cog.outcome{suppressed}` rows with a sleep reason, and sleep-cancels as
   terminal outcomes with a sleep-cancel reason — no thought vanishes.

---

### Edge Cases

- **`gru.attacked` wake** (no `agent.woke` event): the replica applies the same
  reducer arm, so the mirror reads awake; the gate never blocks the victim.
  (That path arms no planner trigger today — pre-existing, out of scope.)
- **Paused-world Guardian nudge on a sleeping villager**: `plan()` already
  clears pending for asleep agents — unchanged.
- **Meeting suppression** (`sim.AtMeeting`): unchanged, orthogonal.
- **Mirror lag** (an `agent.slept` batch not yet absorbed at dequeue/landing):
  the untouched `rungUnavailable` ladder rung remains the authoritative
  backstop; a rare residual rejection is correct behavior, not a defect.
- **Sleep-then-quick-wake during a call**: the cancel may kill a thought whose
  agent re-woke moments later; the wake trigger re-arms a fresh (and fresher)
  thought — accepted.
- **Conversation scenes / narrator / embedder**: out of scope; scenes already
  gate on `Asleep` at scene start (convo.go) and land through the social door,
  not the intent ladder.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The mind maintains a worker-visible per-agent unavailability
  mirror (asleep OR dead), updated by the absorb goroutine after each applied
  batch (the existing `md.tick` atomic-mirror pattern). Workers never read the
  absorb-owned replica.
- **FR-002**: `runPlan` consults the mirror at dequeue, BEFORE emitting
  `cog.thought` or invoking the tool-use loop. Unavailable ⇒ no model call; a
  terminal `cog.outcome` with `Outcome: suppressed` and a reason naming sleep
  (e.g. "asleep at dequeue") is recorded through the telemetry door; the
  per-agent in-flight flag is released; no re-arm (the wake trigger owns
  resumption).
- **FR-003**: The sleep-skip does NOT bump the spec-037 router-suppression
  counters (`RecordSuppression`) — the horizon surface's `SuppressedCount`
  keeps meaning "router suppressed". The skip is distinguished by its recorded
  reason alone.
- **FR-004**: Absorbing `agent.slept` or `agent.died` for an agent whose
  planner call is in flight cancels that call's context (per-agent cancel
  slot, race-safe). The loop terminates without landing an intent; the
  terminal outcome (existing `unusable` vocabulary) carries a reason
  attributable to the sleep/death-cancel, distinct from `callTimeout`.
  Cancellation targets ONLY the planner slot — consolidation, narrator,
  meeting, reconcile, and scene workers are untouched.
- **FR-005**: A cancelled call must not trigger the loop's transport retry nor
  be recorded as provider failure; the latency estimator must not adopt
  sleep-cancelled wall times as spikes (verify against the estimator's
  observation path; adjust only if a real poisoning path exists).
- **FR-006**: The enqueue-time gate in `plan()` and the landing ladder
  (`rungUnavailable`) are byte-unchanged. Zero diff under `internal/sim/`.
- **FR-007**: No new event types, no new outcome vocabulary, no payload-schema
  changes: reuse `OutcomeSuppressed` / `OutcomeUnusable` with reason strings.
  Reducer, replay, and the chronicle digest catalog are untouched.
- **FR-008**: Tests alongside code: dequeue-skip (SC-001), in-flight cancel
  (SC-002), wake re-arm regression (SC-005), consolidation-untouched,
  dead-agent parity, and mirror-update-on-batch coverage.

## Success Criteria *(mandatory)*

- **SC-001**: Unit: a queued job whose agent slept before dequeue produces
  zero `runLoop` invocations and exactly one `cog.outcome{suppressed}` with a
  sleep reason.
- **SC-002**: Unit: `agent.slept` absorbed mid-call cancels the in-flight
  planner context; no `InjectIntent` lands; the terminal outcome's reason is
  sleep-attributable.
- **SC-003**: Soak (card AC #2): a seeded soak (≥ 3 game-days, measurement-run
  dials) shows "is asleep" `agent.intent_rejected` at ≤ 1 per game-day —
  near zero against the 905/29 ≈ 31/game-day baseline (≥ 97% reduction) — and
  zero planner `cog.thought` rows for agents asleep at submit. Counts recorded
  on the board task.
- **SC-004**: `git diff` shows no change under `internal/sim/`;
  `go test -race ./...` green.
- **SC-005**: Regression: after `agent.woke`, the agent's next planner call
  proceeds (debounce permitting), and consolidation still runs on
  `agent.slept`.

## Assumptions

- `sim.AgentCount` (8) fits a single atomic word for the mirror; the exact
  representation (bitmask vs per-agent atomics) is the implementer's choice.
- The soak baseline rate is playtest-1's 905 over 29 game-days (~31.2/day);
  the soak need not run 29 days — per-game-day rate is the comparison unit.
- Implementation tier: **Opus 4.8** (constitution P.V — scheduling logic in
  `internal/mind` orchestration, named explicitly; recorded on the board card).

## Open questions

None — gate placement, wake re-arm, asleep-time-legitimate cognition
(consolidation), and determinism implications were all resolvable from the
code and doctrine (see Diagnosis and Decision above).
