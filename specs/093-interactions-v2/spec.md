# Feature Specification: Interaction system v2 — social primitives on the tool substrate

**Feature Branch**: `task-23-interactions-v2`

**Created**: 2026-07-29

**Status**: Design for ratification (TASK-23 design session; implementation is future
work, expected to be sliced into multiple tasks)

**Input**: TASK-23 + its ideation trail: 2026-07-22 re-grounding (primitives as
tool-registry entries per TASK-53, invoked through the TASK-52 loop — never a
bespoke parallel system), reorient 2026-07-26 move 12 (DF-pole drama generator;
chronicle rubric-legibility), guardian-directives ideation (reuse the spec 084
completion-predicate vocabulary for teach/order-shaped interactions).

> **Design-session artifact.** Ratifies the v2 interaction design. Operator review
> of this PR is the ratification act; **Open Questions** below are the parts most
> deserving of review pushback.

## The design in one paragraph

Five new social primitives — **argue, trade, teach, comfort, conspire** — join
talk_to as tool-registry entries with per-agent rosters, executed through the
existing agent tool loop and landing as recorded events through the normal doors.
Multi-party **scenes** generalize the hail protocol (form → conduct → dissolve)
so interactions survive movement and speed. Every interaction emits situated
memories with honest provenance (TASK-79 rules) and typed **relationship deltas**
that nightly consolidation folds into durable relationship state. LLM cost is
shaped by classing each primitive in the cognition registry (points/staleness
budgets), so the governor and cognition horizon meter social cognition exactly
like planning. The chronicle gets one digest-grammar entry per primitive, written
rubric-legible so future social exercises (curriculum) can grade off recorded
events, not prose.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Villagers do more together than talk (Priority: P1)

As a player, I want to watch villagers argue over a contested chest, trade wood
for cooked meals, teach each other where the quarry is, comfort the bereaved, and
conspire about the gru, so the village reads as a society with texture instead of
an endless string of chats.

**Acceptance Scenarios**:

1. **Given** two villagers with opposed claims (e.g. contested storage), **When**
   one's mind selects argue, **Then** an argue scene runs through the tool loop,
   lands as recorded events, moves both relationship states, and is narrated in
   the chronicle with its own digest entry.
