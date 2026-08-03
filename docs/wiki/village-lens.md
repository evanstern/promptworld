---
name: village-lens
description: The spec-060 village-lens completion — the villager strip (a one-row colonist-bar glanceability read under the header, folding to a header count badge) and the map's condition overlays (needs-critical, suppressed-mind, dying-fire), all style variants over already-legended glyphs, never new ones
kind: component
sources:
  - internal/tui/views.go
  - internal/tui/layout.go
  - internal/tui/help.go
  - internal/sim/agents.go
  - internal/tui/tiles.go
verified_against: aeb0c17a98a8ae1b27fff9111bd009e21841b21c
---

# Village lens (villager strip + map condition overlays)

Spec 060 (TASK-129, reorientation decision 12) is "the village-lens
glanceability win": a one-row colonist-bar strip under the header giving an
at-a-glance read of the roster without opening the villagers tab, plus three
map condition overlays that make a villager's or fire's trouble visible on
the map itself. Both pieces are display-only — no new state, no new tuning constants, no
cursor or keys — reusing the thresholds and telemetry the sim/reflex layers
already compute (D1: presentation, never a new source of truth).

## How it works

**The villager strip** (`villagerStripView`/`villagerStripGlyph`,
`views.go`, US1): one borderless row, `N villagers` (living + dead, matching
the villagers tab's roster count) plus one name-initial glyph per villager
in stable roster order — dead faint `†` (never the map's dead-agent-on-grave
carve-out; the strip has no tile/grave concept), asleep dim/lowercase, awake
bright — the SAME priority/style rules the map's own agent glyphs use, so a
player who already reads the map's vocabulary reads this strip for free.
Width overflow sheds glyphs from the END with a trailing `…N` count (never
a mid-glyph truncation, the corpus-wide "shed, never silently cut"
discipline) — `villagerStripView`'s exact-fit search tries showing everyone
first, then backward-scans for the widest prefix that leaves room for the
overflow marker.

**Row budget and fold order** (`layout.go`): `villagerStripRows` = 1,
unconditionally wanted per world (per [[stage-defaults]]'s authority table,
"on" at every stage in widescreen — it carries no stage-off default the way
the lesson row does, so `computeRows` never takes a `wantsVillagerStrip`
toggle). `rowBudget` gains a `VillagerStrip` field, and the fold cascade has
THREE steps in ruled total order (`patterns/layout.md` ruling a): the
villager strip folds FIRST (step 2, reclaiming 1 row the moment body would
dip below `bodyMin`), then the lesson row (step 3, spec 055), then the
guardian strip LAST (step 4, spec 050 — its budget stays visible by
relocating into the minibuffer's dormant line rather than folding away). The
map legend is body-internal and folds before all three, outside this
function's accounting.

**The header fallback** (`villagerCountBadge`, `headerView`): whenever the
strip isn't rendering this frame — narrow layout (never carried there, per
stage-defaults ruling b) or a widescreen fold under height pressure
(`computeRows(...).VillagerStrip == 0`) — a `[N villagers]` badge appends to
the header, so the count is never lost to folding. Absent entirely with no
roster yet (a pre-connect header stays byte-identical to pre-060).

**Map condition overlays** (`renderMapGrid`, US2, panels/map.md "Wave 5"):
three STYLE variants layered over glyphs the legend already names — never
new glyphs of their own. Since spec 068 ([[tile-registry]]), the styles named
below resolve through the tile registry's classed tokens
(`internal/tui/tiles.go` — `styleAgentCritical`/`styleAgentSuppressed`/
`styleFireDying` are token-derived aliases, not literals of their own), but
the colors, emphasis, and priority rule are the identical pre-registry bytes:

- **Needs-critical** (`needsCritical(n sim.Needs)`): true when a living
  villager's Health/Food/Warmth/Rest crosses the SAME danger-band
  thresholds the reflex's PREP-gate/survival rungs already treat as "in
  danger" (`sim.SurvivalNearDeathBelow`/`SurvivalStarvingRearm`/
  `SurvivalFreezingRearm` — spec 059's existing exports — plus a new
  `sim.DangerRestBelow`, the one export this feature needed since rest has
  no survival-watch kind of its own to piggyback on; it aliases the
  existing `dangerRestBelow` constant rather than naming a new number).
  Morale carries no such band in sim today, so it's excluded — nothing to
  reuse. Rendered `styleAgentCritical` (bold + underline, red). Since spec
  083 this overlay is also, by construction, the map presentation of the
  neglect detector ([[executor-needs-survival]]): the detector fires on the
  same exported band constants this predicate reads, so a neglect-firing
  villager is always already painted critical — no new token, glyph, or
  legend row; the subsumption is a pinned contract
  (`TestRenderMapGridNeglectFiringRendersCritical`,
  `internal/tui/village_lens_test.go`), and the chronicle's whole-line
  alert ([[tui-chronicle-feed]]) is the event-shaped surface.
