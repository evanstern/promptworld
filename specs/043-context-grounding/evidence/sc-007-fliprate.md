# Evidence: SC-007 flip-rate measurement (T027)

**Feature**: specs/043-context-grounding | **Recorded against**: `main` @ `72125c8`
(merged task-105-context-grounding)

## What SC-007 asks (quickstart §Behavioral flip-rate — measured, not gated per-commit)

Compare the rate of same-agent food↔warmth intent alternation against the
frozen world-01 baseline from the TASK-101 spike, using the same
flip-counting method: forage↔goto_warmth-style A→B transitions where A and B
fall in different goal classes. Target: worst-agent flip rate drops by
**≥50%** relative to baseline (worst agent ≤ 36 flips/game-day), understanding
that the spec's assumption frames the 50% target as this feature *combined
with* whatever of TASK-103/TASK-104 (instinct arbitration, recovery intents)
has landed — **neither has landed yet**, so this measures context-grounding
(spec 043) alone.

## Method

Counting script: `/tmp/sc007/analysis/flip_count.py` (ad hoc, not committed —
reads `world.db`'s `events` table directly, not the repo). For each agent, in
tick order over `agent.intent_set` events:

- classify each `goal` into **food** `{forage, hunt}`, **warmth**
  `{goto_warmth, build_fire, refuel_fire}`, or ignored (all other goals);
- a **flip** is a transition between two consecutive *classified* records for
  that agent where the class changes (food→warmth or warmth→food);
- report total flips, flips within ≤200 ticks of the prior classified
  intent_set, and flips normalized to flips/game-day (ticks-per-day = 86400,
  confirmed from `internal/sim/governance.go`'s `DayIndex`/`gruNightIndex`
  and consistent with the baseline's own 523,670 ticks ≈ 6.06 days).

This is the exact class split and adjacency rule TASK-101's spike used
(`agent.intent_set` payload fields `agent`/`goal`/`source`, confirmed against
`internal/sim/agents.go`'s `IntentSetPayload` and `internal/sim/policy.go`'s
goal-string vocabulary before counting).

**Baseline** (TASK-101 spike, world-01, 523,670 ticks ≈ 6.06 game-days):

| Agent | Total flips | ≤200-tick flips | Flips/game-day |
|---|---|---|---|
| Sage (worst) | 436 | 334 | ≈72 |
| Fern | 350 | — | ≈58 |
| Oak | 294 | — | ≈49 |
| Rowan | 204 | — | ≈34 |
| Birch | 202 | — | ≈33 |
| Cedar | 153 | — | ≈25 |
| Hazel | 57 | — | ≈9 |
| Ash | 13 | — | ≈2 |

**SC-007 bar**: worst agent ≤ 36 flips/game-day (50% of Sage's ≈72).

## Deviation from the plan — two samples instead of one

The original plan was a single throwaway world at 32x for 2-4 game-days. That
run (`sc007-flip`) was stopped partway through by the user (~1 game-day in)
before reaching the planned floor. Rather than throw the partial data away
and re-run from scratch, the orchestrator directed analyzing it as **Sample
A** and adding a second, independently-running post-043 world the user
already had live (**Sample B**, world `t2`) as a naturally-occurring
higher-fidelity comparison at a different speed/model posture. `t2` was
**never stopped, paused, or written to** — its `world.db` (+`-wal`/`-shm`)
was copied to a scratch path and only the copy was queried; its own daemon
(pid confirmed still running post-analysis) was left completely alone.

## Sample A — `sc007-flip`, fresh throwaway world, cogito:3b @ 32x

Built from a clean `git archive origin/main` export (repo root untouched,
left on `main`). World created path-form (`promptworld new
/tmp/sc007/worlds/sc007-flip`) so it never touched the real worlds registry
— not `world-01/02/03/myworld-01/test/t1`. `llm.json` left at the fresh-world
default (`llm.DefaultConfig()`): local provider `cogito:3b`
(`openai_compat`, `http://localhost:11434/v1`, `tool_mode: "json"`), serving
the `planner` route — the same tier `world-01` ran. Confirmed via `GET
localhost:11434/api/tags` that `cogito:3b` was pulled and served
(`capabilities: ["completion","tools"]`) before starting. Calibrated
(`promptworld calibrate --tier local`) and the daemon restarted once so
planner/conversation thoughts weren't suppressed at 32x (calibration seeds
only at daemon boot); this restart happened at tick ~1135, negligible against
an 85k-tick run.

Run span: tick 0 → 85,307 (continuous; one restart for calibration pickup,
no gap in the intent-lifecycle event stream) = **85,307 ticks ≈ 0.987
game-days**. Stopped by the user before reaching the planned 2-4 day floor.

| Agent | Total flips | ≤200-tick flips | Flips/game-day | n classified intents |
|---|---|---|---|---|
| Ash (worst) | 5 | 0 | 5.06 | 26 |
| Birch | 3 | 0 | 3.04 | 25 |
| Rowan | 3 | 0 | 3.04 | 26 |
| Cedar | 2 | 0 | 2.03 | 21 |
| Fern | 2 | 0 | 2.03 | 27 |
| Hazel | 2 | 0 | 2.03 | 19 |
| Oak | 2 | 0 | 2.03 | 38 |
| Sage | 2 | 0 | 2.03 | 42 |

**Worst agent: Ash, 5 flips total (0 within ≤200 ticks), 5.06 flips/game-day.**
Manual inspection of Ash's food↔warmth transitions shows they're all >700
ticks apart (spot-checked: deltas of 740–12,260 ticks) — no thrash pattern
visible at all, consistent with zero ≤200-tick flips.

vs. baseline worst (Sage, ≈72/day): **93.0% reduction** — comfortably clears
the ≥50% target and the ≤36/day bar.

## Sample B — `t2`, player-inhabited world, gemma4:12b-mlx planner @ 4x

Confirmed post-043 (every `cog.thought` event since world creation carries
`block_bytes`, including `self_history`/`known_places` — spec-043 blocks;
zero pre-043 planner thoughts in the log) and a single continuous run (one
`daemon.started`, no restarts). **Important difference from Sample A and the
world-01 baseline**: `t2`'s `llm.json` routes `planner` to a provider named
`gemma` (`gemma4:12b-mlx`, a larger/different model than `cogito:3b`, served
from a separate host), not `cogito:3b` — so Sample B is **not** an
apples-to-apples tier match; it's a different (larger) planner model running
at a slower, non-accelerated 4x instead of 32x. The coordinator's framing —
"the fairer regime for judging decision quality" because cognition keeps up
better at 4x — is a real trade against the tier mismatch, not a wash; both
cut in different directions and neither substitutes for a same-tier,
full-length run.

Run span (copy taken at the analysis point, world untouched thereafter): tick
0 → 73,132 = **73,132 ticks ≈ 0.846 game-days**.

| Agent | Total flips | ≤200-tick flips | Flips/game-day | n classified intents |
|---|---|---|---|---|
| Ash (worst) | 28 | 18 | 33.08 | 75 |
| Sage | 27 | 10 | 31.90 | 88 |
| Oak | 15 | 5 | 17.72 | 57 |
| Cedar | 13 | 6 | 15.36 | 66 |
| Rowan | 9 | 3 | 10.63 | 64 |
| Fern | 9 | 1 | 10.63 | 41 |
| Birch | 3 | 0 | 3.54 | 133 |
| Hazel | 1 | 0 | 1.18 | 28 |

**Worst agent: Ash, 28 flips total (18 within ≤200 ticks), 33.08
flips/game-day** (Sage close behind at 31.90/day). Spot-checked Ash's
transition log directly: a genuine thrash cluster from tick ~48,000–57,900 —
reflex-issued `forage` alternating with planner/plan-issued `goto_warmth` /
`refuel_fire` every ~130–250 ticks, the same reflex-vs-planner
counter-scheduling pattern TASK-101 documented in world-01 — this feature
visibly reduces but does not eliminate the pattern under this posture.

vs. baseline worst (Sage, ≈72/day): **54.1% reduction** — clears the ≥50%
target and sits under the ≤36/day bar, but by a narrow **8% margin** (33.08
vs. 36.0), not the comfortable margin Sample A shows.

## Verdict

Both samples clear the SC-007 bar (worst agent ≤ 36 flips/game-day) as
measured:

| Sample | Span (game-days) | Worst agent | Flips/day | vs. baseline (≈72/day) | Bar (≤36) |
|---|---|---|---|---|---|
| A (`sc007-flip`, cogito:3b, 32x) | 0.987 | Ash | 5.06 | −93.0% | met, wide margin |
| B (`t2`, gemma4:12b-mlx, 4x) | 0.846 | Ash | 33.08 | −54.1% | met, narrow margin |

**Honest caveats — report the number without spin either way:**

- **Neither sample reaches the planned run length.** The quickstart/spec ask
  a multi-day (2-4 game-day) run; both samples here are under one game-day
  (0.99 and 0.85 days respectively), against a baseline measured over 6.06
  game-days. Short spans mean small per-agent flip counts (as few as 1-5),
  so the per-game-day normalization amplifies noise — a single extra or
  missing flip swings the rate by ~1/day in Sample A and more in Sample B.
  Neither sample should be read as a precise rate; both are directional.
- **Sample B is not a controlled comparison**: different planner model
  (`gemma4:12b-mlx` vs. `cogito:3b`), different speed (4x vs. 32x/world-01's
  presumed comparable acceleration), and a player-inhabited world with
  unknown session history/interventions rather than a scratch world run
  end-to-end for this measurement alone. Its result (33.08/day, an 8% margin
  under the bar) is the weaker and more cautionary of the two data points —
  it's the one that should carry more weight if a single verdict is wanted,
  since it shows real, visible thrash surviving in the log even after this
  feature landed.
- **TASK-103 (instinct-yields-to-intelligence arbitration) and TASK-104
  (needs-conditioned recovery intents) have not landed.** SC-007's own
  Assumptions section frames the ≥50% target as being for spec 043 *combined
  with* whatever of that work has landed — this measures spec 043 alone, so
  a future combined measurement is expected to (not guaranteed to) do
  better, not worse, once the reflex/planner layer-fight direction (A) and
  needs-conditioned recovery (B) also land.
- Sample A's `llm.json` also carries default `cloud` routes
  (`consolidation`/`narrator`/`drama`/`metatron` → `claude-opus-4-8`) with no
  `ANTHROPIC_API_KEY` set in this environment; those routes ran degraded
  (no key) but do not touch the `planner` route this measurement counts.

**Bottom line**: both available samples clear the ≥50%-reduction /
≤36-flips-per-game-day bar as measured, but on short runs (under a
game-day each) against a 6-day baseline, and one sample (`t2`) clears it by
only 8% — not the wide, confident margin a full-length same-tier run would
be expected to produce. This is evidence *for* SC-007, not proof of it; a
longer same-tier run (the originally planned 2-4 game-days at `cogito:3b`)
would tighten the estimate materially, especially given how close Sample B
sits to the bar.

## Cleanup

- `sc007-flip`'s daemon was already stopped (confirmed: no matching process)
  before this analysis; no daemon was left running by this task.
- `/tmp/sc007` (throwaway world, ad hoc binary, analysis script, `t2` copy)
  was removed in full after this evidence was recorded.
- `t2` and its live daemon (pid confirmed running, untouched) were never
  stopped, paused, or written to — only its `world.db`/`-wal`/`-shm` were
  copied to a scratch path and the copy was queried.
- No changes were made to `world-01`, `world-02`, `world-03`, `myworld-01`,
  `test`, or `t1`.
