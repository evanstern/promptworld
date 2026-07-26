---
name: guardian-orders
description: The event-sourced standing-orders subsystem (spec 029/052/059) — GuardianOrder entity model, caps/TTL/prune, and the player-facing monitor_and_act/cancel_order surface. Event dispatch splits to [[guardian-order-events]], matching/triggering/deferral to [[guardian-order-triggering]], and the boot-seeded survival watches to [[guardian-survival-watches]]. Load for the entity model, tool surface, and cross-subsystem connections.
kind: component
sources:
  - internal/guardian/orders.go
  - internal/sim/guardian.go
  - internal/sim/executor.go
  - internal/sim/loop.go
  - internal/sim/state.go
  - internal/guardian/turn.go
  - internal/guardian/toolcalls.go
  - internal/tool/registry.go
verified_against: cffd9a79bbed61ccac573d97c6cf544565b40336
---

# Guardian's standing orders

A standing order is a pre-authorized watch-and-act instruction (spec 029, TASK-27):
the player tells the guardian "when Rowan next falls asleep, send her a comforting
vision" and walks away. The condition is compiled ONCE at placement into structural
predicates evaluated for free as world events stream past; when it matches, the guardian
wakes and performs the pre-authorized action through exactly the [[guardian]] console
turn's guarded machinery. Orders are **one-shot** (fire once, consumed), event-sourced
(they ride `sim.State` through snapshots and replay), and never fire during
reconstruction — replay rebuilds their state but only live observation triggers them.

## The entity and its lifecycle

`sim.GuardianOrder` (`internal/sim/guardian.go`, data-model §1) is the event-sourced
record: `ID` (`"ord-<placedTick>-<seq>"`, deterministic, no RNG — `nextOrderID` in
`orders.go`), `Origin` (`"player"` | `"system"`), `Condition` (the original NL, ≤300
runes), `Action` (the NL action instruction, ≤400 runes), `EventTypes` (the structural
predicate — non-empty, drawn from the observable vocabulary), `Agent` (a villager index,
or `-1` for any), `Keywords` (a lowercase coarse text filter, ≤6), `Confirm` (fuzzy —
needs the watch confirm), `PlacedTick`, `ExpiresTick`, `Status`
(`active` → `triggered` | `cancelled` | `expired`, one-way), and — since spec
054 — `PlacedSeq` (`omitempty`): the placement event's store seq, stamped by
the reducer at apply time from the event envelope (the `Memory.Seq`
precedent), never trusted from the payload (like `Status`) — it lets
[[scenario-machinery]]'s `OrderPlacedEvidence` re-locate a recorded
`metatron.order_placed` without a log scan.

**Caps and bounds** (`sim` constants): at most `GuardianPlayerOrderCap` (3) ACTIVE
**player-origin** orders may stand concurrently — system-origin deferral orders are
exempt (they are bookkeeping for an already-authorized act, FR-012). Every order carries
a TTL in game days, player-specifiable, default 3, bounded `GuardianOrderTTLMinDays`..
`GuardianOrderTTLMaxDays` (1..7); the reducer validates `ExpiresTick - PlacedTick`
against the same `ticksPerGameDay` (`24*3600`) literal the turn side computes from
(mirrored in `orders.go` so the door and the placer can never diverge). The
`GuardianOrders` slice is pruned to retain every active order plus the most recent
`guardianOrderRetain` (32) non-active ones (`pruneGuardianOrders` — deterministic,
order-preserving, so replay prunes identically), giving the status/trail recent history
without unbounded growth.

## Event sourcing

Four event types carry the lifecycle — `metatron.order_placed`, `order_triggered`,
`order_cancelled`, `order_expired` — dispatched through [[sim-state-reducer]]'s
`applyGuardian` arm; `transitionGuardianOrder` resolves the cancel/expiry/trigger
race so exactly one terminal ever lands. See [[guardian-order-events]] for the
full event-by-event breakdown, the placement door's rejection rules, and the
spec-054 `PlacedSeq` stamping [[scenario-machinery]] relies on.

## Live matching, triggering, and deferral

`matchOrders` runs live-only in [[guardian]]'s absorb loop and enqueues at most
one job per order per batch; `triggerWorker` then runs a structural order
straight through `runTrigger`, or, for a fuzzy order, first confirms via one
cheap [[llm-orchestrator]] `KindGuardianWatch` call. A daytime `send_omen`
never lands directly — `landOmen`'s day path defers itself as a system-origin
order whose one-shot trigger re-runs the omen the instant night falls,
charge-free at placement. See [[guardian-order-triggering]] for the full
matching, confirm, execution-step, and deferral mechanics.

