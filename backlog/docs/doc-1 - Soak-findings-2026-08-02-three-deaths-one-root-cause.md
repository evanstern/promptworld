---
id: doc-1
title: 'Soak findings 2026-08-02 - three deaths, one root cause'
type: specification
created_date: '2026-08-03 01:45'
updated_date: '2026-08-03 01:45'
---
**Status:** findings only — nothing carded, no code changed. For another agent to review, challenge, and (if it holds up) turn into board tasks.

**Full report (charts, tables, evidence):** https://claude.ai/code/artifact/3405ff25-cf1f-43dd-99de-c410f80e34bb

## What this is

A read of the last two overnight soak runs, looking for bugs, tuning problems, behaviours worth making reflexive, and behaviours worth guarding against. Analysis only — the root checkout was never modified and no worktree was cut.

| Run | Span | Scale | Deaths |
|---|---|---|---|
| `soak-world` (Run A, cloud-tier model) | 12.0 game-days | 132,299 events | 3 |
| `soak-qwen` (Run B, local qwen) | 5.9 game-days, still running at time of writing | ~95,000 events | 0 |

Run B's zero deaths are **not** an improvement — 80% of its decisions are `wander` and it discards 73% of its nightly consolidation. Different models, so the two runs are not a difficulty comparison.

## The headline

All three deaths trace to engineering defects rather than difficulty, and the same class of defect disables the rescue system built to prevent them. The guardian went **0 for 2** on villagers who were actively dying.

## P0 — one root cause, three faces

The shared assumption: *a villager who is busy, asleep, or without a precomputed option does not need a turn.*

### 1. A villager can be captured by an intent and never get another turn

Sage's last `agent.intent_set` was `warm_up` at tick 173,131. Sage died at 216,780 — **43,649 ticks later** — with zero intents, paths, work, or eats in between, awake the whole time (no `agent.slept`, rest 0 at death).

`survivalDecision`'s first rung is *eat from inventory the moment hunger bites* (`internal/sim/policy.go`). It never ran, because `decideIntent` is only consulted when the agent is idle. Sage starved holding 5 `food_cooked`.

`agent.recovery_stalled` fired **once** in the whole run — for Fern, same goal (`warm_up`), and Fern lived. So the stall detector exists and missed this.

### 2. The wake gate is inverted triage

`wakeReason` (`internal/sim/executor.go:1847`) wakes a sleeper only for an emergency they can *already* act on: hunger requires `hasAnyFood()`, cold requires `warmthLadder()` to return actionable. Intent was to bound sleep/wake churn.

Effect: villagers who can save themselves get woken; those who cannot are left asleep. It also bypasses the mind layer entirely — a woken villager could forage, search, or seek help.

Rowan slept from tick 230,460 and **never woke**; health 1000 → 0, food hit 0 mid-sleep, died 1,620 ticks before dawn.

### 3. Rescues cannot land

- **Sage** — `guardian.item_granted` 5 `food_cooked` at 198,351. Already frozen since 173,131. Never ate. Died 216,780.
- **Rowan** — `guardian.place_revealed` fire at (54,34) at 238,598, while asleep. A vision does not wake a sleeper.

Rowan's cascade, which is the clearest single artifact in the whole analysis:

| tick | event |
|---|---|
| 230,460 | Rowan sleeps, warmth 596 |
| 238,598 | Guardian reveals fire (54,34) — lit, `fuel_until` 241,298 |
| 241,298 | **That fire burns out.** Nobody refuels it |
| 243,600 | Warmth crosses `exposureWakeBelow` (150). Wake gate finds no actionable remedy → stays asleep. **Health 697** |
| 257,580 | Dies |

The rescue missed by 2,302 ticks (~38 game-minutes). Nothing re-checks that a delivered remedy still exists.

## P1 — missing capabilities

- **No way to run away.** No `flee` / `hide` / `defend` / `take shelter` anywhere in the goal vocabulary (grepped). Birch took 5 gru strikes over 3,048 ticks (750→521→281→41→0) and never moved. After the first strike his next act was `goal:"chop"`, `source:"reflex"` — the warmth ladder, chopping firewood while being eaten.
- **Nothing can outpace deliberation.** Gru attacks every 600 ticks; `planner_cadence_ticks` is 1,800. Any threat faster than the planning cycle is unopposable by construction.
- **Consolidation voids whole nights over one prose field.** `validate.go:153` fails `bad_narrative` when the narrative is empty or >1,200 chars, discarding every promotion, fade and belief edit for that night. Run B: `bad_narrative` 26 of 48 nights (54%); only 13 accepted (27%) vs 67/74 (91%) in Run A. The same file already keeps a best-first prefix for over-long lists, calling them "enthusiasm, not corruption" — the text field just never got that treatment.
- **Preconditions unchecked before proposing.** 735 of 1,030 plan failures (71%) are precondition-class. Top: `you know of no warm place` 194, `lacks inputs for craft_planks` 120, `no such agent to seek` 75, stale-generation discards 50, `<name> is dead` 27.

