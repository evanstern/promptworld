package sim

import (
	"encoding/json"
	"os"
	"sort"
	"testing"
)

// TestMeasureAmbientVolume is the spec-104 T015/SC-001 measurement driver:
// paired seed-1337 baseline (legacy emission) vs fixed (coalescing regime,
// K=10) synthetic headless runs over a month-scale horizon, counting emitted
// rows per family. Gated behind PROMPTWORLD_MEASURE=1 (minutes of CPU) — run:
//
//	PROMPTWORLD_MEASURE=1 go test ./internal/sim/ -run TestMeasureAmbientVolume -v -timeout 60m
//
// The paired worlds legitimately record DIFFERENT event streams (volume is
// the comparison, never bytes — spec.md determinism doctrine); rows are
// normalized per lived game-day since reflex-only survival horizons can
// differ. Results are recorded in
// specs/104-ambient-event-coalescing/measurement.md.
func TestMeasureAmbientVolume(t *testing.T) {
	if os.Getenv("PROMPTWORLD_MEASURE") == "" {
		t.Skip("measurement driver — set PROMPTWORLD_MEASURE=1 to run")
	}
	const seed = 1337
	const days = 29
	const ticks = int64(days) * 86400

	run := func(k int64) (counts map[string]int64, total int64, payloadBytes int64, lastLivingTick int64, endTick int64) {
		m := testMap(seed)
		s := NewState(seed, m)
		tun := defaultTuning()
		tun.NeedsCheckpointMinutes = k // 0 = legacy baseline, 10 = fixed
		if err := s.Apply(NewTuningEvent(0, tun)); err != nil {
			t.Fatal(err)
		}
		counts = map[string]int64{}
		for s.Tick < ticks && !s.Ended {
			next := s.Tick + 1
			s.AdvanceTo(next)
			evs := stepEvents(s, m, next)
			s.Tick = next
			for _, e := range evs {
				if err := s.Apply(e); err != nil {
					t.Fatalf("apply %s at %d: %v", e.Type, next, err)
				}
				counts[e.Type]++
				total++
				payloadBytes += int64(len(e.Payload)) + 80 // envelope estimate: seq/tick/type/wall_time overhead
			}
			if livingCount(s) > 0 {
				lastLivingTick = s.Tick
			}
		}
		return counts, total, payloadBytes, lastLivingTick, s.Tick
	}

	report := func(label string, counts map[string]int64, total, payloadBytes, lastLiving, end int64) float64 {
		days := float64(end) / 86400
		t.Logf("=== %s: %d events over %.2f game-days (last living tick %d, ended=%v) ===",
			label, total, days, lastLiving, end < ticks)
		var types []string
		for typ := range counts {
			types = append(types, typ)
		}
		sort.Slice(types, func(i, j int) bool { return counts[types[i]] > counts[types[j]] })
		for _, typ := range types[:minInt(12, len(types))] {
			t.Logf("  %-28s %9d  (%.1f/game-day)", typ, counts[typ], float64(counts[typ])/days)
		}
		ambient := counts["agent.moved"] + counts["agent.needs_changed"] + counts["gru.moved"] +
			counts["agent.path_started"] + counts["agent.path_truncated"]
		t.Logf("  ambient families total %d (%.1f/game-day, %.1f%% of all rows); ~payload bytes %d",
			ambient, float64(ambient)/days, 100*float64(ambient)/float64(total), payloadBytes)
		return float64(ambient) / days
	}

	bc, bt, bb, bl, be := run(0)
	fc, ft, fb, fl, fe := run(10)
	basePerDay := report("baseline (legacy)", bc, bt, bb, bl, be)
	fixedPerDay := report("fixed (coalesced K=10)", fc, ft, fb, fl, fe)
	ratio := basePerDay / fixedPerDay
	t.Logf("ambient rows/game-day: baseline %.1f vs fixed %.1f — %.1fx reduction", basePerDay, fixedPerDay, ratio)
	if ratio < 4 {
		t.Errorf("SC-001 demands >= 4x ambient reduction, measured %.2fx", ratio)
	}

	// Machine-readable line for measurement.md.
	out, _ := json.Marshal(map[string]any{
		"seed": seed, "horizon_days": days,
		"baseline": map[string]any{"total": bt, "bytes": bb, "end_tick": be,
			"moved": bc["agent.moved"], "needs": bc["agent.needs_changed"], "gru": bc["gru.moved"]},
		"fixed": map[string]any{"total": ft, "bytes": fb, "end_tick": fe,
			"path_started": fc["agent.path_started"], "path_truncated": fc["agent.path_truncated"],
			"needs": fc["agent.needs_changed"], "gru": fc["gru.moved"]},
		"ambient_per_day": map[string]float64{"baseline": basePerDay, "fixed": fixedPerDay, "ratio": ratio},
	})
	t.Logf("MEASUREMENT %s", out)
}
