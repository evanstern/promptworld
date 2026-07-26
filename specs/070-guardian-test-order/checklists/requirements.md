# Specification Quality Checklist: Guardian worker shutdown joins

**Purpose**: Validate specification completeness and quality before planning
**Created**: 2026-07-26
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] Implementation detail limited to the pinned diagnosis + decided fix shape (the defect is code-level by nature)
- [x] Focused on the value: a trustworthy test suite
- [x] All mandatory sections completed

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers (tier checkpoint resolved by operator 2026-07-26: Sonnet)
- [x] Requirements testable (count/race matrices)
- [x] Edge cases identified (in-flight job at Close, double Close, future workers)
- [x] Scope bounded (pre-Close enqueue nondeterminism explicitly out)
- [x] Dependencies identified (spec 069 gate interaction; guardian.md re-pin)

## Feature Readiness

- [x] FRs map to acceptance scenarios and SCs
- [x] Success criteria measurable

## Notes

- Near-trivial by constitution standards now that the diagnosis is pinned;
  kept as a compact spec because spec 070 was already claimed and the shutdown
  semantics change deserves a decision record.
