---
id: TASK-111
title: >-
  Metatron survival autonomy: genesis watch orders, act-on-own authority,
  targeting context
status: To Do
assignee: []
created_date: '2026-07-25 03:00'
updated_date: '2026-07-25 03:10'
labels: []
dependencies: []
priority: high
ordinal: 19000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
World-01 evidence: charges regenerated to cap and sat unused while Ash starved (day 2) and Oak froze (day 6) — the angel is structurally turn-less (turns only on player chat or order match; world-01 had almost no orders). 3 of its 4 miracles were door-rejected on invalid coordinates because the turn prompt never includes positions/passability. Decision (user 2026-07-24): the angel ACTS on its own for survival — not merely warn. Scope: (1) genesis-seeded system-origin watch orders (near-death, starvation, exposure), exempt from the 3-player-order cap, non-expiring; (2) carve survival out of the initiative frame (metatron/turn.go:749) so system survival turns may send visions/work miracles without player authorization — clock control and non-survival orders stay player-authority; (3) add a villager-positions + passability digest to miracle tool guidance (metatron/turn.go:510, tool/derive.go:235). Near-term slice of the larger direction in the agentization task; machinery built here must survive that redesign.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 new worlds start with the three survival watch orders active; existing worlds gain them via a one-time backfill or miracle door
- [ ] #2 a survival order match can land a vision or miracle with no player in the loop, still charge-gated
- [ ] #3 miracle guidance includes live positions/passability; invalid-target rejections drop to ~0
- [ ] #4 guardrails intact: no villager removal, no free miracles, charge economy unchanged
<!-- AC:END -->
