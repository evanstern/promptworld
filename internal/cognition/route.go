package cognition

import "fmt"

// Verdict is the router's decision plus the arithmetic that produced it —
// recorded verbatim in suppression telemetry so every gate is auditable.
type Verdict struct {
	Allow               bool
	Class               string
	Points              int
	PredictedWallMs     int64
	PredictedDriftTicks int64
	BudgetTicks         int64
	Arithmetic          string
}

// Route decides whether a decision may go to the model: pure arithmetic over
// the class's registered values, the calibrated seconds-per-point, and the
// current speed in game ticks per real second. No model, no randomness, no
// wall-clock reads (FR-007). ticksPerSecond <= 0 (uncapped max speed) always
// suppresses — prediction at unbounded speed is meaningless; the existing
// refusal to run max speed with an LLM keeps this branch theoretical.
func Route(dc DecisionClass, ticksPerSecond, secondsPerPoint float64) Verdict {
	v := Verdict{Class: dc.Class, Points: dc.Points, BudgetTicks: dc.BudgetTicks}
	wallSec := float64(dc.Points) * secondsPerPoint
	v.PredictedWallMs = int64(wallSec * 1000)
	if ticksPerSecond <= 0 {
		v.Arithmetic = fmt.Sprintf("%dpt x %.1fs/pt at uncapped speed - suppressed", dc.Points, secondsPerPoint)
		return v
	}
	v.PredictedDriftTicks = int64(wallSec * ticksPerSecond)
	v.Allow = v.PredictedDriftTicks <= dc.BudgetTicks
	rel := "<="
	if !v.Allow {
		rel = ">"
	}
	v.Arithmetic = fmt.Sprintf("%dpt x %.1fs/pt x %gx = %d ticks %s budget %d",
		dc.Points, secondsPerPoint, ticksPerSecond, v.PredictedDriftTicks, rel, dc.BudgetTicks)
	return v
}

// RoutePaused is Route's paused counterpart (spec 040 D2, decision-6): a frozen
// world cannot drift, so predicted drift is zero — within every staleness budget
// — and the verdict allows every class, whatever the SET speed (including
// uncapped max). Wall time still passes while frozen, so PredictedWallMs is
// predicted exactly as Route does it; only the drift math is replaced. The
// arithmetic names the paused state (FR-005) in place of the set-speed drift
// string. Pure — the mind supplies the paused fact from its event-reduced
// replica, never a wall-clock read (FR-006). Route is left untouched (SC-005).
func RoutePaused(dc DecisionClass, secondsPerPoint float64) Verdict {
	wallSec := float64(dc.Points) * secondsPerPoint
	return Verdict{
		Allow:               true,
		Class:               dc.Class,
		Points:              dc.Points,
		PredictedWallMs:     int64(wallSec * 1000),
		PredictedDriftTicks: 0,
		BudgetTicks:         dc.BudgetTicks,
		Arithmetic: fmt.Sprintf("%dpt x %.1fs/pt while paused = 0 ticks <= budget %d",
			dc.Points, secondsPerPoint, dc.BudgetTicks),
	}
}
