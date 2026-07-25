# Phase 0 Research: `?` help overlay in the TUI

**Spec**: `specs/045-tui-help-overlay/spec.md` | **Date**: 2026-07-25

Grounded against code (file:line verified 2026-07-25). No NEEDS CLARIFICATION remain.

## R1 — Overlay slot: head of the key-dispatch chain, `Model` state

**Decision**: A `helpOpen` bool (plus `helpTier`, `helpSection`, `helpScroll`) on `Model`.
Check in `handleKey` immediately after the `ctrl+c` guard (tui.go:524) and **before**
`m.mbFocused` (tui.go:527), with the open-trigger `?` matched only when `!m.mbFocused`
(minibuffer keeps `?` as a buffer rune, tui.go:763-764 — FR-001). While open, the overlay
owns the keyboard: its own navigation, dismissal by `esc`/`?`; everything else inert
(FR-012). This makes help the new head of the esc-release chain: help → minibuffer →
decisions → detail → solo → home.

**Rationale**: modes here are derived predicates checked in priority order, not an enum;
the head-of-chain slot is the only place that guarantees "every mode, one keypress" and
one-layer esc release. `?` binds nowhere today (verified — no collision); an unmatched
rune currently no-ops silently in non-entry modes, which this feature fixes.

## R2 — Rendering: body replacement, solo-zoom precedent (no z-compositing)

**Decision**: `helpPanelView` rendered *instead of* the body — in `widescreenView` the
map‖dock body swaps out (the `m.solo` / `soloPanelView` precedent, views.go:220-239,
291-304), in `narrowView` the pane body swaps (views.go:81-97) — keeping header/
minibuffer/footer chrome. Sized with the same
`styleBox.Width(cols-2).Height(rows-2).Render(clipContent(...))` discipline so the
exact-height invariant holds (`TestWidescreenViewExactHeight` gains help states).

**Rationale**: the TUI has no z-layer compositing; all shipped "overlays" are full-region
body replacements. Keeping chrome visible also keeps the walkthrough self-referential
(the header being explained stays on screen).

## R3 — Content: static pages in a new `internal/tui/help.go`, derived not duplicated

**Decision**: Page content lives in a new `help.go` as pure functions/tables (the
`digest.go` style). Mode key pages are static per-mode tables (source of truth:
`docs/design/tui/patterns/keymap.md`), basic tier = the keys the footer hints advertise,
advanced tier = the remainder + layered globals. The glyph-legend page is **derived
beside the renderer's legend** (`renderMapGrid`'s legend string, views.go:615-617)
via a shared table both consume — one source, so spec-044's grave glyph and any future
glyph surface in both automatically (FR-005). Header-anatomy content enumerates
`headerView`'s elements and every conditional badge (running/PAUSED + governed-speed
suffix + `[degraded]` + `[llm: …]` + `[suppressed: …]`, views.go:99-135) — including the
ENDED posture landing with spec 044.

**Rationale**: FR-005/SC-003 demand the legend and keymap cannot silently drift; a shared
table plus a mechanical sweep test is the codebase's established anti-drift move
(TestCatalogSweep precedent).

## R4 — Scroll/paging: copy the `chronicleDetailPane` pager exactly

**Decision**: `helpScroll` follows the shipped pager pattern: increment unbounded on
`J`, floor-0 on `K`, clamp at render, overflow footer `"… (+N more — J to scroll)"`
(views.go:1097-1141; scroll state idiom tui.go:815-821). Section navigation cycles
pages (keys / walkthrough / reference); tier advance on `?`… no — `?` toggles closed
(spec edge case); tier advance uses a dedicated key shown on the overlay's own footer.

**Rationale**: two shipped instances of this pager idiom (`chronDetailScroll`,
`villDecisionsScroll`) — a third copy is consistency, not invention. FR-009 small-
terminal usability comes from clamp-at-render.

## R5 — Footer hints advertise `?`

**Decision**: Append `· ? help` to each branch of `footerView`'s per-mode switch
(views.go:197-216) and update the keymap doc's footer table
(docs/design/tui/patterns/keymap.md:74-82). No test pins exact hint strings (verified),
so this is additive.

## R6 — No-LLM identity is structural

**Decision**: All content is static strings in the binary; the overlay reads no
`status.LLM`, no replica LLM state, no network. `New`/`View` already tolerate nil
status/replica (tui.go:151-153, views.go:108-110); tests exercise the overlay with nil
status to prove SC-004 byte-identity.

## R7 — Pull-reference seam (US4)

**Decision**: The overlay's section list includes a "Lessons" reference section whose
entries come from a `helpLessons` table in `help.go` — today empty except a contract
comment; the future first-occurrence-lesson feature appends entries (id, title, body)
with no structural change. Documented in contracts/help-content.md (SC-006).

**Rationale**: mirrors the `detailActions` attachment-point idiom (tui.go:996-1005) —
reserve the seam as data, not machinery.

## R8 — Known collision: task-31 (spec 044) in flight

`task-31-run-outcomes-morgue` will touch `headerView` (ENDED token, T009), the legend
line (grave glyph, T025), `footerView` (inert-clock hint), and `views_test.go`. This
feature must expect a rebase before PR; the shared-table design (R3) is what keeps the
merge mechanical. Overlay content must document the ENDED posture and grave glyph once
they exist — if 044 lands first, include them; if not, the shared table picks them up
when it does.
