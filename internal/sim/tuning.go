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

	// Spec 097 (perception of absence, FR-004): the four grounded-observation
	// dials. Doctrine defaults, human-tuned only:
	//   - dedup window 2 game-hours: pacing back and forth over one spot inside
	//     a working stretch is one observation, but a morning and an evening
	//     visit are two.
	//   - base salience 2: below salTalk (3) — background texture that never
	//     crowds the working window; the mind-side surprise bump (D4) is what
	//     promotes a disconfirming observation.
	//   - disconfirm retain 70%: each disconfirming visit keeps 70% of the
	//     belief's effective confidence — faster than the 8-game-day silence
	//     half-life but bounded, so a myth survives several visits before
	//     trending under the confidence floor (dials, not cliffs — D3).
	//   - confirm boost +10: a confirming visit adds 10 effective-confidence
	//     points (capped at 100) and re-anchors the decay clock.
	defaultObservationDedupTicks         = 2 * 3600 // 2 game-hours
	defaultObservationBaseSalience       = 2
	defaultBeliefDisconfirmRetainPercent = 70
	defaultBeliefConfirmBoost            = 10

	// Spec 102 (guardian agentization): the steward cadence dial — how often
	// (game-seconds) the agentized guardian's scheduled cognition lane comes
	// due. 0 is the OFF switch and the DEFAULT: agentization is opt-in per
	// world (FR-007), so an absent dial leaves every pre-102 world's guardian
	// purely event-driven, byte-identical to before. A nonzero value opts the
	// world in and sets the cadence (clamped to [min,max] below — never
	// hotter than 10 game-minutes).
	defaultStewardCadenceTicks = 0

	// Spec 104 (ambient event coalescing, FR-008): the needs-checkpoint
	// cadence — agent.needs_changed emits every K game-minutes per living
	// agent plus immediately on any danger-band/near-death/zero crossing.
	// K=1 reproduces today's per-minute emission byte-for-byte (the escape
	// hatch). The field doubles as the COALESCING REGIME marker: a recorded
	// sim.tuning_applied payload that lacks it resolves to 0 = legacy — the
	// executor keeps per-step agent.moved / per-minute needs / gru.moved
	// emission and the derived-advancement engine (advance.go) stays
	// structurally inert, so pre-104 logs and snapshots fold to
	// hash-identical state. New worlds pin K in their genesis tuning event
	// (defaultTuning below, spec 057), turning the regime on from tick 0.
	defaultNeedsCheckpointMinutes = 10
)

