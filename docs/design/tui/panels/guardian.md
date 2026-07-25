---
title: Panel — guardian (fiction-layer tab content)
class: panel
status: shipped
verified_against: cb89a4c7811962243ac907e0aeed43619b4d4f2d
sources:
  - internal/tui/views.go
  - internal/tui/tui.go
  - internal/tui/decisions.go
---

# Panel: guardian

The dock's fiction-layer tab (D10: the skin boundary is a file boundary — this
page never renders engine telemetry; that content lives on
[systems.md](systems.md)). Split from the pre-v2 monolithic `dock.md`'s
"metatron" tab section. Container chrome (tab row, badges, tab-switch keys)
is [dock.md](dock.md); the *input* line is
[minibuffer.md](minibuffer.md) — this page is history + status only.

**Re-verified, spec 053/TASK-125**: before this feature, this claim held for
the widescreen dock tab (`dockTabContent`'s `paneMetatron` case only ever
called the transcript body — it never rendered telemetry) but not yet for
the narrow fallback's combined pane (`metatronView` used to also render the
provider table/spend/horizon block inline, below the standing-orders block).
Spec 053 relocated that block out to [systems.md](systems.md)'s
`systemsView`, so this page's "never renders engine telemetry" claim is now
byte-true for both layouts — the pane header line itself also moved to a
function shared verbatim with [pages/guardian-console.md](../pages/guardian-console.md)
(`guardianHeaderLine`), so it can never drift from what the console page
shows in its own header line.

## Mockup

```
┌─ chronicle │ {{skin.guardian.tab_label}} (active/upper) │ villagers ┐
├───────────────────────────────────────────────────────────┤
│ {{skin.guardian.name}} · charges ⚡⚡· · stable (charter.md)│
│  · stage: The Written Word (charter locked to tutor)       │
│                                                             │
│ you    why is Rowan hoarding wood?                         │
│ {{skin.guardian.epithet}}  Rowan's memory holds three       │
│        nights of Ash letting the fire die. Trust: −2.      │
│ »      send_vision — landed                                │
│ 👁      watch set (ord-3): "gru sighted within 5 tiles"     │
│ ⏲      paused                                               │
│ you    what does ash want                                  │
│ {{skin.guardian.epithet}}  ⋮ thinking…                      │
│                                                             │
│ 👁 standing orders (1)                                      │
│   ord-3~ [console · day 4 · watching] "gru sighted…"       │
└───────────────────────────────────────────────────────────┘
```

## Structure

1. **Header line** — `{{skin.guardian.name}} · charges <⚡…·…>` (the charge
   bank, filled/empty glyphs to `sim.MetatronChargeCap`), followed, once a
   world has an active charter/instruction surface, by the spec-021
   instruction/capability provenance summary: charter flavor + `(charter.md)`,
   a skill count when non-zero, the granted-tool short-form summary (quiet for
   a full-grant default world, `tools: none` for conversation-only), and,
   since spec 046, a `stage: <skin name>` segment appending
   `(charter locked to <preset>)` while the stage-1 instruction lock binds —
   empty (header byte-identical to pre-046) for a pre-ladder/ungated world.
2. **Transcript** — alternating `you` / `{{skin.guardian.epithet}}` rows, text
   wrapped to tab width; newest at bottom, opens scrolled to bottom. A
   question in flight renders one `{{skin.guardian.epithet}} ⋮ thinking…` row.
   Special rows interleave in landed order (never batched separately):
   - **`⚡` vision/omen line** — a landed `send_vision`/`send_omen` act, "⚡
     {{skin.guardian.epithet}} → target(s): text".
   - **`👁` watch set / watch released** — a standing order placed or
     cancelled (spec 029).
   - **`⏲` clock line** — a landed `pause`/`start`/`adjust_speed` meta tool
     call.
   - **`»` inline verdict row (miracle feedback, spec 016/020)** — every
     turn-scoped tool call the loop saw (landed or refused) appends one
     `» tool — phrase` row styled as telemetry (dim), distinct from
     `you`/`{{skin.guardian.epithet}}` rows — the SAME plain-language
     `verdictGlossary` the villagers decisions sub-view uses
     ([villagers.md](villagers.md)), so a refused `work_miracle` or `send_omen`
     reads in the transcript exactly as it would in a decision trace.
   - **Unreachable** — a failed console call renders one error-styled
     "the {{skin.guardian.epithet}} is unreachable: `<err>`" line below the
     transcript rather than as a transcript row.
