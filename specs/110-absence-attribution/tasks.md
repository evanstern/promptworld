# Tasks: Absence attribution (TASK-173, spec 110)

**Spec**: `spec.md` · **Plan**: `plan.md` · **Branch**: `task-173-absence-attribution`

Phases are internal breakdown, not PR boundaries — every phase lands as commits on this
one branch and merges in TASK-173's single PR. One implementer agent is dispatched per
phase (fresh context each time); the handoff between phases is this file's tick state,
the spec dir, and the branch's commits — never chat context.

## Phase 1: Ledger and classifier

- [X] T001 Add `harvestLedgerWindow` (4 game-days in ticks) and `harvestLedgerCap` to the
      const block in `internal/mind/narrate.go`, each with the comment recording its
      measured justification (lag tops out at 3 game-days; 352 distinct locations over
      12 game-days).
- [X] T002 Add the harvest ledger to the Mind: a coordinate-keyed store of
      `{agentID, tick}` with oldest-first eviction on both age (window) and count (cap).
      Owned by the absorb goroutine; no locking beyond what the Mind already holds.
- [X] T003 Populate the ledger from the existing `agent.chopped` / `agent.quarried`
      absorb arm in `internal/mind/mind.go:315-346` — the arm already decodes
      `sim.HarvestPayload{Agent, X, Y}`; add the ledger write without disturbing the
      spec-081 witness re-arm logic that shares the arm.
- [X] T004 Add `attributedHarvest(x, y int, atTick int64) (agentID int, ok bool)` — a
      pure classifier over the ledger, exported within the package and directly testable.
- [X] T005 Unit tests for T002–T004: population, age eviction at the window edge, cap
      eviction, exact-coordinate matching, and a miss returning `ok == false`.
- [X] T006 Verify Phase 1 is inert: `chronicleNote` output is unchanged by this phase.
      Assert it with a test that renders a correction line before and after the ledger
      exists.

## Phase 2: Coalesced narration

- [ ] T007 Add the per-chapter correction tally to the Mind (attributed count, distinct
      attributed locations, harvester id set), reset alongside `md.narrLines` in
      `closeChapter`.
- [ ] T008 Route `chronicleNote`'s `agent.map_corrected` arm
      (`internal/mind/narrate.go:159-172`) through `attributedHarvest` on the first
      `Gone` fact — the existing first-fact convention. Attributed: fold into the tally,
      emit no line (FR-003). Unexplained: emit today's line, byte-identical (FR-004).
- [ ] T009 Emit the single attributed summary line at `closeChapter`, naming the
      correction count, the distinct-location count, and the harvesters resolved to
      names through the same roster path the existing lines use.
- [ ] T010 FR-008: a chapter with zero attributed corrections produces no summary line
      and is byte-identical to today.
- [ ] T011 Unit tests for T007–T010, including the mixed chapter of User Story 2
      (40 attributed + 1 unexplained → one summary line plus one untouched anomaly line)
      and the SC-002 assertion that corrections contribute at most one line per chapter.

## Phase 3: Prompt and telemetry

- [ ] T012 FR-005: mark the coalesced line as ordinary background rather than storyline
      material in the narrator prompt (`internal/mind/narrate.go:~705-712`), without
      disturbing the existing "group by storyline, not by hour" instruction for
      everything else.
- [ ] T013 FR-007: record per-chapter attributed vs unexplained correction counts on the
      Mind's existing telemetry path, so the soak reads the outcome directly.
- [ ] T014 Tests for T012–T013; then `go build ./...`, `go test ./...`, and
      `go test -race ./internal/mind/...` green (SC-005).

## Phase 4: Evidence

- [ ] T015 SC-004: replay the preserved soak world's event log through the ledger +
      classifier and assert 100% precision — no correction lacking a coordinate-matching
      harvest is classified attributed, and the 3 known anomalies classify as
      unexplained.
- [ ] T016 Run the soak: ≥12 game-days, same scenario, on `gemma4:12b-mlx`; and on
      `qwen3.6` where feasible (SC-006 — a single-model run records its reason).
- [ ] T017 Record the runbook's four required measurements on TASK-173: (a) count and
      share of absence-themed chronicle entries; (b) whether any *named* absence
      storyline slug appears; (c) the harvest-explained share of corrections; (d) the
      count of genuinely-unexplained absences and evidence they still surfaced.
- [ ] T018 Tick board AC#2 and AC#3 against T017's evidence — not before.

## Phase 5: Grounding and PR

- [ ] T019 Re-verify and re-pin `docs/wiki/chronicle.md` in-branch against the actual
      diff (honest re-pin: classify RE-PIN-ONLY vs NEEDS-REVIEW, amend prose first).
      Check `agent-mind.md` / `mental-maps.md` sources and re-pin only if touched.
- [ ] T020 Regenerate `docs/player/` if any `docs/wiki/` note changed; confirm with
      `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`.
- [ ] T021 Update the runbook's execution log row, then run
      `node scripts/check-merge-drift.mjs pr` from the worktree; resolve every blocking
      finding and treat its semantic-overlap warnings as the companion-artifact checklist.
- [ ] T022 Open the PR from the worktree. Merge with `gh pr merge --merge` — never
      squash: this branch carries in-branch wiki re-pins.
