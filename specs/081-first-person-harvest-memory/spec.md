# Feature Specification: First-person harvest memory (mental-map update at chop/quarry time)

**Feature Branch**: `081-first-person-harvest-memory`

**Created**: 2026-07-26

**Status**: Draft

**Board Task**: TASK-159

**Input**: User description: "Bug: the agent.chopped reducer appends the tile to s.Cleared but removes the tree place-fact from no one's mental map — not the actor's, not on-scene witnesses'; agent.quarried has the same gap. The spec 041 US3 perception sweep then finds each remembered fact absent from ground truth and emits agent.map_corrected plus a salience-5 memory 'The tree at (x,y) had been felled when you looked' — so every chop yields the actor a delayed third-party-voiced discovery of their own act. Verified live on world worldy: all 103 chops self-corrected within ~2-10 ticks; 120 real removals fanned out into 346 loss memories, 75% of all agent.memory_added, flooding the WindowK=10 window and driving false 'the world is barren' journals/conversations. Fix (operator decision 2026-07-26): at chop/quarry time, remove the place fact from the mental map of the actor and of every agent within witnessRadius; the actor's memory of the act is FIRST-PERSON, replacing the later map_corrected discovery. agent.map_corrected keeps its intended narrative: only agents who were elsewhere at act time discover the change on return."

## The problem (observed, world "worldy", 2026-07-26)

