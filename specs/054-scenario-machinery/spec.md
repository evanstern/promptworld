# Feature Specification: Scenario incident-schedule machinery (director-lite scheduled emissions)

**Feature Branch**: `054-scenario-machinery`

**Created**: 2026-07-25

**Status**: Draft

**Input**: User description: "Scenario incident-schedule machinery (board TASK-119; learning-game synthesis Wave 2 / reorient D11+D4, Wave 4). The one honest new design step: nothing today injects at a future tick on schedule. V1 rides an executor-style scenario block in world config (deterministic, replay-safe by the charge-regen argument), NOT the InjectSocial door. Powers scenario worlds: promptworld new --scenario first-night = seeded world + authored incident schedule + event-derived rubric + morgue epitaph on failure. Plus the exercise panel (D11): framing, attach briefing, live rubric gauges, per-exercise visibility vocabulary (D4), pass/fail, scenario-cadence narration. The live state-watching storyteller is the post-v1 graduation; leave a documented seam."

## Standing resolutions this spec applies

- **Live rubric gauges**: the parked synthesis question (headline-live vs
  full-breakdown-at-end) was resolved by the authored
  `docs/design/tui/panels/exercise.md` — live per-term gauges with
  met/pending markers and backing counts, at every stage; the D4
  forecast/fog vocabulary governs only incident-schedule visibility, never
  the gauges' shape. This spec implements the page as authored.
- **Substrate already shipped (spec 046)**: `sim.ExerciseDefinition` +
  `FirstNightExercise`/`TheLawExercise` content, the `curriculum.*` reducer
  arms, the per-user unlock observer (same-batch evidence contract), the
  unlock gate (`EvaluateUnlock`), and the reserved `Manifest.Scenario` block
  all exist with doc comments naming TASK-119 as their consumer/producer.
  This feature is the production machinery those seams await.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The rubric machinery: exercises pass and unlock deterministically (Priority: P1)

A player on a scenario world plays the exercise; when the recorded events
satisfy the exercise's rubric (e.g. first-night: dawn of day 2 reached, zero
deaths, a watch was placed), the world itself — not a model, not the client —
emits `curriculum.exercise_passed` and, when the unlock gate's evidence
conjuncts hold, `curriculum.stage_unlocked` in the same batch. Replays
reproduce the pass exactly; the per-user unlocks record updates through the
existing observer.

**Why this priority**: the production emitter is the missing half of spec
046's curriculum — without it no exercise can ever pass outside test
fixtures. Everything else (panel, ceremony linkage, narration) reads these
events.

**Independent Test**: run a seeded scenario world to its pass condition
(compressed clock); assert the two events land in one batch, state latches,
unlocks.json updates, and a replay of the log reproduces the same events at
the same seqs.

**Acceptance Scenarios**:

1. **Given** a first-night scenario world where no villager has died,
   **When** dawn of day 2 arrives and the rubric's remaining terms are
   satisfied, **Then** the executor emission class lands
   `curriculum.exercise_passed{exercise, stage, evidence}` and (per the
   unlock gate) `curriculum.stage_unlocked` in the same batch, exactly once.
2. **Given** the same world replayed from its event log, **When** the fold
   reaches the same tick, **Then** state matches — the pass is recorded
   history, not a live-only judgment.
3. **Given** a scenario world where a rubric-violating event lands (a death
   before dawn), **When** the pass boundary arrives, **Then** no pass is
   emitted — and the exercise reports failed only when `run.ended` lands
   (the exercise outcome is pass, in-progress, or failed-by-run-end; a
   survivable stumble keeps it in progress if the rubric allows).
4. **Given** an ambient (non-scenario) world, **When** any of this
   machinery is consulted, **Then** nothing changes — byte-identical
   behavior (no scenario block → no evaluation, no emissions).

---

### User Story 2 - Authored incidents land on schedule, deterministically (Priority: P1)

A scenario's definition carries an authored incident schedule ("the gru
emerges at 22:00 near the north woods"). At the scheduled game time the
incident's event lands as an executor emission — a pure function of (state,
scenario config, tick) — indistinguishable in kind from charge regen or
order expiry: no LLM, no injection door, replay-safe.

