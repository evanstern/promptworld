---
name: sim-state-world-fields
description: Per-world field catalog on sim.State — structures/piles, the social fabric, the gru/stranger incident state, PairTalks, the chronicle ring, Guardian charges/orders, and governance. The run-outcome and progression fields (Deaths/RunEnd, charter/skills observation, morgue epilogues, curriculum, tuning, report card) are split to [[sim-state-outcome-fields]].
kind: component
sources:
  - internal/sim/state.go
  - internal/sim/agents.go
  - internal/sim/stranger.go
verified_against: 376afd4cee54839a545bc88409f3c485c2f5149d
---

# Sim state: world & social fields

Split from [[sim-state-reducer]] (summary-style, corpus-spec v2): everything
`sim.State` carries about the shared world rather than any one agent —
structures/piles, the social fabric, the gru/stranger incident state, the
chronicle ring, the Guardian's charge bank and standing orders, and village
governance. The run-outcome and progression fields are split to
[[sim-state-outcome-fields]].

The catalog: **structures** (`fire`/`shelter`/`oven`/`chest`, fires carrying a
`FuelUntil`; chests (spec 013 US3) carrying a permanent `Owner` — the builder's
agent index, zero-value round-tripping unambiguously since every chest has one —
and a `Store *Inventory` capped at `chestCap` via the same derived `bulk()` used
for agents; spec 032 US1 adds two wall kinds, `wall_plank`/`wall_stone`,
carrying `HP` — current health, 1..`wallMaxHP(kind)`, derived from kind and
never stored as a separate max, the fire lit-ness doctrine — plus a non-wall
`path` kind carrying no `HP` and never blocking passage; spec 044 US4 adds a
`grave` kind, reducer-placed only — never player-built — marking a death
tile), cleared trees,
harvested forage, den cooldowns, `Quarried` depleted
rock outcrops (spec 012, permanent, `omitempty`), `Piles []Pile` — the per-tile
ground commons of dropped/spilled goods (spec 013 US2): event-sourced overlay
state like `Quarried`, one pile per tile (a reducer invariant), non-food a flat
count, food batch-tracked (`FoodBatch{Kind, N, SpoilAt}`, same `(Kind,SpoilAt)`
merges), spears AND axes (spec 032 US2, a `Spears` clone) sorted ascending,
`omitempty` — the social
fabric — relation edges, the debt ledger, the rumor registry with per-holder
variants and the bounded conversation-record ring ([[social-fabric]]) — the
consolidated inner life: per-agent beliefs (each now carrying a `Reinforced`
decay-clock anchor since spec 030, `omitempty` — 0 is a legacy grandfather that
never decays; [[nightly-consolidation]] covers the decay curve), self-narrative,
and the
once-per-night consolidation ledger ([[nightly-consolidation]]) — the
[[gru]] (`Gru *Gru`, nil while not abroad; `omitempty` keeps pre-TASK-10
snapshots valid) — joined since spec 077 by the incident state
([[event-types-scenario-incidents]]): `ColdSnapUntil int64` (the cold
snap's read-time expiry latch — `coldSnapActive` is `tick < ColdSnapUntil`,
no end event exists), `Stranger *Stranger` (`{X, Y, Night, LastMove,
LastTake}`, nil while not abroad — the gru's entity precedent), and
`StrangerTakes []StrangerTake` (`{Tick, X, Y, Kind, N}`, a ring bounded at
32 on the standing-order prune precedent — the durable take ledger
zero-wanted rubric terms count), all `omitempty` so pre-077 snapshots
round-trip byte-identically — and, since spec 061 ([[social-fabric]]), the conversation
loop damper's world truth: `PairTalks []PairTalk` (`omitempty`, the
Journal/Hail/Map pointer-precedent's slice sibling) — an unordered per-pair
last-exchange tick, `{A,B,Tick}` with a stored `A<B` invariant and the slice
kept sorted by `(A,B)` so canonical bytes never depend on talk arrival order
(a map would marshal non-deterministically); updated by the `agent.talked`
arm alongside `LastTalk`, absent record ≡ never talked, so every pre-061
snapshot round-trips byte-identically — and the narrated story: the bounded
`State.Chronicle`
ring ([[chronicle]], TASK-11), which rides snapshots so attaching clients
get catch-up history for free — the Guardian's charge bank
(`GuardianCharges`, JSON tag `metatron_charges` (frozen, spec 052 ruling 2), genesis 1,
deliberately not `omitempty` so a
spent-to-zero bank round-trips as 0; [[guardian]], TASK-12) — the standing-order
substrate (`GuardianOrders []GuardianOrder`, spec 029, `omitempty` — here an
empty order set genuinely IS the zero value, unlike the charge bank, so a
pre-029 snapshot with the field absent unmarshals to nil; [[guardian-orders]])
— the guardian plan layer (`Designations []Designation` and
`Directives []Directive`, spec 084, both `omitempty` on the same precedent —
pre-084 snapshots load byte-identical, no format bump;
[[guardian-designations]])
— the faith economy (`Faith *FaithState`, spec 085, `omitempty`, nil =
genesis 50 read only through the nil-safe `FaithScore()` accessor — the
`Tuning` nil-means-default shape; and `Prophecies []Prophecy`, the
guardian's declared machine-checkable claims, `omitempty` on the plan-layer
precedent; [[guardian-faith]])
— and the village's
law ([[governance]], TASK-13): `MeetingPlace` (set once), the `Meeting`
lifecycle (including the TASK-36 emergent-gathering watch fields
`GatherStart/GatherX/GatherY`), the `MeetingConvention` (TASK-36, nil until a
source establishes it — pre-TASK-36 snapshots load nil, a village with no
standing agreement to meet), and the `Norms` list with monotonic
`NextNormID`/`NextProposalID`, all zero-valued in pre-TASK-13 snapshots (a
lawless village) — and the run-outcome & progression fields (the
`Deaths`/`RunEnd` ledger and `Ended` latch, the charter/skills observation
fingerprints and coordinates, `MorgueEpilogues`, `CurriculumPasses`/
`StagesUnlocked`, `Tuning`, `GuardianReportCard`), all `omitempty` and all
split to [[sim-state-outcome-fields]]
(executor types in `agents.go`; memories belong to
[[agent-mind]]). Its
`Apply(event)` method is the **only** event-driven mutation path — the live loop and
crash recovery run the exact same code, which is what makes replay provably equal to
live execution. Spec 012 bumped the save format to v2, and spec 013 (inventory &
storage — bulk cap, piles, chests, theft, rot) bumped it again to **v3**
([[world-save-directory]]); a v1 world's `Inventory` (just `wood`/`food`) cannot
decode under this build at all — [[world-migration]] is the bridge, chaining 1→2→3
in one run and landing as a single wholesale-replace event rather than incremental
`Apply` calls (see [[sim-state-apply-world]]).

## Connections

Back to [[sim-state-reducer]] for the whole `State`/`Apply` picture and the
other split-off notes. [[sim-state-outcome-fields]] carries this note's
run-outcome & progression half. [[social-fabric]] owns the
relation/debt/rumor types; [[guardian]] and [[guardian-orders]] own the
charge bank and standing-order lifecycle; [[governance]] owns
`Meeting`/`Norms`; [[event-types-scenario-incidents]] owns the
gru/stranger incident vocabulary. The Apply arms that mutate most of these
fields live in [[sim-state-apply-world]].
