---
name: tui-client
description: The Bubble Tea full-screen client — a widescreen map+dock composite over a live world replica maintained by log shipping; the guardian tab label and chronicle Type-column alias resolve through the boot-frozen [[skin]]. Split into six child notes (corpus-spec v2): [[tui-client-mechanics]] (connection/header/layout), [[tui-map-view]], [[tui-dock-tabs]], [[tui-villagers-tab]], [[tui-chronicle-feed]], [[tui-input-help]].
kind: component
sources:
  - internal/tui/tui.go
  - internal/tui/views.go
  - internal/tui/layout.go
  - internal/tui/grammar.go
  - internal/tui/digest.go
  - internal/tui/decisions.go
  - internal/tui/help.go
  - internal/tui/lessons.go
  - internal/tui/reportcard.go
  - internal/tui/tiles.go
verified_against: 4c66d240b2715706964f02cfd2396256c9957d8e
---

# TUI client

`internal/tui` is the attachable full-screen client (`promptworld ui <dir>`), built on
Bubble Tea + Lipgloss. Its core idea: the map renders from a **live replica** of
`sim.State` that the client maintains by log shipping — fetch the state snapshot, then
apply every pushed event through the exact `Apply` reducer the daemon runs. The TUI is
a read replica of the world.

## Surfaces

This note is a summary-style split point (corpus-spec v2 note-size budget): the
client's substance now lives in six child notes, one per surface family. Each
child links back here; this note keeps only routing summaries.

**Connection, header, and layout** — [[tui-client-mechanics]]. How the client
dials the daemon and maintains the live replica by log shipping, the header's
postmortem/governed-speed/LLM-badge/suppressed-horizon readouts, reconnect
resilience, and the widescreen/narrow layout's three-step fold cascade
(villager strip → lesson row → guardian strip).

**Map view** — [[tui-map-view]]. The map camera region: terrain/agent/
structure glyph rendering resolved through the tile registry, condition
overlays, camera pan and jump-to-source, and the legend's stockpile-zone/
chest inspection additions.

**Dock tabs (chronicle, guardian, systems)** — [[tui-dock-tabs]]. The dock's
tab bar, the guardian tab's skin-resolved label, the guardian console
full-height page, and the chronicle/guardian/systems tab contents (LLM
provider table, cognition horizon block, standing-orders block). The
scenario-only exercise tab is covered there too.

**Villagers tab** — [[tui-villagers-tab]]. The villagers roster, detail view,
and decisions sub-view's client-side decision-trace projection.

**Chronicle feed and digest grammar** — [[tui-chronicle-feed]]. The
chronicle pane's narrated/raw toggle and the digest grammar: a
`digestRegistry` entry per cataloged event type, family color-role tints,
high-salience whole-line alerts, and the inspect-mode detail pane.

**Input, focus, and help overlay** — [[tui-input-help]]. The focus contract
and time controls, and the full-screen help overlay (per-mode key tables,
the shared map-glyph legend, the first-occurrence lessons pull-reference,
and the ceremonies/guardian sections).

## Connections

[[ipc-client]] is the transport; [[ipc-protocol]]'s `state` command exists for this
replica pattern; [[sim-state-reducer]] supplies the shared `Apply`; [[chronicle]]
fills the story pane and [[event-types]] the raw feed; [[tile-registry]] (spec 068)
owns every map tile's glyph/style/binding this note's map region and help-overlay
glyph walkthrough render from; [[cli-promptworld]] mounts
it as the `ui` subcommand. The header's governed-speed suffix and the two
governor digest lines read [[cognition]]'s `ShedThreshold` and the
`clock.governor_shed`/`clock.governor_recovered` payload the [[daemon-lifecycle]]
governor sampler emits through the loop. The guardian pane's standing-orders
block and transcript lines project [[guardian-orders]]' `Status.Orders`/
`TurnResult` fields verbatim, with no client-side re-derivation. The header's
`[llm: …]` badge and the guardian pane's per-provider condition line read
[[llm-provider-health]]'s `ProviderStatus.Condition`/`ConditionDetail`/
`ConditionRemedy` fields off the same polled `Status.LLM`. The header's
`[suppressed: …]` badge and the guardian pane's `horizonLines` block both
read the polled `Status.Horizon` — [[ipc-server]]'s `horizonClasses`
composition backed by [[cognition]]'s `LiveHorizon` and
[[llm-orchestrator]]'s `SuppressionCounts` — with no client-side
re-derivation, the same "polled, not projected" posture as the LLM condition
surfaces. The guardian pane's stage segment and the two curriculum digest
rows are [[curriculum-ladder]]'s TUI surfaces (spec 046), reading the
angel's `Status.Stage`/lock fields and the `curriculum.*` event payloads.
The exercise dock tab is [[scenario-machinery]]'s spec-054 TUI surface,
reading `sim.EvaluateRubric`/`sim.ExerciseOutcome` over the replica and the
manifest's `Scenario` block. [[mental-maps]]'s four place-knowledge event types render through
the raw digest feed with no dedicated pane of their own — the map/prompt
side of the feature lives entirely in [[agent-mind]]/[[executor]], not here.
[[takeover-surfaces]] (spec 056) owns the ceremony/postmortem takeover
family that dispatches ahead of every mode this note describes, the shared
report-card renderer, and the help overlay's ceremonies section;
[[grounded-feedback]] (spec 063) owns the `explain` tool's guardian-side
integration, the report-card producer whose stored note this note's
`consoleCard`/digest surfaces render, and the help overlay's D9 guardian
section. [[stage-defaults]] (spec 066) owns the single authority table
`lessonRowDefault` now delegates to and the live re-resolution/first-
occurrence-arrival plumbing wired into `Update`'s `statusMsg` case;
[[village-lens]] (spec 060) owns the villager strip's own content/overflow
rendering and the map's three condition overlays this note's map-region
paragraph describes.

## Operational notes

Rendering requires no daemon round trips — map updates come from pushed events, so the
UI stays smooth at max speed (the chronicle simply scrolls fast). The four spec-029
standing-order event types (`metatron.order_placed`/`order_triggered`/
`order_cancelled`/`order_expired`) carry `digestRegistry` entries (digest.go —
"Guardian set a watch: …" / "…watch came true/released/lapsed", the placed
condition truncated to 80 runes and quoted through the same speech helper as
nudge text; the id-only lifecycle payloads reference the watch by id), so order
activity reaches the raw chronicle feed as well as the dedicated guardian-pane
block and transcript lines above; `TestCatalogSweep` pins the coverage against
[[event-types]]' backticked catalog.
Unit tests cover pane
navigation, replica application, ring capping, quit behavior, the widescreen layout
math (layout.go), the digest grammar (per-family digests + the catalog sweep in
digest_test.go, plain/segment equivalence under wrap), focus-contract key
routing in both layouts, exact-height rendering invariants across sizes and dense
content (including all help-overlay states), the help overlay itself
(help_test.go — per-mode routing, tier/section navigation, the keymap sweep,
no-LLM byte-identity, a zero-side-effect soak, and the lessons-seam fixture),
the lessons projection (lessons_test.go — an exactly-once fixture sweep
across two worlds plus a restart, queue/decay/spacing, catalog↔overlay
equality, seen-file fault tolerance, and a no-`{{`-literal render sweep;
`testModel` isolates `PROMPTWORLD_HOME` so the suite never touches a real
home directory),
and resize round-trips with live selection; an expect-driven PTY smoke test
drives the real binary. When real systems land, dock tabs graduate from stubs without
changing the replica machinery.

