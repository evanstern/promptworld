---
title: Explaining The Screen
aliases: [What Do All Those Things Mean]
tags: [screen-orientation, symbols, whatis]
type: note
created: 2026-07-24
updated: 2026-07-24
related: ["[[Game-Player-Docs]]", "[[Quickstart-Guide-Patterns]]", "[[In-Game-Contextual-Help]]"]
---

# Explaining The Screen

How TUI games answer "what am I looking at?" — the exact class of question as "what is the
damn angel thing?". The reference example is the NetHack Guidebook, whose chapter 3 is
literally titled **"What do all those things on the screen mean?"** ([[_grounding]] § NetHack).

## The NetHack pattern: questions as headings, regions as structure

- Chapters 2 and 3 of the Guidebook are the player's own questions used verbatim as
  headings: *"What is going on here?"* then *"What do all those things on the screen mean?"*
  ([[_grounding]]).
- The screen chapter is organized **by screen region**, not by concept: the status lines
  (bottom), the message line (top), the map (everything else). The reader locates the region
  they're confused about and reads that subsection ([[_grounding]]).
- Symbols are enumerated in a reference table (`.` floor, `#` corridor, `|`/`−` walls…) but
  the table is prefaced with **"You need not memorize all these symbols"** — because the
  game answers in place ([[_grounding]]).

## Docs delegate to the game: the whatis convention

NetHack's Guidebook teaches two in-game commands instead of asking the reader to memorize
the symbol table: `/` (whatis) explains any symbol or word, `;` (glance) identifies a
visible thing quickly ([[_grounding]]). The document's job shrinks to *teaching the one key
that answers all future "what is that?" questions*. Rogue established the same convention
with `?` for the command list ([[_grounding]] § Classic Rogue).

## The opposite pole: Angband's command-first manual

Angband's *Playing the Game* page skips screen orientation entirely — it opens with command
architecture, keyset choice (original vs roguelike/vi keys), and notation, then gives
alphabetical command tables ([[_grounding]] § Angband). This works as a reference for
players already oriented, and shows the two poles a manual can start from: **"here is what
you see"** (NetHack) vs **"here is what you can do"** (Angband).

## Keys that document themselves

RogueBasin records the genre convention of **mnemonic bindings as self-documentation**: `q`
quaff, `w` wear/wield, `e` eat — plus NetHack's `#` extended commands that are typed out as
words (#dip, #loot) rather than memorized ([[_grounding]] § Classic Rogue). The binding
scheme itself reduces what the docs must carry.

## Grounding

- [[_grounding]] — §§ "NetHack", "Angband", "Classic Rogue and roguelike documentation conventions"
- [NetHack 3.6 Guidebook](https://www.nethack.org/v367/Guidebook.html)
- [Angband Manual: Playing the Game](https://angband.readthedocs.io/en/latest/playing.html)
- [RogueBasin: Preferred Key Controls](https://www.roguebasin.com/index.php/Preferred_Key_Controls)
