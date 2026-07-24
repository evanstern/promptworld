package daemon

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/mind"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/world"
)

// throughputTicks is one full game day — SC-005's unit of measure.
const throughputTicks = 24 * 3600

// runGameDay drives a fresh seeded world through one game day at max speed
// and returns the wall time, optionally with the embedder wired to a stubbed
// endpoint — the SC-005 measurement primitive. Everything else (seed, map,
// store, loop) is identical between the two arms.
func runGameDay(t *testing.T, withEmbedder bool) (time.Duration, int64) {
	t.Helper()
	dir := t.TempDir() + "/w"
	w, err := world.Create(dir, "bench", 42)
	if err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(filepath.Join(dir, "world.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	var embedCalls atomic.Int64
	state := sim.NewState(w.Manifest.Seed, w.Map())
	state.Speed = "max" // pre-Run configuration; this arm never replays
	var emb *mind.Embedder
	notify := func(evs []store.Event) {
		if emb != nil {
			emb.Observe(evs)
		}
	}
	loop := sim.NewLoop(state, w.Map(), st, notify)

	if withEmbedder {
		srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			embedCalls.Add(1)
			var req struct {
				Input []string `json:"input"`
			}
			json.NewDecoder(r.Body).Decode(&req)
			data := make([]map[string]any, len(req.Input))
			for i, text := range req.Input {
				data[i] = map[string]any{"index": i, "embedding": []float32{float32(len(text)), 0.5}}
			}
			rw.Header().Set("Content-Type", "application/json")
			json.NewEncoder(rw).Encode(map[string]any{"data": data})
		}))
		defer srv.Close()
		routes := map[string]llm.RouteConfig{string(llm.KindEmbedding): {Chain: []string{"embedder"}}}
		for _, k := range llm.Kinds() {
			routes[string(k)] = llm.RouteConfig{Chain: []string{"chat"}}
		}
		orch, err := llm.New(llm.Config{
			MonthlyBudgetUSD: 100,
			Providers: map[string]llm.ProviderConfig{
				"chat":     {Transport: llm.ProviderOpenAICompat, Endpoint: "http://127.0.0.1:1/v1", Model: "never-called"},
				"embedder": {Transport: llm.ProviderOpenAICompat, Endpoint: srv.URL, Model: "all-minilm"},
			},
			Routes: routes,
		}, st)
		if err != nil {
			t.Fatal(err)
		}
		defer orch.Close()
		emb, err = mind.NewEmbedder(orch, loop, nil, w.Map(), w.Manifest.Seed, state.Marshal())
		if err != nil {
			t.Fatal(err)
		}
		defer emb.Close()
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	start := time.Now()
	go func() { done <- loop.Run(ctx) }()
	deadline := time.Now().Add(5 * time.Minute)
	for {
		s, err := loop.Do("status", "")
		if err != nil {
			t.Fatal(err)
		}
		if s.Tick >= throughputTicks {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("game day not reached in 5m (tick %d)", s.Tick)
		}
		time.Sleep(50 * time.Millisecond)
	}
	elapsed := time.Since(start)
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("loop.Run: %v", err)
	}
	return elapsed, embedCalls.Load()
}

// TestEmbeddingThroughputSC005 (spec 042 T024, SC-005): wall-clock for one
// game day at max speed, embedding off vs on, same seed. Embeds are fully off
// the tick path (async driver, stubbed endpoint), so the expected delta is
// ≈0; SC-005's budget is 10%. The hard assertion is a generous 25% tripwire —
// wall-clock noise on a shared machine must not flake the suite — while the
// LOGGED numbers are the recorded SC-005 evidence (report/board artifact).
func TestEmbeddingThroughputSC005(t *testing.T) {
	if testing.Short() {
		t.Skip("SC-005 wall-clock measurement skipped in -short mode")
	}
	off, _ := runGameDay(t, false)
	on, calls := runGameDay(t, true)
	delta := (on.Seconds() - off.Seconds()) / off.Seconds() * 100
	t.Logf("SC-005: one game day (%d ticks) at max speed — embedding off %.2fs, on %.2fs (%.1f%% delta, %d embed calls)",
		throughputTicks, off.Seconds(), on.Seconds(), delta, calls)
	if delta > 25 {
		t.Errorf("embedding-on game day %.2fs vs off %.2fs — %.1f%% over the 25%% tripwire (SC-005 budget is 10%%)",
			on.Seconds(), off.Seconds(), delta)
	}
}
