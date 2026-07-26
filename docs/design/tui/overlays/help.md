---
title: Overlay — help (`?`)
class: overlay
status: shipped
verified_against: 2d7a54940f8512340143e8ca5a8ba53e6e196aa2
sources:
  - internal/tui/help.go
  - internal/tui/tiles.go
  - internal/tui/tui.go
  - internal/tui/look.go
  - internal/worlds/unlocks.go
---

# Overlay: help

The context-sensitive help overlay (spec 045/TASK-116). Extracted from
`patterns/keymap.md` into its own overlay page (this feature, FR-005) — the
mode-key tables that page's "Mode: help overlay" section used to carry now
live here, reconciled against `internal/tui/help.go` and
`specs/045-tui-help-overlay/contracts/help-content.md`. `patterns/keymap.md`
stays the one printable reference card and only points here.

**Hybrid status, stated plainly**: `status: shipped`. Sections 1–3 shipped
with spec 045/055; **Section 4 (ceremony replay) shipped with spec 056
(TASK-127)**, **Section 5 (the guardian, D9) shipped with spec 063
(TASK-115)**, and **the badge deep-link row — previously retagged
`unbuilt (pending TASK-142, layer-2)` by spec 075's semantic-cells lint
pass — shipped with spec 074-look-cursor (TASK-142)**, and **the guardian
section's forward-ladder block — shipped with spec 078-tui-ladder-view
(TASK-152)** — every row's control-table entry below now names a real
renderer symbol; nothing on this page remains `unbuilt`.

## Mockup

```
┌ HELP · keys ──────────────────────────────────────────────┐
│ Global (home) — basic         (t: tier · n/p: mode)        │
│                                                            │
│ 2               select the chronicle tab; press again to  │
│                 solo it (or return home if already solo'd) │
│ 3               select the {{skin.guardian.tab_label}} tab; │
│                 press again to solo it (or return home)    │
│ 4               select the villagers tab; press again to   │
│                 solo it (or return home if already solo'd) │
│ m               focus the minibuffer — ask the              │
│                 {{skin.guardian.epithet}}                   │
│ space           pause / resume the clock                  │
│ q               quit                                       │
│                                                            │
│ ctrl+c — quit, from anywhere                                │
└──────────────────────────────────────────────────────────┘
 esc/? close · tab section · t tier · n/p mode · J/K scroll
```

## Layering

