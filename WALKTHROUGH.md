# Spec 046 (curriculum ladder) — quickstart.md walkthrough outcomes

Run 2026-07-25 on branch `task-68-curriculum-ladder`, against the tree at the
commit this file is added in. All commands run headlessly in a scratch
`PROMPTWORLD_HOME` under system temp (`/tmp/pw046-walkthrough-home.*`), never
the repo — deleted after the run.

## §1 Hermetic proof

```
go test ./internal/world -run Stage -v          # PASS (3/3)
go test ./internal/metatron -run 'Stage|Ceiling|Preset' -v  # PASS (10/10)
go test ./internal/worlds -run Unlocks -v       # PASS (6/6)
go test ./internal/sim -run Curriculum -v       # PASS
go test ./internal/tui -run TestCatalogSweep -v # PASS
```

All green.

## §2 Creation UX walkthrough

```
$ promptworld stages
The Voice — you speak, it acts (stage-1) ... earned: yes (every player's floor)
The Written Word — ... (stage-2) ... earned: no — choosable only with new --stage stage-2 --override
The Craft — ... (stage-3) ... earned: no — choosable only with new --stage stage-3 --override
The Stewardship — ... (stage-4) ... earned: no — choosable only with new --stage stage-4 --override

$ promptworld new v1 --stage stage-1
created world "v1" ...
stage: The Voice (stage-1)

$ promptworld new v2 --stage stage-3
promptworld new: new: The Craft (stage-3) is not yet earned — creating here
would skip The Written Word, The Craft; pass --override to proceed anyway
(the world's config records the override honestly)
exit 1  # as expected — the informed-override error

$ promptworld new v2 --stage stage-3 --override
created world "v2" ...
stage: The Craft (stage-3) [overridden]

$ promptworld status v1 --json
{"world":{...,"stage":"stage-1"}, ...}

$ promptworld status v2 --json
{"world":{...,"stage":"stage-3","stage_overridden":true}, ...}
```

Every line matches quickstart.md §2 verbatim (the error message names the skipped
stages by skin display name — "The Written Word, The Craft" — and the override
flag is honestly recorded in `world.json` and echoed on status).

`v1/world.json` recorded `"charter_preset": "tutor"` (the stage-1 default, opt-out
available via `--charter-preset default`); `v1/charter.md` was seeded with the
`persona.TutorCharter` text verbatim. `v2/world.json` recorded no `charter_preset`
key (stage-3, no preset stamped — equivalent to `"default"`).

## §3 Gating proof (US2 / SC-002)

Proven by the hermetic suite (§1) rather than a live model call — this sandbox has
no reachable LLM endpoint, and `metatron.Turn` requires one to produce a reply. The
model-free surfaces (`promptworld start`/`stop`/`status`/`metatron` status-peek) were
additionally exercised live end-to-end against a real daemon boot:

```
$ promptworld start v1
daemon started (pid ...): tick 1 (day 1 06:00) — running, speed 4x (4.0 ticks/s effective)

$ promptworld metatron v1     # status peek, no message — model-free
charges ⚡·· (1/3) · default charter · charter.md at .../v1/charter.md
--- recent notes ---
# The soul of Metatron
*The reign begins. The angel has seen nothing yet.*

$ promptworld status v1 --json
{"world":{...,"stage":"stage-1"}, "clock":{...}, "llm":{...}, "horizon":[...]}

$ promptworld stop v1
daemon stopped
```

The daemon boots cleanly with the stage stamped, the status wire carries it, and
the model-free metatron peek runs without touching the network. The full
turn-level proof (roster == ceiling exactly, the stage-1 does-not-bind notice,
skills not composing, cross-stage determinism) is `internal/metatron`'s
`TestStageCeilingRosterTable` / `TestStageOneInstructionLock` /
`TestStageTwoChartersBindSkillsDoNot` / `TestCrossStageDeterminism` — all green
in §1.

## §4 Unlock chain (US3 / SC-003, SC-004 — fixture-driven until TASK-119)

Fixture-driven in Go, per the spec's own framing ("proven now by fixture
emission" — production emission is TASK-119's rubric machinery, unbuilt).
Covered end-to-end by:

- `internal/sim.TestEvaluateUnlockFixtureChain` / `TestEvaluateUnlockGateConjuncts`
  (SC-004 negative case included: a stage-2 pass with default-charter evidence
  does NOT unlock stage-3).
- `internal/daemon.TestCurriculumObserverUpsertsUnlockWithEvidence` (the daemon
  observer upserts `unlocks.json` with an evidence pointer at the proving pass
  event, no orchestrator involved).
- `internal/mind.TestChronicleNoteStageUnlocked` (the chronicle line).
- `cmd/promptworld.TestCmdNewEarnedStageNeedsNoOverride` /
  `TestCmdStagesHumanOutput` (`stages`/`new` honor an unlocks-record entry with
  no override needed).

All green in §1/the full suite (`go test ./...`).

## §5 Docs (US5 / SC-006)

```
$ node .claude/skills/player-docs/scripts/check-freshness.mjs --check
13 fresh, 0 stale, 0 missing, 0 broken-ref
exit 0
```

Four new stage quickstarts (`stage-1-the-voice.html` .. `stage-4-the-stewardship.html`)
bring the page count from nine to thirteen; all fresh.

