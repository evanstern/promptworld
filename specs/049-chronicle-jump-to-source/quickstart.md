# Quickstart: validating chronicle jump-to-source (spec 049)

## Prerequisites

- A built client: `go build ./...`
- A running world with a few villagers (any existing world; `promptworld ps`
  to find one, or start a throwaway per README).

## Automated validation

```bash
go test -race ./...                            # full suite incl. new tests
go test -race ./internal/tui -run 'Jump|Mouse|DetailAction|CatalogSweep' -v
node scripts/check-tui-design.mjs --changed    # design-doc same-PR gate
```

Expected: all green. The catalog sweep proves every cataloged event type
resolves to jump-or-hint (SC-002 totality); mouse tests prove click routing
and the running-clock no-op; a keyboard-regression test proves existing
inspect keys are unchanged (FR-005).

## Manual validation (attached TUI)

1. `promptworld attach <world>` in a widescreen terminal; wait for events.
2. `space` to pause → inspect mode. Select (`j`/`k`) an event about a
   villager (e.g. `agent.foraged`).
   - Detail pane bottom-right shows `⏎ jump to <name> (x,y)` (US3).
3. Press `⏎` → map centers on that villager; map title reads
   `MAP · panned (c to recenter)` (US1 AS-1). `c` → follow resumes (AS-2).
4. Select a world-lifecycle/telemetry event (no subject) → bar shows
   `no location for this event`; `⏎` moves nothing (AS-3).
5. Click a different chronicle event line → selection moves there AND the
   camera jumps (US2 AS-1). Resume (`space`) → clicking lines does nothing
   (AS-2).
6. Narrow terminal (< widescreen breakpoint): pause on the chronicle pane,
   `⏎` on a located event → you land on the map pane, centered (FR-007);
   switch back — selection preserved.

## Re-ground checklist (after merge)

- `docs/design/tui/panels/chronicle.md` + `patterns/keymap.md` amended and
  re-pinned in the merged PR (gate held).
- `/grounding-wiki:wiki-update` (touches `internal/tui` sources of
  `docs/wiki/tui-client.md`).
- `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`;
  regenerate player docs if stale (keys-reference gains the new bindings).
