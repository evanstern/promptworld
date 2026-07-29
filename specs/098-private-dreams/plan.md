# Implementation Plan: Private dreams — consolidation clustering + habituation (TASK-99)

**Branch**: `task-99-private-dreams` | **Date**: 2026-07-29 | **Spec**: [spec.md](spec.md)

## Summary
Per-agent embedding-cluster detection + habituation in the nightly consolidation
slot (geometry-first, ambiguous-band LLM fallback within the existing slot),
outcomes as recorded salience-revision/merge events, rngAt-seeded zeroable
boundary jitter, dials in the tuning manifest.

## Technical Context
**Language**: Go. **Surfaces**: internal/mind (consolidation pass), internal/sim
(event types + apply arms for salience-revision/merge, memory store weights),
tuning manifest, TUI digest, event-types.md. **Testing**: privacy perturbation,
clustering/habituation units, routing, replay byte-identity, -race.
**Constraints**: single-store inputs only; no new LLM classes; additive events;
noise via rngAt only.

## Constitution Check
I–IV: PASS (spec 098 records D1–D4 incl. the AC#4 noise decision with rationale;
one branch/PR; test+seeded-world evidence; wiki re-pins in-branch — memory/
consolidation/embedding notes expected NEEDS-REVIEW).
V: PASS — **Opus** (internal/mind orchestration; replay-doctrine surface);
recorded on the board task.

## Project Structure
No new packages; vault references stay research-side (already cited on card).
