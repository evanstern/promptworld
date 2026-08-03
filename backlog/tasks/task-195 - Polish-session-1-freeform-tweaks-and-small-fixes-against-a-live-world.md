---
id: TASK-195
title: 'Polish session 1: freeform tweaks and small fixes against a live world'
status: In Progress
assignee: []
created_date: '2026-08-03 17:33'
updated_date: '2026-08-03 17:33'
labels: []
dependencies: []
ordinal: 177001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A single long-running card that holds one rapid polish session. Instead of a spec per fix, we
watch a live world, discuss small bugs / UI tweaks / gameplay nits as they come up, record each
decision on this card before code is written, and land the whole session on one branch in one PR.

## Use cases

- As a player, I want the small rough edges I keep bumping into — a misaligned pane, a confusing
  label, a villager doing something obviously silly — fixed quickly, without each one waiting on
  its own full spec cycle.
- As the operator, I want to sit in front of a real running world, point at what looks wrong, and
  have the decision written down before anyone writes code, so a fast session still leaves a
  paper trail.
- As a reviewer, I want one PR whose card lists every item worked, the diagnosis behind it, and
  whether it was decided ad-hoc or escalated to a spec.

## Session workflow (operator-authored, this session's contract)

1. Worktree + branch off origin/main; this card is claimed for the session's duration.
2. A runnable promptworld is kept live: world-02 in place (~/.promptworld/worlds/world-02),
   restarted against the branch build as we iterate.
3. This card is the only task; there are no per-item task cards.
4. Development sub-loop, repeated:
   a. Discuss a feature, bug, or tweak.
   b. Optionally ground the topic with research / wiki reads.
   c. Record the decision and its diagnosis on this card (Decision log below) BEFORE implementing.
5. Ad-hoc after a loop, decide whether the accumulated items warrant spec(s).
   a. If yes — write the spec(s), link via spec-bridge, execute them.
   b. If no — continue the sub-loop at 4a.
6. Execute any specs created.
7. Decide whether to go again (a fresh session is usually the right move at this point).
8. Test: operator visual QA on the live world; optionally a team review.
9. ONLY THEN: wiki re-pin, player docs, design references, PDLC gates. Nothing is re-pinned
   per item — grounding happens once, at the end, in-branch, before the PR.

Scope guard: this flow is for polish — FE/TUI changes and small gameplay tweaks against decisions
already made. Anything needing real design thought leaves this session and goes through the full
PDLC cycle on its own card.

## Decision log

(entries appended as the session runs — one per item: what, diagnosis with file:line, decision,
ad-hoc vs spec)
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Every item worked in this session has a decision-log entry on this card — item, file:line diagnosis, and the decision — recorded before its implementation
- [ ] #2 Any item exceeding the trivial-exemption bar (surgical fix + complete file:line diagnosis + ACs on this card) is escalated to a Spec Kit spec and linked via spec-bridge before implementation
- [ ] #3 All session work lands on a single branch and a single PR; no per-item task cards or PRs are created
- [ ] #4 Operator visual QA passes on the live world for every shipped item before the PR is opened
- [ ] #5 Grounding is done once at the end, in-branch: wiki re-pinned, player docs regenerated, tui-design amended where internal/tui changed, and the pr merge-drift gate is green
<!-- AC:END -->
