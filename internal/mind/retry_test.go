package mind

// Spec 105 (TASK-172) T002: the truncation-aware submit ladder in isolation —
// the FR-001 detection matrix and the FR-002 doubling arithmetic, against a
// scripted submitter that records each attempt's requested budget.

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/evanstern/promptworld/internal/llm"
)

// respModel returns queued full Responses in order (Stop, OutputTokens,
// CostUSD included), recording each request's MaxTokens; exhausted → error.
type respModel struct {
	mu        sync.Mutex
	replies   []llm.Response
	requested []int64
}

func (m *respModel) Submit(_ context.Context, req llm.Request) (llm.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requested = append(m.requested, req.MaxTokens)
	if len(m.requested) > len(m.replies) {
		return llm.Response{}, context.DeadlineExceeded
	}
	return m.replies[len(m.requested)-1], nil
}

// parseWant fails until it sees the given text.
func parseWant(want string) func(string) error {
	return func(text string) error {
		if text == want {
			return nil
		}
		return fmt.Errorf("cut mid-field")
	}
}

func ladderMind(m Submitter) *Mind { return &Mind{orch: m} }

// TestTruncRetryStopMaxTokens: the primary signal — parse failure +
// StopMaxTokens retries with a doubled budget and an accepted retry returns
// clean, cost accrued across both attempts.
func TestTruncRetryStopMaxTokens(t *testing.T) {
	model := &respModel{replies: []llm.Response{
		{Text: `{"cut`, Stop: llm.StopMaxTokens, OutputTokens: 1024, CostUSD: 0.01},
		{Text: `ok`, Stop: llm.StopEndTurn, OutputTokens: 1500, CostUSD: 0.02},
	}}
	var hook [][2]int64
	res, err := ladderMind(model).submitWithTruncationRetry(time.Second,
		llm.Request{MaxTokens: 1024}, parseWant("ok"),
		func(retry int, from, to int64) { hook = append(hook, [2]int64{from, to}) })
	if err != nil {
		t.Fatal(err)
	}
	if res.ParseErr != nil || res.Truncated || res.Retries != 1 {
		t.Errorf("res = %+v, want parsed, retries 1", res)
	}
	if got := res.CostUSD; got != 0.03 {
		t.Errorf("CostUSD = %v, want 0.03 (accrued across attempts)", got)
	}
	if want := []int64{1024, 2048}; len(model.requested) != 2 ||
		model.requested[0] != want[0] || model.requested[1] != want[1] {
		t.Errorf("requested budgets = %v, want %v", model.requested, want)
	}
	if len(hook) != 1 || hook[0] != [2]int64{1024, 2048} {
		t.Errorf("onRetry saw %v, want [[1024 2048]]", hook)
	}
}

