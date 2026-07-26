---
name: guardian-order-events
description: The four standing-order lifecycle events (order_placed/triggered/cancelled/expired) — reducer dispatch through applyGuardian, the door dry-run's rejection rules, and transitionGuardianOrder's cancel/expiry/trigger race resolution. Split from [[guardian-orders]]; load when tracing event-sourced order lifecycle mechanics.
kind: component
sources:
  - internal/sim/guardian.go
  - internal/sim/loop.go
  - internal/sim/state.go
  - internal/sim/executor.go
verified_against: 8495b34ffb9ee5dc02e224025f0a23313bbab900
---

# Guardian order events

A [[guardian-orders]] standing order's lifecycle is fully event-sourced: four
event types carry it, all cataloged in [[event-types]] and dispatched through
[[sim-state-reducer]]'s `applyGuardian` arm:

- **`metatron.order_placed`** (payload = the whole `GuardianOrder`) — landed through
  the [[sim-loop]] `InjectSocial` door by `placeOrder` (a `monitor_and_act` call) or by
  `deferOmen` (a system deferral). The reducer dry-run is the door authority: it rejects
  a duplicate id in ANY status, an unknown `Origin`, empty `EventTypes` (an uncompilable
  condition), a TTL outside 1..7 days, an out-of-range `Agent`, an over-long
  condition/action, and a player placement beyond the cap. The payload's `Status` is
  IGNORED — an order always lands `active`, then the retention prune runs.
  Since spec 054, `PlacedSeq` is likewise IGNORED on the payload — the
  reducer stamps it from the event's own store seq (`e.Seq`, identical live
  and in replay: `Loop.stampSeqs` pre-assigns the same `last+i+1`
  `AppendEvents` records, [[sim-loop]]), so [[scenario-machinery]]'s
  `OrderPlacedEvidence` can trust it as a re-locatable identity.
- **`metatron.order_triggered`** (`OrderTriggeredPayload{id, matched_type, matched_tick}`)
  — injected by the trigger worker when a match fires; NEVER emitted during replay. See
  [[guardian-order-triggering]] for the matching and execution mechanics that land this
  event.
- **`metatron.order_cancelled`** (`OrderIDPayload{id}`) — injected by a `cancel_order`
  call.
- **`metatron.order_expired`** (`OrderIDPayload{id}`) — **executor-emitted**, a pure
  function of `(state, tick)` exactly like `metatron.charge_regenerated`: the [[executor]]
  emits it once when an active order's `ExpiresTick` elapses, so it is reproduced
  deterministically in replay without the guardian running (unlike a trigger). It is NOT on
  the `injectSocialWhitelist` — `order_placed`/`order_cancelled`/`order_triggered` are
  the injected three; expiry is produced sim-side.

`transitionGuardianOrder` performs every active→terminal move and rejects an unknown or
already-non-active id — this is where the **cancel/expiry/trigger race resolves**:
exactly one terminal lands, and the loser hits a non-active order and refuses at the
door. Replay reconstructs order state through `json.Unmarshal` + `Apply` alone;
`matchOrders` runs only in the absorb goroutine, so a predicate can never match during
reconstruction (the edge-case guarantee — triggering is a live-observation behavior).
`GuardianOrder.ExpiresTick` is a SHIFT field in the miracle `rebaseTicks` taxonomy
(shifted only while active; `PlacedTick` is KEEP) — see [[guardian-miracles]].

## Connections

[[guardian-orders]] is the parent this note splits from — the entity model,
caps/TTL/prune mechanics, and the player-facing tool surface live there.
[[event-types]] catalogs the four order events. [[sim-state-reducer]] holds
`applyGuardian`, `transitionGuardianOrder`, and `pruneGuardianOrders`, plus
the `State.GuardianOrders` field. [[sim-loop]] owns the `InjectSocial` door
and the `stampSeqs` precedent `PlacedSeq` relies on. [[executor]] emits
`metatron.order_expired`. [[scenario-machinery]] is the spec-054 consumer of
`PlacedSeq` — its `OrderPlacedEvidence` constructor re-locates a watch's
placement event for the first-night exercise's rubric evidence.
[[guardian-miracles]] shares the `rebaseTicks` taxonomy that shifts an active
order's expiry across a time snap. Spec: `specs/029-metatron-agency/`
(TASK-27) — `data-model.md`, `contracts/events.md`.
