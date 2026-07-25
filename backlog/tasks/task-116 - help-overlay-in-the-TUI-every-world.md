---
id: TASK-116
title: '? help overlay in the TUI (every world)'
status: To Do
assignee: []
created_date: '2026-07-25 04:43'
labels:
  - learning-game
  - tui
dependencies: []
ordinal: 87000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Learning-game synthesis Wave 1 (operator decision 8: onboarding is every-world, TUI-level). Context-sensitive per pane/mode: current keys first (Cogmind basic/advanced tiering), then a NetHack-style screen-region walkthrough — header anatomy (speed suffix, llm/suppressed badges), map glyph legend, dock tabs. Content is static strings derivable from the keymap design doc (docs/design/tui/patterns/keymap.md) + footer-hint machinery. Load-bearing rationale: a no-LLM world has no tutor (reflex-only villages are first-class), so the overlay is the charter-independent floor beneath an angel that may be absent, down, or mid-repair — not redundant with the tutor. 45 years of roguelike ? convention. Grounding: Analysis-In-Game-First-Teaching rec 3 (R1).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 ? opens a context-sensitive overlay in every pane; basic keys page first, advanced tier behind a second press
- [ ] #2 Screen-region walkthrough covers header anatomy, map glyph legend, and dock tabs
- [ ] #3 Works identically in no-LLM (reflex-only) worlds
- [ ] #4 Every pushed lesson (first-occurrence projection, when it lands) is also reachable from the overlay's pull reference
<!-- AC:END -->
