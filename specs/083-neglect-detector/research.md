# Research: Neglect detector (spec 083 / TASK-133)

**Direction source**: `docs/design/thrash-detection-research.md` §1.3 (the finding), §3
(injection design sketch), §4 (the split-and-sequence recommendation);
`docs/design/evidence/task-106/` (analyze.py schema notes, raw_results.json aggregates,
summary.json). The detector design is substantially pre-decided there; this file records
how each pre-decision maps onto the current codebase and the handful of decisions the
research left open.

## R1 — Trigger arithmetic (encoding §1.3's definition)

**Decision**: per (agent, need ∈ {food, warmth, rest}), on the per-game-minute needs
heartbeat, fire exactly when ALL hold (pre-tick state):

```
alive && !asleep
&& needValue(a.Needs, need) < band(need)            // band: 350/350/250 (spec 062)
&& since := neglect.Since(need); since != 0
&& nextTick - since >= neglectWindowTicks            // T = 7200 (2 game-hours)
&& last := neglect.ClassIntent(need); last == 0 || nextTick - last >= neglectWindowTicks
&& !neglect.Fired(need)
```

The band-entry anchor makes "below critical for T" exact (continuous — the needs arm
clears it on any recovery to/above the band); the class-intent anchor makes "zero
intents in its class over the window" exact and sliding (an intent early in a long
episode defers firing until a full T has passed since it). Both clauses use the same T:
the window that proves the peril is the window that proves the inaction.

**Alternatives rejected**:
- *Scan the `IntentLog` ring at firing time* — capacity is 8 (`intentLogCap`); eight
  wander/chop records inside an hour would evict an older class intent and manufacture
  a false "zero intents" over the full window. The ring is a prompt surface, not an
  exact counter. Dedicated reducer-derived anchors (the `LastMindIntentDone` precedent)
  are exact at O(1) state.
- *Post-decay values (the heartbeat's `n`)* — the near-death latch uses them, but the
  anchors are reducer state that folds AFTER the batch; mixing pre-tick anchors with
  post-decay values makes the predicate un-evaluable outside the sweep (breaks the R8
  probe). Pre-tick everywhere (the `recoveryHoldEvents` purity precedent); the
  one-heartbeat lag is immaterial against T = 7200.

## R2 — Critical bands: reuse spec 062's danger bands

**Decision**: `dangerFoodBelow` (350) / `dangerWarmthBelow` (350) / `dangerRestBelow`
(250) — the constants the reflex survival rungs trigger AT, the recovery preemption
reads, and the map's needs-critical overlay reads via exported aliases
(`SurvivalStarvingRearm`, `SurvivalFreezingRearm`, `DangerRestBelow`). No new numbers.

**Why**: (a) one home — "critical" already means exactly this everywhere; (b) the FP
argument writes itself: below these bands the reflex survival rungs are live, so a
healthy mind stack produces class intents at exactly these levels and resets the clock —
a firing proves the whole stack failed for two hours; (c) the map-overlay subsumption
(spec FR-012) is by construction, honoring "never a new channel".

**Rejected**: the emergency floors (150, `exposureWakeBelow`/hunger-wake) — firing that
late wastes the runway that makes the percept actionable (Oak at warmth 150 had ~37 min
to warmth 0, then the health drain); the guardian's emergency watches already own that
altitude (spec §Composition).

## R3 — T = 7200 (two game-hours), promoted-dial-READY

Oak's trajectory at the night rate (`warmthLossCold` 4/min): band entry 350 → 0 in
87.5 min; at T = 120 min warmth is 0 and health ≈ 900 (33 min × `healthLoss` 3/min);
death needs health 0 — ≈ 5.5 h after warmth hits 0. So the percept lands ≈5 game-hours
before the death the same slide produces. T also dwarfs every scheduling cadence
(reflex grace 120 ticks + 20-tick stagger; planner cadence default 1800), so "no class
intent for T" is dozens of missed opportunities, never a beat artifact. Research §2's
W = 2h cell being the too-tight end FOR THRASH does not bind here — neglect has no
episode-merging or K-count to destabilize. Named const, one doctrine home, rationale
comment, NOT tuning.json (§6 doctrine; explicit card instruction).

## R4 — Class dictionary: next to the goal registry, goal-name-only

**Decision**: `internal/sim/policy.go`, beside the resolver registry (research §2's rot
warning verbatim). v1: food = {forage, hunt, cook}; warmth = {goto_warmth, warm_up,
build_fire, refuel_fire}; rest = {sleep}. Anti-rot test: every member resolves in the
goal registry. Notes against the research's dictionary: FOOD adds `cook` (research §2
names it as serving food; it exists as a goal), drops `eat` (never an intent goal —
analyze.py's class had it because the evidence script counted events, not goals; the
direct `agent.ate` path self-resolves via the needs reset). WARMTH adds `warm_up`
(spec 064, post-dates the research). `chop` deliberately excluded — Oak's fatal window
was full of reflex chops.

