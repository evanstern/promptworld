# Tasks: Live Cognition-Horizon Surface

**Input**: Design documents from `specs/037-live-horizon-surface/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md,
contracts/status-horizon.md, quickstart.md

**Tests**: included — the repo convention is table-driven tests alongside code
(testing-strategy wiki note), and SC-002/SC-004 demand pinned invariants.

**Organization**: grouped by user story; US1 is the MVP increment.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: parallelizable (different files, no dependency on an incomplete task)
- **[Story]**: US1 (live verdict), US2 (counters), US3 (CLI parity)

## Phase 1: Foundational (blocking prerequisite)

**Purpose**: the one shared arithmetic base both wire composition and every
render surface delegate to (research R1, spec FR-002).

- [ ] T001 Add `ClassStanding{Class, Suppressed, Verdict}` and
      `LiveHorizon(ticksPerSecond, secPerPtFor) []ClassStanding` to
      `internal/cognition/horizon.go`; re-base `SuppressedAt` as a
      suppressed-names filter over `LiveHorizon` (one watched-class iteration
      total). Extend `internal/cognition/horizon_test.go`: table tests across
      the speed ladder, exclusion (`ok=false` omits the class), uncapped
      (`tps ≤ 0` → all included classes suppressed with `Route`'s uncapped
      phrasing), and the invariant `SuppressedAt ≡ suppressed names of
      LiveHorizon` (SC-002; data-model.md invariant).

**Checkpoint**: `go test ./internal/cognition/` green; existing spec-035
callers (`set_speed` warning, boot warning) byte-identical.

---

## Phase 2: User Story 1 — Live per-class verdict at the current speed (P1) 🎯 MVP

**Goal**: a player at 32x sees WHICH classes are suppressed and WHY, in the
header (compact) and the metatron dock pane (detail), within one status poll.

**Independent Test**: quickstart.md §2 — hot world at 32x shows the badge +
per-class block; 1x clears them within one poll; governed world reads the
EFFECTIVE speed; no-LLM world byte-identical.

- [ ] T002 [US1] Add `HorizonClass{Class, Suppressed, Verdict, Calibrated,
      SuppressedCount}` (json per contracts/status-horizon.md) and
      `StatusData.Horizon []HorizonClass` with `json:"horizon,omitempty"` to
      `internal/ipc/protocol.go`, doc comment citing the contract's presence
      rules (LLM worlds only; never empty-slice).
- [ ] T003 [US1] Compose the horizon in `statusDataFull`
      (`internal/ipc/server.go`): when `s.llm != nil`, call
      `cognition.LiveHorizon(cs.Speed.TicksPerSecond(), resolver)` where the
      resolver maps class → `s.llm.EstimateForKind(llm.Kind(class))`
      (`ok=false` excludes); fill `Calibrated` from
      `s.llm.CalibratedAt(name) != ""`; `SuppressedCount` stays 0 until US2.
      Tests in `internal/ipc`: composition happy path, provider-missing
      exclusion, uncapped speed, calibrated flag, and the no-LLM
      byte-identity test (reply JSON has no `horizon` key — SC-004).
- [ ] T004 [US1] Header badge in `headerView`
      (`internal/tui/views.go`): when ≥1 `m.status.Horizon` entry is
      suppressed, append a warn-styled `[suppressed: <classes>]` badge on the
      `[llm: …]` badge pattern; absent otherwise. Tests in
      `internal/tui/views_test.go`: badge present/absent, both layouts,
      narrow fallback, class-name ordering follows wire order.
- [ ] T005 [US1] Metatron-pane horizon block (`horizonLines` beside
      `llmProviderLines` in `internal/tui/views.go`): one row per entry —
      class, plain-language standing at the current speed ("thinking at 8x" /
      "suppressed at 32x"), remedy phrase from `suppressed × calibrated`
      (uncalibrated → "calibrate or slow down", calibrated → "slow down"),
      via a phrase helper in the `verdictGlossary` posture (no raw enum
      strings; the wire `verdict` arithmetic may render verbatim as detail).
      Tests: rows for mixed standings, remedy split, absent block for no-LLM.

**Checkpoint**: US1 fully demo-able — quickstart §2 passes end-to-end.

---

## Phase 3: User Story 2 — Suppression counters (P2)

**Goal**: per-class "skipped N" counts, daemon-lifetime, identical for every
client.

**Independent Test**: quickstart.md §3 — counts grow while hot, freeze (not
reset) at 1x, match the world's suppressed `cog.outcome` events, reset only on
daemon restart.

- [ ] T006 [P] [US2] Add mutex-guarded per-class suppression counts to
      `llm.Orchestrator` (`internal/llm/llm.go`): `RecordSuppression(class
      string)` (O(1) bump) and `SuppressionCounts() map[string]int64`
      (defensive copy). Tests in `internal/llm/llm_test.go`: increment,
      copy-isolation, concurrent bumps under `-race`.
- [ ] T007 [P] [US2] Add the optional `suppressionCounting` seam in
      `internal/mind/telemetry.go` (the `estimating` pattern) and call it
      from `emitSuppressed` before the detached event emit — a fake
      orchestrator without the seam is a silent no-op. Test with a counting
      fake: every `emitSuppressed` class reaches the seam; absorb path never
      blocks/panics with a seamless fake.
- [ ] T008 [US2] Fold counts into the status composition
      (`internal/ipc/server.go`): `SuppressionCounts()` read once per status,
      `SuppressedCount` per entry keyed by class. Extend the ipc composition
      test: entries carry counts; unknown-class counts (unwatched, e.g.
      `chronicle`) don't leak extra entries.
- [ ] T009 [US2] Render `skipped N` on each metatron-pane horizon row
      (`internal/tui/views.go`), omitted when the count is 0 and the class is
      thinking. Extend the pane render tests.

**Checkpoint**: US1 + US2 — quickstart §3 passes; counts visible in TUI and
raw status JSON.

---

## Phase 4: User Story 3 — Headless status parity (P3)

**Goal**: `promptworld status <world>` prints the same live horizon.

**Independent Test**: quickstart.md §4 — LLM world shows the section; no-LLM
world output unchanged.

- [ ] T010 [US3] Render a horizon section in `renderStatusHuman`
      (`cmd/promptworld/commands.go`): one line per entry — standing at
      current speed, remedy when suppressed, `skipped N` — nothing when
      `Horizon` is absent. Tests in `cmd/promptworld/commands_test.go`:
      suppressed/thinking/mixed renders, no-LLM output byte-identical.

**Checkpoint**: all three stories independently verifiable.

---

## Phase 5: Polish & lifecycle gates

- [ ] T011 Full sweep: `go build ./... && go vet ./... && go test ./...`
      green (includes `-race` for the new counter paths where the suite
      already runs it); execute quickstart.md scenarios §2–§4 against a live
      world and record outcomes on the board task.
- [ ] T012 Grounding freshness (constitution IV): after merge, run
      `/grounding-wiki:wiki-update` to re-pin the notes whose sources changed
      (`cognition`, `tui-client`, `ipc-protocol`, `ipc-server`,
      `llm-orchestrator`, `agent-mind`, `cli-promptworld`), then
      `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`
      and refresh `docs/player/` if stale.

---

## Dependencies & Execution Order

- **Phase 1 (T001)** blocks T003 (composition calls `LiveHorizon`).
- **US1**: T002 → T003 → T004 → T005 (T004/T005 share `views.go` — sequential).
- **US2**: T006 ∥ T007 (different packages), then T008 (needs T006 + T002/T003
  shapes), then T009 (shares `views.go` with US1 tasks — after T005).
- **US3**: T010 after T002 (wire shape); independent of TUI tasks.
- **Polish**: T011 after all implementation; T012 after merge.

### Parallel opportunities

- T006 and T007 ([P]) — `internal/llm` vs `internal/mind`.
- T010 can proceed in parallel with T004/T005/T009 (different files) once
  T003 lands.

## Implementation Strategy

MVP = Phase 1 + US1 (T001–T005): the "silent world" defect is fixed and
demo-able. US2 adds weight (counters), US3 adds parity — each independently
testable per its quickstart section. One branch, one PR (TASK-41); commit per
task or logical group.
