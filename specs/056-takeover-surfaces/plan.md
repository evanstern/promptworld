# Implementation Plan: Takeover surfaces — ceremony + postmortem

**Branch**: `056-takeover-surfaces` | **Date**: 2026-07-25 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/056-takeover-surfaces/spec.md`

## Summary

One takeover family in `internal/tui`: an overlay-owner state (none /
ceremony / postmortem + deferred-ceremony flag) rendered in the existing
body-replacement slot (help-overlay discipline), with postmortem-always-wins
precedence; the postmortem composes run-end line + (scored only) the new
shared report-card renderer + morgue evidence rows from replica facts; the
ceremony composes the D6 authorship chapter + the same renderer
(authoritative instrument). Dismiss/replay per the pages: `esc`, ceremony
`q`-detach framing (D13), global `p` reopen (ended only), auto-open on
attach to ended worlds, `?`-overlay ceremony-replay entry. The renderer is
one implementation with concluded/live marker modes, composable into spec
053's console card seam (production wiring is TASK-115's).

## Technical Context

**Language/Version**: Go 1.24 (repo toolchain; no new deps)

**Primary Dependencies**: Bubble Tea/lipgloss (existing); `internal/sim` replica facts (death ledger, CurriculumPasses, exercise definitions); `internal/skin` lookups (TASK-121's contract when merged)

**Storage**: none — all content from recorded events/state; seen/replay content stored in the log + per-user unlocks (existing)

**Testing**: `go test -race ./...`; precedence interleaving fixtures (SC-003); ambient/scored matrix (SC-002); renderer three-site equivalence (SC-005); attach-to-ended auto-open; narrow exact-height (existing harness)

**Target Platform**: terminal client

**Project Type**: single package (`internal/tui`) + help-content amendment

**Performance Goals**: render-path only; morgue rows derive from the replica's ledger per open, not per frame

**Constraints**: body-replacement discipline (chrome visible, exact height); takeovers never stack; ENDED posture/read-only keys unaffected by dismissal; no new fiction literals (skin rules); same-PR design-doc amendments

**Scale/Scope**: est. 600–1,000 LOC incl. tests; 4–5 design pages amended

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Artifact-Grounded Action** — PASS: spec dir `specs/056-*`; TASK-127 linked pre-implementation; the two parked questions resolved from the authored pages (artifact-derived, recorded in spec).
- **II. One Task, One PR** — PASS: `.worktrees/task-127`, one branch, one PR.
- **III. Gates Over Assertions** — PASS: design gate + race suite + spec-bridge gate.
- **IV. Grounding Freshness** — PASS (planned): `internal/tui` sources → tui-client.md re-pin + player-docs in re-ground.
- **V. Model-Tiered Workflow** — PASS: planned on Fable 5; implementation on **Sonnet** (single-package view/overlay state machine, tests alongside — routine tier); escalate per rubric only if overlay/focus interactions fail gates.

**Post-Phase-1 re-check**: PASS — one overlay-owner enum + one renderer; no new packages. No Complexity Tracking entries.

## Project Structure

### Documentation (this feature)

```text
specs/056-takeover-surfaces/
├── plan.md
├── research.md          # overlay-owner state, renderer contract, morgue-row derivation,
│                        #   replay surfaces, skin posture
├── data-model.md
├── quickstart.md
├── contracts/
│   └── takeovers.md     # trigger/precedence/dismiss grammar; renderer contract
└── tasks.md
```

### Source Code (repository root)

```text
internal/tui/
├── tui.go               # takeover owner state + precedence transitions; p key (ended only);
│                        #   auto-open on runEnded() at connect; deferred-ceremony flag
├── views.go             # postmortemView, ceremonyView (body-replacement slot, help precedent);
│                        #   reportCardView (shared renderer, concluded/live modes);
│                        #   morgue evidence rows from replica facts
├── help.go              # ceremony-replay entry (spec 045 content contract amendment)
├── tui_test.go          # precedence interleavings; p/esc/q; attach-to-ended auto-open
├── render_test.go       # ambient/scored matrix; renderer three-site equivalence; exact-height
└── focus_test.go        # overlay/focus regressions

docs/design/tui/
├── overlays/ceremony.md     # specified → shipped, real symbols
├── overlays/postmortem.md   # specified → shipped, real symbols (renderer authored here)
├── overlays/help.md         # ceremony-replay entry recorded
├── patterns/keymap.md       # p live; takeover keys; parity gaps
├── pages/guardian-console.md# card-seam note names the shipped renderer
└── re-pins on all touched pages
```

**Structure Decision**: all in `internal/tui`; the renderer is a plain
view function so all three sites (two here, one via 115) share it without
interface ceremony beyond spec 053's existing consoleCard.

## Complexity Tracking

No constitution violations — table intentionally empty.
