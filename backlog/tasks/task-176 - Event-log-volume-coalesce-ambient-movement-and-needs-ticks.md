---
id: TASK-176
title: 'Event-log volume: coalesce ambient movement and needs ticks'
status: In Progress
assignee: []
created_date: '2026-07-30 16:42'
updated_date: '2026-07-30 18:52'
labels: []
dependencies: []
ordinal: 144000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Two-thirds of the event log is per-tick ambient noise — tiny needs deltas, single-step moves, the gru wandering — which balloons the database and slows every offline read (compare, tail, migrate) without adding story or replay value at that granularity.

As a player, I want a month-old world to stay snappy to open, compare, and migrate, not drag a multi-hundred-megabyte log behind it.
As an operator mining a playtest, I want the event log dominated by meaningful events, not movement ticks I have to filter past.

Evidence (playtest-1, 29 game-days): 1,011,063 events, 230MB world.db. agent.needs_changed (332,752) + agent.moved (332,525) + gru.moved (122,382) = 78% of all events. Candidate approaches: coalesce runs of steps into path segments, emit needs on threshold crossings rather than every delta, sample ambient-mover positions. Constraint: replay/determinism doctrine (TASK-75) — any coalescing must preserve whatever the reducer needs for byte-identical replay, so this may land as emission-shape change, not lossy compaction.

Spec: specs/104-ambient-event-coalescing
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A month-scale world's event volume is reduced several-fold for ambient movement/needs families without breaking replay byte-identity or downstream consumers (chronicle, digest, TUI)
- [ ] #2 The chosen approach is recorded as a design decision (emission shape vs compaction) with the determinism doctrine explicitly addressed
- [ ] #3 Spec phase: Contracts + the derived-progress engine (D1)
- [ ] #4 Spec phase: Movement at exact per-step fidelity (D2 — the hard slice)
- [ ] #5 Spec phase: Needs thinning (D3)
- [ ] #6 Spec phase: Gru derived motion (D4)
- [ ] #7 Spec phase: Consumers (D5)
- [ ] #8 Spec phase: Whole-system proofs + measurement
- [ ] #9 Spec phase: Grounding (spec 069 in-branch; spec 047 TUI gate)
<!-- AC:END -->



## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
sweep dispatch (runbook playtest-1-findings-sweep): tier Opus 4.8 — replay-determinism doctrine (spec 092/TASK-75) + architectural emission-shape change with cross-package consumers (sim reducer, TUI digest grammar, event-catalog wiki notes) per constitution P.V hard-slice rubric. Spec 104-ambient-event-coalescing. Design-fork operator checkpoint recorded in runbook.

OPERATOR RULING 2026-07-30 (spec 104 design fork, sweep checkpoint): Arm A emission-shape change ADOPTED — path-segment movement with EXACT per-step sighting fidelity (deterministic segment advancement or baked sighting payloads; encounter/seek behavior byte-identical), needs emission at bounded interval + immediate band-crossing emission, gru position sampling. NO log-format bump (spec 097 place_observed precedent). Old-world relief (one-time archive-and-fresh-log for existing worlds) OUT of spec 104 scope — existing snapshot-cut migrate covers it. Arm B (offline compaction) rejected: inverts log-is-truth doctrine, seq renumbering near-disqualifying.
<!-- SECTION:NOTES:END -->