// TestTruncRetryOutputTokensGuard: the router honesty guard — an unmapped
// stop reason (StopOther) with OutputTokens at the requested budget still
// counts as truncation on a parse failure.
func TestTruncRetryOutputTokensGuard(t *testing.T) {
	model := &respModel{replies: []llm.Response{
		{Text: `{"cut`, Stop: llm.StopOther, OutputTokens: 1024},
		{Text: `ok`, Stop: llm.StopOther, OutputTokens: 900},
	}}
	res, err := ladderMind(model).submitWithTruncationRetry(time.Second,
		llm.Request{MaxTokens: 1024}, parseWant("ok"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.ParseErr != nil || res.Retries != 1 {
		t.Errorf("res = %+v, want accepted retry via output-tokens guard", res)
	}
}

// TestTruncRetryParseSuccessNeverRetries: FR-001 — a reply that parses is
// judged on content only; even a max_tokens stop consumes no retry (the
// object completed before the cut).
func TestTruncRetryParseSuccessNeverRetries(t *testing.T) {
	model := &respModel{replies: []llm.Response{
		{Text: `ok`, Stop: llm.StopMaxTokens, OutputTokens: 1024},
	}}
	res, err := ladderMind(model).submitWithTruncationRetry(time.Second,
		llm.Request{MaxTokens: 1024}, parseWant("ok"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Retries != 0 || res.Truncated || res.ParseErr != nil {
		t.Errorf("res = %+v, want single clean attempt", res)
	}
	if len(model.requested) != 1 {
		t.Errorf("attempts = %d, want 1", len(model.requested))
	}
}

// TestTruncRetryGarbageNeverRetries: a parse failure WITHOUT a truncation
// signal is a content failure, not a budget failure — no retry, not truncated.
func TestTruncRetryGarbageNeverRetries(t *testing.T) {
	model := &respModel{replies: []llm.Response{
		{Text: `humming, no json`, Stop: llm.StopEndTurn, OutputTokens: 40},
	}}
	res, err := ladderMind(model).submitWithTruncationRetry(time.Second,
		llm.Request{MaxTokens: 1024}, parseWant("ok"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.ParseErr == nil || res.Truncated || res.Retries != 0 {
		t.Errorf("res = %+v, want unretried non-truncation parse failure", res)
	}
}

// TestTruncRetryLadderExhausted: 1024→2048→4096, still truncated at the
// clamp — 3 attempts, 2 consumed retries, terminal Truncated=true, cost
// accrued across all three.
func TestTruncRetryLadderExhausted(t *testing.T) {
	cut := func(tok int64) llm.Response {
		return llm.Response{Text: `{"cut`, Stop: llm.StopMaxTokens, OutputTokens: tok, CostUSD: 0.01}
	}
	model := &respModel{replies: []llm.Response{cut(1024), cut(2048), cut(4096)}}
	res, err := ladderMind(model).submitWithTruncationRetry(time.Second,
		llm.Request{MaxTokens: 1024}, parseWant("ok"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.ParseErr == nil || !res.Truncated || res.Retries != 2 {
		t.Errorf("res = %+v, want exhausted-truncated with 2 retries", res)
	}
	if want := []int64{1024, 2048, 4096}; fmt.Sprint(model.requested) != fmt.Sprint(want) {
		t.Errorf("requested budgets = %v, want %v", model.requested, want)
	}
	if res.CostUSD != 0.03 {
		t.Errorf("CostUSD = %v, want 0.03", res.CostUSD)
	}
}

// TestTruncRetryNoHeadroom: a start budget already at the clamp yields a
// single attempt — a truncated reply terminates immediately (loud, FR-002).
func TestTruncRetryNoHeadroom(t *testing.T) {
	model := &respModel{replies: []llm.Response{
		{Text: `{"cut`, Stop: llm.StopMaxTokens, OutputTokens: 4096},
	}}
	res, err := ladderMind(model).submitWithTruncationRetry(time.Second,
		llm.Request{MaxTokens: llm.MaxTokenBudget}, parseWant("ok"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Truncated || res.Retries != 0 || len(model.requested) != 1 {
		t.Errorf("res = %+v (attempts %d), want single truncated attempt", res, len(model.requested))
	}
}

// TestTruncRetryOddStartClampsThenStops: a start whose double overshoots the
// clamp escalates once TO the clamp, then cannot escalate further — the
// second retry would not raise the budget, so the ladder stops there.
func TestTruncRetryOddStartClampsThenStops(t *testing.T) {
	cut := llm.Response{Text: `{"cut`, Stop: llm.StopMaxTokens}
	model := &respModel{replies: []llm.Response{cut, cut}}
	res, err := ladderMind(model).submitWithTruncationRetry(time.Second,
		llm.Request{MaxTokens: 3000}, parseWant("ok"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if want := []int64{3000, 4096}; fmt.Sprint(model.requested) != fmt.Sprint(want) {
		t.Errorf("requested budgets = %v, want %v", model.requested, want)
	}
	if !res.Truncated || res.Retries != 1 {
		t.Errorf("res = %+v, want truncated after one clamped retry", res)
	}
}

// TestTruncRetryTransportErrorMidLadder: a transport failure on any attempt
// surfaces as the Submit error — the caller's existing defer/carry semantics
// own it (FR-009); no result is fabricated.
func TestTruncRetryTransportErrorMidLadder(t *testing.T) {
	model := &respModel{replies: []llm.Response{
		{Text: `{"cut`, Stop: llm.StopMaxTokens, OutputTokens: 1024},
		// second attempt: queue exhausted → transport error
	}}
	_, err := ladderMind(model).submitWithTruncationRetry(time.Second,
		llm.Request{MaxTokens: 1024}, parseWant("ok"), nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want the transport error surfaced", err)
	}
	if len(model.requested) != 2 {
		t.Errorf("attempts = %d, want 2", len(model.requested))
	}
}
