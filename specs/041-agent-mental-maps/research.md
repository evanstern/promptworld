# Phase 0 Research: Per-Agent Mental Maps

**Date**: 2026-07-24 · **Spec**: [spec.md](spec.md) · **Corpus**:
`research/Agent-Mental-Maps/` (56-source grounding, 10 notes, commit c70c53f) ·
**Architecture facts**: verified against `main` (state/reducer/policy/mind seams; see
plan.md Technical Context)

All spec-level unknowns were resolved in the 2026-07-24 clarification session (see spec
Clarifications). This document records the *technical* decisions that turn those answers
into a design, each with rationale and rejected alternatives.

## D1. Representation: flat serialized state, tree as documented growth path

**Decision**: `MentalMap` is serialized as (a) an explored-tile **bitmap** (one bit per
tile, base64-encoded in JSON, sized W×H from the world manifest) and (b) a flat,
deterministically-ordered **place-fact list** (known dynamic entities with kind, position,
last-seen tick, provenance, source). No quadtree in the serialized state. The layered-3D
extension path (spec FR-016) and any future quadtree/octree index are documented in
data-model.md as derived structures behind the same accessors.

**Rationale**: the research corpus is unambiguous that at 64x64 the memory case for trees
evaporates (512 B/bitmap, 4 KiB/byte-layer; `Memory-Footprint-for-Many-Private-Maps`), that
the occupancy payload — not the container — is the load-bearing semantics
(`Occupancy-Grids-and-Belief-Maps`: explicit *unknown* state), and that the canonical
guidance is "start with your existing game world representation, then optimize if needed"
(Amit Patel, `Hierarchical-Pathfinding-Over-Known-Space`). Locally, replay doctrine wants
canonical JSON bytes with fixed field order — a flat bitmap + ordered list is trivially
canonical; a serialized tree is not. The tree family's actual payoff (compression of
coherent space, hierarchical pathfinding) arrives with bigger/multi-layer maps, and the
quadtree→octree lift changes only branching factor (`Hierarchical-Spatial-Trees`), so
deferring the tree costs nothing structurally.

**Alternatives considered**: serialized quadtree (rejected: canonical-bytes and
migration complexity now, zero memory/speed benefit at this scale); per-cell log-odds byte
grid (rejected: probabilistic occupancy is overkill when perception is exact within radius —
promptworld perception has no sensor noise; three-state explored/known/unknown suffices);
storing knowledge only as situated memories (rejected: memories are a salience-decayed
narrative window, not a queryable index; resolvers need O(1) deterministic lookup).

## D2. Write paths: silent derived explored-bits; explicit events for facts

**Decision**: two write paths, split by observability need.
- **Explored bitmap**: updated *derivationally inside existing reducer arms* (movement,
  spawn, teleport-miracle): when an agent's position changes, mark tiles within the
  perception radius explored. Pure function of (state, event) → replay-identical; no new
  events; no chronicle noise.
- **Place-facts**: mutated *only by new recorded events*. The executor's per-beat sweep
  compares each awake agent's map against ground truth within perception radius and emits
  diffs: `agent.saw` (new/changed fact witnessed) and `agent.map_corrected` (remembered fact
  absent in reality). Talk transfer emits `social.place_told`; divine grant emits
  `metatron.place_revealed`. Reducer arms apply them to the target agent's map.

**Rationale**: matches the two existing precedents exactly — derived silent bookkeeping
(relation/idle stamps) vs. baked-at-emission observable events (situated memories,
"context is baked at emission, never re-derived"). Fact changes are the narratively
meaningful knowledge events spec FR-017 requires (discover/correct/told/revealed → chronicle
+ TUI digest); explored-bits are high-frequency and meaningless individually. Emitting
explored-tile events per move step would flood the event log (a move event every 5 ticks per
agent); deriving facts in the reducer without events would hide corrections from the
chronicle and from `mind.absorb` triggers.

**Alternatives considered**: all-events (rejected: log bloat, ~17×17 tile payloads per
step); all-derived (rejected: FR-017 unobservable, no absorb triggers, no digest rows);
mind-side map (rejected outright: violates model isolation — the map gates *reducer-side*
resolution, so it must be reducer state; the mind gets a plan-time snapshot like the
journal).

## D3. Gating mechanics: filter the match predicates, keep BFS geometry

