---
id: TASK-179
title: >-
  Model tiers: Opus 5 plans, gates, and is the escalation target — retire Fable
  5 from the rubric
status: Done
assignee: []
created_date: '2026-08-01 04:19'
updated_date: '2026-08-01 04:41'
labels:
  - doctrine
dependencies: []
priority: high
ordinal: 147000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Retire Fable 5 as this project's planning tier. Opus 5 becomes the model that writes specs, plans, gates, and reviews implementer reports, and it also becomes the tier that hard implementation slices escalate to; routine spec execution stays on Sonnet. Nothing about how work is done changes — this closes a drift between the artifacts and the model the operator actually runs.

As the operator, when I open a session on this repo, I want the constitution's tier rubric to name the model I actually run, so a sweep can pin an explicit model ID from it instead of silently inheriting my session model.

As a sweep orchestrator, when I author a runbook, I want one Opus ID (claude-opus-5) to resolve both the planning tier and the senior implementation tier, so tier-to-ID resolution is unambiguous at dispatch time.

As an implementer agent, I want my default tier (Sonnet) and my escalation target (Opus 5) stated once, with the agent definition and the constitution agreeing.

Trivial-exemption diagnosis (constitution, Development Workflow — surgical, complete file:line diagnosis, ACs on the card; no code changes, three prose files):
- .specify/memory/constitution.md:87 'Planning tier — Claude Fable 5'; :91 'Senior implementation tier — Claude Opus 4.8'
- CLAUDE.md:187, 190, 193, 196, 200 — the mirrored Model-tiered workflow block
- .claude/agents/spec-implementer.md:7-8, 21, 28, 31 — description and escalation rubric prose; the 'model: sonnet' frontmatter is correct and does not change

Out of scope: specs/*/plan.md Fable references are historical records of what planned each spec and stay as-is. praxisflux needs no change — pdlc/skills/sweep/SKILL.md:96 sources the tier rubric from the host project's rubric, and praxis ships no agent definitions and pins no models.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Constitution Principle V names Claude Opus 5 (claude-opus-5) as the planning/gating tier and as the senior-implementation escalation target, with Sonnet retained as the default implementation tier; amended via speckit-constitution with the Sync Impact Report updated and the version bumped from 1.2.0
- [x] #2 CLAUDE.md's Model-tiered workflow block mirrors the amended principle and cites the new constitution version
- [x] #3 rg --hidden 'Fable|Opus 4\.8' over CLAUDE.md, .specify/, and .claude/ returns no hits
- [x] #4 Lands as one PR merged with gh pr merge --merge; no wiki note pins the three files (verified), so no re-pin rides the PR
- [x] #5 Implementation tiers are pinned by explicit model ID in agent-definition frontmatter: .claude/agents/spec-implementer.md carries claude-sonnet-5 as the default tier, and a new .claude/agents/spec-implementer-opus.md carries claude-opus-5 as the escalation target
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Branch task-179-opus5-planning-tier, PR #152 open — https://github.com/evanstern/promptworld/pull/152

Evidence:
- .specify/memory/constitution.md — Principle V amended, version 1.2.0 -> 1.3.0, Sync Impact Report updated, Last Amended 2026-08-01. Planning tier = Claude Opus 5 (claude-opus-5) with scope extended to the non-implementation lifecycle verbs; senior implementation tier = Claude Opus 5 (same ID); implementation tier = Claude Sonnet (claude-sonnet-5, default, unchanged). Dispatches must state the tier's model instead of inheriting the session model. MINOR bump justified in the report: principle structure unchanged, no prior compliant work invalidated.
- CLAUDE.md — Model-tiered workflow block rewritten, version reference now v1.3.0.
- .claude/agents/spec-implementer.md — description + escalation rubric restated for Opus 5; model: sonnet frontmatter unchanged.

AC4 caveat: rg --hidden 'Fable|Opus 4.8' over CLAUDE.md, .specify/, .claude/ returns three hits, all inside the constitution's Sync Impact Report where they are the required record of what changed. No hits in any normative text. AC ticked on that reading.

Gates: node scripts/check-merge-drift.mjs pr -> verdict=pass, no findings. No wiki note pins the three files (frontmatter sources: scanned across docs/wiki/), so no re-pin or player-docs regen rides the PR.

AC5 outstanding: merge. gh pr merge --merge was denied by this session's permission classifier, so the merge is left to the operator — appropriate for a doctrine ratification. Card stays In Progress until PR #152 merges.

Scope amendment during the PR, recorded before merge.

While filing the praxis-side card I read praxis's own field logs and found the mechanism I had just written into Principle V is unreliable: docs/design/board-cost-test-runbook.md:339 in praxis records 'Agent tool model param silently ignored (3 fable dispatches killed early); pinned via .claude/agents/opus-implementer.md, actual model verified from transcript'. praxis fixed it by pinning explicit model IDs in agent-definition frontmatter, which is why that repo carries opus-implementer.md and sonnet-implementer.md.

So commit 12df6090 moves promptworld's pin to the same place:
- .claude/agents/spec-implementer-opus.md ADDED — senior tier as its own definition, model: claude-opus-5. Escalation now means dispatching this agent instead of passing a model param.
- .claude/agents/spec-implementer.md — frontmatter alias 'sonnet' replaced with explicit claude-sonnet-5; rubric points at the opus definition and cites the field case.
- Principle V and CLAUDE.md restated to match: the pin is the frontmatter model ID, escalation is choosing the other definition.

This deviates from AC3 as originally written ('frontmatter stays model: sonnet'). Same model, more explicit ID — AC3 replaced with the accurate criterion rather than ticked on a technicality.

Also corrected in the PR body: praxis ships no agent definitions IN ITS PLUGIN PAYLOAD, but the praxis repo itself does carry implementer agent defs for its own development. The original wording was too broad.

Branch freshened from origin/main by merge (baseLag 2 -> 0); pr gate re-run after both changes: verdict=pass, no findings.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Constitution v1.3.0 ratified and merged as PR #152, merge commit fe1a110e — a true merge, two parents, so no pins were rewritten.

Principle V now reads: planning and gating run on Claude Opus 5 [claude-opus-5], and so does every non-implementation verb of the lifecycle — sweep orchestration, reorientation, refactor triage, grounding, review. High-complexity implementation slices escalate to Opus 5 as well; routine slices stay on Sonnet [claude-sonnet-5]. Fable 5 and Opus 4.8 are retired from the rubric, surviving only in the Sync Impact Report as the record of what changed and in specs/*/plan.md as historical records of what planned each spec.

The mechanism matters more than the names. Tiers are pinned by explicit model ID in agent-definition frontmatter: .claude/agents/spec-implementer.md carries claude-sonnet-5, and the new .claude/agents/spec-implementer-opus.md carries claude-opus-5. Escalation means dispatching the other definition, NOT passing a model parameter — praxis field evidence from 2026-07-31 records that parameter being silently ignored, with three dispatches running on the orchestrator's model before being killed. That evidence arrived mid-PR and changed the design; the deviation from the original AC3 was recorded and the AC replaced rather than ticked past.

Gates: check-merge-drift pr passed with no findings, branch freshened from main by merge. No wiki note pins any of the four files, so no re-pin or player-docs regen was required.

Follow-on filed: TASK-180 here for two root-guard board-sync defects hit while claiming this task, and praxis TASK-91 for pdlc:bootstrap planting no model-tier rubric while pdlc:sweep requires one.
<!-- SECTION:FINAL_SUMMARY:END -->
