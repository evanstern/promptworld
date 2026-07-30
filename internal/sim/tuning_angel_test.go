package sim

import (
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/worldmap"
)

// --- spec 102: the angel_cadence_ticks opt-in dial ---

// TestAngelCadenceDefaultOff pins FR-007's compat half: no tuning at all, and
// a tuning manifest that never mentions the dial, both resolve to 0 (off) —
// every pre-102 world keeps a purely event-driven guardian.
func TestAngelCadenceDefaultOff(t *testing.T) {
	s := NewState(1, worldmap.Generate(1, 32, 32))
	if got := s.AngelCadence(); got != 0 {
		t.Fatalf("nil tuning AngelCadence() = %d, want 0 (off)", got)
	}
	parsed, warns, err := ParseTuning([]byte(`{"planner_cadence_ticks": 1800}`))
	if err != nil || len(warns) != 0 {
		t.Fatalf("parse: %v (warns %v)", err, warns)
	}
	if parsed.AngelCadenceTicks != 0 {
		t.Fatalf("absent dial parsed to %d, want 0", parsed.AngelCadenceTicks)
	}
	// A pre-102 pinned world (state carries defaults) must compare Equal to a
	// pre-102 manifest — no redundant tuning event on restart.
	if !parsed.Equal(s.EffectiveTuning()) {
		t.Fatal("pre-102 manifest not Equal to pre-102 effective tuning — restart would append a redundant event")
	}
}

// TestAngelCadenceParseClamp pins the dial's clamp doctrine: 0 = off passes;
// negatives fail toward off; 1..min-1 clamps up to the floor; huge clamps to
// the ceiling; in-band passes untouched. Every clamp warns.
func TestAngelCadenceParseClamp(t *testing.T) {
	cases := []struct {
		raw      string
		want     int64
		wantWarn bool
	}{
		{`{"angel_cadence_ticks": 0}`, 0, false},
		{`{"angel_cadence_ticks": -5}`, 0, true},
		{`{"angel_cadence_ticks": 60}`, 600, true},
		{`{"angel_cadence_ticks": 7200}`, 7200, false},
		{`{"angel_cadence_ticks": 900000}`, 86400, true},
	}
	for _, c := range cases {
		parsed, warns, err := ParseTuning([]byte(c.raw))
		if err != nil {
			t.Fatalf("%s: %v", c.raw, err)
		}
		if parsed.AngelCadenceTicks != c.want {
			t.Errorf("%s → %d, want %d", c.raw, parsed.AngelCadenceTicks, c.want)
		}
		if (len(warns) > 0) != c.wantWarn {
			t.Errorf("%s → warns %v, wantWarn=%v", c.raw, warns, c.wantWarn)
		}
	}
}

// TestAngelCadenceEventRoundTrip pins the event leg: a NewTuningEvent carries
// the dial, the Apply arm resolves it onto state, and the accessor reads it.
// A recorded pre-102 payload (field absent) resolves to off.
func TestAngelCadenceEventRoundTrip(t *testing.T) {
	s := NewState(1, worldmap.Generate(1, 32, 32))
	tuned := defaultTuning()
	tuned.AngelCadenceTicks = 3600
	ev := NewTuningEvent(0, tuned)
	if !strings.Contains(string(ev.Payload), `"angel_cadence_ticks":3600`) {
		t.Fatalf("event payload missing the dial: %s", ev.Payload)
	}
	if err := s.Apply(ev); err != nil {
		t.Fatal(err)
	}
	if got := s.AngelCadence(); got != 3600 {
		t.Fatalf("AngelCadence() = %d after apply, want 3600", got)
	}
	// Pre-102 recorded payload: no angel field → off, never an error.
	s2 := NewState(1, worldmap.Generate(1, 32, 32))
	pre := NewTuningEvent(0, defaultTuning())
	if strings.Contains(string(pre.Payload), "angel_cadence_ticks") {
		// New events DO carry the field (full-set doctrine) — simulate a
		// genuinely old payload by stripping via decode into the old shape.
		old := []byte(`{"refuel_dying_below":10800,"fire_burn_per_wood":14400,"gru_emerge_per_mille":600,"planner_cadence_ticks":1800,"encounter_cooldown_ticks":7200}`)
		pre.Payload = old
	}
	if err := s2.Apply(pre); err != nil {
		t.Fatal(err)
	}
	if got := s2.AngelCadence(); got != 0 {
		t.Fatalf("pre-102 payload resolved AngelCadence() = %d, want 0", got)
	}
}
