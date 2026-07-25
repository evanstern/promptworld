---
title: TUI and Roguelike UI Craft
aliases: [Roguelike Interface Design, Terminal UI Principles]
tags: [tui, roguelike, ascii, message-log, accessibility]
type: note
created: 2026-07-25
updated: 2026-07-25
related: ["[[Game-UI-UX]]", "[[Dwarf-Fortress-Interface]]", "[[Chat-and-Agent-Console-Rendering]]", "[[Recurring-Interface-Patterns]]"]
---

# TUI and Roguelike UI Craft

The documented craft knowledge of building text-grid game interfaces — Cogmind's UI writing
(Josh "Kyzrati" Ge), NetHack/DCSS conventions, the message-log problem, terminal constraints,
and ASCII as a rendering medium. Facts cited in [[_grounding]] §TUI & Roguelike Interface
Design Principles.

## Cogmind's documented principles (Josh Ge)

- ASCII as "a simple easily readable representation designed to facilitate decision-making";
  every item gets custom ASCII art; tiles ship as the accessibility alternative
  ([Cogmind the Roguelike](https://www.gridsagegames.com/blog/2015/04/cogmind-roguelike/)).
- **Input parity as a hard rule**: every action reachable by both mouse and keyboard; hotkeys
  embedded visibly in the UI so they're discoverable without memorization ([same](https://www.gridsagegames.com/blog/2015/04/cogmind-roguelike/)).
- **Redundant multi-channel feedback**: important events delivered simultaneously via color,
  symbol, sound, and animation (low health = audio warning + red ALERT + oscillating frame)
  ([same](https://www.gridsagegames.com/blog/2015/04/cogmind-roguelike/)).
- **Map-first information design**: push information out of the log onto the map itself via
  animated overlays and labels ("map dynamics") ([Map Dynamics](https://www.gridsagegames.com/blog/2014/02/map-dynamics/)).
- **The grid as architecture**: a fixed 160×60 cell grid; map cells occupy two terminal cells
  to read square; gameplay is balanced against the guaranteed 50×50 visible map area; scaling
  means changing cell size, never grid extent ([Full UI Upscaling Part 1](https://www.gridsagegames.com/blog/2024/01/full-ui-upscaling-part-1-history-and-theory/)).
- Mixed font dimensions: narrow fonts for text panels, wide/square for the map — "the best of
  both worlds" ([Fonts in Roguelikes](https://www.gridsagegames.com/blog/2014/09/fonts-in-roguelikes/)).

## The message log problem

- Cogmind keeps the log deliberately small (6 lines) so it stays "a secondary source of
  information while the player focuses attention on the map," expandable to full height;
  messages are 4–5 words with the important word first; routine events are never logged
  ([Message Log](https://www.gridsagegames.com/blog/2014/02/message-log/)).
- NetHack shows one message line with `--More--` gating when text would scroll past, plus
  Ctrl+P recall; `MSGTYPE` config lets users force, suppress, or dedupe specific messages
  ([NetHack wiki: Message](https://nethackwiki.com/wiki/Message), [MSGTYPE](https://nethackwiki.com/wiki/MSGTYPE)).
- DCSS documents message coalescing ("You hit the imp! You freeze the imp! The imp dies!" →
  "You freeze the imp to death!") ([DCSS dev wiki](https://crawl.develz.org/wiki/doku.php?id=dcss:brainstorm:interface:interface_implementables)).

## DCSS's explicit interface philosophy

- The manual states: "The interface is radically designed to make gameplay easy — this sounds
  trivial, but we mean it," and "all tedious, but necessary, chores should be automated"
  (autoexplore, Autofight); goals are meaningful decisions and no grinding
  ([crawl manual](https://github.com/crawl/crawl/blob/master/crawl-ref/docs/crawl_manual.rst)).
- Commentary credits autoexplore with "quickly getting the player to the interesting
  decisions that are the meat of the game" ([Zhang essay](https://www.jorgezhang.com/2020/06/dungeon-crawl-stone-soup-the-greatest-roguelike-of-all-time-and-what-it-can-tell-us-about-game-design/index.html)).

## Terminal constraints and ASCII as a medium

- Terminal cells are ~2:1 tall:wide — square-looking layouts need doubled horizontal gutters
  or two-cell map tiles; terminals expose 16 themeable ANSI colors under 256/truecolor modes
  ([Textual grid styles](https://textual.textualize.io/styles/grid/), [styles guide](https://textual.textualize.io/guide/styles/)).
- Glyph vocabulary is a shared convention (`@` you, `.` floor, `<`/`>` stairs, letters =
  creatures), dating to Rogue; CP437 supplies the box-drawing/shading chrome used by DF
  ([RogueBasin](https://www.roguebasin.com/index.php/User_interface_features);
  [Fonts in Roguelikes](https://www.gridsagegames.com/blog/2014/09/fonts-in-roguelikes/)).
- Ge's documented ASCII trade-off: glyphs stay identifiable densely clustered at small sizes
  and "distill information for tactical decision-making into its purest form," but many
  players can't make the conceptual leap and "the average player will always choose to play
  with a tileset" ([ASCII vs Tiles](https://www.gridsagegames.com/blog/2015/02/ascii-vs-tiles/)).
- Key conventions: vi-keys vs numpad movement (support both), mnemonic command letters,
  `#` extended-command prefix for rare actions ([RogueBasin: Preferred Key Controls](https://www.roguebasin.com/index.php/Preferred_Key_Controls)).

## Accessibility

- A widely-circulated essay documents that modern reactive TUIs (Ink, Bubble Tea, tcell named)
  are hostile to screen readers — cursor teleporting, full-screen re-renders, no accessibility
  tree — concluding "for the blind user, a dumb, linear CLI stream is infinitely superior to
  a 'smart' TUI"; documented mitigations: hidden cursor, VT100 scroll regions, single vertical
  list layouts ([The text mode lie](https://xogium.me/the-text-mode-lie-why-modern-tuis-are-a-nightmare-for-accessibility)).
- Cogmind ships an audio log and visible SFX for hearing-impaired players; Ge's guidance is
  that some accessibility must be designed in from the start
  ([ModDB](https://www.moddb.com/games/cogmind/news/audio-accessibility-features-for-roguelikes);
  [How to Make a Roguelike](https://www.gridsagegames.com/blog/2018/10/how-to-make-a-roguelike/)).

## Grounding

- [[_grounding]] — §"TUI & Roguelike Interface Design Principles" (all claims above).
