---
name: sim-loop-injection-doors-telemetry
description: Child of [[sim-loop-injection-doors]] — the observability/administrative half of InjectSocial's whitelist: cog.* telemetry, journal entries, belief_reinforced, the spec-042 memory-embedding companions, charter_observed/morgue.epilogue/skills_observed, guardian.report_card, and meeting.proposal_rephrased. Load when tracing why one of these types is or isn't whitelisted.
kind: component
sources:
  - internal/sim/loop.go
verified_against: 0af53ec6d211c71e298072c045c67ccbbd13b61d
---

# Sim loop — injection doors, the observability/administrative whitelist

Child of [[sim-loop-injection-doors]]: the observability and administrative
half of `InjectSocial`'s whitelist — telemetry, journal, belief, memory,
and morgue/report-card recording types — as distinct from the parent's
world/social-effecting types (visions, miracles, orders, designations,
prophecy).

## How it works

`meeting.proposal_rephrased` swaps an enacted norm's text and nothing else;
the `cog.*` telemetry — `cog.thought`, `cog.outcome`,
`cog.recalibration_recommended`, and (since spec 017) `cog.tool_call` (the
tool-use loop's per-call trace, [[tool-loop]]) — is whitelisted as reducer
no-ops so the [[cognition]] layer's observability is recorded, never
silent; and (since spec 019, US3) `journal.entry_written`/
`journal.entry_deleted` — the two mind-injectable journal mutations, whose
reducer dry-run enforces the rune budget (written) and entry existence
(deleted) before either lands; and (since spec 030 US2, FR-008)
`agent.belief_reinforced` — the grounded-observation seam that re-anchors a
held belief's decay clock (spec 030 ships the whitelist entry and reducer
arm only, no in-tree emitter yet); and (since spec 042 US1/US2) three more:
`agent.memory_embedded`/`agent.situation_embedded` — the mind-side
embedder's two vector companions ([[memory-retrieval]]), state-mutating
unlike the `cog.*` telemetry — door ordering guarantees a memory's
embedding companion never precedes the memory itself, since the embedder
only observes an `agent.memory_added` AFTER it is committed and notified;
and `cog.memory_divergence`, the shadow-mode selector's rank-divergence
record, riding the same reducer-no-op `cog.*` isolation class as the
telemetry types above; and (since spec 044 US2) two more:
`guardian.charter_observed` — the Guardian turn pipeline's
fingerprint-at-effect stamp, the event-sourced charter-revision timeline
the [[morgue]] aligns deaths against, whose reducer arm (and so the
dry-run) enforces a non-empty fingerprint — and `morgue.epilogue`, the
narrator's recorded mourning prose after a death or the run's end,
appending only the bounded `State.MorgueEpilogues` ring (never simulation
state, which is why it also survives the ended-world narrowing the parent
describes); and (since spec 077 FR-006) `guardian.skills_observed` — the
skills-observation twin of `charter_observed`: the bound skill-file set a
turn ran under, emitted on fingerprint change by the same pipeline
(`observeSkills`), whose reducer arm (and so the dry-run) enforces a
non-empty fingerprint AND a non-empty name list (an empty bound set is
never an observation); and (since spec 063, [[grounded-feedback]])
`guardian.report_card` — the guardian's report-card producer's stored
attribution note, recorded prose only, never simulation state; a
run-ending card rides `morgue.epilogue` instead, so this type deliberately
does NOT join the ended-world narrowing.

Since spec 036, whitelist membership is also readable from outside the
package via `InjectableSocialEvent(t)`, the single-source accessor both the
tool coverage gate and the bundle boot gate ([[bundle-tools]]) enforce
against — this accessor covers every type on both halves of the whitelist,
parent and child alike.

## Connections

Parent [[sim-loop-injection-doors]] covers the door mechanism, the
world/social-effecting whitelist half, and `InjectOperator`.
[[cognition]] owns the `cog.*` telemetry family; [[memory-retrieval]] owns
the memory-embedding companions and divergence record; [[morgue]] consumes
`charter_observed`/`morgue.epilogue`; [[grounded-feedback]] injects
`guardian.report_card`; [[bundle-tools]] enforces
`InjectableSocialEvent(t)` alongside the tool coverage gate.
