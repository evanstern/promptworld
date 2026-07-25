# Specification Quality Checklist: Needs-Conditioned Recovery Intents

**Purpose**: Validate specification completeness before planning
**Created**: 2026-07-25
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] Implementation named only where it is decided design (condition-on-intent, warm_up)
- [x] Focused on outcome (recovery completes on the need; no arrive-idle vacuum)
- [x] All mandatory sections completed

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers — Direction B decision + 057 audit Gap C
      + 062's shipped substrate resolve the space
- [x] Requirements testable and unambiguous
- [x] Success criteria measurable (recover-then-release; extended Sage scenario; Oak-night wake)
- [x] Edge cases identified (abort/preemption/staleness — no new immunity; opt-in
      condition; replay/snapshot compat; reflex-parity change enumerated as intended)
- [x] Scope bounded (no prompt work; no new verbs beyond warm_up; TUI expected no-op)
- [x] Dependencies and assumptions identified

## Feature Readiness

- [x] All FRs have acceptance coverage
- [x] Board AC map: #1→US1+US2/FR-002/005, #2→SC-002, #3→SC-001

## Notes

- US3 (interruptibility) is deliberately co-P1: an un-interruptible loiter
  re-creates death-by-neglect (TASK-133) from the other side.
