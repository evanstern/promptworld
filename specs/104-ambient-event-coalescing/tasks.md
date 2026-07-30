# Tasks: Ambient event coalescing — Arm A (TASK-176)

**Input**: `specs/104-ambient-event-coalescing/spec.md` (operator rulings 1-5,
2026-07-30, binding) + `plan.md` (D1-D6). Tests land alongside the code they
prove, in the same task. Tier: Opus 4.8 (card-recorded).

## Phase 1: Contracts + the derived-progress engine (D1)

- [X] T001 Research note (`specs/104-ambient-event-coalescing/research.md`):
  pin the within-tick ordering contract (needs minutes → agent steps by index
  → gru beat, matched to today's stepEvents emission order), the
  `PathSegment`/`NeedsSyncTick` field shapes and omitempty sentinels
  (spec-083 Neglect precedent cited), and the K-dial name/default/clamp —
  decisions + rationale before any code.
- [X] T002 `internal/sim/advance.go`: `State.AdvanceTo(tick)` engine —
  strictly-before semantics, fixed processing order, per-item watermarks,
  idempotence; `Apply` calls it before dispatch. Unit tests for ordering,
  idempotence, monotonicity, and no-op on a state with no pending derived
  work (old-log posture) in `internal/sim/advance_test.go`.

## Phase 2: Movement at exact per-step fidelity (D2 — the hard slice)

- [X] T003 Payloads + arms: `agent.path_started{agent, path, move_every,
  phase}` and `agent.path_truncated{agent, x, y}` in payloads.go/state.go —
  segment install and truncate-with-position arms; `Agent.Path *PathSegment`
  (omitempty; pre-change snapshot round-trip byte-identity test).
- [X] T004 Advancement steps: scheduled step execution — position update +
  `markExplored` + `notePresence` at each step's tick, path-tile 2x rule
  evaluated against state at that tick, cadence numbers read from the payload
  never the constant (FR-006); interleaving across agents' segments.
- [X] T005 Emission rewire (executor.go): walks emit `path_started` instead of
  per-step `agent.moved`; the FULL truncation set emits `path_truncated` —
  re-decision, absorb re-arm, hail pause, death, teleport, world pause
  (truncate-all), unreachable; spec-097 arrival observation still fires at
  the arrival tick, exactly once (test).
- [X] T006 **Sighting equivalence harness (the ruling-2 / FR-004 proof)**:
  paired synthetic logs (per-step vs coalesced) for single walk, crossing
  walks with mutual sightings, path-tile speedup, mid-walk truncation,
  sleeper/waker bystander, pause mid-walk — `State.Marshal()` bytes equal at
  EVERY tick boundary, including canonical per-agent mental-map bytes
  (explored bitmap + peer sightings with `Seen` ticks).

## Phase 3: Needs thinning (D3)

- [X] T007 K dial: `needs_checkpoint_minutes` in tuning.go (spec 048 pattern —
  clamp [1,60], default 10, genesis-pinned per spec 057); dial tests.
- [X] T008 Derived per-minute decay in advancement (trajectory anchor +
  neglect band anchors move with it); `NeedsSyncTick` watermark set by the
  `agent.needs_changed` arm; double-decay guard proven: an old log (per-minute
  events) replays with advancement contributing nothing (byte-identity test
  vs the pre-change arm's fold).
- [X] T009 Emission rewire (executor.go): checkpoint every K game-minutes +
  immediate band/near-death/zero crossing emission with absolutes; death path
  unchanged. Tests: K=1 reproduces today's emission byte-for-byte; K=10
  derived minute values equal K=1 folded values; crossings emitted at the
  same minute as today (guardian survival-watch latency test through
  internal/sim/guardian.go + internal/guardian/orders.go match paths).

## Phase 4: Gru derived motion (D4)

- [X] T010 Move stalk/prowl into advancement (shared decision function; RNG
  purpose "gru-prowl" preserved); retire `gru.moved` emission; arm retained —
  old-log gru replay byte-identity test; executor attack/sighting emission
  over advanced positions (ordering test: gru beat after agent steps).

## Phase 5: Consumers (D5)

- [X] T011 TUI: replica `AdvanceTo` on tick pushes (map positions per-step
  smooth); digest rows for `path_started`/`path_truncated`; historic
  `agent.moved`/`gru.moved` rows kept; `TestCatalogSweep` green.
- [X] T012 Mind: replica `AdvanceTo` per absorb batch; `armEncounters` becomes
  the first-adjacency sweep over the advanced replica (same adjacency
  moments, pair cooldown preserved — test against a scripted crossing-walk
  scenario).

## Phase 6: Whole-system proofs + measurement

- [ ] T013 Full determinism battery: `TestDeterminismSameSeedSameTimeline`
  (incl. per-agent canonical map bytes), kill-9 recovery equivalence with a
  mid-segment/mid-needs-window snapshot + tail, pre-change seeded fixture
  world replays green; `go test -race ./...` green (FR-002/FR-007).
- [ ] T014 Spec-092 audit + doctrine bookkeeping: audit-note rows for every
  advancement-read constant (decay constants, witnessRadius class, gru
  cadence + RNG purpose), definition-site "retune requires format bump"
  comments (spec 094 pattern) (FR-005/FR-006).
- [ ] T015 **Volume measurement (the SC-001 proof)**: paired seed-1337
  baseline-vs-fixed synthetic runs at playtest-1-class dials, month-scale
  (29+ game-days, accelerated); record rows/game-day per family, total
  events, and db size projection in
  `specs/104-ambient-event-coalescing/measurement.md`; assert ≥4x ambient
  reduction; worlds preserved under ~/.promptworld/measure/ for review
  (FR-001; live playtest world never touched).

## Phase 7: Grounding (spec 069 in-branch; spec 047 TUI gate)

- [ ] T016 Wiki re-pins per plan.md's owed list (event-log, event-types
  family, sim-state-reducer family + replay-hazards audit, mental-maps
  family, executor family, gru, deterministic-rng, world-tuning family,
  snapshots, tui-chronicle-feed, tui-client/map-view, agent-mind/
  mind-driver-triggers, testing notes) — the event-type catalog notes gain
  the two new rows and the retired-emission annotations; rulings 1-5 mirrored
  into the event-log/reducer doctrine notes (FR-005, SC-004).
- [ ] T017 `docs/player/` regenerated (player-docs skill; freshness probe
  green) and `docs/design/tui/` chronicle pages re-verified + re-pinned
  (`node scripts/check-tui-design.mjs --changed` clean).
- [ ] T018 Gate sweep before PR: `node scripts/check-merge-drift.mjs pr` exit
  0 (wiki-repin-missing / player-docs-stale clean); spec-bridge sync after
  merge is orchestrator-owned bookkeeping, not this branch.

## Dependencies

T001 → T002 → {T003 → T004 → T005 → T006} and {T007 → T008 → T009} in
parallel; T010 after T004 (shares the engine's ordering slots); T011/T012
after T005 (types exist); T013/T014 after all code phases; T015 after T013;
T016-T018 last (grounding rides the finished branch).
