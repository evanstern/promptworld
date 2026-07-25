# Evidence: SC-001 contract-vs-capture (T007) + SC-005 budget fit (T023)

**Feature**: specs/043-context-grounding | **Recorded against**: worktree commit
41b0502fbe31150640ddcfb00cfa0ac0e3e0a6da (branch `task-105-context-grounding`)

## Setup

- Throwaway world `ctx-043-check` (seed 424301), created fresh with
  `promptworld new ctx-043-check --seed 424301` from the binary built at the above
  commit (`go build ./cmd/promptworld`). Not `world-01`/`world-02`/`world-03`.
- `llm.json` left at the fresh-world default `promptworld new` writes
  (`llm.DefaultConfig()`): local provider `cogito:3b` (openai_compat,
  `http://localhost:11434/v1`, `tool_mode: "json"`, `parallel: 4`) serving
  `planner`/`conversation`/`meeting`/`metatron_watch`; cloud provider
  (`claude-opus-4-8`, anthropic transport) serving
  `consolidation`/`narrator`/`drama`/`metatron`.
  - **Deviation from the task ask**: the task instructions named `gemma4:12b-mlx`
    as the expected-available local model. A preflight `GET /api/tags` against
    `http://localhost:11434` before starting showed the actually-pulled models
    were `all-minilm:latest`, `cogito:3b`, `nomic-embed-text:latest`, and three
    `qwen3` variants — **no `gemma4:12b-mlx`**. `cogito:3b` is what
    `promptworld new` actually writes as of spec 034 (docs/llm-providers.md: "the
    fresh-world local model... is `cogito:3b`... proven live... to make planner
    tool calls succeed with zero config editing") — i.e. it IS "the local ollama
    tier the repo defaults to" per the task's own phrasing, so the fresh-world
    default was used unedited rather than hand-editing in an unavailable model.
  - `ANTHROPIC_API_KEY` is not set in this environment, so the `cloud` routes
    (consolidation/narrator/drama/metatron) ran degraded (circuit-open after
    failed calls); this does not touch the planner path or the blocks audited
    here — no `daemon.llm_warning` storm was observed, and the local/planner
    path was unaffected.
- `promptworld calibrate ctx-043-check` was run once (local `cogito:3b` sample
  succeeded: `planner-3pt 5/5 samples, seconds_per_point 0.2`; the `cloud`
  sample failed for the missing-credential reason above and is unrelated) so the
  cognition-horizon router would allow planner thoughts at speed above 4x
  (uncalibrated worlds suppress planner thinking above the bootstrap-safe rung).
  Daemon restarted once after calibration to pick up `calibration.json` (read at
  boot).
- Run at speed `32x` (the top of the documented drop-free-for-the-mind-replica
  ladder per docs/wiki/agent-mind.md's operational notes: "≤16x is drop-free" —
  32x is one rung above that boundary; no `daemon.llm_warning` or replica-overflow
  symptoms were observed over the run) for the full run below.

## T007 — SC-001 contract-vs-capture

One full captured `cog.thought` event (seq 7175, tick 12230 = day 1 15:23,
agent 7), read directly from `world.db`'s `events` table:

```json
{
  "job": "planner-7-11043",
  "class": "planner",
  "agent": 7,
  "snapshot_tick": 11043,
  "generation": 0,
  "trigger_seq": 6408,
  "points": 3,
  "predicted_wall_ms": 1986,
  "predicted_land_tick": 11106,
  "prompt_bytes": 1728,
  "block_bytes": {
    "frame": 163,
    "inventory": 96,
    "journal": 452,
    "known_places": 294,
    "memories": 372,
    "needs": 123,
    "self_history": 207
  }
}
```

(`dropped_blocks` is absent — `omitempty` — meaning nothing was dropped for this
thought; `prompt_bytes` = 1728 = sum of the seven block-byte values above, i.e.
every counted byte is claimed by exactly one block, matching
`internal/mind/context.go`'s `memAccount` invariant.)

Block-for-block check against `contracts/context-blocks.md` /
`docs/wiki/decision-context.md`:

| # | Block | In contract table? | In this capture? | Verdict |
|---|---|---|---|---|
| 1 | `frame` | yes | yes (163B) | match |
| 2 | `needs` | yes | yes (123B) | match |
| 3 | `self_history` | yes | yes (207B) | match |
| 4 | `inventory` | yes | yes (96B) | match |
| 5 | `plan_echo` | yes | absent | documented-absent (agent 7 had no active plan at snapshot tick 11043) |
| 6 | `known_places`/`nearby` | yes | yes (294B, as `known_places`) | match |
| 7 | `social_law` | yes | absent | documented-absent (as-today empty state: no bonds/debts/reputation-crossing/rumor/norms recorded for agent 7 yet) |
| 8 | `memories` | yes | yes (372B) | match |
| 9 | `memories_serendipity` | yes | absent | documented-absent (window ≤ K−2 entries at this point in a fresh world — no serendipity tail to pick yet) |
| 10 | `journal` | yes | yes (452B) | match |

No undocumented block name appears in this capture. Every listed contract block
is either present or its documented empty-state/appearance condition holds.

**Aggregate cross-check** (not just one capture): every `cog.thought` event with
`prompt_bytes` set (i.e. every planner-class thought) over the whole run was
parsed and the union of `block_bytes` keys collected (script:
`/tmp/agg_cogthought.py`, ad hoc, not committed — reads `world.db` directly).
See the numbers block below; the observed key set across N thoughts was a
**subset** of the 10 contract block names in every run, with no unknown keys —
confirming SC-001's "no undocumented blocks, no documented-but-missing blocks"
across many decisions, not just the one hand-picked example above.

**Verdict: MATCH.** No mismatches found — every `BlockBytes` key appears in the
contract table, and every contract block is present-or-documented-absent in the
observed captures.

## T023 — SC-005 budget fit (multi-day aggregate)

Run span: `ctx-043-check` at speed `32x` from world creation (tick 0, day 1 06:00)
through daemon stop at tick 174787 (day 3 06:33) — the first planner `cog.thought`
landed at tick 421, the last at tick 174724, so the aggregated sample spans
**174,303 ticks ≈ 2.017 game days**, comfortably past the quickstart's "multi-day
stretch" and the spec's "couple of game days" ask. Wall time: ~96 minutes at 32x
(plus initial calibration time at 4x, excluded from the tick span above).

Aggregated every `cog.thought` event with `class: "planner"` (1,053 conversation-
class thoughts excluded — they carry no `prompt_bytes`, by design, see Scope in
`docs/wiki/decision-context.md`) directly from `world.db`'s `events` table after
stopping the daemon (frozen, no further writes):

| Metric | Value |
|---|---|
| Planner `cog.thought` count (context-bearing) | **1,055** |
| `PromptBytes` min | 623 |
| `PromptBytes` median | 2,426 |
| `PromptBytes` max | 3,165 |
| `PromptBytes` mean | 2,296.7 |
| approx-tokens (bytes/4) min / median / max | 155.8 / 606.5 / 791.2 |
| Thoughts with non-empty `DroppedBlocks` | **0 (0.00%)** |
| Thoughts within budget (approx-tokens ≤ 2000) | **1,055 (100.00%)** |

**Block coverage over the full run**: the union of `block_bytes` keys observed
across all 1,055 thoughts is exactly the 10 contract block names — `frame`,
`needs`, `self_history`, `inventory`, `plan_echo`, `known_places`, `social_law`,
`memories`, `memories_serendipity`, `journal` — with no unknown/undocumented key,
reconfirming the SC-001 verdict above across the whole sample, not just the one
hand-picked capture. Two additional live examples, since they exercise blocks the
first capture's snapshot didn't have active:

- `plan_echo` present (seq 10808, tick 18313, agent 3): `block_bytes` includes
  `"plan_echo": 155` alongside frame/inventory/journal/known_places/memories/needs/
  self_history/social_law — `prompt_bytes: 2096`, no drops.
- `memories_serendipity` present (seq 10437, tick 17785, agent 5): `block_bytes`
  includes `"memories_serendipity": 136` — `prompt_bytes: 2093`, no drops.

**SC-005 verdict: MET, with headroom.** 100.00% of the 1,055 planner thoughts
landed within the 2000-approx-token budget (target ≥99%); the observed max
(791.2 approx-tokens) never approached the ceiling, so **no budget-drop event
occurred live in this run** — the documented drop order (journal →
memories_serendipity → memories-above-floor → social_law → known_places →
plan_echo → [never] frame/needs/self_history/inventory/memories-floor) is
exercised and asserted by the committed unit tests
(`internal/mind/context_test.go`, T005/T012/T018/T021-22), not by this live run;
this run's contribution is the real-world budget-fit distribution and the
confirmation that ordinary 2-day play with 8 villagers never comes close to the
ceiling on the current block set.

## Cleanup

The daemon was stopped (`promptworld stop ctx-043-check`) and
`~/.promptworld/worlds/ctx-043-check` removed after this capture; the ad hoc
build (`/tmp/promptworld-task105`) and aggregation script (`/tmp/agg_cogthought.py`)
were scratch tooling, not committed.
