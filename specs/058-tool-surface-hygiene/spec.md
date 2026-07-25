# Feature Specification: Tool Surface Hygiene (clamp expressive text; prune dead verbs)

**Feature Branch**: `058-tool-surface-hygiene`

**Created**: 2026-07-25

**Status**: Draft

**Input**: TASK-110 — diagnosis (2026-07-24, world-01 event log): 807
rejected_malformed tool calls, ~93% "exceeds text cap", across ALL tiers
including cloud (muse 403, talk_to 308, plus reason-cap overruns). This is cap
design, not model grammar: every model writes past 200-rune caps. Decisions
recorded on the task: truncate-with-notice for EXPRESSIVE fields only; clamp
set_plan to its step cap instead of rejecting; keep strict rejection for
structural failures; conversation STAYS on gemma (no re-route, no tool_mode
change); roster shrinks — collect_water and bathe leave the villager loop
roster (water has no consumer), revisit only if a thirst need is designed.

**Reality check (2026-07-25, sweep grounding)**: world-01's numbers predate the
current registry (talk_to today carries no text param). The expressive text
surfaces TODAY: `say` (300-byte cap), `gist` (200-byte), `muse` (200-rune), and
the optional per-action `reason` (200-rune, `ReasonCapRunes`). The rejection
site is the tool-loop arg validation (`internal/toolloop/loop.go`, Text kind);
the set_plan step rejection is the landing guard (`internal/sim/landing.go`,
`PlanStepCap`). The spec targets those surfaces as they exist now.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Expressive overruns are clamped, not wasted (Priority: P1)

A model writes a 350-rune musing against a 200-rune cap. Instead of the whole
tool call being rejected (a wasted LLM turn and a villager that does nothing),
the text is truncated rune-safely to the cap, the action proceeds, and the
truncation is visible: the model's tool-result feedback says the text was cut,
and telemetry distinguishes a clamped acceptance from a clean one.

**Why this priority**: ~93% of 807 rejections were this shape — the single
biggest source of wasted model turns and villager inaction.

**Independent Test**: drive the tool loop with over-cap text on each expressive
field; the call lands (event emitted with clamped text), feedback names the
clamp, telemetry records it.

**Acceptance Scenarios**:

1. **Given** a muse/say/gist call whose text exceeds its cap, **When** the loop
   validates it, **Then** the text is truncated at the cap (rune-safe — never
   splitting a UTF-8 sequence, byte caps included), the call proceeds, and the
   emitted event carries the clamped text.
2. **Given** an acting world tool whose optional `reason` exceeds its cap,
   **When** validated, **Then** the reason is clamped and the action proceeds.
3. **Given** any clamped call, **When** the model receives its tool result,
   **Then** the result names the field and the clamp (so the model can adapt),
   and the call's telemetry verdict is distinguishable from a clean acceptance.
4. **Given** a structural failure (unknown tool, missing required arg, wrong
   type, unknown enum value, non-string text), **When** validated, **Then** it
   is REJECTED exactly as today — clamping never applies to structure.

---

### User Story 2 - Oversized plans are clamped to the step cap (Priority: P1)

A model submits a 5-step plan against the 3-step cap. The first `PlanStepCap`
steps are accepted (with the same clamp visibility as US1) instead of the
whole plan being rejected.

**Why this priority**: 34 world-01 rejections; same wasted-turn shape as US1.

**Independent Test**: set_plan with >cap steps lands with exactly the first cap
steps; the feedback names the clamp.

**Acceptance Scenarios**:

1. **Given** a set_plan with more than `PlanStepCap` steps, **When** it lands,
   **Then** the plan in effect is the first `PlanStepCap` steps and the tool
   result says the plan was clamped.
2. **Given** a set_plan whose steps contain a structural failure (unknown verb,
   malformed step), **When** validated, **Then** it rejects as today.

---

### User Story 3 - Dead verbs leave the villager roster (Priority: P2)

`collect_water` and `bathe` no longer appear in the villager loop roster: the
model never sees them, never spends turns on them (water has no consumer —
collection fell 72/day → 0 once novelty wore off; bathe used once in 6 days).
The world-side machinery stays (the verbs remain executable by the sim; only
the villager-facing tool surface shrinks), so reintroduction is a
roster/gloss change, not a rebuild. Prose glosses that advertise the pruned
verbs move in the same change.

