package mind

// Truncation-aware submit ladder (spec 105, TASK-172): the playtest-1 failure
// was a night's digest outgrowing its response budget — the cut JSON parsed as
// "unparseable" and the night was rejected, silently, every night from day 20
// on. This helper wraps a single-shot Submit in a parse-first retry loop:
// detect a cut reply mechanically at the transport level, re-submit the SAME
// prompt with a doubled budget clamped at the shared ceiling, and report what
// happened so the caller can record it. Consolidation and the narrator
// (chapter + morgue epilogue) drive it; the tool-loop and conversation sites
// are untouched.

import (
	"context"
	"time"

	"github.com/evanstern/promptworld/internal/llm"
)

// maxTruncationRetries bounds the ladder: at most 2 retries (3 attempts),
// doubling the budget each time. Compiled-in doctrine, not an llm.json knob —
// the operator's existing per-kind budget moves the ladder's START and
// llm.MaxTokenBudget bounds its END (spec 105 assumptions).
const maxTruncationRetries = 2

// truncResult is one ladder run's outcome. Resp is the FINAL attempt's
// response; CostUSD accrues across every attempt (each attempt is a real,
// billed call). Retries counts consumed truncation retries (0 on a clean
// first attempt). When ParseErr is non-nil the run failed: Truncated says
// whether the terminal failure was still a detected truncation (ladder
// exhausted or no headroom) as opposed to a genuine garbage reply.
type truncResult struct {
	Resp      llm.Response
	Retries   int
	CostUSD   float64
	ParseErr  error
	Truncated bool
}

// truncated is the mechanical detection (FR-001), consulted ONLY after a
// parse failure: the provider said so (StopMaxTokens, primary), or — for
// OpenAI-compatible routers that surface an unmapped finish reason — the
// reply's output tokens reached the attempt's requested budget (the router
// honesty guard). No text heuristics. A reply that parses is never consulted:
// a complete JSON object that hit max_tokens on trailing junk is still a
// complete object.
func truncated(resp llm.Response, requested int64) bool {
	return resp.Stop == llm.StopMaxTokens ||
		(requested > 0 && resp.OutputTokens >= requested)
}

// submitWithTruncationRetry submits req and, when the reply fails the
// caller's parse AND is detected truncated, re-submits the SAME request with
// the budget doubled from the attempt's value, clamped at llm.MaxTokenBudget,
// at most maxTruncationRetries times. Each ATTEMPT gets its own timeout (the
// callers' existing per-call bound). onRetry, if non-nil, is invoked before
// each re-submission with the 1-based retry ordinal, the budget that
// truncated, and the escalated budget — the caller's telemetry hook.
//
// A transport error on ANY attempt returns it immediately: the callers'
// existing transport semantics (defer / carry, no marker) apply unchanged. A
// parse success returns at once — a budget retry is never triggered by
// content judgment (FR-001). No headroom (budget already at the clamp) means
// the loop cannot escalate and the run terminates truncated after that
// attempt.
func (md *Mind) submitWithTruncationRetry(timeout time.Duration, req llm.Request,
	parse func(text string) error, onRetry func(retry int, from, to int64)) (truncResult, error) {
	var res truncResult
	for {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		resp, err := md.orch.Submit(ctx, req)
		cancel()
		if err != nil {
			return res, err
		}
		res.Resp = resp
		res.CostUSD += resp.CostUSD
		res.ParseErr = parse(resp.Text)
		if res.ParseErr == nil {
			res.Truncated = false
			return res, nil
		}
		if !truncated(resp, req.MaxTokens) {
			res.Truncated = false
			return res, nil
		}
		res.Truncated = true
		next := req.MaxTokens * 2
		if next > llm.MaxTokenBudget {
			next = llm.MaxTokenBudget
		}
		// Ladder exhausted: out of retries, or no headroom left to escalate
		// into (a start already at the clamp yields a single attempt).
		if res.Retries >= maxTruncationRetries || next <= req.MaxTokens {
			return res, nil
		}
		res.Retries++
		if onRetry != nil {
			onRetry(res.Retries, req.MaxTokens, next)
		}
		req.MaxTokens = next
	}
}
