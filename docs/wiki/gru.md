---
name: gru
description: The nocturnal sight-triggered predator — an event-sourced entity that wounds the healthy but can finish the already-weakened (spec 044 escalation); fire light and shelter are absolute safety; encounters seed rumors and omens
kind: component
sources:
  - internal/sim/gru.go
verified_against: 9b4ed5aef5bfea50b67fac10f8e2153f065a814d
---

# The gru

The gru (TASK-10) makes night dangerous the Zork way. It is an **entity, not a
phenomenon** — a positioned body in event-sourced state (`State.Gru`, nil
while not abroad) — because sight-triggering needs geometry, the TUI needs something
to render, and rumors need something to have been seen. Its design contract:
it wounds the healthy; since spec 044 (US3) it can also *finish* the
already-weakened. A healthy villager always survives a single attack (the
survival floor holds), but one already below the near-death band can be
killed outright — lethality traces back to a preventable spiral (no fire, no
food, wounds untreated), never a fresh random roll. Deaths from neglect (starvation/exposure/collapse) remain the
heartbeat's; a gru kill is the one death emitted from `gruStep` itself.

## How it works

**Lifecycle**: at 22:00 a seeded per-night roll (`rngAt("gru-emerge")` against
`s.GruEmergePerMille()`, default 600 per mille — a per-world [[world-tuning]]
dial since spec 048, `defaultGruEmergePerMille` in `tuning.go`) decides
whether it comes out; if so it slips in from a seeded passable, unlit border
tile (`gru.emerged{night, x, y}`). Since spec 054 the roll is skipped
ENTIRELY when the armed incident schedule lands a `gru_emerges` entry that
same night (`gruScheduledTonight`) — the schedule preempts the dice, never
two spawn mechanisms in one night, and the skip consumes no RNG draw
(`rngAt` is coordinate-seeded, no stream), so unscheduled nights roll
exactly as before. At
06:00 it is gone (`gru.withdrew{day}`, state nil). Every decision is a pure
function of (seed, night/tick) — [[deterministic-rng]] — so the whole
predator replays.

**Sight**: it sees live agents within Manhattan `gruSightRadius` (8) — **unless
they are protected**. Protection is fire light (`gruLightRadius = 3`, strictly
wider than `warmAt`'s fire radius 2, so a warm agent is always a safe
agent) or a shelter tile. The gru also never steps into protected tiles, so it
visibly circles the firelight. Protection is absolute, not probabilistic, and
sits upstream of target selection, unchanged by the spec 044
escalation (FR-016): a protected agent can never even be attacked, so never
killed.

**Decomposed for read-only export** (spec 074, TASK-142): `litAt` is now a
thin wrapper over a private `lightSource(s, x, y) (lit bool, source string)`
core in this file — same scan, same radius, byte-for-byte the same
predicate — so `internal/sim/env.go`'s exported `EnvAt(s, x, y, tick)` can
report a tile's light truth for the TUI's look-cursor TILE view without a
second, drift-prone radius constant. `warmAt`'s sibling
decomposition lives in [[executor-world-state]] (`terrain.go`'s
`warmthSource`). A pure read-side export; neither reducer nor tuning
behavior changed.

