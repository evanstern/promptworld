---
id: TASK-102
title: >-
  Embed traffic cannot clear provider-health preflight conditions (model-missing
  sticks)
status: To Do
assignee: []
created_date: '2026-07-24 21:59'
labels:
  - llm
  - observability
dependencies: []
references:
  - specs/042-embedding-memory-retrieval/quickstart.md
priority: low
ordinal: 85000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Found during spec 042 T025 walkthrough: the provider-health preflight compares model ids and raises model-missing (e.g. bare all-minilm vs tagged all-minilm:latest), but embedding calls bypass observeSuccess, so successful embed traffic never clears the condition — the warning is persistent and spurious. Fix direction: either route Embed() results through the same health-observation path as chat calls, or exempt embedding providers from id-compare preflight (they prove themselves per call). Workaround documented in spec 042 quickstart + docs/llm-providers.md: pin the fully tagged model id.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Successful embed calls clear (or never raise) provider-health conditions for their provider
- [ ] #2 A world configured with a bare model alias that resolves at the endpoint does not carry a permanent spurious warning
<!-- AC:END -->
