---
name: sim-state-reducer
description: sim.State and Apply — the single event-driven mutation path used identically live and in replay; canonical JSON for hashing. Field catalog and per-family Apply-arm detail split across six children; load this note for the whole-state model, Tick handling, and Marshal/Hash — the children for arm-level detail.
kind: component
sources:
  - internal/sim/state.go
  - internal/sim/agents.go
  - internal/sim/terrain.go
  - internal/sim/journal.go
  - internal/sim/morgue.go
  - internal/sim/miracles.go
verified_against: 63390f122bdf4e1b7abf518a8be83de725f06230
---

# Sim state & reducer

`sim.State` is the whole world in one struct — clock state, the living
world's agents (needs/intents/inventories/memories), structures/piles/
terrain overlays, the social fabric, the Guardian's charge bank and
standing orders, village governance, and the run's outcome ledger — and its
`Apply(event)` method is the **only** event-driven mutation path: the live
loop and crash recovery run the exact same code, which is what makes replay
provably equal to live execution. Spec 012 bumped the save format to v2,
and spec 013 (inventory & storage) bumped it again to **v3**
([[world-save-directory]]); a v1 world's `Inventory` (just `wood`/`food`)
cannot decode under this build at all — [[world-migration]] is the bridge,
chaining 1→2→3 in one run and landing as a single wholesale-replace event
rather than incremental `Apply` calls.

`NewState(seed, m)` is genesis; `Apply` switches on event type across
roughly eighty payload shapes ([[event-types]]). Both the full per-agent/
per-world field catalog and the arm-by-arm `Apply` detail are split into
six children (summary-style, corpus-spec v2) — read the one your task
needs, not the whole set:

- [[sim-state-agent-fields]] — clock state and the per-agent field catalog:
  needs/intents/inventories/memories, the `Journal`/mental-map pointers, the
  `IntentLog` ring, `NeedsAnchor`, and `LastMindIntentDone`.
- [[sim-state-world-fields]] — the shared-world field catalog: structures/
  piles, the social fabric, Guardian charges/standing orders, governance,
  the morgue/run-outcome ledger, the curriculum ladder, world-tuning, and
  the Guardian report card.
- [[sim-state-apply-agents]] — genesis placement plus the core per-agent
  `Apply` arms that fire every tick: clock/night/forage, intent/movement/
  eating/talk/needs/death, the v2 crafting/gather yields, the v3 storage
  events, and the wall demolish/repair family.
- [[sim-state-intent-lifecycle]] — the intent-ring closure arms
  (`agent.build_failed`, the spec-043/062/064 completion/rejection/stall
  stamps), the hail family, and `agent.died`'s spill/`Deaths`/grave/
  `run.ended` effects.
- [[sim-state-apply-world]] — the unexported map/scenario fields and the
  world/governance dispatch arms: mental-map knowledge growth, the gru/
  governance/miracle/guardian-order dispatchers, curriculum unlocks,
  `sim.tuning_applied`, and `world.migrated`'s wholesale replace.
- [[sim-state-cognition-arms]] — the cognition/telemetry arms: memory
  growth (`memory_added`/`memory_embedded`/`situation_embedded`), the
  journal family, the plan family, and the `cog.*` no-op telemetry types.

## How it works

**Tick is deliberately not event-sourced**: quiet ticks (no events) advance
the clock without a log row. The live loop mutates `state.Tick` directly;
recovery sets it to `max(snapshot tick, last event tick)` and re-lives any
quiet tail deterministically.

Canonical bytes: `Marshal()` uses `encoding/json` over structs only (fixed
field order — payload shapes like `AgentMovedPayload` are structs, never
maps), so equal states produce identical bytes. `Hash()` is SHA-256 of
that, used by [[snapshots]] verification and the determinism tests.
Wall-clock time never appears in state.

Unknown event types — including `daemon.*` and `world.created` — are
recorded history but state no-ops, so new event types never break old
replay. `world.forked` (spec 076, [[world-forking]]) carries an EXPLICIT
no-op arm beside `world.created` (self-documentation; the type is never
"unknown" to future totality checks): the no-op is load-bearing — it keeps
a fork's state at the fork tick byte-identical to its parent's, since
lineage lives in the event and the manifest mirror, never on `State`.

