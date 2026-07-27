---
id: TASK-137
title: >-
  Charter-delta experiment: default vs authored charter on a seeded world
  (TASK-111 AC5)
status: In Progress
assignee: []
created_date: '2026-07-25 19:37'
updated_date: '2026-07-27 02:09'
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
- [ ] #1 Two same-seed runs (default vs authored charter) with induced crises completed and compared
- [ ] #2 Delta (or proven absence of delta) recorded with numbers here and on TASK-111; AC#5 resolution recorded
- [ ] #3 Verdict on the teaching-game premise recorded: measurable delta (charter is a real player verb) or proven absence (charter surface is decoration — redesign follow-up carded); TASK-112's hold resolved accordingly
<!-- AC:END -->



## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Experiment launched (2026-07-26, operator go). Design: two same-seed (1337) worlds at ~/.promptworld/measure/task-137-{default,authored}, stage-4 --override (full guardian roster, no curriculum gating), all routes gemma4:12b-mlx, HARSH tuning.json (fire_burn_per_wood=3600 — 4x faster burnout; gru_emerge_per_mille=1000 — gru every night; dogfooding spec 048), 2 game-days each @ 8x. Arms run SEQUENTIALLY (parallel would couple both arms through the shared gemma queue — latency confound). Arm A: stock DefaultCharter (post-059, carries the survival duty). Arm B: authored survival-focused charter overwriting charter.md. Metrics: deaths, near-death recoveries, survival-turn count + actions, charges spent on survival, miracle door rejections (feeds TASK-136). This run also produces TASK-136's rejection-rate sample.

RE-TASKED (2026-07-26, operator): the morning's v4 arm worlds were discarded before any meaningful span — main advanced under them (FormatVersion 5 / spec 068 terrain gen with marsh+sand, spec 069/070, guardian package rename) and the binary was rebuilt by a sibling session. The experiment now runs on the CURRENT build: both arms recreated at launch time with the fresh binary (same-seed pairing holds — both arms share whatever map seed 1337 generates under the new terrain gen; the authored charter text is preserved in this session and re-planted at creation). Runs scheduled for tonight per operator (rig must be idle — the gemma endpoint + CPU interfere with active work).

Scheduled (operator, 2026-07-26): PARALLEL arms at 22:00 tonight -> ~04:00 done, rig free by morning. Symmetric-coupling caveat (shared gemma queue, same seed + dials both arms) will be recorded in the evidence doc. At launch: rebuild binary from current main (v5), recreate both arms fresh (new terrain gen), re-plant the authored charter from docs/design/evidence/task-137/authored-charter.md, harsh dials (fire_burn_per_wood=3600, gru_emerge_per_mille=1000), stage-4 --override, all-routes gemma, calibrate once + copy, verify horizon/tuning/watches before walking away. Trigger armed in-session.
<!-- SECTION:NOTES:END -->
