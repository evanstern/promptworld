# Data Model: Guardian console + systems split (spec 053)

No persistence, no wire changes. Client-side state only.

## Console page state

| Field | Type | Meaning |
|---|---|---|
| console | bool | page open |
| consoleReturn | prior-view snapshot (solo, dockTab) | where `G`-toggle/esc lands back |
| consoleScroll | int | tail-anchored scrollback offset; reset on close |
| consoleNotice | string (one-shot) | "charter changed — next turn binds it" / $EDITOR error notice |

**Transitions**: closed →(G)→ open; open →(G / 1 / esc-with-nothing-focused)→
closed (restore snapshot). esc-release ordering gains the console layer
(focus-contract rule 3): minibuffer → villager detail → console → solo →
home. Overlay (`?`) over the console follows existing overlay rules.

## Dock tab extension

`dockTab` enum + `paneNames` gain `systems` (key `5`). Per-tab state,
unseen-badge, same-key solo zoom, narrow pane reachability inherit from the
existing machinery. The systems content renderer owns the relocated
telemetry; the guardian tab renderer loses it.

## consoleCard (the seam)

`interface { renderCard(width int) string }` + an always-empty-this-feature
`[]consoleCard` composed between turn stream and read surface. Producers
arrive with TASK-127 (shared report-card renderer) and TASK-115
(stopping-point production). Tests pin the composition position with a test
fake.

## $EDITOR round-trip

Pre-exec: (path, content-hash) of `charter.md`. Post-exec callback:
re-hash → changed ⇒ consoleNotice set; error/unset `$EDITOR` ⇒ error notice.
No other state; the daemon's per-turn re-read is the binding mechanism.
