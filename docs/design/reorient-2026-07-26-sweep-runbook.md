# Reorient 2026-07-26 board moves — sweep runbook (2026-07-26)

**You (the session reading this) are the ORCHESTRATOR** for the tasks below. Run each
through the host project's full PDLC — spec → link → worktree → delegated implementation →
PR → merge → re-ground — parallelizing within lanes, merging serially, treating merge
conflicts as routine. Direction is decided; do not re-litigate it:
`docs/design/reorient-2026-07-26-ui.md` (the merged synthesis, PR #110) wins.
Plan-of-record is the board; this file carries only ordering, doctrine, and the log.

**Status:** signed-off · operator sign-off on lanes: 2026-07-26 (lanes specified verbatim
by the operator in the `/pdlc:sweep` invocation — Waves A–D with dependencies and the
TASK-111/136/137 pause — which this runbook transcribes rather than re-derives).

## Read first (in this order)

1. `docs/design/reorient-2026-07-26-ui.md` — the synthesis that produced these tasks
   (decisions 1–8, board-moves table, open questions).
2. `docs/design/tui/INDEX.md` + the spec-047 gate (`scripts/check-tui-design.mjs`);
   the spec-051/065/069 gate `scripts/check-merge-drift.mjs`.
3. `backlog task list --plain` — live state; other sessions move it while you work.
4. The task you're about to execute (`backlog task view TASK-<n> --plain`).

## State when this runbook was written (2026-07-26, origin/main 93cb578)

- **Done already:** TASK-144/145/147 (the synthesis's "board allows a clean sweep run"
  gate — all verified Done). Specs 001–071 taken; **next free spec number is 072**.
- **In flight in other sessions — PAUSED by the operator, do not touch:** TASK-111,
  TASK-136, TASK-137 (cards, branches, worktrees all off-limits; their In Progress
  status is NOT a blocking finding for this sweep).
- **Queued (this runbook's scope):** TASK-149 → TASK-150 (Lane A); TASK-154 → TASK-142
  (Lane B, parallel with A); TASK-67 (Lane C, after TASK-149); TASK-151 + TASK-152 +
  TASK-153 (Lane D, tail).
- No worktrees besides root; root clean on main at 93cb578.

## Execution lanes (dependency-ordered; parallelize within a lane)

Rule of thumb: DEVELOP in parallel, MERGE serially — tasks below share file footprints,
so concurrent PRs will conflict; the lanes bound how bad it gets.

**Lane A — honesty (start immediately; strictly serial, shared files):**
- **TASK-149 (Opus 4.8 — cross-package sim state/reducer + every TUI card surface;
  doctrine-adjacent: grading truth at the teaching moment; persists new state)** —
  unify postmortem/ceremony/console cards on `sim.EvaluateRubric`; author `the-law`'s
  evaluator (persist charter `Default` into state, `internal/sim/scenario.go:277`).
  HIGH. Its merged contract unblocks Lanes C and D's TASK-151.
- **TASK-150 (Sonnet — single-script lint extension + design-doc cell fixes)** —
  semantic lint in `check-tui-design.mjs`; fix `overlays/postmortem.md` ×7 and
  `panels/exercise.md:110` same-PR. Serial after TASK-149: both amend the same two
  design pages.

**Lane B — UI lane (parallel with Lane A; within the lane, 154 merges first):**
- **TASK-154 (Sonnet — tooling/test, single surface)** — mouse-parity sweep test over
  the control tables' `keys+mouse` column; burn down `patterns/keymap.md`'s rollout
  note. Small; merges before TASK-142 so the new gate covers 142's mode table.
- **TASK-142 (Sonnet — view/rendering feature in `internal/tui`; AC7's light level may
  need a small read-only sim derivation, which stays inside the rubric's routine tier;
  escalate to Opus only on failed gates, as an operator checkpoint)** — look-cursor
  mode with TILE pane, DF hierarchy, tile-registry whatis, mouse parity. The sweep's
  biggest slice; nothing else fights for its input-handling files once 154 lands.

**Lane C — iteration (after TASK-149 merges):**
- **TASK-67 (Opus 4.8 — cross-package architectural: world-lifecycle fork, fresh
  identity, lineage events, determinism harness, compare surface)** — `promptworld
  fork` + rubric-first duel scoreboard sharing `reportCardView` + `sim.EvaluateRubric`.
  HIGH. v1 only; HTML retelling is phase 2 (not in this sweep); dual-TUI deferred.

**Lane D — content tail (after TASK-149 for 151; 152/153 independent; droppable last):**
- **TASK-151 (Opus 4.8 — new reducer-valid event kinds across sim reducer + chronicle
  digest (TestCatalogSweep) + scenario evaluators; replay-safety doctrine-adjacent)** —
  2–3 exercises per stage, ~3 incident kinds, lesson tranche 2.
- **TASK-152 (Sonnet — single-package deterministic TUI view)** — forward-ladder block
  in the `?` guardian section; `overlays/help.md` byte-identity row same-PR.
- **TASK-153 (Sonnet — content-only, player-docs skill is the home)** — quickstart
  first-prompt pass. Runs full Spec Kit unless the constitution's trivial exemption
  provably holds at spec time (content pass ≠ surgical fix, so default is full rigor).

Record the model tier + rubric justification on each board task at dispatch
(one-way escalation only; escalations are operator checkpoints).

## Per-PR gates this project enforces (enumerated — implementers cannot miss these)

- **Merge-drift gate: present at `scripts/check-merge-drift.mjs`.** Mandatory at every
  choke point, hook-enforced: `node scripts/check-merge-drift.mjs session` at each
  task start (janitor + drift matrix); `node scripts/check-merge-drift.mjs claim --dir
  NNN-slug` before creating any `specs/NNN-*` dir; `node scripts/check-merge-drift.mjs
  worktree --spec NNN --task TASK-<n>` before every `git worktree add`; `node
  scripts/check-merge-drift.mjs pr` from the worktree before every `gh pr create` AND
  after every rebase — nonzero exit blocks, no bypass flag.
- **TUI authority gate (spec 047):** any PR touching `internal/tui/` runs
  `node scripts/check-tui-design.mjs --changed` and amends `docs/design/tui/` in the
  same PR (re-verify + re-pin every affected page). Applies to 149, 150, 142, 154, 152,
  and 67 (its scoreboard shares `reportCardView`).
- **Wiki-in-PR lifecycle (spec 069):** the branch re-pins every wiki note whose pinned
  sources it touches (`wiki-repin-missing` blocks); wiki changes regenerate
  `docs/player/` in the same PR (`player-docs-stale` blocks; probe:
  `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`).
- **Tests:** `go test ./...` green in the worktree, re-run after every rebase. New
  event types must satisfy `TestCatalogSweep` (chronicle digest coverage) — bites
  TASK-151 especially.
- **Merge-commit-only:** `gh pr merge --merge`, never squash (squash rewrites branch
  pins — observed hazard on this repo).
- **Spec rigor:** full Spec Kit + `spec-bridge:link` BEFORE implementation for every
  task (operator: all non-trivial, except possibly TASK-153 — see Lane D).
- **Claim-before-work (spec 065):** first commit moves the card to In Progress AND
  creates the spec-dir stub, pushed immediately; rejected push = stop-the-lane signal.
  Specs/board bookkeeping commit to main at root; the task branch carries code only.

## Concurrency & conflict doctrine

- **Hotspots:** `internal/tui/views.go` (149, 142, 152, 67); `internal/tui/tui.go`
  input dispatch (142, 154, 152); `docs/design/tui/overlays/postmortem.md` (149, 150);
  `docs/design/tui/patterns/keymap.md` (142, 154); `docs/design/tui/overlays/help.md`
  (142's badge deep-link, 152's byte-identity row); `internal/sim/scenario.go` (149,
  151); the wiki `tui-*` notes and the regenerated `docs/player/` pages (nearly every
  PR — regenerate after rebase, not before); `backlog/tasks/*` (scoped adds only).
- Rebase, never merge-commit into a task branch; take main's side for anything you
  didn't deliberately change; re-run gates after every rebase.
- Two hotspot-heavy PRs never merge within one re-ground cycle without a rebase between.
- Conflicting with a sibling session's open PR → the smaller PR merges first.
- **Claim before work** as enumerated above; never force-push a claim.
- **A rejected push means you lost the race:** fetch, re-read the board and `specs/`.
  If another session now holds that task or number, STOP the lane and surface it to
  the operator. Unrelated rejection with the task+number still free → rebase and
  re-push the claim.
- Verify a PR is merged (`gh api repos/{owner}/{repo}/pulls/<n> --jq .merged`) before
  deleting its branch/worktree; never delete+recreate a closed PR's head.
- The paused TASK-111/136/137 lanes keep their remote branches; the session gate may
  report them — that is expected, not a finding.

## Operator checkpoints (do not proceed silently)

- **TASK-142 spec time — reverse-jump home** (synthesis open question 4: rider on 142,
  rider on 154, or own card). Default taken by this sweep: leave it UNSCHEDULED (it is
  the delta's net-new unscheduled rec); surface the recommendation in the final report
  rather than folding it into 142's scope. Folding it in would be a silent scope grow.
- **TASK-142 spec time — pull-surface budget** (open question 3): record the tension in
  the spec's notes; no navigation ruling is taken in this sweep.
- **TASK-67 spec — fork budget-meter semantics** RESOLVED at spec time, deviating from
  this runbook's original recommendation with evidence (spec 076 research R4): the spend
  meter is per-world (world meta table + per-world llm.json) — no machine-global wallet
  exists to share. Encoded: the fork INHERITS the wallet (llm.json + llm_spend_* meta
  copied verbatim; forking never mints fresh budget; independent meters thereafter).
  Surfaced to the operator in the sweep report.
- Tier escalations (one-way Sonnet → Opus; record rubric justification on the task).
- Lane amendments (amend this file, note why, tell the operator).

## Done means

All eight tasks (149, 150, 154, 142, 67, 151, 152, 153) Done on the board, each via its
own merged PR (merge commits, on main); `go test ./...`, `check-tui-design.mjs`,
player-docs freshness, and `check-merge-drift.mjs session` all green on main; wiki pins
current (re-pinned in-branch per spec 069); `git worktree list` shows only root; this
file's execution log complete and status flipped to done. TASK-111/136/137 untouched.

## Execution log

| date | task | PR | merge | notes |
|------|------|----|-------|-------|
| 2026-07-26 | TASK-152 | #118 | fb8c865 | Ladder view live; StageEarned substrate relocated (parity by construction); reconciled in-branch with #116/#117 |
| 2026-07-26 | TASK-67 | #116 | 4daf75c | Lane C complete; fork+compare shipped; wallet-inheritance decision ratified in-spec (R4); surfaced pre-existing wiki budget debt from PR #115 (carded) |
| 2026-07-26 | TASK-142 | #115 | 011ce4e | Lane B complete; look-cursor + TILE pane + mouse parity + badge deep-link; merged with 5-commit docs-only base lag (zero overlap, drift gate green) to preserve 26 in-branch pins |
| 2026-07-26 | TASK-150 | #114 | 495d8cb | Lane A complete; red-run proved 8 cells (7 postmortem + 1 help.md); help.md badge cell retagged pending TASK-142 |
| 2026-07-26 | TASK-149 | #113 | f78358a | Lane A anchor merged; unblocks 150/67/151. Doctrine tension logged: implementer merged origin/main into the branch (not rebase) to preserve 43 in-branch pin hashes — pr gate green; operator may want to ratify merge-over-rebase for pin-carrying branches |
| 2026-07-26 | TASK-154 | #112 | 86b776d | Lane B first merge; mutation check proven; player-docs pin gap found+fixed in-branch (freshness probe must run directly, not just via pr gate) |
