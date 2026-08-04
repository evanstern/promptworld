---
name: ipc-status-signals
description: How ipc.Server composes optional StatusData signals — set_speed's uncalibrated/teaching-posture warnings, the governor/morgue/stage/scenario folds, and horizonClasses
kind: component
sources:
  - internal/ipc/server.go
verified_against: 5761edb18e2b5fb49c6a03a050b0d871f5546c05
---

# IPC server: StatusData signal composition

Split from [[ipc-server]] (session/dispatch mechanics and the miracle door stay
there): the optional per-world signals `ipc.Server` folds into `StatusData` replies
— speed-change warnings, teaching posture, governor/morgue/stage/scenario snapshot
folds, and the horizon-class listing.

## How it works

`set_speed` enforces the speed
policy (TASK-20): `max` is refused with an actionable error whenever the world
has an LLM configured (`llm != nil`) — uncapped ticking is for pure-sim worlds;
the watchable ceiling is 32x ([[game-clock]]); the spec 028 governor changes
nothing here — it never touches uncapped speed, so this refusal is unchanged.
Since spec 035 (FR-002), a `set_speed` reply that clears the `max` refusal
additionally carries `uncalibratedWarning(speed)` on `StatusData.Warning`:
non-empty only when the world has an orchestrator and the requested speed's
ticks/sec would suppress one or more of [[cognition]]'s watched classes
(`SuppressedAt`) evaluated at their CURRENT live estimates, gated to classes
whose serving provider is still bootstrap-seeded (`s.llm.CalibratedAt(name)
== ""` — a calibrated provider's live drift is the governor's signal, not
this one). The warning never blocks the speed change — it composes after the
`max` gate, which is unchanged and still evaluated first — and `status`,
`pause`, `resume` always pass `""` so their replies stay byte-identical.
`replyStatus(id, cmd, speed, warning)` carries the extra parameter through to
every caller; only the `set_speed` case supplies a non-empty value — since
spec 039 that value is `setSpeedWarning(speed)`, not `uncalibratedWarning`
directly.

Since spec 039 (US2/US4, `contracts/posture.md`), a teaching world
(`s.w.Manifest.Teaching`) layers a second, independent advisory on top:
`postureWarning(speed)` fires when the requested speed exceeds the
planner-safe posture — computed live via [[cognition]]'s `MaxSafeSpeed` over
the planner-serving provider's `EstimateForKind` estimate, the SAME call the
daemon's boot default uses ([[daemon-lifecycle]]) — and composes, for every
watched class `LiveHorizon` finds suppressed at the requested speed, the
router's `Verdict.Arithmetic` string verbatim plus a plain-language
`postureConsequence` (`"villagers will stop deep-thinking (reflex only)"` for
`planner`, `"conversations will be skipped"` for `conversation`, `"meetings
fall back to template speeches"` for `meeting`, a generic degrade phrase
otherwise), joined under an `above teaching posture Nx:` prefix. Unlike
`uncalibratedWarning`, this fires for a CALIBRATED teaching world too — that
is the point of a soft cap. `setSpeedWarning(speed)` composes the two,
posture first then uncalibrated, newline-joined when both fire (either may be
empty), and is what the `set_speed` handler now calls instead of
`uncalibratedWarning` directly. `postureStatus()` computes the teaching
world's `*PostureStatus{Rung, Calibrated}` the status-family reply carries
(`StatusData.Posture`, [[ipc-protocol]]) via the identical `MaxSafeSpeed` call
— nil when no provider serves the planner class, so the field stays absent
rather than reporting an ungrounded rung. `statusDataFull` sets it only when
`s.w.Manifest.Teaching`, alongside the existing `Horizon` assignment.

