# Feature Specification: Claim-Before-Work Protocol

**Feature Branch**: `065-claim-before-work`

**Created**: 2026-07-25

**Status**: Draft

**Board Task**: TASK-139

**Input**: User description: "Turn git push rejection into the mutual-exclusion primitive for concurrent sessions: the first commit of any task claims the board card and the spec number, pushed immediately; a rejected push means you lost the race — fetch, re-read, and stop the lane if another session holds the claim. Enforce at the merge-drift gates so the second session is stopped mechanically rather than politely."

## Problem

The spec-number race fired five times on 2026-07-25 alone, the last collision caused by
the fix for the previous one: one session renumbered `grounded-feedback` to 063
(ff9f8e1) while another was concurrently claiming 063 for `needs-conditioned-recovery`
(e9e75c1), forcing a second reclaim to 064 (4387d32). Earlier the same evening two
sessions independently executed the *same* 059 renumber. In every case neither session
could see the other's in-flight decision at the moment it acted.

The insight this feature encodes: **git push rejection IS the compare-and-swap.** Git
has no lock file, but it has a serialization point — whoever pushes first wins, and the
loser gets a non-fast-forward rejection. Today that rejection reads as "annoying; rebase
and carry on," which is exactly how duplicate work happens. Treating the rejection as
*signal* upgrades a polite convention into mutual exclusion.

## Scope Honesty *(what this fixes, and what it does not)*

- **Two sessions claiming the same TASK**: fixed — exactly this.
- **Two sessions claiming the same SPEC NUMBER**: fixed, but *only* because the number
  claim rides the same push as the card claim. Without that clause the protocol does
  nothing for numbers.
- **In-flight CODE invisible from other clones**: NOT fixed by the claim. Card moves
  already propagate today (seven tasks read In Progress in every clone while their
  branches existed on one machine only). Code visibility needs the separate
  push-on-first-commit rule for task branches — the two rules are *paired* in this
  feature, never conflated.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A session claims a task before working it (Priority: P1)

An orchestrating session picks up a TASK from the board. Before any spec authoring or
code, its first commit *claims* the work: it moves the board card to In Progress AND
claims the spec number by creating the spec directory, then pushes immediately (never
force-pushing). If the push is accepted, the session owns the task and the number in
every clone that fetches. If the push is rejected, the session lost the race: it
fetches, re-reads the board and `specs/`, and — if another session now holds that task
or that number — stops the lane and surfaces the collision to the operator.

**Why this priority**: this is the protocol itself; every other requirement exists to
enforce or audit it.

**Independent Test**: on a fresh clone, execute the claim sequence for a new task and
verify a single pushed commit contains both the card move and the spec directory; then
simulate a rejected push and verify the documented stop path is followed.

**Acceptance Scenarios**:

1. **Given** a To Do task and a free spec number, **When** a session starts the task,
   **Then** its first pushed commit contains the card move to In Progress and the new
   `specs/NNN-<slug>/` directory, and the push precedes all spec authoring and code.
2. **Given** two sessions racing for the same task or number, **When** the slower
   session's claim push is rejected, **Then** that session fetches, observes the other
   session's claim, stops the lane, and surfaces the collision to the operator instead
   of rebasing and continuing.
3. **Given** the doctrine documents, **When** a reader consults the project grounding
   (CLAUDE.md) or a sweep runbook authored from the template, **Then** the claim
   sequence, the push-immediately rule, the never-force-push rule, and the
   rejected-push-means-stop rule are all stated.

---

### User Story 2 - The gates stop the second session mechanically (Priority: P1)

A second session, unaware of the first session's claim, tries to start the same work.
The merge-drift gates stop it at the existing choke points: cutting a worktree for a
task whose card is not In Progress on `origin/main` warns (the session skipped the
claim, or the claim never propagated); creating a spec directory whose number is
already taken on `origin/main` is blocked *at claim time* — before the directory is
authored against — not detected later at PR time.

**Why this priority**: without mechanical enforcement the protocol is another polite
request, which is the failure mode being replaced.

