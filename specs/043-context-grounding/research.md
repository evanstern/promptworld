# Research: Per-Turn Context Grounding

**Feature**: specs/043-context-grounding | **Date**: 2026-07-24
**Code surfaces verified at**: main (94800d9)

No NEEDS CLARIFICATION markers existed in the Technical Context; the research below
resolves the design unknowns the spec's Assumptions deferred, each grounded in the
current code.

## R1 — Self-history state: reducer-maintained intent ring

**Decision**: add `Agent.IntentLog []IntentRecord` (ring, cap 8) maintained entirely by
reducer arms. `IntentRecord{Goal, Source, Reason, Tick, Outcome, OutcomeTick}`.
`agent.intent_set` appends (source already on the payload: `IntentSetPayload.Source`
"reflex"|"planner"|"plan", agents.go:807, plan.go:95); `agent.intent_done`,
`agent.build_failed`, `agent.intent_rejected`, `agent.plan_expired` stamp the outcome
on the matching open record.

**Rationale**: today only `LastGoal`/`LastGoalTick` exist (agents.go:123-124, set at
state.go:604-605) — single slot, no source, no outcome, TUI-only. A reducer-derived
ring is deterministic and replay-safe by construction (state is rebuilt from events),
costs O(8) memory per agent, and gives the prompt everything US1 needs without reading
the event store on the mind path.

**Alternatives considered**: (a) scanning the event stream at prompt time — non-local,
couples the mind to store access, and violates the "assembly from world state alone"
determinism posture; (b) extending `LastGoal` in place — loses the alternation
visibility FR-003 requires (a single slot cannot show A→B→A).

## R2 — Need trajectories: windowed anchor snapshot, not a history ring

**Decision**: add `Agent.NeedsAnchor Needs` + `NeedsAnchorTick int64`, refreshed inside
the existing `agent.needs_changed` reducer arm (state.go:1411) whenever
`tick - NeedsAnchorTick >= trajectoryWindowTicks` (default 1800 — one planner cadence).
Direction per need = sign(current − anchor) with a ±10 (of 1000) deadband → rendered as
rising / falling / steady.

**Rationale**: `agent.needs_changed` already fires every game-minute
(executor.go:104-113) but the reducer only overwrites current values — history exists
solely in the event log. A two-value anchor gives direction with O(1) state and zero
new events; the deadband satisfies the spec's no-flicker scenario (US2 AS-3).

**Alternatives considered**: (a) full ring of samples — more state for no additional
prompt-visible signal (the prompt renders direction, not curves); (b) EMA slope — less
explainable to operators and needs tuning; (c) deriving from the event log at prompt
time — same determinism/local-access objection as R1(a).

## R3 — Plan echo: render existing plan state, no state change

**Decision**: render `Agent.Plan` (agents.go:103, `PlanStep` plan.go:23-35) directly in
the prompt: remaining steps in order, head marked "next", each step's guard (`When`)
and deadline (`Until`) in plain words. Plan end reaches self-history via the R1 ring
(`agent.plan_expired` stamps the outcome; `plan_step_started` needs no record — the
fired step lands as its own `intent_set` with source "plan").

**Rationale**: everything US3 needs already exists on the agent; `userPrompt` simply
never reads it (verified: prompt.go has zero references to Plan/Intent/LastGoal). Pure
rendering work.

**Alternatives considered**: adding a plan-progress field — unnecessary; head-of-slice
already encodes progress (`agent.plan_step_started` pops, state.go:547-561).

## R4 — Memory relevance: consume spec 042 as-is

**Decision**: no changes to selection. `SelectMemoriesRelevant` (memory.go:422) with
the agent's rolling situation vector (`Agent.SitVec`, refreshed per cadence bucket by
the embedder driver, embedder.go:212-228) is the relevance mechanism; nil/absent vector
already degrades to legacy selection (memory.go:423-425), and vectorless memories score
neutral 0.5 (memory.go:500-514). This feature's only obligations: the assembler treats
the returned window as one budgeted block, and the quickstart validates the degraded
path end-to-end (FR-006, US4 AS-4).

