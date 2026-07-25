# Quickstart: Per-Turn Context Grounding — validation

**Feature**: specs/043-context-grounding
**Prerequisites**: Go toolchain; a scratch world (`promptworld` daemon) for the live
checks; world-01's event db (`~/.promptworld/worlds/world-01/world.db`) for SC-007's
baseline comparison.

## Unit-level (deterministic, no LLM)

```sh
go test ./internal/sim -run 'IntentLog|NeedsAnchor'     # ring transitions, anchor window, deadband
go test ./internal/mind -run 'Context|PromptFrame'      # block rendering, empty states, budget drops
go test ./internal/mind -run 'Shadow'                   # spec-042 shadow invariant still holds
go test ./internal/daemon -run 'Replay'                 # replay determinism unaffected
```

Expected: all pass. Covers SC-002 (self-history accuracy per source), SC-003
(trajectory correctness incl. steady deadband), SC-005 mechanics (budget + drop order),
FR-005 empty-state (no stale plan echo).

## Contract check (SC-001)

1. Read `contracts/context-blocks.md` and `docs/wiki/decision-context.md`.
2. Capture one real assembled prompt: run a scratch world with LLM on, take any
   villager thought, and read its `cog.thought` event (`BlockBytes` keys) plus the
   prompt text via the decision-trace view.
3. Verify block-for-block: every `BlockBytes` key appears in the contract table; every
   contract block is present or its documented empty-state condition holds.

## Thrash-episode context replay (SC-004)

1. From world-01's db, extract Sage's episode window (ticks 265,411–266,631; the
   TASK-101 spike documents it).
2. Reconstruct sim state at each planner thought in the window (replay to tick) and run
   the assembler.
3. Verify the assembled context contains: the reflex-issued `forage` in `self_history`
   (source "instinct"), the alternation visible across ≥3 records, and warmth rendered
   with a falling direction while walking away — inspection of assembled text only,
   no model in the loop.

## Budget fit over a run (SC-005, FR-010)

1. Run a scratch world at moderate speed for a multi-day stretch.
2. Aggregate `cog.thought` events: `PromptBytes` distribution, count of non-empty
   `DroppedBlocks`.
3. Expected: ≥99% of thoughts within budget; every overflow lists its dropped blocks.

## Retrieval relevance + degraded mode (SC-006, US4)

1. Planted-memory test (unit): agent with situationally relevant + irrelevant
   high-salience memories → window includes the relevant items ≥80% across seeds,
   within budget.
2. Degraded: config with no embedding route → daemon runs, thoughts proceed, `memories`
   block falls back to legacy selection, journal block still term-matches; no
   `daemon.llm_warning` storm.

## Behavioral flip-rate (SC-007 — measured, not gated per-commit)

1. Baseline: world-01 flip counts from the TASK-101 spike method (forage↔goto_warmth
   A→B→A transitions within ≤200 ticks, per agent).
2. Post-change: equivalent multi-day run on the same map/seed posture with this feature
   (and whatever of TASK-103/104 has landed) enabled; recount with the same script.
3. Expected: worst-agent flip rate ≤50% of baseline; record the comparison on TASK-105.
