# Implementation Plan: Wiki grounding moves inside the PR cycle

**Branch**: `069-wiki-in-pr-gate` (task branch: `task-145-wiki-in-pr-gate`) | **Date**: 2026-07-26 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/069-wiki-in-pr-gate/spec.md`

## Summary

Escalate the merge-drift pr gate from advising about wiki-pinned sources to
BLOCKING PRs whose branches don't carry their own wiki re-verification
(pin-vs-branch predicate at the branch tip), add an in-branch player-docs
freshness block, and move the doctrine (CLAUDE.md + constitution IV) to the
in-PR lifecycle with merge-commit-only merges and a derived-state-only rule
for post-merge main commits.

## Technical Context

**Language/Version**: Node.js ESM (`scripts/*.mjs`), no deps beyond node: builtins

**Primary Dependencies**: git plumbing via the script's existing `git()` helper;
`.claude/skills/player-docs/scripts/check-freshness.mjs` (spawned, exit-code contract)

**Storage**: none

**Testing**: `node --test scripts/check-merge-drift.test.mjs scripts/claim-protocol.test.mjs`
(fixture git repos in temp dirs)

**Target Platform**: darwin/linux dev machines (same as today's script)

**Project Type**: repo tooling + doctrine docs

**Performance Goals**: pr mode stays interactive (<~2s overhead; note-at-tip reads
scale with touched notes, not corpus size)

**Constraints**: session/worktree/claim modes behavior-unchanged (FR-006); no
bypass flag (FR-010); hook (`scripts/hooks/merge-drift-hook.mjs`) needs no change
— it already blocks on exit 1

**Scale/Scope**: ~1 script + its test file, CLAUDE.md, constitution; ~5 files

## Constitution Check

- **I. Artifact-Grounded Action** — PASS: operator direction recorded on
  TASK-145 + runbook; decisions in spec/research; gate output is the artifact.
- **II. One Task, One PR** — PASS: TASK-145 ↔ `task-145-wiki-in-pr-gate` ↔ one
  PR (script + tests + CLAUDE.md + constitution amendment together).
- **III. Gates Over Assertions** — PASS: this spec strengthens that principle's
  machinery.
- **IV. Grounding Freshness** — PASS: `scripts/` files are not wiki-pinned
  (verified: no note lists them as sources); CLAUDE.md/constitution are not
  wiki-pinned either. This PR self-applies the new rule trivially (no pinned
  sources touched) but MUST pass the new gate it introduces before opening.
- **V. Model-Tiered Workflow** — PASS: script/tests/CLAUDE.md dispatched to
  spec-implementer on **Opus 4.8** (doctrine-adjacent behavior change to
  SDLC-critical infrastructure; a gate defect blocks every future PR in every
  lane — recorded on TASK-145). The constitution amendment itself is
  planning-tier work: the orchestrator runs `speckit-constitution` and commits
  the amendment onto the same task branch.

**Post-Phase-1 re-check**: PASS — no new projects/deps; no Complexity Tracking
entries.

## Project Structure

### Documentation (this feature)

```text
specs/069-wiki-in-pr-gate/
├── CLAIM.md
├── spec.md
├── checklists/requirements.md
├── plan.md              # this file
├── research.md          # R1–R5
└── tasks.md
```

(data-model.md / contracts/ / quickstart.md folded into spec Decisions +
research — the "entities" are two findings and a predicate; a separate model
document would restate the spec verbatim.)

### Source Code (repository root)

```text
scripts/
├── check-merge-drift.mjs        # gatePr: predicate, finding swap, player-docs spawn,
│                                #   scoped malformed escalation (D1–D4)
└── check-merge-drift.test.mjs   # fixture-repo matrices for US1/US2 (D5)
CLAUDE.md                        # PDLC loop + rules rewrite (FR-007)
.specify/memory/constitution.md  # Principle IV amendment (FR-008, orchestrator)
docs/design/pdlc-hardening-runbook.md  # execution log row (orchestrator, post-merge)
```

## Design

- **D1 — note-at-tip reader**: `loadWikiNotesAt(ref, notePaths, cwd)` reading
  `git show <ref>:<path>` through the existing frontmatter parser; used only
  for notes whose sources intersect `branchFiles` (cheap).
- **D2 — predicate in gatePr**: for each `wikiSourcesOverlap` hit, evaluate
  (note ∈ branchFiles) ∧ (pin readable at T) ∧ (`merge-base --is-ancestor pin
  T`) ∧ (`rev-list pin..T -- matched` empty). Fail → `wiki-repin-missing`
  block finding (message: note, matched sources, remedy). Pass → no finding at
  all (FR-002). `wiki-note-malformed` at T for a predicate-needed note →
  block (R5).
- **D3 — player-docs spawn**: `branchFiles.some(f => f.startsWith('docs/wiki/'))`
  → spawn checker (path from `CHECK_MERGE_DRIFT_PLAYER_DOCS_CHECKER` env or
  default), map exit 1 → `player-docs-stale` block, exit 2 →
  `player-docs-env-error` block, 0 → nothing.
- **D4 — docs**: CLAUDE.md loop diagram gains "wiki grounding (in-branch)"
  between build and PR; Grounding-freshness and Player-docs rules rewritten to
  the pre-PR choke point; merge-commit-only line; step-7 derived-state-only
  paragraph citing spec 065 + pdlc:sweep re-ground. Constitution IV amended
  via speckit-constitution (v1.1.0 → 1.2.0, Sync Impact Report).
- **D5 — tests**: fixture matrices per spec US1 (block / re-pinned pass /
  re-touched source block / unreachable pin block / no-overlap unchanged) and
  US2 (stub checker exits 0/1/2); plus existing-harness regression (all prior
  tests green untouched).

## Complexity Tracking

No violations — intentionally empty.
