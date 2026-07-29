# Tasks: Move-miracle target freshness (TASK-166)

**Input**: `specs/091-move-miracle-freshness/spec.md` (decision (a) ratified there,
with replay analysis — plan is folded into the spec; single-mechanism change).

## Phase 1: Door-side name re-resolution

- [X] T001 Implement FR-001/FR-002/FR-003 in the guardian move path
  (internal/guardian/turn.go move arg handling + the door validation/emission it
  feeds): class=villager + name ⇒ resolve live position, x/y advisory; refusal
  before charge on unknown/dead name; recorded event carries resolved coords;
  `applyEntityMoved` (internal/sim/miracles.go) untouched.
- [X] T002 FR-004: one-line name-preference gloss in the move guidance
  (internal/tool/derive.go) and in the raced-refusal message.

## Phase 2: Tests

- [ ] T003 Unit tests per FR-005: raced-move-lands, dead/unknown-name pre-charge
  refusal, coordinate-only unchanged, structure/pile unchanged.
- [ ] T004 Replay byte-identity regression over a log containing pre-fix
  entity_moved events; `go test -race ./...` green.

## Phase 3: Grounding + probe

- [ ] T005 Re-verify + re-pin wiki notes whose sources this branch touches
  (guardian / miracle / tool-registry notes listing turn.go, miracles.go,
  derive.go); regenerate docs/player pages if the probe demands; merge-drift pr
  gate green from the worktree.
- [ ] T006 Live probe on a seeded MEASURE world (never playtest-1) at 8x per the
  TASK-163 recipe: demonstrate a name-addressed raced move landing; write
  docs/design/evidence/task-166/results.md. (Orchestrator runs or supervises the
  probe; the implementer prepares the world/dials command list in the evidence
  file's header.)
