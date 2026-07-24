package cognition

import (
	"strings"
	"testing"
)

// TestHorizonSummaryBootstrap (spec 035 T003): at the bootstrap local seed
// (20 s/pt) the summary matches contracts/warnings.md §1's worked example —
// planner and conversation suppressed above 16x, meeting OK through 32x.
func TestHorizonSummaryBootstrap(t *testing.T) {
	got := HorizonSummary(BootstrapLocalSecPerPt)
	for _, want := range []string{
		"planner suppressed above 16x",
		"conversation suppressed above 16x",
		"meeting OK at 32x",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("HorizonSummary(%.1f) = %q, missing %q", BootstrapLocalSecPerPt, got, want)
		}
	}
}

// TestHorizonSummaryCalibrated: a fast, calibrated rig (0.94 s/pt) clears
// every watched class at the ladder ceiling.
func TestHorizonSummaryCalibrated(t *testing.T) {
	got := HorizonSummary(0.94)
	for _, class := range []string{"planner", "conversation", "meeting"} {
		if !strings.Contains(got, class+" OK at 32x") {
			t.Errorf("HorizonSummary(0.94) = %q, want %q OK at 32x", got, class)
		}
	}
}

// TestHorizonSummaryAgreesWithRoute: the summary's suppressed/OK verdicts
// agree with Route across the full ladder, at both bootstrap and calibrated
// seconds-per-point — the property FR-006 demands structurally.
func TestHorizonSummaryAgreesWithRoute(t *testing.T) {
	for _, secPerPt := range []float64{BootstrapLocalSecPerPt, BootstrapCloudSecPerPt, 0.94, 5.0} {
		summary := HorizonSummary(secPerPt)
		for _, class := range watchedClasses {
			dc, ok := ClassFor(class)
			if !ok {
				t.Fatalf("watched class %q not registered", class)
			}
			allowedAt32 := Route(dc, 32, secPerPt).Allow
			saysOKAt32 := strings.Contains(summary, class+" OK at 32x")
			if allowedAt32 != saysOKAt32 {
				t.Errorf("secPerPt=%.2f class=%s: Route.Allow@32x=%v but summary OK-at-32x=%v (%q)",
					secPerPt, class, allowedAt32, saysOKAt32, summary)
			}
		}
	}
}

// TestSuppressedAtAgreesWithRoute (spec 035 contracts/warnings.md Test
// obligations): SuppressedAt says a class is suppressed exactly when Route
// disallows it at that (ticksPerSecond, secPerPt) point, across the ladder ×
// watched-class matrix.
func TestSuppressedAtAgreesWithRoute(t *testing.T) {
	perClassSecPerPt := map[string]float64{
		"planner":      BootstrapLocalSecPerPt,
		"conversation": BootstrapLocalSecPerPt,
		"meeting":      BootstrapLocalSecPerPt,
	}
	lookup := func(class string) (float64, bool) {
		sp, ok := perClassSecPerPt[class]
		return sp, ok
	}
	for _, tps := range horizonLadder {
		suppressed := SuppressedAt(tps, lookup)
		suppressedSet := make(map[string]bool, len(suppressed))
		for _, c := range suppressed {
			suppressedSet[c] = true
		}
		for _, class := range watchedClasses {
			dc, _ := ClassFor(class)
			wantSuppressed := !Route(dc, tps, perClassSecPerPt[class]).Allow
			if suppressedSet[class] != wantSuppressed {
				t.Errorf("tps=%g class=%s: SuppressedAt says suppressed=%v, Route says suppressed=%v",
					tps, class, suppressedSet[class], wantSuppressed)
			}
		}
	}
}

// TestSuppressedAtGateExcludesClass: secPerPtFor returning ok=false (the
// seed-state gate — a calibrated provider) removes the class from the result
// entirely, even at a speed where the bootstrap arithmetic would suppress it.
func TestSuppressedAtGateExcludesClass(t *testing.T) {
	got := SuppressedAt(32, func(class string) (float64, bool) {
		return 0, false // every class gated out (calibrated world)
	})
	if len(got) != 0 {
		t.Errorf("gated-out lookup must yield no suppressed classes, got %v", got)
	}
}

