---
id: TASK-137
title: >-
  Charter-delta experiment: default vs authored charter on a seeded world
  (TASK-111 AC5)
status: In Progress
assignee: []
created_date: '2026-07-25 19:37'
updated_date: '2026-07-26 16:26'
labels:
  - mvls
  - guardian-survival
dependencies: []
priority: medium
ordinal: 107000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Follow-up carded from the MVLS sweep (operator request 2026-07-25) so TASK-111's AC5 anti-self-grading guard doesn't rot: prove charter quality measurably changes autonomous survival performance. Design: same world seed, two runs — DefaultCharter (now carrying the survival-duty wording, spec 059) vs an authored/tuned charter — with induced survival crises (e.g. tuning.json dials to make nights harsh); compare autonomous survival outcomes: deaths, near-death recoveries, survival-turn actions taken, charges spent on survival. The point is falsifiability: if charter wording makes no measurable difference, the charter surface is decoration and TASK-111's autonomy rests on the frame alone — that finding is as valuable as the delta. Tick TASK-111 AC#5 with the evidence.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Two same-seed runs (default vs authored charter) with induced crises completed and compared
- [ ] #2 Delta (or proven absence of delta) recorded with numbers here and on TASK-111; AC#5 resolution recorded
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Experiment launched (2026-07-26, operator go). Design: two same-seed (1337) worlds at ~/.promptworld/measure/task-137-{default,authored}, stage-4 --override (full guardian roster, no curriculum gating), all routes gemma4:12b-mlx, HARSH tuning.json (fire_burn_per_wood=3600 — 4x faster burnout; gru_emerge_per_mille=1000 — gru every night; dogfooding spec 048), 2 game-days each @ 8x. Arms run SEQUENTIALLY (parallel would couple both arms through the shared gemma queue — latency confound). Arm A: stock DefaultCharter (post-059, carries the survival duty). Arm B: authored survival-focused charter overwriting charter.md. Metrics: deaths, near-death recoveries, survival-turn count + actions, charges spent on survival, miracle door rejections (feeds TASK-136). This run also produces TASK-136's rejection-rate sample.

RE-TASKED (2026-07-26, operator): the morning's v4 arm worlds were discarded before any meaningful span — main advanced under them (FormatVersion 5 / spec 068 terrain gen with marsh+sand, spec 069/070, guardian package rename) and the binary was rebuilt by a sibling session. The experiment now runs on the CURRENT build: both arms recreated at launch time with the fresh binary (same-seed pairing holds — both arms share whatever map seed 1337 generates under the new terrain gen; the authored charter text is preserved in this session and re-planted at creation). Runs scheduled for tonight per operator (rig must be idle — the gemma endpoint + CPU interfere with active work).
<!-- SECTION:NOTES:END -->
