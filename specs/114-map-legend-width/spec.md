# Feature Specification: Map Legend Width Policy

**Feature Branch**: `task-191-map-legend-width`

**Created**: 2026-08-02

**Status**: Draft

**Board Task**: TASK-191

**Input**: User description: "The map legend must never exceed the terminal width it is rendered into, at every size, and must signal truncation."

## Overview

The map legend is the strip beneath the map naming what every symbol means — terrain,
structures, agent initials, condition overlays, and (when in view) stockpile and chest
inspection detail. It is the map panel's one designated inspection surface.

At the two terminal widths players actually use, it does not fit, and the two failures
are opposite in kind:

- **Below the widescreen breakpoint (80 columns)** the legend is emitted with no width
  clamp at all — 354–356 columns of text into an 80-column terminal. The terminal
  soft-wraps it into roughly five rows, which pushes the map itself off the top of the
  screen. The player loses the thing the legend exists to explain.
- **At and above the breakpoint (112+ columns)** the legend is clamped, but cut mid-token
  with no marker. At 112 columns the player sees `~water ♠wood "forage` and nothing more
  — three symbols of roughly twenty — with no indication that anything was omitted.

Both are horizontal overflow. This is explicitly **distinct** from the spec 112 FR-008
narrow-fallback caveat, which blesses an 80x30 frame having fewer *rows* than 30 because
the narrow renderer has no fold arithmetic. That caveat says nothing about *columns*.
Having no fold arithmetic is not a licence to emit a 355-column line into an 80-column
terminal.

This feature is also the resolution of a deferral the design authority registered against
itself: `docs/design/tui/panels/map.md:209` deferred the look-cursor question but named
"legend-line overflow pain becomes the actual bottleneck" as its own re-open trigger. The
frame evidence below is that trigger firing.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The legend stops destroying the narrow layout (Priority: P1)

A player runs the client in a standard 80-column terminal — the default width of an
unmaximized terminal on most machines. They open the map. Today the legend wraps into
about five rows and shoves the map grid off the top of the screen, so the primary view of
the village is unusable at the most common terminal size. After this change the legend
occupies exactly one row, and the map stays put.

**Why this priority**: This is the severe one. It breaks the primary screen at the most
common terminal size, and it breaks it by destroying *other* content, not merely its own.
Shipping only this story already delivers a usable narrow map.

**Independent Test**: Render any fixture at 80x30 and confirm no emitted line exceeds 80
columns and the map grid is fully visible. Testable with the existing frame harness alone,
with no dependency on Story 2.

**Acceptance Scenarios**:

1. **Given** a terminal 80 columns wide, **When** the player views the map,
   **Then** the legend occupies exactly one row and no line exceeds 80 columns.
2. **Given** a terminal 80 columns wide, **When** the player views the map,
   **Then** the map grid and its bordering box are fully visible, none of it scrolled
   off-screen by legend wrap.
3. **Given** any committed frame fixture and state at 80x30, **When** the frame is
   dumped, **Then** every line in it is at most 80 columns wide.

---

### User Story 2 - Truncation is honest (Priority: P2)

A player on a 112-column terminal reads the legend and sees it end after three symbols.
Today there is no way to tell whether the village contains only three kinds of terrain or
whether the legend was silently cut. After this change a truncated legend ends in a
visible ellipsis, so "there is more than this" is legible at a glance.

**Why this priority**: A silent cut is a correctness problem in the information the UI
conveys — the player draws a false conclusion about the world. But it degrades
understanding rather than destroying the layout, so it ranks below Story 1.

**Independent Test**: Render at 112x30 and confirm the legend ends in the ellipsis marker
whenever content was dropped, and does not when everything fit. Independent of Story 1's
narrow path.

**Acceptance Scenarios**:

1. **Given** a legend longer than the space available, **When** it renders,
   **Then** it ends in a visible ellipsis marker.
2. **Given** a legend that fits the space available, **When** it renders,
   **Then** it carries no ellipsis marker and is not padded.
