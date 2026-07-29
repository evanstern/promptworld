# Tasks: Determinism scope note + reducer-constants doctrine (TASK-75)

**Input**: `specs/092-determinism-doctrine/spec.md` (docs/doctrine task — plan folded
into the spec; minimal code = comments only).

## Phase 1: Audit

- [ ] T001 Re-verify the card's anchors and sweep internal/sim's Apply/reducer
  paths for every site re-deriving outcomes from mutable gameplay constants;
  produce the audit list (file:line + constants) — FR-004.

## Phase 2: Doctrine + docs

- [ ] T002 Per-log-not-per-seed determinism limit in deterministic-rng.md, the
  EffectiveRate-owning note, and README's determinism paragraph; correct every
  per-seed claim — FR-001.
- [ ] T003 Reducer-constants doctrine (emitter-computes default; re-derive
  exception ⇒ format_version + migration; TASK-134 pointer; spec-019 precedent;
  spec-048 genesis-pin residual scope) + the audit list in the owning wiki note —
  FR-002/FR-004.
- [ ] T004 Hazard comments at each audited site (comment-only; go test ./... green,
  gofmt clean) — FR-003.

## Phase 3: Grounding + gates

- [ ] T005 Re-pin all touched wiki notes; regenerate docs/player pages the probe
  flags (README is a pinned input); merge-drift pr gate exits 0 from the
  worktree — FR-005.
