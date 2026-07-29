package sim

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/evanstern/promptworld/internal/store"
)

// Private dreams (spec 098, TASK-99): the nightly consolidation slot's
// clustering + habituation layer — the model-free half. Once per night, the
// mind's consolidation driver runs PlanDream over ONE agent's recorded memory
// vectors (spec 042) and lands the outcomes as recorded events; the reducer
// arms below apply what was recorded and never re-derive a cluster decision
// (spec 092 doctrine — emitter computes, reducer applies).
//
// Privacy is by construction (D1): PlanDream's only memory input is the one
// agent's own store — there is no shared vector table and no cross-agent
// index anywhere in this file, so two agents with identical experiences
// consolidate to independent outcomes.
//
// Routing is geometry-first (D2, the SAGE/RecMem economics): clear-cut
// cluster decisions resolve by embedding density alone; only the ambiguous
// band is handed back to the caller for the consolidation LLM slot the
// nightly phase already owns. No new LLM call classes exist — a night with
// no consolidation call consults nobody and the band retries next night.

// dreamMinClusterSize is the smallest member count that reads as a routine
// cluster: a pair is coincidence, three-and-up is a habit. Doctrine constant,
// deliberately not a tuning.json dial (the five spec-098 dials are the
// promoted set; this stays versioned with the code).
const dreamMinClusterSize = 3

// DreamReasonHabituation is the recorded reason a salience revision carries
// when the dream pass down-weights a routine cluster member.
const DreamReasonHabituation = "habituation"

// --- event payloads (landed via the whitelisted injection door) ---

// SalienceRevisedPayload sets one memory's salience to an absolute recorded
// value (never a delta): the emitter computed the habituation outcome, the
// reducer applies it verbatim. Identified by the same durable (tick, hash)
// pair the consolidation family uses.
type SalienceRevisedPayload struct {
	Agent    AgentRef `json:"agent"`
	MemTick  int64    `json:"mem_tick"`
	TextHash string   `json:"text_hash"`
	Salience int      `json:"salience"`
	Reason   string   `json:"reason,omitempty"` // audit surface ("habituation")
}

// MemoryMergedPayload folds a routine cluster's absorbed members into their
// kept representative: the merged refs are removed, the kept memory's
// salience is set to the recorded value. Emitter-computed like everything
// else in the family; vanished targets no-op.
type MemoryMergedPayload struct {
	Agent    AgentRef    `json:"agent"`
	Kept     MemoryRef   `json:"kept"`
	Merged   []MemoryRef `json:"merged"`
	Salience int         `json:"salience"` // kept memory's post-merge salience
}

// --- the pure clustering pass (D1/D2/D4) ---

// DreamRevision is one planned habituation down-weight: the memory at Ref
// drops to the absolute Salience recorded here.
type DreamRevision struct {
	Ref      MemoryRef
	Salience int
}

// DreamMerge is one planned cluster fold: Merged members are absorbed into
// Kept, whose salience becomes Salience. Merged may be empty when the
// nightly merge budget was already spent — the cluster then habituates only.
type DreamMerge struct {
	Kept     MemoryRef
	Merged   []MemoryRef
	Salience int
}

// DreamGroup is one ambiguous-band cluster: cohesion near the density
// threshold, so geometry alone must not decide (D2). The caller may consult
// the existing consolidation LLM slot with Size/Examples; a "fold" verdict
// lands the precomputed Revisions/Merge as recorded events, a "keep" verdict
// lands nothing (the decision itself rides the consolidation marker).
type DreamGroup struct {
	Size      int
	Examples  []string // up to dreamGroupExamples member texts, oldest first
	Revisions []DreamRevision
	Merge     DreamMerge
}

// dreamGroupExamples caps the member texts a DreamGroup carries for the
// consult prompt.
const dreamGroupExamples = 3

// DreamPlan is PlanDream's outcome: the clear-cut habituations and merges
// geometry decided alone, plus the ambiguous groups reserved for the
// consolidation slot. All fields may be empty — a store with no dense
// neighborhoods dreams of nothing.
type DreamPlan struct {
	Revisions []DreamRevision
	Merges    []DreamMerge
	Ambiguous []DreamGroup
}

