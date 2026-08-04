---
id: TASK-197
title: >-
  Guardian inventory tool: let the Guardian read and change what a villager
  carries, and make it look first
status: In Progress
assignee: []
created_date: '2026-08-03 19:57'
updated_date: '2026-08-04 03:42'
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

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Spec 116 (specs/116-guardian-inventory-tool), three pieces on one branch / one PR:
1. THE EYE — per-villager inventory CONTENTS mirror in the guardian's absorb snapshot, plus inspect_pack: a charge-free Read tool rendering a deterministic pack sheet (every kind + count, carried bulk, free bulk vs sim.BulkCap, an explicit no-food line). Pure addition; granted at every stage (survey_site/explain/brief_myths precedent).
2. THE DISCIPLINE — a per-turn look-first ledger on turnDispatch and a gate in the send_vision / work_miracle handlers, firing ONLY on survival-origin turns (turnOrigin.survival), per-villager, expressed as the existing repairable rejected_gate whose reason names inspect_pack. Prose was already tried and failed: the carry-headroom digest line was in the world-03 prompt and the Guardian read past it.
3. THE HAND — a take_item miracle kind (1 charge, world-shaping, NOT in the stage-1/2 ceiling) landing guardian.item_taken plus one situated first-person memory in a single atomic batch. Removed goods MOVE into the pile on the villager's own tile via the existing agent.dropped machinery — reject-whole never clamp, spear/axe wear ordering and food-batch spoilage preserved, total goods conserved.

Sequencing: phase 3 (read) is the shippable MVP; phase 4 (gate) is what makes it get used; phase 5 (removal) lands last because it is the only piece that mutates the world.

TIER: Opus 5 (.claude/agents/spec-implementer-opus.md), delegated. Rubric justification: doctrine-adjacent behavior change — the Guardian gains a NEW way to reach into the world (write access to a villager's inventory) and a NEW structural refusal on its own survival turn. Both are capability-surface changes, not routine single-package work.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-08-03 — Spec 116 authored and linked (spec.md, plan.md, tasks.md T001-T046, contracts/pack-access.md). Stub landed on main first per the stub-first rule; spec-bridge check exits 0.

Six design decisions were RESOLVED FROM ARTIFACTS rather than re-asked, and are recorded in spec.md's Assumptions section, each FLAGGED FOR OPERATOR REVIEW (the spec 084/085/107 precedent for capability-surface decisions):
- A1 the look-first gate is structural, not prose — the card's own world-03 evidence forecloses a prompt-only remedy.
- A2 the gate keys on turnOrigin.survival, NOT on message text. Consequence stated plainly: it does not fire on an ordinary console turn where the Guardian might also misread a pack. Deliberate narrow scope — the death being fixed happened on a survival turn. Classifying whether a vision 'concerns food' would mean reading model prose at the door: non-deterministic, unauditable, untestable.
- A3 taken goods DROP TO A PILE at the villager's feet; they are not unmade. The card says 'lift the deadweight out', not destroy the village's labour, and the agent.dropped arm already does exactly this transfer.
- A4 take_item is priced and staged as give_item is (1 charge, excluded from the stage-1/2 ceiling) — same class of act pointed the other way.
- A5 inspect_pack is granted at every stage — read-only, widens no acting capability.
- A6 AC#5 is verified STRUCTURALLY. 'never receives a vision claiming they are carrying food' cannot be asserted against model prose; its testable projection is SC-001 + SC-002 together — the Guardian cannot send that villager a vision at all without first receiving a sheet that says in words that they carry no food. The spec claims that projection and no more.

2026-08-03 — PR #165 opened (https://github.com/evanstern/promptworld/pull/165). Implementation complete on task-197-guardian-inventory-tool, tasks.md T001-T046 all ticked.

TIER SERVED: Opus 5 via .claude/agents/spec-implementer-opus.md (the definition's frontmatter pin, not a model parameter). Rubric: doctrine-adjacent behavior change — new guardian write access into a villager's inventory plus a new structural refusal on the survival turn.

WHAT SHIPPED: (1) inspect_pack — charge-free Read tool over a new deep-copied agentInv mirror, deterministic sheet naming every kind/count, carried+free bulk vs sim.BulkCap, and an explicit food line; granted at every stage. (2) The look-first gate — on survival-origin turns only, send_vision / give_item / take_item at a villager not yet inspected THIS turn is rejected_gate naming inspect_pack; per-villager, per-turn, origin-keyed (never message text). (3) take_item — 1-charge miracle landing guardian.item_taken plus one situated first-person memory in one atomic batch, moving goods into the tile's pile via the agent.dropped rules (reject-whole, wear ordering, food batching, conservation invariant); excluded from the stage-1/2 ceiling; IPC door parity.

GATES: go build/vet/test all clean (23 pkgs, exit 0); check-tui-design --changed passed; check-merge-drift pr exit 0; 82 wiki notes re-verified and re-pinned; 16 player docs fresh.

GROUNDING FINDING (corrected in-branch): the standing wiki claim 'a miracle never mints a new persistent entity' is now FALSE — take_item's pileFor is create-or-merge, so a removal onto a bare tile mints a sim.Pile. It is the only miracle kind that can. Nothing is destroyed either way (inventory + tile pile totals conserved).

SPEC CORRECTION (found by the implementer, verified, fixed rather than left wrong): the contract's 'preserve FoodBatch identity and spoilage' clause was unsatisfiable — carried food holds no spoilage (Inventory.FoodRaw/FoodCooked/Meals are bare ints; only Pile.FoodBatch has SpoilAt). A pack-to-pile transfer MINTS a rot window, exactly as a villager's own drop does.

FOLLOW-ONS FLAGGED, NOT INVENTED (candidates for cards if the operator wants them): (a) vocabulary seam — the sheet renders storage names ('spears') while take_item's enum uses grant names ('spear'), so a model reading the sheet may call take_item(item='spears') and take a repairable whole-refusal; (b) no CLI verb for removals — the IPC door has parity but 'promptworld work' has no 'take' verb, so an operator cannot reach take_item from the CLI.

AWAITING: operator review of PR #165, including the six spec Assumptions (A1-A6) flagged for review — most notably A2 (the gate does NOT fire on ordinary console turns, only survival turns) and A6 (AC#5 is verified structurally, since model prose cannot be asserted against).
<!-- SECTION:NOTES:END -->
