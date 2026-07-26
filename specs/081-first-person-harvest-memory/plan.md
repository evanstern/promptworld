# Implementation Plan: First-person harvest memory (mental-map update at chop/quarry time)

**Branch**: `task-159-first-person-harvest-memory` | **Date**: 2026-07-26 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/081-first-person-harvest-memory/spec.md`

## Summary

Completing a chop/quarry today mutates world overlays (`s.Cleared`/`s.Quarried`)
but no mental map, so the spec 041 US3 perception sweep later "corrects" the
actor's and every on-scene witness's map — minting third-party-voiced loss
memories for acts they performed or watched (75% of all memories on world
worldy). Fix: the `agent.chopped` / `agent.quarried` reducer arms also remove
the matching place-fact from the actor's map and from the map of every awake
living villager within `witnessRadius` (derived from the same pre-mutation
state the emitter checked — the established spear-hunt/axe-yield reducer
idiom); the executor mints a first-person act memory for the actor (the
`salHunt` precedent) as a companion `agent.memory_added`; the mind's absorb
trigger extends to chop/quarry events so an on-scene witness whose current
intent targets the felled tile still re-arms. `agent.map_corrected` is
untouched and keeps firing for genuinely absent agents.

## Technical Context

**Language/Version**: Go (module `promptworld`, toolchain per `go.mod`)

**Primary Dependencies**: stdlib + internal packages only for this slice —
`internal/sim` (executor, state reducer, mental map, memory), `internal/mind`
(absorb triggers), `internal/store` (event log), `internal/worldmap` (terrain)

**Storage**: event-sourced SQLite log (`world.db`, events + snapshots);
state is a pure fold of events — no schema change (no new event types)

**Testing**: `go test ./...`; table-driven unit tests beside code in
`internal/sim` (`mentalmap_test.go`, `memory_*_test.go` precedents) and the
existing replay/fork determinism harnesses (`fork_event_test.go`,
`governor_replay_test.go` shapes)

**Target Platform**: daemon on macOS/Linux (unchanged)

**Project Type**: single Go project, event-sourced simulation daemon + TUI

**Performance Goals**: reducer-arm work is O(agents × facts-at-tile) per
chop/quarry event — negligible against the existing per-tick sweep; no new
per-tick cost

**Constraints**: replay determinism (state mutation only in reducers, pure
function of event + prior state, same code version); memories accrete only
via `agent.memory_added` (TestMemoriesAccrete); no new event types
(contracts/events.md posture for state-derived behavior); witness radius =
the perception sweep's `witnessRadius` (one perceptual reality)

**Scale/Scope**: 2 reducer arms, 1 executor emit site per act (2 acts), 1
absorb-trigger extension, ~4 wiki notes re-pinned + player docs regenerated

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Artifact-Grounded Action** — PASS: diagnosis pinned to file:line +
  worldy event-log counts on TASK-159; spec/plan/tasks under
  `specs/081-first-person-harvest-memory/`; operator decision (first-person
  memory, 2026-07-26) recorded in spec Assumptions.
- **II. One Task, One PR** — PASS: TASK-159 ↔ branch
  `task-159-first-person-harvest-memory` in `.worktrees/task-159` ↔ one PR;
  root stays on main.
- **III. Gates Over Assertions** — PASS: spec-bridge linked (marker on
  TASK-159); claim gate passed for 081; worktree gate passed
  (`--spec 081 --task TASK-159`); pr gate will enforce spec 069 grounding.
- **IV. Grounding Freshness** — PASS (planned): wiki notes whose sources this
  touches (`mental-map-perception`, `mental-maps`, `event-types-mental-map`,
  `sim-state-reducer`-family, memory notes) re-verified and re-pinned ON the
  task branch; `docs/player/` regenerated in-branch; merge-commit-only.
- **V. Model-Tiered Workflow** — PASS: planning artifacts on Fable 5;
  implementation delegated to `spec-implementer` pinned **Opus 4.8** —
  rubric: doctrine-adjacent behavior change (memory formation / perception
  doctrine) AND reducer/replay-contract surface (cross-cutting
  executor+reducer+mind). Tier choice + justification recorded on TASK-159.

No violations → Complexity Tracking not required.

*Post-Phase-1 re-check (2026-07-26): design added no projects, no new event
types, no new packages — gates still pass.*

## Project Structure

### Documentation (this feature)

```text
specs/081-first-person-harvest-memory/
├── CLAIM.md             # spec 065 claim stub (TASK-159 attribution)
├── spec.md              # feature specification
├── plan.md              # this file
├── research.md          # Phase 0: design decisions + rationale
├── data-model.md        # Phase 1: entities & state transitions
├── quickstart.md        # Phase 1: end-to-end validation guide
├── contracts/
│   └── events.md        # Phase 1: event/reducer contract deltas
├── checklists/
│   └── requirements.md  # spec quality checklist (done)
└── tasks.md             # Phase 2 (/speckit-tasks — not this command)
```

### Source Code (repository root)

```text
internal/sim/
├── executor.go          # chop/quarry emit sites: + actor act-memory companion
│                        # (situatedMemoryEvent, salHunt shape); perception
│                        # sweep untouched
├── state.go             # agent.chopped / agent.quarried reducer arms: + map-
│                        # fact removal for actor & in-radius awake witnesses
├── mentalmap.go         # fact-removal primitive reused (map_corrected arm's)
├── memory.go            # + salChop/salQuarry consts + first-person texts
├── mentalmap_test.go    # + act-time removal, same-tick sweep, asleep/absent
├── memory_test.go       # + act-memory accretion tests (TestMemoriesAccrete
│                        # posture preserved)
└── (replay determinism covered by existing fold/replay test harness + new case)

