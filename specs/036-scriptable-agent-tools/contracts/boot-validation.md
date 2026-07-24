# Contract: Bundle Boot Validation & Error Reporting

**Consumers**: `internal/bundle/{load,validate}.go`, daemon boot path, world authors reading
the daemon log.

Runs once at daemon boot, after `tool.Validate()` / `sim.ValidateToolCoverage()` and before
the sim loop starts. Produces a frozen `BundleSet` plus a `BootReport`.

## Discovery (deterministic)

1. Root: `<worldDir>/bundles/` (absent ⇒ empty BundleSet, no error).
2. Bundles: direct child dirs, ascending bytewise name order; dotfile dirs skipped; non-dir
   entries ignored.
3. Tools within a bundle: `tools/` direct child dirs, same ordering; ≤16 (excess ⇒ bundle-level
   error, whole bundle rejected — a cap breach is structural).
4. Unknown extra files anywhere in a bundle are ignored (docs/licenses allowed).

## Validation ladder (per bundle)

| # | Check | Failure scope |
|---|---|---|
| B1 | bundle name matches `[a-z0-9_-]{1,64}` | bundle rejected |
| B2 | `SOUL.md` (if present) ≤4000 chars, valid UTF-8 | **bundle rejected** (clarification #1: broken identity) |
| B3 | `capabilities.json` (if present) parses per `manifestDoc` schema | **bundle rejected** (broken permissions) |
| B4 | ≤16 tool dirs | bundle rejected |
| T1 | `tool.json` present, strict-decodes, `name` == folder name | tool skipped |
| T2 | params well-formed (mirror `internal/tool/validate.go` rules) | tool skipped |
| T3 | `events` non-empty (or narration-only), ⊆ `injectSocialWhitelist` | tool skipped |
| T4 | exactly one of `effects` / `script` | tool skipped |
| T5 | declarative: every effect template well-formed; producible event types == `events` | tool skipped |
| T6 | script: file exists, `starlark.SourceProgram` parses, defines `apply` | tool skipped |
| T7 | `limits.max_steps` ∈ (0, 1_000_000]; `charges` ≥ 0 | tool skipped |
| C1 | tool name vs built-in registry (`tool.Lookup`) | tool skipped (built-in wins), **warning** |
| C2 | tool name vs earlier-loaded bundle tool | tool skipped (first wins), **warning** |

## BootReport entry shape

Every skip/rejection emits one entry — logged to the daemon log at boot and retrievable for
the status surface:

```text
bundle=<name> [tool=<name>] file=<relative path> rule=<B1…C2> severity=error|warning
msg=<human sentence naming the specific problem and the offending value>
```

SC-005 acceptance: the message must name the file and the specific problem — "manifest
declares event 'metatron.heal' which is not an injectable event type" not "invalid manifest".

## Guarantees

- A world with zero valid bundles boots normally (bundles are additive).
- Validation is read-only: no file in the bundle is ever modified or deleted by the engine.
- The same world dir always yields the same BundleSet and the same BootReport ordering
  (deterministic discovery + ladder).
- Post-freeze, invocation-time failures are per-invocation only (script-api.md failure table);
  a tool cannot be evicted from the roster mid-run.
