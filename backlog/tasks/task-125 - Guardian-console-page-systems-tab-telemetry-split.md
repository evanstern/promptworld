---
id: TASK-125
title: Guardian console page + systems-tab telemetry split
status: In Progress
assignee: []
created_date: '2026-07-25 14:44'
updated_date: '2026-07-25 18:05'
labels:
  - learning-game
  - tui
dependencies:
  - TASK-123
priority: high
ordinal: 95000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Reorient 2026-07-25 (decisions 1/2, D5/D10), Wave 3. The guardian conversation becomes a first-class full-height page: document-style turns in a fixed measure, composer for the minibuffer's focused state, charter/skills READ surface with binding status + honest lock notices, $EDITOR write handoff ('charter changed — next turn binds it'), report cards as console cards at stopping points (D5). Provider table, horizon rows, and spend move to a new systems dock tab (D10) — the guardian tab becomes fiction-only, making the TASK-121 skin boundary a file boundary.

Scope ruling (spec 053, per D5): this task ships the console CARD SEAM (inline slot + interface at stopping points); the report-card renderer itself belongs to TASK-127 (shared renderer on overlays/postmortem.md) and TASK-115 (card content/production) and composes here unchanged when they land.

Spec: specs/053-guardian-console
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Guardian page renders document-style turns; telemetry no longer shares the guardian tab
- [ ] #2 Charter/skills readable in-console with binding status; $EDITOR handoff confirmed in-TUI
- [ ] #3 Report card renders as a console card at run end / pause / exercise resolution; badges between
- [ ] #4 Spec phase: Setup
- [ ] #5 Spec phase: Foundational
- [ ] #6 Spec phase: User Story 2 — Telemetry moves out (P1) 🎯 merge-risk first
- [ ] #7 Spec phase: User Story 1 — The console page (P1)
- [ ] #8 Spec phase: User Story 3 — Charter/skills read surface + $EDITOR (P2)
- [ ] #9 Spec phase: Polish & Cross-Cutting Concerns
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Model tier: Sonnet (spec-implementer default). Rubric: single-package view/rendering slice (new page + dock tab + framework-standard ExecProcess), tests alongside — routine tier per constitution Principle V; runbook Lane 2 notes escalate-to-Opus if gates fail. Dispatched by UI-sweep orchestrator.
<!-- SECTION:NOTES:END -->
