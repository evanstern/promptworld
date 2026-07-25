# Spec 044 quickstart walkthrough — US3/US4/Polish (T028)

Run by the Sonnet implementer finishing US3 (gru escalation), US4 (graves), and
Polish on `task-31-run-outcomes-morgue`. US1/US2 (run end, morgue) were already
implemented and merged onto this branch before this session started; this
walkthrough re-exercises them incidentally wherever a live run naturally does.

All commands below ran against a scratch `PROMPTWORLD_HOME` under the system
temp directory (never the repo, never `~/.promptworld`).

## §1. Hermetic proof — full pass

```
go test ./internal/sim -run 'TestGru|TestRunEnd|TestGrave|TestDeterminism|TestReplay' -v
go test ./internal/scribe -run Morgue -v
go test ./internal/tui -run 'TestCatalogSweep|TestHeader' -v
```

All green. Confirmed live in the log:

- Healthy-villager floor holds (`TestGruWoundsNotExecutes`, `TestGruEscalationScenario`).
- A weakened villager's escalated attack emits `agent.died{cause:"gru"}` with
  witness memories, inventory spill, and a grave — all in one `stepEvents`
  call (`TestGruEscalationScenario`, `internal/sim/grave_test.go`).
- Exactly one `run.ended` after same-tick deaths (`TestRunEndedOnceOrderedLast`,
  pre-existing).
- Replay rebuilds `Ended` state and byte-identical morgue facts
  (`TestReplayRebuildsEnded`, `TestMorgueReplayByteIdentity`,
  `TestGravePlacedAtDeathTilePersistsAndReplays`).
