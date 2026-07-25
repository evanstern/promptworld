# Implementation Plan: Conversation Loop Damper

**Branch**: `task-109-convo-damper` | **Date**: 2026-07-25 | **Spec**: [spec.md](spec.md)

## Summary

Two layered gates on conversation founding, per the proven diagnosis
(`docs/design/evidence/task-109/`): (1) **sim-side pair cooldown** — an
event-sourced unordered pair→last-exchange-tick record on `sim.State` (updated
by the `agent.talked` reducer arm, spec-044 Deaths-ledger precedent), consulted
by the hail founding path (`internal/sim/hail.go` `hailable`/`hailStep`) using
the existing `s.EncounterCooldown()` dial (spec 048); refusal surfaces through
the landing outcome so the planner learns "you spoke recently". (2) **mind-side
novelty SHIM** — scene founding (`internal/mind/convo.go`) requires a new
above-floor salience memory since the pair's last exchange and injects the last
gist into the scene prompt; marked removable with a greppable marker.

## Technical Context

**Language**: Go 1.26.4 · **Deps**: none new · **Testing**: `go test ./...`;
hail/convo/replay suites · **Constraints**: replay determinism; pre-061
snapshot compat (omitempty); rebase taxonomy (pair ticks SHIFT); ambient +
encounter gates byte-identical behavior; no format_version bump · **Scope**:
`internal/sim/{state,hail,executor?}.go`, `internal/mind/{convo,mind}.go`,
landing outcome surface, tests alongside.

## Constitution Check

- **I Artifact-grounded** — PASS (diagnosis evidence committed; decisions on
  task + spec). **II One task, one PR** — PASS (`.worktrees/task-109`).
  **III Gates** — PASS.
- **IV Grounding freshness** — PASS with follow-through: social-fabric,
  agent-mind, hail-protocol wiki notes re-pin post-merge.
- **V Model-tiered** — PASS: **Opus 4.8** — internal/mind orchestration
  (scene founding) + doctrine-adjacent sim gate + cross-package state/gate
  semantics. Recorded at dispatch.

**Post-Phase-1 re-check**: PASS.

## Phase 0 research decisions

- **R1 — pair record shape**: deterministic canonical marshal — a sorted slice
  of `{A, B int; Tick int64}` with A<B invariant (maps marshal
  non-deterministically; the state's canonical-bytes discipline forbids them
  unless already handled). Bounded by C(agents,2)=28 for 8 villagers.
  `json:"pair_talks,omitempty"`.
- **R2 — where the update happens**: the `agent.talked` reducer arm (both
  participants known there). Verify hail-founded and ambient talks both emit
  `agent.talked` with the pair (diagnosis says yes — LastTalk is already
  written on every agent.talked).
- **R3 — where the sim gate lives**: `hailable()` gains the pair-cooldown
  check (it already receives the hailer/hailee context) — this gates the
  landing door's hail rungs AND `hailStep`; confirm both route through it.
  Refusal message through the existing landing outcome/refusal surface (the
  same channel "no warmth anywhere"-class refusals use).
- **R4 — novelty gate site**: mind scene founding (`maybeStartConversation` or
  its successor in convo.go). Reads the replica's memory store for
  since-last-exchange salience; last-exchange tick comes from the SAME sim
  pair record via the replica (one source of truth). Gist lookup via the
  convo_record machinery.
- **R5 — SHIM marking**: marker string `SHIM(TASK-109)` at every gate site +
  removal condition in a doc block; SC-005 makes it greppable.

## Project Structure

```text
internal/sim/state.go        PairTalks record + agent.talked arm update
internal/sim/hail.go         hailable pair-cooldown gate + refusal
internal/sim/tuning.go       (no change — EncounterCooldown accessor exists)
internal/mind/convo.go       novelty SHIM + last-gist prompt context
internal/sim/*_test.go       gate/replay/rebase/compat tests
internal/mind/*_test.go      novelty + gist tests
specs/061-conversation-loop-damper/  spec.md, plan.md, tasks.md, checklists/
docs/design/evidence/task-109/       diagnosis (committed with this spec)
```

## Complexity Tracking

None.
