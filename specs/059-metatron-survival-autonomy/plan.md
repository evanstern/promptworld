# Implementation Plan: Metatron Survival Autonomy

**Branch**: `task-111-metatron-survival` | **Date**: 2026-07-25 | **Spec**: [spec.md](spec.md)

## Summary

Three moves on the spec-029 order machinery: (1) system-origin survival watches
(near-death, starvation, exposure) — origin-keyed exemptions (cap, TTL, cancel)
+ genesis/boot seeding via the established seed-if-absent pattern
(`seedMeetingConvention` / spec-057 genesis-pin family); (2) a survival-turn
frame: the initiative non-negotiable (`internal/metatron/turn.go` ~826,
"never on your own initiative") becomes turn-origin-conditional — survival
turns carry an authority carve-out for visions/miracles, everything else keeps
the restrictive frame verbatim; (3) a token-bounded targeting digest (villager
positions/conditions + passability) in miracle-capable prompts
(`internal/metatron/turn.go` prompt assembly + `internal/tool/derive.go`
miracle guidance), with a door round-trip regression test.

## Technical Context

**Language**: Go 1.26.4 · **Deps**: none new · **Testing**: `go test ./...`;
`internal/metatron` turn/orders suites · **Constraints**: replay determinism
(orders are event-sourced); no format_version bump; charter changes via the
charter-observed mechanism; machinery must survive TASK-112 agentization
(origin-keyed data + frame logic, no deeper coupling) · **Scope**:
`internal/metatron/{orders,turn,charter}.go`, `internal/sim/metatron.go`
(MetatronOrder origin semantics if needed), `internal/tool/derive.go`, the
genesis/boot seam, tests alongside.

## Constitution Check

- **I Artifact-grounded** — PASS. **II One task, one PR** — PASS
  (`.worktrees/task-111`). **III Gates** — PASS.
- **IV Grounding freshness** — PASS with follow-through: metatron wiki notes
  re-pin post-merge.
- **V Model-tiered** — PASS: **Opus 4.8** — doctrine-adjacent authority change
  in metatron turn logic (the initiative doctrine is the load-bearing safety
  frame); cross-file order/turn/charter semantics. Recorded at dispatch.

**Post-Phase-1 re-check**: PASS.

## Phase 0 research decisions

- **R1 — origin model**: MetatronOrder already carries an origin
  (`placeOrder(origin string, …)`); system watches use a distinct origin value;
  cap/TTL/cancel checks key on it. Verify the event payload persists origin
  (it should — orders are event-sourced); if not, extend the payload
  compatibly (omitempty).
- **R2 — seeding**: genesis for new worlds + boot seed-if-absent for existing
  (both patterns shipped: meeting convention, spec-057 tuning pin). Absence
  test = "no standing system-origin watches in state".
- **R3 — frame carve-out**: the non-negotiable string at turn.go:826-831 is
  compile-time; make the initiative block a function of turn origin
  (survival-watch match vs everything else). Keep the restrictive text
  byte-identical for non-survival turns (FR-005) — tests pin both frames.
- **R4 — watch matching**: spec-029 matching drives turns from world events;
  near-death/starvation/exposure conditions reuse existing danger-band
  constants (find them: needs thresholds in internal/sim — hungryAt-family /
  health floors). Debounce: rely on existing order-match/turn serialization;
  add nothing speculative.
- **R5 — digest**: assemble from replica state the turn already holds;
  hard token budget (follow the existing turn-prompt budget discipline);
  door round-trip test proves digest coordinates are acceptable.

## Project Structure

```text
internal/metatron/orders.go     origin exemptions (cap/TTL/cancel)
internal/metatron/turn.go       survival frame carve-out; digest assembly
internal/metatron/charter.go    charter wording via charter-observed mechanism
internal/sim/metatron.go        order origin persistence (if extension needed)
internal/tool/derive.go         miracle guidance digest hook
<genesis/boot seam>             seed system watches (new + existing worlds)
internal/metatron/*_test.go     SC-001..004 incl. both-frames pin + door round-trip
specs/059-metatron-survival-autonomy/  spec.md, plan.md, tasks.md, checklists/
```

## Complexity Tracking

None.
