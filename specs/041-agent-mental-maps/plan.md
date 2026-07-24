# Implementation Plan: Per-Agent Mental Maps

**Branch**: `task-96-agent-mental-maps` | **Date**: 2026-07-24 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/041-agent-mental-maps/spec.md`

## Summary

Replace omniscient nearest-target resolution with per-agent private spatial knowledge. Each
villager gains a reducer-owned `MentalMap` (explored-tile bitmap + known place-facts with
provenance and last-seen ticks). Goal resolvers and the survival reflex resolve targets
against the acting agent's map only; the planner prompt renders known places instead of the
global structure list; a `search` verb walks toward the nearest frontier of unexplored space;
perception diffs, talk transfers, and divine reveals mutate the map exclusively through new
recorded events (plus silent derived explored-bit bookkeeping in the reducer), preserving
bit-identical replay. Technical approach grounded in `research/Agent-Mental-Maps/`
(occupancy-grid semantics with explicit unknown; fog-of-war three-state convention;
frontier exploration; provenance-conditioned merge) — see [research.md](research.md) for the
decision log.

## Technical Context

**Language/Version**: Go 1.26.4 (module `github.com/evanstern/promptworld`)

**Primary Dependencies**: stdlib; modernc SQLite (existing event store); no new dependencies

**Storage**: existing event-sourced SQLite store (`events` + `snapshots`); `MentalMap` rides
`Agent` inside `State` JSON snapshots (pointer + `omitempty` for byte-stable old snapshots)

**Testing**: `go test ./...` (+ `-race`); replay-byte-identical harnesses
(`internal/sim/*_replay_test.go`, `internal/mind/replay_test.go`); e2e full-binary
determinism (`e2e/determinism_e2e_test.go`); TUI catalog sweep (`internal/tui/digest_test.go`)

**Target Platform**: darwin/linux daemon (existing `promptworld` binary)

**Project Type**: single Go module; feature spans `internal/sim` (state, policy, executor,
memory), `internal/mind` (prompt, handlers), `internal/tool` (roster), `internal/tui`
(digest), `internal/world` + `internal/sim/migrate.go` (format migration)

**Performance Goals**: no regression to tick throughput at `speed max` (perception sweeps are
O(radius²) per mover per beat on a 64x64 map; explored-bit updates are bitwise)

**Constraints**: bit-identical replay (single `State.Apply` mutation path; events pure
function of (state, nextTick); no wall-clock/live RNG); snapshot byte-stability for
pre-feature worlds; model isolation (map writes only through recorded events; mind reads a
plan-time snapshot)

**Scale/Scope**: 64x64 default grid (W/H from manifest), 8 villagers; per-agent map ≈ 512 B
bitmap + O(dozens) place-facts — negligible; 5 new event types; ~1 new tool verb

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.* (v1.1.0)

| Principle | Status | Evidence |
|---|---|---|
| I. Artifact-Grounded Action | PASS | Spec/plan/tasks under `specs/041-*` on main; research corpus committed (`research/Agent-Mental-Maps/`, c70c53f); board TASK-96 carries decisions |
| II. One Task, One PR | PASS | Single branch `task-96-agent-mental-maps` in `.worktrees/task-96`; all phases land as commits there; one PR |
| III. Gates Over Assertions | PASS | spec-bridge link + sync after tasks phase; `TestCatalogSweep` enforces event/doc coverage; `ValidateToolCoverage` enforces tool↔resolver coverage; replay tests gate determinism claims |
| IV. Grounding Freshness | PASS (planned) | Implementation touches sources of wiki notes (reflex-policy, agent-mind, cognition, tool-registry, event-types, sim-state-reducer, snapshots, social-fabric, prompt/tool-loop); `wiki-update` + player-docs freshness check are explicit tasks before Done |
| V. Model-Tiered Workflow | PASS | Planning (this doc) on Fable 5; implementation delegated to `spec-implementer` at **Opus 4.8** — rubric: cross-package (sim+mind+tool+tui+world), doctrine-adjacent (changes what agents can know), determinism/concurrency-sensitive. Tier recorded on TASK-96 |

Post-Phase-1 re-check: PASS — design introduces no new projects, no new dependencies, no
violation rows; Complexity Tracking stays empty.

## Project Structure

### Documentation (this feature)

```text
specs/041-agent-mental-maps/
├── spec.md              # Feature spec (clarified 2026-07-24)
├── plan.md              # This file
├── research.md          # Phase 0: decision log (representation, write paths, gating, migration)
├── data-model.md        # Phase 1: MentalMap/PlaceFact, event payloads, transitions, 3D extension path
├── quickstart.md        # Phase 1: validation scenarios mapped to SC-001..SC-008
├── contracts/
│   └── knowledge-events.md  # Phase 1: new event types, digest lines, prompt-section contract
└── tasks.md             # Phase 2 (/speckit-tasks)
```

### Source Code (repository root)

```text
internal/sim/
├── agents.go            # Agent.Map *MentalMap field; constants (perception radius reuse, decay half-lives)
├── mentalmap.go         # NEW: MentalMap, PlaceFact, explored bitmap codec, known()/decay helpers
├── mentalmap_test.go    # NEW: unit + snapshot byte-stability twin tests
├── state.go             # reducer arms: agent.saw / agent.map_corrected / social.place_told / metatron.place_revealed; derived explored-bit bookkeeping on movement arms
├── policy.go            # goalResolvers consult acting agent's map; decideIntent reflex parity; search resolver
├── path.go              # nearestKnown / frontier BFS helpers (deterministic order preserved)
├── executor.go          # perception sweep (diff map vs ground truth → events); talk transfer hook beside TellableFor
├── memory.go            # knowledge-event situated memories (discover/correct/told/revealed)
├── migrate.go           # v3→v4 transform: seed knowledge for existing worlds
└── toolcheck.go         # coverage: search tool ⊆ resolvers

internal/world/world.go  # FormatVersion 3 → 4
internal/mind/
├── prompt.go            # known-places section replaces Village line (retires first-6 cap); unexplored summary
└── handlers.go          # search verb door (existing goal-door pattern)
internal/tool/…          # LoopRosterVillager: add search tool
internal/tui/digest.go   # digest registry entries + fixtures for new event types
docs/wiki/event-types.md # backticked rows for new events (test-enforced)
cmd/promptworld/…        # (only if genesis seeding needs a tick-0 event; default: NewState genesis)
```

**Structure Decision**: single-module feature following the existing reducer/executor/mind
split; all new sim state in `internal/sim/mentalmap.go` beside its peers (journal.go,
plan.go precedents); no new packages.

## Complexity Tracking

*No constitution violations — table intentionally empty.*
