# Card-format policy (TASK-168) — sweep runbook (2026-07-29)

**You (the session reading this) are the ORCHESTRATOR** for the task below. Run it
through the host project's full PDLC — spec → link → worktree → delegated implementation →
PR → merge → re-ground — parallelizing within lanes, merging serially, treating merge
conflicts as routine. Direction is decided; do not re-litigate it: the TASK-168 card
(operator-authored, with the operator's own format examples) wins. Plan-of-record is the
board; this file carries only ordering, doctrine, and the log.

**Status:** signed-off · operator sign-off on lanes: 2026-07-29
<!-- Only the OPERATOR flips draft → signed-off (the author never pre-fills it). An
     executing session must refuse a runbook whose status it cannot verify. -->

## Read first (in this order)

1. The TASK-168 card itself (`backlog task view TASK-168 --plain`) — it is both the
   direction source AND the first example of the format it mandates (operator-authored
   good/less-good/bad examples included).
2. Project CLAUDE.md — the Backlog.md and Spec Kit blocks are the likely durable homes
   the policy lands in or is referenced from.
3. `backlog task list --plain` — live state; other sessions move it while you work.
4. The task you're about to execute (`backlog task view TASK-168 --plain`).

## State when this runbook was written (2026-07-29, origin/main b3601b0)

- **Done already:** TASK-163 (PR #128, merged 46f55b1) was the last merge; board current.
- **In flight in other sessions (do not duplicate; expect their merges):** none — no
  In Progress tasks, no local task branches, no stale worktrees.
- **Paused — untouched (`paused` label in the task's frontmatter `labels:`; excluded
  from lane conflict analysis; never claim, rebase, or clean their
  branches/worktrees):** none.
- **Queued (this runbook's scope):** TASK-168 only.

## Execution lanes (dependency-ordered; parallelize within a lane)

Rule of thumb: DEVELOP in parallel, MERGE serially. One task → one lane; no conflict
analysis needed beyond the hotspots below.

**Lane 1 — start immediately:**
- **TASK-168 (Sonnet — routine tier: docs-only policy authoring, no code, single
  surface; constitution Principle V "doc reconciliation" class)** — write the
  card-format policy (gist-first opening + 'As a <role>' use cases, with the
  operator's good/less-good/bad examples) into a durable tracked home card authors
  and the spec agent load (CLAUDE.md Backlog.md block and/or a doc it references);
  state when use cases may be skipped (pure infra/bookkeeping cards omit use cases,
  never the gist); point spec-phase guidance at the gist section as the primary
  statement of intent. Repo-local only — NO praxisflux plugin/template changes
  (AC #3 is a scope gate, not a suggestion).

Spec rigor: full Spec Kit (specify → clarify only if ambiguous → plan → tasks →
implement) — the trivial exemption does not apply (no file:line surgical diagnosis on
the card; the durable home is a design decision the spec phase makes). Spec number
candidate: **087** (086 is the highest on origin/main at authoring time) — re-verify
with the claim gate at claim time; renumber on conflict.

Record the model tier + rubric justification on the board task at dispatch
(one-way escalation only; escalations are operator checkpoints).

## Per-PR gates this project enforces (enumerated — implementers cannot miss these)

- **Merge-drift gate: present at `scripts/check-merge-drift.mjs`.** Mandatory at every
  choke point — four modes, probed at the sweep's precondition gate (session mode ran
  2026-07-29: verdict=pass), invocations verbatim:
  `node scripts/check-merge-drift.mjs session` at sweep start (janitor + drift matrix) ·
  `node scripts/check-merge-drift.mjs claim --dir 087-card-format-policy` before
  creating the `specs/087-*` dir ·
  `node scripts/check-merge-drift.mjs worktree --spec 087 --task TASK-168` before
  `git worktree add` ·
  `node scripts/check-merge-drift.mjs pr` from the worktree before every
  `gh pr create` AND after every history move (merge-in) — nonzero exit blocks; its
  wiki-repin-missing / player-docs-stale / player-docs-env-error findings block with
  no bypass flag.
- **TUI design gate:** `node scripts/check-tui-design.mjs --changed` before any PR
  touching `internal/tui/` — this task touches none, so it should pass trivially; run
  it anyway if the diff unexpectedly grows.
- **Wiki-in-PR (spec 069):** no wiki note pins CLAUDE.md as a source (verified: only
  `docs/wiki/guardian.md` mentions it, in prose; its `sources:` are
  `internal/guardian/guardian.go` + `internal/skin/skin.go`), so no re-pin is
  expected — but the pr gate probes this mechanically; trust the gate, not this note.
- **Merge-commit-only:** `gh pr merge --merge`; never squash (squash rewrites any
  in-branch pins — observed hazard on this repo).
- **Board hygiene:** the task branch never commits `backlog/`; card moves happen at
  root via the `backlog` CLI, committed scoped to the specific task files and pushed
  immediately (board-sync exception, TASK-161).
- **Re-ground obligations:** after merge — tick spec 087 tasks.md at root-derived
  worktree flow (ticks ride a branch, never a root commit), `spec-bridge:sync` moves
  TASK-168 to Done (the ONLY sanctioned path to Done for a linked task), wiki refresh
  only if the merge touched a note's sources (none expected), player-docs freshness
  probe via the pr gate.

## Concurrency & conflict doctrine

- **Hotspots:** `CLAUDE.md` (project root — every session loads it; any concurrent
  policy-block edit conflicts textually), `backlog/tasks/` (board-sync at root only),
  `.specify/` guidance files if AC #4 lands there. All low-traffic right now (no
  sibling sessions in flight at authoring time).
- **Paused tasks are not live lanes:** none exist at authoring time; if one appears
  mid-sweep, it is never claimed, rebased, or cleaned.
- Reconcile by what the branch carries: a **pin-carrying branch merges `origin/main`
  in** (squash/rebase/force-push all stale carried pins); a **pin-free branch
  rebases** — but NOTE: this repo bans rebases everywhere (root-read-only doctrine,
  TASK-160), so on promptworld the reconcile move is ALWAYS merge-in, pin-carrying or
  not. Take main's side for anything you didn't deliberately change.
- **Honest re-pins only — a merge-in never justifies a pin bump.** Route every staled
  or conflicted pin through the wiki-update classifier: read
  `git diff <old-pin>..<merge-commit> -- <sources>`, classify RE-PIN-ONLY vs
  NEEDS-REVIEW (re-verify prose against the diff BEFORE bumping). The merge commit is
  the re-pin *target* once verified, never the *justification*.
- After every history move: re-run gates AND the freshness probe unconditionally.
- **Claim before work:** the FIRST commit of the task branch claims it — board card →
  In Progress (root board-sync commit, pushed immediately — the mutual-exclusion
  event) AND the `specs/087-*` stub directory on the branch. Push the branch on first
  commit (`git push -u origin task-168-card-format-policy`); never force-push a claim.
- **A rejected push means you lost the race:** fetch, re-read the board and `specs/`.
  If another session holds TASK-168 or number 087, STOP and surface to the operator.
  Unrelated rejection with task+number still free → fetch, merge `origin/main` into
  the claim branch, re-push plain.
- Verify a PR is merged (`gh api repos/{owner}/{repo}/pulls/<n> --jq .merged`) before
  deleting its branch/worktree; never delete+recreate a closed PR's head.

## Operator checkpoints (do not proceed silently)

- **Lane sign-off** (this file, before execution) — the only pre-planned checkpoint.
- The durable home for the policy (CLAUDE.md block vs. referenced doc) is a spec-phase
  design decision, not an operator checkpoint — the card explicitly delegates it
  ("e.g., the Backlog.md block of the project CLAUDE.md, or a doc it references").
  Surface it to the operator only if the spec phase finds the options genuinely
  contested.
- Tier escalations (Sonnet → Opus); lane amendments (amend this file, note why, tell
  the operator).

## Done means

TASK-168 Done on the board via spec-bridge:sync after its single merged PR
(merge commit, spec 087 linked and ticked); merge-drift session gate green on main;
no wiki pins staled (none expected — verify, don't assume); player-docs freshness
passing; `git worktree list` shows no sweep worktrees; this file's log complete and
status flipped to done.

## Execution log

| date | task | PR | merge | notes |
|------|------|----|-------|-------|
