# Feature Specification: Design-gate semantic lint — shipped pages cannot carry unbuilt renderer cells

**Feature Branch**: `075-design-gate-semantic-lint` (task branch: `task-150-design-gate-semantic-lint`)

**Created**: 2026-07-26

**Status**: Draft

**Input**: TASK-150 (reorient 2026-07-26 decision 2, merged position 2 face two —
the doc→code honesty face: `scripts/check-tui-design.mjs` validates pins and
table headers but never cell content, so `overlays/postmortem.md` shipped —
`status: shipped`, freshly pinned — while carrying seven `unbuilt (wave 4)`
renderer cells for renderers that exist and are tested).

## Grounding (pinned inventory, 2026-07-26)

The complete current inventory of `unbuilt (wave` **renderer cells** on
`status: shipped` pages (all 25 corpus pages are `shipped` today) is **eight**,
not seven — the survey found one beyond the board card's count:

1. `docs/design/tui/overlays/postmortem.md` — **seven** control-table rows
   whose renderer cell reads `unbuilt (wave 4)` (with varying annotations),
   identified by their `control/region` value, NOT by line number (the sibling
   spec 072 branch amends this file first and lines will drift):
   *postmortem takeover*, *run-end narrated line*, *morgue evidence rows*,
   *report card (scored runs only)*, *dismiss*, *replay via reopen key*,
   *replay via reattach*. Every one of these behaviors is backed by real,
   tested code: `postmortemView` (`internal/tui/views.go`, exercised in
   `internal/tui/render_test.go`), `reportCardView` via the `reportCard`
   consoleCard wrapper and `buildChecklistCard` (`internal/tui/reportcard.go`,
   `internal/tui/reportcard_test.go`), `Model.runEnded()` (the dual-source
   trigger, `internal/tui/exercise.go` et al.).
2. `docs/design/tui/overlays/help.md` — **one** row (*badge deep-link focus
   (layer-2)*) whose renderer cell reads `unbuilt (wave 4, layer-2)`. Unlike
   the postmortem cells this one is **truthful** — the badge deep-link is
   genuinely unbuilt and was folded into TASK-142's lane (reorient merged
   position 4 / board move 4). Its defect is the rotted *wave* marker, not the
   unbuilt claim.

Prose occurrences of `unbuilt (wave 4)` outside renderer cells exist
(`overlays/help.md` hybrid-status paragraph) and are NOT lint targets.

The corpus already holds the honest-deferral convention this lint must not
break: `panels/guardian-strip.md` carries a renderer cell reading
`unbuilt (pending TASK-118)` — an unbuilt claim tied to a **live task id**.
That form stays legal; the rot class is the *wave* marker: build-schedule
bookkeeping that goes stale the moment the wave ships, leaving either a lie
(renderer exists) or an orphaned pointer (feature slipped to another lane).

Coordination: sibling branch `task-149-report-card-truth` (spec 072) merges
BEFORE this task implements. Its FR-010 amends `overlays/postmortem.md`'s
known-simplification note and mockups, `overlays/ceremony.md`, and
`panels/exercise.md`'s stale "TASK-127, unbuilt" pointer (the board card's
"exercise.md:110"); its scope statement explicitly leaves the `unbuilt (wave`
cells to this task. This spec therefore specifies every fix **by content**,
never by line number.

## Decisions

**D1 — detection rule.** New check `semantic-cells` in
`scripts/check-tui-design.mjs`: on every page whose frontmatter is
`status: shipped`, every row of a canonical control table (the existing
7-column header the script already recognizes) whose **renderer column**
(column 4) contains the literal substring `unbuilt (wave` is a violation.
Scope is status-conditional, not directory-conditional: any corpus page a
future author marks `specified` may carry unbuilt cells freely; the moment it
flips to `shipped`, wave-marked cells block. `unbuilt (pending TASK-<n>)`
cells and prose occurrences are never flagged.

**D2 — fail, not warn.** The reorient decision says "warn/fail"; this spec
resolves it to **fail** (a violation, exit 1, like every existing check). The
script has no warning tier, and a non-blocking warning would not have
prevented the observed rot — the gate's value is at the PR choke point, where
blocking is the point. Adding a severity tier to a zero-dependency script for
one check is complexity without a customer.

**D3 — symbol-existence check: OUT.** The task's optional extension
("grep-level check that named renderer symbols exist in `internal/tui`") is
**not implemented**, decision recorded here. Grounds, from the full
renderer-column survey of the corpus:

