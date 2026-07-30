---
name: guardian-miracle-guarantees
description: Two guarantees around a landed miracle — the spec-059 targeting digest that gives the guardian aim (living villagers' tiles/needs/carry headroom plus adjacent passable tiles, prompt surface only) and replay determinism (a miracle event carries only door-resolved values, so Apply re-derives nothing and replay matches live application byte-for-byte). Split from [[guardian-miracles]].
kind: component
sources:
  - internal/guardian/turn.go
  - internal/tool/derive.go
  - internal/sim/miracles.go
verified_against: cf65debb44c1e17b54c0f3421d11e1e8cc28576c
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
villager's tile, health/food/warmth, live carry headroom, and the passable
tiles immediately adjacent, assembled by `buildTargetingDigest` from the
absorb-mirrored `agentXY`/`agentNeeds` snapshots (never the live replica) and
the static map's own `Passable` — the door stays the authority, this is aim
guidance only. `tool.GuardianTargetingGuidance()` ([[tool-registry]]) supplies
the one-line prose pointer introducing it. Prompt surface only — no new event,
no new door, and the reducer dry-run (`applyEntityMoved`/
`applyEntityRemoved`'s presence/placement checks, above) remains the sole
authority on whether a digest-derived coordinate actually lands.

Carry headroom (spec 095 FR-001, TASK-167) rides the SAME per-villager line
and the SAME `needMirror` snapshot as health/food/warmth: `needMirror` grew a
`Bulk` field (`internal/guardian/orders.go`), refreshed in `mirrorState`
(`internal/guardian/guardian.go`) from `sim.Bulk(inv)` — the same exported
derived-value accessor [[executor-world-state]] documents — never a second
copy of the reducer's own `bulk()` arithmetic. `buildTargetingDigest` renders
it against `sim.BulkCap` as free units (`"carrying U/C, F free"`); dead
villagers carry no line at all, so they carry no headroom either. The
give_item gloss (`derive.go`'s `miracleKindArgs["give_item"]`,
[[tool-registry]]) points the model at this field and restates the door's own
reject-whole rule (FR-011) so a grant lands within the cap on the first
attempt instead of bouncing off `applyItemGranted`
([[guardian-miracle-mechanics]], unmodified — the door's rejection message is
pinned byte-unchanged by `TestMiracleGrantOverCapWholeReject`,
`internal/sim/miracles_test.go`).

**Replay determinism**: a miracle event carries only door-resolved, already-decided
values (a tick, an index, a kind, a coordinate) — never a name or a day/HH:MM string
— so `Apply` re-derives nothing at replay time; the same event applied to the same
prior state always produces the same result. `TestMiracleReplayByteIdentity`,
`TestMiracleSnapReplayByteIdentity`, and `TestMiracleGrantReplayByteIdentity`
(`internal/sim/miracles_test.go`) prove each type replays to the same state hash as
live application. Spec 091's door-side move-target freshness (villager moves
naming a target now resolve that villager's LIVE position at the door,
[[guardian-miracle-doors]]) is exactly this guarantee's doctrine at work: the
resolved coordinate — never the name — is what the recorded
`guardian.entity_moved` carries, so `applyEntityMoved` needed no change and
every previously-recorded move (coordinate-addressed, as all pre-091 moves
were) replays identically; `TestMoveFreshnessReplayByteIdentical`
(`internal/sim/miracles_test.go`) pins it. `sim.State.m` (the unexported, unserialized static map attached by
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
