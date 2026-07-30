---
name: event-types-guardian-memory
description: Guardian memory-store event rows split from [[event-types]] (spec 102): guardian.memory_added/embedded/promoted/faded, guardian.salience_revised, guardian.memory_merged, guardian.consolidated — the agent.* consolidation family's guardian-side twins. Load when tracing the agentized guardian's store, its nightly consolidation, or its dream outcomes.
kind: concept
sources:
  - internal/sim/guardian_memory.go
  - internal/sim/loop.go
verified_against: d0645811c9783d1248dc65ed0fcf0b37524dd8fd
---

# Event types — guardian memory-store events (spec 102)

Back to [[event-types]] for the payload-grammar conventions and the full
event-domain index; [[guardian-agentization]] is the feature note.

All seven types are ADDITIVE vocabulary (spec 094 discipline — no format
bump): pre-102 logs never carried them, and a non-agentized world's guardian
never emits them (tuning `steward_cadence_ticks` 0 = off). All are injected
through the `InjectSocial` door (whitelisted in `internal/sim/loop.go`) and
reduced by `applyGuardianMemory` (`internal/sim/guardian_memory.go`) —
emitter computes, reducer applies, vanished targets no-op (spec 092). None
carries an agent ref: the store is the ONE guardian's own (D5 single-store
privacy).

| Type | Payload | Reducer effect |
|---|---|---|
| `guardian.memory_added` | `{text, salience}` (`GuardianMemoryPayload`) | append `sim.Memory{Text, Salience(clamped 1..10), Tick: envelope, Seq: envelope, Subject: -1, Origin: digest}` to `State.GuardianMemories`; refuses empty or >400-byte text; past 400 entries drops the lowest-salience (ties oldest) |
| `guardian.memory_embedded` | `{mem_seq, vec, model}` (`GuardianMemoryEmbeddedPayload`) | attach `Vec`/`VecModel` to the memory whose `Seq == mem_seq` (the `agent.memory_embedded` shape; emitted by the shared embedder driver) |
| `guardian.memory_promoted` | `{mem_tick, text_hash, boost}` | salience += boost, clamped at 10 — the nightly promote outcome |
| `guardian.memory_faded` | `{mem_tick, text_hash}` | remove the memory whole — the nightly fade outcome |
| `guardian.salience_revised` | `{mem_tick, text_hash, salience, reason}` | set salience to the recorded absolute value — the dream pass's habituation (reason `habituation`) |
| `guardian.memory_merged` | `{kept, merged[], salience}` (`MemoryRef`s) | remove merged members, set the kept memory's salience — the dream pass's cluster fold |
| `guardian.consolidated` | `{night, up_to, outcome, reason?, promoted?, faded?, dream_folded?, dream_kept?, cost_usd?}` | outcome `accepted` advances `State.GuardianMemUpTo` to `up_to` (closing the episodic buffer); `rejected` markers record and change nothing |

Memory identity is the shared `(tick, MemoryHash(text))` pair
(`sim.MemoryRef`) exactly as the villager consolidation family uses;
`guardian.memory_embedded` alone keys by `Seq` (its target may be re-hashed
by later merges, and seq is the emission-stable identity — the spec 042
`mem_seq` convention).

Chronicle digest rows for all seven live in `internal/tui/digest.go`
(`TestCatalogSweep` gates coverage both directions), rendered under the
skin's guardian display name — the payloads themselves are skin-free
(spec 052 ruling 1).
