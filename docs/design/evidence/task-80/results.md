# TASK-80 perception of absence — soak evidence (US3, SC-002)

One-game-day soak of the grounded arrival-observation channel (spec 097) on a
fresh seeded MEASURE world — never the operator's playtest world.

## Setup

- World: `~/.promptworld/measure/task-80-soak` — `promptworld new --seed 1337
  --stage stage-4 --override`, default dials (no `tuning.json`), fresh genesis.
- Binary: branch commit `46fe6024` (all spec-097 sim behavior: D1 arrival
  guard, radius dedup, dials; spec 098 merged into the branch afterward is
  consolidation-side only and does not touch the emission path).
- LLM: LOCAL ONLY — every route (planner, conversation, consolidation,
  narrator, guardian) on `gemma4:12b-mlx` at `mbpro-m1.local:11434`; no
  embedding route (memories vectorless); zero paid spend (`spent_usd: 0`).
- Run: tick 1 → 86,934 (one full game-day + shutdown tail) at a steady 8x
  (`effective_rate` 8.0 throughout, never degraded), 2026-07-29.

## Observation-channel volume (bounded per D4)

| Metric | Value |
|---|---|
| `agent.place_observed` events (whole day) | 1,959 |
| Companion Origin-`observed` memories | 1,959 (exactly 1:1 — both-or-neither dedup held) |
| `agent.moved` events | 11,398 — observations are 17% of moves (the pre-fix per-step flood was 100%) |
| `agent.intent_done` | 2,097 — observations ≈ intent-completing arrivals, minus dedup and zero-distance intents |
| Per agent | 202–289 (mean ≈ 245) per game-day |
| By 6-game-hour bucket | 829 / 673 / 440 / 4 / 13 — activity-shaped; sleep hours are near-zero |
| Salience of every observation memory | 2 (the `observation_base_salience` dial), uniformly |

Boundedness reading: the event rate tracks deliberate arrivals (17% of
movement, zero while asleep), never steps; the dedup collapses same-place
repeats inside the 2-game-hour window (bucket 0 is the exploration burst on a
fresh world — everywhere is a first visit — and volume falls monotonically as
maps fill in). The working-memory WINDOW is bounded by construction (top-K =
10, and salience-2 entries decay below every ordinary event tier within a
day), which is US3's criterion.

**Flagged (state growth, not window flooding):** the memory LIST still grows
by ~245 observation memories per agent per active day; nightly consolidation
faded only 56 memories across 7 agent-nights this run. The window never
floods, but long-run state/snapshot size will creep. Candidate follow-up: an
age-based auto-fade for below-floor Origin-`observed` memories (or folding
them into spec 098's habituation pass).

## Survival behavior

| Metric | Value |
|---|---|
| Deaths | 1 — Fern, starvation, tick 84,780 (04:33 night 1) |
| Fires/structures built | 8 |
| Forages | 108 |
| Chops | 62 |
| Eats | 28 |
| Consolidations | 7 accepted markers, 14 beliefs formed, 56 memories faded |

The observation channel is additive by construction — it changes no needs, no
movement, no reflex decision, and emits nothing an executor arm acts on; its
only behavioral surface is memory content reaching planner prompts (at the
lowest salience tier in the window). The one starvation death is within the
ordinary band for a fresh gemma-piloted world's first night (this agent also
lost an early planner call to a local-model timeout); nothing ties it to the
new channel. Flagged for the orchestrator rather than claimed away.

## Belief reconciliation (day-1 expectations)

`agent.belief_reinforced`: 0 this run — expected for a one-day soak: beliefs
only exist after the night-1 consolidation (tick ≈ 57,600+), and the few
sleep-to-dawn hours left almost no post-belief arrivals (buckets 3–4: 17
observations). Of the 14 gemma-authored beliefs, none was reconcilable by the
matcher: one carries coordinates + a feature word the vocabulary lacks
("quarry at (46,30)" — `quarry` is not in `beliefFeatureVocab`; noted as a
cheap vocabulary addition), one names a feature without coordinates ("a rock
… far to the east"). The confirm/disconfirm/silence paths, the bounded decay
arithmetic, and the SC-001 myth-dies-slowly trajectory are pinned by unit
tests through the real reducer (`internal/mind/reconcile_test.go`,
`internal/sim/observe_test.go`).

## Verdict

- SC-002 (soak half): observation-memory counts bounded and activity-shaped;
  no working-window flooding; dedup exact (events == memories).
- Survival behavior ordinary (one first-night starvation, mechanism
  provably untouched by the channel except via low-salience prompt texture).
- World preserved at `~/.promptworld/measure/task-80-soak` for re-analysis.
