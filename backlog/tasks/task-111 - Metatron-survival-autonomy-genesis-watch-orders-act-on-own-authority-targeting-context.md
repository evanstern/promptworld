---
id: TASK-111
title: >-
  Metatron survival autonomy: genesis watch orders, act-on-own authority,
  targeting context
status: In Progress
assignee: []
created_date: '2026-07-25 03:00'
updated_date: '2026-07-25 22:09'
labels:
  - learning-game
  - guardian-survival
  - mvls
dependencies: []
priority: high
ordinal: 19000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
World-01 evidence: charges regenerated to cap and sat unused while Ash starved (day 2) and Oak froze (day 6) — the angel is structurally turn-less (turns only on player chat or order match; world-01 had almost no orders). 3 of its 4 miracles were door-rejected on invalid coordinates because the turn prompt never includes positions/passability. Decision (user 2026-07-24): the angel ACTS on its own for survival — not merely warn. Scope: (1) genesis-seeded system-origin watch orders (near-death, starvation, exposure), exempt from the 3-player-order cap, non-expiring, boot-seeded for existing worlds; (2) survival carve-out of the initiative frame (metatron/turn.go ~826) so system survival turns may send visions/work miracles without player authorization — clock control and non-survival orders stay player-authority; (3) villager-positions + passability digest in miracle tool guidance. Near-term slice of the agentization direction (TASK-112); machinery built here must survive that redesign.

Spec: specs/059-metatron-survival-autonomy
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 new worlds start with the three survival watch orders active; existing worlds gain them via a one-time backfill or miracle door
- [x] #2 a survival order match can land a vision or miracle with no player in the loop, still charge-gated
- [ ] #3 miracle guidance includes live positions/passability; invalid-target rejections drop to ~0
- [x] #4 guardrails intact: no villager removal, no free miracles, charge economy unchanged
- [ ] #5 Anti-self-grading guard: charter quality measurably changes autonomous survival performance on a seeded world (default-charter vs authored-charter delta)
- [x] #6 Spec phase: Foundational
- [x] #7 Spec phase: User Story 1 — Survival watches from birth (P1)
- [x] #8 Spec phase: User Story 2 — Survival authority carve-out (P1)
- [x] #9 Spec phase: User Story 3 — Targeting digest (P2)
- [x] #10 Spec phase: Polish & Cross-Cutting
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
MVLS sweep dispatch (2026-07-25): implementer tier Opus 4.8 — constitution V rubric: doctrine-adjacent authority change in metatron turn logic (the initiative doctrine is the safety frame); cross-file order/turn/charter semantics. Lane 1; merges last of the lane.

spec-bridge sync: Foundational: 1/1 · User Story 1 — Survival watches from birth (P1): 3/3 · User Story 2 — Survival authority carve-out (P1): 4/4 · User Story 3 — Targeting digest (P2): 2/2 · Polish & Cross-Cutting: 1/2

PR #90 squash-merged as 7367216. Implementer deviations gated and accepted: matchSurvival band predicate over needs_changed with live-only hysteresis latches (turn-arming state, mind-arming class — never replayed); starvation/exposure fire at the death-cause predicates (Food==0/Warmth==0) with re-arm at hungry/cold thresholds; zero-charge survival turns run and are recorded (one model call); digest passability via static worldmap.Passable (door authoritative); TutorCharter unchanged. Boot seed-if-absent covers new worlds via first boot (genesis-event variant not needed). Wiki re-pin dispatched; T012 completes with the batched player-docs refresh.

Gating note (2026-07-25): human ACs 1/2/4 proven by merged tests. AC3's second clause (invalid-target rejections drop to ~0 LIVE) and AC5 (charter-delta experiment on a seeded world) require live-world evidence the merge cannot provide — task stays In Progress pending those measurements; surfaced to operator via sweep report. Mechanism halves are done: digest ships + door round-trip regression green.
OPERATOR DECISIONS (2026-07-25, team review):
(1) AC#5 (anti-self-grading guard) FOLDS INTO specs/059-metatron-survival-autonomy as an explicit FR/SC, with an OPERATOR CHECKPOINT for the seeded A/B eval spend (the same treatment TASK-122 got). Review finding: AC#5 exists on this card and NOWHERE in the spec — zero hits for charter-quality / self-grading / default-charter. This is consistent with the gating note above: AC5 is precisely the clause the merge could not prove. Folding it into the spec is what makes the pending live-evidence measurement a defined artifact rather than an open-ended ask.
(2) THIS TASK KEEPS SPEC NUMBER 059. The duplicate is specs/059-grounded-feedback (TASK-115), which renumbers instead — see TASK-115.
(3) DECISION OVERTAKEN BY EVENTS — RECORDED AS DEBT, NOT AS A PLAN. The operator decided merge order TASK-121 -> TASK-111, with 111 bound to the skin-token contract (extending D2). PR #90 merged this task FIRST, ~15 minutes after that decision was recorded and before it reached the MVLS session. Consequence to carry: TASK-121's sweep now rebases through this task's new survival-turn code in internal/metatron/turn.go and orders.go, and any survival order text / seeded soul header this task landed in Metatron voice must be swept by 121 and made to pass 052's T008 fiction-denylist. Add those sites to 121's sweep inventory rather than assuming the skin-token binding held.

REVIEW FINDINGS (pinned-coordinate note now historical — the work merged): this task's three board coordinates had all drifted before dispatch — turn.go:749 was orderStatuses' loop body (the initiative frame is 817-831); turn.go:510 was landNudgeBatch's doc comment (guidance composes at :873); derive.go:235 was MetatronToolGuidance's func line. Kept as evidence for the standing rule: re-pin file:line diagnoses at dispatch, since the constitution's trivial-exemption and both runbooks lean on them. Also: internal/metatron is a THREE-program hotspot (052, 059-metatron, 115) that no runbook lists — specs 052 T012/T014, 059-metatron T005/T007 and 059-grounded-feedback T005 hold three incompatible expectations of the single constant metatronInitiativeFrame. That collision is now live in merged code.

Live-evidence follow-ups carded (operator request 2026-07-25): TASK-136 (AC#3 live rejection-rate measurement) and TASK-137 (AC#5 charter-delta experiment). This task's remaining ACs resolve from their evidence.

Renumber executed (2026-07-25): the 059 collision is resolved — THIS task keeps specs/059-metatron-survival-autonomy; TASK-115's spec moved to specs/063-grounded-feedback (not 060 — village-lens/conversation-loop-damper/instinct-yields claimed 060/061/062 while the collision sat open). spec-bridge check green: 63 linked tasks, none exceeding artifacts.

Sync deviation (2026-07-25, deliberate): the bridge derives Done from spec artifacts (all phases complete), but operator-authored ACs #3b/#5 remain open by explicit operator decision pending TASK-136/137 live evidence — status HELD at In Progress. The spec-phase AC #10 (Polish) is checked; only the live-evidence ACs gate Done.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-07-25 04:42
---
Learning-game synthesis (2026-07-25): survival autonomy is the SURVIVAL LANE of the three-lane initiative frame (tutor lane: ungraded speech + read-only explain tool, charge-free; survival lane: autonomous, charge-gated — this task; ambition lane: player-authorized, unchanged). The lane's competence ceiling is an open operator question that gates the TASK-112 spec — machinery built here must not preempt it. See docs/design/learning-game-synthesis.md.
---
<!-- COMMENTS:END -->
