# Data Model: Run outcomes, the morgue file, death escalation, and graves

**Spec**: [spec.md](./spec.md) | **Plan**: [plan.md](./plan.md) | **Date**: 2026-07-25

All types follow the codebase's event-sourcing discipline: canonical-JSON payload structs
(never maps), reducer-owned state mutation, `omitempty` on new `State` fields so
pre-feature snapshots stay byte-identical (no `format_version` bump).

## State fields (internal/sim/state.go)

| Field | Type | Semantics |
|---|---|---|
| `State.Ended` | `bool, omitempty` | Latched by the `run.ended` reducer arm; never cleared by any event. Read by `Loop.Run` (idle branch), `handleCommand` (mutating-command refusal), status surfaces, TUI replica. |
| `State.Deaths` | `[]DeathRecord, omitempty` | Reducer-appended death ledger (the `agent.died` arm appends). Ratified US1 implementation deviation: `stepEvents` is pure over pre-tick state and cannot scan the log, and per-agent tick/cause stamps would lose within-tick ordering — the ledger is the minimal mechanism that lets `RunEndedPayload.Deaths` carry the full run. Application order == event order; ≤ agentCount entries. Caveat: pre-044 worlds upgraded mid-run under-report pre-upgrade deaths in the eventual `RunEnd.Deaths` (snapshots predating the ledger). |
| `State.RunEnd` | `*RunEnd, omitempty` | Summary facts recorded at run end (final tick, deaths with causes, final cause) — the machine-readable record FR-001/FR-005 require, carried in snapshots. |
| `State.CharterFingerprint` | `string, omitempty` | Most recent effective-charter content hash (short hex), maintained by the `metatron.charter_observed` arm. Lets the morgue and status name "the revision in force" without scanning. |

### RunEnd

```
RunEnd {
  Tick       int64        // tick of the final death / run-end declaration
  Deaths     []DeathRecord // every death of the run, in event order
  FinalCause string       // cause of the last death
}
DeathRecord { Agent int; Tick int64; Cause string }
```

`Deaths` is reconstructible from the log; it is kept on state (small, ≤ agentCount) so
status and morgue render without a scan on the hot path.

## Event types (new / changed)

| Type | Emitter | Payload | Reducer effect |
|---|---|---|---|
| `run.ended` | `stepEvents`, same batch as the final `agent.died`, after all same-tick deaths, guarded by `!State.Ended` | `RunEndedPayload{Tick, Deaths, FinalCause}` (outcome-shaped) | Sets `Ended=true`, `RunEnd` |
| `agent.died` (changed) | existing heartbeat **plus** `gruStep` on an escalated kill | existing `DiedPayload{Agent, Cause}`; cause vocabulary gains `"gru"` | unchanged (Dead, intent/hail clear, inventory spill) **plus** grave placement (below) |
| `metatron.charter_observed` | Metatron turn pipeline when the effective charter's content hash differs from `State.CharterFingerprint` | `CharterObservedPayload{Fingerprint string, Default bool}` | Sets `CharterFingerprint` |
| `morgue.epilogue` | mind narrator worker (LLM-gated), landed via `InjectSocial`; **whitelist entry required** (`injectSocialWhitelist`, loop.go:193) | `MorgueEpiloguePayload{Agent int (or -1 for run-end), Text string}` | Appends to a bounded `State.MorgueEpilogues` ring (chronicle-ring pattern) so the scribe replica can render it |

Ordering guarantees: within the final batch — all `agent.died` (+ witness memories,
grave placements via reducer) strictly before `run.ended`. `morgue.epilogue` arrives in a
later batch (asynchronous narrator) or never; the factual render is complete without it.

## Structures / places

| Entity | Representation | Notes |
|---|---|---|
| **Grave** | `Structure{Kind: "grave", X, Y}` created inside the `agent.died` reducer arm (spill-idiom) | Closed-vocabulary updates: `Structure.Kind` comment (agents.go:237), `PlaceFact.Kind` comment (mentalmap.go:49-51), `placeFactKinds` (tool/registry.go:430), prompt landmark set (prompt.go:204), TUI glyph switch + legend (views.go). Durable freshness horizon (default — no `factHorizon` case). Occupies the tile for `buildSite` (deliberate, see research R10). |

