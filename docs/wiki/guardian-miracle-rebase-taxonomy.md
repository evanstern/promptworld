---
name: guardian-miracle-rebase-taxonomy
description: The SHIFT/KEEP taxonomy (rebaseTicks) every tick-anchored int64 state field must classify — which fields a time snap shifts forward to preserve in-flight durations vs. which stay put as history/identity, and the build-time completeness guard. Split from [[guardian-miracles]]; load when adding a new tick-anchored field or auditing a time-snap's effect on existing state.
kind: component
sources:
  - internal/sim/miracles.go
verified_against: 510a3c3133e120d84cd50525dbc4ee0d3ec01cdc
---

# Guardian's miracle rebase taxonomy

Split from [[guardian-miracles]] (summary-style, corpus-spec v2) — the
SHIFT/KEEP classification every tick-anchored `int64` field must carry.

**Shift-semantics re-base taxonomy** (`rebaseTicks` in `miracles.go`): the SINGLE
authority for how a time snap preserves in-flight durations while history stays
put (FR-009). Every tick-anchored `int64` field anywhere in the state tree MUST be
classified SHIFT or KEEP in its doc comment:

- **SHIFT** (`+delta`) — a future deadline, or an anchor from which an elapsed/
  remaining duration is measured (shifting preserves that duration across the
  jump). A SHIFT field whose zero value means "unset/never" is shifted only when
  non-zero. SHIFT fields: `Agent.IdleSince` (shifted unconditionally — its zero is
  genesis-idle, a real tick, not a "never" sentinel), `Agent.LastTalk`/`LastGive`,
  `Intent.WorkStart`, `AgentHail.Until`, `PlanStep.Until`, `Guard.Tick`,
  `Structure.FuelUntil`, `Harvest.Regrow`, `DenUse.Ready`, `FoodBatch.SpoilAt`,
  `Debt.Due`, `Belief.Reinforced` (spec 030: the decay-curve anchor, elapsed =
  tick − Reinforced; shifted only when non-zero — a legacy grandfathered belief
  stays at 0 so it never decays), `Gru.LastAttack`, `Meeting.OpenedTick`,
  `Meeting.GatherStart`, and (spec 029) `GuardianOrder.ExpiresTick` — shifted ONLY
  for ACTIVE orders, so a standing order's remaining lifetime survives the jump (a
  consumed order's deadline is a spent artifact, left put). Spec 041
  ([[mental-maps]]) adds `PlaceFact.Seen` and `PeerSighting.Seen`, the mental
  map's freshness anchors (fresh iff `now − Seen < horizon`, the
  `Belief.Reinforced` shape) — shifted unconditionally when non-zero, since a
  snap would otherwise instantly stale every villager's spatial knowledge;
  `applyEntityMoved`'s villager case (below) shares the same derived
  bookkeeping a walked step gets. Spec 043 adds `Agent.NeedsAnchorTick`, the
  need-trajectory window's edge anchor (elapsed = tick − NeedsAnchorTick gates
  the anchor roll in the `agent.needs_changed` arm; 0 = unset sentinel, stays
  0) — shifted so a snap preserves the window's remaining time instead of
  forcing an immediate anchor reset that would wipe every villager's
  trajectory sense; `NeedsAnchor` itself holds need levels, not ticks, so it
  needs no entry. Spec 061 ([[social-fabric]], [[sim-state-reducer]]) adds
  `PairTalk.Tick`, the conversation loop damper's per-pair last-exchange
  anchor (cooldown elapsed = tick − Tick, the `Agent.LastTalk` shape) —
  shifted UNCONDITIONALLY: unlike most SHIFT fields here, a PRESENT
  `PairTalk` record is always a real exchange tick (absence of the record
  itself, not a zero value, means "never talked"), so there is no zero
  sentinel to guard. Spec 062 ([[reflex-policy]], [[sim-state-reducer]]) adds
  `Agent.LastMindIntentDone`, the reflex PREP gate's yield-window anchor
  (elapsed = tick − LastMindIntentDone gates `prepYields`; 0 = never
  mind-driven, the permanent sentinel every no-planner world's agents carry)
  — shifted only when non-zero, the `Belief.Reinforced`/`NeedsAnchorTick`
  shape: a snap must preserve the window's remaining deference rather than
  spuriously arming or clearing it. Spec 077
  ([[event-types-scenario-incidents]]) adds `State.ColdSnapUntil` (the cold
  snap's read-time expiry — the `Structure.FuelUntil` shape, only-non-zero)
  and `Stranger.LastMove`/`Stranger.LastTake` (the entity's cadence anchors,
  the `Gru.LastAttack` shape) — all SHIFT, so a snap preserves a live
  snap's remaining window and the stranger's cooldowns.
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
  only-non-zero), and spec 063's `GuardianReportCard.Tick`/`Seq`/`Citations`
  (KEEP) as new tick-anchored
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

[[guardian-miracles]] is the parent — `applyTimeSnapped`
([[guardian-miracle-mechanics]]) is this taxonomy's sole caller
(`rebaseTicks`), spending 2 charges before applying it. [[guardian-orders]]
shares this taxonomy (`GuardianOrder.ExpiresTick` is SHIFT, `PlacedTick` is
KEEP). [[mental-maps]] shares it too (`PlaceFact.Seen`/`PeerSighting.Seen`
SHIFT, `PlaceFact.Detail` KEEP). [[social-fabric]]/[[sim-state-reducer]]
share it since spec 061 (`PairTalk.Tick` SHIFT). [[grounded-feedback]]
shares it since spec 063 (`GuardianReportCard.Tick`/`Seq`/`Citations`
KEEP).
