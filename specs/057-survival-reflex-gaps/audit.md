# Reflex-layer survival audit (spec 057 / TASK-108, US4)

**Verified against:** `internal/sim/policy.go` `decideIntent` (the reflex ladder)
and `internal/sim/executor.go` `wakeReason` (the sleep/wake gate that decides
when the ladder runs at all), on branch `task-108-survival-reflex`.

**Doctrine under audit** (control-surface report §7, 2026-07-24): *"Survival is
table stakes: villagers are simple people who at minimum know how to not die.
Err toward survival instinct wherever the reflex layer has a gap."* Applied here
to **fire/warmth**; eat and sleep are audited for parity but their gaps are
carded, not patched (this spec's boundary is fire — TASK-103 owns arbitration,
TASK-104 owns recovery semantics).

**Method.** Every rung of `decideIntent` is walked in ladder order. For each: the
need it protects, the thresholds it keys on, and a gap disposition — **no gap**,
**fixed here**, or **carded** (with the owning board task). The cold build-fire
rungs are additionally proven cell-by-cell, reflex-only, by
`TestColdNightReflexMatrix` (`internal/sim/reflex_matrix_test.go`, SC-003).

## Thresholds referenced (all `internal/sim/agents.go`)

| Constant | Value | Meaning |
|---|---|---|
| `hungryAt` | 350 | hunger arms the eat / get-food rungs |
| `satietyAt` | 900 | eating stops here |
| `tiredAt` | 250 | daytime nap threshold |
| `coldNightBelow` | 350 | cold-night warmth effect threshold |
| `nearDeathBelow` / `nearDeathResetAt` | 200 / 400 | near-death latch |
| `fireWoodCost` | 2 | wood to build a fire |
| `fireWarmRadius` | 2 (Manhattan) | warmth range of a lit fire |
| `stockFoodRawTo` | 8 | opportunistic larder target |
| `RefuelDyingBelow` | 10800 (was 3600) | reflex refuel trigger window — **raised in this spec** |

## Ladder audit (in `decideIntent` order)

