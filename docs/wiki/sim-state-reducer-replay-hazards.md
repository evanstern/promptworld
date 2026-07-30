---
name: sim-state-reducer-replay-hazards
description: Child of [[sim-state-reducer]] — the spec-092 (TASK-75) reducer-constants replay-hazard doctrine and full audit of every internal/sim/state.go Apply arm that re-derives an outcome from a mutable gameplay constant instead of carrying it in the payload; the spec-048 genesis-tuning-pin's partial, narrower mitigation.
kind: pattern
sources:
  - internal/sim/state.go
  - internal/sim/agents.go
  - internal/sim/recipes.go
  - internal/sim/terrain.go
verified_against: cf65debb44c1e17b54c0f3421d11e1e8cc28576c
---

# Reducer-constants replay-hazard audit (spec 092 / TASK-75)

Split from [[sim-state-reducer]] (summary-style, corpus-spec v2) — the note
budget there only has room for the doctrine's one-paragraph statement; this
child carries the full audit and reconciliation.

## How it works

**Default**: the payload carries the outcome; the reducer copies it
verbatim. Spec 019's `agent.memory_added` is the precedent —
`Where`/`Why`/`Conv`/`Origin` are "baked at emission, never re-derived at
render or replay, so live and replay agree" ([[sim-state-cognition-arms]]).
This is the safe shape: a later balance change can never alter what an OLD
log replays to, because the number that mattered was already written into
the event at the time it happened.

**Exception**: several `Apply` arms instead re-derive an outcome from a bare
package constant — reading it fresh at replay time rather than trusting a
payload field. That is harmless as long as the constant never changes; it
becomes a **replay hazard** the moment someone retunes it, because an OLD
log then replays through the CURRENT build's constant, not the value that
was live when the log was recorded. Reducer-re-derives is the exception,
and — since spec 094 shipped (TASK-134, [[event-log]]) — it now REQUIRES
bumping `store.LogFormatVersion` and shipping a migration before the
constant may change: a pure rename translates the log
(`sim.LogFormatV1Renames`); a semantic retune snapshot-cuts
([[world-migration]]'s decision rule). This note is the audit surface that
migration work consumes going forward.

**Partial existing mitigation — spec 048's genesis-tuning-pin**: five of
the constants below were promoted to per-world tuning dials
(`RefuelDyingBelow`, `FireBurnPerWood`, `GruEmergePerMille`,
`PlannerCadenceTicks`, `EncounterCooldownTicks`, [[world-tuning]]), and
since spec 057 every FRESH world pins its effective set into its own log
at genesis ([[world-tuning-boot-seeding]]) — a post-057 world's replay is
immune to a later default change for exactly those five. The residual
scope is already documented there: pre-057 worlds and any world produced
by `promptworld migrate` follow compiled defaults, not a pin — the same
hazard this doctrine describes, for those five dials, on those worlds.
Every OTHER constant audited below has no promotion and no pin at all, on
any world, regardless of age.

## Audit (FR-004) — reducer arms that re-derive from a mutable constant

Swept from `internal/sim/state.go`'s `Apply` (this note's pin); each site
below also carries a short "Replay hazard (spec 092/TASK-75)" code comment.

| `Apply` arm | Site | Re-derived from | Value(s) |
|---|---|---|---|
| `agent.foraged` | state.go:1131 | `forageYieldV2` | 2 |
| `agent.chopped` | state.go:1154-1158 | `chopYieldBare`/`chopYieldAxe` | 1 / 3 |
| `agent.hunted` | state.go:1196-1200 | `huntYieldBare`/`huntYieldSpear` | 8 / 12 |
| `agent.hunted` | state.go:1212 | `denCooldownSec` | 21600 (6h) |
| `agent.built` | state.go:1230 | `recipeFor("build_"+kind)` input costs | recipes.go table |
| `agent.built` (fire) | state.go:1243 | `FireBurnPerWood()` | tuning-covered — see mitigation above |
| `agent.built` (wall) | state.go:1259 | `wallMaxHP(kind)` → `wallPlankHP`/`wallStoneHP` | 200 / 600 |
| `agent.quarried` | state.go:1295-1299 | `quarryYieldBare`/`quarryYieldAxe` | 1 / 3 |
| `agent.collected_water` | state.go:1322 | `collectWaterYield` | 1 |
| `agent.crafted` | state.go:1351-1357 | `recipeFor(goal)` input/output table, `spearDurability`, `axeDurability` | recipes.go table; 3; 10 |
| `agent.wall_chipped` | state.go:1485 | `demolishChipHP` | 100 |
| `agent.wall_repaired` | state.go:1530-1535 | `wallRepairMaterial(kind)`, `wallMaxHP(kind)`, `repairHPPerUnit` | — ; — ; 100 |
| `agent.talked` | state.go:2136-2137 | `talkMoraleBonus` | 50 |

Lower-tier — listed for FR-004 completeness ("every site") but these
re-derive a CLASSIFICATION of an already-carried absolute value, not a
produced resource amount; a retune reclassifies an old log's derived
flags/anchors, never the recorded needs numbers themselves, so the blast
radius is narrower:

| `Apply` arm | Site | Re-derived from |
|---|---|---|
| `agent.needs_changed` | state.go:1873-1876 | `nearDeathBelow`/`nearDeathResetAt` (NearDeath latch) |
| `agent.needs_changed` | state.go:1888 | `trajectoryWindowTicks` (anchor-roll cadence) |
| `agent.needs_changed` | state.go:1901 | `recoveryDangerBand`/neglect band constants |
| `agent.memory_added` | [[sim-state-cognition-arms]] | `GenerationBumpSalience` (9, the generation-bump threshold) |

Not audited here: `clock.degraded`'s `EffectiveRate` is payload-carried
(the loop measures it once and the reducer copies it verbatim) — it is the
DETERMINISM-SCOPE hazard ([[deterministic-rng]], [[sim-loop]]: replay is
exact per-log, but two machines on the same seed can measure different
wall-clock rates), a different concern from this doctrine's
re-derive-from-a-constant hazard. The rename half of this doctrine is
[[event-log]]'s to own; this note owns the re-derive half only.

## Reconciling with the spec-019 precedent

Spec 019 chose emitter-computes for `Memory`'s situated fields specifically
because a memory's phrasing is expensive to re-derive and cheap to carry
once ([[sim-state-cognition-arms]]); this doctrine generalizes that choice
into the DEFAULT posture for every new `Apply` arm, and treats the audited
table above as the accumulated exception set — grandfathered, not a
pattern to keep extending.

## Connections

Back to [[sim-state-reducer]] for the one-paragraph doctrine statement and
the rename half. [[event-log]] is the `store.LogFormatVersion` gate this
audit's migration surface serves; [[world-migration]] is the
translate-vs-snapshot-cut decision rule. [[sim-state-cognition-arms]] holds
the spec-019 emitter-computes precedent. [[world-tuning-boot-seeding]]
documents the spec-048 genesis-pin's residual scope.
