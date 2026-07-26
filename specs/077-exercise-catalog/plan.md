# Implementation Plan: Exercise catalog wave — exercises, incidents, lessons

**Branch**: `077-exercise-catalog` (task branch: `task-151-exercise-catalog`) | **Date**: 2026-07-26 | **Spec**: [spec.md](spec.md)

## Summary

Three content waves on shipped machinery, sim-first. (1) **Incidents**: grow the closed
kind vocabulary by `cold_snap` / `forage_blight` / `stranger_arrives` — constants +
compile arms + emission arms behind the existing `incidentSource` seam, reducer-total
event arms, preconditions as named TASK-28-reusable predicates, digest/catalog rows for
every new type. (2) **Exercises**: nine cataloged definitions (3/2/2/2 by stage), a
production `EvaluateRubric` arm per exercise, `scenarioRubricEvents` generalized to
per-exercise boundaries with sanctioned-constructor evidence assembly (charter
coordinates persisted; `metatron.skills_observed` added as the stage-3 evidence event).
(3) **Lessons**: four tranche-2 entries including the fold-based wrong-thing detector.
Design pages, wiki re-pins, and player docs ride the PR.

## Technical Context

**Language**: Go (sim + Bubble Tea TUI). **Testing**: `go test ./...`; `driveTicks`
determinism harness for per-exercise pass fixtures; reducer round-trip + genesis-replay
byte-identity suites; `TestCatalogSweep` / `TestExerciseRubricTermsAreCatalogedEventTypes`
/ `TestScenarioSchedulesCompile` / `TestScenarioVocabularyMirrorsSimCatalog` as the
existing gates this content must keep green. **Scope**: `internal/sim/{scenario,
curriculum, state, guardian, agents, executor}.go` + new `internal/sim/stranger.go`;
`internal/guardian/` (skills observation emission); `internal/world/world.go`
(vocabulary mirror); `internal/tui/{grammar,digest,lessons,help}.go`;
`docs/design/tui/` (chronicle-grammar, help, lesson-row, exercise, map);
`docs/wiki/` re-pins + `docs/player/` regen. **Constraints**: replay determinism (events
only; per-decision RNG with new purpose tags); snapshot byte-identity (`omitempty`, no
format bump); ambient worlds byte-identical to pre-077; no new channels/dials/status
fields (spec FR-025); merge-commit-only.

## Constitution Check

- **I. Artifact-grounded** — PASS: decision chain reorient decision 5 / merged position 5
  → TASK-151 → this spec; the faith-event deferral produces a board rider (FR-020), not
  silence.
- **II. One task, one PR** — PASS: TASK-151 ↔ `task-151-exercise-catalog` ↔ one PR;
  the three ACs are phases, not PRs.
- **III. Gates** — PASS: catalog sweeps, schedule-compile pins, design gate, merge-drift
  claim/worktree/pr gates at their choke points.
- **IV. Grounding freshness** — PASS (planned): touched sources are pinned by
  `scenario-machinery(.md/-surfacing)`, `curriculum-ladder-progression`,
  `event-types` + children, `executor-tick-subsystems`, `gru`, `tui-chronicle-feed`,
  `tui-input-help`/`tui-client`, `sim-state-world-fields`, `world-save-manifest-fields`,
  guardian children; reconciliation computed from the actual branch diff, re-pinned
  in-branch; player docs regenerated; merge commit only.
- **V. Model tiers** — PASS: planning tier authored this cycle; implementation
  dispatches to `spec-implementer` on **Opus 4.8** (recorded on TASK-151: new
  reducer-valid event kinds across sim reducer + digest + evaluators; replay-safety
  doctrine-adjacent).

**Post-Phase-1 re-check**: PASS — no new violations; Complexity Tracking empty.

## Design

### D1 — Incident kinds (sim core, the extension points the file names)

`internal/sim/scenario.go`: three constants beside `IncidentGruEmerges`; `compileIncident`
arms with per-kind param validation (data-model §1); `scenarioIncidentEvents` arms
emitting the payloads of data-model §2 after named-predicate preconditions
(`coldSnapActive`, `blightableTiles`, `strangerEntryValid` — exported-shape TBD by
implementer, but each a standalone predicate TASK-28 can call). Windows: cold snap ends
at its own `until_tick`; blight/stranger lapse at next dawn (`nextDawnTick`, shipped).
Document the TASK-28 seam (ambient rolls + per-kind preemption twins) beside
`gruScheduledTonight` (FR-014).

### D2 — Reducer arms + entity step

- `sim.cold_snap` → `State.ColdSnapUntil`; needs heartbeat reads a cold-snap rate via a
  nil-safe accessor-style helper reusing the `warmthLossCold` arithmetic path (rate
  constant beside it, e.g. doubled loss; exact constant is content, pinned in code).
