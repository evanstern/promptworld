# Implementation Plan: Guardian console page + systems-tab telemetry split

**Branch**: `053-guardian-console` | **Date**: 2026-07-25 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/053-guardian-console/spec.md`

## Summary

Two coupled moves on the dock/pages architecture: (1) a new first-class
full-screen page (`G`) rendering the guardian conversation document-style
over the existing `Model.transcript`, with the standard minibuffer as its
composer, a charter/skills read surface fed by the existing status fields, a
`$EDITOR` shell-out (`tea.ExecProcess`) with changed-file confirmation, and a
documented inline card seam (renderer plugs in via TASK-127/115); (2) a
fourth dock tab **systems** (key `5`) receiving the relocated telemetry
renderers (`llmProviderLines`, `horizonLines`/`horizonRow`, spend/wallet
lines) out of the guardian tab, which keeps fiction-layer content only (D10
skin boundary as a file boundary).

## Technical Context

**Language/Version**: Go 1.24 (repo toolchain)

**Primary Dependencies**: Bubble Tea (existing; adds `tea.ExecProcess` usage for $EDITOR), lipgloss

**Storage**: N/A (console has no persistence; charter stays the on-disk file the daemon owns)

**Testing**: `go test -race ./...`; page navigation/focus-contract tests (tui_test.go, focus_test.go harness); render tests for console blocks + systems/guardian split over LLM and no-LLM fixtures (SC-002); $EDITOR round-trip with a scripted fake editor; keymap regression suite

**Target Platform**: terminal client

**Project Type**: single Go module; dominant package `internal/tui` (+ possible small status read helpers)

**Performance Goals**: no render regression; console renders the transcript tail, not the whole history per frame

**Constraints**: focus contract unrelaxed (minibuffer is the only focusable input); dock state machine unchanged for tabs 2–4; systems tab carries zero skin tokens; same-PR design-doc amendments across ≥6 pages; expect mid-flight rebases over TASK-121/124/126 merges

**Scale/Scope**: the sweep's biggest TUI slice — new page + tab + shell-out; est. 800–1,400 LOC delta incl. tests; 6+ design pages amended

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Artifact-Grounded Action** — PASS: spec dir `specs/053-*`; TASK-125 linked pre-implementation; scope ruling (card seam) recorded on spec + board.
- **II. One Task, One PR** — PASS: `.worktrees/task-125`, one branch, one PR.
- **III. Gates Over Assertions** — PASS: check script + race suite + spec-bridge gate.
- **IV. Grounding Freshness** — PASS (planned): `internal/tui` sources → wiki + player-docs re-ground (understanding-the-screen page will need regeneration).
- **V. Model-Tiered Workflow** — PASS: planned on Fable 5; implementation on **Sonnet** per the runbook Lane 2 assignment (view/rendering code in one package, tests alongside; no concurrency/doctrine logic — the $EDITOR exec is framework-standard). Escalation trigger: if gates fail twice or the focus-contract regressions prove architectural, escalate to Opus per the rubric (runbook: "escalate to Opus if gates fail").

**Post-Phase-1 re-check**: PASS — one new page-level state + one tab enum extension; no new abstractions beyond the card-seam interface (deliberately minimal: a slice of renderable cards, empty until later tasks).

## Project Structure

### Documentation (this feature)

```text
specs/053-guardian-console/
├── plan.md
├── research.md          # seam decisions (ExecProcess, tab enum, card interface)
├── data-model.md        # page state, card seam, tab extension
├── quickstart.md
├── contracts/
│   └── console-and-systems.md  # navigation grammar, split inventory, card-seam interface
└── tasks.md
```

### Source Code (repository root)

```text
internal/tui/
├── tui.go               # page state (console open/return-target); `G`/`5`/`e` key
│                        #   handling; ExecProcess cmd + result msg; dockTab enum + paneNames
├── views.go             # consoleView (document blocks, header, read surface, card seam,
│                        #   composer pairing); systems tab content renderer (relocated
│                        #   llmProviderLines/horizonLines/spend); guardian tab minus telemetry
├── help.go              # help overlay content gains G/5 (per overlays/help.md content rules)
├── focus_test.go        # focus contract with the console page in the loop
├── views_test.go        # split render tests (SC-002); console block rendering
├── tui_test.go          # navigation round-trips; $EDITOR round-trip (fake editor);
│                        #   keymap regression
└── layout.go            # (read-only) page uses existing full-screen math

docs/design/tui/
├── pages/guardian-console.md  # specified → shipped, real symbols, card seam noted
├── panels/systems.md          # specified → shipped (tab exists now), key 5
├── panels/guardian.md         # content list minus telemetry; re-pin
├── panels/dock.md             # 4-tab row; re-pin
├── patterns/keymap.md         # G/5/e + console scroll keys; parity gaps; footer hints
└── pages/solo-views.md        # narrow reachability of systems; re-pin
```

**Structure Decision**: all in `internal/tui`. The console is a page (like
solo views), not an overlay and not a dock tab — it gets its own top-level
branch in `View()` beside the widescreen/narrow fork, per the design page's
navigation section.

## Complexity Tracking

No constitution violations — table intentionally empty.
