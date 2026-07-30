---
id: TASK-176
title: 'Event-log volume: coalesce ambient movement and needs ticks'
status: To Do
assignee: []
created_date: '2026-07-30 16:42'
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
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A month-scale world's event volume is reduced several-fold for ambient movement/needs families without breaking replay byte-identity or downstream consumers (chronicle, digest, TUI)
- [ ] #2 The chosen approach is recorded as a design decision (emission shape vs compaction) with the determinism doctrine explicitly addressed
<!-- AC:END -->
