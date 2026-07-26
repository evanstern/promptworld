---
id: TASK-148
title: >-
  Reorient runs need ownership: per-checkout run registry + project-wide Stop
  gate can't handle concurrent sessions
status: To Do
assignee: []
created_date: '2026-07-26 16:30'
labels:
  - design
  - tooling
dependencies: []
priority: medium
ordinal: 118000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Observed live 2026-07-26: two reorient runs begun the same day (15:58, 16:25) while an unrelated session worked TASK-143/147. Failure modes: (1) .handoff/reorient/runs/ is per-checkout shared mutable state and the plugin's Stop gate fires in EVERY session for ANY in-flight run — constant nagging in lanes that don't own the run, and no way to distinguish mine/theirs/orphaned (we abandoned run 15-58-26 as orphaned; the owning session re-created it at 16:25). (2) The synthesis target is date-keyed (docs/design/reorient-YYYY-MM-DD-ui.md), so same-day runs collide on one output path. (3) No claim primitive — unlike specs (claim-before-work, spec 065: pushed artifact + push-rejection mutex), reorient runs have no origin-visible ownership or liveness. Rethink directions (undecided): owner field (session id) + heartbeat on the run manifest with the gate scoped to the owning session only; or align with claim-before-work (a tracked, pushed claim artifact per run); run-id-keyed synthesis targets; explicit takeover/abandon semantics that surface WHO began a run and when. NOTE: the durable fix lands in the praxisflux reorient plugin (~/neumo/projects/praxis — version-lockstep + merge-commit-only laws apply there), not in this repo; this card tracks the promptworld-side pain and the design conversation.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A design decision (recorded artifact) on run ownership: gate scoping, claim/liveness mechanism, and target-path uniqueness for concurrent reorient runs
- [ ] #2 Plugin change implemented in the praxis repo per its laws, and promptworld's plugin version bumped to pick it up
<!-- AC:END -->
