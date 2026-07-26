---
title: Pattern — keymap
class: pattern
status: shipped
verified_against: 66e36e9a7a627161d4b2ec95dcc18aa0f4f91d20
sources:
  - internal/tui/tui.go
  - internal/tui/help.go
  - internal/tui/views.go
  - internal/tui/look.go
---

# Pattern: keymap

Every key, every mode. Three modes total; a key means one thing per mode. The
footer always shows the current mode's primary hints. This is the one
printable reference card — one page, every binding — kept deliberately
separate from [../overlays/help.md](../overlays/help.md) (this feature
extracted the in-app help overlay's content there; this page never grows a
walkthrough or lessons section of its own).

## Mode: global (minibuffer unfocused — the normal state)

| Key | Action |
|---|---|
| `1` | home composite (from solo: return home; on home: map is already primary) |
| `2` / `3` / `4` / `5` | select dock tab chronicle / `{{skin.guardian.tab_label}}` / villagers / systems; **same key again** → solo zoom; again → back home |
| `6` | select the exercise tab (spec 054 — scenario worlds only; inert on ambient worlds, which have no such tab); same solo-zoom grammar as `2`–`5` |
| `G` | open the guardian console — a full-screen page for the conversation, charter/skills, and `$EDITOR` (spec 053; shadowed by inspect mode's and the villagers tab's own `G`, "Mode: console" below) |
| `m` | focus the minibuffer |
| `x` | dismiss the active lesson row (spec 055, `panels/lesson-row.md`); strict no-op when nothing is active |
| `space` | pause / resume the clock |
| `[` / `]` | speed down / up |
| `←↑↓→` | pan the map |
| `c` | recenter camera (resume following) |
| `r` | chronicle: toggle raw ↔ narrated |
| `a` / `t` | chronicle: filter by agent / thread |
| `q` | quit |
| `p` | reopen the postmortem takeover (only while the run has ended; spec 056) — reachable from every mode below `?`/minibuffer focus in `handleKey`'s chain (global, console, inspect, villagers), not shown again per-mode |
| `ctrl+c` | quit (from **any** mode) |

## Mode: minibuffer focused (after `m`)

| Key | Action |
|---|---|
| printable keys, `space` | append to buffer (visibly) |
| `backspace` | delete |
| `↑` / `↓` | input history |
| `⏎` | send (empty buffer: release focus instead) |
| `esc` | release focus |
| `ctrl+c` | quit |

No other key does anything silently — see
[focus-contract.md](focus-contract.md) rule 4.

## Mode: inspect (clock paused + chronicle visible; layered on global)

**TASK-60 (spec 018-chronicle-digest)**: the detail pane is always on (no
`⏎` needed to see it — panels/chronicle.md "Mode 2"); `J`/`K` scroll the pane
when its content overflows. **Spec 049**: `⏎` fills the seam this page used
to describe as reserved — it jumps the map camera to the selected event's
subject (`panels/chronicle.md` "Jump-to-source"; `contracts/jump-to-source.md`),
an honest no-op with a visible hint when the event has none.

| Key | Action |
|---|---|
| `j` / `k` | select next / previous event (also resets detail pane scroll) |
| `g` / `G` | jump to first / last (also resets detail pane scroll) |
| `J` / `K` | scroll the detail pane down / up |
| `⏎` | jump: center the map camera on the selected event's subject (unlocatable: no-op, actions bar names the absence) |
| `space` | resume (exits inspect, clears selection and detail scroll) |

All global keys stay live in inspect mode; `j/k/g/G/J/K` are additions, not
replacements. (Map pan keeps the arrow keys; inspect deliberately uses `j/k`
so the two never collide. `J`/`K` mirror `j`/`k` one layer up — selecting a
row vs. scrolling what it shows.)

## Mode: villagers (the villagers tab is the thing visible; layered on global, TASK-56)

