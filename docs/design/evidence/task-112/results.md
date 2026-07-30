# TASK-112 (spec 102) — guardian agentization: soak + ceiling evidence

Run completed 2026-07-30. Worlds preserved under `~/.promptworld/measure/`
(`task-112-soak-{control,default,authored}`) — never deleted, re-queryable.

## Protocol

Three same-seed worlds (seed 1337, stage-4 override, the branch binary),
all-local LLM, $0 spend — villagers, guardian, angel, consolidation on
`cogito:3b` @ `localhost:11434`; embeddings on `all-minilm`; no cloud, and
neither the running TASK-164 measurement endpoints nor the operator
playtest were touched:

| arm | tuning | charter | final tick |
|---|---|---|---|
| control | none (agentization OFF) | default | 199,829 |
| default | `angel_cadence_ticks: 3600` | default (ceiling ON) | 199,762 |
| authored | `angel_cadence_ticks: 3600` | task-137 authored (ceiling LIFTED) | 199,805 |

Calibrated once (`cogito:3b` ≈ 0.26 s/pt), speed 16x, run concurrently to
**2.31 game days** each (target ≥ 2 days = 172,800 ticks; all exceeded).
Zero `clock.governor_shed`, zero router suppressions in any arm — the lane
ran inside its budget the whole way. All numbers below are pure SQL over
each preserved `world.db` (queries at the end). Note: the soak binary
predates the final suppression-detach commit (a telemetry-latency fix; no
suppression ever fired in this run, so behavior is identical).

## SC-001 — the lane runs, within budget

| arm | angel thoughts | adapted (quiet) | landed (acted) | unusable | suppressed |
|---|---|---|---|---|---|
| control | 0 | — | — | — | — |
| default | 55 | 54 | 0 | 1 | 0 |
| authored | 55 | 52 | 3 | 0 | 0 |

55 scheduled turns per agentized arm over 2.31 game days at a 1-game-hour
cadence — the phase-preserving schedule held for the whole run. The control
arm carries **zero** angel-class events: opt-in is real (FR-007).

## Budget safety — villager cognition latency unchanged within tolerance

Planner `cog.outcome` (landed):

| arm | landed | mean wall ms | p95 wall ms | mean staleness ticks |
|---|---|---|---|---|
| control | 687 | 7,687 | 18,351 | 1,084 |
| default | 689 | 7,903 (+2.8%) | 18,981 (+3.4%) | 1,125 (+3.8%) |
| authored | 635 | 8,034 (+4.5%) | 19,453 (+6.0%) | 1,156 (+6.6%) |

All three arms shared ONE `cogito:3b` endpoint concurrently, so cross-arm
contention is the dominant confound; even so the agentized arms sit within
~5–7% of control on every metric, no planner was ever suppressed, and no
governor shed fired. Villager cognition latency unchanged within tolerance.

## SC-002 — ceiling proof

**Default arm (ceiling ON): only ceiling-permitted autonomous actions
across the entire run.** The complete `angel-*` tool-call spectrum:

| tool | verdicts |
|---|---|
| `explain` | 204 read_ok · 22 rejected_malformed · 31 unlanded |
| `brief_myths` | 36 read_ok · 2 unlanded |
| `survey_site` | 35 read_ok · 2 rejected_malformed · 1 unlanded |

That is the modest set — `{explain, survey_site, brief_myths}` — and
NOTHING else: **zero landed acts under angel jobs, zero charge spend, zero
world-acting guardian events** beyond the three sim-seeded genesis survival
watches (present identically in the control arm). Every scheduled turn
closed `adapted` ("observed; no act taken").

**Authored arm (ceiling LIFTED): the lifted behavior, live.** The scheduled
lane attempted the full lifted roster (send_vision/send_omen/watches/
designations/canonize/work_miracle — most attempts malformed or
gate-refused on the 3B model, all recorded in the trail) and **landed 3
autonomous acts**:

- `angel-metatron-82804` — `work_miracle` `give_item` (1 meal to Ash) →
  recorded `guardian.item_granted`, charge spent through the reducer door;
- one further `work_miracle` `give_item` → the second
  `guardian.item_granted`;
