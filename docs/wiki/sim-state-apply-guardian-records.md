---
name: sim-state-apply-guardian-records
description: sim.State.Apply's guardian-facing record arms — the standing-order lifecycle (place/trigger/cancel/expire), charter_observed fingerprint tracking, morgue epilogues, the guardian report card, curriculum unlock latching, and the tuning-dial snapshot with its nil-safe accessors. Split from [[sim-state-apply-world]].
kind: component
sources:
  - internal/sim/guardian.go
  - internal/sim/reportcard.go
  - internal/sim/curriculum.go
  - internal/sim/morgue.go
verified_against: 72f82f41f7aa2e345572105894cd0fb7c02fc0aa
---

# Sim state: guardian-facing record dispatch arms

Split from [[sim-state-apply-world]] (corpus-spec v2 size-budget split,
summary-style): the `Apply` arms that record guardian-issued and
curriculum/tuning state — standing orders, charter observation, morgue
epilogues, the guardian report card, curriculum unlocks, and the tuning
snapshot. All route through `applyGuardian`/`applyReportCard`/
`applyCurriculum`/`applyMorgueEpilogue` alongside `guardian.charge_regenerated`/
`guardian.nudged`'s own `applyGuardian` arm ([[sim-state-apply-world]]).

## How it works

**Standing orders** (since spec 029, `applyGuardian`): `guardian.order_placed`
validates and appends (id uniqueness, origin, non-empty `event_types`, a
1..7-game-day ttl, valid agent index, condition/action length caps, and —
player-origin only — the 3-order active cap) then prunes to the active set
plus the most recent 32 non-active; `guardian.order_triggered`/
`guardian.order_cancelled`/`guardian.order_expired` each transition one
order from active to a terminal status via shared
`transitionGuardianOrder`, rejecting an unknown id or one not active
([[guardian-orders]]).

**Charter observation** (since spec 044 US2, `applyGuardian`):
`guardian.charter_observed` validates a non-empty fingerprint (so
`InjectSocial`'s dry-run refuses a blank one at the door) then sets
`State.CharterFingerprint` — state keeps only the CURRENT fingerprint, the
full revision timeline being the log's observation sequence the [[morgue]]
aligns each death against.

**Morgue epilogues** (spec 044 US2): `morgue.epilogue` dispatches to
`applyMorgueEpilogue` in `morgue.go`: validate the agent index (`-1` =
run-end epilogue) and non-empty text, then append the bounded
`State.MorgueEpilogues` ring (`morgueEpilogueCap` 32).

**The guardian report card** (spec 063, [[grounded-feedback]]):
`guardian.report_card` dispatches to `applyReportCard` in `reportcard.go`:
validate-not-clamp — non-empty fingerprint, non-empty note capped at 1200
runes, and every cited seq strictly less than the event's own seq (a card
can never cite the future) — then keeps only the LATEST card on
`State.GuardianReportCard`; the log alone carries every prior card.

**Curriculum unlocks** (spec 046, [[curriculum-ladder]]): the
`curriculum.*` pair dispatches to `applyCurriculum` in `curriculum.go` —
validate-not-clamp, the guardian arm's contract, both types being the
executor emission class (pure functions of recorded state: a landed event
always re-applies cleanly in replay, a malformed fixture is rejected at the
door): `curriculum.exercise_passed` checks a non-empty exercise id and the
closed stage vocabulary (`validLadderStage`, `stage-1`..`stage-4` — the
reducer-side twin of `world.ValidStage`, kept local so the deterministic
core never imports the save-directory package) then appends the bounded
pass ring; `curriculum.stage_unlocked` also rejects `stage-1` (the ladder's
unearned floor — only stages 2..4 ever unlock) and any stage already
latched (once per world per stage), and does NOT cross-check
`CurriculumPasses` — that ring is pruned past 32, so the gate-conjunct
evaluation (`EvaluateUnlock`) happens at emission time, never on re-apply.

**Tuning snapshot** (spec 048, [[world-tuning]]): `sim.tuning_applied`
joins this validate-not-clamp family: the payload is always the FULL
effective dial set — the five base dials plus, since spec 098, the resolved
dream block ([[private-dreams]]; a pre-098 event's absent block stays nil ≡
defaults) — (never a delta, never re-clamped here — clamping
is `ParseTuning`'s job daemon-side), so the arm is a pure, idempotent
`s.Tuning = &TuningState{...}` assignment — replay re-applies it
identically, and the daemon boot seed never double-counts. `State.Tuning
*TuningState` (`omitempty`, no `format_version` bump) is nil until the
first such event; nil reads as the default dial set through the nil-safe
accessors (`RefuelDyingBelow()`, `FireBurnPerWood()`, `GruEmergePerMille()`,
`PlannerCadence()`, `EncounterCooldown()`) every other promoted call site
(`agents.go`'s fire-fuel arm ([[sim-state-apply-agents]]), [[reflex-policy]],
[[gru]], [[agent-mind]]'s cadence/encounter scheduling) reads through
instead of the retired raw constants.

## Connections

Parent [[sim-state-apply-world]] summarizes this note and its sibling
`world.migrated`/`world.forked` arms. [[guardian-orders]] owns the
standing-order lifecycle this dispatches; [[morgue]] owns
`applyMorgueEpilogue` and the death-alignment timeline;
[[grounded-feedback]] owns the report card's consumer (spec 063);
[[curriculum-ladder]] owns `EvaluateUnlock` and the unlock evidence
constructor; [[world-tuning]] owns the manifest and dial defaults;
[[reflex-policy]], [[gru]], [[agent-mind]], and [[sim-state-apply-agents]]
read the nil-safe tuning accessors.
