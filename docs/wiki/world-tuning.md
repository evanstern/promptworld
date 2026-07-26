---
name: world-tuning
description: The spec-048 world tuning manifest (tuning.json) — the promotion path for doctrine constants to per-world dials: the five first-promoted dials (refuel-dying window, fire burn per wood, gru emergence chance, planner cadence, encounter cooldown), TuningState, nil-safe accessors, and the manifest's clamp table; boot seeding, the genesis pin, and replay independence split into [[world-tuning-boot-seeding]]
kind: component
sources:
  - internal/sim/tuning.go
  - internal/daemon/daemon.go
  - internal/world/world.go
verified_against: 4c66d240b2715706964f02cfd2396256c9957d8e
---

# World tuning manifest

Spec 048 (TASK-107, `docs/design/control-surface-and-calibration.md` §6) is
the promotion path from "hand-edit a constant and rebuild" to "edit a
per-world manifest and restart": an optional, operator-authored `tuning.json`
in the world directory that moves a small, named set of doctrine constants
onto per-world dials. Every dial defaults to exactly its old constant, so an
absent file is byte-for-byte the pre-048 world — the mechanism is additive,
never a behavior change by itself.

## How it works

**The five promoted dials** (`internal/sim/tuning.go`, defaults and clamp
bounds in the table below): `RefuelDyingBelow` (the reflex's
fire-refuel-dying window — raised from 3600 by spec 057 / TASK-108 on
world-01 burnout evidence), `FireBurnPerWood` (fuel added per wood),
`GruEmergePerMille` (nightly emergence chance), `PlannerCadenceTicks` (the
mind driver's per-agent baseline cadence), and `EncounterCooldownTicks` (the
per-pair encounter cooldown). Each was previously a
bare package constant (`refuelDyingBelow`/`fireBurnPerWood` in `agents.go`,
`gruEmergePerMille` in `gru.go`, the exported `sim.PlannerCadenceTicks` in
`agents.go`, `encounterCooldownTicks` in `mind/mind.go`); spec 048 relocates
all five defaults into `tuning.go` renamed `default*`
(`defaultRefuelDyingBelow`, `defaultFireBurnPerWood`,
`defaultGruEmergePerMille`, `defaultPlannerCadenceTicks`,
`defaultEncounterCooldownTicks`) — the single home for these doctrine values.

**`TuningState`** (`tuning.go`) is the fully-resolved effective set: a
non-nil `TuningState` always carries all five fields with defaults filled in
and clamps applied — never a sparse struct. `sim.State.Tuning *TuningState`
(`json:"tuning,omitempty"`) carries it, event-sourced: **nil means the
default set**, which is what every pre-048 snapshot and world has, and
`omitempty` keeps those snapshots byte-identical with **no `format_version`
bump** (the spec-044 `MorgueEpilogues` precedent). `State.EffectiveTuning()`
returns the resolved set either way (`*s.Tuning` or `defaultTuning()`).

**Nil-safe accessors** (`RefuelDyingBelow()`, `FireBurnPerWood()`,
`GruEmergePerMille()`, `PlannerCadence()`, `EncounterCooldown()`, all methods
on `*State`) are the ONLY consumption path: every reducer call site
([[executor]]'s fire-fuel arm, [[reflex-policy]]'s refuel rung, [[gru]]'s
emergence roll) and every mind-side call site ([[agent-mind]]'s per-agent
cadence/stagger and encounter-cooldown gate, reading off `md.replica`) go
through these instead of the retired raw constants — "absent tuning.json ==
current constants" is structurally true, not a convention to remember. Two
RNG-bucketing reuses of the cadence period are a deliberate exception:
`internal/sim/memory.go`'s serendipity-tail seeding and `social.go`'s
`SecretShareRoll` both bucket on `tick/defaultPlannerCadenceTicks` — the
tuning-manifest DEFAULT constant, never the tuned `PlannerCadence()` dial —
so a tuned world's cadence moves the mind driver's schedule without also
reshuffling these two unrelated seeded picks. `fireFuelCap` (the fuel
ceiling a refuel is truncated to) is deliberately NOT promoted (research R6)
and stays a plain constant beside the promoted `FireBurnPerWood`.

