# Contract: Bundle Tool Manifest (`tool.json`)

**Consumers**: bundle authors (write), `internal/bundle/manifest.go` (parse/validate),
`internal/tool` derivation (via synthesized `tool.Tool`).

One file per tool at `<worldDir>/bundles/<bundle>/tools/<tool>/tool.json`. UTF-8 JSON object.
Unknown top-level keys are rejected at boot (strict decode) so typos fail loudly.

## Schema

```jsonc
{
  "name": "cast_light",              // REQUIRED [a-z0-9_]{1,48}; must equal folder name
  "description": "Conjure light …", // REQUIRED 1–500 chars; LLM-facing PromptGloss
  "params": [                        // OPTIONAL ≤8 entries
    {
      "name": "target",              // REQUIRED [a-z0-9_]{1,32}
      "kind": "agent_name",          // REQUIRED agent_name|text|enum|number
      "required": true,              // OPTIONAL default false
      "description": "who to bless", // OPTIONAL ≤200 chars
      "enum": ["soft", "blinding"],  // REQUIRED iff kind=enum; 1–16 non-empty values
      "min": 0, "max": 10,           // OPTIONAL iff kind=number (min ≤ max)
      "max_bytes": 500               // OPTIONAL iff kind=text (default 500)
    }
  ],
  "events": [                        // REQUIRED unless narration-only (then may be
    "metatron.entity_moved",         //   exactly ["agent.memory_added"])
    "agent.memory_added"             // MUST be ⊆ injectSocialWhitelist (boot gate)
  ],
  "charges": 1,                      // OPTIONAL default 0; gate minimum, not the spend
  "effects": [ /* see below */ ],    // exactly one of effects | script
  "script": "tool.star",             // relative file name; must exist and parse at boot
  "limits": { "max_steps": 100000 }  // script only; default 100000, ceiling 1000000
}
```

## Declarative `effects` entries

Templates over the closed effect vocabulary (data-model.md “EffectTemplate”). String fields
support `{args.<param>}` and `{invoker}` substitution only (no expressions, no nesting).

```jsonc
"effects": [
  { "kind": "move_entity",  "target": "{args.target}", "to_x": "{args.x}", "to_y": "{args.y}" },
  { "kind": "grant_item",   "target": "{args.target}", "item": "bread", "qty": 2 },
  { "kind": "snap_time",    "to_tick": "{args.to_tick}" },
  { "kind": "remove_entity","target": "structure@{args.x},{args.y}" },
  { "kind": "narrate",      "text": "{args.target} vanished in a poof of smoke",
    "recipients": "all_living" }                        // "target" | "all_living" | ["name", …]
]
```

## Target addressing (spec 082)

A `target` string is a spec-082 **address**. The normative grammar, normalization,
resolution semantics, and error taxonomy live in
`specs/082-target-addressing/data-model.md`; the single parser implementing it is
`internal/target` (one authority — no consumer re-implements it). Substitution
(`{args.x}`/`{invoker}`) runs FIRST; the parser sees resolved strings, in declarative and
script mode alike (one shared compile path — no script-specific address surface).

Forms:

| Form | Example | Meaning |
|---|---|---|
| bare name | `Rega` | living villager by name (v1 compat — unchanged, byte-identical) |
| typed name | `villager:Rega` | same, explicit |
| point | `structure@12,7` | the entity of that class on the tile |
| rect | `structure@1,5..3,9` | inclusive rectangle — **reserved** (designations, below) |
| line | `structure@1,5->1,9` | inclusive axis-aligned line — **reserved** (designations, below) |

**Reserved-prefix rule (compat)**: a target matching `^(villager|structure|pile|terrain)[@:]`
MUST parse as a structured address — if malformed it is a compile-time syntax error, never a
fallback to a villager name. Every other string is a bare villager name. No roster name
contains `@` or `:`, and future roster additions MUST NOT — this rule is what makes the
extension backward-compatible for every v1-legal target.

Per-kind form matrix (spec 082 data-model.md §4 — the authority):

