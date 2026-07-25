---
title: Brief and Assumptions
aliases: []
tags: [brief, game-ui, tui]
type: note
created: 2026-07-25
updated: 2026-07-25
related: ["[[Game-UI-UX]]"]
---

# Brief and Assumptions

## The request (restated)

> "Do a targeted evaluation of multiple game interfaces in the TUI world. Also pull information
> about the user interfaces, UI/UX and game design of/from the following sources/topics:
> Dwarf Fortress · RimWorld · Terraria · Smallworld LLM and Clones · ChatGPT / Claude Code
> console (agent back and forth and rendering of chat turns, etc). The goal is to find as much
> grounding information as we can about good UI and UX design for a game like this. DF is a big
> one."

## Assumptions

- **"A game like this"** = promptworld: a terminal-first world simulation observed/steered by a
  player, where LLM-driven agents live in a grid world — so the relevant lens is *TUI-rendered
  world views, information-dense sim interfaces, and agent-conversation rendering*, not generic
  game HUDs.
- **"Smallworld LLM and Clones"** is read as the Stanford **"Generative Agents" Smallville**
  demo (Park et al. 2023) and its open-source clones (AI Town, etc.) — interfaces for watching
  and inspecting LLM agents in a small simulated town. Not the Days of Wonder board game
  "Small World".
- **"Targeted evaluation"** — per the vault pipeline, this phase gathers and structures *facts*
  about each interface (what it does, how it works, what is documented about its reception).
  The opinionated evaluation/comparison is a follow-on `analyze-vault` pass on this branch.
- **Dwarf Fortress gets extra depth** ("DF is a big one"): both the classic keyboard/ASCII
  interface and the 2022 Steam "Premium" mouse/graphics redesign, since the before/after is
  itself high-value UX evidence.
- Terraria is not a TUI, but is in scope as requested — mined for inventory/crafting/HUD
  patterns that transfer.
- A sixth angle is added beyond the five named sources: **general TUI/roguelike UI design
  principles** (e.g. Cogmind's UI dev writing, roguelike UX literature), because it directly
  serves the stated goal and DF/RimWorld discussions constantly reference it.

## Open questions (flagged for the user)

- Should the eventual analysis optimize for *observation-mostly* play (watching agents) or
  *command-heavy* play (ordering agents around)? The gathering covers both.
- Is the target renderer a raw terminal (cells + ANSI) or a TUI-styled web/desktop surface?
  Facts gathered here try to stay renderer-agnostic.
