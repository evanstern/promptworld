# Measurement: ambient event volume, baseline vs fixed (T015 / SC-001 / FR-001)

**Run**: 2026-07-30, this branch (`task-176-ambient-event-coalescing`),
paired seed-1337 synthetic headless worlds, 29 game-days each
(2,505,600 ticks), driven at the sim layer with the exact live-loop
semantics (start-of-tick `AdvanceTo` → `stepEvents` → apply), counting every
emitted row. Reproducible in-tree:

```
PROMPTWORLD_MEASURE=1 go test ./internal/sim/ -run TestMeasureAmbientVolume -v -timeout 60m
```

Both worlds ran the FULL 29-day horizon with all eight villagers alive at
the end (last living tick = 2,505,600 in both), so no per-day normalization
distortion applies. Baseline = legacy emission (regime off, K-carrying
tuning event with `needs_checkpoint_minutes: 0`); fixed = coalescing regime
at the doctrine default K=10. Same seed, same map, same dials otherwise.

## Measured (not extrapolated)

| | baseline (legacy) | fixed (coalesced) | reduction |
|---|---|---|---|
| `agent.moved` | 358,383 (12,358/day) | 0 | — |
| `agent.path_started` | — | 73,730 (2,542/day) | 4.9× vs agent.moved |
| `agent.path_truncated` | — | 0 | (no walls built by reflex agents) |
| `agent.needs_changed` | 334,080 (11,520/day) | 36,292 (1,251/day) | 9.2× |
| `gru.moved` | 151,179 (5,213/day) | 0 (fully derived) | ∞ |
| **ambient families** | **843,642 (29,091/day, 72.1% of rows)** | **110,022 (3,794/day, 25.3% of rows)** | **7.7×** |
| all events | 1,169,572 | 435,717 | 2.68× |
| ~serialized bytes (payload+envelope est.) | 178.7 MB | 83.4 MB | 2.14× |

- **SC-001 gate: ambient rows/game-day drop 7.7× ≥ the required 4×.** The
  test asserts the ≥4× floor mechanically.
- **US2 (log reads as story): the three families fall from 72% of all rows
  to 25%** — the majority of a coalesced log is intent lifecycle, memories,
  and observations.
- Movement: 4.9× (headless reflex agents make many short walks — wander/
  search churn at ~2,500 intents/day; minded worlds take longer, rarer
  walks, so 4.9× is the conservative end of the 5–15× scoping estimate).
- Needs: 9.2× at K=10 (checkpoints + crossings ≈ 1,251/day vs 11,520/day —
  close to the theoretical 10× ceiling; crossings supply the remainder).
- Gru: 151k rows → 0 (motion fully derived, ruling 4).

## Projected to playtest-1 (extrapolated — stated arithmetic)

Playtest-1 (29 game-days, minded agents): 1,011,063 events / 230 MB, with
the ambient families at 78% of rows (agent.needs_changed 332,752 +
agent.moved 332,525 + gru.moved 122,382 = 787,659). Applying the MEASURED
per-family reductions to playtest-1's own mix:

- needs 332,752 / 9.2 ≈ 36,200
- movement 332,525 / 4.9 ≈ 67,900 (conservative; minded walks are longer,
  so the true divisor is likely higher)
- gru 122,382 → 0
- non-ambient 223,404 unchanged (spec 104 deliberately does not touch them)

Projected total ≈ **327,500 events ≈ 3.1× under the 1,011,063 baseline**;
projected size, scaling bytes by the same row mix (ambient rows are smaller
than average, so scaling by rows is conservative), ≈ **75 MB ≈ 3× under
230 MB**. What is measured vs extrapolated: every number in the first table
is measured on this branch; the playtest-1 projection applies measured
per-family ratios to playtest-1's recorded family counts — the minded-world
walk-length effect (fewer, longer walks ⇒ better than 4.9× movement
coalescing) is noted but NOT claimed.

## Deviations from tasks.md's letter

- Worlds were synthetic in-memory sim-layer runs (the T015 "synthetic/
  headless run is fine" allowance), not daemon worlds preserved under
  `~/.promptworld/measure/` — the dispatch for this task forbids writing
  there; the env-gated in-tree test IS the preserved, re-runnable artifact.
- Dials were the doctrine defaults rather than playtest-1's "harsh" set: the
  ambient families' cadences (movement beats, the per-minute heartbeat, gru
  beats) are dial-independent, so the family ratios transfer; only the
  story-event mix (deaths, fires) varies with dials.
