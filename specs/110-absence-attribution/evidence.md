# Evidence: spec 110 (TASK-173) — Phase 4

**Recorded**: 2026-08-02 · **Branch**: `task-173-absence-attribution` · **Spec**: `spec.md`
· **Runbook**: `docs/design/task-173-absence-attribution-runbook.md`

This file is the durable record of the runbook's four required measurements and of every
success criterion Phase 4 can decide. It states what was **measured** and what was
**inferred**, separately, and it records the criteria this route cannot decide rather than
substituting a proxy for them.

## Route (runbook amendment, operator-decided)

Not a fresh live soak: a **replay** of the preserved soak worlds' own event logs through
the Mind's absorb path, so the before/after comparison runs on byte-identical input, plus
a **re-narration** of the replayed chapters against the live models.

Preserved worlds (read-only; every run below used a `sqlite3 .backup` copy under `/tmp`,
never the originals):

| | Run A (primary) | Run B (secondary) |
|---|---|---|
| path | `~/Claude/soak-evidence/2026-08-02/soak-world/world.db` | `~/Claude/soak-evidence/2026-08-02/soak-qwen/world.db` |
| model | `gemma4:12b-mlx` | `qwen3.6:latest` |
| seed | 1337 | 1337 |
| events | 132,299 | 114,635 |
| span | 12.02 game-days (24 closed chapters) | ~6.68 game-days (13 closed chapters), **mid-flight snapshot — a floor** |

## The harness (T015)

`internal/mind/replay_evidence_test.go` — env-gated, skipped by the ordinary
`go test ./...`.

```
PW_REPLAY_DB=/tmp/runA.db PW_REPLAY_OUT=/tmp/runA \
  go test ./internal/mind/ -run TestReplayEvidence -v
```

It lives as a test, not a `cmd/` tool, because the measurement **is** `md.narrLines` — an
unexported field reached only through the absorb path. A `cmd/` binary would have forced
exporting the chapter buffer purely to observe it.

What it does:

1. Builds the world's map and a genesis `sim.State` from the manifest (`seed 1337`,
   64×64, `terrain_gen 2`) and drives **every event, in `seq` order, through
   `md.absorb`**, capturing each `narrJob` the narrator worker would have received.
2. Runs the log **twice**. The `after` pass is the branch as it stands. The `before` pass
   is the same production code with the harvest ledger wiped after every absorbed event,
   so every correction takes the FR-004 unexplained branch — which spec 110 keeps
   byte-identical to pre-110. The before side is therefore a reconstruction of today's
   `main` produced *by the same code*, not a second implementation of it.
3. Builds an **independent** ground truth for SC-004 from the raw log (every
   `agent.chopped`/`agent.quarried` tick at every coordinate) and judges the production
   classifier against it.

**Validation of the before pass**: its per-chapter correction-line counts reproduce
`spec.md`'s measured table **exactly** — 89, 92, 68, 99, 51, 53, 60, 92, 56, 24, 23, 44
for the twelve day chapters — and it reproduces "5 of 12 day chapters over the 120-line
ring". Its *total*-narratable-line column runs a few lines either side of the spec's
(137 vs 141, 155 vs 158, 109 vs 107, …): the spec's totals were an SQL estimate of which
events yield lines, the replay's are the lines the code actually appends. The corrections
column, which is what SC-002 is about, matches to the unit.

## The four runbook measurements

### (c) Harvest-explained share of corrections — MEASURED

Run A, 833 `agent.map_corrected` events over 12.02 game-days:

| | count | share |
|---|---|---|
| classified **attributed** by the production ledger | 830 | 99.6% |
| classified **unexplained** | 3 | 0.4% |
| false attributions (attributed with no harvest at those coordinates in the window) | **0** | **0%** |

Independently reproduced by SQL against the same DB, joining each correction's first
`Gone` fact to the most recent harvest at those exact coordinates: 716 same game-day, 76
at 1 day, 35 at 2, 3 at 3, 3 with no harvest ever = 833. Identical to `spec.md`'s table
and to the orchestrator's own 830/833 measurement.

Run B, 457 corrections over ~6.68 game-days: 452 attributed, 5 unexplained, **0 false
attributions**.

### (d) Genuinely-unexplained absences, and that they still surface — MEASURED

Run A's 3 unexplained corrections all concern one tile — a cache of goods at (61,6) that
three villagers went for on day 9 — and **no** `agent.chopped`/`agent.quarried` exists at
(61,6) anywhere in the 132,299-event log. Each still produced its own chronicle line,
verbatim in the pre-110 wording:

```
[day 9 19:25] Oak went looking for the cache of goods at (61,6) and found it gone.
[day 9 20:30] Ash went looking for the cache of goods at (61,6) and found it gone.
[day 9 21:38] Birch went looking for the cache of goods at (61,6) and found it gone.
```

Before the change those three sat among 53 look-alikes in that chapter; after it they sit
beside one line that names the ordinary cause. Run B's 5 unexplained corrections likewise
each kept their own line.

### (a) Count and share of absence-themed chronicle entries — MEASURED

### (b) Whether any named absence storyline slug appears — MEASURED

Both come out of the T018 re-narration; the table is in the SC-001 section below. In one
line: **no named absence storyline slug appears in any re-narrated run**, against a named
one in every before-side run carrying 19–43% of that run's entries; and absence-themed
entries fall from **32–50%** of entries to **3–7%**.

## SC-002 (volume) — MEASURED, met

Per chapter, correction lines **before → after** (Run A). "corr" counts every
correction-derived line — the per-event "found it gone" lines plus the one coalesced
summary; "tot" is every line the chapter appended, including those the 120-line ring later
evicted.

| chapter | corr (before) | tot (before) | share | corr (after) | tot (after) | share |
|---|---|---|---|---|---|---|
| day 1, dawn to nightfall | 89 | 137 | 64% | 1 | 49 | 2% |
| the night after day 1 | 4 | 18 | 22% | 1 | 15 | 6% |
| day 2, dawn to nightfall | 92 | 155 | 59% | 1 | 64 | 1% |
| the night after day 2 | 0 | 3 | 0% | 0 | 3 | 0% |
| day 3, dawn to nightfall | 68 | 109 | 62% | 1 | 42 | 2% |
| the night after day 3 | 0 | 8 | 0% | 0 | 8 | 0% |
| day 4, dawn to nightfall | 99 | 152 | 65% | 1 | 54 | 1% |
| the night after day 4 | 15 | 27 | 55% | 1 | 13 | 7% |
| day 5, dawn to nightfall | 51 | 106 | 48% | 1 | 56 | 1% |
| the night after day 5 | 4 | 14 | 28% | 1 | 11 | 9% |
| day 6, dawn to nightfall | 53 | 130 | 40% | 1 | 78 | 1% |
| the night after day 6 | 7 | 14 | 50% | 1 | 8 | 12% |
| day 7, dawn to nightfall | 60 | 116 | 51% | 1 | 57 | 1% |
| the night after day 7 | 3 | 14 | 21% | 1 | 12 | 8% |
| day 8, dawn to nightfall | 92 | 156 | 58% | 1 | 65 | 1% |
| the night after day 8 | 15 | 25 | 60% | 1 | 11 | 9% |
| day 9, dawn to nightfall | 56 | 112 | 50% | **4** | 60 | 6% |
| the night after day 9 | 6 | 20 | 30% | 1 | 15 | 6% |
| day 10, dawn to nightfall | 24 | 77 | 31% | 1 | 54 | 1% |
| the night after day 10 | 9 | 24 | 37% | 1 | 16 | 6% |
| day 11, dawn to nightfall | 23 | 89 | 25% | 1 | 67 | 1% |
| the night after day 11 | 12 | 21 | 57% | 1 | 10 | 10% |
| day 12, dawn to nightfall | 44 | 103 | 42% | 1 | 60 | 1% |
| the night after day 12 | 3 | 13 | 23% | 1 | 11 | 9% |

- **Attributed corrections contribute exactly one line per chapter**, in all 24 chapters of
  Run A and all 13 of Run B. Day 9's four correction lines are one summary plus the three
  genuine anomalies FR-004 keeps by design.
- **Corrections' share of the day chapters falls from a 48–65% band (median 50.5% by the
  replay's own totals; `spec.md` measured median ~57% against its SQL totals) to 1–2%.**
  The worst share anywhere after the change is 12%, on an 8-line night chapter carrying a
  single summary line.
- **Ring overflow, Run A: 5 chapters before → 0 after.** The five overflowing day chapters
  silently evicted 17, 35, 32, 10 and 36 non-correction lines respectively — 130 builds,
  gifts, assemblies and musings dropped. After the change, no Run A chapter reaches the
  ring at all.
- **Run B (finding, recorded honestly): 7 of 13 chapters overflow the ring both before
  AND after.** Corrections were never Run B's overflow driver — its chapters run 148–205
  lines *after* attribution, of which **599 of 921 retained lines are villager musings**.
  The test asserts the precise SC-002 claim (no chapter overflows *on corrections'
  account* — removing their lines would not bring it under the ring) and Run B passes it,
  but the qwen run's line budget is dominated by a different source. That is outside spec
  110's scope and is worth a follow-up card; it is not evidence for or against this
  change.

