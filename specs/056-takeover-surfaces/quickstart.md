# Quickstart: validating takeover surfaces (spec 056)

## Automated validation

```bash
go test -race ./...
go test -race ./internal/tui -run 'Takeover|Ceremony|Postmortem|ReportCard' -v
node scripts/check-tui-design.mjs --changed
```

Covers: precedence interleavings (SC-003), ambient/scored matrix (SC-002),
renderer three-site equivalence (SC-005), attach-to-ended auto-open,
exact-height in both layouts.

## Manual validation

1. Attach a world and let it die (or attach to an ended world) — the
   postmortem seizes the screen: run-end line + morgue rows (NO report card
   on an ambient world). `esc` dismisses; clock keys stay read-only; `p`
   reopens.
2. On a scenario world (spec 054) that fails: the report card renders above
   the morgue rows with met/missed markers.
3. Fixture (or spec-054 live) unlock: the ceremony seizes the screen —
   authorship-voice chapter + checklist. `q` detaches with "the world keeps
   running"; re-attach, open `?` — the ceremony-replay entry re-renders it.
4. Interleave: with the ceremony open, end the run — postmortem replaces
   it; the missed ceremony is still in `?`.
5. Narrow terminal: both takeovers full-screen, same content.

## Re-ground checklist (after merge)

- Both overlay pages shipped + help/keymap/guardian-console re-pins (in-PR).
- `/grounding-wiki:wiki-update` (tui-client.md).
- player-docs freshness → regenerate (keys-reference gains `p`; losing-is-fun
  paragraph may reference the postmortem takeover).
