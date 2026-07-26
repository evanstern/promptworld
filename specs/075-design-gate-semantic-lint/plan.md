# Implementation Plan: Design-gate semantic lint

**Branch**: `075-design-gate-semantic-lint` (task branch: `task-150-design-gate-semantic-lint`) | **Date**: 2026-07-26 | **Spec**: [spec.md](spec.md)

## Summary

One new check (`semantic-cells`) in `scripts/check-tui-design.mjs` — a
`status: shipped` page may not carry `unbuilt (wave` in a canonical
control-table renderer cell — plus the eight cell fixes that make the lint's
own PR pass: `overlays/postmortem.md` ×7 renamed to real, code-verified
symbols; `overlays/help.md` ×1 retagged to the live-owner form
(`unbuilt (pending TASK-142, layer-2)`, the `guardian-strip.md` precedent);
`panels/exercise.md` verified clean on the post-072 base. Red/green proof by
running the extended script before and after the fixes.

## Technical Context

**Language**: Node >= 18 ESM (the script; zero npm dependencies, read-only) +
Markdown design-corpus edits. **Testing**: the script itself is the test
vehicle — red run (exit 1, exactly 8 findings) then green run (exit 0);
`go test ./...` per doctrine though no Go changes. **Scope**:
`scripts/check-tui-design.mjs`, `docs/design/tui/overlays/postmortem.md`,
`docs/design/tui/overlays/help.md`, `docs/design/tui/panels/exercise.md`
(verification; edit only on residue). **Constraints**: no exit-code or output
shape changes; no severity tier; no symbol-existence check (spec D2/D3);
sequenced after spec 072's merge (shared pages).

## Constitution Check

- **I–III** — PASS: decision trail on TASK-150 + reorient synthesis + this
  spec; one task, one branch, one PR; the deliverable IS a gate, and its own
  PR is gated by it (spec FR-008 self-test) plus the merge-drift choke points.
- **IV** — PASS (planned): no `docs/wiki/` note lists the script or any
  touched page as a source (verified 2026-07-26, spec SC-005) — expected
  no-repin, no player-docs regeneration; both probes still run in-branch via
  the pr gate. Design-corpus pins: every page this branch edits re-verifies
  and re-pins `verified_against` to a branch commit — merge-commit-only
  (`gh pr merge --merge`) protects those branch hashes.
- **V** — PASS: implementation dispatches to `spec-implementer` on **Sonnet**
  (single-script lint extension + design-doc cell fixes — routine tier per the
  rubric; tier + justification already recorded on TASK-150's board notes).
  Planning model authors these artifacts only.

**Post-Phase-1 re-check**: PASS (no new violations; Complexity Tracking empty).

## Design

- **D1 — check placement**: new loop in `main()` after the `control-tables`
  check, iterating the already-built `parsed` map (every corpus page, not just
  panels/overlays — scope is by `status: shipped`, spec D1). Reuse
  `parseFrontmatter` for status and the existing header-detection idiom
  (`CANONICAL_HEADER`, the header + `/^\|[\s:|-]+\|$/` separator pair) to find
  canonical tables; then walk following lines while they are `|`-prefixed
  7-column rows, take column index 3 (renderer), trim, and flag rows whose
  cell contains the literal `unbuilt (wave`. Violation message carries
  1-based line number + cell text + remedy ("name the real renderer symbol,
  or retag to `unbuilt (pending TASK-<n>)` with a live owner").
- **D2 — conventions untouched**: `violate('semantic-cells', rel, msg)` rides
  the existing report/JSON/exit machinery; no flags added; runs in the
  structural pass so both bare and `--changed` invocations enforce it.
- **D3 — red run first**: before touching any page, run the extended script
  and capture the failing output — MUST be exactly 8 `semantic-cells`
  violations (7 postmortem, 1 help). This is SC-001's proof artifact; record
  it in implementation notes and the PR body.
- **D4 — postmortem.md cell fixes (by `control/region`, never line number)**:
  verify each symbol against code in-branch (this IS the page's re-verify),
  then rename:
  - *postmortem takeover* → `postmortemView` (`internal/tui/views.go`);
    trigger stays described by the data-source column.
  - *run-end narrated line* → the composing site inside `postmortemView`
    (implementer verifies the exact symbol/anchor in `views.go`); keep the
    shares-wording annotation.
  - *morgue evidence rows* → the morgue-row rendering inside `postmortemView`
    (implementer verifies); keep the content-pre-exists annotation.
  - *report card (scored runs only)* → `reportCardView` via the `reportCard`
    wrapper / `buildChecklistCard` (`internal/tui/reportcard.go`); keep the
    shared-with annotation.
  - *dismiss* → the postmortem's `esc` handling path (implementer verifies —
    the ceremony page's precedent is `handleTakeoverKey`,
    `internal/tui/tui.go`).
  - *replay via reopen key* → the `p` key handler (implementer verifies in
    `tui.go`).
  - *replay via reattach* → `Model.runEnded()` dual-source posture (symbol
    already verified; keep the reuses annotation).
  Every other column of every row stays as spec 072 left it. Re-pin
  `verified_against` to a branch commit.
- **D5 — help.md fix**: badge deep-link row renderer cell →
  `unbuilt (pending TASK-142, layer-2)`; hybrid-status prose paragraph
  updated to describe the pending-task form; re-pin.
- **D6 — exercise.md verification**: on the post-072 base,
  `grep -n "unbuilt" docs/design/tui/panels/exercise.md` — expect no stale
  "TASK-127, unbuilt" claim (spec 072 FR-010 amended it). Clean → file
  untouched, verification recorded in notes. Residue → correct by content,
  re-pin, record.
- **D7 — header contract line**: script header gains
  `// Contract (semantic-cells): specs/075-design-gate-semantic-lint/spec.md`
  alongside the existing 047 contract line; 047's contract file untouched.
- **D8 — gates, in order**: worktree cut only after spec 072's PR merges and
  root ff-pulls: `node scripts/check-merge-drift.mjs worktree --spec 075
  --task TASK-150`; claim re-touch already covered
  (`claim --dir 075-design-gate-semantic-lint` passes on own dir). In-branch
  before PR: green `node scripts/check-tui-design.mjs` and `--changed`;
  `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`
  (expected clean — `docs/wiki/` untouched); `go test ./...`;
  `node scripts/check-merge-drift.mjs pr` from the worktree. Merge with
  `gh pr merge --merge` (merge-commit-only; branch-hash pins). Post-merge:
  derived state only (spec-bridge sync, tasks.md ticks, runbook log).

## Complexity Tracking

Empty — no violations.
