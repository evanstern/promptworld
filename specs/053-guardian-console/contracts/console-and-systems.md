# UI Contract: guardian console + systems tab (spec 053)

Design authority: `docs/design/tui/pages/guardian-console.md`,
`panels/systems.md`, `panels/guardian.md`, `panels/dock.md` — amended to
shipped/current in the same PR (gate rules 1–3).

## §1 Navigation grammar

| Input | From | Effect |
|---|---|---|
| `G` | home / solo / narrow (minibuffer unfocused) | open the console full-screen, snapshotting the prior view |
| `G` | console | close, restore prior view |
| `1` / `esc` (nothing focused) | console | close, restore prior view |
| `m` | console | focus the standard minibuffer (composer) |
| `e` | console (minibuffer unfocused) | $EDITOR shell-out on charter.md |
| `J`/`K` | console | scrollback down/up (tail-anchored; reset on close) |
| `5` | global | select systems dock tab; again → solo zoom; again → home (existing tab grammar) |

While the minibuffer is focused, `G`/`e`/`5`/`J`/`K` type into the buffer
(focus-contract rule: no silent stealing). Parity: `G`/`e` keyboard-only at
birth, recorded parity gaps.

## §2 Console composition (top → bottom)

1. Header line (guardian pane-header data, page-width).
2. Document-style turn stream (tail-anchored): labeled blocks
   `you · HH:MM` / `<epithet> · HH:MM`; special rows ⚡/👁/⏲/» inline in
   stream order; timestamps omitted when the entry carries none.
3. Card seam: `[]consoleCard` composed here — EMPTY this feature
   (producers: TASK-127/115). Seam symbol documented on the design page.
4. Charter/skills read surface (bordered): charter provenance
   (default / player-authored / preset-locked) + binding status; skills
   count + binding status; lock notices name the unlocking stage.
   `[e] edit ($EDITOR)` hint on the charter row.
5. One-shot notice line (post-$EDITOR), when set.
6. Composer: the standard minibuffer, unchanged states/contract/transport.
7. Footer: `G back · esc back · m ask · space pause · q quit · ? help`.

## §3 The split inventory (D10)

MOVES to systems tab (key `5`): provider table rows + health continuation
lines + `(unattributed)` row + `spend $X of $Y` wallet line + budget notice;
`🜂 cognition horizon` block (header, class rows, remedies, skipped counts).
STAYS on guardian tab: pane header (name/charges/stage), transcript rows,
standing-orders block, instruction/capability provenance lines, unseen
badge. The systems tab carries zero skin tokens; the guardian tab's fiction
strings resolve per TASK-121's contract when merged (rebase note).

## §4 $EDITOR handoff

Suspend TUI → run `$EDITOR <world>/charter.md` → restore. Changed content
(hash) → exactly one "charter changed — next turn binds it". Unchanged →
nothing. Exit error / `$EDITOR` unset → one honest notice. Never blocks or
pauses the world; never partially applies (the daemon's per-turn re-read is
the only binding mechanism).

## §5 Non-goals

No in-TUI editor; no skills-file picker (direct-file editing remains the
skills authoring path this wave); no card content; no new IPC fields; no
change to the dock solo-zoom machine for tabs 2–4; no second badge system.
