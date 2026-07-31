---
name: guardian-agentization
description: Spec 102 (TASK-112) — the guardian as a first-class autonomous agent on the villager construct: the "steward" scheduled cognition lane (opt-in steward_cadence_ticks dial, router-gated, sheds before villager survival), the guardian's own memory store + nightly consolidation (shared dream phase), the charter-compiled deliberate-incompetence ceiling, and the structural tutor/world channel split. Load when tracing scheduled guardian turns, the ceiling, or guardian memories.
kind: feature
sources:
  - internal/guardian/angel.go
  - internal/guardian/ceiling.go
  - internal/guardian/consolidate.go
  - internal/guardian/tutor.go
  - internal/sim/guardian_memory.go
  - internal/sim/tuning.go
  - internal/cognition/registry.go
  - internal/cognition/schedule.go
verified_against: bd13e84095eb7dd15d2911e524dd0ca0d467bda8
---

# Guardian agentization — the caretaker on the villager construct

Spec 102 (TASK-112) moves the guardian onto the **same agent construct**
villagers run (D1): a scheduled cognition lane with a real budget, a
structured memory store with nightly consolidation (spec 098's dream phase
included), a persona surface compiled from charter/skills/capabilities, and
decision-trail telemetry. The player-chat console and the [[guardian-orders]]
machinery are UNCHANGED event-driven doors — agentization **adds** the
scheduled lane beside them (D2/D6), it replaces nothing.

**Opt-in per world (FR-007):** everything here arms only when
`tuning.json` sets `steward_cadence_ticks > 0` ([[world-tuning]] dial,
`internal/sim/tuning.go`: 0 = off = default; nonzero clamps to 600..86400;
negatives clamp to 0 — never opting a world in by accident). A non-opted
world's guardian is byte-identical to pre-102, and pre-102 logs replay
unchanged (all new event types are additive vocabulary).

## The steward scheduled lane (`internal/guardian/angel.go`)

(The spec's D2 "angel" design vocabulary; every SERIALIZED spelling was
de-themed to "steward" by the operator rename ruling 2026-07-30, pre-merge
— Go identifiers keep the descriptive angel* names.)

The `"steward"` decision class (`internal/cognition/registry.go`, contract row
amended in `specs/007-cognition-horizon/contracts/registry.md`): 5 points,
**900-tick budget — deliberately below planner's 1200**, so under saturation
the router sheds the caretaker's ambient turns before villager survival
cognition (D2; pinned both as MaxSafeSpeed ordering and pointwise). Kind
`"steward"` routes cheap-first (`local → cloud` default, backfilled into
pre-102 `llm.json`s) and is a frozen serialized identifier (spec 052 ruling 2), de-themed so it
needs no fiction-sweep annotation at all.

`scheduleAngel` runs on the absorb goroutine after each applied batch (the
`matchOrders` position — live-only by construction, replay runs no
guardian): first sighting of an opted-in world arms one cadence out; a due
tick routes through the SAME `cognition.Route`/`RoutePaused` gate villager
classes use, records suppressions as `cog.outcome` events, and advances due
via the shared `cognition.NextPhasePreservingDue` (the TASK-44 arithmetic,
moved from mind so both lanes share one schedule implementation). An
admitted turn runs through the **shared `runTurn` body** under the
single-flight turn slot — same roster/handler/gate composition as console
and triggered turns — with jobPrefix `steward` (chain
`steward-metatron-<tick>`, visible on the TUI decision trail via the widened
`-metatron-` attribution in `internal/tui/decisions.go`), a `[cadence]`
transcript marker, and a terminal `cog.outcome` (`landed` with a queued
player moment when an act landed; `adapted` for a quiet observe-only turn).
D6 holds structurally: nothing on this lane emits `guardian.order_triggered`
— the order door keeps its one arbiter.

## The guardian memory store (`internal/sim/guardian_memory.go`)

