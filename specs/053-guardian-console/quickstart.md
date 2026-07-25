# Quickstart: validating the guardian console + systems split (spec 053)

## Prerequisites

- `go build ./...`; a running world with llm.json (for systems content) and
  one no-LLM world (for absence honesty).

## Automated validation

```bash
go test -race ./...
go test -race ./internal/tui -run 'Console|Systems|Editor|Focus' -v
node scripts/check-tui-design.mjs --changed
```

## Manual validation

1. Attach widescreen; press `G` — full-height console: header, labeled turn
   blocks, charter/skills panel, minibuffer beneath, footer. `G` again —
   back exactly where you were (repeat from a solo view and from narrow).
2. `m`, ask something — turn appears in the stream; busy state renders in
   the composer as everywhere else.
3. Press `5` — systems tab: provider table, health lines, spend/wallet,
   horizon block. Press `3` — guardian tab: transcript/orders only, NO
   telemetry. `5` twice — systems solo zoom. Narrow terminal — systems
   reachable as a pane.
4. On the console press `e` — $EDITOR opens charter.md; edit, save, quit —
   "charter changed — next turn binds it" renders once. Re-enter `e`, quit
   without changes — no confirmation. `EDITOR= promptworld attach …` and
   `e` — honest hint.
5. Stage-1 tutor world: read surface shows the preset lock + skills locked
   notices naming the unlocking stages.
6. No-LLM world: systems tab shows no horizon block, states content honestly.

## Re-ground checklist (after merge)

- Both new-surface pages `shipped` + all touched pages re-pinned (in-PR).
- `/grounding-wiki:wiki-update` (tui-client.md at minimum).
- player-docs freshness check → regenerate (understanding-the-screen +
  keys-reference gain G/5).
