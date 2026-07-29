# TASK-163 Guardian tool-call competence — results (runs of 2026-07-27 → 2026-07-28)

Two-stage mitigation of the 80% privileged-action rejection floor measured by
TASK-136 (8/10 rejected, gemma4:12b-mlx guardian, pre-fix binary — see
[task-137/results.md](../task-137/results.md)). Stage 1 swapped the guardian
turn's model tier; stage 2 surfaced the `give_item` grant vocabulary the
measurement showed was invisible. Both stages measured live on the TASK-137
recipe: same-seed (1337) arm pairs, stage-4 `--override`, harsh dials
(`fire_burn_per_wood=3600`, `gru_emerge_per_mille=1000`), 8x, arm B running the
authored charter ([task-137/authored-charter.md](../task-137/authored-charter.md)
verbatim), every non-guardian route on `gemma4:12b-mlx` as in the baseline.
Guardian turn route (`metatron`): `cc/claude-sonnet-5` via the 9router proxy,
head-only chain (a proxy failure skips the turn rather than falling back to
gemma and contaminating the sample).

## Headline

| Condition (binary) | Privileged attempts | Landed | Rejected | Rate |
|---|---|---|---|---|
| Baseline: gemma guardian, pre-fix (TASK-136/137 run, `072dd71`) | 10 | 2 | 8 | **80%** |
| Stage 1: sonnet guardian, pre-fix binary (`89c78b9`, 3 game-days) | 7 | 2 | 5 | **71%** |
| Stage 2: sonnet guardian, vocabulary surfaced (`2eef530`, 3 game-days + probe battery) | 14 | 9 | 5 | **36%** |

The stage-1 headline barely moved — but its taxonomy flipped entirely, which is
the finding that drove stage 2:

- **Stage 1 (tier only):** `send_vision` went 2/2 landed (baseline: 0/3 — the
  missing/hallucinated-coordinate class died with the model swap). ALL five
  remaining rejections were one signature: `work_miracle give_item` with a
  guessed item kind (`"food"` ×4, `"forage"` ×1). The grant vocabulary
  (`grantableKind`, internal/sim/miracles.go) appeared nowhere in the tool
  schema, the turn guidance, or the rejection message — a competent model
  could only guess. Two villagers (Ash, Sage) starved while the guardian fed
  synonyms to a door that wouldn't name what it wanted.
- **Stage 2 (tier + vocabulary):** zero vocabulary failures, zero hallucinated
  coordinates, zero malformed arguments — the baseline's entire model-side
  taxonomy is extinct. The five remaining rejections are two world-mechanics
  classes:
  - **Carry-cap quantity bounds ×2** — correct item kinds (`food_cooked`,
    `meals`), quantity over the villager's carry capacity. The door's message
    carries the numbers; in one case the guardian repaired to a valid quantity
    and landed it in the same turn, in the other it explained the constraint
    to the player accurately instead of blind-retrying. The door is now
    teaching, and the model is learning from it mid-turn.
  - **Position-freshness races ×3** — well-formed `move` calls whose surveyed
    source coordinates staled while the model was thinking (at 8x, ~30–60s of
    latency is 240–480 ticks of villager walking). An architectural property
    of privileged moves at speed, not a competence or guidance gap. One
    autonomous recovery observed: after a raced villager-move rejection, the
    guardian moved the STRUCTURE to the villager instead — landed.

## Stage-2 sample composition (labeled strata)

Autonomous (3 game-days, harsh dials, watches only — no player contact):

| Arm | Attempts | Landed | Rejected | Deaths |
|---|---|---|---|---|
| default | 2 (`send_vision` ×2: Birch exposure t73,620; Sage exposure t245,340) | 2 | 0 | **0** |
| authored | 2 (`move` villager raced; `move` structure-to-villager landed, t254,340) | 1 | 1 | **0** |

Prompted probe battery (post-window, 5 standardized asks per arm through the
normal `promptworld guardian` chat surface — same tool schema, same validity
door; run because three harsh-dial days produced only 4 autonomous attempts):

