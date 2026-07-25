---
id: TASK-102
title: >-
  Embed traffic cannot clear provider-health preflight conditions (model-missing
  sticks)
status: Done
assignee: []
created_date: '2026-07-24 21:59'
updated_date: '2026-07-25 05:06'
labels:
  - llm
  - observability
dependencies: []
references:
  - specs/042-embedding-memory-retrieval/quickstart.md
priority: medium
ordinal: 1000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Found during spec 042 T025 walkthrough: the provider-health preflight compares model ids and raises model-missing (e.g. bare all-minilm vs tagged all-minilm:latest), but embedding calls bypass observeSuccess, so successful embed traffic never clears the condition — the warning is persistent and spurious. Fix direction: either route Embed() results through the same health-observation path as chat calls, or exempt embedding providers from id-compare preflight (they prove themselves per call). Workaround documented in spec 042 quickstart + docs/llm-providers.md: pin the fully tagged model id.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Successful embed calls clear (or never raise) provider-health conditions for their provider
- [x] #2 A world configured with a bare model alias that resolves at the endpoint does not carry a permanent spurious warning
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Delivered in PR #74 (squash-merged 2026-07-25). Fix direction (a): Embed() success clears the preflight condition via clearPreflightCondition — preserves boot-time preflight value for the now-load-bearing embedding class (TASK-105 context) while killing the sticky spurious warning. Regression test seeds the condition exactly as preflight would (verified failing without fix); tool-silent guard test proves embeds never touch non-preflight conditions. Full -race suite green. Doc notes updated (spec-042 quickstart, docs/llm-providers.md). Wiki re-pin owed: docs/wiki/llm-provider-health.md names TASK-102 as an open gap and sources internal/llm/llm.go — running wiki-update next.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-07-25 04:54
---
Implementation started 2026-07-25. Tier: Sonnet spec-implementer (rubric: single-package routine slice — provider-health observation path for embeds, diagnosis and ACs pinned on card; trivial-track exemption applies). Worktree .worktrees/task-102.
---
<!-- COMMENTS:END -->
