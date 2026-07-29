# Implementation Plan: Determinism scope + reducer-constants doctrine (TASK-75)

**Branch**: `task-75-determinism-doctrine` | **Date**: 2026-07-29 | **Spec**: [spec.md](spec.md)

## Summary
Docs/doctrine change: per-log-not-per-seed determinism limit corrected across
deterministic-rng.md, the EffectiveRate-owning note, and README; reducer-constants
doctrine (emitter-computes default; re-derive exception requires format-version
bump + migration, TASK-134 pointer) recorded in the owning wiki note with a full
audit list of re-derive sites; comment-only annotations at those sites.

## Technical Context
**Language**: Markdown + Go comments only (zero behavior change).
**Testing**: go test ./... green (unchanged behavior); wiki freshness gate;
player-docs probe (README is a pinned input); merge-drift pr gate.

## Constitution Check
I–IV: PASS (spec 092; one branch/PR; gate runs as evidence; re-pins in-branch).
V: PASS — Sonnet (doc reconciliation; routine).

## Project Structure
docs/wiki/deterministic-rng.md, sim-loop/EffectiveRate note, event-log or
sim-state-reducer note (doctrine home), README.md, internal/sim comment sites,
docs/player regeneration as probed.
