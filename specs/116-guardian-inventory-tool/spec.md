# Feature Specification: Guardian Inventory Tool

**Feature Branch**: `task-197-guardian-inventory-tool`

**Created**: 2026-08-03

**Status**: Draft

**Board Task**: TASK-197

**Input**: Board card TASK-197 — "Guardian inventory tool: let the Guardian read and change
what a villager carries, and make it look first."

## Overview

The Guardian can read a villager's health, hunger and warmth, and it can put things into a
villager's hands. It cannot look inside the pack it is reaching into, and it cannot take
anything out. Both halves of that gap killed a man in world-03.

Cedar was dying — health 24/1000, starving — and carrying twenty wood and four planks. His
pack was full: 24 of 24 bulk used, zero free. The Guardian sent him a vision reading *"You
are carrying food in your own hands this very moment… eat now, whatever you hold."* Cedar was
carrying no food at all. Across both of its watch turns the Guardian spent every call on
`survey_site` and `explain` — it looked at the ground eight times and never once at the man.
It never attempted a miracle, and could not have helped if it had: `give_item` rejects an
over-cap grant **whole** rather than clamping (`internal/sim/miracles.go:463`), and Cedar had
zero free bulk, so the one intervention that would have saved him was sealed shut by the same
full pack nobody could see.

Three distinct failures compound here, and all three must be addressed:

- **The Guardian cannot see pack CONTENTS.** `buildTargetingDigest`
  (`internal/guardian/turn.go:1255`) already prints `carrying 24/24, 0 free to receive` per
  living villager, and the guardian's mirror (`mirrorState`, `internal/guardian/guardian.go:488`)
  carries only that gross `Bulk` integer. Gross bulk cannot distinguish the three cases that
  demand three different interventions: **has no food**, **has food and will not eat**, and
  **has no room to pick food up**. Cedar was the third; the Guardian treated him as the second.
- **Prose did not make it look.** The digest line *was there* in world-03, in the prompt, and
  the Guardian read past it. A fact buried in a targeting digest is evidently not enough; the
  looking has to be a call the Guardian makes, and — where a life is at stake — one it cannot
  skip.
- **The Guardian has no way to empty a pack.** Its whole write surface into an inventory is
  `give_item`, which only adds. A full pack is therefore a locked door: the Guardian can see
  the lock (`0 free to receive`) and has no key.

This feature delivers all three: a read tool the Guardian calls to open a villager's pack, a
structural look-first gate on the survival watch turn, and a `take_item` miracle that lifts
goods out of a pack and sets them at the villager's feet — every reach-in leaving a durable
event and a situated first-person memory, never a silent mutation.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The Guardian can open a villager's pack (Priority: P1)

**As the Guardian, when a villager is starving, I want to open their pack and see exactly what
they carry, so I can tell the difference between "has no food", "has food and won't eat", and
"has no room to pick food up".**

The Guardian calls `inspect_pack(villager)` and receives a deterministic fact sheet naming
every kind it carries and how many of each, its carried bulk, and its free bulk against
`sim.BulkCap`. The sheet is free: no charge moves, no event lands, and the call does not spend
the turn's one act — looking is looking, not an act (the `survey_site` / `explain` posture).

**Why this priority**: nothing else in the feature is usable without it. The write path is
guesswork without a read, and the look-first gate has nothing to gate on.

