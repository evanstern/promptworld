# Phase 0 Research: TUI Design Reference v2

**Feature**: 047-tui-design-reference-v2 | **Date**: 2026-07-25

All Technical Context unknowns resolved. The three Wave 0 rulings are derived HERE, at
the planning tier — the implementer authors them into pages; they do not re-derive them.

## R1. Reconciliation grounding source

**Decision**: reconcile against `docs/wiki/tui-client.md` (387 lines, pinned at
723c464) as the primary account of shipped reality, with spot-verification reads of
`internal/tui/*.go`; the per-surface introducing specs fill the control tables'
`introduced-by` column.

**Rationale**: the wiki note is the de-facto current reference (accurate, pinned,
organized by code file); re-deriving shipped behavior from Go source alone would repeat
work the wiki already gates. Code reads confirm, not discover.

**Alternatives considered**: (a) code-only reconciliation — slower, re-derives what the
wiki proves; (b) spec-by-spec replay of 013–046 — most specs don't touch the TUI;
the TUI-relevant set is enumerable: **015** (villagers tab), **018** (chronicle
digest), **020** (decision trace view), **021** (instruction surface), **024**
(provider table), **028/031/033** (throttle/adoption/debt surfaces), **029** (standing
orders block), **034** (provider-health badge + rows), **035** (calibration UX),
**037** (horizon rows + suppressed badge), **039** (speed posture), **044** (ENDED
token, morgue), **045** (help overlay), **046** (stage segment, ladder). This set
seeds the staleness sweep checklist in tasks.md.

## R2. Ruling (a) — bottom-chrome row budget and fold order

**Decision**: the widescreen row budget at full stage-1–2 chrome is:

```
totalRows
├─ header:         1
├─ villager strip: 1   (D12; widescreen default-on, all stages)
├─ body:           remainder (map ∥ dock; both full body height)
├─ lesson row:     2   (decision 5; default-on at stages 1–2 only —
│                       0 rows when stage-defaulted off; ≤2 lines, fixed 2 when on)
├─ guardian strip: 1   (decision 7; always-on, all stages)
├─ minibuffer:     3   (unchanged; bordered single line)
└─ footer:         1
```

Fixed chrome at stages 1–2: 9 rows (was 5) → body = totalRows − 9; a 30-row terminal
keeps 21 body rows, a 24-row terminal 15. **Fold order**: chrome folds when body rows
would drop below `bodyMin = 10`, reclaiming in this total order, one step at a time,
until body ≥ 10 or the floor is reached:

1. map legend (existing body-internal shed — stays first)
2. villager strip → folds into a header count badge (e.g. `[12 villagers]` segment)
3. lesson row → folds to a header badge (`[lesson]`); content remains via `?` (pull)
4. guardian strip → folds its content into the minibuffer's dormant line (budget text
   prefixes the dormant hint) — folds LAST because decision 7 says the budget is
   always visible; the fold keeps it visible, one row cheaper.

**Floor layout**: header + body(≥10) + minibuffer(3) + footer — the pre-reorientation
stack. Terminals too short even for that (< 15 rows) keep the existing behavior
(panels shed lowest-priority rows; no new rule).

**Rationale**: folding in inverse order of doctrine strength — the legend is already
sheddable; the villager strip is Wave 5 glanceability (weakest claim); the lesson row
has an explicit designed fallback (badge+overlay is its own stage-3+ default, so the
fold reuses a designed state rather than inventing one); the guardian strip is the only
element a decision says is *always* visible, so it never disappears — it relocates.
`bodyMin = 10` keeps the dock's smallest useful tab (villagers roster header + a few
rows) and the map viewport genuinely usable; below that the composite was already
degenerate pre-reorientation.

**Alternatives considered**: fixed height breakpoints (e.g. fold at < 28 rows) —
rejected: the budget-driven rule (`bodyMin`) adapts to any chrome combination and
matches the existing "rows get scarce → shed lowest-priority" doctrine in layout.md;
hiding the guardian strip entirely at the floor — rejected: contradicts decision 7.

## R3. Ruling (b) — narrow-fallback behavior for the new chrome

**Decision**: in the narrow (< 112 cols) single-pane layout:

- **guardian strip**: carried — 1 row above the minibuffer, identical content
  (decision 7's "always visible" is width-independent).
- **lesson row**: carried with the same stage defaults as widescreen (on at 1–2,
  badge+overlay at 3+/pre-ladder); same fold rule as R2 applies against `bodyMin`.
- **villager strip**: NOT carried — narrow shows the header count badge form; the
  villagers solo view is the drill-down.
- **guardian console / systems tab / exercise panel**: reachable as solo views (the
  existing narrow pattern — solo-views.md); no new narrow-specific rendering.
- **ceremony / postmortem**: take over the full screen in narrow exactly as in
  widescreen (takeovers are layout-independent); linear-stream projections (D1) are
  unaffected.

**Rationale**: narrow's contract is "today's single-pane UI, never deleted" — additive
chrome must justify each row it takes from an already-short layout. The strip and
lesson row carry doctrine (decisions 5/7); the villager strip carries none in narrow
(its value is glanceability *beside* the map, which narrow doesn't render).

**Alternatives considered**: overlay-only lesson delivery in narrow — rejected: narrow
terminals plausibly host new players (SSH sessions); stages 1–2 is exactly where
pushed delivery matters.

## R4. Ruling (c) — help overlay byte-identity classification

**Decision**: `overlays/help.md` classifies every section explicitly:

| Section | Class |
|---|---|
| keys (tiered) | **byte-identical** with nil status (generated from the static keymap; the existing `help_test.go` sweep guarantee) |
| screen walkthrough | **byte-identical** with nil status |
| guardian section (D9) | **stage-keyed, model-free**: content is a pure function of the stage value; for a given stage the bytes are constant; nil status renders the pre-ladder variant (all verbs). Never LLM-derived. |
| lessons registry | **status-derived** (active/seen state per user); the registry's catalog text is static, its state columns are live |
| badge deep-link focus | **status-derived** (which row is pre-focused depends on active badges); content unchanged |
| ceremony replay entries | **status-derived** (which ceremonies exist depends on run history); replayed content is stored, not regenerated |

The no-LLM floor guarantee restated: with nil status AND no LLM configured, the
overlay renders the keys, walkthrough, and pre-ladder guardian sections byte-identically
on every invocation — spec 045's contract, deliberately amended (per D9) from
"never derived from live status" to "the floor set is byte-identical; status-derived
sections degrade to their static catalog with state columns empty".

**Rationale**: preserves what spec 045's invariant actually protects (deterministic,
model-free help always available) while admitting the reorientation's dynamic
additions under an explicit classification instead of eroding the invariant silently.

## R5. Check script mechanics

**Decision**: `scripts/check-tui-design.mjs`, Node ≥ 18 ESM, zero npm deps, read-only,
exit codes 0/1/2 (clean / violations / usage-or-env error) — the `check-freshness.mjs`
contract shape. Checks, in order:

1. **File-set**: every file under `docs/design/tui/` matches the taxonomy (INDEX,
   anatomy, pages/, panels/, overlays/, patterns/); no orphan classes.
2. **Pins**: every `.md` carries YAML frontmatter with `verified_against: <40-hex>`
   resolving to a commit known to the repo.
3. **Control tables**: every `panels/*` and `overlays/*` page contains exactly one
   table whose header row is the canonical column set (contracts/control-table.md).
4. **Anatomy completeness**: every reference file is reachable from `anatomy.md` (no
   unmapped pages), and every anatomy row's target file exists.
5. **Same-PR gate** (`--changed [range]`, default `origin/main...HEAD`): if the range
   touches `internal/tui/` and no path under `docs/design/tui/` is touched in the same
   range, fail, naming the touched TUI files. A design-doc-only change never fails
   this check.

**Rationale**: each check is mechanically decidable from file structure + git — no
semantic judgment, so no false authority. The same-PR rule extends INDEX rule 4 from
"record deviations" to "any change amends the reference" exactly as the reorientation
prescribes; range-diff detection is the strongest enforcement available in a repo with
no CI, and matches how the project actually gates (documented scripts + session gates).

**Alternatives considered**: pin-vs-source-diff staleness (like the wiki's model:
fail when pinned commit's `internal/tui` differs from HEAD's) — rejected for v1: the
reference pins *pages to verification moments*, not pages to source files (a page like
`patterns/stage-defaults.md` describes an unbuilt surface with no source yet); the
same-PR rule covers the drift vector the convention actually failed on (specs 044/046
landing TUI changes with no doc amendment). Revisitable once all surfaces are built.
Git pre-commit hook — rejected: worktree-hostile and bypassable; the project's gate
culture is session/PR-level scripts.

## R6. Taxonomy split mechanics

**Decision**: `panels/dock.md` survives as the tab-*container* page only (tab row,
badges, tab-switching keys, solo-zoom seam); tab content moves to `panels/guardian.md`
(fiction layer — the skin boundary file, D10), `panels/systems.md` (telemetry — never
skinned), `panels/villagers.md`, with `panels/chronicle.md` already separate. The help
overlay section of `patterns/keymap.md` moves to `overlays/help.md`; keymap.md stays
the one printable reference card and gains the decision-8 parity rule.

**Rationale**: the analyses' taxonomy verbatim; the container/content split keeps
exactly one owner per visible element (anatomy's invariant).

## R7. Skin-token documentation conventions (ahead of TASK-121)

**Decision**: `patterns/skin-tokens.md` defines the *documentation* convention only:
fiction strings render as `{{skin.<domain>.<name>}}` placeholders in mockups and as
`skin.<domain>.<name>` in control-table skin-token columns, with the default-skin
(angel fiction) value listed beside each token in that page's token index table.
Non-fiction strings use the literal and `—` in the skin-token column. The runtime
lookup contract (file format, resolution, fallback) is explicitly deferred to
TASK-121's spec, which MUST adopt or amend this page in its own PR.

**Rationale**: satisfies D2's sequencing (the doc never writes new bare fiction
literals) without pre-empting TASK-121's design authority over the runtime contract.

## R8. Implementation tier

**Decision**: Sonnet (default tier), via the `spec-implementer` agent.

**Rationale (rubric)**: the work is doc reconciliation and authoring plus one
single-file zero-dep Node script — the rubric's routine tier by name ("doc
reconciliation", single-package tooling). All judgment-heavy design (the three rulings,
taxonomy, script semantics, operator questions) is resolved in spec/research at the
planning tier, so the implementing slices are execution against fixed decisions.
Escalation to Opus 4.8 only if a slice fails gates (one-way, recorded on TASK-123).
