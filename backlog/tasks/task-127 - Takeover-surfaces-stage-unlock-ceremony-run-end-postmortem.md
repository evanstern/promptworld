---
id: TASK-127
title: 'Takeover surfaces: stage-unlock ceremony + run-end postmortem'
status: Done
assignee: []
created_date: '2026-07-25 14:44'
updated_date: '2026-07-26 02:22'
labels:
  - learning-game
  - tui
dependencies:
  - TASK-123
priority: medium
ordinal: 97000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Reorient 2026-07-25 (decision 6, D6), Wave 4. One takeover-surface family (body-replacement slot precedent): the ceremony seizes the screen on curriculum.stage_unlocked (identity earned, what it grants, proving evidence, player-authorship voice per D6 — 'your charter proved The Written Word'); the postmortem seizes on run.ended (rubric outcome, report card, epitaphs with charter alignment, retry/fork jump-offs). Voice asymmetry: success speaks player authorship; failure speaks the morgue's no-blame evidence register. Both dismissable and replayable from pull surfaces (?, stages, morgue) — explicit AC. Open questions parked in the synthesis: ambient postmortem contents; ceremony score voice.

Spec: specs/056-takeover-surfaces
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Ceremony takeover on unlock while attached; player-authorship voice; skin-tokened
- [x] #2 Postmortem takeover on run.ended; morgue-evidence register; retry/fork jump-offs
- [x] #3 Both replayable from pull surfaces; dismiss is one keypress; never stack
- [x] #4 Spec phase: Setup
- [x] #5 Spec phase: Foundational
- [x] #6 Spec phase: User Story 1 — Postmortem takeover (P1)
- [x] #7 Spec phase: User Story 2 — Ceremony takeover (P1)
- [x] #8 Spec phase: Polish & Cross-Cutting Concerns
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Model tier: Sonnet (spec-implementer default). Rubric: single-package overlay state machine + rendering, tests alongside — routine tier per constitution Principle V. Both parked operator questions verified RESOLVED in the authored overlay pages (ambient postmortem = morgue-only; ceremony = both voices, instrument authoritative) — runbook checkpoint condition not met, proceeding per 'the pages win'. DISPATCH GATED on TASK-121's skin-contract merge (Lane 3 ordering); spec complete and ready.

Dispatched (UI-sweep orchestrator, handoff 2026-07-25b step 3): spec-implementer on Sonnet per recorded rubric; worktree .worktrees/task-127 fast-forwarded to 9386e6a before dispatch. Gate condition met (TASK-121 merged, PR #94). Parallel with TASK-115 (Opus); merge order: smaller first, serial, re-ground between. Implementer warned of pre-existing red TestCatalogSweep on main (TASK-140 hotfix in flight).

Implementation complete (Sonnet spec-implementer): 3 commits, tip 223771c; all 11 spec tasks done; PR #99 open, gated by orchestrator — design gate green, race suite green except the pre-existing TestCatalogSweep red that PR #98 (TASK-140) fixes; merge order: #98 first. Seven implementer judgment calls reviewed and accepted at the planning tier, notably: (1) rubric met/missed markers are event-presence-based until TASK-119's rubric machinery lands (documented in both overlay pages as a known simplification — reconcile at the 119/127 rebase); (2) scored/ambient detection reads Manifest.Scenario client-side (FR-006, no new IPC); (3) takeover wins over console page and help, esc peels one layer. reportCardView seam (TASK-115): unexported views.go symbols — reportCardView/reportCardFact/reportCardFactsFromEvents/reportCardFactsFromEvidence + consoleCard wrapper. Done flip held on PR #99 merge.

spec-bridge sync: Setup: 1/1 · Foundational: 2/2 · User Story 1 — Postmortem takeover (P1): 3/3 · User Story 2 — Ceremony takeover (P1): 2/2 · Polish & Cross-Cutting Concerns: 3/3 — status In Progress → Done

Merged via PR #99 (squash ded11c2) after a two-round rebase over TASK-119 (union-resolved the connectedMsg briefing-reset × postmortem-auto-open hunk; full -race suite green post-resolution). Human ACs #1-3 on merge evidence: ceremony takeover skin-tokened w/ authorship chapter (D6); postmortem in morgue register w/ retry/fork jump-offs; both replayable (? overlay section 4 + p key), single-keypress dismiss, postmortem-always-wins no-stack. Design pages re-pinned to squash commit on main post-merge. Known simplification recorded: rubric markers event-presence-based pending 119-emitter reconciliation (now that both are merged, a follow-up may tighten — noted, not blocking).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
All spec tasks complete (Setup: 1/1 · Foundational: 2/2 · User Story 1 — Postmortem takeover (P1): 3/3 · User Story 2 — Ceremony takeover (P1): 2/2 · Polish & Cross-Cutting Concerns: 3/3). Derived Done by spec-bridge sync.
<!-- SECTION:FINAL_SUMMARY:END -->
