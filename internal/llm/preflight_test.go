package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// --- preflight test scaffolding ---

// preflightOrch builds an orchestrator with one openai_compat provider ("local",
// endpoint + model under test) and one anthropic provider ("cloud", exempt from
// preflight). It is the standard fixture for the models-list probe tests.
func preflightOrch(t *testing.T, localEndpoint, model string) *Orchestrator {
	t.Helper()
	t.Setenv("PROMPTWORLD_TEST_KEY", "test-key")
	cfg := Config{
		MonthlyBudgetUSD: 100,
		Providers: map[string]ProviderConfig{
			"local": {Transport: ProviderOpenAICompat, Endpoint: localEndpoint, Model: model},
			"cloud": {Transport: ProviderAnthropic, Model: "claude-x", InputUSDPerMTok: 5, OutputUSDPerMTok: 25, APIKeyEnv: "PROMPTWORLD_TEST_KEY"},
		},
		Routes: defaultRoutes(),
	}
	o, err := New(cfg, testStore(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(o.Close)
	return o
}

// modelsServer serves a /models listing with the given model ids (OpenAI-compat
// shape). Any other path 404s, so a probe that hit the wrong path is caught.
func modelsServer(t *testing.T, ids ...string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			http.NotFound(w, r)
			return
		}
		data := make([]map[string]any, 0, len(ids))
		for _, id := range ids {
			data = append(data, map[string]any{"id": id})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "list", "data": data})
	}))
	t.Cleanup(srv.Close)
	return srv
}

// closedEndpoint returns a base URL whose port was bound then immediately
// released — a connection to it refuses at once (the "endpoint unreachable"
// fixture, no timeout wait).
func closedEndpoint(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	return "http://" + addr
}

// compressClock shrinks the preflight cadence + timeout for the lifecycle tests
// and restores them after, so a 60s/5s real-time loop runs in milliseconds.
func compressClock(t *testing.T, interval, timeout time.Duration) {
	t.Helper()
	oi, ot := preflightInterval, preflightTimeout
	preflightInterval, preflightTimeout = interval, timeout
	t.Cleanup(func() { preflightInterval, preflightTimeout = oi, ot })
}

// startPreflight launches RunPreflight in its own goroutine (as production
// code does) and registers a cleanup that cancels ctx and BLOCKS until the
// goroutine has actually returned. TASK-93: a bare `go o.RunPreflight(ctx)`
// plus `defer cancel()` only asks the goroutine to stop — it does not wait
// for it to — so the goroutine's reads of preflightInterval/preflightTimeout
// (ticker construction, probeModels' context.WithTimeout) could still be in
// flight when compressClock's own t.Cleanup ran and wrote those same package
// vars back to their real-time values: a data race that only surfaced under
// -race. Cleanups run LIFO, so registering this one AFTER compressClock's
// call guarantees it runs FIRST: cancel, then wait for the goroutine to
// exit, and only THEN is compressClock allowed to restore the vars.
func startPreflight(t *testing.T, o *Orchestrator, ctx context.Context, cancel context.CancelFunc) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		o.RunPreflight(ctx)
	}()
	t.Cleanup(func() {
		cancel()
		<-done
	})
}

// capturePreflightLog redirects preflightLogf into a thread-safe buffer for the
// duration of a test, returning a reader for the accumulated lines.
func capturePreflightLog(t *testing.T) func() []string {
	t.Helper()
	var mu sync.Mutex
	var lines []string
	orig := preflightLogf
	preflightLogf = func(format string, args ...any) {
		mu.Lock()
		defer mu.Unlock()
		lines = append(lines, fmt.Sprintf(format, args...))
	}
	t.Cleanup(func() { preflightLogf = orig })
	return func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), lines...)
	}
}

// --- probe classification (T006) ---

