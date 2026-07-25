---
title: Page — guardian console
class: page
status: specified
verified_against: c8906da39be3a5b861c2272af37db0a83dcded7a
---

# Page: guardian console

The Staged Cockpit's headline surface (reorientation decisions 1/2, D5): the
guardian conversation elevated to a first-class, full-height page — the
"document-style" reading experience the compact dock tab
([panels/guardian.md](../panels/guardian.md)) deliberately doesn't try to be.
**Not built** — specified spec-before-build for Wave 3.

## Mockup

```
 promptworld — tick 8801 · day 4 08:12 · running · speed 1x (1.0 t/s)
┌─ {{skin.guardian.name}} · charges ⚡⚡· · stage: The Written Word ──────────┐
│                                                                          │
│ you · 08:04                                                             │
│   why is Rowan hoarding wood?                                           │
│                                                                          │
│ {{skin.guardian.epithet}} · 08:04                                       │
│   Rowan's memory holds three nights of Ash letting the fire die.        │
│   Trust toward Ash: −2.                                                 │
│                                                                          │
│ » 08:05  send_vision — landed                                          │
│                                                                          │
│ ┌─ charter · skills ────────────────────────────────────────────────┐  │
│ │ charter.md — player-authored, binds now          [e] edit ($EDITOR)│  │
│ │ skills/ — 2 files, binds from stage 3 (locked)                     │  │
│ └──────────────────────────────────────────────────────────────────┘  │
│                                                                          │
│ ┌─ report card · first-night ─────────────────────────────────────┐    │
│ │ ✓ village survives to dawn · ✓ watch placed before nightfall     │    │
│ └───────────────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────────────────┘
┌──────────────────────────────────────────────────────────────────────────┐
│ ⏎ m — speak with the {{skin.guardian.epithet}}…                         │
└──────────────────────────────────────────────────────────────────────────┘
  G back to map · esc back · m ask · space pause · q quit · ? help
```

## Structure

1. **Header line** — same content as the guardian dock tab's pane header
   (name, charge bank, instruction/capability provenance, stage segment,
   [panels/guardian.md](../panels/guardian.md)) — one shared data source,
   rendered larger here because the page has the room.
2. **Document-style turns** — unlike the compact dock tab's dense
   alternating rows, each turn renders as its own labeled block (speaker +
   timestamp, full unwrapped-except-by-width text, blank-line separated) —
   a chat-document reading experience for extended conversation, not a
   glance. The same special-row vocabulary as the dock tab carries over
   unchanged: `⚡` omen/vision lines, `👁` watch set/released, `⏲` clock
   lines, `»` inline verdict rows (miracle feedback) — one shared special-
   row vocabulary, two renderings (compact vs. document).
3. **Charter/skills read surface** (decision 2) — a bordered sub-panel
   showing `charter.md`'s provenance (default/player-authored/preset-locked)
   and binding status, plus the skill file count and its own binding status
   (locked below stage 3, per [[curriculum-ladder]]). An `[e] edit
   ($EDITOR)` action shells out to the player's `$EDITOR` on the real file;
   there is **no in-TUI text editor** (decision 2's explicit ruling). On
   return from `$EDITOR`, if the file changed, the console shows a
   one-line confirmation: **"charter changed — next turn binds it"** —
   honest about the re-read-every-turn timing (edits are live by
   construction, [[metatron]]), never implying the running turn changes
   retroactively.
4. **Report-card cards** (D5) — inline cards at natural stopping points
   (run end, pause, exercise resolution), rendered by the SAME shared
   report-card renderer [overlays/postmortem.md](../overlays/postmortem.md)
   uses for a scored run's takeover — one spec, authored fully on that
   page (FR-018's scored/ambient ruling lives there); this page only
   composes it inline between turns rather than as a takeover.

## Composer

The console's input is the **same minibuffer** every other page uses
([panels/minibuffer.md](../panels/minibuffer.md)) — `patterns/focus-
contract.md`'s "the minibuffer is the only focusable input" ground rule is
NOT relaxed for this page. The console presents the minibuffer as its
composer (visually paired directly beneath the turn stream, no separate
input widget), but its states (dormant/focused/busy), its focus contract,
and its IPC transport are unchanged from `panels/minibuffer.md`. This is a
deliberate scope constraint (`patterns/focus-contract.md`'s "exactly one
client" carries forward): the console elevates *reading*, not the input
mechanism.

## Navigation (new, specified here)

The guardian console is **not** reached through the dock's solo-zoom seam
(`pages/solo-views.md`'s state machine is unchanged and does not include
it) — it is deliberately a first-class destination, not a zoomed tab. This
page specifies a new global key: **`G`** (shift-g) opens the guardian
console from any mode (home, solo, narrow); `1`, `G` (toggle), or `esc`
(when nothing else is focused) return to whatever was open before. The
guardian dock tab ([panels/guardian.md](../panels/guardian.md)) remains the
always-available compact view inside the ordinary composite — the console
is the deliberate elevated destination for extended conversation and
charter/skill authoring, not a replacement for it.

## Stage defaults

Per `patterns/stage-defaults.md` (the authority): **reachable at every
stage** via its own key — decision 3's "everything remains reachable"
applies here exactly as everywhere else. Its *content* is stage-shaped only
through the existing spec-046 stage ceiling and charter lock already
governing `panels/guardian.md` — no additional default-visibility rule is
introduced for this page itself.

## Linear-stream / CLI projection (D1)

`promptworld metatron <dir> [message…]` ([[metatron]]) is the existing,
already-shipped CLI equivalent of a console turn — a non-TUI player can
converse with the guardian, and `promptworld metatron` (no message) can
surface status/provenance the same way `metatron_status` does. The charter/
skills read surface's content is literally the on-disk `charter.md`/
`skills/*.md` files — always readable and editable directly, `$EDITOR`
handoff or not. This page adds no capability a linear client lacks; it only
adds a richer TUI presentation of what already exists.

## Control table

| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
| open/close console | closed · open | new page-level state | `unbuilt (wave 3)` | `G` (open/toggle) · — | reorient decisions 1/2 | — |
| document-style turn block | you · guardian · special-row (⚡/👁/⏲/») | `Model.transcript` (shared with `panels/guardian.md`) | `unbuilt (wave 3)` | — | reorient decision 1 | `skin.guardian.epithet` |
| charter/skills read surface | default · player-authored · locked | `metatron.Status` (provenance, binding, skill count) | `unbuilt (wave 3)` | — (display-only) | reorient decision 2 | — |
| `$EDITOR` handoff | idle · shelled-out · returned | on-disk `charter.md`/`skills/*.md` | `unbuilt (wave 3)` | `e` · — | reorient decision 2 | — |
| "charter changed — next turn binds it" confirmation | absent · shown once | file mtime/content diff on `$EDITOR` return | `unbuilt (wave 3)` | — | reorient decision 2 | — |
| report-card card (inline) | absent · shown | shared with `overlays/postmortem.md`'s report-card renderer | `unbuilt (wave 3/4)` | — | reorient D5 | — |
| composer | dormant · focused · busy | `panels/minibuffer.md` (same component, unchanged) | `minibufferView` (existing) | `m` focus · — | TASK-34 (reused) | `skin.guardian.epithet` |

**Parity rollout**: `G` (open console) and `e` ($EDITOR handoff) have no
mouse target — recorded from birth as a parity gap, per decision 8's
incremental rollout (formal doctrine in `patterns/keymap.md`, T024).
