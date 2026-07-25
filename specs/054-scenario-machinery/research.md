# Research: Scenario machinery (spec 054)

## R1 — Where the machinery attaches in the executor

**Decision**: `stepEvents` (internal/sim/executor.go:21, the pure
(state, map, nextTick) → events function called from loop.go:545) gains one
scenario consultation alongside its existing cadence emissions (charge regen
:53-57, order expiry :59-70): due incidents from the armed incident source,
then the rubric evaluation. The armed scenario config is boot-frozen state
on the Loop (the seed's threading precedent), never mutated.

**Rationale**: curriculum.go:14-18 already declares `curriculum.*` "the
EXECUTOR emission class (the metatron.order_expired precedent)… The
production emitter is TASK-119's scenario rubric machinery"; the whitelist
contrast at loop.go:247-250 confirms executor emissions need no door.

## R2 — The incident source seam (FR-002's director seam)

**Decision**: a one-method interface in `internal/sim/scenario.go`:
`incidentsDue(s *State, nextTick int64) []incident` — v1's only
implementation compiles the exercise definition's authored schedule
(game-day + HH:MM → absolute tick via the existing clock arithmetic, at
arm time) and returns entries whose tick ≤ nextTick and not yet emitted
(latched by observing state, not by internal mutation — e.g. gru_emerges is
due while `tick ≥ scheduledTick ∧ s.Gru == nil ∧ !firedThisNight(state)`;
the recorded event IS the latch, keeping the source pure and replay-safe).
The post-v1 live director is a documented second implementation site.

**Rationale**: purity doctrine (executor.go:11-14) — the source may not
carry mutable "already fired" flags or restarts/replays desync; deriving
the latch from state mirrors how charge regen derives from
`MetatronCharges < cap`.

## R3 — gru_emerges preemption (no double spawn)

**Decision**: on a night with a scheduled emergence, the schedule wins and
the random roll is skipped: `gruStep`'s emergence branch (gru.go:96, the
`rngAt("gru-emerge")` gate at :235) gains a scenario-awareness check —
scheduled-tonight → no roll; the scheduled incident emits `gru.emerged` at
its authored tick/position through the same event shape `applyGru` (:286)
already validates. Preconditions (no gru abroad) are checked at emission;
a failed precondition skips silently (US2 AS-2).

**Rationale**: one mechanism per night; the reducer stays the authority;
the RNG streams are untouched (no new purpose tags, no consumed draws —
skipping the roll on scheduled nights is itself deterministic since the
schedule is config).

## R4 — Rubric evaluation and the pass boundary

**Decision**: evaluated in the same stepEvents pass, pure over state:
first-night's terms map to state facts the reducer already maintains —
day boundary (`clock.SecondOfDay(nextTick)` dawn check + day arithmetic),
zero deaths (`len(state death ledger) == 0` for the window), watch placed
(`state.MetatronOrders` history / CurriculumPasses evidence via recorded
`metatron.order_placed` — the evaluator reads the facts, the emitted
payload carries `EvidenceRef`s built by the sanctioned constructors
(curriculum.go:242-251)). On satisfaction at the boundary tick: emit
`curriculum.exercise_passed` and, when `EvaluateUnlock` (curriculum.go:201)
grants, `curriculum.stage_unlocked` — same batch, pass first (the daemon
observer's same-batch contract, internal/daemon/curriculum.go:44-57).
Emission is once-only via the state latch the reducer maintains
(CurriculumPasses dedupe + StagesUnlocked latch, curriculum.go:117-144).
Failure is never emitted — `run.ended` (executor.go:344-366) IS the fail
signal (Status.Ended's comment names this consumer, protocol.go:203-212).

**Rationale**: every piece is a shipped seam whose doc comment awaits this
feature; inventing a parallel evidence path would violate the
"ONLY sanctioned way" constructor contract.

## R5 — Boot wiring + manifest validation + CLI

**Decision**: `world.Open` validates `Scenario.Exercise` against the
compiled catalog ids (the `ValidStage`/`ValidCharterPreset` idiom,
world.go:110-133) — additive, old manifests without the block unchanged.
The daemon arms the loop's scenario at boot from the manifest (SetStage
discipline). `promptworld new --scenario <id>` (commands.go's new FlagSet,
:102-267): resolves the definition, stamps Scenario + the exercise's
`Stage`/`Seed`/preset (stage-1 → tutor preset default per :150-152),
earned-stage gate unchanged (:115-152); unknown id → refusal listing
`sim.ScenarioExercises` ids (the :142-143 refusal voice).

## R6 — Status + exercise tab

**Decision**: Status gains `scenario_exercise` (id) and
`scenario_outcome` (`in_progress`/`passed`/`failed`) — additive omitempty,
composed like `Status.Stage` (server.go:191; protocol.go:175-181 pattern);
outcome derives from replica facts (CurriculumPasses vs Ended). The TUI
exercise tab (key `6`, present only when the status carries an exercise id)
reads: framing/rubric text from `sim.ScenarioExercises[id]` (compiled
content, importable — the tui already imports sim), progress from the
replica (event counts + state facts), visibility mode from the definition
override else stage default. Briefing = a per-attach bool on the Model;
any-key dismiss swallows one key only while the exercise tab is visible.
Expect a dock-enum rebase over TASK-125's systems tab (key `5`).

## R7 — Narrator + morgue hooks

**Decision**: `chronicleNote`'s switch (internal/mind/narrate.go:62-290)
gains cases: `curriculum.exercise_passed` → line + `closeChapter` at the
boundary ("the exercise's outcome narrated in the score voice"); a
scenario world's `run.ended` closes its chapter likewise (ambient run.ended
narration is already the epilogue path — the new chapter trigger fires only
when a scenario is armed, keeping ambient cadence byte-identical). Morgue:
`writeRunSummary` (morgue.go:367-404) gains the exercise-outcome line when
the world's manifest carries a scenario and no pass is on the log —
no-blame register per curriculum.go:285 ("failure is a story, not a
scold").
