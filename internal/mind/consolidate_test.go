package mind

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/persona"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

// setupConsol builds a mind whose replica has agent 0 (Ash) carrying a
// known episodic buffer, driven by the given scripted model.
func setupConsol(t *testing.T, model Submitter) (*harness, *Mind) {
	t.Helper()
	h := newHarness(t, "") // harness's own mock is unused; we build our own mind
	// Pause the world: consolidation is command-driven (inject_social works
	// while paused), and a running executor pollutes the store with its own
	// memories — at max speed the world reaches night 1 (the gru!) within
	// the assertion windows.
	if _, err := h.loop.Do("pause", ""); err != nil {
		t.Fatal(err)
	}
	m := h.m
	state := sim.NewState(42, m)
	state.Agents[0].Memories = []sim.Memory{
		{Text: "Saw a wolf at the treeline.", Salience: 7, Tick: 100, Subject: -1},
		{Text: "Ate two berries.", Salience: 1, Tick: 200, Subject: -1},
		{Text: "Cedar promised me firewood.", Salience: 5, Tick: 300, Subject: 2, Tone: 20},
	}

	md, err := New(model, h.loop, h.loop, m, 42, state.Marshal(), [sim.AgentCount]string{}, testLoopRounds, testPlannerTokens, testConsolidationTokens, "", noopLoop)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(md.Close)
	return h, md
}

func sleptEvent(t *testing.T, tick int64, agent int) store.Event {
	t.Helper()
	b, err := json.Marshal(sim.AgentPayload{Agent: sim.Ref(agent)})
	if err != nil {
		t.Fatal(err)
	}
	return store.Event{Tick: tick, Type: "agent.slept", Payload: b}
}

// goodConsolidation is a valid output for setupConsol's buffer, in Ash's
// nature.
func goodConsolidation() string {
	return fmt.Sprintf(`{
  "nature": "%s",
  "gist": "A day of wolves, thin meals, and a promise from Cedar.",
  "promote": ["m1"],
  "fade": ["m2"],
  "beliefs": [{"id": 0, "statement": "Cedar keeps his word.", "confidence": 55, "provenance": "witnessed", "source": -1, "subject": 2}],
  "narrative": "I keep this village fed and I keep my head. The wolf worries me; the fire does not."
}`, persona.Anchors["Ash"])
}

// TestConsolidationAcceptedLands is US1 AC-1: one sleep → one atomic batch
// (promote, fade, gist, belief, narrative, marker), and a second sleep the
// same night does nothing (AC-2).
func TestConsolidationAcceptedLands(t *testing.T) {
	h, md := setupConsol(t, &scriptedModel{replies: []string{goodConsolidation()}})

	md.maybeConsolidate(sleptEvent(t, 80000, 0))

	markers := h.waitEvents(t, 10*time.Second, func(e store.Event) bool {
		return e.Type == "agent.consolidated"
	})
	if len(markers) != 1 {
		t.Fatalf("markers = %d, want 1", len(markers))
	}
	var mp sim.ConsolidatedPayload
	json.Unmarshal(markers[0].Payload, &mp)
	if mp.Outcome != sim.ConsolidationAccepted || mp.Night != 1 || mp.UpTo != 300 {
		t.Errorf("marker = %+v", mp)
	}

	all, _ := h.st.EventsSince(0, 0)
	counts := map[string]int{}
	for _, e := range all {
		counts[e.Type]++
	}
	for typ, want := range map[string]int{
		"agent.memory_promoted": 1, "agent.memory_faded": 1,
		"agent.belief_revised": 1, "agent.narrative_set": 1,
		"agent.consolidated": 1,
	} {
		if counts[typ] != want {
			t.Errorf("%s = %d, want %d", typ, counts[typ], want)
		}
	}

	// Feed the recorded marker back to the replica (in the daemon the loop
	// notify does this) — a second sleep the same night must not consolidate
	// again.
	md.absorb(markers)
	if md.replica.Agents[0].LastConsolidatedNight != 1 {
		t.Fatal("replica did not absorb the marker")
	}
	md.maybeConsolidate(sleptEvent(t, 81000, 0))
	time.Sleep(300 * time.Millisecond)
	all, _ = h.st.EventsSince(0, 0)
	n := 0
	for _, e := range all {
		if e.Type == "agent.consolidated" {
			n++
		}
	}
	if n != 1 {
		t.Errorf("same-night second sleep consolidated again: %d markers", n)
	}
}

