---
title: Deity Multiplayer
aliases: [Competing Gods, Guardian-to-Guardian, Indirect Control Multiplayer]
tags: [god-games, indirect-control, faith, guardians]
type: note
created: 2026-08-01
updated: 2026-08-01
related: ["[[Game-Multiplayer]]", "[[Multiplayer-Shapes]]", "[[Inter-Settlement-Interaction]]", "[[Promptworld-Baseline]]"]
---

# Deity Multiplayer

The brief asks how *Guardians* interact — plural deities in one fiction. This is a narrow,
well-established genre question. The god-game lineage is the only body of shipped work where the
player's unit of agency is a divine intermediary rather than a unit or an avatar, and it has a
consistent answer about what multiplayer means in that frame.

## The control style, as the genre defines it

"A god game is a strategy game where the player acts as a divine or otherworldly entity
influencing a population from above, reshaping the world through **indirect means** like raising
terrain, summoning weather, and commanding worshipers **rather than controlling individual
units**" ([[_grounding]] § God games).

The genre's recurring mechanical furniture, as listed in the sources:
- terraforming / world-level intervention
- indirect unit control
- **faith or belief as a resource**
- **miracle or power systems**

Lineage: "Peter Molyneux's Populous in 1989 invented the template; Black & White, Spore, From
Dust, and Reus carried it forward."

## The multiplayer form the genre actually uses

Where god games have multiplayer, it is **competing deities over separate populations**: "players
may sometimes compete against other players with their own population of supporters. In Populous
specifically, the user is a deity that must lead their followers **against the followers of a
rival deity** in order to conquer and capture them" ([[_grounding]] § God games).

Three structural features of this form:

1. **Each deity has their own population.** The unit of ownership is a people, not a territory or
   an army. This maps onto Shape B / Shape C in [[Multiplayer-Shapes]] — parallel settlements
   under separate patrons.
2. **Deities do not act on each other directly.** The genre's grammar is deity → own followers →
   world → other followers. Contact between gods is mediated by their peoples.
3. **The faith economy is the scoreboard and the fuel at once.** Because intervention capacity
   derives from the population's belief/prosperity, a deity's power is an *output* of how well
   their people are doing, which makes neglect self-punishing and makes conquest of followers a
   direct transfer of power. Godus was pitched on exactly this blend: "the power, growth and scope
   of Populous with the detailed construction and **multiplayer excitement** of Dungeon Keeper."

## What the genre does not supply

The surveyed material documents the competitive form and is largely silent on:
- **Cooperative pantheons** — two deities patronising the *same* people, with divided or
  overlapping domains.
- **Conflict resolution between simultaneous divine acts** on the same object.
- **Asynchronous divine play** — a deity acting on a world while the rival is away.

These are gaps in the published record rather than settled answers, and they are recorded as open
questions on [[Game-Multiplayer]].

## The nearest non-god-game analogues for multiple influencers

Where the survey does have evidence about multiple actors influencing one shared simulation:

- **Eco** replaces divine authority with **political authority**: a collectively ratified
  constitution, proposed and voted laws that "restrict what other players can do," taxes and
  grants as incentive levers ([[_grounding]] § Eco). This is a shipped model for *many players
  steering one population indirectly*, resolved through an explicit decision procedure rather
  than through raw power.
- **Haven & Hearth** replaces it with **claims and upkeep**: villages own an influence radius
  that "prevents non-members from interacting with village-owned objects," sustained by a drained
  **authority pool** ([[Inter-Settlement-Interaction]]). This is a shipped model for *bounded
  spheres of influence* between rival groups.
- **Smallville / Generative Agents** shows the LLM-agent version of the influencer role, where the
  human's part is framed as "user control" and intervention on an autonomous population — the
  documented case being a single suggestion cascading into an agent-organised party
  ([[_grounding]] § LLM-agent simulations).

## Applicability to promptworld

promptworld's Guardian already implements the genre's full mechanical vocabulary, single-player
([[Promptworld-Baseline]]):

- **Indirect influence only** — villagers are sealed; the Guardian acts through omens, visions,
  designations, directives, prophecy, and standing orders.
- **A charge economy** priced per miracle, with a gratis doctrine and premium-priced acts.
- **An endogenous faith loop** — village faith is event-sourced over a five-reason delta table,
  and "charge regen is a pure faith-band function." This is precisely the genre's "power derives
  from your people's belief" property, already built.
- **An editable charter** defining the Guardian's own persona and competence ceiling.
- **Canonized regions and missions** as durable world-plan artifacts — a Guardian can name places
  and hold standing intentions that outlive a session.

The consequences for a multi-Guardian design, from the genre evidence:
- The natural published form is **one Guardian per village**, which coincides with the
  parallel-settlements shape and with TASK-65's "parallel villages" option.
- A **shared-village, multi-Guardian** design has no strong shipped precedent in the god-game
  lineage; the closest evidence is Eco's political layer, where multiple actors share one
  population and resolve conflict through a ratified procedure — and promptworld already has a
  governance subsystem (norms, votes, daily meetings under an event-sourced convention) that
  occupies that conceptual slot for *villagers*.
- Faith-as-fuel means that in any multi-Guardian design, **the relative standing of Guardians is
  already derivable from the simulation** rather than needing a separate scoring system.

## Grounding

- [[_grounding]] § God games; § Eco; § Haven & Hearth; § LLM-agent simulations; § promptworld's
  own architecture
- [Grokipedia — God game](https://grokipedia.com/page/God_game)
- [Galaxus — Populous, Black & White and beyond](https://www.galaxus.at/en/page/populous-black-white-and-beyond-the-evolution-of-the-god-simulation-39003)
- [TechCrunch — Project Godus](https://techcrunch.com/2012/11/23/god-complex-peter-molyneux-kicks-off-first-kickstarter-with-project-godus-a-grand-plan-to-recreate-the-entire-god-game-genre)
