---
name: nightly-belief-decay
description: Split from [[nightly-consolidation]] — how a belief's confidence decays purely on read (sim.EffectiveConfidence, an 8-game-day half-life since its Reinforced tick), the legacy-grandfather case for pre-spec-030 beliefs, the below-floor exclusion from model-facing prompts, and when the decay clock re-anchors
kind: component
sources:
  - internal/sim/consolidate.go
verified_against: a761a45cb3b437613b808408c6c7f30d11bd9eb9
---

# Nightly consolidation: belief confidence decay

Split from [[nightly-consolidation]] (sleep-triggered soul digestion): how a
held belief's confidence fades between nights, and when its decay clock
re-anchors to a fresh reinforcement.

## How it works

A newly formed belief always anchors its decay clock to formation
(`Reinforced = tick`, spec 030 US2); a revision refreshes that anchor ONLY when
`Direct` is true — a nightly retelling of pure hearsay changes the stored
confidence but must not keep the clock eternally fresh.

**Belief confidence decay** (spec 030 US2, `sim.EffectiveConfidence`,
`internal/sim/consolidate.go`): a belief's stored `Confidence` never mutates and
no decay event is ever logged — decay is computed purely on read, the same
precedent as memory recency (`SelectMemories` scores on read). Effective
confidence halves every `BeliefHalfLifeDays` (8) game-days since the belief's
`Reinforced` tick — an order of magnitude slower than a memory's own one-day
recency half-life, so convictions outlive vividness. A belief with
`Reinforced == 0` (any belief formed before spec 030) is a legacy grandfather:
it never decays until a revision or an `agent.belief_reinforced` event first
stamps an anchor. Below `BeliefConfidenceFloor` (20 — just under the rumor
tellability floor of 25, so the story keeps being retold after nobody stakes a
decision on it) a belief stops driving behavior: read sites should drop it from
model-facing prompts (`sim.PromptBeliefs` is the shared exclusion helper; the
nightly held-beliefs prompt on [[nightly-consolidation]] is the one documented
exception, marking rather than dropping so faded beliefs stay revisable) and
the scribe renders it hedged rather than as a live conviction ([[agent-mind]]).
**The reinforcement seam has its producer** (spec 097): `agent.belief_reinforced`
(`BeliefReinforcedPayload{Agent, BeliefID, Kind?, Confidence?}`) re-anchors a
held belief's clock to now, and — since spec 097's additive `Kind`/`Confidence`
fields — also copies an emitter-computed new stored confidence when `Kind` is
`confirmed` or `disconfirmed` (clamped 0–100; the legacy bare shape stays the
pure re-anchor, replaying byte-identically). The producer is the
perception-of-absence channel ([[executor-perception-observation]]):
`internal/mind/reconcile.go` judges each `agent.place_observed` against the
observer's place-beliefs — CONFIRMED (feature present) sets confidence to
effective + the `belief_confirm_boost` dial; DISCONFIRMED (place observed,
feature absent from the exhaustive scan) sets it to effective ×
`belief_disconfirm_retain_percent` — faster than this note's silence half-life
but bounded, so a myth survives several visits before trending under the floor.
Silence (never visiting) keeps exactly the decay described above.

## Connections

[[nightly-consolidation]] owns the nightly call that revises beliefs and
computes each revision's `Direct` flag (`enforceProvenance`), the input this
note's re-anchoring rule reads; [[agent-mind]] renders effective (decayed)
confidence into soul.md's Beliefs section; [[event-types]] catalogs
`agent.belief_reinforced`'s payload shape;
[[executor-perception-observation]] owns the spec-097 observation channel
that produces the reinforcements.

Back to [[nightly-consolidation]] for the trigger/ledger, the consolidation
call, the firewall validator, the provenance gate, and landing.
