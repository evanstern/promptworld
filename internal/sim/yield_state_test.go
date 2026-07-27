package sim

// Spec 062 (instinct yields to intelligence). T001: the yield-window anchor
// (Agent.LastMindIntentDone) is event-sourced, snapshot-compatible (omitempty),
// and armed ONLY by non-reflex intent completions. T005 (below): the PREP gate
// — the yield window suppresses prep and decays, the danger bands suppress
// regardless of the window, reflex completions never arm, and a no-planner
// drive matches pre-062 behavior except the enumerated danger-band suppressions.

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// prepAgent configures agent 0 as an idle, fed, rested, warm DAYTIME villager
// who already KNOWS a fire (so the first-fire prep rung skips), carries no wood
// (so the refuel top-up rung skips), and has an empty larder plus a KNOWN
// forage tile — the state in which the larder PREP rung is the sole decider, so
// an ungated decideIntent resolves to "forage". Returns the state, map, agent,
// and the drive tick. Skips the test if the seed has no reachable forage.
func prepAgent(t *testing.T, seed uint64) (*State, *worldmap.Map, *Agent, int64) {
	t.Helper()
	const now int64 = 5000
	m := testMap(seed)
	s := NewState(seed, m)
	s.Night = false
	a := &s.Agents[0]
	a.Dead = false
	a.Asleep = false
	// Every need comfortably above its danger band, so only the yield window (or
	// a deliberately lowered need) can gate prep.
	a.Needs = Needs{Health: 1000, Food: 600, Rest: 600, Warmth: 600, Morale: 600}
	a.Inv = Inventory{} // no wood ⇒ refuel top-up skips
	// A known lit fire off-tile: first-fire prep skips (knows a fire), and the
	// agent is NOT standing in it (no real structure), so warmAt stays false.
	a.Map.upsertFact(PlaceFact{Kind: "fire", X: a.X + 6, Y: a.Y, Seen: now,
		Provenance: ProvenanceWitnessed, Detail: now + 100000})
	// A known forage tile ⇒ the empty-larder rung resolves to forage.
	spot, ok := nearest(m, s, a.X, a.Y, func(x, y int) bool {
		return effectiveKind(m, s, x, y) == worldmap.Forage
	})
	if !ok {
		t.Skipf("seed %d has no reachable forage for the prep rung", seed)
	}
	a.Map.upsertFact(PlaceFact{Kind: "forage", X: spot.X, Y: spot.Y, Seen: now, Provenance: ProvenanceWitnessed})
	a.Inv.FoodRaw = 0 // below stockFoodRawTo ⇒ the larder rung is armed
	return s, m, a, now
}

func goalOf(d decision) string {
	if d.intent == nil {
		return d.directEvent
	}
	return d.intent.Goal
}

// TestYieldAnchorOmitemptySnapshotCompat (T001): an agent that never completed
// a mind-sourced intent carries no last_mind_intent_done key, so its canonical
// bytes stay identical to a pre-062 snapshot; and a pre-062 snapshot (the key
// absent) round-trips to the 0 sentinel.
func TestYieldAnchorOmitemptySnapshotCompat(t *testing.T) {
	m := testMap(42)
	s := NewState(42, m)

	// Never mind-driven ⇒ the field is absent from the marshalled state.
	if got := s.Marshal(); bytes.Contains(got, []byte("last_mind_intent_done")) {
		t.Fatalf("a never-mind-driven world serialized the yield anchor:\n%s", got)
	}

	// A pre-062 agent snapshot (no last_mind_intent_done key) unmarshals to 0.
	pre062 := []byte(`{"name":"Old","x":1,"y":1,"needs":{"health":1000,"food":600,"rest":600,"warmth":600,"morale":600},"inv":{},"idle_since":0,"last_talk":0}`)
	var a Agent
	if err := json.Unmarshal(pre062, &a); err != nil {
		t.Fatalf("pre-062 agent snapshot did not round-trip: %v", err)
	}
	if a.LastMindIntentDone != 0 {
		t.Fatalf("absent yield anchor unmarshalled to %d, want the 0 sentinel", a.LastMindIntentDone)
	}

	// Once armed, the field serializes with its tick (the omitempty tail stays
	// compact for every un-armed agent).
	a.LastMindIntentDone = 4321
	if got, _ := json.Marshal(a); !bytes.Contains(got, []byte(`"last_mind_intent_done":4321`)) {
		t.Fatalf("an armed yield anchor did not serialize:\n%s", got)
	}
}

