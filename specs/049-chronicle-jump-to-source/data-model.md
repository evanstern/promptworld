# Data Model: Chronicle jump-to-source (spec 049)

No persistent data, no wire-protocol changes. All entities are client-side,
per-frame or per-keypress values inside `internal/tui`.

## Subject (resolution result)

The outcome of resolving one chronicle event to a jump target.

| Field | Type | Meaning |
|---|---|---|
| `name` | string | display name of the primary actor, or a per-type place label when only a payload position exists |
| `x`, `y` | int | map tile the camera centers on |
| `ok` | bool | false → event is unlocatable (hint path) |

**Validation**: resolution order is fixed (FR-002): live primary actor →
explicit payload position → not locatable. Only known top-level payload
fields are consulted (bounded work). Dead/despawned actors fall through to
the payload position, never to a stale coordinate.

## detailAction (existing type, now populated)

`internal/tui/tui.go:1159` — `{Label string}`; gains whatever minimal fields
the actions bar needs (at most a disabled/enabled distinction expressed by
which label is returned). `detailActions(e)` returns exactly one action for
every event: the jump affordance (`⏎ jump to <name> (x,y)`) or the honest
absence (`no location for this event`). Never nil after this feature —
SC-002's totality is a property of this function and is what the catalog
sweep asserts.

## chronHitRegion (new, render-time geometry record)

Filled by the chronicle renderers each `View()`; consumed by the mouse
handler the next `Update()`.

| Field | Type | Meaning |
|---|---|---|
| panel origin/size | ints | rendered chronicle list rectangle in screen cells |
| row → event index | slice | which ring-buffer event each rendered row shows (wrapped events occupy several rows, all mapping to one index) |
| valid | bool | false while the chronicle isn't rendered (other tab, running mode, narrow non-chronicle pane) |

**State transitions**: overwritten every frame; invalidated whenever the
chronicle body isn't part of the frame. A click is only actionable when
`valid && clock paused` (FR-004).

## Camera pan (existing fields, new writer)

`Model.panX/panY` (`tui.go:85`) gain one new writer, `centerCameraOn(x, y)`:
`pan = target − liveWandererCentroid` (research R1). Existing readers,
clamps, and the `c`-recenter reset are unchanged.
