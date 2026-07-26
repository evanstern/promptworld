---
name: tui-client-mechanics
description: How the TUI client connects to the daemon and maintains its live world replica by log shipping, the header's postmortem/governed-speed/LLM-badge/suppressed-horizon readouts, reconnect resilience, and the widescreen/narrow layout's three-step fold cascade (villager strip -> lesson row -> guardian strip). Split from [[tui-client]]; read when touching tui.go, views.go (headerView), layout.go, or digest.go's debtPercent.
kind: component
sources:
  - internal/tui/tui.go
  - internal/tui/views.go
  - internal/tui/layout.go
  - internal/tui/digest.go
verified_against: 048259bb42b03cc6ebeb13a49f367c2e3a7d4d37
---

# TUI client mechanics: connection, header, and layout

Split from [[tui-client]] (corpus-spec v2 size-budget split, summary-style):
this note covers the client's connection lifecycle, header status readouts,
reconnect resilience, and layout composition. See [[tui-client]] for the map
view, dock tabs, villagers tab, chronicle feed, and input/help overlay.

## Connecting and maintaining the replica

`Model` holds the world handle, an `ipc.Client`, the replica, the latest polled
`StatusData`, and a chronicle ring (`chronicleCap = 500` events). All protocol calls
run inside `tea.Cmd`s so the UI never blocks on the socket.

Connection (`connect`): dial → `FetchState` (state JSON + the `last_seq` it reflects)
→ unmarshal into a fresh `sim.NewState(seed)` → `Subscribe(since: last_seq)` — the
replica starts gapless by construction. `listen` delivers one push per invocation and
`Update` re-arms it. `applyEvent` skips seqs already folded into the snapshot, applies
the rest to the replica, bumps its tick, and appends to the chronicle ring.

## Header status: postmortem posture and governed speed

**Postmortem posture** (spec 044): once the run is over, the header's clock
state renders a bold-red `ENDED` token (`styleEnded` — a finality register
`PAUSED`'s amber deliberately doesn't carry) that outranks `PAUSED`
regardless of the clock state the run end landed under. The predicate is
`Model.runEnded()` (tui.go), dual-source by necessity: the replica's
`State.Ended` covers clients attaching after the fact (the snapshot path
never replays folded events), while the pushed `run.ended` event (folded by
`applyEvent`) and the 1-second status poll (`StatusData.Clock.Ended`) cover
the live transition without a reconnect. The same predicate makes the clock
keys (space, `[`, `]`) inert client-side — the daemon's refusal error would
otherwise read as a disconnect — and swaps the footer's pause/resume hint
for `run ended (read-only)` in every mode; all reading surfaces stay fully
functional ([[morgue]]). Since spec 056, `runEnded()` also gates a full-screen
**postmortem takeover** distinct from this header posture — see
[[takeover-surfaces]] for the takeover family (ceremony + postmortem), which
owns the keyboard above every mode this note describes.

