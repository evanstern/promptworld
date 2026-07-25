# UI Contract: jump-to-source (spec 049)

The behavioral contract for the chronicle's jump action. Design authority:
`docs/design/tui/panels/chronicle.md` (control table row: jump-to-source) and
`docs/design/tui/patterns/keymap.md` (inspect mode `⏎`; parity doctrine).
Where this contract and those pages disagree after implementation, the pages
are amended in the same PR (gate rule 1) — they are the durable form.

## §1 Key grammar (inspect mode — clock paused, chronicle visible)

| Input | Precondition | Effect |
|---|---|---|
| `⏎` | selected event locatable | center map camera on subject (pan-equivalent); actions bar showed the affordance beforehand |
| `⏎` | selected event unlocatable | no camera change; actions bar's `no location for this event` is the visible explanation |
| left-click on a rendered chronicle event row | clock paused | select that row, then apply the `⏎` rules above to it |
| left-click on chronicle while clock runs | — | no-op (FR-004) |
| left-click outside the chronicle list rows | — | no-op (nothing else is mouse-bound by this feature) |

All existing keys (`j/k/g/G/J/K`, `space`, globals) are byte-identical in
behavior before and after this feature (FR-005).

## §2 Subject resolution (FR-002)

1. **Primary actor**: the event type's single distinguished actor (speaker /
   initiator / affected agent — the same field the digest grammar treats as
   the line's subject). If that agent index resolves in the live replica and
   the agent is alive → its current position.
2. **Payload position**: else, the event type's explicit recorded coordinates
   (where it happened) if present.
3. **Unlocatable**: else, the honest-hint path.

Deterministic per (event, replica-state); bounded to known top-level fields.

## §3 Actions bar (detail pane bottom-right slot)

| Selected event | Bar renders |
|---|---|
| locatable | `⏎ jump to <name> (x,y)` (name/coords live-resolved) |
| unlocatable | `no location for this event` |

The bar is never empty in inspect mode after this feature; `detailActions`
never returns nil (totality — SC-002).

## §4 Camera semantics

A successful jump is exactly a computed pan: map title flips to
`MAP · panned (c to recenter)`, auto-follow suspends, `c` restores follow.
Centering clamps at map edges identically to manual panning. In the narrow
fallback, a successful jump additionally lands the player on the map pane;
selection and paused state survive the round-trip back to the chronicle.

## §5 Parity doctrine ratification (decision 8)

This feature ships the corpus's first control with a real mouse target
(keyboard and mouse landing together). `patterns/keymap.md`'s parity-rollout
statement ("zero controls have a real mouse target") is updated by this PR;
`panels/chronicle.md`'s parity-rollout note records the jump row as compliant
while its sibling controls (filters, selection, scroll) remain keyed-only and
tracked.

## §6 Non-goals

No click targets outside the chronicle list; no hover states; no
click-to-pause; no jump history; no changes to the linear `attach`/`tail`
streams (D1 — the payloads already carry positions).
