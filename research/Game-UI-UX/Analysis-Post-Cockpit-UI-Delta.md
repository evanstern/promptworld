---
title: Analysis — Post-Cockpit UI Delta
aliases: [UI Delta Analysis 2026-07-26, Post-Staged-Cockpit Evaluation]
tags: [analysis, game-ui, tui, teaching-game, delta]
type: analysis
created: 2026-07-26
updated: 2026-07-26
related: ["[[Game-UI-UX]]", "[[Analysis-Teaching-Game-TUI]]", "[[Analysis-Tile-Vocabulary-Expansion]]", "[[TUI-and-Roguelike-UI-Craft]]", "[[Chat-and-Agent-Console-Rendering]]", "[[RimWorld-Interface]]", "[[Dwarf-Fortress-Interface]]", "[[Terraria-Interface]]", "[[LLM-Agent-Sim-Interfaces]]", "[[Recurring-Interface-Patterns]]", "[[Brief-and-Assumptions]]"]
---

# Analysis — Post-Cockpit UI Delta

_The 2026-07-26 re-run evaluation: with the entire Staged Cockpit course of action shipped
(TASK-123…129, PRs #98–#103; spec 063 grounded feedback; spec 066 stage defaults; spec 068
tile registry), how well does the project serve the teaching-game lens NOW, which of this
branch's findings are implemented, and what does the corpus say the next UI work is?
Written for the 2026-07-26 reorientation run; the operator's ratified decisions (recorded
in `docs/design/reorient-2026-07-26-ui.md`, decisions 1–10) are treated as fixed
constraints. Cross-grounded against the sibling Game-Gameplay-Patterns,
Learning-Game-Design, and Game-Player-Docs delta drafts. [[Analysis-Teaching-Game-TUI]]
and [[Analysis-Tile-Vocabulary-Expansion]] are the immutable prior layers this note
builds on; superseded items from them are declared below, never edited in place._

## Verdict

**The 2026-07-25 course of action shipped essentially in full, and it shipped this
corpus's recommendations, not just its own claims — but the two clauses this branch
certified strongest each carry one newly exposed flaw, and both are now decided fixes.**
All 25 files in `docs/design/tui/` carry `status: shipped` with `verified_against` pins;
the check script (`scripts/check-tui-design.mjs`) exists and gates PRs; every surface the
synthesis specced — guardian console, systems split, lesson row, guardian/villager strips,
ceremony/postmortem, stage defaults, jump-to-source — is verifiable in both the design
corpus and `internal/tui/`. Lens clause (b) has the largest page in the app; clause (a)
gained the villager strip and condition overlays; clause (c) runs from pushed lesson row
to pulled `?` guardian section to graded report card; clause (d) — the one that flatly
failed last run — now has a real page-by-page authority with a freshness gate.

The two flaws, surfaced by the sibling branches and conceded here:

1. **Clause (d) passes structurally, not yet semantically.** The docs branch caught the
   gate lying once: `overlays/postmortem.md` is `status: shipped` with a fresh pin while
   seven renderer cells still read `unbuilt (wave 4)` — the script validates pins and
   table headers, never cell content (INDEX.md rule 4 admits this). My prior "extension,
   not replacement" verdict oversold what the gate mechanizes. Decision 2 (semantic lint)
   is the fix.
2. **Clause (b)'s feedback half is corpus-correct in form and wrong in content.** The
   report-card card seam I credited as the shipped HITL-card pattern
   ([[Chat-and-Agent-Console-Rendering]]) grades by generic event presence, rendering a ✓
   on `agent.died: 2` at a failed first-night — a false checkmark at the game's most
   salient teaching moment. Under Sylvester's noise doctrine ("noise is signal that fails
   to transmit meaning" — [[RimWorld-Interface]]) this is worse than noise: it is
   counter-signal, and RimWorld's itemized numeric thoughts work precisely because every
   number is true. Decision 1 (unify on `sim.EvaluateRubric`) is the fix.

**The single biggest remaining gap owned by this branch is unchanged and now scheduled:
the map's inspection dead-end.** The DF `k` look-cursor pattern — the one row of the
corpus's progressive-disclosure chain still missing — is specced as TASK-142 and, per
decision 4, runs in the next UI lane with the DF content-hierarchy amendment and the tile
registry's `meaning` rows as whatis content. Until it ships, the map is legible at a
glance but not interrogable at a point.

## The operator decisions as given constraints

