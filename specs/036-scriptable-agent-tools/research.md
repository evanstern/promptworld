# Phase 0 Research: Scriptable Agent Tools

**Feature**: [spec.md](spec.md) | **Plan**: [plan.md](plan.md) | **Date**: 2026-07-24

All spec-level clarifications were resolved in the spec's Clarifications session (2026-07-24).
This document resolves the four design-level unknowns the plan depends on. Code anchors were
verified against the working tree at plan time (see plan.md Technical Context).

## R1. Scripting runtime — DECIDED: `go.starlark.net` (starlark-go)

**Decision**: embed Starlark via the official `go.starlark.net` package as the script runtime
for scripted bundle tools. New direct dependency (the module currently has five direct deps and
no script runtime, direct or indirect).

**Rationale**:
- **Hermetic by design.** Starlark has no I/O, no filesystem, no network, no clock, and no
  ambient randomness in its core; sandboxing is by *construction*, not by subtraction. The only
  capabilities a script has are the ones we pass in as predeclared values (`args`, `world`).
- **Deterministic by design.** Dict iteration is insertion-ordered, string/number semantics are
  fixed, and the language spec explicitly targets reproducible evaluation — matching the
  project's replay-determinism constraint (byte-identical `State.Hash()`).
- **Native execution budgets.** `starlark.Thread` supports a step limit
  (`Thread.SetMaxExecutionSteps`) that aborts evaluation deterministically when exceeded — a
  *logical* budget, unlike wall-clock deadlines, so the same script + inputs always aborts at
  the same point on every machine. Recursion is disabled by default; loops are bounded by
  iterables.
- **Right-sized surface.** Functions, conditionals, comprehensions — enough for "cast light
  behaves differently at night"; not enough to build an OS. Static parse
  (`starlark.SourceProgram`) supports the boot-time "script parses" gate cheaply.
- **Maintained Go-native implementation** (the Bazel language), pure Go, no cgo.

