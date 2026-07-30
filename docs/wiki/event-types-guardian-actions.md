---
name: event-types-guardian-actions
description: Guardian miracle-action and gru event rows split from [[event-types]]: guardian.time_snapped/item_granted/entity_moved/entity_removed, guardian.region_named (spec 101, the canonization miracle), the gru.* family. Load when tracing miracle cost/gratis mechanics, the canonization miracle's charge shape, or the gru antagonist's emergence/attack lifecycle including the spec 044 escalated-kill path.
kind: concept
sources:
  - internal/sim/miracles.go
  - internal/sim/gru.go
  - internal/sim/guardian.go
  - internal/sim/regions.go
verified_against: 0af53ec6d211c71e298072c045c67ccbbd13b61d
---

# Event types — guardian miracle-action & gru events

Back to [[event-types]] for the payload-grammar conventions and the full
event-domain index.


Spec 086 (agent-named payloads): every agent-referencing field in this
family's payloads is a `sim.AgentRef` — the wire carries
`{"id":N,"name":"…"}` objects (lists element-wise), the name stamped at
emission from the fixed roster via `Ref`/`Refs`; sentinels marshal
`{"id":-1,"name":""}`. Legacy bare-int rows decode through the dual-shape
unmarshal forever and reducer arms fold `.ID`s only — the conventions and
the normative back-compat matrix live on [[event-types]] ("Agent
references are named refs").
| Type | Payload struct | Emitted by | Reducer effect |
|---|---|---|---|
| `guardian.time_snapped` | `TimeSnappedPayload{to_tick, gratis}` in `internal/sim/miracles.go` | angel's turn reply or the `promptworld miracle` CLI/IPC door (spec 016, [[guardian-miracles]]), injected via `InjectSocial` | rejects a target at or before the current tick (forward-only); spends 2 charges (the dearest miracle) unless `gratis`; `rebaseTicks` shifts every relative-duration field forward by the jump so remaining durations are preserved, then `State.Tick = to_tick`; the skipped regeneration boundaries mint no charges |
| `guardian.item_granted` | `ItemGrantedPayload{agent, kind, qty, gratis}` | angel's turn reply or the CLI/IPC door, injected | validates a living villager, a known item kind (since TASK-163 checked against `tool.GrantKinds()`, whose rejection enumerates the full grant vocabulary instead of naming only the bad guess — [[guardian-miracle-mechanics]]), positive qty, and the bulk cap (reject-whole, never clamp); spends 1 charge unless `gratis`; adds `qty` units (a spear grant appends `qty` fresh `spearDurability` entries, kept sorted) |
| `guardian.entity_moved` | `EntityMovedPayload{class, x, y, to_x, to_y, gratis}` (`class` ∈ villager\|structure\|pile) | angel's turn reply or the CLI/IPC door, injected | validates presence at the source and the destination's placement rule (villager/pile → passable, structure → buildSite); spends 1 charge unless `gratis`; relocates the entity (a moved villager drops its intent and goes idle at the landing tick; a moved structure carries its `FuelUntil`/`Owner`/`Store`; a moved pile merges onto any pile already at the destination) |
| `guardian.entity_removed` | `EntityRemovedPayload{class, x, y, gratis}` (`class` ∈ structure\|pile\|terrain; villager is rejected — never removable) | angel's turn reply or the CLI/IPC door, injected | validates presence; spends 1 charge unless `gratis`; deletes the structure (a chest first spills its `Store` to a ground pile — goods are never silently destroyed) or the pile (with contents), or overlays the terrain through the executor's own vocabulary (tree→`Cleared`, forage→`Harvested` with regrow, rock→`Quarried`; an already-overlaid tile is rejected as a no-op target) |
| `guardian.region_named` | `RegionNamedPayload{id, x, y, radius, name, feature_kind, feature_x, feature_y, gratis}` in `internal/sim/regions.go` | the `canonize_region` tool, injected via `InjectSocial` | validates id/radius/name bounds, world bounds, circle-overlap against every existing named region (the second-christening refusal), the active cap, and — when `feature_kind` is given — its membership in the narrower `canonizeFeatureKinds` vocabulary (excludes fire/chest), containment within the region, and `buildSite`; spends **2 charges flat** unless `gratis` (the D4 charge-shape decision — the dearest-miracle band, no cooldown state); appends the `Region` artifact (spec-084 shape: deterministic id, no terminal event in v1) and, when named, a fresh `Structure` at the feature site |
| `gru.emerged` / `gru.moved` / `gru.sighted` / `gru.attacked` / `gru.withdrew` | **spec 104 (ruling 4): `gru.moved` is emitted on LEGACY worlds only** — under the coalescing regime gru motion is fully derived (the shared `gruMoveDecision` runs in the advancement engine each beat, [[gru]]/[[sim-state-reducer]]); its arm is retained forever so old logs replay unchanged, and the `gru.emerged` arm stamps the `Gru.Done` beat watermark while coalescing; payload structs in `internal/sim/gru.go`; `GruAttackedPayload.Health` (spec 044 US3): >= `gruWoundFloor` when pre-attack health was >= `nearDeathBelow` (healthy targets never die from one attack); may be 0 when the target was already weakened (pre-attack health < `nearDeathBelow`) — immediately followed, same batch, by `agent.died{cause:"gru"}` | `gruStep` (executor tick) | `State.Gru` lifecycle/position; sighting latch; attack sets absolute post-wound health, wakes victim, clears intent ([[gru]]); reducer-total (vanished gru no-ops); a killing blow (health 0) additionally lands the standard `agent.died` fallout in the same batch |

## Connections

The `guardian.*` miracle
family is emitted through [[guardian]]'s two doors and reduced in
`internal/sim/miracles.go` — see [[guardian-miracles]] for the cost table,
gratis doctrine, and the shift-semantics re-base taxonomy. `guardian.region_named`
(spec 101) is priced separately from that table (it is not one of the four
FROZEN `work_miracle` kinds `tool.miracleCosts` covers) but shares this
family's door/gratis shape; its region artifact clones the spec-084
designation/directive discipline ([[guardian-designations]]) rather than the
miracle payload shape, and its villager-coined name surfaces through
`describePlace` (situated place text — `internal/sim/memory.go`, not one of
this note's own sources) ahead of the generic terrain phrase.

[[gru]] covers the
escalated-kill semantics on `gru.attacked`/`agent.died`.
