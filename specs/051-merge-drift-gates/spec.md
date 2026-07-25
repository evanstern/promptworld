# Feature Specification: Merge-Drift Gates

**Feature Branch**: `051-merge-drift-gates`

**Created**: 2026-07-25

**Status**: Draft

**Input**: User description: "Merge-drift gates: a deterministic script (scripts/check-merge-drift.mjs) that gates the parallel worktree SDLC against merge drift at three choke points — session start (janitor pass: detect merged-but-present worktrees, ff-pull root, n-way merge-tree drift matrix across live branches, wiki/player-docs freshness flags), before cutting a worktree (fresh origin/main + spec-number collision check), and before opening a PR (merge-tree vs origin/main hard-fail on predicted textual conflicts, warnings on semantic overlaps: backlog/, wiki-pinned sources, internal/tui/, stale branch base). No daemon, no GitHub Actions, no PR comments — findings land as backlog notes on affected tasks plus script exit codes. Script never rebases or resolves a live task's branch; resolution stays with the owning session. Enforcement documented in CLAUDE.md following the spec-047 check-tui-design.mjs precedent."

## Problem Statement *(context)*

This project runs a highly parallel SDLC: multiple concurrent sessions each work a TASK
in its own worktree, and PRs merge to main independently. Drift between those parallel
lines fails in three distinct layers, each already observed in this repository:

1. **Textual merge conflicts** — an in-flight branch conflicts with what just landed on
   main, discovered only at PR time or after.
2. **Semantic collisions that merge cleanly** — spec-number collisions in `specs/NNN-*/`,
   concurrent writes to `backlog/`, two branches touching the same wiki-pinned sources,
   squash-merges rewriting branch-hash provenance pins. Git never flags these.
3. **Post-merge hygiene lag** — merged-but-still-present worktrees, root checkout behind
   origin, grounding freshness (wiki / player docs / TUI design reference) not re-checked,
   board not synced.

Today these are caught by memory and discipline. This feature converts them into gates
against physical evidence, per Constitution Principle III (Gates Over Assertions). By
deliberate decision there is **no daemon, no CI service, and no external annotation**:
detection runs fresh at the choke points where a finding is actionable, because a fresh
check is cheap and a background watcher's report would only ever be consumed at those
same choke points anyway.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - PR gate: no doomed PR gets opened (Priority: P1)

A session working a TASK in its worktree is ready to open the PR. Before opening it, the
session runs the PR gate. The gate freshly syncs against the remote mainline, predicts
whether merging this branch would conflict, and checks for clean-merging semantic
overlaps with what landed on main since the branch was cut. A predicted textual conflict
blocks (non-zero exit) with a file-by-file account; semantic overlaps (board files,
wiki-pinned sources, TUI design surface, stale branch base) warn with the exact evidence.
The owning session — and only the owning session — then resolves before opening the PR.

**Why this priority**: this is the moment of maximum leverage — the one choke point every
TASK passes through, and the one where a conflict caught costs minutes instead of a
broken merge or a bricked PR. It is independently valuable even if no other gate ships.

**Independent Test**: cut two branches from the same base that edit the same lines of the
same file; merge one to main; run the PR gate on the other. The gate must block and name
the conflicting files. Run it on a branch with no overlap: the gate must pass.

**Acceptance Scenarios**:

1. **Given** a branch whose changes textually conflict with commits now on origin/main,
   **When** the PR gate runs from that worktree, **Then** it exits non-zero and reports
   each conflicting file.
2. **Given** a branch that merges cleanly but touches a file listed as a source by a
   `docs/wiki/` note, **When** the PR gate runs, **Then** it passes but emits a warning
   naming the note(s) whose grounding the merge will affect.
3. **Given** a branch that merges cleanly and overlaps nothing, **When** the PR gate
   runs, **Then** it exits zero with a clean report.
4. **Given** a branch whose base is behind origin/main, **When** the PR gate runs,
   **Then** the report states how many mainline commits the base is missing.
