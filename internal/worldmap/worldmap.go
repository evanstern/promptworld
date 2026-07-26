// Package worldmap generates and represents the village terrain: a pure
// function of (seed, width, height, generation version), so the map is
// never persisted — every process that knows the manifest regenerates it
// identically. Terrain is a
// flat tile slice (index y*W+x), the representation that scales to DF-style
// sizes later; only dynamic things (buildings, when they exist) will be
// event-sourced on top.
package worldmap

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
)

type TileKind uint8

// There is deliberately no structure/building tile kind: worlds start cold
// (Minecraft-style), and structures arrive later as event-sourced state
// layered over the terrain, never as generated terrain.
const (
	Grass TileKind = iota
	Water
	Tree
	Forage
	Rock // spec 012: rock outcrops, quarryable stone (impassable while standing)
	// Depleted is an effective-kind-only value (internal/sim/terrain.go): a
	// quarried-out rock outcrop. It never appears in a generated Map — it is
	// produced only by the sim package's overlay merge, to mark ground that
	// is passable but NOT buildable and NOT quarryable again (distinct from
	// Grass, which both Cleared trees and Harvested forage revert to).
	Depleted
	// Marsh and Sand (spec 068): walkable ground covers generated only for
	// GenMarshSand worlds — marsh is the moist shoreline ground, sand the
	// drier shoreline ring. Cosmetic-plus-naming only: passable like grass,
	// never buildable, no resource affordances, no overlays. APPENDED after
	// Depleted so every pre-existing kind's byte value (and with it every
	// legacy Map.Hash() stream) stays frozen.
	Marsh
	Sand
)

// DefaultSize is the v1 village area; the representation itself is
// size-agnostic.
const DefaultSize = 64

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Map struct {
	W, H  int
	Tiles []TileKind
	// Dens are animal home sites (huntable wildlife territory); the animal
	// entities themselves are TASK-5's.
	Dens []Point
}

func (m *Map) At(x, y int) TileKind {
	return m.Tiles[y*m.W+x]
}

func (m *Map) InBounds(x, y int) bool {
	return x >= 0 && x < m.W && y >= 0 && y < m.H
}

// Passable is walkable terrain: grass, forage, and the spec 068 ground
// covers (marsh, sand). Water, standing trees, and rock outcrops block
// movement.
func (m *Map) Passable(x, y int) bool {
	if !m.InBounds(x, y) {
		return false
	}
	k := m.At(x, y)
	return k == Grass || k == Forage || k == Marsh || k == Sand
}

// Buildable sites are plain grass: flat, dry, unforested (AC#2 — worlds
// start with no structures but room to build them).
func (m *Map) Buildable(x, y int) bool {
	return m.InBounds(x, y) && m.At(x, y) == Grass
}

func (m *Map) CountKind(k TileKind) int {
	n := 0
	for _, t := range m.Tiles {
		if t == k {
			n++
		}
	}
	return n
}

