---
name: guardian-miracle-rebase-shift-fields
description: Child of [[guardian-miracle-rebase-taxonomy]] — the full SHIFT half of the rebaseTicks taxonomy: every tick-anchored int64 field a time snap must shift forward to preserve an in-flight duration, spec by spec (029 guardian orders through 084 designations/085 prophecy/041 mental maps/043 needs anchor/061 pair-talk/062 reflex yield/077 incidents/083 neglect).
kind: component
sources:
  - internal/sim/miracles.go
verified_against: 72f82f41f7aa2e345572105894cd0fb7c02fc0aa
---

# Guardian's miracle rebase taxonomy — the SHIFT fields

Child of [[guardian-miracle-rebase-taxonomy]]: the full SHIFT half of
`rebaseTicks`'s classification — every tick-anchored `int64` field a time
snap must shift forward (`+delta`) to preserve an in-flight duration. See
the parent for the KEEP half and the build-time completeness guard.

## SHIFT fields, spec by spec

A future deadline, or an anchor from which an elapsed/remaining duration is
measured (shifting preserves that duration across the jump). A SHIFT field
whose zero value means "unset/never" is shifted only when non-zero. SHIFT
fields: `Agent.IdleSince` (shifted unconditionally — its zero is
genesis-idle, a real tick, not a "never" sentinel), `Agent.LastTalk`/`LastGive`,
`Intent.WorkStart`, `AgentHail.Until`, `PlanStep.Until`, `Guard.Tick`,
`Structure.FuelUntil`, `Harvest.Regrow`, `DenUse.Ready`, `FoodBatch.SpoilAt`,
`Debt.Due`, `Belief.Reinforced` (spec 030: the decay-curve anchor, elapsed =
tick − Reinforced; shifted only when non-zero — a legacy grandfathered belief
stays at 0 so it never decays), `Gru.LastAttack`, `Meeting.OpenedTick`,
`Meeting.GatherStart`, and (spec 029) `GuardianOrder.ExpiresTick` — shifted ONLY
for ACTIVE orders, so a standing order's remaining lifetime survives the jump (a
consumed order's deadline is a spent artifact, left put); spec 084
([[guardian-designations]]) adds `Directive.ExpiresTick`, the same
classification verbatim (ACTIVE only), while both plan entities' other
tick fields — `Designation.PlacedTick`/`PlacedSeq`,
`Directive.IssuedTick`/`PlacedSeq` — are history/identity KEEP (a
designation carries no future deadline at all, and the `designation`
place facts' `Seen` anchors ride the existing per-agent `PlaceFact.Seen`
SHIFT loop with no new code); spec 085 ([[guardian-faith]]) adds
`Prophecy.DeadlineTick`, the `Directive.ExpiresTick` classification
verbatim again (ACTIVE only — a settled prophecy's deadline is a spent
artifact), with `Prophecy.DeclaredTick`/`PlacedSeq` history/identity
KEEP and `FaithState` untouched entirely (it carries no tick fields). Spec 041
([[mental-maps]]) adds `PlaceFact.Seen` and `PeerSighting.Seen`, the mental
map's freshness anchors (fresh iff `now − Seen < horizon`, the
`Belief.Reinforced` shape) — shifted unconditionally when non-zero, since a
snap would otherwise instantly stale every villager's spatial knowledge;
`applyEntityMoved`'s villager case shares the same derived bookkeeping a
walked step gets. Spec 043 adds `Agent.NeedsAnchorTick`, the
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
snap's remaining window and the stranger's cooldowns. Spec 083
([[executor-needs-survival]], [[sim-state-agent-fields]]) adds the six
`NeglectState` tick anchors — `FoodSince`/`WarmthSince`/`RestSince` (the
band-entry anchors: elapsed = tick − Since gates the neglect window in
`NeglectDue`; 0 = not in band) and `FoodIntent`/`WarmthIntent`/`RestIntent`
(the last-class-intent stamps: elapsed = tick − Intent is the zero-intent
clause; 0 = never) — all SHIFT only-non-zero, the
`Belief.Reinforced`/`NeedsAnchorTick` elapsed-anchor shape: left unshifted,
a snap would age every in-band episode past the window and fire the
detector the instant it landed. The `*Fired` latches are bools (episode
state, no tick field, no taxonomy entry). Spec 097
([[executor-perception-observation]]) adds `ObservationMark.Tick` — the
grounded-observation dedup anchor on `Agent.LastObs` (elapsed = tick − Tick
gates the dedup window; never zero once the pointer exists) — SHIFT, the
`Belief.Reinforced` shape.

## Connections

Parent [[guardian-miracle-rebase-taxonomy]] holds the KEEP half and the
`TestRebaseTaxonomyComplete` build guard. [[guardian-designations]] and
[[guardian-faith]] own the spec 084/085 SHIFT additions;
[[mental-maps]]/[[social-fabric]]/[[reflex-policy]]/
[[event-types-scenario-incidents]]/[[executor-needs-survival]] each own a
spec's worth of the fields listed above.
