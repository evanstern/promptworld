---
title: Pattern — keymap
class: pattern
status: shipped
verified_against: d0dd109a73faf66eff14c81d748f155f424b801b
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
([focus-contract.md](focus-contract.md) rule 3): minibuffer → villager detail →
solo → home — each press of `esc` releases exactly one layer. With no world state
loaded (or an empty roster) `j/k/g/G/⏎` are strict no-ops.

## New global keys (specified, unbuilt — this feature)

Three new global keys this feature's new-surface pages introduce; none
exist in `internal/tui` yet (`unbuilt` in each page's own control table):

| Key | Action | Specified in |
|---|---|---|
| `G` | open the guardian console (toggle back with `G` again, or `1`/`esc`) | [pages/guardian-console.md](../pages/guardian-console.md) |
| `x` | dismiss the active lesson row | [panels/lesson-row.md](../panels/lesson-row.md) |
| `p` | reopen the postmortem takeover (only while the run has ended) | [overlays/postmortem.md](../overlays/postmortem.md) |

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
3. **Rollout is incremental, honestly tracked.** This doctrine shipped in
   this same feature (spec 049, decision 8) alongside its own first
   compliant control: chronicle jump-to-source (`⏎` · click line —
   `panels/chronicle.md`) is the corpus's first row with a real mouse
   target, landing keyboard and mouse together as the doctrine requires
   (rule 1). Every other control in `docs/design/tui/` still has a key but
   no mouse target — `internal/tui` predates this doctrine entirely for
   everything else. Every panel/overlay page in this corpus carries a
   **"Parity rollout"** note listing its keyed-but-mouseless controls rather
   than silently marking them `—` as if display-only; a control graduates
   out of that note the moment its page's control table gains a real mouse
   target, one control (or one page) at a time. This page's own footer/
   mode-key tables have no mouse column of their own (they're the printable
   card, not a control table) — the authoritative per-control mouse status
   lives on each control's owning panel/overlay page, cross-referenced from
   here by page link.

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
- Today's "keys 1–4 swap the whole screen" behavior survives only in the narrow
  fallback ([../pages/solo-views.md](../pages/solo-views.md)).
