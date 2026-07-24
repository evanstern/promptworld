package sim

// Per-agent mental maps (spec 041): each villager's PRIVATE spatial knowledge —
// an explored-terrain bitmap plus a list of known place-facts with provenance
// and last-seen ticks. Reducer-owned world state riding Agent.Map: facts are
// mutated ONLY by recorded knowledge events (agent.saw and, in later phases,
// agent.map_corrected / social.place_told / metatron.place_revealed), while
// explored bits are silent derived bookkeeping inside existing reducer arms
// (research D2 — the relation/idle-stamp precedent). Everything here is a pure
// function of its inputs so live and replay agree byte-for-byte.

import (
	"encoding/base64"
	"sort"
)

// MentalMap is one agent's private spatial knowledge (spec 041, research D1):
// a base64 W×H explored bitset (row-major, bit set = terrain shape known) and
// the known dynamic entities as place-facts. Facts are kept sorted by
// (Kind, X, Y) at every mutation so the canonical JSON bytes are independent
// of discovery order (upsert = binary-search insert/replace). Explored bits
// only ever set — knowledge of terrain shape never un-explores; facts are what
// go stale (read-time horizon) or get removed (correction, US3).
type MentalMap struct {
	Explored string      `json:"explored"`
	Facts    []PlaceFact `json:"facts,omitempty"`
}

