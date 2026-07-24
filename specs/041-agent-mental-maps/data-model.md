# Phase 1 Data Model: Per-Agent Mental Maps

**Date**: 2026-07-24 · **Plan**: [plan.md](plan.md) · **Decisions**: [research.md](research.md)

## Core entities

### MentalMap (new; one per agent, reducer-owned)

Attached as `Agent.Map *MentalMap` with `json:"map,omitempty"` — pointer + omitempty so
pre-feature snapshots round-trip byte-identically (Journal precedent).

| Field | Type | JSON | Semantics |
|---|---|---|---|
| Explored | string | `explored` | Base64 of a W×H bitset (row-major, bit set = tile explored). Sized from the world map at creation; ~512 B raw at 64x64 |
| Facts | []PlaceFact | `facts,omitempty` | Known dynamic entities, deterministic order (see ordering rule) |

**Ordering rule**: `Facts` is kept sorted by (Kind, X, Y) at every reducer mutation —
canonical JSON bytes regardless of discovery order; upsert = binary-search insert/replace.

**Invariants**:
- A fact's position is always within map bounds; at most one fact per (Kind, X, Y).
- `Explored` bits only ever set (monotone) — *knowledge of terrain shape* never un-explores;
  facts are what go stale/removed.
- Dead agents: map retained in state for historical fidelity but excluded from all read
  paths (resolvers, prompts, talk transfer) — no posthumous knowledge channels.

### PlaceFact (new)

