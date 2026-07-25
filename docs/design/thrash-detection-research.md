# Thrash detection as a percept — research findings & recommendation (TASK-106)

**Date**: 2026-07-25 · **Ground truth**: world-01 days 1–7 event log
(`world.v3.db`, 271,886 events, archived at migration) · **Reproducible
evidence**: `docs/design/evidence/task-106/` (`analyze.py`, stdlib-only,
read-only; `raw_results.json`; `summary.json`) · **Direction source**: spike
TASK-101; candidate definition from the task card.

## 1. What the data actually showed

Four findings that reshape the problem:

1. **Thrash is daytime and village-wide, not a three-villager quirk.** Nearly
   every large flip episode starts within minutes of dawn (06:00) and runs to
   dusk; days 4–5 are village-wide storms (six of eight villagers). TASK-101's
   Sage/Fern/Oak framing under-counted the blast radius. (Sage's "436 flips"
   reconciles as 415 strict / 456 subsequence forage↔goto_warmth.)
2. **Flip volume ≠ pathology.** Oak's 86-flip day 4 *gained* +723 food and
   +902 warmth — productive shuttling. The genuinely bad episodes are
   flat-or-negative-need ones: Sage day 5 (152 flips, food −56 over 15.6h),
   Sage day 6 20:50 (25.7 flips/h, zero gain, 134 tiles walked per need
   point), Fern day 6 (+12 food for 773 tiles), Cedar day 5 (44 flips,
   food −365). **Any detector without a need-progress clause is unusable**:
   827–2,736 raw firings per parameter cell are healthy interleaving; the
   clause removes 100% of them by construction.
3. **Oak's death was NOT thrash.** In the final ~6 hours before dying of
   exposure, Oak emitted only reflex `chop` and planner `wander` while warmth
   drained 636→0. No oscillation detector can catch **death-by-neglect** —
   it needs its own detector: *need below critical threshold for T with zero
   intents in that need's class*. This is arguably the higher-value percept
   (it is the shape that actually killed a villager).
4. **Alternative metrics**: all-class switch-rate is a weak discriminator
   (healthy Hazel out-scores dead Oak — seek/wander churn dominates).
   Wasted-travel ratio separates pathological windows by 20–100× (2.2–2.7
   tiles/need-point productive vs 59.5–134 pathological) but explodes when
   needs are capped — usable only as a secondary confirm inside a detection
   window, guarded on "at least one need below cap".

## 2. Detection definition (AC #1) — chosen, with FP/FN analysis

**Definition**: over a sliding window of **W = 4 game-hours** ending at each
food/warmth-class intent, count class transitions between FOOD =
{forage, hunt} and WARMTH = {goto_warmth, build_fire, refuel_fire}; fire when
A→B→A transitions ≥ **K = 8**, **AND** neither food nor warmth improved by
more than 5 points (0–1000 scale) across the window, **AND** at least one of
the two needs is below cap−ε at window end (the caps guard). Merge firing
points ≤ W apart into one detection episode; aggregate simultaneous per-agent
detections into a village-level event to avoid 6× alarm spam.

**Performance on world-01** (11 labeled-bad episodes, 8 agents, 6.2 days):

| Cell | episodes | bad caught | healthy-interleave FPs |
|---|---|---|---|
| W=4h K=8 + clause | **19** | **10/11** | 0 (clause removes all ~1,251 raw) |
| W=2h K=8 + clause | 17 | 9/11 | 0, but misses the slow Fern grind |
| W=8h (any K) | 30–46 | 11/11 | unstable — episode splitting, late-window rebounds re-admit healthy storms |

The single W=4h "miss" (Sage day 7 06:13) improved both needs ~650 points —
correctly reclassified as productive interleaving, not a false negative.
Zero detections on the two never-thrashing agents (Ash, Hazel).

**Known misses, accepted for v1**: three-way rotations (food↔warmth↔rest);
oscillation with period > W; class-dictionary drift (`cook` serves food; `eat`
never appears as an intent goal — classes must live next to the goal registry
or they rot). Death-by-neglect is out of scope for THIS detector by design
(see §4).

## 3. Percept/memory injection design sketch (AC #2)

- **Where the detector runs**: reducer-side, as a pure function of the intent
  history and need trajectory the state already carries (spec 043 gave agents
  self-history; the detector is the same substrate read deterministically).
  Emitting a `sim.thrash_detected` event keeps it replay-visible and
  device-independent — the mind replica absorbs it like any event. (A
  mind-side observer was considered and rejected: non-replayable, and the
  percept would vanish from the morgue/chronicle record.)
- **What it injects**: a high-salience observation memory in the agent's own
  voice-of-evidence: *"I have walked between the fire and the berry patch
  five times this afternoon; I am no warmer and no better fed."* Salience
  floor high enough to survive memory selection into the next planner prompt
  (the spec-042 retrieval machinery decides rendering; the memory just has to
  win on salience+recency, which a fresh high-salience observation does).
- **Optional planner beat**: on first detection of an episode, nudge the
  planner schedule (the paused-nudge / hail-style arming precedent) so the
  insight lands while it still matters — NOT a forced re-plan every firing
  (the episode merge + village aggregation bound the rate).
- **Cooldown**: one injection per agent per detection episode (the merged
  unit), not per firing point.
- **Dials**: W, K, the need-progress epsilon, and the salience floor are
  promoted-dial-ready constants (single home), NOT tuning.json entries until
  live evidence earns them (§6 doctrine).

## 4. Recommendation (AC #3)

**GO — but sequenced, and split in two:**

1. **Card the neglect detector as its own task, priority above the thrash
   percept.** *Critical need + zero intents in its class for T* is the shape
   that killed Oak, it is simpler than the oscillation detector, and it
   composes with TASK-111 (a survival watch the angel can act on) and
   TASK-108/103 (the reflex/arbitration layer it backstops).
2. **Implement the thrash percept (this definition) AFTER TASK-103/104 land
   and the TASK-122 re-measure runs.** The arbitration fix attacks the
   *cause* of the day-4/5 storms (reflex-vs-planner counter-scheduling); the
   percept is the agent-visible *symptom* layer. Measuring flip-rate post-103/
   104 first tells us how much thrash survives to be worth percepting, and
   world-01's episodes remain the regression corpus either way.

Proposed implementation shape when it goes: one spec covering the
`sim.thrash_detected` event + injection memory + episode/village aggregation,
with the world-01 labeled episodes as its acceptance corpus (detector must
catch 10/11 with zero healthy-interleave FPs, reproducible via
`evidence/task-106/analyze.py`).
