# Feature Specification: Chronicle Raw Feed Wrapping

**Feature Branch**: `task-195-polish-session-1`

**Created**: 2026-08-03

**Status**: Draft

**Board Task**: TASK-195

**Input**: User description: "The raw feed view does not word-wrap when the line is too long.
Particularly hard with 'thoughts' or conversations. I'd like the text to wrap and line up with
the 4th column of text (the name of the villager and so on)."

## Overview

The chronicle's raw feed is the player's live window on everything the world emits. It renders
as a four-column table — tick, time of day, event type, and a summary that begins with the
villager's name:

```
 390295 18:24  agent.saw             Hazel saw tree at (23,34)
 390458 18:27  agent.moved           Oak → (30,40)
```

For short mechanical events this reads well. For the events players actually care about —
a villager's thought, a line of conversation — the summary is prose, and prose is long. Today
that prose is **cut off at the right edge with an ellipsis**. The player sees the beginning of
a thought and never the end of it. The most interesting content in the game is the content the
feed is least able to show.

Two independent causes produce this, and both must be addressed:

- **Wrapping is disabled at every width a player actually reads at.** The chronicle body is
  rendered with a maximum-wrap budget of one line, raised to three only when the pane is
  narrower than 60 columns. In the full-width solo view — the view a player opens *to read* —
  the budget is one, which means truncate.
- **Where wrapping does happen, it has no hanging indent.** The wrap operates on the whole
  flattened line, prefix and summary together, so continuation lines begin at column zero.
  They collide with the tick column and the table stops reading as a table.

The requested behavior fixes both: long summaries wrap, and every continuation line begins at
the summary column, so the first three columns stay a clean left rail and the prose forms a
single aligned block:

```
 394500 19:45  social.conversation_turn  Hazel: "I keep thinking about the chest by the river,
                                         and whether Oak actually saw who took it — nobody
                                         will say it out loud."
 394663 19:48  agent.moved               Oak → (30,40)
```

The alignment column is **not a constant**. The tick and type column widths are recomputed for
each visible window, so the indent is whatever the first three columns currently measure. A
hardcoded indent would misalign the moment a wider tick or a longer event type scrolls into
view.

This feature also has a **testability precondition**. None of the three frame fixtures emits a
long conversation or thought, so no committed frame reproduces the defect and no before/after
frame diff can be produced. Since the frame diff is this repo's review artifact for UI change,
a fixture must carry a long-prose event before the change is reviewable.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A thought can be read to its end (Priority: P1)

A player watches the feed at their normal terminal size and a villager thinks something long.
Today the thought is cut mid-sentence with an ellipsis and the rest is unrecoverable — the feed
is the only place it is shown. After this change the whole thought is present, continuing onto
as many lines as it needs.

**Why this priority**: This is the entire complaint. Prose events are the ones players read the
feed *for*, and they are precisely the ones currently unreadable. Shipping only this story
already delivers the value, even with continuation lines starting at column zero.

**Independent Test**: Render a fixture carrying a long conversation or thought at a full-width
size and confirm the complete summary text appears across multiple lines with no ellipsis.
Requires no part of Story 2.

**Acceptance Scenarios**:

1. **Given** an event whose summary is longer than the available width, **When** the raw feed
   renders it at a full-width size, **Then** the entire summary text is present in the frame
   and no ellipsis marker appears.
2. **Given** an event whose summary fits the available width, **When** the raw feed renders it,
   **Then** it occupies exactly one line, unchanged from today.
3. **Given** a long summary, **When** it wraps, **Then** it breaks between words and never
   mid-word.

---

### User Story 2 - The feed still reads as a table (Priority: P2)

A player scanning the feed relies on the left rail — tick, time, type — to find things. When a
long event wraps, its continuation lines must not intrude on that rail. Every continuation line
begins exactly where the summary column begins, so the prose forms one aligned block and the
rail stays unbroken.

