---
name: world-tuning-dial-catalog
description: Child of [[world-tuning]] — the per-dial catalog: the five spec-048 promoted dials, spec 097's four observation dials, spec 102's steward_cadence_ticks opt-in, spec 104's needs_checkpoint_minutes regime marker, the full JSON-key/default/clamp table, and the spec-098 dream dial block. Load for any single dial's key, default, bounds, or origin story.
kind: component
sources:
  - internal/sim/tuning.go
verified_against: 9b4ed5aef5bfea50b67fac10f8e2153f065a814d
---

# World tuning — the dial catalog

Child of [[world-tuning]]: every dial the manifest carries, with its origin,
JSON key, default, and clamp bounds. The parent owns the model (TuningState,
nil-safe accessors, the manifest file's parse/fail-closed rules).

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
all five defaults into `tuning.go` renamed `default*` — the single home for
these doctrine values.
**Spec 097 adds four grounded-observation dials** (born as dials, no retired
constants — [[executor-perception-observation]]): two executor-read
(observation dedup window, base salience), two read off the mind's replica
by the belief reconciler (disconfirm retain, confirm boost) — nine total.
**Spec 102 adds `steward_cadence_ticks`** ([[guardian-agentization]]): the
guardian-agentization OPT-IN switch and cadence — 0 = off = default,
nonzero clamps to 600..86400, negatives clamp to 0 (never opting a world in
by accident); `omitempty` on state and payload keeps pre-102 bytes
identical; read via `State.StewardCadence()`.
**Spec 104 adds `needs_checkpoint_minutes`** (K, default 10, clamp [1,60]):
the needs-checkpoint cadence — `agent.needs_changed` emits every K
game-minutes plus immediately on any danger-band/near-death/zero crossing;
K=1 reproduces the per-minute heartbeat byte-for-byte. The field DOUBLES as
the coalescing-REGIME marker: a pre-104 recorded `sim.tuning_applied`
payload (field absent) resolves to 0 = LEGACY — deliberately NOT the
doctrine default, unlike the spec-097 dials — so old worlds keep per-step
`agent.moved` / per-minute needs / `gru.moved` emission and the derived
advancement engine stays structurally inert on their folds
([[sim-state-reducer]]). Because a sparse manifest resolves the absent key
to the default 10, ANY `tuning.json` flips an old world to the regime at
its next boot — deterministic and event-recorded; the `sim.tuning_applied`
arm stamps every advancement watermark (`NeedsSyncTick`, `NeedsEmitted`,
`Gru.Done`) at the flip tick so no event-covered past is ever re-derived.
Accessors: `AmbientCoalescing()` (the regime predicate) and
`NeedsCheckpointK()`.

| Field | JSON key | Default | Clamp bounds |
|---|---|---|---|
| RefuelDyingBelow | `refuel_dying_below` | 10800 | [0, 86400] |
| FireBurnPerWood | `fire_burn_per_wood` | 14400 | [600, 86400] |
| GruEmergePerMille | `gru_emerge_per_mille` | 600 | [0, 1000] |
| PlannerCadenceTicks | `planner_cadence_ticks` | 1800 | [60, 86400] |
| EncounterCooldownTicks | `encounter_cooldown_ticks` | 7200 | [0, 86400] |
| ObservationDedupTicks | `observation_dedup_ticks` | 7200 | [0, 86400] |
| ObservationBaseSalience | `observation_base_salience` | 2 | [1, 10] |
| BeliefDisconfirmRetainPercent | `belief_disconfirm_retain_percent` | 70 | [0, 100] |
| BeliefConfirmBoost | `belief_confirm_boost` | 10 | [0, 100] |
| NeedsCheckpointMinutes | `needs_checkpoint_minutes` | 10 | [1, 60] (0 = legacy, unreachable from a manifest) |

Out-of-range values of a KNOWN field clamp to the nearest bound with an
operator-visible warning (the `llm/config.go` `normalizeTokenBudget` style)
— the clamped value applies and gets recorded. Structural problems
(malformed JSON, wrong types, unrecognized field names) fail `ParseTuning`
outright — a hard boot error naming file and problem (fail-closed): a typo
must never silently run as a no-op.

**The dream dial block (spec 098, [[private-dreams]]).** `TuningState` also
carries `Dream *DreamTuning` — nil ≡ the default dream set, so every
pre-098 snapshot and recorded payload stays byte-identical (`omitempty`,
no format bump). The manifest keys stay flat
(`dream_*_per_mille`/`dream_merge_cap_per_night`; defaults 900/30/500/4/15);
any present key resolves the FULL block. The pointer takes `TuningState` out of bare-`==`
comparability — the boot seed compares via `Equal`, and `State.DreamDials()`
is the nil-safe consumption path.
