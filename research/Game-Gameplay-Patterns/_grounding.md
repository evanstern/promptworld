---
title: Game Gameplay Patterns — Grounding
aliases: []
tags: [grounding]
type: source
created: 2026-07-24
updated: 2026-07-24
related: ["[[Game-Gameplay-Patterns]]"]
---

# Game Gameplay Patterns — Grounding

> Source-of-truth artifact. This is the raw, cited output of a research pass (the `deep-research`
> skill, or a direct web-search fan-out). Keep it close to verbatim — do not editorialize, prune,
> or draw conclusions here. Knowledge notes and analyses cite *into* this file.

**Research question:** What gameplay patterns do colony sims, roguelikes, god games, and
LLM-agent sims use for drama pacing, failure, indirect control, observation-driven play, and
difficulty framing — RimWorld, Dwarf Fortress, Cogmind, Populous/Black & White/WorldBox,
Progress Quest, Generative Agents (Smallville)?
**Method:** web-search fan-out (8 parallel searches + 3 direct page fetches) · 2026-07-24

---

## RimWorld — the AI storyteller (director-driven drama)

- RimWorld is explicitly marketed as a **"story generator"** rather than a traditional
  strategy game; its AI storyteller is "modeled after the AI Director from Left 4 Dead" and
  "analyzes your situation and decides which event she thinks will make the best story"
  ([About RimWorld](https://rimworldwiki.com/wiki/About_RimWorld),
  [AI Storytellers — RimWorld Wiki](https://rimworldwiki.com/wiki/AI_Storytellers)).
- Architecture: "story watchers" monitor game state and "incident generators" trigger events
  in response; "the storyteller builds pressure, then releases it, creating emotional arcs
  similar to a narrative campaign"
  ([Game Developer: procedurally generated storytelling](https://www.gamedeveloper.com/design/rimworld-dwarf-fortress-and-procedurally-generated-story-telling)).
- **Three storyteller personalities double as pacing/difficulty dials**
  ([AI Storytellers — RimWorld Wiki](https://rimworldwiki.com/wiki/AI_Storytellers),
  [GamepadSquire storyteller guide](https://gamepadsquire.com/blog/rimworld/rimworld-storytellers-difficulty-guide/)):
  - **Cassandra Classic** — "story events on a classic increasing curve of challenge and
    tension"; cooldown-based, aiming for 1–2 major threats per quadrum (~every 7–10 in-game
    days) with at least two days between major events; difficulty scales with time and
    colony wealth.
  - **Phoebe Chillax** — "long break between challenges… enough time to rest and recover";
    slower pacing "gives new players a more relaxed experience"; still "hits as hard as
    anyone" at high difficulty.
  - **Randy Random** — "generates random events, and he doesn't care if they make a story of
    triumph or utter hopelessness. It's all drama to him."
- The storyteller "is the main mechanism used to determine game difficulty and play style"
  — narrative persona and difficulty selection are the same control surface
  ([GamepadSquire](https://gamepadsquire.com/blog/rimworld/rimworld-storytellers-difficulty-guide/)).

## Dwarf Fortress — simulation-driven emergent narrative, "losing is fun"

- DF has "no predefined goal, endpoint, or win condition — it is up to the player to decide
  how to play and what goals they set themselves"; the developers "intend for players to
  learn through failure," per the motto "losing is fun!"
  ([Cartlidge 2024, *Interpreting Dwarf Fortress*](https://journals.sagepub.com/doi/full/10.1177/15554120231162418)).
- "Because there is no win condition, every playthrough will eventually end in disaster…
  part of the fun is also discovering how your lovingly crafted civilization meets its doom"
  ([Cartlidge 2024](https://journals.sagepub.com/doi/full/10.1177/15554120231162418)).
- DF "achieves the ultimate goal of any complex simulation: emergent narrative driven by
  uncompromising logic"; emergent narrative = "stories imagined and possibly retold by
  players recounting their experiences in a game, often adding details not present in the
  game itself" ([Medium: DF engineering marvel](https://medium.com/@christian.marques/dwarf-fortress-an-engineering-marvel-of-the-21st-century-2ba3a1e9b95f),
  [ResearchGate: Characterization and Emergent Narrative in Dwarf Fortress](https://www.researchgate.net/publication/356686095_Characterization_and_Emergent_Narrative_in_Dwarf_Fortress)).
- **Boatmurdered** — a community succession-game chronicle — "has been praised as an example
  of Dwarf Fortress' potential for emergent storytelling, and credited for introducing both
  Dwarf Fortress and the Let's Play format to a broader audience"
  ([Wikipedia: Boatmurdered](https://en.wikipedia.org/wiki/Boatmurdered)). Notable: the
  celebrated story artifact is a *player-written retelling*, not game output.

## The DF vs RimWorld contrast (simulation-driven vs director-driven)

Source: [Game Developer: "How Dwarf Fortress and Rimworld tell radically different stories"](https://www.gamedeveloper.com/design/dwarf-fortress-and-rimworld-tell-very-different-stories)
(fetched 2026-07-24); [Game Developer: procedurally generated storytelling](https://www.gamedeveloper.com/design/rimworld-dwarf-fortress-and-procedurally-generated-story-telling).

- "RimWorld reacts to player progress, shaping events to keep the story engaging, while
  Dwarf Fortress reacts to its own systems, not to player pacing."
- RimWorld "prioritizes streamlined gameplay and player control… manageable scope… with
  intentional pacing"; DF "prioritizes unfettered simulation without concern for user
  experience," generating "unpredictable emergent narratives through deep, interconnected
  systems — prioritizing complexity over accessibility."
- Player attachment differs by design: in RimWorld the author "deliberately avoids
  attachment to settlers because 'they're not gonna be around for too long'"; in DF players
  "become deeply invested in individual dwarves," giving them "nicknames after they survive
  particularly harrowing events" and "engraving slabs to memorialize them."
- Memorable line: "Dwarf Fortress does not strive to be a good video game. That is simply
  not one of the game's priorities. It's a good video game seemingly by coincidence."

## Roguelike permadeath — failure as content

- Rogue "redefined punishment in games by introducing the idea of permadeath as a core
  design feature… a philosophy about embracing failure as a part of the journey"
  ([LitRPG Reads](https://litrpgreads.com/blog/permadeath-the-heart-of-roguelike-gameplay),
  [RogueBasin: Permadeath](https://www.roguebasin.com/index.php/Permadeath)).
- Rogue co-creator Glenn Wichman: "more important to the roguelike genre… is simply the
  inability to undo decisions that lead to death, rather than permanently stopping someone's
  playthrough" — permadeath was "never supposed to be about pain"
  ([Game Developer interview](https://www.gamedeveloper.com/design/-i-rogue-i-co-creator-permadeath-was-never-supposed-to-be-about-pain-)).
- "Unlike an RPG, where starting over can involve doing the same hundred page conversation
  again, a roguelike presents you with fresh challenges every game" — each death starts a
  new narrative arc rather than a repeat; irreversibility "increases the feeling of
  responsibility for your actions… The player is more attached to their character, for fear
  of losing them" ([Cartridge: What is a Roguelike?](https://cartridge.gg/blog/what-is-a-roguelike/),
  [Game Developer: Death in Gaming](https://www.gamedeveloper.com/design/death-in-gaming-roguelikes-and-quot-rogue-legacy-quot-)).

## God games — indirect control and the divine intermediary

- Genre definition: "the player does not move individual units like in an RTS, but instead
  influences the environment or general directives, and the simulation responds"; god games
  are "a subgenre of artificial life game, where players use supernatural powers to
  indirectly influence a population of simulated worshipers"
  ([Galaxus: evolution of the god sim](https://www.galaxus.at/en/page/populous-black-white-and-beyond-the-evolution-of-the-god-simulation-39003),
  [Game Developer: God Games: Impostors in the Pantheon](https://www.gamedeveloper.com/design/god-games-impostors-in-the-pantheon)).
- Intervention vocabulary: "causing natural disasters, blessing crops, or answering
  prayers"; "manifestations include miracles, disasters, or environmental changes," plus
  terraforming ([Galaxus](https://www.galaxus.at/en/page/populous-black-white-and-beyond-the-evolution-of-the-god-simulation-39003)).
- **Mana/power economy:** "players must economize quantities of power or mana derived from
  the size and prosperity of their population of worshipers… a positive feedback loop where
  more power allows the player to help their population grow and gain more power"
  ([Game Developer: Impostors in the Pantheon](https://www.gamedeveloper.com/design/god-games-impostors-in-the-pantheon)).
- Canon: Populous (Molyneux) defined the genre; Black & White and WorldBox continue it —
  "almost divine intervention options without directly controlling individual units"
  ([Galaxus](https://www.galaxus.at/en/page/populous-black-white-and-beyond-the-evolution-of-the-god-simulation-39003)).

## Observation-driven / idle play

- Idle-game framing: "you set up initial conditions for a simulation and watch the resulting
  complexity unfold. You can interact with the simulation sometimes as it runs, but most of
  the time you just enjoy watching it play out"
  ([Bits n Pixels: idle games](https://bitsnpixels.org/p/best-idle-clicker-games-2025)).
- **Progress Quest** (EverQuest satire): "you make an RPG character… then hit start, and the
  rest of the experience happens on its own. As the game runs, your character does quests on
  their own, defeats monsters on their own and chooses their own upgrades… the game has its
  own staying power in the fact that you accrue progress over time and events indicate a
  narrative about the digital person running on your machine"
  ([Bits n Pixels](https://bitsnpixels.org/p/best-idle-clicker-games-2025)).

## Generative Agents ("Smallville") — the player's role in an LLM-agent sim

Source: [Park et al. 2023, *Generative Agents: Interactive Simulacra of Human Behavior*](https://arxiv.org/abs/2304.03442)
(arXiv fetched 2026-07-24); [Emergent Mind topic page](https://www.emergentmind.com/topics/generative-agents-smallville).

- Setting: 25 agents in a Sims-like sandbox town (library, cafe, college, shops, houses),
  each seeded with "a paragraph description of their personality and motivations with no
  other information provided" ([Emergent Mind](https://www.emergentmind.com/topics/generative-agents-smallville)).
- Architecture: agents "store a complete record of the agent's experiences using natural
  language, synthesize those memories over time into higher-level reflections, and retrieve
  them dynamically to plan behavior" ([arXiv](https://arxiv.org/abs/2304.03442)).
- **Human interaction is natural-language and persona-based**: "end users interact with the
  sandbox environment through natural language," taking on personas to converse with agents
  ([arXiv](https://arxiv.org/abs/2304.03442), [Emergent Mind](https://www.emergentmind.com/topics/generative-agents-smallville)).
- Emergence showcase: "starting with only a single user-specified notion that one agent
  wants to throw a Valentine's Day party, the agents autonomously spread invitations…, make
  new acquaintances, ask each other out on dates to the party, and coordinate to show up for
  the party together at the right time" — information diffusion, relationship formation, and
  temporal coordination from one seeded intention ([arXiv](https://arxiv.org/abs/2304.03442)).

## Cogmind — dynamic depth and difficulty framing

Source: [Grid Sage Games: "Rebranding Difficulty Modes"](https://www.gridsagegames.com/blog/2019/09/rebranding-difficulty-modes/)
(fetched 2026-07-24); [Grid Sage Games: Adjustable Difficulty](https://www.gridsagegames.com/blog/2017/02/adjustable-difficulty/);
[Cogmind FAQ](https://www.gridsagegames.com/cogmind/faq.html).

- **Dynamic depth:** "at the simplest level you can spend 15 minutes shooting robots without
  worrying about details, just attaching the highest-rated parts you find," while deeper
  players "examine stats and abilities, put together a focused build, hack terminals for
  intel, and explore alternate areas with different story consequences"
  ([Cogmind FAQ](https://www.gridsagegames.com/cogmind/faq.html)).
- **Difficulty rebranding:** original names were "Easy/Easier/Easiest" below the default —
  "me being way too conservative, emphasizing the 'proper' way to play, and everything else
  is below that." Renamed to **Rogue / Adventurer / Explorer** "specifically to avoid
  psychological bias": "Would you rather lower the difficulty setting from the default to
  'easier,' or switch from Rogue to Adventurer?"
  ([Rebranding Difficulty Modes](https://www.gridsagegames.com/blog/2019/09/rebranding-difficulty-modes/)).
- Before rebranding "only about 9.5% of players were using non-default modes" — players
  didn't know the settings existed (not prominent in menus, manual, or tutorials) and the
  game defaulted to the hardest mode. Fix: **a dedicated difficulty selection menu on first
  startup** "to ensure every player made an informed choice"
  ([Rebranding Difficulty Modes](https://www.gridsagegames.com/blog/2019/09/rebranding-difficulty-modes/)).
- Trade-off acknowledged: Cogmind "relies on a fairly tight design," and easier settings
  "somewhat destabilize that design, resulting in a less consistent difficulty curve"
  ([Adjustable Difficulty](https://www.gridsagegames.com/blog/2017/02/adjustable-difficulty/)).

## Sources

- https://rimworldwiki.com/wiki/About_RimWorld — RimWorld as "story generator", L4D AI Director lineage
- https://rimworldwiki.com/wiki/AI_Storytellers — storyteller personalities and mechanics
- https://gamepadsquire.com/blog/rimworld/rimworld-storytellers-difficulty-guide/ — storyteller = difficulty control surface
- https://www.gamedeveloper.com/design/rimworld-dwarf-fortress-and-procedurally-generated-story-telling — story watchers / incident generators
- https://www.gamedeveloper.com/design/dwarf-fortress-and-rimworld-tell-very-different-stories — sim-driven vs director-driven contrast
- https://journals.sagepub.com/doi/full/10.1177/15554120231162418 — Cartlidge 2024, DF finitude/absurdity/narrative
- https://www.researchgate.net/publication/356686095_Characterization_and_Emergent_Narrative_in_Dwarf_Fortress — emergent narrative definition
- https://medium.com/@christian.marques/dwarf-fortress-an-engineering-marvel-of-the-21st-century-2ba3a1e9b95f — DF simulation depth
- https://en.wikipedia.org/wiki/Boatmurdered — the canonical DF community chronicle
- https://www.roguebasin.com/index.php/Permadeath — permadeath conventions
- https://www.gamedeveloper.com/design/-i-rogue-i-co-creator-permadeath-was-never-supposed-to-be-about-pain- — Wichman on irreversibility
- https://cartridge.gg/blog/what-is-a-roguelike/ — failure as fresh narrative arc
- https://litrpgreads.com/blog/permadeath-the-heart-of-roguelike-gameplay — permadeath philosophy
- https://www.gamedeveloper.com/design/death-in-gaming-roguelikes-and-quot-rogue-legacy-quot- — death and attachment
- https://www.galaxus.at/en/page/populous-black-white-and-beyond-the-evolution-of-the-god-simulation-39003 — god-game genre evolution
- https://www.gamedeveloper.com/design/god-games-impostors-in-the-pantheon — indirect control, mana economy
- https://bitsnpixels.org/p/best-idle-clicker-games-2025 — idle/observation framing, Progress Quest
- https://arxiv.org/abs/2304.03442 — Park et al., Generative Agents (Smallville)
- https://www.emergentmind.com/topics/generative-agents-smallville — Smallville setup and interaction summary
- https://www.gridsagegames.com/blog/2019/09/rebranding-difficulty-modes/ — Rogue/Adventurer/Explorer rebranding
- https://www.gridsagegames.com/blog/2017/02/adjustable-difficulty/ — difficulty vs tight design trade-off
- https://www.gridsagegames.com/cogmind/faq.html — dynamic depth
