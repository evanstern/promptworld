# Implementation Plan: Instinct Yields to Intelligence

**Branch**: `task-103-instinct-yields` | **Date**: 2026-07-25 | **Spec**: [spec.md](spec.md)

## Summary

Restructure `decideIntent` (`internal/sim/policy.go`) into explicitly
classified SURVIVAL and PREP rungs; gate PREP on (a) a yield window armed by
non-reflex intent completion and (b) per-need danger bands; add the day-branch
warmth rung (night ladder mirrored); add the bounded night frontier-search
fallback; prove the world-01 loop dead with a scripted regression. The yield
signal is event-sourced on `sim.State` (the intent-completion events already
carry source — state records the last non-reflex completion tick per agent).

## Technical Context

**Language**: Go 1.26.4 · **Deps**: none new · **Testing**: `go test ./...`;
policy/reflex suites + the new regression · **Constraints**: replay
determinism; pre-062 snapshot compat (omitempty); rebase taxonomy (completion
tick SHIFTs); no-LLM degraded mode byte-compatible except danger-band
suppression; no tuning.json additions · **Scope**: `internal/sim/policy.go`
(the ladder), `internal/sim/state.go` (yield state + intent-completion arm),
`internal/sim/agents.go` (danger-band constants beside the existing
thresholds), tests alongside.

## Constitution Check

- **I Artifact-grounded** — PASS (spike TASK-101 notes + 057 audit are the
  direction record). **II One task, one PR** — PASS (`.worktrees/task-103`).
  **III Gates** — PASS.
- **IV Grounding freshness** — PASS with follow-through: reflex-policy,
  sim-state-reducer, decision-context wiki notes re-pin post-merge.
- **V Model-tiered** — PASS: **Opus 4.8** — doctrine-adjacent behavior change
  at the heart of the sim reducer's decision ladder; the slice the whole MVLS
  program hinges on. Recorded at dispatch.

**Post-Phase-1 re-check**: PASS.

## Phase 0 research decisions

- **R1 — yield state shape**: per-agent `LastMindIntentDone int64`
  (tick of the most recent completed non-reflex intent), on the Agent struct
  (agents already serialize per-agent state) with omitempty; updated by the
  intent-completion reducer arm reading the intent's source. Reflex
  completions never write it. Rebase: SHIFT.
- **R2 — rung classification**: SURVIVAL = eat-now, hungry-get-food (+its
  search), night warmth ladder, NEW day warmth rung, exhausted-nap, sleep;
  PREP = first-fire prep (build/chop when no fire known), non-dying refuel
  top-up, larder stock-to-8. The TASK-108 dying-fire refuel (< 3h window) is
  SURVIVAL — it is the burnout-prevention reflex; the audit table is the
  authority; encode as two ordered rung groups in decideIntent.
- **R3 — danger bands**: warmth/food/rest bands anchored at the existing
  survival-rung triggers (`hungryAt`, the cold/`warmAt` boundary, `tiredAt`)
  — "in danger" == "a survival rung would or will imminently act". New named
  constants `dangerFoodBelow`/`dangerWarmthBelow`/`dangerRestBelow` beside
  the existing thresholds, values = those thresholds unless the implementer
  finds an evidenced reason to pad (flag it). Dial-ready, not dialed.
- **R4 — yield window**: `prepYieldTicks = 1800` (one default planner
  cadence — deliberately the CONSTANT, not the tuned dial: the window is
  arbitration doctrine, not scheduling; a cadence-tuned world shouldn't
  silently stretch instinct's deference). Dial-ready, not dialed.
- **R5 — day warmth rung**: reuse the night branch's exact seek → refuel-dying
  → build ladder via a shared helper, gated on the warmth danger band during
  day (night keeps its unconditional not-warm trigger).
- **R6 — night search fallback**: reuse `nearestFrontier` (the hungry-search
  shape) as the last rung before terminal sleep, only when cold + no known
  warmth + no wood + no choppable known.
- **R7 — regression**: scripted state (agent cold, fed, larder low, planner
  goto_warmth completing at a fire) driven through the executor for N ticks;
  assert no prep intent within the window and warmth recovery; the inverse
  assertion documents the old behavior by running the scenario with the
  window/bands zeroed.

## Project Structure

```text
internal/sim/policy.go        rung groups; PREP gate; day warmth rung; night search rung
internal/sim/state.go         LastMindIntentDone update in the completion arm
internal/sim/agents.go        danger-band + prepYieldTicks constants (doctrine home)
internal/sim/policy_test.go   US1-US3 scenario tests + no-LLM parity drive
internal/sim/thrash_regression_test.go  US4 Sage-shape regression
specs/062-instinct-yields/    spec.md, plan.md, tasks.md, checklists/
```

## Complexity Tracking

None.