// TestConsolidationTransportFailureDefers is US1 AC-3: tier down → no
// marker, no events, buffer intact for the next sleep.
func TestConsolidationTransportFailureDefers(t *testing.T) {
	h, md := setupConsol(t, &scriptedModel{}) // no replies → every call errors

	md.maybeConsolidate(sleptEvent(t, 80000, 0))
	time.Sleep(500 * time.Millisecond)

	all, _ := h.st.EventsSince(0, 0)
	for _, e := range all {
		if strings.HasPrefix(e.Type, "agent.consolidated") ||
			e.Type == "agent.narrative_set" || e.Type == "agent.belief_revised" {
			t.Fatalf("deferred consolidation leaked %s", e.Type)
		}
	}
	if md.replica.Agents[0].LastConsolidatedNight != 0 {
		t.Error("deferred attempt must not close the night")
	}
	if got := len(md.replica.Agents[0].EpisodicBuffer()); got != 3 {
		t.Errorf("buffer = %d memories, want 3 (intact)", got)
	}
}

// TestConsolidationMalformedRejected: garbage output lands ONLY a rejected
// marker; the buffer survives (retry next night).
func TestConsolidationMalformedRejected(t *testing.T) {
	h, md := setupConsol(t, &scriptedModel{replies: []string{"the model hums a tune with no json"}})

	md.maybeConsolidate(sleptEvent(t, 80000, 0))
	markers := h.waitEvents(t, 10*time.Second, func(e store.Event) bool {
		return e.Type == "agent.consolidated"
	})
	var mp sim.ConsolidatedPayload
	json.Unmarshal(markers[0].Payload, &mp)
	if mp.Outcome != sim.ConsolidationRejected || mp.Reason == "" {
		t.Errorf("marker = %+v, want rejected with reason", mp)
	}
	all, _ := h.st.EventsSince(0, 0)
	for _, e := range all {
		switch e.Type {
		case "agent.memory_promoted", "agent.memory_faded", "agent.belief_revised", "agent.narrative_set", "agent.memory_added":
			t.Fatalf("rejected night leaked %s", e.Type)
		}
	}
}

// TestConsolidationEmptyBufferSkips: nothing to digest → skipped_empty
// marker, zero model calls.
func TestConsolidationEmptyBufferSkips(t *testing.T) {
	model := &scriptedModel{replies: []string{goodConsolidation()}}
	h := newHarness(t, "")
	state := sim.NewState(42, h.m)
	md, err := New(model, h.loop, h.loop, h.m, 42, state.Marshal(), [sim.AgentCount]string{}, testLoopRounds, testPlannerTokens, testConsolidationTokens, "", noopLoop)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(md.Close)

	md.maybeConsolidate(sleptEvent(t, 80000, 1))
	markers := h.waitEvents(t, 10*time.Second, func(e store.Event) bool {
		return e.Type == "agent.consolidated"
	})
	var mp sim.ConsolidatedPayload
	json.Unmarshal(markers[0].Payload, &mp)
	if mp.Outcome != sim.ConsolidationSkippedEmpty {
		t.Errorf("outcome = %q", mp.Outcome)
	}
	model.mu.Lock()
	calls := model.calls
	model.mu.Unlock()
	if calls != 0 {
		t.Errorf("empty night spent %d model calls", calls)
	}
}

