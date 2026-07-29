# Implementation Plan: wiki corpus size-budget debt

**Branch**: `task-165-wiki-size-budget` | **Date**: 2026-07-29 | **Spec**: [spec.md](spec.md)

## Summary

Doc-only reconciliation: resolve all 26 size-budget findings (24 note bodies, 2
capsule descriptions) via tighten / summary-style split / justified exemption,
regenerate CAPSULES.md by script, update INDEX.md, refresh stale docs/player pages,
and land with the freshness gate at 0 findings.

## Technical Context

**Language/Version**: Markdown corpus + Node scripts (no code changes)

**Primary Dependencies**: grounding-wiki plugin 0.39.0 (`gates/cli.mjs freshness`,
`scripts/capsules.mjs`); player-docs project skill (`check-freshness.mjs`, page regen)

**Storage**: `docs/wiki/`, `docs/player/` (tracked files)

**Testing**: gate runs are the tests — grounding-wiki freshness exit 0; player-docs
probe exit 0; merge-drift pr gate exit 0

**Target Platform**: repo docs

**Project Type**: corpus maintenance

**Performance Goals**: N/A

**Constraints**: no hand-edits to CAPSULES.md (derived); split children carry full
honest frontmatter; no facts dropped anywhere

**Scale/Scope**: 25 notes + CAPSULES.md + INDEX.md + affected docs/player pages

## Constitution Check

- **I**: PASS — inventory table in spec.md is the paper trail; gate runs are evidence.
- **II**: PASS — one branch, one PR.
- **III**: PASS — the deliverable IS gate-greening; every remedy verified by the gate.
- **IV**: PASS — this task is grounding maintenance itself; pins stay honest
  (children pinned to the branch commit whose content was verified).
- **V**: PASS — **Sonnet** tier: doc reconciliation, the rubric's named routine
  slice. Escalation trigger: none.

Post-Phase-1 re-check: PASS.

## Project Structure

```text
docs/wiki/            # tightened/split notes + new children; INDEX.md updated
docs/wiki/CAPSULES.md # regenerated via capsules.mjs only
docs/player/          # pages stale-ed by wiki edits, regenerated
specs/089-wiki-size-budget/  # this spec
```

**Structure Decision**: no source code; the corpus-spec split procedure governs file
shapes.

## Research (inline — one decision)

- **D1 Remedy selection**: tighten for small overages (<~300 chars) where prose
  redundancy covers the gap; split for large overages following the corpus's
  existing parent/child idiom (reflex-policy, sim-state, executor families show the
  pattern); exempt only with a written reason. Rationale: minimizes churn on
  pages/pins for small cases, keeps big notes routable. Alternative (split
  everything) rejected: 26 splits would double the diff and stale far more player
  pages for no routing gain.

## Complexity Tracking

No violations; table omitted.
