# Feature Specification: Consolidation truncation-aware retry + acceptance observability

**Feature Branch**: `task-172-consolidation-truncation`

**Created**: 2026-07-30

**Status**: Draft

**Input**: TASK-172 — nightly memory consolidation silently collapses as worlds age.
Playtest-1 evidence (29 game-days, consolidation on cloud Sonnet): acceptance degrades
monotonically — night 2: 7/9 accepted; night 11: 3/13; night 17: 1/15; nights 20–29:
0/8 EVERY night (ten straight nights, all 8 villagers, silently). The logged invalid
sample is JSON cut mid-field; day-29 narration also died with "unterminated JSON
object" (same failure family). Root cause live on main:
`internal/mind/consolidate.go:166` sends `md.consolidationTokens` (llm.json
`max_tokens.consolidation`, spec 025 US2, default 1024) with no truncation-aware
retry — a truncated reply parses as "unparseable" and lands a rejected marker,
night after night. Only 232 `agent.consolidated` events landed all run.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A 30-day-old villager still consolidates (Priority: P1)

As a player, I want my villagers to keep remembering and growing over a long run —
when a night's digest outgrows its output budget, the mind should notice the cut and
retry with a bigger budget, so a 30-day world feels like week one but deeper, not
like villagers who stopped forming long-term memories on day 12.

**Acceptance Scenarios**:

1. **Given** a consolidation call whose reply comes back truncated (provider stop
   reason `max_tokens`, or an unparseable reply whose output tokens reached the
   requested budget), **When** the worker processes it, **Then** the SAME prompt is
   re-submitted with a doubled response budget (up to the shared 4096 clamp, at most
   2 retries) and an accepted retry lands the night normally, with the consumed
   retries visible on the night's `agent.consolidated` marker.
2. **Given** a late-world villager (episodic buffer larger than `maxBufferSent`,
   a dozen-plus held beliefs including faded ones — the shape that overflowed 1024
   in playtest-1), **When** the scripted model needs more than the configured budget
   to emit the full JSON object, **Then** the night is accepted on a retry, not
   rejected "unparseable".
3. **Given** the ladder is exhausted (still truncated at the 4096 clamp, or the
   configured budget already sits at the clamp), **Then** the night lands a rejected
   marker with the distinct reason `truncated` (never the misleading
   `unparseable`), the buffer stays intact, and the next sleep retries.
4. **Given** a reply that parses and validates on the first attempt, **Then**
   behavior is byte-identical to today — no retry is consumed even if the stop
   reason is `max_tokens` (the object completed before the cut).

---

### User Story 2 - The operator sees the blackout (Priority: P1)

As an operator, I want a loud signal when a background cognitive pipeline starts
failing every night — per-night acceptance in the daemon log, and an escalating
WARNING when every villager's consolidation has failed for consecutive nights —
instead of discovering a ten-day blackout after the fact.

**Acceptance Scenarios**:

1. **Given** a night in which consolidations landed markers, **When** the night
   closes (the mind observes a later night/day), **Then** one daemon-log summary
   line reports that night's acceptance: accepted / rejected (by reason) /
   skipped-empty counts.
2. **Given** two or more CONSECUTIVE nights in which at least one consolidation was
   attempted and none was accepted, **Then** the summary escalates to a
   `WARNING:`-prefixed line naming the streak length and a remedy (raise
   `max_tokens.consolidation`, check the serving provider).
3. **Given** each consumed truncation retry, **Then** a `cog.outcome` record with
   the existing `retried` outcome (class `consolidation`, reason naming the
   truncation and the escalated budget) is injected — the existing telemetry
   vocabulary, no new event type.

---

### User Story 3 - The chronicle survives the same cut (Priority: P2)

As a player reading the story feed, I want day-29's narration to survive a truncated
reply the same way — the narrator call is the sibling failure site (fixed 800-token
budget, same "unterminated JSON object" death), so the retry helper generalizes to
it.

**Acceptance Scenarios**:

1. **Given** a chapter or morgue-epilogue narration whose reply is detected
   truncated, **When** the worker processes it, **Then** it is retried with the same
   doubling ladder (from 800, clamped at 4096) before the existing failure handling
   (carry lines / gap in the story) applies.

---

### Edge Cases

- **Parse-first discipline**: truncation is only consulted AFTER a parse failure. A
  reply that parses is judged on content exactly as today (validator rejection is a
  content judgment, never retried for budget — a complete JSON object that hit
  `max_tokens` on trailing junk is still a complete object).
- **Router honesty guard**: some OpenAI-compatible routers surface an unmapped
  finish reason (`StopOther`). The secondary signal — parse failure AND
  `OutputTokens >= the attempt's requested budget` — catches truncation there
  mechanically, without text heuristics.
- **No headroom**: configured budget already at the 4096 clamp → single attempt;
  a truncated reply lands `truncated` immediately (loud, actionable).
- **Unchanged semantics** (all byte-identical to today): transport/tier failure
  still lands NO marker (defer, next sleep retries); the router gate is consulted
  once per night, never re-consulted mid-ladder; empty buffers still close with
  `skipped_empty`; a parse failure that is NOT truncation still rejects
  `unparseable`.
- **Retries reuse the job snapshot**: the spec-098 dream geometry pass ran once and
  its batch already landed — a retry re-sends the SAME prompt (same ordinal labels,
  same `[gN]` groups), never re-plans dreams and never re-snapshots the replica.
- **Cost accrual**: every attempt's cost accrues into the night's marker `CostUSD`
  and the spend meter as normal; worst case is ~3× one call's input tokens on a
  truncated night — negligible against the ceiling, $0 on a LAN router.
- **Timeouts**: `consolidateCallTimeout` (3 min) bounds each ATTEMPT; worst case
  ~9 min per agent through the single-flight worker — the night is hours long.
- **Restart amnesia**: the nightly counters and the consecutive-failure streak are
  mind-side in-memory (rebuilt from live absorb only; the replica is seeded from a
  state snapshot, not a log replay). A mid-blackout restart delays the WARNING by
  up to the threshold — accepted; reducing a streak into world state for a log line
  is rejected (reducer/state churn, replay surface).
- **Replay/versioning (spec 094 doctrine)**: no new event type. Marker payload gains
  additive `omitempty` fields; `truncated` is a new value in an existing free-form
  reason field; `cog.outcome{retried}` is existing whitelisted vocabulary. No
  format-version bump; old logs replay byte-identically.
- **Input bounding rejected as the mechanism**: the prompt is already input-bounded
  (`maxBufferSent` 60, newest kept), and the output contract is cap-bounded
  (≤5 promotes, ≤8 fades, ≤4 beliefs, gist ≤200 chars, narrative ≤1200 chars) —
  output size is NOT input-proportional, so shrinking the sent buffer discards the
  night's evidence without bounding the reply. Budget escalation fixes the actual
  failure; `maxBufferSent` stays as-is.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Truncation detection is transport-level and mechanical, consulted only
  on parse failure: `Response.Stop == llm.StopMaxTokens` (primary), OR
  `Response.OutputTokens >= the attempt's requested MaxTokens` (router guard). A
  parse success never triggers a budget retry.
- **FR-002**: On detected truncation, the same request is re-submitted with the
  budget doubled from the attempt's value, clamped at the shared token-budget
  ceiling (4096, today unexported `maxTokenBudget` in `internal/llm/config.go` —
  exported by this feature as the single source), at most 2 retries (3 attempts).
  The ladder starts at the effective configured budget
  (`llm.json max_tokens.consolidation`, default 1024).
- **FR-003**: A ladder exhausted while still truncated lands the night's rejected
  marker with reason `truncated`, distinct from `unparseable`; the episodic buffer
  stays intact and the next sleep retries from the ladder's start.
- **FR-004**: Each consumed retry injects a `cog.outcome` record with the existing
  `sim.OutcomeRetried` outcome (class `consolidation`; reason names truncation and
  the escalated budget), following the `emitSuppressed` synthetic-job shape — no
  new event type, no whitelist change.