3. **Given** a truncated legend, **When** it renders, **Then** the ellipsis is the final
   visible character and the total width still does not exceed the space available.

---

### User Story 3 - A wider terminal is rewarded (Priority: P3)

A player drags their terminal wider. Each additional column of width reveals more of the
legend, up to the full untruncated text — so the player learns that widening the window is
how you read the whole symbol key.

**Why this priority**: This is the property that makes the truncation feel like a
considered policy rather than an arbitrary cap. It largely falls out of Stories 1 and 2,
so it is mostly a guard against a fixed-cap implementation that would satisfy them while
violating the spirit.

**Independent Test**: Render the same fixture and state at increasing widths and confirm
visible legend content is monotonically non-decreasing.

**Acceptance Scenarios**:

1. **Given** two terminal widths where the second is wider, **When** the same fixture and
   state render at both, **Then** the legend at the wider size shows at least as much
   content as at the narrower one.
2. **Given** a terminal wide enough for the entire legend, **When** it renders,
   **Then** the full legend is shown with no ellipsis.

---

### Edge Cases

- **Extremely narrow terminal** (fewer columns than the ellipsis marker plus one
  character): the legend must degrade to something renderable rather than producing a
  negative width, an empty-string crash, or a panic.
- **A legend that fits exactly**: no ellipsis, no padding, no off-by-one truncation of the
  final character.
- **Double-width glyphs at the truncation boundary**: cutting must not land mid-glyph and
  must not produce a line that measures wider than intended once the terminal renders the
  wide character.
- **Styled text**: the legend carries display styling. Truncation must not sever a style
  escape sequence, which would bleed color into the rest of the screen.
- **Inspection content present**: when stockpiles and chests are in view the legend grows
  substantially. The width policy must hold with that content present, not only on the
  bare terrain key.
- **Look-cursor active**: the map title changes while the look cursor is on. The legend's
  width policy must be unaffected by title state.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The map legend MUST render to at most the width made available to it, at
  every terminal size, on both the narrow and widescreen paths.
- **FR-002**: The narrow path MUST clamp the legend at the point of render. It MUST NOT
  emit an over-wide line and rely on the terminal to wrap it.
- **FR-003**: A legend that had content removed for width MUST end in a visible ellipsis
  marker.
- **FR-004**: A legend that fits MUST render unchanged — no ellipsis and no padding.
- **FR-005**: Visible legend content MUST be monotonically non-decreasing in terminal
  width: widening never shows less.
- **FR-006**: Truncation MUST be safe for styled text — no severed escape sequence may
  leak styling beyond the legend.
- **FR-007**: Truncation MUST be measured in display columns, not byte or rune count, so
  that double-width glyphs are accounted for correctly.
- **FR-008**: The legend MUST remain a single row on every path. Wrapping the legend to a
  second row is not an accepted remedy for overflow.
- **FR-009**: The system MUST provide a regression guard that fails when any committed
  frame contains a line wider than that frame's declared terminal width.
- **FR-010**: The guard in FR-009 MUST cover the whole committed frame matrix, not the map
  legend alone, so that future over-width regressions on any surface are caught.
- **FR-011**: The design authority for the map panel MUST record the width policy, and
  MUST resolve the legend-overflow deferral that named this condition as its re-open
  trigger.
- **FR-012**: Behavior above the breakpoint MUST NOT otherwise regress: the legend keeps
  its position as the map panel's last row and its existing content and ordering.

### Out of Scope

- **Segment-priority shedding.** When space is short this feature truncates the tail. It
  does not reorder or selectively drop legend segments (for example, dropping chest
  inspection detail before the terrain key). That is a plausible follow-on and would
  materially improve what a narrow player sees, but it changes what the legend *says*
  rather than only how wide it is, and it is not needed to stop the layout damage.
