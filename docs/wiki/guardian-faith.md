---
name: guardian-faith
description: The endogenous mana loop (spec 085) — event-sourced village faith (FaithState, the faith.changed executor sweep over the five-reason delta table) and charge regeneration as a pure faith-band function (FaithRegenCadenceTicks, the scenario-spiral/ambient-floor posture fork). The prophecy channel splits to [[guardian-prophecy]]. Load when tracing faith movements or the regen cadence.
kind: component
sources:
  - internal/sim/faith.go
  - internal/sim/executor.go
  - internal/sim/guardian.go
  - internal/tool/registry.go
verified_against: 63390f122bdf4e1b7abf518a8be83de725f06230
---

# Guardian faith and prophecy — the endogenous mana loop

Spec 085 (TASK-118) closes the god-game mana loop: power derives from the
prosperity of the flock. Better prompting → fulfilled directives and true
prophecies → more faith → faster charge regeneration → more capacity to
help the village. Faith is a world fact kept strictly in fiction — never a
badge, streak, or score surface (overjustification caution; the
`faith.*` prefix is BANNED from every exercise rubric,
`TestRubricHygieneNoTutorLaneTerms`).

## Faith state and the one writer (`internal/sim/faith.go`)

`State.Faith *FaithState{Score int}` (`omitempty` — pre-085 snapshots
round-trip byte-identically, no format bump). **nil means genesis 50**
(`FaithGenesis`), read ONLY through the nil-safe `s.FaithScore()` — the
`State.Tuning` nil-means-default precedent; the first fold materializes
the struct. The `faith.changed` reducer arm (`applyFaith`) is the ONLY
writer: it validates the reason domain and delta SIGN against the
doctrine table (magnitudes fold as recorded — dial-ready retuning never
breaks old logs), then folds `Score = clamp(Score+delta, 0, 100)`.

`faith.changed{delta, reason, source_id}` is **executor-emitted only** —
NOT on `injectSocialWhitelist` ([[sim-loop-injection-doors]]): whitelist
absence refuses any injected forgery (a console-injectable score would be a
cheat surface). Faith derives EXCLUSIVELY from recorded `faith.changed`
events — never retroactively from old `directive.fulfilled` rows, so
pre-085 logs replay byte-identically (research R1).

## The five-reason delta table (one home, promoted-dial-ready)

| Reason | Source event (scanned in batch) | `source_id` | Delta |
|---|---|---|---|
| `directive_fulfilled` | `directive.fulfilled` — the [[guardian-designations]] TASK-118 seam, consumed | directive id | **+8** |
| `directive_expired` | `directive.expired` | directive id | **−4** |
| `villager_died` | `agent.died` | agent index (decimal string) | **−6** |
| `prophecy_fulfilled` | `prophecy.fulfilled` | prophecy id | **+12** |
| `prophecy_failed` | `prophecy.failed` | prophecy id | **−15** (asymmetric vs +12 — claim-spam is negative-EV) |

