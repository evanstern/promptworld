# Quickstart: validating the guardian strip (spec 050)

## Prerequisites

- `go build ./...`
- A running world (`promptworld ps`; any existing world).

## Automated validation

```bash
go test -race ./...
go test -race ./internal/tui -run 'Strip|RowBudget|Fold' -v
node scripts/check-tui-design.mjs --changed
```

Expected: all green. The render fixture sweep proves the segment presence
matrix (SC-002); the layout sweep proves row arithmetic + fold order across
terminal heights (SC-003).

## Manual validation (attached TUI)

1. `promptworld attach <world>` in a widescreen terminal.
   - One borderless line sits directly above the minibuffer: charge glyphs +
     `(N/cap)`, `next +1 @ <time>` (if bank below cap), `👁 N standing orders`.
   - No faith segment appears anywhere (TASK-118 unshipped).
2. Switch dock tabs (`2`/`3`/`4`) — the strip stays put on all of them
   (US1 AS-1).
3. Spend a charge (ask the guardian for an intervention) — bank and regen
   segments update next frame (US1 AS-2). With the bank at cap, confirm the
   regen segment is absent (US2 AS-2).
4. Shrink the terminal height until chrome folds — the strip row disappears
   LAST, and the minibuffer's dormant line gains the budget prefix; focus the
   minibuffer (`m`) — the focused input line is unchanged (US3 AS-1).
5. Narrow terminal (< 112 cols): the strip renders above the minibuffer
   exactly as widescreen (US3 AS-2).

## Re-ground checklist (after merge)

- `panels/guardian-strip.md` shipped + re-pinned; `patterns/layout.md`,
  `panels/minibuffer.md` re-verified/re-pinned in the merged PR.
- `/grounding-wiki:wiki-update` (touches `docs/wiki/tui-client.md` sources).
- `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`;
  regenerate player docs if stale (screen-orientation page gains the strip).
