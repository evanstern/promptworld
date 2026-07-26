# Research: First-person harvest memory (spec 081)

No NEEDS CLARIFICATION markers survived the spec (the operator decision of
2026-07-26 resolved voice and scope). Phase 0 therefore records the design
decisions and the code evidence they rest on.

## D1 — Where the map-fact removal lives

**Decision**: inside the existing `agent.chopped` / `agent.quarried` reducer
arms (`internal/sim/state.go:1057` / `:1176`), deriving actor + witnesses from
the same pre-mutation state the emitter checked. No new event types.

**Rationale**: the reducer already re-derives behavior from pre-mutation state
(axe yield at state.go:1040-1054, "the spear-hunt precedent"), so an in-arm
derivation is the established idiom and stays a pure function of
(event, prior state) — replay-deterministic under one code version.
`MentalMap.removeFact` (mentalmap.go:302-315, the `agent.map_corrected`
arm's primitive) is reused verbatim; it preserves canonical order and nils a
drained list, so canonical state bytes stay stable. The mind's replica applies
the same reducer, so witness maps update there for free.

**Alternatives considered**:
- *New companion event (e.g. `agent.fact_removed`) per affected agent*: makes
  the log self-describing but adds an event type for silent bookkeeping,
  contra the "no new types for state-derived behavior" posture
  (contracts/events.md precedent cited at executor.go quarry arm), and bloats
  the log by O(witnesses) per harvest.
- *Emit `agent.map_corrected` for on-scene parties*: semantically wrong (the
  spec's whole point — watching is not discovering) and would keep minting
  the discovery memory FR-006 forbids.

## D2 — Witness set definition

**Decision**: every agent `w ≠ actor` with `!Dead && !Asleep` and
`abs(w.X-x)+abs(w.Y-y) <= witnessRadius` at the event tick, matched against
their map by `(kind, x, y)` regardless of provenance.

**Rationale**: identical predicate to the perception sweep's eligibility
(executor.go:452 `a.Dead || a.Asleep || a.Map == nil` skip; diamond radius
throughout) — one perceptual reality, no second constant (spec Assumption).
Provenance-blind removal implements the spec's hearsay edge case: watching
the tree fall overrides whoever told you about it. `removeFact` is a no-op
when the fact is absent, which covers the actor-never-knew edge case.

## D3 — Actor act memory

**Decision**: executor companion event at the chop/quarry emit sites —
`situatedMemoryEvent(nextTick, i, salChop|salQuarry, where, in.Reason,
OriginAction, "Felled the tree at (%d,%d)." | "Quarried the outcrop at
(%d,%d).")` riding the same batch as the act event. New constants
`salChop = 4`, `salQuarry = 4` in memory.go's salience block.

**Rationale**: exact shape of the hunt precedent (executor.go:1221,
`salHunt = 4` at memory.go:238 — "Hunted at the den and came back with
meat."): first-person, origin action, situated by the actor's stand tile,
salience in the low non-generation-interrupting band, well below the
`rumorMinSalience`/generation-bump thresholds. Memories accrete only via
`agent.memory_added` (TestMemoriesAccrete), which companion events satisfy.
This consciously supersedes the "completed chops mint no memory"
spam-avoidance posture for these two acts (operator decision recorded in the
spec); the deferred re-evaluation of memory-worthiness is out of scope.

## D4 — Absorb-trigger parity for on-scene witnesses (FR-007)

**Decision**: extend the mind driver's absorb switch (internal/mind/mind.go)
so `agent.chopped` / `agent.quarried` also arm any agent within
`witnessRadius` of the cleared tile whose current intent's
(TargetX,TargetY) or (ResX,ResY) equals the cleared coordinates — mirroring
the `agent.map_corrected` arm-on-matching-intent logic at mind.go:290-311.

**Rationale**: today the witness's re-arm rides the later correction event;
silent removal would otherwise leave them walking to a stump (spec edge
case). The actor needs nothing new — `agent.chopped` already arms its actor
(mind.go:263 case list). The replica's positions at absorb time are the
event-tick positions (replica applies the same reducer batch), so the radius
check reads the same state the reducer used.

## D5 — Same-tick sweep race

**Decision**: no production change; add a regression test pinning the
ordering.

**Rationale**: the sweep's correction half reads pre-batch state — at the act
tick the tree is still present in ground truth (the `cleared` overlay map at
executor.go:466 is built from pre-batch `s.Cleared`), so no correction can
fire in the act's own batch; by any later sweep the reducer has already
removed on-scene facts. The test exists to keep this true if sweep scheduling
ever changes (FR-006's "including a perception pass landing on the act tick").

## D6 — Scope guard: forage / piles / structures / miracles

**Decision**: untouched (FR-010; spec Assumption for miracles).

**Rationale**: forage regrows and never corrects (`groundFactPresent` treats
it as availability, not existence); pile/structure lifecycles have their own
perception behavior; miracle terrain removal is not an agent act — mysterious
loss is the correct narrative for divine intervention. `removeTerrain`
(miracles.go:546) shares the overlay vocabulary but is deliberately NOT given
map-removal behavior.

## D7 — Implementation tier

**Decision**: `spec-implementer` pinned **Opus 4.8**.

**Rationale** (constitution V rubric): doctrine-adjacent behavior change
(memory formation + perception doctrine, the narrative substrate) and a
reducer/replay-contract surface spanning `internal/sim` executor + reducer +
`internal/mind` absorb — cross-package with determinism stakes. Recorded on
TASK-159.
