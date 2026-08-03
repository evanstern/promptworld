# Feature Specification: Absence attribution — harvest-explained map corrections stay mundane

**Feature Branch**: `task-173-absence-attribution`

**Created**: 2026-08-02

**Status**: Draft

**Input**: TASK-173, re-opened 2026-08-02. Harvest-explained map corrections still earn
mystery-grade narrative weight. Two independent soak worlds on current `main` — one on
`gemma4:12b-mlx` (12.02 game-days), one on `qwen3.6` (5.69 game-days) — each carry a
**named** absence storyline in the chronicle (`[the-missing-trees]`,
`[the-disappearance-of-res...]`), so the failure is not model-specific.

**Board task**: TASK-173 · **Runbook**: `docs/design/task-173-absence-attribution-runbook.md`

---

## Problem

A villager who walks to a remembered tree and finds a stump has learned something
ordinary: someone chopped it. The chronicle instead reads it as a phenomenon, and over a
week of play the phenomenon becomes the story of the village.

The card's re-open block already establishes what is **not** broken, and this spec does
not revisit it: TASK-159/spec 081 fixed the memory layer (absence-flavoured memories are
1,234 of 18,283 = **6.7%** of all memories, against the 75% showstopper baseline), and
spec 097 cut the correction rate (69.3/game-day vs playtest-1's ~101/day at a comparable
harvest rate). The remaining failure is downstream, at **narrative salience**.

### The amplification mechanism, measured

The narrator does not see the memory layer's 6.7%. It sees `md.narrLines` — the
per-chapter line buffer that `chronicleNote` fills, one line per notable event, capped at
`narrMaxLines = 120` (`internal/mind/narrate.go:30-33, 293-295`). Every
`agent.map_corrected` with a non-empty `Gone` list contributes one line
(`narrate.go:159-172`): *"X went looking for the pine at (12,7) and found it gone."*

Measured on the preserved 12.02-game-day soak world
(`/Users/evanstern/.claude/jobs/ca35de11/tmp/soak/soak-world/world.db`), counting the
events that actually produce chronicle lines, grouped by chapter boundary
(`sim.day_started` / `sim.night_started`):

| chapter | correction lines | total narratable lines | corrections' share |
|---------|------------------|------------------------|--------------------|
| 0  | 89 | 141 | 63% |
| 2  | 92 | 158 | 58% |
| 4  | 68 | 107 | 64% |
| 6  | 99 | 145 | 68% |
| 8  | 51 | 96  | 53% |
| 10 | 53 | 124 | 43% |
| 12 | 60 | 106 | 57% |
| 14 | 92 | 149 | 62% |
| 16 | 56 | 102 | 55% |
| 18 | 24 | 71  | 34% |
| 20 | 23 | 80  | 29% |
| 22 | 44 | 98  | 45% |

**Corrections are the majority of the narrator's input on every full day chapter** —
median ~57%, peak 68%. The 6.7% minority signal at the memory layer is a ~57% majority
signal at the narrator's prompt. A model told "group by storyline, not by hour"
(`narrate.go:712`) and handed a list that is more than half *"found it gone"* is not
malfunctioning when it names a vanishing-landscape storyline. It is summarising its input
correctly.

There is a second, compounding effect: **five of the twelve day chapters exceeded the
120-line ring** (141, 158, 145, 124, 149 lines). The ring drops oldest-first, so the
overflow silently evicted *other* content — builds, gifts, assemblies, musings — while
corrections, being the most numerous, survived by sheer volume. Absence does not merely
dominate the story; it displaces the rest of it.

### The corrections are mundane, and provably so

Same world, joining each correction's first `Gone` fact to the earliest
`agent.chopped` / `agent.quarried` at those exact coordinates:

| bucket | count |
|--------|-------|
| harvest same game-day | 716 |
| harvest 1 game-day earlier | 76 |
| harvest 2 game-days earlier | 35 |
| harvest 3 game-days earlier | 3 |
| **no harvest at those coordinates, ever** | **3** |

