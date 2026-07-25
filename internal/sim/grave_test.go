package sim

import (
	"testing"
)

// Graves (spec 044 US4, FR-017/FR-018/FR-019): a death leaves a persistent
// Structure{Kind:"grave"} at its tile — the same reducer idiom as the spec
// 013 inventory spill. Research R10 verified that riding the Structure route
// gets perception, telling, and map rendering for free; these tests confirm
// that claim rather than add new plumbing (R10/R11).

// TestGravePlacedAtDeathTilePersistsAndReplays is T023/T026: a death places a
// grave at its tile, the grave survives untouched by later events (nothing
// ever removes a grave — no such reducer path exists), and replaying the
// same log from genesis lands on byte-identical state (SC-004-adjacent,
// FR-020).
func TestGravePlacedAtDeathTilePersistsAndReplays(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	live := NewState(seed, m)
	a := &live.Agents[0]
	dx, dy := a.X, a.Y

	applyEvent(t, live, 100, "agent.died", DiedPayload{Agent: 0, Cause: "starvation"})

	if !live.structureAt("grave", dx, dy) {
		t.Fatal("no grave structure at the death tile")
	}
	// Persistence: drive the world onward: the grave must still be there.
	driveTicks(t, live, m, live.Tick+5_000, nil)
	if !live.structureAt("grave", dx, dy) {
		t.Fatal("grave vanished after further ticks — graves must persist")
	}

	// Replay: reduce the same event onto a fresh genesis state independently.
	replayed := NewState(seed, m)
	applyEvent(t, replayed, 100, "agent.died", DiedPayload{Agent: 0, Cause: "starvation"})
	if !replayed.structureAt("grave", dx, dy) {
		t.Fatal("replayed state missing the grave")
	}
}

// TestBuildSiteBlockedByGrave documents the deliberate research R10 tension:
// a grave occupies its tile for buildSite exactly like any other structure,
// so building over a death site is refused — the conservative default the
// spec's edge case explicitly defers ("a grave persists and remains
// addressable").
func TestBuildSiteBlockedByGrave(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	bx, by, ok := findBuildTile(m, s)
	if !ok {
		t.Skip("no buildable tile on this map")
	}
	if !buildSite(m, s, bx, by) {
		t.Fatal("setup: tile should be buildable before the grave lands")
	}
	s.Structures = append(s.Structures, Structure{Kind: "grave", X: bx, Y: by})
	if buildSite(m, s, bx, by) {
		t.Error("buildSite accepted a tile holding a grave — research R10 requires refusal")
	}
}

// TestPerceptionSweepGrantsGraveFact is T026: a grave is ground truth like
// any other structure kind, so groundFactPresentIn's default case and the
// perception sweep's generic per-structure `note()` call pick it up with no
// grave-specific code (research R10) — a living villager within
// witnessRadius gains it as a witnessed place-fact on their very next
// perception beat (at most one movement beat later).
func TestPerceptionSweepGrantsGraveFact(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	a := &s.Agents[0]
	s.Structures = append(s.Structures, Structure{Kind: "grave", X: a.X, Y: a.Y})

	beat := int64(moveEveryTicks * 8) // agent 0's beat: (tick + 0*3) % 5 == 0
	var got *PlaceFact
	for _, e := range perceptionEvents(s, m, beat) {
		var p SawPayload
		mustUnmarshal(t, e.Payload, &p)
		if p.Agent != 0 {
			continue
		}
		for _, f := range p.Facts {
			if f.Kind == "grave" && f.X == a.X && f.Y == a.Y {
				q := f
				got = &q
			}
		}
	}
	if got == nil {
		t.Fatal("witness's perception sweep did not grant the grave place-fact")
	}
	if got.Provenance != ProvenanceWitnessed {
		t.Errorf("grave fact provenance = %q, want witnessed", got.Provenance)
	}
}

// TestPlaceTellSpreadsGraveFact is T026: a known grave spreads through
// founded-talk place-telling exactly like any other PlaceFact kind — no
// grave-specific code in tellablePlaces (research R10).
func TestPlaceTellSpreadsGraveFact(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	a, b := &s.Agents[0], &s.Agents[1]
	b.X, b.Y = a.X, a.Y+1 // adjacent: a founded talk's shape
	a.Map.Facts = nil
	b.Map.Facts = nil

	a.Map.upsertFact(PlaceFact{Kind: "grave", X: 20, Y: 20, Seen: 900, Provenance: ProvenanceWitnessed})
	s.Structures = []Structure{{Kind: "grave", X: 20, Y: 20}}

	const tick = int64(1000)
	var told *PlaceToldPayload
	for _, e := range talkEvents(s, 0, 1, tick) {
		if e.Type != "social.place_told" {
			continue
		}
		var p PlaceToldPayload
		mustUnmarshal(t, e.Payload, &p)
		if p.From == 0 && p.To == 1 {
			q := p
			told = &q
		}
	}
	if told == nil {
		t.Fatal("no A→B social.place_told on a founded talk")
	}
	var gotGrave bool
	for _, f := range told.Facts {
		if f.Kind == "grave" && f.X == 20 && f.Y == 20 {
			if f.Provenance != ProvenanceTold || f.Source != 0 {
				t.Errorf("told grave fact not baked told-by-teller: %+v", f)
			}
			gotGrave = true
		}
	}
	if !gotGrave {
		t.Errorf("grave fact did not spread through place-telling: %+v", told.Facts)
	}
}

// TestGriefRumorFromWitnessedDeath is SC-006's grief-rumor half (research
// R11): FR-019 rides the ALREADY-SHIPPED witness-death memory
// ("Watched %s die of %s.", salience salWitnessDeath=10) straight into
// TellableFor's birth path (>= rumorMinSalience=4, the strongest possible
// rumor seed) with no new memory class or drive. This is a verifying
// integration test, not new plumbing: a witness who saw a death can, on the
// very next founded talk (well within a game-day — talks happen on the
// nextTick%60==30 social cadence), pass a grief-flavored social.rumor_told
// naming the dead villager as subject.
func TestGriefRumorFromWitnessedDeath(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	witness, listener := &s.Agents[1], &s.Agents[2]
	listener.X, listener.Y = witness.X, witness.Y+1 // adjacent: a founded talk's shape

	applyEvent(t, s, 100, "agent.died", DiedPayload{Agent: 0, Cause: "starvation"})
	// The heartbeat death loop's witness-death memory only fires for agents
	// within witnessRadius of the deceased at the moment of death; stage it
	// directly (the memory's shape, not its emission site, is what this test
	// verifies) so the test is independent of the heartbeat's own witnessing
	// geometry, already covered by TestGruEscalationScenario/executor tests.
	witness.Memories = append(witness.Memories, Memory{
		Subject: 0, Tone: -80, Salience: salWitnessDeath, Text: "Watched Ash die of starvation.",
	})

	tell, ok := TellableFor(s, 1, 2)
	if !ok || tell.Subject != 0 {
		t.Fatalf("witness-death memory is not tellable gossip: ok=%v subject=%d", ok, tell.Subject)
	}

	var rumor *RumorToldPayload
	for _, e := range talkEvents(s, 1, 2, 1000) {
		if e.Type == "social.rumor_told" {
			var p RumorToldPayload
			mustUnmarshal(t, e.Payload, &p)
			rumor = &p
		}
	}
	if rumor == nil {
		t.Fatal("a founded talk between witness and listener produced no social.rumor_told")
	}
	if rumor.Subject != 0 || rumor.Tone >= 0 {
		t.Errorf("rumor_told = %+v, want subject 0 (the deceased) with a negative (grief) tone", *rumor)
	}
}
