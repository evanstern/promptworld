# UI sweep — session handoff (2026-07-25, evening)

**You (the session reading this) are the ORCHESTRATOR continuing the UI sweep.**
The original instructions are `docs/design/reorient-2026-07-25-ui-runbook.md` — its
per-task SDLC loop, lanes, and conflict doctrine still govern; **this handoff
supersedes its "State when written" section.** The prior orchestrator session
("ui-refactor-work") ran Lanes 1–2 to completion and staged Lane 3+; it was halted
cleanly by the operator with everything committed. Direction remains decided — do
not re-plan it.

## Board state (verified against artifacts at handoff)

| Task | Spec | State | Evidence |
|---|---|---|---|
| TASK-124 chronicle jump-to-source | 049 | **Done** | PR #84, re-ground complete |
| TASK-126 guardian strip | 050 | **Done** | PR #83, re-ground complete |
| TASK-125 console + systems split | 053 | **Done** | PR #87, re-ground complete |
| TASK-117 lesson row | 055 | **Done** (other orchestrator's) | PR #88 |
| TASK-121 skinnable guardian | 052 | **In Progress — merged, one AC open** | PR #94 merged (70acb2e); wiki re-ground done (10b3247, notes renamed guardian.md + new skin.md); **AC #6's player-docs half NOT done** — freshness reports 6 stale + 1 broken-ref from the note renames. First action below. |
| TASK-119 scenario machinery | 054 | **In Progress — branch READY TO MERGE** | `.worktrees/task-119` @ `0ad218b`, 13 commits, fully rebased over PR #94's rename; full `-race` suite + design gate + merge-drift `pr` gate all green |
| TASK-127 takeover surfaces | 056 | In Progress — spec complete, linked; implementation NOT started | `.worktrees/task-127` pristine at origin/main |
| TASK-115 grounded feedback | **063** (renumbered off a 059 collision) | In Progress — spec complete, linked; implementation NOT started | `.worktrees/task-115` pristine at origin/main |
| TASK-129 village lens | 060 | In Progress — spec complete, linked; not dispatched | no worktree yet |
| TASK-128 stage defaults | — | To Do — **deliberately unspecced**: write its spec against the NOW-shipped tabs/rows (systems 5, exercise 6, lesson row, strips) | `patterns/stage-defaults.md` is the authority page |
| TASK-67 fork duels | — | To Do, optional tail | only if the operator asks or the queue is dry |

Halted implementer agents from the prior session are NOT resumable across
sessions — dispatch fresh `spec-implementer` agents; nothing is lost (127/115
worktrees have zero commits; their specs are the full instructions).

## Resume order

1. **Player-docs refresh** (Sonnet agent, doc-only, commits only `docs/player/`):
   run the `player-docs` skill; fix the broken source ref (pages referencing
   renamed wiki notes `metatron*.md` → `guardian*.md`) and the guardian-vocabulary
   sweep (page renames like playing-via-metatron.html are the skill's call). Then
   check TASK-121 AC #6 and mark it Done with `--final-summary`.
2. **Merge TASK-119**: from `.worktrees/task-119` — fetch; if main moved, rebase
   (expect pin-orphaning only; see "Pin fix procedure" below), re-run gates, then
   `check-merge-drift pr` → push → PR → merge → verify `gh api ... .merged` →
   cleanup worktree/branch → re-ground (wiki notes: executor, curriculum-ladder,
   sim-loop, tui-client, cli-promptworld, morgue, chronicle, guardian-orders;
   then player docs) → tick specs/054 tasks.md → sync board → Done.
   Its PR body should credit two cross-feature fixes riding it: the PR #90
   survival-seeder seq pre-stamp (replay fix) and the console-suppresses-briefing
   interaction rule. **After merging, check TASK-135** (curriculum production
   wiring) — spec 054's rubric emitter likely satisfies it; note that on 135.
3. **Dispatch TASK-127 (Sonnet) and TASK-115 (Opus)** in parallel from their
   worktrees (fast-forward each to origin/main first; re-cut if stale). The prior
   session's dispatch prompts are reproducible from the specs; key context to
   include: skin contract is on main (new fiction strings → skin lookup + default
   table + doc twin + completeness test; fiction-denylist sweeps will fail bare
   literals); guardian package is `internal/guardian`; 127 ships the shared
   reportCardView renderer, 115 consumes it (115's new event `guardian.report_card`
   is NEW vocabulary — allowed; existing `metatron.*` names are FROZEN).
   Merge serially, smaller first, re-ground between.
4. **Spec TASK-128** (full Spec Kit; Sonnet, watch for escalation — it touches
   every mode's layout) — now that everything it governs exists. Then dispatch.
5. **Dispatch TASK-129** (spec 060 ready; Sonnet).
6. Re-check the runbook's "Done means" list; TASK-67 remains optional.

## Standing rulings (operator decisions this session — do not relitigate)

- **THE one-way door: persisted `metatron.*` event names NEVER rename until
  TASK-134** (event-log format_version + migration) ships. Operator team-review
  decision: names get MIGRATED, not aliased; TASK-121 merged as the
  freeze-everything interim (verified byte-identical); the real rename is 134's
  AC #4. Sweep prose/TUI/Go identifiers/docs freely; NEW `guardian.*` event
  names are fine; existing serialized vocabulary is frozen (annotations at every
  definition site cite spec 052 ruling 2).
- **Live rubric gauges** (119): resolved by panels/exercise.md — live per-term.
- **Ceremony voice / ambient postmortem** (127): resolved in the overlay pages —
  both voices instrument-authoritative; ambient = morgue-only, no report card.
- **Report card** = rubric checklist (127's renderer) + attribution note (115),
  one artifact — spec 063 standing resolution 1.
- **Tutor guide** = compiled game substrate, never a player skill (stage-3 lock
  untouched) — spec 063 standing resolution 2.
- **129's strip is display-only** (page ruling; board AC's parity clause is
  vacuous); look-cursor deferred with the ruling recorded in spec 060.
- Interrupt-policy watch item (ceremony fatigue) stands unresolved by design.

## Session-start checks (operator-requested; they caught everything that mattered)

Run all three before acting, and again after any long gap:

1. **In-Progress-vs-branches reconcile**: `backlog task list --plain` In Progress
   set vs `git branch --list 'task-*'` + `git worktree list` — explain every
   mismatch (specs-done-but-not-dispatched is a legitimate state if the task's
   notes say so).
2. **Spec-number collision scan**: `ls specs/ | sed -E 's/^([0-9]+)-.*/\1/' | sort | uniq -d`
   — and ALWAYS `git fetch` + re-check `origin/main:specs/` immediately before
   pushing a new NNN. Numbers are claimed hourly; three collisions happened today
   (051, 055, 058→059→063). Renumber yours, never theirs, when theirs is merged.
3. **Runbook-claims vs git**: the table above vs `gh pr list --state merged` and
   the board. Plus `node scripts/check-merge-drift.mjs session` (hook-enforced
   choke-point gates from spec 051; `worktree` before cutting, `pr` before
   `gh pr create`).

## Hard-won mechanics (this repo, this week)

- **Concurrent sessions are constant** (an MVLS survival lane + others share this
  repo AND the root checkout; main merges roughly hourly). Merge serially,
  smaller-first; rebase before merge; verify the residual base gap is
  board/wiki/spec-only (`git diff --name-only HEAD...origin/main`) before merging
  over it.
- **Pin fix procedure**: every rebase orphans `docs/design/tui/` `verified_against`
  pins (they reference rewritten hashes). After the final rebase: map old pin →
  the rebased equivalent commit (match by subject), sed across `docs/design/tui/`,
  amend the tip docs commit, re-run `check-tui-design.mjs --changed`.
- **Run backlog/spec-bridge/tick commands from repo ROOT only** — inside
  `.worktrees/` the CLI resolves the branch's stale board and "Task not found"
  lies to you.
- **`.specify/feature.json` is shared root state** other sessions overwrite;
  don't trust it — manage spec dirs explicitly (cp the template yourself).
- Wiki re-grounds and player-docs refreshes **delegate well to Sonnet agents**
  (doc reconciliation tier); implementer reports come back for planning-tier
  gating. Implementers NEVER touch backlog/ or tick tasks.md — orchestrator does
  bookkeeping from root.
- Rebase reconciliation across feature boundaries is real work: when a sibling
  merge lands new code inside a package your branch renames/sweeps, send the
  branch back to ITS OWN implementer agent with explicit doctrine (it has the
  context); budget for 2–4 rebase rounds per long-lived branch.

## Done means (unchanged from the runbook)

All of TASK-124/125/126/127/128/129/117/119/115/121 Done via merged PRs; every
affected spec-047 page `status: shipped` with fresh pins; `check-tui-design.mjs`
green on main; wiki + `docs/player/` fresh on main; no stale worktrees.
