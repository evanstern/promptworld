---
name: deterministic-rng
description: Stateless randomness — every random decision is a PCG seeded from (world seed, purpose, tick, index), so replay needs no RNG state
kind: pattern
sources:
  - internal/sim/rng.go
verified_against: 0af53ec6d211c71e298072c045c67ccbbd13b61d
---

# Deterministic RNG

promptworld has no long-lived random stream. Every random decision constructs a fresh
`math/rand/v2` PCG seeded purely from its coordinates, making randomness a pure
function — the key trick that lets crash recovery and replay work without ever
persisting RNG state.

## How it works

`sim.rngAt(seed uint64, purpose string, tick int64, index int) *rand.Rand`:

- `purpose` (e.g. `"wander"`, `"genesis"`) is FNV-64a hashed and XORed into the world
  seed, giving each decision family an independent stream;
- the second PCG seed word mixes `tick` (via the splitmix64 constant
  `0x9e3779b97f4a7c15`) with the entity `index`.

Consequences:

- **Replay-free**: recovery rebuilds state from events; when the loop then re-lives
  quiet ticks, each tick's random decisions regenerate identically because they depend
  only on (seed, purpose, tick, index) — nothing consumed earlier matters.
- **Order-independent**: entities draw from independent generators, so refactoring
  iteration order can't shift anyone's rolls.
- **Seed-sensitive**: different world seeds diverge immediately (tested by
  `TestDifferentSeedsDiverge` in `internal/sim/sim_test.go`).

`sim.BundleRand(seed, purpose, tick, index) float64` (spec 036) is the one
exported face of the pattern: a thin wrapper returning `rngAt(...).Float64()`,
consumed by the bundle script runtime as `world.rand(purpose, index)` so
Starlark tools ([[bundle-tools]]) draw replay-identical randomness from their
coordinates alone. The `"bundle:<tool>:<purpose>"` namespacing lives in the
caller (`internal/bundle/worldview.go`); the wrapper stays a generic seeded
accessor.

## Spec 104 — a purpose string becomes replay-load-bearing

Ambient event coalescing (spec 104) is the first feature to make a
`purpose` string matter at REPLAY time, not just at live-emission time: on
a coalescing-regime world `gru.moved` is never recorded, so the gru's
stalk/prowl roll (`rngAt(seed, "gru-prowl", t, 0)`, [[gru]]) is re-derived
from scratch by the derived-progress engine (`internal/sim/advance.go`)
inside `Apply` itself, at whatever tick a replay or a live tick currently
folds — the SAME purpose string, tick, and index the retired per-beat
emitter used, so the re-roll reproduces the identical draw either way. That
makes `"gru-prowl"` (and the neighbor order it feeds) a replay hazard on
the spec-092 audit ([[sim-state-reducer]]'s reducer-constants doctrine): a
retune requires the spec-094 log-format-version machinery, never a bare
edit, because a coalesced log's replay depends on drawing it identically
forever, not just on the tick it was first drawn.

## Connections

The [[reflex-policy]] draws wander targets through this; [[sim-state-reducer]]'s genesis
agent placement uses purpose `"genesis"`. The pattern is what makes
[[sim-loop]]-level replay determinism (SC-006) cheap: the [[event-log]] plus the seed is
a complete description of REPLAYING a run.

## Determinism scope (spec 092/TASK-75) — per-log, not per-seed

The guarantee this pattern buys is **replay of a given log is exact**, not
"the same seed always produces the same run on any machine." Every random
decision here is a pure function of (seed, purpose, tick, index), so two
REPLAYS of the identical log always agree — but a live run's [[event-log]]
also carries `clock.degraded`'s `EffectiveRate` ([[sim-loop]]), a wall-clock
measurement of achieved throughput that becomes part of canonical state
(hashed). Two machines (or two live runs on the same machine under different
load) started from the same seed can measure different `EffectiveRate` values
and diverge in `Hash()` even though every RNG draw agreed — the divergence
enters through the clock, not through this file's randomness. "Same seed ⇒
byte-identical history" is only true within a single already-recorded log's
replay, never as a cross-machine or cross-run promise.

## Operational notes

Future systems (TASK-4 procgen, TASK-5 executor) should draw randomness the same way —
new purpose tags, never a shared stateful generator — or the replay contract breaks.
Research note R3 in `specs/001-world-daemon/research.md` records the deviation from a
single seeded stream and why.
