# Implementation Plan: agent.intent_failed for non-build goals (TASK-95)

**Branch**: `task-95-intent-failed-loud` | **Date**: 2026-07-29 | **Spec**: [spec.md](spec.md)

## Summary
Generalize the agent.build_failed pattern: emit `agent.intent_failed` (goal,
enumerated reason, position) + a situated failure memory wherever the executor's
non-build work goals resolve invalid/contested today as a bare intent_done; close
the intent ring via the existing stampIntentOutcome path; add the mind re-arm
entry, digest grammar entry, and event-types.md documentation.

## Technical Context
**Language**: Go. **Surfaces**: internal/sim/executor.go (invalid exits +
contested re-checks), internal/sim/state.go (apply arm via stampIntentOutcome),
internal/sim/memory.go (situated constructor), mind re-arm list, TUI digest
grammar, docs/event-types.md. **Testing**: per-goal failure matrix, replay
byte-identity, TestCatalogSweep, go test -race ./...
**Constraints**: additive event type (no format-version bump per spec 094
doctrine); successful paths byte-identical; build_* goals untouched.

## Constitution Check
I–IV: PASS (spec 096; one branch/PR; test matrix as evidence; wiki re-pins
in-branch — executor/event-types/sim-state notes expected NEEDS-REVIEW).
V: PASS — Sonnet (pattern extension, tests alongside); escalation trigger noted
in spec Assumptions.

## Project Structure
See Technical Context; no new packages.
