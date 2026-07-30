---
name: testing-strategy
description: How correctness is proven, routed by test layer/subsystem — unit determinism & loop replay, IPC/e2e integration, miracle/memory/build-verb suites, guardian & standing-orders, run-outcome & decision-context, reflex-arbitration & recovery, curriculum-ladder & takeover, grounded-feedback & persona. See linked children for per-suite detail.
kind: pattern
sources:
  - internal/sim/sim_test.go
  - internal/sim/guardian_test.go
  - internal/daemon/context_replay_test.go
  - internal/ipc/ipc_test.go
  - internal/mind/replay_test.go
  - internal/guardian/orders_test.go
  - internal/sim/scenario_test.go
  - internal/daemon/scenario_boot_test.go
  - internal/mind/scenario_narrate_test.go
  - internal/scribe/scenario_morgue_test.go
  - internal/tui/exercise_test.go
  - internal/world/scenario_test.go
  - cmd/promptworld/scenario_test.go
  - internal/tui/takeover_test.go
  - internal/tui/render_test.go
  - internal/tui/console_test.go
  - internal/tool/explain_test.go
  - internal/guardian/explain_test.go
  - internal/guardian/tutor_guide_test.go
  - internal/guardian/reportcard_test.go
  - internal/guardian/skin_battery_test.go
  - internal/sim/reportcard_test.go
  - internal/sim/rubric_hygiene_test.go
  - internal/tui/reportcard_test.go
  - internal/tui/help_guardian_test.go
verified_against: 9b4ed5aef5bfea50b67fac10f8e2153f065a814d
---

# Testing strategy

The spec's success criteria (determinism, crash-lossless resume, detach-isolation)
are only provable by tests, so the suite is layered: pure-logic harnesses at the
package level, an in-process integration layer, and binary-level e2e that execs the
real `promptworld`.

## How it works

Test suites are grounded here by layer/subsystem (corpus-spec v2 split,
2026-07-26): each entry below is a one-paragraph summary of a child note that
carries the full suite-level detail moved out of this note verbatim; every
child links back here.

