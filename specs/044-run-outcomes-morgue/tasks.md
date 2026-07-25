# Tasks: Run outcomes, the morgue file, death escalation, and graves

**Input**: Design documents from `/specs/044-run-outcomes-morgue/`

**Prerequisites**: plan.md, spec.md, research.md (R1–R13), data-model.md, contracts/

**Tests**: included — the spec's Independent Tests and Success Criteria (replay
byte-identity, determinism, catalog sweep) are explicit test obligations.

**Organization**: grouped by user story; each story is an independently testable
increment. All work lands on the single `task-31` branch in `.worktrees/task-31`
(one task, one PR).

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

- [x] T001 Create worktree `.worktrees/task-31` from fresh origin/main; confirm
      `go build ./... && go test ./...` green at base

## Phase 2: Foundational (blocking prerequisites)

- [x] T002 Add `State.Ended bool` + `State.RunEnd *RunEnd` (both `omitempty`) and the
      `RunEnd`/`DeathRecord` types to internal/sim/state.go; prove snapshot
      byte-compat for pre-feature worlds (marshal/unmarshal round-trip test in
      internal/sim/state_test.go or sim_test.go)
- [x] T003 [P] Add `livingCount(s *State) int` helper in internal/sim/state.go and use
      it to replace at least the death-adjacent ad-hoc `.Dead` loops it supersedes

**Checkpoint**: state carries the ended shape; nothing emits it yet.

## Phase 3: User Story 1 — The run ends, and the story survives it (P1) 🎯 MVP

**Goal**: last death → exactly one `run.ended`, sim halts permanently, everything stays
readable live/offline/after restart; postmortem posture in the TUI.

**Independent Test**: seeded world, force all deaths; verify run-end declaration, time
frozen, all reading surfaces work — including across a daemon restart (spec US1).

- [x] T004 [US1] `RunEndedPayload` struct + same-batch emission in `stepEvents` after
      the heartbeat death loop, guarded by `!s.Ended`, ordered after all same-tick
      `agent.died`; add top-of-`stepEvents` guard emitting nothing once ended —
      internal/sim/executor.go (research R1; contracts/events.md ordering)
- [x] T005 [US1] `run.ended` reducer arm setting `Ended`/`RunEnd` in
      internal/sim/state.go
- [x] T006 [US1] Ended branch in `Loop.Run` (no timer, modeled on the paused branch at
      loop.go:404-417) + `handleCommand` gating per contracts/status.md (refuse
      pause/resume/set_speed/govern/inject_intent; keep status/state/subscribe;
      inject_social accepts only recorded-prose types) — internal/sim/loop.go (R2)
- [x] T007 [US1] Surface ended fact: `Loop.status()` → additive `omitempty`
      `ClockStatus.Ended`/`EndedDay` — internal/sim/loop.go,
      internal/ipc/protocol.go, internal/ipc/server.go (R3)
- [x] T008 [P] [US1] `promptworld status` ended posture: human line + `--json` field on
      both live and offline-snapshot paths — cmd/promptworld/commands.go:504-549
- [x] T009 [US1] TUI postmortem posture, dual-source (replica `State.Ended` for
      snapshot-attach + pushed `run.ended`/status poll for live transition): `ENDED`
      header token replacing running/`PAUSED`, clock keys inert with footer hint —
      internal/tui/tui.go, internal/tui/views.go (R12)
- [x] T010 [P] [US1] `run.ended` digest row in internal/tui/digest.go + catalog row in
      docs/wiki/event-types.md (TestCatalogSweep enforces the pair)
- [x] T011 [P] [US1] `chronicleNote` line for `run.ended` in internal/mind/narrate.go
- [x] T012 [US1] Tests: same-tick double death → exactly one `run.ended` ordered last;
      ended world emits nothing on further ticks; replay rebuilds `Ended`
      (`TestReplayRebuildsState` extension); command refusal on ended loop; TUI header
      `ENDED` test per the badge-test pattern — internal/sim/sim_test.go,
      internal/sim/loop_test.go (or nearest), internal/tui/views_test.go

**Checkpoint**: US1 fully demonstrable per quickstart §2 — deployable MVP.

