package mind

// Per-night consolidation acceptance observability (spec 105, US2): the
// playtest-1 blackout — ten straight nights of 0/8 accepted consolidations —
// was invisible until a post-mortem read the event log. The mind now keeps
// per-night outcome counters fed from live-absorbed agent.consolidated
// markers and, when it observes a night has closed (any tick or marker from a
// later night), logs ONE summary line for it; two or more CONSECUTIVE nights
// with at least one attempted (non-empty-skip) consolidation and zero
// acceptances escalate the line to a WARNING: naming the streak and the
// remedy. In-memory only, absorb-goroutine-owned: no timer, no new event
// type, no state in the replica. Boot is silent by construction — the
// replica is seeded from a state snapshot, never a log replay through
// absorb, so the counters only ever see live nights. A mid-blackout restart
// therefore delays the WARNING by up to the threshold (accepted, spec 105
// edge cases): reducing a streak into world state for a log line was
// rejected as reducer/state churn.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/evanstern/promptworld/internal/sim"
)

// nightWarnThreshold is the consecutive fully-failed-night count at which the
// summary escalates to a WARNING: line (FR-007). Compiled-in doctrine, like
// the ladder dials.
const nightWarnThreshold = 2

// nightCounts is one night's outcome tally.
type nightCounts struct {
	accepted int
	rejected map[string]int // by marker reason ("" keyed as "unknown")
	skipped  int            // skipped_empty markers (not attempts)
}

func (c *nightCounts) attempted() int {
	n := c.accepted
	for _, v := range c.rejected {
		n += v
	}
	return n
}

// nightReport accumulates marker outcomes per night index and flushes each
// night's summary once a later night is observed. Owned by the absorb
// goroutine — never locked, never touched by workers.
type nightReport struct {
	nights map[int64]*nightCounts
	streak int // consecutive flushed nights with ≥1 attempt and 0 acceptances
}

// record ingests one live-absorbed agent.consolidated marker.
func (r *nightReport) record(p sim.ConsolidatedPayload) {
	if p.Night <= 0 {
		return
	}
	if r.nights == nil {
		r.nights = map[int64]*nightCounts{}
	}
	c := r.nights[p.Night]
	if c == nil {
		c = &nightCounts{rejected: map[string]int{}}
		r.nights[p.Night] = c
	}
	switch p.Outcome {
	case sim.ConsolidationAccepted:
		c.accepted++
	case sim.ConsolidationRejected:
		reason := p.Reason
		if reason == "" {
			reason = "unknown"
		}
		c.rejected[reason]++
	case sim.ConsolidationSkippedEmpty:
		c.skipped++
	}
}

// flushBefore closes every tracked night strictly before the given night
// index, in order, returning one summary log line per closed night (FR-006)
// — escalated to a WARNING: line while the consecutive fully-failed streak is
// at or past the threshold (FR-007, repeating each further failed night until
// a night accepts). A night with zero attempts (all empty skips) neither
// extends nor resets the streak: FR-007 counts attempted nights.
func (r *nightReport) flushBefore(night int64) []string {
	if len(r.nights) == 0 {
		return nil
	}
	var due []int64
	for n := range r.nights {
		if n < night {
			due = append(due, n)
		}
	}
	sort.Slice(due, func(i, j int) bool { return due[i] < due[j] })
	var lines []string
	for _, n := range due {
		c := r.nights[n]
		delete(r.nights, n)
		attempted := c.attempted()
		switch {
		case attempted > 0 && c.accepted == 0:
			r.streak++
		case c.accepted > 0:
			r.streak = 0
		}
		line := fmt.Sprintf("mind: consolidation night %d: %d accepted, %d rejected%s, %d skipped-empty",
			n, c.accepted, attempted-c.accepted, rejectedBreakdown(c.rejected), c.skipped)
		if attempted > 0 && c.accepted == 0 && r.streak >= nightWarnThreshold {
			line = fmt.Sprintf("mind: WARNING: consolidation has failed every attempt for %d consecutive nights"+
				" (night %d: 0 of %d accepted) — raise llm.json max_tokens.consolidation or check the serving provider",
				r.streak, n, attempted)
		}
		lines = append(lines, line)
	}
	return lines
}

// rejectedBreakdown renders the by-reason split, e.g. " (truncated 3,
// unparseable 1)"; empty when nothing was rejected. Reasons sort
// lexicographically so the line is deterministic.
func rejectedBreakdown(rejected map[string]int) string {
	if len(rejected) == 0 {
		return ""
	}
	reasons := make([]string, 0, len(rejected))
	for reason := range rejected {
		reasons = append(reasons, reason)
	}
	sort.Strings(reasons)
	parts := make([]string, 0, len(reasons))
	for _, reason := range reasons {
		parts = append(parts, fmt.Sprintf("%s %d", reason, rejected[reason]))
	}
	return " (" + strings.Join(parts, ", ") + ")"
}
