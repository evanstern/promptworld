---
id: TASK-112
title: 'Metatron agentization: the angel as a first-class autonomous agent'
status: Done
assignee: []
created_date: '2026-07-25 03:00'
updated_date: '2026-07-30 20:36'
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

Spec: specs/102-guardian-agentization
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 spec exists under specs/ and is linked to this task via spec-bridge before implementation starts
- [x] #2 angel runs on scheduled cognition (not only player/order triggers), gated by the cognition horizon like villager classes
- [x] #3 angel uses the shared agent construct: persona/context files, memory, tool loop with its god-mode roster
- [x] #4 all existing guardrails demonstrably intact post-redesign
- [x] #5 Anti-self-grading guard: charter quality measurably changes autonomous performance on a seeded world (default vs authored charter delta)
- [x] #6 Channel split is doctrine: the tutor voice (converse + explain tool) spends no charges, lands no world events, earns no faith, and is excluded from every rubric; world-acting is the graded artifact
- [x] #7 Spec phase: Angel class + scheduled lane
- [x] #8 Spec phase: Shared construct
- [x] #9 Spec phase: Doctrine enforcement
- [x] #10 Spec phase: Surfaces + evidence
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
board-sweep-2026-07-29 lane 2: spec 102 landed + linked (AC1 satisfied). OPERATOR RULING (2026-07-30, in-session checkpoint): deliberate-incompetence ceiling ADOPTED — world-acting initiative only, never tutor facts, never order compliance; encoded as D3 (default charter compiles the cap as data). Tier: Opus (card-stated, cross-package). DISPATCH GATED on TASK-164 evidence per the inherited runbook checkpoint (arm A ~75%).

OPERATOR RULING (2026-07-30): dispatch NOW in parallel with 164 arm B — the A/B delta is the measurement instrument (FR-006, runs on the agentized build post-merge), not a design input; arm A already validates the door under the default charter. Checkpoint satisfied visibly, not silently.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-07-25 04:42
---
Reframed per learning-game synthesis (2026-07-25): agentization is 'the player programs an agent' — curriculum stage 3, not background AI. Encodes the three-lane initiative frame; tutoring requires NO doctrine relaxation (explaining is speech, not an act — rides the existing converse channel + one read-only tool grant). Open question gating this spec: the deliberate-incompetence ceiling (what the angel must never do well without a good charter); if adopted, incompetence applies to world-acting only, never tutor facts.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Merged via PR #146. The guardian is a first-class autonomous agent on the shared villager construct: steward cognition class (de-themed serialized vocabulary per operator ruling — zero fiction carve-outs) with scheduled cognition budgeted/gated like every class; shared memory model + consolidation incl. dream phase; incompetence ceiling ADOPTED as default-charter data (soak-proven: default arm read-only, authored arm 3 landed autonomous acts); structural tutor split byte-identity-pinned; all five guardrails unmodified green; order door single-arbiter preserved; bundle tools excluded from scheduled roster (recorded ruling); opt-in per world. AC5 behavior delta proven live in-soak; the survival OUTCOME delta rides TASK-164 arm B (FR-006 EVIDENCE-PENDING, harness prepared). Opus tier; spec 102 all tasks done; reconciled across both concurrent sweeps.
<!-- SECTION:FINAL_SUMMARY:END -->
