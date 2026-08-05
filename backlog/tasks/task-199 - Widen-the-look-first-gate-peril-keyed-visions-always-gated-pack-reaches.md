---
id: TASK-199
title: 'Widen the look-first gate: peril-keyed visions, always-gated pack reaches'
status: To Do
assignee: []
created_date: '2026-08-05 02:36'
labels:
  - guardian
  - tools
  - survival
dependencies: []
priority: high
ordinal: 181001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The Guardian is currently only forced to look inside a villager's pack when the game wakes it
because someone is dying. This card widens that discipline so it also holds when the player is
the one driving — a god that misreads what you are carrying is no less wrong for having been
asked a question instead of woken by an alarm.

Spec 116 shipped the look-first gate keyed on the survival watch turn. Operator review of that
spec's Assumption A2 (2026-08-04) overturned the narrow scope: the trigger becomes the
villager's PERIL, not which turn woke the Guardian, and reaching into a pack requires a look
every time regardless.

## Use cases

- As a player, when I ask the Guardian to send a word to a villager who is starving, I want it
  to have looked in that villager's pack first — the same care it takes when the alarm wakes
  it, because my asking does not make the villager any less close to death.
- As a player, when I ask the Guardian to speak to a perfectly healthy villager, I do not want
  to pay for a pack inspection that has nothing to do with what I asked.
- As the Guardian, when I am about to put something into a villager's hands or take something
  out of them, I want to have looked in the pack first every single time, so my working is
  never refused at the door for a fullness I could simply have checked.

## The ratified rule

- `send_vision` at villager V — gated on ANY turn origin when V is in a survival band
  (near-death / starvation / exposure, the same bands the watches match).
- `give_item` / `take_item` at villager V — gated on ANY turn, for ANY villager, healthy or
  not. A look is free and it prevents the whole-reject door bounce outright.
- `send_omen` — never gated. It addresses a group, not a pack.
- Unchanged: the refusal stays a repairable `rejected_gate` naming `inspect_pack` and the
  villager, so the model looks and retries inside the loop's round cap; the ledger stays
  per-turn and per-villager; the door remains the semantic authority.

## Rejected alternatives, with the reason each lost

- **Universal gating** (every vision, every villager, every turn): simplest rule to state, but
  it taxes every vision in the game with an extra round, including routine console play to
  healthy villagers. The harm being prevented is misreading an imperiled villager's pack; a
  healthy villager told something inaccurate about their inventory is a far smaller wrong.
- **Pack reaches only** (leave visions on the survival-origin trigger): cheapest, but it leaves
  Assumption A2's stated gap open — a console turn could still send a starving villager a
  vision contradicting what they carry, which is exactly the world-03 failure.

## Notes for the specifier

- The existing gate lives in `internal/guardian/toolcalls.go` (`turnDispatch.survival`,
  `turnDispatch.looked`, `lookFirstRefusal`, `packReachingKind`) and is documented in
  `docs/wiki/guardian-survival-watches.md` and `docs/wiki/guardian-turn-loop.md`.
- The survival bands already exist as a reusable predicate: `survivalBand(kind, needMirror)` in
  `internal/guardian/orders.go`, the same one `survivalFlag` uses to annotate the targeting
  digest. Peril-keying should read THAT, not a second threshold literal — the bands the gate
  fires on must never drift from the bands the watches match.
- `turnDispatch.survival` may become unnecessary for the vision arm once the trigger is the
  villager's condition; check whether it still earns its place before leaving it in.
- Spec 116's Assumption A2 is superseded by this card. The new spec should say so explicitly
  and record the two rejected alternatives above, so a future reader finds the reasoning rather
  than re-deriving it.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 send_vision at a villager in a survival band is gated on ANY turn origin, not only a survival watch turn
- [ ] #2 send_vision at a villager NOT in a survival band is ungated on an ordinary console turn — no inspection is demanded for a routine word to a healthy villager
- [ ] #3 give_item and take_item are gated on ANY turn for ANY villager, healthy or imperiled
- [ ] #4 send_omen remains ungated on every turn
- [ ] #5 The survival bands the gate fires on are read from the SAME survivalBand predicate the watches match — no second threshold literal exists in the codebase
- [ ] #6 The refusal remains a repairable rejected_gate naming inspect_pack and the villager, and a corrected call lands within the same turn's round cap
- [ ] #7 Regression: the world-03 case is caught on a console turn, not only on a survival watch turn
<!-- AC:END -->
