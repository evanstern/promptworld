# Contract: Legend Width Invariant

**Spec**: [../spec.md](../spec.md) | **Plan**: [../plan.md](../plan.md)

The UI contract this feature establishes. Stated so it can be tested rather than asserted
(constitution Principle III).

## C1 — Nothing exceeds its terminal

> For every fixture, every state, and every size in the committed frame matrix, every line
> of the rendered frame measures at most that size's declared width, in **display
> columns**.

- Measured with a display-width function, never `len()` and never rune count. Rune count
  understates width wherever double-width glyphs appear (the guardian strip's `⚡` and
  `👁`), which is precisely how the original audit could produce false negatives.
- Declared width comes from `frameSizes` (`cmd/promptworld/frames.go:73`), not from parsing
  filenames, so the guard and the dumper can never disagree about what a size means.
- Enforced by the FR-009 guard in `cmd/promptworld/frames_test.go`.
- **Known exceptions** are enumerated in the guard's deny-list, each naming the follow-up
  card that retires it. The deny-list is bidirectional: an unlisted frame that fails is a
  failure, and a listed frame that *passes* is also a failure, so a fixed entry cannot be
  left behind.

## C2 — Truncation is signalled

> A legend that had content removed for width ends in `…`. A legend that fits carries no
> `…` and is not padded.

Both halves are load-bearing. The second prevents a "just always append an ellipsis"
implementation from passing, which would lie in the opposite direction — telling the player
content was hidden when nothing was.

## C3 — Monotonic in width

> For any fixture and state, and any two widths `w1 < w2`, the legend rendered at `w2`
> displays at least as much content as at `w1`.

This forbids a fixed cap that happens to satisfy C1 and C2 while ignoring available space.
The player's mental model — "widen the window, read more" — is the actual deliverable.

## C4 — One row, always

> The legend occupies exactly one rendered row on both the narrow and widescreen paths.

Wrapping to a second row is not an accepted remedy for overflow (FR-008): it would trade a
horizontal violation for a vertical one, and the map panel's row budget is already the
scarcer resource. `docs/design/tui/panels/map.md:130` states the legend is shed *entirely*
before the viewport shrinks — it never grows into a second row.

## C5 — Styling does not leak

> Truncation never severs an ANSI escape sequence.

A cut mid-escape leaves the terminal in whatever style was active, bleeding color into
every row below. This is why rune-slicing helpers are unusable on the legend regardless of
how convenient they are.

## Non-contracts

Explicitly *not* promised, so a reviewer does not read them in:

- **Which segments survive.** Tail truncation means the composition order in
  [../data-model.md](../data-model.md) decides. Segment-priority shedding is out of scope.
- **Ellipsis behavior on other surfaces.** Other truncating surfaces keep their existing
  per-surface behavior; `clipLine` and `clipContent` are unchanged (research.md R2). C2
  binds the legend only.
- **Any minimum amount of legend content.** At a sufficiently narrow terminal the legend
  may reduce to almost nothing. C1 and C4 still hold; there is no floor on information.
