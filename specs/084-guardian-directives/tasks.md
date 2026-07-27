# Tasks: guardian directives and designations (spec 084)

**Input**: Design documents from `/specs/084-guardian-directives/`
**Prerequisites**: spec.md, data-model.md (normative entities/rung),
contracts/events.md (normative vocabulary), plan.md, research.md

**Tests**: fixtures and suites ARE deliverables — reducer tables, replay
byte-identity, the reflex matrix, the interruption drive, and the
composition proof land with the code they prove, same commits.

**Tier**: Opus 4.8 via `spec-implementer` (recorded on TASK-157 —
cross-package, reflex-arbitration doctrine, injection door, decision
context). Planning/gating stays on Fable 5.

## Phase 1: Setup

- [X] T001 Worktree already cut (`.worktrees/task-157`, branch
  `task-157-guardian-directives`; claim landed per spec 065/TASK-160
  flow). Confirm baseline: `go test ./...` green in the worktree; branch
  fresh vs `origin/main` — if stale, `git merge origin/main` INTO the
  branch (NEVER rebase; sibling lanes task-133 and later task-118/17 are
  in flight).

## Phase 2: Foundational — grammar entry point, entities, doors, sweeps (blocks all user stories)

- [X] T002 `internal/target/target.go`: exported `ParseLocus` (bare
  point/rect/line locus; same normalization + `Tiles()`) + table tests
  in `target_test.go`; stdlib-only import pin stays green; zero bundle
  behavior change (plan D1, FR-003)
- [X] T003 `internal/sim` new file(s) (`plans.go` or split):
  `Designation`/`Directive` structs, `GuardianDesignationCap` (16) /
  `GuardianDirectiveCap` (3) / shared TTL constants, generalized one-way
  transition + retention prune, fulfillment predicate functions
  (data-model §6) (plan D2, FR-002/005/008)
- [X] T004 Reducer arms for all seven event types per
  contracts/events.md §1 (validate-not-clamp; `Status`/`PlacedSeq`
  reducer-stamped; occupancy/bounds/caps/TTL/text validation; the
  `designation.placed` all-villager place-fact grant), wired into
  `Apply`'s dispatch; `"designation"` joins the `PlaceFact` kind
  vocabulary + `factHorizon` in `internal/sim/mentalmap.go` (plan D2,
  FR-004/006/009)
- [X] T005 `internal/sim/loop.go`: four `injectSocialWhitelist` entries
  (`designation.placed/cancelled`, `directive.issued/cancelled`) with
  doctrine comments; ended-world narrowing inherited (plan D3, FR-004/009)
- [X] T006 `internal/sim/executor.go` `stepEvents`: designation-
  fulfillment → directive-fulfillment → directive-expiry sweeps (incl.
  all-targets-dead), the once-only `charge_regenerated` idiom, fixed
  order per research R14 (plan D4, FR-004/009)
- [X] T007 `internal/sim/miracles.go` `rebaseTicks`: active-directive
  `ExpiresTick` SHIFT arm + KEEP doc comments (data-model §10, FR-014)
- [X] T008 Reducer/lifecycle test suites: per-arm validation tables,
  race pairs (one terminal), prune determinism, place-fact grant,
  sweep once-only + ordering, rebase taxonomy, from-genesis replay
  byte-identity over a full-lifecycle fixture log (SC-004; FR-014)

**Checkpoint**: entities live end-to-end in sim with replay proof — no
tool, prompt, or UI surface yet.

## Phase 3: US2 — designations placed, announced, rendered, fulfilled (P1) 🎯 MVP

- [X] T009 [US2] `internal/tool/registry.go`: `place_designation` /
  `cancel_designation` declared per data-model §9 (Enums, Text locus,
  Events); buildable-structure-kind mirror + guardian-side drift test
  (plan D7)