**Why this priority**: Without it, Story 1 trades one readability problem for another: wrapped
text starting at column zero is hard to tell from a new event, and the scannability of the
first three columns is lost. It is the difference between wrapping and wrapping *well*.

**Independent Test**: Render a fixture with a wrapped event and confirm every continuation line
begins at the same column as its own summary's first character, and that the columns to the
left of that point are blank on continuation lines.

**Acceptance Scenarios**:

1. **Given** an event that wraps to more than one line, **When** it renders, **Then** every
   line after the first begins at the same column as the summary column of the first line.
2. **Given** a visible window whose widest tick or longest event type changes, **When** the
   feed re-renders, **Then** the continuation indent moves with the recomputed column widths
   and still aligns to the summary column.
3. **Given** a wrapped event, **When** it renders, **Then** no continuation line contains any
   tick, time, or type text.
4. **Given** consecutive wrapped events, **When** they render, **Then** a reader can tell where
   one event ends and the next begins.

---

### User Story 3 - Narrow panes degrade sensibly (Priority: P3)

A player using the chronicle in the narrow dock still gets readable prose. The indent must not
be applied so aggressively that the remaining text column collapses into a sliver, and the
feed must never produce a negative or zero-width text area.

**Why this priority**: It protects Story 2 from becoming a defect at small sizes. It affects a
narrower set of situations than either story above, but without it the change is a regression
at the widths where the pane is tightest.

**Independent Test**: Render a wrapped event at the narrowest supported size and confirm the
text column retains a usable minimum width and no line exceeds the pane.

**Acceptance Scenarios**:

1. **Given** a pane so narrow that indenting to the summary column would leave less than a
   usable minimum of text width, **When** a long event renders, **Then** the layout falls back
   to a reduced or zero indent rather than producing a sliver column.
2. **Given** any supported pane width, **When** a long event renders, **Then** no emitted line
   exceeds the pane width.
3. **Given** any supported pane width, **When** a long event renders, **Then** the text area is
   never zero or negative width, and rendering never fails.

---

### Edge Cases

- **A single word longer than the text column** (a long path, a URL, an unbroken token): must
  render without an infinite loop and without silently vanishing.
- **Empty summary**: an event with no summary text must render its left rail and nothing else,
  exactly as today.
- **Styled prose**: feed text carries per-segment styling by role and family. Wrapping must
  break the plain text, never a style escape sequence, or color bleeds across the screen.
- **Whole-line event tiers**: alerts and labeled-voice families render as one uniformly styled
  line rather than prefix-plus-summary. Their wrap behavior must remain coherent and must not
  lose the whole-line styling.
- **The row budget**: the feed shows the tail of the event ring within a fixed row budget. Once
  events can occupy several rows each, the budget must still be honored exactly — the body must
  not overrun its allotted height, and the newest event must remain visible.
- **Selection**: a selected row is rendered in reverse. A wrapped selected event must present
  as one coherent selection, not a highlighted first line with unhighlighted remainder.
- **Double-width glyphs** at a wrap boundary must not be split, and must not make a line
  measure wider than the pane.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The raw chronicle feed MUST wrap summary text that exceeds the available width
  rather than truncating it, at every pane width where the feed is displayed.
- **FR-002**: Wrapping MUST break between words. A word MUST NOT be split except when it alone
  exceeds the text column width.
- **FR-003**: Every continuation line of a wrapped event MUST begin at the column where that
  event's summary begins on its first line.
- **FR-004**: The continuation indent MUST be derived from the currently computed column widths
  of the visible window, never from a fixed constant.
- **FR-005**: Continuation lines MUST NOT contain tick, time, or event-type content.
- **FR-006**: When the indent would leave less than a defined minimum text width, the feed MUST
  reduce or drop the indent rather than render a sliver text column.
- **FR-007**: No emitted feed line MAY exceed the width of the pane it renders into, at any
  supported size.
- **FR-008**: The feed MUST continue to honor its row budget exactly once events can occupy
  multiple rows, and the most recent event MUST remain visible.
