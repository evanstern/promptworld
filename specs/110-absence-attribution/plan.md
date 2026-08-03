# Implementation Plan: Absence attribution (TASK-173)

**Spec**: `specs/110-absence-attribution/spec.md` · **Branch**: `task-173-absence-attribution`
· **Runbook**: `docs/design/task-173-absence-attribution-runbook.md`

## Summary

Give the chronicle narrator a way to tell an ordinary stump from a genuine anomaly, and
stop the ordinary ones from crowding out the rest of the story.

The whole change lives in `internal/mind`. The Mind already absorbs
`agent.chopped`/`agent.quarried` (`mind.go:315-346`) and already renders chronicle lines
in the same absorb goroutine against a current replica (`narrate.go:64` `chronicleNote`).
We add (1) a bounded coordinate-keyed **harvest ledger** fed by those two absorb arms,
(2) a **classifier** on the `agent.map_corrected` line path that asks the ledger whether
the vanished thing was harvested within 4 game-days, (3) a per-chapter **tally** that
replaces N attributed correction lines with **one** attributed summary line at
`closeChapter`, and (4) a prompt line telling the narrator that summary is background,
not storyline.

Unexplained corrections are untouched — same wording, same position. That is the whole
of AC#3's mechanism: three anomalies among 830 look-alikes are invisible; three beside
one "ordinary harvesting" line are conspicuous.

## Technical Context

**Language/stack**: Go, single module. No new dependencies.

**Files expected to change**

| file | change |
|------|--------|
| `internal/mind/narrate.go` | ledger type + tally; `chronicleNote`'s `agent.map_corrected` arm routes attributed vs unexplained; `closeChapter` emits the summary line and resets the tally; narrator prompt gains the background-line instruction |
| `internal/mind/mind.go` | the existing `agent.chopped`/`agent.quarried` absorb arm additionally records the harvest into the ledger (the arm already decodes `HarvestPayload` — `{Agent, X, Y}`, `agents.go:1500-1504`) |
| `internal/mind/narrate_test.go` (or a new `absence_attribution_test.go`) | unit coverage for FR-002/003/004/008 and SC-003/SC-004 |
| `docs/wiki/chronicle.md` | re-verify + re-pin: `internal/mind/narrate.go` is one of its pinned sources |
| `docs/player/` | regenerate if any `docs/wiki/` note changes (spec 069 gate) |

