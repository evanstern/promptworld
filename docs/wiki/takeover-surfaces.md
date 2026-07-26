---
name: takeover-surfaces
description: The spec-056 takeover family — the stage-unlock ceremony and the run-end postmortem, two full-screen pages that own the keyboard above every other TUI mode. The shared report-card renderer (D5) that composes into both takeovers, the guardian console's card seam, and the help overlay's ceremony replay is split to [[report-card-renderer]].
kind: component
sources:
  - internal/tui/tui.go
  - internal/tui/views.go
  - internal/tui/help.go
  - internal/skin/skin.go
verified_against: f3869df03de326d646d68ea716690da81e4840ca
---

# Takeover surfaces (ceremony + postmortem)

Spec 056 (TASK-127) adds a **takeover family**: two full-screen pages —
the stage-unlock **ceremony** and the run-end **postmortem** — that own the
body slot above EVERYTHING else in the TUI client (help, minibuffer,
console, inspect, villagers, global alike). "The takeover IS the event, not
a mode alongside the others" is the family's own framing: unlike the
guardian console (spec 053, which most global keys still pass through to),
a takeover swallows every key it doesn't explicitly name.

## How it works

**Ownership** (`internal/tui/tui.go`): `Model.takeover` is a `takeoverKind`
(`takeoverNone` / `takeoverCeremony` / `takeoverPostmortem`). Precedence is
total and one-way: postmortem always wins, replacing an open ceremony
unconditionally; takeovers never stack, and a same-kind arrival replaces
rather than queues. `handleKey`'s dispatch checks `m.takeover != takeoverNone`
immediately after `ctrl+c` — ahead of the help overlay, minibuffer focus,
the console, and every mode below — routing to `handleTakeoverKey`, whose
entire vocabulary is `esc` (dismiss one layer; on the postmortem also
latches `postmortemDismissed`) and `q` (quit/detach, the same unconditional
`m.quit()` every other quit path uses — `View()`'s own `runEnded()` check
picks the honest wording). Every other key, including `?`, is swallowed —
the help overlay does not open while a takeover is up.

**Transitions** (`applyEvent`, tui.go): folded AFTER the replica's own
`Apply` already latched the fact, so the takeover's render-time derivation
always finds the replica already reflecting whatever just landed — no
content is captured into a second `Model` field.

- `run.ended` → `takeoverPostmortem` unconditionally, clearing
  `postmortemDismissed` (a genuine run end always reopens, regardless of an
  earlier dismissal this session — the dismiss guard is only against
  re-annoying a player across a mere reconnect).
- `curriculum.stage_unlocked` → `takeoverCeremony`, UNLESS the postmortem
  already owns the slot, in which case the ceremony is deferred
  (`ceremonyDeferred = true`) rather than interrupting — its content stays
  reachable only through the replay surfaces below, never regenerated.

