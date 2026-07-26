---
name: tui-input-help
description: The focus contract and time controls (space/[/]/q), and the full-screen help overlay: per-mode key tables (including the look-cursor mode's own page), the badge deep-link, the shared map-glyph legend, the first-occurrence lessons pull-reference, and the ceremonies/guardian sections. Split from [[tui-client]]; read when touching help.go, lessons.go, or look.go's key layer.
kind: component
sources:
  - internal/tui/tui.go
  - internal/tui/help.go
  - internal/tui/lessons.go
  - internal/tui/tiles.go
  - internal/tui/look.go
  - internal/worlds/unlocks.go
verified_against: b6a20eaa4da1073a69959a5aff69591d931103a9
---

# TUI input, focus, and help overlay

Split from [[tui-client]] (corpus-spec v2 size-budget split, summary-style):
this note covers the focus contract, time controls, and the help overlay. See
[[tui-client]] for the map view, dock tabs, and chronicle feed.

## Focus contract and time controls

Input follows the **focus contract** (`docs/design/tui/patterns/focus-contract.md`):
viewing never captures typing; `m` focuses the minibuffer (amber border, inline
`esc release · ⏎ send` hint), `esc` always releases, and no keypress is a
silent no-op — the old rule where the guardian pane owned every key while
active is gone. Time controls (minibuffer unfocused): space toggles
pause/resume based on last-known status; `[`/`]` step through `speedSteps`
(1x → 4x → 8x → 16x → 32x — max is deliberately off the watchable ladder,
TASK-20); `q` detaches — the world keeps running. On an ended world (spec
044) all three clock keys are no-ops and the footer says so — see the
postmortem posture above.

**The look-cursor mode's pane focus** (spec 074-look-cursor, TASK-142,
[[tui-dock-tabs]] owns the mode's TILE view itself) is the contract's newest
scope note: `⏎`/`tab` moving keyboard focus into the TILE pane draws the
SAME amber-border rule the minibuffer/console sub-panels do, but it is a
**drawn selection scope, not a text client** — no printable key ever
buffers there, in any of the mode's three focus layers (cursor/pane/drill).
The "exactly one text-capture client" claim (the minibuffer, reused
verbatim by the guardian console) stays true; `internal/tui/focus_test.go`'s
`TestLookModeNeverCapturesText` is the mechanical proof. `handleLookKey`
(`look.go`) itself sits in `handleKey`'s chain between the console check and
the inspect check — layered exactly like `handleInspectKey`/
`handleVillagersKey`, claiming the whole contested key set so those two
dormant layers (their own tab-visibility predicates read false during the
borrow) never spuriously fire.

## Help overlay

**Help overlay** (spec 045/TASK-116; `help.go`): `?` from any non-text-entry
mode opens a context-sensitive help overlay — the head of the key-dispatch and
esc-release chain (help → minibuffer → decisions → detail → solo → home) —
checked in `handleKey` right after ctrl+c and only when the minibuffer is
unfocused, so a focused `?` still types into the buffer. The overlay freezes
the mode it opened from (`helpMode`) and owns the keyboard while open: `t`
flips the mode page's basic/advanced key tiers, `tab`/`shift+tab` cycle the
overlay's sections — mode keys · screen walkthrough · lessons pull-reference
· (spec 056) ceremonies replay · (spec 063) the guardian, five in all — `n`/`p`
page across all seven mode pages (global/home, minibuffer, inspect, villagers
roster/detail, solo/narrow, and — spec 074-look-cursor — look-cursor tile
inspection; how the minibuffer page stays reachable), `J`/`K` scroll via the
standard pager idiom, and `esc`/`?` dismiss exactly one layer; every other
key is inert. `currentHelpMode` checks the look-cursor mode right after the
console branch (before inspect/villagers), for the same reason the console
check itself sits first: `Model.dockTab`/`active` persist unchanged
underneath the borrow and would otherwise mis-route.

**Badge deep-link (spec 074-look-cursor, FR-011)**: `openHelp` gains a
pre-focus step — with at least one conditional header badge active
(`[degraded]`, `[llm: …]`, `[suppressed: …]`, checked in that header order),
the overlay opens directly on the screen-walkthrough section, scrolled so
that badge's `headerAnatomy` row is visible (`firstActiveBadgeRow`, an index
resolved from the SAME shared table `helpWalkthroughLines` renders, so the
scroll target can never drift from what's actually on the page). No active
badge keeps the open byte-identical to before this feature (keys section,
scroll 0) — it applies corpus-wide, not only from the guardian tab.

Rendering is body replacement (the solo
zoom slot) in both layouts — chrome stays, output remains exactly
terminal-height. Content is static tables in `help.go`: per-mode key rows
(basic ≈ the footer-hinted set), `headerAnatomy` rows covering every
conditional badge, dock-tab rows, and a `mapGlyphs` table **shared with
`renderMapGrid`'s legend line** (`legendGlyphLine`) so the overlay's glyph
walkthrough and the map legend cannot silently diverge — extracting it also
fixed a real gap: the gru's `G` was drawn but never listed in the legend
text. Since spec 068 ([[tile-registry]]), `mapGlyphs` is the tile registry
itself — grown from the spec-045 glyph-key triple into full rows carrying a
style token, state variants, and a world binding, relocated to
`internal/tui/tiles.go`; `help.go` keeps only `legendGlyphLine` and the
walkthrough renderer that read it, so the two surfaces still cannot silently
diverge, and the registry's OWN row order is frozen (append-only) so this
history stays accurate. Since spec 060, both surfaces also carry a `conditionOverlayNote`
prose line naming the three map condition overlays ([[village-lens]]) — a
note rather than a `mapGlyphs` row, since every overlay is a style variant
of an already-legended glyph, never a new one. The shared table gained the `✝` grave row with spec 044, the `░` marsh /
`▒` sand rows with spec 068, and the `S` stranger row with spec 077 (the
night trickster entity, [[event-types-scenario-incidents]] — appended, so
the pinned prefix stays byte-identical; [[tile-registry]],
[[worldmap-generation]]); the grave row is the
dead-agent-on-grave carve-out in `renderMapGrid` (above) that keeps the row
honest — without it the map could never actually show the glyph it
advertises. The lessons section is the pull half of the
**first-occurrence lessons projection** (spec 055/TASK-117; `lessons.go`):
`populateHelpLessons` fills `helpLessons` 1:1 from `lessonCatalog` at client
init, so the overlay lists every lesson the push half can ever show — the
placeholder line survives only as a defensive empty-table branch. The push
half is the lesson row above: `lessonTriggers.ingest` folds the same event
stream `applyEvent` feeds the decision-trace projection, firing each of the
12 static catalog entries (5 mechanics — first suppression/gru
attack/charge-regen/order-expiry/death; 7 prompting since spec 077's
tranche 2 — first rejected `cog.tool_call` (`rejected_*` verdicts only,
narrowed by spec 077: an explain answer's `read_ok` is not a refusal),
first custom charter, first fuzzy order, first explain answer, first
report card, first skill file, and the same-refusal-pattern wrong-thing
detector — the catalog's ONE stateful trigger seam, a bounded
session-local `lessonFold` counting rejections per identical non-empty
reason, firing on the third strike, never on mixed reasons;
`first-faith-event` is deliberately absent, riding TASK-118) at most once
per player: one active lesson at a time (text line + `→` pointer line ending
`(? for more · x dismiss)`), dwelling until its done-signal event or a
global `x` dismiss, with a bounded FIFO queue whose stale entries decay
rather than surface late. Seen-state is per-user and cross-world
(`internal/worlds`' `lessons-seen.json` beside `unlocks.json` — same
load-tolerant/advisory/atomic-write discipline; marked when a lesson
SURFACES, so a decayed queue entry can still fire later), and every catalog
string resolves skin tokens through `lessonSkinResolve` (a bounded
default-table fallback from spec 052's contract until the skin runtime
merges). The ceremonies section (spec 056) and the guardian section (spec
063, D9) are the overlay's other two status-derived additions — see
[[takeover-surfaces]] for the former (a stored-content replay of every
unlocked stage, sharing the live ceremony's own rendering helpers) and
[[grounded-feedback]] for the latter (the stage-keyed,
model-free "what asking looks like" page, also covering spec 078's
forward-ladder block below it). All content is
model-independent — byte-identical with nil status/replica (the no-LLM floor
beneath an absent angel). Footer hints advertise `· ? help` in every mode
except the focused minibuffer; while the overlay is open the footer shows the
overlay's own hint. A keymap sweep test (`help_test.go`) mechanically ties
every advertised binding to a real handler and every handled binding to
exactly one tier of its mode page.

## Back to parent

[[tui-client]] links here for input/focus/help; that note's own Connections
section lists [[takeover-surfaces]] and [[grounded-feedback]] as the help
overlay's ceremonies/guardian sections, and [[stage-defaults]] for the
lessons projection's stage-shaped defaults.
