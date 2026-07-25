---
title: Brief and Assumptions
aliases: []
tags: [brief]
type: note
created: 2026-07-24
updated: 2026-07-24
related: ["[[Game-Player-Docs]]"]
---

# Brief and Assumptions

## The request (restated)

> "Add a vault that researches examples of user documentation for a game such as this. Not
> looking for inspiration for features. Just how to easily convey to a user how to use the
> game. Not a description of the game or how it works. Just… how do you play this game? What
> is the damn angel thing? That kind of thing. Check out RimWorld. Check out Dwarf Fortress.
> Check out docs for other TUI games."

## Constraints carried into the research

- **In scope:** how existing games *convey* play — quickstart guides, manuals, in-game help,
  "what does that thing on screen mean" explanations, controls references, first-session
  walkthroughs.
- **Out of scope:** game features, mechanics design, or anything that reads as feature
  inspiration; marketing-style descriptions of what a game *is*.
- **Named comparators:** RimWorld, Dwarf Fortress, and terminal/TUI games generally
  (NetHack, Angband, Cogmind, Cataclysm: DDA, Caves of Qud were used).

## Assumptions

- "A game such as this" = promptworld: a TUI-fronted ambient simulation where the player
  mostly reads a story feed and intervenes through an unfamiliar named entity (Metatron —
  "the damn angel thing"). The research therefore weights: (a) games with terminal
  interfaces, (b) games with steep unfamiliar-concept load, (c) docs that answer "what am I
  looking at?" and "what do I actually *do*?".
- The eventual consumer of this research is `docs/player/` (plain-language HTML player docs)
  and possibly in-app help — but this branch only gathers patterns; it recommends nothing.

## Open questions

- Whether promptworld's docs should be primarily out-of-game (web pages, like DF's wiki) or
  in-game (contextual help, like RimWorld/Cogmind) is a judgment call — deferred to a future
  analysis note.
