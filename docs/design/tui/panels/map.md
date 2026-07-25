---
title: Panel — map (terrain camera viewport)
class: panel
status: shipped
verified_against: c8906da39be3a5b861c2272af37db0a83dcded7a
sources:
  - internal/tui/views.go
---

# Panel: map

The terrain camera viewport. **Rendering is unchanged** from the current
`renderMapGrid`/`mapView`/`mapPanelView` (`internal/tui/views.go`) — glyph
grid, 2 terminal columns per tile, camera following the wanderer centroid,
night dimming. This doc only respecifies its sizing, its place in the
composite, and (this feature) reconciles its glyph/overlay inventory against
everything shipped since TASK-34.

## Mockup (in-composite)

```
┌─ MAP · following centroid ────────────────┐
│ ~ ~ ~ ~ " " ♠ ♠ ♠ ♠ ♠ " " . . . . ▲ . .   │
│ ~ ~ ~ " " ♠ ♠ A ♠ ♠ " " . . ⌂ ⌂ . . . .   │
│ ~ ~ " ♠ ♠ ♠ R ♠ " " . . . ⌂ . B . . .     │
│ ~ . . ᴥ . " " . . . . S . . . " " . .     │
│                                           │
│ ~ water ♠ wood " forage ᴥ den ▲ fire ⌂ ⌂  │
└───────────────────────────────────────────┘
```

## Sizing

- Composite mode: map gets **all columns left of the dock** (see
  [../patterns/layout.md](../patterns/layout.md)); viewport tiles =
  `(mapCols - borders) / 2` wide, `bodyRows - legend` tall — same formula
  family as today's `vw/vh` computation, with the new width input.
- Solo/fallback mode: full terminal width, as today.
- Viewport stays a camera window clamped to world size — never letterboxed,
  never scaled.

## Glyph inventory (reconciled against shipped reality)

Every glyph `renderMapGrid` can draw, and the priority a tile resolves in
(`tile()`, top wins): gru (`G`) > agents > structures > piles > dens >
terrain. Two dynamic-overlay carve-outs exist within that order:

- **Dead-agent-on-grave** (spec 044 US4): a dead agent standing on a tile
  that also holds a grave renders the grave glyph (`✝`, `styleGrave`)
  instead of the plain dead marker (`†`) — every post-044 death places its
  grave at the dead agent's own frozen tile, so the usual agent-over-
  structure priority would otherwise permanently hide it. A graveless dead
  agent (pre-044 replay/history) still renders plain `†`.
- **Fire lit/cold** (spec 012): lit while the current tick is before the
  structure's `FuelUntil` (`▲`, `styleFire`); cold once fuel runs out (`△`,
  faint `styleFireCold`).

| Glyph | Meaning | Introduced-by |
|---|---|---|
| `~` `♠` `"` `^` | water · wood · forage · rock outcrop (static terrain) | TASK-34 |
| `,` | depleted (quarried) rock outcrop — dynamic overlay | spec 012 |
| `ᴥ` | a gru's den (static) | TASK-34 |
| `▲` / `△` | fire, lit / cold | spec 012 |
| `⌂` `▣` `☐` | shelter · oven · chest | spec 013 US3 |
| `%` | ground pile (dropped goods) — dynamic overlay, own priority layer | spec 013 US2 |
| `▤` / `▩` | wall, plank / stone; dim (`styleWallDamaged`) below `sim.WallMaxHP` | spec 032 |
| `·` (tan) | a paved path — terrain-level, distinct from plain ground's dim `·` | spec 032 US3 |
| `✝` | a grave (dead-agent-on-grave carve-out above) | spec 044 US4 |
| `G` | the gru — highest render priority | TASK-34 |
| `A`/`a`/`†` | agent by initial — uppercase awake, lowercase asleep, `†` dead | TASK-34 |

Night dimming (`m.replica.Night`): every terrain-level style gains `.Faint(true)`
— glyph identity never changes, only its brightness.

## Inspection (map legend, spec 013 T021/T026)

The legend — the map's one designated inspection surface, content grows the
line rather than adding a second row — appends, for whatever's currently in
view, a stockpile-zone summary per pile cluster (`pileZones`, 4-neighbor
Manhattan adjacency, a render-side-only flood fill) and an owner+contents+
fullness entry per chest (`describeChest`). Legend stays pinned as the
panel's last row; drop the legend before shrinking the viewport when rows
get scarce (this page's own shed order, distinct from `patterns/layout.md`'s
chrome-row fold order, which sheds this legend *first* system-wide).

## Behavior

- Title row states the camera mode: `following centroid` or `panned (c to
  recenter)`.
- Arrow keys pan whenever the minibuffer is unfocused — from home,
  regardless of selected dock tab. `c` recenters and resumes following.
- Agents render as name-initial glyphs (existing behavior).

## Wave 5 — condition overlays (specified stub)

Village-lens completion (reorientation Wave 5) adds map condition overlays —
glanceable at-a-glance state layered on the terrain (e.g. a danger/threat
tint, a resource-scarcity tint). **Not built**; this page reserves the row
below as the spec-before-build stub until Wave 5 authors the real behavior.
No renderer exists yet; nothing on screen changes because of this row.

## Control table

| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
| terrain tile | water · wood · forage · rock · quarried · path · plain | `world.Map()` + dynamic overlays | `renderMapGrid`/`tile` | — (display-only) | TASK-34 | — |
| agent glyph | awake · asleep · dead · dead-on-grave | `replica.Agents` | `renderMapGrid` | — | TASK-34/spec 044 | — |
| structure glyph | fire lit/cold · shelter · oven · chest · wall ok/damaged · grave | `replica.Structures` | `renderMapGrid` | — | specs 012/013/032/044 | — |
| pile overlay | present | `replica.Piles` | `renderMapGrid` | — | spec 013 US2 | — |
| gru glyph | abroad | `replica.Gru` | `renderMapGrid` | — | TASK-34 | — |
| camera pan | following · panned | `Model.panX`/`panY` | `renderMapGrid` | `←↑↓→` · — | TASK-34 | — |
| camera recenter | — | `Model.panX`/`panY` reset | `mapPanelView` | `c` · — | TASK-34 | — |
| legend / inspection line | terrain-only · +piles · +chests | in-view piles/chests | `renderMapGrid` (legend half) | — (display-only) | spec 013 | — |
| condition overlay (stub) | unbuilt | — | `unbuilt (wave 5)` | — | reorient Wave 5 | — |

**Parity rollout**: pan (`←↑↓→`) and recenter (`c`) have no mouse target
today; tracked here rather than omitted (decision 8, formal doctrine in
`patterns/keymap.md`, T024).