// TestYieldAnchorArmedOnlyByMindCompletion (T001/FR-003): the intent-completion
// reducer arm stamps LastMindIntentDone from the completing intent's ring-record
// source — set for planner/plan/meeting completions, NEVER for reflex ones.
func TestYieldAnchorArmedOnlyByMindCompletion(t *testing.T) {
	for _, tc := range []struct {
		source    string
		wantArmed bool
	}{
		{"planner", true},
		{"plan", true},
		{"meeting", true},
		{"reflex", false},
		{"", false},
	} {
		t.Run("source="+tc.source, func(t *testing.T) {
			m := testMap(42)
			s := NewState(42, m)
			a := &s.Agents[0]

			// The intent lands with the source under test, then completes.
			set := store.Event{Tick: 1000, Type: "agent.intent_set",
				Payload: mustPayload(IntentSetPayload{Agent: Ref(0), Goal: "goto_warmth", TargetX: a.X, TargetY: a.Y, Source: tc.source})}
			done := store.Event{Tick: 1200, Type: "agent.intent_done",
				Payload: mustPayload(AgentPayload{Agent: Ref(0)})}
			if err := s.Apply(set); err != nil {
				t.Fatal(err)
			}
			if err := s.Apply(done); err != nil {
				t.Fatal(err)
			}

			if tc.wantArmed && a.LastMindIntentDone != 1200 {
				t.Fatalf("%q completion armed the window to %d, want 1200", tc.source, a.LastMindIntentDone)
			}
			if !tc.wantArmed && a.LastMindIntentDone != 0 {
				t.Fatalf("%q completion armed the window to %d, want it never armed (0)", tc.source, a.LastMindIntentDone)
			}
		})
	}
}

// TestPrepBaselineForagesWhenUngated pins the harness: with no yield window
// armed and every need healthy, the prep group fires exactly as pre-062 — the
// larder rung forages. This is the control the suppression tests below flip.
func TestPrepBaselineForagesWhenUngated(t *testing.T) {
	s, m, a, now := prepAgent(t, 42)
	a.LastMindIntentDone = 0 // never mind-driven

	if prepYields(s, a, now) {
		t.Fatal("healthy, unarmed villager: prep must NOT yield (pre-062 parity)")
	}
	if g := goalOf(decideIntent(s, m, 0, now)); g != "forage" {
		t.Fatalf("ungated prep should forage the larder, got %q", g)
	}
}

// TestYieldWindowSuppressesPrepAndDecays (T005, US1 AS1/AS4): a non-reflex
// intent that completed within prepYieldTicks suppresses prep (the agent
// wanders instead of foraging); once the window elapses with the agent still
// idle, prep resumes exactly as today (yield decays — instinct never strands
// the idle loop forever).
func TestYieldWindowSuppressesPrepAndDecays(t *testing.T) {
	s, m, a, now := prepAgent(t, 42)

	// Armed: a planner intent completed 100 ticks ago (< prepYieldTicks 1800).
	a.LastMindIntentDone = now - 100
	if !prepYields(s, a, now) {
		t.Fatal("within the yield window, prep must yield")
	}
	if g := goalOf(decideIntent(s, m, 0, now)); g == "forage" {
		t.Fatalf("prep fired inside the yield window (got %q) — the layer fight is not over", g)
	}

	// Boundary: exactly prepYieldTicks elapsed ⇒ window closed, prep resumes.
	a.LastMindIntentDone = now - prepYieldTicks
	if prepYields(s, a, now) {
		t.Fatal("at exactly prepYieldTicks the window has closed; prep must resume")
	}
	if g := goalOf(decideIntent(s, m, 0, now)); g != "forage" {
		t.Fatalf("prep should resume once the window decays, got %q", g)
	}
}

// TestReflexCompletionNeverArmsWindow (T005, US1): a reflex-sourced completion
// never arms the yield window, so a villager whose only recent intent was a
// reflex one keeps doing prep — instinct never yields to itself (the no-planner
// deadlock guard). Driven through the reducer end-to-end, then decideIntent.
func TestReflexCompletionNeverArmsWindow(t *testing.T) {
	s, m, a, now := prepAgent(t, 42)

	set := store.Event{Tick: now - 50, Type: "agent.intent_set",
		Payload: mustPayload(IntentSetPayload{Agent: Ref(0), Goal: "wander", TargetX: a.X + 1, TargetY: a.Y, Source: "reflex"})}
	done := store.Event{Tick: now - 40, Type: "agent.intent_done", Payload: mustPayload(AgentPayload{Agent: Ref(0)})}
	if err := s.Apply(set); err != nil {
		t.Fatal(err)
	}
	if err := s.Apply(done); err != nil {
		t.Fatal(err)
	}
	if a.LastMindIntentDone != 0 {
		t.Fatalf("a reflex completion armed the window to %d — instinct yielded to itself", a.LastMindIntentDone)
	}
	if prepYields(s, a, now) {
		t.Fatal("after only a reflex completion, prep must not yield")
	}
	if g := goalOf(decideIntent(s, m, 0, now)); g != "forage" {
		t.Fatalf("reflex-only villager should keep doing prep, got %q", g)
	}
}

