# Data Model: Per-Turn Context Grounding

**Feature**: specs/043-context-grounding | **Date**: 2026-07-24

All new state is reducer-derived from existing events (replay-safe by construction).
No store schema changes; no new event types — one payload extension.

## IntentRecord (new, internal/sim)

One entry in a villager's recent-intent ring.

| Field | Type | Meaning |
|---|---|---|
| Goal | string | intent goal name (tool-registry vocabulary) |
| Source | string | `"planner"` \| `"reflex"` \| `"plan"` — verbatim from `IntentSetPayload.Source` |
| Reason | string | the stated reason when present (planner/plan); empty for reflex — never invented |
| Tick | int64 | tick the intent landed (`agent.intent_set`) |
| Outcome | string | `""` (executing) \| `"done"` \| `"failed"` \| `"rejected"` \| `"expired"` — stamped by the closing event |
| OutcomeTick | int64 | tick the outcome landed; 0 while executing |

**Ring**: `Agent.IntentLog []IntentRecord`, newest last, capacity `intentLogCap = 8`
(append drops oldest). Rationale: 8 records cover several game-hours of intents — more
than the alternation window FR-003 needs — at fixed cost.

**State transitions** (reducer arms, internal/sim/state.go):

- `agent.intent_set` → append `{Goal, Source, Reason, Tick}`; the previous record, if
  still open, stays open (an override is visible as open-then-new, order preserved).
- `agent.intent_done` → stamp newest matching open record `Outcome="done"`.
- `agent.build_failed` → stamp `Outcome="failed"`.
- `agent.intent_rejected` → append-and-close `{Goal, Source:"planner", Outcome:"rejected"}`
  (a rejected intent never landed, so it has no open record).
- `agent.plan_expired` → stamp/append `Outcome="expired"` for the plan's pending step
  (plan end visible at next thought, FR-005).

**Validation**: Source restricted to the three known values; unknown sources recorded
verbatim but rendered as "unknown" (never guessed).

## Needs anchor (new fields on Agent, internal/sim)

| Field | Type | Meaning |
|---|---|---|
| NeedsAnchor | Needs | needs snapshot at the last window edge |
| NeedsAnchorTick | int64 | tick of that snapshot; 0 = unset (first window) |

Refreshed in the `agent.needs_changed` arm when
`tick - NeedsAnchorTick >= trajectoryWindowTicks` (const, default 1800). Direction per
need at render time: `current - anchor` with deadband `trajectoryDeadband = 10`
(of 1000) → rising / falling / steady; unset anchor → steady (first-thought empty
state, edge case 1).

## CogThoughtPayload extension (internal/sim/cognition.go)

| Field (new) | Type | Meaning |
|---|---|---|
| PromptBytes | int | total assembled user-prompt bytes |
| BlockBytes | map[string]int | rendered size per included block, keyed by contract block name |
| DroppedBlocks | []string | blocks dropped by budget, in drop order |

Reducer no-op unchanged; old events decode with zero values (backward-readable).

## Context block (assembler unit, internal/mind/context.go)

| Field | Type | Meaning |
|---|---|---|
| Name | string | contract name (see contracts/context-blocks.md) |
| Priority | int | drop priority; higher = dropped later; survival blocks = never |
| Render | func(state) string | pure renderer; "" = block absent (no empty headers) |

Assembly: fixed contract order; measure each block; if total approx-tokens
(`bytes/4`) > `contextBudgetTokens` (const 2000, tuning-manifest dial), drop whole
blocks lowest-priority-first until within budget; record sizes + drops into the
thought telemetry.

## Relationships

- IntentLog ← reduced from intent lifecycle events; rendered by the self-history block;
  also the future input of TASK-106's thrash detector (same taxonomy).
- NeedsAnchor ← reduced from `agent.needs_changed`; rendered by the trajectories block.
- `Agent.Plan` (existing) → rendered by the plan-echo block; no new state.
- `Agent.SitVec` + `SelectMemoriesRelevant` (existing, spec 042) → the memories block.
- `Agent.Journal` (existing) + deterministic term match → the journal block.
- Block sizes/drops → `cog.thought` → decision-trace view (existing surface).
