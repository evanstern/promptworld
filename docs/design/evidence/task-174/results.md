# TASK-174 / spec 103 — conversation-outcome JSON robustness: soak evidence (FR-005, SC-001)

**Status: PARTIAL.** A real soak ran against this branch's own code for ~25
real minutes and was then stopped and torn down (bounded window — see
"Disposition" below). It did not reach the ≥ 20 founded-scenes target. What
it measured is recorded honestly below; **SC-001 is NOT yet demonstrated**
and the soak's daemon/world no longer exist (per the operator's bounded-window
instruction, they must not outlive the implementation session).

**Branch**: `task-174-conversation-outcome-json`. **World**: built and run
entirely under the session scratchpad (never `~/.promptworld/measure/` or
`~/.promptworld/worlds/` — several other sweep lanes had live daemons under
`~/.promptworld/measure/` at the same time; this soak never touched that
preserved state, and no world/binary from it survives this report). **Seed**:
1337. **Stage**: stage-4 (overridden, matching the task-81/99 measurement-run
precedent — a fresh village, never the operator's playtest world).

## Setup (local-only routes, zero paid spend)

- Binary: built from this branch (`go build ./cmd/promptworld`) into the
  session scratchpad, so the soak exercised the branch's own
  `ResponseSchema`/`convoOutcomeSchema`/`sayReplySchema` code, not a stale
  release binary.
- World: `promptworld new <scratchpad>/soak-world --seed 1337 --stage stage-4
  --override`.
- `llm.json`: every kind (including `conversation`) routed to a single
  `local` provider — `openai_compat`, `http://mbpro-m1.local:11434/v1`,
  `gemma4:12b-mlx` (the measurement-run recipe's villager model); `embedding`
  on a separate local Ollama endpoint (`all-minilm:latest`). No `anthropic`/
  9router provider declared at all — no paid-spend surface in this world's
  config at any point.
- Speed: started at the daemon's uncalibrated default (4x), raised to 8x once
  `promptworld calibrate` confirmed conversation-class cognition is not
  suppressed at that speed (`local` measured ~20s/pt on a 3-point planner
  shape; calibrate's own caveat: measured one-call-at-a-time, live concurrent
  load runs the effective rate higher — confirmed true here, see below).

## D6 metrics — query source (this part is DONE, reusable, verified)

`docs/design/evidence/task-174/queries.sql` runs directly against a world's
`world.db` (SQLite, WAL mode, safely readable with `sqlite3 -readonly` while
the daemon has it open) and reproduces D6 exactly:

- **outcome parse-failure count** = `cog.outcome{retried}` events with reason
  prefix `"outcome: "`, plus terminal `cog.outcome{unusable}` events with
  reason prefix `"outcome: "` and non-empty `raw`.
- **abandoned-scene rate** = terminal `unusable` / all terminal conversation-job
  `cog.outcome` events (job prefix `conversation-`, excluding the non-terminal
  `retried` marker).

The query file was exercised against the live, populated `world.db` below and
required one portability fix along the way: this machine's `sqlite3` CLI
(3.x) misparses a bare `/` division operator sitting alone on its own script
line as a statement terminator — division must share a line with its
operands. Fixed in the committed query file; not a promptworld code issue.
**This query file is the durable, reusable deliverable of T006** — it can be
pointed at any world's `world.db` (including a future, longer soak) with no
further changes.

## What the soak actually measured before it was stopped

Over ~25 real minutes at 4x→8x, the world reached game tick 9321 (day 1,
~08:35 — under 3 game-hours of the 29-game-day playtest-1 baseline). Exactly
**one** conversation scene founded in that window, and it did **not** exercise
the schema fix at all: it died to a **transport** timeout (`context deadline
exceeded` at 120s: `predicted_wall_ms: 284492`, `actual_wall_ms: 120000`), the
pre-existing TASK-42 transport-never-retried path — not a parse failure.

Final query run against the (now-deleted) `world.db`:

```
outcome_parse_failures: 0
founded_scenes:         1
abandoned_scenes:       1
abandoned_pct:          100.0   (n=1 — not statistically meaningful)
outcome breakdown: unusable=1 (reason: transport timeout, not an outcome parse failure)
```

**Root cause of the slow rate**: several other sweep lanes were concurrently
running on this same machine for the whole window (visible in `ps` as other
worktrees' own soaks and `go test -race` runs, plus system load average
15–20 for most of the window), and — critically — at least one of those other
sessions' own soaks (`task-112-soak-{default,control,authored}`, confirmed
long-running via `ps`) appears to also target Ollama-family local endpoints,
so the shared `mbpro-m1.local` host was under real contention. A single
conversation-outcome call is budgeted for far less than the 120s transport
timeout under uncontended conditions (TASK-58's calibration data); hitting
that ceiling here reflects host contention, not the schema fix.

## Disposition (per the operator's bounded-window instruction)

The soak was given a bounded window inside this implementation session, did
not reach ≥ 20 founded scenes in that window, and was then torn down
completely: `promptworld stop <scratchpad>/soak-world`, followed by deleting
the entire scratchpad directory (binary + world + `world.db`). **No daemon
and no world from this soak survive this report.** Nothing under
`~/.promptworld/` was ever touched by it.

## What remains to reach SC-001 (≥ 20 founded scenes)

The soak needs to be re-run for a longer, uncontended (or at least
less-contended) window — a fresh scratchpad world (never
`~/.promptworld/measure/` or `~/.promptworld/worlds/`), the same `llm.json`
shape documented above, run until `founded_scenes` from
`docs/design/evidence/task-174/queries.sql` reaches ≥ 20, then:

1. Re-run the query file against that world's `world.db`; record the five
   result blocks in this file, replacing the single-scene numbers above.
2. Record the final playtest-1-vs-soak comparison table (outcome
   parse-failure count, abandoned-scene rate side by side) and state the
   SC-001 verdict — being careful, per D6, to separate a transport-timeout
   abandonment (like the one scene recorded here) from an outcome-parse-
   failure abandonment; only the latter is what the schema fix targets.
3. Tear the soak world down the same way once evidence is captured (or hand
   it to the operator's own measurement-run recipe worlds under
   `~/.promptworld/measure/`, which ARE meant to persist, if a longer-lived
   run is preferred over a scratchpad one).

## Playtest-1 baseline (for the eventual final comparison)

22 outcome parse failures; 83 `cog.outcome{unusable}`; 62/293 conversations
abandoned (21%) — conversation routed to local `gemma4:12b` via Ollama,
29 game-days, no structured-output schema (pre-TASK-174 code).
