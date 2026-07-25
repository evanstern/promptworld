---
name: testing-strategy
description: How correctness is proven — unit determinism harness, in-process IPC integration, binary-level e2e quickstart scenarios, race detector
kind: pattern
sources:
  - internal/sim/sim_test.go
  - internal/sim/migrate_test.go
  - internal/sim/whole_feature_test.go
  - internal/sim/miracles_test.go
  - internal/sim/guardian_test.go
  - internal/sim/origin_test.go
  - internal/sim/belief_evidence_test.go
  - internal/sim/belief_decay_test.go
  - internal/sim/belief_reinforced_test.go
  - internal/sim/wall_test.go
  - internal/sim/axe_test.go
  - internal/sim/path_speed_test.go
  - internal/sim/run_end_test.go
  - internal/sim/morgue_test.go
  - internal/sim/grave_test.go
  - internal/sim/gru_test.go
  - internal/sim/toolcheck_test.go
  - internal/sim/curriculum_test.go
  - internal/sim/intentlog_test.go
  - internal/sim/needsanchor_test.go
  - internal/sim/journal_test.go
  - internal/sim/memory_test.go
  - internal/sim/yield_state_test.go
  - internal/sim/day_warmth_test.go
  - internal/sim/night_search_test.go
  - internal/sim/thrash_regression_test.go
  - internal/sim/reflex_matrix_test.go
  - internal/sim/needs_recovery_test.go
  - internal/mind/context_test.go
  - internal/daemon/context_replay_test.go
  - internal/world/migrate_test.go
  - internal/ipc/ipc_test.go
  - internal/mind/replay_test.go
  - internal/mind/provenance_test.go
  - internal/mind/belief_read_sites_test.go
  - internal/guardian/guardian_test.go
  - internal/guardian/guardian_gaps_test.go
  - internal/guardian/orders_test.go
  - internal/guardian/charter_observed_test.go
  - internal/guardian/stage_test.go
  - internal/daemon/curriculum_test.go
  - internal/worlds/unlocks_test.go
  - internal/world/world_test.go
  - cmd/promptworld/stages_test.go
  - internal/mind/epilogue_test.go
  - internal/persona/persona_test.go
  - e2e/daemon_e2e_test.go
  - e2e/determinism_e2e_test.go
verified_against: 0b42d204718819d773be44c40ed26d42aba055f8
---

# Testing strategy

The spec's success criteria (determinism, crash-lossless resume, detach-isolation)
are only provable by tests, so the suite is layered: pure-logic harnesses at the
package level, an in-process integration layer, and binary-level e2e that execs the
real `promptworld`.

## How it works

