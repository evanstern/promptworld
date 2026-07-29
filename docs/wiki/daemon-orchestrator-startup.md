---
name: daemon-orchestrator-startup
description: Daemon boot step 6a — notify fan-out (scribe, curriculum observer, mind, guardian, scenario handoffs), the LLM-orchestrator gate, governor sampler, provider-health hook, preflight goroutine, embedding driver
kind: pipeline
sources:
  - internal/daemon/daemon.go
  - internal/daemon/curriculum.go
verified_against: 72f82f41f7aa2e345572105894cd0fb7c02fc0aa
---

# Daemon boot: notify fan-out and orchestrator gate

Split from [[daemon-lifecycle]] (full sequence overview there, between steps 5 and
7 of [[daemon-boot-recovery]]): the first half of step 6 — wiring the always-on
notify-fan-out consumers, then, only when `llm.json` exists, building the LLM
orchestrator plus its governor/health/preflight/embedding companions. The
teaching-posture default, tool-loop warnings, and calibration/estimator seeding
that follow in the same step split to [[daemon-cognition-calibration]].

## How it works

Notify fan-out + companions: the loop's notify goes to the IPC broadcast, the
always-on soul scribe (which since spec 044 is constructed with the open
store as its event source — `scribe.New(dir, seed, map, snapshot, st)` —
because the [[morgue]] render is a pure fold over the FULL event history,
which the boot snapshot alone cannot provide; reads are rare, per death or
boot, and briefly serialize with the loop's appends on the store's single
connection), the curriculum-ladder unlock observer (spec 046 US3,
`curriculumObserver(w)` in `internal/daemon/curriculum.go` — always-on
like the scribe and wired BEFORE the LLM gate, so a no-model world still
records its unlocks: on observing `curriculum.stage_unlocked` it upserts
the per-user `~/.promptworld/unlocks.json` record with the world's
name/path and a pointer to the same batch's `curriculum.exercise_passed`
event as evidence; `worlds.UpsertUnlock` warns-and-continues on any
failure, so this advisory record can never perturb the loop — on a
pre-scenario world (no rubric emitter reachable, [[curriculum-ladder]])
it simply sits idle), and — when an orchestrator exists — the mind driver
([[agent-mind]]) and the Guardian component ([[guardian]], attached to the
server via `SetGuardian` for the console); all consumers are non-blocking by
contract. On a scenario world (spec 054, [[scenario-machinery]]), the
scribe (`scr.SetScenario(exercise)`) and, when an orchestrator exists, the
mind (`md.SetScenario(exercise)`) each receive the armed exercise id right
after construction and before the loop starts — the scribe's call also
re-renders the morgue immediately, so an already-ended scenario world's
run summary carries the exercise line from the very first boot render on
restart. The LLM
orchestrator ([[llm-orchestrator]]) starts only when `llm.json` exists
(`llm.LoadConfig` → `llm.New` → `srv.SetLLM`), closed on exit — config-gated,
fully outside the loop, so inference failures can never touch the simulation.
Inside that same conditional branch, spec 028's adaptive-throttle governor
sampler is built and started: `newGovernorSampler(orch, loop)` wired to the
server via `srv.SetGovernor` and run in its own goroutine
(`go sampler.run(ctx)`), sampling aggregate staleness debt every
`cognition.GovernorCadence` off the loop's non-blocking status door and
issuing shed/recover decisions through the loop's `Govern` door — a no-LLM
world builds zero governor machinery (FR-003, SC-004; see [[cognition]] for
the debt arithmetic and controller, [[sim-loop]] for the `govern` command).
In the same conditional branch, spec 034's provider-health surface is wired:
`orch.SetConditionHook` installs a closure that prints a `daemon: WARNING
llm provider …` (or recovered) log line and lands a durable
`daemon.llm_warning` event through `loop.InjectOperator` — the loop's
single-writer-preserving operator-event door ([[sim-loop]]); if the loop
isn't running (the shutdown window) the durable leg is dropped and the log
line is the sole record. `go orch.RunPreflight(ctx)` then starts the
boot-time + periodic model-existence probe in its own goroutine, fired and
forgotten under the shutdown ctx exactly like the governor sampler — boot
never blocks or fails on its results (see [[llm-provider-health]]).
Also in this branch, spec 042's embedding driver ([[memory-retrieval]]):
`mind.New` now also takes `w.Manifest.MemoryRelevance` (the world's
`memory_relevance` mode flag), and, ONLY when `orch.HasEmbedding()` reports
`llm.json` routes the `embedding` kind, boot builds
`mind.NewEmbedder(orch, loop, warnFn, w.Map(), w.Manifest.Seed,
state.Marshal())` — a peer of the mind driver that watches committed
`agent.memory_added` events — appends its `Observe` to the notify fan-out's
consumers, and prints one boot line naming the embedding model and
provider. An absent embedding route prints the off-switch line instead
(`daemon: embedding off (no "embedding" route in llm.json — memories stay
vectorless)`) and builds nothing — the same absence-is-the-feature-switch
doctrine as a world with no `llm.json` at all. `warnFn`'s shape mirrors the
provider-health hook just above (a daemon-log WARNING plus a durable
`daemon.llm_warning`, `kind: "embedding-unavailable"`, through the same
`loop.InjectOperator` door) but is debounced by the embedder driver itself,
not by [[llm-provider-health]]'s detectors — which never RAISE a condition
from embed traffic; since TASK-102 a successful embed does CLEAR a stale
preflight condition on the embedding provider ([[llm-provider-health]]).

## Connections

[[llm-provider-health]] is what the condition hook and preflight goroutine wired
here (spec 034) drive; its durable event rides [[sim-loop]]'s `InjectOperator`
door. [[memory-retrieval]] is the spec 042 embedding driver wired here only
when `orch.HasEmbedding()`; its failure warning shares [[sim-loop]]'s
`InjectOperator` door and the `daemon.llm_warning` event type with
[[llm-provider-health]] but is a separate, debounced-by-the-driver signal.
[[curriculum-ladder]] is what the always-on unlock observer wired here (spec 046)
serves. [[cognition]] supplies the debt arithmetic the governor sampler drives.
See [[daemon-boot-recovery]] for the surrounding boot steps and
[[daemon-cognition-calibration]] for the teaching-posture/calibration half of
this step.