## Connections

[[sim-loop]] generates events via the [[executor]] and applies them here;
[[daemon-lifecycle]] replays the [[event-log]] through `Apply` at startup;
[[event-types]] lists every payload struct; [[world-migration]] is the
sole producer of `world.migrated`. [[guardian-miracles]] covers the miracle
payload shapes, cost table, and the `rebaseTicks` shift-semantics taxonomy
`applyTimeSnapped` uses (spec 029's standing-order `ExpiresTick` shift,
spec 043's `NeedsAnchorTick` shift, spec 061's `PairTalks[].Tick` shift).
The other sibling notes this reducer touches — [[guardian-orders]],
[[mental-maps]], [[memory-retrieval]], [[decision-context]],
[[reflex-policy]], [[curriculum-ladder]], [[world-tuning]],
[[scenario-machinery]], [[grounded-feedback]], [[morgue]] — are linked from
whichever of the six split-off notes above owns the relevant field or arm;
follow those, not this note, for arm-level detail.

## Operational notes

`EffectiveRate`/`Degraded` are part of state (hence snapshots) but only
change via explicitly emitted transition events, so unloaded same-machine
runs stay byte-deterministic. Adding a state field means adding events that
set it — direct mutation outside `Apply` (except `Tick`) breaks the replay
contract.

## Spec 086 — refs fold by ID; names never enter state

Every agent-referencing payload field is a `sim.AgentRef` (`{id,name}` on
the wire, dual-shape unmarshal accepting legacy bare ints forever). Two
laws hold in this reducer, permanently: (1) **no `AgentRef` is reachable
from `sim.State`** — arms fold `ref.ID` (and `refIDs(...)` for lists) into
the unchanged int-typed state entities, so `Marshal()`/`Hash()` bytes for
any pre-086 history are byte-identical (`TestNoAgentRefInState`,
`TestPre086ReplayByteIdentity`); (2) **no arm ever validates a name** —
injected pre-086 rows replay through these same arms on every recovery, so
name validation lives exclusively at the live-emission choke points
(`mustPayload`, the `InjectSocial` door), which replay never traverses.
The four state-shared payload types split into wire mirrors
(`DirectiveIssuedPayload`, `OrderPlacedPayload`, `ProphecyDeclaredPayload`,
`DeathRef`) whose arms fold `.ID`s into the unchanged entities.

## Spec 092 — reducer-constants replay-hazard doctrine (TASK-75)

**Default**: the payload carries the outcome; the reducer copies it verbatim.
Spec 019's `agent.memory_added` is the precedent — `Where`/`Why`/`Conv`/`Origin`
are "baked at emission, never re-derived at render or replay, so live and
replay agree" ([[sim-state-cognition-arms]]). This is the safe shape: a later
balance change can never alter what an OLD log replays to, because the number
that mattered was already written into the event at the time it happened.

**Exception**: several `Apply` arms instead re-derive an outcome from a bare
package constant — reading it fresh at replay time rather than trusting a
payload field. That is harmless as long as the constant never changes; it
becomes a **replay hazard** the moment someone retunes it, because an OLD log
then replays through the CURRENT build's constant, not the value that was live
when the log was recorded. Reducer-re-derives is the exception, and it
requires an explicit format-version bump + migration before the constant may
change — **spec 094 / TASK-134** is building that machinery (an event-log-level
format stamp plus a translating migration mode) precisely because today's
`world.json` `FormatVersion` only catches whole-state-SHAPE breaks, not a
same-shape semantic retune like these. This note records the doctrine and the
audit surface the migration work consumes; it does not implement the bump.

