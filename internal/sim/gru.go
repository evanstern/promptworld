package sim

import (
	"encoding/json"
	"fmt"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// The gru (TASK-10): a nocturnal, sight-triggered predator. It is an entity —
// a positioned body in event-sourced state — not a phenomenon: sight needs
// geometry, the TUI needs something to render, and rumors need something to
// have been seen. It wounds the healthy; it can finish the already-fallen
// (spec 044 US3): a healthy villager always survives a single attack (the
// wound floor holds), but a villager already weakened below the near-death
// band can be killed outright — lethality traces back to a preventable
// spiral (no fire, no food, wounds untreated), never a fresh random roll.
//
// Safety is spatial and absolute: fire light (wider than fire warmth, so a
// warm agent is always a safe agent) and shelter tiles make an agent
// invisible to the gru, and the gru itself never steps into protected tiles.
// These rules sit upstream of target selection and are unchanged by
// escalation (FR-016).

// Gru is the predator's event-sourced state; nil means it is not abroad.
// Seen is a bitmask of agents who have sighted it tonight (one omen memory
// per agent per night). LastVictim lets the near-death heartbeat name the
// gru as the cause instead of cold or hunger.
type Gru struct {
	X          int   `json:"x"`
	Y          int   `json:"y"`
	LastAttack int64 `json:"last_attack,omitempty"`
	LastVictim int   `json:"last_victim,omitempty"`
	Seen       uint8 `json:"seen,omitempty"`
}

const (
	gruSightRadius    = 8   // Manhattan; how far the gru sees and is seen
	gruLightRadius    = 3   // fire light: strictly wider than fireWarmRadius
	gruMoveEveryTicks = 4   // slightly faster than agents' 5
	gruWound          = 250 // health torn off per attack
	gruWoundFloor     = 1   // healthy targets: the gru wounds, it does not execute
	gruAttackCooldown = 600 // ticks between wounds (10 game-minutes)
	// gru emergence chance is spec-048-promoted: the default lives in tuning.go
	// (defaultGruEmergePerMille) and the roll reads State.GruEmergePerMille().
	gruSpawnTries = 128

	salGruAttack  = 9
	salGruWitness = 7
	salGruSighted = 6
	toneGruAttack = -60
)

// Static invariant (spec 044 R4/data-model.md "Validation rules"): gruWound
// must stay >= nearDeathBelow, so an escalated hit against a weakened target
// (health < nearDeathBelow) always lands at exactly 0 — never a "survives at
// some positive value" ambiguity that would make the kill non-obvious from
// the constants alone. Assigning a negative difference to a uint fails to
// compile, so a future constant edit that breaks the invariant is caught
// here, not at runtime.
const _ uint = gruWound - nearDeathBelow

// gruNightIndex is the 1-based game night a tick belongs to (day 1 = night 1).
func gruNightIndex(tick int64) int64 { return tick/86400 + 1 }

// lightSource decomposes litAt's own scan into the light truth plus its
// source — today "fire" (firelight) is the only source the sim models (spec
// 074 research R4: the board's "canopy" example has no mechanic behind it,
// so env.go's EnvAt deliberately omits a canopy note rather than invent
// one) — kept as a named source rather than a bare bool so a future light
// mechanic can extend this without another decomposition. litAt is a thin
// wrapper over this (SC-006: the two can never diverge).
func lightSource(s *State, x, y int) (lit bool, source string) {
	for _, st := range s.Structures {
		if st.Kind == "fire" && abs(st.X-x)+abs(st.Y-y) <= gruLightRadius {
			return true, "fire"
		}
	}
	return false, ""
}

// litAt: within fire light. Deliberately wider than warmAt's fire radius so
// everyone huddled close enough to be warm is also protected.
func litAt(s *State, x, y int) bool {
	lit, _ := lightSource(s, x, y)
	return lit
}

// gruProtected: light and shelter are safety — the gru neither sees agents
// here nor sets foot on these tiles.
func gruProtected(s *State, x, y int) bool {
	return litAt(s, x, y) || s.structureAt("shelter", x, y)
}

// gruStep is the predator's whole turn, called from stepEvents: emergence at
// nightfall, withdrawal at dawn, and while abroad — sightings, one attack,
// or one move. Pure over (pre-tick state, map, next tick) like everything
// else in the executor.
func gruStep(s *State, m *worldmap.Map, night bool, nextTick int64) []store.Event {
	var events []store.Event
	emit := func(typ string, payload any) {
		events = append(events, store.Event{Tick: nextTick, Type: typ, Payload: mustPayload(payload)})
	}

	if clock.SecondOfDay(nextTick) == nightStartSecond && s.Gru == nil {
		// Scenario preemption (spec 054, research R3): on a night with a
		// scheduled gru_emerges incident the schedule wins and the random
		// roll is skipped entirely — never two spawn mechanisms in one night.
		// Skipping consumes no RNG draw (rngAt is coordinate-seeded, no
		// stream), so ambient nights and unscheduled scenario nights roll
		// exactly as before.
		if gruScheduledTonight(s, nextTick) {
			return events
		}
		if x, y, ok := gruEmergence(s, m, nextTick); ok {
			emit("gru.emerged", GruEmergedPayload{Night: gruNightIndex(nextTick), X: x, Y: y})
		}
		return events
	}
	if s.Gru == nil {
		return nil
	}
	if !night {
		day, _, _, _ := clock.GameTime(nextTick)
		emit("gru.withdrew", GruWithdrewPayload{Day: day})
		return events
	}

	g := s.Gru

	// Sightings: any live agent with open eyes near enough sees it — safe
	// ones included (watching it prowl past the firelight is what rumors
	// are made of). Once per agent per night.
	for i := range s.Agents {
		a := &s.Agents[i]
		if a.Dead || a.Asleep || g.Seen&(1<<i) != 0 {
			continue
		}
		if abs(a.X-g.X)+abs(a.Y-g.Y) <= gruSightRadius {
			emit("gru.sighted", GruSightedPayload{Agent: Ref(i), X: g.X, Y: g.Y})
			events = append(events, situatedMemoryEvent(nextTick, i, salGruSighted,
				PlaceAt(s, a.X, a.Y), "", OriginAction, "Saw the gru prowling in the dark."))
		}
	}

	// Prey: nearest visible (unprotected) agent, ties to the lowest index.
	target, targetDist := -1, gruSightRadius+1
	for i := range s.Agents {
		a := &s.Agents[i]
		if a.Dead || gruProtected(s, a.X, a.Y) {
			continue
		}
		if d := abs(a.X-g.X) + abs(a.Y-g.Y); d < targetDist {
			target, targetDist = i, d
		}
	}

	// Adjacent prey, claws ready: wound (absolute post-wound health, floored)
	// — UNLESS the victim was already weakened below the near-death band
	// before this attack landed (spec 044 US3/R4): a healthy villager keeps
	// the survival floor (gruWoundFloor), but an already-weakened one takes
	// the floor of 0 and can die outright. The predicate is the raw pre-attack
	// health, never the hysteresis-latched NearDeath bool (research R4) — a
	// villager who recovered past the band is not "already weakened" just
	// because the latch hasn't cleared. Waking with a cleared intent hands a
	// surviving victim to the reflex, which flees to warmth: the curfew,
	// emergent.
	if target >= 0 && targetDist <= 1 &&
		(g.LastAttack == 0 || nextTick-g.LastAttack >= gruAttackCooldown) {
		a := &s.Agents[target]
		floor := gruWoundFloor
		if a.Needs.Health < nearDeathBelow {
			floor = 0
		}
		health := maxInt(floor, a.Needs.Health-gruWound)
		emit("gru.attacked", GruAttackedPayload{Agent: Ref(target), Health: health})
		events = append(events, situatedMemoryEvent(nextTick, target, salGruAttack,
			PlaceAt(s, a.X, a.Y), "", OriginAction, "The gru came out of the dark and tore into me."))
		for w := range s.Agents {
			wa := &s.Agents[w]
			if w == target || wa.Dead || wa.Asleep {
				continue
			}
			if abs(wa.X-a.X)+abs(wa.Y-a.Y) <= witnessRadius {
				events = append(events, situatedMemoryAboutEvent(nextTick, w, target, toneGruAttack, salGruWitness,
					PlaceAt(s, wa.X, wa.Y), OriginWitness, "Saw the gru attack %s in the dark.", a.Name))
			}
		}
		// Escalated kill (spec 044 R5): gru attacks land off the %60 needs
		// heartbeat, so the executor's witness-death block (executor.go) never
		// runs for them — this replicates that idiom inline: the standard
		// agent.died event (cause "gru") so the entire existing death path
		// (reducer Dead flag, inventory spill, grave, chronicle, morgue) runs
		// unchanged (FR-015), plus the same witness-death memory loop.
		if health == 0 {
			emit("agent.died", DiedPayload{Agent: Ref(target), Cause: "gru"})
			for w := range s.Agents {
				if w == target || s.Agents[w].Dead {
					continue
				}
				if abs(s.Agents[w].X-a.X)+abs(s.Agents[w].Y-a.Y) <= witnessRadius {
					events = append(events, situatedMemoryAboutEvent(nextTick, w, target, -80, salWitnessDeath,
						PlaceAt(s, s.Agents[w].X, s.Agents[w].Y), OriginWitness, "Watched %s die of %s.", a.Name, "the gru"))
				}
			}
		}
		return events
	}

	if nextTick%gruMoveEveryTicks != 0 {
		return events
	}
	if target >= 0 && targetDist > 1 {
		// Stalk: the neighbor that closes the gap, never a protected tile.
		// Greedy, not BFS — a monster that can be baffled by water and
		// firelight is the right monster.
		a := &s.Agents[target]
		bx, by, best := g.X, g.Y, targetDist
		for _, d := range neighborOrder {
			nx, ny := g.X+d[0], g.Y+d[1]
			if !passable(m, s, nx, ny) || gruProtected(s, nx, ny) {
				continue
			}
			if nd := abs(a.X-nx) + abs(a.Y-ny); nd < best {
				bx, by, best = nx, ny, nd
			}
		}
		if bx != g.X || by != g.Y {
			emit("gru.moved", GruMovedPayload{X: bx, Y: by})
			return events
		}
	}
	// Prowl: seeded drift through the dark.
	var open [4][2]int
	n := 0
	for _, d := range neighborOrder {
		nx, ny := g.X+d[0], g.Y+d[1]
		if passable(m, s, nx, ny) && !gruProtected(s, nx, ny) {
			open[n] = [2]int{nx, ny}
			n++
		}
	}
	if n > 0 {
		p := open[rngAt(s.Seed, "gru-prowl", nextTick, 0).Uint64N(uint64(n))]
		emit("gru.moved", GruMovedPayload{X: p[0], Y: p[1]})
	}
	return events
}

// gruEmergence rolls whether the gru comes out tonight and picks a passable,
// unprotected border tile to slip in from. Both pure functions of
// (seed, night).
func gruEmergence(s *State, m *worldmap.Map, nextTick int64) (x, y int, ok bool) {
	night := gruNightIndex(nextTick)
	if rngAt(s.Seed, "gru-emerge", night, 0).Uint64N(1000) >= s.GruEmergePerMille() {
		return 0, 0, false
	}
	r := rngAt(s.Seed, "gru-spawn", night, 0)
	for try := 0; try < gruSpawnTries; try++ {
		switch r.Uint64N(4) {
		case 0:
			x, y = int(r.Uint64N(uint64(m.W))), 0
		case 1:
			x, y = int(r.Uint64N(uint64(m.W))), m.H-1
		case 2:
			x, y = 0, int(r.Uint64N(uint64(m.H)))
		case 3:
			x, y = m.W-1, int(r.Uint64N(uint64(m.H)))
		}
		if passable(m, s, x, y) && !gruProtected(s, x, y) {
			return x, y, true
		}
	}
	return 0, 0, false
}

type (
	GruEmergedPayload struct {
		Night int64 `json:"night"`
		X     int   `json:"x"`
		Y     int   `json:"y"`
	}
	GruMovedPayload struct {
		X int `json:"x"`
		Y int `json:"y"`
	}
	GruSightedPayload struct {
		Agent AgentRef `json:"agent"`
		X     int      `json:"x"`
		Y     int      `json:"y"`
	}
	GruAttackedPayload struct {
		Agent AgentRef `json:"agent"`
		// Absolute post-wound value: >= gruWoundFloor when the pre-attack
		// health was >= nearDeathBelow (healthy targets never die from one
		// attack); may be 0 when the target was already weakened (pre-attack
		// health < nearDeathBelow) — spec 044 US3 escalation. A 0 payload here
		// is immediately followed, in the same batch, by agent.died{Cause:"gru"}.
		Health int `json:"health"`
	}
	GruWithdrewPayload struct {
		Day int64 `json:"day"`
	}
)

// applyGru is the reducer arm for gru.* events. Reducer-total like the rest:
// events aimed at a vanished gru no-op rather than error.
func (s *State) applyGru(e store.Event) error {
	switch e.Type {
	case "gru.emerged":
		var p GruEmergedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		s.Gru = &Gru{X: p.X, Y: p.Y}
	case "gru.moved":
		var p GruMovedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		if s.Gru != nil {
			s.Gru.X, s.Gru.Y = p.X, p.Y
		}
	case "gru.sighted":
		var p GruSightedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		if s.Gru != nil && p.Agent.ID >= 0 && p.Agent.ID < len(s.Agents) {
			s.Gru.Seen |= 1 << p.Agent.ID
		}
	case "gru.attacked":
		var p GruAttackedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		if p.Agent.ID < 0 || p.Agent.ID >= len(s.Agents) {
			return fmt.Errorf("apply %s: agent %d out of range", e.Type, p.Agent.ID)
		}
		a := &s.Agents[p.Agent.ID]
		a.Needs.Health = p.Health
		a.Asleep = false
		a.Intent = nil
		a.IdleSince = e.Tick
		if s.Gru != nil {
			s.Gru.LastAttack = e.Tick
			s.Gru.LastVictim = p.Agent.ID
			s.Gru.Seen |= 1 << p.Agent.ID
		}
	case "gru.withdrew":
		s.Gru = nil
	}
	return nil
}
