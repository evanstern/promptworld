# Tasks: Consolidation truncation-aware retry + acceptance observability (TASK-172)

**Input**: `specs/105-consolidation-truncation/spec.md`

## Phase 1: Detection + ladder helper

- [X] T001 Export the token-budget clamp in `internal/llm/config.go`
  (`maxTokenBudget` → `llm.MaxTokenBudget`, value 4096, all internal uses
  updated; behavior byte-identical) (FR-002).
- [X] T002 `internal/mind/retry.go` (new): truncation-aware submit helper —
  Submit loop with caller-supplied parse closure; detection per FR-001
  (parse-first; `Stop == llm.StopMaxTokens` primary, `OutputTokens >= requested
  budget` router guard); doubling ladder from the caller's start budget clamped
  at `llm.MaxTokenBudget`, ≤2 retries; returns final response, consumed retries,
  accrued cost, terminal-truncation flag. Tests alongside in
  `internal/mind/retry_test.go`: detection matrix (max_tokens stop,
  output-tokens guard on `StopOther`, parse-success-never-retries even with
  max_tokens stop, non-truncation parse failure never retries), ladder
  arithmetic 1024→2048→4096 and the no-headroom (start == clamp) single-attempt
  case (FR-001, FR-002).

## Phase 2: Consolidation integration

- [ ] T003 `internal/sim/consolidate.go`: additive `omitempty`
  `ConsolidatedPayload.Retries int` + `ConsolidationReasonTruncated` const;
  reducer arm untouched. Round-trip + old-payload byte-compat test in
  `internal/sim/consolidate_test.go` (FR-003, FR-005).
- [ ] T004 `internal/mind/consolidate.go`: `runConsolidation` drives the T002
  helper (same job snapshot and prompt on every attempt — dream geometry never
  re-run); accepted nights stamp `Retries` and accrued `CostUSD` on the marker;
  terminal truncation lands `ConsolidationReasonTruncated` (buffer intact);
  non-truncation parse failure still lands `unparseable`; each consumed retry
  emits `cog.outcome{retried}` (class `consolidation`, `emitSuppressed`
  synthetic-job shape, detached injection) (FR-002–FR-005, FR-009).
- [ ] T005 Regression tests in `internal/mind/consolidate_test.go` with the
  scripted submitter: (a) the LATE-WORLD fixture — buffer > `maxBufferSent`,
  12+ held beliefs including below-floor faded ones, model truncates at 1024
  and completes at 2048 → accepted marker with `Retries: 1` and both attempts'
  cost (SC-001); (b) ladder exhaustion → rejected `truncated`, buffer intact,
  next night retries; (c) first-attempt success byte-identical, no retry
  consumed; (d) transport failure still lands no marker (FR-009, FR-010).

## Phase 3: Narrator generalization

- [ ] T006 `internal/mind/narrate.go`: chapter + morgue-epilogue Submits adopt
  the helper (ladder 800→1600→3200); existing carry/gap semantics apply only
  after the ladder. Tests in `internal/mind/narrate_test.go`: truncated chapter
  retried and landed; truncated epilogue retried; ladder-exhausted chapter falls
  into today's carry path (FR-008, SC-004).

## Phase 4: Per-night observability

- [ ] T007 `internal/mind/nightreport.go` (new) + absorb hook: per-night outcome
  counters keyed by `sim.NightIndex`, fed from live-absorbed
  `agent.consolidated` markers; flush one summary log line when a later night is
  observed; consecutive attempted-none-accepted streak ≥2 escalates to a
  `WARNING:` line with remedy text, repeating each further failed night until a
  night accepts; in-memory only, no timer, no new event. Tests alongside:
  counter/flush behavior, streak arm/repeat/reset, no summary from
  pre-live/boot state (FR-006, FR-007, SC-002).

## Phase 5: Docs + grounding

- [ ] T008 `docs/event-types.md`: `agent.consolidated` `retries` field +
  `truncated` reason; `cog.outcome` retried-reason note for the consolidation/
  narrator classes. Full-suite `go test -race ./...` green; existing replay
  fixtures byte-identical; confirm zero `internal/mind/parse.go` diff (FR-009,
  SC-005).
- [ ] T009 Wiki re-pins in-branch for every note whose sources this branch
  touches (expected: nightly-consolidation, chronicle, morgue-epilogues,
  event-types-memory-consolidation, event-types-cognition-telemetry prose if
  amended, llm-orchestrator child covering token budgets, plus any
  mind.go/telemetry.go-sourced notes); regenerate `docs/player/`; merge
  `origin/main` in AFTER TASK-112 lands (plan Risks) and re-verify; merge-drift
  pr gate exit 0.
