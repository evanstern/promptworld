# Tasks: Exercise catalog wave — exercises, incidents, lessons

**Input**: Design documents from `/specs/077-exercise-catalog/`
**Prerequisites**: spec.md, plan.md, research.md, data-model.md

**Tests**: included alongside code. Tier: **Opus 4.8** per the board record (new
reducer-valid event kinds across sim reducer + chronicle digest + scenario evaluators;
replay-safety doctrine-adjacent).

**Organization**: phases map to the board ACs — Phases 2–3 ↔ AC #2 (incident
vocabulary: reducer-valid, ambient-indistinguishable, severity-grammar surfaced),
Phases 4–5 ↔ AC #1 (2–3 exercises per stage with production evaluators + pass
emission), Phase 6 ↔ AC #3 (lesson tranche 2 + wrong-thing detector). Phases 7–9 are
design-authority, grounding, and close-out gates.

## Phase 1: Setup

- [ ] T001 Cut the task worktree from fresh origin/main and baseline:
  `node scripts/check-merge-drift.mjs worktree --spec 077 --task TASK-151`, then
  `git worktree add .worktrees/task-151 -b task-151-exercise-catalog origin/main`;
  `go test ./internal/sim/ ./internal/tui/` green before changes (worktree
  `.worktrees/task-151`)

## Phase 2: Foundational — state, reducer arms, evidence coordinates (blocks all later phases)

- [ ] T002 New state fields per data-model §3 (`ColdSnapUntil`, `Stranger`,
  `StrangerTakes`, `CharterObservedSeq/Tick`, `SkillsFingerprint/SkillsObservedSeq/Tick`)
  with zero-value doc comments, in `internal/sim/state.go`; snapshot byte-identity test
  (pre-077 fixture round-trips unchanged), in `internal/sim/state_test.go` (plan D2,
  spec FR-005/006/010/012, edge "pre-077 world")
- [ ] T003 Reducer arms: `sim.cold_snap` (latch `ColdSnapUntil`), `sim.forage_blighted`
  (append `Harvest` overlays, idempotent re-apply, validate-at-door), and the
  `metatron.charter_observed` arm's `CharterObservedSeq/Tick` stamp (envelope, the
  `PlacedSeq` precedent), dispatched from `internal/sim/state.go` into
  `internal/sim/{executor-adjacent files per house layout}`; per-arm apply + genesis
  replay round-trip tests (plan D2, spec FR-010/011/005)
- [ ] T004 `metatron.skills_observed`: payload struct + `applyGuardian` arm persisting
  fingerprint + Seq/Tick, in `internal/sim/guardian.go`; turn-time observation emission
  mirroring `charter_observed` (fingerprint the bound skill set post-`stageSkills`; emit
  on change; empty set never emits — stages 1–2 structurally silent), in
  `internal/guardian/`; tests both sides (plan D6, spec FR-006)
- [ ] T005 Evidence constructors: `CharterEvidenceFromState` (omit when Seq==0 — pre-077
  honesty) and `SkillsObservedEvidence` (`Custom: true` by construction), with
  doc-comment honesty contracts beside `CharterObservedEvidence`/`OrderPlacedEvidence`,
  in `internal/sim/curriculum.go`; constructor tests (plan D3, spec FR-004, research R7/R8)
- [ ] T006 Rebase taxonomy: classify every new tick-anchored field (data-model §3 —
  `ColdSnapUntil`/`Stranger.LastMove/LastTake` SHIFT; take/observation ticks KEEP) in
  `rebaseTicks`, with tests beside the existing taxonomy suite, in
  `internal/sim/miracles.go` + `internal/sim/miracles_test.go` (plan D5, spec FR-015,
  edge "time snap past an incident window")

## Phase 3: User Story 2 — incident vocabulary grows to three new kinds (P1, board AC #2)

**Goal**: `cold_snap`, `forage_blight`, `stranger_arrives` fire from authored schedules
as reducer-valid, ambient-indistinguishable, replay-safe events.

**Independent Test**: per-kind scheduled-fixture emission + precondition + window-lapse
+ replay byte-identity suites.

- [ ] T007 [US2] Kind constants + `compileIncident` arms with param validation
  (`Hours` [1,24], `Radius` [1,8]) + `IncidentScheduleEntry.Radius/Hours` fields, in
  `internal/sim/scenario.go`; compile-error table tests in
  `internal/sim/scenario_test.go` (plan D1, spec FR-009, data-model §1)
- [ ] T008 [US2] Named precondition predicates (`coldSnapActive`, `blightableTiles` with
  deterministic row-major tile walk, `strangerEntryValid`) + `scenarioIncidentEvents`
  emission arms per data-model §2, with the TASK-28 ambient/preemption seam documented
  beside `gruScheduledTonight`, in `internal/sim/scenario.go` (plan D1, spec
  FR-009/013/014, research R1)
