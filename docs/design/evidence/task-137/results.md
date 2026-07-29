# TASK-137 Charter-delta experiment — results (run of 2026-07-26 22:12 → 2026-07-27 ~04:10)

Two same-seed (1337) worlds, stage-4 `--override`, 2 game-days each at 8x under
harsh dials (`fire_burn_per_wood=3600` — 4× faster burnout;
`gru_emerge_per_mille=1000` — gru every night), all routes
`gemma4:12b-mlx` @ `mbpro-m1.local:11434` (parallel 4, capacity 4,
`tool_mode: "native"` — see deviation note). Arm A: stock DefaultCharter
(post-059 survival duty). Arm B: authored survival charter
([authored-charter.md](authored-charter.md), planted verbatim, diff-verified
after daemon start). Arms ran in PARALLEL per operator schedule; the
symmetric-coupling caveat (shared gemma queue, same seed and dials) applies to
both arms identically. Worlds preserved at
`~/.promptworld/measure/task-137-{default,authored}` (stopped; default ended
tick 178,263 / day 3 07:31, authored 177,807 / day 3 07:23).

## Headline numbers

| Metric | A: default charter | B: authored charter |
|---|---|---|
| Deaths | **1** (Rowan/agent 7, starvation, tick 132,780) | **1** (Rowan/agent 7, starvation, tick 120,780) |
| Guardian interventions **landed** | **0** | **2** (both `send_vision`, both targeting the starving villager, food-seeking guidance, ticks 101,663 / 117,298) |
| Guardian interventions **attempted** | 3 (`send_vision` ×3) | 7 (`send_vision` ×2 landed, `work_miracle` ×5 rejected) |
| Rejected at the targeting/validity gate | 3/3 (100%) | 5/7 (71%) |
| Gru emergences / attacks | 2 / 3 (all victims survived) | 2 / 2 (all victims survived) |
| Metatron charge regens (spend-cycle proxy) | 2 | 4 |
| Guardian survival watches placed | 3 (near-death, starvation, exposure) | 3 (same) |

## Reading

1. **Charter wording measurably changed guardian behavior.** The authored-arm
   guardian engaged earlier and harder: 7 privileged-action attempts vs 3, and
   it was the only arm to land any intervention — two targeted, on-topic
   visions to the villager its starvation watch flagged, 19k and 3.5k ticks
   before that villager died. The default-arm guardian attempted only after
   its watch fired and landed nothing.
2. **…but not the survival outcome, in this sample.** Both arms lost the same
   same-seed villager to the same cause; the authored arm actually lost them
   ~12k ticks EARLIER (within-run stochasticity — with n=1 crises per arm, the
   death-time delta is noise). On outcomes alone, the charter surface made no
   measurable difference in one 2-game-day pairing. TASK-111 AC5's
   falsifiability demand is answered in the narrow sense: the charter is NOT
   decoration (behavior moved), but a single authored rewrite did not buy a
   life here.
3. **The binding constraint is tool competence, not motivation** — the
   experiment's clearest finding, and TASK-136's sample: 8 of 10 privileged
   attempts across both arms were rejected by the targeting/validity door.
   Failure modes, all model-side grammar/grounding fumbles by gemma4:12b-mlx:
   - `send_vision` place grants missing `place_x`/`place_y` (2×, default arm);
   - a place grant naming a tree at (56,34) where the world has none —
     hallucinated coordinates (1×, default arm);
   - `work_miracle give_item` with item kind `"fire"`/empty — not an item
     (3×, authored arm);
   - `work_miracle move` with an empty entity class (2×, authored arm).
   The authored charter pushed the guardian to the door more often; the door
   correctly refused malformed requests every time. A guardian that could
   TARGET would plausibly have converted its 5 rejected miracles into a
   starvation rescue — the charter delta may be understated by the tool floor.
4. **Harsh dials worked as designed**: gru emerged both nights in both arms
   (`sim.tuning_applied` carries both dial sets), attacks landed but killed no
   one; the actual killer was starvation under time pressure — the crisis the
   watches exist for.

## Caveats

- n=1 crisis per arm; one seed; one model. No statistical claim — this is a
  falsifiability probe, not a study.
- Parallel arms share the gemma queue: latency coupling is symmetric but
  compresses both guardians' effective turn counts equally; a sequential
  re-run would decouple it.
- `tool_mode` deviation: launch config said `"json"` implicitly via defaults;
  gemma4:12b-mlx on this ollama ignores `response_format: json_schema`
  (fenced/empty envelopes; calibrate circuit-opened), so BOTH arms ran
  `"native"` tool-calls (probe-verified). Symmetric, but the model's native
  tool-call argument quality is itself the dominant failure mode above —
  a stronger model or a `json`-honoring backend is the first knob for any
  re-run.
- The two agent.died payloads carry no guardian attribution; death causes read
  directly from the event (`starvation`).

## Feeds

- **TASK-111 AC5**: resolved — charter quality is behaviorally live (not
  self-grading decoration), outcome delta unproven at n=1; recorded on both
  cards.
- **TASK-136**: rejection-rate sample recorded — 8/10 rejected overall;
  100% (3/3) default arm, 71% (5/7) authored arm; failure taxonomy above.
- Re-run suggestions (not carded): sequential arms, stronger model tier for
  `metatron_watch`, and/or a targeting-digest prompt asking the guardian to
  read coordinates from its watch context before granting places.

## Addendum (2026-07-28, TASK-163): confound bound superseded

The 8/10 (80%) mechanical-noise bound quoted above is superseded for any run
on the fixed binary (`2eef530`+) with a sonnet-tier guardian turn. TASK-163's
two-stage mitigation (guardian turn route → `cc/claude-sonnet-5`, then the
`give_item` grant vocabulary surfaced in schema/guidance/rejection) measured
**36% (5/14)** overall rejection with **~0% model-side noise** — zero
vocabulary, grammar, or coordinate-hallucination failures; the residue is
world-mechanics bounds (carry cap, self-repairing in-turn) and 8x
position-freshness races on `move`. Prompt-attribution comparisons (this
experiment's successor TASK-164, the TASK-67 fork duel) should quote
TASK-163's bound and hold the guardian model tier fixed as an experiment
variable. Full data: [task-163/results.md](../task-163/results.md).
