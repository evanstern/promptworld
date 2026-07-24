# Quickstart: Validating Per-Agent Mental Maps

**Date**: 2026-07-24 · **Spec**: [spec.md](spec.md) · Success criteria SC-001..SC-008.

## Prerequisites

- Go 1.26.4; repo checkout of branch `task-96-agent-mental-maps` (`.worktrees/task-96`).
- No external services. All validation is `go test` + the built binary.

## Fast loop (unit + package)

```bash
go test ./internal/sim/ ./internal/mind/ ./internal/tool/ ./internal/tui/
go test -race ./...
```

Must-pass suites mapped to criteria:

| Criterion | Proof |
|---|---|
| SC-001 knowledge-gated resolution | new `internal/sim` tests: resolver returns known-far target over unknown-near (spec US1 scenario 3); "you know of no" error when map lacks kind |
| SC-002 divergent prompts, no 6-cap | `internal/mind/prompt_test.go` additions per contracts §3 |
| SC-003 replay determinism | existing + new twins: `TestReplayRebuildsState`, `TestDeterminismSameSeedSameTimeline`, mental-map replay-byte-identical test (pattern: `internal/mind/replay_test.go`), snapshot byte-stability twin (`TestMapOmitemptyStable` — old snapshot without `map` field round-trips byte-identically) |
| SC-004 search terminates/expands | `internal/sim` tests: frontier target is explored∧adjacent-to-unexplored; repeated search monotonically grows explored bits; fully-explored map → exhaustion error |
| SC-005 stale correction | test: structure removed while agent away → arrival emits `agent.map_corrected`, fact gone, memory stamped |
| SC-006 talk transfer | test: A knows fact, B doesn't → social beat emits `social.place_told` ≤2 facts, B's map gains told-provenance fact with A's Seen tick |
| SC-007 viability | `TestVillageSurvivesTwoDays` (must stay green with gating + seeding); migration test: v3 world loads, agents hold seeded knowledge |
| SC-008 3D path documented | data-model.md § Layered-3D extension path exists (review item, not a test) |

Event/catalog gates (red until wired — run early and often):

```bash
go test ./internal/tui/ -run TestCatalogSweep
go test ./internal/sim/ -run TestValidateToolCoverage   # or the boot-time check's test twin
```

## Full-binary determinism (e2e)

```bash
go test ./e2e/ -run TestDeterminism_FullBinary   # two seed-777 worlds to tick 25000, event histories line-identical
```

## Migration check

```bash
go test ./internal/sim/ -run Migrate
# manual: point the binary at a FormatVersion-3 world save →
#   promptworld migrate <world-dir> && promptworld daemon <world-dir>
# expect: agents hold explored home areas + witnessed facts for existing structures; no starvation wave
```

## Live smoke (manual, optional)

```bash
promptworld new /tmp/mmtest --seed 777 && promptworld daemon /tmp/mmtest &
promptworld tui /tmp/mmtest
```

Watch for: villagers' first prompts show only spawn-area knowledge; a villager choosing
`search`; digest lines for discoveries; a talk followed by a told-provenance action; the
chronicle narrating a correction ("the fire was cold when she arrived").

## Post-implementation gates (constitution)

```bash
# wiki re-pin (sources touched: reflex-policy, agent-mind, cognition, tool-registry,
# event-types, sim-state-reducer, snapshots, social-fabric, tool-loop, worldmap-generation)
/grounding-wiki:wiki-update
node .claude/skills/player-docs/scripts/check-freshness.mjs --check   # then player-docs if stale
# board sync
spec-bridge:sync
```
