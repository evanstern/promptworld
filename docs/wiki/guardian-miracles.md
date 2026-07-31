---
name: guardian-miracles
description: The four charge-priced world-edit events (time snap, item grant, entity move, entity remove) the guardian's other mediated act spends from — overview and terminology ("miracle" name vs "working" player-facing noun) only. Mechanics/cost table [[guardian-miracle-mechanics]], SHIFT/KEEP rebase taxonomy [[guardian-miracle-rebase-taxonomy]], the two landing doors [[guardian-miracle-doors]], targeting digest/replay-determinism [[guardian-miracle-guarantees]].
kind: component
sources:
  - internal/sim/miracles.go
  - cmd/promptworld/work.go
  - internal/ipc/server.go
verified_against: fc1a8314f3f71a33c5e2145c914d5cbb511d9196
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
`work_miracle` tool id, the four `guardian.*` event types, the `miracle` IPC/CLI
command, and this note's own name all keep it, unchanged. The PLAYER-FACING word is
now "working" — the default [[skin]]'s `WorkingNoun()`/`WorkingNounPlural()`
(`"working"`/`"workings"`) resolve wherever the guardian's turn or moment text
names the act (`mt.sk().WorkingNoun()`, [[guardian]]); the canonical CLI verb is
`promptworld work` (below). A custom skin may re-voice the display noun; the tool
id, event vocabulary, and cost/validation mechanics below can never move.

## How it works

The four event types (`guardian.time_snapped`/`item_granted`/`entity_moved`/
`entity_removed`) are dispatched by `applyMiracle` (`miracles.go`); each arm
validates before spending a charge or mutating state — reject-whole, never
clamp — so a rejected miracle spends nothing and a recorded one always
re-applies in replay. The dearest miracle (time snap) costs 2 charges,
every other costs 1, authoritatively from `tool.MiracleCost`; `gratis` is
reachable only through the operator's `--force` door, structurally absent
from the guardian's own tool-call contract. See
[[guardian-miracle-mechanics]] for each kind's validation rule and the
cost/gratis doctrine.

A time snap must preserve every in-flight duration while history stays put
(FR-009): `rebaseTicks` classifies every tick-anchored `int64` state field
as SHIFT (a future deadline or duration anchor, shifted forward with the
jump) or KEEP (a historical timestamp or identity, left alone), and a
build-time test fails when a new such field appears unclassified. See
[[guardian-miracle-rebase-taxonomy]] for the full taxonomy and its
cross-feature reach.

Two thin doors land a miracle through the SAME `BuildMiracleBatch` +
`InjectSocial` path so they can never drift: the guardian's turn
(`work_miracle` -> `landMiracle`, gated by `capabilities.json` and the
curriculum-stage ceiling ahead of the reducer) and the operator CLI/IPC
door (`promptworld work`, `--force` for gratis). Fixed perception-memory
templates attach to each kind at `SalDream`, exactly as memorable as an
omen or vision. See [[guardian-miracle-doors]] for both call paths.

Since spec 059, any turn granted `work_miracle` carries a token-bounded
targeting digest — living villagers' tiles/needs and adjacent passable
tiles — so the guardian can aim a move/remove at a tile the door will
actually accept; the reducer dry-run remains the sole authority on whether
it lands. Every miracle event carries only door-resolved values, so replay
reproduces live application byte-for-byte. See
[[guardian-miracle-guarantees]] for both.

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
This note's summary-style splits — [[guardian-miracle-mechanics]],
[[guardian-miracle-rebase-taxonomy]], [[guardian-miracle-doors]], and
[[guardian-miracle-guarantees]] — hold the mechanics summarized above; each
links back here.

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
