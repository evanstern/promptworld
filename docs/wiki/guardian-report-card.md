---
name: guardian-report-card
description: Spec-063's report-card producer — a cheap-chain attribution note citing recorded events (guardian.report_card / KindReportCard), stored on State.GuardianReportCard or the morgue.epilogue channel on run end, and the console card seam that composes it beneath the rubric checklist. Split from [[grounded-feedback]]; load for the attribution half of the feedback layer.
kind: component
sources:
  - internal/guardian/reportcard.go
  - internal/guardian/guardian.go
  - internal/sim/reportcard.go
  - internal/sim/state.go
  - internal/sim/loop.go
  - internal/llm/llm.go
  - internal/llm/config.go
  - internal/cognition/registry.go
  - internal/skin/skin.go
  - internal/tui/reportcard.go
  - internal/tui/tui.go
  - internal/tui/digest.go
  - internal/tui/grammar.go
verified_against: 74fe956813aa6be54e65156ae9bfcb91745cbb8d
---

# Guardian report card

Spec 063 (TASK-115)'s second grounded-feedback surface — see
[[grounded-feedback]] for the parent note and [[explain-tutor-guide]] for the
orientation half. The report card is a durable record tying what the
guardian did back to the charter text that authorized it; it attributes, it
never scores.

## The report-card producer

**The report-card producer** (`internal/guardian/reportcard.go`, US4): a
stopping-point consumer on the digest worker's own notify-consumer pattern
(the `digestWorker` shape — `cardQ`, `reportCardWorker`). The absorb
goroutine collects the guardian's own recorded activity into a bounded
`cardTrail` ring (`cardTrailMax` 48 — `cog.tool_call` records from guardian
turns, landed `guardian.*` acts) and fires a bounded card job at three
stopping points: `run.ended`, `curriculum.exercise_passed`, and a
`clock.paused` episode (debounced — at most one card per pause episode via
`cardPauseSpent`, re-armed on `clock.resumed`). Both trigger scans run on
the ABSORB goroutine, after the replica applies the batch (the
`matchOrders` discipline, [[guardian-order-triggering]]) — a card can never
fire during replay, since no guardian runs during reconstruction.
`enqueueCard` is activity-gated: with no new guardian activity since the
last card (tracked via `cardDoneSeq`, the graded high-water mark), nothing
is queued — the deterministic checklist half stands alone.

`produceCard` runs ONE cheap-chain call (`llm.KindReportCard`, below) per
job: the system prompt casts the model as the guardian "writing a brief
report card on your own recent service," instructed to attribute outcomes
to the charter's own words and cite evidence by `seq N` — `cardPrompt`
renders the charter revision in force, the recorded trail with real seqs,
and the R1 mechanics facts (`tool.ExplainSheet("charges", …)` — the SAME
fact source `explain` serves, so the card's own arithmetic can never
disagree with what a player could ask for directly). `validateCitations`
(`citationRe`, `\bseq\s+(\d+)\b`) extracts every cited seq and checks it
against the fed trail — a note citing an unrecorded seq is dropped WHOLE
(deterministic degradation beats a plausible fabrication); every other
failure path (no orchestrator, empty critique, a door rejection) is a
silent skip with one log line, never error theater, and play is never
blocked.

## Storage — two doors, one intent

