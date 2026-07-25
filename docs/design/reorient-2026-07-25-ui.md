# promptworld — reorientation synthesis: UI/UX for the teaching-game pivot (2026-07-25)

**Status:** synthesis complete; board moves executed with operator sign-off 2026-07-25
(commit b6ac1b3 — TASK-123…129 created, TASK-115/117/119/121/67 rescoped).
**Briefing page:** `reorient-2026-07-25-ui-briefing.html` (this directory) — published at
<https://claude.ai/code/artifact/4ae421da-7494-49e9-aafc-3f838f09c15a>.
**Run:** `promptworld-2026-07-25-13-58-18` (reorient).
**Lens:** promptworld is pivoting into a staged prompting-skills teaching game (curriculum
ladder TASK-68, skinnable guardian TASK-121, the eight ratified learning-game decisions of
2026-07-25) while remaining an ambient, terminal-first LLM-agent world sim. The UI/UX must:
(a) make a living LLM-agent village legible at a glance; (b) make prompting the guardian the
central, rewarding player verb; (c) teach itself in play; and (d) be fully specifiable — page
by page, feature by feature, control by control — so both humans and AI implementers can
build new features and update existing UI directly from the design doc.
**Corpus:** `research/Game-UI-UX` (Analysis-Teaching-Game-TUI),
`research/Game-Gameplay-Patterns` (Analysis-Play-Loop-Surfaces),
`research/Learning-Game-Design` (Analysis-Pedagogy-To-UI),
`research/Game-Player-Docs` (Analysis-UI-Reference-And-Help-Stack).

## TL;DR

All four evaluations converged on one diagnosis: the TUI is a world-class **watching**
instrument — its digest grammar, verdict glossary, remedy-carrying suppression rows, and `?`
overlay match or beat the documented best practice of every surveyed game — but the pivot's
central verb (prompting the guardian) has the smallest surface in the app, the curriculum has
almost no living screen, and the page-by-page design reference has rotted (~6 specs stale),
so lens clause (d) fails today. The operator chose the fullest remedy: the **Staged
Cockpit** — a guardian console page with telemetry split out, map-legibility wins, a
teaching chrome (lesson row, guardian strip, takeover ceremonies), scenario surfaces, and
stage-shaped layout *defaults* — all additive on the existing five-region anatomy. **No big
refactor is necessary or advisable**; the replica/reducer machinery and the five-region
layout are the asset. The first deliverable is `docs/design/tui/` v2: the control-by-control
reference, rebuilt as the living authority with a freshness gate, authored spec-before-build
for every new surface.

## Decisions

Operator decisions, steering round of 2026-07-25 (verbatim intent; FIXED constraints):

1. **Headline direction: Staged Cockpit.** Guardian Console (the guardian conversation
   becomes a first-class full-height page; telemetry splits out of the guardian tab into a
   systems tab) + Village Lens (map/legibility wins: jump-to-source, villager strip,
   condition overlays) + stage-shaped progressive **layout defaults**.
2. **Charter/skills authoring home: in-TUI read + `$EDITOR` write.** The console renders
   charter/skills with binding status and honest lock notices; authoring hands off to
   `$EDITOR`; the TUI confirms "charter changed — next turn binds it". No in-TUI text editor.
3. **Stage may shape TUI layout defaults — defaults only.** Which panels/tabs are visible by
   default is stage-resolved; everything remains reachable at every stage and via `?`;
   pre-ladder worlds get everything. Capability locks remain angel-only (spec 046 doctrine
   untouched).
4. **UI authority: `docs/design/tui/` v2 with a freshness gate.** The design reference
   becomes the living page-by-page authority (pages/ · panels/ · overlays/ · patterns/ +
   uniform control tables with skin-token columns); files carry `verified_against` pins; a
   check script + gate enforces same-PR amendments (TASK-82 precedent).