**Why this priority**: prompt-surface hygiene — every dead verb is tokens and
model attention spent on a non-choice. P2 because it's small and independent.

**Acceptance Scenarios**:

1. **Given** the villager loop roster, **When** enumerated, **Then**
   collect_water and bathe are absent, and the roster's gloss text no longer
   advertises them.
2. **Given** replay of a world whose log contains historical collect_water /
   bathe intents, **When** replayed, **Then** it reproduces exactly (the
   executor still honors the verbs).
3. **Given** the code, **When** read at the prune site, **Then** a comment
   records the revisit condition (a designed thirst need) per the task
   decision.

---

### Edge Cases

- Clamping exactly at the cap boundary mid-rune (byte caps): truncate to the
  last whole rune below the cap; never emit invalid UTF-8.
- Empty-after-clamp is impossible (caps ≥ 200); no special case.
- Metatron/expressive tools outside the villager roster (vision/miracle text,
  narrate): OUT of scope — their caps and rejection behavior are unchanged
  (villager surfaces only, per the task).
- Replay compatibility: events always carried post-validation text; clamped
  text is what lands in events, so replay determinism is untouched. Historical
  logs with rejected calls replay as before (rejections produced no world
  events).
- The conversation route and tool_mode: explicitly untouched (task decision).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Tool-loop text validation MUST clamp (rune-safe) instead of
  reject when a Text argument of a villager-roster tool exceeds MaxRunes or
  MaxBytes, for expressive fields: say.text, gist (muse's), muse.text, and the
  shared optional `reason`. The clamp MUST be visible in the model-facing tool
  result and in telemetry as a distinct verdict/annotation.
- **FR-002**: Structural validation (unknown tool, missing required arg, type
  mismatches, enum violations) MUST keep rejecting exactly as today.
- **FR-003**: set_plan MUST accept the first `PlanStepCap` steps of an
  oversized plan (clamp-with-notice) instead of rejecting; structurally
  invalid steps still reject.
- **FR-004**: `collect_water` and `bathe` MUST leave the villager loop roster
  and its glosses; the sim executor MUST keep honoring both verbs (replay +
  possible future reintroduction); the revisit condition MUST be recorded in a
  code comment at the prune site.
- **FR-005**: Events MUST carry the clamped (post-validation) text so replay
  and downstream consumers see exactly what was accepted.
- **FR-006**: No re-route and no tool_mode change for conversation (recorded
  decision — the spec forbids drive-by "improvements" here).

### Key Entities

- **Expressive field**: a Text param whose content is prose for humans/memory
  (say.text, gist, muse.text, reason) as opposed to structural args.
- **Clamp notice**: the model-facing tool-result annotation + telemetry
  verdict distinguishing clamped-accepted from clean-accepted.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Over-cap expressive text on every villager expressive field lands
  as a clamped acceptance in tests; zero rejections from length alone on those
  fields.
- **SC-002**: Over-cap set_plan lands with exactly PlanStepCap steps in tests.
- **SC-003**: All structural rejection tests (existing suite) pass unchanged.
- **SC-004**: The villager roster enumeration test shows collect_water/bathe
  absent; existing executor tests for both verbs still pass (machinery kept).
- **SC-005**: A clamped call is distinguishable from a clean one in telemetry
  in tests (the world-01 measurement of clamp rates becomes possible).

## Assumptions

- "Expressive" is defined by enumeration (the four fields), not by a new
  param attribute — unless the implementer finds an existing natural flag
  (e.g. Effect==Expressive + reason) that covers exactly these; either is
  acceptable if the covered set is exactly the four.
- talk_to needs no clamp work today (no text param); the world-01 talk_to
  rejections were against a prior registry shape.
- Metatron surfaces excluded; meeting/norm text (NormTextMax truncation in
  mind/meeting.go) already clamps and is untouched.
- The roster prune does not remove the oven/bathe world machinery or recipes;
  only the villager tool surface and glosses.
