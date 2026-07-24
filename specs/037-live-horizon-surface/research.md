# Phase 0 Research: Live Cognition-Horizon Surface

No NEEDS CLARIFICATION markers survived the spec. This file records the design
decisions the plan builds on, each grounded in existing code/doctrine.

## R1 — Where the per-class verdict arithmetic lives

**Decision**: add `cognition.LiveHorizon(ticksPerSecond float64, secPerPtFor
func(class string) (secPerPt float64, ok bool)) []ClassStanding` to
`internal/cognition/horizon.go`, where `ClassStanding` carries the class name,
the suppressed flag, and the full `Verdict` (arithmetic string included).
Re-base `SuppressedAt` as a filter over `LiveHorizon` (suppressed names only).

**Rationale**: spec 035 FR-006 established the single-implementation doctrine:
every operator-facing horizon readout delegates to the one watched-class
iteration + `Route` call so no surface can disagree with the router
(`internal/cognition/horizon.go:57-83`). The live status surface needs richer
per-class data than `SuppressedAt`'s `[]string`, so the structured function
becomes the base and the string filter sits on top — still exactly one
iteration.

**Alternatives considered**: composing directly in `internal/ipc/server.go`
with raw `Route` calls — rejected: duplicates the watched-class loop and the
exclusion semantics already encoded in `SuppressedAt`, the exact drift FR-006
exists to prevent.

## R2 — Who owns suppression counters and how the mind reports

**Decision**: the `llm.Orchestrator` owns a process-lifetime, mutex-guarded
`map[string]int64` of per-class suppression counts —
`RecordSuppression(class string)` increments, `SuppressionCounts()
map[string]int64` returns a copy. The mind calls it from `emitSuppressed`
(`internal/mind/telemetry.go:77`) through an optional interface seam
(`suppressionCounting`), exactly like the existing `estimating` seam
(`telemetry.go:93-107`): a test fake or nil orchestrator is a silent no-op.

**Rationale**: `emitSuppressed` is the SINGLE terminal for every router
suppression (five call sites: planner `mind.go:326`, consolidation
`consolidate.go:100`, conversation `convo.go:115`, chronicle `narrate.go:269`,
meeting `meeting.go:59`) — one hook covers all classes, watched or not. The
orchestrator is already the daemon's live-telemetry hub (per-provider
estimators, `PendingCognition`, `StatusSnapshot`) and the ipc server already
holds it (`s.llm`), so counts reach status with zero new wiring. A no-LLM
world constructs no orchestrator → no counter machinery, satisfying FR-009 by
construction (the spec-028 governor precedent).

**Alternatives considered**:
- Counter struct in `internal/cognition` — rejected: the package is
  deliberately stateless/pure (no goroutines, no locks doctrine); a mutexed
  counter would be its first mutable singleton and needs new daemon→mind→server
  wiring besides.
- Daemon counts by watching its own event log for `cog.outcome`/suppressed —
  rejected: a subscriber loop and JSON re-parse to learn something the mind
  knows at emit time; also couples counts to log delivery.
- TUI-side counting from replayed events — rejected: resets on reconnect,
  invisible to the CLI, and each client would count differently (spec's
  "every client sees the same numbers").

**Increment cost**: one mutex bump on the absorb goroutine, matching the
"telemetry must never block the absorb loop" doctrine (`telemetry.go:74-76`
detaches the event emit; the counter bump is O(1) and lock-scoped).

## R3 — Wire shape on the status reply

**Decision**: `StatusData` gains `Horizon []HorizonClass
\`json:"horizon,omitempty"\`` (`internal/ipc/protocol.go:108`), composed in
`statusDataFull` (`internal/ipc/server.go:213`) only when `s.llm != nil`. Per
entry: `class`, `suppressed`, `verdict` (the router's arithmetic string
verbatim), `calibrated`, `suppressed_count`. Classes whose kind has no
admissible serving provider (`EstimateForKind` ok=false) are omitted from the
slice.

