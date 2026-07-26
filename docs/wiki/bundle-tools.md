---
name: bundle-tools
description: Drop-in persona/tool bundles (spec 036, TASK-85) — manifest or Starlark tools compiled to whitelisted effect batches, boot-frozen from <world>/bundles/, sandboxed and replay-deterministic; effect targets address villagers, structures, piles, and terrain through the spec-082 grammar
kind: component
sources:
  - internal/bundle/manifest.go
  - internal/bundle/effects.go
  - internal/bundle/load.go
  - internal/bundle/validate.go
  - internal/bundle/handlers.go
  - internal/bundle/script.go
  - internal/bundle/worldview.go
  - internal/target/target.go
verified_against: 510a3c3133e120d84cd50525dbc4ee0d3ec01cdc
---

# Bundle tools

`internal/bundle` (spec 036, TASK-85) makes the angel's expressive tools pluggable:
a **bundle** is a folder dropped into `<world>/bundles/` ([[world-save-directory]]'s
`BundlesDir()`) holding tools as `tools/<name>/tool.json` manifests with optional
`tool.star` Starlark scripts, plus optional persona parts (`SOUL.md`,
`capabilities.json`). The core invariant is **effects, not events**: a tool —
declarative or scripted — produces *effect dicts* from a closed five-kind
vocabulary (`move_entity`, `remove_entity`, `grant_item`, `snap_time`, `narrate`),
and the compiler in `effects.go` is the sole `store.Event` factory. A script has
no API for constructing an event, so the [[sim-loop]] `InjectSocial` whitelist
stays the isolation boundary by construction, and tick-simulated world verbs
(hunt/forage/build) remain native Go — v1 scriptability is expressive-only.

## How it works

**Lifecycle**: `Discover(worldDir)` runs once at daemon boot
([[daemon-lifecycle]]), scanning direct child dirs in ascending bytewise order
(the `skills/` discipline — deterministic, dotfiles skipped, unknown files
ignored) and freezing a `BundleSet`. Validation is a ladder: bundle-level rules
(name shape; `SOUL.md` ≤4,000 chars; `capabilities.json` parses; ≤16 tools)
reject the whole bundle, tool-level rules (strict manifest decode; param rules
mirroring `tool.Validate`; declared `events` ⊆ the whitelist via
`sim.InjectableSocialEvent`; script compiles and defines `apply`; step-cap
bounds) skip just that tool — every skip/rejection is a `BootReport` entry
naming file, rule id, and offending value, logged at boot. Name collisions:
built-ins (`tool.Lookup`) always win; among bundles, first-loaded wins — both
warn. Editing a bundle needs a world restart; there is no hot reload.

**Manifest → tool**: `manifest.go` synthesizes a real `tool.Tool`
(Effect Expressive, `PromptGloss` from the description, `Cost.Charges` as the
gate minimum — the reducer's per-event price table stays authoritative,
[[guardian-miracles]]), so schema derivation, arg validation, and guidance prose
ride the existing [[tool-registry]] machinery; the registry itself is never
touched — the turn assembly appends `BundleSet.Roster()` and `Handlers()` after
`grantedRoster` ([[guardian]]).

**Invocation**: the handler (per tool, built by `handlers.go` over an
`InvocationContext{State, Tick, Invoker, Seed, Inject}`) resolves declarative
`{args.x}`/`{invoker}` templates or calls the script's `apply(args, world)`,
compiles effects, asserts every produced event type is declared in the manifest,
and lands the batch through `InjectSocial` — the existing atomic probe-copy dry
run, so batches land all-or-nothing and a failed invocation spends no charge.
Failures come back as `rejected_gate` outcomes the model can retry within
[[tool-loop]]'s round cap, never infrastructure errors.