FR-007 telemetry was captured off the same replay — 22 lines for Run A's 24 chapters, the
two chapters with zero corrections correctly logging nothing (FR-008):

```
mind: chronicle "day 1, dawn to nightfall" corrections: 89 attributed (42 locations, 6 harvesters), 0 unexplained
mind: chronicle "day 9, dawn to nightfall" corrections: 53 attributed (…), 3 unexplained
```

And the coalesced line itself, as the narrator receives it:

```
[day 1 22:00] Ordinary harvesting: 89 remembered things the villagers went for had already
been felled or quarried, at 42 locations, by Ash, Birch, Rowan, Fern, Oak and Sage.
Routine village business.
```

## SC-003 (anti-suppression) — MEASURED, met

Every correction the classifier did not attribute produced its own line, byte-identical to
the pre-110 wording: 3 of 3 in Run A, 5 of 5 in Run B. The harness fails the test if any
unexplained correction emits no line, and fails it if any attributed correction emits one.
Unit coverage for the same property is in `internal/mind/absence_attribution_test.go`.

## SC-004 (precision) — MEASURED, met

**100% precision on 833 replayed real corrections** (Run A) and on 457 (Run B): zero
corrections were classified attributed without a harvest at exactly those coordinates
inside the 4-game-day window, judged against a ground truth built from the raw log rather
than from the ledger under test.

Recall, for completeness (not an SC): 830 of 833 in Run A — the ledger attributed every
correction the independent join says is explainable, so the window and cap cost nothing on
this population.

## SC-001 (the outcome) — T018, re-narration against the live models

Each replayed chapter was handed to a live model with the **production narrator prompts**
(`narrateSystemPrompt` / `narrateUserPrompt`), the production request body
(`{model, messages, stream:false, max_tokens}` + `reasoning_effort` where the run's
`llm.json` set one), the production parser (`parseNarration`), and the production
truncation ladder (800 → 1600 → 3200, `maxTruncationRetries = 2`). The thread list is
**closed-loop**: each chapter is offered the slugs the preceding *re-narrated* chapters
produced, never the original run's — otherwise the old absence slug would be handed back
to the model as a thread to continue.

Two before-sides are reported. The **original live runs** are the true before-side (the
same models narrating the same worlds live, on the un-coalesced buffers). The **replayed
before** runs are a controlled same-model, same-harness comparison — the identical
pipeline over the before buffers, so the only variable is the coalescing.

| run | narrator | entries | absence-themed entries | named absence storyline |
|---|---|---|---|---|
| **before** — original live run A | `gemma4:12b-mlx` | 56 | **18 (32%)** | **`the-missing-trees`** — 15 entries (27%) |
| **before** — replayed run A | `qwen3.6` | 61 | **25 (41%)** | **`vanishing-landmarks`** — 19 entries (31%) |
| **before** — original live run B | `qwen3.6` | 37 | **17 (46%)** | **`the-disappearance-of-res…`** — 7 entries (19%) |
| **before** — replayed run B | `qwen3.6` | 28 | **14 (50%)** | **`vanishing-resources-and…`** — 12 entries (43%) |
| **after** — replayed run A | `gemma4:12b-mlx` | 33 | **2 (6%)** | **none** |
| **after** — replayed run A | `qwen3.6` | 57 | **4 (7%)** | **none** |
| **after** — replayed run B | `qwen3.6` | 32 | **1 (3%)** | **none** |

"Absence-themed" is a word-boundary match of
`missing|vanish*|disappear*|gone|empty|empties|emptying|emptied|absence|absent|thinning`
against the entry text. On the original live run A that measure returns **18** — exactly
the numerator the card recorded as "18 of 90". (The card's denominator, 90, is the line
count of `chronicle.md`, not its 56 entries; the like-for-like entry-denominated baseline
is therefore **18/56 = 32%**, not 20%.)

**(b) is met outright and is the unambiguous half.** Every before-side run — four of
four, two models, two worlds, live and replayed — grew a named absence storyline carrying
19–43% of its entries. **No after-side run grew one.** After the change the storylines are
`hazels-influence`, `village-tensions`, `night-fears-and-necessit`, `birch-sage-suspicion`,
`rowan-fern-contention`, `the-gru-threat` — people and the predator, not the landscape.

