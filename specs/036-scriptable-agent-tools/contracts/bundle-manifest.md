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
  { "kind": "remove_entity","target": "{args.target}" },
  { "kind": "narrate",      "text": "{args.target} vanished in a poof of smoke",
    "recipients": "all_living" }                        // "target" | "all_living" | ["name", …]
]
```

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
