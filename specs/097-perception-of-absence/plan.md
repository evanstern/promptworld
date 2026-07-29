# Implementation Plan: Perception of absence (TASK-80)

**Branch**: `task-80-perception-of-absence` | **Date**: 2026-07-29 | **Spec**: [spec.md](spec.md)

## Summary
Executor-side `agent.place_observed` on intent-completing arrivals (exhaustive
within placeScanRadius, emitter-computes, additive type) + low-salience deduped
situated memory; mind-side belief reconciliation through the TASK-79 seam
(confirm boost / bounded disconfirmation decay / silence unchanged) with dials in
the tuning manifest; digest + docs; soak evidence on a measure world.

## Technical Context
**Language**: Go. **Surfaces**: internal/sim (executor arrival path, memory
constructor, reducer apply arm, event catalog), internal/mind (belief
reconciliation via TASK-79 seam), TUI digest grammar, event-types.md, tuning
manifest dials. **Testing**: determinism/replay, dedup, belief-path units, soak.
**Constraints**: no LLM in the emission path; additive event (no format bump);
playtest world untouched — soak on a measure world.

## Constitution Check
I–IV: PASS (spec 097 records D1–D5 with rationale; one branch/PR; soak + test
evidence; wiki re-pins in-branch — executor-social-perception, sim-state,
event-types, belief/memory notes expected NEEDS-REVIEW).
V: PASS — **Opus** (cross-package sim+mind, epistemics doctrine-adjacent);
recorded on the board task.

## Project Structure
No new packages; dials via spec 048 manifest; evidence under
docs/design/evidence/task-80/.