- **Suppressed-mind** (`agentSuppressedMind(i)`): true when an agent's
  latest decision-trace chain (`m.traces.chainsFor(i)`, most-recent-first —
  [[tui-client]]'s decision-trace projection) is a router suppression — the
  map form of spec 037's "a skipped thought is visible." Rendered
  `styleAgentSuppressed` (faint, a cooler hue than critical red so the
  priority order reads correctly before a player has learned the
  vocabulary).
- **Dying fire** (a third, still-LIT fire state, US2 AS3): inside
  `State.RefuelDyingBelow()`'s window (spec 057's existing dial — the exact
  remaining-fuel threshold the reflex's own refuel-before-cold rule already
  keys on, never inventing a second window) but before `FuelUntil` actually
  passes. Rendered `styleFireDying` (bold, a warmer/redder warn tone,
  distinct from both plain lit orange and cold gray).

Priority when both agent conditions apply: needs-critical wins over
suppressed-mind (physical danger over cognitive telemetry, FR-003/AS4) —
with neither present, rendering is byte-identical to pre-060. The map's
compact legend line and the `?` help overlay's walkthrough both gain
`conditionOverlayNote` ("needs-critical & suppressed-mind mark agents, dying
marks fire — needs-critical wins when both apply") — a prose note rather
than a `mapGlyphs` row, since every overlay is a style variant of an
already-legended glyph, never a new one to key a row on.

Since spec 114 the legend clamps to the terminal width ([[tui-map-view]]), and
`conditionOverlayNote` sits near the tail of the composed line — so on a narrow
terminal it is truncated away and the `?` overlay's walkthrough becomes the only
place the overlay styles are named. The anti-drift discipline is unchanged (one
note, rendered by both surfaces); what changed is that the legend's copy is
width-contingent and the overlay's is not.

## Connections

[[stage-defaults]] (spec 066) is the authority table this feature's villager
strip already had an "on at every stage" row waiting in — `computeRows`'s
new `VillagerStrip` field composes with that table's resolution, not a
second gate. [[tui-client]] hosts `Model`, `renderMapGrid`, the
decision-trace projection (`chainsFor`) the suppressed-mind overlay reads,
and the row-budget/fold-order machinery this feature extends.
[[reflex-policy]]/[[guardian-orders]] own the danger-band constants
`needsCritical` reuses (`SurvivalNearDeathBelow`/`StarvingRearm`/
`FreezingRearm`, spec 059) and the new `sim.DangerRestBelow` sibling.
[[world-tuning]] owns `State.RefuelDyingBelow()`, the dying-fire window's
source. [[tile-registry]] (spec 068) is where the overlay styles now
resolve from (classed tokens, not literals) — the overlay behavior is
unchanged, only the styling mechanism moved. [[cognition]]/[[agent-mind]]
decide a router suppression; [[tui-client]]'s decision-trace projection
ingests it client-side.

## Operational notes

Purely a rendering feature: no new event, tuning constant, or
persisted state. `internal/tui/village_lens_test.go` covers the strip's
overflow/shed behavior and both agent overlays' priority ordering;
`layout_test.go` extends the fold-cascade coverage to the three-step order;
`views_test.go`/`tui_test.go` extend existing render/update assertions.

## Spec 086 — the reverse jump (the lens's other direction)

Standing resolution 1 amended deliberately: the villager strip gains exactly
ONE mouse affordance — clicking a glyph centers the map camera on that
villager (`handleStripHitClick` → `centerCameraOn`; dead villagers jump to
their grave; the `…N` overflow marker is never a target) — still no cursor,
no keys. Its keyboard path is the villagers tab's `J`
over the SAME roster ordering. The villagers-tab roster gains the matching
click (select + jump; narrow switches the active pane to the map). Both
controls ship with mouse-parity oracle entries
(`internal/tui/mouseparity_test.go`) and corpus rows
(`docs/design/tui/panels/villager-strip.md`, `panels/villagers.md`).
