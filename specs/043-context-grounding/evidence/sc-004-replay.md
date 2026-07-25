# Evidence: SC-004 thrash-window context replay (T013 / T024b)

**Feature**: specs/043-context-grounding | **Recorded against**: branch
`task-105-context-grounding` (this evidence + the T024/T025 code committed on top
of worktree commit 7374d1baa4e06edf62bc65859c7cabe1ec6147b3)

## What SC-004 asks (quickstart §Thrash-episode context replay)

Reconstruct Sage's decision context inside the documented thrash window
(world-01 ticks 265,411–266,631, per the TASK-101 spike) and verify — by
inspecting assembled text only, no model in the loop — that the context shows:

- the reflex-issued (instinct) `forage` in `self_history`,
- the `forage` ↔ `goto_warmth` alternation visible across several records, and
- warmth rendered with a direction (the US2 trajectory) while the agent walks.

## How it was reconstructed

- Test: `TestSageThrashWindowContextReplay` in
  `internal/daemon/context_replay_test.go` — env-guarded (skips unless
  `PROMPTWORLD_WORLD01_DB` points at a world-01 `world.db`).
- Helper: `replayToTick(seed, map, store, cutoff)` in `internal/daemon/daemon.go`
  — the minimal replay-to-arbitrary-tick capability the US1 slice noted was
  missing (`recoverState` only replays to head). It replays the event log from
  genesis, applies every event at or before the cutoff in seq order, and
  assembles Sage's prompt via the exported `mind.AssembleUserPrompt`.
- **Read-only against the real save**: world-01 is a legacy `format_version 3`
  world (this build is v4). The test copies `world.json` + `world.db` into its
  own temp dir and opens only the copy, and this run additionally pointed the env
  var at a `/tmp/world-01-copy/` copy of the save — the real
  `~/.promptworld/worlds/world-01/` dir was never opened or written.
  `world.Open`'s format gate was bypassed deliberately: a read-only historical
  replay needs only the seed + map dims (the map regenerates deterministically
  from the seed), and the current reducer replays the format-stable
  intent-lifecycle events faithfully. Migration was **not** used: it cuts a fresh
  covering snapshot and archives the old event log, which would discard exactly
  the pre-snapshot intent events this reconstruction depends on.
- **Cutoff tick**: 265,864 (inside the documented window). Sage is agent 7
  (`sim.AgentNames[7] == "Sage"`).
- **Tolerated legacy events**: the replay skipped and tallied `metatron.nudged: 2`
  — two omen events the current reducer's night-gate invariant
  (`an omen may land only at night`) rejects, recorded under older code. They are
  unrelated to the intent history; the test asserts that **zero**
  `agent.intent_set` / `intent_done` / `intent_rejected` / `build_failed` /
  `plan_expired` events were skipped, so the `self_history` reconstruction is
  trustworthy.

Command (run once locally):

```sh
cp -r ~/.promptworld/worlds/world-01/{world.json,world.db} /tmp/world-01-copy/
PROMPTWORLD_WORLD01_DB=/tmp/world-01-copy/world.db \
  go test ./internal/daemon -run 'TestSageThrashWindowContextReplay' -count=1 -v
```

Result: `--- PASS: TestSageThrashWindowContextReplay (0.33s)`. Toolchain
`go1.26.4 darwin/arm64`.

## Assembled decision context — Sage (agent 7) at tick 265,864

