---
title: Overlay — help (`?`)
class: overlay
status: shipped
verified_against: c8906da39be3a5b861c2272af37db0a83dcded7a
sources:
  - internal/tui/help.go
  - internal/tui/tui.go
---

# Overlay: help

The context-sensitive help overlay (spec 045/TASK-116). Extracted from
`patterns/keymap.md` into its own overlay page (this feature, FR-005) — the
mode-key tables that page's "Mode: help overlay" section used to carry now
live here, reconciled against `internal/tui/help.go` and
`specs/045-tui-help-overlay/contracts/help-content.md`. `patterns/keymap.md`
stays the one printable reference card and only points here.

## Mockup

```
┌ HELP · keys ──────────────────────────────────────────────┐
│ Global (home) — basic         (t: tier · n/p: mode)        │
│                                                            │
│ 2               select the chronicle tab; press again to  │
│                 solo it (or return home if already solo'd) │
│ 3               select the {{skin.guardian.tab_label}} tab; │
│                 press again to solo it (or return home)    │
│ 4               select the villagers tab; press again to   │
│                 solo it (or return home if already solo'd) │
│ m               focus the minibuffer — ask the              │
│                 {{skin.guardian.epithet}}                   │
│ space           pause / resume the clock                  │
│ q               quit                                       │
│                                                            │
│ ctrl+c — quit, from anywhere                                │
└──────────────────────────────────────────────────────────┘
 esc/? close · tab section · t tier · n/p mode · J/K scroll
```

## Layering

