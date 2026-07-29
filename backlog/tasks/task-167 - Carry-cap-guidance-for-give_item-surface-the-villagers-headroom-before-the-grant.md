---
id: TASK-167
title: >-
  Carry-cap guidance for give_item: surface the villager's headroom before the
  grant
status: To Do
assignee: []
created_date: '2026-07-29 13:59'
labels:
  - mvls
  - guardian-survival
dependencies: []
priority: low
ordinal: 135000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Carded from TASK-163's evidence (docs/design/evidence/task-163/results.md, PR #128 merged 46f55b1): the other residual rejection class on the fixed binary — 2 of 5 — was give_item refused on the carry cap with CORRECT item kinds (food_cooked qty 200, meals qty 2 to a full villager). The enumerating door message already carries the numbers ('would exceed the carry cap (N/M already used)') and same-turn self-repair was observed once (qty 200 -> 8 -> landed), so this is polish, not a floor. CONSTRAINT: spec 016 FR-011 rules out door-side clamping — 'reject whole, never clamp to a partial delivery' (comment at the cap check in internal/sim/miracles.go applyItemGranted); the fix surface is guidance/context, not door semantics. Candidate shapes (decision needed): (a) villager carry headroom in the miracle-capable prompt digest (the spec 059 positions/passability digest already flows there); (b) a qty note in the give_item guidance gloss (internal/tool/derive.go miracleKindArgs, the TASK-163 pattern — but headroom is per-villager live state, not static vocabulary, so a static gloss can only say 'small quantities fit; the refusal names the cap'); (c) leave as-is and rely on the teaching door + repair loop, recorded as the deliberate posture.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Decision recorded on this card: digest headroom vs static gloss vs teaching-door-only, honoring FR-011 (no clamping)
- [ ] #2 If a change ships: implementation + tests, and a probe shows first-try quantities land (or the deliberate no-change posture is recorded with rationale and this card closes on that decision)
<!-- AC:END -->
