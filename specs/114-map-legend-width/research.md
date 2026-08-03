# Phase 0 Research: Map Legend Width Policy

**Spec**: [spec.md](./spec.md) | **Board**: TASK-191 | **Date**: 2026-08-02

Three questions had to be resolved before the plan could commit to an approach. All three
are now closed; no NEEDS CLARIFICATION remains.

---

## R1 — Which truncation primitive?

**Decision**: `ansi.Truncate(s string, length int, tail string)` from
`github.com/charmbracelet/x/ansi`, promoted from an indirect to a direct dependency.

**Rationale**: the spec's FR-003, FR-004, FR-006 and FR-007 together demand a truncator
that is simultaneously (a) ANSI-escape-safe, (b) display-column-aware rather than
rune-aware, (c) able to append a tail, and (d) a no-op when the string already fits.
`ansi.Truncate` is documented as satisfying exactly this set:

> Truncate truncates a string to a given length, adding a tail to the end if the string is
> longer than the given length. This function is aware of ANSI escape codes and will not
> break them, and accounts for wide-characters (such as East-Asian characters and emojis).

and its implementation short-circuits `if sw := StringWidth(s); sw <= length { return s }`,
which is FR-004 for free. The module is already in the build graph at v0.10.1 as an
indirect dependency of lipgloss, so this adds no new module to `go.sum` — only a `go.mod`
promotion.

**Alternatives considered**:

- **`truncateRunes` (`internal/tui/digest.go:94`)** — the package's existing `…`
  truncator. Rejected: it counts runes, so it fails FR-007 on double-width glyphs, and it
  slices the string directly, so it fails FR-006 by severing ANSI escapes. The legend is
  wrapped in `styleDim.Render`, so it always carries escapes. This helper is correct for
  its own callers (plain contract fields) and is left alone.
- **`lipgloss.NewStyle().MaxWidth(n)`** — what `clipLine` uses today. ANSI-safe and
  width-aware, but has no tail parameter, which is the whole defect in FR-003. Retained
  for `clipLine`'s own purpose (see R2).
- **Hand-rolled grapheme walk** — rejected as re-implementing a maintained upstream
  primitive that the project already depends on.

---

## R2 — Global `clipLine` change, or a legend-local helper?

**Decision**: **legend-local.** Add a small map-panel helper that clamps the legend with an
ellipsis. Leave `clipLine` (`internal/tui/views.go:1777`) and `clipContent`
(`internal/tui/views.go:1800`) exactly as they are.

**Rationale** — four independent reasons, any two of which would settle it:

1. **They are different jobs.** `clipLine`'s own documentation calls it a crop, and
   `clipContent`'s doc block (`views.go:1787-1799`) describes it as a defense against a
   specific lipgloss failure mode: `Height()` pads but never truncates, so one overlong
   line silently grows a panel past its row budget. It is a **layout safety net** — a
   last-resort guarantee that a panel renders to exactly its handed size. Signalling
   "content was omitted" to the player is a **content policy**. Folding the second into
   the first means every future panel silently inherits a content decision it never asked
   for.

2. **Double-ellipsis hazard.** The chronicle already truncates with its own `…`
   (`internal/tui/grammar.go:284` and `:554`), and its output then passes through
   `clipContent`. If `clipLine` also appended `…`, any chronicle line whose
   grammar-level truncation still exceeds the panel's usable width would be re-truncated
   and re-tailed. The observable frames already show grammar's `…` surviving into the
   dock (`mid-game__home__112x30.txt:17`, `19:51 neglect_d… Birch is dangerously cold and`),
   so the two layers demonstrably compose today. Adding a tail to the lower layer puts
   that composition at risk for no gain.

3. **Blast radius.** `clipLine` is reached from roughly 18 call sites across
   `views.go`, `help.go`, `exercise.go`, and `reportcard.go`. Several are structural rows
   (box interiors, padded rows, the minibuffer) where an ellipsis is simply wrong. A global
   change would rewrite a large fraction of the 132 committed frames, turning a targeted
   two-line defect fix into a matrix-wide golden churn — and burying the actual behavior
   change in the diff, which defeats the project's own "the frame diff *is* the review
   artifact" rule.

4. **Precedent.** Per-surface ellipsis is the established pattern in this package: the
   chronicle does it (`grammar.go`), the villager strip does it (the `…N` overflow
   marker), the digest does it (`truncateRunes`). A legend-local helper is the idiomatic
   choice, not the exceptional one.

**What this costs**: surfaces other than the legend keep truncating silently. That is
pre-existing behavior, is not in this spec's scope, and is now *detectable* — the FR-009
guard makes any over-width line visible even where no ellipsis is added.

**Alternative considered**: extend `clipLine` with a variadic tail parameter defaulting to
empty, so existing callers are unchanged and only the legend opts in. Rejected as strictly
worse than a named helper: same call-site count, same review surface, but it hides a
content policy inside a layout primitive's signature and invites future callers to reach
for it casually.

