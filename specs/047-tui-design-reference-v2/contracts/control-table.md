# Contract: canonical control table

Every `panels/*.md` and `overlays/*.md` page carries **exactly one** control table.
The check script asserts the header row byte-exactly (after whitespace
normalization); authors and sweep scripts rely on the column order.

## Canonical header

```markdown
| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
```

## Column value grammar

| Column | Grammar | Examples |
|---|---|---|
| control/region | short name of one visible element | `provider row`, `tab badge`, `composer` |
| states | `·`-separated state names | `dormant · focused · error` |
| data source | `Status.<field>` \| `event:<type>` \| `replica:<path>` \| `static` \| `per-user:<file>` | `Status.Providers`, `event:curriculum.stage_unlocked`, `static` |
| renderer | Go symbol \| `unbuilt (wave N)` | `llmProviderLines`, `unbuilt (wave 3)` |
| keys+mouse | `<keys> · <mouse>`; `—` if display-only; mouse REQUIRED for every actionable control (decision 8) | `⏎ · click row`, `? · click badge`, `—` |
| introduced-by | `spec NNN` \| `TASK-N` \| `reorient <decision>` | `spec 024`, `reorient D11` |
| skin-token | `skin.<domain>.<name>` \| `—` | `skin.guardian.name`, `—` |

## Rules

1. One table per page — additional tables (e.g. a token index) must not reuse the
   canonical header.
2. No bare fiction literal anywhere in the row — fiction goes through `skin-token`;
   mockups use `{{skin.<domain>.<name>}}` placeholders.
3. `data source` may name raw registry/engine values (this column is
   engineer-facing); the player-facing projection of the same information is
   plain-language, with raw values only behind the debug/inspector toggle (FR-020).
4. Display-only elements use `—` in keys+mouse; an actionable control with a key but
   no mouse target is a parity violation to be listed in the page's
   "parity rollout" note, not silently omitted (decision 8 — incremental rollout,
   honestly tracked).