| Field | Type | JSON | Semantics |
|---|---|---|---|
| Kind | string | `kind` | Closed vocabulary: structure kinds (`fire`, `shelter`, `oven`, `chest`, `wall_plank`, `wall_stone`, `path`) + resource kinds (`tree`, `forage`, `rock`, `water_edge`, `den`, `pile`) |
| X, Y | int | `x`, `y` | Tile position |
| Seen | int64 | `seen` | Game tick the fact was last perceived by the *original observer* (talk transfer copies the teller's value — secondhand is never fresher) |
| Provenance | string | `prov` | `witnessed` \| `told` \| `revealed` (aligned with the existing Belief vocabulary) |
| Source | int | `src,omitempty` | Teller agent index for `told`; -1/omitted otherwise |
| Detail | int64 | `detail,omitempty` | Kind-specific scalar baked at emission (fires: FuelUntil as last seen; piles: 0) — display/decay input, never re-derived |

**Freshness (read-time, D6)**: a fact is *fresh* iff `now - Seen < horizonFor(Kind)`.
Volatile horizons (lit-fire detail ≈ game-hours) vs durable (structure existence ≈ many
game-days); named constants beside `PlannerCadenceTicks`. Stale facts are invisible to
resolvers and prompts; only `agent.map_corrected` or death physically removes facts.

**State transitions** (fact lifecycle):

```
(unknown) --agent.saw / social.place_told / metatron.place_revealed--> known(fresh)
known(fresh) --game time > horizon--> stale (invisible; still stored)
stale --agent.saw (re-perceived)--> known(fresh, Seen refreshed)
known|stale --agent.map_corrected (perceived absent)--> (removed)
```

### Agent (modified)

`Map *MentalMap` added after `Journal` (append-only field position; snapshot
byte-stability twin test required, per `TestAxesOmitemptyStable` precedent).

## New event payloads (canonical structs, not maps)

### `agent.saw` — perception upserts facts (emitter: executor perception sweep)

```go
type SawPayload struct {
    Agent int         `json:"agent"`
    Facts []PlaceFact `json:"facts"` // new or changed within perception radius this beat; bounded by radius
}
```

Reducer: upsert each into `Agents[Agent].Map.Facts` (provenance `witnessed`, Seen = event
tick). Digest-only (no chronicle line). Not an absorb trigger.

### `agent.map_corrected` — perception found remembered facts gone

```go
type MapCorrectedPayload struct {
    Agent int         `json:"agent"`
    Gone  []PlaceFact `json:"gone"` // the remembered facts, as remembered (for narration)
}
```

Reducer: remove matching facts; stamp a situated memory ("The fire at the woods was cold
and dead when you looked.", Origin witness). Chronicle grammar + digest row. Absorb trigger
when a removed fact matches the agent's current Intent target (re-arm planner).

### `social.place_told` — talk transfer (emitter: talkEvents sidecar, D5)

```go
type PlaceToldPayload struct {
    From  int         `json:"from"`
    To    int         `json:"to"`
    Facts []PlaceFact `json:"facts"` // ≤2 per direction; Seen = teller's Seen; Provenance told; Source = From
}
```

Reducer: upsert into receiver's map only where absent-or-staler; situated memories both
sides ("Told Birch about the fire by the rock."/"Birch told you of a fire at…"). Chronicle
grammar + digest row.

### `metatron.place_revealed` — divine grant (emitter: send_vision extension)

```go
type PlaceRevealedPayload struct {
    Agent int         `json:"agent"`
    Facts []PlaceFact `json:"facts"` // Provenance revealed
}
```

Reducer: upsert; situated memory Origin omen. Chronicle grammar + digest row. Must be added
to `injectSocialWhitelist` (rides the Metatron mind path).

## Implementation addenda (accepted deviations, 2026-07-24)

Recorded during US1 review; these refine the model above and are the as-built truth.

- **PeerSighting** (new sub-entity): agent positions are NOT place-facts. `MentalMap.Peers
  []PeerSighting{Agent, X, Y, Seen}` (sorted by agent index, `peers,omitempty`), maintained
  *derivationally* (D2's silent class) from movement/wake reducer arms — mover and awake
  living bystanders within perception radius record each other. Rationale: (Kind,X,Y)-keyed
  facts for constantly-moving agents would leave trails demanding saw/corrected event churn
  per move — the log-flood D2 rejected. `talk_to`/`seek` resolve to the last sighting.
- **Availability vs existence split**: *existence* of a place is belief-gated (the map);
  *availability at a known place* (harvest/cleared/quarried overlays, pile presence, den
  readiness, wall damage, chest contents) stays a ground condition in the resolver's `also`
  closure. Belief-pure availability livelocked the reflex (a just-harvested tile's fact
  persists until US3 correction → village re-targets the nearest depleted spot forever;
  soak 0/8 survivors). Structures remain belief-pure; arrival re-validation is the
  correction moment (US3). Fire lit-ness is predicted from the remembered `Detail`
  (FuelUntil as last seen), never read live.
- **`PlaceFact.Source`**: 0-omitted, meaningful only under `Provenance=="told"` (a -1
  sentinel cannot ride `omitempty`).
- **Migration grants all agents (incl. dead) non-nil maps**: snapshot-over-genesis merge
  semantics would otherwise resurrect genesis maps for map-absent dead agents, breaking
  byte-agreement between snapshot-load and from-genesis replay.
- **Death knowledge** is out of scope for the fact model (live `Dead` check retained at the
  door); flagged as possible follow-up.
- **Correction memories ride as companion `agent.memory_added` events** (same batch as
  `agent.map_corrected`), not a reducer-side append — the memories-accrete-via-events
  invariant forbids arms appending `Agent.Memories` directly.
- **Correction scope = absence, not availability lapse**: chopped trees, quarried-out rock,
  drained piles, and removed structures correct; harvested forage, cooling dens, and
  shorelines persist (the place still exists). A burned-out fire is *not* corrected — the
  structure persists; lit-ness staleness rides the remembered `Detail`, refreshed on sight,
  and the prompt renders "(likely burned out by now)".
- **Walls and paths render grouped in the prompt** (count + nearest), not individually —
  per-tile runs would bloat the prompt; landmark kinds (fire/shelter/oven/chest) render
  individually with no cap, which is what SC-002 measures.
- **`metatron.place_revealed` stamps normatively in the reducer**: the emitter bakes place
  identity only; the arm stamps `Seen` (landing tick), `Provenance` (revealed), and
  `Detail` (ground truth at landing) — the model-side emitter cannot know the landing tick
  under InjectSocial re-stamping, and the stamps are a pure function of (state, event).
  The arm validates rather than clamps: a vision naming a false place rejects the whole
  batch at the dry-run door (the god reveals what is).
- **Polish perf relief (T034)**: per-beat local overlay sets, `groundFactPresentIn`, and a
  `MarkExplored` no-change fast path — emission order and payload bytes proven unchanged.
  Known honest costs left for follow-up: gated resolvers run more BFS on
  know-something-reach-nothing cases; the InjectSocial dry-run's full-state round-trip got
  heavier with fact-bearing state (candidate: structured State clone).
- **US3 re-examination of the availability split (accepted)**: permanent-availability
  ground conditions (forage/den/chest/wall) stay load-bearing forever; absence-class
  conditions (tree/rock/pile) became redundancy-with-benefits under correction — kept as
  shipped; relaxing them is a T035 tunable if the soak wants stronger epistemics.

## Modified behaviors (no new entities)

| Seam | Change |
|---|---|
| `goalResolvers` match predicates | candidates require a fresh fact of the right kind at (x,y) in the acting agent's map (gated kinds); terrain passability untouched (D3) |
| `decideIntent` reflex ladder | same predicates (full parity); get-food rung may fall back to `search` |
| `search` (new goal + tool) | nearest frontier tile: explored ∧ passable ∧ adjacent-to-unexplored; instant goal; exhaustion error when no frontier (D4) |
| `talk_to`/`seek` resolver | target position from map (last-known) instead of live coordinates; existing landing guard covers misses |
| Movement reducer arms | derived explored-bit marking within perception radius on position change (D2) |
| `userPrompt` | known-places section per D8; first-6 truncation retired |
| Genesis / migration | spawn-area explored at `NewState`; v3→v4 migrate transform grants explored area + witnessed facts for existing structures/piles (D7) |

## Layered-3D extension path (spec FR-016 — documentation, not implementation)

The representation extends to stacked grid layers (the planned verticality model:
independent W×H levels joined at portal tiles — stairs/ladders) without changing cell or
fact semantics:

1. **Explored**: one bitmap **per layer** (`Explored []string`, index = layer). The 2D case
   is the degenerate single-layer array. Bits stay monotone per layer.
2. **PlaceFact** gains `Z int` (layer index), zero-valued and omitted in 2D — additive,
   snapshot-stable. Fact ordering extends to (Kind, Z, X, Y).
3. **Portals are place-facts** (`Kind: "stair_up"/"stair_down"` at their tile): an agent can
   only route across layers through portals it *knows*, which is exactly the corpus's
   floor-stair topology (per-layer metric maps + portal graph;
   `Extending-to-3D-and-Layered-Grids`).
4. **Search generalizes per-layer**: frontier BFS runs on the agent's current layer;
   known portals join the search graph as edges, so "search the cellar" emerges from the
   same nearest-frontier rule over the portal-connected known space.
5. **Growth valve**: if layers multiply until flat bitmaps hurt, the serialized form can
   move to a quadtree-per-layer (or octree) behind the same accessors — the corpus
   establishes the payload lifts unchanged (log-odds/tri-state in quadtree/octree; OctoMap
   precedent). Nothing in the event contracts changes: payloads carry facts and positions,
   never representation.
