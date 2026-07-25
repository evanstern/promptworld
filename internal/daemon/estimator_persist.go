package daemon

import (
	"context"
	"time"

	"github.com/evanstern/promptworld/internal/cognition"
)

// estimatorPersistCadence is how often the daemon flushes live per-provider
// seconds-per-point estimates to estimator_state.json (TASK-113): frequent
// enough that a crash between flushes loses at most one cadence of learned
// drift, infrequent enough it is not a hot dial and touches disk rarely. Not a
// doctrine constant in the cognition sense (it governs I/O cadence, not
// routing arithmetic), so it lives here rather than in internal/cognition.
const estimatorPersistCadence = 5 * time.Minute

// estimatorSource is the daemon's read seam onto the orchestrator's live
// estimates — narrow so tests can persist from a fake without a live model,
// mirroring pendingSource/statusSource above.
type estimatorSource interface {
	SnapshotEstimators() map[string]float64
}

// estimatorPersister is the daemon-owned goroutine that writes the live
// estimator snapshot to path every estimatorPersistCadence; Run additionally
// calls Flush once, synchronously, after the sim loop stops — so learned s/pt
// survives both a crash (periodic flush) and a clean shutdown (final flush),
// closing the world-01 amnesia gap (control-surface report §4 row 3, §7:
// estimator process-lifetime + 36 restarts = 36 resets to the optimistic
// floor). Constructed only when an LLM orchestrator exists — a no-LLM world
// runs no persister and writes no file.
type estimatorPersister struct {
	src  estimatorSource
	path string
	// now is the persisted saved_at clock, a field (not time.Now directly) so
	// tests can freeze it.
	now func() time.Time
}

func newEstimatorPersister(src estimatorSource, path string) *estimatorPersister {
	return &estimatorPersister{src: src, path: path, now: time.Now}
}

// run flushes every estimatorPersistCadence until ctx is canceled, then
// returns — Run's own final Flush call after loop.Run(ctx) covers shutdown,
// so this goroutine does not need its own last-gasp write on cancel.
func (p *estimatorPersister) run(ctx context.Context) {
	ticker := time.NewTicker(estimatorPersistCadence)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = p.Flush()
		}
	}
}

// Flush snapshots the orchestrator's live estimates and writes them as a
// full-file replace. Errors are the caller's to log — a failed persist never
// blocks or fails boot/shutdown (the same warn-not-fatal posture as every
// other boot-loaded config file in this package).
func (p *estimatorPersister) Flush() error {
	state := &cognition.EstimatorState{
		SavedAt:   p.now().UTC().Format(time.RFC3339),
		Providers: p.src.SnapshotEstimators(),
	}
	return state.Save(p.path)
}
