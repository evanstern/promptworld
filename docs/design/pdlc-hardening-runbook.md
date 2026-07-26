# PDLC hardening sweep — runbook (2026-07-26)

**You (the session reading this) are the ORCHESTRATOR** for the tasks below. Run each
through the host project's full PDLC — spec → link → worktree → delegated implementation →
PR → merge → re-ground — parallelizing within lanes, merging serially, treating merge
conflicts as routine. Direction is decided; do not re-litigate it: the operator's
2026-07-26 process direction (recorded on TASK-145: "wiki updates belong to the PR",
lifecycle design → code → approval → wiki grounding → PR → merge → close task + commit
main) and the three board cards win. Plan-of-record is the board; this file carries only
ordering, doctrine, and the log.

**Status:** signed-off · operator sign-off on lanes: 2026-07-26 (lanes as written;
player-docs placement decided IN-PR; praxisflux upstream halves confirmed out of scope
— follow-up lives in ~/neumo/projects/praxis)
<!-- Only the OPERATOR flips draft → signed-off (the author never pre-fills it). An
     executing session must refuse a runbook whose status it cannot verify. -->

## Read first (in this order)

1. TASK-145's description (carries the operator direction verbatim) and the
   [[wiki-updates-belong-to-the-pr]] session memory it mirrors; TASK-141's close-out
   notes (the motivating case: 10 wiki-overlap warnings, 3-commit post-merge tail).
2. Project gate docs: CLAUDE.md (PDLC block, merge-drift block, claim protocol,
   worktrees), `.specify/memory/constitution.md` (v1.1.0 — Principles I–V),
   `docs/design/tui/` INDEX (gate applies only to `internal/tui/` PRs — none expected
   here).
3. `backlog task list --plain` — live state; other sessions move it while you work.
4. The task you're about to execute (`backlog task view TASK-<n> --plain`).

## State when this runbook was written (2026-07-26, origin/main 93df57f)

