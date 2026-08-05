---
id: TASK-197
title: >-
  Guardian inventory tool: let the Guardian read and change what a villager
  carries, and make it look first
status: Done
assignee: []
created_date: '2026-08-03 19:57'
updated_date: '2026-08-05 02:40'
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
- [x] #1 A guardian-facing read tool returns a named villager's full inventory — every kind and its count — plus carried bulk and free bulk against sim.BulkCap
- [x] #2 The Guardian's survival watch turn reads a villager's inventory before sending any vision or miracle concerning food, so its message cannot contradict what that villager carries
- [x] #3 A guardian-facing write path can both add to and remove from a villager's inventory, respecting the carry cap, so a full pack can be relieved and then fed
- [x] #4 Every guardian reach-in emits a durable event and a situated first-person memory for the villager — no silent mutation of a pack
- [x] #5 Regression from world-03: a starving villager at 24/24 carrying zero food never receives a vision claiming they are carrying food
- [x] #6 The read tool is exercised in the guardian's own tests against a fixture villager whose pack is full of non-food
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

2026-08-04 — OPERATOR REVIEW of the spec 116 Assumptions (A1-A6), recorded:

- A2 OVERTURNED (widen). The turn-origin trigger is too narrow. Ratified replacement: the gate becomes PERIL-KEYED for visions and UNCONDITIONAL for pack reaches — send_vision is gated on ANY turn origin when the target villager is in a survival band; give_item and take_item are gated on ANY turn for ANY villager (a look is cheap there and it prevents the whole-reject door bounce outright). send_omen stays ungated (it addresses a group, not a pack). Rejected alternatives: universal gating (taxes every vision in the game with an extra round, including routine console play to healthy villagers) and pack-reaches-only (leaves A2's stated gap open — a console turn could still send an imperiled villager a vision contradicting their pack). Carried forward as TASK-199.
- A6 AGREED as written. AC#5 stays discharged structurally (SC-001 + SC-002); no change.
- A1, A3, A4, A5 were NOT addressed in this review and remain flagged as originally recorded. They stand as implemented; nothing blocks on them.

Both follow-ons approved for work: the storage/grant vocabulary mismatch (with an explicit operator instruction to CENTRALIZE the definitions rather than patch the symptom — magic strings are the failure mode being designed out) and the missing CLI take verb. Both carried forward as TASK-200.

2026-08-04 (same review, correction to the note above) — A1, A3, A4 and A5 are RATIFIED by the operator. The preceding note recorded them as 'not addressed'; that was my under-reading of the review, corrected here rather than by rewriting the earlier entry.

All six spec 116 Assumptions are therefore now resolved:
- A1 RATIFIED — the look-first gate is structural, not prose. Prompt-only remedies are foreclosed by the world-03 evidence (the carry-headroom digest line was present and read past).
- A2 OVERTURNED and widened — see the note above; peril-keyed visions, unconditional pack reaches. Carried forward as TASK-199.
- A3 RATIFIED — taken goods drop to a pile at the villager's feet; they are never unmade. Total goods conserved.
- A4 RATIFIED — take_item is priced and staged exactly as give_item (1 charge, excluded from the stage-1/2 ceiling).
- A5 RATIFIED — inspect_pack is granted at every stage; read-only, widens no acting capability.
- A6 AGREED as written — AC#5 stays discharged structurally via SC-001 + SC-002.

No assumption on spec 116 remains open. The only carried-forward work is TASK-199 (A2's widening) and TASK-200 (the two approved follow-ons).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Merged via PR #165 (merge commit 877045a5, a true two-parent merge — the in-branch wiki pin 5761edb is reachable from main, so no re-pin was staled). Spec 116, 46/46 tasks.

The guardian-side half of the world-03 death is closed. Three fixes, because three failures compounded: it could not see pack CONTENTS (only gross bulk), it did not look at all, and it had no way to empty a full pack.

1. inspect_pack — a charge-free Read tool (Effect Read / Gate None, off the turn's one-act budget, granted at every stage on the survey_site/explain/brief_myths profile) returning a deterministic sheet: every kind and count, carried and free bulk against sim.BulkCap, and an explicit statement of whether any food is carried. Backed by a new agentInv mirror refreshed in the same absorb batch as needMirror, with Spears/Axes deep-copied — the replica keeps mutating them, and a test pins the aliasing regression.

2. The look-first gate — on survival-origin turns ONLY, a send_vision or a work_miracle of kind give_item/take_item aimed at a villager whose pack was not opened THAT turn is refused as rejected_gate naming inspect_pack and the villager. Per-villager, per-turn, origin-keyed. Repairable inside the loop's round cap, so a skipped look costs one round, not a turn. Prose had already been tried and had already failed: the carry-headroom digest line WAS in the world-03 prompt and the guardian read past it.

3. take_item — a 1-charge miracle (guardian.item_taken, excluded from the stage-1/2 ceiling on the work_miracle precedent) lifting goods out of a pack into the pile on the villager's own tile via the agent.dropped rules. Reject-whole never clamp; spear/axe wear leaves most-worn-first; food gets a pile rot window; total units in (inventory + tile pile) are conserved. Every reach-in lands its event plus one situated first-person memory in a single atomic batch — no pack mutation is silent. IPC operator door has parity through the same BuildMiracleBatch.

AC#5 is discharged STRUCTURALLY, as recorded in spec Assumption A6: 'never receives a vision claiming they are carrying food' cannot be asserted against model prose, so its testable projection is SC-001 + SC-002 together — the guardian cannot send that villager a vision at all without first receiving a sheet that says, in words, that they carry no food.

GATES: go build/vet/test clean (23 packages); check-tui-design --changed passed; check-merge-drift pr exit 0; 82 wiki notes re-verified and re-pinned; 16 player docs fresh. Phase 8 ticks landed as derived state on branch spec-116-grounding-ticks (merged --no-ff at root, commit 2550e707).

GROUNDING FINDING corrected in-branch: the standing wiki claim 'a miracle never mints a new persistent entity' is now FALSE — take_item's pileFor is create-or-merge, so a removal onto a bare tile mints a sim.Pile. It is the only miracle kind that can; nothing is destroyed either way. Also corrected a spec clause of my own that was unsatisfiable against the code: carried food holds no spoilage (Inventory food fields are bare ints; only Pile.FoodBatch has SpoilAt), so a pack-to-pile transfer MINTS a rot window rather than preserving one.

TIER: Opus 5 via .claude/agents/spec-implementer-opus.md, delegated per constitution Principle V — doctrine-adjacent (new guardian write access into the world plus a new structural refusal).

OPEN FOR OPERATOR REVIEW (recorded, not blocking): spec Assumptions A1-A6, most notably A2 — the gate keys on turnOrigin.survival rather than message text, so it deliberately does NOT fire on an ordinary console turn where the guardian could also misread a pack.

FOLLOW-ONS flagged, not carded (operator's call): (a) vocabulary seam — the sheet renders storage names ('spears') while take_item's enum uses grant names ('spear'), so a model reading the sheet can take a repairable whole-refusal; (b) no CLI verb for removals — the IPC door has parity but 'promptworld work' has no 'take' verb.
<!-- SECTION:FINAL_SUMMARY:END -->
