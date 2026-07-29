---
name: daemon-cognition-calibration
description: Daemon boot step 6b — teaching-posture default, tool-loop config warnings, mind/guardian wiring (bundles/stage/skin), ValidateKinds gate, calibration profile + estimator-state seeding, the estimator persister goroutine
kind: pipeline
sources:
  - internal/daemon/daemon.go
  - internal/daemon/estimator_persist.go
verified_against: a5df40921577bc194478bb29c42af2b10bf11ea8
---

# Daemon boot: teaching posture and calibration seeding

Split from [[daemon-lifecycle]] (full sequence overview there): the second half
of step 6, continuing right after [[daemon-orchestrator-startup]]'s LLM
orchestrator/governor/embedding wiring, still inside the `llm.json`-gated
branch — deriving the teaching-posture boot default, surfacing tool-loop config
warnings, wiring the mind/guardian consumers, and seeding calibration +
persisted estimator state.

## How it works

On a teaching world (spec 039 US1/US3, `w.Manifest.Teaching` —
[[world-save-directory]]), boot also derives and prints the teaching-posture
default: `orch.EstimateForKind(llm.Kind("planner"))`'s live seconds-per-point
feeds [[cognition]]'s `MaxSafeSpeed("planner", est)` for the highest
planner-safe ladder rung, mapped to a `clock.Speed` via
[[game-clock]]'s `SpeedForRate`; `teachingPostureBootLine` prints it in a
calibrated flavor (the planner-serving provider's `CalibratedAt` is set) or
a provisional one that also prompts `promptworld calibrate <world>` — the
pessimistic bootstrap-seeded rung still applies either way, just honestly
labeled. No planner-serving provider means no posture line and no default.
Boot also surfaces the agent tool-use loop's config warnings the same
warn-not-error way as the concurrency knob (`llmCfg.Local.Workers()`'s
`workersWarn`): `llmCfg.Rounds()` (an out-of-range `loop_max_rounds`), both
tiers' `ToolModeResolved()` (an unknown `tool_mode`), and — since spec 025
(TASK-72) — the three per-kind token budgets (`llmCfg.PlannerTokens()`/
`GuardianTurnTokens()`/`ConsolidationTokens()`, an out-of-range
`max_tokens.<key>`) each print one line and
clamp/default rather than aborting boot (TASK-52, [[llm-orchestrator]]). The
normalized round cap and effective budgets then thread into both loop
consumers: `mind.New(..., loopRounds, plannerTokens, consolidationTokens)`
and `guardian.New(orch, loop, loop, ..., loopRounds, guardianTurnTokens)`
(followed by `mt.SetBundles(bundleSet)` — spec 036 hands the boot-frozen
bundle surface to the turn assembly, [[bundle-tools]] — and
`mt.SetStage(w.Manifest.Stage, w.Manifest.CharterPreset)` — spec 046 US2
hands the immutable stage + charter preset from the opened manifest the
same boot-frozen way, so the stage tool ceiling and the stage-1
instruction lock cannot be tampered mid-run, [[curriculum-ladder]] — and
`mt.SetSkin(worldSkin)`, handing the same boot-frozen display skin
`srv.SetSkin` (below) gave the status surface to the guardian turn
assembly's prompts, spec 052 FR-003) —
since spec 029 (US5) the loop is passed twice: once as the `Injector` it
was always passed as, once as the new `LoopControl` seam Guardian's
`pause`/`start`/`adjust_speed` meta tools drive ([[guardian-orders]],
[[sim-loop]]'s `Loop.Do` — the same two-interfaces-one-value pattern
`mind.New(loop, loop)` already used for the mind driver).
Before the orchestrator is built, `cognition.ValidateKinds(llm.Kinds())` is a
hard startup gate: every call kind must resolve to a registered decision class
before a model is ever reachable ([[cognition]]). After it is built,
`cognition.LoadProfile(w.CalibrationPath())` seeds the seconds-per-point
estimators (`orch.SeedCalibration`, which since spec 035 also records each
provider's `calibratedAt` from the profile — [[llm-orchestrator]]); a
missing or unreadable `calibration.json` falls back to pessimistic
bootstrap defaults
(`cognition.BootstrapLocalSecPerPt`/`BootstrapCloudSecPerPt` — fail toward
reflex, never toward stale action), and since spec 035 (FR-001,
contracts/warnings.md §1) both branches print the full
`uncalibratedBootWarning(worldName)` block instead of a one-line hint: the
UNCALIBRATED statement, `cognition.HorizonSummary` evaluated at the
bootstrap seeds (the identical string `promptworld calibrate` itself
prints, FR-006 — [[cognition]]), and the exact `promptworld calibrate
<world>` command to run. The profile-seeded branch is untouched and stays
byte-identical (US2 AC2). After calibration seeding,
`cognition.LoadEstimatorState(w.EstimatorStatePath())` +
`orch.SeedPersisted` raise each provider's seed to any higher persisted
live estimate (TASK-113, max(seed, persisted) — a malformed
`estimator_state.json` downgrades to a warning, never a crash), and a
daemon-side persister goroutine flushes `orch.SnapshotEstimators()` back to
that file every 5 minutes plus once synchronously after `loop.Run(ctx)`
returns, so learned drift survives restarts.
`orch.SetRecalibrateHook(md.RecalibrateSignal)` wires
the drift signal: a provider's estimator breaching its spike-rate threshold lands
as `cog.recalibration_recommended` telemetry.

## Connections

[[cognition]] supplies the startup kind gate, the calibration profile it seeds
into the orchestrator, (spec 035) the `HorizonSummary` the boot warning block
quotes verbatim, and the `MaxSafeSpeed` the teaching-posture default computes
from; [[game-clock]]'s `SpeedForRate` turns that rung into the `clock.Speed`
later applied through the loop's `set_speed` door (in [[daemon-boot-recovery]]'s
step 8). [[guardian-orders]] is what the `LoopControl` seam wired here (spec
029) drives. [[bundle-tools]] is the boot-frozen bundle surface handed to the
turn assembly; [[curriculum-ladder]] is what the `SetStage` handoff serves.
[[llm-orchestrator]] is what the token-budget warnings clamp and what
`SeedCalibration`/`SeedPersisted` write into. See [[daemon-orchestrator-startup]]
for the LLM orchestrator/governor/embedding half of this step, and
[[daemon-boot-recovery]] for the surrounding boot sequence.