2. **Given** a hungry woodcutter and a fed cook, **When** trade is selected and
   both sides' offers validate against REAL inventories at the executor door,
   **Then** goods move atomically (contested re-check semantics — reject whole,
   never partial), a debt entry is recorded when the exchange is lopsided by
   consent, and failure is loud (TASK-95's intent_failed pattern).
3. **Given** a villager who knows the quarry location and one who doesn't,
   **When** teach completes, **Then** the learner's mental map/knowledge gains
   exactly the taught fact through the existing knowledge-gating substrate
   (spec 041), with provenance "taught by <name>" (hearsay rules apply — taught
   facts are testimony, not observation, per TASK-79).
4. **Given** a villager whose kin died, **When** another selects comfort, **Then**
   the mourner's relationship to the comforter strengthens and grief-related
   need/mood effects apply per the tuning dials.
5. **Given** two villagers with a shared grievance and privacy, **When** conspire
   completes, **Then** a shared secret exists in the social fabric (existing
   secrets substrate) with both as holders, discoverable through the existing
   rumor machinery.

---

### User Story 2 - Scenes form and dissolve like the hail protocol, at any speed (Priority: P1)

As a villager in the game, when I begin an interaction, I want my counterpart(s)
held in the scene the way hails already pause talk_to targets, and the scene to
dissolve cleanly on completion, interruption (danger reflex), or timeout — so
interactions land at 8x+ instead of evaporating mid-walk.

**Acceptance Scenarios**:

1. **Given** any primitive, **When** the initiator's intent lands, **Then** scene
   formation reuses/extends the hail pause-close-found mechanics for 2..N
   participants; SURVIVAL reflexes and the DIRECTIVE rung interrupt scenes (the
   spec 062 arbitration order is unchanged).
2. **Given** a scene in progress, **When** a participant's danger band trips,
   **Then** the scene dissolves with a recorded interruption outcome — no
   participant is left paused (leak-proof: the TASK-109 pair-cooldown lesson).

---

### User Story 3 - Interactions become relationship memory, not transcript soup (Priority: P2)

As the nightly consolidation, I want each scene's outcome to arrive as a typed
relationship delta + situated memories (attribution-preserving gists), so
long-term relationships evolve from what happened, and dreams stay private
(TASK-99's per-agent constraint).

**Acceptance Scenarios**:

1. **Given** a completed scene, **Then** each participant holds their OWN situated
   memory (perspective-correct, provenance-tagged) and a typed delta
   (trust/warmth/debt/grievance axes — exact axes are an open question) lands in
   the social fabric through recorded events.
2. **Given** consolidation, **When** scenes are folded, **Then** folding operates
   strictly within one agent's store; taught facts and secrets keep provenance.

---

### User Story 4 - Social cognition is budgeted like all cognition (Priority: P2)

As the governor, I want each primitive classed in the cognition registry with its
own points/staleness budget (argue/conspire costlier than comfort), so drama
never starves survival planning and degrades gracefully under load.

**Acceptance Scenarios**:

1. **Given** saturation, **When** the governor sheds load, **Then** social classes
   shed before planner/survival classes per their configured priorities, and the
   damper (novelty gate) applies across ALL interaction kinds, not just talk.

## Requirements *(mandatory)*

### Functional Requirements (the ratified skeleton)

- **FR-001**: Primitives are tool-registry entries (world/expressive classes as
  appropriate) with per-agent rosters; no bespoke interaction engine.
- **FR-002**: Scenes generalize the hail protocol to 2..N participants with
  form/conduct/dissolve lifecycle, reflex interruption, timeout, and leak-proof
  pause bookkeeping.
- **FR-003**: All effects land as recorded events through existing doors
  (InjectSocial whitelist extended deliberately, per-type); replay-deterministic;
  loud failure per the TASK-95 pattern.
- **FR-004**: Trade validates both sides at the executor door against real
  inventories/carry rules (reject whole); consensual imbalance records a debt.
- **FR-005**: Teach transfers knowledge through spec 041 gating with testimony
  provenance; teach/order-shaped interactions reuse the spec 084
  completion-predicate vocabulary where a checkable goal exists.
- **FR-006**: Typed relationship deltas + per-participant situated memories feed
  consolidation; private-dreams constraint holds.
- **FR-007**: Each primitive is a cognition-registry class with budget dials;
  novelty/damper machinery covers all kinds.
- **FR-008**: Chronicle digest-grammar entry per primitive, rubric-legible
  (a future social exercise can grade from events alone).

### Open Questions (flagged for PR review — not yet decided)

- **OQ-1**: Relationship delta axes — reuse the existing relationship scalar(s)
  or introduce trust/warmth/debt/grievance as separate axes?
- **OQ-2**: Conspire's product — always a secret, or can it seed norm proposals
  (votes substrate) when the grievance is governance-shaped?
- **OQ-3**: Trade negotiation depth — single offer/accept/refuse per scene (lean)
  vs multi-round haggling (rich, costlier)? Lean is the working default.
- **OQ-4**: Scene size cap (working default: 4) and whether the meeting/curfew
  layer can host scenes.
- **OQ-5**: Does argue ever move need/mood state (morale), or only relationships?

## Success Criteria *(mandatory)*

- **SC-001**: This spec is ratified via PR review; the future implementation can
  be sliced (registry entries → scenes → deltas/consolidation → budget/chronicle)
  without another design session.
- **SC-002**: Every mechanism above rides an EXISTING substrate (registry, tool
  loop, hail, social fabric, spec 041/084, cognition registry, digest grammar) —
  no new parallel systems anywhere in the design.
- **SC-003**: The design satisfies replay determinism on its face: all state
  changes are recorded events; LLM output only ever selects among validated
  choices.

## Assumptions

- Ordering per the card: after TASK-157 (Done) — the directive substrate exists to
  reuse. TASK-27's ordering note is obsolete (Metatron v2 Done).
- Implementation tiering and slicing are decided when the implementation tasks are
  carded; expected Opus for scene/arbitration slices, Sonnet for registry/digest
  slices.
