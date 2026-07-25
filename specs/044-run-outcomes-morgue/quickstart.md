# Quickstart: validating spec 044 end-to-end

Prerequisites: `go build ./... && go test ./...` green at the feature branch head.
References: [contracts/events.md](./contracts/events.md),
[contracts/morgue-document.md](./contracts/morgue-document.md),
[contracts/status.md](./contracts/status.md), [data-model.md](./data-model.md).

## 1. Hermetic proof (no daemon)

```bash
go test ./internal/sim -run 'TestGru|TestRunEnd|TestGrave|TestDeterminism|TestReplay' -v
go test ./internal/scribe -run Morgue -v
go test ./internal/tui -run 'TestCatalogSweep|TestHeader' -v
```

Expected: healthy-villager floor holds; weakened-villager gru kill emits `agent.died`
cause `"gru"` + witness memories; exactly one `run.ended` after same-tick deaths; replay
rebuilds `Ended` state and byte-identical morgue facts; digest catalog complete.

## 2. Live run-end walkthrough (US1 + US2)

```bash
promptworld new demo-044 --seed 7 && promptworld start demo-044
# ... let it run; force deaths via the operator door (test hook) or a seeded
# no-fire/no-food scenario; watch:
promptworld attach demo-044        # deaths → final death → run.ended pushed live;
                                   # header flips to ENDED without reconnect
promptworld status demo-044 --json # .clock.ended == true, ended day present
cat <world-dir>/morgue.md          # epitaph per death (7 factual fields), run summary
promptworld stop demo-044 && promptworld start demo-044
promptworld status demo-044 --json # still ended after restart (replay)
promptworld pause demo-044         # refused: "run has ended"
```

No-LLM validation: run the world with no `llm.json` — morgue.md must be complete
(facts only, no epilogue blocks). Then on an LLM world, verify a blockquoted epilogue
appears after a later death and no factual byte changed.

## 3. Evidence alignment (US2, SC-003)

Edit `charter.md` mid-run, place a standing order, then cause a death. The epitaph's
"angel's watch" section must name the new charter fingerprint (player-authored) and the
active order's condition/action — readable from `morgue.md` alone.

## 4. Escalation demonstration (US3, SC-005)

Seeded world, two villagers meet the gru at night in the open: pre-attack health 1000 →
survives at 750 (wounded); pre-attack health 150 (< 200 near-death band) → dies, cause
`"gru"`. Replay the world: identical outcomes.

## 5. Graves & grief (US4, SC-006)

After a witnessed death: grave glyph on the TUI map + legend entry; witness's mental map
holds the grave place-fact (provenance witnessed) within one movement beat; within one
game-day a grief rumor/conversation referencing the death appears in the social feed.

## 6. The reader test (SC-007)

Hand `morgue.md` + `chronicle.md` from a lost run to a reader with no other context:
they must be able to answer "what killed the village, and what were my angel's
instructions at each death."
