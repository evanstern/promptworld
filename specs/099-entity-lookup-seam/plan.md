# Implementation Plan: Entity-lookup seam (TASK-76)

**Branch**: `task-76-entity-lookup-seam` | **Date**: 2026-07-29 | **Spec**: [spec.md](spec.md)

## Summary
Route every positional entity lookup through one accessor type (v1 = existing
scans, semantics-identical incl. tie-breaks); prove bit-identical replay; record
the store-error fatal-stands decision (D2) as a wiki note + site comment. No
index, no retry code.

## Technical Context
**Language**: Go. **Surfaces**: internal/sim (state.go, terrain.go, executor.go
call sites + rot sweep), one new accessor file; docs/wiki operational note.
**Testing**: determinism/replay harness, -race suite, grep-clean check.
**Constraints**: zero behavior change; tail task — merges last, droppable.

## Constitution Check
I–IV: PASS (spec 099 records D1/D2; one branch/PR; harness evidence; re-pins
in-branch). V: PASS — Sonnet (mechanical refactor; routine).

## Project Structure
internal/sim/lookup.go (accessor) + call-site edits; docs/wiki note for D2.
