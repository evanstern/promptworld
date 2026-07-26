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
verified_against: 048259bb42b03cc6ebeb13a49f367c2e3a7d4d37
---

# TUI input, focus, and help overlay

Split from [[tui-client]] (corpus-spec v2 size-budget split, summary-style);
covers focus contract, time controls, help overlay — the parent keeps map
view, dock tabs, chronicle feed.

## Focus contract and time controls

Input follows the **focus contract** (`docs/design/tui/patterns/focus-contract.md`):
viewing never captures typing; `m` focuses the minibuffer (amber border, inline
`esc release · ⏎ send` hint), `esc` always releases, and no keypress is a
silent no-op — the old active-guardian-pane-owns-every-key rule is gone.
Time controls (minibuffer unfocused): space toggles pause/resume from
last-known status; `[`/`]` step `speedSteps` (1x → 4x → 8x → 16x → 32x —
max deliberately off the watchable ladder, TASK-20); `q` detaches — the
world keeps running. On an ended world (spec 044) all three clock keys are
no-ops and the footer says so (postmortem posture above).

**The look-cursor mode's pane focus** (spec 074-look-cursor, TASK-142;
[[tui-dock-tabs]] owns the TILE view) is the contract's newest scope note:
`⏎`/`tab` focusing the TILE pane draws the SAME amber-border rule as the
minibuffer/console sub-panels, but it is a **drawn selection scope, not a
text client** — no printable key ever buffers there in any of its three
focus layers (cursor/pane/drill). The "exactly one text-capture
client" claim (the minibuffer, reused verbatim by the guardian console)
holds — `internal/tui/focus_test.go`'s `TestLookModeNeverCapturesText`
proves it mechanically. `handleLookKey` (`look.go`) sits in `handleKey`'s
chain between the console and inspect checks — layered like
`handleInspectKey`/`handleVillagersKey`, claiming the whole contested key
set so the two dormant layers (whose tab-visibility predicates read false
during the borrow) never spuriously fire.

## Help overlay

Spec 045/TASK-116 (`help.go`): `?` from any non-text-entry
mode opens context-sensitive help — head of the key-dispatch/esc-release
chain (help → minibuffer → decisions → detail → solo → home) — checked in
`handleKey` right after ctrl+c, minibuffer unfocused only, so a focused `?`
still types into the buffer. The overlay freezes its opening mode
(`helpMode`) and owns the keyboard while open: `t` flips the page's
basic/advanced key tiers, `tab`/`shift+tab` cycle the five sections (mode
keys · screen walkthrough · lessons pull-reference · spec-056 ceremonies
replay · spec-063 guardian), `n`/`p` page the seven mode pages (global/home,
minibuffer, inspect, villagers roster/detail, solo/narrow, spec-074
look-cursor tile inspection; how the minibuffer page stays reachable),
`J`/`K` scroll (standard pager idiom), `esc`/`?` dismiss exactly one layer;
every other key is inert. `currentHelpMode` checks look-cursor right after
the console branch (before inspect/villagers) — same reason the console
check sits first:
`Model.dockTab`/`active` persist unchanged under the borrow and would
otherwise mis-route.

**Badge deep-link (spec 074-look-cursor, FR-011)**: `openHelp` gains a
pre-focus step — with any conditional header badge active (`[degraded]`,
`[llm: …]`, `[suppressed: …]`, in header order), the overlay opens
on the screen-walkthrough section, scrolled to that badge's `headerAnatomy`
row (`firstActiveBadgeRow`, resolved from the SAME table
`helpWalkthroughLines` renders, so the scroll target cannot drift from the
page). No active badge keeps the open byte-identical to before this feature
(keys section, scroll 0) — applying corpus-wide, not only from the guardian
tab.

