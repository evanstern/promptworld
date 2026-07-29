---
name: mental-maps
description: Per-agent private spatial knowledge (spec 041) — retires villager omniscience via a private MentalMap. Overview + spec history here; detail lives in three children — [[mental-map-model]] (the type/freshness/derived bookkeeping), [[mental-map-perception]] (resolution + perception sweep + search goal + graves), [[mental-map-propagation]] (talk/vision telling + prompt rendering). Load this note first for orientation.
kind: component
sources:
  - internal/sim/mentalmap.go
  - internal/tool/registry.go
verified_against: 63390f122bdf4e1b7abf518a8be83de725f06230
---

# Mental maps

Spec 041 retires the villagers' omniscience: before this feature, [[reflex-policy]]'s
resolvers and [[agent-mind]]'s prompt read the world's ground truth directly, so
every villager always knew where every resource, structure, and neighbor stood.
Each agent now carries a private `MentalMap` — explored terrain plus a list of
known place-facts with provenance and a last-seen tick — and target resolution,
the prompt's world description, and hail/talk resolution all read through it
instead. Two villagers with different histories now see different worlds.

## How it works

Each agent's `MentalMap` (`internal/sim/mentalmap.go`) replaces the old
ground-truth reads with three parts: an `Explored` terrain bitmap, a sorted
list of known `PlaceFact`s with provenance and a last-seen tick, and a list
of `PeerSighting`s. This note now summarizes three split-off children that
carry the mechanics:

[[mental-map-model]] carries the type itself — the `Explored` bitmap and
`PlaceFact`/`PeerSighting` fields, the closed `Kind` vocabulary (including
spec 044's `grave`), read-time freshness horizons (`factHorizonVolatileTicks`/
`factHorizonDurableTicks`), the knowledge predicates that test them
(`knownFreshFact`, `knowsAnyFresh`, `warmKnownPredicate`), the derived
(eventless) bookkeeping that grows the bitmap and peer sightings on movement,
replay determinism, and the genesis/v3→v4-migration knowledge grants that
seed a world's maps.

[[mental-map-perception]] carries how the map is READ by target resolution
(`nearestKnown`/`nearestKnownAdjacentTo`, `talk_to`/`seek` against
`peerSightingOf`) and GROWN by the perception sweep (`perceptionEvents`,
`agent.saw`/`agent.map_corrected`), plus the search goal's `nearestFrontier`
frontier fallback (US4) and graves as a perceived fact kind (spec 044).

[[mental-map-propagation]] carries the three channels knowledge travels
through besides direct perception: `tellablePlaces`/`social.place_told` in
talk (US5), the divine `send_vision` place-grant (FR-014), and rendering the
acting agent's own map into the decision prompt's `known_places` block (US2).

## Connections

[[reflex-policy]] is this note's primary consumer — every goal resolver and
the reflex ladder read through `nearestKnown`/`knowsAnyFresh`/
`warmKnownPredicate`/`peerSightingOf`, and `search`/`nearestFrontier` live in
its files. [[executor]] hosts the perception sweep (`perceptionEvents`) and
the talk sidecar (`tellablePlaces`) that grow and correct the map, plus the
derived explored/sighting bookkeeping on movement events. [[agent-mind]]
renders `knownPlaces` from it in the prompt — since spec 043 as
[[decision-context]]'s `known_places` block (`internal/mind/context.go`) —
and re-arms the planner on a targeted `agent.map_corrected`. [[sim-state-reducer]] carries `Agent.Map` and
the four knowledge-event Apply arms. [[event-types]] catalogs
`agent.saw`/`agent.map_corrected`/`social.place_told`/`guardian.place_revealed`.
[[social-fabric]] is where the place-telling sidecar rides, beside rumors and
gifts. [[guardian]] is the `send_vision` place grant's door, sharing this
note's closed vocabulary via `internal/tool/registry.go`'s `placeFactKinds`.
[[guardian-miracles]] shares the `rebaseTicks` SHIFT/KEEP taxonomy and is
what a miracle-moved villager's derived bookkeeping updates.
[[world-migration]] carries the v3→v4 knowledge grant; [[world-save-directory]]
is the format-version gate this bumped to v4. [[tui-client]] renders the four
new event types through the raw digest feed with no dedicated pane of its
own. [[chronicle]] narrates three of the four events (not `agent.saw`, too
chatty). [[tool-registry]] declares `search` and `send_vision`'s place-grant
params.

## Operational notes

A villager's knowledge only ever grows or corrects through recorded events —
there is no silent forgetting beyond the read-time freshness horizon, and a
stale fact is invisible to resolvers/prompt without being physically removed
until a correction, a death, or — since spec 081 — a chop/quarry the villager
performed or watched in radius removes it ([[mental-map-perception]]).
`internal/sim/mentalmap_test.go` is
the subsystem's own suite, alongside the v3→v4 migration, rebase-taxonomy,
determinism, and vision-place-reveal coverage [[testing-strategy]] tracks.
Exact freshness-horizon values are tuning (clarify Q5), soak-validated
rather than derived from first principles.

## Spec 086 — perception payloads carry named refs

`agent.saw`/`agent.map_corrected`'s `agent`, `social.place_told`'s
`from`/`to`, and `guardian.place_revealed`'s `agent` are `sim.AgentRef` on
the wire. `PlaceFact.Source` (`src,omitempty`) deliberately stays a bare
int — `PlaceFact` is state-resident (mental maps) AND rides four payloads,
so the R2 no-refs-in-state invariant wins; it is a frozen, rationale-carrying
allowlist entry in `TestPayloadAgentRefSweep` (the payloads' top-level
actor fields carry the refs).
