---
name: guardian-miracle-doors
description: The two doors a miracle lands through — the guardian's turn (work_miracle tool call -> landMiracle) and the operator CLI/IPC door (promptworld work, --force for gratis) — both thin translators onto the shared BuildMiracleBatch + InjectSocial path, plus the fixed perception-memory templates each miracle kind attaches. Split from [[guardian-miracles]]; load for either door's call path.
kind: component
sources:
  - internal/guardian/miracle_batch.go
  - internal/guardian/turn.go
  - internal/guardian/toolcalls.go
  - cmd/promptworld/work.go
  - internal/ipc/server.go
verified_against: 72f82f41f7aa2e345572105894cd0fb7c02fc0aa
---

# Guardian's miracle doors

Split from [[guardian-miracles]] (summary-style, corpus-spec v2) — the
perception-memory templates and the two doors that build and land a
miracle batch.

**Perception memories** (`BuildMiracleBatch` in `internal/guardian/miracle_batch.go`):
the shared, door-neutral batch-builder both channels call, so the miracle event and
its perception memories can never drift between the operator and guardian paths. It
only COMPOSES — validation lives entirely in the reducer arms above, enforced by the
`InjectSocial` dry-run, so both doors reject identically and a recorded miracle
always re-applies in replay. `MiracleParams` is the door-neutral, already-resolved
input (villager names resolved to indices, day/`HH:MM` resolved to a tick, by the
caller). Fixed, deterministic memory templates land at `SalDream` — miracles are
exactly as memorable as one of the guardian's omens or visions:

- `time_snap` touches every living villager (`s.LivingAgents()`) with
  `"The light lurched across the sky; a great span of time passed in a single
  breath."`
- `give_item` touches only the granted villager with a rendered
  `"You found N <item> in your hands, as if set there by an unseen giver."`
- `move` touches the moved villager only when `class == "villager"`, resolved via
  `s.VillagerAt(x,y)` — the SAME helper the reducer's `applyEntityMoved` and this
  builder both call, so a tile-addressed move and its memory can never name
  different villagers — with `"An unseen hand lifted you and set you down in a
  strange place."`
- `remove` touches nobody in v1 (no villager is directly affected by a structure/
  pile/terrain removal).

