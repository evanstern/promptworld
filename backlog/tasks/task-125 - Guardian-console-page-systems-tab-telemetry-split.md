---
id: TASK-125
title: Guardian console page + systems-tab telemetry split
status: Done
assignee: []
created_date: '2026-07-25 14:44'
updated_date: '2026-07-25 18:49'
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
- [x] #1 Guardian page renders document-style turns; telemetry no longer shares the guardian tab
- [x] #2 Charter/skills readable in-console with binding status; $EDITOR handoff confirmed in-TUI
- [x] #3 Report card renders as a console card at run end / pause / exercise resolution; badges between
- [x] #4 Spec phase: Setup
- [x] #5 Spec phase: Foundational
- [x] #6 Spec phase: User Story 2 — Telemetry moves out (P1) 🎯 merge-risk first
- [x] #7 Spec phase: User Story 1 — The console page (P1)
- [x] #8 Spec phase: User Story 3 — Charter/skills read surface + $EDITOR (P2)
- [x] #9 Spec phase: Polish & Cross-Cutting Concerns
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Model tier: Sonnet (spec-implementer default). Rubric: single-package view/rendering slice (new page + dock tab + framework-standard ExecProcess), tests alongside — routine tier per constitution Principle V; runbook Lane 2 notes escalate-to-Opus if gates fail. Dispatched by UI-sweep orchestrator.

spec-bridge sync: 14/14 tasks done — merged via PR #87 (5af30cc), clean rebase, all gates green post-rebase incl. the new check-merge-drift pr gate. AC #3 satisfied per the recorded card-seam scope ruling (seam shipped; renderer/production = TASK-127/115 per D5). Judgment calls recorded in-PR: G shadowed in inspect/villagers modes (layered bindings win, documented in keymap.md); skills-locked line carries no file count (status surface honesty).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Guardian console + systems split shipped via spec 053 + PR #87 (merge 5af30cc). Full-height console page on G (document-style turns over the shared transcript, standard-minibuffer composer, charter/skills read surface with honest lock notices, $EDITOR handoff with content-hash confirmation, empty consoleCard seam for TASK-127/115); systems dock tab on 5 holds all relocated telemetry — the guardian tab is fiction-only, making TASK-121's skin boundary a file boundary. 7 design pages amended (guardian-console.md/systems.md → shipped) + re-pinned in-PR; wiki re-verified (tui-client console/systems prose, llm-provider-health location fix, 02812b5); player-docs refresh dispatched. All gates green post-rebase incl. the new merge-drift pr gate.
<!-- SECTION:FINAL_SUMMARY:END -->
