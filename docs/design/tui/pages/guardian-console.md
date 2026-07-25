---
title: Page — guardian console
class: page
status: shipped
verified_against: 39e2407850ef4b4e8493846e28b12b5a445a0b39
sources:
  - internal/tui/views.go
  - internal/tui/tui.go
---

# Page: guardian console

The Staged Cockpit's headline surface (reorientation decisions 1/2, D5): the
guardian conversation elevated to a first-class, full-height page — the
"document-style" reading experience the compact dock tab
([panels/guardian.md](../panels/guardian.md)) deliberately doesn't try to be.
**Built** (spec 053, TASK-125) — the report-card CARD content is the one
exception: D5 assigns it to TASK-115/127, so the control table below keeps
that one row `unbuilt (wave 3/4)`, naming the shipped composition seam
(`consoleCard`, `Model.consoleCards`, `views.go` `consoleView`) it will plug
into.

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

The mockup's `· 08:04` timestamp suffixes are representative, not literal:
`Model.transcript` (the shared data source, `internal/tui/tui.go`) carries
no per-entry timestamp in this client — a turn renders as a bare `you` /
`{{skin.guardian.epithet}}` label with no time suffix (the honesty rule:
never invent what the data doesn't carry). The `»` verdict row similarly
renders through the compact tab's own label (`"note"`, styled as telemetry,
`classifyTranscriptLine`) rather than the literal glyph — one shared
classification, reused verbatim, not a second vocabulary.

## Structure

1. **Header line** (`guardianHeaderLine`, `internal/tui/views.go`) — same
   content as the guardian dock tab's pane header (charge bank, instruction/
   capability provenance, stage segment,
   [panels/guardian.md](../panels/guardian.md)) — the exact same string, not
   a re-derivation, so the two renderings can never silently disagree.
2. **Document-style turns** (`consoleTurnLines`) — unlike the compact dock
   tab's dense alternating rows, each conversational turn renders as its own
   labeled block (speaker label, full width-wrapped text, blank-line
   separated) — a chat-document reading experience for extended
   conversation, not a glance. The same special-row vocabulary as the dock
   tab carries over unchanged, via the same `classifyTranscriptLine`
   classification the compact tab uses: `⚡` omen/vision lines, `👁` watch
   set/released, `⏲` clock lines, and inline verdict rows (working feedback,
   labeled `note`) — one shared special-row vocabulary, two renderings
   (compact vs. document). Tail-anchored scrollback (`J`/`K`,
   `consoleScrollWindow`), reset on close.
3. **Card seam** (`consoleCardLines`) — the composition point between the
   turn stream and the read surface; see item 4 below.