- **The other two over-width defect classes.** The same frame audit found the header row
  over-width in 11 frames (81–83 columns at 80x30) and the scenario keybind footer
  over-width in 2 frames (121 columns at 112/113). They are separate findings with
  separate causes and are not fixed here — though the FR-009 guard will *detect* them, so
  they must be resolved or explicitly registered before the guard can run clean.
- **The look-cursor / map interrogation surface.** Already shipped under spec 074; this
  feature does not revisit it.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Zero of the 132 committed frames contain a legend line wider than that
  frame's declared terminal width — down from 21 today.
- **SC-002**: At 80 columns the map grid is fully visible, with the legend on exactly one
  row — where today the legend consumes roughly five rows and displaces the grid.
- **SC-003**: A player looking at a shortened legend can tell it was shortened, without
  resizing the terminal or consulting any other screen.
- **SC-004**: Widening the terminal from 80 to 160 columns never decreases the amount of
  legend content shown at any intermediate width.
- **SC-005**: An intentionally over-width line introduced anywhere in the frame matrix
  causes the regression guard to fail.
- **SC-006**: The map design reference states the width policy, and carries no unresolved
  deferral whose re-open trigger is legend overflow.

## Assumptions

- **The approved remedy is clamp-plus-ellipsis, not content reflow.** The operator was
  shown and approved the target shape
  `│ night · [21,8–46,24 of 64×64] · ~water ♠wood "for… │` before this spec was written.
  Tail truncation with an ellipsis is therefore the sanctioned baseline, and
  segment-priority shedding is out of scope above.
- **The narrow legend keeps its current position** outside the map's bordering box, where
  it renders today. This feature clamps it; it does not relocate it, because
  `docs/design/tui/pages/solo-views.md` holds the narrow fallback to "renders unchanged"
  and moving it would exceed that.
- **The committed four-size frame matrix is the verification surface.** It already
  straddles the layout boundaries that matter (80 narrow, 112 at-breakpoint, 113 even
  split, 160 roomy), so no new sizes are needed to prove the policy.
- **The frame harness is the review artifact.** Per the project's TUI design loop, this
  change is reviewed as a before/after diff of real generated frames, not a prose
  description, and every affected committed frame is regenerated in the same branch.
- **Where the shared clipping helper should change is a planning decision**, deliberately
  not fixed here: the ellipsis behavior could land in the shared line-clipping helper used
  across roughly 18 call sites, or in a legend-local helper following the chronicle's
  existing per-surface precedent. Both satisfy every requirement above; they differ in
  blast radius and in how many committed frames are rewritten. `/speckit-plan` decides.

## Evidence

Frame audit over all 132 committed frames in `docs/design/tui/frames/`, comparing each
line's width against the width declared in its filename. 31 frames carry at least one
over-width line; 21 of those are the legend class this feature fixes.

| Frame | Line | Declared | Actual | Defect |
| --- | --- | --- | --- | --- |
| `scenario__*__80x30.txt` (7 frames) | 29 | 80 | 356 | no clamp |
| `mid-game__*__80x30.txt` (7 frames) | 29 | 80 | 355 | no clamp |
| `empty__*__80x30.txt` (7 frames) | 26 | 80 | 354 | no clamp |
| `mid-game__home__112x30.txt` | 22 | 112 | 112 | cut mid-token, no ellipsis |

Diagnosis, complete to file and line:

- `internal/tui/views.go:1750` — the narrow fallback returns the grid box concatenated
  with the raw legend, outside the box, with no clip and no width set. The comment at
  `internal/tui/views.go:1500` asserts that the shared clipping helper "already clips an
  over-wide legend"; that is true of the widescreen path at `internal/tui/views.go:937`
  and false of this one. The stale assumption is the defect.
- `internal/tui/views.go:1784` — the shared line-clipping helper crops to a maximum width
  with no ellipsis, which produces the mid-token cut on the widescreen path.

Measurement caveat: the audit counted runes, while a few chrome rows carry double-width
characters. Rune counting *understates* display width, so it yields false negatives, never
false positives. All 21 legend findings are genuine.
