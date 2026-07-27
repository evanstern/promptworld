---
name: guardian-designations
description: The guardian's durable plan layer (spec 084) — designations (event-sourced world plan artifacts with structural fulfillment predicates), hard directives (TTL-bounded villager bindings landed through the injection door), the DIRECTIVE reflex rung between SURVIVAL and PREP, the neverDrop directive context block, and the charge-free survey_site read tool. Load when tracing place_designation/issue_directive, the designation./directive. event lifecycle, or why a villager walks to a marked site.
kind: component
sources:
  - internal/sim/plans.go
  - internal/sim/policy.go
  - internal/sim/executor.go
  - internal/guardian/plans.go
  - internal/guardian/survey.go
  - internal/tool/registry.go
  - internal/mind/context.go
verified_against: c61cd6c04ddfcd2a976c14a49ba071e8fd768a73
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

## The announcement grant

The `designation.placed` arm upserts one
`PlaceFact{Kind:"designation", Provenance: revealed, Seen: e.Tick}` at the
anchor tile (point tile / line first endpoint / rect min corner) into EVERY
living villager's mental map — the spec-041 place-grant machinery fanned
out reducer-side (one event, deterministic; map-less agents skipped, the
reducer stays total). `"designation"` joins the closed `PlaceFact`
vocabulary with a 7-game-day horizon (the max directive TTL);
`placeFactKinds` (send_vision's reveal enum) is NOT extended.
`renderKnownPlaces` renders the fact as an individually-named landmark
enriched from `State.Designations` (`designationLandmark`,
`internal/mind/prompt.go` — `PlaceFact.Detail` is an int64 scalar and
cannot carry a phrase). Cancellation/fulfillment never retracts the fact —
remembered history, the burned-out-fire precedent.

## The villager side — hardness (operator ruling, FIRM)

- **DIRECTIVE reflex rung** (`directiveDecision`,
  `internal/sim/policy.go`): SURVIVAL → **DIRECTIVE** → (prepYields?) PREP
  → wander. Unconditioned by the yield window and danger bands — survival
  always preempts (the rung sits after `survivalDecision`), directives
  preempt prep/wander (before the `prepYields` consult). The rung is
  STATELESS: it re-derives the oldest active directive addressing the
  agent from state at every idle decision, so interruption-resume needs
  zero new code (hails/conversations pause intents through the existing
  machinery; the rung simply re-resolves afterward). Routing: build at
  the site when reflex-expressible with materials in hand
  (`reflexBuildable` — the closed fire/chest/oven set); wall lines build
  the first gap in enumeration order (adjacent-stand) or walk toward it;
  zones walk to presence; otherwise walk to the site via the new
  `heed_directive` goal (instant on arrival, the `search` completion
  shape, never model-facing) or fall through to the ladder — the planner,
  reading the block, owns the clever part (no idle-at-site deadlock).
- **`directive` context block** ([[decision-context]],
  [[context-block-inventory]]): the eleventh spec-043 block, between
  `plan_echo` and `known_places`, priority `neverDrop` — ≤2 active
  directives addressing the agent, oldest first: framing text VERBATIM
  (recorded event data — the firewall's only prose channel, riding the
  issue batch's per-target "The Guardian charges you: …" companion
  memories exactly like visions), the bound designation's
  kind/site/requirement, plain-words days left. Empty renders `""` — a
  directive-free prompt is byte-identical to pre-084.

## Guardian-side surfaces

Tool handlers (`internal/guardian/plans.go`) parse the call, mint ids, and
land through `InjectSocial`, mapping door rejections to repairable
`rejected_gate` counsel (`planRefusal`); the turn prompt carries active
designations and directives (`writeDesignations`/`writeDirectives` — id,
kind, site, days-left, the `writeStandingOrders` shape) so counsel stays
truthful (FR-015). `survey_site` (`internal/guardian/survey.go`) is Effect
Read / Gate None — a deterministic site fact sheet (terrain mix, nearest
water/tree/rock, structures kind+tile, passability incl. wall-blocked
tiles) built turn-side from the absorb mirrors + static map (the
`buildTargetingDigest` pattern); out-of-bounds input returns a repairable
in-fiction miss naming the world bounds, never a hard error; radius clamps
1..8, default 4. On the TUI map, active designations render through three
appended [[tile-registry]] rows (`◇` site, `┄` wall-line segment, `◦` zone
perimeter — interior unmarked) beneath every real entity; consumed marks
stop rendering.

Time snap: an ACTIVE directive's `ExpiresTick` SHIFTs (the
`GuardianOrders` classification verbatim); everything else on both
entities is history/identity KEEP
([[guardian-miracle-rebase-taxonomy]]).

**Scope guards** (FR-016): no mission machinery (TASK-158), no faith
accounting (TASK-118 consumes the `directive.fulfilled` contract), no
bundle grammar/matrix change, no interruption-machinery change.