- [ ] T009 [US2] Cold-snap severity mechanics: harsher outdoor night warmth loss while
  `tick < ColdSnapUntil` through the existing `warmthLossCold` arithmetic path (rate
  constant beside it), in `internal/sim/agents.go`; heartbeat test proving snap-window
  vs ambient-night rates and read-time expiry, in `internal/sim/day_warmth_test.go` or
  sibling (plan D2, spec FR-010, research R2)
- [ ] T010 [US2] The stranger entity: `internal/sim/stranger.go` — `strangerStep` from
  `stepEvents` (order pinned by test: after `gruStep`, before the social beat),
  greedy store-seeking movement on `rngAt("stranger-prowl"/"stranger-take")`, bounded
  takes mutating pile/chest stock via the agent-withdrawal state shapes +
  `StrangerTakes` ring (32), fire-light/shelter avoidance via predicates shared with the
  gru (extracted, not duplicated), dawn departure, witness/victim-adjacent situated
  memories (rumor fuel); `applyStranger` dispatch in `internal/sim/state.go` (plan D2,
  spec FR-012, research R4)
- [ ] T011 [US2] Per-kind behavior suites: emission at authored tick, precondition-failed
  silent skip (blight on exhausted patch; stranger with one abroad; snap during snap),
  window lapse after time-snap, once-only via state latch, stranger+gru same-night
  independence, and genesis-replay byte-identity per kind, in
  `internal/sim/scenario_test.go` + `internal/sim/stranger_test.go` (spec US2 scenarios,
  edges; SC-002)
- [ ] T012 [US2] Ambient-world regression proof: a no-scenario world's event stream under
  the new binary is byte-identical to pre-077 (determinism harness), in
  `internal/sim/sim_test.go` or sibling (spec FR-017, SC-002)
- [ ] T013 [US2] Digest grammar: `familyByNamespace["stranger"]` (gru/threat voice),
  `stranger.took` into the whole-line alert set, `digestRegistry` + `catalogFixture`
  rows for all seven new types (data-model §2; blight uses first-fact-plus-count), in
  `internal/tui/grammar.go` + `internal/tui/digest.go` + `internal/tui/digest_test.go`;
  `TestCatalogSweep` green (spec FR-016, SC-003, plan D7)

## Phase 4: User Story 1a — emitter generalization (blocks 4b; board AC #1)

- [ ] T014 [US1] `ExerciseDefinition.BoundaryDay` + `boundaryDue(def, nextTick)` (fixed
  dawn-of-day-N / rolling every-dawn from day 2); `scenarioRubricEvents` drops the
  `first-night`-only guard, keeps batch-death scan + `hasCurriculumPass` latch +
  pass→unlock same-batch order, and assembles evidence via the sanctioned constructors
  keyed by satisfied term type (data-model §4), in `internal/sim/scenario.go` +
  `internal/sim/curriculum.go` (plan D3, spec FR-003/004)
- [ ] T015 [US1] Emitter generalization tests: first-night byte-identical behavior
  (regression), rolling boundary passes at first satisfying dawn and never re-emits,
  fixed-boundary miss emits nothing forever (`in_progress` until `run.ended`), all-dead
  dawn suppressed for every exercise, in `internal/sim/scenario_test.go` (spec US1
  scenario 6, edges)

## Phase 5: User Story 1b — nine exercises with production evaluators (P1, board AC #1)

**Goal**: catalog 3/2/2/2 by stage; every exercise evaluates and emits for real.

**Independent Test**: per-exercise `driveTicks` pass fixture + the no-default-arm sweep.

- [ ] T016 [US1] Evaluator helpers (`deathsByCause`, `storedFoodTotal`,
  `playerOrderSince`; death-cause field names verified against the spec-044 ledger) +
  seven new rubric arms in `EvaluateRubric` per data-model §5, in
  `internal/sim/scenario.go`; per-arm table tests (the `TestTheLawRubricTable`
  precedent), in `internal/sim/scenario_test.go` (plan D4, spec FR-002)
- [ ] T017 [US1] The seven new `ExerciseDefinition`s + `BoundaryDay` on the two shipped
  ones (first-night: 2; the-law: 0/rolling), seeds 46103–46109, schedules per
  data-model §5, in `internal/sim/curriculum.go`; `TestScenarioSchedulesCompile` stays
  green; position-validity pins per authored position on its own seed's map (the
  `TestFirstNightSchedulePositionValid` precedent), in `internal/sim/scenario_test.go`
  (spec FR-001/008)
- [ ] T018 [US1] Catalog sweeps: no cataloged id reaches `EvaluateRubric`'s default arm
  (new sweep test); `TestExerciseRubricTermsAreCatalogedEventTypes` green over the grown
  terms (needs T013), in `internal/sim/scenario_test.go` + `internal/tui/digest_test.go`
  (spec FR-002, SC-001)