**Alternatives considered**:
- **gopher-lua** (task sketch's other candidate): sandbox by subtraction — must strip `os`,
  `io`, `math.random`, and friends and keep them stripped across upgrades; its execution
  limiting idiom is context-deadline based (wall-clock ⇒ non-deterministic abort points).
  Rejected: every sandbox property we need is manual where Starlark's is structural.
- **goja (JavaScript)**: full ES runtime; much larger attack/determinism surface (Date, Math.
  random to stub out, event-loop semantics); heavier than the problem. Rejected.
- **cel-go**: expression-only — no multi-statement logic or local bindings suited to composing
  multi-effect batches. Too small. Rejected.
- **WASM (wazero)**: strongest isolation, but authors would need a compile toolchain — kills
  the "drop a folder" authoring loop. Determinism achievable but budget/metering machinery is
  heavyweight. Rejected for v1; remains the escape hatch if bundles ever need untrusted-grade
  isolation.

**Residual risks & mitigations**: Starlark has no native *memory* cap → mitigate with the step
cap (allocation is step-bounded), input-size caps on args, and hard output caps (≤32 events,
text byte caps) enforced by the effect compiler after `apply()` returns. Float ops are IEEE-754
and deterministic across our platforms; the effect compiler additionally rejects NaN/Inf in
numeric effect fields so no float weirdness can reach a payload.

## R2. Registration seam — DECIDED: compose at turn assembly, never touch the registry

**Decision**: bundle tools do NOT enter `internal/tool/registry.go`'s compile-time literal
registry. A new `internal/bundle` package produces a boot-frozen `BundleSet` holding synthesized
`tool.Tool` values + handler factories; the metatron turn assembly (`internal/metatron/turn.go:144-198`)
composes `grantedRoster(grant) + bundleSet.Roster()` into the `toolloop.Job` roster and merges
the handler maps.

**Rationale**:
- The registry is a leaf package whose **registration order is byte-identity load-bearing** for
  derived prompt strings (`registry.go:525-533`); injecting dynamic entries there risks
  perturbing every derived surface for zero benefit.
- The toolloop already takes per-job `Roster []tool.Tool` + `Handlers map[string]Handler`
  (`internal/toolloop/loop.go:53-72`) — the seam exists; built-in metatron tools already flow
  through it (`turn.go:188-198`, `toolcalls.go:69-96`). Bundle tools ride the identical path,
  so roster-membership, schema validation, and failure feedback (`loop.go:299-382`) apply
  unchanged.
- Boot-time coverage checks stay honest: `tool.Validate()` + `sim.ValidateToolCoverage()` keep
  gating the static registry; `bundle.Validate()` applies the *same* rules (params well-formed,
  events ⊆ whitelist — mirroring `internal/sim/toolcheck.go:62-67`) to synthesized tools.
- One derived-surface fix is required: `tool.MetatronToolGuidance` reads a hardcoded
  description map (`derive.go:201-222`) and would render bundle tools with empty guidance →
  add a `PromptGloss` fallback (benefits future built-ins too).

**Alternatives considered**: (a) dynamic `tool.Register()` API — rejected: breaks the leaf
package's immutability + ordering guarantees and lets bundle state leak into global surfaces;
(b) a parallel second registry consulted everywhere — rejected: every consumer would need
dual-lookup; the per-job roster already *is* the composition point.

## R3. Bundle layout & lifecycle — DECIDED: `<worldDir>/bundles/<name>/`, boot-frozen

**Decision**: bundles live under a single `bundles/` folder in the world dir:

```text
<worldDir>/bundles/gandalf/
├── SOUL.md                      # optional persona/charter fragment
├── capabilities.json            # optional grant narrowing (intersection with world grant)
└── tools/
    └── cast_light/
        ├── tool.json            # manifest (required)
        └── tool.star            # script (optional — absent ⇒ manifest-only tool)
```

Discovery, validation, and freezing happen once at daemon boot into a `BundleSet`; editing a
bundle requires a world restart (spec assumption: hot reload out of scope).

**Rationale**:
- Mirrors the established world-dir file-loading precedent — `charter.md`, `skills/*.md`,
  `capabilities.json` (`internal/metatron/charter.go:27-46,82-123,241-306`) — including its
  determinism discipline: direct children only, dotfiles/unknown files ignored, **ascending
  bytewise name order** (which also implements clarification #2's "first-loaded wins").
- A dedicated `bundles/` sibling of `skills/` keeps `world.Open`'s manifest validation
  untouched and gives `World` a single new accessor (`BundlesDir()`), consistent with
  `world.go:156-171`.
- Boot-freeze (vs the charter's per-turn reread) is what FR-008's validation gate and SC-005's
  error reporting attach to; it also means invocation-time behavior can't drift mid-run,
  protecting the determinism story. SOUL fragments and grants are *read from the frozen set*
  at turn time, composing with the per-turn charter/skills reads.
- Persona `capabilities.json` uses **intersection semantics** (can narrow the world grant,
  never widen it): the world owner's grant remains the authority; a dropped folder cannot
  self-escalate.

**Alternatives considered**: top-level `<worldDir>/<name>/` folders per the original idea
sketch — rejected: collides with existing reserved names (`agents/`, `metatron/`, `skills/`)
and makes discovery/validation ambiguous; a `bundles/` root is one mkdir away for authors.

## R4. Effect vocabulary & script output contract — DECIDED: effects, not events

**Decision**: scripts and manifests both express results as **effect dicts** from a closed v1
vocabulary; a single compiler (`internal/bundle/effects.go`) is the only code path that turns
effects into `store.Event`s. v1 vocabulary (zero reducer/whitelist changes):

| Effect | Compiles to | Payload source |
|---|---|---|
| `move_entity` | `metatron.entity_moved` | `internal/sim/miracles.go:44-52` (`EntityMovedPayload`) |
| `remove_entity` | `metatron.entity_removed` | `EntityRemovedPayload` |
| `grant_item` | `metatron.item_granted` | `ItemGrantedPayload` |
| `snap_time` | `metatron.time_snapped` | `TimeSnappedPayload` |
| `narrate` | `agent.memory_added` (per recipient: `target` / `all_living` / `named`) | `MemoryAddedPayload`, mirroring `miracle_batch.go:83-88` |

**Rationale**:
- FR-004/FR-005 require that loaded tools cannot invent event types or hand-craft payloads;
  making the compiler the sole event factory enforces this *structurally* — a script literally
  has no API for constructing an event, only for requesting audited effects.
- Both authoring modes share one pipeline: a manifest-only tool is just a constant effect list
  with `{args.x}` template substitution; a scripted tool computes the same list. One compiler,
  one validation path, one test surface.
- Invocation-time gate: compiled event types must be ⊆ the manifest's declared `events` (which
  boot validation already proved ⊆ `injectSocialWhitelist`, mirroring
  `internal/sim/toolcheck.go:62-67`) — then the batch rides `InjectSocial`'s existing atomic
  probe-copy dry run (`internal/sim/loop.go:534-572`) and reducer-side validate-then-spend
  (`internal/sim/miracles.go:78-91`). The engine's existing behavior satisfies the clarified
  charge policy (failures spend nothing) with no new code.
- **Charge**: per-event costs remain reducer-authoritative (`tool.MiracleCostsByEvent()`,
  `registry.go:148-154`); the manifest's `charges` field is the *gate minimum* (`Cost.Charges`)
  exactly like `work_miracle`'s — no bundle can undercut the reducer's price.
- Vocabulary growth ("heal") is deliberately out of v1 (spec assumption): each new effect =
  new event type + reducer arm + whitelist entry + replay test, one audited effect at a time.

**Alternatives considered**: exposing raw event construction to scripts (rejected: FR-005
violation by construction); an imperative engine API (`move_agent(...)` mutating mid-script) —
rejected in the task's design sketch already: re-entrancy, non-atomicity, and a far larger
sandbox surface vs. the pure `effects-out` model.
