# Tasks: faith-driven charge regeneration (spec 085)

**Input**: Design documents from `/specs/085-faith-regen/`
**Prerequisites**: spec.md, data-model.md (normative state/events/
predicates/curve), plan.md, research.md

**Tests**: fixtures and suites ARE deliverables — reducer tables, sweep
determinism, band × posture matrices, replay byte-identity (new
lifecycle AND pre-085 logs AND genesis-band schedule), and the strip
render states land with the code they prove, same commits.

**Tier**: Opus 4.8 via `spec-implementer` (recorded on TASK-118 —
reducer doctrine: faith as event-sourced state, regen as pure function;
doctrine-adjacent by definition). Planning/gating stays on Fable 5.

## Phase 1: Setup

- [ ] T001 Worktree already cut (`.worktrees/task-118`, branch
  `task-118-faith-regen`; claim landed per spec 065/TASK-160/161 flow).
  Confirm baseline: `go test ./...` green in the worktree; branch fresh
  vs `origin/main` — if stale, `git merge origin/main` INTO the branch
  (NEVER rebase; sibling lanes may be in flight).

## Phase 2: Foundational — faith state, fold, curve (blocks all user stories)

- [ ] T002 `internal/sim/faith.go`: `FaithState` + `State.Faith`
  (`omitempty`), `FaithGenesis`, nil-safe `FaithScore()`, the delta
  table constants (one home, data-model §3), `FaithChangedPayload`,
  exported `FaithRegenCadenceTicks(score, scenario)` band table
  (data-model §6) — no wiring yet (plan D1, FR-001/002/004)
- [ ] T003 `faith.changed` reducer arm wired into `Apply` (reason
  domain + sign validation, clamp fold, materialize-on-first); NOT
  whitelisted — doctrine comment in `loop.go` beside the
  `charge_regenerated` note (plan D1/D4, FR-002)
- [ ] T004 Fold/accessor test table: reason domain, sign mismatch
  refused, clamp both ends, materialize-on-first, nil-accessor genesis,
  pre-085 snapshot round-trip byte-identity (`omitempty`) (FR-001/002,
  SC-001)

**Checkpoint**: faith folds and reads; nothing emits it yet.

## Phase 3: US1 — the faith accounting sweep (P1) 🎯 MVP

- [ ] T005 [US1] `internal/sim/executor.go`: `faithEvents(s, batch,
  nextTick)` batch-scanning sweep (run-end idiom) emitting one
  `faith.changed` per `directive.fulfilled` / `directive.expired` /
  `agent.died` source in batch order, with the cannot-move emission
  gate; positioned after all source emitters, before
  `scenarioRubricEvents`/run-end (plan D3, FR-003)
- [ ] T006 [US1] Sweep test drive: directive fulfillment mints in-batch
  (+8, source_id = directive id); expiry −4; per-death −6; same-batch
  pileup order-determinism; score-100/0 no-emit edges; a bare
  `designation.fulfilled` mints nothing; ended-world silence (US1
  AS-1..4/6, SC-001)
- [ ] T007 [US1] Replay proofs: from-genesis byte-identity over a
  fixture log with every Phase-3 reason; a pre-085 fixture log replays
  byte-identically (no retroactive faith) (FR-014, US1 AS-5, SC-001)

**Checkpoint**: the mana loop's income side is live and replay-proven.

## Phase 4: US2 — regen as a pure function of faith + the posture decision (P1)

- [ ] T008 [US2] `internal/sim/executor.go`: regen check rewritten to
  `FaithRegenCadenceTicks` (cadence-0 short-circuit; same event, same
  empty payload; `chargeRegenTicks` survives as the steady-band value)
  (plan D3, FR-004)
- [ ] T009 [US2] Band × posture test matrix + boundary/off-boundary
  drives per band (the `TestChargeRegen` clone); scenario-forsaken
  emits nothing over a multi-day drive; ambient-forsaken emits exactly
  at 24h boundaries; genesis-band world schedule byte-identical to
  pre-085 (FR-004/005, US2 AS-1..5, SC-002/003)
- [ ] T010 [US2] Doctrine comments carry the FR-005 posture decision +
  reversal lever at the band table (the spec's AC #4 artifact echoed
  where the code lives); `guardian.go` export comment reconciled
  (legacy constant = steady band / TUI fallback) (FR-004/005)

**Checkpoint**: AC #3 and AC #4 are code + tests + recorded decision.

## Phase 5: US3 — prophecy: declare, verify, judge (P2)

- [ ] T011 [US3] `internal/sim/prophecy.go`: `Prophecy` +
  `ProphecyClaim` + `State.Prophecies` (`omitempty`),
  `GuardianProphecyCap`, normalized-claim equality, per-kind
  fulfil/fail predicates (data-model §5, shared by sweep and arms),
  retention prune (plan D2, FR-006/007)
- [ ] T012 [US3] Reducer arms: `prophecy.declared` door (full
  data-model §4 table incl. already-true and duplicate rejection;
  `Status`/`PlacedSeq` reducer-stamped) + terminal arms (re-validate,
  one-way); `prophecy.declared` joins `injectSocialWhitelist` (plan
  D2/D4, FR-008)