**Independent Test**: call `inspect_pack` against a fixture villager whose pack is full of
non-food (Cedar's twenty wood and four planks) and assert the sheet names `wood 20`,
`planks 4`, `carrying 24/24`, `0 free`, and states that no food is carried. Call it twice with
identical arguments and assert byte-identical text.

**Acceptance Scenarios**:

1. **Given** a living villager carrying 20 wood and 4 planks and nothing else, **When**
   `inspect_pack` names them, **Then** the sheet lists exactly those two kinds with their
   counts, reports `carrying 24/24 — 0 free`, and says in plain words that they carry no food.
2. **Given** a living villager carrying nothing, **When** `inspect_pack` names them, **Then**
   the sheet says the pack is empty and reports `carrying 0/24 — 24 free`.
3. **Given** a name that is not a living villager, **When** `inspect_pack` names it, **Then**
   the sheet is an honest in-fiction miss naming the living roster — never a hard error (the
   `explain` unknown-topic shape).
4. **Given** two identical `inspect_pack` calls in one turn, **Then** both sheets are
   byte-identical.

---

### User Story 2 - The Guardian looks before it speaks (Priority: P1)

**As a player, when the Guardian speaks to my villager, I want its words to match what that
villager is really carrying — a god that misreads the pack is worse than a silent one.**

On a **survival watch turn** — the turn the Guardian's own nature wakes it for, where a
villager may die and a wrong word is a death — the Guardian may not send a vision to, or work
a grant/removal upon, a villager whose pack it has not opened **in that same turn**. A call
that skips the looking is refused at the door as a `rejected_gate` naming the exact repair, so
the model calls `inspect_pack` and retries within the loop's round cap. This is the
established refusal-with-counsel pattern, not a new failure mode: a corrected call lands
normally.

**Why this priority**: this is the failure that killed Cedar. The read tool alone would have
sat unused exactly as the digest line did.

**Independent Test**: run a survival-origin turn against the Cedar fixture; assert a
`send_vision` at Cedar without a prior `inspect_pack(Cedar)` is refused with a reason naming
`inspect_pack`, and that the same call after an `inspect_pack(Cedar)` lands.

**Acceptance Scenarios**:

1. **Given** a survival watch turn woken for Cedar, **When** the model calls `send_vision`
   targeting Cedar with no prior `inspect_pack(Cedar)` this turn, **Then** the call is
   `rejected_gate` with a reason instructing it to open Cedar's pack first, nothing lands, and
   no charge is spent.
2. **Given** the same turn after `inspect_pack(Cedar)` has returned, **When** `send_vision`
   targets Cedar, **Then** it lands exactly as it would have before this feature.
3. **Given** a survival watch turn, **When** the model calls `work_miracle(give_item)` or
   `work_miracle(take_item)` at a villager whose pack it has not opened this turn, **Then**
   the call is refused the same way.
4. **Given** an ordinary (non-survival) console or system turn, **When** `send_vision` targets
   anyone, **Then** the gate does not apply and behavior is unchanged.
5. **Given** a survival turn in which `inspect_pack(Hazel)` was called, **When**
   `send_vision` targets Cedar, **Then** it is refused — the gate is per-villager, not
   per-turn.

---

### User Story 3 - The Guardian can empty a pack it means to fill (Priority: P2)

**As the Guardian, when I find a starving villager whose hands are too full to accept a gift,
I want to be able to lift the deadweight out of their pack, so my miracle is not refused at
the door.**

A new miracle kind, `take_item`, removes a named quantity of a named kind from a living
villager's inventory and sets it down as a pile on the tile they stand on — the goods are
*moved*, not unmade, so nothing of the village's labour is destroyed and the villager (or a
neighbour) can pick them back up. It is priced and gated exactly as `give_item` is: one
charge, world-shaping, absent from the stage-1/stage-2 ceiling.

Removal is **reject-whole, never clamp**, matching `applyItemGranted`'s discipline: a request
to take more than is carried is refused naming the actual count, so the Guardian corrects the
quantity rather than receiving a silent partial.

**Why this priority**: this is the half that turns a diagnosis into a rescue. Without it the
Guardian can see Cedar's locked pack and still do nothing about it.

**Independent Test**: with a fixture villager at 24/24 carrying 20 wood, apply
`take_item(villager, wood, 12)`; assert the inventory drops to 12 bulk, a pile at their tile
holds 12 wood, one charge is spent, and a subsequent `give_item(food_cooked, 4)` — which was
refused before — now lands.

**Acceptance Scenarios**:

1. **Given** Cedar at 24/24 carrying 20 wood and 4 planks, **When** the Guardian works
   `take_item(Cedar, wood, 20)`, **Then** Cedar carries 4/24, a pile on Cedar's tile gained
   20 wood, and one charge was spent.
2. **Given** that same state, **When** the Guardian then works `give_item(Cedar, meals, 4)`,
   **Then** the grant lands (it would have been rejected whole before the removal).