**Decision**: `resolveGoal`/`goalResolvers` and `decideIntent` (reflex — full parity per
clarify) take the acting agent's map into their match predicates: `nearest`/
`nearestAdjacentTo` keep their deterministic BFS geometry over ground-truth passability, but
the `match` closure returns true only for candidates the agent *knows* — a place-fact of the
right kind at that position (structures, piles, resource sites, dens) or, for `talk_to`, a
known/last-seen agent position. Terrain passability (walkability, water-as-obstacle) stays
common knowledge, exactly as clarified. On resolver failure the landing ladder returns the
knowledge-flavored rejection `"rejected_guard: you know of no <kind>"` — distinct from the
existing `"no <kind> reachable"` (spec FR-004; subsumes TASK-95's distinction for gated
verbs).

**Rationale**: smallest possible seam — predicates already flow through two helpers; the
BFS itself is knowledge-neutral geometry. Keeping passability ground-truth avoids the
deadlock/starvation class the research flags for movement-gated designs
(`Fog-of-War-and-Limited-Perception`) and matches the clarify decision. Arrival
re-validation (the existing contested/re-validate pattern at `executeAtTarget`) is the
staleness correction moment: work against a vanished target fails, the perception sweep
emits `agent.map_corrected`, and the planner re-arms — US3 falls out of existing machinery.

**Alternatives considered**: post-filtering candidate lists (rejected: `nearest` early-exits
by BFS distance; filtering after breaks nearest-known semantics); a parallel known-space BFS
(rejected: duplicates path.go for no gain at this scale); gating `nextStep` movement
(rejected by clarify Q1).

## D4. Search verb: frontier BFS, honest exhaustion

**Decision**: new villager tool `search` → goal `"search"` resolver: BFS from the agent
over passable ground-truth terrain to the nearest tile that is explored, passable, and
adjacent to ≥1 unexplored tile (the frontier, per `Frontier-Based-Exploration`); intent
lands there; arrival is an instant goal (like wander) — the perception sweep then expands
the map and the planner re-arms with fresh knowledge. If no reachable frontier exists
(map fully explored), the resolver errors `"rejected_guard: nothing left unexplored"` —
honest exhaustion (FR-009/AC). `wander` is untouched (FR-010). The reflex ladder may fall
back to `search` on its get-food rung when no food source is known.

**Rationale**: Yamauchi-style frontier selection is the standard, terminates by
construction (explored set grows monotonically), and reuses `nearest` with a frontier
predicate — deterministic tie-breaking for free. At 4,096 tiles a full-scan frontier check
is trivially cheap (corpus: millions of cell updates/sec envelope).

**Alternatives considered**: information-gain frontier ranking (rejected: needless
sophistication at this scale; nearest-frontier is deterministic and legible); biasing
`wander` toward unknown (rejected: FR-010 keeps wander distinct; a modal behavior is harder
to narrate and test).

## D5. Talk transfer: deterministic sidecar to the rumor slot

**Decision**: in `talkEvents`, beside the existing `TellableFor` rumor exchange, run a
deterministic place-knowledge exchange: each participant offers up to **2** place-facts the
other lacks (or holds staler), selected by fixed criteria (freshest first, then nearest to
the listener, then coordinate order); emit `social.place_told` per direction. Received facts
land with provenance `told`, `Source` = teller, and the *teller's* last-seen tick (never
fresher). Every completed talk runs the exchange (clarify Q3: automatic, no LLM tool).

**Rationale**: `TellableFor` is the proven template — deterministic selection from private
state during the social beat, one event, reducer applies to the listener. Carrying the
teller's last-seen tick makes secondhand knowledge inherently staler (decays sooner), which
implements "lower initial confidence" without a separate trust scalar — provenance +
staleness are the trust model, matching the corpus's reliability-conditioning direction
(`Multi-Agent-Map-Sharing`, `Belief-Decay-and-Stale-Knowledge`) with village-scale
simplicity.

**Alternatives considered**: explicit share tool (rejected by clarify); merging full maps on
talk (rejected: instant knowledge homogenization destroys the information asymmetry the
feature exists to create; corpus notes team-vision pooling is the degenerate case);
noisy-OR/confidence-scalar merge (rejected: promptworld perception is noiseless; timestamp
staleness already orders evidence).

## D6. Decay: read-time freshness horizon, no mutation