Ratified 2026-07-26 (`docs/design/reorient-2026-07-26-ui.md` §Decisions); recorded here,
not re-argued. The ones this branch's evaluation feeds directly:

- **Decision 4 — TASK-142 runs in the next UI lane**, amended with (i) DF's fixed
  tile-content hierarchy — agents → piles/chests → structures → terrain — the `k`-cursor
  scan-order fact ([[Dwarf-Fortress-Interface]]), and (ii) the spec-068 tile registry's
  `meaning` rows as the TILE pane's whatis content, plain-language per FR-020. This
  closes my re-run open question 1 and adopts the docs branch's whatis framing (below).
- **Decision 8 — mouse-parity sweep test commissioned** as a small task: parse the
  control tables' `keys+mouse` column, assert every non-`—` cell has a handler, burn
  down `patterns/keymap.md`'s rollout note. Closes my open question 2; converts
  decision-8-of-2025 from doctrine-with-a-worklist into a gate — Cogmind's hard rule
  ([[TUI-and-Roguelike-UI-Craft]]) finally mechanized.
- **Decisions 1–3, 5–7, 9** (report-card truth; semantic lint; TASK-67 → HIGH as the
  iteration rung; exercise catalog + ~3 incident kinds; in-TUI ladder in the `?`
  guardian section; first-prompt quickstart step; duel-before-faith lane order) —
  primarily sibling-branch findings; this branch's UI contributions to each are folded
  into the reconciliation below.
- **Decision 10 — parked watch items**, genuinely open, including three that were this
  branch's re-run questions: chronicle/world search `/` (revisit past ~20 villagers),
  colorblind/contrast palette research (with the web/tileset projection behind it), and
  the villager-strip ambient-action badge. Kept open below, not smuggled into
  recommendations.

## What shipped against this branch's findings — verified

Each row verified against the design corpus, `internal/tui/`, and git history — not
against claims.

- **Chat-console conventions → guardian console** (`pages/guardian-console.md`, spec
  053/TASK-125). Document-style turns (`consoleTurnLines`) per ChatGPT's non-bubble
  fixed-measure convention; `$EDITOR` charter handoff with the honest "charter changed —
  next turn binds it" line; report-card cards at stopping points via
  `rebuildConsoleCards` (spec 063). Streaming-as-trust is served minimally but
  defensibly: a `⋮ thinking…` busy row plus incremental `»` verdict rows
  (`panels/guardian.md`) — the NN/g overload evidence argues *against* full token
  autoscroll ([[Chat-and-Agent-Console-Rendering]]), so this depth is right. Prior open
  question 1 (turn rendering depth) resolved: it was a layout project.
- **RimWorld's fixed-region grammar → village lens** (spec 060/TASK-129). The villager
  strip (`panels/villager-strip.md`) is the colonist bar transplanted with the
  corpus-correct twist of reusing the map's exact glyph/style rules — Sylvester's
  metaphor-vocabulary discipline. Map condition overlays are RimWorld 1.5's mood-glow
  move: recolor, never re-glyph, exactly [[Analysis-Tile-Vocabulary-Expansion]]'s
  state-variant rule.
- **Click-to-jump → chronicle jump-to-source** (spec 049; `panels/chronicle.md`:
  `⏎ · click line`, `contracts/jump-to-source.md`; unlocatable events get an honest
  actions-bar note). My named "no jump-to-source anywhere" gap is closed, and it landed
  keyboard+mouse together as the parity doctrine demanded — the corpus's first real
  mouse target.
- **Learning helper → lesson row + stage defaults** (spec 055; spec 066/TASK-128).
  Progressive disclosure as layout defaults — the Terraria staged-codex finding served
  without capability-lock doctrine creep ([[Terraria-Interface]]).
- **DF pause-and-center → takeover pair** (spec 056). Decision 6-of-2025 shipped with
  the stipulated mitigations; the D1 linear-stream floor held (ceremony replayable from
  `stages`, postmortem from the morgue).
