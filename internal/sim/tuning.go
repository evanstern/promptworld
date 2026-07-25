package sim

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/evanstern/promptworld/internal/store"
)

// --- spec 048 world tuning manifest (tuning.json) ---
//
// A boot-loaded, clamp-validated, event-logged promotion path for doctrine
// constants (docs/design/control-surface-and-calibration.md §6, TASK-107).
// Five former constants are promoted to per-world dials. Every dial defaults
// to its current doctrine constant, so an absent tuning.json is byte-for-byte
// the pre-048 behavior: TuningState is a pointer on sim.State tagged
// json:"tuning,omitempty" and nil ≡ the default set.
//
// The default constants below are the SINGLE home for these doctrine values
// (relocated here from agents.go / gru.go / mind.go by spec 048). Reducer and
// mind-replica reads go through the nil-safe accessors on *State; only
// tuning.go, the two RNG-bucketing reuses in memory.go/social.go (deliberately
// pinned to the default cadence — they are not the "planner cadence" dial),
// and in-package tests reference the raw default constants.
const (
	// Reflex refuels a fire with < 3 game-hours of fuel left. Raised from 3600
	// (1 h) by spec 057 / TASK-108: world-01 built 8 fires and lost 42 to
	// burnout over 6 days (warmth 848→82, an exposure death) because the 1-hour
	// window left villagers racing burnout. This is a doctrine DEFAULT change,
	// not a dial-semantics change — the clamp range, manifest override, and
	// event discipline are unchanged, and a tuning.json value still wins. New
	// worlds pin this default at genesis (sim.tuning_applied), so the change
	// reaches only pre-057 worlds (intended live effect) and never rewrites the
	// replay of a post-057 world.
	defaultRefuelDyingBelow       = 10800    // 3 game-hours (spec 057; was 3600 = 1 h)
	defaultFireBurnPerWood        = 4 * 3600 // 4 game-hours of fuel per wood
	defaultGruEmergePerMille      = 600      // per-mille chance the gru emerges per night
	defaultPlannerCadenceTicks    = 1800     // 30 game-minutes: the mind driver's per-agent baseline
	defaultEncounterCooldownTicks = 2 * 3600 // 2 game-hours: per-pair encounter cooldown
)

// Clamp bounds per contracts/tuning.md. Out-of-range values of KNOWN fields
// clamp to the nearest bound with an operator-visible warning (the
// llm/config.go normalizeTokenBudget shape). Structural problems (malformed
// JSON, wrong types, unknown fields) fail boot instead (ParseTuning).
const (
	minRefuelDyingBelow, maxRefuelDyingBelow             = 0, 86400
	minFireBurnPerWood, maxFireBurnPerWood               = 600, 86400
	minGruEmergePerMille, maxGruEmergePerMille           = 0, 1000
	minPlannerCadenceTicks, maxPlannerCadenceTicks       = 60, 86400
	minEncounterCooldownTicks, maxEncounterCooldownTicks = 0, 86400
)

// TuningState is the fully-resolved effective dial set carried on sim.State
// (event-sourced). A non-nil TuningState always carries all five fields with
// defaults filled in and clamps applied — never a sparse/partial struct. nil
// means "all defaults" and is the state of every pre-048 world. This is the
// shape the sim.tuning_applied payload snapshots and the accessors read.
type TuningState struct {
	RefuelDyingBelow       int64  `json:"refuel_dying_below"`
	FireBurnPerWood        int64  `json:"fire_burn_per_wood"`
	GruEmergePerMille      uint64 `json:"gru_emerge_per_mille"`
	PlannerCadenceTicks    int64  `json:"planner_cadence_ticks"`
	EncounterCooldownTicks int64  `json:"encounter_cooldown_ticks"`
}

