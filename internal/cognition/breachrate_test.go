package cognition

import "testing"

// TestBreachRateLoweredAdoptsOnFewerSpikes (TASK-113 optional stretch, taken
// as a SEPARATE commit from the estimator-persistence fix per the control-
// surface report's suggestion, §4 row 3): characterizes the measurable effect
// of BreachRate 0.3 -> 0.2 — the sustained-drift evidence bar drops from "at
// least 7 of the last WindowSize(20) samples spike" (0.35 > 0.3) to "at least
// 5" (0.25 > 0.2). This pins the new exact boundary: one fewer spike than the
// bar must NOT breach, the bar itself must. If BreachRate changes again, this
// test's own arithmetic (derived from the constants, not hardcoded) keeps it
// honest.
func TestBreachRateLoweredAdoptsOnFewerSpikes(t *testing.T) {
	minSpikes := 0
	for n := 1; n <= WindowSize; n++ {
		if float64(n)/float64(WindowSize) > BreachRate {
			minSpikes = n
			break
		}
	}
	if minSpikes == 0 {
		t.Fatal("setup: no spike count in [1, WindowSize] breaches BreachRate")
	}

	// One spike short of the bar: no breach — still one-shot territory.
	if fired := feedSpikePattern(t, minSpikes-1); fired {
		t.Errorf("%d/%d spikes (below the %v bar) breached; want no breach", minSpikes-1, WindowSize, BreachRate)
	}

	// Exactly at the bar: breach fires.
	if fired := feedSpikePattern(t, minSpikes); !fired {
		t.Errorf("%d/%d spikes (at the %v bar) did not breach; want breach", minSpikes, WindowSize, BreachRate)
	}
}

// feedSpikePattern fills exactly one full window (WindowSize samples) with
// nSpikes spikes (100.0, > SpikeFactor*seed) front-loaded and the remainder
// normal (10.0), and reports whether any Sample call returned a non-nil
// Adoption — i.e. whether the window's spike rate breached BreachRate.
func feedSpikePattern(t *testing.T, nSpikes int) bool {
	t.Helper()
	e := NewEstimator(10.0)
	fired := false
	for i := 0; i < WindowSize; i++ {
		v := 10.0
		if i < nSpikes {
			v = 100.0
		}
		if e.Sample(v) != nil {
			fired = true
		}
	}
	return fired
}