Scoped to whichever tab is on screen — the dock shows one tab at a time, so this
never collides with inspect mode's `j/k/J/K` (chronicle) or the map's arrow-pan.
Unlike inspect mode, this does **not** require the clock to be paused.

| Key | Roster view | Detail view |
|---|---|---|
| `j` / `k` | select next / previous villager (clamped) | no-op (unless the decisions sub-view is open — scrolls it) |
| `g` / `G` | jump to first / last villager | no-op |
| `⏎` | open detail for the selected villager | no-op |
| `d` | no-op (falls through to global — no global binding either, so inert) | toggle the decisions sub-view (spec 020, TASK-63) |
| `esc` | (falls through to global — releases solo/home) | close the decisions sub-view if open, else close detail → back to roster |

`esc` follows "esc always releases" ordering
([focus-contract.md](focus-contract.md) rule 3, spec 053 amendment): minibuffer
→ villager detail → **console** → solo → home — each press of `esc` releases
exactly one layer. With no world state loaded (or an empty roster)
`j/k/g/G/⏎` are strict no-ops.

## Mode: look-cursor (`v` toggles it on; spec 074-look-cursor, layered on
global exactly like inspect/villagers)

`v` enters from the global mode whenever the map is the thing on screen
(widescreen home — never solo — or the narrow map pane) and a world is
attached; a strict documented no-op otherwise (the `x`-key precedent).
While active, the dock body is borrowed by a transient **TILE view**
([../panels/dock.md](../panels/dock.md)) — never a numbered tab, `m.dockTab`
untouched — and the villagers/inspect key layers go dormant (their own
tabs' visibility predicates read false during the borrow, the
villagers-mode scoping precedent: "the TILE view is the thing visible").
Unclaimed global keys (`space`, `[`/`]`, `m`, `q`, `p`, `?`) keep working;
`a`/`t`/`r` (chronicle-only) and `1` are reachable but inert, since the
chronicle isn't the thing visible and `1`'s state is already true
(`lookEntryAllowed()` required it). `G` (console) exits the mode first, then
opens the console exactly as it would from global — the borrow never
survives being covered by a body-replacing surface (same rule for a solo
zoom via the exit-and-select digits below).

Three focus layers inside the mode, released one at a time by `esc`
(`patterns/focus-contract.md` rule 3): **cursor** (the default) → **pane**
(`⏎`/`tab`, amber border — focus is drawn) → **drill** (`⏎` on a drillable
row). `tab`/`shift+tab` are a second way back toward the cursor from pane/
drill (the same direction `esc` releases), claimed at every layer so they
never leak to the global dock-cycle/solo-zoom underneath the borrow.

| Key | Cursor focus | Pane focus | Drill focus |
|---|---|---|---|
| `v` | exit the mode (from anywhere inside it) | same | same |
| `hjkl` / arrows | move the cursor 1 tile, pushing the camera at a 2-tile margin | select a TILE-pane row (`j`/`k` only; `left`/`right` a claimed no-op) | scroll the drill-in pane (`up`/`down` alias `J`/`K`; `left`/`right` a claimed no-op) |
| `H`/`J`/`K`/`L` | jump the cursor 8 tiles, clamped to the world edge, never wrapping | `H`/`L` a claimed no-op; `J`/`K` alias `j`/`k` | `J`/`K` scroll the pane; `H`/`L` a claimed no-op |
| `c` | snap the camera onto the cursor (`centerCameraOn`) | — | — |
| `⏎` / `tab` | focus the TILE pane | `⏎` drills into the selected row; `tab`/`shift+tab` release to cursor | `tab`/`shift+tab` release to pane |
| `esc` | exit the mode (camera resumes centroid-following) | release to cursor focus | release to pane focus |
| `2`–`6` | exit the mode and select that dock tab (`selectTab` — pressing the tab already selected before entry zooms it solo, the one honest consequence of routing digits through the single `selectTab` grammar rather than a second one) | same | same |