**Rationale**: TASK-98/spec 042 merged exactly this capability, including divergence
telemetry; re-touching it would duplicate a fresh, tested subsystem.

## R5 — Journal inclusion: deterministic term-match excerpts (embeddings deferred)

**Decision**: assembler-side selection using the existing deterministic
`Journal.SearchJournal` (substring, most-recent-first, cap 8 — journal.go:131) with
query terms derived from the current situation (the two worst needs' names and the
active/last intent goal — the same signals `renderSituation` uses, embedder.go:257).
Include at most 2 entries, each excerpted to ≤300 runes, as the lowest-priority block.
Embedding journal entries at write (the spec-042 pattern applied to
`journal.entry_written`) is explicitly deferred — noted as a follow-on, not built here.

**Rationale**: journal entries carry no vectors today and no ranking layer exists; a
deterministic term match delivers FR-007 (relevant, bounded, no reasoning turns spent)
without expanding the embedding pipeline mid-feature. The villager's own journal read
tools stay unchanged and complementary.

**Alternatives considered**: (a) embed-at-write — right long-term shape, but couples
this feature to new embedder traffic and event types; deferred; (b) always include the
latest entry — fails "chosen for the moment" (US4).

## R6 — Budget accounting: bytes/4 approx-tokens as a shipped helper

**Decision**: promote the test-only heuristic (`tokensApprox = bytes/4`,
prompt_test.go:163-166) into a small shipped helper used by the assembler; budget
default `contextBudgetTokens = 2000` approx-tokens as a package const with the
TASK-108-style comment marking it a tuning-manifest dial (TASK-107 fallback pattern).

**Rationale**: no tokenizer exists in production (verified — only the static Fibonacci
points model, registry.go:37-42, and output-side MaxTokens budgets, config.go:51);
bytes/4 is stable across the local tiers in use and the budget is a guardrail, not a
billing meter. Exact tokenization would add a dependency for no behavioral gain.

## R7 — Assembly & drop order: block assembler in internal/mind/context.go

**Decision**: a block registry where each context block declares a name, priority, and
renderer; assembly renders in fixed order, measures each block (bytes + approx-tokens),
and on budget overflow drops whole blocks lowest-priority-first. Drop order (first
dropped → last): journal excerpts → serendipity memory tail → retrieved memories above
a floor of 4 → social/law detail → plan echo → trajectories → self-history → needs/
inventory/position (never dropped). Every drop emits into the per-thought telemetry
(R8). Deterministic: same state → same blocks → same drops.

**Rationale**: FR-008 requires a documented order protecting survival-relevant blocks;
whole-block dropping keeps rendering deterministic and the contract explainable
(contracts/context-blocks.md is the normative list).

## R8 — Observability: sizes on cog.thought

**Decision**: extend `CogThoughtPayload` (cognition.go:52-62) with `PromptBytes int`
and `BlockBytes map[string]int` (block name → bytes; dropped blocks recorded with
negative size or a parallel `Dropped []string` — final shape in data-model.md). Emitted
where thoughts are already stamped (telemetry.go:138-146). Reducer stays no-op
(state.go:526-529), so replay is unaffected; old logs remain readable (missing fields
zero-value).

**Rationale**: FR-009/FR-010 want per-thought context observability through existing
decision-trace surfaces; `cog.thought` is exactly that surface and today records no
prompt sizing at all.

## R9 — Context inventory home: docs/wiki/decision-context.md

**Decision**: the US5 inventory is a grounding-wiki note (`docs/wiki/
decision-context.md`) listing every block (source of truth, appearance conditions,
caps, drop priority) plus the deliberate absences; pinned to the verified commit like
every note, so the existing freshness gate (Principle IV) flags drift — satisfying US5
AS-2 with zero new tooling. `contracts/context-blocks.md` in the spec dir is the
design-time contract; the wiki note is the living, re-pinned projection.

**Rationale**: the wiki already has the pin/freshness machinery and is the corpus
operators read; inventing a second audit surface would rot.
