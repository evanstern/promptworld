# Feature Specification: Perception of absence — grounded arrival observations make beliefs falsifiable

**Feature Branch**: `task-80-perception-of-absence`

**Created**: 2026-07-29

**Status**: Draft

**Input**: TASK-80 — architectural root of the Thornspire finding (2026-07-23,
world-01): the sim only tells agents what happened, never what is NOT there;
confabulated beliefs are unfalsifiable by construction. Depends on TASK-79's
reinforcement seam (Done). Pairs deliberately with TASK-81 (canonization) so the
god can make a myth true before reality debunks it.

## Design decisions (the card's spec questions, resolved)

- **D1 Trigger — intent-completing arrivals, not every step.** The grounded
  observation emits when an agent ARRIVES at the destination its intent chose
  (walk-goal completion), not on every wander step tile. Rationale: the card's
  flood worry is real (every step is an arrival); the falsifiability need is
  about places agents deliberately went. Reason interpretation (did the intent's
  reason name a phenomenon?) stays OUT of the reducer — the executor emits
  uniformly; the MIND weighs relevance (D3). No LLM in the emission path.
- **D2 Event shape — exhaustive-within-radius.** `agent.place_observed` carries
  the agent, position, and the complete set of feature/entity kinds actually
  present within the existing placeScanRadius (the describePlace substrate).
  Absence is IMPLIED by exhaustiveness: anything not listed was not there. No
  "absence_of: X" field — the reducer cannot know what an agent expected;
  expectation lives mind-side.
- **D3 Belief interaction — mind-side, through the TASK-79 seam.** On processing
  a place_observed memory, beliefs referencing that location get: confirmation
  boost when the believed feature appears in the observation; DISCONFIRMATION
  decay when the location was observed and the feature is absent — faster than
  ambient silence decay but bounded so myths die slowly (dials, not cliffs);
  silence (never visiting) keeps today's decay. Observation-vs-belief matching
  is the mind's judgment (it may use LLM there; never in the sim).
- **D4 Salience/flooding — low base salience + dedup window.** Observation
  memories enter at low salience; repeat observations of an unchanged place
  within a dedup window collapse (no working-window flooding). Disconfirming
  observations get a mind-side salience bump through the D3 pathway (surprise is
  memorable), not an executor-side one.
- **D5 Determinism.** Emission is executor-side, a pure function of world state
  at the arrival tick; payload carries the full observation (emitter-computes,
  spec 092/094 doctrine — additive event type, no format-version bump).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Reality gets a voice at Thornspire (Priority: P1)

As a villager in the game who walked to the forest's edge because I believed
"the tendrils and stones of Thornspire" were there, I want my arrival to record
what actually IS there, so my belief can lose confidence from my own experience
instead of living forever on social reinforcement.

**Acceptance Scenarios**:

1. **Given** an agent arriving at an intent destination, **When** the walk
   completes, **Then** `agent.place_observed` is emitted with the exhaustive
   within-radius feature set, deterministically and replay-safe.
2. **Given** a belief naming a feature at/near that location, **When** the
   observation lacks the feature, **Then** belief confidence decays through the
   TASK-79 seam, faster than silence but slower than a cliff (dials); **When**
   the observation contains it, confidence is reinforced.
3. **Given** repeated arrivals at an unchanged place inside the dedup window,
   **Then** no additional observation memory accumulates in the working window.

---

### User Story 2 - Legibility: the decision trail shows the observation (Priority: P2)

As a player using the decision-trace view, I want grounded observations visible
(chronicle digest entry + decision trail), so "went to Thornspire, found
nothing" is readable evidence of why the myth faded.

**Acceptance Scenarios**:

1. **Given** the new event type, **Then** the TUI digest grammar renders it and
   event-types.md documents it (TestCatalogSweep green); belief-confidence
   movement it caused is visible where TASK-79's reinforcement already surfaces.

---

### US3 - The soak proves no flooding (Priority: P2)

**Acceptance Scenarios**:

1. **Given** a seeded MEASURE world (never the playtest) run for at least one
   game-day at 8x, **Then** working-memory windows show observation memories
   bounded per D4 (evidence recorded under docs/design/evidence/task-80/), and
   ordinary survival behavior is unchanged.

## Requirements *(mandatory)*

- **FR-001**: `agent.place_observed` per D1/D2/D5 — executor-emitted on
  intent-completing arrivals, exhaustive within placeScanRadius, deterministic,
  additive (no format-version bump).
- **FR-002**: Situated observation memory with low base salience + dedup window
  (D4); provenance "observed" (first-person, strongest class per TASK-79's
  hygiene).
- **FR-003**: Mind-side belief reconciliation through the TASK-79 seam (D3):
  confirmation boost, disconfirmation decay (faster than silence, dial-bounded),
  silence unchanged; matching logic lives in internal/mind only.
- **FR-004**: Dials in the tuning manifest (spec 048): dedup window, base
  salience, disconfirmation multiplier, confirmation boost.
- **FR-005**: Chronicle digest entry + event-types.md; TestCatalogSweep green.
- **FR-006**: Tests: emission determinism + replay byte-identity for existing
  logs; dedup behavior; belief movement paths (confirm/disconfirm/silence);
  soak evidence per US3.

## Success Criteria *(mandatory)*

- **SC-001**: On a seeded world with an implanted false place-belief, visits
  drive confidence measurably down through recorded events alone; the belief
  survives multiple visits (myths die slowly) but trends to extinction.
- **SC-002**: Existing replay fixtures byte-identical; soak shows bounded
  observation-memory counts.
- **SC-003**: TASK-81's canonization can flip the same scenario: after a feature
  is made real, the next visit CONFIRMS (verified in 81's own AC; this spec only
  requires the confirm path to exist).

## Assumptions

- The describePlace/placeScanRadius machinery (executor-social-perception note)
  is the substrate; TASK-79's reinforcement seam exposes confirm/decay hooks.
- Tier: **Opus** — cross-package (sim executor/reducer + mind belief semantics),
  doctrine-adjacent epistemics. Escalation N/A (already senior tier).