Rendering is body replacement (the solo zoom slot) in both layouts — chrome
stays; output exactly terminal-height. Content is static tables in
`help.go`: per-mode key rows (basic ≈ the footer-hinted set),
`headerAnatomy` rows for every conditional badge, dock-tab rows, and a
`mapGlyphs` table **shared with `renderMapGrid`'s legend line**
(`legendGlyphLine`) so glyph walkthrough and map legend cannot silently
diverge — extraction fixed a real gap: the gru's `G` was drawn but never in
the legend text. Since spec 068 ([[tile-registry]]),
`mapGlyphs` IS the tile registry — the spec-045 glyph-key triple grown into
full rows (style token, state variants, world binding), relocated to
`internal/tui/tiles.go`; `help.go` keeps only `legendGlyphLine` and the
walkthrough renderer reading it, preserving the invariant; the registry's
OWN row order is frozen (append-only), keeping this history accurate. Since spec 060 both surfaces carry a `conditionOverlayNote`
prose line naming the three map condition overlays ([[village-lens]]) — a
note, not a `mapGlyphs` row: every overlay is a style variant of an
already-legended glyph, never a new one. The table gained the `✝` grave row
(spec 044), `░` marsh / `▒` sand (spec 068), and `S` stranger (spec 077 —
the night trickster entity, [[event-types-scenario-incidents]];
appended so the pinned prefix stays byte-identical;
[[worldmap-generation]]); the grave row
stays honest via the dead-agent-on-grave carve-out in `renderMapGrid`
(above) — without it the map could never show the glyph it advertises. The lessons section is the pull half of the
**first-occurrence lessons projection** (spec 055/TASK-117; `lessons.go`):
`populateHelpLessons` fills `helpLessons` 1:1 from `lessonCatalog` at client
init, so the overlay lists every lesson the push half can ever show — the
placeholder line survives only as a defensive empty-table branch. The push
half is the lesson row above: `lessonTriggers.ingest` folds the same event
stream `applyEvent` feeds the decision-trace projection, firing the 12
static catalog entries (5 mechanics — first suppression/gru
attack/charge-regen/order-expiry/death; 7 prompting since spec 077's
tranche 2 — first rejected `cog.tool_call` (`rejected_*` verdicts only;
spec 077 narrowed: an explain answer's `read_ok` is not a refusal),
first custom charter/fuzzy order/explain answer/report card/skill file, and
the same-refusal-pattern wrong-thing detector — the catalog's ONE stateful
trigger seam, a bounded session-local `lessonFold` counting rejections per
identical non-empty reason, third-strike firing, never on mixed reasons;
`first-faith-event` is deliberately absent, riding TASK-118) at
most once per player: one active lesson at a time (text line + `→` pointer
line ending `(? for more · x dismiss)`), dwelling until its done-signal
event or a global `x` dismiss, with a bounded FIFO queue whose stale
entries decay rather than surface late. Seen-state is per-user,
cross-world (`internal/worlds`' `lessons-seen.json` beside `unlocks.json` — same
load-tolerant/advisory/atomic-write discipline; marked when a lesson
SURFACES, so a decayed queue entry can still fire later); every catalog
string resolves skin tokens via `lessonSkinResolve` (a bounded
default-table fallback from spec 052's contract until the skin runtime
merges). The ceremonies (spec 056) and guardian (spec 063, D9) sections are
its other two status-derived additions: [[takeover-surfaces]]
covers the former (a stored-content replay of every unlocked stage, sharing
the live ceremony's rendering helpers), [[grounded-feedback]] the latter
(the stage-keyed, model-free "what asking looks like" page, also covering
spec 078's forward-ladder block below it). All content is model-independent
— byte-identical with nil status/replica (the no-LLM floor beneath an
absent angel). Footer hints advertise `· ? help` in every mode except the
focused minibuffer; an open overlay swaps in its own footer hint. A keymap
sweep test (`help_test.go`) mechanically ties
every advertised binding to a real handler and every handled binding to
exactly one tier of its mode page.

## Back to parent

[[tui-client]] links here for input/focus/help; its Connections section
lists [[takeover-surfaces]]/[[grounded-feedback]] (the help
overlay's ceremonies/guardian sections) and [[stage-defaults]] (the lessons
projection's stage-shaped defaults).
