# Tasks: Target-addressing grammar for bundle effects

**Input**: Design documents from `/specs/082-target-addressing/`
**Prerequisites**: spec.md, data-model.md (normative grammar), plan.md

**Tests**: fixtures ARE deliverables (FR-010); byte-identity + replay +
error-taxonomy tests land with the code they prove, same commits.

**Tier**: Opus 4.8 via `spec-implementer` (recorded on TASK-97 — cross-package,
two-consumer grammar). Planning/gating stays on Fable 5.

## Phase 1: Setup

- [X] T001 Worktree already cut (`.worktrees/task-97`, branch
  `task-97-target-addressing`; claim landed per spec 065/TASK-160 flow).
  Confirm baseline: `go test ./internal/bundle/ ./internal/sim/` green in
  the worktree; branch is fresh vs `origin/main` (merge main INTO the
  branch if stale — never rebase).

## Phase 2: Foundational — the grammar package (blocks all user stories)

- [X] T002 `internal/target/target.go` (new): `Parse` per data-model.md
  §1–2 — reserved-prefix rule, four classes, name/point/rect/line forms,
  normalization (rect min/max corners; line endpoint-order preserved,
  axis-aligned enforced; whitespace rules), non-negative-integer
  coordinates; typed errors exposing taxonomy classes
  `syntax`/`class`/`form` (plan D1, FR-001)
- [X] T003 `Address.Tiles()` in the same file: point single-tile; rect
  row-major (y then x, ascending); line endpoint-order stepping — pure
  function, no state (data-model.md §2; US3 AC1/AC2)
- [X] T004 `internal/target/target_test.go`: table tests over every
  data-model.md §1 example row + normalization + enumeration order +
  error-class assertions; the stdlib-only import assertion (leaf-safety
  pin — SC-004, the TASK-157 seam's structural guarantee) (plan D1)

## Phase 3: US1 — structures & piles by class+tile; remove_entity becomes real (P1) 🎯 MVP

- [X] T005 [US1] `internal/sim`: exported deterministic read-only presence
  probes for structure/pile-on-tile + map dims for bounds (over today's
  unexported `structureIndexAt`/`pileAt`); doc comments per the
  `VillagerAt` one-helper discipline; NO reducer/payload/state changes
  (plan D4; spec Assumptions)
- [X] T006 [US1] `internal/bundle/effects.go`: `compileOne` dispatch
  through `target.Parse` per the data-model.md §4 matrix — `move_entity`
  villager(name/point via `VillagerAt`)/structure@/pile@ →
  `EntityMovedPayload{Class, X, Y, ToX, ToY, Gratis:false}`;
  `remove_entity` structure@/pile@/terrain@ → `EntityRemovedPayload`;
  `grant_item` villager forms only; bounds checks for tile forms; ❌ cells
  → `form` errors (villager-removal doctrine mirrored; rect/line messages
  name the TASK-157 reservation); bare-name path byte-identical to v1
  (plan D2/D3, FR-002/003/004/005/006/007 — script mode rides the same
  path, no script.go structural change)
- [X] T007 [US1] Error-taxonomy table test in
  `internal/bundle/effects_test.go`: every §5 class × representative
  effect kind — message names effect index, field, offending address;
  whole-invocation rejection, no charge (SC-005; US1 AC3/AC5)
- [X] T008 [US1] Fixture world `internal/bundle/testdata/worlds/<addressing>/`
  (structure + chest-with-contents + pile + tree) with declarative tools
  covering literal and `{args}`-templated class+tile targets, plus a
  scripted tool composing `class@x,y` strings (plan D6, FR-010; US1
  AC1/AC2/AC4)
- [X] T009 [US1] Byte-identity test: compiled structure/pile move + remove
  payloads equal `guardian.BuildMiracleBatch`'s for the same inputs
  (dogfood-move precedent extended) — SC-002
- [X] T010 [US1] Replay determinism: new fixtures ride the existing
  `replay_test.go`/`script_replay_test.go` byte-identity pattern including
  delete-bundle-dir-before-replay; chest removal spills contents per the
  unchanged reducer (SC-001/SC-003)

## Phase 4: US2 — terrain removal (P2)

- [X] T011 [US2] `terrain@X,Y` remove path: compile-time bounds-only
  (removability stays reducer-side), fixture tree-removal case lands and
  clears the tile; `move_entity` terrain → `form` error; dry-run rejection
  of a non-removable tile proven whole-invocation (US2 AC1/AC2/AC3;
  extends T006/T008 — same files, kept a separate slice so US1 is an
  independently green MVP)

## Phase 5: US3 — the designation seam, contract-named (P3)

- [X] T012 [US3] Bundle-side reserved-form proof: rect/line targets fail
  compile with the reservation error naming TASK-157/designations
  (effects_test case; US3 AC3 — parser work already landed in Phase 2)
- [X] T013 [US3] Amend
  `specs/036-scriptable-agent-tools/contracts/bundle-manifest.md`:
  "Target addressing" section (grammar summary → 082 data-model.md as
  normative; §4 matrix; error behavior; reserved-prefix/compat rule;
  `text` param-kind guidance) + named **"Designation addressing (TASK-157
  seam)"** section (parser package + leaf-safety, rect/line normalization
  + enumeration order, bundle-reserved status) + compatibility note
  (additive value-space extension) (plan D5, FR-008)
