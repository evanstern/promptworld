# Feature Specification: Wiki grounding moves inside the PR cycle

**Feature Branch**: `069-wiki-in-pr-gate`

**Created**: 2026-07-26

**Status**: Draft

**Input**: Operator direction (2026-07-26, recorded on TASK-145): "wiki updates
belong to the PR"; lifecycle (1) design (2) code (3) approval (4) wiki grounding
(5) PR (6) merge (7) close task + commit main. Motivating case: TASK-141 merged
with 10 `wiki-sources-overlap` warnings and left a 3-commit post-merge tail on
main (wiki re-pin, player-docs refresh, board close-out).

## Current state (pinned)

- `scripts/check-merge-drift.mjs` pr mode emits `wiki-sources-overlap` at
  severity **warn** (`gatePr`, ~line 1433): branch touches files a wiki note
  pins → advisory only. The PreToolUse hook (`scripts/hooks/merge-drift-hook.mjs`)
  blocks PR creation only on exit 1, and warnings never produce exit 1.
- CLAUDE.md's PDLC loop and Grounding-freshness rule prescribe
  `/grounding-wiki:wiki-update` AFTER merging; the player-docs rule runs "after"
  wiki-update. Constitution v1.1.0 Principle IV says re-verification must happen
  before a change "is done" but names no choke point, so practice defaulted to
  post-merge.
- Board/spec bookkeeping (card moves, tasks.md ticks, spec-bridge sync, runbook
  logs) commits to main at root — spec 065's claim protocol and the pdlc:sweep
  re-ground step already design this ("step 7").

## Decisions

1. **The pin-vs-branch rule (the blocking predicate).** For a PR branch with
   tip T, merge-base B against origin/main, and changed files F: for every wiki
   note N whose `sources:` intersect F, the branch itself must carry the
   re-verification —
   - N is modified on the branch (N ∈ F), AND
   - N's `verified_against` at T is a commit reachable from T, AND
   - no matched source of N changed after that pin on the branch
     (`git rev-list <pin>..T -- <matched sources>` is empty).
   Any note failing this blocks pr mode (exit 1) with rule
   `wiki-repin-missing` naming the note, the matched sources, and the exact
   remedy. The old warn-level `wiki-sources-overlap` finding is replaced by
   this rule for notes that fail, and SUPPRESSED for notes that pass (a
   satisfied re-pin is not noise).
2. **Player docs ride the same PR.** When the branch modifies any
   `docs/wiki/*.md`, pr mode runs the existing checker
   (`.claude/skills/player-docs/scripts/check-freshness.mjs --check`) in the
   worktree and BLOCKS on exit 1 (rule `player-docs-stale`). Exit 2
   (environment error) blocks with its own message rather than passing
   silently. Operator decision 2026-07-26: in-PR, not post-merge.
3. **Merge-commit-only for PRs.** In-branch pins are branch commit hashes; a
   squash merge rewrites them out of main's history and stales every pin the
   PR carried (observed previously on this repo — the squash-rewrites-pins
   hazard). Doctrine: PRs merge as merge commits (`gh pr merge --merge`),
   documented in CLAUDE.md. The gate does not (cannot) enforce the merge
   button; the doctrine line and the PR template of habit do.
4. **Step 7 is already designed — name it, don't reinvent it.** "Close task +
   commit main" = spec 065's claim/board protocol plus the pdlc:sweep
   re-ground step: spec-bridge sync, tasks.md ticks, board Done, runbook log —
   **derived state only**. After this spec, those are the ONLY sanctioned
   post-merge main commits; grounding content (wiki notes, player docs, design
   references) always rides the PR.
5. **No bypass flag.** A gate with an escape hatch is a warning with extra
   steps. Emergencies go through the operator editing hook config, which is
   visible and deliberate.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A code PR cannot open without its wiki grounding (Priority: P1)

A task branch changes `internal/*.go` files that wiki notes pin. The developer
(human or agent) runs the pr gate before opening the PR: it exits 1 and names
each note needing a re-pin. After running `/grounding-wiki:wiki-update` inside
the worktree (notes re-verified, re-pinned to a branch commit), the gate exits
0 and the PR carries code + wiki atomically.

**Why this priority**: This is the operator direction itself.