- `sim.forage_blighted` → append `Harvest` overlays (skip already-harvested — idempotent
  re-apply; validates coordinates at the door).
- `stranger.*` → new `internal/sim/stranger.go`: `Stranger` state struct, `applyStranger`
  dispatch from `state.go` (the `applyGru` pattern), and `strangerStep` called from
  `stepEvents` adjacent to `gruStep` (order: after `gruStep`, before the social beat —
  pinned by a determinism test). Movement/take cadence constants mirror the gru's; RNG
  purposes `"stranger-prowl"`/`"stranger-take"`; protection predicates shared with the
  gru (extract, don't duplicate). Witness/victim-adjacent memories via
  `situatedMemoryEvent` idioms (rumor fuel).
- `metatron.skills_observed` → `applyGuardian` arm persisting fingerprint + envelope
  Seq/Tick; `charter_observed` arm additionally stamps `CharterObservedSeq/Tick`.

### D3 — Emitter generalization + evidence

`ExerciseDefinition.BoundaryDay` (data-model §4). `scenarioRubricEvents`: boundary check
becomes `boundaryDue(def, nextTick)` (fixed-day dawn or rolling every-dawn from day 2);
keep batch-death scan, `hasCurriculumPass` latch, pass→unlock ordering. Evidence
assembly walks the satisfied terms and applies the sanctioned constructor per term type
(data-model §4); `CharterEvidenceFromState` + `SkillsObservedEvidence` land in
`curriculum.go` beside their siblings with the same doc-comment honesty contracts.

### D4 — Nine evaluator arms

`EvaluateRubric` switch grows seven cases (plus the two shipped); each a small pure
function in `scenario.go` following `firstNightRubric`/`theLawRubric` style. Shared
helpers: `deathsByCause(s)`, `storedFoodTotal(s)`, `playerOrderSince(s, tick)`,
`firstNightWatch` (reused). A catalog sweep test asserts every cataloged id returns
non-default arms (spec SC-001). Death-cause field names verified against the spec-044
ledger at implementation time (Assumption pinned in spec.md).

### D5 — Rebase taxonomy

`rebaseTicks` classifications (data-model §3): `ColdSnapUntil` SHIFT;
`Stranger.LastMove/LastTake` SHIFT; `StrangerTakes[i].Tick`, `CharterObservedTick`,
`SkillsObservedTick` KEEP (historical/log coordinates — seq-paired ticks must keep
pointing at their recorded events). Tests beside the existing taxonomy suite; the
compile-time completeness idiom the taxonomy note documents applies.

### D6 — Guardian skills observation

`internal/guardian/`: mirror the charter observation pipeline — fingerprint the BOUND
skill set at turn/status assembly (post-`stageSkills`, so stages 1–2 never observe),
emit `metatron.skills_observed` on fingerprint change through the same door
`metatron.charter_observed` uses. Empty bound set = no event (absence is not an
observation). Names list = bound file names in deterministic order.

### D7 — Digest + lessons (TUI)

`grammar.go`: `familyByNamespace["stranger"]` → gru/threat family; alert set +
`stranger.took`. `digest.go`: registry rows per data-model §2 (payload → readable line;
`sim.forage_blighted` uses the first-fact-plus-count shape for its tile list);
`catalogFixture` rows. `lessons.go`: four entries + the `lessonFold` seam (data-model
§6); `populateHelpLessons` untouched (1:1 by construction). No dock/tab/status changes.

### D8 — Design pages + grounding

Amend: `patterns/chronicle-grammar.md` (rows + alert tier), `overlays/help.md` (12
entries), `panels/lesson-row.md` (tranche 2), `panels/exercise.md` (nine-exercise
catalog + incident vocabulary table), `panels/map.md` (stranger glyph — the gru's red-G
precedent; if the tile registry's entity conventions require a registry row, follow
spec 068's single-source rule). Then `check-tui-design.mjs --changed` re-verify +
re-pin. Wiki: body amendments + re-pins per Constitution Check IV; consider a new
`event-types-scenario-incidents.md` child for the incident rows if the existing children
don't fit (author's judgment at wiki-update time; parent must backtick all new types
either way). Player docs regen. Board rider on TASK-118 for `first-faith-event`
(post-merge derived state).

## Phase ordering (tasks.md mirrors this)

Setup → Foundational sim state/events (D2 arms, D5 taxonomy, D6 skills event — blocks
everything) → US2 incidents (D1 + stranger step; digest rows) → US1 exercises (D3
emitter + D4 evaluators + catalog + world mirror + per-exercise fixtures) → US3 lessons
(D7 lessons half) → design pages (D8) → grounding → close-out. US2 precedes US1 because
five exercise schedules reference the new kinds.
