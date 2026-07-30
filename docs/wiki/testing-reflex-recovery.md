---
name: testing-reflex-recovery
description: Spec-062 reflex-arbitration rungs (PREP gate, day-warmth, night-search, the Sage thrash-regression proof) and the spec-064 needs-conditioned recovery lifecycle (hold/complete/abort/yield). Split out of [[testing-strategy]].
kind: pattern
sources:
  - internal/sim/toolcheck_test.go
  - internal/sim/yield_state_test.go
  - internal/sim/day_warmth_test.go
  - internal/sim/night_search_test.go
  - internal/sim/thrash_regression_test.go
  - internal/sim/reflex_matrix_test.go
  - internal/sim/needs_recovery_test.go
verified_against: d0645811c9783d1248dc65ed0fcf0b37524dd8fd
---

# Reflex-arbitration & recovery suites

**Reflex-arbitration suites** (spec 062, TASK-103, [[reflex-policy]]): the
"instinct yields to intelligence" restructuring is proven per rung group.
`internal/sim/yield_state_test.go` (T001/T005) covers the PREP gate itself:
the yield-window anchor (`Agent.LastMindIntentDone`) is event-sourced,
snapshot-compatible (`omitempty`), and armed ONLY by a non-reflex intent
completion; the gate matrix proves the window suppresses prep and decays,
either danger band (food/warmth/rest) suppresses regardless of the window, a
reflex-sourced completion never arms the window, and a no-planner drive
matches pre-062 behavior except the enumerated danger-band suppressions
(SC-003). `internal/sim/day_warmth_test.go` (US2, T007) covers the new day
warmth rung: seek/relight KNOWN warmth (AS1), build with wood in hand (AS2),
a healthy-warmth day byte-identical to pre-062 (AS3), and
`TestDayWarmthDoesNotChopTheDeviation` — the flagged plan deviation — pinning
that the day rung stops before the night ladder's chop tail.
`internal/sim/night_search_test.go` (US3, T009) covers the bounded night
frontier-search fallback: a cold, nothing-known-or-carried-or-choppable agent
with a reachable frontier searches rather than sleeping cold
(`TestNightSearchFallbackSearchesWhenFrontierReachable`), and an unreachable
frontier still falls through to sleep, today's terminal floor.
`internal/sim/thrash_regression_test.go` (US4, T010, SC-001) is the
deliverable: the scripted world-01 Sage-shape scenario (cold, fed, a planner
`goto_warmth` completing at a fire, larder below stock target) encodes BOTH
proofs in one deterministic test — the OLD flip (absent the gate's inputs,
the reflex forages the agent away from the fire it was just sent to) and the
NEW hold-and-recover (with the gate's inputs live, zero prep intents fire
inside the yield window and the warmth trajectory recovers) — expressed by
nullifying the gate's INPUTS (an unarmed window, healthy warmth) rather than
mutating the immutable `prepYieldTicks`/danger-band constants, the
const-respecting encoding of the pre-062 world. `internal/sim/reflex_matrix_test.go`'s
cold-night truth table is updated for Gap A: the `wood=0`/nothing-known cells
that used to resolve `sleep` now resolve `search` (`TestColdNightReflexMatrix`),
with the no-frontier fall-through to `sleep` still pinned separately.

**Needs-conditioned recovery suites** (spec 064, TASK-104,
[[executor]]/[[reflex-policy]]): `internal/sim/needs_recovery_test.go` proves
the hold/complete/abort/yield lifecycle per user story. US1 (warm_up holds and
completes on warmth): `TestConditionlessIntentByteIdentical` (SC-004, FR-001)
pins that a condition-free `Intent`/`WorkStartedPayload`/`IntentSetPayload`
marshals without any of the new keys; `TestConditionRidesIntentSetPayload`
proves the condition rides the recorded `intent_set` and a malformed need is
dropped at the reducer door; `TestRecoverThenRelease` (SC-001) drives a
conditioned hold to a deterministic threshold-crossing completion tick, twice,
proving no premature `intent_done` and identical completion ticks across
runs; `TestAlreadySatisfiedCompletesImmediately` proves a condition already
met on arrival completes at once with no `work_started` (checked, not
assumed); `TestReplayDeterminismOverRecoverySpan` (FR-008) proves the whole
recover-then-release event log is byte-identical across two runs;
`TestWarmUpResolverAndClamp` pins the single `clampWarmUp`/`ClampWarmUp` clamp
home (default, in-range, below-floor, above-cap) and the `warm_up` resolver's
default/clamped threshold; `TestReflexWarmthRungsIssueConditioned` proves
BOTH the day and night reflex warmth rungs issue `goto_warmth` carrying the
doctrine-default condition; `TestReflexHoldDoesNotArmYieldWindow` and
`TestReflexHoldNoArriveIdleWander` (the enumerated no-LLM behavior change)
prove a reflex-issued hold never arms the spec-062 yield window and never
wanders mid-recovery. US3 (interruptibility, no new immunity):
`TestRecoveryAbortsWhenSourceDead` (SC-003) proves a dead-source hold aborts
with the distinct `agent.recovery_stalled` outcome within the stall-window
bound, stamping the ring `"stalled"`; `TestSurvivalPreemptsRecovery` (SC-003)
proves a higher-priority survival need in its danger band ends a hold at once
(via `preemptsRecovery`'s ladder-order doctrine — food outranks warmth,
never the reverse) rather than letting it start or continue holding;
`TestHoldingIntentOverriddenLikeAnyActive` proves a planner override replaces
a holding recovery exactly as it replaces any active intent (no new
immunity). US2 (the mechanism is generic): `TestSecondConsumerSharesMechanism`
(SC-004) drives a REST-conditioned intent through the identical
`executeAtTarget`→`recoveryHoldEvents` path — satisfied-on-arrival completion
and dead-source abort both — proving the condition machinery is need-generic,
not warm_up-private (sleep itself is deliberately NOT refactored onto it, per
plan R6 — the proof is at the mechanism level). US4 (wake to cold, audit Gap
C): `TestSleeperWakesToColdEmergency` (SC-005) proves a sleeper below
`exposureWakeBelow` with an actionable warmth ladder wakes;
`TestCozySleeperSleepsThrough` and `TestColdWakeIsEmergencyFloorNotDangerBand`
are the controls — a cozy sleeper and a merely-danger-band-cold one (above the
150 floor, below the 350 band) both stay asleep, pinning the wake as an
EMERGENCY floor, not the routine danger band; `TestSleeperNotWokenWhenNothingActionable`
is the churn-bound control. SC-002 end-to-end:
`TestSageWarmUpHeldToThresholdThenReleased` extends the 062 Sage-shape
scenario through a planner `warm_up`, proving the agent holds to the doctrine
threshold with ZERO mid-recovery prep dispatches or wandering, then releases
— the arrive-idle-vacuum dead end-to-end. `internal/sim/toolcheck_test.go`
gains a `"warm_up": 0` entry (arrival triggers the hold, not a timed work
cycle, like its `goto_warmth` sibling).

## Connections

Part of the [[testing-strategy]] suite map (split out during the corpus-spec v2
restructure); see that note for the full layered test picture and links to
sibling suites.
