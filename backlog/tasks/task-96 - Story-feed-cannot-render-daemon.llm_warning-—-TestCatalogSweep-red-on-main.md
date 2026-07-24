---
id: TASK-96
title: Story feed cannot render daemon.llm_warning — TestCatalogSweep red on main
status: To Do
assignee: []
created_date: '2026-07-24 19:37'
labels:
  - bug
  - tui
dependencies: []
priority: high
ordinal: 80000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
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
