---
id: TASK-100
title: Story feed cannot render daemon.llm_warning — TestCatalogSweep red on main
status: To Do
assignee: []
created_date: '2026-07-24 19:51'
updated_date: '2026-07-25 03:10'
labels:
  - bug
  - tui
dependencies: []
priority: high
ordinal: 10000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Refiled from a TASK-96 ID collision (two sessions created TASK-96 concurrently 2026-07-24; the mental-maps task keeps the ID — spec 041 references it).

Found 2026-07-24 during TASK-78 baseline validation: `go test ./internal/tui -run TestCatalogSweep` FAILS on main (verified at 404a98f): 'docs/wiki/event-types.md backticks "daemon.llm_warning" but the catalog fixture doesn't cover it' (internal/tui/digest_test.go:208).

Diagnosis (complete): spec 034 introduced the `daemon.llm_warning` event (provider-health preflight; spec 038 made it loud) and the wiki note docs/wiki/event-types.md:94,147,200 documents it, but internal/tui has NO digest registry entry and NO catalogFixture row for it (grep 'llm_warning' internal/tui/ → zero hits). The sweep test cross-checks wiki-backticked event types against the fixture, so it correctly flags the gap: the story feed has no renderer for daemon.llm_warning events.

Surgical fix (trivial-exemption candidate per constitution Development Workflow): add a digestRegistry entry rendering LLMWarningPayload{provider, kind, detail, remedy?, active} (raise vs clear flavors, same operator-facing class as daemon.started/stopped) + matching catalogFixture row in internal/tui/digest_test.go.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 TestCatalogSweep passes on main (go test ./internal/tui)
- [ ] #2 daemon.llm_warning events render a sensible story-feed line for both active=true (raise) and active=false (clear)
- [ ] #3 Catalog fixture row added; no other digest behavior changes
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Dedup 2026-07-24: TASK-92 and TASK-94 were duplicate cards for this same defect (three sessions independently hit the red sweep) — both archived in favor of this card. Folded in from TASK-94: the event is an operator-facing state-no-op with no digest renderer, so a bare fixture row alone would fail the fixture→registry cross-check — the chosen resolution (this card's surgical fix: digestRegistry entry with raise/clear flavors + fixture row) should be recorded here when it lands (TASK-94 AC#2).
<!-- SECTION:NOTES:END -->
