package sim

import (
	"encoding/json"
	"testing"

	"github.com/evanstern/promptworld/internal/store"
)

// pausedNudgeTimeline scripts the paused authoring loop (spec 040, TASK-77): the
// world pauses, Guardian lands a vision on one villager WHILE paused (the nudge
// spends the genesis charge and appends a dream memory), then the world resumes.
// Applied at tick boundaries exactly as the loop's inject door and replay do.
func pausedNudgeTimeline() map[int64][]store.Event {
	pl := func(v any) json.RawMessage { return mustPayload(v) }
	return map[int64][]store.Event{
		500: {
			{Tick: 500, Type: "clock.paused", Payload: pl(struct{}{})},
			{Tick: 500, Type: "metatron.nudged", Payload: pl(GuardianNudgedPayload{
				Form: "vision", Targets: []int{0}, Text: "the river is rising"})},
			{Tick: 500, Type: "agent.memory_added", Payload: pl(MemoryAddedPayload{
				Agent: 0, Text: "You saw a vision: the river is rising",
				Salience: SalDream, Subject: -1, Origin: OriginOmen})},
		},
		1500: {{Tick: 1500, Type: "clock.resumed", Payload: pl(struct{}{})}},
	}
}

// TestPausedNudgeReplayByteIdentical (spec 040 FR-006/FR-007, SC-004): a run whose
// log contains a paused nudge session — clock.paused, the metatron.nudged spend +
// dream memory landed while frozen, and clock.resumed — replays byte-identically
// from genesis. Every one of those events is a pure function of the log (charge
// spend, memory append, and the paused flag all reduce deterministically), so the
// paused-authoring exception adds no wall-clock or out-of-band state to replay.
func TestPausedNudgeReplayByteIdentical(t *testing.T) {
	const seed, ticks = 456, 3000
	m := testMap(seed)

	live := NewState(seed, m)
	log := driveTicks(t, live, m, ticks, pausedNudgeTimeline())

	// Guard: the log must actually carry the paused nudge session, or the test
	// proves nothing.
	var paused, nudged, resumed int
	for _, e := range log {
		switch e.Type {
		case "clock.paused":
			paused++
		case "metatron.nudged":
			nudged++
		case "clock.resumed":
			resumed++
		}
	}
	if paused != 1 || nudged != 1 || resumed != 1 {
		t.Fatalf("timeline carried paused=%d nudged=%d resumed=%d, want 1 each", paused, nudged, resumed)
	}
	// The vision spent the genesis charge — proof the nudge actually reduced.
	if live.GuardianCharges != GuardianGenesisCharges-1 {
		t.Fatalf("charges = %d after the vision, want %d (the spend must reduce)", live.GuardianCharges, GuardianGenesisCharges-1)
	}

	// Replay from genesis: reduce the logged events, align the clock, re-live the
	// quiet tail — exactly the recovery contract (mirrors TestGovernorReplayByteIdentical).
	replayed := NewState(seed, m)
	for _, e := range log {
		if err := replayed.Apply(e); err != nil {
			t.Fatalf("replay apply %s: %v", e.Type, err)
		}
		replayed.Tick = e.Tick
	}
	driveTicks(t, replayed, m, ticks, nil)

	if live.Hash() != replayed.Hash() {
		t.Fatalf("paused-nudge replay diverged from live:\nlive:     %s\nreplayed: %s",
			string(live.Marshal()), string(replayed.Marshal()))
	}
}
