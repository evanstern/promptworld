# Tasks: World Tuning Manifest (tuning.json)

**Input**: Design documents from `/specs/048-tuning-manifest/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/tuning.md, quickstart.md

**Tests**: included — the spec's success criteria (SC-001, SC-003, SC-005) explicitly
demand test evidence (suite-unchanged, replay hash equality, one proof per dial),
and replay determinism is a hard invariant of this codebase.

**Organization**: grouped by user story. US1 (manifest mechanism) and US2 (event
discipline) are co-P1 and share the foundational TuningState; they are ordered
US1 → US2 because the event seed consumes the parser. US3 wires the five dials.
US4 is docs.

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

*No setup tasks — existing Go module, no new dependencies, no new packages.
Worktree creation is TASK-107 process (constitution II), not a spec task.*

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: the TuningState type, defaults, clamps, and accessors that every
story consumes.

- [x] T001 Create `internal/sim/tuning.go`: `TuningState` struct (five int fields,
      JSON tags per data-model.md), `default*` constants relocated/renamed from
      their current homes (`defaultRefuelDyingBelow`, `defaultFireBurnPerWood`,
      `defaultGruEmergePerMille`, `defaultPlannerCadenceTicks`,
      `defaultEncounterCooldownTicks` — values 3600, 14400, 600, 1800, 7200),
      per-field clamp bounds per contracts/tuning.md, and `defaultTuning()`
      returning the fully-resolved default set. Remove/alias the old consts in
      `internal/sim/agents.go` (lines ~671, ~705, ~709) and `internal/sim/gru.go`
      (line ~46); the mind-side `encounterCooldownTicks` const in
      `internal/mind/mind.go:36` is removed in T013.
- [x] T002 Add `Tuning *TuningState` field to `State` in `internal/sim/state.go`
      with tag `json:"tuning,omitempty"` (placed with the other omitempty
      late-additions, `state.go:92-106`), plus nil-safe accessor methods on
      `*State` in `internal/sim/tuning.go`: `RefuelDyingBelow()`,
      `FireBurnPerWood()`, `GruEmergePerMille()`, `PlannerCadence()`,
      `EncounterCooldown()` — tuned value when `Tuning != nil`, else default.
- [x] T003 [P] Add snapshot-compatibility test in `internal/sim/tuning_test.go`:
      a pre-048 snapshot JSON (no `tuning` key) unmarshals to `Tuning == nil`
      with all accessors returning defaults, and a state with `Tuning == nil`
      marshals with no `tuning` key (byte-identical guarantee, no
      format_version bump).

**Checkpoint**: TuningState exists, accessors compile, defaults proven equal to
the old constants.

---

## Phase 3: User Story 1 — Operator tunes a dial without editing code (P1) 🎯 MVP

**Goal**: `tuning.json` read at boot; absent file == current behavior; clamps
warn; malformed/unknown fails boot.

**Independent Test**: boot with no file → unchanged; boot with one dial set →
that value effective; out-of-range → clamped+warned; garbage → boot refused.

- [x] T004 [US1] Implement `ParseTuning(data []byte) (*TuningState, []string, error)`
      in `internal/sim/tuning.go`: strict decode
      (`json.Decoder.DisallowUnknownFields`) into a sparse pointer-fields
      carrier, resolve against defaults into a full set, clamp out-of-range
      values collecting `llm/config.go`-style warning strings
      (`tuning.json <field> <raw> out of range (max <bound>) — clamped to <v>`);
      unknown key / wrong type / malformed JSON → error.
- [x] T005 [P] [US1] Table-driven parse/clamp tests in
      `internal/sim/tuning_test.go`: empty object → default set, sparse file
      resolves missing fields to defaults, each field clamps at both bounds
      with the documented warning text, unknown key errors, wrong type errors,
      malformed JSON errors (contract table in contracts/tuning.md is the
      test table).
- [x] T006 [P] [US1] Add `TuningPath()` helper to `internal/world/world.go`
      (`filepath.Join(w.Dir, "tuning.json")`, the `CalibrationPath()` pattern
      at world.go:311).
- [x] T007 [US1] Daemon boot loading in `internal/daemon/daemon.go`: after
      `recoverState`/`seedMeetingConvention` (~line 116), read
      `w.TuningPath()`; absent → skip silently; present → `sim.ParseTuning`,
      print each warning line, fail boot on error with file path + problem.
      (Seeding the event is T009 — this task may land as one function with a
      TODO seam or split, but the load/validate/warn/fail behavior is complete
      here.)

**Checkpoint**: US1 acceptance scenarios 1, 3, 4 demonstrable (scenario 2's
"value takes effect" is fully true once US3 wires consumers; the effective set
is already computed and reported).

---

## Phase 4: User Story 2 — Replays reproduce tuned behavior (P1)

**Goal**: effective values live in the event log; replay never reads the file;
no redundant events on restart.

**Independent Test**: run tuned, delete tuning.json, replay log → identical
state; restart with unchanged file → no new event.

- [x] T008 [US2] Reducer arm for `sim.tuning_applied` in
      `internal/sim/state.go` `Apply` dispatch (near the governance arm,
      ~line 1694): decode full-set payload into `s.Tuning` (idempotent), plus
      `NewTuningEvent(state, tick, set)` constructor in
      `internal/sim/tuning.go` (the `NewConventionEvent` pattern,
      governance.go:628). Payload type `TuningAppliedPayload` beside the other
      payload structs in state.go.
- [x] T009 [US2] `seedTuning(w, st, state, parsed)` in
      `internal/daemon/daemon.go` (the `seedMeetingConvention` shape,
      daemon.go:486): compute effective set (parsed manifest, or default set
      when file absent — but absent file seeds NOTHING; only a present file
      participates), compare against in-effect set (`state.Tuning`, nil ≡
      defaults), and when different `state.Apply` + `st.AppendEvents` one
      `sim.tuning_applied` before the loop starts and before `mind.New`
      (~daemon.go:331). Equal sets append nothing.
- [x] T010 [P] [US2] Reducer/seed tests in `internal/sim/tuning_test.go`:
      apply `sim.tuning_applied` → accessors return payload values; re-apply
      idempotent; snapshot round-trip carries `Tuning`; a log with defaults →
      tuned → (no event) restart sequence yields exactly one tuning event.
- [x] T011 [US2] Replay determinism test (quickstart §4, SC-003) in
      `internal/sim/tuning_test.go` (or `internal/daemon` if it needs the
      store): drive a state under a tuned set past fire-refuel + gru-night +
      encounter activity via events, hash final state, replay same events into
      a fresh state, assert hash equality — no file involved anywhere in the
      replay path.
- [x] T012 [P] [US2] Old-log compatibility test: replaying a pre-048 event
      sequence (no tuning events) leaves `Tuning == nil` and defaults in
      effect (spec FR-007); secondary consumers survive the new event kind —
      extend the chronicle/digest "unknown event tolerance" coverage if such a
      test exists, else assert `Apply` on an event-kind-unaware consumer path
      doesn't panic (grep `internal/tui/digest.go`, `internal/sim/chronicle.go`
      for their default arms and add the minimal assertion).

**Checkpoint**: tuning is fully event-sourced; SC-003 evidence exists.

---

## Phase 5: User Story 3 — The five earned dials consume the manifest (P2)

**Goal**: every promoted call site reads the accessor, not a constant.

**Independent Test**: per-dial tests prove the tuned value drives behavior
(SC-005).

- [x] T013 [US3] Convert reducer-side call sites to accessors:
      `internal/sim/policy.go:147` (`f.Detail-tick < s.RefuelDyingBelow()`),
      `internal/sim/executor.go:832` and `internal/sim/state.go:915`
      (`FireBurnPerWood()`), `internal/sim/gru.go:235`
      (`GruEmergePerMille()`). Verify no other non-test reads of the old
      consts remain (`grep -rn "refuelDyingBelow\|fireBurnPerWood\|gruEmergePerMille" internal/ --include="*.go" | grep -v default | grep -v _test`).
- [x] T014 [US3] Convert mind-layer call sites to replica accessors:
      `internal/mind/mind.go:176,384,432` and
      `internal/mind/embedder.go:156,212` (`replica.PlannerCadence()`),
      `internal/mind/mind.go:317` (`md.replica.EncounterCooldown()`, removing
      the `encounterCooldownTicks` const at mind.go:36); update the
      `sim.PlannerCadenceTicks` read in `internal/daemon/daemon.go:341`
      (boot printout) to the state accessor. Keep `sim.AgentCount` stagger
      math intact (`replica.PlannerCadence()/sim.AgentCount`).
- [x] T015 [P] [US3] Per-dial proof tests (SC-005, quickstart §5): in
      `internal/sim/tuning_test.go` — tuned `RefuelDyingBelow` changes the
      refuel-reflex trigger window (the `food_fire_test.go:378` scenario under
      a non-default value); tuned `FireBurnPerWood` moves the `FuelUntil`
      deadline (the `food_fire_test.go:222` scenario); tuned
      `GruEmergePerMille` at 0 and 1000 flips `gruEmergence` (the
      `gru_test.go:43` roll under tuned state). In `internal/mind` — tuned
      `PlannerCadence` shifts `nextDue` stagger and embedder bucket edges;
      tuned `EncounterCooldown` gates/admits a re-encounter at the boundary.
- [x] T016 [US3] Fix test-suite fallout from const renames: update in-package
      test references (`internal/sim/recipes_test.go:82` names
      `fireBurnPerWood`; `internal/sim/gru_test.go:43` names
      `gruEmergePerMille`; `internal/sim/whole_feature_test.go:285-287`;
      `internal/mind` tests naming `sim.PlannerCadenceTicks` or
      `encounterCooldownTicks`) to the `default*` consts or accessors as
      semantically appropriate, keeping every existing assertion's meaning.
      `go test ./...` green (SC-001 suite-unchanged evidence).

**Checkpoint**: all five dials manifest-driven; full suite green.

---

## Phase 6: User Story 4 — Design report points at the mechanism (P3)

- [x] T017 [US4] Update §6 of `docs/design/control-surface-and-calibration.md`:
      promotion path steps 1–2 marked shipped via `tuning.json` +
      `sim.tuning_applied` (spec 048 / TASK-107), the five promoted dials
      listed with file/clamps pointer to
      `specs/048-tuning-manifest/contracts/tuning.md`, step 3 (hot exposure)
      still future. Also update the §2 inventory rows for the five dials to
      note "tuning.json-promoted".

---

## Phase 7: Polish & Cross-Cutting

- [x] T018 Manual smoke per quickstart.md §§1–3 on a scratch world (new world,
      no file → silent; add file → one event, warnings on clamp; garbage →
      boot refusal). Record transcript/evidence in the PR description.
- [x] T019 Run `node scripts/check-tui-design.mjs --changed` (spec 047 gate) —
      expected no-op since `internal/tui/` is untouched; if the digest/TUI was
      touched by T012, amend `docs/design/tui/` accordingly in the same PR.
- [x] T020 Post-merge (root, after PR lands): `/grounding-wiki:wiki-update` for
      notes sourcing `internal/sim/state.go`, `agents.go`, `gru.go`, `policy.go`,
      `executor.go`, `internal/mind/mind.go`, `internal/mind/embedder.go`,
      `internal/world/world.go`, `internal/daemon/daemon.go`; then player-docs
      freshness check (`node .claude/skills/player-docs/scripts/check-freshness.mjs --check`).

---

## Dependencies

```text
Phase 2 (T001→T002; T003 after T002)
  └─▶ US1: T004 (needs T001); T005 ∥ T006 ∥ (after T004) T007 (needs T004+T006)
        └─▶ US2: T008 (needs T002); T009 (needs T007+T008); T010/T011/T012 after T008
              └─▶ US3: T013 ∥ T014 (need T002); T015 after T013/T014; T016 last in phase
                    └─▶ US4: T017 (any time after scope is fixed; sequenced here)
                          └─▶ Polish: T018 (needs everything); T019; T020 post-merge
```

US1→US2 ordering is soft (the parser feeds the seed); US3 strictly needs
Phase 2 only, but lands after US2 so the per-dial tests can lean on the event
path. T020 runs at root after merge — it is the constitution-IV follow-through,
not a branch commit.

## Parallel Example

After T004 lands: T005, T006 in parallel. After T008: T010, T011, T012 in
parallel. T013 and T014 touch disjoint packages — parallel. T015 splits
per-dial across files if desired.

## Implementation Strategy

MVP = Phase 2 + US1 + US2 (the mechanism, fully event-sourced, zero dials
consuming it would still pass US1/US2 tests — but do not stop there; US3 is
the payoff). Single worktree `.worktrees/task-107`, branch
`task-107-tuning-manifest`, commits per phase, one PR. Implementer tier:
**Opus 4.8** (cross-package + doctrine-adjacent + mind scheduling — see
plan.md Constitution Check V).
