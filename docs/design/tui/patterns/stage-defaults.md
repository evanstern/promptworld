---
title: Pattern — stage-shaped layout defaults
class: pattern
status: shipped
verified_against: 1b3fe329e64431b66f9995a1f8c8e5fc979dafb7
sources:
  - internal/tui/stagedefaults.go
  - internal/tui/layout.go
  - internal/tui/tui.go
  - internal/tui/lessons.go
  - internal/sim/scenario.go
---

# Pattern: stage-defaults

The authority table for reorientation decision 3: **stage may shape TUI
layout defaults — defaults only.** Every other page in this corpus that
states a stage-default visibility value references this table rather than
restating its own; if a value ever needs to change, it changes here once.

## The ruling, restated precisely

1. Stage shapes what's **visible by default**, never what's **reachable**.
   Every surface below remains reachable at every stage: through the help
   overlay (`overlays/help.md`), through a solo view
   (`pages/solo-views.md`), or (for a folded chrome row) through its own
   pull path back to full content.
2. **Pre-ladder worlds get everything** — the default set for a pre-ladder
   world (`Stage == ""`) is the union of every stage's default-on set, the
   same "ungated, stage-4 semantics" posture `internal/guardian`'s stage
   ceiling already carries ([[curriculum-ladder]]).
3. **Capability locks stay guardian-only** (spec 046 doctrine, untouched by
   this feature): nothing in this table is a capability lock. The stage
   ceiling and the stage-1 charter lock govern what the *guardian* may do;
   this table only governs what the *layout* shows by default. A player at
   any stage can always reach any surface's full content — only its
   always-visible chrome placement varies.
4. These are **layout** defaults; they compose with, but are distinct from,
   the row **fold order** (`patterns/layout.md` ruling a) — see
   "Composition with the fold order" below.

## Per-surface stage defaults

| Surface | Stage 1 | Stage 2 | Stage 3 | Stage 4 | Pre-ladder | Narrow |
|---|---|---|---|---|---|---|
| Lesson row (`panels/lesson-row.md`) | **on** | **on** | badge + overlay-only | badge + overlay-only | badge + overlay-only | same as widescreen (carried, `patterns/layout.md` R3) |
| Guardian strip (`panels/guardian-strip.md`) | **on** | **on** | **on** | **on** | **on** | **on** (carried, R3) |
| Villager strip (`panels/villager-strip.md`) | **on** | **on** | **on** | **on** | **on** | **off** (folds to header count badge, R3 — never carried) |
| Exercise tab (`panels/exercise.md`) | present iff the world carries a scenario | present iff the world carries a scenario | present iff the world carries a scenario | present iff the world carries a scenario | present iff the world carries a scenario | present iff the world carries a scenario (solo-view only, R3) |
| Incident-visibility vocabulary (exercise panel, D4) | `forecast` | `forecast` | `fog` | `fog` | `forecast` (everything) | same as widescreen |
| Systems tab (`panels/systems.md`, once built) | on | on | on | on | on | on |
| Guardian console (`pages/guardian-console.md`) | reachable (own key) | reachable | reachable | reachable | reachable | reachable |
| Help overlay guardian section (D9) | shows stage 1's content | shows stage 2's content | shows stage 3's content | shows stage 4's content | shows the pre-ladder (all-verbs) variant | unaffected by width |
| Unlock ceremony (`overlays/ceremony.md`) | fires stages 1→2, 2→3, 3→4 | ″ | ″ (3→4 only) | never (stage 4 is terminal — nothing unlocks past it) | never (no stage progression exists) | fires identically (takeovers are layout-independent, R3) |
| Postmortem (`overlays/postmortem.md`) | fires on `run.ended`, every world | ″ | ″ | ″ | ″ (ambient/pre-ladder worlds still get the takeover — FR-018's ambient ruling governs its *content*, not whether it fires) | fires identically (R3) |

**Exercise tab and incident vocabulary are world-shaped, not stage-shaped**:
a world either carries a `Manifest.Scenario` block or it doesn't
([[world-save-directory]]); when it does, the tab is present regardless of
the world's `Stage` field, because the two scenarios that ship today
(`first-night`/stage-1, `the-law`/stage-2) already imply their stage by
construction. The incident-visibility *vocabulary value* (`forecast`/`fog`)
is genuinely stage-keyed (D4), independent of which scenario is running.

**Re-verified for spec 054 (TASK-119)**: the two exercise rows above are
now REAL — `internal/tui` gates the tab on the manifest block
(`Model.exerciseID`) and `internal/sim`'s `IncidentVisibilityFor` implements
this table's vocabulary column exactly (definition override wins; forecast
at stages 1–2 and pre-ladder, fog from stage 3), pinned by
`TestIncidentVisibilityVocabulary`.

