# Implementation Plan: Mouse-parity sweep test

**Branch**: `073-mouse-parity-sweep` (task branch: `task-154-mouse-parity-sweep`) | **Date**: 2026-07-26 | **Spec**: [spec.md](spec.md)

## Summary

One new test file (`internal/tui/mouseparity_test.go`) that parses every
canonical-header control table under `docs/design/tui/`, classifies each
`keys+mouse` cell, and enforces — both directions — that every non-'—' mouse
claim has a hand-audited oracle entry whose live `tea.MouseMsg` dispatch
proves the handler; plus a prose amendment to
`docs/design/tui/patterns/keymap.md` doctrine rule 3 naming the mechanized
gate. No non-test code, no new mouse features.

## Technical Context

**Language**: Go (package `tui` tests), stdlib only (`os`, `path/filepath`,
`strings`, `testing`; `bubbletea` for `tea.MouseMsg`). **Corpus access**:
relative read from the package dir — `../../docs/design/tui` — the
`TestStageDefaultsSweep` precedent (`internal/tui/stagedefaults_test.go:125`).
**Scope**: `internal/tui/mouseparity_test.go` (new) +
`docs/design/tui/patterns/keymap.md` (rule 3 amendment + re-pin).
**Constraints**: spec FR-008 — gate the existing surface only.

## Constitution Check

- **I–III** — PASS: decision trail on TASK-154 + reorient decision 8 + this
  spec; one task, one branch (`task-154-mouse-parity-sweep`), one PR; the
  deliverable IS a gate, and the design/merge-drift gates run at their choke
  points.
- **IV** — PASS (planned): no wiki note pins the touched files (spec
  Assumptions); pr gate probes expected clean, verified in-branch, gate is
  the authority. `patterns/keymap.md` is design corpus, not wiki — its
  freshness gate is `check-tui-design.mjs --changed` (same-PR re-verify +
  re-pin), which this branch runs and satisfies.
- **V** — PASS: dispatched to `spec-implementer` on **Sonnet** (board note on
  TASK-154: tooling/test, single surface — routine tier per the rubric; no
  concurrency, no cross-package, no doctrine-adjacent behavior change).
  Escalation unused unless gates fail.

**Post-Phase-1 re-check**: PASS — design below adds no violations.

## Design

- **D1 — Parser** (`parseControlTables` in the test file): walk
  `../../docs/design/tui` (`filepath.WalkDir`, `.md` only; `t.Fatalf` if the
  root is unreadable — FR-006). In each file, find lines equal (after
  whitespace-normalizing the cell padding) to the canonical header
  (`| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |`
  — mirror `CANONICAL_HEADER` in `scripts/check-tui-design.mjs`); skip the
  `|---|` divider; consume subsequent `|`-prefixed rows. Split each row on
  `|`, trim cells, take column 5 (index 5 of the split, matching the awk/JS
  gate's column order).
- **D2 — Cell classifier** (`classifyKeysMouse`): (a) no ` · ` substring and
  cell begins with `—` → display-only; (b) otherwise split on ` · `, mouse
  half = last segment, keys half = the rest rejoined; mouse half exactly `—`
  → tracked gap; (c) anything else → shipped claim
  `{page, control, mouseClaim}` where `page` is the corpus-relative path
  (e.g. `panels/chronicle.md`) and `control` is column 1. A cell containing
  ` · ` whose *first* segment is `—` is malformed → fail loudly (defensive;
  none exist).
- **D3 — Oracle** (`mouseParityOracle`): `map[mouseClaimKey]func(t *testing.T)`
  with exactly one initial entry —
  `{page: "panels/chronicle.md", control: "jump-to-source", claim: "click line"}`.
  Its check builds `widescreenModel(t)` + `seedEvents`, enters inspect mode
  (paused, selection live), calls `m.View()` to populate the `chronHit` hit
  region, computes an in-region coordinate from `m.chronHit`, sends
  `mouseLeftRelease(x, y)` through `Update`, and asserts the documented
  effect: `chronSelected` set to the clicked row's event and the camera
  centered per `jumpToSource` (crib the existing live-mouse tests near
  `internal/tui/tui_test.go:872`). Doc comment states the oracle's audit
  contract: an entry may only be added alongside the corpus cell and code it
  proves (keymap doctrine rule 1 — keyboard and mouse land together).
- **D4 — The sweep** (`TestMouseParitySweep`): parse (D1/D2); assert
  non-vacuity (≥ 1 table, ≥ 1 shipped claim — FR-006); direction 1: every
  shipped claim has an oracle entry, and run its check as a named subtest
  (`t.Run(page+"/"+control, ...)`) — FR-002/003; direction 2: every oracle
  key was seen in the corpus — FR-004; rollout-note honesty: for every parsed
  file with ≥ 1 tracked-gap row, the file's text contains `**Parity rollout**`
  — FR-005. Error messages name page file + control/region.
- **D5 — keymap.md amendment** (FR-007): within `## Input-parity doctrine`
  rule 3 of `docs/design/tui/patterns/keymap.md`, state that per-control
  tracking is now mechanized by `TestMouseParitySweep`
  (`internal/tui/mouseparity_test.go`): every non-'—' mouse cell must carry a
  live-dispatch proof, rollout notes are asserted present wherever a table
  has keyed-but-mouseless rows, and a control graduates out of its note by
  gaining the real mouse cell + oracle entry + passing proof in one PR.
  Keep rules 1–2 and the rest of the page untouched; update the page's
  `verified_against` pin to a branch commit (the `--changed` gate's same-PR
  re-verify duty; merge-commit-only doctrine keeps the pin reachable).
- **D6 — Proof commands** (board AC + SC-001..004):
  `go test ./internal/tui/ -run TestMouseParity -v`; the SC-001 mutation
  check (temporary local edit of one `—` cell → sweep fails naming it →
  revert); `go test ./...`; `node scripts/check-tui-design.mjs --changed`;
  `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`
  (expected clean); `node scripts/check-merge-drift.mjs pr` from the
  worktree.

## Complexity Tracking

Empty — no violations.
