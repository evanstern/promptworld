---
title: Analysis — Teaching-Game TUI
aliases: [TUI Direction Analysis, Staged Cockpit Analysis, Game-UI-UX Evaluation]
tags: [analysis, game-ui, tui, teaching-game]
type: analysis
created: 2026-07-25
updated: 2026-07-25
related: ["[[Game-UI-UX]]", "[[TUI-and-Roguelike-UI-Craft]]", "[[Chat-and-Agent-Console-Rendering]]", "[[RimWorld-Interface]]", "[[Dwarf-Fortress-Interface]]", "[[Terraria-Interface]]", "[[LLM-Agent-Sim-Interfaces]]", "[[Recurring-Interface-Patterns]]", "[[Brief-and-Assumptions]]"]
---

# Analysis — Teaching-Game TUI

_How well does promptworld's current TUI serve the teaching-game pivot lens — (a) village
legible at a glance, (b) prompting the guardian as the central rewarding verb, (c) the UI
teaches itself in play, (d) fully specifiable page-by-page — and what direction should the
UI take? Written for the 2026-07-25 reorientation run; the operator's steering decisions
(recorded in `docs/design/reorient-2026-07-25-ui.md`) are treated as fixed constraints._

## Verdict

**The current TUI is a world-class watching instrument whose acting surfaces are underbuilt
relative to the pivot, and the ratified direction — the Staged Cockpit — is the right fix:
it composes a guardian-first console, map-first legibility wins, and stage-shaped layout
defaults onto the existing five-region anatomy rather than replacing it.**

The watching craft already implements documented best practice: the chronicle digest
grammar is Cogmind's message-log discipline (terse importance-first lines, routine never
logged, whole-line alerts for exactly four high-salience types — [[TUI-and-Roguelike-UI-Craft]])
fused with RimWorld's severity color language ([[RimWorld-Interface]]); the verdict
glossary and remedy-carrying suppression rows operationalize Sylvester's noise doctrine
("noise is signal that fails to transmit meaning"); the roster → detail → decisions
drill-down reproduces the dominant LLM-sim inspector chain (ambient glyph → click → full
action → memory stream — [[LLM-Agent-Sim-Interfaces]]); and the spec-045 `?` overlay with
tiering and a sweep-tested keymap matches Cogmind's context-help program. At 8–20
villagers the design sits safely inside the documented 25-agent per-agent-inspection
threshold, far from Project Sid's aggregate-dashboard cliff.

The single biggest gap, pre-decisions, was that **the game's central verb had the smallest
surface in the app**: a 3-row minibuffer that truncates input to its visible tail, feeding
a dock tab where guardian replies share space with provider tables, horizon rows, and
order telemetry — while chat-console corpus evidence (document-style turns, fixed-measure
columns, streaming as trust, prompt controls, HITL cards — [[Chat-and-Agent-Console-Rendering]])
describes a far richer conversation surface, and the stage-2 curriculum verb (charter
authoring) had no in-game surface at all. Secondary gaps: no jump-to-source anywhere
(RimWorld's click-to-jump letters, DF's zoom-to-event — [[RimWorld-Interface]],
[[Dwarf-Fortress-Interface]]); the map is a camera, not an information surface (Ge's
map-dynamics doctrine — [[TUI-and-Roguelike-UI-Craft]]); curriculum presence in the client
is one header segment plus two feed lines against Terraria's staged-progress evidence
([[Terraria-Interface]]); and input parity — Cogmind's hard rule — was absent.

Lens clause (d) was the quiet superpower and the quiet liability at once: `docs/design/tui/`
is exactly the page-by-page reference shape the lens demands, but it froze at its original
widescreen scope and no longer describes roughly six specs of shipped dock content. The
verdict stands as **extension, not replacement** — the region anatomy, focus contract, and
mockup-first page conventions are corpus-correct — but extension now explicitly means a
reconciliation-and-taxonomy v2 with a freshness gate, per operator decision 4.

## The operator decisions as given constraints

