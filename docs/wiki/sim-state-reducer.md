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

## Replay doctrine: names and re-derivations are versioned (spec 094)

`Apply`'s switch keys on persisted event-type NAMES, and several arms
RE-DERIVE state from doctrine constants at apply time (the hunt-yield
derivation from `huntYieldSpear` in `agent.hunted`'s arm is the canonical
example — spec 092's determinism/emitter-computes doctrine, TASK-75, audits
this class). Both are replay contracts: renaming a type or retuning a
re-derived constant changes what replaying an EXISTING log computes.
Either change REQUIRES bumping `store.LogFormatVersion` ([[event-log]]) and
shipping a migration — a pure rename TRANSLATES the log (the spec 094
`metatron.*`→`guardian.*` rename, `sim.LogFormatV1Renames`, is the first);
a semantic break SNAPSHOT-CUTS ([[world-migration]]'s decision rule).
`TestGatherTuningMirror` pins the tuning values against the ratified log
format so a retune must confront the doctrine, and the load gates
([[event-log]], [[world-save-directory]]) make an unmigrated old log
unreachable by these arms. Persisted reference strings inside preserved
payloads (`EvidenceRef.Type`) keep old names; readers normalize via
`sim.CanonicalEventType` — the log's own type column is never aliased at
read.

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
