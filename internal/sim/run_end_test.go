package sim

// Run-outcome tests (spec 044 US1, T012): the run-end declaration fires
// exactly once, ordered after every same-tick death; an ended world emits
// nothing forever after; replay rebuilds the ended posture; the loop's
// command door refuses mutation once ended; and the new State fields are
// omitempty-stable so pre-044 snapshots stay byte-identical.

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/store"
)

// starveAll stages every villager one heartbeat from a starvation death —
// the TestStarvationDeath posture, reused as the run-end trigger.
func starveAll(s *State) {
	for i := range s.Agents {
		s.Agents[i].Needs.Food = 0
		s.Agents[i].Needs.Health = 3
	}
}

// TestRunEndedOnceOrderedLast (FR-001, edge "two die on the same tick"):
// every villager dies on the same heartbeat — the hardest same-tick case —
// and exactly one run.ended lands, in that same batch, after every
// agent.died, as the batch's last event, carrying the full death ledger in
// event order.
func TestRunEndedOnceOrderedLast(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	starveAll(s)
	log := driveTicks(t, s, m, 120, nil)

	var died []DeathRecord
	var ends []store.Event
	endIdx, lastIdx := -1, -1
	for i, e := range log {
		switch e.Type {
		case "agent.died":
			var p DiedPayload
			if err := json.Unmarshal(e.Payload, &p); err != nil {
				t.Fatal(err)
			}
			died = append(died, DeathRecord{Agent: p.Agent.ID, Tick: e.Tick, Cause: p.Cause})
		case "run.ended":
			ends = append(ends, e)
			endIdx = i
		}
		lastIdx = i
	}
	if len(died) != agentCount {
		t.Fatalf("%d/%d agents died", len(died), agentCount)
	}
	if len(ends) != 1 {
		t.Fatalf("run.ended fired %d times, want exactly once", len(ends))
	}
	if endIdx != lastIdx {
		t.Errorf("run.ended is event %d of %d — it must close the batch (nothing may follow it)", endIdx, lastIdx)
	}
	if ends[0].Tick != died[len(died)-1].Tick {
		t.Errorf("run.ended tick %d, want the final death's tick %d (same batch, no extra tick)",
			ends[0].Tick, died[len(died)-1].Tick)
	}
	var p RunEndedPayload
	if err := json.Unmarshal(ends[0].Payload, &p); err != nil {
		t.Fatal(err)
	}
	if len(p.Deaths) != agentCount {
		t.Fatalf("payload carries %d deaths, want %d", len(p.Deaths), agentCount)
	}
	for i, d := range p.Deaths {
		if (DeathRecord{Agent: d.Agent.ID, Tick: d.Tick, Cause: d.Cause}) != died[i] {
			t.Errorf("payload death %d = %+v, want %+v (event order)", i, d, died[i])
		}
	}
	if want := died[len(died)-1].Cause; p.FinalCause != want {
		t.Errorf("FinalCause = %q, want %q", p.FinalCause, want)
	}
	if !s.Ended || s.RunEnd == nil {
		t.Fatal("reducer did not latch Ended/RunEnd")
	}
	if s.RunEnd.Tick != p.Tick || s.RunEnd.FinalCause != p.FinalCause || len(s.RunEnd.Deaths) != len(p.Deaths) {
		t.Errorf("State.RunEnd %+v does not mirror the payload %+v", s.RunEnd, p)
	}
}

// TestEndedWorldEmitsNothing (FR-002): once ended, further ticks produce no
// events at all — simulated time is frozen by the executor's guard, not by
// the loop's pacing.
func TestEndedWorldEmitsNothing(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	s := NewState(seed, m)
	starveAll(s)
	driveTicks(t, s, m, 120, nil)
	if !s.Ended {
		t.Fatal("world should have ended")
	}
	tail := driveTicks(t, s, m, 24*3600, nil) // a full further game day
	if len(tail) != 0 {
		t.Fatalf("ended world emitted %d events (first: %s)", len(tail), tail[0].Type)
	}
}

// TestReplayRebuildsEnded (FR-004, the TestReplayRebuildsState pattern):
// reducing the logged events over genesis reproduces the ended posture —
// which is exactly what a daemon restart's recovery does.
func TestReplayRebuildsEnded(t *testing.T) {
	const seed = 42
	m := testMap(seed)
	live := NewState(seed, m)
	starveAll(live)
	// Stage the same pre-drive mutation on the replay side: the staging is
	// test scaffolding standing in for recorded history, so both sides must
	// share it (the real recovery path replays needs events instead).
	log := driveTicks(t, live, m, 120, nil)

	replayed := NewState(seed, m)
	starveAll(replayed)
	for _, e := range log {
		if err := replayed.Apply(e); err != nil {
			t.Fatalf("replay apply %s: %v", e.Type, err)
		}
		replayed.Tick = e.Tick
	}
	driveTicks(t, replayed, m, 120, nil) // quiet tail, as recovery re-lives it

	if !replayed.Ended {
		t.Fatal("replay did not rebuild Ended")
	}
	if live.Hash() != replayed.Hash() {
		t.Fatalf("replayed ended state diverged from live:\nlive:     %s\nreplayed: %s",
			string(live.Marshal()), string(replayed.Marshal()))
	}
}

