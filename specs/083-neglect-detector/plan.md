# Implementation Plan: Neglect detector — critical need with zero intents in its class

**Branch**: `083-neglect-detector` (task branch: `task-133-neglect-detector`) | **Date**: 2026-07-26 | **Spec**: [spec.md](spec.md)

## Summary

Encode the TASK-106 research's neglect definition (§1.3) against the shipped substrate:
(1) sim — three reducer arms maintain per-agent `NeglectState` anchors (band-entry,
last class intent, episode latch); a pure heartbeat sweep in `stepEvents` fires ONE
`sim.neglect_detected` + a salience-9 companion memory per (agent, need) episode when a
need has sat below its spec-062 danger band for `neglectWindowTicks` with zero
class-goal intents over the same window. (2) tui — the event joins the shipped
whole-line alert tier (`isAlertType`, the spec-077 `stranger.took` five-touch
precedent) and the map's existing needs-critical overlay is pinned as already covering
the state. (3) validation — CI fixtures derived from Oak's documented death window
(fires, with runway) and healthy windows (silent), plus an env-guarded
`PROMPTWORLD_WORLD01_DB` probe over the real log. Thresholds are promoted-dial-READY
consts, never tuning.json. Wiki re-pins + player docs ride the branch (pr gate).

## Technical Context

**Language**: Go (sim + Bubble Tea TUI). **Testing**: `go test ./...`; scripted-timeline
replay tests (`governor_replay_test.go` idiom); env-guarded recorded-log probe
(`context_replay_test.go` / `replayToTick` precedent); TUI render + `TestCatalogSweep`.
**Scope**: `internal/sim/{agents,state,executor,policy,memory,miracles}.go` + tests;
`internal/tui/{grammar,digest}.go` + `digest_test.go` + render tests;
`docs/wiki/event-types.md` (sweep-required) + reconciliation re-pins; `docs/player/`
regen; `docs/design/tui/` 3 pages; `specs/018-chronicle-digest/contracts/
digest-grammar.md` §3 row. **Constraints**: replay determinism (reducer-only writes;
executor sweep pure over pre-tick state + tick); snapshot byte-identity (`omitempty`
pointer); no new tuning.json dials; no new TUI tier/channel/token; no injection-door
change (executor emission class); rebase taxonomy updated (SHIFT anchors).

## Constitution Check

- **I. Artifact-grounded** — PASS: decision chain is TASK-106 research → operator
  decision 2026-07-25 → TASK-133 → reorient move 13 → this spec; the not-in-repo log
  is handled by pinned fixtures + an evidence-recorded probe, not by assertion.
