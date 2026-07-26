# TASK-122 — Full-length SC-007 flip-rate re-measure (compound MVLS build)

**Date**: 2026-07-26 · **World**: `~/.promptworld/measure/task-122` (seed 4242,
path-form/off-registry, retained for reproducibility) · **Binary**: post-104
main (`5acb5b5` line — specs 048, 057, 058, 059, 061, 062, 064 all merged) ·
**Tier**: gemma4:12b-mlx on every route (the SC-007 Sample-B planner tier;
zero cloud), calibrated · **Speed**: 4x for the first 1.6 game-hours, then 8x
(operator decision; horizon verified green — planner/conversation/meeting all
thinking, none suppressed) · **Span**: ticks 120–449,726 = **5.204 game-days**
(exceeds the ≥4-day floor) · **Method**: `flip_count.py` (this directory) —
the SC-007/TASK-101 counting rule verbatim, with spec 064's `warm_up`
classified into the warmth class; script validated against the world-01
baseline (reproduces Sage ≈72–74/day).

## Result

| Agent | Flips | Fast (≤200t) | Flips/game-day | n classified |
|---|---|---|---|---|
| 0 | 19 | 3 | 3.65 | 148 |
| 1 | 34 | 6 | 6.53 | 152 |
| 2 (worst) | 35 | 6 | **6.73** | 266 |
| 3 | 21 | 1 | 4.04 | 148 |
| 4 | 16 | 1 | 3.07 | 122 |
| 5 | 11 | 0 | 2.11 | 142 |
| 6 | 19 | 2 | 3.65 | 120 |
| 7 | 15 | 1 | 2.88 | 140 |

**Worst agent: 6.73 flips/game-day — clears the ≤36 bar by 81% and sits 90.7%
below the world-01 baseline worst (≈72/day).** Fast flips (≤200 ticks — the
thrash signature) total 20 across all eight agents over 5.2 days; baseline
Sage alone had 334–352 in 6.06 days.

## Survival context (low flips are meaningless in a dead village)

- **Zero deaths** in 5.204 game-days (baseline world-01: 2 deaths in 6.23).
- Final needs: **all 8 villagers at health 1000 and warmth 1000**, food
  463–550 — nobody near a danger band at span end.
- The three system survival watches stood armed the whole run and **never
  triggered** — no villager ever crossed near-death/starvation/exposure.
- `warm_up` (spec 064) used 8 times; `recovery_stalled` fired 2 times (the
  abort path working in the wild).
- Conversation volume: 61 scenes ≈ 11.7/day village-wide (world-01 ran ≈176/
  day through the ungated hail path — the spec 061 damper's effect).

## Comparison ladder

| Measurement | Build | Tier/speed | Span | Worst flips/day |
|---|---|---|---|---|
| Baseline (TASK-101) | pre-043 | cogito world-01 mixed | 6.06 d | ≈72 (Sage) |
| Sample A (SC-007) | 043 only | cogito:3b @ 32x | 0.99 d | 5.06 |
| Sample B (SC-007) | 043 only | gemma4 @ 4x, player world | sub-1 d | cleared by 8% |
| **This run** | **043+057+058+059+061+062+064** | **gemma4 @ 8x, 5.2 d** | **5.204 d** | **6.73** |

The compound build holds Sample-A-class flip suppression at the Sample-B tier
over a full-length span — the estimate Sample B's 8%-margin sub-day sample
could not give.

## Regime notes (honest caveats)

1. **Mixed speed**: first 1.6 game-hours at 4x, remainder 8x. Negligible
   (0.07% of span).
2. **Planning was structurally degraded at 8x**: set_plan landed 27 vs 287
   rejected-stale (~85% dead) — the fixed 1200-tick staleness budget vs
   wall-clock thought latency. Carded as **TASK-141**. Bias direction is
   conservative for this measurement (independent single decisions give MORE
   flip opportunities, and the bar cleared anyway). Sample A (32x) shared
   this regime.
3. **Expressive clamping was heavy** (229 muse clamps) — the spec 058
   clamp-don't-reject path working as designed; `rejected_malformed` totaled 7
   (all structural).
4. Fresh world ≠ world-01's scar tissue: no gru-trauma lore, no day-4/5
   village-wide storm conditions reproduced. The world-01 replay corpus
   (TASK-106 evidence) remains the regression set for the thrash-percept
   work if it proceeds.

## Reproduce

```sh
python3 docs/design/evidence/task-122/flip_count.py ~/.promptworld/measure/task-122/world.db
```
