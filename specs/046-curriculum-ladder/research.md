# Phase 0 Research: The curriculum ladder

**Spec**: `specs/046-curriculum-ladder/spec.md` | **Date**: 2026-07-25

Grounded against code at ffb4207. No NEEDS CLARIFICATION remain.

## R1 — Stage is an additive `world.json` Manifest field, write-once

**Decision**: `Stage string json:"stage,omitempty"` (values `stage-1..stage-4`) +
`StageOverridden bool json:"stage_overridden,omitempty"` on `world.Manifest`
(world.go:29-61). Absent = pre-ladder world = ungated (stage-4 semantics) — the spec's
compatibility edge case. Value validated at `world.Open` (the `MemoryRelevance` closed-
vocabulary switch precedent, world.go:173-177). Set at creation only (`promptworld new
--stage`, the `--teaching` set-after-create pattern, commands.go:157-165); **no toggle
command exists or will** — stage is immutable for the world's lifetime (FR-002), so
`SetTeaching`-style mutation is deliberately not replicated.

**Rationale**: additive-omitempty is the documented no-format-bump pattern (world.go:
49-52); manifest fields already carry world-creation-time posture (`Teaching`,
`Meeting`, `MemoryRelevance`).

## R2 — Stage gating = a ceiling intersection at the manifest-load sites

**Decision**: a stage→ceiling table (per the spec's ladder: allowed tool names +
miracle kinds per stage) intersected into the `grantSet` immediately after
`loadManifest`, exactly like bundle narrowing (`narrowGrantForBundles`,
charter.go:382-450 — intersection-only). Two call sites, both must apply it:
`runTurn` (turn.go:145-152, before `grantedRoster` at :164) and the status twin
(turn.go:679). Because the intersection happens before `grantedRoster`, all three
gating layers (declared roster, prose guidance, door checks) inherit it from the one
roster — the spec-021 can't-disagree property holds for stages with zero extra wiring.

**Rationale**: FR-004 structural absence falls out of the existing derivation;
intersection-only means a player's own `capabilities.json` can still narrow further
but never exceed the stage.

## R3 — Stage-1 instruction lock: preset-sourced effective charter + honest notice

**Decision**: `CharterPreset string json:"charter_preset,omitempty"` on the Manifest
(`""`/`default` | `tutor`). At stage-1, the turn pipeline serves the preset constant as
the effective charter and skips `loadSkills`; if `charter.md`'s bytes differ from the
preset (player edited it), the existing notice machinery (notices collected
turn.go:143-158, prefixed at :294-296) appends: *"charter.md does not bind at this
stage — The Written Word (stage-2) unlocks instruction authoring"*. Status
(turn.go:674-689) reports the lock the same way. Stage-2+ behaves exactly as today.
`loadCharter`'s missing-file restore and `charterIsDefault` become preset-aware.

**Rationale**: FR-005's "honest notice, not silent ignoring" maps 1:1 onto the shipped
notice channel; sourcing from the preset constant (not the file) makes the lock
tamper-proof rather than advisory.

## R4 — Per-user unlocks record: `~/.promptworld/unlocks.json`, registry doctrine

**Decision**: a new file beside `known_worlds.json` (`worlds.Root()`, home.go:19-52),
managed in `internal/worlds`: load-tolerant read (missing/corrupt → empty, never an
error — registry.go:29-54), `.tmp`+`os.Rename` atomic write (registry.go:92-109),
**advisory-never-authority** (FR-008: worlds' histories are the authority; entries
carry stage, world name+path, evidence pointers (event seqs/types + tick), earned-at
tick). Home-dir failure = warn-and-continue (the endpoint-lease precedent,
lease.go:52-58).

**Rationale**: the registry is the only cross-world store in the codebase and its
doctrine ("every read tolerates lies, every write heals") is exactly the spec's
auditable-convenience posture. No XDG/UserConfigDir precedent exists — `~/.promptworld`
is the established user scope.

## R5 — Pass/unlock events: executor-emitted contract now, activation with TASK-119