// TestPersonaBytesSurviveConsolidation is FR-007's observable half: a full
// accepted cycle leaves every persona.md byte-identical (the consolidator
// has no filesystem access at all — this is the canary).
func TestPersonaBytesSurviveConsolidation(t *testing.T) {
	dir := t.TempDir()
	if err := persona.Genesis(dir); err != nil {
		t.Fatal(err)
	}
	sum := func() [sim.AgentCount][16]byte {
		var out [sim.AgentCount][16]byte
		for i, name := range sim.AgentNames {
			b, err := os.ReadFile(persona.PersonaPath(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			out[i] = md5.Sum(b)
		}
		return out
	}
	before := sum()

	h, md := setupConsol(t, &scriptedModel{replies: []string{goodConsolidation()}})
	md.maybeConsolidate(sleptEvent(t, 80000, 0))
	h.waitEvents(t, 10*time.Second, func(e store.Event) bool {
		return e.Type == "agent.consolidated"
	})

	if before != sum() {
		t.Fatal("persona.md changed across a consolidation cycle")
	}
	// And the files are genesis-locked: not writable.
	if err := os.WriteFile(persona.PersonaPath(dir, "Ash"), []byte("mutiny"), 0o644); err == nil {
		t.Fatal("persona.md was writable (mode should be 0444)")
	}
}

// --- spec 105 (TASK-172): truncation-aware retry ladder regression tests ---

// lateWorldTick is a day-29 evening sleep — the playtest-1 blackout's world
// age (night index 29).
const lateWorldTick = 28*86400 + 78000

// setupLateWorld builds the playtest failure shape (spec 105 FR-010): agent 0
// carries an episodic buffer LARGER than maxBufferSent and 14 held beliefs
// including below-floor faded ones — the prompt volume that overflowed the
// 1024-token response budget in playtest-1.
func setupLateWorld(t *testing.T, model Submitter) (*harness, *Mind) {
	t.Helper()
	h := newHarness(t, "") // harness's own mock is unused; we build our own mind
	if _, err := h.loop.Do("pause", ""); err != nil {
		t.Fatal(err)
	}
	state := sim.NewState(42, h.m)
	// 70 memories since the last consolidation: buffer > maxBufferSent (60),
	// newest 60 ride the prompt.
	for i := 0; i < 70; i++ {
		state.Agents[0].Memories = append(state.Agents[0].Memories, sim.Memory{
			Text:     fmt.Sprintf("Day-29 doing %d: hauled wood, watched the ridge, traded words.", i+1),
			Salience: 1 + i%7, Tick: lateWorldTick - int64(70-i)*600, Subject: -1,
		})
	}
	// 14 held beliefs; the odd-ID half were last reinforced on day 1, so 28
	// days of half-life decay puts their effective confidence below the floor
	// (faded — still listed by ID in the prompt, spec 030).
	state.NextBeliefID = 15
	for i := 1; i <= 14; i++ {
		reinforced := int64(lateWorldTick - 3600) // fresh
		if i%2 == 1 {
			reinforced = 100 // day 1: faded by day 29
		}
		state.Agents[0].Beliefs = append(state.Agents[0].Beliefs, sim.Belief{
			ID: i, Statement: fmt.Sprintf("Long-held notion %d about the village.", i),
			Confidence: 60, Provenance: sim.ProvenanceInferred, Source: -1, Subject: -1,
			Tick: 100, Reinforced: reinforced,
		})
	}
	if faded := len(state.Agents[0].Beliefs) - len(sim.PromptBeliefs(state.Agents[0].Beliefs, lateWorldTick)); faded < 7 {
		t.Fatalf("fixture defect: %d faded beliefs, want 7", faded)
	}
	md, err := New(model, h.loop, h.loop, h.m, 42, state.Marshal(), [sim.AgentCount]string{}, testLoopRounds, testPlannerTokens, testConsolidationTokens, "", noopLoop)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(md.Close)
	return h, md
}

// lateWorldGood is a valid consolidation for setupLateWorld's buffer, in
// Ash's nature — the reply the model "needed more than 1024 tokens" to emit.
func lateWorldGood() string {
	return fmt.Sprintf(`{
  "nature": "%s",
  "gist": "Twenty-nine days in: the village holds, the ridge stays quiet, and I keep the fire fed.",
  "promote": ["m57", "m60"],
  "fade": ["m1", "m2", "m3"],
  "beliefs": [
    {"id": 3, "statement": "The north ridge goes quiet at dusk.", "confidence": 70, "provenance": "inferred", "source": -1, "subject": -1, "evidence": ["m57"]},
    {"id": 0, "statement": "The wood store outlasts a hard week when we haul daily.", "confidence": 45, "provenance": "inferred", "source": -1, "subject": -1, "evidence": ["m58", "m59"]}
  ],
  "narrative": "Twenty-nine days of keeping this village fed and warm. I count the wood, I watch the ridge, I keep my head. The work is long but it is mine."
}`, persona.Anchors["Ash"])
}

// truncCut returns a truncated-JSON llm.Response — the reply cut mid-field at
// the given budget, stop reason max_tokens (the playtest failure sample).
func truncCut(budget int64) llm.Response {
	text := lateWorldGood()
	return llm.Response{Text: text[:len(text)/2], Stop: llm.StopMaxTokens,
		OutputTokens: budget, CostUSD: 0.01, Tier: llm.TierLocal}
}

// TestConsolidationLateWorldRetryAccepted is SC-001, the playtest shape
// reproduced and fixed: a late-world night truncates at 1024 and completes at
// 2048 — accepted marker with Retries 1, both attempts' cost, and a
// cog.outcome{retried} record naming the escalation.
func TestConsolidationLateWorldRetryAccepted(t *testing.T) {
	model := &respModel{replies: []llm.Response{
		truncCut(1024),
		{Text: lateWorldGood(), Stop: llm.StopEndTurn, OutputTokens: 1400, CostUSD: 0.02, Tier: llm.TierLocal},
	}}
	h, md := setupLateWorld(t, model)

	md.maybeConsolidate(sleptEvent(t, lateWorldTick, 0))

	markers := h.waitEvents(t, 10*time.Second, func(e store.Event) bool {
		return e.Type == "agent.consolidated"
	})
	if len(markers) != 1 {
		t.Fatalf("markers = %d, want 1", len(markers))
	}
	var mp sim.ConsolidatedPayload
	json.Unmarshal(markers[0].Payload, &mp)
	if mp.Outcome != sim.ConsolidationAccepted || mp.Night != 29 {
		t.Fatalf("marker = %+v, want accepted night 29", mp)
	}
	if mp.Retries != 1 {
		t.Errorf("marker Retries = %d, want 1", mp.Retries)
	}
	if mp.CostUSD != 0.03 {
		t.Errorf("marker CostUSD = %v, want 0.03 (both attempts accrued)", mp.CostUSD)
	}
	if model.attempts() != 2 {
		t.Errorf("attempts = %d, want 2", model.attempts())
	}
	if want := []int64{1024, 2048}; fmt.Sprint(model.budgets()) != fmt.Sprint(want) {
		t.Errorf("requested budgets = %v, want %v", model.budgets(), want)
	}

	// The consumed retry is visible as a cog.outcome{retried} in the existing
	// vocabulary (FR-004): class consolidation, reason naming both budgets.
	outcomes := h.waitEvents(t, 10*time.Second, func(e store.Event) bool {
		if e.Type != "cog.outcome" {
			return false
		}
		var p sim.CogOutcomePayload
		return json.Unmarshal(e.Payload, &p) == nil && p.Outcome == sim.OutcomeRetried
	})
	if len(outcomes) != 1 {
		t.Fatalf("retried cog.outcome records = %d, want 1", len(outcomes))
	}
	var op sim.CogOutcomePayload
	json.Unmarshal(outcomes[0].Payload, &op)
	if op.Class != "consolidation" || !op.Retried ||
		!strings.Contains(op.Reason, "1024") || !strings.Contains(op.Reason, "2048") {
		t.Errorf("retried record = %+v, want class consolidation naming 1024→2048", op)
	}

	// The digest landed: the night's batch is complete, buffer consumed.
	md.absorb(markers)
	if md.replica.Agents[0].LastConsolidatedNight != 29 {
		t.Error("replica did not absorb the accepted marker")
	}
	if got := len(md.replica.Agents[0].EpisodicBuffer()); got != 0 {
		t.Errorf("buffer after accepted night = %d memories, want 0", got)
	}
}

// TestConsolidationLadderExhaustedTruncated is FR-003: still cut at the 4096
// clamp — the night lands rejected with the DISTINCT reason "truncated"
// (never "unparseable"), the buffer stays intact, and the next night's sleep
// retries from the ladder's start and can accept.
func TestConsolidationLadderExhaustedTruncated(t *testing.T) {
	model := &respModel{replies: []llm.Response{
		truncCut(1024), truncCut(2048), truncCut(4096),
		{Text: lateWorldGood(), Stop: llm.StopEndTurn, OutputTokens: 1400, CostUSD: 0.02, Tier: llm.TierLocal},
	}}
	h, md := setupLateWorld(t, model)

	md.maybeConsolidate(sleptEvent(t, lateWorldTick, 0))
	markers := h.waitEvents(t, 10*time.Second, func(e store.Event) bool {
		return e.Type == "agent.consolidated"
	})
	var mp sim.ConsolidatedPayload
	json.Unmarshal(markers[0].Payload, &mp)
	if mp.Outcome != sim.ConsolidationRejected || mp.Reason != sim.ConsolidationReasonTruncated {
		t.Fatalf("marker = %+v, want rejected %q", mp, sim.ConsolidationReasonTruncated)
	}
	if mp.Retries != 2 || mp.CostUSD != 0.03 {
		t.Errorf("marker retries/cost = %d/$%v, want 2/$0.03", mp.Retries, mp.CostUSD)
	}
	if want := []int64{1024, 2048, 4096}; fmt.Sprint(model.budgets()) != fmt.Sprint(want) {
		t.Errorf("requested budgets = %v, want %v", model.budgets(), want)
	}

	// Buffer intact; the ledger closed the night without consuming it.
	md.absorb(markers)
	if got := len(md.replica.Agents[0].EpisodicBuffer()); got != 70 {
		t.Errorf("buffer after truncated night = %d memories, want 70 (intact)", got)
	}

	// The NEXT night retries from the ladder's start and accepts.
	md.maybeConsolidate(sleptEvent(t, lateWorldTick+86400, 0))
	accepted := h.waitEvents(t, 10*time.Second, func(e store.Event) bool {
		if e.Type != "agent.consolidated" {
			return false
		}
		var p sim.ConsolidatedPayload
		return json.Unmarshal(e.Payload, &p) == nil && p.Outcome == sim.ConsolidationAccepted
	})
	if len(accepted) != 1 {
		t.Fatalf("next-night accepted markers = %d, want 1", len(accepted))
	}
	var ap sim.ConsolidatedPayload
	json.Unmarshal(accepted[0].Payload, &ap)
	if ap.Night != 30 || ap.Retries != 0 {
		t.Errorf("next-night marker = %+v, want night 30 with 0 retries", ap)
	}
	if got := model.budgets()[3]; got != 1024 {
		t.Errorf("next night started at %d tokens, want 1024 (ladder resets)", got)
	}
}

// TestConsolidationFirstAttemptByteIdentical is FR-009's accepted-path half:
// a clean first attempt consumes no retry, its marker carries NO retries key
// (byte-compat with pre-105 markers), and no cog.outcome{retried} is emitted.
func TestConsolidationFirstAttemptByteIdentical(t *testing.T) {
	model := &respModel{replies: []llm.Response{
		// max_tokens stop on a reply that PARSES: the object completed before
		// the cut — content judgment only, never a budget retry (FR-001).
		{Text: goodConsolidation(), Stop: llm.StopMaxTokens, OutputTokens: 1024, CostUSD: 0.01, Tier: llm.TierLocal},
	}}
	h, md := setupConsol(t, model)

	md.maybeConsolidate(sleptEvent(t, 80000, 0))
	markers := h.waitEvents(t, 10*time.Second, func(e store.Event) bool {
		return e.Type == "agent.consolidated"
	})
	var mp sim.ConsolidatedPayload
	json.Unmarshal(markers[0].Payload, &mp)
	if mp.Outcome != sim.ConsolidationAccepted || mp.Retries != 0 {
		t.Fatalf("marker = %+v, want accepted with 0 retries", mp)
	}
	if bytes.Contains(markers[0].Payload, []byte("retries")) {
		t.Errorf("zero-retry marker leaked a retries key: %s", markers[0].Payload)
	}
	if model.attempts() != 1 {
		t.Errorf("attempts = %d, want 1 (parse success never retries)", model.attempts())
	}
	time.Sleep(300 * time.Millisecond) // retried records inject detached; give one a chance to leak
	all, _ := h.st.EventsSince(0, 0)
	for _, e := range all {
		if e.Type != "cog.outcome" {
			continue
		}
		var p sim.CogOutcomePayload
		if json.Unmarshal(e.Payload, &p) == nil && p.Outcome == sim.OutcomeRetried {
			t.Errorf("clean first attempt emitted a retried record: %+v", p)
		}
	}
}

// TestConsolidationTransportFailureMidLadder is FR-009's transport half under
// the ladder: attempt 1 truncates, attempt 2 dies in transport — NO marker
// lands (the night never happened as far as the ledger cares), buffer intact.
func TestConsolidationTransportFailureMidLadder(t *testing.T) {
	model := &respModel{replies: []llm.Response{truncCut(1024)}} // attempt 2 → error
	h, md := setupLateWorld(t, model)

	md.maybeConsolidate(sleptEvent(t, lateWorldTick, 0))
	time.Sleep(500 * time.Millisecond)

	all, _ := h.st.EventsSince(0, 0)
	for _, e := range all {
		if e.Type == "agent.consolidated" {
			t.Fatalf("mid-ladder transport failure landed a marker: %s", e.Payload)
		}
	}
	if md.replica.Agents[0].LastConsolidatedNight != 0 {
		t.Error("deferred attempt must not close the night")
	}
	if got := len(md.replica.Agents[0].EpisodicBuffer()); got != 70 {
		t.Errorf("buffer = %d memories, want 70 (intact)", got)
	}
}
