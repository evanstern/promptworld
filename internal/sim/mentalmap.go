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
	// Peers are last-seen sightings of the other villagers (spec 041 US1,
	// T013): where the agent last saw each of them, keyed and sorted by agent
	// index. Maintained DERIVATIONALLY (research D2's explored-bit class):
	// villagers cross each other's sight constantly, so event-carrying every
	// sighting would flood the log the way per-step explored events would —
	// the sightings are high-frequency and meaningless individually, and the
	// narratively-loud encounters already ride memories and rumors. Updated by
	// notePresence from the position-changing and waking reducer arms; only
	// ever upserted (a sighting is never forgotten, only superseded).
	Peers []PeerSighting `json:"peers,omitempty"`
}

// PeerSighting is one remembered villager position: agent index, the tile the
// agent was last seen on, and the tick of that sighting. talk_to/seek resolve
// against this — never against live coordinates (spec 041 FR/US1).
type PeerSighting struct {
	Agent int   `json:"agent"`
	X     int   `json:"x"`
	Y     int   `json:"y"`
	Seen  int64 `json:"seen"`
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

// FrontierDirection returns the compass phrase ("north-east") for the
// dominant direction of the NEAREST unexplored terrain from (fromX, fromY),
// and false when the whole map is explored — the prompt's one-line
// orientation toward the unknown (spec 041 US2, contracts §3). Nearest by
// Manhattan distance, ties broken row-major (deterministic); a clearly
// dominated axis component (less than half the other) is dropped so a mostly-
// eastward frontier reads "east", not "north-east". Exported for
// internal/mind.
func (mm *MentalMap) FrontierDirection(w, h, fromX, fromY int) (string, bool) {
	bits := exploredBytes(mm.Explored, w, h)
	best, bx, by := -1, 0, 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			if bits[i/8]&(1<<(i%8)) != 0 {
				continue
			}
			if d := abs(x-fromX) + abs(y-fromY); best < 0 || d < best {
				best, bx, by = d, x, y
			}
		}
	}
	if best < 0 {
		return "", false // fully explored
	}
	dx, dy := bx-fromX, by-fromY
	if 2*abs(dx) < abs(dy) {
		dx = 0
	}
	if 2*abs(dy) < abs(dx) {
		dy = 0
	}
	var ns, ew string
	switch {
	case dy < 0:
		ns = "north"
	case dy > 0:
		ns = "south"
	}
	switch {
	case dx > 0:
		ew = "east"
	case dx < 0:
		ew = "west"
	}
	switch {
	case ns != "" && ew != "":
		return ns + "-" + ew, true
	case ns != "":
		return ns, true
	case ew != "":
		return ew, true
	}
	// The standing tile itself is unexplored (a bare test map): no bearing.
	return "all around", true
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

// --- knowledge predicates (spec 041 US1, research D3) -----------------------
//
// The resolver/reflex gate: candidates must hold a FRESH fact of the right
// kind at their position in the ACTING agent's map. Standalone funcs over
// *Agent so a nil map (a dead migrated native; bare test agents) uniformly
// means "knows nothing".

// knownFreshFact returns the agent's fresh fact of kind at (x, y), if held.
func knownFreshFact(a *Agent, kind string, x, y int, now int64) (PlaceFact, bool) {
	if a.Map == nil {
		return PlaceFact{}, false
	}
	f, ok := a.Map.factAt(kind, x, y)
	if !ok || !factFresh(f, now) {
		return PlaceFact{}, false
	}
	return f, true
}

// knownFactAt reports a fresh fact of kind at (x, y) — the per-tile match
// closure the gated resolvers include (O(log facts), BFS-safe).
func knownFactAt(a *Agent, kind string, x, y int, now int64) bool {
	_, ok := knownFreshFact(a, kind, x, y, now)
	return ok
}

// knowsAnyFresh reports whether the agent holds ANY fresh fact of kind — the
// knowledge-emptiness test behind the "you know of no <kind>" rejection
// (contracts §4), checked BEFORE reachability so the two failure classes stay
// distinct: knowing none is an epistemic failure; knowing some but reaching
// none keeps the existing "no <kind> reachable" phrasing.
func knowsAnyFresh(a *Agent, kind string, now int64) bool {
	if a.Map == nil {
		return false
	}
	for _, f := range a.Map.Facts {
		if f.Kind == kind && factFresh(f, now) {
			return true
		}
	}
	return false
}

