# Research — TUI Look-Cursor Mode (spec 074)

Phase 0 findings. The board design is operator-reviewed and DECIDED (bindings, layout,
borrow seam); every decision below is an *implementation-approach* resolution grounded
in the code as read 2026-07-26 (root main = 93cb578), never a re-litigation of the
design. File:line references are the evidence trail.

## R1 — Mode layering: where the look layer sits in `handleKey`

`handleKey` (`internal/tui/tui.go:914-973`) dispatches in priority order: `ctrl+c` →
takeover → help overlay → exercise-briefing eater → `?` → minibuffer focus → `p`
(ended-only) → console → inspect → villagers → global.

- **Decision**: the look layer slots between the console check and the inspect check —
  a `m.lookActive` gate calling `handleLookKey(msg) (Model, tea.Cmd, bool)`, the exact
  `handleInspectKey`/`handleVillagersKey` handled-bool shape. Unclaimed keys fall
  through to `handleGlobalKey` (so `space`, `[`/`]`, `m`, `q` keep working, FR-013).
- **Dormancy of inspect/villagers during the borrow**: `inspecting()` (`tui.go:876`)
  gates on `chronicleVisible()` and `villagersVisible()` (`tui.go:816`) on
  `dockTab`/`active` — both would spuriously read "visible" while the TILE view has
  borrowed the body. Rather than patch every predicate, the look layer sits ABOVE
  them and claims the whole contested key set (`j/k/g/G/J/K/⏎/esc/d/a/t/r` are all
  either look-claimed or safe); additionally `chronicleVisible()`/
  `villagersVisible()`/`exerciseVisible()` each gain a `!m.lookActive &&` guard so
  render-side and mouse-side consumers (`handleMouse`'s `!m.inspecting()` guard,
  `currentHelpMode`, reply-badging via `guardianVisible` — which must badge, not
  stream, during a borrow) agree that the TILE view is the thing visible. This is the
  villagers-mode scoping precedent stated on the board ("the TILE view is what's
  visible in the dock").
- **`v` entry point**: a new `case "v"` in `handleGlobalKey`, gated on the map being
  visible (widescreen home — not solo, not console — or narrow with `active ==
  paneMap`) and `m.gameMap != nil`; strict documented no-op otherwise (the `x`/`6`
  precedent). Entry from inspect/villagers mode works because neither layer claims
  `v` (verified against `handleInspectKey`/`handleVillagersKey` key sets), so `v`
  falls through to global.
- **Alternatives considered**: patching only the predicates (no layer) — rejected:
  the mode needs its own key table anyway; a second dispatch home would split the esc
  chain across files.

## R2 — Borrow seam: transient body, untouched tab state

`dockTabsRow` (`views.go:988-1023`) renders labels from a static list keyed off
`m.dockTab`; `dockTabContent` (`views.go:1027-1048`) dispatches the body by
`m.dockTab`. Per-tab state (chronicle selection/scroll, villager selection, etc.)
lives on Model fields keyed by tab, never torn down on switch (dock.md: "a tab never
re-mounts on selection").

- **Decision**: the borrow NEVER writes `m.dockTab`. While `m.lookActive`:
  `dockTabsRow` appends/replaces the active-highlight with a highlighted
  `TILE (x,y)` pseudo-label (normal tab labels render inactive-dim), and
  `dockTabContent` short-circuits to `tileBody(width, height)` before the `dockTab`
  switch. AC #2's "prior state intact" then holds *by construction* — nothing about
  the underlying tab changed, exactly how the guardian console leaves
  `dockTab/solo/active` untouched as its own return target (`openConsole`,
  `tui.go:1346-1353`).
- **`2`–`6` exit-and-select**: the look layer claims the digits, clears the mode
  state, and delegates to the existing `selectTab` — one behavior, no duplication.
  Note the solo-zoom grammar: `selectTab` on the already-selected tab zooms solo;
  exiting the mode first means "press `2` while borrowing over chronicle" selects
  (not zooms) only if we route digits as *plain tab selection*. Decision: the look
  layer maps digits to `selectTab(k)` after clearing mode state — if `k` was already
  the selected tab this yields solo zoom, which is acceptable and honest (`selectTab`
  is the single grammar; the board text "exit the mode and select that tab" is
  satisfied; documenting the same-tab-solo consequence lands in keymap.md's mode
  table).
- **Covered-by-body-replacing-surfaces rule** (spec edge case): `G` (console) and
  solo zoom exit the mode before opening; help and takeovers layer above without
  ending it (they restore what was beneath by construction — `closeHelp`
  `tui.go:1032`, `handleTakeoverKey` `tui.go:1099`).

## R3 — Cursor/camera math: one shared origin helper

`renderMapGrid` (`views.go:1114-1139`) computes the camera origin inline:
`cx,cy = wandererCentroid() + pan`, then `x0 = clampInt(cx-vw/2, 0, W-vw)` (same for
y). The key handler needs the same numbers to push the camera and clamp the cursor,
and the mouse handler needs them to hit-test — three consumers.

- **Decision**: extract `cameraOrigin(vw, vh) (x0, y0)` on Model (pure over
  centroid+pan+clamps) — the `wandererCentroid` extraction precedent from spec 049
  (research R1 there), which exists precisely so camera writers and the renderer
  can't drift. `renderMapGrid` calls it; the look layer and mouse hit-testing call it
  with the current viewport dims recomputed from `computeColumns`/`computeRows`/
  `mapViewportTiles` (`layout.go` — all pure, so the handler needs no cached
  geometry). A tiny `mapViewportDims()` Model helper wraps that recomputation for
  both consumers (widescreen home vs narrow map pane use different formulas —
  `mapPanelView` vs `mapView`).
- **Camera push**: after a cursor move, if the cursor is within 2 tiles of the
  viewport edge, adjust `panX/panY` by the overshoot so the cursor sits exactly at
  the margin — clamped by the same origin clamps (at the world edge the camera
  stops; the cursor may reach the viewport border, spec edge case).
- **`c` in-mode**: snaps the camera onto the cursor — `pan = cursor −
  wandererCentroid()` (the `centerCameraOn` jump-to-source formula, one more caller
  of the same math). Outside the mode `c` keeps its recenter meaning untouched.
- **Exit**: `panX, panY = 0, 0` — "resumes centroid-following" is literally the
  existing following state (`mapPanelView` title logic, `views.go:914-917`).
- **Cursor rendering**: the cursor tile renders with a background-highlight style
  transform over whatever glyph `tile()` resolves (the `feedSelect`
  background-highlight precedent, layout.md style tokens) — a style transform, never
  a glyph change (spec-068 FR-003 discipline), applied in `renderMapGrid` when
  `lookActive` and `(x,y) == cursor`. Mode-off rendering is untouched → the
  `TestTilesIdentityPin` goldens keep passing (SC-007).

## R4 — Warmth + light derivation (FR-007): export one pure sim helper

The mechanics exist, unexported:

- `warmAt(s, x, y, tick)` (`internal/sim/terrain.go:140`) — within `fireWarmRadius`
  of a LIT fire (`tick < FuelUntil`), or exactly on a shelter tile.
- `litAt(s, x, y)` (`internal/sim/gru.go:70`) — within `gruLightRadius` of a fire
  (deliberately wider than warmth; note: litAt does not check `FuelUntil` today —
  the helper must present what the mechanic actually does, not an idealization).
- Day/night: `s.Night` (`state.go:37`); `decayNeeds` (`executor.go:1472-1504`) keys
  warmth gain/loss on exactly (warm, night).

- **Decision**: add `internal/sim/env.go` with an exported, pure, read-only
  `EnvAt(s *State, x, y int, tick int64) EnvSample` where `EnvSample` carries
  `Warm bool` + `WarmSource` (`fire` / `shelter` / none) and `Lit bool` +
  `LightSource` (`daylight` / `firelight` / none) — decomposed from the same loops
  `warmAt`/`litAt` run (shared private cores so the predicates and the sample can
  never disagree; `warmAt`/`litAt` become thin wrappers). No reducer changes, no new
  constants, no persistence. SC-006's test asserts `EnvAt(...).Warm == warmAt(...)`
  and `EnvAt(...).Lit == litAt(...)` across a structure/tick matrix.
- **Level/note vocabulary (TUI-side mapping, plain language per FR-020)**:
  - Warmth: `Warm/fire` → "warm — in a lit fire's radius"; `Warm/shelter` → "warm —
    shelter cover"; else day → "mild — daylight"; else night → "cold — night, no
    cover"; a water tile appends "open water" as terrain flavor (from the registry
    row, not a sim claim).
  - Light: day → "bright — daylight"; night+`Lit` → "lit — firelight (gru-safe)";
    night+shelter → "dark — indoors (gru-safe: shelter)"; else night → "dark".
    The gru-safety notes are honest restatements of `gruProtected`
    (`gru.go:80-83` — light OR shelter), the teaching payoff of exposing light at
    all.
  - Meter: a 3-step discrete gauge (e.g. `▮▮▯`) per level — presentation detail,
    finalized in the plan; the LEVELS are the contract, the glyphs are not.
  - The board AC's "canopy" example has no sim mechanic behind it today (no
    tree-cover effect exists in `decayNeeds`/`litAt`); the note vocabulary therefore
    omits it rather than invent an unbacked claim — the AC's list is illustrative
    ("fire radius, shelter cover, open water; daylight, canopy, indoors, firelight"),
    and honesty-doctrine (merged position 2) forbids decorating with mechanics that
    don't exist. If tree cover ever becomes a mechanic, the note gains it then.

