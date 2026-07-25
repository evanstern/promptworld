# Quickstart: validating the lesson row (spec 055)

Prerequisites: repo at the task branch; Go toolchain; no LLM required (the feature is
model-free by contract).

## 1. Automated validation

```sh
go test -race ./internal/tui/ ./internal/worlds/   # feature tests + regressions
go test -race ./...                                # full sweep before PR
node scripts/check-tui-design.mjs --changed        # spec-047 gate (must pass)
node scripts/check-merge-drift.mjs pr              # spec-051 gate (from the worktree)
```

Expected: all green. The lesson tests must include, at minimum:
- the SC-001 fixture sweep — each of the 8 catalog triggers fires exactly once across
  a simulated two-world + restart sequence (fresh seen-record, replayed event
  fixtures);
- the SC-002 catalog↔overlay equality test;
- render tests: two-line row above the guardian strip, suffix present, `[lesson]`
  badge at stage 3+/pre-ladder and under fold pressure, narrow-carry (SC-003);
- seen-state tolerance: missing/corrupt/read-only file boots clean (SC-004);
- token resolution: no `{{` in any rendered string, default skin (SC-005).

## 2. Manual smoke (one covered trigger end-to-end)

```sh
rm -f ~/.promptworld/lessons-seen.json        # fresh user
promptworld new smoke-055 && promptworld ui smoke-055
```

1. Let the world run until a covered first occurrence lands (a `cog.outcome`
   suppression at speed is the quickest; `[` `]` to raise speed) — the two-line lesson
   appears above the guardian strip, ending `(? for more · x dismiss)`.
2. Press `?`, `tab` to the lessons section — the pushed lesson is listed (placeholder
   line gone).
3. Press `x` — row clears. Trigger the same event again — no lesson.
4. Quit, relaunch `promptworld ui smoke-055` — still no repeat;
   `cat ~/.promptworld/lessons-seen.json` shows the id.
5. Second world (`promptworld new smoke-055b`) — same trigger, still no repeat.

## 3. Design-gate companion (same PR)

- `docs/design/tui/panels/lesson-row.md`: `status: shipped`, renderer symbols filled,
  `verified_against` re-pinned.
- `docs/design/tui/patterns/keymap.md`: `x` moved into the global table; re-pinned.
- `node scripts/check-tui-design.mjs --changed` green after both amendments.
