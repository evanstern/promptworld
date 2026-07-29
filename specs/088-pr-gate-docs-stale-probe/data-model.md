# Data model — spec 088 (TASK-162)

## Trigger set (pr-mode docs-stale probe)

| Trigger | Derivation | Fires the probe when |
|---------|-----------|----------------------|
| wiki prefix | static `docs/wiki/` prefix (existing) | any branch diff path starts with `docs/wiki/` |
| declared page sources | union of `promptworld-docs:source` tags across `docs/player/*.html` at the branch tip | any branch diff path equals a declared source (covers README.md, docs/llm-providers.md, spec 046 quickstart sources — whatever the pages actually pin) |
| history move | `git rev-list --merges origin/main..<tip>` non-empty | the branch contains any merge commit since diverging from origin/main |

The probe (player-docs freshness checker) is invoked AT MOST ONCE per run regardless
of how many triggers match (FR-004).

## Finding rules

| Rule | Severity | New/changed | Emitted when |
|------|----------|-------------|--------------|
| `player-docs-stale` | block | trigger set widened (rule unchanged) | checker exit 1 on any trigger |
| `player-docs-env-error` | block | trigger set widened (rule unchanged) | checker missing or exit ≥2 on any trigger |
| `tui-design-stale` | block | **new** | tui-design checker exit 1 (failing pages named from its `--json` report) |
| `tui-design-env-error` | block | **new** | tui-design checker missing or exit 2 |
| `tui-surface` | warn | unchanged | branch touches `internal/tui/` (reminder retained) |

Tui-design delegation trigger: branch diff touches `internal/tui/`, `docs/design/tui/`,
or any source a design page pins; invoked at most once per run; env-overridable checker
path (`CHECK_MERGE_DRIFT_TUI_DESIGN_CHECKER`) for fixtures.

## Fixture matrix (all in `scripts/check-merge-drift.test.mjs`)

| # | Fixture | Expected |
|---|---------|----------|
| F1 | branch touches README.md only; checker stub exits 1 | exit 1, `player-docs-stale` |
| F2 | branch touches a declared spec-046 source; checker stub exits 0 | exit 0, no player-docs finding |
| F3 | branch tip contains merge of main; diff touches no pinned source; checker stub exits 1 | exit 1, `player-docs-stale` |
| F4 | same as F3 with checker stub exit 0 | exit 0 (re-trigger adds no false block) |
| F5 | branch touches tui-pinned source; tui stub exits 1 | exit 1, `tui-design-stale` |
| F6 | F5 after re-pin (tui stub exits 0) | exit 0; `tui-surface` warn only |
| F7 | pinned input + history move combined; both stubs exit 1 | one `player-docs-stale`, no duplicates |
| F8 | branch touches nothing pinned, no merge commits | checker not invoked (existing test 069/US2 preserved) |
| F9 | checker missing on a README-only trigger | exit 1, `player-docs-env-error` |

Existing fixtures must pass unchanged (SC-004).
