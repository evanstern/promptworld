package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// withToolSilentThreshold lowers the tool-silence threshold for a test and
// restores it after — crossing the real 8 by hand is a pathological run, and
// the var exists precisely so the detector's boundary is testable at a small n.
func withToolSilentThreshold(t *testing.T, n int) {
	t.Helper()
	old := toolSilentThreshold
	toolSilentThreshold = n
	t.Cleanup(func() { toolSilentThreshold = old })
}

// nativeProvider is a bare provider whose resolved tool mode is native — enough
// state for the unit-level detector tests (observeSuccess needs only the name,
// the resolved tool mode, and the condition slot).
func nativeProvider() *provider {
	return &provider{name: "local", cfg: ProviderConfig{Transport: ProviderOpenAICompat}}
}

// TestToolSilentRaisesAtThreshold (spec 034 US2, AC#1): the detector raises
// tool-silent at EXACTLY the threshold — not before — and fires the transition
// hook once, then never again as the counter climbs past it (AC#2 persistence
// rides the status field, not a repeated hook).
func TestToolSilentRaisesAtThreshold(t *testing.T) {
	withToolSilentThreshold(t, 4)
	var rec condRecorder
	o := &Orchestrator{}
	o.SetConditionHook(rec.hook)
	p := nativeProvider()

	// One short of the threshold: nothing raised, nothing fired.
	for i := 0; i < toolSilentThreshold-1; i++ {
		o.observeSuccess(p, true, false)
	}
	if got := p.conditionSnapshot().kind; got != "" {
		t.Fatalf("raised before threshold (count %d): slot=%q", toolSilentThreshold-1, got)
	}
	if fires := rec.snapshot(); len(fires) != 0 {
		t.Fatalf("hook fired before threshold: %+v", fires)
	}

	// The threshold-crossing completion raises tool-silent with the contract's
	// detail and the native-mode remedy.
	o.observeSuccess(p, true, false)
	c := p.conditionSnapshot()
	if c.kind != CondToolSilent {
		t.Fatalf("slot=%q, want tool-silent at threshold", c.kind)
	}
	wantDetail := fmt.Sprintf("%d consecutive tool-free completions on tool-carrying calls", toolSilentThreshold)
	if c.detail != wantDetail {
		t.Errorf("detail=%q, want %q", c.detail, wantDetail)
	}
	if want := `set providers.local.tool_mode to "json" and restart`; c.remedy != want {
		t.Errorf("remedy=%q, want %q", c.remedy, want)
	}

	// Past the threshold the counter keeps climbing, but the hook does NOT
	// re-fire — the raise is idempotent (identical kind+detail).
	o.observeSuccess(p, true, false)
	o.observeSuccess(p, true, false)
	fires := rec.snapshot()
	if len(fires) != 1 || !fires[0].active || fires[0].kind != string(CondToolSilent) {
		t.Fatalf("want exactly one raise fire, got %+v", fires)
	}
}

// TestToolSilentResetOnToolCall (spec 034 US2, AC#3): a tool call landing
// mid-run resets the consecutive counter, so a subsequent partial run does not
// prematurely trip the detector — isolated tool-free completions never
// accumulate into a false positive.
func TestToolSilentResetOnToolCall(t *testing.T) {
	withToolSilentThreshold(t, 4)
	o := &Orchestrator{}
	p := nativeProvider()

	// Climb to one short of the threshold, then a landed tool call resets the run.
	for i := 0; i < toolSilentThreshold-1; i++ {
		o.observeSuccess(p, true, false)
	}
	o.observeSuccess(p, true, true)
	if got := p.conditionSnapshot().kind; got != "" {
		t.Fatalf("landed tool call left a condition: %q", got)
	}

	// Another full-minus-one run must still not raise — the counter really reset.
	for i := 0; i < toolSilentThreshold-1; i++ {
		o.observeSuccess(p, true, false)
	}
	if got := p.conditionSnapshot().kind; got != "" {
		t.Fatalf("raised despite the mid-run reset: slot=%q", got)
	}
	// The threshold-th completion of the fresh run crosses.
	o.observeSuccess(p, true, false)
	if got := p.conditionSnapshot().kind; got != CondToolSilent {
		t.Fatalf("did not raise after reset + a full run: %q", got)
	}
}

