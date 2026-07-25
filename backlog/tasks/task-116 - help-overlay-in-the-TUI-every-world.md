---
id: TASK-116
title: '? help overlay in the TUI (every world)'
status: In Progress
assignee: []
created_date: '2026-07-25 04:43'
updated_date: '2026-07-25 06:23'
labels:
  - learning-game
  - tui
dependencies: []
ordinal: 87000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Learning-game synthesis Wave 1 (operator decision 8: onboarding is every-world, TUI-level). Context-sensitive per pane/mode: current keys first (Cogmind basic/advanced tiering), then a NetHack-style screen-region walkthrough — header anatomy (speed suffix, llm/suppressed badges), map glyph legend, dock tabs. Content is static strings derivable from the keymap design doc (docs/design/tui/patterns/keymap.md) + footer-hint machinery. Load-bearing rationale: a no-LLM world has no tutor (reflex-only villages are first-class), so the overlay is the charter-independent floor beneath an angel that may be absent, down, or mid-repair — not redundant with the tutor. 45 years of roguelike ? convention. Grounding: Analysis-In-Game-First-Teaching rec 3 (R1).

Spec: specs/045-tui-help-overlay
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 ? opens a context-sensitive overlay in every pane; basic keys page first, advanced tier behind a second press
- [ ] #2 Screen-region walkthrough covers header anatomy, map glyph legend, and dock tabs
- [ ] #3 Works identically in no-LLM (reflex-only) worlds
- [ ] #4 Every pushed lesson (first-occurrence projection, when it lands) is also reachable from the overlay's pull reference
- [ ] #5 Spec phase: Setup
- [ ] #6 Spec phase: Foundational (blocking prerequisites)
- [ ] #7 Spec phase: User Story 1 — `?` answers "what can I press right now" (P1) 🎯 MVP
- [ ] #8 Spec phase: User Story 2 — the screen explained (P2)
- [ ] #9 Spec phase: User Story 3 — the floor holds with no angel (P3)
- [ ] #10 Spec phase: User Story 4 — pushed lessons findable again (P4)
- [ ] #11 Spec phase: Polish & Cross-Cutting
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
spec-bridge sync: Setup: 0/1 · Foundational (blocking prerequisites): 0/2 · User Story 1 — `?` answers "what can I press right now" (P1) 🎯 MVP: 0/5 · User Story 2 — the screen explained (P2): 0/2 · User Story 3 — the floor holds with no angel (P3): 0/1 · User Story 4 — pushed lessons findable again (P4): 0/2 · Polish & Cross-Cutting: 0/3
<!-- SECTION:NOTES:END -->
