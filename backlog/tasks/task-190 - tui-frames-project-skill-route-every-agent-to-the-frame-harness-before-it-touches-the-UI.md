---
id: TASK-190
title: >-
  tui-frames project skill: route every agent to the frame harness before it
  touches the UI
status: In Progress
assignee: []
created_date: '2026-08-03 02:52'
updated_date: '2026-08-03 03:03'
labels:
  - tui
  - design
  - tooling
  - dx
dependencies: []
ordinal: 172001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A project skill that teaches any agent working in this repo how to actually look at the terminal UI before changing it — read the generated frame, never imagine it — plus the handful of repo rules that otherwise cost a fresh session an hour of blocked commits. Today that knowledge lives in whatever the operator remembers to paste into a new session.

As the operator starting a UI session, I want to say "let's fix the chronicle pane" and have the agent already know how to see the chronicle pane, so I am not re-teaching the harness every time.

As an agent given a UI task in this repo, I want to be told to read `docs/design/tui/frames/mid-game__home__160x50.txt` rather than infer the layout from lipgloss calls, so what I change is grounded in the characters that actually reach the screen.

As a reviewer, I want UI changes to arrive as a before/after frame diff, so I can see what moved without checking out the branch and reproducing the state by hand.

As the operator, I want an agent that hand-edits a generated frame to be caught, because such an edit is not a design change — it is a false claim about what the client renders, and the next `--dump` silently erases it.

## Why now

TASK-187 (spec 112, merged in PR 157) shipped the frame harness: three fixture worlds, `promptworld frames`, and 132 generated frames under `docs/design/tui/frames/`. The capability exists; the *routing* to it does not. Nothing tells an agent the frames are there, and nothing stops one from reasoning about the UI the old way — which is the habit the harness was built to end.

The repo also has three traps a fresh session reliably hits, each costing a blocked commit and a confusing error before the cause is clear: the root checkout is read-only and rejects edits outside `.worktrees/`; the root-guard hook parses the raw command string, so a commit needs `-F <file>` and must run completely bare (no `2>&1`, no `| tail`, no `&&`) or its message text is misread as pathspecs; and a squash merge silently rewrites in-branch pins out of history. All three are documented, none are discoverable at the moment they bite.

## Scope

A `.claude/skills/tui-frames/` project skill carrying: how to see a screen (read the frame file; render live for an off-matrix size), the design loop (target frame first, implement until `--dump` matches, the diff is the review artifact), the authority-vs-evidence distinction (`docs/design/tui/pages|panels` say what a surface should be; frames say what it is; disagreement is a finding), the generated-file rule, the narrow-fallback caveat from spec 112 FR-008, and the three repo traps.

Plus a freshness/guard check script, in the shape of the existing `player-docs` skill (`.claude/skills/player-docs/scripts/check-freshness.mjs --check`): regenerate the matrix, confirm a clean tree, and fail loudly if a committed frame was hand-edited.

## Out of scope

- Changing the harness itself, the fixtures, or any renderer behavior.
- Wiring frames into `scripts/check-tui-design.mjs` as a regenerate-and-diff gate — still its own decision, unchanged by this card.
- Any actual UI/UX change. This card ships the routing, not a redesign.

Spec: specs/113-tui-frames-skill
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A .claude/skills/tui-frames/SKILL.md exists with frontmatter matching the repo's project-skill shape (name, description, metadata.author, user-invocable), so it is both operator-invocable and model-invocable on a UI task
- [ ] #2 The skill states, unambiguously, that a screen is SEEN by reading docs/design/tui/frames/<fixture>__<state>__<WxH>.txt and that generated .txt frames are never hand-edited
- [ ] #3 It documents rendering an off-matrix size live, the four committed sizes and what boundary each straddles, and --interactive for feel rather than layout
- [ ] #4 It states the design loop: target frame authored first, implement until --dump matches, and the frame diff is the review artifact
- [ ] #5 It states the authority-vs-evidence rule — docs/design/tui/pages|panels are the design authority, frames are evidence, disagreement is a finding to surface rather than silently resolve
- [ ] #6 It carries the spec 112 FR-008 caveat that narrow frames are content-height, so an agent does not 'fix' correct renderer behavior
- [ ] #7 It carries the three repo traps: root checkout read-only plus the worktree recipe, the bare 'git commit -F <file>' requirement, and merge-commit-only
- [ ] #8 A guard script under the skill's scripts/ regenerates the matrix and exits nonzero when the committed frames differ from a fresh dump, catching a hand-edited frame
- [ ] #9 The guard script passes on main at merge time, and running it leaves the working tree clean
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Model tier (constitution Principle V v1.3.0), recorded at dispatch 2026-08-02: Sonnet, model ID claude-sonnet-5, via the spec-implementer agent definition (the definition's frontmatter is the pin; no model parameter, which is silently ignored). Rubric justification: single-package authoring — one SKILL.md and one small Node guard script, no cross-package surface, no concurrency or scheduling logic, no doctrine-adjacent behavior change in the running system. This is the default tier and the escalation rubric does not fire. The doctrine CONTENT was authored at the planning tier (Opus 5) in specs/113-tui-frames-skill/ and handed over; the implementer types, grounds and tests it. Spec: specs/113-tui-frames-skill. Branch: task-190-tui-frames-skill, pushed at 84ecc4b5.

PR open 2026-08-02: https://github.com/evanstern/promptworld/pull/158 (branch task-190-tui-frames-skill, spec 113). Implemented by the Sonnet tier as pinned; orchestrator verified independently rather than on report. Gates: check-frames guard exit 0, check-tui-design all passed, player-docs 16 fresh 0 stale, check-merge-drift pr verdict=pass with NO findings. Guard proven by two independent negative controls on different frames (implementer tampered empty__home__160x50.txt, orchestrator tampered scenario__help__113x30.txt) — both named the file and exited 1, both restored clean. Guard runtime 1.4s, dumps to a temp dir so a green result is evidence rather than a self-fulfilling side effect. No wiki re-pin needed: branch touches no pinned source, verified. MERGE WITH --merge, NOT SQUASH.
<!-- SECTION:NOTES:END -->