**Partial existing mitigation — spec 048's genesis-tuning-pin**: five of the
constants below were promoted to per-world tuning dials (`RefuelDyingBelow`,
`FireBurnPerWood`, `GruEmergePerMille`, `PlannerCadenceTicks`,
`EncounterCooldownTicks`, [[world-tuning]]), and since spec 057 every FRESH
world pins its effective set into its own log at genesis
([[world-tuning-boot-seeding]]) — a post-057 world's replay is immune to a
later default change for exactly those five. The residual scope is already
documented there: pre-057 worlds and any world produced by `promptworld
migrate` follow compiled defaults, not a pin — the same hazard this doctrine
describes, for those five dials, on those worlds. Every OTHER constant audited
below has no promotion and no pin at all, on any world, regardless of age.

### Audit (FR-004) — reducer arms that re-derive from a mutable constant

Swept from `internal/sim/state.go`'s `Apply` (this note's pin); each site
below also carries a short "Replay hazard (spec 092/TASK-75)" code comment.
TASK-134/spec-094 consumes this table as its migration surface.

| `Apply` arm | Site | Re-derived from | Value(s) |
|---|---|---|---|
| `agent.foraged` | state.go:1130 | `forageYieldV2` | 2 |
| `agent.chopped` | state.go:1153-1157 | `chopYieldBare`/`chopYieldAxe` | 1 / 3 |
| `agent.hunted` | state.go:1195-1199 | `huntYieldBare`/`huntYieldSpear` | 8 / 12 |
| `agent.hunted` | state.go:1211 | `denCooldownSec` | 21600 (6h) |
| `agent.built` | state.go:1229 | `recipeFor("build_"+kind)` input costs | recipes.go table |
| `agent.built` (fire) | state.go:1242 | `FireBurnPerWood()` | tuning-covered — see mitigation above |
| `agent.built` (wall) | state.go:1258 | `wallMaxHP(kind)` → `wallPlankHP`/`wallStoneHP` | 200 / 600 |
| `agent.quarried` | state.go:1294-1298 | `quarryYieldBare`/`quarryYieldAxe` | 1 / 3 |
| `agent.collected_water` | state.go:1321 | `collectWaterYield` | 1 |
| `agent.crafted` | state.go:1350-1356 | `recipeFor(goal)` input/output table, `spearDurability`, `axeDurability` | recipes.go table; 3; 10 |
| `agent.wall_chipped` | state.go:1484 | `demolishChipHP` | 100 |
| `agent.wall_repaired` | state.go:1529-1534 | `wallRepairMaterial(kind)`, `wallMaxHP(kind)`, `repairHPPerUnit` | — ; — ; 100 |
| `agent.talked` | state.go:2135-2136 | `talkMoraleBonus` | 50 |

Lower-tier — listed for FR-004 completeness ("every site") but these
re-derive a CLASSIFICATION of an already-carried absolute value, not a
produced resource amount; a retune reclassifies an old log's derived
flags/anchors, never the recorded needs numbers themselves, so the blast
radius is narrower:

| `Apply` arm | Site | Re-derived from |
|---|---|---|
| `agent.needs_changed` | state.go:1872-1875 | `nearDeathBelow`/`nearDeathResetAt` (NearDeath latch) |
| `agent.needs_changed` | state.go:1887 | `trajectoryWindowTicks` (anchor-roll cadence) |
| `agent.needs_changed` | state.go:1898-1901 | `recoveryDangerBand`/neglect band constants |
| `agent.memory_added` | [[sim-state-cognition-arms]] | `GenerationBumpSalience` (9, the generation-bump threshold) |

Not audited here: `clock.degraded`'s `EffectiveRate` is payload-carried (the
loop measures it once and the reducer copies it verbatim) — it is the
DETERMINISM-SCOPE hazard ([[deterministic-rng]], [[sim-loop]]: replay is exact
per-log, but two machines on the same seed can measure different wall-clock
rates), a different concern from this doctrine's re-derive-from-a-constant
hazard.

### Reconciling with the spec-019 precedent

Spec 019 chose emitter-computes for `Memory`'s situated fields specifically
because a memory's phrasing is expensive to re-derive and cheap to carry
once ([[sim-state-cognition-arms]]); this doctrine generalizes that choice
into the DEFAULT posture for every new `Apply` arm, and treats the audited
table above as the accumulated exception set — grandfathered, not a pattern
to keep extending.
