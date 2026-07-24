---
id: TASK-41
title: Surface the cognition horizon live in the TUI/status
status: In Progress
assignee: []
created_date: '2026-07-21 13:47'
updated_date: '2026-07-24 18:11'
labels:
  - ux
  - tui
dependencies: []
priority: medium
ordinal: 6000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
From TASK-39: suppression exists only in raw cog.outcome payloads; no TUI/status surface. At high speed even the narrator is router-gated (narrate.go:260-261), so the world goes silent with no explanation — 'nothing is happening' with no cause shown. Add a live per-class verdict at current speed (e.g. header/souls indicator: 'conversations suppressed at 32x — calibrate or slow down') and/or suppression counters. Natural home: the TASK-34 dock (new tab or status strip); pairs with the metatron cloud-tier-unreachable spinner fix noted on TASK-34.

Spec: specs/037-live-horizon-surface
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Spec phase: Foundational (blocking prerequisite)
- [x] #2 Spec phase: User Story 1 — Live per-class verdict at the current speed (P1) 🎯 MVP
- [x] #3 Spec phase: User Story 2 — Suppression counters (P2)
- [x] #4 Spec phase: User Story 3 — Headless status parity (P3)
- [ ] #5 Spec phase: Polish & lifecycle gates
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Full Spec Kit: specs/037-live-horizon-surface (spec → plan → tasks, clarify skipped — no ambiguity markers; choices grounded in drift audit + specs 028/034/035 precedents, recorded in research.md). 12 tasks: T001 cognition.LiveHorizon foundational; T002-T005 US1 wire field + status composition + TUI header badge + metatron-pane block (MVP); T006-T009 US2 orchestrator suppression counters via mind seam; T010 US3 CLI status parity; T011-T012 sweep + wiki re-pin. One PR off .worktrees/task-41 branch task-41-live-horizon-surface; spec docs commit to main at root.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Sequencing: depends on TASK-34's dock landing (its natural home is a dock status surface) — open/merge the TASK-34 PR first. Scope note: pair with the metatron cloud-tier-unreachable spinner fix recorded on TASK-34 (same 'daemon knows, UI doesn't say' family).

Re-grounding 2026-07-22: narrate.go router gate holds (~:261). Blocker cleared — TASK-34's dock is Done and landed; the sequencing note about merging the TASK-34 PR first is obsolete. Unblocked.

TASK-66 / decision-6 (2026-07-23): horizon legibility is a PREREQUISITE for classroom mode — a suppressed planner without a visible verdict reads to a learner as 'my prompt did nothing.' The client-decided teaching posture (TASK-77 chain-completion, TASK-78 soft speed cap) folds its learner-facing legibility needs into this task rather than a new one; spec 024 US6 (routing legibility) is the adjacent surface.

Drift audit 2026-07-23: PARTIALLY overtaken — the TUI decisions view now marks suppressed chains ('didn't think': internal/tui/decisions.go:232, :358; rendered views.go:1711), so 'suppression exists only in raw cog.outcome payloads' is no longer true. Still missing (remaining scope): the live per-class horizon verdict at current speed (header/status 'conversations suppressed at 32x') and suppression counters. Narrate gate moved to internal/mind/narrate.go:266-269 (was ~:260).

Model tier (constitution V rubric): Opus 4.8 — cross-package slice (cognition, llm, mind, ipc, tui, cmd) touching internal/cognition arithmetic doctrine and an internal/mind telemetry seam on the absorb goroutine; both are explicitly senior-tier territory ('cross-package/architectural', 'internal/llm, internal/cognition, internal/mind orchestration'). Single implementer run, all 5 phases, batched commits per phase.

Implementation complete (Opus 4.8 spec-implementer, single run): T001-T010 done on task-41-live-horizon-surface (460b777 foundational, c56bf22 US1, 396046b US2, ccf97c4 US3, 06bcc02 polish). Gates: go build/vet/test green; -race green for the new counter paths. T011 live validation (quickstart §2-§4, scratch worlds, branch binary): §2 PASS — 32x shows 'planner/conversation suppressed at 32x — calibrate or slow down', meeting thinking, spec-035 set_speed warning agrees with the horizon; §3 PASS — planner skipped grew 0→32 while hot, counts retained (not reset) after dropping to 1x, all classes thinking again; §4 PASS — no-LLM world prints no horizon section and the status JSON carries no horizon key. Two PRE-EXISTING failures found and reproduced on base main (NOT from this work): internal/tui TestCatalogSweep (event-types.md backticks daemon.llm_warning, fixture lacks a row) and internal/llm TestPreflightBootNeverFails -race flake (compressClock cleanup races RunPreflight) — filing as separate tasks. Remaining: PR + merge, then T012 wiki re-pin + player-docs.

spec-bridge sync: Foundational (blocking prerequisite): 1/1 · User Story 1 — Live per-class verdict at the current speed (P1) 🎯 MVP: 4/4 · User Story 2 — Suppression counters (P2): 4/4 · User Story 3 — Headless status parity (P3): 1/1 · Polish & lifecycle gates: 1/2
<!-- SECTION:NOTES:END -->