- **FR-005**: `sim.ConsolidatedPayload` gains an additive `omitempty` field
  `Retries int` (consumed truncation retries this night, accepted or rejected);
  `CostUSD` accrues across all attempts. Reducer effect unchanged (telemetry-only
  fields).
- **FR-006**: The mind logs ONE per-night consolidation acceptance summary line
  (counts by outcome, rejected split by reason) when it observes the night has
  closed, from live absorb only — no summary spew on boot, no new event, no timer.
- **FR-007**: When ≥2 CONSECUTIVE nights each had ≥1 attempted (non-empty-skip)
  consolidation and 0 acceptances, the summary escalates to a `WARNING:`-prefixed
  daemon-log line carrying the streak length and remedy text; it repeats each
  further failed night until a night accepts.
- **FR-008**: The narrator's chapter and morgue-epilogue calls
  (`internal/mind/narrate.go`, fixed `narrMaxTokens` 800) go through the same
  truncation-aware retry helper (ladder 800 → 1600 → 3200, same clamp/attempt
  rules); their existing failure semantics (carry lines / story gap) apply only
  after the ladder.
- **FR-009**: All untouched paths stay byte-identical: transport failure (no
  marker), router-gate suppression, empty-buffer skip, validator rejection, and
  every accepted first-attempt night. No new event types; no format-version bump;
  `internal/mind/parse.go` is not modified (spec-103 hotspot).
- **FR-010**: Regression tests cover the late-world shape, not just day-1: a
  fixture agent with an episodic buffer exceeding `maxBufferSent` and 12+ held
  beliefs (including below-floor faded ones) whose scripted model truncates at the
  configured budget and completes at the escalated one; plus the detection matrix
  (stop-reason, output-tokens guard, parse-success-no-retry), ladder exhaustion →
  `truncated`, non-truncation parse failure → `unparseable` unchanged, narrator
  retry, and the summary/WARNING counters.

## Success Criteria *(mandatory)*

- **SC-001**: The playtest failure shape is reproduced and fixed in tests: a
  late-world consolidation that truncates at 1024 lands accepted via the ladder,
  with `Retries` on its marker — AC#1/AC#3 of TASK-172.
- **SC-002**: Sustained failure is loud: a scripted two-night blackout produces the
  per-night summary lines and the `WARNING:` escalation — AC#2.
- **SC-003**: A terminally-truncated night is distinguishable from a garbage reply
  in the durable record (`truncated` vs `unparseable` on the marker) and each
  consumed retry is visible as `cog.outcome{retried}`.
- **SC-004**: Narrator chapter/epilogue calls survive a first-attempt truncation
  (scripted) — the day-29 "unterminated JSON object" death path retries instead.
- **SC-005**: `go test -race ./...` green; existing replay fixtures byte-identical;
  no `internal/mind/parse.go` diff.

## Assumptions

- Both provider callers already map truncation onto `llm.StopMaxTokens`
  (`mapFinishReason("length")`, `mapAnthropicStop(StopReasonMaxTokens)`) and the
  worker threads `cr.stop`/`cr.outTok` into every completed `Response` — verified
  on this branch; no transport changes needed beyond exporting the budget clamp.
- Ladder dials (×2 per retry, 2 retries, threshold 2 nights) are compiled-in
  defaults, not new llm.json knobs: the operator's existing
  `max_tokens.consolidation` knob moves the ladder's START, and the clamp bounds
  its END — the ladder makes the default self-healing rather than adding tuning
  surface.
- A status-wire/TUI condition surface (à la [[llm-provider-health]] dead-tier) is
  OUT of scope: the card's AC names "telemetry/log summary" as the observability
  surface; if the ladder proves insufficient in a future playtest, a status
  condition is a separate deliverable.
- Tier: Opus 4.8 (per the board card and constitution P.V — cross-package
  `internal/mind` + `internal/llm` seam, failure-handling in async mind
  orchestration).

## Open questions

None — every design fork (detection signal, ladder policy, input bounding,
observability channel, status-surface scope) is resolved above from the card's ACs,
the transport's verified capabilities, and existing repo vocabulary.
