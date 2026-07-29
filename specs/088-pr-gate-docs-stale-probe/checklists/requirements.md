# Specification Quality Checklist: merge-drift pr gate — docs-stale probe on all pinned sources + history moves

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-07-29
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — file:line anchors cite the defect site, not a prescribed fix
- [x] Focused on user value and business needs (gate integrity for reviewers/sessions)
- [x] Written for non-technical stakeholders (gate behavior in given/when/then terms)
- [x] All mandatory sections completed

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Success criteria are technology-agnostic
- [x] All acceptance scenarios are defined
- [x] Edge cases are identified
- [x] Scope is clearly bounded (session/worktree modes and TASK-165 finding class excluded)
- [x] Dependencies and assumptions identified (stateless history-move interpretation recorded)

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria
- [x] User scenarios cover primary flows
- [x] Feature meets measurable outcomes defined in Success Criteria
- [x] No implementation details leak into specification

## Notes

- Card AC3's "since the last probe" resolved as graph-detectable history moves
  (stateless gate) — recorded in Assumptions; flag at plan review if contested.
