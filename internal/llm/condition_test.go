package llm

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// condFire records one condition-hook invocation for transition assertions.
type condFire struct {
	provider string
	kind     string
	detail   string
	remedy   string
	active   bool
}

// condRecorder is a thread-safe SetConditionHook target: conditions can be
// raised from worker/preflight goroutines, so the recorder guards its slice.
type condRecorder struct {
	mu    sync.Mutex
	fires []condFire
}

func (r *condRecorder) hook(provider, kind, detail, remedy string, active bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fires = append(r.fires, condFire{provider, kind, detail, remedy, active})
}

func (r *condRecorder) snapshot() []condFire {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]condFire(nil), r.fires...)
}

// TestConditionPrecedence exercises the one-slot precedence rule (spec 034
// data-model.md): endpoint-unreachable > model-missing > tool-silent. A raise
// whose kind ranks below the active condition is dropped; an equal-or-higher
// kind replaces (reclassify). Each genuine transition fires the hook exactly
// once with active=true; a dropped raise fires nothing.
func TestConditionPrecedence(t *testing.T) {
	var rec condRecorder
	o := &Orchestrator{}
	o.SetConditionHook(rec.hook)
	p := &provider{name: "local"}

	// Healthy → model-missing: a genuine raise.
	o.raiseCondition(p, CondModelMissing, "model absent", "pull it")
	if got := p.conditionSnapshot().kind; got != CondModelMissing {
		t.Fatalf("after model-missing raise, slot = %q, want model-missing", got)
	}

	// tool-silent while model-missing holds: lower precedence → no-op, no fire.
	o.raiseCondition(p, CondToolSilent, "8 tool-free", "switch mode")
	if got := p.conditionSnapshot().kind; got != CondModelMissing {
		t.Fatalf("lower-precedence raise overwrote the slot: got %q, want model-missing", got)
	}

	// endpoint-unreachable while model-missing holds: higher precedence →
	// reclassify.
	o.raiseCondition(p, CondEndpointUnreachable, "conn refused", "start server")
	if got := p.conditionSnapshot().kind; got != CondEndpointUnreachable {
		t.Fatalf("higher-precedence raise did not reclassify: got %q, want endpoint-unreachable", got)
	}

	// model-missing while unreachable holds: lower precedence → no-op.
	o.raiseCondition(p, CondModelMissing, "model absent again", "pull it")
	if got := p.conditionSnapshot().kind; got != CondEndpointUnreachable {
		t.Fatalf("lower-precedence raise displaced unreachable: got %q", got)
	}

	// Clear returns to healthy.
	o.clearCondition(p)
	if got := p.conditionSnapshot().kind; got != "" {
		t.Fatalf("after clear, slot = %q, want healthy (empty)", got)
	}

	// Exactly three transitions fired: raise(model-missing), reclassify(unreachable), clear.
	fires := rec.snapshot()
	want := []condFire{
		{"local", "model-missing", "model absent", "pull it", true},
		{"local", "endpoint-unreachable", "conn refused", "start server", true},
		{"local", "endpoint-unreachable", "conn refused", "", false},
	}
	if len(fires) != len(want) {
		t.Fatalf("hook fired %d times, want %d: %+v", len(fires), len(want), fires)
	}
	for i := range want {
		if fires[i] != want[i] {
			t.Errorf("fire[%d] = %+v, want %+v", i, fires[i], want[i])
		}
	}
}

// TestConditionTransitionsOnly proves the hook fires on transitions only —
// never on a repeated identical raise, and never on a clear of an already
// healthy provider (spec 034: transitions only, so the durable log stays quiet).
func TestConditionTransitionsOnly(t *testing.T) {
	var rec condRecorder
	o := &Orchestrator{}
	o.SetConditionHook(rec.hook)
	p := &provider{name: "local"}

	// First raise: one fire.
	o.raiseCondition(p, CondToolSilent, "8 consecutive tool-free completions", "set json")
	// Identical raise (same kind AND detail): not a transition — no fire.
	o.raiseCondition(p, CondToolSilent, "8 consecutive tool-free completions", "set json")
	o.raiseCondition(p, CondToolSilent, "8 consecutive tool-free completions", "set json")
	// Same kind, new detail: a reclassify — fires.
	o.raiseCondition(p, CondToolSilent, "12 consecutive tool-free completions", "set json")
	// Clear: fires once.
	o.clearCondition(p)
	// Clear again while healthy: no fire.
	o.clearCondition(p)

	fires := rec.snapshot()
	want := []condFire{
		{"local", "tool-silent", "8 consecutive tool-free completions", "set json", true},
		{"local", "tool-silent", "12 consecutive tool-free completions", "set json", true},
		{"local", "tool-silent", "12 consecutive tool-free completions", "", false},
	}
	if len(fires) != len(want) {
		t.Fatalf("hook fired %d times, want %d (transitions only): %+v", len(fires), len(want), fires)
	}
	for i := range want {
		if fires[i] != want[i] {
			t.Errorf("fire[%d] = %+v, want %+v", i, fires[i], want[i])
		}
	}
}

