# UI sweep runbook — orchestrator instructions (reorientation 2026-07-25)

**Status: awaiting operator sign-off** (original had no status line; snapshot refreshed
2026-07-25 ~17:30 by a third orchestrator session — see the refreshed snapshot and
execution log below; flip this line to `signed-off — executing` once the operator
approves).

**You (the session reading this) are the ORCHESTRATOR** for the remaining UI-sweep tasks
from the 2026-07-25 reorientation. Your job: run each queued task through this project's
full PDLC automatically — spec → link → worktree → delegated implementation → PR → merge →
re-ground — parallelizing where the lanes below allow, and treating merge conflicts as
routine, not exceptional. Do not re-litigate direction: the synthesis
(`docs/design/reorient-2026-07-25-ui.md`, decisions 1–8 + D1–D13) and spec 047's authored
design pages are decided. Plan-of-record is the board.

## Read first (in this order)

1. `docs/design/reorient-2026-07-25-ui.md` — decisions + waves (the why).
2. `docs/design/tui/INDEX.md` — spec 047's authority statement, taxonomy, and **gate
   rules**; then skim `docs/design/tui/anatomy.md` and the `status: specified` pages for
   whichever task you're about to run (each new surface already has an authored page).
3. `backlog task list --plain` — live state; other sessions move it while you work.
4. The board task you're executing (`backlog task view TASK-<n> --plain`) — its rescope
   notes carry the decision references.

## State when this runbook was written (2026-07-25 ~15:00)

