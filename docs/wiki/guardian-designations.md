---
name: guardian-designations
description: The guardian's durable plan layer (spec 084) — designations (event-sourced world plan artifacts with structural fulfillment predicates) and hard directives (TTL-bounded villager bindings landed through the injection door): the entity model and event vocabulary. Villager response and guardian-turn surfaces split to [[guardian-designations-reception]]. Load when tracing place_designation/issue_directive or the designation./directive. event lifecycle.
kind: component
sources:
  - internal/sim/plans.go
  - internal/sim/executor.go
verified_against: 376afd4cee54839a545bc88409f3c485c2f5149d
---

# Guardian designations and directives — the durable plan layer

Spec 084 gives the guardian the DF/RimWorld player's plan-making verbs: a
free **survey** (`survey_site`), durable **designations** staked on the map,
and hard **directives** binding villagers to a designation. All five tools
are **charge-free** (recorded decision, research R3: charges price world
EDITS; the plan layer edits nothing physical — villagers still do all the
work by their own logic) and granted at **every curriculum stage** (the
`monitor_and_act` teaching-primitive precedent; both flagged for operator
review in the spec's Assumptions).

## The two entities (`internal/sim/plans.go`)

`sim.Designation` and `sim.Directive` clone the [[guardian-orders]] entity
discipline verbatim: deterministic ids (`dsg-<tick>-<seq>` /
`dir-<tick>-<seq>`, the `nextOrderID` shape — `nextPlanID`,
`internal/guardian/plans.go` — no RNG), one-way status doors
(`transitionDesignation`/`transitionDirective`: exactly one terminal ever
lands; the race loser finds a non-active entity and is refused),
reducer-stamped `PlacedSeq` (payload value ignored), validate-not-clamp
arms, and an active + most-recent-32-non-active retention prune
(`prunePlanEntities`, the `pruneGuardianOrders` algorithm generalized).
State rides `State.Designations`/`State.Directives` (`omitempty` — pre-084
snapshots load byte-identical, no format bump).

- **Designation**: kind `structure_site` (point + required buildable
  `StructureKind`) | `wall_line` (axis-aligned line, optional wall-kind
  narrowing, ≤32 tiles) | `settlement_zone` (normalized rect, ≤256 tiles,
  `MinStructures` 1..12 default 3); loci stored NORMALIZED as ints —
  payloads are self-contained, replay never re-parses. Status
  `active → fulfilled | cancelled`. Cap **16 active**
  (`GuardianDesignationCap`). Door-side occupancy fulfillability: a
  different-kind structure on a site tile (or a non-wall structure on a
  line tile) is rejected; zones are never occupancy-checked; a same-kind
  pre-existing structure is legal (consecrate what stands — the sweep
  fulfills at the next boundary).
- **Directive**: binds living villagers (`Targets`, resolved to ascending
  indices at issue; `Village` marks an "everyone" issue) to an ACTIVE
  designation, with the guardian's framing `Text` (1..400 runes) and a TTL
  (1..7 game days, default 3 — the SHARED `GuardianOrderTTL*` constants).
  Status `active → fulfilled | cancelled | expired`. Cap **3 active**
  (`GuardianDirectiveCap`).

**Addressing — the one-parser law**: designation loci parse through
`target.ParseLocus` (`internal/target`), a bare-locus entry point over the
SAME spec-082 grammar/normalization/`Tiles()` enumeration the bundle
compiler uses; the 082 class table and reserved-prefix surface are
untouched ([[bundle-tools]]). `sim.DesignationTiles` reconstructs the
enumeration from the stored ints for the fulfillment predicates and the
map renderer — no designation-side grammar copy exists.

**Fulfillment predicates** (`designationFulfilled` — pure state checks, no
clock/RNG/IO, evaluated identically by the executor sweep and the reducer
arm): structure-of-kind-at-tile (site); a qualifying wall on EVERY
enumerated line tile; ≥ `MinStructures` structures within the rect.

## The seven-event vocabulary and its doors

Contract: `specs/084-guardian-directives/contracts/events.md` (normative).
The [[guardian-orders]] split, verbatim ([[sim-loop-injection-doors]]):

| Type | Door |
|---|---|
| `designation.placed` / `designation.cancelled` / `directive.issued` / `directive.cancelled` | INJECTED via `InjectSocial` (whitelist + validating arms in `applyPlan`) by the four guardian tools |
| `designation.fulfilled` / `directive.fulfilled` / `directive.expired` | EXECUTOR-emitted from `stepEvents` — pure over (state, tick), the `charge_regenerated` pattern; NOT whitelisted, so a forged injection is refused structurally; replay reproduces the whole lifecycle with no guardian running |

Sweep order within a tick is fixed (research R14): designations first
(slice order), then per-directive fulfilled-before-expired — so exactly one
terminal lands per boundary, and a designation fulfilled at tick T yields
its bound directives' `directive.fulfilled` at T+1 (the documented one-tick
lag). `directive.expired` also fires when no targeted villager remains
alive (the un-executable clause — no TTL wait). `directive.fulfilled`
carries `{id, designation_id, targets, issued_tick}` — **the TASK-118
faith-accounting seam** (recorded contract; this spec adds no faith state —
spec 085 consumes it: the faith sweep mints `faith.changed{+8}` beside every
`directive.fulfilled` and `{−4}` beside every `directive.expired`,
[[guardian-faith]]).
All four `directive.*` types join `observableEventTypes` (enum-only,
12 → 16), so `monitor_and_act` standing orders watch the directive
lifecycle through the unmodified `matchOrders` path — zero new trigger
code (AC #7; e.g. a re-issue loop on `directive.expired` is the operator
ruling's named anti-thrash workaround).

## Reception, villager response, and turn surfaces

Split into [[guardian-designations-reception]]: the mental-map
announcement grant that reveals a placed designation to every villager,
the DIRECTIVE reflex rung (between SURVIVAL and PREP) and the `neverDrop`
directive context block that make a villager act on one, and the
guardian-side turn surfaces — tool handlers, `survey_site`, TUI rendering,
miracle rebase treatment, and the spec's scope guards (FR-016: no mission
machinery, no faith accounting here, no bundle-grammar or
interruption-machinery change).
