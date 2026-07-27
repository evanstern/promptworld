---
name: chronicle
description: The narrated story feed — a cloud narrator compresses notable events into chapter entries (chronicle.entry) on a snapshot-carried ring, the ambient world's catch-up mechanism; spec 044 rides morgue epilogues on the same single-flight worker and router class. Load for narration behavior, chapter/epilogue failure handling, or the chronicle/morgue views.
kind: component
sources:
  - internal/sim/chronicle.go
  - internal/mind/narrate.go
  - internal/scribe/scribe.go
verified_against: c61cd6c04ddfcd2a976c14a49ba071e8fd768a73
---

# Chronicle

The chronicle is the catch-up mechanism for the ambient world (the grounding's
core posture: days pass unattended; the story must be readable on return). A
cloud-tier narrator compresses the event stream into a few story entries per
game day; entries are events like everything else, and a bounded ring on
`State` carries the readable history to every attaching client for free.

## How it works

**The event and the ring** (`internal/sim/chronicle.go`): `chronicle.entry`
carries `ChronicleEntryPayload{day, from_tick, to_tick, text, thread, agents}`.
The reducer appends it to `State.Chronicle`, a bounded ring (`chronicleCap`
= 256 — weeks of story at the narrator's ~2 chapters/game-day; the
[[event-log]] keeps everything forever). Because the ring rides
`State.Marshal`, the TUI's state-snapshot fetch and daemon recovery both
deliver narrated history with no extra protocol.
`thread` is a stable lowercase slug naming a storyline across chapters;
`agents` are roster indices, the basis of the TUI's agent filter.

