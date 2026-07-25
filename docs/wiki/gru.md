---
name: gru
description: The nocturnal sight-triggered predator — an event-sourced entity that wounds the healthy but can finish the already-weakened (spec 044 escalation); fire light and shelter are absolute safety; encounters seed rumors and omens
kind: component
sources:
  - internal/sim/gru.go
verified_against: cea7b8f83fa07f9fcfefe4dd861aa05a78448f1b
---

# The gru

The gru (TASK-10) makes night dangerous the Zork way. It is an **entity, not a
phenomenon** — a positioned body in event-sourced state (`State.Gru`, nil while it
is not abroad) — because sight-triggering needs geometry, the TUI needs something
to render, and rumors need something to have been seen. Its design contract:
it wounds the healthy; since spec 044 (US3) it can also *finish* the
already-weakened. A healthy villager always survives a single attack (the
survival floor holds), but a villager whose health has already fallen below
the near-death band can be killed outright — lethality traces back to a
preventable spiral (no fire, no food, wounds untreated), never a fresh random
roll. Deaths from neglect (starvation/exposure/collapse) remain the
heartbeat's; a gru kill is the one death emitted from `gruStep` itself.

## How it works

**Lifecycle**: at 22:00 a seeded per-night roll (`rngAt("gru-emerge")` against
`s.GruEmergePerMille()`, default 600 per mille — spec 048 promotes this to a
per-world [[world-tuning]] dial, the default living in `tuning.go` as
`defaultGruEmergePerMille`) decides whether it comes out; if so it slips in
from a seeded passable, unlit border tile (`gru.emerged{night, x, y}`). At
06:00 it is gone (`gru.withdrew{day}`, state nil). Every decision is a pure
function of (seed, night/tick) — [[deterministic-rng]] — so the whole
predator replays.

**Sight**: it sees live agents within Manhattan `gruSightRadius` (8) — **unless
they are protected**. Protection is fire light (`gruLightRadius = 3`, strictly
wider than `warmAt`'s fire radius 2, so a warm agent is always a safe agent) or
standing on a shelter tile. The gru also never steps into protected tiles, so it
visibly circles the firelight. Protection is absolute, not probabilistic — and
these rules sit upstream of target selection, unchanged by the spec 044
escalation (FR-016): a protected agent can never be killed because it can
never be attacked at all.

**Movement** (`gru.moved{x, y}`, one tile per `gruMoveEveryTicks = 4`, slightly
faster than agents' 5): greedy chase toward the nearest visible agent (ties to
the lowest index), seeded prowl when nobody is visible. Deliberately greedy
rather than BFS — a monster that can be baffled by water and firelight is the
right monster.

**Attack** (`gru.attacked{agent, health}`): adjacent + visible + a
10-game-minute cooldown (`gruAttackCooldown = 600`). The payload carries the
**absolute post-wound health** (outcome convention), a `gruWound = 250` drop —
and since spec 044 (US3) the floor is conditional on the victim's raw
pre-attack health. At or above `nearDeathBelow` (200) the drop floors at
`gruWoundFloor = 1`: a healthy target is wounded, never executed. Below
`nearDeathBelow` the floor is 0: the attack can kill outright. The predicate is
deliberately the raw pre-attack health, never the hysteresis-latched
`Agent.NearDeath` bool (research R4) — a villager who recovered past the band
is not "already weakened" just because the latch hasn't cleared. A compile-time
invariant (`const _ uint = gruWound - nearDeathBelow` in `gru.go`) pins
`gruWound >= nearDeathBelow`, so an escalated hit against a weakened target
always lands at exactly 0 — never an ambiguous "survives at some positive
value". When the post-wound health is 0, `gruStep` emits the standard
`agent.died{cause: "gru"}` immediately after the attack, in the same batch, so
the entire existing death path — reducer `Dead` flag, inventory spill, the
`State.Deaths` ledger entry, the grave, chronicle, [[morgue]] — runs unchanged
(FR-015); because gru attacks land off the %60 needs heartbeat, the executor's
witness-death block never runs for them, so `gruStep` replicates that idiom
inline (a "Watched X die of the gru" witness memory loop over agents within
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
`situatedMemoryAboutEvent` and `Where` set to the remembering agent's own tile
at the moment of the memory (the sighter's or victim's tile for sighting/attack,
the witness's own tile for the witness memory) — not the gru's position. Since
spec 030 each also carries the required `origin` provenance stamp: the sighter's
and the victim's own memories (sighting, attack) stamp `OriginAction` — they are
the agent's own direct experience of the moment, not a report of someone else's
act — while the nearby witness memory stamps `OriginWitness`, matching every
other witness site in the sim. [[event-types]] catalogs the family.

## Connections

[[executor]] calls `gruStep` from `stepEvents` (same purity contract) and hosts
the run-end detection an escalated kill can trip (a gru kill that fells the
last living villager lands in the same batch as the `run.ended` declaration);
[[sim-state-reducer]] dispatches `gru.*` to `applyGru` and carries the death
fallout (ledger, grave); [[morgue]] mourns gru deaths under the charter
revision then in force; [[reflex-policy]] supplies the flee-to-warmth response;
[[social-fabric]] turns witness memories into rumors; [[tui-client]] renders it
as a red G; [[worldmap-generation]] bounds its spawn border.

## Operational notes

Live proving (seed 42, 1257 game days, run under the pre-044 unconditional
floor): the gru emerged ~60% of nights, attacked 186 times, never killed (zero
deaths in the whole run — every wound left its victim at 750; under today's
rules those healthy victims would fare identically, since escalation only
reaches targets already below 200 health), emitted zero events outside
22:00–06:00, and a witnessed attack
on Fern propagated as a rumor through the entire village with confidence decaying
80 → 35. Sightings are personal (subject −1) and thus omen material, not gossip;
only *witnessed attacks* seed rumors, which makes gru rumors appropriately rare.
