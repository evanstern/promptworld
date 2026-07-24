# Implementation Plan: Teaching-World Speed Posture (Calibrated Soft Cap)

**Branch**: `task-78-teaching-speed-posture` | **Date**: 2026-07-24 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/039-teaching-speed-posture/spec.md`

## Summary

Add a per-world **teaching posture**: an optional `Teaching bool` on `world.Manifest`.
For teaching worlds, the daemon defaults the clock at each boot to the **posture rung**
— the highest capped-ladder speed at which the planner class still routes to the model,
computed live from the planner-serving provider's estimate (never stored). Raising speed
above the posture succeeds (soft cap) and the set_speed reply's existing additive
`Warning` channel (spec 035 pattern) carries the router's own per-class horizon
arithmetic. Uncalibrated teaching worlds are prompted to calibrate at boot and on speed
changes. Status gains an additive, teaching-only `Posture` block for TASK-68 stage
presets. No engine/routing change anywhere (decision-4 stands).

## Technical Context

**Language/Version**: Go 1.x (existing module)

**Primary Dependencies**: stdlib only (existing project posture); no new deps

**Storage**: `world.json` manifest (one new optional bool); `calibration.json` read-only

**Testing**: `go test ./...` — table tests beside code (project convention)

**Target Platform**: macOS/Linux daemon + CLI (existing)

**Project Type**: CLI + daemon (existing single Go module)

**Performance Goals**: posture computation is O(ladder × classes) pure arithmetic per
set_speed/status/boot — negligible; no hot-path work

**Constraints**: replay byte-identity (boot-time posture speed lands as a recorded
`clock.speed_set` event via the normal loop command path); non-teaching worlds
byte-identical replies (`omitempty` additive fields only); decision-4 doctrine — the
engine never refuses a speed

**Scale/Scope**: ~6 files touched across `internal/world`, `internal/cognition`,
`internal/ipc`, `internal/daemon`, `cmd/promptworld`; no schema migration
(additive manifest field, FormatVersion unchanged)

## Constitution Check

*GATE: evaluated against constitution v1.1.0 before Phase 0; re-checked after Phase 1.*

- **I. Artifact-Grounded Action** — PASS: shape derives from decision-6,
  docs/design/horizon-vs-learner-iteration-speed.md, and spec 035's warning contract;
  this plan + spec dir are the new artifacts; board task TASK-78 linked via
  spec-bridge before implementation.
- **II. One Task, One PR** — PASS: TASK-78 ↔ branch `task-78-teaching-speed-posture`
  in `.worktrees/task-78` ↔ one PR. Spec phases are internal breakdown.
- **III. Gates Over Assertions** — PASS: spec-bridge gate mirrors phases to TASK-78;
  status advances only with artifacts.
- **IV. Grounding Freshness** — PASS (planned): touched files are listed as sources by
  wiki notes (cognition-horizon, calibration, ipc-protocol, speed-ladder); PR is not
  done until `/grounding-wiki:wiki-update` re-pins and player-docs freshness passes.
- **V. Model-Tiered Workflow** — PASS: planning on Fable 5; implementation delegated to
  `spec-implementer`. Tier call: **Opus 4.8** — the slice is cross-package
  (world/cognition/ipc/daemon/cmd), touches `internal/cognition` (rubric-listed
  package), and the boot-time speed event is replay-determinism-adjacent
  (doctrine-adjacent behavior). Recorded on TASK-78 with this justification.

**Post-Phase-1 re-check**: PASS — design adds one optional manifest bool, one exported
pure function, two additive reply fields, one boot hook; no violations, Complexity
Tracking empty.

## Project Structure

### Documentation (this feature)

```text
specs/039-teaching-speed-posture/
├── spec.md
├── plan.md              # This file
├── research.md          # Phase 0
├── data-model.md        # Phase 1
├── quickstart.md        # Phase 1
├── contracts/
│   └── posture.md       # Phase 1 — wire/CLI/boot surfaces
└── tasks.md             # Phase 2 (/speckit-tasks)
```

### Source Code (repository root)

```text
internal/world/world.go          # Manifest: + Teaching bool `json:"teaching,omitempty"`
internal/cognition/horizon.go    # + MaxSafeSpeed(class, secPerPt) — exported maxOK loop
internal/ipc/protocol.go         # StatusData: + Posture *PostureStatus (omitempty);
                                 #   Warning doc-comment widened to posture case
