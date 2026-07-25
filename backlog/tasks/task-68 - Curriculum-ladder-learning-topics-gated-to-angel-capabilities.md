---
id: TASK-68
title: 'Curriculum ladder: learning topics gated to angel capabilities'
status: To Do
assignee: []
created_date: '2026-07-23 03:28'
updated_date: '2026-07-25 06:20'
labels:
  - review-2026-07-22
  - teaching-game
  - learning-game
dependencies:
  - TASK-64
priority: high
ordinal: 15000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
From the 2026-07-22 team review (new-ideas item 6), SHAPED by the client's stated progression (2026-07-22): "here's Metatron with base settings, you can learn to prompt him like he's Claude/ChatGPT, he has some basic tools" -> "now you learn to edit his instructions (a CLAUDE.md/SKILL.md topic)" -> "now you pick additional things your angel can do in the world (tools)". A learning-topic gate-to-feature pathway: completing a topic unlocks the next capability tier.

Scope: (a) Curriculum design artifact: the ladder of stages, each stage naming the prompt-engineering concept taught (conversational prompting; instruction-file authoring; capability/tool design; indirect influence and prompt-injection awareness — which the Metatron fiction already teaches natively), the world features it requires, and its pass signal. (b) Stage presets: world templates/configs per stage — stage 1 world grants base tools only; stage 2 enables charter editing; stage 3 opens the tool manifest (all substrate from TASK-64). Seeded scenario worlds as exercises ("survive the first night", "get your law passed") using existing systems (needs, gru, norms/votes, secrets) as lesson material, with the chronicle as the score narrative. (c) Gating mechanism: how a stage unlocks — self-serve manual unlock vs artifact-gated (this repo's educate plugin already models topic lifecycles and progress gates; evaluate reusing its shape for player-facing lessons vs a simpler in-game unlock file). Keep v1 gating simple: a per-world stage field in config that the capability manifest reads.

Depends on TASK-64 (instruction surface + tool gating is the substrate). The horizon decision (TASK-66) informs per-stage speed posture.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Curriculum ladder artifact exists: stages, concept taught per stage, required features, pass signal per stage
- [ ] #2 Per-stage world presets exist and are creatable (new --stage or template worlds)
- [ ] #3 Stage gating mechanism decided and implemented (capability manifest honors the world stage)
- [ ] #4 At least two seeded scenario exercises defined with their score-narrative framing
- [ ] #5 Reviewed with the client against their three-stage progression before implementation
- [ ] #6 Scenario/stage pass signals are event-derived and surface in-game (chronicle/status), not only in docs
- [ ] #7 Stages are fiction-named identities (never easy-mode framing), presented as an informed first-startup choice
- [ ] #8 Stage unlocks are artifact-gated on event-log evidence, not menu toggles
- [ ] #9 Each stage ships a per-stage quickstart page via the player-docs skill
- [ ] #10 Stage-1 orientation is delivered in-game via the tutor charter preset
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Curriculum decisions ratified (operator, 2026-07-25, spec session): (1) LADDER = 4 stages — client 3 (conversational prompting / instruction authoring / capability design) + capstone graduation stage (full roster, ambient endgame, canonization TASK-81 as signature). (2) NAMES = skin-provided display identities over neutral substrate ids (stage-1..4 + concept descriptions); default guardian-skin names drafted in spec for AC#5 client review — see TASK-121 pivot (de-themed skinnable persona). (3) UNLOCK GATES = scenario pass signals: 1→2 stage-1 scenario pass (first-night; run not ended + rubric pass, TASK-119 machinery); 2→3 stage-2 scenario pass with custom charter revision in force (spec 044 charter fingerprint proves it); 3→4 granted-tool act contributes to a stage-3 scenario pass. (4) UNLOCK HOME = per-user unlocks file, entries point at proving world+events (auditable); world creation offers earned stages + explicit informed override (you-are-skipping-X notice); self-directed-engineer audience posture. Spec next: specs/046, skin-aware per TASK-121 sequencing (68 spec does not wait for 121 implementation).
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-07-25 04:42
---
Promoted to High per learning-game synthesis (docs/design/learning-game-synthesis.md, 2026-07-25): this task is the spine of the v1 game. Absorbed ACs from both vault analyses (Analysis-Learning-Game-Fit, Analysis-In-Game-First-Teaching): Cogmind difficulty-rebrand lesson (stages as identities), artifact-gated unlocks per repo gate doctrine, DF per-audience quickstarts. Operator decisions 1-3 (staged sessions, director-lite, hybrid scoring) are given constraints.
---
<!-- COMMENTS:END -->
