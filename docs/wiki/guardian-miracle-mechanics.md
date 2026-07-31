---
name: guardian-miracle-mechanics
description: The four miracle event types (time_snapped/item_granted/entity_moved/entity_removed) — reducer dispatch, per-arm validation (validate-not-clamp, reject-whole), and the cost table/gratis doctrine (2 charges for a snap, 1 for the rest; gratis reachable only via the operator's --force door). Split from [[guardian-miracles]]; load for a specific miracle kind's validation rule or its charge cost.
kind: component
sources:
  - internal/sim/miracles.go
  - internal/tool/registry.go
  - internal/guardian/toolcalls.go
verified_against: fc1a8314f3f71a33c5e2145c914d5cbb511d9196
---

# Guardian's miracle mechanics

Split from [[guardian-miracles]] (summary-style, corpus-spec v2) — the four
event types' reducer semantics and the cost/gratis doctrine.

**The four event types** (`internal/sim/miracles.go`, canonical JSON, struct-ordered):

| Event | Payload | Effect |
|---|---|---|
| `guardian.time_snapped` | `TimeSnappedPayload{to_tick, gratis}` | jumps `State.Tick` forward to `to_tick`, forward-only (a target at or before the current tick is rejected whole, before any spend); shifts every relative-duration field via `rebaseTicks` first |
| `guardian.item_granted` | `ItemGrantedPayload{agent, kind, qty, gratis}` | provisions a living villager with `qty` known items, reject-whole (never clamp) if it would exceed the carry cap |
| `guardian.entity_moved` | `EntityMovedPayload{class, x, y, to_x, to_y, gratis}` (`class` ∈ villager\|structure\|pile) | relocates the entity from `(x,y)` to `(to_x,to_y)` |
| `guardian.entity_removed` | `EntityRemovedPayload{class, x, y, gratis}` (`class` ∈ structure\|pile\|terrain; villager is always rejected) | deletes the entity or overlays the terrain |

`applyMiracle` in `miracles.go` is the reducer dispatcher `sim.State.Apply` routes
these four types to (alongside `applyMetatron` for `guardian.charge_regenerated`/
`guardian.nudged` — [[sim-state-reducer]]). Every arm's validation — presence at the
source, the destination's placement rule, item kind/quantity — precedes both the
charge spend and the mutation, so a rejected miracle spends nothing and leaves no
partial application (validate-not-clamp, reject-whole):

- **`applyEntityMoved`**: `villager`/`pile` destinations must be `passable`;
  `structure` destinations must satisfy `buildSite`. A moved structure carries its
  `FuelUntil`/`Owner`/`Store` along whole; a moved pile merges onto any pile already
  at the destination (`movePile`); a moved villager drops its intent and goes idle
  at the landing tick (cancel-and-replan) — villagers may share a tile, so no
  destination-exclusivity check applies to a villager move. Since spec 041
  ([[mental-maps]]), a teleported villager also gets the SAME derived
  mental-map bookkeeping a walked step gets: its landing surroundings mark
  explored and mutual peer sightings with anyone nearby update — a miracle
  move is knowledge-transparent, not a blind teleport.
- **`applyEntityRemoved`**: a villager is always rejected ("a villager can never be
  removed" — v1 doctrine). A removed chest first spills its `Store` to a ground pile
  via `spillInventory` (the same death-spill vocabulary `agent.died` uses) before
  deletion, so goods are never silently destroyed; a removed pile is destroyed with
  its contents (the explicit, operator-visible destruction the miracle names).
  `removeTerrain` overlays a tree/forage/rock tile through the SAME vocabulary the
  executor's own harvest completions use (chop→`Cleared`, forage→`Harvested` with a
  regrow deadline, quarry→`Quarried`, permanent) — a removed tile is a state the
  executor could already have produced on its own; spec 068's marsh/sand ground
  covers are deliberately absent from this switch — they have no depleted state the
  executor could ever produce, so they fall to the same "holds no removable
  terrain" refusal grass and water already draw ([[worldmap-generation]],
  [[tile-registry]]); an already-overlaid tile is
  rejected as a no-op target.
