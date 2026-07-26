# Implementation Plan: Tile Registry + New Terrain Tiles

**Branch**: `task-143-tile-registry` | **Date**: 2026-07-26 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/068-tile-registry/spec.md`

## Summary

Grow the TUI's existing shared glyph table (`mapGlyphs`, help.go) into a full tile registry
— one data table carrying glyph, legend name, overlay meaning, classed style tokens
(semantic-16 / material-256), and tile-kind binding — consumed by `renderMapGrid`'s tile
resolution, `legendGlyphLine`, and the `?` overlay walkthrough, with zero per-tile style
literals left in views.go (byte-identical output for the shipped vocabulary). Then add two
walkable, non-buildable ground covers — marsh `░` (256:65) and sand `▒` (256:180) —
generated deterministically for new worlds only, gated by `terrain_gen:2` in the manifest
plus a FormatVersion 4→5 bump so old software refuses (never silently regenerates) new
worlds, with `promptworld migrate` carrying v4 worlds forward terrain-unchanged.
Full rationale in [research.md](research.md).

## Technical Context

**Language/Version**: Go (module `github.com/evanstern/promptworld`; toolchain per go.mod)

**Primary Dependencies**: charmbracelet lipgloss/bubbletea (TUI styling/rendering) — already
in use; no new dependencies

**Storage**: world save dir (`world.json` manifest); terrain never persisted — regenerated
from (seed, dims, terrain_gen)

**Testing**: `go test ./...`; package conventions — worldmap determinism/fraction tests,
tui rendering + coverage sweeps (help_test.go), sim overlay tests

**Target Platform**: darwin/linux terminals (256-color assumed, 16-color degradation
handled by glyph distinctness)

**Project Type**: single Go repo — daemon + TUI client + CLI

**Performance Goals**: no regression in map render path (registry lookup replaces switch;
table is ~20 rows, O(1)-ish by kind)

**Constraints**: FR-004 byte-identical rendering for shipped vocabulary; deterministic
generation (same seed+dims+gen → identical map, every platform); FormatVersion break must
ship with a working `promptworld migrate` 4→5

**Scale/Scope**: 4 packages (`internal/tui`, `internal/worldmap`, `internal/world`,
`internal/sim`), ~2 new terrain kinds, ~20-row registry, 1 manifest field + version bump

## Constitution Check

*GATE: v1.1.0. Evaluated pre-Phase-0 and re-checked post-Phase-1 design — PASS.*

- **I. Artifact-Grounded Action** — PASS: scope ratified by operator on the board
  (TASK-143); decisions derived from the Game-UI-UX analysis vault artifacts; this plan +
  research.md are the new artifacts.
- **II. One Task, One PR** — PASS: TASK-143 ↔ `task-143-tile-registry` branch in
  `.worktrees/task-143` ↔ one PR. Spec docs commit to main at root (project convention).
- **III. Gates Over Assertions** — PASS: spec-bridge link before implementation;
  merge-drift worktree/pr gates; SC-001/004/005 are test-enforced, not asserted.
- **IV. Grounding Freshness** — PASS (planned): touched files are listed as sources by
  `docs/wiki/tui-client.md`, `world-model.md`, `world-migration.md` (at minimum) —
  `/grounding-wiki:wiki-update` + player-docs freshness check are post-merge steps in
  tasks.md, and the task's AC #4 carries them.
- **V. Model-Tiered Workflow** — PASS: planned/spec'd on Fable 5; implementation delegated
  to `spec-implementer` at **Opus 4.8** (cross-package + format-version break + determinism-
  sensitive generation; rubric justification recorded on TASK-143). Never inline.

No violations → Complexity Tracking left empty.

## Project Structure

### Documentation (this feature)

```text
specs/068-tile-registry/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0 — decisions R1-R8
├── data-model.md        # Phase 1 — registry/table/manifest shapes
├── quickstart.md        # Phase 1 — end-to-end validation guide
├── contracts/
│   └── tile-registry.md # Phase 1 — registry invariants + manifest contract
├── checklists/
│   └── requirements.md  # Spec quality checklist (passed)
└── tasks.md             # Phase 2 (/speckit-tasks — not created by plan)
```

### Source Code (repository root)

```text
internal/
├── tui/
│   ├── help.go          # glyphEntry → registry rows (glyph/name/meaning + style + binding)
│   ├── tiles.go         # NEW: the tile registry + style-token table (single source)
│   ├── views.go         # renderMapGrid tile() resolves via registry; literals removed
│   └── *_test.go        # byte-identity pin, token-class sweep, coverage sweeps extended
├── worldmap/
│   ├── worldmap.go      # Marsh/Sand kinds; Passable; generation-version parameter
│   ├── noise.go         # (unchanged)
│   └── worldmap_test.go # determinism + nonzero-count + legacy-identity (Hash) tests
├── world/
│   ├── world.go         # FormatVersion 5; Manifest.TerrainGen (omitempty); Open/New
│   └── migrate*.go      # 4→5 migration, terrain-preserving
└── sim/
    ├── terrain.go       # sim passable() adds new kinds
    ├── memory.go        # featureDesc names marsh/sand; kind-switch sweep
    ├── executor.go      # scan/look switch sweep
    └── miracles.go      # target-validation switch sweep
```

**Structure Decision**: single Go repo, four existing packages touched; one new file
(`internal/tui/tiles.go`) so the registry has a single obvious home; no new packages.

## Phase 1 Design Highlights

- **Registry shape** (data-model.md): rows keyed by role — terrain kinds, effective kinds
  (path/depleted), structure kinds, markers (pile/den/grave/dead/gru) — each row: glyph,
  legend name, overlay meaning, style token ref(s), state-variant transforms (dying, cold,
  damaged) referencing the same glyph per FR-003.
- **Token table**: every color used by a tile resolves through a named token classed
  `semantic16` or `material256`; night dimming and condition overlays remain transforms
  applied over token-resolved styles.
- **Manifest contract** (contracts/tile-registry.md): `format_version: 5`;
  `terrain_gen` omitted (legacy) or `2` (marsh+sand); Open-rejection and migrate behavior
  pinned as the FR-007 mechanism.
- **Post-design constitution re-check**: PASS (above reflects the re-check).
