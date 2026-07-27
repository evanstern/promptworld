package mind

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// stubEmbedClient is a settable-failure EmbedClient: one deterministic vector
// per text while healthy, a transport error while failing. Warm calls are
// counted and fail independently (the pin is best-effort by contract).
type stubEmbedClient struct {
	mu        sync.Mutex
	fail      bool
	warmErr   error
	warmCalls atomic.Int64
	// calls counts TEXTS attempted (success or failure; the coalescing worker
	// may merge jobs into one call) — tests sequence on it so "recover after
	// N failures" is deterministic, never a timing bet.
	calls atomic.Int64
}

func (c *stubEmbedClient) setFail(fail bool) {
	c.mu.Lock()
	c.fail = fail
	c.mu.Unlock()
}

func (c *stubEmbedClient) Embed(_ context.Context, texts []string) ([][]float32, string, error) {
	defer c.calls.Add(int64(len(texts)))
	c.mu.Lock()
	fail := c.fail
	c.mu.Unlock()
	if fail {
		return nil, "", errors.New("stub endpoint down")
	}
	vecs := make([][]float32, len(texts))
	for i, t := range texts {
		vecs[i] = []float32{float32(len(t))}
	}
	return vecs, "stub-model", nil
}

func (c *stubEmbedClient) WarmEmbedding(context.Context) error {
	c.warmCalls.Add(1)
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.warmErr
}

// recordingInjector hands each injected batch to the test — the
// synchronization point for the driver's async worker.
type recordingInjector struct{ ch chan []store.Event }

func (r *recordingInjector) InjectSocial(events []store.Event) error {
	r.ch <- events
	return nil
}

// eventStream flattens injected batches into an ordered event cursor: the
// worker COALESCES queued jobs into one injection (a deliberate cost choice —
// see embedJob), so batch boundaries are load-dependent while event ORDER is
// the guarantee under test.
type eventStream struct {
	ch  chan []store.Event
	buf []store.Event
}

func (s *eventStream) next(t *testing.T) store.Event {
	t.Helper()
	if len(s.buf) == 0 {
		select {
		case batch := <-s.ch:
			s.buf = batch
		case <-time.After(5 * time.Second):
			t.Fatal("no injected event within 5s")
		}
	}
	e := s.buf[0]
	s.buf = s.buf[1:]
	return e
}

// idle reports whether nothing is buffered or immediately pending.
func (s *eventStream) idle() bool {
	if len(s.buf) > 0 {
		return false
	}
	select {
	case batch := <-s.ch:
		s.buf = batch
		return len(s.buf) == 0
	default:
		return true
	}
}

// memAdded builds one committed agent.memory_added event as the notify path
// delivers it: real store seq, payload carrying agent + text.
func memAdded(seq int64, agent int, text string) store.Event {
	b, _ := json.Marshal(sim.MemoryAddedPayload{Agent: sim.Ref(agent), Text: text, Salience: 3, Subject: sim.Ref(-1)})
	return store.Event{Seq: seq, Tick: 100, Type: "agent.memory_added", Payload: b}
}

// newTestEmbedder starts a driver over a fresh genesis replica (tick 0 —
// bucket edges fire only when a test advances the tick past a cadence
// boundary on purpose).
func newTestEmbedder(t *testing.T, client EmbedClient, inj SocialInjector, warn func(string)) *Embedder {
	t.Helper()
	m := worldmap.Generate(42, 64, 64)
	state := sim.NewState(42, m)
	e, err := NewEmbedder(client, inj, warn, m, 42, state.Marshal())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(e.Close)
	return e
}

// nextCompanion reads the next injected event off the flattened stream and
// asserts it is a memory companion.
func nextCompanion(t *testing.T, s *eventStream) sim.MemoryEmbeddedPayload {
	t.Helper()
	e := s.next(t)
	if e.Type != "agent.memory_embedded" {
		t.Fatalf("unexpected injected event: %s %s", e.Type, e.Payload)
	}
	var p sim.MemoryEmbeddedPayload
	if err := json.Unmarshal(e.Payload, &p); err != nil {
		t.Fatal(err)
	}
	return p
}

