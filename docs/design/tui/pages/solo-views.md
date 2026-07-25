---
title: Page — solo views + narrow fallback
class: page
status: shipped
verified_against: c8906da39be3a5b861c2272af37db0a83dcded7a
sources:
  - internal/tui/views.go
  - internal/tui/tui.go
---

# Page: solo views + narrow fallback

Two ways the home composite gets replaced: deliberately (solo zoom) and by
necessity (terminal too narrow).

## Solo zoom

Pressing a dock tab's key **twice** zooms that tab to full width (first press
selects the tab in the dock; second press, while it is already selected,
goes solo). The map is never solo'd separately — it is already the primary,
always-visible region of the home composite, so `1` has one job: return home
(see patterns/keymap.md: "on home: map is already primary").

```
state machine (per key k ∈ {2,3,4}):
  home, tab≠k   --k-->  home, tab=k
  home, tab=k   --k-->  solo(k)
  solo(k)       --k-->  home, tab=k          (same key toggles back)
  solo(k)       --1 or esc-->  home, tab=k
  solo(k)       --k2 (k2≠k)-->  solo(k2)     (switch which tab is solo'd,
                                               stay solo — "same component,
                                               two widths" holds even mid-zoom)
```

Implementation note (TASK-34): the state machine above only specifies the
*same-key* transitions out of `solo(k)`. A different dock-tab key pressed
while solo'd switches which tab is zoomed rather than silently returning home
— solo stays "the dock at full width" ([../panels/dock.md](../panels/dock.md)),
so tab-switching keeps working at every width, and only the same key or
`1`/`esc` drops back to the composite.

**New tabs/pages this feature's later phases add**: the state machine and
"same component, two widths" rule extend to any future dock tab with no
change here — `panels/systems.md` (Wave 2–3) and `panels/exercise.md`
(Wave 4) both reuse this exact seam once they exist as selectable tabs.
[pages/guardian-console.md](guardian-console.md) (Wave 3, authored in Phase 4
of this feature) is **not** reached through this mechanism — it is a
full-height page of its own, not a dock tab, so it has its own navigation
(specified on its own page when authored).

### Mockup — solo chronicle (`2` `2`)

```
 promptworld · attached · day 4 · 08:12 · 1×                          PAUSED
 ┌─ CHRONICLE · raw · paused — j/k select · ⏎ expand · r narrated ──────────┐
 │ #1198 08:09 agent.talked              {"a":"Sable","b":"Birch"}          │
 │ #1201 08:11 social.conversation_turn  {"Ash"→"Rowan"} "the fire's low ag…│
 │▌#1202 08:11 social.conversation_turn  ◂ expanded                        ▐│
 │▌  {                                                                     ▐│
 │▌    "seq": 1202, "tick": 8846,                                          ▐│
 │▌    "type": "social.conversation_turn",                                 ▐│
 │▌    "payload": {                                                        ▐│
 │▌      "conv": 102,                                                      ▐│
 │▌      "speaker": 1,     // Rowan                                        ▐│
 │▌      "listener": 0,    // Ash                                          ▐│
 │▌      "text": "I stacked wood at dawn, ask Birch"                       ▐│
 │▌    }                                                                   ▐│
 │▌  }                                                                     ▐│
 │ #1203 08:12 social.rumor_told         {"Birch"→"Sable"} "ash lets the f…"│
 └───────────────────────────────────────────────────────────────────────────┘
 ┌─────────────────────────────────────────────────────────────────────────────┐
 │ ⏎ m — speak with the {{skin.guardian.epithet}}…                            │
 └────────────────────────────────────────────────────────────────────────────┘
  2 back to map · space resume · q quit
```

(This mockup predates the always-on detail pane — TASK-60/spec 018 replaced
the `⏎`-expand interaction it shows with the always-on detail split; see
[../panels/chronicle.md](../panels/chronicle.md) "Mode 2" for the current,
accurate inspect-mode rendering. Kept here only as the illustrative solo-at-
full-width framing; the chronicle content itself follows chronicle.md, not
this mockup.)

### Solo rules

- Solo renders the **same component** as the dock tab, just wider — one
  implementation, two widths ([../panels/chronicle.md](../panels/chronicle.md),
  [../panels/guardian.md](../panels/guardian.md),
  [../panels/systems.md](../panels/systems.md) once built,
  [../panels/villagers.md](../panels/villagers.md)). No solo-only features.
- The minibuffer and footer persist in every solo view; the map's live state
  keeps updating underneath and is intact on return.
- Tab state (scroll position, filters, selection) survives the round trip
  home → solo → home.

## Narrow fallback

Below the widescreen breakpoint ([../patterns/layout.md](../patterns/layout.md)),
the app renders **today's single-pane UI unchanged**: header + tab bar + one
active pane + footer, keys `1–4` swap panes exactly as the current
`internal/tui` does.

```
 promptworld · day 4 · 08:12 · 1×
 [ map ] chronicle  guardian  villagers
 ┌───────────────────────────────────┐
 │ ~ ~ " ♠ ♠ A ♠ " . . ⌂ . B . .     │
 │ ~ . . ᴥ . " . . . S . . " " .     │
 └───────────────────────────────────┘
  1-4 panes · space pause · q quit
```

- The two focus-contract fixes still apply in fallback mode: the focus
  contract ([../patterns/focus-contract.md](../patterns/focus-contract.md))
  governs the guardian pane's input, and the chronicle grammar
  ([../patterns/chronicle-grammar.md](../patterns/chronicle-grammar.md))
  formats the feed. Layout is the only thing that degrades.
- Crossing the breakpoint (resize) swaps layouts live; no state is lost.
- Narrow-fallback behavior for the reorientation's new chrome (villager
  strip, lesson row, guardian strip) is ruled in
  [../patterns/layout.md](../patterns/layout.md) (research.md R3): the
  guardian strip and lesson row are carried, the villager strip is not
  (folds to the header count badge instead).

## Control table

| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
| solo zoom (widescreen) | home(tab) · solo(tab) | `Model.solo`, `Model.dockTab` | `selectTab`, `soloPanelView` | same key twice · — | TASK-34 | — |
| solo → home | solo(k) · home | `Model.solo` | `selectTab` | `1`/`esc` · — | TASK-34 | — |
| solo tab switch | solo(k) · solo(k2) | `Model.dockTab` | `selectTab` | different tab key · — | TASK-34 | — |
| narrow pane switch | map · chronicle · guardian · villagers | `Model.active` | `narrowView`, `tabsView` | `1`/`2`/`3`/`4` · — | TASK-34 | — |

**Parity rollout**: every control above has a key but no mouse target today;
tracked here rather than omitted (decision 8, formal doctrine in
`patterns/keymap.md`, T024).
