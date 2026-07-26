# Feature Specification: TUI Look-Cursor Mode — Tile Inspection with a Focusable Tile Pane

**Feature Branch**: `task-142-look-cursor`

**Created**: 2026-07-26

**Status**: Draft

**Input**: Board task TASK-142 (operator-reviewed design, mock at
https://claude.ai/code/artifact/d998d6d2-bb8f-4d2c-9689-cde5b7de1961, revision
"fixed-geometry-env-levels") + reorientation synthesis decision 4 / merged position 4
(`docs/design/reorient-2026-07-26-ui.md`): map interrogation completes the inspector
chain — Smallville's inspector chain and DF's `k` cursor both run through the map;
promptworld's runs only through lists.

**Re-opens a recorded deferral.** `docs/design/tui/panels/map.md` ("Look-cursor:
evaluated and deferred", spec 060 standing resolution 2) parked exactly this feature
pending a playtesting signal. Operator request 2026-07-26 is that signal; the map.md
resolution note is amended in this feature's same PR (FR-006). The design on the board
task is DECIDED (operator-reviewed): bindings, layout, and the borrow seam below are
requirements, not options.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Point at any tile from the keyboard (Priority: P1)

As a player watching the village, I press `v` and get a look-cursor on the map: a
highlighted tile I can move with `hjkl` or the arrow keys (`H/J/K/L` jump 8 tiles),
spawned at the camera center. The map title tells me where I am
(`MAP · cursor (x,y) · c center · esc exit`). When I push the cursor toward the
viewport edge the camera follows it at a 2-tile margin; `c` snaps the camera onto the
cursor; `esc` (or `v` again) exits and the camera resumes following the villager
centroid.

**Why this priority**: the cursor is the substrate every other story stands on — no
cursor, no tile to inspect. It is independently valuable even before the TILE view
exists: "which exact tile is that fire on" is answerable from the title's coordinates.

**Independent Test**: enter the mode on a connected world, walk the cursor to every map
edge and back, watch the camera track at the margin, snap with `c`, exit with `esc`,
and confirm the camera resumes centroid-following — all without the dock changing size.

**Acceptance Scenarios**:

1. **Given** the widescreen home composite, **When** the player presses `v`, **Then**
   the cursor appears at the camera-center tile and the map title reads
   `MAP · cursor (x,y) · c center · esc exit`.
2. **Given** the mode is active, **When** the player presses `l` (or right-arrow),
   **Then** the cursor moves one tile right; **When** the player presses `L`, **Then**
   it moves 8 tiles right, clamped to the world edge.
3. **Given** the cursor 2 tiles from the viewport's right edge, **When** it moves
   right again, **Then** the camera pans right so the cursor stays at least 2 tiles
   inside the viewport (until the world edge stops the camera).
4. **Given** the mode is active with the camera pushed away from the centroid,
   **When** the player presses `esc` once (with nothing focused, no drill-in),
   **Then** the mode exits and the camera resumes following the centroid.
5. **Given** the mode is active, **When** the player presses arrow keys, **Then** they
   move the cursor — never the free camera-pan they drive outside the mode.

---

### User Story 2 - Read what a tile is and what's on it (Priority: P1)

While the cursor is up, the dock body is borrowed by a transient **TILE view** (the
solo-zoom-seam style — not a numbered tab): the tab row shows a highlighted
`TILE (x,y)` pseudo-label, and the body lists the tile in DF's fixed hierarchy —
agents (with needs and current intent) → piles/chests (contents) → structures →
terrain — with a header carrying the tile's plain-language "whatis" prose from the
tile registry's `meaning` rows plus warmth and light levels (meter + plain-language
note). Pressing `2`–`5` (`6` on scenario worlds) exits the mode and restores that tab
with its prior state intact — chronicle selection included.

**Why this priority**: this is the inspection payoff — decision 4's "the map cannot be
interrogated" gap. It makes the look-cursor the third in-place lookup after `?` and the
guardian's explain tool, shrinking what external docs must carry.

**Independent Test**: park the cursor on a busy tile (agent + pile + structure), read
the TILE view top to bottom, confirm the fixed scan order, the registry-sourced whatis
line, and the warmth/light header; press `2` and confirm the chronicle returns with
its previous selection.

**Acceptance Scenarios**:

1. **Given** the mode is active, **When** the dock renders, **Then** its body is the
   TILE view for the cursor tile and the tab row shows a highlighted `TILE (x,y)`
   pseudo-label in place of the normal active-tab highlight.
2. **Given** a tile holding an agent, a pile, and a shelter, **When** the TILE view
   renders, **Then** rows appear in exactly the order agents → piles/chests →
   structures → terrain (stable scan order, every tile, every time).
3. **Given** any tile, **When** the TILE view header renders, **Then** it shows the
   tile's registry `meaning` prose plus a warmth level and a light level, each a
   discrete meter with a plain-language source note (fire radius, shelter cover, open
   water; daylight, indoors, firelight, dark).
4. **Given** the chronicle tab was selected with an event selected before `v`,
   **When** the player presses `2` from the mode, **Then** the mode exits, the
   chronicle tab is active, and the previous selection/scroll state is intact.
5. **Given** an empty grass tile, **When** the TILE view renders, **Then** it shows
   the terrain row and whatis prose only — no blank sections, no placeholder noise.

---

### User Story 3 - Dig into what the tile holds (Priority: P2)

From the cursor, `⏎` (or `tab`) moves focus into the tile pane — the pane border turns
amber (focus is drawn, focus-contract rule 2). `j/k` select rows; `⏎` drills in: an
agent row opens the villager-detail renderer family for that villager; an event row
opens the raw-JSON debug inspector (the FR-020 boundary: plain language by default,
raw behind the explicit inspector drill); a chest or pile row opens its contents.
`esc` releases exactly one layer per press: drill-in → pane rows → map cursor → exit
mode → (global esc behavior).

**Why this priority**: completes the inspector chain — from a map tile to the same
deep surfaces the lists already reach — but the cursor and TILE view are useful
without it.

**Independent Test**: focus the pane, select an agent row, drill into the villager
detail, then unwind with four `esc` presses back to centroid-following, counting
exactly one visible layer released per press.

**Acceptance Scenarios**:

1. **Given** the mode is active with cursor focus, **When** the player presses `⏎`
   (or `tab`), **Then** the tile pane border renders amber and `j/k` move a visible
   row selection.
2. **Given** the pane is focused on an agent row, **When** the player presses `⏎`,
   **Then** the pane body shows that villager's detail (identity/needs/intent —
   the villager-detail renderer family), still inside the same pane geometry.
3. **Given** a drill-in is open, **When** the player presses `esc` four times,
   **Then** the layers release one per press: drill-in → pane rows (focus back to
   cursor) → cursor (mode exits) → and the fourth press performs the normal global
   `esc` (nothing left from this feature to release).
4. **Given** the pane is focused, **When** the player presses any key, **Then** no
   key is silently swallowed without a visible effect, and no text is captured — the
   minibuffer remains the focus contract's only text-input client.

---

### User Story 4 - Mouse parity ships with the keyboard (Priority: P2)