## P2 — tuning and self-defeating systems

- **Fires:** 15 built, 322 refuelled, **144 burned out**. Root cause of the 194 no-warm-place failures and of Rowan's dead hearth.
- **Wander is 72.5% of all decisions** in Run A (14,340 of 19,779), 80% in Run B. It is `wanderDecision`, the terminal idle filler.
- **Governance oscillates.** 17 motions: curfew added 8× (passed 4× narrowly — 4–1, 5–1, 4–2, 3–2), repealed 3× (passed **unanimously** every time — 6–0, 6–0, 4–0). `add_repay_debts` tabled 5× with identical text by the same villager, failed 3–3 every time. Ties fail; nothing blocks re-tabling a defeated motion unchanged.
- **Nobody shares.** 2 `social.gave` in 12 days (both Birch→Sage, both wasted on a villager too stuck to eat). 15 deposits vs 2 withdrawals. One survivor ended holding 16 cooked meals.
- **Memory unbounded.** 2,300–3,243 memories per villager and climbing, against 536 fades run-wide; 15,343 `place_observed` at `observation_base_salience: 2` mints noise like "Looked around: standing trees".
- **Estimator pessimism.** Daemon logs `world is UNCALIBRATED`; observed `predicted_wall_ms` 15,341 vs `actual_wall_ms` 8,296. Over-prediction drives `clock.governor_shed` (×9) and 103 suppressed planner calls.

## Methodological correction — please keep this one

An earlier draft of this analysis claimed *"knowledge of the map predicts survival"* — dead villagers knew 6/7/14 places, survivors 25–41. **The operator challenged it as a survivorship artifact and was right. The claim is withdrawn.**

It was wrong twice: (1) it read the wrong field — snapshot `agents[].known` is the villager's **rumours**, not places; real map knowledge is `agents[].map.facts`; (2) on the correct field the pattern collapses — Birch died with 901 map facts, **more than three of the five survivors** (Oak 817, Fern 822, Cedar 896). Normalising by days lived reverses it: the dead explored at 90–133 facts/day, survivors at 68–81, because exploration is fast early and saturates.

Treat any other correlation in here that scales with lifetime with the same suspicion. The findings above are event-mechanism claims (specific ticks, specific code paths), not correlations, which is why I have more confidence in them.

## Evidence — and a warning that it is perishable

Raw runs live under another job's scratch dir:

- `/Users/evanstern/.claude/jobs/ca35de11/tmp/soak/soak-world/` — `world.db`, `daemon.log`, `chronicle.md`, `morgue.md`
- `/Users/evanstern/.claude/jobs/ca35de11/tmp/soak/soak-qwen/` — same, daemon may still be running (PID in `daemon.pid`)
- `progress.log`, `progress-qwen.log` — 15-minute sampled counters

**These are deleted when that job is cleaned up.** If any of this is to be reproducible, the two `world.db` files should be copied somewhere durable first. Open them read-only (`file:...?mode=ro`) — Run B's daemon may still be writing.

All counts above came from SQL over the append-only `events` table; code claims were read in the source.

## What I'd want a reviewer to challenge

1. **Is the captured-intent bug really one bug?** I inferred "the intent never completed" from the absence of `agent.intent_done` after 173,130. A reviewer should confirm against the executor whether `warm_up` has a completion or timeout path at all, and why `recovery_stalled` caught Fern but not Sage.
2. **Is the wake gate's actionable guard load-bearing?** It was added to bound churn. Removing it may reintroduce sleep/wake thrash — worth measuring rather than assuming.
3. **Does `guardian.place_revealed` interact correctly with fire freshness?** The reveal stamps `Seen` at landing and `Detail` from ground truth, so a revealed fire that later burns out becomes silently useless. Should a rescue re-verify, or should the guardian refuel instead of pointing?
4. **Sampling.** One 12-day run and one partial run, two different models, one seed each. Nothing here is established across seeds.

## Proposed cards — NOT created

Deliberately not carded: the board was contended by concurrent sessions when this was written, and the operator asked to hold. Suggested shape if a reviewer agrees:

| Priority | Card |
|---|---|
| P0 | Intent deadline + hand the turn back to the reflex ladder on expiry |
| P0 | Wake on the emergency band, not on a precomputed remedy |
| P0 | Rescue delivery: vision wakes a sleeper, grant is evaluated on arrival, remedy re-verified |
| P1 | Add a flee action and a predator-response reflex |
| P1 | Consolidation: clamp narrative/gist instead of rejecting the night |
| P1 | Precondition checking before a plan step is proposed |
| P2 | Timing constants: planner cadence vs attack interval, `exposureWakeBelow`, fire burn window |
| P2 | Stop minting memories for already-seen terrain |
| P2 | Break tied votes; block unchanged re-tabling of a defeated motion |