---

## R3 — Where does the narrow path get its width budget?

**Decision**: clamp at each of the two call sites, using the width budget already in scope
there. `renderMapGrid` keeps returning the legend unclamped.

| Path | Function | Budget | Why |
| --- | --- | --- | --- |
| Narrow (<112 cols) | `mapView` (`views.go:1737`) | `m.width` | The legend renders **outside** the map box, so the full terminal width is genuinely available to it. `m.width` is already in scope and already drives the viewport calculation two lines above. |
| Widescreen (112+) | `mapPanelView` (`views.go:907`) | `cols-4` | The legend renders **inside** a `styleBox` whose `.Width(cols-2)` is then reduced by `Padding(0,1)`. `clipContent` computes the same `boxWidth-2`, so `cols-4` is the true usable width. |

**Rationale**: `renderMapGrid` receives viewport dimensions in *tiles* (`vw, vh`), not
panel columns, so it cannot compute either budget. Pushing the clamp to the callers puts
it where the width is known and keeps one legend-composition site with two presentation
budgets — the same "one renderer, two widths" shape the chronicle already uses
(`views.go:1810-1811`: "One body renderer shared by the narrow pane, the dock tab, and the
solo view — differing only in (width, height, maxWrap)").

**Layering consequence**: because the legend now arrives already within budget,
`clipContent` on the widescreen path becomes a no-op for that line rather than the thing
doing the cutting. The safety net stays in place and stays untouched; it simply stops
being load-bearing for the legend. The stale comment at `views.go:1500` asserting that
`clipContent` "already clips an over-wide legend" must be corrected in the same change —
it is the false belief that allowed the narrow path to ship unclamped.

**Alternative considered for the narrow budget**: clamp to the map box's rendered width
(77 columns at an 80-column terminal) rather than `m.width` (80), so the legend aligns
flush with the box border. Rejected: it discards three columns of information for a
cosmetic gain, and FR-005 favors showing more. Recorded here because it is a reasonable
taste call an operator might want to revisit — it is a one-constant change if so.

---

## R4 — Where does the regression guard live?

**Decision**: a new Go test in `cmd/promptworld/frames_test.go`, alongside the existing
matrix tests.

**Rationale**: that file already owns the frame matrix contract —
`TestDumpFramesWritesEveryCombination`, `TestDumpFramesMatchesCommittedMatrix`,
`TestDumpFramesPrunesStaleFrames`, and `TestFrameSizesStraddleTheBoundaries` all live
there, and `frameSizes` (`frames.go:73`) is the declared-width source the guard needs. A
test in this file runs under plain `go test ./...` with no new tooling, no new script, and
no CI wiring — which matches the project's stated preference for gates that are exit codes
rather than daemons.

**Alternative considered**: extend `.claude/skills/tui-frames/scripts/check-frames.mjs`.
Rejected as the primary home — that script checks *freshness* (committed matrix vs. a fresh
dump) and is invoked by a skill, not by the build. A width invariant belongs in the test
suite so it fails for anyone running tests, not only for someone who thought to run the
skill's checker.

**The scope problem this exposes, and how the plan handles it**: a guard honest enough to
satisfy FR-010 (cover the whole matrix, not just the legend) fails immediately on two
defect classes this feature does not fix — the header row (20 frames, 81–83 columns at
80x30) and the scenario keybind footer (2 frames, 121 columns at 112/113). Three options:

- Scope the guard to the legend line only — rejected, violates FR-010 and leaves the
  matrix unguarded.
- Fix all three classes here — rejected, this is the option the operator explicitly did
  **not** choose when scoping the work, and the header and footer have unrelated causes.
- **Chosen**: the guard covers every line of every frame, with a small explicit
  registry of known-failing frames, each entry naming the follow-up card. The registry is
  a deny-list that can only shrink: a frame not in it must pass, and a frame in it that
  *starts* passing fails the test too, so the debt cannot silently persist after someone
  fixes it. Follow-up cards for the header and footer classes are filed as part of this
  work.

---

## Constitution alignment

- **Principle I (Artifact-Grounded Action)** — every decision above cites the file and
  line it was derived from; the evidence table in `spec.md` is reproducible by re-running
  the frame audit.
- **Principle III (Gates Over Assertions)** — R4 converts "we fixed the legend" from a
  claim into a test that fails when it stops being true.
- **Principle IV (Grounding Freshness)** — `docs/design/tui/panels/map.md` is a pinned
  wiki-adjacent design authority and is amended in-branch (FR-011), per the wiki-in-PR
  lifecycle.
- **Principle V (Model-Tiered Workflow)** — single package (`internal/tui`) plus a test in
  `cmd/promptworld`, view/rendering code, tests alongside code. This is the routine tier:
  `.claude/agents/spec-implementer.md` (claude-sonnet-5). No escalation trigger fires — no
  cross-package architecture change, no concurrency or governor logic, no doctrine-adjacent
  behavior change, no prior failed attempt.
