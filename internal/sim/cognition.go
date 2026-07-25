package sim

import "encoding/json"

// Cognition-horizon telemetry (TASK-32, specs/007-cognition-horizon).
// These event types are recorded observability with zero state effect:
// reducer no-ops whitelisted on the inject_social door (cog.*) or emitted by
// the loop itself alongside the verdict they describe
// (agent.intent_rejected). Payload field order is canonical — histories are
// byte-comparable (contracts/events.md).

// Thought outcomes: every requested thought terminates in exactly one
// (FR-015). Silent failure is eliminated.
const (
	OutcomeLanded        = "landed"
	OutcomeAdapted       = "adapted"
	OutcomeRejectedStale = "rejected-stale"
	OutcomeRejectedGuard = "rejected-guard"
	OutcomeSuperseded    = "superseded"
	OutcomeExpired       = "expired"
	OutcomeUnavailable   = "rejected-unavailable"
	OutcomeUnusable      = "unusable"
	OutcomeSuppressed    = "suppressed"
	// OutcomeClamped (spec 058 US2/FR-003): a set_plan landing whose step
	// count exceeded PlanStepCap accepted anyway — the first PlanStepCap
	// steps landed, truncated at the guard (internal/sim/landing.go) rather
	// than the whole plan being rejected. Distinguishes a clamped acceptance
	// from a clean OutcomeLanded (the toolloop.VerdictLandedClamped analog,
	// one layer down — the reducer's own outcome vocabulary, not the driver's
	// verdict enum).
	OutcomeClamped = "clamped"
	// OutcomeRetried is a NON-TERMINAL marker (TASK-42, conversation
	// robustness): one scene reply failed to parse and the scene continued
	// via one retry. It carries the failed reply's raw text; consumers that
	// sum job completions MUST filter it out (contracts/telemetry.md rule 1).
	OutcomeRetried = "retried"
)

// Rejection classification (FR-013): prediction-miss is an infrastructure
// signal (kept out of tuning heuristics as a spike); world-change means the
// world moved on — supersede/guards working as intended.
const (
	RejectKindPredictionMiss = "prediction-miss"
	RejectKindWorldChange    = "world-change"
)

// GenerationBumpSalience: an agent.memory_added at or above this salience
// bumps Agent.Generation (FR-014). The salience table defines "emergency":
// near-death 9, witnessed death 10, exile 9 — dreams (8) do not interrupt.
const GenerationBumpSalience = 9

// PredictionMissFactor: a landing whose actual wall time exceeded its
// prediction by this factor is classified prediction-miss, not world-change
// — infra noise that must stay out of budget-tuning heuristics (FR-013).
const PredictionMissFactor = 3

// CogThoughtPayload — cog.thought: a model call passed the router and was
// enqueued. trigger_seq is the event-log seq of the stimulus that armed the
// trigger (0 = pure cadence): the causality edge stimulus → thought.
type CogThoughtPayload struct {
	Job               string `json:"job"`
	Class             string `json:"class"`
	Agent             int    `json:"agent"`
	SnapshotTick      int64  `json:"snapshot_tick"`
	Generation        int64  `json:"generation"`
	TriggerSeq        int64  `json:"trigger_seq"`
	Points            int    `json:"points"`
	PredictedWallMs   int64  `json:"predicted_wall_ms"`
	PredictedLandTick int64  `json:"predicted_land_tick"`
	// Context-grounding observability (spec 043, FR-009/FR-010): the assembled
	// user-prompt size, the per-block byte breakdown (keyed by contract block
	// name), and the blocks the size budget dropped, in drop order. Additive,
	// LAST, omitempty — pre-043 cog.thought events and every non-planner
	// emission (conversation thoughts carry none) marshal byte-identically, and
	// old logs decode with these zero-valued. Reducer stays a no-op (cog.*),
	// so replay is unaffected.
	PromptBytes   int            `json:"prompt_bytes,omitempty"`
	BlockBytes    map[string]int `json:"block_bytes,omitempty"`
	DroppedBlocks []string       `json:"dropped_blocks,omitempty"`
}

