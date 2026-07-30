---
name: private-dreams
description: The spec-098 private dream phase — per-agent density clustering + habituation over recorded memory vectors in the nightly consolidation slot: geometry-first routing with an ambiguous-band consult, recorded salience-revision/merge outcomes, rngAt-seeded zeroable boundary jitter, and the five dream dials
kind: component
sources:
  - internal/sim/dream.go
  - internal/mind/consolidate.go
  - internal/sim/tuning.go
verified_against: 9b4ed5aef5bfea50b67fac10f8e2153f065a814d
---

# Private dreams (consolidation clustering + habituation)

Spec 098 (TASK-99, build-order step 3 of the memory-retrieval vault
analysis): each night, alongside the [[nightly-consolidation]] call, a
villager's mind reweighs its own memory store — dense clusters of
near-duplicate memories (forty berry-patch runs) habituate and fold so the
one vivid night stays sharp. Dreams are **private by construction** (D1):
`sim.PlanDream` (`internal/sim/dream.go`) takes ONE agent's memory snapshot
and nothing else — no shared vector table, no cross-agent index — proven by
the perturbation test (perturb one agent's store; the other's plan and
reduced post-night state are byte-identical, `TestPlanDreamPrivacyPerturbation`).

## How it works

**Geometry first (D2, the SAGE/RecMem economics).** The pass runs in the
consolidation worker (`internal/mind/consolidate.go runConsolidation`) over
the `consolJob`'s enqueue-time snapshot (memories, world seed, resolved
dials). Leader clustering in append order over vectored memories only
(same-model, the FR-009 guard; zero-magnitude vectors incomparable):
membership at cosine ≥ density−band, per-mille. Clusters of 3+
(`dreamMinClusterSize`) are candidates; a candidate's cohesion (mean
member↔leader cosine) at or above density+band is decided by geometry
ALONE — no model call — while the between band is reserved for the
consolidation LLM slot the night already owns (no new cognition classes; at
most `maxDreamGroupsSent` (4) groups ride the existing prompt as labeled
`[gN]` summaries with a `routine` output field; unknown labels are coerced
away, never a rejected night). A night whose buffer is empty makes no call,
so the band waits for a later night.

**Outcomes are recorded events (D3, spec 092 doctrine).** A decided cluster
keeps its most salient member (ties newer) vivid as representative, merges
the oldest members into it while the shared per-night cap lasts
(`agent.memory_merged`), and habituates the rest — salience × factor,
floor 1 — as absolute recorded values (`agent.salience_revised`,
reason `habituation`). Geometry outcomes inject IMMEDIATELY as their own
batch, so a deferred or rejected LLM night never undoes them; a folded
ambiguous group lands its precomputed events inside the accepted atomic
batch, and the marker carries `dream_folded`/`dream_kept` (the keep
decision's only trace). Reducer arms (`applyDream`) are total — vanished
targets no-op — and replay applies recorded outcomes with zero re-derivation
and zero model calls ([[event-types-memory-consolidation]] has the rows).

**Seeded boundary jitter (D4, adopted minimally).** Each candidate's
cohesion is nudged by a [[deterministic-rng]] draw — purpose-keyed
`rngAt(seed, "dream", night, agentIdx)`, amplitude `jitter_per_mille`,
zeroable — before classification, so borderline clusters occasionally
merge or survive differently ACROSS SEEDS while any one log replays
byte-identically. Jitter moves only the decision boundary; membership uses
the raw bar, so habituation of true duplicates is stable.

**The five dials** ride the [[world-tuning]] manifest as a nil≡default
`DreamTuning` block on `TuningState` (flat manifest keys, clamps in
`tuning.go`): `dream_density_per_mille` (900), `dream_ambiguous_band_per_mille`
(30), `dream_habituation_per_mille` (500), `dream_merge_cap_per_night` (4),
`dream_jitter_per_mille` (15, zeroable). Consumed via `State.DreamDials()`
off the mind's replica at snapshot time.

## Connections

[[nightly-consolidation]] owns the slot, trigger, and firewall this pass
rides; since spec 102 the agentized guardian's night runs the SAME
`PlanDream` over its own store (seat `sim.GuardianSeat`, outcomes landing
as `guardian.salience_revised`/`guardian.memory_merged` via
`GuardianDreamEvents` — [[guardian-agentization]]; single-store privacy
holds trivially there);
[[memory-retrieval]] records the vectors it clusters (no vectors — no
embedding route or `memory_relevance` off — means the pass finds nothing);
[[world-tuning]] carries the dials; [[deterministic-rng]] the jitter
pattern; [[sim-loop]]'s InjectSocial door whitelists the two event types;
[[tui-chronicle-feed]] digests them. Upstream design record:
`specs/098-private-dreams/` and `research/Agent-Memory-Retrieval/`
(SAGE density gate; RecMem cluster-triggered consolidation).

## Operational notes

Proven by `internal/sim/dream_test.go` (privacy perturbation, routing,
noise-zeroed equivalence, cross-seed boundary variance, reducer totality,
replay byte-identity) and `internal/mind/dream_test.go` (geometry lands
despite a rejected night; a routine verdict folds in the accepted batch;
group-less prompts stay byte-identical pre-098). Cost: zero new LLM calls —
the consult rides the existing nightly consolidation call or waits.
