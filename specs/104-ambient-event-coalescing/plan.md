# Implementation Plan: Ambient event coalescing — Arm A (TASK-176)

**Branch**: `task-176-ambient-event-coalescing` | **Date**: 2026-07-30 | **Spec**: [spec.md](spec.md)

## Summary

Arm A per the operator rulings (spec.md fork section, 2026-07-30): one derived-
progress engine plus three emission changes. (1) Per-step `agent.moved` is
replaced for new walks by `agent.path_started` (+ `agent.path_truncated` when
reality deviates), with movement, explored-bitmap growth, and mutual
tick-stamped peer sightings advanced derivedly at EXACT per-step fidelity
(ruling 2 — the hard slice; design in D1/D2). (2) `agent.needs_changed` keeps
its type and payload but thins to K-game-minute checkpoints plus immediate
band-crossing emissions, K a genesis-pinned tuning dial (ruling 3; D3).
(3) `gru.moved` emission retires — gru motion becomes fully derived from
(state, seed, tick) (D4). Additive vocabulary, NO log-format bump (ruling 4);
all old reducer arms retained forever; old-world relief is a non-goal
(ruling 5).

## Technical Context

**Language**: Go. **Surfaces**: internal/sim (executor.go, state.go, gru.go,
payloads.go, tuning.go, mentalmap.go, NEW advance.go), internal/mind (replica
advancement + encounter arming), internal/tui (replica advancement, digest
rows, map positions), tests across all three; docs/wiki + docs/player +
docs/design/tui grounding. internal/store and internal/world untouched (no
format bump, no migration mode).
**Constraints**: spec 092/TASK-75 — every new payload fully baked at emission,
no reducer arm re-derives an outcome from a mutable constant; new derived
computation reads constants only in the already-grandfathered D2
derived-bookkeeping class, and every such constant joins the spec-092 audit
(retune ⇒ format bump). Live playtest world never touched by tests.

## Constitution Check

- **I-IV: PASS.** Rulings artifact-recorded (spec fork section + board card);
  one branch/PR; equivalence harness + volume measurement as gate evidence;
  wiki re-pins + player-docs + TUI design reference ride this branch (spec 069
  pr gate; spec 047 gate — internal/tui/ changes).
- **V: PASS — Opus 4.8** (card-recorded rubric: cross-package architectural
  change, reducer/determinism doctrine surface, replay byte-identity proofs).
  Planning/gating stays on Fable 5; implementation delegates to
  spec-implementer with `model: opus`.

## Design

### D1 — The derived-progress engine (`internal/sim/advance.go`)

The one new mechanism everything else rides. In this codebase events are the
only state mutation path and quiet ticks advance only the clock; this feature
adds a second sanctioned derived channel, generalizing the spec-041 D2
precedent (`markExplored`/`notePresence` are already eventless pure functions
of (state, event)) to eventless pure functions of **(state, tick range)**:

- `State.AdvanceTo(tick)` processes all pending derived progress with
  scheduled tick **strictly before** `tick`, in a single fixed order: ticks
  ascending; within a tick, the same order today's `stepEvents` produces its
  emissions (needs minutes, then agent movement steps by agent index, then
  gru beat) — the within-tick ordering contract is pinned in a research note
  (T002) and enforced by the equivalence harness (D6).
- `Apply(e)` calls `AdvanceTo(e.Tick)` first, then applies `e`. This is what
  makes every fold path — live loop, recovery, `replayToTick`, morgue fold,
  TUI/mind replicas — advance identically with zero per-consumer wiring:
  anything that applies events gets derived progress for free. Consumers that
  progress PAST the last event (live loop per tick, replicas on tick pushes,
  `replayToTick` to its cutoff, recovery to `max(snapshot, last event tick)`)
  call `AdvanceTo` explicitly.
- Idempotent and monotone: advancement never runs a scheduled item twice
  (per-item watermarks live in the state fields that carry the pending work —
  D2's segment cursor, D3's needs sync tick) and never rolls back.
- Snapshots: pending derived work is ordinary marshaled state (segment +
  watermarks), so snapshot round-trip and kill-9 recovery reproduce mid-walk /
  mid-window progress exactly (existing recovery suites extended).

### D2 — Movement at EXACT per-step fidelity (ruling 2 — the hard slice)

