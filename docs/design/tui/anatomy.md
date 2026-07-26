---
title: Anatomy — region index
class: index
status: shipped
verified_against: 9905c70f326d027a34425081442270bdb80648f7
---

# Anatomy: the region index

Every visible region, strip, badge, and chrome row the shipped client (or a
spec-before-build reorientation surface) can render, mapped to exactly one
owning file. Start here (or `INDEX.md`) and reach any control's authoritative
page in at most 2 hops (`anatomy.md` → page, or `INDEX.md` → `anatomy.md` →
page).

**Completeness**: every file this feature's Project Structure names now
exists (spec 047 landed in two slices on one task branch, TASK-123 — one
PR); the "every `pages/`/`panels`/`overlays` file is the target of ≥1 row"
invariant (data-model.md) holds for the whole table below, not a subset of
it. **Existing** and **built** are different axes, though: several rows
below name a page whose `status:` frontmatter is `specified` — the design
doc is real and complete (mockup, control table, stage defaults, linear
projection), but the surface itself has no `internal/tui` code yet. Those
rows are marked **(specified)** rather than omitted or bracketed as
forthcoming.

## Header region

| Region | Owning file | Notes |
|---|---|---|
| world name / disconnected-retry state | `pages/home.md` | always shown |
| clock state segment (running · PAUSED · ENDED) | `pages/home.md` | `ENDED` outranks `PAUSED` (spec 044); read-only postmortem posture |
| speed segment + governed-speed suffix | `pages/home.md` | specs 028/039; teaching soft-cap rides this same surface |
| `[degraded]` badge | `pages/home.md` | pre-existing |
| `[llm: provider kind]` badge | `pages/home.md` | spec 034 |
| `[suppressed: classes]` badge | `pages/home.md` | spec 037 |
| villager strip (one row under header) | `panels/villager-strip.md` **(specified)** | D12; widescreen default-on all stages (`patterns/stage-defaults.md`); narrow folds to a header count badge (`patterns/layout.md` ruling b) |

## Map region

| Region | Owning file | Notes |
|---|---|---|
| terrain tile (water/wood/forage/rock/quarried/path/plain) | `panels/map.md` | night-dimmed (`.Faint`) uniformly |
| agent glyph (awake/asleep/dead/dead-on-grave) | `panels/map.md` | spec 044 dead-agent-on-grave carve-out |
| structure glyphs (fire, shelter, oven, chest, wall, grave) | `panels/map.md` | specs 012/013/032/044 |
| pile overlay | `panels/map.md` | spec 013 US2 |
| gru glyph | `panels/map.md` | highest render priority |
| camera pan / recenter | `panels/map.md` | arrow keys, `c` |
| legend / inspection line (glyph key + in-view piles/chests) | `panels/map.md` | spec 013 T021/T026; first row shed when height is scarce |
| map condition overlay | `panels/map.md` | Wave 5 stub row, unbuilt |

## Dock region (tab container + per-tab content)

| Region | Owning file | Notes |
|---|---|---|
| tab row, tab-switch keys, solo-zoom seam | `panels/dock.md` | container chrome only (D10 split) |
| unseen-reply badge dot | `panels/dock.md` | guardian tab only today |
| chronicle tab (default) — running feed + paused inspect | `panels/chronicle.md` | spec 018/TASK-60; line format governed by `patterns/chronicle-grammar.md` |
| guardian tab — pane header (name, charge bank, instruction/capability provenance, stage segment) | `panels/guardian.md` | specs 021/046 |
| guardian tab — transcript (you/guardian rows, `⚡`/`👁`/`⏲`/`»` special rows) | `panels/guardian.md` | specs 016/029/020 |
| guardian tab — standing-orders block | `panels/guardian.md` | spec 029 |
| systems tab — provider table + health rows | `panels/systems.md` **(hybrid: content shipped, tab unbuilt)** | specs 024/034 |
| systems tab — cognition horizon block | `panels/systems.md` **(hybrid)** | spec 037; calibration remedy voice spec 035 |
| villagers tab — roster | `panels/villagers.md` | spec 015 |
| villagers tab — detail (identity/objective/inventory/beliefs/memories) | `panels/villagers.md` | spec 015 |
| villagers tab — decisions sub-view | `panels/villagers.md` | spec 020 |
| exercise tab (framing, attach briefing, rubric gauges, forecast/fog vocabulary, pass/fail) | `panels/exercise.md` | spec 054 (D11/D4), key `6`; world-shaped not stage-shaped (`patterns/stage-defaults.md`) |

