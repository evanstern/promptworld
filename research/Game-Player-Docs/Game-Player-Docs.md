---
title: Game-Player-Docs
aliases: [Player Documentation Patterns, How-To-Play Docs]
tags: [moc, player-docs, onboarding, tui-games]
type: moc
created: 2026-07-24
updated: 2026-07-25
related: []
---

# Game-Player-Docs

How complex simulation and TUI games (RimWorld, Dwarf Fortress, NetHack, Angband, Cogmind,
Cataclysm: DDA, Caves of Qud) write player-facing "how do I actually play this?"
documentation — controls, the first session, and "what is that thing on my screen?" — as
distinct from describing the game or its features.

## Scope

**In:** structure, tone, and delivery of how-to-play docs — quickstart guides, manuals,
in-game contextual help, symbol/screen explanations, controls references.
**Out:** game mechanics or feature design; marketing descriptions of what a game is;
any recommendation for promptworld's own docs (that's an analysis, not this branch).
Constraints and assumptions: [[Brief-and-Assumptions]].

## What is known

The corpus shows three distinct documentation layers, each with its own conventions:

- **The first-session walkthrough** — a scoped, do-this-then-this quickstart that teaches
  notation first, walks a "minimal viable session," reassures about failure up front, and
  hands off to deeper reference at the end ([[Quickstart-Guide-Patterns]]).
- **Screen orientation** — the player's own questions as headings (NetHack's "What do all
  those things on the screen mean?"), organized by screen region, with symbol tables the
  reader is told *not* to memorize because an in-game lookup key answers in place
  ([[Explaining-The-Screen]]).
- **In-flow help** — RimWorld's event-triggered, auto-retiring learning helper and Cogmind's
  four-layer hierarchy (reduce the need first → context help at the point of question →
  bounded tutorial → tiered reference) ([[In-Game-Contextual-Help]]).
- **The reference manual** — never mixes lore with controls, segments and tags for lookup,
  shows-then-tells with call-outs, ships a reference card, and stays honest either by being
  generated from game data or by delegating to in-game lookup
  ([[Manual-Structure-Conventions]]).

## Notes

- [[Brief-and-Assumptions]] — the request's constraints and the assumptions taken
- [[Quickstart-Guide-Patterns]] — first-session walkthrough structure and tone (DF, CDDA)
- [[Explaining-The-Screen]] — "what am I looking at?" patterns (NetHack, Angband, Rogue)
- [[In-Game-Contextual-Help]] — adaptive/contextual help systems (RimWorld, Cogmind, Qud, DF Steam)
- [[Manual-Structure-Conventions]] — reference-layer organization and separation rules

## Analyses

- [[Analysis-In-Game-First-Teaching]] — resolves the deferred in-game vs out-of-game
  question under the learning-game lens (verdict: in-game primary, out-of-game reference;
  operator-ratified 2026-07-25), with reconciled recommendations and proposed backlog moves

## Open questions

- How do games whose player mostly *observes* (god games, ambient sims) rather than issues
  frequent commands document the observe/intervene split? The corpus here is command-driven.
- What does the RimWorld learning helper's lesson *content* look like (length, structure per
  lesson)? The wiki page describes the trigger system, not lesson anatomy.
- ~~Whether promptworld's docs should be primarily out-of-game or in-game~~ — resolved in
  [[Analysis-In-Game-First-Teaching]] (in-game primary, out-of-game reference).

## Grounding

- [[_grounding]] — the research pass this branch is built on (web-search fan-out, 2026-07-24)
- [DF Quickstart guide](https://dwarffortresswiki.org/Quickstart_guide)
- [NetHack 3.6 Guidebook](https://www.nethack.org/v367/Guidebook.html)
- [Grid Sage Games: Tutorials and Help](https://www.gridsagegames.com/blog/2016/07/tutorials-help/)
- [RimWorld Wiki: Learning helper](https://rimworldwiki.com/wiki/Learning_helper)
