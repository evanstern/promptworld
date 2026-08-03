---
id: TASK-194
title: >-
  Musings flood the chronicle on the qwen default: 599 of 921 retained lines,
  ring overflowing 7 of 13 chapters
status: To Do
assignee: []
created_date: '2026-08-03 05:47'
labels:
  - debt
dependencies: []
ordinal: 176001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Villagers' private musings crowd the story feed the same way harvest corrections used to. On the local model the project now ships as its default, two thirds of everything the narrator reads is villagers musing to themselves, and the buffer that holds a chapter's material overflows and silently throws the rest away.

As a player, I want the chronicle to tell me what happened in my village - who built, who fought, who died - not to be two thirds one villager's inner monologue.

As the guardian reading the story feed, when a chapter overflows, I want to lose the least important material rather than whatever happened earliest in the day.

Found while measuring spec 110 (TASK-173), which fixed exactly this mechanism for map corrections. Evidence, from the spec 110 replay harness over the preserved soak world ~/Claude/soak-evidence/2026-08-02/soak-qwen (qwen3.6, the TASK-184 shipped default):

- Its chapters overflow the 120-line narrLines ring 7 of 13 times BOTH before and after spec 110 - corrections were never its driver.
- 599 of 921 retained lines (65%) are agent.thought source=musing lines.
- The overflow is oldest-first (internal/mind/narrate.go, narrMaxLines = 120), so the evicted material is whatever happened earliest, not whatever mattered least.

Spec 110 established the shape of the fix and the measurement method: classify, then coalesce rather than suppress, and prove it by replaying a preserved world's own log rather than running a fresh soak. Neither the classifier nor the coalescing generalises for free - musings have no equivalent of the harvest ledger's ground truth, so what 'ordinary' means for a musing is the open design question this card carries.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Musings' share of the narrator's per-chapter line buffer is measured before/after on a preserved world of at least 12 game-days
- [ ] #2 Chapters no longer overflow narrMaxLines on account of musings, or the overflow policy stops being oldest-first
- [ ] #3 Distinctive or plot-bearing musings still reach the chronicle - the fix does not silence the inner life wholesale
<!-- AC:END -->
