# Feature Specification: Per-Agent Mental Maps

**Feature Branch**: `task-96-agent-mental-maps`

**Created**: 2026-07-24

**Status**: Draft

**Input**: User description: "Per-agent mental maps: knowledge-gated spatial memory replaces omniscient nearest-target resolution (TASK-96). Each villager agent holds a private mental map of the 64x64 grid with an explicit unknown state; verb target resolution consults the agent's spatial knowledge instead of the current omniscient whole-map BFS; the planner prompt renders only agent-known structures/places; unknown space becomes deliberately explorable; remembered content can go stale and is corrected on arrival; spatial knowledge can spread socially through talk. Representation must have a documented extension path to future layered 3D grids. Replay determinism is a hard constraint. Grounding corpus: research/Agent-Mental-Maps/ vault branch (commit c70c53f)."

## Clarifications

### Session 2026-07-24

- Q: What should an agent's spatial knowledge gate? → A: Target choice only — terrain is
  common knowledge (villagers are natives); pathfinding runs on ground truth; knowledge
  gates what to walk toward (structures, fires, resource sites, agents).
- Q: Does the deterministic survival reflex use the same private map as cognition? → A:
  Full parity — reflex resolves targets against the same mental map; world viability is
  protected by seeded home knowledge and the search behavior, never by omniscient fallback.
- Q: How does spatial knowledge spread when villagers talk? → A: Automatically on every
  completed talk encounter — a small bounded set of relevant place-facts transfers with
  told-by provenance and lower initial confidence; no LLM tool call required.
- Q: How does the LLM planner consume the mental map? → A: Rendered prompt text only (the
  known-places section replaces the global structure list); read-only map-query tools are
  deferred to a future task.
- Q: Decay rates for remembered dynamic facts? → A: Adopt the research-aligned default —
  confidence decays toward unknown (never toward false) on the existing belief half-life
  family, with volatile kinds (lit fires) decaying faster than durable ones (buildings);
  exact constants are set during planning/tuning.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Agents act only on places they know (Priority: P1)

A villager who wants to warm up, cook, refuel a fire, or visit a structure can only target
one they have personally seen, witnessed being built, or been told about. If they know of no
such place, the action fails honestly — "you don't know of any lit fire" — instead of the
simulation silently teleport-targeting the nearest one anywhere on the map. The story a
watcher reads becomes epistemically honest: villagers go to places *they* know, not places
the world knows.

**Why this priority**: This is the core inversion — replacing omniscient nearest-target
resolution with knowledge-gated resolution. Every other story builds on the private map this
one introduces. Without it, nothing else in the feature is observable.

**Independent Test**: Seed a world where a fire exists far outside a villager's known area.
Ask the villager (via its normal cognition) to warm up. Verify the intent fails with an
"unknown to you" outcome rather than resolving to the distant fire. Then let the villager
witness the fire and verify the same action now resolves to it.

**Acceptance Scenarios**:

1. **Given** a villager who has never seen or heard of any fire, **When** their turn resolves
   a warmth/cooking verb, **Then** the action is rejected with a distinct "no known target"
   outcome (not "none exists"), and the rejection is visible to the villager's next planning
   turn.
2. **Given** a villager who witnessed a fire being built at (x,y), **When** they later choose
   a warmth verb from elsewhere, **Then** the resolved target is that fire (or the nearest
   *known* fire if several are known), never an unknown-but-closer one.
3. **Given** two fires — one known and far, one unknown and near, **When** the villager seeks
   warmth, **Then** they travel to the known far one.

---

### User Story 2 - The map fills in through perception, and the prompt tells only what the agent knows (Priority: P2)

As a villager moves through the world, what they perceive around them (structures, terrain
features, resource sites) becomes part of their private map. Their planning prompt describes
their surroundings and their *known* places — every known structure, with no arbitrary cap —
and never mentions places they have not learned about. Two villagers standing in different
histories see different worlds.

**Why this priority**: This is the write path and the presentation layer. US1 gates actions
on knowledge; US2 makes knowledge accrue and makes the difference visible in each agent's
mind. It also retires the current defect where only the first six structures (in creation
order) are ever shown.

