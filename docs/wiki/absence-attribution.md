---
name: absence-attribution
description: Child of [[chronicle]] — spec 110's harvest ledger and per-chapter correction tally: why a remembered tree found gone stops earning a chronicle line of its own, how attributed corrections coalesce into one closeChapter summary while unexplained ones keep theirs byte-identical, the four dials, and the measured before/after. Load when tracing why a map correction did or did not reach the narrator.
kind: component
sources:
  - internal/mind/narrate.go
  - internal/mind/mind.go
verified_against: ebc30ee1990dbcdf17576a2fbe81fdbeb175411d
---

# Absence attribution

Split from [[chronicle]] (spec 110, TASK-173). The chronicle's narrator kept
naming a village-wide mystery out of ordinary firewood; this note is why, and
what stopped it.

## The problem was the narrator's input, not its judgment

Every `agent.map_corrected` with a non-empty `Gone` list used to contribute its
own line — *"X went looking for the pine at (12,7) and found it gone."*
Measured on a preserved 12.02-game-day world, those lines were the **majority of
every full day chapter** of `md.narrLines` (median ~57%, peak 68%). Five of
twelve day chapters also overran `narrMaxLines` (120), and because that ring
drops oldest-first, the overflow silently evicted builds, gifts, assemblies and
musings while corrections survived on sheer volume.

A model asked to *"group by storyline, not by hour"* over a list that is more
than half *found it gone* is not malfunctioning when it names a
vanishing-landscape storyline. It is summarising its input correctly — which it
did in **four of four** measured runs, across two local models and two worlds.

The corrections were mundane and provably so: **830 of 833** matched an
`agent.chopped`/`agent.quarried` at exactly those coordinates, and every match
fell within **3 game-days**. Exactly 3 were genuine anomalies.

Note what this does NOT say: the memory layer was fine. Absence-flavoured
memories were 6.7% of all memories (TASK-159/spec 081 had already fixed that),
and [[executor-perception-observation]]'s dedup and disconfirmation decay were working.
The 6.7% minority at the memory layer was a ~57% majority at the narrator's
prompt, and that gap is the whole bug.

## The mechanism

Two absorb-owned pieces on the [[agent-mind]] Mind, both lock-free exactly as
`narrLines` is, both fed and reset from the single absorb goroutine:

- **`harvestLedger`** (`Mind.harvests`) — a coordinate-keyed
  `(x, y) → {agent, tick}` store, written from the `agent.chopped`/
  `agent.quarried` absorb arm ([[mind-driver-triggers]]) alongside the spec-081
  witness re-arm that shares it. Eviction is oldest-first on **age**
  (`harvestLedgerWindow`) and on **count** (`harvestLedgerCap`); cap-eviction
  ties break by coordinate, so the surviving set is a function of the event log
  alone. `lookup` re-applies the age test, so an entry eviction has not yet
  reached still misses. A harvest recorded *after* a correction never attributes
  it.
- **`correctionTally`** (`Mind.corrTally`) — the per-chapter accumulator:
  attributed count, distinct locations, harvester id set, and the unexplained
  count.

`chronicleNote` routes each correction's **first** `Gone` fact — the same
first-fact convention the pre-110 line already used — through
`attributedHarvest`:

- **attributed** → folds into the tally, emits **no** line;
- **unexplained** → emits the pre-110 line **byte-identical**, so a genuine
  anomaly is untouched.

`closeChapter` then appends exactly **one** line for the whole chapter, stamped
at the chapter's close (so it sorts last) and opened by `corrSummaryMarker`:

    Ordinary harvesting: 40 remembered things the villagers went for had
    already been felled or quarried, at 12 locations, by Ash and Rowan.
    Routine village business.

Harvester ids are sorted before name resolution, so the wording is a function of
the log too. Singular/plural agree. A chapter that attributed nothing appends
nothing and is byte-identical to pre-110.

`narrateUserPrompt` adds one clause telling the narrator that such a line is
ordinary background whose cause is already named, and not to build an entry or a
thread around it. The clause is emitted **only when a line carrying the marker is
actually present** — it scans the job's lines rather than carrying a flag, so it
survives the retry-carry path where a failed chapter's summary line rides into
the next chapter's job, and a correction-free chapter's prompt stays
byte-identical. The existing *"group by storyline, not by hour"* instruction is
untouched.

## Telemetry

Per-chapter counts go to the daemon log, following [[nightly-consolidation]]'s
`nightReport` precedent — a counter flushed as one summary line:

    mind: chronicle "day 3, dawn to nightfall" corrections: 40 attributed (12 locations, 2 harvesters), 1 unexplained

Deliberately a log line and not an event: no sim event type, payload, or
`format_version` moves, so [[deterministic-rng]] and the
[[event-log]] format are untouched. A chapter with no corrections of either kind
logs nothing — an absent line reads as 0/0, not as missing data.

## Dials

| name | value | why |
|---|---|---|
| `harvestLedgerWindow` | 4 game-days | the measured harvest→correction lag topped out at 3 game-days (716 same-day, 76 at 1, 35 at 2, 3 at 3, 0 beyond) |
| `harvestLedgerCap` | 4096 entries | 352 distinct harvested locations over 12 game-days observed; the cap bounds a long run, it is not a tuning knob |
| `corrSummaryMarker` | `"Ordinary harvesting:"` | opens the coalesced line; the prompt clause and the evidence harness both key off it |
| `narrMaxLines` | 120 | pre-existing chapter ring ([[chronicle]]); the overflow this note's change relieves |

The failure direction is deliberate: a **missed** attribution merely restores
pre-110 behaviour for one line, while a **false** attribution would hide a real
mystery. Matching is exact-coordinate for that reason.

## Measured outcome

Replay of the preserved worlds' own event logs, then re-narration through the
production narrator path:

- Corrections fall from 48–65% of each day chapter to **1–2%**; attributed
  corrections contribute exactly one line per chapter across all 24 + 13
  replayed chapters; ring overflow **5 chapters → 0**, recovering 130
  non-correction lines that had been evicted.
- **No named absence storyline** in any re-narrated run, against one in all four
  before-runs carrying 19–43% of their entries.
- Attribution precision **100%** over 1,290 replayed real corrections, judged
  against ground truth built from the raw log rather than the ledger under test.
- Every genuinely-unexplained correction kept its own line (3/3 and 5/5).

Recorded limits: replay exercises the narrator's input faithfully but re-runs its
output against a live model, so this is evidence about the same game-days rather
than a fresh world's emergent dynamics, and it cannot exercise the
narrator→villager feedback loop. Full evidence, method, and caveats:
`specs/110-absence-attribution/evidence.md`.

## Known adjacent gap

The same amplification mechanism is live for **musings** on the `qwen3.6`
default: on that world the 120-line ring overflows 7 of 13 chapters both before
and after spec 110, and 599 of 921 retained lines are `agent.thought`
`source=musing`. Spec 110 addresses corrections only; the musing case is carded
separately (TASK-194).
