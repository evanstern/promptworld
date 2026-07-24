# Tasks: Paused Authoring Chain-Completion

**Input**: Design documents from `/specs/040-paused-chain-completion/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/recorded-events.md, quickstart.md

**Tests**: INCLUDED — the spec's success criteria are test-shaped (SC-002…SC-005
demand byte-identity, determinism, and bounded-round proofs), and the
constitution's gate culture requires the proof artifacts. Write each story's
tests first and watch them fail before implementing.

**Organization**: grouped by user story (US1 wake, US2 truthful routing, US3
running-world byte-identity), each independently testable. Board: TASK-77 —
all phases land on the single branch `task-77-paused-chain-completion` in
`.worktrees/task-77`, one PR.

**Tier**: doctrine-adjacent `internal/mind` change — implementation executes on
the **Opus 4.8** `spec-implementer` tier (constitution V; recorded on TASK-77).

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

- [ ] T001 Create the TASK-77 worktree from fresh origin/main: `git worktree add .worktrees/task-77 -b task-77-paused-chain-completion origin/main` (root stays on main; all subsequent tasks execute inside `.worktrees/task-77/`)

## Phase 2: Foundational

No foundational tasks — every consumed shape already exists (`State.Paused`
reducer, `metatron.nudged` payload, arm/debounce machinery, Verdict struct;
see research.md grounding table). User stories can start immediately after T001.

---

## Phase 3: User Story 1 — A nudge wakes the nudged villager while paused (Priority: P1) 🎯 MVP

**Goal**: while paused, a landed Metatron nudge arms each targeted villager for
exactly one debounce-bounded planner round at the frozen tick, with the nudge
event recorded as the arming stimulus.

**Independent Test**: harness world at 16x (planner never suppressed at 16x, so
this story needs no routing change): pause, inject the nudge landing batch,
assert exactly one planner thought at the frozen tick with `trigger_seq` = the
nudge's Seq; nudge again, assert no second thought.

### Tests for User Story 1 (write first, must fail)

- [ ] T002 [US1] Harness test `TestPausedNudgeWakesTargetOnce` in internal/mind/telemetry_test.go (pattern: `TestPauseInFlightThoughtLandsAtFrozenTick`, 16x): pause via `h.loop.Do("pause","")`, inject the `metatron.nudged` + `agent.memory_added` batch (shape: internal/metatron/turn.go:422-427), assert exactly one planner `cog.thought`/`cog.outcome` for the target at the frozen tick with `staleness_ticks == 0` and `trigger_seq` equal to the nudge event's Seq (contracts C2/C3); then inject a second nudge in the same pause and assert the model call count for that agent does not increase (debounce bound, spec US1 scenario 2)
- [ ] T003 [P] [US1] Harness test `TestPausedOmenArmsOnlyTargets` in internal/mind/telemetry_test.go: paused world, multi-target omen batch → each living awake target gets exactly one thought; a non-targeted villager gets none (spec US1 scenario 3)

### Implementation for User Story 1

- [ ] T004 [US1] Add the paused-gated wake to `absorb()`'s arm switch in internal/mind/mind.go (after the existing cases, mind.go:212-237): `case "metatron.nudged":` unmarshal `sim.MetatronNudgedPayload`, and only when `md.replica.Paused`, `md.arm(t, e.Seq)` for each `t` in `Targets` (research.md D1; FR-001/FR-002/FR-003). No other bounding: the debounce is the bound (D4)
- [ ] T005 [US1] Prove the doctrine tests still hold: `go test ./internal/mind/ -run 'TestPause|TestResume' -v` — `TestPauseStartsNoNewThoughts`, `TestPauseInFlightThoughtLandsAtFrozenTick`, `TestPauseConversationLandsAtFrozenTick`, `TestResumeNoBurst` all green alongside T002/T003

**Checkpoint**: paused nudge → one frozen-tick thought works end-to-end at
non-suppressing speeds; US1 demonstrable alone.

---

## Phase 4: User Story 2 — Paused routing tells the truth (Priority: P1)

**Goal**: a paused world routes every decision class at zero predicted drift
(allow), the recorded arithmetic names the paused state, and the land-tick
prediction (prompt + telemetry) agrees the thought lands at the frozen tick.

**Independent Test**: pure cognition table test plus mind-level routeVerdict
assertions on a paused replica whose SET speed suppresses the planner —
verdict allow with "paused" arithmetic; resume → today's strings.

### Tests for User Story 2 (write first, must fail)

- [ ] T006 [P] [US2] Pure table test `TestRoutePaused` in internal/cognition/route_test.go: for planner/conversation/meeting classes assert `Allow == true`, `PredictedDriftTicks == 0`, `PredictedWallMs == Points × spp × 1000`, `BudgetTicks`/`Points`/`Class` populated, and `Arithmetic` exactly matches contract C1 (`"%dpt x %.1fs/pt while paused = 0 ticks <= budget %d"`) — including that it contains the word `paused`
- [ ] T007 [P] [US2] Mind test `TestRouteVerdictPausedAllowsAtSuppressingSpeed` in internal/mind/telemetry_test.go: replica with `Paused = true` and a set speed whose arithmetic suppresses the planner (and separately `tps <= 0`, the uncapped case — paused must win, spec US2 scenario 3): `routeVerdict("planner", …)` returns allow + C1 arithmetic; with `Paused = false` the verdict is byte-identical to `cognition.Route`'s output (FR-005)
- [ ] T008 [P] [US2] Mind test `TestPausedThoughtPredictsFrozenLanding` in internal/mind/telemetry_test.go: on a paused replica, `newMeta` yields `predictedLandTick == snapshotTick`, and a planner job prompt built while paused carries NO `futureDated` prefix (prompt.go:63-69 no-ops at landing ≤ now); the recorded `cog.thought.predicted_land_tick` equals the snapshot tick (contract C3)

### Implementation for User Story 2

- [ ] T009 [US2] Add pure `RoutePaused(dc DecisionClass, secondsPerPoint float64) Verdict` to internal/cognition/route.go per research.md D2: Allow true, drift 0, wall-ms predicted, C1 arithmetic string; do NOT touch `Route` (SC-005)
- [ ] T010 [US2] Consult the paused flag in `routeVerdict` (internal/mind/telemetry.go:61-71): immediately after `ClassFor`, `if md.replica.Paused { return cognition.RoutePaused(dc, md.secondsPerPoint(kind)) }` — before the `tps <= 0` branch so paused wins at uncapped (D2)
- [ ] T011 [US2] Truthful land prediction in `newMeta` (internal/mind/telemetry.go:38-54): when `md.replica.Paused`, set `predictedLandTick = snapshotTick` instead of the tps projection (D3)
- [ ] T012 [US2] Composition test `TestPausedNudgeThinksAtSuppressingSpeed` in internal/mind/telemetry_test.go: harness world at a planner-suppressing speed (pattern: `TestPlannerSuppressedAtHighSpeed`), pause, inject nudge batch → the thought is ATTEMPTED and LANDS at the frozen tick (US1 wake + US2 routing together — the full decision-6 chain), with zero `cog.outcome` suppressions while paused (SC-003)

**Checkpoint**: the full learner loop (pause → nudge → thought under charter)
works even on worlds paused from suppressing speeds.

---

## Phase 5: User Story 3 — The running world is untouched (Priority: P2)

**Goal**: byte-identical unpaused behavior — no new wake stimuli, no verdict
changes, determinism harness green.

**Independent Test**: land a nudge while RUNNING and prove no nudge-attributable
arm; full existing suite green.

### Tests for User Story 3 (write first where new, must fail against a buggy ungated impl)

- [ ] T013 [P] [US3] Mind test `TestRunningNudgeDoesNotArm` in internal/mind/mind_test.go: unpaused replica, feed a `metatron.nudged` event through `absorb()` for an agent inside its debounce-fresh window, assert `pending` is NOT set by the nudge (unit-level, no harness needed) — guards FR-003 directly so a future un-gating fails loudly
- [ ] T014 [US3] Determinism + regression gate: `go build ./... && go vet ./... && go test ./...` from the worktree — entire existing suite (including internal/sim byte-identical replay tests and internal/mind/replay_test.go) green with zero test edits outside the files this feature adds to (SC-004/SC-005); if no existing replay scenario contains a paused nudge-thought session, add one following internal/sim/governor_replay_test.go's byte-identical pattern

**Checkpoint**: all three stories proven; unpaused world provably today's.

---

## Phase 6: Polish & Cross-Cutting

- [ ] T015 Run quickstart.md validation end-to-end (sections 1–3) inside the worktree and record the output in the PR description; tick spec ACs on board TASK-77 (`backlog task edit 77 --check-ac …` from repo root) and `spec-bridge:sync`
- [ ] T016 Post-merge grounding (root, on main, after the PR merges): `/grounding-wiki:wiki-update` for notes sourcing internal/mind/mind.go, internal/mind/telemetry.go, internal/cognition/route.go (agent-mind, cognition-horizon et al.), then `node .claude/skills/player-docs/scripts/check-freshness.mjs --check` and refresh player docs if stale (constitution IV)

---

## Dependencies & Execution Order

- **T001** blocks everything (worktree).
- **US1 (T002–T005)** and **US2 (T006–T011)** are independent of each other and
  can proceed in parallel after T001; T012 needs both T004 and T010/T011.
- **US3**: T013 needs T004 (it tests the gate on the new case); T014 needs all
  implementation tasks.
- **Polish**: T015 after T014; T016 strictly post-merge.
- Within stories: tests (T002/T003, T006–T008, T013) before their
  implementations (T004, T009–T011); watch them fail first.

### Parallel Opportunities

- T002 ∥ T003 (same file — write in one pass if convenient); T006 ∥ T007 ∥ T008
  (route_test.go vs telemetry_test.go).
- After T001: US1's tests and US2's pure cognition work (T006/T009) touch
  disjoint packages and can run in parallel.

## Implementation Strategy

MVP = Phase 3 (US1): at watchable non-suppressing speeds the learner loop
already completes with the wake alone. Phase 4 (US2) extends it to worlds
paused from suppressing speeds and makes every paused verdict truthful. Phase 5
proves the blast radius is zero. One branch, one PR (TASK-77); commit per task
or logical group.
