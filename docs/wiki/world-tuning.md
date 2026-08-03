---
name: world-tuning
description: The spec-048 world tuning manifest (tuning.json) — the promotion path from doctrine constants to per-world dials: TuningState, nil-safe accessors, fail-closed parse rules. Per-dial catalog (specs 048/097/098/102/104) split into [[world-tuning-dial-catalog]]; boot seeding, genesis pin, and replay independence into [[world-tuning-boot-seeding]].
kind: component
sources:
  - internal/sim/tuning.go
  - internal/daemon/daemon.go
  - internal/world/world.go
verified_against: 4efa712bb90538c9c195d23146077e7fc535e511
---

# World tuning manifest

Spec 048 (TASK-107, `docs/design/control-surface-and-calibration.md` §6) is
the promotion path from "hand-edit a constant and rebuild" to "edit a
per-world manifest and restart": an optional, operator-authored `tuning.json`
in the world directory that moves a small, named set of doctrine constants
onto per-world dials. Every dial defaults to exactly its old constant — an
absent file is byte-for-byte the pre-048 world.

## How it works

**The dial catalog** — every dial (the five spec-048 promotions, spec 097's
four observation dials, spec 102's `steward_cadence_ticks`, spec 104's
`needs_checkpoint_minutes` regime marker, the spec-098 dream block) with
JSON keys, defaults, and clamp bounds lives in
[[world-tuning-dial-catalog]]. Highlights the model below depends on:
`needs_checkpoint_minutes` doubles as the coalescing-regime marker
(absent-in-recorded-payload resolves to 0 = legacy, deliberately not the
doctrine default), and the dream block rides a nil≡default pointer.

**`TuningState`** (`tuning.go`) is the fully-resolved effective set: a
non-nil `TuningState` always carries all fields with defaults filled in
and clamps applied — never a sparse struct. `sim.State.Tuning *TuningState`
(`json:"tuning,omitempty"`) carries it, event-sourced: **nil means the
default set**, which is what every pre-048 snapshot and world has, and
`omitempty` keeps those snapshots byte-identical with **no `format_version`
bump** (the spec-044 `MorgueEpilogues` precedent). `State.EffectiveTuning()`
returns the resolved set either way (`*s.Tuning` or `defaultTuning()`).

**Nil-safe accessors** (`RefuelDyingBelow()`, `FireBurnPerWood()`,
`GruEmergePerMille()`, `PlannerCadence()`, `EncounterCooldown()`, plus one
per spec-097 dial — all methods on `*State`) are the ONLY consumption path: every reducer call site
([[executor]]'s fire-fuel arm, [[reflex-policy]]'s refuel rung, [[gru]]'s
emergence roll) and every mind-side call site ([[agent-mind]]'s per-agent
cadence/stagger and encounter-cooldown gate, reading off `md.replica`) go
through these instead of the retired raw constants — "absent tuning.json ==
current constants" is structurally true, not a convention to remember. Two
RNG-bucketing reuses (serendipity-tail seeding in `memory.go`,
`SecretShareRoll` in `social.go`) deliberately bucket on
`tick/defaultPlannerCadenceTicks` — the DEFAULT constant, never the tuned
dial — so a tuned cadence never reshuffles those seeded picks. `fireFuelCap` (the fuel
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

**The `sim.tuning_applied` event** (`tuning.go`'s `NewTuningEvent`) carries
the FULL effective set — base fields, the four spec-097 pointer fields, and
the RESOLVED dream block — never a delta, so replay can establish tuning
state from any single event without scanning history. Absent spec-097
fields resolve to the doctrine defaults at Apply (`resolveTuning`), never
to zero; a pre-098 event's absent dream block stays nil ≡ defaults. The
reducer arm (`state.go`) is a pure, idempotent `s.Tuning = &t` assignment —
re-applies cleanly on replay, boot seed never double-counts.

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
