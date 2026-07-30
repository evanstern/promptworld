---
name: context-block-inventory
description: The full block table (frame/needs/self_history/inventory/plan_echo/directive/known_places/social_law/memories/memories_serendipity/journal — eleven since spec 084) a villager planner prompt is assembled from - source of truth, appearance condition, cap, and drop priority per block - plus the as-implemented rendering rules. Split from [[decision-context]] (The blocks + Rendering rules sections).
kind: concept
sources:
  - internal/mind/context.go
  - internal/mind/prompt.go
  - internal/sim/agents.go
  - internal/sim/plan.go
  - internal/sim/guard.go
  - internal/sim/memory.go
  - internal/sim/journal.go
verified_against: 9b4ed5aef5bfea50b67fac10f8e2153f065a814d
---

# Decision-context block inventory

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
| 5b | `directive` | the guardian's hard command(s) — ≤2 ACTIVE directives addressing this agent, oldest first: framing text verbatim, the bound designation's kind/site/fulfillment requirement, plain-words days left | `renderDirective`/`directiveGoalPhrase`/`directiveTimeLeft`; `State.Directives`, `State.Designations` (spec 084) | no active directive addressing the agent (or bound designation non-active — the orphan the sweep expires) → omitted entirely; a directive-free prompt is byte-identical to pre-084 | `directiveRenderCap` = 2; text ≤400 runes at the door | never (a hard command is never shed) |
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

## Connections

[[decision-context]] is the parent note this child was split from — it owns
the scope, size budget, drop order, observability fields, and deliberate
absences this block table feeds; [[agent-mind]] hosts `plan()`, which calls
the assembler this table describes, once per enqueued planner job;
[[memory-retrieval]] owns the `memory_relevance` mode gate behind the
`memories` block (block 8); [[mental-maps]] owns the per-agent map state
`known_places`/`nearby` (block 6) renders from; [[social-fabric]] /
[[governance]] own the state `social_law` (block 7) renders from;
[[agent-journal]] owns the journal state and `SelectJournalExcerpts`, the
selector the `journal` block (block 10) renders.

