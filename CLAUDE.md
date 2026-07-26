<!-- pdlc:grounding BEGIN v0.21.0 — planted by pdlc:bootstrap; refreshed wholesale on update. Keep project-specific edits OUTSIDE this block. -->
# promptworld — praxis development lifecycle (PDLC)

This project is developed with the **praxisflux** plugin suite. This block is the always-on
grounding: it names the loop, each plugin's role, and the rules that hold between them. The
procedures live in the plugins' skills (lazy-loaded); this block makes the rules apply even
when no skill has triggered.

## The loop

Ground the codebase → plan as specs → build → re-ground → teach/render:

```
grounding-wiki (docs/wiki) ──corpus──▶ codebase-to-course (docs/course)
        │
        └─grounding─▶ spec/plan ──▶ build ──▶ wiki-update (re-ground) ──▶ …
```

## Plugin roles (entry skills)

- **grounding-wiki** — the code-grounded corpus in `docs/wiki/`: per-concept notes pinned to
  the commit they were verified against. Build once with `/grounding-wiki:wiki-build`; after
  merging changes that touch files any note lists as sources, run `/grounding-wiki:wiki-update`.
- **codebase-to-course** — interactive single-page HTML course in `docs/course/`, for
  non-technical readers. Reads `docs/wiki/` as its primary input when present.
- **build** — implements a SPEC handed off through `.handoff/` (`/build:implement`) and
  returns findings to the producer.
- **research** — drop-anywhere cited-fact vaults (`research:research-vault` → `analyze-vault`
  → `vault-artifact`) for grounding external topics.
- **spec-bridge** — the kanban view over Spec Kit specs (see the Spec Kit block below, if
  opted in).
- **pdlc** — the lifecycle's own verbs: `pdlc:bootstrap` (re)stamps this grounding after
  plugin upgrades; `/pdlc:sweep` orchestrates a set of board tasks through the whole loop —
  an authored, operator-signed-off runbook, then spec → PR → merge → re-ground per task,
  parallel lanes with serial merges.

## Rules that always hold

- **Artifact-grounded action:** never do anything without leaving a durable paper trail
  and/or gating against real physical evidence in the project — a file, a git commit, a
  task/issue. Artifacts that survive for human review are the only currency of state and
  decision: a choice living only in a chat turn, or a commitment left as prose where its
  durable home is the tracker, did not happen. Decisions are derived FROM artifacts and
  produce NEW artifacts; a question an existing artifact or principle already answers is
  resolved from it, not re-asked as a preference.
- **One TASK, one PR:** a TASK is a top-level deliverable and maps 1:1 to a pull request —
  one task, one branch, one PR. An EPIC (whatever the task system calls it) groups
  deliverable TASKs and gets no PR of its own; a SUBTASK is internal work breakdown and
  never gets its own PR: subtasks land as commits on the parent TASK's single branch and
  merge together in that TASK's one PR. A PR exists only where it carries a stated reason
  for a human to approve (a policy ratified, a posture changed, a contract made binding) —
  never a diff for its own sake; work too small to give a reviewer a real decision merges
  into the deliverable it serves.
- **Gates:** a status can never exceed the artifacts that prove it. Plugins ship Stop hooks
  that enforce this; when a gate blocks, produce the missing artifact — don't argue with the
  gate or edit derived state by hand.
- **Handoffs:** plugins compose only through files + gates, never by calling each other.
  Payloads ride the gitignored `.handoff/` transport; evidence lives in tracked state.
- **Grounding freshness:** `docs/wiki/` is load-bearing, not decoration. Changes that touch
  pinned sources aren't done until the wiki is re-pinned (`/grounding-wiki:wiki-update`).
- **Corpus loading:** when a grounded corpus is present (`docs/wiki/` or similar), load its
  `INDEX.md` first and route; load notes just-in-time — never bulk-load the corpus.
  Whole-corpus orientation reads `CAPSULES.md` when it exists; without one, INDEX plus
  just-in-time notes.

<!-- pdlc:peer:backlog BEGIN -->
## Backlog.md — the board (officially supported peer)

Backlog.md is this project's kanban; the board is the plan of record. Statuses flow
**To Do → In Progress → Done**.