**Independent Test**: with a claim present on `origin/main`, run the worktree gate for
an unclaimed task (expect a warning) and attempt to create a colliding spec directory
(expect a block naming the taken directory and the next free number).

**Acceptance Scenarios**:

1. **Given** a task whose board card is still To Do on `origin/main`, **When** the
   worktree gate runs for that task, **Then** it warns that the card is not In Progress
   on `origin/main` (claim missing or unpushed).
2. **Given** a task whose card is In Progress on `origin/main`, **When** the worktree
   gate runs for that task, **Then** no card-claim warning is raised.
3. **Given** spec number NNN already taken on `origin/main`, **When** a session
   attempts to create a new `specs/NNN-*` directory, **Then** the attempt is blocked at
   creation time with the taken directory and the next free number named.
4. **Given** spec number NNN taken only in the local worktree (not on `origin/main`),
   **When** the claim-time check runs, **Then** the verdict is based on `origin/main`
   after a fresh fetch, not on stale local state.

---

### User Story 3 - In-flight work is auditable from any clone (Priority: P2)

A task branch is pushed to the remote on its first commit, so any other session (or the
operator) can see from any clone which tasks have in-flight implementation — not just
which cards are In Progress. A local-only task branch with commits is a visible gate
finding, not an invisible fact of one machine.

**Why this priority**: pairs with the claim to close the visibility gap the claim alone
does not cover; without it, seven In Progress cards can again correspond to branches
that exist on one machine only.

**Independent Test**: create a task branch with a commit but no remote counterpart and
verify the session gate reports it; push it and verify the finding clears.

**Acceptance Scenarios**:

1. **Given** a local task branch with at least one commit and no upstream on the
   remote, **When** the session gate runs, **Then** it reports the branch as unpushed /
   not auditable from other clones.
2. **Given** doctrine documents, **When** a session cuts a task branch, **Then** the
   documented rule requires pushing the branch on its first commit.

---

### User Story 4 - Two-session race simulation (Priority: P2)

The protocol's core claim — "the second session is stopped rather than duplicating
work" — is demonstrated end to end: two simulated sessions race for the same task/spec
number against a shared remote; the first push wins; the second session's path hits the
rejected push and the gates, and stops.

**Why this priority**: the protocol was born from five real races in one day; a
simulation is the only honest proof short of waiting for the sixth.

**Independent Test**: an automated test (or a documented, reproducible manual run)
whose transcript shows the losing session stopped by rejection + gates.

**Acceptance Scenarios**:

1. **Given** two clones of a shared repository and one free spec number, **When** both
   execute the claim sequence concurrently, **Then** exactly one claim lands and the
   other clone's push is rejected; after fetching, the loser's gates warn/block on the
   now-taken task and number.

### Edge Cases

- Remote unreachable at claim-check time: the claim-time spec-number check follows the
  existing gate posture (fail closed with a distinct exit/usage error, matching
  worktree/pr modes) rather than silently passing on stale data.
- The claim push is rejected for an unrelated reason (someone pushed board notes):
  after fetch + rebase, if the task and number are still free, the claim is re-pushed —
  a rejection is a *stop-and-look* signal, not an unconditional abort.
- A session re-entering its own task (card already In Progress on origin/main because
  this same session claimed it earlier): the worktree gate's card check passes; the
  claim is idempotent from the owner's perspective.
- Renumbering an existing spec directory (the 059/063 cases): a rename lands as a new
  `specs/NNN-*` path and is subject to the same claim-time collision check.
- A spec directory created empty as a claim (no spec.md yet): counts as taken for
  numbering purposes the moment it is on `origin/main` — the claim stub is enough.
- Tasks that legitimately need no spec (trivial exemption): the card claim still
  applies; only the spec-number clause is moot.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Project doctrine (CLAUDE.md) MUST state the claim-before-work protocol:
  the first commit of any task claims the card (In Progress) and the spec number
  (directory creation) before any spec authoring or code; it is pushed immediately;
  force-push is never used; a rejected push is a stop-the-lane signal requiring fetch,
  re-read of board + `specs/`, and operator surfacing when another session holds the
  claim.
