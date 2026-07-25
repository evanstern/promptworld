# Research: Guardian strip (spec 050)

## R1 — Where the strip's data already lives

**Decision**: reuse verbatim the guardian tab header's sources: charge bank =
`m.status.Clock.MetatronCharges` vs `sim.MetatronChargeCap`
(`internal/tui/views.go:1509-1512` renders exactly the glyph run the design
page mocks); standing orders = `len(m.status.Orders)` (the tab's
standing-orders block's input); regen boundary = next tick `t` with
`t % chargeRegenTicks == 0` after the current tick (the executor's rule,
`internal/sim/executor.go:53-56`), formatted with the app's existing
game-clock time formatter.

**Rationale**: the design page names these sources field-for-field; the strip
is "a glanceable TUI-side restatement", so sharing the source is the honesty
guarantee — the strip and the tab can never disagree.

**Alternatives considered**: subscribing to `metatron.charge_regenerated`
events to maintain a countdown — rejected: derivable pure state; event
tracking adds a second source of truth.

## R2 — Exposing the regen cadence to the client

**Decision**: `chargeRegenTicks` is unexported in `internal/sim`; add a
read-only exported accessor (or exported const) mirroring the existing
`MetatronChargeCap` export pattern (`internal/sim/metatron.go:20-22`,
precedent noted at `agents.go:843`). The TUI computes
`next = tick + (chargeRegenTicks − tick % chargeRegenTicks)` and renders it
as game time.

**Rationale**: the client replica already imports `sim` for the cap; the
cadence is the same kind of doctrine constant. No IPC change, no new field on
the wire (D1 holds trivially — the CLI can derive the same number).

**Alternatives considered**: putting next-regen on the status message —
rejected: wire change for a derivable value.

## R3 — Row-budget integration and fold order

**Decision**: extend `rowBudget`/`computeRows` (`internal/tui/layout.go`)
with a `Strip` row: present when `totalRows − (header+strip+minibuffer+
footer) ≥ bodyMin` per layout.md's re-derived budget; when it folds (last in
the ruled order — the lesson row and villager strip don't exist yet, so in
this feature's code the strip is the only foldable new chrome and folds at
exactly the ruled threshold), `minibufferView`'s dormant branch renders the
relocated form (budget prefix + existing placeholder). `computeRows` stays a
pure function of `totalRows` (plus a strip-visible flag), preserving
layout.go's "pure functions, Update computes once per WindowSizeMsg" design.

**Rationale**: layout.md ruling a is a total order; implementing only this
feature's step keeps the arithmetic honest without building the whole
stage-defaults fold machinery early (TASK-128's job). The pure-function shape
is what `layout_test.go`'s sweep tests can exhaustively check.

**Alternatives considered**: implementing the full 5-step fold state machine
now — rejected: three of its rows don't exist; TASK-128 owns the machinery.

## R4 — Honest-degradation rulings the authored page left open

**Decision** (recorded back onto `panels/guardian-strip.md` same-PR):

1. **Full bank** → regen segment omitted (the executor only fires below cap;
   forecasting an arrival that isn't scheduled would be a lie).
2. **No status snapshot yet** → row present but blank (layout stability
   without invented zeros).
3. **Width pressure** → truncate segments right-to-left (faith → orders →
   regen → bank; the bank is the headline) with `…`; never wrap.
4. **Zero orders** → `👁 0 standing orders` renders (true value, not absent
   mechanic — distinguished from the faith case).

**Rationale**: the page's own rule ("degrading to absence rather than to a
misleading zero") applied case-by-case; spec US2/edge cases fix these and the
gate requires the page to carry them once shipped.

## R5 — Narrow-fallback carry

**Decision**: the narrow single-pane fallback already stacks
`content + minibuffer + footer` (`views.go:1558` vicinity); insert the strip
line above the minibuffer in that stack unconditionally (ruling b:
width-independent), sharing the same `guardianStripView(width)` output at
narrow width (truncation per R4.3 handles the columns).

**Rationale**: one renderer, two call sites — the design page's "exactly as
widescreen does".
