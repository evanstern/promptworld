# Implementation Plan: Move-miracle target freshness (TASK-166)

**Branch**: `task-166-move-miracle-freshness` | **Date**: 2026-07-29 | **Spec**: [spec.md](spec.md)

*Plan was authored folded into spec.md's Decision section at specify time; this
file records it separately for the spec-artifact contract. Content matches what
shipped (PR #133, merged).*

## Summary
Door-side name re-resolution for villager moves (decision (a)): guardian move path
in internal/guardian/turn.go resolves the named villager's live position at
validation/emission; x/y advisory; applyEntityMoved untouched; guidance gloss in
internal/tool/derive.go.

## Technical Context
**Language**: Go. **Testing**: go test -race; replay byte-identity regression.
**Constraints**: emitter-computes (recorded event carries resolved coords);
validation precedes charge; no reducer change; no format_version bump.

## Constitution Check
I–IV: PASS (spec 091 + card AC1 decision recorded; one branch/PR; probe + tests
as evidence; wiki re-pins in-branch — 3 prose-amended, 13 re-pinned).
V: PASS — Sonnet (narrow door fix); escalation trigger (reducer semantics change)
did not fire.

## Project Structure
internal/guardian/turn.go (landMiracle move arm), internal/tool/derive.go (gloss),
internal/guardian/move_freshness_test.go, internal/sim/miracles_test.go (replay
regression), docs/design/evidence/task-166/ (live probe).