**Decision**: two new event types under a new `curriculum.*` namespace:
`curriculum.exercise_passed` (executor-emitted, pure function of (state, tick) — the
`metatron.order_expired` precedent, loop.go:233-235: **no whitelist entry**) and
`curriculum.stage_unlocked` (derived from a pass; same emission class). This feature
ships: payload structs, reducer arms (no-op-safe), `familyByNamespace` row for
`curriculum` (grammar.go:75-88), `digestRegistry` rows + fixture rows +
`docs/wiki/event-types.md` rows (TestCatalogSweep enforces the trio,
digest_test.go:210-245), a `chronicleNote` case narrating the unlock
(narrate.go:56-262), and unlock-record writing on observing the event. **What emits
the pass in production is TASK-119's rubric machinery** (unbuilt; its AC #2 puts
scenario definitions in world config) — until it lands, tests emit fixture pass events
to prove the chain end-to-end (spec assumption: gates activate as 119's signals
arrive).

**Rationale**: the seam is the deliverable; contract-first matches how 044 defined
`run.ended` for 119's future consumption.

## R6 — Tutor preset: authored constant in `internal/persona`, selected at `new`

**Decision**: `persona.TutorCharter` const beside `DefaultCharter`
(persona/charter.go:7-33); `promptworld new --stage stage-1` seeds it by default at
stage-1 (opt-out to plain default), via `persona.Genesis`'s charter-seeding step
(files.go:63-70) parameterized by preset. Absent-safe by construction: it's charter
text — no model, no new mechanics (FR-012).

## R7 — Exercises ship as content contracts shaped for TASK-119's config block

**Decision**: the two exercise definitions (*first-night*, *the-law*) ship as (a) a
contract doc (contracts/exercises.md — stage, seed, framing, rubric terms, pass
signal, score-narrative framing) and (b) Go data structs + a `scenario`-block Manifest
shape reserved additively (the `Meeting *MeetingConfig` block precedent,
world.go:40-45, consumed at boot by `seedMeetingConvention`, daemon.go:466-473).
TASK-119 implements the consumption; 046 guarantees the definitions parse and the
rubric terms are event-derivable.

## R8 — Stage identity display: interim skin table

**Decision**: a `stageIdentity` table (id → display name, one-line description) in a
new `internal/skin` package stub holding the default guardian strings (The Voice /
The Written Word / The Craft / The Stewardship — client-approved). TASK-121 absorbs
this package as the skin substrate; stage ids and semantics never move.

## R9 — Creation UX: `promptworld stages` + informed `--stage` with `--override`

**Decision**: `promptworld stages` lists the four identities (name, concept, grants,
unlock evidence, earned-or-not from the unlocks record). `promptworld new --stage
<id>`: earned stages proceed; unearned stages error with the informed message naming
the skipped concept(s) unless `--override` is given, which records
`StageOverridden: true` in the manifest. Default `--stage` = stage-1 for new players
(no unlocks record), else highest earned. Human + `--json` symmetric.

**Rationale**: CLI-first matches every existing surface; the informed-choice contract
(FR-003) is a message + explicit flag, not an interactive prompt (nothing in `cmdNew`
is interactive today).

## R10 — Status/TUI visibility

**Decision**: `WorldStatus` gains `Stage`/`StageOverridden` (additive omitempty,
protocol.go:171-175; the `Posture` comment at protocol.go:160-169 already anticipates
the TASK-68 carrier); `postureStatusLine` precedent (commands.go:845-857) renders the
stage line; the TUI metatron console summary (tui.go:185-222, views.go:1425-1442)
gains the stage + granted-surface line. Unlock moments surface via the chronicle
(R5) — FR-009's two surfaces.

## R11 — Sequencing dependencies (recorded, not blocking)

`metatron.charter_observed` (stage-2 gate evidence) is being implemented on the
task-31 branch (044 US2) — in flight, not on main at plan time. 046's gate
*definitions* reference it by contract; the fixture-driven tests stub the evidence
until 044 merges. TASK-121 (skins) and TASK-119 (scenarios) consume/replace R8/R7
seams later; TASK-67 (forks) gets the no-double-count contract via evidence pointers
naming a specific world path + event seqs (R4).

## R12 — Test strategy

(a) Manifest round-trip + validation (stage enum, absent = ungated); (b) gating: per
stage, the derived roster equals the ladder table exactly (table-driven over
`grantedRoster` post-intersection), doors refuse beyond-stage acts, prose/declaration/
door coherence; (c) stage-1 lock: edited charter.md + skills ignored with notice,
identical world mechanics across stages (same-seed determinism run at two stages
diffing world events only); (d) unlocks: fixture pass event → record entry with
evidence pointers → `stages`/creation honor it; corrupt/missing record tolerated;
(e) catalog sweep + chronicle line; (f) player-docs freshness gate over the four new
quickstarts.
