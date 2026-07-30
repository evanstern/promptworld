# TASK-81 / spec 101 — canonization miracle: seeded-world demonstration (SC-001)

**Date**: 2026-07-29/30 · **World**: `~/.promptworld/measure/task-81-demo`
(preserved, running) · **Seed**: 1337 · **Stage**: stage-4 (overridden) ·
**Branch**: `task-81-canonization-miracle`

## Setup (local-only routes, zero paid spend)

- `llm.json`: villager routes (planner/conversation/consolidation/narrator/
  meeting/report_card) on local gemma (`http://mbpro-m1.local:11434/v1`,
  `gemma4:12b-mlx`); `embedding` on local ollama
  (`http://localhost:11434/v1`, `all-minilm:latest`); only `metatron`/
  `metatron_watch` on the 9router proxy (`localhost:20128`,
  `cc/claude-sonnet-5`, CC-subscription passthrough, $0 marginal — the
  task-99/task-80 measurement-run recipe verbatim). No paid API traffic.
- `world.json`: seed 1337, stage-4 (overridden), default dials, fresh
  genesis (the task-99-demo precedent — a NEW measure world, never
  world-01/the operator's playtest world).
- Run: day 1 06:00 → day 29 (natural play at 8x/16x, one operator time-snap
  to day 1 21:00 to reach the night the deferred omen needed, plus — to
  reach the guardian's 2-charge premium bank without an hours-long
  wall-clock wait — a bounded `llm.json`-absent `max`-speed interval
  (deterministic executor-only ticks, no cognition calls, no cost) from
  day 2 06:00 to day 29 05:00, then `llm.json` restored and normal-speed
  play resumed). Zero deaths across the run.

## The loop, driven live through the real daemon + a real LLM guardian

**1 — Implant a place-myth** (belief injection via the guardian's own
`send_omen`, deferred to nightfall by the daytime-omen doctrine since the
ask landed at day-1 09:08): targets `everyone`, text `"In dreams you
glimpse Thornspire: an ancient wall stands near (36,10), where the ground
is generous with forage."` The standing order (`ord-11203-0`) triggered on
the REAL `sim.night_started` event at day 1 22:00 (tick 57600); the omen
landed for all 8 villagers as a direct-perception memory
(`origin:"omen"`, salience 8):

```
#9218 t57600 guardian.order_triggered {"id":"ord-11203-0","matched_type":"sim.night_started",...}
#9262 t57649 guardian.nudged {"form":"omen","targets":[Ash..Sage],"text":"In dreams you glimpse Thornspire..."}
#9263-9270  agent.memory_added × 8  "You witnessed an omen: In dreams you glimpse Thornspire..."
```

Nightly consolidation (gemma) folded the omen memory into recorded
beliefs for at least one villager (Ash): `"The land feels stirred; there
is an old wall near the forge-patch."` (confidence 80, `provenance:
witnessed`, evidence citing the omen event). **Observed limitation,
consistent with TASK-80's own evidence** (`docs/design/evidence/task-80/
results.md`, "none was reconcilable by the matcher... one names a feature
without coordinates"): the local model's consolidation paraphrase dropped
the literal `(36,10)` coordinate in favor of a descriptive
("forge-patch"), which `internal/mind/reconcile.go`'s coordinate-regex
matcher cannot act on — a pre-existing, documented risk of the belief
pipeline, not something spec 101 introduces or could fix (D3 deliberately
adds no belief-authoring machinery). The mechanical confirm/disconfirm
path itself is proven deterministically by `internal/sim/regions_test.go`'s
`TestCanonizedFeatureObservable` and the existing `internal/mind/
reconcile_test.go` suite, independent of any one LLM's phrasing.

**2 — Brief + canonize, through the real guardian console**
(`promptworld guardian task-81-demo "..."`, live tool calls, `tier:niner`):
a single turn walked `explain` → `survey_site(36,10,r=6)` →
`brief_myths(limit=5)` → `explain("workings")` → `canonize_region` — the
model consulted the read-only surfaces before acting, exactly D5's design
intent. The FIRST canonize attempt (feature at the region's exact center)
was refused at the door — decades of in-world activity (day 29) had left
something on that exact tile that failed `buildSite`:

```
#624328 cog.tool_call {"tool":"canonize_region","args":{...,"feature_x":36,"feature_y":10},"verdict":"rejected_gate","reason":"something already stands there, or the ground will not bear it"}
```

— the door refusing exactly as `TestRegionFeatureRequiresBuildSite` pins,
live. A second turn retried at an offset feature site (`37,10`, still
inside the radius-6 region) and it LANDED for real:

```
#624358 t2415974 guardian.region_named {"id":"reg-2415952-0","x":36,"y":10,"radius":6,"name":"Thornspire","feature_kind":"wall_stone","feature_x":37,"feature_y":10,"gratis":false}
#624359          cog.tool_call {"tool":"canonize_region",...,"verdict":"landed","tier":"niner"}
```

Charges moved 3 → 1 (the D4 2-charge premium, verified against the real
bank).

**3 — Arrival confirms — the D1 place-text integration, live**: within
~7 real minutes of natural play after the canonization, Ash's reflex
wander landed her at (32,8) — inside the region's radius (distance ≈4.5 of
6) but away from the placed wall — and the SAME arrival-observation
channel spec 097 shipped (UNMODIFIED by this branch) produced:

```
#625143 agent.moved            {"agent":"Ash","x":32,"y":8}
#625144 agent.memory_added     {"agent":"Ash","text":"Looked around: a forage patch at Thornspire (32,8).","origin":"observed"}
#625145 agent.place_observed   {"agent":"Ash","x":32,"y":8,"radius":2,"kinds":["forage"]}
```

**"at Thornspire (32,8)"** — the villager-coined name the guardian just
made real, surfacing through `describePlace`/`featureDesc` with ZERO new
perception code, exactly as `TestDescribePlaceUsesRegionName` pins at the
unit level. This is the demo's headline shot: the myth is now geography,
narrated in the game's own first-person voice, through the pre-existing
channel.

**4 — The raised feature itself confirms — the D3 proof, live**: 25 real
seconds later (32x, natural reflex movement, no scripting), Ash's next
forage step landed her at (36,9) — within `placeScanRadius` (2) of the
placed `wall_stone` at (37,10) — and the observation channel picked up the
feature exhaustively, unprompted:

```
#625634 t2419895 agent.moved          {"agent":"Ash","x":36,"y":9}
#625635          agent.memory_added   {"agent":"Ash","text":"Looked around: a forage patch, a stone wall at Thornspire (36,9).","origin":"observed"}
#625636          agent.place_observed {"agent":"Ash","x":36,"y":9,"radius":2,"kinds":["forage","wall_stone"]}
```

**"a forage patch, a stone wall at Thornspire (36,9)"** — the guardian's
raised feature reads back through the UNMODIFIED spec-097 channel exactly
like a villager-built one, with zero new perception code (`observedKinds`,
`internal/sim/observe.go`, untouched by this branch). This is the D3 wiring
claim proven live, not just at the unit level: a myth naming a real feature
would confirm the instant a belief citing it and the coordinate survives
consolidation intact (the one link this run couldn't force — see the
documented limitation above).

## Verdict

- **SC-001** (the full loop: myth → brief → canonize → arrival confirms →
  narrated with the coined name): demonstrated live through the real
  daemon, the real (niner/gemma) LLM routes, and the real spec-097
  channel — steps 1-4 above are verbatim log excerpts, not a scripted
  fixture. Both halves of "arrival confirms" landed live: the region's
  name (step 3) and the raised feature (step 4).
- **SC-002** (zero new perception/entity schemes): grep-level and
  live-level both hold — `guardian.region_named`'s door reuses `buildSite`
  (proven live by the first attempt's refusal), and the observation channel
  emitted both the coined name and the raised feature's kind with no code
  path added or changed.
- Belief-confirmation via the mind-side matcher specifically depends on the
  local consolidation model preserving a literal coordinate in its
  paraphrase — an existing, documented risk (task-80's own evidence) outside
  this feature's scope; the mechanism itself is pinned deterministically
  by the unit suites named above.
- World preserved (stopped, never deleted) at
  `~/.promptworld/measure/task-81-demo` for re-analysis.

## Reproduce

```sh
promptworld new ~/.promptworld/measure/task-81-demo --seed 1337 --stage stage-4 --override
# llm.json as above (gemma villager routes, niner metatron routes, embedder)
promptworld start task-81-demo
promptworld guardian task-81-demo "send an omen to everyone naming a coordinate + a real feature word"
# wait for night (or snap to it); wait for/drive nightly consolidation
promptworld guardian task-81-demo "call brief_myths, then canonize_region at that coordinate"
promptworld tail task-81-demo | grep -E 'region_named|place_observed|Thornspire'
```