Clicking a map tile moves the cursor there — entering the mode if it was inactive.
Clicking a row in the TILE pane selects it; clicking the selected row again (or a
row's drill affordance) drills in. This is the map's first real mouse target
(input-parity doctrine, decision 8 rule 1: keyboard and mouse land together).

**Why this priority**: decision 8 rule 1 makes parity a ship-together requirement for
new controls, not a follow-up; it rides the same hit-region machinery the chronicle
click already proved.

**Independent Test**: with the mode inactive, click a visible tile and confirm the
mode enters with the cursor there; click a TILE pane row and confirm selection; click
it again and confirm the drill-in.

**Acceptance Scenarios**:

1. **Given** the home composite with the mode inactive, **When** the player
   left-clicks inside the map grid, **Then** the mode enters with the cursor on the
   clicked tile.
2. **Given** the mode active, **When** the player clicks a different tile, **Then**
   the cursor moves there (no mode churn).
3. **Given** the TILE pane visible, **When** the player clicks one of its rows,
   **Then** that row is selected (acquiring pane focus if the cursor held it); a
   second click on the selected row drills in.
4. **Given** the help overlay open or the minibuffer focused, **When** the player
   clicks the map, **Then** nothing happens (the same no-op guards the chronicle
   click already honors).

---

### User Story 5 - Help opens on the badge that's on screen (Priority: P3)

When a conditional header badge is active (`[degraded]`, `[llm: …]`,
`[suppressed: …]`), pressing `?` opens the help overlay pre-focused on that badge's
row in the screen-walkthrough section — the retained layer-2 row from
`overlays/help.md`, folded into this lane by reorientation decision 4.

**Why this priority**: a cheap, specified addition that closes the docs branch's one
unshipped item; fully independent of the cursor stories.

**Independent Test**: force an active `[llm: …]` badge, press `?`, and confirm the
overlay opens on the screen section scrolled to that badge's `headerAnatomy` row;
with no badge active, confirm `?` opens on the keys section exactly as today.

**Acceptance Scenarios**:

1. **Given** a header badge is active, **When** the player presses `?`, **Then** the
   overlay opens on the screen-walkthrough section with the badge's row visible
   (pre-focused).
2. **Given** no conditional badge is active, **When** the player presses `?`,
   **Then** the overlay opens byte-identically to today (keys section, basic tier).
3. **Given** the overlay was opened pre-focused, **When** the player pages/cycles
   sections, **Then** all existing help navigation behaves unchanged.

---

### Edge Cases

- **Cursor at the map edge**: the cursor clamps to `[0, W) × [0, H)`; the camera
  clamps to the world exactly as `renderMapGrid` does today, so near the edge the
  2-tile push margin degrades gracefully (the camera stops, the cursor may reach the
  viewport border). A shift-jump (`H/J/K/L`) that would overshoot clamps to the edge,
  never wraps.
- **Mode entry while chronicle-inspect is open**: `v` while paused with the chronicle
  selected borrows the dock like any other tab — the inspect layer goes dormant (its
  `j/k` never fire during the borrow; the TILE view is the thing visible, the
  villagers-mode scoping precedent), chronicle selection and detail scroll are
  preserved, and exiting the mode (or pressing `2`) restores inspect mode exactly as
  it was.
- **Narrow-layout fold cascade interaction**: the mode adds zero chrome rows and
  never participates in `computeRows`' fold cascade; fold pressure while the mode is
  active resizes the body exactly as it does today, and the cursor's camera-push math
  reads the current viewport each frame, so a fold mid-mode changes the margin
  geometry without crashing or resizing panels beyond what the fold itself did. In
  the narrow fallback the mode is available while the map pane is the active pane;
  the TILE view renders as a transient body replacement of that single pane (see
  FR-012).
- **Dead/empty tiles**: a dead agent on the tile lists in the agents band with its
  dead status (drilling in shows the frozen detail); a grave lists under structures;
  a truly empty tile renders header + terrain row only. No band renders an empty
  heading.
- **No world state**: with `gameMap == nil` or `replica == nil`, `v` is a strict
  no-op (the `x`-key documented-no-op precedent), and a mid-mode disconnect keeps the
  cursor usable over the static terrain with replica-derived bands empty.
- **Mode vs. higher layers**: takeover, help overlay, and a focused minibuffer all
  sit above the mode in the key chain — they own the keyboard while open and restore
  it after; the mode persists beneath them. Opening the guardian console (`G`) or a
  solo zoom (via `2`–`6`) exits the mode first — the borrow is transient and never
  survives being covered by a body-replacing surface.
- **World smaller than the viewport**: cursor and camera clamps use the same
  `renderMapGrid` viewport-clamp math (`vw > W → vw = W`), so tiny fixture worlds
  behave.
- **Run ended**: the mode works on an ended world — inspection is a reading surface,
  and every reading surface stays functional post-mortem (spec 044 posture).

## Requirements *(mandatory)*

### Functional Requirements

FR-001 through FR-010 map 1:1 to the board task's ten acceptance criteria.

- **FR-001** *(AC #1)*: `v` MUST toggle look-cursor mode; while active, the cursor
  MUST move with `hjkl` and the arrow keys (shift variants `H/J/K/L` = 8-tile jump)
  and MUST push the camera when it comes within 2 tiles of the viewport edge. The
  cursor spawns at the camera-center tile; exiting resumes centroid-following.
- **FR-002** *(AC #2)*: while the mode is active the dock MUST show a transient TILE
  view (terrain, structures, piles, agents with needs/intent, recent events at that
  tile) borrowed in the solo-zoom-seam style — a highlighted `TILE (x,y)` pseudo-label
  in the tab row, never a numbered tab; `2`–`5` (`6` on scenario worlds) MUST exit the
  mode and restore the chosen tab with its prior state intact.
- **FR-003** *(AC #3)*: `⏎`/`tab` MUST focus the tile pane with focus drawn
  (focus-contract rule 2: amber border); `j/k` + `⏎` MUST select and drill into
  agents/contents/events; `esc` MUST release exactly one layer per press —
  drill-in → pane rows → map cursor → exit mode — back to centroid-following
  (focus-contract rule 3).
- **FR-004** *(AC #4)*: the feature MUST introduce no new focusable text input; the
  focus contract's "exactly one client" claim MUST remain true, and
  `patterns/focus-contract.md` MUST say so in the same PR (a scope note like the
  villager-strip/console precedents).
- **FR-005** *(AC #5)*: mouse parity MUST ship with the keyboard (decision 8 rule 1):
  a left-click on a map tile moves the cursor (entering the mode if inactive); a
  click on a TILE pane row selects it, a second click on the selected row drills in.
  Clicks are no-ops while the help overlay is open or the minibuffer is focused.
- **FR-006** *(AC #6)*: `docs/design/tui/` MUST be re-verified and amended in the
  same PR — `panels/map.md` (deferral note re-opened + control-table rows),
  `patterns/keymap.md` (new mode table + `v` binding + footer hints),
  `panels/dock.md` (the borrow seam), `patterns/focus-contract.md` (scope note),
  `anatomy.md` (TILE view row), `overlays/help.md` (badge deep-link row flipped from
  unbuilt) — and `node scripts/check-tui-design.mjs --changed` MUST pass.
- **FR-007** *(AC #7)*: every tile's inspector header MUST list warmth and light
  levels — a discrete meter plus a plain-language source note (fire radius, shelter
  cover, open water; daylight, indoors, firelight, dark). The levels MUST derive from
  the sim's own mechanics (the `warmAt`/`litAt`/day-night truths `decayNeeds` and the
  gru already key on), exposed through a small read-only, pure sim-side helper —
  never re-derived TUI-side from duplicated radii, and never a reducer change.
- **FR-008** *(AC #8)*: map and dock panel geometry MUST stay fixed
  (`patterns/layout.md` column budget): entering the mode, focusing the pane, and
  drilling in swap content only — never panel size or position.
- **FR-009** *(AC #9)*: the TILE pane MUST list contents in DF's fixed hierarchy —
  agents → piles/chests → structures → terrain — a stable scan order on every tile.
- **FR-010** *(AC #10)*: the TILE pane's whatis prose MUST come from the spec-068
  tile registry's `meaning` rows (`internal/tui/tiles.go`) — plain language per the
  spec-047 FR-020 boundary — making the look-cursor the third in-place lookup after
  `?` and the guardian's explain tool. A tile row added to the registry MUST reach
  the TILE view with no renderer edit (the SC-002 seam extends to a fourth surface).

Feature-scoped requirements beyond the board ACs:

- **FR-011** *(badge deep-link — folded into this lane per decision 4)*: `?` pressed
  while a conditional header badge is active MUST open the help overlay pre-focused
  on that badge's `headerAnatomy` row in the screen-walkthrough section; with no
  active badge, open behavior is byte-identical to today. `overlays/help.md`'s
  layer-2 control-table row and byte-identity classification MUST be updated in the
  same PR (content stays static; only which row is pre-focused is status-derived —
  the classification already says exactly this).
- **FR-012** *(narrow fallback)*: in the narrow (<112 cols) layout the mode MUST be
  available while the map pane is active: `v` raises the cursor on the map pane and
  `⏎`/`tab` swaps that pane's body to the TILE view (a transient body replacement —
  the solo-views "one component, two widths" posture); the esc chain is unchanged.
  Content swap only, never a layout change (FR-008's spirit at one-pane scale).
- **FR-013** *(layering)*: the mode MUST be a key mode layered like inspect/villagers
  — below takeover, help, minibuffer focus, and the console in `handleKey`'s chain.
  While the borrow is active the inspect and villagers key layers MUST be dormant
  (the TILE view is the thing visible — the villagers-mode scoping precedent), and
  unclaimed global keys (`space`, `[`/`]`, `m`, `q`, `p`, `?`) MUST keep working.
  Opening the console or a solo zoom exits the mode; help/takeover layer above it
  without ending it.
- **FR-014** *(help completeness)*: the help overlay's keys section MUST document the
  mode (a look-cursor mode page reachable via `n`/`p`, and `?` pressed during the
  mode opens help frozen on it — the spec-045 completeness discipline), and the
  footer MUST show the mode's primary hints while it is active.

### Key Entities

- **Look-cursor mode state**: whether the mode is active; cursor tile (x, y); which
  focus layer holds the keyboard (map cursor / pane rows / drill-in); selected pane
  row; open drill-in target. Client-side only — never persisted, never in the replica.
- **TILE view**: the transient dock-body renderer for the cursor tile — header
  (coords, whatis prose, warmth/light levels) + banded rows in the fixed hierarchy.
  Not a dock tab: no `pane` enum value, no key digit, no cycle membership.
- **Tile environment sample**: the read-only sim-derived answer for one tile at one
  tick — warmth level + source (lit-fire radius / shelter cover / ambient day /
  night exposure) and light level + source (daylight / firelight / indoors / dark).
- **Hit regions**: per-frame map-grid and TILE-pane click geometry (the
  `chronHitRegion` precedent) — recorded at render, consumed by the mouse handler,
  invalidated when their surface wasn't part of the frame.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: from the home composite, a player can identify what is on any visible
  tile — including needs/intent of an agent standing there — in ≤ 4 keypresses
  (`v`, moves, read), without leaving the map or opening any list surface.
- **SC-002**: 100% of tiles the map can draw produce a TILE view header with
  non-empty whatis prose (registry `meaning` coverage extends to the fourth surface;
  the registered-test-tile round-trip proves no renderer edit is needed).
- **SC-003**: rendered panel geometry (map and dock width × height) is byte-stable
  across mode entry, pane focus, drill-in, and exit — verified by a rendering test
  diffing panel dimensions before/during/after.
- **SC-004**: the esc chain releases exactly one layer per press in 100% of states
  (drill-in, pane focus, cursor, mode-off), test-enumerated.
- **SC-005**: every new control lands with both a key and a mouse target in its
  control-table row (zero new parity-gap entries in the rollout notes), and
  `node scripts/check-tui-design.mjs --changed` passes on the PR.
- **SC-006**: the warmth/light levels shown for a tile agree with the sim's own
  mechanics on that tick (warm ⇔ `decayNeeds` would treat an agent there as warm;
  lit ⇔ the gru's light protection holds there) — asserted against the exported
  helper, not re-derived constants.
- **SC-007**: `go test ./...` green; existing map/legend/overlay byte-identity pins
  (`TestTilesIdentityPin` and friends) pass unchanged — the mode-off rendering is
  byte-identical to today.

## Assumptions

- **Reverse jump stays unscheduled** (runbook default, open question 4 resolved for
  this sweep): the strip-glyph/roster-row → camera-center reverse jump is OUT OF
  SCOPE here. Recorded recommendation: it deserves its own small board card once this
  lands — it is a camera writer like jump-to-source, not a cursor feature, and this
  spec's shared camera-origin helper (plan) is the seam it would reuse.
- The mode is a client-side presentation feature: no daemon/IPC changes, no replica
  or reducer changes; the one sim-side addition (FR-007) is a pure, read-only
  derivation helper over existing state.
- "Recent events at that tile" is bounded to the client's existing event ring
  (`m.events`) filtered by each event's *recorded* payload position via the
  spec-049 subject registry — no store queries, no new IPC. Events whose registry
  entry carries no position (or no registry entry) simply don't list; the raw-JSON
  drill keeps whatever does list fully inspectable.
- The exercise briefing any-key eater, when showing, continues to outrank the mode's
  entry key exactly as it outranks every other unfocused key (its check sits higher
  in `handleKey`); this is existing-order fallout, not a new rule.
- Skin neutrality: the TILE view renders engine truths (registry meanings, needs,
  intents, events) — no fiction-layer strings, so no new skin tokens.

## Notes

- **Pull-surface budget tension (reorientation open question 3 — recorded, not
  ruled)**: decisions 4 and 6 plus D9 all grow tab-cycled overlay/pane surfaces; the
  DF "three ways of scrolling" caution asks at what point the help/inspect stack
  needs its own navigation ruling. This spec deliberately makes NO navigation ruling:
  the TILE view reuses the existing `j/k`-select / `J/K`-scroll / `esc`-release
  grammar verbatim rather than inventing a new one, which is this feature's whole
  contribution to containing the tension. The question stays parked with its
  original resurfacing trigger (operator judgment as pull surfaces accumulate);
  re-raise it if a future surface cannot reuse the existing grammar.
