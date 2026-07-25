# Contract: scripts/check-tui-design.mjs

Read-only structural validator + same-PR gate for `docs/design/tui/`. Node ≥ 18,
ESM, zero npm dependencies (stdlib `fs`/`path`/`child_process` only). The script
NEVER writes files.

## Usage

```
node scripts/check-tui-design.mjs [--json] [--changed [<range>]]
```

- no flags: structural checks only (file-set, pins, control tables, anatomy).
- `--changed [<range>]`: additionally run the same-PR gate over the git range
  (default `origin/main...HEAD`).
- `--json`: emit the report as JSON instead of human-readable lines.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | all checks pass |
| 1 | ≥ 1 violation (each reported with file + actionable message) |
| 2 | usage/environment error (bad flag, not a git repo, docs/design/tui unreadable) |

## Checks

| id | Assertion | Violation message must name |
|---|---|---|
| `file-set` | every `.md` under `docs/design/tui/` is `INDEX.md`, `anatomy.md`, or lives in `pages/`\|`panels/`\|`overlays/`\|`patterns/`; frontmatter `class` matches its directory | the offending file and its class/location mismatch |
| `pins` | frontmatter present; `verified_against` is 40-hex and resolves to a commit | file + `missing`/`malformed`/`unresolvable` |
| `control-tables` | every `panels/*`/`overlays/*` page has exactly one table with the canonical header (contracts/control-table.md), and zero additional tables reusing that header | file + `missing`/`duplicate`/`non-canonical header` |
| `anatomy` | every row target in `anatomy.md` exists; every `pages/`/`panels/`/`overlays/` file is targeted by ≥ 1 anatomy row | the unmapped file or dangling target |
| `same-pr` | (only with `--changed`) if the range touches `internal/tui/**`, the range must also touch `docs/design/tui/**` | the touched TUI files + "amend docs/design/tui in this PR (re-verify + re-pin affected pages)" |

## JSON report shape

```json
{
  "ok": false,
  "checks": [
    { "check": "pins", "status": "violation",
      "file": "docs/design/tui/panels/systems.md",
      "message": "verified_against missing" }
  ]
}
```

## Gate wiring

- `INDEX.md` states the rule: any PR touching `internal/tui/` runs
  `node scripts/check-tui-design.mjs --changed` and amends the reference in the same
  PR (extends old INDEX rule 4 from "record deviations" to "any change").
- `CLAUDE.md` gains one line beside the player-docs freshness rule naming the script
  and when to run it (before opening any PR that touches `internal/tui/`).
