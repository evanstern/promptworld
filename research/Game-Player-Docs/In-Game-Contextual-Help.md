---
title: In-Game Contextual Help
aliases: [Adaptive Tutorials, Learning Helper Pattern]
tags: [contextual-help, tutorial, adaptive]
type: note
created: 2026-07-24
updated: 2026-07-24
related: ["[[Game-Player-Docs]]", "[[Explaining-The-Screen]]", "[[Manual-Structure-Conventions]]"]
---

# In-Game Contextual Help

How games answer questions **at the moment they arise** instead of (or in addition to) a
document read beforehand. RimWorld and Cogmind are the two most articulated systems in the
corpus.

## RimWorld's learning helper: push lessons, triggered by events

RimWorld ships no manual. Its adaptive "learning helper" panel (top-right of screen) works
as follows ([[_grounding]] § RimWorld):

- **Event-triggered:** "if something happens relating to a concept the player hasn't
  learned, that lesson will be activated and shown."
- **Auto-retired:** lessons are marked learned when the player performs the interaction, or
  can be dismissed manually; remaining lessons surface "as needed by circumstance, or on a
  slow timer."
- **Stated invariant:** "you'll never be shown a lesson you already know, but always be
  shown the lessons you need to know now."
- **Doubles as pull reference:** the helper "can be expanded and searched for any lesson,"
  so the same lesson corpus serves both push (triggered) and pull (searched) use.
- A one-time scripted tutorial covers the basics; the adaptive helper takes over after
  ([[_grounding]]).

## Cogmind's four layers, and the ordering principle

The Cogmind developer's design post lays out an explicit hierarchy ([[_grounding]] § Cogmind):

1. **Reduce the need for help first** — "the easiest way to help a player is to make it so
   they don't actually need help in the first place" (intuitive controls, self-explanatory
   labels).
2. **Context help at the point of question** — any stat or UI element can be clicked/keyed
   for an explanation, because "specific questions are best handled at the point where
   they're asked."
3. **Tutorial as a bounded space** — a static four-room level, shown only on the first three
   runs, teaching by sequenced encounters rather than pop-ups.
4. **Tiered references** — `?` opens *basic* and *advanced* command pages separately, plus a
   searchable ~8,000-word in-game manual.

A recorded caution on attention: tutorial messages "hot pink and blinking… and still people
sometimes miss them" — visibility of in-flow help is unreliable, which is part of why the
same content also exists as searchable reference ([[_grounding]]).

## Retrofit cases: DF Steam and Caves of Qud 1.0

- Dwarf Fortress Premium added a basic in-game tutorial + help section; the tutorial removes
  early decision load by auto-choosing embark site, dwarves and supplies. Community
  assessment: onboarding improved, but "you're gonna need to spend some time with your nose
  in a wiki" — in-game help did not displace the external corpus ([[_grounding]] § Dwarf Fortress).
- Caves of Qud 1.0 added its tutorial **together with** a UI overhaul ("modernized menus,
  clarified tooltips, improved readability without diluting depth") — docs and interface
  clarity shipped as one change, and reviews still note the tutorial only *starts* the
  learning ([[_grounding]] § Caves of Qud).

## Grounding

- [[_grounding]] — §§ "RimWorld", "Cogmind", "Dwarf Fortress", "Caves of Qud"
- [RimWorld Wiki: Learning helper](https://rimworldwiki.com/wiki/Learning_helper)
- [Grid Sage Games: Tutorials and Help](https://www.gridsagegames.com/blog/2016/07/tutorials-help/)
