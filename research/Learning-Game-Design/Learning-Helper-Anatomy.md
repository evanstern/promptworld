---
title: Learning Helper Anatomy
aliases: [RimWorld ConceptDef, Adaptive Lesson System Internals]
tags: [rimworld, adaptive-help, lesson-system, contextual-help]
type: note
created: 2026-07-25
updated: 2026-07-25
related: ["[[Learning-Game-Design]]", "[[Teaching-Through-Play]]", "[[Observe-Intervene-Onboarding]]"]
---

# Learning Helper Anatomy

The per-lesson mechanics of RimWorld's adaptive learning helper, from the community
wiki and the community-decompiled 1.x source ([[_grounding]] § Area 5). This goes below
the player-facing description ("lessons trigger when relevant, retire when performed") to
how a lesson is actually defined, scored, and scheduled.

## A lesson is a data definition (`ConceptDef`)

Each lesson is one `ConceptDef` with, per the decompiled source
([ConceptDef.cs](https://github.com/Chillu1/RimWorldDecompiled/blob/master/RimWorld/ConceptDef.cs)):

- `helpText` — the instructional content (plus `helpTextController` for controller mode);
- `priority` — a float weight (lessons with `priority <= 0` are "triggered direct");
- `needsOpportunity` — whether the lesson only makes sense in a triggering context;
- `opportunityDecays` (default true) — whether a missed teaching moment expires;
- `highlightTags` — UI elements to highlight while the lesson is active
  (`HighlightAllTags()` applies the emphasis);
- `gameMode` — which program state the lesson applies in.

So a lesson bundles **content + trigger condition + UI pointer + decay policy** in one def,
and mods can add lessons the same way ([[_grounding]] § Area 5;
[RimWorld Wiki: Learning helper](https://rimworldwiki.com/wiki/Learning_helper)).

## Selection: desire scoring against a knowledge model

`LessonAutoActivator` schedules lessons
([LessonAutoActivator.cs](https://github.com/Chillu1/RimWorldDecompiled/blob/master/RimWorld/LessonAutoActivator.cs)):

- Gameplay code registers teachable moments via `TeachOpportunity()` with typed weights:
  `GoodToKnow` = 60, `Important` = 80, `Critical` = 100.
- Each concept's desire ≈
  `(priority + opportunity/100 × 60) × (1 − PlayerKnowledgeDatabase.GetKnowledge(concept))`
  — importance and timeliness, scaled by the player's remaining knowledge gap.
- **Knowledge itself decays** (`KnowledgeDecayRate = 0.00015`), so long-unused concepts can
  resurface; opportunities decay faster (`OpportunityDecayRate = 0.4`) when the def allows.
- Anti-spam gates: check every 15 frames, teach only the `MostDesiredConcept()`, only if
  desire exceeds a `RelaxDesire` threshold, fewer than 3 concepts are active, and no lesson
  is running; `timeSinceLastLesson` spaces lessons out.

## Retirement

Lessons are "automatically marked as learned when the player does the necessary
interaction, and can be marked as learned manually" (dismissal); remaining lessons surface
"as needed by circumstance, or on a slow timer"
([RimWorld Wiki: Learning helper](https://rimworldwiki.com/wiki/Learning_helper)). The
helper and a one-time scripted tutorial shipped together in Alpha 15 (2016), the tutorial
covering basics before "the adaptive learning helper starts up"
([Ludeon Alpha 15 announcement](https://ludeon.com/blog/2016/08/alpha-15-tutorial-and-drugs-released/)).

## Grounding

- [[_grounding]] — § Area 5 (RimWorld learning-helper per-lesson anatomy)
- [Chillu1/RimWorldDecompiled: ConceptDef.cs](https://github.com/Chillu1/RimWorldDecompiled/blob/master/RimWorld/ConceptDef.cs)
- [Chillu1/RimWorldDecompiled: LessonAutoActivator.cs](https://github.com/Chillu1/RimWorldDecompiled/blob/master/RimWorld/LessonAutoActivator.cs)
- [RimWorld Wiki: Learning helper](https://rimworldwiki.com/wiki/Learning_helper)