- [ ] T019 [US1] `world.ValidScenarioExercise` mirrors all nine ids;
  `TestScenarioVocabularyMirrorsSimCatalog` green, in `internal/world/world.go` (+ its
  test) (spec FR-007)
- [ ] T020 [US1] Pass-emission fixtures per exercise class (`driveTicks` harness):
  the-law with `CharterEvidenceFromState` evidence → same-batch `stage_unlocked{stage-3}`
  (US1 scenario 3); toolsmith with `Custom: true` skills evidence → stage-4 unlock (US1
  scenario 4); a stage-4 pass with no unlock (scenario 5); cold-dawn/stranger-at-the-gate
  full pass paths under their incidents; pre-077-snapshot degradation (charter Seq==0 →
  evidence omitted, pass waits — edge case), in `internal/sim/scenario_test.go` +
  `internal/sim/curriculum_test.go` (spec SC-001, edges)

## Phase 6: User Story 3 — lesson tranche 2 + wrong-thing detector (P2, board AC #3)

- [ ] T021 [US3] The `lessonFold` seam (bounded per-reason rejection counter, cap 32) +
  optional `FoldTrigger` on `lessonEntry`, per-event `Trigger` predicates untouched, in
  `internal/tui/lessons.go` (plan D7, spec FR-019, research R10)
- [ ] T022 [US3] Four catalog entries per data-model §6 (`first-explain-answer`,
  `first-report-card`, `first-skill-file`, `same-refusal-pattern`), skin-tokened,
  prompting tier, in `internal/tui/lessons.go`; projection tests: each surfaces once,
  never again; fold triggers on 3 same-reason rejections, never on mixed; catalog↔help
  1:1 population test green at 12 entries, in `internal/tui/lessons_test.go` (spec
  FR-018/021, SC-004)

## Phase 7: Design authority — pages amended, gate green

- [ ] T023 Amend `docs/design/tui/patterns/chronicle-grammar.md` (seven new digest rows +
  `stranger.took` alert-tier addition), `overlays/help.md` ("(8 catalog entries)" → 12),
  `panels/lesson-row.md` (tranche 2), `panels/exercise.md` (nine-exercise catalog +
  incident vocabulary/params table), `panels/map.md` (stranger glyph, gru precedent;
  follow spec 068's single-source rule if the registry conventions apply) (spec FR-022,
  plan D8)
- [ ] T024 `node scripts/check-tui-design.mjs --changed` from the worktree: re-verify +
  re-pin every flagged page; gate passes (spec SC-005)

## Phase 8: Grounding — wiki-in-PR obligations (in-branch, pr-gate enforced)

- [ ] T025 `/grounding-wiki:wiki-update` reconciliation over the branch diff;
  body-amendment re-pins expected on `docs/wiki/scenario-machinery.md` (kind vocabulary,
  emitter generalization, 9-exercise operational notes),
  `scenario-machinery-surfacing.md`, `curriculum-ladder-progression.md` (the-law
  emission real; stage-3 evidence design landed), `event-types.md` + children (new
  incident/skills rows — a new `event-types-scenario-incidents.md` child if the existing
  children don't fit; parent backticks all new types for the sweep),
  `event-types-curriculum-events.md`, `executor-tick-subsystems.md` (strangerStep),
  `gru.md` (shared protection predicates), `tui-chronicle-feed.md` (rows + alert set),
  `tui-input-help.md`/`tui-client.md` (lessons 12), `sim-state-world-fields.md`,
  `world-save-manifest-fields.md` (vocabulary mirror), guardian children
  (`guardian-instruction-surface.md` — skills observation); computed re-pins for every
  other note listing touched sources — all pinned to branch commits (spec FR-023, SC-006)
- [ ] T026 Regenerate `docs/player/` via the `player-docs` skill (wiki changed in T025);
  `node .claude/skills/player-docs/scripts/check-freshness.mjs --check` passes in-branch
  (spec FR-023, SC-006)

## Phase 9: Polish & close-out

- [ ] T027 Full proof: `gofmt -l` clean; `go test ./...` green (snapshots byte-identical;
  ambient-world regression T012 green); `node scripts/check-merge-drift.mjs pr` from the
  worktree exits 0; PR opens carrying code + design + wiki + player docs together; merge
  via `gh pr merge --merge` only (spec SC-005/006, FR-024)
- [ ] T028 Post-merge (root): spec-bridge sync, board AC ticks, tasks.md ticks, runbook
  execution-log row, AND the `first-faith-event` rider note appended to TASK-118 via the
  `backlog` CLI — derived state only, no grounding content on main (spec FR-020/024)
