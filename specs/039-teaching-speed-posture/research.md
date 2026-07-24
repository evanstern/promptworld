# Phase 0 Research: Teaching-World Speed Posture

All Technical Context unknowns resolved. Each decision is grounded in repo artifacts
(decision-6, docs/design/horizon-vs-learner-iteration-speed.md, spec 035, live code).

## R1 — Where the posture rung comes from

**Decision**: export `cognition.MaxSafeSpeed(class string, secPerPt float64) float64`,
extracted from `HorizonSummary`'s existing per-class loop (internal/cognition/
horizon.go:39-44: highest `horizonLadder` rung where `Route(dc, sp, secPerPt).Allow`);
`HorizonSummary` re-implements on top of it.

**Rationale**: the number TASK-78 needs ("highest calibrated planner-safe rung") is
computed today only as `maxOK` inside a string formatter and never returned. Extracting
it (a) avoids a second arithmetic that could drift from the router (spec FR-004 —
`Route` remains the single rule, same guarantee spec 035 FR-006 relies on), (b) keeps
cognition a leaf package (pure float in/out, no clock import — mirrors the existing
`horizonLadder` mirror of `clock.CappedLadder()`).

**Alternatives considered**: computing the loop inline in `ipc/server.go` (duplicates
arithmetic; rejected); returning the whole per-class table (YAGNI — planner is the
posture class per decision-6; the warning already gets other classes via
`Route().Arithmetic`).

## R2 — Which provider's seconds-per-point

**Decision**: `Orchestrator.EstimateForKind(llm.Kind("planner"))` — the serving
(admissible chain-head) provider's live estimate; provenance via
`Orchestrator.CalibratedAt(name)` (`""` = bootstrap-seeded = provisional).

**Rationale**: spec 024 gave each provider its own estimator (llm.go:520-533);
`EstimateForKind` (llm.go:711-718) already answers "who actually serves planner work
and at what rate" — exactly TASK-78's "recompute from the profile the planner class
actually routes to". Recalibration flows through `SeedCalibration` re-seeding, so the
posture follows it with zero new plumbing (spec SC-005).

**Alternatives considered**: reading `calibration.json` directly at posture time
(bypasses live estimator drift and provider admissibility — wrong provider on
fallback; rejected); storing the computed rung in the manifest (decision-6 explicitly
forbids hard-coding; stale-vs-profile disagreement; rejected).

## R3 — How the teaching default is applied (and replay determinism)

**Decision**: at daemon boot, for a teaching world with an orchestrator, issue the
normal loop `set_speed` command with the posture rung after calibration seeding —
landing as a recorded `clock.speed_set` event, exactly like an operator set_speed
(internal/sim/loop.go:492-497). Initial state stays `clock.DefaultSpeed`
(internal/sim/state.go:114) — untouched.

**Rationale**: replay byte-identity is doctrine (spec 036 proved it; decision-6
consequences reaffirm it). A recorded event is the one mechanism that both changes the
default and replays identically. It also naturally yields spec US1/AC2: next boot
recomputes from the (new) profile and records the new rung. Operator overrides last
until restart — the posture is a *default*, not a leash (soft posture).

**Alternatives considered**: changing `state.go`'s initial Speed (unrecorded divergence
between manifest-aware boot and replay — breaks byte-identity; rejected); persisting
"operator overrode, don't re-default" state (new mutable state for marginal benefit;
restart-resets-to-posture is the honest ambient default per decision-6; rejected).

## R4 — Warning channel and composition with spec 035

**Decision**: reuse `StatusData.Warning` on the set_speed reply (ipc/protocol.go
additive field, spec 035 FR-002). New composer `postureWarning` sits beside
`uncalibratedWarning` (ipc/server.go:282-298); when both produce text they are joined
newline-separated in one Warning value. CLI keeps rendering via `setSpeedLine`
(cmd/promptworld/commands.go:857-863). The posture warning body lists, per suppressed
watched class, `Route(...).Arithmetic` verbatim (route.go:37-38 — "3pt x 17.0s/pt x
32x = 1632 ticks over budget 1200") plus the degrade consequence in plain language.

**Rationale**: spec 035 established the exact contract this feature needs — additive,
never blocks, `max`-gate untouched, warning-augmented success (contracts/warnings.md).
A second channel would fork the pattern for no gain. Unlike `uncalibratedWarning`
(gated to bootstrap-seeded providers), the posture warning fires on **calibrated**
teaching worlds too — that is its point; the two are complementary, not overlapping,
and an uncalibrated teaching world can legitimately get both (calibrate prompt +
posture override arithmetic).

**Alternatives considered**: blocking + `--force` confirm flow (violates decision-6
"soft cap, warn-with-override" and spec 035's nothing-blocks contract; rejected);
warning on the Metatron `adjust_speed` tool path (the angel's surface is the spec 037
live horizon; in-fiction phrasing of router arithmetic is TASK-41 territory; deferred).

## R5 — Posture visibility for TASK-68 (status surface)

**Decision**: additive `StatusData.Posture *PostureStatus` (`omitempty`), present only
for teaching worlds with an orchestrator: `{Rung string, Calibrated bool}`. The
teaching marker itself is read from `world.json`; the CLI status output gains one line.

**Rationale**: spec 037's `Horizon` field set the precedent (present-only-with-
orchestrator, omitempty, non-LLM replies byte-identical). Stage presets (TASK-68) need
"is teaching + effective rung + provenance" — exactly this triple; the manifest bool
alone is not enough because the rung is derived. FR-008 byte-identity for non-teaching
worlds falls out of omitempty.

**Alternatives considered**: folding into `WorldStatus` (would appear for all worlds —
breaks byte-identity; rejected); a new IPC verb (heavier than a status field; rejected).

## R6 — Manifest field and migration

**Decision**: `Teaching bool` with `json:"teaching,omitempty"` on `world.Manifest`
(internal/world/world.go:27-44). No FormatVersion bump (stays 3). Set at creation via
`promptworld new --teaching`; toggled offline via a new `promptworld teaching <world>
[on|off]` subcommand using a world-package read-modify-write helper.

**Rationale**: additive optional bool — old manifests unmarshal to `false` (spec
US4/AC3, FR-001, FR-008); `omitempty` keeps non-teaching `world.json` files
byte-identical on rewrite. FormatVersion bumps are for shape changes that old readers
would misread; a defaulting bool is not one (precedent: `Meeting *MeetingConfig` was
added additively). A CLI toggle honors "never hand-edit derived/config state" ergonomics
and gives TASK-68 presets a scriptable hook.

**Alternatives considered**: a string `posture` enum field ("teaching"/"") for future
postures (speculative; bool is the decided scope, rename cost later is one migration;
rejected for YAGNI); requiring world re-creation to toggle (hostile to classroom
operation; rejected).

## R7 — Model tier for implementation (constitution V)

**Decision**: Opus 4.8 via `spec-implementer` `model` param.

**Rationale**: rubric hits — touches `internal/cognition` (listed package), the boot
path writes a recorded event (replay-determinism-adjacent, doctrine decision-4/-6
boundary must stay exact: warn, never block), and the slice spans five packages.

**Alternatives considered**: Sonnet default (fails the rubric's cross-package +
doctrine-adjacent tests; rejected).
