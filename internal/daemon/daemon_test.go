package daemon

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/cognition"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/world"
)

// openWorldWithMeeting writes a manifest carrying a meeting block and opens it.
func openWorldWithMeeting(t *testing.T, meeting string) *world.World {
	t.Helper()
	dir := t.TempDir()
	manifest := `{"name":"w","seed":42,"format_version":4,"tick_game_seconds":1` + meeting + `}`
	if err := os.WriteFile(filepath.Join(dir, world.ManifestName), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	w, err := world.Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return w
}

// TestSeedMeetingConventionConfig (TASK-36 AC#2): a manifest-declared
// convention lands as the establishing event on boot, sets state, and rides
// the log so a replayed daemon re-establishes it without re-injecting.
func TestSeedMeetingConventionConfig(t *testing.T) {
	w := openWorldWithMeeting(t, `,"meeting":{"convene":"11:30","open":"12:00","x":7,"y":9}`)
	st, err := store.Open(w.DBPath())
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	state := sim.NewState(w.Manifest.Seed, w.Map())
	if err := seedMeetingConvention(w, st, state); err != nil {
		t.Fatalf("seed: %v", err)
	}

	c := state.MeetingConvention
	if c == nil {
		t.Fatal("boot seed did not establish the convention")
	}
	if c.ConveneSecond != 11*3600+1800 || c.OpenSecond != 12*3600 || c.Source != "config" {
		t.Errorf("convention = %+v, want 11:30/12:00/config", c)
	}
	if state.MeetingPlace == nil || state.MeetingPlace.X != 7 || state.MeetingPlace.Y != 9 {
		t.Errorf("meeting place = %+v, want the config coords (7,9)", state.MeetingPlace)
	}

	// The event is persisted: a fresh state rebuilt from the log carries it.
	replayed := sim.NewState(w.Manifest.Seed, w.Map())
	var establishes int
	if err := st.ReplayEvents(0, func(e store.Event) error {
		if e.Type == "meeting.convention_established" {
			establishes++
		}
		return replayed.Apply(e)
	}); err != nil {
		t.Fatal(err)
	}
	if establishes != 1 {
		t.Errorf("%d convention_established events in the log, want exactly one", establishes)
	}
	if replayed.MeetingConvention == nil || *replayed.MeetingConvention != *c {
		t.Errorf("replay convention %+v != live %+v", replayed.MeetingConvention, c)
	}

	// Idempotent: a second boot with the convention already in state injects
	// nothing (the guard), so no second event lands.
	if err := seedMeetingConvention(w, st, replayed); err != nil {
		t.Fatal(err)
	}
	var after int
	st.ReplayEvents(0, func(e store.Event) error {
		if e.Type == "meeting.convention_established" {
			after++
		}
		return nil
	})
	if after != 1 {
		t.Errorf("re-seed injected a duplicate: %d establish events", after)
	}
}

// TestSeedMeetingConventionAbsent: no manifest meeting block → no convention,
// no event (emergent default).
func TestSeedMeetingConventionAbsent(t *testing.T) {
	w := openWorldWithMeeting(t, ``)
	st, err := store.Open(w.DBPath())
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	state := sim.NewState(w.Manifest.Seed, w.Map())
	if err := seedMeetingConvention(w, st, state); err != nil {
		t.Fatal(err)
	}
	if state.MeetingConvention != nil {
		t.Errorf("no manifest meeting block should leave the world convention-less, got %+v", state.MeetingConvention)
	}
	var events int
	st.ReplayEvents(0, func(e store.Event) error { events++; return nil })
	if events != 0 {
		t.Errorf("%d events written for a convention-less boot, want none", events)
	}
}

// TestSeedMeetingConventionNoCoords: a manifest meeting without x/y derives the
// place, so the convention still lands with a concrete meeting place.
func TestSeedMeetingConventionNoCoords(t *testing.T) {
	w := openWorldWithMeeting(t, `,"meeting":{"convene":"11:30","open":"12:00"}`)
	st, err := store.Open(w.DBPath())
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	state := sim.NewState(w.Manifest.Seed, w.Map())
	if err := seedMeetingConvention(w, st, state); err != nil {
		t.Fatal(err)
	}
	if state.MeetingConvention == nil || state.MeetingPlace == nil {
		t.Fatalf("convention/place missing: conv=%+v place=%+v", state.MeetingConvention, state.MeetingPlace)
	}
}

// TestUncalibratedBootWarningContainsContractElements (spec 035 US2, T011):
// the boot warning block carries all three contracts/warnings.md §1
// elements — the uncalibrated statement, the per-class horizon at bootstrap
// seeds, and the exact calibrate command for this world.
func TestUncalibratedBootWarningContainsContractElements(t *testing.T) {
	got := uncalibratedBootWarning("demo-ux")
	if !strings.Contains(got, "UNCALIBRATED") || !strings.Contains(got, "bootstrap defaults") {
		t.Errorf("missing the uncalibrated statement: %q", got)
	}
	if !strings.Contains(got, "planner") || !strings.Contains(got, "conversation") {
		t.Errorf("missing the per-class horizon summary: %q", got)
	}
	if !strings.Contains(got, "promptworld calibrate demo-ux") {
		t.Errorf("missing the exact calibrate command with the real world name: %q", got)
	}
}

// TestTeachingPostureReplayByteIdentical (spec 039 T017/R3, spec 036 replay
// doctrine): the boot-time posture default is applied through the loop's normal
// set_speed command door, so it lands as a recorded clock.speed_set event. The
// world's log therefore replays to a byte-identical state — the posture speed
// included — with no unrecorded manifest-aware divergence. Initial sim state is
// left at clock.DefaultSpeed (untouched); only the recorded event moves it.
func TestTeachingPostureReplayByteIdentical(t *testing.T) {
	dir := t.TempDir() + "/w"
	w, err := world.Create(dir, "teach", 42)
	if err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(w.DBPath())
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	state := sim.NewState(w.Manifest.Seed, w.Map())
	if state.Speed != clock.DefaultSpeed {
		t.Fatalf("initial sim speed = %s, want the untouched default %s", state.Speed, clock.DefaultSpeed)
	}

	ctx, cancel := context.WithCancel(context.Background())
	loop := sim.NewLoop(state, w.Map(), st, nil)
	done := make(chan error, 1)
	go func() { done <- loop.Run(ctx) }()

	// Apply the posture default exactly as daemon boot does — the normal command
	// door — so it is recorded.
	if _, err := loop.Do("set_speed", clock.Speed16x); err != nil {
		t.Fatalf("posture set_speed: %v", err)
	}
	// Let a few ticks accrue around the recorded event.
	deadline := time.Now().Add(5 * time.Second)
	for {
		if s, err := loop.Do("status", ""); err == nil && s.Tick >= 3 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("world did not advance past tick 3")
		}
		time.Sleep(20 * time.Millisecond)
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("loop.Run: %v", err)
	}
	live := state.Marshal()

	// Exactly one recorded clock.speed_set(16x) carries the posture.
	events, err := st.EventsSince(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	var speedSets int
	for _, e := range events {
		if e.Type != "clock.speed_set" {
			continue
		}
		speedSets++
		var p sim.SpeedSetPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			t.Fatal(err)
		}
		if p.Speed != clock.Speed16x {
			t.Errorf("recorded speed_set = %s, want 16x", p.Speed)
		}
	}
	if speedSets != 1 {
		t.Errorf("%d clock.speed_set events, want exactly 1 (the posture default)", speedSets)
	}

	// Replay the log through the daemon's own recovery path: byte-identical state.
	replayed, err := recoverState(w, st)
	if err != nil {
		t.Fatalf("recoverState: %v", err)
	}
	if replayed.Speed != clock.Speed16x {
		t.Errorf("replayed speed = %s, want 16x (the posture rode the recorded event)", replayed.Speed)
	}
	if got := string(replayed.Marshal()); got != string(live) {
		t.Errorf("replay diverged from the live state:\nlive:     %s\nreplayed: %s", live, got)
	}
}