**The manifest file** (`<world-dir>/tuning.json`, `world.TuningPath()`) is
sparse JSON — set only the dials you want to move — a peer of
`manifest.json`/`calibration.json`/`llm.json`, never written by `promptworld
new`. `ParseTuning(data)` (`tuning.go`) decodes it with
`DisallowUnknownFields` (an unrecognized key fails boot rather than silently
no-opping — a typo'd dial must never look like it moved) into a full
effective `TuningState` resolved against the defaults, returning per-field
clamp warnings and a structural error separately:

| Field | JSON key | Default | Clamp bounds |
|---|---|---|---|
| RefuelDyingBelow | `refuel_dying_below` | 10800 | [0, 86400] |
| FireBurnPerWood | `fire_burn_per_wood` | 14400 | [600, 86400] |
| GruEmergePerMille | `gru_emerge_per_mille` | 600 | [0, 1000] |
| PlannerCadenceTicks | `planner_cadence_ticks` | 1800 | [60, 86400] |
| EncounterCooldownTicks | `encounter_cooldown_ticks` | 7200 | [0, 86400] |

Out-of-range values of a KNOWN field clamp to the nearest bound with an
operator-visible warning (`tuning.json <field> <raw> out of range (<bound
kind> <bound>) — clamped to <bound>`, the `llm/config.go`
`normalizeTokenBudget` style) — the clamped value is what applies and gets
recorded. Structural problems — malformed JSON, a wrong-typed value, or an
unrecognized field name — fail `ParseTuning` outright, which the daemon
boot treats as a hard boot error naming the file and the problem
(fail-closed): clamping handles a merely out-of-range value of a
correctly-named field, but a typo must never silently run as a no-op.

**The `sim.tuning_applied` event** (`tuning.go`'s `NewTuningEvent`) carries
the FULL effective five-field set, never a delta, so replay can establish
tuning state from any single event without scanning history. The reducer arm
(`state.go`) is a pure, idempotent `s.Tuning = &TuningState{...}` assignment
— it re-applies cleanly on replay and the boot seed never double-counts.

**Boot seeding, the genesis pin, and replay independence** split into
[[world-tuning-boot-seeding]]: how `daemon.go`'s `seedTuning` applies the
manifest at boot (absent-file no-op vs present-file clamp-and-compare), the
spec-057 genesis pin that fixes a freshly-created world's doctrine at birth,
and why replay derives tuning EXCLUSIVELY from logged `sim.tuning_applied`
events, never `tuning.json` itself.

## Connections

[[world-save-directory]] hosts `TuningPath()` (`tuning.json`, a `manifest.json`
peer, never validated by that package); [[sim-state-reducer]] owns
`State.Tuning` and the one `sim.tuning_applied` Apply arm; [[reflex-policy]]
reads `RefuelDyingBelow()` for the reflex's refuel rung; [[executor]] reads
`FireBurnPerWood()` for the fire-build/refuel fuel window; [[gru]] reads
`GruEmergePerMille()` for the nightly emergence roll; [[agent-mind]] reads
`PlannerCadence()` for the mind driver's per-agent stagger/schedule and
`EncounterCooldown()` for the first-adjacency encounter gate;
[[memory-retrieval]] and [[decision-context]] document the two RNG-bucketing
call sites (`memory.go`, `social.go`) deliberately pinned to the DEFAULT
cadence constant rather than this dial; [[event-types]] catalogs
`sim.tuning_applied`'s payload shape. See [[world-tuning-boot-seeding]] for
who calls `seedTuning` and the upstream design record.

## Operational notes

Five dials promoted so far; every other doctrine constant (`fireFuelCap`,
the talk cooldown, the trajectory window, etc.) stays a plain constant until
its own promotion earns evidence the same way — this manifest is a
mechanism, not a blanket config surface. The recorded follow-on decision
(not yet implemented): once a world's tuned values prove out, they become
the new default for freshly-created worlds — a separate future change, not
this feature's scope. Secondary event consumers (TUI digest, chronicle,
morgue render) must tolerate the new `sim.tuning_applied` event type without
crashing; none of them render it today (no obligation in the spec).