- **FR-009**: Events whose summary fits on one line MUST render with their column padding
  intact — the left rail is never reflowed. In the full-width views this means
  byte-identical to today. **In the narrow dock it does not**, and deliberately so: the dock
  already wrapped, and its wrap path collapsed every run of whitespace, so its column padding
  was being destroyed on rows that never needed wrapping at all. Repairing that is a
  consequence of this feature, not a regression, and it changes committed dock frames. See the
  "Discovered during implementation" note below.
- **FR-010**: Wrapping MUST preserve each character's styling role, so that wrapped prose is
  colored exactly as unwrapped prose is.
- **FR-011**: A selected wrapped event MUST render as one coherent selection across all of its
  lines.
- **FR-012**: At least one frame fixture MUST emit an event whose summary is long enough to
  wrap at the committed frame sizes, so the behavior is visible in the committed frame matrix.
- **FR-013**: The design authority for the chronicle panel and the digest-grammar line-format
  contract MUST be amended to describe the wrapping and indent behavior, in the same change.

### Key Entities

- **Feed line**: one event rendered for display — a left rail (tick, time, type) plus a summary
  of styled segments. The unit that may now occupy multiple physical rows.
- **Column layout**: the per-window measurement of tick and type column widths, from which the
  summary column position — and therefore the continuation indent — is derived.
- **Row budget**: the number of physical rows the feed body may occupy, which multi-row events
  must now be reconciled against.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A player can read any villager thought or conversation line in full from the
  feed, without resizing the terminal, opening another view, or consulting a file.
- **SC-002**: Zero prose events are shown with their text cut off at the pane edge, at any
  supported size.
- **SC-003**: For every wrapped event in the committed frame matrix, all continuation lines
  start at the same column as that event's summary.
- **SC-004**: No line in any committed frame exceeds the frame's width.
- **SC-005**: The feed body occupies no more than its allotted rows at every committed size,
  with the newest event visible.
- **SC-006**: Events short enough to fit on one line render exactly as before, so the change is
  invisible for the mechanical events that make up most of the feed.
- **SC-007**: The committed frame matrix contains at least one visibly wrapped, indented event,
  so the behavior can be reviewed from frames alone.

## Discovered during implementation

**The narrow dock's column alignment was already broken.** The dock has wrapped since before
this feature, and its wrap path budgeted with a routine that collapses runs of whitespace —
which is exactly what the feed's column padding is made of. Every dock row, including short
ones that never needed wrapping, was therefore reflowed from

```
19:12 moved       Fern → (36,26)
```

to

```
19:12 moved Fern → (36,26)
```

losing the type column entirely. Nobody had noticed, because the committed frames recorded the
collapsed form as if it were intended.

The fix this feature needed anyway — return a line that fits verbatim, and wrap only the
summary — repairs it. Committed `mid-game` and `scenario` dock frames change as a result,
including frames for a fixture this feature never touched. That is the repair showing up, not
drift, and FR-009 above was amended to say so rather than claim a byte-identity the diff
contradicts.

## Assumptions

- **Wrap depth in the full-width view is unbounded.** A long thought wraps to as many lines as
  it needs rather than being capped and re-truncated. Capping would reproduce the original
  complaint in a subtler form — the player would still lose the end of the sentence. The row
  budget, which already drops the oldest events, is the mechanism that keeps the feed bounded.
  The narrow dock's existing cap is treated as a separate, retained behavior.
- **The minimum text width below which the indent yields** is a layout constant chosen during
  planning; this spec requires only that one exist and be honored.
- **Auto-follow behavior is unchanged.** The feed continues to show the tail of the ring; this
  feature changes how a line is laid out, not which events are selected for display.
- **The narrated chronicle view is out of scope.** It already wraps prose to the pane width and
  is not what the request refers to.
- **Only the raw feed's own rendering is in scope.** Event payloads, the digest grammar's
  choice of summary text, and the event ring are untouched.
