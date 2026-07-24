# Implementation Plan: Loud Build Failure & Occupancy Tolerance

**Branch**: `task-91-loud-build-failure` | **Date**: 2026-07-24 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/038-loud-build-failure/spec.md`

## Summary

Builds cancelled by mid-work re-validation currently resolve as a bare
`agent.intent_done` (executor.go:684-687) — indistinguishable from success —
and the wall guard (`executor.go:657`) insta-cancels on transient occupancy of
the reserved tile. Fix: (1) a new `agent.build_failed` event + situated
failure memory emitted by the executor when a build's site genuinely fails;
(2) split the wall guard so occupancy no longer cancels — during work only
site validity is checked, at the completion moment an occupied res tile defers
completion (never entomb) up to a fixed grace window, after which the build
fails loudly; (3) catalog + TUI rendering for the new event; (4) regression
tests including replay byte-identity.

## Technical Context

**Language/Version**: Go (module `promptworld`, toolchain per go.mod)

**Primary Dependencies**: stdlib only in `internal/sim`; Bubble Tea TUI in `internal/tui`

**Storage**: event-sourced world log (SQLite `world.db`); all state changes ride events through reducers (`internal/sim/state.go`)

**Testing**: `go test ./...`; sim tests are deterministic replay-style unit tests (`internal/sim/*_test.go`)

**Target Platform**: macOS/Linux CLI + TUI

**Project Type**: single Go project, event-sourced simulation

**Performance Goals**: no new per-tick allocations in the hot executor path; guard split is branch-only

**Constraints**: deterministic replay byte-identity (any new event/memory must flow through reducers identically on record and replay); never entomb an agent in a wall

**Scale/Scope**: 2 packages touched (`internal/sim`, `internal/tui`) + 1 wiki doc; ~6 files

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. Artifact-Grounded Action | PASS | Root-caused on TASK-91 (file:line evidence); design choice recorded on the task; this spec dir is the plan of record. |
| II. One Task, One PR | PASS | TASK-91 → one branch `task-91-loud-build-failure` in `.worktrees/task-91`, one PR; spec phases are internal breakdown. |
| III. Gates Over Assertions | PASS | Spec linked to board via `spec-bridge:link` before implementation; status moves only via `spec-bridge:sync` against artifacts. |
| IV. Grounding Freshness | PASS (with follow-up) | `docs/wiki/event-types.md` is edited as part of the deliverable; wiki notes sourcing `executor.go`/`state.go` re-pinned via `/grounding-wiki:wiki-update` after merge. |
| V. Model-Tiered Workflow | PASS | Planning/gating on Fable 5; implementation delegated to `spec-implementer` subagent; tier choice + rubric justification recorded on TASK-91. |

Post-Phase-1 re-check: PASS — design adds one event type, one guard split, one
constant; no new packages, no new abstractions. No Complexity Tracking entries
needed.

## Project Structure

### Documentation (this feature)

```text
specs/038-loud-build-failure/
├── spec.md
├── plan.md              # This file
├── research.md          # Phase 0: decisions (event name, grace design, salience)
├── data-model.md        # Phase 1: event payload, state transitions
├── quickstart.md        # Phase 1: validation guide
├── contracts/
│   └── agent-build-failed.md   # Phase 1: event contract (catalog row source)
└── tasks.md             # Phase 2 (/speckit-tasks)
```

### Source Code (repository root)

```text
internal/sim/
├── executor.go     # guard split (639-687), failure emission, deferred completion
├── agents.go       # BuildFailedPayload struct; wallOccupancyGraceTicks constant
├── state.go        # agent.build_failed reducer (clear intent, stamp IdleSince)
├── memory.go       # (read-only: situatedMemoryEvent, salience table)
├── wall_test.go    # regression tests (TestWallOccupancyGuard rewrite + new tests)
└── whole_feature_test.go  # replay byte-identity expected-event sets

internal/tui/
└── digest.go       # render agent.build_failed distinctly (alongside :271)

docs/wiki/
└── event-types.md  # catalog row for agent.build_failed; pass-through note
```

**Structure Decision**: existing single-project layout; all sim changes stay
inside `internal/sim` (event-sourced executor/reducer pattern), rendering in
`internal/tui/digest.go`, documentation in `docs/wiki/event-types.md`. No new
packages or files except tests ride existing files.

## Design (Phase 1 digest)

1. **New event `agent.build_failed`** — `BuildFailedPayload{Agent int, Goal string, Reason string}`
   (styled after `IntentRejectedPayload`, cognition.go:87). Emitted by the
   executor *instead of* bare `agent.intent_done` when a build goal's mid-work
   re-validation genuinely fails. Reducer effect identical to
   `agent.intent_done` (state.go:544-554): clear intent, stamp `IdleSince` —
   so planner re-arm behavior (mind.go:218 absorb) is preserved by subscribing
   `agent.build_failed` alongside the existing re-arm list.
2. **Guard split (walls, executor.go:651-657)**:
   - During work (`nextTick - WorkStart < workDuration`): validity = `buildSite(m, s, in.ResX, in.ResY)` only. Occupancy ignored.
   - At/after due tick: if site invalid → loud fail. If `agentAt(s, in.ResX, in.ResY)` → defer: return no event this tick; completion fires the first tick the tile is clear.
   - Grace bound: if `nextTick - in.WorkStart >= workDuration + wallOccupancyGraceTicks` and tile still occupied → loud fail (reason: site blocked too long). No new state needed — the bound derives from `WorkStart`, keeping replay trivially deterministic.
   - `wallOccupancyGraceTicks = 120` (constant in agents.go next to `buildWallTicks = 600`).
3. **Other build goals** (`build_fire/shelter/oven/chest/path`, executor.go:647-650):
   validity check unchanged (`buildSite(Target)`), but failure now emits
   `agent.build_failed` + failure memory instead of bare `intent_done`. The
   generic `!valid` exit at executor.go:684-687 stays as-is for non-build
   goals (forage/chop/hunt/…, explicitly out of scope).
4. **Failure memory** — emitted with the failure event:
   `situatedMemoryEvent(nextTick, i, salShelter, <place>, in.Reason, OriginAction, text)`
   (memory.go:189); text names the structure and the cause, e.g.
   "My stone wall was never built — the site was blocked too long." Salience
   reuses the wall-build tier (`salShelter`=6) so failure is as memorable as
   success.
5. **TUI** — digest.go renders `agent.build_failed` as
   `"<name>'s <goal> failed — <reason>"`, visibly distinct from "finished".
   Check the sim-loop event pass-through whitelist (event-types.md:206) and
   include `agent.build_failed` wherever `agent.intent_rejected` is included.
6. **Docs** — event-types.md: new row (`agent.build_failed | BuildFailedPayload{agent, goal, reason} | executor (build re-validation) | intent cleared`); amend the `agent.intent_done` row to note build failures no longer funnel through it.
7. **Tests** — see quickstart.md; `TestWallOccupancyGuard` (wall_test.go:300)
   is rewritten for the new semantics; new tests cover passerby-tolerance,
   squatter-grace-fail, site-vanished-fail (wall + one non-wall build), and
   `TestReplayByteIdentityWallsAxesPaths` expected sets gain the new events.

## Complexity Tracking

No constitution violations — table intentionally empty.
