# Implementation Plan: Report-card truth — unify all card surfaces on sim.EvaluateRubric

**Branch**: `072-report-card-truth` (task branch: `task-149-report-card-truth`) | **Date**: 2026-07-26 | **Spec**: [spec.md](spec.md)

## Summary

Two moves, one doctrine. (1) sim: persist the charter authorship flag
(`State.CharterCustom = !CharterObservedPayload.Default`) via the existing
`metatron.charter_observed` reducer arm, and add `EvaluateRubric`'s `the-law` arm over
state facts (`Norms` + the new flag) — removing the documented blocker at
`internal/sim/scenario.go:277`. (2) tui: replace the three presence-based fact builders
with ONE shared resolver whose sources are, in precedence order: recorded pass
(instrument, all-met) → `EvaluateRubric` concluded (run ended) → `EvaluateRubric` live —
routed through by all three card surfaces. Design pages amended and re-pinned in the same
PR; wiki re-pins and player-docs regen ride the branch (pr gate).

## Technical Context

**Language**: Go (sim + Bubble Tea TUI). **Testing**: `go test ./...`; table-driven sim
rubric tests (`TestFirstNightRubricTable` precedent); Model-fixture TUI render tests.
**Scope**: `internal/sim/state.go`, `internal/sim/guardian.go`,
`internal/sim/scenario.go`, `internal/tui/views.go`, `internal/tui/reportcard.go`, their
tests; `docs/design/tui/` (3 pages amended + all flagged pages re-pinned); `docs/wiki/`
re-pins + `docs/player/` regen as the diff demands. **Constraints**: replay determinism
(new state via reducer only); snapshot byte-identity for existing worlds (`omitempty`);
`reportCardView` renderer and `consoleCard` seam contracts unchanged; no change to
`scenarioRubricEvents`' emission set (spec FR-009).

## Constitution Check

- **I. Artifact-grounded** — PASS: decision chain is reorient decision 1 → TASK-149 →
  this spec; design/wiki amendments are tasks, not afterthoughts.
- **II. One task, one PR** — PASS: TASK-149 ↔ `task-149-report-card-truth` ↔ one PR;
  phases are internal breakdown.
- **III. Gates** — PASS: `check-tui-design.mjs --changed`, `check-merge-drift.mjs pr`,
  and the spec-bridge mirror gate all run at their choke points.
- **IV. Grounding freshness** — PASS (planned): touched sources are pinned by (at least)
  `report-card-renderer.md`, `scenario-machinery.md`, `scenario-machinery-surfacing.md`,
  `guardian-report-card.md`, `tui-dock-tabs.md`, `takeover-surfaces.md`,
  `sim-state-world-fields.md`, `sim-state-reducer.md`, `event-types-guardian-morgue.md`
  and the other notes listing `views.go`/`state.go`/`guardian.go`; the reconciliation is
  computed from the actual branch diff and re-pinned IN this branch; player docs
  regenerated if the wiki changes. Merge with `gh pr merge --merge`.
- **V. Model tiers** — PASS: this spec/plan/tasks cycle is the planning tier;
  implementation dispatches to `spec-implementer` on **Opus 4.8** (tier recorded on
  TASK-149: cross-package sim state/reducer + all TUI card surfaces; doctrine-adjacent —
  grading truth at the teaching moment; persists charter Default into state).

**Post-Phase-1 re-check**: PASS — no new violations; Complexity Tracking empty.

## Design

### D1 — Persist the charter authorship flag (sim, reducer-only writer)

- `internal/sim/state.go` (~line 121, beside `CharterFingerprint`): add
  `CharterCustom bool \`json:"charter_custom,omitempty"\`` with a doc comment naming spec
  072 and the conservative zero value (false = not known player-authored → term pending;
  the "unknown honestly" rule for pre-feature snapshots).
- `internal/sim/guardian.go` `applyGuardian`, `case "metatron.charter_observed"` (~line
  422-434): after `s.CharterFingerprint = p.Fingerprint`, add
  `s.CharterCustom = !p.Default`. The payload has carried `Default` since spec 044, so
  genesis replays of existing logs populate the field with no migration; the v1→v2
  migration (`migrate.go`) carries neither charter field and needs no change.
- No new event type, no whitelist change, no `format_version` bump.