**Movement** (one tile per `gruMoveEveryTicks = 4`, slightly faster than
agents' 5): greedy chase toward the nearest visible agent (ties to the
lowest index), seeded prowl (`rngAt(seed, "gru-prowl", tick, 0)`) when
nobody is visible. Deliberately greedy rather than BFS — a monster that can
be baffled by water and firelight is the right monster. The decision lives
in ONE home, `gruMoveDecision` (gru.go), shared by two shapes since spec
104 (ruling 4): **legacy worlds** emit `gru.moved{x, y}` per beat exactly
as always (arm retained forever — old logs replay unchanged); under the
**coalescing regime** ([[world-tuning]]'s `needs_checkpoint_minutes`
marker) NO `gru.moved` is emitted at all — the derived-advancement engine
([[sim-state-reducer]], `advance.go`) runs the same decision at each beat
tick over the advanced state (agent steps first within the tick, then the
gru — an attack recorded at the beat tick precludes the move, exactly the
emitter's exclusivity), behind the marshaled `Gru.Done` watermark. The
cadence, the "gru-prowl" purpose string, and the stalk/protection geometry
are therefore replay-load-bearing for coalesced logs
([[sim-state-reducer-replay-hazards]]) — a retune requires the spec-094
format machinery. `gru.emerged`/`withdrew`/`sighted`/`attacked` remain
events either way (rare, story-bearing).

**Attack** (`gru.attacked{agent, health}`): adjacent + visible + a
10-game-minute cooldown (`gruAttackCooldown = 600`). The payload carries the
**absolute post-wound health** (outcome convention), a `gruWound = 250` drop —
and since spec 044 (US3) the floor is conditional on the victim's raw
pre-attack health. At or above `nearDeathBelow` (200) the drop floors at
`gruWoundFloor = 1`: a healthy target is wounded, never executed. Below
`nearDeathBelow` the floor is 0: the attack can kill outright. The predicate is
deliberately the raw pre-attack health, never the hysteresis-latched
`Agent.NearDeath` bool (research R4) — an agent recovered past the band is
not "already weakened". A compile-time invariant
(`const _ uint = gruWound - nearDeathBelow`) pins
`gruWound >= nearDeathBelow`, so an escalated hit against a weakened target
lands at exactly 0, never a positive remainder. When the post-wound health is 0, `gruStep` emits the standard
`agent.died{cause: "gru"}` immediately after the attack, in the same batch, so
the entire existing death path (reducer `Dead` flag, inventory spill,
`State.Deaths` ledger, grave, chronicle, [[morgue]]) runs unchanged
(FR-015); gru attacks land off the %60 needs heartbeat, so the executor's
witness-death block never runs for them and `gruStep` replicates that idiom
inline (a "Watched X die of the gru" witness loop over agents within
`witnessRadius`). For a surviving victim, the reducer arm (`applyGru`) wakes
them and clears their intent, handing them to [[reflex-policy]], which at
night flees to warmth — the night curfew, emergent rather than scripted. The
heartbeat's near-death memory names "the gru" as the cause when the last wound
was recent (`LastVictim` / `LastAttack` on the `Gru` struct).

**Story fuel**: the victim keeps a salience-9 memory; awake witnesses within
`witnessRadius` keep a subject-tagged, tone-negative memory about the victim
(salience 7 ≥ `rumorMinSalience`), which [[social-fabric]]'s `TellableFor` serves
as gossip — a witnessed attack becomes a village-wide rumor with mutating
confidence. Any awake agent within sight range — safe ones by the fire included —
gets one `gru.sighted{agent, x, y}` plus an omen memory per night (a `Seen`
bitmask on the `Gru` struct latches it). Sighting, attack, and witness memories
are situated (spec 019): each is built with `situatedMemoryEvent`/
`situatedMemoryAboutEvent` and `Where` set to the remembering agent's own
tile at the moment of the memory — never the gru's position. Since spec 030
each also carries the required `origin` provenance stamp: the sighter's and
the victim's own memories (sighting, attack) stamp `OriginAction` (their own
direct experience), while the nearby witness memory stamps `OriginWitness`,
matching every other witness site in the sim. [[event-types]] catalogs the
family.

**Shared protection predicates (spec 077):** `gruProtected` (fire light +
shelter, `litAt`'s wider-than-warmth radius) is no longer the gru's alone —
the stranger ([[event-types-scenario-incidents]], `internal/sim/stranger.go`)
reuses it verbatim for both its movement avoidance and its entry-tile
validity (`strangerEntryValid`), so "light is safety" is one rule with two
consumers, never two copies. The two entities are independent: separate
latches (`State.Gru` / `State.Stranger`), legal on the same night, and the
stranger takes goods where the gru takes health.

## Connections

[[executor]] calls `gruStep` from `stepEvents` (same purity contract) and
hosts the run-end detection an escalated kill can trip (a kill felling the
last living villager lands in the same batch as `run.ended`);
[[sim-state-reducer]] dispatches `gru.*` to `applyGru` and carries the death
fallout (ledger, grave); [[morgue]] mourns gru deaths under the charter
revision then in force; [[reflex-policy]] supplies the flee-to-warmth response;
[[social-fabric]] turns witness memories into rumors; [[tui-client]] renders it
as a red G; [[worldmap-generation]] bounds its spawn border; [[scenario-machinery]]
shares the night-emergence preemption check (`gruScheduledTonight`) and
owns the scheduled emission; [[tui-dock-tile-view]] (the look-cursor TILE view) reads
`sim.EnvAt` (this file's `lightSource` core) for a tile's light level/note.

## Operational notes

Live proving (seed 42, 1257 game days, run under the pre-044 unconditional
floor): the gru emerged ~60% of nights, attacked 186 times, never killed —
every wound left its victim at 750 (healthy victims fare identically
today — escalation only reaches targets below 200) — and
emitted zero events outside 22:00–06:00; a witnessed attack on Fern
propagated village-wide with confidence decaying 80 → 35.
Sightings are personal (subject −1) and thus omen material, not gossip;
only *witnessed attacks* seed rumors, which makes gru rumors appropriately
rare.
