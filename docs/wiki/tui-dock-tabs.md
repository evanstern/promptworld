---
name: tui-dock-tabs
description: The dock's tab bar, the guardian tab's skin-resolved label, the guardian console full-height page, and the chronicle/guardian/systems tab contents (LLM provider table, cognition horizon block, standing-orders block). Split from [[tui-client]]; the villagers tab is its own sibling note, [[tui-villagers-tab]]. Read when touching tui.go's dock dispatch or the console/tab rendering in views.go.
kind: component
sources:
  - internal/tui/tui.go
  - internal/tui/views.go
  - internal/tui/reportcard.go
verified_against: 0fd2104c59c54be8e8071d319fa4ce192083faf3
---

# TUI dock tabs (chronicle, guardian, systems)

Split from [[tui-client]] (corpus-spec v2 size-budget split, summary-style):
this note covers the dock's tab bar, the guardian console, and the chronicle/
guardian/systems tab contents. The villagers tab is [[tui-villagers-tab]]; the
scenario-only exercise tab follows below. See [[tui-client]] for the map view,
chronicle digest grammar, and input/help overlay.

## Dock tabs

The **dock** hosts four tabs (five on a scenario world, see below) — keys
`2`/`3`/`4`/`5` select, the same key again zooms the tab solo, `1`/`esc`
return to the composite. Since spec 052
the guardian tab's own label is [[skin]] data, never a compiled-in string:
`Model.paneName(p)` (`tui.go`) resolves every OTHER tab from the static
`paneNames` table but resolves `paneGuardian` through `m.sk().TabLabel()`
(`m.sk()` reads the boot-resolved `*skin.Skin` the daemon's status carries,
`FromFacts`-rebuilt client-side, [[skin]]) — the default renders `guardian`
byte-identically, an alternate skin's tab label renders live with no client
restart. The tab bar (`views.go`) and the help overlay's dock-tab table
(`help.go`) both call `paneName`, so the two can never disagree. Since spec 053
the guardian conversation also has a first-class full-height page: the
**guardian console** (`Model.console`, `consoleView`), opened with global
`G` from home/solo/narrow (shadowed only where inspect/villagers modes
already bind `G` to jump-to-last) and closed with `G`/`1`/`esc` — document-
style labeled turn blocks over the shared transcript (`consoleTurnLines`,
special ⚡/👁/⏲/» rows inline), tail-anchored `J`/`K` scrollback, an
a `consoleCard` composition seam (spec 053's empty seam, populated by spec
056's shared rubric-checklist renderer and spec 063's report-card producer —
[[takeover-surfaces]], [[grounded-feedback]]; `Model.consoleCards`,
`rebuildConsoleCards`),
a charter/skills read surface from status fields
(`charterReadSurfaceLines`: provenance, binding, honest lock notices), an
`e` → `tea.ExecProcess($EDITOR, charter.md)` handoff with a content-hash
"charter changed — next turn binds it" one-shot notice, and the standard
minibuffer as its composer (states/focus contract unchanged). The tabs:
**chronicle** (default;
see below), **guardian** (the guardian transcript — replies stream here, or
badge the tab `guardian •` when it isn't visible; the pane header shows the charge
bank plus the spec-021 instruction/capability provenance summary — charter
default/custom, skill-file count when non-zero, and the granted-tool summary from
`Status.GrantedTools`, quiet for a full-grant default world — [[guardian]] —
joined, since spec 046, by a `stage: <skin name>` segment
(`consoleStageSummary` in tui.go) that appends `(charter locked to
<preset>)` while the stage-1 instruction lock binds, read off the polled
`Status.Stage`/`CharterLocked`/`CharterPreset` with no client-side
re-derivation; empty (header byte-identical) for a pre-ladder/ungated
world — [[curriculum-ladder]]; the
transcript itself gains a `👁 watch set`/`👁 watch released` line for a
placed/cancelled standing order and a `⏲` line for a landed pause/start/
adjust_speed meta tool call, alongside the existing `⚡` vision/omen line;
below the transcript, a `👁 standing orders (n)` block (spec 029,
`orderStatusLines`, [[guardian-orders]]) renders one compact row per order
from `Status.Orders` — id, a `~` fuzzy marker, origin, remaining game-day,
status, and condition — present only while orders stand. Since spec 053
(TASK-125, the D10 telemetry split) the guardian tab carries
fiction-layer content ONLY — the telemetry below moved to a fourth
**systems** dock tab (key `5`, `systemsView`/`systemsContentBody`, never
skinned by design): the LLM provider table since spec 024 — `llmProviderLines`,
one row per provider with name, model, up/down glyph, queue, inflight/slots, a
contended marker, and spend share, plus an `(unattributed)` row for pre-024
months, followed by a `spend $X of $Y` wallet line — [[llm-orchestrator]]; a
provider carrying an active health condition (spec 034) gains an indented
continuation line rendering the condition's detail and remedy in the pane's
error style, immediately below its row — [[llm-provider-health]]; beside the
provider table, since spec 037 (US1, FR-006) the pane also gains a
`horizonLines` block, one dim "🜂 cognition horizon" header plus a
`horizonRow` per watched class from the polled horizon — a thinking class
renders a plain "<class> thinking at <speed>"; a suppressed class is
warn-styled with its remedy ("suppressed at <speed> — calibrate or slow
down" for an uncalibrated class, "… — slow down" for a calibrated one,
`horizonRemedy`) and carries the router's own verdict arithmetic verbatim as
a dim trailing detail — no raw enum ever reaches the screen. A trailing dim
"· skipped N" (the class's `SuppressedCount`) appears on every suppressed row
and on a thinking row only once it has ever been suppressed (N > 0), so a
never-suppressed class shows no count clutter). The villagers tab is covered separately in [[tui-villagers-tab]].

## Exercise tab

Since spec 054 (TASK-119), a fifth dock tab, **exercise** (key `6`,
`internal/tui/exercise.go`), joins the row ONLY on a scenario world
(`Model.exerciseID()` reads the attached world's manifest `Scenario` block
plus a live `sim.ExerciseByID` re-check — world-shaped, not stage-shaped; an
ambient world carries no tab, no help row, and no footer-hint digit at all,
never an inert `6`). An attach-time briefing (framing + incident-visibility
mode) shows once per attach and is dismissed by any key while visible
(`exBriefingDismissed`, reset on reconnect — the one deliberate any-key
eater, gated to fire only while the exercise tab is the thing on screen and
the minibuffer is unfocused); afterward the body renders one gauge row per
`sim.EvaluateRubric` term (met/pending marker, backing event count — the
SAME pure function the executor's pass precondition and, since spec 072,
every report-card surface read ([[report-card-renderer]]), so the panel,
the emitter, and the cards can never disagree; both cataloged exercises
evaluate for real), an incident-schedule line under forecast
visibility (omitted, never blanked, under fog), and a pass/fail banner once
`sim.ExerciseOutcome` resolves ([[scenario-machinery]] owns the whole
subsystem). `paneKey`/`dockTabKey` extend to `6`; `nextDockTab`/`prevDockTab`
became `Model` methods so the dock-cycle order can consult `exerciseID()`
and include or skip the tab.

## Back to parent

[[tui-client]] links here for the dock's non-villager tabs; that note's own
Connections section lists [[llm-orchestrator]], [[llm-provider-health]],
[[cognition]], [[ipc-server]], [[curriculum-ladder]], and
[[scenario-machinery]] as the tabs' underlying data sources.