- Start from `backlog task list --plain`; read a task with `backlog task view TASK-x --plain`.
- Record plans (`--plan`), progress (`--append-notes`), and tick acceptance criteria
  (`--check-ac <n>`) as they come true; finish with `--final-summary` and `-s Done`.
- **One task, one PR:** a top-level TASK gets one branch and one PR. Dotted-id subtasks
  (TASK-x.y) are internal breakdown — they ride the parent task's branch and merge in its
  PR, never their own.
- **Never hand-edit** files under `backlog/` — always the `backlog` CLI, so metadata and
  relationships stay consistent.
<!-- pdlc:peer:backlog END -->

<!-- pdlc:peer:spec-kit BEGIN -->
## Spec Kit — specs drive the work (officially supported peer)

Features are specified with GitHub Spec Kit (`specify`) under `specs/NNN-<feature>/`
(spec.md, plan.md, tasks.md). The spec dir is the source of truth for its feature.

- Put a spec on the board with `spec-bridge:link`; after working a spec, run
  `spec-bridge:sync` to move the linked task, re-mirror phase criteria, and record progress.
- The bridge gate blocks a linked task's status from exceeding what the spec artifacts
  prove — produce the artifact, then sync.
- A spec's linked task is the deliverable: it lands as **one PR**. Spec phases and their
  mirrored criteria are internal breakdown, not PR boundaries.
<!-- pdlc:peer:spec-kit END -->
<!-- pdlc:grounding END -->


## TUI design reference — the UI authority gate (spec 047)

`docs/design/tui/` is the living page-by-page, control-by-control UI authority
(spec 047). Before opening any PR that touches `internal/tui/`, run
`node scripts/check-tui-design.mjs --changed` and amend `docs/design/tui/` in the
same PR (re-verify + re-pin every affected page).

## Merge-drift gates (spec 051)

`scripts/check-merge-drift.mjs` gates the parallel worktree SDLC against merge drift
at three choke points. No daemon, no CI, no PR comments — findings are exit codes plus
(optionally) board notes. Run:

- session start (root): `node scripts/check-merge-drift.mjs session`
- before cutting a worktree: `node scripts/check-merge-drift.mjs worktree [--spec NNN]`
- before opening a PR (from the worktree): `node scripts/check-merge-drift.mjs pr`

Exit 0 = pass (clean or warnings-only), 1 = blocked (do not proceed), 2 = usage/env
error. The script never rebases, merges, commits to, or resets any task branch or its
worktree — resolution always stays with the owning session. pr mode also enforces the
wiki-in-PR lifecycle (spec 069, next block): `wiki-repin-missing`, `player-docs-stale`,
and `player-docs-env-error` are blocking findings, with no bypass flag.

These gates are also hook-enforced at the harness level (`.claude/settings.json`,
`scripts/hooks/merge-drift-hook.mjs`): a PreToolUse hook blocks `gh pr create` and
`git worktree add` when the corresponding gate exits nonzero, and a SessionStart hook
injects the session gate's report as context without ever blocking session start.

## Wiki-in-PR lifecycle (spec 069)

A task's lifecycle is (1) design (2) code (3) approval (4) wiki grounding + player
docs, in-branch (5) PR (6) merge (7) close task + commit main. Grounding rides the
PR; the pr gate enforces it, and step 7 is bookkeeping only.

