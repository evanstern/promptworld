package cognition

import (
	"os"
	"path/filepath"
	"testing"
)

// TestEstimatorStateRoundTrip (TASK-113): Save then Load reproduces the exact
// per-provider seconds-per-point map — the persistence half of the AC#1
// contract ("learned s/pt survives a daemon restart").
func TestEstimatorStateRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "estimator_state.json")
	s := &EstimatorState{
		SavedAt:   "2026-07-25T00:00:00Z",
		Providers: map[string]float64{"cloud": 2.76, "gemma": 0.9, "cogito": 1.48},
	}
	if err := s.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := LoadEstimatorState(path)
	if err != nil {
		t.Fatalf("LoadEstimatorState: %v", err)
	}
	if got.SavedAt != s.SavedAt {
		t.Errorf("SavedAt = %q, want %q", got.SavedAt, s.SavedAt)
	}
	for name, want := range s.Providers {
		if got.Providers[name] != want {
			t.Errorf("Providers[%q] = %g, want %g", name, got.Providers[name], want)
		}
	}
}

// TestEstimatorStateMissingIsLegal mirrors LoadProfile's missing-file posture:
// no persisted history yet is legal, not an error — boot reseeds from
// calibration/bootstrap alone.
func TestEstimatorStateMissingIsLegal(t *testing.T) {
	s, err := LoadEstimatorState(filepath.Join(t.TempDir(), "estimator_state.json"))
	if err != nil || s != nil {
		t.Errorf("missing file: s=%v err=%v; want nil, nil", s, err)
	}
}

// TestEstimatorStateMalformedErrorsWithoutPanic mirrors LoadProfile's
// malformed-file posture: an error the caller downgrades to a warning, never
// a crash.
func TestEstimatorStateMalformedErrorsWithoutPanic(t *testing.T) {
	path := filepath.Join(t.TempDir(), "estimator_state.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadEstimatorState(path); err == nil {
		t.Error("malformed file loaded without error")
	}
}

// TestReseedValueTakesMax (AC#2): reseed is max(calibration/bootstrap seed,
// persisted estimate) — a persisted value only ever RAISES the seed, never
// lowers it below a fresher calibration or the pessimistic bootstrap floor.
func TestReseedValueTakesMax(t *testing.T) {
	state := &EstimatorState{Providers: map[string]float64{"cloud": 5.0}}

	if got := ReseedValue(1.2, state, "cloud"); got != 5.0 {
		t.Errorf("persisted (5.0) > calSeed (1.2): got %g, want 5.0", got)
	}
	if got := ReseedValue(9.0, state, "cloud"); got != 9.0 {
		t.Errorf("calSeed (9.0) > persisted (5.0): got %g, want 9.0 (never lowered)", got)
	}
	if got := ReseedValue(3.0, state, "gemma"); got != 3.0 {
		t.Errorf("no persisted entry for %q: got %g, want calSeed 3.0 unchanged", "gemma", got)
	}
	if got := ReseedValue(3.0, nil, "cloud"); got != 3.0 {
		t.Errorf("nil state: got %g, want calSeed 3.0 unchanged", got)
	}
}

// TestRestartStormNoLongerReproduces (AC#3, regression-shaped — world-01 ran
// this exact daemon and can't be replayed here, so this pins the mechanism
// the control-surface report diagnosed: docs/design/control-surface-and-
// calibration.md §4 row 3 — calibrate measures a single-call floor (~10s/pt
// here, matching BootstrapCloudSecPerPt's order of magnitude) while live
// contention runs ~6x higher (~60s/pt, matching world-01's cloud 0.449→2.76
// s/pt ~6x drift); a process-lifetime estimator re-seeded from the floor on
// every restart re-enters the spike-rate breach (cog.recalibration_recommended)
// each time, 92 times across world-01's 36 restarts.
//
// Without persistence, EVERY restart repeats the SAME climb from the floor:
// this test drives that climb twice from a fresh NewEstimator(calSeed) and
// shows it fires an Adoption both times — the storm reproducing on schedule.
// With persistence, a restart reseeds from max(calSeed, persisted) instead:
// this test shows a fresh estimator seeded that way absorbs a full window of
// the SAME live load with zero adoptions — already living at the true value,
// not climbing back up to it.
func TestRestartStormNoLongerReproduces(t *testing.T) {
	const calSeed = 10.0 // the calibration/bootstrap floor
	const live = 60.0    // sustained live contention, ~6x the floor

	// climbFromFloor seeds fresh at calSeed (today's process-lifetime-only
	// behavior on every restart) and feeds one full window of live load,
	// reporting how many adoption episodes fired.
	climbFromFloor := func() (adoptions int, final float64) {
		e := NewEstimator(calSeed)
		for i := 0; i < WindowSize; i++ {
			if e.Sample(live) != nil {
				adoptions++
			}
		}
		return adoptions, e.Estimate()
	}

	// Restart #1 (pre-persistence world, and also world-01's very first climb):
	// the estimator has never seen this load before, so an adoption climbing
	// from the floor to the live value is EXPECTED and correct — not the bug.
	adoptions1, learned := climbFromFloor()
	if adoptions1 == 0 {
		t.Fatal("setup: first climb from the floor produced no adoption — test's live/calSeed gap is too small to be a spike")
	}

	// Restart #2 WITHOUT persistence: today's actual behavior. The estimator
	// forgot everything restart #1 learned and climbs from the floor again —
	// this IS the storm (repeats every restart; world-01 hit it 92 times
	// across 36 restarts).
	adoptions2, _ := climbFromFloor()
	if adoptions2 == 0 {
		t.Fatal("regression baseline broke: restart WITHOUT persistence should still re-climb from the floor")
	}

	// Restart #2 WITH persistence (TASK-113 fix): the daemon persisted
	// `learned` from restart #1 and reseeds via ReseedValue(calSeed, state, name)
	// instead of NewEstimator(calSeed) directly.
	state := &EstimatorState{Providers: map[string]float64{"cloud": learned}}
	reseed := ReseedValue(calSeed, state, "cloud")
	if reseed != learned {
		t.Fatalf("reseed = %g, want the persisted learned value %g", reseed, learned)
	}
	e3 := NewEstimator(reseed)
	fixedAdoptions := 0
	for i := 0; i < WindowSize; i++ {
		if e3.Sample(live) != nil {
			fixedAdoptions++
		}
	}
	if fixedAdoptions != 0 {
		t.Errorf("estimator reseeded from persisted state fired %d adoption(s) absorbing the SAME live load it already learned — the restart storm still reproduces", fixedAdoptions)
	}
	if got := e3.Estimate(); got < live*0.9 || got > live*1.1 {
		t.Errorf("estimate after reseed+one window = %g, want it to stay near the live value %g (not re-climb from the floor)", got, live)
	}
}
