---
title: Manual Structure Conventions
aliases: [Reference Doc Patterns]
tags: [manuals, reference, structure]
type: note
created: 2026-07-24
updated: 2026-07-24
related: ["[[Game-Player-Docs]]", "[[Quickstart-Guide-Patterns]]", "[[In-Game-Contextual-Help]]"]
---

# Manual Structure Conventions

What the corpus shows about how the *reference* layer of game docs is organized — the
document a player returns to, as distinct from the first-session walkthrough
([[Quickstart-Guide-Patterns]]) and in-flow help ([[In-Game-Contextual-Help]]).

## Separation rules from instructional-design sources

- **Never mix lore/background with controls.** "Authors should resist mixing background
  information with controls, as such mixtures complicate a gamer's referencing and make a
  complicated game seem that much more complex" ([[_grounding]] § General manual-writing).
  The reference layer answers "how do I", not "what is this game about".
- **Segment into logical, tagged sections** — "information is easier to reference when it is
  neatly tagged, described and categorized"; a common outline pattern serves sim/strategy
  games generally ([[_grounding]]).
- **Show first, then tell**: for complex screens, "screenshots with call-outs" *start* the
  instructions, with descriptions per call-out following ([[_grounding]]).
- **Reference cards for cognitive load**: complex games "rely on reference cards for players
  to remember chunks and their contents" — a one-page keys/commands card is a standard
  artifact, not an extra ([[_grounding]]).

## Observed manual shapes in the corpus

- **NetHack Guidebook**: 12 chapters ordered from orientation ("What is going on here?",
  "What do all those things on the screen mean?") → Commands → world-content chapters
  (rooms, monsters, objects) → meta (options, scoring). Player-question headings up front,
  taxonomy behind ([[_grounding]] § NetHack).
- **Angband manual**: notation and keyset choice first, then alphabetical command tables per
  keyset, then interaction systems. Pure reference; assumes orientation happened elsewhere
  ([[_grounding]] § Angband).
- **Cogmind manual**: ~8,000 words, lives *inside* the game, browsable and searchable; the
  command reference is split into basic vs advanced pages so newcomers aren't shown the full
  surface at once ([[_grounding]] § Cogmind).
- **Tiered corpus around the game**: DF and CDDA both rely on an external wiki as the deep
  reference, with the quickstart explicitly handing off to it ("What Next?" section); CDDA
  additionally has a **generated-from-game-data reference** (the Hitchhiker's Guide, built
  from the game's own JSON, versioned per release) that stays accurate where hand-written
  pages rot ([[_grounding]] §§ Dwarf Fortress, Cataclysm).

## Keeping reference honest

Two mechanisms appear for keeping reference docs true to the running game: generating them
from game data (CDDA's Hitchhiker's Guide), and putting the canonical explanation *in* the
game so the doc only teaches the lookup command (NetHack's `/` whatis; Cogmind's context
help) ([[_grounding]]).

## Grounding

- [[_grounding]] — §§ "General manual-writing / instructional-design findings", "NetHack",
  "Angband", "Cogmind", "Cataclysm: Dark Days Ahead"
- [Game Developer: Manuals: They Can Be Good](https://www.gamedeveloper.com/production/manuals-they-can-be-good)
- [Analog Game Studies: Rules for Writing Rules](https://analoggamestudies.org/2014/10/the-rules-for-writing-rules-how-instructional-design-impacts-good-game-design/)
