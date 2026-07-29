# Feature Specification: Carry-cap headroom guidance for give_item

**Feature Branch**: `task-167-give-item-headroom`

**Created**: 2026-07-29

**Status**: Draft

**Input**: TASK-167 — carded from TASK-163's evidence (2/5 residual rejections:
give_item refused on the carry cap with CORRECT item kinds; one same-turn
self-repair observed). Constraint: spec 016 FR-011 — reject whole, never clamp.

## Decision (card AC #1)

**Chosen: (a) live carry headroom in the miracle-capable prompt digest**, plus a
one-line static note in the give_item gloss. The spec 059 positions/passability
digest already flows live per-villager state into miracle-capable turns — adding
each villager's carry headroom (used/cap) there gives the model the number BEFORE
it picks a quantity, preventing the wasted attempt.

- Rejected (b)-only: a static gloss cannot carry per-villager live state — it can
  only say "small quantities fit", which the evidence shows the model effectively
  already learns from the refusal.
- Rejected (c) leave-as-is: the teaching door + repair loop works but costs a
  wasted attempt (and one observed non-repair); first-try landing matters because
  TASK-164's charter-delta re-run runs next and every rejection class left in the
  door is noise in its outcome attribution.
- FR-011 honored: the door is untouched — no clamping, no door-side changes at
  all. This is guidance/context only.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The guardian grants what fits, first try (Priority: P1)

As the guardian deciding to feed a villager, I want the villager's live carry
headroom visible in my turn context, so my give_item quantity fits on the first
attempt instead of bouncing off the cap and retrying.

**Acceptance Scenarios**:

1. **Given** a miracle-capable guardian turn, **When** the prompt digest is
   assembled, **Then** each villager's line carries carry headroom (used/cap or
   free units) from live state, deterministic from the snapshot.
2. **Given** a full villager, **When** the model reads the digest, **Then** the
   headroom shows 0 free — the model can choose a different response (smaller
   qty is impossible; explain or act otherwise) without a door round-trip.
3. **Given** the give_item gloss, **Then** it carries one added line noting the
   digest's headroom field and that oversized grants are refused whole (FR-011).

---

### Edge Cases

- Digest size: headroom adds a few tokens per villager; the digest stays within
  its existing budget shape (no new sections — extend the existing per-villager
  line).
- Dead villagers: no headroom shown (existing digest rules for dead agents
  unchanged).
- The door and its refusal message are byte-unchanged; replay untouched (no new
  events, no reducer changes).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The miracle-capable turn digest's per-villager line includes live
  carry headroom (free bulk units and cap), read from the same snapshot as
  positions.
- **FR-002**: The give_item guidance gloss gains one line referencing the
  headroom field; no other tool-surface changes.
- **FR-003**: The door (internal/sim/miracles.go applyItemGranted cap check) is
  untouched; FR-011 reject-whole semantics preserved (test asserts message and
  behavior unchanged).
- **FR-004**: Tests: digest contains headroom for living villagers (correct
  arithmetic vs inventory bulk), gloss line present, door behavior regression.

## Success Criteria *(mandatory)*

- **SC-001**: A probe-style unit test shows the digest carrying correct headroom
  for a villager with a partially-filled inventory.
- **SC-002**: Zero door/reducer diff (git diff shows no internal/sim/miracles.go
  behavior change).
- **SC-003**: Card AC#2 satisfied via the change + tests; live first-try landing
  is expected to be demonstrated opportunistically during TASK-164's re-run
  (recorded there), not by a dedicated probe world.

## Assumptions

- The spec 059 digest assembly is in the guardian turn context path
  (internal/guardian, near the positions digest); carry bulk/cap live on the
  agent inventory per spec 013 (executor-world-state note).
- Sonnet tier: single-surface guidance/digest change with tests.
