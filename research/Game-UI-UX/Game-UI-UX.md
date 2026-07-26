---
title: Game-UI-UX
aliases: [TUI Game Interfaces, Game Interface Design Research]
tags: [game-ui, tui, ux, moc]
type: moc
created: 2026-07-25
updated: 2026-07-26
related: []
---

# Game-UI-UX

What is documented about good UI/UX design for information-dense simulation games and agent
interfaces — surveyed across Dwarf Fortress, RimWorld, Terraria, LLM-agent world sims
(Smallville and clones), conversational AI consoles (ChatGPT, Claude Code), and the general
TUI/roguelike interface craft. Gathered as grounding for designing the interface of a
terminal-first LLM-agent world game.

## Scope

**In:** how each surveyed interface presents world state, entity internals, events, and time;
how each handles input, discoverability, and teaching; documented developer rationale and
reception evidence (including what community mods reveal about native gaps); terminal/text-grid
rendering constraints and accessibility facts.

**Out:** verdicts on which patterns promptworld should adopt (that's an analysis note);
gameplay/narrative design (covered by the separate Game-Gameplay-Patterns branch); player
documentation conventions (Game-Player-Docs branch).

## What is known

- **Dwarf Fortress** is the depth extreme: a one-codebase text grid whose classic keyboard UI
  was the canonical usability cautionary tale, whose 2022 mouse-first redesign sold a million
  copies while drawing its own criticisms, and whose community tooling (Dwarf Therapist,
  DFHack) maps the native UI's gaps — [[Dwarf-Fortress-Interface]].
- **RimWorld** is the documented accessible counterpoint: fixed screen regions, severity-coded
  letters with click-to-jump, a contextual learning helper, and a precise visual grammar for
  pawn internals, backed by Tynan Sylvester's published design philosophy —
  [[RimWorld-Interface]].
- **Terraria** shows a decade of QoL waves converging on the same endpoints (search,
  batch operations, contextual lookup, staged codex unlocks) and two documented input
  re-mappings (console, mobile) — [[Terraria-Interface]].
- **LLM-agent sims** (Smallville, AI Town, Project Sid) establish the inspector patterns for
  watching agents — ambient glyph → click → full state → memory stream — plus replay-vs-live
  cost trade-offs and the scale threshold where dashboards replace per-agent views —
  [[LLM-Agent-Sim-Interfaces]].
- **Chat/agent consoles** document turn rendering: streaming as trust signal (and its overload
  failure modes), document-style turns, collapsed tool/thinking blocks, interrupt-with-work-
  preserved, human-in-the-loop approval cards, and how Claude Code renders all this in a
  terminal at 60fps — [[Chat-and-Agent-Console-Rendering]].
- **TUI/roguelike craft** supplies the substrate rules: input parity, multi-channel feedback,
  map-first information design, small importance-filtered logs, terminal cell geometry and
  color limits, ASCII's documented strengths/costs, and TUI screen-reader accessibility
  hazards — [[TUI-and-Roguelike-UI-Craft]].
- Patterns recurring independently across all six areas (event feeds, progressive disclosure,
  time controls, contextual teaching, automation of tedium, mods-as-gap-detectors, scale
  thresholds) are mapped in [[Recurring-Interface-Patterns]].
- The request, assumptions, and reading of "Smallworld LLM" are recorded in
  [[Brief-and-Assumptions]].

## Notes

- [[Brief-and-Assumptions]] — the request restated, assumptions, open questions for the user
- [[Dwarf-Fortress-Interface]] — classic ASCII UI, Steam Premium redesign, community tooling layer
- [[RimWorld-Interface]] — screen regions, alerts/letters, learning helper, pawn visual grammar
- [[Terraria-Interface]] — inventory/crafting QoL waves, HUD, Guide/Bestiary, input re-mappings
- [[LLM-Agent-Sim-Interfaces]] — Smallville, AI Town, Project Sid: watching and steering LLM agents
- [[Chat-and-Agent-Console-Rendering]] — chat-turn conventions, Claude Code's terminal UI, agent-loop patterns
- [[TUI-and-Roguelike-UI-Craft]] — Cogmind/DCSS/NetHack principles, message logs, terminal constraints, accessibility
- [[Recurring-Interface-Patterns]] — the cross-source pattern map

## Analyses

- [[Analysis-Tile-Vocabulary-Expansion]] — 2026-07-26: what the corpus supports for
  expanding the tile/glyph vocabulary — new tile types from the roguelike/CP437 convention
  dictionary, a semantic (severity-first, token-externalized) color grammar, letter avatars
  with emoji confined to the inspector layer, and fonts/tilesets as swappable skins over one
  grid. Names the corpus gaps (no concrete font candidates, no tested colorblind palette).
  Rendered as `tile-vocabulary-expansion-briefing.html`, published at
  <https://claude.ai/code/artifact/05415ad3-3efd-4693-9330-c626f4435731>.
- [[Analysis-Teaching-Game-TUI]] — the 2026-07-25 reorientation evaluation: the current
  TUI as watching instrument vs the teaching-game lens, the Staged Cockpit direction
  under the operator's ratified decisions, and the reconciled cross-branch position.
  Rendered (merged with the run's sibling analyses) as the briefing page
  `docs/design/reorient-2026-07-25-ui-briefing.html` (outside the vault — the page
  carries cross-branch synthesis content), published at
  <https://claude.ai/code/artifact/4ae421da-7494-49e9-aafc-3f838f09c15a>.

## Open questions

- Should the eventual analysis weight observation-mostly play (watching agents) or
  command-heavy play (ordering agents)? Both are grounded here.
- Is the target renderer a raw terminal or a TUI-styled web/desktop surface? Accessibility
  facts differ sharply between them.
- Cogmind's and RimWorld's numbers-heavy visual grammars are documented; no source yet
  grounds how such grammars perform for *natural-language* state (agent thoughts/dialogue)
  at density — a possible follow-up research pass.

## Grounding

- [[_grounding]] — six-section web-search fan-out (2026-07-25): Dwarf Fortress, RimWorld,
  Terraria, LLM-agent sims, chat/agent consoles, TUI/roguelike craft.
- Key primary sources: [Tarn Adams interviews](https://www.gamedeveloper.com/programming/how-tarn-adams-upgraded-and-optimized-dwarf-fortress-for-its-official-steam-release),
  [Tynan Sylvester's Alpha 15 post](https://ludeon.com/blog/2016/08/alpha-15-tutorial-and-drugs-released/),
  [Grid Sage Games UI series](https://www.gridsagegames.com/blog/2015/04/cogmind-roguelike/),
  [Generative Agents paper](https://ar5iv.labs.arxiv.org/html/2304.03442),
  [Claude Code docs](https://code.claude.com/docs/en/interactive-mode).
