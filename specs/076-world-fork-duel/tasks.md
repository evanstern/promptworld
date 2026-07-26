# Tasks: World fork + duel v1 — `promptworld fork` and the rubric-first scoreboard

**Input**: Design documents from `/specs/076-world-fork-duel/`
**Prerequisites**: spec.md, plan.md, research.md, data-model.md

**Tests**: included alongside code. Tier: **Opus 4.8** per the board record
(cross-package architectural: world lifecycle + store + sim vocabulary + TUI exports +
two CLI verbs; determinism-doctrine-adjacent).

**Organization**: phases map to the board ACs — Phases 2–3 ↔ AC #1/#2/#4/#5 (fork verb,
lineage, determinism, budget), Phase 4 ↔ AC #7 (rubric-first scoreboard), Phase 5 ↔
AC #3 (divergence + interleaved chronicles). AC #6 (spec linked before implementation) is
Phase 1's link step. Phases 6–8 are the design/wiki gates and close-out.

## Phase 1: Setup

- [X] T001 Link this spec to the board BEFORE implementation (AC #6): `spec-bridge:link`
      for `specs/076-world-fork-duel/` ↔ TASK-67 from the repo root; verify the card
      carries the Spec marker (orchestrator step — recorded here so the gate is a task,
      not an assumption)
- [X] T002 Cut the task worktree from fresh origin/main: `git fetch origin && git pull
      --ff-only` at root, `node scripts/check-merge-drift.mjs worktree --spec 076 --task
      TASK-67`, then `git worktree add .worktrees/task-67 -b task-67-world-fork-duel
      origin/main`; push the branch on first commit (claim protocol); baseline
      `go test ./internal/sim/ ./internal/world/ ./internal/store/ ./internal/tui/
      ./internal/worlds/` green before changes (worktree `.worktrees/task-67`)

## Phase 2: Foundational — lineage vocabulary + store helper (blocks Phase 3)

- [X] T003 `WorldForkedPayload` struct (data-model §1) in the shared payload block +
      `case "world.forked":` recorded-history no-op arm beside `world.created`, in
      `internal/sim/state.go`; reducer test: applying the event mutates nothing
      (marshal-identical state before/after), in `internal/sim/sim_test.go` or a new
      `internal/sim/fork_event_test.go` (plan D1, spec FR-007)
- [X] T004 [P] `world.forked` digest registry entry + `catalogFixture` row ("forked from
      `<parent>` at day D, HH:MM"), in `internal/tui/digest.go` +
      `internal/tui/digest_test.go` — `TestCatalogSweep` totality holds with the new
      event type (plan D1, spec FR-009)
- [X] T005 [P] `LineageConfig` + `Manifest.Lineage *LineageConfig json:"lineage,omitempty"`
      + `Open` structural validation (present block: non-empty `parent`, `fork_tick >= 0`);
      tests: lineage-less `world.json` round-trips byte-identically, validation table, in
      `internal/world/world.go` + `internal/world/world_test.go` (plan D2, spec FR-008)
- [X] T006 [P] `Store.MetaByPrefix(prefix) (map[string]string, error)` + test, in
      `internal/store/store.go` + `internal/store/store_test.go` (plan D3, spec FR-012)

## Phase 3: User Story 1 — the fork verb; both run side by side (P1, board ACs #1/#2/#4/#5)

**Goal**: `promptworld fork` produces a runnable, self-contained, provenance-carrying,
determinism-proven copy at the latest snapshot, inheriting the parent's wallet.

**Independent Test**: e2e — new → run past a snapshot → stop → fork → start both → both
answer status, `ps` shows both; unit — replay-to-hash identities.

- [X] T007 [US1] `world.Fork(srcDir, destDir, newName) (*ForkResult, error)` ceremony in
      new `internal/world/fork.go` (plan D3 steps 1–7: Open + live-daemon refusal +
      LatestValidSnapshot boundary (nil → refuse with remedy) + empty-dest check with
      best-effort cleanup on error + fresh log prefix stream + boundary snapshot +
      `world.forked` append + meta seed/format stamp + `llm_spend_*` copy + fork manifest
      (name/created_at/lineage new, ALL else verbatim, seed carried) + R9 sidecar
      copy/skip + `ForkResult` per data-model §4) (spec FR-002..005, FR-007/008,
      FR-012)
- [X] T008 [US1] Ceremony tests in new `internal/world/fork_test.go`: happy path
      (contiguity 1..N+1, boundary snapshot hash verifies, `world.forked` payload exact,
      manifest field-by-field carry, scribe views/runtime/archives NOT copied); refusal
      table (no snapshot / non-empty dest / live pidfile / bad name); partial-failure
      cleanup (plan D3, spec US1 scenarios 1/4, edge cases)
- [X] T009 [US1] Determinism proofs (board AC #4, FR-010 a+b) in
      `internal/world/fork_test.go`: genesis replay of the fork's log through
      `sim.NewState` + `Apply` marshal-hashes equal to the boundary snapshot's
      `state_hash` AND to the parent's state hash at the same (tick, seq) — the
      byte-identity property (spec US1 scenario 3, SC-003)
- [X] T010 [US1] Wallet inheritance proof (board AC #5, SC-006): parent store with
      seeded `llm_spend_<month>` total + per-provider keys → `Fork` → `llm.NewMeter`
      over the fork's store reports the same spent/attribution, in
      `internal/world/fork_test.go` (spec FR-012/013)
- [X] T011 [US1] CLI `fork` subcommand: new `cmd/promptworld/fork.go` + dispatch/usage in
      `cmd/promptworld/main.go` — resolveWorld source, `new`-convention name/path dest,
      `--at latest-snapshot` sole accepted value (default; other values refused naming
      the follow-on), summary print (boundary day/HH:MM + tick, events carried,
      truncated tail, lineage, ended-boundary warning, spend line, start-both
      next-steps); CLI tests in `cmd/promptworld/fork_test.go` (plan D4, spec FR-001/002,
      US1 scenario 5)
- [X] T012 [US1] e2e in new `e2e/fork_e2e_test.go` (SC-001 + FR-010 c): create pure-sim
      world, run at max past a snapshot, stop (cuts final snapshot), fork, start BOTH,
      both answer `status` and appear in `ps --json` as running, stop both; then replay
      the FORK's full log from genesis and assert its hash matches its own final
      snapshot's `state_hash` — the fork passes the determinism harness independently
      (plan D4, spec US1 scenarios 1–3)

## Phase 4: User Story 2 — the duel scoreboard (P2, board AC #7)

**Goal**: `promptworld compare` renders one honest rubric card per world through the ONE
spec-072 resolver — plain language, truthful glyphs, postmortem register on a loss.

**Independent Test**: decided-duel fixture renders winner all-✓ (from pass), loser ✗ with
`agent.died: N`; no raw enum in the output; duel rows == postmortem rows.

- [X] T013 [US2] Export the spec-072 resolver replica-parametric (data-model §5):
      `ResolveRubricFacts(state, def, pass)`, `RecordedPassFor(state, exercise)`,
      `ReportCardFact`/`ReportCardMode` aliases + exported mode consts,
      `RenderReportCard` over `reportCardView`; `Model.resolveReportCardFacts`/
      `Model.recordedPassFor` become thin wrappers, in `internal/tui/reportcard.go`
      (+ shim in `internal/tui/views.go` if the renderer wrapper lives there); existing
      spec-072 TUI tests pass unchanged — the no-behavior-change proof (plan D5, spec
      FR-018/019)
- [X] T014 [US2] `worlds.OfflineState(w) (*sim.State, int64, error)` extracted from
      `OfflineSnapshot` (which now calls it); existing probe/ps tests pass unchanged, in
      `internal/worlds/probe.go` (plan D6, spec FR-015)
- [X] T015 [US2] CLI `compare` scoreboard: new `cmd/promptworld/compare.go` + dispatch/
      usage in `cmd/promptworld/main.go` — `duelReport`/`duelSide` (data-model §6),
      per-side OfflineState + manifest-scenario exercise lookup + `RecordedPassFor` +
      `ResolveRubricFacts` + plain-language outcome map (data-model §7; raw enums never
      print), lineage-derived default window (`--since` override), header with lineage
      line + running-world as-of note + different-exercises note, `RenderReportCard` per
      side, ambient-world honest no-scorecard note (plan D6, spec FR-015/016/018/019/020,
      US2 scenarios 1–5)
- [X] T016 [US2] Scoreboard tests in new `cmd/promptworld/compare_test.go`: SC-004 — the
      decided-duel fixture (winner recorded pass → all-✓ evidence-backed; loser
      run-ended → ✗ `agent.died: N`), raw-enum sweep over rendered output
      (`in_progress` never appears), cross-surface identity (duel rows byte-equal to the
      resolver-fed postmortem rows for the same state), live `…` markers on a running
      fixture, ambient honesty, different-exercise note (spec SC-004)

## Phase 5: User Story 3 — divergence + interleaved chronicles (P3, board AC #3)

**Goal**: the drill-down — where the two stories split, and the two chronicles against
each other; zero divergence rendered honestly.

**Independent Test**: shared-prefix fixture logs — marker at first differing STORY event;
machinery-only difference → no divergence; interleave labeled and tick-ordered.

- [X] T017 [US3] Divergence scan + chronicle interleave in `cmd/promptworld/compare.go`:
      post-window story-event streams (exclude `daemon.*`/`clock.*`/`cog.*`/`llm.*`;
      compare tick/type/payload, never wall_time/seq — research R7) → first mismatch as
      `divergence` rendered with game day/time; nil → the identical-since-fork line;
      `chronicle.entry` events `from_tick >= since` merged stable by FromTick with world
      labels and the divergence marker in timeline position (plan D6, spec FR-017,
      US3 scenarios 1–4)
- [X] T018 [US3] Divergence tests in `cmd/promptworld/compare_test.go`: SC-005 — marker
      placement on a diverging fixture pair; a pair differing ONLY in machinery events
      renders NO divergence; the zero-divergence line pinned; interleave ordering +
      labels; `--since` override window (spec SC-005)

## Phase 6: Design reference — authority gate (spec 047)

- [X] T019 `node scripts/check-tui-design.mjs --changed` from the worktree: re-verify +
      re-pin every page flagged by the `internal/tui` diff (reportcard.go/views.go/
      digest.go pins); amend only where a page states something the diff falsifies (none
      expected — verify, don't assume); gate passes (plan D7, spec FR-021)

## Phase 7: Grounding — wiki-in-PR obligations (in-branch, pr-gate enforced)

- [ ] T020 `/grounding-wiki:wiki-update` reconciliation over the branch diff;
      review-work re-pins expected on `world-save-directory.md`,
      `world-save-manifest-fields.md`, `world-save-path-helpers.md`,
      `cli-world-lifecycle.md`/`cli-promptworld.md`, `event-log.md`, `snapshots.md`,
      `sim-state-reducer.md`, the clock/world event-types child (new `world.forked`
      row), `instance-manager.md`, `report-card-renderer.md` (fourth consumer +
      exported surface), `llm-budget-degraded-mode.md` (fork wallet-inheritance
      note) and `deterministic-rng.md` (re-verify); a new `world-forking` concept note
      if the update plan calls for one; computed re-pins for other notes listing touched
      sources — all pinned to branch commits (plan D8, spec FR-021)
- [ ] T021 Regenerate `docs/player/` via the `player-docs` skill (wiki changed in T020);
      `node .claude/skills/player-docs/scripts/check-freshness.mjs --check` passes
      in-branch (plan D8, spec FR-021)

## Phase 8: Polish & close-out

- [ ] T022 Full proof: `gofmt -l` clean; `go test ./...` green (including
      `TestCatalogSweep` with the new event and every pre-existing snapshot/manifest
      byte-identity test — SC-007); `node scripts/check-merge-drift.mjs pr` from the
      worktree exits 0; PR opens carrying code + design re-pins + wiki + player docs
      together; merge via `gh pr merge --merge` ONLY (squash rewrites branch pins —
      observed hazard)
- [ ] T023 Post-merge (root): `git worktree remove .worktrees/task-67` + branch delete +
      ff-pull; spec-bridge sync, board AC ticks (#1–#7), tasks.md ticks, runbook
      execution-log row — derived state only, no grounding content on main

## Dependencies & Execution Order

- T001 → T002 → everything (link before implementation — AC #6; worktree before code).
- Phase 2: T003 → T007 (payload before ceremony); T004, T005, T006 parallel [P] with
  each other and with T003 (different packages).
- Phase 3 chain: {T003, T005, T006} → T007 → T008 → T009 → T010; T007 → T011 → T012.
  T012 also needs T004 (full-binary build compiles the digest).
- Phase 4: T013 ∥ T014 (different packages; both independent of Phase 3) → T015 → T016.
  T015's lineage-window path needs T005; its fixtures are cheapest built on T007's forks
  (soft dependency — synthetic logs acceptable).
- Phase 5: T015 → T017 → T018.
- T019 after all `internal/tui` code (T004, T013). T020 after ALL code; T021 after T020;
  T022 after everything; T023 post-merge.
- **MVP** = Phases 2–3 (the fork verb with lineage, determinism, and budget proofs);
  Phase 4 makes it a duel; Phase 5 completes board AC #3.

**Parallel opportunities**: T004/T005/T006 [P]; Phase 4's T013/T14 can start alongside
Phase 3 (disjoint files); compare fixtures (T016/T018) can be authored while T015 lands.
