# UI sweep runbook — orchestrator instructions (reorientation 2026-07-25)

**Status: signed-off — executing** (operator sign-off 2026-07-25 ~17:35: third
orchestrator session claims TASK-117 → TASK-127, serial; TASK-119 stays unclaimed until
its half-authored spec 054 is attributed; snapshot refreshed below).
**Amendment 2026-07-25 ~18:25**: TASK-127 was claimed by a sibling session at ~18:18
(linked to spec 056-takeover-surfaces, Sonnet tier + checkpoint ruling recorded,
dispatch gated on TASK-121's merge) — session 3's claim reduces to TASK-117 only, per
the no-duplication doctrine. Operator notified in-session.

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

## State — QUERY IT, DO NOT READ IT HERE

*(Two hand-maintained snapshots lived here — "State when this runbook was written ~15:00" and
"Snapshot refresh ~17:30". Both were deleted 2026-07-25 after the team review found them
false: they listed TASK-121 and TASK-125 as in flight after 125 merged (PR #87), called
TASK-117 "unblocked and unclaimed" after it was Done (PR #88), and cited worktree paths that
do not exist on every machine. The execution log below and the board are the real record.)*

Run these at every dispatch — seconds, and they cannot be stale:

```sh
backlog task list --plain                     # live board state
git fetch origin && git branch -r             # what actually has a branch
ls -d specs/*/ | sed 's|specs/0*||;s|-.*||' | sort -n | tail -1   # highest spec number
```

- **Queued (this runbook's scope), execution order:** TASK-121, 124, 126, 125, 119, 117,
  127, 115, 128, 129; TASK-67 optional tail. Consult the execution log below for what has
  actually landed — do not infer it from prose in this file.
- **Standing rulings that are NOT state and remain valid:** TASK-127's two parked operator
  questions are RESOLVED in the spec-047 pages, which win per this runbook (ambient
  postmortem = morgue evidence only, `overlays/postmortem.md` FR-018; ceremony interrupt
  policy = decision 6 stands, with the reopening signal named as a watch item in
  `overlays/ceremony.md`) — verify score voice against the page at dispatch.
- **Amendment 2026-07-25 (operator, team review):** the decided merge order TASK-121 → TASK-111
  was overtaken — TASK-111 merged first (PR #90). TASK-121's sweep must now rebase through
  111's survival-turn code in `internal/metatron/{turn,orders}.go` and add any Metatron-voice
  order text / soul header it landed to the sweep inventory, rather than assuming the
  skin-token binding held. TASK-115's spec renumbered off the 059 collision to
  `specs/063-grounded-feedback`.

## Execution log

| Task | Spec | PR | Merge | Date | Notes |
|---|---|---|---|---|---|
| TASK-124 | 049 | #84 | c388c41 | 2026-07-25 | Lane 1; jump-to-source + parity retrofit |
| TASK-126 | 050 | #83 | f18e9a4 | 2026-07-25 | Lane 1; guardian strip |
| TASK-130 | — | #81 | ef696aa | 2026-07-25 | adjacent player-docs refresh (not in queue) |
| TASK-121 | 052 | #94 | 70acb2e | 2026-07-25 | Lane 1; skin contract + sweep merged; AC #6 re-ground closed by MVLS player-docs 998ea90 + wiki 10b3247 |
| TASK-125 | 053 | #87 | — | 2026-07-25 | Lane 2; guardian console + systems split |
| TASK-117 | 055 | #88 | dfa73d7 | 2026-07-25 | Lane 3; session 3 — lesson row shipped (Sonnet), 18/18 tasks, board Done |
| TASK-127 | 056 | #99 | ded11c2 | 2026-07-25 | Lane 3; takeovers + reportCardView seam (Sonnet); 2-round rebase over 119 |
| TASK-140 | — | #98 | 87f7251 | 2026-07-25 | hotfix (trivial-exempt): main-red TestCatalogSweep — recovery_stalled catalog row (Sonnet) |
| TASK-119 | 054 | #101 | d220645 | 2026-07-25 | Lane 2; scenario machinery + rubric emitter (Opus); closed TASK-135 as production-wired |
| TASK-115 | 063 | #100 | bdb0686 | 2026-07-25 | Lane 3; explain/tutor/report card (Opus); checklist-above-note seam plugged into 127's renderer |
| TASK-128 | 066 | #102 | 24ae434 | 2026-07-25 | Lane 4; stage defaults + authority-page parity sweep (Sonnet) |
| TASK-129 | 060 | #103 | 7e3c2b5 | 2026-07-25 | Lane 5; villager strip + map overlays (Sonnet). SWEEP COMPLETE — all 10 queued tasks Done; TASK-67 remains optional, not started |

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
