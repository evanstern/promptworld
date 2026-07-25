# Tasks: Per-Turn Context Grounding

**Input**: Design documents from `/specs/043-context-grounding/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/context-blocks.md, quickstart.md

**Tests**: included — the spec's success criteria are test-shaped (SC-002/003/005 are deterministic assertions) and the constitution's gates require artifact evidence.

**Organization**: grouped by user story; delivery order US5 → US1 → US2 → US3 → US4 per spec priorities. All stories are commits on ONE branch (`.worktrees/task-105`, board TASK-105) merging in ONE PR (constitution Principle II). Model tiers per plan Constitution Check: US5 = Sonnet; US1-US4 = Opus 4.8 (reducer + `internal/mind` orchestration, doctrine-adjacent).

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

- [x] T001 Create worktree `.worktrees/task-105` (branch `task-105-context-grounding` from origin/main); confirm `go build ./...` green at fork point

---

## Phase 2: Foundational (blocking prerequisites)

**Purpose**: the assembler skeleton every story renders through, and the telemetry fields every story reports through.

- [x] T002 Block assembler skeleton in `internal/mind/context.go`: `contextBlock{Name, Priority, Render}` registry, fixed contract order, per-block byte measurement, `approxTokens = bytes/4` helper, `contextBudgetTokens = 2000` const with TASK-107 tuning-manifest dial comment, whole-block drop ascending priority (contracts/context-blocks.md is normative)
- [x] T003 Wrap existing `userPrompt` content into blocks (`frame`, `needs`, `inventory`, `known_places`/`nearby`, `social_law`, `memories`, `memories_serendipity`) in `internal/mind/prompt.go` with byte-identical output when no new blocks render and budget is not exceeded — assert via existing `internal/mind/shadow_test.go` invariant plus a new golden-equality test in `internal/mind/context_test.go`
- [x] T004 Extend `CogThoughtPayload` with `PromptBytes`, `BlockBytes`, `DroppedBlocks` in `internal/sim/cognition.go`; stamp them where thoughts are emitted in `internal/mind/telemetry.go`; reducer stays no-op (`internal/sim/state.go`)
- [x] T005 Assembler unit tests in `internal/mind/context_test.go`: deterministic assembly (same state ⇒ same bytes), drop order under a shrunk test budget, never-dropped blocks survive extreme budgets, telemetry fields populated

**Checkpoint**: prompt output unchanged for existing worlds; sizes observable per thought.

---

## Phase 3: US5 — Operators can see exactly what an agent knew (P1) 🎯 co-MVP

**Goal**: durable, freshness-gated context inventory. **Independent test**: read the note, capture a live decision context, verify block-for-block (SC-001).

- [x] T006 [US5] Write `docs/wiki/decision-context.md`: every block from contracts/context-blocks.md (source of truth, appearance conditions, caps, drop priority) plus deliberate absences (LastGoal history beyond the ring, full event log, other agents' private state); pin to current commit per wiki note conventions; add to `docs/wiki/INDEX.md`
- [x] T007 [US5] Contract-vs-capture check per quickstart §Contract-check: capture one real `cog.thought` `BlockBytes` + prompt text from a scratch world, verify against the note, record the capture in `specs/043-context-grounding/` as evidence (SC-001)

**Checkpoint**: US5 shippable alone (documents present-state even before US1-US4 blocks land; note updated as they do).

---

## Phase 4: US1 — An agent knows what it was just doing (P1) 🎯 co-MVP

**Goal**: reducer-maintained intent ring rendered as the `self_history` block. **Independent test**: per-source context capture + thrash-window replay (SC-002, SC-004).

- [x] T008 [US1] Add `IntentRecord` type + `Agent.IntentLog` ring (cap 8) in `internal/sim/agents.go` per data-model.md
- [x] T009 [US1] Reducer arms in `internal/sim/state.go`: `agent.intent_set` appends {Goal, Source, Reason, Tick}; `agent.intent_done`/`agent.build_failed` stamp outcomes; `agent.intent_rejected` appends-and-closes; `agent.plan_expired` stamps/append `expired`
- [x] T010 [P] [US1] Ring transition tests in `internal/sim/intentlog_test.go`: all five arms, override-in-quick-succession ordering, ring wraparound at cap, rejected-never-landed shape
- [x] T011 [US1] `self_history` block renderer in `internal/mind/context.go` + wire into `internal/mind/prompt.go`: last ≤4 records newest-first, sources in plain words (planner→"you decided", reflex→"instinct", plan→"your plan"; unknown→"unknown", never invented), outcomes, "no prior activity" empty state (contract block 3)
- [x] T012 [US1] Renderer tests in `internal/mind/context_test.go`: each source wording, empty state, alternation visible across ≥3 records (SC-002)
- [x] T013 [US1] Thrash-window context replay per quickstart §SC-004: reconstruct Sage's episode (world-01 ticks 265,411–266,631), assert assembled context shows the instinct redirection + alternation; record result in the spec dir

**Checkpoint**: MVP complete (US5+US1) — the model can see what it was just doing and operators can audit it.

---

## Phase 5: US2 — An agent feels which way its needs are moving (P2)

**Goal**: trajectory arrows on the needs block. **Independent test**: deterministic needs-movement scenario (SC-003).

- [x] T014 [US2] Add `Agent.NeedsAnchor`/`NeedsAnchorTick` in `internal/sim/agents.go`; refresh in the `agent.needs_changed` reducer arm in `internal/sim/state.go` when `tick - anchorTick >= trajectoryWindowTicks` (const 1800, tuning-dial comment)
- [x] T015 [P] [US2] Anchor tests in `internal/sim/needsanchor_test.go`: window-edge refresh, unset-anchor first window, sleep-spanning window reflects overnight fall (spec edge case)
- [x] T016 [US2] Trajectory rendering in the `needs` block (`internal/mind/context.go`): rising/falling/steady from `current − anchor` with deadband ±10; steady never flickers (SC-003); tests in `internal/mind/context_test.go`

**Checkpoint**: "warmth 43 and falling" distinguishable from "warmth 43 and rising".

---

## Phase 6: US3 — An agent continues its plan instead of restarting it (P3)

**Goal**: `plan_echo` block. **Independent test**: standing-plan capture, echo disappears on clear (FR-005).

- [x] T017 [US3] `plan_echo` renderer in `internal/mind/context.go` reading `Agent.Plan` (`internal/sim/plan.go` PlanStep): remaining steps in order, head marked next, guards + `Until` deadlines in plain words; omitted when no plan (contract block 5)
- [x] T018 [US3] Tests in `internal/mind/context_test.go`: active-plan echo content, no-stale-echo after `agent.plan_expired`/completion, plan end visible in self-history at next thought (via US1 ring)

**Checkpoint**: plans are commitments the deciding mind can see.

---

## Phase 7: US4 — What an agent remembers is chosen for the moment (P4)

**Goal**: relevance-fed memories + journal excerpts under the budget. **Independent test**: planted-memory selection + degraded mode (SC-006), budget fit (SC-005).

- [x] T019 [US4] Journal term-match selection per research R5: exported excerpt helper in `internal/sim/journal.go` (≤2 entries via `SearchJournal`, query terms = two worst needs + active/last goal, ≤300-rune excerpts); `journal` block renderer in `internal/mind/context.go` (contract block 10, first dropped)
- [x] T020 [P] [US4] Journal selection tests in `internal/sim/journal_test.go` + block tests in `internal/mind/context_test.go`: relevant-entry match, no-match omission, excerpt cap, determinism
- [x] T021 [US4] Memories-block floor: protect 4 entries, drop-above-floor + serendipity-tail drop wiring in `internal/mind/context.go` (contract blocks 8-9); degraded-mode passthrough test (no SitVec ⇒ legacy selection, block renders, nothing crashes)
- [x] T022 [US4] Planted-memory relevance test per quickstart §SC-006 in `internal/mind/context_test.go`: relevant items included ≥80% across seeds within budget
- [x] T023 [US4] Multi-day budget-fit run per quickstart §SC-005: aggregate `cog.thought` PromptBytes/DroppedBlocks from a scratch world, verify ≥99% within budget, record numbers on TASK-105

**Checkpoint**: all spec stories implemented.

---

## Phase 8: Polish & cross-cutting

- [x] T024 Replay determinism: extend `internal/daemon` replay coverage (mirror `embed_replay_test.go`) to assert IntentLog/NeedsAnchor state and assembled prompts reproduce byte-identically from the event log
- [x] T025 [P] `go vet ./...`, `gofmt -l` clean, full `go test ./...` green in the worktree
- [x] T026 Update `docs/wiki/decision-context.md` final pin + re-pin touched notes via `/grounding-wiki:wiki-update` after merge (agent-mind, memory-retrieval, event-types, sim-state-reducer); then player-docs freshness check (`node .claude/skills/player-docs/scripts/check-freshness.mjs --check`)
- [ ] T027 SC-007 baseline comparison run per quickstart (flip-rate vs world-01, spike counting method); record on TASK-105 (measured, not merge-gating)

---

## Dependencies & execution order

- Phase 2 (T002-T005) blocks all stories except US5's T006 (the note can draft from contracts while the skeleton builds — [P] with T002-T005).
- US1 (T008-T013): T008 → T009 → {T010 [P], T011} → T012 → T013.
- US2 (T014-T016) independent of US1 (different fields/arms) — may run [P] with Phase 4 after Phase 2.
- US3 (T017-T018) depends only on Phase 2 (+T009 for the plan-end-in-history assertion).
- US4 (T019-T023) depends on Phase 2; T019/T020 [P] with T021.
- Phase 8 last; T026 post-merge; T027 any time after US1+US2 land (stronger after all).

## Implementation strategy

MVP = Phase 2 + US5 + US1 (T001-T013): self-history visible, audited, thrash-replay evidence in hand. Then US2 (cheap, high leverage), US3, US4 as increments on the same branch; single PR when the branch is complete per Principle II. Tier plan: T006-T007 Sonnet; everything touching reducers/`internal/mind` (T002-T005, T008-T023) Opus 4.8; T024-T027 with the closing implementer. Escalation one-way Sonnet→Opus per constitution.