**Mouse** (US4, decision 8 rule 1 — ships together with the keyboard): a
left-click on a map tile moves the cursor there, entering the mode first if
it was inactive (`panels/map.md`); a click on a TILE-pane row selects it
(acquiring pane focus if the cursor held it); a second click on the already-
selected row drills in (`panels/dock.md`). No-ops while the help overlay is
open or the minibuffer is focused, the same guards every other click in
this corpus honors.

**Narrow fallback** (FR-012): `v` raises the cursor on the map pane while it
is active; `⏎`/`tab` swaps that single pane's body to the TILE view (content
swap only — the "one component, two widths" posture, never a layout
change); the esc chain is unchanged.

## Mode: console (the guardian console page is open; spec 053, layered on global)

The console ([pages/guardian-console.md](../pages/guardian-console.md)) is a
full-screen **page**, not a dock tab or an overlay — reached by `G` from the
global mode (home, solo, narrow), never from inside inspect mode or the
villagers tab, both of which already bind `G` to their own jump-to-last
(above) and keep it. While the console is open, its own key layer owns a
handful of keys; every other global key still works underneath (the
console's own footer names them), unlike inspect/villagers mode's narrower
"layered, most global keys pass through unchanged" shape.

| Key | Action |
|---|---|
| `G` | close, restoring whatever was open before (`1`/unfocused `esc` do the same) |
| `m` | focus the minibuffer (the composer) — in place, never switching the narrow fallback's active pane |
| `e` | `$EDITOR` shell-out on the world's `charter.md` (contract §4) |
| `J` / `K` | scrollback down (toward the tail) / up (toward older turns) |
| everything else (`space`/`q`/`[`/`]`/`2`-`5` — plus `6` on scenario worlds/pan/`a`/`t`/`r`) | falls through to the global mode unchanged |

While the minibuffer is focused inside the console, `G`/`e`/`5`/`J`/`K` all
type into the buffer instead — no silent stealing (focus-contract.md rule 4).

## Mode: takeover (the ceremony or postmortem is open; spec 056, layered
above every other mode)

The takeover family ([overlays/ceremony.md](../overlays/ceremony.md),
[overlays/postmortem.md](../overlays/postmortem.md)) owns the keyboard
before EVERYTHING else — `handleKey`'s first check after `ctrl+c`, ahead
of help, minibuffer focus, the console, inspect, and villagers mode alike.
Unlike "Mode: console"'s "most global keys pass through unchanged" shape,
a takeover swallows every key it doesn't name below — the takeover IS the
event, not a mode layered alongside the others.

