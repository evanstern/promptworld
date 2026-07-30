---
name: event-types-memory-consolidation
description: Memory-embedding/consolidation event rows split from [[event-types]]: agent.memory_embedded/situation_embedded, journal.entry_written/deleted, the consolidation family, agent.belief_reinforced. Load when tracing spec 019 journals, spec 030 epistemic hygiene (Origin/belief evidence), or spec 042 embedding retrieval.
kind: concept
sources:
  - internal/sim/agents.go
  - internal/sim/journal.go
  - internal/sim/consolidate.go
verified_against: 1fae0d8536eb43e43eaa7b747aaeaf0b6e05ac83
---

# Event types — memory embedding & consolidation events

Back to [[event-types]] for the payload-grammar conventions and the full
event-domain index.

Spec 086 (agent-named payloads): every agent-referencing field in this
family's payloads is a `sim.AgentRef` — the wire carries
`{"id":N,"name":"…"}` objects (lists element-wise), the name stamped at
emission from the fixed roster via `Ref`/`Refs`; sentinels marshal
`{"id":-1,"name":""}`. Legacy bare-int rows decode through the dual-shape
unmarshal forever and reducer arms fold `.ID`s only — the conventions and
the normative back-compat matrix live on [[event-types]] ("Agent
references are named refs").
Spec 019 (grounded memories — situated
episodic memories + agent journal) adds **no** format bump: every addition is
`omitempty`, byte-stable against pre-019 logs. `MemoryAddedPayload` gains
situated context (`where`/`why`/`conv`), `IntentSetPayload` gains `reason`
(carried onto the intent so the executor can bake it into a memory's `why` at
completion), and TWO new whitelisted event types — `journal.entry_written` /
`journal.entry_deleted` — drive the agent-authored journal ([[agent-mind]]).

Spec 030 (epistemic hygiene — honest
provenance, hearsay decay, attribution-preserving gists) is likewise format-stable:
`MemoryAddedPayload` gains `Origin` (`omitempty`, the closed-vocabulary
provenance class — `action`/`witness`/`report`/`omen`/`gist`/`digest`, defined in
`internal/sim/memory.go` — stamped at every emission site; absent classifies as
secondhand, the conservative default, and `DirectPerception` is the only test the
belief validator runs against it), `BeliefRevisedPayload` gains `Evidence`
(`omitempty`, the resolved `MemoryRef{tick, hash}` identities a revision cites)
and `Direct` (`omitempty`, whether any cited evidence is direct perception — the
revision only refreshes the belief's decay anchor when true), and
`ConsolidatedPayload` gains `Coerced` (`omitempty`, telemetry: how many beliefs
the validator downgraded off `"witnessed"` for lack of direct evidence, never a
rejection). One new whitelisted type, `agent.belief_reinforced`,
re-anchors a held belief's decay clock; since spec 097 its producer exists —
the perception-of-absence reconciler (row below).

Spec 042 ([[memory-retrieval]] — embedding-augmented memory retrieval) is also
format-stable: `Memory` gains `omitempty` `Seq`/`Vec`/`VecModel` (`Seq` the
emitting `agent.memory_added` event's store seq — a pre-042 memory's field is
absent/0) and `Agent` gains `omitempty` `SitVec`/`SitVecModel`/`SitVecTick`,
so every pre-042 snapshot round-trips byte-identically. THREE new whitelisted
event types drive it: `agent.memory_embedded` and `agent.situation_embedded`
(the mind-side embedder's two state-mutating vector companions) and
`cog.memory_divergence` (a reducer-no-op telemetry record, the `cog.*`
convention) — full shapes and reducer effects in the table below. A world's
`memory_relevance` manifest flag (`""`/`shadow`/`on`, validated at
`world.Open`) gates whether the embedder and divergence recording run at all;
it carries no event of its own.

| Type | Payload struct | Emitted by | Reducer effect |
|---|---|---|---|
| `agent.memory_embedded` (spec 042 US1) | `MemoryEmbeddedPayload{agent, mem_seq, vec, model}` in `internal/sim/agents.go` | the mind-side embedder driver, injected via `InjectSocial` after observing the target's `agent.memory_added` commit ([[memory-retrieval]]) | scans the agent's memories newest-first for the one whose `Seq` equals `mem_seq` and copies `vec`/`model` onto it verbatim (the reducer never computes or inspects a vector); a `mem_seq` of 0 never matches (a pre-042, seq-less memory has no target identity); a vanished target (agent died, memory consolidated away) is a deliberate no-op, never an error |
| `agent.situation_embedded` (spec 042 US1) | `SituationEmbeddedPayload{agent, tick, text, vec, model}` in `internal/sim/agents.go` | the mind-side embedder driver, injected via `InjectSocial` at planner cadence ([[memory-retrieval]]) | overwrites `Agent.SitVec`/`SitVecModel`/`SitVecTick` with `vec`/`model`/`tick` — later events simply replace earlier ones, no history kept; `text` (the rendered situation template) rides the payload as an audit surface only, stored nowhere in state |
| `journal.entry_written` | `JournalWrittenPayload{agent, text}` (`journal.go`) | mind journal tool (`write_journal_entry`, injected via `InjectSocial` — spec 019 US3) | the ONLY journal-growth path: appends a reducer-id'd `JournalEntry{id, tick, text}` to the agent's `Journal` via `appendEntry`, which enforces the per-agent `journalBudgetRunes` (4000) rune budget INSIDE `Apply` — the `InjectSocial` dry-run turns an over-budget append into a door rejection, so no over-budget event lands (SC-005, [[agent-mind]]) |
| `journal.entry_deleted` | `JournalDeletedPayload{agent, entry}` (`journal.go`) | mind journal tool (`delete_from_journal`, injected) | removes the entry with that id from the agent's `Journal` (survivor order preserved, ids never reused or renumbered so freed runes are immediately reclaimable); a missing id errors at the door |
| consolidation family: `agent.memory_promoted` / `agent.memory_faded` / `agent.belief_revised` / `agent.narrative_set` / `agent.consolidated` | payload structs in `internal/sim/consolidate.go`; contract in `specs/004-nightly-consolidation/contracts/` (spec 030 additions in `specs/030-epistemic-hygiene/contracts/`) | consolidation driver (injected) | salience boost / memory removal / belief create-or-revise / narrative replace / once-per-night ledger ([[nightly-consolidation]]); all reducer-total (vanished targets no-op); spec 030 threads two payload additions through — `belief_revised`'s `evidence` (the validator's resolved `MemoryRef{tick, hash}` citations) and `direct` (whether any cited evidence is direct perception; only a `direct` revision refreshes the belief's `Reinforced` decay anchor — a myth retold nightly on hearsay alone never re-anchors), and `consolidated`'s `coerced` (telemetry: beliefs the validator downgraded off `"witnessed"` for lack of direct evidence, never a rejection); spec 105 adds `consolidated`'s `retries` (`omitempty`, telemetry: consumed truncation retries that night, accepted or rejected — `cost_usd` accrues across all ladder attempts) and the distinct rejected-marker reason `truncated` (`ConsolidationReasonTruncated`: the retry ladder exhausted while the reply was still detected cut — a budget failure, unlike `unparseable`'s garbage reply; no reducer change, the marker still closes the night without advancing `ConsolidatedUpTo`) |
| `agent.belief_reinforced` (spec 030 FR-008; spec 097 `kind`/`confidence`) | `BeliefReinforcedPayload{agent, belief_id, kind?, confidence?}` (`consolidate.go`) — `kind` `""` legacy \| `confirmed` \| `disconfirmed` | the spec-097 reconciler (`internal/mind/reconcile.go`), injected on an `agent.place_observed` landing ([[executor-perception-observation]]): confirm = effective + boost dial, disconfirm = effective × retain dial | re-anchors `Reinforced` to `e.Tick`; a Kind-stamped payload also copies `confidence` (clamped 0–100) — mind computes, reducer copies; a bare pre-097 payload stays the pure re-anchor; a vanished id no-ops |
| `agent.salience_revised` (spec 098) | `SalienceRevisedPayload{agent, mem_tick, text_hash, salience, reason}` (`dream.go`) | the private-dream pass (injected): habituation of a cluster member | sets the `(tick, hash)` memory's salience to the recorded value (clamped 1..`MaxSalience`); vanished target no-ops |
| `agent.memory_merged` (spec 098) | `MemoryMergedPayload{agent, kept, merged, salience}` (`dream.go`) | the private-dream pass (injected): a cluster folded into its representative | removes each `merged` `(tick, hash)` member, sets `kept`'s salience to the recorded value; vanished targets no-op |

Spec 098 (private dreams, [[nightly-consolidation]]) is format-stable: the
two whitelisted rows above carry the dream pass's recorded outcomes, and
`ConsolidatedPayload` gains `omitempty` `dream_folded`/`dream_kept`.

Spec 105 (truncation-aware retry, [[nightly-consolidation]]) is likewise
format-stable: `ConsolidatedPayload` gains `omitempty` `retries` (LAST, so a
zero-retry marker marshals byte-identically to a pre-105 one), `truncated`
joins the free-form rejected-marker reason vocabulary, and each consumed
retry rides the EXISTING `cog.outcome{retried}` record
([[cognition-horizon-telemetry]]) with class `consolidation` (the narrator's
ladder uses class `chronicle`) — no new event type, no whitelist change, no
format-version bump.

## Connections

[[memory-retrieval]] owns `agent.memory_embedded`/
`agent.situation_embedded`/`cog.memory_divergence` end to end — the mind-side
embedder driver emits all three through [[sim-loop]]'s `InjectSocial` door,
and [[sim-state-reducer]] reduces the first two (the third is a `cog.*`
no-op).
