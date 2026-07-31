---
name: executor-social-perception
description: Guarded plans (timed conditional multi-step intents), hails (talk_to pause/close/found), and the situated-memory/origin provenance every emitted memory carries; routes to [[executor-perception-observation]] for the perception sweep and spec-097 arrival observations. Load for planner-plan timing, hail/talk mechanics, or memory provenance questions.
kind: component
sources:
  - internal/sim/executor.go
  - internal/sim/plan.go
  - internal/sim/memory.go
  - internal/sim/observe.go
verified_against: fc1a8314f3f71a33c5e2145c914d5cbb511d9196
---

# Executor — guarded plans, hails, perception, and memory provenance

Child of [[executor]] — the planner- and perception-facing surfaces: timed
conditional plans, the hail/talk courtesy pause, the per-agent perception
sweep, and the situated/origin provenance every emitted memory carries.

## How it works

**Guarded plans** (TASK-32, `plan.go`): a planner reply may land as a short
conditional plan — up to `PlanStepCap` (3) `PlanStep`s, each with a goal, an
optional `When` guard, and `Until` validity deadline (default window
`PlanDefaultWindowTicks`, 2 game-hours). The steps live on `Agent.Plan` in
deterministic state (`agent.plan_set`); each idle tick the executor evaluates
the head step via `planStepEvents` *before* falling through to the reflex:
holding (guard false, window open) emits nothing, expiry or a failed goal
resolution clears the whole plan with `agent.plan_expired` (a broken sequence
is not resumed), and firing emits `agent.plan_step_started` plus the intent
with source `plan`. No model runs at firing — timed guards are the sole
act-at-time-T mechanism. `Agent.Generation` (also TASK-32) counts high-salience
interrupts: the reducer bumps it on memories at/above `GenerationBumpSalience`
(9); in-flight thoughts snapshotted under an older generation are superseded
when they land ([[cognition]]).

**Hails** (TASK-47, `hail.go`): a `talk_to` landing flags its target down —
`social.hailed` pauses the target for `hailWindowTicks` (480, 8 game-minutes)
so the hailer can close distance. The per-tick `hailStep` sweep runs *before*
the per-agent loop: a hailer within Manhattan 1 of its paused target founds
the talk deterministically (`social.hail_met` + the `talkEvents` shape,
bypassing the ambient `canTalk` cooldown — met is checked before expiry so an
on-time arrival wins the edge tick); otherwise closing emits
`social.hail_expired` and the target resumes untouched. A paused agent (`hailPaused`) skips the
reflex, plan-step evaluation, and en-route movement, but keeps decaying,
keeps its intent/plan as they were, and still works if already standing on
its intent target. `hailable` (same file) is the exemption predicate: dead,
asleep, already-hailed, actively-hailing, meeting-pinned, or beyond
`hailRadius` (64) targets are never paused. A plan-step `talk_to` firing
hails exactly as a planner landing. The ambient beat's talk founding is
shared with the sweep via `talkEvents` (`executor.go`). Since spec 041,
`talkEvents` also carries a place-knowledge sidecar (US5, [[mental-maps]]):
every founded talk, hail-founded included, exchanges up to `placeTellCap`
fresh facts per direction the other party lacks or holds staler
(`tellablePlaces`), landing one `social.place_told` each way plus a companion
situated memory on both sides.

**Perception & grounded observations** — split to
[[executor-perception-observation]] (spec 089 size budget): the spec-041
perception sweep (`perceptionEvents` — at most one `agent.saw` + one
`agent.map_corrected` per beat, spec 081's act-time map removals, the
chop/quarry act memories) and the spec-097 grounded arrival observations
(`agent.place_observed` on intent-completing arrivals, exhaustive within
`placeScanRadius`, `Agent.LastObs` dedup window, Origin-`observed` companion
memory, mind-side belief reconciliation). That child owns the full
mechanics; this note keeps the memory-provenance vocabulary below.

**Memory emission**: the executor emits `agent.memory_added` events from the
salience table in `memory.go` ([[agent-mind]]) alongside memorable
happenings — since spec 019 (US1) every one is *situated*. The emission
sites go through the situated constructors (`situatedMemoryEvent`/
`situatedMemoryToned`/`situatedMemoryAboutEvent`, `memory.go`; T008b removed
the pre-019 bare `memoryEvent`/`memoryEventToned`/`memoryAboutEvent` once
every site had migrated, so no sim memory can be emitted unsituated). Each
bakes a `Where` — the acting-or-witnessing agent's tile via `PlaceAt` →
`describePlace`, a deterministic Manhattan-radius nearest-feature scan that
names a station ("the fire") or terrain ("the woods" — since spec 068,
`featureDesc` also names marsh/sand tiles individually, "the marsh"/"the sand
flat", as NOTABLE terrain rather than the generic "" plain grass gets,
[[worldmap-generation]]) — and, for a driven personal act, a `Why` (the
completing intent's `Reason`, `""` for reflex/witness) into the
`MemoryAddedPayload`, composing both into the memory text via `situateText`;
the [[chronicle]]/scribe render what the payload carries with no
re-derivation, so replay is byte-identical. Build completions situate
through `placeForBuild`, which excludes the just-built structure kind from
the scan so "Built a fire" resolves to the tile as it was ("at the woods
(x,y)"), never "at the fire" (T024). Gossip/witness memories carry no `Why`
— a witness did not drive the act.

Since spec 030, all three situated constructors also take a required `origin`
param — the emission-stamped provenance class the compiler forces every call
site to declare, so no new memory site can land unstamped. `origin` is a
closed vocabulary (`memory.go`): `OriginAction` (an own executed act),
`OriginWitness` (a seen event — `situatedMemoryAboutEvent`'s usual value),
`OriginReport` (learned of at any distance, e.g. a chest-owner's theft
notification), `OriginOmen` (a delivered omen/dream/working — the guardian's
FROZEN payload value, spec 052 ruling 2), `OriginGist` (a conversation
summary into memory), `OriginDigest` (a nightly day-gist), and — spec 097 —
`OriginObserved` (a grounded arrival observation, first-person); an
absent/legacy origin (`""`, any pre-030 payload) classifies secondhand, the
conservative direction. `DirectPerception(origin)` is the pure helper —
true only for `OriginAction`/`OriginWitness`/`OriginOmen`/`OriginObserved` — the belief
validator ([[nightly-consolidation]]) reads it to decide whether a memory
counts as direct perception; it is the ONLY signal used, no text inspection.
`Memory.Origin` (`omitempty`) rides the same copied-at-Apply, never-re-derived
doctrine as `Where`/`Why`/`Conv`, so replay stays byte-identical and a pre-030
Memory (field absent) reduces to `Origin` `""`.

## Connections

Parent: [[executor]]. Child: [[executor-perception-observation]] (the
perception sweep + spec-097 arrival observations). [[cognition]] consumes
`Agent.Generation` to supersede in-flight thoughts; [[mental-maps]] owns the
per-agent map the sweep populates/corrects and the talk sidecar's
place-knowledge exchange;
[[agent-mind]]/[[chronicle]] render situated-memory payloads with no
re-derivation; [[nightly-consolidation]] hosts the belief validator reading
`Origin`/`DirectPerception` off these memories; [[sim-loop]] is the door
(`InjectIntent`) a planner plan or `talk_to` lands through.

## Operational notes

Spec 030's `Origin` stamping is exercised by `origin_test.go`.