- **Done:** TASK-123 (spec 047, PR #80, merge 948c162) — the design reference v2 exists;
  all ten new-surface pages are authored at `status: specified`.
- **In flight elsewhere (do not duplicate, expect their merges):** TASK-130 (player-docs
  keys-reference, PR #81 open), spec 048 / TASK-107 (tuning manifest), TASK-105 (context
  grounding, In Progress). Other orchestrators like you may also be running.
- **Queued (this runbook's scope), in execution order below:** TASK-121, 124, 126, 125,
  119, 117, 127, 115, 128, 129; TASK-67 optional tail.

## Snapshot refresh (2026-07-25 ~17:30)

- **Done since authoring:** TASK-124 (spec 049, PR #84, merge c388c41), TASK-126
  (spec 050, PR #83, merge f18e9a4), TASK-130 (player-docs keys-reference, PR #81,
  merge ef696aa). Lane 1's two Sonnet slices are merged and re-grounded.
- **In flight (do not duplicate):** TASK-121 (spec 052 fully authored on main incl.
  `contracts/skin-contract.md`; worktree `.worktrees/task-121` carries US1 + T005
  commits), TASK-125 (linked to spec 053, In Progress, Sonnet + card-seam scope ruling
  recorded, worktree cut at 622e559). Outside this runbook: TASK-131 (spec 051
  merge-drift gates, worktree present), TASK-107 (uncommitted wiki edits at root —
  leave them).
- **Ambiguous claim — TASK-119:** `specs/054-scenario-machinery/` exists with spec.md +
  requirements checklist only (no plan/tasks/link; board still To Do). Either a stalled
  pre-work artifact or a session mid-specify. Confirm before claiming; if claimed,
  resume from spec 054 rather than renumbering.
- **Unblocked and unclaimed:** TASK-117 (Lane 3 gate satisfied — 121's skin-token
  contract is published on main via spec 052), TASK-127 (its two parked operator
  questions are RESOLVED in the spec-047 pages, which win per this runbook: ambient
  postmortem = morgue evidence only, `overlays/postmortem.md` FR-018 ruling; ceremony
  interrupt policy = decision 6 stands, reopening signal named as a watch item,
  `overlays/ceremony.md` "Interrupt-policy watch item" — verify score voice against the
  page at dispatch). Still blocked: TASK-115 (needs 125's console surface), TASK-128
  (Lane 4, after 125/117/119), TASK-129 (Lane 5 tail).

## Execution log

| Task | Spec | PR | Merge | Date | Notes |
|---|---|---|---|---|---|
| TASK-124 | 049 | #84 | c388c41 | 2026-07-25 | Lane 1; jump-to-source + parity retrofit |
| TASK-126 | 050 | #83 | f18e9a4 | 2026-07-25 | Lane 1; guardian strip |
| TASK-130 | — | #81 | ef696aa | 2026-07-25 | adjacent player-docs refresh (not in queue) |
| TASK-121 | 052 | — | — | in flight | Lane 1; contract published on main |
| TASK-125 | 053 | — | — | in flight | Lane 2; dispatched by sibling session |

## Execution lanes (dependency-ordered; parallelize within a lane)

Rule of thumb: DEVELOP in parallel, MERGE serially — almost every task below touches
`internal/tui/` and `docs/design/tui/`, so concurrent PRs will conflict; the lanes bound
how bad it gets.

**Lane 1 — start immediately, in parallel (3 worktrees):**
- **TASK-121 (Opus)** — *spec phase is the blocker, not the sweep.* Produce the skin-token
  CONTRACT first (spec + `patterns/skin-tokens.md` alignment); the code/docs rename sweep
  is the same task's implementation but 117/115 only need the published contract. Full
  Spec Kit; injection-adjacent (persona voice composed beneath the fixed frame) → Opus.
- **TASK-124 (Sonnet)** — chronicle `⏎`/click jump-to-source + parity doctrine in
  `patterns/keymap.md`. Small; merge FIRST.
- **TASK-126 (Sonnet)** — guardian strip. Small; merge second; rebase over 124.

**Lane 2 — after Lane 1's TUI merges (2 worktrees in parallel):**
- **TASK-125 (Sonnet, escalate to Opus if gates fail)** — guardian console page +
  systems-tab split. The big TUI slice; expect conflicts with everything; rebase daily.
- **TASK-119 (Opus)** — scenario machinery. Mostly `internal/sim`/daemon (deterministic
  scheduled emissions — executor-emission doctrine) + the exercise panel; low TUI overlap
  with 125 except views wiring. Opus: sim-loop/determinism.

**Lane 3 — after 121's contract (and 125 for TASK-115):**
- **TASK-117 (Sonnet)** — lesson row; consumes skin tokens; client-side projection
  precedent (decisions.go).
- **TASK-127 (Sonnet)** — ceremony + postmortem overlays; `run.ended` and
  `curriculum.stage_unlocked` already exist (fixtures until 119's emitter lands).
  NOTE: `overlays/ceremony.md` / `overlays/postmortem.md` were authored by spec 047 — if
  the parked open questions (ambient postmortem contents; ceremony score voice) are still
  marked open in those pages, ASK THE OPERATOR before implementing; if the pages resolve
  them, the pages win.
- **TASK-115 (Opus)** — explain tool + tutor + report card; metatron turn pipeline =
  injection-adjacent; needs 121's contract and 125's console surface for the card.

**Lane 4 — after the tabs/rows it governs exist (125, 117, 119):**
- **TASK-128 (Sonnet, watch for escalation)** — stage-shaped layout defaults; touches
  every mode's layout; keep it last of the chrome work.

**Lane 5 — tail:**
- **TASK-129 (Sonnet)** — villager strip + map overlays.
- **TASK-67 (Opus, optional)** — fork duels; only if the operator asks or the queue is dry.

Record the model tier + rubric justification on each board task when dispatching
(constitution Principle V; escalation is one-way Sonnet → Opus).

## The per-task SDLC loop (repeat verbatim for every task)

1. **Root freshness:** at repo root (pinned to main): `git fetch origin && git pull --ff-only`.
2. **Spec:** full Spec Kit unless the task is genuinely trivial-exempt (almost none of
   these are): `speckit-specify` → `speckit-clarify` (only if ambiguous) → `speckit-plan`
   → `speckit-tasks`. **Spec-number collision check:** `git fetch` and list
   `origin/main:specs/` before claiming the next NNN — 048 is taken; concurrent sessions
   take numbers constantly; renumber on conflict.
3. **Link:** `spec-bridge:link` the spec to the board task BEFORE implementation.
4. **Worktree:** `git worktree add .worktrees/task-<N> -b task-<N>-<slug> origin/main`.
   Never check out branches at root. One task, one worktree, one PR.
5. **Dispatch implementation** to the `spec-implementer` agent (model per lane above) —
   the orchestrator plans and gates, never implements inline. Subtasks = commits on the
   task branch, never their own PRs.
6. **The spec-047 gate (every PR touching `internal/tui/`):** run
   `node scripts/check-tui-design.mjs --changed` in the worktree; amend the affected
   `docs/design/tui/` pages in the SAME PR (flip the surface's page
   `status: specified → shipped`, fill real renderer symbols, re-pin `verified_against`).
   A TUI PR without its doc amendment is not mergeable.
7. **Pre-PR:** rebase onto fresh `origin/main`; re-run `go test -race ./...` and the check
   script AFTER the rebase (tripwire tests get updated by concurrent sessions); then open
   the PR from the worktree (`gh pr create`), body ending with the standard generated-with
   footer.
8. **Merge serially:** before merging, re-check the branch is based on current
   `origin/main` (rebase again if a sibling merged meanwhile). After merge: verify with
   `gh api repos/{owner}/{repo}/pulls/<n> --jq .merged` BEFORE deleting anything; then
   `git worktree remove .worktrees/task-<N>`, `git branch -d`, ff-pull root. NEVER
   delete+recreate a closed PR's head branch — open a fresh PR instead.
9. **Re-ground (the merge is not the end):** `spec-bridge:sync`; tick tasks.md at root
   (spec docs commit directly to main); `/grounding-wiki:wiki-update` when the merge
   touched files any wiki note lists as sources (`docs/wiki/tui-client.md` will need
   re-pinning after EVERY one of these tasks); then
   `node .claude/skills/player-docs/scripts/check-freshness.mjs --check` and run the
   `player-docs` skill if stale; mark the board task Done with `--final-summary`.
10. **Board hygiene:** `git add` specific task files only, never `backlog/` wholesale;
    run backlog/spec-bridge/tasks.md commands from repo root, never inside `.worktrees/`.

## Concurrency & conflict doctrine (this WILL happen)

- **Hotspots:** `internal/tui/{tui,views,layout,help,digest,decisions}.go`,
  `docs/design/tui/**`, `docs/wiki/tui-client.md`, `docs/player/**`, `backlog/tasks/*`,
  `README.md`, `specs/` numbering, `.specify/feature.json` (per-worktree; don't fight it).
- On conflict: rebase (never merge-commit into the task branch), take main's side for
  anything you didn't intentionally change, re-run the check script + race tests, and
  re-verify your design-doc amendments still describe the post-rebase reality.
- Two TUI-heavy PRs must not merge within the same re-ground cycle without a rebase
  between them — the wiki/player-docs freshness gates will otherwise thrash.
- If another session's PR conflicts with your open PR, prefer letting the SMALLER one
  merge first, regardless of whose it is.
- Concurrent sessions rebase main frequently: before diagnosing "my branch broke," fetch
  and compare against `origin/main` first.

## Operator checkpoints (do not proceed silently)

- The parked open questions from the synthesis, IF still open in the spec-047 pages:
  ambient postmortem contents (→ TASK-127), ceremony score voice (→ TASK-127),
  live-vs-end rubric gauges (→ TASK-119). Ask as concrete either/or decisions.
- Any one-way door not covered by decisions 1–8/D1–D13.
- Escalating a slice from Sonnet to Opus (record the rubric justification on the task).

## Done means

All of TASK-124/125/126/127/128/129/117/119/115/121 are Done on the board, each via its
own merged PR; every affected spec-047 page flipped to `status: shipped` with a fresh pin;
`check-tui-design.mjs` green on main; wiki notes and `docs/player/` fresh on main; no
stale worktrees under `.worktrees/`.