**Independent Test**: fixture-repo tests in the existing
`scripts/check-merge-drift.test.mjs` node:test harness — branch touching a
pinned source with (a) no note change → block; (b) note re-pinned to branch
HEAD → pass; (c) note re-pinned, then source touched again after the pin →
block.

**Acceptance Scenarios**:

1. **Given** a branch touching `internal/sim/landing.go` (pinned by a note)
   with no change to that note, **When** pr mode runs, **Then** exit 1 with
   `wiki-repin-missing` naming the note and `internal/sim/landing.go`.
2. **Given** the same branch after the note is re-pinned to a branch commit
   later than the last source change, **When** pr mode runs, **Then** exit 0
   and no `wiki-sources-overlap`/`wiki-repin-missing` finding for that note.
3. **Given** a re-pinned note whose pinned commit is NOT reachable from the
   branch tip (e.g. pin left pointing at origin/main after a rebase),
   **When** pr mode runs, **Then** exit 1 with the reachability failure named.
4. **Given** a branch touching no pinned sources, **When** pr mode runs,
   **Then** behavior is unchanged from today.

---

### User Story 2 - Player docs cannot go stale through a merge (Priority: P1)

A branch's wiki re-pin moves `verified_against` on notes that player pages
cite. The pr gate runs the player-docs freshness checker in the worktree and
blocks until the pages are regenerated in the same branch.

**Why this priority**: Operator decision — player docs follow the same
principle; without this, the gate just moves the post-merge tail one artifact
downstream.

**Independent Test**: fixture test with a stubbed checker script path
(injectable for tests) asserting exit-1 → block, exit-0 → pass, exit-2 →
block-with-env-error; plus one integration run in this repo's real worktree
during implementation.

**Acceptance Scenarios**:

1. **Given** a branch modifying a wiki note a player page cites without
   updating the page, **When** pr mode runs, **Then** exit 1 with
   `player-docs-stale`.
2. **Given** the page regenerated in-branch (meta pins current), **When** pr
   mode runs, **Then** no `player-docs-stale` finding.
3. **Given** a branch touching no `docs/wiki/` files, **When** pr mode runs,
   **Then** the checker is not invoked at all.

---

### User Story 3 - The doctrine says what the gate enforces (Priority: P2)

A fresh session reading CLAUDE.md and the constitution learns the in-PR
lifecycle — wiki grounding before the PR, merge-commit-only, and step 7 as
derived-state-only main commits — without having watched this sweep happen.

**Why this priority**: gates without doctrine breed workarounds; doctrine
without gates breeds drift. Both land together, but the gate (US1/US2) is the
enforcement and could ship alone.

**Independent Test**: doc review against the four TASK-145 ACs; constitution
version bumped with a Sync Impact Report; CLAUDE.md PDLC loop diagram shows
wiki grounding inside the build step.

**Acceptance Scenarios**:

1. **Given** the merged PR, **When** reading CLAUDE.md, **Then** the PDLC loop
   places wiki-update (and player-docs) INSIDE the task branch before the PR,
   names merge-commit-only, and the "Player docs" rule no longer says
   "after /grounding-wiki:wiki-update re-pins" as a post-merge step.
2. **Given** the constitution, **Then** Principle IV names the pr gate as the
   choke point (MINOR bump via speckit-constitution, Sync Impact Report
   updated).
3. **Given** the spec's step-7 section, **Then** it cites spec 065 and the
   sweep re-ground step as the existing design, with the derived-state-only
   boundary stated.

---

### Edge Cases

- **Wiki-only PRs** (e.g. a corpus restructure): they touch notes but no
  pinned sources → US1's predicate is vacuous per note unless that note's own
  sources changed; the player-docs check (US2) still runs. No special case.
- **A branch deletes a source file**: deletion is a change to a pinned source
  → the note must be re-verified in-branch (its `sources:` list updated), same
  predicate.
- **A branch deletes a note** while touching its sources: N ∈ F holds
  (deletion is a modification); the predicate's pin clauses are vacuous for a
  file that no longer exists at T — deleting a note IS a re-verification
  outcome (structural drift), allowed.
- **Test-file sources**: notes pin `*_test.go` files too (testing-strategy
  note); the predicate treats them like any source.
- **Merge-from-main refresh commits** on the branch: pins stay reachable from
  T; sources not re-touched after the pin still pass — no false block.
- **Malformed/missing frontmatter** in a touched note at T: existing
  `wiki-note-malformed` finding escalates to block severity in pr mode ONLY
  for notes the predicate needs (a note you must re-pin but whose pin is
  unreadable cannot pass).