// TestRunEndOmitemptyStable (T002): the three spec-044 State additions are
// omitempty, so a living world's snapshot carries none of their keys and a
// pre-044 snapshot round-trips byte-identically — no format_version bump.
func TestRunEndOmitemptyStable(t *testing.T) {
	s := NewState(42, testMap(42))
	pre := s.Marshal()
	for _, key := range []string{`"deaths"`, `"ended"`, `"run_end"`} {
		if bytes.Contains(pre, []byte(key)) {
			t.Errorf("living-world snapshot leaked the spec-044 key %s:\n%s", key, pre)
		}
	}
	var back State
	if err := json.Unmarshal(pre, &back); err != nil {
		t.Fatal(err)
	}
	if got := back.Marshal(); !bytes.Equal(got, pre) {
		t.Errorf("pre-044 snapshot did not round-trip byte-identically:\npre:  %s\npost: %s", pre, got)
	}
	// An ended world does serialize the posture (it must survive snapshots).
	end := RunEndedPayload{Tick: 60, Deaths: DeathRefs([]DeathRecord{{Agent: 0, Tick: 60, Cause: "starvation"}}), FinalCause: "starvation"}
	if err := s.Apply(store.Event{Tick: 60, Type: "run.ended", Payload: mustPayload(end)}); err != nil {
		t.Fatal(err)
	}
	got := s.Marshal()
	if !bytes.Contains(got, []byte(`"ended":true`)) || !bytes.Contains(got, []byte(`"run_end":{"tick":60`)) {
		t.Errorf("ended snapshot missing posture keys:\n%s", got)
	}
}

// newEndedHarness builds a Loop (goroutine not running — the newGovernHarness
// pattern) over a world whose run has ended through the real event path.
func newEndedHarness(t *testing.T) *Loop {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "world.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	m := testMap(7)
	s := NewState(7, m)
	end := RunEndedPayload{Tick: 1, Deaths: DeathRefs([]DeathRecord{{Agent: 0, Tick: 1, Cause: "starvation"}}), FinalCause: "starvation"}
	if err := s.Apply(store.Event{Tick: 1, Type: "run.ended", Payload: mustPayload(end)}); err != nil {
		t.Fatal(err)
	}
	return NewLoop(s, m, st, nil)
}

// runCommand pushes one command straight through handleCommand and returns
// the reply.
func runCommand(t *testing.T, l *Loop, cmd command) commandResult {
	t.Helper()
	cmd.reply = make(chan commandResult, 1)
	if err := l.handleCommand(cmd); err != nil {
		t.Fatalf("handleCommand(%s): %v", cmd.name, err)
	}
	return <-cmd.reply
}

// TestEndedCommandGating (FR-002/FR-003, contracts/status.md): an ended
// world refuses every clock/world-mutating command with an explicit "run has
// ended" error and emits nothing; reads serve, and inject_social narrows to
// recorded prose (chronicle.entry today).
func TestEndedCommandGating(t *testing.T) {
	l := newEndedHarness(t)
	seqBefore := l.st.LastSeq()

	refused := []command{
		{name: "pause"},
		{name: "resume"},
		{name: "set_speed", speed: clock.Speed16x},
		{name: "govern", govern: &governArgs{to: clock.Speed16x, debt: 1.0, jobs: 1}},
		{name: "inject_intent", inject: &InjectArgs{Agent: 0, Goal: "forage", TargetAgent: -1}},
		{name: "inject_social", social: []store.Event{
			{Type: "social.rumor_told", Payload: mustPayload(RumorToldPayload{From: Ref(0), To: Ref(1), Subject: Ref(2), Text: "x"})},
		}},
	}
	for _, cmd := range refused {
		res := runCommand(t, l, cmd)
		if res.err == nil || !strings.Contains(res.err.Error(), "run has ended") {
			t.Errorf("%s on an ended world: err = %v, want a \"run has ended\" refusal", cmd.name, res.err)
		}
	}
	if l.st.LastSeq() != seqBefore {
		t.Errorf("refused commands appended events: last seq %d → %d", seqBefore, l.st.LastSeq())
	}

	// Reads keep serving, and status reports the posture.
	res := runCommand(t, l, command{name: "status"})
	if res.err != nil {
		t.Errorf("status on an ended world errored: %v", res.err)
	}
	if !res.status.Ended || res.status.EndedDay == 0 {
		t.Errorf("status = %+v, want Ended=true with a non-zero EndedDay", res.status)
	}
	if res.status.EffectiveRate != 0 {
		t.Errorf("EffectiveRate = %v on an ended world, want 0", res.status.EffectiveRate)
	}
	res = runCommand(t, l, command{name: "state"})
	if res.err != nil || len(res.state) == 0 {
		t.Errorf("state on an ended world: err=%v, %d state bytes", res.err, len(res.state))
	}

	// Recorded prose about the ended run still lands (contracts/status.md).
	res = runCommand(t, l, command{name: "inject_social", social: []store.Event{
		{Type: "chronicle.entry", Payload: mustPayload(ChronicleEntryPayload{Day: 1, Text: "The village fell."})},
	}})
	if res.err != nil {
		t.Errorf("chronicle.entry on an ended world refused: %v", res.err)
	}
	if l.st.LastSeq() != seqBefore+1 {
		t.Errorf("chronicle.entry did not land: last seq %d, want %d", l.st.LastSeq(), seqBefore+1)
	}
}
