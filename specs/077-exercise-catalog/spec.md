# Feature Specification: Exercise catalog wave — 2–3 exercises per stage + incident vocabulary growth

**Feature Branch**: `077-exercise-catalog` (task branch: `task-151-exercise-catalog`)

**Created**: 2026-07-26

**Status**: Draft

**Input**: TASK-151 / reorient 2026-07-26 decision 5 (docs/design/reorient-2026-07-26-ui.md,
merged position 5 — "content is now the constraint, not surfaces"). Depends on TASK-149
(spec 072, merged — rubric truth + the-law production evaluator). Tier: Opus 4.8 (recorded
on the board task: new reducer-valid event kinds across sim reducer + chronicle digest +
scenario evaluators; replay-safety doctrine-adjacent).

## Problem (pinned)

The loop is built but the game is one lesson long. The real current state, inventoried
from code:

- **Exercises**: two ship (`sim.ScenarioExercises`, `internal/sim/curriculum.go:388`).
  `first-night` (stage-1, seed 46101) is complete: production evaluator
  (`firstNightRubric`), pass emission (`scenarioRubricEvents`), and an authored schedule
  (one `gru_emerges` entry). `the-law` (stage-2, seed 46102) has a production EVALUATOR
  since spec 072 (`theLawRubric` over `State.Norms` + the persisted `State.CharterCustom`
  flag TASK-149 added) but **no boundary tick, no evidence assembly, and no pass
  emission** — `scenarioRubricEvents` carries an explicit `first-night`-only guard
  (`internal/sim/scenario.go:415-424`, spec 072 FR-009: "exercise-catalog content work,
  not this guard's"). Stages 3 and 4 have zero exercises; the stage-2→3 unlock therefore
  still needs `--override`, and the stage-3→4 gate's "player-granted tool's contributing
  act" evidence (deferred as "TASK-119's exercise design" in `EvaluateUnlock`'s doc) has
  never been designed.
