# Tasks: Needs-Conditioned Recovery Intents

**Input**: spec.md, plan.md (specs/064-needs-conditioned-recovery/)

**Tests**: included — SC-001..005 demand deterministic evidence; new intent
payload fields make replay/snapshot coverage mandatory.

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

*None.*

---

## Phase 2: Foundational

- [x] T001 Intent condition fields (plan R1): `UntilNeed`/`UntilValue`
      (omitempty) through the intent lifecycle — intent_set payload, active
      record, snapshot (`internal/sim/agents.go`, `state.go`); closed-set
      validation at the door; pre-064 compat test (absent fields ≡
      arrive-and-done, empty marshal byte-identical); rebase check (no new
      tick anchors expected — assert).
- [x] T002 Doctrine constants (`internal/sim/agents.go`): `warmthRecoverTo`
      (default recovery threshold, healthy margin above dangerWarmthBelow,
      chosen against the needs scale) + `recoveryStallTicks` (abort window) —
      named, dial-ready, NOT tuning.json.

---

## Phase 3: User Story 1 — warm_up (P1)

- [x] T003 [US1] Executor hold-at-target + per-tick completion check
      (`internal/sim/executor.go`, plan R2): conditioned intent stays active
      after arrival; completes (normal intent_done) on threshold crossing;
      already-satisfied completes immediately.
- [x] T004 [US1] `warm_up` planner tool (`internal/tool/registry.go` +
      `internal/mind/handlers.go`): target resolution like goto_warmth,
      optional `until_warmth` clamped-with-notice (058 posture), default
      `warmthRecoverTo`.
- [x] T005 [US1] Reflex issuance (`internal/sim/policy.go`): the 062 day and
      night warmth-recovery rungs issue the conditioned form with the
      doctrine default; source discipline unchanged (reflex completions never
      arm the 062 yield window).
- [x] T006 [P] [US1] Tests: recover-then-release determinism (SC-001, exact
      crossing tick); planner arg respected + clamped; reflex hold (no
      arrive-idle-wander, day and night); replay determinism over a recovery
      span.

---

## Phase 4: User Story 3 — interruptibility (P1)

- [x] T007 [US3] Abort-on-no-progress (plan R4): while holding, no net gain
      over `recoveryStallTicks` → distinct abort outcome; agent re-decides.
- [x] T008 [P] [US3] Tests: fire-dies-mid-recovery aborts; preemption path
      interrupts a holding intent exactly as any active intent (no new
      immunity); existing staleness window ends a stuck recovery (SC-003).

---

## Phase 5: User Story 2 — generic mechanism proof (P2)

- [x] T009 [US2] Rest analog (plan R6): align sleep's end conditions to the
      shared condition-check helper behavior-preservingly; if not
      behavior-preserving, prove via a rest-conditioned test variant and FLAG.
- [x] T010 [P] [US2] Tests: shared fields/check for the second consumer;
      existing no-condition intents byte-identical through the full suite
      (SC-004).

---

## Phase 6: User Story 4 — wake to cold (P2)

- [x] T011 [US4] Cold-emergency wake arm in the wake gate
      (`internal/sim/executor.go`, plan R5) reusing the exposure-band
      constants; cozy sleep untouched.
- [x] T012 [P] [US4] Tests: the Oak-final-night shape wakes the sleeper;
      cozy-at-fire control sleeps through (SC-005).

---

## Phase 7: Integration proof

- [x] T013 Extended Sage scenario (SC-002): the 062 thrash regression
      extended with warm_up — agent held at the fire to threshold, then
      released; zero mid-recovery dispatches end-to-end through the executor.

---

## Phase 8: Polish & Cross-Cutting

- [x] T014 Full gates in worktree: `go test ./...`, `go vet ./...`,
      `node scripts/check-tui-design.mjs --changed` (if warm_up surfaces in
      internal/tui goal-label tables, amend docs/design/tui/ same-PR).
- [ ] T015 Post-merge (root): wiki re-pin (executor, reflex-policy,
      tool-registry, agent-mind, event-types), player-docs freshness,
      spec-bridge sync.
