# TASK-164 — charter-delta outcome re-run (post tool-competence floor)

**Status: COMPLETE** — run 2026-07-29 → 2026-07-31 by the board-sweep-2026-07-29
orchestrator under the operator-approved design (n=1 same-seed pair, AC#1
checkpoint 2026-07-29) with the recorded horizon amendment (2026-07-30): arm A
over-ran while the session idled, and the strict 3-game-day window proved
vacuous under the harsh dials (zero privileged attempts before tick 259,200),
so the pair was matched at arm A's actual horizon — **tick 498,187 ≈ 6.76
game-days** — the evidence doc's own "longer horizon" alternative.

## Design

TASK-137/163 recipe on current main (post 163/166/167 door fixes + 134 rename):
seed 1337, stage-4 `--override`, harsh dials (`fire_burn_per_wood=3600`,
`gru_emerge_per_mille=1000`), 8x, villagers gemma4:12b-mlx, guardian route
`niner` = cc/claude-sonnet-5 via 9router head-only. Arm A = compiled default
charter; arm B = `docs/design/evidence/task-137/authored-charter.md` verbatim.
Sequential arms. Worlds preserved (stopped, never deleted):
`~/.promptworld/measure/task-164-default`, `~/.promptworld/measure/task-164-authored`.
NOTE: pre-102 binary lineage for arm A start; both arms predate spec 102's
steward lane (no scheduled cognition — watch/chat-triggered turns only), so
this measures the CHARTER surface on the event-driven guardian, the direct
TASK-137 comparison.

## Headline (both arms, tick ≤ 498,187)

| | Arm A — default | Arm B — authored |
|---|---|---|
| Privileged attempts (work_miracle) | 4 | 4 |
| Landed / rejected | **4 / 0** | **3 / 1** (gate) |
| Reads (survey / explain) | 10 / 4 | 8 / 16 |
| Other acts | — | 4 omens, 1 vision, 3 nudges |
| Orders placed / triggered | 3 / 0 | 5 / 2 |
| Report cards | 0 | 2 |
| **Deaths** | **1** (starvation, t411,180 ≈ day 5.8) | **4** (starvation t79,980; exposure t85,800; gru t147,825; gru t233,005) |

## Findings

1. **Tool competence floor: FIXED, confirmed at horizon.** 7/8 privileged
   miracle attempts landed across both arms (87.5%) vs TASK-136's 20% baseline
   and TASK-163's 64%. Zero position-freshness rejections (TASK-166's class
   extinct in this sample), zero carry-cap rejections (TASK-167's class
   extinct). The single rejection was a stage-gate refusal. Outcome
   attribution is no longer confounded by tool fumbling — the card's purpose
   for this re-run is achieved.
2. **Charter → behavior delta: CONFIRMED, large.** The authored charter
   produced a 4x richer act stream (omens/visions/nudges/reports absent
   entirely from the default arm) and 4x the explain traffic — consistent
   with TASK-137's behavior finding, now on a competent door.
3. **Charter → survival outcome delta at n=1: NEGATIVE (authored arm worse).**
   1 death (default) vs 4 (authored), same seed, same dials. This is the
   honest headline: at n=1 with `gru_emerge_per_mille=1000`, gru encounters
   are highly sensitive to LLM-timing divergence between arms, and two of the
   four authored-arm deaths are gru kills; no causal attribution is licensed.
   What the data DOES rule out: the hypothesis that the authored charter
   straightforwardly improves survival outcomes on this recipe. TASK-111
   AC#5's outcome-delta caveat should record: behavior delta proven twice
   (137, 164); outcome delta remains UNPROVEN in either direction at n=1,
   with the 164 sample directionally against.
4. **Recommended next instrument** (feeds TASK-112 FR-006, harness prepared in
   spec 102's evidence doc, + spec 107's mission scenario): n≥3 seeds, the
   agentized (steward-lane) build, mission-scenario scoring in addition to
   survival — outcome variance at these dials swamps n=1.

## Raw evidence

Ledger queries (per arm, `world.db`): privileged calls —
`SELECT tool,verdict,COUNT(*) FROM events WHERE type='cog.tool_call' AND tier='niner' AND tick<=498187 GROUP BY 1,2`
(via json_extract); deaths — `agent.died` rows; acts — `guardian.%` type
counts. Full tables above; worlds preserved for re-query.

## Feeds

- **TASK-164 AC#2**: both arms re-run post-163 at a matched horizon under the
  operator-approved amended design — SATISFIED.
- **TASK-164 AC#3**: this doc + TASK-111 AC#5 caveat update — SATISFIED.
- **TASK-112 FR-006 (EVIDENCE-PENDING)**: this run predates the steward lane;
  the agentized-arm instrument is the recommended follow-on (finding 4).