// Hash fingerprints the full terrain + dens for determinism checks (AC#3).
func (m *Map) Hash() string {
	h := sha256.New()
	h.Write([]byte{byte(m.W), byte(m.W >> 8), byte(m.H), byte(m.H >> 8)})
	for _, t := range m.Tiles {
		h.Write([]byte{byte(t)})
	}
	for _, d := range m.Dens {
		h.Write([]byte{byte(d.X), byte(d.X >> 8), byte(d.Y), byte(d.Y >> 8)})
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Generation tuning: fractions of the map, chosen so a 64×64 village has a
// real lake, real woods, and plenty of open ground.
const (
	waterFraction  = 0.18 // lowest-lying land floods
	treeFraction   = 0.24 // moistest dry land forests
	rockFraction   = 0.06 // highest-elevation remaining dry grass, after trees (spec 012 research R1)
	foragePerMille = 45   // ~4.5% of open grass carries forage
	// marshFraction (spec 068 R4, GenMarshSand only): the moistest fraction
	// of shoreline grass — open grass 4-adjacent to water, after every
	// existing pass — becomes marsh; the remaining shoreline grass becomes
	// sand. Only open grass converts, so the water/tree/rock/forage
	// fractions are untouched by the new kinds.
	marshFraction  = 0.4
	denCount       = 4
	denMinDistance = 12
)

// Terrain generation versions (spec 068 FR-006/FR-007): the manifest's
// terrain_gen field selects the algorithm, so a world's terrain is a pure
// function of (seed, w, h, gen) forever. GenLegacy (0, the absent-field
// default) is the pre-068 algorithm, bit-identical — TestLegacyGenerationHashPin
// is its gate. GenMarshSand (2, what `promptworld new` writes) adds the
// marsh/sand shoreline pass. There is deliberately no value 1: the field
// went straight to 2 so absent/0 reads unambiguously as "legacy".
const (
	GenLegacy    = 0
	GenMarshSand = 2
)

// Generate is the legacy generation path, deterministic: same (seed, w, h)
// → identical Map, on every platform (integer/hash noise only, no float
// randomness sources). Kept as the no-version signature so every legacy
// caller keeps exactly today's output; version-aware callers (world.Map)
// use GenerateV.
func Generate(seed uint64, w, h int) *Map {
	return GenerateV(seed, w, h, GenLegacy)
}

// GenerateV generates terrain under a manifest-selected generation version
// (spec 068 R5). gen == GenMarshSand adds the marsh/sand shoreline pass;
// any other value — including future versions a newer manifest might carry,
// which world.Open refuses before terrain is ever generated — takes the
// legacy path.
func GenerateV(seed uint64, w, h int, gen int) *Map {
	if w <= 0 {
		w = DefaultSize
	}
	if h <= 0 {
		h = DefaultSize
	}
	m := &Map{W: w, H: h, Tiles: make([]TileKind, w*h)}

	// Two independent fractal noise fields: elevation and moisture.
	height := make([]float64, w*h)
	moist := make([]float64, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			height[i] = fbm(seed, "elevation", x, y)
			moist[i] = fbm(seed, "moisture", x, y)
		}
	}

	// Water: flood everything below the waterFraction percentile of height.
	waterLine := percentile(height, waterFraction)
	for i, hv := range height {
		if hv <= waterLine {
			m.Tiles[i] = Water
		}
	}

	// Trees: the moistest dry tiles, as patches (moisture is spatially
	// correlated noise, so thresholding yields woods, not salt-and-pepper).
	var dryMoist []float64
	for i, t := range m.Tiles {
		if t != Water {
			dryMoist = append(dryMoist, moist[i])
		}
	}
	treeLine := percentileTop(dryMoist, treeFraction)
	for i, t := range m.Tiles {
		if t == Grass && moist[i] >= treeLine {
			m.Tiles[i] = Tree
		}
	}

	// Rock outcrops: the highest-elevation ~6% of dry grass remaining after
	// trees (spec 012 research R1), scored by the existing elevation field
	// plus a small hash-jitter (purpose "rock") so patches get a coherent-but-
	// textured edge instead of a smooth ridge line. Reuses the elevation
	// field rather than adding a new noise pass — correlated noise already
	// yields patches, exactly like trees claiming the moistest fraction.
	var dryHeight []float64
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			if m.Tiles[i] == Grass {
				dryHeight = append(dryHeight, height[i]+rockJitter(seed, x, y))
			}
		}
	}
	rockLine := percentileTop(dryHeight, rockFraction)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			if m.Tiles[i] == Grass && height[i]+rockJitter(seed, x, y) >= rockLine {
				m.Tiles[i] = Rock
			}
		}
	}

	// Forage: scattered over remaining grass by per-tile hash.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			if m.Tiles[i] == Grass && hash2(seed, "forage", x, y)%1000 < foragePerMille {
				m.Tiles[i] = Forage
			}
		}
	}

	// Marsh + sand (spec 068 R4, GenMarshSand only): the shoreline pass runs
	// AFTER every legacy pass — it claims only open grass 4-adjacent to
	// water, so trees, rocks, and forage keep their exact legacy placement —
	// and BEFORE dens, so a den never sits on the new ground covers. The
	// moistest marshFraction of the shoreline candidates (the same moisture
	// field the tree pass thresholds) becomes marsh — low-lying wet ground —
	// and the drier remainder becomes sand, the shoreline ring. Marsh wins
	// where both apply (it takes the top of the percentile split).
	if gen == GenMarshSand {
		waterAdjacent := func(x, y int) bool {
			for _, d := range [4][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}} {
				nx, ny := x+d[0], y+d[1]
				if nx >= 0 && nx < w && ny >= 0 && ny < h && m.Tiles[ny*w+nx] == Water {
					return true
				}
			}
			return false
		}
		var shoreMoist []float64
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if m.Tiles[y*w+x] == Grass && waterAdjacent(x, y) {
					shoreMoist = append(shoreMoist, moist[y*w+x])
				}
			}
		}
		if len(shoreMoist) > 0 {
			marshLine := percentileTop(shoreMoist, marshFraction)
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					i := y*w + x
					if m.Tiles[i] != Grass || !waterAdjacent(x, y) {
						continue
					}
					if moist[i] >= marshLine {
						m.Tiles[i] = Marsh
					} else {
						m.Tiles[i] = Sand
					}
				}
			}
		}
	}

	// Animal dens: deterministic candidate stream; grass only, spread out.
	for n := 0; len(m.Dens) < denCount && n < 10_000; n++ {
		x := int(hash2(seed, "den-x", n, 0) % uint64(w))
		y := int(hash2(seed, "den-y", n, 0) % uint64(h))
		if m.At(x, y) != Grass {
			continue
		}
		ok := true
		for _, d := range m.Dens {
			if abs(d.X-x)+abs(d.Y-y) < denMinDistance {
				ok = false
				break
			}
		}
		if ok {
			m.Dens = append(m.Dens, Point{X: x, Y: y})
		}
	}

	return m
}

// rockJitter is a small deterministic nudge (purpose "rock") added to the
// elevation score before the rock-outcrop percentile cut: elevation stays
// the dominant, spatially-correlated signal (coherent patches), while the
// jitter roughens the patch boundary and breaks exact-tie ordering.
func rockJitter(seed uint64, x, y int) float64 {
	return float64(hash2(seed, "rock", x, y)%1000) / 1000 * 0.05
}

func percentile(values []float64, frac float64) float64 {
	s := append([]float64(nil), values...)
	sort.Float64s(s)
	idx := int(frac * float64(len(s)))
	if idx >= len(s) {
		idx = len(s) - 1
	}
	return s[idx]
}

// percentileTop returns the threshold above which frac of values lie.
func percentileTop(values []float64, frac float64) float64 {
	return percentile(values, 1-frac)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