// The spec-098 dream dials (private dreams — consolidation clustering +
// habituation, dream.go). Human-tuned doctrine defaults, recorded with their
// rationale in specs/098-private-dreams/spec.md (D2/D4):
//   - density 900‰: cosine 0.90 reads as a near-duplicate neighborhood on the
//     spec-042 sentence embeddings (identical texts embed at 1.0);
//   - band 30‰: membership gathers at 0.87 and geometry alone decides only at
//     0.93+ — between the bars the existing consolidation slot is consulted;
//   - habituation 500‰: a routine cluster member's salience halves per night
//     (floor 1), the memory-recency half-life rhyme;
//   - merge cap 4: at most four absorbed members per agent-night, so a
//     collapse is gradual and each night's batch stays small;
//   - jitter 15‰: D4's minimal dream-noise adoption — a ±0.015 boundary
//     nudge, rngAt-seeded and zeroable (0 = pre-noise outcomes exactly).
const (
	defaultDreamDensityPerMille       = 900
	defaultDreamAmbiguousBandPerMille = 30
	defaultDreamHabituationPerMille   = 500
	defaultDreamMergeCapPerNight      = 4
	defaultDreamJitterPerMille        = 15
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
	// Spec 097 dials. Dedup 0 = every arrival observes; salience stays inside
	// the 1..10 memory band; the two belief dials are percentages/points.
	minObservationDedupTicks, maxObservationDedupTicks                 = 0, 86400
	minObservationBaseSalience, maxObservationBaseSalience             = 1, 10
	minBeliefDisconfirmRetainPercent, maxBeliefDisconfirmRetainPercent = 0, 100
	minBeliefConfirmBoost, maxBeliefConfirmBoost                       = 0, 100
	// Spec 104: the manifest can never author the 0 legacy sentinel — the
	// floor is 1 (per-minute), so "legacy" is reachable only from pre-104
	// recorded payloads.
	minNeedsCheckpointMinutes, maxNeedsCheckpointMinutes = 1, 60
	// Dream dials (spec 098): per-mille geometry bounds; the band is capped at
	// 500 so the membership bar can never go negative against a mid-range
	// density; the merge cap is bounded well under any plausible store size;
	// jitter is capped at 200‰ so noise can widen dreams but never dominate.
	minDreamDensityPerMille, maxDreamDensityPerMille             = 0, 1000
	minDreamAmbiguousBandPerMille, maxDreamAmbiguousBandPerMille = 0, 500
	minDreamHabituationPerMille, maxDreamHabituationPerMille     = 0, 1000
	minDreamMergeCapPerNight, maxDreamMergeCapPerNight           = 0, 64
	minDreamJitterPerMille, maxDreamJitterPerMille               = 0, 200
	// Steward cadence (spec 102): 0 = off (the opt-in switch); a NONZERO value
	// clamps to this band — floor 600 (10 game-minutes; hotter would let one
	// guardian outdraw the whole village's planner budget), ceiling one game
	// day. The 0-vs-band split is enforced by a dedicated clamp in
	// ParseTuning, not clampI64.
	minStewardCadenceTicks, maxStewardCadenceTicks = 600, 86400
)

// TuningState is the fully-resolved effective dial set carried on sim.State
// (event-sourced). A non-nil TuningState always carries all fields with
// defaults filled in and clamps applied — never a sparse/partial struct. nil
// means "all defaults" and is the state of every pre-048 world. This is the
// shape the sim.tuning_applied payload snapshots and the accessors read.
// Spec 097 adds the four grounded-observation dials; a pre-097 tuning_applied
// payload (fields absent) resolves them to the doctrine defaults at Apply
// (state.go), never to zero.
type TuningState struct {
	RefuelDyingBelow       int64  `json:"refuel_dying_below"`
	FireBurnPerWood        int64  `json:"fire_burn_per_wood"`
	GruEmergePerMille      uint64 `json:"gru_emerge_per_mille"`
	PlannerCadenceTicks    int64  `json:"planner_cadence_ticks"`
	EncounterCooldownTicks int64  `json:"encounter_cooldown_ticks"`
	// Spec 097 (perception of absence, FR-004).
	ObservationDedupTicks         int64 `json:"observation_dedup_ticks"`
	ObservationBaseSalience       int64 `json:"observation_base_salience"`
	BeliefDisconfirmRetainPercent int64 `json:"belief_disconfirm_retain_percent"`
	BeliefConfirmBoost            int64 `json:"belief_confirm_boost"`
	// StewardCadenceTicks (spec 102) is the guardian-agentization opt-in dial:
	// 0 = off (the default — every pre-102 world), nonzero = the scheduled
	// angel lane's cadence in game-seconds. omitempty keeps every pre-102
	// snapshot and recorded payload byte-identical (the spec-094 additive
	// discipline, no format bump).
	StewardCadenceTicks int64 `json:"steward_cadence_ticks,omitempty"`
	// NeedsCheckpointMinutes (spec 104, FR-008): the needs-checkpoint cadence
	// K, doubling as the coalescing-regime marker — 0 means LEGACY (a pre-104
	// recorded payload; per-step/per-minute emission, advancement inert),
	// never authorable from a manifest (clamp floor 1). omitempty keeps a
	// legacy TuningState's canonical bytes identical to pre-104.
	NeedsCheckpointMinutes int64 `json:"needs_checkpoint_minutes,omitempty"`
	// Dream (spec 098) is the private-dream dial block. nil ≡ the default
	// dream set — exactly the State.Tuning nil convention one level down —
	// which keeps every pre-098 snapshot and recorded sim.tuning_applied
	// payload byte-identical (omitempty; the spec-094 additive discipline,
	// no format-version bump). A non-nil block is always fully resolved.
	// NOTE: the pointer makes bare == on TuningState compare pointer
	// identity — dial comparisons go through Equal, never ==.
	Dream *DreamTuning `json:"dream,omitempty"`
}

