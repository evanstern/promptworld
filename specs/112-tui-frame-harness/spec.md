# Spec 110 — TUI frame harness: fixture worlds and headless frame dump

**Board task:** TASK-187
**Status:** planning
**Created:** 2026-08-02

## Problem

Looking at a TUI frame currently costs a compile-and-run cycle. The renderer is trivially
inspectable — `internal/tui/views.go:75` `View() string` returns exactly the characters
that reach the terminal, and `internal/tui/tui.go:381` `New(w *world.World) Model` needs
only a world directory — but the only existing headless path is inside Go tests
(`internal/tui/tui_test.go:39` `widescreenModel`, `internal/tui/focus_test.go:308`
`seedEvents`). Because looking is expensive, nobody looks, and the UI is developed by
reasoning about lipgloss calls rather than about rendered output.

The cost of that shows up as documented drift. `docs/design/tui/pages/home.md` carries a
"Reconciliation correction" section recording that its hand-drawn mockup showed a
right-aligned villager-count header segment and raw inline-JSON feed lines the shipped
renderer never rendered; a human reconstructed that by hand. The mockups are prose
drawings unattached to the renderer, so nothing detects their drift.

Two audiences need the same artifact. A human wants to open the real UI on an interesting
world instantly, without running a daemon, configuring an LLM, or waiting for a village to
develop. An agent implementing a UI change needs to read the frame it is producing rather
than guess at it — for a terminal UI this is uniquely achievable, because the frame *is* a
string.

## Goal

Make a frame cheap to produce, deterministic enough to diff, and identical for both
audiences.

## Users

- **The operator doing UI/UX work** — wants the real interactive UI on a canned world in
  one command, and wants to compare a proposed layout against the current one.
- **The implementing agent** — wants to print any screen at any size and read it back.
- **The PR reviewer** — wants a UI change to arrive as a before/after of real frames.
- **The design-doc reader** — wants mockups generated from the renderer, not drawn.

## Functional requirements

### FR-001 — Three fixture worlds
Three canned worlds, built deterministically from a checked-in recipe:

| id | shape | exercises |
| --- | --- | --- |
| `empty` | fresh stage-1, no events, no daemon | cold start, zero-event chronicle, empty-state and disconnected rendering |
| `mid-game` | populated later-stage ambient world, deep event backlog, mixed roster | dense/overflow case, truncation, the three teaching chrome rows |
| `scenario` | a `--scenario` world from the spec 054 catalog | the exercise dock tab and lesson row, absent from ambient worlds |

The **recipe** is version-controlled, not generated world directories — a world dir's
logs churn on every regeneration and would make the repo noisy.

Fixture construction MUST NOT depend on the operator's earned-stage unlock state
(`promptworld new` refuses unearned stages without `--override`, spec 046), or fixtures
build differently on different machines.

### FR-002 — Headless frame render
A command renders one (fixture, page, state, size) combination and prints the frame to
stdout. Shape:

```
promptworld frames --fixture mid-game --state home --size 160x50
```

It MUST run with no daemon, no LLM configured, and no sim advancing.

### FR-003 — Determinism
Rendering the same combination twice MUST produce byte-identical output. No wall-clock
time, live tick counter, or unpinned randomness may reach a frame. Note the header alone
renders tick, day, wall time and speed — `internal/clock` is the seam to freeze.

### FR-004 — Plain text by default
Dumped frames MUST be plain text with ANSI escapes suppressed, so they diff. An opt-in
flag preserves color codes for live eyeballing. `internal/tui/render_test.go` already
forces a `termenv` profile — reuse that seam rather than inventing a second one.

### FR-005 — Scene matrix dump
A driver writes every combination to `docs/design/tui/frames/`, one file per combination.
The starting matrix is the state set already enumerated in `internal/tui/render_test.go`
— `home`, `solo`, `inspect`, `inspect-solo`, `villagers-solo`, `villagers-detail-solo`,
`guardian-solo`, `help`, `help-advanced`, `help-walkthrough`, `help-lessons` — across
sizes straddling the widescreen breakpoint and the 50/50 column split.

> **Amendment (orchestrator, 2026-08-02, after phase 2).** This list said `metatron-solo`
> when written, copied verbatim from `render_test.go`. That name cannot ship in a
> production `.go` file: `internal/lint/fiction_sweep_test.go` (spec 052 SC-001) bans bare
> fiction literals in non-test Go sources, and the old spelling survived in
> `render_test.go` only because that sweep skips `_test.go` files. Hoisting the state
> registry into `design.go` is exactly the case the gate exists to catch. Renamed to
> **`guardian-solo`** — the state poses `paneGuardian` either way, and every other name in
> the list is already post-spec-052 vocabulary. TASK-187's card carries the same
> correction.

### FR-008 — narrow frames are content-height, not terminal-height

Discovered during phase 2 and recorded here so the matrix dump does not encode a false
invariant. `TestWidescreenViewExactHeight` asserts `len(lines) == height` only at widths at
or above the widescreen breakpoint, because the narrow fallback has no fold arithmetic
(`docs/design/tui/patterns/layout.md`, ruling b). A frame rendered at 80×30 is therefore
content-height and shorter than 30 lines — correct renderer behavior, not a defect. The
matrix dump, and any height assertion over it, MUST NOT assume terminal-height frames below
the breakpoint.

### FR-006 — Interactive launch
The operator can launch the real interactive TUI against any fixture in one command and
drive it by keyboard.

### FR-007 — Fidelity
A frame dumped by the harness MUST equal what the same Model renders at the same size.
This is the requirement that makes every other one trustworthy.

## Out of scope

- Wiring the frames into `scripts/check-tui-design.mjs` as a regenerate-and-diff gate —
  a posture change deserving its own review decision. This spec is its prerequisite.
- Replacing prose mockups in `docs/design/tui/pages/*.md` with generated frames.
- Color or palette judgment: perceived color is not readable from a frame dump.
- tui.studio integration or an MCP over it — evaluated and declined as an authority on
  TASK-187's card.

## Acceptance criteria (mirrors TASK-187)

1. Three fixtures build deterministically on any machine, independent of unlock state.
2. One command renders any (fixture, page, state, size) with no daemon/LLM/sim.
3. The full matrix regenerated twice is byte-identical.
4. Frames are plain text by default; an opt-in flag keeps ANSI.
5. The matrix covers every `render_test.go` state across breakpoint-straddling sizes,
   one file per combination under `docs/design/tui/frames/`.
6. The `mid-game` fixture shows awake, asleep and dead villagers and an overflowing
   chronicle, so truncation and the chrome rows are visible.
7. The `scenario` fixture renders the exercise tab and lesson row; ambient ones do not.
8. The operator can launch the interactive TUI against any fixture in one command.
9. A test asserts harness output equals `View()` for at least one page per fixture.

## Lane 0 precondition (resolved in this branch)

`docs/design/tui/anatomy.md` pinned `4eb6471a`, a commit orphaned by a squash merge,
making `check-tui-design.mjs` exit 1 repo-wide and blocking every TUI PR. Repaired to
`012032fb` — the commit that landed the file's current content — following the precedent
of `6318cf8b`. See the runbook's Lane 0 for the full reasoning.