| Arm | Attempts | Landed | Rejected |
|---|---|---|---|
| default | 5 | 3 (`give_item food_cooked` qty-repaired, `give_item spear`, `send_vision`) | 2 (cap ×1, race ×1) |
| authored | 5 | 3 (`give_item spear`, `give_item wood`, `send_vision`) | 2 (cap ×1, race ×1) |

One default-arm ask (wood grant at 0/3 charges) drew a charge-aware decline
with no tool call — correctly excluded from the validity denominator. Charge
economy held throughout: every landed miracle spent its charge; nothing landed
free.

## Survival note (n=1, not a claim)

The fixed run lost **zero villagers** across both arms in 3 harsh-dial
game-days. The tier-only run on the same seed and dials lost two (both
starvation, both villagers the guardian had correctly targeted with grants the
door refused on vocabulary). At n=1 per condition this is an anecdote, but it
is the mechanism TASK-111's survival lane exists for, now observed end-to-end:
watch fires → guardian targets → call lands → nobody dies.

## Caveats

- Small n throughout; one seed; one model per stage. Falsifiability probes,
  not studies.
- The probe battery is player-prompted (ambition lane) rather than
  watch-triggered (survival lane): turn context differs, but the tool schema
  and validity door — the surfaces under measurement — are identical.
- Stage-1 run day 1 shared the gemma endpoint with an active playtest world
  (calibration 9.9 s/pt vs baseline 9.3); guardian calls ride the 9router
  route and were unaffected.
- Binary drift across stages: baseline `072dd71`, stage 1 `89c78b9` (spec 086
  payload census landed between — plumbing, not guardian prompting), stage 2
  `2eef530` (the vocabulary fix itself). Stage 2's autonomous window closes at
  tick 259,200; battery calls run after it on the same still-live worlds.
- The 8x position-race class would shrink at lower speeds; at 1x the
  survey-to-call gap is ~8 ticks, not ~400.

## Feeds

- **TASK-163 AC#2**: satisfied — ≥10 attempts (14), rate 36%, materially below
  the 80% baseline, evidence recorded here.
- **TASK-163 AC#3 / TASK-136 / TASK-137 confound bound**: the quoted
  mechanical-noise bound for prompt-attribution comparisons (TASK-137 charter
  delta, TASK-67 fork duel) is superseded: on the fixed binary with a sonnet
  guardian, **model-side tool-call noise is ~0%** (0/14 vocabulary, grammar, or
  grounding-hallucination failures); residual rejections are world-mechanics
  bounds (carry cap — self-repairing) and position races (speed-dependent).
  Addendum added to [task-137/results.md](../task-137/results.md).
- **TASK-164 (charter-delta outcome re-run)**: the competence floor that
  confounded TASK-137's outcome measurement is cleared; the re-run can now
  attribute outcomes to the charter surface. Note the guardian model tier is a
  new experiment variable to hold fixed.
- **TASK-112 dispatch checkpoint** (guardian-directives runbook): the
  tool-competence evidence this checkpoint gated on now exists.
- Residual classes worth their own consideration (not carded here): a
  freshness token for move targets (re-resolve the entity by name at the door
  rather than by stale coordinates), and quantity guidance (surface the carry
  cap or clamp-with-note semantics).

## Raw evidence

Worlds preserved (stopped): `~/.promptworld/measure/task-163-{default,authored}`
(stage 1) and `~/.promptworld/measure/task-163-fixed-{default,authored}`
(stage 2). Ledger query, per world:

```sql
SELECT json_extract(payload,'$.snapshot_tick'), json_extract(payload,'$.tool'),
       json_extract(payload,'$.verdict'), json_extract(payload,'$.args'),
       json_extract(payload,'$.reason')
FROM events
WHERE type='cog.tool_call'
  AND json_extract(payload,'$.tier')='niner'
  AND json_extract(payload,'$.tool') IN ('send_vision','work_miracle')
ORDER BY seq;
```
