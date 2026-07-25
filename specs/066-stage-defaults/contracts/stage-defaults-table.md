# Contract: code table ↔ authority page parity (spec 066)

## Parties

- **Producer**: `docs/design/tui/patterns/stage-defaults.md` — the spec 047
  authority page owning every per-surface, per-stage default value.
- **Consumer**: `internal/tui/stagedefaults.go` — the code table the TUI
  resolves from.
- **Enforcer**: the parity sweep test in
  `internal/tui/stagedefaults_test.go` (TestStageDefaultsSweep or equivalent),
  run in the ordinary `go test ./...` suite.

## Contract

1. The sweep test parses the authority page's "Per-surface stage defaults"
   markdown table at test time (path-relative read, `TestCatalogSweep`
   precedent) and asserts cell-for-cell parity with the code table: same
   surface rows, same six columns (Stage 1–4, Pre-ladder, Narrow), same
   normalized cell values.
2. A default value change is therefore a two-file change (page + code table)
   or the build breaks. Neither side may gain a row/column the other lacks.
3. Surfaces absent from the page's table are NOT stage-shaped; the code table
   must not carry rows the page lacks.
4. Cell vocabulary is the page's own (`on`, `badge + overlay-only`,
   `present iff the world carries a scenario`, `forecast`/`fog`, `reachable`,
   fire rules, fold references); the test normalizes formatting (whitespace,
   emphasis) but never meaning.
5. The resolution function's fail-open rule (unrecognized/empty stage →
   pre-ladder union) is asserted by its own unit test; it is part of this
   contract even though it has no table cell.

## Change protocol

Amend the authority page first (spec 047 gate re-verifies + re-pins it in the
same PR), then the code table, in one commit — the sweep test proves they
moved together.
