---
name: guardian-miracle-rebase-taxonomy
description: The SHIFT/KEEP taxonomy (rebaseTicks) every tick-anchored int64 state field must classify — the KEEP half (history/identity fields) and the build-time completeness guard here; the full SHIFT half splits to [[guardian-miracle-rebase-shift-fields]]. Split from [[guardian-miracles]]; load when adding a new tick-anchored field or auditing a time-snap's effect on existing state.
kind: component
sources:
  - internal/sim/miracles.go
verified_against: a761a45cb3b437613b808408c6c7f30d11bd9eb9
---

# Guardian's miracle rebase taxonomy

Split from [[guardian-miracles]] (summary-style, corpus-spec v2) — the
SHIFT/KEEP classification every tick-anchored `int64` field must carry.

**Shift-semantics re-base taxonomy** (`rebaseTicks` in `miracles.go`): the SINGLE
authority for how a time snap preserves in-flight durations while history stays
put (FR-009). Every tick-anchored `int64` field anywhere in the state tree MUST be
classified SHIFT or KEEP in its doc comment:

- **SHIFT** (`+delta`) — a future deadline, or an anchor from which an
  elapsed/remaining duration is measured (shifting preserves that duration
  across the jump). A SHIFT field whose zero value means "unset/never" is
  shifted only when non-zero. The full field-by-field, spec-by-spec list
  (029 guardian orders through 084/085 plans/faith, 041 mental maps, 043
  needs anchor, 061 pair-talk, 062 reflex yield, 077 incidents, 083
  neglect) splits into [[guardian-miracle-rebase-shift-fields]].
- **KEEP** — a historical timestamp or an identity/counter; rewriting it would
  rewrite history or break a reference. `Agent.Generation`, `Agent.LastGoalTick`,
  `Memory.Tick`, `Memory.Conv` (spec 019: a conversation-ref identity, the same
  founding-talk tick as `ConvoRecord.Conv` — an identity, not a duration anchor),
  `Memory.Seq` (spec 042: the emitting event's store seq — an identity, never a
  clock value), `Agent.SitVecTick` (spec 042: when the agent's situation text
  was rendered — history/audit, the `Memory.Tick` shape), `JournalEntry.Tick`
  (spec 019: when the entry was written, a historical
  timestamp), `Belief.Tick`, `ChronicleEntry.Tick`/`Day`/`FromTick`/`ToTick`,
  `GuardianOrder.PlacedTick` (spec 029: when the order was placed, history),
  `IntentRecord.Tick`/`IntentRecord.OutcomeTick` (spec 043: when an intent
  landed / when its outcome landed — the recent-intent ring is a historical
  self-history log, the `Memory.Tick` shape, never a future deadline),
  `PlaceFact.Detail` (spec 041: a remembered value baked at emission, never
  re-derived — for a fire it mirrors the FuelUntil last seen, so shifting it
  would rewrite what the agent remembers rather than what is; the perception
  sweep simply re-witnesses the shifted reality on the next look),
  `GuardianReportCard.Tick`/`Seq`/`Citations` (spec 063,
  [[grounded-feedback]]: when the card landed, the card event's own identity,
  and the cited event seqs — history and identities, never deadlines),
  spec 077's `Stranger.Night` (the 1-based arrival night — identity),
  `StrangerTake.Tick` (when the take happened — ledger history, the
  `DeathRecord.Tick` shape), and `State.CharterObservedSeq/Tick` +
  `State.SkillsObservedSeq/Tick` (log coordinates re-locating recorded
  observation events — identities, the `Memory.Seq`/`EvidenceRef.Tick`
  shape: rewriting them would point pass evidence at ticks where nothing
  was recorded), and every
  other identity/history field — see the doc comment for the full list.
  `TestRebaseTaxonomyComplete` caught both spec-019 additions, the spec-030
  `Belief.Reinforced` field, (later) spec 029's `GuardianOrder.ExpiresTick`/
  `PlacedTick`, spec 041's `PlaceFact`/`PeerSighting` fields, spec 042's
  `Memory.Seq`/`Agent.SitVecTick` fields, spec 043's `IntentRecord.Tick`/
  `OutcomeTick` (KEEP) and `Agent.NeedsAnchorTick` (SHIFT), spec 061's
  `PairTalk.Tick` (SHIFT), spec 062's `Agent.LastMindIntentDone` (SHIFT,
  only-non-zero), spec 063's `GuardianReportCard.Tick`/`Seq`/`Citations`
  (KEEP), and spec 083's six `NeglectState` anchors (SHIFT, only-non-zero)
  as new tick-anchored
  `int64` fields requiring classification, confirming the taxonomy guard holds
  across features outside miracles' own spec.

`TestRebaseTaxonomyComplete` (`internal/sim/miracles_test.go`) is the taxonomy guard:
it fails the build when a new tick-anchored `int64` field appears in the state
structs without a classification entry here, so the taxonomy can never silently
drift from the struct definitions. `PlanStep.Until` and `Guard.Tick` are shifted even
though `specs/016-metatron-miracles/data-model.md` did not list them — a deviation
recorded in `rebaseTicks`'s doc comment: both are genuine future deadlines FR-009's
catch-all ("any future duration-anchored state") requires shifting, since leaving
them unshifted would expire a pending plan step or fire a timed guard the instant a
snap jumped past its absolute tick.

## Connections

[[guardian-miracle-rebase-shift-fields]] is this note's own split-off
child — the full SHIFT half. [[guardian-miracles]] is the parent — `applyTimeSnapped`
([[guardian-miracle-mechanics]]) is this taxonomy's sole caller
(`rebaseTicks`), spending 2 charges before applying it. [[guardian-orders]]
shares this taxonomy (`GuardianOrder.ExpiresTick` is SHIFT, `PlacedTick` is
KEEP). [[mental-maps]] shares it too (`PlaceFact.Seen`/`PeerSighting.Seen`
SHIFT, `PlaceFact.Detail` KEEP). [[social-fabric]]/[[sim-state-reducer]]
share it since spec 061 (`PairTalk.Tick` SHIFT). [[grounded-feedback]]
shares it since spec 063 (`GuardianReportCard.Tick`/`Seq`/`Citations`
KEEP).
