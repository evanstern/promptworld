---
name: event-types-guardian-orders
description: Guardian standing-order event rows split from [[event-types]]: metatron.charge_regenerated/nudged/order_placed/order_triggered/order_cancelled/order_expired. Load when tracing charge regen, nudge validation, or the spec 029/059 standing-order and survival-watch lifecycle.
kind: concept
sources:
  - internal/sim/guardian.go
  - internal/daemon/daemon.go
verified_against: 4c66d240b2715706964f02cfd2396256c9957d8e
---

# Event types — guardian standing-order events

Back to [[event-types]] for the payload-grammar conventions and the full
event-domain index.

Spec 029 (guardian agency — [[guardian]], [[guardian-orders]]) likewise adds
**no** format bump: `State` gains `GuardianOrders []GuardianOrder`
(`omitempty` — an empty order set is genuinely zero-value, unlike
`GuardianCharges`'s spent-to-zero precedent, so a pre-029 snapshot with the
field absent unmarshals to nil), and FOUR new event types drive a standing
order's lifecycle — `metatron.order_placed` (monitor_and_act), one-shot
`metatron.order_triggered` (the trigger worker, live-only, never replayed),
`metatron.order_cancelled` (cancel_order), and executor-emitted
`metatron.order_expired` (the `charge_regenerated` pattern: a pure function
of state + tick, so it reproduces on replay with no angel running). The same
spec retires `nudge_dream`/`nudge_omen` from the tool registry in favor of
`send_vision`, so `metatron.nudged`'s `form` domain is now `vision` (exactly
one living target, any hour) / `omen` (≥1 living targets, night-only,
`State.Night`) / `dream` (legacy, grandfathered: accepted on replay for
historical events, but no live tool can produce a new one).

Spec 059 (metatron survival autonomy — [[guardian-orders]], [[guardian]]) is
also format-stable: `GuardianOrder` gains `omitempty` `Survival` (`""` = an
ordinary structural order, the pre-059 shape; `near_death`/`starvation`/
`exposure` names one of the three canonical system-origin survival watches),
so a pre-059 order round-trips byte-identically. No new event types: the
three watches ride the EXISTING `metatron.order_placed`/`order_cancelled`/
`order_expired` lifecycle, but origin-keyed — a survival watch is exempt from
the player-order cap and the TTL bounds at the `order_placed` door, refused
outright at the `order_cancelled` door (the angel's own nature, not a player
configuration), and skipped entirely by the executor's `order_expired` sweep
(non-expiring by origin, never a giant TTL). The daemon's
`seedSurvivalWatches` ([[daemon-lifecycle]]) lands the three watches once via
ordinary `order_placed` events, replay-safe like `seedMeetingConvention`/
`seedTuning`.

| Type | Payload struct | Emitted by | Reducer effect |
|---|---|---|---|
| `metatron.charge_regenerated` | `ChargeRegeneratedPayload{}` in `internal/sim/guardian.go` | executor, absolute 6-game-hour boundaries below cap | `GuardianCharges` +1, cap 3 ([[guardian]]) |
| `metatron.nudged` | `GuardianNudgedPayload{form, targets, text}` | Guardian console turn (injected, TASK-12) | validates (charges > 0, form ∈ vision\|omen\|dream, living targets, text cap) then `GuardianCharges` −1; `vision` (spec 029, replaces `dream` as the live one-target form) needs exactly one living target at any hour; `omen` needs ≥1 living targets AND `State.Night`; `dream` is legacy-only (grandfathered exactly-one-target validation so historical events replay, but no tool can emit a new one); villager memories ride companion `agent.memory_added` events in the same atomic batch |
| `metatron.order_placed` (spec 029, [[guardian-orders]]) | `GuardianOrder{id, origin, condition, action, event_types, agent, keywords?, confirm?, placed_tick, expires_tick, status, survival?, placed_seq?}` | the guardian's `monitor_and_act` tool (injected via `InjectSocial`), or — since spec 059 — the daemon's boot-time `seedSurvivalWatches` for the three canonical system survival watches | validates (non-empty id not reused by any past order regardless of status, `origin` ∈ player\|system, non-empty `event_types`, `agent` index valid or −1 for any, `condition` ≤300 chars, `action` ≤400 chars, and — player-origin only — fewer than 3 already-active player orders, `GuardianPlayerOrderCap`; system-origin deferral orders are exempt from the cap); a non-empty `survival` (spec 059: `near_death`\|`starvation`\|`exposure`) MUST be `origin: system` and is EXEMPT from the ttl 1..7 game days bound (non-expiring by nature, `expires_tick` ignored); the payload's `status` is ignored — a landed order is always `active`; since spec 054 the payload's `placed_seq` is likewise IGNORED — the reducer stamps `PlacedSeq` from the event's own store seq (`e.Seq`) instead, so it agrees live and in replay ([[scenario-machinery]]'s `OrderPlacedEvidence` re-locates a watch's placement this way); `GuardianOrders` appended then pruned to every active order plus the most recent 32 non-active (`pruneGuardianOrders`) |
| `metatron.order_triggered` | `OrderTriggeredPayload{id, matched_type, matched_tick}` | the angel's trigger worker (injected, live-only — NEVER emitted during replay, since the matching runs off live events the replica sees post-batch) | the named order transitions active → triggered (one-shot consumption); rejects an unknown id or one not currently active; a survival watch (spec 059) never lands this — it is non-consuming, so its own trigger runs no `order_triggered` at all ([[guardian-orders]]) |
| `metatron.order_cancelled` | `OrderIDPayload{id}` | the guardian's `cancel_order` tool, injected | the named order transitions active → cancelled; same rejection rule as triggered; since spec 059 a survival watch is refused outright regardless of status — it is the guardian's own nature, not a player configuration, and the player order surface cannot release it |
| `metatron.order_expired` | `OrderIDPayload{id}` | executor, `stepEvents`, once per order once `nextTick >= expires_tick` for an active order (the `charge_regenerated` pattern — a pure function of state + tick, so replay reproduces it without any angel running) | the named order transitions active → expired, freeing its slot against the player cap; since spec 059 a survival watch is skipped by this sweep entirely (origin-keyed TTL exemption, never evaluated) |

## Connections

The standing-order
lifecycle (spec 029) reduces in `internal/sim/guardian.go` alongside
`charge_regenerated`/`nudged` — see [[guardian-orders]] for the placement
validation, trigger-matching, and confirm/degradation mechanics; `order_placed`/
`order_triggered`/`order_cancelled` are whitelisted in [[sim-loop]]'s
`InjectSocial` door exactly like the miracle types, while `order_expired`
needs no whitelist entry (executor-emitted, the `charge_regenerated`
precedent). Since spec 059, [[daemon-lifecycle]]'s `seedSurvivalWatches` is a
second `order_placed` emitter (boot-time, alongside `monitor_and_act`), and
the `Survival` discriminator's cap/TTL/cancel exemptions reduce alongside the
rest of `applyGuardian`.

[[scenario-machinery]] also owns
`GuardianOrder.PlacedSeq` (spec 054) as an `OrderPlacedEvidence` consumer,
reduced alongside the rest of `applyGuardian`.
