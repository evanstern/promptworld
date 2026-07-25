---
name: morgue
description: The run's legacy document (spec 044, TASK-31) — morgue.md at the world root, one factual epitaph per death plus a run-end summary, rendered as a deterministic genesis replay fold over the full event log, with narrated morgue.epilogue events blockquoted beneath the facts; charter/orders evidence alignment, a no-blame vocabulary contract, and regenerability doctrine
kind: component
sources:
  - internal/scribe/morgue.go
  - internal/sim/morgue.go
  - internal/sim/state.go
  - internal/mind/narrate.go
  - internal/world/world.go
verified_against: d9d74924621b8816bbb4608afe48c41cda4321d7
---

# Morgue

The morgue is the run's legacy document (spec 044, TASK-31): `morgue.md` at
the world root (`world.MorguePath`, beside the charter and daemon files) —
one factual epitaph per death, closed by a village-level run-end summary
when the last villager falls, with the narrator's recorded epilogues
blockquoted beneath each section's facts. A run ends exactly once
(`run.ended`, emitted by the [[executor]]'s `stepEvents` as a batch's LAST
event when every villager still living at the tick's start died within the
batch; the reducer latches `State.Ended` and sets `State.RunEnd` — a
terminal posture no event ever clears), and the morgue is what that run
leaves behind: evidence of what happened and what the angel's instructions
were when it happened, never a score.

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
from the live replica, whose relations, debts, and standing orders keep
moving after a death and would change a prior section's bytes on a later
render. Prior sections' factual bytes therefore never change, replaying the
same history reproduces the factual content byte-identically (SC-004), and
the whole factual side is model-free: it renders with the LLM subsystem off.
The scribe ([[agent-mind]]) re-renders on every batch carrying `agent.died`,
`run.ended`, or `morgue.epilogue`, and once at boot — an empty world gets
the "*No one has died. The village lives.*" posture line.

**What an epitaph holds** (`captureEpitaph`, `writeEpitaph`): name, death
day, and cause; curated deeds (`morgueDeedNote` — the [[chronicle]]'s
notable-event vocabulary restricted to the deed-shaped subset: builds,
gifts, broken promises, chest thefts, gru encounters, governance arcs, norm
violations — each line rendered against the state AS THE EVENT FOUND IT, so
a broken promise still names its open debt); merged memories ("what they
carried in memory"); standing bonds ("who mattered to them", relations in
canonical state order); open debts at death, both directions; and the
angel's watch at that moment (below). Every epitaph closes with the frame
line "_Stated as evidence; the reader draws the lesson._"

**The deterministic caps**: memories merge the at-death retained set with a
lifetime scan of `agent.memory_added` at salience ≥ `morgueNotableSalience`
(7) — necessary because [[nightly-consolidation]] deletes consolidated
memories from state, so at-death retention under-reports a long life —
deduped by (tick, text), ordered salience-desc, capped at `morgueMemoryCap`
(12). Deeds cap at `morgueDeedCap` (20) per epitaph and the run summary's
notable events at `morgueRunEventCap` (60), most recent kept with an
explicit elision line ("_(N earlier deeds not shown)_") — truncation is
always stated, never silent.

**Charter and orders — the evidence alignment** (FR-008): each epitaph's
"the angel's watch at that moment" aligns the death against the
event-sourced charter-revision timeline: the render collects every
`metatron.charter_observed` in the fold and pins the MOST RECENT observation
at or before the death tick — "charter revision `<fingerprint>` (default |
player-authored), in force since day N", or an explicit "no charter
observation recorded before this death" ([[metatron]] owns the emission: the
effective charter's content hash stamped at each turn). Since spec 046
([[curriculum-ladder]]) the observation's `default` flag is preset-aware —
default means the effective text equals the WORLD's charter-preset constant,
so a tutor-preset world running its stage-1 orientation charter records
`default: true`; the seam this fixes is exactly this evidence line — preset
text is authored by the game, and must never render here as a player-authored
charter. Beside it stand the
standing orders ACTIVE at the death moment, condition → action with watch
subjects, read from the folding state ([[metatron-orders]]). Instruction and
outcome sit together as evidence; the reader draws the lesson.

