---
id: TASK-149
title: 'Report-card truth: unify all card surfaces on sim.EvaluateRubric'
status: To Do
assignee: []
created_date: '2026-07-26 17:57'
labels:
  - game-ui
  - pedagogy
dependencies: []
references:
  - docs/design/reorient-2026-07-26-ui.md
priority: high
ordinal: 119000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Reorient 2026-07-26 decision 1 (docs/design/reorient-2026-07-26-ui.md). The report card in postmortem/ceremony/console grades by generic event presence (reportCardFactsFromEvents/Evidence, internal/tui/views.go:863-880) and renders ✓ on a failing outcome (agent.died: 2) at the game's most salient teaching moment, while sim.EvaluateRubric sits unused outside the exercise tab. Unify all card surfaces on EvaluateRubric (✗ renders on failure) and author the-law's real evaluator (persist the charter Default flag into state — documented blocker at internal/sim/scenario.go:277). Sequence after TASK-144 (same code).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 All report-card surfaces (postmortem, ceremony, console card) derive verdicts from sim.EvaluateRubric; a failed term renders ✗
- [ ] #2 the-law has a production evaluator: charter Default flag persisted into state; live gauges stop rendering permanently pending
- [ ] #3 Design pages' 'known simplification' notes (overlays/postmortem.md, panels/exercise.md) amended same-PR; check-tui-design.mjs --changed passes
<!-- AC:END -->
