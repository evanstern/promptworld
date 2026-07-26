---
title: Panel — map (terrain camera viewport)
class: panel
status: shipped
verified_against: 7e3c2b5f5f23eb8e5fcb37d0f867dbc6f46a289b
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
- **Fire lit/dying/cold** (spec 012; dying added spec 060 US2): lit while
  the current tick is before the structure's `FuelUntil` (`▲`, `styleFire`);
  **dying** — a third, still-lit state — once the remaining fuel drops
  inside `State.RefuelDyingBelow()`'s window (spec 057; the SAME window the
  reflex's own refuel-before-cold rule keys on, so this never invents a
  second threshold): still `▲`, but the warn-toned `styleFireDying` instead
  of plain `styleFire` — steady, no blink (standing resolution 3); cold once
  fuel actually runs out (`△`, faint `styleFireCold`).

| Glyph | Meaning | Introduced-by |
|---|---|---|
| `~` `♠` `"` `^` | water · wood · forage · rock outcrop (static terrain) | TASK-34 |
| `,` | depleted (quarried) rock outcrop — dynamic overlay | spec 012 |
| `ᴥ` | a gru's den (static) | TASK-34 |
| `▲` / `▲` (warn) / `△` | fire, lit / dying / cold | spec 012, dying spec 060 |
| `⌂` `▣` `☐` | shelter · oven · chest | spec 013 US3 |
| `%` | ground pile (dropped goods) — dynamic overlay, own priority layer | spec 013 US2 |
| `▤` / `▩` | wall, plank / stone; dim (`styleWallDamaged`) below `sim.WallMaxHP` | spec 032 |
| `·` (tan) | a paved path — terrain-level, distinct from plain ground's dim `·` | spec 032 US3 |
| `✝` | a grave (dead-agent-on-grave carve-out above) | spec 044 US4 |
| `G` | the gru — highest render priority | TASK-34 |
| `A`/`a`/`†` | agent by initial — uppercase awake, lowercase asleep, `†` dead; a living agent's STYLE (not case) additionally carries a condition overlay (below) | TASK-34, overlays spec 060 |

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

**Condition-overlay naming** (spec 060 US2 AS5, FR-003): the legend also
carries a prose note (`conditionOverlayNote`, `internal/tui/help.go`) naming
the needs-critical/suppressed-mind/dying-fire marker styles above — the
same `agentGlyphNote`/`mapControlNote` precedent (a note, not a new
`mapGlyphs` row, since every overlay is a style variant of an
already-legended glyph). The help overlay's glyph walkthrough page renders
the identical note (FR-005 anti-drift discipline).

## Behavior

- Title row states the camera mode: `following centroid` or `panned (c to
  recenter)`.
- Arrow keys pan whenever the minibuffer is unfocused — from home,
  regardless of selected dock tab. `c` recenters and resumes following.
- Agents render as name-initial glyphs (existing behavior).

## Wave 5 — condition overlays (shipped, spec 060/TASK-129)

Village-lens completion adds three glanceable condition overlays, layered on
top of the existing agent/fire glyphs (`renderMapGrid`'s agents loop and
fire branch, `internal/tui/views.go`) — style variants of already-legended
glyphs, not new glyphs:

- **Needs-critical** (`styleAgentCritical`, bold+underlined red) — a living
  villager with any need at its existing danger-band threshold: the SAME
  thresholds the reflex's own PREP-gate/survival rungs already treat as "in
  danger" (`needsCritical`, `internal/tui/views.go`) — `sim.
  SurvivalNearDeathBelow` (health), `sim.SurvivalStarvingRearm` (food),
  `sim.SurvivalFreezingRearm` (warmth), and the new `sim.DangerRestBelow`
  export (rest — the one need with no prior exported threshold). Morale has
  no danger band in sim today, so it is deliberately not part of this check.
  Clears the frame recovery crosses back over the threshold.
- **Suppressed-mind** (`styleAgentSuppressed`, faint magenta) — a living
  villager whose latest decision-trace chain (`Model.traces`,
  `internal/tui/decisions.go`) is a router suppression (`agentSuppressedMind`)
  — the map form of spec 037's "a skipped thought is visible." Clears the
  moment a later, non-suppressed outcome lands for that agent.
- **Dying fire** (`styleFireDying`, warn orange, steady) — see the glyph
  inventory above; still lit (`▲`), not a new glyph.

**Priority** (needs-critical > suppressed-mind, physical danger over
cognitive telemetry): when both hold for the same living agent, only the
needs-critical style renders — never blended, never suppressed-mind. A dead
agent never carries either overlay (no needs, no mind to suppress).

**No blink, ever** (standing resolution 3): "pulse" is a style, never
terminal blink — accessibility + terminal-variance doctrine, identical to
the existing damaged-wall/cold-fire faded-treatment precedent.

**Look-cursor: evaluated and deferred** (spec 060 standing resolution 2).
The feature's task asked whether a free-roaming look-cursor should exist
alongside the growing legend line. Conclusion: chronicle jump-to-source
(spec 049) plus the legend's own inspection line (piles/chests, above)
already cover "what is that tile" for every shipped surface; a look-cursor
would add a FOURTH inspection modality with its own key-mode for marginal
gain over what these three overlays already make visible from orbit.
Deferred — re-open if playtesting shows legend-line overflow pain becomes
the actual bottleneck (not addressed by a look-cursor anyway, which doesn't
touch the legend).

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
| condition overlay | needs-critical · suppressed-mind · dying-fire | `replica.Agents` (`Needs`) · `Model.traces` · `replica.Structures` (`FuelUntil`) | `renderMapGrid` (`needsCritical`, `agentSuppressedMind`, fire branch) | — (display-only) | reorient Wave 5, spec 060 | — |

**Parity rollout**: pan (`←↑↓→`) and recenter (`c`) have no mouse target
today; tracked here rather than omitted (decision 8, formal doctrine in
`patterns/keymap.md`, T024).