// defaultTuning returns the fully-resolved default dial set (every field equal
// to its doctrine constant). It is what a sparse manifest is resolved against
// and what nil TuningState is equivalent to.
func defaultTuning() TuningState {
	return TuningState{
		RefuelDyingBelow:       defaultRefuelDyingBelow,
		FireBurnPerWood:        defaultFireBurnPerWood,
		GruEmergePerMille:      defaultGruEmergePerMille,
		PlannerCadenceTicks:    defaultPlannerCadenceTicks,
		EncounterCooldownTicks: defaultEncounterCooldownTicks,
	}
}

// --- accessors (nil-safe, on *State) ---
//
// These are the ONLY consumption path for the promoted dials. Every reducer
// call site holds a *State; every mind call site reads them off its replica
// (md.replica / e.replica), which absorbs sim.tuning_applied like every other
// event. nil Tuning returns the default constant, making "absent file ==
// current constants" structurally true.

// RefuelDyingBelow is the remaining-fuel window under which the reflex refuels
// a known fire (game-seconds).
func (s *State) RefuelDyingBelow() int64 {
	if s.Tuning != nil {
		return s.Tuning.RefuelDyingBelow
	}
	return defaultRefuelDyingBelow
}

// FireBurnPerWood is the fuel (game-seconds) one wood adds to a fire, still
// truncated by the unpromoted fireFuelCap.
func (s *State) FireBurnPerWood() int64 {
	if s.Tuning != nil {
		return s.Tuning.FireBurnPerWood
	}
	return defaultFireBurnPerWood
}

// GruEmergePerMille is the per-mille chance the gru emerges on a given night.
func (s *State) GruEmergePerMille() uint64 {
	if s.Tuning != nil {
		return s.Tuning.GruEmergePerMille
	}
	return defaultGruEmergePerMille
}

// PlannerCadence is the mind driver's per-agent baseline cadence (game-seconds)
// that drives planner scheduling, boot stagger, and embedder bucket edges.
func (s *State) PlannerCadence() int64 {
	if s.Tuning != nil {
		return s.Tuning.PlannerCadenceTicks
	}
	return defaultPlannerCadenceTicks
}

// EncounterCooldown is the per-pair cooldown (game-seconds) gating a new
// planner encounter between two agents who recently converged.
func (s *State) EncounterCooldown() int64 {
	if s.Tuning != nil {
		return s.Tuning.EncounterCooldownTicks
	}
	return defaultEncounterCooldownTicks
}

// EffectiveTuning returns the full in-effect dial set (tuned values or their
// defaults); nil Tuning resolves to the default set. The daemon boot seed
// compares a parsed manifest against this — TuningState is comparable — to
// decide whether the effective set differs from what state already carries and
// thus whether to append a sim.tuning_applied event (spec 048 FR-004).
func (s *State) EffectiveTuning() TuningState {
	if s.Tuning != nil {
		return *s.Tuning
	}
	return defaultTuning()
}

// --- manifest parsing (spec 048 US1) ---

// tuningManifest is the sparse on-disk carrier: pointer fields so an absent key
// stays nil (resolved to the default) and a present key is distinguishable from
// a zero value. DisallowUnknownFields rejects typos; a wrong-typed value fails
// json.Decode with the field name.
type tuningManifest struct {
	RefuelDyingBelow       *int64  `json:"refuel_dying_below"`
	FireBurnPerWood        *int64  `json:"fire_burn_per_wood"`
	GruEmergePerMille      *uint64 `json:"gru_emerge_per_mille"`
	PlannerCadenceTicks    *int64  `json:"planner_cadence_ticks"`
	EncounterCooldownTicks *int64  `json:"encounter_cooldown_ticks"`
}

