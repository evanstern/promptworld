---
title: Analysis — UI Reference and Help Stack
aliases: [Design-Doc v2 Structure, Help-Surface Layering]
tags: [analysis, player-docs, teaching-game, ui-reference]
type: analysis
created: 2026-07-25
updated: 2026-07-25
related: ["[[Game-Player-Docs]]", "[[Quickstart-Guide-Patterns]]", "[[Explaining-The-Screen]]", "[[In-Game-Contextual-Help]]", "[[Manual-Structure-Conventions]]", "[[Analysis-In-Game-First-Teaching]]"]
---

# Analysis — UI Reference and Help Stack

_The UI/UX projection of this branch under the reorientation lens of 2026-07-25
(`docs/design/reorient-2026-07-25-ui.md`): promptworld as a staged prompting-skills
teaching game whose UI must (a) make the village legible at a glance, (b) make prompting
the guardian the central verb, (c) teach itself in play, and (d) be fully specifiable —
page by page, control by control. [[Analysis-In-Game-First-Teaching]] is decided context
(in-game primary, out-of-game reference; the eight learning-game decisions); this note
builds the two layers that analysis did not cover: the in-TUI help-stack layering, and
the structure of the UI design reference itself._

## Constraints taken as given

The operator's reorientation decisions (recorded in
`docs/design/reorient-2026-07-25-ui.md`, steering round 2026-07-25) are fixed:

1. Headline direction **Staged Cockpit** — guardian console as a first-class page,
   telemetry split to a systems tab, village-lens legibility work, stage-shaped layout
   *defaults* (everything reachable at every stage; capability locks stay angel-only).
2. Charter/skills authoring: in-TUI read + `$EDITOR` write; no in-TUI editor.
3. Stage shapes TUI **defaults only**.
4. **UI authority: `docs/design/tui/` v2 with a freshness gate** — the living
   page-by-page reference (pages/ · panels/ · overlays/ · patterns/, uniform control
   tables with skin-token columns, `verified_against` pins, check script + same-PR gate).
5. First-occurrence lessons render in a **dedicated lesson row** above the minibuffer —
   one active lesson, ≤2 lines, dwell-until-done, anti-spam; default-on at stages 1–2,
   badge+overlay-only at stage 3+/pre-ladder.
6. **Both big moments seize the screen**: stage-unlock ceremony on
   `curriculum.stage_unlocked`, postmortem takeover on `run.ended`; both dismissable and
   replayable (`?`, `stages`, morgue).
7. Guardian strip (charge/regen/orders/faith) always visible above the minibuffer.
8. Input parity (keyboard + mouse) ratified as doctrine, rolled out incrementally.