### D2 — the-law production evaluator (sim)

- `internal/sim/scenario.go` `EvaluateRubric` (~line 280): add
  `case "the-law": return theLawRubric(s)`; rewrite the doc comment's blocker sentence
  (lines 276-279) to state the shipped derivation.
- New `theLawRubric(s *State) []RubricTerm` beside `firstNightRubric`:
  - Term 1 `{Label: "a village law adopted", Event: "meeting.proposal_resolved",
    Met: len(s.Norms) > 0, Count: len(s.Norms)}` — every `Norms` entry exists only via
    `resolveProposal` on a passed proposal (`governance.go:397`), so the count is a
    faithful adopted-law ledger (repealed norms stay, adopted-ever semantics — spec
    Assumptions).
  - Term 2 `{Label: "a player-authored charter in force",
    Event: "metatron.charter_observed", Met: s.CharterFingerprint != "" &&
    s.CharterCustom, Count: <1 if CharterFingerprint != "" else 0>}` — latest observation
    wins, matching how the fingerprint itself is kept.
- `scenarioRubricEvents` (~line 389) keeps its `def.ID != "first-night"` early return,
  with its comment updated to cite spec FR-009 (the-law's boundary/evidence/pass emission
  is exercise-catalog content work — evidence needs the observed event's Seq/Tick, which
  state deliberately does not retain).

### D3 — One shared fact resolver (tui)

- New in `internal/tui/reportcard.go` (or `views.go` beside the renderer — implementer's
  call, one file owns it):
  - `reportCardFactsFromRubric(terms []sim.RubricTerm) []reportCardFact` — Term =
    `Label`, Met = `Met`, Backing = `fmt.Sprintf("%s: %d", Event, Count)`.
  - `reportCardFactsFromPass(terms []sim.RubricTerm, evidence []sim.EvidenceRef)
    []reportCardFact` — every fact Met = true (the recorded pass proves all terms held —
    `scenarioRubricEvents` emits only when every term is Met; re-read, never re-grade);
    Backing prefers a matching evidence ref (`"<type> · seq <n>"`, first match by Event
    type), else the rubric backing.
  - `(m Model) resolveReportCardFacts(def sim.ExerciseDefinition, pass
    *sim.CurriculumPass) ([]reportCardFact, reportCardMode)` — the ONE precedence switch,
    generalizing `buildChecklistCard`'s current one: pass != nil → concluded +
    FromPass; `m.runEnded()` → concluded + FromRubric; else live + FromRubric. Labels
    always come from `sim.EvaluateRubric(m.replica, def, m.replica.Tick)`; `m.replica ==
    nil` → no facts (caller renders nothing).
- Rewire the three surfaces to it:
  - `postmortemReportCard` (`views.go:659`): look up this exercise's recorded pass
    (the `buildChecklistCard` loop) and resolve; drop the events-ring derivation.
  - `ceremonyReportCardFor` (`views.go:756`): `provingPass` still identifies the pass;
    found → resolver with that pass; aged-out → resolver with nil pass (concluded
    `EvaluateRubric` fallback — spec edge case).
  - `buildChecklistCard` (`reportcard.go:87`): keep the stopping-point gate exactly as
    is; replace its facts/mode switch with the resolver.
- Delete `reportCardFactsFromCounts`, `reportCardFactsFromEvents`,
  `reportCardFactsFromEvidence`, `humanizeEventType` (`views.go:843-885`) and their
  direct tests (FR-003 — deleted, not stranded).
- Untouched: `reportCardView`, `reportCardMode`, the `reportCard` consoleCard wrapper,
  `rebuildConsoleCards` composition order, the exercise tab (already on
  `EvaluateRubric`; it gains truth on the-law worlds via D1/D2 with zero code change).

### D4 — Tests

- sim (`internal/sim/scenario_test.go` + `guardian_test.go` where the arm lives):
  `TestTheLawRubricTable` on the `TestFirstNightRubricTable` model — default-charter
  observation → charter term unmet; custom → met; norm adoption → law term met with
  count; nothing observed → pending/0. Reducer test: `metatron.charter_observed` with
  `Default: true/false` sets `CharterCustom` false/true. Replay-equivalence assertion:
  fold the same events live vs from genesis, identical `EvaluateRubric` output (US2-5) —
  piggyback the `TestFirstNightGenesisPassAndReplayEquivalence` pattern.