**The no-blame vocabulary contract** (contract invariant 2): the factual
render carries no scoring or blame language — `TestMorgueBannedVocabulary`
(`morgue_test.go`) pins the rendered document against word-boundary matches
of the banned words `score`/`grade`/`blame`/`fault`/`should have` ("default"
is provenance, not blame), and the spec-044 event payloads carry no scoring
fields by contract (`contracts/events.md`).

**The narrated epilogues** (`internal/sim/morgue.go`,
`internal/mind/narrate.go`): prose is a separate, optional layer. An
absorbed `agent.died` or `run.ended` queues an epilogue job on the
[[chronicle]]'s single-flight narrator worker (same `chronicle` decision
class and `KindNarrator` route — no new model-call class); one call under a
fixed elegiac no-invention prompt lands the text as a recorded
`morgue.epilogue{agent, text}` event through `InjectSocial` — `agent` is the
mourned villager, or `-1` for the run end. The reducer arm
(`applyMorgueEpilogue`) rejects empty text and an out-of-range agent and
appends to `State.MorgueEpilogues`, a bounded ring (`morgueEpilogueCap` =
32 — a run produces at most one per death plus one for the run end; the cap
only guards against a misbehaving narrator re-mourning) on the chronicle-
ring pattern, so the scribe replica and attaching clients read it from
state. `morgue.epilogue` is one of the two prose types an ENDED world's
narrowed injection door still accepts ([[sim-loop]]'s `endedProseWhitelist`,
with `chronicle.entry`) — the run-end epilogue lands AFTER `run.ended` by
construction. The render blockquotes each epilogue after its section's facts
(facts before prose: removing every epilogue leaves a complete document),
and epilogues are collected in the same fold but EXCLUDED from the
byte-identity requirement — a suppressed, dropped, or failed epilogue is a
gap in the prose, never a stall of the factual record (FR-010).

**The run-end summary** (`writeRunSummary`, FR-009): rendered from
`State.RunEnd` (set verbatim by the `run.ended` arm from the payload's
complete death ledger — `State.Deaths` accretes in the `agent.died` arm
precisely so the emission needs no log scan): run length, the day-stamped
population decline curve, every death with cause, and the run's notable
events in the same curated deed vocabulary.

**Regenerability doctrine** (FR-011): `morgue.md` is a derived view, never a
source of truth — exactly like the chronicle and village charter files. A
deleted or hand-edited morgue is healed byte-identically (facts) by the next
render, because everything factual folds from the [[event-log]] alone.

## Connections

[[chronicle]] is the sibling prose system — the narrator worker that writes
epilogues is the chronicle's own, and `morgue.epilogue` shares the
ended-world door slot with `chronicle.entry`; [[event-types]] catalogs
`run.ended`, `morgue.epilogue`, and `metatron.charter_observed`;
[[executor]] emits the `run.ended` declaration and freezes an ended world
(the `stepEvents` guard); [[sim-state-reducer]] holds the `agent.died` death
ledger + grave placement, the `run.ended` latch, and the epilogue ring;
[[sim-loop]] owns the ended posture (mutators refused, `InjectSocial`
narrowed to recorded prose) and the `Status.Ended`/`EndedDay` surface;
[[metatron]] emits the charter-revision observations the epitaphs align
against; [[metatron-orders]] is the standing-order evidence; [[mental-maps]]
carries the `grave` place-fact kind a death leaves in the world;
[[agent-mind]] hosts the scribe that renders the file and the absorb hook
that queues epilogues; [[world-save-directory]] is where `morgue.md` lives.
Spec: `specs/044-run-outcomes-morgue/`.

## Operational notes

The morgue accretes across the whole run — it exists (with the no-deaths
line) from first boot, grows an epitaph per death, and closes with the run
summary; it is readable at any point, not only after the run ends. The
facts/prose split is the operating rule to remember: everything a test or a
dispute should rely on is in the deterministic fold; the blockquoted
epilogues are best-effort narration that may simply be absent (no llm.json,
a suppression, a full queue) without making the document incomplete.