- **Incidents**: the closed vocabulary is ONE kind, `IncidentGruEmerges`
  (`internal/sim/scenario.go:26-31`). The director seam (`incidentSource`) and the
  schedule compiler exist and are exactly the extension points the file's own comment
  names ("a new constant here, a compile arm in compileIncident, and an emission arm in
  scenarioIncidentEvents").
- **Lessons**: eight catalog entries (`internal/tui/lessonCatalog`, 5 mechanics + 3
  prompting). Nothing teaches the spec-063 feedback surfaces (explain, report card), the
  stage-3 skill-file concept, or the wrong-thing pattern (repeated same-cause refusals).

This spec is the delivery vehicle for decision 5: 2–3 hand-authored exercises per stage
with production evaluators, the incident vocabulary grown to ~3 new kinds entering through
the shipped severity grammar (no new channels), and lesson tranche 2 plus the first
wrong-thing detector riding as content.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Every stage has 2–3 exercises with production evaluators (Priority: P1)

A player runs `promptworld new --scenario <id>` for any of nine cataloged exercises —
three at stage-1, two each at stages 2–4. Every exercise's exercise-tab gauges evaluate
for real over state facts (no permanently-pending production exercise), and every exercise
EMITS `curriculum.exercise_passed` at its boundary when its rubric holds — including
`the-law`, whose pass emission (boundary, evidence, emit) this spec completes, unblocking
the stage-2→3 unlock without `--override`, and the new stage-3 exercises, whose
`Custom` evidence opens the stage-3→4 gate for the first time.

**Why this priority**: board AC #1 and the reorientation's core diagnosis — the ladder
exists but has almost nothing to climb. Everything else in this spec supplies pressure
(US2) or narration (US3) for these exercises.

**Independent Test**: per-exercise deterministic fixtures (the `driveTicks` harness) reach
each exercise's pass boundary with its terms satisfied and assert
`curriculum.exercise_passed` (with correct evidence) plus, where the gate grants,
`curriculum.stage_unlocked`; a catalog sweep test asserts NO cataloged exercise falls
through to `EvaluateRubric`'s default pending arm.

**Acceptance Scenarios**:

1. **Given** the shipped catalog, **When** `sim.ScenarioExercises` is enumerated by
   stage, **Then** stage-1 has 3 entries, stages 2–4 have 2 each (9 total), every entry
   compiles (`TestScenarioSchedulesCompile`) and every rubric term names a cataloged
   event type (`TestExerciseRubricTermsAreCatalogedEventTypes`).
2. **Given** any cataloged exercise, **When** `EvaluateRubric(s, def, tick)` runs,
   **Then** a production arm evaluates every term from state facts — the default
   render-pending arm is reached by NO cataloged id (sweep-tested).
3. **Given** a `the-law` world where a norm is adopted while a player-authored charter is
   in force, **When** the next dawn boundary arrives, **Then**
   `curriculum.exercise_passed` is emitted carrying a `metatron.charter_observed`
   evidence ref with `custom: true` (re-located via the newly persisted observation
   Seq/Tick coordinates), and `curriculum.stage_unlocked{stage-3}` lands in the same
   batch.
4. **Given** a stage-3 `toolsmith` world where a player skill file is bound and observed
   and the guardian subsequently places an order, **When** the dawn boundary arrives with
   zero deaths, **Then** the pass's evidence includes a `Custom: true` entry and the
   stage-3→4 gate grants (`EvaluateUnlock` stage-3 conjunct satisfied for the first time
   by production machinery).
5. **Given** a stage-4 exercise pass, **Then** `curriculum.exercise_passed` is recorded
   but no `curriculum.stage_unlocked` follows (stage-4 is graduation — existing
   `nextLadderStage` behavior, unchanged).
6. **Given** a fixed-boundary exercise whose terms are unmet at its boundary dawn,
   **Then** nothing is emitted (failure is never an event; `run.ended` remains the sole
   fail signal, and `ExerciseOutcome` continues to report `in_progress` on a live world).

---

### User Story 2 - The incident vocabulary grows to cold snap, forage blight, and stranger arrival (Priority: P1)

An exercise author (and, post-TASK-28, the ambient director) can schedule three new
incident kinds beside `gru_emerges`: **`cold_snap`** (a bounded window of harsher
night cold), **`forage_blight`** (a patch of forage stricken barren until a far regrow
deadline), and **`stranger_arrives`** (a trickster entity that slips in at night, takes
from unattended stores, and is gone by dawn). Each lands as reducer-valid events
indistinguishable from an ambient cause — no `scenario`/`authored` marker in any payload,
preconditions identical to what an ambient dice path would check (the `gru.emerged`
precedent) — and surfaces through the SHIPPED severity grammar: digest rows in existing
family voices and tiers, no new alert channel, no new push surface.

**Why this priority**: board AC #2. The exercises of US1 are inert without authored
pressure; three kinds is the minimum vocabulary that makes stages 2–4 feel different from
stage-1.

**Independent Test**: per-kind reducer round-trip tests (emit → apply → replay from
genesis → byte-identical state); a scheduled-incident fixture per kind asserting the
emission, its preconditions, its window lapse (time-snap past the window skips silently),
and its once-only state latch; `TestCatalogSweep` green over the new event types.

**Acceptance Scenarios**:

1. **Given** an armed schedule with a `cold_snap` entry, **When** its tick arrives,
   **Then** `sim.cold_snap` is emitted and the reducer latches `State.ColdSnapUntil`;
   while the snap holds, outdoor night warmth loss runs at the harsher cold-snap rate
   through the same needs-heartbeat arithmetic ambient night cold already uses; past
   `ColdSnapUntil` behavior reverts with no end event (read-time check, the belief-decay
   precedent).
2. **Given** an armed `forage_blight` entry over a patch containing unharvested forage,
   **When** it fires, **Then** one `sim.forage_blighted` event carries the stricken tile
   list (deterministic walk order) and the reducer marks each tile with the EXISTING
   harvested-regrow overlay at the blight's far deadline — a villager (and its mental
   map) experiences exactly what heavy picking already produces; a patch with no
   unharvested forage skips silently (precondition-failed class, never retried).
