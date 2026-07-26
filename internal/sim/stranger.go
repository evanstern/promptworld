package sim

import (
	"encoding/json"
	"fmt"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// The stranger (spec 077 US2, research R4): a nocturnal trickster that slips
// in at an authored entry tile, moves toward unattended stores, takes bounded
// goods, and is gone by dawn. Like the gru it is an ENTITY — a positioned
// body in event-sourced state, not a phenomenon: sight needs geometry, the
// TUI needs something to render, rumors need something to have been seen, and
// the "nothing is taken" rubric terms need a durable ledger to count.
//
// Safety rules are the gru's, shared not duplicated: fire light and shelter
// tiles are absolute — the stranger never steps into a protected tile
// (gruProtected), so goods kept inside the light are structurally safe.
// It takes, it never harms: no attack arm exists, and the whole-line alert
// tier it touches is theft's (stranger.took beside social.chest_taken).
//
// Determinism: every random decision is per-decision seeded (rngAt purposes
// "stranger-prowl"; deterministic scans everywhere else) — no stream, per
// deterministic-rng's standing instruction. All effects are events through
// the reducer; genesis replay re-applies them with no scenario armed.

// Stranger is the trickster's event-sourced state; nil means it is not
// abroad. Night is the 1-based arrival night (history/identity). LastMove/
// LastTake are cadence anchors stamped by the reducer arms (the
// Gru.LastAttack shape) — both SHIFT across a time snap (miracles.go).
type Stranger struct {
	X        int   `json:"x"`
	Y        int   `json:"y"`
	Night    int64 `json:"night"`
	LastMove int64 `json:"last_move,omitempty"`
	LastTake int64 `json:"last_take,omitempty"`
}

// StrangerTake is one ledger record of a take — the durable trace zero-wanted
// rubric terms count ("nothing is taken" ⇔ an empty ledger). Tick is a
// historical fact (KEEP across a time snap, the DeathRecord.Tick shape).
type StrangerTake struct {
	Tick int64  `json:"tick"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
	Kind string `json:"kind"`
	N    int    `json:"n"`
}

const (
	// strangerTakeRetain bounds the take ledger (the guardianOrderRetain
	// precedent): the most recent 32 takes keep the trail readable without
	// unbounded state growth; rubric counting only ever needs "zero or not".
	strangerTakeRetain = 32
	// strangerMoveEveryTicks is the movement cadence — the gru's pace
	// (gruMoveEveryTicks), anchored on LastMove rather than a modulo so the
	// cadence survives a time snap (LastMove SHIFTs).
	strangerMoveEveryTicks = 4
	// strangerTakeCooldownTicks spaces takes (the gruAttackCooldown shape):
	// one bounded take per 10 game-minutes while it stands on an unattended
	// store.
	strangerTakeCooldownTicks = 600
	// strangerTakeMax bounds one take's quantity — CONTENT, like the rubric
	// thresholds: it nibbles, it does not empty the larder in one grab.
	strangerTakeMax = 2

	salStrangerTheft = 7 // a witnessed theft — rumor fuel (salGruWitness's tier)
)

// strangerLootKinds is the fixed order a take drains a store in — the
// canonical goods vocabulary minus the durability-carrying tools (spears/
// axes stay: a trickster pockets goods, not armaments, and the flat
// Kind/N ledger shape has no durability slot). Determinism depends on this
// order exactly as the rot sweep depends on foodKinds.
var strangerLootKinds = []string{
	"wood", "stone", "water", "planks", "refined_stone",
	"food_raw", "food_cooked", "meals",
}

// strangerStep is the trickster's whole turn, called from stepEvents after
// gruStep and before the governance/social beats (order pinned by test).
// Pure over (pre-tick state, map, next tick) like everything else in the
// executor; an ambient world (no stranger ever arrived) returns nil on its
// first check — byte-identical pre-077 behavior (spec 077 FR-017).
func strangerStep(s *State, m *worldmap.Map, night bool, nextTick int64) []store.Event {
	if s.Stranger == nil {
		return nil
	}
	var events []store.Event
	emit := func(typ string, payload any) {
		events = append(events, store.Event{Tick: nextTick, Type: typ, Payload: mustPayload(payload)})
	}

	if !night {
		// Gone by dawn (research R4) — the gru.withdrew shape.
		day, _, _, _ := clock.GameTime(nextTick)
		emit("stranger.departed", StrangerDepartedPayload{Day: day})
		return events
	}

	st := s.Stranger

	// Take: standing on an unattended store with goods, claws — hands — ready.
	if kind, n := strangerLootAt(s, st.X, st.Y); n > 0 && strangerStoreUnattended(s, st.X, st.Y) &&
		(st.LastTake == 0 || nextTick-st.LastTake >= strangerTakeCooldownTicks) {
		emit("stranger.took", StrangerTookPayload{X: st.X, Y: st.Y, Kind: kind, N: n})
		// A seen theft is village rumor fuel (research R4): every awake
		// villager near enough marks the moment — the gru's witness idiom.
		for w := range s.Agents {
			wa := &s.Agents[w]
			if wa.Dead || wa.Asleep {
				continue
			}
			if abs(wa.X-st.X)+abs(wa.Y-st.Y) <= witnessRadius {
				events = append(events, situatedMemoryEvent(nextTick, w, salStrangerTheft,
					PlaceAt(s, wa.X, wa.Y), "", OriginWitness, "Saw a stranger take from our stores in the dark."))
			}
		}
		return events
	}

	// Move: cadence-gated (LastMove anchor — the snap-safe gru pace).
	if st.LastMove != 0 && nextTick-st.LastMove < strangerMoveEveryTicks {
		return events
	}
	if tx, ty, ok := strangerNearestStore(s, st.X, st.Y); ok && (tx != st.X || ty != st.Y) {
		// Stalk the larder: the neighbor that closes the gap, never a
		// protected tile — greedy, not BFS, exactly the gru's hunt (a
		// trickster baffled by firelight is the right trickster).
		bx, by, best := st.X, st.Y, abs(tx-st.X)+abs(ty-st.Y)
		for _, d := range neighborOrder {
			nx, ny := st.X+d[0], st.Y+d[1]
			if !passable(m, s, nx, ny) || gruProtected(s, nx, ny) {
				continue
			}
			if nd := abs(tx-nx) + abs(ty-ny); nd < best {
				bx, by, best = nx, ny, nd
			}
		}
		if bx != st.X || by != st.Y {
			emit("stranger.moved", StrangerMovedPayload{X: bx, Y: by})
			return events
		}
	}
	// Prowl: seeded drift through the dark (per-decision RNG, no stream).
	var open [4][2]int
	n := 0
	for _, d := range neighborOrder {
		nx, ny := st.X+d[0], st.Y+d[1]
		if passable(m, s, nx, ny) && !gruProtected(s, nx, ny) {
			open[n] = [2]int{nx, ny}
			n++
		}
	}
	if n > 0 {
		p := open[rngAt(s.Seed, "stranger-prowl", nextTick, 0).Uint64N(uint64(n))]
		emit("stranger.moved", StrangerMovedPayload{X: p[0], Y: p[1]})
	}
	return events
}

// strangerStoreUnattended reports whether no living villager stands adjacent
// to (or on) the store tile — the "unattended" precondition (research R4):
// a watched larder is safe.
func strangerStoreUnattended(s *State, x, y int) bool {
	for i := range s.Agents {
		a := &s.Agents[i]
		if !a.Dead && abs(a.X-x)+abs(a.Y-y) <= 1 {
			return false
		}
	}
	return true
}

// strangerLootAt picks what one take at (x,y) would carry off: the first
// stocked kind in strangerLootKinds order, capped at strangerTakeMax — from
// the tile's ground pile first, else its chest. (0 when the tile holds no
// takeable goods.) Deterministic: fixed kind order, fixed store precedence.
func strangerLootAt(s *State, x, y int) (string, int) {
	if pile := s.pileAt(x, y); pile != nil {
		for _, kind := range strangerLootKinds {
			if a := pile.avail(kind); a > 0 {
				return kind, minInt(a, strangerTakeMax)
			}
		}
	}
	if ch := s.chestAt(x, y); ch != nil && ch.Store != nil {
		for _, kind := range strangerLootKinds {
			if a := carriedCount(*ch.Store, kind); a > 0 {
				return kind, minInt(a, strangerTakeMax)
			}
		}
	}
	return "", 0
}

// strangerNearestStore finds the nearest unattended, unprotected store tile
// holding takeable goods — ground piles first, then chests, ties to the
// earliest in slice order (deterministic by construction). ok=false when no
// store tempts tonight (the stranger just prowls).
func strangerNearestStore(s *State, x, y int) (tx, ty int, ok bool) {
	best := -1
	consider := func(cx, cy int) {
		if gruProtected(s, cx, cy) || !strangerStoreUnattended(s, cx, cy) {
			return
		}
		if _, n := strangerLootAt(s, cx, cy); n == 0 {
			return
		}
		if d := abs(cx-x) + abs(cy-y); best < 0 || d < best {
			tx, ty, best = cx, cy, d
		}
	}
	for i := range s.Piles {
		consider(s.Piles[i].X, s.Piles[i].Y)
	}
	for i := range s.Structures {
		if s.Structures[i].Kind == "chest" {
			consider(s.Structures[i].X, s.Structures[i].Y)
		}
	}
	return tx, ty, best >= 0
}

type (
	// StrangerArrivedPayload — stranger.arrived: the entity slips in at a
	// passable, unprotected tile. NO authored/scenario marker (spec 077
	// FR-013): an ambient arrival would record exactly these fields.
	StrangerArrivedPayload struct {
		Night int64 `json:"night"`
		X     int   `json:"x"`
		Y     int   `json:"y"`
	}
	StrangerMovedPayload struct {
		X int `json:"x"`
		Y int `json:"y"`
	}
	// StrangerTookPayload — stranger.took: one bounded take from the store
	// at (x,y). Whole-line alert tier in the chronicle (theft is theft —
	// social.chest_taken's tier).
	StrangerTookPayload struct {
		X    int    `json:"x"`
		Y    int    `json:"y"`
		Kind string `json:"kind"`
		N    int    `json:"n"`
	}
	StrangerDepartedPayload struct {
		Day int64 `json:"day"`
	}
)

// applyStranger is the reducer arm for stranger.* events. Reducer-total like
// applyGru: events aimed at a vanished stranger no-op rather than error, and
// the take arm clamps defensively to what the store holds (the
// agent.withdrew posture — replay fidelity over re-validation).
func (s *State) applyStranger(e store.Event) error {
	switch e.Type {
	case "stranger.arrived":
		var p StrangerArrivedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		s.Stranger = &Stranger{X: p.X, Y: p.Y, Night: p.Night}
	case "stranger.moved":
		var p StrangerMovedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		if s.Stranger != nil {
			s.Stranger.X, s.Stranger.Y = p.X, p.Y
			s.Stranger.LastMove = e.Tick
		}
	case "stranger.took":
		var p StrangerTookPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		if p.N <= 0 {
			return fmt.Errorf("apply %s: take of %d %s not positive", e.Type, p.N, p.Kind)
		}
		// Goods leave through the same state shapes agent withdrawal uses:
		// pile first, else chest — clamped to what is actually there.
		if pile := s.pileAt(p.X, p.Y); pile != nil {
			if isFoodKind(p.Kind) {
				pile.takeFood(p.Kind, p.N)
			} else {
				pile.takeNonFood(p.Kind, p.N)
			}
			s.removeEmptyPileAt(p.X, p.Y)
		} else if ch := s.chestAt(p.X, p.Y); ch != nil && ch.Store != nil {
			n := p.N
			if c := carriedCount(*ch.Store, p.Kind); n > c {
				n = c
			}
			if n > 0 {
				addItems(ch.Store, []Item{{Kind: p.Kind, N: n}}, -1)
			}
		}
		// The ledger records the event as recorded (payload N), bounded ring.
		s.StrangerTakes = append(s.StrangerTakes, StrangerTake{
			Tick: e.Tick, X: p.X, Y: p.Y, Kind: p.Kind, N: p.N,
		})
		if drop := len(s.StrangerTakes) - strangerTakeRetain; drop > 0 {
			s.StrangerTakes = append([]StrangerTake(nil), s.StrangerTakes[drop:]...)
		}
		if s.Stranger != nil {
			s.Stranger.LastTake = e.Tick
		}
	case "stranger.departed":
		s.Stranger = nil
	}
	return nil
}
