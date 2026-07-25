---
id: TASK-117
title: First-occurrence lessons projection (RimWorld-style learning helper)
status: To Do
assignee: []
created_date: '2026-07-25 04:43'
labels:
  - learning-game
  - tui
dependencies: []
ordinal: 88000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Learning-game synthesis Wave 2 (operator decision 8). RimWorld trigger model over the event stream, via the same client-side-projection pattern as the decision-trace view (TASK-63): first cog.outcome{suppressed} -> 'a thought was skipped for speed — press ? or ask Metatron why'; first gru.attacked, first charge regen, first metatron.order_expired, etc. Invariant: never a lesson you already know — auto-retire per lesson; seen-lessons state lives client-side/TUI-level (decided); exact home (per-user file vs per-world client state) and reset semantics are an open minor question. In scenario worlds, director-lite scheduled incidents double as lesson triggers. Cogmind caution recorded in the corpus: 'hot pink and blinking… and still people sometimes miss them' — hence every pushed lesson must also live in the ? overlay's pull reference. Lessons are model-free strings. Grounding: Analysis-In-Game-First-Teaching rec 4 (R3).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 First-occurrence events trigger one-line lessons in the feed; each lesson fires at most once (auto-retire, persisted seen-state)
- [ ] #2 Lesson trigger taxonomy covers at minimum: suppression, gru attack, charge regen, order expiry, first death
- [ ] #3 All pushed lessons are also reachable from the ? overlay
<!-- AC:END -->
