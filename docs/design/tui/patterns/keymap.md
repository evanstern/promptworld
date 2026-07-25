---
title: Pattern — keymap
class: pattern
status: shipped
verified_against: c8906da39be3a5b861c2272af37db0a83dcded7a
sources:
  - internal/tui/tui.go
  - internal/tui/help.go
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
| `2` / `3` / `4` | select dock tab chronicle / `{{skin.guardian.tab_label}}` / villagers; **same key again** → solo zoom; again → back home |
| `m` | focus the minibuffer |
| `space` | pause / resume the clock |
| `[` / `]` | speed down / up |
| `←↑↓→` | pan the map |
| `c` | recenter camera (resume following) |
| `r` | chronicle: toggle raw ↔ narrated |
| `a` / `t` | chronicle: filter by agent / thread |
| `q` | quit |
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
`⏎` needed to see it — panels/chronicle.md "Mode 2"); `⏎` is reserved for a
future jump-off actions bar (contract §5 "Extension point") rather than
toggling expansion, and `J`/`K` scroll the pane when its content overflows.

| Key | Action |
|---|---|
| `j` / `k` | select next / previous event (also resets detail pane scroll) |
| `g` / `G` | jump to first / last (also resets detail pane scroll) |
| `J` / `K` | scroll the detail pane down / up |
| `⏎` | reserved — no-op today (future jump-off actions) |
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
([focus-contract.md](focus-contract.md) rule 3): minibuffer → villager detail →
solo → home — each press of `esc` releases exactly one layer. With no world state
loaded (or an empty roster) `j/k/g/G/⏎` are strict no-ops.

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
global            2 chronicle 3 {{skin.guardian.tab_label}} 4 villagers (again: solo) · m ask · space pause · q quit · ? help
minibuffer        esc release · ⏎ send · ↑↓ history
inspect           j/k select · J/K scroll detail · space resume · m ask · ? help
villagers roster  j/k select · ⏎ inspect · space pause · q quit · ? help
villagers detail  esc back · space pause · q quit · ? help
```

Minibuffer's hint carries no `? help`: focused, `?` types into the buffer
(FR-001) — advertising it as a help trigger in that one mode would be wrong.
Minibuffer help is still reachable, just from any other mode's overlay
(`n`/`p` paging above).

## Migration notes

- `tab`/`shift+tab` pane cycling may remain as aliases for dock-tab cycling; not
  load-bearing.
- Today's "keys 1–4 swap the whole screen" behavior survives only in the narrow
  fallback ([../pages/solo-views.md](../pages/solo-views.md)).
