# Phase 0 Research: Paused Authoring Chain-Completion

No NEEDS CLARIFICATION markers existed in the Technical Context — decision-6
and the TASK-77 drift audit pre-answered the scope questions. This document
records the design decisions and the code-verified grounding behind them
(all pins re-verified against the working tree on 2026-07-24, post-31646b0).

## D1: Arm on `metatron.nudged`, not on the per-target memory events

**Decision**: the new `absorb()` case matches `metatron.nudged` and arms each
index in its `Targets` slice, gated on `md.replica.Paused`.

**Rationale**: the landing batch is `[metatron.nudged, agent.memory_added × N]`
(internal/metatron/turn.go:410-439, `landNudgeBatch`). `metatron.nudged`
carries `Targets []int` (internal/sim/metatron.go:40) — one event, explicit
targets, and its `Seq` is the correct causality edge to record as the arming
stimulus (`pendingSeq`, mind.go:243-250: "the arming stimulus event — the
causality edge recorded on the eventual cog.thought"). Order within the batch
is irrelevant to correctness: `run()` calls `absorb(batch)` for the whole batch
before `plan()` (mind.go:193-195), so the dream memories are already in the
replica when the planner prompt snapshots.

**Alternatives considered**:
- *Arm on `agent.memory_added` with `Origin == OriginOmen`*: rejected —
  OriginOmen is shared by miracle grounding memories (internal/sim/memory.go:46
  "a delivered omen/dream/miracle"), which would silently widen the wake to
  landings decision-6 did not bless, and the memory event's Seq is one step
  removed from the nudge that caused it.
- *Arm unconditionally (running too) *: rejected — spec FR-003 / task AC #4:
  unpaused behavior byte-identical; running-world nudge cadence is today's.

## D2: Paused verdict is a pure `cognition.RoutePaused`, checked before speed math

**Decision**: add `RoutePaused(dc DecisionClass, secondsPerPoint float64) Verdict`
to internal/cognition/route.go (Allow=true, PredictedDriftTicks=0, wall-ms
still predicted, Arithmetic naming the paused state). `routeVerdict`
(internal/mind/telemetry.go:61-71) consults `md.replica.Paused` immediately
after `ClassFor` — **before** the `tps <= 0` branch, so paused wins even on a
world whose set speed is uncapped max (spec US2 scenario 3: a frozen world does
not drift, whatever the set speed).

**Rationale**: spec 007 put all routing arithmetic in the pure cognition
package ("pure arithmetic … no model, no randomness, no wall-clock reads",
route.go:17-22); the paused rule is more arithmetic, so it lives there and
stays independently table-testable. The mind supplies the paused fact from its
event-reduced replica (`State.Paused`, internal/sim/state.go:25, reduced from
`clock.paused`/`clock.resumed` at state.go:337-340), which the mind provably
receives — the daemon registers `md.Observe` on the loop's notify fan-out
(internal/daemon/daemon.go:254) and `absorb()` applies every event to the
replica (mind.go:206-211).

**Arithmetic string**: `"%dpt x %.1fs/pt while paused = 0 ticks <= budget %d"`
— same house style as Route's existing strings, states drift 0 ≤ budget, and
names the paused state (FR-005). Exact format is pinned in
contracts/recorded-events.md.

**Alternatives considered**:
- *Extend `Route` with a `paused bool` parameter*: rejected — every existing
  caller (mind, live horizon surface, tests) would need a byte-identical-risk
  edit; a separate function leaves Route untouched (SC-005).
- *Special-case only the planner class*: rejected — spec FR-004 says every
  class; `routeVerdict` is the shared chokepoint for planner, conversation,
  meeting, consolidation, chronicle, so one branch covers all truthfully.

## D3: Paused thoughts predict landing at the snapshot tick

**Decision**: in `newMeta` (internal/mind/telemetry.go:38-54), when the replica
is paused, set `predictedLandTick = snapshotTick` (and skip the tps-based
projection).

**Rationale**: this is FR-004's "predicted drift zero" carried through to the
one other place set-speed prediction leaks while frozen. The planner class is
FutureDated (internal/cognition/registry.go:37), and `plan()` prefixes the
prompt with "your decision will take effect around <predictedLandTick>"
(mind.go:350-352). While frozen, the truth is that the decision lands at the
frozen tick — and `futureDated(now, landing)` already returns "" when
`landing <= now` (internal/mind/prompt.go:63-69), so the fix makes the prompt,
the recorded `cog.thought.PredictedLandTick`, and the gate agree (spec 007
FR-016: "prompt and gate never disagree") with zero new mechanism.

**Alternatives considered**: leaving `predictedLandTick` at set-speed
projection — rejected: it records a future-dated prediction for a thought the
router just allowed on the grounds that no ticks will pass; the record would
contradict the verdict.

## D4: One round is bounded by the existing debounce — no new counter

**Decision**: no bounding mechanism is added. The arm sets `pending[i]`; the
first planned thought sets `lastPlanned[i] = frozen tick`; the 300-game-tick
debounce (`planDebounceTicks`, mind.go:44-49, checked at mind.go:315) cannot
reopen while the clock is frozen, so a second nudge arms `pending` but never
yields a second thought until resume.

**Rationale**: decision-6 blessed exactly this shape ("bounded by
construction… the same shape as decision-4's blessed catch-up round"). Existing
skip semantics compose unchanged: dead/asleep villagers are cleared at
plan-time (mind.go:305-308), meetings hold the arm pending (mind.go:309-311),
single-flight holds (mind.go:318-320).

**Known edge (documented in spec)**: a villager nudged inside a still-closed
debounce window gets the memory but no frozen-tick thought — the designed
bound, legible via the TASK-41 horizon surface, not new mechanism here.

## D5: Determinism and the existing doctrine tests

- `TestPauseStartsNoNewThoughts` (internal/mind/telemetry_test.go:270) stays
  green and stays true: pause alone still starts nothing — the new wake
  requires a nudge landing, which only the angel/operator can cause.
- Replay determinism: both fixes read only event-reduced replica state
  (`Paused`, `Tick`) and registry constants — no wall clock, no randomness.
  Landed frozen-tick effects enter the log as ordinary events; the
  byte-identical replay pattern to follow is
  internal/sim/governor_replay_test.go (which already covers a mid-governed
  pause).
- The live horizon surface (spec 037) intentionally keeps showing set-speed
  verdicts while paused — it describes the world's running posture, not the
  paused-authoring exception; changing it is display-layer scope this feature
  does not touch (candidate follow-up if classroom use shows confusion).

## Verified grounding pins (2026-07-24)

| Fact | Pin |
|------|-----|
| Metatron chat has no pause gate | internal/ipc/server.go:334-355 |
| Nudge landing batch shape (`metatron.nudged` + dream memories) | internal/metatron/turn.go:410-439 |
| `MetatronNudgedPayload.Targets []int` | internal/sim/metatron.go:37-43 |
| absorb arm switch (wake stimuli) | internal/mind/mind.go:206-241 |
| arm() + pendingSeq causality edge | internal/mind/mind.go:243-250 |
| planDebounceTicks = 300 game ticks; debounce check | internal/mind/mind.go:44-49, 315-317 |
| plan() skip semantics (dead/asleep/meeting/in-flight) | internal/mind/mind.go:301-320 |
| routeVerdict computes at set speed | internal/mind/telemetry.go:61-71 |
| newMeta set-speed land prediction | internal/mind/telemetry.go:38-54 |
| futureDated no-ops when landing ≤ now | internal/mind/prompt.go:63-69 |
| Pure Route + Verdict.Arithmetic house style | internal/cognition/route.go |
| planner FutureDated, 3pt/1200t | internal/cognition/registry.go:37 |
| `State.Paused` reduced from clock events | internal/sim/state.go:25, 337-340 |
| Mind observes the full notify fan-out | internal/daemon/daemon.go:254 |
| Pause doctrine tests to keep green | internal/mind/telemetry_test.go:218-333 |
