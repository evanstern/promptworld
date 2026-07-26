---
name: executor-social-perception
description: Guarded plans (timed conditional multi-step intents), hails (talk_to pause/close/found), the per-agent perception sweep populating mental maps, and the situated-memory/origin provenance every emitted memory carries. Load for planner-plan timing, hail/talk mechanics, or memory provenance questions.
kind: component
sources:
  - internal/sim/executor.go
  - internal/sim/plan.go
  - internal/sim/memory.go
verified_against: 4c66d240b2715706964f02cfd2396256c9957d8e
---

# Executor — guarded plans, hails, perception, and memory provenance

Child of [[executor]] — the planner-facing and perception-facing surfaces:
timed conditional plans, the hail/talk courtesy pause, the per-agent
perception sweep, and the situated/origin provenance every emitted memory
carries.

## How it works

**Guarded plans** (TASK-32, `plan.go`): a planner reply may land as a short
conditional plan — up to `PlanStepCap` (3) `PlanStep`s, each with a goal, an
optional `When` guard, and an `Until` validity deadline (default window
`PlanDefaultWindowTicks`, 2 game-hours). The steps live on `Agent.Plan` in
deterministic state (`agent.plan_set`); each idle tick the executor evaluates
the head step via `planStepEvents` *before* falling through to the reflex:
holding (guard false, window open) emits nothing, expiry or a failed goal
resolution clears the whole plan with `agent.plan_expired` (a broken sequence
is not resumed), and firing emits `agent.plan_step_started` plus the intent
with source `plan`. No model runs at firing time — timed guards are the sole
act-at-time-T mechanism. `Agent.Generation` (also TASK-32) counts
high-salience interrupts: the reducer bumps it on memories at or above
`GenerationBumpSalience` (9), and in-flight thoughts snapshotted under an
older generation are superseded when they land ([[cognition]]).

**Hails** (TASK-47, `hail.go`): a `talk_to` landing flags its target down —
`social.hailed` pauses the target for `hailWindowTicks` (480, 8 game-minutes) so
the hailer can close distance. The per-tick `hailStep` sweep runs *before* the
per-agent loop: a hailer within Manhattan 1 of its paused target founds the talk
deterministically (`social.hail_met` + the `talkEvents` shape, bypassing the
ambient `canTalk` cooldown — met is checked before expiry so an on-time arrival
wins the edge tick); otherwise the window closing emits `social.hail_expired`
and the target resumes untouched. A paused agent (`hailPaused`) skips the
reflex, plan-step evaluation, and en-route movement, but keeps decaying,
keeps its intent and plan exactly as they were, and still works if already
standing on its intent target. `hailable` (same file) is the exemption
predicate: dead, asleep, already-hailed, actively-hailing, meeting-pinned, or
beyond `hailRadius` (64) targets are never paused. A plan-step `talk_to` firing
hails exactly as a planner landing does. The ambient beat's talk founding is
shared with the sweep via `talkEvents` (`executor.go`). Since spec 041,
`talkEvents` also carries a place-knowledge sidecar (US5, [[mental-maps]]):
every founded talk, hail-founded included, exchanges up to `placeTellCap`
fresh facts per direction the other party lacks or holds staler
(`tellablePlaces`), landing one `social.place_told` each way plus a companion
situated memory on both sides.

