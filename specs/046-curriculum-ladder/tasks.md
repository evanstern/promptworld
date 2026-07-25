# Tasks: The curriculum ladder

**Input**: Design documents from `/specs/046-curriculum-ladder/`

**Prerequisites**: plan.md, spec.md (client-reviewed), research.md (R1–R12),
data-model.md, contracts/ (4)

**Tests**: included — gating coherence, cross-stage determinism, catalog sweep, and
the fixture-driven unlock chain are explicit obligations.

**Organization**: grouped by user story; one `task-68` branch in `.worktrees/task-68`
(one task, one PR).

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

- [ ] T001 Create worktree `.worktrees/task-68` from fresh origin/main; confirm
      `go build ./... && go test ./...` green at base

## Phase 2: Foundational (blocking prerequisites)

- [ ] T002 Manifest fields `Stage`/`StageOverridden`/`CharterPreset` (additive
      omitempty) + closed-vocabulary validation at `world.Open` (MemoryRelevance
      precedent); absent = ungated — internal/world/world.go + tests
- [ ] T003 [P] `internal/skin` stub: `stageIdentity` table with the client-approved
      guardian names (The Voice / The Written Word / The Craft / The Stewardship) +
      one-line identity descriptions — internal/skin/skin.go
- [ ] T004 [P] `curriculum.*` event types: `exercise_passed`/`stage_unlocked` payload
      structs + reducer arms (bounded state records; unknown-type-safe) —
      internal/sim/curriculum.go + state wiring, per contracts/events.md

**Checkpoint**: stage facts exist; nothing gates or announces yet.

## Phase 3: User Story 2 — The world grants what the stage teaches (P1) 🎯 co-MVP

**Goal**: stage ceiling intersection + stage-1 instruction lock; mechanics untouched.

**Independent Test**: one world per stage — roster/instruction behavior matches the
ladder exactly; stage-1 edits get the honest notice; same-seed histories identical
across stages.

- [ ] T005 [US2] Stage ceiling table + `intersect(grant, stageCeiling)` at both
      manifest load sites (turn.go:145-152 before grantedRoster; status twin
      turn.go:679), bundle-narrowing idiom; pin exact stage-1 tool names against the
      live roster and record them in contracts/stage-gating.md —
      internal/metatron/charter.go, turn.go (R2)
- [ ] T006 [US2] Stage-1 instruction lock: preset-sourced effective charter
      (preset-aware loadCharter/charterIsDefault), skills not composed, does-not-bind
      notice via the existing notice channel; Status gains stage/lock provenance —
      internal/metatron/charter.go, turn.go (R3)
- [ ] T007 [US2] Tests: table-driven post-intersection roster per stage == ladder;
      door refusal beyond stage; declaration/prose/door coherence; stage-1 lock +
      notice (edited charter + skill files present); **cross-stage determinism diff**
      (same seed/commands at stage-1 vs stage-4 ⇒ identical world-event history) —
      internal/metatron tests + internal/sim determinism extension (FR-006)

**Checkpoint**: gate-to-feature pathway real; co-MVP with US1.

## Phase 4: User Story 1 — Choosing a stage is choosing an identity (P1) 🎯 co-MVP

**Goal**: informed creation UX; stage recorded and visible.

**Independent Test**: create at each stage (earned/overridden); identity presentation,
informed override, durable stage fact, status visibility.

- [ ] T008 [US1] `promptworld stages` command: identity table (skin stub names),
      concept, grants, unlock evidence, earned state (unlocks record read) —
      cmd/promptworld/commands.go (R9)
- [ ] T009 [US1] `promptworld new --stage <id>` (+ `--override`): earned-stage checks,
      informed error naming skipped concepts, manifest stamping, default stage
      selection (stage-1 for new players, else highest earned); `--charter-preset`
      opt-out — cmd/promptworld/commands.go (R9)
- [ ] T010 [P] [US1] Status surfaces: `WorldStatus.Stage`/`StageOverridden` (additive
      omitempty), CLI posture-style stage line (human + --json, live + offline), TUI
      metatron pane stage line — internal/ipc/protocol.go, internal/ipc/server.go,
      cmd/promptworld/commands.go, internal/tui/{tui.go,views.go} (R10)
- [ ] T011 [US1] Tests: creation flows (earned/unearned/override/default), manifest
      immutability (no mutation path exists), status rendering, absent-stage worlds
      unchanged — cmd + ipc + tui tests

**Checkpoint**: the ladder's front door works end to end.

## Phase 5: User Story 3 — Earning the next stage, told in-game (P2)