- [X] T010 [US2] `internal/guardian`: designation handlers — locus parse
  via `ParseLocus`, per-kind form/param checks (partial-args refused,
  the parseReveal shape), `dsg-` id minting (`nextOrderID` clone),
  `InjectSocial` landing, door rejection → `rejected_gate` counsel;
  `turn.go` `writeDesignations` prompt section (plan D8, FR-015)
- [X] T011 [US2] `internal/tui/tiles.go`: three registry rows (site /
  wall segment / zone perimeter; semantic16 tokens) + map tile
  resolution renders them beneath real entities, active-only; registry
  sweep/identity tests (plan D10, FR-007)
- [X] T012 [US2] `docs/design/tui/`: amend `panels/map.md` (+ every page
  `node scripts/check-tui-design.mjs --changed` names) — re-verify +
  re-pin same-PR (spec-047 gate; plan D11, FR-007)
- [X] T013 [US2] End-to-end designation tests: place each kind through
  the tool door → state + grant + render asserted; build-at-site flips
  `fulfilled` via the sweep; cancel races; occupancy/bounds/cap
  rejections; pre-existing-structure placement fulfills next boundary
  (spec US2 scenarios; SC-004 fixture feeds T008's replay log)

**Checkpoint**: US2 independently testable — the plan layer exists,
visible on the map and in villager knowledge.

## Phase 4: US3 — directives through the injection door; observable lifecycle (P1)

- [X] T014 [US3] `internal/tool/registry.go`: `issue_directive` /
  `cancel_directive` declared (send_omen target vocabulary, TTL default
  3); `observableEventTypes` += the four `directive.*` types (12 → 16)
  (plan D7, FR-009/010)
- [X] T015 [US3] `internal/guardian`: directive handlers — target
  resolution (comma-names / `"everyone"`, all-living-or-reject),
  `dir-` id minting, atomic batch with companion `agent.memory_added`
  per target, rejections as `rejected_gate`; `turn.go`
  `writeDirectives` section (plan D8, FR-009/015)
- [X] T016 [US3] Door/firewall suites: rejection table (dead target,
  unknown/non-active designation, TTL, cap, text runes); atomic
  companion batch; firewall audit extension — directive text reaches
  villager prompts only via state-rendered surfaces (spec US3
  scenarios 2/6, AC #4)
- [X] T017 [US3] Composition proof (AC #7 / SC-003): integration test —
  `monitor_and_act` watch on `directive.fulfilled` triggers through
  UNMODIFIED `matchOrders` when a directive fulfills; assert zero
  diff in `internal/guardian/orders.go` matching code

**Checkpoint**: directives land, expire, fulfill, and are watchable;
villager behavior still unchanged.

## Phase 5: US4 — the villager side: block + DIRECTIVE rung + interruption proof (P1)

- [X] T018 [US4] `internal/mind/context.go`: `renderDirective` +
  `fixedBlocks` insertion (`directive`, `neverDrop`, between
  `plan_echo` and `known_places`); `renderKnownPlaces` handles the
  `designation` landmark kind; golden prompt tests — directive-present
  rendering AND directive-free byte-identity (plan D9, FR-011; SC-006)
- [X] T019 [US4] `internal/sim/policy.go`: `directiveDecision` per
  data-model §8 (oldest-active selection, per-kind routing table,
  orphan fall-through), called after `survivalDecision` / before the
  `prepYields` consult; `heed_directive` goal through the intent
  machine + goal/duration mirrors + drift test (plan D5, FR-012;
  research R13)
- [X] T020 [US4] Reflex matrix suites (the `reflex_matrix_test.go`
  pattern): survival preempts directive; directive preempts prep and
  wander; per-kind routing cells (build-with-materials / walk-to-site /
  at-site fall-through / wall-line next-tile order / zone presence);
  directive-free parity drive (SC-001, SC-006; AC #5)
- [X] T021 [US4] Interruption-resume drive (SC-002 / AC #6): reflex-only
  fixture — hail mid-walk pauses directed work, conversation runs,
  next idle decision re-resolves the same directive; PR body records
  the zero-interruption-code diff obligation (FR-013)
- [X] T022 [US4] Reflex-only end-to-end (SC-001): issued structure-site
  directive → fed, warm villager walks, builds, designation fulfills,
  `directive.fulfilled` recorded — no planner call in the drive

**Checkpoint**: all four hardness properties proven (survival-first,
directive-before-free-time, interruption-friendly, resumable).

## Phase 6: US1 — survey_site (P2)

- [X] T023 [US1] `internal/tool/registry.go` + `derive.go`:
  `survey_site` declared (Effect Read, Gate None; x/y/radius);
  `guardianToolDesc` entry — renders under `GuardianReadGuidance`,
  skipped by acting guidance (plan D7, FR-001)
- [X] T024 [US1] `internal/guardian/survey.go` + `handleSurvey`
  (`handleExplain` dispatch shape): deterministic sheet — terrain mix,
  nearest water/tree/rock distances, structures, passability; bounds →
  repairable in-fiction miss (plan D8, FR-001)
- [X] T025 [US1] Survey suites (SC-005): byte-identity for identical
  (args, state); no event appended, no charge, acting cardinality
  untouched; bounds-miss shape; guidance-path render assertion (spec
  US1 scenarios)

## Phase 7: Cross-cutting surfaces

- [X] T026 `internal/tui/grammar.go`: digest rows for all seven event
  types; `TestCatalogSweep` green (SC-007; contracts §3)
- [X] T027 Full-suite pass: `go test ./...` including the spec-062
  parity drives, degraded-mode survival suites, and the miracle/replay
  suites — no regression anywhere (SC-004/006)

## Phase 8: Grounding + gates (the wiki-in-PR lifecycle, spec 069)

- [X] T028 Wiki re-pins IN-BRANCH per plan.md's re-pin set (the pr gate
  is the authority: every note whose `sources:` this branch touched),
  including the NEW plan-layer note (e.g. `guardian-designations`) +
  INDEX row + event-types family row; `docs/player/` regenerated
  (`node .claude/skills/player-docs/scripts/check-freshness.mjs --check`
  green) (SC-008)
- [x] T029 Gates from the worktree: `node scripts/check-tui-design.mjs
  --changed` exit 0 (T012 pages amended); `node
  scripts/check-merge-drift.mjs pr` exit 0; open the ONE PR (body:
  hardness ruling, charge-free plan-layer + stage-availability
  assumptions flagged for the operator, the zero-interruption-code diff
  note); merge with `gh pr merge --merge` ONLY (SC-008)
- [x] T030 Post-merge bookkeeping (derived state only, authored on a
  branch and merged per TASK-160): spec-bridge sync, AC ticks
  (#1–#7 as proven), board card to Done with final summary; worktree
  cleanup

## Dependencies & execution order

- Phase 2 blocks everything (T002 → T003 → T004 → T005/T006/T007 → T008).
- Phase 3 (US2) needs Phase 2; Phase 4 (US3) needs Phase 2 + T009-ish
  designation placement to bind against (T013 fixtures reusable).
- Phase 5 (US4) needs Phases 3–4 (a directive to execute).
- Phase 6 (US1) needs only Phase 2's map/state access — may run in
  parallel with Phases 3–5 after T002.
- Phases 7–8 last; T029 gates the PR; T030 is post-merge.

## Notes

- One TASK, one PR: every phase lands as commits on
  `task-157-guardian-directives`; dotted subtasks never get their own
  PR.
- OUT of scope (FR-016): TASK-158 missions, TASK-118 faith accounting
  (its contract is `directive.fulfilled`'s payload — contracts §3),
  bundle grammar/matrix changes, any interruption/pause machinery
  change.
- Push the branch on first commit; a rejected push is a
  stop-the-lane signal (spec 065).
