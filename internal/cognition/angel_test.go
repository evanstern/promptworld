package cognition

import "testing"

// --- spec 102 (guardian agentization): the "angel" decision class ---

// TestAngelClassRegistered pins the scheduled lane's registry row (D2): the
// class exists, the "angel" llm kind resolves to it, and the registry still
// validates whole.
func TestAngelClassRegistered(t *testing.T) {
	dc, ok := ClassFor("angel")
	if !ok {
		t.Fatal("angel class not registered")
	}
	if dc.Points != 5 || dc.BudgetTicks != 900 || dc.Degrade != DegradeSkip {
		t.Fatalf("angel class values drifted: %+v", dc)
	}
	kdc, ok := ClassForKind("angel")
	if !ok || kdc.Class != "angel" {
		t.Fatalf("kind %q resolved to %+v (ok=%v), want the angel class", "angel", kdc, ok)
	}
	if err := Validate(); err != nil {
		t.Fatalf("registry invalid with angel class: %v", err)
	}
}

// TestAngelShedsBeforeVillagerSurvival is the D2 shed-order pin: under
// saturation the angel must be suppressed BEFORE the villager survival class
// (planner — the reflex-floored class that keeps villagers alive-in-mind).
// Two forms, both across a realistic seconds-per-point sweep:
//
//  1. Max-safe-speed ordering: the angel's highest routable ladder speed
//     never exceeds the planner's.
//  2. Pointwise: at any (speed, secPerPt) where the angel routes, the
//     planner routes too — there is no operating point where the caretaker
//     thinks while villager survival cognition is shed.
func TestAngelShedsBeforeVillagerSurvival(t *testing.T) {
	angel, _ := ClassFor("angel")
	planner, _ := ClassFor("planner")
	spps := []float64{0.5, 1, 2, 5, 10, 20, 60}
	for _, spp := range spps {
		if a, p := MaxSafeSpeed("angel", spp), MaxSafeSpeed("planner", spp); a > p {
			t.Errorf("at %.1fs/pt: angel safe to %gx but planner only %gx — the angel must shed first", spp, a, p)
		}
		for _, tps := range []float64{1, 4, 8, 16, 32} {
			av, pv := Route(angel, tps, spp), Route(planner, tps, spp)
			if av.Allow && !pv.Allow {
				t.Errorf("at %.1fs/pt %gx: angel allowed while planner suppressed — shed order inverted", spp, tps)
			}
		}
	}
	// Concreteness: at the bootstrap local seed the angel sheds at 16x while
	// the planner still routes — the shed gap is real, not vacuous.
	if v := Route(angel, 16, BootstrapLocalSecPerPt); v.Allow {
		t.Errorf("angel at 16x/bootstrap should be suppressed, got allow (%s)", v.Arithmetic)
	}
	if v := Route(planner, 16, BootstrapLocalSecPerPt); !v.Allow {
		t.Errorf("planner at 16x/bootstrap should still route, got suppressed (%s)", v.Arithmetic)
	}
}

// TestNextPhasePreservingDueShared pins the shared cadence arithmetic (SC-004:
// one schedule implementation for the planner and angel lanes) — the TASK-44
// semantics, now exported here.
func TestNextPhasePreservingDueShared(t *testing.T) {
	cases := []struct{ due, tick, cadence, want int64 }{
		{100, 50, 30, 100},   // not overdue: untouched
		{100, 100, 30, 130},  // due == tick advances one cadence
		{100, 250, 30, 280},  // multiple cadences skipped, phase kept (100 mod 30 == 280 mod 30)
		{100, 250, 0, 100},   // degenerate cadence: untouched
		{100, 99999, 30, 100029}, // long stall, phase preserved
	}
	for _, c := range cases {
		if got := NextPhasePreservingDue(c.due, c.tick, c.cadence); got != c.want {
			t.Errorf("NextPhasePreservingDue(%d,%d,%d) = %d, want %d", c.due, c.tick, c.cadence, got, c.want)
		}
		if c.cadence > 0 {
			if got := NextPhasePreservingDue(c.due, c.tick, c.cadence); got%c.cadence != c.due%c.cadence {
				t.Errorf("phase not preserved: due %d cadence %d → %d", c.due, c.cadence, got)
			}
		}
	}
}