**Unit determinism & replay harness** ([[testing-unit-harness]]): The package-level determinism harness (`internal/sim/sim_test.go`'s `driveTicks`, spec 041's mental-map diffing — since spec 104 also calling `AdvanceTo` at the start of each driven tick, the live loop's own convention) proven over the full [[executor]] behavior suites and the spec-012/013/032 world-migration fixture chains (v1→v4), plus the loop-era live-vs-replay proof (`internal/mind/replay_test.go`) through a real `Loop`+`loopMind`. Spec 104 adds its own equivalence/determinism battery (`internal/sim/advance_test.go`) and an env-gated multi-day measurement driver (`internal/sim/measure_test.go`).

**IPC integration & e2e harness** ([[testing-integration-e2e]]): The in-process IPC integration suite (`internal/ipc/ipc_test.go`: status round trip, subscribe-from-zero, idempotent commands, the governor/calibration/horizon status-fold coverage, large-reply handling) and the binary-level `e2e/` suite (hermetic `PROMPTWORLD_HOME`, the always-on/pause/crash-resume/detach quickstart scenarios).

**Miracle pricing & reducer suites** ([[testing-miracle-suites]]): Miracle-cost derivation (`sim.miracleCost` ≡ the tool registry's single authoritative price source) and the full miracle reducer/IPC round-trip coverage: move/remove/grant/time-snap arms, charge doctrine, and `TestRebaseTaxonomyComplete`'s tick-taxonomy completeness guard.

**Memory-provenance & build-verb suites** ([[testing-memory-and-build-suites]]): Memory-origin/belief-decay proofs (`origin_test.go`'s direct-vs-secondhand classification, `belief_decay_test.go`'s half-life curve, `belief_reinforced_test.go`'s re-anchoring) and the spec-032 walls/axes/paths build-verb unit suites.

**Guardian behavior & standing-order suites** ([[testing-guardian-suites]]): Guardian package behavioral coverage (turn serialization, the agency surface, the handler-firewall audit, channel-gated concurrency) plus the standing-order lifecycle proven on both the reducer door (`internal/sim/guardian_test.go`) and the guardian-side matcher/trigger machinery (`internal/guardian/orders_test.go`).

**Run-outcome & decision-context suites** ([[testing-run-outcome-context]]): The run-end/morgue/grave/escalation surface (`run_end_test.go`, `morgue_test.go`, `grave_test.go`, `gru_test.go`) and the decision-context assembler suites (the recent-intent ring, journal excerpts, the golden-identity prompt, and the replay-determinism harness in `internal/daemon/context_replay_test.go`).

**Reflex-arbitration & recovery suites** ([[testing-reflex-recovery]]): The spec-062 "instinct yields to intelligence" reflex-arbitration rungs (the PREP gate, day-warmth, night-search, and the Sage-shape thrash-regression proof) and the spec-064 needs-conditioned recovery lifecycle (hold/complete/abort/yield).

**Curriculum-ladder & takeover suites** ([[testing-curriculum-takeover]]): The spec-046 staged-worlds curriculum-ladder proof (unlock gate logic, stage-ceiling rosters, unlock records, CLI stage resolution) and the spec-056 takeover dispatch/precedence/dismiss matrix.

**Grounded-feedback & persona suites** ([[testing-feedback-persona]]): The spec-063 grounded-feedback mirror-drift pins and report-card producer/consumer suites, plus the persona lifecycle suite (index-aligned maps, genesis charter/journal seeding, secret events).

## Connections

Exercises [[sim-loop]], [[sim-state-reducer]], [[deterministic-rng]] (unit),
[[ipc-server]]/[[ipc-client]] (integration), and [[cli-promptworld]]/
[[daemon-lifecycle]] (e2e). [[guardian-miracles]] and [[guardian-orders]] cover the
reducer arms and doors these suites exercise; [[tool-registry]]'s spec-032 world verbs
(walls/axe/path) are what the new whole-feature and unit suites drive.
[[agent-mind]]/[[tool-loop]] are what the
loop-era replay suite drives through a real `Loop` + `loopMind`; the
provenance/belief-decay suites prove the substrate [[guardian]]'s omen/vision/miracle
memories now stamp. [[mental-maps]]'s own dedicated suite
(`internal/sim/mentalmap_test.go`) sits alongside the v3→v4 migration,
rebase-taxonomy, determinism, and vision-place-reveal coverage this note
tracks. [[scenario-machinery]]'s spec-054 suite spans
`internal/sim/scenario_test.go` (schedule compilation, incident/rubric
emission, `TestScenarioSchedulesCompile`), `internal/daemon/
scenario_boot_test.go` (boot arming end to end), `internal/mind/
scenario_narrate_test.go` and `internal/scribe/scenario_morgue_test.go` (the
chronicle/morgue surfaces), and `internal/tui/exercise_test.go` (the
exercise dock tab) — plus `internal/world/scenario_test.go` and
`cmd/promptworld/scenario_test.go` for the manifest validation and `new
--scenario` CLI paths. [[takeover-surfaces]]'s spec-056 suite
(`internal/tui/takeover_test.go`, `render_test.go`, `console_test.go`,
`help_test.go`'s ceremonies extension) is this note's other TUI-family
addition. [[grounded-feedback]]'s spec-063 suite spans
`internal/tool/explain_test.go` plus three cross-package mirror-drift pins
(`internal/sim`, `internal/toolloop`, `internal/tui`), `internal/guardian`'s
`explain_test.go`/`tutor_guide_test.go`/`reportcard_test.go`/
`skin_battery_test.go`, `internal/sim/reportcard_test.go`/
`rubric_hygiene_test.go`, and `internal/tui/reportcard_test.go`/
`help_guardian_test.go`. Since spec 078 (TASK-152), `help_guardian_test.go`
also carries the forward-ladder suite: a runtime-derived parity test
(`TestHelpLadderMatchesStagesJSONSubstrate` — expected rows computed from
the same substrate `stages --json` marshals, zero hardcoded stage
ids/counts/prose so an upstream catalog change flows through untouched),
byte-identity for fixed inputs, the nil-status/nil-replica/no-unlocks-file
floor, a replica-only mid-session unlock (earned, no audit pointer yet),
the override-without-laundering edge case, and 80x24 pager reachability.
Manual
validation results live in `specs/001-world-daemon/quickstart-results.md`.

## Operational notes

`go test -race ./...` runs everything in ~3 min (e2e dominates at ~187 s; measured
2026-07-23 during TASK-74 — the note's earlier ~25 s figure predates the e2e suite's
growth). E2E timing assertions
use deliberately loose bounds against CI jitter; tighten only with longer windows.
The executor behavior suites are seed-pinned: policy tuning that changes behavior
legitimately requires re-verifying (not deleting) the survival assertions.
