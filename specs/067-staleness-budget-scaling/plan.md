# Implementation Plan: Staleness budget scaling — planning must survive clock speed

**Branch**: `067-staleness-budget-scaling` | **Date**: 2026-07-26 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/067-staleness-budget-scaling/spec.md`

## Summary

The landing gate (`internal/sim/landing.go:rungStale`) compares actual
game-tick staleness to a fixed per-class `BudgetTicks`, so a constant-wall-time
thought is punished linearly harder as clock speed rises — planning is
structurally dead above ~4x on local tiers while every horizon surface reports
it healthy. Fix: reinterpret `BudgetTicks` as the 1x budget (wall-clock
patience) and enforce `BudgetTicks × ticksPerSecond(effective speed at the
landing tick)` at the two *delivery* gates (reducer landing rung + mind-side
convo pre-abort), leaving every *scheduling* gate (Route, RoutePaused,
governor debt, horizon surfaces) untouched. The scale source is the
event-sourced `state.Speed`, so the gate stays a pure function of
event-sourced state and replay determinism holds.

## Technical Context

**Language/Version**: Go (module `promptworld`, toolchain per `go.mod`)

**Primary Dependencies**: stdlib only at the change sites; internal packages
`internal/cognition` (budget doctrine, leaf), `internal/sim` (reducer),
`internal/mind` (mind-side pre-check), `internal/clock` (Speed →
TicksPerSecond)

**Storage**: event-sourced world log (no schema change; reason strings only)

**Testing**: `go test ./...`; table-driven unit tests + reducer replay test
mirroring `internal/sim/governor_replay_test.go`

**Target Platform**: daemon (darwin/linux), no platform-specific code

**Project Type**: single Go module, internal packages

**Performance Goals**: n/a (one float multiply per landing)

**Constraints**: gate must be a pure function of event-sourced state (replay
determinism); 1x behavior bit-identical; `Route`/horizon doctrine untouched

**Scale/Scope**: 2 gate sites + 1 helper + doctrine comment + telemetry string
+ tests + wiki re-pin; ~6 files touched

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Artifact-Grounded Action** — PASS: diagnosis pinned in spec.md from the
  recorded TASK-122 measurement; decisions recorded in research.md; landing
  outcomes remain recorded events.
- **II. One Task, One PR** — PASS: TASK-141 ↔ branch
  `task-141-staleness-budget-scaling` ↔ one PR; spec docs commit to main.
- **III. Gates Over Assertions** — PASS: board mirrors spec via spec-bridge;
  merge-drift gates run at worktree/PR choke points.
- **IV. Grounding Freshness** — PASS (planned): `internal/sim/landing.go`,
  `internal/cognition/registry.go`, `internal/mind/convo.go` are wiki-note
  sources; `/grounding-wiki:wiki-update` re-pin is an explicit task (spec
  FR-006/SC-005 also *adds* the residual-gap documentation).
- **V. Model-Tiered Workflow** — PASS (planned): implementation delegated to
  `spec-implementer`. Tier: **Opus 4.8** — the slice is cross-package
  (`internal/cognition` + `internal/sim` + `internal/mind`), touches
  scheduling/governor-adjacent cognition doctrine, and changes
  doctrine-adjacent behavior (registry doctrine reinterpretation) — three
  independent rubric triggers.

**Post-Phase-1 re-check**: PASS — design adds one pure helper in the doctrine
package; no new projects, no new dependencies, no Complexity Tracking entries
needed.

## Project Structure

### Documentation (this feature)

```text
specs/067-staleness-budget-scaling/
├── CLAIM.md             # claim stub (spec 065)
├── spec.md              # feature spec (diagnosis + decision + FRs/SCs)
├── checklists/requirements.md
├── plan.md              # this file
├── research.md          # Phase 0: decisions R1–R5
├── data-model.md        # Phase 1: entities touched (no new state)
├── contracts/landing-gate.md  # Phase 1: gate rule + reason grammar
├── quickstart.md        # Phase 1: validation guide
└── tasks.md             # Phase 2 (/speckit-tasks)
```

### Source Code (repository root)

```text
internal/
├── cognition/
│   ├── registry.go      # DecisionClass doctrine comment (FR-005) +
│   │                    #   EffectiveBudgetTicks helper (R3)
│   └── registry_test.go # helper table tests (scaling, 1x identity, uncapped)
├── sim/
│   ├── landing.go       # rungStale consumes the helper + scaled reason (R4)
│   ├── landing_test.go  # updated signature + acceptance-scenario tables
│   └── staleness_replay_test.go  # new: speed.set mid-flight replay proof (R5)
├── mind/
│   └── convo.go         # scene-staleness pre-abort consumes the helper (R2#4)
└── clock/               # untouched (TicksPerSecond already exists)

