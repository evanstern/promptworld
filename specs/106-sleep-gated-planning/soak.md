# T007 — Soak evidence: measurement queries (SC-003)

**Status: queries implemented + documented; the live soak run itself remains**
(see "What remains" below). The soak needs a real local planner model serving
multi-game-day traffic — the measurement-run recipe (paired seed worlds, harsh
dials, the operator's local router for LLM routes) — which is an operator
measurement run, not something a CI-shaped environment can produce
meaningfully. Nothing here is fabricated: the numbers stay blank until a real
run fills them.

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

## What remains

- [ ] The live ≥ 3-game-day soak itself (real planner model, measurement
      dials), run by the operator/orchestrator.
- [ ] Counts from queries 1–3 recorded on the board task (SC-003).