- [X] T014 [US3] Update `docs/bundles.md` authoring guide: address forms
  with examples per effect kind, error messages authors will see, the
  designation reservation (plan D5, FR-008)

## Phase 6: Grounding + gates (in-branch, wiki-in-PR lifecycle)

- [X] T015 Wiki re-pin ON THIS BRANCH: update `docs/wiki/bundle-tools.md`
  (grammar paragraph replaces the TASK-97-limitation operational note;
  sources gain `internal/target/target.go`), re-verify + re-pin to a
  branch commit; re-pin every other note the pr gate names (candidates:
  notes pinning the `internal/sim` file T005 touches, e.g.
  `sim-state-world-fields.md`/`sim-state-reducer.md`;
  `tool-registry-guardian-tools.md` only if `internal/tool` sources were
  touched — expected NOT). The gate is the authority — produce what it
  names (constitution IV; plan D7)
- [X] T016 Player docs: `docs/wiki/` changed ⇒ regenerate `docs/player/`
  via the player-docs project skill, then run the probe DIRECTLY:
  `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`
  → exit 0 (spec 069; no bypass)
- [X] T017 Conditional catalog gate: NO new event type is expected
  (FR-009); IF implementation introduced one anyway, add its
  `digestRegistry` entry + `catalogFixture` row + `docs/wiki/event-types.md`
  row so `TestCatalogSweep` (`internal/tui/digest_test.go`) passes — never
  merge with the sweep red or the type undocumented
- [ ] T018 Full gates from `.worktrees/task-97`: `gofmt -l` clean;
  `go test ./...` green; `node scripts/check-merge-drift.mjs pr` exits 0
  (includes wiki-repin-missing / player-docs-stale findings). Then ONE PR
  for the whole task; merge with `gh pr merge --merge` ONLY
  (merge-commit-only — in-branch pins are branch hashes; squash stales
  them). SC-006

## Phase 7: Post-merge bookkeeping (derived state, lands by merge — TASK-160)

- [ ] T019 After the PR merges: spec-bridge sync (TASK-97 → Done as
  artifacts prove), tick this tasks.md, runbook execution-log row
  (`docs/design/faith-directives-sweep-runbook.md`), unblock note on
  TASK-157 (its dependency is met; point it at the contract seam section).
  ALL of it authored on a branch in a worktree and landed on main by merge
  (PR or manual `git merge --no-ff` at root) — NOTHING commits directly at
  root, bookkeeping included (TASK-160 iron-clad rule). Then worktree +
  branch cleanup, ff-pull root.

## Dependencies & execution order

- T001 → Phase 2 (T002 → T003 → T004) → Phase 3.
- Within Phase 3: T005 [P] and (T002-dependent) T006 can start in
  parallel; T006 depends on T005; T007–T010 depend on T006 (T008 [P] with
  T007; T009/T010 after T008).
- T011 after T006/T008. T012 after T006. T013–T014 [P] after T011/T012
  (contract documents the final matrix).
- Phase 6 after all code+docs: T015 → T016; T017 anytime post-code;
  T018 last, gates everything. T019 post-merge only.
- **MVP** = Phases 2–3 (US1: the board's headline — remove_entity real,
  structures/piles addressable). US2/US3 ride the same single PR
  (one task, one PR — phases are internal breakdown, never PR boundaries).