**Auto-open on attach** (`Update`'s `connectedMsg` case): a fresh connect to
an already-ended world opens the postmortem the same way the live
transition does (the dual-source `runEnded()` posture — `State.Ended` in
the snapshot needs no live `run.ended` replay), unless this CLIENT SESSION
already dismissed it (`postmortemDismissed` is per-session, not
per-reconnect, so a transient resync never re-annoys); `p` clears the flag
and reopens it from any mode, any time after a run has ended (inert on a
live world) — the global reopen `handleKey` checks after help/minibuffer so
it neither types into the buffer nor breaks help's own swallow-everything
rule.

**Layout-independent rendering** (`views.go`'s `takeoverView`): a first-class
page like the guardian console, checked ahead of it in `View()` — header +
body + footer regardless of widescreen/narrow layout, no minibuffer, no
dock/tabs (neither takeover accepts text input or has a second pane).
`fitTakeoverLines` truncates overflowing content to the panel's row budget
with a trailing "… (+N more)" count — the takeover family has no scroll key
at all, so an implausibly tall death ledger sheds its tail rather than
growing past `Height()`.

**The postmortem** (`postmortemView`): the narrated run-end line
(`postmortemRunEndLine`, computed directly from the recorded `run.ended`
payload's `RunEnd.FinalCause` rather than waiting on the async LLM-narrated
[[chronicle]] chapter — the takeover must render within one frame and stay
model-free; the wording mirrors `internal/mind/narrate.go`'s own `run.ended`
line verbatim), the report card when the world is scored (below), then the
[[morgue]]'s no-blame evidence rows always: one line per death — name, day,
cause, and the closest `metatron.charter_observed` fingerprint at or before
that death tick, scanned from the client's own bounded chronicle ring
(`closestCharterObservation`) rather than a file read; "unknown" is the
honest answer once the ring has rotated past the relevant observation on a
very long run, never a guess.

**The ceremony** (`ceremonyView`): the stage identity is always
`replica.StagesUnlocked`'s LAST entry while `takeover == takeoverCeremony`
— this is exactly the event that appended it, so no second `Model` field
remembers which stage is open. Renders the skin's D6 authorship-voice
narrated chapter (`Skin.CeremonyChapter(stage)` — detail split to
[[report-card-renderer]]) plus the report card that earned it
(`ceremonyReportCardFor`), title `"<STAGE NAME> — unlocked"`.
`provingPass` re-applies the SAME gate conjuncts `sim.EvaluateUnlock` uses to
find which recorded `CurriculumPass` earned the stage — a read-only twin for
replay purposes only (`EvaluateUnlock` itself refuses to re-evaluate an
already-latched stage); honest-false once the qualifying pass has aged out
of the bounded 32-entry `CurriculumPasses` retention, in which case the
report card identifies the proving exercise by stage (`provingStageFor`)
and grades `sim.EvaluateRubric` over the current replica with concluded
markers (spec 072's pinned fallback — current-state truth) instead of the
pass's own authoritative `Evidence`.

**Replayability** — split to [[report-card-renderer]]: the help overlay's
`?` ceremonies section re-renders every unlocked stage's chapter + report
card from the SAME helpers the live ceremony used, so a missed or dismissed
ceremony is never permanently lost.

## The shared report-card renderer (D5)

Split to [[report-card-renderer]]: one rubric-checklist implementation
(`reportCardFact`/`reportCardView`/`reportCardMode`), several sites — the
postmortem (concluded), the ceremony (concluded), and, via a `consoleCard`
wrapper, spec 053's guardian-console card seam (production wiring completed
by [[grounded-feedback]], TASK-115) — plus, since spec 072, the ONE shared
fact resolver (`resolveReportCardFacts`) all three surfaces derive verdicts
through: the recorded pass (all-met, evidence-backed) else
`sim.EvaluateRubric` over the replica, so a failed term renders honest ✗
everywhere and the surfaces can never disagree.

## Connections

[[scenario-machinery]] is the source of the report card's underlying
rubric data on a scenario world (`sim.ExerciseDefinition.RubricTerms`,
`sim.EvaluateUnlock`'s gate conjuncts, `sim.CurriculumPass.Evidence`) and of
`m.scenarioExercise()`, the manifest-fact lookup both takeovers share;
[[curriculum-ladder]] owns `sim.EvaluateUnlock`/`CurriculumPasses`/
`StagesUnlocked` the ceremony and its replay read; [[morgue]] is the
postmortem's no-blame evidence vocabulary and the death ledger
(`State.RunEnd.Deaths`) it projects; [[guardian]]'s `metatron.charter_observed`
timeline is what `closestCharterObservation` scans; [[event-types]]
catalogs `run.ended`/`curriculum.stage_unlocked`; [[skin]] supplies
`CeremonyChapter` and every other fiction string these pages render;
[[grounded-feedback]] (spec 063) is the production consumer that wires the
shared report-card renderer into the guardian console's live card seam and
the guardian's own attribution note; [[tui-client]] hosts the takeover
dispatch inside `Model`'s broader key/render loop.

## Operational notes

Both takeovers are static/derived renders — no IPC round trip, no new
event beyond the ones that already trigger them (`run.ended`,
`curriculum.stage_unlocked`), and the ceremony's replay section works with
zero model calls. `internal/tui/takeover_test.go`, `render_test.go`, and
`console_test.go` cover the dispatch/precedence/dismiss matrix, layout-
independent rendering, and the `consoleCard` seam composition respectively;
`help_test.go` extends the keymap/section sweep to the new ceremonies
section.
