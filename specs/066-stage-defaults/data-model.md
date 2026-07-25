# Data Model: Stage-shaped TUI layout defaults (spec 066)

No persisted state is added. All entities are in-memory TUI model state.

## stageDefaultsTable (static, code)

The per-surface × per-stage starting-visibility table, mirrored 1:1 from
`docs/design/tui/patterns/stage-defaults.md` (the authority page) and held to
it by the parity sweep test (contracts/stage-defaults-table.md).

| Field | Meaning |
|---|---|
| surface id | stable key per governed surface (lesson-row, guardian-strip, villager-strip, exercise-tab, incident-vocabulary, systems-tab, guardian-console, help-guardian-section, ceremony, postmortem) |
| per-stage posture | one value per column: stage-1..stage-4, pre-ladder, narrow — values are the authority page's cell vocabulary (`on`, `badge+overlay-only`, `world-shaped`, `forecast`/`fog`, `reachable`, fire rules) |

Validation: the sweep test parses the page's markdown table and asserts
cell-for-cell equality; unknown surface ids or missing columns fail the build.

## startingVisibleSet (derived, per resolution)

The resolved set of default-visible surfaces for one world at one moment.
Derived by `resolve(stage, world-shape)`:

- `stage == ""` or unrecognized → the union (pre-ladder) column, fail-open.
- world-shaped axes (exercise tab presence ← `Manifest.Scenario`) resolve
  independently of stage; stage-keyed vocabulary (incident forecast/fog)
  resolves from stage alone.

Consumed by the existing fold pipeline (`rowBudget` inputs), tab presence,
and help-overlay section selection. Never consulted by capability machinery.

## surfaceOverrides (session-only)

| Field | Meaning |
|---|---|
| surface id → forced visibility | recorded when the player explicitly toggles a governed surface this session |

Rules: an override outranks re-resolution (stage change never rewrites an
overridden surface); cleared only by session end; never persisted.

## State transitions

```
boot            → resolve(stage₀) → startingVisibleSet
stage change    → resolve(stage₁) applied to all surfaces WITHOUT overrides;
                  newly-appearing surfaces route through first-occurrence lessons
explicit toggle → surfaceOverrides[surface] = player's choice
fold pressure   → patterns/layout.md ruling (a), unchanged, applied on top
```
