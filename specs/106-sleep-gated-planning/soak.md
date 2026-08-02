# T007 — Soak evidence: measurement queries (SC-003)

**Status: COMPLETE — soak run 2026-08-01/02, SC-003 passes on every clause.**
Measured counts are in "Results" below; the queries that produced them are
unchanged from when this file was written. Nothing here is fabricated — every
number is reproducible by running the queries below against the recorded world.

## Baseline (playtest-1, the card's evidence)

- 905 "is asleep" `agent.intent_rejected` over 29 game-days ≈ **31.2/game-day**.
- Target (SC-003): **≤ 1/game-day** (≥ 97% reduction) over a ≥ 3-game-day soak,
  and **zero** planner `cog.thought` rows for agents asleep at submit.

One game-day = 86,400 ticks (`sim.NightIndex`).

## Queries (sqlite3, against the soak world's `world.db`)

All counts are derivable from the event log alone (US3): every skip is a
`cog.outcome{suppressed}` with a dequeue reason, every cancel a
`cog.outcome{unusable}` with an in-flight reason, and the ladder's residual
rejections keep their exact pre-106 shape.

### 1. "is asleep" rejections per game-day (the SC-003 headline)

```sh
sqlite3 world.db "
SELECT COUNT(*) AS asleep_rejections,
       (SELECT MAX(tick) FROM events) / 86400.0 AS game_days,
       COUNT(*) / ((SELECT MAX(tick) FROM events) / 86400.0) AS per_game_day
FROM events
WHERE type = 'agent.intent_rejected'
  AND json_extract(payload, '$.reason') LIKE '% is asleep';"
```

(Dead-agent parity, same shape: `LIKE '% is dead'`.)

### 2. Planner thoughts submitted while asleep (target: 0)

A planner `cog.thought` whose agent's most recent sleep-state edge before the
thought's row was `agent.slept` (agent refs are spec-086 `{id,name}` objects
on current rows):

```sh
sqlite3 world.db "
SELECT COUNT(*) FROM events t
WHERE t.type = 'cog.thought'
  AND json_extract(t.payload, '$.class') = 'planner'
  AND (SELECT e.type FROM events e
        WHERE e.type IN ('agent.slept', 'agent.woke')
          AND json_extract(e.payload, '$.agent.id') =
              json_extract(t.payload, '$.agent.id')
          AND e.seq < t.seq
        ORDER BY e.seq DESC LIMIT 1) = 'agent.slept';"
```

Caveat: this log-derived predicate does not see `gru.attacked`'s eventless
wake (spec edge case); a thought between a victim's `agent.slept` and its
attack would false-positive. Cross-check any nonzero count against
`gru.attacked` rows before calling it a defect.

### 3. The gate's own work (layer accounting, US3)

```sh
sqlite3 world.db "
SELECT json_extract(payload, '$.reason') AS reason, COUNT(*)
FROM events
WHERE type = 'cog.outcome'
  AND json_extract(payload, '$.reason') IN
      ('asleep at dequeue', 'dead at dequeue',
       'cancelled in flight: agent slept', 'cancelled in flight: agent died')
GROUP BY reason;"
```

Dequeue rows must carry `outcome = suppressed`, in-flight rows
`outcome = unusable`. Sanity: suppressed dequeue rows have NO matching
`cog.thought` for the same `job` (the router-suppression precedent);
in-flight rows have exactly one.

### 4. FR-003 cross-check

The daemon's spec-037 `SuppressedCount` (horizon surface) must NOT grow by
the query-3 dequeue count — sleep-skips are distinguished by reason only,
never counted as router suppressions.

## Run recipe (for the operator measurement run)

Paired seed worlds, measurement dials (memory: seed 1337, harsh dials, local
router for LLM routes), ≥ 3 game-days each; compare query 1's `per_game_day`
against the 31.2 baseline and record both counts on TASK-175's board card.
Any test world lives under a temp directory — never `~/.promptworld/worlds/`
or `~/.promptworld/measure/` — and its daemon is stopped and deleted after.

## Results (soak run 2026-08-01/02)

**Run of record.** Throwaway world in a session scratchpad — never
`~/.promptworld/worlds/` or `~/.promptworld/measure/` — seed 1337, stage-4
(`--override`), all LLM routes on a single local provider, 16x, run to
**tick 1,037,280 = 12.005 game-days**, four times the ≥ 3-game-day bar. Binary
built from `main` with spec 106 merged (PR #148).

*Recorded deviation from the recipe above:* stage-4 defaults were used rather
than harsh dials, because this world was shared with TASK-174's conversation
soak, whose founded-scene target harsh dials would have starved. SC-003's
metric is dial-independent — villagers keep the same sleep cycle either way,
and the gate acts at dequeue/in-flight, not on survival pressure. Operator
approved the sharing 2026-08-01.

### Query 1 — "is asleep" rejections per game-day (the SC-003 headline)

| | baseline (playtest-1) | target | **measured** |
|---|---|---|---|
| `agent.intent_rejected` "is asleep" | 905 over 29 game-days | — | **0** |
| per game-day | 31.2 | ≤ 1 | **0.0** |

A 100% reduction, against a target that allowed up to 1 per game-day.

### Query 2 — planner thoughts submitted while asleep

**0** (target 0). The `gru.attacked` eventless-wake caveat noted above never
had to be applied: the count is zero, so there is no false positive to
cross-check.

### Query 3 — the gate's own work (proves the zero is not vacuous)

A zero in query 1 would be meaningless if nobody slept. They did — **244
`agent.slept` / 242 `agent.woke`** edges over the run — and the gate is
visibly doing the work:

| reason | outcome | count |
|---|---|---|
| `asleep at dequeue` | `suppressed` | 102 |
| `cancelled in flight: agent slept` | `unusable` | 88 |
| `dead at dequeue` | `suppressed` | 1 |
| `cancelled in flight: agent died` | `unusable` | 1 |

**192 planner round-trips prevented over 12 game-days**, every one of which
would previously have become an "is asleep" rejection. Outcome tagging matches
the spec exactly: dequeue rows `suppressed`, in-flight rows `unusable`.

### Independent confirmation

A second soak (same seed and stage, different local model — `qwen3.6:latest`,
run for TASK-174) reproduces the result in a separate world: **0 "is asleep"
rejections**, with the sleep gate again firing normally. SC-003 is not an
artifact of one world or one model.

## Caveat: the dead-agent path is leakier than the sleep path

Recorded because the recipe above frames dead-agent as a same-shape parity
check, and the parity does **not** hold. Over the same run:

- `% is asleep` rejections: **0**, with 190 sleep-side gate actions.
- `% is dead` rejections: **27**, with only 2 dead-side gate actions.

This does not affect SC-003, which is scoped to sleep, and it is not a
regression — the dead path simply never received the equivalent coverage.
Flagged as a candidate follow-on rather than silently absorbed.

## What remains

- [x] The live ≥ 3-game-day soak itself — run 2026-08-01/02 at 12.005 game-days.
- [x] Counts from queries 1–3 recorded on the board task (SC-003).
