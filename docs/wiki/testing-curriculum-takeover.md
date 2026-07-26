---
name: testing-curriculum-takeover
description: Staged-worlds curriculum-ladder proof (unlock gate logic, stage ceilings, unlock records, CLI stage resolution) and the takeover dispatch/precedence/dismiss matrix. Split out of [[testing-strategy]].
kind: pattern
sources:
  - internal/sim/curriculum_test.go
  - internal/guardian/stage_test.go
  - internal/daemon/curriculum_test.go
  - internal/worlds/unlocks_test.go
  - internal/world/world_test.go
  - cmd/promptworld/stages_test.go
  - internal/tui/takeover_test.go
  - internal/tui/render_test.go
  - internal/tui/console_test.go
verified_against: b3f4da3c29e3cbbd933e366abe76a5d6ef0f2be9
---

# Curriculum-ladder & takeover suites

**Curriculum-ladder suites** (spec 046, TASK-68, [[curriculum-ladder]]): the
staged-worlds surface is proven per layer. Reducer-side,
`internal/sim/curriculum_test.go` pins the two `curriculum.*` arms — pass
recording with evidence, door validation (empty exercise id, unknown stage),
the pass ring's 32-cap prune, the once-per-(world,stage) unlock latch with
duplicate and stage-1 rejection — plus the pure gate logic:
`TestEvaluateUnlockGateConjuncts` walks all three gates (stage-1: any pass;
stage-2: only a `custom` charter-observed evidence entry — SC-004's negative
case pins that a default/preset-charter pass never opens it; stage-3: any
`custom` evidence entry), `TestCharterObservedEvidence` pins the sanctioned
constructor's `Custom = !payload.Default` derivation and its
wrong-type/bad-payload rejections, `TestEvaluateUnlockFixtureChain` drives
the full fixture pass→unlock chain, `TestExerciseDefinitionsParse` proves the
two shipped exercise definitions are well-formed, and
`TestCurriculumReplayDeterministic` proves a log carrying both types replays
byte-identically. Guardian-side, `internal/guardian/stage_test.go` (US2,
~500 lines) pins the gate-to-feature pathway: `TestStageCeilingRosterTable`
(per stage, the post-intersection roster equals the contract's ceiling
exactly — stage-1/-2 the four-tool watch set, stage-3/-4 and pre-ladder the
full roster), manifest intersection within the ceiling, the door refusing
beyond-stage acts, three-layer declaration/prose/door coherence, the stage-1
instruction lock's honesty (`TestStageOneInstructionLock` — the compiled-in
preset binds regardless of edits, the notice names the unlocking stage),
stage-2 charter-binds-skills-don't, `TestCrossStageDeterminism` (stage gating
never perturbs the sim, FR-006), preset resolution, an ungated pre-ladder
world byte-unchanged, and the tutor preset hot-reloading like any charter;
`charter_observed_test.go` gains the preset legs
(`TestCharterObservationTutorPresetIsDefault` — a stage-1 tutor-preset
world's observation honestly records `default: true`, so it can never
masquerade as player authorship — and
`TestCharterObservationEndedStageOneCoexists`). Daemon-side,
`internal/daemon/curriculum_test.go` drives the always-on unlock observer
with fixture events: upsert with the pass-event evidence pointer, non-
curriculum events ignored, a missing pass in the batch tolerated, and the
recorded path matching the world fixture. `internal/worlds/unlocks_test.go`
pins the record doctrine — missing/corrupt file loads empty (never an
error), atomic upsert-and-reload, same-stage overwrite, load-time healing
that drops malformed entries but KEEPS entries whose world path no longer
exists, and an unresolvable home warning-and-continuing. CLI-side,
`cmd/promptworld/stages_test.go` covers `new`'s stage resolution (stage-1
default for a new player, highest-earned otherwise, unearned refusal naming
skipped concepts unless `--override`, the override recorded honestly,
invalid stage/preset rejected, tutor-preset opt-out) and both `stages`
outputs; `internal/world/world_test.go` gains the manifest legs (stage
round-trip, absent-stage = ungated, `Open` rejecting a bad `stage` or
`charter_preset`), and `internal/skin/skin_test.go` pins the four
client-approved stage identities.

**Takeover suites** (spec 056, TASK-127, [[takeover-surfaces]]):
`internal/tui/takeover_test.go` covers the dispatch/precedence/dismiss
matrix — `run.ended` always wins over an open ceremony, a deferred ceremony
never interrupts the postmortem, `esc`/`q`/`p` behave per mode, and the
per-session `postmortemDismissed` flag survives a mere reconnect but never a
genuine new `run.ended`. `internal/tui/render_test.go` proves the takeover
renders layout-independently (widescreen and narrow alike) and that
`fitTakeoverLines` sheds an overflowing death ledger's tail with an honest
count rather than growing past the panel budget. `internal/tui/console_test.go`
proves the `reportCard` wrapper composes into `Model.consoleCards`
unmodified — the D5 shared-renderer seam. `internal/tui/help_test.go`
extends the keymap/section sweep to the new ceremonies section.

## Connections

Part of the [[testing-strategy]] suite map (split out during the corpus-spec v2
restructure); see that note for the full layered test picture and links to
sibling suites.
