# Data Model — TUI Look-Cursor Mode (spec 074)

Phase 1 shapes. Everything here is client-side presentation state (`internal/tui`)
except `EnvSample` (`internal/sim`, derived read-only). Nothing persists; nothing
enters the replica, reducer, or IPC surface.

## Look-cursor mode state (Model fields)

| Field | Type | Meaning |
|---|---|---|
| `lookActive` | `bool` | the mode is on; gates the key layer, the dock borrow, the cursor render, and the `!lookActive` visibility guards |
| `lookX`, `lookY` | `int` | cursor tile, always clamped to `[0, W) × [0, H)` |
| `lookFocus` | `lookFocusKind` | which layer holds the keyboard within the mode: `lookFocusCursor` / `lookFocusPane` / `lookFocusDrill` |
| `lookSel` | `int` | selected TILE-pane row index (into the flattened selectable row list), clamped at read time like `villSelected` |
| `lookDrill` | `lookDrillRef` | open drill-in target: kind (`agent` / `pile` / `chest` / `event`) + stable index (agent index / structure identity / `m.events` index); zero value = none |
| `lookDrillScroll` | `int` | drill-in pane scroll (the `villDecisionsScroll` / `chronDetailScroll` idiom, clamped at render) |
| `mapHit` | `*mapHitRegion` | per-frame map-grid click geometry (pointer — the `chronHit` pattern) |
| `tileHit` | `*tileHitRegion` | per-frame TILE-pane click geometry (pointer) |

**Invariants**

1. `lookFocus != lookFocusCursor ⇒ lookActive` (pane/drill focus exists only inside
   the mode).
2. `lookDrill` non-zero ⇒ `lookFocus == lookFocusDrill`.
3. The minibuffer remains the focus contract's only text-capture client: no look
   state ever captures printable keys into a buffer (FR-004). Pane "focus" is a
   selection scope, drawn per focus-contract rule 2, not text focus.
4. `dockTab`, `solo`, `active`, and every per-tab state field are never written by
   look code except the exit-and-select digits, which delegate to `selectTab` after
   clearing look state (research R2).
5. Entering the mode: `lookX, lookY = cameraOrigin(vw,vh) + (vw/2, vh/2)` (camera
   center); exiting: all look fields zeroed AND `panX, panY = 0, 0` (resume
   following).

## State transitions (the esc chain, FR-003 / SC-004)

```
            v (map visible)                        ⏎ / tab                 ⏎ on row
  mode off ────────────────▶ cursor ─────────────────────────▶ pane ────────────────▶ drill
           ◀──────────────── (esc | v)  ◀───────────────────── (esc)  ◀────────────── (esc)
```

- One layer per `esc` press, exactly: drill → pane → cursor → off. `esc` from off is
  the untouched global esc.
- Digits `2`–`6`: any look state → mode off + `selectTab(k)` (prior tab state intact).
- `G` (console) and solo-zoom transitions: mode off first, then the existing behavior.
- Help overlay / takeover: layer above without changing look state; on dismiss the
  mode is exactly as it was.
- Mouse: map click → cursor state (entering the mode if off, moving the cursor if
  on, from any look focus); pane row click → pane focus + selection; second click on
  the selected row → drill.

## TILE view row model

The TILE body is a header plus an ordered list of `tileRow`s — the single source for
rendering, keyboard selection, and the `tileHit` region (they cannot disagree because
they are the same slice).

| Field | Type | Meaning |
|---|---|---|
| `band` | enum | `agents` / `pilesChests` / `structures` / `terrain` / `events` — rendered strictly in this order (FR-009; `terrain` before `events` because events are history, not contents) |
| `label` | string | the rendered row line (plain language; FR-020 default) |
| `drill` | `lookDrillRef` | zero for non-drillable rows (terrain, band headings are not rows) |

**Assembly (per frame, from cursor tile):**

- **agents**: every `replica.Agents[i]` with `(X,Y) == (lookX,lookY)` — name,
  awake/asleep/dead, needs bars, intent goal; the gru joins this band when abroad on
  the tile (registry meaning as its label, non-drillable).
- **pilesChests**: `replica.Piles` at the tile (`summarizePileContents`) and chest
  structures at the tile (`describeChest`).
- **structures**: non-chest structures at the tile — registry legend name + state
  (fire lit/dying/cold via the same `FuelUntil`/`RefuelDyingBelow` logic
  `renderMapGrid` uses; wall damaged; grave); dens (static map feature) list here.
