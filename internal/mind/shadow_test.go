package mind

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// shadowFixtureState builds one deterministic replica for the invariant test:
// agent 0 carries an old, low-salience, situation-matching memory (vectored)
// buried under louder unrelated ones, plus a recorded situation vector — a
// state where the augmented ranking PROVABLY differs from the legacy one, so
// the shadow invariant is tested where it bites, not where the rankings
// trivially agree.
func shadowFixtureState() (*sim.State, *worldmap.Map) {
	m := worldmap.Generate(42, 64, 64)
	s := sim.NewState(42, m)
	s.Tick = 130_000
	a := &s.Agents[0]
	a.Memories = append(a.Memories, sim.Memory{
		Text: "Robbed at the river.", Salience: 2, Tick: 100, Subject: -1,
		Seq: 1, Vec: []float32{1, 0}, VecModel: "all-minilm",
	})
	for i := int64(1); i <= 40; i++ {
		a.Memories = append(a.Memories, sim.Memory{
			Text: "Routine day.", Salience: 5, Tick: 100_000 + i*600, Subject: -1,
			Seq: 1 + i, Vec: []float32{0, 1}, VecModel: "all-minilm",
		})
	}
	a.SitVec = []float32{1, 0}
	a.SitVecModel = "all-minilm"
	a.SitVecTick = 129_000
	return s, m
}

// shadowRun drives one bare-Mind plan pass (the TestRunningNudgeDoesNotArm
// pattern — absorb-goroutine code run inline, race-free) in the given mode and
// returns the enqueued jobs' prompts plus every cog.* event the pass injected.
func shadowRun(t *testing.T, mode string) (prompts map[int]string, injected []store.Event) {
	t.Helper()
	s, m := shadowFixtureState()
	inj := &recordingInjector{ch: make(chan []store.Event, 64)}
	md := &Mind{
		replica: s, m: m, k: sim.WindowK,
		social:          inj,
		planQ:           make(chan planJob, sim.AgentCount),
		memoryRelevance: mode,
	}
	md.plan()

	prompts = map[int]string{}
	for {
		select {
		case job := <-md.planQ:
			prompts[job.agent] = job.system + "\n---\n" + job.prompt
			continue
		default:
		}
		break
	}
	// Divergence injections detach (go emitCog); drain to quiescence — a
	// fresh grace window after every arrival, so slow goroutine scheduling
	// never undercounts. In "" mode nothing may arrive — the first quiet
	// window is what proves the negative.
	for {
		select {
		case batch := <-inj.ch:
			injected = append(injected, batch...)
		case <-time.After(500 * time.Millisecond):
			return prompts, injected
		}
	}
}