4. **Charter/skills read surface** (decision 2, `charterReadSurfaceLines`/
   `charterReadSurfaceBox`) — a bordered sub-panel showing `charter.md`'s
   provenance (default/player-authored/preset-locked) and binding status,
   plus the skill file count and its own binding status (locked below stage
   3, per [[curriculum-ladder]]) — sourced entirely from the same
   `metatron.Status` fields the compact tab's header already reads (no
   client-side file parsing). An `[e] edit ($EDITOR)` hint appears on the
   charter line when it binds; there is **no in-TUI text editor** (decision
   2's explicit ruling). On return from `$EDITOR`, if the file's content
   hash changed, the console shows a one-line confirmation:
   **"charter changed — next turn binds it"** — honest about the
   re-read-every-turn timing (edits are live by construction, `[[metatron]]`),
   never implying the running turn changes retroactively.
5. **Report-card cards** (D5) — **not built this feature**: the inline
   composition slot exists (`consoleCard` interface, `Model.consoleCards`,
   composed between the turn stream and the read surface) and is always
   empty. TASK-127's shared report-card renderer
   ([overlays/postmortem.md](../overlays/postmortem.md)) and TASK-115's
   stopping-point production (run end, pause, exercise resolution) are the
   producers that plug into this seam later — the chronicle-⏎
   reserved-seam precedent: an honest, already-wired attachment point beats
   a placeholder that renders nothing meaningful.

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

## Navigation

The guardian console is **not** reached through the dock's solo-zoom seam
(`pages/solo-views.md`'s state machine is unchanged and does not include
it) — it is deliberately a first-class destination, not a zoomed tab: a
`Model.console` page-level flag branching at the top of `View()`, ahead of
the widescreen/narrow fork (`internal/tui/views.go`). **`G`** (shift-g) opens
the guardian console from home, solo, and narrow; `1`, `G` (toggle), or
`esc` (when nothing else is focused) return to whatever was open before.
`dockTab`/`solo`/`active` are never touched while the console is open, so
they already ARE the "prior view" — nothing is snapshotted or restored
beyond the flag itself. The guardian dock tab
([panels/guardian.md](../panels/guardian.md)) remains the always-available
compact view inside the ordinary composite — the console is the deliberate
elevated destination for extended conversation and charter/skill authoring,
not a replacement for it.

**Scoping note**: inspect mode (paused chronicle, `chronJumpLast`) and the
villagers tab (`villJumpLast`) already bind `G` to their own jump-to-last
before this feature existed — both are checked ahead of the global mode in
`handleKey`'s dispatch, so they keep their own `G` unchanged (FR-001 itself
names only "home, solo, and narrow" as where the console's `G` applies).
While the console is open, its own key layer (`handleConsoleKey`) owns `G`/
`1`/`esc`/`m`/`e`/`J`/`K`; every other global key (`space`/`q`/`[`/`]`/
`2`-`5`/pan/`a`/`t`/`r`) still reaches `handleGlobalKey` underneath, matching
the console's own footer hints. The esc-release chain gains one layer:
minibuffer → villager detail → **console** → solo → home.

## Stage defaults

Per `patterns/stage-defaults.md` (the authority): **reachable at every
stage** via its own key — decision 3's "everything remains reachable"
applies here exactly as everywhere else. Its *content* is stage-shaped only
through the existing spec-046 stage ceiling and charter lock already
governing `panels/guardian.md` — no additional default-visibility rule is
introduced for this page itself.

## Linear-stream / CLI projection (D1)

`promptworld guardian <dir> [message…]` (`[[metatron]]`; the pre-052
`metatron` spelling stays a hidden compat alias) is the existing,
already-shipped CLI equivalent of a console turn — a non-TUI player can
converse with the guardian, and `promptworld guardian` (no message) can
surface status/provenance the same way `metatron_status` does. The charter/
skills read surface's content is literally the on-disk `charter.md`/
`skills/*.md` files — always readable and editable directly, `$EDITOR`
handoff or not. This page adds no capability a linear client lacks; it only
adds a richer TUI presentation of what already exists.

## Control table

| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
| open/close console | closed · open | `Model.console` | `consoleView`, `openConsole`/`closeConsole` | `G` (open/toggle) · — | spec 053 | — |
| document-style turn block | you · guardian · special-row (⚡/👁/⏲/note) | `Model.transcript` (shared with `panels/guardian.md`) | `consoleTurnLines` | — | spec 053 | `skin.guardian.epithet` |
| scrollback | tail (0) · scrolled N | `Model.consoleScroll` | `consoleScrollWindow` | `J`/`K` · — | spec 053 | — |
| card seam (composition point) | always empty this feature | `Model.consoleCards` (`consoleCard` interface) | `consoleCardLines` | — | spec 053 (seam only) | — |
| charter/skills read surface | default · player-authored · preset-locked; skills bound · locked | `metatron.Status` (`CharterDefault`/`CharterLocked`/`CharterPreset`/`Skills`/`SkillsLocked`) | `charterReadSurfaceLines`/`charterReadSurfaceBox` | — (display-only) | spec 053 | — |
| `$EDITOR` handoff | idle · shelled-out · returned | on-disk `charter.md` (content hash, pre/post) | `startEditorHandoff` (`tea.ExecProcess`) | `e` · — | spec 053 | — |
| "charter changed — next turn binds it" confirmation | absent · shown once · error notice | pre/post content-hash comparison (`editorRoundTripMsg`) | `Model.consoleNotice` | — | spec 053 | — |
| report-card card (inline) | absent · shown | shared with `overlays/postmortem.md`'s report-card renderer | `unbuilt (wave 3/4)` — plugs into `consoleCard`/`Model.consoleCards` above | — | reorient D5 | — |
| composer | dormant · focused · busy | `panels/minibuffer.md` (same component, unchanged) | `minibufferView` (existing) | `m` focus · — | TASK-34 (reused) | `skin.guardian.epithet` |

**Parity rollout**: `G` (open console) and `e` ($EDITOR handoff) have no
mouse target — recorded from birth as a parity gap, per decision 8's
incremental rollout (formal doctrine in `patterns/keymap.md`).
