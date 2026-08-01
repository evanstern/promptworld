---
title: Inter-Settlement Interaction
aliases: [Other Villages, Trade Raid Visit, Territory Claims]
tags: [villages, trade, diplomacy, territory, factions]
type: note
created: 2026-08-01
updated: 2026-08-01
related: ["[[Game-Multiplayer]]", "[[Multiplayer-Shapes]]", "[[Deity-Multiplayer]]", "[[World-Size-and-Density]]", "[[Promptworld-Baseline]]"]
---

# Inter-Settlement Interaction

The brief asks how a Guardian works with "other villages." That is a design question with a
substantial answer in this game family — and notably, the richest published model comes from a
game with **no multiplayer at all**. Dwarf Fortress shows that "other settlements" is a
simulation feature first and a multiplayer feature only incidentally.

## The Dwarf Fortress model — other settlements as simulated civilizations

DF has no human multiplayer, but models a full world of other sites the fortress interacts with
([[_grounding]] § Dwarf Fortress inter-settlement).

**Civilizations as entities.** A civilization is defined by "recognizable names and symbols, art
forms, and languages, communities possessing common ethics and values, the development of
advanced sites and structures … governing systems … interactions with the environment, and the
creation of complex relationships and **diplomacy** between other entities, which can result in
global networks of **trade**, the **sharing of knowledge**, and conflicts due to disagreements,
which may break out into **war**." Worldgen "controls the number of civilizations in each world."

**Sites are placed on a world map, keyed to terrain.** Mountain civs at mountain edges (`Ω`),
plains civs as lowland wooden towns (`+ * ☼ #`), forest civs in heavy forest (`î ¶`), connected
during worldgen "by either roads or tunnels."

**The interaction verbs:**
- **Caravans on a seasonal cadence** — "the three main civilizations (humans, dwarves, elves) will
  send caravans to you in Spring/Summer/Autumn, and how you trade/gift with them can adjust their
  attitude towards you and your dwarven civilization as well as future trading."
- **Diplomacy via a visiting liaison** — "when a caravan visits your base, an Outpost Liaison will
  trigger a Diplomacy button … by which you can ask for the caravan to bring specific goods next
  year."
- **Persistent diplomatic state** — "Peace is the normal state … civilized entities will trade
  with you and send diplomats. **War** causes civilizations to engage in full-on warfare … entities
  you are at war with will potentially send **sieges** to your fort, especially if you attack them
  first."

Three properties worth isolating: interaction is **episodic** (caravans arrive on a schedule
rather than continuously); it is **stateful** (attitude persists and accumulates across visits);
and it is **asymmetric** (the other settlement is an abstraction that sends representatives, not a
place you must simulate at full fidelity).

## The RimWorld Together model — the same verbs, but the other village is a player

RimWorld Together takes DF's structure and puts a human behind each site: players share "the same
planet" while "keeping separate colonies and pace," and "every action is synced, including
**visiting, raiding, trading, creating factions, roads, sites**." The advertised feature set is
"visit, raid, and spy other players," "trade with other players in real time," and "create and
manage custom factions" ([[_grounding]] § RimWorld).

The verb set is nearly identical to DF's — trade, visit, raid, faction relations — which suggests
these verbs are the family's natural inter-settlement grammar regardless of who is on the other
end.

## The Haven & Hearth model — territory as the primitive

Where DF and RimWorld abstract other settlements as sites on a map, Haven & Hearth makes
settlements **spatial claims in one continuous world** ([[_grounding]] § Haven & Hearth):

- **A claim with a centre and a radius**: "villages have a village claim or idol which sends out
  influence in a large area, and this area can be expanded with statues and banners."
- **Concrete geometry**: "banners expand village claims at a range of 30 tiles in a square shape,
  making a 61×61 claim around it."
- **Membership defines rights**: "villages work as a shared claim which prevents non-members from
  interacting with village-owned objects, and special village officials get certain benefits."
- **Upkeep as a governor on size**: villages "maintain a pool of **authority** … the idol,
  banners, and statues **drain authority** and lose effectiveness if the authority pool is fully
  drained." Territory therefore costs something ongoing.
- **A higher tier**: players "found **Realms** to lay claim to entire Provinces by capturing
  ancient **Thingwalls**."
- **Politics is emergent and out-of-band**: the game hosts "a forum dedicated to discussing
  in-game politics, village relations and matters of justice."

## The Eco model — one settlement, many players, politics as the interaction

Eco does not have inter-village interaction so much as **intra-world politics**: a ratified
constitution governing how laws are proposed and approved, player currencies, and laws that
"restrict what other players can do" or steer behaviour via "taxes or government grants," all
under a shared deadline and a shared environmental externality ([[_grounding]] § Eco). It is the
reference for *negotiated* rather than *transactional* interaction.

## The verb grammar, consolidated

Across the survey, inter-settlement interaction resolves into a small, recurring set:

| Verb | DF | RW Together | H&H | Eco |
|---|---|---|---|---|
| Trade goods | caravans, seasonal | real-time trade | player markets | currencies, markets |
| Visit | liaison/diplomats | visit | free movement | free movement |
| Raid / war | sieges, war state | raid, spy | claim conflict | — |
| Diplomacy / standing | persistent attitude | faction management | village relations | law and votes |
| Territory | site on world map | site on planet | claim radius + upkeep | shared world |
| Knowledge exchange | "sharing of knowledge" | — | — | tech/skill specialisation |

## Applicability to promptworld

- promptworld today has **one village per world**, and no world-map layer: "each world run is one
  save directory and at most one daemon process; multiple worlds mean multiple daemons"
  ([[Promptworld-Baseline]]). "Other villages" is not a modelled concept.
- The game does, however, already have **one shipped inter-group primitive**: the spec-077
  **stranger** entity family — an outside actor arriving as a scenario incident. That is
  structurally the DF caravan/liaison pattern in miniature: an outsider arrives, interacts, and
  leaves, without the outside settlement being simulated.
- It also has the **social and governance machinery** the verbs would attach to — relationships,
  rumors, debts, secrets, conversations, norms, votes, a daily meeting under an event-sourced
  convention, and a village charter — which is the substrate DF's "attitude persists across
  visits" property would need.
- The Guardian's **canonization** miracle already names durable regions as world-plan artifacts,
  and **faith** is already an event-sourced village-level quantity — the two pieces a claim/
  influence model (Haven & Hearth style) or a patron-standing model (god-game style) would rest on
  ([[Deity-Multiplayer]]).
- Per [[World-Size-and-Density]], the family's standard way to add other settlements is a
  **separate abstract world-map layer**, not more tiles on the tactical map.

## Grounding

- [[_grounding]] § Dwarf Fortress inter-settlement; § RimWorld; § Haven & Hearth; § Eco;
  § promptworld's own architecture
- [DF Wiki — Civilization](https://dwarffortresswiki.org/index.php/Civilization)
- [DF Wiki — Diplomacy](https://dwarffortresswiki.org/index.php/Diplomacy)
- [Haven and Hearth Wiki — Village](https://havenandhearth.fandom.com/wiki/Village)
- [Steam Workshop — RimWorld Together](https://steamcommunity.com/sharedfiles/filedetails/?id=3005289691)
