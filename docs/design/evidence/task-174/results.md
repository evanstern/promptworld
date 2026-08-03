# TASK-174 / spec 103 — conversation-outcome JSON robustness: soak evidence (FR-005, SC-001)

**Status: COMPLETE. SC-001 demonstrated — 0 outcome parse failures and 0 abandoned
scenes over 92 founded scenes / 9.37 game-days.**

Two full soaks were run (2026-08-01 → 2026-08-03). Together they do more than measure the
fix: they isolate *why* the first soak looked like the fix had not worked, and the answer
turned out to be a provider defect rather than anything in spec 103's code.

Both used the same binary, seed (1337), stage (stage-4 `--override`), speed (16x), and a
throwaway world outside `~/.promptworld/` — never the playtest world or any preserved
measurement world. Every route pointed at a single zero-priced local provider, so no
paid-spend surface existed in either world's config at any point. **Only the local model
differed.** That is the whole experiment.

## Result

| | playtest-1 (baseline) | soak A — `gemma4:12b-mlx` | soak B — `qwen3.6:latest` |
|---|---|---|---|
| horizon | 29 game-days | 11.97 game-days | 9.37 game-days |
| founded scenes | 293 | 90 | **92** |
| outcome parse failures | 22 | 10 | **0** |
| scenes killed by an outcome parse failure | not separated | 3 | **0** |
| abandoned scenes (all causes) | 62 (21.2%) | 14 (15.6%) | **0 (0.0%)** |

Soak B's complete set of non-`landed` conversation outcomes, across 92 scenes:

| outcome | reason | count | note |
|---|---|---|---|
| `landed` | — | 90 | |
| `suppressed` | `nothing new since last exchange` | 2 | the novelty gate working as designed — not a failure |
| `retried` | `utterance turn 3: bad say JSON: unexpected end of JSON input` | 1 | recovered; the scene completed |

So the run contains exactly **one** imperfect event in 92 scenes, it was non-terminal, and
it was not on the route this spec fixed.

## AC #2 / SC-001 verdict

The board card's AC #2 asks for the *abandoned-outcome* rate — operator ruling 2026-08-01,
binding: **scenes abandoned because the outcome could not be parsed**, not all-cause
abandonment.

- Soak A: **3** scenes (3.3%).
- Soak B: **0** scenes (0.0%), on the model the project now ships by default.

All-cause abandonment, recorded for completeness rather than as the criterion: 21.2%
baseline → 15.6% (A) → **0.0%** (B).

## Why soak A still failed — the finding that mattered

Spec 103 restored structured outputs: the caller sets a JSON Schema and
`internal/llm/providers.go` sends it as an OpenAI-compat
`response_format: {type: json_schema}` envelope. Soak A ran that code and still produced
10 outcome parse failures.

Direct probing of the rig (2026-08-02) found the cause outside promptworld entirely:
**Ollama's MLX engine silently discards schema constraints.** On `gemma4:12b-mlx`, all
three constraint mechanisms returned free prose — OpenAI-compat `json_schema` strict, the
same non-strict, and Ollama's own native `/api/chat` `format` parameter. Its own GGUF
sibling `gemma4:latest` honored all three. The split tracks `details.format`
(`safetensors` = MLX = ignores; `gguf` = honors), not the model family or size.

So spec 103's code was correct throughout, and the constraint never reached the sampler.
Every downstream symptom — parse failures, retries, abandoned scenes — looked like a
promptworld defect while the provider returned HTTP 200 and a normal-looking completion.
Nothing in the daemon log, `status`, or `calibrate` reported it; `calibrate` could not
complete against that model at all.

This produced **TASK-184 / spec 109** (default moved to a GGUF model, hazard documented,
merged in PR #155) and **TASK-185** (a daemon-start capability probe, so the next
occurrence costs seconds rather than twelve game-days). Soak B is the re-run on the new
default.

## A correction, recorded deliberately

An interim reading of soak A at **23** founded scenes reported AC #2 as **0** and was
**wrong**. At 90 scenes the true figure was 3. The zero was a small-sample artifact, and it
was reported to the operator before the sample was adequate.

The lesson is retained here because it nearly closed this task on a false negative: a count
of zero over a small sample is not evidence of absence, and the ≥20-founded-scenes bar in
this spec is a floor for having *any* signal, not a threshold for a defensible rate. Soak
B's zero is credible because it is a zero over 92 scenes **and** because it was predicted in
advance by an independent mechanism test, not discovered by looking.

## Residual: the utterance route (TASK-183)

The one retry in soak B is on the say/utterance route, and its character changed between the
runs — worth recording, because it re-scopes TASK-183:

- Soak A: `no JSON object in reply`, **no raw payload** — the model emitted prose. It killed
  11 of soak A's 14 abandoned scenes.
- Soak B: `unexpected end of JSON input`, **raw payload present** — the model emitted
  well-formed JSON that was **truncated**. Recovered on retry; zero scenes lost.

Under a constraint-honoring model the utterance route's failure stops being "the reply isn't
JSON" and becomes "the reply ran out of tokens," which is a materially smaller problem with a
different fix.

## Free corroboration for TASK-175

Both worlds independently reproduce spec 106's result: **0** `"is asleep"`
`agent.intent_rejected` in each, against a 31.2/game-day baseline — soak B across 554
`agent.slept` edges. Recorded in `specs/106-sleep-gated-planning/soak.md`.

## Reproducing

`docs/design/evidence/task-174/queries.sql` runs unchanged against either world's `world.db`
and reproduces the D6 metrics exactly. Both worlds are preserved under the session
scratchpad.

**Query note:** these `world.db` files are WAL-mode. Once a daemon has stopped cleanly and
no `-wal` file remains, `sqlite3 -readonly` fails with `unable to open database file (14)`;
open without `-readonly`. Reading a *running* world with `-readonly` works fine.
