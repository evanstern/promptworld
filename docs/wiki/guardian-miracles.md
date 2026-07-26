---
name: guardian-miracles
description: The four charge-priced world-edit events (time snap, item grant, entity move, entity remove) — cost table, operator-only gratis doctrine, shift-semantics re-base taxonomy, perception memories, and the two doors; package renamed internal/metatron → internal/guardian (spec 052/TASK-121); the frozen "miracle" mechanics (tool id, event types, IPC/CLI command name) now display to the player as "working" via the boot-frozen [[skin]]'s WorkingNoun, canonical CLI verb `promptworld work` (hidden `miracle` alias)
kind: component
sources:
  - internal/sim/miracles.go
  - internal/guardian/miracle_batch.go
  - internal/guardian/turn.go
  - internal/guardian/toolcalls.go
  - internal/tool/registry.go
  - internal/tool/derive.go
  - internal/ipc/server.go
  - cmd/promptworld/work.go
verified_against: d304e8adb64fdf40e24bfeca3ca3420e8a840a35
---



# Guardian's miracles

Miracles (spec 016) are four direct, charge-priced world edits — spent from the same
bank as a [[guardian]] omen or vision, but landing a concrete change rather than a
villager's subjective experience. Like an influence, a miracle lands through
`Loop.InjectSocial` as
one atomic, whitelisted batch; the reducer validates rather than clamps, so an
invalid miracle is rejected wholesale before recording and a recorded miracle always
re-applies cleanly in replay (spec 016 R1). No new persistent entities exist —
miracles only mutate fields already in `sim.State`.

Terminology (spec 052, TASK-121): "miracle" is the frozen mechanics name — the
`work_miracle` tool id, the four `metatron.*` event types, the `miracle` IPC/CLI
command, and this note's own name all keep it, unchanged. The PLAYER-FACING word is
now "working" — the default [[skin]]'s `WorkingNoun()`/`WorkingNounPlural()`
(`"working"`/`"workings"`) resolve wherever the guardian's turn or moment text
names the act (`mt.sk().WorkingNoun()`, [[guardian]]); the canonical CLI verb is
`promptworld work` (below). A custom skin may re-voice the display noun; the tool
id, event vocabulary, and cost/validation mechanics below can never move.

## How it works

**The four event types** (`internal/sim/miracles.go`, canonical JSON, struct-ordered):

| Event | Payload | Effect |
|---|---|---|
| `metatron.time_snapped` | `TimeSnappedPayload{to_tick, gratis}` | jumps `State.Tick` forward to `to_tick`, forward-only (a target at or before the current tick is rejected whole, before any spend); shifts every relative-duration field via `rebaseTicks` first |
| `metatron.item_granted` | `ItemGrantedPayload{agent, kind, qty, gratis}` | provisions a living villager with `qty` known items, reject-whole (never clamp) if it would exceed the carry cap |
| `metatron.entity_moved` | `EntityMovedPayload{class, x, y, to_x, to_y, gratis}` (`class` ∈ villager\|structure\|pile) | relocates the entity from `(x,y)` to `(to_x,to_y)` |
| `metatron.entity_removed` | `EntityRemovedPayload{class, x, y, gratis}` (`class` ∈ structure\|pile\|terrain; villager is always rejected) | deletes the entity or overlays the terrain |

`applyMiracle` in `miracles.go` is the reducer dispatcher `sim.State.Apply` routes
these four types to (alongside `applyMetatron` for `metatron.charge_regenerated`/
`metatron.nudged` — [[sim-state-reducer]]). Every arm's validation — presence at the
source, the destination's placement rule, item kind/quantity — precedes both the
charge spend and the mutation, so a rejected miracle spends nothing and leaves no
partial application (validate-not-clamp, reject-whole):

- **`applyEntityMoved`**: `villager`/`pile` destinations must be `passable`;
  `structure` destinations must satisfy `buildSite`. A moved structure carries its
  `FuelUntil`/`Owner`/`Store` along whole; a moved pile merges onto any pile already
  at the destination (`movePile`); a moved villager drops its intent and goes idle
  at the landing tick (cancel-and-replan) — villagers may share a tile, so no
  destination-exclusivity check applies to a villager move. Since spec 041
  ([[mental-maps]]), a teleported villager also gets the SAME derived
  mental-map bookkeeping a walked step gets: its landing surroundings mark
  explored and mutual peer sightings with anyone nearby update — a miracle
  move is knowledge-transparent, not a blind teleport.