## R5 — "Recent events at that tile": recorded positions via the subject registry

`resolveSubject` (`internal/tui/digest.go:1889-1912`) resolves a *jump target*
(live actor position preferred). For "what happened HERE" the live-actor preference
is wrong — an event belongs to the tile where it was recorded.

- **Decision**: a `tileEvents(x, y) []int` helper over `m.events` (the existing
  client ring) that decodes each event through the same `subjectRegistry` and keeps
  it iff its candidate carries explicit payload coordinates (`hasPos`) equal to
  (x, y). Bounded work, registry-bounded types, never decodes unknown payloads —
  `resolveSubject`'s own posture. Most recent first, capped by the pane budget.
  Events with actor-only candidates don't list (their historical tile is simply not
  recorded); this is stated in the spec's assumptions rather than silently implied.
- Drill-in reuses `formatInspector`/`chronicleDetailPane`'s raw-JSON rendering
  (`views.go:1861-1907`) — the FR-020 "raw behind an explicit inspector drill"
  boundary, one renderer family, no fork.

## R6 — Mouse targets: two new hit regions, the `chronHitRegion` pattern

`chronHitRegion` (`tui.go:303-315`) is the proven shape: a pointer field on Model so
value-receiver `View()` can record geometry each frame; `handleMouse`
(`tui.go:1688-1711`) consumes it next Update; invalidated whenever the surface wasn't
rendered. `tea.WithMouseCellMotion` is already on (`cmd/promptworld/commands.go:856`).

