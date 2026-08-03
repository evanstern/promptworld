---
id: TASK-196
title: >-
  Full pack makes gathering fail silently — a villager can starve standing on
  food
status: In Progress
assignee: []
created_date: '2026-08-03 19:57'
updated_date: '2026-08-03 20:19'
labels:
  - bug
  - sim
  - survival
dependencies: []
priority: high
ordinal: 178001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
When a villager's hands are already full, gathering does nothing at all — and nobody is told. The villager
walks to the berries, does the full work of picking them, and comes away empty, with no message, no memory, and
no sign that anything went wrong. This card makes a full pack fail out loud, so the villager can notice, drop
something, and eat.

Observed in world-03: Cedar died of starvation on day 1 while standing on a live forage patch. Cedar had dropped
all six of his food at tick 10904 to carry more wood, then filled his pack to exactly 24/24 (20 wood + 4 planks,
bulkCap = 24). Every one of the ~20 forages that followed completed the full work cycle and yielded nothing;
every eat that followed was rejected 'Cedar has nothing to eat'. Cedar never ate once in his life. The event log
shows zero agent.foraged and zero agent.intent_failed for him after tick 8350 — just intent_done, over and over.

## Use cases

- As a villager, when my pack is full and I try to forage, I want to feel that my hands are too full to carry any
  more, so I can drop something and eat instead of starving on top of the food.
- As a player watching the chronicle, when a villager gathers and comes back with nothing, I want to see why,
  instead of watching a silent success repeat until they die.
- As the Guardian, when I look for the reason a villager cannot feed themselves, I want the real cause on the
  record, so I intervene on the pack and not on the berries.

## Diagnosis (pinned)

- internal/sim/executor.go:1361-1372 — the US1-AS1 zero-space guard emits a bare agent.intent_done when
  freeBulk(a.Inv) == 0. That is the same event a SUCCESSFUL gather emits, so a no-op is indistinguishable from a
  harvest to the villager, the planner, and the chronicle.
- internal/sim/agents.go:1295,1360-1376 — bulkCap = 24; bulk() counts one per unit of every kind.
- internal/sim/executor.go:1669-1671,1733-1741 — spec 096 already ships the right mechanism: agent.intent_failed
  with a closed reason vocabulary, paired same-tick with a situated first-person failure memory.
- internal/sim/state.go:1138-1156 — the intent_failed reducer closes the IntentLog record 'failed', which is
  exactly the signal the next planner thought needs and never got.

The fix is to route the zero-space guard through intentFailedEvents under a new 'pack full' reason, preserving
the no-yield/no-depletion invariant the guard exists to protect.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A gather (forage/chop/hunt/quarry/collect_water) completed with zero free bulk emits agent.intent_failed with a stable 'pack full' reason instead of a bare agent.intent_done
- [x] #2 The failure carries a situated first-person memory naming the full pack, at salIntentFailed salience, on the same tick as the failure event
- [x] #3 The US1-AS1 invariant is preserved: at zero free bulk there is still no yield and no depletion (Harvested and inventory both unchanged)
- [x] #4 The failed gather closes the IntentLog record 'failed' so the next planner thought can see the goal did not finish
- [x] #5 TestBulkZeroSpaceGatherNoEventNoDepletion is updated to assert the loud failure, and go test ./internal/sim/... passes
- [x] #6 A regression test reproduces the world-03 shape: a full-pouch agent standing on forage gets a distinguishable failure, not a silent success
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented on branch task-196-full-pack-silent-gather; draft PR #162 (https://github.com/evanstern/promptworld/pull/162).

Root cause confirmed from world-03's event log (17,469 events), not inferred: Cedar has ZERO agent.ate events for his whole life, 3 agent.foraged (last at tick 8350), and ZERO agent.intent_failed — which rules out the 'tile depleted' hypothesis, since that path emits intentFailTargetGone. He dropped 6 food_raw at tick 10904 to carry wood, reaching 20 wood + 4 planks = exactly bulkCap 24. Every gather after that hit the zero-space guard at internal/sim/executor.go and resolved as a bare agent.intent_done — the same event a successful harvest emits.

Fix routes the guard through spec 096's existing intentFailedEvents under a new intentFailPackFull reason, rather than inventing a mechanism. The guard's no-yield/no-depletion invariant is unchanged; only its audibility. 'pack full' is a distinct reason from 'contested' because the remedy is the agent's alone (drop or store).

Verification: go build clean, go vet clean, go test ./... exit 0 across 23 packages, zero failures. TestBulkZeroSpaceGatherNoEventNoDepletion updated (it asserted the bare intent_done by name); new TestIntentFailedForagePackFullWorld03 pins the world-03 shape.

Grounding rides the PR (spec 069): 26 wiki notes re-verified against 012f715f, 6 edited (3 of them stated outright that the zero-free-bulk case is 'skipped entirely — no event', which this change makes false), 20 re-pinned unchanged. Player docs re-stamped, 16/16 fresh. check-merge-drift pr => exit 0.

MODEL TIER (Principle V): implemented INLINE on claude-opus-5 rather than dispatching spec-implementer. Deviation from the delegation rule, taken deliberately and recorded here: this session carries an explicit operator directive not to call the Agent tool unless requested, which conflicts with Principle V's 'never implements inline'. Flagged to the operator in-session rather than silently resolved either way. Spec Kit trivial exemption applies (surgical fix, file:line diagnosis pinned on this card before work, ACs on this card).
<!-- SECTION:NOTES:END -->
