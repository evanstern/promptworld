package cognition

import (
	"strings"
	"testing"
)

func TestRegistryValidates(t *testing.T) {
	if err := Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestRegistryContractValues(t *testing.T) {
	// Pinned to contracts/registry.md — a drift here is a doctrine change
	// and must be made there first.
	want := map[string]struct {
		points int
		budget int64
		deg    Degrade
		future bool
	}{
		"planner":       {3, 1200, DegradeReflex, true},
		"conversation":  {13, 7200, DegradeSkip, false},
		"meeting":       {2, 3600, DegradeTemplate, false},
		"consolidation": {5, 28800, DegradeSkip, false},
		"chronicle":     {5, 86400, DegradeSkip, false},
		"metatron":      {5, 86400, DegradeSkip, false},
		// spec 102 (guardian agentization): the scheduled ("steward") lane — budget
		// BELOW planner's so the angel sheds first (contract registry.md,
		// amended in-branch with the spec-102 row).
		"steward": {5, 900, DegradeSkip, false},
	}
	if len(registry) != len(want) {
		t.Fatalf("registry has %d classes, contract has %d", len(registry), len(want))
	}
	for name, w := range want {
		dc, ok := ClassFor(name)
		if !ok {
			t.Fatalf("class %q missing", name)
		}
		if dc.Points != w.points || dc.BudgetTicks != w.budget || dc.Degrade != w.deg || dc.FutureDated != w.future {
			t.Errorf("class %q = %+v, want %+v", name, dc, w)
		}
	}
}

// TestEffectiveBudgetTicks (spec 067 FR-001): the effective budget is the 1x
// budget times the tick rate for every registry class at every capped rung —
// identity at 1x, exact products up the ladder — and the unscaled base at
// uncapped speed (tps <= 0), where the gate must not multiply by zero.
func TestEffectiveBudgetTicks(t *testing.T) {
	rates := []float64{1, 4, 8, 16, 32}
	for name, dc := range registry {
		for _, tps := range rates {
			if got, want := dc.EffectiveBudgetTicks(tps), dc.BudgetTicks*int64(tps); got != want {
				t.Errorf("%s at %gx = %d, want %d", name, tps, got, want)
			}
		}
		for _, tps := range []float64{0, -1} {
			if got := dc.EffectiveBudgetTicks(tps); got != dc.BudgetTicks {
				t.Errorf("%s at uncapped tps %g = %d, want unscaled base %d", name, tps, got, dc.BudgetTicks)
			}
		}
	}
	// Pin the spec-067 contract's reference table verbatim
	// (contracts/landing-gate.md) — a drift here is a doctrine change and
	// must be made there first.
	table := map[string][5]int64{
		"planner":      {1200, 4800, 9600, 19200, 38400},
		"conversation": {7200, 28800, 57600, 115200, 230400},
		"meeting":      {3600, 14400, 28800, 57600, 115200},
	}
	for name, want := range table {
		dc, ok := ClassFor(name)
		if !ok {
			t.Fatalf("class %q missing", name)
		}
		for i, tps := range rates {
			if got := dc.EffectiveBudgetTicks(tps); got != want[i] {
				t.Errorf("%s at %gx = %d, contract table says %d", name, tps, got, want[i])
			}
		}
	}
}

func TestClassForKind(t *testing.T) {
	for kind, class := range map[string]string{
		"planner": "planner", "conversation": "conversation",
		"meeting": "meeting", "consolidation": "consolidation",
		"narrator": "chronicle", "drama": "chronicle", "metatron": "metatron",
	} {
		dc, ok := ClassForKind(kind)
		if !ok || dc.Class != class {
			t.Errorf("ClassForKind(%q) = %q, %v; want %q", kind, dc.Class, ok, class)
		}
	}
	if _, ok := ClassForKind("no-such-kind"); ok {
		t.Error("unknown kind resolved")
	}
}

func TestValidateKindsNamesOffender(t *testing.T) {
	if err := ValidateKinds([]string{"planner", "conversation"}); err != nil {
		t.Fatalf("known kinds: %v", err)
	}
	err := ValidateKinds([]string{"planner", "oracle"})
	if err == nil {
		t.Fatal("unregistered kind passed")
	}
	if !strings.Contains(err.Error(), `"oracle"`) {
		t.Errorf("error does not name the offender: %v", err)
	}
}
