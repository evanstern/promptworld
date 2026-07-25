# Tasks: First-occurrence lessons projection (lesson row)

**Input**: Design documents from `/specs/055-lesson-row/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: included — the spec's success criteria (SC-001…SC-005) name mechanical
verification explicitly, and repo convention is tests alongside code.

**Organization**: foundation → US1 (P1, the MVP) → US2 (P2) → US3 (P3) → polish.

## Phase 1: Foundational (blocking prerequisites)

- [ ] T001 [P] Per-user seen-state record in `internal/worlds/lessons.go`:
      `LessonsSeenPath()`, `LoadLessonsSeen()`, `MarkLessonSeen(id, world string)` —
      mirror `unlocks.go` (load-tolerant, advisory, temp-file+rename atomic write);
      schema per `contracts/seen-state-file.md`.
- [ ] T002 [P] Tests in `internal/worlds/lessons_test.go`: missing/corrupt/
      unknown-version file loads empty; read-only dir swallows write failure; upsert
      preserves existing entries; version + unknown-field tolerance.
- [ ] T003 Catalog skeleton in `internal/tui/lessons.go`: `lessonEntry` type
      (data-model.md), `lessonCatalog` with all 8 entries (ids, tiers, skin-tokened
      text/pointer/title/body prose per `contracts/lessons-catalog.md`), and the
      `lessonSkinResolve` seam per research.md R1 (delegate to spec 052 runtime if
      merged; else default-table fallback for the tokens the catalog uses).

**Checkpoint**: `go test ./internal/worlds/` green; catalog compiles; no TUI wiring yet.

## Phase 2: User Story 1 — first encounter teaches itself, exactly once (P1) 🎯 MVP

**Goal**: mechanics-tier lessons surface once-ever in the new row; dwell/dismiss/queue
/decay; seen-state persists.

**Independent test**: quickstart.md §2 steps 1, 3–5 (suppression lesson end-to-end).

- [ ] T004 [US1] Trigger projection in `internal/tui/lessons.go`: `lessonTriggers`
      with `ingest(e store.Event)` — predicates for the 5 mechanics triggers
      (research.md R3 table: `cog.outcome` suppressed, `gru.attacked`,
      `metatron.charge_regenerated`, `metatron.order_expired`, `agent.died`), seen-set
      check, and the row state machine (one-active, FIFO queue, `lessonQueueDecay`,
      `lessonSpacing`, mark-seen-on-surface) per data-model.md transitions.
- [ ] T005 [US1] Wire into the client in `internal/tui/tui.go`: load `LessonsSeen` at
      init; call `lessonTriggers.ingest` at the same event-arrival seam as
      `decisionTraces.ingest` (decisions.go:150 precedent); persist via
      `MarkLessonSeen` when a lesson surfaces.
- [ ] T006 [US1] `x` key in `internal/tui/tui.go` global dispatch: dismiss the active
      lesson (record seen, start spacing); strict no-op when none active (documented
      fallthrough, keymap doctrine).
- [ ] T007 [US1] Render the row in `internal/tui/views.go` + chrome budget in
      `internal/tui/layout.go`: two lines, no border, above the guardian strip; line 2
      = pointer + `(? for more · x dismiss)` suffix appended by the renderer; row
      absent when no active lesson (control-table `none` state); all strings through
      `lessonSkinResolve`.
- [ ] T008 [US1] Done-signal clearing in `internal/tui/lessons.go`: catalog `done`
      predicates (v1: `first-order-expired` clears on `metatron.order_placed`);
      clearing records seen + starts spacing.
- [ ] T009 [US1] Tests in `internal/tui/lessons_test.go`: SC-001 fixture sweep (each
      mechanics trigger exactly once across simulated two-world + restart with a real
      temp seen-file); queue order, decay drop, spacing; `x` dismiss; done-signal
      clear; corrupt/missing seen-file boots clean (SC-004); no `{{` in rendered
      output (SC-005, default skin).

**Checkpoint**: US1 fully functional — MVP shippable.

## Phase 3: User Story 2 — the player's own prompting practice (P2)

**Goal**: the 3 prompting-tier lessons ride the same machinery.

**Independent test**: fixture-fire each prompting event for a fresh user; exactly once
each.

- [ ] T010 [P] [US2] Prompting-tier predicates + prose in `internal/tui/lessons.go`:
      `cog.tool_call` verdict ≠ landed, `metatron.charter_observed{default: false}`,
      `metatron.order_placed{fuzzy: true}` (payload-field vocabulary per
      decisions.go:180 / digest.go:962,997).
- [ ] T011 [P] [US2] Tests in `internal/tui/lessons_test.go`: exactly-once sweep for
      the 3 prompting triggers; fuzzy/default payload-field discrimination (a
      non-fuzzy order fires nothing; a default charter fires nothing).

**Checkpoint**: all 8 catalog entries live.

## Phase 4: User Story 3 — always findable again; the row knows its place (P3)

**Goal**: pull half via `?`; stage defaults; fold + narrow behavior.

**Independent test**: quickstart.md §2 step 2 + stage/fold render assertions.

- [ ] T012 [US3] Populate `helpLessons` from `lessonCatalog` at client init (boot-time
      skin resolution) in `internal/tui/lessons.go`/`tui.go`; test asserts id-for-id
      catalog↔overlay equality and that the placeholder line no longer renders
      (SC-002).
- [ ] T013 [US3] Stage defaults in `internal/tui/views.go`/`layout.go`:
      `lessonRowDefault(stage string) bool` — row on at stages 1–2; `[lesson]` header
      badge + overlay-only at stage 3+ and pre-ladder (`""` stage) — reading the
      existing status stage (tui.go:115 / ipc protocol Stage), per research.md R6.
- [ ] T014 [US3] Fold + narrow in `internal/tui/layout.go`/`views.go`: under height
      pressure the row folds to the `[lesson]` badge BEFORE the guardian strip
      (layout.md ruling (a) restricted to built foldables, research.md R5); narrow
      (< 112 cols) carries the row with identical stage defaults (ruling (b)).
- [ ] T015 [US3] Render tests in `internal/tui/lessons_test.go` (or
      `render_test.go` per its conventions): badge at stage 3+/pre-ladder; fold order
      vs guardian strip; narrow carry; row never exceeds 2 rows at any width/stage
      (SC-003).

**Checkpoint**: all three stories complete.

## Phase 5: Polish & gates (same PR)

- [ ] T016 [P] Amend `docs/design/tui/panels/lesson-row.md`: `status: specified →
      shipped`, fill real renderer symbols in the control table, add sources, re-pin
      `verified_against`.
- [ ] T017 [P] Amend `docs/design/tui/patterns/keymap.md`: move `x` from "New global
      keys (specified, unbuilt)" into the global-mode table; re-pin.
- [ ] T018 Run the gates from the worktree and fix what they surface:
      `go test -race ./...`, `node scripts/check-tui-design.mjs --changed`,
      `node scripts/check-merge-drift.mjs pr`; then the quickstart.md §2 manual smoke.

## Dependencies

- T001/T002 ∥ (same package, one author order: T001 → T002); T003 depends on nothing
  else and may run ∥ with T001.
- US1 (T004–T009) needs T001 + T003. Within US1: T004 → T005 → (T006 ∥ T007 ∥ T008) →
  T009.
- US2 (T010–T011) needs T004 (the projection); independent of US3.
- US3 (T012–T015) needs T003 (T012) and T007 (T013/T014); independent of US2.
- Polish needs everything (T016/T017 ∥; T018 last).

## Implementation strategy

MVP = Phase 1 + US1 (T001–T009): a shippable teaching row with the 5 mechanics
lessons. US2 is catalog additions; US3 completes the design-system obligations. All
phases land as commits on `task-117-lesson-row` and merge in TASK-117's single PR.
