# Feature Specification: Mouse-parity sweep test — control tables become a gate

**Feature Branch**: `073-mouse-parity-sweep` (task branch: `task-154-mouse-parity-sweep`)

**Created**: 2026-07-26

**Status**: Draft

**Input**: TASK-154 (reorient 2026-07-26 decision 8, merged position 6 —
"verification culture, extended": convert input-parity doctrine into a gate,
the project's house style). Board ACs: (1) a test parses control tables and
fails on any non-'—' mouse cell without a handler; (2) `patterns/keymap.md`'s
rollout note updated to reflect mechanized tracking.

## Grounding (verified against the working tree, 2026-07-26)

**The documented mouse surface.** Every page in `docs/design/tui/` that
carries a control table uses the canonical header
(`specs/047-tui-design-reference-v2/contracts/control-table.md`, byte-asserted
by `scripts/check-tui-design.mjs`):

```
| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
```

The `keys+mouse` column's contracted grammar is `<keys> · <mouse>`, with `—`
for a display-only control. 18 files carry the header today — across
`panels/`, `overlays/`, **and `pages/`** (the contract text names only
panels/overlays; `pages/home.md`, `pages/solo-views.md`,
`pages/guardian-console.md` carry tables too, so the sweep walks the whole
corpus and keys on the header, not the directory). A full cell inventory
shows exactly four cell shapes in the wild:

1. `—`, `— (display-only)`, `— (auto-follow)` — no ` · ` separator; the cell
   starts with `—`: display-only, no input claim at all.
2. `` `r` · — `` (and variants with prose in the keys half, e.g.
   `` `5` select · — ``, `any key (consumed, one press, this tab only) · —``)
   — keyed, mouse half exactly `—`: a **tracked parity gap**, honest per
   contract rule 4, listed in the page's `**Parity rollout**` note.
3. `` `1`/`esc` (home) · a different tab key (switch) · — `` (dock.md) — the
   keys half itself contains ` · `; the mouse half is the **last**
   ` · `-separated segment.
4. `` `⏎` · click line `` (chronicle.md jump-to-source) — the corpus's ONE
   non-'—' mouse cell today: a shipped mouse claim.

**"Planned" vs `—` (edge case, resolved from the artifacts):** the corpus has
no in-cell "planned" marker. A control whose mouse action is legitimately
unbuilt-yet-documented is expressed as mouse half `—` **plus** an entry in the
page's `**Parity rollout**` note (contract rule 4; keymap.md doctrine rule 3).
All 18 table-bearing pages already carry that note. Therefore: any non-'—'
mouse half is a *shipped* claim and must have a real handler — there is no
third state for the sweep to guess about.

**The real mouse surface.** `internal/tui/tui.go:577` routes `tea.MouseMsg`
to `Model.handleMouse` (`tui.go:1688`), which implements exactly one
behavior: left-button release inside the chronicle's last-rendered inspect
rows (hit region `m.chronHit`) selects that row and applies `jumpToSource` —
the code twin of the one documented claim. Test helpers already exist:
`mouseLeftRelease(x, y)` (`internal/tui/tui_test.go:81`),
`widescreenModel(t)`, and live mouse-dispatch tests around
`tui_test.go:872`.

**Precedents.** `TestStageDefaultsSweep`
(`internal/tui/stagedefaults_test.go:124`) parses a design-corpus page from
`../../docs/design/tui/` and compares it against a code table, both
directions. `TestHelpKeymapSweep` (`internal/tui/help_test.go:317`) compares
a rendered surface against a hand-audited in-test oracle (`realModeKeys`) and
pairs it with `TestHelpKeymapSweepLiveDispatch`, which proves representative
entries against the *running* dispatch. This feature composes the two shapes.

## Decision — what counts as handler evidence

A prose mouse claim like `click line` is not machine-mappable to a Go symbol,
so "a case exists in `handleMouse`" cannot be asserted from the string alone.
Handler evidence is therefore **two-layered**, mirroring the keymap-sweep
precedent:

1. **Oracle membership.** The test file carries a hand-audited oracle table
   mapping each `(page, control/region, mouse claim)` triple to a live-dispatch
   check. The sweep fails when the corpus carries a non-'—' mouse cell with no
   oracle entry (a mouse target documented without proof), and when the oracle
   carries an entry the corpus no longer documents (a stale proof).
2. **Live dispatch.** Each oracle entry is executable: it builds a `Model` in
   the documented state, sends the real `tea.MouseMsg` the claim describes
   through `Model.Update`, and asserts the documented observable effect. A
   static list alone never satisfies the sweep — the handler must demonstrably
   act.

Rejected alternative: grepping `internal/tui` source for handler symbols named
in the renderer column — the renderer column names *renderers*, not input
handlers, and string-matching source text proves presence, not behavior.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The control tables gate mouse parity (Priority: P1)

A developer documents a mouse target in any control table (or ships one in
code) — `go test ./internal/tui/` fails unless the claim and a passing
live-dispatch proof land together.

**Why this priority**: this is decision 8's conversion of doctrine into a
gate — board AC #1; without it the `keys+mouse` column can silently drift
from the code, exactly the dishonesty the doctrine forbids.

**Independent Test**: run the new test against the current corpus (must pass:
the one claim, jump-to-source, has a real handler); mutate a copy of a table
cell from `—` to a fake mouse claim and confirm the sweep fails with a
message naming the page and control.

**Acceptance Scenarios**:

1. **Given** the shipped corpus and code, **When**
   `go test ./internal/tui/ -run TestMouseParity` runs, **Then** it passes,
   having parsed every canonical-header table in `docs/design/tui/` and
   verified the chronicle jump-to-source claim by live dispatch.
2. **Given** a control table gains a non-'—' mouse cell with no oracle entry,
   **When** the sweep runs, **Then** it fails naming the page file and the
   control/region.
3. **Given** the oracle carries an entry whose corpus cell was removed or
   reverted to `—`, **When** the sweep runs, **Then** it fails naming the
   stale entry.
4. **Given** a page whose table has a keyed-but-mouseless row
   (`<keys> · —`) but no `**Parity rollout**` note, **When** the sweep runs,
   **Then** it fails naming the page (contract rule 4's honesty requirement,
   mechanized).

---

### User Story 2 - The doctrine page says tracking is mechanized (Priority: P2)

A reader of `docs/design/tui/patterns/keymap.md`'s input-parity doctrine
(rule 3, "Rollout is incremental, honestly tracked") learns that the tracking
is now enforced by a named test, and what graduating a control requires
(real mouse cell + oracle entry + live proof, same PR).

**Why this priority**: board AC #2; the rollout note is the doctrine's
human-facing face and currently describes hand-tracking only.

**Independent Test**: read the amended rule 3 — it names the test and states
the graduation mechanics; `node scripts/check-tui-design.mjs --changed`
passes with the page re-verified and re-pinned on this branch.

**Acceptance Scenarios**:

1. **Given** the amended `patterns/keymap.md`, **When** rule 3 is read,
   **Then** it states that non-'—' mouse cells are enforced by the sweep test
   (named, with its file path) and that parity-rollout notes are asserted to
   exist wherever a table carries a keyed-but-mouseless row.
2. **Given** the amendment, **When** `check-tui-design.mjs --changed` runs on
   the branch, **Then** the page's `verified_against` pin is a branch commit
   and the gate passes.

---

### Edge Cases

- **Annotated dashes** (`— (display-only)`, `— (auto-follow)`): cells with no
  ` · ` separator whose text starts with `—` are display-only — skipped, never
  parsed as claims.
- **Multiple ` · ` separators in one cell** (dock.md row): the mouse half is
  the last segment; everything before it is the keys half.
- **Multi-action mouse cells** (none exist today): a future mouse half naming
  several actions (e.g. `click row / wheel scroll`) is ONE claim string keyed
  as-is in the oracle; its single check must prove every named action. The
  sweep does not invent a sub-grammar the contract doesn't define.
- **Tables outside panels/** (pages/, overlays/, future patterns/): the sweep
  keys on the canonical header anywhere under `docs/design/tui/`, so a table
  added to a new directory is swept automatically. Non-canonical tables
  (keymap.md's mode tables, token indexes) never match the header — contract
  rule 1 guarantees the header is not reused — so they are naturally ignored.
- **`·` inside other columns** (states column, renderer column): irrelevant —
  the parser splits the row on `|` first and only inspects column 5.
- **Corpus goes missing** (test run from a stripped checkout): the test fails
  loudly on the unreadable corpus root rather than passing vacuously
  (`TestStageDefaultsSweep` precedent: `t.Fatalf` on read error).
- **Zero claims parsed**: the sweep asserts it found ≥ 1 non-'—' claim and
  ≥ 1 control table overall — a refactor that silently breaks the parser or
  moves the corpus cannot pass as an empty sweep.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: A new Go test in `internal/tui` MUST parse every markdown table
  under `docs/design/tui/` whose header row equals the canonical control-table
  header, extract each row's `keys+mouse` cell (column 5), and classify it:
  display-only (cell starts with `—`, no ` · `), tracked gap (mouse half —
  the last ` · `-separated segment — exactly `—`), or shipped mouse claim
  (anything else).
- **FR-002**: For every shipped mouse claim, the test MUST require an entry in
  an in-test oracle keyed by (corpus-relative page path, control/region,
  mouse-claim string), and MUST fail with the page + control named when none
  exists.
- **FR-003**: Every oracle entry MUST carry an executable check that sends the
  claimed mouse event(s) as real `tea.MouseMsg` values through `Model.Update`
  and asserts the documented effect; the sweep runs every check. The initial
  oracle has exactly one entry: chronicle jump-to-source (`click line` →
  row selection + `jumpToSource`, per `panels/chronicle.md` and
  `contract §1` in `handleMouse`'s doc comment).
- **FR-004**: The test MUST fail on stale oracle entries — an oracle key with
  no matching shipped claim in the parsed corpus (both directions, precedent
  `TestStageDefaultsSweep`).
- **FR-005**: The test MUST fail when a parsed table contains ≥ 1 tracked-gap
  row but the page has no `**Parity rollout**` marker (contract rule 4).
- **FR-006**: The test MUST fail (not skip) when the corpus root is
  unreadable, and MUST assert non-vacuity (≥ 1 table parsed, ≥ 1 shipped
  claim found).
- **FR-007**: `docs/design/tui/patterns/keymap.md` doctrine rule 3 MUST be
  amended to state the tracking is mechanized: name the test and its file, and
  state the graduation contract (a control leaves a rollout note by gaining a
  real mouse cell + oracle entry + passing live proof in the same PR). No
  other doctrine rule changes.
- **FR-008**: No new mouse features, no `internal/tui` non-test code changes,
  no control-table cell changes — this feature gates the existing surface
  only.

### Key Entities

- **Control table** — canonical-header markdown table; column 5 =
  `keys+mouse` (`<keys> · <mouse>` | `—`).
- **Shipped mouse claim** — a non-'—' mouse half; today exactly one:
  `panels/chronicle.md` jump-to-source, `click line`.
- **Oracle** — in-test map from claim triple to live-dispatch check
  (`realModeKeys` / `stageDefaultsTable` precedent shape).
- **Parity rollout note** — the `**Parity rollout**` prose block each
  table-bearing page carries for its keyed-but-mouseless controls.

## Success Criteria *(mandatory)*

- **SC-001**: Board AC #1 — the sweep test exists, parses the corpus control
  tables, and fails on any non-'—' mouse cell without a proven handler;
  demonstrated by the mutation check in US1's independent test.
- **SC-002**: Board AC #2 — keymap.md rule 3 reflects mechanized tracking;
  `node scripts/check-tui-design.mjs --changed` passes on the branch with the
  page re-pinned.
- **SC-003**: `go test ./...` green; the new test passes against the shipped
  corpus with no code or table changes beyond FR-007's prose amendment.
- **SC-004**: The merge-drift pr gate passes from the worktree (including
  wiki-repin and player-docs probes — expected clean: no wiki note pins the
  touched files; verified below).

## Assumptions

- No wiki note pins `docs/design/tui/patterns/keymap.md` or the (new) test
  file as a source (verified: `grep "^  - docs/design/tui" docs/wiki/*.md` is
  empty; a new file cannot be pinned) — so the pr gate's `wiki-repin-missing`
  and `player-docs-stale` probes are expected clean. The gate remains the
  authority; if it disagrees, produce the re-pin, don't argue.
- The canonical header stays byte-stable — it is already asserted by
  `check-tui-design.mjs` (`CANONICAL_HEADER`), so the Go parser and the JS
  gate drift together or not at all.
- `mouseLeftRelease` / `widescreenModel` test helpers remain available in
  package `tui`'s test files for the live-dispatch check to reuse.
