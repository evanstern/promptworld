---
name: executor-needs-recovery-and-neglect
description: Child of [[executor-needs-survival]] — spec 064's needs-conditioned recovery holds (Intent.UntilNeed/UntilValue, recoveryHoldEvents, warm_up, the cold-emergency wake arm) and spec 083's neglect detector (NeglectDue, sim.neglect_detected, the Agent.Neglect anchors/latch). Load for warm_up/recovery-hold mechanics or the neglect percept's exact firing rule.
kind: component
sources:
  - internal/sim/executor.go
  - internal/sim/agents.go
verified_against: 864d2a3bcff4b3113739d596befc72229a84d4b8
---

# Executor — needs-conditioned recovery and the neglect detector

Child of [[executor-needs-survival]]: spec 064's needs-conditioned recovery
holds and spec 083's neglect detector — the two most recent additions to
the needs/survival loop, both keyed off the same danger-band doctrine the
parent's heartbeat maintains.

## How it works

**Needs-conditioned recovery** (spec 064, from TASK-101 spike Direction B):
`Intent` gains an OPTIONAL completion condition — `UntilNeed` (closed set
`warmth`/`rest`/`food`) and `UntilValue`, a need level — plus `HoldRef`,
the need level at the hold's anchor tick; all three
`omitempty`, so a conditionless intent (every pre-064 intent) marshals
byte-identically. With `UntilNeed` set, `executeAtTarget` intercepts BEFORE
the per-goal switch, handing off to `recoveryHoldEvents` instead of the goal's
default arrive-and-done: the intent HOLDS at its target (visibly recovering,
not idle), checked every tick against the live need — already-satisfied
completes at once, a threshold crossing completes normally (arming the spec-062
yield window iff the ring source `isMindSource`, so a reflex-issued recovery
never arms it), a higher-priority survival need (reflex-ladder order:
food > warmth > rest) crossing into ITS danger band ends the hold so the
agent re-decides (no new preemption immunity — a hold is LESS sticky than an
ordinary intent, never more), and no net gain over a `recoveryStallTicks`
(300, ~5 needs heartbeats) window aborts with the distinct `agent.recovery_stalled`
outcome (dead fire, displaced source, unreachable threshold) rather than
loitering forever. `warm_up` is the evidenced consumer — a planner tool
resolving exactly like `goto_warmth` but carrying the condition, its
optional `until_warmth` clamped (spec 058 clamp-with-notice posture) into
`[warmthRecoverFloor, needMax]` (`warmthRecoverFloor` = `dangerWarmthBelow`,
350; `needMax` 1000) via the single `clampWarmUp`/`ClampWarmUp` clamp home,
defaulting to doctrine constant `warmthRecoverTo` (800, a healthy margin
above the danger band) when absent — and the [[reflex-policy]] day AND night
warmth rungs (`reachKnownWarmth`) now issue the same conditioned `goto_warmth`
at the doctrine default, so a reflex-driven recovery also holds at the fire
instead of arriving, idling, and wandering off cold (the world-01
arrive-idle-vacuum, Direction B). `wakeReason` (US4, the audit's
Gap C) gains a matching cold-emergency wake arm — a sleeper whose warmth falls
below `exposureWakeBelow` (150 — exactly the hunger-emergency wake's shape
and magnitude, deliberately deviating from the plan's nominated 350: an
emergency floor, not the routine-dip danger band, so a sleeper isn't roused
merely for being cold) wakes only when night AND the reflex's warmth
ladder finds something actionable (the hunger wake's "food in hand" analog,
the churn bound); a cozy fire-side sleeper sleeps through untouched.
`wakeReason` now takes the state/map/tick that ladder check needs, not the
bare `(agent, night)` of before. Held-pinned villagers are excluded from the
emergent-gathering quorum ([[governance]]) — a survival hold is not an
elective assembly.

**Neglect detector** (spec 083, from the TASK-106 research §1.3 — the shape
that killed Oak on world-01 day 7): the same `%60` heartbeat opens with a
sweep, BEFORE the decay loop, over every living AWAKE agent × need in the
fixed food→warmth→rest order, evaluating the exported pure predicate
`NeglectDue(a, need, tick)` (`agents.go`) against PRE-tick state: need below
its spec-062 danger band (`dangerFoodBelow`/`dangerWarmthBelow` 350,
`dangerRestBelow` 250 — reused, one home) with the band-entry anchor a full
`neglectWindowTicks` (7200, two game-hours; promoted-dial-READY doctrine
const, not tuning.json) old, zero class-goal intents (`needClassGoals`,
[[reflex-policy]]) over the same sliding window, and the episode's fired
latch clear. On true it emits ONE `sim.neglect_detected`
(`NeglectDetectedPayload{agent, need, level, since}` — executor emission
class, no injection-door entry) followed immediately by a companion
salience-9 `agent.memory_added` (`salNeglect` = `GenerationBumpSalience`, a
deliberate join of the near-death/exile interrupt band — the generation bump
IS the research's planner beat; fixed per-need voice-of-evidence text,
`neglectMemoryText`, `OriginWitness`, `Why` empty). The anchors/latch live
on `Agent.Neglect` ([[sim-state-agent-fields]]), written only by reducer
arms; emitting the pair before the beat's `agent.needs_changed` means a
same-beat recovery folds latch-then-reset and closes the episode cleanly.
Sleepers are skipped at the beat (their inaction is sleep — the spec-064
wake ladder owns sleeping emergencies) while anchors keep accruing, so a
still-critical waker fires on its next heartbeat; the one-per-episode latch
re-arms only when the need recovers to/above its band. The chronicle renders
the event whole-line alert; the map's needs-critical overlay already covers
the state by construction ([[tui-chronicle-feed]], [[village-lens]]).

## Connections

Parent [[executor-needs-survival]] covers the heartbeat, fire fuel, eating,
and run-end detector this recovery/neglect machinery shares a beat with.
[[reflex-policy]] issues the conditioned `goto_warmth`/`warm_up`, owns the
`needClassGoals` dictionary the neglect detector's zero-intent clause
reads, and owns the wake ladder `wakeReason` consults; [[governance]]
excludes held-pinned villagers from its quorum; [[sim-state-agent-fields]]
owns `Agent.Neglect`'s anchor/latch fields; [[tui-chronicle-feed]] and
[[village-lens]] are the neglect percept's rendering surfaces.
