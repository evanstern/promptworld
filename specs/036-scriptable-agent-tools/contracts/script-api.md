# Contract: Bundle Tool Script API (`tool.star`)

**Consumers**: bundle authors (write), `internal/bundle/script.go` (execute),
`internal/bundle/worldview.go` (provide `world`).

Language: Starlark (`go.starlark.net`), recursion disabled, step-capped
(`limits.max_steps`, default 100k, hard ceiling 1M). The script file must define:

```python
def apply(args, world):
    """Pure function: (args, world) -> list of effect dicts."""
    ...
    return effects
```

## Inputs

- `args` — a frozen dict of the validated invocation arguments (post toolloop schema check).
  Types: `agent_name`/`text`/`enum` → `str`; `number` → `int`.
- `world` — frozen, read-only, **invoker-scoped** view (clarification #3):

| Member | Type | Meaning |
|---|---|---|
| `world.tick` | int | current game tick |
| `world.time_of_day` | str | `"night" \| "morning" \| "day" \| "evening"` (derived from game clock) |
| `world.map_width`, `world.map_height` | int | map dimensions |
| `world.agents()` | list of struct | living agents: `.name` str, `.x` int, `.y` int, `.alive` bool |
| `world.agent(name)` | struct or `None` | lookup by exact name |
| `world.rand(purpose, index)` | float in [0,1) | ONLY randomness; deterministic — engine-side `rngAt(worldSeed, "bundle:<tool>:" + purpose, tick, index)` |

Deliberately absent (and must stay absent without spec-level review): private memories,
beliefs, relationships, journals, pending orders, LLM/provider state, wall-clock time,
filesystem, network, environment. There is no `load()` (module loading disabled) and no
`print` side-channel into world state (print goes to the daemon debug log only).

## Output

Return a Starlark list of effect dicts drawn from the closed vocabulary
(data-model.md “EffectTemplate / Effect”; resolved values, no `{args.x}` templating needed):

```python
def apply(args, world):
    if world.time_of_day == "night":
        return [
            {"kind": "narrate", "text": "A soft light blooms over " + args["target"],
             "recipients": "all_living"},
        ]
    return [{"kind": "narrate", "text": "The light is invisible in daylight.",
             "recipients": "target", "target": args["target"]}]
```

Rules enforced by the effect compiler after `apply()` returns:
- Result must be a list of dicts, each with a known `kind` and exactly the fields that kind
  requires (unknown fields rejected).
- ≤32 events after `narrate` recipient expansion; `narrate.text` ≤500 bytes; numeric fields
  must be finite ints in range.
- Compiled event types must be ⊆ the manifest's declared `events`.
- Empty list is a valid no-op result.

## Failure semantics (all reject the whole invocation, land nothing, spend no charge)

| Failure | Detected | Reported |
|---|---|---|
| parse error / no `apply` | boot | boot error naming file (tool skipped) |
| runtime error (`fail()`, type error) | invocation | `ResultForModel` to invoking agent |
| step cap exceeded | invocation | deterministic abort, `ResultForModel` |
| malformed effect dict | invocation | compiler rejection, `ResultForModel` |
| undeclared event type | invocation | compiler rejection, `ResultForModel` |
| batch fails probe dry-run | invocation | `InjectSocial` rejection, `ResultForModel` |

## Determinism obligations (FR-011, SC-003)

Given identical (args, world snapshot, world seed, tick), `apply()` MUST produce identical
effects — guaranteed structurally: Starlark's deterministic evaluation + no ambient
capabilities + `world.rand` derived from the seeded `rngAt` pattern. Replay never re-executes
scripts (events are self-contained data), so bundle edits/removals cannot corrupt replay.
