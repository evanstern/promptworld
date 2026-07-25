# Data Model: TUI Design Reference v2

**Feature**: 047-tui-design-reference-v2 | **Date**: 2026-07-25

The "data" here is the reference corpus's structure — the entities the check script
parses and the acceptance sweeps audit.

## Reference page

One Markdown file under `docs/design/tui/`, owning exactly one page, panel, overlay,
or pattern.

| Field | Where | Rule |
|---|---|---|
| class | path | one of `pages/`, `panels/`, `overlays/`, `patterns/`, or top-level (`INDEX.md`, `anatomy.md` only) |
| frontmatter | YAML block at top | required on every `.md`; schema in contracts/frontmatter-and-pins.md |
| `verified_against` | frontmatter | 40-hex commit; must resolve in the repo |
| `status` | frontmatter | `shipped` (describes built surface) \| `specified` (spec-before-build page, no code yet) |
| mockup | body | required on pages/panels/overlays for visual surfaces: fenced ASCII block + per-callout prose |
| control table | body | required on every `panels/*` and `overlays/*` page: exactly one, canonical header (contracts/control-table.md) |
| linear projection | body | required section on every new-surface page: the `attach`/`tail`/CLI equivalent (D1) |
| stage defaults | body | required on stage-shaped surfaces: default visibility per stage + pre-ladder |

**State transitions**: `specified` → `shipped` when the implementing wave lands; that
PR amends the page (same-PR gate) — flipping `status`, filling renderer symbols, and
re-pinning `verified_against`.

## Control table row

The AI-parseable unit. Columns (canonical, ordered — grammar in
contracts/control-table.md):

1. `control/region` — the visible element
2. `states` — enumerated visual/behavioral states
3. `data source` — Status field / event type / replica path (plain-language by
   default; raw registry names allowed here — this column is engineer-facing; the
   player-facing projection stays plain-language per FR-020's toggle ruling)
4. `renderer` — Go symbol; `unbuilt (wave N)` on `specified` pages
5. `keys+mouse` — keyboard binding(s) `·` mouse target; `—` if display-only
6. `introduced-by` — spec NNN / TASK-N / reorient decision id
7. `skin-token` — `skin.<domain>.<name>` or `—` (never a bare fiction literal)

## Anatomy index entry (`anatomy.md`)

One row per visible screen element.

| Field | Rule |
|---|---|
| region | named visible element (header segment, badge, strip, panel, overlay…) |
| owning file | relative path; must exist; every panel/overlay/page file must be the target of ≥1 row (completeness both directions) |
| notes | stage-default visibility, fold behavior reference |

**Invariant**: exactly one owning file per element; zero unmapped visible elements
(SC-001); zero unreferenced reference files.

## Check report (script output)

| Field | Rule |
|---|---|
| check | one of `file-set`, `pins`, `control-tables`, `anatomy`, `same-pr` |
| status | `ok` \| `violation` |
| detail | file + actionable message ("amend docs/design/tui/… in this PR", "pin missing on …") |
| exit code | 0 all ok · 1 ≥1 violation · 2 usage/env error |

`--json` emits the report as JSON (machine-readable, mirrors check-freshness.mjs).

## Ruling records

Not entities with schemas, but fixed content with a required home:

| Ruling | Home | Content (from research.md) |
|---|---|---|
| (a) row budget + fold order | `patterns/layout.md` | 9-row stage-1–2 chrome stack; `bodyMin = 10`; fold order legend → villager strip → lesson row → guardian strip (relocates, never disappears); floor layout |
| (b) narrow fallback | `patterns/layout.md` + each new-surface page | strip carried, lesson row carried w/ stage defaults, villager strip badge-only, new pages solo-view, takeovers full-screen |
| (c) byte-identity | `overlays/help.md` | section classification table; restated floor guarantee (spec 045 contract amended per D9) |
| FR-018 ambient postmortem | `overlays/postmortem.md` | morgue evidence only; report card in scored runs only |
| FR-019 ceremony voice | `overlays/ceremony.md` | both voices; instrument (rubric) authoritative |
| FR-020 audience | `patterns/…` (control-table conventions) + `overlays/help.md` | plain-language default; raw registry values behind explicit debug/inspector toggle |