Plus the lead-adopted defaults, of which these bind this branch directly: **D2**
(skin-token contract ships before TASK-115/117 write new fiction literals; the design
doc renders fiction strings as tokens), **D5** (report card renders as a guardian-console
card at natural stopping points), **D8** (seen-lessons state per-user), **D9** (the `?`
overlay gains a "the guardian" section as static-per-stage, model-free content; spec
045's content contract amended deliberately), **D10** (telemetry split), **D11**
(exercise panel + attach briefing), **D12** (villager strip), **D13** (`q`-detach is the
blessed stopping point).

These resolve four questions the round-2 report left open: design-doc authority →
option (a) with the gate (decision 4); guardian content in `?` → static-per-stage
section (D9); skin tokens now vs sweep later → tokens first (D2); seen-lessons home →
per-user (D8).

## Verdict

Under the corpus's four-layer hierarchy ([[In-Game-Contextual-Help]]: reduce need →
context help → bounded tutorial → tiered reference), promptworld's stack is now complete
*on paper*: layer 1 shipped (digest grammar, verdict glossary, remedy-carrying rows —
`docs/wiki/tui-client.md`), layer 4 shipped (spec 045 `?` overlay with tiers and the
lessons seam), layer 3 is spec 046's exercises plus the D11 exercise panel, layer 2 is
carded (TASK-115 explain tool, TASK-117 lessons, now the decision-5 lesson row). The
remaining biggest gap is lens clause (d): **the design reference is not yet the buildable
authority it claims to be.** `docs/design/tui/INDEX.md` promises "an implementer should
be able to build every screen from these files," but the files froze at the TASK-34
widescreen scope — `panels/dock.md`'s metatron tab predates six specs of shipped content
(provider table, horizon rows, standing orders, stage segment, verdict rows), and the
help overlay exists in the reference only as a section appended to `patterns/keymap.md`.
The de-facto current reference is the monolithic wiki note `docs/wiki/tui-client.md` —
accurate and pinned, but organized by code file, failing the corpus's segment-and-tag
rule ("information is easier to reference when it is neatly tagged, described and
categorized", [[_grounding]] § General manual-writing). Decision 4 fixes this by fiat;
this note's job is to specify *what the fixed thing looks like* (§ Design-doc v2 below)
and how the help layers map onto the Staged Cockpit (§ Help-stack layering).

## Help-stack layering under the Staged Cockpit

The four layers, mapped onto the decided frame:

- **Layer 1 — reduce the need** ([[_grounding]] § Cogmind: "make it so they don't
  actually need help in the first place"): the shipped digest/glossary/remedy surfaces,
  *plus* decision 3's stage-shaped defaults, which are themselves a layer-1 move — a
  stage-1 screen with fewer visible panels needs less explaining. The corpus's "dynamic
  depth" phrasing (playable in 15 minutes, deeper stats on demand) is exactly
  progressive defaults with everything reachable via `?`.
- **Layer 2 — context help at the point of question** ("specific questions are best
  handled at the point where they're asked", [[_grounding]] § Cogmind): the lesson row
  (decision 5) with its UI-pointer field; chronicle `⏎` jump-to-source (D3); the
  always-on detail pane; the `explain` tool as the pull half of TASK-115; and the
  badge deep-link recommendation retained below.
- **Layer 3 — bounded tutorial** (Cogmind's four-room level, "teaching through
  sequenced encounters rather than modal pop-ups"): scenario exercises (spec 046
  first-night/the-law) + the D11 exercise panel and attach-time briefing + the stage-1
  TutorCharter. The scenario IS the tutorial; the panel gives it the on-screen goal
  frame.
- **Layer 4 — tiered reference** (`?` basic/advanced, searchable manual): the spec 045
  overlay — keys (tiered), screen walkthrough, lessons pull-reference, and now the D9
  guardian section — with `docs/player/` as the out-of-game deep reference in the
  DF-wiki role ([[Manual-Structure-Conventions]]), including the four per-stage
  quickstarts ([[Quickstart-Guide-Patterns]] audience splitting, shipped via TASK-68
  AC#9).

**In-game vs `docs/player/` split, restated:** in-game holds everything needed *during*
play at the deterministic no-LLM floor (keys, screen, glyphs, stage verbs, active
lesson, ceremonies' replay); `docs/player/` holds pre-play orientation, per-stage
quickstarts, and depth — the corpus's quickstart "What next?" hand-off runs *toward*
the wiki-role pages, and the pages' own job shrinks to teaching the lookup keys
(`?`, `m`) per the whatis-delegation pattern ([[Explaining-The-Screen]]).

**The prompting-verb gap, closed by decision.** Round 2 found the deterministic floor
taught zero prompting (the overlay's three sections were keys/screen/lessons; TASK-117's
trigger taxonomy was mechanics-only). D9 adds the guardian section (stage identity +
concept, granted verbs, one example ask per verb — renderable from `stagesLadder` and
the stage ceiling with no model), and the prompting-lesson taxonomy is adopted into the
lessons work: first rejected tool call, first `metatron.charter_observed` custom, first
fuzzy order — the first-occurrences of the *verb being taught*, not just of sim
mechanics. Every lesson string carries its pull-path suffix, honoring the corpus caution
that pushed help gets missed ("hot pink and blinking… and still people sometimes miss
them", [[_grounding]] § Cogmind).

## Reconciliation with the sibling branches

- **Staged Cockpit (Game-UI-UX branch) — adopted as the frame.** That branch's verdict
  ("the central verb has the smallest surface in the app") is complementary to this
  branch's ("the floor teaches no prompting"): the guardian console page gives the
  verb its surface; this branch's layers say how the surface teaches. The sibling's
  never-build-an-in-TUI-text-editor instinct survived into decision 2, which this
  branch endorses — the corpus never shows a game teaching authoring inside a widget;
  quickstart-style docs teach *workflows* ([[Quickstart-Guide-Patterns]]), and
  "$EDITOR + the TUI confirms binding" is a specifiable workflow.
- **Lesson row (Learning-Game-Design branch) vs this branch's badge-deep-link — both,
  different layers.** The sibling's dwell-row argument (a feed line scrolls off at
  speed; RimWorld's helper is a persistent region) won decision 5, and rightly: this
  branch's own corpus says RimWorld's helper "sits in the top-right of the screen" as
  a persistent panel, not a log line ([[_grounding]] § RimWorld). The badge deep-link
  is a *separate, retained* recommendation: it is layer-2 context help for header
  badges (`?` opens pre-focused on the active badge's anatomy row), not lesson
  delivery — the two compose (a lesson's pointer field may name the overlay section).
- **Exercise panel, ceremony, postmortem (Gameplay-Patterns + Learning-Game-Design
  branches) — adopted; the design-doc taxonomy accommodates them** (§ below). The
  ceremony/postmortem takeovers are a new *surface class* (body-replacing overlays,
  the help overlay's rendering precedent), so the v2 taxonomy gives overlays/ three
  files, not one.
- **The unlock voice (Learning-Game-Design branch) — adopted as D6**: attribution to
  the player's authorship, skin-resolved. Consistent with this branch's tutor-lane
  doctrine in [[Analysis-In-Game-First-Teaching]] (the graded artifact is the
  player's text).
- **Telemetry split (Gameplay-Patterns branch, D10) — adopted; it also fixes a
  reference-structure problem this branch flagged**: the metatron tab mixed skinnable
  fiction with never-skinned engine telemetry, which would have made the control
  tables' skin-token column ambiguous row by row. Splitting the tab makes the token
  boundary a *file* boundary.

**Decision 6 vs the corpus — risk recorded, not relitigated.** The operator chose
maximum salience: both big moments seize the screen. The corpus evidence this branch
carries points the other way in its one articulated case: Cogmind's tutorial teaches
"through sequenced encounters rather than modal pop-ups" ([[_grounding]] § Cogmind) —
modal interruption was the explicitly rejected alternative in the only corpus system
that states a position. The corpus also supports the operator's side of the trade:
non-interruptive delivery demonstrably gets missed (the same source's blinking-message
caution). The decision's own mitigation — dismissable and replayable from `?`,
`stages`, and the morgue — is precisely this branch's push/pull invariant ("the same
content also exists as searchable reference", [[In-Game-Contextual-Help]]), so the
risk is bounded **iff replayability is enforced as an acceptance criterion** on the
ceremony and postmortem surfaces, not left as intent. That requirement belongs in
`overlays/ceremony.md` and `overlays/postmortem.md` as a stated invariant.

## The design-doc v2 structure (seeds the deliverable)

Decision 4 makes `docs/design/tui/` the living authority. The corpus's reference-layer
rules ([[Manual-Structure-Conventions]]) translate as follows:

**Taxonomy** (pages/ · panels/ · overlays/ · patterns/, per decision 4):

- `INDEX.md` — authority statement, the gate rules, the region-anatomy diagram.
- `anatomy.md` — the screen-region index: every visible region, strip, and badge maps
  to its owning file. This is NetHack's organize-by-screen-region device
  ([[Explaining-The-Screen]]) applied to the doc's own navigation — the
  human-scannable entry point.
- `pages/` — `home.md` (composite; stage-shaped defaults), `guardian-console.md`
  (decision 1), `solo-views.md`.
- `panels/` — `map.md`, `chronicle.md`, `guardian.md` (fiction-layer content only;
  the skin boundary, D10), `systems.md` (telemetry; never skinned), `exercise.md`
  (D11, scenario worlds), `villagers.md`, `villager-strip.md` (D12),
  `lesson-row.md` (decision 5), `guardian-strip.md` (decision 7), `minibuffer.md`.
- `overlays/` — `help.md` (extracted from keymap.md; four sections incl. the D9
  guardian section; the lessons pull registry), `ceremony.md` and `postmortem.md`
  (decision 6; replayability invariant stated).
- `patterns/` — `focus-contract.md`, `chronicle-grammar.md`, `keymap.md` (+ the
  decision-8 input-parity rule), `layout.md` (+ row-budget math absorbing the three
  new permanent rows), `skin-tokens.md` (D2: fiction strings as tokens; token
  naming/lookup contract), `stage-defaults.md` (decision 3's defaults machinery).

**Control tables — the AI-parseable unit.** Per panel/overlay, one table with a fixed
header across every file: `control/region · states · data source (Status field /
event type / replica) · renderer (Go symbol) · keys (+mouse per decision 8) ·
introduced-by (spec) · skin-token`. Uniform columns are what let an AI implementer —
and a sweep script — query the reference like CDDA's generated-from-data guide
([[Manual-Structure-Conventions]] "Keeping reference honest"). Fiction strings appear
*only* as skin tokens, which mechanizes the corpus's never-mix-lore-with-controls rule.

**Anti-drift, mechanized:** (a) every file carries a `verified_against` pin (the wiki
convention); (b) a check script + gate enforces same-PR amendment for any PR touching
`internal/tui` — extending INDEX rule 4 from "deviation" to "any change" (the TASK-82
`check-freshness.mjs` precedent); (c) extend the `help_test.go` sweep-test idea
outward where cheap — a test parsing `keymap.md`'s tables against the key dispatch
makes the keymap un-rottable, the strongest honesty mechanism the corpus records
(generation-from-data beats convention).

**Keep the conventions that already match the corpus:** mockup-first pages (ASCII
mockup + per-callout prose = "screenshots with call-outs *start* the instructions",
[[_grounding]] § General manual-writing); `keymap.md` stays one printable
reference-card page ([[Manual-Structure-Conventions]] reference cards); mnemonic-key
and reserved-seam conventions codified in a short "binding rules" section
([[Explaining-The-Screen]] "Keys that document themselves").

## Recommendations as they now stand (ranked)

1. **Build design-doc v2 per the structure above** (decision 4) — the run's deliverable;
   everything else lands *as pages in it* before code.
2. **Skin-token contract first** (D2): `patterns/skin-tokens.md` + the TASK-121 spec's
   token contract precede any new fiction literals in lessons, overlay content, or
   ceremony copy (concrete sites found in the read-only pass: `internal/tui/help.go`
   carries 5 Metatron literals today; footer hints and `stagesLadder` identities are
   further token consumers).
3. **Lesson row** (decision 5) with the adopted internals — one active, spaced,
   opportunity-decay, UI-pointer field, length budget — and the prompting-verb trigger
   taxonomy folded in; every lesson string names its pull path; the lessons registry
   feeds `helpLessons` (spec 045 seam).
4. **`?` guardian section** (D9): static-per-stage, model-free; the deterministic floor
   now teaches the central verb.
5. **Ceremony + postmortem overlays** (decision 6) with the replayability invariant as
   an explicit AC — the corpus-derived mitigation for the interrupt policy.
6. **Badge deep-link** (retained, cheap): with a header badge active, `?` opens
   pre-focused on that badge's screen-walkthrough row — layer-2 help for dynamic state.
7. **Report card home** per D5 (guardian-console card at stopping points), specified in
   `pages/guardian-console.md` before TASK-115 builds.

## Open questions (genuinely still open)

- **Audience → advanced tier** (carried from the synthesis's Q3): does the overlay's
  advanced tier — and the control tables' data-source column as projected player-side —
  expose raw registry values, or stay plain-language-only? Unresolved by the eight
  decisions.
- **Vertical row budget**: decisions 5 + 7 + D12 add up to three permanent rows in an
  exact-height layout; `patterns/layout.md` v2 must state the small-terminal fallback
  order (which strip folds first) — a design constraint no decision has ruled on.
- **The overlay's no-LLM invariant, restated**: spec 045 pinned content as "never
  derived from live status"; D9 amends that deliberately, and ceremony replay + the
  lessons registry pull further in the dynamic direction. `overlays/help.md` must state
  precisely which sections remain byte-identical with nil status (the floor guarantee)
  and which are status-derived.
- **Lesson row in the narrow fallback**: decision 5 specifies the widescreen composite;
  whether the narrow single-pane layout carries the row, a badge, or overlay-only is
  unspecified.
- **Corpus gaps carried forward** (unchanged from [[Analysis-In-Game-First-Teaching]]):
  observe-mostly games' documentation of the observe/intervene split remains ungrounded
  here (the Learning-Game-Design branch now covers learning-helper anatomy).

## Basis

- [[_grounding]] — Cogmind's four layers and modal-avoidance stance; RimWorld's
  persistent helper panel and push/pull duality; NetHack's region organization and
  whatis delegation; CDDA's generated-from-data reference; manual-writing separation
  and tagging rules; reference cards
- [[In-Game-Contextual-Help]] — the layer ordering and the missed-message caution
- [[Explaining-The-Screen]] — region-index navigation; self-documenting keys
- [[Manual-Structure-Conventions]] — segment-and-tag, show-then-tell, reference honesty
- [[Quickstart-Guide-Patterns]] — audience-split quickstarts, hand-off structure
- [[Analysis-In-Game-First-Teaching]] — the decided in-game-primary frame this note
  builds on
- `docs/design/reorient-2026-07-25-ui.md` — the eight operator decisions and D1–D13
  (fixed constraints)
- Sibling reorientation drafts (Game-UI-UX, Game-Gameplay-Patterns,
  Learning-Game-Design branches), reconciled in prose above per vault isolation —
  the Staged Cockpit frame, the lesson-row dwell argument, the exercise
  panel/ceremony/postmortem surfaces, the unlock-attribution voice, the telemetry
  split
- Codebase/board evidence from the read-only pass: `docs/wiki/tui-client.md`,
  `docs/wiki/curriculum-ladder.md`, `docs/design/tui/` (INDEX, keymap, chronicle),
  `specs/045-tui-help-overlay/contracts/help-content.md`, `internal/tui/help.go`
  (5 Metatron literals), `docs/player/` (13 pages incl. the shipped
  screen-orientation and keys pages), TASK-68/82/115/117/121