5. **Given** a branch and origin/main that both modified the same file under `backlog/`,
   **When** the PR gate runs, **Then** it warns and names the task file(s).

---

### User Story 2 - Session-start janitor: drift is surfaced before work begins (Priority: P2)

A new session starts at the repo root. It runs the session gate, which syncs with the
remote and reports: whether the root checkout is behind (and fast-forwards it), which
worktrees belong to branches whose work has already landed on main (cleanup-eligible,
with the exact remediation commands), the n-way drift matrix across all live branches
(which pairs are predicted to conflict with each other, informing merge order), and
whether commits newly landed on main touched sources pinned by the grounding surfaces
(wiki, player docs, TUI design reference) so freshness re-checks get queued.

**Why this priority**: this is where post-merge hygiene lag dies. It found a live example
during design: a merged PR whose worktree was still present and whose root checkout was
one commit behind, with nothing noticing.

**Independent Test**: merge a branch, leave its worktree in place, leave root un-pulled;
run the session gate. It must fast-forward root, identify that worktree as
cleanup-eligible, and prescribe its removal.

**Acceptance Scenarios**:

1. **Given** a worktree whose branch tip is an ancestor of origin/main, **When** the
   session gate runs, **Then** the worktree is reported cleanup-eligible with exact
   remediation steps.