- one `monitor_and_act` → `ord-180001-0` ("any villager dies during the
  current day") — a proactively placed watch.

Same seed, same model, same cadence: the ONLY difference between the arms
is the charter text the ceiling compiles from. The default arm structurally
cannot act; the authored arm demonstrably does (the AC#5 behavior delta).
One wording wrinkle for follow-up: an angel-placed watch records
`origin:"player"` (the field names the placement DOOR — the player-tool
surface — not the initiating turn; the `cog.tool_call` trail carries the
true angel attribution).

## Guardian memory + consolidation health (D5)

| arm | memory_added | memory_embedded | consolidated markers | dream outcomes |
|---|---|---|---|---|
| default | 85 | 85 | 2 (both `rejected: unparseable`) | 1 `guardian.memory_merged` |
| authored | 52 | 52 | 2 (both `rejected: unparseable`) | 0 |

Every memory got its embedding companion (the shared embedder driver,
live). Both arms' LLM consolidation nights were rejected `unparseable` —
`cogito:3b` fails the JSON reply contract at the same rate it fails the
VILLAGER night (villager `agent.consolidated` this run: 50–75% rejected per
arm), so this is model quality, not a contract defect; the marker path
recorded both nights honestly and the buffers stayed intact (the designed
degrade). The spec-098 dream phase fired INDEPENDENTLY of the failed LLM
nights, exactly per doctrine: the default arm landed a real
`guardian.memory_merged` (three near-identical notes clustered by recorded
vectors; two absorbed into the kept representative). The acceptance path is
unit-pinned (`TestGuardianNightConsolidates`).

## FR-006 — default-vs-authored outcome instrument: EVIDENCE-PENDING

The full anti-self-grading outcome-delta instrument is the TASK-137/164
paired-arm recipe (harsh dials, cloud-guardian arms, longer horizon,
probe-battery supplement — ledger queries in
`docs/design/evidence/task-163/results.md`). It runs POST-MERGE on the
agentized build as the TASK-164 follow-on. Harness prepared by this run:

- both arms boot, schedule, remember, and consolidate on the agentized
  build with only the charter differing;
- the ceiling guarantees the arms differ structurally (above), so the
  instrument measures OUTCOME delta, not whether a behavior delta exists;
- recipe deltas vs TASK-164: set `angel_cadence_ticks` on both arms; route
  `angel` head-only at the guardian's provider; add the `angel-*` job
  filter to the rejection-rate ledger queries.

## Queries (the recorded protocol)

Read-only immutable opens (`file:<db>?immutable=1`), per arm:

```sql
-- angel lane activity
SELECT json_extract(payload,'$.outcome'), COUNT(*) FROM events
 WHERE type='cog.outcome' AND json_extract(payload,'$.class')='angel' GROUP BY 1;
-- ceiling proof: every angel tool call
SELECT json_extract(payload,'$.tool'), json_extract(payload,'$.verdict'), COUNT(*)
  FROM events WHERE type='cog.tool_call'
   AND json_extract(payload,'$.job') LIKE 'angel-%' GROUP BY 1,2;
-- landed acts under angel jobs (0 on default; 3 on authored)
SELECT COUNT(*) FROM events WHERE type='cog.tool_call'
   AND json_extract(payload,'$.job') LIKE 'angel-%'
   AND json_extract(payload,'$.verdict')='landed';
-- planner latency (budget safety)
SELECT COUNT(*), ROUND(AVG(json_extract(payload,'$.actual_wall_ms'))),
       ROUND(AVG(json_extract(payload,'$.staleness_ticks')))
  FROM events WHERE type='cog.outcome'
   AND json_extract(payload,'$.class')='planner'
   AND json_extract(payload,'$.outcome')='landed';
-- store + consolidation + dream
SELECT type, COUNT(*) FROM events WHERE type LIKE 'guardian.memory%'
   OR type IN ('guardian.consolidated','guardian.salience_revised') GROUP BY type;
-- world-acting guardian events (any origin)
SELECT type, COUNT(*) FROM events WHERE type IN ('guardian.nudged',
 'guardian.time_snapped','guardian.item_granted','guardian.entity_moved',
 'guardian.entity_removed','guardian.order_placed','designation.placed',
 'directive.issued','prophecy.declared','guardian.region_named') GROUP BY type;
```