- **`applyItemGranted`**: validates a living, in-range agent index, a `grantableKind`
  (the `Inventory` key vocabulary plus `"spear"`/`"axe"` singular), and a positive
  quantity. Since TASK-163, `grantableKind` checks a package-level `grantableKinds`
  set built once from `tool.GrantKinds()` — the SAME authoritative grant vocabulary
  `work_miracle`'s `item` param now declares as an `Enum`
  ([[tool-registry-guardian-tools]]) — rather than a second hand-written switch
  literal, so the door's acceptance can never drift from the schema the model is
  shown; the rejection for an unrecognized kind now enumerates that vocabulary
  (`"unknown item kind %q (grantable: %s)"`) instead of naming only the bad guess
  (live measurement: a guardian repeatedly guessed `"food"`/`"forage"` against the
  old bare message). One bulk per granted unit, exactly like a carried item — a grant of
  `qty` items always costs `qty` bulk regardless of kind, so the cap check is
  `bulk(*inv)+qty > bulkCap`. A spear grant appends `qty` fresh `spearDurability`
  entries to `Inv.Spears`, kept sorted ascending (hunts spend the most-worn first);
  since spec 032 (US2) an axe grant is the same clone against `Inv.Axes` with
  the same fresh-`axeDurability` value the `craft_axe` verb produces, sorted
  the same way.
- **`applyTimeSnapped`**: rejects a non-forward target before any spend or mutation;
  spends 2 charges (the dearest miracle) unless gratis; calls `rebaseTicks`, then
  sets `State.Tick = to_tick`. FR-010 (a snap mints no charges across the skipped
  regeneration boundaries) needs no code of its own — regeneration only fires when
  the executor *processes* a boundary crossing, and a snap processes no interval.

**Cost table and gratis doctrine**: the time snap costs 2 charges; every other
miracle costs 1. Since spec 021 (TASK-64) the AUTHORITATIVE per-kind table lives in
the leaf [[tool-registry]] (`tool.MiracleCost(kind)` / `tool.MiracleCostsByEvent()`,
`internal/tool/registry.go`, beside `miracleKinds`); `sim.miracleCost` (`miracles.go`,
a keyed map — never iterated into state, for determinism) is now DERIVED from
`tool.MiracleCostsByEvent()` rather than a second literal, and the guardian's turn
prompt renders costs from the same source (`tool.GuardianToolGuidance`), so one edit
propagates to enforcement and every rendering (`TestMiracleCostDerivedFromTool`
pins the derivation). Pricing remains doctrine, not caller input — a payload never
carries its own price, so replay re-validates every spend identically (R2).
`spendMiracleCharge(eventType, gratis)` is the shared validate/spend helper every
arm calls last, after all other validation passes: with `gratis` it returns
immediately, waiving ONLY the charge (every other validation still runs in full);
without it, it errors if the bank can't pay and decrements it otherwise. `gratis` is
reachable from exactly one surface: the `promptworld work --force` CLI/IPC door
(canonical since spec 052 FR-008; `promptworld miracle --force` survives as a
hidden compat alias, same handler)
([[cli-promptworld]], [[ipc-protocol]], [[ipc-server]]) — the operator's cheat door
the guardian structurally cannot reach. The guardian's turn contract — since spec 017 the
`work_miracle` tool call, parsed into `miracleArgs` (`internal/guardian/toolcalls.go`;
the retired `turnReply.Miracle` anonymous struct carried the identical flat field
set pre-loop) — has **no gratis field at all** — a model can emit `"gratis":true`
in its tool-call arguments and it is simply dropped at unmarshal, nothing to
sanitize or forget. `landMiracle` (the guardian's landing path) calls the shared
builder with `gratis` hardcoded `false`, so a model-driven miracle
is unconditionally charged (contracts §1, FR-007/SC-005).

## Connections

[[guardian-canonization]] (spec 101) clones this no-gratis-param guarantee
for `canonize_region`, a SEPARATE tool outside this four-kind family, priced
outside `miracleCost`.

[[guardian-miracles]] is the parent — see there for the two landing doors
and the shift-semantics re-base taxonomy a time snap triggers.
[[guardian]] hosts the turn's `work_miracle` tool call that reaches
`landMiracle`, which calls the shared builder these arms back.
[[worldmap-generation]] and [[tile-registry]] own the terrain vocabulary
`applyEntityRemoved`'s `removeTerrain` reuses; [[mental-maps]] owns the
derived explored/sighting bookkeeping a miracle-moved villager gets.
[[event-types]] catalogs all four payload shapes.
