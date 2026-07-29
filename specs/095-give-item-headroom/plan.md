# Implementation Plan: Carry-cap headroom guidance for give_item (TASK-167)

**Branch**: `task-167-give-item-headroom` | **Date**: 2026-07-29 | **Spec**: [spec.md](spec.md)

## Summary
Guidance/context-only change: extend the spec 059 miracle-capable prompt digest's
per-villager line with live carry headroom (free/cap), add one line to the
give_item gloss, leave the door byte-unchanged (FR-011 reject-whole preserved).

## Technical Context
**Language**: Go. **Surfaces**: internal/guardian (digest assembly),
internal/tool/derive.go (gloss). **Testing**: unit tests for digest arithmetic +
gloss presence + door regression; go test -race ./...
**Constraints**: no door/reducer/event changes; digest stays in its existing
per-villager line shape.

## Constitution Check
I–IV: PASS (spec 095 records the AC#1 decision with rejected alternatives; one
branch/PR; tests as evidence; wiki re-pins in-branch for touched sources —
guardian turn/digest and tool-registry notes likely NEEDS-REVIEW one-liners).
V: PASS — Sonnet (single-surface guidance change; routine). Escalation trigger:
none.

## Project Structure
internal/guardian/<digest assembly site>, internal/tool/derive.go, tests
alongside; wiki re-pins + player-docs probe as gated.