internal/mind/           # absorb trigger: chop/quarry events match witness
│                        # intent targets (map_corrected parity) + test
docs/wiki/               # re-pins: mental-map-perception, mental-maps,
│                        # event-types-mental-map, memory/salience notes
docs/player/             # regenerated (player-docs skill) in-branch
docs/design/tui/         # NOT touched (no internal/tui changes — gate no-op)
```

**Structure Decision**: single-project layout; all behavior lands in
`internal/sim` + a bounded absorb extension in `internal/mind`. No new
packages, files stay beside their concerns per existing convention.

## Design decisions (Phase 0 summary — full rationale in research.md)

1. **Removal lives in the reducer arms, not new events.** The
  `agent.chopped`/`agent.quarried` arms derive actor + in-radius awake
  witnesses from the same pre-mutation state the emitter checked (established
  idiom: axe/spear yield re-derivation). Pure function of event + prior state
  → replay-deterministic; no event-log schema change.
2. **Actor memory is an executor companion event** — `situatedMemoryEvent`
  with new `salChop`/`salQuarry` in the `salHunt` band (4), origin action,
  first-person text ("Felled the tree at (x,y)." / "Quarried the outcrop at
  (x,y)."), riding the same batch as the act (buildFailedEvents shape).
3. **Absorb parity via event matching, not synthetic corrections.** The mind
  driver already treats `agent.map_corrected` as an absorb trigger keyed on
  the agent's intent target; extend the trigger set so `agent.chopped`/
  `agent.quarried` at coordinates matching a witness's current intent target
  re-arm identically. No `agent.map_corrected` is ever emitted for on-scene
  parties.
4. **Same-tick sweep needs no code change, only a test.** The sweep computes
  from pre-batch state (tree still present at the act tick → no correction);
  by the next sweep the reducer has removed the fact. A regression test pins
  this ordering.
5. **Asleep/dead within radius are excluded** from removal (perception parity
  with the sweep's awake-living filter); they keep the fact and the
  return-discovery narrative.

## Complexity Tracking

No Constitution Check violations — table intentionally empty.
