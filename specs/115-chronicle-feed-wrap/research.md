# Research: Chronicle Raw Feed Wrapping

**Phase 0** · spec 115 · 2026-08-03

Every unknown below was resolved by reading the current renderer, not by assumption. File
references are to the branch `task-195-polish-session-1` at the time of writing.

---

## R1 — Where the wrap decision actually lives

**Finding.** Two call sites choose the wrap budget, and both choose "truncate" for the widths a
player reads at:

- `internal/tui/views.go:1073-1076` (`dockTabContent`) — `maxWrap := 1`, raised to `3` only when
  `width < 60`. The solo view and any dock ≥ 60 columns therefore get `1`.
- `internal/tui/views.go:2286` (`chronicleView`, the narrow fallback) — passes `1` unconditionally.

`maxWrap` threads down to two renderers that must agree:

- `styleWrapLine` (`internal/tui/grammar.go:424`) — the styled path, used for ordinary rows.
- `wrapOrTruncatePlain` (`internal/tui/grammar.go:542`) — the plain path, used by `styleWholeLine`
  for the alert and labeled-voice tiers.

Both take `maxWrap <= 1` to mean "truncate with `…`". The two are deliberately kept
byte-equivalent (the existing plain-equivalence tests referenced in `styleWrapLine`'s own comment).

**Decision.** Thread the change through **both** paths with a shared parameter rather than
patching only the styled path. Any divergence would show up as alert rows aligning differently
from ordinary rows in the same frame.

**Alternatives considered.** Patch `styleWrapLine` only — rejected: alert and labeled-voice rows
would keep truncating, so the fix would appear to work until a death or a norm violation scrolled
past.

---

## R2 — Expressing "wrap as far as it needs"

**Finding.** `maxWrap` is an `int` whose current domain is `1` (truncate) and `> 1` (cap at N
lines, ellipsis on the last). There is no value meaning "unbounded", and both renderers normalize
out-of-domain values upward (`if maxWrap < 1 { maxWrap = 1 }`).

**Decision.** Extend the domain with **`0` = unbounded**, leaving `1` and `> 1` exactly as they
are. The existing `maxWrap <= 1` truncate guards become `maxWrap == 1`, and the upward
normalizations are removed so `0` survives to the wrap branch.

Call sites become: solo and the narrow fallback pass `0`; the narrow dock keeps `3`.

**Rationale.** Preserves every existing behavior by value, so the dock's capped wrap and the
one-line truncate path are untouched and their tests keep passing unmodified. The alternative — a
`wrapPolicy` struct carrying `MaxLines` and `Indent` — reads better at the call site but rewrites
every caller and both renderers for a change that two integers express, and the constitution's
simplicity constraint disfavors the larger diff.

**Risk accepted and mitigated.** `0` meaning "unbounded" is not self-evident from the type. Both
renderers get a named constant (`wrapUnbounded`) and a comment at the declaration, and the
contract documents the domain.

---

## R3 — The alignment column

**Finding.** The prefix is built by `chronicleLinePrefix` (`internal/tui/grammar.go:299-307`):

- solo: `<tick> <HH:MM>  <type>  ` — width `TickWidth + 1 + 5 + 2 + TypeWidth + 2`
- dock: `<HH:MM> <type>  ` — the tick column is dropped entirely

`TickWidth` and `TypeWidth` come from `computeChronicleColumns`, which measures **only the rows
about to render** and is recomputed every frame. They are not stable across frames.

**Decision.** The indent is `len([]rune(prefix))`, computed per row from the prefix the row
already builds. It is never stored, never a constant, and never shared between frames.

**Alternatives considered.** A package-level constant, or measuring once per window — both
rejected: the column widths legitimately change between frames as a wider tick or a longer event
type scrolls into view, and either alternative would misalign exactly then. This is FR-004.

---

## R4 — The minimum text width below which the indent yields

**Finding.** The narrow dock can be well under 60 columns while the prefix is ~20 (dock) or ~36
(solo) columns wide. Indenting unconditionally would leave a text column of single-digit width,
which renders prose as a ragged vertical strip — worse than not wrapping.

**Decision.** A named constant `minWrapTextWidth = 24`. When `width - indent < minWrapTextWidth`,
the row wraps at full width with **zero** indent rather than a reduced one.

**Rationale for the all-or-nothing fallback.** A partial indent aligns to nothing — it is neither
the summary column nor the left margin, and it reads as a mistake. Alignment is a property that is
either exact or absent. 24 columns is roughly four to five average English words, the point below
which greedy wrapping produces more line breaks than words per line.

**Alternatives considered.** Proportional indent scaling (rejected, above); dropping wrap entirely
in narrow panes (rejected — it reintroduces the original complaint precisely where the pane is
tightest and truncation bites hardest).

---

## R5 — The row budget under multi-row events

**Finding.** `chronicleRawBody` (`internal/tui/views.go:1955-1995`) already reconciles this:

1. it slices the tail to `entryRows` **events**, on the stated assumption that each event
   contributes at least one physical line;
2. it renders them;
3. it re-splits the joined output into **physical lines** and trims from the top
   (`all = all[len(all)-entryRows:]`).

Step 3 is the existing dock-wrap overshoot trim. It already does the right thing for unbounded
wrap: the newest event stays at the bottom and the budget is honored exactly.

**Decision.** No change to the budget machinery. FR-008 is satisfied by existing code; the
implementation must verify it rather than rebuild it.

**Named consequence.** Trimming physical lines from the top can leave an **orphaned continuation
line** — the tail of an event whose first line was trimmed — at the top of the feed. This is
accepted: it is the top edge of a scrolling window, the same place a partially-scrolled line sits
in any pager. It is called out here so a reviewer does not read it as a defect.

**Cost.** Rendering up to `entryRows` events that may each produce several lines, then discarding
the overshoot, is bounded by the visible budget and unchanged in order of magnitude from today's
dock behavior.

---

## R6 — Selection across a wrapped event

**Finding.** `renderChronicleRow` builds one styled line per wrapped segment and calls
`paintStyledLine(sl, fam, selected)` for each, and `styleWholeLine` applies `.Reverse(true)` per
physical line. The selected flag is per **row**, not per line.

**Decision.** No change needed — every physical line of a wrapped selected event already receives
the reverse treatment. The implementation must add a test proving it, because nothing currently
exercises a selected row that occupies more than one line.

---

## R7 — The fixture that makes this reviewable

**Finding.** `midGameFeed` (`internal/tui/fixtures.go:323-378`) emits an eight-event ambient cycle
plus roster-state events. Every payload it produces digests to a short summary — coordinates,
names, causes. No fixture emits free prose.

The two digests that carry player-authored prose are `agent.thought`
(`internal/tui/digest.go:710`, rendering `<name> thought: "<text>" (<source>)`) and
`social.conversation_turn` (`internal/tui/digest.go:828`, rendering
`<speaker>→<listener> "<text>"`). Both wrap their text in `speech()` segments, so a fixture event
of either type exercises the styled path, the speech role, and the wrap together.

**Decision.** Extend `midGameFeed` with one `agent.thought` and one `social.conversation_turn`,
each carrying text long enough to wrap at **160 columns** — the widest committed size, and
therefore the one that binds. Place them near the tail so they fall inside the visible window at
all four committed sizes.

**Consequence to accept deliberately.** Adding events to `midGameFeed` changes tick spacing and
the visible window for **every** `mid-game` frame in the committed matrix, so a large number of
frames will re-dump with shifted content. That is expected churn, not drift — but it must be
stated in the PR so a reviewer is not asked to audit hundreds of changed lines without knowing
why.

**Alternative considered.** A fourth fixture dedicated to prose — rejected: it multiplies the
committed frame matrix by a third for one behavior, and spec 112 chose three fixtures
deliberately.

---

## Summary of decisions

| # | Decision |
|---|---|
| R1 | Thread the change through both the styled and plain wrap paths, never one alone |
| R2 | `maxWrap == 0` means unbounded; `1` and `> 1` keep today's exact meaning |
| R3 | Indent = the row's own prefix width, recomputed per frame, never a constant |
| R4 | `minWrapTextWidth = 24`; below it the indent drops to zero rather than shrinking |
| R5 | Row budget needs no change — verify the existing physical-line trim, don't rebuild it |
| R6 | Selection already spans wrapped lines — prove it with a test |
| R7 | Extend `midGameFeed` with a long thought and a long conversation turn |
