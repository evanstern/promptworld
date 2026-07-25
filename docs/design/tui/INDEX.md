---
title: TUI design reference — the living UI authority
class: index
status: shipped
verified_against: c8906da39be3a5b861c2272af37db0a83dcded7a
---

# TUI design reference v2

## Authority statement

`docs/design/tui/` is the living, page-by-page, control-by-control authority
for the promptworld TUI (`internal/tui`, the Bubble Tea client). An
implementer — human or AI — assigned a change to any shipped surface, or
picking up any reorientation wave, should be able to build or modify the
screen from these files alone, without re-deriving intent from Go source.
Where these files and the running client disagree, that is a bug in *this
corpus* (or an unamended PR) — file the fix here in the same change that
touches `internal/tui/`.

This v2 rebuild (spec 047, reorientation 2026-07-25 decision 4) replaces the
TASK-34 single-directory reference: it reconciles the corpus with everything
shipped since (specs 013–046), splits the monolithic dock tab along the D10
skin boundary, and adds the taxonomy, control-table, and freshness-gate
machinery below. Nothing about the region anatomy or the reuse-the-renderers
ground rules changes — v2 is reconciliation and extension, not a redesign.

## v2 taxonomy

Four classes, plus two top-level index files:

- **`pages/`** — what fills the whole terminal (or the whole solo-view slot).
- **`panels/`** — one file per dock tab, plus each standalone strip/row that
  isn't a dock tab (lesson row, guardian strip, villager strip, minibuffer).
- **`overlays/`** — takeover surfaces that replace the body in place (help,
  ceremony, postmortem) — same body-replacement slot as solo zoom, never
  stacking.
- **`patterns/`** — cross-cutting rules that apply wherever they're relevant
  (focus, chronicle grammar, keymap, layout, skin tokens, stage defaults).
- **[`anatomy.md`](anatomy.md)** — the region index: every visible element →
  its owning file, both directions complete.
- **`INDEX.md`** (this file) — authority statement, taxonomy, gate rules.

## File map

Every file the finished feature lands, per plan.md's Project Structure.
`[recon]` = existing v1 content reconciled against shipped reality in this
feature; `[new]` = authored fresh in this feature (spec 047's ten new-surface
pages plus the taxonomy/gate machinery); dates/waves in parentheses note when
a `[new]` page's *surface* actually ships in `internal/tui`, per its own
`status:` frontmatter field.

```text
docs/design/tui/
├── INDEX.md                         [recon] this file
├── anatomy.md                       [new]   region index — every visible element → owning file
├── pages/
│   ├── home.md                      [recon] widescreen composite, stage-shaped defaults
│   ├── guardian-console.md          [new]   full-height guardian page (decisions 1/2, D5 — Wave 3)
│   └── solo-views.md                [recon] full-width zoom + narrow fallback
├── panels/
│   ├── map.md                       [recon] terrain camera viewport
│   ├── chronicle.md                 [recon] feed panel: running scroll + paused inspect
│   ├── dock.md                      [recon] tab-container chrome ONLY (tabs, badges, solo-zoom seam)
│   ├── guardian.md                  [new]   fiction-layer tab content, split from dock.md (D10)
│   ├── systems.md                   [new]   telemetry tab content, split from dock.md (D10)
│   ├── villagers.md                 [new]   villagers tab content, split from dock.md
│   ├── exercise.md                  [new]   scenario exercise panel (D11, D4 — Wave 4)
│   ├── lesson-row.md                [new]   first-occurrence lesson row (decision 5 — Wave 4)
│   ├── guardian-strip.md            [new]   action-budget strip above minibuffer (decision 7 — Wave 2)
│   ├── villager-strip.md            [new]   colonist-bar strip under header (D12 — Wave 5)
│   └── minibuffer.md                [recon] the guardian input line and its states
├── overlays/
│   ├── help.md                      [new]   extracted from patterns/keymap.md + guardian section (D9 — Wave 4)
│   ├── ceremony.md                  [new]   unlock takeover (decision 6, FR-019 — Wave 4)
│   └── postmortem.md                [new]   run-end takeover (decision 6, FR-018 — Wave 4)
└── patterns/
    ├── focus-contract.md            [recon] who owns the keyboard, when
    ├── chronicle-grammar.md         [recon] event line format + JSON inspector
    ├── keymap.md                    [recon] every key, every mode; input-parity doctrine (decision 8)
    ├── layout.md                    [recon] breakpoints, row/column budget, fold order (rulings a/b)
    ├── skin-tokens.md               [new]   skin-token conventions + the spec 052 runtime contract (doc twin)
    └── stage-defaults.md            [new]   stage-resolved visibility defaults (decision 3 — Wave 4)

scripts/
└── check-tui-design.mjs             [new]   structural + same-PR gate (Wave 0 mechanization)
```

## Gate rules

