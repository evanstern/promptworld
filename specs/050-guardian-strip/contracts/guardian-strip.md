# UI Contract: guardian strip (spec 050)

Behavior contract for the always-visible action-budget row. Design authority:
`docs/design/tui/panels/guardian-strip.md` (amended to `shipped` in the same
PR, carrying rulings §2 below) and `docs/design/tui/patterns/layout.md`
rulings a/b.

## §1 Placement

Exactly one borderless row directly above the minibuffer: widescreen
composite (every dock tab, every solo view retains chrome per existing
composite rules) and narrow fallback alike. Never 2 rows; overflow truncates
with `…`.

## §2 Segment grammar (left → right, ` · ` separated)

| Segment | Form | Presence rule |
|---|---|---|
| charge bank | `⚡⚡· (2/3)` — filled glyphs, empty `·` padding, numeric | status snapshot exists |
| regen | `next +1 @ 12:00` | status exists AND bank below cap (full bank: omitted — nothing is scheduled) |
| standing orders | `👁 2 standing orders` | status exists (0 renders as a true zero) |
| faith | — | omitted entirely until the faith mechanic ships (never a dash before then) |

Pre-status (connecting): the row is present and blank — layout stays stable,
no invented values. Width pressure truncates segments right-to-left (faith →
orders → regen → bank): the bank is the last thing standing.

## §3 Fold behavior (layout.md ruling a step 4)

The strip is the LAST chrome row to fold. Folded, its content relocates into
the minibuffer's DORMANT line as a prefix
(`⚡⚡· 👁2 · ⏎ m — speak with the …`), truncated to width; the focused and
busy minibuffer states are byte-identical to today. Unfolding restores the
row and the plain dormant line.

## §4 Honesty invariants

1. Every rendered segment is a true statement about the current world state
   (SC-002's fixture sweep is the mechanical form of this).
2. The strip and the guardian tab header can never disagree — same fields,
   same frame.
3. No new fiction literals: all strip text is non-fiction chrome (skin-token
   column `—`); the dormant placeholder's epithet belongs to the minibuffer
   page, unchanged here.

## §5 Non-goals

No keys, no mouse targets (display-only — no parity gap); no stage-defaults
machinery (TASK-128); no faith rendering (TASK-118); no IPC/status wire
changes.
