# Feature Specification: Per-Turn Context Grounding

**Feature Branch**: `043-context-grounding`

**Created**: 2026-07-24

**Status**: Draft

**Input**: User description: "Per-turn context grounding for villager minds — audit and intent-driven context assembly (board TASK-105, from spike TASK-101). Today each villager thought gets a userPrompt with instantaneous needs, inventory, structures, nearby agents, social/law context, and a salience-ranked memory window — but nothing about the agent's own recent behavior. Feature: (1) a durable audit of exactly what context each thought receives today (present vs absent); (2) add self-history to the decision prompt: current/last intent with its source (planner/reflex/plan), need trajectories (level + direction), and an active-plan echo so a thought continues its plan instead of restarting; (3) relevance-based memory retrieval and selective journal-entry inclusion feeding the prompt (building on spec 042 embedding retrieval), assembled efficiently and with intent under a measured per-thought token budget. Goal: agents that can see their own situation and recent behavior make better-composed decisions — evidence: world-01 forage↔goto_warmth thrash (Sage 436 flips) where the model could not know a reflex had just redirected it."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - An agent knows what it was just doing (Priority: P1)

When a villager's mind is asked "what do you do next?", the decision context includes the
villager's own recent behavior: the intent it is currently executing or most recently
completed, where that intent came from (its own deliberation, an instinctive reflex, or a
step of its standing plan), and how the attempt ended if it ended (completed, failed,
superseded). A villager that was just yanked toward the berry patch by instinct can see
that this happened and reason about it, instead of rediscovering its situation from
scratch every thought.

**Why this priority**: This is the smallest change that attacks the observed failure
directly. In the world-01 evidence, the deliberating mind repeatedly declared warmth
"dangerously low" with no way to know an instinct had just re-dispatched the body to
forage — 436 food↔warmth flips for one villager. Self-history is the missing sense organ;
everything else in this feature sharpens it.

**Independent Test**: Can be fully tested by capturing the assembled decision context for
an agent whose previous intent was issued by each possible source and confirming the
self-history appears, is accurate, and names the source — plus a replayed thrash scenario
showing the context now carries what the model previously could not see.

**Acceptance Scenarios**:

1. **Given** a villager whose last intent was issued by the instinct/reflex layer,
   **When** its mind is next asked to decide, **Then** the decision context states that
   intent, that it came from instinct, and its outcome so far.
2. **Given** a villager whose last intent came from its own prior deliberation, **When**
   it next decides, **Then** the decision context states that intent and that the
   villager chose it itself, so the model can knowingly continue or knowingly change
   course.
3. **Given** a villager taking its very first thought in a world, **When** it decides,
   **Then** the context contains a well-formed "no prior activity" self-history rather
   than a missing or malformed block.
4. **Given** a villager whose recent intents alternated between two goals, **When** it
   next decides, **Then** the recent-intent history (not just the single last intent)
   is visible so the alternation itself is perceivable.

---

### User Story 2 - An agent feels which way its needs are moving (Priority: P2)

Each need in the decision context carries direction, not just level: "warmth 43 and
falling" is a different situation from "warmth 43 and rising", and today the model
cannot tell them apart. Direction is derived from the need's recent movement over a
short window of game time.

**Why this priority**: Trajectory is the cheapest form of foresight. The thrash pattern
persisted because every thought saw a static snapshot: warmth 45 looked equally urgent
whether the villager had just left the fire (rising, safe to finish another errand) or
was drifting away from it (falling, must commit to recovery). Direction lets a single
thought distinguish recovering from deteriorating without any new machinery for planning.

**Independent Test**: Drive a villager's needs up and down deterministically and assert
the rendered trajectory (level + direction) matches the actual recent movement for each
need, including the steady case.

**Acceptance Scenarios**:

1. **Given** a villager warming at a fire whose warmth rose over the recent window,
   **When** it decides, **Then** warmth is presented as its level plus a rising
   direction.
2. **Given** a villager walking away from warmth into the cold, **When** it decides,
   **Then** warmth is presented as falling.
3. **Given** a need that has not meaningfully changed over the window, **When** the
   villager decides, **Then** the need is presented as steady, not as noise-driven
   rising/falling flicker.

---

### User Story 3 - An agent continues its plan instead of restarting it (Priority: P3)

When a villager has a standing multi-step plan, the decision context echoes that plan:
the steps, which step is next, and the conditions under which it remains valid. A thought
taken mid-plan can see the plan and choose to continue it, revise it, or consciously
abandon it — instead of being structurally unaware that a plan exists.

**Why this priority**: The plan mechanism already exists but is invisible to the mind
that authored it; a plan can only be continued by accident. This story converts plans
from fire-and-forget scripts into commitments the agent can reason about. It ranks below
self-history and trajectories because it only helps once agents actually hold plans, and
today plans are rare precisely because nothing reinforces them.

**Independent Test**: Give a villager a standing plan, capture the next decision context,
and confirm the plan echo names the remaining steps and next step; confirm the echo
disappears once the plan completes or is cleared.

