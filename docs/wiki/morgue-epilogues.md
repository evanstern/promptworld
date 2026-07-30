---
name: morgue-epilogues
description: The morgue's narrated prose layer (internal/sim/morgue.go, internal/mind/narrate.go) — the single-flight narrator worker's morgue.epilogue events, the epilogue ring, the ended-world door slot shared with chronicle.entry, and the spec-063 report card that lands on this same channel on run end. Split from [[morgue]]; load for the prose (vs. fact) side of the run's legacy document.
kind: component
sources:
  - internal/sim/morgue.go
  - internal/mind/narrate.go
  - internal/sim/state.go
verified_against: 9b4ed5aef5bfea50b67fac10f8e2153f065a814d
---

# Morgue narrated epilogues

Prose is a separate, optional layer beneath the [[morgue]]'s factual fold —
this note covers it.

**The narrated epilogues** (`internal/sim/morgue.go`,
`internal/mind/narrate.go`): an absorbed `agent.died` or `run.ended` queues an
epilogue job on the [[chronicle]]'s single-flight narrator worker (same
`chronicle` decision class and `KindNarrator` route — no new model-call
class); one call under a fixed elegiac no-invention prompt lands the text as
a recorded `morgue.epilogue{agent, text}` event through `InjectSocial` —
`agent` is the mourned villager, or `-1` for the run end. The reducer arm
(`applyMorgueEpilogue`) rejects empty text and an out-of-range agent and
appends to `State.MorgueEpilogues`, a bounded ring (`morgueEpilogueCap` =
32 — a run produces at most one per death plus one for the run end; the cap
only guards against a misbehaving narrator re-mourning) on the chronicle-ring
pattern, so the scribe replica and attaching clients read it from state.
`morgue.epilogue` is one of the two prose types an ENDED world's narrowed
injection door still accepts ([[sim-loop]]'s `endedProseWhitelist`, with
`chronicle.entry`) — the run-end epilogue lands AFTER `run.ended` by
construction. The render blockquotes each epilogue after its section's facts
(facts before prose: removing every epilogue leaves a complete document),
and epilogues are collected in the same fold but EXCLUDED from the
byte-identity requirement — a suppressed, dropped, or failed epilogue is a
gap in the prose, never a stall of the factual record (FR-010). Since spec
063 ([[grounded-feedback]]), a run-ending guardian report card ALSO lands on
this same `morgue.epilogue` channel — `agent -1`, beside the narrator's own
run-end epilogue, prefixed "Report card (under charter `<fingerprint>`):
…" — since the ended door already narrows to recorded prose and no new
door entry is needed; a non-run-ending card instead rides its own
`guardian.report_card` type ([[event-types]]).

## Connections

[[morgue]] is the parent note this splits from — the deterministic factual
fold each epilogue is blockquoted beneath lives there. [[chronicle]] is the
sibling prose system — the narrator worker that writes epilogues is the
chronicle's own, and `morgue.epilogue` shares the ended-world door slot with
`chronicle.entry`. [[sim-loop]] owns the ended posture and the narrowed
`InjectSocial` door. [[sim-state-reducer]] holds the epilogue ring's reducer
arm. [[grounded-feedback]] (spec 063) is what lands a run-ending report card
on this channel — see [[guardian-report-card]] for its producer mechanics.
[[event-types]] catalogs `morgue.epilogue`. Spec: `specs/044-run-outcomes-morgue/`.

## Spec 086 — the epilogue's agent is a named ref

`MorgueEpiloguePayload.Agent` is a `sim.AgentRef` (−1 = the run-end
epilogue, marshaled `{"id":-1,"name":""}` — sentinels are legal and never
fake a name); a mourned villager's ref carries their roster name — death
never blanks a name (names are the fixed roster constant). The arm folds
`.ID` into the state ring.