// CogOutcomePayload — cog.outcome: the single terminal record of a thought.
// Router suppressions carry the routing arithmetic in reason and have no
// matching cog.thought (no call was made).
type CogOutcomePayload struct {
	Job             string `json:"job"`
	Class           string `json:"class"`
	Agent           int    `json:"agent"`
	Outcome         string `json:"outcome"`
	SnapshotTick    int64  `json:"snapshot_tick"`
	LandingTick     int64  `json:"landing_tick"`
	StalenessTicks  int64  `json:"staleness_ticks"`
	PredictedWallMs int64  `json:"predicted_wall_ms"`
	ActualWallMs    int64  `json:"actual_wall_ms"`
	Kind            string `json:"kind,omitempty"`
	Reason          string `json:"reason,omitempty"`
	// Raw / Retried (TASK-42): raw is the verbatim model reply on a scene
	// parse failure (bounded, truncated on a rune boundary); retried marks a
	// terminal scene outcome whose run consumed ≥1 retry. Both omitempty, so
	// every pre-TASK-42 emission stays byte-identical (FR-009).
	Raw     string `json:"raw,omitempty"`
	Retried bool   `json:"retried,omitempty"`
}

// IntentRejectedPayload — agent.intent_rejected: the loop refused a landing
// intent. Its own type (not just telemetry) so souls/chronicle can later
// notice refused intentions without parsing cog.* payloads.
type IntentRejectedPayload struct {
	Agent          int    `json:"agent"`
	Goal           string `json:"goal"`
	Reason         string `json:"reason"`
	StalenessTicks int64  `json:"staleness_ticks"`
}

// CogToolCallPayload — cog.tool_call: one record per tool call the loop saw
// (spec 017, FR-007) — landed, rejected, read, or unlanded. Reducer no-op
// like every cog.* type (recorded observability, zero state effect);
// {Job, Ordinal} is the correlation key (ordinals are 1-based, dense per
// job, in model-emission order across every round). Field order is
// canonical (contracts/events.md) — future additive fields go LAST,
// omitempty, so existing cog.tool_call events keep replaying
// byte-identically (TASK-32 pattern).
type CogToolCallPayload struct {
	Job     string `json:"job"`
	Ordinal int    `json:"ordinal"`
	Tool    string `json:"tool"`
	// Args is the call's raw arguments, copied verbatim up to the 2 KiB cap
	// (toolloop.capArgs; larger payloads truncate to
	// {"_truncated":true,"prefix":"…"}); omitempty for zero-argument calls.
	Args json.RawMessage `json:"args,omitempty"`
	// Verdict is the toolloop.Verdict enum (data-model.md §5): landed |
	// rejected_gate | rejected_cardinality | rejected_unknown |
	// rejected_malformed | read_ok | read_error | unlanded.
	Verdict string `json:"verdict"`
	// Reason is omitempty but REQUIRED (non-empty) for every rejected_* and
	// read_error verdict — the queryable rejection explanation (AC#5).
	// Enforced by emitters (mind/metatron), not this type.
	Reason string `json:"reason,omitempty"`
	Tier   string `json:"tier"`
	// SnapshotTick is the world tick the call's cognition snapshotted at.
	SnapshotTick int64 `json:"snapshot_tick"`
}

// NewCogToolCallPayload assembles a cog.tool_call payload from a recorded
// call's plain fields (spec 017 T018). It lives sim-side — next to the payload
// it builds — with only plain/std-lib argument types (no toolloop or mind
// import), so BOTH loop consumers reach it without a shared helper package or
// dependency inversion: the mind (T018) and metatron (T020) each unpack their
// own toolloop.CallRecord and call this, sharing one authority for the payload's
// field set. verdict is the toolloop.Verdict enum stringified by the caller;
// reason must be non-empty for every rejected_* / read_error verdict — the
// caller enforces that (contracts/events.md), this constructor only shapes.
func NewCogToolCallPayload(job string, ordinal int, tool string, args json.RawMessage, verdict, reason, tier string, snapshotTick int64) CogToolCallPayload {
	return CogToolCallPayload{
		Job:          job,
		Ordinal:      ordinal,
		Tool:         tool,
		Args:         args,
		Verdict:      verdict,
		Reason:       reason,
		Tier:         tier,
		SnapshotTick: snapshotTick,
	}
}