**Why this priority**: co-P1 — the "director-lite" half of the task name;
the exercise's pressure must be authored, not left to the RNG's whims, for
a scenario to teach reliably.

**Independent Test**: run the seeded scenario twice from genesis; the
incident events land at identical ticks with identical payloads both times;
a world whose schedule is exhausted emits nothing further.

**Acceptance Scenarios**:

1. **Given** first-night's schedule entry (gru emerges at 22:00), **When**
   that tick arrives and the incident's precondition holds (no gru already
   abroad), **Then** `gru.emerged` lands at the authored position via the
   executor path, replacing that night's emergence roll (the schedule
   preempts the dice; the dice never double-spawn on a scheduled night).
2. **Given** an incident whose precondition fails at its tick (a gru is
   already abroad), **When** the tick arrives, **Then** the incident is
   skipped — recorded nowhere, invented nowhere (the schedule proposes; the
   reducer-valid world disposes).
3. **Given** two runs of the same scenario world (same seed, same config),
   **When** their logs are compared, **Then** scheduled emissions are
   byte-identical.
4. **Given** the post-v1 live director idea, **When** a developer reads the
   incident machinery, **Then** one documented seam names where a
   state-watching director would attach (the incident-source interface),
   without any live-director code existing.

---

### User Story 3 - `promptworld new --scenario first-night` (Priority: P2)

A player creates a scenario world in one command: seeded with the exercise's
authored seed, stamped with the scenario block and the exercise's stage
(with its charter preset), ready to attach — the schedule armed, the rubric
live, the briefing waiting.

**Why this priority**: the packaging that makes US1/US2 reachable by a
player rather than a test harness.

**Independent Test**: `promptworld new <dir> --scenario first-night`,
inspect the manifest (scenario block, stage, seed), attach and confirm the
exercise is live; an unknown scenario name refuses with the catalog listed.

**Acceptance Scenarios**:

1. **Given** the command with a valid scenario id, **When** the world is
   created, **Then** its manifest carries the scenario block and the
   exercise's stage/seed/preset, and the daemon arms the machinery at boot.
2. **Given** an unknown scenario id, **When** the command runs, **Then** it
   refuses with the known-scenario catalog (the stage-gate refusal voice).
3. **Given** a scenario id whose stage the user hasn't earned (per the
   existing earned-stage gate), **When** the command runs, **Then** the
   existing stage-gate rules apply unchanged (scenario implies its stage; it
   never bypasses the earn gate).

---

### User Story 4 - The exercise panel (Priority: P2)

On a scenario world the dock gains an **exercise** tab: an attach-time
briefing (framing + this stage's incident-visibility mode, dismissed by any
key, once per attach), then live rubric gauges (one row per term:
plain-language label, met/pending marker, backing event count), the incident
line (forecast: upcoming incidents with approximate times; fog: omitted),
and the pass/fail banner. Present only on scenario worlds; reachable in
narrow like any tab.

**Why this priority**: D11 — the scenario's living screen; P2 because the
machinery (US1/US2) must exist first and is independently testable.

**Independent Test**: attach a scenario world; verify briefing-once,
gauges tracking live events, forecast at stage 1 vs fog on a stage-3 world,
pass banner after US1's pass, failed banner after run end.

**Acceptance Scenarios**:

1. **Given** a fresh attach to a scenario world, **When** the exercise tab
   first renders, **Then** the briefing shows (framing + visibility mode);
   any key dismisses it for this attach; re-attaching shows it again.
2. **Given** the dismissed briefing, **When** rubric-relevant events land,
   **Then** the gauges update live (met/pending + counts) — same replica
   data, no extra IPC.
3. **Given** a stage-1–2 or pre-ladder scenario world, **When** the panel
   renders, **Then** the incident line forecasts the schedule; **Given**
   stage 3+, **Then** the line is omitted (fog) — and a per-exercise
   visibility override in the definition wins over the stage default (D4:
   a vocabulary, not a boolean).
4. **Given** `curriculum.exercise_passed` lands, **Then** the pass banner
   renders (and the ceremony overlay, when TASK-127 ships it, reads the
   same two events — this panel is the ambient view, no coupling).
5. **Given** `run.ended` with no pass, **Then** the banner reads failed —
   consistent with the postmortem, distinct surface.