- **terrain**: exactly one row — the effective-kind resolution `tile()` uses
  (path/quarried override, else base terrain), labeled with the registry row's
  legend name; whatis prose in the header comes from the same row's `Meaning`
  (FR-010).
- **events**: `tileEvents(lookX, lookY)` — `m.events` filtered to entries whose
  subject-registry candidate carries recorded payload coordinates equal to the tile
  (research R5), most recent first, capped to the remaining pane budget.

**Header:** `TILE (x,y) · <registry meaning>` + warmth/light meter lines (below).

## EnvSample (`internal/sim`, exported, pure)

```go
type EnvSample struct {
    Warm       bool   // ≡ warmAt(s, x, y, tick)
    WarmSource string // "fire" | "shelter" | "" (when !Warm)
    Lit        bool   // ≡ litAt(s, x, y)
}
// EnvAt(s *State, x, y int, tick int64) EnvSample
```

- Shared private cores with `warmAt` (`terrain.go`) / `litAt` (`gru.go`) — the
  predicates become wrappers, so sample and mechanics cannot disagree (SC-006).
- Day/night is NOT part of the sample (it is world state, `State.Night`); the TUI
  combines `Night + EnvSample + on-shelter` into the displayed levels:

| Condition | Warmth level + note | Light level + note |
|---|---|---|
| `WarmSource == "fire"` | warm — in a lit fire's radius | (light column independent) |
| `WarmSource == "shelter"` | warm — shelter cover | dark — indoors (gru-safe: shelter) at night |
| `!Warm && !Night` | mild — daylight | bright — daylight |
| `!Warm && Night` | cold — night, no cover | per `Lit` below |
| `Night && Lit` | — | lit — firelight (gru-safe) |
| `Night && !Lit && !shelter` | — | dark |

Water tiles append the terrain-flavor note "open water" (from the registry row, not
a sim claim). No canopy note — no mechanic backs it (research R4).

## Hit regions

```go
type mapHitRegion struct {
    valid            bool
    originX, originY int // screen cell of tile (x0, y0)
    x0, y0           int // world coords of the top-left rendered tile
    vw, vh           int // viewport size in tiles
    // stride: 2 terminal columns per tile (the "glyph + space" render), 1 row per tile
}

type tileHitRegion struct {
    valid            bool
    originX, originY int   // screen cell of the first selectable row
    width            int   // column span of a hit
    rowIndex         []int // rowIndex[i] = tileRow index for screen row originY+i (-1 = non-selectable line)
}
```

- Recorded by `mapPanelView`/`mapView` and the TILE body renderer respectively, each
  `View()`; invalidated by default every frame (the `chronHitRegion` lifecycle,
  `tui.go:303-315`).
- Map hit → world tile: `(x0 + (msgX-originX)/2, y0 + (msgY-originY))`, rejected
  outside `[0,vw)×[0,vh)`.

## Design-reference deltas (FR-006 — same-PR doc rows, listed here so tasks can pin them)

| Page | Amendment |
|---|---|
| `panels/map.md` | deferral note re-opened (names this spec + the operator signal); control-table rows: look-cursor toggle/move/jump (`v`, `hjkl`+arrows, `H/J/K/L` · click-tile), in-mode `c` snap, cursor tile highlight; title-row cursor state; parity note updated (map gains its first mouse target) |
| `patterns/keymap.md` | "Mode: look-cursor" table (layered like inspect/villagers; dormancy rule; digits exit-and-select); `v` in the binding-selection note (mnemonic: `l` is cursor-right in-mode, "inspect" already names the chronicle mode → `v`iew); footer hints rows |
| `panels/dock.md` | the borrow seam: `TILE (x,y)` pseudo-label states, "not a tab" (no digit, no cycle membership), state-preservation rule; control-table row with click targets |
| `patterns/focus-contract.md` | scope note: the TILE pane is a drawn selection scope, not a text client — "exactly one client" unchanged (FR-004) |
| `anatomy.md` | dock region row: TILE view (transient borrow) → `panels/dock.md`; map region row: look cursor → `panels/map.md` |
| `overlays/help.md` | badge deep-link row: renderer symbol replaces `unbuilt (wave 4, layer-2)`; keys section gains the look-mode page row |