`faithEvents(s, batch, nextTick)` is the accounting sweep
([[executor-tick-subsystems]]): pure over (pre-tick state, THIS tick's
batch, tick) — the run-end detector's idiom — scanning the batch for the
five sources and emitting one `faith.changed` per source in batch order,
positioned AFTER every source emitter and BEFORE the scenario rubric and
run-end detection, so companions land in the same batch. The **cannot-move
gate** skips any emission whose fold could not move the clamped score
(accounting for this batch's prior faith emissions — the
`charge_regenerated` below-cap idiom); a partial clamp still emits the
doctrine delta (the fold clamps). Deliberately excluded sources (research
R3): a bare `designation.fulfilled` (villager initiative is not the
guardian's word), ambient accrual, time decay, tutoring (TASK-112 AC #6),
`metatron.nudged`.

## Regen as a pure faith-band function (the FR-005 posture decision)

The fixed 6-game-hour regen constant became one row of a band table.
`FaithRegenCadenceTicks(score, scenario)` — exported; the executor's
boundary check and the daemon's status projection both read THIS function,
so the wire and the sim can never disagree ([[ipc-protocol]]):

| Band | Score | Cadence |
|---|---|---|
| fervent | ≥ 75 | 4 game hours |
| steady | 40..74 | 6 game hours (`chargeRegenTicks` — the genesis band: a world with no faith events keeps the pre-085 schedule byte-identically) |
| wavering | 15..39 | 12 game hours |
| forsaken | < 15 | **scenario: 0 — no regen (the authentic spiral)** · **ambient: 24 game hours (the floor)** |

The check keeps the pre-085 shape — `nextTick % cadence == 0 && charges <
cap`, absolute boundaries, the same `metatron.charge_regenerated` empty
payload — so replay determinism is inherited. Cadence 0 short-circuits the
check entirely. The **posture fork** keys on the boot-frozen
`s.scenario != nil` (the spec-054 incident-sweep precedent): a run-shaped
scenario world may die of the spiral (the morgue teaches); a persistent
ambient world floors at once per game day, and the charge-free plan verbs
([[guardian-designations]]) keep directive-earned faith reachable with an
empty bank — the endogenous exit. The REVERSAL LEVER is the one band table
plus the fork argument; the recorded future promotion is `tuning.json
faith_floor_cadence_ticks` (0 = no floor). `GuardianChargeRegenTicks`
stays exported as the steady-band value — the TUI forecast fallback
against a pre-085 daemon.

## Prophecy — the staked vision

The charge-priced `prophesy` tool, `sim.Prophecy`'s entity discipline, the
closed claim-kind predicate table, and the fulfil-before-fail verification
sweep split into [[guardian-prophecy]].

## Surfaces

- **Tool**: `prophesy` (Gate Charge 1, appended last to `guardianTools`;
  claim_kind Enum over the closed vocabulary; kind-conditional params
  handler-refused when partial/foreign — `assembleClaim`,
  `internal/guardian/prophecy.go`). Stage availability follows
  `send_vision` (`stage1CeilingTools`). `prophecy.*` (3) join `observableEventTypes` (enum-only,
  16 → 19 — standing orders can watch "when my prophecy fails");
  `faith.changed` deliberately does NOT (v1: the strip already surfaces
  it continuously).
- **Turn prompt** (FR-013): the faith score in fiction (`faithBandWord`)
  plus active prophecies (id, claim, days left — `writeProphecies`,
  [[guardian-turn-loop]]).
- **Wire**: `ClockStatus.Faith *int` + `FaithRegenTicks int64` (the
  EFFECTIVE cadence, 0 = no regen scheduled), served from the two sim
  functions ([[ipc-status-extensions]]); the TUI strip's fourth segment
  renders `faith N` / `faith —` (nil pointer = older daemon) with the
  forecast on the wire cadence (`docs/design/tui/panels/guardian-strip.md`
  §4).
- **Chronicle**: four in-fiction digest rows (devotion/doubt wording,
  never numbers first); the `first-faith-event` lesson (direction-neutral
  copy) closed the spec-077 FR-020 rider ([[tui-input-help]]).

**Scope guards** (FR-015): no metatron agentization, no mission machinery,
no faith from tutoring, no faith influence on villager behavior/prompts/
cognition (`internal/cognition`/`internal/mind` untouched), no
rumor-driven faith spread (companions personal, `Subject: -1`), no
tuning.json promotion, no badge surface.

## Spec 086 — agent-named payloads

`faith.changed` keeps its FROZEN three-field shape for every reason except
one: when `reason` is `villager_died` the payload now also carries an
additive `agent` ref — `Agent *AgentRef `json:"agent,omitempty"`` — the
named `{id,name}` form of the index the string `source_id` still encodes
(the SourceID encoding is untouched; it also names directive/prophecy ids
for the other reasons and stays on the sweep's allowlist). `omitempty`
keeps every non-death reason's emission byte-identical to spec 085's shape
(`TestFaithChangedByteShapes`). `prophecy.declared` now rides the wire as
the `ProphecyDeclaredPayload` mirror — `Targets []AgentRef` plus a claim
mirror whose `Agent *AgentRef` is set only for the `survives` kind (index 0
carries a FULL ref — the struct form has no omitempty-zero hazard) — while
the state `Prophecy`/`ProphecyClaim` keep bare ints; the arm folds `.ID`s
(`internal/sim/prophecy.go`).

## Connections

[[guardian-prophecy]] is this note's own split-off child — the staked-vision
channel whose fulfilled/failed outcomes feed the delta table above.
[[guardian-designations]] shares the TASK-118 seam (`directive.fulfilled`/
`directive.expired`); [[guardian-orders]] is the entity discipline both
prophecies and standing orders clone; [[ipc-protocol]]/[[ipc-status-extensions]]
serve the faith/regen wire fields; [[guardian-turn-loop]] renders the turn
prompt's faith/prophecy lines.
