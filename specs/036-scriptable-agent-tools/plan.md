# Implementation Plan: Scriptable Agent Tools — Pluggable Bundle-Defined Tools

**Branch**: `task-85-scriptable-agent-tools` | **Date**: 2026-07-24 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/036-scriptable-agent-tools/spec.md`

## Summary

Make angel/expressive tools loadable from bundle folders dropped into a world directory, instead of hard-coded in the compile-time registry. A bundle tool is a manifest (name, description, params, declared events, charge gate, effects) plus an optional Starlark script; both compile to an **effect batch** that the existing `InjectSocial` door validates atomically against the `injectSocialWhitelist` and lands through the existing reducer path — scripts never mutate state and never construct raw events. v1 scope: metatron/expressive tools only; tick-simulated world verbs stay native Go.

The engine already provides the hard parts end-to-end: atomic dry-run batch validation on a probe state copy (`internal/sim/loop.go:534-572`), whitelist gating (`internal/sim/loop.go:193-256`), reducer-side validate-then-spend charge economy (`internal/sim/miracles.go:78-91`), a deterministic seeded-RNG pattern (`internal/sim/rng.go:11-16`), and a per-world-dir file-loading precedent (`internal/metatron/charter.go`). What's new is: a bundle loader/validator package, an effect→event compiler, a sandboxed Starlark executor, an invoker-scoped world view, and the roster/handler/prompt merge seam in the metatron turn assembly.

## Technical Context

**Language/Version**: Go 1.26.4 (module `github.com/evanstern/promptworld`)

**Primary Dependencies**: existing — `anthropic-sdk-go`, `modernc.org/sqlite`, bubbletea/lipgloss (TUI). **New direct dependency: `go.starlark.net` (starlark-go)** for the script runtime — decision and alternatives in [research.md](research.md) R1.

**Storage**: existing SQLite event log (`world.db`, single-writer WAL); bundles are plain files under `<worldDir>/bundles/` — no new storage.

**Testing**: `go test` (stdlib); determinism via the established live-apply vs replay `State.Hash()` byte-identity pattern (`internal/sim/miracles_test.go:398` et al.); sandbox-escape and cap-exhaustion tests; dogfood equivalence test (built-in vs bundle twin).

**Target Platform**: the promptworld daemon (darwin/linux), same as today.

**Project Type**: single Go module — CLI + daemon + TUI.

**Performance Goals**: bundle tool invocation adds negligible latency vs built-in tools (script step cap default 100k steps ≈ sub-millisecond; hard engine ceiling 1M steps). Boot-time validation of a world with ≤32 bundles completes in <1s.

**Constraints**: replay determinism is load-bearing — byte-identical `State.Hash()` after replay; scripts get no wall clock, no I/O, no ambient randomness (seeded `rngAt`-derived helper only); whitelist isolation — no event type outside `injectSocialWhitelist` can ever land; atomic all-or-nothing batches (existing probe-copy dry run).

**Scale/Scope**: tens of bundles per world, ≤16 tools per bundle (cap, mirrors `maxSkillFiles=8` discipline); effect batches capped (≤32 events, text caps per existing `Cost.TextCapBytes` conventions).

## Constitution Check

*GATE: evaluated against constitution v1.1.0 before Phase 0; re-checked after Phase 1 design — PASS (no violations).*

- **I. Artifact-Grounded Action** — PASS. Spec, clarifications, this plan, research, data model, and contracts are durable artifacts; the feature itself extends the artifact discipline (bundles are files; every effect lands as a recorded event).
- **II. One Task, One PR** — PASS. TASK-85 → one branch (`.worktrees/task-85`) → one PR. Spec phases are internal breakdown.
- **III. Gates Over Assertions** — PASS. Boot validation gates bundles against real files; the whitelist + probe dry-run gates batches against real state; spec-bridge gates board status against spec artifacts.
- **IV. Grounding Freshness** — PASS with obligation: this feature touches files pinned by wiki notes (`tool-registry.md`, `sim-loop.md`, `metatron-miracles.md`, `metatron.md`, `deterministic-rng.md`, `event-types.md`, likely `world-save-directory.md`) — `/grounding-wiki:wiki-update` is required before the TASK is done, followed by the `player-docs` freshness check.
- **V. Model-Tiered Workflow** — PASS. Plan authored on the planning tier. Implementation slices are cross-package/architectural (new `internal/bundle` package, sim/tool/metatron integration, sandboxed runtime, determinism-critical code) → **Opus 4.8** per the escalation rubric; routine slices (fixture bundles, docs reconciliation, quickstart validation) → Sonnet. Tier choices recorded on TASK-85 at task-generation time.

**Post-Phase-1 re-check (2026-07-24)**: design introduces one new package and one new direct dependency, both minimal and justified in research.md R1/R2; no violations. Complexity Tracking left empty.

## Project Structure

### Documentation (this feature)

```text
specs/036-scriptable-agent-tools/
├── spec.md              # Feature specification (clarified 2026-07-24)
├── plan.md              # This file
├── research.md          # Phase 0: runtime choice, registration seam, layout, view scope
├── data-model.md        # Phase 1: manifest schema, effect vocabulary, entities
├── quickstart.md        # Phase 1: end-to-end validation guide
├── contracts/
│   ├── bundle-manifest.md   # tool.json contract (authoring surface)
│   ├── script-api.md        # Starlark script contract (apply(), world view, effect dicts)
│   └── boot-validation.md   # validation rules + error reporting contract
└── tasks.md             # Phase 2 (/speckit-tasks — not created by plan)
```

### Source Code (repository root)

```text
internal/bundle/                 # NEW package: everything bundle-specific
├── manifest.go                  # tool.json parse + schema validation → tool.Tool synthesis
├── load.go                      # world-dir discovery (deterministic order), BundleSet, collision rules
├── validate.go                  # boot-time gate: schema, events ⊆ whitelist, script parse, caps
├── effects.go                   # effect vocabulary + effect→store.Event compiler (both modes)
├── script.go                    # Starlark executor: sandbox, step cap, seeded RNG, apply() contract
├── worldview.go                 # invoker-scoped read-only world view exposed to scripts
└── *_test.go                    # unit tests incl. sandbox-escape + cap tests

