---
title: Panel — minibuffer (guardian input)
class: panel
status: shipped
verified_against: c8906da39be3a5b861c2272af37db0a83dcded7a
sources:
  - internal/tui/views.go
---

# Panel: minibuffer (guardian input)

The **only text input in the app**. One bordered line above the footer, present on
every widescreen page (home and solo views). Replaces typing directly into the
guardian pane. Governed by the focus contract
([../patterns/focus-contract.md](../patterns/focus-contract.md)); transcript and
replies live in the dock's guardian tab ([guardian.md](guardian.md)).

**Guardian strip (Wave 2, decision 7)**:
[guardian-strip.md](guardian-strip.md) — specified, `unbuilt` today — pairs
an always-visible action-budget line directly above this panel, so the
minibuffer reads as *the* verb. Nothing on this page changes when that strip
ships; it is a new sibling chrome row, not an edit to the minibuffer itself.
It is also this page's fold-last destination: under height pressure the
guardian strip's content relocates into this panel's **dormant-state** line
rather than disappearing (`patterns/layout.md` ruling a, step 4) — the
focused and busy states are unaffected.

**Guardian console composer**: [pages/guardian-console.md](../pages/guardian-console.md)
— specified, `unbuilt` today — reuses this exact component as its composer;
it introduces no second focusable input (`patterns/focus-contract.md`'s
"exactly one client" rule is not relaxed).

## States

**Reconciliation note**: the box carries no title text in the shipped
renderer (`minibufferView`) — the boxes below omit the illustrative
`┌─ METATRON ─┐` title the v1 mockup showed, which never matched
`internal/tui`; the panel is a plain bordered box, one content line, exactly
3 rows.

### 1 · Dormant (default)

```
┌────────────────────────────────────────────────────────────┐
│ ⏎ m — speak with the {{skin.guardian.epithet}}…            │
└──────────────────────────────────────────────────────────────┘
```

Dim border, dim placeholder that names the focus key. Zero keyboard ownership —
every global key works.

### 2 · Focused (`m`)

```
┌────────────────────────────────────────────────────────────┐  ← amber border
│ why did rowan lie about the wood▌       esc release · ⏎ send │
└──────────────────────────────────────────────────────────────┘
```

Amber border + live cursor + the exit hint rendered **inside the panel chrome**,
right-aligned. The focused state documents its own escape, every time it is drawn.

### 3 · Busy (question sent)

```
┌────────────────────────────────────────────────────────────┐
│ ⋮ the {{skin.guardian.epithet}} is answering…  esc to background │
└──────────────────────────────────────────────────────────────┘
```

- Focus is released automatically on send; `esc` (or any navigation) just proceeds —
  busy never blocks the UI.
- When the reply arrives: if the dock is on the guardian tab it streams
  there ([guardian.md](guardian.md)); otherwise the tab badges
  `{{skin.guardian.tab_label}} •` and the minibuffer flashes one dim,
  literal line — `answer arrived — 3 to read` (`Model.mbFlash`; the "3" is a
  fixed string, not a live unread count — a shipped-reality quirk this doc
  preserves rather than silently prettifies) — before returning to dormant.

## Control table

| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
| minibuffer box | dormant · focused · busy · flash | `Model.mbFocused`/`mbBusy`/`mbFlash` | `minibufferView` | `m` focus · — | TASK-34 | `skin.guardian.epithet` |
| input buffer | typing | `Model.mbInput` | `minibufferView` | printable keys · — | TASK-34 | — |
| input history | — | session-scoped history | `minibufferView` | `↑`/`↓` · — | TASK-34 | — |
| send | sent → busy | `Model.mbInput` | `sendConsole` | `⏎` · — | TASK-34 | — |
| release focus | focused → dormant | `Model.mbFocused` | `minibufferView` | `esc` (or empty-buffer `⏎`) · — | TASK-34 | — |
| reply flash | shown once · none | `Model.mbFlash` | `minibufferView` | — (display-only) | TASK-34 | — |

**Parity rollout**: focus (`m`), send (`⏎`), history (`↑`/`↓`), and release
(`esc`) have no mouse target today; tracked here rather than omitted
(decision 8, formal doctrine in `patterns/keymap.md`, T024).

## Rules

- Input history: `↑`/`↓` while focused cycle previous questions (session-scoped).
- Multi-line input is out of scope. Implementation note (TASK-34, B3): "wrap within
  the single logical line" turned out to be ambiguous — soft-wrapping a long input
  across multiple *rendered* rows grows the box past its fixed 3-row budget, which
  visually collides with the row-count invariant every other panel follows (see
  patterns/layout.md's Composition notes). The input display instead truncates to its
  visible tail (cursor glued to the right edge, like a normal terminal input line);
  the right-aligned hint is dropped first if there's no room for both. The box is
  always exactly 3 rows regardless of how long the question is.
- `⏎` on an empty buffer releases focus (no-op send).
- The minibuffer is chromeless-adjacent to the footer: footer hints while focused
  shrink to the minibuffer-mode keys only (see
  [../patterns/keymap.md](../patterns/keymap.md)).
- IPC send/receive is the existing guardian console protocol
  (`specs/005-metatron/contracts/console-protocol.md`) — transport unchanged, only
  the surface moves.
