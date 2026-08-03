# Contract: Chronicle raw feed wrap and indent

**Spec 115** · amends the digest-grammar contract's "Line format" section and
`docs/design/tui/panels/chronicle.md`.

This contract is the normative statement of what a raw-feed row looks like once its summary
exceeds the available width. It is written against the rendered characters — the thing a frame
file contains — not against the functions that produce them.

---

## §1 Row shape (unchanged for short rows)

A raw-feed row is a **prefix** followed by a **summary**.

```
solo:  <tick> <HH:MM>  <type>  <summary>
dock:          <HH:MM> <type>  <summary>
```

The tick is right-aligned to the widest tick in the visible window; the type is left-aligned and
padded to the widest type in the visible window, capped at 26 runes (solo) or 10 (dock). Both
widths are recomputed for every frame from exactly the rows about to render.

A row whose summary fits within the pane renders on one line, **verbatim** — the prefix's column
padding is never reflowed. In the full-width views this is byte-identical to the pre-115
rendering. In the narrow dock it is not: the dock's pre-115 wrap path collapsed every run of
whitespace, so its column padding was being destroyed on rows that never needed wrapping. The
padded form is correct and is what this contract requires.

## §2 Wrap budget

The wrap budget is an integer with three domains:

| value | meaning |
| --- | --- |
| `0` | **unbounded** — the summary wraps to as many lines as it needs |
| `1` | truncate to one line, ending in `…` when content was dropped |
| `> 1` | wrap, capped at that many lines, `…` on the last line when content was dropped |

Budgets by surface:

| surface | budget |
| --- | --- |
| solo (full-width chronicle) | `0` |
| dock, pane width < 60 | `3` |
| dock, pane width ≥ 60 | `0` |
| narrow fallback chronicle | `0` |

## §3 Continuation indent

When a row occupies more than one physical line:

1. Every line after the first begins with exactly `indent` spaces, where `indent` is the rune
   width of that row's own prefix.
2. No continuation line contains tick, time, or type text.
3. `indent` is derived from the current frame's column widths. It is never a stored constant and
   never carried between frames.

**Narrow-pane fallback.** If `width - indent < 24`, the row wraps at full width with `indent = 0`.
The fallback is all-or-nothing: a reduced indent is never emitted, because an indent that is
neither the summary column nor the margin aligns to nothing.

## §4 Wrapping rules

1. Wrapping breaks between words. A word is split only when that single word exceeds the text
   column width on its own.
2. The text column width is `width - indent` for continuation lines and `width` for the first
   line (which carries the prefix).
3. No emitted line exceeds the pane width, at any size.
4. The text column width is never zero or negative; rendering never fails on a narrow pane.
5. Wrapping operates on plain text. A style escape sequence is never split.
6. Each character keeps the style role it had before wrapping, so wrapped prose is colored
   exactly as unwrapped prose.

## §5 Tiers

The alert tier and the labeled-voice families render as one uniformly styled line rather than
prefix-plus-summary. They obey §3 and §4 identically — same indent, same wrap — and keep their
whole-line styling across every physical line.

## §6 Selection

A selected row renders in reverse across **all** of its physical lines. A wrapped selected row is
one visually coherent selection, never a highlighted first line above unhighlighted remainder.

## §7 Row budget

The feed body occupies no more than its allotted rows. Once rows may occupy several lines each,
the budget is enforced on **physical lines**, trimmed from the top, so the newest row is always
visible at the bottom.

**Accepted consequence.** Trimming from the top can leave an orphaned continuation line — the tail
of a row whose first line was trimmed — at the very top of the feed. This is the top edge of a
scrolling window and is correct, not a defect.

## §8 Evidence

At least one committed frame fixture emits a row long enough to wrap at 160 columns, so §3 and §4
are demonstrable from the committed frame matrix without running the client.
