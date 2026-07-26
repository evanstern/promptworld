# Specification Quality Checklist: Report-card truth — unify all card surfaces on sim.EvaluateRubric

**Purpose**: Validate specification completeness and quality before planning
**Created**: 2026-07-26
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] Implementation detail limited to the pinned problem sites and decided shapes (the
  defect and its fix are code-level by nature; file:line references are the diagnosis,
  verified against HEAD)
- [x] Focused on the value: truthful grading at the game's most salient teaching moment
- [x] All mandatory sections completed

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers — the direction is ratified (reorient 2026-07-26
  decision 1, merged position 2); resolver precedence, aged-out fallback, and zero-value
  semantics decided from existing doctrine (spec 063 re-read-never-re-grade, "unknown
  honestly") and recorded in the spec
- [x] Requirements testable (regression on the exact motivating case; rubric table
  matrices; replay equivalence; gate exit codes)
- [x] Edge cases identified (pre-flag snapshots, old-log replay, aged-out pass,
  non-evaluator exercises, post-pass death, ambient worlds, nil replica)
- [x] Scope bounded (the-law pass emission out — FR-009; semantic lint out — FR-011 /
  TASK-150; `reportCardView` renderer and console seam unchanged — FR-005)
- [x] Dependencies identified (TASK-144/spec 070 merged; TASK-67 duel depends on this
  landing first; wiki-in-PR gate obligations enumerated)

## Feature Readiness

- [x] FRs map 1:1 onto the three board ACs (AC1 ↔ FR-001..005, AC2 ↔ FR-006..009,
  AC3 ↔ FR-010..011) and onto tasks.md phases 3/2+4/5
- [x] Success criteria measurable (SC-001..005)

## Notes

- Tier: Opus 4.8 per the board record (cross-package sim state/reducer + all TUI card
  surfaces; doctrine-adjacent). One-way escalation already resolved at board-authoring
  time; no checkpoint pending.
