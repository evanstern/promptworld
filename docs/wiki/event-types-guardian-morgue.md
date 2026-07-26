---
name: event-types-guardian-morgue
description: Guardian morgue/report-card event rows split from [[event-types]]: metatron.charter_observed, metatron.skills_observed (spec 077), morgue.epilogue, guardian.report_card, chronicle.entry. Load when tracing spec 044 charter-fingerprint observation, the spec 077 skills observation, run epilogues, or the spec 063 grounded-feedback report card.
kind: concept
sources:
  - internal/sim/guardian.go
  - internal/sim/morgue.go
verified_against: 93837e1885bff17114df75e5382ac60dee24776a
---

# Event types — guardian morgue & report-card events

Back to [[event-types]] for the payload-grammar conventions and the full
event-domain index.

Spec 063 (the grounded feedback layer — [[grounded-feedback]], TASK-115) is
also format-stable: `State` gains `omitempty` `GuardianReportCard
*GuardianReportCard` (the reducer keeps only the LATEST stored card; the
log keeps history, so a pre-063 snapshot round-trips byte-identically). ONE
new whitelisted type, `guardian.report_card`
(`GuardianReportCardPayload{fingerprint, note, citations?}`), rides
[[sim-loop]]'s `InjectSocial` door — a run-ending card instead rides the
EXISTING `morgue.epilogue` channel (agent `-1`), so this new type
deliberately does NOT join `endedProseWhitelist`. No new tool-registry
event vocabulary: `explain` is `Effect: Read` and lands no event of its own.

| Type | Payload struct | Emitted by | Reducer effect |
|---|---|---|---|
| `metatron.charter_observed` (spec 044 US2) | `CharterObservedPayload{fingerprint, default}` in `internal/sim/guardian.go` — `fingerprint` is a short content hash (12 hex chars) of the EFFECTIVE charter text a turn ran under; `default` marks the authored default | the guardian turn pipeline (`observeCharter`, injected via `InjectSocial` at charter load), only when the fingerprint differs from `State.CharterFingerprint` — the first turn always emits; skipped on ended worlds | sets `State.CharterFingerprint`, (spec 072) `State.CharterCustom = !default` — the persisted authorship flag charter-conjunct rubrics read ([[scenario-machinery]]) — and (spec 077 FR-005) `State.CharterObservedSeq/Tick` from the event envelope (the `PlacedSeq` apply-time-stamp precedent), the coordinates `CharterEvidenceFromState` re-locates for pass evidence without a log scan; the log's observation sequence is the event-sourced charter-revision timeline the morgue aligns each death against (most recent observation ≤ the death, [[guardian]]) |
| `metatron.skills_observed` (spec 077 FR-006) | `SkillsObservedPayload{fingerprint, names}` in `internal/sim/guardian.go` — `fingerprint` is a short content hash (12 hex chars) over the BOUND skill set's names+texts in composition order; `names` the bound file list | the guardian turn pipeline (`observeSkills`, the `observeCharter` twin — injected via `InjectSocial` at skills bind), only when the fingerprint differs from `State.SkillsFingerprint`; an EMPTY bound set never emits (absence is not an observation), so stages 1–2 — where `stageSkills` refuses to bind — are structurally silent; skipped on ended worlds | sets `State.SkillsFingerprint` and `State.SkillsObservedSeq/Tick` from the envelope — the coordinates `SkillsObservedEvidence` reads (`Custom: true` by construction: only players author skill files, and only stage-3+ binds them — the stage-3→4 gate's evidence, [[curriculum-ladder-progression]]) |
| `morgue.epilogue` (spec 044 US2) | `MorgueEpiloguePayload{agent, text}` in `internal/sim/morgue.go` — `agent` is the mourned villager, or −1 for the run-end epilogue | mind narrator worker (injected via `InjectSocial` on absorbing `agent.died`/`run.ended`; LLM-gated, so structurally absent in no-model worlds); since spec 063 also the guardian's report-card producer at run end (agent −1, prefixed "Report card (under charter `<fingerprint>`): …" — [[grounded-feedback]], since the ended door already narrows to recorded prose and no new door entry is needed); one of the two prose types an ENDED world's door still accepts (`endedProseWhitelist`) | appends the bounded `State.MorgueEpilogues` ring (chronicle pattern) for replica/scribe rendering; the morgue's factual render never depends on it — narrator absence or failure is a gap, never a stall |
| `guardian.report_card` (spec 063 US4, [[grounded-feedback]]) | `GuardianReportCardPayload{fingerprint, note, citations?}` in `internal/sim/reportcard.go` | the guardian's report-card producer (injected via `InjectSocial`) at a non-run-ending stopping point (an exercise pass, or a debounced pause episode) — never on an ended world, where the run-end card rides `morgue.epilogue` instead | validates (non-empty fingerprint/note, note ≤1200 runes, every cited seq strictly precedes this event's own seq — a card can never cite the future) then keeps only the LATEST card on `State.GuardianReportCard` (`omitempty`); the log keeps every prior card as history |
| `chronicle.entry` | `ChronicleEntryPayload{day, from_tick, to_tick, text, thread, agents}` in `internal/sim/chronicle.go` | narrator driver (injected, TASK-11) | appends the bounded `State.Chronicle` ring ([[chronicle]]) |

## Connections

[[morgue]] owns the spec 044 surface — the [[executor]] emits
`run.ended`, the guardian turn pipeline injects `metatron.charter_observed`
(reduced in `internal/sim/guardian.go`), and the mind narrator injects
`morgue.epilogue` (reduced in `internal/sim/morgue.go`);

[[grounded-feedback]] owns `guardian.report_card` end to
end (spec 063) — the payload and reducer arm in `internal/sim/reportcard.go`,
the report-card producer in `internal/guardian` that injects it, and the
`State.GuardianReportCard` field this reducer keeps;

[[morgue]] is where a
run-ending card's note lands instead (the existing `morgue.epilogue`
channel);

[[takeover-surfaces]] is the spec-056 consumer that composes the
report card's rubric checklist into the postmortem and ceremony takeovers.
