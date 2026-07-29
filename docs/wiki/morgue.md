---
name: morgue
description: The run's legacy document (spec 044, TASK-31) — morgue.md at the world root, one factual epitaph per death plus a run-end summary, rendered as a deterministic genesis replay fold over the full event log; charter/orders evidence alignment, a no-blame vocabulary contract, and regenerability doctrine. The narrated prose layer (epilogues) splits to [[morgue-epilogues]].
kind: component
sources:
  - internal/scribe/morgue.go
  - internal/sim/morgue.go
  - internal/sim/state.go
  - internal/world/world.go
verified_against: 72f82f41f7aa2e345572105894cd0fb7c02fc0aa
---

# Morgue

The morgue is the run's legacy document (spec 044, TASK-31): `morgue.md` at
the world root (`world.MorguePath`, beside the charter and daemon files) —
one factual epitaph per death, closed by a village-level run-end summary
when the last villager falls. A run ends exactly once (`run.ended`, emitted
by the [[executor]]'s `stepEvents` as a batch's LAST event when every
villager still living at the tick's start died within the batch; the
reducer latches `State.Ended` and sets `State.RunEnd` — terminal, no event
ever clears it), and the morgue is what that run leaves behind: evidence of
what happened and what the guardian's instructions were, never a score.

## How it works

**A genesis replay fold, not the live replica**
(`internal/scribe/morgue.go`, `renderMorgue`): the render is a whole-file
regeneration and a PURE FOLD over the recorded history
(`contracts/morgue-document.md`). A fresh reducer state
(`sim.NewState(seed, map)`) replays the FULL event log from seq 0 through
the `EventSource` seam (`ReplayEvents` — `scribe.New` gained a variadic
`EventSource` parameter; the daemon passes the `*store.Store`, and a scribe
constructed without one renders no morgue), and each death's fields are
captured from the FOLDING state AT that event (`captureEpitaph`,
post-apply, so the epitaph includes the death's own ledger entry) — never
from the live replica, whose state keeps moving after a death and would
change a prior section's bytes later. Prior sections' factual bytes
therefore never change, replaying the same history reproduces the factual
content byte-identically (SC-004), and the whole factual side is model-free.
The scribe ([[agent-mind]]) re-renders on every batch carrying `agent.died`,
`run.ended`, or `morgue.epilogue`, and once at boot — an empty world gets
the "*No one has died. The village lives.*" posture line. Since spec 054,
`Scribe.SetScenario(exercise)` installs the armed scenario's exercise id and
re-renders immediately, so an already-ended scenario world's run summary
carries the exercise line from the first boot render on restart — called
once, by the daemon, right after `scribe.New` and before the sim loop
starts ([[scenario-machinery]], [[daemon-lifecycle]]).

**What an epitaph holds** (`captureEpitaph`, `writeEpitaph`): name, death
day, and cause; curated deeds (`morgueDeedNote` — the [[chronicle]]'s
notable-event vocabulary restricted to the deed-shaped subset: builds,
gifts, broken promises, chest thefts, gru encounters, governance arcs, norm
violations — each line rendered against the state AS THE EVENT FOUND IT, so
a broken promise still names its open debt); merged memories; standing
bonds (relations in canonical state order); open debts at death, both
directions; and the guardian's watch at that moment (below). Every epitaph
closes with the frame line "_Stated as evidence; the reader draws the
lesson._"

**The deterministic caps**: memories merge the at-death retained set with a
lifetime scan of `agent.memory_added` at salience ≥ `morgueNotableSalience`
(7) — necessary because [[nightly-consolidation]] deletes consolidated
memories from state, so at-death retention under-reports a long life —
deduped by (tick, text), ordered salience-desc, capped at `morgueMemoryCap`
(12). Deeds cap at `morgueDeedCap` (20) per epitaph and the run summary's
notable events at `morgueRunEventCap` (60), most recent kept with an
elision line ("_(N earlier deeds not shown)_") — truncation is always
stated, never silent.

**Charter and orders — the evidence alignment** (FR-008): each epitaph's
"the guardian's watch at that moment" aligns the death against the
event-sourced charter-revision timeline: the render collects every
`guardian.charter_observed` in the fold and pins the MOST RECENT observation
at or before the death tick — "charter revision `<fingerprint>` (default |
player-authored), in force since day N", or an explicit "no charter
observation recorded before this death" ([[guardian]] owns the emission: the
effective charter's content hash stamped at each turn). Since spec 046
([[curriculum-ladder]]) the observation's `default` flag is preset-aware —
default means the effective text equals the WORLD's charter-preset constant,
so a tutor-preset world running its stage-1 orientation charter records
`default: true` — preset text is authored by the game, and must never render
here as a player-authored charter. Beside it stand the standing orders
ACTIVE at the death moment, condition → action with watch subjects, read
from the folding state ([[guardian-orders]]). Instruction and outcome sit
together as evidence; the reader draws the lesson.

