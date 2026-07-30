# Implementation Plan: Consolidation truncation-aware retry + acceptance observability (TASK-172)

**Branch**: `task-172-consolidation-truncation` | **Date**: 2026-07-30 | **Spec**: [spec.md](spec.md)

## Summary

Give the nightly consolidation call (and the sibling narrator call) a
truncation-aware retry ladder: detect a cut reply mechanically at the transport
level (`Response.Stop == llm.StopMaxTokens`, or parse-failure +
`OutputTokens >= requested budget`), re-submit the same prompt with a doubled
budget clamped at the shared 4096 ceiling (≤2 retries), and make the outcome
observable — `Retries` on the `agent.consolidated` marker, a distinct `truncated`
rejection reason, `cog.outcome{retried}` per consumed retry, a per-night
acceptance summary log line, and a `WARNING:` escalation after 2 consecutive
fully-failed nights.

## Technical Context

**Language**: Go. **Surfaces**:

- `internal/llm/config.go` — export the token-budget clamp (`maxTokenBudget` 4096
  → `MaxTokenBudget`), single source for the ladder ceiling. Behavior unchanged.
- `internal/mind/retry.go` (new) — the shared truncation-aware submit helper: a
  loop of Submit → caller-supplied parse closure → detect (FR-001) → double &
  clamp (FR-002), returning the final response, consumed-retry count, accrued
  cost, and whether the terminal failure was truncation. Consolidation and the
  narrator both drive it; the tool-loop and conversation sites are untouched.
- `internal/mind/consolidate.go` — `runConsolidation` adopts the helper around
  the Submit at line ~162; terminal truncation lands the marker with reason
  `truncated`; consumed retries ride the marker and emit `cog.outcome{retried}`
  (the `emitSuppressed` synthetic-job shape in `internal/mind/telemetry.go` is
  the template — class `consolidation`, detached injection). The dream geometry
  pass and the job snapshot are NOT re-run on retry.
- `internal/sim/consolidate.go` — `ConsolidatedPayload.Retries int` (additive,
  `omitempty`, telemetry-only; reducer arm unchanged) + a
  `ConsolidationReasonTruncated` const beside the outcome consts.
- `internal/mind/narrate.go` — chapter + epilogue Submits adopt the helper
  (ladder 800→1600→3200); existing carry/gap failure handling applies after it.
- `internal/mind/nightreport.go` (new) + a hook in the absorb path — per-night
  outcome counters keyed by `sim.NightIndex`, fed from live-absorbed
  `agent.consolidated` events; a night flushes (one summary log line, FR-006)
  when a marker or tick from a later night is observed; a consecutive
  all-attempted-none-accepted streak ≥2 escalates the line to `WARNING:` with
  remedy text (FR-007). In-memory only; the replica is snapshot-seeded so boot
  does not replay history through absorb (restart amnesia accepted, spec edge
  case).
- `docs/event-types.md` — marker field/reason additions; `docs/wiki` re-pins +
  `docs/player` regen ride this branch per the pr gate.

**Testing**: scripted-submitter unit tests in `internal/mind`
(detection matrix, ladder arithmetic, late-world fixture per FR-010, narrator
retry, night-report counters), payload round-trip in `internal/sim`,
`go test -race ./...`, replay byte-identity on existing fixtures.

**Constraints**: no new event types, no whitelist or format-version change, no
`internal/mind/parse.go` diff (spec-103 hotspot), transport-failure /
router-gate / empty-buffer / validator-rejection paths byte-identical.

## Constitution Check

I–IV: PASS (spec 105; one branch/PR; test matrix as evidence; wiki re-pins
in-branch — nightly-consolidation, chronicle/morgue-epilogues,
event-types-memory-consolidation and any mind.go/telemetry.go-sourced notes
expected NEEDS-REVIEW). V: PASS — **Opus 4.8** per the board card's dispatch
ruling: cross-package (`internal/mind` worker + `internal/llm` budget seam),
failure-handling in async mind orchestration — the hard-slice rubric applies.

## Project Structure

See Technical Context; two new files (`internal/mind/retry.go`,
`internal/mind/nightreport.go`), no new packages.

## Risks

- **Merge ordering (operator ruling, LANE 2)**: this branch's PR merges only
  AFTER TASK-112 (guardian agentization, spec 102) lands — spec 102 also touches
  `internal/mind`. Before opening the PR: merge `origin/main` INTO this branch
  (never rebase), reconcile any `internal/mind` conflicts (mind.go wiring and
  telemetry.go are the likely overlap; consolidate.go/narrate.go are expected
  clean), re-run the full suite, and re-pin any wiki note the merge re-touches.
- **Spec-103 parse.go hotspot**: design keeps detection transport-level so this
  branch carries NO parse.go diff; if implementation discovers a need to touch
  it, changes must be minimal/additive and flagged in the PR body.
- **Router stop-reason fidelity**: a router that neither maps `finish_reason`
  nor reports usage tokens would defeat both detection signals; the failure mode
  is today's behavior (reject `unparseable`), now at least visible in the
  nightly summary — no regression, documented in the wiki note.
- **Estimator skew**: retry attempts are completed calls and feed the
  seconds-per-point estimator like any single-shot call today (consolidation is
  night-scale; skew risk is negligible and self-correcting). No `SkipObserve`.
