---
name: event-types
description: Event taxonomy overview — the payload-struct/canonical-JSON convention, the outcome-payload doctrine, and a routing index into the 15 domain-split event catalogs (clock/world, agent intents/vitals, mental map, harvesting, crafting/building, social memory/consolidation, social protocol, cognition telemetry, curriculum, scenario incidents, guardian orders/morgue/actions). Load a child for a type's payload/reducer row; load this note for cross-cutting conventions.
kind: concept
sources:
  - internal/sim/state.go
  - internal/sim/gru.go
verified_against: 4c66d240b2715706964f02cfd2396256c9957d8e
---

# Event types

Every event has a namespaced `type` and a canonical-JSON payload defined as a Go
struct in `internal/sim` (structs, never maps, so bytes are deterministic; core
payloads live in `state.go`, families note their own file below).
This catalog is the contract downstream consumers (chronicle, guardian digests, the
TUI) will read. [[event-log]] stores every event; [[ipc-protocol]] pushes them to
subscribers verbatim.

## Event catalog, by domain

The full per-type catalog (92 event types across the format's history, specs
012 through 077) is split by event domain — each child inherits this note's
`verified_against` pin, carries the domain's own history-of-format-changes
prose and catalog rows verbatim, and links back here. Spec 077's seven new
types: the scenario-incident family — `sim.cold_snap`, `sim.forage_blighted`,
and the stranger entity's `stranger.arrived` / `stranger.moved` /
`stranger.took` / `stranger.departed` ([[event-types-scenario-incidents]]) —
plus the guardian's `metatron.skills_observed` skills observation
([[event-types-guardian-morgue]], the `metatron.charter_observed` twin).

- [[event-types-clock-world]] — Clock/scheduler and world-lifecycle events — pause/resume/speed/governor, day/night, forage regrowth, world genesis/migration/forking (`world.forked`, spec 076), daemon lifecycle and LLM-provider warnings.
- [[event-types-agent-intents]] — Agent intent lifecycle — intent_set/work_started/intent_done/recovery_stalled/build_failed/moved, including the spec 062/064 yield-window and needs-conditioned recovery arms.
- [[event-types-agent-vitals]] — Agent vitals and mortality — needs_changed, died (with the spec 044 death ledger/grave), run.ended, and sleep/wake.
- [[event-types-mental-map]] — Perception and mental-map events — moved, saw, map_corrected, place_told, place_revealed (spec 041's per-agent spatial knowledge family).
- [[event-types-harvesting-consumption]] — Harvesting and consumption — forage/chop/hunt/quarry/collect_water yields, food rot, cook/bathe/refuel/eat, spear/axe breakage, fire burnout.
- [[event-types-crafting-building]] — Crafting, building and goods movement — crafted/built, wall chip/destroy/repair, drop/pick_up/deposit/withdraw, and the spec 012/013/032 format-bump history.
- [[event-types-social-memory]] — Social and memory-authoring events — talked, memory_added, thought, the social.* family, chest_taken, and the spec 061 PairTalks damper.
- [[event-types-memory-consolidation]] — Memory embedding and consolidation — memory_embedded/situation_embedded, journal entries, the nightly consolidation family, belief_reinforced.
- [[event-types-social-protocol]] — Governance and hail-protocol events — meeting/norm families, convention_established, gathering_observed, and the hail family.
- [[event-types-cognition-telemetry]] — Cognition telemetry and planning — cog.thought/outcome/tool_call/memory_divergence, intent_rejected, plan_set/plan_step_started/plan_expired.
- [[event-types-curriculum-events]] — Curriculum-ladder events — exercise_passed and stage_unlocked, the spec 046/054 staged-world unlock gates (emission generalized to the nine-exercise catalog by spec 077).
- [[event-types-scenario-incidents]] — Scenario-incident events (spec 077) — cold snap, forage blight, and the stranger entity family: ambient-indistinguishable authored pressure.
- [[event-types-guardian-orders]] — Guardian standing-order events — charge_regenerated, nudged, order_placed/triggered/cancelled/expired, and the spec 029/059 survival-watch lifecycle.
- [[event-types-guardian-morgue]] — Guardian morgue and report-card events — charter_observed (+ spec 077's skills_observed twin), morgue.epilogue, guardian.report_card, chronicle.entry.
- [[event-types-guardian-actions]] — Guardian miracle actions and gru events — time_snapped/item_granted/entity_moved/entity_removed, and the gru emerged/moved/sighted/attacked/withdrew family.

## Conventions

Conventions: `clock.*` are applied player/scheduler commands; `sim.*`, `agent.*`,
and (spec 044) `run.*` are world happenings (pure functions of state + seed +
tick — `run.ended` is executor-emitted, never injectable); `morgue.*` (spec 044)
is injected narrator prose about the run, the chronicle's pattern; `daemon.*` are process
bookkeeping, wall-time dependent, and excluded from determinism comparisons (as are
`clock.*` in the binary-level test, since their ticks depend on command timing).
Payloads record **outcomes** (positions reached, absolute need values), never dice
rolls, so replay needs no RNG. Unknown types are no-ops in the reducer, so adding
types is backward-compatible with old replay code. The `cog.*` family
(TASK-32, [[cognition]]) is recorded observability —
explicit reducer no-ops whose wall-time fields are recorded input, so no failure
is silent and thought chains are walkable from the log alone; their payload
field order is canonical per `specs/007-cognition-horizon/contracts/events.md`.
`agent.intent_rejected` shares that observability role but since spec 043 is
no longer a no-op — its ring append (deterministic from the event alone) keeps
it replay-safe.
`world.migrated` (spec 012 US6) is the one exception to "payloads are small
outcomes" — its payload embeds the entire canonical `sim.State`, by design: it is
the single record standing in for the whole pre-break history, and the reducer's
`state.Seed` check keeps it total (a mismatched payload no-ops rather than erroring).

## Operational notes

The outcome-payload
convention ([[deterministic-rng]]) is load-bearing — keep it; `gru.attacked`
carrying absolute post-wound health (never the wound roll) is the pattern.