| # | Rung (condition → intent) | Need protected | Thresholds | Disposition |
|---|---|---|---|---|
| 1 | `Food < hungryAt && hasAnyFood` → `agent.ate` | eat (from hand) | `hungryAt` 350, `satietyAt` 900 | **no gap** — eats the moment hunger bites if carrying any food form |
| 2 | `Food < hungryAt` → forage/hunt known, else `search` frontier | eat (acquire) | `hungryAt` 350 | **no gap** — knowledge-gated with a frontier-search fallback, so ignorance turns into exploration, not omniscience (spec 041 US4). This is the one rung that already searches when it knows nothing. |
| 3a | night & `!warmAt` & known warmth reachable → `goto_warmth` | warmth (reach) | `fireWarmRadius` 2 | **no gap** — walks to a fire it remembers as lit (matrix col 1) |
| 3b | night & `!warmAt` & known cold/dying fire & `Wood≥1` → `refuel_fire` | warmth (relight) | `RefuelDyingBelow` 10800 | **no gap; improved here** — relights the fire it remembers rather than building a second; the raised window arms it 3× earlier. Closes the US3 "stale warmth belief" candidate gap (matrix col 3). |
| 3c | night & `!warmAt` & `Wood≥fireWoodCost` & build site → `build_fire` | warmth (make) | `fireWoodCost` 2 | **no gap** — the cold build-fire reflex the task asked to ensure; proven by matrix col 2 / wood≥2 |
| 3d | night & `!warmAt` & known tree → `chop` | warmth (get fuel) | — | **no gap** — wood-acquisition rung; proven by matrix col 2 / wood=1+tree |
| 3e | night, otherwise → `sleep` | rest | — | **gap (carded)** — a cold villager with `<2` wood, no known warmth, and no known tree sleeps in place. There is no night search-for-warmth/wood fallback (unlike rung 2's frontier search). See Gap A. |
| 4 | day & `Rest < tiredAt` → `sleep` (at known warmth if not already warm) | rest (+ warmth parity) | `tiredAt` 250 | **partial gap (carded)** — naps toward known warmth, but only when *tired*. A cold-but-not-tired villager has no day-time warmth-seeking rung. See Gap B (the TASK-103 day-branch warmth gap). |
| 5 | day & `!knowsAnyFresh("fire")` → `build_fire` (wood≥2) else `chop` | village has a fire | `fireWoodCost` 2 | **no gap** — prepares the village fire from the agent's own belief (spec 041) |
| 6 | day → `reflexRefuelIntent` (known cold/dying fire + wood) | keep the fire alive | `RefuelDyingBelow` 10800 | **no gap; improved here** — the raised window is the core burnout fix (world-01: 42 burnouts vs 8 builds) |
| 7 | day & `FoodRaw < stockFoodRawTo` → forage/hunt known (no search) | larder | `stockFoodRawTo` 8 | **no gap by design** — opportunistic top-up only; deliberately does NOT mount frontier expeditions (constant treks desync the forage-regrow rotation, spec 041 note) |
| 8 | day, nothing urgent → `wander` | — | — | **no gap** — idle filler |

### Sleep/wake gate (`wakeReason`, executor.go) — audited for parity

`wakeReason` wakes a sleeper only on (a) daybreak with `Rest ≥ 600`, or (b) a
hunger emergency it can act on (`Food < 150` **and** food in hand). It does
**not** wake a villager to worsening **cold** or **health**. So a villager that
enters rung 3e's `sleep` while cold will not be re-woken by the warmth/health
decay itself. This is the US3 "wake/sleep boundary" candidate gap → Gap C.

## Gaps & dispositions

- **Gap A — no night search-for-warmth/wood fallback (rung 3e).** *Carded, not
  fixed.* Fixing it means sending cold villagers wandering the night to look for
  wood or warmth — a behavior change that trades exposure risk for gru-exposure
  and forage-rotation risk, and it overlaps the reflex/planner arbitration
  design. Out of the fire-adjacent *surgical* boundary. **Owner: TASK-103**
  (reflex/planner arbitration). Recommend a dedicated follow-on card:
  "night cold: search-for-warmth/wood fallback vs. sleep-in-place".
- **Gap B — day branch never proactively seeks warmth (rung 4).** *Carded, not
  fixed.* The daytime ladder only moves toward warmth when *napping*; a cold,
  rested villager during the day has no warmth-seeking rung, so daytime warmth
  decay is unaddressed until night. This is exactly the arbitration gap the task
  notes name. **Owner: TASK-103** ("policy.go day branch never considers
  warmth"). The code in this spec deliberately does not preempt that design.
- **Gap C — sleepers don't wake to cold (`wakeReason`).** *Carded, not fixed.*
  The wake gate ignores warmth/health emergencies to avoid 4am sleep/wake churn;
  adding a cold/health wake condition is a recovery-semantics change touching
  that churn-avoidance design. **Owner: TASK-104** (recovery semantics).
  Recommend a follow-on card: "wake a sleeper on a warmth/health emergency
  (churn-bounded)".

## Coverage & outcome

- **Rung coverage: 100%** — every `decideIntent` rung above has a need,
  thresholds, and a disposition; the sleep/wake gate is audited for parity.
- **Fixed in this spec:** the refuel window (rungs 3b/6) — `RefuelDyingBelow`
  3600 → 10800 (US1). No other reflex rung was changed: the matrix proved the
  cold build-fire ladder already produces the doctrine outcome for all nine
  cells, so US3 required verification, not patching.
- **Carded (with owners):** Gap A → TASK-103; Gap B → TASK-103; Gap C →
  TASK-104. These are survival gaps outside the fire-adjacent surgical boundary;
  they seed the next survival tasks with evidence rather than anecdotes.

> Board note: Gaps A and C recommend *new* follow-on cards under their owning
> tasks; creating those board entries is orchestrator/planning-tier work (this
> branch does not edit `backlog/`). This audit is the durable evidence for them.
