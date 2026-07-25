# Contract: per-user seen-state file (spec 055)

**Path**: `~/.promptworld/lessons-seen.json` — sibling of `unlocks.json`, resolved by
the same home-dir logic (`internal/worlds` owns it; the TUI never touches the path
directly).

**Schema** (version 1):

```json
{
  "version": 1,
  "seen": {
    "<lesson-id>": { "first_shown": "<RFC3339>", "world": "<world-name>" }
  }
}
```

**Semantics** (the unlocks.json discipline, D8):

- **Load-tolerant**: missing / unreadable / malformed / unknown-version ⇒ empty record,
  normal boot, no player-visible error. Ever.
- **Advisory, never authority**: nothing in the game gates on this file. Worst failure
  mode anywhere in its lifecycle is a repeated lesson.
- **Atomic write**: temp-file-then-rename per upsert. A failed write is swallowed; the
  session's in-memory set still suppresses repeats until exit.
- **Per-user**: no world component in path or key; ids are global lesson ids.
- **Write timing**: a lesson is recorded when it SURFACES (becomes the active row
  entry) — queued-then-decayed entries are not recorded.
- **Reset**: deleting the file. No in-app reset in v1.
- **Compatibility**: readers ignore unknown top-level and per-entry fields; writers
  preserve `version` and write only fields in this contract.
