# Tasks: Loud Build Failure & Occupancy Tolerance

**Input**: Design documents from `/specs/038-loud-build-failure/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/agent-build-failed.md, quickstart.md

**Tests**: Spec explicitly requires regression tests (US1–US3 acceptance scenarios, FR-009, TASK-91 AC #4) — test tasks are included and written to fail first.

**Organization**: Tasks grouped by user story; all work lands on the single TASK-91 branch (`.worktrees/task-91`, branch `task-91-loud-build-failure`) per One-Task-One-PR.

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

**Purpose**: Worktree + baseline

- [X] T001 Create worktree `.worktrees/task-91` with branch `task-91-loud-build-failure` from fresh `origin/main`; verify `go test ./internal/sim/` is green at baseline

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The event type itself — every story emits, asserts, or renders it

- [X] T002 Define `BuildFailedPayload{Agent int, Goal string, Reason string}` and `wallOccupancyGraceTicks = 120` in internal/sim/agents.go (payload styled after `IntentRejectedPayload`, cognition.go:87; constant next to `buildWallTicks`, agents.go:649); define the two stable reason strings (`site no longer buildable`, `site blocked too long`) as constants
- [X] T003 Add `agent.build_failed` reducer case in internal/sim/state.go mirroring the `agent.intent_done` case (state.go:544-554): `a.Intent = nil`, `a.IdleSince = e.Tick`
- [X] T004 Subscribe `agent.build_failed` to planner re-arm in internal/mind/mind.go `absorb` (mind.go:218, alongside `agent.intent_done`) and add it to any sim-loop pass-through whitelist that carries `agent.intent_rejected` (locate via docs/wiki/event-types.md:206 reference)

**Checkpoint**: event type exists end-to-end (emit→reduce→re-arm) but nothing emits it yet; all existing tests still green

---

## Phase 3: User Story 1 — Loud, distinguishable build failure (Priority: P1) 🎯 MVP

**Goal**: Site-vanished mid-work build cancellation emits `agent.build_failed` (builder, goal, reason) instead of bare `agent.intent_done`; documented and rendered distinctly.

**Independent Test**: invalidate a build site mid-work in a sim test; assert the failure event with reason `site no longer buildable`, no bare `intent_done`, TUI digest renders a failure line.

### Tests for User Story 1 (write first, must fail)

- [X] T005 [US1] Add `TestWallBuildSiteVanishedFailsLoud` in internal/sim/wall_test.go: invalidate the wall's res tile site mid-work (e.g. structure placed on it); assert exactly one `agent.build_failed{goal: build_wall_stone, reason: site no longer buildable}`, zero `agent.built`, zero bare `agent.intent_done` for that intent, materials unspent
- [X] T006 [P] [US1] Add a non-wall case (e.g. `TestBuildSiteVanishedFailsLoudOven` in internal/sim/oven_test.go or wall_test.go): invalidate `buildSite(Target)` for a `build_oven` mid-work; same assertions with `goal: build_oven`

### Implementation for User Story 1

- [X] T007 [US1] In internal/sim/executor.go: for build goals in the validity switch (`build_fire/shelter/oven/chest/path` at 647-650, walls at 651-657), route site-invalid mid-work failure to emit `agent.build_failed` (+ memory event, see T010) instead of falling through to the bare `agent.intent_done` exit at 684-687; non-build goals keep the existing exit unchanged
- [X] T008 [P] [US1] Render `agent.build_failed` in internal/tui/digest.go (alongside :271) as `"<name>'s <goal> failed — <reason>"`, never "finished"
- [X] T009 [P] [US1] Document `agent.build_failed` in docs/wiki/event-types.md per contracts/agent-build-failed.md (new row + amend the `agent.intent_done` row at :115)

**Checkpoint**: US1 tests green; failed builds are loud and documented

---

## Phase 4: User Story 2 — Builder remembers the failure (Priority: P1)

**Goal**: every `agent.build_failed` is paired same-tick with a situated first-person failure memory (OriginAction) stating the build did NOT complete and why.

**Independent Test**: in the US1 failure tests, assert the paired `agent.memory_added` for the builder.

### Tests for User Story 2 (write first, must fail)

- [X] T010 [US2] Extend T005/T006 tests (internal/sim/wall_test.go): assert a same-tick `agent.memory_added` for the builder with `Origin: OriginAction`, salience `salShelter`, text stating the build did not complete and naming the cause

### Implementation for User Story 2

- [X] T011 [US2] In internal/sim/executor.go, emit the failure memory with every `agent.build_failed` via `situatedMemoryEvent(nextTick, i, salShelter, <place>, in.Reason, OriginAction, <text>)` (memory.go:189), text per data-model.md (first-person, names structure + cause); factor a small helper so all build-failure sites emit event+memory consistently

**Checkpoint**: failure belief is falsifiable — event + memory always travel together

---

## Phase 5: User Story 3 — Passerby no longer kills the build (Priority: P2)

**Goal**: wall guard split — occupancy ignored during work; at due tick an occupied res tile defers completion (never entomb); past `wallOccupancyGraceTicks` the build fails loudly.

**Independent Test**: passerby crosses res tile mid-work → wall completes; squatter parks on it → deferral then loud failure within grace bound.

### Tests for User Story 3 (write first, must fail)

- [X] T012 [US3] Rewrite `TestWallOccupancyGuard` in internal/sim/wall_test.go (:300) for new semantics: agent parked on res tile at due tick → no `agent.built` while occupied, completion defers, then exactly one `agent.build_failed{reason: site blocked too long}` + failure memory once `nextTick - WorkStart >= workDuration + wallOccupancyGraceTicks`; planks unspent, no wall, no bare `intent_done`
- [X] T013 [P] [US3] Add `TestWallBuildToleratesPasserby` in internal/sim/wall_test.go: second agent steps onto res tile mid-work and leaves before due tick → `agent.built`, wall stands with HP, planks spent, zero `agent.build_failed`
- [X] T014 [P] [US3] Add deferred-completion case in internal/sim/wall_test.go: occupant present at due tick, leaves within grace → wall completes on first clear tick (assert never entombed: no `agent.built` on any tick the tile was occupied)

### Implementation for User Story 3

- [X] T015 [US3] Split the wall guard in internal/sim/executor.go (651-657 + completion path 807-815): during work validity = `buildSite` only; at/after due tick — site invalid → loud fail (site no longer buildable); `agentAt(Res)` → return no event this tick (defer); `nextTick - WorkStart >= workDuration + wallOccupancyGraceTicks` while occupied → loud fail (site blocked too long); tile clear → existing `agent.built` completion

**Checkpoint**: all three stories independently green

---

## Phase 6: Polish & Cross-Cutting

- [X] T016 Update `TestReplayByteIdentityWallsAxesPaths` expected-event sets in internal/sim/whole_feature_test.go (:617) to include `agent.build_failed` + paired memory; verify byte-identical replay (FR-009)
- [X] T017 Run full quickstart.md validation: `go test ./...` green; confirm every invariant in data-model.md (one resolution per build intent, event+memory pairing, never entomb)
- [X] T018 [P] Reconcile any other docs/wiki notes that describe the old "invalid → intent_done" behavior (grep docs/wiki for `intent_done`); note in the implementer report which notes need `/grounding-wiki:wiki-update` re-pinning post-merge

---

## Dependencies & Execution Order

- **Phase 1 → Phase 2 → stories**: T002 blocks everything; T003/T004 block emission tests.
- **US1 (P1)**: after Phase 2. T005/T006 before T007; T008/T009 parallel with T007.
- **US2 (P1)**: extends US1's failure paths — T010 before T011; T011 touches the same executor helper as T007 (sequential with it).
- **US3 (P2)**: independent of US2 conceptually but edits the same executor switch — run after US1/US2 to avoid same-file conflicts. T012–T014 before T015.
- **Polish**: T016 after all emissions exist; T017 last.

### Parallel Opportunities

- T006 ∥ T005 (different test funcs); T008 ∥ T009 ∥ T007 (different files); T013 ∥ T014 (different test funcs); T018 ∥ T016.
- Single implementer expected — parallelism is optional batching, not required.

## Implementation Strategy

MVP = Phases 1–3 (US1): loud failure alone breaks the phantom-belief loop's
observability gap. US2 makes it falsifiable in-mind; US3 removes the dominant
failure trigger. Deliver sequentially on the one TASK-91 branch; commit per
task or logical group; stop at each checkpoint to validate.