// TestConditionClearOnSuccess proves the clear-on-success path (spec 034
// FR-004, "traffic is truth"): a completed provider call clears any active
// condition and fires the hook with active=false. Driven through the fake
// caller (the package's httptest OpenAI-compat server) and a provider pin so
// the served provider is deterministic.
func TestConditionClearOnSuccess(t *testing.T) {
	t.Setenv("PROMPTWORLD_TEST_KEY", "test-key")
	var lh, ch atomic.Int64
	local := mockLocal(t, &lh)
	cloud := mockCloud(t, &ch)
	o, err := New(testConfig(local.URL, cloud.URL, 100), testStore(t))
	if err != nil {
		t.Fatal(err)
	}
	defer o.Close()

	var rec condRecorder
	o.SetConditionHook(rec.hook)

	// Raise a preflight-class condition on local (fires active=true).
	o.raiseCondition(o.providers["local"], CondModelMissing,
		`model "test-local" not served`, "pull it")
	if got := o.providers["local"].conditionSnapshot().kind; got != CondModelMissing {
		t.Fatalf("setup: local slot = %q, want model-missing", got)
	}

	// A successful call on local clears the condition (worker success path).
	if _, err := o.Submit(context.Background(),
		Request{Kind: KindPlanner, Provider: "local", Prompt: "x"}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	if lh.Load() == 0 {
		t.Fatal("mock local never served the call")
	}
	if got := o.providers["local"].conditionSnapshot().kind; got != "" {
		t.Fatalf("condition did not clear on success: slot = %q", got)
	}

	fires := rec.snapshot()
	if len(fires) != 2 {
		t.Fatalf("want 2 fires (raise + clear), got %d: %+v", len(fires), fires)
	}
	if last := fires[len(fires)-1]; last.active || last.provider != "local" || last.remedy != "" {
		t.Errorf("clear fire = %+v, want {provider:local active:false remedy:empty}", last)
	}
}

// TestStatusSnapshotConditionExport proves the wire export (spec 034): a healthy
// provider serializes byte-identically to a pre-034 world (all three condition
// fields omitempty-dropped), and an active condition surfaces all three.
func TestStatusSnapshotConditionExport(t *testing.T) {
	t.Setenv("PROMPTWORLD_TEST_KEY", "test-key")
	var lh, ch atomic.Int64
	local := mockLocal(t, &lh)
	cloud := mockCloud(t, &ch)
	o, err := New(testConfig(local.URL, cloud.URL, 100), testStore(t))
	if err != nil {
		t.Fatal(err)
	}
	defer o.Close()

	// Healthy world: no condition keys anywhere in the marshaled status.
	healthy, err := json.Marshal(o.StatusSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"condition", "condition_detail", "condition_remedy"} {
		if strings.Contains(string(healthy), key) {
			t.Errorf("healthy status must not emit %q, got: %s", key, healthy)
		}
	}

	// Raise a condition on local: its row carries the three fields; cloud (still
	// healthy) leaves them empty.
	o.raiseCondition(o.providers["local"], CondModelMissing,
		`model "test-local" not served by `+local.URL, "ollama pull test-local")

	snap := o.StatusSnapshot()
	localRow := provStatus(snap, "local")
	if localRow.Condition != string(CondModelMissing) {
		t.Errorf("local Condition = %q, want model-missing", localRow.Condition)
	}
	if !strings.Contains(localRow.ConditionDetail, "not served") {
		t.Errorf("local ConditionDetail = %q, want the evidence text", localRow.ConditionDetail)
	}
	if localRow.ConditionRemedy != "ollama pull test-local" {
		t.Errorf("local ConditionRemedy = %q, want the pull command", localRow.ConditionRemedy)
	}
	if cloudRow := provStatus(snap, "cloud"); cloudRow.Condition != "" {
		t.Errorf("healthy cloud row leaked a condition: %q", cloudRow.Condition)
	}

	// The active row marshals all three keys.
	raised, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"condition":"model-missing"`, `"condition_detail":`, `"condition_remedy":`} {
		if !strings.Contains(string(raised), key) {
			t.Errorf("active status missing %q, got: %s", key, raised)
		}
	}
}