**The two doors**: both are thin translators onto the SAME `BuildMiracleBatch` +
`InjectSocial` path, so they cannot drift. (Spec 036's bundle tools
([[bundle-tools]]) are a third batch producer on the same door: their effect
compiler builds the identical payload structs — including the trailing
perception `agent.memory_added` pattern this note describes — which is what the
dogfood equivalence test pins byte-identical to `BuildMiracleBatch`'s output.)

- **The guardian's turn** (`internal/guardian/turn.go`, `toolcalls.go`): since spec
  017 the turn runs through [[tool-loop]]'s bounded loop ([[guardian]]); "at most
  one mediated act per turn" is now the driver's cardinality rule (one acting call
  lands, every other call this cognition is rejected) rather than a hand-written
  nudge-wins-over-miracle precedence — the model calls `work_miracle` (or one of the
  other acting tools: `send_vision`/`send_omen`/`monitor_and_act`/`cancel_order`/the
  meta tools, spec 029) and whichever lands first ends the turn. Since spec 021 the
  world's
  `capabilities.json` can withhold `work_miracle` entirely or restrict its `kind`
  vocabulary ([[guardian]]): an ungranted tool/kind is structurally absent from the
  declared schema and guidance, its handler is never installed, and `landMiracle`
  additionally refuses via the grant check ("that miracle is not granted in this
  world") — defense in depth ahead of the reducer dry-run, which remains the final
  authority. Since spec 046 ([[curriculum-ladder]]) a staged world's curriculum
  stage caps the grant the same way, upstream of the manifest: the stage-1/-2
  ceiling grants NO miracle kinds at all, so `work_miracle` is structurally
  absent from the guardian's roster until stage-3 opens the full grantable surface
  ([[guardian]]'s `applyStageCeiling` — intersection-only, so a manifest can
  narrow within the ceiling but never exceed it). The operator CLI/IPC door
  below is stage-blind — the ceiling gates the guardian, not the operator. `handleMiracle` parses the call's
  arguments into `miracleArgs` and calls `landMiracle`, which builds a probe
  `sim.State` from the `agentXY`/`alive` snapshot (`mt.agentXY`, mirrored per
  batch by the absorb goroutine in `mirrorState`, so the turn worker never
  races the live replica) BEFORE the kind switch, then resolves `MiracleParams`
  and calls `BuildMiracleBatch` with `gratis=false`. Since spec 091 (door-side
  move-target freshness), a `move` call with `class="villager"` AND a
  `villager` name re-resolves that villager's LIVE position from the SAME probe
  and uses it as the move's source coordinates — the model's surveyed `x`/`y`
  become advisory once a name is supplied, so a move ordered by name cannot
  lose a race to the villager's own walking during model latency. An unknown or
  dead name refuses BEFORE the reducer ever runs (mirroring `landVision`'s "no
  villager named %q"/"%s is beyond reach now" resolution), spending no charge.
  A villager move with NO name, and every structure/pile move, takes the
  untouched coordinate-addressed path — byte-identical to pre-091 behavior,
  including the residual race a bare coordinate address can still lose; when
  that race trips the reducer's "no living villager at (x,y)" refusal, the
  wrapped in-fiction message appends a one-line suggestion to name the
  villager instead. The recorded `guardian.entity_moved` event always carries
  whichever coordinates the door actually used (resolved-by-name or surveyed)
  — emitter-computes, so `applyEntityMoved` ([[guardian-miracle-mechanics]]) is
  unmodified and replay of any previously-recorded move is unaffected. A
  reducer rejection becomes a `rejected_gate` outcome the loop
  feeds back to the model within its round cap (the in-fiction wording is
  unchanged, just no longer necessarily turn-ending), exactly like a refused
  omen or vision; a landed miracle appends a soul-file line and is recorded in the
  transcript with a `✨` prefix.
  Tool-call contract: `work_miracle(kind, day, time, villager, item, qty, class,
  x, y, to_x, to_y)`, no gratis parameter (`internal/tool` registry's
  `miracleParams`, [[tool-registry]]). `TurnResult.Miracle` (`{kind, summary}`) is
  what the console surfaces; every call the loop saw — landed or rejected — also
  lands as a `cog.tool_call` telemetry event ([[event-types]], `toolcalls.go`).
- **The operator CLI/IPC door** (`cmd/promptworld/work.go` → IPC `miracle`
  command — the wire command name is FROZEN, spec 052 ruling 2 —
  → `internal/ipc/server.go`'s `handleMiracle`): `promptworld work
  <world> <snap-time|give|move|remove> ... [--force]` (canonical since spec 052
  FR-008; `promptworld miracle ...` survives as a hidden compat alias — same
  handler, same behavior). `handleMiracle` needs only
  `srv.loop` — never `srv.llm` or `srv.guardian` — so it works on pure-sim worlds
  with no guardian or orchestrator configured. It fetches the current state via
  `loop.DoState` (to resolve door-side name/tile lookups — `give_item`'s villager
  name through `sim.AgentIndexByName`, `time_snap`'s day/`HH:MM` through
  [[game-clock]]'s `clock.ParseTimeOfDay`/`clock.TickAt`), builds `MiracleParams`,
  calls `BuildMiracleBatch`, and lands it through `loop.InjectSocial`. `--force`
  sets `MiracleArgs.Gratis`, the one field that reaches `gratis=true`. Replies with
  `MiracleData{kind, charges, gratis, summary}`.

## Connections

[[guardian-miracles]] is the parent — the reducer validation each door's
build ultimately hits lives in [[guardian-miracle-mechanics]]; a time-snap
door call triggers [[guardian-miracle-rebase-taxonomy]]. [[guardian]] hosts
the turn's side of the guardian door (`hasWorkMiracle`, capabilities/stage
gating ahead of the reducer). [[tool-loop]] is the turn's bounded driver;
[[tool-registry]] declares `work_miracle`
([[guardian-miracle-mechanics]]'s cost table). [[bundle-tools]] is a third
batch producer on the same door, pinned byte-identical to
`BuildMiracleBatch`'s output. [[ipc-protocol]] freezes the `miracle` wire
command name; [[cli-promptworld]] documents `promptworld work` (hidden
`miracle` alias). [[game-clock]]'s `TickAt`/`ParseTimeOfDay` resolve a
time-snap target for both doors.
