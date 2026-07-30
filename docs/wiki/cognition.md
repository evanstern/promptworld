---
name: cognition
description: The cognition horizon substrate — the decision-class registry (Fibonacci points, 1x staleness budgets scaled by clock speed at the delivery gates via EffectiveBudgetTicks, spec 067) and the pure Route/RoutePaused LLM-vs-reflex routing arithmetic and delivery-gate doctrine; estimation/calibration, horizon summaries/telemetry, and the adaptive-throttle governor split into child notes
kind: component
sources:
  - internal/cognition/doc.go
  - internal/cognition/registry.go
  - internal/cognition/route.go
verified_against: 04ff15001bd8a74f7c2965889c0d318fc0dc03a9
---

# Cognition horizon

`internal/cognition` (TASK-32, decision-4, specs/007-cognition-horizon) is the
deterministic substrate that scopes LLM authority by decision timescale versus
turn latency in game time: a model turn takes real seconds while the world
keeps ticking, and the drift scales with speed. Rather than capping speed to
protect cognition, the package decides — with no model in the loop — which
decisions may go to the model at the current speed, and what deterministic
floor covers when they may not. It is stdlib-only and imports nothing from
`internal/mind`, `internal/sim`, or `internal/llm`; all three depend on it,
never the reverse.

## How it works

**Registry** (`registry.go`): each model-reaching decision class is a
`DecisionClass{Class, Points, BudgetTicks, Degrade, FutureDated}`. `Points` is
the thought cost in Fibonacci points (the closed set 1/2/3/5/8/13 — ordinal,
host-independent, a property of the prompt shape); `BudgetTicks` is the
staleness budget in game ticks AT 1x — wall-clock patience (spec 067,
TASK-141). The scheduling gates (`Route`/`RoutePaused`, governor debt, every
horizon surface) hold it fixed against the fiction, while the delivery gates
(the [[sim-loop]] landing rung, [[social-fabric]]'s scene pre-abort) enforce
it scaled by the event-sourced clock speed through
`EffectiveBudgetTicks(ticksPerSecond)` — `BudgetTicks × ticksPerSecond`, the
base budget unscaled at uncapped speed — so a constant-wall-time thought is
judged the same at every capped speed instead of dying rejected-stale above
~4x on local tiers. Seven classes are
registered (`planner` 3pt/1200t degrading to reflex, `conversation`
13pt/7200t, `meeting` 2pt/3600t degrading to a template, `consolidation`
5pt/28800t, `chronicle` 5pt/86400t, `metatron` 5pt/86400t, and — spec 102,
[[guardian-agentization]] — `steward` 5pt/900t `DegradeSkip`, the guardian's
scheduled lane, budgeted BELOW planner so the caretaker sheds first; the
`"steward"` kind maps onto it); values are
doctrine — changing one is a reviewed code change, never runtime tuning.
`schedule.go` exports `NextPhasePreservingDue` — the TASK-44 cadence
advance, moved from `internal/mind` so both scheduled lanes share it.
The `musing` class retired with spec 017 — musing is now a roster tool inside
the planner's tool-use loop and shares the `planner` class's 3pt/1200t gate
rather than its own former 1pt/3600t budget ([[agent-mind]], [[tool-loop]]).
Spec 029 and spec 063 ([[grounded-feedback]]) add no new classes: the angel's
fuzzy-order confirm kind (`metatron_watch`, [[guardian-orders]]) and the
guardian's report-card kind (`report_card`, `DegradeSkip` — an unavailable
chain means the deterministic card parts stand alone) both map onto the
EXISTING `metatron` class (5pt/86400t) via one-line `kindToClass` entries on
the narrator/drama→`chronicle` precedent — same actor, event-triggered not
cadence-scheduled, so the spec-007 registry doctrine contract stays untouched.
`kindToClass` maps every LLM call kind (as a string, keeping the package leaf)
to a class; `ValidateKinds` enforces FR-002 at daemon start — an unmapped
kind, a non-Fibonacci point value, or a non-positive budget is a fatal
startup error. `Degrade` names the suppression floor: `skip` (recorded, not
silent), `reflex`, or `template` (a `faster-tier` variant existed but was
never wired past skip and was removed as dead code, TASK-71).

