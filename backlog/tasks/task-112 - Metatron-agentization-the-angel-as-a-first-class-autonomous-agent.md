---
id: TASK-112
title: 'Metatron agentization: the angel as a first-class autonomous agent'
status: To Do
assignee: []
created_date: '2026-07-25 03:00'
updated_date: '2026-07-25 03:10'
labels: []
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
<!-- AC:END -->
