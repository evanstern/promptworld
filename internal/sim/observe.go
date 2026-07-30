package sim

import (
	"sort"
	"strings"

	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// Perception of absence (spec 097, TASK-80): grounded arrival observations.
//
// The sim only ever told agents what happened — never what is NOT there — so a
// confabulated place-belief was unfalsifiable by construction (the Thornspire
// finding, 2026-07-23). This file is the observation channel that gives
// reality a voice: when a walker ARRIVES at the destination its intent chose
// (the step that lands it on the target tile — D1: intent-completing arrivals,
// never every wander step), the executor emits one `agent.place_observed`
// carrying the COMPLETE set of feature/entity kinds actually present within
// placeScanRadius (D2: exhaustive-within-radius — absence is implied by
// exhaustiveness; the reducer cannot know what an agent expected, so there is
// no "absence_of" field). Emission is a pure function of world state at the
// arrival tick, payload fully baked at emission (D5, the spec 092/094
// emitter-computes doctrine; additive event type, no format-version bump), and
// NO model runs anywhere in this path — observation-vs-belief matching is the
// mind's judgment (internal/mind/reconcile.go, D3).
//
// A companion first-person situated memory rides the same batch (memories
// accrete only via agent.memory_added — TestMemoriesAccrete), at the low
// ObservationBaseSalience dial with the new OriginObserved provenance class.
// Repeat observations of an UNCHANGED place within the ObservationDedupTicks
// window collapse entirely — no event, no memory (D4: no working-window
// flooding, and the mind-side reconciliation rate is bounded by the same
// window since it triggers off the event).

// PlaceObservedPayload is the grounded arrival observation (spec 097 FR-001):
// the agent, where it stood, the scan radius, and the exhaustive set of
// feature/entity kinds present within that radius at the arrival tick. Kinds
// is the mental-map fact vocabulary (mentalmap.go: structure kinds as-is,
// "pile", "tree", "forage", "rock", "water_edge", "den"), deduplicated and
// sorted — a deterministic set, not a per-tile listing. An empty Kinds means
// the scan found NOTHING notable: that emptiness is the whole point of the
// channel (perception of absence), so the field is never omitted.
type PlaceObservedPayload struct {
	Agent  AgentRef `json:"agent"`
	X      int      `json:"x"`
	Y      int      `json:"y"`
	Radius int      `json:"radius"`
	Kinds  []string `json:"kinds"`
}

// ObservationMark is the reducer-maintained dedup anchor (D4): the last
// grounded observation this agent recorded — where, what it saw (canonical
// sorted kinds), and when. Written ONLY by the agent.place_observed reducer
// arm; the emission site compares the next arrival against it, so dedup is a
// pure function of event-sourced state (replay-identical by construction).
// Tick is a duration anchor (elapsed = tick − Tick gates the window) — SHIFT
// under a time snap (rebaseTicks, miracles.go), never zero once set.
type ObservationMark struct {
	X     int      `json:"x"`
	Y     int      `json:"y"`
	Kinds []string `json:"kinds,omitempty"`
	Tick  int64    `json:"tick"`
}

// observedKinds scans the tiles within placeScanRadius (Manhattan) of (x,y)
// and returns the deduplicated, sorted set of feature/entity kinds actually
// present — the D2 exhaustive observation. The vocabulary and presence rules
// are the perception sweep's (perceptionEvents): structure kinds as-is,
// ground piles, overlay-aware resource tiles (a cleared tree or harvested
// forage spot is NOT a tree/forage — effectiveKind already says Grass), the
// water shoreline, and dens. Deterministic: fixed scan order plus a final
// sort; pure function of (state, map, x, y).
func observedKinds(s *State, m *worldmap.Map, x, y int) []string {
	set := map[string]bool{}
	for dy := -placeScanRadius; dy <= placeScanRadius; dy++ {
		for dx := -placeScanRadius; dx <= placeScanRadius; dx++ {
			tx, ty := x+dx, y+dy
			if !m.InBounds(tx, ty) || abs(dx)+abs(dy) > placeScanRadius {
				continue
			}
			switch effectiveKind(m, s, tx, ty) {
			case worldmap.Tree:
				set["tree"] = true
			case worldmap.Forage:
				set["forage"] = true
			case worldmap.Rock:
				set["rock"] = true
			case worldmap.Water:
				if waterEdge(m, tx, ty) {
					set["water_edge"] = true
				}
			}
		}
	}
	for _, d := range m.Dens {
		if abs(d.X-x)+abs(d.Y-y) <= placeScanRadius {
			set["den"] = true
		}
	}
	for _, st := range s.Structures {
		if abs(st.X-x)+abs(st.Y-y) <= placeScanRadius {
			set[st.Kind] = true
		}
	}
	for _, p := range s.Piles {
		if abs(p.X-x)+abs(p.Y-y) <= placeScanRadius {
			set["pile"] = true
		}
	}
	kinds := make([]string, 0, len(set))
	for k := range set {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	return kinds
}

// observationDeduped reports whether an observation of (x,y) seeing kinds at
// tick collapses into the agent's last one (D4): the SAME PLACE — within
// placeScanRadius of the last observation, where the two scan discs largely
// overlap — with an identical kind set, inside the dedup window. Radius-based
// rather than exact-tile: an agent pacing adjacent tiles of one clearing is
// revisiting an unchanged place, not discovering new ones (the card's flood
// worry), and a genuinely changed place always re-observes because its kind
// set differs. The anchor never slides on a suppressed observation (no event,
// no anchor move), so drift beyond the radius always observes afresh. kinds
// must be canonical (sorted, deduped) — observedKinds' output.
func observationDeduped(a *Agent, x, y int, kinds []string, tick, window int64) bool {
	lo := a.LastObs
	if lo == nil || abs(lo.X-x)+abs(lo.Y-y) > placeScanRadius || tick-lo.Tick >= window {
		return false
	}
	if len(lo.Kinds) != len(kinds) {
		return false
	}
	for i := range kinds {
		if lo.Kinds[i] != kinds[i] {
			return false
		}
	}
	return true
}

// kindNoun renders one observed kind as the phrase the observation memory
// lists ("a fire", "standing trees"). Closed over the mental-map fact
// vocabulary; an unknown kind falls back to its underscore-split words.
func kindNoun(kind string) string {
	switch kind {
	case "fire":
		return "a fire"
	case "shelter":
		return "a shelter"
	case "oven":
		return "an oven"
	case "chest":
		return "a chest"
	case "wall_plank":
		return "a plank wall"
	case "wall_stone":
		return "a stone wall"
	case "path":
		return "a path"
	case "grave":
		return "a grave"
	case "pile":
		return "goods on the ground"
	case "tree":
		return "standing trees"
	case "forage":
		return "a forage patch"
	case "rock":
		return "a rock outcrop"
	case "water_edge":
		return "the water's edge"
	case "den":
		return "a den"
	}
	return "a " + strings.ReplaceAll(kind, "_", " ")
}

// observedMemoryText composes the observation memory's base text from the
// canonical kind set. First person and terse; the empty set says so plainly —
// "nothing of note" IS the discovery this channel exists to record.
func observedMemoryText(kinds []string) string {
	if len(kinds) == 0 {
		return "Looked around: nothing of note here."
	}
	nouns := make([]string, len(kinds))
	for i, k := range kinds {
		nouns[i] = kindNoun(k)
	}
	return "Looked around: " + strings.Join(nouns, ", ") + "."
}

// placeObservedEvents builds the grounded arrival observation for agent i
// standing at (x,y) — the companion situated memory FIRST, then the
// agent.place_observed event, or nothing at all when the observation dedups
// (D4: both-or-neither, so the memory window, the event stream, and the
// mind-side reconciliation are all bounded by the same window). The memory
// precedes the observation deliberately — a deviation from the map_corrected
// order (event then memories), recorded here: the mind's absorb loop
// reconciles beliefs when the place_observed event lands and reads the
// companion memory (for the D4 surprise bump) off its replica, so the memory
// must already be applied by then. Pure function of (state, map, i, tick);
// stepEvents doctrine: reads s, never mutates it.
func placeObservedEvents(s *State, m *worldmap.Map, i int, x, y int, nextTick int64) []store.Event {
	a := &s.Agents[i]
	kinds := observedKinds(s, m, x, y)
	if observationDeduped(a, x, y, kinds, nextTick, s.ObservationDedupTicks()) {
		return nil
	}
	why := ""
	if a.Intent != nil {
		why = a.Intent.Reason
	}
	return []store.Event{
		situatedMemoryEvent(nextTick, i, int(s.ObservationBaseSalience()),
			PlaceAt(s, x, y), why, OriginObserved, "%s", observedMemoryText(kinds)),
		{Tick: nextTick, Type: "agent.place_observed", Payload: mustPayload(PlaceObservedPayload{
			Agent: Ref(i), X: x, Y: y, Radius: placeScanRadius, Kinds: kinds,
		})},
	}
}
