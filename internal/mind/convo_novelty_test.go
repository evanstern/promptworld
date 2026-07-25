package mind

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

// TASK-109 / spec 061 Phase 4 (US3): the mind-side novelty SHIM at scene
// founding. These pin SC-003 (no-new-memory founding refused; new-memory
// founding admitted WITH the last gist in the prompt) and SC-005 (the
// SHIM(TASK-109) marker is greppable in the source).

// capturingModel is a scripted model that also records every request prompt, so
// a test can prove the last-conversation gist reached the scene prompt.
type capturingModel struct {
	mu      sync.Mutex
	replies []string
	calls   int
	prompts []string
}

func (m *capturingModel) Submit(_ context.Context, req llm.Request) (llm.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.prompts = append(m.prompts, req.Prompt)
	if m.calls >= len(m.replies) {
		return llm.Response{}, context.DeadlineExceeded
	}
	r := m.replies[m.calls]
	m.calls++
	return llm.Response{Text: r, Tier: llm.TierLocal}, nil
}

// noveltyMind builds a convo-ready Mind over a caller-shaped state (the
// setupConvo shape, but with the pair already adjacent so the founding
// decision is the novelty gate, not geometry).
func noveltyMind(t *testing.T, model Submitter, mutate func(*sim.State)) (*harness, *Mind) {
	t.Helper()
	h := newHarness(t, "")
	m := h.m
	state := sim.NewState(42, m)
	state.Agents[0].X, state.Agents[0].Y = 10, 10
	state.Agents[1].X, state.Agents[1].Y = 10, 11
	if mutate != nil {
		mutate(state)
	}
	md, err := New(model, h.loop, h.loop, m, 42, state.Marshal(), [sim.AgentCount]string{},
		testLoopRounds, testPlannerTokens, testConsolidationTokens, "", noopLoop)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(md.Close)
	return h, md
}

// TestNoveltyShimRefusesWhenNothingNew is SC-003 / US3-AS1: past the pair
// cooldown but with no above-floor memory on either side since the last
// exchange, scene founding is refused with a queryable "nothing new" outcome —
// and no scene lands.
func TestNoveltyShimRefusesWhenNothingNew(t *testing.T) {
	model := &scriptedModel{replies: convoScript(
		`{"gist":"x","topics":["t"],"tones":[1,1],"retold":null}`)}
	h, md := noveltyMind(t, model, func(s *sim.State) {
		// A salient memory, but formed BEFORE the pair's prior exchange (tick
		// 1000): stale, so it is NOT novelty.
		s.Agents[0].Memories = append(s.Agents[0].Memories,
			sim.Memory{Text: "Built the fire.", Salience: 6, Tick: 500, Subject: -1})
	})

	md.maybeStartConversation(store.Event{Tick: 5000, Type: "agent.talked",
		Payload: mustJSON(t, sim.TalkedPayload{A: 0, B: 1})}, 1000)

	outs := h.waitEvents(t, 5*time.Second, func(e store.Event) bool {
		var p sim.CogOutcomePayload
		return e.Type == "cog.outcome" && json.Unmarshal(e.Payload, &p) == nil &&
			p.Class == "conversation" && p.Outcome == sim.OutcomeSuppressed &&
			strings.Contains(p.Reason, "nothing new")
	})
	if len(outs) == 0 {
		t.Fatal("novelty SHIM did not emit a 'nothing new' suppressed outcome")
	}
	all, _ := h.st.EventsSince(0, 0)
	for _, e := range all {
		if strings.HasPrefix(e.Type, "social.conversation") {
			t.Fatalf("a no-novelty encounter founded a scene: %s", e.Type)
		}
	}
}

// TestNoveltyShimAdmitsWithGistContext is SC-003 / US3-AS2: with a fresh
// above-floor memory since the last exchange the scene founds, and the pair's
// previous-exchange gist rides into the scene prompt as "already talked about"
// context (the last-gist SHIM).
func TestNoveltyShimAdmitsWithGistContext(t *testing.T) {
	model := &capturingModel{replies: convoScript(
		`{"gist":"found it","topics":["goat"],"tones":[1,1],"retold":null}`)}
	h, md := noveltyMind(t, model, func(s *sim.State) {
		// A salient memory formed AFTER the prior exchange (tick 1000): genuine
		// novelty for agent 0.
		s.Agents[0].Memories = append(s.Agents[0].Memories,
			sim.Memory{Text: "Raised a shelter.", Salience: 6, Tick: 1500, Subject: -1})
		// A prior conversation record between the pair — its gist is the scene's
		// "already talked about" context (convo_record machinery).
		s.Conversations = []sim.ConvoRecord{{Conv: 1000, Tick: 1000,
			Participants: []int{0, 1}, Gist: "the missing goat"}}
	})

	md.maybeStartConversation(store.Event{Tick: 5000, Type: "agent.talked",
		Payload: mustJSON(t, sim.TalkedPayload{A: 0, B: 1})}, 1000)

	convs := h.waitEvents(t, 10*time.Second, func(e store.Event) bool {
		return e.Type == "social.conversation"
	})
	if len(convs) == 0 {
		t.Fatal("a novel encounter did not found a scene")
	}
	model.mu.Lock()
	prompts := append([]string{}, model.prompts...)
	model.mu.Unlock()
	var sawGist bool
	for _, pr := range prompts {
		if strings.Contains(pr, "the missing goat") {
			sawGist = true
			break
		}
	}
	if !sawGist {
		t.Error("the pair's last gist did not enter the scene prompt (last-gist SHIM context)")
	}
}

// TestNoveltyShimFirstContactFounds: a never-talked pair (priorExchange 0) is
// vacuously novel — first contact always founds, no salient memory required.
func TestNoveltyShimFirstContactFounds(t *testing.T) {
	model := &scriptedModel{replies: convoScript(
		`{"gist":"hello","topics":["greeting"],"tones":[1,1],"retold":null}`)}
	h, md := noveltyMind(t, model, nil)

	md.maybeStartConversation(store.Event{Tick: 5000, Type: "agent.talked",
		Payload: mustJSON(t, sim.TalkedPayload{A: 0, B: 1})}, 0)

	if convs := h.waitEvents(t, 10*time.Second, func(e store.Event) bool {
		return e.Type == "social.conversation"
	}); len(convs) == 0 {
		t.Fatal("first-contact (priorExchange 0) scene did not found")
	}
}
