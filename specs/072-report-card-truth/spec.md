# Feature Specification: Report-card truth — unify all card surfaces on sim.EvaluateRubric

**Feature Branch**: `072-report-card-truth` (task branch: `task-149-report-card-truth`)

**Created**: 2026-07-26

**Status**: Draft

**Input**: TASK-149 / reorient 2026-07-26 decision 1 (docs/design/reorient-2026-07-26-ui.md,
merged position 2 — "one honesty doctrine, two enforcement faces"). Sequenced after
TASK-144 (spec 070, merged — same guardian/report-card code).

## Problem (pinned)

The report card rendered at the game's most salient teaching moment lies. Three surfaces
share one renderer (`reportCardView`, `internal/tui/views.go:809` — the D5
one-renderer-three-sites rule) but grade by **generic event presence**
(`reportCardFactsFromCounts/FromEvents/FromEvidence`, `views.go:849-885`): a term is "met"
the first time its cataloged event type appears at all. For a zero-wanted term this is
backwards — a postmortem after two deaths renders `✓ agent died (agent.died: 2)` on the
exercise's FAILING outcome (`overlays/postmortem.md`'s own mockup shows the false ✓).
Meanwhile the true evaluator, `sim.EvaluateRubric` (`internal/sim/scenario.go:280`) — the
SAME pure derivation the executor's pass emission reads — sits used by exactly one surface
(the exercise tab's live gauges, `internal/tui/exercise.go:50`).

Second face of the same gap: `EvaluateRubric` has a production arm only for `first-night`.
`the-law` (stage-2) falls through to the default arm and renders every term permanently
pending, because its charter conjunct is not state-derivable — the reducer keeps only
`State.CharterFingerprint`, not the recorded `CharterObservedPayload.Default` flag (the
documented blocker in `EvaluateRubric`'s doc comment, `internal/sim/scenario.go:277-279`).

The pedagogy corpus is unambiguous (merged position 2): a false ✓ at the failure moment
mis-teaches the rubric vocabulary the curriculum ladder rests on, and it violates the
morgue's no-blame register, which only teaches if the evidence is true. Downstream, the
TASK-67 fork duel compares report cards — comparing false checkmarks is worse than no duel
(merged position 3's order constraint: this task lands first).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A failed term renders ✗ on every card surface (Priority: P1)

A player whose village dies during `first-night` opens the postmortem takeover. The report
card shows `✗` on "no villager dies" with the honest backing count (`agent.died: 2`) —
because every card surface (postmortem overlay, ceremony takeover, guardian-console card)
now derives its per-term verdicts from `sim.EvaluateRubric`, the same pure derivation the
pass emitter reads. The panel, the emitter, and every card can no longer disagree.

**Why this priority**: this is the board's AC #1 and the decision's core: grading truth at
the teaching moment. Everything else in this spec exists to serve or unblock it.

**Independent Test**: drive a scenario-world Model fixture to a run-end with recorded
deaths and assert the postmortem card renders `✗` on the death term; repeat for the
ceremony (recorded pass → all-met) and the console card seam.

**Acceptance Scenarios**:

1. **Given** a `first-night` world where two villagers died and the run ended with no
   pass, **When** the postmortem takeover renders, **Then** the report card shows `✗` for
   "no villager dies" with backing `agent.died: 2`, and `✗`/`✓` on every other term per
   `EvaluateRubric` — never a presence-derived ✓.
2. **Given** a recorded `CurriculumPass` for the exercise, **When** the ceremony (or any
   card surface) renders that exercise's card, **Then** every term renders met — the
   recorded pass is the instrument and by construction (`scenarioRubricEvents` emits only
   when every term is Met) proves all terms held at pass time; the card is a re-read of
   that record, never a re-grade (spec 063 doctrine).
3. **Given** a scenario world still running (no pass, no run end), **When** the console
   card renders at a stopping point, **Then** unmet terms show the live pending marker
   (`…`) and met terms `✓`, both from `EvaluateRubric` over the replica.
4. **Given** any two surfaces rendering the same exercise at the same replica state,
   **When** both render, **Then** their per-term verdicts are identical (one shared fact
   resolver, by construction).
5. **Given** all three surfaces, **Then** term labels are `EvaluateRubric`'s hand-authored
   plain language ("no villager dies"), not the mechanical event-type gloss
   (`humanizeEventType`'s "agent died") — the generic fact builders are deleted, not
   bypassed.

---

### User Story 2 - the-law has a production evaluator; gauges stop rendering permanently pending (Priority: P2)

A player on a `the-law` world opens the exercise tab. The gauges evaluate for real: "a
village law adopted" flips ✓ when a proposal resolves into a norm; "a player-authored
charter in force" flips ✓ when a guardian turn runs under a custom (non-default) charter —
because the charter `Default` flag is now persisted into sim state by the existing
`metatron.charter_observed` reducer arm, removing the documented blocker.

**Why this priority**: board AC #2. Independently valuable (live gauges become honest on
the second shipped exercise) and it feeds US1: without it, unified card surfaces on a
`the-law` world would truthfully but uselessly render everything pending forever.

**Independent Test**: fold a `metatron.charter_observed` event with `Default: false` into
a state and assert `EvaluateRubric(s, TheLawExercise, tick)` marks the charter term met;
with `Default: true`, unmet. Fold a passed `meeting.proposal_resolved` and assert the law
term met. Replay the same events from genesis and assert identical term results.

**Acceptance Scenarios**:

1. **Given** a `the-law` world with no charter observation yet, **When** the gauges
   render, **Then** the charter term is pending with count 0.
2. **Given** a recorded `metatron.charter_observed` with `Default: true` (the preset
   charter), **When** evaluated, **Then** the charter term stays unmet (SC-004's negative
   case: the game's authorship never satisfies the player-authorship term).
3. **Given** a recorded `metatron.charter_observed` with `Default: false`, **When**
   evaluated, **Then** the charter term is met.
4. **Given** a passed proposal that appended a norm (`State.Norms` non-empty), **When**
   evaluated, **Then** the law term is met with the norm count as backing.
5. **Given** the same event log replayed from genesis on a fresh state, **When**
   `EvaluateRubric` runs at the same tick, **Then** every term result is byte-identical to
   the live fold — the flag enters state ONLY through the event-sourced reducer arm.

---

### User Story 3 - The design reference tells the truth about the shipped behavior (Priority: P3)

An implementer (human or AI) reads `docs/design/tui/overlays/postmortem.md` to build on
the report card. The "known simplification" note describing presence-based grading is
gone — replaced by the shipped `EvaluateRubric` contract — the scored mockup shows the
truthful markers, and `panels/exercise.md`'s stale pointers are amended. The pages are
re-verified and re-pinned in the same PR.

**Why this priority**: board AC #3 and constitution Principle IV — the reference is the
UI authority (spec 047); shipping the behavior change without amending it recreates the
exact doc→code honesty gap this reorientation diagnosed.

**Independent Test**: `node scripts/check-tui-design.mjs --changed` passes on the branch;
the amended pages carry no "known simplification" note describing presence grading.

**Acceptance Scenarios**:

1. **Given** the code change, **When** `node scripts/check-tui-design.mjs --changed` runs
   on the branch, **Then** it passes — every page whose pinned sources this branch touches
   is re-verified and re-pinned.
2. **Given** `overlays/postmortem.md` and `overlays/ceremony.md`, **Then** their identical
   "known simplification" notes are rewritten to describe the shipped rubric-derived
   grading, and postmortem.md's scored mockup shows `✗` on the death row.
3. **Given** `panels/exercise.md`, **Then** its stale "TASK-127, unbuilt" pointer (line
   110) is corrected and the page records that both cataloged exercises now carry
   production evaluators (non-evaluator exercises still render pending — the honest
   default arm).

---

### Edge Cases

- **World snapshotted before the flag existed, custom charter already in force**: the
  snapshot carries `CharterFingerprint` but no `CharterCustom`; the zero value is `false`,
  so the charter term renders pending — honest degradation, never a false ✓. It self-heals
  on the next `metatron.charter_observed` emission (a charter revision); it does NOT
  self-heal merely by waiting, because the pipeline emits only on fingerprint change. This
  is the documented cost of the conservative zero value and is acceptable (the "unknown
  honestly" house rule).
- **Replay of old logs (worlds created before this feature)**: `CharterObservedPayload`
  has carried `Default` since spec 044, so a genesis replay populates the new field
  correctly for every post-044 log; pre-044 logs contain no charter events at all and the
  term is honestly pending. No migration arm is needed — the v1→v2 migration
  (`internal/sim/migrate.go`) predates `CharterFingerprint` itself and carries neither.
- **Recorded pass has aged out of the bounded 32-entry `CurriculumPasses` retention**
  (ceremony replay via the `?` overlay): fall back to `EvaluateRubric` over the current
  replica with concluded markers — current-state truth, honest about what it can still
  see. Near-unreachable in v1 (one exercise per world never prunes the ring) but pinned
  here so the fallback is a decision, not an accident.
- **Exercise with no production evaluator arm** (future catalog content): the default
  `EvaluateRubric` arm returns its terms unmet. Live surfaces render pending (`…`);
  concluded surfaces render `✗` — "not provably met" is the honest concluded reading, and
  a recorded pass (which can only exist once an evaluator exists) overrides it anyway.
- **Death lands after a recorded pass** (passed at dawn, villager dies later, run ends):
  the pass is the instrument — card surfaces for that exercise render all terms met
  (scenario 2 above), while the postmortem's morgue rows still carry the deaths. The
  exercise outcome and the run outcome are distinct facts and both render truthfully.
- **Ambient (unscored) world**: unchanged — no rubric exists, no card renders
  (`scenarioExercise` honest fallback, FR-018 boundary preserved).
- **Replica not yet attached** (`m.replica == nil`): no card renders — `EvaluateRubric`
  needs state; a card invented from the events ring alone is exactly the mechanism being
  deleted.

## Requirements *(mandatory)*

### Functional Requirements

Mapped to the three board ACs: AC1 ↔ FR-001..005 (US1), AC2 ↔ FR-006..009 (US2),
AC3 ↔ FR-010..011 (US3).

- **FR-001**: All three report-card surfaces — postmortem overlay
  (`postmortemReportCard`), ceremony takeover (`ceremonyReportCardFor`), and the
  guardian-console card (`buildChecklistCard`) — MUST derive per-term facts through ONE
  shared resolver whose verdict source is `sim.EvaluateRubric` over the replica (or a
  recorded pass, FR-002). A failed term on a concluded surface MUST render `✗`.
- **FR-002**: When a recorded `CurriculumPass` exists for the exercise, the resolver MUST
  render every term met (concluded markers), backed by the pass's own `Evidence` where a
  term's event type matches — the instrument-authoritative, re-read-never-re-grade rule
  (spec 063 / FR-019 precedent). No pass + run ended → concluded `EvaluateRubric` facts;
  no pass + live → live-mode (`…`) `EvaluateRubric` facts.
- **FR-003**: The generic presence-based fact builders (`reportCardFactsFromCounts`,
  `reportCardFactsFromEvents`, `reportCardFactsFromEvidence`) and the mechanical label
  gloss (`humanizeEventType`) MUST be removed from `internal/tui` — deleted, not left as
  dead alternates a future call site could resurrect.
- **FR-004**: Term labels on every surface MUST be `RubricTerm.Label` (hand-authored
  plain language) with the backing reference rendered from `RubricTerm.Event` and
  `RubricTerm.Count` (or the pass evidence ref) — the mockups' "village survives to dawn"
  phrasing becomes real, sourced from the evaluator.
- **FR-005**: The shared renderer `reportCardView` and its mode vocabulary
  (met/pending live, met/missed concluded — `contracts/takeovers.md` §4) are UNCHANGED;
  only fact derivation moves. The console card seam contract (`consoleCard`) is unchanged.
- **FR-006**: The reducer MUST persist the charter authorship flag into state: a new
  `State.CharterCustom bool` (JSON `charter_custom,omitempty`) set to
  `!CharterObservedPayload.Default` by the existing `metatron.charter_observed` arm in
  `applyGuardian` — the ONLY writer. No other code path may set it (replay determinism:
  new persisted state enters via the event-sourced reducer path, never ad hoc).
- **FR-007**: `sim.EvaluateRubric` MUST gain a production `the-law` arm evaluating both
  terms from state facts: the law term met when `State.Norms` is non-empty (every entry is
  appended only by a passed `meeting.proposal_resolved` — `resolveProposal`), count = the
  norm count; the charter term met when `CharterFingerprint != "" && CharterCustom`,
  count = 1 when an observation is on state else 0. The doc-comment blocker
  (`scenario.go:277-279`) MUST be rewritten to describe the shipped behavior.
- **FR-008**: `omitempty` on the new field keeps every existing snapshot byte-identical
  (no `format_version` bump) — the house rule every post-044 state field follows.
- **FR-009**: `scenarioRubricEvents` (pass emission) is OUT of scope for `the-law`: it
  keeps its `first-night`-only guard. Authoring `the-law`'s boundary tick, evidence
  assembly (`CharterObservedEvidence` needs the recorded event's Seq/Tick, which state
  does not retain), and pass emission is exercise-catalog content work (reorient decision
  5's wave), not this task's. The evaluator and gauges are this task's deliverable.
- **FR-010**: `docs/design/tui/overlays/postmortem.md` (known-simplification note +
  scored mockup), `docs/design/tui/overlays/ceremony.md` (its identical note), and
  `docs/design/tui/panels/exercise.md` (stale TASK-127 pointer; evaluator coverage note)
  MUST be amended in this PR, and every page `check-tui-design.mjs --changed` flags MUST
  be re-verified and re-pinned in this PR (spec 047 authority gate).
- **FR-011**: The semantic LINT of decision 2 (extending `check-tui-design.mjs` to flag
  `unbuilt (wave` cells on shipped pages) is explicitly OUT of scope — that is TASK-150.
  This task amends prose/mockups its own behavior change touches, nothing more.

### Key Entities

- **`sim.RubricTerm`** — `{Label, Event, Met, Count}` (`internal/sim/scenario.go:266`):
  the already-shipped evaluated-term shape; becomes the single verdict currency for every
  card surface.
- **`State.CharterCustom`** — new event-sourced bool beside `CharterFingerprint`
  (`internal/sim/state.go:121`): whether the most recently observed effective charter was
  player-authored (`!Default`). Zero value = not known player-authored (conservative).
- **`reportCardFact`** — `{Term, Met, Backing}` (`internal/tui/views.go:802`): the
  renderer's row shape, unchanged; now built from `RubricTerm`s or a recorded pass.
- **Shared fact resolver** — one Model-level function replacing the three per-surface
  derivations; the point where pass-vs-rubric-vs-live resolution happens exactly once.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001** (board AC #1): on a deaths-recorded, run-ended `first-night` fixture, all
  three surfaces render `✗` on the death term; a test pins the exact motivating case
  (`agent.died: 2` → `✗`, never `✓`).
- **SC-002** (board AC #2): on a `the-law` fixture, the exercise tab's gauges flip per
  scenarios US2-1..4 — no term renders permanently pending; a genesis replay reproduces
  identical `EvaluateRubric` results (US2-5).
- **SC-003** (board AC #3): `node scripts/check-tui-design.mjs --changed` passes on the
  branch with the three named pages amended; no page any longer documents presence-based
  grading as shipped behavior.
- **SC-004**: `go test ./...` green; existing snapshot fixtures load unchanged
  (`omitempty` — FR-008).
- **SC-005**: the wiki-in-PR gate passes: every wiki note whose pinned sources this branch
  touches is re-verified and re-pinned on the branch, and `docs/player/` is regenerated if
  the wiki changes (`node scripts/check-merge-drift.mjs pr` exits 0 from the worktree).

## Assumptions

- TASK-144 (spec 070, guardian Close join) is merged — the report-card test fixtures this
  spec edits are deterministic; sequencing satisfied (all spec-070 tasks ticked).
- `State.Norms` entries exist ONLY via `resolveProposal` on a passed proposal (verified:
  `appendNorm`'s only callers) — so norm-count is a faithful "law adopted" fact. Repealed
  norms remain on state (`Active: false`) and still count as adopted-ever, which matches
  the exercise's "get a norm adopted" teaching goal.
- The `the-law` charter term reads "a player-authored charter **in force**" as: the most
  recent observation was custom. A later revert to the default charter flips it back off —
  correct for a term about present force, and consistent with how the fingerprint itself
  is kept (latest observation wins).
- The guardian package's replica mirror (`internal/guardian/guardian.go`) needs no change:
  card surfaces read the TUI's own `sim.State` replica, not the guardian's mirror.
- `internal/sim/curriculum.go` is not touched: `CharterObservedEvidence` remains the
  sanctioned pass-evidence constructor for the future emitter (FR-009's content work).
