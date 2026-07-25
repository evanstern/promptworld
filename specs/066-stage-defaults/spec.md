# Feature Specification: Stage-shaped TUI layout defaults

**Feature Branch**: `066-stage-defaults`

**Created**: 2026-07-25

**Status**: Draft

**Input**: User description: "Stage-shaped TUI layout defaults (TASK-128, reorientation 2026-07-25 decision 3, Wave 4). Which panels/tabs/chrome are visible BY DEFAULT in the promptworld TUI is stage-resolved... The authority table for every per-surface default value is docs/design/tui/patterns/stage-defaults.md; the spec must derive requirements from that table rather than inventing values. Acceptance: (1) stage-resolved default visibility exists with every surface reachable at every stage; (2) pre-ladder worlds byte-identical to the ungated full layout."

## The authority page governs

`docs/design/tui/patterns/stage-defaults.md` (spec 047 corpus, `class: pattern`) is
the single authority for every per-surface, per-stage default value in this
feature. This spec derives its requirements from that table and NEVER restates a
value the table owns — if a value below ever disagrees with the page, the page
wins and this spec is defective. Changing a default is a change to that page
first, this feature's implementation second.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A stage-1 player boots into the focused layout (Priority: P1)

A player starting a fresh stage-1 world opens the TUI and sees only the surfaces
that matter to a beginner: the map, the narrated chronicle, the guardian strip,
and the lesson row. The trace/raw-feed density and operator-grade chrome that
would overwhelm a first session are not in the default frame — but nothing is
locked away: the help overlay, solo views, and each surface's own pull path still
reach everything.

**Why this priority**: This is reorientation decision 3's core promise — the
learning game boots at the altitude of the learner. Without it the stage ladder
teaches into a cockpit built for stage 4.

**Independent Test**: Boot a stage-1 world; compare the default visible set
against the authority table's Stage 1 column; then reach every non-default
surface via help overlay / solo view / pull path without changing the world's
stage.

**Acceptance Scenarios**:

1. **Given** a world whose curriculum stage is 1, **When** the TUI boots,
   **Then** the starting visible set is exactly the authority table's Stage 1
   column (lesson row on, guardian strip on, villager strip on, systems tab on,
   exercise tab iff the world carries a scenario, incident vocabulary
   `forecast`).
2. **Given** the stage-1 default layout, **When** the player invokes any
   non-default surface through the help overlay or its solo view or its pull
   path, **Then** the surface presents its full content — identical to what a
   stage-4 player would see.
3. **Given** any stage's default layout, **When** height pressure forces rows to
   fold, **Then** the fold order is exactly `patterns/layout.md` ruling (a),
   applied on top of the stage's starting set — the order itself never varies by
   stage.

---

### User Story 2 - Pre-ladder worlds are untouched (Priority: P1)

A player running a pre-ladder world (no curriculum stage) sees exactly the
full layout they see today — every default-on surface from every stage's union,
with no visual or behavioral difference from the current ungated TUI.

**Why this priority**: Regression safety is the board task's second acceptance
criterion. Pre-ladder worlds are the existing installed base; decision 3 must be
provably invisible to them.

**Independent Test**: Render a pre-ladder world's frames before and after this
feature at the same terminal size and world state; the output must be
byte-identical.

**Acceptance Scenarios**:

1. **Given** a world with no curriculum stage (`Stage == ""`), **When** the TUI
   renders any frame, **Then** the output is byte-identical to the pre-feature
   ungated full layout at the same size and state.
2. **Given** a world whose stage value is unrecognized, **When** the TUI boots,
   **Then** it takes the pre-ladder posture (fail-open to everything), never a
   narrower set.

---

### User Story 3 - Surfaces arrive with the stage (Priority: P2)

A player whose world crosses a stage boundary mid-session sees the newly
default-on surfaces appear (each announced through the existing first-occurrence
lesson machinery), and surfaces whose default narrows at the new stage adopt
their new posture — unless the player has explicitly toggled that surface this
session, in which case the player's choice wins.

**Why this priority**: The ladder is live progression, not a boot-time
configuration; but it builds on US1's resolution logic, so it lands second.

**Independent Test**: Drive a world across a stage unlock while the TUI is
attached; observe default re-resolution, first-occurrence announcements, and
the preservation of an explicit in-session toggle.

**Acceptance Scenarios**:

1. **Given** an attached TUI on a stage-2 world, **When** the world unlocks
   stage 3, **Then** the lesson row adopts its stage-3 default (badge +
   overlay-only) and the incident vocabulary switches `forecast → fog`, per the
   authority table.
2. **Given** a player who explicitly turned a surface on or off this session,
   **When** a stage unlock re-resolves defaults, **Then** that surface keeps the
   player's explicit setting.
3. **Given** a surface newly default-on at the just-unlocked stage, **When** it
   first appears, **Then** the existing first-occurrence lesson machinery
   announces it exactly once (no duplicate or suppressed lesson).

