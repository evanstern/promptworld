<!-- pdlc:grounding BEGIN v0.8.0 — planted by pdlc:bootstrap; refreshed wholesale on update. Keep project-specific edits OUTSIDE this block. -->
# promptworld — praxis development lifecycle (PDLC)

This project is developed with the **praxisflux** plugin suite. This block is the always-on
grounding: it names the loop, each plugin's role, and the rules that hold between them. The
procedures live in the plugins' skills (lazy-loaded); this block makes the rules apply even
when no skill has triggered.

## The loop

Ground the codebase → plan as specs → build + re-ground in the branch → PR →
merge → teach/render:

```
grounding-wiki (docs/wiki) ──corpus──▶ codebase-to-course (docs/course)
        │
        └─grounding─▶ spec/plan ─▶ build ─▶ wiki-update + player-docs ─▶ PR ─▶ merge ─▶ …
                                 └──────── one task branch ──────────┘
```

Re-grounding happens INSIDE the task branch, before the PR (spec 069): the
wiki re-pin and regenerated player docs ride the same PR as the code they
verify.

## Plugin roles (entry skills)

- **grounding-wiki** — the code-grounded corpus in `docs/wiki/`: per-concept notes pinned to
  the commit they were verified against. Build once with `/grounding-wiki:wiki-build`; when a
  task branch touches files any note lists as sources, run `/grounding-wiki:wiki-update` in
  that branch's worktree BEFORE opening the PR — the re-pin rides the PR (spec 069).
- **codebase-to-course** — interactive single-page HTML course in `docs/course/`, for
  non-technical readers. Reads `docs/wiki/` as its primary input when present.
- **build** — implements a SPEC handed off through `.handoff/` (`/build:implement`) and
  returns findings to the producer.
- **research** — drop-anywhere cited-fact vaults (`research:research-vault` → `analyze-vault`
  → `vault-artifact`) for grounding external topics.
- **spec-bridge** — the kanban view over Spec Kit specs (see the Spec Kit block below, if
  opted in).

## Rules that always hold

- **Artifact-grounded action:** never do anything without leaving a durable paper trail
  and/or gating against real physical evidence in the project — a file, a git commit, a
  task/issue. Artifacts that survive for human review are the only currency of state and
  decision: a choice living only in a chat turn, or a commitment left as prose where its
  durable home is the tracker, did not happen. Decisions are derived FROM artifacts and
  produce NEW artifacts; a question an existing artifact or principle already answers is
  resolved from it, not re-asked as a preference.
- **One TASK, one PR:** a TASK is a top-level deliverable and maps 1:1 to a pull request —
  one task, one branch, one PR. A SUBTASK (whatever the task system calls it) is internal
  work breakdown and never gets its own PR: subtasks land as commits on the parent TASK's
  single branch and merge together in that TASK's one PR.
- **Gates:** a status can never exceed the artifacts that prove it. Plugins ship Stop hooks
  that enforce this; when a gate blocks, produce the missing artifact — don't argue with the
  gate or edit derived state by hand.
- **Handoffs:** plugins compose only through files + gates, never by calling each other.
  Payloads ride the gitignored `.handoff/` transport; evidence lives in tracked state.
- **Grounding freshness:** `docs/wiki/` is load-bearing, not decoration. A branch that
  touches pinned sources must carry its own re-pin: run `/grounding-wiki:wiki-update` in the
  worktree and commit it on the branch BEFORE the PR — the pr gate BLOCKS
  (`wiki-repin-missing`, spec 069) until the branch itself re-verifies every note whose
  sources it touched. Re-pinning is pre-PR work, never a post-merge tail on main.
- **Player docs:** `docs/player/` (plain-language HTML for non-engineers) is generated from
  the wiki + README.md + docs/llm-providers.md by the `player-docs` project skill. When a
  branch changes `docs/wiki/`, regenerate the pages IN THE SAME BRANCH — the pr gate runs
  `node .claude/skills/player-docs/scripts/check-freshness.mjs --check` and BLOCKS on
  staleness (`player-docs-stale`, spec 069); stale pages cannot ride through a merge.

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
  `docs/wiki/` without regenerating `docs/player/`. There is no bypass flag;
  emergencies go through the operator editing hook config, visibly and deliberately.
- **Merge-commit-only:** merge PRs with `gh pr merge --merge`. In-branch re-pins are
  branch commit hashes; a squash merge rewrites them out of main's history and stales
  every pin the PR carried (observed on this repo — the squash-rewrites-pins hazard).
- **Step 7 is derived state only:** after the merge, the ONLY sanctioned main commits
  are bookkeeping DERIVED from the merged artifacts — board card moves, spec-bridge
  sync, tasks.md ticks, runbook execution logs — per spec 065's claim protocol and
  the pdlc:sweep re-ground step. Grounding content (wiki notes, player docs, design
  references) always rides the PR, never a post-merge commit.

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

- **The claim commit:** the FIRST commit of any task claims the work — it moves the
  board card to In Progress AND creates the spec directory (`specs/NNN-<slug>/`; a stub
  claims the number), before any spec authoring or code. Push it immediately. Never
  force-push.
- **Rejected push = stop-the-lane signal:** fetch, re-read the board and `specs/`, and
  if another session now holds that task or that number, STOP the lane and surface the
  collision to the operator — do not rebase and continue. If the rejection was
  unrelated (someone pushed board notes) and the task and number are still free,
  rebase and re-push the claim.
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

## Git worktrees — root stays on main, branches live in `.worktrees/`

The root checkout is **pinned to `main`** — never check out a task branch there. All
branch work happens in worktrees under the repo-local, gitignored `.worktrees/` folder.

- **Create:** when starting a TASK branch, make a worktree instead of switching:
  `git worktree add .worktrees/task-<N> -b task-<N>-<slug> origin/main`
  (dir named `task-<N>`, matching the task id). Do the work, commit, and open the PR
  from inside `.worktrees/task-<N>`.
- **Root freshness:** keep the root current with `git fetch origin && git pull --ff-only`
  — at session start and always before cutting a new worktree, so branches fork from
  fresh `origin/main`.
- **Cleanup:** after a TASK's PR merges, `git worktree remove .worktrees/task-<N>`,
  delete the branch (`git branch -d …`), and ff-pull the root.
- One TASK, one worktree — this is the same "one task, one branch, one PR" rule; the
  worktree is just where that branch lives.

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
