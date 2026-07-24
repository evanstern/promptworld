# Data Model: Loud Build Failure & Occupancy Tolerance

**Date**: 2026-07-24 · **Spec**: [spec.md](spec.md) · **Decisions**: [research.md](research.md)

## New event: `agent.build_failed`

| Field | Type | Meaning |
|-------|------|---------|
| `agent` | int | Index of the builder whose build failed |
| `goal` | string | The build goal that failed (e.g. `build_wall_stone`, `build_oven`) |
| `reason` | string | Human-readable cause (`site no longer buildable`, `site blocked too long`) |

- **Emitter**: executor, mid-work re-validation of build goals only.
- **State effect (reducer)**: identical to `agent.intent_done` — `a.Intent = nil`,
  `a.IdleSince = e.Tick`. No material spend, no structure.
- **Consumers**: mind `absorb` (re-arms planner, same as `intent_done`); TUI
  digest (renders as failure with reason); event catalog / tests.

## Extended constant

- `wallOccupancyGraceTicks = 120` — ticks past the due tick
  (`WorkStart + workDuration`) that a wall completion may defer on an occupied
  reserved tile before failing loudly. Pure constant; no persisted state.

## Accompanying memory (existing type, new emission site)

`agent.memory_added` via `situatedMemoryEvent` with:

| Field | Value |
|-------|-------|
| `origin` | `OriginAction` |
| `salience` | `salShelter` (6) — same tier as the wall-built success memory |
| `why` | the intent's driving `Reason` (planner reason, already on the Intent) |
| `text` | first-person, names the structure and the cause, states it did NOT complete |

## State transitions (wall build intent)

```
intent_set ──▶ (travel) ──▶ work_started ──▶ working
                                              │  per tick: buildSite(Res) valid?
                                              │    no ──▶ agent.build_failed + memory  [intent cleared]
                                              ▼
                                        due tick reached (WorkStart + 600)
                                              │  buildSite invalid ──▶ agent.build_failed + memory
                                              │  agentAt(Res) ──▶ defer (no event this tick)
                                              │      └─ past WorkStart + 600 + 120 ──▶ agent.build_failed + memory
                                              ▼  tile clear
                                        agent.built  [materials spent, wall + HP, intent cleared]
```

Non-wall builds (`fire/shelter/oven/chest/path`): same diagram minus the
occupancy branch — site invalid at any tick after landing → `agent.build_failed`
+ memory; otherwise unchanged completion.

## Invariants

1. Never entomb: `agent.built` for a wall is never emitted while an agent
   stands on the reserved tile (unchanged, now enforced by deferral instead of
   cancellation).
2. A build that passed landing resolves in exactly one of: `agent.built` or
   `agent.build_failed` — never a bare `agent.intent_done`.
3. Every `agent.build_failed` is accompanied (same tick) by an
   `agent.memory_added` failure memory for the same agent.
4. All new events flow through reducers; recorded worlds replay
   byte-identically.
