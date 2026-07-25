---
id: TASK-138
title: >-
  Merge-drift gate: player-docs evidence is objects, suppressing all repeat
  findings
status: In Progress
assignee: []
created_date: '2026-07-25 19:47'
updated_date: '2026-07-25 20:35'
labels:
  - gates
  - review-2026-07-25
  - code-quality
dependencies: []
priority: high
ordinal: 108000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Team review 2026-07-25 finding, verified at runtime and in source. scripts/check-merge-drift.mjs:521 extracts player-docs 'touched' evidence as OBJECTS, not strings:

  (parsed) => (parsed.pages || []).filter((p) => p.verdict !== 'fresh').flatMap((p) => p.sources || [])

p.sources elements are {path, recorded, current, fresh}. Compare the TUI extractor 10 lines below at :531, which correctly does .map((c) => c.file).

Three consequences, ascending severity:
1. COSMETIC — :677 and :686 join the array for the report, printing '[object Object], [object Object], ...'. Observed live in session-mode output (30 of them).
2. DEDUP BROKEN — :512 does [...new Set(touched)]; a Set of distinct object identities dedups nothing, so the 'touched sources' list is inflated with duplicates.
3. SILENT SUPPRESSION (the real bug) — fingerprint() at :268/:281 is computed from that same evidence via [...evidence].sort().join(','), so EVERY player-docs staleness finding fingerprints to the identical string. :741 skips writing a board note when existingText.includes(f.fingerprint). Therefore the FIRST player-docs staleness note ever written to a task suppresses ALL future ones on that task, regardless of which pages actually went stale.

Fix (one line): .flatMap((p) => (p.sources || []).map((s) => s.path))

Then verify: the fingerprint changes when a different page/source goes stale, and a second distinct staleness finding on the same task does write its note.

Trivial-exemption candidate per the constitution: surgical fix, complete file:line diagnosis pinned above, ACs on this card. Regression test belongs alongside — this class of bug (evidence shape feeding a dedup key) is invisible to exit-code testing.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 check-merge-drift.mjs:521 maps sources to their .path string; report output shows real paths, not [object Object]
- [x] #2 Set-based dedup at :512 actually dedups (duplicate sources collapse)
- [x] #3 Two DIFFERENT player-docs staleness findings on one task produce two different fingerprints and both write board notes
- [x] #4 Regression test covers the evidence-shape -> fingerprint path so a future extractor returning objects fails loudly
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Claimed 2026-07-25 by the review session, using the TASK-139 protocol: card moved and pushed BEFORE any work, so a competing session sees the claim. Constitution trivial exemption applies — surgical one-line fix, complete file:line diagnosis pinned on this card, ACs present; no Spec Kit. Implementer tier: Sonnet (single-file, mechanical, complete diagnosis) per Principle V rubric. No spec number claimed (exempt task).

PR #95 open (branch task-138-drift-evidence-shape), all four ACs proven in-branch; awaiting operator merge — the review session's merge attempt was blocked by the harness permission classifier, not by a gate.

Fix: scripts/check-merge-drift.mjs player-docs extractor now maps sources to .path. Live session run shows 26 readable, deduped paths and zero [object Object] (was 30 undeduped objects).

Adds scripts/check-merge-drift.test.mjs — the FIRST test for any scripts/*.mjs in this repo (node:test, zero deps; no prior convention existed). Run: node --test scripts/check-merge-drift.test.mjs

DEVIATION, gated and accepted: to test the real extractor rather than a copy, fingerprint() and computePlayerDocsSurface() are exported and the trailing main() call is guarded. The implementer's FIRST guard used process.argv[1] === new URL(import.meta.url).pathname, which silently disabled the gate entirely — import.meta.url resolves symlinks and argv[1] does not, so invoking through a symlinked ancestor produced NO output and exit 0. This machine reaches the repo at both ~/Claude/Code/promptworld and ~/projects/promptworld, so it was live, not theoretical. Rejected and replaced with a realpath comparison that DEFAULTS TO RUNNING when it cannot tell, plus a regression test that spawns the CLI through a symlink. Lesson worth keeping: the fix for a silently-suppressing gate nearly shipped a silently-skipping one.

Verification: 4/4 tests pass; reverting the fingerprint fix fails 3 with two IDENTICAL fingerprints for different findings; reverting the guard fails the symlink test on empty stdout. go build/vet clean. check-merge-drift pr = pass, no findings.
<!-- SECTION:NOTES:END -->
