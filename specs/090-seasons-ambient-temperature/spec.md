# Feature Specification: Seasons and ambient temperature

**Feature Branch**: `task-28-seasons-ambient-temperature`

**Created**: 2026-07-29

**Status**: Design ratified via PR review (TASK-28 design session; implementation is future work)

**Input**: TASK-28 — "Seasons and ambient temperature: design session" + the card's
recorded pre-session decisions (operator, 2026-07-20) and re-grounding notes
(2026-07-22/23, reorient 2026-07-26).

> **Design-session artifact.** This spec is the deliverable of TASK-28: it ratifies
> the seasons/temperature design so a future implementation task can build from it.
> No code ships with this spec. Constants below are DESIGN TARGETS for the
> implementing task and the TASK-30 calibration worksheet, not pinned code.

## Pre-decided constraints (operator, recorded on the card — not re-litigated here)

1. **Two seasons only** — hot and cold, alternating; no four-season calendar.
2. **Diurnal swing is in** — temperature drops at night, peaks 13:00–14:00,
   troughs pre-dawn 04:00–05:00.
3. **Purpose is decision-3** — seasons exist to turn the labor-budget screw
   (strife doctrine), not for flavor.
4. **Hard invariant** — temperature is derivable from (seed, tick) with no new
   persisted state beyond events; replay determinism is test-enforced.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The year turns and the village feels it (Priority: P1)

As a player watching a long run, I want the world to move through alternating hot
and cold seasons — visible in the chronicle and on the status surfaces — so that
time has texture and the village's fortunes rise and fall with the calendar rather
than repeating one eternal day.

**Why this priority**: the calendar is the substrate every other part of this design
(temperature curve, scarcity, incidents) hangs off.

**Acceptance Scenarios**:

1. **Given** a world at default season length (10 game-days), **When** day 11
   begins, **Then** the season has flipped (hot→cold), a season-transition event is
   in the log, and the chronicle narrates the turn.
2. **Given** the same seed and log, **When** the world replays, **Then** every
   season boundary lands on the identical tick (pure function of tick — season
   index = day ÷ length, alternating).
3. **Given** TASK-14-shaped 30-day runs, **When** run at default length, **Then**
   the run crosses 3 season transitions (10-day length chosen exactly for this).

---

### User Story 2 - Cold is a curve, not a switch (Priority: P1)

As a villager in the game, I want how fast I lose warmth to follow a continuous
ambient temperature — coldest just before dawn, warmest in early afternoon, colder
across the whole cold season — so my survival decisions (when to travel, when to
refuel, when to sleep) have a landscape to be smart about, instead of the current
binary "night = −4/min, day = +2/min".

**Why this priority**: replaces the binary night-cold with the mechanism the whole
design exists to deliver; the labor screw (US3, TASK-30) is calibrated against it.

**Acceptance Scenarios**:

1. **Given** any (seed, tick), **When** ambient temperature is computed, **Then** it
   equals seasonal baseline + diurnal swing with trough at 04:00–05:00 and peak at
   13:00–14:00, and two machines computing it for the same (seed, tick) agree.
2. **Given** an agent outdoors with no fire, **When** ambient is below comfort,
   **Then** warmth decays proportional to the gap (clamped to a min/max rate);
   **When** ambient is at/above comfort, warmth recovers — reproducing today's
   hot-season daytime behavior as the calibration anchor.
3. **Given** fire/shelter/sleep, **When** warmth is evaluated, **Then** their
   modifiers apply on top of the ambient gap (fire and shelter as effective-ambient
   boosts; sleep per the existing sleep rules) — spec 012's fuel mechanics
   (2 wood → 8 game-hours, +4h/refuel, cap 12h) are unchanged by this design.
4. **Given** the hot-season default dials, **When** a pre-seasons world's behavior
   is compared, **Then** survival outcomes are approximately today's (compat
   anchor: hot season ≈ current world, cold season is the new pressure).

---

### User Story 3 - Winter is the ant-and-grasshopper screw (Priority: P2)

As a player, I want the cold season to slow or stop forage regrowth and thin den
yields, so surviving winter requires warm-season stockpiles — making storage
(spec 013 chests/piles), sharing, and planning ahead the difference between a
village that endures and one that starves.

**Why this priority**: this is decision-3's economic half; it depends on US1/US2
but its dials belong to TASK-30's calibration worksheet.

**Acceptance Scenarios**:

1. **Given** the cold season, **When** forage regrowth is evaluated, **Then** a
   seasonal multiplier (design target: 0×–0.25× of hot-season regrowth) applies
   deterministically.
2. **Given** den/hunt yields in the cold season, **When** resolved, **Then** the
   seasonal thinning multiplier applies through the existing deterministic yield
   path (emitter-computes; payload carries the outcome).
3. **Given** a solo agent with no stockpile at cold-season start, **When** the
   season runs its course under TASK-30's calibrated dials, **Then** the agent is
   in structural deficit (the worksheet proves it; this spec only requires the
   mechanism to exist).

---

### Edge Cases

- Speed compression (up to 32x/max): temperature is a function of tick, so
  compression changes nothing about determinism; the cognition horizon still
  governs how often minds SEE the temperature (no new mind-loop coupling).