docs/wiki/               # re-pin: cognition-horizon note gains the
                         # scheduling-vs-delivery split + residual gap (FR-006)
```

**Structure Decision**: existing package layout; the only new file is the
replay test. The helper lives in `internal/cognition` (R3) so both delivery
gates share one implementation.

## Design

### D1 — `EffectiveBudgetTicks` (internal/cognition/registry.go)

```
func (dc DecisionClass) EffectiveBudgetTicks(ticksPerSecond float64) int64
```

- `ticksPerSecond <= 0` → returns `dc.BudgetTicks` (uncapped guard, mirrors
  Route's posture; theoretical branch — Route suppresses everything at
  uncapped speed).
- Otherwise → `int64(float64(dc.BudgetTicks) * ticksPerSecond)`. Ladder values
  are exact small floats; the product is exact for every registry budget
  (deterministic across platforms).
- Doctrine comment on `DecisionClass` rewritten per FR-005: `BudgetTicks` is
  the staleness budget **at 1x** — wall-clock patience — enforced scaled at
  the delivery gates; values remain reviewed-code doctrine (decision-4
  posture preserved), never runtime tuning.

### D2 — Landing gate (internal/sim/landing.go)

`rungStale(class string, staleness int64)` grows a `ticksPerSecond float64`
parameter (or equivalently reads it at the one call site,
`landing.go:76`, from `l.state.Speed.TicksPerSecond()` — event-sourced,
FR-002). Rejection reason becomes the R4 grammar:
`staleness %d > budget %d (%d at 1x × %gx)`; at uncapped speed the reason
keeps the unscaled form. `internal/tui/decisions.go` maps outcome codes and
never parses reasons — no renderer change (SC-004).

### D3 — Convo pre-abort (internal/mind/convo.go:412)

The scene-staleness check compares against
`cc.meta.class.EffectiveBudgetTicks(replica speed's TicksPerSecond())` instead
of raw `BudgetTicks`, keeping the pre-check and the landing gate one
predicate (R2 #4). Mind-side only; no replay surface.

### D4 — Tests (proof shape, R5)

1. `registry_test.go`: helper table — 1x identity, 4x/8x/32x products,
   uncapped fallback, every registry class.
2. `landing_test.go`: spec US1 scenarios — 2000 ticks @8x lands, 1300 @1x
   rejects (regression), >9600 @8x rejects with scaled-reason grammar; existing
   1x assertions untouched (SC-003).
3. `staleness_replay_test.go`: record a run with a pending thought spanning a
   `speed.set`, replay the log, assert identical landing outcomes (SC-002,
   US2).
4. Full suite green: `go test ./...` (includes existing replay + governor
   suites).

### D5 — Evidence + grounding (post-merge tail)

- Measured-run half of SC-001: rejected-stale share on the TASK-122 measure
  world profile at 8x, recorded as board-task evidence (unit-test arm
  suffices if rerun is impractical — spec assumption).
- `/grounding-wiki:wiki-update` re-pins notes sourcing the touched files and
  adds the FR-006 residual-gap documentation; player-docs freshness check
  after.

## Complexity Tracking

No Constitution Check violations — table intentionally empty.