1. **Same-PR amendment (extends TASK-34's old rule 4 from "record deviations"
   to "any change"):** a PR touching `internal/tui/` MUST touch this reference
   in the same PR — re-verify every affected page, amend what changed
   (including a pure refactor with zero visible change: re-verify and bump the
   affected pages' pins), and record any deviation forced by implementation
   reality. The convention alone already failed twice (specs 044 and 046
   shipped TUI changes with no doc amendment) — this repo has no CI, so the
   check script below is the mechanized version of this rule.
2. **Run the check before any PR touching `internal/tui/`:**
   ```bash
   node scripts/check-tui-design.mjs --changed
   ```
   Validates: every file's taxonomy placement, every `verified_against` pin,
   every panel/overlay page's canonical control-table header
   (contracts/control-table.md), `anatomy.md` completeness (every reference
   file reachable, no unmapped visible element), and the same-PR range check
   itself. A doc-only changeset (no `internal/tui/` touched) always passes the
   last check trivially.
3. **New surfaces are authored here spec-before-build.** A reorientation wave
   implementing one of the ten new-surface pages amends that page in the same
   PR that ships the surface: flips `status: specified` → `shipped`, fills in
   real renderer symbols, re-pins `verified_against`.
4. **Mockups are representative, not screenshots.** The check script verifies
   pins and structure, never pixel content; semantic drift between a mockup
   and the real rendering is a same-PR review responsibility, not something
   the script can catch.

## FR-020 audience ruling (corpus-wide convention)

Control tables' `data source` column may name raw registry/engine values —
that column is engineer-facing (contracts/control-table.md). Every
**player-facing** projection of the same information — a control table's
plain-language callout prose, the help overlay's advanced tier, any
mockup-adjacent explanation — stays **plain-language by default**; raw
registry values surface to a player only behind an explicit debug/inspector
toggle (a *mode*, not a tier). This is an operator ruling (2026-07-25, closing
reorientation open question 4 / FR-020) and applies corpus-wide: every page
authored or reconciled in this feature follows it, not just the pages that
happen to name it explicitly.

## History: TASK-34 decision record

Preserved verbatim from the v1 `INDEX.md` (the original widescreen-layout
spike) — this is *why* the five-region anatomy exists at all, and v2 changes
none of it:

> **Chosen direction: B + C hybrid** (from the TASK-34 spike mockups).
>
> - **B — tabbed dock:** the right side of the widescreen composite is a
>   single dock with tabs (chronicle · metatron · villagers). One tab visible
>   at a time; the dock is the extension point for future displays/controls.
> - **C — Metatron minibuffer:** Metatron's *input* leaves the pane system
>   entirely and becomes a one-line minibuffer above the footer — the only
>   text input in the app. Angel replies land in the dock's **metatron** tab
>   (this resolves C's open question about where replies live).
> - Option A (stacked right rail) was considered and rejected: its three-way
>   rail split starves the chronicle of rows and its docked one-line Metatron
>   loses history.
>
> Visual mockups from the spike:
> https://claude.ai/code/artifact/dfb04194-b379-4733-a586-9882b5e0746e
> (exploratory; where it disagrees with these files, these files win.)

Every widescreen frame is composed of exactly five regions (unchanged by v2;
the reorientation's new chrome — lesson row, guardian strip, villager strip —
is additive on top, per `patterns/layout.md`'s re-derived row budget):

```
┌─ header ─ promptworld · day 4 · 08:12 · 1× · [PAUSED] ──────────────────┐
│ ┌─ MAP ──────────────────────────────┐ ┌─ DOCK ───────────────────────┐  │
│ │                                    │ │ chronicle │ metatron │ villagers│  │
│ │  camera viewport over terrain      │ ├──────────────────────────────┤  │
│ │  (existing renderer, resized)      │ │                              │  │
│ │                                    │ │  active tab content          │  │
│ │                                    │ │                              │  │
│ └────────────────────────────────────┘ └──────────────────────────────┘  │
│ ┌─ METATRON minibuffer (dormant: 1 dim line · focused: amber border) ─┐  │
│ └──────────────────────────────────────────────────────────────────────┘ │
└─ footer ─ key hints ──────────────────────────────────────────────────────┘
```

## Ground rules for the implementer

1. These docs specify *behavior and composition*, not Go structure. Reuse the
   existing renderers (`internal/tui/views.go`, `help.go`) wherever a panel
   says "content unchanged" or names a real renderer symbol.
2. The narrow-terminal fallback preserves today's single-pane UI — never
   delete it; every new chrome element states its own narrow behavior
   (`patterns/layout.md` ruling b).
3. Mockup content (agent names, event payloads) is representative; bind to
   the real event types in `internal/sim` and names from the replica. Fiction
   strings render as skin tokens (`patterns/skin-tokens.md`), never bare
   literals.
4. Any deviation forced by implementation reality gets recorded back into
   these files in the same PR — the gate rules above make this mechanical,
   not just a convention.
