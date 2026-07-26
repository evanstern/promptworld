# Tasks: Neglect detector — critical need with zero intents in its class

**Input**: Design documents from `/specs/083-neglect-detector/`
**Prerequisites**: spec.md, research.md, data-model.md, plan.md

**Tests**: included alongside code; tier is Opus 4.8 per the board record
(reducer/percept event + high-salience memory injection + world-01 validation;
cognition-adjacent).

**Organization**: phases map to the board ACs — Phases 2–3 ↔ AC #2 (deterministic
percept event + high-salience memory, replay-visible), Phase 4 ↔ AC #1 (validated
against Oak's death window / silent on healthy windows), Phase 5 ↔ the reorient move-13
alert obligation (shipped severity grammar: chronicle whole-line alert + map overlay).
Board AC #3 (composition with TASK-111 survival watches) is satisfied by spec.md's
Composition section — no code task exists for it by design. Phases 6–7 are the
wiki-in-PR obligations and close-out.

## Phase 1: Setup

- [X] T001 Worktree ready (claim already landed per TASK-160 flow): from
      `.worktrees/task-133`, merge fresh `origin/main` INTO the branch if stale (never
      rebase), then baseline `go test ./internal/sim/ ./internal/tui/` green before
      changes

## Phase 2: Foundational — derived state, class dictionary, reducer arms (blocks all)

- [X] T002 `NeglectState` struct + `Neglect *NeglectState` omitempty pointer on `Agent`
      + need-keyed accessors + `neglectWindowTicks = 7200` doctrine const block, in
      `internal/sim/agents.go` (plan D1, data-model §1/§5, spec FR-001/005)
- [X] T003 `needClassGoals` map + `needClassOf` beside the goal-resolver registry, in
      `internal/sim/policy.go`; `TestNeedClassGoalsResolve` pins every member resolves
      (anti-rot), in `internal/sim/policy_test.go` (plan D1, data-model §6, spec FR-003)
- [X] T004 Extend the `agent.needs_changed` arm (band anchors set/clear + latch clear
      on recovery) and the `agent.intent_set` arm (class stamp after `appendIntent`);
      add the `sim.neglect_detected` arm (Fired latch), in `internal/sim/state.go`
      (~1718 / ~845); arm unit tests in `internal/sim/neglect_test.go` (plan D1,
      data-model §2, spec FR-001)
- [X] T005 `rebaseTicks` shifts the six `*Since`/`*Intent` anchors (non-zero only) +
      taxonomy doc-comment rows, in `internal/sim/miracles.go`; extend the rebase test
      (plan D1, data-model §7, spec FR-009)

## Phase 3: User Story 1 — the deterministic percept + high-salience memory (P1, board AC #2)

**Goal**: one `sim.neglect_detected` + one salience-9 companion memory per episode,
replay-visible, reducer-only writes.

**Independent Test**: scripted-history fold → sweep fires once, memory added,
generation bumps; genesis replay hash-identical.

- [X] T006 [US1] `NeglectDetectedPayload` beside `NeedsPayload`/`DiedPayload` in
      `internal/sim/agents.go` (~1262) (plan D2, data-model §3, spec FR-004)
- [X] T007 [US1] Factored pure predicate `NeglectDue(a *Agent, need string, tick int64)
      bool` (pre-tick state only, exported for the D5 probe) + the heartbeat sweep in
      the `%60` block beside the near-death latch — awake-only, fixed need order, event
      then companion memory in-batch — in `internal/sim/executor.go` (plan D2, spec
      FR-004/008)
- [X] T008 [US1] `salNeglect = 9` in the salience table (interrupt-band rationale
      comment — research R6) + the three fixed per-need voice-of-evidence texts
      (`OriginWitness`, `Why` empty), in `internal/sim/memory.go` (plan D2, data-model
      §4, spec FR-006)
- [X] T009 [US1] Mechanism tests: fires exactly once per episode (second window
      silent); re-fires after recovery + relapse; class intent inside window defers;
      asleep skipped at the beat; generation bump lands; live-vs-replay hash identity
      with the detector's own events in the log (`governor_replay_test.go` idiom);
      snapshot byte-identity for pre-083 fixtures — in `internal/sim/neglect_test.go`
      (plan D4, spec US1 scenarios, SC-003/004/006)

## Phase 4: User Story 2 — validated against Oak's death window, silent on healthy (P1, board AC #1)

**Goal**: the documented world-01 evidence becomes the regression corpus — in-repo
fixtures binding, real-log probe as recorded evidence.

**Independent Test**: `go test ./internal/sim/` (fixtures);
`PROMPTWORLD_WORLD01_DB=… go test ./internal/daemon/ -run Neglect` (probe).

- [X] T010 [US2] Oak-shaped recorded fixture (warmth 636→0 at 4/min, only reflex
      `chop` + planner `wander` records): fires at band-entry + 7200 — before the
      trajectory's death tick, health ≈ 900 at firing (runway assertion) — in
      `internal/sim/neglect_test.go` (plan D4, spec FR-007, SC-001)