- **`applyEntityRemoved`**: a villager is always rejected ("a villager can never be
  removed" — v1 doctrine). A removed chest first spills its `Store` to a ground pile
  via `spillInventory` (the same death-spill vocabulary `agent.died` uses) before
  deletion, so goods are never silently destroyed; a removed pile is destroyed with
  its contents (the explicit, operator-visible destruction the miracle names).
  `removeTerrain` overlays a tree/forage/rock tile through the SAME vocabulary the
  executor's own harvest completions use (chop→`Cleared`, forage→`Harvested` with a
  regrow deadline, quarry→`Quarried`, permanent) — a removed tile is a state the
  executor could already have produced on its own; spec 068's marsh/sand ground
  covers are deliberately absent from this switch — they have no depleted state the
  executor could ever produce, so they fall to the same "holds no removable
  terrain" refusal grass and water already draw ([[worldmap-generation]],
  [[tile-registry]]); an already-overlaid tile is
  rejected as a no-op target.
- **`applyItemGranted`**: validates a living, in-range agent index, a `grantableKind`
  (the `Inventory` key vocabulary plus `"spear"`/`"axe"` singular), and a positive
  quantity. One bulk per granted unit, exactly like a carried item — a grant of
  `qty` items always costs `qty` bulk regardless of kind, so the cap check is
  `bulk(*inv)+qty > bulkCap`. A spear grant appends `qty` fresh `spearDurability`
  entries to `Inv.Spears`, kept sorted ascending (hunts spend the most-worn first);
  since spec 032 (US2) an axe grant is the same clone against `Inv.Axes` with
  the same fresh-`axeDurability` value the `craft_axe` verb produces, sorted
  the same way.
- **`applyTimeSnapped`**: rejects a non-forward target before any spend or mutation;
  spends 2 charges (the dearest miracle) unless gratis; calls `rebaseTicks`, then
  sets `State.Tick = to_tick`. FR-010 (a snap mints no charges across the skipped
  regeneration boundaries) needs no code of its own — regeneration only fires when
  the executor *processes* a boundary crossing, and a snap processes no interval.