5. **First-occurrence lessons render in a dedicated lesson row.** One always-visible row
   above the minibuffer: one active lesson, ≤2 lines, dwells until done/dismissed, points at
   a key/tab, anti-spam (one active, spaced, opportunity-decay). Default-on at stages 1–2;
   badge+overlay-only default at stage 3+ and pre-ladder (per decision 3's defaults).
6. **Interrupt policy: BOTH big moments seize the screen.** The stage-unlock ceremony takes
   over immediately when `curriculum.stage_unlocked` fires on an attached client, and the
   postmortem takes over on `run.ended`. (Operator chose maximum salience over the
   never-interrupt-play rule.) Both surfaces are dismissable and replayable (`?`, `stages`,
   morgue) — replayability-from-pull is an explicit AC on both.
7. **Action budget always visible: guardian strip above the minibuffer.** One line pairing
   the budget (charge bank, regen, standing-order count, faith once TASK-118 lands) with the
   input line, making the minibuffer read as THE verb.
8. **Input parity ratified as doctrine.** Every action reachable by both keyboard and mouse;
   keyboard remains primary and complete; rollout incremental. keymap.md gains the rule, and
   the v2 control tables gain a mouse column so a parity sweep test can exist.

### Lead-adopted defaults (presented to the operator for veto; unvetoed as of this writing)

D1. Terminal-first renderer; `attach`/`tail` linear streams stay first-class specified
    surfaces (screen-reader floor); any web surface is a separate later projection.
D2. Skin tokens first: TASK-121's spec ships the token contract BEFORE TASK-115/117 write
    new fiction literals; the design doc renders all fiction strings as skin tokens.
D3. Chronicle `⏎` jump-to-source gets built (fills the reserved seam); click-a-line too.
D4. Incident-schedule visibility is a per-exercise field: forecast at stages 1–2, fog from
    stage 3.
D5. Report card (TASK-115) renders as a guardian-console card at natural stopping points
    (run end / pause / exercise resolution; badges between) and inside the postmortem.
D6. Unlock attribution voice: the player's authorship ("your charter proved The Written
    Word"), skin-resolved.
D7. Fork-duel (TASK-67) v1: rubric-first scoreboard with drill-down, then the shareable
    HTML retelling; dual side-by-side TUI deferred.
D8. Seen-lessons state lives per-user (the `unlocks.json` precedent).
D9. The `?` overlay gains a "the guardian" section — static-per-stage, model-free (stage
    identity/concept, granted verbs, one example ask per verb); spec 045's content contract
    is amended deliberately for it.
D10. Telemetry split: provider table, horizon rows, spend move to a "systems" dock tab; the
    guardian tab carries fiction-layer content only — the skin boundary becomes a file
    boundary.
D11. Scenario worlds get an exercise panel (fourth dock tab): framing, event-derived rubric
    gauges, pass/fail state, incident forecast per D4, attach-time briefing, and a
    scenario-cadence narration trigger so the chronicle score-narrative renders during
    short runs.
D12. Villager strip: a one-row colonist-bar-style strip under the header, stage-defaulted.
D13. `q`-detach is the blessed stopping point; the detach affordance says "the world keeps
    running".

## Merged positions

**1. The diagnosis (all four branches).** Watching craft is shipped and corpus-validated:
the digest grammar is Cogmind's log discipline fused with RimWorld's severity language
(Game-UI-UX); the villager drill-down reproduces the Smallville inspector chain
(Game-UI-UX); the `?` overlay is Cogmind's tiered context help with a mechanically-enforced
keymap (Game-Player-Docs). The gaps were all *game* surfaces: no designed scenario screen
(Game-Gameplay-Patterns), no living curriculum presence — one header segment, feed lines
that scroll away at 16x, a CLI-only ladder, the WorldBox discoverability failure
(Learning-Game-Design) — and a guardian conversation violating chat-console conventions in
a telemetry-crowded tab (Game-UI-UX).

**2. The Staged Cockpit (decision 1) as one coherent shell.** Game-UI-UX's three directions
compose rather than compete: the guardian console page gives the verb a first-class surface
(document-style turns, composer, charter read + `$EDITOR` write per decision 2, report cards
per D5); the systems-tab split (D10) simultaneously decongests the console and draws
TASK-121's skin boundary as a file boundary; the village-lens wins (D3 jump-to-source, D12
villager strip, condition overlays) serve at-a-glance legibility; and stage-shaped *defaults*
(decision 3) make the layout itself the progressive-disclosure instrument while touching no
capability doctrine. Everything is additive on the five-region anatomy — the DF-Premium
guardrail (redesigns that orphan the old layout's logic draw their own criticism) holds.

**3. The teaching chrome (Learning-Game-Design + Game-Player-Docs merged).** The lesson row
(decision 5) is RimWorld's ConceptDef transplanted: one active lesson, dwell-until-done,
UI-pointer, spacing, opportunity decay, per-user seen state (D8) — with Game-Player-Docs'
taxonomy split (mechanics lessons AND prompting-verb lessons: first rejected tool call,
first custom charter observed, first fuzzy order), every string skin-token-resolved (D2) and
carrying its pull path. The guardian strip (decision 7) pairs budget with input so the
minibuffer reads as THE verb. The `?` overlay grows the guardian section (D9) so prompting
is taught at the deterministic no-LLM floor. Badge deep-link (a `?` opening focused on the
active badge's row) is retained as a cheap layer-2 addition.

**4. Scenario and takeover surfaces (Game-Gameplay-Patterns + Learning-Game-Design merged).**
Scenario worlds get the exercise panel (D11): framing line, live event-derived rubric gauges
(the decision-trace projection pattern), per-exercise forecast/fog (D4 — the Cassandra-vs-DF
dial as data), attach-time briefing (the Portal safe-room), and scenario-cadence narration so
the ratified "chronicle as score narrative" actually renders inside a short run (the narrator's
~2 chapters/game-day cadence would otherwise produce zero entries). The unlock ceremony and
the postmortem are ONE takeover-surface family (decision 6) with a deliberate voice
asymmetry: success speaks in the player's authorship ("your charter proved The Written
Word", D6); failure speaks in the morgue's no-blame evidence register. Both replayable from
pull surfaces; the ceremony doubles as a session stopping point (D13's shape).

**5. The design doc v2 (Game-Player-Docs, adopted by all).** `docs/design/tui/` becomes the
living authority (decision 4): taxonomy `pages/ · panels/ · overlays/ · patterns/` plus a
top-level `anatomy.md` mapping every visible region/badge to its owning file; `panels/dock.md`
splits into per-tab panels; overlays get real pages (help, ceremony, postmortem); uniform
control tables (control/region · states · data source · renderer · keys+mouse · introduced-by
· skin-token) are the AI-parseable unit; mockup-first pages and the printable keymap card
stay; `verified_against` pins + a check script + a same-PR gate mechanize freshness (the
convention alone already failed for specs 044/046). New surfaces are authored here
spec-before-build.

**6. Doctrine adopted.** Input parity (decision 8) with a mouse column so a parity sweep
test can exist; terminal-first with the linear-stream accessibility floor (D1 — every new
surface needs a specified `attach`/CLI projection or the "text mode lie" hazard grows);
skin-tokens-first sequencing (D2) so TASK-115/117 never write literals TASK-121 must sweep.

## Course of action

No big refactor: every wave is additive on the five-region anatomy and the replica machinery;
the one genuinely structural item is the bottom-chrome row budget, which Wave 0 must rule on.

- **Wave 0 — the deliverable: `docs/design/tui/` v2.** Reconcile with shipped reality
  (specs 013–046), adopt the v2 taxonomy + control tables + pins + check script + gate, and
  author the new-surface pages spec-before-build: guardian console, systems tab, exercise
  panel, ceremony overlay, postmortem overlay, lesson row, guardian strip, villager strip,
  stage-defaults pattern, `?` guardian section. Wave 0 also RULES on: bottom-chrome row
  budget and fold order (lesson row + guardian strip + minibuffer + footer stack ~4 rows at
  stages 1–2), narrow-fallback behavior for the new chrome, and the `?` overlay's
  no-LLM-byte-identity invariant restated for status-derived sections.
- **Wave 1 — skin-token contract** (TASK-121 spec): the token lookup contract ships before
  any new fiction literal (D2); the sweep implementation may lag.
- **Wave 2 — quick wins:** chronicle `⏎`/click jump-to-source (D3, the parity retrofit's
  first act), guardian strip (decision 7), telemetry systems-tab split (D10).
- **Wave 3 — guardian console page** (decision 1/2, D5): document-style turns, composer,
  charter/skills read surface with binding status + `$EDITOR` handoff, report-card cards.
- **Wave 4 — teaching surfaces:** lesson row (TASK-117 per decision 5), exercise panel +
  scenario-cadence narration (TASK-119 per D11/D4), ceremony + postmortem takeovers
  (decision 6), `?` guardian section (D9), stage-default machinery (decision 3).
- **Wave 5 — village lens completion:** villager strip (D12), map condition overlays,
  look-cursor evaluation.

## Board moves

| # | Task | Move |
|---|------|------|
| 1 | NEW | **TUI design reference v2** (Wave 0) — reconcile docs/design/tui with specs 013–046; taxonomy (overlays/, per-tab panels, anatomy.md); uniform control tables incl. mouse + skin-token columns; verified_against pins + check script + same-PR gate; author all new-surface pages; rule on row budget / narrow fallback / overlay invariant. HIGH — this run's deliverable. |
| 2 | NEW | **Chronicle jump-to-source + parity retrofit start** — fill the reserved `⏎` seam; click-a-line; keymap.md parity rule (decision 8, D3). |
| 3 | NEW | **Guardian console page + systems-tab telemetry split** (Waves 2–3; decisions 1/2, D5/D10). |
| 4 | NEW | **Guardian strip** — always-visible budget line above the minibuffer (decision 7). |
| 5 | NEW | **Takeover surfaces: unlock ceremony + postmortem** — one surface family, voice asymmetry (decision 6, D6); replayable-from-pull ACs. |
| 6 | NEW | **Stage-shaped layout defaults machinery** (decision 3). |
| 7 | NEW | **Villager strip + map condition overlays** (D12, Wave 5). |
| 8 | EDIT TASK-115 | Add ACs: render surface = guardian-console card at stopping points + postmortem (D5); guardian verbs reachable from the deterministic `?` floor (D9); skin tokens (D2). |
| 9 | EDIT TASK-117 | Add: dedicated lesson row w/ dwell + pointer + one-active/spacing/decay (decision 5); prompting-lesson tier (first rejected tool call, first custom charter, first fuzzy order); strings skin-tokened + pull-path suffix; per-user seen state (D8). |
| 10 | EDIT TASK-119 | Add UI ACs: exercise panel page (D11) w/ live rubric gauges; visibility VOCABULARY per exercise, not a boolean (D4); attach briefing; scenario-cadence narration; ceremony trigger linkage. |
| 11 | EDIT TASK-121 | Enumerate sweep sites (help.go's 5 literals, footer hints, stagesLadder, design-doc mockups, player-docs page names); token contract ships first (D2); skin boundary = guardian/systems tab split (D10); design doc renders fiction as tokens. |
| 12 | EDIT TASK-67 | Record D7: v1 = rubric-first scoreboard sharing the postmortem's rubric renderer + glossary; then HTML retelling; dual-TUI deferred. |

## Open questions

Genuinely still the operator's; parked with their resurfacing moment:

1. **Ambient postmortem contents** — does the unscored ambient world's `run.ended` takeover
   include the report card, or morgue evidence only (the hybrid-scoring boundary on
   screen)? → decide when Wave 0 authors `overlays/postmortem.md`.
2. **Score voice inside the ceremony** — instrument (rubric checklist), fiction (narrated
   chapter), or both with one authoritative — → decide in `overlays/ceremony.md`.
3. **Live rubric gauges vs gaming-the-metric** — headline-live vs full-breakdown-at-end →
   decide in TASK-119's spec.
4. **Audience for advanced tiers** (raw registry values player-visible?) — carried from the
   learning-game synthesis Q3 → decide when Wave 0 specs the control tables' data-source
   column exposure.
5. **Interrupt-policy watch item** — decision 6 stands; named reopening signal: playtest
   evidence of ceremony fatigue or mid-crisis seizure complaints.

## Unresolved tensions (named by the analyses, carried consciously)

- **Vertical budget contention** (all four): ~4 fixed bottom-chrome rows at stages 1–2 in an
  exact-height layout — Wave 0's first ruling.
- **Pre-ladder worlds get the least pushed teaching** while plausibly hosting new players
  (lesson row defaults off there) — revisit when stage adoption data exists.
- **Forecast fog at stage 3+ removes the legibility that taught stages 1–2** — TASK-119
  specs a visibility vocabulary, not a boolean (D4 as amended by move 10).
- **Accessibility drift** — each new surface needs its linear-stream projection specified
  (D1) or the TUI-only chrome grows the screen-reader gap.
