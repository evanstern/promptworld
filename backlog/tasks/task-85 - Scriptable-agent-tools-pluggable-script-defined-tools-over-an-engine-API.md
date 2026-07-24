---
id: TASK-85
title: 'Scriptable agent tools: pluggable script-defined tools over an engine API'
status: Done
assignee: []
created_date: '2026-07-24 03:02'
updated_date: '2026-07-24 20:19'
labels:
  - idea
dependencies: []
priority: high
ordinal: 5500
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Idea capture (2026-07-23): make agent/angel tools scriptable and pluggable instead of hard-coded. An embedded scripting layer (e.g. Lua) calls a stable engine API surface (move_agent, emit_event, broadcast, heal, ...). A 'tool' becomes a script + manifest: e.g. a teleport tool that calls move_agent on itself and emits 'vanished in a poof of smoke'. Existing built-in tools would be converted to the same form (major shift, highly extensible). Personas become installable bundles dropped into the world folder, e.g. gandalf/{SOUL.md, tools/cast_light.lua, influence_verbal.lua, water_magic.lua}. Key work: (1) design the engine API surface area, (2) pick/sandbox the scripting runtime, (3) tool manifest -> LLM tool schema, (4) convert existing tools, (5) bundle install/validation. Needs a spec before implementation.

Spec: specs/036-scriptable-agent-tools
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Full Spec Kit spec (specify -> clarify -> plan -> tasks) authored and linked to this task via spec-bridge:link before implementation starts
- [x] #2 Script tools are pure functions (args + read-only world view -> event batch + narration); every emitted event is validated against the InjectSocial whitelist — scripts cannot mutate state directly or invent event types
- [x] #3 v1 scope holds: instantaneous angel/expressive tools are scriptable; tick-simulated villager world verbs (hunt/forage/build...) remain native Go
- [x] #4 Persona bundle (SOUL/charter fragment + capabilities.json + tools/ manifests+scripts) installs by dropping a folder into the world dir, with boot-time validation: manifest schema, declared events subset-of whitelist, script parses, step/memory caps set
- [x] #5 Determinism preserved: scripts have no wall clock and no unseeded RNG; replaying a world containing scripted-tool events reproduces identical state hashes
- [x] #6 At least one existing metatron tool is re-expressed as a loadable bundle tool (dogfood) proving the manifest -> registry -> derive -> handler pipeline
- [x] #7 Spec phase: Setup
- [x] #8 Spec phase: Foundational (Blocking Prerequisites)
- [x] #9 Spec phase: User Story 1 — Declarative tool bundle end-to-end (Priority: P1) 🎯 MVP
- [x] #10 Spec phase: User Story 2 — Dogfood: built-in re-expressed as bundle (Priority: P2)
- [x] #11 Spec phase: User Story 3 — Scripted tools, sandboxed + deterministic (Priority: P3)
- [x] #12 Spec phase: User Story 4 — Persona bundles (Priority: P4)
- [x] #13 Spec phase: Polish & Cross-Cutting
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Grounded design sketch (from codebase exploration 2026-07-23, pre-spec):

**Core reframe — scripts emit events, they don't call an API.** The engine is event-sourced with two doors (sim.Loop.InjectIntent / InjectSocial) and a whitelist (internal/sim/loop.go:190 injectSocialWhitelist); replay determinism is load-bearing (State.Hash). So a tool script is a pure function: (args, read-only world view) -> event batch + narration. The engine validates the batch against the whitelist and lands it through the existing door. Teleport = script returns [entity_moved, broadcast("vanished in a poof of smoke")]. Safer and simpler than an imperative move_agent() API: no re-entrancy, trivially sandboxable, deterministic by construction, and ValidateToolCoverage (declared Events subset-of whitelist) extends unchanged.

**Scope cut — two species of tool; only one is scriptable in v1:**
- Instantaneous event-batch tools (metatron miracles/nudges, expressive/social) — already "build batch + inject" (sim.BuildMiracleBatch is the hand-coded prototype). v1 target.
- Tick-simulated world verbs (hunt = 900 ticks + spear durability, via InjectIntent + executor) — NOT scriptable in v1; scripting those means scripting the executor. Stay native Go.

