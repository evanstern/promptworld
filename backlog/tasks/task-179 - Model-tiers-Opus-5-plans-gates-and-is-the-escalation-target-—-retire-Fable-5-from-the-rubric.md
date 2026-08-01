---
id: TASK-179
title: >-
  Model tiers: Opus 5 plans, gates, and is the escalation target — retire Fable
  5 from the rubric
status: In Progress
assignee: []
created_date: '2026-08-01 04:19'
updated_date: '2026-08-01 04:21'
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
- [ ] #1 Constitution Principle V names Claude Opus 5 (claude-opus-5) as the planning/gating tier and as the senior-implementation escalation target, with Sonnet retained as the default implementation tier; amended via speckit-constitution with the Sync Impact Report updated and the version bumped from 1.2.0
- [ ] #2 CLAUDE.md's Model-tiered workflow block mirrors the amended principle and cites the new constitution version
- [ ] #3 .claude/agents/spec-implementer.md description and escalation rubric name Opus 5 as the escalation target; frontmatter stays model: sonnet
- [ ] #4 rg --hidden 'Fable|Opus 4\.8' over CLAUDE.md, .specify/, and .claude/ returns no hits
- [ ] #5 Lands as one PR merged with gh pr merge --merge; no wiki note pins the three files (verified), so no re-pin rides the PR
<!-- AC:END -->