- **Done already:** TASK-141 (spec 067, PR #104 merged; wiki + player-docs re-grounded).
- **In flight in other sessions (do not duplicate; expect their merges):**
  - TASK-111 / spec 059 — In Progress, derives Done-eligible; operator says ignore, its
    session owns the sync. Do not touch its card or spec.
  - TASK-143 / spec 068 (tile registry) — worktree `.worktrees/task-143` active, clean
    vs main in the drift matrix. Expect merges touching `internal/tui`, map/terrain
    code, and later wiki notes; refetch before every merge.
  - Root has uncommitted board-note edits to TASK-136/TASK-137 (another session's) and
    an untracked `.claude/settings.local.json:` file — leave both alone; never
    `git add backlog/` wholesale.
- **Queued (this runbook's scope):** TASK-145 → TASK-144 → TASK-146 (merge order;
  145 and 144 develop in parallel).

## Execution lanes (dependency-ordered; parallelize within a lane)

Rule of thumb: DEVELOP in parallel, MERGE serially.

**Lane 1 — start immediately:**
- **TASK-145 (Opus 4.8 for the gate/hook code — doctrine-adjacent behavior change to
  SDLC-critical infrastructure: a defect in `scripts/check-merge-drift.mjs` blocks every
  future PR in every lane; cross-cutting scripts + hooks + doctrine docs)** — the in-PR
  wiki gate: escalate `pr`-mode `wiki-sources-overlap` warn → BLOCK when the branch
  touches pinned sources without re-pinning those notes in-branch (pin must equal branch
  HEAD or an ancestor — the merge hash doesn't exist at PR time); CLAUDE.md PDLC block +
  constitution Principle IV updated; player-docs placement documented; step-7
  reconciliation named by artifact. The **constitution amendment itself is planning-tier
  work** (speckit-constitution, orchestrator hands) — only the script/hook changes are
  dispatched. Its CONTRACT (the new gate rule) governs lanes 2–3's PRs, which is why it
  merges first.
- **TASK-144 (Sonnet — single-package test fix, routine; escalate to Opus ONLY if the
  shared-state diagnosis lands in concurrency machinery — that's an operator
  checkpoint)** — de-flake `internal/guardian/TestReportCardRunEndRidesEpilogue`
  (order-dependent; fails deterministically in isolation). Disjoint file footprint from
  TASK-145 → parallel development is safe; it merges whenever ready, smaller-first
  relative to 145 if both are open.

**Lane 2 — tail (droppable), after BOTH lane-1 merges:**
- **TASK-146 (orchestrator-led grounding-docs work — constitution V leaves grounding
  docs in the orchestrator's hands; fan mechanical note-splits out to Sonnet subagents,
  keep capsule/routing judgment and the final gate flip with the orchestrator)** —
  corpus-spec v2 adoption: split/exempt the 35 over-budget notes, rewrite the 6
  oversized capsules, generate CAPSULES.md, freshness gate green in v2 (failure) mode.
  MUST NOT develop in parallel with anything that re-pins wiki notes (its footprint is
  `docs/wiki/` wholesale — the sweep's #1 hotspot), and lands under 145's new in-PR
  rules deliberately, as their first full exercise.

Record the model tier + rubric justification on each board task at dispatch
(one-way escalation only; escalations are operator checkpoints).

## Per-PR gates this project enforces (enumerated)

- **Merge-drift gate: present at `scripts/check-merge-drift.mjs`.** Mandatory at every
  choke point: `node scripts/check-merge-drift.mjs session` at sweep start and before
  each new task (janitor + drift matrix); `node scripts/check-merge-drift.mjs claim
  --dir NNN-slug` before creating any `specs/NNN-*` dir; `node
  scripts/check-merge-drift.mjs worktree --spec NNN --task TASK-<n>` before every
  `git worktree add`; `node scripts/check-merge-drift.mjs pr` from the worktree before
  every PR open AND after every rebase — nonzero exit blocks. Hook-enforced besides
  (`scripts/hooks/merge-drift-hook.mjs`); note the hook pattern-matches command TEXT, so
  board-card descriptions must not contain the literal PR-creation command string.
- **Wiki grounding rides the PR (operator direction 2026-07-26, in force for this sweep
  NOW, enforced by the gate after TASK-145 merges):** any branch touching files a wiki
  note pins re-pins/rewrites those notes IN THE SAME BRANCH before the PR opens. No
  post-merge main tails.
- **Player-docs follow the same principle** (proposed, confirmed at sign-off): when a
  branch's wiki re-pin stales `docs/player/` pages, regenerate them in the same branch
  (`node .claude/skills/player-docs/scripts/check-freshness.mjs --check` must pass
  in-branch).
- Go-touching PRs (TASK-144): `gofmt -l` clean on touched files; `go test ./...` green
  in the worktree, re-run after every rebase.
- `internal/tui/` PRs: `node scripts/check-tui-design.mjs --changed` + same-PR
  `docs/design/tui/` amendment — no lane here expects to trigger it.
- Spec rigor: full Spec Kit per task (none qualifies for the trivial exemption —
  TASK-144's diagnosis is incomplete by definition); `spec-bridge:link` BEFORE
  implementation; tasks.md ticks and board commands from repo root, never inside
  `.worktrees/`; add specific board task files to git, never `backlog/` wholesale;
  board writes via the `backlog` CLI only (`--priority`, not `-p`, which means parent).
- Tentative spec numbers (claim gate is authoritative): TASK-145 → 069,
  TASK-144 → 070, TASK-146 → 071.

## Concurrency & conflict doctrine

- **Hotspots:** `docs/wiki/**` (the sweep's own subject — lane 2 owns it exclusively);
  `scripts/check-merge-drift.mjs` + `scripts/hooks/` (TASK-145 only); `CLAUDE.md` and
  `.specify/memory/constitution.md` (TASK-145 + any concurrent session's doc edits);
  `backlog/tasks/*` (every session — scoped adds only); `.specify/feature.json`
  (per-worktree resolution; root copy churns — take main's side on conflicts).
- Rebase, never merge-commit into a task branch **when resolving conflicts**; take
  main's side for anything you didn't deliberately change; re-run gates after every
  rebase. (Precedent note: TASK-141's branch used a merge-from-main on a zero-overlap
  base refresh because its branch was already pushed and force-push is forbidden —
  acceptable for a clean fast-forwardable refresh, but conflicts resolve by rebase.)
- Two hotspot-heavy PRs never merge within one re-ground cycle without a rebase between.
- Conflicting with a sibling session's open PR → the smaller PR merges first.
- **Claim before work:** the FIRST commit of any task claims it — board card →
  In Progress AND the spec number's directory stub — before any spec authoring or code.
  Push immediately (`git push -u origin <branch>` on first commit); never force-push.
- **A rejected push means you lost the race:** fetch, re-read the board and `specs/`.
  If another session now holds that task or number, STOP the lane and surface it.
  Unrelated rejection with the task+number still free → rebase and re-push the claim.
- Verify a PR is merged (`gh api repos/<owner>/<repo>/pulls/<n> --jq .merged`) before
  deleting its branch/worktree; never delete+recreate a closed PR's head.

## Operator checkpoints (do not proceed silently)

- **At sign-off (this document):** (a) player-docs placement = in-PR (proposed above);
  (b) TASK-145 amends constitution Principle IV via `speckit-constitution` — a MINOR
  version bump whose wording the operator sees in the PR;
  (c) the praxisflux upstream halves (grounding-wiki skill prose "commit together with
  or immediately after", pdlc sweep step 9's post-merge wiki refresh) are OUT of this
  sweep's scope — follow-up work in `~/neumo/projects/praxis` under its version-lockstep
  laws; this sweep changes only promptworld's own gates/docs.
- **TASK-144 tier escalation** if the flake's root cause is concurrency machinery.
- **TASK-146 gate flip:** generating CAPSULES.md flips freshness to v2 failure mode
  repo-wide — it lands only in a PR that makes the whole corpus pass in v2 mode; if the
  split stalls, the PR closes unmerged and main stays on v1 warnings.
- Dropping/reordering lanes = runbook amendment: amend this file, note why, tell the
  operator.

## Done means

TASK-145, TASK-144, TASK-146 each Done on the board via its own merged PR; the new
wiki-overlap BLOCK live in `scripts/check-merge-drift.mjs` and documented in CLAUDE.md +
constitution; wiki freshness gate green in v2 mode with CAPSULES.md present; player-docs
13/13 fresh; `go test ./...` green on main; no sweep worktrees in `git worktree list`;
this file's execution log complete and status flipped to done. TASK-146 alone may be
dropped (tail lane) — if dropped, that is recorded here, not rounded up to done.

## Execution log

| date | task | PR | merge | notes |
|------|------|----|-------|-------|
| 2026-07-26 | TASK-145 | #106 | merge commit (origin 922223b lineage) | spec 069; gate live: wiki-repin-missing + player-docs blocks; constitution v1.2.0; first PR under review of its own doctrine |
| 2026-07-26 | TASK-144 | #107 | merge commit | spec 070; Close joins workers; first live block of the 069 gate — 3 findings satisfied in-branch |