**Independent Test**: Walk one villager past a cluster of structures while a second villager
stays away. Compare the two planner prompts: the traveler lists the cluster, the homebody
does not. Build a seventh structure in the traveler's presence and verify it appears in
their prompt.

**Acceptance Scenarios**:

1. **Given** a villager whose route passes within perception range of a structure, **When**
   their next planning turn renders, **Then** that structure appears among their known places
   with its location.
2. **Given** a village with more than six structures all known to a villager, **When** their
   prompt renders, **Then** all known structures are represented (no first-six truncation).
3. **Given** a structure the villager has never perceived or been told about, **When** their
   prompt renders, **Then** that structure does not appear in any form.

---

### User Story 3 - Remembered places can be wrong, and reality corrects them (Priority: P3)

A villager's map records what they last saw, not what is currently true. A fire remembered
as lit may have burned out; a structure remembered as standing may have been demolished.
When the villager arrives where they remember something, what is actually there overwrites
the memory — and the surprise is visible in their story ("the fire was cold when she
arrived").

**Why this priority**: Staleness is what makes private maps *matter* narratively and
mechanically. Without correction, wrong knowledge is permanent; with it, agents live the
believe-act-discover loop. Depends on US1/US2 existing.

**Independent Test**: Let a villager learn a fire's location, move them away, let the fire
burn out, then have them seek warmth. Verify they travel to the remembered fire, discover it
cold, their map updates, and their next action re-plans from corrected knowledge.

**Acceptance Scenarios**:

1. **Given** a villager who knows a fire at (x,y) and that fire has since burned out,
   **When** they arrive seeking it, **Then** their map no longer records a lit fire there,
   and the arrival discovery is observable in their memory/event trail.
2. **Given** a villager arriving anywhere, **When** their perception covers tiles whose
   remembered contents differ from reality, **Then** the map is corrected to reality for
   those tiles (both additions and removals).

---

### User Story 4 - Villagers can search the unknown deliberately (Priority: P4)

A villager who needs something they know no location for can *search for it*: move toward
the nearest boundary of their unexplored space and look around, repeatedly, until they find
a target or give up. Search is directed at the unknown, not a random shuffle near their
feet.

**Why this priority**: Without a search behavior, knowledge-gating (US1) risks deadlock —
an agent that knows nothing can do nothing. Exploration turns "I don't know of any X" into a
plan instead of a dead end. Depends on US1's unknown state existing.

**Independent Test**: Place a villager with a mostly-unknown map and a need whose target
exists only in unexplored territory. Verify the villager's search behavior reaches unknown
space, expands their map, eventually perceives the target, and then acts on it normally.

**Acceptance Scenarios**:

1. **Given** a villager with no known forage sites and unexplored map regions, **When** they
   attempt to forage and choose to search, **Then** their movement is directed toward
   unexplored space (not a random walk within already-known tiles).
2. **Given** a search in progress, **When** the sought target enters perception, **Then**
   the search concludes and the villager proceeds against the now-known target.
3. **Given** a villager whose map is fully explored and who still knows no matching target,
   **When** they attempt the verb, **Then** they receive the honest "none known" outcome
   (searching is not offered as a false hope).

---

### User Story 5 - Spatial knowledge spreads through talk (Priority: P5)

When villagers talk, useful place-knowledge can pass between them: "there's a fire by the
east woods." Knowledge learned secondhand is distinguishable from firsthand witness — it can
be wrong, it can be stale, and the receiving villager's story reflects who told them.

**Why this priority**: Social spread is what makes private maps a *social* system — scouts,
directions, and spatial rumors become possible. It is last because it layers on everything
above and touches the existing conversation machinery.

**Independent Test**: Villager A knows a fire location; villager B does not. Let them talk.
Verify B can subsequently target that fire, that B's knowledge is marked as told (not
witnessed), and that B's prompt/story reflect learning it from A.

**Acceptance Scenarios**:

1. **Given** villager A knows a place villager B does not, **When** they complete a talk
   encounter, **Then** B may act on that place afterward, and B's record of it carries
   told-by-A provenance.
