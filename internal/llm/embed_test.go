package llm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
)

// mockEmbeddings is an OpenAI-compatible /embeddings server: one fixed-shape
// vector per input text, index-ordered, counting hits — no live Ollama in any
// test (spec 042 plan constraint).
func mockEmbeddings(t *testing.T, hits *atomic.Int64) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/embeddings" {
			http.NotFound(w, r)
			return
		}
		hits.Add(1)
		var req struct {
			Model string   `json:"model"`
			Input []string `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		data := make([]map[string]any, len(req.Input))
		for i, text := range req.Input {
			data[i] = map[string]any{
				"index":     i,
				"embedding": []float32{float32(len(text)), float32(i), 0.5},
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": data})
	}))
	t.Cleanup(srv.Close)
	return srv
}

// TestOpenAICompatEmbed (spec 042 T002): the transport posts {"model","input"}
// to endpoint+"/embeddings", decodes data[].embedding index-ordered, and
// returns the PROVIDER'S configured model string — the authoritative vector
// identity — one vector per input text.
func TestOpenAICompatEmbed(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/embeddings" {
			http.NotFound(w, r)
			return
		}
		json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		// Rows deliberately out of order: the decoder must honor the index field.
		json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{
			{"index": 1, "embedding": []float32{4, 5, 6}},
			{"index": 0, "embedding": []float32{1, 2, 3}},
		}})
	}))
	t.Cleanup(srv.Close)

	o := newOpenAICompat(srv.URL, "all-minilm", "", "", "")
	vecs, model, err := o.Embed(context.Background(), []string{"first", "second"})
	if err != nil {
		t.Fatal(err)
	}
	if body["model"] != "all-minilm" {
		t.Errorf("wire model = %v, want all-minilm", body["model"])
	}
	if in, ok := body["input"].([]any); !ok || len(in) != 2 || in[0] != "first" {
		t.Errorf("wire input = %v, want the two texts verbatim", body["input"])
	}
	if model != "all-minilm" {
		t.Errorf("returned model = %q, want the provider's configured model", model)
	}
	want := [][]float32{{1, 2, 3}, {4, 5, 6}}
	if !reflect.DeepEqual(vecs, want) {
		t.Errorf("vectors = %v, want %v (index-ordered)", vecs, want)
	}
}

// TestOpenAICompatEmbedCountMismatch: a server replying with the wrong number
// of vectors is a transport error, never a silent partial result.
func TestOpenAICompatEmbedCountMismatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{
			{"index": 0, "embedding": []float32{1}},
		}})
	}))
	t.Cleanup(srv.Close)
	o := newOpenAICompat(srv.URL, "m", "", "", "")
	if _, _, err := o.Embed(context.Background(), []string{"a", "b"}); err == nil {
		t.Fatal("1 vector for 2 inputs must error")
	}
}

// TestAnthropicEmbedUnsupported (spec 042 T002): the anthropic transport's
// Embed is the typed refusal — ErrEmbeddingUnsupported, never an HTTP call.
func TestAnthropicEmbedUnsupported(t *testing.T) {
	a := newAnthropicCaller(CloudConfig{Model: "claude"})
	if _, _, err := a.Embed(context.Background(), []string{"x"}); !errors.Is(err, ErrEmbeddingUnsupported) {
		t.Errorf("anthropic Embed err = %v, want ErrEmbeddingUnsupported", err)
	}
}

// embedRoutes is validRoutes plus the embedding route to a dedicated
// openai_compat provider — the positive baseline for the kind's validation.
const embedRoutes = `{"planner":["gemma"],"conversation":["gemma"],"meeting":["gemma"],` +
	`"consolidation":["anthropic"],"narrator":["anthropic"],"drama":["anthropic"],"metatron":["anthropic"],` +
	`"metatron_watch":["gemma","anthropic"],"embedding":["embedder"]}`

// v2EmbedBody is v2Body with a third, embedding-capable provider declared.
func v2EmbedBody(routes string) string {
	return `{"monthly_budget_usd":100,` +
		`"providers":{` +
		`"gemma":{"transport":"openai_compat","endpoint":"http://x/v1","model":"g"},` +
		`"embedder":{"transport":"openai_compat","endpoint":"http://x/v1","model":"all-minilm"},` +
		`"anthropic":{"transport":"anthropic","model":"claude","input_usd_per_mtok":5}},` +
		`"routes":` + routes + `}`
}

// TestEmbeddingRouteOptional (spec 042 T001): an ABSENT embedding route is the
// subsystem's off switch — the config loads with no error and no backfill (the
// deliberate deviation from the warn-backfill pattern), and the resolved routes
// carry no embedding entry.
func TestEmbeddingRouteOptional(t *testing.T) {
	var logged []string
	prev := configWarnf
	configWarnf = func(format string, args ...any) { logged = append(logged, format) }
	defer func() { configWarnf = prev }()

	cfg, err := LoadConfig(writeConfigFile(t, v2Body(validRoutes)))
	if err != nil {
		t.Fatalf("config without an embedding route must load: %v", err)
	}
	_, routes, err := cfg.resolveRegistry()
	if err != nil {
		t.Fatalf("resolveRegistry: %v", err)
	}
	if _, ok := routes[KindEmbedding]; ok {
		t.Error("absent embedding route was backfilled — absence must mean OFF")
	}
	for _, l := range logged {
		if strings.Contains(l, "embedding") {
			t.Errorf("absent embedding route produced a warn line %q — must be silent at the config layer", l)
		}
	}
}

// TestEmbeddingRouteValidates (spec 042 T001): a present embedding route gets
// the full chain validation, and naming an anthropic-transport provider
// anywhere in its chain is a boot error naming the offender (the Messages API
// serves no embeddings endpoint).
func TestEmbeddingRouteValidates(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string
	}{
		{
			"embedding routed to the anthropic transport",
			v2EmbedBody(strings.Replace(embedRoutes, `"embedding":["embedder"]`, `"embedding":["anthropic"]`, 1)),
			"embeddings",
		},
		{
			"embedding route names undeclared provider",
			v2EmbedBody(strings.Replace(embedRoutes, `"embedding":["embedder"]`, `"embedding":["ghost"]`, 1)),
			"ghost",
		},
		{
			"embedding route with empty chain",
			v2EmbedBody(strings.Replace(embedRoutes, `"embedding":["embedder"]`, `"embedding":[]`, 1)),
			"embedding",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := LoadConfig(writeConfigFile(t, c.body))
			if err == nil {
				t.Fatalf("expected a boot error naming %q, got nil", c.want)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("boot error %q does not name %q", err.Error(), c.want)
			}
		})
	}
	// Positive control: the valid embedding baseline loads and resolves.
	cfg, err := LoadConfig(writeConfigFile(t, v2EmbedBody(embedRoutes)))
	if err != nil {
		t.Fatalf("valid embedding config rejected: %v", err)
	}
	_, routes, err := cfg.resolveRegistry()
	if err != nil {
		t.Fatalf("resolveRegistry: %v", err)
	}
	if rc, ok := routes[KindEmbedding]; !ok || !reflect.DeepEqual(rc.Chain, []string{"embedder"}) {
		t.Errorf("embedding route = %+v, want chain [embedder]", rc)
	}
}

// TestOrchestratorEmbed (spec 042 T001/T002): New() resolves the embedding
// route into the Embed surface — HasEmbedding/EmbeddingProvider report it, and
// Embed serves vectors through the routed provider's /embeddings endpoint with
// the provider's model as the recorded identity. Without the route the surface
// reports off and Embed refuses with ErrEmbeddingOff.
func TestOrchestratorEmbed(t *testing.T) {
	var hits atomic.Int64
	srv := mockEmbeddings(t, &hits)

	cfg := Config{
		MonthlyBudgetUSD: 100,
		Providers: map[string]ProviderConfig{
			"gemma":    {Transport: ProviderOpenAICompat, Endpoint: "http://x/v1", Model: "g"},
			"embedder": {Transport: ProviderOpenAICompat, Endpoint: srv.URL, Model: "all-minilm"},
		},
		Routes: map[string]RouteConfig{
			string(KindPlanner):       {Chain: []string{"gemma"}},
			string(KindConversation):  {Chain: []string{"gemma"}},
			string(KindMeeting):       {Chain: []string{"gemma"}},
			string(KindConsolidation): {Chain: []string{"gemma"}},
			string(KindNarrator):      {Chain: []string{"gemma"}},
			string(KindDrama):         {Chain: []string{"gemma"}},
			string(KindMetatron):      {Chain: []string{"gemma"}},
			string(KindMetatronWatch): {Chain: []string{"gemma"}},
			string(KindEmbedding):     {Chain: []string{"embedder"}},
		},
	}
	o := newOrch(t, cfg, testStore(t))
	if !o.HasEmbedding() {
		t.Fatal("HasEmbedding = false with an embedding route declared")
	}
	if name, model, ok := o.EmbeddingProvider(); !ok || name != "embedder" || model != "all-minilm" {
		t.Errorf("EmbeddingProvider = (%q, %q, %v), want (embedder, all-minilm, true)", name, model, ok)
	}
	vecs, model, err := o.Embed(context.Background(), []string{"a memory"})
	if err != nil {
		t.Fatal(err)
	}
	if model != "all-minilm" || len(vecs) != 1 || len(vecs[0]) != 3 {
		t.Errorf("Embed = (%v, %q), want one 3-dim mock vector under all-minilm", vecs, model)
	}
	if hits.Load() != 1 {
		t.Errorf("embeddings endpoint hits = %d, want 1", hits.Load())
	}
	// Submit must never dispatch the embedding kind — it is not a chat route.
	if _, err := o.Submit(context.Background(), Request{Kind: KindEmbedding, Prompt: "x"}); !errors.Is(err, ErrUnknownKind) {
		t.Errorf("Submit(embedding) err = %v, want ErrUnknownKind", err)
	}

	// Routeless orchestrator: the surface reports off and Embed refuses.
	delete(cfg.Routes, string(KindEmbedding))
	off := newOrch(t, cfg, testStore(t))
	if off.HasEmbedding() {
		t.Error("HasEmbedding = true without an embedding route")
	}
	if _, _, err := off.Embed(context.Background(), []string{"x"}); !errors.Is(err, ErrEmbeddingOff) {
		t.Errorf("Embed without a route err = %v, want ErrEmbeddingOff", err)
	}
}
