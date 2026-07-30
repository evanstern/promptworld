# Tasks: Sleep-gated planning (TASK-175)

**Input**: `specs/106-sleep-gated-planning/spec.md` (gate placement ratified
there: pre-submit gate + in-flight cancel, mind-side, ladder untouched).

## Phase 1: Unavailability mirror + pre-submit gate (US1 layer 1)

- [X] T001 Add the per-agent unavailability mirror (asleep|dead) to `Mind`,
  stored atomically by absorb at batch end beside `md.tick.Store` (FR-001);
  seed it in `New` from the freshly unmarshaled replica (the `tickRate`
  precedent).
- [X] T002 Dequeue gate at the top of `runPlan`: unavailable ⇒ no
  `cog.thought`, no `runLoop`; emit one terminal `cog.outcome{suppressed}`
  with a sleep/death reason via the telemetry door; no re-arm; `planInFlight`
  released by the existing defer (FR-002). Do NOT bump `RecordSuppression`
  (FR-003).
- [X] T003 Tests: SC-001 (slept-in-queue ⇒ zero runLoop invocations, one
  suppressed outcome), dead-agent parity, mirror-updates-on-batch, and the
  FR-003 counter non-bump.

## Phase 2: In-flight cancel (US1 layer 2)

- [X] T004 Per-agent race-safe cancel slot: `runPlan` registers its call
  context's cancel before `md.runLoop` and clears it after; absorb invokes it
  on `agent.slept` / `agent.died` for that agent (FR-004). Terminal outcome
  reason is sleep/death-attributable, distinct from `callTimeout`; planner
  slot only — consolidation/narrator/meeting/reconcile/scene workers
  untouched.
- [X] T005 Tests: SC-002 (mid-call `agent.slept` cancels, nothing lands,
  reason attributable), consolidation-still-runs-on-slept, and the FR-005
  verification (cancel consumes no transport retry, no estimator spike
  adoption — adversarial check; fix only if a real path exists).

## Phase 3: Wake resumption + regression (US2)

- [X] T006 Regression tests: SC-005 — `agent.woke` re-arms and the next
  planner call proceeds past the gate (mirror awake in the same batch);
  `plan()`'s enqueue-time gate and `internal/sim/` byte-unchanged (SC-004
  zero-diff assertion is review-level; the test pins behavior).

## Phase 4: Soak evidence (card AC #2)

- [ ] T007 Seeded soak (≥ 3 game-days, measurement-run dials): count
  "is asleep" `agent.intent_rejected` per game-day (target ≤ 1 vs baseline
  ~31) and planner `cog.thought` rows for asleep-at-submit agents (target 0);
  record the counts on the board task (SC-003).
  **PARTIAL** — measurement queries implemented + documented in
  [soak.md](soak.md); the live soak needs a real local planner model over
  ≥ 3 game-days (an operator measurement run, not runnable meaningfully from
  the implementation environment). Remaining: the run itself + counts on the
  board card.

## Phase 5: Reconcile with spec 102 + grounding + gates

- [ ] T008 After TASK-112's PR lands (operator ruling: this PR merges only
  after it): merge main INTO this branch (never rebase), re-run
  `go test -race ./...`, re-verify the absorb/runPlan seams.
- [ ] T009 Re-verify + re-pin touched wiki notes (behavior-review:
  mind-driver-triggers, tool-use-dispatch, planner-telemetry; computed
  re-verifies for the other mind.go/telemetry.go-sourced notes); regenerate
  `docs/player/` if wiki changed; `node scripts/check-merge-drift.mjs pr`
  exit 0.
