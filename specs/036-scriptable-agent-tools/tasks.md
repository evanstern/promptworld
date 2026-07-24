# Tasks: Scriptable Agent Tools — Pluggable Bundle-Defined Tools

**Input**: Design documents from `/specs/036-scriptable-agent-tools/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/ (all present)

**Tests**: INCLUDED — the spec's success criteria (SC-002/003/006) and constitution demand
determinism, sandbox, and equivalence proofs; tests ride alongside code per project convention.

**Organization**: grouped by user story; each phase is an independently testable increment.

**Model tiers (constitution V)**: Phases 2–5 are cross-package/architectural + determinism-critical
→ **Opus 4.8** via `spec-implementer`. Phase 1 setup and Phase 7 polish/doc tasks → Sonnet.
Phase 6 (persona composition over proven seams) → Sonnet, escalate on gate failure.

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

**Purpose**: dependency + package skeleton

- [x] T001 Add `go.starlark.net` as a direct dependency (`go get go.starlark.net@latest && go mod tidy`) in go.mod
- [x] T002 [P] Create `internal/bundle/` package skeleton (doc.go with package comment stating the invariants: effects-not-events, boot-frozen, whitelist-subset) and `internal/bundle/testdata/` fixture root
- [x] T003 [P] Add `BundlesDir()` accessor (`<dir>/bundles`) to internal/world/world.go with unit coverage in internal/world/world_test.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: manifest model, effect compiler, loader/validator, BundleSet — everything every story needs

**⚠️ CRITICAL**: no user story work until this phase completes

- [x] T004 Implement manifest parsing + `tool.Tool` synthesis in internal/bundle/manifest.go per contracts/bundle-manifest.md: strict JSON decode, param mapping to `tool.Param` kinds, name/description/charges/limits validation mirroring internal/tool/validate.go rules
- [x] T005 [P] Unit tests for manifest parsing in internal/bundle/manifest_test.go: valid manifest, unknown keys rejected, name≠folder rejected, param rule violations, charges/limits bounds (rules T1/T2/T7 of contracts/boot-validation.md)
- [x] T006 Implement the effect vocabulary + effect→`store.Event` compiler in internal/bundle/effects.go per research.md R4 and data-model.md: five kinds (`move_entity`, `remove_entity`, `grant_item`, `snap_time`, `narrate` with recipient expansion mirroring internal/metatron/miracle_batch.go:83-88), `{args.x}`/`{invoker}` template substitution, batch caps (≤32 events, text ≤500 bytes), NaN/Inf rejection, declared-events subset check
- [x] T007 [P] Unit tests for the effect compiler in internal/bundle/effects_test.go: each kind compiles to the correct event type+payload, template substitution, recipient expansion (`target`/`all_living`/named), cap violations, undeclared event type rejected, empty-batch narration-only allowed
- [x] T008 Implement bundle discovery + `BundleSet` in internal/bundle/load.go per contracts/boot-validation.md: deterministic bytewise ordering, ≤16 tools/bundle, dotfile/unknown-file handling, `Roster()` (synthesized tools in deterministic order), `BootReport` with rule ids + file paths, collision rules C1 (built-in wins via `tool.Lookup`) and C2 (first-loaded wins)
- [x] T009 Implement the boot validation ladder in internal/bundle/validate.go: B1–B4 bundle-level (SOUL cap 4000 chars, capabilities.json parses per `manifestDoc` schema), T1–T7 tool-level (events ⊆ `sim` whitelist — export or mirror the whitelist check per internal/sim/toolcheck.go:62-67; script parse via `starlark.SourceProgram` + `apply` presence), per-tool skip vs bundle-reject scoping per clarification #1
- [x] T010 [P] Unit tests for load+validate in internal/bundle/load_test.go using fixtures under internal/bundle/testdata/: valid bundle loads; off-whitelist event skips tool (T3); malformed capabilities.json rejects bundle (B3); oversized SOUL rejects bundle (B2); collision C1/C2 warnings; deterministic ordering; BootReport messages name file + problem (SC-005)

**Checkpoint**: `go test ./internal/bundle/...` green; loader/compiler proven in isolation

---

## Phase 3: User Story 1 — Declarative tool bundle end-to-end (Priority: P1) 🎯 MVP

**Goal**: drop a manifest-only bundle folder → tool on the metatron roster → invocation lands validated events + narration

**Independent Test**: quickstart.md Scenarios 1–2

- [ ] T011 [US1] Implement `BundleSet.Handlers()` handler factory in internal/bundle/handlers.go: validated args → template expansion → effect compile → declared-events check → `sim.Loop.InjectSocial`; failures return `toolloop.Outcome{Verdict: rejected_*, ResultForModel: <specific reason>}` per internal/toolloop/loop.go conventions (no charge spent on failure — reducer already guarantees)
- [ ] T012 [US1] Add `PromptGloss` fallback to `MetatronToolGuidance` in internal/tool/derive.go (tools absent from `metatronToolDesc` render their PromptGloss + param surface instead of empty description), with unit coverage in internal/tool/derive_test.go
- [ ] T013 [US1] Load + validate the `BundleSet` at daemon boot (alongside `tool.Validate()`/`sim.ValidateToolCoverage()` call sites), store it on the daemon, and log every BootReport entry; world with zero/absent `bundles/` boots unchanged
- [ ] T014 [US1] Merge bundle surface into the metatron turn assembly in internal/metatron/turn.go + toolcalls.go: roster = `grantedRoster(grant) + bundleSet.Roster()` (grant-filtered), handlers = `turnHandlers(d) + bundleSet.Handlers(...)`, guidance via T012 fallback; converse/read tools unaffected
- [ ] T015 [P] [US1] Create the teleport fixture bundle in internal/bundle/testdata/worlds/declarative/bundles/demo/tools/teleport/tool.json (move_entity + narrate "vanished in a poof of smoke" to all_living) per quickstart Scenario 1
- [ ] T016 [US1] Integration test in internal/metatron/bundle_integration_test.go: boot a test world with the teleport fixture → roster contains `teleport` with derived schema; invoking the handler moves the villager, adds narration memories to all living agents, event log contains only declared types
- [ ] T017 [US1] Replay byte-identity test `TestBundleToolReplayByteIdentity` in internal/sim (or internal/bundle) following internal/sim/miracles_test.go:398 pattern: live-apply a bundle-tool batch vs replay → identical `State.Hash()` (SC-003)
- [ ] T018 [P] [US1] Boot-rejection integration test: world with an off-whitelist fixture bundle boots, BootReport error names file+rule, valid sibling tool still on roster (quickstart Scenario 2, SC-005)

**Checkpoint**: MVP — declarative bundles fully functional and provably deterministic

---

## Phase 4: User Story 2 — Dogfood: built-in re-expressed as bundle (Priority: P2)

**Goal**: one existing metatron tool ships as a bundle twin with equivalent observable behavior (AC #6)

**Independent Test**: quickstart.md Scenario 5

- [ ] T019 [P] [US2] Author the dogfood bundle in examples/bundles/dogfood-move/tools/miracle_move/ (tool.json re-expressing the `work_miracle` move kind: `move_entity` + the exact perception memory text from internal/metatron/miracle_batch.go:40-43), tracked in-repo as the shipped example
- [ ] T020 [US2] Equivalence test in internal/metatron/dogfood_test.go: identical-seed worlds — built-in `work_miracle{kind:move}` vs bundle `miracle_move` with same args → equivalent events, narration memories, and charge deduction (SC-004)
- [ ] T021 [US2] Collision test in internal/bundle/load_test.go: install a bundle tool named `work_miracle` → C1 boot warning, built-in wins, exactly one roster entry (clarification #2)

**Checkpoint**: bundle format proven sufficient for real shipped tools

---

## Phase 5: User Story 3 — Scripted tools, sandboxed + deterministic (Priority: P3)

**Goal**: `tool.star` scripts compute effect batches from args + invoker-scoped world view under hard caps

**Independent Test**: quickstart.md Scenarios 3–4

- [ ] T022 [US3] Implement the invoker-scoped world view in internal/bundle/worldview.go per contracts/script-api.md: `tick`, `time_of_day`, `map_width/height`, `agents()`, `agent(name)`, `rand(purpose,index)` backed by the internal/sim/rng.go:11-16 `rngAt` pattern with purpose `"bundle:<tool>:<purpose>"`; frozen starlark values; nothing else exposed
- [ ] T023 [US3] Implement the Starlark executor in internal/bundle/script.go: program compiled once at boot, per-invocation `starlark.Thread` with step cap from manifest limits (default 100k, ceiling 1M), recursion off, no `load()`, predeclared `args` (frozen dict) + `world`; `apply()` return value fed to the T006 compiler; deterministic abort → descriptive error
- [ ] T024 [US3] Wire script-mode tools through the handler factory in internal/bundle/handlers.go (script path replaces template expansion; identical downstream compile/validate/inject) and extend boot validation T6 to compile-check the program
- [ ] T025 [P] [US3] Sandbox + cap unit tests in internal/bundle/script_test.go: step-cap exhaustion aborts deterministically with no state change; no time/io/net/module access (probe each); undeclared event kind from script rejected by compiler; malformed return shapes rejected; `fail()` propagates as invocation failure (SC-006)
- [ ] T026 [P] [US3] Create the cast_light scripted fixture in internal/bundle/testdata/worlds/scripted/bundles/demo/tools/cast_light/{tool.json,tool.star} branching on `world.time_of_day` per contracts/script-api.md example
- [ ] T027 [US3] Integration test in internal/metatron/bundle_integration_test.go: cast_light under night vs day produces the branch-correct narration; only declared events land (quickstart Scenario 3)
- [ ] T028 [US3] Scripted replay byte-identity test (variant of T017) exercising `world.rand`: live vs replay hashes identical; second test proving replay is bundle-independent — delete the bundle dir, replay still reproduces the hash (SC-003, FR-011)

**Checkpoint**: scripting runtime live, sandboxed, provably deterministic

---

## Phase 6: User Story 4 — Persona bundles (Priority: P4)

**Goal**: `bundles/<name>/{SOUL.md, capabilities.json, tools/}` installs identity + grants + tools as one unit

**Independent Test**: quickstart.md Scenario 6

- [ ] T029 [US4] Load SOUL.md fragments into `BundleSet.SoulFragments()` (cap 4000 chars, load order) in internal/bundle/load.go and append them to the metatron system prompt after charter in internal/metatron/turn.go (`turnSystemPrompt` seam), with prompt-content assertion in internal/metatron/turn_test.go
- [ ] T030 [US4] Implement persona `capabilities.json` intersection narrowing in internal/bundle/load.go + application after the world-level grant in internal/metatron/turn.go (persona can narrow, never widen — reuse `grantSet` semantics from internal/metatron/charter.go:144-149), with unit tests covering narrow/no-widen/absent-file cases
- [ ] T031 [P] [US4] Create the persona fixture in internal/bundle/testdata/worlds/persona/bundles/gandalf/ (SOUL.md, capabilities.json narrowing miracle_kinds, two tools — one valid, one with a broken manifest)
- [ ] T032 [US4] Persona integration test in internal/metatron/bundle_integration_test.go: SOUL fragment present in system prompt; narrowed kind absent from work_miracle enum; valid tool on roster; broken tool skipped with T-rule BootReport while SOUL + grant + sibling stay active (clarification #1, quickstart Scenario 6)

**Checkpoint**: all four user stories independently functional

---

## Phase 7: Polish & Cross-Cutting

- [ ] T033 Run the full quickstart.md validation (Scenarios 1–6) against a scratch world; record outcomes in the PR description
- [ ] T034 [P] `go vet ./...` + `go test ./... -count=1` full-suite green; boot-time validation of a 32-bundle world <1s sanity check (plan Technical Context)
- [ ] T035 [P] Author docs/bundles.md — bundle authoring guide distilled from contracts/ (manifest reference, script API, validation errors), linked from README.md
- [ ] T036 Reconcile board: tick TASK-85 ACs #2–#6 with evidence, record tier choices per phase, `spec-bridge:sync`

**Post-merge obligation (constitution IV, not a PR gate)**: `/grounding-wiki:wiki-update` re-pin
(tool-registry, sim-loop, metatron-miracles, metatron, deterministic-rng, event-types,
world-save-directory notes) then `player-docs` freshness check.

---

## Dependencies & Execution Order

- **Phase 1 → Phase 2 → Phase 3 (MVP)**: strictly ordered; T004/T006 before T008/T009 (loader validates via compiler); T011–T014 sequential (handler → derive fallback → boot → turn merge share seams), T015/T018 parallel to late Phase-3 work
- **Phase 4 (US2)**: needs Phase 3 (declarative pipeline + boot). Independent of Phases 5–6
- **Phase 5 (US3)**: needs Phase 2 (compiler) + Phase 3's handler seam (T011); independent of Phase 4
- **Phase 6 (US4)**: needs Phase 3 (load/turn seams); tools inside personas may be declarative only → independent of Phase 5, but scripted persona tools need Phase 5
- **Phase 7**: after all stories

### Parallel opportunities

- T002/T003 in parallel after T001
- T005, T007, T010 (tests) parallel with the next implementation task in their phase
- Phases 4 and 5 can run in parallel after Phase 3 (different packages/files except bundle_integration_test.go — coordinate T020/T027 merges)
- T015, T019, T026, T031 (fixtures) are always parallelizable

## Implementation Strategy

MVP-first: Phases 1–3 deliver the complete declarative pipeline (US1) and the determinism proof —
stop-and-validate point. Then US2 (dogfood, cheap, high evidence value), US3 (the runtime),
US4 (packaging). Commit per task or logical group on the single `task-85` branch; one PR.
