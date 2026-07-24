package mind

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
)

// stubEmbedClient is a settable-failure EmbedClient: one deterministic vector
// per text while healthy, a transport error while failing. Warm calls are
// counted and fail independently (the pin is best-effort by contract).
type stubEmbedClient struct {
	mu        sync.Mutex
	fail      bool
	warmErr   error
	warmCalls atomic.Int64
}

func (c *stubEmbedClient) setFail(fail bool) {
	c.mu.Lock()
	c.fail = fail
	c.mu.Unlock()
}

func (c *stubEmbedClient) Embed(_ context.Context, texts []string) ([][]float32, string, error) {
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

// memAdded builds one committed agent.memory_added event as the notify path
// delivers it: real store seq, payload carrying agent + text.
func memAdded(seq int64, agent int, text string) store.Event {
	b, _ := json.Marshal(sim.MemoryAddedPayload{Agent: agent, Text: text, Salience: 3, Subject: -1})
	return store.Event{Seq: seq, Tick: 100, Type: "agent.memory_added", Payload: b}
}

// nextCompanion reads one injected companion batch or fails the test.
func nextCompanion(t *testing.T, ch chan []store.Event) sim.MemoryEmbeddedPayload {
	t.Helper()
	select {
	case batch := <-ch:
		if len(batch) != 1 || batch[0].Type != "agent.memory_embedded" {
			t.Fatalf("unexpected injected batch: %+v", batch)
		}
		var p sim.MemoryEmbeddedPayload
		if err := json.Unmarshal(batch[0].Payload, &p); err != nil {
			t.Fatal(err)
		}
		return p
	case <-time.After(5 * time.Second):
		t.Fatal("no companion injected within 5s")
		return sim.MemoryEmbeddedPayload{}
	}
}

// TestEmbedderEmissionOrdered (spec 042 T010, contract §2): companions land in
// the committed log's emission order — per agent AND globally, since the
// worker is single-flight FIFO — each targeting its memory's store seq and
// carrying the model identity verbatim.
func TestEmbedderEmissionOrdered(t *testing.T) {
	client := &stubEmbedClient{}
	inj := &recordingInjector{ch: make(chan []store.Event, 16)}
	e := NewEmbedder(client, inj, nil)
	defer e.Close()

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
		p := nextCompanion(t, inj.ch)
		if p.Agent != want.agent || p.MemSeq != want.seq {
			t.Fatalf("companion = agent %d seq %d, want agent %d seq %d (emission order)", p.Agent, p.MemSeq, want.agent, want.seq)
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
	warns := make(chan string, 16)
	e := NewEmbedder(client, inj, func(detail string) { warns <- detail })
	defer e.Close()

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

	// Recover: the next memory embeds — and because the worker is FIFO, its
	// companion arriving proves the three failed ones were already skipped.
	client.setFail(false)
	e.Observe([]store.Event{memAdded(13, 0, "four")})
	if p := nextCompanion(t, inj.ch); p.MemSeq != 13 {
		t.Fatalf("post-recovery companion targets seq %d, want 13 — a failed memory was retried (backfill is forbidden)", p.MemSeq)
	}
	select {
	case batch := <-inj.ch:
		t.Fatalf("extra injection after the recovery companion: %+v (skipped memories must stay vectorless forever)", batch)
	default:
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
	e := NewEmbedder(client, inj, nil)
	defer e.Close()

	long := make([]byte, embedTextCapBytes+512)
	for i := range long {
		long[i] = 'a'
	}
	e.Observe([]store.Event{memAdded(10, 0, string(long))})
	nextCompanion(t, inj.ch)
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
	warns := make(chan string, 4)
	e := NewEmbedder(client, inj, func(detail string) { warns <- detail })
	defer e.Close()

	e.Observe([]store.Event{memAdded(10, 0, "still embeds")})
	if p := nextCompanion(t, inj.ch); p.MemSeq != 10 {
		t.Fatalf("companion seq = %d, want 10", p.MemSeq)
	}
	if got := client.warmCalls.Load(); got != 1 {
		t.Errorf("warm calls at start = %d, want exactly 1 (re-warms ride the hourly ticker)", got)
	}
	if len(warns) != 0 {
		t.Errorf("a failed warm pin fired the failure-episode warning — it must stay best-effort")
	}
}