**Decision**: place-facts are never mutated by time. Staleness is evaluated at *read* time:
a fact older than its kind's freshness horizon no longer satisfies resolver predicates or
prompt rendering (drops to "unknown again"), with volatile kinds on short horizons (lit-fire
status ≈ hours) and durable kinds on long ones (buildings ≈ many game days). Horizons are
named constants beside the existing cadence constants. Facts are physically removed only by
`agent.map_corrected` (perception found them gone) or agent death.

**Rationale**: replay doctrine punishes per-tick mutation sweeps (every decay step is state
churn in snapshots); read-time evaluation is the established pattern (`KnownRumor`
confidence-decay, belief half-life anchors). Decay-toward-unknown (never toward false) is
the corpus's central staleness rule (`Belief-Decay-and-Stale-Knowledge`). Exact constants
are tuning (clarify Q5) — set at implementation, soak-validated.

**Alternatives considered**: periodic decay sweep events (rejected: log noise + snapshot
churn); no decay (rejected: FR-012; permanently-known fires make staleness meaningless).

## D7. Seeding & migration: genesis grant + v3→v4 transform

**Decision**:
- **New worlds**: `NewState` genesis (pure function of seed+map, never event-replayed) marks
  each agent's spawn surroundings explored (radius = perception radius; worlds start with
  zero structures, so no facts to grant). Villagers begin knowing their landing area and
  nothing else.
- **Existing worlds**: bump `FormatVersion` 3→4; the `promptworld migrate` transform grants
  each living agent (a) explored tiles around their current position and (b) witnessed
  place-facts for **all current structures and piles** — they have lived in this village;
  amnesia would be both unfair and viability-threatening. Rides the established
  `world.migrated` full-state event.

**Rationale**: genesis-side seeding keeps tick-0 event logs clean and is replay-safe by
construction (genesis is reconstructed from seed, not replayed); the migration transform is
the exact precedent shipped for prior format bumps. Generous migration knowledge matches
spec edge-case "cold start" viability and the fiction (natives, not strangers).

**Alternatives considered**: tick-0 seed events at `promptworld new` (workable precedent
exists, but adds log noise and a second seeding path to keep in sync — rejected);
`omitempty`-only additive field without format bump (rejected: loaded old worlds would have
all-nil maps → every villager knows nothing → mass starvation; semantics of existing worlds
change, which is exactly what format bumps are for).

## D8. Prompt rendering: grouped known-places section, complete for structures

**Decision**: `userPrompt` replaces the `Village:` line with a known-places section built
from the agent's map (fresh facts only, D6): **all** known structures individually
("fire at (12,34) — seen recently; shelter at (8,20)"), resource knowledge grouped by kind
with count + nearest ("you know 3 forage spots, nearest (22,41)"), and a one-line
unexplored orientation ("land to the north-east is unknown to you"). Secondhand facts are
marked ("Birch told you of a fire at (40,12)"). No fixed truncation of structures —
grouping, not dropping, bounds the size (retires the first-6 bug; SC-002).

**Rationale**: rendered-text-only surface per clarify Q4; grouping keeps prompt size
bounded on a 64x64 world without violating the all-known-structures criterion; provenance
phrasing feeds the social/narrative layer the feature exists for.

**Alternatives considered**: read-only map query tools (deferred by clarify Q4); rendering
raw fact rows (rejected: prompt-bloat and robotic voice — existing prompt style is prose).

## D9. New event types and their gates

**Decision**: four new event kinds — `agent.saw`, `agent.map_corrected`,
`social.place_told`, `metatron.place_revealed` (Metatron's `send_vision` gains an optional
place grant). Each ships with: payload struct (canonical field order), reducer arm,
emitter, TUI digest registry entry + catalog fixture row, `docs/wiki/event-types.md`
backticked row, `chronicleNote` grammar for the narratively-loud ones (corrections,
tellings, reveals; `agent.saw` stays digest-only), and absorb triggers where the planner
should re-arm (correction of the acting agent's current target). Situated memories
accompany discovery/correction/told/revealed at appropriate salience so knowledge events
enter the memory window (FR-017).

**Rationale**: this is the test-enforced checklist for any new event in this codebase
(`TestCatalogSweep` cross-checks registry, fixtures, and wiki doc); enumerating it in Phase
0 prevents the known failure mode (TASK-92/94 class: event added, catalog red).

**Alternatives considered**: one generic `agent.map_changed` event (rejected: digest/
chronicle need per-kind voice; provenance is the point).
