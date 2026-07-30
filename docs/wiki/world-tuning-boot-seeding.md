---
name: world-tuning-boot-seeding
description: Split from [[world-tuning]] — how tuning.json is seeded into the event log at daemon boot (absent/present-file behavior, clamp warnings), the spec-057 genesis pin that fixes a new world's doctrine at birth, and why replay derives tuning EXCLUSIVELY from logged sim.tuning_applied events, never the file itself
kind: component
sources:
  - internal/daemon/daemon.go
  - internal/sim/tuning.go
verified_against: 0af53ec6d211c71e298072c045c67ccbbd13b61d
---

# World tuning: boot seeding, genesis pin, and replay independence

Split from [[world-tuning]] (the spec-048 world tuning manifest): how the
manifest actually gets applied at boot, how a freshly-created world pins its
doctrine at birth, and why a running world's replay never depends on the
manifest file surviving.

## Boot seeding

**Boot seeding** (`daemon.go`'s `seedTuning`, spec 048): follows the
`seedMeetingConvention` shape — build event → `state.Apply` →
`st.AppendEvents` — so the seed lands in the log like genesis. Called right
after `seedMeetingConvention` and before the loop starts and before
`mind.New`, so no tick and no planner schedule ever runs ahead of the tuned
values ([[daemon-lifecycle]]). Behavior:

- **Absent file**: seeds nothing — state keeps whatever tuning it already
  carries (defaults for a fresh world, or values already in the log from a
  prior boot's event). This is the "absent tuning.json == current behavior"
  invariant (FR-001).
- **Present file**: `ParseTuning` clamps and warns (each warning printed as
  `daemon: <warning>`) or fails boot. The resolved effective set is compared
  against `state.EffectiveTuning()` — via `TuningState.Equal` since spec 098
  (the dream pointer block resolves by value; bare `==` would re-append on
  every restart) —; identical sets append nothing (a restart with an
  unchanged file never grows the log), and a differing set prints one
  `daemon: tuning.json applied: <the full dial set, dream block included>`
  line, builds the event via `sim.NewTuningEvent(state.Tick, *parsed)`
  (which always carries the RESOLVED spec-098 dream block, [[private-dreams]]),
  applies it, and appends it.

**Spec 104's coalescing dial** (`NeedsCheckpointMinutes`, clamp [1,60],
default 10): joins the genesis pin like every other dial — a post-057
world's effective value is fixed at birth. `resolveTuning` treats an ABSENT
field in a sparse manifest as the doctrine default (10), never 0, so a
sparse `tuning.json` written before spec 104 existed still resolves to the
coalescing regime once the daemon upgrades; a pre-057 or migrated world
with no genesis pin at all keeps compiled defaults, the same documented
determinism hazard every other un-pinned dial carries. Applying ANY
`tuning.json` on an old, still-legacy world (`NeedsCheckpointMinutes`
previously 0) flips it into the regime at that boot: the `sim.tuning_applied`
arm's legacy→coalescing branch ([[sim-state-apply-agents]]) stamps every
living agent's `NeedsSyncTick`/`NeedsEmitted` and `Gru.Done` to the seed
event's own tick at the same moment, so the derived engine's floor starts
exactly at the flip with the past already covered by recorded events.

## The genesis pin

**The genesis pin** (spec 057 / TASK-108): `promptworld new` seeds one
`sim.tuning_applied` event carrying the full current default set among the
genesis events (`GenesisTuningEvent` in `tuning.go`, appended right after
`world.created` in `cmd/promptworld/commands.go`), so a post-057 world's
effective doctrine is fixed in its own log at birth — later changes to any
`default*` constant never rewrite its replay. `promptworld migrate`
deliberately does NOT back-fill the pin: pre-057 and migrated worlds follow
compiled defaults, a documented determinism hazard
(`control-surface-and-calibration.md` §6, the TASK-75 class). The boot seed
compares a manifest against the pinned set exactly as before — no
`seedTuning` change was needed.

## Replay and file independence

**Replay and file independence** (FR-005/FR-007): replay derives tuning
values EXCLUSIVELY from `sim.tuning_applied` events in the log — defaults
until the first one (which for post-057 worlds is the genesis pin), never
`tuning.json` itself. A world that ran under tuned values replays identically
even if the file is later edited, deleted, or the whole directory copied
elsewhere; a pre-048 log simply contains no such event, so pre-048 worlds and
logs load and replay unchanged on compiled defaults. Editing `tuning.json`
while the daemon runs has no effect until the next boot (no hot exposure —
that is §6 step 3 and explicitly out of scope here).

## Connections

[[daemon-lifecycle]] calls `seedTuning` right after the meeting-convention
seed, before the loop and `mind.New`. Upstream design record:
`specs/048-tuning-manifest/` (spec.md, data-model.md, `contracts/tuning.md`),
and `docs/design/control-surface-and-calibration.md` §6.

Back to [[world-tuning]] for the promoted dials, `TuningState`, the
nil-safe accessors, and the manifest file's clamp table.