**Cost table and gratis doctrine**: the time snap costs 2 charges; every other
miracle costs 1. Since spec 021 (TASK-64) the AUTHORITATIVE per-kind table lives in
the leaf [[tool-registry]] (`tool.MiracleCost(kind)` / `tool.MiracleCostsByEvent()`,
`internal/tool/registry.go`, beside `miracleKinds`); `sim.miracleCost` (`miracles.go`,
a keyed map — never iterated into state, for determinism) is now DERIVED from
`tool.MiracleCostsByEvent()` rather than a second literal, and the guardian's turn
prompt renders costs from the same source (`tool.GuardianToolGuidance`), so one edit
propagates to enforcement and every rendering (`TestMiracleCostDerivedFromTool`
pins the derivation). Pricing remains doctrine, not caller input — a payload never
carries its own price, so replay re-validates every spend identically (R2).
`spendMiracleCharge(eventType, gratis)` is the shared validate/spend helper every
arm calls last, after all other validation passes: with `gratis` it returns
immediately, waiving ONLY the charge (every other validation still runs in full);
without it, it errors if the bank can't pay and decrements it otherwise. `gratis` is
reachable from exactly one surface: the `promptworld work --force` CLI/IPC door
(canonical since spec 052 FR-008; `promptworld miracle --force` survives as a
hidden compat alias, same handler)
([[cli-promptworld]], [[ipc-protocol]], [[ipc-server]]) — the operator's cheat door
the guardian structurally cannot reach. The guardian's turn contract — since spec 017 the
`work_miracle` tool call, parsed into `miracleArgs` (`internal/guardian/toolcalls.go`;
the retired `turnReply.Miracle` anonymous struct carried the identical flat field
set pre-loop) — has **no gratis field at all** — a model can emit `"gratis":true`
in its tool-call arguments and it is simply dropped at unmarshal, nothing to
sanitize or forget. `landMiracle` (the guardian's landing path) calls the shared
builder with `gratis` hardcoded `false`, so a model-driven miracle
is unconditionally charged (contracts §1, FR-007/SC-005).

**Shift-semantics re-base taxonomy** (`rebaseTicks` in `miracles.go`): the SINGLE
authority for how a time snap preserves in-flight durations while history stays
put (FR-009). Every tick-anchored `int64` field anywhere in the state tree MUST be
classified SHIFT or KEEP in its doc comment:

- **SHIFT** (`+delta`) — a future deadline, or an anchor from which an elapsed/
  remaining duration is measured (shifting preserves that duration across the
  jump). A SHIFT field whose zero value means "unset/never" is shifted only when
  non-zero. SHIFT fields: `Agent.IdleSince` (shifted unconditionally — its zero is
  genesis-idle, a real tick, not a "never" sentinel), `Agent.LastTalk`/`LastGive`,
  `Intent.WorkStart`, `AgentHail.Until`, `PlanStep.Until`, `Guard.Tick`,
  `Structure.FuelUntil`, `Harvest.Regrow`, `DenUse.Ready`, `FoodBatch.SpoilAt`,
  `Debt.Due`, `Belief.Reinforced` (spec 030: the decay-curve anchor, elapsed =
  tick − Reinforced; shifted only when non-zero — a legacy grandfathered belief
  stays at 0 so it never decays), `Gru.LastAttack`, `Meeting.OpenedTick`,
  `Meeting.GatherStart`, and (spec 029) `GuardianOrder.ExpiresTick` — shifted ONLY
  for ACTIVE orders, so a standing order's remaining lifetime survives the jump (a
  consumed order's deadline is a spent artifact, left put). Spec 041
  ([[mental-maps]]) adds `PlaceFact.Seen` and `PeerSighting.Seen`, the mental
  map's freshness anchors (fresh iff `now − Seen < horizon`, the
  `Belief.Reinforced` shape) — shifted unconditionally when non-zero, since a
  snap would otherwise instantly stale every villager's spatial knowledge;
  `applyEntityMoved`'s villager case (below) shares the same derived
  bookkeeping a walked step gets. Spec 043 adds `Agent.NeedsAnchorTick`, the
  need-trajectory window's edge anchor (elapsed = tick − NeedsAnchorTick gates
  the anchor roll in the `agent.needs_changed` arm; 0 = unset sentinel, stays
  0) — shifted so a snap preserves the window's remaining time instead of
  forcing an immediate anchor reset that would wipe every villager's
  trajectory sense; `NeedsAnchor` itself holds need levels, not ticks, so it
  needs no entry. Spec 061 ([[social-fabric]], [[sim-state-reducer]]) adds
  `PairTalk.Tick`, the conversation loop damper's per-pair last-exchange
  anchor (cooldown elapsed = tick − Tick, the `Agent.LastTalk` shape) —
  shifted UNCONDITIONALLY: unlike most SHIFT fields here, a PRESENT
  `PairTalk` record is always a real exchange tick (absence of the record
  itself, not a zero value, means "never talked"), so there is no zero
  sentinel to guard. Spec 062 ([[reflex-policy]], [[sim-state-reducer]]) adds
  `Agent.LastMindIntentDone`, the reflex PREP gate's yield-window anchor
  (elapsed = tick − LastMindIntentDone gates `prepYields`; 0 = never
  mind-driven, the permanent sentinel every no-planner world's agents carry)
  — shifted only when non-zero, the `Belief.Reinforced`/`NeedsAnchorTick`
  shape: a snap must preserve the window's remaining deference rather than
  spuriously arming or clearing it.
- **KEEP** — a historical timestamp or an identity/counter; rewriting it would
  rewrite history or break a reference. `Agent.Generation`, `Agent.LastGoalTick`,
  `Memory.Tick`, `Memory.Conv` (spec 019: a conversation-ref identity, the same
  founding-talk tick as `ConvoRecord.Conv` — an identity, not a duration anchor),
  `Memory.Seq` (spec 042: the emitting event's store seq — an identity, never a
  clock value), `Agent.SitVecTick` (spec 042: when the agent's situation text
  was rendered — history/audit, the `Memory.Tick` shape), `JournalEntry.Tick`
  (spec 019: when the entry was written, a historical
  timestamp), `Belief.Tick`, `ChronicleEntry.Tick`/`Day`/`FromTick`/`ToTick`,
  `GuardianOrder.PlacedTick` (spec 029: when the order was placed, history),
  `IntentRecord.Tick`/`IntentRecord.OutcomeTick` (spec 043: when an intent
  landed / when its outcome landed — the recent-intent ring is a historical
  self-history log, the `Memory.Tick` shape, never a future deadline),
  `PlaceFact.Detail` (spec 041: a remembered value baked at emission, never
  re-derived — for a fire it mirrors the FuelUntil last seen, so shifting it
  would rewrite what the agent remembers rather than what is; the perception
  sweep simply re-witnesses the shifted reality on the next look),
  `GuardianReportCard.Tick`/`Seq`/`Citations` (spec 063,
  [[grounded-feedback]]: when the card landed, the card event's own identity,
  and the cited event seqs — history and identities, never deadlines), and
  every
  other identity/history field — see the doc comment for the full list.
  `TestRebaseTaxonomyComplete` caught both spec-019 additions, the spec-030
  `Belief.Reinforced` field, (later) spec 029's `GuardianOrder.ExpiresTick`/
  `PlacedTick`, spec 041's `PlaceFact`/`PeerSighting` fields, spec 042's
  `Memory.Seq`/`Agent.SitVecTick` fields, spec 043's `IntentRecord.Tick`/
  `OutcomeTick` (KEEP) and `Agent.NeedsAnchorTick` (SHIFT), spec 061's
  `PairTalk.Tick` (SHIFT), spec 062's `Agent.LastMindIntentDone` (SHIFT,
  only-non-zero), and spec 063's `GuardianReportCard.Tick`/`Seq`/`Citations`
  (KEEP) as new tick-anchored
  `int64` fields requiring classification, confirming the taxonomy guard holds
  across features outside miracles' own spec.