- **II. One task, one PR** — PASS: TASK-133 ↔ `task-133-neglect-detector` ↔ one PR;
  phases are internal breakdown; the TASK-111 watch composition is spec consideration
  only (AC #3), explicitly not a second deliverable.
- **III. Gates** — PASS: `TestCatalogSweep`, `check-tui-design.mjs --changed`,
  `check-merge-drift.mjs pr` (wiki-repin + player-docs findings), spec-bridge mirror.
- **IV. Grounding freshness** — PASS (planned): touched sources are pinned by (at
  least) `executor-needs-survival.md`, `sim-state-agent-fields.md`,
  `sim-state-apply-agents.md`, `sim-state-intent-lifecycle.md`, `reflex-policy.md`,
  `agent-memory-window.md`, `guardian-miracle-rebase-taxonomy.md`, `event-types.md` (+
  the routed child note), and the chronicle/map TUI notes; reconciliation computed from
  the actual branch diff, re-pinned IN this branch; `docs/player/` regenerated; merge
  with `gh pr merge --merge` only (squash rewrites branch pins — observed hazard).
- **V. Model tiers** — PASS: this spec/plan/tasks cycle is the planning tier;
  implementation dispatches to `spec-implementer` on **Opus 4.8** (recorded on
  TASK-133: reducer/percept event + high-salience memory injection + world-01 log
  validation; cognition-adjacent).

**Post-Phase-1 re-check**: PASS — no new violations; Complexity Tracking empty.

## Design

### D1 — Derived state + reducer arms (sim)

- `internal/sim/agents.go`: `Neglect *NeglectState` on `Agent` (omitempty pointer,
  Journal/Hail/Map precedent) + the `NeglectState` struct and need-keyed accessors —
  data-model.md §1 shapes verbatim. Spec-083 doctrine const block: `neglectWindowTicks
  = 7200` (data-model §5 comment).
- `internal/sim/state.go`: extend the `agent.needs_changed` arm (~1718) — band anchors
  set/cleared + latch cleared on recovery, per data-model §2; extend the
  `agent.intent_set` arm (~845/872) — `needClassOf(p.Goal)` stamp after `appendIntent`;
  new `sim.neglect_detected` arm — set the need's Fired latch.
- `internal/sim/policy.go`: `needClassGoals` map + `needClassOf` beside the
  goal-resolver registry (data-model §6); `TestNeedClassGoalsResolve` pins every member
  against the registry (anti-rot, research §2).
- `internal/sim/miracles.go`: `rebaseTicks` shifts the six `*Since`/`*Intent` anchors
  (non-zero only); taxonomy doc comment amended. Extend the existing rebase test.

### D2 — Detector sweep + payload + memory (sim)

- `internal/sim/executor.go`: inside the `%60` needs-heartbeat block (beside the
  near-death latch), per living **awake** agent, per need in {food, warmth, rest}:
  evaluate the factored predicate `neglectDue(a, need, nextTick) bool` (pre-tick state
  only — the `recoveryHoldEvents` purity precedent; factored so the D5 probe can call
  it over replayed state). On true: emit `sim.neglect_detected`
  (`NeglectDetectedPayload{Agent, Need, Level, Since}`, data-model §3) then the
  companion `situatedMemoryEvent(…, salNeglect, PlaceAt, "", OriginWitness,
  neglectMemoryText(need))` — event first, memory second, same batch (map-corrected
  companion shape). Fixed need iteration order (food, warmth, rest) for determinism.
- `internal/sim/agents.go` (payload home ~1262): `NeglectDetectedPayload` beside
  `NeedsPayload`/`DiedPayload`.
- `internal/sim/memory.go`: `salNeglect = 9` in the salience table (comment: joins the
  near-death/exile interrupt band deliberately — research.md R6) + the three fixed
  per-need texts (data-model §4).
- No `loop.go` change: executor emission class needs no whitelist entry (the
  `charge_regenerated` doctrine comment).

### D3 — Chronicle alert + map pin (tui)

- `internal/tui/grammar.go`: add `"sim.neglect_detected"` to `isAlertType`'s switch
  (the ONLY alert wiring — `styleFeedAlert` whole-line path already exists).
- `internal/tui/digest.go`: `digestRegistry["sim.neglect_detected"]` — deterministic
  per-need wording, e.g. `⟨Name⟩ is dangerously cold and has done nothing about it
  (warmth 0)` / `…starving… (food N)` / `…exhausted… (rest N)`; comment `// alert
  (neglect tier — spec 083)` (the stranger.took comment precedent).
- `internal/tui/digest_test.go`: `catalogFixture` row (sample payload + expected plain
  summary) — `TestCatalogSweep` enforces registry↔fixture both directions.
- Map: **no code**. New render test pins that an agent fixture in the neglect-firing
  state paints `styleAgentCritical` via the existing `needsCritical` overlay
  (subsumption contract, spec FR-012).
- `specs/018-chronicle-digest/contracts/digest-grammar.md` §3: add the type's row with
  the exact wording (spec-077 precedent).

### D4 — Fixture validation (CI-binding, board AC #1)

New `internal/sim/neglect_test.go` (+ the sweep/arm units where they live):

- **Arm units**: needs arm sets/clears anchors + latch on band crossings; intent arm
  stamps exactly the classed goals (table over `needClassGoals` + a non-class goal);
  `sim.neglect_detected` arm sets the latch.
- **Oak-shaped fixture** (documented shape: warmth 636→0 at 4/min, only reflex `chop`
  + planner `wander` intent records): fold the scripted history through `Apply`, run
  `stepEvents` across heartbeats — fires exactly once at band-entry + 7200 with the
  spec's arithmetic (warmth 0, health ≈ 900 on the same trajectory ⇒ ≈5 h runway
  asserted as "fires before the death tick the fixture's decay produces"); a second
  full window adds nothing; recovery above 350 then relapse + window fires once more.
- **Healthy fixtures**: class intent inside every window (Oak-day-4 shuttling shape) ⇒
  silent; dip-and-recover before T ⇒ silent, anchors reset; asleep at the would-fire
  heartbeat ⇒ silent that beat (fires after waking if still due).
- **Replay determinism**: live-driven firing world vs genesis replay of its log —
  `live.Hash() == replayed.Hash()` (`governor_replay_test.go` idiom); memory seq
  identical (stampSeqs path).
- **Snapshot byte-identity**: existing fixture snapshots load/marshal unchanged
  (`omitempty` — piggyback the standing round-trip tests).

### D5 — World-01 probe (env-guarded, board AC #1 evidence)