**Runtime:** recommend starlark-go over gopher-lua — hermetic/deterministic by design (no I/O, no clock, no ambient randomness; built-in step limits). gopher-lua needs hand-stripping os/io/math.random. Either way: no wall clock, RNG seeded from event log only, hard step/memory caps. Decide in spec.

**Bundle shape:**
gandalf/
  SOUL.md              (persona/charter fragment)
  capabilities.json    (existing grant mechanism, internal/metatron/charter.go)
  tools/<name>/{tool.json, tool.star}   (manifest: name, description, params, declared events, charge cost)
Manifest maps onto the existing Tool struct so derive.go keeps generating LLM schemas + prompt gloss with no new machinery; loaded tools append to registry + roster. Precedent for world-dir file loading: charter.md / skills/*.md loaders (per-turn fresh).

**Phasing:**
1. Manifest-only declarative tools — parameterized macros over existing primitives (metatron.entity_moved / entity_removed / item_granted / time_snapped + broadcast/dream), zero script runtime. Proves load -> register -> schema -> handler pipeline.
2. Script runtime for conditional logic, sandboxed as above.
3. Widen the primitive event vocabulary one audited event at a time — e.g. heal does not exist today (needs only change via gameplay events); heal = new bounded needs_changed-style event = new store.Event type + State.Apply reducer arm + whitelist entry + charge cost. The "API surface" IS the whitelisted event vocabulary; grow it deliberately. Charge costs plug into the existing miracle economy so bundles can't be free-cast god mode.

**Key code anchors:** internal/tool/{registry.go,derive.go,roster.go,validate.go}; internal/toolloop/loop.go (Handler map); internal/sim/loop.go:190 (whitelist), internal/sim/miracles.go (primitives), internal/sim/toolcheck.go; internal/metatron/charter.go (charter/skills/capabilities loaders).
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Priority/ordering rationale (2026-07-23): placed HIGH, tail of the HIGH group (ordinal 76000) — above all mediums, below the four live-defect highs (TASK-84/40/86/87). Reason: very large surface area (internal/tool, toolloop, sim whitelist/reducers, metatron loaders, world-dir format) means longer drift = more staleness risk on a bigger issue; do it sooner rather than let the design sketch rot. Constitution tiering: architectural/cross-package -> Opus 4.8 implementation slices; full Spec Kit mandatory (already an AC).

Spec Kit complete 2026-07-24: specs/036-scriptable-agent-tools (specify -> clarify [4 Qs: persona per-tool failure, built-ins win collisions, invoker-scoped world view, failures free] -> plan -> tasks, 36 tasks / 7 phases). Constitution Check PASS pre-Phase-0 and post-Phase-1. Tier decisions (Principle V rubric): Phases 2-5 (new internal/bundle package, cross-package metatron/tool/sim seams, sandboxed starlark runtime, replay-determinism-critical code) -> Opus 4.8 spec-implementer; Phases 1, 6, 7 (dep setup, persona composition over proven seams, docs/polish) -> Sonnet, one-way escalation on gate failure. Runtime decision: go.starlark.net (hermetic+deterministic by construction, native step caps) over gopher-lua (research.md R1).

Phase 2 foundational landed (worktree commit d36da4d): manifest strict-decode + tool.Tool synthesis, 5-kind effect compiler byte-compatible with BuildMiracleBatch payloads, deterministic loader + B1-B4/T1-T7/C1-C2 ladder, sim.InjectableSocialEvent accessor (pure refactor of toolcheck). Known v1 limitation recorded: remove_entity/move_entity address living villagers by name only; structure/pile/terrain addressing grammar unspecified by contracts — deferred, needs follow-up card. NaN/Inf script-value guard deliberately deferred to Phase 5 executor.

Phase 3 US1 MVP landed (36eae00, rebased onto post-#41/#68 main): handler factory (InvocationContext{State,Tick,Invoker,Inject} — bundle importable by metatron, no reverse edge), daemon boot BundleSet wiring + BootReport logging, turn-assembly roster/handler merge (built-ins first), teleport fixture, TestBundleToolReplayByteIdentity green. AC#2 proven for declarative path (pure args+state -> whitelist-validated batch; scripts get same guarantee in Phase 5). AC#3 holds (expressive-only surface; world verbs untouched). Cosmetic note for Phase 6: capabilities.json naming a bundle tool emits a spurious 'unknown tool ignored' notice from loadManifest.

Correction: AC#2/#3 unticked — both assert script-tool properties; they become true in Phase 5 (US3), not with the declarative-only MVP.

Phases 4+5 landed (c184bf2 dogfood, 6167581 starlark runtime). AC#6: miracle_move bundle twin byte-identical to work_miracle{kind:move} (events+payloads+State.Hash+charge). AC#2/#3/#5 now true: apply(args, world) pure functions, frozen invoker-scoped view, sandbox proofs (step-cap deterministic abort, no clock/io/load, NaN/float rejection at script->Effect conversion), world.rand via new sim.BundleRand over rngAt; replay byte-identity incl. bundle-dir-deleted independence (FR-011). time_of_day pinned to sim's 22:00-06:00 night definition. Remaining: Phase 6 personas (Sonnet), Phase 7 polish.

Phase 6 landed (00dfd35): SOUL fragments into system prompt (charter->souls->skills order, byte-identical when absent), persona grant intersection (narrow-never-widen, commutative across personas, reaches roster+guidance+door), loadManifest cosmetic-notice fix, gandalf fixture + Scenario 6 integration test. Sonnet tier, no escalation needed.

Phase 7 polish landed (ead97e7 after rebase): quickstart 6/6 validated — Scenarios 1-2 manual with captured boot logs (clean load line; T3 and B3 rejection lines naming file+rule+value), 3-6 via the automated suite (named test-to-scenario mapping in PR body). Boot-perf: 32 bundles/256 tools discover+validate in 43-58ms. docs/bundles.md authoring guide linked from README. Branch task-85-scriptable-agent-tools pushed (7 commits, rebased on main); PR creation retrying against GitHub API outage.

Merged to main as 18a0376 (direct merge per user instruction, GitHub PR API outage; branch + worktree cleaned up, remote branch deleted). Full suite green on merged main except TestCatalogSweep — pre-existing, refiled as TASK-100. Spec 036 at 36/36 tasks, Done-eligible per spec-bridge. HOLDING at In Progress for the constitution-IV wiki re-pin: this change touches wiki-pinned sources (tool-registry, sim-loop, metatron-miracles, metatron, deterministic-rng, event-types, world-save-directory) and a concurrent session has in-flight edits to docs/wiki/ at root right now — wiki-update must run after their work lands, then TASK-85 -> Done + player-docs freshness check.

spec-bridge sync: Setup: 3/3 · Foundational (Blocking Prerequisites): 7/7 · User Story 1 — Declarative tool bundle end-to-end (Priority: P1) 🎯 MVP: 8/8 · User Story 2 — Dogfood: built-in re-expressed as bundle (Priority: P2): 3/3 · User Story 3 — Scripted tools, sandboxed + deterministic (Priority: P3): 7/7 · User Story 4 — Persona bundles (Priority: P4): 4/4 · Polish & Cross-Cutting: 4/4 — status In Progress → Done
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
All spec tasks complete (Setup: 3/3 · Foundational (Blocking Prerequisites): 7/7 · User Story 1 — Declarative tool bundle end-to-end (Priority: P1) 🎯 MVP: 8/8 · User Story 2 — Dogfood: built-in re-expressed as bundle (Priority: P2): 3/3 · User Story 3 — Scripted tools, sandboxed + deterministic (Priority: P3): 7/7 · User Story 4 — Persona bundles (Priority: P4): 4/4 · Polish & Cross-Cutting: 4/4). Derived Done by spec-bridge sync.
<!-- SECTION:FINAL_SUMMARY:END -->
