---
name: daemon-lifecycle
description: Process lifecycle overview — daemon.Run's validate/recover/wire/tick/exit shape, signal-driven graceful shutdown, IsRunning, replayToTick; boot sequence detail routes to daemon-boot-recovery, daemon-orchestrator-startup, daemon-cognition-calibration
kind: pipeline
sources:
  - internal/daemon/daemon.go
verified_against: 6318cf8b53e407765f0c9793f5355a7af4777ed7
---

# Daemon lifecycle

`daemon.Run(dir)` is the foreground primitive that turns a save directory into a
living world: validate, recover, bind, tick, and — on any exit path — leave the
directory in a state the next start can resume from losslessly.

## How it works

The startup sequence runs nine steps in order — tool-registry gates through
snapshot+replay recovery and seeding (steps 0-5), then the LLM-orchestrator gate
with its governor/health/embedding wiring and teaching-posture/calibration
seeding (step 6), then IPC wire-up and `loop.Run` (steps 7-8). It splits into
three notes by substance (summary-style split — each keeps its own full detail
and links back here):

- [[daemon-boot-recovery]] — steps 0-5 and 7-8: tool-registry/bundle gates,
  pidfile/registry, store/meta validation, contiguity, snapshot+replay recovery
  and event seeding (meeting convention, tuning, survival watches, scenario
  arming), then IPC wire-up, `daemon.started`, and `loop.Run`. Also keeps this
  note's Operational notes (measured recovery timing, boot output).
- [[daemon-orchestrator-startup]] — step 6, first half: the always-on notify
  fan-out (scribe, curriculum observer, mind, guardian, scenario handoffs), then,
  gated on `llm.json`, the LLM orchestrator plus its governor sampler,
  provider-health hook, preflight goroutine, and embedding driver.
- [[daemon-cognition-calibration]] — step 6, second half: the teaching-posture
  boot default, tool-loop config warnings, mind/guardian wiring (bundles, stage,
  skin), the `ValidateKinds` gate, and calibration profile + persisted
  estimator-state seeding (including the 5-minute persister goroutine).

Shutdown: ctx cancellation (signal or `shutdown` cmd) returns from `Run` after the
loop's final snapshot; `daemon.stopped` is appended; deferred cleanup closes the
server (removing the socket), the store, and the pidfile — the pidfile only if it
is still ours (a slow shutdown can overlap a successor daemon that has already
claimed it; the CLI's stop wait is 30 s to match). SIGKILL skips all of this —
that is the crash path recovery is tested against.

`IsRunning(dir)` (used by CLI `start`/`stop`/`status`) reads the pidfile and probes
liveness without touching the world — deliberately, since TASK-147: it reads
`world.PidPathIn(dir)` (a pure path join, not a validating `Open`) rather than going
through a `*World`, so pidfile liveness stays checkable for a world this build can no
longer `world.Open` (e.g. an older `format_version`). Before this, `IsRunning` opened
the world first, so a running old-version daemon could never be detected — let alone
stopped — by a newer binary; `migrate` also refuses to touch a live daemon, so the two
gates combined could deadlock a world at an unsupported version.

`replayToTick(seed, m, st, cutoff)` (spec 043) sits beside `recoverState` as a
read-only reconstruction primitive the boot path never calls: it rebuilds
state as of an arbitrary tick by replaying the event log from genesis (a
snapshot may postdate the cutoff, so genesis replay is the only cutoff-correct
source), skipping — not stopping on — events past the cutoff, and tallying by
type, rather than aborting on, events the current reducer rejects, so a
legacy-format save whose manifest `world.Open` would refuse can still be
reconstructed from just its seed + map. Its consumer is the spec-043
replay-determinism harness (`internal/daemon/context_replay_test.go`,
[[decision-context]] / [[testing-strategy]]).

## Connections

[[cli-runtime-control]] runs this via `daemon` and detaches it via `start`;
[[sim-loop]] is the foreground engine; [[ipc-server]] the concurrent face;
[[event-types]] defines the `daemon.*` bookkeeping events it emits. See
[[daemon-boot-recovery]], [[daemon-orchestrator-startup]], and
[[daemon-cognition-calibration]] for the connections specific to each boot
phase.
