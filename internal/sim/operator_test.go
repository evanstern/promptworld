package sim

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/evanstern/promptworld/internal/store"
)

// newOperatorHarness builds a Loop over a real store whose goroutine is NOT
// running, so inject_operator's whitelist + tick-stamping are exercised straight
// through handleCommand — no timers, no ticks. The state sits at the given tick.
func newOperatorHarness(t *testing.T, tick int64) *Loop {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "world.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	m := testMap(7)
	s := NewState(7, m)
	s.Tick = tick
	return NewLoop(s, m, st, nil)
}

// llmWarningEvent builds a daemon.llm_warning store event with a deliberately
// wrong tick, so a passing test proves the loop re-stamps it.
func llmWarningEvent(t *testing.T, wrongTick int64) store.Event {
	t.Helper()
	payload, err := json.Marshal(LLMWarningPayload{Provider: "local", Kind: "model-missing", Detail: "absent", Active: true})
	if err != nil {
		t.Fatal(err)
	}
	return store.Event{Tick: wrongTick, Type: "daemon.llm_warning", Payload: payload}
}

// TestInjectOperatorStampsTick proves the operator door lands a whitelisted
// daemon event and re-stamps its tick from the loop state (single-writer
// authority over the log position), leaving world state untouched (reducer
// no-op).
func TestInjectOperatorStampsTick(t *testing.T) {
	l := newOperatorHarness(t, 4242)
	ev := llmWarningEvent(t, 999) // wrong tick — the loop must overwrite it

	cmd := command{name: "inject_operator", social: []store.Event{ev}, reply: make(chan commandResult, 1)}
	if err := l.handleCommand(cmd); err != nil {
		t.Fatalf("handleCommand(inject_operator): %v", err)
	}
	if res := <-cmd.reply; res.err != nil {
		t.Fatalf("inject_operator returned err: %v", res.err)
	}

	all, err := l.st.EventsSince(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	var landed *store.Event
	for i := range all {
		if all[i].Type == "daemon.llm_warning" {
			landed = &all[i]
		}
	}
	if landed == nil {
		t.Fatal("daemon.llm_warning never landed in the store")
	}
	if landed.Tick != 4242 {
		t.Errorf("landed tick = %d, want the loop's 4242 (re-stamped)", landed.Tick)
	}
}

// TestInjectOperatorRejectsNonWhitelisted proves the door is a strict boundary:
// an event type outside injectOperatorWhitelist is rejected with an error and
// NOTHING is appended (all-or-nothing).
func TestInjectOperatorRejectsNonWhitelisted(t *testing.T) {
	l := newOperatorHarness(t, 10)
	// A social type is valid at the MIND's door but not the operator's — the two
	// whitelists are deliberately disjoint.
	bad := store.Event{Type: "agent.thought", Payload: json.RawMessage(`{}`)}

	cmd := command{name: "inject_operator", social: []store.Event{bad}, reply: make(chan commandResult, 1)}
	if err := l.handleCommand(cmd); err != nil {
		t.Fatalf("handleCommand should surface the rejection via reply, not return: %v", err)
	}
	res := <-cmd.reply
	if res.err == nil {
		t.Fatal("non-whitelisted operator event was accepted, want rejection")
	}

	all, err := l.st.EventsSince(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range all {
		if e.Type == "agent.thought" {
			t.Fatal("a rejected operator batch still appended an event")
		}
	}
}

// TestInjectOperatorEmptyRejected proves an empty batch is a clean error, never
// a silent success.
func TestInjectOperatorEmptyRejected(t *testing.T) {
	l := newOperatorHarness(t, 1)
	cmd := command{name: "inject_operator", social: nil, reply: make(chan commandResult, 1)}
	if err := l.handleCommand(cmd); err != nil {
		t.Fatalf("handleCommand: %v", err)
	}
	if res := <-cmd.reply; res.err == nil {
		t.Fatal("empty operator batch was accepted, want an error")
	}
}

// TestInjectOperatorErrorsAfterLoopStop proves the shutdown-window degrade path
// (spec 034 T008b): once the loop goroutine has stopped, InjectOperator returns
// an error rather than blocking forever — the daemon hook falls back to the log
// line only.
func TestInjectOperatorErrorsAfterLoopStop(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "world.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	m := testMap(7)
	s := NewState(7, m)
	s.Paused = true // idle in the command/ctx select, no busy ticking
	l := NewLoop(s, m, st, nil)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { _ = l.Run(ctx); close(done) }()
	cancel()
	<-done // the loop has returned and closed l.done

	ev := llmWarningEvent(t, 0)
	deadline := time.Now().Add(2 * time.Second)
	for {
		err := l.InjectOperator([]store.Event{ev})
		if err != nil {
			return // expected: the stopped loop refuses the injection
		}
		if time.Now().After(deadline) {
			t.Fatal("InjectOperator never errored after the loop stopped")
		}
		time.Sleep(time.Millisecond)
	}
}