`statusData`/`statusDataFull` fold two optional per-world snapshots into the
`clock`/`llm` sections the same way: the orchestrator's `StatusSnapshot` when
`SetLLM` attached one, and — since spec 028 — the daemon governor's debt
reading through a local `Governor` interface (`GovernorStatus() (debt float64,
jobs int)`, `SetGovernor`, kept narrow like `Angel` so `ipc` never imports
`internal/daemon`); a nil governor (no-LLM world) leaves the clock section's
governor fields at their `omitempty` zero, byte-identical to pre-028
([[cognition]], [[daemon-lifecycle]]). Since spec 044, `statusData` also
copies the loop's run-end posture into the clock section — `Ended`/`EndedDay`
straight from `sim.Status` — where the same `omitempty` discipline keeps a
living world's status bytes unchanged ([[ipc-protocol]], [[morgue]]). Since
spec 046, `statusData` likewise composes the world section's
`Stage`/`StageOverridden` straight from the opened manifest
(`s.w.Manifest.Stage`/`.StageOverridden`, [[curriculum-ladder]]) — additive
`omitempty` again, so a pre-ladder world's bytes are unchanged
([[ipc-protocol]]). Since spec 054, `statusData` also copies the loop's
scenario facts into the world section — `ScenarioExercise`/`ScenarioOutcome`
straight from `cs.ScenarioExercise`/`ScenarioOutcome` (the loop's status
snapshot, so the pair is coherent with `Tick`) — the same `omitempty`
discipline keeps an ambient world's status bytes unchanged
([[ipc-protocol]], [[scenario-machinery]]).

Since spec 037 (`contracts/status-horizon.md`), `statusDataFull` additionally
sets `StatusData.Horizon` via `horizonClasses(cs)` whenever an orchestrator is
attached — so `status`/`pause`/`resume`/`set_speed` all carry it alike, unlike
the `set_speed`-only `Warning`. `horizonClasses` delegates to
[[cognition]]'s `LiveHorizon` at the loop's EFFECTIVE speed
(`cs.Speed.TicksPerSecond()`, post-governor), resolving each watched class's
live estimate through `EstimateForKind` (a class whose kind has no admissible
serving provider is excluded, `ok=false`), and folds in
`s.llm.CalibratedAt(name) != ""` for the `Calibrated` flag and
`s.llm.SuppressionCounts()` for each entry's `SuppressedCount` — a class never
suppressed reads 0 from the map's zero value. Unlike `uncalibratedWarning`
below, which gates OUT calibrated classes, `horizonClasses` INCLUDES them
(research R4): calibration only changes the client's remedy phrasing, never
membership. Returns nil (never an empty slice) when nothing is included, so
`omitempty` keeps the field absent for a no-LLM world.

## Connections

`uncalibratedWarning` reads [[cognition]]'s `SuppressedAt` and
[[llm-orchestrator]]'s `EstimateForKind`/`CalibratedAt`, riding
`StatusData.Warning` ([[ipc-protocol]]), rendered by [[cli-runtime-control]]'s
`setSpeedLine`. `horizonClasses` reads [[cognition]]'s `LiveHorizon` and
[[llm-orchestrator]]'s `EstimateForKind`/`CalibratedAt`/`SuppressionCounts`,
riding `StatusData.Horizon` ([[ipc-protocol]]), rendered by
[[cli-runtime-control]]'s `horizonStatusLines` and [[tui-client]]'s header
badge + guardian-pane `horizonLines`. `postureWarning`/`postureStatus` read
[[cognition]]'s `MaxSafeSpeed` and [[game-clock]]'s `SpeedForRate` (the same
call [[daemon-lifecycle]]'s boot default makes), riding
`StatusData.Warning`/`StatusData.Posture` ([[ipc-protocol]]), rendered by
[[cli-runtime-control]]'s `setSpeedLine`/`postureStatusLine`. The scenario
folds read [[scenario-machinery]]'s `sim.ExerciseOutcome`, riding
`StatusData.ScenarioExercise`/`ScenarioOutcome` ([[ipc-protocol]]), rendered
by [[cli-runtime-control]]'s `scenarioStatusLine` and the [[tui-client]]
exercise tab. See [[ipc-server]] for session dispatch, the miracle door, and
transport mechanics.