// DreamTuning is the resolved spec-098 dream dial block (dream.go's
// PlanDream consumes it): the cluster density threshold and ambiguous band
// (per-mille cosine), the habituation weight factor (per-mille), the
// per-night merge cap, and the D4 boundary-jitter amplitude (per-mille,
// zeroable). A non-nil DreamTuning always carries all five fields resolved
// and clamped, mirroring TuningState's own never-sparse rule.
type DreamTuning struct {
	DensityPerMille       int64 `json:"density_per_mille"`
	AmbiguousBandPerMille int64 `json:"ambiguous_band_per_mille"`
	HabituationPerMille   int64 `json:"habituation_per_mille"`
	MergeCapPerNight      int64 `json:"merge_cap_per_night"`
	JitterPerMille        int64 `json:"jitter_per_mille"`
}

// defaultDream returns the fully-resolved default dream dial block.
func defaultDream() DreamTuning {
	return DreamTuning{
		DensityPerMille:       defaultDreamDensityPerMille,
		AmbiguousBandPerMille: defaultDreamAmbiguousBandPerMille,
		HabituationPerMille:   defaultDreamHabituationPerMille,
		MergeCapPerNight:      defaultDreamMergeCapPerNight,
		JitterPerMille:        defaultDreamJitterPerMille,
	}
}

// EffectiveDream resolves the dream block: the carried values, or the
// default set when the block is nil (every pre-098 TuningState).
func (t TuningState) EffectiveDream() DreamTuning {
	if t.Dream != nil {
		return *t.Dream
	}
	return defaultDream()
}

// Equal compares two effective dial sets BY VALUE — the comparison the boot
// seed uses (daemon seedTuning). Bare == would compare the Dream pointer's
// identity and re-append a redundant sim.tuning_applied on every restart.
func (t TuningState) Equal(o TuningState) bool {
	return t.RefuelDyingBelow == o.RefuelDyingBelow &&
		t.FireBurnPerWood == o.FireBurnPerWood &&
		t.GruEmergePerMille == o.GruEmergePerMille &&
		t.PlannerCadenceTicks == o.PlannerCadenceTicks &&
		t.EncounterCooldownTicks == o.EncounterCooldownTicks &&
		// Spec 097: the four grounded-observation dials compare by value like
		// the base five.
		t.ObservationDedupTicks == o.ObservationDedupTicks &&
		t.ObservationBaseSalience == o.ObservationBaseSalience &&
		t.BeliefDisconfirmRetainPercent == o.BeliefDisconfirmRetainPercent &&
		t.BeliefConfirmBoost == o.BeliefConfirmBoost &&
		// Spec 102: the angel cadence compares by value like the base five.
		t.StewardCadenceTicks == o.StewardCadenceTicks &&
		t.NeedsCheckpointMinutes == o.NeedsCheckpointMinutes &&
		t.EffectiveDream() == o.EffectiveDream()
}

