# Quickstart: validating TUI Design Reference v2

**Feature**: 047-tui-design-reference-v2. Run everything from the worktree root.

## Prerequisites

- Node ≥ 18 (`node --version`)
- A checkout with the feature branch (`.worktrees/task-123`)

## 1. Structural checks pass on the landed tree (SC-003 clean half)

```bash
node scripts/check-tui-design.mjs
echo $?   # expect 0
```

## 2. Each violation class fails actionably (SC-003 seeded half)

Seed each violation in a scratch commit (or stash-revert cycle), run the script,
assert exit 1 and a message naming the file:

| Seed | Expected violation |
|---|---|
| delete `verified_against` from one page | `pins` |
| change one control-table header column name | `control-tables` |
| add `docs/design/tui/panels/orphan.md` with valid frontmatter but no anatomy row | `anatomy` |
| add `docs/design/tui/rogue.md` (top-level, class `panel`) | `file-set` |
| commit touching `internal/tui/tui.go` only, then `--changed HEAD~1...HEAD` | `same-pr` |

Then restore the tree and re-run check 1.

## 3. Same-PR gate ignores doc-only changes

```bash
# on a branch where only docs/design/tui/ changed:
node scripts/check-tui-design.mjs --changed origin/main...HEAD
echo $?   # expect 0
```

## 4. Corpus acceptance sweeps

- **SC-001 (coverage)**: for each surface in the R1 staleness set (specs 015, 018,
  020, 021, 024, 028/031/033, 029, 034, 035, 037, 039, 044, 045, 046), find its
  documentation via `anatomy.md` → owning page; confirm the page describes the
  shipped rendering per `docs/wiki/tui-client.md`. Zero misses.
- **SC-002 (tables/tokens)**: `node scripts/check-tui-design.mjs` covers table
  presence/shape mechanically; additionally
  `grep -rn "Metatron" docs/design/tui/ --include="*.md"` — hits allowed only inside
  `patterns/skin-tokens.md`'s default-skin value column and `INDEX.md`'s verbatim
  TASK-34 history blockquote (fiction appears only as tokens elsewhere).
- **SC-004 (ten pages)**: confirm all ten new-surface pages/sections exist with
  mockup + control table + stage defaults + linear-stream projection:
  `pages/guardian-console.md`, `panels/systems.md`, `panels/exercise.md`,
  `overlays/ceremony.md`, `overlays/postmortem.md`, `panels/lesson-row.md`,
  `panels/guardian-strip.md`, `panels/villager-strip.md`,
  `patterns/stage-defaults.md`, and the guardian section inside `overlays/help.md`.
- **SC-005 (2 hops)**: pick 5 arbitrary controls (e.g. provider row, stage segment,
  suppression row, tab badge, dormant minibuffer); from `INDEX.md` reach each's
  authoritative page in ≤ 2 hops (INDEX → anatomy → page, or INDEX → page).
- **SC-006 (rulings)**: in `patterns/layout.md`, verify the row arithmetic sums to
  terminal height at stage 1–2 and stage 3+ defaults and the fold order is total
  with a stated floor; every new-surface page states its narrow behavior; every
  section listed in `overlays/help.md` is classified byte-identical vs
  status-derived per research.md R4.

## 5. Regression

```bash
go test ./...            # untouched code stays green
node .claude/skills/player-docs/scripts/check-freshness.mjs --check   # wiki/player-docs unaffected
```