// ParseTuning decodes a sparse tuning.json into a full effective TuningState.
// It returns the resolved set, a slice of operator-facing clamp warnings (one
// per out-of-range known field, in the llm/config.go style), and an error for
// structural problems (malformed JSON, wrong types, unknown field names) that
// must fail boot (fail-closed — a typo'd dial must not silently no-op).
func ParseTuning(data []byte) (*TuningState, []string, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var m tuningManifest
	if err := dec.Decode(&m); err != nil {
		return nil, nil, fmt.Errorf("tuning.json: %w", err)
	}

	t := defaultTuning()
	var warns []string
	clampI64 := func(field string, raw *int64, dst *int64, min, max int64) {
		if raw == nil {
			return
		}
		v := *raw
		switch {
		case v < min:
			warns = append(warns, fmt.Sprintf("tuning.json %s %d out of range (min %d) — clamped to %d", field, v, min, min))
			v = min
		case v > max:
			warns = append(warns, fmt.Sprintf("tuning.json %s %d out of range (max %d) — clamped to %d", field, v, max, max))
			v = max
		}
		*dst = v
	}

	clampI64("refuel_dying_below", m.RefuelDyingBelow, &t.RefuelDyingBelow, minRefuelDyingBelow, maxRefuelDyingBelow)
	clampI64("fire_burn_per_wood", m.FireBurnPerWood, &t.FireBurnPerWood, minFireBurnPerWood, maxFireBurnPerWood)
	if m.GruEmergePerMille != nil {
		// per-mille is uint64: only a high-side clamp is reachable (the min is 0).
		v := *m.GruEmergePerMille
		if v > maxGruEmergePerMille {
			warns = append(warns, fmt.Sprintf("tuning.json gru_emerge_per_mille %d out of range (max %d) — clamped to %d", v, uint64(maxGruEmergePerMille), uint64(maxGruEmergePerMille)))
			v = maxGruEmergePerMille
		}
		t.GruEmergePerMille = v
	}
	clampI64("planner_cadence_ticks", m.PlannerCadenceTicks, &t.PlannerCadenceTicks, minPlannerCadenceTicks, maxPlannerCadenceTicks)
	clampI64("encounter_cooldown_ticks", m.EncounterCooldownTicks, &t.EncounterCooldownTicks, minEncounterCooldownTicks, maxEncounterCooldownTicks)

	return &t, warns, nil
}

// --- event constructor (spec 048 US2) ---

// TuningAppliedPayload is the sim.tuning_applied event body: the full effective
// dial set (never a delta), so replay can establish tuning from any single
// event without scanning history.
type TuningAppliedPayload struct {
	RefuelDyingBelow       int64  `json:"refuel_dying_below"`
	FireBurnPerWood        int64  `json:"fire_burn_per_wood"`
	GruEmergePerMille      uint64 `json:"gru_emerge_per_mille"`
	PlannerCadenceTicks    int64  `json:"planner_cadence_ticks"`
	EncounterCooldownTicks int64  `json:"encounter_cooldown_ticks"`
}

// GenesisTuningEvent builds the sim.tuning_applied event a world pins at
// creation (spec 057 / TASK-108 US2): the FULL current default dial set, so a
// world's effective doctrine is fixed in its own log at birth and later changes
// to any default* constant never reach back into its replay. `promptworld new`
// seeds exactly one of these among its genesis events; the daemon boot seed
// (seedTuning) then compares a tuning.json against this pinned set exactly as
// spec 048 specifies (no event on equality, one on difference). migrate does
// NOT back-fill this — pre-057 and migrated worlds carry no pin and follow
// compiled defaults (a documented determinism hazard, §6).
func GenesisTuningEvent(tick int64) store.Event {
	return NewTuningEvent(tick, defaultTuning())
}

// NewTuningEvent builds the sim.tuning_applied event carrying the full
// effective set, for the daemon to seed on boot when it differs from what
// state already carries (the NewConventionEvent pattern, governance.go:632).
func NewTuningEvent(tick int64, t TuningState) store.Event {
	return store.Event{Tick: tick, Type: "sim.tuning_applied",
		Payload: mustPayload(TuningAppliedPayload{
			RefuelDyingBelow:       t.RefuelDyingBelow,
			FireBurnPerWood:        t.FireBurnPerWood,
			GruEmergePerMille:      t.GruEmergePerMille,
			PlannerCadenceTicks:    t.PlannerCadenceTicks,
			EncounterCooldownTicks: t.EncounterCooldownTicks,
		})}
}
