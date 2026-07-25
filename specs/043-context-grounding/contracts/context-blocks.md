# Contract: Decision-Context Blocks

**Feature**: specs/043-context-grounding | **Consumers**: `internal/mind` assembler,
`docs/wiki/decision-context.md` (living projection), TASK-106 detector (taxonomy reuse)

The villager decision prompt is assembled from named blocks in this fixed order. Every
block is deterministic given world state. A block that has nothing to say renders as
empty string (omitted entirely — no empty headers). Sizes are measured per block;
budget overflow drops whole blocks in ascending drop-priority order.

| # | Block name | Content | Empty state | Drop priority |
|---|---|---|---|---|
| 1 | `frame` | time, phase, position, future-dating line | never empty | never dropped |
| 2 | `needs` | five needs, each with trajectory arrow (rising/falling/steady) | never empty | never dropped |
| 3 | `self_history` | last ≤4 IntentLog records, newest first: goal, source in plain words ("you chose", "instinct sent you", "your plan's step"), outcome; explicit "you have not acted yet" on empty | first thought → "no prior activity" line | never dropped |
| 4 | `inventory` | carried items | never empty (renders zero-state) | never dropped |
| 5 | `plan_echo` | active plan: remaining steps in order, head marked next, guards + deadlines in plain words | no active plan → omitted | 6 |
| 6 | `known_places` / `nearby` | existing mental-map blocks (spec 041), unchanged content | as today | 5 |
| 7 | `social_law` | existing social/law context, unchanged content | as today | 4 |
| 8 | `memories` | relevance-blended window (spec 042) — protected floor of 4; entries above the floor drop first | as today | 3 (above floor), floor never dropped |
| 9 | `memories_serendipity` | the 2 serendipity tail picks | absent when window ≤ k-2 | 2 |
| 10 | `journal` | ≤2 term-matched entries, each ≤300-rune excerpt, with entry ids | no match → omitted | 1 (first dropped) |

## Rules

- **Order is normative.** Blocks render in table order; renumbering is a contract
  change and must update the wiki note.
- **Sources named honestly.** `self_history` renders Source verbatim-mapped:
  planner → "you decided", reflex → "instinct", plan → "your plan". Reflex records
  have no Reason and none is invented (edge case: instinct honesty).
- **Overrides visible.** Two records landing in quick succession (instinct override)
  appear as consecutive entries in landing order; the open-then-superseded shape is
  preserved (edge case: rapid succession).
- **Budget.** `contextBudgetTokens` (approx-tokens = bytes/4) default 2000; dial is
  tuning-manifest-ready (TASK-107 const-fallback pattern). Drops are whole-block,
  ascending drop priority; drops recorded in `cog.thought.DroppedBlocks`.
- **Determinism.** Identical world state ⇒ identical bytes. No wall clock, no RNG
  outside the existing seeded serendipity picks.
- **Degraded modes.** No `SitVec`/embedder → `memories` falls back to legacy selection
  (unchanged block name); journal term match is model-free and always available.
- **Observability.** `cog.thought` carries `PromptBytes`, `BlockBytes` (by these
  names), `DroppedBlocks`.
