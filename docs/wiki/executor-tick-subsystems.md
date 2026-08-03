---
name: executor-tick-subsystems
description: The ancillary subsystems stepEvents drives each tick beyond agent bodies and goals — Guardian charge regen and order expiry, scenario incidents/rubric, the gru's turn, the stranger's turn (spec 077), the social adjacency beat, and the governance/meeting/curfew layer. Load to see what else runs inside one tick besides needs and intents.
kind: component
sources:
  - internal/sim/executor.go
verified_against: 012f715f55d8d87317e601ad75686c599d277349
---

# Executor — tick subsystems

Child of [[executor]] — the non-agent-body subsystems `stepEvents` drives
every tick: Guardian charge/order upkeep, scenario incidents and rubric
evaluation, the gru's turn, the social beat, and the governance layer.

## How it works

Each tick, `stepEvents` regenerates Guardian's nudge charges (`guardian.charge_regenerated` at absolute boundaries of the
faith-band cadence while below the cap — spec 085 replaced the fixed
6-game-hour constant with `FaithRegenCadenceTicks(FaithScore, scenario)`:
fervent 4h / steady 6h (the genesis band — a world with no faith events
keeps the pre-085 schedule byte-identically) / wavering 12h / forsaken
scenario-stopped-or-ambient-24h-floor — [[guardian-faith]], [[guardian]])
and, per tick, sweeps
`State.GuardianOrders` for any active standing order whose `ExpiresTick` the
new tick has reached, emitting `guardian.order_expired` (spec 029, the
`charge_regenerated` pattern — a pure function of state + tick, so a
lapsed watch reproduces on replay with no guardian running — [[guardian-orders]];
since spec 059 a survival watch is skipped by this sweep entirely — it is
non-expiring by origin, not a timed order, so `ExpiresTick` is never
consulted for it). Spec 084 adds the plan-layer sweeps immediately after the
order-expiry sweep, the same once-only idiom: active designations whose
structural predicate holds emit `designation.fulfilled` (slice order), then
per active directive fulfilled-before-expired — `directive.fulfilled` when
the bound designation is fulfilled, else `directive.expired` on TTL elapse
or when no targeted villager remains alive — so exactly one terminal lands
per boundary and a designation fulfilled at tick T fulfills its directives
at T+1 (the documented one-tick lag — [[guardian-designations]]). Spec 085
adds two more sweeps: the prophecy verification sweep (`prophecyEvents`,
immediately after the directive sweep — per active prophecy, fulfil before
fail, once, with per-living-target `OriginReport` companion memories) and
the faith accounting sweep (`faithEvents`, AFTER every faith-source emitter
and BEFORE the scenario rubric/run-end detection — scans THIS tick's batch
for `directive.fulfilled`/`directive.expired`/`agent.died`/
`prophecy.fulfilled`/`prophecy.failed` and emits one `faith.changed` per
source in batch order, skipping emissions the clamp would swallow —
[[guardian-faith]]); spec 107 adds the mission outcome sweep
(`missionEvents`, immediately after the prophecy sweep — per active
mission, completed before failed, once: `guardian.mission_completed` when
every linked designation is fulfilled, `guardian.mission_failed` with
per-link status evidence at the deadline — [[guardian-missions]]);
its reflex fires only on agents idle past `reflexGraceTicks` (120). Since spec 054, an armed scenario world's `stepEvents` also consults its
incident schedule (`scenarioIncidentEvents`) immediately BEFORE `gruStep` —
a scheduled `gru.emerged` preempts that night's random emergence roll, so
never two spawn mechanisms in one night — and its rubric evaluator
(`scenarioRubricEvents`) immediately AFTER every emitter and BEFORE run-end
detection, emitting `curriculum.exercise_passed` (+ same-batch
`curriculum.stage_unlocked`) at the exercise's pass boundary; an ambient
world (no scenario armed) enters neither branch, byte-identical to pre-054
— [[scenario-machinery]] owns the whole subsystem (grown to four incident
kinds by spec 077, with the rubric emitter generalized to per-exercise dawn
boundaries). `stepEvents` also runs the
[[gru]]'s whole turn (`gruStep`) each tick, and the heartbeat's near-death memory
names "the gru" as the cause when the last wound was recent. Since spec 104,
`gruStep`'s movement half is gated on `AmbientCoalescing()`: under the
regime it emits no `gru.moved` at all — the shared `gruMoveDecision` runs
instead inside the derived-progress engine's own per-beat sweep
(`advanceGruBeat`, `internal/sim/advance.go`), behind a `Gru.Done`
watermark stamped by the `gru.emerged` arm — while a legacy world keeps the
per-beat `gru.moved` emission byte-identically; the attack half (and its
inline witness-death handling above) is untouched either way. Immediately
AFTER `gruStep` and before the governance/social beats, spec 077 adds the
stranger's turn (`strangerStep`, `internal/sim/stranger.go` — order pinned
by test): nil-check first, so an ambient world where no `stranger.arrived`
ever landed pays one comparison and emits nothing; while abroad it seeks
the nearest unattended store (greedy, `gruProtected`-avoiding), takes
bounded goods on a cooldown, and departs at dawn —
[[event-types-scenario-incidents]] catalogs the family. The per-minute social beat
(`socialEvents`, [[social-fabric]]) runs the adjacency ladder — repay an open
debt, give to a starving neighbor, or talk (chat-while-working, cooldown-bounded)
with a verbatim rumor fallback — and the hourly due-check breaks overdue debts
(also emitting a `norm.violated` when a repay-debts norm is in force — [[governance]]).
`stepEvents` further runs the whole governance layer (TASK-13, `governanceEvents` in
`governance.go`): the daily meeting lifecycle — gated since TASK-36 on an
event-sourced meeting convention (convene at the convention's hour with attendee
intent pinning to `attend_meeting`, open, speaking-turn beats, timebox+grace
close; no convention → the per-minute emergent-gathering watch runs instead) —
and the per-minute curfew/exile violation detectors. `attend_meeting` is the one
intent goal the executor sets itself (never planner-choosable): arrival idles at
the meeting place until close, and stale pins clear when the meeting ends.
`stepEvents` stays a pure function of (pre-tick state, map, next tick);
every effect is an event through [[sim-state-reducer]] — the determinism and replay guarantees of
the substrate hold unchanged over the whole layer.

## Connections

Parent note: [[executor]]. [[guardian]] owns the nudge-charge regeneration
this section drives; [[guardian-orders]] owns the standing-order expiry
sweep; [[scenario-machinery]] owns the whole incident/rubric subsystem this
section calls into twice per tick; [[gru]] owns the whole predator turn;
[[social-fabric]] owns the adjacency ladder (debt/give/talk) this section's
per-minute social beat runs; [[governance]] owns the meeting lifecycle,
emergent-gathering watch, and curfew/exile detectors this section drives.
