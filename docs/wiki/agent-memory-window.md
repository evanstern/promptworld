---
name: agent-memory-window
description: Situated episodic memory: the fixed salience table, the Origin provenance classifier, and the deterministic top-K working-memory window (legacy salience/recency, or the spec-042/043 relevance-blended selector), plus shadow-mode divergence telemetry. Split from [[agent-mind]] (Memories section).
kind: component
sources:
  - internal/sim/memory.go
  - internal/mind/prompt.go
  - internal/mind/telemetry.go
verified_against: d304e8adb64fdf40e24bfeca3ca3420e8a840a35
---

# Agent memory window

**Memories** (`internal/sim/memory.go`): the executor emits `agent.memory_added`
events from a fixed salience table (talk 3★ … death witnessed 10★); the reducer
appends them to `Agent.Memories`. Spec 012's crafting economy added four entries
— `salSpearBroke` (8, the spear that spent its last use), `salOvenBuilt` (7,
village-visible), `salBath` (5, medium and positive), and `salFireOut` (3, a
cold fire nearby is background texture, not formative) — all, like the
pre-existing `SalDream` (8), kept below `GenerationBumpSalience` (9,
[[cognition]]) on purpose: memorable enough to surface in the working window,
never so high they'd interrupt an in-flight generation the way near-death or
exile do. Spec 041 added two more on the low end: `salMapCorrected` (5, a
mental-map correction — [[mental-maps]]) and `salPlaceTold` (3, the talk band,
for giving/getting directions between villagers). Spec 032 added `salAxeBroke`
(8) on the same band, the axe that spent
its last harvest use. Spec 013's storage economy added two more on the same band:
`salChestBuilt` (7, village-visible, the oven precedent) and `salTaking` (7,
a non-owner withdrawal from a chest — suffered by the owner and witnessed by
neighbors, above the rumor-eligibility floor so the owner's subject-tagged
memory seeds gossip). It also added `memoryEventToned`, a `memoryEvent`
variant for a personal (non-gossip, `Subject: -1`) memory that still carries
an explicit tone — `toneBath` (40), `toneOvenBuilt` (30), and spec 013's
`toneChestBuilt` (20) are positive; the taking itself is recorded (since spec 019 through
`situatedMemoryAboutEvent`, see below) with `theftMemoryTone` (−60, negative) for
both the owner and nearby witnesses, alongside a trust/affection hit on the owner→taker
relationship edge — the existing gossip and relation machinery carries a
chest theft the same way it carries any other trust violation ([[social-fabric]]).

Since spec 030, every memory also carries an `Origin` — a closed-vocabulary
provenance class stamped at emission (`OriginAction`/`OriginWitness`/
`OriginReport`/`OriginOmen`/`OriginGist`/`OriginDigest`), a required parameter
on every situated constructor so a new, unstamped emission site cannot compile.
`DirectPerception(origin)` is the pure classifier over it: an own act, a
witnessed event, or a delivered omen/dream are direct perception; a
chest-owner's any-distance report, a conversation gist, a nightly digest, or an
absent/legacy origin (`""`) are secondhand — the conservative default, since
hygiene may under-grant "witnessed" but never over-grant it. This is the ONLY
signal [[nightly-consolidation]]'s belief validator reads to gate a model's
"witnessed" provenance claim — no text inspection, no heuristics.
`SelectMemories` is the deterministic working
window: salience halved per game-day of age, top K−2, plus 2 seeded serendipity
picks from the oldest half (bucketed to `defaultPlannerCadenceTicks` — the
tuning-manifest default constant, deliberately NOT [[world-tuning]]'s tunable
`State.PlannerCadence()` dial, [[memory-retrieval]] covers the distinction),
presented reverse-chronologically. K = `WindowK` (10). Prompts never see the
whole soul.

Since spec 042, which SELECTOR fills that window is gated by the world's
`memory_relevance` flag (`world.json`, validated at `world.Open`, threaded
boot-static into `mind.New` as `memoryRelevance` and reused by `convo.go`'s
scene snapshot): `selectWindow` (prompt.go) sends `""` and `"shadow"` through
the legacy window, byte-identical to today's, while `"on"` selects the
relevance-blended window conditioned on the agent's recorded
situation vector (`Agent.SitVec`, nil until the async embedder has rendered
one — a nil query is the legacy window inside the selector). Since spec 043
both routes run through one shared annotated selector
(`sim.SelectMemoriesWindow` via `selectWindowAnnotated`, each entry flagged
scored-pick vs serendipity-tail for the context assembler's drop accounting;
`StripSelected` of it reproduces `SelectMemories`/`SelectMemoriesRelevant`
byte-for-byte — [[memory-retrieval]]).
In shadow/on mode, `plan()` also records one `cog.memory_divergence` per
enqueued plan job (`recordDivergence`, telemetry.go) — both rankings computed,
only the legacy one served, the evidence a shadow→on gate decision reads.
[[memory-retrieval]] owns the embedding pipeline, the situation vector, and the
relevance-scoring math behind this gate.

Spec 019 (US1) made every sim-emitted episodic memory **situated**: the three
bare constructors are gone — `memoryEvent`/`memoryAboutEvent`/`memoryEventToned`
were replaced (T008b removed the bare forms once every emission site migrated)
by `situatedMemoryEvent`/`situatedMemoryAboutEvent`/`situatedMemoryToned`, so no
memory site can emit unsituated: a site must pick a situated constructor and
therefore a `Where`. Salience/subject/tone semantics are unchanged — this layer
situates, it does not re-weigh. `Where` is a `*sim.MemoryPlace{X,Y,Desc}` baked
at emission by `PlaceAt` (coords always — FR-001; `describePlace`'s deterministic
Manhattan feature scan, radius `placeScanRadius` (2), naming the nearest station
or terrain as a noun phrase — "the fire", "the woods" — or "" when nothing
notable is near); a build completion uses `placeForBuild`/`describePlaceExcept`
to describe its tile WITHOUT the just-placed kind, so "Built a fire" never
resolves to "at the fire" (T024). `Why` is the driving intent's reason verbatim
("" for reflex, and never set on a witness memory — a witness did not drive the
act). `situateText` composes the situated text once, in the grammar order
`<base>[ at <desc> (x,y) | at (x,y)][ — <why>]` (splicing the where-clause before
the base's trailing period), so every call site situates identically and the
scribe never re-derives. The reduced `Memory` gains `Where`/`Why`/`Conv`, all
`omitempty`, copied verbatim by the `agent.memory_added` arm — a pre-019 payload
still produces a pre-019-shaped memory (FR-007/FR-014). Conversation gist
memories are situated differently — `Where` plus a `Conv` transcript ref, text
unchanged ([[social-fabric]]). Villagers also keep a self-authored journal
([[agent-journal]]).

## Connections

[[agent-mind]] is the parent note this child was split from — the mind
driver's prompt assembly is what calls into this window; [[memory-retrieval]]
owns the spec-042 embedding pipeline, the situation vector, and the
relevance-scoring math behind the `memory_relevance` gate; [[decision-context]]
owns the per-turn `memories`/`memories_serendipity` blocks this window feeds
and their drop priority under the context budget; [[world-tuning]] owns the
tunable `PlannerCadence` dial the serendipity bucketing deliberately does NOT
read; [[social-fabric]] carries the trust/relation side-effects a chest theft
memory also produces; [[cognition]] owns `GenerationBumpSalience`, the
ceiling every memory-emission salience is kept under on purpose.

