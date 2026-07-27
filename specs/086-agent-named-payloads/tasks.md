# Tasks: agent-named payloads (spec 086)

**Input**: Design documents from `/specs/086-agent-named-payloads/`
**Prerequisites**: spec.md, data-model.md (normative census/enforcement/
back-compat/rider contracts), plan.md, research.md

**Tests**: suites ARE deliverables — marshal/unmarshal tables, the three
enforcement sweeps, per-family emission drives, the pre-086 replay
byte-identity fixture, the `names = nil` digest proof, the hit-rate
test, and the two oracle entries land with the code they prove, same
commits.

**Batching discipline**: the census migration (Phase 3) is long and
mechanical. Each batch task flips ONE file-cluster's payload types +
their emission sites + their reducer-arm `.ID` reads, and MUST end with
`go build ./...` green and that cluster's emission-drive test green —
never leave the tree red across batch boundaries. The compiler is the
checklist (a flipped type breaks every un-migrated call site).

**Tier**: Opus 4.8 via `spec-implementer` (recorded on TASK-17 —
repo-wide payload migration + mechanical enforcement + back-compat
replay). Planning/gating stays on Fable 5.

## Phase 1: Setup

- [X] T001 Worktree already cut (`.worktrees/task-17`, branch
  `task-17-agent-named-payloads`; claim landed per spec 065/TASK-160/161
  flow). Confirm baseline: `go test ./...` green in the worktree; branch
  fresh vs `origin/main` — if stale, `git merge origin/main` INTO the
  branch (NEVER rebase; sibling lanes may be in flight).

## Phase 2: Foundational — the type, the catalog, the rails (blocks all user stories)

- [X] T002 `internal/sim/agentref.go`: `AgentRef` (struct marshal;
  dual-shape `UnmarshalJSON`), `Ref`/`Refs` constructors,
  `validateRefs` reflection walk; doc comments carry the R2/R3 laws
  (never in State; never validated in Apply) (plan D1, FR-001)
- [X] T003 Type test tables: marshal shape + fixed field order + unicode
  name fixture; unmarshal bare int / object / `[]AgentRef` over both /
  pointer; constructor sentinels and out-of-range; `validateRefs`
  accept/reject matrix (FR-001, SC-002)
- [X] T004 `internal/sim/payloads.go`: `PayloadCatalog` seeded with every
  CURRENT payload type (pre-migration shapes — the catalog exists before
  the flip so the sweep can drive the migration); doc-anchored
  completeness test vs `docs/wiki/event-types.md` backticks; catalog
  covers `journal.*` and `sim.tuning_applied` (outside the tui catalog
  today) (plan D2, FR-006)
- [X] T005 `TestPayloadAgentRefSweep` (frozen vocabulary + four-entry
  allowlist with rationale strings, data-model §5), landing with the
  vocabulary check scoped to already-migrated types via a census
  progress constant so the sweep is never red in any commit — flipped
  unconditional at T015 once the census completes; plus
  `TestNoAgentRefInState` (green from day one — State has no refs)
  (plan D2, FR-004/006, SC-002)
- [X] T006 Emission-door rails: `mustPayload` calls `validateRefs`
  (panic contract); `InjectSocial` decodes via `PayloadCatalog` +
  `validateRefs`, refusing the batch pre-dry-run; mutation tests (unnamed
  in-roster ref panics / is refused); explicit test that `Apply` accepts
  unnamed shapes (no arm validation — replay law R3) (plan D5, FR-005,
  SC-002)

**Checkpoint**: the type, catalog, and both doors exist; nothing is
migrated yet; all gates green.

## Phase 3: US1 — the census migration (P1) 🎯 — batched, compiler-driven

Each batch: flip the cluster's payload fields (data-model §3), rewrite
its emission sites to `Ref`/`Refs`, switch its reducer arms to `.ID`,
extend the cluster's emission-drive test, `go build ./...` +
cluster tests green.

- [X] T007 [US1] Batch A — core agent lifecycle
  (`internal/sim/agents.go` rows 1–31 + `state.go` rows 30–31:
  intent/work/recovery/harvest/built/build_failed/needs/died/neglect/
  talked/memory/thought/hail/crafting/consumption/tools/walls/storage/
  moved/AgentPayload): fields + executor emission sites + arms
  (FR-002)
- [X] T008 [US1] Batch B — social + mental map
  (`internal/sim/social.go` rows 33–39, `mentalmap.go` rows 48–51:
  relation/gave/rumor incl. subject/secret/conversation turn+scene/
  chest_taken/saw/place_told/place_revealed/map_corrected; PlaceFact.
  Source stays bare — allowlist §5.1): fields + sim emission sites +
  arms (FR-002)