- [X] T011 [US2] Healthy fixtures: class-intent-inside-window (Oak-day-4 shuttling
      shape) silent; dip-and-recover-before-T silent with anchors reset — in
      `internal/sim/neglect_test.go` (plan D4, spec FR-007, SC-001)
- [X] T012 [US2] Env-guarded world-01 probe (the `TestSageThrashWindowContextReplay`
      copy-and-replay idiom + `replayToTick`): predicate true at sampled ticks in Oak's
      final ~6 h, false on sampled labeled-healthy episodes and Ash/Hazel; skips
      without `PROMPTWORLD_WORLD01_DB`; record one run's output as task evidence — in
      `internal/daemon/` beside the Sage test (plan D5, spec FR-008, SC-002)

## Phase 5: User Story 3 — shipped severity channels: chronicle alert + map overlay (P2, reorient move-13 obligation)

**Goal**: the alert enters existing tiers only — the five-touch `stranger.took`
precedent; no new tier, channel, or token.

**Independent Test**: `TestCatalogSweep` green; alert render + overlay-subsumption
tests green.

- [X] T013 [US3] `"sim.neglect_detected"` case in `isAlertType`
      (`internal/tui/grammar.go`) + `digestRegistry` entry with deterministic per-need
      wording (`internal/tui/digest.go`) + `catalogFixture` row
      (`internal/tui/digest_test.go`) (plan D3, spec FR-010/011)
- [X] T014 [US3] Backticked `sim.neglect_detected` mention in
      `docs/wiki/event-types.md` (TestCatalogSweep doc↔catalog anti-drift) + the §3
      wording row in `specs/018-chronicle-digest/contracts/digest-grammar.md` (plan D3,
      spec FR-011)
- [X] T015 [US3] Render tests: chronicle row renders whole-line `styleFeedAlert`; a
      neglect-state agent fixture paints `styleAgentCritical` on the map grid (the
      FR-012 subsumption pin — no tiles.go change), in `internal/tui/` tests (plan D3,
      spec FR-012, SC-005)
- [X] T016 [US3] Amend `docs/design/tui/patterns/chronicle-grammar.md` (alert-tier
      enumeration + color-roles `alert` row), `panels/chronicle.md`, `panels/map.md`
      (condition overlays name neglect); then
      `node scripts/check-tui-design.mjs --changed` from the worktree — re-verify +
      re-pin every flagged page; gate passes (plan D6, spec FR-013, SC-006)

## Phase 6: Grounding — wiki-in-PR obligations (in-branch, pr-gate enforced)

- [ ] T017 `/grounding-wiki:wiki-update` reconciliation over the branch diff;
      review-work re-pins expected on `docs/wiki/executor-needs-survival.md`,
      `sim-state-agent-fields.md`, `sim-state-apply-agents.md`,
      `sim-state-intent-lifecycle.md`, `reflex-policy.md`, `agent-memory-window.md`
      (the "kept below 9 on purpose" sentence now has a deliberate exception),
      `event-types.md` + routed child note, `guardian-miracle-rebase-taxonomy.md`,
      `guardian-survival-watches.md` (composition seam, per judgment), and the
      chronicle/tiles TUI notes; computed re-pins for the rest — all pinned to branch
      commits (plan D7, spec SC-007)
- [ ] T018 Regenerate `docs/player/` via the `player-docs` skill (wiki changed in
      T017); run the probe directly:
      `node .claude/skills/player-docs/scripts/check-freshness.mjs --check` passes
      in-branch (plan D7, spec SC-007)

## Phase 7: Polish & close-out

- [ ] T019 Full proof from the worktree: `gofmt -l` clean; `go test ./...` green
      (TestCatalogSweep included; pre-083 snapshots byte-identical);
      `node scripts/check-merge-drift.mjs pr` exits 0; PR opens carrying code + design
      + wiki + player docs together; merge via `gh pr merge --merge` ONLY (squash
      rewrites in-branch pins — observed hazard) (spec SC-006/007)
- [ ] T020 Post-merge bookkeeping — TASK-160 all-by-merge: board AC ticks, spec-bridge
      sync, tasks.md ticks, runbook execution-log row are authored on a branch
      (task branch or short-lived bookkeeping branch) and land on main via merge —
      derived state only, no grounding content, nothing committed directly at root

## Dependencies & Execution Order

- T001 → everything. T002 → {T003, T004, T005, T006}. T003+T004 → T007 (sweep reads
  anchors + dictionary). T007+T008 → T009 → {T010, T011}. T007 (exported predicate) →
  T012. T006+T007 → T013 → T014 → T015; T016 after T013–T015.
- T017 after all code (T002–T016); T018 after T017; T019 after everything in-branch;
  T020 post-merge.
- **MVP** = Phases 2–4 (the detector, injected and validated — board ACs #1+#2);
  Phase 5 completes the move-13 alert obligation. AC #3 is already satisfied by
  spec.md §Composition.

**Parallel opportunities**: T003 ∥ T004 ∥ T005 (different files); T010/T011 authoring ∥
T013–T015 (sim tests vs tui membership); T016 page prose can start once T013's wording
is fixed.
