---
id: TASK-140
title: >-
  main broken: story feed cannot render agent.recovery_stalled —
  TestCatalogSweep red on main
status: Done
assignee: []
created_date: '2026-07-25 23:47'
updated_date: '2026-07-26 01:53'
labels: []
dependencies: []
priority: high
ordinal: 110000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
TASK-104 (spec 064, PR #96) shipped the agent.recovery_stalled event and its post-merge wiki re-verify (574021e) documented it in docs/wiki/event-types.md, but internal/tui never got a digest entry — TestCatalogSweep (internal/tui/digest_test.go:265) fails on origin/main, blocking every branch's gate battery. Same shape as TASK-100 (daemon.llm_warning). Diagnosis: emitter internal/sim/executor.go:867 emits RecoveryStalledPayload{Agent, Goal, Need} (internal/sim/agents.go:1219); internal/tui/digest.go digestRegistry (~line 121) needs a row (pattern: agent.build_failed at digest.go:308 + subject-candidate table at digest.go:1629); catalogFixture in digest_test.go needs the matching fixture row (pattern: line 73). Trivial-exempt from Spec Kit: surgical fix, complete file:line diagnosis here, ACs below.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 digestRegistry renders agent.recovery_stalled (agent name, goal, need — honest abort voice per event-types.md row), with subject-candidate entry
- [x] #2 catalogFixture row added; TestCatalogSweep green
- [x] #3 full go test -race ./... green in the worktree
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Model tier: Sonnet (spec-implementer, default). Rubric: single-package TUI view/rendering fix with tests alongside — routine tier per constitution Principle V. Trivial-exempt from Spec Kit (surgical, file:line diagnosis on card, ACs on card). Dispatched by UI-sweep orchestrator; found while gating TASK-119's rebase (branch test failure reproduced identically on origin/main).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Merged via PR #98 (squash 87f7251). digestRegistry + subject-candidate rows for agent.recovery_stalled mirroring agent.build_failed; catalogFixture row ('Birch's warm_up stalled — warmth not recovering'). TestCatalogSweep green on main again (verified post-merge at root); full -race suite was green in the worktree (22 packages). Sonnet spec-implementer, trivial-exempt. Wiki note tui-client.md lists digest.go as a source — re-pin rides the TASK-119 re-ground cycle.
<!-- SECTION:FINAL_SUMMARY:END -->
