---
name: agent-persona-soul
description: How each villager's identity is seeded and rendered: eight write-once personas plus Guardian's editable charter at genesis, and the scribe's regenerable soul.md (beliefs, narrative, memories) rendered from the event log on memory/death/consolidation/governance/journal events. Split from [[agent-mind]] (Personas + Souls sections).
kind: component
sources:
  - internal/persona/personas.go
  - internal/persona/files.go
  - internal/scribe/scribe.go
  - internal/scribe/morgue.go
verified_against: d304e8adb64fdf40e24bfeca3ca3420e8a840a35
---

# Agent persona and soul rendering

**Personas** (`internal/persona`): eight authored natures, written exactly once by
`promptworld new` at mode 0444 into `agents/<name>/persona.md` — no post-genesis
write path exists anywhere (the structural half of the persona firewall; the
validation half is [[nightly-consolidation]]'s validator, fed by the authored
`persona.Anchors` and `persona.DriftMarkers`). `Load` reads them as the mind's
stable prompt prefixes. Genesis also seeds Guardian's `charter.md` (the ONE
player-editable prompt, never overwritten once present — [[guardian]]; since
spec 046 `Genesis` takes an optional variadic charter-preset parameter, so a
`"tutor"` world seeds `persona.TutorCharter` instead of the default —
[[curriculum-ladder]]), and the
salience table gains `SalDream` (8) for nudge memories. Since spec 019 (US3)
`Genesis` also seeds an empty `agents/<name>/journal.md` beside `soul.md`
(`JournalPath`, files.go) — a regenerable view of the agent's journal state the
scribe rewrites on every `journal.*` event, unlike the once-and-frozen persona
([[agent-journal]]).

**Souls** (`internal/scribe`): an always-on daemon component with its own replica
renders `agents/<name>/soul.md` (dated, starred memories, death freezes the header;
since TASK-9 also a "Who I am becoming" narrative section and a Beliefs section
with provenance and — since spec 030 — **effective**, decayed confidence
(`sim.EffectiveConfidence`, computed against the replica's current tick, never
the stored value; [[nightly-consolidation]]) rather than the stored number; a
belief whose effective confidence has fallen below the floor renders with no
number at all, as a hedged "half-remembered: ..." line, grouped after the
still-live convictions so the live ones read first) on memory/death/consolidation
events; since TASK-11
it also renders `chronicle.md` from the narrated story ring on `chronicle.entry`
events ([[chronicle]]), and since TASK-13 `village_charter.md` from the norm state
on governance events ([[governance]]); and since spec 019 (US3) it also renders
each agent's `journal.md` (`renderJournal`, on `journal.entry_written`/
`journal.entry_deleted` — a `jDirty` set kept separate from the soul `dirty`
set, since a journal mutation touches only that one file, souls unaffected;
[[agent-journal]]); and since spec 044 (US2) `morgue.md` (`renderMorgue` — a
whole-file re-render on `agent.died`/`run.ended`/`morgue.epilogue`, folding
the FULL event log through a fresh reducer state via the variadic
`EventSource` `scribe.New` now takes; [[morgue]]). The files are regenerable views — the event
log remains the only truth, so souls survive restarts and travel with the save dir.

Spec 019 (T024) kept soul.md's memory lines byte-identical for the common case:
a situated memory already carries its where/why IN THE TEXT (`situateText`, above),
so re-rendering them as a trailing suffix duplicated them ("Built a fire at the
woods (10,40). · at the woods (10,40)"). `memorySuffix` therefore renders ONLY
the one thing NOT in the text — a conversation memory's `Conv` ref, as
`· [conv <id>]`; place and why yield no suffix. A pre-019 memory (and any
non-conversation memory) yields "", so its line is byte-identical to the pre-019
format (FR-006/FR-014/SC-007); the structured `Where`/`Why` fields stay on the
`Memory` for programmatic consumers, only the redundant render is dropped.

## Connections

[[agent-mind]] is the parent note this child was split from — it hosts the
mind driver and tool-use loop that read the persona/soul state described
here; [[nightly-consolidation]] is the belief validator fed by the authored
`persona.Anchors`/`persona.DriftMarkers` and the digestion pass that writes
into the soul at sleep; [[guardian]] owns the player-editable `charter.md`
seeded here; [[curriculum-ladder]] owns the `tutor` charter preset threaded
through `Genesis`; [[agent-journal]] owns `journal.md`, rendered by the same
scribe machinery on its own dirty-set; [[chronicle]], [[governance]], and
[[morgue]] own the narrated `chronicle.md`, `village_charter.md`, and
`morgue.md` views the same always-on scribe daemon renders.

