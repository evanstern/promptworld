package guardian

import (
	"os"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

// --- spec 102 T002: the guardian's own memory + nightly consolidation ---

// seedGuardianMemory applies one guardian.memory_added to BOTH the replica
// and the injector-backed state (what live absorption would produce).
func seedGuardianMemory(t *testing.T, mt *Guardian, inj *stateInjector, tick, seq int64, text string, sal int) {
	t.Helper()
	ev := store.Event{Tick: tick, Seq: seq, Type: "guardian.memory_added",
		Payload: mustJSON(sim.GuardianMemoryPayload{Text: text, Salience: sal})}
	if err := mt.replica.Apply(ev); err != nil {
		t.Fatal(err)
	}
	if err := inj.state.Apply(ev); err != nil {
		t.Fatal(err)
	}
	mt.mirrorState()
}

// TestRecordMemoryGatedOnAgentization pins FR-007's live half: a non-opted
// world's guardian records NOTHING to the store — the door never sees a
// guardian.memory_added — while an opted-in world's does.
func TestRecordMemoryGatedOnAgentization(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "quiet")
	mt.recordMemory("should not land", 5)
	if len(inj.batches) != 0 {
		t.Fatal("non-agentized recordMemory reached the door")
	}
	optIn(t, mt, inj, 600)
	mt.recordMemory("the first remembered moment", 5)
	if len(inj.batches) != 1 || inj.batches[0][0].Type != "guardian.memory_added" {
		t.Fatalf("agentized recordMemory batches: %+v", inj.batches)
	}
	if len(inj.state.GuardianMemories) != 1 {
		t.Fatalf("state store: %+v", inj.state.GuardianMemories)
	}
}

// TestGuardianNightConsolidates drives one full agentized night: buffer →
// shared dream pass (vectorless store: no geometry events) → one
// KindConsolidation call → promote/fade/gist/marker landed as one atomic
// batch through the door, high-water mark advanced.
func TestGuardianNightConsolidates(t *testing.T) {
	mt, orch, inj, _ := newTestGuardian(t, `{"gist":"A hard night; Ash nearly slipped away.","promote":["m1"],"fade":["m2"]}`)
	optIn(t, mt, inj, 600)
	seedGuardianMemory(t, mt, inj, 100, 11, "Ash nearly starved at dusk", 7)
	seedGuardianMemory(t, mt, inj, 200, 12, "routine dawn, nothing stirred", 3)

	mt.replica.Tick = 90000 // night 2
	mt.maybeConsolidateNight(store.Event{Tick: 90000, Type: "sim.night_started", Payload: []byte(`{}`)})
	var job guardianConsolJob
	select {
	case job = <-mt.consolQ:
	default:
		t.Fatal("agentized night queued no consolidation")
	}
	if job.night != sim.NightIndex(90000) || len(job.buffer) != 2 {
		t.Fatalf("job = night %d, buffer %d", job.night, len(job.buffer))
	}
	mt.runConsolidation(job)

	// The one model call rode the SHARED consolidation kind (route + class).
	reqs := orch.requests()
	if len(reqs) != 1 || reqs[0].Kind != llm.KindConsolidation {
		t.Fatalf("consolidation calls: %+v", reqs)
	}
	if !strings.Contains(reqs[0].Prompt, "[m1]") || !strings.Contains(reqs[0].Prompt, "[m2]") {
		t.Fatalf("prompt missing m-labels:\n%s", reqs[0].Prompt)
	}

	// State after the night: promoted boosted, faded gone, gist added,
	// high-water mark advanced (buffer closed).
	st := inj.state
	if st.GuardianMemUpTo != 200 {
		t.Fatalf("GuardianMemUpTo = %d, want 200", st.GuardianMemUpTo)
	}
	var sawGist, sawPromoted bool
	for _, m := range st.GuardianMemories {
		if strings.Contains(m.Text, "hard night") {
			sawGist = true
		}
		if m.Text == "Ash nearly starved at dusk" {
			sawPromoted = true
			if m.Salience != 10 {
				t.Fatalf("promoted salience %d, want 10", m.Salience)
			}
		}
		if m.Text == "routine dawn, nothing stirred" {
			t.Fatal("faded memory survived")
		}
	}
	if !sawGist || !sawPromoted {
		t.Fatalf("night outcomes missing (gist %v, promoted %v): %+v", sawGist, sawPromoted, st.GuardianMemories)
	}
	// (In production the door re-stamps the gist's envelope tick to the
	// landing tick, so it opens the NEXT night's buffer; this fixture's
	// injector applies verbatim with Tick 0, so the buffer simply closes.)
	for _, m := range st.GuardianEpisodicBuffer() {
		if m.Text == "Ash nearly starved at dusk" || m.Text == "routine dawn, nothing stirred" {
			t.Fatalf("consolidated memory still in the buffer: %+v", m)
		}
	}

	// One night per index: the same boundary again queues nothing.
	mt.maybeConsolidateNight(store.Event{Tick: 90000, Type: "sim.night_started", Payload: []byte(`{}`)})
	select {
	case <-mt.consolQ:
		t.Fatal("second boundary of the same night queued again")
	default:
	}
}

