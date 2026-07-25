# Tasks: Tool Surface Hygiene

**Input**: spec.md, plan.md (specs/058-tool-surface-hygiene/)

**Tests**: included — SC-001..005 demand test evidence.

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

*None.*

---

## Phase 2: Foundational

- [ ] T001 Add the expressive-clamp marker to the param model
      (`internal/tool/tool.go` Param + set it on say.text, gist, muse.text,
      and `reasonParam()` in `internal/tool/registry.go`); registry Validate()
      asserts Clamp only appears on Text params.
- [ ] T002 Rune-safe truncation helper (factor from the
      `internal/mind/meeting.go` NormTextMax idiom) usable for both rune caps
      and byte caps; unit tests over multi-byte boundaries.

---

## Phase 3: User Story 1 — Expressive clamping (P1)

- [ ] T003 [US1] Tool-loop Text validation (`internal/toolloop/loop.go`):
      Clamp-flagged over-cap args are truncated in place (call proceeds);
      validation gains a replace-value path; non-flagged Text args keep
      rejecting. Clamp surfaces in the model-facing tool result naming field +
      cap.
- [ ] T004 [US1] Telemetry: clamped acceptance distinguishable in the
      cog.tool_call verdict surface (`internal/mind` handlers/telemetry) —
      reuse existing verdict shape (SC-005).
- [ ] T005 [P] [US1] Tests: over-cap on each of the four fields lands clamped
      (event carries clamped text, valid UTF-8); structural failures still
      reject (existing suite green) (SC-001, SC-003).

---

## Phase 4: User Story 2 — set_plan step clamp (P1)

- [ ] T006 [US2] Landing guard (`internal/sim/landing.go`): oversized plan →
      first PlanStepCap steps accepted, clamp visible via an outcome of the
      existing shape; structurally invalid steps still reject.
- [ ] T007 [P] [US2] Tests: >cap plan lands with exactly cap steps + notice;
      invalid step still rejects; replay of the clamped landing deterministic
      (SC-002).

---

## Phase 5: User Story 3 — Roster prune (P2)

- [ ] T008 [US3] Remove collect_water + bathe from `LoopRosterVillager`
      (`internal/tool/roster.go`) with the revisit-condition comment (thirst
      need); update glossQuarry/glossBuildOven (`internal/tool/registry.go`)
      and grep for other prompt-surface mentions.
- [ ] T009 [P] [US3] Tests: roster enumeration excludes both; existing executor
      tests for both verbs still green (machinery kept, SC-004); replay of
      historical collect_water/bathe events unaffected.

---

## Phase 6: Polish & Cross-Cutting

- [ ] T010 Full gates in worktree: `go test ./...`, `go vet ./...`,
      `node scripts/check-tui-design.mjs --changed`; re-run after rebase.
- [ ] T011 Post-merge (root): wiki re-pin for notes sourcing
      toolloop/registry/roster; player-docs freshness; spec-bridge sync.