## Survival watches (spec 059)

Three system-origin survival watches — near-death, starvation, exposure —
exist in every world from boot, seeded once via `seedSurvivalWatches`
([[daemon-lifecycle]]). They are cap/TTL-exempt and cancel-refused, matched
live through per-villager hysteresis latches (`matchSurvival`), and carry the
ONE frame carve-out letting the guardian act on its own initiative — a vision
or miracle — to save a life. See [[guardian-survival-watches]] for the full
origin-keyed exemptions, danger bands, and survival-turn mechanics.

## Surfaces

The player reads and cancels orders through the guardian. `monitor_and_act` and
`cancel_order` are registered [[tool-registry]] tools (`monitor_and_act` uses a hand-built
`monitorAndActSchema` — arrays are unrepresentable in the scalar Param model, like
`set_plan` — with `event_types` an enum over `observableEventTypes`, the curated
vocabulary of genuinely-emitted types); `toolcalls.go`'s `handleMonitor`/
`handleCancelOrder` wrap `placeOrder`/`cancelOrder`, mapping a door rejection to
in-fiction counsel as a `rejected_gate` the model may repair. On a curriculum-ladder world
(spec 046, [[curriculum-ladder]]) both tools sit INSIDE the stage-1 capability
ceiling — a ratified amendment added `monitor_and_act`/`cancel_order` to the
stage-1 grant since the first-night exercise teaches the watch as a stage-1
primitive — so standing orders are available at every stage, like omens and
visions ([[guardian]]'s `applyStageCeiling`). The turn prompt carries active orders
(`writeStandingOrders` — id, condition, days-left, fuzzy/structural — FR-017) so the
angel's counsel stays truthful to live state, and the model-free `metatron.Status`
surface lists them (`Status.Orders`, `OrderStatus{id, condition, origin, fuzzy,
expires_day, status}`, FR-016). The fixed frame's `guardianInitiativeFrame` (a
compile-time constant appended last, beneath any charter/soul/skill text — spec
036's persona SOUL fragments stack with the editable text, never above the
frame) binds standing-order and
meta-tool use to player-requested or pre-authorized action only — never the guardian's own
initiative — with the door-side grant gate backing it independently. Since
spec 059 this has exactly ONE carve-out, keyed on turn origin rather than
tool: a **survival-watch** turn's frame (above) permits a vision or miracle
on the guardian's own initiative to save a life; standing orders and clock
control are NOT part of that carve-out and remain player-authority in every
turn, survival included (FR-004).

## Connections

[[guardian-order-events]] holds the four event types and reducer dispatch;
[[guardian-order-triggering]] holds the live-matching/confirm/execution and
daytime-omen-deferral mechanics; [[guardian-survival-watches]] holds the
spec-059 boot-seeded watches. [[guardian]] hosts the trigger worker, the turn
body (`runTurn`) both console and system paths share, the meta-tool/deferral
wiring, and `persona/charter.go`'s `DefaultCharter` (states the survival duty
in-fiction since spec 059). [[tool-registry]] declares `monitor_and_act`/
`cancel_order` and the observable-event vocabulary; [[guardian-miracles]]
shares the `rebaseTicks` taxonomy that shifts an active order's expiry across
a time snap. Spec: `specs/029-metatron-agency/` (TASK-27) — `spec.md`
US2/US3/US4/US6, `data-model.md`, `contracts/events.md`, `contracts/routing.md`;
`specs/059-metatron-survival-autonomy/` — `spec.md`, `plan.md`.

## Operational notes

Orders are one-shot by doctrine — recurring watches are out of scope; the player
may re-place. A triggered turn's own emissions must not re-trigger the order that
produced them (bounded cascade). A confirmed trigger that races its own TTL expiry
resolves at the door: exactly one of triggered/expired lands, never both.
A full game-day of unattended watching with ≤3 active orders adds at most the placement
call plus rate-capped confirms (≤48 cheap calls/order/day worst case) — no unbounded
per-event model cost shape exists (SC-008). On an ENDED world (spec 044,
[[morgue]]) the whole subsystem goes quiet by construction: `stepEvents` emits
nothing, so no order can match or expire, and the `InjectSocial` door narrows
to recorded prose, refusing the three injected order events.