## Morgue document (rendered artifact, never a source of truth)

`morgue.md` at the world dir root (`world.MorguePath()`), whole-file re-rendered by
`internal/scribe` on any batch containing `agent.died` / `run.ended` /
`morgue.epilogue`, and at every boot. Structure contract:
[contracts/morgue-document.md](./contracts/morgue-document.md).

### Epitaph (per death, render-time projection — not a stored struct)

**Ratified US2 deviation (2026-07-25)**: epitaph facts are captured by a **genesis
replay fold** — `renderMorgue` replays the full event log into a fresh reducer state
and snapshots each epitaph *at its death event* — not from the live scribe replica.
Live state violates the append-shape invariant and SC-004: open debts can later break
and orders expire, so a boot re-render from now-state would rewrite past epitaphs.
Cost: one full-log replay per render (per-death/boot — rare; blessed by plan
Performance Goals). "Source" rows below are read from the fold state at the death
event. Deterministic caps (ratified, constants in `internal/scribe/morgue.go`):
12 memories / 20 deeds per epitaph, 60 run-summary notable events — each with an
explicit "(N earlier … not shown)" line; lifetime-memory scan threshold = salience ≥ 7.

| Field | Source | Determinism |
|---|---|---|
| name, death day, days survived | replica agent + `clock.GameTime(tick)` | pure |
| cause | `DiedPayload.Cause` | pure |
| notable memories | retained `Agent.Memories` at death (salience-ranked) ∪ event scan `agent.memory_added` ≥ salience threshold | pure scan |
| relationships | `State.Relations` filtered on the villager | pure |
| debts owed / owing | `State.Debts`, `Status=="open"`, Debtor/Creditor == villager | pure |
| notable deeds | typed event scan over the chronicle notable-event vocabulary, filtered to the villager | pure scan |
| charter evidence | most recent `metatron.charter_observed` ≤ death tick (fingerprint + default/custom) | pure scan / state |
| active orders | `State.MetatronOrders` with `Status=="active"` at death (condition, action, watch subjects) | pure |
| epilogue (optional) | recorded `morgue.epilogue` for this villager | recorded input |

### Run-end summary

Run length (days), population-over-time curve (derived from death ticks vs genesis
count), every death with cause, notable events of the run (same curated vocabulary),
optional run-level epilogue.

## Status / IPC additions (additive, omitempty)

| Surface | Addition |
|---|---|
| `ipc.ClockStatus` | `Ended bool, omitempty` (+ optional run-end day) — governor-trio precedent |
| `StateData` | free — `State.Ended`/`RunEnd` ride the canonical state JSON |
| `promptworld status` (human + `--json`, live + offline snapshot) | ended posture line / field |
| TUI header | `ENDED` state token replacing running/`PAUSED` |

Contract: [contracts/status.md](./contracts/status.md).

## Validation rules

- `run.ended` fires exactly once per world, ever (guard: `State.Ended`).
- Escalation predicate: pre-attack `Needs.Health < nearDeathBelow` (200) — the raw
  threshold, **not** the hysteresis latch `Agent.NearDeath` (see research R4).
- `gruWound (250) ≥ nearDeathBelow (200)` ⇒ an escalated hit always lands at 0; assert in
  tests so a future constant change can't create a "floorless survivor" ambiguity.
- Morgue factual bytes: replay of the same history ⇒ byte-identical render (SC-004);
  epilogues excluded from byte-identity, included in render.
- Ended worlds refuse `pause/resume/set_speed/govern/inject_*` at `handleCommand`;
  `status`/`state`/`subscribe` continue to serve.

## State transitions

```
world running ──(last agent.died … run.ended in same batch)──▶ Ended (terminal)
    Ended: loop idles (no timer) · reads served · restart replays to Ended
    no event clears Ended; forking to a new dir (TASK-67) is not a transition of this world
```
