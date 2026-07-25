# Feature Specification: Chronicle jump-to-source + input-parity retrofit start

**Feature Branch**: `049-chronicle-jump-to-source`

**Created**: 2026-07-25

**Status**: Draft

**Input**: User description: "Chronicle jump-to-source + input-parity retrofit start (board TASK-124; reorient 2026-07-25 D3 + decision 8, Wave 2 quick win). Fill the chronicle inspect-mode reserved ⏎ seam: pressing ⏎ (and clicking a chronicle line, per input parity) on a selected chronicle event whose subject has a map location centers the map camera on that subject; events without a locatable subject are honest no-ops with a visible hint. Ratify the input-parity doctrine in patterns/keymap.md; new bindings documented keys+mouse. Design authority: docs/design/tui/ — same-PR doc amendment gate applies."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Jump from an event to its subject on the map (Priority: P1)

A player reading the chronicle while paused sees an event about a villager
("Ash foraged at (14,9)", "Birch died: exposure") and wants to see where that
happened. They press `⏎` on the selected event and the map camera centers on
the event's subject — no manual arrow-key panning, no scanning the map for a
glyph.

**Why this priority**: This is the reorientation's D3 quick win — the single
biggest map-legibility gap named by the synthesis (RimWorld letter / DF
zoom-to-event precedent). Everything else in this feature hangs off this
action existing.

**Independent Test**: Pause a world with at least one located villager, select
a chronicle event naming that villager, press `⏎`, observe the map camera now
centered on the villager's position.

**Acceptance Scenarios**:

1. **Given** the clock is paused and a chronicle event whose subject is a
   living, located villager is selected, **When** the player presses `⏎`,
   **Then** the map camera centers on that villager's current position and
   camera auto-follow is suspended (exactly as manual panning suspends it).
2. **Given** the same state, **When** the player presses `c` after a jump,
   **Then** the camera resumes its normal follow behavior (existing recenter
   semantics, unchanged).
3. **Given** a selected event that names no locatable subject (e.g. a
   world-lifecycle or LLM-telemetry event), **When** the player presses `⏎`,
   **Then** nothing moves and a visible hint states the event has no location
   — never a silent no-op, never an error.
4. **Given** a selected event whose subject is dead or despawned but whose
   payload carries an explicit position, **When** the player presses `⏎`,
   **Then** the camera centers on that recorded position (the place it
   happened), not on a stale or missing entity.

---

### User Story 2 - Click a chronicle line to select and jump (Priority: P2)

The same action is reachable by mouse: clicking a chronicle event line while
paused selects that line and performs the same jump — the first mouse-bound
control in the app, opening the input-parity rollout.

**Why this priority**: Decision 8 ratifies input parity as doctrine; D3 names
click-a-line as part of the jump feature itself ("click-a-line too"). It ships
with US1 but is separable — keyboard-only delivery would still be a viable
(if doctrine-incomplete) MVP.

**Independent Test**: Pause a world, click a chronicle event line with a
located subject, observe selection moving to that line and the camera
centering on the subject.

**Acceptance Scenarios**:

1. **Given** the clock is paused and the chronicle is visible, **When** the
   player clicks an event line, **Then** the selection moves to that line and
   the jump behavior of US1 fires for it (locatable → center; not → hint).
2. **Given** the clock is running, **When** the player clicks a chronicle
   line, **Then** nothing happens — reading closely is a paused-mode activity
   (existing doctrine), and the running feed's lines move too fast to be
   honest click targets.
3. **Given** any existing keyboard flow, **When** mouse support is enabled,
   **Then** every existing keyboard behavior is unchanged — keyboard remains
   primary and complete.

---

### User Story 3 - The seam advertises itself honestly (Priority: P3)

The detail pane's reserved actions slot (`[future: actions]`) becomes a real
actions bar: when the selected event is locatable it shows the jump affordance
(e.g. `⏎ jump to Ash (14,9)`); when it isn't, it shows the honest absence
(e.g. `no location for this event`). A player who presses `⏎` never wonders
whether the app is broken.

**Why this priority**: Discoverability polish on top of US1 — the corpus's
"reserved seams are wired to a documented no-op" rule, upgraded to "wired to a
visible affordance". Valuable but the jump works without it.

**Independent Test**: Pause, select a locatable event, observe the affordance
text in the detail pane; select an unlocatable one, observe the absence text.

**Acceptance Scenarios**:

1. **Given** inspect mode with a locatable event selected, **When** the detail
   pane renders, **Then** its actions slot names the jump key and the resolved
   subject/position.
2. **Given** an unlocatable event selected, **When** the detail pane renders,
   **Then** the actions slot states there is no location — and this same text
   is the US1 scenario-3 hint (one surface, not two).