// knowsLitFire reports whether the agent remembers ANY fresh fire whose
// remembered Detail (the FuelUntil as last seen) is still ahead of now — the
// cook resolver's knowledge-emptiness test (spec 041 US1).
func knowsLitFire(a *Agent, now int64) bool {
	if a.Map == nil {
		return false
	}
	for _, f := range a.Map.Facts {
		if f.Kind == "fire" && factFresh(f, now) && f.Detail > now {
			return true
		}
	}
	return false
}

// warmKnownPredicate is the knowledge twin of warmAt: warmth the agent can
// PLAN on — within fireWarmRadius of a fire it remembers as lit (remembered
// Detail, the FuelUntil as last seen, still ahead of now — the agent can
// predict burnout from its own knowledge), or on a shelter tile it knows.
// Returns the per-tile predicate plus whether any warm place is known at all
// (the "you know of no warm place" emptiness test). The known fires/shelters
// are captured once, so the predicate is O(known warm places) per tile.
func warmKnownPredicate(a *Agent, now int64) (func(x, y int) bool, bool) {
	if a.Map == nil {
		return func(int, int) bool { return false }, false
	}
	var fires, shelters []PlaceFact
	for _, f := range a.Map.Facts {
		switch {
		case f.Kind == "fire" && factFresh(f, now) && f.Detail > now:
			fires = append(fires, f)
		case f.Kind == "shelter" && factFresh(f, now):
			shelters = append(shelters, f)
		}
	}
	pred := func(x, y int) bool {
		for _, f := range fires {
			if abs(f.X-x)+abs(f.Y-y) <= fireWarmRadius {
				return true
			}
		}
		for _, f := range shelters {
			if f.X == x && f.Y == y {
				return true
			}
		}
		return false
	}
	return pred, len(fires)+len(shelters) > 0
}

// --- peer sightings (spec 041 US1, T013) -------------------------------------

// peerSightingOf returns the agent's last sighting of peer, if any.
func peerSightingOf(a *Agent, peer int) (PeerSighting, bool) {
	if a.Map == nil {
		return PeerSighting{}, false
	}
	i := sort.Search(len(a.Map.Peers), func(j int) bool { return a.Map.Peers[j].Agent >= peer })
	if i < len(a.Map.Peers) && a.Map.Peers[i].Agent == peer {
		return a.Map.Peers[i], true
	}
	return PeerSighting{}, false
}

// sightPeer upserts a sighting, keeping Peers sorted by agent index — the
// canonical-bytes invariant (at most one sighting per peer).
func (mm *MentalMap) sightPeer(peer, x, y int, seen int64) {
	i := sort.Search(len(mm.Peers), func(j int) bool { return mm.Peers[j].Agent >= peer })
	if i < len(mm.Peers) && mm.Peers[i].Agent == peer {
		mm.Peers[i] = PeerSighting{Agent: peer, X: x, Y: y, Seen: seen}
		return
	}
	mm.Peers = append(mm.Peers, PeerSighting{})
	copy(mm.Peers[i+1:], mm.Peers[i:])
	mm.Peers[i] = PeerSighting{Agent: peer, X: x, Y: y, Seen: seen}
}

// notePresence is the derived peer-sighting bookkeeping (D2's explored-bit
// class — silent, no event, pure function of (state, event)): when agent idx
// arrives somewhere — a walked step, a teleport, or waking up — it records
// sightings of every living agent within the witness radius, and every awake
// living agent within that radius records it back. Sleepers neither see nor
// are updated (they look around on agent.woke, which also routes here).
func (s *State) notePresence(idx int, tick int64) {
	a := &s.Agents[idx]
	for j := range s.Agents {
		if j == idx {
			continue
		}
		b := &s.Agents[j]
		if b.Dead || abs(a.X-b.X)+abs(a.Y-b.Y) > witnessRadius {
			continue
		}
		if !a.Asleep && a.Map != nil {
			a.Map.sightPeer(j, b.X, b.Y, tick)
		}
		if !b.Asleep && b.Map != nil {
			b.Map.sightPeer(idx, a.X, a.Y, tick)
		}
	}
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

// MapCorrectedPayload — agent.map_corrected (spec 041 US3, contracts §1): the
// perception sweep found remembered facts ABSENT from ground truth. Gone
// carries the facts AS REMEMBERED (verbatim from the agent's map, canonical
// order — context baked at emission for narration, never re-derived). The
// reducer removes them and stamps a situated witness memory per fact. Absorb
// trigger: the planner re-arms when a removed fact matches the agent's
// current intent target (mind.go).
type MapCorrectedPayload struct {
	Agent int         `json:"agent"`
	Gone  []PlaceFact `json:"gone"`
}
