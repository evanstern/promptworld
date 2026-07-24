# Bundle tools — authoring guide

How to add new agent-callable tools to a world without touching Go code, as of
spec 036 (scriptable agent tools, TASK-85). Formal shapes live in
`specs/036-scriptable-agent-tools/contracts/` (`bundle-manifest.md`,
`script-api.md`, `boot-validation.md`); this is the practical, authoring-facing
distillation.

## What a bundle is

A **bundle** is a folder of tools (and, optionally, a persona identity) dropped
into a world's save directory. The daemon discovers and validates every bundle
once at boot, alongside the built-in tool registry — a bundle can never be
loaded, edited, or revalidated while the daemon is running. Bundles are purely
**additive**: a world with no `bundles/` directory boots exactly as it always
has, and a bundle that fails validation never brings the world down — only the
broken piece is skipped.

Two things a bundle tool is emphatically *not*:

- **Not an event.** A tool never writes to the event log directly; it produces
  a batch of declared **effects**, which the engine compiles into the exact
  same `store.Event` types the built-in Metatron miracles use, then lands them
  through the same atomic, dry-run-probed path (`InjectSocial`).
- **Not live-reloadable.** The set of bundles is frozen at boot
  (`bundle.BundleSet`). Editing or deleting `bundles/` on disk does nothing
  until the next restart — and, deliberately, replay never re-reads or
  re-executes bundle code at all (see [Determinism](#determinism-guarantees)).

## Folder layout

```
<worldDir>/bundles/
  <bundle-name>/
    SOUL.md                 # optional — persona identity fragment
    capabilities.json        # optional — persona grant narrowing
    tools/
      <tool-name>/
        tool.json             # required — manifest (always present)
        tool.star             # required IFF the manifest declares "script"
```

- `<bundle-name>` matches `[a-z0-9_-]{1,64}` and is just a folder — its name
  has no effect on tool names or the roster.
- `<tool-name>` matches `[a-z0-9_]{1,48}` **and must equal the manifest's
  `"name"` field** — a mismatch skips the tool (rule T1).
- A bundle may ship a `SOUL.md` and/or `capabilities.json` with no `tools/` at
  all (a pure persona overlay), or `tools/` with neither (a pure tool pack).
  Both shapes compose in the same bundle.
- Up to 16 tool directories per bundle (a cap breach rejects the *whole*
  bundle — a bundle this large is presumably a bug, not a legitimate need).
- Any other file (`README.md`, a license, notes) is ignored — bundles are
  allowed to carry documentation alongside the machine-read files.

Discovery order is deterministic: bundles load in ascending bytewise folder
name order, and tools within a bundle load the same way. This ordering is what
makes `C2` (see [Collision rules](#collision-rules)) predictable and is part
of why replaying a world always reconstructs the same roster.

## `tool.json` reference

One manifest per tool, strict-decoded UTF-8 JSON — **unknown top-level keys
are rejected**, so a typo in a field name fails loudly at boot instead of
silently doing nothing.

```jsonc
{
  "name": "teleport",                // REQUIRED, [a-z0-9_]{1,48}, == folder name
  "description": "Whisk a villager across the map in a puff of smoke",
  "params": [ /* see below, ≤8 */ ],
  "events": ["metatron.entity_moved", "agent.memory_added"],
  "charges": 1,                      // OPTIONAL, default 0 — gate MINIMUM, not the spend
  "effects": [ /* declarative, see below */ ],   // exactly one of effects | script
  "script": "tool.star",             // relative filename; must exist and parse
  "limits": { "max_steps": 100000 }  // script only; default 100000, ceiling 1000000
}
```

### Params (≤8 per tool)

| Field | Applies to | Meaning |
|---|---|---|
| `name` | all | `[a-z0-9_]{1,32}`, unique within the tool |
| `kind` | all | one of `agent_name`, `text`, `enum`, `number` |
| `required` | all | default `false` |
| `description` | all | optional, ≤200 chars — shown to the model |
| `enum` | `enum` only | REQUIRED, 1–16 non-empty strings |
| `min`, `max` | `number` only | optional bounds; `0, 0` means unbounded; an inverted non-zero pair (`min > max`) is rejected |
| `max_bytes` | `text` only | optional byte cap, default 500 |

The four param kinds mirror the same argument kinds the built-in tool registry
uses, so a bundle tool's schema is derived and presented to the model exactly
like a built-in tool's.

### Effect kinds

Every effect compiles to exactly one event type (except `narrate`, which
expands to one `agent.memory_added` per recipient). This is the **entire**
closed vocabulary — a tool cannot produce any event type outside it:

| Kind | Produces | Required fields | Notes |
|---|---|---|---|
| `move_entity` | `metatron.entity_moved` | `target`, `to_x`, `to_y` | target must be a living villager |
| `remove_entity` | `metatron.entity_removed` | `target` | |
| `grant_item` | `metatron.item_granted` | `target`, `item`, `qty` | `qty` must be 1–99 |
| `snap_time` | `metatron.time_snapped` | `to_tick` | |
| `narrate` | `agent.memory_added` (×N) | `text`, `recipients` | see below |

`recipients` is either the string `"all_living"`, the string `"target"`
(resolves to the invocation's `target` argument — the tool must declare a
param named `target` to use this form), or a JSON array of specific villager
names.

Caps enforced by the compiler, regardless of declarative or scripted origin:
- ≤32 events per batch, after `narrate` recipient expansion.
- `narrate.text` ≤500 bytes.
- Numeric fields must resolve to finite integers in range — no floats, no
  `NaN`/`Inf` can ever reach a payload.
- Every produced event type must be a member of the tool's declared `events`
  — this is checked **again per invocation**, not just at boot, so a script
  can never smuggle an event type past what it promised.

### `charges` and `limits`

- `charges` (≥0, default 0) is the **gate minimum** the tool requires — it
  does not set the price. The reducer remains the sole authority on what a
  landed event actually costs; the manifest cannot waive or override that
  (bundle-produced events always set `Gratis: false`).
- `limits.max_steps` only applies to script-mode tools: the Starlark
  interpreter step budget for one invocation, `(0, 1_000_000]`, defaulting to
  100,000 when the block is absent.

### Declarative `effects` — templating

String fields support exactly two placeholder forms — `{args.<param>}` and
`{invoker}` — substituted verbatim, with **no expressions and no nesting**:

```jsonc
"effects": [
  { "kind": "move_entity", "target": "{args.target}",
    "to_x": "{args.x}", "to_y": "{args.y}" },
  { "kind": "narrate", "text": "{args.target} vanished in a poof of smoke",
    "recipients": "all_living" }
]
```

(This is the real teleport fixture —
`internal/bundle/testdata/worlds/declarative/bundles/demo/tools/teleport/tool.json`.)

Boot validation additionally proves, for declarative tools, that the set of
event types the `effects` array can produce is **exactly equal** to the
declared `events` list — not a subset, not a superset (rule T5). Get one
wrong and the tool is skipped with a message naming the mismatch.

## `tool.star` reference — scripted tools

A script-mode tool swaps `effects` for `script`, and its `tool.json` gains no
extra required fields beyond the shared ones above. The file must define
`apply`:

```python
def apply(args, world):
    """Pure function: (args, world) -> list of effect dicts."""
    if world.time_of_day == "night":
        return [
            {"kind": "narrate", "text": "A soft light blooms over " + args["target"] + ".",
             "recipients": "all_living"},
        ]
    return [
        {"kind": "narrate", "text": "The light is invisible in daylight.",
         "recipients": "target", "target": args["target"]},
    ]
```

(The real `cast_light` fixture —
`internal/bundle/testdata/worlds/scripted/bundles/demo/tools/cast_light/tool.star`.)

`apply()` returns a plain Starlark list of effect dicts drawn from the same
closed vocabulary as declarative `effects`, except values are already
resolved — no `{args.x}` templating inside a script; just build the string or
int directly. An empty list is a valid no-op result.

### Inputs

- `args` — a frozen dict of the already-validated invocation arguments.
  `agent_name`/`text`/`enum` params arrive as `str`; `number` params arrive as
  `int`.
- `world` — a frozen, **invoker-scoped**, read-only view:

  | Member | Type | Meaning |
  |---|---|---|
  | `world.tick` | int | current game tick |
  | `world.time_of_day` | str | `"night" \| "morning" \| "day" \| "evening"` |
  | `world.map_width`, `world.map_height` | int | map dimensions |
  | `world.agents()` | list of struct | living agents: `.name`, `.x`, `.y`, `.alive` |
  | `world.agent(name)` | struct or `None` | lookup by exact name (can report a dead agent's `.alive == False`) |
  | `world.rand(purpose, index)` | float in `[0,1)` | the ONLY randomness available — deterministic |

  `time_of_day`'s night boundary (22:00–06:00) is pinned to the same
  definition the sim engine itself uses for "is it night", so a script's
  branch always agrees with the world's. Deliberately and permanently absent:
  private memories, beliefs, relationships, journals, pending orders,
  LLM/provider state, wall-clock time, filesystem, network, environment. There
  is no `load()` (module loading is disabled) and no `print` side-channel into
  world state (`print` only reaches the daemon debug log).

### `world.rand` and determinism

`world.rand(purpose, index)` is routed through the engine's seeded
`rngAt`-style derivation, namespaced `"bundle:<tool>:<purpose>"` so two tools
(or two purposes within one tool) never draw from the same stream. Given
identical `(args, world snapshot, world seed, tick)`, `apply()` **must**
produce identical effects — guaranteed structurally by Starlark's
deterministic evaluation plus the total absence of ambient capabilities, not
by author discipline.

### Failure semantics

Every failure below rejects the **whole invocation** — nothing lands, no
charge is spent:

| Failure | Detected | Reported |
|---|---|---|
| parse error / no `apply` defined | boot | boot error naming the file (tool skipped, rule T6) |
| runtime error (`fail()`, type error) | invocation | descriptive message back to the invoking agent |
| step cap exceeded | invocation | deterministic abort, descriptive message |
| malformed effect dict | invocation | compiler rejection, descriptive message |
| undeclared event type | invocation | compiler rejection, descriptive message |
| batch fails the reducer's dry-run probe | invocation | `InjectSocial` rejection, descriptive message |

A script that busy-loops past its step cap aborts deterministically — same
inputs, same abort, every time — it does not hang the daemon or land partial
state.

## Persona bundles

A bundle may install an agent identity and a capability narrowing alongside
(or instead of) tools:

- **`SOUL.md`** — a short first-person identity fragment (≤4000 characters,
  valid UTF-8), appended to Metatron's system prompt after the charter. An
  oversized or invalid `SOUL.md` rejects the **whole bundle** (rule B2) —
  broken identity is not something to silently truncate.
- **`capabilities.json`** — narrows what the persona may use from the
  world's existing grant. It can only *narrow*, never widen:

  ```jsonc
  { "miracle_kinds": ["remove", "give_item", "time_snap"] }
  ```

  (The real gandalf fixture —
  `internal/bundle/testdata/worlds/persona/bundles/gandalf/capabilities.json`
  — narrows `work_miracle`'s kind enum to exclude `"move"` even though the
  world grants it.) An explicit `"tools"` list works the same way, over the
  flat tool-name namespace (built-in tools and bundle tools alike). Omitting a
  key narrows nothing on that axis; a malformed `capabilities.json` rejects
  the whole bundle (rule B3). Multiple persona bundles compose by
  intersection — load order does not matter.

A persona's own `tools/` folder validates by the exact same T1–T7 ladder as
any other bundle's tools, and a broken tool in a persona bundle is skipped on
its own — the SOUL fragment, the grant narrowing, and any sibling valid tools
all stay active (this is the same "per-tool failure never brings down its
bundle" rule as everywhere else).

## Boot validation — how errors are reported

Validation runs once at boot, after the built-in tool registry validators and
before the sim loop starts. Every skip or rejection produces one line on the
daemon's boot log, naming the file, the rule, and the specific offending
value — never a generic "invalid manifest":

```
daemon: bundle=demo tool=healtool file=demo/tools/healtool/tool.json rule=T3 severity=error: tool "healtool" declares event "metatron.heal", which is not an injectable event type
daemon: bundles on (1 tool(s) from 1 bundle(s))
```

The ladder, in the order it runs (see `contracts/boot-validation.md` for the
authoritative table):

| Rule | Level | Checks |
|---|---|---|
| B1 | bundle | folder name shape |
| B2 | bundle | `SOUL.md` UTF-8 + ≤4000 chars |
| B3 | bundle | `capabilities.json` parses |
| B4 | bundle | ≤16 tool dirs |
| T1 | tool | `tool.json` present, decodes, name == folder |
| T2 | tool | param shapes |
| T3 | tool | `events` non-empty, every one an injectable event type |
| T4 | tool | exactly one of `effects` / `script` |
| T5 | tool | declarative effects well-formed; producible events == declared |
| T6 | tool | script exists, parses, defines `apply` |
| T7 | tool | `limits.max_steps` and `charges` bounds |
| C1 | tool | collides with a built-in tool name → **built-in wins**, warning |
| C2 | tool | collides with an earlier-loaded bundle tool → **first-loaded wins**, warning |

**Scope matters**: a B-rule failure rejects the entire bundle (SOUL, grant,
and every tool in it); a T-rule failure skips only that one tool — its
siblings, the bundle's SOUL fragment, and its grant narrowing all still take
effect. This is deliberate: a broken identity or a broken permissions file
undermines the whole bundle's trustworthiness, but one broken tool among
several says nothing about the others.

## Collision rules

A tool name can collide two ways, and both resolutions are silent-but-logged
(a warning, not an error — the world still boots with the surviving tool):

- **C1 — vs. a built-in tool.** The built-in always wins; the bundle tool is
  skipped. This is why the dogfood example
  (`examples/bundles/dogfood-move/tools/miracle_move/`) ships under a
  *different* name than the built-in `work_miracle` door it re-expresses —
  naming it identically would simply never install while `work_miracle` is
  granted.
- **C2 — vs. an earlier-loaded bundle tool.** First-loaded wins (bundle load
  order is the deterministic bytewise folder order described above); every
  later bundle declaring the same tool name is skipped.

## Determinism guarantees

Bundle tools sit inside the same replay contract as every other Metatron
miracle: **events are the only thing that is ever replayed.** Concretely:

- A live invocation compiles effects → `store.Event`s once, at invocation
  time, and those events are what gets logged. Replay re-applies the logged
  events through the reducer; it never re-parses a manifest, never
  re-executes a `tool.star` script, and never re-derives a random draw.
- This means a bundle can be edited or deleted after the fact, and replaying
  the world from its event log still reproduces the exact same `State.Hash()`
  — bundle removal cannot corrupt history.
- For a script to produce identical output on identical inputs in the first
  place (the property that makes the logged events trustworthy when they
  *were* produced), the sandbox removes every source of variance: no wall
  clock, no I/O, no environment, no module loading, and the only randomness
  available (`world.rand`) is derived from the world's own seed plus the
  tick and a purpose string — never `math/rand` or any unseeded source.

## Trying it yourself

1. Create `<worldDir>/bundles/<name>/tools/<tool>/tool.json` (see the
   templates above, or copy a fixture under
   `internal/bundle/testdata/worlds/` or `examples/bundles/dogfood-move/`).
2. `promptworld start <world>` (or `promptworld daemon <world>` in the
   foreground) and watch the boot log for `daemon: bundle=...` lines — a
   clean load shows only the `bundles on (N tool(s) from M bundle(s))`
   summary line, with no error/warning lines above it.
3. Fix whatever the log names, restart, repeat.

No LLM configuration is required to prove a bundle loads — boot validation
and roster assembly run independently of any provider being reachable.
