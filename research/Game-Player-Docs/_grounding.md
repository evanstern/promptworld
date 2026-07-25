---
title: Game Player Docs — Grounding
aliases: []
tags: [grounding]
type: source
created: 2026-07-24
updated: 2026-07-24
related: ["[[Game-Player-Docs]]"]
---

# Game Player Docs — Grounding

> Source-of-truth artifact. This is the raw, cited output of a research pass (the `deep-research`
> skill, or a direct web-search fan-out). Keep it close to verbatim — do not editorialize, prune,
> or draw conclusions here. Knowledge notes and analyses cite *into* this file.

**Research question:** How do complex simulation and TUI games (RimWorld, Dwarf Fortress,
NetHack, Angband, Cogmind, Cataclysm: DDA, Caves of Qud) structure player-facing "how do I
actually play this" documentation — controls, first session, and explaining what things on
screen mean — as opposed to describing the game?
**Method:** web-search fan-out (8 parallel searches + 4 direct page fetches) · 2026-07-24

---

## RimWorld — the learning helper (adaptive in-game lessons, no manual)

- RimWorld ships **no traditional manual**; it teaches through an "intelligent learning
  helper" plus an optional scripted tutorial, both added in Alpha 15 (2016). The tutorial
  "walks you through the basics of managing a colony," after which "the adaptive learning
  helper starts up" ([Ludeon Alpha 15 announcement](https://ludeon.com/blog/2016/08/alpha-15-tutorial-and-drugs-released/),
  [PCGamesN coverage](https://www.pcgamesn.com/rimworld/rimworld-alpha-15-patch)).
- The learning helper **sits in the top-right of the screen**; it replaced earlier yellow
  pop-up messages and learning alerts ([RimWorld Wiki: Learning helper](https://rimworldwiki.com/wiki/Learning_helper)).
- Trigger model: "if something happens relating to a concept the player hasn't learned, that
  lesson will be activated and shown on the learning helper," where it can be opened and
  read. Lessons are **auto-marked as learned when the player performs the interaction**, can
  be manually dismissed, and surface "as needed by circumstance, or on a slow timer"
  ([RimWorld Wiki: Learning helper](https://rimworldwiki.com/wiki/Learning_helper)).
- Stated design goal: "you'll never be shown a lesson you already know, but always be shown
  the lessons you need to know now" ([RimWorld Wiki: Learning helper](https://rimworldwiki.com/wiki/Learning_helper)).
- The helper is also a **pull reference**: it "can be expanded and searched for any lesson,
  so you can look up how to do a specific thing at any time"
  ([RimWorld Wiki: Learning helper](https://rimworldwiki.com/wiki/Learning_helper)).
- Everything deeper (mechanics, numbers) lives on the community
  [RimWorld Wiki](https://rimworldwiki.com/), not in official docs.

## Dwarf Fortress — the community Quickstart guide + Steam tutorial

Source: [DF Wiki Quickstart guide](https://dwarffortresswiki.org/Quickstart_guide) (fetched
2026-07-24), plus [DF Wiki: Tutorials](https://www.dwarffortresswiki.org/Tutorials).

- The Quickstart guide is explicitly scoped: "for dwarf fortress mode for those who have
  never played before and quickly want to jump in head-first."
- **Section order:** (1) Common UI Concepts — keybindings, menu navigation, interface
  symbols; (2) World Generation; (3) pointer to the in-game Tutorial; (4) Embark — site
  selection and preparation; (5) **A Minimal Fortress** — entry, storage, food, workshops;
  (6) Beyond a Minimal Fortress; (7) Military (brief); (8) "What Next?" — pointers to
  further reading.
- **Failure reassurance up front:** it opens with "Always remember that **losing is fun!**"
  and "Be prepared to lose a few fortresses before you get all the way through this guide…
  losing means that next time, *you'll remember how you lost*."
- **Acknowledges reader ignorance directly:** phrasing like "you have no idea what to do"
  followed by "That's understandable."
- **Notation is taught explicitly:** e.g. "t means 'press the t key without the shift key'"
  vs "T means 'hold down shift and press the t key.'" Interface objects are defined in plain
  terms as they appear (a stockpile is "where your dwarves will drop things for storage").
- **Why, not just how:** systems get contextual rationale (refuse stays outside "to avoid
  miasma"), and essential steps are distinguished from optional depth.
- The **Steam (Premium) version added a basic in-game tutorial and a good help section**;
  the in-game tutorial "is quite good and strongly suggested for new players, as it will
  automatically choose a site for the fortress as well as dwarves and supplies"
  ([DF Wiki Quickstart guide](https://dwarffortresswiki.org/Quickstart_guide)). Even so,
  community consensus is "you're gonna need to spend some time with your nose in a wiki"
  ([Pro Game Guides comparison](https://progameguides.com/dwarf-fortress/all-differences-between-dwarf-fortress-classic-and-dwarf-fortress-steam/),
  [Steam community guides](https://steamcommunity.com/app/975370/guides/)).
- The wiki maintains **separate quickstarts per audience/mode**: fortress mode, an Adventure
  mode quick start, and a Military quickstart
  ([DF Wiki Quickstart guide](https://dwarffortresswiki.org/Quickstart_guide)).

## NetHack — the Guidebook and in-game "whatis"

Source: [NetHack 3.6 Guidebook](https://www.nethack.org/v367/Guidebook.html) (fetched
2026-07-24); [NetHack Wiki: Guidebook](https://nethackwiki.com/wiki/Guidebook).

- The Guidebook ("A Guide to the Mazes of Menace") is NetHack's canonical manual — "the most
  important documentation included in NetHack itself, introducing map symbols, keyboard
  controls, and options in more detail than the help files"
  ([NetHack Wiki](https://nethackwiki.com/wiki/Guidebook)).
- **Chapter order:** 1. Introduction · 2. **"What is going on here?"** · 3. **"What do all
  those things on the screen mean?"** (status lines, message line, the map) · 4. Commands ·
  5. Rooms and corridors · 6. Monsters · 7. Objects · 8. Conduct · 9. Options · 10. Scoring ·
  11. Explore mode · 12. Credits. Chapters 2–3 are literally the player's questions used as
  headings.
- **Screen-region walkthrough:** chapter 3 explains the display by region — the status lines
  (bottom), the message line (top), and the map (the rest of the screen).
- **Symbols with reassurance:** default symbols are enumerated (`.` floor/ice/doorless
  doorway, `#` corridor/iron bars/tree/kitchen sink/drawbridge, `−` and `|` walls…), but the
  Guidebook says "You need not memorize all these symbols" — because **the game itself
  answers**: the `/` (whatis) command explains any symbol ("You may choose to specify a
  location or type a symbol (or even a whole word) to explain") and `;` (glance) is the
  quick "what type of thing is that visible symbol" command.
- Tone: the introduction is **narrative and welcoming** — it orients through an adventurer's
  framing before any technical instruction.

## Angband — command-first manual

Source: [The Angband Manual: Playing the Game](https://angband.readthedocs.io/en/latest/playing.html)
(fetched 2026-07-24).

- Structured **task-first, not screen-first**: it opens with the command architecture
  (commands + arguments), the two keyset options (original vs roguelike/vi keys), and
  notation for control keys, then gives **alphabetical command summary tables per keyset**,
  then behavioral systems (disturbance, repeat counts, targeting, pathfinding).
- Procedural, example-heavy tone: "To enter a repeat count, type `0`…" with clarifications
  in parentheses and heavy cross-referencing.
- Demonstrates the opposite pole from NetHack: it skips screen-layout orientation entirely.

## Cogmind — layered help in a modern terminal-style roguelike

Source: [Grid Sage Games: "Tutorials and Help: Easing Players into the Game"](https://www.gridsagegames.com/blog/2016/07/tutorials-help/)
(fetched 2026-07-24); [Cogmind manual PDF](https://cdn.steamstatic.com/steam/apps/722730/manuals/manual.pdf?t=1600161133);
[Cogmind FAQ](https://www.gridsagegames.com/cogmind/faq.html).

- First principle stated by the developer: **"The easiest way to help a player is to make it
  so they don't actually need help in the first place"** — reduce the need for docs via
  intuitive controls and self-explanatory UI labeling before writing any docs.
- **Four help layers:** (1) an ~8,000-word **in-game interactive manual**, browsable and
  searchable without leaving the game; (2) **context help** — click/keyboard onto any stat
  or UI element for an immediate explanation, because "specific questions are best handled
  at the point where they're asked"; (3) a **tutorial** — a static four-room intro level
  shown only on a player's first three runs, teaching through sequenced encounters rather
  than modal pop-ups; (4) **tiered command references** via `?` — separate *basic* and
  *advanced* command pages "preventing new players from being overwhelmed."
- Honest observation about player attention: "tutorial messages appear in the message log,
  hot pink and blinking… and still people sometimes miss them."
- Broader stance: "new roguelikes need lower barriers to entry so that those who might enjoy
  the game — but not learning it — make it through that first stage"; Cogmind aims for
  "dynamic depth" — playable in 15 minutes without details, deeper stats available on demand
  ([Grid Sage Games blog](https://www.gridsagegames.com/blog/2015/04/cogmind-roguelike/)).

## Cataclysm: Dark Days Ahead — wiki "Getting started" + data-driven reference

- Community docs split into a wiki [Getting started](https://cataclysmdda.miraheze.org/wiki/Getting_started)
  page (first-day survival walkthrough) plus tiered community [guides](https://cddawiki.danmakudan.com/wiki/index.php/Guides)
  "ranging from basic tutorials to advanced guides."
- The **[Hitchhiker's Guide to the Cataclysm](https://github.com/nornagon/cdda-guide)** is a
  reference whose "data comes directly from the JSON files in the game itself," offline-capable
  and versioned per game release — a generated-from-source reference that stays current where
  hand-written wikis rot.

## Caves of Qud — retrofitting a tutorial onto a dense sim (1.0, 2024)

- For years "the only problem with actually enjoying Caves of Qud is learning to play Caves
  of Qud, which was once relegated to forums, YouTube videos, or fan-made resources"; the
  1.0 release added a beginner tutorial so "a high barrier to entry just got a lot more
  surmountable" ([Destructoid](https://www.destructoid.com/fan-favorite-roguelike-caves-of-qud-mercifully-receives-tutorial-ahead-of-december-release/)).
- 1.0 paired the tutorial with **an interface overhaul that "modernized menus, clarified
  tooltips, and improved readability without diluting depth"** — docs and UI were fixed
  together, not separately ([Game Critix review](https://gamecritix.co.uk/caves-of-qud-review/)).
- Limit acknowledged: "while tutorials help, true mastery demands commitment" — the tutorial
  gets players started; depth still requires ongoing external resources
  ([Game Critix review](https://gamecritix.co.uk/caves-of-qud-review/)).

## Classic Rogue and roguelike documentation conventions

- Rogue's in-game convention: typing `?` brings up the list of all commands; F1/Shift+/ as
  alternates ([Rogue instruction manual PDF](https://shared.fastly.steamstatic.com/store_item_assets/steam/apps/1443430/manuals/Rogue_-_Instruction_Manual.pdf?t=1651574702),
  [Steam discussion](https://steamcommunity.com/app/1443430/discussions/0/3011179244462753211/)).
- **Mnemonic keybindings as self-documentation**: "a general RL theme in keyboard controls
  is the use of mnemonic bindings — 'q' for quaff, 'w' for wear/wield, 'e' for eat"
  ([RogueBasin: Preferred Key Controls](https://www.roguebasin.com/index.php/Preferred_Key_Controls)).
- NetHack's `#` extended-command mode spells out rare commands (#dip, #jump, #loot) rather
  than burdening the key map ([RogueBasin](https://www.roguebasin.com/index.php/Preferred_Key_Controls)).
- Vi-keys (hjkl movement) are the traditional convention inherited from the original Rogue
  ([RogueBasin](https://www.roguebasin.com/index.php/Preferred_Key_Controls)).

## General manual-writing / instructional-design findings

- **Segment content into logical sections**; a well-tested general outline serves adventure,
  RPG, simulation and strategy games ([Game Developer: "Manuals: They Can Be Good"](https://www.gamedeveloper.com/production/manuals-they-can-be-good)).
- **Do not mix background/lore with controls**: "authors should resist mixing background
  information with controls, as such mixtures complicate a gamer's referencing and make a
  complicated game seem that much more complex"
  ([Game Developer](https://www.gamedeveloper.com/production/manuals-they-can-be-good)).
- **Show as well as tell**: screenshots with call-outs should *start* operating
  instructions for complex systems, followed by descriptions per call-out
  ([Game Developer](https://www.gamedeveloper.com/production/manuals-they-can-be-good)).
- **Reference cards for high cognitive load**: complex games "rely on reference cards for
  players to remember chunks and their contents"
  ([Analog Game Studies: Rules for Writing Rules](https://analoggamestudies.org/2014/10/the-rules-for-writing-rules-how-instructional-design-impacts-good-game-design/)).
- Information is "easier to reference when it is neatly tagged, described and categorized"
  ([Game Developer](https://www.gamedeveloper.com/production/manuals-they-can-be-good)).

## Sources

- https://ludeon.com/blog/2016/08/alpha-15-tutorial-and-drugs-released/ — RimWorld Alpha 15: tutorial + adaptive learning helper announcement
- https://www.pcgamesn.com/rimworld/rimworld-alpha-15-patch — press coverage of the same
- https://rimworldwiki.com/wiki/Learning_helper — Learning helper mechanics (trigger model, search, auto-learned)
- https://rimworldwiki.com/wiki/Learning_assistant — earlier learning assistant page
- https://dwarffortresswiki.org/Quickstart_guide — DF Quickstart guide (structure, tone, notation teaching)
- https://www.dwarffortresswiki.org/Tutorials — DF tutorials index
- https://progameguides.com/dwarf-fortress/all-differences-between-dwarf-fortress-classic-and-dwarf-fortress-steam/ — Steam vs classic onboarding
- https://steamcommunity.com/app/975370/guides/ — DF Steam community guides
- https://www.nethack.org/v367/Guidebook.html — NetHack 3.6 Guidebook (chapter structure, whatis/glance)
- https://nethackwiki.com/wiki/Guidebook — Guidebook meta-description
- https://angband.readthedocs.io/en/latest/playing.html — Angband manual, Playing the Game
- https://www.gridsagegames.com/blog/2016/07/tutorials-help/ — Cogmind: Tutorials and Help design post
- https://www.gridsagegames.com/blog/2015/04/cogmind-roguelike/ — Cogmind design overview (dynamic depth)
- https://cdn.steamstatic.com/steam/apps/722730/manuals/manual.pdf?t=1600161133 — Cogmind manual PDF
- https://www.gridsagegames.com/cogmind/faq.html — Cogmind FAQ
- https://github.com/nornagon/cdda-guide — Hitchhiker's Guide to the Cataclysm (data-driven reference)
- https://cataclysmdda.miraheze.org/wiki/Getting_started — CDDA Getting started
- https://cddawiki.danmakudan.com/wiki/index.php/Guides — CDDA community guides index
- https://www.destructoid.com/fan-favorite-roguelike-caves-of-qud-mercifully-receives-tutorial-ahead-of-december-release/ — Caves of Qud 1.0 tutorial
- https://gamecritix.co.uk/caves-of-qud-review/ — Qud 1.0 UI + tutorial assessment
- https://www.roguebasin.com/index.php/Preferred_Key_Controls — roguelike keybinding conventions
- https://shared.fastly.steamstatic.com/store_item_assets/steam/apps/1443430/manuals/Rogue_-_Instruction_Manual.pdf?t=1651574702 — original Rogue instruction manual
- https://www.gamedeveloper.com/production/manuals-they-can-be-good — game manual writing best practices
- https://analoggamestudies.org/2014/10/the-rules-for-writing-rules-how-instructional-design-impacts-good-game-design/ — instructional design for rules
