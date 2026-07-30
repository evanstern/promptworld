package sim

// Derived-progress advancement (spec 104): the second sanctioned eventless
// channel, generalizing the spec-041 D2 precedent (markExplored/notePresence
// as pure functions of (state, event)) to pure functions of (state, tick
// range). Three families ride it — in-flight walk segments (per-step position
// + mental-map bookkeeping at EXACT per-step fidelity, ruling 2), per-minute
// needs decay behind the NeedsSyncTick watermark (ruling 3), and the gru's
// stalk/prowl motion (ruling 4 — fully derived from (state, seed, tick)).
//
// CONVENTION (research.md §1): AdvanceTo(target) executes every pending item
// with scheduled tick STRICTLY BELOW target; Apply calls it before
// dispatching each event, and the live loop calls it at the START of each
// tick. An item scheduled at tick t therefore executes after every event
// recorded at t and before any event at a later tick — on every fold path
// (live loop, recovery, replayToTick, morgue fold, mind/TUI replicas)
// identically, which is what makes derived progress replay-exact with zero
// per-consumer wiring.
//
// ORDER (research.md §2, enforced by the T006 equivalence harness): within
// one derived tick — needs minutes, then agent movement steps by agent
// index, then the gru beat. Idempotent and monotone: each family's watermark
// (Agent.NeedsSyncTick, PathSegment.Done, Gru.Done) lives in marshaled state,
// so snapshots and kill-9 recovery reproduce mid-walk / mid-window progress
// exactly, and no item ever runs twice or rolls back.
//
// Spec-092 discipline: the numbers a segment steps by (MoveEvery, Phase) ride
// the agent.path_started payload, never the compiled constants. The
// advancement-read constants that remain — pathSpeedSlot below, the decay
// constants inside decayNeeds, witnessRadius inside markExplored/notePresence
// (the grandfathered D2 class), gruMoveEveryTicks and the "gru-prowl" RNG
// purpose — are replay-load-bearing for coalesced logs and join the spec-092
// audit (docs/wiki/sim-state-reducer-replay-hazards.md): a retune requires
// the spec-094 format machinery, never a bare edit.

// pathSpeedSlot is the spec-032 path 2x rule's second cadence slot: a
// segment step also fires when (tick+Phase)%MoveEvery == pathSpeedSlot and
// the walker stands ON a path structure — the tile stepped FROM decides,
// exactly the retired per-step emitter's rule. Replay hazard (spec 092/104):
// advancement reads this constant at apply time for coalesced logs — a
// retune requires the spec-094 store.LogFormatVersion bump + migration.
const pathSpeedSlot = 2

// AdvanceTo executes every pending derived item scheduled strictly before
// target, in the fixed order above. Safe on any state (no map, no pending
// work ⇒ no-op); monotone (a target at or below every watermark does
// nothing). The coalescing-regime marker (AmbientCoalescing) gates the needs
// and gru families structurally — a legacy world's fold never enters them —
// while movement is inert on legacy worlds by construction (no
// agent.path_started ever installs a segment).
func (s *State) AdvanceTo(target int64) {
	for {
		t := s.nextDerivedTick(target)
		if t < 0 {
			return
		}
		s.runDerivedTick(t)
	}
}

// nextDerivedTick returns the smallest scheduled derived tick strictly below
// target, or -1 when nothing is pending in range.
func (s *State) nextDerivedTick(target int64) int64 {
	best := int64(-1)
	consider := func(t int64) {
		if t < target && (best < 0 || t < best) {
			best = t
		}
	}
	coalescing := s.AmbientCoalescing()
	for i := range s.Agents {
		a := &s.Agents[i]
		if coalescing && !a.Dead {
			// Next game-minute past the decay watermark. Watermarks are
			// stamped by the regime-flip transition (the tuning arm), by the
			// agent.needs_changed arm, and by derived minutes themselves, so
			// a regime world never has an uncovered gap (research.md §3).
			consider((a.NeedsSyncTick/60 + 1) * 60)
		}
		if seg := a.Path; seg != nil && seg.MoveEvery > 0 && seg.Next < len(seg.Path) {
			consider(segNextCandidate(seg))
		}
	}
	if coalescing && s.Gru != nil {
		consider((s.Gru.Done/gruMoveEveryTicks + 1) * gruMoveEveryTicks)
	}
	return best
}