- tui (`views_test.go`/`takeover_test.go`/`reportcard_test.go`): the motivating
  regression — run-ended `first-night` fixture with 2 recorded deaths renders `✗` on "no
  villager dies" with `(agent.died: 2)` in the postmortem (SC-001); recorded-pass fixture
  renders all-met on ceremony AND console card; live fixture renders `…` on unmet;
  cross-surface identity — same fixture, postmortem and console card rows carry identical
  glyph+label pairs (US1-4). Update fixtures now asserting the old generic labels
  (`TestReportCardChecklistOnly`, `TestReportCardBothComposeChecklistAboveNote`,
  `TestConsoleCardSeamComposesReportCard`, ceremony/postmortem render tests).

### D5 — Design reference amendments (same PR — spec FR-010/011)

- `docs/design/tui/overlays/postmortem.md`: rewrite the "Known simplification" block as
  the shipped contract (verdicts from `sim.EvaluateRubric` / the recorded pass; ✗ on
  failure); fix the scored mockup rows (`✗ no villager dies (agent.died: 2)` etc. — real
  labels, truthful glyphs). Leave the `unbuilt (wave 4)` renderer cells to TASK-150.
- `docs/design/tui/overlays/ceremony.md`: rewrite its identical note (evidence path =
  recorded pass, all-met, instrument-authoritative; aged-out fallback = rubric over
  replica).
- `docs/design/tui/panels/exercise.md`: fix the stale "(TASK-127, unbuilt — no coupling
  here)" pointer (line 110); note both cataloged exercises now evaluate for real and
  non-evaluator exercises render pending (the default arm).
- Run `node scripts/check-tui-design.mjs --changed`; re-verify + re-pin EVERY flagged
  page (`views.go` is a pinned source of ~11 pages; re-verify is per-page judgment,
  amendment only where behavior changed).

### D6 — Wiki + player docs (in-branch, pr-gate enforced)

- `/grounding-wiki:wiki-update` reconciliation over the branch diff; expected review-work
  notes: `report-card-renderer.md` (the derivation change is its subject),
  `scenario-machinery.md` / `scenario-machinery-surfacing.md` (EvaluateRubric arm),
  `sim-state-world-fields.md` / `sim-state-reducer.md` (new field),
  `event-types-guardian-morgue.md` / `guardian-report-card.md` (charter_observed arm),
  `tui-dock-tabs.md`, `takeover-surfaces.md`, `curriculum-ladder-progression.md`;
  computed re-pins for the rest. Regenerate `docs/player/` if any wiki note changes
  (`player-docs` skill; probe: `node .claude/skills/player-docs/scripts/check-freshness.mjs
  --check`).
- Gate: `node scripts/check-merge-drift.mjs pr` from the worktree must exit 0 before the
  PR opens; merge with `gh pr merge --merge` (pins are branch hashes).

## Project Structure

### Documentation (this feature)

```text
specs/072-report-card-truth/
├── CLAIM.md             # claim stub (spec 065) — kept
├── spec.md
├── plan.md              # this file
├── tasks.md
└── checklists/
    └── requirements.md
```

### Source Code (repository root)

```text
internal/sim/state.go          # + CharterCustom field (D1)
internal/sim/guardian.go       # charter_observed arm sets it (D1)
internal/sim/scenario.go       # the-law EvaluateRubric arm; doc comments (D2)
internal/sim/scenario_test.go  # TestTheLawRubricTable, replay equivalence (D4)
internal/tui/views.go          # resolver wiring; generic builders deleted (D3)
internal/tui/reportcard.go     # shared resolver home; buildChecklistCard rewire (D3)
internal/tui/{views,takeover,reportcard,console}_test.go  # D4
docs/design/tui/overlays/{postmortem,ceremony}.md          # D5
docs/design/tui/panels/exercise.md                          # D5
docs/design/tui/**                                          # re-pins per --changed
docs/wiki/** · docs/player/**                               # D6
```

**Structure Decision**: existing two-package split (`internal/sim` evaluator truth,
`internal/tui` presentation) — the whole point is that verdicts live in sim and tui only
renders them; no new packages, files, or seams beyond the one resolver function.

## Complexity Tracking

Empty — no constitution violations.
