---
title: Panel — map (terrain camera viewport)
class: panel
status: shipped
verified_against: 72f82f41f7aa2e345572105894cd0fb7c02fc0aa
sources:
  - internal/tui/views.go
  - internal/tui/tiles.go
  - internal/tui/look.go
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
(`tile()`, top wins): gru (`G`) > stranger (`S`, spec 077 — the second
nocturnal entity shares the gru's precedence tier; the gru wins a legal but
rare shared tile, being the greater threat) > agents > structures > piles >
dens > designation marks (spec 084 — beneath every real world entity, above
the path/terrain tier) > terrain. Two dynamic-overlay carve-outs exist
within that order:

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
| `S` | the stranger — a night trickster after unattended stores (`State.Stranger`, entity like the gru); appended registry row, violet 135 bold | spec 077 |
| `░` | marsh — walkable wet ground near water (new-terrain worlds, `terrain_gen: 2`) | spec 068 |
| `▒` | sand — walkable shoreline flat (new-terrain worlds, `terrain_gen: 2`) | spec 068 |
| `◇` | guardian-marked structure site (`State.Designations`, ACTIVE only — fulfilled/cancelled marks stop rendering, state-derived) | spec 084 |
| `┄` | guardian-marked wall line — one segment per enumerated line tile (`sim.DesignationTiles`, the same enumeration the fulfillment predicate checks) | spec 084 |
| `◦` | guardian-marked settlement zone — PERIMETER tiles only (interior unmarked: extent, never wallpaper) | spec 084 |
| `A`/`a`/`†` | agent by initial — uppercase awake, lowercase asleep, `†` dead; a living agent's STYLE (not case) additionally carries a condition overlay (below) | TASK-34, overlays spec 060 |

Night dimming (`m.replica.Night`): every terrain-level style gains `.Faint(true)`
— glyph identity never changes, only its brightness.

## The tile registry (spec 068)