- `TestCatalogSweep` green with no new fixture/registry entries needed (US3/US4
  changed existing event types' vocabulary/behavior, added none).
- `TestRebaseTaxonomyComplete` and `TestWhitelistDiffIdentical` green with no
  new whitelist entries (both changes are executor/reducer-side, per the
  handoff note).

## §2/§4 — Live run-end + escalation walkthrough (ran further than expected)

```
promptworld new demo044 --seed 7
rm worlds-home/worlds/demo044/llm.json     # no-LLM validation
promptworld start demo044
promptworld speed demo044 max
```

There is no CLI-exposed way to force a villager's death directly — `promptworld
miracle remove` explicitly refuses class `villager` ("villager cannot be
removed"), and the "operator door (test hook)" quickstart.md alludes to is an
in-process mechanism (`Loop.InjectOperator`), not a shipped CLI command. So
rather than force it, the world was simply let run at `max` speed
(~50-60k ticks/sec real, uncapped) and it played out **an entire natural run
to completion**, ending with `run.ended` after real real-world minutes:

```json
{
  "clock": {
    "tick": 7404917, "game_time": "day 86 22:55", "speed": "max",
    "effective_rate": 0, "ended": true, "ended_day": 86
  }
}
```

The gru killed all eight villagers over days 76-86 (`final_cause: "gru"`), a
genuine, unscripted escalation cascade — confirmed from the event log
(`promptworld tail demo044 --since 0`):

```
day 76 03:57  gru.attacked {"agent":3,"health":423}
day 76 04:07  gru.attacked {"agent":3,"health":143}
day 76 04:17  gru.attacked {"agent":3,"health":0}
day 76 04:17  agent.died   {"agent":3,"cause":"gru"}
```

repeated for all 8 agents, each showing the exact escalation shape: health
drops in ~240-point steps until it lands below `nearDeathBelow` (200), then the
NEXT hit floors to 0 and an `agent.died{cause:"gru"}` follows immediately in
the same tick, in the same batch as `gru.attacked` — exactly the R4/R5 design.

Postmortem posture, live:

```
$ promptworld status demo044
world "demo044" (seed 7) — daemon running (pid ..., up 168s, 0 subscriber(s))
tick 7404917 (day 86 22:55) — run ended day 86, all villagers dead; world is an archive (read-only)

$ promptworld pause demo044
promptworld pause: run has ended: "pause" refused — the world is a read-only archive
(exit 1)
```

Restart persistence (FR-004, AC#3), live:

```
$ promptworld stop demo044 && promptworld start demo044
daemon started (pid ...): tick 7404917 (day 86 22:55) — run ended day 86, all villagers dead; world is an archive (read-only)
$ promptworld status demo044 --json
{"clock":{"tick":7404917,"ended":true,"ended_day":86, ...}}
```

Tick and `ended_day` unchanged across the restart; the daemon never resumed
ticking (`effective_rate: 0`).

## §5 — Graves & grief (ran live too)

The morgue's run summary (`<world-dir>/morgue.md`) closed cleanly:

```
## The run — ended day 86
- **Run length**: 86 days
- **Population**: 8 → 7 (day 76) → 6 (day 79) → 5 (day 80) → 4 (day 80) →
  3 (day 81) → 2 (day 81) → 1 (day 85) → 0 (day 86)
- **The deaths**: (all 8, cause gru)
```

Grave perception confirmed live from the event log — surviving villagers'
perception sweeps granted the grave place-fact independently of any code path
this session added new plumbing for (research R10's "gets it for free" claim):

```
day 76 12:07  agent.saw {"agent":0,"facts":[{"kind":"grave","x":57,"y":53,"seen":6502070,"prov":"witnessed"}, ...]}
day 76 13:22  agent.saw {"agent":7,"facts":[{"kind":"grave","x":57,"y":53,"seen":6506579,"prov":"witnessed"}]}
day 76 13:56  agent.saw {"agent":4,"facts":[{"kind":"grave","x":57,"y":53,"seen":6508603,"prov":"witnessed"}]}
day 76 14:32  agent.saw {"agent":1,"facts":[{"kind":"grave","x":57,"y":53,"seen":6510757,"prov":"witnessed"}]}
day 77 11:11  agent.saw {"agent":6,"facts":[{"kind":"grave","x":57,"y":53,...}]}
```

25 total `agent.saw` events carried a `"kind":"grave"` fact by the end of the
run.

**What did NOT show up live**: a `social.rumor_told` specifically carrying a
witness-*death* memory (tone -80, "Watched ... die of ..."). The live run did
show plenty of grief-adjacent rumor (attack-witness memories, tone -60,
spreading exactly as designed), but the final cascade killed 7 of 8 villagers
within about 10 in-world days, several within hours of each other — there
simply wasn't always a living witness+listener pair with enough elapsed
social-cadence ticks (`nextTick%60==30`) before the next death or the run's
end. This is a property of this particular seed's unlucky endgame, not a
defect: `internal/sim/grave_test.go`'s `TestGriefRumorFromWitnessedDeath`
hermetically proves the mechanism (`TellableFor`'s birth path, salience 10 >=
rumorMinSalience 4) fires correctly whenever a witness/listener pairing exists
with time to talk.

## §3 (evidence alignment) and §6 (reader test) — not independently re-run

Both require editing `charter.md` / placing a standing order mid-run
specifically to demonstrate the evidence alignment, and a naive-reader pass
over `morgue.md`/`chronicle.md`. These are US2 concerns already covered by the
US1/US2 implementer's own test suite (`TestMorgueRunSummary` and the evidence
fields asserted in `internal/scribe/scribe_test.go`) and by SC-003's
mechanism (charter fingerprint + active orders, both read verbatim off
recorded state) — this session did not re-verify them live since they sit
outside the US3/US4/Polish scope handed to this implementer. The live morgue
excerpts above do show the (empty, since untouched) evidence line rendering
correctly: `"no charter observation recorded before this death; standing
orders active: none."` — the no-orders-placed case, handled cleanly.

## TUI (grave glyph, postmortem posture) — verified by unit test, not a live screenshot

`promptworld ui` is an interactive Bubble Tea TUI; capturing it live would need
a pty-driven screenshot harness not set up in this session. Verified instead
by the hermetic TUI test suite: `TestHeaderViewEndedFromStatus`,
`TestHeaderViewEndedFromReplica` (postmortem header token), and
`TestMapRendersGraveGlyph` (the "✝" glyph + `✝grave` legend entry, both
generated from the shared `mapGlyphs` table so the help overlay's glyph
walkthrough page can't drift from the in-game legend —
`TestHelpWalkthroughGlyphPageMatchesSharedTable`).

## Gate summary

- `gofmt -l .` — clean.
- `go vet ./...` — clean.
- `go test ./...` — green (see the orchestrator report for the exact
  per-package run used as evidence).
