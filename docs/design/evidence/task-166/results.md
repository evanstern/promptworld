# TASK-166 move-miracle target freshness — live probe

**Status: RESULTS-PENDING** — the implementer (this file's author) prepared the
world/dial/route recipe below per spec 091's US3 and T006; the orchestrator
runs (or supervises) the live probe and fills in the sections below.

## What the probe must demonstrate (spec 091 US3, SC-001/SC-003)

A seeded MEASURE world (never the TASK-14 playtest world) at 8x, guardian
route proxied through 9router, in which:

1. A name-addressed `work_miracle{kind:"move", class:"villager", villager:
   "<Name>", ...}` call is attempted **after** the named villager has walked
   away from whatever coordinates the guardian last surveyed for it (the exact
   race TASK-163's evidence recorded 3/5 residual rejections against).
2. The call **lands** — the door resolves the villager's live position and
   moves it there, instead of refusing with "no living villager at (x,y)".
3. Evidence (the `cog.tool_call` ledger row: verdict, args, tool) is recorded
   here, and TASK-166's card is updated with the outcome — extinct or reduced
   residual "position-freshness races" class, per SC-003.

This is a fresh probe, not a replay of TASK-163's preserved worlds: those were
run on the PRE-fix binary and are the evidence this fix responds to, not a
substrate to rerun the fix against.

## Recipe (TASK-163 pattern, `docs/design/evidence/task-163/results.md`)

Same-seed measure world (seed 1337), stage-4 `--override` (full capability
ceiling — `work_miracle` is stage-gated below stage-3), harsh dials
(`fire_burn_per_wood=3600`, `gru_emerge_per_mille=1000` — TASK-163's values,
chosen to keep the world eventful without an outcome-comparison arm pair;
TASK-166 needs only ONE world, since this probe demonstrates a mechanism
landing, not an outcome delta), 8x, guardian turn route (`metatron`) proxied
through 9router at `localhost:20128`, single-entry (head-only) chain so a
proxy failure skips the turn rather than silently falling back to a local
model and contaminating the sample.

```sh
# 1. create the world (seed 1337, stage-4 override — full grant surface)
promptworld new task-166-probe --at ~/.promptworld/measure/task-166-probe \
  --seed 1337 --stage stage-4 --override

# 2. harsh dials — tuning.json dropped in the world dir (not a CLI flag;
#    spec 048, docs/wiki/world-tuning.md)
cat > ~/.promptworld/measure/task-166-probe/tuning.json <<'EOF'
{ "fire_burn_per_wood": 3600, "gru_emerge_per_mille": 1000 }
EOF

# 3. guardian route -> 9router, head-only chain (hand-edit the world's
#    llm.json — docs/llm-providers.md's providers+routes v2 shape). Add a
#    provider entry for the 9router endpoint and pin metatron to it alone:
#
#      "providers": {
#        "niner": { "transport": "openai_compat",
#                    "endpoint": "http://localhost:20128/v1",
#                    "model": "cc/claude-sonnet-5" }
#        , ... (existing providers untouched)
#      },
#      "routes": {
#        ... (existing routes untouched)
#        "metatron": { "chain": ["niner"], "no_fallback": true }
#      }
#
#    Probe-verify the route is live before starting the run (a proxy that
#    isn't up must fail the boot/first-call loudly, never silently degrade).

# 4. start the world, then set speed to 8x
promptworld start ~/.promptworld/measure/task-166-probe
promptworld speed ~/.promptworld/measure/task-166-probe 8

# 5. calibrate once (docs/wiki/cli-runtime-control.md)
promptworld calibrate ~/.promptworld/measure/task-166-probe
```

Let the world run watch-triggered (autonomous, harsh dials) until at least one
`work_miracle{kind:move,class:villager}` attempt lands; a targeted prompted ask
through `promptworld guardian` chat (naming a villager the operator can see is
mid-walk, e.g. via `promptworld status`/the TUI) is an acceptable supplement
if autonomous attempts run dry within a reasonable window, mirroring TASK-163's
probe-battery supplement.

Ledger query (TASK-163's pattern) to pull the move attempts once the run
window closes:

```sql
SELECT json_extract(payload,'$.snapshot_tick'), json_extract(payload,'$.tool'),
       json_extract(payload,'$.verdict'), json_extract(payload,'$.args'),
       json_extract(payload,'$.reason')
FROM events
WHERE type='cog.tool_call'
  AND json_extract(payload,'$.tier')='niner'
  AND json_extract(payload,'$.tool')='work_miracle'
ORDER BY seq;
```

Stop (never delete) the world when the probe window closes; note its path
under "Raw evidence" below, per this repo's worlds-preserved convention
(`~/.promptworld/measure/`).

## Headline

*(orchestrator fills in after the run — attempts, landed, rejected, and
whether any residual "no living villager at (x,y)" rejection was a
coordinate-only call, per FR-003's expected residual)*

## Raw evidence

*(orchestrator fills in: world path under `~/.promptworld/measure/`, binary
commit the run executed against, the ledger rows the query above returns)*

## Feeds

- **TASK-166 SC-001/SC-003**: pending this probe's outcome.
- **Card update**: TASK-163's residual "position-freshness races" class —
  extinct or reduced to the coordinate-only path this probe should confirm.