**Files deliberately NOT changed**: everything under `internal/sim` (no payload, event
type, or `format_version` change — FR-006, spec 092/094), `internal/tui` (so the
spec-047 design gate does not fire, and `main`'s pre-existing `tui-design` staleness
stays someone else's), and the spec 097/081 memory-side machinery.

**Key constants to introduce** (named, commented, in the `narrate.go` const block beside
`narrMaxLines`):
- `harvestLedgerWindow` — 4 game-days in ticks. Justified by measurement: the observed
  harvest→correction lag tops out at 3 game-days (716 same-day, 76 at 1, 35 at 2, 3 at 3,
  0 beyond), so 4 covers 100% of the explained population with a margin.
- `harvestLedgerCap` — hard entry cap (352 distinct locations over 12 game-days observed;
  a cap an order of magnitude above that is ample). Eviction is oldest-first, and the cap
  exists so a long run cannot grow the ledger without bound, not as a tuning dial.

**Determinism posture**: the ledger is derived purely from absorbed events applied in
order by the single absorb goroutine, so two replays of the same log build the same
ledger. Nothing about it reaches the sim; narrator output continues to enter the world
only as recorded input through `InjectSocial`.

## Constitution Check (v1.3.0)

Run before Phase 0 and re-checked after design. **Result: PASS, no Complexity Tracking
entries required.**

| principle | check | verdict |
|-----------|-------|---------|
| **I. Artifact-Grounded Action** | Evidence is the preserved 12.02-game-day soak world, queried in-spec with the exact SQL buckets reproduced in `spec.md`; the runbook, spec, and board card all carry it. Nothing rests on a chat turn. | PASS |
| **II. One Task, One PR** | TASK-173 → one branch (`task-173-absence-attribution`) → one PR. `tasks.md` phases are internal breakdown, committed to that branch. | PASS |
| **III. Gates Over Assertions** | Success criteria are gate-shaped: build/test/-race, the merge-drift `pr` mode, the spec-069 wiki+player-docs gates, and the runbook's ≥12-game-day evidence bar with four named measurements. No criterion is satisfied by assertion. | PASS |
| **IV. Grounding Freshness** | `internal/mind/narrate.go` is a pinned source of `docs/wiki/chronicle.md`; the branch re-verifies and re-pins it in-branch and regenerates `docs/player/` if the wiki changes. | PASS |
| **V. Model-Tiered Workflow** | Planning (this document, `spec.md`, `tasks.md`) on `claude-opus-5` in the orchestrating session; implementation delegated to `.claude/agents/spec-implementer-opus.md` (`claude-opus-5`, fallback `claude-opus-4-8`) — never inline. Rubric lines fired and the model that served are recorded on the board task. | PASS |
| **Additional constraints** | `backlog/` touched only via the CLI, only at root, only in board-scoped commits; `specs/110-*` is this feature's source of truth and is bridge-linked (marker landed on `main` at `46222529`). | PASS |
| **Spec rigor** | Non-trivial (narrative behaviour change with a soak-scale evidence bar) → full Spec Kit. specify ✓ · clarify — not run, see below · plan ✓ · tasks → next. | PASS |

**On `speckit-clarify`**: the constitution requires it "where ambiguity exists". The two
ambiguities this feature had are both resolved from artifacts rather than preference, per
Principle I: the attribution *locus* (D1 — settled by the card's own re-open block, which
states the memory layer is working and the failure is downstream at narration) and the
*rendering* choice (D2 — the runbook's operator checkpoint 1, decided with the
recommendation stated and flagged to the operator, reversible in a follow-up). No third
ambiguity remains that a clarify round would resolve, so the round is skipped with the
reason recorded here rather than run as ceremony.

## Phases

**Phase 1 — Ledger + classifier (no behaviour change yet)**
Introduce the ledger type, its window/cap eviction, and the absorb-arm population.
Introduce the classifier as a pure, directly-testable function
(`attributedHarvest(fact, tick) (agentID int, ok bool)`). No change to what
`chronicleNote` emits yet, so this phase is provably inert — tests assert the ledger and
classifier alone. Kept separate because it is the only phase whose correctness can be
pinned against real replayed data (SC-004) without narration in the way.

**Phase 2 — Coalesced narration**
Route `chronicleNote`'s correction arm through the classifier; accumulate the tally;
emit the single summary line at `closeChapter`; leave unexplained corrections' lines
byte-identical. Includes FR-008's zero-correction identity case.

**Phase 3 — Prompt + telemetry**
Mark the summary line as background in the narrator prompt (FR-005); record per-chapter
attributed/unexplained counts on the existing telemetry path (FR-007) so the soak can
read the outcome without re-deriving it from the log.

**Phase 4 — Evidence**
Run the ≥12-game-day soak (both models where feasible), record the runbook's four
required measurements on the board task, and tick AC#2/AC#3 against them.

**Phase 5 — Grounding + PR**
Re-verify and re-pin `docs/wiki/chronicle.md` in-branch, regenerate `docs/player/` if the
wiki changed, run the merge-drift `pr` gate, open the PR.

## Risks & mitigations

- **The narrator still finds a storyline in one summary line.** Possible — a single line
  saying "41 remembered spots emptied" could itself read as ominous. Mitigated by FR-005's
  explicit background framing and by naming the harvesters in the line (a cause named is
  not a mystery). Detected by SC-001's soak, which looks for a *named* storyline slug, not
  for the word "gone". If it recurs, the follow-up is prompt-side, not architecture-side.
- **False attribution silences a real mystery.** The window is generous (4 game-days) and
  matching is exact-coordinate, so a genuinely-unharvested location cannot hit the ledger.
  SC-004 pins precision at 100% against replayed real data; D4 records that the failure
  direction was chosen deliberately (a miss restores today's behaviour for one line; a
  false hit would hide an anomaly).
- **A guardian miracle or gru action removes a feature without a harvest event.** Such a
  correction is unexplained by construction and keeps its own line — correct, and the
  desired behaviour: it genuinely is not ordinary harvesting.
- **Chapter boundary races.** The tally is per-chapter state reset alongside
  `md.narrLines` at `closeChapter`, in the same single absorb goroutine — no new
  concurrency. `go test -race ./internal/mind/...` pins it.
- **Soak cost and wall-clock.** ≥12 game-days at 16x is the bar the re-open evidence sets;
  it cannot be traded down without a runbook amendment plus an operator ping (runbook
  checkpoint 4).

## Phase 1 implementation notes (recorded on landing T001–T006)

No deviation from the plan; three details it did not pin down, resolved in the code and
recorded here because Phase 2 depends on them.

1. **Where the ledger lives.** The type, its eviction, and `attributedHarvest` are in
   `internal/mind/narrate.go` (beside the constants); only the absorb-owned `harvests`
   field on the `Mind` struct and the one-line write in the harvest arm are in
   `internal/mind/mind.go`. `attributedHarvest` is a **method on `*Mind`**
   (`md.attributedHarvest(x, y, atTick)`), delegating to `harvestLedger.lookup` — so
   Phase 2's `chronicleNote` arm calls it directly and unit tests can drive the ledger
   without a `Mind`.
2. **Window and matching semantics** (the edge cases Phase 2's tests will sit on):
   attribution requires `0 <= atTick - harvestTick <= harvestLedgerWindow` — the edge is
   **inclusive**, and a harvest recorded at a tick *after* the correction never attributes.
   A second harvest at the same tile **replaces** the first (most recent explains). Lookup
   is read-only and re-applies the age test, so a stale entry eviction has not yet reached
   still misses. Cap eviction breaks tick ties by `(x, y)` so the surviving set is a
   function of the event log alone (FR-006).
3. **`-race` flake, unattributed.** The first `go test -race ./internal/mind/...` run on
   this branch FAILED (562s) with the failing test name lost to a truncated tail; two
   subsequent full `-count=1` runs passed clean (544s, 545s), as did `go build ./...` and
   `go test ./...`. Phase 1 adds no goroutine, no lock, and no shared state outside the
   absorb goroutine, so it cannot be the cause — but it is on the record so a Phase 2/3
   `-race` failure is not mistaken for a first occurrence. If it recurs, capture the whole
   output (`> file 2>&1`), do not `tail` it.

## Phase 2/3 implementation notes (recorded on landing T007–T014)

No deviation from the plan's design; four things it left open, resolved in the code and
recorded here because Phase 4 (the soak) reads them.

1. **Where the FR-007 telemetry surfaces — what the soak queries.** NOT an event: FR-006
   forbids a new sim event type, and every existing `cog.*` payload is a poor fit. It
   follows the **spec 105 `nightReport` precedent** instead — an absorb-owned in-memory
   counter flushed as ONE summary log line per closed chapter, on the daemon's log. The
   line is emitted from `closeChapter` (`internal/mind/narrate.go`,
   `correctionTally.report`) and its exact shape is:

   ```
   mind: chronicle "day 3, dawn to nightfall" corrections: 40 attributed (12 locations, 2 harvesters), 1 unexplained
   ```

   Note the units agree in number (`1 location, 1 harvester`), so a parser must match
   `locations?` / `harvesters?`.
   Grep the daemon log for `corrections: ` (or the fuller `mind: chronicle .* corrections:`)
   to get per-chapter attributed/unexplained counts directly. The label is the chapter
   label `closeChapter` already builds, so day and phase are on the line. **A chapter with
   zero corrections of either kind logs nothing** (FR-008), and the soak must read an
   absent line as 0/0 rather than as missing data.

2. **The coalesced line's exact wording** (`correctionTally.summary`), stamped like every
   other chronicle line with `[<clock.Format(toTick)>] ` and appended LAST in the chapter
   buffer:

   ```
   Ordinary harvesting: 40 remembered things the villagers went for had already been felled or quarried, at 12 locations, by Ash and Rowan. Routine village business.
   ```

   Singular agreement is handled (`1 remembered thing … at 1 location, by Ash`).
   Harvester names are sorted by agent id so the wording is a function of the event log
   alone (FR-006). The constant `corrSummaryMarker = "Ordinary harvesting:"` is the
   line's identity, and the FR-005 prompt instruction quotes that same constant — SC-001's
   soak query can key off it. Deliberately no negation ("not a mystery") in the line: the
   plan's risk register worried the summary could itself read as ominous, and priming a
   model with the word it must avoid is the wrong mitigation. Framing is carried by the
   named cause plus the prompt.

3. **The FR-005 prompt addition is conditional, not static.** `narrateUserPrompt` scans
   the job's lines for `corrSummaryMarker` and only then appends, after the log and before
   the ongoing-threads block:

   ```
   Any "Ordinary harvesting:" line above is ordinary background, not storyline material: its cause is already named and settled. Do not build an entry or a thread around it.
   ```

   Conditional because FR-008 demands a zero-correction chapter be byte-identical to
   today — a static addition would change every prompt in the world. Scanning the lines
   rather than flagging the job also survives the retry-carry path, where a failed
   chapter's summary line rides into the next chapter's job. The existing
   "Group by storyline, not by hour." instruction is untouched (pinned by test).

4. **Two judgment calls Phase 2 made that the spec did not pin.**
   - *An attributed correction still opens the chapter's window.* It emits no line but
     still sets `md.narrFrom` when that is zero, exactly as its per-event line used to, so
     a chapter whose earliest event is an attributed correction keeps today's `fromTick`.
     The alternative (leaving `narrFrom` to the next narratable event) would have silently
     shortened chapter windows.
   - *Phase 1's `TestChronicleNoteCorrectionLineInertPhase1` is superseded, deliberately.*
     Its first half asserted the attributed case still emits a per-event line — the exact
     assertion Phase 2 inverts — so it is replaced by
     `TestChronicleNoteAttributedCorrectionEmitsNoLine`. Its second half (an unexplained
     correction keeps its line, byte-identical) carries over unchanged and now also pins
     the line's full text. T006's tick stands: Phase 1 *was* inert when it landed.

   Refactor note: `chronicleNote`'s `name` closure became the method `md.agentName` and its
   line-append block became `md.appendNarrLine`, so `closeChapter`'s summary uses the same
   roster path and the same ring discipline (FR-003's "the same roster path the existing
   lines use"). Both are behaviour-preserving extractions.

## Grounding (wiki-in-PR, spec 069)

- `docs/wiki/chronicle.md` pins `internal/mind/narrate.go` — **will** be touched;
  re-verify the note's prose against the diff and re-pin **on this branch**.
- `docs/wiki/agent-mind.md` and `docs/wiki/mental-maps.md` — check their `sources:` at PR
  time; re-pin only if this branch actually touches a listed source (`mind.go` is the
  likely one).
- Re-pins are honest re-pins: read the diff the pin covers, classify RE-PIN-ONLY vs
  NEEDS-REVIEW, amend prose before bumping. A merge-in is never a justification.
- Any `docs/wiki/` change requires regenerating `docs/player/`
  (`node .claude/skills/player-docs/scripts/check-freshness.mjs --check` is the gate's
  probe).
