# Implementation Plan: TUI Look-Cursor Mode — Tile Inspection with a Focusable Tile Pane

**Branch**: `task-142-look-cursor` | **Date**: 2026-07-26 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/074-look-cursor/spec.md`

## Summary

Add a look-cursor key mode to the map (`v`; `hjkl`/arrows move, shift = 8-tile jump,
camera pushes at a 2-tile margin, `c` snaps, `esc` exits) whose active state borrows
the dock body for a transient TILE view — a `TILE (x,y)` pseudo-label in the tab row,
a header carrying the tile-registry whatis prose plus sim-derived warmth/light levels,
and rows in DF's fixed hierarchy (agents → piles/chests → structures → terrain, plus
recent recorded-position events). `⏎`/`tab` focuses the pane (amber border), `j/k`+`⏎`
drill into the villager-detail renderer family / contents / the raw-JSON inspector,
and `esc` releases one layer per press. Mouse parity ships together (click-tile,
click-row — two new hit regions on the `chronHitRegion` pattern). The badge deep-link
(`?` pre-focused on the active header badge's row) folds into the same lane. One small
sim addition: an exported pure `sim.EnvAt` decomposing the existing `warmAt`/`litAt`
mechanics; zero reducer/IPC changes. Full decisions in [research.md](research.md);
shapes in [data-model.md](data-model.md).

## Technical Context

**Language/Version**: Go (module `github.com/evanstern/promptworld`; toolchain per go.mod)

**Primary Dependencies**: charmbracelet bubbletea/lipgloss (already in use — mouse
reporting already enabled via `tea.WithMouseCellMotion`, `cmd/promptworld/commands.go:856`);
no new dependencies

**Storage**: none — mode state is client-side Model fields, never persisted, never in
the replica

**Testing**: `go test ./...`; package conventions — tui key-routing tests
(`tui_test.go`), rendering/geometry tests (`render_test.go`), focus-contract sweeps
(`focus_test.go`), byte-identity pins (`tiles_identity_*.golden` — must pass
unchanged), help byte pins (`help_test.go`); sim helper unit tests

**Target Platform**: darwin/linux terminals (widescreen composite ≥112 cols + narrow
fallback, both in scope per FR-012)

**Project Type**: single Go repo — daemon + TUI client + CLI; this feature touches the
TUI client + one read-only sim helper

**Performance Goals**: no regression in the map render path (cursor highlight is one
style transform on one tile; TILE body assembly is O(agents+structures+piles) per
frame plus a bounded event-ring filter)

**Constraints**: mode-off rendering byte-identical (existing goldens/pins must pass
unchanged, SC-007); fixed panel geometry across all mode states (SC-003); no new
focusable text input (FR-004); no reducer changes (FR-007)

**Scale/Scope**: 2 packages (`internal/tui`, `internal/sim`), ~1 new tui file
(`look.go`) + edits to `tui.go`/`views.go`/`help.go`, 1 new sim file (`env.go`),
6 design pages amended, 3–4 wiki notes re-pinned

## Constitution Check

*GATE: v1.2.0. Evaluated pre-Phase-0 and re-checked post-Phase-1 design — PASS.*

- **I. Artifact-Grounded Action** — PASS: scope ratified on the board (TASK-142 ACs
  + operator-reviewed mock, revision "fixed-geometry-env-levels"); reorientation
  synthesis decision 4 is the deciding artifact; runbook defaults (reverse-jump
  unscheduled; pull-surface tension recorded not ruled) encoded in spec.md rather
  than re-asked. This plan + research.md + data-model.md are the new artifacts.
- **II. One Task, One PR** — PASS: TASK-142 ↔ `task-142-look-cursor` branch in
  `.worktrees/task-142` ↔ one PR. Spec docs commit to main at root (project
  convention); design-page amendments and wiki re-pins ride the PR branch.
- **III. Gates Over Assertions** — PASS: spec-bridge link before implementation;
  merge-drift claim/worktree/pr gates; SC-002/003/004/006/007 are test-enforced;
  AC #6 is gate-enforced (`check-tui-design.mjs --changed`).
- **IV. Grounding Freshness** — PASS (planned, in-branch per spec 069): touched files
  are pinned sources of `docs/wiki/tui-map-view.md`, `tui-dock-tabs.md`,
  `tui-input-help.md`, `tile-registry.md` (tiles.go is read, not edited — re-pin only
  if it changes), plus sim-note fallout from `internal/sim/env.go`'s extraction
  touching `terrain.go`/`gru.go`; re-pins + `docs/player/` regeneration are explicit
  tasks ON THE BRANCH (tasks.md Phase 9) — the pr gate blocks without them, no
  bypass.
- **V. Model-Tiered Workflow** — PASS: planned/spec'd on Fable 5; implementation
  delegated to `spec-implementer` at **Sonnet** (board-recorded tier: view/rendering
  feature in `internal/tui`; the AC7 helper is a read-only sim derivation, inside the
  routine tier). Escalation to Opus only via the rubric as an operator checkpoint.
  Never inline.

No violations → Complexity Tracking omitted.

## Project Structure

### Documentation (this feature)

```text
specs/074-look-cursor/
├── CLAIM.md             # Spec-065 claim stub (kept)
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0 — decisions R1-R10
├── data-model.md        # Phase 1 — mode state, TILE view, EnvSample, hit regions
└── tasks.md             # Phase 2 (/speckit-tasks)
```

### Source Code (repository root)

```text
internal/
├── tui/
│   ├── tui.go           # handleKey: look layer between console and inspect; v entry in
│   │                    #   handleGlobalKey; visibility predicates gain !lookActive guards;
│   │                    #   handleMouse: tileHit → mapHit → chronicle routing;
│   │                    #   openHelp badge pre-focus; currentHelpMode look case;
│   │                    #   Model: look* fields + mapHit/tileHit region pointers
│   ├── look.go          # NEW: handleLookKey (movement/focus/drill/esc chain), cursor
│   │                    #   clamps + camera push, tileBody renderer (header + bands),
│   │                    #   tileEvents filter, env level/note mapping
│   ├── views.go         # cameraOrigin extraction; renderMapGrid cursor highlight +
│   │                    #   in-mode title; mapPanelView/mapView record mapHit;
│   │                    #   dockTabsRow TILE pseudo-label; dockTabContent borrow branch
│   ├── help.go          # helpModeLook keys page; headerAnatomy row-index resolution for
│   │                    #   the badge deep-link; footer hints for the mode
│   └── *_test.go        # key-routing, esc-chain enumeration, geometry pin (SC-003),
│                        #   borrow state-preservation, hierarchy order, whatis round-trip,
│                        #   mouse hit-tests, help pre-focus + byte pins, identity pins
└── sim/
    ├── env.go           # NEW: EnvAt + EnvSample (pure, read-only); shared cores with
    │                    #   warmAt/litAt so predicates and sample cannot disagree
    ├── terrain.go       # warmAt thins to the shared core (behavior unchanged)
    ├── gru.go           # litAt thins to the shared core (behavior unchanged)
    └── env_test.go      # NEW: EnvAt ≡ warmAt/litAt matrix; source attribution cases