// RecalibrationPayload — cog.recalibration_recommended: the live estimator's
// spike rate breached the drift threshold (once per breach episode). Post-spec
// 031 the same episode also ADOPTS: PriorSPerPt/AdoptedSPerPt are additive,
// omitempty fields carrying the adoption arithmetic (prior estimate → window
// median installed). EstimateSPerPt keeps its meaning — "the estimator's
// current estimate at emission" — which post-adoption equals AdoptedSPerPt.
// Pre-031 events lack the two new fields and replay identically (the reducer is
// a no-op either way); see specs/031-.../contracts/adoption-event.md.
type RecalibrationPayload struct {
	Tier           string  `json:"tier"`
	EstimateSPerPt float64 `json:"estimate_s_per_pt"`
	SpikeRate      float64 `json:"spike_rate"`
	Window         int     `json:"window"`
	PriorSPerPt    float64 `json:"prior_s_per_pt,omitempty"`
	AdoptedSPerPt  float64 `json:"adopted_s_per_pt,omitempty"`
}

// MemoryDivergencePayload — cog.memory_divergence (spec 042 US2): one
// selection's rank divergence between the legacy window and the
// relevance-augmented window, recorded at plan/scene-snapshot time by the mind
// while memory_relevance is "shadow" or "on". Reducer NO-OP telemetry
// (recorded observability, cog.* class) — the recorded evidence the US2→US3
// gate decision is made from (FR-006/FR-007). Both windows ride as memory
// Seqs in window order, auditable against their agent.memory_added events.
type MemoryDivergencePayload struct {
	Agent        int     `json:"agent"`
	Tick         int64   `json:"tick"`
	Mode         string  `json:"mode"`
	Legacy       []int64 `json:"legacy"`
	Augmented    []int64 `json:"augmented"`
	Overlap      int     `json:"overlap"`
	Displacement int     `json:"displacement"`
	Vectorless   int     `json:"vectorless"`
	SitTick      int64   `json:"sit_tick"`
}

// NewMemoryDivergencePayload assembles the divergence record from the two
// selected windows — the sim-side payload authority (the NewCogToolCallPayload
// precedent), pure over recorded data so the arithmetic is unit-testable
// beside the selector it audits. Overlap counts memories present in both
// windows; displacement sums the absolute rank distance of each shared
// member. Identity is the memory's stamped Seq; a pre-042 (seq-less, Seq 0)
// memory has no durable identity and never counts as shared — its window slot
// still rides the lists as 0, visibly pre-042. sitTick 0 means no situation
// vector existed, so the rankings are identical by definition.
func NewMemoryDivergencePayload(agent int, tick int64, mode string, legacy, augmented []Memory, vectorless int, sitTick int64) MemoryDivergencePayload {
	l := make([]int64, len(legacy))
	rank := make(map[int64]int, len(legacy))
	for i, m := range legacy {
		l[i] = m.Seq
		if m.Seq != 0 {
			rank[m.Seq] = i
		}
	}
	g := make([]int64, len(augmented))
	overlap, displacement := 0, 0
	for i, m := range augmented {
		g[i] = m.Seq
		if m.Seq == 0 {
			continue
		}
		if li, ok := rank[m.Seq]; ok {
			overlap++
			d := i - li
			if d < 0 {
				d = -d
			}
			displacement += d
		}
	}
	return MemoryDivergencePayload{
		Agent: agent, Tick: tick, Mode: mode,
		Legacy: l, Augmented: g,
		Overlap: overlap, Displacement: displacement,
		Vectorless: vectorless, SitTick: sitTick,
	}
}