## §6 Tutor preset (US5)

Confirmed structurally (§2): `promptworld new v1 --stage stage-1` seeds
`persona.TutorCharter` (guardian-voiced, folk-tale tone — watch first, visions,
omens, watches, the chronicle) into `charter.md`, and the manifest records
`charter_preset: "tutor"`. Live in-game delivery through an actual model turn
was not exercised (no reachable LLM endpoint in this sandbox); the hot-reload
behavior above stage-1 and the stage-1 lock/notice behavior are proven by
`internal/metatron.TestTutorPresetHotReloadsLikeAnyCharter` and
`TestStageOneInstructionLock` (§1, green). No-model worlds retain every bit of
this machinery per FR-014 — see the implementer report's FR-014 note for the
precise scope of what was/wasn't additionally proven there.

## Full gate (T020)

```
go build ./...   # clean
go vet ./...     # clean
gofmt -l .       # no output (all formatted)
go test ./...    # ok, all 22 packages (incl. e2e, cmd/promptworld), no FAIL
```

## Post-merge obligation (Principle IV)

This branch touches wiki-pinned files: `docs/wiki/event-types.md` (new
curriculum.* rows + spec 046 prose + sources list, `verified_against` left at its
prior pin — NOT re-verified against this branch's commit, since that's
`/grounding-wiki:wiki-update`'s job, not this implementation's). Run
`/grounding-wiki:wiki-update` after this branch merges, then re-check
`docs/player/` freshness (already green as of this file, but wiki-update may
re-pin sources the stage quickstarts also draw on).

## T022 reconciliation finding — IMPORTANT, read before merging

**Spec 044 US2 (task-31) has ALREADY MERGED to origin/main** (PR #78, commit
`e48bbaa`) during this implementation — found via `git fetch origin` +
`git log origin/main`. This branch (`task-68-curriculum-ladder`) forked before
that merge and has NOT been rebased onto it, so `internal/metatron/charter.go`
has diverged in two different, overlapping directions:

- **This branch** added: `stageCeiling`/`applyStageCeiling`, `presetCharter`,
  `stageCharter`, `stageLocksCharter`/`stageBindsSkills`, and made
  `loadCharter`/`charterIsDefault` take a variadic `preset` argument.
- **origin/main** (via 044 US2) added: `charterFingerprint`, `observeCharter`,
  and a new event `metatron.charter_observed` with payload
  `sim.CharterObservedPayload{Fingerprint string, Default bool}` — emitted
  whenever the effective charter's fingerprint changes from the last recorded
  one. **Note the polarity**: `Default == true` means the DEFAULT charter is in
  force; `Default == false` means a player-authored revision is in force — the
  INVERSE of this branch's `EvidenceRef.Custom` flag (`Custom == true` means
  player-authored).

This branch's stub, exactly where it needs updating once rebased:

1. `internal/sim/curriculum.go`, `EvaluateUnlock`'s stage-2 branch currently
   checks `ev.Type == "metatron.charter_observed" && ev.Custom`. Once rebased
   onto (or merged with) 044, this should read the REAL
   `sim.CharterObservedPayload` and check `!payload.Default` (or keep the
   `EvidenceRef.Custom` indirection but set `Custom = !observed.Default` at
   whatever call site constructs the `EvidenceRef` — TASK-119's rubric
   machinery, since that's what will actually populate `ExercisePassedPayload.
   Evidence` in production).
2. `internal/tui/digest_test.go`'s `pendingCatalogEventTypes` map currently
   excepts `"metatron.charter_observed"` from
   `TestExerciseRubricTermsAreCatalogedEventTypes`. origin/main's
   `internal/tui/digest.go`/`digest_test.go` already carry real
   `digestRegistry`/`catalogFixture` rows for it (spec 044 US2's own T014
   equivalent) — once merged, delete the exception; the term will satisfy the
   check via the real catalog entry like every other term.
3. `docs/wiki/event-types.md`'s curriculum row references
   "metatron charter-observed fingerprint's `custom` evidence entry" in plain
   text (deliberately NOT backticked, to avoid tripping
   `TestCatalogSweep`'s backtick-sweep before 044 was cataloged here) — once
   rebased, this can safely become a real backticked cross-reference.

**This reconciliation was deliberately NOT performed in this worktree.** Merging
origin/main into this feature branch mid-implementation — resolving real
conflicts in `internal/metatron/charter.go` between two independently-designed
extensions to the same functions, and rewiring a cross-feature gate conjunct to
a just-landed payload shape — is a cross-package, capability-gating-adjacent
change under the project's model-tiering rubric (constitution Principle V:
"concurrency/scheduling/governor logic... doctrine-adjacent behavior changes").
It was judged out of scope for this Sonnet-tier slice and is flagged here for
the planning tier to schedule explicitly (likely as its own reviewed rebase
pass, plausibly Opus-tier given the rubric).

TASK-119 (scenario/rubric machinery — the real production emitter of
`curriculum.exercise_passed`/`stage_unlocked`) and TASK-121 (skins — absorbs
`internal/skin`) are both still unbuilt as of this branch; their seams
(`sim.ExerciseDefinition`, `sim.EvaluateUnlock`, `skin.Stage`/`StageName`) are
unchanged by the above and remain exactly as documented in
`specs/046-curriculum-ladder/research.md` R7/R8/R11.
