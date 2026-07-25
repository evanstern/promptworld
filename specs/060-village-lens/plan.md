# Implementation Plan: Village lens completion — villager strip + map condition overlays

**Branch**: `060-village-lens` | **Date**: 2026-07-25 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/060-village-lens/spec.md`

## Summary

Two additive render features in `internal/tui`, both fully authored in the
design corpus: (1) `villagerStripView` — the D12 colonist-bar row under the
header (count + roster-ordered state-styled glyph run, end-drop overflow),
integrated into the spec-050 fold machinery (folds second, to the header
count badge; narrow = badge only); (2) three map condition overlays in the
`renderMapGrid` agents/structures pass — needs-critical, suppressed-mind
(from the existing client decision-trace projection), dying-fire (fuel
window before `FuelUntil`) — steady styles, documented priority, legend
entries. The spec's standing resolutions carry the plan-level rulings
(display-only strip; look-cursor deferred; no blink).

## Technical Context

**Language/Version**: Go 1.24 · **Primary Dependencies**: existing tui/lipgloss only · **Storage**: none · **Testing**: `go test -race ./...`, fixture render sweeps, layout height/width sweeps, color-profile distinguishability · **Target Platform**: terminal client · **Project Type**: single package (`internal/tui`) · **Performance Goals**: O(roster) strip, O(in-view agents) overlays per frame · **Constraints**: 1-row strip; fold-second ordering; steady styles; same-PR doc gate · **Scale/Scope**: ~300–500 LOC incl. tests; 4 design pages

## Constitution Check

- **I** PASS (spec dir; TASK-129 linked pre-implementation; rulings cite the authored pages).
- **II** PASS (`.worktrees/task-129`, one branch/PR).
- **III** PASS (design gate + race suite + merge-drift pr + spec-bridge).
- **IV** PASS planned (tui-client.md re-pin + player docs re-ground).
- **V** PASS — **Sonnet** (single-package rendering, tests alongside; routine tier).

**Post-design re-check**: PASS — no new state, no new packages; extends the
existing fold machinery and map render pass. No Complexity Tracking rows.

## Project Structure

```text
specs/060-village-lens/   # spec.md (carries the rulings) · tasks.md · this plan
internal/tui/
├── views.go              # villagerStripView; renderMapGrid overlay pass + legend entries
├── layout.go             # strip row + fold-second step (extends spec-050 machinery)
├── views_test.go / layout_test.go / render_test.go
docs/design/tui/
├── panels/villager-strip.md  # → shipped
├── panels/map.md             # stub → real overlay rows + look-cursor deferral
├── patterns/layout.md, pages/home.md  # re-verify/re-pin
```

**Structure Decision**: all in `internal/tui`; the authored pages are the
detailed behavior contract (mockups/budgets verbatim there, not duplicated
here).

## Complexity Tracking

No constitution violations — table intentionally empty.