## Teaching-chrome region (below the body, above the minibuffer)

| Region | Owning file | Notes |
|---|---|---|
| lesson row (one active lesson, ≤2 lines) | `panels/lesson-row.md` **(specified)** | decision 5; on at stages 1–2, badge+overlay at 3+/pre-ladder (`patterns/stage-defaults.md`); folds 3rd (`patterns/layout.md`) |
| guardian strip (charge bank · regen · order count) | `panels/guardian-strip.md` | decision 7; spec 050 shipped (faith reserved, unbuilt pending TASK-118); always-on all stages; folds LAST (relocates into the minibuffer's dormant line, `patterns/layout.md`) |

## Minibuffer region

| Region | Owning file | Notes |
|---|---|---|
| minibuffer box (dormant · focused · busy · flash) | `panels/minibuffer.md` | the only text input; focus contract governs it (`patterns/focus-contract.md`); also the guardian console's composer (`pages/guardian-console.md`) |

## Footer region

| Region | Owning file | Notes |
|---|---|---|
| footer key-hint line (per-mode) | `patterns/keymap.md` | "Footer hints per mode"; overlay's own hint replaces it while help is open |

## Pages (whole-terminal / whole-solo-slot surfaces)

| Region | Owning file | Notes |
|---|---|---|
| widescreen composite (header+map+dock+minibuffer+footer) | `pages/home.md` | resting state; always "underneath" |
| solo zoom (any dock tab at full width) | `pages/solo-views.md` | same component, two widths |
| narrow fallback (single-pane, keys `1`–`5`, plus `6` on scenario worlds) | `pages/solo-views.md` | never deleted; new-chrome narrow rules per `patterns/layout.md` |
| guardian console (full-height guardian page) | `pages/guardian-console.md` **(specified)** | decisions 1/2, D5; not reached via solo zoom — its own `G` key |

## Overlays (takeover, body-replacement)

| Region | Owning file | Notes |
|---|---|---|
| help overlay — keys section (6 mode pages, 2 tiers) | `overlays/help.md` | spec 045 |
| help overlay — the screen section (header anatomy, map glyphs, dock tabs) | `overlays/help.md` | spec 045; anti-drift shared tables |
| help overlay — lessons pull-reference | `overlays/help.md` | seam, ships empty |
| help overlay — the guardian section (Section 4) | `overlays/help.md` **(specified)** | D9; stage identity/concept, granted verbs, example asks |
| unlock ceremony takeover | `overlays/ceremony.md` **(specified)** | decision 6, FR-019; replayable from `?`/`stages`; postmortem wins on conflict |
| postmortem takeover | `overlays/postmortem.md` **(specified)** | decision 6, FR-018; replayable from the morgue; always wins over ceremony |

## Both-directions check

Every file listed above is the target of at least one row (the invariant
this page enforces); every row names exactly one owning file (no visible
element is split across two pages — the guardian/systems split is a content
boundary at the tab level, not a shared element; the report-card renderer
is one spec shared by three call sites — `overlays/postmortem.md`,
`pages/guardian-console.md`, `overlays/ceremony.md` — authored once on
`overlays/postmortem.md` and referenced, not duplicated, from the other
two). `patterns/*.md` files are cross-cutting rules rather than owners of a
single visible region and are referenced from the notes column above where
relevant (`patterns/layout.md` for fold order, `patterns/focus-contract.md`
for the minibuffer's focus chrome, `patterns/chronicle-grammar.md` for
feed-line format, `patterns/skin-tokens.md` for every `{{skin.…}}`
placeholder above, `patterns/stage-defaults.md` for the authoritative
stage-visibility table); they are not required to be — and are not —
anatomy targets themselves (data-model.md scopes the completeness
invariant to `pages/`/`panels/`/`overlays/` files).
