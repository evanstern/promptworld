---
id: TASK-112
title: 'Metatron agentization: the angel as a first-class autonomous agent'
status: In Progress
assignee: []
created_date: '2026-07-25 03:00'
updated_date: '2026-07-30 14:58'
labels:
  - learning-game
dependencies:
  - TASK-111
priority: medium
ordinal: 6000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Direction decision (user, 2026-07-24, firm): the angel should become a SINGULAR AGENT separate from but akin to the villagers — the same basic agent construct (mind loop, memory, cadence, persona/soul files) with a different tool roster, extra context files, extra levers, and god mode on. Today the metatron is a request-driven console (turns only on player chat or order match, internal/metatron); after this, it is an inhabitant of the sim with its own scheduled cognition, observing the world through its replica and acting through its existing charge-gated tools. Implies: an angel decision-class in the cognition registry (points/staleness budget so the governor and horizon gate it like everyone else), an angel cadence, angel memory/consolidation (soul.md grows a real memory model instead of append-only digests), and the charter/skills/capabilities surface becoming its persona-equivalent. Builds on TASK-111 (survival autonomy is the first autonomous behavior and its order machinery must survive this redesign). Guardrail set is non-negotiable and carries over unchanged: no invented events, player words never reach villagers, no free miracles, no villager removal, reducer-side charge economy. Needs a full spec (speckit) before any implementation; expect cross-package work (metatron, mind, cognition, sim) — Opus-tier implement per constitution Principle V.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 spec exists under specs/ and is linked to this task via spec-bridge before implementation starts
- [ ] #2 angel runs on scheduled cognition (not only player/order triggers), gated by the cognition horizon like villager classes
- [ ] #3 angel uses the shared agent construct: persona/context files, memory, tool loop with its god-mode roster
- [ ] #4 all existing guardrails demonstrably intact post-redesign
- [ ] #5 Anti-self-grading guard: charter quality measurably changes autonomous performance on a seeded world (default vs authored charter delta)
- [ ] #6 Channel split is doctrine: the tutor voice (converse + explain tool) spends no charges, lands no world events, earns no faith, and is excluded from every rubric; world-acting is the graded artifact
<!-- AC:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-07-25 04:42
---
Reframed per learning-game synthesis (2026-07-25): agentization is 'the player programs an agent' — curriculum stage 3, not background AI. Encodes the three-lane initiative frame; tutoring requires NO doctrine relaxation (explaining is speech, not an act — rides the existing converse channel + one read-only tool grant). Open question gating this spec: the deliberate-incompetence ceiling (what the angel must never do well without a good charter); if adopted, incompetence applies to world-acting only, never tutor facts.
---
<!-- COMMENTS:END -->
