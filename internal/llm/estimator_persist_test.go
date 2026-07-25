package llm

import (
	"testing"

	"github.com/evanstern/promptworld/internal/cognition"
)

// TestSeedPersistedRaisesAboveCalibrationSeed (TASK-113, AC#2): after
// SeedCalibration seeds each provider from a calibration profile,
// SeedPersisted raises a provider's estimate to its persisted value only
// when that value exceeds the calibration seed — never lowers it.
func TestSeedPersistedRaisesAboveCalibrationSeed(t *testing.T) {
	o := newOrch(t, testConfig("http://127.0.0.1:1", "http://127.0.0.1:1", 100), testStore(t))
	profile := &cognition.Profile{
		CalibratedAt: "2026-07-20T21:40:00Z",
		Tiers: map[string]cognition.TierProfile{
			"local": {SecondsPerPoint: 0.94},
			"cloud": {SecondsPerPoint: 1.2},
		},
	}
	o.SeedCalibration(profile)

	state := &cognition.EstimatorState{
		SavedAt: "2026-07-24T12:00:00Z",
		Providers: map[string]float64{
			"cloud": 2.76, // above the calibration seed 1.2 — must raise
			"local": 0.5,  // below the calibration seed 0.94 — must NOT lower
		},
	}
	o.SeedPersisted(state)

	if got, _, _, _ := o.providers["cloud"].est.Stats(); got != 2.76 {
		t.Errorf("cloud estimate = %g, want persisted 2.76 (raised above calibration seed 1.2)", got)
	}
	if got, _, _, _ := o.providers["local"].est.Stats(); got != 0.94 {
		t.Errorf("local estimate = %g, want calibration seed 0.94 unchanged (persisted 0.5 is lower)", got)
	}
}

// TestSeedPersistedNilIsNoop: SeedPersisted(nil) — the no-persisted-history
// boot path — leaves every provider's estimate exactly as SeedCalibration (or
// bootstrap, if never called) left it.
func TestSeedPersistedNilIsNoop(t *testing.T) {
	o := newOrch(t, testConfig("http://127.0.0.1:1", "http://127.0.0.1:1", 100), testStore(t))
	before, _, _, _ := o.providers["local"].est.Stats()
	o.SeedPersisted(nil)
	after, _, _, _ := o.providers["local"].est.Stats()
	if before != after {
		t.Errorf("SeedPersisted(nil) changed local estimate: %g -> %g", before, after)
	}
}

// TestSeedPersistedNoEntryForProvider: a persisted state that names no entry
// for a given provider leaves that provider on its calibration/bootstrap seed.
func TestSeedPersistedNoEntryForProvider(t *testing.T) {
	o := newOrch(t, testConfig("http://127.0.0.1:1", "http://127.0.0.1:1", 100), testStore(t))
	before, _, _, _ := o.providers["local"].est.Stats()

	state := &cognition.EstimatorState{Providers: map[string]float64{"cloud": 99.0}}
	o.SeedPersisted(state)

	after, _, _, _ := o.providers["local"].est.Stats()
	if before != after {
		t.Errorf("local estimate changed with no persisted entry for it: %g -> %g", before, after)
	}
}

// TestSnapshotEstimatorsReflectsSeeding (TASK-113): SnapshotEstimators — the
// value the daemon persists — reports exactly what SeedCalibration/
// SeedPersisted last set, keyed by provider name.
func TestSnapshotEstimatorsReflectsSeeding(t *testing.T) {
	o := newOrch(t, testConfig("http://127.0.0.1:1", "http://127.0.0.1:1", 100), testStore(t))
	profile := &cognition.Profile{
		CalibratedAt: "2026-07-20T21:40:00Z",
		Tiers: map[string]cognition.TierProfile{
			"local": {SecondsPerPoint: 0.94},
			"cloud": {SecondsPerPoint: 1.2},
		},
	}
	o.SeedCalibration(profile)

	snap := o.SnapshotEstimators()
	if snap["local"] != 0.94 {
		t.Errorf("snapshot local = %g, want 0.94", snap["local"])
	}
	if snap["cloud"] != 1.2 {
		t.Errorf("snapshot cloud = %g, want 1.2", snap["cloud"])
	}
}