6. **Given** an ambient world, **Then** no exercise tab exists at all.

---

### User Story 5 - The story tells the score (Priority: P3)

A short scenario run still produces chronicle narration: the narrator gains
one additional chapter trigger at the exercise's pass/fail boundary, so a
40-minute first-night run yields at least one narrated chapter carrying the
exercise's outcome in the score-narrative voice; and a failed scenario's
morgue carries the exercise outcome in the run summary — failure is a
story, not a scold.

**Why this priority**: D11's cadence fix + the morgue epitaph AC; polish on
top of the machinery.

**Independent Test**: drive a scenario to pass and to fail; assert a
narrated chapter lands at the boundary in both cases and the failed run's
morgue names the exercise outcome.

**Acceptance Scenarios**:

1. **Given** the exercise passes mid-day, **When** the pass lands, **Then**
   the narrator closes a chapter at that boundary (additive to the
   day/night cadence, which is unchanged for ambient worlds).
2. **Given** the run ends before any pass, **When** the postmortem
   machinery runs, **Then** the morgue's run summary names the exercise and
   its failure in the no-blame evidence register.

---

### Edge Cases

- **Pass at the same tick as run end** (dawn arrives as the last villager
  dies): the batch is ordered — deaths land before `run.ended` (existing
  executor ordering) and the pass evaluation sees the post-death state; an
  all-dead dawn is a fail, not a photo-finish pass.
- **World restarted mid-scenario**: schedule and rubric state derive from
  (manifest, recorded events, tick) — nothing in-memory-only; a restart
  resumes exactly (the charge-regen argument).
- **Snapshot upgrade of a pre-054 scenario-stamped world**: the reserved
  block was never consumed before, so arming it on old worlds is fine —
  but evaluation only looks forward (no retroactive pass for events before
  the machinery existed... moot in practice: no pre-054 world has a
  scenario block since nothing ever wrote one).
- **Exercise already passed, world keeps running**: the pass latch is
  once-only (existing reducer dedupe); the panel shows passed; incidents
  keep landing per schedule (the world goes on).
- **Multiple exercises**: v1 = one exercise per scenario world (the
  manifest names one); the definition catalog supports more scenarios, not
  concurrent exercises.
- **Clock-time skew**: schedule times are game-time (day + HH:MM → tick via
  the existing clock arithmetic); time-snap miracles move the clock — the
  schedule keys on absolute ticks computed at boot from the authored times,
  so a snap past an incident's tick lands it on the next boundary check
  honestly (fires late, never twice; document in the contract).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: A deterministic scheduled-emission primitive MUST exist in the
  executor emission class: pure function of (state, boot-frozen scenario
  config, next tick), no LLM, no injection door, no new RNG stream; emitted
  events are existing reducer-valid types.
- **FR-002**: The incident vocabulary v1 MUST ship at least `gru_emerges`
  (authored time + position), preempting that night's random emergence roll
  (no double spawn); the vocabulary MUST be a closed enum with a documented
  seam (an incident-source interface) where the post-v1 live director
  attaches.
- **FR-003**: The rubric evaluator MUST be the production emitter for
  `curriculum.exercise_passed` (+ same-batch `curriculum.stage_unlocked` via
  the existing unlock gate and evidence constructors), evaluated
  deterministically from state each tick, emitting exactly once, honoring
  the existing reducer dedupe/validation and the same-batch observer
  contract.
- **FR-004**: The FirstNightExercise rubric MUST be evaluatable end-to-end:
  dawn-of-day-2 boundary, zero-deaths condition, watch-placed evidence,
  charter-observed evidence for its unlock — using the existing evidence
  constructors (`CharterObservedEvidence` et al.).
- **FR-005**: `promptworld new --scenario <id>` MUST create a world stamped
  with the scenario block, the exercise's stage (+ preset) and seed;
  unknown ids refuse with the catalog; the earned-stage gate applies
  unchanged.
- **FR-006**: The daemon MUST arm the machinery from `Manifest.Scenario` at
  boot (boot-frozen, the SetStage discipline); a world with no scenario
  block is byte-identical to today.
