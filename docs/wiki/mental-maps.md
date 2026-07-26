---
name: mental-maps
description: Per-agent private spatial knowledge (spec 041) — an explored-terrain bitmap plus known place-facts with provenance and freshness horizons, gating target resolution (nearestKnown, search/frontier, last-known-sighting talk_to), rendered into the prompt's known-places section, grown by a perception sweep and corrected when facts go stale, and carried across a v3→v4 migration; spec 044 adds the grave kind, placed by the agent.died reducer arm
kind: component
sources:
  - internal/sim/mentalmap.go
  - internal/sim/executor.go
  - internal/sim/policy.go
  - internal/sim/path.go
  - internal/sim/state.go
  - internal/sim/social.go
  - internal/sim/miracles.go
  - internal/sim/migrate.go
  - internal/mind/prompt.go
  - internal/mind/context.go
  - internal/mind/mind.go
  - internal/guardian/turn.go
  - internal/guardian/toolcalls.go
  - internal/tool/registry.go
verified_against: d304e8adb64fdf40e24bfeca3ca3420e8a840a35
---

# Mental maps

Spec 041 retires the villagers' omniscience: before this feature, [[reflex-policy]]'s
resolvers and [[agent-mind]]'s prompt read the world's ground truth directly, so
every villager always knew where every resource, structure, and neighbor stood.
Each agent now carries a private `MentalMap` — explored terrain plus a list of
known place-facts with provenance and a last-seen tick — and target resolution,
the prompt's world description, and hail/talk resolution all read through it
instead. Two villagers with different histories now see different worlds.

## How it works

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
  since spec 044 US4 — `grave`) and the
  perception-gated resource kinds (`tree`/`forage`/`rock`/`water_edge`/`den`/
  `pile`). Spec 068's marsh/sand ground covers ([[worldmap-generation]])
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
`factHorizonVolatileTicks` (12 game-hours); every other kind gets
`factHorizonDurableTicks` (4 game-days). A stale fact stays stored (invisible
to resolvers/prompt) until a correction or the agent's death removes it —
staleness is never itself a removal. `PlaceFact.Fresh(now)` exports the same
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

**Resolution** ([[reflex-policy]], `internal/sim/path.go`/`policy.go`): every
resolver that targets a place now searches the acting agent's fresh facts, not
ground truth. `nearestKnown`/`nearestKnownAdjacentTo` (`path.go`) are
knowledge-gated twins of `nearest`/`nearestAdjacentTo` — the identical BFS
geometry and tie-breaking, only the match closure differs, so
"nearest known" keeps every ground-truth search's determinism. Availability
that is not itself place knowledge — a harvested forage spot, a cooling den,
wall damage, chest contents, quarry depletion — stays layered on top as an
ordinary ground condition, checked at arrival exactly as before; a fully
resolved goal that fails on arrival re-validates the same way any contested
resource always has. `talk_to`/`seek` resolves to the target's last KNOWN
sighting (`peerSightingOf`) — a stale sighting walks honestly to where the
target was last seen, and the landing/arrival guards (`GuardTargetPresent`)
cover a miss; liveness (`Dead`) stays a live check.

