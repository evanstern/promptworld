---
name: testing-unit-harness
description: Package-level determinism harness (driveTicks, world-migration v1-v4 fixture suites, whole-feature byte-identity replay) and the loop-era live-vs-replay proof through a real Loop+loopMind. Split out of [[testing-strategy]] (corpus-spec v2); load for unit/replay-layer testing questions.
kind: pattern
sources:
  - internal/sim/sim_test.go
  - internal/sim/migrate_test.go
  - internal/sim/whole_feature_test.go
  - internal/world/migrate_test.go
  - internal/mind/replay_test.go
verified_against: b6a20eaa4da1073a69959a5aff69591d931103a9
---

# Unit determinism & replay harness

**Unit determinism harness** (`internal/sim/sim_test.go`): `driveTicks` replicates
the loop's semantics minus the real-time scheduler — commands injected at exact tick
boundaries, terrain threaded through exactly as the live loop does. Now proven over
the full [[executor]]: 30k–40k-tick determinism and replay harnesses, plus behavior
suites — multi-step intent chains with zero input (AC#1), needs decay + self-feeding
and starvation death with recorded cause (AC#2), night warmth mechanics and exposure
death (AC#3), and a two-day unattended village survival run on multiple seeds.
`TestDeterminismSameSeedSameTimeline` additionally diffs each agent's canonical
mental-map bytes across two same-seed runs (spec 041, T010) — the state hash
already covers them, but a targeted per-agent diff localizes a map-determinism
regression instead of just failing the whole-state comparison ([[mental-maps]]).
(Terrain generation has its own determinism/AC suite in `internal/worldmap`, covered
by [[worldmap-generation]].)

Spec 012 and spec 013 each added their own fixture suite spanning both save-format
packages, all in [[world-migration]]'s territory: `internal/sim/migrate_test.go`
builds representative v1 and v2 states and proves both pure transforms'
carry/reset/re-place/spill rules directly, including a v1 fixture that chains both
transforms (1→2→3) in one call; `internal/world/migrate_test.go` drives the full
`Migrate(dir)` ceremony end-to-end against on-disk v1 and v2 fixture worlds (happy
path, replay-from-zero-snapshots determinism, the already-migrated and
already-current guards, uncovered/tolerated event tails, a running-daemon refusal)
for both the v1→v2 and v2→v3 steps. Spec 041 (T009) adds the v3→v4 leg on
both sides, [[mental-maps]]'s knowledge-grant transform: `migrate_test.go`'s
`TestTransformV3GrantsKnowledge` proves the pure `TransformV3State` carries
people and land verbatim while granting each LIVING agent explored terrain
around its position plus witnessed facts for every current structure/pile
(natives, not strangers), leaves a DEAD agent an empty but non-nil map (so a
genesis-seeded state and a migrated one agree on map presence), and mutates
neither its input nor the map argument; `TestTransformV3ChainReducerReplay`
proves the transform's output replays byte-identically through
`world.created` → `world.migrated` from genesis. `world/migrate_test.go`'s
`TestMigrateV3HappyPath` drives the same ceremony end-to-end against an
on-disk v3 fixture world (manifest bump to v4, `world.v3.db` archived, the
covering snapshot's agents each holding the grant), and
`TestMigrateV3ReplayDeterminism` proves deleting every snapshot and
rebuilding from genesis reproduces the post-migration snapshot byte-for-byte,
maps included.

`internal/sim/whole_feature_test.go` carries several byte-identity suites (SC-004/SC-005):
the original spec-012 run, a single scripted-agent script chaining every
resources/food/crafting event kind (quarrying — five bare quarries at
`quarryYieldBare` (1) each, since spec 032 T014 split the old flat `quarryYield`
(2) into bare/axe-assisted tiers and dropped the bare yield to 1 — water, the full
craft chain, both cook stations, bathing, refueling, a spear breaking, a fire
burning out) — rebalanced under spec 013's bulk cap (24) to consume-as-it-goes
rather than hoard a large seeded larder — that replays from genesis to a
byte-identical state hash; and a spec-013 storage suite
(`TestReplayByteIdentityWholeFeatureStorage`) exercising every new
013 event type in one run — `agent.dropped`, `agent.picked_up`, `agent.deposited`,
`agent.withdrew` (both an owner fetch and a non-owner theft with its full companion
batch: `social.chest_taken`, a reason-`theft` `social.relation_changed`, and owner +
witness `agent.memory_added`), `sim.food_rotted`, `agent.built{kind: chest}`, and a
death spill — that also replays to a byte-identical hash. The same file also proves
every new 013 event type no-ops under a pre-013 reducer stub (the unknown-type
convention: an event type the reducer's switch doesn't match falls through to a
total no-op, never an error), so old logs stay safely replayable by builds that
predate a given event kind. `TestReplayByteIdentityWallsAxesPaths` (spec 032 T021,
quickstart scenario 7) is the walls/axes/paths counterpart: one scripted session
chains `craft_axe`, an axe-assisted chop, a `build_wall_plank`, a full
`demolish` (chip then destroy), and a `build_path`, asserts every required event
(`agent.crafted`, `agent.chopped`, `agent.built`, `agent.wall_chipped`,
`agent.wall_destroyed`, and — since spec 038 — `agent.build_failed`) actually
occurred, then replays from genesis (log only) to a byte-identical state hash.
Spec 038 extends the same session with a build that fails LOUDLY: a wall
injected onto the just-built path tile (an unbuildable reserved site) resolves
via `agent.build_failed` plus a paired situated failure memory rather than a
bare `agent.intent_done`, and both new events replay byte-identically alongside
the rest of the session. `TestPre032SnapshotLoadsUnchanged` (spec 032 T021,
research R7) proves a pre-032-shaped snapshot (no structure `hp`, no inventory/pile
`axes`) round-trips unmarshal→marshal byte-identically — the new fields are
additive `omitempty`, so an old save loads unchanged with no format-version bump.
Together these prove: same seed + same command timeline
over 30k ticks → byte-identical event sequences and equal state hashes; different
seeds diverge; replaying the logged events over genesis (then re-living the quiet
tail) reproduces the live state hash exactly; the day/night cycle behaves (nobody
moves at night).

**Loop-era replay determinism** (`internal/mind/replay_test.go`): a real `Loop` +
`loopMind` pair proves live-vs-replay byte identity above the pure-reducer layer.
`TestLoopRunReplayByteIdenticalSC002` (TASK-52, SC-002) drives cognitions, tool
calls, and a muse through the real loop, then asserts a from-genesis replay
reproduces the identical `State` with the model seam invoked zero times.
`TestJournalAndSituatedReplayByteIdentical` (spec 019 US4, T019, SC-003) extends
this to the grounded-memories feature: injected situated memories (place/why,
place/conv), a journal write→write→delete cognition sequence, and a scripted
over-budget write that the gate refuses (landing nothing but a rejected
`cog.tool_call`) — genesis replay reproduces the identical `State` *and*
byte-identical rendered `soul.md`/`journal.md` over both live and replayed
state, with the model seam invoked exactly once per live cognition and zero
times during replay.

## Connections

Part of the [[testing-strategy]] suite map (split out during the corpus-spec v2
restructure); see that note for the full layered test picture and links to
sibling suites.
