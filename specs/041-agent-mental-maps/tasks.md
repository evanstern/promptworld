# Tasks: Per-Agent Mental Maps

**Input**: Design documents from `/specs/041-agent-mental-maps/`

**Prerequisites**: plan.md, spec.md (clarified), research.md (D1–D9), data-model.md,
contracts/knowledge-events.md, quickstart.md

**Tests**: included — the spec's success criteria demand test evidence (SC-001..SC-007 are
test-shaped) and the repo's gates are test-enforced. Tests land in the same slice as the
code they prove.

**Organization**: grouped by user story; every story phase is independently testable.
Implementation is delegated slice-by-slice to `spec-implementer` (**Opus 4.8** — recorded
tier: cross-package, doctrine-adjacent, determinism-sensitive).

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

- [x] T001 Baseline: in `.worktrees/task-96`, run `go test ./...` and record the pre-feature green baseline (plus known-red list, if any) in the PR notes; confirm branch is rebased on current `origin/main`

## Phase 2: Foundational (blocking all stories)

**⚠️ No user-story work until this phase is complete.**

- [x] T002 Create `internal/sim/mentalmap.go`: `MentalMap{Explored string; Facts []PlaceFact}` + `PlaceFact{Kind,X,Y,Seen,Provenance,Source,Detail}` per data-model.md; base64 W×H bitmap codec (row-major); canonical (Kind,X,Y) fact ordering with binary-search upsert/remove; freshness-horizon constants (volatile vs durable kinds) and `fresh(fact, now)`; accessors `Explored(x,y)`, `MarkExplored(x,y,radius)` (map-bounds clipped), `KnownFresh(kind, now)` iterator
- [x] T003 [P] Unit tests in `internal/sim/mentalmap_test.go`: codec round-trip, upsert ordering determinism (shuffled insert → identical bytes), freshness horizons, bounds clipping
- [x] T004 Add `Map *MentalMap json:"map,omitempty"` to `Agent` in `internal/sim/agents.go` (after `Journal`); snapshot byte-stability twin test `TestMapOmitemptyStable` in `internal/sim/state_test.go` (pattern: `TestAxesOmitemptyStable`) proving pre-feature snapshots round-trip byte-identically
- [x] T005 Derived explored-bit bookkeeping in `internal/sim/state.go` reducer arms: on position-changing events (`agent.moved`, genesis-adjacent arms, miracle move) mark mover's surroundings explored within perception radius (reuse `witnessRadius`); research D2 — silent, no new events
- [x] T006 Genesis seeding in `internal/sim/state.go` `NewState`: each agent's spawn surroundings marked explored (perception radius); worlds start with zero structures so no facts granted (research D7)
- [x] T007 Perception sweep + `agent.saw` event: emitter in `internal/sim/executor.go` `stepEvents` (per-beat, awake agents: diff ground truth within perception radius against agent's map → `SawPayload{Agent, Facts}` for new/changed facts only, gated kinds incl. resource tiles/dens/piles per data-model.md); payload struct + reducer arm (upsert witnessed) in `internal/sim/state.go`
- [x] T008 [P] Gates for `agent.saw`: digest registry entry + catalog fixture row in `internal/tui/digest.go` / `internal/tui/digest_test.go`; backticked row in `docs/wiki/event-types.md`; `TestCatalogSweep` green
- [x] T009 Format bump + migration: `FormatVersion` 3→4 in `internal/world/world.go`; v3→v4 transform in `internal/sim/migrate.go` granting each living agent explored area around current position + witnessed facts for all current structures and piles (research D7); migration test (v3 fixture world loads, agents hold seeded knowledge)
- [x] T010 Replay determinism harness: mental-map replay-byte-identical test (pattern: `internal/mind/replay_test.go` `TestJournalAndSituatedReplayByteIdentical`) covering moved/saw sequences; extend `internal/sim/sim_test.go` determinism tests to assert map bytes equal across same-seed runs

**Checkpoint**: maps exist, fill from perception, survive save/replay/migration byte-identically.

---

## Phase 3: User Story 1 — Agents act only on places they know (P1) 🎯 MVP

**Goal**: knowledge-gated target resolution with honest "you know of none" failures; full
reflex parity.

**Independent test** (from spec): fire far outside a villager's known area → warmth verb
fails "unknown to you"; after witnessing the fire, same verb resolves to it; known-far
beats unknown-near.

- [x] T011 [US1] Map-aware matching helpers in `internal/sim/path.go`: `nearestKnown(s, m, agentIdx, fromX, fromY, kind, now)` wrapping `nearest`/`nearestAdjacentTo` with fresh-fact predicates (deterministic BFS order preserved; terrain passability untouched — research D3)
- [x] T012 [US1] Gate `goalResolvers` in `internal/sim/policy.go`: forage/chop/quarry/collect_water/hunt(den)/refuel_fire/cook/goto_warmth/bathe/pick_up/deposit/withdraw/demolish/repair resolve via `nearestKnown`; build-site resolvers stay terrain-based (building needs no prior knowledge of the empty spot); resolver errors use knowledge phrasing `"you know of no <kind>"` distinct from `"no <kind> reachable"` (contracts §4, FR-004)
- [x] T013 [US1] Gate `talk_to`/`seek` resolver in `internal/sim/policy.go`: target position from acting agent's map (last-known fact or currently-perceived), not live coordinates; existing `GuardTargetPresent` landing guard covers misses
- [x] T014 [US1] Reflex parity in `internal/sim/policy.go` `decideIntent`: every rung's target lookup goes through the same map predicates (research D3; clarify Q2 — no omniscient fallback)
- [x] T015 [US1] US1 tests in `internal/sim/policy_test.go` (or sibling): unknown-target rejection wording; known-far-beats-unknown-near (spec US1 scenario 3); witness-then-resolve flow; reflex rung parity (agent with empty map does not resolve to unseen food source); `ValidateToolCoverage` still green
- [x] T016 [US1] Rejection-string plumbing check in `internal/mind/handlers.go`: gated-verb rejection reaches the model verbatim via `VerdictRejectedGate` and lands in the next prompt cycle (extend existing handler test)

**Checkpoint**: MVP — the omniscient resolver is gone for gated verbs; epistemics are honest.

---

## Phase 4: User Story 2 — Prompt renders only known places (P2)

**Goal**: the write path is visible: prompts differ by history; all known structures render;
first-6 cap retired.

**Independent test**: traveler vs homebody prompts diverge; 7th structure built in presence
appears.

- [x] T017 [US2] Known-places section in `internal/mind/prompt.go` `userPrompt`: replace the `Village:` line with map-rendered section per contracts §3 — all fresh known structures individually with provenance phrasing (witnessed/told/revealed), resource kinds grouped with count + nearest, one-line unexplored orientation (dominant frontier direction), explicit empty-state line; delete the `parts[:6]` truncation
- [x] T018 [US2] Prompt tests in `internal/mind/prompt_test.go`: divergent maps → divergent sections; >6 known structures all render; told-provenance phrasing; empty-state line; fully-explored map omits the unknown-land line (SC-002)

**Checkpoint**: two villagers see different worlds; watchers can tell.

---

## Phase 5: User Story 3 — Stale memories corrected by reality (P3)

**Goal**: perception removes vanished facts, observably.

**Independent test**: fire learned → burns out while away → agent seeks it, arrives, map
corrects, next plan uses corrected knowledge.

- [x] T019 [US3] `agent.map_corrected` event: emitter in the T007 perception sweep (remembered fresh facts within perception radius absent from ground truth → `MapCorrectedPayload{Agent, Gone}` carrying facts *as remembered*); reducer arm removes facts + stamps situated memory ("The fire … was cold and dead when you looked.", Origin witness) in `internal/sim/state.go` / `internal/sim/memory.go`
- [x] T020 [P] [US3] Gates for `agent.map_corrected`: digest row + fixture (`internal/tui/digest.go`, `digest_test.go`); `docs/wiki/event-types.md` row; `chronicleNote` grammar line in `internal/mind/narrate.go`
- [x] T021 [US3] Absorb trigger in `internal/mind/mind.go`: `agent.map_corrected` re-arms the planner when a removed fact matches the agent's current intent target (contracts §1)
- [x] T022 [US3] US3 tests in `internal/sim/`: SC-005 flow (structure removed while away → arrival emits correction, fact gone, memory stamped); stale-plan-step interaction (plan step whose target was corrected away fails via existing `agent.plan_expired`, no omniscient re-resolve)

**Checkpoint**: believe–act–discover loop closed.

---

## Phase 6: User Story 4 — Deliberate search of the unknown (P4)

**Goal**: "I know of none" becomes a plan: frontier-directed exploration with honest
exhaustion.

**Independent test**: mostly-unknown map + target only in unexplored land → search reaches
unknown space, map grows, target found, then acted on.

- [x] T023 [US4] Frontier helper in `internal/sim/mentalmap.go` (+`path.go` glue): nearest tile that is explored ∧ passable ∧ adjacent-to-unexplored, via existing deterministic BFS predicate (research D4)
- [x] T024 [US4] `search` goal resolver + duration entry in `internal/sim/policy.go`: instant wander-class goal targeting the frontier tile; exhaustion error `"nothing left unexplored"` when no reachable frontier (contracts §4)
- [x] T025 [US4] `search` tool in the villager roster (`internal/tool/` roster definition, model-facing description per contracts §2); goal-door coverage: `ValidateToolCoverage` green (`internal/sim/toolcheck.go`)
- [x] T026 [US4] Reflex fallback in `internal/sim/policy.go`: get-food rung falls back to `search` when no food source is known and map has frontier (keeps FR-013 parity without omniscience)
- [x] T027 [US4] US4 tests in `internal/sim/`: frontier target properties; repeated search monotonically grows explored coverage and terminates (SC-004); fully-explored exhaustion; wander untouched (FR-010)

**Checkpoint**: ignorance is a solvable problem in-world.

---

## Phase 7: User Story 5 — Spatial knowledge spreads through talk (P5)

**Goal**: automatic bounded place-fact exchange on every completed talk, told-provenance.

**Independent test**: A knows fire, B doesn't → talk → B targets the fire; B's record shows
told-by-A.

- [x] T028 [US5] Talk-transfer sidecar in `internal/sim/executor.go` `talkEvents` (beside `TellableFor`): deterministic selection of ≤2 facts per direction the other lacks-or-holds-staler (freshest, then nearest-to-listener, then coordinate order); emit `social.place_told` per direction with teller's `Seen` and `Source` (research D5)
- [x] T029 [US5] Reducer arm for `social.place_told` in `internal/sim/social.go` (or state.go): upsert absent-or-staler only, provenance told; situated memories both sides ("Told Birch about…"/"Birch told you of…") in `internal/sim/memory.go`
- [x] T030 [P] [US5] Gates for `social.place_told`: digest row + fixture; `docs/wiki/event-types.md` row; `chronicleNote` grammar in `internal/mind/narrate.md`-correct file (`internal/mind/narrate.go`)
- [x] T031 [US5] US5 tests in `internal/sim/`: SC-006 flow (transfer ≤2, told provenance, teller's Seen tick, staler-never-overwrites-fresher); B acts on told fact end-to-end at resolver level; secondhand-of-secondhand keeps original Seen

**Checkpoint**: directions exist; scouts matter.

---

## Phase 8: Polish & Cross-Cutting

- [ ] T032 Metatron reveal channel (FR-014): optional place grant on `send_vision` → `metatron.place_revealed` event; payload + reducer arm (upsert revealed + Origin-omen memory); `injectSocialWhitelist` entry in `internal/sim/loop.go`; Metatron tool arg in `internal/tool/registry.go`; digest row + fixture + `docs/wiki/event-types.md` row + chronicle grammar
- [ ] T033 [P] Dead-agent hygiene: death excludes the agent's map from resolver/prompt/talk-transfer read paths (data-model invariant); test in `internal/sim/`
- [ ] T034 Full-suite gates: `go test ./...` and `go test -race ./...` green; `go test ./e2e/ -run TestDeterminism_FullBinary` green (SC-003 end-to-end)
- [ ] T035 Viability soak (SC-007): `TestVillageSurvivesTwoDays` and `internal/sim/whole_feature_test.go` green with gating + seeding; tune freshness horizons/search cadence if starvation regresses; record chosen constants in research.md D6 addendum
- [ ] T036 Quickstart validation sweep: execute quickstart.md fast loop + migration check; live smoke optional; note outcomes in PR description
- [ ] T037 gofmt/vet sweep over touched files; ensure new exported identifiers have doc comments in repo voice
- [ ] T038 Wiki re-pin (constitution IV): `/grounding-wiki:wiki-update` covering touched sources (reflex-policy, agent-mind, cognition, tool-registry, event-types, sim-state-reducer, snapshots, social-fabric, tool-loop, worldmap-generation as applicable)
- [ ] T039 Player docs freshness: `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`; regenerate `docs/player/` via the player-docs skill if stale
- [ ] T040 Board close-out: `spec-bridge:sync` moves TASK-96 per artifacts; tick board ACs #3–#7 with evidence; PR opened from `.worktrees/task-96` (one task, one PR)

## Dependencies

- Phase 2 blocks all stories. Story order US1 → US2 → US3 → US4 → US5 is the priority
  order; US2+ each depend only on Phase 2 plus (for US3) the T007 sweep and (for US5) the
  fact model — after Phase 2 completes, US2 can proceed in parallel with late US1 tasks if
  needed, but sequential-by-priority is the default posture.
- T032 (Metatron) depends on the fact model only; safe any time after Phase 2.
- T038–T040 strictly last (they gate Done).

## Parallel opportunities

- Within Phase 2: T003 ∥ T004 after T002; T008 ∥ T009/T010 after T007.
- Gate tasks marked [P] (T008, T020, T030) parallel their story's reducer work.
- US4 and US5 touch disjoint seams (path/policy/tool vs executor/social) — parallelizable
  across two implementer runs if desired.

## Implementation strategy

MVP = Phase 2 + Phase 3 (US1): omniscience is gone, failures are honest, maps fill and
persist — shippable and observable on its own. Each subsequent story is an independent
increment landing as commits on `task-96-agent-mental-maps`; spec-bridge sync after each
phase keeps the board honest. Delegation: one `spec-implementer` (Opus 4.8) run per phase
(or per story), with the phase's tasks + relevant design-doc excerpts as the slice brief.