// TestUncalibratedBootWarningNoTeachingFlavor (spec 039 US3 AC3, FR-008): the
// spec-035 uncalibrated boot warning is untouched — it carries no teaching or
// provisional-posture wording, so a NON-teaching uncalibrated world's boot
// output is byte-identical to pre-039. The teaching flavor lives only in
// teachingPostureBootLine, which non-teaching boots never call.
func TestUncalibratedBootWarningNoTeachingFlavor(t *testing.T) {
	got := uncalibratedBootWarning("plain")
	for _, banned := range []string{"teaching", "posture", "provisional"} {
		if strings.Contains(got, banned) {
			t.Errorf("uncalibrated boot warning leaked teaching wording %q: %q", banned, got)
		}
	}
}

// TestTeachingPostureBootLineCalibrated (spec 039 US1, contracts/posture.md §2):
// a measured planner estimate renders the plain "defaulting speed to Nx" line
// with the calibration timestamp and no provisional/calibrate prompt.
func TestTeachingPostureBootLineCalibrated(t *testing.T) {
	got := teachingPostureBootLine("teach", clock.Speed16x, 17.0, "2026-07-24T12:00:00Z")
	for _, want := range []string{"teaching posture", "defaulting speed to 16x", "planner-safe at 17.0s/pt", "calibrated 2026-07-24T12:00:00Z"} {
		if !strings.Contains(got, want) {
			t.Errorf("calibrated boot line missing %q: %q", want, got)
		}
	}
	if strings.Contains(got, "provisional") || strings.Contains(got, "promptworld calibrate") {
		t.Errorf("calibrated boot line must not prompt calibrate: %q", got)
	}
}