internal/ipc/server.go           # + postureWarning composer; posture block in status;
                                 #   set_speed reply composition
internal/daemon/daemon.go        # boot: teaching default speed via loop command;
                                 #   teaching flavor of uncalibrated boot warning
cmd/promptworld/commands.go      # `new --teaching` flag; `teaching <world> [on|off]`
                                 #   toggle; status rendering of posture line
internal/*/..._test.go           # table tests beside each change
```

**Structure Decision**: existing single-module layout; every change is an additive
field, function, or composer inside the package that already owns that concern —
mirror of spec 035's placement so warning surfaces stay in one place (`ipc/server.go`
composers, `daemon.go` boot prints, `commands.go` rendering).

## Design Decisions (Phase 0/1 digest)

1. **Marker**: `Manifest.Teaching bool`, `json:"teaching,omitempty"` — absent field ⇒
   `false` ⇒ byte-identical non-teaching behavior; old worlds load unchanged; no
   FormatVersion bump (additive, defaulting).
2. **Posture rung**: new exported `cognition.MaxSafeSpeed(class string, secPerPt
   float64) float64` — generalizes `HorizonSummary`'s internal `maxOK` loop
   (horizon.go:39-44) over `horizonLadder`; returns 0 when even 1× suppresses.
   `HorizonSummary` refactors onto it (single source; FR-004 by construction: it calls
   `Route`). Posture = `MaxSafeSpeed("planner", est)` where `est` comes from
   `s.llm.EstimateForKind(llm.Kind("planner"))` — the serving provider, so spec 024
   divergence and recalibration are inherited for free.
3. **Boot default**: in daemon boot for `Manifest.Teaching` worlds with an
   orchestrator: after `SeedCalibration`, compute posture and issue the same loop
   `set_speed` command an operator would — it is recorded as `clock.speed_set`, so
   replay stays byte-identical and restart re-derives the posture (spec US1/AC2).
   Uncalibrated ⇒ still apply the (pessimistic) rung but print the teaching calibrate
   prompt (US3); pure-sim teaching world ⇒ no-op.
4. **Soft-cap warning**: new `(*Server).postureWarning(speed)` beside
   `uncalibratedWarning` (server.go:282-298): teaching world + orchestrator + requested
   rung > posture ⇒ compose from each suppressed watched class's
   `Route(...).Arithmetic` verbatim + plain-language consequence (degrade verb).
   Joined with the existing uncalibrated warning when both fire (newline-separated) on
   the one `StatusData.Warning` field — still never blocks, `max` gate untouched.
5. **Status surface**: additive `StatusData.Posture *PostureStatus{Rung string,
   Calibrated bool}` present ONLY for teaching worlds with an orchestrator (omitempty
   — FR-008 byte-identity for everyone else). `WorldStatus` untouched. CLI `status`
   renders one posture line; TASK-68 reads the same field.
6. **CLI**: `promptworld new <dir> --teaching`; `promptworld teaching <world> [on|off]`
   offline manifest toggle (world package helper does the read-modify-write); status
   line shows `teaching posture: 16x (calibrated)` / `(provisional — run calibrate)`.
7. **Out of scope**: Metatron's in-fiction `adjust_speed` path gains no warning (the
   angel's surface is spec 037's live horizon; noted for TASK-41), no TUI changes, no
   engine/routing change, no per-class budget knobs (decision-6 rejected).

## Complexity Tracking

> No Constitution Check violations — table intentionally empty.
