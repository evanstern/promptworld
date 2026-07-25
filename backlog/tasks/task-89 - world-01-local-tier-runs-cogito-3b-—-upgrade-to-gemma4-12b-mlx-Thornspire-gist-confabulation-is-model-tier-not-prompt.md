---
id: TASK-89
title: >-
  world-01 local tier runs cogito:3b — upgrade to gemma4:12b-mlx (Thornspire
  gist confabulation is model-tier, not prompt)
status: In Progress
assignee: []
created_date: '2026-07-24 04:31'
updated_date: '2026-07-25 18:40'
labels:
  - emergent-lore
  - epistemics
  - operations
  - model-tier
  - mvls
dependencies: []
priority: medium
ordinal: 2000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Found during spec 030 (TASK-79) US3 eval, 2026-07-24. The spec 030 gist-attribution eval (specs/030-epistemic-hygiene/eval/decision.md) proved the fact-flattening / asserted-unperformed-action confabulation class ('discussed the glowy tendrils after investigating') is a property of the weak local model, not the outcome prompt: gemma4:12b-mlx (repo default local tier) produces 0/18 defects with the CURRENT prompt (controls 12/12), while cogito:3b — which world-01 actually runs per ~/.promptworld/worlds/world-01/llm.json — produces 3/18 with no improvement from attribution-preserving wording (5/18). Remediation is operational: point world-01's llm.json local tier at gemma4:12b-mlx (endpoint localhost:11434 already serves it). Historical world-01 lore already laundered into memories/beliefs is handled by spec 030's decay + provenance machinery once merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 world-01 llm.json local.model updated to gemma4:12b-mlx and the daemon restarted against it
- [ ] #2 A post-upgrade multi-scene gist sample from world-01 shows zero fact-flattened / asserted-unperformed-action shapes (spot-check recorded on this task)
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
MVLS sweep ops (2026-07-25): (1) verified world-01 llm.json routes already reference only gemma (gemma4:12b-mlx @ mbpro-m1.local:11434) + cloud — no route used cogito:3b any longer; removed the dangling unused cogito provider stanza (backup: llm.json.bak-task89). (2) World was format_version 3; backed up the whole dir (world-01.bak-task89) and ran promptworld migrate — v4, 8 villagers carried across at tick 538823 (day 7 11:40), v3 events archived in world.v3.db. (3) Daemon restarted clean against the migrated world: mind driver on (8 villagers), metatron on, orchestrator on. AC2 pending: multi-scene gist sample to spot-check once new conversations accumulate under gemma.
<!-- SECTION:NOTES:END -->