```
It is day 4 07:51 (daytime). You are at (3, 20).
Needs (0-100): health 100 and steady, food 80 and falling, rest 88 and falling, warmth 46 and rising, morale 100 and steady.
Recently you:
- goto_warmth — you decided this ("Warmth is dangerously low at 45; nearest fire at (6,11) approximately 5-6 tiles away"); still underway
- forage — instinct drove this; still underway
- goto_warmth — you decided this ("Warmth is at 43; fire nearby at (6,11). Must maintain core temperature before hypothermia clouds judgment."); completed
- forage — instinct drove this; still underway
Carrying: 0 wood, 0 stone, 11 water, 0 planks, 4 refined stone, food (2 raw, 7 cooked, 0 meals).
You know of no fires or shelters yet.
Land to the north is unknown to you.
Nearby: Birch (7 tiles away), Cedar (8 tiles away), Fern (8 tiles away).
People: you resent Birch; you resent Hazel; you like Rowan.
Last conversation, with Cedar: Cedar and Sage argue over whether tangible measurements can truly quantify or contain abstract, intangible burdens.
You have heard: "Talked with Oak and others — Discussed morning weather and hunting stones between friends, with Cedar charming but Oak cryptic about his collection."

You remember:
- day 4 06:00 (5★) Survived a freezing night in the open at the fire (36,40).
- day 3 17:02 (6★) A rainbow pointed toward the forest; the village will chase its meaning while I record what it actually was.
- day 3 16:52 (7★) Talked with Birch — Birch questions a strange rainstorm sign while interpreting magic guides, but Sage dismisses signs too hastily, questioning evidence beyond nature.
- day 3 16:47 (7★) Talked with Birch — Birch and Sage discussed omens in storm winds, with uncertainty about whether rainbow signs point directions or are merely reinforcing prior assumptions.
- day 3 16:36 (7★) Talked with Birch — "Sage argues that the storm and rainbow signs point outward towards a deeper world beyond the village boundaries, while Birch is more convinced storm secrets reveal local hidden truths."
- day 3 16:10 (7★) Talked with Birch — "Sage recalls pastel hue arc pointing to hidden treasure they must uncover together."
- day 3 16:05 (7★) Talked with Birch — "Birch recalls a bold rainbow seen in the village sky last day; now shares patterns suggesting deeper forest mysteries and connections to ancient Glowie studies."
- day 3 09:42 (4★) Talked with Birch — Sage questions whether Birch's curiosity is a pursuit of hidden truths or an escape from stillness, to which Birch responds that some secrets are too compelling to ignore.
- day 3 08:52 (4★) Talked with Rowan — Sage warns Rowan against reckless haste, but Rowan insists on proceeding alone to stop the growing rot.
- day 2 10:21 (10★) You witnessed an omen: At midday, the clouds part. A great arc of color stretches across the sky—red, gold, green, blue, violet—so vivid and still that it seems to hang in place, pointing downward like a hand. It leads toward the forest's edge, toward the boundary where the village's search has not yet reached. The seven feel it: something is being shown to them.

From your journal:
- (#2) Day 3, 14:54. The colors persist in our minds, but I fear what they point toward more than I care to admit. Birch sees secrets in the arcs; Rowan sees safety. I see a map of things we may not be ready to find. I must ensure the village remains stable while these visions pull us toward the edge. We …

What do you do next?
```

## IntentLog ring at tick 265,864 (newest last)

```
forage       [reflex]  @264951 -> done
chop         [plan]    @265150
seek         [planner] @265271 -> done
craft_planks [plan]    @265276 -> expired
forage       [reflex]  @265411
goto_warmth  [planner] @265552 -> done
forage       [reflex]  @265731
goto_warmth  [planner] @265864
```

## SC-004 shape confirmed

- **Instinct-issued forage present**: every `forage` record carries source
  `reflex`, rendered "instinct drove this" — the reflex fallback mind, no model.
  (Assertion: prompt contains "instinct drove this".)
- **forage ↔ goto_warmth alternation across ≥3 records**: the ring's tail is the
  textbook thrash — `forage[reflex]@265411 → goto_warmth[planner]@265552 →
  forage[reflex]@265731 → goto_warmth[planner]@265864`. The instinct keeps
  redirecting to `forage`; the planner keeps overriding to `goto_warmth` to fend
  off the cold. Both goals are visible in the `self_history` block.
  (Assertions: block contains both "forage" and "goto_warmth".)
- **Warmth trajectory rendered while walking**: `warmth 46 ... and rising` — the
  US2 trajectory arrow the deciding mind now sees (Sage is walking toward the fire
  at (6,11), warmth climbing off its window-edge anchor). Food and rest read
  `falling`; the deadband holds health/morale `steady`.

This is exactly the redirection the TASK-101 spike documented: the reflex, blind
to the plan, re-issues `forage`; the planner — now able to *see* its own recent
`goto_warmth` decisions in `self_history` and warmth rising — can knowingly
continue toward warmth instead of restarting the thrash. The self-history block
is what makes that visible to the model (US1, FR-003).
