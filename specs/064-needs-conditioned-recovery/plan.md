# Implementation Plan: Needs-Conditioned Recovery Intents

**Branch**: `task-104-needs-recovery` | **Date**: 2026-07-25 | **Spec**: [spec.md](spec.md)

## Summary

Add an optional completion condition (need + threshold) to the Intent model
(event-payload + omitempty snapshot fields), a per-tick completion check in
the executor's active-intent path (arrive → hold-at-target → complete on
threshold), a `warm_up` planner tool (until_warmth arg, clamp-with-notice)
plus reflex issuance from 062's warmth rungs, abort-on-no-progress via a
doctrine window, a cold-emergency arm in the sleep wake gate, and a rest
analog aligning sleep's existing end-conditions to the shared mechanism.

## Technical Context

**Language**: Go 1.26.4 · **Deps**: none new · **Testing**: `go test ./...`;
executor/policy/mind suites + recover-then-release determinism tests ·
**Constraints**: replay determinism (condition in the intent_set payload;
check is a pure state function); pre-064 snapshot/event compat (omitempty);
no format_version bump; no new preemption immunity; no-LLM behavior change
limited to the intended hold-at-fire fix (enumerated) · **Scope**:
`internal/sim/{agents,executor,policy,state}.go` (Intent fields, completion
check, wake gate, reflex issuance), `internal/tool/registry.go` +
`internal/mind/handlers.go` (warm_up tool + args), tests alongside.

## Constitution Check

- **I Artifact-grounded** — PASS (spike Direction B + 057 audit Gap C are the
  record). **II One task, one PR** — PASS (`.worktrees/task-104`).
  **III Gates** — PASS.
- **IV Grounding freshness** — PASS with follow-through: executor,
  reflex-policy, tool-registry, agent-mind notes re-pin post-merge.
- **V Model-tiered** — PASS: **Opus 4.8** — cross-package (sim executor +
  tool registry + mind handler), doctrine-adjacent intent-completion
  semantics. Recorded at dispatch.

**Post-Phase-1 re-check**: PASS.

## Phase 0 research decisions

- **R1 — condition shape**: `Intent.UntilNeed string` + `Intent.UntilValue int`
  (or a small struct), omitempty in the intent_set payload and in the active
  intent record; absent ≡ arrive-and-done. Validate need names against a
  closed set (warmth, rest, food) at the door.
- **R2 — hold semantics**: on arrival with a condition, the executor keeps the
  intent active (agent state visibly "recovering" via the existing goal
  surface) and checks the condition each tick; completion emits the normal
  intent_done (source discipline intact — reflex-issued never arms 062's
  window).
- **R3 — warm_up surfaces**: planner tool `warm_up` (target like goto_warmth;
  optional `until_warmth`, clamped with notice per the 058 posture, default
  `warmthRecoverTo` — new doctrine constant, healthy margin above
  dangerWarmthBelow; pick against the needs scale, spike example 800);
  reflex day/night rungs issue goto-warmth-with-condition (same Intent form).
- **R4 — abort-on-no-progress**: while holding, if the need shows no net gain
  over `recoveryStallTicks` (doctrine constant, small — e.g. a few need-decay
  periods), abort with a distinct outcome (the executor's existing
  reject/outcome vocabulary — nearest existing shape, e.g. an
  `intent_rejected`-family or a dedicated stalled outcome). Covers dead fire,
  displaced source, unreachable threshold.
- **R5 — wake-to-cold**: extend the existing wakeReason gate with a warmth
  arm at the exposure-emergency band (reuse the 059/062 constants — one
  doctrine home); mirrors the hunger-emergency wake shape exactly.
- **R6 — rest analog (US2 proof)**: align sleep's existing end conditions to
  read through the same condition-check helper WITHOUT changing sleep's
  observable behavior (the proof is mechanism-sharing, not new behavior);
  if alignment cannot be behavior-preserving, prove US2 with a
  rest-conditioned variant in tests only and flag it.
- **R7 — staleness/preemption**: no new machinery — the existing intent
  staleness window and the planner/preemption paths apply to holding intents
  exactly as to moving ones; tests pin it.

## Project Structure

```text
internal/sim/agents.go       Intent condition fields; warmthRecoverTo, recoveryStallTicks
internal/sim/executor.go     hold-at-target + per-tick check + abort; wake gate arm
internal/sim/policy.go       062 warmth rungs issue conditioned intents
internal/sim/state.go        payload plumbing (intent_set/active record)
internal/tool/registry.go    warm_up tool (+ until_warmth, Clamp posture)
internal/mind/handlers.go    warm_up handler → InjectArgs
*_test.go                    SC-001..005 (recover-release, extended Sage, abort/
                             preempt/stale, rest analog, Oak-night wake)
specs/064-needs-conditioned-recovery/  spec.md, plan.md, tasks.md, checklists/
```

## Complexity Tracking

None.