// PlaceFact is one known dynamic entity. Kind is a closed vocabulary: the
// structure kinds (fire/shelter/oven/chest/wall_plank/wall_stone/path) plus
// the resource kinds (tree/forage/rock/water_edge/den/pile). Seen is the game
// tick the fact was last perceived by the ORIGINAL observer — talk transfer
// (US5) copies the teller's value, so secondhand is never fresher. Provenance
// reuses the Belief vocabulary (witnessed/told, plus revealed for divine
// grants). Source is the teller's agent index, meaningful ONLY for told facts
// (omitempty: a zero value round-trips as agent 0, which is only ever read
// under told provenance — deviation from data-model.md's -1 sentinel, which
// cannot ride an omitempty int; recorded for the planning tier). Detail is a
// kind-specific scalar baked at emission and never re-derived (fires: the
// FuelUntil as last seen; every other kind 0).
type PlaceFact struct {
	Kind       string `json:"kind"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
	Seen       int64  `json:"seen"`
	Provenance string `json:"prov"`
	Source     int    `json:"src,omitempty"`
	Detail     int64  `json:"detail,omitempty"`
}

// ProvenanceRevealed marks a place-fact granted by a Metatron vision (spec 041
// FR-014) — the third provenance beside the Belief vocabulary's witnessed/told
// (consolidate.go), which place-facts reuse.
const ProvenanceRevealed = "revealed"

// Freshness horizons (spec 041, research D6): a fact is fresh iff
// now − Seen < horizon for its kind. Staleness is evaluated at READ time only
// (resolver predicates, prompt rendering) — time never mutates a fact, so
// snapshots stay churn-free. Volatile kinds (a fire's lit-state, a ground
// pile's very existence) decay to unknown within a game-day fraction; durable
// kinds (buildings, terrain resources, dens) hold for several game days.
// Exact values are tuning (clarify Q5), soak-validated in T035.
const (
	factHorizonVolatileTicks = 12 * 3600 // fires, piles: ~12 game-hours
	factHorizonDurableTicks  = 4 * 86400 // structures/resources: 4 game-days
)

// factHorizon is the per-kind freshness window.
func factHorizon(kind string) int64 {
	switch kind {
	case "fire", "pile":
		return factHorizonVolatileTicks
	}
	return factHorizonDurableTicks
}

// factFresh reports whether a fact still satisfies read paths at now — stale
// facts are invisible to resolvers and prompts but stay stored until a
// correction (or death) physically removes them.
func factFresh(f PlaceFact, now int64) bool {
	return now-f.Seen < factHorizon(f.Kind)
}

// Fresh is factFresh exported for internal/mind (US2 prompt rendering reads
// the plan-time replica's maps through the same read-time horizon the
// resolvers use — one freshness rule, one source).
func (f PlaceFact) Fresh(now int64) bool { return factFresh(f, now) }

// newMentalMap returns an empty map sized for a w×h world: all-zero explored
// bits, no facts. Genesis and migration are the only callers — an agent's map
// is created exactly once and only grows.
func newMentalMap(w, h int) *MentalMap {
	return &MentalMap{Explored: base64.StdEncoding.EncodeToString(make([]byte, (w*h+7)/8))}
}

// exploredBytes decodes the bitmap, sized (grown, never shrunk) to cover a
// w×h world. A zero-value or undersized Explored (hand-built test maps, a
// future map-dimension change) yields a correctly-sized bitset with any
// existing bits preserved; a corrupt encoding decodes as all-unexplored
// rather than erroring — the reducer stays total.
func exploredBytes(encoded string, w, h int) []byte {
	bits, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		bits = nil
	}
	if n := (w*h + 7) / 8; len(bits) < n {
		grown := make([]byte, n)
		copy(grown, bits)
		bits = grown
	}
	return bits
}

// ExploredAt reports whether the agent knows the terrain at (x, y) on a w×h
// world. Out-of-bounds tiles are never explored. Bit layout: index y*w+x,
// row-major, LSB-first within each byte.
func (mm *MentalMap) ExploredAt(w, h, x, y int) bool {
	if x < 0 || y < 0 || x >= w || y >= h {
		return false
	}
	bits := exploredBytes(mm.Explored, w, h)
	i := y*w + x
	return bits[i/8]&(1<<(i%8)) != 0
}

// MarkExplored sets the explored bits within Manhattan distance radius of
// (cx, cy), clipped to the w×h map bounds. Monotone by construction (bits are
// only ever OR-ed in), so replay order of position changes cannot matter.
func (mm *MentalMap) MarkExplored(w, h, cx, cy, radius int) {
	bits := exploredBytes(mm.Explored, w, h)
	for dy := -radius; dy <= radius; dy++ {
		y := cy + dy
		if y < 0 || y >= h {
			continue
		}
		r := radius - abs(dy)
		for dx := -r; dx <= r; dx++ {
			x := cx + dx
			if x < 0 || x >= w {
				continue
			}
			i := y*w + x
			bits[i/8] |= 1 << (i % 8)
		}
	}
	mm.Explored = base64.StdEncoding.EncodeToString(bits)
}

// markExplored is the derived explored-bit bookkeeping hook (research D2):
// position-changing reducer arms call it for the mover — silent, no event, a
// pure function of (state, event) so replay is identical. The perception
// radius is the shared witnessRadius. A map-less agent (dead at migration
// time on a pre-041 world) or a map-less State (bare test states that never
// SetMap) skips, matching the miracle arms' existing s.m dependency.
func (s *State) markExplored(a *Agent, x, y int) {
	if a.Map == nil || s.m == nil {
		return
	}
	a.Map.MarkExplored(s.m.W, s.m.H, x, y, witnessRadius)
}

// factLess is the canonical (Kind, X, Y) fact order — the single comparator
// the sorted Facts invariant, the binary-search upsert, and every emitted
// fact batch share.
func factLess(a, b PlaceFact) bool {
	if a.Kind != b.Kind {
		return a.Kind < b.Kind
	}
	if a.X != b.X {
		return a.X < b.X
	}
	return a.Y < b.Y
}

// sortFacts orders a fact batch canonically ((Kind, X, Y)) — emitters sort
// before baking a payload so event bytes are independent of gather order.
func sortFacts(facts []PlaceFact) {
	sort.Slice(facts, func(i, j int) bool { return factLess(facts[i], facts[j]) })
}

// factIndex binary-searches the sorted Facts for (kind, x, y): the position
// holding it (found true) or the insertion point preserving order.
func (mm *MentalMap) factIndex(kind string, x, y int) (int, bool) {
	probe := PlaceFact{Kind: kind, X: x, Y: y}
	i := sort.Search(len(mm.Facts), func(j int) bool { return !factLess(mm.Facts[j], probe) })
	if i < len(mm.Facts) {
		if f := mm.Facts[i]; f.Kind == kind && f.X == x && f.Y == y {
			return i, true
		}
	}
	return i, false
}

// factAt returns the stored fact at (kind, x, y), if known (fresh or stale).
func (mm *MentalMap) factAt(kind string, x, y int) (PlaceFact, bool) {
	if i, ok := mm.factIndex(kind, x, y); ok {
		return mm.Facts[i], true
	}
	return PlaceFact{}, false
}

// upsertFact inserts or replaces the fact at its (Kind, X, Y) slot, keeping
// Facts sorted — at most one fact per (Kind, X, Y), the reducer invariant.
func (mm *MentalMap) upsertFact(f PlaceFact) {
	i, ok := mm.factIndex(f.Kind, f.X, f.Y)
	if ok {
		mm.Facts[i] = f
		return
	}
	mm.Facts = append(mm.Facts, PlaceFact{})
	copy(mm.Facts[i+1:], mm.Facts[i:])
	mm.Facts[i] = f
}

// removeFact deletes the fact at (kind, x, y), preserving order — the
// agent.map_corrected reducer arm's primitive (US3). A drained list goes nil
// so the omitempty canonical bytes match a never-knew map.
func (mm *MentalMap) removeFact(kind string, x, y int) {
	i, ok := mm.factIndex(kind, x, y)
	if !ok {
		return
	}
	mm.Facts = append(mm.Facts[:i], mm.Facts[i+1:]...)
	if len(mm.Facts) == 0 {
		mm.Facts = nil
	}
}

// KnownFresh returns the agent's fresh facts of a kind at now, in canonical
// order — the resolver predicates' and prompt renderer's read surface.
// Exported for internal/mind (US2).
func (mm *MentalMap) KnownFresh(kind string, now int64) []PlaceFact {
	var out []PlaceFact
	for _, f := range mm.Facts {
		if f.Kind == kind && factFresh(f, now) {
			out = append(out, f)
		}
	}
	return out
}

// SawPayload — agent.saw (spec 041, contracts §1): the perception sweep's
// per-agent diff of ground truth against the agent's map. Facts carries the
// NEW or CHANGED facts only, fully baked at emission (Seen = event tick,
// Provenance witnessed, Detail as perceived) and sorted (Kind, X, Y); the
// reducer upserts them verbatim. Digest-only — no chronicle line, not an
// absorb trigger.
type SawPayload struct {
	Agent int         `json:"agent"`
	Facts []PlaceFact `json:"facts"`
}