3. **Given** Cedar carrying 20 wood, **When** the Guardian works `take_item(Cedar, wood, 30)`,
   **Then** the miracle is rejected whole naming what he actually carries, no charge is spent,
   and his inventory is untouched.
4. **Given** a tile that already holds a pile, **When** a removal lands on it, **Then** the
   goods merge into the existing pile (one pile per tile), exactly as a villager's own drop
   does.
5. **Given** a dead villager or an unknown item kind, **When** a removal names them, **Then**
   it is rejected at the door with an in-fiction reason.

---

### User Story 4 - Nothing reaches into a pack in silence (Priority: P2)

**As a villager, when the Guardian reaches into my pack and takes or leaves something, I want
to remember that it happened, and my neighbours to be able to hear of it.**

Every guardian reach-in — grant or removal — lands as one atomic batch through the injection
door: the durable `guardian.item_granted` / `guardian.item_taken` event, plus one
`agent.memory_added` written in the villager's own first person, in the language of their
world (no player, no game, no outside voice). The memory is a memory like any other: it can be
recalled, consolidated, spoken of, and carried into a conversation with a neighbour. A pack
mutation with no memory attached is a contract violation, not an optimisation.

**Why this priority**: it is the doctrine condition on write access existing at all. The
Guardian gaining a new way to reach into the world is only acceptable if every reach leaves a
trail a human can read afterwards.

**Independent Test**: land a `take_item` and assert the emitted batch contains exactly the
`guardian.item_taken` event and one `agent.memory_added` for the affected villager, and that
replaying the batch reproduces the same state.

**Acceptance Scenarios**:

1. **Given** a landed removal, **Then** the batch carries `guardian.item_taken` and exactly
   one `agent.memory_added` addressed to that villager, at the same salience a grant memory
   uses.
2. **Given** a landed removal, **Then** the villager's memory text is first-person and
   in-fiction, naming what left their pack and where it went.
3. **Given** any `inspect_pack` call, **Then** a `cog.tool_call` record is emitted with a read
   verdict — the looking itself is auditable — and **no** world event lands.
4. **Given** a rejected removal, **Then** the `cog.tool_call` record carries the refusal reason
   and no `guardian.item_taken` and no memory exist.

---

### Edge Cases

- **Spears and axes.** These are slices of remaining-uses, not counts. The read sheet reports
  them as a count plus their remaining uses; a removal takes the **most-worn first** (the front
  of the ascending slice), matching how `agent.dropped` and hunts already spend them, so the
  villager keeps their best tools and the pile receives the worn ones.
- **Food batches.** A villager's food carries spoilage; a pile stores food as `FoodBatch`
  entries. A removal must preserve batch identity into the pile exactly as a villager's own
  drop does, so taken food spoils on its original schedule rather than being refreshed.
- **The tile is not empty of a pile.** One pile per tile: a removal merges into the existing
  pile (create-or-merge), never a second pile.
- **A removal that empties a pack entirely** is legal — the pack simply reads `0/24`.
- **The looking is stale by the time the vision lands.** The sheet is a snapshot of the
  mirrored replica at call time; the world keeps ticking. The gate proves the Guardian *looked
  this turn*, not that nothing changed in the intervening milliseconds — the door remains the
  semantic authority for what actually lands.
- **Survival turn where the endangered villager is already dead** by the time the turn runs:
  `inspect_pack` returns the honest miss (the living-roster shape), and the existing dead-target
  refusals still apply — the gate never traps the Guardian in an unsatisfiable loop.
- **Stage-1 and stage-2 worlds.** `inspect_pack` is granted (read-only, widens nothing);
  `take_item` is not (world-shaping, the `work_miracle` precedent). The look-first gate is
  therefore vacuous at those stages for miracles, and still binding for visions.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001 (READ TOOL)**: A guardian-facing tool `inspect_pack` MUST join the guardian tool
  registry with Effect Read and Gate None, taking a single required `villager` argument, and
  MUST be wired as a handler in `turnHandlers` under the same grant gating every other tool
  uses (structural absence when ungranted).