// TestDangerBandSuppressesPrepRegardlessOfWindow (T005, US1 AS2): a need in its
// danger band yields prep even with the window fully decayed. Uses the durable
// residual case — warmth below its band while standing AT a fire (warmAt), so
// the day-warmth SURVIVAL rung is skipped (it guards on !warmAt) and the
// prep-gate's danger clause is what holds the agent — the one warmth-danger
// suppression that survives the US2 day-warmth rung.
func TestDangerBandSuppressesPrepRegardlessOfWindow(t *testing.T) {
	s, m, a, now := prepAgent(t, 42)
	a.LastMindIntentDone = 0 // window inert — isolate the danger band

	// A real lit fire ON the agent's tile ⇒ warmAt true (recovering in place),
	// but the warmth NEED is still in its danger band.
	s.Structures = append(s.Structures, Structure{Kind: "fire", X: a.X, Y: a.Y, FuelUntil: now + 100000})
	a.Needs.Warmth = dangerWarmthBelow - 1
	if !warmAt(s, a.X, a.Y, now) {
		t.Fatal("agent should be warm standing on a lit fire")
	}
	if !prepYields(s, a, now) {
		t.Fatal("warmth in the danger band must yield prep even with the window decayed")
	}
	if g := goalOf(decideIntent(s, m, 0, now)); g == "forage" {
		t.Fatalf("prep fired while warmth was in its danger band, got %q", g)
	}
}

// TestNoLLMDegradedModeYieldWindowInert (T005, SC-003 / FR-007): in a world
// driven WITHOUT any planner or meeting activity (the permanent degraded mode),
// no non-reflex intent ever completes, so the yield window never arms — every
// agent's LastMindIntentDone stays the 0 sentinel across a long drive. The
// yield-window clause of prepYields is therefore inert in a no-LLM world; the
// ONLY way a planner-free reflex diverges from pre-062 is the danger-band
// clause, enumerated in TestNoLLMDangerBandSuppressionCasesEnumerated.
func TestNoLLMDegradedModeYieldWindowInert(t *testing.T) {
	const seed int64 = 42
	m := testMap(42)
	s := NewState(42, m)
	// A full day/night cycle of pure reflex — no intents injected, no meeting
	// convention established (NewState seeds none), so nothing non-reflex runs.
	driveTicks(t, s, m, 24*3600, nil)
	for i, a := range s.Agents {
		if a.LastMindIntentDone != 0 {
			t.Errorf("agent %d (%s) armed the yield window to %d in a planner-free world — degraded mode diverged",
				i, a.Name, a.LastMindIntentDone)
		}
	}
	_ = seed
}

// TestNoLLMDangerBandSuppressionCasesEnumerated (T005, SC-003): the COMPLETE
// enumeration of where a no-LLM reflex suppresses prep versus pre-062. With the
// yield window inert (LastMindIntentDone == 0), prepYields reduces to its
// danger-band clause. Each need's band is anchored at its survival-rung trigger
// (R3), so the divergence from pre-062 is exactly:
//
//	Food   < dangerFoodBelow (350): the eat / get-food SURVIVAL rung preempts
//	  prep in every reachable case EXCEPT the extreme "knows no food AND no
//	  frontier reachable" tail, where prep is now suppressed → wander (pre-062
//	  would attempt a larder forage that also fails / refuel / wander). Marginal.
//	Rest   < dangerRestBelow (250): by day the nap SURVIVAL rung always preempts
//	  prep, so this band NEVER diverges — it is defensive/subsumed. (Prep is a
//	  day-only group; the night branch returns before it.)
//	Warmth < dangerWarmthBelow (350): the US2 day-warmth SURVIVAL rung preempts
//	  when NOT already warm; when warmAt (recovering AT a fire) the band
//	  suppresses prep → wander (pre-062 had no day warmth handling and would run
//	  prep). This is the one live by-day divergence.
//
// The table asserts the predicate directly; each row states its pre-062
// disposition so the exceptions are the enumerated, reviewed set — not silent.
func TestNoLLMDangerBandSuppressionCasesEnumerated(t *testing.T) {
	m := testMap(42)
	s := NewState(42, m)
	a := &s.Agents[0]
	a.Needs = Needs{Health: 1000, Food: 600, Rest: 600, Warmth: 600, Morale: 600}
	a.LastMindIntentDone = 0 // degraded mode: window inert
	const now int64 = 5000

	cases := []struct {
		name         string
		food         int
		warmth       int
		rest         int
		wantSuppress bool
	}{
		{"all healthy — identical to pre-062 (prep fires)", 600, 600, 600, false},
		{"food at band boundary — not suppressed (strict <)", dangerFoodBelow, 600, 600, false},
		{"warmth at band boundary — not suppressed (strict <)", 600, dangerWarmthBelow, 600, false},
		{"rest at band boundary — not suppressed (strict <)", 600, 600, dangerRestBelow, false},
		{"food in danger — survival preempts except the no-food/no-frontier tail", dangerFoodBelow - 1, 600, 600, true},
		{"warmth in danger — US2 rung preempts unless warmAt (then this diverges)", 600, dangerWarmthBelow - 1, 600, true},
		{"rest in danger — defensive; nap always preempts by day (never diverges)", 600, 600, dangerRestBelow - 1, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a.Needs.Food, a.Needs.Warmth, a.Needs.Rest = c.food, c.warmth, c.rest
			if got := prepYields(s, a, now); got != c.wantSuppress {
				t.Fatalf("prepYields(food=%d warmth=%d rest=%d) = %v, want %v",
					c.food, c.warmth, c.rest, got, c.wantSuppress)
			}
		})
	}
}