**Emission** (executor): when a walk starts, emit
`agent.path_started{agent, path [[x,y],…], move_every, phase}` — the full BFS
path (already deterministic — [[reflex-pathfinding]]), plus the cadence
numbers baked into the payload (`moveEveryTicks`, the agent's phase offset) so
advancement never reads the compiled cadence constant (spec 092). When live
reality deviates from the declared path — intent re-decision, absorb re-arm,
hail pause, death, `guardian.entity_moved` teleport, world pause, target
unreachable — emit `agent.path_truncated{agent, x, y}` carrying the agent's
actual position at truncation (outcome in payload; the reducer arm sets
position and clears the segment, never recomputes where the walk "would have"
been).

**Advancement** (the exactness core): each scheduled step executes at its
scheduled tick doing EXACTLY what today's `agent.moved` arm does — position
update, `markExplored`, `notePresence(agent, stepTick)` — interleaved across
all agents' in-flight segments in per-tick order. Exactness holds because:

- *Step timing*: derived per tick from the payload's cadence numbers plus the
  path-tile 2x rule evaluated against state AT the step's tick — paths are
  event-sourced structures and advancement interleaves with `Apply`, so
  `pathAt(state, tick)` during advancement sees exactly what the live
  executor saw. (The speedup RULE is code, like any reducer arm; the numbers
  ride the payload.)
- *Mutual sightings*: `notePresence` at step tick t reads bystander positions
  at t — correct because ALL agents' segments advance through t together in
  the fixed order before any t-or-later event applies. Sighting `Seen` ticks,
  sighting pairs, and explored bits come out byte-identical to per-step
  emission (D6 harness proves it).
- *Within-tick semantics*: `AdvanceTo(e.Tick)` is strictly-before, so an event
  at tick t observes state including derived steps through t-1 — matching
  today, where tick t's `agent.moved` folds in tick t's batch and is read by
  others from t+1.
- *Pause*: today's executor freezes walkers while paused; advancement is
  tick-driven, so pause TRUNCATES all in-flight segments (one
  `agent.path_truncated` per walker at pause; pause is rare, cost trivial).
  This is emission-side, deterministic, and keeps advancement unconditional.

**Spec-097 arrival**: the executor observes arrival via advanced state at the
arrival tick and emits `agent.place_observed` exactly as today (emission path
unchanged; only the per-step `agent.moved` beside it disappears).

**State shape**: `Agent.Path *PathSegment` (omitempty pointer — the
Journal/Map precedent: nil for every old log and pre-change snapshot, so old
replay bytes never gain the field). Segment carries path, start tick, cadence
numbers, and the step cursor (the advancement watermark).

**Volume**: a 10-30 tile walk collapses from 10-30 rows to 1-2.

### D3 — Needs: K-minute checkpoints + crossings (ruling 3)

**No new event type.** `agent.needs_changed` keeps its payload (absolutes) and
its reducer arm — guardian survival watches (internal/sim/guardian.go:151),
standing-order matching (internal/guardian/orders.go), and the TUI digest row
are untouched by construction.

- *Derived decay*: advancement applies `decayNeeds` per game-minute per living
  agent (environment — night, `warmAt`, shelter, cold snap, asleep — read from
  state as-of-that-minute, exact live and replay by the D1 interleaving). The
  trajectory-anchor roll and neglect band anchors move with it (per-minute
  exact, as today).
- *Emission* (executor, after `AdvanceTo`): emit `agent.needs_changed` with
  current absolutes when (a) K game-minutes have elapsed since the agent's
  last needs emission, OR (b) any need crossed a spec-062 danger band,
  near-death, or zero boundary vs the last-emitted values — crossings at
  today's per-minute latency (ruling 3), OR (c) the death path fires (0
  health ⇒ `agent.died`, unchanged).
- *Double-decay guard*: the `agent.needs_changed` arm sets a per-agent
  `NeedsSyncTick` watermark (`omitempty`); advancement skips minutes ≤ the
  watermark. Old logs (an event every minute) therefore replay with
  advancement contributing nothing — every minute folds from the recorded
  absolutes exactly as today. The watermark is a new arm-set field on an
  existing type: the spec-083 `Neglect` precedent governs (same-build
  replay=live and pre-change snapshot round-trip are the contracts; old-build
  cross-hash equality never was).
- *The K dial*: `needs_checkpoint_minutes` joins `internal/sim/tuning.go`
  (spec 048 pattern — clamp-validated, event-logged as `sim.tuning_applied`,
  genesis-pinned per spec 057 so replay is immune to default retunes).
  Default 10; clamp [1, 60]; K=1 reproduces today's per-minute emission
  byte-for-byte (the escape hatch and a test fixture).

**Volume**: ~11.5k/day → ~1.2k/day checkpoints + crossings at K=10.

### D4 — Gru motion fully derived (`gru.moved` retired)

Gru stalk/prowl is already a pure function of (state, seed, tick): greedy
stalk over agent positions/protections, seeded prowl via
`rngAt(seed, "gru-prowl", tick, 0)`. The movement decision moves into
advancement (shared function; the executor's `gruStep` keeps attack/sighting
emission over advanced positions, ordered after agent steps within the beat
tick). `gru.moved` is no longer emitted; its arm stays forever for old logs.
`gru.emerged`/`withdrew`/`attacked`/`sighted` remain events (rare).
**Volume**: ~122k rows over 29 days → 0. RNG derivation joins the spec-092
audit (the "gru-prowl" purpose string and cadence become replay-load-bearing
for new logs — retune requires the format machinery).

