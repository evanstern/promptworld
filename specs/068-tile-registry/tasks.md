# Tasks: Tile Registry + New Terrain Tiles

**Input**: Design documents from `/specs/068-tile-registry/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/tile-registry.md

**Tests**: INCLUDED — the spec's success criteria are explicitly test-enforced (SC-001–SC-006)
and the contract file names its enforcing tests (C1–C13).

**Organization**: grouped by user story; US1 (registry) is the MVP increment, US2
(marsh/sand) builds on it.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: parallelizable (different files, no dependency on an incomplete task)
- **[Story]**: US1 / US2 from spec.md

## Phase 1: Setup — pin current behavior BEFORE touching anything

**Purpose**: the byte-identity contract (C6) and legacy-terrain contract (C7) need their
"before" truth captured first, against unmodified code.

- [x] T001 Add the byte-identity pin: a fixture test in `internal/tui/tiles_identity_test.go`
      that builds a representative replica (fixed seed via `worldmap.Generate`; agents
      awake/asleep/dead/critical/suppressed; fire lit/dying/cold; walls intact/damaged;
      shelter, oven, chest, pile, den, grave, path, quarried; day AND night) and pins the
      exact output of `renderMapGrid` + `legendGlyphLine()` + the `?` overlay glyph
      walkthrough as committed expected bytes. Must pass against the CURRENT renderer
      before any refactor commit (C6, SC-001).
- [x] T002 [P] Add the legacy-terrain pin: extend `internal/worldmap/worldmap_test.go` with
      a pinned `Map.Hash()` for a fixed (seed, w, h) legacy generation — the value the
      gen-parameter refactor and migration must keep reproducing (C7, feeds C11).

**Checkpoint**: both pins green on unmodified code; commit them first.

---

## Phase 2: Foundational — the registry substrate

**Purpose**: the table both stories consume.

- [x] T003 Create `internal/tui/tiles.go`: the StyleToken table (every color currently in
      `views.go:1581-1644`, classed `semantic16`/`material256` per data-model.md — colors
      unchanged) and the TileEntry registry (glyph, legend name, overlay meaning, style
      token ref, state-variant transforms, binding) for the entire shipped vocabulary,
      growing `glyphEntry`/`mapGlyphs` from `internal/tui/help.go:33-75` rather than
      duplicating it (R1, data-model.md shapes).

**Checkpoint**: `go build ./...` green; registry exists but nothing reads it yet.

---

## Phase 3: User Story 1 — one tile table drives map, legend, overlay (P1) 🎯 MVP

**Goal**: registry is the single presentation source; zero per-tile literals outside it;
output byte-identical for the shipped vocabulary.

**Independent test**: T001's pin + full existing tui suite pass with the renderer fully
registry-driven; a registered test tile reaches map/legend/overlay with no renderer edit.

- [x] T004 [US1] Rewire `legendGlyphLine()` and the `?` overlay glyph walkthrough in
      `internal/tui/help.go` to read the grown registry rows (they already read
      `mapGlyphs`; this keeps that seam intact through the type change).
- [x] T005 [US1] Rewire `renderMapGrid`'s `tile()` leaves in `internal/tui/views.go`
      (lines ~1113-1330) to resolve glyph+style through the registry — priority order
      unchanged (gru > agents > structures > piles > dens > path > depleted > terrain);
      agent letter/case resolution stays in `tile()`, styles come from registry tokens;
      night dimming and condition overlays become transforms over token-resolved styles
      (C5, FR-003). Delete the per-tile style literal block `views.go:1581-1644`.
- [x] T006 [P] [US1] Add the sweep tests in `internal/tui/tiles_test.go`: (a) no per-tile
      `lipgloss` color/style literal outside `tiles.go` — package-source scan, TASK-121
      sweep precedent (C3, SC-004); (b) every StyleToken is classed and semantic16 tokens
      use only ANSI 0-15 (C4); (c) every variant reuses its base row's glyph (C5).
- [x] T007 [P] [US1] Add the round-trip test in `internal/tui/tiles_test.go`: register a
      test-only tile row and assert it renders on the map, appears in `legendGlyphLine()`,
      and appears in the overlay walkthrough with no renderer edits (C1, SC-002); assert
      full decode — every drawable glyph has a row with non-empty name+meaning (C2, SC-005
      — extend help_test.go's existing coverage sweep if that's the natural home).
- [x] T008 [US1] Verify: T001 byte-identity pin + entire `go test ./internal/tui/...`
      green with zero pin edits (C6, SC-001).

**Checkpoint**: US1 delivers alone — registry-driven renderer, byte-identical, MVP done.

---

## Phase 4: User Story 2 — marsh and sand (P2)

**Goal**: two new walkable ground covers in new worlds only, fully decoded by the UI,
invisible to old worlds, refused by old software.

**Independent test**: quickstart.md §2-4 — migrate a v4 world (terrain unchanged), create
a v5 world (marsh+sand present, legend/overlay decode them), old build refuses the v5 world.

- [x] T009 [US2] `internal/worldmap/worldmap.go`: append `Marsh`, `Sand` TileKind values
      (after `Depleted` — byte values of existing kinds frozen, data-model.md); add both
      to `Passable`; `Buildable` unchanged; add the generation-version parameter (legacy
      default reproduces today's output exactly) and the gen=2 placement passes — marsh =
      moist dry-land 4-adjacent to water above a marshFraction percentile, sand = remaining
      dry-land 4-adjacent to water, applied after existing passes with tuning constants
      beside `waterFraction` (R4).
- [x] T010 [US2] Extend `internal/worldmap/worldmap_test.go`: T002's legacy Hash pin still
      exact under the new parameter (C7); gen=2 → Marsh>0 and Sand>0 across a seed sample
      (C8, SC-003); gen=2 determinism (same inputs → same Hash, repeated); Marsh/Sand
      passable, not buildable, absent from legacy maps (C9).
- [x] T011 [US2] `internal/world/world.go` (+ the migrate implementation it points at):
      `FormatVersion` 4 → 5; `Manifest.TerrainGen int` with `json:"terrain_gen,omitempty"`;
      `New` writes `terrain_gen: 2`; `Open` treats absent/0 as legacy and rejects unknown
      values with a clear error (C12); `promptworld migrate` upgrades 4 → 5 WITHOUT setting
      terrain_gen (C11). Update every `worldmap.Generate` caller (daemon, TUI replica
      bootstrap, tests — sweep the repo) to pass the manifest's generation version.
- [x] T012 [P] [US2] World/migration tests in `internal/world/`: migrate preserves terrain
      (`Map.Hash()` equal before/after, C11); Open rejects unknown `terrain_gen` (C12);
      v5+terrain_gen:2 manifest round-trips; version-mismatch rejection still exact (C10).
- [x] T013 [P] [US2] `internal/sim`: add Marsh/Sand to `passable` (`terrain.go:38`);
      `featureDesc` returns "the marsh" / "the sand flat" (`memory.go:107`); sweep every
      terrain-kind switch — `executor.go:496`, `memory.go:126`, `miracles.go:524`,
      `terrain.go:13` — adding deliberate cases or documented fall-throughs (C13, FR-008);
      tests for naming + walkability in the relevant `_test.go` files.
- [x] T014 [US2] `internal/tui/tiles.go`: add the marsh row (`░`, material token 256:65,
      meaning per data-model) and sand row (`▒`, material token 256:180); confirm the
      coverage sweep (T007) and legend/overlay pick them up with no renderer edits — that
      IS SC-002 exercised for real; extend the identity fixture with a gen=2 corner or a
      second small fixture so night-dimmed marsh/sand rendering is pinned (US2-AS5
      distinctness: assert glyphs differ from every registry glyph).

**Checkpoint**: quickstart §§1-3 pass end-to-end.

---

## Phase 5: Polish & cross-cutting

- [x] T015 Full-suite + build verification: `go build ./... && go vet ./... && go test ./...`
      green; run quickstart.md §3 against a fresh world and §2 against a fixture v4 world
      (record evidence on TASK-143 board notes).
- [x] T016 Post-merge re-grounding (runs at root AFTER the PR merges, per project law):
      `/grounding-wiki:wiki-update` for notes sourcing the touched files (at minimum
      `docs/wiki/tui-client.md`, `docs/wiki/world-model.md`, `docs/wiki/world-migration.md`),
      then `node .claude/skills/player-docs/scripts/check-freshness.mjs --check` and refresh
      player docs if stale (constitution IV; board AC #4).

---

## Dependencies & execution order

```
T001, T002 (pins, parallel) → T003 (registry) → US1: T004 → T005 → {T006, T007} → T008
                                                US2: T009 → T010 → T011 → {T012, T013} → T014
US2 starts after US1's T008 (registry must be the live source before new rows prove SC-002).
T015 after T014. T016 after merge.
```

**Parallel opportunities**: T001∥T002; T006∥T007 (different test concerns);
T012∥T013 (different packages).

**MVP scope**: Phase 1 + 2 + User Story 1 (T001-T008) — a shippable, byte-identical
registry refactor. US2 is the visible payoff increment.
