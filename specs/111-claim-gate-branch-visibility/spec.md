# Feature Specification: Claim gate sees spec numbers held by pushed branches

**Feature Branch**: `task-188-claim-gate-branch-visibility`

**Created**: 2026-08-02

**Status**: Claim stub — spec number 111 reserved by this commit (spec 065 claim
protocol). The real spec (problem, requirements, acceptance mapping) lands in the
next commits on this branch, before any implementation.

**Input**: TASK-188 — `scripts/check-merge-drift.mjs` claim mode defines spec-number
ownership solely by presence on `origin/main` (`takenSpecNumbers()`, line 614;
`runClaim()`, line 1387; design intent stated at lines 461-465). Two concurrent
sessions therefore both claimed number 110 on 2026-08-02 — TASK-173 holds
`specs/110-absence-attribution` and TASK-187 holds `specs/110-tui-frame-harness`,
both branches pushed to origin, neither claim stub merged to main, both passed the
gate. Branch-vs-main collision detection already exists (`specNumberCollisions()`,
line 635, used by session and pr modes); branch-vs-branch is the gap this spec closes.
