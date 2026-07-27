package daemon

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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

// TestEmbeddingReplayByteIdentical (spec 042 T011, SC-001/SC-002) extends the
// replay harness (TestTeachingPostureReplayByteIdentical pattern): a seeded
// world runs with the embedder wired to a stubbed /embeddings endpoint — every
// emitted memory converges to a recorded vector (SC-002) — then the log
// replays through the daemon's own recovery path with the embedder NOT wired:
// the replayed state is byte-identical and the embedding endpoint receives
// ZERO calls during replay (SC-001; replay performs no embedding computation).
func TestEmbeddingReplayByteIdentical(t *testing.T) {
	dir := t.TempDir() + "/w"
	w, err := world.Create(dir, "embed", 42)
	if err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(w.DBPath())
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	// Stub embedding server: OpenAI-compat /embeddings (counted — the SC-001
	// meter) plus the Ollama-native /api/embed warm pin (accepted, not an
	// embedding computation). No live Ollama anywhere (plan constraint).
	var embedHits, totalHits atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		totalHits.Add(1)
		switch r.URL.Path {
		case "/embeddings":
			embedHits.Add(1)
			var req struct {
				Input []string `json:"input"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
				return
			}
			data := make([]map[string]any, len(req.Input))
			for i, text := range req.Input {
				data[i] = map[string]any{"index": i, "embedding": []float32{float32(len(text)), 0.25, 0.5}}
			}
			rw.Header().Set("Content-Type", "application/json")
			json.NewEncoder(rw).Encode(map[string]any{"data": data})
		case "/api/embed":
			rw.Header().Set("Content-Type", "application/json")
			json.NewEncoder(rw).Encode(map[string]any{"model": "all-minilm"})
		default:
			http.NotFound(rw, r)
		}
	}))
	defer srv.Close()

	// A v2 registry routing every chat kind to a never-called provider and the
	// embedding kind to the stub — the exact wiring gate daemon boot checks.
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
	if !orch.HasEmbedding() {
		t.Fatal("embedding route declared but HasEmbedding = false")
	}

	// Loop + embedder wired exactly as daemon boot does: the embedder is one
	// more notify consumer injecting through the InjectSocial door.
	state := sim.NewState(w.Manifest.Seed, w.Map())
	var emb *mind.Embedder
	notify := func(evs []store.Event) {
		if emb != nil {
			emb.Observe(evs)
		}
	}
	loop := sim.NewLoop(state, w.Map(), st, notify)
	emb, err = mind.NewEmbedder(orch, loop, nil, w.Map(), w.Manifest.Seed, state.Marshal())
	if err != nil {
		t.Fatal(err)
	}
	defer emb.Close()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- loop.Run(ctx) }()

	// Pause the clock: injections are pause-open, so the run stays quiet and
	// deterministic — every event in the log is one this test placed there.
	if _, err := loop.Do("pause", ""); err != nil {
		t.Fatalf("pause: %v", err)
	}

	// Seed episodic memories through the whitelisted door (the same path a
	// consolidation gist takes). The embedder must converge every one of them
	// to a recorded vector.
	memory := func(agent int, text string) store.Event {
		b, _ := json.Marshal(sim.MemoryAddedPayload{Agent: sim.Ref(agent), Text: text, Salience: 4, Subject: sim.Ref(-1), Origin: sim.OriginAction})
		return store.Event{Type: "agent.memory_added", Payload: b}
	}
	if err := loop.InjectSocial([]store.Event{
		memory(0, "Foraged by the river and found little."),
		memory(0, "Birch shared a meal with me at the fire."),
		memory(1, "The storm broke my lean-to."),
	}); err != nil {
		t.Fatalf("inject memories: %v", err)
	}

	// SC-002: poll until 100% of emitted memories carry a vector + model.
	embedded := func() (bool, *sim.State) {
		raw, _, err := loop.DoState()
		if err != nil {
			t.Fatalf("state: %v", err)
		}
		var s sim.State
		if err := json.Unmarshal(raw, &s); err != nil {
			t.Fatalf("unmarshal state: %v", err)
		}
		total, withVec := 0, 0
		for _, a := range s.Agents {
			for _, m := range a.Memories {
				total++
				if m.Seq == 0 {
					t.Fatalf("memory %q has no stamped seq — the reducer/loop seq stamping broke", m.Text)
				}
				if len(m.Vec) > 0 && m.VecModel == "all-minilm" {
					withVec++
				}
			}
		}
		return total == 3 && withVec == total, &s
	}
	deadline := time.Now().Add(10 * time.Second)
	for {
		if ok, _ := embedded(); ok {
			break
		}
		if time.Now().After(deadline) {
			_, s := embedded()
			t.Fatalf("memories never fully embedded (SC-002); state agents: %+v", s.Agents[0].Memories)
		}
		time.Sleep(20 * time.Millisecond)
	}
	if embedHits.Load() == 0 {
		t.Fatal("no /embeddings calls recorded during the live run")
	}

	cancel()
	if err := <-done; err != nil {
		t.Fatalf("loop.Run: %v", err)
	}
	live := state.Marshal()

	// SC-001: replay through the daemon's own recovery path — the embedder is
	// NOT wired (recoverState touches no model), the replayed state is
	// byte-identical, and the stub receives ZERO calls of any kind.
	hitsBefore := totalHits.Load()
	replayed, err := recoverState(w, st)
	if err != nil {
		t.Fatalf("recoverState: %v", err)
	}
	if got := string(replayed.Marshal()); got != string(live) {
		t.Errorf("replay diverged from the live state:\nlive:     %s\nreplayed: %s", live, got)
	}
	if totalHits.Load() != hitsBefore {
		t.Errorf("replay reached the embedding endpoint (%d extra calls) — replay must perform zero embedding computation", totalHits.Load()-hitsBefore)
	}
	// The vectors in the replayed state came purely from recorded events.
	for ai, a := range replayed.Agents {
		for _, m := range a.Memories {
			if len(m.Vec) == 0 || m.VecModel != "all-minilm" || m.Seq == 0 {
				t.Errorf("replayed agent %d memory %q lost its recorded vector identity: seq=%d model=%q", ai, m.Text, m.Seq, m.VecModel)
			}
		}
	}

	// A full log replay from genesis (not just snapshot recovery) also
	// reproduces the vectors — the raw ReplayEvents leg of the harness.
	fresh := sim.NewState(w.Manifest.Seed, w.Map())
	if err := st.ReplayEvents(0, func(e store.Event) error {
		if e.Tick > fresh.Tick {
			fresh.Tick = e.Tick
		}
		return fresh.Apply(e)
	}); err != nil {
		t.Fatalf("full replay: %v", err)
	}
	if got := string(fresh.Marshal()); got != string(live) {
		t.Errorf("genesis replay diverged from the live state:\nlive:     %s\nreplayed: %s", live, got)
	}
}

// TestOnWorldReplayByteIdentical (spec 042 T022, US3 AC4): a world whose
// manifest carries memory_relevance "on" runs UNPAUSED with the embedder
// wired — natural sim events cross a PlannerCadenceTicks bucket edge, so
// recorded situation vectors land alongside memory companions — then the log
// replays byte-identically through both recovery paths with zero embedding
// calls: selection under "on" is a pure function of exactly this recorded
// state, so replay determinism is untouched by the mode.
func TestOnWorldReplayByteIdentical(t *testing.T) {
	dir := t.TempDir() + "/w"
	w, err := world.Create(dir, "on-world", 42)
	if err != nil {
		t.Fatal(err)
	}
	// Flip the world to "on" on disk (the operator's world.json edit) and
	// re-open through the validated path.
	w.Manifest.MemoryRelevance = world.MemoryRelevanceOn
	data, err := json.MarshalIndent(w.Manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, world.ManifestName), append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	if w, err = world.Open(dir); err != nil {
		t.Fatalf("re-open on-mode world: %v", err)
	}
	if w.Manifest.MemoryRelevance != world.MemoryRelevanceOn {
		t.Fatal("memory_relevance did not persist as on")
	}
	st, err := store.Open(w.DBPath())
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	var totalHits atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		totalHits.Add(1)
		switch r.URL.Path {
		case "/embeddings":
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
		case "/api/embed":
			rw.Header().Set("Content-Type", "application/json")
			json.NewEncoder(rw).Encode(map[string]any{"model": "all-minilm"})
		default:
			http.NotFound(rw, r)
		}
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

	state := sim.NewState(w.Manifest.Seed, w.Map())
	var emb *mind.Embedder
	notify := func(evs []store.Event) {
		if emb != nil {
			emb.Observe(evs)
		}
	}
	loop := sim.NewLoop(state, w.Map(), st, notify)
	emb, err = mind.NewEmbedder(orch, loop, nil, w.Map(), w.Manifest.Seed, state.Marshal())
	if err != nil {
		t.Fatal(err)
	}
	defer emb.Close()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- loop.Run(ctx) }()

	// Spin to cross the 1800-tick cadence bucket fast — through the recorded
	// set_speed door, so genesis replay reproduces the speed byte-identically
	// (the TestTeachingPostureReplayByteIdentical rule: never poke state
	// directly around a running loop).
	if _, err := loop.Do("set_speed", "max"); err != nil {
		t.Fatalf("set_speed: %v", err)
	}

	// Run until a recorded situation vector lands (the bucket edge fired and
	// its companion committed), then freeze the world and seed memories so
	// both companion kinds are in the log.
	countType := func(typ string) int {
		n := 0
		st.ReplayEvents(0, func(e store.Event) error {
			if e.Type == typ {
				n++
			}
			return nil
		})
		return n
	}
	deadline := time.Now().Add(20 * time.Second)
	for countType("agent.situation_embedded") == 0 {
		if time.Now().After(deadline) {
			t.Fatal("no agent.situation_embedded landed within 20s of max-speed running")
		}
		time.Sleep(30 * time.Millisecond)
	}
	if _, err := loop.Do("pause", ""); err != nil {
		t.Fatalf("pause: %v", err)
	}
	memory := func(agent int, text string) store.Event {
		b, _ := json.Marshal(sim.MemoryAddedPayload{Agent: sim.Ref(agent), Text: text, Salience: 4, Subject: sim.Ref(-1), Origin: sim.OriginAction})
		return store.Event{Type: "agent.memory_added", Payload: b}
	}
	if err := loop.InjectSocial([]store.Event{
		memory(0, "Watched the river rise."),
		memory(1, "Shared a meal at the fire."),
	}); err != nil {
		t.Fatalf("inject memories: %v", err)
	}
	for countType("agent.memory_embedded") < 2 {
		if time.Now().After(deadline) {
			t.Fatal("injected memories never gained companions")
		}
		time.Sleep(30 * time.Millisecond)
	}

	cancel()
	if err := <-done; err != nil {
		t.Fatalf("loop.Run: %v", err)
	}
	live := state.Marshal()

	// Stop the driver and drain to endpoint quiescence: a natural memory
	// dequeued just before shutdown may still complete its embed call (its
	// companion is refused by the stopped loop — vectorless forever, fine).
	// Those trailing driver calls are NOT replay computation; meter replay
	// only once the endpoint has gone quiet.
	emb.Close()
	prev := totalHits.Load()
	for {
		time.Sleep(150 * time.Millisecond)
		cur := totalHits.Load()
		if cur == prev {
			break
		}
		prev = cur
	}

	// Replay both recovery paths with the embedder unwired: byte-identical,
	// zero calls of any kind to the embedding endpoint.
	hitsBefore := totalHits.Load()
	replayed, err := recoverState(w, st)
	if err != nil {
		t.Fatalf("recoverState: %v", err)
	}
	if got := string(replayed.Marshal()); got != string(live) {
		t.Error("on-world snapshot recovery diverged from the live state")
	}
	fresh := sim.NewState(w.Manifest.Seed, w.Map())
	if err := st.ReplayEvents(0, func(e store.Event) error {
		if e.Tick > fresh.Tick {
			fresh.Tick = e.Tick
		}
		return fresh.Apply(e)
	}); err != nil {
		t.Fatalf("full replay: %v", err)
	}
	if got := string(fresh.Marshal()); got != string(live) {
		t.Error("on-world genesis replay diverged from the live state")
	}
	if totalHits.Load() != hitsBefore {
		t.Errorf("replay reached the embedding endpoint (%d extra calls)", totalHits.Load()-hitsBefore)
	}
	// The recorded situation vectors — the "on" selector's query inputs — are
	// reproduced purely from the log.
	found := false
	for _, a := range replayed.Agents {
		if len(a.SitVec) > 0 && a.SitVecModel == "all-minilm" && a.SitVecTick > 0 {
			found = true
		}
	}
	if !found {
		t.Error("replayed state carries no situation vector — the on-mode query inputs did not survive replay")
	}
}
