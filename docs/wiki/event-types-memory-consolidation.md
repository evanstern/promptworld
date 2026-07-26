---
name: event-types-memory-consolidation
description: Memory-embedding/consolidation event rows split from [[event-types]]: agent.memory_embedded/situation_embedded, journal.entry_written/deleted, the consolidation family, agent.belief_reinforced. Load when tracing spec 019 journals, spec 030 epistemic hygiene (Origin/belief evidence), or spec 042 embedding retrieval.
kind: concept
sources:
  - internal/sim/agents.go
  - internal/sim/journal.go
  - internal/sim/consolidate.go
verified_against: 8495b34ffb9ee5dc02e224025f0a23313bbab900
---

# Event types — memory embedding & consolidation events

Back to [[event-types]] for the payload-grammar conventions and the full
event-domain index.

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
rejection). One new whitelisted type, `agent.belief_reinforced`
(`BeliefReinforcedPayload{agent, belief_id}`), re-anchors a held belief's decay
clock at the grounded-observation seam — spec 030 ships the consumer (whitelist +
reducer arm) only, no in-tree emitter yet ([[nightly-consolidation]]).

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
| consolidation family: `agent.memory_promoted` / `agent.memory_faded` / `agent.belief_revised` / `agent.narrative_set` / `agent.consolidated` | payload structs in `internal/sim/consolidate.go`; contract in `specs/004-nightly-consolidation/contracts/` (spec 030 additions in `specs/030-epistemic-hygiene/contracts/`) | consolidation driver (injected) | salience boost / memory removal / belief create-or-revise / narrative replace / once-per-night ledger ([[nightly-consolidation]]); all reducer-total (vanished targets no-op); spec 030 threads two payload additions through — `belief_revised`'s `evidence` (the validator's resolved `MemoryRef{tick, hash}` citations) and `direct` (whether any cited evidence is direct perception; only a `direct` revision refreshes the belief's `Reinforced` decay anchor — a myth retold nightly on hearsay alone never re-anchors), and `consolidated`'s `coerced` (telemetry: beliefs the validator downgraded off `"witnessed"` for lack of direct evidence, never a rejection) |
| `agent.belief_reinforced` (spec 030 US2, FR-008) | `BeliefReinforcedPayload{agent, belief_id}` in `internal/sim/consolidate.go` | whitelisted through `InjectSocial`'s injection door (the grounded-observation seam) — ships as consumer only; no in-tree emitter yet, the perception-of-absence work is the intended future producer | re-anchors the named belief's `Reinforced` decay-clock field to `now` (`e.Tick`); a vanished belief id no-ops, reducer-total like its siblings |

## Connections

[[memory-retrieval]] owns `agent.memory_embedded`/
`agent.situation_embedded`/`cog.memory_divergence` end to end — the mind-side
embedder driver emits all three through [[sim-loop]]'s `InjectSocial` door,
and [[sim-state-reducer]] reduces the first two (the third is a `cog.*`
no-op).
