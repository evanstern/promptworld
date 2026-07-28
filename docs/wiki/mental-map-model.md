---
name: mental-map-model
description: Child of [[mental-maps]] — the MentalMap data model itself: the Explored bitmap and PlaceFact/PeerSighting types, read-time freshness horizons, the knowledge predicates that test them, derived (eventless) bookkeeping on movement, replay determinism, and the genesis/v3→v4-migration knowledge grants. Load for the type's field-level shape or its correctness guarantees.
kind: component
sources:
  - internal/sim/mentalmap.go
  - internal/sim/state.go
  - internal/sim/miracles.go
  - internal/sim/migrate.go
verified_against: 63390f122bdf4e1b7abf518a8be83de725f06230
---

# Mental map data model

Child of [[mental-maps]]: `MentalMap`'s own shape and invariants — the
`Explored` bitmap, the `Facts`/`Peers` slices, their freshness rule, the
knowledge predicates that read them, how they mutate without events on
movement, why replay agrees byte-for-byte, and how a fresh or migrated world
seeds them. [[mental-map-perception]] and [[mental-map-propagation]] cover
how the model is grown, read for resolution, and told to others.

**The type** (`internal/sim/mentalmap.go`): `Agent.Map *MentalMap` (`omitempty`,
the Journal/Hail pointer precedent — a never-mapped agent, i.e. every pre-041
snapshot, round-trips byte-identically). `MentalMap{Explored, Facts, Peers}`:

- `Explored` is a base64-encoded W×H row-major bitset (LSB-first per byte) —
  terrain shape known, monotone (bits only ever OR in, `MarkExplored`), never
  un-set. `ExploredAt`/`exploredBytes` decode lazily and grow (never shrink) to
  cover the current map size; a corrupt encoding decodes as all-unexplored
  rather than erroring, keeping the reducer total.
- `Facts []PlaceFact{Kind, X, Y, Seen, Provenance, Source?, Detail?}` are known
  dynamic entities, kept sorted by `(Kind, X, Y)` at every mutation
  (`factLess`/`sortFacts`/binary-search `upsertFact`/`removeFact`) so canonical
  JSON bytes never depend on discovery order — at most one fact per
  `(Kind, X, Y)`. `Kind` is a closed vocabulary: the structure kinds
  (`fire`/`shelter`/`oven`/`chest`/`wall_plank`/`wall_stone`/`path`, plus —
  since spec 044 US4 — `grave`), the
  perception-gated resource kinds (`tree`/`forage`/`rock`/`water_edge`/`den`/
  `pile`), and — since spec 084 — `designation`: the guardian's announced
  plan mark ([[guardian-designations]]), granted reducer-side by the
  `designation.placed` arm to every living villager at the anchor tile
  (provenance `revealed`); deliberately NOT in `placeFactKinds`
  (send_vision reveals real world places only). Spec 068's marsh/sand ground covers ([[worldmap-generation]])
  deliberately do NOT join this vocabulary — they carry no resource
  affordance, so `perceptionEvents` has no fact kind to record for them; a
  villager's map has nothing to say about marsh or sand beyond what it can
  already see on the drawn map, the same as plain grass. `Provenance` reuses the Belief vocabulary — `witnessed`/`told`, plus
  `ProvenanceRevealed` ("revealed") for a divine grant. `Source` is the
  teller's agent index, meaningful only under `told` provenance. `Detail` is a
  kind-specific scalar baked at emission and never re-derived — a fire's
  `FuelUntil` as last seen; every other kind 0.
- `Peers []PeerSighting{Agent, X, Y, Seen}` are last-seen positions of other
  villagers, sorted by agent index (`peerSightingOf`/`sightPeer`) —
  `talk_to`/`seek` resolve against this, never live coordinates.

**Freshness horizons** (research D6): a fact is fresh iff `now − Seen <
factHorizon(kind)`, evaluated at READ time only — time never mutates a
fact, so snapshots stay churn-free. Volatile kinds (`fire`, `pile`) get
`factHorizonVolatileTicks` (12 game-hours); `designation` gets
`factHorizonDesignationTicks` (7 game-days — the max directive TTL, so the
announcement outlives any directive bound to it, spec 084); every other
kind gets `factHorizonDurableTicks` (4 game-days). A stale fact stays stored (invisible
to resolvers/prompt) until something removes it: a perception correction, the
agent's death, or — since spec 081 — a chop/quarry the agent performed or
watched in radius (`removeHarvestedFact`, [[sim-state-reducer]], calling the
same `removeFact` primitive from the `agent.chopped`/`agent.quarried` arms).
Staleness is never itself a removal. `PlaceFact.Fresh(now)` exports the same
test for [[agent-mind]]'s prompt renderer, one freshness rule shared by every
reader.

