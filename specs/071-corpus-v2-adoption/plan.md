# Implementation Plan: Wiki corpus-spec v2 adoption

**Branch**: `071-corpus-v2-adoption` (task branch: `task-146-corpus-v2`) | **Date**: 2026-07-26 | **Spec**: [spec.md](spec.md)

**Note**: the design lives in spec.md's Decisions 1–6 (this feature is a
document restructure; a separate design layer would restate them). This file
carries the execution plan the tasks.md encodes, and exists as the
spec-bridge derivation's plan artifact. Authored at execution time — the
bridge's Done derivation requires plan.md present, which the original
fold-into-spec choice missed; recorded honestly rather than backdated.

## Summary

Adopt corpus-spec v2 for `docs/wiki/`: split/tighten the 37 over-budget note
bodies into ≤8,000-char parents + children (summary-style, parents keep
filenames, children inherit pins), rewrite over-budget capsules for routing,
generate `CAPSULES.md` last, and land only when the freshness gate passes in
v2 failure mode with player docs fresh in-branch.

## Technical Context

**Language**: Markdown corpus + node scripts (capsules.mjs, freshness gate,
player-docs checker). **Testing**: the gates themselves. **Constraints**:
no re-verification (pins inherited), no filename changes on parents,
single-writer INDEX/CAPSULES (orchestrator), workers run git-free in one
shared worktree.

## Constitution Check

- **I–III** — PASS: worklist derived from the gate's own output; decisions in
  spec.md; the v2 gate flip is the enforcement artifact.
- **IV** — PASS: this IS grounding work, riding its own PR under the spec-069
  in-PR rules (player-docs checked in-branch; no Go sources touched so the
  re-pin predicate is vacuous).
- **V** — PASS: orchestrator-led grounding-docs work (grounding docs are the
  orchestrator's hands) with 14 Sonnet subagent batches for mechanical
  splits; audit, INDEX, capsules, and gates stayed with the orchestrator.

## Execution shape (as run)

1. Worklist derivation (sizes per note, split/tighten/exempt buckets per
   spec Decision 2).
2. 14 parallel Sonnet batches over disjoint note sets (A–N), each reporting
   children, INDEX one-liners, char accounting, and exemptions (none needed).
3. Orchestrator tail: corpus-wide size/frontmatter audit (161 notes, 0
   violations), INDEX reconciliation (88 missing child lines inserted,
   derived from each child's routing capsule), CAPSULES.md generation,
   freshness gate in v2 failure mode, player-docs check, single commit, pr
   gate, PR #109 (merge commit).

## Complexity Tracking

Empty — no violations.
