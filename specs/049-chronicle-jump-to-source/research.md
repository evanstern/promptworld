# Research: Chronicle jump-to-source (spec 049)

## R1 — Expressing "center on (x,y)" in the existing camera model

**Decision**: `centerCameraOn(x, y)` sets `panX = x − centroidX`,
`panY = y − centroidY`, where the centroid is the same live-wanderer centroid
`renderMapGrid` computes (`internal/tui/views.go:444-461`). No new camera
state, no new clamping — render-time `clampInt` on the window origin and
`clampGeometry`'s pan cap already bound every value.

**Rationale**: the camera's only state is the pan offset from the wanderer
centroid (`tui.go:85`). Reusing it means `c` (recenter, sets pan to 0,0) and
arrow-key panning keep their exact semantics for free — a jump IS a pan, just
computed instead of accumulated. The spec's "suspends auto-follow exactly as
manual panning does" falls out naturally: any nonzero pan already renders
`MAP · panned (c to recenter)` (`views.go:298`).

**Alternatives considered**: a separate "camera locked to (x,y)" mode —
rejected: new state, new interactions with follow/recenter, and the spec
explicitly wants pan-equivalent behavior. Extracting the centroid into a
shared helper is required either way (jump-time code needs the same number
render-time uses; duplicating the loop risks drift).

## R2 — Bubble Tea mouse enablement

**Decision**: add `tea.WithMouseCellMotion()` to the `tea.NewProgram` call
(`cmd/promptworld/commands.go:745`); handle `tea.MouseMsg` in `Update`,
acting only on left-button *release* (`tea.MouseActionRelease` /
`tea.MouseButtonLeft`) to avoid double-firing on press+release, and only when
the click lands inside the chronicle's rendered row region.

**Rationale**: CellMotion is the standard bubbletea option for click
reporting with per-cell coordinates; AllMotion adds hover churn nothing
consumes. Release-only matches terminal-UI convention. All other mouse events
fall through unhandled — keyboard behavior is untouched (FR-005).

**Alternatives considered**: `tea.WithMouseAllMotion()` — rejected (constant
motion messages, no consumer); OSC 8 hyperlinks — not a pointing device.

## R3 — Click → chronicle line hit-testing

**Decision**: record the chronicle's rendered geometry at render time on the
Model (a small `chronHitRegion` struct: panel origin/size, first visible row's
event index, rows-per-event map or per-row index slice, filled in by the
chronicle body renderers each `View()`), and hit-test `tea.MouseMsg`
coordinates against it. Clicks resolve to the event index rendered on that
row; clicks on non-event rows (borders, detail pane, title) are no-ops.

**Rationale**: geometry in this package is derived fresh every `View()`
(`clampGeometry` comment, `tui.go:1220-1246`) — there is no persistent layout
object to query, so the renderer that already knows which event each row shows
is the only honest source. Recording render geometry for input routing is the
established Bubble Tea pattern for hit-testing composed views.

**Alternatives considered**: re-deriving layout math in the mouse handler —
rejected: duplicates the row-window/wrapping logic (narrow dock wraps events
to ≤3 lines; solo is 1 line) and WILL drift; bubblezone-style ANSI markers —
rejected: new dependency for one click target.

## R4 — Subject resolution (which payload fields, which actor)

**Decision**: a `resolveSubject(e store.Event, replica …) (name string, x, y
int, ok bool)` helper beside the digest catalog in `digest.go`, resolving in
spec FR-002 order: (1) primary actor — the same per-type agent-index fields
the digest grammar already reads (speaker/initiator/agent conventions, e.g.
`agent`, `speaker`, `killer`-style fields per event type), looked up in the
live replica when present and alive → its current `X, Y`; (2) explicit
payload position — per-type coordinate fields (`x`/`y`, `at`, position
objects) as recorded at emission; (3) neither → not locatable. Resolution
inspects only known top-level fields — never a recursive payload scan
(bounded work; `world.migrated` stays cheap and unlocatable).

**Rationale**: `digest.go` is the one place that already encodes per-event-
type payload knowledge (1,201 lines of exactly this); the catalog-sweep test
pattern (`TestCatalogSweep`, TASK-100 precedent) can then enforce SC-002's
jump-or-hint totality mechanically for every cataloged type.

**Alternatives considered**: a generic reflective search for x/y-shaped
fields — rejected: unbounded over oversized payloads and semantically wrong
(a conversation payload may carry both participants' positions; "primary
actor" is a per-type judgment the digest already makes).

## R5 — Narrow-fallback behavior after a successful jump

**Decision**: on a successful jump in the narrow single-pane fallback, switch
the visible pane to the map view (the fallback's existing pane-switching
state), keeping the paused clock and chronicle selection intact so returning
to the chronicle resumes inspect exactly where the player left it.

**Rationale**: FR-007 — an invisible camera move is a silent outcome, which
the spec bans. Widescreen needs no switch (map always visible).

## R6 — Hint surface for unlocatable events

**Decision**: the detail pane's actions bar (the `[future: actions]` slot,
`views.go:1216`) is the single hint surface: locatable events render
`⏎ jump to <name> (x,y)` (mouse: click line), unlocatable render
`no location for this event`. Pressing `⏎` on an unlocatable event changes
nothing visible beyond the already-present hint — the bar IS the honest
no-op's explanation (US3 AS-2: one surface, not two).

**Rationale**: spec 047 documented this exact slot as the attachment surface;
a transient toast/flash would add a new UI primitive for a Wave-2 quick win.

## R7 — Design-page amendments in scope

**Decision**: amend `panels/chronicle.md` (jump-to-source control-table row →
real symbols + `⏎` · click bindings; parity-rollout note gains "first mouse
target shipped"), `patterns/keymap.md` (inspect-mode `⏎` row → jump action;
doctrine rule 3's "zero controls have a real mouse target" statement updated;
migration note), and re-pin `verified_against` on both. `panels/minibuffer.md`
and other pages are untouched (their parity notes don't change).
`scripts/check-tui-design.mjs --changed` gates the PR.