// segNextCandidate is the segment's next candidate beat strictly after its
// watermark: the next tick hitting either cadence slot. A pathSpeedSlot
// candidate may turn out not to step (the walker is not on a path tile at
// that tick — evaluated at execution, against the advanced state); the
// candidate still consumes the watermark so the scan stays monotone.
func segNextCandidate(seg *PathSegment) int64 {
	for t := seg.Done + 1; ; t++ {
		switch (t + seg.Phase) % seg.MoveEvery {
		case 0, pathSpeedSlot:
			return t
		}
	}
}

// runDerivedTick executes everything due at derived tick t, in the fixed
// within-tick order: needs minutes (agents by index), movement steps (agents
// by index), gru beat.
func (s *State) runDerivedTick(t int64) {
	if s.AmbientCoalescing() && t%60 == 0 {
		for i := range s.Agents {
			a := &s.Agents[i]
			if a.Dead || a.NeedsSyncTick >= t {
				continue
			}
			s.advanceNeedsMinute(i, t)
		}
	}
	for i := range s.Agents {
		a := &s.Agents[i]
		seg := a.Path
		if seg == nil || seg.MoveEvery <= 0 || seg.Next >= len(seg.Path) || t <= seg.Done {
			continue
		}
		// The step rule, byte-for-byte the retired per-step emitter's: fire
		// on the base slot, or on the path slot while standing ON a path
		// structure (state at this tick — event-sourced, so advancement sees
		// exactly what a live per-step walker at t would have).
		ph := (t + seg.Phase) % seg.MoveEvery
		if ph == 0 || (ph == pathSpeedSlot && pathAt(s, a.X, a.Y)) {
			// One step: exactly what the agent.moved arm does — position,
			// explored bits, mutual tick-stamped peer sightings (ruling 2).
			tile := seg.Path[seg.Next]
			a.X, a.Y = tile.X, tile.Y
			s.markExplored(a, tile.X, tile.Y)
			s.notePresence(i, t)
			seg.Next++
			if seg.Next >= len(seg.Path) {
				// Arrival: the segment retires derivationally; the executor
				// observes the arrival via advanced state (and emitted the
				// spec-097 place observation at this step's tick, predicted
				// from the segment — executor.go).
				a.Path = nil
				continue
			}
		}
		seg.Done = t
	}
	if s.AmbientCoalescing() && s.Gru != nil && t%gruMoveEveryTicks == 0 && t > s.Gru.Done {
		s.advanceGruBeat(t)
	}
}

// advanceNeedsMinute applies one game-minute of derived decay to agent i at
// minute tick t: the same decayNeeds arithmetic, near-death latch,
// trajectory-anchor roll, and neglect band anchors the agent.needs_changed
// arm folds (one shared helper, foldNeedsAbsolutes, so the two can never
// drift), then stamps the watermark. Environment inputs (night, fire warmth,
// shelter, cold snap, sleep) read from the advanced state at t — events
// recorded at t have already folded (research.md §1's convention; §5 records
// the razor-edge difference from the legacy heartbeat's pre-batch reads when
// an environment-flipping event lands on the same minute boundary).
//
// Replay hazard (spec 092/104): decayNeeds' constants are re-derived at
// apply time for coalesced logs — see the audit note; a retune requires the
// spec-094 format machinery.
func (s *State) advanceNeedsMinute(i int, t int64) {
	a := &s.Agents[i]
	n := decayNeeds(a.Needs, a.Asleep, s.Night, warmAt(s, a.X, a.Y, t),
		s.Lookup().Structure("shelter", a.X, a.Y), coldSnapActive(s, t))
	s.foldNeedsAbsolutes(a, n, t)
	a.NeedsSyncTick = t
}

