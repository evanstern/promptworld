# Tasks: TUI Look-Cursor Mode — Tile Inspection with a Focusable Tile Pane

**Input**: Design documents from `/specs/074-look-cursor/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md

**Tests**: INCLUDED — SC-002 through SC-007 are explicitly test-enforced, and AC #6
is gate-enforced (`check-tui-design.mjs --changed`, merge-drift pr gate).

**Organization**: grouped by user story after a substrate phase; US1 (cursor) + US2
(TILE view) together are the MVP increment; US3 (focus/drill), US4 (mouse), US5
(badge deep-link) build on them. Phase 9 carries the same-PR gate obligations —
design-reference amendments, wiki re-pins, player docs — as explicit tasks (spec 069:
grounding rides the PR, no post-merge tail).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: parallelizable (different files, no dependency on an incomplete task)
- **[Story]**: US1–US5 from spec.md

## Phase 1: Setup — shared substrate (no behavior change)

**Purpose**: the pure helpers every story consumes, landed with the existing suite
byte-green so regressions are attributable.

- [X] T001 Extract `cameraOrigin(vw, vh) (x0, y0)` on Model in
      `internal/tui/views.go` from `renderMapGrid`'s inline centroid+pan+clamp math
      (research R3, the spec-049 `wandererCentroid` extraction precedent), plus a
      `mapViewportDims()` helper recomputing (vw, vh) from
      `computeColumns`/`computeRows`/`mapViewportTiles` for both widescreen home and
      the narrow map pane. `renderMapGrid` calls them; rendered bytes unchanged
      (existing goldens + `TestTilesIdentityPin` must pass untouched).
- [X] T002 [P] Add `internal/sim/env.go`: exported pure `EnvAt(s, x, y, tick)
      EnvSample` (`Warm`+`WarmSource`, `Lit`) sharing private cores with `warmAt`
      (`internal/sim/terrain.go:140`) and `litAt` (`internal/sim/gru.go:70`), which
      thin to wrappers — behavior unchanged (research R4; FR-007's no-reducer-change
      constraint). Tests in `internal/sim/env_test.go`: `EnvAt ≡ warmAt/litAt` across
      a lit/dying/cold-fire × shelter × tick matrix; source attribution (fire vs
      shelter precedence matches `warmAt`'s own loop order) — SC-006.

**Checkpoint**: `go test ./...` green with zero golden edits; commit before any mode
code.

---

## Phase 2: Foundational — mode state, key layer, borrow seam

**Purpose**: the skeleton both P1 stories hang off; blocks all user stories.

- [X] T003 Model state + entry/exit in `internal/tui/tui.go` and new
      `internal/tui/look.go` (data-model.md "Look-cursor mode state"): `lookActive`,
      `lookX/Y`, `lookFocus`, `lookSel`, `lookDrill`, `lookDrillScroll`; `v` case in
      `handleGlobalKey` gated on map-visible + `gameMap != nil` (strict documented
      no-op otherwise — the `x` precedent); entry spawns the cursor at camera center;
      exit zeroes look state and resets `panX/panY` (resume following). `handleKey`
      gains the `lookActive → handleLookKey` layer between the console and inspect
      checks (research R1), with `esc` and `v` exiting from cursor focus.
- [X] T004 Visibility dormancy (research R1):
      `chronicleVisible`/`villagersVisible`/`exerciseVisible` in
      `internal/tui/tui.go` gain `!m.lookActive &&` guards so inspect/villagers key
      layers, reply badging (`guardianVisible` callers), mouse guards, and
      `currentHelpMode` all treat the TILE view as the thing visible during the
      borrow. Console (`G`) and solo-zoom paths exit the mode first; help/takeover
      layer above without ending it (FR-013).
- [X] T005 Borrow seam in `internal/tui/views.go` (research R2): `dockTabsRow`
      renders the highlighted `TILE (x,y)` pseudo-label (normal labels dim-inactive)
      and `dockTabContent` short-circuits to `tileBody` while `lookActive` —
      `m.dockTab` never written, placeholder body acceptable until T009. Digits
      `2`–`6` in `handleLookKey` clear mode state then delegate to `selectTab`.
- [X] T006 Foundational tests in `internal/tui/tui_test.go` /
      `internal/tui/look_test.go`: `v` toggles (and no-ops with no world state /
      while solo / in console); mode entry from chronicle-inspect preserves
      `chronSelected`/`chronDetailScroll` and `2` restores inspect intact (spec edge
      case); villagers/inspect `j/k` never fire during the borrow; `space`/`[`/`]`/
      `m`/`q` still work in-mode (FR-013); guardian reply mid-borrow badges rather
      than streams.

**Checkpoint**: mode enters/exits cleanly with a placeholder TILE body; dock state
preservation proven.

---

## Phase 3: User Story 1 — cursor movement + camera (P1) 🎯 AC #1, AC #8

**Goal**: `v`-cursor with hjkl/arrows, shift jumps, margin push, `c` snap — fixed
geometry throughout.

**Independent test**: walk the cursor to every world edge and back at a fixed
terminal size; camera tracks at the 2-tile margin; panel dimensions never change.

- [X] T007 [US1] Movement + camera in `internal/tui/look.go`: `hjkl`+arrows move 1,
      `H/J/K/L` move 8, clamped to `[0,W)×[0,H)` (never wrap); camera push keeps the
      cursor ≥2 tiles inside the viewport via `cameraOrigin`/`mapViewportDims`
      adjusting `panX/panY`, degrading at world edges (camera stops, cursor may reach
      the border); `c` snaps camera to cursor (the `centerCameraOn` formula).
- [X] T008 [US1] Cursor render + title in `internal/tui/views.go`: `renderMapGrid`
      applies a background-highlight style transform on the cursor tile while
      `lookActive` (style transform, never a glyph change — spec-068 FR-003
      discipline); `mapPanelView`/`mapView` title becomes
      `MAP · cursor (x,y) · c center · esc exit` in-mode. Mode-off bytes untouched.
- [X] T009 [P] [US1] Tests: movement/clamp/jump matrix incl. world-smaller-than-
      viewport fixtures; camera-push geometry (cursor at margin ⇒ pan delta, at world
      edge ⇒ no pan); arrows move the cursor never the free pan in-mode (AS 1.5);
      exit resumes following (`panX/panY == 0`); the SC-003 geometry pin — full
      composite rendered at fixed size in mode-off/cursor/pane/drill states asserts
      per-panel width×height equality (research R10; drill states activate after
      T013, wire the harness now with the first two states).

**Checkpoint**: US1 demonstrable alone — cursor + coordinates answer "which tile is
that" before any TILE content exists.

---

## Phase 4: User Story 2 — the TILE view (P1) 🎯 AC #2, AC #7, AC #9, AC #10

**Goal**: the borrowed dock body shows the tile in DF's fixed hierarchy with
registry whatis prose and warmth/light levels.

**Independent test**: park the cursor on a composed fixture tile (agent + pile +
chest + shelter + fire radius, day and night) and read the pane top-to-bottom.

- [X] T010 [US2] `tileBody` assembly + renderer in `internal/tui/look.go`
      (data-model.md "TILE view row model"): header `TILE (x,y) · <registry
      meaning>`; bands strictly agents (needs bars + intent; gru row when abroad
      here) → piles/chests (`summarizePileContents`/`describeChest`) → structures
      (registry names + fire lit/dying/cold, wall damaged, grave; dens) → terrain
      (one row, effective-kind resolution matching `tile()`'s path/quarried
      overrides) → events; empty bands render nothing (dead/empty-tile edge cases);
      whatis resolves through the registry rows (`internal/tui/tiles.go`) with no
      renderer edit needed for future rows (FR-010).
- [X] T011 [P] [US2] Env header in `internal/tui/look.go`: map
      `sim.EnvAt` + `State.Night` + on-shelter into the warmth/light levels + notes
      table (data-model.md; 3-step discrete meters; gru-safety notes restate
      `gruProtected`; "open water" terrain flavor; NO canopy note — research R4).
- [X] T012 [P] [US2] `tileEvents(x, y)` in `internal/tui/look.go` (research R5):
      filter `m.events` through `subjectRegistry` keeping only recorded-position
      (`hasPos`) matches, most recent first, pane-budget capped; never decodes
      unregistered types.
- [X] T013 [US2] Tests: hierarchy order pinned on the composed fixture (AC #9);
      registry round-trip — `registerTile` a test row, assert its meaning reaches the
      TILE header with no renderer edit (SC-002, the spec-068 SC-002 seam's fourth
      surface); env header cases day/night × fire/shelter/exposed × water flavor
      (AC #7); events-at-tile includes recorded-position matches and excludes
      actor-only events; empty-tile renders header + terrain only (AS 2.5); TILE body
      respects the handed (width, height) budget (clip discipline, layout.md B1).

**Checkpoint**: MVP complete — US1 + US2 is a shippable inspection increment.

---

## Phase 5: User Story 3 — pane focus, drill-in, esc chain (P2) 🎯 AC #3, AC #4

**Goal**: `⏎`/`tab` focus (drawn amber), `j/k`+`⏎` drill, one esc layer per press;
the focus contract's one-client claim stays true.

**Independent test**: the four-esc unwind from an open drill back to
centroid-following, one visible layer per press.

- [X] T014 [US3] Focus + selection in `internal/tui/look.go`/`views.go`: `⏎`/`tab`
      from cursor focus → `lookFocusPane` with the pane border amber
      (`panelFocus` token — focus is drawn, rule 2); `j/k` move `lookSel` over
      drillable rows (clamped, `clampVillSelected` idiom); `esc` releases pane →
      cursor.
- [X] T015 [US3] Drill-ins in `internal/tui/look.go`: `⏎` on an agent row renders
      the villager-detail family (`villagerDetailBody`) for that agent index
      in-pane; an event row renders the raw-JSON inspector family
      (`formatInspector`/`chronicleDetailPane` — the FR-020 raw-behind-a-drill
      boundary); a chest/pile row renders contents detail; `J/K`-style scroll via
      `lookDrillScroll` (render-clamped); `esc` releases drill → pane.
- [X] T016 [US3] Narrow fallback (FR-012, research R7): `v` on the active map pane
      raises the cursor; `⏎`/`tab` swaps that pane's body to the TILE view
      (transient body replacement); same esc chain; `mapView` shares the cursor
      highlight/title from T008.
- [X] T017 [P] [US3] Tests: SC-004 esc-chain enumeration (drill → pane → cursor →
      off → global esc untouched, exactly one layer per press, every state); no new
      text capture — extend `internal/tui/focus_test.go`'s contract sweep to assert
      printable keys in every look state never buffer anywhere (FR-004); complete
      the SC-003 geometry pin's pane/drill states (T009 harness); narrow-mode entry
      /swap/unwind; fold-pressure resize mid-mode neither crashes nor changes any
      panel beyond the fold itself (spec edge case).

---

## Phase 6: User Story 4 — mouse parity (P2) 🎯 AC #5

**Goal**: click-tile moves/enters the cursor; click-row selects, second click
drills — landing keyboard+mouse together (decision 8 rule 1).

- [X] T018 [US4] Hit regions (data-model.md): `mapHitRegion` recorded by
      `mapPanelView`/`mapView` (grid origin, `x0/y0`, `vw/vh`, 2-col stride) and
      `tileHitRegion` recorded by the TILE body renderer (`rowIndex` from the same
      `tileRow` slice) — the `chronHit` pointer/invalidation lifecycle
      (`internal/tui/tui.go:303-315`).
- [X] T019 [US4] `handleMouse` routing in `internal/tui/tui.go` (research R6):
      existing guards (left-release only; help-open/minibuffer-focused no-ops) →
      TILE pane region (select row, acquiring pane focus; second click on the
      selected row drills) → map region (move cursor / enter mode) → existing
      chronicle path unchanged.
- [X] T020 [P] [US4] Tests in `internal/tui/tui_test.go` (the `mouseLeftRelease`
      helper): click-tile enters the mode at the clicked tile / moves an active
      cursor (AS 4.1/4.2); pane row click + double-click drill (AS 4.3); guard
      no-ops (AS 4.4); out-of-region and stride-boundary hit math; chronicle click
      still works mode-off.

---

## Phase 7: User Story 5 — badge deep-link (P3) 🎯 help.md layer-2 row

- [X] T021 [US5] `openHelp` pre-focus in `internal/tui/tui.go`/`help.go` (research
      R8): with ≥1 active conditional header badge (`[degraded]`, `[llm: …]`,
      `[suppressed: …]` — the same predicates `headerView` uses), open on the
      screen-walkthrough section scrolled so the first active badge's
      `headerAnatomy` row is visible (index resolved from the shared table); no
      badge → byte-identical open.
- [X] T022 [P] [US5] Tests: badge-active open lands on the screen section at the
      right row per badge kind; badge-free open byte-identical (existing
      `help_test.go` pins must pass unchanged); overlay navigation after a
      pre-focused open unchanged (AS 5.3).

---

## Phase 8: In-app reference completeness (FR-014)

- [X] T023 `helpModeLook` keys page in `internal/tui/help.go` (research R9):
      `currentHelpMode` gains the look case right after the console branch; the keys
      section pages through the mode's table (`v`, moves, jumps, `c`, `⏎`/`tab`,
      `j/k ⏎`, `2`–`6`, `esc`); footer hints for cursor and pane-focused states in
      `footerView`. Tests: `?` mid-mode freezes on the look page;
      `nextHelpMode`/`prevHelpMode` include it; footer hint per state.

---

## Phase 9: Same-PR gate obligations — design reference, grounding, verification

**Purpose**: AC #6 + constitution IV/spec 069 — these ride the task branch; the
merge-drift pr gate blocks without them (no bypass flag).

- [X] T024 Amend `docs/design/tui/` per data-model.md's delta table, re-verifying +
      re-pinning every touched page: `panels/map.md` (the "Look-cursor: evaluated
      and deferred" resolution note re-opened, naming the operator signal and this
      spec; new control-table rows with real mouse targets; parity note updated),
      `patterns/keymap.md` ("Mode: look-cursor" table; `v` in the binding-selection
      note; footer hints), `panels/dock.md` (borrow seam), `patterns/
      focus-contract.md` (scope note: still exactly one text client — FR-004's "its
      page says so"), `anatomy.md` (TILE view + look-cursor rows),
      `overlays/help.md` (badge deep-link row flipped from `unbuilt (wave 4,
      layer-2)` to its renderer; look-mode keys page row).
- [X] T025 Run `node scripts/check-tui-design.mjs --changed` from the worktree and
      fix every finding (AC #6's gate half; SC-005).
- [X] T026 In-branch wiki re-pins (`/grounding-wiki:wiki-update` scoped to this
      branch's diff): `docs/wiki/tui-map-view.md`, `tui-dock-tabs.md`,
      `tui-input-help.md`, `tile-registry.md` (only if `tiles.go` actually changed),
      plus any sim note sourcing `internal/sim/terrain.go`/`gru.go` (the
      `warmAt`/`litAt` core extraction). Pins must be branch commits reachable from
      the branch tip (spec 069 `wiki-repin-missing`).
- [X] T027 Regenerate `docs/player/` in-branch (player-docs skill) since T026
      touches `docs/wiki/`; verify with
      `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`
      (spec 069 `player-docs-stale`).
- [X] T028 Full verification: `go build ./... && go vet ./... && go test ./...`
      green (SC-007 — all pre-existing goldens/byte pins pass unchanged); run
      `node scripts/check-merge-drift.mjs pr` from the worktree before opening the
      PR; record evidence on TASK-142 board notes. Merge with
      `gh pr merge --merge` (merge-commit-only — in-branch pins die under squash).

---

## Dependencies & execution order

```
T001, T002 (substrate, parallel)
  → T003 → T004 → T005 → T006                      (Phase 2 foundational)
      → US1: T007 → T008 → T009
      → US2: T010 → {T011, T012} → T013            (US2 needs T005's borrow; T011 needs T002)
          → US3: T014 → T015 → T016 → T017          (drill reuses US2's rows)
              → US4: T018 → T019 → T020             (tileHit needs the row model + focus)
US5: T021 → T022                                    (independent — may run any time after T001)
T023 after T014 (footer states exist)
Phase 9: T024 → T025 (gate needs the amendments); T026 → T027; T028 last.
```

**Parallel opportunities**: T001∥T002; T009∥T010 (different concerns once T008
lands); T011∥T012; T017∥T018-prep; US5 (T021–T022) parallel to US3/US4; T024
drafting parallel to Phase 6–8 code (re-verify against final code before T025).

**MVP scope**: Phases 1–4 (T001–T013) — cursor + TILE view, keyboard-only. US3–US5
complete the board ACs; nothing ships before Phase 9's gates pass (one task, one PR).