The steering round (`docs/design/reorient-2026-07-25-ui.md`, decisions 1–8 and lead
defaults D1–D13) resolved this branch's ranked options and most of its open questions.
Recorded here as constraints, not re-argued:

- **Staged Cockpit is the headline** (decision 1): my Direction 2, absorbing Direction 1
  (guardian console page + telemetry split into a systems tab, D10) and Direction 3
  (jump-to-source D3, villager strip D12, condition overlays) as its composition. The
  ranking question this analysis would otherwise adjudicate is closed.
- **Charter authoring: in-TUI read + `$EDITOR` write** (decision 2) — the Direction 4
  instinct ("never build a bad in-TUI text editor") constrains the console exactly as this
  branch's report recommended.
- **Stage shapes layout *defaults* only** (decision 3): everything reachable at every
  stage and via `?`; pre-ladder worlds get everything; capability locks stay angel-only.
  This resolves the doctrine-creep risk that made me rank the Staged Cockpit second.
- **`docs/design/tui/` v2 is the UI authority** (decision 4): pages/panels/overlays/patterns,
  uniform control tables with skin-token columns, `verified_against` pins, a gate.
- **Lesson row** (decision 5), **guardian strip** (decision 7), **exercise panel** (D11):
  the sibling-proposed surfaces adopted — reconciled below.
- **Both big moments seize the screen** (decision 6): the operator overrode the
  badge-until-pause recommendation. Notably, *this branch's corpus supports the override*:
  DF's major announcements "pause gameplay and center the camera" and some RimWorld events
  "pause the game and pop up a choice dialog" ([[Dwarf-Fortress-Interface]],
  [[RimWorld-Interface]]) — interruption for genuinely major events is established
  colony-sim practice, provided it is rare, dismissable, and replayable (all stipulated).
- **Input parity ratified** (decision 8), incremental rollout; **terminal-first with
  first-class linear streams** (D1) — closing this branch's renderer and parity questions.
- **Skin tokens first** (D2): TASK-121's token contract precedes new fiction literals,
  including in the design doc — matching this branch's finding that the design reference
  itself hard-codes the angel fiction.

## Reconciliation with the sibling reports

**Gameplay-patterns report.** Its core finding — "the scenario run, the ratified v1 game,
has no designed screen" — is the same verb-gap thesis as mine seen from the goal side
rather than the input side; the two are complementary, not competing. I **adopt** its
exercise panel (now D11) and its guardian-strip placement above the minibuffer (now
decision 7): pairing budget with input makes the minibuffer read as *the* verb, which is a
better fulfillment of my lens-(b) argument than my own header-segment alternative. Its
fiction/telemetry split of the metatron tab is the same move as my systems-tab proposal
(now D10) with a sharper rationale — the TASK-121 skin boundary needs the split or the
sweep is ambiguous control-by-control. Its staleness audit of `docs/design/tui/`
(dock.md still specifies a transcript-only metatron tab; INDEX rule 4 did not hold for
specs 044/046) *sharpens* rather than overturns my "extension, not replacement" verdict:
what survives is the anatomy, the patterns, and the mockup-first convention; what v2 must
add is reconciliation, an overlays/ class, per-tab panel files, and the gate (decision 4).
Its DF-Premium-guardrail caution and mine are identical.

**Learning-game-design report.** Its lesson-row proposal (persistent dwell surface with a
UI pointer, one active lesson, anti-spam, opportunity decay) initially collides with two
things I defended: the minibuffer's protected dormant state (the silent-tutorial opening)
and the exact-height row-budget discipline of `patterns/layout.md`. Reconciled: the lesson
row is a *display* row, not an input — the focus contract and "minibuffer is the only text
input" doctrine are untouched — and decision 5's stage-defaulting (always-visible at
stages 1–2, badge+overlay at 3+ and pre-ladder) spends the vertical budget exactly where
the corpus says pushed lessons are needed most and reclaims it where they are not. I adopt
its unlock-attribution fix ("your charter proved The Written Word", now D6) — it matches
this branch's own plain-language discipline better than the shipped "Metatron's watcher
earned…" line. Its ceremony recommendation (badge-until-pause) was overridden by decision
6; the flow/interruption evidence it rested on remains true and is carried below as a
residual tension, but the override is corpus-defensible from this branch's own DF/RimWorld
pause-and-center facts.