### D5 — Downstream consumers

- **TUI**: replica calls `AdvanceTo` on tick pushes — map positions stay
  per-step smooth (exact, not interpolated). Digest rows added for
  `agent.path_started` ("Ash sets out for (x,y) (N tiles)") and
  `agent.path_truncated`; `agent.moved`/`gru.moved` rows KEPT for historic
  logs; `TestCatalogSweep` covers the additions. `docs/design/tui/` chronicle
  pages amended in-branch (spec 047 gate).
- **Mind**: replica calls `AdvanceTo` per absorb batch; `armEncounters`
  (mind.go:356, keyed on per-step `agent.moved`) becomes a first-adjacency
  sweep over the advanced replica on `agent.path_started`/tick progress —
  same adjacency moments, same pair cooldown; mind-side heuristic, outside
  replay scope.
- **Guardian**: watches/orders untouched (D3 keeps the type and crossing
  latency).
- **Chronicle/morgue**: fold through `Apply` ⇒ advancement free; narration
  inputs change only in ambient volume.

### D6 — Replay byte-identity strategy (the ruling-2 proof)

1. **Sighting equivalence harness** (`internal/sim/advance_test.go`): for
   scripted scenarios — single walk, two crossing walks (mutual sightings),
   path-tile speedup, mid-walk truncation, sleeper/waker beside a walk, pause
   mid-walk — build paired synthetic logs: (A) today's per-step `agent.moved`
   rows, (B) the coalesced `path_started`(+truncation) shape with identical
   geometry and timing. Assert `State.Marshal()` bytes equal at EVERY tick
   boundary (a `replayToTick`-style sweep), including each agent's canonical
   mental-map bytes (explored bitmap, facts, peer sightings with `Seen`
   ticks). This is FR-004's gate and the within-tick-order contract's
   enforcement.
2. **Needs equivalence**: K=1 world vs today's emission — byte-identical logs
   and states; K=10 world — derived per-minute values equal the K=1 world's
   folded values at every minute (decay exactness), crossings emitted at the
   same minute.
3. **Existing suites**: `TestDeterminismSameSeedSameTimeline` (incl. per-agent
   canonical map bytes), kill-9 recovery equivalence (mid-segment/mid-window
   snapshot + tail), `TestPre086ReplayByteIdentity`-class old-log fixtures
   replay green with pre-existing fields unchanged.
4. **Doctrine bookkeeping**: the spec-092 audit note gains rows for every
   constant the advancement path reads (witnessRadius — already grandfathered
   in the D2 class, decay constants, gru cadence/purpose string); definition
   sites get the "retune requires format bump" comment (spec 094 pattern).

## Wiki re-pins owed (in-branch, spec 069)

`event-log`, `event-types` + children touched (`event-types-agent-intents`
[agent.moved row], `event-types-agent-vitals` [needs row],
`event-types-mental-map`, `event-types-guardian-actions` [gru rows]),
`sim-state-reducer` + `sim-state-apply-agents`, `sim-state-agent-fields`,
`sim-state-reducer-replay-hazards` (audit additions), `mental-maps` /
`mental-map-model` / `mental-map-perception`, `executor` /
`executor-needs-survival` / `executor-tick-subsystems`, `gru`,
`deterministic-rng` (gru prowl derivation), `world-tuning` +
`world-tuning-boot-seeding` (K dial), `snapshots` (recovery interplay),
`tui-chronicle-feed`, `tui-client`/`tui-map-view` (replica advancement),
`agent-mind`/`mind-driver-triggers` (encounter arming), plus `testing-*`
children as suites land. `docs/player/` regenerated; `docs/design/tui/`
chronicle pages re-verified.

## Project Structure (predicted file footprint)

- NEW: `internal/sim/advance.go`, `internal/sim/advance_test.go`,
  `specs/104-ambient-event-coalescing/research.md` (T002 contracts note),
  `specs/104-ambient-event-coalescing/measurement.md` (T016 numbers).
- MODIFIED: `internal/sim/executor.go` (emission), `internal/sim/state.go`
  (new arms, Apply advancement hook, `PathSegment`/watermark fields),
  `internal/sim/payloads.go` (catalog), `internal/sim/gru.go` (motion moves
  out), `internal/sim/tuning.go` (K dial), `internal/sim/agents.go`
  (constants/comments), `internal/mind/mind.go` (replica advance + encounter
  sweep), `internal/tui/tui.go` + `internal/tui/digest.go` (replica advance +
  rows), tests beside each, wiki/player/TUI-design docs per the re-pin list.
