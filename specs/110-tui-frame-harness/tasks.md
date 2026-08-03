# Tasks — spec 110, TUI frame harness

**Board task:** TASK-187 · **Branch:** `task-187-frame-harness`

One task, one PR. Every phase below lands as a commit on this single branch.

## Phase 1 — Lane 0 precondition and fixture core

- [x] T001 Repair the dead design pin: `docs/design/tui/anatomy.md`
      `verified_against` `4eb6471a` → `012032fb`; verify `node scripts/check-tui-design.mjs`
      exits 0. (Done by the orchestrator before dispatch — see the runbook's Lane 0.)
- [x] T002 Add `internal/tui/fixtures.go` with a `Fixture` type and the three recipes
      (`empty`, `mid-game`, `scenario`) as in-process Go values — world, map, event feed,
      villager roster, `ipc.StatusData`, skin/stage config. No shelling out to
      `promptworld new`, so the earned-stage gate (spec 046) is bypassed by construction
      and AC #1's machine-independence holds.
- [x] T003 Inject per-user state. `tui.New()` reads `worlds.LoadLessonsSeen()` and
      `worlds.LoadUnlocks()` from the operator's home dir (plan.md F1) — give the fixture
      path a seam that supplies a fixed canned record instead. This is the primary
      determinism fix; without it frames differ per machine.
- [x] T004 Freeze the clock and can the status so tick/day/wall-time/speed in the header
      are fixed (plan.md F4).
- [x] T005 `mid-game` specifically carries awake, asleep AND dead villagers plus a
      chronicle backlog deep enough to overflow the pane (AC #6); `scenario` is built from
      the spec 054 exercise catalog so the exercise tab and lesson row render (AC #7).

## Phase 2 — Render API

- [x] T006 Add `internal/tui/design.go` exporting `FrameOptions{Fixture, State, Width,
      Height, ANSI}` and `Frame(opts) (string, error)`. It lives in-package because the
      layout-bearing Model fields are unexported (plan.md F2).
- [x] T007 Export `States() []string` as the single registry of state names, and refactor
      `internal/tui/render_test.go` to consume it, so the matrix and the tests cannot
      disagree about what states exist (plan.md F3).
- [x] T008 Pose each state (`home`, `solo`, `inspect`, `inspect-solo`, `villagers-solo`,
      `villagers-detail-solo`, `metatron-solo`, `help`, `help-advanced`,
      `help-walkthrough`, `help-lessons`) reusing the existing test-helper logic rather
      than a parallel implementation.
- [x] T009 ANSI suppressed by default via the `termenv` profile `render_test.go` already
      forces; `ANSI: true` opts back in (AC #4).

## Phase 3 — CLI surface

- [ ] T010 Add `cmd/promptworld/frames.go`: `promptworld frames --fixture <f> --state <s>
      --size <WxH> [--ansi]`, printing one frame to stdout. No daemon, no LLM, no sim
      (AC #2).
- [ ] T011 Register the verb in the command table and add `--list` to enumerate fixtures
      and states.
- [ ] T012 `--interactive` materializes a fixture into a temp world dir and runs the same
      `tea.NewProgram(tui.New(w), tea.WithAltScreen(), tea.WithMouseCellMotion())`
      construction as `cmdUI` (`cmd/promptworld/commands.go:848`), so the interactive view
      and the dumped frame are the same thing (AC #8).
- [ ] T013 Unit tests for flag parsing and size parsing, alongside the code.

## Phase 4 — Matrix dump

- [ ] T014 **Check R3 first:** confirm `scripts/check-tui-design.mjs` accepts a new
      `docs/design/tui/frames/` directory — its taxonomy check rejects files outside
      `pages/panels/overlays/patterns`. If rejected, amend the checker in this same PR
      and say so in the design docs. Do this before generating the matrix, not at PR time.
- [ ] T015 Add `promptworld frames --dump` writing every (fixture, state, size)
      combination to `docs/design/tui/frames/`, one file per combination, over sizes
      straddling the widescreen breakpoint and the 50/50 column split (AC #5).
- [ ] T016 Generate and commit the matrix.
- [ ] T017 Add `docs/design/tui/frames/README.md` explaining what the directory is, how
      to regenerate it, and that it is generated — never hand-edited.

## Phase 5 — Gates and grounding

- [ ] T018 **Determinism test (AC #3):** regenerate the full matrix twice in one process
      and assert byte-identical output; additionally assert the generated output matches
      the committed copy, so an environment-dependent read fails the suite (plan.md R1).
- [ ] T019 **Fidelity test (AC #9):** assert `Frame(opts)` equals `View()` for the same
      posed Model, at least one page per fixture.
- [ ] T020 `gofmt -l` clean, `go build ./...`, `go test ./...` green.
- [ ] T021 Amend `docs/design/tui/` for this feature and re-pin every affected page;
      `node scripts/check-tui-design.mjs --changed <range>` exits 0.
- [ ] T022 Wiki-in-PR (spec 069): re-verify and re-pin, in-branch, any wiki note listing a
      touched file in its `sources:`; if `docs/wiki/` changed, regenerate `docs/player/`.
      Probe with `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`.
- [ ] T023 `node scripts/check-merge-drift.mjs pr` exits 0 from the worktree.
- [ ] T024 Update the runbook execution log and flip its status.
