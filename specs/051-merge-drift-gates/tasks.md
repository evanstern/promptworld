# Tasks: Merge-Drift Gates

**Input**: Design documents from `/specs/051-merge-drift-gates/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: No automated test tasks — per plan.md/R11, validation runs the quickstart.md
fixture-repo scenarios; each story phase ends with its validation task.

**Organization**: Nearly everything lands in one file
(`scripts/check-merge-drift.mjs`), so tasks are largely sequential; story phases remain
independently testable increments. Contracts are normative: gate-cli.md (CLI surface),
detection-rules.md (git plumbing, cited as §N below), report-schema.md (output).

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

**Purpose**: script skeleton with the full CLI surface, so every later task slots into
a stable frame

- [x] T001 Create `scripts/check-merge-drift.mjs` skeleton per contracts/gate-cli.md: shebang + header comment (contract pointer, Node ≥ 18, zero-dep rule, mutation whitelist), mode/flag parsing with usage errors → exit 2 (unknown mode/flag, `--no-fetch` outside session, `--spec` outside worktree, pr mode on main/root), environment preflight (in a git repo, git ≥ 2.38 via `git version`), Finding/GateRun structures with verdict computation (max severity; pass/warnings → exit 0, blocked → exit 1) and severity table from data-model.md, human-line and `--json` emitters matching contracts/report-schema.md field names and stable ordering (findings by severity rank → rule → first evidence)

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: shared git plumbing every mode consumes

**⚠️ CRITICAL**: complete before any story phase

- [x] T002 In `scripts/check-merge-drift.mjs`, add git plumbing helpers per detection-rules.md: `execFileSync` git wrapper, fetch-with-outcome (success / failure signal for FR-014 handling by mode), merge-tree wrapper (§2: `--write-tree --name-only`, parse exit 0 → tree OID, exit 1 → OID + conflicted paths, other → env error), changed-file-set + baseLag computation (§3), RootState reader (§8: onMain, behind/ahead, dirty)
- [x] T003 In `scripts/check-merge-drift.mjs`, add live-branch enumeration (§1: worktree branches via `git worktree list --porcelain` ∪ local branches not contained in origin/main, `main` excluded), branch→task attribution from `task-<N>-<slug>` (§7), and finding fingerprints (§7: first 12 hex of SHA-256 of `gate|rule|branch|sorted evidence`)
- [x] T004 In `scripts/check-merge-drift.mjs`, add the wiki frontmatter parser (tolerant `sources:`/`verified_against:` extraction from `docs/wiki/*.md`, malformed note → info finding, §3/§6) and the semantic-overlap primitives (§3): backlog same-path overlap, wiki-pinned-source overlap (naming affected notes), `internal/tui/` prefix match, spec-number extraction + taken-number lookup via `git ls-tree origin/main specs/` (§5)

**Checkpoint**: all detection primitives callable; modes are wiring + severities

## Phase 3: User Story 1 — PR gate: no doomed PR gets opened (Priority: P1) 🎯 MVP

**Goal**: `pr` mode blocks predicted textual conflicts and warns on clean-merging
semantic overlaps, from inside a task worktree

**Independent Test**: quickstart.md scenario 1 — conflicting fixture branch → exit 1
naming the file; clean branch → exit 0

- [x] T005 [US1] In `scripts/check-merge-drift.mjs`, implement the `pr` mode driver per gate-cli.md: resolve target branch (`--branch` or HEAD; refuse `main`), fail-closed fetch (failure → exit 2 with `remote-unverified` message, FR-014), merge-tree vs fetched origin/main → `textual-conflict` block finding per conflicted file (FR-002)
- [x] T006 [US1] In `scripts/check-merge-drift.mjs`, wire pr-mode warn findings at the data-model.md severities: `stale-base` (baseLag > 0, FR-003), `backlog-overlap`, `wiki-sources-overlap`, `tui-surface` (message points at `node scripts/check-tui-design.mjs --changed`), `spec-number-collision` (branch-added `specs/NNN-*` vs origin, warn), `root-not-main` block, `dirty-worktree` info (FR-004)
- [x] T007 [US1] Validate US1 against quickstart.md scenario 1 (fixture repo): blocked run names `internal/tui/view.txt` and carries the stale-base/wiki/tui warnings; clean-branch control run exits 0; record actual output in the task notes

**Checkpoint**: MVP — the highest-leverage gate works end to end

## Phase 4: User Story 2 — Session-start janitor (Priority: P2)

**Goal**: `session` mode fast-forwards root, identifies cleanup-eligible worktrees
(incl. squash-merged), computes the n-way drift matrix, and flags stale grounding

**Independent Test**: quickstart.md scenario 2 — squash-merged fixture worktree reported
`cleanupReason: "empty-contribution"`; dirty worktree never eligible; matrix flags the
conflicting pair

- [x] T008 [US2] In `scripts/check-merge-drift.mjs`, implement the `session` mode driver: fetch-or-degrade (`unverifiedAgainstRemote: true`, verdict capped at warnings, `remote-unverified` warn finding; `--no-fetch` takes the same path), root handling (`root-not-main` block; guarded ff-pull per §8 when behind ∧ not ahead ∧ clean, `fastForwarded` reported)
- [x] T009 [US2] In `scripts/check-merge-drift.mjs`, implement cleanup eligibility (§4: ancestor OR empty-contribution via merged-tree-OID == `origin/main^{tree}`, AND clean worktree status) producing `cleanup-eligible` warn findings + `cleanupPrescriptions` (exact `git worktree remove` / `git branch -d` commands), and `--apply-cleanup` applying exactly those prescriptions and nothing else (FR-006, FR-009)
- [x] T010 [US2] In `scripts/check-merge-drift.mjs`, implement the drift matrix (FR-007): each live branch vs origin/main plus every unordered pair, lexicographic order, `pairwise-conflict`/`textual-conflict` warn findings with merge-order-visible messages; plus per-branch session-severity overlap findings (`backlog-overlap`, `spec-number-collision` warn; `stale-base` info)
- [x] T011 [US2] In `scripts/check-merge-drift.mjs`, implement grounding surfaces (FR-008, §6): wiki internally (per-note `git diff --name-only <verified_against> origin/main -- <sources…>`, unresolvable pin → info), player-docs and tui-design by delegation (`--json`/exit-code of their checkers, absent → `checker: "absent"` info), producing `grounding-stale` warn findings naming surface + touched sources
- [x] T012 [US2] Validate US2 against quickstart.md scenario 2: squash-cleanup detection, `--apply-cleanup` removes only task-1, dirty task-2 excluded, matrix + wiki-stale flags present; record actual output in the task notes

**Checkpoint**: US1 + US2 independently functional

## Phase 5: User Story 3 — Worktree-cut gate (Priority: P3)

**Goal**: `worktree` mode blocks stale roots and taken spec numbers before a branch
exists

**Independent Test**: quickstart.md scenario 3 — stale root → exit 1; `--spec` collision
→ exit 1 with next free number; fresh root + free number → exit 0

- [x] T013 [US3] In `scripts/check-merge-drift.mjs`, implement the `worktree` mode driver per gate-cli.md: fail-closed fetch (exit 2), block unless root is on `main` at exactly the fetched origin/main tip (`root-stale` block), `--spec <NNN>` collision block reporting the smallest free number (§5, FR-005)
- [x] T014 [US3] Validate US3 against quickstart.md scenario 3 (all three expected exits); record actual output in the task notes

## Phase 6: User Story 4 — Findings become board artifacts (Priority: P3)

**Goal**: `--notes` records task-attributable findings (severity ≥ warn) as
fingerprint-deduped board notes via the `backlog` CLI

**Independent Test**: quickstart.md scenario 4 — first `--notes` run appends one note
with the fingerprint line; identical second run appends nothing

- [x] T015 [US4] In `scripts/check-merge-drift.mjs`, implement `--notes` (FR-010, §7): BoardNote text per data-model.md (`[merge-drift <gate>] <severity>: <message>` + evidence + `fingerprint:` line), spec-marker attribution fallback (read-only scan of `backlog/tasks/` for `Spec: specs/NNN-…` when branch-name attribution misses), dedup by reading the task file for the fingerprint before appending, writes exclusively via `backlog task edit TASK-<N> --append-notes`, `noteWritten` reflected in the report; `backlog` CLI absent → info finding, never a crash
- [x] T016 [US4] Validate US4 per quickstart.md scenario 4 (append once, dedup on rerun, non-matching branch → report-only); record actual output in the task notes

## Phase 7: Polish & Cross-Cutting

- [x] T017 [P] Add the "Merge-drift gates" section to `CLAUDE.md` adjacent to the spec-047 TUI gate block (FR-013): the three invocations verbatim from contracts/gate-cli.md and when each is mandatory (session start / before `git worktree add` / before opening any PR)
- [x] T018 Run quickstart.md scenario 5 + offline check on this repo: two identical `session --json` runs diff empty (FR-012), wall time < 30 s (SC-002), no new reflog entries in any task worktree (SC-005/FR-009), unreachable-remote behavior per FR-014 (pr/worktree exit 2, session degrades); record timings and outputs in the task notes
- [ ] T019 PR hygiene for the TASK-131 branch: check whether any `docs/wiki/` note lists `scripts/` or `CLAUDE.md` files among its `sources:` (if so, re-pin via `/grounding-wiki:wiki-update` in the same PR), and run `node scripts/check-tui-design.mjs --changed` (expect: no `internal/tui/` touches, passes trivially)

## Dependencies & Execution Order

- **Setup (T001)** → **Foundational (T002–T004, sequential — same file)** → story phases
- **US1 (T005–T007)**: after T004. MVP — validate before proceeding.
- **US2 (T008–T012)**: after T004; independent of US1 (reuses foundational primitives, not US1 wiring)
- **US3 (T013–T014)**: after T002; independent of US1/US2
- **US4 (T015–T016)**: after T003; exercised via US1's pr mode, so run after US1
- **Polish (T017–T019)**: T017 any time after gate-cli.md is stable; T018–T019 last

### Parallel Opportunities

Single-file implementation → implementation tasks are sequential by design. Genuinely
parallel: T017 (CLAUDE.md, different file) against any implementation task; validation
tasks T007/T012/T014/T016 each run as soon as their story completes.

## Implementation Strategy

MVP = Phase 1 + 2 + US1 (T001–T007): the pr gate alone already prevents doomed PRs.
Then US2 (the janitor + matrix), then US3, US4, polish — each phase a working increment,
validated by its quickstart scenario before moving on. Single implementer
(spec-implementer subagent, Sonnet tier per plan.md Constitution Check); all commits on
the one TASK-131 branch in `.worktrees/task-131`.