**The no-blame vocabulary contract** (contract invariant 2): the factual
render carries no scoring or blame language — `TestMorgueBannedVocabulary`
pins the document against word-boundary matches of `score`/`grade`/`blame`/
`fault`/`should have` ("default" is provenance, not blame), and the spec-044
event payloads carry no scoring fields by contract (`contracts/events.md`).

**The narrated epilogues** are a separate, optional prose layer beneath the
facts above (facts before prose — removing every epilogue leaves a complete
document); epilogues are EXCLUDED from the byte-identity requirement, so a
suppressed or failed one is a gap in the prose, never a stall of the
factual record (FR-010). See [[morgue-epilogues]] for the full mechanics —
the narrator job, the reducer arm, the ring cap, the ended-door whitelist,
and the spec-063 report-card landing.

**The run-end summary** (`writeRunSummary`, FR-009): rendered from
`State.RunEnd` (set verbatim by the `run.ended` arm from the payload's
complete death ledger — `State.Deaths` accretes in the `agent.died` arm, so
the emission needs no log scan): run length, the day-stamped population
decline curve, every death with cause, and the run's notable events in the
same curated deed vocabulary. Since spec 054 ([[scenario-machinery]]), on a
scenario world `writeRunSummary` also takes the scribe's `scenarioExercise`
id and appends one "**The exercise**: `<id>` — `<outcome>`. _Stated as
evidence; the reader draws the lesson._" line, `<outcome>` derived via
`sim.ExerciseOutcome` — the same no-blame register: failure is stated,
never scored. An ambient world (no scenario armed) renders this section
byte-identically to pre-054.

**Regenerability doctrine** (FR-011): `morgue.md` is a derived view, never a
source of truth — like the chronicle and village charter files. A deleted
or hand-edited morgue is healed byte-identically (facts) by the next
render, since everything factual folds from the [[event-log]] alone.

## Connections

[[morgue-epilogues]] holds the narrated prose layer this note splits off.
[[event-types]] catalogs `run.ended`, `morgue.epilogue`, and
`guardian.charter_observed`; [[executor]] emits `run.ended` and freezes an
ended world; [[sim-state-reducer]] holds the `agent.died` death ledger,
grave placement, and the `run.ended` latch; [[sim-loop]] owns the ended
posture and `Status.Ended`/`EndedDay`; [[guardian]] emits the
charter-revision observations epitaphs align against; [[guardian-orders]]
is the standing-order evidence; [[mental-maps]] carries the `grave`
place-fact kind; [[agent-mind]] hosts the scribe; [[world-save-directory]]
is where `morgue.md` lives; [[scenario-machinery]] is behind the run
summary's exercise-outcome line; [[takeover-surfaces]]' postmortem page also
renders the report-card prose. Spec: `specs/044-run-outcomes-morgue/`.

## Operational notes

The morgue accretes across the whole run — it exists (with the no-deaths
line) from first boot, grows an epitaph per death, and closes with the run
summary; it is readable at any point, not only after the run ends. The
facts/prose split is the operating rule: everything a test or dispute should
rely on is in the deterministic fold; the blockquoted epilogues are
best-effort narration that may simply be absent without making the document
incomplete.
