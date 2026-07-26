---
id: TASK-146
title: >-
  Wiki corpus-spec v2 adoption: CAPSULES.md + 35 notes over the 8000-char size
  budget
status: To Do
assignee: []
created_date: '2026-07-26 15:38'
labels:
  - wiki
  - tech-debt
dependencies: []
priority: low
ordinal: 116000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Surfaced during the TASK-141 wiki re-pin (2026-07-26): this corpus has no CAPSULES.md, so it predates corpus-spec v2. Running capsules.mjs generates it — and adoption flips the freshness gate's size budgets from warn-only to FAILURES: 35 notes are over the 8000-char body budget (event-types 69.5k, testing-strategy 47.2k, executor 40.5k, tui-client 37.8k, sim-state-reducer 36.2k, agent-mind 33.7k...) and 6 descriptions exceed the 500-char capsule budget (guardian.md 1286). A generated CAPSULES.md was reverted to keep the gate green. Adopting means: split the oversized notes summary-style (or mark size_budget_exempt with reasons), rewrite the 6 oversized capsules for routing, then generate CAPSULES.md once and keep it regenerated on every wiki pass. Worth doing with/after TASK-145 since the in-PR wiki gate will make every branch pay the gate's costs.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 All notes within the 8000-char budget or explicitly size_budget_exempt with a reason
- [ ] #2 All description capsules within 500 chars
- [ ] #3 CAPSULES.md generated and the freshness gate passes in v2 (failure) mode
<!-- AC:END -->
