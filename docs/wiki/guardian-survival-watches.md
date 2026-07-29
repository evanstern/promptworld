---
name: guardian-survival-watches
description: The spec-059 boot-seeded survival watches (near-death, starvation, exposure) — origin-keyed cap/TTL/cancel exemptions, live-only hysteresis matching (matchSurvival), and the survival-turn frame permitting the guardian to act on its own initiative to save a life. Split from [[guardian-orders]]; load when tracing the three genesis watches' mechanics.
kind: component
sources:
  - internal/sim/guardian.go
  - internal/guardian/orders.go
  - internal/guardian/turn.go
  - internal/sim/executor.go
  - internal/daemon/daemon.go
verified_against: 6a5344a12cdc8858909ca7cf209d55025135e9d5
---

# Guardian survival watches

Every order in [[guardian-orders]] is player-authored and player-authority; three
**system-origin survival watches** — near-death, starvation, exposure — exist
in EVERY world without any player action: seeded once at boot
(`seedSurvivalWatches`, [[daemon-lifecycle]], called right after `seedTuning`
and before the loop starts) via `sim.SurvivalWatchDefs`, the single home both
the boot seeder and tests build from. Since spec 054, the seeder pre-assigns
each placement event's store seq (`st.LastSeq() + i + 1`, the `Loop.
stampSeqs` precedent) before applying — the reducer's `order_placed` arm
stamps `PlacedSeq` from that envelope, so a boot-live watch's `PlacedSeq`
agrees with what `AppendEvents` records rather than diverging at Seq 0; boot
is single-writer, so `AppendEvents` then re-assigns the identical values.
They are the guardian's nature, not a
player configuration:

- **Origin-keyed exemptions** (`internal/sim/guardian.go`'s `applyGuardian`):
  a non-empty `GuardianOrder.Survival` (`near_death`\|`starvation`\|
  `exposure`, `sim.IsSurvivalKind`) must carry `Origin: "system"` and is
  refused otherwise; it is exempt from the TTL bounds entirely (`ExpiresTick`
  is set to `PlacedTick` as an honest placeholder and never validated or
  consulted — non-expiring by nature, not a giant TTL) and is refused
  outright at the `order_cancelled` door regardless of status, in-fiction
  ("that watch is my own nature… I cannot set it aside", `cancelOrder`).
  System-origin was already cap-exempt (parent note) — the three watches never
  count against `GuardianPlayerOrderCap`. The executor's `order_expired`
  sweep (`stepEvents`) skips a survival watch outright, the same origin-keyed
  exemption mirrored on the reducer's `order_placed` arm.
- **Danger bands** (`internal/guardian/orders.go`'s `survivalBand`, FR-008):
  each watch REUSES an existing sim doctrine constant rather than a new
  dial — `near_death` the `nearDeathBelow`/`nearDeathResetAt` hysteresis
  band, `starvation`/`exposure` the exact `Food == 0`/`Warmth == 0` predicate
  that drains health and stamps those `agent.died` causes, with `hungryAt`/
  `coldNightBelow` as the RE-ARM (recovery) thresholds — promoted-dial-ready
  (named, one home) but deliberately NOT added to `tuning.json` (dials are
  earned by evidence, not speculatively added).
- **Live matching, hysteresis-latched** (`matchSurvival`, distinct from the
  structural `orderMatches`): evaluated against `agent.needs_changed`
  payloads in the live absorb batch. A per-watch, per-villager
  `survivalLatch` (absorb-owned, guarded by `stateMu`, deliberately **NOT
  event-sourced** — matching is live-only by construction, the same
  guarantee `orderMatches` already gives every structural order) debounces a
  villager staying in-band: it fires once on entry, then stays silent until
  the villager recovers past the re-arm threshold (clearing the latch) and
  later relapses. `pendingTrigger` still gates one in-flight survival turn
  per watch per batch, exactly like a structural order.
- **`runSurvivalTrigger`** (vs. `runTrigger`) differs from an ordinary
  trigger in three ways that follow from a survival watch's nature: (1) it
  lands NO `metatron.order_triggered` — the watch is non-expiring and
  non-consuming, so it never transitions out of `active`, ever; (2) there is
  NO empty-bank precheck — the turn ALWAYS runs (the edge case: "the turn
  still happens… it must not burn the match silently"), so at zero charges
  every acting tool refuses in-fiction and the transcript + a helpless
  moment are the honest record of a watch that woke to find itself
  powerless; (3) the turn runs with `turnOrigin{system: true, survival:
  true, seed: "The survival watch has woken you: <villager> is <peril> and
  may die. <order.Action>"}` — the seed names the endangered villager and
  peril so the turn can aim — and its soul/transcript record is marked
  distinctly, `[survival watch]` rather than the ordinary `[watch]`
  (`recordTurn`), attributing the authority trail to the survival duty
  (FR-007).
- **The survival-turn frame** (`internal/guardian/turn.go`): `turnOrigin.survival`
  is the ONLY thing that changes the turn's system prompt —
  `buildTurnSystemPrompt(survival, …)` swaps ONLY the initiative frame,
  `guardianSurvivalFrame` in place of the ordinary `guardianInitiativeFrame`
  (byte-identical otherwise; a non-survival call site still composes the
  pre-059 prompt verbatim, FR-005). The survival frame is the ONE carve-out:
  for THIS peril alone the guardian may send a vision or work a miracle on its
  own initiative — no player authorization needed, charge cost unchanged —
  while the world's clock and every OTHER standing order remain the
  player's alone to command, exactly as before (FR-004). See [[guardian]]'s
  Turns section for the frame-composition mechanics shared with every other
  console/system turn.
- **The miracle targeting digest** (`buildTargetingDigest`, spec 059 US3) is
  a SEPARATE, independently-gated prompt addition (any turn whose granted
  roster offers `work_miracle`, not just a survival turn) — see
  [[guardian-miracles]] for its mechanics.
- **Composition with the spec-083 neglect detector**
  ([[executor-needs-survival]]): the two are DISJOINT layers with a
  deliberate ordering — neglect fires UPSTREAM at the danger bands
  (350/350/250) plus two hours of proven inaction, giving the VILLAGER the
  first chance to save itself (the salience-9 percept + generation bump),
  while these watches fire at the later emergencies (health < 200,
  Food == 0, Warmth == 0) as the guardian's harder backstop — on Oak's
  trajectory the percept lands ≈5 game-hours before the exposure watch's
  moment, and no surface double-alarms (different events, different
  consumers, different times). `matchSurvival` is untouched by spec 083;
  what composes today with zero new code is only that
  `sim.neglect_detected` sits in the world log/chronicle any guardian turn's
  event window can read. A fourth system-origin watch KIND matching the
  event structurally (the `orderMatches` shape) is a named, deliberately
  unbuilt seam — a guardian-authority change owed its own deliverable
  (spec 083 §Composition).

## Connections

[[guardian-orders]] is the parent standing-orders subsystem this note splits
from — the entity model, cap/TTL/prune mechanics for player-authored orders,
and the player-facing tool surface live there. [[daemon-lifecycle]] seeds the
three watches at boot, right after `seedTuning`. [[guardian]] hosts the turn
assembly (`buildTurnSystemPrompt`) whose survival frame this note describes,
and the absorb loop `matchSurvival` runs in. [[guardian-miracles]] shares the
targeting-digest mechanics a survival turn granting `work_miracle` uses.
[[sim-state-reducer]] holds `applyGuardian`'s origin-keyed exemptions.
[[executor]] skips a survival watch in its `order_expired` sweep. Spec:
`specs/059-metatron-survival-autonomy/` — `spec.md`, `plan.md`.
