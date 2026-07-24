package main

// promptworld divergence — the spec-042 US2 rank-divergence summary
// (contracts/relevance-scoring.md §3): aggregate every recorded
// cog.memory_divergence event into per-agent/per-game-day gate metrics —
// mean overlap@K and the promoted-memory share (selections where relevance
// pulled in at least one memory the legacy ranking excluded) — plus a
// whole-run row. This is the recorded evidence the shadow→on flip (FR-007)
// is decided from; the decision itself is an operator artifact on the board
// task, never an auto-flip. Reads the event log offline (the `tail`
// pattern), so a stopped world summarizes fine.

import (
	"encoding/json"
	"flag"
	"fmt"
	"sort"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/world"
)

// divTicksPerDay converts a selection tick to its game day (1 tick = 1 game
// second, 24h days — the sim's half-life arithmetic uses the same constant).
const divTicksPerDay = 24 * 3600

// divKey groups divergence records: one row per (agent, day).
type divKey struct {
	agent int
	day   int64
}

// divRow is one aggregated summary row. Sums accumulate during the sweep;
// the mean accessors divide at render time.
type divRow struct {
	n               int
	overlapShareSum float64 // Σ overlap/K per selection (K = that selection's window size)
	promoted        int     // selections with ≥1 relevance-promoted memory
	displacementSum int
	vectorlessSum   int
}

func (r divRow) meanOverlap() float64      { return r.overlapShareSum / float64(r.n) }
func (r divRow) promotedShare() float64    { return float64(r.promoted) / float64(r.n) }
func (r divRow) meanDisplacement() float64 { return float64(r.displacementSum) / float64(r.n) }
func (r divRow) meanVectorless() float64   { return float64(r.vectorlessSum) / float64(r.n) }

// add folds one divergence record into the row. Overlap@K normalizes by the
// legacy window's size so k=10 planner and k=5 scene selections aggregate on
// one scale; an empty legacy window is an identical-empty selection (share 1,
// nothing promotable). A promoted memory is a non-zero seq present in the
// augmented window and absent from the legacy one — relevance pulled it in.
func (r *divRow) add(p sim.MemoryDivergencePayload) {
	r.n++
	if len(p.Legacy) == 0 {
		r.overlapShareSum++
	} else {
		r.overlapShareSum += float64(p.Overlap) / float64(len(p.Legacy))
	}
	inLegacy := make(map[int64]bool, len(p.Legacy))
	for _, seq := range p.Legacy {
		if seq != 0 {
			inLegacy[seq] = true
		}
	}
	for _, seq := range p.Augmented {
		if seq != 0 && !inLegacy[seq] {
			r.promoted++
			break
		}
	}
	r.displacementSum += p.Displacement
	r.vectorlessSum += p.Vectorless
}

// aggregateDivergence sweeps the records into sorted (agent, day) rows plus
// the whole-run total — pure, so the arithmetic is unit-testable without a
// store.
func aggregateDivergence(recs []sim.MemoryDivergencePayload) (keys []divKey, rows map[divKey]divRow, total divRow) {
	rows = map[divKey]divRow{}
	for _, p := range recs {
		k := divKey{agent: p.Agent, day: p.Tick / divTicksPerDay}
		row := rows[k]
		row.add(p)
		rows[k] = row
		total.add(p)
	}
	for k := range rows {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].agent != keys[j].agent {
			return keys[i].agent < keys[j].agent
		}
		return keys[i].day < keys[j].day
	})
	return keys, rows, total
}

// divAgentName renders an agent index through the canonical roster, falling
// back to the bare index for out-of-roster values in old logs.
func divAgentName(idx int) string {
	if idx >= 0 && idx < len(sim.AgentNames) {
		return sim.AgentNames[idx]
	}
	return fmt.Sprintf("#%d", idx)
}

func cmdDivergence(args []string) error {
	fs := flag.NewFlagSet("divergence", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "machine-readable output")
	dir, err := parseWorldFlags(fs, args)
	if err != nil {
		return err
	}
	w, err := world.Open(dir)
	if err != nil {
		return err
	}
	// History comes read-only from the store, daemon or not (the tail pattern).
	st, err := store.Open(w.DBPath())
	if err != nil {
		return err
	}
	defer st.Close()

	var recs []sim.MemoryDivergencePayload
	if err := st.ReplayEvents(0, func(e store.Event) error {
		if e.Type != "cog.memory_divergence" {
			return nil
		}
		var p sim.MemoryDivergencePayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return nil // a malformed row is skipped, never fatal to the summary
		}
		recs = append(recs, p)
		return nil
	}); err != nil {
		return err
	}
	if len(recs) == 0 {
		return fmt.Errorf("no cog.memory_divergence events recorded — set world.json memory_relevance to \"shadow\" and run the world")
	}
	keys, rows, total := aggregateDivergence(recs)

	if *jsonOut {
		type jsonRow struct {
			Agent            string  `json:"agent"`
			Day              int64   `json:"day"`
			Selections       int     `json:"selections"`
			MeanOverlapK     float64 `json:"mean_overlap_at_k"`
			PromotedShare    float64 `json:"promoted_share"`
			MeanDisplacement float64 `json:"mean_displacement"`
			MeanVectorless   float64 `json:"mean_vectorless"`
		}
		out := struct {
			Rows  []jsonRow `json:"rows"`
			Total jsonRow   `json:"total"`
		}{}
		for _, k := range keys {
			r := rows[k]
			out.Rows = append(out.Rows, jsonRow{
				Agent: divAgentName(k.agent), Day: k.day, Selections: r.n,
				MeanOverlapK: r.meanOverlap(), PromotedShare: r.promotedShare(),
				MeanDisplacement: r.meanDisplacement(), MeanVectorless: r.meanVectorless(),
			})
		}
		out.Total = jsonRow{Agent: "ALL", Day: -1, Selections: total.n,
			MeanOverlapK: total.meanOverlap(), PromotedShare: total.promotedShare(),
			MeanDisplacement: total.meanDisplacement(), MeanVectorless: total.meanVectorless()}
		b, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	}

	fmt.Printf("rank divergence: %d selections (spec 042 US2 gate evidence — SC-003 wants ≥1 full game day)\n", total.n)
	fmt.Printf("%-8s %5s %6s %14s %15s %14s %12s\n",
		"agent", "day", "n", "mean overlap@K", "promoted-share", "displacement", "vectorless")
	for _, k := range keys {
		r := rows[k]
		fmt.Printf("%-8s %5d %6d %14.2f %15.2f %14.1f %12.1f\n",
			divAgentName(k.agent), k.day, r.n, r.meanOverlap(), r.promotedShare(), r.meanDisplacement(), r.meanVectorless())
	}
	fmt.Printf("%-8s %5s %6d %14.2f %15.2f %14.1f %12.1f\n",
		"ALL", "-", total.n, total.meanOverlap(), total.promotedShare(), total.meanDisplacement(), total.meanVectorless())
	fmt.Println("\nthe shadow→on flip (FR-007) is an operator decision recorded on the board task — cite these numbers.")
	return nil
}
