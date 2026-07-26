# Data Model: First-person harvest memory (spec 081)

No new persisted entities and no new event types. The feature changes when an
existing entity (the place fact) leaves an existing collection (the mental
map), and adds two salience constants + two memory texts.

## Entities (existing, behavior extended)

### PlaceFact (`internal/sim/mentalmap.go`)
- **Fields** (unchanged): `Kind` (closed vocabulary: tree, rock, forage,
  water_edge, den, pile, structure kinds, grave), `X`, `Y`, `Seen` (tick),
  `Prov` (`witnessed` | `told` | `revealed`), optional `Detail`.
- **Invariant** (unchanged): at most one fact per (Kind, X, Y), canonical
  (Kind, X, Y) sort order; drained list goes nil.
- **New exit path**: act-time removal — see transitions below.

### MentalMap (`internal/sim/mentalmap.go`)
- Per-villager `Facts []PlaceFact`. Grown by `agent.saw` / `social.place_told`
  / `metatron.place_revealed`; shrunk today only by the `agent.map_corrected`
  reducer arm via `removeFact`; after this feature also shrunk by the
  `agent.chopped` / `agent.quarried` arms via the same `removeFact` primitive.

### Memory (`internal/sim/memory.go`)
- Unchanged shape; accretes only via `agent.memory_added`.
- **New constants**: `salChop = 4`, `salQuarry = 4` (the `salHunt` band —
  memorable, below every generation-interrupting / rumor-seed threshold).
- **New texts** (first-person, situated, origin action):
  - chop → `"Felled the tree at (%d,%d)."`
  - quarry → `"Quarried the outcrop at (%d,%d)."`

## State transitions

### On `agent.chopped{agent: i, x, y}` (reducer arm, state.go)
1. Existing mutations unchanged (yield, axe wear, intent clear,
   `s.Cleared += (x,y)`).
2. **New**: `Agents[i].Map.removeFact("tree", x, y)` (no-op if absent).
3. **New**: for every `w ≠ i` with `!Dead && !Asleep && Map != nil` and
   `abs(X-x)+abs(Y-y) <= witnessRadius` — positions read from the same
   pre-mutation state as the yield derivation:
   `Agents[w].Map.removeFact("tree", x, y)`.

### On `agent.quarried{agent: i, x, y}` (reducer arm, state.go)
Identical shape with kind `"rock"` and `s.Quarried += (x,y)`.

### On chop/quarry emit (executor, same batch)
- **New**: companion `agent.memory_added` for the actor
  (`situatedMemoryEvent`, salience `salChop`/`salQuarry`), riding the same
  event batch as the act. Witnesses get NO memory event.

### Perception sweep (`perceptionEvents`) — unchanged
- Correction half still emits `agent.map_corrected` + discovery memory for
  any agent whose map retains a fresh fact absent from ground truth — after
  this feature that set is exactly the agents absent/asleep at act time
  (their facts were not removed).

### Mind absorb (`internal/mind/mind.go`) — extended
- `agent.chopped` / `agent.quarried`: in addition to arming the actor
  (existing), arm any agent within `witnessRadius` of (x,y) whose live
  intent has `(TargetX,TargetY) == (x,y)` or `(ResX,ResY) == (x,y)` —
  `agent.map_corrected` parity for the re-arm-on-lost-premise rule.

## Validation rules

- Removal matches on (kind, x, y) only — provenance-blind (hearsay edge case).
- Only awake living villagers with a non-nil map participate as witnesses
  (perception parity; dead/asleep keep their facts).
- Determinism: all removals derive from (event payload, pre-mutation state);
  no wall-clock, no randomness, no executor-only knowledge.
