---
name: stage-defaults
description: The spec-066 stage-shaped TUI layout defaults — a single authority table governing which chrome surfaces default on/off per curriculum stage (never a capability lock), the resolution engine (startingVisibleSet, session overrides), and the live re-resolution/first-occurrence-arrival plumbing that fires on a genuine stage change
kind: pattern
sources:
  - internal/tui/stagedefaults.go
  - internal/tui/layout.go
  - internal/tui/tui.go
verified_against: 7e3c2b5f5f23eb8e5fcb37d0f867dbc6f46a289b
---

# Stage-shaped TUI layout defaults

Spec 066 (TASK-128, reorientation decision 3) is the promised absorption of
every per-surface stage default (`lessonRowDefault`, the exercise tab's
incident visibility, and others authored ad hoc across earlier specs) into
ONE table. The ruling it enforces: stage shapes what's **visible by
default**, never what's **reachable** — every governed surface stays fully
reachable at every stage (the help overlay, a solo view, or a folded row's
own pull path); the curriculum ladder's capability lock ([[curriculum-ladder]],
[[guardian]]) is untouched and orthogonal — this table governs the layout
only.

## How it works

**The authority table** (`docs/design/tui/patterns/stage-defaults.md`'s "Per-
surface stage defaults" table, mirrored cell-for-cell by
`stageDefaultsTable` in `internal/tui/stagedefaults.go`): ten governed
surface rows (`surfaceLessonRow`, `surfaceGuardianStrip`,
`surfaceVillagerStrip`, `surfaceExerciseTab`, `surfaceIncidentVocabulary`,
`surfaceSystemsTab`, `surfaceGuardianConsole`,
`surfaceHelpGuardianSection`, `surfaceCeremony`, `surfacePostmortem`) ×
six columns (stage-1..stage-4, pre-ladder, narrow). A parity sweep test
(`TestStageDefaultsSweep`, the `digest.go` `TestCatalogSweep` precedent)
parses the design doc's own markdown table at test time and asserts byte
equality against `stageDefaultsTable` (after stripping the page's own
emphasis/backtick/ditto formatting) — a default changes on the page first,
the Go table second, or the build breaks. Most rows are carried for
completeness/parity rather than gating anything today: the guardian strip,
systems tab, and guardian console are unconditional at every stage; only the
lesson row, the villager strip (its own surface didn't exist until
[[village-lens]] shipped it), and the incident-vocabulary/help-guardian-
section content selectors actually vary today.

**Resolution** (`resolveStageDefaults(stage, hasScenario)`): `stageColumnFor`
selects the table column — stage-1..stage-4 map directly, and `""` OR any
UNRECOGNIZED value takes the Pre-ladder column (the fail-open rule: an
unrecognized stage never takes a narrower posture than pre-ladder). The
result is a `startingVisibleSet{LessonRowOn, GuardianStripOn,
VillagerStripOn, SystemsTabOn, ExerciseTabOn, IncidentVocabulary,
GuardianConsoleReachable, HelpGuardianVariant}` — the exercise tab's
`ExerciseTabOn` field mirrors the caller's `hasScenario` input independent
of the resolved column, since its presence is WORLD-shaped, not stage-shaped
(FR-006; the real gate stays `exerciseID() != ""`, [[scenario-machinery]]).
`Model.stageResolve()` is the one consolidated read (`resolveStageDefaults`
plus any session `stageOverrides` layered on top via `applyOverrides`) for
frame-level assertions and any future call site that wants one fact rather
than composing `currentStage()`/`exerciseID()` by hand — no production tab
gate was rerouted through it except `lessonRowDefault`, which is now a pure
delegation (`resolveStageDefaults(stage, false).LessonRowOn`) with unchanged
behavior.

**Session overrides** (`surfaceOverrides map[string]bool`,
`Model.stageOverrides`): an in-session, per-surface visibility choice a
player made explicitly would outrank re-resolution until the session ends
and is never persisted — `applyOverrides` is the mechanism, proven against
synthetic input (`TestApplyOverridesPrecedence`), but NO production key sets
an entry today (no in-session command exists yet to toggle a governed
surface); the field exists so a future toggle command has a session-scoped
place to land without a second mechanism.

**Live re-resolution + first-occurrence arrivals** (`tui.go`'s `Update`,
`statusMsg` case, US3): on a GENUINE live stage change — `hadStatus` is true
(never the first status poll, which is boot resolution rendered continuously
via `currentStage()` reads, not an "arrival") and `m.currentStage()` differs
from before — both the previous and next `startingVisibleSet`s are resolved
and diffed via `newlyOnSurfaces` (fixed order: lesson-row, guardian-strip,
villager-strip, systems-tab; the exercise tab is deliberately excluded, since
its presence is world-shaped, never a stage-driven arrival). Each newly-on
surface routes through the EXISTING first-occurrence lesson machinery
(`announceSurfaceArrival` → `lessonTriggers`, spec 055's `lessonCatalog` —
never a second dedupe mechanism) via `surfaceArrivalLessonID`, a map that is
EMPTY today: under the CURRENT authority table no numbered-stage transition
ever widens a governed row (every row is either constant-on or narrows going
up the ladder), so this is forward-compatible plumbing exercised only by
synthetic fixtures (`TestNewlyOnSurfacesDiffExactlyOnce` and siblings,
`stagedefaults_test.go`) rather than any live transition — kept as real,
tested code rather than dead weight, for whichever future table revision
does widen a surface. Takeovers (ceremony/postmortem, [[takeover-surfaces]])
fire off their own event triggers (`run.ended`/`curriculum.stage_unlocked`),
independent of this re-resolution path (FR-008).

## Connections

[[tui-client]] hosts `Model`, the `Update` dispatch this note's
re-resolution step lives inside, and the existing `lessonRowDefault`/
`wantsLessonRow` call sites this feature absorbed without changing their
behavior. [[village-lens]] (spec 060) is the villager strip's own surface —
the first governed row whose "on"-at-every-stage default this table already
anticipated before the row existed. [[scenario-machinery]] owns the
exercise tab's world-shaped presence gate and `IncidentVisibilityFor`'s
per-definition override, which this table's `IncidentVocabulary`
stage-only default composes beneath (`incidentVocabularyColumn`).
[[grounded-feedback]] owns the D9 help-guardian-section CONTENT this table's
`HelpGuardianVariant` field only names the selector for (the content itself
is a separate, not-yet-authored deliverable). [[takeover-surfaces]] owns the
ceremony/postmortem takeover surfaces this table's rows describe but does
not gate (they fire on their own events, independent of stage
re-resolution). [[curriculum-ladder]] and [[guardian]] own the ORTHOGONAL
capability lock/stage ceiling this table is deliberately NOT — ruling 3
above is the explicit boundary between the two.

## Operational notes

`TestStageDefaultsSweep` is the structural guard: editing a default here
without first amending `docs/design/tui/patterns/stage-defaults.md` fails
the build. Under the table as shipped, `computeRows`/`lessonRowDefault`
behavior is provably unchanged from pre-066 — this feature is an
absorption/consolidation, not a new default anywhere the sweep and
`TestApplyOverridesPrecedence`/`TestNewlyOnSurfacesDiffExactlyOnce` don't
already cover.