- **FR-007**: Status MUST carry additive omitempty scenario facts (exercise
  id; pass/fail/in-progress state) so a linear client reads the outcome
  model-free (D1; the `Status.Ended` comment already names this consumer).
- **FR-008**: A new dock tab **exercise** MUST render on scenario worlds
  only: attach-once briefing (framing + visibility mode, any-key dismiss),
  live rubric gauges (term label, met/pending, backing count), incident
  line per the visibility vocabulary, pass/fail banner — per
  `panels/exercise.md` as authored; absent entirely on ambient worlds.
- **FR-009**: Incident visibility MUST be a per-exercise vocabulary value
  with the stage-keyed default (forecast at stages 1–2/pre-ladder, fog from
  stage 3; definition override wins) — never a boolean (D4).
- **FR-010**: The narrator MUST gain one additional chapter trigger at the
  exercise pass/fail boundary (additive; ambient cadence unchanged), and
  the morgue's run summary MUST name the exercise outcome on a failed
  scenario run (no-blame register).
- **FR-011**: The design reference MUST be amended in the same PR:
  `panels/exercise.md` specified → shipped (real symbols, tab key recorded),
  `panels/dock.md` (5-tab row on scenario worlds), `patterns/keymap.md`
  (tab key + briefing dismiss + parity notes), `patterns/stage-defaults.md`
  re-verified (visibility vocabulary now real), re-pins on all touched
  pages.
- **FR-012**: All new machinery MUST be covered by determinism tests
  (twice-run identical logs; replay equivalence) and the full `-race`
  suite.

### Key Entities

- **Incident schedule entry**: (incident kind, game-time, kind-specific
  params e.g. position) — authored data on the exercise definition;
  compiled to absolute ticks at boot.
- **Incident source (the seam)**: the interface producing due incidents for
  a tick — v1 has exactly one implementation (the authored schedule); the
  live director is a documented future second.
- **Rubric evaluation state**: derived per tick from (state, definition) —
  term satisfaction + the pass boundary; nothing persisted beyond the
  recorded events.
- **Scenario block** (`Manifest.Scenario`): the reserved manifest field,
  now consumed — names the exercise id; boot-frozen.
- **Exercise tab state**: briefing-dismissed (per attach), the panel's
  gauge/banner projections over the replica.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Two genesis runs of the same scenario world produce
  byte-identical scheduled-emission and pass/unlock event sequences; replay
  of a recorded run reproduces state exactly.
- **SC-002**: A player can go from nothing to a live first-night scenario
  in one command and see the briefing, gauges, and forecast without
  touching config files.
- **SC-003**: The first-night exercise is completable end-to-end by play:
  pass emits once, unlocks stage 2 in the per-user record, the ceremony
  trigger pair is on the log, and the panel/CLI both report it.
- **SC-004**: A failed scenario run's morgue names the exercise outcome;
  a sub-one-game-day run still yields ≥1 narrated chapter.
- **SC-005**: Ambient worlds are provably untouched: the no-scenario path's
  behavior is byte-identical (regression suite green with zero fixture
  changes for ambient tests).
- **SC-006**: The design gate passes with `panels/exercise.md` shipped and
  every touched page re-pinned in the PR.

## Assumptions

- The exercise tab's key is `6` (dock digits continue; TASK-125 takes `5`
  for systems — expect a rebase over that merge; if 125 hasn't merged
  first, still reserve `6` and leave `5` untouched).
- One exercise per scenario world (v1); the definition catalog
  (`sim.ScenarioExercises`) is compiled-in content.
- Schedule times are authored in game time and compiled to absolute ticks
  at boot; a time-snap past an incident fires it on the next check (late,
  never twice, never invented retroactively on replay — the emission is
  what's recorded).
- The gauges read the replica only (event counts + state facts); framing
  and rubric text come from the compiled definition looked up by the
  status-carried exercise id.
- The briefing's "any key" dismiss swallows exactly one keypress on the
  exercise tab only (never a global key-eater).
- Model tier: Opus 4.8 (sim-loop/determinism + executor emission class =
  senior tier per the runbook Lane 2 assignment).
- Dependencies: spec 046 substrate (shipped); TASK-125's systems tab merges
  around the same window (dock-enum rebase expected); TASK-127's ceremony
  reads this feature's events later (no coupling now).
