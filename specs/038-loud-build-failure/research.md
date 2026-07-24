# Research: Loud Build Failure & Occupancy Tolerance

**Date**: 2026-07-24 · **Spec**: [spec.md](spec.md)

All unknowns were resolved by direct code exploration (read-only pass over
`internal/sim`, `internal/mind`, `internal/tui`, `docs/wiki`) before this plan;
no external research needed. File:line evidence also lives on TASK-91.

## D1 — Occupancy: tolerance, not pathing avoidance

- **Decision**: the build tolerates occupancy of the reserved tile; pathing is
  untouched.
- **Rationale**: pathfinding is unweighted deterministic BFS
  (`internal/sim/path.go`) with binary `passable()` (`terrain.go:38-51`) that
  consults no reservations; `ResX/ResY` live only on each agent's `Intent`
  (`agents.go:58-59`) with no queryable index. Soft-avoid would require a new
  tile-cost concept plus a reservation registry consulted inside `passable`,
  which every movement step and `buildSite` call — huge blast radius, and a
  hard block risks agents walling themselves out of routes. Tolerance is a
  branch-level change in one executor case.
- **Alternatives considered**: (a) reservation index + soft-avoid costs —
  rejected for blast radius and new concepts; (b) hard-impassable res tiles —
  rejected: turns social proximity into route denial and can strand agents.
- Recorded on TASK-91 (AC #3) before speccing.

## D2 — Grace bound derived from WorkStart, no new state

- **Decision**: deferred completion fails loudly when
  `nextTick - WorkStart >= workDuration + wallOccupancyGraceTicks`; between
  due-tick and that bound, an occupied res tile simply produces no event that
  tick and completion fires the first clear tick.
- **Rationale**: `Intent.WorkStart` (`agents.go:60`) is already event-sourced
  (`agent.work_started` reducer, `state.go:532-543`), so the bound is a pure
  function of existing replayed state — replay byte-identity holds with zero
  new fields. A per-tick occupancy counter would need new Intent state and a
  new event or reducer mutation to stay deterministic.
- **Consequence**: the bound is "time past due", not "continuous occupancy" —
  intermittent occupancy that keeps the tile busy at every due tick still hits
  the bound; any single clear tick completes the build. This satisfies the
  spec's "bounded, never forever, never silent" requirement with the simplest
  deterministic mechanism.
- **Constant**: `wallOccupancyGraceTicks = 120` — 20% of `buildWallTicks`
  (600, `agents.go:649`); long enough for a passerby/short chat, short enough
  to resolve within a fraction of the build itself. Tuning later is out of
  scope (spec assumption).

## D3 — Event name and payload

- **Decision**: `agent.build_failed`, payload
  `BuildFailedPayload{Agent int, Goal string, Reason string}`.
- **Rationale**: mirrors the existing failure-shaped payload
  (`IntentRejectedPayload{Agent, Goal, Reason, ...}`, `cognition.go:87`) so
  consumers/tooling see a familiar shape; named for what happened
  (build failed mid-work) as distinct from `agent.intent_rejected`
  (up-front landing rejection, `landing.go:47-65`). Grep confirms no existing
  `*failed`/`*cancelled` agent event to reuse.
- **Alternatives considered**: (a) overload `agent.intent_rejected` — rejected:
  its documented semantics are pre-acceptance, observability-only, "none"
  state effect, while this event must clear the intent; (b) add a status field
  to `agent.intent_done` — rejected: changes an existing payload consumed
  everywhere and would still render as "finished" in old tooling.

## D4 — Failure memory written by the executor, not the mind

- **Decision**: the executor emits the situated failure memory (via
  `situatedMemoryEvent`, `memory.go:189`, `OriginAction`) in the same tick as
  `agent.build_failed`.
- **Rationale**: the mind's `absorb` (`mind.go:218`) writes no memories from
  intent resolution — it only re-arms the planner; all action memories are
  emitted inline by the executor today (`executor.go:758,761,774,797,815`).
  Following that pattern keeps memory writing in one place and on the
  deterministic event stream.
- **Salience**: reuse the wall-build tier (`salShelter`=6, `memory.go:228-258`)
  so the failure competes equally with the success memory it falsifies.

## D5 — Scope of loud failure

- **Decision**: loud failure applies to the build goals in the executor's
  validity switch (`build_fire/shelter/oven/chest/path` site checks,
  `executor.go:647-650`; walls, `651-657`). The generic bare-`intent_done`
  exit (`executor.go:684-687`) remains for non-build goals; completion-time
  no-op re-checks (craft/cook/bathe/deposit, `executor.go:853-898,566-619`)
  stay as-is.
- **Rationale**: TASK-91's ACs target builds; the silent-`intent_done` pattern
  is systemic (~10 goals) but widening now multiplies replay-test churn and
  belief-model risk. Follow-up task to be filed for non-build goals.

## D6 — Mind re-arm and pass-through

- **Decision**: `agent.build_failed` joins the planner re-arm list in the
  mind's `absorb` (`mind.go:218`, alongside `agent.intent_done`) and any
  sim-loop pass-through whitelist that carries `agent.intent_rejected`
  (see `docs/wiki/event-types.md:206`).
- **Rationale**: a failed builder must re-plan exactly like a finished one
  (FR-008), and observers must actually receive the event (FR-001/SC-005).