2. **Given** a worktree whose branch was squash-merged (tip is not an ancestor, but the
   branch's cumulative diff against origin/main is empty), **When** the session gate
   runs, **Then** it is likewise reported cleanup-eligible.
3. **Given** a root checkout behind origin/main with no local divergence, **When** the
   session gate runs, **Then** the root is fast-forwarded and the movement reported.
4. **Given** two live branches that each modify the same lines of the same file, **When**
   the session gate runs, **Then** the drift matrix reports that pair as
   conflict-on-merge regardless of which merges first.
5. **Given** new commits on origin/main touching files listed as sources by wiki notes,
   **When** the session gate runs, **Then** the affected grounding surface is flagged for
   a freshness re-check.
6. **Given** a worktree with uncommitted local changes, **When** the session gate runs,
   **Then** it is never reported cleanup-eligible and the dirty state is noted.

---

### User Story 3 - Worktree-cut gate: branches fork from fresh, collision-free ground (Priority: P3)

A session is about to cut a new TASK worktree. It runs the worktree gate, which verifies
the root checkout sits exactly at the freshly-fetched origin/main tip, and — when the
TASK will create a new spec directory — that the intended spec number is unused on
origin. Cutting from a stale base or onto a taken spec number is blocked before the
branch exists.

**Why this priority**: cheapest of the three to work around manually, but spec-number
collisions have occurred in this project (numbers 049 and 050 were both taken on origin
while a session assumed 049 was free — during the design of this very feature).

**Independent Test**: hold root one commit behind origin/main and run the worktree gate:
it must block. Pull, pass a spec number that exists on origin: it must block and name the
colliding directory. Pass a free number at fresh tip: it must pass.

**Acceptance Scenarios**:

1. **Given** a root checkout not at the current origin/main tip, **When** the worktree
   gate runs, **Then** it exits non-zero and instructs a fast-forward pull first.
2. **Given** an intended spec number already present under `specs/` on origin/main,
   **When** the worktree gate runs with that number, **Then** it exits non-zero and names
   the colliding spec directory and the next free number.
3. **Given** fresh root and a free spec number, **When** the worktree gate runs, **Then**
   it exits zero.

---

### User Story 4 - Findings become board artifacts (Priority: P3)

When a gate produces a blocking finding or warning that implicates a board TASK (the
task owning the affected branch, or the task whose files overlap), that finding is
recorded as a note on the affected task — so the plan of record carries the warning, the
finding survives the session that discovered it, and the next session working that task
inherits it. No findings are published to any external service: no PR comments, no
labels, no CI annotations.

**Why this priority**: Constitution Principle I — a finding living only in a gate's
stdout did not happen. Ordered after the gates themselves because the exit codes already
block; notes make the evidence durable.

**Independent Test**: trigger a PR-gate conflict on a branch whose TASK id is derivable
from the branch name; verify a note appears on that task describing the finding.

**Acceptance Scenarios**:

1. **Given** a gate finding that implicates an identifiable board task, **When** the
   gate run is handled, **Then** the task carries a note recording the finding (severity,
   evidence, gate, date), written via the board CLI only.
2. **Given** a finding implicating no identifiable task, **When** the gate run completes,
   **Then** the finding still appears in the gate's report output and exit status.

---

### Edge Cases

- **Remote unreachable**: the PR and worktree gates fail closed (a verdict about
  origin/main cannot be issued from stale data); the session gate degrades to a local
  report explicitly marked as unverified-against-remote.
- **Squash-merged branches**: ancestor checks alone miss them; cleanup eligibility must
  also accept the empty-cumulative-diff signal (US2 scenario 2).
- **Dirty worktrees**: uncommitted changes exclude a worktree from cleanup eligibility
  and are noted; conflict prediction uses committed state only.
- **Root checkout not on main**: a violation of the worktree doctrine in its own right —
  every gate reports it as a blocking finding.
- **No live branches / no worktrees**: the drift matrix is trivially empty; gates pass
  green rather than erroring.
- **Pairwise-only conflicts**: two branches that each merge cleanly with main but
  conflict with each other are exactly what the n-way matrix exists to surface; neither
  branch's PR gate alone would see it.
- **Branch with no derivable task id**: findings are report-only (US4 scenario 2); the
  gate must not guess a task.
- **Very many live branches**: pairwise matrix cost grows quadratically; the gate must
  remain usable at this project's realistic scale (see SC-002) and report matrix size.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide a single deterministic gate command with three
  modes — session start, worktree cut, and PR open — each producing a human-readable
  report and a machine-readable verdict (exit status: pass / warnings / blocked).
- **FR-002**: The PR gate MUST predict, against a freshly-fetched origin/main and without
  modifying any worktree or branch, whether merging the current branch would produce
  textual conflicts, and MUST block (non-zero exit) when it would, naming each
  conflicting file.
- **FR-003**: The PR gate MUST report how far the branch's merge base lags the current
  origin/main tip, and warn when it lags at all.
- **FR-004**: The PR gate MUST warn on clean-merging semantic overlaps between the branch
  and commits on origin/main since the merge base, covering at minimum: same-file
  touches under `backlog/`; files listed as sources by any `docs/wiki/` note; files under
  `internal/tui/` (with a pointer to the spec-047 design-reference gate); and new
  `specs/NNN-*` directories whose number collides with one on origin/main.
- **FR-005**: The worktree gate MUST block unless the root checkout is exactly at the
  freshly-fetched origin/main tip, and — when given an intended spec number — MUST block
  if that number exists under `specs/` on origin/main, reporting the next free number.
- **FR-006**: The session gate MUST identify cleanup-eligible worktrees: those whose
  branch tip is an ancestor of origin/main, or whose cumulative diff against origin/main
  is empty (squash-merge case), and which have no uncommitted changes. For each, it MUST
  prescribe the exact remediation (worktree removal, branch deletion). It MUST
  fast-forward the root checkout when the root is behind with no local divergence.
- **FR-007**: The session gate MUST compute an n-way drift matrix: for every pair of live
  branches (and each branch against origin/main), whether merging both is predicted to
  conflict, presented so merge-order consequences are visible.
- **FR-008**: The session gate MUST flag grounding-freshness exposure: when commits new
  on origin/main since the last known-fresh state touch files pinned as sources by the
  wiki, the player docs, or the TUI design reference, the affected surface MUST be named
  for re-check (delegating to each surface's existing freshness checker where one
  exists).
- **FR-009**: No gate may ever rebase, merge, commit to, or otherwise modify any task
  branch or its worktree. The only permitted mutations are: fetching from the remote,
  fast-forwarding the root checkout on main, and — only on explicit opt-in per run —
  applying the prescribed cleanup of worktrees already verified cleanup-eligible under
  FR-006. Conflict resolution always belongs to the branch's owning session.
- **FR-010**: Every blocking finding or warning that implicates an identifiable board
  task MUST be recorded as a note on that task via the board CLI (never by editing board
  files directly), including the gate, severity, evidence, and date.
- **FR-011**: The system MUST NOT depend on any external service beyond the git remote
  itself: no hosting-provider API calls, no CI workflows, no PR comments or labels.
- **FR-012**: Gate runs MUST be deterministic: identical repository and remote state
  yields an identical verdict and findings.
- **FR-013**: The three gate invocations and when each is mandatory MUST be documented in
  the project's always-on instructions (CLAUDE.md), following the spec-047 TUI
  design-reference gate precedent.
- **FR-014**: When the remote cannot be reached, the PR and worktree gates MUST fail
  closed; the session gate MUST still run its local checks while marking the report
  unverified against the remote.

### Key Entities

- **Gate**: one of the three choke points (session start, worktree cut, PR open); has a
  mode, a verdict (pass / warnings / blocked), and a report.
- **Finding**: a single detected condition — severity (block / warn / info), the gate
  that produced it, evidence (files, commits, branches), and an optional affected board
  task.
- **Drift matrix**: the pairwise conflict-prediction table across live branches plus
  origin/main; the session gate's core artifact.
- **Cleanup-eligible worktree**: a worktree whose branch's work is verifiably contained
  in origin/main (ancestor or empty-diff) and which holds no uncommitted changes.
- **Grounding surface**: a freshness-tracked documentation layer (wiki, player docs, TUI
  design reference) whose pinned sources can be invalidated by mainline commits.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: After adoption, zero PRs are opened that a PR-gate run at open time would
  have blocked — every textual conflict with mainline is surfaced to the owning session
  before the PR exists.
- **SC-002**: A full gate run (any mode) completes in under 30 seconds at this project's
  scale (up to ~10 live branches), so it is never worth skipping.
- **SC-003**: Zero spec-number collisions reach origin after adoption.
- **SC-004**: 100% of merged-but-still-present worktrees are identified as
  cleanup-eligible by the first session gate run after their merge, including
  squash-merged branches.
- **SC-005**: Zero mutations of any task branch or worktree by gate runs, ever —
  verifiable from git history and reflogs.
- **SC-006**: Every blocking finding that implicates a board task is traceable to a note
  on that task; a sampled audit finds no orphaned blocking findings.
- **SC-007**: Grounding surfaces affected by a mainline merge are flagged by the next
  session gate run, so no stale-pin state survives a session boundary undetected.

## Assumptions

- The gates are invoked by sessions at the documented choke points (CLAUDE.md-enforced
  convention, per the spec-047 precedent); no hook wiring is in scope for v1, though the
  gates' exit codes are designed so hooks could consume them later.
- "Live branches" means branches checked out in `.worktrees/` plus any local branches
  ahead of origin/main; branches existing only on the remote are out of scope for the
  drift matrix in v1 (the PR gate covers each at its own choke point).
- The affected-task mapping uses the established branch naming convention
  (`task-<N>-<slug>`) and the spec-bridge markers; when neither identifies a task, the
  finding stays report-only.
- Janitor cleanup defaults to prescribing remediation; applying it requires explicit
  opt-in on the invocation (FR-009). Fast-forwarding the root on main is always safe and
  automatic.
- The wiki's per-note source lists, the player-docs freshness checker, and the TUI
  design-reference checker are the authoritative inputs for FR-008; this feature reuses
  them rather than re-deriving freshness.
- Conflict prediction operates on committed state; predicting conflicts against
  uncommitted work is out of scope.
- Findings recorded as board notes are advisory history; the gate's exit code — recomputed
  fresh at the next choke point — remains the enforcement mechanism.
