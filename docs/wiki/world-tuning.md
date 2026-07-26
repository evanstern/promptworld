---
name: world-tuning
description: The spec-048 world tuning manifest (tuning.json) — the promotion path for doctrine constants to per-world dials, clamp-validated at boot, event-logged as sim.tuning_applied so replay never reads the file, and the five first-promoted dials (refuel-dying window, fire burn per wood, gru emergence chance, planner cadence, encounter cooldown)
kind: component
sources:
  - internal/sim/tuning.go
  - internal/daemon/daemon.go
  - internal/world/world.go
verified_against: d304e8adb64fdf40e24bfeca3ca3420e8a840a35
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

**The five promoted dials** (`internal/sim/tuning.go`): `RefuelDyingBelow`
(the reflex's fire-refuel-dying window, default 10800 ticks / 3 game-hours —
raised from 3600 by spec 057 / TASK-108 on world-01 burnout evidence),
`FireBurnPerWood` (fuel added per wood, default 14400 ticks / 4 game-hours),
`GruEmergePerMille` (nightly emergence chance, default 600 per mille),
`PlannerCadenceTicks` (the mind driver's per-agent baseline cadence, default
1800 ticks / 30 game-minutes), and `EncounterCooldownTicks` (the per-pair
encounter cooldown, default 7200 ticks / 2 game-hours). Each was previously a
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

**Boot seeding** (`daemon.go`'s `seedTuning`, spec 048): follows the
`seedMeetingConvention` shape — build event → `state.Apply` →
`st.AppendEvents` — so the seed lands in the log like genesis. Called right
after `seedMeetingConvention` and before the loop starts and before
`mind.New`, so no tick and no planner schedule ever runs ahead of the tuned
values ([[daemon-lifecycle]]). Behavior:

- **Absent file**: seeds nothing — state keeps whatever tuning it already
  carries (defaults for a fresh world, or values already in the log from a
  prior boot's event). This is the "absent tuning.json == current behavior"
  invariant (FR-001).
- **Present file**: `ParseTuning` clamps and warns (each warning printed as
  `daemon: <warning>`) or fails boot. The resolved effective set is compared
  against `state.EffectiveTuning()`; identical sets append nothing (a
  restart with an unchanged file never grows the log), and a differing set
  prints one `daemon: tuning.json applied: <all five fields>` line, builds
  the event via `sim.NewTuningEvent(state.Tick, *parsed)`, applies it, and
  appends it.

**The genesis pin** (spec 057 / TASK-108): `promptworld new` seeds one
`sim.tuning_applied` event carrying the full current default set among the
genesis events (`GenesisTuningEvent` in `tuning.go`, appended right after
`world.created` in `cmd/promptworld/commands.go`), so a post-057 world's
effective doctrine is fixed in its own log at birth — later changes to any
`default*` constant never rewrite its replay. `promptworld migrate`
deliberately does NOT back-fill the pin: pre-057 and migrated worlds follow
compiled defaults, a documented determinism hazard
(`control-surface-and-calibration.md` §6, the TASK-75 class). The boot seed
compares a manifest against the pinned set exactly as before — no
`seedTuning` change was needed.

**Replay and file independence** (FR-005/FR-007): replay derives tuning
values EXCLUSIVELY from `sim.tuning_applied` events in the log — defaults
until the first one (which for post-057 worlds is the genesis pin), never
`tuning.json` itself. A world that ran under tuned values replays identically
even if the file is later edited, deleted, or the whole directory copied
elsewhere; a pre-048 log simply contains no such event, so pre-048 worlds and
logs load and replay unchanged on compiled defaults. Editing `tuning.json`
while the daemon runs has no effect until the next boot (no hot exposure —
that is §6 step 3 and explicitly out of scope here).

## Connections

[[world-save-directory]] hosts `TuningPath()` (`tuning.json`, a `manifest.json`
peer, never validated by that package); [[daemon-lifecycle]] calls
`seedTuning` right after the meeting-convention seed, before the loop and
`mind.New`; [[sim-state-reducer]] owns `State.Tuning` and the one
`sim.tuning_applied` Apply arm; [[reflex-policy]] reads `RefuelDyingBelow()`
for the reflex's refuel rung; [[executor]] reads `FireBurnPerWood()` for the
fire-build/refuel fuel window; [[gru]] reads `GruEmergePerMille()` for the
nightly emergence roll; [[agent-mind]] reads `PlannerCadence()` for the mind
driver's per-agent stagger/schedule and `EncounterCooldown()` for the
first-adjacency encounter gate; [[memory-retrieval]] and [[decision-context]]
document the two RNG-bucketing call sites (`memory.go`, `social.go`)
deliberately pinned to the DEFAULT cadence constant rather than this dial;
[[event-types]] catalogs `sim.tuning_applied`'s payload shape. Upstream
design record: `specs/048-tuning-manifest/` (spec.md, data-model.md,
`contracts/tuning.md`), and `docs/design/control-surface-and-calibration.md`
§6, the promotion-path report this feature implements.

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