1. `?` opens the overlay from every mode except text-entry (minibuffer
   focused: `?` is a buffer character, per the focus contract — minibuffer
   help is content reachable from other modes' overlays instead).
2. While open, the overlay owns the keyboard entirely: its own navigation +
   dismissal only; every other key is inert — no silent fallthrough to the
   mode beneath.
3. Dismissal (`esc` or `?`, toggle) releases exactly **one** layer — the
   esc-release chain beneath (minibuffer → decisions → detail → solo → home)
   is untouched, and the overlay is never stacked a second time.
4. Opening/using/dismissing sends no IPC command, emits no event, and
   changes no world or layout state beyond the help overlay's own fields
   (`helpOpen`, `helpMode`, `helpPageMode`, `helpTier`, `helpSection`,
   `helpScroll`) — rendering is body replacement in the solo-zoom slot
   (`helpPanelView`), chrome (header/minibuffer/footer) stays visible, output
   remains exactly terminal-height.

## Sections & tiers

### Section 1 — keys (frozen at open)

The mode you were in when you pressed `?` (`m.helpMode`, frozen via
`currentHelpMode`) determines the starting page, but `n`/`p` page across
**every** mode's key table regardless of where you opened from — the only way
to read the minibuffer's page, since `?` never opens from there. Six mode
pages total (`helpModeKey`):

| Mode page | Covers |
|---|---|
| Global (home) | `handleGlobalKey`'s dispatch — tabs, ask, pause, speed, pan, recenter, chronicle filters, quit |
| Minibuffer (focused) | `handleMinibufferKey` — send, history, release, typing |
| Inspect (paused, chronicle) | `handleInspectKey` — selection, jump, detail scroll, resume |
| Villagers (roster) | `handleVillagersKey`, roster state — selection, jump, open detail |
| Villagers (detail) | same handler, detail/decisions state — toggle decisions, back, scroll |
| Solo zoom / narrow fallback | the identical global dispatch, partitioned into a different basic/advanced split for its footer |

Each page has a **basic** tier (≈ the footer-hinted set for that mode) and an
**advanced** tier (`t` toggles) — every real, working binding appears in
exactly one tier of its mode's page; `help_test.go`'s keymap sweep
mechanically ties every advertised row to a real handler and every handled
binding to exactly one tier (SC-003/FR-003).

### Section 2 — the screen (`tab`/`shift+tab` to reach)

- **Header anatomy** (`headerAnatomy`) — every element and conditional badge
  `headerView` can render, including ones the player may never have seen yet
  (governed speed, an `[llm: …]` condition, `[suppressed: …]`) — see
  [../pages/home.md](../pages/home.md) for the reconciled header-segment
  inventory this table mirrors.
- **Map glyphs** (`mapGlyphs`) — rendered long-form from the *same* table
  `renderMapGrid`'s compact legend line reads (`legendGlyphLine`), so the
  overlay's glyph walkthrough and the in-game legend cannot silently diverge
  — see [../panels/map.md](../panels/map.md) for the reconciled glyph
  inventory (this feature added the wall/path/quarried/pile/grave rows to
  both the map page and, by construction, this shared table).
- **Dock tabs** (`dockTabs`) — key/name/purpose for each dock tab, read off
  `dockTabKey`/`paneNames` (never a second, hand-maintained list) — currently
  chronicle/`{{skin.guardian.tab_label}}`/villagers; a future tab (systems,
  exercise) is a new entry here with no structural change.

### Section 3 — lessons (pull reference)

`helpLesson{id, title, body}` entries render on demand; the table ships empty
today (`helpLessons` is nil) and the section shows a placeholder line
("lessons appear here as the village teaches them") until the future
first-occurrence lesson projection (decision 5, Phase 4 of this feature)
registers entries here — a content addition, no structural change to this
overlay's navigation or rendering (SC-006).

### [Placeholder] Section — the guardian (decision 9)

**Not authored in this slice.** The reorientation's `?`-overlay guardian
section — static-per-stage, model-free content (stage identity/concept,
granted verbs, one example ask per verb) — lands in a later phase of this
feature (T018, after this page exists). This heading is a deliberate
placeholder so `anatomy.md` and `INDEX.md`'s file map can point at "the
guardian section of `overlays/help.md`" today without a broken link, and so
a reader knows exactly where it will land.

## [Placeholder] Byte-identity classification (research.md R4)

**Not authored in this slice.** The reorientation's Wave 0 ruling (c) —
which sections of this overlay stay byte-identical with nil status (the
deterministic no-LLM floor, spec 045's original invariant) versus which are
now status-derived (the guardian section, lessons registry, badge deep-link
focus, ceremony replay entries) — lands in a later phase of this feature
(T023, after T011 and T018 both exist). Today's overlay content in full
(keys, screen walkthrough, and the still-empty lessons stub) remains exactly
what spec 045 shipped: static, local, model-independent, byte-identical on
every no-LLM world by construction (`help.go`'s own doc comment; no
daemon/IPC/event/world-state read anywhere in this file).

## Control table

| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
| open/dismiss overlay | closed · open | `Model.helpOpen` | `openHelp`/`closeHelp` | `?` (open, any non-text mode) · `esc`/`?` (dismiss) · — | spec 045 | — |
| section cycle | keys · screen · lessons | `Model.helpSection` | `helpPanelView` | `tab`/`shift+tab` · — | spec 045 | — |
| tier toggle (keys section) | basic · advanced | `Model.helpTier` | `helpKeysLines` | `t` · — | spec 045 | — |
| mode paging (keys section) | 6 mode pages | `Model.helpPageMode` | `helpKeysLines`, `nextHelpMode`/`prevHelpMode` | `n`/`p` · — | spec 045 | — |
| pager scroll | — | `Model.helpScroll` | `paginateHelpContent` | `J`/`K` · — | spec 045 | — |
| header anatomy row | static | `headerAnatomy` | `helpWalkthroughLines` | — (display-only) | spec 045 | — |
| map glyph row | static | `mapGlyphs` (shared with `legendGlyphLine`) | `helpWalkthroughLines` | — | spec 045 | — |
| dock tab row | static | `dockTabs` | `helpWalkthroughLines` | — | spec 045 | — |
| lessons pull-reference entry | empty (placeholder line) · populated | `helpLessons` | `helpLessonsLines` | — | spec 045 (seam); content Phase 4 | — |
| the guardian section | unbuilt | — | `unbuilt (wave 4)` | — | reorient D9 | — (placeholder — T018) |

**Parity rollout**: every control above has a key but no mouse target today;
tracked here rather than omitted (decision 8, formal doctrine in
`patterns/keymap.md`, T024).
