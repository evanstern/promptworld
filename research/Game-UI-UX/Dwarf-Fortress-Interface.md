---
title: Dwarf Fortress Interface
aliases: [DF UI, Dwarf Fortress UI]
tags: [game-ui, dwarf-fortress, ascii, redesign]
type: note
created: 2026-07-25
updated: 2026-07-25
related: ["[[Game-UI-UX]]", "[[TUI-and-Roguelike-UI-Craft]]", "[[RimWorld-Interface]]", "[[Recurring-Interface-Patterns]]"]
---

# Dwarf Fortress Interface

How Dwarf Fortress presents an enormous simulation through a text grid — the classic
keyboard/ASCII interface, the 2022 Steam "Premium" mouse-first redesign, and the community
tooling layer that grew around the native UI's gaps. Facts cited in [[_grounding]] §Dwarf Fortress.

## The classic keyboard/ASCII interface

- The world renders entirely in text glyphs in sixteen colors (`☺` dwarf, `c` cat, `d` dog);
  a 3D volume is viewed one z-level slice at a time (`<`/`>` to move vertically), with numpad
  or arrow-key cursor movement ([Wikipedia](https://en.wikipedia.org/wiki/Dwarf_Fortress);
  [DF wiki: Controls](https://dwarffortresswiki.org/index.php/DF2014:Controls)).
- Single-keypress top-level menus drive everything: `d` designations, `b` building, `k` look,
  `u` units, `v` view creature. Key meanings change per screen; each screen lists its own
  controls at the bottom ([DF wiki: Controls](https://dwarffortresswiki.org/index.php/DF2014:Controls)).
- Menu scrolling is famously inconsistent: because arrow keys stay bound to map scrolling,
  lists scroll with `+`/`-` and page with `*`/`/`; veterans describe "at least three different
  ways of scrolling" across menus ([DF wiki: Controls](https://dwarffortresswiki.org/index.php/DF2014:Controls);
  [Steam discussions](https://steamcommunity.com/app/975370/discussions/0/3416557114751453595)).
- The learning curve is documented as the game's defining UX fact: no in-game tutorial for
  ~16 years; reviewers described basic actions as "like teaching a beetle to cook" (RPS) and
  "banging two rocks together" (Ars Technica) ([Wikipedia](https://en.wikipedia.org/wiki/Dwarf_Fortress)).
- Tarn Adams' stated reason the UI stayed that way: "there were always new features to work
  on, and the game was perpetually unfinished anyway; why work on usability when those menus
  would surely be changing?" ([PC Gamer 2019](https://www.pcgamer.com/tutorials-and-mouse-support-could-make-dwarf-fortress-on-steam-vastly-easier-to-play/)).

## Information-density patterns

- Announcements appear top-left; major ones pause the game and center the camera on the event.
  A dated announcements list supports zoom-to-event for some entries
  ([DF wiki: Announcement](https://dwarffortresswiki.org/index.php/DF2014:Announcement)).
- The `k` look cursor lists everything on a tile in fixed hierarchy (creatures → objects →
  buildings → tile); list screens support "zoom to creature," jumping from list-state to
  map-state ([DF wiki: Controls guide](https://dwarffortresswiki.org/index.php/DF2014:Controls_guide)).
- Deep per-dwarf state (emotions, favorite gems, grudges, relationships) is surfaced through
  text screens ([Stack Overflow blog](https://stackoverflow.blog/2021/12/31/700000-lines-of-code-20-years-and-one-developer-how-dwarf-fortress-is-built/)).

## The Steam Premium redesign (Dec 2022)

- Shipped: pixel-art tileset (by hired community modders Mayday and Meph), full mouse support,
  navigable menus with scrollbars/tabs, an interactive tutorial, widescreen scaling
  ([Kitfox devlog](https://kitfoxgames.itch.io/dwarf-fortress/devlog/460421/dwarf-fortress-is-out-now);
  [PCGamesN](https://www.pcgamesn.com/dwarf-fortress/menus)).
- The UI was designed mostly by Tarn and Zach Adams themselves; a professional UX person was
  brought in "later than we should have" and produced "our nice building menu"
  ([Game Developer interview](https://www.gamedeveloper.com/programming/how-tarn-adams-upgraded-and-optimized-dwarf-fortress-for-its-official-steam-release)).
- Classic and Premium are one codebase: "the switch… is just swapping out some glyphs. The
  grid structure underneath is the same everywhere" — the UI model is grid-native regardless
  of renderer ([Game Developer interview](https://www.gamedeveloper.com/programming/how-tarn-adams-upgraded-and-optimized-dwarf-fortress-for-its-official-steam-release)).
- Adoption evidence: ~300k copies in week one, 1M+ by April 2025, Metacritic 93, vs roughly
  $15k/month donations pre-Steam ([Wikipedia](https://en.wikipedia.org/wiki/Dwarf_Fortress)).
- Documented trade-offs of the redesign: PC Gamer found the classic keyboard interface
  "followed a reliable logic with clean separation between playspace and menu information,
  while the new interface feels scattered," with overwhelming visual noise in busy forts and
  keyboard shortcuts lost to mouse-first design ([PC Gamer review](https://www.pcgamer.com/dwarf-fortress-review/));
  DSOGaming documented scroll-position resets, no batch operations, and a "barebones" tutorial
  ([DSOGaming](https://www.dsogaming.com/special/reviews/dwarf-fortress-premium-pc-review/)).

## The community UX layer (evidence of native gaps)

- **Dwarf Therapist**: an external GUI reading game memory to manage labor assignments,
  stats, and moods in a sortable grid — wiki pull-quote: "It makes Dwarf Fortress playable"
  ([DF wiki: Dwarf Therapist](https://dwarffortresswiki.org/index.php/Utility:Dwarf_therapist)).
- **DFHack**: memory-access library shipping interface enhancements and automation for
  "aspects of gameplay many players find toilsome"; its `manipulator` embeds a Therapist-style
  labor grid in-game ([DFHack docs](https://docs.dfhack.org/en/stable/docs/Introduction.html)).
- The tileset modders were eventually hired to make the official tileset; Adams publicly
  endorsed third-party tools ([Variety](https://variety.com/2019/gaming/news/dwarf-fortress-steam-itchio-1203162351/);
  [Wikipedia](https://en.wikipedia.org/wiki/Dwarf_Fortress)).

## Grounding

- [[_grounding]] — §"Dwarf Fortress UI/UX — Cited Fact Digest" (all claims above).
