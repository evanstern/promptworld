---
id: TASK-95
title: >-
  Loud failure for non-build goals: silent bare intent_done on invalid/contested
  resolution
status: In Progress
assignee: []
created_date: '2026-07-24 18:27'
updated_date: '2026-07-29 21:17'
labels:
  - enhancement
dependencies: []
priority: medium
ordinal: 5000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Follow-up from TASK-91 / spec 038 (research D5): the silent-failure pattern fixed for the seven build_* goals is systemic — forage/chop/hunt/demolish/repair/quarry/cook/bathe invalid exits and the contested no-op re-checks (craft/cook/bathe/deposit/withdraw, executor.go) still resolve as a bare agent.intent_done indistinguishable from success, with no failure memory. Extend the agent.build_failed pattern (or a generalized agent.intent_failed) to these goals so agents can falsify beliefs about non-build actions too. Blast radius: replay expected-event sets, TUI digest, event-types.md, mind re-arm list.

Spec: specs/096-intent-failed-loud
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Invalid/contested non-build resolutions emit a distinct failure event + situated memory, documented in event-types.md
- [ ] #2 Replay byte-identity preserved; regression tests cover at least one gather and one station goal
- [ ] #3 Spec phase: Event + emission
- [ ] #4 Spec phase: Surfaces
- [ ] #5 Spec phase: Tests + grounding
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
board-sweep-2026-07-29 lane 1: spec 096 landed + linked; additive event type, no format-version bump per spec 094 doctrine. Tier: Sonnet — build_failed pattern extension with tests; escalation trigger: new reducer semantics beyond stampIntentOutcome.
<!-- SECTION:NOTES:END -->
