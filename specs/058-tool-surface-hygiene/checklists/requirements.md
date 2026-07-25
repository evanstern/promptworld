# Specification Quality Checklist: Tool Surface Hygiene

**Purpose**: Validate specification completeness and quality before planning
**Created**: 2026-07-25
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] Implementation named only where it IS the product surface (tool/param names)
- [x] Focused on user value (wasted model turns, prompt hygiene)
- [x] All mandatory sections completed

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements testable and unambiguous
- [x] Success criteria measurable
- [x] Edge cases identified (rune-safe boundaries, replay, out-of-scope surfaces)
- [x] Scope clearly bounded (villager surfaces only; conversation route frozen)
- [x] Dependencies and assumptions identified

## Feature Readiness

- [x] All FRs have acceptance coverage
- [x] Scenarios cover primary flows

## Notes

- Ambiguities resolved as assumptions: expressive set defined by enumeration
  (four fields); talk_to excluded (no text param today); metatron surfaces
  untouched. World-01 counts reconciled against the current registry in the
  spec's Reality check block.