| Effect kind | villager (name / `villager:` / `villager@X,Y`) | `structure@X,Y` | `pile@X,Y` | `terrain@X,Y` | rect / line |
|---|---|---|---|---|---|
| `move_entity` | ✅ | ✅ | ✅ | ❌ form error | ❌ form error |
| `remove_entity` | ❌ form error (a villager can never be removed — reducer doctrine, mirrored compiler-side; the reducer arm stays authoritative) | ✅ | ✅ | ✅ | ❌ form error |
| `grant_item` | ✅ | ❌ | ❌ | ❌ | ❌ |

Resolution is deterministic and reads only the invocation's state snapshot:
villager names first-match by roster index (case-insensitive, trimmed); `villager@X,Y` via
`sim.State.VillagerAt` (first living by agent index — the miracle door's own choice);
`structure@`/`pile@` via the one-per-tile presence probes (`HasStructureAt`/`HasPileAt`);
`terrain@` is bounds-checked only (removability stays reducer-side). Compiled payloads are
byte-identical to `guardian.BuildMiracleBatch`'s for the same class+tiles, `Gratis:false`.

Errors: every target failure (`syntax`/`class`/`form`/`bounds`/`unresolved` — spec 082
data-model.md §5) rejects the WHOLE invocation atomically as a descriptive message naming
the effect index, the field, and the offending address; nothing lands, no charge is spent.
Reducer-side rejections (occupied site, already-overlaid terrain, charge shortfall) remain
the dry run's, unchanged.

**Param-kind guidance**: an argument that carries an address (or composes into one) is
declared kind `text` — `agent_name` params stay villager-name-validated and are for
name-only targets.

## Designation addressing (TASK-157 seam)

The grammar deliberately serves a second consumer: guardian **designations** (TASK-157 —
settlement zone, structure site, wall line). This contract guarantees the seam:

- **One parser, leaf-safe**: `internal/target` imports the standard library only (pinned by
  test), so `internal/tool` — which may import neither `sim` nor `bundle` — can adopt it
  without a cycle or a new dependency edge. Designation params parse through the SAME
  package bundle effects use.
- **Point, rect, line**: a structure site is a point; a settlement zone is an inclusive
  rect whose corners normalize to (min,min)-(max,max) at parse; a wall line is an inclusive
  axis-aligned line whose endpoint ORDER IS PRESERVED (direction is author intent).
  `Address.Tiles()` enumerates deterministically — rect row-major (y ascending, then x
  ascending), line in endpoint order stepping ±1 — as a pure function of the Address, so
  any two consumers enumerate identically.
- **Bundle effects reserve the shapes**: rect and line forms are rejected in every bundle
  effect with a form error naming the reservation ("reserved for designation consumers,
  TASK-157"). Widening the bundle matrix later is a one-table change; nothing in the
  grammar hard-codes the current class set.

## Validation (boot — see contracts/boot-validation.md)

1. Strict JSON decode; `name` matches folder; param rules mirror `internal/tool/validate.go`.
2. `events` ⊆ `injectSocialWhitelist` AND `events` == union of event types producible by
   `effects`/script declaration (declarative mode: computed exactly; script mode: `events` is
   the promise, enforced per-invocation).
3. Exactly one of `effects` / `script`; script file exists, parses (`starlark.SourceProgram`),
   and defines `apply`.
4. `limits.max_steps` within (0, 1_000_000]; `charges` ≥ 0.

## Invocation-time guarantees

- Args validated against the derived JSON schema by the toolloop before any bundle code runs.
- Compiled event types MUST be ⊆ manifest `events`, else the whole invocation is rejected.
- Batch lands atomically via `InjectSocial` (probe dry run); failure feeds a descriptive
  `ResultForModel` back to the invoking agent and spends no charge.

## Compatibility

- Adding an OPTIONAL manifest key is backward-compatible; changing semantics of an existing
  key or shrinking `injectSocialWhitelist` is breaking and requires a boot-error message that
  names the offending manifest (spec edge case: vocabulary drift).
- Spec 082's target addressing is an ADDITIVE value-space extension of the existing `target`
  key: every v1-legal target (a bare villager name) compiles byte-identically (the
  reserved-prefix rule above), and no new manifest key, event type, or boot rule was added.
  The one behavior change is deliberate: `remove_entity` with a villager-designating target
  now fails at compile time with the doctrine message instead of compiling an event the
  reducer would reject.
