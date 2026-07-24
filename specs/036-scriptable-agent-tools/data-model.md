# Phase 1 Data Model: Scriptable Agent Tools

**Feature**: [spec.md](spec.md) | **Plan**: [plan.md](plan.md) | **Research**: [research.md](research.md)

Entities below are the feature's durable shapes. Wire/authoring details live in
[contracts/bundle-manifest.md](contracts/bundle-manifest.md) and
[contracts/script-api.md](contracts/script-api.md).

## Bundle

A folder under `<worldDir>/bundles/<name>/`. The unit of installation.

| Field | Type | Rules |
|---|---|---|
| `Name` | string | = folder basename; `[a-z0-9_-]{1,64}`; unique within the world (bytewise-first wins, later skipped with boot warning) |
| `Soul` | string (optional) | contents of `SOUL.md`; cap 4000 chars (mirrors `persona.CharterMaxChars`); appended to the metatron system prompt after `charter.md` |
| `Grant` | grant doc (optional) | `capabilities.json`, same schema as the world-level file (`manifestDoc`: `tools`, `miracle_kinds`); **intersection** with the world grant — narrows, never widens |
| `Tools` | []BundleTool | from `tools/<tool>/`, ascending bytewise name order; ≤16 per bundle |

**Lifecycle**: discovered → validated → frozen into `BundleSet` at daemon boot. No hot reload.
Per-tool validation failure skips the tool (loud boot error); invalid `SOUL.md` (over-cap) or
malformed `capabilities.json` rejects the whole bundle (clarification #1). Unknown extra files
are ignored.

## BundleTool

One tool: `tools/<tool>/tool.json` (+ optional `tool.star`). Synthesized into a `tool.Tool`
(`internal/tool/tool.go:108-124`) with `Effect: Expressive`, `Gate: Charge|None`, and
`Events` from the manifest — so the toolloop, schema derivation, and validation reuse existing
machinery unchanged.

| Field | Type | Rules |
|---|---|---|
| `Name` | string | `[a-z0-9_]{1,48}`; collision: built-in (`tool.Lookup`) wins → skip + warn; among bundles, first-loaded wins |
| `Description` | string | 1–500 chars; becomes `PromptGloss` (LLM-facing) |
| `Params` | []ManifestParam | ≤8; maps 1:1 onto `tool.Param` kinds (`agent_name`, `text`, `enum`, `number`) with the same validation rules (`internal/tool/validate.go`) |
| `Events` | []string | non-empty unless narration-only; MUST be ⊆ `injectSocialWhitelist` (boot gate, mirrors `internal/sim/toolcheck.go:62-67`) and MUST equal the union of event types its effects/script can produce |
| `Charges` | int ≥0 | gate minimum (`Cost.Charges`); actual spend stays reducer-authoritative per event type |
| `Effects` | []EffectTemplate | declarative mode; required iff no script |
| `Script` | file ref | `tool.star`; required iff no `Effects`; mutually exclusive with `Effects` in v1 |
| `Limits` | Limits | script mode only: `max_steps` (default 100_000, ceiling 1_000_000) |

## EffectTemplate / Effect

The closed v1 effect vocabulary (research.md R4). An **EffectTemplate** is an effect with
`{args.<param>}` / `{invoker}` placeholders (manifest mode); an **Effect** is the resolved dict
(script mode returns these directly). The compiler (`internal/bundle/effects.go`) is the sole
`store.Event` factory.

| Effect kind | Required fields | Compiles to |
|---|---|---|
| `move_entity` | `target` (agent/structure/pile name or id), `to_x`, `to_y` | `metatron.entity_moved` |
| `remove_entity` | `target` | `metatron.entity_removed` |
| `grant_item` | `target` (agent), `item`, `qty` (1–99) | `metatron.item_granted` |
| `snap_time` | `to_tick` ≥ current tick | `metatron.time_snapped` |
| `narrate` | `text` (≤ 500 bytes), `recipients` (`target` \| `all_living` \| list of names) | one `agent.memory_added` per recipient (`Salience: SalDream`, `Origin: OriginOmen`, per `miracle_batch.go:83-88`) |

**Batch rules**: ≤32 events after expansion; empty batch allowed only when the tool is
narration-only and narration resolved to ≥0 recipients (a fully empty result is a valid no-op
narration tool per spec assumption); NaN/Inf rejected in numeric fields; all-or-nothing landing
via `InjectSocial`'s probe dry run.

## BundleSet

Boot-frozen aggregate the daemon holds; the metatron turn assembly reads it.

| Field | Type | Notes |
|---|---|---|
| `Bundles` | []Bundle | valid ones, load order preserved |
| `Roster()` | []tool.Tool | synthesized tools, deterministic order (bundle order × tool order) |
| `Handlers(deps)` | map[string]toolloop.Handler | one handler per tool: resolve args → run script or expand templates → compile effects → declared-events check → `InjectSocial` |
| `SoulFragments()` | []string | in load order, appended to system prompt |
| `GrantNarrowing()` | grant intersection | applied after the world-level grant |
| `BootReport` | []BootIssue | every skip/rejection: bundle, file, rule violated, message (SC-005) |

## WorldView (script-facing, read-only, invoker-scoped)

Snapshot facade constructed per invocation from `sim.State` — only what the metatron's
existing prompt surfaces legitimately expose (clarification #3). Full API in
[contracts/script-api.md](contracts/script-api.md).

| Member | Exposes |
|---|---|
| `world.tick`, `world.time_of_day` | game clock (tick, derived day-phase string) |
| `world.agents()` | living agents: `name`, `x`, `y`, `alive` |
| `world.agent(name)` | one agent or `None` |
| `world.map_width`, `world.map_height` | map dims |
| `world.rand(purpose, index)` | float64 from `rngAt(seed, "bundle:<tool>:"+purpose, tick, index)` — the only randomness |

Explicitly absent: private memories, beliefs, relationships, journals, orders, LLM state, wall
clock. Adding a member is a deliberate, reviewed API change.

## State transitions

```text
folder on disk ─boot─▶ Validated(BundleTool) ─freeze─▶ BundleSet
                │ fail: tool-level → skipped (BootIssue)
                │ fail: bundle-level (SOUL/capabilities) → bundle rejected (BootIssue)
invocation: args ─validateArgs (toolloop)─▶ script|templates ─▶ effects
        ─compile─▶ []store.Event ─declared⊆manifest check─▶ InjectSocial
        ─probe dry-run─▶ land (all) │ reject (none, Outcome.ResultForModel explains)
```

No new persisted state: bundles are files; landed events are ordinary `store.Event`s (replay
never re-executes tool logic — FR-011 holds by construction).