**Player-docs report.** Its design-doc program — control tables as the AI-parseable unit,
`verified_against` pins, a freshness check script, same-PR amendment enforcement — is the
concrete mechanism my lens-(d) verdict needed and is now decision 4; I adopt it wholesale,
with one emphasis from this branch: keep the mockup-first page convention and the
printable-card keymap, both corpus-backed, so the v2 doc stays human-scannable as well as
machine-queryable. Its finding that the deterministic help floor teaches *zero prompting*
(the `?` overlay covers keys/screen/lessons, never the guardian) strengthens my lens-(b)
gap from an angle I missed; the guardian overlay section (D9) closes it. Its skin-token
sequencing (D2) matches my TASK-121 design-doc finding exactly.

## Ranked recommendations under the decisions

With direction fixed, ranking shifts from *which direction* to *build order within it*:

1. **`docs/design/tui/` v2 first** (decision 4). Every other item below specifies into it;
   until it exists, "build from the design doc" is false for the metatron tab and the
   lens's clause (d) is unmet. Scope: reconcile shipped reality (specs 013–046), split
   dock.md into per-tab panels, add overlays/ (help, ceremony, postmortem) and the new
   surfaces (guardian console page, systems tab, exercise panel, guardian strip, lesson
   row, villager strip), control tables with skin-token columns, pins + gate. Record
   decision 8's parity rule in keymap.md.
2. **Skin-token contract** (D2, TASK-121 spec) before TASK-115/117 write new fiction
   literals — otherwise the sweep grows unboundedly and the v2 doc rots on day one.
3. **Guardian console + telemetry split** (decision 1 + D10 + decision 2): the full-height
   console page with document-style turns, report-card cards at stopping points (D5), the
   charter/skills read surface with binding status, `$EDITOR` handoff and the
   "charter changed — next turn binds it" confirmation. This is the lens-(b) payoff.
4. **Village Lens quick wins** (cheap, high yield, independent): chronicle `⏎`
   jump-to-source filling the reserved seam (D3), villager strip (D12), guardian strip
   (decision 7), incremental mouse targets under the parity doctrine (decision 8).
5. **Stage-default machinery + teaching surfaces** (decisions 3, 5, 6; D11): per-stage
   layout defaults, the lesson row, the exercise panel riding the scenario machinery, and
   the two takeover surfaces (ceremony on stage unlock, postmortem on run end), each with
   a linear-stream equivalent (the stages CLI, the morgue file) per D1.
6. **`?` overlay guardian section** (D9) — static-per-stage, model-free, the deterministic
   floor finally teaching the verb.

## Tensions & tradeoffs

Named unresolved tensions — synthesis input, not failures:

- **Vertical budget contention.** Decisions 5 + 7 add up to two new chrome rows above the
  minibuffer's fixed three, plus header, footer, and legend. On a 30-row terminal at
  stages 1–2 that is roughly a fifth of the screen spent on chrome. `patterns/layout.md`'s
  row arithmetic and the narrow-fallback breakpoint need re-derivation, and a minimum
  *height* breakpoint (which row collapses first: lesson row → guardian strip → legend?)
  must be specified in v2. Unresolved; no decision covers it.
- **Takeover salience vs. flow.** Decision 6 is corpus-defensible (DF/RimWorld
  pause-and-interrupt precedent) but the learning-game branch's evidence that mid-play
  interruption harms flow does not vanish; the mitigation burden moves to execution —
  takeovers must be rare by construction (two event types), instantly dismissable, and
  never able to stack. If live use shows the ceremony landing mid-crisis, the fallback is
  a pause-first variant of the same surface, not a redesign.
