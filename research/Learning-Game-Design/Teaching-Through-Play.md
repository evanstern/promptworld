---
title: Teaching Through Play
aliases: [Silent Tutorials, Onboarding Curve Design]
tags: [tutorial, onboarding, flow, level-design]
type: note
created: 2026-07-25
updated: 2026-07-25
related: ["[[Learning-Game-Design]]", "[[Puzzle-Pedagogy-Patterns]]", "[[Learning-Helper-Anatomy]]"]
---

# Teaching Through Play

What the tutorial/onboarding design literature says about teaching mechanics inside play
rather than beside it ([[_grounding]] § Area 2). Three bodies of evidence: Portal's
level-design method, George Fan's ten rules, and flow/difficulty-curve theory.

## Portal: the sequenced silent tutorial

- Each test chamber isolates "some key aspect of a mechanic" and asks "how to teach it to
  players while still challenging them" — Valve's own description of the design method
  ([Valve Developer Community](https://developer.valvesoftware.com/wiki/Portal_Design_And_Detail)).
- Levels build cumulatively: "Every level in these games builds slowly upon the previous
  one, introducing new concepts to the player while also requiring them to call upon all of
  the previous skills that they have learned" ([[_grounding]] § Portal). The opening is a
  safe, confined room where controls can be learned with no threat.
- The method was validated empirically, not authored: weekly playtests (test Friday, discuss
  Monday, fix all week) from one week into development; Gabe Newell called playtesting "our
  secret weapon." Unintended solutions that "made the tester feel smart" were kept
  ([[_grounding]]; [GMTK: Valve's "Secret Weapon"](https://gmtk.substack.com/p/valves-secret-weapon)).

## George Fan's ten tutorial rules (GDC 2012)

From "How I Got My Mom to Play Through Plants vs. Zombies" ([[_grounding]] § George Fan;
[GDC Vault](https://www.gdcvault.com/play/1015541/How-I-Got-My-Mom)): blend the tutorial
into the game; prioritize doing over reading; spread out mechanic introductions; one
performed action teaches ("Once they see the results of their action, that's often all it
takes"); maximum "eight words on the screen at any given moment"; unobtrusive messaging;
adaptive messaging (tips only for players doing the wrong thing); avoid noise ("like being
the little boy who cried wolf, and the player will tune out"); visuals that encode
function; leverage existing knowledge.

## The general "no more tutorials" principles

Design writing converges on: teach on a need-to-know basis, one mechanic at a time;
reinforce or players forget; show-don't-tell by forcing discovery (Super Metroid's
creatures demonstrate the wall-jump before the player needs it); consistent visual
affordances; real-world knowledge as free tutorialization
([Game Developer: No More Tutorials!](https://www.gamedeveloper.com/audio/no-more-tutorials-how-to-convey-information-through-design),
[[_grounding]]).

## Flow theory as the difficulty-curve substrate

Csikszentmihalyi's model — anxiety when challenge exceeds skill, boredom when skill exceeds
challenge, flow in the "fuzzy safe zone" between — is the standard theoretical frame.
Jenova Chen's "Flow in Games" thesis applies it to games and argues for widening the flow
zone per player, since "a game designed for the 'average' player might be boring to a
'hardcore' player and frustratingly difficult for a 'novice' player"; Sweetser & Wyeth's
GameFlow model (2005) turns flow into evaluable criteria (concentration, challenge, skills,
control, clear goals, feedback, immersion, social interaction)
([[_grounding]] § Flow; [Chen thesis PDF](https://www.jenovachen.com/flowingames/Flow_in_games_final.pdf)).

## Grounding

- [[_grounding]] — § Area 2 (Tutorial and onboarding-curve literature)
- [Valve Developer Community: Portal — Designing Test Chambers](https://developer.valvesoftware.com/wiki/Portal_Design_And_Detail)
- [Game Developer: George Fan's 10 tutorial tips](https://www.gamedeveloper.com/design/gdc-2012-10-tutorial-tips-from-i-plants-vs-zombies-i-creator-george-fan)
- [Jenova Chen: Flow in Games](https://www.jenovachen.com/flowingames/Flow_in_games_final.pdf)
