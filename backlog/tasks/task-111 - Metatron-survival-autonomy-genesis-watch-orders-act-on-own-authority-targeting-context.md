---
id: TASK-111
title: >-
  Metatron survival autonomy: genesis watch orders, act-on-own authority,
  targeting context
status: In Progress
assignee: []
created_date: '2026-07-25 03:00'
updated_date: '2026-07-25 19:30'
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
- [ ] #10 Spec phase: Polish & Cross-Cutting
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
MVLS sweep dispatch (2026-07-25): implementer tier Opus 4.8 — constitution V rubric: doctrine-adjacent authority change in metatron turn logic (the initiative doctrine is the safety frame); cross-file order/turn/charter semantics. Lane 1; merges last of the lane.

spec-bridge sync: Foundational: 1/1 · User Story 1 — Survival watches from birth (P1): 3/3 · User Story 2 — Survival authority carve-out (P1): 4/4 · User Story 3 — Targeting digest (P2): 2/2 · Polish & Cross-Cutting: 1/2

PR #90 squash-merged as 7367216. Implementer deviations gated and accepted: matchSurvival band predicate over needs_changed with live-only hysteresis latches (turn-arming state, mind-arming class — never replayed); starvation/exposure fire at the death-cause predicates (Food==0/Warmth==0) with re-arm at hungry/cold thresholds; zero-charge survival turns run and are recorded (one model call); digest passability via static worldmap.Passable (door authoritative); TutorCharter unchanged. Boot seed-if-absent covers new worlds via first boot (genesis-event variant not needed). Wiki re-pin dispatched; T012 completes with the batched player-docs refresh.

Gating note (2026-07-25): human ACs 1/2/4 proven by merged tests. AC3's second clause (invalid-target rejections drop to ~0 LIVE) and AC5 (charter-delta experiment on a seeded world) require live-world evidence the merge cannot provide — task stays In Progress pending those measurements; surfaced to operator via sweep report. Mechanism halves are done: digest ships + door round-trip regression green.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-07-25 04:42
---
Learning-game synthesis (2026-07-25): survival autonomy is the SURVIVAL LANE of the three-lane initiative frame (tutor lane: ungraded speech + read-only explain tool, charge-free; survival lane: autonomous, charge-gated — this task; ambition lane: player-authorized, unchanged). The lane's competence ceiling is an open operator question that gates the TASK-112 spec — machinery built here must not preempt it. See docs/design/learning-game-synthesis.md.
---
<!-- COMMENTS:END -->
