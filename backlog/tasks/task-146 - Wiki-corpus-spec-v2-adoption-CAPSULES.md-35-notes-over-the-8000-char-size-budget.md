---
id: TASK-146
title: >-
  Wiki corpus-spec v2 adoption: CAPSULES.md + 35 notes over the 8000-char size
  budget
status: Done
assignee: []
created_date: '2026-07-26 15:38'
updated_date: '2026-07-26 16:51'
labels:
  - wiki
  - tech-debt
dependencies: []
priority: low
ordinal: 116000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Surfaced during the TASK-141 wiki re-pin (2026-07-26): this corpus has no CAPSULES.md, so it predates corpus-spec v2. Running capsules.mjs generates it — and adoption flips the freshness gate's size budgets from warn-only to FAILURES: 35 notes are over the 8000-char body budget (event-types 69.5k, testing-strategy 47.2k, executor 40.5k, tui-client 37.8k, sim-state-reducer 36.2k, agent-mind 33.7k...) and 6 descriptions exceed the 500-char capsule budget (guardian.md 1286). A generated CAPSULES.md was reverted to keep the gate green. Adopting means: split the oversized notes summary-style (or mark size_budget_exempt with reasons), rewrite the 6 oversized capsules for routing, then generate CAPSULES.md once and keep it regenerated on every wiki pass. Landed deliberately AFTER TASK-145 so the corpus restructure is the first full exercise of the in-PR grounding rules.

Spec: specs/071-corpus-v2-adoption
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 All notes within the 8000-char budget or explicitly size_budget_exempt with a reason
- [x] #2 All description capsules within 500 chars
- [x] #3 CAPSULES.md generated and the freshness gate passes in v2 (failure) mode
- [x] #4 Spec phase: Setup
- [x] #5 Spec phase: User Story 1 — the corpus passes v2 (P1)
- [x] #6 Spec phase: User Story 2 — downstream intact (P1)
- [x] #7 Spec phase: Polish
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Claim per spec 065: specs/071-corpus-v2-adoption/ stubbed; compact spec authored (decisions: summary-style splits with parents keeping filenames; children inherit pins; split/tighten/exempt rubric; CAPSULES.md last; orchestrator-led with Sonnet fan-out per runbook lane 2). Sweep runbook: docs/design/pdlc-hardening-runbook.md.

spec-bridge sync: Setup: 1/1 · User Story 1 — the corpus passes v2 (P1): 4/4 · User Story 2 — downstream intact (P1): 2/2 · Polish: 2/2. Batch accounting: 14 Sonnet workers, 37 notes → 109 children, all verbatim-move ratios ≥100%, 0 exemptions; corpus-wide audit 161 notes 0 violations; INDEX coverage complete.

spec-bridge sync: Setup: 1/1 · User Story 1 — the corpus passes v2 (P1): 4/4 · User Story 2 — downstream intact (P1): 2/2 · Polish: 2/2 — status In Progress → Done
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
All spec tasks complete (Setup: 1/1 · User Story 1 — the corpus passes v2 (P1): 4/4 · User Story 2 — downstream intact (P1): 2/2 · Polish: 2/2). Derived Done by spec-bridge sync.
<!-- SECTION:FINAL_SUMMARY:END -->
