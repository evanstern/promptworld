# Implementation Plan: Conversation-outcome JSON robustness (TASK-174)

**Branch**: `task-174-conversation-outcome-json` | **Date**: 2026-07-30 | **Spec**: [spec.md](spec.md)

## Summary
Restore the TASK-58 structured-outputs transport (additive
`Request.ResponseSchema`/`SchemaName`, attached by `openai_compat.callNative`
as `response_format {type: json_schema}` iff set and tool-free; Anthropic
ignores it) and stamp two generated schemas on the conversation scene's
single-shot calls: `convoOutcomeSchema` on `Mind.outcome`, `sayReplySchema` on
`Mind.utterance`. No cloud fallback (D2 — the scene provider pin is ratified
doctrine); the TASK-42 retry/abandon ladder is unchanged as the safety net.
Parse-failure rate is measured over existing `cog.outcome` telemetry (D6) and
proven on a soak against a small local model.

## Technical Context
**Language**: Go. **Surfaces**: `internal/llm` (Request shape + openai_compat
transport only — the worker, chain-walk, leases, breaker are untouched),
`internal/mind` (new `convo_schema.go`; two Submit sites in `convo.go`).
**Testing**: provider payload tests (envelope iff schema; byte-identical
baseline; no attach beside Tools), schema single-source tests, Request-capture
via the existing `Submitter` fakes in `convo_test.go`, full suite `-race`.
**Constraints**: zero `parse.go` diff (spec 105 hotspot, D5/SC-003); no
reducer/event/payload changes — replay is model-free and stays so; prompts
byte-identical; soak on a MEASURE world with local-only routes.

## Constitution Check
I–IV: PASS — spec 103 records D1–D6 with rationale from durable artifacts
(TASK-58 commit `f6bd31ae`, TASK-71 drop `db45c44b`, spec 024 US3 provider-pin
doctrine, spec 011/TASK-42 telemetry contract); one branch/PR; test + soak
evidence; wiki re-pins ride the branch.
V: PASS — **Sonnet** (routine single-subsystem, established pattern, no
concurrency/doctrine surface); recorded on the board card by the sweep
dispatch.

## Project Structure

Files changed (predicted):

- `internal/llm/llm.go` — restore `ResponseSchema`/`SchemaName` on `Request`
  (additive, omitempty, documented incl. the no-attach-beside-Tools rule).
- `internal/llm/providers.go` — `callNative` attaches the envelope when set
  and `len(req.Tools) == 0`.
- `internal/llm/providers_test.go` — envelope iff schema; byte-identical when
  unset; not attached when Tools present.
- `internal/mind/convo_schema.go` (NEW) — `convoOutcomeSchema` +
  `sayReplySchema`, built once from `gistCapBytes`/`sayCapBytes`/`sceneCap`
  (single source of truth; no hand-copied literals).
- `internal/mind/convo_schema_test.go` (NEW) — schema validity/round-trip,
  caps == registry caps, tones `maxItems` == `sceneCap`, no `anyOf`.
- `internal/mind/convo.go` — `outcome()` and `utterance()` Submits stamp the
  schemas; nothing else moves.
- `internal/mind/convo_test.go` — Request-capture assertions on both sites;
  existing TASK-42 ladder tests stay green unmodified.
- `docs/design/evidence/task-174/` — soak evidence: queries, counts, baseline
  comparison (FR-005).
- Wiki re-pins (in-branch, spec 069): prose amendment +
  re-pin on `social-fabric-conversations` (the note's retry paragraph gains
  the constrained-decoding layer); source-touched re-verifies on the notes
  pinning `internal/llm/llm.go` / `providers.go` / `internal/mind/convo.go`
  (`llm-orchestrator`, `llm-provider-registry`, `llm-chain-walk-dispatch`,
  `llm-concurrency-leases`, `llm-budget-degraded-mode`,
  `llm-preflight-detection`, `llm-provider-health`,
  `guardian-order-triggering`, `guardian-report-card`, `memory-retrieval`).
  `parse.go` untouched ⇒ its notes don't re-pin. Player-docs freshness probe
  after any wiki prose change.

No new packages, no config/schema-file surface, no tuning dials.