- [X] T009 [US1] Batch C — consolidation + journal + cognition + plans
  (`consolidate.go` rows 42–47, `journal.go` rows 40–41, `cognition.go`
  rows 52–55, `plan.go` rows 56–57): fields + arms; `cog.tool_call` Args
  exempt (allowlist §5.2) (FR-002)
- [X] T010 [US1] Batch D — governance + gru + guardian actions + prose
  (`governance.go` rows 65–69 incl. embedded ProposalPayload + yeas/
  nays/witnesses, `gru.go` rows 63–64, `miracles.go` row 70,
  `morgue.go` row 71, `chronicle.go` row 72, `guardian.go` row 60
  nudged targets): fields + sim emission sites + arms (FR-002)
- [X] T011 [US1] Batch E — the SPLITS (data-model §4):
  `DirectiveIssuedPayload`, `OrderPlacedPayload`,
  `ProphecyDeclaredPayload` (+ claim mirror, agent-0-legal),
  `DeathRef` for `RunEndedPayload.Deaths`; arms fold `.ID`s into
  unchanged entities; door dry-runs validate mirrors;
  `TestNoAgentRefInState` still green (plan D4, FR-003)
- [X] T012 [US1] Batch F — `faith.changed` additive died-agent ref
  (`faith.go` row 73, `omitempty`, set iff reason villager_died; other
  reasons byte-identical to spec 085 emissions — regression-pinned)
  (FR-002)
- [X] T013 [US1] Batch G — out-of-sim emitters: `internal/mind` (convo,
  consolidate, telemetry, handlers, embedder, narrate, meeting),
  `internal/guardian` (turn, miracle_batch, plans, orders, prophecy,
  reportcard), `internal/bundle/effects.go`,
  `internal/persona/files.go` — construction sites to `sim.Ref`/`Refs`
  (the roster constant; no state needed) (plan D6, FR-002)
- [X] T014 [US1] Family emission-drive suite consolidated: fixture
  drives per family asserting named refs FROM LOG BYTES ALONE, incl.
  sentinels (−1 any/personal/target), posthumous references, injected
  batches, bundle/persona sites (US1 AS-1..6, SC-001)

**Checkpoint**: every emitter names its agents; the compiler + drives
prove the census; AC #1 is code + tests.

## Phase 4: US2 — enforcement flipped fully on (P1)

- [X] T015 [US2] `TestPayloadAgentRefSweep` flipped unconditional (the
  census-scoped constant from T005 removed); catalog updated to
  post-migration zero values incl. mirrors; synthetic-violation tests
  (vocabulary-tagged bare int fails; dead allowlist entry fails)
  (FR-006, US2 AS-1/2, SC-002)
- [X] T016 [US2] tui↔sim catalog weld: `catalogFixture` keys ⊆
  `sim.PayloadCatalog` asserted in `internal/tui/digest_test.go`
  (FR-006, US2 AS-1)
- [X] T017 [US2] Door mutation coverage completed against migrated
  types: every whitelisted agent-bearing type refused when unnamed;
  `mustPayload` panic paths per family (FR-005, US2 AS-3, SC-002)

**Checkpoint**: AC #3 — a future unnamed agent reference cannot land by
any path without a red test or a refused emission.

## Phase 5: US3 — back-compat proven and documented (P1)

- [X] T018 [US3] Pre-086 fixture log (checked in; spans the payload
  families incl. injected social rows): from-genesis replay
  byte-identity (`Marshal`/`Hash`) vs recorded pre-086 state bytes;
  pre-086 snapshot decode + stored-hash verification; `world.migrated`
  fixture shape untouched (FR-004, US3 AS-1..3, SC-003)
- [X] T019 [US3] Mixed-era proofs: live-continuation test (old rows keep
  bytes, new rows named); legacy-shape rows through every migrated arm
  (fold identity, no name rejection); same-seed double-run post-086
  event-history byte-equality (US3 AS-4/5, FR-012, SC-003)
- [X] T020 [US3] The documented contract: data-model §6 matrix
  re-grounded into the wiki event-types conventions (rides Phase 8's
  re-pin; THIS task adds the doc text to the note source so AC #4's
  "documented" clause has a durable home beyond the spec dir) (US3
  AS-6, SC-003)

**Checkpoint**: AC #4 — old worlds replay byte-identically; the contract
is written where consumers will look.

## Phase 6: US4 — chronicle names + jump-to-source hit rate (P2)

- [X] T021 [US4] `internal/tui/grammar.go`/`digest.go`: `refName` helper
  (payload name first, `agentName` fallback); agent-bearing digest rows
  switched; `resolvePayloadNames`/`m.agentNames()` untouched as the
  historic fallback (plan D7, FR-007, US4 AS-2)