2. **Given** knowledge passed by talk that was already stale when shared, **When** B acts on
   it and arrives, **Then** US3's correction applies and B's map records reality.

---

### Edge Cases

- **Cold start**: villagers at world creation must begin with seeded knowledge of their home
  area (spawn surroundings and founding structures), or every agent starves before learning
  anything. Newly created worlds must remain viable.
- **Survival reflex**: the deterministic reflex layer that keeps bodies alive between LLM
  thoughts must consult the same private map — an agent must not "reflex-know" what it
  cognitively doesn't — while world viability (edge case above plus search behavior) keeps
  ignorance from being lethal by construction.
- **Demolition/decay while away**: structures removed while no one watches must not linger
  forever in every absent villager's map; correction happens on next perception (US3), and
  confidence in old sightings decays with time.
- **Target agent moves**: talking to another villager can no longer assume their live global
  position; resolution uses last-known/perceived position, and the existing landing guard
  (target present within radius at arrival) already handles misses.
- **Plan steps queued on stale knowledge**: a multi-step plan whose later step names a
  place that stops being known (or corrected away) must fail that step honestly, not
  silently re-resolve omnisciently.
- **Dead villagers**: a dead agent's map is retired with it; no posthumous knowledge
  channels.
- **Metatron (the god agent)**: divine actions and visions are exempt from knowledge gating
  (omniscience is in character) and remain a deliberate channel for *granting* villagers
  place-knowledge.
- **Replay**: every map mutation must derive deterministically from simulation state and
  recorded events; replaying a world file reproduces every agent's map bit-identically.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Each villager agent MUST hold a private spatial knowledge state ("mental map")
  over the world grid in which every tile/fact is in exactly one of: unknown, known-current
  (perceived now), or known-remembered (perceived or learned earlier, possibly stale).
- **FR-002**: The mental map MUST record, for each known place-fact, at minimum: what, where,
  when last perceived/learned, and how it was learned (witnessed firsthand vs told
  secondhand vs divinely revealed), consistent with the existing witnessed/told provenance
  vocabulary.
- **FR-003**: Target resolution for world-acting verbs MUST consult only the acting agent's
  mental map for knowledge-gated target kinds; resolution MUST NOT scan ground-truth world
  state for targets the agent does not know.
- **FR-004**: When an agent knows no matching target, the verb MUST fail with an outcome
  distinct from "no such thing exists/reachable" — the agent (and its next planning turn)
  learns "you know of none", phrased so the difference is visible in the decision trail.
- **FR-005**: Static terrain (ground, water, rock, paths as terrain) is treated as known to
  all villagers by default; dynamic content (structures, fires, resource sites, piles,
  agents) is knowledge-gated. Movement/pathfinding over terrain remains ground-truth;
  *choosing where to go* is what knowledge gates.
- **FR-006**: Perception MUST write to the map: whenever an agent's perception covers tiles
  (arrival, passing through, witnessing events), the map records what is actually present —
  including recording *absence* where remembered content is gone (correction).
- **FR-007**: Villagers MUST begin (at world creation and at agent creation) with seeded
  knowledge of their home area — spawn surroundings and founding structures — sufficient for
  world viability.
- **FR-008**: The planner prompt MUST render the agent's known places instead of the global
  structure list: all known structures/places relevant to the agent (no fixed first-N
  truncation), each with location, and nothing the agent does not know.
- **FR-009**: An agent MUST be able to search deliberately: a directed exploration behavior
  that moves toward the nearest reachable boundary of its unexplored space, available when a
  verb fails for lack of a known target; when the map has no unexplored reachable space the
  behavior reports exhaustion honestly.
- **FR-010**: The existing undirected wander remains available; search (FR-009) is a
  distinct, unknown-directed behavior.
- **FR-011**: Every completed talk encounter MUST automatically transfer a bounded amount
  of relevant place-knowledge between participants (no LLM tool call required); transferred
  knowledge carries told-by provenance and lower initial confidence than witnessed
  knowledge, and the sharing is observable in both agents' records.
- **FR-012**: Confidence in remembered dynamic facts MUST decay with game time toward
  unknown (not toward false), on a scale consistent with the existing belief half-life
  machinery; decayed-out facts stop gating targets as known.
