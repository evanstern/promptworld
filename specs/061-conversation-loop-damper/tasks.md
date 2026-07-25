# Tasks: Conversation Loop Damper

**Input**: spec.md, plan.md (specs/061-conversation-loop-damper/)

**Tests**: included — SC-001..005 demand test evidence; new reducer-visible
state means replay + rebase coverage is mandatory.

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

*None.*

---

## Phase 2: Foundational — pair record (US2)

- [x] T001 [US2] `PairTalks` on `sim.State` (sorted slice, A<B invariant,
      `json:"pair_talks,omitempty"`, canonical-bytes safe) + update in the
      `agent.talked` reducer arm (`internal/sim/state.go`); rebase taxonomy:
      pair ticks classified SHIFT (`rebaseTicks` + miracles_test).
- [x] T002 [P] [US2] Tests: record updates on both hail-founded and ambient
      talks (unordered — one record per pair); pre-061 snapshot (no
      `pair_talks` key) loads as never-talked and marshals byte-identically
      when empty; snapshot round-trip; replay determinism over a talk-heavy
      sequence (SC-004).

---

## Phase 3: User Story 1 — sim-side hail cooldown (P1)

- [x] T003 [US1] Pair-cooldown check in `hailable()`
      (`internal/sim/hail.go`) consuming `s.EncounterCooldown()`; confirm both
      the landing door's hail rungs and `hailStep` route through it; refusal
      surfaces via the landing outcome with a "spoke recently" message the
      planner sees.
- [x] T004 [P] [US1] Tests: the world-01 loop shape — hail-founded talk →
      immediate replanned talk_to → exactly one scene + informative refusal;
      past-cooldown retry founds; dial=0 world (TuningState) makes the gate
      vacuous (SC-001).
- [x] T005 [P] [US1] Regression pins: ambient-beat gate (`canTalk`) and
      encounter-arming gate (`pairSeen`) suites pass byte-identically; hail's
      deliberate bypass of the AMBIENT cooldown (TASK-47) preserved for
      first/post-cooldown talks (SC-002).

---

## Phase 4: User Story 3 — novelty SHIM (P2)

- [x] T006 [US3] Novelty gate at mind scene founding
      (`internal/mind/convo.go`): require a new above-floor salience memory on
      either side since the pair's last exchange (last-exchange tick from the
      replica's PairTalks — one source of truth); salience floor as a named
      promoted-dial-ready constant (NOT in tuning.json).
- [x] T007 [US3] Last-gist context: a founded scene's prompt carries the
      pair's previous exchange gist (convo_record machinery) as
      "this already happened" context.
- [x] T008 [US3] SHIM marking: `SHIM(TASK-109)` marker at every gate site +
      removal-condition doc block (model tiers make it unnecessary); note in
      the spec dir README or plan appendix (SC-005 greppable).
- [x] T009 [P] [US3] Tests: no-new-memory founding refused; new-memory
      founding admitted with gist in prompt; marker grep test optional but the
      marker must exist (SC-003, SC-005).

---

## Phase 5: Polish & Cross-Cutting

- [x] T010 Full gates in worktree: `go test ./...`, `go vet ./...`,
      `node scripts/check-tui-design.mjs --changed`; re-run after rebase
      (siblings 110/108/111 merge first — expect a rebase).
- [x] T011 Post-merge (root): wiki re-pin (social-fabric, hail-protocol,
      agent-mind, sim-state-reducer notes), player-docs freshness,
      spec-bridge sync.
