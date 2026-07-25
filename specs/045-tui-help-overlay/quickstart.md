# Quickstart: validating spec 045 end-to-end

Prerequisites: `go build ./... && go test ./internal/tui` green.
References: [contracts/help-content.md](./contracts/help-content.md),
[data-model.md](./data-model.md).

## 1. Hermetic proof

```bash
go test ./internal/tui -run 'Help|TestWidescreenViewExactHeight|TestFocusContract' -v
```

Expected: `?` opens from every mode incl. narrow fallback; tier advance works; esc/`?`
release exactly one layer; keymap sweep green (every listed key real, every real key
listed once); exact-height holds with the overlay open at all sizes; nil-status
(no-LLM) content byte-identical.

## 2. Live walkthrough

```bash
promptworld new demo-045 && promptworld start demo-045 && promptworld attach demo-045
```

- In home view press `?` → overlay, basic global keys first; advance tier → advanced
  page; `?` again → dismissed, screen exactly as before, clock never paused.
- `2` (chronicle) + `space` (pause) → inspect mode → `?` → inspect keys first.
- `4` → villagers → `⏎` (detail) → `?` → detail keys; esc closes only the overlay,
  detail still open; esc again → roster (one layer per press).
- `m` (minibuffer) → type `what?` → the `?` appears in the buffer (never hijacked).
- Navigate to the screen walkthrough → verify header anatomy (all badges explained),
  full glyph legend, dock tabs; scroll with the pager on a small terminal.
- Lessons section → renders the placeholder line (empty table today).

## 3. No-LLM floor (US3)

Repeat §2 on a world with no `llm.json`: identical overlay content and behavior.

## 4. Anti-drift check (SC-003 / FR-005)

Temporarily add a fake binding or glyph in a scratch branch → the sweep test fails;
revert. (Proves the mechanical gate bites.)
