---
name: event-types
description: Event taxonomy overview — the payload-struct/canonical-JSON convention, the outcome-payload doctrine, and a routing index into the 15 domain-split event catalogs (clock/world, agent intents/vitals, mental map, harvesting, crafting/building, social memory/consolidation, social protocol, cognition telemetry, curriculum, scenario incidents, guardian orders/morgue/actions). Load a child for a type's payload/reducer row; load this note for cross-cutting conventions.
kind: concept
sources:
  - internal/sim/state.go
  - internal/sim/gru.go
verified_against: PENDING_MERGE_SHA
---

# Event types

Every event has a namespaced `type` and a canonical-JSON payload defined
as a Go struct in `internal/sim` (structs, never maps, so bytes are
deterministic; core payloads live in `state.go`, families note their own
file below). This catalog is the contract downstream consumers (chronicle,
guardian digests, the TUI) read. [[event-log]] stores every event;
[[ipc-protocol]] pushes them to subscribers verbatim.

**Type names are versioned** (spec 094): a rename = `store.LogFormatVersion`
bump + translating migration ([[event-log]]); `guardian.*` action types
were `metatron.*` in log format 1.

## Event catalog, by domain

The full per-type catalog (93 types, specs 012 through 083) is split by
event domain — each child inherits this note's `verified_against` pin and
carries the domain's own format-history prose and catalog rows.

