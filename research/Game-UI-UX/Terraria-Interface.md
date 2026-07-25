---
title: Terraria Interface
aliases: [Terraria UI]
tags: [game-ui, terraria, inventory, hud]
type: note
created: 2026-07-25
updated: 2026-07-25
related: ["[[Game-UI-UX]]", "[[Recurring-Interface-Patterns]]"]
---

# Terraria Interface

How Terraria's inventory/crafting/HUD design evolved over a decade of quality-of-life waves,
how it handles crafting discoverability (the Guide NPC, Bestiary, Journey Mode), and how its
mouse-grid UI was re-mapped for controllers and touch. Not a TUI — mined for transferable
patterns. Facts cited in [[_grounding]] §Terraria.

## Inventory and crafting

- 40 main slots + a 10-slot hotbar on keys 0–9 (lockable against accidental swaps), plus
  dedicated slot classes for coins, ammo, armor, accessories, trash
  ([Terraria wiki: Inventory](https://terraria.wiki.gg/wiki/Inventory)).
- The crafting menu shows **only currently craftable recipes** (ingredients in hand or in an
  open chest), not a full recipe browser ([Inventory](https://terraria.wiki.gg/wiki/Inventory)).
- QoL arrived in documented waves: 1.2 (2013) extended crafting menu + minimap; 1.3 (2015)
  Quick Stack to All Nearby Chests + item favoriting; 1.3.1 (2016) inventory sorting + smart
  interact + native controller support; 1.4 (2020) craft-from-chest, gold recipe highlight,
  Void Bag overflow storage; 1.4.4 (2022) 9999 stacks, equipment loadouts (F1–F3), and a
  visual effect showing quick-stacked items flying to their chests; 1.4.5 crafting category
  tabs + search bar ([official changelogs](https://terraria.wiki.gg/wiki/1.2), [1.3.0.1](https://terraria.wiki.gg/wiki/1.3.0.1), [1.3.1](https://terraria.wiki.gg/wiki/1.3.1), [1.4.0.1](https://terraria.wiki.gg/wiki/1.4.0.1), [1.4.4](https://terraria.wiki.gg/wiki/1.4.4)).

## HUD

- Health/mana offer six selectable display styles (icons vs bars, with/without numerics)
  ([Health](https://terraria.wiki.gg/wiki/Health)); buffs are icons below the hotbar with
  remaining duration, right-click to cancel ([Buffs](https://terraria.wiki.gg/wiki/Buffs)).
- The minimap (1.2) reveals tiles "the brightest they have ever been seen"; boss health bars
  only arrived in 1.4, auto-tracking the most recently damaged or closest boss
  ([1.2](https://terraria.wiki.gg/wiki/1.2); [Boss health bar](https://terraria.wiki.gg/wiki/Boss_health_bar)).
- Event/invasion progress bars, breath meters above the player, and map icons with white
  outlines for visibility were incremental additions (1.3, 1.4)
  ([1.3.0.1](https://terraria.wiki.gg/wiki/1.3.0.1); [1.4.0.1](https://terraria.wiki.gg/wiki/1.4.0.1)).

## Discoverability

- The Guide NPC answers "what can I craft with this?" for any item shown to him (added 1.0.5),
  and his help dialogue is dynamic — adjusting to bosses defeated, NPCs present, and inventory
  ([Guide](https://terraria.wiki.gg/wiki/Guide)).
- Early-game opacity is a documented criticism thread: no tutorial and a slow start
  (PC Gamer 2011), "without a handy guide, many would be lost" (Destructoid 2013), progression
  "intentionally opaque" (PC Gamer) ([reviews in grounding](https://www.pcgamer.com/terraria-review/)).
- 1.4's documented responses: the Bestiary — a 546-entry in-game catalog whose entries unlock
  in four stages by kill count, with 60 filters and search — and Journey Mode's
  research-and-duplicate Power Menu ([Bestiary](https://terraria.wiki.gg/wiki/Bestiary);
  [Journey Mode](https://terraria.wiki.gg/wiki/Journey_Mode)).

## Input remapping across platforms

- Console (per David Welch, who led the adaptation): two cursor modes shipped because neither
  sufficed alone — Auto Cursor (stick-direction targeting) and Manual Cursor (mouse-like);
  combat became directional aiming; D-pad bindings and a dedicated grapple button pulled
  frequent actions off the hotbar ([Welch, Medium](https://medium.com/@watsonwelch/terraria-on-console-10-years-later-part-2-the-controls-bb84cac14700)).
- Mobile split the PC's single multi-panel inventory screen into tabbed screens, added a
  magnifying-glass cursor ("it's hard to see under your fingers"), and used a persistent
  categorized recipe list that unlocks as ingredients are found
  ([Welch, Medium part 3](https://medium.com/@watsonwelch/the-making-of-terraria-mobile-part-3-crafting-a-new-ui-4fb84708c767)).
- Gamepad uses a radial hotbar with an option to cycle via bumpers instead
  ([Game controls](https://terraria.wiki.gg/wiki/Game_controls)).

## Grounding

- [[_grounding]] — §"Terraria UI/UX — Cited Fact Digest" (all claims above).