// PlanDream is the per-agent density clustering pass (spec 098 FR-001/002):
// a pure function of ONE agent's memory snapshot, the world seed, the night
// index, the agent index, and the resolved dream dials — deterministic, no
// model, no other agent's state reachable.
//
// Mechanics (contract recorded in specs/098-private-dreams/spec.md D2/D4):
//
//   - Only memories carrying a recorded vector participate; clustering never
//     crosses embedding models (the FR-009 cross-model guard, as in
//     relevance01) and zero-magnitude vectors are incomparable.
//   - Leader clustering in append order: a memory joins the first cluster
//     whose leader's cosine similarity reaches the membership bar
//     (density − band, per-mille); otherwise it founds a new cluster.
//   - Clusters of dreamMinClusterSize+ members are candidates. Each
//     candidate's cohesion (mean member↔leader cosine) is compared against
//     the clear bar (density + band) — after the D4 boundary jitter, a
//     zeroable rngAt-seeded nudge purpose-keyed (seed, "dream", night,
//     agent), drawn once per candidate in cluster order. At or above the
//     clear bar geometry decides alone; below it the cluster is ambiguous
//     and reserved for the consolidation slot. Jitter moves only this
//     boundary — membership uses the raw bar, so habituation of true
//     duplicates stays stable (D4's bound).
//   - A decided (or precomputed ambiguous) cluster keeps its most salient
//     member (ties: newer, then later) vivid as the representative; the
//     oldest non-representatives are merged into it while the shared
//     per-night merge budget lasts (clear and ambiguous clusters draw from
//     the one budget in cluster order), and every remaining member is
//     habituated to salience*factor (per-mille, floor 1) — a revision is
//     planned only when the value actually drops.
func PlanDream(memories []Memory, seed uint64, night int64, agentIdx int, d DreamTuning) DreamPlan {
	memberBar := float64(d.DensityPerMille-d.AmbiguousBandPerMille) / 1000
	clearBar := float64(d.DensityPerMille+d.AmbiguousBandPerMille) / 1000

	// Leader clustering over the vectored subset, append order.
	type cluster struct {
		leader  int
		members []int // includes leader
	}
	var clusters []*cluster
	for i := range memories {
		m := &memories[i]
		if len(m.Vec) == 0 {
			continue
		}
		joined := false
		for _, c := range clusters {
			l := &memories[c.leader]
			if m.VecModel != l.VecModel {
				continue
			}
			if cos, ok := cos32(m.Vec, l.Vec); ok && cos >= memberBar {
				c.members = append(c.members, i)
				joined = true
				break
			}
		}
		if !joined {
			clusters = append(clusters, &cluster{leader: i, members: []int{i}})
		}
	}

	// One jitter stream per (seed, night, agent) — deterministic, zeroable.
	var plan DreamPlan
	r := rngAt(seed, "dream", night, agentIdx)
	budget := d.MergeCapPerNight
	for _, c := range clusters {
		if len(c.members) < dreamMinClusterSize {
			continue
		}
		// Cohesion: mean member↔leader cosine over the non-leader members.
		var sum float64
		n := 0
		for _, mi := range c.members {
			if mi == c.leader {
				continue
			}
			if cos, ok := cos32(memories[mi].Vec, memories[c.leader].Vec); ok {
				sum += cos
				n++
			}
		}
		if n == 0 {
			continue
		}
		score := sum / float64(n)
		if d.JitterPerMille > 0 {
			// D4: the boundary nudge — drawn even for clusters that end up
			// clear, so classification order never shifts the stream.
			score += (r.Float64()*2 - 1) * float64(d.JitterPerMille) / 1000
		}

		// Representative: max salience, ties newer tick, ties later index.
		rep := c.members[0]
		for _, mi := range c.members[1:] {
			m, best := &memories[mi], &memories[rep]
			if m.Salience > best.Salience ||
				(m.Salience == best.Salience && (m.Tick > best.Tick || (m.Tick == best.Tick && mi > rep))) {
				rep = mi
			}
		}

		// Merge the oldest non-representatives while the shared budget lasts.
		merge := DreamMerge{
			Kept:     MemoryRef{Tick: memories[rep].Tick, Hash: MemoryHash(memories[rep].Text)},
			Salience: memories[rep].Salience,
		}
		merged := map[int]bool{}
		for _, mi := range c.members { // append order == oldest first
			if budget <= 0 {
				break
			}
			if mi == rep {
				continue
			}
			merge.Merged = append(merge.Merged, MemoryRef{Tick: memories[mi].Tick, Hash: MemoryHash(memories[mi].Text)})
			merged[mi] = true
			budget--
		}

		// Habituate every surviving non-representative that actually drops.
		var revs []DreamRevision
		for _, mi := range c.members {
			if mi == rep || merged[mi] {
				continue
			}
			m := &memories[mi]
			newSal := int(int64(m.Salience) * d.HabituationPerMille / 1000)
			if newSal < 1 {
				newSal = 1
			}
			if newSal < m.Salience {
				revs = append(revs, DreamRevision{
					Ref:      MemoryRef{Tick: m.Tick, Hash: MemoryHash(m.Text)},
					Salience: newSal,
				})
			}
		}

		if score >= clearBar {
			plan.Revisions = append(plan.Revisions, revs...)
			if len(merge.Merged) > 0 {
				plan.Merges = append(plan.Merges, merge)
			}
		} else {
			g := DreamGroup{Size: len(c.members), Revisions: revs, Merge: merge}
			for _, mi := range c.members {
				if len(g.Examples) >= dreamGroupExamples {
					break
				}
				g.Examples = append(g.Examples, memories[mi].Text)
			}
			plan.Ambiguous = append(plan.Ambiguous, g)
		}
	}
	return plan
}

