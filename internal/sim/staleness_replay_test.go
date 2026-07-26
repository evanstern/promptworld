package sim

// Spec 067 US2 (SC-002): the staleness landing gate is a pure function of
// event-sourced state (tick, snapshot tick, effective speed) and recorded
// intent fields — landing outcomes across a mid-flight speed change must
// reproduce identically on replay. Mirrors the governor replay proof pattern
// (governor_replay_test.go): record a live run, re-apply the log to a fresh
// state, and check both the state hash and the gate's re-evaluated verdicts
// against the recorded outcomes.

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// stalenessPreState arranges the world both halves of the proof share: paused
// at tick 10000 (staleness fully controlled, exactly the ladder-harness
// posture), everyone sighted. The arrangement is genesis-side — not
// event-sourced — so the replay half reconstructs it the same way before
// re-applying the log, just as replay reconstructs genesis from the seed.
func stalenessPreState(seed uint64, m *worldmap.Map) *State {
	s := NewState(seed, m)
	s.Paused = true
	s.Tick = 10000
	sightAll(s, s.Tick)
	return s
}

// TestStalenessReplayAcrossSpeedChange (spec 067 US2): a classed
// thought is admitted at the default 4x, a player speed change to 16x lands
// mid-flight, and the intent lands — the gate reads the effective speed at
// the LANDING tick, so staleness 6000 (dead under 4x's budget 4800) is
// forgiven at 16x (budget 19200). A second thought straddles a drop to 1x and
// is rejected with the scaled reason. Replaying the recorded log reproduces
// the state byte-identically AND the pure gate, re-evaluated over the
// replayed event-sourced speed at each recorded landing, reaches the same
// verdicts with the same reasons (FR-002).
func TestStalenessReplayAcrossSpeedChange(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	st, err := store.Open(filepath.Join(t.TempDir(), "world.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	live := stalenessPreState(seed, m)
	loop := NewLoop(live, m, st, nil)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- loop.Run(ctx) }()
	stop := func() {
		cancel()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("loop did not stop")
		}
	}

	// Mid-flight speed rise: the thought was admitted under 4x (snapshot
	// 4000 → staleness 6000 > 4800), but the speed.set to 16x lands first, so
	// the landing-tick budget is 19200 and the intent lands.
	if _, err := loop.Do("set_speed", clock.Speed16x); err != nil {
		stop()
		t.Fatal(err)
	}
	args := meteredArgs(0, "wander")
	args.SnapshotTick = 4000
	if err := loop.InjectIntent(args); err != nil {
		stop()
		t.Fatalf("landing after the mid-flight speed rise rejected: %v", err)
	}

	// Mid-flight speed drop: 16x → 1x tightens the budget under the second
	// thought (deterministic and accepted — spec 067 edge case); staleness
	// 2000 > 1200 × 1 rejects with the scaled reason recorded.
	if _, err := loop.Do("set_speed", clock.Speed1x); err != nil {
		stop()
		t.Fatal(err)
	}
	args = meteredArgs(0, "wander")
	args.SnapshotTick = 8000
	if err := loop.InjectIntent(args); err == nil {
		stop()
		t.Fatal("stale landing at 1x executed")
	}
	stop()

	evs, err := st.EventsSince(0, 0)
	if err != nil {
		t.Fatal(err)
	}

	// Guard: the log must carry both speed changes and both landings, with
	// the expected recorded verdicts, or the test proves nothing.
	var speedSets, landed, rejected int
	for _, e := range evs {
		switch e.Type {
		case "clock.speed_set":
			speedSets++
		case "cog.outcome":
			var p CogOutcomePayload
			if err := json.Unmarshal(e.Payload, &p); err != nil {
				t.Fatal(err)
			}
			switch p.Outcome {
			case OutcomeLanded:
				landed++
			case OutcomeRejectedStale:
				rejected++
				if want := "staleness 2000 > budget 1200 (1200 at 1x × 1x)"; p.Reason != want {
					t.Errorf("recorded rejection reason = %q, want %q", p.Reason, want)
				}
			}
		}
	}
	if speedSets != 2 || landed != 1 || rejected != 1 {
		t.Fatalf("log carried speedSets=%d landed=%d rejected=%d, want 2, 1, 1", speedSets, landed, rejected)
	}

	// Replay from the same genesis arrangement: before applying each recorded
	// cog.outcome, re-evaluate the pure gate over the REPLAYED event-sourced
	// speed — the verdict and reason must reproduce exactly (FR-002); then
	// re-apply the event verbatim, exactly the recovery contract.
	replayed := stalenessPreState(seed, m)
	for _, e := range evs {
		if e.Type == "cog.outcome" {
			var p CogOutcomePayload
			if err := json.Unmarshal(e.Payload, &p); err != nil {
				t.Fatal(err)
			}
			reason := rungStale(p.Class, p.StalenessTicks, replayed.Speed.TicksPerSecond())
			switch p.Outcome {
			case OutcomeLanded, OutcomeAdapted, OutcomeClamped:
				if reason != "" {
					t.Errorf("replayed gate rejects the recorded landing %s: %q", p.Job, reason)
				}
			case OutcomeRejectedStale:
				if reason != p.Reason {
					t.Errorf("replayed rejection reason = %q, recorded %q", reason, p.Reason)
				}
			}
		}
		if err := replayed.Apply(e); err != nil {
			t.Fatalf("replay apply %s: %v", e.Type, err)
		}
		if e.Tick > replayed.Tick {
			replayed.Tick = e.Tick
		}
	}
	if live.Hash() != replayed.Hash() {
		t.Fatalf("staleness replay diverged from live:\nlive:     %s\nreplayed: %s",
			string(live.Marshal()), string(replayed.Marshal()))
	}
	if replayed.Speed != clock.Speed1x {
		t.Errorf("final replayed Speed = %q, want 1x", replayed.Speed)
	}
}
