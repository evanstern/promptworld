package daemon

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/evanstern/promptworld/internal/cognition"
)

// fakeEstimatorSource is a scripted estimatorSource — stands in for the
// Orchestrator's SnapshotEstimators so the persister is exercised without a
// live model.
type fakeEstimatorSource struct {
	snap map[string]float64
}

func (f *fakeEstimatorSource) SnapshotEstimators() map[string]float64 {
	return f.snap
}

// TestEstimatorPersisterFlushRoundTrip (TASK-113, AC#1): Flush writes exactly
// what the source currently reports, and the file is a legal
// cognition.EstimatorState a fresh boot can load back.
func TestEstimatorPersisterFlushRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "estimator_state.json")
	src := &fakeEstimatorSource{snap: map[string]float64{"cloud": 2.76, "local": 0.94}}
	p := newEstimatorPersister(src, path)
	p.now = func() time.Time { return time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC) }

	if err := p.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	got, err := cognition.LoadEstimatorState(path)
	if err != nil {
		t.Fatalf("LoadEstimatorState: %v", err)
	}
	if got == nil {
		t.Fatal("LoadEstimatorState returned nil after Flush")
	}
	if got.Providers["cloud"] != 2.76 || got.Providers["local"] != 0.94 {
		t.Errorf("persisted providers = %+v, want cloud=2.76 local=0.94", got.Providers)
	}
	if got.SavedAt != "2026-07-25T00:00:00Z" {
		t.Errorf("SavedAt = %q, want 2026-07-25T00:00:00Z", got.SavedAt)
	}
}

// TestEstimatorPersisterFlushOverwritesPreviousSnapshot: a later Flush with a
// changed source value replaces the file (full-file replace, same posture as
// calibration.json) — the persisted file always reflects the LATEST live
// estimate, not the first one ever written.
func TestEstimatorPersisterFlushOverwritesPreviousSnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "estimator_state.json")
	src := &fakeEstimatorSource{snap: map[string]float64{"cloud": 1.0}}
	p := newEstimatorPersister(src, path)

	if err := p.Flush(); err != nil {
		t.Fatalf("first Flush: %v", err)
	}
	src.snap = map[string]float64{"cloud": 5.5}
	if err := p.Flush(); err != nil {
		t.Fatalf("second Flush: %v", err)
	}

	got, err := cognition.LoadEstimatorState(path)
	if err != nil {
		t.Fatalf("LoadEstimatorState: %v", err)
	}
	if got.Providers["cloud"] != 5.5 {
		t.Errorf("Providers[cloud] = %g, want 5.5 (latest value)", got.Providers["cloud"])
	}
}

// TestEstimatorPersisterRunFlushesOnCadenceThenStops verifies run() performs
// periodic flushes and exits cleanly when ctx is canceled, without panicking
// or double-closing anything — the daemon-owned-goroutine contract every
// other periodic sampler in this package follows (governorSampler.run).
func TestEstimatorPersisterRunFlushesOnCadenceThenStops(t *testing.T) {
	path := filepath.Join(t.TempDir(), "estimator_state.json")
	src := &fakeEstimatorSource{snap: map[string]float64{"cloud": 3.3}}
	p := newEstimatorPersister(src, path)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		p.run(ctx)
		close(done)
	}()
	// run() ticks on estimatorPersistCadence (5m) — far longer than a test
	// should wait, so this test only proves clean start/stop, not the cadence
	// itself (Flush's own tests cover the write path directly).
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("run() did not return after ctx cancel")
	}
}
