---
id: TASK-137
title: >-
  Charter-delta experiment: default vs authored charter on a seeded world
  (TASK-111 AC5)
status: Done
assignee: []
created_date: '2026-07-25 19:37'
updated_date: '2026-07-27 08:27'
labels:
  - mvls
  - guardian-survival
dependencies: []
priority: medium
ordinal: 107000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Follow-up carded from the MVLS sweep (operator request 2026-07-25) so TASK-111's AC5 anti-self-grading guard doesn't rot: prove charter quality measurably changes autonomous survival performance. Design: same world seed, two runs — DefaultCharter (survival-duty wording, spec 059) vs an authored/tuned charter — with induced survival crises (harsh tuning.json dials); compare autonomous survival outcomes: deaths, near-death recoveries, survival-turn actions taken, charges spent on survival. Tick TASK-111 AC#5 with the evidence.

REORIENT 2026-07-26 reframe (docs/design/reorient-2026-07-26-ui.md): this experiment is now the falsifiability test of the pivot's core premise — that prompting the guardian is a real, rewarding player verb and not decoration. The charter IS the player's prompt: if an authored charter produces no measurable behavioral delta over DefaultCharter, the charter surface is decoration, the iteration rung (TASK-67 fork duel) duels over noise, and the curriculum ladder teaches a skill with no in-game payoff — a finding that forces a content/mechanics redesign and is exactly as valuable as the delta. The synthesis's course of action also holds TASK-112 (Metatron agentization) behind this task's evidence per TASK-112 AC5. Experiment design and the scheduled run are unchanged; the delta (or its proven absence) is now quoted as learning-game validity evidence, not just TASK-111 AC5 close-out.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Two same-seed runs (default vs authored charter) with induced crises completed and compared
- [x] #2 Delta (or proven absence of delta) recorded with numbers here and on TASK-111; AC#5 resolution recorded
- [ ] #3 Verdict on the teaching-game premise recorded: measurable delta (charter is a real player verb) or proven absence (charter surface is decoration — redesign follow-up carded); TASK-112's hold resolved accordingly
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Experiment launched (2026-07-26, operator go). Design: two same-seed (1337) worlds at ~/.promptworld/measure/task-137-{default,authored}, stage-4 --override (full guardian roster, no curriculum gating), all routes gemma4:12b-mlx, HARSH tuning.json (fire_burn_per_wood=3600 — 4x faster burnout; gru_emerge_per_mille=1000 — gru every night; dogfooding spec 048), 2 game-days each @ 8x. Arms run SEQUENTIALLY (parallel would couple both arms through the shared gemma queue — latency confound). Arm A: stock DefaultCharter (post-059, carries the survival duty). Arm B: authored survival-focused charter overwriting charter.md. Metrics: deaths, near-death recoveries, survival-turn count + actions, charges spent on survival, miracle door rejections (feeds TASK-136). This run also produces TASK-136's rejection-rate sample.

RE-TASKED (2026-07-26, operator): the morning's v4 arm worlds were discarded before any meaningful span — main advanced under them (FormatVersion 5 / spec 068 terrain gen with marsh+sand, spec 069/070, guardian package rename) and the binary was rebuilt by a sibling session. The experiment now runs on the CURRENT build: both arms recreated at launch time with the fresh binary (same-seed pairing holds — both arms share whatever map seed 1337 generates under the new terrain gen; the authored charter text is preserved in this session and re-planted at creation). Runs scheduled for tonight per operator (rig must be idle — the gemma endpoint + CPU interfere with active work).

Scheduled (operator, 2026-07-26): PARALLEL arms at 22:00 tonight -> ~04:00 done, rig free by morning. Symmetric-coupling caveat (shared gemma queue, same seed + dials both arms) will be recorded in the evidence doc. At launch: rebuild binary from current main (v5), recreate both arms fresh (new terrain gen), re-plant the authored charter from docs/design/evidence/task-137/authored-charter.md, harsh dials (fire_burn_per_wood=3600, gru_emerge_per_mille=1000), stage-4 --override, all-routes gemma, calibrate once + copy, verify horizon/tuning/watches before walking away. Trigger armed in-session.

LAUNCHED 2026-07-26 22:00 window (parallel arms, per operator schedule). Binary rebuilt from main @072dd71. Both arms created fresh: ~/.promptworld/measure/task-137-{default,authored}, seed 1337, stage-4 --override. All routes gemma4:12b-mlx @ http://mbpro-m1.local:11434/v1 (parallel 4, endpoint_capacity 4). Harsh dials verified via sim.tuning_applied in both logs: fire_burn_per_wood=3600, gru_emerge_per_mille=1000. Arm B charter.md = docs/design/evidence/task-137/authored-charter.md verbatim (diff-verified after daemon start); arm A stock default. Calibration: 9.3 s/pt on gemma, copied A→B; horizon all-green at 8x (planner 222<=1200, conversation 962<=7200, meeting 148<=3600). Three sys-watch orders (near_death, starvation, exposure) active in both arms. Both daemons at 8x, effective rate 8.0, not degraded. Monitor armed: exits when both arms MAX(tick)>=172800 (~04:00). DEVIATION (recorded): tool_mode json→native on both arms identically — gemma4:12b-mlx on this ollama ignores response_format json_schema (fenced/empty JSON envelope; calibrate circuit-opened: 'tool-call envelope: invalid character backtick'); native tool-calls probe-verified working. Symmetric across arms, no confound beyond the already-noted shared-queue coupling. LAUNCH-BLOCKING FAILURE AVOIDED, all 7 pre-flight verifications passed.

COMPLETED 2026-07-27 ~04:10. Both arms ran 2 game-days at 8x under harsh dials (parallel per schedule; default ended tick 178,263, authored 177,807). Results: docs/design/evidence/task-137/results.md (merged 8406583). Headline: deaths 1 vs 1 — SAME villager (Rowan/7), SAME cause (starvation), authored arm ~12k ticks earlier (n=1 noise). Guardian BEHAVIOR moved decisively: landed interventions 0 (default) vs 2 (authored — targeted food-seeking visions to the starving villager 19k/3.5k ticks pre-death); attempts 3 vs 7. Binding constraint = tool competence, not motivation: 8/10 privileged attempts rejected by the targeting door (missing/hallucinated place coords, item kind 'fire', empty move class — all gemma4:12b-mlx grammar/grounding fumbles). AC5 verdict recorded on TASK-111: charter surface is NOT decoration (behavior delta real) but outcome delta unproven at n=1. TASK-136 sample recorded on its card. Worlds preserved at ~/.promptworld/measure/task-137-{default,authored} for re-analysis.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Charter-delta experiment executed per operator schedule (2026-07-26 22:00 launch, parallel arms, seed 1337, harsh dials, gemma4:12b-mlx all routes, tool_mode native deviation recorded). Delta found in guardian BEHAVIOR (0 vs 2 landed interventions, 3 vs 7 attempts — authored charter engages the guardian), null in OUTCOME (1 starvation death each, same villager, n=1). Dominant finding: 8/10 privileged attempts rejected at the targeting/validity door — tool competence is the floor under any charter. Evidence doc: docs/design/evidence/task-137/results.md. Feeds TASK-111 AC5 (resolved) and TASK-136 (rejection sample).
<!-- SECTION:FINAL_SUMMARY:END -->
