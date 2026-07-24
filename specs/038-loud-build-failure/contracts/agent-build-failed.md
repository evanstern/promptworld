# Event Contract: `agent.build_failed`

Source row for the `docs/wiki/event-types.md` catalog (FR-002).

| Property | Value |
|----------|-------|
| Type | `agent.build_failed` |
| Payload | `BuildFailedPayload{agent int, goal string, reason string}` |
| Emitter | executor — mid-work re-validation of build goals (`build_fire`, `build_shelter`, `build_oven`, `build_chest`, `build_path`, `build_wall_plank`, `build_wall_stone`) |
| State effect | intent cleared (`Intent = nil`, `IdleSince` stamped) — same as `agent.intent_done` |
| Side effects | none (no material spend, no structure); paired same-tick `agent.memory_added` (OriginAction) for the builder |
| Reasons | `site no longer buildable` (any build goal); `site blocked too long` (walls only, after `wallOccupancyGraceTicks` past due) |
| Mind | re-arms the builder's planner (same list as `agent.intent_done`) |
| TUI digest | rendered as a failure line naming builder, goal, reason — never "finished" |
| Pass-through | included wherever `agent.intent_rejected` is whitelisted (see event-types.md sim-loop pass-through) |

## Contract guarantees

1. Emitted **instead of** — never in addition to — `agent.intent_done` for the
   failing build intent.
2. `reason` is stable vocabulary (the two strings above), suitable for tests
   and tooling to match on.
3. Exactly one `agent.build_failed` per failed build intent.
4. Deterministic: identical event bytes on record and replay.

## Catalog updates required

- New row for `agent.build_failed` (as above).
- Amend `agent.intent_done` row (event-types.md:115): "executor
  (done/invalid/unreachable)" → build-goal invalid cases now emit
  `agent.build_failed`; `intent_done` remains the resolution for successful
  non-build completion paths and non-build no-ops.
