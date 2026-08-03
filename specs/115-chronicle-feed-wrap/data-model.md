# Data Model: Chronicle Raw Feed Wrapping

**Phase 1** · spec 115

This feature introduces **no persisted state, no event payload change, and no new type**. It
changes how three existing in-memory values are computed and threaded. They are documented here
because the correctness of the feature is entirely a matter of which of them is authoritative.

---

## Existing entities (behavior changed, shape unchanged)

### `chronicleLine`

One event prepared for display: sequence, tick, time-of-day string, event type, family, and a
`Summary` of styled segments.

- **Unchanged.** The wrap happens downstream of this value.
- Relationship: one `chronicleLine` now maps to **one or more physical output lines**, where
  before it always mapped to exactly one. This is the single conceptual change in the feature.

### `chronicleColumns`

The per-visible-window column measurement: `Dock`, `TickWidth`, `TypeWidth`.

- **Unchanged in shape.** Gains a second consumer: the continuation indent is derived from it via
  the prefix, alongside its existing role of padding the columns.
- **Validation rule:** it is recomputed from exactly the rows about to render, every frame. Any
  code path that caches it across frames, or derives an indent from anything else, violates
  FR-004.

### Wrap budget (`maxWrap`)

An `int` parameter threaded from the view layer into both wrap renderers.

- **Domain extended** from `{1, >1}` to `{0, 1, >1}`; `0` means unbounded. See the contract §2.
- **Validation rule:** the value must survive from call site to renderer without upward
  normalization. The current `if maxWrap < 1 { maxWrap = 1 }` clamps must be removed, or `0`
  silently becomes `1` and the feature does nothing.

---

## New derived values (computed, never stored)

### Continuation indent

- **Definition:** the rune width of the row's own prefix.
- **Lifetime:** one row, one frame. Never stored, never cached, never shared.
- **Validation rule:** when `width - indent < minWrapTextWidth`, the indent is `0`, not a reduced
  value.

### `minWrapTextWidth`

- A layout constant, value `24`.
- The minimum text column width below which the indent yields entirely.

---

## State transitions

None. The feed is a pure projection of the event ring onto a character grid; this feature changes
the projection, not any state machine.

---

## Invariants the implementation must preserve

1. A row whose summary fits on one line renders byte-identically to its pre-115 output.
2. Every physical line of a row is at most the pane width.
3. The text column width is never zero or negative.
4. The styled path and the plain path produce the same characters for the same input.
5. Column widths, and therefore the indent, are recomputed every frame.
