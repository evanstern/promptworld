---
name: agent-mind
description: The thinking layer's orientation note - the three foundational separations (persona vs soul, events vs files, mind vs loop) plus one-paragraph summaries of the split-out detail; see agent-persona-soul, agent-memory-window, mind-driver-triggers, tool-use-dispatch, villager-tool-handlers, and planner-telemetry for the full accounts.
kind: component
sources:
  - internal/mind/mind.go
  - internal/mind/handlers.go
verified_against: 0af53ec6d211c71e298072c045c67ccbbd13b61d
---

# Agent mind

TASK-7's thinking layer: eight villagers with authored natures, growing souls, and
planner thoughts from the local model — while replay stays byte-deterministic and
model-free. Three separations do all the work: persona vs soul (fixed vs grown),
events vs files (truth vs view), and mind vs loop (I/O vs determinism).

## How it works

**Identity and soul rendering** ([[agent-persona-soul]]): eight write-once
personas seeded at genesis alongside Guardian's editable charter, and the
scribe's regenerable soul.md/journal.md/chronicle.md/village_charter.md/
morgue.md views rendered from the event log — souls survive restarts because
the log is the only truth.

**The memory window** ([[agent-memory-window]]): every sim-emitted memory is
situated (Where/Why/Conv, an Origin provenance class), scored from a fixed
salience table, and served through a deterministic top-K working-memory
window — legacy salience/recency, or since spec 042/043 a relevance-blended
selector gated by the world's `memory_relevance` flag, with shadow-mode
divergence telemetry recorded per enqueued plan job.

**Driver cadence and prompt content** ([[mind-driver-triggers]]): per-agent
planner cadence with a phase-preserving stagger, the trigger set that arms a
due thought (wake, completion, nightfall, encounters, mental-map corrections,
on-scene harvest acts re-arming intent-matched witnesses (spec 081), paused
Guardian nudges), and the social-law/known-places/village-law blocks a
villager's own history and mental map render into its own prompt. Since
spec 104, `absorb` also walks the replica's own derived progress
(`AdvanceTo`) after each batch and, on a coalescing-regime world, sweeps for
newly-adjacent pairs itself — the per-event `agent.moved` encounter trigger
never fires for a coalesced walk.

**The cognition gate and tool-use loop dispatch** ([[tool-use-dispatch]]): the
cognition-horizon gate that suppresses a thought whose predicted drift would
already be stale on arrival, the immutable per-job snapshot and `thoughtMeta`
identity, and how `runPlan` drives the bounded tool-use loop against the
villager roster — the retired free-text reply contract now lives as tool
schemas the loop driver validates before dispatch.

**Villager tool handlers** ([[villager-tool-handlers]]): every acting tool
(world verbs, `set_plan`, `muse`, the journal writes) wraps an existing
landing door instead of mutating the world directly; musing itself, retired
as a scheduled channel by spec 017, is now just one more roster tool
competing for the same turn as any other action.

**Planner outcome telemetry** ([[planner-telemetry]]): every buffered tool
call lands as a `cog.tool_call` event on every termination path, single goals
still land through `Loop.InjectIntent`'s existing ladder, and `runPlan`'s
terminal switch mirrors the pre-loop outcome paths (landed, rearmed on a bare
rejection, or an explicit `unusable` when no door was ever reached) — no
cognition thread ends silently.

## Connections

[[mental-maps]] is the spec 041 per-agent knowledge subsystem the prompt's
known-places section renders from and whose corrections re-arm the planner;
[[memory-retrieval]] is the spec 042 subsystem behind the `memory_relevance`
mode gate `selectWindow` reads and the `cog.memory_divergence` telemetry
`plan()` records;
[[executor]] emits memories and runs the intents; [[reflex-policy]] shares
`resolveGoal` and provides the fallback; [[cognition]] owns the decision-class
registry, the router the mind gates on, and the latency estimate behind
predictions and future-dating — since spec 024 the mind reads it through the
orchestrator's `EstimateForKind` seam (the kind's admissible chain-head
provider's estimate; `RecalibrateSignal` is per-provider, the breaching name
riding the recorded payload's `Tier` field, kept for replay-schema stability);
[[llm-orchestrator]] carries the calls
(routed by the kind's provider chain in llm.json); [[tool-loop]] is `runPlan`'s
driver (spec 017) — `md.runLoop`
wraps `toolloop.Run`, `tool.LoopRosterVillager()` ([[tool-registry]]) is its
declared roster (since spec 019 including the four journal tools), and
`villagerHandlers` (handlers.go) wraps every acting tool's landing door;
[[agent-journal]] is the spec 019 self-authored notebook the journal handlers
read and write; [[sim-loop]]'s `inject_intent` command is the only door into
deterministic space and since TASK-32 the owner of landing-time validation
(staleness ladder, generation and guard checks); [[event-types]] catalogs the
new events; the [[tui-client]]
souls pane shows each agent's newest memory. [[nightly-consolidation]] digests each
day's memories into the soul at sleep; TASK-8 turned the talk primitive into real
conversations. The mind also hosts the [[chronicle]] narrator (TASK-11): absorb
collects notable events as named log lines and day/night boundaries hand chapters
to a single-flight cloud worker — since spec 044 an absorbed
`agent.died`/`run.ended` also queues a [[morgue]] epilogue job on that same
worker (`queueEpilogue`, mind.go/narrate.go) — and the [[governance]] phrasing driver (TASK-13,
`meeting.go`): enacted proposals get one best-effort `llm.KindMeeting` call
rephrasing the template text in the proposer's voice, injected as
`meeting.proposal_rephrased`; every failure leaves the template standing.

## Operational notes

Live-verified against real Ollama: personas visibly steer reasoning (Hazel: "will
charm my way into doing it"), souls accrete and survive restarts, persona hashes
stay intact. Known gap: at `max` speed the mind replica can drop event batches
(overflow policy) — resync-on-overflow is future work; ≤16x is drop-free. Planner
volume at 4x ≈ 16 calls/game-hour for 8 agents, all on zero-priced local
providers.
