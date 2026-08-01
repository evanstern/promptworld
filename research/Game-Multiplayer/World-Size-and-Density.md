---
title: World Size and Density
aliases: [Map Size Scaling, Content Density]
tags: [world-size, density, map-design, scaling]
type: note
created: 2026-08-01
updated: 2026-08-01
related: ["[[Game-Multiplayer]]", "[[Scale-Ceilings-and-Interest-Management]]", "[[Inter-Settlement-Interaction]]", "[[Multiplayer-Shapes]]", "[[Promptworld-Baseline]]"]
---

# World Size and Density

Does a world have to get bigger to hold more players? The surveyed evidence separates three
questions that are easy to conflate: how much **space** exists, how much **stuff** is in it, and
how much of it the server must **simulate**. Only the third reliably tracks player count.

## Finding 1 — density, not extent, produces the sense of scale

"The true perception of scale is dictated by **playable density** — the amount of meaningful
content packed into every unit of space — which explains why games with smaller maps can feel
vastly larger than those with expansive but empty territories" ([[_grounding]] § World size,
density, and player count).

The counter-case is stated numerically: "to achieve real-world density in a 60,000 square mile
game world, you would need a population the size of Earth, and your game likely won't have that
scale of billions of players — no game currently does." Enlarging a world without a matching
supply of content produces emptiness, not scale.

## Finding 2 — the server cost of a big world is generation and simulation, not area

The area itself is nearly free; what it holds is not.

- **Minecraft**: "chunk generation is the #1 cause of lag spikes. When players explore new
  territory, the server has to generate terrain, caves, structures, and biomes in real-time."
  A larger world is only expensive **as players spread into it** ([[_grounding]] § Minecraft).
- **Valheim**: pressure comes from **ZDOs per area**, not from map extent. A single base with
  "5,000+ building pieces" saturates the default ~64 KB/s cap regardless of how big the world is
  ([[_grounding]] § Valheim).
- **Project Zomboid**: "a multiplayer server has to process zombie AI, player actions, world
  changes, inventory activity, and persistent multiplayer data **across the whole map**" — here
  the whole map *is* the simulated set, and area does cost ([[_grounding]] § Project Zomboid).
- **Eco**: the player ceiling "depends on **world size and server resources**" — world size is
  named as a co-determinant of the ceiling, in the direction of *bigger world = fewer players
  supportable*, not more ([[_grounding]] § Eco).

Minecraft's **simulation distance** dial encodes the distinction directly: view distance governs
what is *sent*, simulation distance governs how far out the world *ticks at all*
([[Scale-Ceilings-and-Interest-Management]]).

## Finding 3 — spreading players out is a load-management technique

Where a bigger world helps, it helps by *separating* players, and it is implemented as
partitioning rather than as raw area:

- **Zoning**: "an efficient way to distribute server load … a zone can typically host ~100
  players" ([[_grounding]] § Interest management).
- **Valheim's zones**: authority is assigned per-zone to the first player who enters, so distance
  between players means independent authority domains ([[_grounding]] § Valheim).
- **Screeps' shards**: separate maps, separate databases, separate runtime servers — the strongest
  form, and effectively "more worlds" rather than "one bigger world" ([[_grounding]] § Screeps).

The literature also warns which mechanics defeat this: "mechanics like large-scale PvP battles or
**player-driven settlements** cannot be handled through traditional zoning" — because settlements
concentrate entities and players in one place by design.

## Finding 4 — the alternative to a bigger map is a second map layer

Dwarf Fortress models many settlements without enlarging the playable site: it keeps a small
tactical embark and puts civilizations, sites, roads, and tunnels on a **separate world map**, from
which caravans and diplomats arrive at the site ([[_grounding]] § Dwarf Fortress inter-settlement,
and [[Inter-Settlement-Interaction]]). RimWorld Together does the same for multiplayer — players
hold separate colonies on a shared **planet**, and interaction (visits, raids, trade) is expressed
at the planet layer ([[Multiplayer-Shapes]]).

This is the structural answer to "do we need more space for more settlements": in this game family,
additional settlements are usually added as **entries on an abstract world map**, not as more tiles
on the tactical one.

## Finding 5 — scaling can be a rules change instead of a space change

"Scaling in game design means a game retains a similar experience regardless of the number of
players by **changing rules, numbers, or other design elements based on player count**"
([[_grounding]] § World size, density, and player count). Resource yields, event frequency, and
threat scaling are the usual levers; none require a larger map.

## Applicability to promptworld

- The current world is a **64×64 tile grid** (`DefaultSize = 64`), **seeded and regenerated rather
  than stored**, and the daemon simulates all of it — there is no simulation-distance concept and
  no chunk streaming ([[Promptworld-Baseline]]). The Minecraft "generation is the lag source"
  finding therefore does not transfer; the PZ "whole map is simulated" finding does.
- Because terrain is regenerated from a seed rather than persisted, **map extent is cheap in
  storage** but is paid for in per-tick simulation and in agent pathfinding/perception work over
  the grid.
- The population that drives cost is **agents, not tiles**: eight villagers with model-gated
  cognition. Under the survey's framing, adding tiles without adding inhabitants moves the world
  toward the "expansive but empty" end.
- promptworld has **no world-map layer** today — one world is one save directory with one map and
  at most one daemon. The DF/RimWorld-Together pattern of "many settlements on an abstract layer"
  would be a new structure, not an enlargement of the existing one.
- promptworld does already have a **multi-world primitive**: the spec-076 fork ceremony (fresh
  prefix log at a snapshot boundary, `world.forked` lineage, seed carried) plus fork/compare duel
  doors, and machine-wide `ps` enumeration of running worlds. That is closer to Screeps-style
  shards than to a bigger single map.

## Grounding

- [[_grounding]] § World size, density, and player count; § Minecraft; § Valheim; § Project
  Zomboid; § Eco; § Interest management; § Screeps; § Dwarf Fortress inter-settlement;
  § promptworld's own architecture
- [TV Tropes — Content Density vs. Width](https://tvtropes.org/pmwiki/pmwiki.php/Main/SlidingScaleOfContentDensityVsWidth)
- [Delphi Digital — Overcoming the limits of scale](https://members.delphidigital.io/reports/overcoming-the-limits-of-scale-in-virtual-worlds)
- [Games Precipice — Player Count & Scalability](https://www.gamesprecipice.com/player-count-scalability/)