**Goal**: pass → unlock → chronicle/status → per-user record, fixture-proven.

**Independent Test**: fixture pass event drives the full chain; negative case: default-
charter pass unlocks nothing.

- [ ] T012 [US3] Unlock derivation: gate-conjunct evaluation over a recorded pass
      (stage-2 conjunct: player-authored `metatron.charter_observed` evidence — stub
      the type if 044 US2 hasn't merged; reconcile on rebase), `stage_unlocked`
      emission exactly once per (world, stage) — internal/sim/curriculum.go per
      contracts/events.md + contracts/unlocks-record.md gate conjuncts
- [ ] T013 [US3] Per-user unlocks record: read/heal/write (`.tmp`+rename, advisory
      doctrine), upsert on observing `stage_unlocked`, evidence pointers —
      internal/worlds/unlocks.go + daemon observer wiring (R4)
- [ ] T014 [P] [US3] Announcement wiring: `chronicleNote` case for `stage_unlocked`;
      `familyByNamespace["curriculum"]`; digestRegistry + fixture + event-types.md
      rows (TestCatalogSweep) — internal/mind/narrate.go,
      internal/tui/{grammar.go,digest.go}, docs/wiki/event-types.md
- [ ] T015 [US3] Tests: fixture chain (pass → unlock → chronicle line → record entry
      → stages/new honor it); SC-004 negative (default-charter evidence ⇒ no unlock);
      corrupt/missing record tolerance; once-only unlock; catalog sweep —
      internal/sim, internal/worlds, internal/tui tests

**Checkpoint**: the ladder moves (on fixtures until TASK-119 emits for real).

## Phase 6: User Story 4 — Two exercises exist and teach (P2)

**Goal**: first-night + the-law as parseable, 119-consumable content.

**Independent Test**: definitions parse; rubric terms are cataloged event types;
stage-2 rubric requires the charter conjunct.

- [ ] T016 [US4] Exercise definition structs + the two shipped definitions + reserved
      additive Manifest block shape (Meeting precedent; consumed by TASK-119 later)
      per contracts/exercises.md — internal/sim/curriculum.go (or internal/world),
      with parse/validation tests proving every rubric term is a cataloged event type

**Checkpoint**: exercises are real content with a real consumer contract.

## Phase 7: User Story 5 — The stage has a floor and a guide (P3)

**Goal**: tutor preset + per-stage quickstarts.

**Independent Test**: tutor-preset world orients unprompted (LLM world); freshness
gate green over 13 pages.

- [ ] T017 [US5] `persona.TutorCharter` const + `Genesis` preset parameter + `new`
      wiring (stage-1 default, opt-out) — internal/persona/{charter.go,files.go},
      cmd/promptworld/commands.go (R6)
- [ ] T018 [P] [US5] Four per-stage quickstart pages via the player-docs skill:
      author pages, EXPECTED_PAGES + SKILL.md mapping table + index.html nav +
      "nine pages" description string updates; freshness gate green —
      docs/player/, .claude/skills/player-docs/ (R8 of research; SC-006)
- [ ] T019 [US5] Tests/validation: tutor preset seeds and hot-reloads like any
      charter; no-model world retains all gating/preset/unlock function (FR-014);
      freshness script exit 0 — internal/persona tests + script run

## Phase 8: Polish & Cross-Cutting

- [ ] T020 Full gate: `go test ./...`, gofmt/vet, TestCatalogSweep, determinism
      harnesses unmodified
- [ ] T021 Run quickstart.md walkthrough (§2–§6) live; record outcomes in the PR
      description; note post-merge wiki-update obligation (Principle IV)
- [ ] T022 [P] Reconcile sequencing notes: if 044 US2 merged during this work, drop
      the charter_observed stub and bind the real event; record TASK-119/121 seam
      handoffs on the board task

## Dependencies & Execution Order

- Setup → Foundational → US2 → US1 → US3 → US4 → US5 → Polish. US2 before US1 so
  creation UX demos real gating. US3 depends on Foundational events + US1's record
  read path. US4/US5 independent of US3 (parallel-eligible).
- Cross-branch: 044 US2 (charter_observed) in flight — T012 stubs, T022 reconciles.

## Implementation Strategy

**MVP = US2 + US1** (staged worlds with real gating and an informed front door).
Tier notes (constitution V): **Opus 4.8** for the gating slice (T005–T007 — capability
gating in the turn pipeline is injection-adjacent doctrine); **Sonnet** for T002–T004,
US1, US3–US5, Polish (routine additive surfaces). Recorded on TASK-68 at dispatch.
