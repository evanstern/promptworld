---
name: event-types-guardian-actions
description: Guardian miracle-action and gru event rows split from [[event-types]]: metatron.time_snapped/item_granted/entity_moved/entity_removed, the gru.* family. Load when tracing miracle cost/gratis mechanics or the gru antagonist's emergence/attack lifecycle including the spec 044 escalated-kill path.
kind: concept
sources:
  - internal/sim/miracles.go
  - internal/sim/gru.go
  - internal/sim/guardian.go
verified_against: 22bb41c887ef6a34c55a77b9b989b299f4dc6857
---

# Event types — guardian miracle-action & gru events

Back to [[event-types]] for the payload-grammar conventions and the full
event-domain index.

| Type | Payload struct | Emitted by | Reducer effect |
|---|---|---|---|
| `metatron.time_snapped` | `TimeSnappedPayload{to_tick, gratis}` in `internal/sim/miracles.go` | angel's turn reply or the `promptworld miracle` CLI/IPC door (spec 016, [[guardian-miracles]]), injected via `InjectSocial` | rejects a target at or before the current tick (forward-only); spends 2 charges (the dearest miracle) unless `gratis`; `rebaseTicks` shifts every relative-duration field forward by the jump so remaining durations are preserved, then `State.Tick = to_tick`; the skipped regeneration boundaries mint no charges |
| `metatron.item_granted` | `ItemGrantedPayload{agent, kind, qty, gratis}` | angel's turn reply or the CLI/IPC door, injected | validates a living villager, a known item kind, positive qty, and the bulk cap (reject-whole, never clamp); spends 1 charge unless `gratis`; adds `qty` units (a spear grant appends `qty` fresh `spearDurability` entries, kept sorted) |
| `metatron.entity_moved` | `EntityMovedPayload{class, x, y, to_x, to_y, gratis}` (`class` ∈ villager\|structure\|pile) | angel's turn reply or the CLI/IPC door, injected | validates presence at the source and the destination's placement rule (villager/pile → passable, structure → buildSite); spends 1 charge unless `gratis`; relocates the entity (a moved villager drops its intent and goes idle at the landing tick; a moved structure carries its `FuelUntil`/`Owner`/`Store`; a moved pile merges onto any pile already at the destination) |
| `metatron.entity_removed` | `EntityRemovedPayload{class, x, y, gratis}` (`class` ∈ structure\|pile\|terrain; villager is rejected — never removable) | angel's turn reply or the CLI/IPC door, injected | validates presence; spends 1 charge unless `gratis`; deletes the structure (a chest first spills its `Store` to a ground pile — goods are never silently destroyed) or the pile (with contents), or overlays the terrain through the executor's own vocabulary (tree→`Cleared`, forage→`Harvested` with regrow, rock→`Quarried`; an already-overlaid tile is rejected as a no-op target) |
| `gru.emerged` / `gru.moved` / `gru.sighted` / `gru.attacked` / `gru.withdrew` | payload structs in `internal/sim/gru.go`; `GruAttackedPayload.Health` (spec 044 US3): >= `gruWoundFloor` when pre-attack health was >= `nearDeathBelow` (healthy targets never die from one attack); may be 0 when the target was already weakened (pre-attack health < `nearDeathBelow`) — immediately followed, same batch, by `agent.died{cause:"gru"}` | `gruStep` (executor tick) | `State.Gru` lifecycle/position; sighting latch; attack sets absolute post-wound health, wakes victim, clears intent ([[gru]]); reducer-total (vanished gru no-ops); a killing blow (health 0) additionally lands the standard `agent.died` fallout in the same batch |

## Connections

The `metatron.*` miracle
family is emitted through [[guardian]]'s two doors and reduced in
`internal/sim/miracles.go` — see [[guardian-miracles]] for the cost table,
gratis doctrine, and the shift-semantics re-base taxonomy.

[[gru]] covers the
escalated-kill semantics on `gru.attacked`/`agent.died`.
