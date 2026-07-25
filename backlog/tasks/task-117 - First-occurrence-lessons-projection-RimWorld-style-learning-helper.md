---
id: TASK-117
title: First-occurrence lessons projection (RimWorld-style learning helper)
status: In Progress
assignee: []
created_date: '2026-07-25 04:43'
updated_date: '2026-07-25 18:21'
labels:
  - learning-game
  - tui
dependencies: []
ordinal: 88000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Learning-game synthesis Wave 2 (operator decision 8). RimWorld trigger model over the event stream, via the same client-side-projection pattern as the decision-trace view (TASK-63): first cog.outcome{suppressed} -> 'a thought was skipped for speed — press ? or ask Metatron why'; first gru.attacked, first charge regen, first metatron.order_expired, etc. Invariant: never a lesson you already know — auto-retire per lesson; seen-lessons state lives client-side/TUI-level (decided); exact home (per-user file vs per-world client state) and reset semantics are an open minor question. In scenario worlds, director-lite scheduled incidents double as lesson triggers. Cogmind caution recorded in the corpus: 'hot pink and blinking… and still people sometimes miss them' — hence every pushed lesson must also live in the ? overlay's pull reference. Lessons are model-free strings. Grounding: Analysis-In-Game-First-Teaching rec 4 (R3).

Spec: specs/055-lesson-row
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 First-occurrence events trigger one-line lessons in the feed; each lesson fires at most once (auto-retire, persisted seen-state)
- [ ] #2 Lesson trigger taxonomy covers at minimum: suppression, gru attack, charge regen, order expiry, first death
- [ ] #3 All pushed lessons are also reachable from the ? overlay
- [ ] #4 Dedicated lesson row with dwell + UI pointer + one-active/spacing/decay; stage-defaulted per decision 5
- [ ] #5 Prompting-lesson tier included in the minimum taxonomy
- [ ] #6 Lesson strings skin-tokened, each naming its pull path; seen-state per-user
- [ ] #7 Spec phase: Foundational (blocking prerequisites)
- [ ] #8 Spec phase: User Story 1 — first encounter teaches itself, exactly once (P1) 🎯 MVP
- [ ] #9 Spec phase: User Story 2 — the player's own prompting practice (P2)
- [ ] #10 Spec phase: User Story 3 — always findable again; the row knows its place (P3)
- [ ] #11 Spec phase: Polish & gates (same PR)
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Reorient 2026-07-25 rescope (decision 5, D2/D8): lessons render in a dedicated lesson row above the minibuffer — one active lesson, <=2 lines, dwells until done/dismissed, points at a key/tab (highlight field), anti-spam (one active, spaced, opportunity-decay). Default-on stages 1-2; badge+overlay-only default stage 3+/pre-ladder (stage-defaults machinery, TASK-128). Trigger taxonomy gains a PROMPTING tier: first rejected tool call, first custom charter observed, first fuzzy order. Every lesson string skin-tokened and carries its pull path ('press ? -> lessons'). Seen-state decided: per-user file (D8, unlocks.json precedent). Row page authored in TASK-123 before build.

Linked to spec 055 (specify+plan+tasks on main at 1170f46). Tier: Sonnet per constitution P.V rubric — single-package TUI view/rendering + client-side projection with tests alongside; no concurrency/governor logic, no doctrine-adjacent behavior. Runbook Lane 3 (session 3 claim).
<!-- SECTION:NOTES:END -->