- **In-branch grounding (gated):** the pr gate blocks (`wiki-repin-missing`) unless
  the branch itself re-verifies every wiki note whose pinned sources it touches —
  note re-pinned on the branch, pin reachable from the branch tip, no source touched
  after the pin. It likewise blocks (`player-docs-stale`) when the branch changes
  `docs/wiki/` without regenerating `docs/player/` — the plain-language HTML pages
  generated from the wiki + `README.md` + `docs/llm-providers.md` by the
  `player-docs` project skill
  (`node .claude/skills/player-docs/scripts/check-freshness.mjs --check` is the
  gate's probe). There is no bypass flag;
  emergencies go through the operator editing hook config, visibly and deliberately.
- **Merge-commit-only:** merge PRs with `gh pr merge --merge`. In-branch re-pins are
  branch commit hashes; a squash merge rewrites them out of main's history and stales
  every pin the PR carried (observed on this repo — the squash-rewrites-pins hazard).
- **Step 7 is derived state only:** after the merge, the ONLY sanctioned main changes
  are bookkeeping DERIVED from the merged artifacts — board card moves, spec-bridge
  sync, tasks.md ticks, runbook execution logs — per spec 065's claim protocol and
  the pdlc:sweep re-ground step. Since TASK-160/TASK-161, board card moves commit
  directly at root under the board-sync exception (scoped to `backlog/`, pushed
  immediately), while all other derived state (spec-bridge sync under `specs/`,
  tasks.md ticks, runbook execution logs) is authored on a branch in a worktree
  and lands via merge — never committed directly at root. Grounding
  content (wiki notes, player docs, design references) always rides the PR, never a
  post-merge commit.

## Model-tiered workflow (constitution Principle V, v1.2.0)

Three tiers, enforced by delegation (`.specify/memory/constitution.md`, Principle V):

- **Fable 5 plans and gates:** specs (`speckit-specify`), clarify, plans (`speckit-plan`),
  task generation (`speckit-tasks`), analysis, board/task creation, and review/gating of
  implementer reports stay on the main session's planning model. Never implements inline.
- **Opus 4.8 implements the hard slices:** cross-package/architectural changes;
  concurrency/scheduling/governor logic (`internal/llm`, `internal/cognition`,
  `internal/mind` orchestration); doctrine-adjacent behavior changes; any slice whose
  prior Sonnet attempt failed gates or shipped live defects; adversarial verification
  passes on request. Select via the Agent tool's `model` param on the
  `spec-implementer` agent.
- **Sonnet implements the routine slices (default):** single-package features,
  view/rendering code, tests alongside code, doc reconciliation.

The escalation rubric lives in `.claude/agents/spec-implementer.md`; escalation is
one-way Sonnet → Opus, and the tier choice + rubric justification is recorded on the
board task.

**Spec rigor (constitution, Development Workflow):** every non-trivial TASK runs full
Spec Kit (specify → clarify where ambiguous → plan → tasks → implement) with the spec
linked to the board via `spec-bridge:link` BEFORE implementation starts. Trivial
exemption only when ALL hold: surgical fix, complete file:line diagnosis pinned on the
task, and ACs on the board task.

## Claim-before-work protocol (spec 065)

Git push rejection is the mutual-exclusion primitive for concurrent sessions: whoever
pushes first wins, and the loser's non-fast-forward rejection is a **signal**, never an
"annoying; rebase and carry on."

- **The claim commit:** claiming a task is TWO immediate pushes. (1) The card
  move: `backlog task edit TASK-x -s "In Progress"` at root, committed at root
  scoped to the card file alone and pushed to main at once (the board-sync
  exception, TASK-161) — THIS push is the mutual-exclusion event, and its
  rejection is the stop-the-lane signal. (2) The spec stub: the task branch's
  first commit creates `specs/NNN-<slug>/` (the stub claims the number),
  pushed with the branch and landed on main via an immediate manual
  `git merge --no-ff` at root. Never force-push.
- **Rejected push = stop-the-lane signal:** fetch, re-read the board and `specs/`, and
  if another session now holds that task or that number, STOP the lane and surface the
  collision to the operator — do not rebase and continue. If the rejection was
  unrelated (someone pushed board notes) and the task and number are still free,
  fetch and `git pull --no-rebase` (merge, never rebase) at root, then push again.
- **Task branches push on first commit:** `git push -u origin task-<N>-<slug>` as soon
  as the branch has a commit, so in-flight work is auditable from any clone. The
  session gate reports any local task branch with commits but no remote counterpart
  (`branch-unpushed`).
- **Gates (hook-enforced):** before creating any new `specs/NNN-*` directory, run
  `node scripts/check-merge-drift.mjs claim --dir NNN-slug` — it blocks (exit 1) when
  the number is taken on origin/main under a different name, and passes when you
  re-touch your own claimed dir. Cut worktrees with
  `node scripts/check-merge-drift.mjs worktree --spec NNN --task TASK-<n>` — it warns
  (`card-not-claimed`) when the task's card is not In Progress on origin/main, and with
  `--task` it accepts a spec dir already claimed BY that task.

## Root checkout is READ-ONLY — worktree + merge only, no rebases (iron-clad)

NOTHING is modified in the root checkout directly — no file edits, no direct
commits to main, no exceptions — code, docs, wiki, player docs, specs — with
ONE ratified carve-out: Backlog.md board state (see the board-sync bullet
below). Every change is authored on a branch in a
worktree under the repo-local, gitignored `.worktrees/` folder and reaches main
ONLY by merging that branch back: a PR (`gh pr merge --merge`) or a manual
`git merge --no-ff <branch>` run at root. **Rebases are forbidden everywhere in
this repo.** The only history operations are merges — main into a branch to
freshen it, branch into main to land it.

- **Create:** `git worktree add .worktrees/task-<N> -b task-<N>-<slug> origin/main`
  (dir named `task-<N>`, matching the task id); work, commit, and open PRs from
  inside the worktree.
- **Land:** PR merge, or — when the operator has ratified the change — a manual
  `git merge --no-ff` at root, then push. Never squash, never rebase.
- **Freshen a stale branch** by merging main INTO it; never rebase it (rebases
  and squash merges stale in-branch wiki pins — observed hazards on this repo).
- **Sanctioned git ops at root:** `fetch`, `pull --ff-only`, `merge` (plus the
  `git commit` that concludes a CONFLICTED merge, MERGE_HEAD present), `push`,
  `worktree add`/`remove`, `branch -d`, and reads (`status`/`log`/`diff`/…).
  Everything else that writes files or history at root is out.
- **The ONE exception — board sync (TASK-161):** `backlog/` is the plan of
  record and the concurrent-session mutex, so it lives current on MAIN. Edit it
  at root via the `backlog` CLI only, commit at root scoped to `backlog/` paths
  alone (`git add <specific task files> && git commit` — never `-a`, never mixed
  paths), and push immediately. Task branches never commit `backlog/` — the
  board has a single home. The root-guard hook allows exactly this shape of
  commit and no other. Non-board bookkeeping (tasks.md ticks, runbook execution
  logs, spec-bridge mirrors under `specs/`) is NOT excepted: it rides a branch
  and merges.
- **Root freshness:** `git fetch origin && git pull --ff-only` at session start
  and before cutting any worktree.
- **Cleanup:** after a branch lands, `git worktree remove .worktrees/task-<N>`,
  `git branch -d task-<N>-<slug>`, ff-pull the root.
- **Hook-enforced (TASK-160):** `scripts/hooks/root-guard-hook.mjs` (PreToolUse,
  wired in `.claude/settings.json`) blocks direct commits, cherry-picks,
  reverts, `git am`, squash merges, and branch creation at root; `git rebase`
  anywhere in the repo; force-pushes; and Write/Edit of non-gitignored files in
  the root checkout. No bypass flag — emergencies go through the operator
  editing hook config, visibly and deliberately.

## educate — Socratic learning layer (planted by educate:start, adapted for PDLC)

This project also hosts **educate** lessons (Socratic grounding/Q&A sessions) under
`topics/<topic-slug>/<NNN>-<lesson-slug>/`. Files inside a lesson use bare names:
`checklist.md`, `raw-notes.md`, and (when produced) `deck.html`, `guide.md`.

- **Lifecycle (exact words):** `scaffolded` → `taught` → `spec'd` → `built` → `decked` → `done`.
  Scaffold by copying `topics/.template/` at the start.
- **Note-taking cadence (enforced):** `raw-notes.md` is maintained live — one Session-log
  entry per question→answer exchange, written before the next question. A turn with no
  note is an incomplete turn.
- **Run lessons via the `educate:lesson` skill;** delegated builds go through the
  `.handoff/` transport to `build:implement` (see the PDLC block above — same rules).
- **Gate:** `topics/<topic>/progress.json` is the machine source of truth; sync/check with
  `node <educate-plugin>/scripts/progress.mjs --root <root> <topic> --sync|--check`.
  Never advance a lesson's status past the artifacts on disk.
- Decks are single self-contained HTML files built FROM `topics/.template/deck.html`.
