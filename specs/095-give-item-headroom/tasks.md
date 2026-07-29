# Tasks: Carry-cap headroom guidance for give_item (TASK-167)

**Input**: `specs/095-give-item-headroom/spec.md` (decision (a)+gloss ratified there).

## Phase 1: Digest + gloss

- [ ] T001 Add live carry headroom (free units + cap) to the miracle-capable
  digest's per-villager line (spec 059 assembly path), same snapshot as
  positions; dead villagers unchanged (FR-001).
- [ ] T002 One-line headroom reference in the give_item gloss
  (internal/tool/derive.go), noting reject-whole (FR-002).

## Phase 2: Tests

- [ ] T003 Digest arithmetic test (partially-filled inventory ⇒ correct
  free/cap), gloss presence test, door regression test (message + reject-whole
  behavior byte-unchanged) (FR-003/FR-004); go test -race ./... green.

## Phase 3: Grounding + gates

- [ ] T004 Re-verify + re-pin touched wiki notes; player-docs probe; merge-drift
  pr gate exit 0.
