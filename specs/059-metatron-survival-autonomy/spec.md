# Feature Specification: Metatron Survival Autonomy

**Feature Branch**: `059-metatron-survival-autonomy`

**Created**: 2026-07-25

**Status**: Draft

**Input**: TASK-111 — world-01 evidence: charges regenerated to cap and sat
unused while Ash starved (day 2) and Oak froze (day 6); the angel is
structurally turn-less (turns only on player chat or order match, and world-01
had almost no orders); 3 of its 4 miracles were door-rejected on invalid
coordinates because the turn prompt never includes positions/passability.
Decision (user, 2026-07-24): the angel ACTS on its own authority for survival —
not merely warn. Near-term slice of the agentization direction (TASK-112);
machinery built here must survive that redesign.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Survival watches exist from birth (Priority: P1)

Every world carries three standing system-origin survival watches — near-death,
starvation, exposure — that exist without any player action: seeded at world
creation, and back-seeded at boot for worlds that predate this feature (the
established seed-if-absent pattern). They are exempt from the player-order cap,
non-expiring, and cannot be cancelled by the player order surface (they are the
angel's nature, not a player configuration).

**Why this priority**: the structural fix — world-01's angel never turned
because nothing armed it; watches are the arming mechanism that already exists
(spec 029), minus the player dependency.

**Independent Test**: create a world → the three watches exist with system
origin; boot a pre-059 world → same; the player cap still admits its full
player-order count; cancel-by-player refuses.

**Acceptance Scenarios**:

1. **Given** a new world, **When** created, **Then** the log carries the three
   system survival watches (event-sourced, like all orders).
2. **Given** a pre-059 world booting after upgrade, **When** state carries no
   system survival watches, **Then** boot seeds them once (replay-safe seed
   pattern; no duplicates on later boots).
3. **Given** the three system watches standing, **When** the player places
   their full allowance of orders, **Then** the cap counts only player orders.
4. **Given** a player cancel naming a system watch, **When** processed, **Then**
   it refuses with an in-fiction message.

---

### User Story 2 - The angel acts on survival without permission (Priority: P1)

When a system survival watch matches (a villager near death, starving, or
freezing), the resulting turn is a SURVIVAL turn: the angel may send visions
and work miracles on its own authority — still charge-gated, still the same
tools — without player authorization. The initiative doctrine changes from
"never on your own initiative" to "on your own initiative ONLY for survival":
clock control, non-survival orders, and everything else stay player-authority.
The carve-out is visible in the angel's charter/turn framing so its narration
stays honest about what it may do.

**Why this priority**: the doctrine half of the fix — watches arming turns is
useless if the turn frame forbids acting. Co-P1 with US1; they ship together
or not at all.

**Independent Test**: drive a survival-watch match → the turn's frame permits
vision/miracle without player text; a non-survival turn keeps the restrictive
frame; clock tools remain refused on angel initiative in both.

**Acceptance Scenarios**:

1. **Given** a survival-watch turn and charges available, **When** the angel
   chooses a vision or miracle, **Then** it executes without player
   authorization (charge cost unchanged).
2. **Given** the same turn, **When** the angel attempts clock control or
   placing a non-survival order on its own initiative, **Then** the existing
   restriction holds.
3. **Given** a player-chat turn (non-survival), **When** framed, **Then** the
   initiative text is today's restrictive doctrine, unchanged.
4. **Given** a survival turn that acts, **When** the soul/chronicle records it,
   **Then** the record attributes the action to the survival duty (auditable
   authority trail).

---

### User Story 3 - Miracles can aim (Priority: P2)

Survival (and all) miracle-capable turns include a compact targeting digest:
villager names with positions and conditions, and passability guidance for the
relevant area, so miracles stop dying at the door on invalid coordinates
(world-01: 3 of 4 door-rejected).

**Why this priority**: an angel with authority but no aim still fails its
watch; P2 only because US1+US2 are the structural change and this is prompt
surface.

**Independent Test**: a survival turn's prompt contains the digest; a
miracle placed on a digest-listed position passes the door in tests.

**Acceptance Scenarios**:

1. **Given** a miracle-capable turn, **When** the prompt is assembled, **Then**
   it carries villager positions/conditions and passability guidance for the
   targeted area (compact, token-bounded).
2. **Given** the digest names a tile, **When** the angel targets it, **Then**
   the door accepts the coordinates (round-trip proven in a test).

---

### Edge Cases

- Multiple villagers in crisis at once: one survival turn may address any/all;
  watches don't queue duplicate turns for the same crisis while one is in
  flight (reuse the order-match debounce semantics if they exist; otherwise
  the turn loop's natural serialization bounds it).
- Zero charges at match time: the turn still happens (vision may be free? — if
  all tools are charge-gated and none affordable, the angel may only narrate;
  it must not burn the match silently — record the helpless turn).
- Player pause/paused world: survival turns respect the pause exactly like
  player-chat turns do today.
- System watches and the TTL bounds: non-expiring means exempt from the
  order-TTL validation, not a 10000-day hack — model it as origin-based
  exemption.
- TASK-112 survivability: watches are data (event-sourced orders with a system
  origin), the carve-out is turn-frame logic keyed on turn origin — both
  transfer to an agentized metatron; nothing here hard-codes the request-driven
  console model deeper.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Three system-origin survival watch orders (near-death,
  starvation, exposure) MUST exist in every world: seeded at world creation
  for new worlds and boot-seeded once for existing worlds lacking them
  (event-sourced, replay-deterministic, no duplicates across boots).
- **FR-002**: System watches MUST be exempt from the player-order cap and TTL
  expiry, and MUST refuse player cancellation, all keyed on order origin.
- **FR-003**: A turn triggered by a system survival watch MUST carry a
  survival-authority frame permitting visions and miracles without player
  authorization; charge costs unchanged.
- **FR-004**: Clock control and self-placed non-survival orders MUST remain
  outside angel initiative in ALL turn frames (the carve-out is survival-only).
- **FR-005**: Non-survival turns MUST keep today's restrictive initiative
  framing verbatim in spirit (no drive-by doctrine loosening).
- **FR-006**: Miracle-capable turn prompts MUST include a token-bounded
  villager positions/conditions + passability digest sufficient for the door
  to accept digest-derived coordinates.
- **FR-007**: Survival-turn actions MUST be attributable in the durable record
  (soul/chronicle) to the survival duty.
- **FR-008**: All new thresholds (what counts as near-death/starving/freezing
  for watch matching) MUST reuse existing doctrine constants where they exist;
  any NEW threshold introduced here MUST be promoted-dial-ready (documented
  const, single home) but NOT speculatively added to tuning.json (dials are
  earned by evidence).

### Key Entities

- **System survival watch**: a MetatronOrder with system origin; cap/TTL/cancel
  exemptions; three canonical instances.
- **Survival turn frame**: the turn prompt variant carrying the survival
  authority carve-out.
- **Targeting digest**: the positions/conditions/passability block in
  miracle-capable prompts.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: New and pre-059 worlds both show the three system watches
  standing after boot, proven by tests; replay determinism suites green.
- **SC-002**: In tests, a survival-watch match yields a turn that can execute a
  miracle with zero player input; the same tool call on a player-chat turn
  without authorization keeps today's refusal behavior.
- **SC-003**: Clock-control refusal on angel initiative proven in BOTH frames.
- **SC-004**: A miracle targeted from digest coordinates passes the door in
  tests (the world-01 3-of-4 rejection shape has a regression test).
- **SC-005**: Full suite green; no format_version bump.

## Assumptions

- The spec-029 order machinery (match → turn arming) works and is the vehicle;
  this spec adds origin semantics + frame carve-out + digest, not a new
  scheduler. If matching is currently player-order-only in a way that can't
  carry system origin, that's plan-phase discovery, not scope change.
- "Near-death, starvation, exposure" watch conditions reuse the needs/health
  thresholds the sim already defines (danger bands); no new tunable dials
  unless plan-phase finds none exist (then FR-008 applies).
- The digest is prompt surface only (no new events); its token budget rides
  the existing turn prompt budget discipline.
- Charter file wording changes ride this PR (the charter is versioned/observed
  via metatron.charter_observed — follow that mechanism, not raw edits).