3. **Given** an armed `stranger_arrives` entry, **When** it fires, **Then**
   `stranger.arrived` places a positioned entity (`State.Stranger`, the gru's
   entity-not-phenomenon precedent) that moves (`stranger.moved`) toward unattended
   stores on deterministic per-decision RNG (new purpose tags, never a stream), takes
   bounded goods (`stranger.took`, appended to a bounded `State.StrangerTakes` ledger),
   never enters firelight or shelter (the gru's protection rules), and departs at dawn
   (`stranger.departed`, state nil).
4. **Given** any of the three kinds, **Then** its payload carries NO field
   distinguishing an authored emission from an ambient one, and its emission-time
   preconditions (no snap active / unharvested forage present / no stranger abroad +
   passable unprotected entry tile) are expressed as reusable predicates an ambient
   emitter (TASK-28) can call unchanged.
5. **Given** a world with recorded incidents of all three kinds, **When** its log is
   replayed from genesis (no scenario runtime armed), **Then** the folded state is
   byte-identical to the live run — recorded events are the only persistence.
6. **Given** an ambient (no-scenario) world, **Then** its tick path and state bytes are
   unchanged by this feature — no ambient emission of the new kinds ships in 077 (the
   ambient dice paths are TASK-28's recorded scope; see Assumptions).
7. **Given** the chronicle raw feed, **Then** every new event type renders through a
   `digestRegistry` row in an existing family voice — `sim.*` in the sim family,
   `stranger.*` in the gru/threat family — with `stranger.took` joining the existing
   whole-line alert tier beside `social.chest_taken` (theft is theft); no new tier, no
   new channel.

---

### User Story 3 - Lesson tranche 2 + the first wrong-thing detector land as catalog content (Priority: P2)

A player sees, at most once each and at the moment they first become true: a lesson when
the guardian first answers through `explain` (the grounded-feedback surface), a lesson
when the first report card is produced, a lesson when their first skill file is observed
in force (stage-3), and — the first wrong-thing detector — a lesson when the guardian's
tool calls have been refused three times for the SAME stated reason, pointing the player
at the decision trace and their charter rather than at the refusal itself.

**Why this priority**: board AC #3. Pure client-side content on the shipped spec-055
machinery; valuable alone but its skill-file entry depends on US1's new
`metatron.skills_observed` event.

**Independent Test**: `lessonTriggers` unit tests per new entry (trigger event → surfaces
once, never again); a fold test for the repeated-refusal detector (two same-reason
refusals → nothing; third → surfaces; three different-reason refusals → nothing); the
existing catalog↔help-overlay 1:1 population test extended to the grown catalog.

**Acceptance Scenarios**:

1. **Given** a `cog.tool_call` with tool `explain` and a read-ok verdict, **When**
   ingested, **Then** `first-explain-answer` surfaces (prompting tier), once per player.
2. **Given** a `guardian.report_card` event, **Then** `first-report-card` surfaces
   (prompting tier).
3. **Given** a `metatron.skills_observed` event (US1's new type), **Then**
   `first-skill-file` surfaces (prompting tier).
4. **Given** three `cog.tool_call` events with `rejected_*` verdicts carrying the same
   non-empty `Reason`, **When** the third is ingested, **Then** `same-refusal-pattern`
   surfaces — the detector's fold is bounded, session-local, and resets nothing else;
   mixed reasons never trigger it.
5. **Given** the grown catalog, **Then** the help overlay's lessons section lists all 12
   entries 1:1 (`populateHelpLessons`), every guardian reference is skin-tokened, and
   `docs/design/tui/overlays/help.md`'s "(8 catalog entries)" row plus
   `panels/lesson-row.md` are amended in the same PR.
6. **"First faith event" is explicitly OUT of scope**: TASK-118 (faith) has not run —
   there is no faith event type to trigger on. The entry is recorded as a content rider
   on TASK-118 (the catalog is append-only; it lands there), never stubbed here.

---

### Edge Cases

- **Incident firing on a world whose player never earned the stage** (`--stage N
  --override` or a scenario world created via the override path): incidents read ONLY the
  armed scenario — earned-state is a per-user convenience record that world behavior
  never reads (spec 046 doctrine, unchanged). The incident fires; only the exercise
  panel's incident-visibility mode differs by manifest stage
  (forecast stages 1–2, fog from stage-3 — `IncidentVisibilityFor`, unchanged).
- **Replay of a world with recorded incidents**: replay arms no scenario runtime
  (`State.scenario` is unexported, never serialized) — the recorded `sim.cold_snap` /
  `sim.forage_blighted` / `stranger.*` events re-apply through reducer-total arms and
  reproduce the state exactly; no arm may consult the schedule, the manifest, or wall
  clock.
- **Evaluator on a world created pre-077**: every new state field
  (`ColdSnapUntil`, `Stranger`, `StrangerTakes`, `CharterObservedSeq/Tick`,
  `SkillsFingerprint/SkillsObservedSeq/Tick`) is `omitempty` with a conservative zero
  value — a pre-077 snapshot round-trips byte-identically (no `format_version` bump, the
  spec-072 `CharterCustom` precedent) and evaluators render honestly pending/unmet, never
  a false ✓.
- **A pre-077 `the-law` world resumed under the new daemon**: pass emission begins
  evaluating at the next dawn boundary. If the rubric already holds, the pass lands at
  that dawn — a legitimately earned, late-recorded pass, not retroactive invention (the
  evidence coordinates come from state persisted by the charter arm; a pre-077 snapshot
  without persisted Seq/Tick renders the evidence entry absent and the pass waits for the
  next charter observation to stamp them — honest degradation, self-healing on the next
  charter revision, the spec-072 zero-value posture).
- **Time snap past an incident window**: skipped silently, never retried (the shipped
  `windowEnd` contract, per kind). `ColdSnapUntil` and every new tick-anchored field is
  classified in the miracle rebase SHIFT/KEEP taxonomy
  (`guardian-miracle-rebase-taxonomy`): a snap forward SHIFTs an active snap's remaining
  window and the stranger's cooldowns; the blight rides the existing, already-classified
  `Harvest.Regrow`.
- **Stranger and gru abroad the same night**: legal — they are independent entities with
  independent latches; the stranger's protection rules keep it out of firelight exactly
  like the gru, and neither preempts the other's schedule entry.
- **All-dead dawn / same-batch death at a boundary**: the generalized emitter keeps the
  shipped batch-scan guard — a rubric-violating `agent.died` in the boundary batch
  suppresses the pass (never a photo-finish pass), for every exercise.
- **Blight patch already exhausted** (every forage tile harvested when the entry fires):
  precondition fails, incident skips silently — the schedule proposes, the reducer-valid
  world disposes (US2 AS-2 class).
- **Exercise passed, villager dies later**: pass is the instrument (spec 072); the
  report card renders the recorded pass, the postmortem carries the deaths — two facts,
  both true, both rendered (unchanged).
- **Rolling-boundary exercise never satisfied**: outcome stays `in_progress` until
  `run.ended` — no time-limit fail is introduced; failure is still never an event.

## Requirements *(mandatory)*

### Functional Requirements

Mapped to the three board ACs: AC1 ↔ FR-001..008, AC2 ↔ FR-009..017, AC3 ↔ FR-018..021,
cross-cutting gates FR-022..025.

**AC1 — exercises with production evaluators**

- **FR-001**: `sim.ScenarioExercises` MUST grow to nine hand-authored exercises: stage-1
  `first-night` (existing), `cold-dawn`, `stranger-at-the-gate`; stage-2 `the-law`
  (existing), `blighted-larder`; stage-3 `toolsmith`, `fog-watch`; stage-4 `long-winter`,
  `stewards-charge` — ids, seeds, concepts, framings, rubric terms, schedules, and
  boundaries per data-model.md's inventory table. Every exercise's seed is unique
  (46101–46109) and every schedule compiles at boot.
- **FR-002**: `EvaluateRubric` MUST carry a production arm for every cataloged exercise,
  each a pure function over state facts (no log scan). A sweep test MUST assert no
  cataloged id reaches the default pending arm. The default arm itself stays (honest
  rendering for future non-evaluator content).
- **FR-003**: `scenarioRubricEvents` MUST generalize from its `first-night`-only guard to
  a per-exercise boundary: `ExerciseDefinition` gains a boundary declaration
  (fixed dawn-of-day-N, or rolling every-dawn from day 2 — data-model.md), the emitter
  evaluates exactly at boundary dawns, and the shipped guards (same-batch death scan,
  `hasCurriculumPass` once-only latch, pass-then-unlock same-batch ordering) apply to
  every exercise unchanged.
- **FR-004**: Evidence assembly MUST generalize through the sanctioned constructors only:
  order evidence via `OrderPlacedEvidence` (shipped), charter evidence via a new
  state-sourced constructor reading `CharterObservedSeq/Tick/Custom` (FR-005), and
  skill evidence via a new `SkillsObservedEvidence` (FR-006). No freehand `EvidenceRef`
  construction anywhere.
- **FR-005**: The `metatron.charter_observed` reducer arm MUST additionally persist the
  observation's envelope coordinates — `State.CharterObservedSeq`/`CharterObservedTick`
  (`omitempty`, the `GuardianOrder.PlacedSeq` precedent) — so `the-law`'s pass evidence
  is state-derivable, removing the spec-072 FR-009 blocker ("state does not retain" the
  Seq/Tick).
- **FR-006**: A new event type `metatron.skills_observed` MUST record the bound
  skill-file set the guardian's turn ran under (fingerprint + names), emitted through the
  guardian's existing turn-time observation pipeline exactly as `metatron.charter_observed`
  is (emit on fingerprint change only), reducer-persisted as
  `State.SkillsFingerprint/SkillsObservedSeq/SkillsObservedTick`. Skill files bind only
  from stage-3 and only players author them, so `SkillsObservedEvidence` derives
  `Custom: true` by construction — the stage-3→4 gate's long-deferred evidence design.
- **FR-007**: `world.ValidScenarioExercise` MUST mirror the grown catalog id set
  (`TestScenarioVocabularyMirrorsSimCatalog` stays green), and `promptworld new
  --scenario` / the exercise tab / status lines require NO code change beyond the catalog
  (they enumerate `sim.ScenarioExercises` already).
- **FR-008**: Each new exercise with an authored position (stranger entry tile, blight
  center) MUST pin position validity on its own seed's map in a test (the
  `TestFirstNightSchedulePositionValid` precedent).

**AC2 — incident vocabulary**

- **FR-009**: The incident kind vocabulary MUST grow by three constants — `cold_snap`,
  `forage_blight`, `stranger_arrives` — each with a compile arm (`compileIncident`), an
  emission arm (`scenarioIncidentEvents`), kind-specific schedule parameters
  (data-model.md), and a per-kind window (`windowEnd`) making "fires late, never twice"
  pure via state latches only.
- **FR-010**: `sim.cold_snap{night, until_tick}` — reducer latches
  `State.ColdSnapUntil`; the needs heartbeat's outdoor night warmth loss MUST read a
  harsher rate while `tick < ColdSnapUntil` through the same arithmetic path as ambient
  night cold; no end event (read-time expiry). Precondition: no snap already active.
- **FR-011**: `sim.forage_blighted{x, y, radius, tiles, regrow_tick}` — one event per
  firing (the `sim.food_rotted` merge precedent), tiles resolved in deterministic walk
  order; the reducer marks each via the EXISTING `Harvest{X, Y, Regrow}` overlay with the
  blight's far deadline. Precondition: at least one unharvested forage tile in the patch.
- **FR-012**: `stranger.arrived/moved/took/departed` — a positioned entity
  (`State.Stranger`, gru precedent): greedy movement toward the nearest unattended
  store on per-decision seeded RNG (new purpose tags), bounded takes appended to
  `State.StrangerTakes` (ring-bounded), absolute avoidance of firelight/shelter (the
  gru's protection radii), departure at the next dawn. Takes mutate pile/chest contents
  through the same state shapes agent withdrawal already uses. Precondition: no stranger
  abroad, entry tile passable and unprotected.
- **FR-013**: No payload of any new type may carry an authored/scenario marker; every
  emission-time precondition MUST be a named predicate reusable verbatim by a future
  ambient emitter (TASK-28's recorded seam).
- **FR-014**: Ambient dice paths for the new kinds are OUT of scope (TASK-28,
  reorient move #11 "dual-duty"); consequently no per-kind preemption check (the
  `gruScheduledTonight` twin) ships for them yet — but the seam and its contract MUST be
  documented where `gruScheduledTonight` lives, so TASK-28 adds rolls + preemption as one
  move.
- **FR-015**: Every new tick-anchored state field MUST be classified in the miracle
  rebase SHIFT/KEEP taxonomy (`rebaseTicks`), with tests (`ColdSnapUntil` SHIFT;
  `Stranger`'s move/take cooldown ticks SHIFT; ledger record ticks KEEP as historical
  facts — final classification argued in plan.md).
- **FR-016**: All new events MUST satisfy `TestCatalogSweep`: a `digestRegistry` row +
  `catalogFixture` row per type; `familyByNamespace` gains `"stranger"` mapped onto the
  gru/threat family voice; `stranger.took` joins the whole-line alert set beside
  `social.chest_taken`; `sim.cold_snap`/`sim.forage_blighted` render in the sim family
  tint. `docs/wiki/event-types.md` (the sweep's scanned parent) MUST backtick the new
  types, with full rows in the appropriate child notes.
- **FR-017**: All effects are events through the reducer; genesis replay of a recorded
  run reproduces state byte-identically with no scenario armed; an ambient world's tick
  path is byte-identical to pre-077 (no new code runs when `s.scenario == nil` and no
  new events exist in the log).

**AC3 — lesson content**

- **FR-018**: `lessonCatalog` MUST grow by four entries — `first-explain-answer`,
  `first-report-card`, `first-skill-file`, `same-refusal-pattern` — all prompting tier,
  skin-tokened, with triggers per data-model.md; catalog count 8 → 12, help-overlay
  population stays 1:1 by the existing structural test.
- **FR-019**: The `same-refusal-pattern` detector MUST extend the lesson machinery with a
  minimal bounded fold (per-reason rejection counter, session-local, capped) — the ONE
  stateful trigger seam, designed so per-event triggers stay pure predicates.
- **FR-020**: `first-faith-event` MUST NOT ship or be stubbed: TASK-118 has not run. The
  deferral MUST be recorded as a rider note on TASK-118 (board note — derived state) so
  the tranche's fourth entry lands with faith itself.
- **FR-021**: No lesson fires more than once per player (per-user seen record, shipped
  machinery unchanged); the new entries introduce no new dismissal semantics.

**Cross-cutting gates**

- **FR-022**: `docs/design/tui/` amendments ride this PR: `patterns/chronicle-grammar.md`
  (new digest rows + alert-tier addition), `overlays/help.md` + `panels/lesson-row.md`
  (catalog 12), `panels/exercise.md` (grown exercise catalog + incident vocabulary),
  `panels/map.md` (stranger glyph, gru precedent) — plus every page
  `check-tui-design.mjs --changed` flags, re-verified and re-pinned in this PR.
- **FR-023**: Wiki re-pins ride the branch (pr gate, no bypass): body amendments expected
  on `scenario-machinery.md`, `scenario-machinery-surfacing.md`,
  `curriculum-ladder-progression.md`, `event-types.md` + children (curriculum rows;
  new incident/skills rows), `executor-tick-subsystems.md`, `gru.md`,
  `tui-chronicle-feed.md`, `tui-input-help.md`/`tui-client.md` (lessons),
  `sim-state-world-fields.md`, `world-save-manifest-fields.md` (scenario vocabulary),
  guardian children (skills observation); `docs/player/` regenerated.
- **FR-024**: Merge is `gh pr merge --merge` only; post-merge main commits are derived
  state only (board/bridge/tasks ticks + the TASK-118 rider note).
- **FR-025**: No new tuning dials, no new IPC/status fields, no new dock tabs or push
  surfaces — this wave is content on shipped machinery; any temptation otherwise is a
  scope question for the operator, not a silent addition.

### Key Entities

- **`sim.ExerciseDefinition`** — gains a boundary declaration and kind-parameterized
  schedule entries; stays compiled-in CONTENT (never player data).
- **Incident kinds** — `cold_snap`, `forage_blight`, `stranger_arrives` beside
  `gru_emerges`: constants + compile/emission arms behind the shipped `incidentSource`
  seam.
- **`State.ColdSnapUntil`**, **`State.Stranger` / `State.StrangerTakes`**,
  **`State.CharterObservedSeq/Tick`**, **`State.SkillsFingerprint/SkillsObservedSeq/Tick`**
  — new `omitempty` event-sourced state; reducer is the only writer; zero values honest.
- **`metatron.skills_observed`** — the stage-3 evidence event, `charter_observed`'s twin.
- **`SkillsObservedEvidence` / state-sourced charter evidence constructor** — the third
  and fourth sanctioned `EvidenceRef` constructors.
- **`lessonCatalog` tranche 2** — four entries + the bounded refusal fold.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001** (board AC #1): the catalog enumerates 9 exercises (3/2/2/2 by stage); a
  sweep test proves every cataloged exercise has a production `EvaluateRubric` arm; every
  exercise's pass emission is proven on a deterministic fixture, including `the-law` with
  charter evidence and a stage-3 pass whose `Custom` evidence unlocks stage-4.
- **SC-002** (board AC #2): three new incident kinds fire from authored schedules in
  fixtures; per-kind replay round-trips are byte-identical; a pre-077 ambient world's
  event stream is byte-identical under the new binary (determinism harness).
- **SC-003** (board AC #2): `TestCatalogSweep` and
  `TestExerciseRubricTermsAreCatalogedEventTypes` pass over the grown vocabulary; no new
  alert channels — `stranger.took` is the only addition to the existing whole-line tier.
- **SC-004** (board AC #3): the four lesson entries surface exactly once each in
  projection tests; the refusal fold triggers on 3 same-reason rejections and never on
  mixed reasons; help overlay lists 12.
- **SC-005**: `node scripts/check-tui-design.mjs --changed` passes with the named pages
  amended; `go test ./...` green; existing snapshots load byte-identical.
- **SC-006**: `node scripts/check-merge-drift.mjs pr` exits 0 from the worktree (wiki
  re-pins + player docs in-branch); PR merges as a merge commit.

## Assumptions

- **TASK-149 merged** (verified in code: `theLawRubric`, `State.CharterCustom`, the
  shared fact resolver) — the report-card surfaces need NO change here; new exercises'
  evaluators feed `resolveReportCardFacts` automatically.
- **Ambient emission of the new incident kinds belongs to TASK-28** (reorient move #11:
  "ambient drama supply AND authorable scenario incident vocabulary"). 077 ships shapes,
  reducer arms, preconditions-as-predicates, and scheduled emission; TASK-28 adds the
  dice and the per-kind preemption twins. Recorded here and in the wiki so the seam is a
  decision, not a drift.
- **`first faith event` is post-TASK-118** (not run; lane order decision 9 puts TASK-118
  after TASK-67). Deferred with a rider note, per FR-020.
- **Skill files are inherently player-authored** (no game-shipped skill files exist;
  stage-1/2 lock them out entirely), so `metatron.skills_observed` evidence is
  `Custom: true` by construction — the honest twin of the charter's derived-inverse rule.
- **`State.Deaths` entries carry the death cause** (the spec-044 ledger), so
  cause-scoped rubric terms ("no villager freezes/starves") are state-derivable; exact
  field names pinned at plan time.
- **One exercise per world** stays the v1 posture; the 32-entry pass ring is never pruned
  past a same-exercise pass in practice (shipped reasoning on `hasCurriculumPass`,
  unchanged).
- Storage-stock rubric terms read pile/chest contents from state (the v3 storage
  economy); thresholds are exercise content pinned in data-model.md, tuned only by
  editing content.