1. `?` opens the overlay from every mode except text-entry (minibuffer
   focused: `?` is a buffer character, per the focus contract — minibuffer
   help is content reachable from other modes' overlays instead).
2. While open, the overlay owns the keyboard entirely: its own navigation +
   dismissal only; every other key is inert — no silent fallthrough to the
   mode beneath.
3. Dismissal (`esc` or `?`, toggle) releases exactly **one** layer — the
   esc-release chain beneath (minibuffer → decisions → detail → solo → home)
   is untouched, and the overlay is never stacked a second time.
4. Opening/using/dismissing sends no IPC command, emits no event, and
   changes no world or layout state beyond the help overlay's own fields
   (`helpOpen`, `helpMode`, `helpPageMode`, `helpTier`, `helpSection`,
   `helpScroll`) — rendering is body replacement in the solo-zoom slot
   (`helpPanelView`), chrome (header/minibuffer/footer) stays visible, output
   remains exactly terminal-height.

## Sections & tiers

### Section 1 — keys (frozen at open)

The mode you were in when you pressed `?` (`m.helpMode`, frozen via
`currentHelpMode`) determines the starting page, but `n`/`p` page across
**every** mode's key table regardless of where you opened from — the only way
to read the minibuffer's page, since `?` never opens from there. Seven mode
pages total (`helpModeKey`; spec 074-look-cursor added the seventh):

| Mode page | Covers |
|---|---|
| Global (home) | `handleGlobalKey`'s dispatch — tabs, ask, pause, speed, pan, recenter, chronicle filters, quit |
| Minibuffer (focused) | `handleMinibufferKey` — send, history, release, typing |
| Inspect (paused, chronicle) | `handleInspectKey` — selection, jump, detail scroll, resume |
| Villagers (roster) | `handleVillagersKey`, roster state — selection, jump, open detail |
| Villagers (detail) | same handler, detail/decisions state — toggle decisions, back, scroll |
| Solo zoom / narrow fallback | the identical global dispatch, partitioned into a different basic/advanced split for its footer |
| Look-cursor (tile inspection) | `handleLookKey` family (`internal/tui/look.go`) — cursor move/jump/snap, pane focus/select, drill-in, exit-and-select; spec 074, checked right after the console branch in `currentHelpMode` (the mode borrows the dock, so the villagers/inspect branches would otherwise mis-route) |

Each page has a **basic** tier (≈ the footer-hinted set for that mode) and an
**advanced** tier (`t` toggles) — every real, working binding appears in
exactly one tier of its mode's page; `help_test.go`'s keymap sweep
mechanically ties every advertised row to a real handler and every handled
binding to exactly one tier (SC-003/FR-003).

### Section 2 — the screen (`tab`/`shift+tab` to reach)

- **Header anatomy** (`headerAnatomy`) — every element and conditional badge
  `headerView` can render, including ones the player may never have seen yet
  (governed speed, an `[llm: …]` condition, `[suppressed: …]`) — see
  [../pages/home.md](../pages/home.md) for the reconciled header-segment
  inventory this table mirrors.
- **Map glyphs** (`mapGlyphs`) — rendered long-form from the *same* table
  `renderMapGrid`'s compact legend line reads (`legendGlyphLine`), so the
  overlay's glyph walkthrough and the in-game legend cannot silently diverge
  — see [../panels/map.md](../panels/map.md) for the reconciled glyph
  inventory (this feature added the wall/path/quarried/pile/grave rows to
  both the map page and, by construction, this shared table). Spec 068 grew
  the table into the tile REGISTRY (`internal/tui/tiles.go` — glyph, name,
  meaning, classed style token, world binding; see map.md's "tile registry"
  section) and appended the marsh `░` / sand `▒` rows: the walkthrough and
  the compact legend picked both up from the shared table with no edit to
  this overlay's rendering, exactly the claim this seam exists to keep.
- **Dock tabs** (`dockTabs`) — key/name/purpose for each dock tab, read off
  `dockTabKey`/`paneNames` (never a second, hand-maintained list) — now four:
  chronicle/`{{skin.guardian.tab_label}}`/villagers/systems (spec 053 added
  the fourth, proving the "a future tab is a new entry here with no
  structural change" claim); a future tab (exercise, Wave 4) is a fifth with
  the same zero-structural-change property.

### Section 3 — lessons (pull reference)

`helpLesson{id, title, body}` entries render on demand. **Shipped** (spec
055, TASK-117, Wave 4): `populateHelpLessons` (`internal/tui/lessons.go`)
fills the table 1:1 from the lesson row's own catalog
([panels/lesson-row.md](../panels/lesson-row.md), decision 5) at every
client boot — the placeholder line ("lessons appear here as the village
teaches them") now renders only in the degenerate case of an empty table
(`helpLessonsLines`' own defensive branch; never true in a running client
past this feature). Exactly the content addition this seam was built for —
no structural change to this overlay's navigation or rendering (SC-006).

### Section 4 — ceremonies (pull reference, spec 056/TASK-127)

`ceremonyReplayLines` (`internal/tui/help.go`) — **shipped**. One entry per
stage `replica.StagesUnlocked` names (the durable per-world facts, never a
second event scan), each re-rendering the SAME title, D6 authorship chapter
(`skin.CeremonyChapter`), and report card (`ceremonyReportCardFor`) the live
`overlays/ceremony.md` takeover showed when the unlock actually happened —
stored, never regenerated (research R5). Degrades to an honest placeholder
line ("no stage has unlocked yet in this world") before any stage has
unlocked. This is the explicit AC `overlays/ceremony.md`'s replayability
section depends on (spec.md US2-AS2/FR-013): a player who missed or
dismissed a ceremony is never permanently denied its content.

### Section 5 — the guardian (decision 9)

**Built** (spec 063, TASK-115): `helpGuardianLines` (`internal/tui/help.go`).
A fourth overlay section (alongside keys/screen/lessons, `tab`/`shift+tab`
cycles to it like the other three; its panel title resolves the skin
epithet via `helpSectionLabel`), deliberately teaching prompting at the
no-LLM floor: **static-per-stage, model-free** content — never generated by
a model, never read from live status beyond the one stage value that
selects which static page to show.

**Content, per stage** (`world.StagesLadder`, `internal/world` — the
skin-independent ladder content this section reads, RELOCATED there from
`cmd/promptworld/stages.go` by spec 063 so `promptworld stages` and this
section render from one source):

1. **Stage identity/concept** — the skin-resolved identity name
   (`skin.Stage`) and the one prompt-engineering concept that stage teaches
   (conversational prompting → instruction authoring → capability design →
   mastery), plus the ladder's grants line.
2. **Granted verbs** — the tool names the stage ceiling grants, read
   through `guardian.StageCeilingVerbs` — the SAME
   `applyStageCeiling`-over-full-grant intersection the turn's roster runs
   ([[curriculum-ladder]]), never a second, hand-maintained tool list. The
   stage-1/2 ceiling gained `explain` (spec 063 — the tutor stage's
   grounding tool).
3. **One example ask per verb** — a sample player phrasing per granted
   verb, resolved through the per-verb example-ask token family (keyed by
   the frozen tool id, e.g. `skin.guardian.example_ask.send_vision` —
   [../patterns/skin-tokens.md](../patterns/skin-tokens.md)) — the
   teaching payload this section exists for: the whole game is
   prompt-engineering the guardian, so the help overlay's deterministic
   floor must include how to ask, not just what exists.

**Rendering rule**: a pure function of the world's `Status.Stage` value
(plus the boot-frozen world skin) — for a given stage the bytes are
constant, exactly like every other section on this page
(`TestHelpGuardianByteIdenticalPerStatus`, SC-005). **Nil status renders
the pre-ladder variant** (all verbs, no lock, matching a pre-ladder world's
"ungated, stage-4 semantics" posture, [[curriculum-ladder]]) — never a
blank or an error. This section is **never LLM-derived**: unlike the
guardian tab's live transcript, this section reads only the static
`world.StagesLadder` table plus the one polled `Status.Stage` value that
selects a row from it.

**The ladder — shipped (reorient decision 6, spec 078/TASK-152)**:
`helpLadderLines` (`internal/tui/help.go`) appends a forward-looking block
below the current-stage teaching content above — the WorldBox
discoverability gap this closes: a TUI-only player previously had no
in-game way to see what's next or how to earn it. It iterates
`world.StageOrder` (never a hardcoded stage list/count) and renders, per
stage, the same fields `promptworld stages` prints — the skin-resolved
identity, the concept it teaches, the unlock evidence (graduation wording
at the stage whose `UnlockEvidence` is empty), and its earned/next state.
`next` marks exactly the first stage the earned rule hasn't cleared yet — a
pure derivation, not a new field `stages --json` carries.

The earned rule is single-sourced (spec 063 T014's one-source-two-surfaces
precedent, extended from catalog content to earned state):
`worlds.(*Unlocks).StageEarned` — stage-1's unconditional floor, or a
per-user unlocks-record entry — RELOCATED from `cmd/promptworld/stages.go`'s
former package-main `stageEarned` so `cmdStages` and this block render from
one substrate. A record-earned stage additionally carries the audit pointer
(proving world + exercise) exactly when `stages --json` would populate
`proving_world`/`exercise`; a stage latched in the live
`replica.StagesUnlocked` (a mid-session unlock, before or regardless of
whether the record write lands) shows earned immediately with no audit
pointer yet — the record is the only place that pointer exists.

The attached world's current stage gets a you-are-here marker
(`◀ this world`); a `stage_overridden` world states the override honestly
on that marker (`by override — not earned`) without ever laundering it
into an earned claim — the row's earned state stays record-derived
regardless (the `stages` command's own honesty posture).

Unlike the stage-keyed block above (byte-identical per `Status.Stage`
scalar), the ladder is **per-user forward content, not per-connection**: it
renders identically with nil status (no you-are-here marker, ladder still
shown) and updates live the moment a stage unlocks mid-session, with zero
per-frame disk reads — the unlocks record loads once at client boot
(`Model.unlocks`, the `populateHelpLessons`/`LoadLessonsSeen` precedent) and
is unioned with `replica.StagesUnlocked` only at render time. No LLM call,
no IPC command, no event emission — see "Byte-identity classification"
below for its own row.

**The deliberate spec-045 amendment (D9)**: spec 045 originally forbade any
status-derived content in the help overlay ("never derived from live
status"). This section is the deliberate, named exception the reorientation
authorizes: reading `Status.Stage` to select which STATIC page to show does
not reintroduce the hazard the original invariant guarded against (no LLM
call, no world-state dependency beyond one scalar) — see "Byte-identity
classification" below for the precise restatement of what the no-LLM floor
guarantee still covers.

**Badge deep-link — shipped (spec 074-look-cursor, FR-011)**: `?` opening
pre-focused on the active badge's row (e.g. opening help from a screen
showing `[llm: …]` scrolls Section 2 straight to that badge's
`headerAnatomy` row) — `openHelp`'s pre-focus step (`internal/tui/tui.go`):
with ≥1 active conditional badge (`[degraded]`, `[llm: …]`, `[suppressed:
…]`, checked in that header order — the first active one wins), `helpSection`
is set to the screen-walkthrough section and `helpScroll` to that badge's
`headerAnatomy` row index (`firstActiveBadgeRow`, resolved from the SAME
shared table `helpWalkthroughLines` renders, so the scroll target can never
drift from the rendered row). No active badge → the open is byte-identical
to before this feature (keys section, scroll 0). It applies corpus-wide (any
mode, any active badge), not only to the guardian section.

## Byte-identity classification (research.md R4)

The Wave 0 ruling (c), restated verbatim from research.md — every section of
this overlay, classified:

| Section | Class |
|---|---|
| keys (tiered) | **byte-identical** with nil status (generated from the static keymap; the existing `help_test.go` sweep guarantee) |
| screen walkthrough | **byte-identical** with nil status |
| the guardian (D9, above) | **stage-keyed, model-free**: content is a pure function of the stage value; for a given stage the bytes are constant; nil status renders the pre-ladder variant (all verbs). Never LLM-derived. |
| forward ladder (guardian section, above) | **unlocks-record-derived, model-free**: catalog text (identity/concept/evidence) is static per stage; earned/next columns come from the per-user unlocks record union'd with the live `replica.StagesUnlocked`; you-are-here comes from the stage/override scalars. Never LLM-derived; an absent/corrupt/unresolvable record renders the stage-1-floor-only ladder, never an error — **shipped** (reorient D6/spec 078, TASK-152) |
| lessons registry | **status-derived** (active/seen state per user); the registry's catalog text is static, its state columns are live |
| badge deep-link focus | **status-derived** (which row is pre-focused depends on active badges); content unchanged — **shipped** (spec 074-look-cursor/TASK-142) |
| ceremony replay entries | **status-derived** (which ceremonies exist depends on run history — `replica.StagesUnlocked`; replayed content is stored, not regenerated) — **shipped** (spec 056/TASK-127) |

**The no-LLM floor guarantee, restated**: with nil status AND no LLM
configured, the overlay renders the keys, walkthrough, and pre-ladder
guardian sections byte-identically on every invocation — spec 045's
contract, deliberately amended (per D9) from "never derived from live
status" to "the floor set is byte-identical; status-derived sections
degrade to their static catalog with state columns empty." This preserves
what spec 045's invariant actually protects (deterministic, model-free help
always available) while admitting the reorientation's dynamic additions
under an explicit classification instead of eroding the invariant silently.

## Control table

| control/region | states | data source | renderer | keys+mouse | introduced-by | skin-token |
|---|---|---|---|---|---|---|
| open/dismiss overlay | closed · open | `Model.helpOpen` | `openHelp`/`closeHelp` | `?` (open, any non-text mode) · `esc`/`?` (dismiss) · — | spec 045 | — |
| section cycle | keys · screen · lessons · the guardian | `Model.helpSection` | `helpPanelView` (title via `helpSectionLabel`) | `tab`/`shift+tab` · — | spec 045; guardian section spec 063 | `skin.guardian.epithet` (guardian section title) |
| tier toggle (keys section) | basic · advanced | `Model.helpTier` | `helpKeysLines` | `t` · — | spec 045 | — |
| mode paging (keys section) | 7 mode pages | `Model.helpPageMode` | `helpKeysLines`, `nextHelpMode`/`prevHelpMode` | `n`/`p` · — | spec 045; 7th (look-cursor) page spec 074 | — |
| pager scroll | — | `Model.helpScroll` | `paginateHelpContent` | `J`/`K` · — | spec 045 | — |
| header anatomy row | static | `headerAnatomy` | `helpWalkthroughLines` | — (display-only) | spec 045 | — |
| map glyph row | static | `mapGlyphs` (shared with `legendGlyphLine`) | `helpWalkthroughLines` | — | spec 045 | — |
| dock tab row | static | `dockTabs` | `helpWalkthroughLines` | — | spec 045 | — |
| lessons pull-reference entry | populated (12 catalog entries, spec 077 tranche 2) · empty (placeholder line, degenerate only) | `helpLessons`, populated from `lessonCatalog` (`populateHelpLessons`) | `helpLessonsLines` | — | spec 045 (seam); content spec 055/TASK-117 | — |
| the guardian section (stage identity/concept) | per-stage · pre-ladder (nil status) | `Status.Stage`, `world.StagesLadder`, `skin.Stage` | `helpGuardianLines` | — (display-only) | reorient D9 / spec 063 | `skin.stage.stage-N.name`/`.line` |
| the guardian section (granted verbs) | per-stage ceiling | `guardian.StageCeilingVerbs` (the turn grant's own `applyStageCeiling` intersection) | `helpGuardianLines` | — | reorient D9 / spec 063 | — |
| the guardian section (example ask per verb) | static, per verb | the per-verb example-ask token family (skin-tokens.md) | `helpGuardianLines` | — | reorient D9 / spec 063 | `skin.guardian.example_ask.send_vision` (one per verb, keyed by tool id) |
| the guardian section (forward ladder) | per-stage identity/concept/evidence (static) · earned/next/audit-pointer (per-user record ∪ live `StagesUnlocked`) · you-are-here (stage/override scalars) | `world.StageOrder`/`StagesLadder`, `Model.unlocks` (`worlds.Unlocks.StageEarned`), `replica.StagesUnlocked`, `skin.Stage` | `helpLadderLines` | — (display-only) | reorient D6 / spec 078 | `skin.stage.stage-N.name`/`.line` |
| badge deep-link focus (layer-2) | unfocused · pre-focused on active badge | active header badge at open | `openHelp`/`firstActiveBadgeRow` (`internal/tui/tui.go`) | — (display-only; the badge itself has no click target) | spec 074-look-cursor (retained per D9 discussion, shipped here) | — |
| look-cursor keys page | basic · advanced | `handleLookKey` dispatch (`look.go`) | `helpKeysLines` (page 7) | `n`/`p` to reach (`?` mid-mode freezes on it) · — | spec 074 | — |
| ceremony replay entries | none · N replayable | `replica.StagesUnlocked`/`CurriculumPasses` | `ceremonyReplayLines` (`internal/tui/help.go`), shared rendering with `overlays/ceremony.md` | `tab`/`shift+tab` to reach · — | reorient FR-013 | — |

**Parity rollout**: every control above has a key but no mouse target today;
tracked here rather than omitted (decision 8, formal doctrine in
`patterns/keymap.md`, T024).
