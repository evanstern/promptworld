# Quickstart: validating stage-shaped layout defaults (spec 066)

## Prerequisites

- Repo built: `go build ./...`
- A daemonized world you can set stages on (or the test corpus below — the
  automated path needs no live world).

## Automated validation (the gate path)

```sh
# Parity: code table == authority page, cell for cell
go test -race -run TestStageDefaults ./internal/tui/

# Pre-ladder byte-identity + per-stage starting-set frames + fold composition
go test -race ./internal/tui/

# Everything (includes reachability sweep and first-occurrence exactly-once)
go test -race ./...

# Design gate (page specified → shipped, pins fresh)
node scripts/check-tui-design.mjs --changed
```

Expected: all green; the pre-ladder golden-frame test proves SC-002's
byte-identity; existing `layout_test.go` fold tests pass unmodified (SC-004).

## Manual validation (live TUI)

1. **Stage-1 boot (US1)**: create/attach a stage-1 world
   (`promptworld new --stage stage-1 …` per spec 046's flow), attach the TUI.
   Expect the authority table's Stage 1 column: map + narrated chronicle,
   guardian strip, lesson row on; incident vocabulary `forecast` if the world
   carries a scenario. Open the help overlay (`?`) and reach a non-default
   surface (e.g. a solo view) — full stage-independent content.
2. **Pre-ladder (US2)**: attach a pre-stage world — the layout must be
   indistinguishable from the pre-feature TUI (everything on).
3. **Live unlock (US3)**: drive a stage unlock (curriculum ladder flow);
   expect the ceremony, then re-resolved defaults for the new stage (e.g.
   `forecast → fog` crossing into stage 3; lesson row folding to badge), with
   any surface you explicitly toggled this session left exactly as you set it.

## References

- Authority values: `docs/design/tui/patterns/stage-defaults.md` (the table;
  see contracts/stage-defaults-table.md for the parity contract)
- Fold composition: `docs/design/tui/patterns/layout.md` ruling (a)
- Entities/state: data-model.md