- Worlds created before seasons exist: genesis tuning pin (spec 048) carries the
  dial set — pre-seasons worlds keep binary-era dials and replay byte-identically;
  seasons activate only for worlds whose tuning enables them. No event-log
  format_version bump is required by THIS design (no persisted-name changes, no
  reducer re-derivation from mutable constants — the TASK-75/134 doctrine is the
  reference).
- Mid-season tuning edits (tuning.json promotion path): take effect from the apply
  tick forward via the recorded sim.tuning_applied event, as spec 048 already
  defines; ambient remains pure given (seed, tick, active tuning at that tick).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: A two-season alternating calendar (hot/cold) derived purely from tick
  and a season-length dial (default 10 game-days), with a season-transition event
  emitted at each boundary and narrated by the chronicle.
- **FR-002**: A deterministic ambient temperature function of (seed, tick):
  seasonal baseline + diurnal swing (trough 04:00–05:00, peak 13:00–14:00), no
  persisted state, identical across machines for the same inputs.
- **FR-003**: Warmth decay/recovery proportional to the comfort–ambient gap
  (rate-clamped), replacing the binary warmthLossCold/warmthGainDay constants;
  fire/shelter/sleep act as modifiers; hot-season defaults reproduce today's
  behavior within calibration tolerance.
- **FR-004**: Seasonal scarcity multipliers on forage regrowth and den/hunt yields,
  flowing through the existing deterministic yield paths (emitter-computes).
- **FR-005**: All new dials live in the world-tuning manifest (spec 048): season
  length, baselines, swing amplitude, comfort point, decay slope/clamps, scarcity
  multipliers. Pre-seasons worlds are unaffected (genesis pin compat).
- **FR-006**: Replay determinism test coverage: same log replays byte-identically;
  same seed + same tuning reproduces identical temperature series.
- **FR-007**: Surfacing: current season + a coarse temperature cue are visible to
  players (status/TUI surface — exact placement is the implementing task's spec 047
  design-gate work) and to minds via the existing percept/context assembly (so
  planners can reason "winter is coming").

### Decisions ratified by this design session

- **D1 — Cold snaps and storms are INCIDENTS, not ambient machinery.** A seeded
  cold snap or a storm (dousing outdoor fires — the absorbed TASK-29 question:
  YES, storms douse outdoor fires) is an IncidentScheduleEntry kind in the spec 054
  director-lite vocabulary (TASK-151's catalog is the natural home). The ambient
  function stays smooth and pure; shocks arrive through the incident door as
  recorded events. Rationale: keeps the pure function simple, makes shocks
  authorable per-scenario, and reuses an existing deterministic mechanism.
- **D2 — Night length does NOT vary by season in v1.** Night stays 22:00–06:00;
  the cold season bites through temperature, not darkness. Rationale: varying night
  touches the clock layer and every night-gated mechanic (gru, curfew, musing
  cadence) for flavor the temperature curve already delivers. Revisit only if
  playtests show winter lacks teeth.
- **D3 — Hot season is the compat anchor.** Hot-season defaults are calibrated to
  approximate today's binary behavior, so the change reads as "winter was added",
  not "the world was retuned" — the actual retune is TASK-30's mandate, layered on
  this substrate.
- **D4 — Temperature is world-truth, perception rides existing channels.** No new
  perception events; minds learn temperature through the existing percept/context
  assembly and villagers feel it through needs. (TASK-80's observation channel is
  orthogonal.)

### Key Entities

- **Season**: hot|cold, derived (day ÷ seasonLength) alternating; never persisted
  beyond its transition events.
- **Ambient temperature**: pure function value at a tick; units are abstract
  degrees anchored at comfort = 0 gap.
- **Season dials** (tuning manifest): seasonLengthDays, hotBaseline, coldBaseline,
  diurnalAmplitude, comfortPoint, decaySlope, decayClamp, forageRegrowthColdMult,
  huntYieldColdMult.

## Success Criteria *(mandatory)*

- **SC-001**: This spec is ratified via its PR review (the TASK-28 deliverable) and
  a future implementation task can be scoped from it without a second design
  session.
- **SC-002**: The design satisfies the card's hard invariant on its face: every
  mechanism above is a pure function of (seed, tick, tuning) plus recorded events —
  confirmed by the edge-case analysis (no new persisted state anywhere).
- **SC-003**: TASK-30's calibration worksheet can consume this spec's dial list
  as-is (its food/wood/labor arithmetic plugs into FR-003/FR-004 multipliers).
- **SC-004**: The 30-day proving-run shape crosses 3 season transitions at the
  default season length.

## Assumptions

- Current code anchors verified on the card (2026-07-23 drift audit):
  warmthLossCold=4 at agents.go:463, warmthGainDay=2 at agents.go:465, night
  22:00–06:00 (game-clock), spec 012 fuel mechanics, spec 048 tuning manifest +
  genesis pin. The implementing task re-verifies before building.
- Implementation tiering, phase breakdown, and the spec 047 TUI design-gate work
  belong to the future implementation task, which will extend this spec dir with
  plan.md/tasks.md when it is carded and claimed.
