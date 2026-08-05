---
id: TASK-200
title: 'One authority for item kinds, plus the CLI take verb'
status: To Do
assignee: []
created_date: '2026-08-05 02:37'
labels:
  - tools
  - refactor
  - cli
dependencies: []
priority: high
ordinal: 182001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The game currently spells the same thing two different ways depending on which part of the code
is asking — a villager's pack says "spears", the Guardian's tool for taking things says "spear",
and a Guardian that reads the first and types the second gets refused. This card makes one
authoritative list of what a thing is called, so the two can never disagree again, and adds the
missing operator command for taking goods out of a pack.

Two deliverables, one surface: both touch how an item kind is named and parsed, so doing them
apart would mean two passes over the same files.

## Use cases

- As the Guardian, when I read that a villager is carrying "spears" and then ask to take one, I
  want that to work — not to be refused because the taking verb spells it "spear".
- As a player, I want the Guardian to stop wasting a turn's round on a refusal that only exists
  because two parts of the game disagree about a word.
- As the operator, when I want to lift something out of a villager's pack from the command line
  — the same working the Guardian can already do, and that I can already do over IPC — I want a
  verb for it instead of being told the kind is unknown.
- As a developer, when I add a new kind of thing to the world, I want to declare its names once
  and have every surface pick them up, instead of hunting down hand-written string lists.

## Deliverable 1 — one authority for item kinds

Today the vocabularies are separate hand-maintained literals that must be kept in step by
convention and drift tests:

- `internal/tool/registry.go` — `itemKinds` (storage vocabulary, plural `spears`/`axes`) and
  `grantKinds` (grant vocabulary, singular `spear`/`axe`), deliberately distinct and documented
  as "do not merge the two lists".
- `internal/guardian/pack.go` — `packKinds` (a third ordering) plus a `packCount` switch over
  bare strings.
- `internal/sim` — `Inventory` struct fields, `grantableKinds`, `addItems`, `bulk()`, the
  `agent.dropped` arm's `"spears"` special-case, and `applyItemTaken`'s `"spear"`/`"axe"` arms.

The two vocabularies exist for a real reason (a pack holds many spears, each with its own
remaining uses; a grant delivers one fresh spear), so the fix is NOT to collapse them into one
string list. It is to make them two projections of ONE declared kind, so the relationship is
expressed once in data instead of maintained by hand in five places — and so a storage name can
always be resolved to its grant name and back.

The operator's standing instruction, recorded verbatim: centralize the definitions, do not
patch the symptom. Magic strings are the failure mode being designed out.

## Deliverable 2 — the CLI take verb

`take_item` shipped in spec 116 with a Guardian door and an IPC door, but `promptworld work` has
no `take` verb, so an operator cannot reach it from the command line. Add it alongside the
existing grant verb, parsing its item kind through the centralized vocabulary from deliverable 1.

## Notes for the specifier

- The mismatch is live and reachable today: `inspect_pack`'s sheet renders `- spears: 2 (uses
  left: 3, 7)` while `take_item`'s `item` argument accepts `tool.GrantKinds()`, i.e. `spear`.
  A Guardian that reads the sheet and echoes the word it saw takes a whole-refusal. It is
  repairable (the rejection enumerates the accepted kinds) but it is a wasted round, and it was
  flagged by the spec 116 implementer as a new place a pre-existing split can bite.
- Whatever shape the centralization takes, the existing drift tests (`TestGrantKindsMirrorTool`,
  `TestMiracleKindsMirrorTool`, and the `ItemKinds` cross-checks) should end up ENFORCED BY
  CONSTRUCTION rather than merely re-pointed — if the vocabularies can no longer disagree, the
  tests should be asserting that impossibility, not re-checking two lists.
- `internal/tool` is a leaf package (it cannot import `internal/sim` or `internal/guardian`), so
  the authority has to live somewhere both can reach. That constraint is what produced the
  mirrored-list-plus-drift-test pattern in the first place; the spec should confront it directly
  rather than rediscover it.
- Watch the food kinds while centralizing: carried food is a bare count on `sim.Inventory`,
  while a pile stores `FoodBatch` entries with a `SpoilAt`. The "is this kind food" question is
  currently answered by at least two separate lists.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The storage vocabulary (spears/axes) and the grant vocabulary (spear/axe) are two projections of one declared kind, not two independently hand-maintained string lists
- [ ] #2 A storage name resolves to its grant name and back through the shared authority, so the inspect_pack sheet and take_item's item argument can never disagree
- [ ] #3 Naming a kind by its storage spelling in take_item either works or is refused with a message naming the correct spelling — a Guardian echoing the sheet is never silently wrong
- [ ] #4 internal/guardian/pack.go's packKinds ordering and packCount string switch are derived from the shared authority, not a third hand-written list
- [ ] #5 The existing vocabulary drift tests assert the impossibility of disagreement by construction, rather than re-checking two parallel lists
- [ ] #6 internal/tool remains a leaf package — the centralization does not introduce an import from tool into sim or guardian
- [ ] #7 promptworld work gains a take verb reaching take_item through the same shared batch builder the guardian and IPC doors use, parsing its item kind through the centralized vocabulary
- [ ] #8 go test ./... passes and no behavior changes for any existing kind: grants, drops, hunts, and harvests spend and store exactly as before
<!-- AC:END -->