`TestRebaseTaxonomyComplete` (`internal/sim/miracles_test.go`) is the taxonomy guard:
it fails the build when a new tick-anchored `int64` field appears in the state
structs without a classification entry here, so the taxonomy can never silently
drift from the struct definitions. `PlanStep.Until` and `Guard.Tick` are shifted even
though `specs/016-metatron-miracles/data-model.md` did not list them — a deviation
recorded in `rebaseTicks`'s doc comment: both are genuine future deadlines FR-009's
catch-all ("any future duration-anchored state") requires shifting, since leaving
them unshifted would expire a pending plan step or fire a timed guard the instant a
snap jumped past its absolute tick.

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
  arguments into `miracleArgs` and calls `landMiracle`, which resolves
  `MiracleParams` from an `agentXY` snapshot (`mt.agentXY`, mirrored per batch by
  the absorb goroutine in `mirrorState`, so the turn worker never races the live
  replica), builds a probe `sim.State` from it, and calls `BuildMiracleBatch` with
  `gratis=false`. A reducer rejection becomes a `rejected_gate` outcome the loop
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

[[guardian]] hosts the guardian's door (`landMiracle`, the `work_miracle` tool call
parsed into `miracleArgs`); [[guardian-orders]] shares this note's `rebaseTicks`
taxonomy (a standing order's `ExpiresTick` is a SHIFT field);
[[mental-maps]] shares it too (`PlaceFact.Seen`/`PeerSighting.Seen` SHIFT,
`PlaceFact.Detail` KEEP) and is what a miracle-moved villager's derived
explored/sighting bookkeeping updates; [[social-fabric]]/[[sim-state-reducer]]
share it since spec 061 (`PairTalk.Tick` SHIFT, the conversation loop
damper's per-pair ledger);
[[sim-loop]] whitelists the four event types in `injectSocialWhitelist` and
reattaches the static map to the dry-run probe; [[sim-state-reducer]] dispatches to
`applyMiracle` and carries the unexported `m *worldmap.Map` field the reducer arms
need; [[event-types]] catalogs the four payload shapes; [[ipc-protocol]] and
[[ipc-server]] define and implement the `miracle` wire command; [[cli-promptworld]]
is the `promptworld work` operator door (hidden `miracle` alias); [[game-clock]]'s `TickAt`/
`ParseTimeOfDay` resolve a time-snap target; [[world-migration]]'s `MigrateState`
attaches the map so a migrated state is miracle-ready like a fresh genesis.
[[tool-loop]] is the guardian's door since spec 017: `work_miracle` is a declared
loop tool ([[tool-registry]]'s `LoopRosterGuardian`) whose handler
(`toolcalls.go`) wraps `landMiracle` exactly as described above.
[[guardian-orders]] documents the spec 059 survival watches whose turn frame
permits a miracle on the guardian's own initiative (charge cost unchanged) —
this note's cost/rebase/door mechanics are identical either way, only the
turn's authorization frame differs; [[mental-maps]] is the closed
place-fact/passability vocabulary the targeting digest draws its adjacency
guidance from. [[grounded-feedback]] (spec 063) shares this note's
`GuardianReportCard.Tick`/`Seq`/`Citations` rebaseTicks classification, and
its `explain` tool's `costs`/`workings` fact sheets read the SAME
`tool.MiracleCost`/`miracleKindArgs` source this note's cost table and
guardian prompt do — a described price can never disagree across all three
surfaces.

## Operational notes

A miracle never mints a new persistent entity — it edits fields already in
`sim.State`. On an ENDED world (spec 044, [[morgue]]) no miracle can land
through either door: `InjectSocial` narrows to recorded prose about the
finished run, so all four miracle event types are refused at the command
gate in [[sim-loop]]. A villager is the one class that can never be removed by any door
(v1 doctrine); this is enforced in the reducer, not just at the doors, so it holds
even against a forged event. The gratis flag's only reachable path is the CLI/IPC
`--force` flag — if a future surface needs to grant it, that is a deliberate design
decision to record, not a default to fall into.
