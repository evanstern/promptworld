---
title: Panel — villagers (roster, detail, decisions)
class: panel
status: shipped
verified_against: a30ee798ff6cc6316256d7833aead1e8a4c9a849
sources:
  - internal/tui/views.go
  - internal/tui/decisions.go
---

# Panel: villagers

The dock's villager-inspector tab (renamed from "souls" — spec 015/TASK-56),
split out of the pre-v2 monolithic `dock.md` into its own page (this
feature). A three-view drill-down — roster → detail → decisions — one
renderer per view, rendering width-aware exactly like every other dock tab
(wrap/condense columns; drop the least important column first when narrow).
No fiction content here — villager names/inventory/memories are simulation
data, not guardian-skin strings.

## Mockup — roster (default)

Today's per-villager summary line plus a selection cursor glyph (`▌`) on the
selected row. Wide (≥40 cols) keeps name/status/goal/position, the needs
bars, and the full carried-inventory line; narrow drops to name + status +
health. Rows beyond the height budget are dropped from the bottom (never a
partial row).

```
┌─ chronicle │ {{skin.guardian.tab_label}} │ VILLAGERS ┐
├───────────────────────────────────┤
│▌ Ash    awake · chop · (12,9)     │
│    health █████ food ███░░ ...    │
│    carry 2w 0st ... · spear 1(2)  │
│  Rowan  asleep · idle · (4,3)     │
│    health ████░ food ████░ ...    │
│    carry 0w 0st ...               │
└───────────────────────────────────┘
```

- `j`/`k` move the cursor, `g`/`G` jump first/last, `⏎` opens the selected
  villager's detail view.
- Selection is clamped to the roster and survives tab switches, reconnects,
  and world restarts.

## Mockup — detail (after `⏎`)

Sections render in a fixed priority order and truncate from the **bottom**
when height runs short — memories are shed first, identity/objective/
inventory are never pushed off-screen.

```
┌─ chronicle │ {{skin.guardian.tab_label}} │ VILLAGERS ┐
├───────────────────────────────────┤
│ ASH                                │
│                                    │
│ Ash · awake · (12,9)               │
│ health █████ food ███░░ rest ████░ │
│                                    │
│ objective: chop → (13,9) (current) │
│                                    │
│ inventory:                         │
│   wood 2                          │
│   spears 1 (uses left: 2)          │
│                                    │
│ memories                           │
│ 08:41 (4★) chopped a fine oak      │
│ 07:02 (2★) shared bread with Rowan │
└───────────────────────────────────┘
```

1. **identity/vitals** — name, awake/asleep/dead status (dead villagers still
   render — the morgue view), position, needs bars.
2. **objective** — the active `Intent.Goal` marked *current*; else the
   reducer-stamped `Agent.LastGoal` + tick marked *last*; else "no objective
   yet".
3. **inventory** — every carried kind itemized with counts, spear wear
   included; an empty pack says so plainly.
4. **beliefs/narrative** — consolidated beliefs and self-narrative, shown
   only when nightly consolidation has produced them.
5. **memories** — episodic memories, most recent first, bounded to whatever
   height remains.

## Mockup — decisions (after `d`, spec 020/TASK-63)

```
┌─ chronicle │ {{skin.guardian.tab_label}} │ VILLAGERS ┐
├───────────────────────────────────┤
│ ASH · decisions                    │
│                                    │
│ 08:11 cog.thought — hungry, cold   │
│  1. send_vision — landed          │
│  outcome: landed                  │
│                                    │
│ 07:40 cog.thought — thirsty        │
│  didn't think because suppressed  │
└───────────────────────────────────┘
```

The villager's recent cognitions as causal chains, most-recent-first: a
when/class header, the stimulus line, each tool call as
`ordinal. tool — phrase (reason)`, and the terminal outcome (or an explicit
`in progress — no outcome yet` marker); router suppressions render as one
`didn't think because …` entry. Verdicts and outcomes render ONLY through the
sweep-tested plain-language `verdictGlossary` — raw enum strings never reach
the screen.

## Behavior

- `esc` closes the detail view back to the roster (roster selection
  preserved) — before the existing solo-release chain, same "esc always
  releases" ordering as everywhere else
  ([../patterns/focus-contract.md](../patterns/focus-contract.md) rule 3).
  From decisions, `esc` closes decisions first, then detail, then solo, then
  home — one layer per press.
- The detail and decisions views update live as world events arrive — they
  render straight from the replica each frame; the decision-trace projection
  is client-side (`decisions.go`), bounded (`decisionChainCap` 20 chains per
  agent), and resets on reconnect like the replica.
- Selection state survives tab switches and is clamped on reconnect.

## Control table

| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
| roster row | selected · unselected | `replica.Agents` | `villagerRosterBody` | `j`/`k` select · — | spec 015 | — |
| roster selection jump | first · last | `Model.villSel` | `villagerRosterBody` | `g`/`G` · — | spec 015 | — |
| open detail | roster → detail | `Model.villDetail` | `villagersBody` | `⏎` · — | spec 015 | — |
| identity/vitals section | — | `sim.Agent` | `villagerIdentitySection` | — | spec 015 | — |
| objective section | current · last · none | `Agent.Intent.Goal`/`LastGoal` | `villagerObjectiveSection` | — | spec 015 | — |
| inventory section | itemized · empty | `sim.Agent` inventory fields | `villagerInventorySection` | — | spec 015 | — |
| beliefs/narrative section | present · omitted | nightly consolidation output | `villagerBeliefsSection` | — | spec 015 | — |
| memories section | itemized · empty | episodic memory store | `villagerMemoriesLines` | — | spec 015 | — |
| close detail | detail → roster | `Model.villDetail` | `villagersBody` | `esc` · — | spec 015 | — |
| decisions sub-view toggle | closed · open | `Model.villDecisions` | `villagerDecisionsBody` | `d` · — | spec 020 | — |
| decision chain row | landed · refused · in progress · suppressed | client-side decision-trace projection (`Model.traces`) | `renderDecisionChain`, `decisionOutcomeLine`, `verdictGlossary` | — | spec 020 | — |
| decisions scroll | — | `Model.villDecisions` scroll offset | `villagerDecisionsBody` | `j`/`k` · — | spec 020 | — |

**Parity rollout**: no control on this page has a mouse target today; tracked
here rather than omitted (decision 8, formal doctrine in `patterns/keymap.md`,
T024).