- **FR-002 (SHEET CONTENT)**: `inspect_pack`'s sheet MUST report, for the named living
  villager: every inventory kind they carry and its count; the carried bulk and the free bulk
  against `sim.BulkCap`; and an explicit statement of whether any edible kind is carried.
- **FR-003 (DETERMINISM)**: The sheet MUST be a pure function of (arguments, mirrored state)
  with a fixed kind ordering and no clock reads, so identical calls in one turn return
  byte-identical text (the `survey_site` contract).
- **FR-004 (FREE LOOKING)**: `inspect_pack` MUST spend no charge, land no world event, and
  MUST NOT consume the turn's one mediated act.
- **FR-005 (HONEST MISS)**: An unknown or non-living `villager` MUST return a read-ok sheet
  naming the living roster, never a hard error.
- **FR-006 (CONTENTS MIRROR)**: The guardian's `mirrorState` snapshot MUST carry per-villager
  inventory CONTENTS alongside the existing gross `Bulk`, refreshed in the same absorb batch,
  read by the turn worker under `stateMu` and never from the replica directly.
- **FR-007 (LOOK-FIRST GATE)**: On a turn whose origin is the survival watch, a `send_vision`
  targeting villager V, or a `work_miracle` of kind `give_item` or `take_item` upon V, MUST be
  refused as `rejected_gate` unless `inspect_pack(V)` returned earlier in the SAME turn's tool
  loop. The refusal reason MUST name `inspect_pack` and the villager, so the model can repair
  within the round cap.
- **FR-008 (GATE SCOPE)**: The look-first gate MUST apply per-villager (looking at one does not
  license acting on another) and MUST NOT apply on non-survival turns. It MUST NOT inspect
  message text to decide whether a vision "concerns food" — the survival origin is the trigger.
- **FR-009 (REMOVAL MIRACLE)**: A miracle kind `take_item` MUST join the miracle vocabulary
  with the same charge price as `give_item`, mapping to a new event type
  `guardian.item_taken`, taking `villager`, `item`, and `qty`.
- **FR-010 (REJECT WHOLE)**: A removal that names more units than the villager carries, an
  unknown item kind, a non-positive quantity, a dead villager, or an out-of-range index MUST be
  rejected WHOLE at the reducer — no charge spent, no partial application — with a reason
  naming what is actually carried.
- **FR-011 (GOODS ARE MOVED, NOT UNMADE)**: A landed removal MUST transfer the removed units
  into the pile on the villager's own tile (create-or-merge, one pile per tile), preserving
  spear/axe wear ordering (most-worn first) and food batch identity, exactly as `agent.dropped`
  does.
- **FR-012 (SHARED BATCH BUILDER)**: `BuildMiracleBatch` MUST gain the `take_item` arm so the
  operator's IPC `miracle` door and the guardian's `landMiracle` compose an identical batch and
  can never drift (spec 016 R6).
- **FR-013 (SITUATED MEMORY)**: Every landed removal MUST include exactly one
  `agent.memory_added` for the affected villager, first-person and in-fiction, at the same
  salience and origin a grant memory uses, in the same atomic batch as the removal event.
- **FR-014 (AUDIT TRAIL)**: Every `inspect_pack` call MUST produce a `cog.tool_call` record;
  every refused reach-in MUST produce one carrying the refusal reason. No pack mutation may
  land without its event and memory.
- **FR-015 (STAGE CEILING)**: `inspect_pack` MUST join the stage-1 ceiling roster (read-only,
  the `survey_site` / `brief_myths` precedent); `take_item` MUST NOT (world-shaping, the
  `work_miracle` precedent).
- **FR-016 (GUIDANCE)**: The tool guidance surfaces the model reads MUST describe
  `inspect_pack` and `take_item`, and `give_item`'s existing carry-headroom gloss MUST name
  `take_item` as the remedy for a full pack.
- **FR-017 (RENDERING)**: `guardian.item_taken` MUST render in the chronicle digest as a
  first-class event with a plain-language summary, not fall through as an unknown type.
- **FR-018 (REPLAY)**: A recorded `guardian.item_taken` MUST re-apply cleanly during
  reconstruction, producing byte-identical state — the reducer arm validates before it mutates.

### Key Entities

