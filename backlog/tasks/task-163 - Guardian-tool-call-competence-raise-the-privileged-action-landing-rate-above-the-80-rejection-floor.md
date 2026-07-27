---
id: TASK-163
title: >-
  Guardian tool-call competence: raise the privileged-action landing rate above
  the 80% rejection floor
status: In Progress
assignee: []
created_date: '2026-07-27 13:20'
updated_date: '2026-07-27 23:33'
labels:
  - mvls
  - guardian-survival
dependencies: []
priority: high
ordinal: 131000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Carded from TASK-136's live measurement (2026-07-27, docs/design/evidence/task-137/results.md): 8 of 10 guardian privileged-action attempts were rejected at the targeting/validity door (80%; 3/3 default arm, 5/7 authored arm) — the spec-059 targeting digest shipped and functions, but rejections did not drop toward 0 because every failure was model-side tool-call incompetence on gemma4:12b native tool-calls: send_vision place grants missing place_x/place_y (2), hallucinated tree coords (1), work_miracle give_item with non-item kind 'fire'/empty (3), work_miracle move with empty entity class (2). This floor is the mechanical-noise/confound bound for all prompt-attribution comparisons (TASK-137 charter delta, TASK-67 fork duel) and a checkpoint for TASK-112 dispatch (docs/design/guardian-directives-runbook.md). Candidate mitigations from the evidence doc, not pre-decided: stronger model tier for metatron_watch; prompt shaping so the guardian reads coordinates from its watch-context digest before granting places; schema/argument validation with in-turn repair retry. Event evidence: ~/.promptworld/measure/task-137-{default,authored}/world.db (cog.tool_call verdict/reason fields).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Failure taxonomy triaged to a chosen mitigation (model tier, prompt shaping, arg-repair retry, or a combination) with the decision and rationale recorded on this card
- [ ] #2 Post-fix run with >=10 privileged-action attempts: rejection rate measured, materially below the 80% baseline, evidence doc recorded
- [ ] #3 The new rate replaces 80% as the quoted confound bound for prompt-attribution comparisons (TASK-137 evidence doc updated; TASK-67 duel scoreboard cites it)
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
AC#1 TRIAGE DECISION (2026-07-27, operator-steered): mitigation = STRONGER MODEL TIER for the guardian turn via the 9router proxy (openai_compat @ localhost:20128, probe-verified live). World-config only (llm.json routes) — no code change; prompt shaping and arg-repair retry DEFERRED unless the tier change fails to move the rate. Model ladder per operator: cc/claude-sonnet-5 first; escalate to cc/claude-opus-4-8 if the rate does not move materially; cc/claude-fable-5 sanctioned as an occasional probe only (thinking model, wall-clock heavy — never the always-on turn model). Rationale: the TASK-137 taxonomy shows all 8 rejections were model-side grammar/grounding fumbles (gemma4:12b-mlx native tool-calls); the spec-059 digest and the door behaved correctly, so the model is the knob — exactly the evidence doc's first re-run suggestion. CORRECTION to this card's framing: the failing privileged tool calls ride the guardian TURN (route 'metatron', the KindMetatron tool loop) — 'metatron_watch' is the bare yes/no order-match confirm (spec 029, never a tool loop) and stays cheap-first local. Experiment design: re-run the TASK-137 recipe (both arms, seed 1337, stage-4 --override, harsh dials fire_burn_per_wood=3600 / gru_emerge_per_mille=1000, 2 game-days @ 8x, authored charter re-planted verbatim from docs/design/evidence/task-137/authored-charter.md) with ONE variable changed: routes.metatron -> cc/claude-sonnet-5 (pricing 3/15 usd-per-mtok); every other route stays gemma4:12b-mlx tool_mode native as in the baseline. Binary rebuilt from current main — spec 086 payload census landed since baseline @072dd71 (payload plumbing, not guardian prompting; recorded for comparability). Worlds at ~/.promptworld/measure/task-163-{default,authored}.

RUN LOG (2026-07-27): both arms launched ~14:35 (binary main@89c78b9, sonnet 0.8 s/pt via 9router, gemma 9.9 s/pt — mild playtest-1 contention vs baseline 9.3, recorded). 2-game-day mark reached ~20:35. Interim tallies at tick 172800 — DEFAULT arm: 7 privileged attempts, 2 landed (both send_vision, clean coordinates — the baseline's 0/3 send_vision failure mode is GONE on sonnet), 5 rejected, ALL FIVE the same signature: work_miracle give_item with item kind 'food' x4 / 'forage' x1 -> 'unknown item kind'. 2 deaths (Ash tick ~66k, Sage night 2, both starvation, both villagers the guardian tried to feed). AUTHORED arm: zero watch fires, zero attempts, zero deaths in 2 days (same-seed stochastic divergence — its village never hit crisis). DEVIATION: extended both arms to 3 game-days (target tick 259200, ~23:30) because total attempts 7 < the AC#2 bar of 10. CODE DIAGNOSIS pinned mid-run: the give_item vocabulary lives ONLY in grantableKind (internal/sim/miracles.go:472 — wood/stone/water/planks/refined_stone/food_raw/food_cooked/meals/spear/axe); the work_miracle tool schema exposes 'item' as a bare string (internal/guardian/turn.go:88), no enum, no guidance mention, and the rejection reason does not enumerate valid kinds — the guardian can only guess. On sonnet the ENTIRE rejection residue is this one un-surfaced enum; grammar/grounding failures (baseline taxonomy) have vanished.

RUN COMPLETE (2026-07-27 ~23:40, extended to 3 game-days/tick 259200 per the <10-attempts deviation): FINAL SAMPLE 7 privileged attempts, 2 landed, 5 rejected (71% vs baseline 80%). Taxonomy is CATEGORICALLY different from baseline: send_vision 2/2 landed (baseline 0/3 — coordinate grounding fixed by the tier change); work_miracle 0/5, every rejection the SAME un-surfaced enum (give_item item='food' x4 / 'forage' x1). Authored arm: zero crises, zero attempts, zero deaths across all 3 days (same-seed stochastic divergence); default arm deaths 2 (Ash, Sage — both starvation, both villagers the guardian tried to feed through the vocabulary gap). VERDICT on AC#1's ladder: the tier change is NECESSARY but NOT SUFFICIENT — it eliminated the model-side failure class entirely but cannot beat the 80% headline alone because the item enum is invisible to any model. Per AC#1's recorded activation condition ('prompt shaping deferred UNLESS the tier change fails to move the rate'), the prompt-shaping leg is now ACTIVE: surface grantableKind's vocabulary (internal/sim/miracles.go:472) in the work_miracle tool schema item field (internal/guardian/turn.go:88) and enumerate valid kinds in the door's rejection reason (internal/sim/miracles.go:430), single source of truth, then re-measure tier+enum combined. Implementer tier: Sonnet (constitution V rubric: surgical two-file surface change, no authority/concurrency/doctrine semantics; file:line diagnosis pinned above). Worlds preserved at ~/.promptworld/measure/task-163-{default,authored}, daemons stopped.
<!-- SECTION:NOTES:END -->