- **The deterministic teaching floor → spec 063** (explain tool, tutor guide, help
  overlay's D9 guardian section). The player-docs finding I adopted last run — the `?`
  floor "taught zero prompting" — is closed.
- **[[Analysis-Tile-Vocabulary-Expansion]], executed with unusual fidelity** (spec
  068/TASK-143, `internal/tui/tiles.go`, PR #105): one registry feeding map, legend, and
  `?` overlay ("one grid model, swappable skins" — DF's glyph-swap architecture,
  [[Dwarf-Fortress-Interface]]); named tokens classed `semantic16` vs `material256` (the
  analysis's palette rule, verbatim in code comments); variants as style transforms that
  *cannot* change the glyph (FR-003); marsh `░` / sand `▒` from the CP437 shading tier —
  "texture, not object." Byte-identity pins guarded the migration. Decision 4 now makes
  the registry a fourth surface's content source (the look-cursor whatis pane).
- **The vertical-budget tension, ruled.** `patterns/layout.md` specifies `bodyMin = 10`
  and the fold order (villager strip → lesson row → guardian strip last, because the
  budget stays visible). [[Analysis-Teaching-Game-TUI]]'s top unresolved tension is
  resolved in the direction the corpus supported. Prior open question 2 closed.
- **FR-020 audience ruling** (INDEX.md): raw registry values are engineer-facing table
  content; player projections stay plain-language, debug values behind a *mode*. Prior
  open question 3 closed.

## Reconciliation with the sibling reports

**Game-Player-Docs draft.** Two adoptions and one concession. *Adopted:* its TASK-142
whatis framing — the tile registry's glyph/name/`meaning` rows are exactly the
NetHack-`/`-whatis content the look-cursor pane should render (now decision 4) — this
*strengthens* my registry finding: one table now feeds four surfaces, the
one-legend-source discipline compounding. *Adopted:* the semantic-drift lint (decision 2)
— its `postmortem.md` seven-stale-cells discovery, with `exercise.md:110` in miniature.
*Conceded:* my draft verdict "lens (d) now passes" was too strong; the corpus's own
lesson (convention-dependent reference rots; only generation-from-data or mechanized
lint survives) applies to the gate I was crediting. One residual divergence, named not
papered over: the docs branch defers help-overlay *search* while my corpus tracks
chronicle/world search as the genre's QoL convergence endpoint (Terraria 1.4.5, RimWorld
1.5 `Z`, Dubs Mint as mod-layer gap detector — [[Recurring-Interface-Patterns]]). Both
are parked under decision 10; my position for the eventual reopening: build ONE search
grammar (`/`) spanning chronicle, help, and world-find rather than per-surface searches
— DF classic's "three different ways of scrolling" ([[Dwarf-Fortress-Interface]]) is the
documented cost of letting input grammars accrete per surface.

**Learning-Game-Design draft.** *Adopted:* the report-card truth finding (decision 1) —
it intersects and upgrades my TASK-144 note: I flagged the flaky test as guarding the
lens-(b) feedback surface; the pedagogy branch showed the surface itself mis-teaches.
From this corpus's side the diagnosis is the same fact in interface terms: the card
*form* (stopping-point cards, evidence-cited, no-blame register) is corpus-correct; the
*derivation* violates the one property that makes RimWorld's numeric-thought grammar and
Cogmind's log discipline work — every displayed signal is true. Sequencing after
TASK-144 (same code) is right. *Adopted:* the in-TUI ladder view (decision 6) — the
forward ladder (identity · concept · earned/next · unlock evidence) in the `?` guardian
section is Terraria's staged-Bestiary visibility fact applied at the deterministic floor;
my addition: it spends `?` overlay budget, see Tensions. *Aligned, no collision:* its
teaching-signal instrumentation and fire-once/retry questions are parked (decision 10);
nothing in this corpus decides them.

**Game-Gameplay-Patterns draft.** *Adopted:* fork-duel elevation (decision 3) and
duel-before-faith (decision 9). My prior layer only reaffirmed D7; the gameplay branch
showed all D7 prerequisites shipped and the loop's compare rung missing. This branch's
UI contribution: the duel's v1 scoreboard reuses `reportCardView` + the postmortem
register (one renderer family, glossary discipline held), and dual side-by-side TUI
stays deferred — consistent with the five-region anatomy's guardrail and the
replay-vs-live cost finding ([[LLM-Agent-Sim-Interfaces]]: comparison over recorded runs
is the cheap, proven shape; a live dual client is not). *Adopted with a UI rider:* the
exercise catalog + incident vocabulary wave (decision 5). Rider from this corpus: every
new incident kind spends alert-vocabulary budget — the chronicle grammar's whole-line
alert class is deliberately small (Cogmind's importance-filtered log discipline,
[[TUI-and-Roguelike-UI-Craft]]), and new incidents should enter through the *shipped*
severity channels (chronicle severity line + map condition overlay + tile-registry
variant where terrain-visible), never a new channel per incident. The
small-stable-vocabulary rule from [[Analysis-Tile-Vocabulary-Expansion]] applies to
event vocabulary exactly as to glyphs. *No collision:* its metric-gaming watch item
(all-terms-live gauges) is parked under decision 10; from this side, live gauges are
RimWorld's honest-meter grammar and fine at one-exercise scale.

## What remains unserved, ranked under the decisions

With decisions 1–9 scheduled, this branch's residue ranks as:

1. **TASK-142 execution quality** (decision 4; next UI lane). The corpus-critical
   details to hold: fixed panel geometry (Cogmind's fixed-grid rule — mode changes swap
   content, never panel size, already AC#8), the DF content hierarchy, one-esc-one-layer,
   parity shipping keyboard+mouse together, and the registry-`meaning` whatis rows. The
   warmth/light header (AC#7) is RimWorld's cell-info readout — good.
2. **Reverse jump: strip/roster → camera.** RimWorld's colonist bar is click-to-jump;
   the shipped strip is display-only. `⏎`/click on a strip glyph or roster row centering
   the camera (via the existing `centerCameraOn` + jump-to-source contract) completes
   the bidirectional list↔map loop the corpus documents ("zoom to creature",
   [[Dwarf-Fortress-Interface]]; colonist-bar click, [[RimWorld-Interface]]). Small;
   natural rider on the TASK-142 lane or the parity-sweep task. Not yet on any card —
   this is the one net-new recommendation this delta leaves unscheduled.
3. **Parked, kept open (decision 10):** chronicle/world search (reopen past ~20
   villagers — the [[Recurring-Interface-Patterns]] scale-threshold row); colorblind/
   contrast palette research (the token table `styleTokens` is enumerable and is the
   mechanism — Cogmind's externalized-colors precedent — but the palette itself remains
   ungrounded, [[Analysis-Tile-Vocabulary-Expansion]] open question 2); the web/tileset
   projection (architecturally cheap now the registry exists, but gated on the font and
   palette gaps — analysis open question 1); the strip ambient-action badge (the
   Smallville chain's layer-1.5, [[LLM-Agent-Sim-Interfaces]] — terminal-safe only as a
   strip/roster badge, never an on-grid emoji).

## Superseded and resolved prior-layer items

Declared here; the prior notes stay untouched.

- [[Brief-and-Assumptions]] open questions (observation-vs-command; renderer target):
  superseded by the ratified lens and D1 — already declared in
  [[Analysis-Teaching-Game-TUI]]; restated for this layer.
- [[Analysis-Teaching-Game-TUI]] open questions: **Q1** (turn rendering depth) resolved
  — shipped as layout-project depth with a liveness row; **Q2** (minimum-height fold
  order) resolved by `patterns/layout.md` rulings a/b; **Q3** (audience) resolved by the
  FR-020 corpus-wide ruling; **Q4** (fork-duel revisit criterion) resolved by decision 3
  — the criterion turned out to be "its prerequisites all shipped"; **Q5**
  (seen-lessons reset) partially open — shipped v1 improves on fire-once
  (queued-then-decayed lessons re-fire, per the pedagogy branch's read), full decay
  parked under decision 10 (revisit >15 lessons).
- [[Analysis-Tile-Vocabulary-Expansion]]: the verdict's four rules are all implemented
  (spec 068); open questions 1 (fonts) and 2 (colorblind palette) remain open, now
  parked as decision-10 watch items; open question 3 (natural-language state at density)
  remains the MOC's live research gap — and is *more* pressing now that the guardian
  console renders exactly that content class at full-page height.
- The path-vs-grass color-only distinction (the vocabulary's one never-color-alone
  violation) persists, acknowledged in `tiles.go` comments; fine at its stakes, still
  not a pattern to extend.

## Tensions & tradeoffs

- **Pull-surface budget.** Decisions 4 (whatis pane), 6 (ladder block), and the shipped
  D9 section all grow the pull-teaching surfaces, whose only navigation is tab-cycling;
  the docs branch deferred help search. The legend-budget caution from
  [[Analysis-Tile-Vocabulary-Expansion]] generalizes: the teaching surface degrades
  before the map does. No decision covers overlay-internal navigation; it is the likely
  next instance of the search convergence pattern and should be folded into the parked
  search item when it reopens, not solved separately.
- **Incident vocabulary vs alert discipline.** Decision 5's ~3 incident kinds are the
  first real test of whether the chronicle's deliberately small whole-line-alert class
  survives content growth. The rider above (enter through shipped channels) is a
  discipline, not a mechanism; if incident kinds keep growing, the alert class needs the
  same registry treatment tiles got.
- **Gates guard what they can parse.** Decision 2's lint checks for `unbuilt` markers
  and (optionally) symbol existence; decision 8's sweep asserts handler existence, not
  target-geometry correctness. Both are the right mechanizations, and both leave a
  semantic remainder that stays a same-PR review duty (INDEX.md rule 4). The corpus's
  honesty ceiling — only generation-from-data fully survives — is not reachable for
  mockup-first design pages; the residue should be named in the gate docs, not implied
  away.
- **The duel raises the compare stakes on report-card truth.** Decision 3 makes
  `reportCardView` the scoreboard renderer; decision 1 must therefore land first or the
  duel ships comparisons of untrue checkmarks — the sequencing (decision 1 after
  TASK-144, decision 3 in the duel lane) must hold that order. Named because nothing
  mechanically enforces it yet.

## Confidence & open questions

High confidence on the shipped-ness verification (every fit row checked against code
and pinned pages, not claims) and on the reconciliation positions (each adoption traces
to a corpus fact). Moderate on the pull-surface-budget tension — it extrapolates the
legend-budget evidence to overlays, which the corpus supports by analogy, not by a
documented failure.

Genuinely open (all with recorded homes):

1. **Decision-10 watch items** owned or co-owned by this branch: search grammar scope
   when it reopens (one `/` across surfaces — this branch's position — or per-surface);
   colorblind palette research (prerequisite for any skin beyond the default and for the
   web/tileset projection); strip ambient-action badge (strip-width cost at 12+
   villagers).
2. **Reverse jump (ranked item 2)** — the one unscheduled recommendation; needs a home
   (TASK-142 rider vs parity-sweep rider vs its own small card).
3. **Natural-language state at density** — the MOC's standing research gap, unresolved
   by anything shipped; the guardian console is now the surface that would consume the
   answer.

## Basis

- [[_grounding]] — all six sections; every corpus claim above cites through its
  per-source note.
- [[Analysis-Teaching-Game-TUI]], [[Analysis-Tile-Vocabulary-Expansion]] — the immutable
  prior layers; this note is their delta.
- [[TUI-and-Roguelike-UI-Craft]] — Cogmind parity/fixed-grid/log discipline,
  externalized colors, accessibility floor.
- [[Chat-and-Agent-Console-Rendering]] — document-style turns, streaming-as-trust and
  its overload failure modes, HITL cards.
- [[RimWorld-Interface]] — colonist bar, click-to-jump, learning helper, honest-meter
  grammar, Sylvester's noise doctrine.
- [[Dwarf-Fortress-Interface]] — `k`-cursor hierarchy, zoom-to-creature, glyph-swap
  architecture, input-grammar accretion cautionary tale.
- [[Terraria-Interface]] — staged disclosure, QoL/search convergence waves.
- [[LLM-Agent-Sim-Interfaces]] — inspector chain, ambient-action layer, replay-vs-live
  cost, scale thresholds.
- [[Recurring-Interface-Patterns]] — the cross-source rows (search convergence, scale
  thresholds, mods-as-gap-detectors) this delta reasons across.
- [[Brief-and-Assumptions]] — superseded framing, declared above.
- Operator constraints: `docs/design/reorient-2026-07-26-ui.md` (decisions 1–10);
  prior-run constraints `docs/design/reorient-2026-07-25-ui.md` (decisions 1–8,
  D1–D13). Sibling drafts (Game-Gameplay-Patterns, Learning-Game-Design,
  Game-Player-Docs delta evaluations, 2026-07-26) read in cross-grounding; referenced
  in prose by branch name per vault isolation rules.
- Project surfaces verified (cited by path, outside this vault by design):
  `docs/design/tui/` (INDEX.md, anatomy.md, pages/, panels/, overlays/, patterns/ — all
  25 files `status: shipped` with pins), `scripts/check-tui-design.mjs`,
  `internal/tui/tiles.go`, `internal/tui/tui.go` (mouse handling),
  `panels/chronicle.md` control table (jump-to-source row), `patterns/keymap.md`
  (parity rule + rollout note), `patterns/layout.md` (fold-order rulings),
  `pages/guardian-console.md`, `panels/guardian.md` (`⋮ thinking…` row),
  `panels/villager-strip.md`, `panels/map.md:168` (look-cursor deferral), board views
  (TASK-142, list), `git log --oneline --since=2026-07-25`.
