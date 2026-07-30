# Feature Specification: Canonization miracle — the guardian makes consensus lore real

**Feature Branch**: `task-81-canonization-miracle`

**Created**: 2026-07-29

**Status**: Draft

**Input**: TASK-81 — the "yes, and" answer to emergent mythology (Thornspire,
2026-07-23): instead of only letting reality debunk myths (spec 097, merged),
give the god the power to answer them. Realigned 2026-07-26: sibling of the
designation family — REUSE spec 084's durable-artifact entity machinery
(event-sourced guardian artifacts, deterministic IDs, one-way status, prune
discipline), never a parallel scheme. Depends: TASK-79 (Done), TASK-157 (Done),
spec 097 (merged — arrival observations CONFIRM canonized features).

## Decisions

- **D1 — Toponymy first: named regions as world state.** A named region is a
  guardian-authored durable artifact (spec 084 discipline): center + radius +
  villager-coined name, landing as a recorded `guardian.region_named` event
  through the normal door. Regions flow into describePlace/situated-memory
  place text, the chronicle's location vocabulary, and spec 097 observations.
- **D2 — Instantiation starts minimal: name-a-region + place-one-feature.**
  The canonize working can (a) christen a region, and/or (b) place ONE feature
  of an EXISTING placeable kind (structure kinds the build system already
  knows, e.g. standing-stone as a new decorative structure kind ONLY if an
  existing kind cannot serve — prefer existing kinds v1). No new flora/terrain
  taxonomy in v1.
- **D3 — Perception rides existing channels.** Discovery on arrival (spec 097's
  observation channel picks up the new feature/name — the next visit CONFIRMS
  instead of disconfirming) and/or an omen announcing it (existing send_omen).
  No new perception machinery.
- **D4 — Economy: canonization is a big act.** Charged per miracle doctrine at
  a premium (design target: 2 charges or a charge + cooldown — implementer
  picks within doctrine and records it); gratis/operator door unchanged.
- **D5 — Candidate briefing is read-only.** The guardian's myth briefing (AC5)
  derives dominant candidate lore from the existing rumor/belief corpus via a
  read-only tool/context surface — no new events, no self-grading.

## User Scenarios & Testing *(mandatory)*

### US1 - The player canonizes Thornspire (Priority: P1)

As a player whose villagers collectively invented "Thornspire at the forest's
edge", I want to tell the guardian to make it real — christen the region and
raise a feature there — so the myth becomes geography and the dreamers were
right.

**Acceptance Scenarios**:

1. **Given** a canonize working naming a region (center/radius/name), **When**
   the door validates and charges (D4), **Then** `guardian.region_named` lands
   (spec 084 artifact discipline: deterministic ID, one-way status), replay-
   deterministic.
2. **Given** an optional feature placement, **Then** it lands through the
   existing entity/build placement rules (valid site validated pre-charge).
3. **Given** the named region, **Then** situated place descriptions and the
   chronicle use the villager-coined name for positions within it.

### US2 - The village discovers the miracle (Priority: P1)

1. **Given** a canonized region/feature, **When** a villager next arrives
   (spec 097 channel), **Then** the observation CONFIRMS beliefs naming it
   (the disconfirmation loop inverts — myth becomes verified knowledge).
2. **Given** an omen sent about it, **Then** existing omen rules apply
   unchanged.

### US3 - The guardian briefs on candidate myths (Priority: P2)

1. **Given** the belief/rumor corpus, **When** the player asks (or the
   guardian's turn context includes the briefing surface), **Then** dominant
   place-myths are summarized read-only (D5) as canonization candidates.

### Edge Cases

- Region name collisions: second christening of an overlapping region refuses
  at the door (one name per ground truth; renames are future work).
- Replay: pre-101 logs byte-identical (additive events, spec 094 doctrine —
  no format bump); recorded canonizations re-apply cleanly.
- The live playtest world is never touched; acceptance demo runs on a seeded
  measure world with a manufactured myth (implanted beliefs), NOT world-01
  (the card's world-01 wording predates the playtest freeze).

## Requirements *(mandatory)*

- **FR-001**: Named regions as spec-084-discipline artifacts on sim.State,
  landing via `guardian.region_named`; flow into place text, chronicle, and
  spec 097 observations (D1).
- **FR-002**: Canonize working: name-region and/or place-one-existing-kind
  feature, validated pre-charge, premium-charged (D2, D4).
- **FR-003**: Discovery via existing channels only (D3).
- **FR-004**: Read-only myth briefing surface (D5).
- **FR-005**: Digest + event-types.md; TestCatalogSweep; tool gloss per the
  TASK-163 pattern.
- **FR-006**: Tests: door validation/refusal, artifact discipline, place-text
  integration, spec-097 confirm path, replay byte-identity; -race green.
- **FR-007**: Acceptance demo on a seeded measure world: implant a place-myth,
  canonize it, observe the next arrival CONFIRM; evidence at
  docs/design/evidence/task-81/.

## Success Criteria *(mandatory)*

- **SC-001**: The demo shows the full loop: myth (implanted) → brief → canonize
  → arrival confirms → chronicle narrates with the coined name.
- **SC-002**: Zero new perception/entity schemes (grep-level: region artifacts
  use the 084 shape; discovery uses 097 events).
- **SC-003**: Replay fixtures byte-identical.

## Assumptions

- Tier: Sonnet (card + guardian-directives draft tier call: narrow reducer arm
  + one tool, reusing 157's machinery); escalate if toponymy forces
  worldmap/state restructuring beyond an additive regions field.