**Acceptance Scenarios**:

1. **Given** a villager with an active plan whose second step is pending, **When** it
   next decides, **Then** the context shows the plan, marks the pending step as next,
   and shows the plan's validity conditions.
2. **Given** a villager whose plan was cleared (expired, guard failed, superseded),
   **When** it next decides, **Then** the context contains no stale plan echo, and the
   self-history (US1) records that the plan ended and why.

---

### User Story 4 - What an agent remembers is chosen for the moment (Priority: P4)

The memories and journal excerpts included in a decision context are selected for
relevance to the villager's current situation — its needs, location, recent activity, and
the entities around it — blended with the existing importance/recency ranking, and the
total situational context is assembled under an explicit, measured size budget per
thought. What no longer fits is dropped in a deliberate, documented order rather than by
accident of ranking.

**Why this priority**: This is the "richer grounding" half of the feature — high value
but dependent on the relevance-retrieval capability (spec 042) that is still landing, and
meaningful only after the self-awareness blocks (US1-US3) define what the budget must
protect. Thoughts run only a handful of reasoning turns, so a moderately sized,
deliberately assembled context is affordable; an unbounded one is not.

**Independent Test**: Construct a villager with a memory store containing both
situationally relevant and irrelevant high-importance memories; capture the assembled
context and confirm relevant items are present, the assembly respects the configured
budget, and the drop order is the documented one.

**Acceptance Scenarios**:

1. **Given** a cold villager near a fire with old memories about fire-building and
   unrelated high-importance social memories, **When** it decides, **Then** the included
   memory window contains the situationally relevant memories without exceeding the
   budget.
2. **Given** a villager whose journal contains an entry relevant to its current
   predicament, **When** it decides, **Then** a bounded excerpt of that entry is included
   without the villager having to spend its own reasoning turns fetching it.
3. **Given** an assembled context that would exceed the per-thought budget, **When** the
   context is finalized, **Then** content is dropped in the documented priority order and
   the overflow event is observable to operators.
4. **Given** the relevance-retrieval capability is unavailable (degraded mode), **When**
   a villager decides, **Then** context assembly falls back to the existing
   importance/recency selection and the thought still proceeds.

---

### User Story 5 - Operators can see exactly what an agent knew (Priority: P1)

A durable, reviewable inventory documents every block of context a villager receives at
decision time — what is present, what is deliberately absent, and where each block comes
from — and stays pinned to the code it describes. Additions from this feature (US1-US4)
land in the same inventory. An operator debugging "why did my villager do that?" can read
what the villager knew, and the running system can be checked against the inventory.

**Why this priority**: This is the audit half of the feature request and the cheapest
deliverable; every later decision about what to add or drop from context is grounded in
it. It shares P1 with self-history because it can ship first and alone and still deliver
value.

**Independent Test**: Read the inventory document; verify each listed block against a
captured real decision context from a running world; confirm listed absences are in fact
absent.

**Acceptance Scenarios**:

1. **Given** the inventory document, **When** compared against a captured decision
   context from a live world, **Then** every block in the capture is described by the
   inventory and every inventory block appears (or is marked conditional) in the capture.
2. **Given** a later change that adds or removes a context block, **When** the freshness
   gate runs, **Then** an out-of-date inventory is flagged the same way stale grounding
   notes are flagged today.

---

### Edge Cases

- First thought of a newly created villager: no prior intent, no trajectories (no
  history window yet), no plan — every block must render a well-formed empty state.
- A villager that just woke from sleep: the trajectory window spans the sleep period;
  direction must reflect what the sleeper experienced (e.g. warmth fell overnight), not
  read as noise.
- The previous intent was rejected or superseded before landing: self-history must show
  the attempt and its fate, not present it as if it executed.
- Instinct-issued intents carry no stated reasoning: self-history must render them
  honestly (source: instinct) without inventing a rationale.
- Two intents landed in quick succession (instinct override during deliberation):
  recent-intent history must preserve order and sources so the override is visible.
- Degraded mode (no language model, or relevance retrieval unavailable): context
  assembly must not block the simulation; fall back to existing selection.
- Budget pressure: when self-history + trajectories + plan echo + retrieved memories
  cannot all fit, the documented drop order applies; survival-relevant blocks are
  protected.
- Replay determinism: assembled context must be reproducible from world state alone, so
  identical replays produce identical prompts.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The project MUST contain a durable context inventory documenting every
  block of decision-time context a villager receives — its content, its source of truth,
  and conditions under which it appears — plus an explicit list of what is deliberately
  absent. The inventory MUST be pinned to the code it describes and participate in the
  existing grounding-freshness regime.
- **FR-002**: A villager's decision context MUST include a self-history block: the
  current or most recent intent, its source (deliberation, instinct, or plan step), and
  its outcome so far (executing, completed, failed, rejected, superseded).
- **FR-003**: The self-history block MUST cover a short window of recent intents (not
  only the last one), preserving order and per-intent source, so alternation between
  goals is itself visible to the deciding mind.