// nextSituation reads the next injected event and asserts it is a situation
// companion.
func nextSituation(t *testing.T, s *eventStream) sim.SituationEmbeddedPayload {
	t.Helper()
	e := s.next(t)
	if e.Type != "agent.situation_embedded" {
		t.Fatalf("unexpected injected event: %s %s", e.Type, e.Payload)
	}
	var p sim.SituationEmbeddedPayload
	if err := json.Unmarshal(e.Payload, &p); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestEmbedderEmissionOrdered (spec 042 T010, contract §2): companions land in
// the committed log's emission order — per agent AND globally, since the
// worker is single-flight FIFO — each targeting its memory's store seq and
// carrying the model identity verbatim.
func TestEmbedderEmissionOrdered(t *testing.T) {
	client := &stubEmbedClient{}
	inj := &recordingInjector{ch: make(chan []store.Event, 16)}
	stream := &eventStream{ch: inj.ch}
	e := newTestEmbedder(t, client, inj, nil)

	e.Observe([]store.Event{
		memAdded(10, 0, "Foraged by the river."),
		memAdded(11, 0, "Saw Birch there."),
		memAdded(12, 1, "Chopped wood."),
		// Non-memory events in the batch are ignored, including our own
		// companions echoing back through notify.
		{Seq: 13, Tick: 100, Type: "agent.moved", Payload: json.RawMessage(`{"agent":0,"x":1,"y":1}`)},
	})
	e.Observe([]store.Event{memAdded(14, 0, "Ate a meal.")})

	wantSeqs := []struct {
		agent int
		seq   int64
	}{{0, 10}, {0, 11}, {1, 12}, {0, 14}}
	for _, want := range wantSeqs {
		p := nextCompanion(t, stream)
		if p.Agent.ID != want.agent || p.MemSeq != want.seq {
			t.Fatalf("companion = agent %d seq %d, want agent %d seq %d (emission order)", p.Agent.ID, p.MemSeq, want.agent, want.seq)
		}
		if p.Model != "stub-model" || len(p.Vec) == 0 {
			t.Errorf("companion missing vector/model: %+v", p)
		}
	}
}

// TestEmbedderFailureDebounce (spec 042 T010, research D8): a failure EPISODE
// warns exactly once no matter how many memories it swallows; a success closes
// the episode and the next episode warns once again.
func TestEmbedderFailureDebounce(t *testing.T) {
	client := &stubEmbedClient{}
	client.setFail(true)
	inj := &recordingInjector{ch: make(chan []store.Event, 16)}
	stream := &eventStream{ch: inj.ch}
	warns := make(chan string, 16)
	e := newTestEmbedder(t, client, inj, func(detail string) { warns <- detail })

	e.Observe([]store.Event{
		memAdded(10, 0, "one"),
		memAdded(11, 0, "two"),
		memAdded(12, 1, "three"),
	})

	// First failure of the episode warns; the other two stay quiet.
	select {
	case <-warns:
	case <-time.After(5 * time.Second):
		t.Fatal("no warning within 5s of the first failure")
	}
	// Sequence, don't race: recover only after the worker has ATTEMPTED all
	// three queued memory texts (possibly coalesced into fewer calls) —
	// otherwise a still-queued one would legitimately succeed post-recovery
	// and read as a forbidden retry.
	deadline := time.Now().Add(5 * time.Second)
	for client.calls.Load() < 3 {
		if time.Now().After(deadline) {
			t.Fatalf("worker attempted %d embed texts within 5s, want 3", client.calls.Load())
		}
		time.Sleep(5 * time.Millisecond)
	}

	// Recover: the next memory embeds — and because the worker is FIFO, its
	// companion arriving proves the three failed ones were already skipped.
	client.setFail(false)
	e.Observe([]store.Event{memAdded(13, 0, "four")})
	if p := nextCompanion(t, stream); p.MemSeq != 13 {
		t.Fatalf("post-recovery companion targets seq %d, want 13 — a failed memory was retried (backfill is forbidden)", p.MemSeq)
	}
	if !stream.idle() {
		t.Fatal("extra companions after the recovery one — skipped memories must stay vectorless forever")
	}
	if len(warns) != 0 {
		t.Errorf("%d extra warnings buffered for one failure episode, want 0", len(warns))
	}

	// A NEW episode warns once again (the success re-armed the debounce).
	client.setFail(true)
	e.Observe([]store.Event{memAdded(14, 0, "five"), memAdded(15, 0, "six")})
	select {
	case <-warns:
	case <-time.After(5 * time.Second):
		t.Fatal("second failure episode never warned")
	}
	if len(warns) != 0 {
		t.Errorf("second episode warned more than once")
	}
}

// TestEmbedderTextTruncationCap (spec 042 T010, FR-011): text beyond the
// fixed-byte cap is cut at exactly that boundary before the model call — the
// deterministic preparation rule.
func TestEmbedderTextTruncationCap(t *testing.T) {
	var gotLen atomic.Int64
	client := &lenRecordingClient{gotLen: &gotLen}
	inj := &recordingInjector{ch: make(chan []store.Event, 4)}
	stream := &eventStream{ch: inj.ch}
	e := newTestEmbedder(t, client, inj, nil)

	long := make([]byte, embedTextCapBytes+512)
	for i := range long {
		long[i] = 'a'
	}
	e.Observe([]store.Event{memAdded(10, 0, string(long))})
	nextCompanion(t, stream)
	if got := gotLen.Load(); got != embedTextCapBytes {
		t.Errorf("embedded text length = %d bytes, want the fixed cap %d", got, embedTextCapBytes)
	}
}

// lenRecordingClient records the byte length of the text it was asked to embed.
type lenRecordingClient struct{ gotLen *atomic.Int64 }

func (c *lenRecordingClient) Embed(_ context.Context, texts []string) ([][]float32, string, error) {
	c.gotLen.Store(int64(len(texts[0])))
	return [][]float32{{1}}, "stub-model", nil
}
func (c *lenRecordingClient) WarmEmbedding(context.Context) error { return nil }

// TestEmbedderWarmPin (spec 042 T008 warm-pin, T010): the driver warms the
// model once at start, and a FAILING warm never disables it — memories still
// embed and no failure-episode warning fires (the pin is best-effort; only
// cold-load latency depends on it).
func TestEmbedderWarmPin(t *testing.T) {
	client := &stubEmbedClient{warmErr: fmt.Errorf("404 page not found")}
	inj := &recordingInjector{ch: make(chan []store.Event, 4)}
	stream := &eventStream{ch: inj.ch}
	warns := make(chan string, 4)
	e := newTestEmbedder(t, client, inj, func(detail string) { warns <- detail })

	e.Observe([]store.Event{memAdded(10, 0, "still embeds")})
	if p := nextCompanion(t, stream); p.MemSeq != 10 {
		t.Fatalf("companion seq = %d, want 10", p.MemSeq)
	}
	if got := client.warmCalls.Load(); got != 1 {
		t.Errorf("warm calls at start = %d, want exactly 1 (re-warms ride the hourly ticker)", got)
	}
	if len(warns) != 0 {
		t.Errorf("a failed warm pin fired the failure-episode warning — it must stay best-effort")
	}
}

// TestEmbedderSituationCadence (spec 042 T012, research D5): when the
// replica's tick crosses a PlannerCadenceTicks bucket edge, the driver renders
// each live agent's deterministic situation string from replica state, embeds
// it, and injects agent.situation_embedded — text (the audit surface), render
// tick, vector, and model all recorded. Memories in the same batch keep their
// place ahead of the bucket's situation renders (one FIFO worker), and the
// same bucket never fires twice.
func TestEmbedderSituationCadence(t *testing.T) {
	client := &stubEmbedClient{}
	inj := &recordingInjector{ch: make(chan []store.Event, 64)}
	stream := &eventStream{ch: inj.ch}
	e := newTestEmbedder(t, client, inj, nil)

	// Reference replica: absorb exactly what the driver observes, so the
	// expected situation strings are the same deterministic render.
	m := worldmap.Generate(42, 64, 64)
	ref := sim.NewState(42, m)
	moved, _ := json.Marshal(sim.AgentMovedPayload{Agent: sim.Ref(0), X: 3, Y: 4})
	batch := []store.Event{
		memAdded(4, 1, "Crossed the meadow."),
		{Seq: 5, Tick: ref.PlannerCadence() + 1, Type: "agent.moved", Payload: moved},
	}
	for _, ev := range batch {
		ref.Apply(ev)
		if ev.Tick > ref.Tick {
			ref.Tick = ev.Tick
		}
	}

	e.Observe(batch)

	// The batch's memory companion lands first (FIFO), then the bucket's
	// situations, one per live agent in index order (batch boundaries are the
	// coalescing worker's business; order and content are the contract).
	if p := nextCompanion(t, stream); p.Agent.ID != 1 || p.MemSeq != 4 {
		t.Fatalf("memory companion = %+v, want agent 1 seq 4 ahead of the bucket's situations", p)
	}
	for i := 0; i < sim.AgentCount; i++ {
		p := nextSituation(t, stream)
		if p.Agent.ID != i {
			t.Fatalf("situation companion order: got agent %d at position %d", p.Agent.ID, i)
		}
		if p.Tick != ref.Tick {
			t.Errorf("agent %d situation tick = %d, want the render tick %d", i, p.Tick, ref.Tick)
		}
		if want := renderSituation(ref, i); p.Text != want {
			t.Errorf("agent %d situation text = %q, want the deterministic render %q", i, p.Text, want)
		}
		if p.Model != "stub-model" || len(p.Vec) == 0 {
			t.Errorf("agent %d situation companion missing vector/model: %+v", i, p)
		}
	}

	// A later event INSIDE the same bucket fires nothing more.
	inside, _ := json.Marshal(sim.AgentMovedPayload{Agent: sim.Ref(0), X: 4, Y: 4})
	e.Observe([]store.Event{{Seq: 6, Tick: ref.PlannerCadence() + 500, Type: "agent.moved", Payload: inside}})
	e.Observe([]store.Event{memAdded(7, 2, "sync marker")})
	if p := nextCompanion(t, stream); p.MemSeq != 7 {
		t.Fatalf("expected only the sync-marker companion, got %+v (same bucket must not re-fire)", p)
	}
	if !stream.idle() {
		t.Fatal("same-bucket event re-fired the situation leg")
	}
}

// TestRenderSituationDeterministic (spec 042 T012): the template is a pure
// render — identical state, identical bytes — and carries every D5 component:
// time phase, position + place, the two worst needs, the active intent
// verb + reason, and nearby names.
func TestRenderSituationDeterministic(t *testing.T) {
	m := worldmap.Generate(42, 64, 64)
	s := sim.NewState(42, m)
	a := &s.Agents[0]
	a.X, a.Y = 3, 4
	a.Needs.Warmth = 200
	a.Needs.Food = 350
	a.Intent = &sim.Intent{Goal: "forage", Reason: "the stores are low."}
	// Put agent 1 adjacent so the nearby clause renders.
	s.Agents[1].X, s.Agents[1].Y = 4, 4

	got := renderSituation(s, 0)
	if got != renderSituation(s, 0) {
		t.Fatal("renderSituation is not deterministic over identical state")
	}
	for _, want := range []string{"(3,4)", "worst needs: warmth 20, food 35", "doing: forage — the stores are low.", s.Agents[1].Name} {
		if !strings.Contains(got, want) {
			t.Errorf("situation %q missing component %q", got, want)
		}
	}
	// Idle agents render the idle clause, never a fabricated intent.
	a.Intent = nil
	if got := renderSituation(s, 0); !strings.Contains(got, "idle") {
		t.Errorf("idle agent situation %q missing the idle clause", got)
	}
}
