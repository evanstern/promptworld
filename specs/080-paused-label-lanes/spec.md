# Feature Specification: paused-label lanes downgrade to info

**Feature Branch**: `paused-lane-marker` (board: TASK-155) ·
**Created**: 2026-07-26 · Counterpart of **praxisflux TASK-55**
(`specs/015-paused-lane-marker` in the praxis repo), which ships the
convention; this spec ships the gate behavior.

## The convention (shared with praxisflux)

An operator pauses an In Progress task without moving it on the board: the
card carries a **`paused` label**, set/cleared only via
`backlog task edit TASK-<n> --labels …` (never a hand edit), machine-findable
in the task file's frontmatter `labels:` list. Provenance rides an
append-note at pause time — "paused by \<who\> \<date\>: \<why\>" — and
clearing the label gets a matching resume note.

## Requirements

- **FR-001**: `scripts/check-merge-drift.mjs` detects a paused task by
  reading the `paused` label from the task file's frontmatter `labels:` list
  (block-list and inline `[a, b]` YAML forms), in the MAIN worktree's working
  tree — the pause is a local operator act on the board of record, and origin
  lag must not resurrect blocking findings for a lane the operator just
  paused.
- **FR-002**: a paused task's branch/worktree **drift findings** downgrade
  from block/warn to **info**, with the pause cited as evidence
  (`paused:TASK-<n>` appended) and named in the message, mode-uniformly:
  - session: `textual-conflict`, `pairwise-conflict` (either side paused),
    `stale-base`*, `branch-unpushed`, `backlog-overlap`,
    `spec-number-collision` (branch-sourced), `cleanup-eligible`
    (*already info in session mode; unchanged);
  - pr: `textual-conflict` (was block), `stale-base`, `backlog-overlap`,
    `spec-number-collision` for a gated branch attributed to a paused task;
  - worktree: emits no branch-scoped drift findings today, so the downgrade
    is vacuously satisfied; any future branch-scoped finding must route
    through the same `makeDriftFinding` wrapper.
- **FR-003**: session mode **never prescribes cleanup** of a paused task's
  branch/worktree: no `cleanupPrescriptions` entry, so `--apply-cleanup`
  can never remove it; an info-level `cleanup-eligible` finding keeps the
  parked lane visible.
- **FR-004** (deliberate non-downgrades): pausing is **not a gate bypass** —
  - ownership protections stay blocking: claim mode's and worktree mode's
    `spec-number-collision` guard a paused task's claimed number exactly as
    before (they protect the paused task, they are not noise about it);
  - pr mode's spec-069 grounding blocks (`wiki-repin-missing`,
    `wiki-note-malformed`, `player-docs-stale`, `player-docs-env-error`)
    stay blocking regardless of pause.
- **FR-005**: downgraded findings are info-severity, so `--notes` never
  writes them to the board (applyNotes skips info) — a paused card is not
  spammed with drift notes.

## Coverage

Regression tests in `scripts/check-merge-drift.test.mjs` ("spec 080" section)
per the repo's existing bare-origin + clone fixture convention: session-mode
downgrades (pairwise either-side, textual vs main, branch-unpushed) with live
controls unchanged, cleanup never prescribed for a paused branch (prescribed
for the live control), pr-mode exit-1 → exit-0 flip with the pause as
evidence, spec-069 blocks surviving a pause, and worktree/claim ownership
blocks surviving a pause.