- Renderer cells are free prose, not a symbol grammar: file paths
  (`cmd/promptworld/stages.go (unchanged, pre-existing)`), symbols legitimately
  outside `internal/tui` (`internal/mind/narrate.go` `chronicleNote`), skin
  tokens (`skin.CeremonyChapter`), method syntax (`Model.runEnded()`), struct
  fields (`lessonTriggers.Dismiss`), composite annotations ("existing quit
  path (unchanged)"), and cross-page references. Grep-level extraction of "the
  symbol" from these mis-flags honest cells at high rates.
- The observed failure class (face two: the doc claims unbuilt for built code)
  is fully covered by D1. The inverse class (a named symbol that does not
  exist) has zero observed instances across the 25-page corpus.
- If symbol verification is ever wanted, the right home is the mouse-parity
  sweep (reorient decision 8, spec 073's lane), which already commits to
  parsing control-table columns with a real cell grammar — not a grep bolted
  onto this small task.

**D4 — the fixes ride the same PR (self-test).** The lint must not fail its
own PR: all eight cells are amended in this branch, and
`node scripts/check-tui-design.mjs` (and `--changed`) exits 0 at the branch
tip. The seven postmortem cells are renamed to the **real renderer symbols
verified against code in-branch** (that verification IS the page's re-verify
step; `verified_against` re-pins to a branch commit). The help.md cell is
retagged from the wave form to the live-owner form —
`unbuilt (pending TASK-142, layer-2)` — following `guardian-strip.md`'s
existing precedent, and help.md's hybrid-status prose paragraph (which
describes that very cell's marking convention) is updated to match.

**D5 — contract residency.** The script's header cites spec 047's
`contracts/check-script.md`. The new check belongs to THIS spec: the header
comment gains a second contract line citing
`specs/075-design-gate-semantic-lint/spec.md`; spec 047's contract document is
not retro-edited (a spec dir is the source of truth for its own feature).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The gate catches semantic lies on shipped pages (Priority: P1)

An author (human or AI) marks a design page `status: shipped` while its
control table still carries `unbuilt (wave …)` renderer cells; the gate the
project already runs before every TUI-touching PR
(`node scripts/check-tui-design.mjs --changed`) blocks with a per-cell,
actionable finding instead of passing structurally.

**Why this priority**: this is board AC #1 and the reorient's converged
finding — "gates verify meaning, not just structure"; the reference lied while
the gate was green.

**Independent Test**: run the extended script on a tree that still carries the
current cells — it exits 1 reporting exactly the eight known violations, each
with file, line, and the offending cell.

**Acceptance Scenarios**:

1. **Given** the extended script and the pre-fix corpus, **When**
   `node scripts/check-tui-design.mjs` runs, **Then** it exits 1 and reports
   exactly 8 `semantic-cells` violations (7 in `overlays/postmortem.md`, 1 in
   `overlays/help.md`), each naming file + line + cell content and the remedy
   (name the real renderer symbol, or retag to `unbuilt (pending TASK-<n>)`).
2. **Given** a page with `status: specified` carrying `unbuilt (wave 4)`
   cells, **When** the script runs, **Then** no `semantic-cells` violation is
   reported for it.
3. **Given** `panels/guardian-strip.md`'s `unbuilt (pending TASK-118)` cell,
   **When** the script runs, **Then** it is not flagged.
4. **Given** `--json`, **Then** the violations ride the existing
   `{ ok, checks: [{check, status, file, message}] }` shape unchanged.

---

### User Story 2 - The corpus stops lying (Priority: P1)

An implementer reads `overlays/postmortem.md` to build on the postmortem
surface and finds renderer cells naming the real, tested symbols — not a
claim that seven shipped behaviors are unbuilt.

**Why this priority**: board AC #2; the false cells actively misdirect the
next implementer at the exact page the reorient calls the game's most salient
teaching surface.

**Independent Test**: `grep -rn "unbuilt (wave" docs/design/tui/` returns
nothing in any renderer cell; the extended script exits 0.

**Acceptance Scenarios**:

1. **Given** the amended `overlays/postmortem.md`, **Then** each of the seven
   rows names the real symbol(s) its behavior runs through — verified against
   code in-branch — preserving the rows' existing sharing/annotation content
   (e.g. the report-card row keeps its shared-with note), and
   `verified_against` is re-pinned to a branch commit.
2. **Given** the amended `overlays/help.md`, **Then** the badge deep-link row
   reads `unbuilt (pending TASK-142, layer-2)` (live owner, no wave marker),
   the hybrid-status prose paragraph describes the pending-task form, and
   `verified_against` is re-pinned.
3. **Given** `panels/exercise.md` on the post-072 base, **Then** it carries no
   stale "TASK-127, unbuilt" claim — expected already amended by spec 072
   (merges first); this branch VERIFIES that by content and corrects any
   residue, editing (and re-pinning) the file only if residue exists.
4. **Given** the branch tip, **Then** `node scripts/check-tui-design.mjs` and
   `node scripts/check-tui-design.mjs --changed` both exit 0 — the lint does
   not fail its own PR.

---

### Edge Cases

- **Truthful unbuilt on a shipped page** (the help.md case): legal via the
  `pending TASK-<n>` form — the lint bans the rotting wave marker, not
  honesty. A genuinely unbuilt sub-feature on a hybrid-status page points at
  its live owner instead.
- **Prose `unbuilt (wave`** outside a canonical control table: not flagged —
  only rows of a table whose header matches the canonical 7-column set are
  scanned. (help.md's prose paragraph is updated in this PR for consistency
  with its retagged cell, as authoring, not enforcement.)
- **Pages without control tables** (patterns/, INDEX.md, anatomy.md, pages/
  files without a canonical table): nothing to scan; no new requirement on
  them.
- **Non-canonical 7-column tables**: already a `control-tables` violation on
  panels/overlays; `semantic-cells` scans only canonical-headed tables and
  does not duplicate that finding.
- **Case/spacing**: literal substring match on the trimmed cell; the corpus
  authors the marker lowercase and the check stays dumb and predictable — no
  case folding, no regex classes.
- **Missing `status:`**: not this check's concern (frontmatter presence is the
  `pins` check's job); only the exact value `shipped` arms the scan.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `scripts/check-tui-design.mjs` MUST gain a `semantic-cells`
  check: for every corpus page whose frontmatter `status` is `shipped`, every
  data row of a canonical control table whose renderer column (column 4)
  contains the literal substring `unbuilt (wave` is a violation, reported with
  file, line number, the offending cell text, and the remedy.
- **FR-002**: the check runs in the structural pass (always — with and without
  `--changed`); exit-code (0/1/2), human, and `--json` output conventions are
  unchanged; the script stays read-only, Node >= 18 ESM, zero npm
  dependencies.
- **FR-003**: `unbuilt (pending TASK-<n>)` cells, prose occurrences, and pages
  whose `status` is not `shipped` MUST NOT be flagged.
- **FR-004**: no symbol-existence check is implemented (D3 — decision
  recorded; out of scope).
- **FR-005**: `overlays/postmortem.md`'s seven wave-marked renderer cells
  (identified by `control/region` value, D4's list) MUST be renamed to the
  real symbols verified against code in-branch, preserving each row's other
  columns and sharing annotations; the page's `verified_against` re-pins to a
  branch commit.
- **FR-006**: `overlays/help.md`'s badge deep-link renderer cell MUST be
  retagged to `unbuilt (pending TASK-142, layer-2)` and the hybrid-status
  prose paragraph updated to match; `verified_against` re-pins.
- **FR-007**: `panels/exercise.md` MUST be verified (by content) to carry no
  stale "TASK-127, unbuilt" claim on the post-072 base; residue, if any, is
  corrected and the page re-pinned — otherwise the file is untouched.
- **FR-008**: the extended script MUST exit 0 at the branch tip (self-test:
  the lint cannot fail its own PR), and MUST have been demonstrated to exit 1
  with exactly the eight expected findings before the cell fixes (red/green
  proof recorded in the implementer's notes/PR).
- **FR-009**: the script's header comment MUST additionally cite
  `specs/075-design-gate-semantic-lint/spec.md` as the contract for the new
  check (D5); spec 047's contract file is not modified.

### Key Entities

- **`semantic-cells` check** — new violation class in
  `scripts/check-tui-design.mjs`, alongside `file-set`, `pins`,
  `control-tables`, `anatomy`, `same-pr`.
- **Canonical control table** — the existing 7-column header constant
  (`CANONICAL_HEADER`); the new check reuses the script's table detection, and
  scans data rows following a canonical header + separator.
- **The eight cells** — Grounding items 1–2; the fix inventory.

## Success Criteria *(mandatory)*

- **SC-001** (board AC #1): the red-run proof — the extended check on the
  pre-fix tree reports exactly the eight known violations and exits 1;
  recorded as an artifact (implementation notes / PR body).
- **SC-002** (board AC #2): `overlays/postmortem.md`'s seven cells name real,
  code-verified symbols; `panels/exercise.md` verified clean of the stale
  TASK-127 unbuilt claim on the post-072 base; `overlays/help.md` retagged —
  all in this one PR.
- **SC-003**: `node scripts/check-tui-design.mjs` and `--changed` exit 0 at
  the branch tip; `node scripts/check-merge-drift.mjs pr` passes from the
  worktree.
- **SC-004**: `go test ./...` green (a JS + docs change should not move it;
  run anyway per doctrine).
- **SC-005**: no wiki re-pin expected — no `docs/wiki/` note lists
  `scripts/check-tui-design.mjs` or any touched page as a source (verified
  2026-07-26); the pr gate's wiki and player-docs probes still run in-branch
  and must pass.

## Assumptions

- Spec 072 (`task-149-report-card-truth`) merges before this branch cuts; the
  branch forks from an `origin/main` that already carries its postmortem.md /
  ceremony.md / exercise.md amendments. If exercise.md's stale pointer somehow
  survives, FR-007's residue clause covers it — no coordination stall.
- All 25 corpus pages are `status: shipped` today, so the status-conditional
  scope currently equals whole-corpus scope; the conditionality exists for
  future `specified` pages, not present behavior.
- Amending only `docs/design/tui/` and `scripts/` keeps the `same-pr` check
  trivially green (no `internal/tui/` files change).

## Out of scope

- **Stale ownership pointers in design prose** (e.g. "owned by TASK-119" for
  a Done task) — the reorient's recorded third residue class; a review
  responsibility, not mechanically caught by this lint (merged position 2's
  scope note). Recorded here so the boundary is deliberate.
- Symbol-existence verification (D3).
- Any warning/severity tier for the check script (D2).
- `internal/tui/` code changes of any kind.