**The narrator driver** (`internal/mind/narrate.go`, hosted by the
[[agent-mind]] Mind): `chronicleNote` (absorb goroutine, replica already
current) turns notable events into pre-named factual log lines — deaths (and,
since spec 044 US1, the run-over declaration: `run.ended` closes the story
with "The last villager died of `<final_cause>`. The village stands empty —
the run has ended.", the factual line; the [[morgue]] carries the full
record), builds, [[gru]] emergence/sightings/attacks (since spec 044 US3/R5
the `gru.attacked` line renders ONLY for a non-lethal hit — `p.Health > 0` —
because an escalated attack landing at health 0 is a kill whose same-batch
`agent.died` line already carries the death, so "left them wounded" must
stay silent on a killing blow), conversations with gist+topics,
rumors told, gifts, broken promises, chest thefts (spec 013's
`social.chest_taken`, rendered "X took from Y's chest without asking" — the
same narrative weight as a broken promise, a trust violation), musings, and
(spec 041) three [[mental-maps]] beats — a discovery gone stale
(`agent.map_corrected`, "X went looking for the <kind> at (x,y) and found it
gone"), directions changing hands (`social.place_told`, "X told Y about the
<kind> at (x,y)"), and a divine reveal (`metatron.place_revealed`, "A vision
showed X the <kind> at (x,y)") — each voiced by the first fact in the event's
canonical order, and (TASK-13) the whole
[[governance]] arc: assemblies with attendance named (meeting hours rendered
from the TASK-36 convention, including an emergent convention's birth),
grievances raised, proposals tabled/passed/voted down with
tallies, exiles, and witnessed norm violations — each stamped with in-world
time. Since spec 046 ([[curriculum-ladder]]) a `curriculum.stage_unlocked`
event also earns a line — "The village's watcher earned `<stage>`.", the stage
rendered through the skin package's display name (`skin.StageName`) — one of
the ladder's two required in-game unlock surfaces (the other is the CLI status
line). Since spec 054 ([[scenario-machinery]]) a `curriculum.exercise_passed`
event on a scenario world also earns a line — "The watcher's exercise —
`<exercise>` — was passed: the village made it through." `sim.night_started`
closes the day chapter, `sim.day_started` closes
the night chapter; a chapter with no lines spends no call. Since spec 054
([[scenario-machinery]]), a scenario world's `Mind.SetScenario(exercise)`
(installed once at boot, before the loop starts) arms ONE additional chapter
trigger at the exercise's pass/fail boundary — additive to the day/night
cadence: `curriculum.exercise_passed` always closes a
chapter, and `run.ended` closes one only when the mind's scenario id is set
— so a sub-one-game-day scenario run still yields a narrated chapter
carrying the outcome, and an ambient world (no scenario armed) never fires
the extra trigger. Since TASK-32,
`closeChapter` also consults the [[cognition]] router (`routeVerdict` with the
`chronicle` decision class, `llm.KindNarrator`) before enqueueing: the class's
day-scale staleness budget passes at every watchable speed, but a suppression
(possible at future faster speeds) emits a `cog.outcome{suppressed}` record
and drops the chapter — a gap in the story, no call spent. The chapter job
snapshots the lines plus up to 8 recent thread slugs (offered for reuse) to a
single-flight worker: one `llm.KindNarrator` call ([[llm-orchestrator]] cloud
tier, 3-minute cap, MaxTokens 800) asking for strict JSON of 1–3 entries.
`parseNarration` validates (texts trimmed/capped, threads slugified, agent
names resolved against the roster, unknowns dropped) and the batch lands
atomically through `InjectSocial` — `chronicle.entry` is whitelisted in
[[sim-loop]]'s injection door, so narrator output enters the world only as
recorded input, like all model output.

**Morgue epilogues ride the same worker** (spec 044 US2, `narrate.go`): an
absorbed `agent.died` or `run.ended` also queues a `narrJob{epilogue: true}`
via `queueEpilogue` — the SAME single-flight narrator worker and the SAME
router class as chapters (`routeVerdict("chronicle", llm.KindNarrator)`).
The job's lines are a replica-built fact sheet
(`epilogueFacts`: name/cause/day, standing bonds, up to 8 highest-salience
retained memories; `runEpilogueFacts`: every death of the run from the
`run.ended` payload's ledger, `agent` = -1 for the run end) and the worker's
`runEpilogue` makes one `KindNarrator` call under a fixed elegiac
no-invention system prompt, then lands the prose as a recorded
`morgue.epilogue` event through `InjectSocial` — one of the two prose types
an ENDED world's door still accepts (the run-end epilogue lands AFTER
`run.ended`). Failure discipline is the chronicle's own: a
suppressed verdict, a full queue, a transport error, or empty output is a
logged GAP in the morgue's prose, never a stall or retry — the [[morgue]]'s
factual record never waits on it.

**Failure honesty**: a transport/tier failure carries the chapter's lines
into the next boundary via a 1-slot retry buffer (merged oldest-first, capped
at `narrMaxLines` = 120, oldest dropped); unusable model output is dropped —
a gap in the story, never a stall or a retry loop; a full chapter queue (8)
drops with a log line. No llm.json → no narrator; the world just has no
narrated story.

**Views**: the [[tui-client]] chronicle pane renders the replica's ring with
agent/thread filters and a raw-feed fallback; the [[agent-mind]] scribe
renders `chronicle.md` in the save dir — the offline catch-up artifact,
regenerated from recovered state at every daemon start. Since spec 044 the
scribe (`scribe.go`) also renders `morgue.md` ([[morgue]]): `scribe.New`
gained a variadic `EventSource` (the event-log read surface — the daemon
passes the `*store.Store`; call sites passing none render no morgue) and
`renderMorgue` re-runs on every batch carrying `agent.died`, `run.ended`, or
`morgue.epilogue`.

## Connections

[[event-types]] catalogs `chronicle.entry`; [[sim-state-reducer]] holds the
ring; [[sim-loop]] whitelists the injection (narrower on an ended world);
[[llm-orchestrator]] routes `KindNarrator` to the cloud tier; [[tui-client]]
and the scribe render it; [[snapshots]] carry the ring through recovery.
[[mental-maps]], [[morgue]], [[curriculum-ladder]], and [[scenario-machinery]]
each own an event type this note's narrator voices — see How it works above
for which.

## Operational notes

Live-proven (chronicle-proof world, 32x, gemma local + 9router cloud): chapters
landed at both boundaries, the narrator reused thread slugs across chapters
unprompted beyond the offered list, gru night drama narrated from real events,
and the ring survived a daemon restart. Cost: ~2 narrator calls per game day — noise against the $100/month
ceiling. The chapter buffer is in-memory: a daemon restart loses the current
chapter's collected lines (the story resumes at the next boundary).

## Spec 086 — chronicle.entry agents are named refs

`ChronicleEntryPayload.Agents` is `[]AgentRef` on the wire (`{id,name}`
objects; legacy bare-int lists decode dual-shape forever); the reducer arm
folds `refIDs(...)` into the state ring's `ChronicleEntry.Agents []int` —
names never enter state (R2). The narrator (`internal/mind/narrate.go`)
constructs refs via `sim.Refs`; the injection door refuses an unnamed
in-roster ref at live emission.
