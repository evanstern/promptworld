---
name: guardian-order-triggering
description: How a standing order fires — live-only structural/fuzzy matching in the absorb path (matchOrders/orderMatches), the confirm-then-execute trigger worker (runConfirm/runTrigger), and the daytime-omen deferral that re-runs send_omen the instant night falls. Split from [[guardian-orders]]; load when tracing match-to-act mechanics.
kind: component
sources:
  - internal/sim/guardian.go
  - internal/guardian/orders.go
  - internal/guardian/turn.go
  - internal/guardian/digest.go
  - internal/llm/llm.go
  - internal/llm/config.go
verified_against: 6a5344a12cdc8858909ca7cf209d55025135e9d5
---

# Guardian order triggering

A [[guardian-orders]] standing order's condition is compiled once at
placement into structural predicates; this note covers what happens from the
moment one of those predicates matches a live event through to the
pre-authorized action landing — matching, the fuzzy confirm step, execution,
and the one case (a daytime omen) where landing means deferring instead.

## Live matching (the absorb path)

`matchOrders` runs in [[guardian]]'s `run()` loop AFTER the replica applies each
batch and the mirror refreshes, so it is live-only by construction. `orderMatches`
is a PURE predicate — no state, no model call, evaluated free per event
(SC-001): the event type is one of the order's `EventTypes`; if the order
pins an agent (`>= 0`) the event concerns that villager (`eventConcernsAgent`
probes the `agent`/`from`/`to` payload fields — a best-effort structural
match, never a false positive); if the order lists keywords the lowercased
payload contains at least one. Only active orders match.

Orders fire in **order-id order** within a batch, at most once: `pendingTrigger`
(stateMu-guarded) dedups an order already queued but not yet resolved, and one job is
enqueued per order per batch. A structural hit enqueues a `triggerJob` onto the buffered
`triggerQ` (a full queue logs and drops the order). A **fuzzy** order (Confirm) matches
structurally here too, but its hit is routed as a CONFIRM job and is rate-capped to one
confirm per `confirmRateTicks` (1800 ticks = 30 game minutes) per order via the
absorb-owned `lastConfirmTick` map — NOT event-sourced (a skipped confirm is an economy
decision, never world history); a rate-capped hit is logged and skipped so a storm of
matching events never triggers a flood of watch calls (FR-009, SC-008).

## Trigger execution

`triggerWorker` consumes `triggerQ` FIFO (one worker, so triggered turns and confirms
serialize with each other and — via the shared `turnBusy` — with console turns). A
structural job fires straight through `runTrigger`; a fuzzy job first runs `runConfirm`.

**The fuzzy confirm** (`confirmOrder`, spec 029 US6, `contracts/routing.md`): ONE bare
`Submit` on [[llm-orchestrator]]'s new `llm.KindGuardianWatch` kind (routed to a cheap
default chain — `local`→`cloud` in `internal/llm/config.go`), `MaxTokens` 16, a fixed
yes/no system prompt (`confirmSystem`), and a user prompt rendering the order's condition
plus the matched event in the digest vocabulary (`describeEvent`, reading static
`sim.AgentNames`). Reply contract (`confirmYes`): the first token, lowercased and
stripped of punctuation, must be exactly `"yes"` — anything else, empty, garbage, or an
error is a NO. A no/failed verdict leaves the order armed with NO retry (a single call,
not a loop); only the in-flight marker is cleared so a later hit can confirm again,
subject to the rate cap.

**`runTrigger`** fires one matched order:

1. Land `metatron.order_triggered` through the door — the dry-run enforces the order is
   STILL active, so a cancel/expiry that raced the match wins here and the trigger is
   abandoned silently.
2. **Empty-bank precheck** (`knownActEmptyBank`): a system-origin (deferral) order's
   action is a known charge spend, so an empty bank short-circuits to an honest moment —
   no model call, no cloud cost. A free-form player order's action may be advisory or a
   meta act, so it still runs the turn.
3. Acquire `turnBusy` with a **bounded wait** (`acquireTurnBusy`, `systemTurnBusyWait`
   90s): system turns WAIT for the single-flight slot (unlike the console's fail-fast
   `ErrTurnBusy`), but a wedged console turn degrades the trigger to an honest moment
   rather than hanging.
4. Run the pre-authorized action as a **system-authored turn**: `runTurn` with
   `turnOrigin{system: true, jobPrefix: "watch", seed: order.Action}` — the SAME
   [[tool-loop]]-driven turn body a console message uses (same roster/handler/gate
   composition, same `cog.tool_call` telemetry, same retry marker). The framing differs:
   the transcript opens with a `[watch]` origin marker over the order's action (never a
   player-text line — a triggered turn has no player text), the correlation id is
   `watch-metatron-<tick>`, and moment consumption is suppressed (the player-facing queue
   stays intact for the next console open; the trigger worker queues the turn's OWN
   moment).
5. `queueMoment` from the outcome: `triggeredMoment` names the landed act on success
   (omen/vision/miracle — since spec 052 the moment text renders the display noun
   through the active [[skin]], `mt.sk().FormNoun(form)`/`.WorkingNoun()`, never the
   frozen tool id or recorded form value), `degradedMoment` maps a failed turn to ONE model-free honest
   moment per failure family — `ErrBudgetExhausted`/`ErrTierDown`/`ErrTierBusy` →
   "my sight dimmed", otherwise "I faltered" — never a retry (FR-011). Moments accrete to
   `metatron/soul.md` and the queue so the next console reply leads with what the guardian
   did while the player was away (SC-003).

## Daytime-omen deferral

A daytime `send_omen` never lands and never refuses (FR-012): `landOmen`'s day path calls
`deferOmen` ([[guardian]]'s `turn.go`), which places a **system-origin** standing order —
`EventTypes` `["sim.night_started"]`, TTL 1 game day, cap-exempt — whose one-shot trigger
re-runs the omen the instant night falls. Placement is FREE; the charge is spent at
trigger-time landing, not at placement (SC-004). The action seed leads the night system
turn back to `send_omen` with the promised targets and text (terse framing keeps typical
renderings within the 400-rune action cap; a near-cap omen can exceed it and is refused
at placement with counsel to shorten). `"everyone"` is preserved as the target word so
the night turn re-resolves against whoever lives THEN; a named list re-sends to those
still living. The `monitor_and_act` grant is NOT required — a deferral carries `send_omen`'s
gate, so a world granting `send_omen` but withholding `monitor_and_act` can still defer.
Cancelling the deferral before nightfall wins: the omen never lands and no charge is spent.

## Connections

[[guardian-orders]] is the parent this note splits from — the entity model,
event sourcing (see [[guardian-order-events]]), caps/TTL, and the player-facing
tool surface live there. [[guardian]] hosts `matchOrders`/`triggerWorker` and
the turn body (`runTurn`) both console and system paths share. [[llm-orchestrator]]
routes the fuzzy `KindGuardianWatch` confirm to its cheap chain. [[tool-loop]]
drives the system-authored turn exactly as it drives the console turn. [[skin]]
renders a triggered moment's display noun. [[guardian-survival-watches]] covers
the boot-seeded watches' own distinct trigger path (`runSurvivalTrigger`), which
does not route through this note's `runTrigger`. Spec:
`specs/029-metatron-agency/` (TASK-27) — `spec.md` US2/US3/US4/US6,
`contracts/routing.md`.
