---
id: TASK-197
title: >-
  Guardian inventory tool: let the Guardian read and change what a villager
  carries, and make it look first
status: In Progress
assignee: []
created_date: '2026-08-03 19:57'
updated_date: '2026-08-04 02:49'
labels:
  - guardian
  - tools
  - survival
dependencies: []
priority: high
ordinal: 179001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The Guardian can see a villager's health, hunger and warmth, but not what is actually in their pack — so it
gives advice that cannot be followed. This card gives the Guardian a tool to look inside a villager's pack and to
change what is in it, and makes the Guardian look before it speaks or acts.

Observed in world-03: with Cedar dying at 24/24 carrying zero food, the Guardian sent him a vision reading 'You
are carrying food in your own hands this very moment... eat now, whatever you hold.' Cedar was carrying twenty
wood and four planks and not one scrap of food. Across both of its watch turns the Guardian spent every call on
survey_site and explain — it looked at the ground eight times and never once at the man. It also never attempted
a miracle, and could not have: give_item is rejected whole at the carry cap, and Cedar had zero free bulk, so the
only intervention that would have saved him was sealed shut by the same full pack nobody could see.

## Use cases

- As the Guardian, when a villager is starving, I want to open their pack and see exactly what they carry, so I
  can tell the difference between 'has no food', 'has food and won't eat', and 'has no room to pick food up'.
- As the Guardian, when I find a starving villager whose hands are too full to accept a gift, I want to be able
  to lift the deadweight out of their pack, so my miracle is not refused at the door.
- As a player, when the Guardian speaks to my villager, I want its words to match what that villager is really
  carrying — a god that misreads the pack is worse than a silent one.
- As a villager, when the Guardian reaches into my pack and takes or leaves something, I want to remember that it
  happened, and my neighbours to be able to hear of it.

## Notes for the specifier

- internal/guardian/turn.go:1256-1294 (buildTargetingDigest) already prints 'carrying N/24, M free to receive'
  per living villager, so gross bulk is visible — but not the CONTENTS, which is what distinguishes the three
  cases above. The digest was there in world-03 and the Guardian still got it wrong; a one-line summary buried in
  a targeting digest is evidently not enough, hence an explicit tool it must call.
- internal/sim/miracles.go:463-466 — give_item rejects an over-cap grant WHOLE rather than clamping, which is why
  read access alone does not close the world-03 hole; removal is the missing half.
- Write access is a doctrine-adjacent change (the Guardian gaining a new way to reach into the world), so this
  wants full Spec Kit and an explicit decision on what is auditable: every reach-in should leave an event and a
  situated memory, never a silent mutation.
- Related: TASK-167 (carry-cap guidance for give_item, spec 095) is the same blindness seen from the granting
  side; TASK-196 is the villager-side half of the world-03 death.

Spec: specs/116-guardian-inventory-tool
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A guardian-facing read tool returns a named villager's full inventory — every kind and its count — plus carried bulk and free bulk against sim.BulkCap
- [ ] #2 The Guardian's survival watch turn reads a villager's inventory before sending any vision or miracle concerning food, so its message cannot contradict what that villager carries
- [ ] #3 A guardian-facing write path can both add to and remove from a villager's inventory, respecting the carry cap, so a full pack can be relieved and then fed
- [ ] #4 Every guardian reach-in emits a durable event and a situated first-person memory for the villager — no silent mutation of a pack
- [ ] #5 Regression from world-03: a starving villager at 24/24 carrying zero food never receives a vision claiming they are carrying food
- [ ] #6 The read tool is exercised in the guardian's own tests against a fixture villager whose pack is full of non-food
<!-- AC:END -->