**The search goal** (US4, research D4): `search` is a new World tool
(`internal/tool/registry.go`, `Effect: World, Gate: Resolvable, Cost.DurationTicks:
0, PlanStep: true, ReflexEligible: true`, appended after `build_path` so no
existing tool's registration position shifts) resolving to `nearestFrontier`
(`path.go`) — the closest reachable tile the agent's map marks EXPLORED that
4-neighbors at least one UNEXPLORED in-bounds tile (Yamauchi-style), decoding
the explored bitmap once per search; not found means the reachable world is
fully explored, the search verb's honest exhaustion. Completion is
wander-class ([[executor]]) — instant on arrival, since the walk itself did
the exploring (movement marks explored terrain and the perception sweep
witnesses what's there). The reflex ladder ([[reflex-policy]]) falls back to
`search` on the hungry rung ONLY when the agent knows of no forage and no
ready den — hunger-only, so a fed villager never mounts an expedition just to
top up the larder. Spec 062 (US3, 057 audit Gap A) adds a second,
independent reflex call site: a cold NIGHT with no known warmth, insufficient
wood, and no known tree to chop also falls back to `search` — one rung above
terminal sleep in [[reflex-policy]]'s bounded frontier-search fallback — so
`nearestFrontier` now backs two separate reflex triggers (hunger, and
cold-with-nothing-left-to-try), each still bounded by the same
fully-explored-fails-honestly floor.

**The perception sweep** (`internal/sim/executor.go`, `perceptionEvents`, T007):
each awake living villager, on the same staggered per-agent cadence movement
uses (a fifth of a full per-tick sweep, T034's hot-path relief), diffs ground
truth within `witnessRadius` against its map and emits at most one
`agent.saw` (new/changed structures, piles, standing trees, unharvested
forage, unquarried rock, water shoreline, dens — fully baked, `Seen` = this
tick, provenance `witnessed`) and, when a remembered fresh fact is genuinely
ABSENT from ground truth (`groundFactPresent` — a chopped tree, a
quarried-out outcrop, a drained pile, a removed structure; a merely-harvested
forage spot or cooling den still exists, only its availability lapsed), one
`agent.map_corrected` (US3, T019) naming the gone facts. A correction's gone
facts each ride a companion situated first-person discovery memory
(`mapCorrectedText`, `salMapCorrected` = 5) in the same batch as
`agent.memory_added` — memories accrete only via that event, never appended
directly by a reducer arm (a deviation from data-model.md's "reducer stamps a
situated memory" phrasing, recorded for the planning tier). `agent.saw` is
digest-only, deliberately no chronicle line (too chatty) and not an absorb
trigger; `agent.map_corrected` IS an absorb trigger — [[agent-mind]]'s `absorb`
re-arms the planner only when a removed fact matches the agent's OWN current
intent target or resolved coordinates, so a correction elsewhere in the map
stays quiet, carried into the next scheduled round as a memory instead.

**Graves** (spec 044 US4, [[morgue]]): a death leaves a persistent marker —
the `agent.died` reducer arm ([[sim-state-reducer]]) appends
`Structure{Kind: "grave"}` at the death tile, the same reducer-internal idiom
as the inventory spill, unconditionally (the `Structures` slice has no
per-tile uniqueness invariant outside the `buildSite` gate on NEW builds, so
a grave coexists with whatever already stands there; appended last, it wins
the map view's per-tile glyph, and it blocks future building via `buildSite`'s
blanket any-structure check). No new knowledge machinery: `grave` is simply a
new entry in the closed vocabulary above (`placeFactKinds` in
`internal/tool/registry.go` mirrors it for `send_vision`'s `place_kind` Enum),
so the ordinary perception sweep witnesses a grave, talk can pass it on, a
vision can reveal it, and the prompt's landmark set (below) names it
individually.

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

**Directions exchanged in talk** (US5, research D5, [[social-fabric]]):
`tellablePlaces` (`internal/sim/executor.go`'s `talkEvents`) selects up to
`placeTellCap` (2) fresh facts per direction the other party lacks or holds
staler — freshest → nearest-to-listener → coordinate order — for EVERY
founded talk, hail-founded included, riding beside the rumor slot. One
`social.place_told{from, to, facts}` per direction, facts baked with `told`
provenance, the TELLER's `Seen` (secondhand is never fresher — staleness IS
the trust model), `Source` = the immediate teller. The reducer's `applySocial`
arm upserts into the RECEIVER's map only where absent or staler (a receiver's
own fresher knowledge never loses to secondhand); companion situated memories
on both sides ride the same batch at `salPlaceTold` = 3 (the talk band) —
"Told X about the fire at (x,y)." / "X told you of a fire at (x,y)."

**The divine reveal** (FR-014, [[guardian]]): `send_vision` carries an
OPTIONAL place-grant triple, `place_kind`/`place_x`/`place_y`, all riding
together — `internal/guardian/toolcalls.go`'s `parseReveal` refuses a partial
triple as a `rejected_gate` before anything lands. `place_kind`'s Enum is
`placeFactKinds` (`internal/tool/registry.go`) — this note's closed
vocabulary hand-mirrored, since `tool` must not import `sim`; a drift there
can only over- or under-offer the model, never land a false fact, since the
reducer dry-run (`groundFactPresent`) is the semantic authority that the
place is real. `landVision` composes one `metatron.place_revealed` event plus
a companion `agent.memory_added` ("The vision showed you the fire at
(x,y).", `SalDream`, `Origin: sim.OriginOmen`) as extra events riding the SAME
atomic `landNudgeBatch` call as the vision's own nudge memory — the grant
lands with the vision or not at all. The reducer arm stamps `Seen` (the
landing tick) and `Detail` (ground truth at landing, `groundFactDetail`)
NORMATIVELY — the model-side emitter cannot know either, so it bakes only the
place identity.

**Rendering the known world** (US2, contracts §3, [[agent-mind]]): the
decision prompt retires the old blanket "Village: `<first six structures>`"
line and the bare-distance nearby-agent scan in favor of `knownPlaces`
(`internal/mind/prompt.go`). Since spec 043 that section is one named block
of the assembled decision context ([[decision-context]]):
`internal/mind/context.go`'s `renderKnownPlaces` — the `known_places`
contract block — wraps `knownPlaces` plus the peer-sighting "Nearby" line,
and is droppable (priority 5) under context-budget pressure, unlike the
survival blocks. The content is the
acting agent's OWN mental map, never `State.Structures`: landmark structures
(fire/shelter/oven/chest, and since spec 044 `grave` — a death site is
exactly the individually-named, narratively-weighted place the landmark set
exists for, never grouped with the count+nearest resource kinds)
individually with provenance flavor (witnessed
plain, told naming the teller, revealed naming the vision; a fire the agent
remembers as burned out says so), everything place-shaped grouped per kind
with count + nearest (walls/paths/resources — grouping bounds the size a
long run without dropping information), one orientation line toward the
nearest unexplored land (`FrontierDirection`, silent on a fully-explored map),
and an explicit empty state ("You know of no fires or shelters yet.") — the
model must always be able to tell "I know none" from silence. The nearby-agent
line walks the map's `Peers` in agent-index order, rendering a remembered
position (even the asleep flavor is only ever added when the peer is
verifiably still there) rather than a live one. `State.MapDims()` sizes the
bitmap read without the `State` ever serializing the map itself.

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

[[reflex-policy]] is this note's primary consumer — every goal resolver and
the reflex ladder read through `nearestKnown`/`knowsAnyFresh`/
`warmKnownPredicate`/`peerSightingOf`, and `search`/`nearestFrontier` live in
its files. [[executor]] hosts the perception sweep (`perceptionEvents`) and
the talk sidecar (`tellablePlaces`) that grow and correct the map, plus the
derived explored/sighting bookkeeping on movement events. [[agent-mind]]
renders `knownPlaces` from it in the prompt — since spec 043 as
[[decision-context]]'s `known_places` block (`internal/mind/context.go`) —
and re-arms the planner on a targeted `agent.map_corrected`. [[sim-state-reducer]] carries `Agent.Map` and
the four knowledge-event Apply arms. [[event-types]] catalogs
`agent.saw`/`agent.map_corrected`/`social.place_told`/`metatron.place_revealed`.
[[social-fabric]] is where the place-telling sidecar rides, beside rumors and
gifts. [[guardian]] is the `send_vision` place grant's door, sharing this
note's closed vocabulary via `internal/tool/registry.go`'s `placeFactKinds`.
[[guardian-miracles]] shares the `rebaseTicks` SHIFT/KEEP taxonomy and is
what a miracle-moved villager's derived bookkeeping updates.
[[world-migration]] carries the v3→v4 knowledge grant; [[world-save-directory]]
is the format-version gate this bumped to v4. [[tui-client]] renders the four
new event types through the raw digest feed with no dedicated pane of its
own. [[chronicle]] narrates three of the four events (not `agent.saw`, too
chatty). [[tool-registry]] declares `search` and `send_vision`'s place-grant
params.

## Operational notes

A villager's knowledge only ever grows or corrects through recorded events —
there is no silent forgetting beyond the read-time freshness horizon, and a
stale fact is invisible to resolvers/prompt without being physically removed
until a correction (or death) does so. `internal/sim/mentalmap_test.go` is
the subsystem's own suite, alongside the v3→v4 migration, rebase-taxonomy,
determinism, and vision-place-reveal coverage [[testing-strategy]] tracks.
Exact freshness-horizon values are tuning (clarify Q5), soak-validated
rather than derived from first principles.
