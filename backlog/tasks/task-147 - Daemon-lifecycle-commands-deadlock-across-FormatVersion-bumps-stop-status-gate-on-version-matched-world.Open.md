---
id: TASK-147
title: >-
  Daemon lifecycle commands deadlock across FormatVersion bumps: stop/status
  gate on version-matched world.Open
status: In Progress
assignee: []
created_date: '2026-07-26 15:57'
updated_date: '2026-07-26 15:57'
labels:
  - bug
  - cli
dependencies: []
priority: high
ordinal: 117000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Shipped by PR #105 (spec 068, FormatVersion 4→5). Live incident 2026-07-26: a v4 world's running daemon cannot be stopped by the new binary (stop → world.Open rejects v4 with the migrate hint) while migrate refuses the locked running world — a deadlock any future FormatVersion bump recreates. Complete diagnosis: cmdStop (cmd/promptworld/commands.go:590) and cmdStatus (commands.go:623) call world.Open(dir) before touching the daemon, but only need SockPath/PidPath — which are pure path joins on the dir (internal/world/world.go:447-448, daemon.sock / daemon.pid); daemon.IsRunning(dir) already takes the bare dir. Fix: make daemon lifecycle commands version-agnostic — derive socket/pid paths from the dir without the version-gated Open (audit other lifecycle commands for the same pattern; world-content commands keep the gate). Operator workaround used: kill -TERM $(cat <world>/daemon.pid). Trivial exemption (constitution): surgical, file:line diagnosis pinned here, ACs below.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 promptworld stop reaches and gracefully stops a running daemon in a world whose format_version is NOT the current one (regression test with a deliberately wrong-version manifest)
- [ ] #2 promptworld status reports the daemon for a wrong-version world instead of erroring at Open
- [ ] #3 Commands that genuinely read/write world content keep the version gate (migrate hint unchanged); only daemon-lifecycle paths bypass it
- [ ] #4 go build/vet/test green; one PR
<!-- AC:END -->
