---
id: TASK-166
title: >-
  Move-miracle target freshness: re-resolve the entity at the door instead of
  trusting surveyed coordinates
status: Done
assignee: []
created_date: '2026-07-29 13:59'
updated_date: '2026-07-29 19:45'
labels:
  - mvls
  - guardian-survival
dependencies: []
priority: medium
ordinal: 134000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Carded from TASK-163's evidence (docs/design/evidence/task-163/results.md, PR #128 merged 46f55b1): on the fixed binary, 3 of 5 residual privileged-action rejections were position-freshness races — the guardian surveys, forms a well-formed 'move' with the surveyed source coordinates, and the villager walks away during model latency (at 8x, 30-60s of thinking = 240-480 ticks), so the door refuses 'no living villager at (x,y)'. Not a competence or guidance gap; an architectural property of coordinate-addressed moves at speed. Observed workaround by the guardian itself: move the structure to the villager instead. Candidate fixes (decision needed): (a) when kind=move, class=villager, and a villager NAME is supplied, the door re-resolves the villager's live position at apply time and treats x/y as advisory; (b) a freshness token binding the call to the survey tick with a re-resolve grace window; (c) guidance instructing name-only villager moves. Anchor points: internal/sim/miracles.go applyEntityMoved (source class MUST be present at (x,y) — validated at the dry-run door), internal/guardian/turn.go move arg parsing (villager name field already exists in miracleArgs). Any fix must preserve the door's determinism and replay semantics (validation precedes the charge; recorded moves re-apply cleanly).

Spec: specs/091-move-miracle-freshness
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Decision recorded on this card: door-side name re-resolution vs freshness token vs guidance-only, with replay/determinism implications analyzed
- [x] #2 Implementation + tests: a raced villager-move (entity moved after survey) lands under the chosen mechanism; structure/pile moves unchanged; replay of pre-fix recorded moves still applies cleanly
- [x] #3 Probe verification on a live world: name-addressed villager moves land at 8x; evidence appended here
- [x] #4 Spec phase: Door-side name re-resolution
- [x] #5 Spec phase: Tests
- [x] #6 Spec phase: Grounding + probe
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
board-sweep-2026-07-29 lane 2: AC1 decision recorded in spec 091 — (a) door-side name re-resolution, x/y advisory; replay analysis in spec.md (reducer arm untouched, emitter-computes). Tier: Sonnet — narrow door fix, single mechanism; escalation trigger: any change to applyEntityMoved reducer semantics.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Merged via PR #133 (merge commit on main; branch tip 3b26820). Door-side name re-resolution shipped: villager moves naming the villager resolve live position at the door, x/y advisory; reducer untouched, replay byte-identity regression-tested; coordinate/structure/pile paths unchanged. Live probe on preserved measure world (~/.promptworld/measure/task-166-probe): guardian call surveyed (59,20), door resolved (62,18) after 139-tick race, landed. Evidence: docs/design/evidence/task-166/results.md. TASK-163 residual position-freshness class retired to the deliberate coordinate-only remainder.
<!-- SECTION:FINAL_SUMMARY:END -->
