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

- [X] T007 Add the per-chapter correction tally to the Mind (attributed count, distinct
      attributed locations, harvester id set), reset alongside `md.narrLines` in
      `closeChapter`.
- [X] T008 Route `chronicleNote`'s `agent.map_corrected` arm
      (`internal/mind/narrate.go:159-172`) through `attributedHarvest` on the first
      `Gone` fact — the existing first-fact convention. Attributed: fold into the tally,
      emit no line (FR-003). Unexplained: emit today's line, byte-identical (FR-004).
- [X] T009 Emit the single attributed summary line at `closeChapter`, naming the
      correction count, the distinct-location count, and the harvesters resolved to
      names through the same roster path the existing lines use.
- [X] T010 FR-008: a chapter with zero attributed corrections produces no summary line
      and is byte-identical to today.
- [X] T011 Unit tests for T007–T010, including the mixed chapter of User Story 2
      (40 attributed + 1 unexplained → one summary line plus one untouched anomaly line)
      and the SC-002 assertion that corrections contribute at most one line per chapter.

## Phase 3: Prompt and telemetry

- [X] T012 FR-005: mark the coalesced line as ordinary background rather than storyline
      material in the narrator prompt (`internal/mind/narrate.go:~705-712`), without
      disturbing the existing "group by storyline, not by hour" instruction for
      everything else.
- [X] T013 FR-007: record per-chapter attributed vs unexplained correction counts on the
      Mind's existing telemetry path, so the soak reads the outcome directly.
- [X] T014 Tests for T012–T013; then `go build ./...`, `go test ./...`, and
      `go test -race ./internal/mind/...` green (SC-005).

## Phase 4: Evidence (replay + re-narrate — runbook amendment 2026-08-02, operator-decided)

The evidence bar's window is unchanged (≥12 game-days of real data); the route is replay
of the preserved soak world's own event log rather than a fresh live soak, so the
before/after comparison is controlled on identical input. See the runbook's amended
"HOST ADDITION — the evidence bar" section, including its recorded limitation.

Preserved world (READ-ONLY — never mutate, it is the before-side of the comparison):
`/Users/evanstern/.claude/jobs/ca35de11/tmp/soak/soak-world/world.db` (12.02 game-days).
Its `world.db` is WAL-mode; `sqlite3 -readonly` FAILS on it now the daemon has stopped —
open it without `-readonly`, and copy it before any experiment that could write.

- [ ] T015 Build the replay harness: feed the preserved world's event log through the
      Mind's absorb path in order and capture, per chapter, the exact `md.narrLines`
      buffer the narrator would receive. Offline — no model calls in this task.
- [ ] T016 SC-004 (precision) and SC-003 (anti-suppression), from T015's replay:
      assert 100% precision — no correction lacking a coordinate-matching harvest is
      classified attributed — and that the 3 known genuinely-unexplained corrections
      each still produce their own line.
- [ ] T017 SC-002 (volume), from T015's replay: report per chapter the correction lines
      before vs after, and confirm attributed corrections contribute at most one line
      per chapter and no chapter overflows `narrMaxLines` on their account. The
      before-side baseline to reproduce is in `spec.md`'s per-chapter table (median 57%,
      peak 68%, 5 of 12 day chapters over the 120-line ring).
- [ ] T018 SC-001 (the actual outcome): re-narrate the replayed chapters with the live
      model — `gemma4:12b-mlx`, and `qwen3.6` too where feasible (SC-006) — and report
      whether any **named** absence storyline slug appears, plus the count and share of
      absence-themed entries against the 18/90 (20%) baseline.
- [ ] T019 Record all four measurements on TASK-173 (the runbook's (a)–(d)), then tick
      board AC#2 and AC#3 against that evidence — not before.

## Phase 5: Grounding and PR

- [ ] T020 Re-verify and re-pin `docs/wiki/chronicle.md` in-branch against the actual
      diff (honest re-pin: classify RE-PIN-ONLY vs NEEDS-REVIEW, amend prose first).
      Phase 2/3 changed `internal/mind/narrate.go` substantially, so this is a genuine
      NEEDS-REVIEW, not a mechanical bump. Check `agent-mind.md` / `mental-maps.md`
      sources and re-pin only if touched.
- [ ] T021 Regenerate `docs/player/` if any `docs/wiki/` note changed; confirm with
      `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`.
- [ ] T022 Update the runbook's execution log row, then run
      `node scripts/check-merge-drift.mjs pr` from the worktree; resolve every blocking
      finding and treat its semantic-overlap warnings as the companion-artifact checklist.
- [ ] T023 Open the PR from the worktree. Merge with `gh pr merge --merge` — never
      squash: this branch carries in-branch wiki re-pins.