**Knowledge predicates** (`mentalmap.go`, research D3): `knownFreshFact`/
`knownFactAt` test a fresh fact at one tile (a nil map — a dead migrated
native, or a bare test agent — uniformly means "knows nothing");
`knowsAnyFresh(a, kind, now)` is the knowledge-emptiness test behind every
"you know of no `<kind>`" rejection, checked BEFORE reachability so the two
failure classes stay distinct (knowing none is epistemic; knowing some but
reaching none is the pre-existing "no `<kind>` reachable" phrasing);
`knowsLitFire`/`warmKnownPredicate` gate the cook/warmth rungs on remembered
`Detail` (a fire's `FuelUntil` as last seen, still ahead of `now` — the agent
predicts burnout from its own knowledge, never a live read).

**Derived bookkeeping** (research D2): position-changing reducer arms silently
grow the mover's explored bitmap and update peer sightings — no event, a pure
function of (state, event): `agent.moved`, `agent.woke`, and a `villager`-class
`metatron.entity_moved` (a miracle-teleported villager is knowledge-transparent,
not a blind teleport) all call `markExplored`/`notePresence`
(`internal/sim/state.go`, `internal/sim/miracles.go`). `notePresence` records a
sighting between the arriving agent and every living, AWAKE agent within
`witnessRadius` — mutual, since villagers cross each other's sight constantly
and event-carrying every sighting would flood the log the way per-step
explored events would.

**Genesis and migration** (research D7): `NewState` grants each agent explored
surroundings at `witnessRadius` around its landing tile and zero facts
(cold-start worlds have no structures yet); a second pass seeds mutual peer
sightings for villagers who spawn within sight of each other — nothing else.
The v3→v4 migration transform (`internal/sim/migrate.go`'s `TransformV3State`,
[[world-migration]]) grants the same knowledge to an UPGRADED world: each
living agent gets explored terrain around its position plus witnessed facts
for every current structure and ground pile (natives, not strangers — a
migrated villager already lives in its village, so it is never handed a blank
map that would force it to re-discover a home it has always known); a DEAD
agent gets an empty but non-nil map (genesis now seeds maps for everyone, and
a replica/recovery unmarshal MERGES a snapshot over a genesis state, so a
map-absent agent would silently resurrect the genesis map there while a
from-genesis replay produces the transform's own value — an explicit empty
map is what makes the two paths agree byte-for-byte). This bumped the save
format to **v4** ([[world-save-directory]]).

**Replay determinism**: every mutation is either a pure derived function of
(state, event) — `markExplored`/`notePresence` — or a recorded event whose
payload is fully baked at emission (never re-derived at Apply time), so live
and replay agree byte-for-byte; `TestDeterminismSameSeedSameTimeline`
additionally diffs each agent's canonical map bytes across two same-seed
runs. `metatron.time_snapped`'s `rebaseTicks` (`internal/sim/miracles.go`,
[[guardian-miracles]]) classifies `PlaceFact.Seen`/`PeerSighting.Seen` as
SHIFT (the freshness anchor, so a time snap cannot instantly stale every
villager's knowledge) and `PlaceFact.Detail` as KEEP (a remembered value,
never rewritten — the perception sweep simply re-witnesses the shifted
reality on the next look).

## Connections

Parent [[mental-maps]] summarizes this note and links every sibling child;
[[mental-map-perception]] is the primary reader of this type (freshness,
predicates); [[mental-map-propagation]] is what upserts told/revealed facts
into it; [[sim-state-reducer]] carries `Agent.Map` and the reducer arms that
call the derived bookkeeping here; [[guardian-miracles]] shares the
`rebaseTicks` SHIFT/KEEP taxonomy; [[world-migration]] carries the v3→v4
knowledge grant; [[world-save-directory]] is the format-version gate this
bumped to v4; [[worldmap-generation]] is why marsh/sand never join the
`Kind` vocabulary.