// TestShadowInvariant (spec 042 T017, FR-006, contracts/relevance-scoring.md
// §3): with memory_relevance "shadow", every enqueued plan job's prompt is
// BYTE-IDENTICAL to "" mode on the same state — the augmented ranking provably
// differs here, yet nothing about the cognition changes. Divergence events are
// present in shadow (one per enqueued job, mode-stamped, seq-listed) and
// absent in "" mode; applied to state they are reducer no-ops, so replay
// stays green.
func TestShadowInvariant(t *testing.T) {
	offPrompts, offEvents := shadowRun(t, "")
	shadowPrompts, shadowEvents := shadowRun(t, "shadow")

	// The invariant: prompts byte-identical across modes, per agent.
	if len(offPrompts) == 0 {
		t.Fatal("no plan jobs enqueued — the fixture never exercised the prompt path")
	}
	if len(shadowPrompts) != len(offPrompts) {
		t.Fatalf("shadow enqueued %d jobs, off enqueued %d — behavior changed", len(shadowPrompts), len(offPrompts))
	}
	for agent, off := range offPrompts {
		if shadow, ok := shadowPrompts[agent]; !ok || shadow != off {
			t.Errorf("agent %d prompt differs between \"\" and \"shadow\" — the shadow invariant is broken:\noff:    %q\nshadow: %q", agent, off, shadowPrompts[agent])
		}
	}

	// Off mode records nothing.
	for _, e := range offEvents {
		if e.Type == "cog.memory_divergence" {
			t.Errorf("\"\" mode emitted a divergence event: %s", e.Payload)
		}
	}

	// Shadow records one divergence per enqueued job, and agent 0's proves the
	// rankings genuinely diverged (the augmented window pulled the old
	// situation-matching memory in — displacement/overlap reflect it).
	var divs []sim.MemoryDivergencePayload
	for _, e := range shadowEvents {
		if e.Type != "cog.memory_divergence" {
			continue
		}
		var p sim.MemoryDivergencePayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			t.Fatal(err)
		}
		divs = append(divs, p)
	}
	if len(divs) != len(shadowPrompts) {
		t.Fatalf("%d divergence events for %d plan jobs, want one per job", len(divs), len(shadowPrompts))
	}
	var agent0 *sim.MemoryDivergencePayload
	for i := range divs {
		if divs[i].Mode != "shadow" {
			t.Errorf("divergence mode = %q, want shadow", divs[i].Mode)
		}
		if divs[i].Agent == 0 {
			agent0 = &divs[i]
		}
	}
	if agent0 == nil {
		t.Fatal("no divergence record for agent 0")
	}
	if agent0.SitTick != 129_000 {
		t.Errorf("agent 0 sit_tick = %d, want the recorded situation tick 129000", agent0.SitTick)
	}
	if len(agent0.Legacy) != sim.WindowK || len(agent0.Augmented) != sim.WindowK {
		t.Errorf("windows = %d/%d seqs, want %d each", len(agent0.Legacy), len(agent0.Augmented), sim.WindowK)
	}
	inAugmented := false
	for _, seq := range agent0.Augmented {
		if seq == 1 { // the old situation-matching memory's seq
			inAugmented = true
		}
	}
	if !inAugmented {
		t.Error("the situation-matching memory (seq 1) never entered the augmented window — the fixture lost its bite")
	}
	if agent0.Overlap == len(agent0.Legacy) && agent0.Displacement == 0 {
		t.Error("rankings identical for agent 0 — the invariant was tested where it cannot bite")
	}

	// Replay green: divergence events are reducer no-ops — state bytes
	// unchanged by applying them.
	s, _ := shadowFixtureState()
	before := string(s.Marshal())
	for _, e := range shadowEvents {
		if e.Type != "cog.memory_divergence" {
			continue
		}
		if err := s.Apply(e); err != nil {
			t.Fatalf("apply divergence event: %v", err)
		}
	}
	if got := string(s.Marshal()); got != before {
		t.Error("cog.memory_divergence mutated state — replay would diverge")
	}
}

// TestOnModeConsumesRelevantWindow (spec 042 T019, contracts §3 row 3): with
// memory_relevance "on", the SAME fixture state feeds the plan prompt the
// relevance-augmented window — the old situation-matching memory the legacy
// window excludes now reaches the model — and divergence is still recorded,
// mode-stamped "on". Off-mode prompts remain free of it, so this is the one
// mode that changes agent-visible behavior, exactly as gated.
func TestOnModeConsumesRelevantWindow(t *testing.T) {
	offPrompts, _ := shadowRun(t, "")
	onPrompts, onEvents := shadowRun(t, "on")

	const marker = "Robbed at the river."
	off, on := offPrompts[0], onPrompts[0]
	if off == "" || on == "" {
		t.Fatal("agent 0 enqueued no plan job")
	}
	if strings.Contains(off, marker) {
		t.Fatalf("legacy window already carries the old memory — fixture lost its bite:\n%s", off)
	}
	if !strings.Contains(on, marker) {
		t.Errorf("\"on\" prompt missing the relevance-promoted memory %q:\n%s", marker, on)
	}

	// Divergence keeps recording in "on" (contract §3), mode-stamped.
	found := false
	for _, e := range onEvents {
		if e.Type != "cog.memory_divergence" {
			continue
		}
		var p sim.MemoryDivergencePayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			t.Fatal(err)
		}
		if p.Agent == 0 {
			found = true
			if p.Mode != "on" {
				t.Errorf("divergence mode = %q, want on", p.Mode)
			}
		}
	}
	if !found {
		t.Error("\"on\" mode stopped recording divergence — the guardrail must keep running")
	}
}
