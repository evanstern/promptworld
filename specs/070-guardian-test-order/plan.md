# Implementation Plan: Guardian worker shutdown joins

**Branch**: `070-guardian-test-order` (task branch: `task-144-guardian-close-join`) | **Date**: 2026-07-26 | **Spec**: [spec.md](spec.md)

## Summary

One-file surgical fix: `sync.WaitGroup` joins the four worker goroutines
`guardian.New` spawns; `Close` = `close(done)` + `wg.Wait()`. Proof by
repetition (`-count=50` isolation, `-count=10 -race` package) — no test-file
changes required.

## Technical Context

**Language**: Go. **Testing**: `go test` count/race matrices. **Scope**:
`internal/guardian/guardian.go` (+ `docs/wiki/guardian.md` re-pin if the pr
gate demands — spec SC-004). **Constraints**: no API shape change; no
timeout on the join.

## Constitution Check

- **I–III** — PASS (diagnosis + decision recorded on TASK-144 and this spec;
  one task, one PR; gates run at choke points).
- **IV** — PASS (planned): `internal/guardian/guardian.go` is a pinned source
  of `docs/wiki/guardian.md`; the re-verification rides THIS branch (the
  in-PR doctrine, whether or not spec 069's gate has merged when this PR
  opens). Player-docs: no page cites guardian.md → freshness check expected
  clean; verify in-branch.
- **V** — PASS: dispatched to spec-implementer on **Sonnet** (operator
  checkpoint resolved 2026-07-26: analysis was the concurrency-risky half and
  is complete; the fix is prescribed). One-way escalation unused.

**Post-Phase-1 re-check**: PASS.

## Design

- **D1**: `wg sync.WaitGroup` field on `Guardian`; `wg.Add(1)` immediately
  before each of the four `go` statements in `New` (`guardian.go:259-262`);
  `defer mt.wg.Done()` as the first line of each worker func.
- **D2**: `Close` = `close(mt.done); mt.wg.Wait()`. Preserve any existing
  double-close guard; add none if absent.
- **D3**: Proof commands (also the board AC): 
  `go test ./internal/guardian/ -run 'TestReportCard' -count=50`;
  `go test ./internal/guardian/ -count=10 -race`; `go test ./...`.
- **D4**: In-branch grounding: re-verify `docs/wiki/guardian.md` against the
  Close-semantics change (its body may state shutdown behavior) and re-pin to
  a branch commit; run the player-docs freshness check in-branch (expected
  no-op).

## Complexity Tracking

Empty — no violations.
