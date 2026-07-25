# Feature Specification: Takeover surfaces — stage-unlock ceremony + run-end postmortem

**Feature Branch**: `056-takeover-surfaces`

**Created**: 2026-07-25

**Status**: Draft

**Input**: User description: "Takeover surfaces (board TASK-127; reorient 2026-07-25 decision 6 + D6/D13/D5, Wave 4). One takeover-surface family with a deliberate voice asymmetry: the stage-unlock ceremony seizes the screen when curriculum.stage_unlocked fires (success speaks in the player's authorship voice — 'your charter proved The Written Word'); the postmortem seizes it on run.ended (failure speaks in the morgue's no-blame evidence register). Both dismissable and replayable from pull surfaces (explicit ACs). Ships the shared report-card renderer (D5: one renderer, three sites). Design authority: docs/design/tui/overlays/ceremony.md and overlays/postmortem.md (authored by spec 047 at status: specified — both parked operator questions are RESOLVED in the pages: ambient postmortem = morgue evidence only, no report card; ceremony = both voices, instrument authoritative). Fixtures until TASK-119's emitter lands where needed."

## Standing resolutions this spec applies

- **Ambient postmortem contents** (parked question 1): resolved by
  `overlays/postmortem.md` — morgue evidence only; the report card renders
  only on scored/scenario runs, in front of (never instead of) the morgue
  register.
- **Ceremony score voice** (parked question 2): resolved by
  `overlays/ceremony.md` FR-019 — both voices always render together,
  narrated chapter in the player's-authorship register (D6) plus the rubric
  checklist, with the checklist authoritative.
- **Interrupt policy** (question 5): decision 6 stands; the watch item
  (ceremony fatigue / mid-crisis seizure complaints) is recorded on the
  page, not acted on here.
- **Report-card renderer home** (D5): authored on `overlays/postmortem.md`;
  this feature ships it and its ceremony/postmortem call sites; the guardian
  console's inline-card production is TASK-115's (the console's card seam
  already exists via spec 053).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The postmortem takeover (Priority: P1)

When the run ends, the postmortem seizes the screen on every attached
client: the narrated run-end line, then (scored runs only) the report card,
then the morgue's no-blame evidence rows — one row per death with name, day,
cause, and the closest charter observation. `esc` dismisses (the ENDED
posture persists); `p` reopens any time after; attaching to an already-ended
world opens it automatically.

**Why this priority**: run.ended already exists in production (spec 044) —
this story is fully exercisable today without fixtures, and the postmortem
is the family's precedence winner (the other overlay defers to it).

**Independent Test**: end a world (or attach to an ended one); the takeover
opens with morgue rows; dismiss, reopen with `p`; on a scenario fixture,
the report card renders above the morgue rows.

**Acceptance Scenarios**:

1. **Given** an attached client on a live world, **When** `run.ended`
   lands, **Then** the takeover opens immediately (body-replacement, chrome
   visible, exactly terminal-height), leading with the run-end line and the
   morgue evidence rows.
2. **Given** an ambient (unscored) world, **Then** NO report card renders —
   morgue register only. **Given** a scored/scenario run, **Then** the
   report card renders first (met/missed markers — the run is concluded),
   followed by the same morgue rows.
3. **Given** the takeover dismissed with `esc`, **Then** the client returns
   to normal rendering with the ENDED posture and read-only clock keys
   intact; `p` reopens it from anywhere while the run is ended.
4. **Given** a fresh client attaching to an already-ended world, **Then**
   the takeover opens automatically on connect (the dual-source runEnded
   posture).
5. **Given** `q` on the takeover, **Then** normal quit/detach — no "world
   keeps running" framing (it isn't).

---

### User Story 2 - The unlock ceremony takeover (Priority: P1)

When a stage unlocks, the ceremony seizes the screen: a skin-resolved
narrated chapter in the player's own authorship voice ("your charter proved
The Written Word…"), plus the rubric checklist that earned it (the
instrument, authoritative). `esc` dismisses; `q` detaches with the "world
keeps running" blessed-stopping-point framing (D13). Replayable from both
pull surfaces.

