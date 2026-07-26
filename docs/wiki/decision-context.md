---
name: decision-context
description: The per-turn decision-context inventory (spec 043) — every named block a villager's planner prompt is assembled from, its source of truth, appearance/empty-state conditions, size caps, and drop priority under the context budget, plus the deliberate absences
kind: concept
sources:
  - internal/mind/context.go
  - internal/mind/prompt.go
  - internal/mind/mind.go
  - internal/mind/telemetry.go
  - internal/sim/cognition.go
  - internal/sim/agents.go
  - internal/sim/state.go
  - internal/sim/journal.go
  - internal/sim/memory.go
  - internal/sim/plan.go
  - internal/sim/guard.go
verified_against: d304e8adb64fdf40e24bfeca3ca3420e8a840a35
---

# Decision context (per-turn context grounding)

Spec 043 (TASK-105, from the TASK-101 spike) is the durable answer to "what did this
villager know when it decided that?" Every villager planner prompt is assembled from
named **blocks** in a fixed contract order (`contracts/context-blocks.md`,
`specs/043-context-grounding/`), each a pure function of world state, each measured,
with a documented drop order for when the assembled context would exceed the
per-thought size budget. This note is the living projection of that contract against
the actual assembler (`internal/mind/context.go`); the two must move together.

**Scope**: this inventory covers the **planner** decision prompt only —
`assembleContext`/`assembleBudget`, called once per enqueued planner job
(`internal/mind/mind.go`'s `plan()`); `AssembleUserPrompt`
(`internal/mind/prompt.go`) exports the same assembly minus the per-thought
future-dating line, the pure-function entry point replay tooling and the TUI
capture path use to reproduce a thought's exact bytes from a bare `*sim.State`.
Conversation-scene and meeting prompts are a
different, unbudgeted surface and are out of scope here. `cog.thought`'s
`PromptBytes`/`BlockBytes`/`DroppedBlocks` are stamped ONLY for `Class == "planner"`
thoughts — every other class's `cog.thought` carries them zero-valued
(`omitempty`, `internal/sim/cognition.go`).

## The blocks (contract order)

Rendered in this exact order; a block with nothing to say returns `""` (omitted
entirely — no empty header ever renders). Priority is the drop rank under budget
pressure: **higher number = dropped later**; `neverDrop` blocks are never shed.

| # | Block | Content | Source of truth | Empty / appearance condition | Cap | Drop priority |
|---|---|---|---|---|---|---|
| 1 | `frame` | future-dating line (when the decision won't land instantly) + game time, day/night phase, position | `renderFrame`, `clock.Format`, `Agent.X/Y` | never empty; future-dating line itself is blank at uncapped/paused speeds (`futureDated`) | — | never |
| 2 | `needs` | five needs (0-100 scale) each with a rising/falling/steady trajectory arrow | `renderNeeds` + `trajectory`; `Agent.Needs`, `Agent.NeedsAnchor`/`NeedsAnchorTick` | never empty; before any anchor exists every need reads steady (edge case: first thought) | — | never |
| 3 | `self_history` | last ≤4 `IntentRecord`s, newest first: goal, source in plain words, stated reason (only when one was recorded), outcome | `renderSelfHistory`/`selfHistoryLine`; `Agent.IntentLog` (ring, cap 8) | first thought (empty ring) → explicit "no prior activity yet — this is your first decision" line, never silence | show ≤4 of ring cap 8 | never |
| 4 | `inventory` | carried resources/items (wood/stone/water/planks/refined stone/food/meals, spears) | `renderInventory`; `Agent.Inv` | never empty (zero-state renders as zero counts) | — | never |
| 5 | `plan_echo` | active plan's remaining steps in order, head marked "next" the rest "then", each with its guard and validity deadline in plain words | `renderPlanEcho`/`guardPhrase`; `Agent.Plan []PlanStep` | no active plan → omitted entirely (no stale echo); plan end (completed/expired/guard-failed/superseded) surfaces instead via `self_history` at the next thought | `PlanStepCap` = 3 steps | 6 |
| 6 | `known_places` / `nearby` | spec-041 known-places section (landmarks with provenance — since spec 044 US4 including graves — place-shaped groups, orientation) + a peer-sighting "Nearby" line | `renderKnownPlaces`/`knownPlaces`; `Agent.Map` | as today (unchanged content) | peer scan radius 10 tiles | 5 |
| 7 | `social_law` | bonds/debts/reputation/rumor + village-law context (active norms, exile judgments) | `renderSocialLaw` = `socialContext` + `villageLaw` | as today (unchanged content) | — | 4 |
| 8 | `memories` | working-memory window: relevance-blended when `memory_relevance` is `"on"` with a recorded situation vector, legacy salience/recency otherwise; a protected floor of the most-recent 4 non-serendipity entries | `buildMemLines`/`renderMemLines`; `Agent.Memories`, `Agent.SitVec`, `sim.SelectMemories`/`SelectMemoriesRelevant` | no memories → no header at all (not an empty list) | window `WindowK` (10) minus 2 serendipity; floor 4 (`memoryFloor`) | 3 above the floor; the floor itself never drops |
| 9 | `memories_serendipity` | the 2 serendipity tail picks from the oldest half of memory, seeded per `defaultPlannerCadenceTicks` bucket (the tuning-manifest default constant, deliberately not the [[world-tuning]]-tunable cadence dial) | same window (`buildMemLines`), tagged `serendipity` | absent when the window has ≤ K−2 scored entries (no tail to pick) | 2 entries | 2 |
| 10 | `journal` | ≤2 term-matched excerpts from the villager's own journal, each ≤300 runes, with entry ids | `renderJournal`; `Agent.Journal.SelectJournalExcerpts(situationTerms)` | no term match → omitted entirely | `JournalExcerptCap` = 2, `JournalExcerptRunes` = 300 | 1 (first dropped) |

A fixed, unmeasured closer (`"\nWhat do you do next?"`) always follows the last
rendered block; it is never counted toward the budget and never dropped.

## Rendering rules (as implemented)

- **Sources named honestly** (`self_history`, block 3): `IntentRecord.Source` maps
  `"planner"` → "you decided this", `"reflex"` → "instinct drove this", `"plan"` →
  "your plan's step"; anything else → "source unknown". A stated `Reason` renders
  only when the record actually carries one (planner/plan intents) — reflex records
  carry none and none is ever invented.
- **Overrides visible**: an intent that landed while an earlier one was still open
  leaves the earlier record open (not retroactively closed), so two records landing
  in quick succession appear as consecutive entries — open-then-superseded, in
  landing order.
- **Trajectory** (`needs`, block 2): `trajectory(current, anchor, hasAnchor)` compares
  the current need value to the window-edge anchor with a `trajectoryDeadband` of
  ±10 (of the raw 0-1000 scale, one point on the displayed 0-100 scale); no anchor
  yet (first `trajectoryWindowTicks` = 1800 ticks of an agent's life) always reads
  steady rather than a spurious rising/falling.
- **Plan echo** (block 5): guards render in plain second-person words via
  `guardPhrase` (`GuardTargetAlive`/`GuardTargetPresent`/`GuardNotSuperseded`/
  `GuardAfterTick`/`GuardBeforeTick`), never the raw predicate name; `Until` renders
  as "valid until `<clock>`" when set.
- **Memory chunk accounting** (blocks 8-9): the window is ONE contiguous rendered
  region so nothing is reordered by the floor/serendipity split — with nothing
  dropped, the concatenation is byte-identical to the pre-043 single "You remember:"
  list. Under pressure the serendipity tail sheds first, then above-floor entries;
  the floor's `memHeader` ("\nYou remember:\n") rides with whichever tier is kept,
  and `memAccount` attributes it to the `memories` block name so every rendered byte
  is claimed by exactly one block in telemetry.
- **Journal terms** (block 10): `situationTerms` is a pure function of agent state —
  the two lowest-valued needs' names plus the active-or-last intent goal — so
  selection is deterministic and model-free; it is available in every degraded mode.

## Budget and drop order

`contextBudgetTokens` = 2000 approx-tokens (`bytes/4` — no tokenizer in production;
a package const today, designed as a TASK-107 tuning-manifest dial with this value as
the const-fallback). While the assembled total exceeds the budget, whole blocks are
shed **lowest-priority-first**, recorded in `cog.thought.DroppedBlocks` in the order
they were dropped:

```
journal (1) → memories_serendipity (2) → memories above-floor (3) →
social_law (4) → known_places (5) → plan_echo (6) → [never dropped]
frame, needs, self_history, inventory, and the memories floor (4 entries)
```

If only protected (never-drop) content remains and the budget is still exceeded, the
loop stops rather than shedding survival-relevant content — the contract protects
`frame`/`needs`/`self_history`/`inventory`/the memory floor absolutely; an overflow
in that state is a budget that cannot be met, not a bug to paper over.

## Observability

Every planner `cog.thought` event carries:

- `PromptBytes` (int) — total assembled user-prompt bytes.
- `BlockBytes` (map[string]int) — bytes actually rendered per kept block, keyed by
  the block names in the table above.
- `DroppedBlocks` ([]string) — the blocks the budget shed, in drop order; empty when
  nothing was dropped.

All three are `omitempty` and additive-last on `CogThoughtPayload`: a pre-043 event
log, and every non-planner `cog.thought` today, decodes with them zero-valued — the
reducer stays a no-op for `cog.*`, so replay is unaffected by this feature entirely.

## Deliberate absences

The prompt does NOT include, by design:

- **The full event log.** A villager sees only its own reducer-derived state
  (`IntentLog`, `NeedsAnchor`, `Plan`, `Memories`, `Journal`, `Map`) — never the raw
  event stream, never other agents' event history.
- **`IntentLog` beyond the ring/the shown window.** The ring itself is capped at 8
  records (`intentLogCap`); `self_history` renders only the newest 4 of those. Older
  intent history is not reconstructable from the prompt — it lives only in the event
  log, for operator/replay inspection outside the model's view.
- **Other agents' private state.** No other agent's needs, inventory, plan, journal,
  memories, or intent history appear in a villager's own prompt. What a villager
  knows about a peer comes only through its own mental map (`known_places`/`nearby`,
  block 6) and the social-fabric surfaces (`social_law`, block 7) — never a direct
  read of another agent's `Agent` struct.
- **Other agents' journals.** A villager's `read_journal`/`search_journal` tools
  (spec 019) read only its OWN journal snapshot; the assembler's automatic
  `journal` block (block 10) is likewise self-only.
- **Exact numeric trajectories.** `needs` (block 2) renders a direction word
  (rising/falling/steady), never the anchor value, the raw delta, or the window
  length in ticks — the model gets a qualitative signal, not a number to
  over-index on.
- **The full memory store or full journal.** Both are windowed/excerpted
  deterministically (blocks 8-10); nothing outside the rendered window/excerpts is
  visible to a given thought, by the same doctrine that has governed the memory
  window since [[agent-mind]]'s `SelectMemories`.
- **Conversation-scene and meeting prompts.** This inventory is the planner surface
  only (see Scope, above); those prompts are built and budgeted separately and are
  not audited here.
- **A raw predicate/guard name.** Plan guards (block 5) always render through
  `guardPhrase`'s plain-words mapping — the closed guard vocabulary itself
  (`GuardTargetAlive` etc.) never appears verbatim in a prompt.

## Connections

- `contracts/context-blocks.md` (`specs/043-context-grounding/`) is the normative
  contract this note projects; block order, names, and drop priorities are
  contract-owned — a renumbering there is a contract change and must update this
  note in the same commit.
- [[agent-mind]] — the mind driver (`plan()`) that calls the assembler once per
  enqueued planner job, and the memory-window/tool-loop machinery the assembled
  prompt feeds into.
- [[memory-retrieval]] — owns the `memory_relevance` mode gate, the situation
  vector, and the relevance-scoring math behind the `memories` block (block 8) when
  the mode is `"shadow"`/`"on"`.
- [[mental-maps]] — owns the per-agent map state `known_places`/`nearby` (block 6)
  renders from.
- [[social-fabric]] / [[governance]] — own the state `social_law` (block 7) renders
  from (bonds/debts/reputation/rumor; active norms/exile judgments).
- [[agent-journal]] — owns the journal state and the deterministic term-match
  selector (`SelectJournalExcerpts`) the `journal` block (block 10) renders.
- [[cognition]] — `cog.thought`'s existing decision-trace surface, extended (not
  replaced) by this feature's `PromptBytes`/`BlockBytes`/`DroppedBlocks` fields.
- [[event-types]] — catalogs `CogThoughtPayload`'s fields including this feature's
  additive-last, `omitempty` extension.
- Upstream design record: `specs/043-context-grounding/` (spec.md, data-model.md,
  `contracts/context-blocks.md`, quickstart.md), and the TASK-101 spike this feature
  answers.

## Operational notes

- Live-verified (SC-001) against a scratch world's `cog.thought` events: see
  `specs/043-context-grounding/evidence/sc-001-capture.md` for the captured
  `BlockBytes`/`DroppedBlocks`/`PromptBytes` and the block-for-block check against
  this table.
- This note is pinned to the commit above, re-pinned post-merge (T026) alongside
  the other touched notes ([[agent-mind]], [[memory-retrieval]],
  [[event-types]], [[sim-state-reducer]]).
