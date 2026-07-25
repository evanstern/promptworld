# Research: Guardian console + systems split (spec 053)

## R1 — The console as a page-level state, not a solo view

**Decision**: a `Model.console bool` (+ return-target snapshot of
solo/dockTab state) branching at the top of `View()` beside the
widescreen/narrow fork. `pages/solo-views.md`'s state machine is untouched —
`G` toggles the page; `1`/`esc` (nothing focused) closes it restoring the
prior view; the esc-release ordering (focus-contract rule 3) gains one layer:
minibuffer → villager detail → console → solo → home.

**Rationale**: the design page's navigation section rules the console is
"deliberately a first-class destination, not a zoomed tab"; reusing solo
would entangle two state machines the corpus keeps separate.

## R2 — $EDITOR shell-out

**Decision**: `tea.ExecProcess(exec.Command($EDITOR, charterPath), callback)`
— Bubble Tea's standard suspend/exec/restore. Before exec: stat + hash
`charter.md`; on the callback: re-stat/re-hash — changed → set a one-shot
"charter changed — next turn binds it" line on the console; editor exit
error or `$EDITOR` unset → one-shot honest notice instead. The world keeps
running (the daemon is a separate process; only the client's terminal is
handed off).

**Rationale**: decision 2 ("no in-TUI editor"); ExecProcess is the
framework-blessed mechanism, restoring altscreen state itself. Content hash
(not mtime alone) avoids false confirmations from editors that touch files
on open.

**Alternatives considered**: watching the file continuously — rejected: the
daemon already re-reads per turn; the console only owes the player the
post-edit acknowledgment.

## R3 — Systems tab wiring

**Decision**: extend the dock tab enum + `paneNames` (tui.go:48) with
`systems`, key `5`, dispatched through the existing `dockTabContent`
one-renderer-per-tab switch (dock.md: "adding a future tab = a new label in
the row + a content renderer; no new layout"). The content renderer composes
the already-shipped `llmProviderLines`, spend/wallet lines, and
`horizonLines`/`horizonRow` — moved out of the metatron/guardian tab
renderer (views.go:1436-1560 vicinity), which keeps pane header, transcript,
standing orders, and provenance lines. Solo zoom and narrow reachability
come free from the existing tab machinery.

**Rationale**: panels/systems.md is reconciliation-accurate for the content
("nothing left to build for the content itself, only the tab") — this is a
relocation, and the split IS the D10 skin boundary.

## R4 — Document-style turn rendering

**Decision**: a `consoleTurns` renderer over the same `Model.transcript`
entries the compact tab reads, emitting labeled blocks (`you · HH:MM` /
`<epithet> · HH:MM`) with blank-line separation and width wrapping; the
special-row vocabulary (⚡/👁/⏲/» lines) renders inline in stream order
unchanged. Tail-anchored with scrollback (`J`/`K`, matching the detail-pane
scroll vocabulary; reset on close). Timestamps: transcript entries that
carry no time render without the timestamp suffix rather than inventing one
(honesty rule) — the design page mockup's timestamps are representative.

**Rationale**: one shared vocabulary, two renderings (design page §2);
inventing a second transcript store would violate the replica-single-source
rule.

## R5 — Charter/skills read surface

**Decision**: render from the status surface only (`metatron_status` fields:
charter provenance `charter_default`, `charter_locked`, `charter_preset`,
`skills` list, `skills_locked`, stage) — the exact fields spec 046 added;
lock notices reuse the stage-name vocabulary via the skin lookup (TASK-121's
contract, if merged; else the existing skin.StageName). No client-side file
parsing (FR-004); the `[e]` action targets the file path the status/world
handle already knows.

**Rationale**: the daemon is the authority on binding status (stage forks
live there); duplicating that logic client-side would drift.

## R6 — The card seam (scope ruling operationalized)

**Decision**: `type consoleCard interface{ renderCard(width int) string }` +
`Model.consoleCards []consoleCard` composed between the turn stream and the
read surface at render time; this feature ships the type, the composition
slot, and its tests with ZERO producers (slice always empty). TASK-127's
shared report-card renderer and TASK-115's stopping-point production plug in
later. The design page's control-table row stays `unbuilt (wave 3/4)` for
the card CONTENT with a note naming the shipped seam symbol.

**Rationale**: the chronicle-⏎ reserved-seam precedent — an honest,
already-wired attachment point beats a placeholder card that renders
nothing meaningful.

## R7 — Design-page amendments in scope

pages/guardian-console.md (shipped; symbols; seam note), panels/systems.md
(shipped; key 5), panels/guardian.md (telemetry rows removed → pointer to
systems.md), panels/dock.md (4-tab row + key), patterns/keymap.md (G/5/e +
console scroll + footer hints + parity gaps for G/e), pages/solo-views.md
(systems narrow reachability), overlays/help.md (content gains the new keys
per its content contract). All re-pinned; `check-tui-design.mjs --changed`
gates.