**(a) is met on the criterion's intent and MISSED on the letter of the ≤5% bar.**
Absence-themed entries fall 32% → 6% (gemma, run A), 41% → 7% (qwen, run A) and 50% → 3%
(qwen, run B): a 5–17× reduction, but two of the three after-runs land 1–2 points above
the spec's "at most 5%". Reading the residue rather than the number: of the seven
surviving matches across all three after-runs, **none is about a remembered map feature
being gone**. They are villagers' own words and fears — a conversation about
"disappearances in a desolate marsh", Cedar "counting disappearing shadows", a rumor that
Birch has disappeared, a debate over "missing marks" in the forest, Oak confessing to a
missing rock (a cause named, not a mystery). The absence storyline the card opened this
task about is at **0%**.

**Verdict, stated plainly**: SC-001's storyline clause passes; its ≤5% clause fails by
1–2 points on two of three runs when scored with the crude keyword measure, and passes at
0% when scored on what the criterion was written to catch. The spec did not define the
measure, and the criterion's own baseline citation ("18/90 = 20%") uses a denominator
that does not match its numerator. **This is a finding for the planning tier to rule on,
not something Phase 4 will round in its own favour.**

### Environment trait, not a spec finding

`gemma4:12b-mlx` on Ollama emits a long `reasoning` field ahead of its content. On 7 of
run A's 24 chapters it spent the whole ladder (800 → 1600 → 3200 tokens) reasoning and
returned empty or truncated JSON, which the production parser correctly treats as an
unusable reply — "a gap in the story, never a stall". Those chapters, day 9 among them,
produced no entries. This is the known MLX/structured-output trait recorded in project
memory, it is identical on `main`, and it is why the gemma after-run has 33 entries where
the qwen after-run has 57. `qwen3.6` with its own `reasoning_effort: none` had 1 parse
failure in 24 chapters. **The gemma gap does not favour the after side**: the same trait
was in force during the original live run, whose 56 entries still grew the absence
storyline.

## SC-005 (gates)

`go build ./...`, `go test ./...`, and `go test -race ./internal/mind/...` — results in
the commit message that lands this file.

## SC-006 (cross-model) — met

Both models, both worlds:

- **Offline half** (SC-002/SC-003/SC-004) run on both preserved worlds: Run A 833
  corrections / 0 false attributions / 5 → 0 ring overflows; Run B 457 corrections / 0
  false attributions / one line per chapter.
- **Narration half** run on both models: `gemma4:12b-mlx` over all 24 of run A's chapters,
  `qwen3.6` over all 24 of run A's and all 13 of run B's, each with a matching before-side
  pass except gemma's, whose before-side is the original live run itself (a second
  2.6-hour gemma pass would have measured what the live run already measured). Neither
  model names an absence storyline after the change; both name one before it.

## Reproducing this

```
sqlite3 "file:$HOME/Claude/soak-evidence/2026-08-02/soak-world/world.db?mode=ro" \
  ".backup /tmp/runA.db"

# offline: SC-002 / SC-003 / SC-004
PW_REPLAY_DB=/tmp/runA.db PW_REPLAY_OUT=/tmp/runA \
  go test ./internal/mind/ -run TestReplayEvidence -v

# SC-001: re-narrate (PW_RENARRATE_MODE=before for the control pass)
PW_REPLAY_DB=/tmp/runA.db PW_NARRATE_MODEL=qwen3.6:latest PW_NARRATE_EFFORT=none \
  PW_RENARRATE_OUT=/tmp/after.json \
  go test ./internal/mind/ -run TestReplayRenarrate -v -timeout 60m
```

## What Phase 4 did NOT test

- **A fresh live soak.** The route is replay by operator decision (runbook amendment).
  Replay cannot show a *feedback* effect: in a live run the narrator's entries re-enter
  the world as recorded input and colour later villager cognition, so a live soak could in
  principle drift somewhere the replay cannot see. The replay holds the villagers' behaviour
  fixed at the original run's and changes only what the narrator was told.
- **Run B as a 12-game-day sample.** It is a mid-flight snapshot at ~6.68 game-days; every
  Run B count is a floor.
- **One seed.** Both preserved worlds ran seed 1337. They differ by model, not by seed, so
  they are not two independent samples of world behaviour.
- **A follow-up worth carding, found on the way**: Run B's chapters overflow the 120-line
  ring 7 times of 13 both before AND after this change, because **599 of its 921 retained
  lines are villager musings**. The mechanism spec 110 fixes for corrections is live for
  musings on the qwen default, and nothing in this spec addresses it.
