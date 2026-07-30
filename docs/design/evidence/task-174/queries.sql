-- spec 103/TASK-174 D6 metrics — computed from cog.outcome events alone,
-- against a world's world.db (readable live, WAL mode, while the daemon
-- runs: `sqlite3 -readonly <world>/world.db < queries.sql`).
--
-- Contract (spec 011/TASK-42, restated in spec 103 D6): OutcomeRetried is a
-- NON-TERMINAL marker — a scene that consumed its one retry and continued.
-- Every other cog.outcome value for a job is terminal. Transport/admission
-- errors are never retried, so a `cog.outcome{retried}` whose reason carries
-- the "outcome: " prefix is exactly one outcome-call parse failure; a
-- terminal `cog.outcome{unusable}` whose reason carries "outcome: " and a
-- non-empty raw is a parse-killed scene (a transport-abandoned scene carries
-- no raw at all).

-- 1. Outcome parse-failure count: retried markers with the outcome-call
--    reason prefix, plus parse-killed terminal unusable events.
SELECT
  (SELECT COUNT(*) FROM events
     WHERE type = 'cog.outcome'
       AND json_extract(payload, '$.class') = 'conversation'
       AND json_extract(payload, '$.outcome') = 'retried'
       AND json_extract(payload, '$.reason') LIKE 'outcome: %')
  +
  (SELECT COUNT(*) FROM events
     WHERE type = 'cog.outcome'
       AND json_extract(payload, '$.class') = 'conversation'
       AND json_extract(payload, '$.outcome') = 'unusable'
       AND json_extract(payload, '$.reason') LIKE 'outcome: %'
       AND COALESCE(json_extract(payload, '$.raw'), '') != '')
  AS outcome_parse_failures;

-- 2. Founded scenes: terminal cog.outcome events (job prefix
--    "conversation-", excluding the non-terminal "retried" marker) — one row
--    per scene that was admitted past the router/novelty gates.
SELECT COUNT(*) AS founded_scenes
FROM events
WHERE type = 'cog.outcome'
  AND json_extract(payload, '$.class') = 'conversation'
  AND json_extract(payload, '$.job') LIKE 'conversation-%'
  AND json_extract(payload, '$.outcome') != 'retried';

-- 3. Abandoned scenes: terminal unusable conversation-job events (dies to
--    either a parse failure with no retry left, or a transport error, or the
--    TASK-32 staleness pre-abort lands separately as rejected-stale and is
--    NOT counted as abandoned-by-outcome-parse-failure here).
SELECT COUNT(*) AS abandoned_scenes
FROM events
WHERE type = 'cog.outcome'
  AND json_extract(payload, '$.class') = 'conversation'
  AND json_extract(payload, '$.job') LIKE 'conversation-%'
  AND json_extract(payload, '$.outcome') = 'unusable';

-- 4. Abandoned-scene rate = (3) / (2), reported as a percentage the same way
--    the playtest-1 baseline is (62/293 = 21%).
SELECT
  (SELECT COUNT(*) FROM events
     WHERE type = 'cog.outcome' AND json_extract(payload, '$.class') = 'conversation'
       AND json_extract(payload, '$.job') LIKE 'conversation-%'
       AND json_extract(payload, '$.outcome') = 'unusable') AS abandoned,
  (SELECT COUNT(*) FROM events
     WHERE type = 'cog.outcome' AND json_extract(payload, '$.class') = 'conversation'
       AND json_extract(payload, '$.job') LIKE 'conversation-%'
       AND json_extract(payload, '$.outcome') != 'retried') AS founded,
  ROUND(100.0 *
    (SELECT COUNT(*) FROM events
       WHERE type = 'cog.outcome' AND json_extract(payload, '$.class') = 'conversation'
         AND json_extract(payload, '$.job') LIKE 'conversation-%'
         AND json_extract(payload, '$.outcome') = 'unusable') /
    NULLIF((SELECT COUNT(*) FROM events
       WHERE type = 'cog.outcome' AND json_extract(payload, '$.class') = 'conversation'
         AND json_extract(payload, '$.job') LIKE 'conversation-%'
         AND json_extract(payload, '$.outcome') != 'retried'), 0), 1)
  AS abandoned_pct;

-- 5. Breakdown by terminal outcome value, for a sanity check alongside (2)-(4).
SELECT json_extract(payload, '$.outcome') AS outcome, COUNT(*) AS n
FROM events
WHERE type = 'cog.outcome'
  AND json_extract(payload, '$.class') = 'conversation'
  AND json_extract(payload, '$.job') LIKE 'conversation-%'
GROUP BY outcome
ORDER BY n DESC;
