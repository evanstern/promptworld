# Tasks: Instinct Yields to Intelligence

**Input**: spec.md, plan.md (specs/062-instinct-yields/)

**Tests**: included — SC-001..003 demand deterministic test evidence; the
regression corpus is the deliverable proving the layer fight is over.

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

*None.*

---

## Phase 2: Foundational

- [x] T001 Yield state: per-agent last non-reflex intent-completion tick
      (`internal/sim/state.go` completion arm reads intent source; Agent field
      with omitempty; reflex completions excluded); rebase taxonomy SHIFT
      classification; snapshot-compat test (pre-062 fixture, empty marshal).
- [x] T002 Doctrine constants (`internal/sim/agents.go`): danger bands
      (`dangerFoodBelow`/`dangerWarmthBelow`/`dangerRestBelow`, anchored per
      plan R3) + `prepYieldTicks` (1800, plan R4) — named, single home,
      dial-ready comments, NOT tuning.json.
- [x] T003 Rung classification in `internal/sim/policy.go`: restructure
      decideIntent into ordered SURVIVAL and PREP groups per plan R2 (audit
      table authority; TASK-108 dying-fire refuel = SURVIVAL; top-up refuel,
      first-fire prep, larder stock = PREP). Pure restructure — behavior
      byte-identical before T004 (prove with existing suite green at this
      commit).

---

## Phase 3: User Story 1 — Prep yields (P1)

- [x] T004 [US1] PREP gate: skip the PREP group when (tick − last non-reflex
      completion) < prepYieldTicks OR any need below its danger band; SURVIVAL
      group unconditioned.
- [x] T005 [P] [US1] Tests: yield window suppresses prep post-planner
      completion and decays; danger band suppresses prep regardless of window;
      reflex completions never arm the window; no-LLM parity drive (SC-003 —
      planner-free reflex matches pre-062 intents except danger-band
      suppression cases, enumerated).

---

## Phase 4: User Story 2 — Day warmth rung (P1)

- [x] T006 [US2] Shared warmth-ladder helper (seek known → refuel-dying →
      build) used by the night branch unchanged and by a NEW day rung gated on
      the warmth danger band, ordered before all PREP.
- [x] T007 [P] [US2] Tests: day + cold + known warmth → seek; day + cold + no
      known + wood → build; day + warm → today's day branch byte-identical.

---

## Phase 5: User Story 3 — Night search fallback (P3, droppable by amendment)

- [x] T008 [US3] Frontier-search rung (reuse nearestFrontier) between the
      night ladder's last rung and terminal sleep, gated per FR-005/plan R6.
- [x] T009 [P] [US3] Tests: cold night, nothing known/carried/choppable,
      frontier reachable → search; no frontier → sleep (today's terminal).

---

## Phase 6: User Story 4 — Thrash regression (P1)

- [x] T010 [US4] `internal/sim/thrash_regression_test.go`: the Sage-shape
      scenario (plan R7) — new arbitration: zero prep in window + warmth
      recovers; zeroed-doctrine variant reproduces the old flip (one test,
      both proofs).

---

## Phase 7: Polish & Cross-Cutting

- [x] T011 Full gates in worktree: `go test ./...`, `go vet ./...`,
      `node scripts/check-tui-design.mjs --changed`; re-run after rebase
      (siblings 110/109/111 may merge first).
- [x] T012 Post-merge (root): wiki re-pin (reflex-policy, sim-state-reducer,
      decision-context), player-docs freshness, spec-bridge sync.