**Rationale**: additive-`omitempty` is the pinned pattern for exactly this
("pre-028 status bytes are byte-identical" — `protocol.go:140-148`, spec 034's
LLM badge fields, spec 035's `Warning`). Composing in `statusDataFull` reuses
the existing 1 s poll (FR-010) and the same fold point as the governor debt
and LLM snapshot. Structured facts (not a pre-composed sentence) let the
header badge, the dock block, and the CLI each phrase for their voice while
the arithmetic string stays the router's own (FR-002 audit trail).

**Alternatives considered**: folding into `llm.Status` — rejected: that struct
is the per-provider table the metatron pane renders verbatim; the horizon is a
cross-cutting composition (cognition × llm × clock) and belongs to the status
envelope, like `Warning`.

## R4 — Calibrated classes: included, remedy differs

**Decision**: the live surface includes every watched class with a serving
provider, calibrated or not (unlike the spec-035 set_speed warning, which
gates to uncalibrated providers). The `calibrated` flag drives phrasing only:
uncalibrated → "calibrate or slow down"; calibrated → "slow down".

**Rationale**: spec-035 research R2 gated the WARNING to uncalibrated worlds
because a calibrated world's drift has other signals (adoption events, the
governor) — but this surface answers a different question: "what is suppressed
right now?" A calibrated world at 32x with conversation suppressed is exactly
the "silent world" defect TASK-41 exists to fix. Telling an already-calibrated
player to recalibrate would be a false remedy (spec FR-007 / edge case).

## R5 — TUI placement

**Decision**: two surfaces off the same polled `Horizon` slice:
1. **Header badge** (`headerView`, `internal/tui/views.go:98`): when ≥1 entry
   is suppressed, append a warn-styled `[suppressed: planner, conversation]`
   badge following the `[llm: …]` badge pattern (`views.go:124-128`); absent
   otherwise. Renders in both widescreen and single-pane layouts (the header
   is shared).
2. **Metatron dock pane block** (`horizonLines`, beside `llmProviderLines`,
   `views.go:1482+`): one row per horizon entry — class, plain-language
   standing at the current speed, remedy when suppressed, and `skipped N`
   count. The pane already hosts the LLM provider table, spend, and health
   conditions: the established cognition-ops surface.

**Rationale**: the task note says "header/souls indicator … natural home: the
TASK-34 dock (new tab or status strip)". The header IS the always-visible
status strip; the metatron pane is where operators already look for
model-side state. A new dock tab for three rows was rejected (tab keys are
scarce; spec assumption pinned "existing pane"). The villagers pane was
rejected: it is per-agent, while the horizon is world-level.

**Phrasing**: plain-language via a small glossary-style helper, consistent
with the `verdictGlossary` doctrine (raw enum strings never reach the screen);
the verbatim arithmetic string is available to the pane row (it is already
operator-facing telemetry text elsewhere: chronicle digest reasons).

## R6 — CLI status parity

**Decision**: `cmd/promptworld/status.go` renders a horizon section for LLM
worlds (one line per entry: standing + count + remedy), nothing for no-LLM
worlds (US3).

**Rationale**: the daemon already knows; the payload already carries it; the
CLI render is a few lines and serves headless/classroom scripting. Output for
no-LLM worlds is unchanged since the field is absent (`omitempty`).

## R7 — Effective speed and uncapped max

**Decision**: compose the verdict at `sim.Status.Speed` (the loop's
post-governor EFFECTIVE speed, `internal/sim/loop.go:31`), converting via the
existing `clock.Speed.TicksPerSecond()`. No pause special-case: the verdict
answers "what may think at this speed."

**Rationale**: the router itself reads effective speed (spec 028 FR-010), so
the surface must too (spec FR-003). `Route` already returns
suppressed-with-dedicated-phrasing at `tps ≤ 0` ("at uncapped speed —
suppressed", `internal/cognition/route.go:27-30`), so max speed needs no
special handling anywhere above `Route`.