## Phase 4: User Story 2 — The morgue file (P2)

**Goal**: per-death factual epitaphs + charter/orders evidence in one regenerable
`morgue.md`; run-end summary; optional narrated epilogue that never touches facts.

**Independent Test**: no-AI world, cause a death, verify all seven factual fields
against history; add a narrator, verify separated epilogue and unchanged facts (spec US2).

- [x] T013 [P] [US2] `MorguePath()` helper beside the path cluster —
      internal/world/world.go:205-236
- [x] T014 [US2] Charter revision identity: content hash of the effective charter at
      turn load; emit `metatron.charter_observed{fingerprint, default}` when it
      differs from `State.CharterFingerprint` (first turn always emits); reducer arm +
      `State.CharterFingerprint (omitempty)` — internal/metatron/turn.go,
      internal/metatron/charter.go, internal/sim/state.go (R8)
- [x] T015 [US2] `morgue.epilogue` event: payload struct, `injectSocialWhitelist` entry
      (internal/sim/loop.go:193), bounded `State.MorgueEpilogues` ring + reducer arm
      (chronicle-ring pattern) — internal/sim/state.go, internal/sim/chronicle.go
      pattern (R9)
- [x] T016 [US2] `renderMorgue` in internal/scribe/scribe.go: whole-file render per
      contracts/morgue-document.md — epitaphs (replica relations/debts/retained
      memories + typed event scans for deeds and lifetime notable memories + charter
      timeline alignment + active orders), run-end summary from `State.RunEnd`,
      recorded epilogues blockquoted; render on batches containing
      agent.died/run.ended/morgue.epilogue and at boot (R6, R7)
- [x] T017 [US2] Narrator epilogue job: on absorbing `agent.died`/`run.ended`, enqueue
      an epilogue on the existing single-flight narrator worker (`llm.KindNarrator`,
      chronicle decision class), land prose via InjectSocial as `morgue.epilogue`;
      chronicle failure doctrine (gap, never stall) — internal/mind/narrate.go (R9)
- [x] T018 [P] [US2] Digest + catalog rows for `metatron.charter_observed` and
      `morgue.epilogue` — internal/tui/digest.go, docs/wiki/event-types.md
- [x] T019 [US2] Tests: golden-file morgue render on a scripted no-LLM history (all
      seven fields, SC-002); replay byte-identity of factual render (SC-004); evidence
      alignment (edited charter + active order named at death, SC-003); epilogue
      separation (facts byte-identical with narrator on/off); regeneration after file
      deletion — internal/scribe/scribe_test.go, internal/sim tests for new arms

**Checkpoint**: every death leaves a complete, replayable epitaph; run end closes the
document.

## Phase 5: User Story 3 — The gru can finish the wounded (P3)

**Goal**: survival floor conditional on the near-death band; gru kills flow through the
unchanged death path with cause `"gru"`.

**Independent Test**: staged night encounter — healthy villager survives with floor
intact; weakened villager dies, reproducibly under replay (spec US3).

- [ ] T020 [US3] Conditional floor at internal/sim/gru.go:131 — pre-attack
      `Needs.Health < nearDeathBelow` ⇒ floor 0, else `gruWoundFloor`; emit
      `agent.died{Cause: "gru"}` from `gruStep` after `gru.attacked` on a kill, with
      inline witness-death memory loop (executor.go:137-146 idiom, `salWitnessDeath`);
      update doctrine comment (gru.go:12-20), `GruAttackedPayload.Health` doc
      (gru.go:232), `DiedPayload` cause doc (agents.go:862-865) (R4, R5)
- [ ] T021 [P] [US3] `chronicleNote` branch for a killing gru attack (narrate.go:85-89
      "left them wounded" must not render for a death); verify digest `agent.died`
      alert covers cause "gru" — internal/mind/narrate.go, internal/tui/digest.go
- [ ] T022 [US3] Tests: amend `TestGruWoundsNotExecutes` (healthy half stays; the
      health-50 case inverts to a kill assertion); new `gruTestState` scenario —
      healthy + weakened victims, single `stepEvents`, assert died/survived + cause +
      witness memories + spill; static assert `gruWound >= nearDeathBelow`; replay
      determinism rides existing harnesses — internal/sim/gru_test.go (R13)

