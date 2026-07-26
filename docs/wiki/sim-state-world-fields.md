---
name: sim-state-world-fields
description: Per-world field catalog on sim.State — structures/piles, the social fabric, Guardian charges/orders, governance, morgue/run-outcome, curriculum ladder, world-tuning, and the Guardian report card
kind: component
sources:
  - internal/sim/state.go
  - internal/sim/agents.go
  - internal/sim/morgue.go
  - internal/sim/curriculum.go
verified_against: 8495b34ffb9ee5dc02e224025f0a23313bbab900
---

# Sim state: world & social fields

Split from [[sim-state-reducer]] (summary-style, corpus-spec v2): everything
`sim.State` carries about the shared world rather than any one agent —
structures/piles, the social fabric, the Guardian's charge bank and standing
orders, village governance, the run's morgue/outcome ledger, the curriculum
ladder's world-visible facts, the world-tuning dial set, and the Guardian's
latest report card.

agent (every no-planner world stays there forever), structures (`fire`/`shelter`/`oven`/`chest`, fires carrying a
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
— and the village's
law ([[governance]], TASK-13): `MeetingPlace` (set once), the `Meeting`
lifecycle (including the TASK-36 emergent-gathering watch fields
`GatherStart/GatherX/GatherY`), the `MeetingConvention` (TASK-36, nil until a
source establishes it — pre-TASK-36 snapshots load nil, a village with no
standing agreement to meet), and the `Norms` list with monotonic
`NextNormID`/`NextProposalID`, all zero-valued in pre-TASK-13 snapshots (a
lawless village) — and, since spec 044 ([[morgue]]), the run's outcome: the
`Deaths []DeathRecord` ledger (`{agent, tick, cause}`, appended by the
`agent.died` arm in application = event order, bounded by the agent count —
it exists so the run-end emission stays a pure function of (state, batch)
rather than a log scan), the `Ended bool` terminal latch and `RunEnd *RunEnd`
summary (`{tick, deaths, final_cause}`, set once by the `run.ended` arm and
never cleared by any event, so snapshot+replay restores the ended posture on
restart for free), `CharterFingerprint` (the most recent effective-charter
content hash a Guardian turn ran under — the full revision timeline lives in
the event log) with, since spec 072, its authorship twin `CharterCustom bool`
(`charter_custom` — whether that most recent observation was player-authored,
`!CharterObservedPayload.Default`, set only by the same
`metatron.charter_observed` arm; the conservative false zero value means a
pre-072 snapshot with a custom charter in force reads "not known
player-authored" until the next revision is observed — the-law's rubric
charter conjunct reads it, [[scenario-machinery]]) and, since spec 077,
the observation COORDINATES `CharterObservedSeq/CharterObservedTick`
(stamped by the same arm from the event envelope — what
`CharterEvidenceFromState` re-locates pass evidence with; zero = a pre-077
snapshot, the evidence honestly absent until the next observation stamps
them) plus the skills-observation triple
`SkillsFingerprint`/`SkillsObservedSeq`/`SkillsObservedTick` (set only by
the `metatron.skills_observed` arm — the stage-3 evidence substrate,
[[curriculum-ladder-progression]]), and the
`MorgueEpilogues []MorgueEpilogue` bounded ring
(`morgueEpilogueCap` 32, the chronicle pattern) of narrator mourning prose —
all `omitempty`, so every pre-044 snapshot stays byte-identical with no
format bump — and, since spec 046 ([[curriculum-ladder]]), the ladder's
world-visible facts: `CurriculumPasses []CurriculumPass` (a bounded ring,
`curriculumPassRetain` 32 on the standing-order prune precedent, of recorded
exercise passes each carrying `EvidenceRef{type, seq, tick, custom}` audit
pointers back into this world's log) and `StagesUnlocked []string` (the
once-per-(world,stage) unlock latch — no cap needed, at most one entry per
ladder stage), both `omitempty` so a pre-046 snapshot with neither field
round-trips byte-identically; the per-user unlocks record is a PROJECTION of
these events, this state being the replayable authority — and, since spec 048
([[world-tuning]]), the effective world-tuning dial set: `Tuning *TuningState`
(`omitempty`, the Journal/Hail/Map pointer precedent — nil means the five
promoted doctrine defaults, set only by the `sim.tuning_applied` arm (see [[sim-state-apply-world]]), no
`format_version` bump) — and, since spec 063 ([[grounded-feedback]]), the
guardian's latest attribution note: `GuardianReportCard *GuardianReportCard`
(`omitempty`; `{Tick, Seq, Fingerprint, Note, Citations}` — the reducer keeps
only the most recent card, the `Tuning`/`Journal` pointer-precedent's
single-value sibling, since re-opening the console card seam re-reads the
stored note rather than re-grading; nil until the first card lands)
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
other five split-off notes. [[social-fabric]] owns the relation/debt/rumor
types; [[guardian]] and [[guardian-orders]] own the charge bank and
standing-order lifecycle; [[governance]] owns `Meeting`/`Norms`; [[morgue]]
owns `Deaths`/`RunEnd`/`MorgueEpilogues`; [[curriculum-ladder]] owns
`CurriculumPasses`/`StagesUnlocked`; [[world-tuning]] owns `Tuning`;
[[grounded-feedback]] owns `GuardianReportCard`. The Apply arms that mutate
most of these fields live in [[sim-state-apply-world]].
