# Data Model — Tile Registry + New Terrain Tiles (spec 068)

Phase 1. Shapes only — field names are indicative; the implementer may adjust identifiers,
never semantics. Invariants live in [contracts/tile-registry.md](contracts/tile-registry.md).

## StyleToken

The unit a future skin would override. One table, package `internal/tui` (new `tiles.go`).

| Field | Type | Notes |
|---|---|---|
| name | identifier | e.g. `tokWater`, `tokShelter`, `tokCritical` |
| class | enum: `semantic16` \| `material256` | analysis palette rule (FR-002) |
| color | terminal color ref | semantic16 ⇒ ANSI 0–15 slot; material256 ⇒ fixed 256 code |
| emphasis | flags | bold / faint / underline, as today |

**Classing of the shipped palette** (colors unchanged — classing is metadata):
semantic16: water ANSI 4, tree ANSI 2, forage ANSI 3, den ANSI 5, agent-critical ANSI 1,
dead red. material256: shelter 130, suppressed 135, chest 136, path 137, oven 166, pile
178, gru 196, fire-dying 202, fire 208, depleted/cold/damaged 240, grave 244, rock 245,
wall 250. New: marsh 65, sand 180 (material256).

## TileEntry (the registry row)

Grown from `glyphEntry` (help.go:33) — same table keeps feeding legend + overlay.

| Field | Type | Notes |
|---|---|---|
| glyph | string | the character(s) drawn (e.g. `▤▩` walls share one legend row, as today) |
| name | string | compact legend token text |
| meaning | string | plain-language `?` overlay walkthrough sentence |
| style | StyleToken ref | base presentation |
| variants | map state → transform | e.g. fire: dying→tokFireDying, cold→tokSpent; wall: damaged→tokSpent. Same glyph, different style only (FR-003) |
| binding | Binding | what world thing this row renders (below) |

**Binding** is one of: terrain kind (`worldmap.Grass/Water/Tree/Forage/Rock/Marsh/Sand`),
effective-kind override (`Depleted`, path), structure kind string (`fire`, `shelter`,
`oven`, `chest`, `wall_plank|wall_stone`), or marker (pile, den, grave, dead, gru, agent).
Agent rows keep their special resolution (initial letter, case, condition overlays) — the
registry supplies their *styles*, resolution order stays in `tile()`.

**Resolution order** (unchanged, FR/edge case): gru > agents > structures > piles > dens >
path > depleted > base terrain > grass default. The registry replaces the *leaf* of each
branch (which glyph+style), never the priority logic.

## Terrain kinds (worldmap)

```
Grass, Water, Tree, Forage, Rock, Depleted (effective-only)  — existing, values frozen
Marsh, Sand                                                   — new, appended AFTER existing
```

- Appended enum values so existing `TileKind` byte values (and `Map.Hash()` streams) are
  untouched.
- `Passable` (worldmap): + Marsh, Sand. `Buildable`: unchanged (Grass only).
- sim `passable` (terrain.go): + Marsh, Sand.
- `effectiveKind`: no overlays for the new kinds (nothing depletes/clears them).

## Manifest (internal/world)

| Field | Change |
|---|---|
| `format_version` | constant 4 → **5**; `Open` keeps rejecting mismatches with migrate hint |
| `terrain_gen` | **new**, int, `omitempty`. Absent/0 ⇒ legacy generation (bit-identical to today). `2` ⇒ marsh+sand generation. `promptworld new` writes 2; `migrate` 4→5 leaves it absent |

Generation entry point gains the version: legacy callers keep exact current output;
`terrain_gen`-aware callers pass the manifest value. Determinism contract extends to
(seed, w, h, **gen**).

## State transitions

None — marsh/sand have no lifecycle (no clear/harvest/quarry/regrow), no structure
interactions, no resource keys in mental maps or policy.

## Agent-facing naming (sim/memory.go)

`featureDesc` gains: Marsh → "the marsh", Sand → "the sand flat". Grass stays "" —
ordinary ground. Every other `switch` over terrain kinds (executor scan, miracles target
validation, memory salience) is swept for exhaustiveness; new kinds behave as open ground
wherever grass does, but are never described by a wrong/fallback label (FR-008).
