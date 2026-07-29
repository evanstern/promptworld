# Feature Specification: Loud failure for non-build goals — agent.intent_failed

**Feature Branch**: `task-95-intent-failed-loud`

**Created**: 2026-07-29

**Status**: Draft

**Input**: TASK-95 — follow-up from TASK-91 / spec 038 (research D5): the
silent-failure pattern fixed for the seven build_* goals is systemic. Invalid exits
for forage/chop/hunt/demolish/repair/quarry/cook/bathe and the contested no-op
re-checks (craft/cook/bathe/deposit/withdraw in the executor) still resolve as a
bare agent.intent_done indistinguishable from success, with no failure memory.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A villager knows the hunt failed (Priority: P1)

As a villager in the game, when my hunt or forage or repair evaporates because the
target was gone or the resource was contested away mid-walk, I want a real failure
event and a situated memory of it — so my beliefs can be falsified by experience
(the phantom-wall lesson from TASK-91) instead of my mind reading a bare
"intent done" as success.

**Acceptance Scenarios**:

1. **Given** any non-build work goal whose completion re-validation fails (target
   absent/invalid at arrival or contested away at the no-op re-check), **When**
   the executor resolves it, **Then** a distinct failure event (generalized
   `agent.intent_failed`, following the `agent.build_failed` shape: goal, reason,
   position) is emitted instead of a bare `agent.intent_done`, plus a situated
   failure memory for the agent.
2. **Given** a successful completion, **Then** behavior is byte-identical to
   today (events, yields, memories unchanged).
3. **Given** the failure event, **When** the mind's re-arm list processes it,
   **Then** the agent re-arms exactly as `build_failed` already does (no stuck
   intent, no double re-arm).

---

### User Story 2 - The chronicle and player can see the failure (Priority: P2)

As a player reading the story feed, I want failed intents narrated distinctly, so
"tried to hunt and found the den empty" reads differently from a completed hunt.

**Acceptance Scenarios**:

1. **Given** `agent.intent_failed`, **Then** the TUI digest grammar renders it
   (no `TestCatalogSweep` regression — every new type has a digest entry) and
   event-types.md documents it.

---

### Edge Cases

- **Replay/versioning (spec 094 doctrine)**: adding a NEW event type is additive —
  old logs never contain it, no format-version bump required (the doctrine's bump
  triggers are renames and reducer-re-derivation changes, neither applies).
  Existing replay expected-event tests updated only where fixtures newly exercise
  failure paths.
- Goals covered: the card's list — forage, chop, hunt, demolish, repair, quarry,
  cook, bathe invalid exits; craft, cook, bathe, deposit, withdraw contested
  re-checks. build_* goals keep their existing `agent.build_failed` (no rename,
  no consolidation — spec 094 just froze vocabulary churn; a build_failed→
  intent_failed migration is explicitly OUT of scope).
- Reason strings: deterministic, enumerated (target-gone / contested / invalid),
  payload-carried (emitter-computes; the reducer stamps outcome state without
  re-deriving anything).
- Working-memory pressure: failure memories use the same salience shape as
  build-failure memories (no new flooding vector).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: A generalized `agent.intent_failed` event (goal, enumerated reason,
  position; emitter-computes) emitted for every invalid/contested non-build
  resolution the card enumerates; `agent.intent_done` remains success-only for
  these goals.
- **FR-002**: Situated failure memory per occurrence, same provenance/salience
  discipline as build failures.
- **FR-003**: Reducer arm applies the event (intent ring closure like
  build_failed via the existing stampIntentOutcome path); mind re-arm list
  includes the new type.
- **FR-004**: TUI digest grammar entry + event-types.md documentation;
  TestCatalogSweep green.
- **FR-005**: Replay byte-identity for existing logs; regression tests cover at
  least one gather goal (hunt or forage) and one station goal (cook or
  deposit/withdraw) per card AC#2, both invalid-exit and contested variants.
- **FR-006**: No behavior change to successful completions or build_* goals.

## Success Criteria *(mandatory)*

- **SC-001**: For every enumerated goal, the invalid and contested paths emit
  intent_failed (test matrix), and no path resolves as a bare intent_done that
  isn't a success.
- **SC-002**: Existing replay fixtures replay byte-identically.
- **SC-003**: TestCatalogSweep green; event-types.md lists the new type with its
  reasons.

## Assumptions

- The `agent.build_failed` machinery (emission shape, memory constructor,
  stampIntentOutcome closure, re-arm) is the pattern to generalize; blast radius
  is the card's list: replay expected-event sets, TUI digest, event-types.md,
  mind re-arm list.
- Tier: Sonnet — pattern extension alongside tests; escalation trigger: if the
  intent-ring closure demands new reducer semantics beyond the existing
  stampIntentOutcome path.