---

### Edge Cases

- **Multi-agent events** (conversation turns Ash→Rowan, social events with
  several participants): the jump targets the event's primary actor (speaker /
  initiator) — one deterministic subject per event, documented in the design
  page.
- **Narrow-terminal fallback** (map and chronicle are separate views): the
  jump still centers the camera; the player lands on the map view so the
  effect is visible, not invisible behind another pane.
- **Subject position off the current camera viewport edge / map edge**:
  centering clamps to the map's existing camera bounds (same clamping as
  manual panning; no new camera math).
- **Oversized payloads** (`world.migrated` embedding full state): unlocatable
  → honest hint; the actions bar must not scan the entire payload to decide
  (bounded work, same windowing discipline the detail pane already has).
- **Event references an agent index that no longer resolves** in the replica
  (post-migration, morgue): fall back to any explicit payload position; if
  none, unlocatable hint.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: In inspect mode (clock paused, chronicle visible), `⏎` on the
  selected chronicle event MUST center the map camera on the event's resolved
  subject location, suspending camera auto-follow exactly as manual panning
  does.
- **FR-002**: Subject resolution MUST be deterministic and two-step: (1) the
  event's primary actor's current position in the live world state, if that
  actor exists and is located; else (2) an explicit position carried in the
  event payload, if present; else the event is unlocatable.
- **FR-003**: `⏎` on an unlocatable event MUST be a no-op accompanied by a
  visible hint stating the event has no location. No silent no-ops, no errors.
- **FR-004**: Clicking a chronicle event line while paused MUST select that
  line and then apply FR-001/FR-003 to it. Clicking while the clock runs MUST
  do nothing.
- **FR-005**: Mouse enablement MUST NOT change any existing keyboard behavior;
  the keyboard alone MUST still reach 100% of app functionality (parity
  doctrine rule 2).
- **FR-006**: The detail pane's reserved actions slot MUST render the live
  affordance for the selected event: the jump binding plus resolved
  subject/position when locatable, the no-location text when not.
- **FR-007**: In the narrow-terminal fallback, a successful jump MUST leave
  the player looking at the map view centered on the subject.
- **FR-008**: The design reference MUST be amended in the same PR:
  `panels/chronicle.md`'s jump-to-source control-table row flips from
  `unbuilt` to real renderer symbols with its keys+mouse binding recorded;
  `patterns/keymap.md`'s inspect-mode `⏎` row flips from "reserved — no-op"
  to the jump action; affected pages' parity-rollout notes and
  `verified_against` pins are updated (gate rules 1–3).
- **FR-009**: The input-parity doctrine already authored in
  `patterns/keymap.md` MUST be ratified by this feature shipping its first
  compliant control: the jump lands keyboard+mouse together, and its
  control-table row is the corpus's first row with a real mouse target.
- **FR-010**: The linear-stream/CLI projection (D1) loses nothing: the jump is
  pure navigation over information the streams already carry (event payloads
  include positions); no new TUI-only information is introduced.

### Key Entities

- **Chronicle event**: stored world event (`seq`, `tick`, `type`, `payload`);
  the jump's input. Payloads name agents by index and often carry explicit
  positions.
- **Subject**: the event's single resolved jump target — a primary actor with
  a live position, or an explicit payload position.
- **Detail-pane actions bar**: the previously-reserved slot that now renders
  the per-event affordance or its honest absence.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: From any chronicle event about a located villager, the player
  reaches that villager on the map in exactly one action (one keypress or one
  click), replacing today's unbounded manual panning.
- **SC-002**: 100% of chronicle events respond to the jump action with either
  a camera move or a visible no-location hint — zero silent outcomes, verified
  over a representative event-catalog sweep.
- **SC-003**: A keyboard-only player retains access to 100% of the feature
  (parity rule 2 holds; mouse adds no exclusive capability).
- **SC-004**: The design-reference gate (`check-tui-design.mjs --changed`)
  passes on the feature's PR with the affected pages amended and re-pinned in
  that same PR.

## Assumptions

- Clicking while the clock runs is deliberately inert (existing "pausing is
  the way to read closely" doctrine); no click-to-pause shortcut in this
  feature.
- The primary actor of a multi-participant event is the speaker/initiator the
  event type already distinguishes; no new event fields are needed.
- Camera centering reuses existing camera/clamping machinery; this feature
  adds no new camera math beyond "center on (x,y)".
- Mouse support is enabled app-wide by this feature but bound only to the
  chronicle line click; other panels' mouse targets arrive incrementally per
  the parity-rollout notes (doctrine rule 3).
- Dependency: TASK-123 (spec 047) is merged — the design pages this feature
  amends exist at `status: shipped`/`specified` with the reserved seam
  documented.