- **Pack sheet** — the deterministic text `inspect_pack` returns. Not stored; composed
  turn-side from the mirrored contents snapshot, the `survey_site` pattern.
- **Inventory contents mirror** — a per-villager snapshot of `sim.Inventory` held beside the
  existing `needMirror`, refreshed on every absorb batch.
- **Look-first ledger** — per-turn, in-memory set of villager indices `inspect_pack` has
  returned for during this tool loop. Lives on the turn's dispatch state; never persisted.
- **`guardian.item_taken`** — the durable removal event: villager ref, item kind, quantity,
  gratis flag. The mirror image of `guardian.item_granted`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Against the world-03 fixture (villager at 24/24 carrying 20 wood + 4 planks,
  starving), `inspect_pack` returns a sheet naming both kinds, `24/24`, `0 free`, and "no
  food" — and the same call twice returns byte-identical bytes.
- **SC-002**: On a survival-origin turn, 100% of `send_vision` / `give_item` / `take_item`
  calls aimed at a villager whose pack has not been opened that turn are refused, and the
  refusal names `inspect_pack`.
- **SC-003**: On the same fixture, the sequence `inspect_pack` → `take_item(wood, 20)` →
  `give_item(meals, 4)` succeeds end to end, where `give_item(meals, 4)` alone is rejected
  whole.
- **SC-004**: Every landed `guardian.item_taken` in the test suite is accompanied by exactly
  one `agent.memory_added` for the affected villager; a removal with no memory fails the suite.
- **SC-005**: A recorded removal replays to byte-identical state, and the total units in
  (villager inventory + tile pile) is unchanged by the removal — nothing is created or
  destroyed.
- **SC-006**: `go test ./...` passes, and the guardian package's own tests exercise
  `inspect_pack` against a fixture villager whose pack is full of non-food.

## Assumptions

These are resolved from existing artifacts and repo doctrine rather than re-asked, per the
project's artifact-grounded-action rule. Each is **flagged for operator review**, following the
spec 084 / 085 / 107 precedent for capability-surface decisions.

- **A1 — The look-first gate is structural, not prose.** The card's own evidence forecloses a
  prompt-only remedy: the carry-headroom digest line was present in world-03 and the Guardian
  read past it. The repo's doctrine is gates over prose ("a status can never exceed the
  artifacts that prove it"), and `rejected_gate` is already the repairable-refusal channel, so
  a skipped look costs one extra round, not a turn.
- **A2 — The gate keys on the survival turn origin, not on message text.** `turnOrigin.survival`
  is already a first-class flag (`internal/guardian/orders.go:714`). Classifying whether a
  vision "concerns food" would require reading model prose at the door — non-deterministic,
  unauditable, and untestable. On a survival turn every word concerns the peril by construction,
  so the origin is the honest trigger. Consequence, stated plainly: the gate does **not** fire
  on an ordinary console turn where the Guardian might also misread a pack. That is the
  deliberate narrow scope — the death being fixed happened on a survival turn.
- **A3 — Taken goods drop to a pile; they are not unmade.** The card asks to "lift the
  deadweight out of their pack", not to destroy the village's labour. The `agent.dropped`
  reducer arm already performs exactly this inventory→pile transfer with create-or-merge, wear
  ordering, and food-batch preservation, so removal reuses proven machinery, conserves total
  goods (SC-005), and leaves the pile itself as a visible artifact of the reach-in.
- **A4 — `take_item` is priced and staged as `give_item` is** — one charge, absent from the
  stage-1/2 ceiling. It is the same class of act (a world-shaping reach into a pack) pointed
  the other way, so a different price or stage would need a reason the artifacts do not supply.
- **A5 — `inspect_pack` is granted at every stage.** It is read-only and widens no acting
  capability, matching `survey_site`, `explain`, and `brief_myths` exactly.
- **A6 — AC#5 is verified structurally.** "A starving villager at 24/24 carrying zero food
  never receives a vision claiming they are carrying food" cannot be asserted against model
  prose. Its testable projection is the conjunction of SC-001 and SC-002: the Guardian cannot
  send that villager a vision at all without first receiving a sheet that says, in words, that
  they carry no food. The spec claims that projection and no more.
