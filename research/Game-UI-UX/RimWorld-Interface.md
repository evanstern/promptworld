---
title: RimWorld Interface
aliases: [RimWorld UI]
tags: [game-ui, rimworld, colony-sim, alerts]
type: note
created: 2026-07-25
updated: 2026-07-25
related: ["[[Game-UI-UX]]", "[[Dwarf-Fortress-Interface]]", "[[Recurring-Interface-Patterns]]"]
---

# RimWorld Interface

How RimWorld surfaces a deep colony simulation legibly — its fixed screen-region layout,
severity-coded event system, self-teaching layer, and pawn-state visual grammar. Widely
documented as the accessible counterpoint to Dwarf Fortress. Facts cited in [[_grounding]]
§RimWorld.

## Screen architecture

- Fixed regions: Architect menu + inspect pane bottom-left (showing per-cell info when
  closed); resources top-left; menu tabs (F1–F8 hotkeys) along the bottom; time controls,
  date, and overlay toggles bottom-right; alerts and the learning helper top-right; colonist
  bar across the top ([RimWorld wiki: User interface](https://rimworldwiki.com/wiki/User_interface);
  [Menus](https://rimworldwiki.com/wiki/Menus)).
- The colonist bar shows every pawn's portrait and status; left-click jumps to the pawn on
  the map. Update 1.5 added optional mood color/glow on the bar below mental-break thresholds
  ([User interface](https://rimworldwiki.com/wiki/User_interface); [1.5 changelog](https://docs.google.com/document/d/e/2PACX-1vSgQmFRFk0FATWTjyZkKRq4oa58sQps4D0kE_uoyKR1y3ZXJT1nIMZSsno7T8cfG-Y6B8lVL3QFnbwQ/pub)).
- Time runs in ticks (60/sec at 1x; 2,500 per in-game hour) with Pause/x1/x3/x6 controls
  (dev-only x15) on keys Space/1/2/3 ([Time](https://rimworldwiki.com/wiki/Time)).
- The work priorities grid encodes three data layers per cell: assignment (check or number
  1–4), skill (border brightness red→white→yellow), and passion (flame icons); a "crunching
  sound" plays when assigning a low-skill pawn ([Work](https://rimworldwiki.com/wiki/Work)).

## Alerts and letters

- Events create "Letters" — envelope icons stacked on the right edge, color-coded by severity:
  blue good, grey neutral, yellow bad, red direct threat; some events instead pause and pop a
  choice dialog ([Events](https://rimworldwiki.com/wiki/Events)).
- Persistent condition alerts (colonist about to break, needs rescue) sit at the top-right
  edge; notifications are click-to-jump ("instantly center the camera on their subject")
  ([Learning helper](https://rimworldwiki.com/wiki/Learning_helper)).
- The History screen keeps the last 200 messages/letters after dismissal, supports pinning,
  and plots colony stats with the same color language (blue/white positive, yellow negative,
  red attacks) ([Menus](https://rimworldwiki.com/wiki/Menus)). 1.4 added duplicate-message
  highlighting and letter bundling; 1.5 added right-click dismissal ([changelogs](https://docs.google.com/document/d/e/2PACX-1vSgQmFRFk0FATWTjyZkKRq4oa58sQps4D0kE_uoyKR1y3ZXJT1nIMZSsno7T8cfG-Y6B8lVL3QFnbwQ/pub)).

## The self-teaching layer

- Alpha 15 (2016) introduced the tutorial (which "locks out irrelevant controls and highlights
  relevant controls") and the learning helper. Tynan Sylvester's design statement: "you'll
  never be shown a lesson you already know, but always be shown the lessons you need to know
  now" — lessons trigger on circumstance, auto-mark as learned by doing, and are searchable
  ([Ludeon blog: Alpha 15](https://ludeon.com/blog/2016/08/alpha-15-tutorial-and-drugs-released/)).
- Sylvester's stated philosophy (Designing Games; interviews): elegance = maximum experience
  per unit comprehension burden; "noise" is signal that fails to transmit meaning; RimWorld
  aims to push Dwarf Fortress's story potential "in a way that's approachable"; a simulation
  "not overly complex, that a player can observe and comprehend"
  ([Game Developer](https://www.gamedeveloper.com/design/how-i-rimworld-i-fleshes-out-the-i-dwarf-fortress-i-formula);
  [GDC 2017 talk](https://www.gdcvault.com/play/1024232/-RimWorld-Contrarian-Ridiculous-and)).

## Pawn-state visual grammar

- Mood is a bar with a target-triangle (instantaneous sum of thoughts) that the bar chases
  over time, plus threshold lines for minor/major/extreme mental breaks; need meters carry
  tendency triangles and hatch-mark thresholds where new thoughts fire
  ([Mood](https://rimworldwiki.com/wiki/Mood); [Needs](https://rimworldwiki.com/wiki/Needs)).
- Every mood influence is an itemized "thought" with an explicit signed number ("Low
  expectations +12") ([Thoughts](https://rimworldwiki.com/wiki/Thoughts)).
- Health is tracked per body part against named capacities (Sight, Moving, Consciousness…)
  in a tabbed pane with an operations scheduler ([Health](https://rimworldwiki.com/wiki/Health)).

## Reception and the mod layer

- Positioned by reviewers as "the game for everyone who wanted to get into Dwarf Fortress but
  couldn't because ASCII" (RPS 2016), while PC Gamer still called the vanilla UI "far from
  intuitive and full of irksome inconsistencies" ([RPS mirror](https://bliss.isa-geek.net/2016/07/18/rimworld-review-early-access/);
  [PC Gamer](https://www.pcgamer.com/rimworld-review/)).
- QoL mods document the residual gaps: Dubs Mint Menus (search over bills/architect/research),
  RimHUD (denser pawn HUD with warnings), Interaction Bubbles (surfacing log-only social text
  as speech bubbles over pawns), Colony Manager (automating repetitive orders)
  ([mod feature lists in grounding](https://github.com/Jaxe-Dev/RimHUD)).
- 1.5 (2024) added map search (hotkey Z) over items/buildings/creatures — vanilla search
  arrived a decade into development ([1.5 changelog](https://docs.google.com/document/d/e/2PACX-1vSgQmFRFk0FATWTjyZkKRq4oa58sQps4D0kE_uoyKR1y3ZXJT1nIMZSsno7T8cfG-Y6B8lVL3QFnbwQ/pub)).

## Grounding

- [[_grounding]] — §"RimWorld UI/UX — Cited Fact Digest" (all claims above).