**Routing** (`route.go`): `Route(dc, ticksPerSecond, secondsPerPoint)` is pure
arithmetic — predicted wall seconds = points × seconds-per-point; predicted
drift ticks = wall seconds × ticks-per-second; allow iff drift ≤ budget. No
model, no randomness, no wall-clock reads (FR-007), so identical inputs always
yield the identical `Verdict`. The verdict carries the arithmetic verbatim as
a string (e.g. `3pt x 17.0s/pt x 32x = 1632 ticks > budget 1200`) so every
suppression is auditable in telemetry. `ticksPerSecond <= 0` (uncapped max
speed) always suppresses — prediction at unbounded speed is meaningless.
`RoutePaused(dc, secondsPerPoint)` (spec 040, decision-6) is `Route`'s paused
counterpart: a frozen world cannot drift, so predicted drift is zero — within
every budget — and the verdict allows every class whatever the SET speed,
including uncapped max; only the drift math is replaced (wall-ms is still
predicted, since wall time passes while frozen), and the arithmetic names the
paused state (`3pt x 20.0s/pt while paused = 0 ticks <= budget 1200`). The
paused fact is the caller's to supply from event-reduced state — the function
stays pure. `Route` itself is untouched, so running-world verdicts are
byte-identical to pre-040.

**Estimation and calibration** split into [[cognition-estimator-calibration]]:
per-provider live seconds-per-point estimation (EWMA, spike rejection,
breach-adoption, persisted estimator state surviving restarts) and the
human-authored calibration profile (`calibration.json`) that seeds it.

**Horizon summaries and telemetry** split into [[cognition-horizon-telemetry]]:
the operator-facing horizon-summary surfaces every status view reads, plus
the two per-call telemetry payloads `internal/sim/cognition.go` records.

**The adaptive-throttle governor** split into [[cognition-governor-debt]]:
the pure `Debt` staleness signal and the `Governor` hysteresis state machine
(spec 028) that turns the player's speed setting into a ceiling, shedding and
recovering one speed notch at a time.

## Connections

The [[agent-mind]] consults `Route` before every enqueue (`routeVerdict` in
`internal/mind/telemetry.go`) and records suppressions and thought outcomes as
`cog.thought`/`cog.outcome` events ([[event-types]]; since spec 043 a planner
`cog.thought` also carries the decision context's assembled sizes,
[[decision-context]]); while paused, `routeVerdict` short-circuits to
`RoutePaused` before the uncapped branch, so paused wins even at max (spec
040), and the suppression floors are [[reflex-policy]] and pre-authored
templates. See [[cognition-estimator-calibration]] for the live per-provider
estimate behind `secondsPerPoint`.

The [[sim-loop]] enforces the budget at landing: an intent whose measured
staleness exceeds its class's `EffectiveBudgetTicks` at the reducer's
event-sourced effective speed (spec 067 — pure, replay-safe) is rejected
(`OutcomeRejectedStale`) at the injection door, naming the scaled budget
(`staleness 9800 > budget 9600 (1200 at 1x × 8x)`). The split is doctrine:
scheduling gates pace cognition against the fiction, delivery gates forgive
queue-inflated latency against the wall. The residual gap is documented, not
closed (spec 067 FR-006): a thought whose WALL latency exceeds `BudgetTicks`
seconds still dies rejected-stale at every capped speed while the horizon —
which predicts from calibrated per-call latency, excluding endpoint queue
wait — keeps reporting the class healthy.

Every operator-facing horizon surface and the daemon's governor sampler read
[[cognition-horizon-telemetry]] and [[cognition-governor-debt]] respectively
rather than re-deriving `Route`'s verdicts; the router itself always reads
whatever EFFECTIVE speed the governor may have shed (spec 028 FR-010,
extending decision-4 from the other side).

## Operational notes

Telemetry: router verdicts land as `cog.outcome` events with the arithmetic
string as the reason (estimator/calibration and horizon/governor telemetry
are on the split notes above).
`OutcomeRetried` (`"retried"`, TASK-42) is the one NON-terminal outcome — a
scene reply failed to parse and continued via one retry; consumers summing
completions must filter it, and the optional `Raw`/`Retried` fields
(omitempty) carry the failed reply's bounded text and the retry flag on
terminals. `OutcomeClamped` (`"clamped"`, spec 058 US2/FR-003) is a TERMINAL
sibling of `OutcomeLanded` on the [[sim-loop]] landing decision: a `set_plan`
submission longer than `PlanStepCap` still lands, truncated to
`PlanStepCap` steps — distinguishing clamped from clean acceptance in the
`cog.outcome` trail, the `sim`-side analog of `toolloop.VerdictLandedClamped`,
not the same value.
Base budgets are never widened automatically — spec 067's speed scaling is
fixed arithmetic over the event-sourced clock rate; persistent suppression or
rejection on one class is a human retune signal.
