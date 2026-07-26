---
id: TASK-147
title: >-
  Daemon lifecycle commands deadlock across FormatVersion bumps: stop/status
  gate on version-matched world.Open
status: Done
assignee: []
created_date: '2026-07-26 15:57'
updated_date: '2026-07-26 17:53'
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
- [x] #1 promptworld stop reaches and gracefully stops a running daemon in a world whose format_version is NOT the current one (regression test with a deliberately wrong-version manifest)
- [x] #2 promptworld status reports the daemon for a wrong-version world instead of erroring at Open
- [x] #3 Commands that genuinely read/write world content keep the version gate (migrate hint unchanged); only daemon-lifecycle paths bypass it
- [x] #4 go build/vet/test green; one PR
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implementation complete on task-147-stop-version-gate (aedcf52, Sonnet tier): stop/status version-agnostic via world.SockPathIn/PidPathIn; root cause was deeper than the card's diagnosis — daemon.IsRunning itself called world.Open and returned not-running for any Open failure (fixed, empirically verified); 5 regression tests; audit left attach/timectl/speed/work gated per the card's carve-out. PR blocked by spec 069 wiki-repin-missing (13 notes) — in-branch re-pin pass dispatched.

PR #108 opened (branch task-147-stop-version-gate @ 0391837): fix + 5 regression tests + spec-069 in-branch grounding (13 notes, player docs) + TASK-144 collision resolved by merge; pr gate warnings-only. Awaiting operator merge.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Fixed and merged (PR #108). stop/status are version-agnostic: cmdStop never opens the world (requireWorldDir + pid/socket paths via new world.SockPathIn/PidPathIn); cmdStatus dials the socket first and maps the new world.ErrFormatVersionMismatch to 'daemon not running' on the offline branch. Deeper root cause fixed: daemon.IsRunning itself called world.Open and reported any Open failure as not-running — now a pure pidfile+liveness check. 5 regression tests incl. a live stand-in daemon killed through a wrong-version manifest. Audit: attach/timectl/speed/work + content commands stay gated per the card's carve-out. Spec-069 grounding rode the branch (13 notes re-verified, 3 prose; player docs refreshed; TASK-144 collision resolved by merge). Sonnet tier (single-package surgical fix, trivial exemption).
<!-- SECTION:FINAL_SUMMARY:END -->
