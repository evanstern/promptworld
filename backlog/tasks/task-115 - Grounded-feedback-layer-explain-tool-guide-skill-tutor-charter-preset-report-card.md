---
id: TASK-115
title: >-
  Grounded feedback layer: explain tool, guide skill, tutor-charter preset,
  report card
status: In Progress
assignee: []
created_date: '2026-07-25 04:43'
updated_date: '2026-07-25 18:55'
labels:
  - learning-game
  - metatron
dependencies: []
ordinal: 86000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Learning-game synthesis Wave 1 (docs/design/learning-game-synthesis.md, 2026-07-25; operator decision 5: the angel IS the tutor). One work item merging three vault recommendations: (a) a read-only, grant-gated, registry-derived explain tool serving deterministic mechanics facts — tool rosters, charge/miracle costs, decision classes/points/budgets, map glyphs — so the angel never confabulates mechanics (the unreliable-manual hazard: the frame forbids inventing events but nothing grounds facts); (b) a default skills/guide.md + tutor-charter preset teaching the base angel to answer how-do-I-play questions through it ('ask your angel' becomes the game's whatis command and a rep of the skill being taught); (c) the post-turn report card — a cheap-chain critique attributing outcomes to charter text ('your charter never mentions coordinates; the miracle was rejected twice for them'), riding the TASK-63 trace. Contract: explain is pull, report card is push, ONE shared data source so the grader never grades on vibes (RimWorld one-corpus/two-deliveries precedent). Tutor lane doctrine: charge-free, faith-free, excluded from every rubric; no initiative-frame relaxation (explaining is speech, not an act). Grounding: Analysis-In-Game-First-Teaching rec 2, Analysis-Learning-Game-Fit rec 4.

Spec: specs/059-grounded-feedback
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 explain tool exists: read-only, registry-derived, grant-gated; answers are derived facts, never model-generated
- [ ] #2 Default guide skill + tutor-charter preset make a fresh world's angel a competent orientation tutor
- [ ] #3 Report card attributes outcomes to charter text citing event-log evidence, on a cheap chain, sharing the explain tool's data source
- [ ] #4 Tutor-lane exclusions hold: no charges spent, no world events, no faith earned, absent from all rubrics
- [ ] #5 Report card surface: guardian-console card at stopping points + postmortem; badge between; never mid-run
- [ ] #6 Guardian verbs + example asks reachable from the deterministic ? floor (D9)
- [ ] #7 All new strings skin-token-resolved (D2)
- [ ] #8 Spec phase: Setup
- [ ] #9 Spec phase: Foundational — the shared data source
- [ ] #10 Spec phase: User Story 1+2 — explain + tutor-lane doctrine (P1)
- [ ] #11 Spec phase: User Story 3 — tutor guide (P2)
- [ ] #12 Spec phase: User Story 4 — report card (P2)
- [ ] #13 Spec phase: User Story 5 — the ? guardian section (P3)
- [ ] #14 Spec phase: Polish & Cross-Cutting Concerns
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Reorient 2026-07-25 rescope (docs/design/reorient-2026-07-25-ui.md, D5/D9/D2): report card renders as a guardian-console card at natural stopping points (run end / pause / exercise resolution; badges between) and inside the postmortem takeover (TASK-127) — never a mid-run interruption. The stage's grantable verbs + one example ask per verb must also be reachable from the deterministic ? floor (D9 guardian section), not only via the tutor. All new fiction strings are skin-token-resolved (D2 — token contract from TASK-121 ships first). Render surface page authored in TASK-123 before build.

Model tier: Opus 4.8 (spec-implementer, model=opus). Rubric: guardian turn-pipeline/prompt-composition (injection-adjacent) + new llm route kind, cross-package — senior tier per constitution Principle V and the runbook Lane 3 assignment. Standing resolutions recorded in spec: report card = checklist + attribution note (one artifact); tutor guide = compiled game substrate, not a player skill (stage-3 lock untouched). DISPATCH GATED on TASK-121's merge (skin contract; console seam already merged via #87).
<!-- SECTION:NOTES:END -->