// TestProbeClassification exercises every branch of the models-list probe
// classifier (contracts/provider-conditions.md): healthy / model-missing /
// endpoint-unreachable / listing-unsupported (404 and garbage JSON and a
// non-listing 2xx body).
func TestProbeClassification(t *testing.T) {
	ctx := context.Background()

	t.Run("healthy", func(t *testing.T) {
		srv := modelsServer(t, "other", "cogito:3b")
		o := preflightOrch(t, srv.URL, "cogito:3b")
		if got := o.probeModels(ctx, o.providers["local"]); got != probeHealthy {
			t.Errorf("probe = %d, want probeHealthy", got)
		}
	})

	t.Run("model-missing (valid listing, id absent)", func(t *testing.T) {
		srv := modelsServer(t, "other-model")
		o := preflightOrch(t, srv.URL, "cogito:3b")
		if got := o.probeModels(ctx, o.providers["local"]); got != probeMissing {
			t.Errorf("probe = %d, want probeMissing", got)
		}
	})

	t.Run("model-missing (empty but valid listing)", func(t *testing.T) {
		srv := modelsServer(t) // {"data":[]} — a real, empty listing
		o := preflightOrch(t, srv.URL, "cogito:3b")
		if got := o.probeModels(ctx, o.providers["local"]); got != probeMissing {
			t.Errorf("probe = %d, want probeMissing (empty listing is not unsupported)", got)
		}
	})

	t.Run("unreachable (closed port)", func(t *testing.T) {
		o := preflightOrch(t, closedEndpoint(t), "cogito:3b")
		if got := o.probeModels(ctx, o.providers["local"]); got != probeUnreachable {
			t.Errorf("probe = %d, want probeUnreachable", got)
		}
	})

	t.Run("listing-unsupported (404)", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		}))
		t.Cleanup(srv.Close)
		o := preflightOrch(t, srv.URL, "cogito:3b")
		if got := o.probeModels(ctx, o.providers["local"]); got != probeUnsupported {
			t.Errorf("probe = %d, want probeUnsupported (404)", got)
		}
	})

	t.Run("listing-unsupported (garbage JSON)", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not json{{{"))
		}))
		t.Cleanup(srv.Close)
		o := preflightOrch(t, srv.URL, "cogito:3b")
		if got := o.probeModels(ctx, o.providers["local"]); got != probeUnsupported {
			t.Errorf("probe = %d, want probeUnsupported (garbage)", got)
		}
	})

	t.Run("listing-unsupported (2xx, no data key)", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"models":["cogito:3b"]}`)) // valid JSON, wrong shape
		}))
		t.Cleanup(srv.Close)
		o := preflightOrch(t, srv.URL, "cogito:3b")
		if got := o.probeModels(ctx, o.providers["local"]); got != probeUnsupported {
			t.Errorf("probe = %d, want probeUnsupported (no data key)", got)
		}
	})
}

// TestProbeSendsAuthHeader proves the probe applies the chat-completions auth
// rule: a Bearer header when a key resolves, none for an open endpoint.
func TestProbeSendsAuthHeader(t *testing.T) {
	ctx := context.Background()
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"id": "m"}}})
	}))
	t.Cleanup(srv.Close)

	t.Setenv("PROMPTWORLD_TEST_KEY", "test-key")
	cfg := Config{
		MonthlyBudgetUSD: 100,
		Providers: map[string]ProviderConfig{
			"local": {Transport: ProviderOpenAICompat, Endpoint: srv.URL, Model: "m", APIKey: "secret-key"},
			"cloud": {Transport: ProviderAnthropic, Model: "claude-x", InputUSDPerMTok: 5, OutputUSDPerMTok: 25, APIKeyEnv: "PROMPTWORLD_TEST_KEY"},
		},
		Routes: defaultRoutes(),
	}
	o, err := New(cfg, testStore(t))
	if err != nil {
		t.Fatal(err)
	}
	defer o.Close()
	o.probeModels(ctx, o.providers["local"])
	if gotAuth != "Bearer secret-key" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer secret-key")
	}
}

// --- condition reconciliation (T006/T007) ---

// TestPreflightHealthyRaisesNothing proves a probe of a served model raises no
// condition and fires no hook — a healthy provider is silent.
func TestPreflightHealthyRaisesNothing(t *testing.T) {
	srv := modelsServer(t, "cogito:3b")
	o := preflightOrch(t, srv.URL, "cogito:3b")
	var rec condRecorder
	o.SetConditionHook(rec.hook)

	if active := o.preflightProbe(context.Background(), o.providers["local"]); active {
		t.Error("healthy probe reported an active condition")
	}
	if got := o.providers["local"].conditionSnapshot().kind; got != "" {
		t.Errorf("healthy provider slot = %q, want empty", got)
	}
	if fires := rec.snapshot(); len(fires) != 0 {
		t.Errorf("healthy probe fired the hook %d times, want 0: %+v", len(fires), fires)
	}
}

// TestPreflightModelMissing proves a valid listing without the configured model
// raises model-missing with the contract's detail + remedy and one raise fire.
func TestPreflightModelMissing(t *testing.T) {
	srv := modelsServer(t, "some-other-model")
	o := preflightOrch(t, srv.URL, "cogito:3b")
	var rec condRecorder
	o.SetConditionHook(rec.hook)

	if active := o.preflightProbe(context.Background(), o.providers["local"]); !active {
		t.Error("model-missing probe reported no active condition")
	}
	c := o.providers["local"].conditionSnapshot()
	if c.kind != CondModelMissing {
		t.Fatalf("slot = %q, want model-missing", c.kind)
	}
	if !contains(c.detail, `"cogito:3b"`) || !contains(c.detail, srv.URL) {
		t.Errorf("detail = %q, want the model + endpoint", c.detail)
	}
	if c.remedy != "ollama pull cogito:3b" {
		t.Errorf("remedy = %q, want the pull command", c.remedy)
	}
	fires := rec.snapshot()
	if len(fires) != 1 || !fires[0].active || fires[0].kind != string(CondModelMissing) {
		t.Fatalf("want one active model-missing fire, got %+v", fires)
	}
}

// TestPreflightUnreachable proves a closed endpoint raises endpoint-unreachable
// with the "start the model server" remedy.
func TestPreflightUnreachable(t *testing.T) {
	endpoint := closedEndpoint(t)
	o := preflightOrch(t, endpoint, "cogito:3b")
	var rec condRecorder
	o.SetConditionHook(rec.hook)

	if active := o.preflightProbe(context.Background(), o.providers["local"]); !active {
		t.Error("unreachable probe reported no active condition")
	}
	c := o.providers["local"].conditionSnapshot()
	if c.kind != CondEndpointUnreachable {
		t.Fatalf("slot = %q, want endpoint-unreachable", c.kind)
	}
	if !contains(c.remedy, "start the model server") {
		t.Errorf("remedy = %q, want the start-server remedy", c.remedy)
	}
	fires := rec.snapshot()
	if len(fires) != 1 || fires[0].kind != string(CondEndpointUnreachable) {
		t.Fatalf("want one endpoint-unreachable fire, got %+v", fires)
	}
}

// TestPreflightListingUnsupportedNeverRaises proves the graceful-skip path: a
// non-listing endpoint raises NO condition, fires NO hook, and emits exactly one
// low-key skip line — never a false model-missing (spec 034 edge case).
func TestPreflightListingUnsupportedNeverRaises(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	o := preflightOrch(t, srv.URL, "cogito:3b")
	var rec condRecorder
	o.SetConditionHook(rec.hook)
	readLog := capturePreflightLog(t)

	if active := o.preflightProbe(context.Background(), o.providers["local"]); active {
		t.Error("unsupported probe reported an active condition")
	}
	if got := o.providers["local"].conditionSnapshot().kind; got != "" {
		t.Errorf("unsupported endpoint raised %q, want no condition", got)
	}
	if fires := rec.snapshot(); len(fires) != 0 {
		t.Errorf("unsupported probe fired the hook %d times, want 0", len(fires))
	}
	lines := readLog()
	if len(lines) != 1 || !contains(lines[0], "does not expose a model listing") {
		t.Errorf("want one skip line, got %v", lines)
	}
}

// TestSetPreflightConditionReclassifies proves both directions of the
// preflight-owned reclassify (data-model: unreachable ⇄ missing), including the
// DOWNWARD case (unreachable → missing) that raiseCondition's precedence guard
// would otherwise drop. Each reclassify fires a clear (of the old kind) then a
// raise (of the new); a steady re-set of the same kind fires nothing.
func TestSetPreflightConditionReclassifies(t *testing.T) {
	var rec condRecorder
	o := &Orchestrator{}
	o.SetConditionHook(rec.hook)
	p := &provider{name: "local"}

	// missing → unreachable (upward): raiseCondition would allow this natively.
	o.setPreflightCondition(p, CondModelMissing, "d1", "r1")
	o.setPreflightCondition(p, CondEndpointUnreachable, "d2", "r2")
	if got := p.conditionSnapshot().kind; got != CondEndpointUnreachable {
		t.Fatalf("after upward reclassify slot = %q, want endpoint-unreachable", got)
	}

	// unreachable → missing (DOWNWARD): must reclassify, not be dropped.
	o.setPreflightCondition(p, CondModelMissing, "d3", "r3")
	if got := p.conditionSnapshot().kind; got != CondModelMissing {
		t.Fatalf("downward reclassify dropped: slot = %q, want model-missing", got)
	}

	// steady re-set of the same kind + detail: no transition.
	o.setPreflightCondition(p, CondModelMissing, "d3", "r3")

	fires := rec.snapshot()
	want := []condFire{
		{"local", "model-missing", "d1", "r1", true},        // first raise
		{"local", "model-missing", "d1", "", false},         // clear before upward reclassify
		{"local", "endpoint-unreachable", "d2", "r2", true}, // upward raise
		{"local", "endpoint-unreachable", "d2", "", false},  // clear before downward reclassify
		{"local", "model-missing", "d3", "r3", true},        // downward raise
	}
	if len(fires) != len(want) {
		t.Fatalf("fired %d times, want %d: %+v", len(fires), len(want), fires)
	}
	for i := range want {
		if fires[i] != want[i] {
			t.Errorf("fire[%d] = %+v, want %+v", i, fires[i], want[i])
		}
	}
}

// TestPreflightAnthropicSkipped proves an anthropic provider is exempt: it is
// never in the eligible set, so no probe (and no false condition) ever touches
// it (FR-001).
func TestPreflightAnthropicSkipped(t *testing.T) {
	srv := modelsServer(t, "cogito:3b")
	o := preflightOrch(t, srv.URL, "cogito:3b")
	eligible := o.preflightEligible()
	if len(eligible) != 1 || eligible[0].name != "local" {
		names := make([]string, len(eligible))
		for i, p := range eligible {
			names[i] = p.name
		}
		t.Fatalf("eligible = %v, want [local] only (anthropic exempt)", names)
	}
}

// --- lifecycle loop (T007) ---

// flipModels is a /models server whose returned ids can be swapped mid-test, so
// a single stable URL can go from model-missing to healthy (a pulled model).
type flipModels struct {
	mu  sync.Mutex
	ids []string
}

func (f *flipModels) set(ids ...string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.ids = ids
}

func (f *flipModels) handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/models" {
		http.NotFound(w, r)
		return
	}
	f.mu.Lock()
	ids := append([]string(nil), f.ids...)
	f.mu.Unlock()
	data := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		data = append(data, map[string]any{"id": id})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"object": "list", "data": data})
}

// TestPreflightReprobeClearsOnPull proves the FR-004 no-restart recovery: a
// world whose model is missing raises the condition at boot, keeps re-probing on
// the compressed cadence, and CLEARS on its own once the model is "pulled"
// (served) — no traffic and no restart required. It also confirms the active
// re-probe re-logs the standing warning (repeat-loudness) while the durable hook
// stays transitions-only (one raise + one clear).
func TestPreflightReprobeClearsOnPull(t *testing.T) {
	compressClock(t, 2*time.Millisecond, time.Second)
	readLog := capturePreflightLog(t)

	flip := &flipModels{ids: []string{"other-model"}} // model absent at boot
	srv := httptest.NewServer(http.HandlerFunc(flip.handler))
	t.Cleanup(srv.Close)

	o := preflightOrch(t, srv.URL, "cogito:3b")
	var rec condRecorder
	o.SetConditionHook(rec.hook)

	ctx, cancel := context.WithCancel(context.Background())
	startPreflight(t, o, ctx, cancel)

	// Condition raised (model-missing) within a few cadences.
	waitFor(t, "model-missing raised", func() bool {
		return o.providers["local"].conditionSnapshot().kind == CondModelMissing
	})

	// Pull the model: the endpoint now serves it.
	flip.set("cogito:3b")

	// Condition clears on its own — no restart.
	waitFor(t, "condition cleared after pull", func() bool {
		return o.providers["local"].conditionSnapshot().kind == ""
	})
	cancel()

	// Hook fired exactly the two transitions (raise, clear), never per re-probe.
	fires := rec.snapshot()
	if len(fires) != 2 {
		t.Fatalf("hook fired %d times, want 2 (raise + clear): %+v", len(fires), fires)
	}
	if !fires[0].active || fires[0].kind != string(CondModelMissing) {
		t.Errorf("fire[0] = %+v, want active model-missing", fires[0])
	}
	if fires[1].active || fires[1].remedy != "" {
		t.Errorf("fire[1] = %+v, want a clear (active=false, remedy empty)", fires[1])
	}
	// Repeat-loudness: at least one re-probe re-logged the standing warning.
	var warned bool
	for _, line := range readLog() {
		if contains(line, "WARNING llm provider") {
			warned = true
			break
		}
	}
	if !warned {
		t.Error("active re-probe never re-logged the standing warning")
	}
}

// TestPreflightBootNeverFails proves FR-002: constructing the orchestrator and
// starting preflight against an UNREACHABLE local endpoint leaves the
// orchestrator fully serving — a call routed to a healthy provider still flows
// while preflight raises the dead provider's condition in the background.
func TestPreflightBootNeverFails(t *testing.T) {
	compressClock(t, 2*time.Millisecond, time.Second)

	var ch atomic.Int64
	cloud := mockCloud(t, &ch)

	t.Setenv("PROMPTWORLD_TEST_KEY", "test-key")
	cfg := Config{
		MonthlyBudgetUSD: 100,
		Providers: map[string]ProviderConfig{
			"local": {Transport: ProviderOpenAICompat, Endpoint: closedEndpoint(t), Model: "cogito:3b"},
			"cloud": {Transport: ProviderAnthropic, Model: "claude-x", Endpoint: cloud.URL, InputUSDPerMTok: 5, OutputUSDPerMTok: 25, APIKeyEnv: "PROMPTWORLD_TEST_KEY"},
		},
		Routes: defaultRoutes(),
	}
	o, err := New(cfg, testStore(t))
	if err != nil {
		t.Fatalf("New must not fail on an unreachable endpoint: %v", err)
	}
	defer o.Close()

	ctx, cancel := context.WithCancel(context.Background())
	startPreflight(t, o, ctx, cancel)

	// The orchestrator serves: a consolidation call routes to the healthy cloud
	// provider and flows, regardless of the dead local endpoint + running probe.
	resp, err := o.Submit(context.Background(), Request{Kind: KindConsolidation, Prompt: "x"})
	if err != nil {
		t.Fatalf("submit to healthy provider failed while preflight ran: %v", err)
	}
	if resp.Provider != "cloud" {
		t.Errorf("served by %q, want cloud", resp.Provider)
	}
	if ch.Load() == 0 {
		t.Error("cloud provider never served the call")
	}

	// And preflight did its job in the background: the dead local endpoint gets
	// an endpoint-unreachable condition without ever affecting serving.
	waitFor(t, "local flagged unreachable", func() bool {
		return o.providers["local"].conditionSnapshot().kind == CondEndpointUnreachable
	})
}

// --- small helpers ---

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("timed out waiting for: %s", what)
}

func contains(s, sub string) bool { return strings.Contains(s, sub) }
