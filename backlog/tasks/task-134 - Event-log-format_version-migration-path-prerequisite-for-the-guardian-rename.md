---
id: TASK-134
title: >-
  Event-log format_version + migration path (prerequisite for the guardian
  rename)
status: In Progress
assignee: []
created_date: '2026-07-25 19:29'
updated_date: '2026-07-29 19:22'
labels:
  - replay-doctrine
  - review-2026-07-25
  - guardian-rename
dependencies: []
priority: high
ordinal: 104000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Operator decision (2026-07-25, team review): the metatron.* persisted event names get MIGRATED, not aliased. That decision has a prerequisite this repo does not have — there is NO format_version anywhere in the event log.

Evidence (team review 2026-07-25): at least 13 persisted event types carry the fiction in their names (metatron.charge_regenerated, metatron.entity_removed, metatron.order_triggered, metatron.time_snapped, ...). Renaming them rewrites the replay of every existing world. The only thing standing between a rename and silently-broken replay today is recipes_test.go:75-76, a VALUE PIN that a deliberate retune simply edits. Related hazard, same root: state.go:887-893 re-derives hunt yield from huntYieldSpear (agents.go:733) at apply time.

Scope: introduce an event-log format_version (or equivalent log-level schema stamp), a migration path that rewrites or translates old logs on load, and the doctrine that a persisted-name or reducer-re-derivation change REQUIRES a version bump + migration. The cmd/promptworld 'migrate' command is the natural driver.

Relationship to TASK-75: 75 is the docs/doctrine task and explicitly scopes migration OUT ('migrating them is future work, not this task'). This task is that future work. 75 should land its doctrine note first or alongside; this task supplies the machinery 75 points at.

Sequencing: this BLOCKS TASK-121 (skinnable guardian). 121's rename sweep cannot honestly proceed until the versioning exists — otherwise the sweep is a one-way replay-compat door with no migration behind it. Non-trivial: full Spec Kit before implementation. Expect Opus tier (replay/reducer doctrine, cross-package).

Spec: specs/094-event-log-format-version
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Event log carries a format_version (or equivalent schema stamp) written at genesis and on every append-path that needs it
- [ ] #2 A migration path translates pre-version logs on load; an existing world-01-shaped log replays byte-identically before and after migration
- [ ] #3 Doctrine recorded: persisted-name changes and reducer-re-derivation changes require a version bump + migration; sites commented
- [ ] #4 TASK-121's metatron.* -> guardian.* rename is demonstrated end-to-end through the migration on a seeded world
- [ ] #5 Wiki re-pinned (event-log / sim-state-reducer notes) and freshness gate green
- [ ] #6 Spec phase: Log stamp + enforcement
- [ ] #7 Spec phase: Translating migration
- [ ] #8 Spec phase: The guardian rename
- [ ] #9 Spec phase: Doctrine + grounding
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Operator ruling (2026-07-25, via UI-sweep orchestrator checkpoint): TASK-121 merged AHEAD of this task as the freeze-everything interim — its branch changes ZERO persisted names (all frozen + annotated at definition sites; pre-052 replay byte-identity proven by its compat suite), so the replay door this task guards stays shut. The real metatron.*→guardian.* rename runs THROUGH this task's migration machinery (AC #4); 121's chronicle Type-column display alias is an interim shim to remove when this lands. Cross-ref: specs/052-skinnable-guardian ruling 2 + freeze annotations.

board-sweep-2026-07-29 lane 1: spec 094 landed + linked. OPERATOR RULING (2026-07-29, in-session checkpoint): ship the REAL metatron.*->guardian.* rename + TASK-121 shim removal through this task, not just a demo. Tier: Opus — card-stated (replay/reducer doctrine, cross-package, migration machinery).
<!-- SECTION:NOTES:END -->
