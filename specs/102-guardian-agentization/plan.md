# Implementation Plan: Guardian agentization (TASK-112)

**Branch**: `task-112-guardian-agentization` | **Date**: 2026-07-30 | **Spec**: [spec.md](spec.md)

## Summary
(Operator rename ruling 2026-07-30: serialized spellings are "steward" —
class/kind `steward`, dial `steward_cadence_ticks`, prefix
`steward-metatron-`; see spec.md. "Angel" below is design vocabulary.)
Move the guardian onto the shared agent construct: angel cognition-registry
class (cadence, budget, governor/horizon-gated) adding a scheduled lane beside
the unchanged event triggers; shared memory/consolidation (incl. spec 098);
charter-compiled persona with the operator-adopted incompetence ceiling as
default data; structural tutor/world channel split; guardrails and TASK-111
order machinery byte-contract-preserved; opt-in per world.

## Technical Context
**Language**: Go. **Surfaces**: internal/metatron+guardian (loop/persona),
internal/mind (shared construct reuse), internal/cognition (angel class),
internal/sim (config/telemetry, doors unchanged), internal/tui (decision
trail), docs. **Testing**: guardrail suite, ceiling enforcement, channel
isolation, budget/shedding, replay byte-identity, multi-day measure-world soak.
**Constraints**: existing doors only; additive events; opt-in compat; playtest
untouched.

## Constitution Check
I–IV: PASS (spec 102 records the firm direction + both operator rulings;
one branch/PR; soak + suite evidence; re-pins in-branch — guardian/mind/
cognition notes expected heavy NEEDS-REVIEW).
V: PASS — **Opus** (card-stated). Recorded on the board task.

## Reuse checklist (SC-004 — reviewed at gate)
mind loop driver · working-memory window · memory store + consolidation entry
(incl. dream phase) · persona compilation seam · cadence/decision-class
registration · decision-trail plumbing. Each: shared code path named in the
implementation report, no forked copies.

## Project Structure
Evidence under docs/design/evidence/task-112/; new wiki note
guardian-agentization.md; dials/config per spec 048 pattern.