**Unit determinism harness** (`internal/sim/sim_test.go`): `driveTicks` replicates
the loop's semantics minus the real-time scheduler — commands injected at exact tick
boundaries, terrain threaded through exactly as the live loop does. Now proven over
the full [[executor]]: 30k–40k-tick determinism and replay harnesses, plus behavior
suites — multi-step intent chains with zero input (AC#1), needs decay + self-feeding
and starvation death with recorded cause (AC#2), night warmth mechanics and exposure
death (AC#3), and a two-day unattended village survival run on multiple seeds.
`TestDeterminismSameSeedSameTimeline` additionally diffs each agent's canonical
mental-map bytes across two same-seed runs (spec 041, T010) — the state hash
already covers them, but a targeted per-agent diff localizes a map-determinism
regression instead of just failing the whole-state comparison ([[mental-maps]]).
(Terrain generation has its own determinism/AC suite in `internal/worldmap`, covered
by [[worldmap-generation]].)

Spec 012 and spec 013 each added their own fixture suite spanning both save-format
packages, all in [[world-migration]]'s territory: `internal/sim/migrate_test.go`
builds representative v1 and v2 states and proves both pure transforms'
carry/reset/re-place/spill rules directly, including a v1 fixture that chains both
transforms (1→2→3) in one call; `internal/world/migrate_test.go` drives the full
`Migrate(dir)` ceremony end-to-end against on-disk v1 and v2 fixture worlds (happy
path, replay-from-zero-snapshots determinism, the already-migrated and
already-current guards, uncovered/tolerated event tails, a running-daemon refusal)
for both the v1→v2 and v2→v3 steps. Spec 041 (T009) adds the v3→v4 leg on
both sides, [[mental-maps]]'s knowledge-grant transform: `migrate_test.go`'s
`TestTransformV3GrantsKnowledge` proves the pure `TransformV3State` carries
people and land verbatim while granting each LIVING agent explored terrain
around its position plus witnessed facts for every current structure/pile
(natives, not strangers), leaves a DEAD agent an empty but non-nil map (so a
genesis-seeded state and a migrated one agree on map presence), and mutates
neither its input nor the map argument; `TestTransformV3ChainReducerReplay`
proves the transform's output replays byte-identically through
`world.created` → `world.migrated` from genesis. `world/migrate_test.go`'s
`TestMigrateV3HappyPath` drives the same ceremony end-to-end against an
on-disk v3 fixture world (manifest bump to v4, `world.v3.db` archived, the
covering snapshot's agents each holding the grant), and
`TestMigrateV3ReplayDeterminism` proves deleting every snapshot and
rebuilding from genesis reproduces the post-migration snapshot byte-for-byte,
maps included.

`internal/sim/whole_feature_test.go` carries several byte-identity suites (SC-004/SC-005):
the original spec-012 run, a single scripted-agent script chaining every
resources/food/crafting event kind (quarrying — five bare quarries at
`quarryYieldBare` (1) each, since spec 032 T014 split the old flat `quarryYield`
(2) into bare/axe-assisted tiers and dropped the bare yield to 1 — water, the full
craft chain, both cook stations, bathing, refueling, a spear breaking, a fire
burning out) — rebalanced under spec 013's bulk cap (24) to consume-as-it-goes
rather than hoard a large seeded larder — that replays from genesis to a
byte-identical state hash; and a spec-013 storage suite
(`TestReplayByteIdentityWholeFeatureStorage`) exercising every new
013 event type in one run — `agent.dropped`, `agent.picked_up`, `agent.deposited`,
`agent.withdrew` (both an owner fetch and a non-owner theft with its full companion
batch: `social.chest_taken`, a reason-`theft` `social.relation_changed`, and owner +
witness `agent.memory_added`), `sim.food_rotted`, `agent.built{kind: chest}`, and a
death spill — that also replays to a byte-identical hash. The same file also proves
every new 013 event type no-ops under a pre-013 reducer stub (the unknown-type
convention: an event type the reducer's switch doesn't match falls through to a
total no-op, never an error), so old logs stay safely replayable by builds that
predate a given event kind. `TestReplayByteIdentityWallsAxesPaths` (spec 032 T021,
quickstart scenario 7) is the walls/axes/paths counterpart: one scripted session
chains `craft_axe`, an axe-assisted chop, a `build_wall_plank`, a full
`demolish` (chip then destroy), and a `build_path`, asserts every required event
(`agent.crafted`, `agent.chopped`, `agent.built`, `agent.wall_chipped`,
`agent.wall_destroyed`, and — since spec 038 — `agent.build_failed`) actually
occurred, then replays from genesis (log only) to a byte-identical state hash.
Spec 038 extends the same session with a build that fails LOUDLY: a wall
injected onto the just-built path tile (an unbuildable reserved site) resolves
via `agent.build_failed` plus a paired situated failure memory rather than a
bare `agent.intent_done`, and both new events replay byte-identically alongside
the rest of the session. `TestPre032SnapshotLoadsUnchanged` (spec 032 T021,
research R7) proves a pre-032-shaped snapshot (no structure `hp`, no inventory/pile
`axes`) round-trips unmarshal→marshal byte-identically — the new fields are
additive `omitempty`, so an old save loads unchanged with no format-version bump.
Together these prove: same seed + same command timeline
over 30k ticks → byte-identical event sequences and equal state hashes; different
seeds diverge; replaying the logged events over genesis (then re-living the quiet
tail) reproduces the live state hash exactly; the day/night cycle behaves (nobody
moves at night).

**Loop-era replay determinism** (`internal/mind/replay_test.go`): a real `Loop` +
`loopMind` pair proves live-vs-replay byte identity above the pure-reducer layer.
`TestLoopRunReplayByteIdenticalSC002` (TASK-52, SC-002) drives cognitions, tool
calls, and a muse through the real loop, then asserts a from-genesis replay
reproduces the identical `State` with the model seam invoked zero times.
`TestJournalAndSituatedReplayByteIdentical` (spec 019 US4, T019, SC-003) extends
this to the grounded-memories feature: injected situated memories (place/why,
place/conv), a journal write→write→delete cognition sequence, and a scripted
over-budget write that the gate refuses (landing nothing but a rejected
`cog.tool_call`) — genesis replay reproduces the identical `State` *and*
byte-identical rendered `soul.md`/`journal.md` over both live and replayed
state, with the model seam invoked exactly once per live cognition and zero
times during replay.

**IPC integration** (`internal/ipc/ipc_test.go`): a real loop + server + store on a
temp world. Proves: status round trip <2 s; subscribe-from-zero delivers strictly
consecutive seqs; abrupt disconnects and wire garbage leave the loop ticking;
commands are idempotent and land in the log as events; the `state` command's
coherence contract holds (no push predates the snapshot's `last_seq`, and a replica
built from it applies subsequent pushes cleanly — the [[tui-client]] pattern); and
`llm_call` routes through a live [[llm-orchestrator]] while a killed inference
endpoint leaves the loop ticking (the package's own suite covers routing, metering,
ceiling refusal, and circuit recovery against httptest mock providers). Spec 028
(adaptive throttle) adds its own status-fold coverage here: a scripted
`Governor` fake proves debt/jobs fold into `StatusData.Clock` exactly like the
LLM snapshot, a no-governor world reports zero governor values, and a
byte-shape test pins the three new fields `omitempty` (a zero status marshals
with none of them present); a `Loop.Govern`-driven test proves status reports
both the effective and requested speed while governed and that a player
`set_speed` below the governed notch collapses `RequestedSpeed` back to empty;
and a regression test pins that `set_speed max` is still refused with an LLM
configured (FR-012) while `32x` is accepted, unchanged by the governor. Spec
035 (calibration UX) adds its own `set_speed`-warning coverage here: an
uncalibrated world raised into suppressing territory (32x) warns naming the
suppressed classes and the calibrate command while the speed change still
applies, and the same world dropped to a non-suppressing notch (4x) carries
none; a calibrated world (`SeedCalibration` with a profile) gets no warning
even at a speed that would suppress at bootstrap estimates (the gate is seed
state, not raw arithmetic); a no-LLM world never warns at any speed; the
pre-035 `set_speed max` refusal still precedes the warning and carries no
`StatusData` at all; `status`/`pause`/`resume` never carry the warning even
on an uncalibrated world sitting at a suppressing speed; and a byte-shape
test pins that a zero `StatusData` omits `warning` entirely. Spec 037 (live
horizon surface) adds its own horizon-composition coverage here: an
uncalibrated 32x world's `status` reply carries one `Horizon` entry per
watched class in `WatchedClasses` order, each with a non-empty verdict
string and `Calibrated=false`, with planner/conversation suppressed and
meeting thinking at the bootstrap seeds; a calibrated world (`SeedCalibration`)
is still fully INCLUDED with `Calibrated=true` (contrast the `Warning` gate,
which excludes it) and nothing suppressed at 32x on a fast rig; composing
directly at `clock.SpeedMax` suppresses every included class with Route's
uncapped phrasing; a no-LLM world's reply carries no `horizon` key at all
(byte-identical to pre-037); and `RecordSuppression` calls surface as each
entry's `SuppressedCount`, keyed correctly per class with an unwatched class
(`chronicle`) never leaking an extra entry. A byte-shape test pins that a
zero `StatusData` omits `horizon` entirely. Spec 039 (teaching-world speed
posture) adds its own posture coverage here: an uncalibrated teaching world
set to 32x applies the speed and carries an `above teaching posture 16x`
override naming the router's verbatim suppressed-class arithmetic and the
plain-language degrade consequence for each; at or below the 16x posture the
same world carries no warning text at all; a non-teaching world never carries
the posture text (only the unchanged spec 035 uncalibrated leg); an
overshooting uncalibrated teaching world gets BOTH texts, posture first then
the calibrate prompt, newline-joined; a CALIBRATED teaching world
(`SeedCalibration`) still warns on override — using the measured seconds-per-
point in its arithmetic and carrying no uncalibrated text — proving the soft
cap is independent of calibration state; the `set_speed max` refusal still
precedes any posture text for a teaching LLM world; and on the status side, a
calibrated teaching world's `status` reply carries `Posture{Rung, Calibrated:
true}`, an uncalibrated one `Posture{Rung, Calibrated: false}`, a non-teaching
world's reply omits the `posture` key entirely (byte-shape pinned on a zero
`StatusData` too), and a teaching-but-pure-sim world (no orchestrator) also
carries no posture block. Large-reply
behavior (TASK-19) is proven against a `fakeDaemon` wire harness that speaks the
protocol from canned replies: a >1 MiB `state` payload round-trips; a reply over
the 64 MiB cap is substituted server-side with an actionable `reply too large`
error (via `net.Pipe` against `session.writeResponse`); and both the substituted
error and a raw over-long line surface promptly as `ErrReplyTooLarge` — never a
hang or silent scanner death.

**E2E** (`e2e/`): `TestMain` builds the binary once and sets a package-wide
hermetic `PROMPTWORLD_HOME` (a temp dir) before running — every subprocess the
package execs inherits it, so no test can write the developer's real
`~/.promptworld` registry (TASK-49; `manager_e2e_test.go`'s `isolatedHome`
layers a per-test override on top). Worlds drop `llm.json`
right after `new` so they are pure-sim — a precondition for `speed max` under
the TASK-20 policy. Scenarios mirror
`specs/001-world-daemon/quickstart.md` — A: always-on + detach-is-not-pause; B:
pause freezes the clock, compression ratios hold (loose tolerances over short
windows; the spec's 5% applies to 5-minute windows); C: kill -9 → lossless resume
within 10 s, restart-while-paused wakes paused, graceful stop idempotent; E: a
`cp -R`'d stopped world runs. `determinism_e2e_test.go` compares two same-seed
daemons' sim histories over their common tick prefix (past tick 25000, so the
full day-1 [[governance]] meeting cycle is inside the compared window),
excluding wall-dependent `daemon.*`/`clock.*` bookkeeping.

**Miracle cost derivation** (`internal/sim/miracles_test.go`, spec 021):
`TestMiracleCostDerivedFromTool` pins `sim.miracleCost` ≡
`tool.MiracleCostsByEvent()` — the sim-side enforcement table is a derivation of
the registry's single authoritative price source, not a mirror, so a price edit
cannot half-propagate ([[tool-registry]], [[guardian-miracles]]).

**Miracle reducer suite** (`internal/sim/miracles_test.go`, spec 016,
[[guardian-miracles]]): per-arm coverage for all four types — move (villager/
structure-whole/pile-merge, impassable/absent-source rejection), remove
(villager rejected, chest spill, pile destruction, terrain routing), grant
(happy path, over-cap whole-reject, unknown kind, dead villager, non-positive
qty, spear shape), and time-snap (forward-only, duration-preserving,
whole-day-no-drift, mints-no-charges-across-skipped-boundaries, while-paused);
plus charge doctrine (insufficient-charge rejection, gratis waives only the
charge, gratis is logged visibly), and `TestRebaseTaxonomyComplete` — the build
fails if a new tick-anchored `int64` field appears anywhere in the state tree
without a SHIFT/KEEP classification in `rebaseTicks`, so the taxonomy can never
silently drift from the state struct (spec 030 extended this to
`Belief.Reinforced`, the decay-curve anchor; spec 041 extends it again to
`PlaceFact.Seen`/`PeerSighting.Seen` (SHIFT) and `PlaceFact.Detail` (KEEP),
[[mental-maps]]/[[guardian-miracles]]; spec 042 extends it once more to
`Memory.Seq` (KEEP — the emitting event's store seq, an identity rather than a
clock value) and `Agent.SitVecTick` (KEEP — when the situation text was
rendered, [[memory-retrieval]]); spec 043 adds `Agent.NeedsAnchorTick` (SHIFT
— the trajectory-window edge anchor, elapsed-anchor shape, 0 = unset) and
`IntentRecord.Tick`/`IntentRecord.OutcomeTick` (KEEP — when an intent and its
outcome landed, self-history like `Memory.Tick`, [[decision-context]]), with
`TestSnapPreservesRemainingDurations` also proving the anchor's LEVELS
(`NeedsAnchor`, need values not ticks) ride a snap untouched while its tick
shifts; spec 044 extends it again with three KEEP
entries — `RunEnd.Tick` (when the run ended: history, the world never ticks
again), `DeathRecord.Tick` (the `NormViolation.Tick` shape), and
`MorgueEpilogue.Tick` (the `ChronicleEntry.Tick` shape); spec 046 adds three
more KEEP entries — `CurriculumPass.Tick` (when the pass was recorded,
history like `Memory.Tick`), `EvidenceRef.Tick` (an audit pointer at a
recorded event's tick, never a deadline), and `EvidenceRef.Seq` (the evidence
event's store seq, an identity like `Memory.Seq` —
[[curriculum-ladder]]); spec 062 adds one SHIFT entry, `Agent.LastMindIntentDone`
(the reflex PREP gate's yield-window anchor, only-non-zero, the
`Belief.Reinforced`/`NeedsAnchorTick` shape — [[reflex-policy]])). Byte-identity replay suites
(`TestMiracleReplayByteIdentity`, `TestMiracleSnapReplayByteIdentity`,
`TestMiracleGrantReplayByteIdentity`) prove each miracle type replays to the
same state hash as live application.

**Memory-provenance and belief-decay suites** (spec 030, [[guardian]]):
`internal/sim/origin_test.go` proves the `DirectPerception` classifier's closed
vocabulary (action/witness/omen direct; report/gist/digest/absent secondhand),
that every situated constructor stamps `Origin` at emission, that the reducer
copies it verbatim, and that a pre-030 memory (no `origin` field) round-trips
byte-identically. `internal/sim/belief_evidence_test.go` proves belief
formation stamps `Belief.Reinforced` to the formation tick regardless of
evidence, that a later revision leaves it untouched (US2 only re-anchors on
direct evidence), and that a log of coerced beliefs replays byte-identically.
`internal/sim/belief_decay_test.go` pins `EffectiveConfidence`'s half-life
curve to the tick (a pure, computed-on-read function — nothing stored ever
mutates), including a fractional-half-life midpoint proving the curve is
continuous, not integer-stepped, and a legacy no-stamp belief grandfathered to
undecayed. `internal/sim/belief_reinforced_test.go` proves the
`agent.belief_reinforced` reducer arm re-anchors a held belief's clock and is a
total no-op against a vanished belief ID, and that a log containing the event
replays byte-identically. On the mind side, `internal/mind/provenance_test.go`
proves the consolidation user prompt instructs the model to cite evidence and
reserve "witnessed" for direct perception, and that deterministic
provenance enforcement coerces rather than rejects; `internal/mind/belief_read_sites_test.go`
proves the nightly consolidation held-beliefs block is the one documented
exception that renders EFFECTIVE (not stored) confidence and marks a faded
belief while still listing it by ID.

**Walls/axes/paths unit suites** (spec 032, [[tool-registry]]): `internal/sim/wall_test.go`
covers wall build/chip/repair/demolish across both materials (plank vs stone
HP and material cost); `internal/sim/axe_test.go` covers `craft_axe` and the
axe-assisted chop/quarry yield and durability countdown to breakage;
`internal/sim/path_speed_test.go` covers a path tile's travel-speed doubling
over a deterministic grass corridor. These sit alongside
`TestReplayByteIdentityWallsAxesPaths` and `TestPre032SnapshotLoadsUnchanged`
above as the feature's full proof. Spec 038 (loud build failure & occupancy
tolerance, TASK-91, [[executor]]) rewrote `wall_test.go`'s occupancy coverage
into a `driveWithOccupant` (per-tick scripted occupant placement) matrix:
`TestWallOccupancyGuard` proves a permanent squatter defers completion then
fails loudly (`agent.build_failed{reason: "site blocked too long"}`) exactly
`wallOccupancyGraceTicks` past the due tick, never a wall, never a spend;
`TestWallBuildToleratesPasserby` proves a mid-work crossing that clears before
the due tick no longer cancels the build; `TestWallBuildDefersThenCompletes`
proves a departure inside the grace window lets completion land on the first
clear tick and never on an occupied one; and `TestWallBuildSiteVanishedFailsLoud`
proves a vanished reserved-tile site fails loudly immediately
(`reason: "site no longer buildable"`) with a same-tick paired failure memory,
never a bare `intent_done`. `builderFailure`, a shared log-scanning helper, is
what the first two of these plus the site-loss test read to assert the count,
reason, tick, and paired-memory invariants. Since spec 041 made `repair`/`demolish`
knowledge-gated ([[reflex-policy]]), the fixture helper `grantStructureFacts`
seeds an agent's map with the walls/structures its resolver test needs to
already KNOW before the ground-truth assertions run.

**IPC miracle round trips** (`internal/ipc/ipc_test.go`, spec 016): the
operator "miracle" command exercised over the real wire on a pure-sim world
(no LLM/guardian) — a move lands, spends a charge, and is visible in the next
state fetch; `--force`/`gratis` lands a miracle against an empty bank and
leaves it untouched at zero, while a non-forced attempt against the same
empty bank is refused; a `give_item` resolves the villager by name and the
grant is visible in the next state fetch; unknown kinds/names are refused
cleanly with the connection surviving.

**Guardian behavioral suites** (`internal/guardian/guardian_test.go`,
`internal/guardian/guardian_gaps_test.go`, TASK-74): the package's own tests
now prove the economy mirror, turn serialization, and context-window
contracts, not just the TASK-64 instruction surface. `guardian_test.go`
(pre-existing) covers turn converse/degraded/fallback paths, influence
landing (charge decrement, atomicity, perception memories), zero-bank
refusal, the firewall sentinel, charter fallbacks, skill-file
eligibility/ordering, the fixed-frame non-negotiables under an adversarial
battery, and capability-manifest gating; spec 029 (TASK-27) extended it with the
agency surface — vision landing and its single-target rejection
(`TestVisionLands`/`TestVisionRejectsMultiTarget`), omen group targeting and
dead-target/day-deferral behavior (`TestOmenLandsOnNamedGroup`,
`TestOmenDeadTargetRefused`, `TestOmenDayDefersToNightfall`,
`TestOmenDayDeferralCapExempt`), the meta tools driving the `LoopControl` seam
including the start-with-speed set_speed-then-resume order and the
converse-never-touches-the-clock guarantee (`TestMetaToolsLandThroughLoopControl`,
`TestMetaToolStartSpeedFailureStopsBeforeResume`, `TestConverseTurnNeverTouchesTheClock`,
`TestMetaToolLoopError`), spec 041's `send_vision` place grant
(`TestVisionPlaceRevealLands`: the full triple lands one atomic
nudge+memory+`metatron.place_revealed`+companion-memory batch and the
target's map gains the revealed fact; `TestVisionPartialPlaceRefused`: a
partial `place_kind`/`place_x`/`place_y` triple is refused before anything
lands, charges untouched; `TestVisionFalsePlaceRefused`: the reducer dry-run
rejects a reveal of a place absent from ground truth, the WHOLE batch
including the vision itself, spending nothing — [[mental-maps]]), the
extended handler-firewall audit (`TestHandlerFirewallAudit`,
SC-007), the fixed `guardianInitiativeFrame` (`TestInitiativeFrameFixed`), the
clock-speed ladder drift guard (`TestClockSpeedsMirrorLadder`), and the
single-origin directive label — a console prompt carries exactly one
"The player says:" and a system turn's carries none
(`TestTurnDirectiveLabelSingleOrigin`); spec 025 (TASK-72)
extended it with
turn retry-visibility tests (a turn whose loop consumed its transport retry
emits the non-terminal `cog.outcome` retried marker; a clean turn emits none)
and turn token-budget plumbing tests (`guardian.New` stores and passes the
`max_tokens.metatron_turn` budget — the llm.json config key itself is frozen
serialized vocabulary, spec 052 ruling 2; the default reproduces 1024). The
tool-loop retry matrix itself lives in `internal/toolloop/retry_test.go`
([[tool-loop]]), with the mind-side twins in `internal/mind/mind_test.go`. `guardian_gaps_test.go` closes what
that suite left untested: `TestChargeMirrorAccrualAndCap` drives
`metatron.charge_regenerated`/`metatron.nudged` through `Observe` → `run()` →
`mirrorState` and proves the bank accrues and caps at `sim.GuardianChargeCap`
without a sim executor; `TestTurnBusyConcurrent` runs two real goroutines
against the `turnBusy` CAS (channel-gated, meaningful under `-race`) to prove
exactly one `Turn` proceeds at a time; `TestObserveNeverBlocks` proves the
notify path drops rather than wedges the caller; `TestAbsorbRefreshesMirrors`
proves an observed batch's effects (alive map, chronicle story tail capped at
8) are visible to the very next turn; and `TestTailOfFile`/
`TestSoulTailWindow`/`TestTranscriptTailTurns` pin the soul/transcript
tail-window truncation rules (`tailOfFile`, the 4000-byte `soulTail`, the
6-whole-turn `transcriptTail`). All new concurrency tests are channel-gated,
never sleep-as-the-only-gate (the TASK-69 flake lesson).

**Standing-order suites** (spec 029, TASK-27, [[guardian-orders]]): the lifecycle
is proven on both sides of the door. Reducer-side (`internal/sim/guardian_test.go`):
`TestGuardianOrderPlacedRejections` and `TestGuardianPlayerOrderCap` pin the
door validation (duplicate id, bad origin, empty event_types, TTL bounds, agent
range, over-long condition/action, and the 3-active player cap with system-origin
exemption); `TestGuardianOrderLifecycle` walks active→terminal transitions and the
cancel/expiry/trigger race (exactly one terminal lands); `TestGuardianOrderExpiryExecutor`
proves the executor emits `metatron.order_expired` as a pure function of state+tick;
`TestGuardianOrdersSnapshotUpgrade` proves a pre-029 snapshot loads with empty order
state; `TestGuardianOrdersReplayIdentically` proves from-genesis replay reconstructs
the order set identically; `TestGuardianOrderPrune` pins the retain-32 rule.
`TestSurvivalWatchReducer` (spec 059, FR-002) covers the origin-keyed
exemptions door-side: a survival watch lands despite an illegal (0-day) TTL
delta (the exemption, not a giant TTL), the cap still bites on player orders
with the three watches standing, a player `order_cancelled` naming a survival
watch is refused (`err` names "survival watch") and leaves it untouched, the
executor's expiry sweep emits zero `order_expired` for standing survival
watches even 30 days out, and an unknown `survival` kind or a player-origin
survival watch is refused at the door.
Guardian-side (`internal/guardian/orders_test.go`, 23 tests): the pure matcher and
agent probe (`TestOrderMatches`, `TestEventConcernsAgent`), id sequencing, placement/
cancel/expiry mirroring and prompt block, handler grant-gating, the end-to-end trigger
firing and its serialization with a console turn through the shared `turnBusy`
(`TestTriggerFiresEndToEnd`, `TestTriggerSerializesWithConsoleTurn`), the cancelled/raced
door resolution, the empty-bank precheck spending nothing and the budget-exhausted
one-moment-no-retry degradation (`TestEmptyBankPrecheckSpendsNothing`,
`TestTriggerBudgetExhaustedOneMomentNoRetry`), `TestReplayReconstructsWithoutFiring`
(replay rebuilds state but fires zero triggers, SC-002), the daytime-omen deferral
landing at nightfall and its cancel-before-night path
(`TestDeferredOmenTriggersAtNightfall`, `TestDeferredOmenCancelledNeverLands`), and the
fuzzy-confirm matrix — no-hit silence, rate-cap skipping, negative/failed verdict leaves
armed with no retry, and a yes verdict triggers (`TestFuzzyNoConfirmWithoutHit`,
`TestFuzzyRateCapSkipsExcess`, `TestFuzzyNegativeVerdictLeavesArmed`,
`TestFuzzyConfirmFailureNoRetry`, `TestFuzzyYesTriggers`). Channel-gated throughout.

**Run-outcome suites** (spec 044, TASK-31, [[morgue]]): the run-end/morgue/
escalation/grave surface is proven per layer. `internal/sim/run_end_test.go`:
`TestRunEndedOnceOrderedLast` (a same-tick multi-death batch declares the run
over exactly once, ordered after every death and its witness memories),
`TestEndedWorldEmitsNothing` (further ticks emit nothing — the `stepEvents`
top guard), `TestReplayRebuildsEnded` (from-genesis replay lands back in the
ended posture), `TestRunEndOmitemptyStable` (the three `State` additions are
`omitempty`, pre-044 snapshots byte-stable), and `TestEndedCommandGating`
(mutating commands refused, reads served, `inject_social` narrowed).
`internal/sim/morgue_test.go` covers the two injected arms
(`TestCharterObservedArm`, `TestMorgueEpilogueArm` — ring append in event
order, cap, agent −1 for the run-end epilogue) and
`TestEndedDoorAcceptsMorgueEpilogue` for the narrowed door's surviving type.
`internal/sim/grave_test.go`: grave placement at the death tile persisting
through replay, `buildSite` refusing a grave tile (research R10's deliberate
tension), the perception sweep granting the grave as a `PlaceFact`, place-tell
spreading it, and the witnessed-death grief rumor (SC-006).
`internal/sim/gru_test.go` gains `TestGruWoundInvariant` (the compile-pinned
`gruWound >= nearDeathBelow` escalation arithmetic) and
`TestGruEscalationScenario` (a weakened victim dies of cause `"gru"` with the
full death fallout; a healthy one keeps the survival floor — [[gru]]).
`internal/sim/toolcheck_test.go`'s `TestWhitelistDiffIdentical` — the
injection-whitelist tripwire — accepts exactly the two declared boundary
widenings (`metatron.charter_observed`, `morgue.epilogue`). On the mind side,
`internal/guardian/charter_observed_test.go` proves the first turn emits the
charter observation, an unchanged fingerprint stays silent, and an ended
world skips it (the shared fixtures pre-seed `charterFP` so turn tests keep
counting exactly the batches they drive); `internal/mind/epilogue_test.go`
proves absorbing a death or `run.ended` queues an epilogue, good prose lands
as ONE `morgue.epilogue`, and a narrator failure is a gap, never a stall
(FR-010).

**Decision-context suites** (spec 043, TASK-105, [[decision-context]]): the
context-grounding surface is proven per layer. Reducer-side,
`internal/sim/intentlog_test.go` pins the recent-intent ring —
`agent.intent_set` appends, done/failed stamp the newest open record, a
rejected intent appends already closed, an expired plan step stamps its open
record (or appends an unfired one), quick-succession overrides preserve
order, wraparound at the cap, and byte-stable replay — and
`internal/sim/needsanchor_test.go` the trajectory anchor's window-edge roll
(unset first window renders steady, refresh at the edge, a sleep spanning the
window). `internal/sim/journal_test.go` gains the `SelectJournalExcerpts`
matrix (term match, no-match ⇒ nil, rune cap, determinism) and
`internal/sim/memory_test.go` the annotated-selector twins:
`TestSelectedWindowMatchesLegacy` proves `StripSelected` of the annotated
window equals the legacy selector byte-for-byte, and the serendipity-tail
flag is pinned for the assembler's drop accounting. Mind-side,
`internal/mind/context_test.go` proves the block assembler: a golden-identity
prompt, determinism, per-block telemetry sizes, the full drop-priority ladder
under a shrunken budget, the protected memory floor and the byte-identical
memories/serendipity accounting split, plan-echo content/guard
phrasing/clearing, the journal block, a planted-memory relevance check, and
an aggregate budget-fit sweep. Daemon-side,
`internal/daemon/context_replay_test.go` is the replay-determinism harness
(T013/T024, SC-004): `TestContextReplayByteIdentical` runs a real unpaused
loop, then proves both recovery paths (snapshot recovery and genesis replay)
rebuild the state byte-identically and `mind.AssembleUserPrompt` reproduces
the assembled decision prompt byte-for-byte from the recovered world;
`TestSageThrashWindowContextReplay` (env-guarded via
`PROMPTWORLD_WORLD01_DB`) reconstructs a historical agent's context at an
exact tick from a COPY of a legacy world.db via the daemon package's
`replayToTick` ([[daemon-lifecycle]]), asserting the assembled text surfaces
the documented reflex thrash — inspection of assembled text only, no model in
the loop.

**Reflex-arbitration suites** (spec 062, TASK-103, [[reflex-policy]]): the
"instinct yields to intelligence" restructuring is proven per rung group.
`internal/sim/yield_state_test.go` (T001/T005) covers the PREP gate itself:
the yield-window anchor (`Agent.LastMindIntentDone`) is event-sourced,
snapshot-compatible (`omitempty`), and armed ONLY by a non-reflex intent
completion; the gate matrix proves the window suppresses prep and decays,
either danger band (food/warmth/rest) suppresses regardless of the window, a
reflex-sourced completion never arms the window, and a no-planner drive
matches pre-062 behavior except the enumerated danger-band suppressions
(SC-003). `internal/sim/day_warmth_test.go` (US2, T007) covers the new day
warmth rung: seek/relight KNOWN warmth (AS1), build with wood in hand (AS2),
a healthy-warmth day byte-identical to pre-062 (AS3), and
`TestDayWarmthDoesNotChopTheDeviation` — the flagged plan deviation — pinning
that the day rung stops before the night ladder's chop tail.
`internal/sim/night_search_test.go` (US3, T009) covers the bounded night
frontier-search fallback: a cold, nothing-known-or-carried-or-choppable agent
with a reachable frontier searches rather than sleeping cold
(`TestNightSearchFallbackSearchesWhenFrontierReachable`), and an unreachable
frontier still falls through to sleep, today's terminal floor.
`internal/sim/thrash_regression_test.go` (US4, T010, SC-001) is the
deliverable: the scripted world-01 Sage-shape scenario (cold, fed, a planner
`goto_warmth` completing at a fire, larder below stock target) encodes BOTH
proofs in one deterministic test — the OLD flip (absent the gate's inputs,
the reflex forages the agent away from the fire it was just sent to) and the
NEW hold-and-recover (with the gate's inputs live, zero prep intents fire
inside the yield window and the warmth trajectory recovers) — expressed by
nullifying the gate's INPUTS (an unarmed window, healthy warmth) rather than
mutating the immutable `prepYieldTicks`/danger-band constants, the
const-respecting encoding of the pre-062 world. `internal/sim/reflex_matrix_test.go`'s
cold-night truth table is updated for Gap A: the `wood=0`/nothing-known cells
that used to resolve `sleep` now resolve `search` (`TestColdNightReflexMatrix`),
with the no-frontier fall-through to `sleep` still pinned separately.

**Needs-conditioned recovery suites** (spec 064, TASK-104,
[[executor]]/[[reflex-policy]]): `internal/sim/needs_recovery_test.go` proves
the hold/complete/abort/yield lifecycle per user story. US1 (warm_up holds and
completes on warmth): `TestConditionlessIntentByteIdentical` (SC-004, FR-001)
pins that a condition-free `Intent`/`WorkStartedPayload`/`IntentSetPayload`
marshals without any of the new keys; `TestConditionRidesIntentSetPayload`
proves the condition rides the recorded `intent_set` and a malformed need is
dropped at the reducer door; `TestRecoverThenRelease` (SC-001) drives a
conditioned hold to a deterministic threshold-crossing completion tick, twice,
proving no premature `intent_done` and identical completion ticks across
runs; `TestAlreadySatisfiedCompletesImmediately` proves a condition already
met on arrival completes at once with no `work_started` (checked, not
assumed); `TestReplayDeterminismOverRecoverySpan` (FR-008) proves the whole
recover-then-release event log is byte-identical across two runs;
`TestWarmUpResolverAndClamp` pins the single `clampWarmUp`/`ClampWarmUp` clamp
home (default, in-range, below-floor, above-cap) and the `warm_up` resolver's
default/clamped threshold; `TestReflexWarmthRungsIssueConditioned` proves
BOTH the day and night reflex warmth rungs issue `goto_warmth` carrying the
doctrine-default condition; `TestReflexHoldDoesNotArmYieldWindow` and
`TestReflexHoldNoArriveIdleWander` (the enumerated no-LLM behavior change)
prove a reflex-issued hold never arms the spec-062 yield window and never
wanders mid-recovery. US3 (interruptibility, no new immunity):
`TestRecoveryAbortsWhenSourceDead` (SC-003) proves a dead-source hold aborts
with the distinct `agent.recovery_stalled` outcome within the stall-window
bound, stamping the ring `"stalled"`; `TestSurvivalPreemptsRecovery` (SC-003)
proves a higher-priority survival need in its danger band ends a hold at once
(via `preemptsRecovery`'s ladder-order doctrine — food outranks warmth,
never the reverse) rather than letting it start or continue holding;
`TestHoldingIntentOverriddenLikeAnyActive` proves a planner override replaces
a holding recovery exactly as it replaces any active intent (no new
immunity). US2 (the mechanism is generic): `TestSecondConsumerSharesMechanism`
(SC-004) drives a REST-conditioned intent through the identical
`executeAtTarget`→`recoveryHoldEvents` path — satisfied-on-arrival completion
and dead-source abort both — proving the condition machinery is need-generic,
not warm_up-private (sleep itself is deliberately NOT refactored onto it, per
plan R6 — the proof is at the mechanism level). US4 (wake to cold, audit Gap
C): `TestSleeperWakesToColdEmergency` (SC-005) proves a sleeper below
`exposureWakeBelow` with an actionable warmth ladder wakes;
`TestCozySleeperSleepsThrough` and `TestColdWakeIsEmergencyFloorNotDangerBand`
are the controls — a cozy sleeper and a merely-danger-band-cold one (above the
150 floor, below the 350 band) both stay asleep, pinning the wake as an
EMERGENCY floor, not the routine danger band; `TestSleeperNotWokenWhenNothingActionable`
is the churn-bound control. SC-002 end-to-end:
`TestSageWarmUpHeldToThresholdThenReleased` extends the 062 Sage-shape
scenario through a planner `warm_up`, proving the agent holds to the doctrine
threshold with ZERO mid-recovery prep dispatches or wandering, then releases
— the arrive-idle-vacuum dead end-to-end. `internal/sim/toolcheck_test.go`
gains a `"warm_up": 0` entry (arrival triggers the hold, not a timed work
cycle, like its `goto_warmth` sibling).

**Curriculum-ladder suites** (spec 046, TASK-68, [[curriculum-ladder]]): the
staged-worlds surface is proven per layer. Reducer-side,
`internal/sim/curriculum_test.go` pins the two `curriculum.*` arms — pass
recording with evidence, door validation (empty exercise id, unknown stage),
the pass ring's 32-cap prune, the once-per-(world,stage) unlock latch with
duplicate and stage-1 rejection — plus the pure gate logic:
`TestEvaluateUnlockGateConjuncts` walks all three gates (stage-1: any pass;
stage-2: only a `custom` charter-observed evidence entry — SC-004's negative
case pins that a default/preset-charter pass never opens it; stage-3: any
`custom` evidence entry), `TestCharterObservedEvidence` pins the sanctioned
constructor's `Custom = !payload.Default` derivation and its
wrong-type/bad-payload rejections, `TestEvaluateUnlockFixtureChain` drives
the full fixture pass→unlock chain, `TestExerciseDefinitionsParse` proves the
two shipped exercise definitions are well-formed, and
`TestCurriculumReplayDeterministic` proves a log carrying both types replays
byte-identically. Guardian-side, `internal/guardian/stage_test.go` (US2,
~500 lines) pins the gate-to-feature pathway: `TestStageCeilingRosterTable`
(per stage, the post-intersection roster equals the contract's ceiling
exactly — stage-1/-2 the four-tool watch set, stage-3/-4 and pre-ladder the
full roster), manifest intersection within the ceiling, the door refusing
beyond-stage acts, three-layer declaration/prose/door coherence, the stage-1
instruction lock's honesty (`TestStageOneInstructionLock` — the compiled-in
preset binds regardless of edits, the notice names the unlocking stage),
stage-2 charter-binds-skills-don't, `TestCrossStageDeterminism` (stage gating
never perturbs the sim, FR-006), preset resolution, an ungated pre-ladder
world byte-unchanged, and the tutor preset hot-reloading like any charter;
`charter_observed_test.go` gains the preset legs
(`TestCharterObservationTutorPresetIsDefault` — a stage-1 tutor-preset
world's observation honestly records `default: true`, so it can never
masquerade as player authorship — and
`TestCharterObservationEndedStageOneCoexists`). Daemon-side,
`internal/daemon/curriculum_test.go` drives the always-on unlock observer
with fixture events: upsert with the pass-event evidence pointer, non-
curriculum events ignored, a missing pass in the batch tolerated, and the
recorded path matching the world fixture. `internal/worlds/unlocks_test.go`
pins the record doctrine — missing/corrupt file loads empty (never an
error), atomic upsert-and-reload, same-stage overwrite, load-time healing
that drops malformed entries but KEEPS entries whose world path no longer
exists, and an unresolvable home warning-and-continuing. CLI-side,
`cmd/promptworld/stages_test.go` covers `new`'s stage resolution (stage-1
default for a new player, highest-earned otherwise, unearned refusal naming
skipped concepts unless `--override`, the override recorded honestly,
invalid stage/preset rejected, tutor-preset opt-out) and both `stages`
outputs; `internal/world/world_test.go` gains the manifest legs (stage
round-trip, absent-stage = ungated, `Open` rejecting a bad `stage` or
`charter_preset`), and `internal/skin/skin_test.go` pins the four
client-approved stage identities.

**Persona lifecycle suite** (`internal/persona/persona_test.go`, TASK-74): on
top of the pre-existing genesis-once/0444/missing-file-load coverage,
`TestPersonaMapsSweepAligned` proves the four index-aligned maps (`Texts`,
`Anchors`, `DriftMarkers`, `Secrets`) stay in lockstep with `sim.AgentNames` —
gaining or losing an entry in any one map fails the sweep;
`TestAnchorsMatchTemperamentLine` pins the documented "deliberately
identical" invariant between `Anchors` and each persona's `**Temperament:**`
line; `TestLoadUnreadableDegrades` proves an unreadable persona file degrades
`Load` to an empty string for that agent only, mirroring the missing-file
contract; `TestGenesisSeedsCharterAndJournal` proves fresh genesis seeds
`charter.md` (= `DefaultCharter`) and a rune-budgeted `journal.md` per agent,
and that an existing `charter.md` is never overwritten; and `TestSecretEvents`
proves the genesis `social.secret_seeded` events are index-aligned,
tick-0, tone `-70`, one per agent.

The whole suite runs under `-race`; it caught a real race (store `lastSeq`, loop
writer vs IPC readers — now atomic).

## Connections

Exercises [[sim-loop]], [[sim-state-reducer]], [[deterministic-rng]] (unit),
[[ipc-server]]/[[ipc-client]] (integration), and [[cli-promptworld]]/
[[daemon-lifecycle]] (e2e). [[guardian-miracles]] and [[guardian-orders]] cover the
reducer arms and doors these suites exercise; [[tool-registry]]'s spec-032 world verbs
(walls/axe/path) are what the new whole-feature and unit suites drive.
[[agent-mind]]/[[tool-loop]] are what the
loop-era replay suite drives through a real `Loop` + `loopMind`; the
provenance/belief-decay suites prove the substrate [[guardian]]'s omen/vision/miracle
memories now stamp. [[mental-maps]]'s own dedicated suite
(`internal/sim/mentalmap_test.go`) sits alongside the v3→v4 migration,
rebase-taxonomy, determinism, and vision-place-reveal coverage this note
tracks. Manual
validation results live in `specs/001-world-daemon/quickstart-results.md`.

## Operational notes

`go test -race ./...` runs everything in ~3 min (e2e dominates at ~187 s; measured
2026-07-23 during TASK-74 — the note's earlier ~25 s figure predates the e2e suite's
growth). E2E timing assertions
use deliberately loose bounds against CI jitter; tighten only with longer windows.
The executor behavior suites are seed-pinned: policy tuning that changes behavior
legitimately requires re-verifying (not deleting) the survival assertions.