**Checkpoint**: death is reachable; a gru kill produces an epitaph (US2) and can end a
run (US1).

## Phase 6: User Story 4 — Graves on the map, grief in the village (P4)

**Goal**: persistent grave structure at each death site — rendered, perceivable,
tellable; grief rides the shipped memory→rumor chain.

**Independent Test**: witnessed death → grave on map, witness holds it as a known
place, grief talk within a game-day (spec US4).

- [ ] T023 [US4] Place `Structure{Kind: "grave"}` in the `agent.died` reducer arm
      (spill idiom, state.go:1426-1466); extend the structure-kind vocabulary comment
      — internal/sim/state.go, internal/sim/agents.go:237 (R10)
- [ ] T024 [P] [US4] Mirror the vocabulary: `PlaceFact.Kind` comment
      (internal/sim/mentalmap.go:49-51), `placeFactKinds`
      (internal/tool/registry.go:430), prompt landmark set
      (internal/mind/prompt.go:204) (R10)
- [ ] T025 [P] [US4] TUI: grave glyph case in `renderMapGrid` (views.go:418-472),
      legend entry (views.go:616), style var — internal/tui/views.go
- [ ] T026 [US4] Tests: grave placed at death tile, persists, replay-identical;
      perception sweep grants the place-fact (witnessed provenance) within one
      movement beat; place-telling spreads it; SC-006 integration — grief
      rumor/conversation referencing the death within one game-day (existing
      witness-death memory as seed, R11); `buildSite` blocked on the grave tile
      (documents the R10 tension) — internal/sim/mentalmap_test.go,
      internal/sim/social_test.go, internal/sim/gru_test.go or sim_test.go

**Checkpoint**: all four stories independently functional.

## Phase 7: Polish & Cross-Cutting

- [ ] T027 Full-suite gate: `go test ./...`, `gofmt`/`go vet`; confirm
      `TestDeterminismSameSeedSameTimeline` + `TestCatalogSweep` green with all new
      types cataloged
- [ ] T028 Run quickstart.md live walkthrough (§2–§6) on a demo world; record outcomes
      in the PR description
- [ ] T029 [P] Reconcile spec-adjacent docs touched by the work (event-types rows
      already landed per-story); note the post-merge `/grounding-wiki:wiki-update`
      obligation in the PR body (constitution Principle IV)

## Dependencies & Execution Order

- **Setup → Foundational → US1 → US2 → US3 → US4 → Polish** is the safe sequential
  order and matches spec priorities.
- **US1** depends only on Foundational.
- **US2** epitaphs (T013–T016 partial, T019) are independent of US1; the run-end
  summary section of T016 consumes US1's `RunEnd`. Implement after US1 as planned.
- **US3** is code-independent of US1/US2 (touches gru.go + tests) but is sequenced
  after them per the spec ("a death that occurs lands in a defined run/record
  structure"). T020 can start in parallel with US2 if desired — different files.
- **US4** is independent of US2/US3; T023 touches state.go's `agent.died` arm, which
  US2's T015/T016 read — coordinate the reducer-arm edits (T023 after T015 lands, or
  same author).

## Parallel Opportunities

- Within US1: T008, T010, T011 in parallel after T004–T007.
- Across stories: T013 (world.go), T020 (gru.go), T024/T025 (tool/tui) touch disjoint
  files and can proceed in parallel once Foundational lands.
- Test tasks (T012, T019, T022, T026) parallelize per package.

## Implementation Strategy

**MVP = Phase 1–3 (US1)**: run end + halt + readable archive + postmortem posture —
demonstrable and shippable alone. Then US2 (the morgue makes deaths legible), US3
(makes deaths reachable), US4 (makes deaths remembered). Each checkpoint is a
deploy/demo point; stop-and-validate between stories per quickstart.

**Tier notes (constitution V)**: US1 (loop/executor halt semantics) and US2's
T014/T015/T017 (turn pipeline, whitelist, narrator) meet the Opus 4.8 rubric
(concurrency/doctrine-adjacent); US3, US4, and the remaining US2 rendering/tests are
Sonnet-eligible. Recorded on TASK-31 at dispatch.