**Perception** (spec 041, `perceptionEvents`): each awake living villager,
on the same staggered per-agent cadence movement uses (a fifth of a full
per-tick sweep, T034's hot-path relief), diffs ground truth within
`witnessRadius` against its own `Agent.Map` and emits at most one `agent.saw`
(new/changed structures, piles, standing trees, unharvested forage, unquarried
rock, water shoreline, dens) and one `agent.map_corrected` (remembered fresh
facts whose place has genuinely vanished — a chopped tree, a quarried-out
outcrop, a drained pile, a removed structure; a merely-harvested forage spot
or cooling den is not gone, only unavailable, so it stays). A correction's
gone facts each ride a companion situated first-person discovery memory
(`mapCorrectedText`, `salMapCorrected`) in the same batch — memories accrete
only via `agent.memory_added`, never appended directly by a reducer arm. Pure
function of (state, map, tick): `stepEvents` reads, never mutates.
[[mental-maps]] owns the mental-map subsystem this sweep populates and
corrects; the executor's role is only the perception beat that drives it.

**Memory emission**: the executor also emits `agent.memory_added` events from the salience table in
`memory.go` ([[agent-mind]]) alongside memorable happenings — and since spec 019
(US1) every one is *situated*. The emission sites now go through the situated
constructors (`situatedMemoryEvent`/`situatedMemoryToned`/`situatedMemoryAboutEvent`,
`memory.go`; T008b removed the pre-019 bare `memoryEvent`/`memoryEventToned`/
`memoryAboutEvent` once every site had migrated, so no sim memory can be emitted
unsituated). Each bakes a `Where` — the acting-or-witnessing agent's tile via
`PlaceAt` → `describePlace`, a deterministic Manhattan-radius nearest-feature scan
that names a station ("the fire") or terrain ("the woods" — since spec 068,
`featureDesc` also names marsh/sand tiles individually, "the marsh"/"the sand
flat", as NOTABLE terrain rather than the generic "" plain grass gets,
[[worldmap-generation]]) — and, for a driven
personal act, a `Why` (the completing intent's `Reason`, `""` for reflex/witness)
into the `MemoryAddedPayload`, and composes both into the memory text via
`situateText`; the [[chronicle]]/scribe render what the payload carries with no
re-derivation, so replay is byte-identical. Build completions situate through
`placeForBuild`, which excludes the just-built structure kind from the scan so
"Built a fire" resolves to the tile as it was ("at the woods (x,y)"), never
"at the fire" (T024). Gossip/witness memories carry no `Why` — a witness did not
drive the act.

Since spec 030, all three situated constructors also take a required `origin`
parameter — the emission-stamped provenance class the compiler now forces every
call site to declare, so no new memory site can land unstamped. `origin` is a
closed vocabulary (`memory.go`): `OriginAction` (an own executed act),
`OriginWitness` (a seen event — `situatedMemoryAboutEvent`'s usual value),
`OriginReport` (learned of at any distance, e.g. a chest-owner's theft
notification), `OriginOmen` (a delivered omen/dream/working — the guardian's
FROZEN payload value, spec 052 ruling 2), `OriginGist` (a
conversation summary written into memory), and `OriginDigest` (a nightly
day-gist); an absent/legacy origin (`""`, any pre-030 payload) classifies as
secondhand, the conservative direction. `DirectPerception(origin)` is the pure
helper — true only for `OriginAction`/`OriginWitness`/`OriginOmen` — that the
belief validator ([[nightly-consolidation]]) reads to decide whether a memory
counts as direct perception; it is the ONLY signal that decision uses, no text
inspection. `Memory.Origin` (`omitempty`) rides the same copied-at-Apply,
never-re-derived doctrine as `Where`/`Why`/`Conv`, so replay stays byte-identical
and a pre-030 Memory (field absent) reduces to `Origin` `""`.

## Connections

Parent note: [[executor]]. [[cognition]] is `Agent.Generation`'s consumer for
superseding in-flight thoughts; [[mental-maps]] owns the per-agent map this
section's perception sweep populates and corrects, and the talk sidecar's
place-knowledge exchange; [[agent-mind]]/[[chronicle]] render what a
situated-memory payload carries with no re-derivation; [[nightly-consolidation]]
hosts the belief validator that reads `Origin`/`DirectPerception` off these
memories; [[sim-loop]] is the door (`InjectIntent`) a planner-issued plan or
`talk_to` lands through.

## Operational notes

Spec 030's `Origin` stamping is exercised by `origin_test.go`.
