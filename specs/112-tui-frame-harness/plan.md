# Plan — spec 112, TUI frame harness

**Spec:** `specs/112-tui-frame-harness/spec.md` · **Board task:** TASK-187

## Constitution check (`.specify/memory/constitution.md` v1.3.0)

| Principle | How this plan complies |
| --- | --- |
| I. Artifact-Grounded Action | The deliverable IS an artifact generator; frames land as tracked files under `docs/design/tui/frames/`, and the fixture recipe is version-controlled. |
| II. One Task, One PR | TASK-187 → one branch (`task-187-frame-harness`) → one PR. The phases below are internal breakdown and land as commits, never their own PRs. |
| III. Gates Over Assertions | Determinism (FR-003) and fidelity (FR-007) are asserted by tests, not by claim. Lane 0's pin repair is verified by `check-tui-design.mjs` exiting 0, not by assertion. |
| IV. Grounding Freshness | The branch amends `docs/design/tui/` in the same PR; any wiki note pinning a touched source is re-verified and re-pinned in-branch per spec 069. |
| V. Model-Tiered Workflow | Cross-package/architectural → Opus 5 (`claude-opus-5`) via the `spec-implementer-opus` agent definition. Rubric justification recorded on TASK-187 and in the runbook. |

## Key technical findings (discovered during planning — these shape the design)

### F1 — `tui.New()` reads per-user state, and that breaks determinism

`internal/tui/tui.go:381` `New()` calls `worlds.LoadLessonsSeen()` and
`worlds.LoadUnlocks()`, both of which read the **operator's home directory**. The same
fixture would therefore render differently depending on which lessons that operator has
seen and which curriculum stages they have unlocked. Any frame harness that ignores this
produces machine-dependent output and FR-003 fails silently.

**Consequence:** the harness needs an injection seam for per-user state, defaulting to a
fixed, empty-or-canned record rather than whatever is on disk. This is the single most
important design constraint in this spec.

### F2 — the layout-bearing Model fields are unexported

`m.width`, `m.height`, `m.solo`, `m.dockTab`, `m.chronSelected`, `m.status` and friends
are package-private, which is why only in-package tests can drive them today. A command
in `cmd/promptworld` cannot construct a posed Model from outside.

**Consequence:** the rendering entry point must live **inside `internal/tui`** as an
exported, documented dev/design API — not a pile of exported fields. `cmd/promptworld`
stays a thin CLI shell over it.

### F3 — the existing test helpers are the proven recipe

`internal/tui/tui_test.go` `testModel`/`widescreenModel` and
`internal/tui/focus_test.go:308` `seedEvents` already build and pose a Model headlessly,
and `internal/tui/render_test.go` already enumerates the state names and forces a
`termenv` profile for stable output. The harness should be the productization of these,
so the tests and the harness cannot disagree about what a state means.

### F4 — the clock is the other nondeterminism source

The header renders tick, day, wall time and speed. Frames must be produced against a
frozen clock and a canned `ipc.StatusData`, never a live one.

## Design

```
cmd/promptworld/frames.go        thin CLI: flag parsing, fixture selection, output routing
        │
        ▼
internal/tui/design.go (new)     exported dev API:
        │                          - FrameOptions{Fixture, State, Width, Height, ANSI}
        │                          - Frame(opts) (string, error)
        │                          - States() []string      // single source of state names
        ▼
internal/tui/fixtures.go (new)   the three recipes: world + events + status + per-user
                                 state, all pinned; shared with the test helpers
```

- **`States()` is the single registry** of state names, consumed by the harness, the
  matrix driver, and `render_test.go`. One list, so the matrix cannot drift from the tests.
- **Fixtures are constructed in-process**, not by shelling out to `promptworld new`. That
  sidesteps the earned-stage gate (spec 046) entirely — satisfying AC #1's
  machine-independence — and keeps the recipe a Go value rather than a directory of logs.
- **Per-user state is injected** (F1) with a fixed canned record per fixture.
- **The clock is frozen** and `ipc.StatusData` is canned (F4).
- **ANSI is suppressed by default** by forcing the `termenv` profile the tests already
  force (F3), with `--ansi` opting back in.

### Interactive launch (FR-006)

`promptworld frames --fixture <f> --interactive` materializes the fixture into a temp
world dir and runs the existing `tea.NewProgram(tui.New(w), …)` path from
`cmd/promptworld/commands.go:848`. Reusing `cmdUI`'s exact program construction is what
keeps the interactive view and the dumped frame the same thing.

### Fidelity test (FR-007 / AC #9)

An in-package test builds a Model through the fixture path, calls `View()` directly, and
asserts equality with `Frame()` output for the same options — at least one page per
fixture. This is the assertion that makes every dumped frame trustworthy.

## Phases

Phased so each lands as its own commit on the single branch. See `tasks.md`.

1. **Lane 0 + fixture core** — pin repair (done); fixture recipes with injected per-user
   state and frozen clock.
2. **Render API** — `internal/tui/design.go`, `States()`, ANSI control, state posing.
3. **CLI surface** — `cmd/promptworld/frames.go`, flags, interactive mode.
4. **Matrix dump** — driver writing `docs/design/tui/frames/`, plus the generated frames.
5. **Gates + grounding** — determinism test, fidelity test, `gofmt`/build/test,
   `check-tui-design --changed`, wiki/player-docs freshness, design-doc amendment.

## Risks

- **R1 — hidden per-user or environment reads beyond F1.** Mitigation: the determinism
  test regenerates the whole matrix twice in one process AND asserts against a committed
  copy, so an environment-dependent read fails the suite rather than the reviewer's eye.
- **R2 — the matrix is large.** Guard against a churny diff: keep sizes to the few that
  straddle the breakpoint and the column split, not a sweep of every width.
- **R3 — `docs/design/tui/frames/` is a new directory in a gate-governed surface.**
  `check-tui-design.mjs` enforces a `pages/panels/overlays/patterns` taxonomy and rejects
  files outside it. Verify `frames/` is either accepted or explicitly exempted, and amend
  the checker in this same PR if not — this is a real possibility that must be checked
  early, not at PR time.
- **R4 — scope creep into the gate wiring.** Explicitly out of scope; the frames land as
  files, and nothing yet fails a build if they go stale.
