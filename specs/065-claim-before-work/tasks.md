# Tasks: Claim-Before-Work Protocol

**Input**: Design documents from `/specs/065-claim-before-work/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md,
contracts/gate-cli-delta.md, quickstart.md

**Tests**: included — FR-006 explicitly requires the two-session race simulation as a
test (or documented run); the automated test was chosen (research.md D6).

**Organization**: grouped by user story; US1 = doctrine (the protocol itself),
US2 = mechanical gates, US3 = branch auditability, US4 = race simulation.

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

*(No scaffolding needed — all work lands in existing files plus one new test file;
the spec directory and board claim already exist as the claim commit.)*

- [x] T001 Re-read normative surfaces before touching code: `specs/051-merge-drift-gates/contracts/gate-cli.md`, `specs/065-claim-before-work/contracts/gate-cli-delta.md`, `specs/065-claim-before-work/data-model.md`

---

## Phase 2: Foundational (blocking prerequisites for the gate stories)

**Purpose**: shared parsing/validation surface both new checks and the hook depend on

- [x] T002 Extend arg parsing in `scripts/check-merge-drift.mjs`: add `claim` to MODES; add `--dir <NNN-slug>` (claim-only, regex `^\d{3,}-[A-Za-z0-9._-]+$`) and `--task TASK-<n>` (worktree-only, regex `^TASK-\d+$`); usage errors (exit 2) on wrong-mode/invalid-value, mirroring existing `--spec`/`--branch` handling
- [x] T003 Add origin/main tree readers in `scripts/check-merge-drift.mjs`: `cardStatusOnOriginMain(originMainTip, taskId, cwd)` (ls-tree `backlog/tasks/`, filename regex `^task-<n>[ .]`, `git show` + frontmatter `status:` parse; missing/unparseable → null) — pure read-only helpers per the mutation whitelist

---

## Phase 3: User Story 2 — the gates stop the second session mechanically (P1) 🎯 MVP

**Goal**: claim-time spec-number block + worktree card-claim warning
**Independent test**: quickstart.md §1 and §2 runs produce the contracted exits/findings

- [x] T004 [US2] Implement `claim` mode (`runClaim`) in `scripts/check-merge-drift.mjs`: fetch (fail closed → exit 2), block (exit 1, rule `spec-number-collision`) iff `takenSpecNumbers()` maps NNN to a different dirname; message names taken dir + next free number; same-dirname passes (idempotence); `--json` report with `mode: "claim"`; attribution via `attributeBySpecDir`
- [x] T005 [US2] Implement `card-not-claimed` warn in worktree mode (`runWorktree`) in `scripts/check-merge-drift.mjs`: when `--task` given, read card status from the fetched origin/main tree (T003 helper); status ≠ `In Progress` (or card missing) → warn finding naming the task, card path evidence, claim-doctrine one-liner; never changes exit code by itself. ALSO make `--spec` claim-aware per the contract delta: `--spec NNN --task TASK-<n>` passes when the taken `specs/NNN-*` dir attributes to `TASK-<n>` via the Spec marker (origin/main tree), blocks when it attributes elsewhere or to none; `--spec` without `--task` keeps pre-065 semantics
- [x] T006 [US2] Extend `pre-bash` in `scripts/hooks/merge-drift-hook.mjs`: match spec-dir-creating Bash commands (`mkdir` with `specs/<NNN>-<slug>` segment; `create-new-feature.sh` with derivable `--number`/`--short-name` target) → run `claim --dir NNN-slug` from the effective dir, block on gate exit ≥ 1; also derive `--task TASK-<n>` from `git worktree add` commands (`task-<n>` in dir/branch args) and pass through to worktree mode; fail-open posture unchanged
- [x] T007 [US2] Add `pre-write` subcommand to `scripts/hooks/merge-drift-hook.mjs`: read PreToolUse stdin JSON, extract `specs/(\d{3,})-([^/]+)/` from `tool_input.file_path`, run `claim --dir NNN-slug` (jurisdiction + fail-open rules identical to pre-bash), exit 2 with findings on stderr when the gate blocks
- [x] T008 [US2] Wire `pre-write` in `.claude/settings.json`: PreToolUse entry with matcher `Write|Edit` invoking `node "$CLAUDE_PROJECT_DIR/scripts/hooks/merge-drift-hook.mjs" pre-write`, alongside the existing Bash entry

**Checkpoint**: quickstart §1/§2 pass by hand; existing `session`/`worktree`/`pr`
behavior unchanged (`node scripts/check-merge-drift.mjs session` clean run)

---

## Phase 4: User Story 3 — in-flight work auditable from any clone (P2)

**Goal**: unpushed task branches are a visible session-gate finding
**Independent test**: quickstart.md §3 — finding appears for a local-only branch with
commits, clears after push

- [x] T009 [US3] Add `branch-unpushed` warn to session mode (`runSession`) in `scripts/check-merge-drift.mjs`: for each live `task-*` branch ahead of its merge base with origin/main, warn when `refs/remotes/origin/<branch>` is absent post-fetch; message prescribes `git push -u origin <branch>`; attribution via `attributeTask` (board-notes eligible, fingerprint-deduped)

---

## Phase 5: User Story 1 — doctrine: the protocol itself (P1)

**Goal**: every session and every sweep runbook states the claim protocol
**Independent test**: quickstart.md §5 greps; a reader can execute the claim from the
doc alone

- [x] T010 [P] [US1] Add "Claim-before-work protocol (spec 065)" block to `CLAUDE.md`, adjacent to the worktrees block: first commit claims card (In Progress) + spec number (directory), pushed immediately; never force-push; rejected push = stop-the-lane signal (fetch, re-read board + specs/, surface to operator if another session holds the claim; unrelated rejection → rebase and re-push); task branches push on first commit; name the `claim`/`--task` gate invocations verbatim
- [ ] T011 [P] [US1] Companion doctrine in praxisflux source repo `~/neumo/projects/praxis/pdlc/skills/sweep/templates/runbook.md` ("Concurrency & conflict doctrine" section): same three-rule protocol for executing sessions; follow that repo's laws (version-lockstep bump, merge-commit-only); record the companion commit hash on TASK-139 — NOT part of this repo's PR

---

## Phase 6: User Story 4 — two-session race simulation (P2)

**Goal**: automated proof the loser is stopped by rejection + gates
**Independent test**: quickstart.md §4 — `node --test scripts/claim-protocol.test.mjs`

- [x] T012 [US4] Create `scripts/claim-protocol.test.mjs` (node:test, stdlib only; fixture pattern from `check-merge-drift.test.mjs`): bare origin + clones A/B in tmpdir with minimal repo shape (backlog/tasks card, specs/, main branch); tests: (1) A's claim push accepted; (2) B's competing push rejected non-fast-forward; (3) post-fetch, B's `claim --dir <NNN-other>` exits 1 naming A's dir; (4) B's `worktree --task` warns `card-not-claimed` for an unclaimed card and stays quiet for A's claimed card; (5) `claim` idempotence — A re-running against its own dirname exits 0; (6) `branch-unpushed` fires for a local-only task branch and clears when pushed

---

## Phase 7: Polish & cross-cutting

- [x] T013 Full regression: `node --test scripts/check-merge-drift.test.mjs scripts/claim-protocol.test.mjs` green; `node scripts/check-merge-drift.mjs session` and `worktree --spec 065 --task TASK-139` behave per contracts; update the gate script's header comment to point at both contracts (051 + 065 delta)
- [x] T014 Grounding freshness: check whether `docs/wiki/` notes list `CLAUDE.md`, `scripts/check-merge-drift.mjs`, or `scripts/hooks/merge-drift-hook.mjs` as sources (`grep -l` over `docs/wiki/*.md` frontmatter); if so, flag for `/grounding-wiki:wiki-update` in the PR flow (orchestrator runs it post-merge)

---

## Dependencies & execution order

- T001 → T002 → T003 (foundational chain)
- US2 (T004–T008): T004 after T002; T005 after T002+T003; T006–T007 after T004/T005; T008 after T007
- US3 (T009): after T002 only — parallel with US2's hook tasks
- US1 (T010, T011): parallel with everything ([P]); T010 should name the final flag spellings, so draft after T002 settles them
- US4 (T012): after T004, T005, T009 (tests all three behaviors)
- T013–T014: last

## Implementation strategy

MVP = Phase 3 (US2): the mechanical stop is the highest-value slice and exercises the
full architecture (gate mode + both hook layers). Then US3 (one finding), US1 (docs),
US4 (simulation locking it all in), polish. Within this repo everything lands on the
single task branch `task-139-claim-before-work` and merges in TASK-139's one PR; T011
is the recorded companion change in the praxis repo.