`internal/daemon/` beside `TestSageThrashWindowContextReplay` (same env guard +
copy-to-tempdir + manifest parse + `worldmap.Generate` + `replayToTick` mechanics):
skip unless `PROMPTWORLD_WORLD01_DB` is set; replay the recorded log to sampled ticks
inside Oak's final ~6 h (death ≈ tick 511,440) and assert `sim`-exported
`NeglectDue(agent, need, tick)` (the D2 predicate, exported for this consumer) holds
for (Oak, warmth); replay to sampled ticks inside labeled healthy episodes
(raw_results.json `episodes` t0/t1 ranges — e.g. Oak day 4) and for Ash/Hazel and
assert it does not. A run's output is recorded as evidence on the task (the spec-043
`evidence/sc-004-replay.md` precedent). Honesty note rides the test's doc comment: the
log is machine-local; CI validation is D4.

### D6 — Design reference amendments (same PR)

- `docs/design/tui/patterns/chronicle-grammar.md`: alert-tier enumeration + color-roles
  `alert` row gain `sim.neglect_detected` (spec 083).
- `docs/design/tui/panels/chronicle.md`: alert-tier mention re-verified.
- `docs/design/tui/panels/map.md`: condition-overlay section names neglect as covered
  by the needs-critical overlay (no new token/glyph/legend row).
- `node scripts/check-tui-design.mjs --changed` from the worktree; re-verify + re-pin
  every flagged page.

### D7 — Wiki + player docs (in-branch, pr-gate enforced)

- `/grounding-wiki:wiki-update` reconciliation over the branch diff; expected
  review-work notes: `executor-needs-survival.md` (the sweep is its subject),
  `sim-state-agent-fields.md` (new Neglect field), `sim-state-apply-agents.md` /
  `sim-state-intent-lifecycle.md` (extended arms), `reflex-policy.md` (policy.go
  dictionary), `agent-memory-window.md` (salNeglect joins the interrupt band — its
  "kept below 9 on purpose" sentence must be amended), `event-types.md` (+ routed
  child note — likely `event-types-agent-vitals.md`; sweep-required backtick),
  `guardian-miracle-rebase-taxonomy.md` (new SHIFT rows), `guardian-survival-watches.md`
  (the composition seam, if judged in-scope), and the chronicle/tiles TUI notes;
  computed re-pins for the rest.
- Regenerate `docs/player/` if any wiki note changes (`player-docs` skill; probe run
  directly: `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`).
- Gates: `node scripts/check-merge-drift.mjs pr` exits 0 from the worktree before the
  PR opens; merge with `gh pr merge --merge` ONLY. Post-merge bookkeeping (board ticks,
  spec-bridge sync, tasks.md, runbook row) is authored on a branch and lands on main by
  merge — TASK-160: nothing commits directly at root.

## Project Structure

### Documentation (this feature)

```text
specs/083-neglect-detector/
├── CLAIM.md          # claim stub (spec 065 / TASK-160 flow) — kept
├── spec.md
├── research.md
├── data-model.md
├── plan.md           # this file
└── tasks.md
```

### Source Code (repository root)

```text
internal/sim/agents.go       # NeglectState + payload + neglectWindowTicks (D1/D2)
internal/sim/state.go        # three reducer arms (D1)
internal/sim/policy.go       # needClassGoals dictionary (D1)
internal/sim/executor.go     # heartbeat sweep + NeglectDue predicate (D2)
internal/sim/memory.go       # salNeglect + per-need texts (D2)
internal/sim/miracles.go     # rebaseTicks SHIFT rows (D1)
internal/sim/neglect_test.go # fixtures + replay determinism (D4)
internal/daemon/*_test.go    # env-guarded world-01 probe (D5)
internal/tui/grammar.go      # isAlertType membership (D3)
internal/tui/digest.go       # digestRegistry entry (D3)
internal/tui/digest_test.go  # catalogFixture row (D3)
internal/tui/*_test.go       # alert render + map-overlay subsumption pin (D3)
specs/018-chronicle-digest/contracts/digest-grammar.md  # §3 row (D3)
docs/design/tui/{patterns/chronicle-grammar,panels/chronicle,panels/map}.md  # D6
docs/wiki/** · docs/player/**                            # D7
```

**Structure Decision**: existing packages only — detection truth in `internal/sim`
(executor emission class + reducer-derived anchors), presentation membership in
`internal/tui`. No new packages, channels, or seams beyond one predicate function and
one dictionary that deliberately live next to what they must not drift from.

## Complexity Tracking

Empty — no constitution violations.
