---
id: TASK-140
title: >-
  main broken: story feed cannot render agent.recovery_stalled —
  TestCatalogSweep red on main
status: To Do
assignee: []
created_date: '2026-07-25 23:47'
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
- [ ] #1 digestRegistry renders agent.recovery_stalled (agent name, goal, need — honest abort voice per event-types.md row), with subject-candidate entry
- [ ] #2 catalogFixture row added; TestCatalogSweep green
- [ ] #3 full go test -race ./... green in the worktree
<!-- AC:END -->