// advanceGruBeat runs the gru's derived movement beat at tick t: the shared
// stalk/prowl decision (gru.go — RNG purpose "gru-prowl" preserved verbatim)
// over the advanced state, skipped when an attack was recorded this very
// tick (Gru.LastAttack == t — the attack precludes the move, exactly the
// legacy emitter's exclusivity). Ordered after agent steps within the tick
// (research.md §2), so the derived gru reads agent positions through t.
func (s *State) advanceGruBeat(t int64) {
	g := s.Gru
	g.Done = t
	if g.LastAttack == t {
		return
	}
	if s.m == nil {
		return // bare test states: no terrain, no move (reducer-total)
	}
	if x, y, ok := gruMoveDecision(s, s.m, t); ok {
		g.X, g.Y = x, y
	}
}

// --- spec 104 payloads and reducer arms ---

// PathStartedPayload — agent.path_started: one intent walk, coalesced. Path
// is the full departure-time BFS route (tiles stepped onto, in order, ending
// on the intent target — the same bfs/neighbor order the per-step walker
// used); MoveEvery/Phase are the cadence numbers baked at emission (spec
// 092: advancement never reads the compiled moveEveryTicks or the agent-
// index stagger). A 10-30 tile walk that was 10-30 agent.moved rows is one
// of these (plus, rarely, an agent.path_truncated).
type PathStartedPayload struct {
	Agent     AgentRef `json:"agent"`
	Path      []Point  `json:"path"`
	MoveEvery int64    `json:"move_every"`
	Phase     int64    `json:"phase"`
}

// PathTruncatedPayload — agent.path_truncated: the walk stopped short of its
// declared path — the executor's blocked-path re-route (a wall built
// mid-segment; the next tick re-plans or resolves unreachable). X/Y carry
// the agent's ACTUAL position at truncation (outcome in payload — the arm
// sets it verbatim and never recomputes where the walk "would have" been).
// Every OTHER walk interruption is recorded by its own event, whose arm
// clears the segment (research.md §4).
type PathTruncatedPayload struct {
	Agent AgentRef `json:"agent"`
	X     int      `json:"x"`
	Y     int      `json:"y"`
}

// applyPathStarted installs the in-flight segment (Done = the departure
// tick, so steps fire strictly after it). Reducer-total: an empty path
// installs nothing.
func (s *State) applyPathStarted(a *Agent, p PathStartedPayload, tick int64) {
	if len(p.Path) == 0 || p.MoveEvery <= 0 {
		a.Path = nil
		return
	}
	a.Path = &PathSegment{
		Path:      append([]Point(nil), p.Path...),
		MoveEvery: p.MoveEvery,
		Phase:     p.Phase,
		Done:      tick,
	}
}

// segStepFiresAt reports whether the agent's in-flight segment fires a step
// at tick t — the executor's arrival-observation prediction reads it against
// pre-tick state, the same beat rule (and the same standing tile) the
// advancement engine will evaluate when the step executes.
func segStepFiresAt(s *State, a *Agent, seg *PathSegment, t int64) bool {
	if seg.MoveEvery <= 0 || t <= seg.Done {
		return false
	}
	ph := (t + seg.Phase) % seg.MoveEvery
	return ph == 0 || (ph == pathSpeedSlot && pathAt(s, a.X, a.Y))
}

// truncateWalk clears an agent's in-flight segment — the shared primitive
// every walk-invalidating reducer arm calls (intent lifecycle, sleep, death,
// gru attack, hail, teleport, pause, time snap). Position stays wherever
// advancement actually walked it: exact through the interrupting event's
// tick by the strictly-before convention. A nil segment (every legacy world)
// is a no-op, so legacy folds are byte-untouched.
func truncateWalk(a *Agent) { a.Path = nil }
