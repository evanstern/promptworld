---
name: guardian-miracle-guarantees
description: Two guarantees around a landed miracle — the spec-059 targeting digest that gives the guardian aim (living villagers' tiles/needs plus adjacent passable tiles, prompt surface only) and replay determinism (a miracle event carries only door-resolved values, so Apply re-derives nothing and replay matches live application byte-for-byte). Split from [[guardian-miracles]].
kind: component
sources:
  - internal/guardian/turn.go
  - internal/tool/derive.go
  - internal/sim/miracles.go
verified_against: cffd9a79bbed61ccac573d97c6cf544565b40336
---

# Guardian's miracle guarantees

Split from [[guardian-miracles]] (summary-style, corpus-spec v2) — the
targeting digest that aims a miracle and the determinism proof behind its
replay.

**Miracle targeting digest** (spec 059 US3): world-01 evidence showed 3 of 4
miracle attempts door-rejected on invalid coordinates — the guardian had
authority to act but no aim. Since spec 059, any turn whose granted roster
offers `work_miracle` (gated by `hasWorkMiracle`, `internal/guardian/turn.go`)
carries a token-bounded targeting digest in its user prompt: every living
villager's tile, health/food/warmth, and the passable tiles immediately
adjacent, assembled by `buildTargetingDigest` from the absorb-mirrored
`agentXY`/`agentNeeds` snapshots (never the live replica) and the static
map's own `Passable` — the door stays the authority, this is aim guidance
only. `tool.GuardianTargetingGuidance()` ([[tool-registry]]) supplies the
one-line prose pointer introducing it. Prompt surface only — no new event,
no new door, and the reducer dry-run (`applyEntityMoved`/
`applyEntityRemoved`'s presence/placement checks, above) remains the sole
authority on whether a digest-derived coordinate actually lands.

**Replay determinism**: a miracle event carries only door-resolved, already-decided
values (a tick, an index, a kind, a coordinate) — never a name or a day/HH:MM string
— so `Apply` re-derives nothing at replay time; the same event applied to the same
prior state always produces the same result. `TestMiracleReplayByteIdentity`,
`TestMiracleSnapReplayByteIdentity`, and `TestMiracleGrantReplayByteIdentity`
(`internal/sim/miracles_test.go`) prove each type replays to the same state hash as
live application. `sim.State.m` (the unexported, unserialized static map attached by
`SetMap`/`NewState`/`MigrateState` — [[sim-state-reducer]]) makes the terrain
vocabulary (`passable`/`buildSite`/`effectiveKind`) available identically live, in
the `InjectSocial` dry-run (`probe.SetMap(l.m)` in [[sim-loop]]'s `handleCommand`),
and in replay, so the map-dependent move/remove-terrain checks can never diverge
between the three contexts.

## Connections

[[guardian-miracles]] is the parent — the reducer dry-run this digest
merely aims at, and the `Apply` dispatch this determinism proof covers,
both live in [[guardian-miracle-mechanics]]. [[guardian]] hosts
`hasWorkMiracle` and the turn prompt this digest rides in. [[tool-registry]]
supplies `GuardianTargetingGuidance()`'s one-line prose pointer.
[[sim-state-reducer]] carries the unexported `m *worldmap.Map` field the
map-dependent checks need identically live, in the dry-run, and in replay;
[[sim-loop]] reattaches the static map to the dry-run probe.
[[world-migration]]'s `MigrateState` attaches the map so a migrated state
is miracle-ready like a fresh genesis.