`State.GuardianMemories` is the guardian's own store on the **shared
`sim.Memory` model** (D5): entries land as `guardian.memory_added`
(reducer-capped at 400 bytes text / 400 memories, lowest-salience-first
eviction), vectors attach via `guardian.memory_embedded` (the SAME
[[memory-retrieval]] embedder driver serves both stores), and the working
window into the turn prompt is the SAME deterministic top-K selector
(`sim.SelectMemories`, seat `sim.GuardianSeat`). Entry points
(`internal/guardian/consolidate.go` `recordMemory`, agentization-gated):
landed acts, watch fires/moments, the 6-hour digest lines (which on
agentized worlds land as memories INSTEAD of soul appends — **soul.md stays
the persona seed, not the memory log**), and console exchanges (recorded as
the guardian's OWN reply words — player text keeps its single sink, the
prompt).

Nightly, on `sim.night_started`, the un-consolidated tail
(`GuardianEpisodicBuffer`, high-water `GuardianMemUpTo`) consolidates
through the shared machinery: `sim.PlanDream` (the spec 098 geometry pass,
[[private-dreams]]) lands clear outcomes as `guardian.salience_revised`/
`guardian.memory_merged`; one `KindConsolidation` call (the shared
"consolidation" class and route) returns the m-label/g-label contract
parsed by the shared `sim.ParseMemLabel`/`ParseRoutineLabels`/
`FirstJSONObject` (moved from mind — one parser, two nightly drivers); the
batch lands promote/fade/gist plus the `guardian.consolidated` marker
(shared `ConsolidationAccepted`/`Rejected` vocabulary).

## The deliberate-incompetence ceiling (`internal/guardian/ceiling.go`)

The operator-adopted D3 ruling as **charter-compilation data**:
`angelCharterLifted` derives the ceiling bit from the EFFECTIVE charter text
each scheduled turn runs under (preset constants and legacy seeds = ceiling
ON; any player-authored revision lifts it — the same derivation as the
recorded `charter_observed` default flag, so they can never disagree).
`applyAngelCeiling` narrows the SCHEDULED turn's grant through the SAME
`intersectGrant` machinery as the stage ceiling: default → the modest
read/counsel set (`explain`, `survey_site`, `brief_myths`; zero miracle
kinds — no spend, no watches, no plans); authored → the full world grant
minus the clock triple (the clock is the player's at ANY ceiling). Capped
tools are structurally absent from declaration, prose, and door alike. The
ceiling caps **initiative only**: console and order-triggered turns never
pass through it, and the modest frame says so verbatim. Bundle tools are
not part of the scheduled lane's roster at ANY ceiling (planning-tier
ruling 2026-07-30) — they remain console/order-driven; revisit via a
future card. Two compile-time
initiative frames (INV-1 appended-last) carry the doctrine into the prompt.
Spec 107 composes ONE layer beside the ceiling: with an active MISSION
(the player's standing pre-authorization, [[guardian-missions]]),
`applyMissionPursuitGrant` union-adds the world-granted pursuit verbs
(survey, the plan verbs, `work_miracle` at the world's full kind grant,
`note_mission_progress`) back into the scheduled roster at ANY ceiling —
initiative stays capped exactly as above (clock, orders, nudges,
accept/cancel_mission all still absent), and a third frame
(`guardianAngelMissionFrame`) carries the carve-out; mission-free
scheduled turns are byte-identical to spec 102.

## The structural tutor/world split (`internal/guardian/tutor.go`)

D4 as a type-level fact: `tutorSurface` is the tutor channel's entire
capability — inert descriptor data in, strings out; no injector, no loop
control, no funcs/chans/interfaces anywhere in its transitive field graph
(reflection-enforced). Converse remains the final-text channel (no roster
entry, no handler). A pure tutor exchange lands only `cog.*` telemetry
(reducer no-ops): world state is byte-identical after it — zero charges,
zero faith, zero rubric-visible facts (`tutor_isolation_test.go`).

## Event vocabulary and surfaces

Seven additive event types (rows: [[event-types-guardian-memory]]), all on
the `InjectSocial` whitelist and in `sim.PayloadCatalog`, with chronicle
digest entries ([[tui-chronicle-feed-guardian-digests]] family). Decision
trail: [[tui-villagers-tab]]'s guardian sentinel now attributes every
`-metatron-` chain (turn/watch/steward).
