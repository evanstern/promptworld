package sim

// The hail protocol (TASK-47): a talk_to landing flags its target down with a
// deterministic, zero-LLM "let's chat" — social.hailed pauses the target in
// place for a bounded game-tick window so the hailer can close distance and
// the pair actually meets. All state transitions are event-sourced through the
// reducer; the sweep here is a pure function of (state, tick) like every other
// executor beat. No wall clock, no randomness.

import (
	"github.com/evanstern/promptworld/internal/store"
)

// Hail tunables (constants, like the executor cadences/radii in agents.go).
const (
	// hailRadius bounds who a landing may flag down (Manhattan). Observed
	// guard failures land at distances 35–50; 64 covers the population with
	// margin (FR-003).
	hailRadius = 64
	// hailWindowTicks is the pause window: 8 game-minutes. Walk speed is
	// 12 tiles/game-minute (moveEveryTicks = 5), so the far edge of hail
	// range costs ~5.3 game-minutes of walking — ~50% margin for detours
	// (FR-005). Denominated in ticks, so wall-speed changes are irrelevant.
	hailWindowTicks = 480
)

// hailPaused reports whether an agent is currently frozen by an unexpired
// hail. The pause suppresses movement only (executor); intent, plan, needs,
// and social participation are untouched (FR-004).
func hailPaused(a *Agent, tick int64) bool {
	return a.Hail != nil && tick < a.Hail.Until
}

// pairCooled reports whether the unordered pair (a,b) exchanged words within
// the encounter cooldown ending at tick — the conversation loop damper's
// deliberate-talk gate (spec 061, TASK-109, docs/design/evidence/task-109). It
// consumes the SAME spec-048 EncounterCooldown() dial the mind's encounter
// arming uses (one pair-cooldown doctrine, one dial — NO new dial). A
// never-talked pair (PairLastTalk 0) and a dial of 0 are both vacuous (never
// cooled) — the canTalk/LastTalk shape. This is the frequency gate the AMBIENT
// canTalk cooldown never covered on the hail founding path that TASK-47
// deliberately routes around it: first-contact and past-cooldown talks still
// bypass the ambient gate exactly as before, and ONLY a within-cooldown PAIR
// is refused.
func (s *State) pairCooled(a, b int, tick int64) bool {
	last := s.PairLastTalk(a, b)
	return last != 0 && tick-last < s.EncounterCooldown()
}

// hailable is the deterministic predicate a talk_to landing consults before
// flagging a target down (FR-003, research D6/D7). Pure function of State: no
// model, no randomness, no wall clock.
func hailable(s *State, hailer, target int, tick int64) bool {
	if target < 0 || target >= len(s.Agents) || target == hailer {
		return false
	}
	t := &s.Agents[target]
	if t.Dead || t.Asleep {
		return false
	}
	// First hail wins: an already-paused target is not re-hailed and its
	// window is never extended (spec US3-3).
	if t.Hail != nil {
		return false
	}
	// Deadlock prevention (D6): an agent already answering its own hail — a
	// target k it flagged down is still paused waiting for it — can never be
	// frozen by an incoming one. Two agents must never end up mutually frozen.
	for k := range s.Agents {
		if s.Agents[k].Hail != nil && s.Agents[k].Hail.By == target {
			return false
		}
	}
	// Meeting-pinned targets are exempt: pulling an attendee out of the pin
	// would corrupt the governance loop (D7, FR-009).
	if t.Intent != nil && t.Intent.Goal == "attend_meeting" {
		return false
	}
	if meetingActive(s) && attendCandidate(s, target) {
		return false
	}
	if abs(s.Agents[hailer].X-t.X)+abs(s.Agents[hailer].Y-t.Y) > hailRadius {
		return false
	}
	// Pair cooldown (spec 061, TASK-109): the deliberate-talk founding leak —
	// no hail is placed for a pair that spoke within the cooldown. This gates
	// EVERY hail-placement site (the landing door's hail rungs AND the plan-step
	// hail) that TASK-47 routed around the ambient canTalk cooldown, so
	// "deliberate" is no longer "unlimited" (docs/design/evidence/task-109).
	if s.pairCooled(hailer, target, tick) {
		return false
	}
	return true
}

// hailStep is the per-tick hail sweep, run before the per-agent loop in
// stepEvents. For each paused target: if its hailer has closed to adjacency
// (Manhattan ≤ 1) the meeting founds deterministically (social.hail_met + the
// ambient beat's talk shape, bypassing the ambient cooldown — the planner
// deliberately chose this conversation, FR-006); otherwise once the window
// closes the pause lifts untouched (social.hail_expired, FR-005). Met is
// checked before expiry so a hailer arriving on the expiry tick wins.
func hailStep(s *State, nextTick int64) []store.Event {
	var events []store.Event
	for i := range s.Agents {
		a := &s.Agents[i]
		if a.Hail == nil {
			continue
		}
		by := a.Hail.By
		if by >= 0 && by < len(s.Agents) {
			h := &s.Agents[by]
			// Pair cooldown backstop (spec 061, TASK-109): the founding itself
			// routes through the same gate as placement, so even a hail placed
			// legitimately never founds a redundant talk if the pair spoke in
			// between (e.g. an ambient beat raced during the walk-over). The
			// pause then simply lifts on window expiry below. hailable already
			// keeps most hails from being placed; this closes the founding path
			// talkEvents itself (the leak the diagnosis pins).
			if !h.Dead && !h.Asleep && abs(h.X-a.X)+abs(h.Y-a.Y) <= 1 &&
				!s.pairCooled(by, i, nextTick) {
				events = append(events, store.Event{Tick: nextTick, Type: "social.hail_met",
					Payload: mustPayload(HailMetPayload{From: Ref(by), To: Ref(i)})})
				events = append(events, talkEvents(s, by, i, nextTick)...)
				continue
			}
		}
		if nextTick >= a.Hail.Until {
			events = append(events, store.Event{Tick: nextTick, Type: "social.hail_expired",
				Payload: mustPayload(HailExpiredPayload{From: Ref(by), To: Ref(i)})})
		}
	}
	return events
}
