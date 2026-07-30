---
name: decision-context
description: The per-turn decision-context inventory (spec 043) - scope, budget, drop order, observability fields (PromptBytes/BlockBytes/DroppedBlocks), and the deliberate absences for the villager planner prompt; see context-block-inventory for the full block table (eleven blocks since spec 084's neverDrop directive block) and per-block rendering rules.
kind: concept
sources:
  - internal/mind/context.go
  - internal/mind/prompt.go
  - internal/mind/mind.go
  - internal/mind/telemetry.go
  - internal/sim/cognition.go
  - internal/sim/state.go
  - internal/sim/agents.go
  - internal/sim/memory.go
  - internal/sim/journal.go
verified_against: 1fae0d8536eb43e43eaa7b747aaeaf0b6e05ac83
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

## The block inventory ([[context-block-inventory]])

The full contract table (eleven blocks since spec 084) — frame, needs, self_history, inventory,
plan_echo, known_places/nearby, social_law, memories, memories_serendipity,
and journal, each with its source of truth, appearance/empty-state
condition, size cap, and drop priority — plus the as-implemented rendering
rules (source naming, override visibility, trajectory bucketing, plan-echo
guard phrasing, memory-chunk accounting, journal-term selection) now live in
[[context-block-inventory]]. This note keeps the scope, the budget and drop
order, the observability contract, and the deliberate absences.

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
  the block names in [[context-block-inventory]]'s table.
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
  [[context-block-inventory]]'s table.
- This note is pinned to the commit above, re-pinned post-merge (T026) alongside
  the other touched notes ([[agent-mind]], [[memory-retrieval]],
  [[event-types]], [[sim-state-reducer]]).
