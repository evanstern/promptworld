# Implementation Plan: Per-Turn Context Grounding

**Branch**: `043-context-grounding` | **Date**: 2026-07-24 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/043-context-grounding/spec.md`

## Summary

Give each villager thought an accurate view of the villager's own recent behavior and
situation: a reducer-maintained intent history ring (goal + source + outcome), need
trajectories from a windowed anchor snapshot, an active-plan echo, and
relevance-selected memories/journal excerpts — all assembled deterministically under an
explicit size budget with a documented drop order, recorded per thought for operators,
and documented in a freshness-gated context inventory wiki note. Consumes spec 042's
embedding relevance machinery (merged, TASK-98); changes no reflex/planner arbitration
(that is TASK-103/104/108).

## Technical Context

**Language/Version**: Go 1.26 (module github.com/evanstern/promptworld)

**Primary Dependencies**: none new — stdlib + existing internal packages
(`internal/sim`, `internal/mind`, `internal/llm`, `internal/cognition`,
`internal/scribe`, `internal/tool`)

**Storage**: event-sourced SQLite world store (append-only `events`); all new agent
state is reducer-derived from events (replayable), no schema change

**Testing**: `go test` — table-driven unit tests + the existing prompt frame-report
pattern (`internal/mind/prompt_test.go`); replay determinism tests mirror
`internal/daemon/embed_replay_test.go`

**Target Platform**: macOS/Linux daemon (existing `promptworld` binary)

**Project Type**: single Go project, internal simulation packages

**Performance Goals**: context assembly is pure in-memory rendering on the mind path —
no new I/O, no new LLM calls; per-thought assembly overhead negligible vs. the LLM call
it precedes

**Constraints**: deterministic replay is doctrine — every new agent-state field must be
derived from reduced events only; shadow-mode byte-identity invariant
(`internal/mind/shadow_test.go`) must keep holding; prompt growth bounded by the new
context budget (default ≈2k approx-tokens, tuning-manifest-ready per TASK-107's const
fallback pattern)

**Scale/Scope**: 8 villagers/world today; context assembly per planner thought
(cadence 1800 ticks/agent) — dozens of assemblies per game-hour, trivial load

## Constitution Check

*GATE: evaluated against constitution v1.1.0 before Phase 0; re-checked after Phase 1.*

- **I. Artifact-Grounded Action** — PASS: spec/plan/research/data-model/contracts in
  `specs/043-context-grounding/`; board task TASK-105 carries decisions; the feature
  itself produces a durable context inventory (wiki note).
- **II. One Task, One PR** — PASS: TASK-105 is the deliverable; all user stories land
  as commits on one branch (`.worktrees/task-105`), one PR. Spec docs commit to main.
- **III. Gates Over Assertions** — PASS: spec-bridge will mirror phase criteria; status
  moves only on artifacts (tests, merged PR).
- **IV. Grounding Freshness** — PASS with obligation: this feature touches
  `internal/mind/prompt.go`, `internal/sim/*` — sources of `docs/wiki/agent-mind.md`,
  `memory-retrieval`, `sim-state-reducer`, `event-types` notes; wiki re-pin
  (`/grounding-wiki:wiki-update`) is an explicit post-merge task, and the new
  `decision-context` inventory note joins the corpus.
- **V. Model-Tiered Workflow** — PASS: plan authored by planning tier; implementation
  delegated — US1/US2/US3/US4 touch `internal/mind` orchestration and reducer state
  (doctrine-adjacent) → **Opus 4.8 tier**; US5 inventory note + fixture-style tests →
  **Sonnet tier**. Tier choices recorded on TASK-105 at dispatch.

**Post-Phase-1 re-check**: PASS — no violations introduced; no Complexity Tracking
entries needed.

## Project Structure

### Documentation (this feature)

```text
specs/043-context-grounding/
├── spec.md
├── plan.md              # This file
├── research.md          # Phase 0: decisions R1-R9
├── data-model.md        # Phase 1: IntentRecord ring, NeedsAnchor, context blocks
├── quickstart.md        # Phase 1: validation scenarios mapped to SCs
├── checklists/requirements.md
├── contracts/
│   └── context-blocks.md  # Phase 1: prompt block contract (order, caps, drop order)
└── tasks.md             # Phase 2 (/speckit-tasks — not created here)
```

### Source Code (repository root)

```text
internal/sim/
├── agents.go        # Agent: +IntentLog ring, +NeedsAnchor/NeedsAnchorTick; IntentRecord type
├── state.go         # reducer arms: intent_set appends record; intent_done/build_failed/
│                    #   intent_rejected/plan_expired stamp outcomes; needs_changed
│                    #   refreshes anchor on window edge
├── cognition.go     # CogThoughtPayload: +PromptBytes, +BlockBytes (per-block sizes)
└── memory.go        # unchanged (042 consumed as-is); journal.go: exported excerpt helper

internal/mind/
├── prompt.go        # userPrompt: self-history, trajectories, plan echo, journal
│                    #   excerpts; block assembler with budget + drop order
├── context.go       # NEW: block assembler (block registry, caps, drop order, sizing)
├── telemetry.go     # cog.thought carries prompt/block sizes
└── prompt_test.go   # block rendering, empty states, budget/drop tests
     context_test.go # NEW: assembler unit tests
     shadow_test.go  # invariant keeps holding

docs/wiki/
└── decision-context.md  # NEW: the context inventory (US5), pinned + freshness-gated
```

**Structure Decision**: existing two-package split holds — `internal/sim` owns
deterministic state (history ring, anchors, payloads), `internal/mind` owns rendering
and assembly. The only new file is `internal/mind/context.go` (assembler) plus its
test; everything else extends files in place.

## Complexity Tracking

No constitution violations; table intentionally empty.