**Why this priority**: co-P1 — the celebration half of decision 6; testable
with fixture events today (production emission is TASK-119's, possibly
merged by implementation time).

**Independent Test**: land a fixture `curriculum.exercise_passed` +
`curriculum.stage_unlocked` batch on an attached client; the ceremony opens
with both voices; dismiss and replay from the `?` overlay entry.

**Acceptance Scenarios**:

1. **Given** an attached client, **When** `curriculum.stage_unlocked`
   lands, **Then** the ceremony opens immediately: title (`<STAGE NAME> —
   unlocked`), narrated chapter (D6 authorship voice, fiction via skin
   tokens), and the rubric checklist for the proving exercise — both
   always, checklist authoritative (FR-019).
2. **Given** the ceremony open, **When** the player presses `q`, **Then**
   detach with the "the world keeps running" affordance (D13); `esc`
   dismisses one layer as everywhere.
3. **Given** a dismissed or missed ceremony, **Then** its content is
   reachable from BOTH pull surfaces: `promptworld stages` (facts,
   already shipped) and a new ceremony-replay entry in the `?` overlay
   (stored content, never regenerated) — a player is never permanently
   denied a milestone's content.
4. **Given** `run.ended` lands while the ceremony is open, **Then** the
   ceremony is dismissed and replaced by the postmortem (postmortem always
   wins); **Given** an unlock lands while the postmortem is open, **Then**
   the ceremony defers (replayable later, never interrupts).

---

### User Story 3 - The shared report-card renderer (Priority: P2)

One rubric-checklist renderer serves three sites: the postmortem (met/missed
— concluded), the ceremony (the earning evidence), and — via the console's
existing card seam — TASK-115's future inline cards (met/pending — live).
Rows: plain-language term, marker, backing event reference.

**Why this priority**: D5's one-renderer rule is what keeps three surfaces
from drifting; P2 because both P1 stories embed it (it ships inside them)
— this story is its extraction and contract.

**Independent Test**: render the same rubric fixture through the renderer
in concluded and live modes; assert identical row content, differing only
in marker vocabulary; plug a fake into the console card seam and assert it
composes.

**Acceptance Scenarios**:

1. **Given** an exercise rubric with mixed outcomes, **When** rendered
   concluded, **Then** met/missed markers; **When** rendered live, **Then**
   met/pending — same rows, same backing references, same truncation
   discipline.
2. **Given** the console's card seam (spec 053's consoleCard interface),
   **When** a report card is wrapped as a card, **Then** it composes in the
   seam's slot without modification — proven by a test, even though
   production wiring is TASK-115's.

---

### Edge Cases

- **No exercise on a scored-looking world / no rubric data**: the report
  card never renders on missing data (the postmortem falls back to the
  ambient form; honesty over decoration).
- **Multiple unlocks queued** (stage 2 and 3 in quick succession —
  hypothetical): ceremonies render one at a time; a ceremony arriving while
  one is open replaces it (same-kind takeovers don't stack; the newest
  milestone wins, both remain replayable).
- **Help overlay open when a takeover fires**: the takeover wins the body
  slot (help is dismissible chrome, the takeover is the event); `?` works
  again after dismissal.
- **`p` on a live (non-ended) world**: inert (the key only exists while
  runEnded()).
- **Narrow terminals**: both takeovers are layout-independent
  (full-screen in narrow, same content, standard wrapping).
- **Skin tokens before/after TASK-121's merge**: the ceremony's fiction
  strings resolve through the skin lookup if merged (expected — Lane 3
  ordering), else via the existing `skin.StageName` + the D6 voice as
  compiled default text; either way no NEW bare fiction literal that
  TASK-121's sweep test would flag.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: A postmortem takeover MUST open immediately on `run.ended`
  (live push) and automatically on attach to an ended world (dual-source
  posture), rendering: run-end narrated line; report card (scored runs
  only, concluded markers); morgue evidence rows (always). Ambient runs
  render NO report card.
