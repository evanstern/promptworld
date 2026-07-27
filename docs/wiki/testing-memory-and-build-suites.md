---
name: testing-memory-and-build-suites
description: Memory-origin/belief-decay proofs (direct-vs-secondhand classification, half-life confidence curve, reinforcement re-anchoring) and the spec-032 walls/axes/paths build-verb unit suites. Split out of [[testing-strategy]].
kind: pattern
sources:
  - internal/sim/origin_test.go
  - internal/sim/belief_evidence_test.go
  - internal/sim/belief_decay_test.go
  - internal/sim/belief_reinforced_test.go
  - internal/sim/wall_test.go
  - internal/sim/axe_test.go
  - internal/sim/path_speed_test.go
  - internal/mind/provenance_test.go
  - internal/mind/belief_read_sites_test.go
verified_against: c61cd6c04ddfcd2a976c14a49ba071e8fd768a73
---

# Memory-provenance & build-verb suites

**Memory-provenance and belief-decay suites** (spec 030, [[guardian]]):
`internal/sim/origin_test.go` proves the `DirectPerception` classifier's closed
vocabulary (action/witness/omen direct; report/gist/digest/absent secondhand),
that every situated constructor stamps `Origin` at emission, that the reducer
copies it verbatim, and that a pre-030 memory (no `origin` field) round-trips
byte-identically. `internal/sim/belief_evidence_test.go` proves belief
formation stamps `Belief.Reinforced` to the formation tick regardless of
evidence, that a later revision leaves it untouched (US2 only re-anchors on
direct evidence), and that a log of coerced beliefs replays byte-identically.
`internal/sim/belief_decay_test.go` pins `EffectiveConfidence`'s half-life
curve to the tick (a pure, computed-on-read function — nothing stored ever
mutates), including a fractional-half-life midpoint proving the curve is
continuous, not integer-stepped, and a legacy no-stamp belief grandfathered to
undecayed. `internal/sim/belief_reinforced_test.go` proves the
`agent.belief_reinforced` reducer arm re-anchors a held belief's clock and is a
total no-op against a vanished belief ID, and that a log containing the event
replays byte-identically. On the mind side, `internal/mind/provenance_test.go`
proves the consolidation user prompt instructs the model to cite evidence and
reserve "witnessed" for direct perception, and that deterministic
provenance enforcement coerces rather than rejects; `internal/mind/belief_read_sites_test.go`
proves the nightly consolidation held-beliefs block is the one documented
exception that renders EFFECTIVE (not stored) confidence and marks a faded
belief while still listing it by ID.

**Walls/axes/paths unit suites** (spec 032, [[tool-registry]]): `internal/sim/wall_test.go`
covers wall build/chip/repair/demolish across both materials (plank vs stone
HP and material cost); `internal/sim/axe_test.go` covers `craft_axe` and the
axe-assisted chop/quarry yield and durability countdown to breakage;
`internal/sim/path_speed_test.go` covers a path tile's travel-speed doubling
over a deterministic grass corridor. These sit alongside
`TestReplayByteIdentityWallsAxesPaths` and `TestPre032SnapshotLoadsUnchanged`
above as the feature's full proof. Spec 038 (loud build failure & occupancy
tolerance, TASK-91, [[executor]]) rewrote `wall_test.go`'s occupancy coverage
into a `driveWithOccupant` (per-tick scripted occupant placement) matrix:
`TestWallOccupancyGuard` proves a permanent squatter defers completion then
fails loudly (`agent.build_failed{reason: "site blocked too long"}`) exactly
`wallOccupancyGraceTicks` past the due tick, never a wall, never a spend;
`TestWallBuildToleratesPasserby` proves a mid-work crossing that clears before
the due tick no longer cancels the build; `TestWallBuildDefersThenCompletes`
proves a departure inside the grace window lets completion land on the first
clear tick and never on an occupied one; and `TestWallBuildSiteVanishedFailsLoud`
proves a vanished reserved-tile site fails loudly immediately
(`reason: "site no longer buildable"`) with a same-tick paired failure memory,
never a bare `intent_done`. `builderFailure`, a shared log-scanning helper, is
what the first two of these plus the site-loss test read to assert the count,
reason, tick, and paired-memory invariants. Since spec 041 made `repair`/`demolish`
knowledge-gated ([[reflex-policy]]), the fixture helper `grantStructureFacts`
seeds an agent's map with the walls/structures its resolver test needs to
already KNOW before the ground-truth assertions run.

## Connections

Part of the [[testing-strategy]] suite map (split out during the corpus-spec v2
restructure); see that note for the full layered test picture and links to
sibling suites.