// TestToolSilentNonToolCallsDontCount (spec 034 US2, AC#4, FR-005): calls that
// carry no tool declarations (conversation, meeting) return no tool calls as a
// matter of course — they must never count toward the threshold, and
// interleaving them between tool-carrying completions must not disturb the
// consecutive count either.
func TestToolSilentNonToolCallsDontCount(t *testing.T) {
	withToolSilentThreshold(t, 3)
	o := &Orchestrator{}
	p := nativeProvider()

	// Non-tool calls alone, well past the threshold, never raise.
	for i := 0; i < toolSilentThreshold*3; i++ {
		o.observeSuccess(p, false, false)
	}
	if got := p.conditionSnapshot().kind; got != "" {
		t.Fatalf("non-tool calls tripped the detector: %q", got)
	}

	// Interleaving non-tool calls does not count and does not reset: it still
	// takes exactly `threshold` tool-carrying completions to raise.
	for i := 0; i < toolSilentThreshold-1; i++ {
		o.observeSuccess(p, true, false)  // counts
		o.observeSuccess(p, false, false) // ignored (neither counts nor resets)
	}
	if got := p.conditionSnapshot().kind; got != "" {
		t.Fatalf("raised early — interleaved non-tool calls miscounted: %q", got)
	}
	o.observeSuccess(p, false, false) // still ignored
	o.observeSuccess(p, true, false)  // the threshold-th tool-carrying completion
	if got := p.conditionSnapshot().kind; got != CondToolSilent {
		t.Fatalf("did not raise at threshold with interleaved non-tool calls: %q", got)
	}
}

// TestToolSilentPrecedence (spec 034 data-model.md): a live preflight condition
// (endpoint-unreachable / model-missing) outranks tool-silent — the detector's
// raise at the threshold must not overwrite it. Once the preflight condition is
// gone, continued tool-free tool-carrying successes accumulate and raise.
func TestToolSilentPrecedence(t *testing.T) {
	withToolSilentThreshold(t, 3)
	o := &Orchestrator{}
	p := nativeProvider()

	// A preflight condition holds the slot. The detector's raise — issued exactly
	// as observeSuccess issues it past the threshold — is precedence-dropped. (In
	// the real worker path a tool-free success clears the preflight kind before it
	// counts, so this guard is the concurrent-race safety net: the preflight probe
	// re-raising the condition between a worker's clear and its raise.)
	o.raiseCondition(p, CondEndpointUnreachable, "conn refused", "start the server")
	o.raiseCondition(p, CondToolSilent,
		fmt.Sprintf("%d consecutive tool-free completions on tool-carrying calls", toolSilentThreshold),
		toolSilentRemedy(p.name, ToolModeNative))
	if got := p.conditionSnapshot().kind; got != CondEndpointUnreachable {
		t.Fatalf("detector overwrote a higher-precedence preflight condition: %q", got)
	}

	// Driving observeSuccess against the still-active unreachable condition shows
	// the real path: the FIRST tool-free success clears the preflight kind (the
	// model answered), then the run climbs to the threshold and raises tool-silent.
	for i := 0; i < toolSilentThreshold; i++ {
		o.observeSuccess(p, true, false)
	}
	if got := p.conditionSnapshot().kind; got != CondToolSilent {
		t.Fatalf("detector never raised after the preflight condition cleared: %q", got)
	}
}

// TestToolSilentClearSemantics (spec 034 US2, FR-004): a tool-free success — the
// tool-silent symptom — must NEVER clear tool-silent, though it DOES clear a
// stale preflight condition (the model answered). Only a landed tool call clears
// tool-silent, and it resets the counter.
func TestToolSilentClearSemantics(t *testing.T) {
	withToolSilentThreshold(t, 2)
	o := &Orchestrator{}
	p := nativeProvider()

	// Raise tool-silent.
	for i := 0; i < toolSilentThreshold; i++ {
		o.observeSuccess(p, true, false)
	}
	if p.conditionSnapshot().kind != CondToolSilent {
		t.Fatal("setup: tool-silent not raised")
	}
	// Tool-free successes — tool-carrying OR not — do not clear tool-silent.
	o.observeSuccess(p, true, false)
	o.observeSuccess(p, false, false)
	if got := p.conditionSnapshot().kind; got != CondToolSilent {
		t.Fatalf("a tool-free success cleared tool-silent: %q", got)
	}
	// A landed tool call clears it and resets the run.
	o.observeSuccess(p, true, true)
	if got := p.conditionSnapshot().kind; got != "" {
		t.Fatalf("a landed tool call did not clear tool-silent: %q", got)
	}

	// A tool-free success DOES clear a stale preflight condition.
	o.raiseCondition(p, CondModelMissing, `model "x" not served`, "pull it")
	o.observeSuccess(p, false, false)
	if got := p.conditionSnapshot().kind; got != "" {
		t.Fatalf("a tool-free success did not clear a preflight condition: %q", got)
	}
}

