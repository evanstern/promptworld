---
id: TASK-157
title: >-
  Guardian directives and designations: the durable plan layer (survey,
  designations, hard directives)
status: In Progress
assignee: []
created_date: '2026-07-26 20:25'
updated_date: '2026-07-26 22:33'
labels:
  - learning-game
dependencies:
  - TASK-97
priority: high
ordinal: 126000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The guardian gains the DF/RimWorld player's plan-making verbs as a durable, checkable, villager-facing goal layer. Three pieces: (1) SURVEY read tools — survey_site(x,y,radius) deterministic site fact sheets (terrain mix, water/wood/rock distances, structures, passability), free and non-acting, the explain-tool + spec-059 targeting-digest pattern pointed at planning; (2) DESIGNATIONS — event-sourced world artifacts (settlement zone, structure site, wall line) placed/cancelled by guardian tools, rendered on the map, announced as village knowledge via the spec-041 place-grant machinery, each carrying a structural fulfillment predicate checked in the reducer (structure-of-kind-K-at-tile is a pure state check); entity discipline clones sim.GuardianOrder (deterministic IDs, one-way status, prune). (3) DIRECTIVES — issue_directive/cancel_directive guardian tools targeting a villager, group, or the village, binding a designation (or other checkable goal) with framing text + TTL, landing as recorded events through the injection door so the prompt firewall holds exactly as for visions; villager-side it becomes a decision-context block (spec 043) and a reflex rung.

DIRECTIVE HARDNESS (operator decision, 2026-07-26, firm): directives are HARD — a DIRECTIVE rung in decideIntent between SURVIVAL and PREP. Villagers first make sure they are not dying, then execute active directives, then free time (prep/wander/fun). HOWEVER conversations, hails, and dynamic world stimuli CAN and SHOULD interrupt directed work — interruption is life and must not be discouraged. If interruptions cause issues (thrash, stalled directives), work around them IN-GAME first (guardian re-issue, TTLs, standing-order watches); code fixes only when no in-game workaround exists.

Grounding: docs/wiki/guardian-orders.md (entity/lifecycle template), docs/wiki/reflex-prep-arbitration.md (the arbitration ladder the new rung joins), docs/wiki/mental-map-propagation.md (knowledge announcement), research/Game-Gameplay-Patterns/Indirect-Control-and-Divine-Intervention.md (world-level verbs doctrine). Feature ideation session 2026-07-26. Directive lifecycle events join observableEventTypes so existing standing orders compose with zero new trigger code. TASK-97's targeting grammar is the designation addressing input (dependency). Follow-up: guardian missions task (accept/decompose/pursue/report, gated on TASK-112).

Spec: specs/084-guardian-directives
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 spec exists under specs/ and is linked to this task via spec-bridge before implementation starts
- [ ] #2 survey read tool returns a deterministic site fact sheet, charge-free and non-acting (Read effect, GuardianReadGuidance path)
- [ ] #3 designations are event-sourced, map-rendered, villager-knowable, and carry structural fulfillment predicates stamped in the reducer
- [ ] #4 issue_directive/cancel_directive land through the injection door as recorded events; prompt firewall demonstrably intact
- [ ] #5 DIRECTIVE reflex rung sits between SURVIVAL and PREP: survival always preempts; directives preempt prep/wander
- [ ] #6 conversations/hails/dynamic stimuli interrupt directed work and the directive resumes afterward without code intervention (in-game-workaround-first doctrine proven)
- [ ] #7 directive lifecycle events join observableEventTypes so standing orders can watch them
- [ ] #8 Spec phase: Setup
- [ ] #9 Spec phase: Foundational — grammar entry point, entities, doors, sweeps (blocks all user stories)
- [ ] #10 Spec phase: US2 — designations placed, announced, rendered, fulfilled (P1) 🎯 MVP
- [ ] #11 Spec phase: US3 — directives through the injection door; observable lifecycle (P1)
- [ ] #12 Spec phase: US4 — the villager side: block + DIRECTIVE rung + interruption proof (P1)
- [ ] #13 Spec phase: US1 — survey_site (P2)
- [ ] #14 Spec phase: Cross-cutting surfaces
- [ ] #15 Spec phase: Grounding + gates (the wiki-in-PR lifecycle, spec 069)
<!-- AC:END -->



## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Sweep claim (runbook docs/design/faith-directives-sweep-runbook.md, signed-off 2026-07-26): spec 084-guardian-directives. Tier: Opus 4.8 — cross-package (guardian tools + sim designation/directive entities + reflex-ladder arbitration, doctrine-adjacent + map render + decision context). Dependency satisfied: TASK-97 merged (PR #123) — internal/target grammar + designation seam live. Collision checkpoint passed: card was To Do, no sibling branch.
<!-- SECTION:NOTES:END -->
