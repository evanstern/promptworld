# Contract: scenario machinery (spec 054)

## §1 Determinism contract

1. Scheduled emissions and rubric passes are executor-class: pure over
   (state, boot-frozen scenario config, next tick); no LLM, no injection
   door, no new RNG purpose tags, no mutable machinery state.
2. Two genesis runs of the same (seed, manifest) produce byte-identical
   scenario event sequences; replaying a recorded log reproduces state
   exactly (the recorded events are the only latches).
3. Ambient worlds (no Scenario block): every code path is dormant;
   behavior byte-identical to pre-054 (regression suite unchanged).

## §2 Event contracts

- `gru.emerged` (scheduled): identical shape/reducer as the random path;
  emitted at the authored tick/position only when preconditions hold (no
  gru abroad); a scheduled night suppresses that night's random roll —
  never two mechanisms in one night; a skipped incident (precondition
  failed) emits nothing and is never retried retroactively.
- `curriculum.exercise_passed{exercise, stage, evidence[]}`: emitted once,
  at the rubric's boundary tick, evidence via the sanctioned constructors.
- `curriculum.stage_unlocked`: same batch as its pass, pass first, only
  when `EvaluateUnlock` grants (existing conjuncts untouched).
- Failure: NO event — `run.ended` with no prior pass IS failed (Status and
  panel derive it; the morgue names it).
- Time-snap interaction: an incident whose tick was snapped past fires at
  the next evaluation (late, once) **within the incident's own night** —
  the never-twice latch is state-derived (no mutable fired-flags, per §1),
  so the due window is `authoredTick ≤ tick < next dawn`; a snap past the
  entire night skips the incident silently (the precondition-failed class:
  its night no longer stands). Replay reproduces whatever was recorded.
  *(Amended 2026-07-25 at implementation, planning-tier approved: unbounded
  fire-late would require a mutable fired-flag the purity doctrine forbids;
  never-twice wins over fire-anytime.)*

## §3 CLI contract

`promptworld new <dir> --scenario <id>`: stamps Scenario{Exercise: id} +
the definition's Stage/Seed/charter preset; unknown id refuses listing the
catalog; earned-stage gate unchanged. `promptworld status` output carries
the exercise id + outcome model-free (D1).

## §4 Exercise tab grammar (panels/exercise.md as authored)

| Input | Precondition | Effect |
|---|---|---|
| `6` | scenario world | select exercise tab; again → solo zoom; again → home |
| any key | briefing showing, exercise tab visible | dismiss briefing (this attach only); the key is consumed |
| — | ambient world | no tab exists; `6` falls through inert |

Panel composition top→bottom: title (`<NAME> · in progress|passed|failed`),
briefing (first render per attach: framing + visibility mode), gauge rows
(term plain-language + met/pending + backing count — live at every stage),
incident line (`incidents (forecast): …` or omitted under fog), pass/fail
banner replacing `in progress`. Narrow: standard solo/narrow pane, no
special rendering. Parity: tab key + dismiss recorded as parity gaps from
birth.

## §5 Narration + morgue

Narrator: one additional `closeChapter` at the exercise pass boundary and
at a scenario world's run end — additive; ambient cadence byte-identical.
Morgue run summary: on a failed scenario run, one line naming the exercise
and its outcome in the no-blame evidence register.

## §6 The director seam (post-v1)

The incident-source interface is the ONLY attachment point a live
state-watching director will use; its contract (pure, state-latched,
reducer-validated) is fixed by §1 and documented at the interface. Nothing
else in this feature may assume "schedule" specifically.