Villagers experience their own wood-cutting and quarrying as someone else's
vandalism. Completing a chop marks the tile cleared in world state but leaves
the tree standing in every mental map — including the chopper's own. The
perception sweep then "corrects" each of those maps a few ticks later, minting
a third-party-voiced discovery memory ("The tree at (x,y) had been felled when
you looked.") for the very agent who swung the axe, and for every bystander who
watched them do it. Since completed chops deliberately mint no act memory of
their own, a villager's productive labor is remembered *exclusively* as
unexplained loss.

Measured impact: 120 real removals (103 chops + 17 quarries) fanned out into
346 loss-discovery memories — 75% of every memory formed in the world — and
the villagers' bounded working-memory window carried them into journals and
conversations about the world "becoming barren", which was false.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - My own harvest is mine, in the first person (Priority: P1)

A villager who fells a tree or quarries out an outcrop knows what they did.
The moment the act completes, their mental map no longer shows the resource,
and they carry a first-person memory of the act ("Felled the tree at (12,7)." /
"Quarried the outcrop at (33,31)."). They never later "discover" their own
harvest as a mysterious loss.

**Why this priority**: This is the showstopper's core — the actor
self-discovery accounted for 103 of the 346 loss memories and is the most
narratively corrosive (an agent gaslit by its own labor). Fixing the actor
alone already removes every guaranteed correction.

**Independent Test**: Run a world where one villager chops one tree; inspect
the event log. The villager's memory stream contains exactly one first-person
felling memory and zero "had been felled when you looked" memories for that
tile, at any later tick.

**Acceptance Scenarios**:

1. **Given** a villager whose mental map holds a tree fact at (x,y), **When**
   they complete a chop at (x,y), **Then** the tree fact leaves their mental
   map at that same moment and they gain a first-person memory of felling the
   tree at (x,y).
2. **Given** a villager who completed a chop at (x,y), **When** any later
   perception pass runs with the villager near (x,y), **Then** no map
   correction and no "had been felled when you looked" memory is produced for
   (x,y) for that villager.
3. **Given** a villager whose mental map holds a rock fact at (x,y), **When**
   they complete a quarry at (x,y), **Then** the rock fact leaves their map and
   they gain a first-person memory of quarrying the outcrop at (x,y).

---

### User Story 2 - Watching a neighbor harvest is not a later mystery (Priority: P2)

A villager standing close enough to see a neighbor fell a tree absorbs that
change into their mental map on the spot, silently — no memory is minted for
merely watching, and they never later "discover" the stump as a fresh loss.

**Why this priority**: On-scene witnesses produced the remaining fan-out
(the 243 corrections beyond the 103 self-corrections include many villagers
who were present at the act). Without this, the barren-world signal shrinks
but does not go away.

**Independent Test**: Run a world with two villagers adjacent to the same
tree; one chops it. The bystander's mental map loses the tree fact at the act
tick, their memory stream gains nothing from the event, and no later
correction fires for that tile for them.

**Acceptance Scenarios**:

1. **Given** an awake villager within witness radius of (x,y) whose map holds
   the matching fact, **When** another villager completes a chop or quarry at
   (x,y), **Then** the fact leaves the witness's map at the act tick with no
   memory minted for the witness.
2. **Given** such a witness, **When** any later perception pass runs near
   (x,y), **Then** no map correction fires for (x,y) for that witness.
3. **Given** a villager within witness radius who is asleep at the act tick,
   **When** the chop completes, **Then** their map keeps the fact (they did
   not see it happen) and the existing return-discovery narrative applies to
   them later.

---

### User Story 3 - Genuine return-discovery still works (Priority: P3)

A villager who was elsewhere when a tree fell still gets the intended spec 041
narrative: returning within sight of the remembered tree, they discover it
gone and carry the situated discovery memory ("The tree at (x,y) had been
felled when you looked.").

**Why this priority**: Regression guard — the correction machinery is the
feature working as designed for absent agents; the fix must narrow it, not
break it.

**Independent Test**: Villager A witnesses a tree, walks out of witness
radius; villager B chops it; A returns. A's log shows exactly one map
correction for the tile with the discovery memory.

**Acceptance Scenarios**:

1. **Given** a villager whose map holds a fresh tree fact and who is outside
   witness radius of (x,y) at the act tick, **When** they later come within
   witness radius of (x,y), **Then** a map correction fires for (x,y) with the
   situated discovery memory, exactly as today.
2. **Given** the same setup, **When** the correction fires, **Then** the fact
   leaves their map and no repeat correction ever fires for that tile.

---

### Edge Cases

- **Actor's map never held the fact** (perception cadence gap: the villager
  reached and worked the tile between sweep passes): the map removal is a
  no-op; the first-person act memory is still minted.
- **Fact known by hearsay**: a witness's fact for the tile carries `told`
  provenance (a neighbor's directions) rather than `witnessed`. Removal at act
  time matches on place and kind, regardless of provenance — watching the tree
  fall overrides whoever told you about it.
- **Same-tick sweep**: a witness's staggered perception pass lands on the act
  tick. No correction may fire at the act tick or later for an on-scene
  party's removed fact — the removal and the sweep must not race.
- **Witness's current intent targets the removed tile** (e.g. walking over to
  chop that same tree): today the map correction doubles as the planner's
  re-arm (absorb) trigger for a matching intent target. Silent removal MUST
  preserve that re-arm behavior for witnesses whose current intent targets the
  removed coordinates; otherwise the witness walks to a stump and fails
  opaquely.
- **Multiple witnesses**: every awake villager within witness radius has their
  matching fact removed independently; villagers holding no fact for the tile
  are unaffected.
- **Dead villagers within radius**: no map updates for the dead (perception
  parity — the sweep only serves awake living villagers).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Completing a chop MUST remove the tree place-fact at the act
  coordinates from the acting villager's mental map, effective at the act
  event itself (not a later pass).
- **FR-002**: Completing a quarry MUST remove the rock place-fact at the act
  coordinates from the acting villager's mental map, effective at the act
  event itself.
- **FR-003**: The acting villager MUST receive a situated first-person act
  memory for a completed chop ("Felled the tree at (x,y).") and quarry
  ("Quarried the outcrop at (x,y).") — the existing first-person act-memory
  shape (the hunt precedent), in the low, non-generation-interrupting salience
  band, minted through the standard memory-accretion event (never appended
  directly to state).
- **FR-004**: Every awake living villager within witness radius of the act
  coordinates at the act tick MUST have the matching place-fact (same place
  and kind, any provenance) removed from their mental map at the act event,
  silently — no memory is minted for witnessing.
- **FR-005**: Villagers outside witness radius at the act tick, asleep, or
  dead MUST be unaffected at act time: their remembered fact stays, and the
  existing return-discovery correction (spec 041 US3) applies unchanged when
  they next perceive the place.
- **FR-006**: No map correction and no discovery memory may ever be produced
  for a villager for a fact their own act or on-scene witnessing already
  removed — including a perception pass landing on the act tick itself.
- **FR-007**: For a witness whose current intent targets the removed
  coordinates (or resolves to them), the act-time removal MUST trigger the
  same planner re-arm the map correction triggers today; a removal elsewhere
  in the map stays quiet, exactly as today.
- **FR-008**: All new state mutations (map-fact removals) MUST live in the
  event-application path as pure functions of the event and prior state, so a
  fresh replay of an event log reproduces identical state — including every
  mental map — under the same code version.
- **FR-009**: Memories MUST continue to accrete only via the standard
  memory-accretion event; the act memory rides the same event batch as the
  act, as a companion event.
- **FR-010**: Foraging, pile draining, and structure removal are explicitly
  unchanged by this feature: forage spots regrow and never correct; pile and
  structure lifecycles keep their current perception behavior.

### Key Entities

- **Place fact**: one remembered resource/landmark in a villager's mental map
  — kind, coordinates, last-seen tick, provenance (`witnessed`/`told`/…).
  This feature adds a new way facts leave a map (act-time removal) alongside
  the existing correction path.
- **Mental map**: the per-villager collection of place facts; grown by the
  perception sweep and hearsay, shrunk today only by corrections, and after
  this feature also by acts the villager performed or watched.
- **Act memory**: the situated first-person memory minted for the actor at
  completion of a chop/quarry — new for these two acts, following the existing
  hunt-memory shape.
- **Map correction**: the existing discovery event + memory pair for a
  remembered place found gone; after this feature it fires only for villagers
  who were genuinely absent (or asleep) at removal time.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: On a fresh world run of comparable length to the observed one,
  zero map corrections name a tile the same villager chopped or quarried
  (baseline: 103 of 103 chops self-corrected).
- **SC-002**: Zero map corrections name a tile the correcting villager was
  within witness radius of (awake) at the moment it was cleared.
- **SC-003**: Every map correction in a fresh run's event log corresponds to a
  villager who was absent or asleep at removal time and later returned — 100%
  genuine return-discoveries (baseline: loss-discovery memories were 75% of
  all memories formed).
- **SC-004**: Each completed chop/quarry yields exactly one first-person act
  memory for its actor and zero memories for on-scene witnesses.
- **SC-005**: Replaying a fresh run's event log from genesis reproduces the
  run's final state exactly, mental maps included.
- **SC-006**: On a spot-check of journals and conversations from a fresh
  multi-day run with normal harvesting, villagers no longer describe
  unexplained resource disappearance or a world going barren absent actual
  scarcity.

## Assumptions

- **Operator decision (2026-07-26), recorded**: the actor's memory of the act
  is first-person. What is stored as memory (whether every completed
  chop/quarry stays memory-worthy, salience tuning, witness memories) may be
  re-evaluated later — out of scope here. This deliberately supersedes the
  earlier "completed chops mint no memory" spam-avoidance posture for these
  two acts; the hunt act-memory precedent shows the shape and band.
- Witness radius for "watched it happen" is the same witness radius the
  perception sweep uses — one perceptual reality, no second constant.
- Cross-version replay (event logs recorded before this change replayed under
  code after it) follows the project's existing posture on reducer semantic
  evolution; this spec requires determinism under a single code version only.
- The exact first-person phrasings above are illustrative of voice and
  content (act, kind, coordinates); final strings may be tuned during
  implementation without a spec change.
- Miracle terrain removal (operator-initiated clearing) is not an agent act
  and keeps today's behavior: nearby villagers discover it via the correction
  path — mysterious loss is the correct narrative for divine intervention.
