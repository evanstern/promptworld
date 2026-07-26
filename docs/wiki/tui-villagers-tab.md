---
name: tui-villagers-tab
description: The villagers tab: the roster (per-agent status/needs/inventory), the detail view, and the decisions sub-view's client-side decision-trace projection (thought/tool_call/outcome chains joined on job id, capped and evicted). Split from [[tui-client]]; read when touching decisions.go or views.go's villager rendering.
kind: component
sources:
  - internal/tui/views.go
  - internal/tui/decisions.go
  - internal/tui/tui.go
verified_against: d304e8adb64fdf40e24bfeca3ca3420e8a840a35
---

# TUI villagers tab

Split from [[tui-client]] (corpus-spec v2 size-budget split, summary-style):
this note covers the villagers tab's roster, detail view, and decisions
sub-view. See [[tui-client]] for the dock's other tabs ([[tui-dock-tabs]]),
the map view, and the chronicle feed.

## Roster, detail, and decisions

The **villagers** tab (renamed from
"souls", spec 015/TASK-56 — now a two-view inspector rather than a flat
roster). The villagers **roster** shows per agent: a selection cursor,
status, current goal, needs gauges, a leading `bulk n/24` derived-load
reading (spec 013 T015, SC-006; `sim.Bulk`/`sim.BulkCap` — the same function
the reducer/executor clamp gathers and crafts against, so the number never
drifts from what an action will actually do), then the full carried-inventory
line — wood/stone/water/planks/refined-stone counts, the food triplet
raw/cooked/meals, and (when carried) a spear count with the most-worn spear's
remaining uses. While the villagers tab is visible, `j`/`k`/`g`/`G` move the
cursor and `⏎` opens the selected villager's **detail view**
(`villagerDetailBody`): identity/vitals, an objective line (active
`Intent.Goal` marked current; else the reducer-stamped `Agent.LastGoal` +
tick marked `last:`; else "no objective yet" — [[sim-state-reducer]]),
itemized inventory, beliefs/narrative when consolidation has produced them,
and episodic memories most-recent-first, each section truncating bottom-up
inside the pane budget. From the detail view, `d` opens the **decisions
sub-view** (spec 020/TASK-63, `villagerDecisionsBody`): the villager's recent
cognitions as causal chains, most-recent-first — a when/class header, the
stimulus line, each tool call as `ordinal. tool — phrase (reason)`, and the
terminal outcome or an explicit `in progress — no outcome yet` marker; router
suppressions render as one `didn't think because …` entry. Chains come from
the client-side **decision-trace projection** (decisions.go): `applyEvent`
feeds every `cog.thought`/`cog.tool_call`/`cog.outcome` into `Model.traces`
before the ring append, joining on the shared job ID, so the stimulus is
resolved once at thought-ingest from the chronicle ring in the digest voice
(a pre-connect trigger degrades to a neutral `stimulus #N` reference,
trigger 0 to a cadence phrase) and the stored chain survives the ring's
500-event eviction. Attribution: the thought/outcome payload's agent, else a
villager job-ID parse for fragments; `turn-metatron-*` jobs go to a sentinel
and `conversation-*` jobs are never ingested. `ingestOutcome` also skips the
NON-terminal `sim.OutcomeRetried` marker (spec 025, TASK-72): the tool-loop
consumers emit it AFTER a landed run's door already recorded the real terminal
outcome, so folding it in would overwrite `landed` with `retried` — the marker
stays in the event log for trail-level retry counting, it just never becomes a
chain's outcome (the same disregard conversation outcomes get via the job-ID
prefix guard). The projection is bounded
(`decisionChainCap` 20 chains per agent, oldest evicted) and resets wholesale
on reconnect like the replica. Verdicts and outcomes render ONLY through the
sweep-tested plain-language `verdictGlossary` — raw enum strings never reach
the screen (an unknown value gets a safe generic phrase). `j`/`k` scroll the
sub-view (render-time clamped), and `esc` unwinds decisions → detail →
roster ahead of the solo-release chain; selection state survives tab
switches and is clamped on reconnect. Full soul.md persona files stay on
disk per [[agent-mind]]. The same glossary feeds the guardian's inline verdict
rows: a `turn-metatron-*` `cog.tool_call` appends one `» tool — phrase`
transcript row at ingest (`guardianVerdictRow`), which
`classifyTranscriptLine` labels `note` and styles as cog telemetry — the
angel's refused and landed calls are visible in the transcript where before
only the RPC reply's `⚡` miracle lines appeared.

## Back to parent

[[tui-client]] links here for the villagers tab; that note's own Connections
section lists [[sim-state-reducer]] and [[agent-mind]] as this tab's
underlying data sources.
