# Feature Specification: Ambient event coalescing — movement and needs ticks

**Feature Branch**: `task-176-ambient-event-coalescing`

**Created**: 2026-07-30

**Status**: Draft — **design fork RESOLVED** (operator rulings 2026-07-30,
recorded below and on the TASK-176 board card; the fork analysis is kept intact
as the decision record per card AC#2)

**Input**: TASK-176 + playtest-1 evidence (29 game-days): 1,011,063 events,
230 MB world.db; `agent.needs_changed` (332,752) + `agent.moved` (332,525) +
`gru.moved` (122,382) = 78% of all events. Constraint named on the card: the
replay/determinism doctrine (spec 092/TASK-75) — any coalescing must preserve
whatever the reducer needs for byte-identical replay.

## What exists vs what's missing (grounded 2026-07-30, this branch's base)

- EXISTS: the append-only event log as sole source of truth (in-schema
  `events_no_update`/`events_no_delete` triggers, `CheckContiguity` demanding
  seq 1..N — internal/store/schema.go, store.go), the single `Apply` mutation
  path shared byte-for-byte by live and replay (internal/sim/state.go), the
  spec-094 log-format stamp + load gate (internal/store/format.go), and both
  migration modes (snapshot-cut and translating — internal/world/migrate.go).
- EXISTS: the three ambient emitters at full per-beat granularity:
  - `agent.needs_changed` — per-game-minute heartbeat, one event per living
    agent per minute carrying absolute values (executor.go:192-270; 8 agents ×
    1,440 min/day ≈ 11.5k/day). Its reducer arm also rolls the spec-043
    trajectory anchor and the spec-083 neglect band anchors (state.go:1912).
  - `agent.moved` — one event per tile step, every `moveEveryTicks = 5` ticks
    while walking (2× on path tiles; executor.go:418). Its reducer arm does the
    spec-041 derived eventless bookkeeping: `markExplored` + `notePresence`
    (mutual peer sightings within `witnessRadius = 8`, tick-stamped) per step
    (state.go:1078-1093).
  - `gru.moved` — one event per gru step, every `gruMoveEveryTicks = 4` at
    night while emerged (gru.go:212-248).
- MISSING (the card's gap): any mechanism — at emission or offline — that keeps
  a month-scale world's ambient event volume proportional to *story*, not to
  ticks. Nothing thins these families; nothing compacts historical spans; the
  log grows ~8 MB/game-day at playtest-1 dials and every offline read (compare,
  tail, migrate, morgue regen) pays for it.

## Why this is not a logging tweak (the load-bearing coupling)

In this architecture events are the ONLY state mutation path ("adding a state
field means adding events that set it" — [[sim-state-reducer]]). Emitting fewer
events therefore means the live state itself updates less often, and everything
that reads live state feels it. The three families are load-bearing well beyond
their own reducer arms:

- `agent.moved` drives the mental-map perception substrate: per-step explored
  bitmap growth and mutual peer sightings ([[mental-map-model]] D2 — sightings
  are what `talk_to`/`seek` resolve against), the mind's first-adjacency
  encounter arming (`armEncounters`, internal/mind/mind.go:356), the spec-097
  arrival-observation gate (the step that lands ON the intent target), and the
  TUI live map's villager positions.
- `agent.needs_changed` drives the guardian's boot-seeded survival watches
  (internal/sim/guardian.go:151 declares `EventTypes: ["agent.needs_changed"]`)
  and the standing-order live matcher's per-villager hysteresis
  (internal/guardian/orders.go:240), the reflex danger bands, the spec-043
  trajectory arrows, and the spec-083 neglect anchors.
- `gru.moved` keeps `State.Gru` current for stalking/protection/sighting
  checks read every tick, and the TUI's threat rendering.
- All three have digest-grammar rows (internal/tui/digest.go) swept by
  `TestCatalogSweep`; chronicle narration and guardian digests query the
  events table by tick/type.
- Persisted seq references exist: `cog.thought` records the arming stimulus
  event's seq as its causality edge, and memory identity/embedding references
  ride event seqs — any scheme that renumbers or removes historical rows
  breaks recorded references, not just row counts.

## THE DESIGN FORK — **RESOLVED (operator rulings, 2026-07-30)**

Two mutually exclusive architectures could deliver AC#1. Both were costed below
and presented at the sweep's operator checkpoint. **The operator ruled; the
rulings are binding** (also recorded on the TASK-176 board card), and the
analysis below is retained verbatim as the decision record per card AC#2:

1. **Arm A (emission-shape change) ADOPTED.** Arm B rejected — it inverts the
   log-is-truth doctrine, and seq renumbering is near-disqualifying.
2. **Movement sighting fidelity (sub-fork 1): EXACT per-step.** Deterministic
   segment advancement or baked sighting payloads — mental-map explored-bitmap
   growth, mutual tick-stamped peer sightings, and `talk_to`/`seek`/encounter
   behavior must be byte-identical to today's per-step emission. This is the
   hard sub-problem; design effort concentrates here (plan.md).
3. **Needs depth (sub-fork 2): interval + crossings.** Emit every K
   game-minutes plus immediately on band crossings (guardian survival watches
   and standing-order hysteresis see crossings at today's latency).
   Needs-as-curve rejected. K becomes a world-tuning dial (spec 048 pattern) —
   decided: `needs_checkpoint_minutes`, genesis-pinned per spec 057, K=1
   reproducing today's emission as the escape hatch (recorded in plan.md D3).
4. **Log-format stamp (sub-fork 3): NO bump** — additive vocabulary per the
   spec-097 `place_observed` precedent.
5. **Old-world relief (sub-fork 4): OUT of scope** — the existing snapshot-cut
   migrate covers it. Recorded as an explicit non-goal below.

### Arm A — emission-shape change (coalesce at the source)

Change what the executor emits so ambient state advances in coarser, fully
payload-carried strokes; the log records fewer, richer events.

- **A-move (path segments):** an intent walk emits one segment event at
  departure — planned path (BFS output is already deterministic), start tick,
  cadence — plus a closing/truncating event on arrival, re-route, or
  interruption. Position becomes a derived read: `State` stores the in-flight
  segment and readers compute the tile from (segment, tick). A typical walk of
  10–30 tiles collapses from 10–30 events to 2. The gru gets the same shape at
  smaller scale (nightly prowl legs), or — cheaper — a sampled variant, since
  no perception bookkeeping hangs off `gru.moved`.
- **A-needs (regime-change emission):** the per-minute heartbeat stops emitting
  unchanged-regime decay. Either (i) bounded-interval thinning — emit absolute
  values every K game-minutes AND immediately on any band/threshold crossing
  (danger bands, near-death, zero) — or (ii) needs-as-curve: emit an event only
  when the decay regime changes (sleep/wake, warmth source, shelter, night
  boundary, cold snap, eating, attack), payload carrying anchor values + rates
  (emitter-computes, spec-092-safe), readers deriving the current value from
  (anchor, rate, tick). (i) is a ~K-fold cut with minimal rework; (ii) cuts
  1,440/agent/day to dozens but reshapes `Agent.Needs` and every reader.

**Costs / risks:**

1. **This is the hard, architectural arm** (already tiered Opus 4.8 on the
   card): position and/or needs become time-derived state, and the derived
   mental-map bookkeeping must be re-specified. Per-step `markExplored`/
   `notePresence` cannot simply fire at segment application time — sightings
   are *mutual* functions of two agents' concurrent walks with tick-stamped
   `Seen` values, so exact preservation requires deterministically advancing
   in-flight segments as other events apply (or baking the sighting outcomes
   into the closing payload), while deliberate coarsening (sightings at
   segment endpoints only) is cheaper but is a behavior change to
   `talk_to`/`seek` resolution and encounter arming that needs paired-seed
   re-measurement. This sub-fork is real and is listed for the operator below.
2. **Live-consumer latency:** guardian survival watches and standing orders
   match `agent.needs_changed` in the live absorb path; thinner emission delays
   trigger latency by up to the emission interval unless crossings always emit
   immediately (variant (i) preserves this by construction; variant (ii) needs
   the matcher re-pointed at derived reads).
3. **Forward-only relief:** Arm A shrinks nothing retroactively — playtest-1's
   230 MB stays 230 MB. (Mitigation: the EXISTING snapshot-cut migration
   already offers old worlds a fresh log with history archived; see hybrid
   note below.)
4. **TUI smoothness:** the live map currently animates per-step; under
   segments the replica either interpolates from the declared path (small TUI
   change) or villagers visibly jump.
5. **Migration/format surface:** additive new types + retired emission of old
   types — see Migration implications; per doctrine this AVOIDS a translating
   or snapshot-cut migration for old logs, with one flagged sub-decision on
   the log-format stamp.

### Arm B — offline compaction (keep emission as-is)

Live emission unchanged; a periodic or operator-invoked offline pass rewrites
historical spans of the log, collapsing ambient runs whose fold is already
covered by a verified snapshot (snapshot-anchored span splice — effectively a
partial snapshot-cut INSIDE the log, on the `world.migrated` wholesale-replace
precedent).

**Costs / risks:**

1. **Doctrine inversion:** [[event-log]] and [[snapshots]] pin "the log is the
   source of truth; snapshots merely accelerate — all snapshots can be
   discarded at the cost of replay-from-genesis." Compaction makes snapshots
   (or spliced state events) *authoritative* for compacted spans and forfeits
   replay-from-genesis within them: `daemon.replayToTick` to a mid-span cutoff
   (the spec-043 determinism harness, deliberately snapshot-free), the
   morgue's deterministic genesis replay fold, and the fork ceremony's
   event-prefix streaming all assume the full log exists.
2. **Mechanical resistance is by design:** the in-schema append-only triggers
   mean compaction must rebuild the DB via the translating-mode swap pattern;
   `CheckContiguity`'s 1..N demand means removed rows force either renumbering
   — which breaks every persisted seq reference (`cog.thought` causality seqs,
   memory/embedding seq identity, snapshot `seq` anchors) — or a relaxed
   contiguity contract that tolerates manifest-declared compacted spans. Both
   are surgery on load-bearing integrity gates.
3. **Downstream history consumers degrade:** chronicle/digest tick/type
   queries, TUI inspect mode, morgue regeneration, and the spec-092 audit
   surface all see fewer rows for compacted spans; every such consumer needs
   an explicit "compacted span" posture.
4. **Byte-identity must be re-scoped:** determinism doctrine would need an
   amendment — byte-identity guaranteed at and beyond compaction boundaries
   (provable: fold(original span) == spliced state), forfeited within them.
5. **Benefit that Arm A lacks:** existing bloated worlds shrink in place, zero
   live-path or behavior change, event vocabulary untouched, and recent
   history (spans younger than the compaction horizon) stays fully inspectable.

### Hybrid note (weakens Arm B's unique advantage)

Arm A plus the **already-shipped** snapshot-cut migration gives old worlds
relief today: `promptworld migrate`-style archive-and-fresh-log (history
archived as `world.vN.db`, never deleted) is precisely "compaction of
everything before now," using ratified machinery and existing doctrine. If the
operator accepts archive-granularity relief for old worlds (rather than
Arm B's sliding-window in-place compaction), Arm B's remaining unique value is
keeping full-fidelity recent history INSIDE one continuously-running world —
worth naming, because it substantially changes the arms' relative cost/benefit.

### Recommendation (as presented at the checkpoint — adopted by ruling 1)

**Arm A**, sub-scoped as: A-needs variant (i) (bounded-interval thinning with
immediate band-crossing emission — the cheap 80% of the needs win with no
consumer latency regression), A-move path segments with the
exact-vs-coarsened-sightings sub-fork explicitly ruled on, and gru sampling.
Rationale: Arm A spends effort on new code paths while leaving every existing
doctrine intact; Arm B spends effort *amending* doctrine (log-is-truth,
contiguity, replay-from-genesis) and touching integrity gates whose whole
purpose is to resist exactly this operation — and its unique benefit (in-place
relief for old worlds) is substantially covered by the existing snapshot-cut
migration. The card's own constraint framing ("this may land as
emission-shape change, not lossy compaction") points the same way. Cost
accepted: Arm A is the harder implementation and gives no retroactive shrink.

### Additional forks discovered (all resolved by rulings 2-5 above)

1. **A-move sighting fidelity:** exact per-step equivalence (segments advanced
   deterministically / sightings baked into closing payloads) vs deliberate
   coarsening to segment endpoints (cheaper; small behavior change to
   `talk_to`/`seek` and encounter arming; needs paired-seed evidence).
2. **A-needs depth:** bounded-interval thinning (i) vs needs-as-curve (ii) —
   several-fold vs orders-of-magnitude, at very different rework cost.
3. **Log-format stamp posture for additive-but-load-bearing vocabulary:** spec
   097's D5 precedent says additive types take NO format bump — but a NEW log
   whose *movement itself* rides new types silently mis-replays on an OLD
   binary (unknown-type no-op), the exact hazard the spec-094 newer-refusal
   gate exists for. Bump `store.LogFormatVersion` (old logs then need a
   stamp-only translation; new logs refuse on old binaries) vs follow the
   097 no-bump precedent (downgrade-replay remains unguarded, as it already
   is for `agent.place_observed`).
4. **Old-world relief:** whether playtest-1-class worlds get a one-time
   archive-and-fresh-log pass (existing machinery) — orthogonal to the arm
   chosen, but it changes what AC#1's "month-scale world" means for worlds
   that already exist.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A month-old world stays snappy (Priority: P1)

As a player, I want a month-old world to stay quick to open, compare, and
migrate — not drag a multi-hundred-megabyte log behind it — so long-running
worlds remain a pleasure rather than a liability.

**Acceptance Scenarios**:

1. **Given** a paired-seed measurement run at playtest-1 dials, **When** the
   fixed build runs the same game-days, **Then** the ambient families
   (`agent.needs_changed` + `agent.moved` + `gru.moved` or their successors)
   land at least several-fold fewer rows per game-day than the baseline.
2. **Given** the reduced log, **When** projected to 29 game-days, **Then**
   total events and on-disk size are several-fold below the 1,011,063 /
   230 MB baseline.

---

### User Story 2 - The log reads as story, not ticks (Priority: P1)

As an operator mining a playtest, I want the event log dominated by meaningful
events, so tailing, filtering, and forensic replay stop wading through
movement and decay ticks.

**Acceptance Scenarios**:

1. **Given** the fixed build, **When** tailing a running world, **Then** the
   three ambient families no longer form the overwhelming majority of rows
   (playtest-1 baseline: 78%).
2. **Given** the chronicle and digest surfaces, **When** rendering a world
   recorded under the new shape, **Then** every persisted type (including any
   new coalesced types) has a digest row (`TestCatalogSweep` green) and
   narration/feed quality is not degraded.

---

### User Story 3 - Replay determinism survives untouched (Priority: P1)

As the daemon (and the determinism harness), I want old logs to replay to
byte-identical states and new worlds to remain byte-deterministic, so the
coalescing changes volume without ever changing truth.

**Acceptance Scenarios**:

1. **Given** an existing pre-change log, **When** replayed on the fixed build,
   **Then** state hashes are byte-identical to the pre-change build's replay
   (retained reducer arms; no rename, no re-derivation change).
2. **Given** a fresh world on the fixed build, **When** run twice at the same
   seed, **Then** `TestDeterminismSameSeedSameTimeline` and the replay/recovery
   suites stay green, including each agent's canonical mental-map bytes.
3. **Given** whichever arm is chosen, **Then** the recorded design decision
   states explicitly how the spec-092 emitter-computes doctrine and the
   spec-094 name/format doctrine are satisfied (card AC#2).

---

### User Story 4 - Perception and vigilance keep working (Priority: P2)

As a villager in the game, I still learn the terrain I walk, still notice who
I pass, and my guardian still notices when I am in danger — coalescing the
bookkeeping must not blind anyone.

**Acceptance Scenarios**:

1. **Given** an agent walking under the new shape, **When** the walk completes,
   **Then** its explored bitmap and peer sightings (positions AND `Seen`
   ticks) are byte-identical to what today's per-step emission would have
   produced (ruling 2: exact), and the spec-097 intent-completing arrival
   observation still fires exactly once.
2. **Given** a villager's need crossing a danger band, **When** the crossing
   occurs, **Then** an `agent.needs_changed` event is emitted at that very
   minute beat (ruling 3), so the guardian's survival watches and standing
   orders trigger with no worse latency than today.
3. **Given** two agents becoming first-adjacent mid-walk, **When** the
   encounter-arming path evaluates, **Then** conversation arming fires for
   the same adjacency moments per-step emission would have armed (pair
   cooldown preserved).

### Edge Cases

- Walk truncation: a wall built mid-segment, an absorb re-arm, a hail pause,
  death, or a `guardian.entity_moved` teleport must close/truncate an in-flight
  segment correctly (Arm A) — the closing event is the record of what actually
  happened, never the plan.
- `guardian.time_snapped` rebase: any new tick-anchored payload/state fields
  (segment start ticks, needs anchors/rates) must be classified SHIFT/KEEP in
  the `rebaseTicks` taxonomy ([[guardian-miracles]]).
- The live playtest world is never migrated or opened by test tooling in this
  task; measurement uses disposable paired-seed worlds (the measurement-run
  recipe), preserved for review.
- Quiet-tick clock behavior is unchanged: ticks were never event-sourced;
  coalescing must not accidentally introduce per-tick heartbeats elsewhere.
- Zero-length walks and zero-delta needs beats: whatever the arm, "nothing
  happened" must still land no event (the current sweep's settled-map
  no-emission precedent).

## Requirements *(mandatory)*

### Non-goals (explicit, per ruling 5)

- **Old-world relief is OUT of scope.** This feature changes emission going
  forward only; existing bloated worlds (playtest-1's 230 MB) are served by
  the EXISTING snapshot-cut migration (archive-and-fresh-log) whenever the
  operator chooses to invoke it. No compaction, no automatic migration, no
  touching of existing world databases ships in this task.
- No offline compaction machinery of any kind (Arm B rejected, ruling 1).
- No log-format-version bump and no new migration mode (ruling 4).

### Functional Requirements

- **FR-001**: Ambient event volume for the movement/needs/gru families is
  reduced several-fold per game-day at playtest-1 dials, measured on a
  paired-seed baseline-vs-fixed run and recorded with the numbers.
- **FR-002**: Existing logs replay byte-identically on the fixed build: no
  persisted type is renamed, no reducer arm changes how it re-derives from a
  recorded payload; all determinism/replay/recovery suites stay green.
- **FR-003**: Every downstream consumer keeps working: chronicle narration,
  digest grammar (rows for any new types; `TestCatalogSweep`), TUI live map
  and inspect mode, guardian survival watches and standing-order matching,
  the mind's encounter arming and arrival-observation reconciliation.
- **FR-004**: The mental-map contract is preserved EXACTLY (ruling 2):
  explored-bitmap growth, mutual peer sightings (positions and `Seen` ticks),
  and the spec-097 arrival observation are byte-identical to per-step
  emission, proven by an equivalence harness comparing per-step vs coalesced
  logs of the same walks at every tick boundary.
- **FR-005**: The adopted design (Arm A, rulings 1-5) is recorded as a design
  decision — this spec's resolved fork section, mirrored into the wiki's
  event-log/reducer doctrine notes — explicitly addressing the spec-092
  emitter-computes doctrine and the spec-094 format doctrine (card AC#2).
- **FR-006**: Whatever new payloads exist are fully baked at emission
  (emitter-computes): no reducer arm may re-derive a coalesced outcome from a
  mutable gameplay constant (cadences, radii, decay rates read at apply time
  are forbidden — they ride the payload).
- **FR-007**: `go test -race ./...` green; the wiki notes whose pinned sources
  this touches are re-pinned in-branch per the wiki-in-PR lifecycle.
- **FR-008**: Needs emission follows ruling 3: an `agent.needs_changed` event
  (existing type, existing absolute-value payload) every K game-minutes per
  living agent AND immediately on any band/threshold crossing (spec-062
  danger bands, near-death, zero). K is a spec-048 world-tuning dial
  (`needs_checkpoint_minutes`), clamp-validated, genesis-pinned per spec 057,
  with K=1 reproducing today's per-minute emission exactly.
- **FR-009**: `gru.moved` emission is retired for new worlds (gru motion
  becomes derived, deterministic from seed + state + tick); its reducer arm
  is retained forever so old logs replay unchanged.

### Key Entities

- **Ambient families**: `agent.needs_changed`, `agent.moved`, `gru.moved` —
  and, under Arm A, their coalesced successor types (additive vocabulary).
- **Derived movement bookkeeping**: explored bitmap + peer sightings — the
  eventless (state, event) functions whose fidelity contract is FR-004's
  subject.
- **Compacted span** (Arm B only): a log range replaced by a spliced,
  snapshot-verified state event, with declared boundaries.

## Determinism doctrine *(explicit, per card AC#2)*

- The governing law is spec 092/TASK-75 ([[sim-state-reducer-replay-hazards]]):
  payloads carry outcomes; reducers copy verbatim; re-deriving from mutable
  constants is the audited exception set, never extended. Every coalesced
  payload in this feature is emitter-computed — segment paths, per-step
  schedules or baked sighting outcomes, needs anchors/rates/absolutes — so a
  later retune of `moveEveryTicks`, decay curves, or radii can never change
  what an old log replays to.
- Byte-identity is per-log and per-build: old logs must fold to identical
  states on the fixed build (FR-002); new worlds must be self-deterministic
  (same seed, same timeline, same bytes — including canonical mental-map
  bytes). Baseline-vs-fixed worlds at the same seed will legitimately record
  DIFFERENT event streams; the paired-seed comparison is a volume measurement,
  never a byte comparison across builds.
- Arm A must additionally prove that any derived-position/derived-needs read
  is a pure function of (state, tick) — the [[mental-map-model]] D2 discipline
  extended — and that replay-to-cutoff (`daemon.replayToTick`) lands the same
  derived values live execution held at that tick.
- Arm B (REJECTED, ruling 1 — retained for the record) would instead have
  AMENDED the doctrine: byte-identity guaranteed at and
  beyond compaction boundaries (provable as fold(original span) == spliced
  state at the boundary), explicitly forfeited within compacted spans — a
  doctrine change the operator must ratify, not a detail.

## Migration implications *(explicit, per dispatch)*

- **Arm A is additive vocabulary**: new event types + retired emission of old
  types. The old arms (`agent.moved`, `agent.needs_changed`, `gru.moved`)
  remain in `Apply` forever — no rename, no re-derivation change — so per the
  spec-094 decision rule NO translating migration and NO snapshot-cut is
  required: old logs load and replay unchanged. The card's question ("check
  whether additive new types + retired emission of old types avoids a format
  bump") resolves YES on current doctrine. **RULED (ruling 4): no
  `store.LogFormatVersion` bump** — the spec-097 `place_observed` precedent
  governs (downgrade-replay remains unguarded for additive vocabulary, as it
  already is today).
- **Arm B requires format surface either way**: a compacted log must declare
  itself (stamp or manifest marker) so tools know spans are compacted; the
  rebuild rides the translating-mode swap pattern (build at
  `world.db.translating`, verify, archive, swap, bump last), and either seq
  renumbering (breaks persisted seq references — effectively disqualifying)
  or a relaxed `CheckContiguity` contract with declared gaps.
- **State-shape changes** (Arm A's in-flight segment field; needs-as-curve's
  anchor/rate fields) follow the `omitempty` pointer precedent
  ([[mental-map-model]]: pre-change snapshots round-trip byte-identically) so
  the world-manifest `format_version` need not bump for the state additions
  alone; if the chosen sub-scope cannot satisfy that, the manifest bump +
  snapshot-cut chain is the fallback and must be called out in plan.md.
- **Existing worlds**: never silently touched. Old-world relief is a
  NON-GOAL (ruling 5) — the operator-invoked archive-and-fresh-log pass
  using existing snapshot-cut machinery covers it, outside this task.

## Success Criteria *(mandatory)*

- **SC-001**: On a paired-seed measurement run at playtest-1 dials, the
  ambient families' rows per game-day drop ≥4× vs baseline, and projected
  29-game-day totals land several-fold under 1,011,063 events / 230 MB —
  numbers recorded on the task.
- **SC-002**: A pre-change log replays on the fixed build to state hashes
  byte-identical to the pre-change build's replay; determinism, recovery,
  and replay-to-cutoff suites green.
- **SC-003**: `TestCatalogSweep` green with the new types; chronicle, TUI map,
  guardian watch, and encounter-arming behavior verified per the exact
  fidelity contract (FR-003/FR-004); the sighting equivalence harness green.
- **SC-004**: The operator's fork resolution (rulings 1-5, recorded above) is
  mirrored into the wiki doctrine notes, explicitly addressing determinism
  and migration (card AC#2).

## Assumptions

- The operator checkpoint is RESOLVED (rulings 1-5, 2026-07-30, recorded in
  the fork section above and on the board card); plan.md and tasks.md are
  authored against those rulings, which are binding.
- Volume arithmetic grounding the several-fold claim: needs at 8 agents ×
  1,440 min/day ≈ 11,520/day (thinning at K=10 with crossings ⇒ ~10×);
  movement at ~1,434 steps/agent/day with typical walks of 10–30 tiles
  (segments ⇒ ~5–15×); gru at ~4,200/night (sampling/legs ⇒ ~5–10×). These
  are estimates for scoping, not commitments; SC-001's measurement is the
  contract.
- The measurement run follows the established recipe (paired seed-1337 worlds,
  harsh dials, worlds preserved for review); the live playtest world is never
  used as a fixture.
- Implementation tier is Opus 4.8 per the card's recorded rubric ruling
  (cross-package, doctrine-adjacent, reducer/determinism surface).
