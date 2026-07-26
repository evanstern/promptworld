---
name: worldmap-generation
description: Seeded village terrain — integer-hash value noise generates water, woods, forage, animal dens, and (since spec 068) a versioned marsh/sand shoreline pass; never persisted, regenerated from the manifest anywhere
kind: component
sources:
  - internal/worldmap/worldmap.go
  - internal/worldmap/noise.go
verified_against: d304e8adb64fdf40e24bfeca3ca3420e8a840a35
---

# World map generation

`internal/worldmap` generates the village terrain as a **pure function of
(seed, width, height, generation version)**. The map is never persisted or sent over
the wire: the daemon, the TUI, and any future tool regenerate the identical map from
the manifest (`world.Map()`). Only dynamic changes (buildings, when TASK-5+ adds
them) will be event-sourced on top of this static base.

## How it works

Representation: `Map{W, H, Tiles []TileKind, Dens []Point}` — a flat slice indexed
`y*W+x`, the shape that scales to DF-style sizes later (the engine requirement from
the grounding session). `TileKind` is `Grass | Water | Tree | Forage | Rock |
Depleted | Marsh | Sand` — `Marsh`/`Sand` (spec 068, TASK-143) are appended AFTER
`Depleted` so every pre-existing kind's byte value, and with it every legacy
`Map.Hash()` stream, stays frozen. **There is deliberately no structure kind**:
worlds start cold (Minecraft-style), so "no structures at seed" holds by
construction, and `Buildable(x,y)` = plain grass (marsh/sand are NOT buildable,
unchanged by spec 068). `Depleted` is an effective-kind-only value: `Generate`
never produces it — it exists only as a value the [[executor]]'s terrain overlay
merge (`internal/sim/terrain.go`) can mark a quarried-out `Rock` tile with, distinct
from `Grass` (which cleared trees and harvested forage both revert to). `Marsh` and
`Sand` are, by contrast, real GENERATED kinds (below) — walkable ground covers with
no resource affordances and no overlay of their own: cosmetic-plus-naming terrain,
like grass but nameable to a villager ([[executor]]'s `featureDesc` names them "the
marsh"/"the sand flat", never the generic "" grass gets).

Noise (`noise.go`): integer-hash value noise — lattice values from FNV-64a of
(seed, purpose, lattice point), smoothstep-bilinear interpolation, summed over three
octaves (cells 16/8/4, halving amplitude) in `fbm`. Pure integer hashing keeps
generation byte-identical across platforms and Go versions, the same discipline as
[[deterministic-rng]].

**Generation versions** (spec 068, FR-006/FR-007): a world's terrain is a pure
function of `(seed, w, h, gen)` forever — the manifest's `terrain_gen` field
([[world-save-directory]]) selects the algorithm. `GenLegacy` (0, the absent-field
default) is the pre-068 algorithm, bit-identical to every world generated before
this feature (`TestLegacyGenerationHashPin` is the gate); `GenMarshSand` (2, what
`promptworld new` writes for every new world) adds the marsh/sand shoreline pass
below. There is deliberately no value 1 — the field jumps straight to 2 so an
absent/zero `terrain_gen` reads unambiguously as "legacy," never "version one of
something." `Generate(seed, w, h)` is kept as the legacy no-version signature (it
calls `GenerateV(seed, w, h, GenLegacy)`) so every existing caller's output is
untouched; version-aware callers — `world.Map()` — call `GenerateV(seed, w, h, gen)`
directly.

`GenerateV` pipeline: elevation and moisture fields (`fbm` with different purpose
tags) → water floods the lowest `waterFraction = 18%` of elevation (percentile
threshold, so every seed gets real water) → trees claim the moistest
`treeFraction = 24%` of dry land (correlated noise ⇒ woods, not salt-and-pepper) →
rock outcrops claim the highest-elevation `rockFraction = 6%` of the dry grass left
after trees (spec 012 research R1) — scored by the same elevation field plus a small
deterministic `rockJitter` hash-nudge (purpose `"rock"`) so patches get a
coherent-but-textured edge rather than a smooth ridge line, reusing the elevation
signal rather than adding a new noise pass, the same idiom as trees claiming the
moistest fraction → forage scatters over remaining grass at ~4.5% (`foragePerMille =
45`) per-tile hash probability → **on `GenMarshSand` only**, the shoreline pass
(below) runs, converting some of the open grass left after forage into marsh/sand →
four (`denCount`) animal `Dens` are picked from a deterministic candidate stream,
grass-only, ≥12 (`denMinDistance`) Manhattan apart — run LAST, after the shoreline
pass, so a den never lands on the new ground covers. Zero dims default to
`DefaultSize = 64`.

**The marsh/sand shoreline pass** (spec 068 R4, `GenMarshSand` only): candidates are
open `Grass` tiles 4-adjacent (`waterAdjacent`, the four cardinal neighbors) to a
`Water` tile — computed AFTER every legacy pass, so trees/rocks/forage keep their
exact legacy placement; only leftover open grass converts. Among those candidates,
the moistest `marshFraction = 0.4` (by the SAME moisture field the tree pass
thresholds, `percentileTop`) become `Marsh` — low-lying wet ground; the drier
remainder of the shoreline candidates become `Sand`, the shoreline ring. A world
with no qualifying shoreline grass (`len(shoreMoist) == 0`) simply gets no marsh or
sand — the pass is a no-op, never an error. Marsh and sand carry no resource
affordances of their own and no overlay ([[executor]]'s `effectiveKind`/`passable`
deliberately have no arm for them — a spec-068 comment in `terrain.go` notes their
kind never changes from what generation produced).

`Passable(x,y)` = in-bounds grass, forage, marsh, or sand (water, standing trees, and
rock outcrops block); `Buildable(x,y)` stays plain grass only — marsh and sand are
walkable but not buildable, exactly like forage. `Hash()` fingerprints tiles + dens
for determinism tests; because `Marsh`/`Sand` are new `TileKind` byte values
appended after every existing one, a legacy (`GenLegacy`) map's `Hash()` is
unaffected by their existence.

## Connections

[[world-save-directory]]'s manifest carries `map_width`/`map_height` and the spec-068
`terrain_gen` field, and `world.Map()` calls `GenerateV` with it; the [[executor]]
overlays dynamic terrain on the generated map (quarrying a `Rock` tile produces the
`Depleted` effective kind; marsh/sand pass through `effectiveKind`/`passable`
unchanged) and moves agents against effective passability; [[sim-state-reducer]]'s
genesis places them on passable tiles; the [[tui-client]] map pane renders tiles and
dens through the [[tile-registry]] (spec 068 adds marsh `░`/sand `▒` rows there).
Dens are huntable food sites with cooldowns as of TASK-5. [[world-migration]]
re-places migrated agents onto a freshly generated map under the world's own
`terrain_gen` (a v4→v5 migrated world's `terrain_gen` stays absent, so it keeps its
legacy terrain, rock outcrops included, exactly as it always has).

## Operational notes

Tuning constants live at the top of `worldmap.go`. Tests assert, across a seed
spread: same seed ⇒ identical `Hash()` (AC#3); water/trees/forage/dens all present
(AC#1); ≥25% buildable open grass (AC#2). Changing any tuning constant or the noise
changes every existing world's terrain on next daemon start — treat generation as
format-versioned behavior once real saves matter; spec 068's own answer to that risk
for the marsh/sand pass specifically is the `terrain_gen` version gate, not a promise
never to add more generation versions later.