830 of 833 corrections (99.6%) are explained by a harvest, and **every** explained
correction's harvest falls within **3 game-days** of it. Exactly 3 corrections in 12.02
game-days are genuine anomalies — independently reproducing the card's 969/972 figure by
a different query. AC#2 is failing not at the margin but on essentially the entire
population.

## Design decisions (resolved here; the runbook's checkpoints carry the two that need the operator)

- **D1 — Attribution is computed at the NARRATOR, from world truth, not from each
  villager's knowledge.** The card's original phrasing ("known harvest activity —
  witnessed or rumored") was written when the target was the memory layer. The re-open
  moved the target: the villagers' own epistemics are working (spec 081/097; the card
  even records villagers reaching for the mundane explanation socially — *"Birch accuses
  Sage of cutting trees based on suspicious tracks"*), and the failure is in the
  chronicle, which is the **guardian's** story feed, voiced by a world-level narrator that
  legitimately knows who chopped what. Gating the chronicle on per-villager knowledge
  would rebuild spec 097's belief machinery in the narrator for no player-visible gain.
  Per-villager attribution is explicitly **out of scope** and stays with the memory layer.
- **D2 — Attributed corrections COALESCE; they are not suppressed.** This is the runbook's
  operator checkpoint 1, decided (a)-with-a-volume-bound: an attributed correction stops
  contributing its own line and instead folds into **one** per-chapter summary line that
  names the cause and the harvesters. Suppressing them outright would trade a false
  mystery for a silent world and would also delete the believe-act-discover beat spec 041
  deliberately built; keeping them one-per-event is the bug. One coalesced line both
  attributes (a cause is named) and bounds the volume (1 line, not ~57% of the buffer).
- **D3 — Mind-side only; no event-shape or determinism change.** The ledger and the
  classification live in the Mind, which already absorbs `agent.chopped`/`agent.quarried`
  (`mind.go:315-346`) and already renders chronicle lines in the same absorb goroutine
  with a current replica. No `agent.map_corrected` payload change, no new event type on
  the emission path, no `format_version` bump — so specs 092/094 are untouched and the
  runbook's operator checkpoint 2 does **not** fire. Narrator output continues to enter
  the world only as recorded input through `InjectSocial`.
- **D4 — Ledger window is 4 game-days.** The measured lag distribution tops out at 3
  game-days; 4 covers 100% of the observed explained population with a margin, and the
  ledger is small (352 distinct harvested locations over 12 game-days). A correction whose
  coordinates have no harvest inside the window is treated as unexplained — the honest
  failure direction, since a missed attribution merely restores today's behaviour for that
  one line, while a false attribution would silence a real mystery.
- **D5 — Unexplained corrections are untouched.** They keep the existing per-event line
  verbatim. AC#3 is served not by adding emphasis but by removing the noise they were
  drowning in: three anomalies among 830 look-alikes are invisible; three anomalies beside
  one "ordinary harvesting" summary line are conspicuous.

## User Scenarios & Testing *(mandatory)*

### User Story 1 — The village stops panicking about firewood (Priority: P1)

As a player, I want real mysteries to stand out, not be drowned by my villagers panicking
about firewood their neighbours chopped.

**Acceptance Scenarios**:

1. **Given** a chapter in which 40 map corrections all match harvests at their exact
   coordinates within the ledger window, **When** the chapter closes, **Then** the
   narrator's line buffer carries **one** attributed summary line for all 40 — naming the
   count, the distinct locations, and the harvesters — and **no** per-correction "found it
   gone" line.
2. **Given** the same chapter, **When** the narrator's input is inspected, **Then**
   corrections account for at most one line of it, and the non-correction content that
   previously overflowed the 120-line ring is retained.

### User Story 2 — A genuine anomaly still lands (Priority: P1)

As a villager in the game, when something really does vanish with no one having touched
it, I want that to be as strange as it deserves.

**Acceptance Scenarios**:

1. **Given** a correction whose coordinates carry no harvest within the ledger window,
   **When** the chapter closes, **Then** it contributes its existing per-event line
   unchanged, alongside (not merged into) the attributed summary line.
2. **Given** a chapter containing both 40 attributed corrections and 1 unexplained one,
   **When** the chapter is narrated, **Then** the unexplained absence is the only absence
   the narrator is invited to treat as notable.

### User Story 3 — The narrator is told which is which (Priority: P2)

As the guardian reading the chronicle, I want the story to name the mundane cause rather
than leave a gap I will read as a mystery.

**Acceptance Scenarios**:

1. **Given** the coalesced summary line, **When** the narrator prompt is assembled,
   **Then** the prompt marks that line as ordinary background and not storyline material.

## Requirements *(mandatory)*

- **FR-001** The Mind MUST maintain a bounded harvest ledger, keyed by `(x, y)`, populated
  from absorbed `agent.chopped` and `agent.quarried` events, carrying the harvester's
  agent id and the harvest tick. Entries older than the D4 window (4 game-days) MUST be
  evicted, and the ledger MUST carry a hard entry cap so a long run cannot grow it without
  bound.
- **FR-002** On absorbing `agent.map_corrected`, the Mind MUST classify the event as
  **attributed** when the first `Gone` fact's coordinates hit the ledger, and
  **unexplained** otherwise, using the same first-fact convention the existing line
  already uses (`narrate.go:159-172`).
- **FR-003** An attributed correction MUST NOT contribute its own line to `md.narrLines`.
  Instead the chapter MUST accumulate attributed corrections and, at `closeChapter`,
  contribute exactly **one** line naming: the number of corrections, the number of
  distinct locations, and the harvesters responsible (resolved to names through the same
  roster path the existing lines use).
- **FR-004** An unexplained correction MUST contribute its existing per-event line,
  unchanged in wording and position.
- **FR-005** The narrator system/user prompt MUST identify the coalesced line as ordinary
  background rather than storyline material.
- **FR-006** The change MUST be confined to `internal/mind`. No sim event payload, event
  type, or `format_version` may change, and replay determinism MUST be unaffected.
- **FR-007** The Mind MUST record per-chapter counts of attributed vs unexplained
  corrections on its existing telemetry path, so a soak can measure the outcome without
  re-deriving it from the event log.
- **FR-008** A chapter with zero corrections MUST be byte-identical to today's behaviour —
  no empty summary line, no spent call.

### Key entities

- **Harvest ledger entry** — `(x, y) → {agentID, tick}`, bounded by age (4 game-days) and
  by count.
- **Chapter correction tally** — per-chapter accumulator: attributed count, distinct
  attributed locations, set of harvester ids, reset at `closeChapter` alongside
  `md.narrLines`.

## Success criteria *(mandatory)*

Mapped to the board card's acceptance criteria and to the runbook's evidence bar
(≥12 game-days; a shorter window cannot disprove the 12.02-game-day soak that re-opened
this card).

- **SC-001 (card AC#2)** In a soak of ≥12 game-days on the same scenario, **no named
  absence storyline** appears in the chronicle ring, and absence-themed chronicle entries
  fall from the soak baseline of 18/90 (20%) to **at most 5%**.
- **SC-002 (card AC#2)** Corrections' share of the narrator's per-chapter line buffer
  falls from the measured median ~57% to **at most one line per chapter**, and no chapter
  overflows `narrMaxLines` on account of corrections.
- **SC-003 (card AC#3)** Every correction with no harvest at its coordinates within the
  window still produces its own chronicle line — verified against the soak's known
  population (3 in 12.02 game-days) and by unit test.
- **SC-004** Attribution precision on replayed real data is **100%** on the soak's
  population: no correction that lacks a coordinate-matching harvest is classified
  attributed.
- **SC-005** `go build ./...`, `go test ./...`, and `go test -race ./internal/mind/...`
  green.
- **SC-006** Where feasible the soak runs on both `gemma4:12b-mlx` and `qwen3.6`, since
  the card establishes the failure reproduces on both; a single-model soak records its
  reason on the card.

## Out of scope

- Per-villager attribution (who knew what, rumours as evidence) — the memory layer's
  business, working per specs 081/097.
- Changing the correction rate itself, the memory salience table, or spec 097's
  disconfirmation decay.
- The pre-existing `tui-design` staleness on `main` (`docs/design/tui/anatomy.md`) — not
  caused by this task and not adopted by it.