// TestGuardianNightNonAgentized pins FR-007: the nightly boundary does
// nothing on a non-opted world — no queue, no call, no events.
func TestGuardianNightNonAgentized(t *testing.T) {
	mt, orch, inj, _ := newTestGuardian(t, "never")
	mt.replica.Tick = 90000
	mt.maybeConsolidateNight(store.Event{Tick: 90000, Type: "sim.night_started", Payload: []byte(`{}`)})
	select {
	case <-mt.consolQ:
		t.Fatal("non-agentized night queued a consolidation")
	default:
	}
	if len(orch.requests()) != 0 || len(inj.batches) != 0 {
		t.Fatal("non-agentized night reached the model or the door")
	}
}

// TestGuardianNightRejectedMarker: an unparseable reply lands a rejected
// marker and leaves the buffer intact for the next night.
func TestGuardianNightRejectedMarker(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "not JSON at all")
	optIn(t, mt, inj, 600)
	seedGuardianMemory(t, mt, inj, 100, 11, "a moment", 5)
	mt.replica.Tick = 90000
	mt.maybeConsolidateNight(store.Event{Tick: 90000, Type: "sim.night_started", Payload: []byte(`{}`)})
	mt.runConsolidation(<-mt.consolQ)
	if inj.state.GuardianMemUpTo != 0 {
		t.Fatalf("rejected night advanced the mark to %d", inj.state.GuardianMemUpTo)
	}
	if buf := inj.state.GuardianEpisodicBuffer(); len(buf) != 1 {
		t.Fatalf("rejected night touched the buffer: %+v", buf)
	}
}

// TestAgentizedDigestBecomesMemory pins D5's core sentence: on an agentized
// world the 6-hour digest lands as structured memories, NOT a soul append —
// soul.md stays the persona seed.
func TestAgentizedDigestBecomesMemory(t *testing.T) {
	mt, _, inj, _ := newTestGuardian(t, "- Ash and Birch quarreled over the chest\n- the fire ran low twice")
	optIn(t, mt, inj, 600)
	soulBefore, _ := readFileString(mt.soulPath())
	mt.runDigest(digJob{label: "day 1, 12:00", lines: []string{"[day 1, 06:10] Ash built a fire."}})
	soulAfter, _ := readFileString(mt.soulPath())
	if soulBefore != soulAfter {
		t.Fatal("agentized digest still appended to soul.md")
	}
	if got := len(inj.state.GuardianMemories); got != 2 {
		t.Fatalf("digest landed %d memories, want 2", got)
	}
	for _, m := range inj.state.GuardianMemories {
		if m.Salience != salGuardianDigest {
			t.Fatalf("digest memory salience %d, want %d", m.Salience, salGuardianDigest)
		}
	}
}

// TestMemoryWindowRidesThePrompt pins the D1 context gain: an agentized
// guardian's turn prompt carries its remembered moments (the shared top-K
// window); a non-agentized turn prompt is byte-free of the block.
func TestMemoryWindowRidesThePrompt(t *testing.T) {
	mt, orch, inj, _ := newTestGuardian(t, "I remember.")
	if _, err := mt.Turn(t.Context(), "hello"); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(orch.requests()[0].Prompt, "Your remembered moments") {
		t.Fatal("non-agentized prompt grew the memory block")
	}
	optIn(t, mt, inj, 600)
	seedGuardianMemory(t, mt, inj, 100, 11, "Ash nearly starved at dusk", 7)
	if _, err := mt.Turn(t.Context(), "how fares Ash?"); err != nil {
		t.Fatal(err)
	}
	reqs := orch.requests()
	last := reqs[len(reqs)-1].Prompt
	if !strings.Contains(last, "Your remembered moments") || !strings.Contains(last, "Ash nearly starved at dusk") {
		t.Fatalf("agentized prompt missing the memory window:\n%s", last)
	}
}

func readFileString(path string) (string, error) {
	b, err := os.ReadFile(path)
	return string(b), err
}
