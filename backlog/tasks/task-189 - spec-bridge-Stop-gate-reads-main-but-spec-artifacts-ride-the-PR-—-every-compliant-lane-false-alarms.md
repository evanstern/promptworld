---
id: TASK-189
title: >-
  spec-bridge Stop gate reads main, but spec artifacts ride the PR - every
  compliant lane false-alarms
status: To Do
assignee: []
created_date: '2026-08-03 01:34'
updated_date: '2026-08-03 03:57'
labels:
  - gate
  - spec-bridge
  - spec-069
  - upstream
dependencies: []
priority: high
ordinal: 171001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The board's spec gate looks for a task's planning documents in the shared main copy of the project, but this project deliberately keeps those documents on the working branch until the pull request merges. The result is that a task which followed every rule correctly gets flagged as if it had skipped its planning work — and because it runs at the end of every turn, the false alarm repeats forever.

As a session finishing a turn, I want the Stop gate's warnings to mean something. If it cries wolf on every compliant lane, I learn to tune it out — and the one real violation it exists to catch gets tuned out with it.

As the operator, when I read "TASK-x is In Progress but only proves To Do", I want that to mean the spec work is genuinely missing — not that it is sitting on a branch exactly where our own doctrine told the session to put it.

As a session picking up someone else's lane, I want to trust the board's status over the gate's verdict, instead of having to open three branches by hand to work out which one is lying.

## Evidence (2026-08-02, all three lanes verified against origin)

| Spec dir | spec.md | plan.md | tasks.md | present on main? |
|---|---|---|---|---|
| specs/110-absence-attribution (TASK-173) | yes | yes | yes | NO — branch only |
| specs/112-tui-frame-harness (TASK-187) | yes | yes | yes | NO — branch only |
| specs/111-claim-gate-branch-visibility (TASK-188) | stub | yes (branch) | yes (branch) | stub only |

The gate fired on TASK-173 and TASK-187 on two consecutive turns, reporting "spec.md missing, plan.md missing, no tasks in tasks.md" for lanes whose artifacts are complete.

TASK-188 IS THE CLEAN REPRODUCTION and the reason this is a gate defect rather than two sloppy lanes. It followed both protocols exactly: spec 065's claim stub merged to main immediately (merge 1b17cd33), and spec 069's rule that spec content rides the PR. Main therefore holds specs/111/spec.md and nothing else. Running spec-bridge:link on TASK-188 today would produce a THIRD false alarm — "plan.md missing, no tasks in tasks.md" — against a lane that did everything right. That is why the link step was deliberately skipped when TASK-188 shipped.

## Diagnosis pinned

The bridge gate probes the FILESYSTEM at the resolved project root, which under this repo's iron-clad read-only-root doctrine is always the main checkout:

- lib/spec-derive.mjs:118-121 — deriveSpecState(specDir) builds has()/read() from existsSync/readFileSync on a plain path; every stage verdict downstream derives from those two.
- gates/bridge.mjs:153 — the missing-artifact message is literally ["spec.md","plan.md"].filter((f) => !existsSync(join(root, specDir, f))).
- gates/bridge.mjs:116-123 — linked tasks are discovered by reading backlog/tasks/ from the same root.
- The gate-runner resolves that root by locating a backlog/ directory, i.e. the root checkout — never a worktree, never a branch.

Nothing in the chain is aware that a task has a branch. The gate's implicit precondition is "a linked task's spec artifacts live on main while it is In Progress", which directly contradicts spec 069's lifecycle ("grounding rides the PR; step 7 is bookkeeping only") and spec 065's claim protocol (the stub alone lands early; content follows on the branch). No lane following current doctrine can satisfy it.

NOTE ON OWNERSHIP: the gate lives in the praxisflux spec-bridge PLUGIN (~/.claude/plugins/marketplaces/praxisflux/spec-bridge), not in this repo. There is no local escape hatch — .spec-bridge.json's entire surface is strictDone plus statusVocabulary (gates/bridge.mjs:25-35), neither of which affects where artifacts are read from. So this card is partly an upstream report and its resolution may be a plugin change this repo consumes.

## Candidate resolutions, NOT pre-decided

(a) Branch-aware derivation, upstream: resolve the linked task's branch and read spec artifacts from it (git show <branch>:<specDir>/<file>), falling back to the working tree when there is no branch. Same defect shape and same remedy as TASK-188 just applied to the claim gate, which learned to read branches for spec numbers — second gate, same lesson.
(b) Downgrade to a warning when the spec dir is absent from the working tree but present on a task branch, keeping the block for the genuinely-missing case.
(c) Reconcile the doctrine instead: land spec.md/plan.md/tasks.md on main at claim time, moving spec artifacts (not grounding) out of spec 069's rides-the-PR boundary.
(d) A per-project config knob selecting the artifact source. Cheapest upstream change, but leaves every other praxisflux consumer with the same false alarm.

Option (a) is the recommendation on file; (c) would trade a false alarm for a doctrine change and deserves explicit operator sign-off before anyone builds it.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The gate no longer reports missing spec artifacts for a linked task whose spec.md/plan.md/tasks.md exist on its own task branch but not yet on main
- [ ] #2 A genuinely missing spec artifact (absent from main AND from the task branch) still blocks, with today's message
- [ ] #3 A linked task with no branch at all behaves exactly as today — working-tree derivation, unchanged verdicts
- [ ] #4 TASK-173, TASK-187 and TASK-188 all pass the gate at their true statuses without any card being set back or any spec content moved to main early
- [ ] #5 The chosen resolution is recorded on this card with its rationale, including whether it lands upstream in the praxisflux spec-bridge plugin or as a doctrine change here
- [ ] #6 If the resolution is upstream, the plugin version carrying it is pinned on this card and verified live in this repo
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Concrete instance observed 2026-08-02, during TASK-191 (unrelated lane). The spec-bridge Stop gate blocked session exit with:

  [spec-bridge] TASK-173 is "In Progress" but specs/110-absence-attribution only proves "To Do": spec.md missing, plan.md missing, no tasks in tasks.md.

Every clause of that finding is false. On branch task-173-absence-attribution (pushed to origin at d8731a56, so auditable from any clone):

  specs/110-absence-attribution/spec.md      present
  specs/110-absence-attribution/plan.md      present
  specs/110-absence-attribution/tasks.md     present, 9 tasks
  specs/110-absence-attribution/evidence.md  present

On main the directory does not exist at all — which is exactly what spec 069 requires, since spec artifacts ride the PR and land at merge. TASK-173 is fully compliant with the wiki-in-PR lifecycle and the gate punishes it for that compliance.

Blast radius beyond the false alarm itself: the gate names a remedy ("set the task back") that is actively wrong. Following it would (a) assert something false about the world, and (b) reach into another session's claimed lane, which the spec 065 claim protocol forbids — a rejected claim is a stop-the-lane signal to surface, never a state to overwrite. A session that trusts the gate's remedy corrupts a peer's claim; a session that ignores it cannot cleanly exit. There is no correct action available to the blocked session, which is what makes this High rather than cosmetic.

Suggested resolution when this is worked: resolve the spec dir against the task's own branch (the board card names the task, and the branch is discoverable as task-<N>-<slug> / already reported by the session gate's branch table) before falling back to main. The merge-drift pr gate already reasons about branch-carried grounding this way, so the pattern exists in-repo.
<!-- SECTION:NOTES:END -->