**Shipped for spec 066 (TASK-128)**: every row above is now REAL code, one
table, one resolver — `internal/tui/stagedefaults.go`'s `stageDefaultsTable`
mirrors this page cell-for-cell (`TestStageDefaultsSweep` parses this page
at test time and fails the build on any drift), and
`resolveStageDefaults(stage, hasScenario)` is the single resolution
function every governed row above reduces to:

- **Lesson row**: `layout.go`'s `lessonRowDefault` now delegates to
  `resolveStageDefaults(...).LessonRowOn` (a pure refactor of the pre-066
  stage-1/2-only check research.md R6 flagged for this table to absorb —
  same behavior, one source now).
- **Guardian strip, systems tab, guardian console**: every column reads
  "on"/"reachable" — nothing gates them by stage in `internal/tui` (they
  were already unconditional); `resolveStageDefaults` carries their values
  for completeness and the parity sweep, not because a gate needed adding.
- **Exercise tab, incident vocabulary**: unchanged since spec 054 above;
  `resolveStageDefaults`'s `ExerciseTabOn`/`IncidentVocabulary` fields
  mirror `Model.exerciseID`/`IncidentVisibilityFor` for one consolidated
  starting-visible-set view (`SC-001` frame assertions), while the real
  render call sites keep calling those functions directly (the latter
  alone carries a per-definition override this table doesn't express).
- **Villager strip**: resolved (`VillagerStripOn`) — real as of spec 060
  (TASK-129, `panels/villager-strip.md`: shipped): `villagerStripView`
  renders it and `computeRows`/`Model.villagerCountBadge` implement its
  fold-to-badge behavior. `resolveStageDefaults`'s `VillagerStripOn` is
  unconditionally true (the table's own "on" at every stage), matching
  `computeRows`'s own no-toggle treatment (`internal/tui/layout.go`'s
  comment on the villager strip's fold accounting) — nothing in
  `internal/tui` actually branches on `VillagerStripOn` yet since there's no
  stage variance to gate; it exists here for completeness and the parity
  sweep, the same posture as the guardian-strip/systems-tab/guardian-console
  bullet above.
- **Unlock ceremony, postmortem**: fire rules, layout-independent by
  design (FR-008) — carried in the table for parity only; TASK-127's
  overlays are a separate, not-yet-merged feature.
- **Help overlay guardian section**: shipped for real by spec 063
  (TASK-115, PR #100, merged during this feature's own implementation) —
  `help.go`'s `helpGuardianLines` already reads `m.currentStage()` with the
  identical fail-open pre-ladder posture (nil/unrecognized status → the
  all-verbs variant) this table's own R3 rule states, sourced from
  `world.StagesLadder`/`guardian.StageCeilingVerbs`/the world skin rather
  than this table. `resolveStageDefaults`'s `HelpGuardianVariant` field
  names which variant that same resolution selects — a second, test-only
  view of the same fact (`TestResolveStageDefaultsHelpGuardianVariant`),
  not a competing implementation.
- **Live re-resolution / explicit overrides / first-occurrence arrival**
  (US3): `tui.go`'s `statusMsg` handler diffs the resolved stage between
  polls (never the first poll) and would route any newly-on governed
  surface through the existing `lessonCatalog` first-occurrence machinery
  (`stagedefaults.go`'s `newlyOnSurfaces`/`announceSurfaceArrival`); under
  the table AS WRITTEN ABOVE this is a no-op on every real transition — no
  row ever widens going up the ladder, only narrows or stays constant —
  proven directly against synthetic fixtures
  (`TestAnnounceSurfaceArrivalExactlyOnce`) rather than any live trigger.
  A `surfaceOverrides` map exists with the correct re-resolution
  precedence, but no in-session command sets one yet.

## Composition with the fold order

Stage defaults decide the **starting** visible set before a terminal's
height forces anything to fold; `patterns/layout.md`'s ruling (a) fold
order (legend → villager strip → lesson row → guardian-strip relocation)
then applies unconditionally on top of whatever that starting set is:

- At **stage 1–2** on a short terminal, all three new chrome rows
  (villager strip, lesson row, guardian strip) start visible, so the fold
  order may need to shed all three in sequence before `bodyMin` is
  satisfied.
- At **stage 3+**/pre-ladder-defaulted-off cases, the lesson row already
  starts folded (badge+overlay), so the fold order has strictly less work:
  only the villager strip and (last) the guardian strip remain foldable.
- The **guardian strip never starts folded** at any stage (decision 7:
  always visible) — it only ever leaves its own row through the fold
  order's relocation step, never through a stage default.

No stage default ever produces a *narrower* fold order than
`patterns/layout.md` states — stage-defaults only changes which rows are
already-folded before fold pressure begins, never the order itself.