- **FR-004**: Each need presented in the decision context MUST carry a direction
  (rising, falling, steady) derived from its movement over a defined recent window of
  game time, with a threshold that prevents noise from flickering the direction.
- **FR-005**: When a villager has an active plan, the decision context MUST echo the
  plan: its steps, which step is next, and its validity conditions; when no plan is
  active there MUST be no stale echo. The end of a plan (completed, expired, guard
  failure, superseded) MUST be visible in self-history at the next thought.
- **FR-006**: Memory selection for the decision context MUST incorporate relevance to
  the villager's current situation (needs, location, recent activity, nearby entities)
  blended with the existing importance/recency ranking, degrading gracefully to the
  existing ranking when relevance retrieval is unavailable.
- **FR-007**: Context assembly MUST be able to include bounded excerpts from the
  villager's own journal when they are relevant to the current situation, without
  consuming the villager's own reasoning turns.
- **FR-008**: Total situational context per thought MUST be assembled under an explicit,
  configurable size budget with a documented drop order for overflow; survival-relevant
  blocks (needs, self-history) MUST be dropped last.
- **FR-009**: Context assembly MUST be deterministic given identical world state, and
  the assembled context (or a faithful summary of its blocks and sizes) MUST be
  observable per thought by operators through the existing decision-trace surfaces.
- **FR-010**: The per-thought context size (and any overflow/drop events) MUST be
  measured and recorded so operators can evaluate budget fit across model tiers.

### Key Entities

- **Self-History Block**: the villager-facing account of its own recent intents — goal,
  source (deliberation / instinct / plan), outcome, and order.
- **Need Trajectory**: a need's current level plus its direction over the recent window
  (rising / falling / steady).
- **Plan Echo**: the villager-facing rendering of its active plan — steps, next step,
  validity conditions.
- **Context Budget**: the per-thought size allowance for assembled context, with a
  documented drop order and overflow observability.
- **Situational Retrieval Query**: the representation of "the current situation" (needs,
  location, recent activity, nearby entities) used to select relevant memories and
  journal excerpts.
- **Context Inventory**: the durable audit document enumerating all decision-time
  context blocks, present and deliberately absent, pinned to the code it describes.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The context inventory exists, and a captured decision context from a live
  world matches it block-for-block (no undocumented blocks, no documented-but-missing
  blocks) at the commit it is pinned to.
- **SC-002**: In captured decision contexts across all intent sources, 100% of thoughts
  taken after at least one prior intent contain an accurate self-history naming the
  prior intent's goal, source, and outcome.
- **SC-003**: In a deterministic needs-movement scenario, 100% of rendered trajectories
  match the actual direction of the need over the window, and a steady need never
  renders as rising or falling.
- **SC-004**: In a replay of the world-01 thrash episode's conditions, the decision
  context at each planner thought contains the information that was previously missing
  (the instinct's redirection and the alternation history) — verifiable by inspection of
  the assembled contexts alone, independent of model behavior.
- **SC-005**: Per-thought assembled context stays within the configured budget in at
  least 99% of thoughts in a multi-day run, with every overflow recorded and attributable
  to its drop decision.
- **SC-006**: In a relevance-retrieval evaluation with planted relevant and irrelevant
  memories, the assembled window includes the situationally relevant items at least 80%
  of the time without exceeding budget, and degraded mode never blocks a thought.
- **SC-007**: In a post-change multi-day run under conditions equivalent to world-01,
  the rate of same-agent food↔warmth intent alternations within a short window drops by
  at least 50% relative to the world-01 baseline (436 flips for the worst villager),
  measured with the same flip-counting method used in the TASK-101 spike.

## Assumptions

- Self-history depth defaults to a small fixed window (the last few intents within
  recent game time) rather than a full log; the exact depth is a tunable detail, not a
  scope question.
- Trajectory direction is computed from existing recorded need movement; no new sensory
  machinery is implied.
- Journal inclusion is automatic and selective (stuffed by the assembler when relevant)
  per the feature request; the villager's existing ability to read its own journal via
  its tools is unchanged and complementary.
- Relevance retrieval builds on the embedding-memory capability being delivered by spec
  042 (record-at-emission vectors, relevance term); this feature consumes that
  capability and adds graceful degradation, it does not re-implement it. If spec 042 is
  not yet merged when implementation starts, US4 waits behind it while US1-US3 and US5
  proceed.
- The per-thought size budget is configurable per world (candidate home: the tuning
  manifest of TASK-107) with a sensible default; thoughts run at most a handful of
  reasoning turns, so a moderate budget is affordable on locally hostable model tiers.
- The decision-trace/observability surfaces that exist today (thought/outcome event
  records and their views) are the delivery vehicle for context observability; this
  feature extends what they carry rather than inventing a new channel.
- Behavioral improvement (SC-007) is expected to compound with the instinct-arbitration
  and recovery-intent work (TASK-103/TASK-104); the 50% target is for this feature in
  combination with whatever of that work has landed at measurement time, against the
  frozen world-01 baseline.