Presentation is data, not code: every glyph above resolves through the tile
registry (`internal/tui/tiles.go`) — one table (`mapGlyphs`, grown from spec
045's shared glyph-key) carrying each tile's glyph, compact legend name,
plain-language overlay meaning, style token, style-only state variants
(fire `dying`, wall `damaged`), and the world binding it renders. `tile()`
(`renderMapGrid`, `internal/tui/views.go`) keeps ONLY the priority logic —
gru > agents > structures > piles > dens > designation marks > path >
quarried > base terrain — and resolves every leaf through the registry's
binding indexes (the stranger row appended per spec 068 FR-009 — the pinned
legend/overlay prefix stays byte-identical, marsh/sand's own precedent; the
three spec-084 designation rows appended the same way, one shared
semantic-16 `designation` token — a plan mark is meaning, not material);
the compact legend line and the `?` overlay walkthrough render from the
same rows, so a tile added to the table reaches all three surfaces with no
renderer edit.

Style tokens are classed per the tile-vocabulary analysis's palette rule:
**semantic-16** (themeable ANSI 0–15: water 4, tree 2, forage 3, den 5, the
agent family) or **material-256** (fixed palette: rock 245, spent 240, fire
208/202, shelter 130, oven 166, pile 178, chest 136, wall 250, path 137,
gru 196, grave 244, suppressed 135, marsh 65, sand 180, stranger 135
bold — spec 077). No per-tile style
literal exists outside `tiles.go` (sweep-tested); byte identity for the
pre-068 vocabulary is pinned by `TestTilesIdentityPin`'s committed goldens
(`internal/tui/testdata/tiles_identity_*.golden` — never regenerate to
"fix" a failure). Marsh/sand appear only on worlds whose manifest carries
`terrain_gen: 2` (new worlds); migrated legacy worlds regenerate their
exact pre-068 terrain and can never draw them.

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
  Clears the frame recovery crosses back over the threshold. This overlay
  also SUBSUMES the spec-083 neglect detector's map presentation by
  construction — the detector fires on the same exported band constants this
  predicate reads, so a neglect-firing villager is always already painted
  needs-critical: no new token, glyph, or legend row exists for neglect
  (`TestRenderMapGridNeglectFiringRendersCritical` pins the subsumption; the
  chronicle's whole-line alert is the event-shaped surface).
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

**Look-cursor: re-opened and shipped** (spec 074-look-cursor, TASK-142 —
re-opens the spec 060 standing resolution 2 deferral below). The deferral's
own resurfacing trigger ("re-open if playtesting shows ... the actual
bottleneck") was never what re-opened this: the operator's 2026-07-26 request
plus reorientation synthesis decision 4 ("the map cannot be interrogated")
is the signal — a distinct gap from legend-line overflow, and this feature's
research (specs/074-look-cursor/research.md) grounds it in the code as read
at re-open time, not a re-litigation of the original evaluation. `v` toggles
a cursor on the map (`hjkl`/arrows move 1, `H/J/K/L` jump 8, a 2-tile margin
pushes the camera, `c` snaps); while active the dock body is borrowed by a
transient TILE view (`panels/dock.md`) — the fourth inspection modality the
original deferral named, now justified by decision 4's gap rather than
legend-line pressure. Original deferral text, preserved for its own history:

> The feature's task asked whether a free-roaming look-cursor should exist
> alongside the growing legend line. Conclusion: chronicle jump-to-source
> (spec 049) plus the legend's own inspection line (piles/chests, above)
> already cover "what is that tile" for every shipped surface; a look-cursor
> would add a FOURTH inspection modality with its own key-mode for marginal
> gain over what these three overlays already make visible from orbit.
> Deferred — re-open if playtesting shows legend-line overflow pain becomes
> the actual bottleneck (not addressed by a look-cursor anyway, which doesn't
> touch the legend).

**Cursor rendering**: a background-highlight style transform
(`styleLookCursor`, `lipgloss.Reverse(true)`) over whatever glyph `tile()`
resolves at the cursor tile — never a new glyph (spec-068 FR-003 discipline
extended; `TestTilesIdentityPin`'s goldens are untouched because mode-off
rendering never reaches this branch). The title row gains a third state:
`MAP · cursor (x,y) · c center · esc exit` — see `patterns/keymap.md`'s new
"Mode: look-cursor" table for the full key set and the borrow's key-layering
rule, and `patterns/focus-contract.md` for the pane-focus scope note.

## Control table

| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
| terrain tile | water · wood · forage · rock · quarried · path · marsh · sand · plain | `world.Map()` + dynamic overlays | `renderMapGrid`/`tile` (tile registry, `tiles.go`) | — (display-only) | TASK-34, marsh/sand spec 068 | — |
| agent glyph | awake · asleep · dead · dead-on-grave | `replica.Agents` | `renderMapGrid` | — | TASK-34/spec 044 | — |
| structure glyph | fire lit/cold · shelter · oven · chest · wall ok/damaged · grave | `replica.Structures` | `renderMapGrid` | — | specs 012/013/032/044 | — |
| pile overlay | present | `replica.Piles` | `renderMapGrid` | — | spec 013 US2 | — |
| gru glyph | abroad | `replica.Gru` | `renderMapGrid` | — | TASK-34 | — |
| camera pan | following · panned | `Model.panX`/`panY` | `renderMapGrid` | `←↑↓→` · — | TASK-34 | — |
| camera recenter | — | `Model.panX`/`panY` reset | `mapPanelView` | `c` · — | TASK-34 | — |
| legend / inspection line | terrain-only · +piles · +chests | in-view piles/chests | `renderMapGrid` (legend half) | — (display-only) | spec 013 | — |
| condition overlay | needs-critical · suppressed-mind · dying-fire | `replica.Agents` (`Needs`) · `Model.traces` · `replica.Structures` (`FuelUntil`) | `renderMapGrid` (`needsCritical`, `agentSuppressedMind`, fire branch) | — (display-only) | reorient Wave 5, spec 060 | — |
| look-cursor toggle + click-move | off · on; cursor tile `(lookX,lookY)` | `Model.lookActive`, `lookX`/`lookY` | `handleGlobalKey` "v", `handleLookCursorKey` (`look.go`) | `v` · click a map tile (enters the mode there if inactive, moves the cursor there if already active) | spec 074 | — |
| look-cursor keyboard move/jump | cursor tile `(lookX,lookY)` | `Model.lookX`/`lookY` | `handleLookCursorKey` (`look.go`) | `hjkl`/arrows move 1 · `HJKL` jump 8 · — | spec 074 | — |
| look-cursor camera snap | following · panned-to-cursor | `Model.panX`/`panY` | `snapCameraToCursor` (`look.go`) | `c` (in-mode) · — | spec 074 | — |
| look-cursor tile highlight | highlighted · plain | `Model.lookActive`, `lookX`/`lookY` | `renderMapGrid` (`styleLookCursor` transform) | — (display-only) | spec 074 | — |

**Parity rollout**: pan (`←↑↓→`) and recenter (`c`, outside the mode) still
have no mouse target — tracked here rather than omitted (decision 8, formal
doctrine in `patterns/keymap.md`). The map's first REAL mouse target ships
with this feature: a left-click on a visible tile enters the look-cursor
mode there (or moves an already-active cursor) — `patterns/keymap.md`'s
mouse-parity note and `panels/dock.md`'s TILE-pane click target are the
other two legs of the same click-tile/click-row parity landing (US4,
decision 8 rule 1: keyboard and mouse land together).
