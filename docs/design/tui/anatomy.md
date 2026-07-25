---
title: Anatomy — region index
class: index
status: shipped
verified_against: c8906da39be3a5b861c2272af37db0a83dcded7a
---

# Anatomy: the region index

Every visible region, strip, badge, and chrome row the shipped client (or a
spec-before-build reorientation surface) can render, mapped to exactly one
owning file. Start here (or `INDEX.md`) and reach any control's authoritative
page in at most 2 hops (`anatomy.md` → page, or `INDEX.md` → `anatomy.md` →
page).

**Completeness note (transient, honest about this slice's scope)**: this
feature lands in phases on one task branch (spec 047, TASK-123 — one PR).
This slice authors the taxonomy, gate conventions, and the full US1
reconciliation (rows 1–ended at "Minibuffer" below); the ten new-surface
pages are authored in a later phase of the *same* feature. Rows below that
name a Phase-4 file mark it `[Phase 4]` — that file does not exist on disk
yet in this slice, so the "every file is the target of ≥1 row" completeness
invariant (data-model.md) holds for the **shipped-reconciliation** rows now
and for the **whole table** once Phase 4 lands, not as two separate
promises: every row that exists today already points at a real file.

## Header region

| Region | Owning file | Notes |
|---|---|---|
| world name / disconnected-retry state | `pages/home.md` | always shown |
| clock state segment (running · PAUSED · ENDED) | `pages/home.md` | `ENDED` outranks `PAUSED` (spec 044); read-only postmortem posture |
| speed segment + governed-speed suffix | `pages/home.md` | specs 028/039; teaching soft-cap rides this same surface |
| `[degraded]` badge | `pages/home.md` | pre-existing |
| `[llm: provider kind]` badge | `pages/home.md` | spec 034 |
| `[suppressed: classes]` badge | `pages/home.md` | spec 037 |
| villager strip (one row under header) | `panels/villager-strip.md` **[Phase 4]** | D12; widescreen default-on all stages; narrow folds to a header count badge (`patterns/layout.md` ruling b, once authored) |

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
| guardian tab — pane header (name, charge bank, instruction/capability provenance, stage segment) | `panels/guardian.md` | specs 021/046 |
| guardian tab — transcript (you/guardian rows, `⚡`/`👁`/`⏲`/`»` special rows) | `panels/guardian.md` | specs 016/029/020 |
| guardian tab — standing-orders block | `panels/guardian.md` | spec 029 |
| systems tab — provider table + health rows | `panels/systems.md` **(hybrid: content shipped, tab unbuilt)** | specs 024/034 |
| systems tab — cognition horizon block | `panels/systems.md` **(hybrid)** | spec 037; calibration remedy voice spec 035 |
| villagers tab — roster | `panels/villagers.md` | spec 015 |
| villagers tab — detail (identity/objective/inventory/beliefs/memories) | `panels/villagers.md` | spec 015 |
| villagers tab — decisions sub-view | `panels/villagers.md` | spec 020 |
| exercise tab (scenario framing, rubric gauges, forecast/fog) | `panels/exercise.md` **[Phase 4]** | D11/D4; Wave 4 |

## Teaching-chrome region (below the body, above the minibuffer)

| Region | Owning file | Notes |
|---|---|---|
| lesson row (one active lesson, ≤2 lines) | `panels/lesson-row.md` **[Phase 4]** | decision 5; on at stages 1–2, badge+overlay at 3+/pre-ladder |
| guardian strip (charge bank · regen · order count · faith) | `panels/guardian-strip.md` **[Phase 4]** | decision 7; always-on all stages; folds LAST (relocates into the minibuffer's dormant line) |

## Minibuffer region

| Region | Owning file | Notes |
|---|---|---|
| minibuffer box (dormant · focused · busy · flash) | `panels/minibuffer.md` | the only text input; focus contract governs it (`patterns/focus-contract.md`) |

## Footer region

| Region | Owning file | Notes |
|---|---|---|
| footer key-hint line (per-mode) | `patterns/keymap.md` | "Footer hints per mode"; overlay's own hint replaces it while help is open |

## Pages (whole-terminal / whole-solo-slot surfaces)

| Region | Owning file | Notes |
|---|---|---|
| widescreen composite (header+map+dock+minibuffer+footer) | `pages/home.md` | resting state; always "underneath" |
| solo zoom (any dock tab at full width) | `pages/solo-views.md` | same component, two widths |
| narrow fallback (single-pane, keys `1`–`4`) | `pages/solo-views.md` | never deleted; new-chrome narrow rules per `patterns/layout.md` (Phase 4) |
| guardian console (full-height guardian page) | `pages/guardian-console.md` **[Phase 4]** | decisions 1/2, D5; not reached via solo zoom — its own navigation |

## Overlays (takeover, body-replacement)

| Region | Owning file | Notes |
|---|---|---|
| help overlay — keys section (6 mode pages, 2 tiers) | `overlays/help.md` | spec 045 |
| help overlay — the screen section (header anatomy, map glyphs, dock tabs) | `overlays/help.md` | spec 045; anti-drift shared tables |
| help overlay — lessons pull-reference | `overlays/help.md` | seam, ships empty |
| help overlay — the guardian section | `overlays/help.md` **[Phase 4 — T018]** | D9; placeholder heading present today |
| unlock ceremony takeover | `overlays/ceremony.md` **[Phase 4]** | decision 6, FR-019; replayable from `?`/`stages` |
| postmortem takeover | `overlays/postmortem.md` **[Phase 4]** | decision 6, FR-018; replayable from the morgue |

## Both-directions check

Every file listed above is the target of at least one row (the invariant
this page enforces); every row names exactly one owning file (no visible
element is split across two pages — the guardian/systems split is a content
boundary at the tab level, not a shared element). `patterns/*.md` files are
cross-cutting rules rather than owners of a single visible region and are
referenced from the notes column above where relevant (`patterns/layout.md`
for fold order, `patterns/focus-contract.md` for the minibuffer's focus
chrome, `patterns/chronicle-grammar.md` for feed-line format,
`patterns/skin-tokens.md` for every `{{skin.…}}` placeholder above,
`patterns/stage-defaults.md` for the Phase-4 stage-visibility table); they
are not required to be — and are not — anatomy targets themselves
(data-model.md scopes the completeness invariant to `pages/`/`panels/`/
`overlays/` files).
