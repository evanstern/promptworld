# Implementation Plan: Sleep-gated planning (TASK-175)

**Branch**: `task-175-sleep-gated-planning` | **Date**: 2026-07-30 | **Spec**: [spec.md](spec.md)

## Summary

Two mind-side layers close the two post-enqueue leak windows behind playtest-1's
905 "is asleep" rejections: (1) an atomic per-agent unavailability mirror
(asleep|dead, absorb-maintained — the `md.tick` mirror precedent) consulted by
`runPlan` at dequeue, before any model call; (2) a race-safe per-agent cancel
slot that aborts an in-flight planner call when absorb applies `agent.slept` /
`agent.died`. The enqueue-time gate in `plan()` and the sim's `rungUnavailable`
ladder rung are byte-unchanged; the ladder stays the authoritative backstop.
Off-log: no reducer, event-schema, or outcome-vocabulary changes.

## Technical Context

**Language**: Go. **Surface**: `internal/mind/mind.go` only (struct fields,
absorb hooks, `runPlan` gate + cancel wiring), plus a small telemetry emit —
prefer inline reuse of `cogOutcomeEvent`/`emitSuppressed`-shaped emission in
mind.go or `internal/mind/telemetry.go`. **Testing**: unit tests alongside
(`runplan_test.go` pattern — the `runLoopOverride` seam scripts the loop;
`Submitter`/`Injector` fakes exist throughout `mind_test.go`), plus a seeded
soak for SC-003. `go test -race ./...`.

**Key mechanics (verified)**:
- `plan()` sets `planInFlight[i]` before enqueue; `runPlan` defers its release
  — the dequeue gate sits at the top of `runPlan` so the defer covers the skip
  path for free.
- `absorb` already switches on `agent.slept` (→ `maybeConsolidate`) and owns
  the replica; the mirror store rides next to the existing
  `md.tick.Store(...)` at batch end, so `gru.attacked`'s eventless wake is
  covered automatically (the replica applied the arm).
- `agent.woke` → `md.arm` already exists (wake resumption, spec FR — verify by
  regression test, no new code).
- `runPlan` builds its own `context.WithTimeout(callTimeout)`; the cancel slot
  wraps/stores that context's cancel per agent (mutex or atomic slot; set
  before `md.runLoop`, cleared after; absorb invokes it on slept/died).
- Skip telemetry: `cog.outcome{suppressed}` through `emitCog` (worker
  goroutine — safe; `emitSuppressed` itself is absorb-side and bumps the
  spec-037 counter, which FR-003 forbids here, so emit the payload directly
  with the sleep reason instead of calling it).

## Constitution Check

- **I–IV: PASS** — decision + rejected alternatives ratified in spec.md; one
  task, one branch, one PR; tests as evidence; in-branch wiki re-pins gated by
  the pr gate.
- **V: PASS — Opus 4.8** (recorded on the board card at dispatch):
  concurrency/scheduling logic in `internal/mind` orchestration is named
  explicitly by the escalation rubric. Absorb-goroutine/worker-goroutine
  ownership rules and a cancellation race are exactly the hard-slice shape.

## Project Structure

- `internal/mind/mind.go` — mirror fields + absorb hooks + `runPlan` gate +
  cancel slot (the entire production diff; keep it surgical)
- `internal/mind/telemetry.go` — only if a small emit helper reads better than
  inline; otherwise untouched
- `internal/mind/sleep_gate_test.go` (or extend `runplan_test.go` /
  `mind_test.go`) — SC-001/002/005 + consolidation-untouched + dead parity
- `specs/106-sleep-gated-planning/` — these artifacts
- `docs/wiki/` re-pins ride the branch (below)

## Risks & mitigations

- **TASK-112 cross-session hotspot (spec 102, guardian agentization) touches
  `internal/mind` on a live sibling branch; this PR merges only AFTER 112
  lands (operator ruling, LANE 2 after TASK-172).** Mitigation: keep the diff
  confined to mind.go + tests; after 112 merges to main, merge main INTO this
  branch (never rebase — pin-staling hazard), re-run `go test -race ./...`,
  re-verify the absorb/runPlan seams still match, and only then re-pin wiki
  notes and open the PR.
- **Estimator/retry accounting on cancelled calls** (FR-005): verify the
  ctx-cancel termination path (`TermCtxDone`) neither consumes the loop's
  transport retry nor feeds a spike observation into the per-provider
  estimator; adjust only if a real poisoning path exists (adversarial check —
  Opus tier).
- **Mirror lag**: sub-second absorb lag can still let a call start or land
  around a sleep boundary; the untouched ladder rung is the backstop and the
  SC-003 soak budget (≤ 1/game-day) absorbs it.
- **Replica drift on `Observe` overflow-drop** (pre-existing): a dropped batch
  could stale the mirror until the next batch; same backstop, no new exposure.

## Grounding (wiki-in-PR, spec 069)

Touching `internal/mind/mind.go` (+ possibly `telemetry.go`) invalidates pins
on: [[mind-driver-triggers]], [[tool-use-dispatch]], [[planner-telemetry]] —
behavior-review re-pins (the gate/cancel change what those notes describe) —
plus computed re-verifies for the other notes listing those files as sources
(agent-mind, cognition, decision-context, journal-tool-integration,
memory-retrieval, mental-map-propagation, mental-map-perception,
agent-memory-window, scenario-machinery-surfacing). Regenerate `docs/player/`
if wiki content changes; `node scripts/check-merge-drift.mjs pr` must exit 0.