internal/metatron/
├── turn.go                      # MODIFIED: merge bundle roster + handlers + SOUL fragments per turn
├── toolcalls.go                 # MODIFIED: handler factory for bundle tools → InjectSocial
└── charter.go                   # (pattern source; capabilities merge for persona bundles)

internal/tool/
└── derive.go                    # MODIFIED: MetatronToolGuidance falls back to PromptGloss
                                 #           for tools absent from metatronToolDesc

internal/sim/
└── (no reducer changes in v1)   # existing whitelist + probe dry-run + spend logic reused as-is

internal/world/
└── world.go                     # MODIFIED: BundlesDir() accessor (`<dir>/bundles`)

internal/daemon (or equivalent boot path)
└── boot                         # MODIFIED: load + validate BundleSet at boot; report skips/errors

testdata / fixtures
└── internal/bundle/testdata/    # fixture bundles: valid, invalid-manifest, off-whitelist,
                                 # over-budget script, dogfood twin, persona bundle
```

**Structure Decision**: one new leaf-ish package `internal/bundle` (depends on `tool`, `sim` read surface, `store`, starlark; nothing depends on it except metatron turn assembly and daemon boot). Bundle tools never enter the compile-time registry (`internal/tool/registry.go` stays a literal); they compose into the per-turn roster/handler/prompt triad in `internal/metatron/turn.go:144-198`, exactly where `grantedRoster` already assembles the granted surface. This keeps `tool` a leaf package and keeps registration-order byte-identity guarantees intact.

## Design Overview (Phase 1 summary)

Full details in [data-model.md](data-model.md) and [contracts/](contracts/).

1. **Two authoring modes, one pipeline.** A manifest-only tool declares `effects[]` — parameterized templates over the audited effect vocabulary. A scripted tool ships `tool.star` defining `apply(args, world)` that *returns the same effect dicts*. Both modes feed the same effect→event compiler (`effects.go`), so scripts can never construct raw `store.Event`s; the compiler is the only event factory, and it only emits whitelisted types the manifest declared.

2. **Effect vocabulary v1** (compiles to existing whitelisted events, zero reducer changes): `move_entity` → `metatron.entity_moved`; `remove_entity` → `metatron.entity_removed`; `grant_item` → `metatron.item_granted`; `snap_time` → `metatron.time_snapped`; `narrate` → `agent.memory_added` (recipient selectors: `target` / `all_living` / `named`), mirroring `BuildMiracleBatch`'s trailing-perception pattern (`internal/metatron/miracle_batch.go:83-88`).

3. **Landing path** (unchanged engine): handler → compile effects → assert every event type ∈ manifest `events` ∩ whitelist → `InjectSocial` → probe-copy dry run → atomic land; reducer arms validate and spend charge (`spendMiracleCharge`) exactly as for built-ins. Failed batches land nothing and spend nothing — already the engine's behavior, satisfying the clarified charge policy for free.

4. **Sandbox** (research.md R1): starlark-go with recursion off, step cap (`manifest.limits.max_steps`, default 100k, hard ceiling 1M), no time/io/net modules, deterministic iteration, output caps (≤32 events, text byte caps). Randomness only via `world.rand(purpose, index)` backed by the existing `rngAt(seed, "bundle:"+tool+":"+purpose, tick, index)` pattern.

5. **Invoker-scoped world view** (clarification #3): `worldview.go` exposes only what the metatron's existing prompt surfaces legitimately expose — clock/game time, living agents (name, position, alive), entity lookup, map dims. No private memories, beliefs, or hidden state. Contract in [contracts/script-api.md](contracts/script-api.md).

6. **Persona bundles**: `<worldDir>/bundles/<name>/` with optional `SOUL.md` (charter fragment appended to the system prompt after `charter.md`, same caps discipline), optional `capabilities.json` (merged with the world-level grant: intersection semantics — a persona can narrow, never widen), and `tools/<tool>/`. Per-tool failure skips the tool; invalid SOUL/capabilities rejects the persona (clarification #1). Built-ins win name collisions; first-loaded (bytewise dir order) wins among bundles (clarification #2).

7. **Boot vs turn**: bundles are discovered, validated, and frozen into a `BundleSet` at daemon boot (FR-008; hot reload out of scope). The metatron turn assembly composes `grantedRoster(grant) + bundleSet.Roster()` and merges `turnHandlers(d) + bundleSet.Handlers(...)`; `tool.MetatronToolGuidance` gains a PromptGloss fallback so bundle tools render real guidance (today's hardcoded map at `internal/tool/derive.go:201-222` would render them empty).

8. **Determinism proof**: a `TestBundleToolReplayByteIdentity` following `internal/sim/miracles_test.go:398`'s live-vs-replay `State.Hash()` pattern, plus a scripted-tool variant exercising `world.rand`.

## Complexity Tracking

> No constitution violations — table intentionally empty.
