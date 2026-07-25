---
id: TASK-116
title: '? help overlay in the TUI (every world)'
status: Done
assignee: []
created_date: '2026-07-25 04:43'
updated_date: '2026-07-25 06:51'
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
- [x] #1 ? opens a context-sensitive overlay in every pane; basic keys page first, advanced tier behind a second press
- [x] #2 Screen-region walkthrough covers header anatomy, map glyph legend, and dock tabs
- [x] #3 Works identically in no-LLM (reflex-only) worlds
- [x] #4 Every pushed lesson (first-occurrence projection, when it lands) is also reachable from the overlay's pull reference
- [x] #5 Spec phase: Setup
- [x] #6 Spec phase: Foundational (blocking prerequisites)
- [x] #7 Spec phase: User Story 1 — `?` answers "what can I press right now" (P1) 🎯 MVP
- [x] #8 Spec phase: User Story 2 — the screen explained (P2)
- [x] #9 Spec phase: User Story 3 — the floor holds with no angel (P3)
- [x] #10 Spec phase: User Story 4 — pushed lessons findable again (P4)
- [x] #11 Spec phase: Polish & Cross-Cutting
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
spec-bridge sync: Setup: 0/1 · Foundational (blocking prerequisites): 0/2 · User Story 1 — `?` answers "what can I press right now" (P1) 🎯 MVP: 0/5 · User Story 2 — the screen explained (P2): 0/2 · User Story 3 — the floor holds with no angel (P3): 0/1 · User Story 4 — pushed lessons findable again (P4): 0/2 · Polish & Cross-Cutting: 0/3

Implementation dispatched 2026-07-25: full spec 045 slice (T001-T016) to spec-implementer on Sonnet (default tier) — rubric: single-package view/rendering code, tests alongside, no concurrency/doctrine surface. Branch task-116-tui-help-overlay in .worktrees/task-116. Known collision with task-31 TUI surfaces recorded in plan.md; rebase step T015.

spec-bridge sync: Setup: 1/1 · Foundational (blocking prerequisites): 2/2 · User Story 1 — `?` answers "what can I press right now" (P1) 🎯 MVP: 5/5 · User Story 2 — the screen explained (P2): 2/2 · User Story 3 — the floor holds with no angel (P3): 1/1 · User Story 4 — pushed lessons findable again (P4): 2/2 · Polish & Cross-Cutting: 3/3 — status In Progress → Done

Human ACs verified against PR #76 (merged 5536b2f): #1 ? opens context-sensitive overlay in every pane, basic→advanced tiers (help_test routing suite); #2 walkthrough covers header anatomy/glyph legend/dock tabs (content-presence tests; gru glyph legend gap closed); #3 byte-identical no-LLM (nil status/replica tests); #4 pull-reference seam shipped — helpLesson table + contract, fixture-entry test proves content-only addition (full delivery rides the future first-occurrence projection). Design note recorded: legend↔overlay drift impossible by construction (shared table), but no independent oracle validates mapGlyphs entries themselves. Post-merge wiki-update obligation open.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
All spec tasks complete (Setup: 1/1 · Foundational (blocking prerequisites): 2/2 · User Story 1 — `?` answers "what can I press right now" (P1) 🎯 MVP: 5/5 · User Story 2 — the screen explained (P2): 2/2 · User Story 3 — the floor holds with no angel (P3): 1/1 · User Story 4 — pushed lessons findable again (P4): 2/2 · Polish & Cross-Cutting: 3/3). Derived Done by spec-bridge sync.
<!-- SECTION:FINAL_SUMMARY:END -->
