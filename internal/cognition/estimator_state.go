package cognition

import (
	"encoding/json"
	"fmt"
	"os"
)

// EstimatorState is the persisted snapshot of every provider's live
// seconds-per-point estimate (TASK-113, estimator_state.json in the world
// dir). It follows calibration.json's discipline as a boot-loaded config file
// — full-file replace, read once at daemon start — with one deliberate
// difference: calibration.json is written only by a human running
// `promptworld calibrate`, while this file IS daemon-written (periodically and
// at shutdown), because its whole purpose is capturing live EWMA drift a
// human never observes directly. It never enters the event log and is never
// read during replay: the estimator sits outside internal/sim's reducer (this
// package is imported BY internal/mind/internal/llm, never the reverse), so
// its presence, absence, or contents change only future routing/throttle
// predictions, never simulated world state (control-surface report §7).
type EstimatorState struct {
	SavedAt   string             `json:"saved_at"`
	Providers map[string]float64 `json:"providers"` // provider name -> seconds-per-point
}

// LoadEstimatorState reads estimator_state.json. A missing file returns
// (nil, nil) — legal; the caller reseeds from calibration/bootstrap alone,
// exactly as if this file never existed. A malformed file returns an error the
// caller downgrades to a warning, never a crash — the same posture
// LoadProfile takes with calibration.json.
func LoadEstimatorState(path string) (*EstimatorState, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("estimator state: %w", err)
	}
	var s EstimatorState
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("estimator state %s: %w", path, err)
	}
	return &s, nil
}

// Save writes the state as a full-file replace, the same shape as
// Profile.Save.
func (s *EstimatorState) Save(path string) error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

// ReseedValue is the AC#2 contract: max(calibration/bootstrap seed, persisted
// estimate) for one provider. calSeed is whatever SeedFor already resolved
// (a recorded calibration.json entry, or a bootstrap default) — this only
// ever raises that seed, when a persisted live estimate exceeds it, never
// lowers it below a fresher human calibration or the pessimistic bootstrap
// floor. A nil state, or no persisted entry for the named provider, returns
// calSeed unchanged.
func ReseedValue(calSeed float64, state *EstimatorState, name string) float64 {
	if state == nil {
		return calSeed
	}
	if v, ok := state.Providers[name]; ok && v > calSeed {
		return v
	}
	return calSeed
}