// DreamEvents renders a plan's clear-cut outcomes (and, for a folded
// ambiguous group, its precomputed outcomes) as the recorded event batch the
// injection door accepts. Pure serialization — no decisions here.
func DreamEvents(agent int, revs []DreamRevision, merges []DreamMerge) []store.Event {
	var out []store.Event
	for _, mg := range merges {
		out = append(out, store.Event{Type: "agent.memory_merged", Payload: mustPayload(MemoryMergedPayload{
			Agent: Ref(agent), Kept: mg.Kept, Merged: mg.Merged, Salience: mg.Salience})})
	}
	for _, rv := range revs {
		out = append(out, store.Event{Type: "agent.salience_revised", Payload: mustPayload(SalienceRevisedPayload{
			Agent: Ref(agent), MemTick: rv.Ref.Tick, TextHash: rv.Ref.Hash,
			Salience: rv.Salience, Reason: DreamReasonHabituation})})
	}
	return out
}

// cos32 is the raw cosine similarity between two recorded float32 vectors,
// accumulated sequentially in float64 in fixed index order (the relevance01
// determinism discipline). ok is false when the vectors are incomparable —
// dimension mismatch or zero magnitude.
func cos32(a, b []float32) (float64, bool) {
	if len(a) == 0 || len(a) != len(b) {
		return 0, false
	}
	var dot, aa, bb float64
	for i := range a {
		av, bv := float64(a[i]), float64(b[i])
		dot += av * bv
		aa += av * av
		bb += bv * bv
	}
	if aa == 0 || bb == 0 {
		return 0, false
	}
	return dot / (math.Sqrt(aa) * math.Sqrt(bb)), true
}

// applyDream is the reducer arm for the two dream event types, dispatched
// from State.Apply. Total like the consolidation arms: vanished targets
// degrade to no-ops, never errors — replay applies recorded outcomes and
// re-derives nothing (spec 092).
func (s *State) applyDream(e store.Event) error {
	agent := func(i int) (*Agent, error) {
		if i < 0 || i >= len(s.Agents) {
			return nil, fmt.Errorf("apply %s: agent %d out of range", e.Type, i)
		}
		return &s.Agents[i], nil
	}
	clampSal := func(v int) int {
		if v < 1 {
			return 1
		}
		if v > MaxSalience {
			return MaxSalience
		}
		return v
	}
	setSalience := func(a *Agent, ref MemoryRef, sal int) {
		for i := range a.Memories {
			m := &a.Memories[i]
			if m.Tick == ref.Tick && MemoryHash(m.Text) == ref.Hash {
				m.Salience = clampSal(sal)
				break
			}
		} // vanished target: no-op
	}
	switch e.Type {
	case "agent.salience_revised":
		var p SalienceRevisedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		a, err := agent(p.Agent.ID)
		if err != nil {
			return err
		}
		setSalience(a, MemoryRef{Tick: p.MemTick, Hash: p.TextHash}, p.Salience)

	case "agent.memory_merged":
		var p MemoryMergedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		a, err := agent(p.Agent.ID)
		if err != nil {
			return err
		}
		for _, ref := range p.Merged {
			for i := range a.Memories {
				if a.Memories[i].Tick == ref.Tick && MemoryHash(a.Memories[i].Text) == ref.Hash {
					a.Memories = append(a.Memories[:i], a.Memories[i+1:]...)
					break
				}
			} // vanished member: no-op
		}
		setSalience(a, p.Kept, p.Salience)
	}
	return nil
}