// TestLiveHorizonStandingAgreesWithRoute (spec 037 SC-002): every entry's
// Suppressed flag and Verdict equal Route's own verdict for that class at that
// (ticksPerSecond, secPerPt) point, across the ladder × watched-class matrix.
func TestLiveHorizonStandingAgreesWithRoute(t *testing.T) {
	lookup := func(string) (float64, bool) { return BootstrapLocalSecPerPt, true }
	for _, tps := range horizonLadder {
		standings := LiveHorizon(tps, lookup)
		if len(standings) != len(watchedClasses) {
			t.Fatalf("tps=%g: got %d standings, want %d", tps, len(standings), len(watchedClasses))
		}
		for i, cs := range standings {
			if cs.Class != watchedClasses[i] {
				t.Errorf("tps=%g entry %d: class=%q, want %q (WatchedClasses order)", tps, i, cs.Class, watchedClasses[i])
			}
			dc, _ := ClassFor(cs.Class)
			want := Route(dc, tps, BootstrapLocalSecPerPt)
			if cs.Suppressed != !want.Allow {
				t.Errorf("tps=%g class=%s: Suppressed=%v, want %v", tps, cs.Class, cs.Suppressed, !want.Allow)
			}
			if cs.Verdict.Arithmetic != want.Arithmetic {
				t.Errorf("tps=%g class=%s: Verdict.Arithmetic=%q, want %q", tps, cs.Class, cs.Verdict.Arithmetic, want.Arithmetic)
			}
		}
	}
}

// TestLiveHorizonExcludesGatedClass: a class whose lookup returns ok=false is
// absent from the result (no entry) — the serving-provider gate, matching
// SuppressedAt's exclusion semantics.
func TestLiveHorizonExcludesGatedClass(t *testing.T) {
	// Only conversation resolves; planner and meeting are gated out.
	standings := LiveHorizon(32, func(class string) (float64, bool) {
		if class == "conversation" {
			return BootstrapLocalSecPerPt, true
		}
		return 0, false
	})
	if len(standings) != 1 || standings[0].Class != "conversation" {
		t.Fatalf("expected only the conversation entry, got %v", standings)
	}
}

// TestLiveHorizonUncappedSuppressesAll: at uncapped max speed (tps <= 0) every
// included class is suppressed with Route's dedicated uncapped phrasing.
func TestLiveHorizonUncappedSuppressesAll(t *testing.T) {
	lookup := func(string) (float64, bool) { return BootstrapLocalSecPerPt, true }
	for _, tps := range []float64{0, -1} {
		for _, cs := range LiveHorizon(tps, lookup) {
			if !cs.Suppressed {
				t.Errorf("tps=%g class=%s: want suppressed at uncapped speed", tps, cs.Class)
			}
			if !strings.Contains(cs.Verdict.Arithmetic, "uncapped speed - suppressed") {
				t.Errorf("tps=%g class=%s: Arithmetic=%q, want Route's uncapped phrasing", tps, cs.Class, cs.Verdict.Arithmetic)
			}
		}
	}
}

// TestSuppressedAtEqualsLiveHorizonNames (spec 037 SC-002, data-model.md
// invariant): SuppressedAt's result is exactly the names of LiveHorizon
// entries with Suppressed==true, across the ladder × a mixed lookup (one class
// gated out) so the single-iteration equivalence holds under exclusion too.
func TestSuppressedAtEqualsLiveHorizonNames(t *testing.T) {
	lookup := func(class string) (float64, bool) {
		if class == "meeting" {
			return 0, false // gated out — must be absent from both surfaces
		}
		return BootstrapLocalSecPerPt, true
	}
	for _, tps := range append([]float64{0}, horizonLadder...) {
		names := SuppressedAt(tps, lookup)
		var want []string
		for _, cs := range LiveHorizon(tps, lookup) {
			if cs.Suppressed {
				want = append(want, cs.Class)
			}
		}
		if strings.Join(names, ",") != strings.Join(want, ",") {
			t.Errorf("tps=%g: SuppressedAt=%v, LiveHorizon suppressed names=%v", tps, names, want)
		}
	}
}
