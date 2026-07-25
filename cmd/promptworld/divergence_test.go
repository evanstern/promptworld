package main

import (
	"math"
	"testing"

	"github.com/evanstern/promptworld/internal/sim"
)

// TestAggregateDivergence (spec 042 T018): the pure aggregation groups by
// (agent, game day), normalizes overlap by each selection's own window size,
// counts a selection as "promoted" when relevance pulled in ≥1 memory the
// legacy ranking excluded (non-zero seqs only — pre-042 zeros never match),
// and the total row folds every record.
func TestAggregateDivergence(t *testing.T) {
	day := int64(divTicksPerDay)
	recs := []sim.MemoryDivergencePayload{
		// Agent 0, day 0: identical k=2 windows — full overlap, nothing promoted.
		{Agent: 0, Tick: 100, Mode: "shadow", Legacy: []int64{5, 6}, Augmented: []int64{5, 6}, Overlap: 2, Displacement: 0, Vectorless: 1},
		// Agent 0, day 0: one memory promoted (seq 9 absent from legacy), half overlap.
		{Agent: 0, Tick: 200, Mode: "shadow", Legacy: []int64{5, 6}, Augmented: []int64{5, 9}, Overlap: 1, Displacement: 0, Vectorless: 0},
		// Agent 0, day 1: k=5 selection, 4/5 overlap, displaced by 3, and seq 7
		// promoted (absent from legacy).
		{Agent: 0, Tick: day + 1, Mode: "shadow", Legacy: []int64{1, 2, 3, 4, 5}, Augmented: []int64{2, 1, 3, 5, 7}, Overlap: 4, Displacement: 3, Vectorless: 2},
		// Agent 1, day 0: zero seqs (pre-042 memories) never count as promoted.
		{Agent: 1, Tick: 300, Mode: "shadow", Legacy: []int64{0, 0}, Augmented: []int64{0, 0}, Overlap: 0, Displacement: 0, Vectorless: 2},
	}
	keys, rows, total := aggregateDivergence(recs)

	wantKeys := []divKey{{0, 0}, {0, 1}, {1, 0}}
	if len(keys) != len(wantKeys) {
		t.Fatalf("keys = %v, want %v", keys, wantKeys)
	}
	for i := range wantKeys {
		if keys[i] != wantKeys[i] {
			t.Fatalf("keys = %v, want %v (sorted agent then day)", keys, wantKeys)
		}
	}

	r00 := rows[divKey{0, 0}]
	if r00.n != 2 || r00.promoted != 1 {
		t.Errorf("agent 0 day 0: n=%d promoted=%d, want 2/1", r00.n, r00.promoted)
	}
	if got, want := r00.meanOverlap(), (1.0+0.5)/2; math.Abs(got-want) > 1e-9 {
		t.Errorf("agent 0 day 0 mean overlap = %v, want %v", got, want)
	}

	r01 := rows[divKey{0, 1}]
	if got, want := r01.meanOverlap(), 0.8; math.Abs(got-want) > 1e-9 {
		t.Errorf("agent 0 day 1 mean overlap = %v, want %v (k=5 normalizes by its own window)", got, want)
	}
	if r01.promoted != 1 {
		t.Errorf("agent 0 day 1 promoted = %d, want 1 (seq 7)", r01.promoted)
	}
	if r01.meanDisplacement() != 3 {
		t.Errorf("agent 0 day 1 mean displacement = %v, want 3", r01.meanDisplacement())
	}

	r10 := rows[divKey{1, 0}]
	if r10.promoted != 0 {
		t.Errorf("zero-seq (pre-042) entries counted as promoted: %d", r10.promoted)
	}

	if total.n != 4 || total.promoted != 2 || total.vectorlessSum != 5 {
		t.Errorf("total = n %d promoted %d vectorless %d, want 4/2/5", total.n, total.promoted, total.vectorlessSum)
	}
}

// TestDivergencePayloadArithmetic (spec 042 T015): the sim-side constructor's
// overlap/displacement arithmetic over stamped seqs — shared members counted
// once with absolute rank distance; zero seqs excluded from identity.
func TestDivergencePayloadArithmetic(t *testing.T) {
	mem := func(seq, tick int64) sim.Memory { return sim.Memory{Seq: seq, Tick: tick, Text: "m", Salience: 3} }
	legacy := []sim.Memory{mem(10, 5), mem(11, 4), mem(0, 3), mem(12, 2)}
	augmented := []sim.Memory{mem(12, 2), mem(10, 5), mem(0, 3), mem(13, 1)}
	p := sim.NewMemoryDivergencePayload(2, 999, "shadow", legacy, augmented, 3, 800)

	if p.Agent != 2 || p.Tick != 999 || p.Mode != "shadow" || p.Vectorless != 3 || p.SitTick != 800 {
		t.Errorf("scalar fields wrong: %+v", p)
	}
	if len(p.Legacy) != 4 || p.Legacy[2] != 0 || len(p.Augmented) != 4 {
		t.Errorf("windows must ride as seqs in window order (zeros visible): %+v", p)
	}
	// Shared: 10 (rank 0→1, d=1) and 12 (rank 3→0, d=3). The zero seq never
	// matches; 11 and 13 are unshared.
	if p.Overlap != 2 || p.Displacement != 4 {
		t.Errorf("overlap/displacement = %d/%d, want 2/4", p.Overlap, p.Displacement)
	}
}
