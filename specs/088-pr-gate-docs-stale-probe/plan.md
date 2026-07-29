# Implementation Plan: merge-drift pr gate — docs-stale probe on all pinned sources + history moves

**Branch**: `task-162-pr-gate-docs-stale-probe` | **Date**: 2026-07-29 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/088-pr-gate-docs-stale-probe/spec.md`

## Summary

Widen the pr-mode docs-stale probe in `scripts/check-merge-drift.mjs` from its single
trigger (`branchFiles` touching `docs/wiki/`, line ~1645) to three triggers: (1) the
branch touches ANY source a `docs/player/` page declares (README.md,
docs/llm-providers.md, spec 046 quickstart sources — derived, not hardcoded), (2) the
branch has undergone a history move (a merge commit in `origin/main..tip`), and (3)
design-reference pin drift (`docs/design/tui/*`) becomes a BLOCKING pr finding by
delegating to `scripts/check-tui-design.mjs` the same way the gate already delegates
to the player-docs checker. All new triggers covered by the existing node:test
synthetic-repo fixture harness (`scripts/check-merge-drift.test.mjs`).

## Technical Context

**Language/Version**: Node.js ESM (`.mjs`), no transpilation — matches the existing script

**Primary Dependencies**: none (node builtins + `git` subprocess; delegated checkers:
`.claude/skills/player-docs/scripts/check-freshness.mjs`, `scripts/check-tui-design.mjs`)

**Storage**: N/A (stateless gate; findings are exit codes + stdout)

**Testing**: `node --test scripts/check-merge-drift.test.mjs` (existing synthetic-repo
fixture harness; env override `CHECK_MERGE_DRIFT_PLAYER_DOCS_CHECKER` pattern already
established — add the analogous override for the tui-design checker)

**Target Platform**: developer workstations (darwin/linux), invoked by hooks and sessions

**Project Type**: repo tooling (single script + test file)

**Performance Goals**: pr mode stays interactive (<2s overhead beyond delegated checkers)

**Constraints**: stateless (no persisted probe timestamps — spec Assumption); exit-code
contract unchanged (0 pass / 1 blocked / 2 env error); no new blocking behavior in
session/worktree/claim modes

**Scale/Scope**: one script (~2000 lines) + its test file; ~5 new fixtures

## Constitution Check

- **I. Artifact-Grounded Action**: PASS — spec 088 + TASK-162 card carry the diagnosis;
  this plan and fixtures are the new artifacts.
- **II. One Task, One PR**: PASS — single branch `task-162-pr-gate-docs-stale-probe`,
  one PR.
- **III. Gates Over Assertions**: PASS — the deliverable strengthens a gate; every new
  trigger is fixture-proven, not asserted.
- **IV. Grounding Freshness**: PASS — wiki notes pinning
  `scripts/check-merge-drift.mjs` as a source get re-verified + re-pinned on this
  branch before the PR (the gate self-applies).
- **V. Model-Tiered Workflow**: PASS — plan/spec on the planning tier; implementation
  dispatched to spec-implementer at **Sonnet** (single-script tooling change with
  fixtures; routine slice). Escalation trigger: none foreseen.

Post-Phase-1 re-check: PASS (no new projects, no new dependencies, no doctrine change —
the gate's severity model and exit contract are preserved).

## Project Structure

### Documentation (this feature)

```text
specs/088-pr-gate-docs-stale-probe/
├── spec.md
├── plan.md              # this file
├── research.md          # decisions: trigger derivation, history-move predicate, tui delegation
├── data-model.md        # trigger set, finding rules, fixture matrix
├── quickstart.md        # how to run the gate + fixtures
├── checklists/requirements.md
└── tasks.md             # /speckit-tasks output
```

### Source Code (repository root)

```text
scripts/
├── check-merge-drift.mjs        # pr-mode probe: trigger widening + tui delegation
├── check-merge-drift.test.mjs   # new fixtures (see data-model.md fixture matrix)
└── check-tui-design.mjs         # UNCHANGED — invoked by the gate with --changed --json
```

**Structure Decision**: all changes live in the two existing script files; the
tui-design checker is consumed as-is through its published CLI contract
(`--changed [<range>] --json`, exit 0/1/2).

## Complexity Tracking

No constitution violations; table omitted.