- [[event-types-clock-world]] — Clock/scheduler and world-lifecycle events — pause/resume/speed/governor, day/night, forage regrowth, genesis/migration/forking (`world.forked`, spec 076), daemon lifecycle and LLM-provider warnings.
- [[event-types-agent-intents]] — Agent intent lifecycle — intent_set/work_started/intent_done/recovery_stalled/build_failed/moved, incl. the spec 062/064 yield-window and needs-conditioned recovery arms.
- [[event-types-agent-vitals]] — Agent vitals and mortality — needs_changed, died (spec 044 death ledger/grave), run.ended, sleep/wake, spec 083's neglect_detected percept.
- [[event-types-mental-map]] — Perception and mental-map events — moved, saw, map_corrected, place_told, place_revealed (spec 041's spatial knowledge family).
- [[event-types-harvesting-consumption]] — Harvesting/consumption — forage/chop/hunt/quarry/collect_water yields, food rot, cook/bathe/refuel/eat, spear/axe breakage, fire burnout.
- [[event-types-crafting-building]] — Crafting/building/goods movement — crafted/built, wall chip/destroy/repair, drop/pick_up/deposit/withdraw, spec 012/013/032 format-bump history.
- [[event-types-social-memory]] — Social/memory-authoring events — talked, memory_added, thought, the social.* family, chest_taken, spec 061's PairTalks damper.
- [[event-types-memory-consolidation]] — Memory embedding/consolidation — memory_embedded/situation_embedded, journal entries, nightly consolidation, belief_reinforced.
- [[event-types-social-protocol]] — Governance/hail-protocol events — meeting/norm families, convention_established, gathering_observed, the hail family.
- [[event-types-cognition-telemetry]] — Cognition telemetry/planning — cog.thought/outcome/tool_call/memory_divergence, intent_rejected, plan_set/plan_step_started/plan_expired.
- [[event-types-curriculum-events]] — Curriculum-ladder events — exercise_passed/stage_unlocked, spec 046/054 staged-world unlock gates (generalized to nine exercises by spec 077).
- [[event-types-scenario-incidents]] — Scenario-incident events (spec 077) — cold snap, forage blight, the stranger entity family: ambient-indistinguishable authored pressure.
- [[event-types-guardian-orders]] — Guardian standing-order events — charge_regenerated, nudged, order_placed/triggered/cancelled/expired, spec 029/059 survival-watch lifecycle.
- [[event-types-guardian-morgue]] — Guardian morgue/report-card events — charter_observed (+ spec 077's skills_observed twin), morgue.epilogue, guardian.report_card, chronicle.entry.
- [[event-types-guardian-actions]] — Guardian miracle actions and gru events — time_snapped/item_granted/entity_moved/entity_removed, the gru emerged/moved/sighted/attacked/withdrew family.
- [[event-types-guardian-plans]] — Guardian plan-layer events (spec 084) — `designation.*`/`directive.*`: injected placement/issue/cancel, executor-emitted fulfillment/expiry, the TASK-118 faith seam (consumed by spec 085).
- [[guardian-faith]] — Faith-economy events (spec 085) — `faith.changed` (executor-emitted, the five-reason delta table) and the prophecy lifecycle `prophecy.declared`/`fulfilled`/`failed`.

## Conventions

Conventions: `clock.*` are applied player/scheduler commands; `sim.*`,
`agent.*`, and (spec 044) `run.*` are world happenings (pure functions of
state + seed + tick — `run.ended` is executor-emitted, never injectable);
`morgue.*` (spec 044) is injected narrator prose about the run, the
chronicle's pattern; `daemon.*` are process bookkeeping, wall-time
dependent, excluded from determinism comparisons (as are `clock.*` in the
binary-level test, since their ticks depend on command timing). Payloads
record **outcomes** (positions reached, absolute need values), never dice
rolls, so replay needs no RNG. Unknown types are no-ops in the reducer, so
adding types is backward-compatible with old replay code. The `cog.*`
family (TASK-32, [[cognition]]) is recorded observability — explicit
reducer no-ops whose wall-time fields are recorded input, so no failure is
silent and thought chains are walkable from the log alone; field order is
canonical per `specs/007-cognition-horizon/contracts/events.md`.
`agent.intent_rejected` shares that role but since spec 043 is no longer a
no-op — its ring append (deterministic from the event alone) keeps it
replay-safe. `world.migrated` (spec 012 US6) is the one exception to
"payloads are small outcomes" — its payload embeds the entire canonical
`sim.State`, by design: the single record standing in for the whole
pre-break history, and the reducer's `state.Seed` check keeps it total (a
mismatched payload no-ops rather than erroring).

### Agent references are named refs (spec 086)

Every agent-referencing payload field is a `sim.AgentRef` — the wire
carries `{"id":2,"name":"Cedar"}` objects (and `[{…},{…}]` lists), the
name stamped at emission from the fixed `AgentNames` roster constant via
the `Ref`/`Refs` constructors, so the log is self-describing with no
replica. Sentinels (canonically −1: any/none/personal) marshal
`{"id":-1,"name":""}` — a fake name on a sentinel is as much a bug as a
missing name on an agent. Four state-shared entities split into wire
mirrors carrying refs while state keeps bare ints
(`DirectiveIssuedPayload`, `OrderPlacedPayload`, `ProphecyDeclaredPayload`,
`DeathRef` in `run.ended`) — no `AgentRef` is ever reachable from
`sim.State` (`TestNoAgentRefInState`). Enforcement is mechanical: the
typed ref, append validation at both live-emission doors (`mustPayload`
panics; `InjectSocial` decodes via `sim.PayloadCatalog` and refuses the
batch), and `TestPayloadAgentRefSweep` over the catalog (frozen tag
vocabulary + frozen allowlist: `PlaceFact.Source`,
`CogToolCallPayload.Args`, `IntentSetPayload.Source`,
`FaithChangedPayload.SourceID`, `PlanStep.Target`, `Guard.Target` — the
last two state-resident, the PlaceFact class).

**The back-compat matrix (normative, permanent — spec 086 data-model §6):**

| Input | Decoder behavior | Renderer behavior |
|---|---|---|
| pre-086 row, `"agent":2` | dual-shape unmarshal → `{2, ""}`; arm reads `.ID` — identical fold | name `""` → replica fallback (`agentName`); grammar-miss rows still via `resolvePayloadNames` |
| post-086 row, `{"id":2,"name":"Cedar"}` | object branch; arm reads `.ID` | payload name rendered directly; no replica needed |
| pre-086 snapshot | state shapes unchanged; decodes as today | n/a |
| migrated v1–v4 world | `world.migrated` embeds State — shape untouched; migration never rewrites payloads | historic rows fall back as above |
| pre-086 log, from-genesis replay | byte-identical `State.Marshal()`/`Hash()` | n/a |
| live pre-086 world continued | new emissions named from the next event on; old rows keep their bytes | mixed feed: new rows payload-named, old rows fallback |

Name validation NEVER lives in a reducer arm — replay accepts unnamed
shapes forever; validation exists only at the live-emission choke points,
which replay never traverses.

## Operational notes

The outcome-payload convention ([[deterministic-rng]]) is load-bearing —
keep it; `gru.attacked` carrying absolute post-wound health (never the
wound roll) is the pattern.