- **Pre-ladder worlds get the least pushed teaching.** Decision 5 defaults the lesson row
  off at stage 3+ *and* pre-ladder, but a pre-ladder world may host a brand-new player
  (the ambient default path), and the Cogmind caution — "hot pink and blinking… and still
  people sometimes miss them" — applies most to badge-only delivery. Mild; the `?`
  pull-reference floor and the every-world overlay hold, but worth watching.
- **Parity without a sweep.** The keymap's honesty is mechanized (`help_test.go` ties every
  advertised binding to a handler); mouse targets have no equivalent yet. Ratifying parity
  as doctrine (decision 8) without a mouse-target sweep risks the DF-classic failure mode
  reappearing in the click dimension. v2's control tables should carry a mouse column so a
  future sweep has something to check.
- **Accessibility drift.** More chrome rows, badges, and takeovers all worsen the
  screen-reader story the corpus documents ("a dumb, linear CLI stream is infinitely
  superior" — [[TUI-and-Roguelike-UI-Craft]]). D1 keeps `attach`/`tail` first-class; the
  discipline to hold is that every new surface (ceremony, postmortem, exercise state) has
  a linear-stream or CLI projection, specified in v2, not retrofitted.

## Confidence & open questions

Confidence is high on the evaluation (the corpus facts and the shipped-surface reads are
directly cited) and moderate-high on build order (decision-constrained; the main
uncertainty is how much of the stage-default machinery the scenario work forces early).

Genuinely open:

1. **Guardian turn rendering depth** — document-style fixed-measure turns with streaming
   affordances, or upgraded terse transcript lines? Not settled by decisions 1/2/D10;
   determines whether the console page is a rendering project or a layout project.
2. **Minimum-height behavior** — the collapse order for the new chrome rows and the
   narrow/short fallback matrix (see the budget tension above).
3. **Audience calibration** (carried from the synthesis's own open question 3) — does the
   `?` overlay's advanced tier and the v2 control tables' data-source column expose raw
   registry values (engineer audience) or stay plain-language-only?
4. **Fork-duel revisit criterion** — D7 defers the dual side-by-side TUI; what evidence
   (demand, scenario telemetry) would reopen it post-v1?
5. **Seen-lessons reset semantics** — D8 fixes the home (per-user, the unlocks.json
   precedent); what resets it (new skin? explicit command? never)?

## Basis

- [[_grounding]] — all six sections; every corpus claim above cites through its per-source note
- [[TUI-and-Roguelike-UI-Craft]] — Cogmind input parity, map dynamics, message-log discipline, accessibility hazards
- [[Chat-and-Agent-Console-Rendering]] — chat-turn conventions, Claude Code terminal craft, HITL cards, collapsible reasoning
- [[RimWorld-Interface]] — severity letters, click-to-jump, learning helper, pawn visual grammar, pause-dialog events
- [[Dwarf-Fortress-Interface]] — classic-UI cautionary tale, Premium redesign trade-offs, pause-and-center announcements, mods-as-gap-detectors
- [[Terraria-Interface]] — QoL waves, staged Bestiary/Journey-Mode disclosure, progress bars, input re-mappings
- [[LLM-Agent-Sim-Interfaces]] — inspector chain, replay-vs-live cost, per-agent inspection scale threshold
- [[Recurring-Interface-Patterns]] — the cross-source pattern map this analysis reasons across
- [[Brief-and-Assumptions]] — the branch's original framing; its observation-vs-command open question is closed by the ratified lens (prompting is the verb; observation is the feedback channel)
- Operator constraints: `docs/design/reorient-2026-07-25-ui.md` (decisions 1–8, D1–D13); project surfaces read in `docs/wiki/tui-client.md`, `docs/wiki/curriculum-ladder.md`, `docs/wiki/metatron.md`, `docs/wiki/chronicle.md`, `docs/design/tui/`, `docs/design/learning-game-synthesis.md` (cited by path — outside this vault by design)
