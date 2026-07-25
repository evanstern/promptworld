---
id: TASK-107
title: 'World tuning manifest: tuning.json promotion path for doctrine constants'
status: To Do
assignee: []
created_date: '2026-07-25 02:59'
updated_date: '2026-07-25 03:10'
labels: []
dependencies: []
priority: high
ordinal: 12000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Keystone from docs/design/control-surface-and-calibration.md §6. A boot-loaded, clamp-validated, event-logged tuning.json in the world dir; every field defaults to the current doctrine constant. First promoted dials: refuelDyingBelow, fireBurnPerWood, gruEmergePerMille, PlannerCadenceTicks, conversation pair cooldown. Values logged as events at boot/change so replays reproduce behavior (calibration.json pattern). Follow-on goal (user decision 2026-07-24): once dialed in on world-01, fold the tuned values back as the standard default for new worlds.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 tuning.json read at boot with per-field clamps; absent file == current constants
- [ ] #2 applied values emitted as events so replay is deterministic
- [ ] #3 the five named dials consume the manifest instead of consts
- [ ] #4 docs/design report §6 updated to point at the mechanism
<!-- AC:END -->
