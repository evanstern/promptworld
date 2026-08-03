---
id: TASK-187
title: >-
  TUI frame harness: three fixture worlds and a headless frame dump for rapid UI
  iteration
status: To Do
assignee: []
created_date: '2026-08-03 00:33'
labels:
  - tui
  - design
  - tooling
  - dx
dependencies: []
ordinal: 169001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Three canned fixture worlds plus a command that prints any TUI screen as plain text, so a UI change can be looked at — by a person or by an agent — without running the game. Today the only way to see a frame is to write a Go test and compile; this makes it one command, and dumps a matrix of frames to disk so UI drift shows up as a diff instead of an investigation.

As the operator doing UI/UX work, I want to open the real UI on a canned world instantly — no daemon, no LLM, no waiting for a village to develop into an interesting state — so I can judge a layout change in seconds instead of minutes.

As the agent implementing a UI change, I want to print the exact frame the terminal will show, so I am reasoning about the real rendered characters instead of guessing from the lipgloss calls that produce them.

As a reviewer of a UI pull request, I want the change to arrive as a before/after of actual frames, so I can see what moved without checking out the branch and reproducing the state by hand.

As anyone reading docs/design/tui, I want the mockups to be generated from the renderer rather than hand-drawn, so nobody has to reconcile a drawing against shipped code again.

## Why now — the drift is already documented

docs/design/tui/pages/home.md carries a section titled "Reconciliation correction" recording that its hand-drawn mockup showed a right-aligned villager-count header segment and raw inline-JSON feed lines that the shipped renderer never rendered. Someone reconstructed that by hand. This session's merge-drift gate separately reports docs/design/tui/anatomy.md as stale (grounding-stale, tui-design). The mockups drift because they are prose drawings unattached to the renderer.

The renderer, meanwhile, is already trivially inspectable and we are not using it. internal/tui/tui.go:381 `New(w *world.World) Model` takes a plain world directory — cmd/promptworld/commands.go:848 (`cmdUI`) proves the TUI runs with no daemon attached. internal/tui/views.go:75 `View() string` returns exactly the characters that reach the screen. internal/tui/tui_test.go:39 (`widescreenModel`) plus internal/tui/focus_test.go:308 (`seedEvents`) already build a Model headlessly at a chosen size — but only from inside a Go test, so every look costs a compile-and-run cycle. Because looking is expensive, nobody looks.

## The three fixtures (operator decision, 2026-08-02)

Selected over a single canonical demo world because most layout defects appear only at the extremes.

1. **empty** — fresh stage-1 world, no events, no daemon. Exercises cold start, zero-event chronicle, empty-state and disconnected rendering.
2. **mid-game** — populated later-stage ambient world with a deep event backlog and a mixed villager roster (awake, asleep, dead). Exercises the dense and overflowing case, the teaching chrome rows (villager strip, lesson row, guardian strip), and truncation.
3. **scenario** — a `--scenario` world from the spec 054 exercise catalog. Exercises the sixth dock tab (exercise), the lesson row, and the stage-shaped chrome an ambient world never shows.

## Scope

- A deterministic fixture builder producing the three worlds on demand from a checked-in recipe (pinned seed, stage, scripted event feed). The recipe is what is version-controlled — not generated world directories, whose logs would churn every regeneration.
- A frame command, shape `promptworld frames --page <p> --size <WxH> --state <s> --fixture <f>`, that builds the Model headlessly and prints `View()`.
- A scene matrix driver that dumps every combination to `docs/design/tui/frames/` as plain-text files.
- The states already enumerated by internal/tui/render_test.go are the starting matrix: home, solo, inspect, inspect-solo, villagers-solo, villagers-detail-solo, metatron-solo, help, help-advanced, help-walkthrough, help-lessons — across sizes straddling the widescreen breakpoint and the 50/50 column split.

## Explicitly out of scope (follow-on cards)

- Wiring the generated frames into scripts/check-tui-design.mjs as a regenerate-and-diff gate. That is a posture change deserving its own review decision, and this card is the prerequisite that makes it possible.
- Replacing the prose mockups inside docs/design/tui/pages/*.md with generated frames. Sequenced after the gate exists.
- Any color or palette judgment. Perceived color is not readable from a frame dump and stays an operator judgment; see the ANSI criterion below.
- tui.studio integration or an MCP over it. Evaluated and declined as an authority: its export is a component tree that would fight this renderer's skins, stage defaults, exact-height invariant and focus contract, and it does not model the ANSI-aware truncation, unicode width and lipgloss column math where TUI defects actually live. It stays a human sketchpad. The MCP worth building, if any, wraps this card's own harness.

## Risks the implementation must handle

- **Determinism is the whole value.** A frame containing wall-clock time, a live tick counter, or unpinned randomness diffs as noise on every run and the artifact becomes worthless. The header alone renders tick, day, wall time and speed. Every such source must be frozen through the fixture (internal/clock is the seam).
- **ANSI must be suppressed by default.** lipgloss emits escape codes; the dumped artifact must be plain text for diffability, with an opt-in flag to keep codes for live eyeballing. internal/tui/render_test.go already forces a termenv profile — reuse that seam rather than inventing one.
- **The stage gate refuses unearned stages.** `promptworld new` blocks stages the player has not earned unless `--override` (spec 046). Fixture generation must not depend on the operator's unlock state, or a fixture will build differently on different machines.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Three fixture worlds — empty, mid-game, scenario — build deterministically from a checked-in recipe on any machine, independent of the operator's earned-stage unlock state
- [ ] #2 A single command renders any (fixture, page, state, size) combination and prints the frame to stdout, with no daemon running, no LLM configured, and no sim advancing
- [ ] #3 Regenerating the full scene matrix twice in a row produces byte-identical output: no wall-clock time, live tick, or unpinned randomness leaks into a frame
- [ ] #4 Dumped frames are plain text with ANSI escapes suppressed by default; an opt-in flag preserves color codes for live eyeballing
- [ ] #5 The scene matrix covers every state enumerated in internal/tui/render_test.go across sizes straddling the widescreen breakpoint, and writes one file per combination under docs/design/tui/frames/
- [ ] #6 The mid-game fixture contains awake, asleep, and dead villagers and a chronicle backlog deep enough to overflow the pane, so truncation and the teaching chrome rows are visible in its frames
- [ ] #7 The scenario fixture renders the exercise dock tab and the lesson row, which the ambient fixtures do not
- [ ] #8 The operator can launch the interactive TUI against any of the three fixtures with one command and drive it by keyboard
- [ ] #9 A frame dumped by the harness matches what the same Model renders in the terminal at the same size — verified by a test asserting harness output equals View() for at least one page per fixture
<!-- AC:END -->
