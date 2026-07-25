# Contract: event types (spec 044)

New and changed event-log types. All payloads are canonical-JSON structs; docs/wiki/
event-types.md gains one catalog row per type (mechanically enforced by the TUI digest
`TestCatalogSweep`).

## `run.ended` (new)

- **Emitter**: `stepEvents` only (pure executor emission — the `metatron.order_expired`
  precedent). Never injectable: no whitelist entries.
- **When**: in the same batch as the run's final `agent.died`, emitted at the **end of
  the batch** (after every same-tick death event and after all per-agent execution for
  the batch, so no event ever trails it), guarded by `!State.Ended` — exactly once per
  world, ever.
- **Payload** `RunEndedPayload`:
  - `tick` (int64) — declaration tick
  - `deaths` ([]{`agent` int, `tick` int64, `cause` string}) — every death, event order
  - `final_cause` (string)
- **Reducer**: sets `State.Ended = true`, `State.RunEnd`. Unknown-type no-op rule means
  old binaries ignore it safely.
- **Consumers**: loop (idle branch + command gating), status surfaces, TUI posture,
  scribe (morgue run-end summary), narrator (optional run epilogue), future TASK-119
  scenario fail signal.

## `agent.died` (changed: cause vocabulary + second emitter)

- **New emitter**: `gruStep`, immediately after its `gru.attacked` event, when the
  escalated attack lands at health 0. Emission includes the inline witness-death memory
  loop (radius 8, `salWitnessDeath`), since gru attacks are off the needs heartbeat.
- **Cause vocabulary**: `"starvation" | "exposure" | "collapse"` + **`"gru"`**. Cause is
  an open string; consumers switch on type, not cause — additive.
- **Reducer (changed)**: existing arm (Dead, wake, intent/hail clear, spec-013 inventory
  spill) **plus** grave placement: `Structure{Kind: "grave"}` at the death tile.
- **Invariant**: escalated deaths are byte-indistinguishable from heartbeat deaths
  downstream of the event (same reducer arm, chronicle, digest, morgue path).

## `gru.attacked` (changed: floor semantics)

- `GruAttackedPayload.Health` doc changes from "≥ gruWoundFloor" to: ≥ gruWoundFloor when
  the pre-attack health was ≥ `nearDeathBelow`; may be 0 when the target was already
  weakened (pre-attack health < `nearDeathBelow`). Payload shape unchanged.
- Doctrine comment (gru.go:12-20) updated: "it wounds the healthy; it can finish the
  already-fallen."

## `metatron.charter_observed` (new)

- **Emitter**: the Metatron turn pipeline, at effective-charter load, only when the
  content hash differs from `State.CharterFingerprint` (first turn always emits).
- **Payload** `CharterObservedPayload`: `fingerprint` (string, short hex of effective
  charter text), `default` (bool — matches `charterIsDefault`).
- **Reducer**: sets `State.CharterFingerprint`.
- **Purpose**: an event-sourced charter-revision timeline; the morgue aligns each death
  against the most recent observation ≤ the death tick (FR-008, SC-003). Evidence only —
  no scoring fields, by contract.

## `morgue.epilogue` (new)

- **Emitter**: mind narrator worker (LLM-gated), landed through `InjectSocial`; requires
  an `injectSocialWhitelist` entry. Absent in no-LLM worlds by construction.
- **Payload** `MorgueEpiloguePayload`: `agent` (int; villager index, or −1 for the
  run-end epilogue), `text` (string, bounded like chronicle lines).
- **Reducer**: appends to a bounded `State.MorgueEpilogues` ring (chronicle-ring
  pattern) for replica rendering.
- **Failure discipline**: chronicle doctrine — narrator absence/failure is a gap, never
  a stall; the factual morgue render never waits on it (FR-010).

## Ordering / determinism guarantees

1. Within the final batch: `agent.died`* → witness memories → `run.ended`.
2. No event after `run.ended` except daemon bookkeeping and recorded narrator prose
   (`morgue.epilogue`, `chronicle.entry`); the executor emits nothing once
   `State.Ended` (guard at the top of `stepEvents`).
3. All new emissions are pure functions of (recorded state, map, tick) — replay
   reproduces them; determinism harnesses must pass unmodified.
