---
name: guardian-designations-reception
description: Child of [[guardian-designations]] — how designations/directives reach villagers (the mental-map announcement grant, the DIRECTIVE reflex rung between SURVIVAL and PREP, the neverDrop directive context block) and the guardian-side turn surfaces (tool handlers, survey_site, TUI rendering, miracle rebase treatment, scope guards). Load for villager response mechanics or guardian-turn plumbing.
kind: component
sources:
  - internal/sim/plans.go
  - internal/sim/policy.go
  - internal/sim/executor.go
  - internal/guardian/plans.go
  - internal/guardian/survey.go
  - internal/mind/context.go
verified_against: 0af53ec6d211c71e298072c045c67ccbbd13b61d
---

# Guardian designations — reception, villager response, and turn surfaces

Child of [[guardian-designations]]: how a placed designation/directive
reaches a villager's mental map and reflex ladder, and the guardian-side
turn surfaces (tool handlers, `survey_site`, TUI rendering) that author and
display them. The parent covers the entity model and event vocabulary.

## The announcement grant

The `designation.placed` arm upserts one
`PlaceFact{Kind:"designation", Provenance: revealed, Seen: e.Tick}` at the
anchor tile (point tile / line first endpoint / rect min corner) into EVERY
living villager's mental map — the spec-041 place-grant machinery fanned
out reducer-side (one event, deterministic; map-less agents skipped, the
reducer stays total). `"designation"` joins the closed `PlaceFact`
vocabulary with a 7-game-day horizon (the max directive TTL);
`placeFactKinds` (send_vision's reveal enum) is NOT extended.
`renderKnownPlaces` renders the fact as an individually-named landmark
enriched from `State.Designations` (`designationLandmark`,
`internal/mind/prompt.go` — `PlaceFact.Detail` is an int64 scalar and
cannot carry a phrase). Cancellation/fulfillment never retracts the fact —
remembered history, the burned-out-fire precedent.

## The villager side — hardness (operator ruling, FIRM)

- **DIRECTIVE reflex rung** (`directiveDecision`,
  `internal/sim/policy.go`): SURVIVAL → **DIRECTIVE** → (prepYields?) PREP
  → wander. Unconditioned by the yield window and danger bands — survival
  always preempts (the rung sits after `survivalDecision`), directives
  preempt prep/wander (before the `prepYields` consult). The rung is
  STATELESS: it re-derives the oldest active directive addressing the
  agent from state at every idle decision, so interruption-resume needs
  zero new code (hails/conversations pause intents through the existing
  machinery; the rung simply re-resolves afterward). Routing: build at
  the site when reflex-expressible with materials in hand
  (`reflexBuildable` — the closed fire/chest/oven set); wall lines build
  the first gap in enumeration order (adjacent-stand) or walk toward it;
  zones walk to presence; otherwise walk to the site via the new
  `heed_directive` goal (instant on arrival, the `search` completion
  shape, never model-facing) or fall through to the ladder — the planner,
  reading the block, owns the clever part (no idle-at-site deadlock).
- **`directive` context block** ([[decision-context]],
  [[context-block-inventory]]): the eleventh spec-043 block, between
  `plan_echo` and `known_places`, priority `neverDrop` — ≤2 active
  directives addressing the agent, oldest first: framing text VERBATIM
  (recorded event data — the firewall's only prose channel, riding the
  issue batch's per-target "The Guardian charges you: …" companion
  memories exactly like visions), the bound designation's
  kind/site/requirement, plain-words days left. Empty renders `""` — a
  directive-free prompt is byte-identical to pre-084.

## Guardian-side surfaces

Tool handlers (`internal/guardian/plans.go`) parse the call, mint ids, and
land through `InjectSocial`, mapping door rejections to repairable
`rejected_gate` counsel (`planRefusal`); the turn prompt carries active
designations and directives (`writeDesignations`/`writeDirectives` — id,
kind, site, days-left, the `writeStandingOrders` shape) so counsel stays
truthful (FR-015). `survey_site` (`internal/guardian/survey.go`) is Effect
Read / Gate None — a deterministic site fact sheet (terrain mix, nearest
water/tree/rock, structures kind+tile, passability incl. wall-blocked
tiles) built turn-side from the absorb mirrors + static map (the
`buildTargetingDigest` pattern); out-of-bounds input returns a repairable
in-fiction miss naming the world bounds, never a hard error; radius clamps
1..8, default 4. On the TUI map, active designations render through three
appended [[tile-registry]] rows (`◇` site, `┄` wall-line segment, `◦` zone
perimeter — interior unmarked) beneath every real entity; consumed marks
stop rendering.

Time snap: an ACTIVE directive's `ExpiresTick` SHIFTs (the
`GuardianOrders` classification verbatim); everything else on both
entities is history/identity KEEP ([[guardian-miracle-rebase-taxonomy]]).

**Scope guards** (FR-016): no mission machinery (TASK-158), no faith
accounting (TASK-118 consumes the `directive.fulfilled` contract), no
bundle grammar/matrix change, no interruption-machinery change.

## Connections

Parent [[guardian-designations]] holds the entity model and event
vocabulary this reception layer consumes. [[decision-context]] and
[[context-block-inventory]] own the directive context block's slot;
[[reflex-policy]] hosts the DIRECTIVE rung among the other reflex rungs;
[[mental-maps]] owns the `PlaceFact` vocabulary the announcement grant
extends; [[tile-registry]] owns the three designation glyph rows;
[[guardian-miracle-rebase-taxonomy]] classifies the directive TTL's SHIFT
treatment.
