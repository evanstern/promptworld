---
name: guardian-canonization
description: The canonization miracle (spec 101) — named regions as durable spec-084-shaped artifacts, the guardian.region_named door (2-charge premium, optional existing-kind feature, circle-overlap refusal), the D1 place-text integration, the D3 unmodified-observation-channel wiring, and the read-only myth-briefing surface (DominantPlaceMyths). Load when tracing region_named/canonize_region/brief_myths.
kind: component
sources:
  - internal/sim/regions.go
  - internal/sim/myths.go
  - internal/guardian/canonize.go
verified_against: cf65debb44c1e17b54c0f3421d11e1e8cc28576c
---

# Guardian canonization — the guardian makes consensus lore real

Spec 101's "yes, and" answer to emergent mythology (Thornspire, 2026-07-23):
where spec 097 ([[executor-perception-observation]]) only lets ground truth
DEBUNK a villager-coined myth, canonization lets the guardian ANSWER it —
christen a named region and, optionally, raise one feature within it, so a
myth the village invented becomes geography the world actually contains.

## The Region entity (`internal/sim/regions.go`)

`sim.Region` clones the spec-084 designation/directive entity discipline
verbatim ([[guardian-designations]]): a deterministic id
(`reg-<placedTick>-<seq>`, the `nextPlanID` shape — no RNG), a
reducer-stamped `PlacedSeq`, and `prunePlanEntities` reuse for shape parity.
UNLIKE a designation, a region has **no terminal event in v1** — renames and
decommissioning are named future work in the spec's own edge cases — so
`Status` is always `"active"`; the prune call is consequently inert by
construction (nothing is ever non-active to drop), kept only so the shape
matches its siblings exactly. `X`/`Y`/`Radius` describe a circle (not one of
the spec-082 target-grammar forms); `Name` is the villager-coined toponym,
1..80 runes (the designation `Label` cap). `GuardianRegionCap` (16) bounds
concurrent regions, matching `GuardianDesignationCap`.

`guardian.region_named` (validate-not-clamp, the InjectSocial dry-run
contract) checks, in order: id uniqueness, radius bounds
(`regionRadiusMin`/`Max`, 2..24), name bounds, world bounds on the center,
**circle-overlap refusal against every existing region** (`circlesOverlap` —
touching circles, distance exactly the sum of radii, are NOT overlap, so
regions may sit edge-to-edge; the spec's own named edge case: "second
christening of an overlapping region refuses at the door — one name per
ground truth"), and the active cap. An OPTIONAL feature (`FeatureKind`
non-empty signals presence, the `send_vision` place-grant convention) is
further checked: membership in `canonizeFeatureKinds` (a NARROWER mirror of
`BuildableStructureKinds` — `shelter`/`oven`/`wall_plank`/`wall_stone`/`path`;
`fire`/`chest` are excluded, their `FuelUntil`/`Owner`+`Store` fields
carrying per-agent-linked lifecycle a divine act has no natural owner for),
containment within the region's circle, and `buildSite` (the SAME rule
`agent.built` and `entity_moved` reuse — no re-derived occupancy check). A
landed feature mirrors `agent.built`'s Structure construction (a wall gets
full `wallMaxHP`; nothing else needs a special field).