// TestToolSilentRemedyByToolMode (spec 034 US2, contracts/provider-conditions.md):
// the remedy switches on the provider's RESOLVED tool mode — native points at the
// json fallback envelope; json (already the fallback) points at the model itself.
func TestToolSilentRemedyByToolMode(t *testing.T) {
	withToolSilentThreshold(t, 1)
	cases := []struct {
		name       string
		toolMode   string
		wantRemedy string
	}{
		{"native", ToolModeNative, `set providers.local.tool_mode to "json" and restart`},
		{"json", ToolModeJSON, "model never emits tool calls even in json mode — use a model suited for tool work"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			o := &Orchestrator{}
			p := &provider{name: "local", cfg: ProviderConfig{Transport: ProviderOpenAICompat, ToolMode: c.toolMode}}
			o.observeSuccess(p, true, false) // threshold 1 → raises immediately
			if got := p.conditionSnapshot().remedy; got != c.wantRemedy {
				t.Errorf("%s remedy=%q, want %q", c.name, got, c.wantRemedy)
			}
		})
	}
}

// TestToolSilentDetectorFromWorker (spec 034 US2, AC#1/AC#3): end-to-end through
// the real worker path — a tool-carrying planner call whose completions come back
// with zero tool calls raises tool-silent at the threshold; interleaved transport
// failures neither count toward it nor reset the run (the breaker owns transport
// health); and a landed tool call clears it.
func TestToolSilentDetectorFromWorker(t *testing.T) {
	// Keep the circuit closed so the transport-failure path stays the error path,
	// not a fast open-circuit refusal.
	oldFTO := failuresToOpen
	failuresToOpen = 1 << 30
	defer func() { failuresToOpen = oldFTO }()
	withToolSilentThreshold(t, 3)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			http.NotFound(w, r)
			return
		}
		switch prompt := lastUserPrompt(r); {
		case strings.Contains(prompt, "boom"):
			http.Error(w, "boom", http.StatusInternalServerError)
		case strings.Contains(prompt, "landtool"):
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{{
					"message": map[string]any{
						"content": "",
						"tool_calls": []map[string]any{{
							"id": "t1", "type": "function",
							"function": map[string]any{"name": "noop", "arguments": "{}"},
						}},
					},
					"finish_reason": "tool_calls",
				}},
				"usage": map[string]any{"prompt_tokens": 1, "completion_tokens": 1},
			})
		default: // a tool-free completion (content only, no tool_calls)
			localReplyJSON(w)
		}
	}))
	defer srv.Close()

	o := newOrch(t, testConfig(srv.URL, "http://unused.invalid", 100), testStore(t))
	local := o.providers["local"]
	tools := []ToolDecl{{Name: "noop", InputSchema: json.RawMessage("{}")}}
	submit := func(prompt string) error {
		_, err := o.Submit(context.Background(), Request{Kind: KindPlanner, Prompt: prompt, Tools: tools})
		return err
	}

	// One short of the threshold: no condition yet.
	for i := 0; i < toolSilentThreshold-1; i++ {
		if err := submit("silent"); err != nil {
			t.Fatalf("silent call %d: %v", i, err)
		}
	}
	if got := local.conditionSnapshot().kind; got != "" {
		t.Fatalf("raised before threshold: %q", got)
	}

	// Transport failures interleave: neither counting nor resetting the run.
	for i := 0; i < 3; i++ {
		if err := submit("boom"); err == nil {
			t.Fatal("a 500 response must surface as an error")
		}
	}
	if got := local.conditionSnapshot().kind; got != "" {
		t.Fatalf("a transport failure raised a condition: %q", got)
	}

	// The threshold-th tool-free completion (proving the failures did not reset
	// the run) raises tool-silent.
	if err := submit("silent"); err != nil {
		t.Fatal(err)
	}
	if got := local.conditionSnapshot().kind; got != CondToolSilent {
		t.Fatalf("detector not wired from the worker: slot=%q, want tool-silent", got)
	}

	// A landed tool call clears it.
	if err := submit("landtool"); err != nil {
		t.Fatal(err)
	}
	if got := local.conditionSnapshot().kind; got != "" {
		t.Fatalf("a landed tool call did not clear tool-silent from the worker: %q", got)
	}
}