- **Decision**: two more per-frame regions — `mapHit` (grid origin cell, `x0/y0`
  world offset, `vw/vh`, 2-terminal-columns-per-tile stride — recorded by
  `mapPanelView`/`mapView`) and `tileHit` (row → band-item mapping, recorded by the
  TILE body renderer). `handleMouse` routes in order: existing guards (release+left,
  help/minibuffer no-ops) → TILE pane region (select; second click on selected row
  drills) → map region (move cursor / enter mode) → existing chronicle path. The
  chronicle path is unreachable during a borrow anyway (R1's visibility guard makes
  `inspecting()` false), so no ambiguity.
- Narrow: `mapView` records `mapHit` the same way; the TILE body records `tileHit`
  wherever it renders. One mechanism, both widths.

## R7 — Narrow fallback ruling (FR-012)

The design mock is widescreen; narrow is unaddressed by the board text but governed
by corpus doctrine: keyboard reaches 100% of functionality (keymap.md doctrine 2),
narrow "is today's single-pane UI, never deleted" (layout.md ruling b), and narrow
terminals plausibly host new players (layout.md's own rationale).

- **Decision**: the mode ships in narrow, scoped to the map pane being active: `v`
  raises the cursor on the map pane; `⏎`/`tab` swaps the pane body to the TILE view
  (transient body replacement — the "one component, two widths" posture); `esc`
  unwinds the same chain. No new narrow chrome, no fold-cascade participation.
- **Alternative considered**: widescreen-only v1 (the villager-strip "NOT carried"
  precedent) — rejected: the strip is glanceability beside a map narrow doesn't
  render; tile interrogation is core functionality with nothing else covering it,
  so excluding narrow would be a keyboard-reachability gap, not a chrome judgment.

## R8 — Badge deep-link mechanics (FR-011)

`openHelp` (`tui.go:1018-1027`) always opens `helpSectionKeys`, scroll 0. The header's
conditional badges are `[degraded]`, `[llm: …]`, `[suppressed: …]` (`headerView`,
anatomy.md header region), each derivable at open time from existing Model state
(`m.status`/`m.llm`/horizon fields — the same predicates `headerView` uses).

- **Decision**: `openHelp` gains a pre-focus step: if ≥1 conditional badge is active,
  set `helpSection = helpSectionScreen` and `helpScroll` to the first active badge's
  `headerAnatomy` row index (resolved from the same shared table
  `helpWalkthroughLines` renders — one source, so the scroll target can't drift from
  the rendered row). Row highlight uses the existing selected-row styling if cheap,
  else "scrolled into view at top" satisfies "pre-focused" (plan finalizes; the
  contract is the row is visible on open). No badge → behavior byte-identical
  (`TestHelp*` byte pins keep passing).
- `overlays/help.md` already classifies this row "status-derived (which row is
  pre-focused …); content unchanged" — the doc flip is control-table `unbuilt
  (wave 4, layer-2)` → shipped renderer symbol, same-PR (FR-006).

## R9 — Help keys section: a seventh mode page (FR-014)

`currentHelpMode` (`tui.go:984-1011`) must return something honest when `?` is
pressed mid-mode (help freezes the opened-from mode). The keys section pages through
per-mode tables (`helpKeysLines`, `nextHelpMode`/`prevHelpMode`).

- **Decision**: add `helpModeLook` — checked in `currentHelpMode` right after the
  console branch (the mode borrows the dock, so the villagers/inspect branches below
  would mis-route exactly the way the console comment warns). Its basic-tier rows:
  `v` toggle, `hjkl/arrows` move, `H/J/K/L` jump, `c` center, `⏎/tab` focus pane,
  `j/k ⏎` select/drill, `2-6` exit to tab, `esc` back one layer. Footer hints gain a
  look-mode line (keymap.md "Footer hints per mode" grows one row; two variants —
  cursor vs pane-focused — if the hint line budget allows, else one combined line;
  plan finalizes).

## R10 — What the fixed-geometry AC pins (FR-008 / SC-003)

`mapPanelView`/`dockPanelView` (`views.go:910-953`) render to exactly the handed
(cols, rows) — layout.md's composition contract (B1). The mode adds no chrome rows
and never touches `computeColumns`/`computeRows`.

- **Decision**: SC-003's test renders the full composite at a fixed size in four
  states (mode off / cursor / pane focus / drill-in) and asserts per-panel
  `lipgloss.Width`/`Height` equality — pinning the AC as a test rather than a
  review-time claim (gates-over-assertions).