**Open ambiguity (accepted for v1, revisit on live evidence)**: kind-parameterized
goals (`pick_up`/`withdraw` of food) are not classed — goal-name granularity can't see
the kind, and classing every pick_up would mask real neglect while hauling wood. The
accepted false-fire (percept while hauling food home critical-hungry for 2 h) is a
truthful nudge, not a pathology.

## R5 — Emission site and event: executor sweep, no injection door

The detector runs in `stepEvents` on the `%60` heartbeat — the sim's ONLY emitter, pure
over (state, tick); the research's "reducer-side" means sim-side-deterministic as
opposed to a mind-side observer (which it explicitly rejected: non-replayable, absent
from the morgue/chronicle record). `sim.neglect_detected` is executor-emitted and so
needs NO `injectSocialWhitelist`/operator-door entry (`loop.go` doctrine: pure function
of state + tick, the `charge_regenerated` comment). The companion memory rides the same
batch immediately after the event via `situatedMemoryEvent` (the map-corrected /
buildFailed companion shape; memories accrete only via `agent.memory_added`).

## R6 — Salience 9: join the interrupt band (deliberate deviation)

`salNeglect = 9 = GenerationBumpSalience`, the salNearDeath/salExiled band. The wiki's
default rule keeps texture memories below 9 "so they never interrupt an in-flight
generation the way near-death or exile do" — and that is exactly why neglect joins the
interrupt class: the percept's entire job is to break a mis-scheduling mind out of its
loop while there is still runway. The generation bump + next-prompt window win
(salience 9, fresh recency) IS research §3's "planner beat", implemented with zero new
nudge machinery; §3's cooldown ("one injection per episode") is the fired latch, which
bounds the interrupt rate by construction. **Rejected**: salience 8 (spear-broke band,
no interrupt) — the memory would wait politely for the next planner cadence, up to
1800 ticks, while the agent freezes; polite is what killed Oak. Origin:
`OriginWitness` — perceiving one's own condition is direct perception either way under
`DirectPerception`, and the villager did not ACT (that's the point).

## R7 — State substrate: one `omitempty` pointer, reducer-only, SHIFT on rebase

One new `Agent` field, `Neglect *NeglectState` (data-model.md) — the Journal/Hail/Map
pointer precedent so pre-083 snapshots round-trip byte-identically. Three writer arms,
all existing event types plus the new one: `agent.needs_changed` (band anchors + latch
clear), `agent.intent_set` (class-intent stamps), `sim.neglect_detected` (fired latch).
Lazily allocated on first non-zero write (replay-identical: the arms are
deterministic). All tick fields are duration anchors, non-zero-only ⇒ SHIFT in
`rebaseTicks` (`miracles.go` taxonomy, the `NeedsAnchorTick`/`IntentRecord.Tick` row) —
the taxonomy test must cover them.

## R8 — Validation: the log is not in-repo; fixture is binding, probe is evidence

Verified: no `.db` exists in-repo; the world-01 log lives machine-local
(`~/.promptworld/worlds/world-01/`, ~106 MB — `world.v3.db` archived at migration
beside the migrated `world.db`). `docs/design/evidence/task-106/raw_results.json`
carries aggregates only (per-episode flip counts/deltas); Oak's day-7 window exists
in-repo ONLY as documented shape (summary.json false_positive_notes[3]; research §1.3:
warmth 636→0 over ~6 h, only reflex chop + planner wander, death day 7 04:04 ≈ tick
511,440). Therefore, honestly:

1. **Binding (CI)**: recorded-fixture tests DERIVED from the documented shape — scripted
   event history folded through `Apply` (the `governor_replay_test.go` scripted-timeline
   idiom), sweep run at heartbeats, firing/silence asserted; plus live-vs-replay hash
   identity with the detector's own events in the log.
2. **Evidence (opt-in)**: env-guarded probe on the `PROMPTWORLD_WORLD01_DB`
   copy-and-replay idiom (`TestSageThrashWindowContextReplay`,
   `internal/daemon/context_replay_test.go` + `replayToTick`, `daemon.go:731`):
   replay the real log to sampled ticks, evaluate the factored predicate over replayed
   state — Oak's window true, labeled healthy episodes and Ash/Hazel false. Skips
   without the env var; a run is recorded as task evidence (the spec-043
   `evidence/sc-004-replay.md` precedent).

The card's AC #1 wording ("validated against Oak's death window in the world-01 v3
log") is satisfied by 1 + 2 together, with the in-repo/not-in-repo split stated rather
than papered over.

## R9 — Alert surfaces: membership, not machinery (spec 077 precedent)

Chronicle: one case in `isAlertType` (grammar.go) + `digestRegistry` entry +
`catalogFixture` row + `docs/wiki/event-types.md` backtick + digest-grammar contract
§3 row — exactly the five touches `stranger.took` shipped as. No family change (`sim`
maps to `familySim`). Map: nothing — `needsCritical` (views.go) already paints
`styleAgentCritical` from the same band constants; a test pins the subsumption. No
village aggregation (research §2's aggregation addressed thrash's same-tick six-agent
storms; neglect is per-agent, episode-latched, rate-bounded by construction).
