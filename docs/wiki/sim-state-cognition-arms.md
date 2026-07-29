---
name: sim-state-cognition-arms
description: sim.State.Apply's cognition/telemetry arms — memory_added/embedded, situation_embedded, the journal family's rune-budgeted mutation path, the plan family, and the cog.* no-op telemetry types
kind: component
sources:
  - internal/sim/state.go
  - internal/sim/journal.go
verified_against: b35a7ffec46ba996741cdba4af9652fcfd163b32
---

# Sim state: cognition & telemetry arms

Split from [[sim-state-reducer]] (summary-style, corpus-spec v2): the arms
that grow an agent's recorded inner life — `agent.memory_added`,
the spec-042 embedding companions (`memory_embedded`/`situation_embedded`),
the journal family's rune-budgeted mutation path, and the plan family — plus
the `cog.*` telemetry types that are explicit reducer no-ops.

`agent.memory_added` copies the payload's situated context — `Where`/`Why`/`Conv`/
`Origin`, all `omitempty` — verbatim onto the appended `Memory` (spec 019/030:
baked at emission, never re-derived at render or replay, so live and replay
agree; `Origin` is the closed-vocabulary provenance class `DirectPerception`
reads to classify direct perception, absent classifying as secondhand, the
conservative default), and
additionally bumps `Agent.Generation` when the memory's
salience is at or above `GenerationBumpSalience` (9) — in-flight thoughts
snapshotted under the old generation are superseded at landing ([[cognition]],
[[sim-loop]]). Two spec 042 arms mutate state from the embedder driver's
companions ([[memory-retrieval]]): `agent.memory_embedded`
(`MemoryEmbeddedPayload{Agent, MemSeq, Vec, Model}`) scans the agent's
memories newest-first for the one whose `Seq` equals `MemSeq` and copies
`Vec`/`Model` onto it verbatim — a zero `MemSeq` never matches (so a pre-042,
seq-less memory can never be mistargeted) and a target that has died or
consolidated away is a deliberate no-op, never an error; `agent.situation_embedded`
(`SituationEmbeddedPayload{Agent, Tick, Text, Vec, Model}`) unconditionally
overwrites the agent's `SitVec`/`SitVecModel`/`SitVecTick` — later events
simply replace earlier ones, no history kept. The journal family (spec 019 US3) is the agent notebook's only
mutation path and, unlike the cognition telemetry types below, does mutate state:
`journal.entry_written` appends a reducer-id'd `JournalEntry` via `appendEntry`,
which enforces the per-agent `journalBudgetRunes` (4000) rune budget INSIDE
`Apply` — the budget participates in the accept/reject decision, so the
`InjectSocial` dry-run turns an over-budget append into a door rejection rather
than trusting handler courtesy (Principle III, the same door-facing gate the
miracle dry-run uses — [[agent-mind]]); `journal.entry_deleted`
removes an entry by id (ids never reused or renumbered, freed runes reclaimable),
a missing id erroring. The budget lives here as a version-stable sim constant,
not config, so a replay can never reject an event that landed live. The plan
family maintains `Agent.Plan`: `agent.plan_set`
replaces the steps, `agent.plan_step_started` pops the head, and
`agent.plan_expired` clears the whole remaining plan (a broken sequence is
not resumed) — and, since spec 043 (FR-005), also stamps the expired step
into the intent ring via `stampOrAppendExpired`: an open record matching the
step's goal closes `"expired"` (goal-matched so a concurrent non-plan intent
is never mis-stamped), otherwise a closed record is appended (the step
expired before ever firing). The hail family (TASK-47) maintains `Agent.Hail *AgentHail`

matching ended posture; the [[executor]] emits the event). The cognition telemetry types — `cog.thought`, `cog.outcome`,
`cog.recalibration_recommended`, (since spec 017)
`cog.tool_call` (the tool-use loop's call trace, [[tool-loop]]), and (since
spec 042 US2) `cog.memory_divergence` (the shadow-mode selector's rank-
divergence record, [[memory-retrieval]]) — are explicit
reducer no-ops: recorded observability with zero state effect.
`agent.intent_rejected`, formerly in that no-op list, is since spec 043 US1
split into its own STATE-MUTATING arm: the refused intent never landed, so
`Intent`/`IdleSince` stay untouched, but the ring gains the appended-closed
`"rejected"` record (see [[sim-state-intent-lifecycle]]) — deterministic from the event alone, so
replay-safe.
The spec-098 dream arms (`agent.salience_revised`/`agent.memory_merged`,
dispatched from the same switch to `applyDream`, `internal/sim/dream.go`)
apply the nightly clustering pass's recorded outcomes — [[private-dreams]]
owns them. Unknown types — including `daemon.*` and `world.created` — are recorded
history but state no-ops, so new event types never break old replay.

## Connections

Back to [[sim-state-reducer]] and its other five split-off notes.
[[cognition]] and [[sim-loop]] consume the `Generation` bump this arm
performs; [[memory-retrieval]] owns `Seq`/`Vec`/`VecModel`/`SitVec*`'s
producer and the `cog.memory_divergence` telemetry; [[agent-journal]] owns
the journal's roster tools and rune-budget doctrine; [[decision-context]]
consumes the plan-echo surface `Agent.Plan` renders. `agent.intent_rejected`'s
appended-closed record is stamped by [[sim-state-intent-lifecycle]].