A non-run-ending card lands as
`guardian.report_card` (`GuardianReportCardPayload{Fingerprint, Note,
Citations}`, whitelisted in [[sim-loop]]'s `InjectSocial` door) — the
reducer arm (`internal/sim/reportcard.go`'s `applyReportCard`) validates
rather than clamps (non-empty fingerprint/note, note ≤ `reportCardNoteMaxRunes`
1200, every citation strictly precedes the card event's own seq — a card
can never cite the future) and keeps only the LATEST card on
`State.GuardianReportCard *GuardianReportCard` (`omitempty`; the log keeps
history). A run-ending card instead rides the EXISTING `morgue.epilogue`
channel (agent `-1`, beside the narrator's own run-end epilogue, prefixed
"Report card (under charter `<fingerprint>`): …") — the ended door already
narrows to recorded prose ([[morgue]]), so `guardian.report_card` deliberately
does NOT join `endedProseWhitelist` ([[morgue]] renders both epilogues in the
same blockquoted section).

## KindReportCard

**`KindReportCard`** (`internal/llm/llm.go`/`config.go`): a new accepted
`Kind` ("report_card", frozen from birth per spec 052 ruling 2), routed by
default `local→cloud` (the `KindGuardianWatch` cheap-first shape) and
classified into the existing `metatron` cognition class
(`internal/cognition/registry.go` — same actor, `DegradeSkip`,
event-triggered at stopping points, never cadence-scheduled). A pre-063
`llm.json` backfills the route from `defaultRoutes()` with a boot log line
(`defaultBackfillKinds`, the `KindGuardianWatch` precedent).

## The report-card console seam

**The report-card console seam** (`internal/tui/reportcard.go`, T012, spec
063 standing resolution 1 — "one composed card artifact"): [[takeover-surfaces]]
shipped the shared `reportCardView` renderer and an empty `consoleCard`
seam; this feature is the production wiring. `rebuildConsoleCards` composes,
in order: the rubric checklist (`buildChecklistCard`, TASK-127's
`reportCard` wrapper — nil until a stopping point is visible in durable
state: a stored note, a recorded pass for the seeded exercise, or the ended
run; facts and marker vocabulary from the spec-072 shared resolver
`resolveReportCardFacts` ([[report-card-renderer]]): the recorded pass
re-read all-met when one exists, else `sim.EvaluateRubric` over the replica
— live `met/pending` until the exercise concludes or the run ends, then
`met/missed` with honest ✗ on failed terms) FIRST — always authoritative —
then the
attribution note (`buildNoteCard`, `noteCard`) — additive prose beneath it,
clearly its own block, never a second scoring computation. Either half
absent degrades to the other alone. Recomposed on connect (a late attach
re-reads the stored card, no badge) and on every stopping-point-relevant
event (`guardian.report_card`, `curriculum.exercise_passed`, `run.ended`) —
a fresh NOTE additionally sets the existing unseen-badge flag
(`guardianUnseen`) when the guardian pane isn't visible: at most a badge
between stopping points, never a takeover (FR-006). `guardian.report_card`
also gets its own [[tui-client]] digest-registry entry in the `"guardian"`
namespace of `familyByNamespace` (guardian voice, the `curriculum`
precedent) — a namespace that, since spec 094's rename, also carries the 13
world-action types formerly under `metatron.*`.

## Connections

[[grounded-feedback]] is the parent note this splits from; [[explain-tutor-guide]]
is its sibling (the orientation half) — the report card's mechanics facts
come from the same `explain` sheets that note describes. [[guardian]] hosts
the absorb loop that drives the report-card producer; [[guardian-miracles]]
shares the mirrored `MiracleCost` source the card's charges arithmetic reads;
[[curriculum-ladder]] owns `sim.EvaluateUnlock`/`StagesUnlocked` the pass-derived
facts read; [[takeover-surfaces]] owns the shared `reportCardView` renderer
and the `consoleCard` seam this feature wires into production;
[[scenario-machinery]] is `sim.ExerciseDefinition`'s `RubricTerms` source both
the checklist card and the producer's mechanics facts ground in;
[[event-types]] catalogs `guardian.report_card`; [[sim-state-reducer]] owns
`State.GuardianReportCard` and the `applyReportCard` arm; [[morgue]] is where
a run-ending card's note actually lands; [[llm-orchestrator]] routes
`KindReportCard`; [[cognition]] classifies it into the `metatron` decision
class; [[skin]] supplies `ReportCardLabel`/`AttributionLabel`/`ExampleAsk`;
[[tui-client]] hosts the digest entry, the namespace mapping, and the
console-card recomposition this note's producer and consumer sides meet at.