---

### Edge Cases

- Stage unlock under fold pressure: a newly default-on row on a short terminal
  must enter through the normal fold-order composition (it may be immediately
  folded by `patterns/layout.md` ruling (a)) — never force the body below
  `bodyMin`.
- A scenario world at stage 3+: exercise tab present (world-shaped), incident
  vocabulary `fog` (stage-keyed) — the two axes resolve independently.
- A pre-ladder world carrying a scenario: exercise tab present, incident
  vocabulary `forecast` (the pre-ladder "everything" posture).
- Narrow terminals: the authority table's Narrow column composes with stage
  defaults exactly as it does today (carried rows stay carried; the villager
  strip folds to the header count badge regardless of stage).
- A surface that does not exist yet in the build (e.g. the villager strip before
  its own feature lands): the default-resolution mechanism must tolerate the
  absence — a table row with no surface is inert, not an error.
- Takeovers (ceremony, postmortem) are layout-independent: they fire per the
  authority table regardless of the stage's default visible set.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The TUI MUST resolve a per-surface starting visible set from the
  world's curriculum stage, with every per-surface, per-stage value taken from
  the authority table in `docs/design/tui/patterns/stage-defaults.md` — no
  value hardcoded anywhere the table does not govern.
- **FR-002**: Every surface MUST remain reachable at every stage — via the help
  overlay, its solo view, or its own pull path — and its full content when
  reached MUST be stage-independent (defaults shape placement, never content or
  capability).
- **FR-003**: Pre-ladder worlds (no stage) MUST receive the union of every
  stage's default-on set, rendering byte-identical to the pre-feature ungated
  layout; an unrecognized stage value MUST take the same fail-open posture.
- **FR-004**: Stage defaults MUST compose with the existing fold order
  (`patterns/layout.md` ruling a) by supplying only the starting visible set;
  the fold order and its `bodyMin` guarantee are unchanged at every stage.
- **FR-005**: On a live stage unlock, the TUI MUST re-resolve defaults for the
  new stage, announcing newly appearing surfaces through the existing
  first-occurrence lesson machinery (exactly once), while preserving any
  surface visibility the player explicitly set during the session.
- **FR-006**: Surface presence axes that the authority table marks world-shaped
  (exercise tab: present iff the world carries a scenario) MUST resolve
  independently of stage; stage-keyed vocabulary (incident `forecast`/`fog`)
  MUST resolve from stage alone.
- **FR-007**: Capability gating MUST be untouched: nothing in default
  resolution may read from or write to the guardian capability/stage-ceiling
  machinery (spec 046 doctrine — layout is never a lock).
- **FR-008**: Takeover overlays (ceremony, postmortem) MUST fire independently
  of the stage's default layout, per the authority table's rows.

### Key Entities

- **Stage-default table**: the per-surface × per-stage (1–4, pre-ladder,
  narrow) starting-visibility mapping; owned by the authority page; consumed,
  never redefined, by the implementation.
- **Starting visible set**: the resolved set of default-visible surfaces for
  one world at one moment; input to the existing fold pipeline.
- **Explicit player toggle**: an in-session, per-surface visibility choice made
  by the player; outranks default re-resolution until the session ends.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For each stage 1–4, a booted world's starting visible set matches
  the authority table's column for that stage exactly (automated frame
  assertions per stage; zero deviations).
- **SC-002**: Pre-ladder worlds render byte-identical to the pre-feature layout
  (golden-frame comparison; zero differing bytes across the frame corpus).
- **SC-003**: A reachability sweep over every surface × every stage reaches
  full content in 100% of combinations without changing the world's stage.
- **SC-004**: All pre-existing fold-order and layout tests pass unmodified
  (fold order provably unchanged).
- **SC-005**: A live stage unlock produces each newly-appearing surface's
  first-occurrence lesson exactly once (no duplicates, no suppressions) in the
  unlock test corpus.

## Assumptions

- The authority page `docs/design/tui/patterns/stage-defaults.md` is complete:
  surfaces absent from its table are NOT stage-shaped and boot with today's
  behavior. Adding a stage-shaped surface later means amending the page first.
- "Traces / raw feed" density from the original board prose is governed by the
  table's actual rows (chronicle presentation is not in the table, so it is not
  stage-shaped in this feature); the board card's prose defers to the page.
- The villager strip's table row applies only once that surface exists (its own
  feature, TASK-129, may land before or after this one); default resolution
  tolerates absent surfaces.
- Explicit in-session toggles outrank stage re-resolution (reasonable default;
  no persistence of toggles across sessions is introduced by this feature).
- The existing first-occurrence lesson machinery (spec 055) is the announcement
  channel; this feature adds no new announcement surface.
- Stage resolution reads the world's existing curriculum stage field; no new
  persisted state is introduced.
