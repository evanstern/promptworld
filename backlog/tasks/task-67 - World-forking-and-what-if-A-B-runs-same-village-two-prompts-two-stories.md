---
id: TASK-67
title: 'World forking and what-if A/B runs (same village, two prompts, two stories)'
status: In Progress
assignee: []
created_date: '2026-07-23 03:28'
updated_date: '2026-07-26 18:52'
labels:
  - review-2026-07-22
  - teaching-game
  - learning-game
dependencies:
  - TASK-149
priority: high
ordinal: 16000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
From the 2026-07-22 team review (new-ideas item 5) — the killer teaching feature. Replay is model-free (LLM outputs are recorded inputs), so you cannot re-run yesterday under a new prompt. But the persistence substrate makes world FORKING cheap: save dirs are fully self-contained and copyable (proven by e2e scenario: copied worlds run), snapshots bound recovery, and each world is its own daemon. Fork the world at a point, diverge the charter/skills, run both live, and compare the chronicles: the most direct way a learner SEES what their prompt change did.

Scope: (a) promptworld fork <world> <new-name> [--at latest-snapshot] — copy the save dir truncated to a chosen snapshot boundary, assign a fresh world identity (name registration, socket, any world-scoped ids) so both run side by side; document the semantics of forking mid-log vs at-snapshot (simplest v1: latest snapshot only). (b) Lineage recorded in the fork (parent world + fork tick) as an event/metadata so provenance is durable. (c) A comparison surface: v1 can be CLI — promptworld compare <a> <b> [--since tick] rendering the two chronicles side-by-side or interleaved with divergence markers; a TUI view can come later. (d) Doctrine note: forks are independent worlds afterward (no merging, ever). Design question to settle in spec: does the fork inherit the LLM budget meter or get its own (interacts with the global monthly ceiling — review flagged cost attribution as coarse).

Depends on nothing, but pairs naturally with the decision-trace view (TASK-63): trace explains one run, fork contrasts two.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 promptworld fork creates a runnable copy at the latest snapshot with fresh identity; both worlds run simultaneously
- [ ] #2 Fork lineage (parent, fork tick) durably recorded in the new world
- [ ] #3 A compare surface renders two chronicles against each other with divergence visible
- [ ] #4 Forked world passes the determinism harness independently (replay to identical hash)
- [ ] #5 Budget-meter semantics for forks decided and documented in the spec
- [ ] #6 Spec Kit spec written and linked via spec-bridge before implementation (non-trivial task)
- [ ] #7 Compare/duel output includes an event-derived rubric (deaths, needs floors, norms passed, charge efficiency, rejected-call rate) rendered plain-language per the glossary discipline (no raw enums in a grade)
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Drift audit 2026-07-23: premises verified — save dirs self-contained/copyable (world-save-directory.md:15-16), snapshots bound recovery (snapshots.md:14-17), replay never re-calls a model (llm-orchestrator.md:20), and no fork/compare subcommand exists yet (main.go:52-88).

Reorient 2026-07-25 rescope (D7): v1 compare surface is the rubric-first scoreboard — plain-language rubric card with drill-down into interleaved chronicles at divergence points — sharing the postmortem's rubric renderer and the verdict-glossary discipline (a lost duel IS a postmortem, TASK-127). The shareable HTML retelling (two chronicles as one artifact — the Boatmurdered move) follows; dual side-by-side live TUI is deferred post-v1.

Reorient 2026-07-26 decision 3 (docs/design/reorient-2026-07-26-ui.md): promoted to HIGH and reframed as the loop's iteration rung — all D7 prerequisites shipped (spec 054 rubric evaluator, spec 056/063 report-card renderer, glossary discipline, postmortem register), so v1 is dramatically cheaper than when scoped. v1 = rubric-first scoreboard sharing reportCardView + sim.EvaluateRubric (AC #7's rubric should be EvaluateRubric terms, not a bespoke list); phase 2 = the Boatmurdered-style shareable HTML retelling (one renderer family); dual side-by-side TUI stays deferred. Depends on TASK-149 — the duel must not compare false checkmarks.

Sweep claim (runbook docs/design/reorient-2026-07-26-sweep-runbook.md): spec 076-world-fork-duel. Tier: Opus 4.8 — cross-package architectural (world-lifecycle fork, fresh identity, lineage events, determinism harness, compare surface). Dependency satisfied: TASK-149 merged (PR #113, f78358a) — duel scoreboard shares resolveReportCardFacts/reportCardView + sim.EvaluateRubric, comparing true verdicts.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-07-25 04:42
---
Rescoped per learning-game synthesis (2026-07-25): reframed from teaching feature to core game verb — the fork-duel is the grader and postmortem (charter A vs B on a seeded fork; 'here is what your prompt change did'). Hybrid-scoring decision applies: duels are a scored surface. Framed as the losing-is-fun postmortem artifact. Grounding: both vault analyses, recommendation 2 in each.
---
<!-- COMMENTS:END -->