// defaultTuning returns the fully-resolved default dial set (every field equal
// to its doctrine constant). It is what a sparse manifest is resolved against
// and what nil TuningState is equivalent to.
func defaultTuning() TuningState {
	return TuningState{
		RefuelDyingBelow:              defaultRefuelDyingBelow,
		FireBurnPerWood:               defaultFireBurnPerWood,
		GruEmergePerMille:             defaultGruEmergePerMille,
		PlannerCadenceTicks:           defaultPlannerCadenceTicks,
		EncounterCooldownTicks:        defaultEncounterCooldownTicks,
		ObservationDedupTicks:         defaultObservationDedupTicks,
		ObservationBaseSalience:       defaultObservationBaseSalience,
		BeliefDisconfirmRetainPercent: defaultBeliefDisconfirmRetainPercent,
		BeliefConfirmBoost:            defaultBeliefConfirmBoost,
		NeedsCheckpointMinutes:        defaultNeedsCheckpointMinutes,
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

// ObservationDedupTicks is the window (game-seconds) inside which a repeat
// observation of an unchanged place collapses — no event, no memory (spec 097
// D4). 0 = every arrival observes.
func (s *State) ObservationDedupTicks() int64 {
	if s.Tuning != nil {
		return s.Tuning.ObservationDedupTicks
	}
	return defaultObservationDedupTicks
}

// ObservationBaseSalience is the salience an arrival-observation memory
// enters at (spec 097 D4) — low, so grounded texture never crowds the
// working-memory window.
func (s *State) ObservationBaseSalience() int64 {
	if s.Tuning != nil {
		return s.Tuning.ObservationBaseSalience
	}
	return defaultObservationBaseSalience
}

// BeliefDisconfirmRetainPercent is the share (0–100) of a belief's effective
// confidence retained per disconfirming observation (spec 097 D3): the
// mind-side reconciler reads it off its replica. Lower = myths die faster.
func (s *State) BeliefDisconfirmRetainPercent() int64 {
	if s.Tuning != nil {
		return s.Tuning.BeliefDisconfirmRetainPercent
	}
	return defaultBeliefDisconfirmRetainPercent
}

// BeliefConfirmBoost is the effective-confidence points a confirming
// observation adds to a matching belief, capped at 100 (spec 097 D3).
func (s *State) BeliefConfirmBoost() int64 {
	if s.Tuning != nil {
		return s.Tuning.BeliefConfirmBoost
	}
	return defaultBeliefConfirmBoost
}

// StewardCadence is the guardian-agentization opt-in dial (spec 102): the
// scheduled angel lane's cadence in game-seconds, 0 = the lane is OFF (the
// default, and every pre-102 world). Read by the guardian off its replica —
// the mind-side dial discipline.
func (s *State) StewardCadence() int64 {
	if s.Tuning != nil {
		return s.Tuning.StewardCadenceTicks
	}
	return defaultStewardCadenceTicks
}

// AmbientCoalescing reports whether the spec-104 coalescing regime is ON for
// this world: movement rides agent.path_started segments, needs thin to
// checkpoints + crossings, gru motion is derived. OFF (legacy) for nil Tuning
// and for every pre-104 recorded tuning payload (field absent ⇒ 0), so old
// worlds keep the old emission shape and the advancement engine stays inert
// on their folds (research.md §3 — the double-fold guard).
func (s *State) AmbientCoalescing() bool {
	return s.Tuning != nil && s.Tuning.NeedsCheckpointMinutes > 0
}

// NeedsCheckpointK is the needs-checkpoint cadence K in game-minutes (spec
// 104 FR-008): agent.needs_changed emits on the K-minute grid plus on band
// crossings. 1 (today's per-minute cadence) while the regime is off — legacy
// worlds emit every minute through the retained heartbeat path, so this
// accessor is only ever consulted under AmbientCoalescing().
func (s *State) NeedsCheckpointK() int64 {
	if s.Tuning != nil && s.Tuning.NeedsCheckpointMinutes > 0 {
		return s.Tuning.NeedsCheckpointMinutes
	}
	return 1
}

// DreamDials is the resolved spec-098 dream dial block (nil-safe): the one
// consumption path for PlanDream's parameters, read off the mind's replica at
// consolidation-snapshot time like the other mind-side dials.
func (s *State) DreamDials() DreamTuning {
	if s.Tuning != nil {
		return s.Tuning.EffectiveDream()
	}
	return defaultDream()
}

// EffectiveTuning returns the full in-effect dial set (tuned values or their
// defaults); nil Tuning resolves to the default set. The daemon boot seed
// compares a parsed manifest against this — via Equal, since spec 098's Dream
// pointer block took TuningState out of bare-==-comparability — to decide
// whether the effective set differs from what state already carries and thus
// whether to append a sim.tuning_applied event (spec 048 FR-004).
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
	// Spec 097 (perception of absence, FR-004).
	ObservationDedupTicks         *int64 `json:"observation_dedup_ticks"`
	ObservationBaseSalience       *int64 `json:"observation_base_salience"`
	BeliefDisconfirmRetainPercent *int64 `json:"belief_disconfirm_retain_percent"`
	BeliefConfirmBoost            *int64 `json:"belief_confirm_boost"`
	// Steward cadence (spec 102): the guardian-agentization opt-in dial.
	StewardCadenceTicks *int64 `json:"steward_cadence_ticks"`
	// Spec 104: the needs-checkpoint cadence / coalescing-regime dial. Like
	// every dial, an absent key resolves to the doctrine default (10) — so
	// ANY tuning.json turns the regime on at next boot (a deterministic,
	// event-recorded, forward-only change; the flip transition stamps the
	// advancement watermarks, state.go).
	NeedsCheckpointMinutes *int64 `json:"needs_checkpoint_minutes"`
	// Dream dials (spec 098): flat keys like every other dial — the manifest
	// stays one sparse level deep for the operator; the nested resolved block
	// is a state/payload shape, not an authoring shape.
	DreamDensityPerMille       *int64 `json:"dream_density_per_mille"`
	DreamAmbiguousBandPerMille *int64 `json:"dream_ambiguous_band_per_mille"`
	DreamHabituationPerMille   *int64 `json:"dream_habituation_per_mille"`
	DreamMergeCapPerNight      *int64 `json:"dream_merge_cap_per_night"`
	DreamJitterPerMille        *int64 `json:"dream_jitter_per_mille"`
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
	clampI64("observation_dedup_ticks", m.ObservationDedupTicks, &t.ObservationDedupTicks, minObservationDedupTicks, maxObservationDedupTicks)
	clampI64("observation_base_salience", m.ObservationBaseSalience, &t.ObservationBaseSalience, minObservationBaseSalience, maxObservationBaseSalience)
	clampI64("belief_disconfirm_retain_percent", m.BeliefDisconfirmRetainPercent, &t.BeliefDisconfirmRetainPercent, minBeliefDisconfirmRetainPercent, maxBeliefDisconfirmRetainPercent)
	clampI64("belief_confirm_boost", m.BeliefConfirmBoost, &t.BeliefConfirmBoost, minBeliefConfirmBoost, maxBeliefConfirmBoost)
	clampI64("needs_checkpoint_minutes", m.NeedsCheckpointMinutes, &t.NeedsCheckpointMinutes, minNeedsCheckpointMinutes, maxNeedsCheckpointMinutes)

	// Steward cadence (spec 102): 0 is the off switch and passes untouched; a
	// nonzero value clamps to the [min,max] band with the standard warning.
	if m.StewardCadenceTicks != nil {
		v := *m.StewardCadenceTicks
		switch {
		case v == 0:
			// explicit off — the default; nothing to clamp
		case v < 0:
			// A negative is nonsense: fail toward OFF, never toward opting a
			// world into agentization it did not ask for.
			warns = append(warns, fmt.Sprintf("tuning.json steward_cadence_ticks %d is negative — clamped to 0 (off)", v))
			v = 0
		case v < minStewardCadenceTicks:
			warns = append(warns, fmt.Sprintf("tuning.json steward_cadence_ticks %d out of range (min %d; 0 = off) — clamped to %d", v, int64(minStewardCadenceTicks), int64(minStewardCadenceTicks)))
			v = minStewardCadenceTicks
		case v > maxStewardCadenceTicks:
			warns = append(warns, fmt.Sprintf("tuning.json steward_cadence_ticks %d out of range (max %d) — clamped to %d", v, int64(maxStewardCadenceTicks), int64(maxStewardCadenceTicks)))
			v = maxStewardCadenceTicks
		}
		t.StewardCadenceTicks = v
	}

	// Dream dials (spec 098): any present key resolves the FULL dream block
	// against its defaults (a non-nil block is never sparse); no key present
	// leaves Dream nil ≡ the default set, so a pre-098 manifest parses to a
	// set Equal to what a pre-098 world already carries.
	if m.DreamDensityPerMille != nil || m.DreamAmbiguousBandPerMille != nil ||
		m.DreamHabituationPerMille != nil || m.DreamMergeCapPerNight != nil ||
		m.DreamJitterPerMille != nil {
		d := defaultDream()
		clampI64("dream_density_per_mille", m.DreamDensityPerMille, &d.DensityPerMille, minDreamDensityPerMille, maxDreamDensityPerMille)
		clampI64("dream_ambiguous_band_per_mille", m.DreamAmbiguousBandPerMille, &d.AmbiguousBandPerMille, minDreamAmbiguousBandPerMille, maxDreamAmbiguousBandPerMille)
		clampI64("dream_habituation_per_mille", m.DreamHabituationPerMille, &d.HabituationPerMille, minDreamHabituationPerMille, maxDreamHabituationPerMille)
		clampI64("dream_merge_cap_per_night", m.DreamMergeCapPerNight, &d.MergeCapPerNight, minDreamMergeCapPerNight, maxDreamMergeCapPerNight)
		clampI64("dream_jitter_per_mille", m.DreamJitterPerMille, &d.JitterPerMille, minDreamJitterPerMille, maxDreamJitterPerMille)
		t.Dream = &d
	}

	return &t, warns, nil
}

// --- event constructor (spec 048 US2) ---

// TuningAppliedPayload is the sim.tuning_applied event body: the full effective
// dial set (never a delta), so replay can establish tuning from any single
// event without scanning history. The spec 097 dials are POINTERS: a pre-097
// recorded payload (fields absent) decodes to nil, and the Apply arm
// (state.go) resolves nil to the doctrine default — never to zero — so old
// worlds replay under the same effective dials new defaults give them (the
// spec 092 additive-field discipline). New events always carry all fields
// (NewTuningEvent below).
type TuningAppliedPayload struct {
	RefuelDyingBelow       int64  `json:"refuel_dying_below"`
	FireBurnPerWood        int64  `json:"fire_burn_per_wood"`
	GruEmergePerMille      uint64 `json:"gru_emerge_per_mille"`
	PlannerCadenceTicks    int64  `json:"planner_cadence_ticks"`
	EncounterCooldownTicks int64  `json:"encounter_cooldown_ticks"`
	// Spec 097 (perception of absence, FR-004).
	ObservationDedupTicks         *int64 `json:"observation_dedup_ticks,omitempty"`
	ObservationBaseSalience       *int64 `json:"observation_base_salience,omitempty"`
	BeliefDisconfirmRetainPercent *int64 `json:"belief_disconfirm_retain_percent,omitempty"`
	BeliefConfirmBoost            *int64 `json:"belief_confirm_boost,omitempty"`
	// Steward cadence (spec 102): pointer + omitempty for the READ side — a
	// pre-102 recorded payload decodes nil, resolved to 0 (off) at Apply, so
	// old logs replay byte-identically with no format bump. New events always
	// carry the field (NewTuningEvent).
	StewardCadenceTicks *int64 `json:"steward_cadence_ticks,omitempty"`
	// Spec 104: the needs-checkpoint / coalescing-regime dial. Pointer +
	// omitempty for the READ side: a pre-104 recorded payload decodes nil,
	// which resolveTuning keeps as 0 = LEGACY — deliberately NOT the doctrine
	// default, unlike the spec-097 dials, because the field is the regime
	// marker and a pre-104 world must fold to the legacy emission shape
	// (research.md §3).
	NeedsCheckpointMinutes *int64 `json:"needs_checkpoint_minutes,omitempty"`
	// Dream (spec 098): newly-emitted events always carry the resolved block
	// (the full-set doctrine); the pointer + omitempty exist for the READ
	// side — a pre-098 recorded event decodes nil, which the apply arm keeps
	// nil ≡ defaults, so old logs replay byte-identically with no format bump.
	Dream *DreamTuning `json:"dream,omitempty"`
}

// resolveTuning turns a decoded tuning_applied payload into the full effective
// TuningState, resolving absent (nil) spec-097 fields to their doctrine
// defaults. Shared by the Apply arm and tests.
func resolveTuning(p TuningAppliedPayload) TuningState {
	t := TuningState{
		RefuelDyingBelow:              p.RefuelDyingBelow,
		FireBurnPerWood:               p.FireBurnPerWood,
		GruEmergePerMille:             p.GruEmergePerMille,
		PlannerCadenceTicks:           p.PlannerCadenceTicks,
		EncounterCooldownTicks:        p.EncounterCooldownTicks,
		ObservationDedupTicks:         defaultObservationDedupTicks,
		ObservationBaseSalience:       defaultObservationBaseSalience,
		BeliefDisconfirmRetainPercent: defaultBeliefDisconfirmRetainPercent,
		BeliefConfirmBoost:            defaultBeliefConfirmBoost,
	}
	if p.ObservationDedupTicks != nil {
		t.ObservationDedupTicks = *p.ObservationDedupTicks
	}
	if p.ObservationBaseSalience != nil {
		t.ObservationBaseSalience = *p.ObservationBaseSalience
	}
	if p.BeliefDisconfirmRetainPercent != nil {
		t.BeliefDisconfirmRetainPercent = *p.BeliefDisconfirmRetainPercent
	}
	if p.BeliefConfirmBoost != nil {
		t.BeliefConfirmBoost = *p.BeliefConfirmBoost
	}
	// Steward cadence (spec 102): absent resolves to 0 — off, the pre-102 world.
	if p.StewardCadenceTicks != nil {
		t.StewardCadenceTicks = *p.StewardCadenceTicks
	}
	// Spec 104: absent resolves to 0 = legacy (the regime marker), never to
	// the doctrine default — see the payload field's comment.
	if p.NeedsCheckpointMinutes != nil {
		t.NeedsCheckpointMinutes = *p.NeedsCheckpointMinutes
	} else {
		t.NeedsCheckpointMinutes = 0
	}
	// Dream (spec 098): a decoded nil stays nil ≡ defaults; a carried block
	// lands as a FRESH copy, never the payload's pointer (the apply-arm
	// doctrine).
	if p.Dream != nil {
		d := *p.Dream
		t.Dream = &d
	}
	return t
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
	// Full-set doctrine: the emitted payload always carries the RESOLVED
	// dream block (spec 098), even when t.Dream is nil — a newly-pinned world
	// fixes its dream doctrine at birth exactly as spec 057 fixed the base
	// five, and later default* changes never reach back into its live runs.
	d := t.EffectiveDream()
	return store.Event{Tick: tick, Type: "sim.tuning_applied",
		Payload: mustPayload(TuningAppliedPayload{
			RefuelDyingBelow:              t.RefuelDyingBelow,
			FireBurnPerWood:               t.FireBurnPerWood,
			GruEmergePerMille:             t.GruEmergePerMille,
			PlannerCadenceTicks:           t.PlannerCadenceTicks,
			EncounterCooldownTicks:        t.EncounterCooldownTicks,
			ObservationDedupTicks:         &t.ObservationDedupTicks,
			ObservationBaseSalience:       &t.ObservationBaseSalience,
			BeliefDisconfirmRetainPercent: &t.BeliefDisconfirmRetainPercent,
			BeliefConfirmBoost:            &t.BeliefConfirmBoost,
			StewardCadenceTicks:           &t.StewardCadenceTicks,
			NeedsCheckpointMinutes:        &t.NeedsCheckpointMinutes,
			Dream:                         &d,
		})}
}
