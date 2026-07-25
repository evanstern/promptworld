# Quickstart: validating scenario machinery (spec 054)

## Prerequisites

- `go build ./...`; no LLM required for the machinery itself (visions/orders
  for a real playthrough need llm.json, but the determinism suite runs dry).

## Automated validation

```bash
go test -race ./...
go test -race ./internal/sim -run 'Scenario|Rubric|Incident' -v   # determinism twins, preemption, rubric tables
go test -race ./internal/tui -run 'Exercise|Briefing' -v
node scripts/check-tui-design.mjs --changed
```

## Manual validation

1. `promptworld new /tmp/fn --scenario first-night` — manifest carries
   scenario block + stage-1 + exercise seed; `--scenario bogus` refuses
   listing the catalog.
2. `promptworld start /tmp/fn && promptworld attach /tmp/fn` — exercise tab
   (key `6`) shows the FIRST NIGHT briefing (framing + "incidents ahead are
   forecast"); any key dismisses; detach + re-attach shows it again.
3. Gauges track live: place a watch (stage-1 grant covers it) — the
   order_placed gauge flips met.
4. At the scheduled time the gru emerges at the authored position (map `G`
   glyph); run the world twice from genesis — identical logs.
5. Survive to dawn of day 2 → pass banner; `curriculum.exercise_passed` +
   `curriculum.stage_unlocked` in one batch on the log; unlocks.json gains
   stage-2; the chronicle gains a narrated chapter at the boundary.
6. Let a run die instead → banner `failed (run ended)`; the morgue run
   summary names the exercise outcome; `promptworld status` shows
   scenario_outcome=failed.
7. Ambient world: no exercise tab; `6` inert; behavior unchanged.

## Re-ground checklist (after merge)

- exercise.md → shipped + dock/keymap/stage-defaults re-pins (in-PR).
- `/grounding-wiki:wiki-update` — executor.md, curriculum-ladder.md,
  sim-loop.md, tui-client.md, cli-promptworld.md, morgue.md, chronicle.md.
- player-docs freshness → regenerate (stage-1 page + keys-reference gain
  the exercise tab).