3. **Standing-orders block** (spec 029) — present only while ≥1 order stands:
   a `👁 standing orders (n)` header, then one compact row per order — id, a
   `~` fuzzy marker, origin, remaining game-day, status, and condition text.
4. **Minibuffer** ([minibuffer.md](minibuffer.md)) — nested at the bottom of
   this pane in the narrow fallback rendering; a standalone chrome row in the
   widescreen composite. Either way, the guardian tab and the minibuffer are
   governed by the same focus contract
   ([../patterns/focus-contract.md](../patterns/focus-contract.md)).

## Behavior

- **Reply arrival while this tab isn't visible**: the reply still streams
  into the transcript state; the dock badges `{{skin.guardian.tab_label}} •`
  instead of stealing the selected tab ([dock.md](dock.md)).
- Scrollback is newest-at-bottom; selecting this tab opens scrolled to the
  bottom.
- All fiction strings on this page — the guardian's proper name, the
  transcript epithet, the busy/unreachable copy — resolve through the skin
  tokens defined in [../patterns/skin-tokens.md](../patterns/skin-tokens.md);
  this is the file the D10 skin boundary protects (systems.md carries none).

## Control table

| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
| pane header — name + charge bank | 0–3 charges filled | `Status.Clock.MetatronCharges`, `sim.MetatronChargeCap` | `guardianHeaderLine` (shared verbatim with the console page, spec 053) | — (display-only) | TASK-12 | `skin.guardian.name` |
| instruction/capability provenance segment | quiet (full grant) · skill count · tools summary · `tools: none` | `metatron.Status` (`ManifestDefault`, `GrantedTools`) | `consoleToolsSummary` | — | spec 021 | — |
| stage segment | present (locked) · present (unlocked) · absent (pre-ladder) | `Status.Stage`/`CharterLocked`/`CharterPreset` | `consoleStageSummary` | — | spec 046 | — |
| transcript row — you | text | player input history | `transcriptRowLines`/`classifyTranscriptLine` | — | TASK-34 | — |
| transcript row — guardian reply | text · `⋮ thinking…` (busy) | `Model.transcript`, `Model.mbBusy` | `transcriptRowLines`/`classifyTranscriptLine` | — | TASK-34 | `skin.guardian.epithet` |
| transcript row — omen/vision (`⚡`) | landed | landed `metatron.nudged` | `tui.go` transcript append | — | spec 029 | — |
| transcript row — watch set/released (`👁`) | placed · released | `metatron.order_placed`/`order_cancelled` | `tui.go` transcript append | — | spec 029 | — |
| transcript row — clock (`⏲`) | landed | landed `pause`/`start`/`adjust_speed` | `tui.go` transcript append | — | spec 029 | — |
| transcript row — inline verdict (`»`) | landed · refused | `cog.tool_call` (turn-scoped) | `metatronVerdictRow`, `callLine`, `verdictGlossary` | — | spec 016/020 | — |
| unreachable notice | shown · hidden | `Model.mbErr` | `metatronView` | — | TASK-34 | `skin.guardian.epithet` |
| standing-orders block | absent · n rows | `Status.Orders` | `orderStatusLines` | — | spec 029 | — |

**Parity rollout**: no control on this page has a mouse target today (same
gap as every other pre-decision-8 surface — see dock.md); tracked here rather
than omitted, formal doctrine lands in `patterns/keymap.md` (T024).