**D4 — the charge-shape decision (RECORDED):** canonization costs
`GuardianRegionCharge` (2 flat charges), unconditionally, no cooldown. The
spec offered two doctrine-legal shapes — a flat premium OR a charge plus a
cooldown; a cooldown needs new per-world cooldown state this feature has no
other use for, while 2 charges reuses the existing "dearest miracle" band
(`guardian.time_snapped`'s price, [[guardian-miracles]]) with zero new
state — the simpler choice. The spend is inline in the reducer arm (the
`guardian.nudged` shape, [[sim-state-apply-guardian-records]]), not the
`miracleCost`/`MiracleCostsByEvent` table ([[tool-registry-guardian-tools]]):
`guardian.region_named` is not one of the four FROZEN `work_miracle` kinds
that table prices, so its cost lives locally as `sim.GuardianRegionCharge`,
cross-referenced by comment into `tool.canonize_region`'s `Cost{Charges: 1}`
declaration (the gate minimum, not the price — the `send_vision`/`prophesy`
precedent). `Gratis` rides the payload for operator-door parity (D4:
"gratis/operator door unchanged") but has no tool-schema param — the
guardian can never waive it, `work_miracle`'s structural-absence guarantee.

## Place-text integration (D1) — `describePlace`/`featureDesc`

A tile inside a canonized region's circle (`regionAt`, `internal/sim/regions.go`)
resolves to the region's coined `Name` in `featureDesc`
(`internal/sim/memory.go`) — checked FIRST, ahead of the structure/terrain
scan, so a villager-coined toponym is the strongest place identity there is:
every situated memory, every conversation gist, and the chronicle narrator's
source material (all built through `describePlace`/`PlaceAt`,
[[executor-social-perception]]) read "at Thornspire (x,y)" for any position
inside the region, even one that also holds a fire or stands in the woods.
Outside every region, behavior is byte-identical to pre-101.

## D3 — the observation channel needs no new machinery

Spec 097's arrival observation ([[executor-perception-observation]]) already
scans `s.Structures` and terrain within `placeScanRadius` into the
exhaustive `Kinds` set (`observedKinds`, `internal/sim/observe.go` —
UNTOUCHED by spec 101). A canonized feature is just another `Structure`, so
it participates in that scan identically to a villager-built one. Mind-side
belief reconciliation (`internal/mind/reconcile.go`, also untouched) matches
a belief's coordinate-referenced feature words against that same `Kinds`
set; a myth naming a real feature (or real ground truth already inside the
region — trees, rock, water) confirms on the next arrival purely because the
ground truth it names is now actually there. The region's NAME plays no
role in belief matching (reconcile.go has no naming vocabulary) — only in
the place-TEXT layer above.

## The myth briefing (D5) — `internal/sim/myths.go`

`State.DominantPlaceMyths(topN)` is a READ-ONLY derivation over the existing
belief corpus — no new events, no stored state, no self-grading, computed
fresh on every call. It walks every LIVING agent's world beliefs
(`Belief.Subject == -1`; a belief about a villager is never a place-myth
candidate, and `Rumor.Subject` is always an agent index, so there is no
rumor-to-place linkage either), extracts a `(x,y)` coordinate via a regex
mirroring `internal/mind/reconcile.go`'s `statementCoordRe` (declared again
here, not imported — sim cannot see mind, the `tool.MiracleKinds`
"mirrored, not imported" pattern), and clusters beliefs into coarse
coordinate buckets (`mythClusterCell`, 8 tiles) so independently-worded
tellings of the same myth still count as one candidate. Each cluster
surfaces its most-repeated wording, holder count, and average confidence;
candidates rank by holders then confidence, capped at `topN` (0 = every
cluster).

## The guardian-side working (`internal/guardian/canonize.go`)

`landCanonizeRegion` mints the `"reg"`-prefixed id (`nextPlanID`,
[[guardian-designations]]'s per-prefix same-tick disambiguator, one more
case) and lands the event through `InjectSocial`, mapping a door rejection
to in-fiction counsel (`canonizeRefusal`, the `planRefusal` shape). The
turn-level pre-check (`charges < sim.GuardianRegionCharge`) mirrors
`landMiracle`'s "no charges banked" gate. `handleBriefMyths` never touches
the live replica: `mt.myths` is a stateMu-guarded cache recomputed in
`mirrorState()` alongside every other turn-worker mirror on every absorbed
batch — the `survey_site` discipline (read tools never race the absorb
goroutine's unlocked `mt.replica.Apply`).

## Operational notes

`regions_test.go` pins the door validation/refusal table, the artifact
discipline, the D1 `describePlace` priority, the D3 `observedKinds` wiring
proof, and replay byte-identity. `myths_test.go` pins `DominantPlaceMyths`'
clustering/ranking. `canonize_test.go` pins the door/refusal/grant/
charge-gate paths and the `CanonizeFeatureKinds` mirror-drift pin.

## Connections

[[guardian-designations]] is the entity-shape precedent Region clones (no
lifecycle beyond christening, unlike designations' fulfilled/cancelled
doors). [[guardian-miracles]]/[[tool-registry-guardian-tools]] document the
charge-priced tool family `canonize_region` joins, at its own separately
priced footing. [[executor-perception-observation]] owns the spec-097
channel this feature deliberately leaves untouched. [[event-types-guardian-actions]]
carries `guardian.region_named`'s event-catalog row.