- **FR-002**: The sweep runbook template (pdlc plugin) MUST carry the same doctrine in
  its concurrency section, so every runbook authored from it instructs executing
  sessions to claim before work and to treat push rejection as signal.
- **FR-003**: The worktree-mode merge-drift gate MUST accept the task identity and warn
  when that task's board card is not In Progress on `origin/main` at the moment the
  worktree is cut (severity: warning — the card may be legitimately mid-propagation,
  and blocking would fight the claim commit itself).
- **FR-004**: Creating a new `specs/NNN-*` directory whose number is already taken on
  `origin/main` MUST be blocked at creation (claim) time, with the taken directory and
  the next free number named — using the same collision source of truth the PR gate
  already uses (`takenSpecNumbers()`), evaluated against a freshly fetched
  `origin/main`.
- **FR-005**: Task branches MUST be pushed to the remote on their first commit, and the
  session-mode gate MUST surface any local task branch with commits but no remote
  counterpart as a finding, so in-flight work is auditable from any clone.
- **FR-006**: A two-session race MUST be demonstrated: an automated test or a
  documented, reproducible manual run showing the second session stopped (rejected push
  + gate warning/block) rather than duplicating the claim.
- **FR-007**: The claim-time spec-number check MUST fail closed when the remote cannot
  be reached, consistent with the existing worktree/pr gate posture.
- **FR-008**: All new enforcement MUST remain within the existing gate architecture —
  exit codes + findings from `scripts/check-merge-drift.mjs`, harness-level enforcement
  via the existing PreToolUse hook wiring — with no daemon, no CI dependency, and no
  writes to task branches or worktrees by the gate itself.

### Key Entities

- **Claim**: the first pushed commit of a task — board card move to In Progress + spec
  directory creation, indivisible in one push. Ownership is defined by presence on
  `origin/main`, not by local state.
- **Claim race**: two sessions attempting overlapping claims; resolved by push
  ordering. The loser is identified by non-fast-forward rejection.
- **Taken spec number**: a `specs/NNN-*` directory present on `origin/main` (content
  optional — a stub claims the number).
- **Unpushed task branch**: a local branch matching the task-branch naming convention
  with commits but no remote counterpart — the audit gap FR-005 closes.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In a two-session race simulation for the same task/number, the losing
  session performs zero duplicate work past the claim point (no second spec directory,
  no second renumber) — stopped by rejection + gates alone.
- **SC-002**: A spec-number collision is surfaced at claim time — before any spec
  content is authored — rather than at PR time; the collision report names the taken
  directory and the next free number.
- **SC-003**: Cutting a worktree for an unclaimed task produces a visible warning in
  100% of runs where the card is not In Progress on `origin/main`.
- **SC-004**: Every task branch created after adoption is visible on the remote from
  its first commit; the session gate reports any exception.
- **SC-005**: Zero spec-number or task-claim collisions reach PR time in the first
  month after adoption (baseline: five in one day on 2026-07-25).

## Assumptions

- The runbook template lives in the praxisflux plugin source repository
  (`~/neumo/projects/praxis`), not in this repository; FR-002 is a companion change
  there, following that repo's own laws (version-lockstep, merge-commit-only). This
  repo's PR carries FR-001 and FR-003–FR-006; the board task records the companion
  change.
- Board state on `origin/main` is readable from the tracked `backlog/tasks/*.md` files
  (status lives in the task file); no backlog CLI invocation is needed inside the gate.
- "Claim time" for spec directories is enforceable at the harness level via the
  existing PreToolUse hook mechanism (the same wiring that gates `gh pr create` and
  `git worktree add`); sessions not running under the harness still get the doctrine
  and the PR-time collision check as backstop.
- The existing task-branch naming convention (`task-<n>-<slug>`) identifies task
  branches for FR-005.
- Card-claim enforcement at worktree time is a warning, not a block, by explicit task
  decision ("should warn"); spec-number collision at claim time is a block ("should
  block").