- **FR-013**: The survival reflex policy MUST resolve targets against the same mental map as
  cognition, so reflex and mind share one epistemic state.
- **FR-014**: The god agent's (Metatron's) capabilities are exempt from knowledge gating,
  and divine communication MUST be able to grant a villager place-knowledge (provenance:
  revealed).
- **FR-015**: All mental-map state MUST be simulation-owned, serialized with world saves,
  and reconstructed bit-identically on replay; no map mutation may originate outside the
  deterministic simulation step.
- **FR-016**: The chosen representation MUST have a documented extension path to layered 3D
  grids (multiple stacked levels joined at portals), covering: per-level knowledge, portal
  knowledge, and how unknown-space search generalizes. Documentation, not implementation.
- **FR-017**: The chronicle/story surfaces MUST be able to distinguish knowledge events —
  discovering, correcting (arrived and it was gone), being told, and revealing — so the
  epistemic life of the village is legible to watchers.

### Key Entities

- **Mental Map**: one per living villager; private spatial knowledge over the world grid;
  explicit unknown state; serialized with the agent.
- **Place-Fact**: a single known thing-at-place: kind, location, last-perceived/learned game
  time, provenance (witnessed / told-by / revealed), confidence.
- **Unknown Region**: the complement of explored space; the target of search; shrinks under
  perception.
- **Knowledge Event**: a discovery, correction, telling, or revelation — the observable
  trace of a map changing, feeding memory/chronicle.
- **Seed Knowledge**: the home-area knowledge granted at world/agent creation for viability.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In a multi-game-day soak run, zero instances of an agent resolving a
  knowledge-gated target it had never witnessed, been told about, or been granted — auditable
  from the event/decision trail.
- **SC-002**: Two villagers with different movement histories produce different known-place
  prompt sections, and a villager's prompt lists 100% of their known structures (a 7th+
  known structure always appears; the first-six cap is gone).
- **SC-003**: A replayed world file reproduces every agent's mental map bit-identically at
  every checkpoint of a soak run.
- **SC-004**: An agent whose map lacks a needed target and who searches reaches previously
  unknown space and grows its explored coverage in a bounded number of game hours;
  observed searches terminate (find, exhaust, or abandon) — no infinite search loops in the
  soak run.
- **SC-005**: When a remembered structure is removed and the agent later perceives its
  location, 100% of such cases correct the map, and a knowledge event records the
  discovery.
- **SC-006**: After a talk transfer, the receiving agent can successfully act on the shared
  place, and its record shows told-by provenance — demonstrated at least once in a natural
  (unscripted) soak run.
- **SC-007**: Newly created worlds remain viable: villagers survive the first game days at a
  rate indistinguishable from pre-feature worlds (no starvation-by-ignorance regression).
- **SC-008**: The 3D extension path document exists and covers per-level knowledge, portals,
  and search generalization.

## Assumptions

The five formerly-open design questions (the research MOC's open questions) were decided in
the 2026-07-24 clarification session — see Clarifications above; they are no longer
assumptions. What remains below are the working defaults this spec rests on.

- **Perception envelope**: perception reuses the existing witness radius (Manhattan 8) as
  the default "you can see it" range; no line-of-sight computation in this feature.
- **talk_to gating**: targeting another villager uses last-known position (perceived or
  told); the existing arrival guard already handles "they moved".
- **3D posture**: future verticality is layered grids joined at portals (stairs-like), per
  the research corpus; this feature only documents the extension path (FR-016).
- **Scale envelope**: 64x64 grid, single layer, 8 villagers; per-agent map memory is
  negligible at this scale (research: ~4 KiB per byte-layer per agent), so representation
  choice optimizes semantics and 3D extensibility, not size.
- **Related work**: TASK-80 (arrival observations) shares the perception write path; TASK-79
  (belief reinforcement) receives correction signals; TASK-76 (entity-lookup seam) is the
  world-side query seam; TASK-95's failure-semantics distinction ("none known" vs "none
  exists") is subsumed by FR-004 for knowledge-gated verbs. Sequencing is a plan concern.