- **FR-002**: A ceremony takeover MUST open immediately on
  `curriculum.stage_unlocked`, rendering the D6 authorship-voice narrated
  chapter AND the authoritative rubric checklist, always both (FR-019 as
  authored).
- **FR-003**: Precedence MUST be: postmortem always wins — it replaces an
  open ceremony; an unlock during an open postmortem defers (replayable,
  never interrupting). Takeovers never stack; body-replacement discipline
  (chrome visible, exact terminal height) holds in widescreen and narrow.
- **FR-004**: Dismissal/replay MUST match the pages: `esc` dismisses both;
  `q` on the ceremony carries the D13 "world keeps running" stopping-point
  framing, `q` on the postmortem is the plain ended-world quit; `p`
  (global, ended-only) reopens the postmortem; the `?` overlay gains a
  ceremony-replay entry (stored content) and `promptworld stages` remains
  the CLI pull surface.
- **FR-005**: The shared report-card renderer MUST ship with concluded
  (met/missed) and live (met/pending) marker modes over rubric terms +
  backing event references, used by both takeovers and composable into the
  console's existing card seam (production wiring there is TASK-115's).
- **FR-006**: All content MUST derive from recorded events/state (morgue
  register facts, rubric evidence, unlock records) — model-free, stored,
  never regenerated; no new IPC fields beyond what exists.
- **FR-007**: The design reference MUST be amended in the same PR: both
  overlay pages specified → shipped (real symbols; precedence/replay rows
  filled), `patterns/keymap.md` (`p`, takeover keys, parity gaps),
  `overlays/help.md` (ceremony-replay entry), `pages/guardian-console.md`
  (card-seam note updated to name the shipped renderer), re-pins
  throughout.
- **FR-008**: Fiction strings MUST resolve through the skin system per the
  edge-case ruling (lookup when TASK-121 is merged; no new bare literals
  regardless).
- **FR-009**: The linear-stream projection (D1) stays sufficient: every
  fact both takeovers show is already on the raw feed / morgue.md /
  `promptworld stages`; this feature adds presentation only.

### Key Entities

- **Takeover state**: which overlay (none/ceremony/postmortem) owns the
  body slot + the deferred-ceremony flag; per the precedence rules.
- **Report-card renderer**: rubric terms × {concluded, live} markers ×
  backing references; one implementation, three sites.
- **Ceremony content record**: the stored facts a replay renders (stage,
  proving exercise, evidence refs) — from the per-user unlocks record +
  event log, never regenerated.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of run endings on an attached client produce the
  postmortem within one frame of the event; 100% of attaches to ended
  worlds open it automatically.
- **SC-002**: The ambient/scored boundary is exact: zero report cards on
  ambient runs, one on scored runs — across the fixture matrix.
- **SC-003**: Precedence holds under interleaving fixtures (ceremony→ended,
  ended→unlock): postmortem wins in both orders; the deferred ceremony
  remains replayable.
- **SC-004**: A dismissed ceremony's content is retrievable via both pull
  surfaces with zero model calls.
- **SC-005**: The same rubric fixture renders identical rows in all three
  renderer sites (postmortem, ceremony, seam-composed card), differing only
  in marker vocabulary.
- **SC-006**: The design gate passes with both overlay pages shipped and
  every touched page re-pinned in-PR.

## Assumptions

- TASK-119's production emitter may or may not be merged when this
  implements; all ceremony-path tests run on fixture events either way (the
  reducer arms exist since spec 046). run.ended is production-real today.
- TASK-125's console card seam (consoleCard) is merged before this
  implements (Lane ordering); if not, the renderer still ships and the
  seam-composition test lands with 115.
- The `?` overlay ceremony-replay entry is a content addition under spec
  045's content contract (amended deliberately, per D9's precedent of
  deliberate help-content amendments).
- Model tier: Sonnet (view/rendering + overlay state machine in one
  package; escalation per rubric if focus/overlay interactions prove
  architectural).
- Dependency note: TASK-121's skin contract expected merged first (Lane 3
  ordering); the edge-case ruling covers either order.
