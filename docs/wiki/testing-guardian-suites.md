---
name: testing-guardian-suites
description: Guardian package behavioral coverage (turn serialization, agency surface, firewall audit, concurrency) plus the standing-order lifecycle proven on both the reducer door and the guardian-side matcher/trigger machinery. Split out of [[testing-strategy]].
kind: pattern
sources:
  - internal/sim/guardian_test.go
  - internal/guardian/guardian_test.go
  - internal/guardian/guardian_gaps_test.go
  - internal/guardian/orders_test.go
verified_against: 657c770f87404b936a0587db1f6b00e81b9f0ee6
---

# Guardian behavior & standing-order suites

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

## Connections

Part of the [[testing-strategy]] suite map (split out during the corpus-spec v2
restructure); see that note for the full layered test picture and links to
sibling suites.