- **Uncommitted worktree changes**: the gate reads the branch TIP (committed
  state), as pr mode does today; an uncommitted re-pin doesn't count until
  committed (`dirty-worktree` info finding already exists).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: pr mode MUST block (exit 1) with rule `wiki-repin-missing` for
  every wiki note whose pinned sources intersect the branch's changed files
  unless the branch itself re-verifies the note per the pin-vs-branch rule
  (note modified on branch; `verified_against` at tip reachable from tip; no
  matched source changed after the pin).
- **FR-002**: pr mode MUST NOT emit the advisory `wiki-sources-overlap` for
  notes that satisfy FR-001's predicate (replaced, not duplicated).
- **FR-003**: When the branch modifies any file under `docs/wiki/`, pr mode
  MUST run the player-docs freshness checker in the worktree and block on
  exit 1 (`player-docs-stale`) or exit 2 (`player-docs-env-error`); it MUST
  NOT invoke the checker when no `docs/wiki/` file changed.
- **FR-004**: The blocking findings MUST name the note(s)/page(s), the
  matched sources, and the remedy command (`/grounding-wiki:wiki-update` in
  the worktree; the player-docs skill for pages) in their message text.
- **FR-005**: `wiki-note-malformed` MUST escalate from warn to block in pr
  mode for exactly the notes FR-001 requires a readable pin from.
- **FR-006**: Existing session/worktree/claim modes MUST be behavior-unchanged
  (the block is a PR-choke-point rule; session keeps its advisory posture).
- **FR-007**: CLAUDE.md MUST be updated: PDLC loop diagram and Grounding
  freshness + Player docs rules moved to in-branch/pre-PR wording;
  merge-commit-only doctrine stated; step-7 boundary (derived-state-only
  post-merge main commits) stated with its citations (spec 065; pdlc:sweep
  re-ground).
- **FR-008**: Constitution Principle IV MUST be amended via
  `speckit-constitution` (MINOR bump) to name the pr gate as the choke point
  and the derived-state-only boundary; Sync Impact Report updated in the same
  change.
- **FR-009**: All new gate logic MUST be covered in the existing node:test
  harness (`scripts/check-merge-drift.test.mjs` pattern, fixture repos where
  needed); the full harness and `claim-protocol.test.mjs` MUST stay green.
- **FR-010**: No bypass flag. The gate's only pass is satisfying the
  predicate.

### Key Entities

- **Wiki note frontmatter** (`sources:` list, `verified_against:` hash) — read
  at the BRANCH TIP for the predicate (new) in addition to origin/main tip
  (existing session logic).
- **pr-mode finding** — new rules `wiki-repin-missing`, `player-docs-stale`,
  `player-docs-env-error`; severity block.
- **Doctrine docs** — CLAUDE.md, `.specify/memory/constitution.md`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The fixture-repo test matrix for US1 (block / pass / re-touched
  source / unreachable pin / untouched sources) passes; running pr mode in a
  real worktree that touches a pinned source without a re-pin exits 1.
- **SC-002**: The player-docs matrix for US2 passes; a wiki-touching branch
  with stale player pages cannot produce a PR through the hook.
- **SC-003**: `node --test scripts/check-merge-drift.test.mjs
  scripts/claim-protocol.test.mjs` fully green; session mode output on this
  repo is byte-comparable (same findings) before/after except where the spec
  changes it.
- **SC-004**: CLAUDE.md and constitution updated per FR-007/FR-008; the next
  task's PR (TASK-144 or TASK-146 in this sweep) runs under the new gate and
  merges with zero post-merge grounding commits on main.
- **SC-005**: TASK-145's four board ACs check true against these artifacts.

## Assumptions

- The praxisflux upstream halves (grounding-wiki skill prose, pdlc:sweep
  step-9 wording) are out of scope (operator decision 2026-07-26); this spec
  changes only promptworld's gates and docs. Until upstream catches up, the
  skills' prose may still SUGGEST post-merge flow — CLAUDE.md's project rules
  override skill prose in this repo.
- `gh pr merge --merge` (merge commit) is the standing merge method; the gate
  does not enforce it (Decision 3).
- The player-docs checker's exit codes (0 fresh / 1 stale / 2 env) are stable
  (spec 026 contract).