**Target addressing** (spec 082, TASK-97): effect `target` strings parse
through `internal/target` — the stdlib-only leaf package that is the ONE
authority on the address grammar (normative:
`specs/082-target-addressing/data-model.md`; manifest surface:
`specs/036-scriptable-agent-tools/contracts/bundle-manifest.md`). Forms: bare
villager name (v1 compat, byte-identical), `villager:<name>`, and
`<class>@X,Y` for villager/structure/pile/terrain; the reserved-prefix rule
(`^(villager|structure|pile|terrain)[@:]`) makes a malformed structured
address a compile error, never a name fallback. The per-kind matrix:
`move_entity` takes villagers (name or tile via `sim.State.VillagerAt` —
first-living-by-index, the miracle door's own choice), structures, and piles;
`remove_entity` takes structures, piles, and terrain and rejects every
villager form compiler-side (mirroring the reducer's never-remove-a-villager
doctrine, which stays authoritative); `grant_item` takes villagers only.
Resolution reads only `CompileInput.State` — roster walk, `VillagerAt`, the
one-per-tile `HasStructureAt`/`HasPileAt` probes, `MapDims` bounds — and
compiled payloads are byte-identical to `BuildMiracleBatch`'s for the same
class+tiles. Every target failure (syntax/class/form/bounds/unresolved) is a
descriptive whole-invocation rejection naming the effect index, field, and
offending address. Rect (`class@X1,Y1..X2,Y2`) and axis-aligned line
(`class@X1,Y1->X2,Y2`) forms parse in the shared package — with normalized
corners, preserved endpoint order, and deterministic `Tiles()` enumeration —
but bundles reject them as reserved for the TASK-157 designation consumers
(settlement zone, structure site, wall line), which will import the same leaf
from `internal/tool`. Live turns resolve against the guardian's probe, which
mirrors structure/pile tiles per absorb batch (the `agentXY` discipline) and
carries the static map.

**Sandbox** (`script.go`): programs compile once at boot and run per-invocation
on a fresh step-capped `starlark.Thread` (manifest `limits.max_steps`, default
100k, hard ceiling 1M) — no `load()`, no clock, no I/O, recursion off, `print`
to the daemon log only; the script→effect conversion rejects floats/NaN/Inf,
unknown fields, and malformed shapes with element-indexed errors.

**World view** (`worldview.go`): scripts see an invoker-scoped, frozen snapshot
— `tick`, `time_of_day` (night pinned to the sim's own 22:00–06:00 definition),
map dims, living agents (name/x/y/alive), `agent(name)` — and draw randomness
ONLY from `world.rand(purpose, index)`, backed by `sim.BundleRand` over the
[[deterministic-rng]] pattern with a `"bundle:<tool>:<purpose>"` namespace.
Private memories, beliefs, relationships, and wall time are deliberately absent.

**Personas**: a bundle's `SOUL.md` fragments stack into the turn prompt after
the charter and before skills (the fixed frame still lands last —
[[guardian-orders]]); its `capabilities.json` NARROWS the world grant by
intersection (`intersectGrant` — a persona can exclude tools/miracle kinds,
never resurrect what the world owner excluded).

**Determinism**: replay never re-executes tool logic — landed events are
self-contained data, proven by the replay byte-identity tests (including
deleting the bundle dir before replay — the spec-082 addressing fixture
repeats the proof for structure/pile/terrain edits). Same (args, state, seed,
tick) ⇒ byte-identical batches; the dogfood bundle
(`examples/bundles/dogfood-move`) pins its output byte-identical to
`BuildMiracleBatch`'s `move` kind, and `TestAddressingMiracleByteIdentity`
extends that pin to every class+tile form.

## Connections

[[daemon-lifecycle]] discovers/freezes the set at boot and hands it to
[[guardian]] via `SetBundles`; [[tool-registry]] supplies the `Tool` model,
schema derivation, and the spec-036 `PromptGloss` guidance fallback;
[[sim-loop]]'s `InjectSocial` door + whitelist (readable via
`InjectableSocialEvent`) is the only path to state; [[guardian-miracles]]
defines the payloads the effect compiler reproduces; [[deterministic-rng]]
backs `world.rand`; [[world-save-directory]] owns the `bundles/` layout;
[[event-types]] catalogs every type a bundle may declare.

## Operational notes

Authoring guide: `docs/bundles.md` (manifest/script reference, target
addresses, validation errors, examples). Specs:
`specs/036-scriptable-agent-tools/` (manifest contract, normative) and
`specs/082-target-addressing/` (address grammar, normative). The former v1
limitation (TASK-97) is resolved: structure/pile/terrain addressing landed
with spec 082, so `remove_entity` is real — a bundle tool moves and removes
structures and piles and clears terrain tiles with miracle-door-identical
payloads. Boot cost is negligible: 32 bundles / 256 tools discover+validate
in ~50 ms.