- [ ] T013 [US3] Verification sweep `prophecyEvents` in `stepEvents`
  (fixed position; fulfil before fail; once-only) + companion
  `OriginReport` terminal memories; `rebaseTicks` SHIFT/KEEP arms for
  `DeadlineTick`/`DeclaredTick` (plan D3/D5, FR-008/014)
- [ ] T014 [US3] `internal/tool/registry.go`: `prophesy` (Gate Charge
  1, params per data-model §4) + `observableEventTypes` +3;
  `internal/guardian`: `handleProphesy` (target resolution, claim
  assembly, `pro-` id minting, atomic `OriginOmen` companion batch,
  `rejected_gate` counsel), `turn.go` faith + prophecies prompt
  sections (plan D6/D7, FR-008/013)
- [ ] T015 [US3] Prophecy test suites: door rejection table; predicate
  table per cell (incl. survives fail-fast, late-truth-after-failed,
  cancelled-designation, all-targets-dead stays judged); terminal
  races; faith companions in-batch (+12/−15); provenance stamps
  (`OriginOmen`/`OriginReport`, `DirectPerception` classes); a
  `monitor_and_act` watch on `prophecy.failed` triggers through
  unmodified `matchOrders`; full-lifecycle from-genesis replay
  byte-identity (US3 AS-1..7, SC-001/004)

**Checkpoint**: AC #2 complete — event-sourced faith with the defined,
log-checkable verification rule, live end to end.

## Phase 6: US4 — the visible surface (P2)

- [ ] T016 [US4] `internal/ipc`: `ClockStatus.Faith *int` +
  `FaithRegenTicks` served from the sim functions (data-model §7);
  CLI/`metatron_status` parity asserted (plan D8, FR-009)
- [ ] T017 [US4] `internal/tui/views.go`: strip fourth segment
  (`faith N` / nil → `faith —`), forecast on the wire cadence with
  legacy fallback + cadence-0 omission; render tests for all states +
  drop order; `render_test.go:599` absence pin flipped (plan D9,
  FR-009, US4 AS-1..3)
- [ ] T018 [US4] `internal/tui/grammar.go`: four in-fiction digest rows;
  `TestCatalogSweep` green (plan D9, FR-010, US4 AS-5)
- [ ] T019 [US4] `internal/tui/lessons.go`: `first-faith-event` row
  (trigger `faith.changed`, tier mechanics, skin tokens,
  direction-neutral copy, strip/guardian-tab pointer); taxonomy tests
  flipped from absence-pin to presence (spec-077 rider closed) (plan
  D9, FR-011, US4 AS-4)
- [ ] T020 [US4] `internal/sim/rubric_hygiene_test.go`: `faith.` joins
  the banned prefixes (the recorded R2 obligation) (FR-012)
- [ ] T021 [US4] `docs/design/tui/panels/guardian-strip.md` §4
  reconciled to shipped (+ any page `node
  scripts/check-tui-design.mjs --changed` names) — re-verify + re-pin
  same-PR (spec-047 gate; plan D10, FR-009)

**Checkpoint**: the strip contract, the lesson rider, and every catalog
gate are satisfied.

## Phase 7: Polish, grounding, gates

- [ ] T022 Full-suite pass: `go test ./...`; scope-guard audit — zero
  diffs under `internal/cognition`/`internal/mind`; no tuning.json
  change; no new RNG; no whitelist entries beyond `prophecy.declared`
  (FR-015)
- [ ] T023 Wiki re-pins in-branch per plan.md's re-pin set
  (`/grounding-wiki:wiki-update`; new `guardian-faith` note; event-types
  family rows) + `docs/player/` regenerated
  (`check-freshness.mjs --check` green) (constitution IV)
- [ ] T024 From the worktree: `node scripts/check-merge-drift.mjs pr`
  exits 0; open the ONE PR (body carries: the FR-005 posture decision +
  lever for operator review, the delta-table normative defaults, the
  untouched-packages review obligation); merge is `gh pr merge --merge`
  (SC-006)
- [ ] T025 Post-merge bookkeeping (orchestrator, per TASK-160/161
  landing laws): board card moves at root (board-sync exception);
  spec-bridge sync + tasks.md ticks ride a branch and land by merge —
  never a direct root commit; worktree removed, branch deleted, root
  ff-pulled (AC #1 recorded: link + sync are the orchestrator's acts)

## Dependencies

- Phase 2 blocks everything; Phase 3 needs T002-T004; Phase 4 needs
  T002 (curve) and benefits from Phase 3 fixtures; Phase 5 needs
  Phases 2-3 (faith companions) and consumes the merged spec-084
  layer; Phase 6 needs Phases 2-5 shapes (wire reads the curve;
  digests/lesson need the event types); Phase 7 last.
- US1/US2 together are the MVP (the card's title mechanic); US3
  completes AC #2's prophecy half; US4 closes the strip contract and
  the spec-077 rider.