docs/design/tui/         # same-PR amendments (FR-006): panels/map.md (deferral
                         #   re-opened + control rows), patterns/keymap.md (mode table,
                         #   v binding, footer hints), panels/dock.md (borrow seam),
                         #   patterns/focus-contract.md (scope note), anatomy.md (TILE
                         #   row), overlays/help.md (deep-link row flipped)
docs/wiki/               # in-branch re-pins: tui-map-view, tui-dock-tabs,
                         #   tui-input-help, tile-registry (as applicable), sim notes
                         #   sourcing terrain.go/gru.go
docs/player/             # regenerated in-branch when the wiki changes (spec 069)
```

**Structure Decision**: single Go repo; two new files (`internal/tui/look.go`,
`internal/sim/env.go`) so the mode and the derivation each have one obvious home; no
new packages, no daemon/IPC surface.

## Phase 1 Design Highlights

- **Key routing** (research R1): `lookActive` layer between console and inspect in
  `handleKey`; claims `v h j k l H J K L` arrows `c ⏎ tab esc 1-6 g G J K d`; digits
  clear mode state then delegate to `selectTab`; everything else falls through to
  global. `chronicleVisible`/`villagersVisible`/`exerciseVisible` gain
  `!m.lookActive` guards so every visibility consumer (badging, mouse, help-mode
  freeze) agrees the TILE view is the thing visible.
- **Borrow seam** (R2): `dockTab` is never written — `dockTabsRow` renders the
  highlighted `TILE (x,y)` pseudo-label and `dockTabContent` short-circuits to
  `tileBody`; per-tab state survives by construction. Console/solo exit the mode;
  help/takeover layer above it.
- **Camera** (R3): extract `cameraOrigin(vw, vh)` (the `wandererCentroid` precedent)
  shared by `renderMapGrid`, the push logic, `c`-snap (`centerCameraOn` formula), and
  mouse hit-testing; exit resets pan to (0,0) = following.
- **TILE view** (data-model.md): header (coords · registry `meaning` whatis · warmth/
  light meters + notes) + bands in DF order — agents (name, awake/asleep/dead,
  needs bars, intent goal) → piles/chests (contents via
  `summarizePileContents`/`describeChest`) → structures (registry names + state) →
  terrain (registry row) → recent events (`tileEvents`, recorded-position filter,
  most recent first). Empty bands render nothing. Row model doubles as the selection
  and hit-region source so keyboard, mouse, and rendering can't disagree.
- **Drill-ins** (R5): agent → `villagerDetailBody` family rendered in-pane for that
  agent index; event → `formatInspector`/`chronicleDetailPane` raw-JSON family
  (FR-020 boundary); chest/pile → contents detail. All inside the same pane budget —
  content swap only (SC-003).
- **Env levels** (R4): `sim.EnvAt` decomposes warm (fire radius / shelter) and lit
  (firelight) sources from the same private cores `warmAt`/`litAt` wrap; TUI maps to
  3-step meters + plain-language notes (daylight / firelight / indoors / dark; the
  gru-safety note restates `gruProtected`). No "canopy" note — no mechanic backs it
  (honesty doctrine).
- **Mouse** (R6): `mapHit` + `tileHit` per-frame regions (the `chronHitRegion`
  pointer pattern); `handleMouse` order: guards → tile pane → map grid → chronicle.
- **Badge deep-link** (R8): `openHelp` pre-focuses the screen section on the first
  active badge's `headerAnatomy` row (index resolved from the shared table); no
  badge → byte-identical open.
- **Help/footer completeness** (R9): `helpModeLook` page + `currentHelpMode` case +
  footer hint line(s) for the mode.
- **Post-design constitution re-check**: PASS (the check above reflects the
  re-check).