**Governed speed** (`headerView` in `views.go`, spec 028 US4): the header's
speed segment renders the EFFECTIVE speed as the world's speed, and — only
while `StatusData.Clock.RequestedSpeed` is set and differs from `Speed` (the
governor has shed at least one notch) — gains a plain-language suffix via
`governedSpeedSuffix`: `"asked 32x — 3 minds in flight, debt 140%"`. An
ungoverned world (`RequestedSpeed` empty) renders byte-identically to
pre-028. Since spec 034, the header also gains a red `[llm: <provider>
<kind>]` badge (the `[degraded]` badge's pattern) whenever any provider
carries an active health condition — `firstLLMCondition` reports the first
name-sorted affected provider; no condition active renders no badge
([[llm-provider-health]]). Since spec 037 (US1, FR-005), the header gains a
further warn-styled `[suppressed: class, class]` badge whenever ≥1 watched
class in the polled `StatusData.Horizon` is currently suppressed —
`suppressedHorizonClasses` filters the wire slice (already in
`WatchedClasses` order) with no client-side re-derivation; a world with no
horizon (no-LLM, or nothing suppressed) shows no badge. `debtPercent` (`digest.go`) is the one shared arithmetic behind both
this suffix and the digest lines below: the measured debt expressed as a
whole percent of `cognition.ShedThreshold`, rounded to the nearest percent.
The raw chronicle feed's digest grammar gains two entries for the same
feature: `clock.governor_shed`/`clock.governor_recovered` each render as
`"governor shed/recovered <from>→<to> debt=N% jobs=N"`, in the terse
`clock.degraded`-line style (the `requested` payload field is omitted here —
the from→to transition already carries the notch delta).

## Resilience

Resilience: errors become `disconnectedMsg` → the header shows the failure and a
2-second retry loop re-dials; a `dropped` push (subscriber overflow) tears the client
down and reconnects from a fresh state snapshot, because the replica may have missed
events. One exception is fatal (TASK-19): `ipc.ErrReplyTooLarge` (a reply over the
protocol's 64 MiB ceiling — reconnecting cannot shrink the state) quits instead of
retrying, rendering the reason in the final view and exposing it via
`Model.FatalErr()`, which `cmdUI` turns into a non-zero exit. A 1-second poll refreshes the clock/status line (quiet ticks produce no
events, so the replica's tick alone would lag).

## Layout

Layout (TASK-34; design reference in `docs/design/tui/` — entry points
`INDEX.md`/`anatomy.md`; since TASK-123's v2 taxonomy the dock's per-tab
content is split across `panels/guardian.md` (fiction layer), `panels/
systems.md` (telemetry), and `panels/villagers.md`, with `panels/dock.md`
itself covering only the tab-container chrome): at ≥112 columns the
client renders the **widescreen composite** — the map on the left and a tabbed
**dock** on the right in a 50/50 split (`computeColumns` in layout.go; the map's
viewport derives from the column budget via `mapViewportTiles`), a one-line
borderless **guardian strip** (spec 050, reorient decision 7: charge-bank
glyphs + `(N/cap)`, a `next +1 @ <time>` regen forecast derived from
`sim.GuardianChargeRegenTicks`, and the replica's standing-order count —
`guardianStripView`, each segment degrading to absence rather than a
misleading zero), a one-line
**Guardian minibuffer** above the footer, and per-mode footer hints. Since
spec 060 (TASK-129, reorient decision 12) a one-line borderless **villager
strip** sits directly under the header — see [[village-lens]] for its
glanceability content and the map's sibling condition overlays. Since
spec 055 (TASK-117, reorient decision 5) a two-line borderless **lesson row**
sits above the guardian strip whenever a first-occurrence lesson is active
and the stage default allows it (`lessonRowDefault` in layout.go, since spec
066 a pure delegation onto the shared stage-defaults table —
[[stage-defaults]]: on at curriculum stages 1–2; at stage 3+/pre-ladder the
row never renders and a
quiet `[lesson]` header badge points at the `?` overlay instead —
`lessonBadgeVisible`); under height pressure the fold cascade now runs THREE
steps in ruled order (`patterns/layout.md` ruling a): the villager strip
folds FIRST (`rowBudget.VillagerStrip`, relocating to a `[N villagers]`
header badge — `villagerCountBadge`), then the lesson row folds to its
badge (`rowBudget.Lesson`), then the guardian strip folds LAST
(`rowBudget.Strip`, `computeRows` keeps it while body ≥ 10 rows); folded,
the guardian strip's content relocates into the minibuffer's dormant
placeholder line instead of hiding — the villager strip and lesson row
have no such relocation, only their header-badge fallback. Below 112
columns it falls back to the original single-pane UI (header + tab bar + one
active pane), unchanged except that the guardian pane carries the same strip
above its minibuffer and the lesson row is carried above whichever pane is
active with identical stage defaults (`narrowView`) — the villager strip is
never carried into the narrow layout at all (stage-defaults ruling b).
`View` output is exactly terminal-height in every mode
(every panel body is clipped to its row budget — `clipContent`), and resizes
re-clamp pan/selection state (`clampGeometry`).

## Back to parent

[[tui-client]] links here for connection/header/layout mechanics; that note's
own Connections and Operational notes sections carry the full cross-reference
list and test coverage for these paths.
