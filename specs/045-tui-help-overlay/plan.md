# Implementation Plan: `?` help overlay in the TUI (every world)

**Branch**: `045-tui-help-overlay` | **Date**: 2026-07-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/045-tui-help-overlay/spec.md`

## Summary

A client-only presentation feature on three shipped TUI idioms. (1) **Layer**: a
`helpOpen` state on `Model`, checked at the head of `handleKey` (after ctrl+c, before
minibuffer focus), owning the keyboard while open — help becomes the head of the
esc-release chain; `?` opens it from every non-text-entry mode (it binds nowhere today).
(2) **Rendering**: a body-replacement panel (the solo-zoom precedent — no z-compositing
exists), chrome kept visible, sized under the exact-height invariant; scrolling copies
the `chronicleDetailPane` pager verbatim. (3) **Content**: static pages in a new
`internal/tui/help.go` — per-mode key tables (basic tier = footer-hinted keys, advanced
tier = the rest + layered globals, source of truth `docs/design/tui/patterns/keymap.md`),
a screen walkthrough (header anatomy incl. every conditional badge, map glyph legend
derived from a table shared with `renderMapGrid` so it cannot drift, dock tabs), and a
lessons pull-reference section that is an empty data table with a documented contract.
No daemon, IPC, event, or world-state changes. Decision log: [research.md](./research.md).

## Technical Context

**Language/Version**: Go 1.24 (single module)

**Primary Dependencies**: charmbracelet/bubbletea + lipgloss (existing TUI stack); no
new dependencies

**Storage**: none — all content is static strings compiled into the client

**Testing**: `go test ./internal/tui` — key-routing tests (focus_test.go patterns),
exact-height render sweep (`TestWidescreenViewExactHeight` gains help states),
content-presence assertions (views_test.go style), a keymap↔overlay mechanical sweep
(SC-003), nil-status no-LLM tests

**Target Platform**: terminal client only (`internal/tui`); zero daemon surface

**Project Type**: single Go project; this feature is confined to `internal/tui` + the
keymap design doc

**Performance Goals**: overlay open/close is pure view-state mutation — no perceptible
latency; no ticking impact (the world runs beneath it untouched)

**Constraints**: exact-height render invariant at all sizes; focus contract (esc
releases exactly one layer; no silent no-op keys while a layer is focused); minibuffer
text entry never hijacked (`?` types); content model-independent (works with nil
status/replica); FR-005 legend anti-drift via shared table

**Scale/Scope**: ~6 modes' key pages + 3 walkthrough sections + reference seam;
one new file (`help.go` + `help_test.go`), edits to `tui.go` (dispatch), `views.go`
(body swap, footer hints, legend table extraction), keymap doc footer table

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.* (v1.1.0)

- **I. Artifact-Grounded Action — PASS.** Derived from TASK-116's board ACs, the
  learning-game synthesis decision 8, and the keymap design doc; produces spec/plan/
  tasks artifacts and the overlay content itself is keymap-doc-derived.
- **II. One Task, One PR — PASS.** TASK-116 ↔ this spec ↔ one branch
  (`.worktrees/task-116`) ↔ one PR.
- **III. Gates Over Assertions — PASS.** spec-bridge gates TASK-116; the SC-003
  keymap↔overlay sweep test is a new mechanical gate created by this feature.
- **IV. Grounding Freshness — PASS (deferred obligation).** Touches files pinned by TUI
  wiki notes; wiki-update after merge. The keymap design doc is updated in-PR (footer
  hint table).
- **V. Model-Tiered Workflow — PASS.** Planned on the planning tier; implementation is
  a **Sonnet** slice by rubric: single-package, view/rendering code, tests alongside —
  no concurrency, no doctrine surface. Recorded on TASK-116 at dispatch.

**Post-design re-check: PASS** — no violations; Complexity Tracking empty.

## Project Structure

### Documentation (this feature)

```text
specs/045-tui-help-overlay/
├── plan.md              # This file
├── research.md          # Phase 0 — R1..R8 decision log
├── data-model.md        # Phase 1 — content/state model
├── quickstart.md        # Phase 1 — validation walkthrough
├── contracts/
│   └── help-content.md  # Overlay structure, tier semantics, lessons-seam contract
└── tasks.md             # Phase 2 (/speckit-tasks)
```

### Source Code (repository root)

```text
internal/tui/
├── help.go          # NEW: page content tables (mode keys basic/advanced, walkthrough
│                    #   sections, lessons reference seam) + helpPanelView
├── help_test.go     # NEW: routing, tiers, content presence, keymap sweep, no-LLM
├── tui.go           # helpOpen/helpTier/helpSection/helpScroll on Model; dispatch at
│                    #   head of handleKey; overlay key handling
├── views.go         # body swap in widescreenView/narrowView; footer hints gain
│                    #   "? help"; legend string extracted to a shared glyph table
│                    #   consumed by both renderMapGrid and help.go
└── (render_test.go, focus_test.go, views_test.go — extended, not restructured)

docs/design/tui/patterns/keymap.md   # footer-hint table gains ?; overlay keys documented
```

**Structure Decision**: all changes inside `internal/tui` plus one design doc. The only
cross-file refactor is extracting the legend format string (views.go:615-617) into a
glyph table both the renderer and the overlay consume (FR-005/SC-003 anti-drift).

## Complexity Tracking

No Constitution Check violations — table intentionally empty.

## Known collision (recorded)

`task-31` (spec 044) is concurrently adding an ENDED header token, a grave glyph +
legend entry, and footer/header tests in the same files (views.go, views_test.go).
Expect a rebase before PR; the shared glyph table makes the legend merge mechanical, and
overlay header-anatomy content must include the ENDED posture once it exists (whichever
branch lands second reconciles — noted in tasks.md polish).