| Key | Ceremony | Postmortem |
|---|---|---|
| `esc` | dismiss (returns to whatever was beneath — help/console state is untouched, not force-closed) | dismiss + latches `postmortemDismissed` (suppresses the next connect's auto-reopen this session; `p` overrides) |
| `q` | quit/detach — the D13 "world keeps running" framing (the run hasn't ended) | quit/detach — the plain, honest ended-world quit (`View()`'s `runEnded()` check drops the "keeps running" claim) |
| `?` | swallowed — the takeover keeps the body slot, help does not open | swallowed |
| everything else | swallowed | swallowed |

## New global keys (specified, unbuilt — this feature)

Nothing remains unbuilt from this table as of spec 056/TASK-127: `G` (open
the guardian console) shipped with spec 053/TASK-125; `x` (dismiss the
lesson row) shipped with spec 055/TASK-117; `p` (reopen the postmortem)
shipped with spec 056/TASK-127 — all three now live in the tables above.

## Mode: exercise briefing (scenario worlds; the exercise tab visible, briefing not yet dismissed — spec 054)

While the exercise tab's attach-time briefing is on screen
([../panels/exercise.md](../panels/exercise.md)), **any key dismisses it and
is consumed** — exactly one keypress, scoped strictly to the exercise tab
being the thing visible (never a global key-eater); `ctrl+c` still quits
(rule 3 outranks everything), the help overlay's own keyboard outranks this
while open, a focused minibuffer keeps its keys (focus was explicitly
acquired, focus-contract rule 1), and an open guardian console (spec 053, a
whole-body takeover) suppresses the eater entirely — the briefing is not
the thing on screen while the console is, so console keys reach the console
and the briefing survives undismissed. Dismissal lasts for this attach
only; re-attaching shows the briefing (and this mode) again.

| Key | Action |
|---|---|
| any key | dismiss the briefing (consumed — no other effect) |
| `ctrl+c` | quit |

## Mode: help overlay (spec 045-tui-help-overlay, TASK-116)

`?` opens a context-sensitive help overlay from every mode above except
minibuffer focus (there it types like any other character — the table's "no
other key does anything silently" rule wins). Its full content, sections,
tiers, and control table are specified in
[../overlays/help.md](../overlays/help.md) (extracted there, this feature);
this card keeps only the top-level binding so the printable page stays
complete without duplicating the overlay's own reference:

| Key | Action |
|---|---|
| `?` | open help (from any mode except minibuffer focus); while open, dismiss (toggle) |
| `esc` | dismiss (same as `?`) |

## Footer hints per mode

```
global            2 chronicle 3 {{skin.guardian.tab_label}} 4 villagers 5 systems (again: solo) · G console · m ask · space pause · q quit · ? help
                  (scenario worlds append "6 exercise" to the tab list — spec 054)
narrow (no dock visible)   1-5 panes · G console · space pause · q quit · ? help
                  (scenario worlds: "1-5,6 panes" — spec 054)
minibuffer        esc release · ⏎ send · ↑↓ history
inspect           j/k select · J/K scroll detail · space resume · m ask · ? help
villagers roster  j/k select · ⏎ inspect · space pause · q quit · ? help
villagers detail  esc back · space pause · q quit · ? help
look-cursor (cursor focus)  hjkl/arrows move · HJKL jump · c center · ⏎/tab pane · 2-6 tab · esc exit · ? help
look-cursor (pane focus)    j/k select · ⏎ drill · esc back · 2-6 tab · ? help
console           G back · esc back · m ask · space pause · q quit · ? help
ceremony          esc dismiss · q — the world keeps running
postmortem        esc dismiss · q quit
```

Minibuffer's hint carries no `? help`: focused, `?` types into the buffer
(FR-001) — advertising it as a help trigger in that one mode would be wrong.
Minibuffer help is still reachable, just from any other mode's overlay
(`n`/`p` paging above).

## Input-parity doctrine (reorientation decision 8)

Ratified doctrine, binding on this entire corpus, not just this page:

1. **Every action is reachable by both keyboard and mouse.** Every
   control-table row's `keys+mouse` column (contracts/control-table.md)
   states the mouse target directly beside the keyboard binding — `—` only
   for a genuinely display-only control, never as a stand-in for "not done
   yet."
2. **Keyboard is primary and complete.** The keyboard alone can always
   reach 100% of the app's functionality — mouse support is additive
   convenience, never a second, richer control scheme with keyboard-only
   gaps. This corpus's mockups and behavior sections are written keyboard-
   first for exactly this reason.
3. **Rollout is incremental, honestly tracked — and mechanized (spec 073,
   TASK-154).** This doctrine shipped in this same feature (spec 049,
   decision 8) alongside its own first compliant control: chronicle
   jump-to-source (`⏎` · click line — `panels/chronicle.md`) is the corpus's
   first row with a real mouse target, landing keyboard and mouse together
   as the doctrine requires (rule 1). Every other control in
   `docs/design/tui/` still has a key but no mouse target — `internal/tui`
   predates this doctrine entirely for everything else, including spec
   053's `G` (guardian console), `5` (systems tab), and `e` ($EDITOR
   handoff), each recorded as a parity gap on its own owning page rather
   than silently shipped as if display-only. Every panel/overlay page in
   this corpus carries a **"Parity rollout"** note listing its
   keyed-but-mouseless controls rather than silently marking them `—` as if
   display-only; a control graduates out of that note the moment its
   page's control table gains a real mouse target, one control (or one
   page) at a time. This page's own footer/mode-key tables have no mouse
   column of their own (they're the printable card, not a control table) —
   the authoritative per-control mouse status lives on each control's
   owning panel/overlay page, cross-referenced from here by page link. Spec
   054's two new bindings — the exercise tab's `6` and the briefing's
   any-key dismiss — land keyboard-only and are tracked as parity gaps on
   `panels/exercise.md`'s own rollout note, exactly as that page promised
   from birth. Spec 074 (look-cursor) lands its own controls compliantly
   too — click-tile (`panels/map.md`, the map's first real mouse target
   ever) and click-row (`panels/dock.md`'s TILE pane) both ship in the same
   PR as their keyboard bindings, each with a `TestMouseParitySweep` oracle
   entry proving the live dispatch — while leaving the map's pre-existing
   pan/recenter gap exactly as tracked (this feature didn't touch those
   controls, so it didn't close that gap either). This tracking is no
   longer hand-audited alone: `TestMouseParitySweep`
   (`internal/tui/mouseparity_test.go`) parses every canonical-header
   control table under `docs/design/tui/` and fails the build when a
   non-`—` mouse cell has no proven handler (a hand-audited oracle entry
   whose live `tea.MouseMsg` dispatch demonstrates the documented effect)
   or when a page carries a keyed-but-mouseless row with no
   `**Parity rollout**` note. **Graduation contract**: a control leaves its
   page's rollout note only when, in the same PR, it gains (1) a real
   mouse cell in the control table (replacing `—`), (2) a matching entry in
   the sweep's oracle, and (3) that entry's live-dispatch check passing —
   never a documentation-only promotion.

## Binding-selection rules

Two conventions this corpus's key choices follow, made explicit here so a
future page's key choice is principled rather than arbitrary:

- **Mnemonic keys**, when an unclaimed letter matches the concept it
  controls, are preferred over an arbitrary free key: `G`uardian console,
  `p`ostmortem reopen, `d`ecisions toggle (existing, spec 020). `x` (lesson
  dismiss) is the one exception — `d`, `l`, and other lesson-adjacent
  letters were already claimed or read as ambiguous in context, so `x` (a
  conventional "close/dismiss" mnemonic in terminal UIs) was chosen instead
  and is called out here rather than silently deviating from the rule.
  `v` (look-cursor toggle, spec 074) is a second recorded exception: `l` is
  already the cursor's own in-mode right-move key, and "inspect" already
  names the chronicle mode, so the mnemonic reaches one word further —
  `v`iew, the plain-language framing this feature's TILE view itself uses.
- **Reserved seams are wired to a documented no-op**, never left an
  undiscovered gap: before spec 049 filled it, the chronicle's `⏎`
  (jump-to-source, Wave 2/D3) was this corpus's precedent — a key that did
  nothing yet still appeared in this card and its owning page's control
  table, with the future behavior named, so a player who pressed it saw
  documented "nothing yet" rather than wondering whether the app was
  broken. The same posture now applies one level down, inside the built
  feature: an *unlocatable* event's `⏎` is still a no-op, but the actions
  bar (`panels/chronicle.md`) names it live, every time, rather than the
  static card-only promise a truly unbuilt seam relies on.

## Migration notes

- `tab`/`shift+tab` pane cycling may remain as aliases for dock-tab cycling; not
  load-bearing.
- Today's "keys 1–5 swap the whole screen" behavior survives only in the narrow
  fallback ([../pages/solo-views.md](../pages/solo-views.md)).
