---
name: tool-use-dispatch
description: The cognition-horizon gate that suppresses a thought before enqueue, the spec-106 sleep gate and in-flight cancel that stop a queued or running thought whose agent slept or died, the per-job immutable snapshot and thoughtMeta identity, and how runPlan drives the bounded tool-use loop against the villager roster — replacing the retired free-text reply contract with tool schemas. Split from [[agent-mind]] (The mind driver, dispatch half).
kind: component
sources:
  - internal/mind/mind.go
  - internal/mind/prompt.go
  - internal/mind/context.go
  - internal/mind/telemetry.go
verified_against: c5e66ee92fa75c00e2b480811e0ca727d5c1a1e1
---

# Cognition gate and tool-use loop dispatch

Before enqueue, each due agent passes the cognition-horizon gate
(`routeVerdict` in telemetry.go, backed by [[cognition]]'s deterministic
router): a planner thought whose predicted drift exceeds its staleness budget
at the current speed is never attempted — a `cog.outcome{suppressed}` records
the arithmetic, and the reflex floor is the degrade action. While the replica
is paused, `routeVerdict` short-circuits to [[cognition]]'s `RoutePaused`
(spec 040): a frozen world cannot drift, so every class routes allow at zero
predicted drift with an arithmetic string naming the paused state — checked
before the uncapped branch, so paused wins even on a world set to max. Since spec 037
(FR-004), `emitSuppressed` (telemetry.go) also reports the suppression to
[[llm-orchestrator]]'s daemon-lifetime per-class counters through the
optional `suppressionCounting{RecordSuppression(class)}` seam (the same
optional-interface pattern as `estimating` below) before the detached event
emit — an O(1) mutex bump that never blocks the absorb loop; a test fake or
nil orchestrator lacking the seam is a silent no-op. The count feeds
[[ipc-server]]'s status-wire horizon (`SuppressedCount` per class). Allowed agents are
enqueued as immutable prompt snapshots to a single-flight-per-agent planner
worker — a model call must never block the absorb loop, or the events channel
overflows at high speed and edge triggers are dropped. Each job carries a
`thoughtMeta` identity (job id, decision class, snapshot tick, agent
generation, trigger seq, predicted wall-ms and landing tick from
[[cognition]]'s latency estimate; while paused, `newMeta` predicts the landing
AT the snapshot tick — spec 040's truth rule: a frozen-tick thought lands now,
not at the set-speed projection) plus a snapshot of every agent's position
(`agentSnap`) — the assumptions guards are built from. Because the planner
class is `FutureDated`, the prompt opens with `futureDated` (prompt.go): "your
decision will take effect around <landing clock> — plan for then" — a no-op
while paused, since landing ≤ now, so prompt, gate, and record agree; since
spec 043 the line rides INTO the assembler as part of the frame block (same
bytes, one owner) rather than being prepended after. The situation itself is
now assembled from named blocks in a fixed contract order under a size
budget (`assembleContext`, context.go — [[decision-context]] owns the block
inventory, sources, and drop priority): `plan()` calls the assembler
directly and stamps the result's `promptBytes`/`blockBytes`/`droppedBlocks`
onto the thought's telemetry (`thoughtMeta`, landing additively and
`omitempty` on `cog.thought` — non-planner paths like conversation scenes
assemble no block set and marshal byte-identically); `userPrompt` and the
exported `AssembleUserPrompt` (prompt.go) remain the no-future-line entry
points — a pure function of world state, so replay tooling and the TUI
capture path reproduce a thought's exact bytes. Each job now
drives a bounded **tool-use loop** (spec 017, `toolloop.Run`, [[tool-loop]])
rather than one bare planner call: `runPlan` builds a `villagerDispatch` (the
job, a wall-clock start, a buffered `CallRecord` sink, and the `doorOutcome`
flag that tells the terminal switch below whether a door already recorded an
outcome) and calls `md.runLoop` — production wires this to `toolloop.Run`
against the concrete `*llm.Orchestrator` (`New`'s `runLoopOverride` variadic
seam installs a scripted driver instead, race-free, for tests that stub the
model through `Submitter`) — with `Kind: llm.KindPlanner`, the persona system
prefix, the situation+memory-window suffix as the loop's `Seed` turn,
`Roster: tool.LoopRosterVillager()`, per-tool handlers from
`md.villagerHandlers`, `MaxRounds: md.loopRounds` (`llm.json`'s
`loop_max_rounds`, normalized), and `MaxTokens: md.plannerTokens` (spec 025,
TASK-72: the operator-tunable `llm.json` `max_tokens.planner` budget, threaded
through `mind.New` like `loopRounds`; default 512, up from the pre-loop 256 —
a tool-era round carries a `tool_use` block alongside any prose, so the budget
grew to avoid truncating a call mid-arguments — with the rationale now living
on the config default in [[llm-orchestrator]]'s `config.go`). `mind.New` also
threads `consolidationTokens` (`max_tokens.consolidation`, default 1024) for
[[nightly-consolidation]]'s Submit. The model
may call read tools first (since spec 019 the villager roster ships two —
`search_journal`/`read_journal`, the first production Read tools, [[agent-journal]])
then must commit to exactly one acting tool — a world verb, `set_plan`, `muse`,
or a journal write/delete — which lands through its existing door; the loop's cardinality rule means every
call after the first acting one is rejected, so "one thought, one act" is
structural now rather than parser-enforced. The retired free-text contract —
`planReply`/`planStepReply`, the parser's `worldGoals`/`validKinds` accept
sets, `validateKindQty`, `plannerReplySchema()`/`plannerSchema`, `parseReply`,
and the golden-prompt fixture (`prompt_golden_test.go`) that pinned it — is
gone: the goal vocabulary, storage kind/qty validation, and the guarded-plan
step cap now live as tool schemas ([[tool-registry]]'s `InputSchema`,
`set_plan`'s authored override) the loop driver itself validates before
dispatch, not a parser the mind ran after the call returned. Since spec 106
two sleep-gating layers sit around the loop, closing the post-enqueue windows
playtest-1 measured (905 "is asleep" landing rejections — thoughts spent on
villagers who reflex-slept while the job waited or ran): at the TOP of
`runPlan`, before `cog.thought` or any model call, a **dequeue gate** consults
the mind's atomic per-agent unavailability mirror (`unavail`, asleep|dead —
absorb refreshes it at batch end beside the `md.tick`/`md.tickRate` mirrors
and `New` seeds it from the snapshot; derived from replica state, so
`gru.attacked`'s eventless wake is covered) and an unavailable agent's job
terminates in one `cog.outcome{suppressed}` with a dequeue reason ("asleep at
dequeue" / "dead at dequeue" — dead first, mirroring `rungUnavailable`'s
ordering) — deliberately WITHOUT the `RecordSuppression` bump, so the spec-037
`SuppressedCount` keeps meaning "router suppressed", and without re-arm (the
wake trigger owns resumption; [[mind-driver-triggers]]). Second, an
**in-flight cancel**: `runPlan` registers its call context's cancel in a
per-agent race-safe slot (`planCancel`, registered before the loop, cleared
after) which absorb fires on `agent.slept`/`agent.died` — the running call
aborts instead of finishing a dead-on-arrival thought. Planner slot only;
consolidation, narrator, meeting, reconcile, and scene workers are untouched,
and the enqueue-time `plan()` skip plus the sim's landing ladder
(`rungUnavailable`, the authoritative backstop for mirror lag) are
byte-unchanged. `systemPrompt`
(prompt.go) no longer renders the goal line/gloss block or the JSON reply
shape — it only frames the choice; the tools themselves carry the vocabulary
via their declared name/description/schema. Since spec 027 (TASK-73) that
frame is a crafted three-part structure — one identity statement (the only
place the frame names the agent), the persona verbatim as its own block
(vanishing cleanly when empty), and name-free second-person task framing
("calling exactly one acting tool... a world action, a short plan (set_plan),
or a passing thought (muse)") — doctrine unchanged, and still a pure function
of (name, personaText) so it stays the cacheable per-agent prefix (the
`cache_control` system block). The rewrite was eval-gated
(specs/027-villager-prompt-quality/eval/): on identical seeded soaks it cut
`rejected_malformed` from 15.34% to 11.50% and roughly halved the muse share;
a worked tool-selection exemplar was measured and REJECTED (worse malformed
rate, +30% tokens, cook-verb anchoring), so the frame ships exemplar-free.

## Connections

[[agent-mind]] is the parent note this child was split from; [[mind-driver-triggers]]
is its sibling child covering what arms a thought before this gate runs;
[[cognition]] owns the deterministic router (`routeVerdict`'s backing
decision-class registry, staleness budgets, and `RoutePaused`) this gate
calls; [[decision-context]] owns the named-block assembler (`assembleContext`/
`assembleBudget`) invoked here and its size-budget drop order; [[tool-loop]]
is `toolloop.Run`, the driver `md.runLoop` wraps; [[tool-registry]] declares
the roster/schemas (`InputSchema`, `set_plan`'s override) that replaced the
retired parser; [[agent-journal]] owns the two journal Read tools on the
roster; [[llm-orchestrator]] carries the model calls this dispatch enqueues.

