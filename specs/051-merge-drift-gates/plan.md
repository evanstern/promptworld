# Implementation Plan: Merge-Drift Gates

**Branch**: `051-merge-drift-gates` | **Date**: 2026-07-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/051-merge-drift-gates/spec.md`

## Summary

One zero-dependency Node script, `scripts/check-merge-drift.mjs`, invoked in three modes
(`session`, `worktree`, `pr`) at the SDLC's choke points. Textual conflict prediction and
the n-way drift matrix ride `git merge-tree --write-tree` (verified working on this
repo's git 2.50.1); semantic collision warnings come from changed-file set intersections
(branch vs mainline since merge-base) against the project's workflow surfaces
(`backlog/`, wiki-pinned sources, `internal/tui/`, `specs/NNN-*`); cleanup eligibility
combines an ancestor check with an empty-contribution merge-tree check (squash case);
grounding freshness delegates to the existing per-surface checkers where they exist and
reads wiki note frontmatter (`sources:` + `verified_against:`) directly otherwise.
Findings print as a report with a tri-state verdict, exit codes gate (0 pass / 1 blocked
/ 2 usage-or-env error), and task-attributable findings are recorded as board notes via
the `backlog` CLI with fingerprint dedup. Follows the spec-047 `check-tui-design.mjs`
precedent for shape, contracts, and CLAUDE.md enforcement.

## Technical Context

**Language/Version**: Node.js ≥ 18, ESM, single file (local toolchain: v24.17.0)

**Primary Dependencies**: none (node stdlib: `fs`, `path`, `child_process`); external
binaries: `git` ≥ 2.38 (`merge-tree --write-tree`; local: 2.50.1), `backlog` CLI
(optional — only for `--notes`), existing freshness checkers
(`scripts/check-tui-design.mjs`, `.claude/skills/player-docs/scripts/check-freshness.mjs`)
invoked as child processes when present

**Storage**: none. Reads git object DB, `docs/wiki/*.md` frontmatter, worktree status.
Writes: git fetch, root ff-pull (session mode, guarded), board notes via `backlog` CLI
(opt-in), worktree/branch cleanup (opt-in). Never writes files under `backlog/` directly;
never modifies any task branch.

**Testing**: quickstart validation scenarios against a disposable fixture repo
(recipe in quickstart.md), matching the spec-047 precedent of contract-documented
manual validation for `.mjs` gate scripts; no automated test harness for scripts exists
in this repo and none is introduced

**Target Platform**: developer machines (darwin/linux), run from repo root or a worktree

**Project Type**: CLI gate script + project-doc updates (CLAUDE.md)

**Performance Goals**: full run < 30 s at ≤ 10 live branches (SC-002); expected cost is
one fetch + O(n²) merge-tree calls at ~10–50 ms each (≤ 55 calls at n = 10)

**Constraints**: deterministic given identical repo+remote state (FR-012); no services
beyond the git remote (FR-011); pr/worktree modes fail closed without a fresh fetch
(FR-014); mutation whitelist per FR-009

**Scale/Scope**: this repo — ~45 wiki notes, ≤ ~10 live branches, one remote (`origin`),
branch convention `task-<N>-<slug>`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment |
|---|---|
| I. Artifact-Grounded Action | PASS — the feature's entire purpose: every gate run yields a report + exit code; task-attributable findings become board notes (FR-010). |
| II. One Task, One PR | PASS — one deliverable (script + contracts + CLAUDE.md section), TASK-131, one branch `task-131-merge-drift-gates` in `.worktrees/task-131`, one PR. Spec docs commit to main per project practice. |
| III. Gates Over Assertions | PASS — implements the principle. All board writes via `backlog` CLI; the script never edits derived state by hand. |
| IV. Grounding Freshness | PASS — FR-008 makes the gate a freshness tripwire. Implementation adds `scripts/check-merge-drift.mjs`; at merge time, re-check whether any wiki note lists `scripts/` files as sources and re-pin if so. |
| V. Model-Tiered Workflow | PASS — plan authored on the planning tier; implementation delegated to `spec-implementer`. Tier: **Sonnet (default)** — single-package script, no concurrency/governor logic, no doctrine-adjacent behavior; rubric justification to be recorded on TASK-131 at implement time. |

**Post-Phase-1 re-check**: design introduces no new projects, no dependencies, no
external services; mutation surface is the FR-009 whitelist verbatim. Still PASS on all
five principles. Complexity Tracking stays empty.

## Project Structure

### Documentation (this feature)

```text
specs/051-merge-drift-gates/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output (fixture-repo validation scenarios)
├── contracts/
│   ├── gate-cli.md      # CLI contract: modes, flags, exit codes, mutation whitelist
│   ├── report-schema.md # report + finding + matrix JSON shapes, fingerprints
│   └── detection-rules.md # the git plumbing per detection, normative
└── tasks.md             # Phase 2 output (/speckit-tasks — NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
scripts/
└── check-merge-drift.mjs    # the entire implementation: single-file ESM gate script

CLAUDE.md                    # new "Merge-drift gates" section (three invocations,
                             # when each is mandatory) next to the spec-047 gate block
```

**Structure Decision**: single-file script under `scripts/`, mirroring
`scripts/check-tui-design.mjs` (spec 047) — same runtime floor (Node ≥ 18), same
zero-dependency rule, same contract-doc pattern, same CLAUDE.md enforcement precedent.
No new directories, packages, or build steps.

## Complexity Tracking

No constitution violations — table intentionally empty.