- [X] T022 [US4] `catalogFixture` rewritten to named payloads;
  `TestCatalogSweep` gains the `names = nil` identical-output assertion
  for agent-bearing types (the AC #2 proof) (FR-007, US4 AS-1, SC-004)
- [X] T023 [US4] `resolveSubject` generic single-ref fallback
  (registry-first; exactly-one-distinct-in-roster ⇒ candidate; several ⇒
  unlocatable; `world.migrated` excluded); hit-rate test
  (registry+fallback > registry-only; pins `journal.entry_written`,
  `morgue.epilogue`, `cog.thought`; multi-ref stays unlocatable)
  (plan D8, FR-008, US4 AS-3..5, SC-004)

**Checkpoint**: AC #2 — new events name themselves with no replica;
jump-to-source finds measurably more of the feed.

## Phase 7: US5 — the reverse-jump rider (P3)

- [ ] T024 [US5] `internal/tui`: `stripHit`/`rosterHit` regions
  (chronHit pointer pattern; frame-top invalidation); renderers record
  geometry (`villagerStripView` glyph columns, `villagerRosterBody` row
  bands); `handleMouse` branches — strip glyph → `centerCameraOn(a.X,
  a.Y)`; roster row → select + jump + narrow pane→map; overflow marker
  and nil replica no-ops; strip standing-resolution comment amended
  (plan D9, FR-009, US5 AS-1/2/4)
- [ ] T025 [US5] `handleVillagersKey`: `J` (roster + detail) centers on
  the selected villager; key test incl. dead-villager grave coords (plan
  D9, FR-010, US5 AS-3/4)
- [ ] T026 [US5] Two `mouseParityOracle` entries
  (`panels/villager-strip.md`/`reverse-jump`/`click glyph`;
  `panels/villagers.md`/`reverse-jump`/`click row`) with real-dispatch
  checks (pan moved; roster also `villSelected` moved); mouse-parity
  sweep green both directions (FR-011, US5 AS-5, SC-005)
- [ ] T027 [US5] Design pages amended and re-verified:
  `panels/villager-strip.md` (control cell `— · click glyph`,
  display-only prose + parity note rewritten, keyboard-path note),
  `panels/villagers.md` (roster row cell + new reverse-jump row
  `J · click row`, parity note), `patterns/keymap.md` (villagers `J`
  row) + anything `node scripts/check-tui-design.mjs --changed` names —
  re-verify + re-pin same-PR (spec 047 gate; plan D10, FR-011, US5 AS-6,
  SC-005)

**Checkpoint**: the rider ships with keyboard parity, oracle coverage,
and amended authority pages.

## Phase 8: Polish, grounding, gates

- [ ] T028 Full-suite + scope-guard audit: `go test ./...`; zero diffs
  under `internal/cognition`/`internal/clock`/`internal/llm`/
  `internal/store`/`internal/world`; `internal/ipc` untouched; no new
  event types; no emission-order changes; stranger payloads untouched
  (FR-012)
- [ ] T029 Wiki re-pins in-branch per plan.md's set — the event-types
  family (parent conventions + every domain child with payload rows),
  event-log, sim-state-reducer family, sim-loop-injection-doors,
  tui-chronicle-feed, village-lens, tui-villagers-tab, plus whatever the
  gate names from `sources:` vs the diff
  (`/grounding-wiki:wiki-update`); `docs/player/` regenerated
  (`node .claude/skills/player-docs/scripts/check-freshness.mjs --check`
  green) (constitution IV)
- [ ] T030 From the worktree: `node scripts/check-merge-drift.mjs pr`
  exits 0; open the ONE PR (body carries: the wire-shape change for
  operator review, the split-law review obligation, the `J` keybinding +
  standing-resolution-1 amendment, the untouched-packages audit); merge
  is `gh pr merge --merge` (SC-006)
- [ ] T031 Post-merge bookkeeping (orchestrator, per TASK-160/161
  landing laws): board card moves at root (board-sync exception, scoped
  commit, pushed immediately); spec-bridge sync + tasks.md ticks ride a
  branch and land by merge — never a direct root commit; worktree
  removed, branch deleted, root ff-pulled

## Dependencies

- Phase 2 blocks everything (T002 blocks all flips; T004/T005/T006 rails
  before batches so every batch lands validated).
- Phase 3 batches are sequential (each ends green); T011 (splits) after
  T007–T010 (the entities' arms are already `.ID`-reading); T013 after
  all sim-side flips (mind/guardian compile against final shapes); T014
  closes the phase.
- Phase 4 needs Phase 3 complete (the sweep flips unconditional only
  when the census is done). Phase 5 needs Phases 3–4 (fixtures replay
  against final shapes). Phase 6 needs Phase 3 (named fixtures). Phase 7
  is independent of 3–6 (pure TUI camera work) but lands after 6 to
  keep `internal/tui` churn ordered. Phase 8 last.
- US1+US2+US3 together are the deliverable's spine (the format, its
  enforcement, its safety); US4 is the visible AC #2 payoff; US5 is the
  operator rider.