// TestTeachingPostureBootLineProvisional (spec 039 US3, contracts/posture.md §2):
// a bootstrap-seeded planner (empty calibratedAt) marks the rung provisional and
// appends the explicit `promptworld calibrate <world>` prompt.
func TestTeachingPostureBootLineProvisional(t *testing.T) {
	got := teachingPostureBootLine("teach", clock.Speed16x, 20.0, "")
	for _, want := range []string{"defaulting speed to 16x", "provisional", "planner-safe at 20.0s/pt bootstrap estimate", "run `promptworld calibrate teach`"} {
		if !strings.Contains(got, want) {
			t.Errorf("provisional boot line missing %q: %q", want, got)
		}
	}
}

// TestTeachingPostureSpeedDerivation (spec 039 US1/SC-005): the boot default is
// clock.SpeedForRate(cognition.MaxSafeSpeed("planner", est)) — the exact
// derivation daemon boot applies. A slower profile yields a lower rung; a
// pathological one clamps to the 1x floor; the number follows the profile with
// no stored value.
func TestTeachingPostureSpeedDerivation(t *testing.T) {
	cases := []struct {
		est  float64
		want clock.Speed
	}{
		{5.0, clock.Speed32x},
		{17.0, clock.Speed16x},
		{20.0, clock.Speed16x},
		{1000.0, clock.Speed1x}, // even 1x suppresses ⇒ clamp to floor
	}
	for _, c := range cases {
		got := clock.SpeedForRate(cognition.MaxSafeSpeed("planner", c.est))
		if got != c.want {
			t.Errorf("posture speed at %.1fs/pt = %s, want %s", c.est, got, c.want)
		}
	}
}

// TestSeedSurvivalWatches (spec 059 US1, SC-001): a fresh world's boot seeds the
// three system-origin survival watches once; they ride the log so a replayed
// daemon re-establishes them; and a second boot with them already standing
// injects nothing (replay-safe idempotence).
func TestSeedSurvivalWatches(t *testing.T) {
	w := openWorldWithMeeting(t, ``)
	st, err := store.Open(w.DBPath())
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	state := sim.NewState(w.Manifest.Seed, w.Map())
	if err := seedSurvivalWatches(w, st, state); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// State now carries exactly the three canonical survival watches, all active,
	// all system origin, one per kind.
	kinds := map[string]int{}
	for i := range state.GuardianOrders {
		o := state.GuardianOrders[i]
		if o.Survival == "" {
			continue
		}
		if o.Origin != sim.GuardianOriginSystem || o.Status != "active" {
			t.Errorf("survival watch %s not active/system: %+v", o.ID, o)
		}
		kinds[o.Survival]++
	}
	for _, k := range []string{sim.SurvivalNearDeath, sim.SurvivalStarvation, sim.SurvivalExposure} {
		if kinds[k] != 1 {
			t.Errorf("survival kind %q seeded %d times, want 1", k, kinds[k])
		}
	}

	// The events persist: a fresh state rebuilt from the log carries them.
	replayed := sim.NewState(w.Manifest.Seed, w.Map())
	var placed int
	if err := st.ReplayEvents(0, func(e store.Event) error {
		if e.Type == "metatron.order_placed" {
			placed++
		}
		return replayed.Apply(e)
	}); err != nil {
		t.Fatal(err)
	}
	if placed != 3 {
		t.Errorf("%d order_placed events in the log, want exactly 3", placed)
	}

	// Idempotent: a second boot with the watches already standing injects nothing.
	if err := seedSurvivalWatches(w, st, replayed); err != nil {
		t.Fatal(err)
	}
	var after int
	st.ReplayEvents(0, func(e store.Event) error {
		if e.Type == "metatron.order_placed" {
			after++
		}
		return nil
	})
	if after != 3 {
		t.Errorf("re-seed injected duplicates: %d order_placed events, want 3", after)
	}
}
