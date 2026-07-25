# Implementation Plan: Tool Surface Hygiene

**Branch**: `task-110-tool-hygiene` | **Date**: 2026-07-25 | **Spec**: [spec.md](spec.md)

## Summary

Three contained changes: (1) the tool-loop Text validation
(`internal/toolloop/loop.go`, Text arm) clamps expressive fields rune-safely
with a notice instead of returning the rejection string — structural arms
untouched; (2) the set_plan landing guard (`internal/sim/landing.go`,
`PlanStepCap` check) truncates to the cap with notice instead of rejecting;
(3) `collect_water` + `bathe` leave `LoopRosterVillager`
(`internal/tool/roster.go`) and the glosses that advertise them
(`internal/tool/registry.go` glossQuarry/glossBuildOven), executor machinery
untouched.

## Technical Context

**Language**: Go 1.26.4 · **Deps**: none new · **Testing**: `go test ./...`;
toolloop + landing + roster tests · **Constraints**: replay determinism
(events carry post-validation text — already true); structural rejection
behavior frozen; conversation route frozen · **Scope**:
`internal/toolloop/loop.go`, `internal/sim/landing.go`,
`internal/tool/{roster,registry}.go`, `internal/mind` telemetry surface for
the clamp verdict, tests alongside.

## Constitution Check

- **I Artifact-grounded** — PASS (spec on main; decisions cite the task).
- **II One task, one PR** — PASS (TASK-110 → `.worktrees/task-110`).
- **III Gates** — PASS (bridge-driven status).
- **IV Grounding freshness** — PASS with follow-through: wiki notes sourcing
  toolloop/registry/roster re-pin post-merge.
- **V Model-tiered** — PASS: **Sonnet** — routine tier: single-mechanism
  changes with complete file:line diagnosis recorded on the task; no
  concurrency, no doctrine arbitration. Recorded on the task at dispatch.

**Post-Phase-1 re-check**: PASS.

## Phase 0 research decisions

- **R1 — where the clamp lives**: in the toolloop validation arm, keyed on an
  explicit is-expressive predicate. Prefer enumerating the four fields via a
  Param flag set at registry definition (e.g. `Clamp: true` on say.text, gist,
  muse.text, reasonParam) so the loop stays generic and the registry stays the
  single source of truth. The validation function must be able to REWRITE the
  arg (clamped value) — check its signature; today it only returns an error
  string, so it gains a mutate-or-return-replacement path. Verdict surface:
  the loop's existing per-call verdict/telemetry (see `internal/mind`
  handlers/telemetry and cog.tool_call events) gains a clamped annotation —
  reuse the existing verdict enum shape rather than inventing a channel.
- **R2 — set_plan clamp site**: the landing guard (`landing.go`) is
  reducer-side; clamping THERE keeps determinism (events carry the clamped
  plan). The tool result notice needs the clamp to be visible to the loop —
  check how landing rejections surface today (OutcomeRejectedGuard) and add a
  clamped-accepted outcome of the same shape.
- **R3 — rune-safe truncation**: reuse the existing byte-cap rune-safe idiom
  (`internal/mind/meeting.go` NormTextMax loop) — factor a tiny shared helper
  if both sites need it; never split UTF-8.
- **R4 — roster prune**: remove the two entries from `LoopRosterVillager`
  (roster.go) + gloss text; grep prompts/derive for remaining mentions. The
  registry Tool defs stay (executor + replay + tests reference them).

## Project Structure

```text
internal/toolloop/loop.go        Text arm: clamp path for Clamp-flagged params
internal/tool/tool.go            Param.Clamp flag (or equivalent)
internal/tool/registry.go        flag the four fields; gloss updates
internal/tool/roster.go          prune collect_water, bathe (+ revisit comment)
internal/sim/landing.go          set_plan step clamp outcome
internal/mind/…                  clamp verdict in telemetry/tool result
*_test.go alongside each         SC-001..005
specs/058-tool-surface-hygiene/  spec.md, plan.md, tasks.md, checklists/
```

## Complexity Tracking

None.
