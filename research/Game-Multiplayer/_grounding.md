---
title: Game-Multiplayer — Grounding
aliases: [Multiplayer Grounding]
tags: [grounding, multiplayer, netcode]
type: source
created: 2026-08-01
updated: 2026-08-01
related: ["[[Game-Multiplayer]]"]
---

# Game-Multiplayer — Grounding

> Source-of-truth artifact. Raw, cited output of a research pass. Kept close to verbatim — not
> editorialized. Knowledge notes cite *into* this file by section.

**Research question:** How do colony sims, agent sims, and open-world sandbox games implement
multiplayer — what synchronization architectures do they use, how do they handle fidelity
(determinism/desync), latency, and scale limits; what multiplayer *gameplay* shapes exist
(shared world vs parallel settlements, per-player deity roles, inter-settlement interaction);
and does world/map size have to scale with player count?

**Method:** web-search fan-out (12 searches) + targeted primary-source fetches · 2026-08-01.
The `deep-research` harness was deliberately not used (operator instruction); this is the
documented fallback path.

---

## § Dwarf Fortress — no multiplayer, and why

Dwarf Fortress ships **no multiplayer mode** in either fortress or adventure mode. Community
and developer discussion attributes this to the architecture rather than to a design refusal:
the game "was built around a single thread," and retrofitting it is "a TON of work" — "games
that use multiple threads are usually built that way from the ground up." The systems that
resist parallelization are named explicitly: "AI, character logic, tasks and errands,
pathfinding, and anything to do with rendering" — precisely the systems any multiplayer
implementation would have to keep consistent across instances. Bay 12 hired a programmer from
the modding community for performance work, and recent versions have added **experimental
multithreading support**, but this is a performance effort, not a multiplayer one.
([Steam — Multi-threading?](https://steamcommunity.com/app/975370/discussions/0/3709306945111910300/),
[Beehaw — DF adds experimental multithreading](https://beehaw.org/post/853274),
[Wikipedia — Dwarf Fortress](https://en.wikipedia.org/wiki/Dwarf_Fortress))

## § Dwarf Fortress — inter-settlement interaction (single-player, simulated)

Although DF has no *human* multiplayer, it models a full world of **other settlements** that
the fortress interacts with. Worldgen "controls the number of civilizations in each world";
a civilization is defined by "recognizable names and symbols, art forms, and languages …
governing systems … and the creation of complex relationships and diplomacy between other
entities, which can result in global networks of trade, the sharing of knowledge, and
conflicts … which may break out into war." Settlement types are placed by terrain: mountain
civs on mountain edges (`Ω`), plains civs as lowland wooden towns (`+`, `*`, `☼`, `#`), forest
civs in heavy forest (`î`, `¶`), connected during worldgen "by either roads or tunnels."

The interaction vocabulary between the player's site and other sites:
- **Caravans**: "the three main civilizations (humans, dwarves, elves) will send caravans to
  you in Spring/Summer/Autumn, and how you trade/gift with them can adjust their attitude
  towards you and your dwarven civilization as well as future trading."
- **Diplomacy**: "when a caravan visits your base, an Outpost Liaison will trigger a Diplomacy
  button … by which you can ask for the caravan to bring specific goods next year."
- **Diplomatic states**: "Peace is the normal state of being for most civilizations you are in
  contact with, and civilized entities will trade with you and send diplomats. War causes
  civilizations to engage in full-on warfare … entities you are at war with will potentially
  send sieges to your fort, especially if you attack them first."
([DF Wiki — Civilization](https://dwarffortresswiki.org/index.php/Civilization),
[DF Wiki — Diplomacy](https://dwarffortresswiki.org/index.php/Diplomacy),
[DF Wiki — World generation](https://dwarffortresswiki.org/index.php/World_generation),
[Slyther Games — How to Trade](https://www.slythergames.com/2022/12/21/dwarf-fortress-how-to-trade/))

## § RimWorld — two independent multiplayer mods, two different shapes

RimWorld ships no official multiplayer. Two community mods exist, and they resolve the design
question in **opposite** directions.

**(a) Zetrith's Multiplayer — one shared colony, deterministic lockstep.** "Multiplayer in
RimWorld relies on all players' games staying in lockstep — every pawn action, every random
event must occur identically for everyone." The mod enforces this with "patches to ensure
deterministic behavior across all game systems," specifically: **fixed seeds** for
deterministic outcomes across clients, **state tracking** monitoring RNG state for desync
detection, and a **push/pop state pattern** to preserve RNG state across non-deterministic
operations.

**(b) RimWorld Together — parallel colonies on a shared planet, independent pace.** Self
described: "an open sourced and community driven mod that allows you to play MULTIPLAYER, with
other people, in the same planet, at the same time, **while keeping separate colonies and
pace**! The mod focuses on creating a seamless and lag free MULTIPLAYER experience, where
players can play together with each other, **regardless of mods or DLCs**, while maintaining
the vanilla experience of the game intact!" The shared surface is the planet and a verb set,
not the tick: "every action is synced, including **visiting, raiding, trading, creating
factions, roads, sites**, and more." Feature list includes "create or join custom servers,"
"visit, raid, and spy other players," "trade with other players in real time and chat with
other players in real time," and "create and manage custom factions." Colonies keep
"independent progression for each player while still allowing interaction through visits,
raids, and trading on a shared planet."
([DeepWiki — rwmt/Multiplayer, Determinism and Desyncs](https://deepwiki.com/rwmt/Multiplayer/7-determinism-and-desyncs),
[Zetrith/Multiplayer Wiki — Desyncs](https://github.com/Zetrith/Multiplayer/wiki/Desyncs),
[Steam Workshop — RimWorld Together](https://steamcommunity.com/sharedfiles/filedetails/?id=3005289691))

## § RimWorld Multiplayer — desync causes, detection, recovery

Causes, per the mod's own wiki: "the game state can still occasionally desynchronize. **Most
of the time it happens due to unsynchronized interface interactions.**" Contributing factors:
mismatched mod versions across players, mods enabled in different orders, different mod
configurations, different RimWorld versions, corrupted game files. Community guidance is
blunter: "the number one cause of desyncs is mod mismatch or incompatible mods. If even one
mod differs between players — different version, load order, or a mod that isn't
multiplayer-safe — you will desync."

Detection and reporting: the mod "automatically generates desync files when synchronization
failures occur," saved to an `MpDesyncs` folder beside the saves directory, retaining "the
latest 10 desync files." Reports capture game state at the moment of failure. Diagnostic
guidance: "the most useful ones are the very first ones after a period of everything behaving
correctly," because they pinpoint the triggering action; reports "after trying to resync are
also useful, because they indicate there could be a problem with the **arbiter**" (a
reference-simulation process the mod runs to adjudicate which client diverged).
([Zetrith/Multiplayer Wiki — Desyncs](https://github.com/Zetrith/Multiplayer/wiki/Desyncs),
[RimWorldHub — Desyncs: Back in Sync Fast](https://rimworldhub.com/post/rimworld_desyncs_lets_get_back_in_sync))

## § Factorio — deterministic lockstep as shipped architecture

"Factorio multiplayer code uses **deterministic lockstep** to synchronize clients, which is a
method of synchronizing a game from one computer to another by **sending only the user inputs**
that control that game, rather than networking the state of the objects in the game itself."
"The Game State in Factorio is the full state of the map, player, entities, everything, and
it's simulated deterministically on all clients based on the actions received from the server."

Determinism is a hard precondition: "a fully deterministic algorithm is required when multiple
instances of Factorio are run so all instances run in a lockstep algorithm and are in sync,
because if functions produce random outputs, you can't use the lockstep architecture, as the
whole system screws up if the functions that process things don't give the same results for
each client, every time."

The named structural cost: "the most fundamental limit of lock step architecture is that **the
game speed is limited by the slowest player**, because to finish a frame input from all other
peers needs to be processed, a peer who can't run the game fast enough will slow the game down
for everyone."
([Factorio Wiki — Desynchronization](https://wiki.factorio.com/Desynchronization),
[FFF #76 — MP inside out](https://www.factorio.com/blog/post/fff-76),
[FFF #147 — Multiplayer rewrite](https://www.factorio.com/blog/post/fff-147),
[Alt-F4 #26 — Putting the Multi in Player](https://alt-f4.blog/ALTF4-26/))

## § Factorio — latency hiding (FFF #83), verbatim mechanism

Problem stated: network latency delays player actions because "all the players need to apply
all the user actions in the same order" to maintain determinism, so an action only executes
after a round trip.

Mechanism: Factorio maintains a **"latency state"** layer duplicating the relevant game state.
Each tick this layer is **reset from the real (authoritative) game state**, then **replays all
buffered local user actions that have not yet been confirmed**. The player sees results
immediately while the server confirmation is in flight.

Actions that ARE latency-hidden: player movement; entity selection; opening/closing GUIs;
building and fast-replacing; mining resources and buildings; picking items to cursor.

Actions that are NOT: "we don't plan to do any latency hiding for interacting with entities
(apart from basic operations like opening, rotating, etc.) or fighting" — complex interactions
requiring cascading state changes.

Key property: the scheme is **self-correcting**, because the latency layer is reinitialized
every tick from authoritative state, so a mispredicted local action cannot accumulate into a
desync. (The post states no specific latency figures or tick rate.)
([FFF #83 — Hide the latency](https://www.factorio.com/blog/post/fff-83))

## § Minecraft — server-authoritative, single-threaded tick, distance-based interest

Tick model: "Minecraft servers aim to run at 20 ticks per second (TPS)"; the goal is "1 tick
every 50 ms." "The Minecraft server runs at a fixed (maximum) update rate of 20 ticks/second,
and if the server is overloaded such that a tick takes more than 50 ms, **the tick rate
drops**" — i.e. the world slows rather than desyncing.

Threading: "Minecraft servers process everything — mob AI, redstone, chunk loading, player
actions — in a **single thread**." "The Minecraft server only uses a single core since it is
single-threaded. However … many tasks can benefit from multiple threads, such as Netty,
plugins, SQL databases."

Interest management is exposed as operator dials: "reducing view distance from the default 10
to 8 (or even 6) massively reduces chunk generation load. **Simulation distance** controls how
far entities tick — 5 is the sweet spot." Load attribution: "player movement, block updates,
mob AI, chunk loading, redstone, item drops, machines, and plugins all compete for processing
time. **Chunk generation is the #1 cause of lag spikes.** When players explore new territory,
the server has to generate terrain, caves, structures, and biomes in real-time."
([ServerHeron — Minecraft TPS explained](https://serverheron.com/knowledge-base/minecraft/minecraft-tps-explained),
[ServerTracker — optimization guide 2026](https://servertracker.gg/blog/minecraft-server-optimization-guide-2026),
[Paper-chan's optimization guide](https://paper-chan.moe/paper-optimization/),
[ouiheberg — Ticks, TPS and main thread](https://www.ouiheberg.com/en/documentation/article/ticks-tps-and-main-thread-on-minecraft-understanding-server-lag))

## § Terraria — server-authoritative state sync (not lockstep)

"Terraria's multiplayer architecture uses a **client-server model** where clients connect to a
server as a middleman, and **clients cannot send or receive things directly from each other**.
The server is the owner of all NPCs, and Terraria syncs position, life, and other data from the
server to the clients whenever `npc.netUpdate` has been set to true."

Non-determinism handling is explicit and manual: "with NPCs, **any non-deterministic decision
must be synced to the clients**, and desync issues happen if random choice values in AI code
are not synced properly." Observed failure modes: "network lag causes rubber-banding and NPC
desync, which are the most complained-about multiplayer issues." Transport contributes:
"Host & Play routes traffic through Steam relay servers instead of a direct connection, which
adds delay and raises packet loss risk."
([tModLoader Wiki — Basic Netcode](https://github.com/tModLoader/tModLoader/wiki/Basic-Netcode),
[WinterNode — Terraria server lag](https://winternode.com/blog/terraria/terraria-server-lag-fixes))

## § Valheim — zone-scoped object replication, and a player ceiling that is a netcode fact

Object model: "every entity in the world — players, creatures, building pieces, and tamed
animals — is represented as a **ZDO** (Zone Data Object)," an object "used for networking that
keeps data in sync between client and server."

Authority model: "the **first player to enter an area acts as the host** for the physics and
logic in that zone" — a "**peer-to-peer within a dedicated server**" hybrid. Two backends
shipped: a direct Steam networking backend (no intermediary, low latency, PC/Steam only) and a
**PlayFab/Azure relay** crossplay backend, where "crossplay players are more likely to
experience lag, timeouts, and disconnects" from the extra hop.

Scale ceiling: "Valheim's **10-player limit is a networking decision, not a game design one**,
as the ZDO object-sync system degrades under higher player counts due to bandwidth pressure."
Community analysis found "hard-coded send and receive rate limits in the ZDO manager," and
"performance degrades noticeably approaching **5–6 players** even with networking
modifications, as large player-built structures and tamed animal populations add pressure
independently of player count."

Replication inefficiency named: "**every ZDO update resends the entire ZDO**, regardless of
what updated, which is expensive from a packet standpoint." "By default, Valheim caps data
transfer at roughly **64 KB/sec**, and in large bases with 5,000+ building pieces, this cap is
reached instantly, causing everything to 'freeze' for the player."
([Edgegap — Valheim backend deep dive](https://edgegap.com/blog/valheim-multiplayer-game-backend-deep-dive),
[James A. Chambers — fixing Valheim lag / send-receive limits](https://jamesachambers.com/fixing-valheim-dedicated-server-lag-modify-send-receive-limits/),
[Valheim Wiki — Zones](https://valheim.fandom.com/wiki/Zones))

## § Project Zomboid — migrating authority to the client that cares

PZ's stated approach to combat feel under latency is **authority transfer per entity**:
"ownership of zombies that are of relevance to the player at hand will be transferred — so the
client impacted by the zombie in gameplay terms will have **zero latency authority** over the
zombie's actions. This aims to make combat identical to that of the single player experience."

The acknowledged residual problem: "challenges remain in making combat effective when attacking
a zombie that's chasing a *friend*, as that zombie will be at far higher risk of suffering from
latency and will use **client side prediction and cover-ups** based on delayed information."

Server load characterization at scale: "lag and desync usually happen when the server cannot
keep up with the amount of simulation happening at once. A multiplayer Project Zomboid server
has to process zombie AI, player actions, world changes, inventory activity, and persistent
multiplayer data **across the whole map**."
([The Indie Stone — Zed Clients devblog](https://projectzomboid.com/blog/news/2020/06/zed-clients/),
[Pine Hosting — Why PZ servers break at scale](https://pinehosting.com/blog/why-project-zomboid-servers-break-at-scale-and-how-to-fix-performance-desync-and-player-limits/))

## § Space Station 13 — one authoritative server, tick lag as an explicit dial

SS13 "is a multiplayer game running on the freeware BYOND game engine. Sessions are typically
hosted on **user-maintained and customized game servers**, and a codebase in SS13 may involve
multiple servers, with each having different code and mechanics."

Scheduling: "SS13 uses a tick-based system to manage game updates. **Ticklag** is the amount of
time between game ticks. The **Master Controller (MC)** is the primary system controlling timed
tasks and events in SS13 (lobby timer, game checks, lighting updates, atmos, etc.)."

The tradeoff is stated as a knob, not a bug: "if you have a low tick lag the server sometimes
lags while it calculates things, whereas if you have a high tick lag there's no lag as such but
**the game moves slow**." Additionally "the tick rate is modified with heuristics during lag
spikes to help manage performance" — i.e. adaptive degradation of simulation rate under load.
([BYOND forum — SS13 optimization](http://www.byond.com/forum/post/1345084),
[Goonstation forums — Let's talk lag](https://forum.ss13.co/archive/index.php?thread-5866.html=),
[Wikipedia — Space Station 13](https://en.wikipedia.org/wiki/Space_Station_13))

## § Eco — many players, one shared simulation, politics as the multiplayer mechanic

Eco is "a simulation game … in which players have to work together to create a civilization on
a virtual planet"; "a survival game where you need to work together with other players."

Scale: "Eco's `MaxConnections` setting controls how many players can be connected at once, with
the **default typically 100**, but the practical limit depends on **world size and server
resources**, as Eco simulates a full ecosystem, economy, and government."

The multiplayer *gameplay* is governance, not combat: "players can collectively decide on a
**constitution** that determines how laws are proposed and approved"; "players have the
possibility to create currencies and establish an economy for trading, form a government and
propose and vote on laws that can **restrict what other players can do** or give incentives by
applying **taxes or government grants** to specific actions." The shared pressure is a
world-level clock plus a shared externality: "while a meteor looms overhead set to strike the
planet in thirty days, a more subtle threat grows from player-interaction with the environment.
To ultimately succeed, players and their community will need to use the tools of government and
economy to find a balance between progress and protection."
([Wikipedia — Eco](https://en.wikipedia.org/wiki/Eco_(2018_video_game)),
[Eco Wiki — Collaboration](https://wiki.play.eco/en/Collaboration),
[XGamingServer — max players](https://xgamingserver.com/docs/eco/max-players),
[Massively OP — Why I Play Eco](https://massivelyop.com/2021/04/07/why-i-play-multiplayer-sandbox-eco-is-so-much-more-than-just-an-ecology-simulator/))

## § Haven & Hearth / Wurm-lineage — territory claims as the inter-village primitive

Haven & Hearth models settlements as **claims** with a radius and an upkeep economy:
- "Villages have a **village claim or idol** which sends out influence in a large area, and this
  area can be expanded with statues and banners to make the village larger."
- "**Banners expand village claims at a range of 30 tiles** in a square shape, making a
  **61×61 claim** around it."
- "Villages work as a shared claim which **prevents non-members from interacting with
  village-owned objects**, and special village officials get certain benefits."
- Upkeep: "villages maintain a pool of **authority**, which is generated as members earn LPs
  based on their intelligence and charisma; the idol, banners, and statues **drain authority**
  and lose effectiveness if the authority pool is fully drained."
- Higher tier: players "can found villages to claim large swathes of land … and found **Realms**
  to lay claim to entire Provinces by capturing ancient **Thingwalls**."
- The politics are player-run and out-of-band: "there is a forum dedicated to discussing in-game
  politics, village relations and matters of justice."
([Haven & Hearth Wiki — Village](https://havenandhearth.fandom.com/wiki/Village),
[H&H forum — Village Claims. How do they work?](https://www.havenandhearth.com/forum/viewtopic.php?f=42&t=50497),
[Haven & Hearth on Steam](https://store.steampowered.com/app/3051280/Haven__Hearth/))

## § God games — the indirect-control genre and its multiplayer form

Genre definition: "a god game is a strategy game where the player acts as a divine or
otherworldly entity influencing a population from above, reshaping the world through **indirect
means** like raising terrain, summoning weather, and commanding worshipers **rather than
controlling individual units**." Common mechanics: "terraforming, indirect unit control,
**faith or belief as a resource**, and miracle or power systems."

Multiplayer form, where it exists, is **competing deities over separate populations**: "players
may sometimes compete against other players with their own population of supporters. In
Populous specifically, the user is a deity that must lead their followers **against the
followers of a rival deity** in order to conquer and capture them." Lineage: "Peter Molyneux's
Populous in 1989 invented the template; Black & White, Spore, From Dust, and Reus carried it
forward." Godus was pitched as blending "the power, growth and scope of Populous with the
detailed construction and **multiplayer excitement** of Dungeon Keeper."
([Grokipedia — God game](https://grokipedia.com/page/God_game),
[Galaxus — evolution of the god simulation](https://www.galaxus.at/en/page/populous-black-white-and-beyond-the-evolution-of-the-god-simulation-39003),
[PCGamesN — best god games](https://www.pcgamesn.com/best-god-games-pc),
[TechCrunch — Project Godus](https://techcrunch.com/2012/11/23/god-complex-peter-molyneux-kicks-off-first-kickstarter-with-project-godus-a-grand-plan-to-recreate-the-entire-god-game-genre))

## § Asynchronous / persistent-world multiplayer as a distinct design family

Definitions in circulation:
- **Drop-in/drop-out**: "allows players to join or leave a session during gameplay."
- **Persistent world**: "the game world is always running, even when players aren't actively
  playing … servers persist, allowing players to log in at any time and interact with an ongoing
  game world."
- **Always-on server**: "except for scheduled downtime … the server is generally operated
  continuously, resulting in a game world which is persistent."
- **Asynchronous multiplayer**: "persistence is your friend since it enables asynchronous
  interactions. This contrasts with fully concurrent gameplay, as games today exist on a
  **spectrum** from fully concurrent to fully asynchronous and everything in between."
- Robustness property: "for persistent worlds, asynchronous multiplayer applications benefit
  from **server crashes having minimal impact on player experience**."
([Heathen KB — Multiplayer](https://kb.heathen.group/game-dev-resources/game-design/multiplayer),
[Game Developer — what I've learned designing multiplayer games](https://www.gamedeveloper.com/design/what-i-ve-learned-about-designing-multiplayer-games-so-far),
[McGill — Persistence in MMOs (PDF)](https://www.cs.mcgill.ca/~adenau/pub/persistance.pdf))

## § Interest management / area-of-interest

Problem statement: "bandwidth limitations are a key challenge in large-scale collaborative
virtual environments, making it **infeasible for all participants to receive all data from all
others**, so data must be filtered according to user interest."

Three canonical approaches: "**Zone-based** (spatial partitioning), **Aura-based**
(subscription area limitation), and **Visibility-based** (visibility filtering)." "Area of
interest can be defined as simply 'interested in any entity under 100 m away' and the method
used to find those entities is independent. Spatial systems suited to arbitrary object sizes
include **quadtrees**." Practical notes: "only objects currently in range 'exist' for the client
to interact with, and **hysteresis prevents excessive updates**"; the mechanism is "a
subscription list outside of a spatial query structure." "Hierarchical architecture and message
dissemination algorithms can greatly save network bandwidth and alleviate each node's workload."
Zoning as load distribution: "zoning is an efficient way to distribute server load, frequently
used in large sandbox games or open-world MMOs like World of Warcraft. Depending on the game
engine used, **a zone can typically host ~100 players**."
([Boulanger — Interest Management for MMGs, McGill thesis (PDF)](https://www.cs.mcgill.ca/~jboula2/thesis.pdf),
[GameDev.net — Interest management in an MMO](https://www.gamedev.net/forums/topic/609123-interest-management-in-an-mmo/),
[Delphi Digital — Overcoming the limits of scale in virtual worlds](https://members.delphidigital.io/reports/overcoming-the-limits-of-scale-in-virtual-worlds))

## § Determinism hazards and desync debugging (general engineering)

Named sources of non-determinism: "**float operations, uninitialized fields, hash map
iteration, and sorts without stable tiebreakers**"; also "a non-deterministic iteration order
or unseeded random number generators."

Go-specific: "**in Go, iterating over maps does not guarantee a consistent order**." This is
called out as a practical determinism defect class in Cosmos SDK blockchain code, "where
non-deterministic iteration behavior inherent in Go maps poses challenges for deterministic
operations."

Floating point: "FP calculations are entirely deterministic, as per the IEEE Floating Point
Standard, **but that doesn't mean they're entirely reproducible across machines, compilers,
OS's**, etc. However, floating-point arithmetic does not cause non-zero simulation variance for
repeated simulation runs when using the **same executable, hardware, configuration and
execution order**."

Debug methodology: "replay tools work by recording the complete initial state and all inputs
from a real game session, then replaying the simulation on different machines or builds and
**comparing state dumps at every tick**. The fastest sanity check is to run the same recorded
input sequence twice and confirm the final world state is identical down to the last value."
([Gaffer On Games — Floating Point Determinism](https://gafferongames.com/post/floating_point_determinism/),
[Ashouri — Go map iteration & determinism in Cosmos SDK](https://ashourics.medium.com/the-challenge-of-gos-map-iteration-in-the-cosmos-sdk-blockchain-a-dive-into-determinism-bd5a99260519),
[Bugnet — How to debug desync in deterministic lockstep games](https://bugnet.io/blog/how-to-debug-desync-in-deterministic-lockstep-games))

## § Screeps — a persistent always-on world where each player's *agent code* is the cost centre

Screeps is the closest published analogue to "a persistent server world where every player's
participation consumes server-side compute per tick."

Tick model: "the game operates on synchronized ticks consisting of two sequential stages:
(1) **player scripts calculation** — all active player scripts execute; (2) **commands
processing** — game world rooms process the resulting commands. After both stages complete,
database updates are applied in bulk, then the system advances to the next tick." Also: "**a
tick ends when all scripts of all players have been executed to the end**."

Per-player compute budget: "your CPU time limit depends on your Global Control Level if you have
activated CPU Unlock, or **fixed at 20** otherwise. Every player gets a basic amount of CPU (20
at the time of writing), and can unlock 10 more per GCL, up to a maximum (**300** currently)."
"CPU time limit is a duration of time in **milliseconds** during which your game script is
allowed to run within one tick. The CPU limit 100 means that after 100 ms execution of your
script will be terminated even if it has not accomplished some work."

Smoothing: "if a script during a tick worked less time than the account CPU baseline limit, the
resulting difference is added to a cumulative **bucket**. You may accumulate up to **10,000
CPU**. If the bucket contains any accumulation, your script can overrun your CPU limit using up
to **500 CPU per tick** from the amount accumulated in the bucket."

Sharding: "the consistent game world is divided into **shards**, each with its own database of
game objects, own game map, and own set of connected runtime servers. You will be able to set
your CPU limit for each shard … and their total sum should match your account limit."

Isolation: scripts run "within Node.js virtual machine contexts using the `vm` library. Each
node process spawns a separate fork without access to the parent process … If execution exceeds
the player-specific timeout, **the entire fork terminates** rather than gracefully stopping."

Infrastructure scale: "run-time computations are done in parallel on **40 quad-core dedicated
servers** … using 160 Intel Xeon E3-1231 v3 cores. **MongoDB** for each shard runs on a 24-core
machine with 128 GB of RAM and handles **30k update requests per second**." "One core
synchronously processes one room or player, preventing race conditions."
([Screeps Docs — Server-side architecture](https://docs.screeps.com/architecture.html),
[Screeps Docs — CPU limit](https://docs.screeps.com/cpu-limit.html),
[Screeps Blog — World Shards launched](https://blog.screeps.com/2017/08/shards/),
[Screeps Blog — Optimizations roadmap](https://blog.screeps.com/2017/06/optimizations/))

## § LLM-agent simulations — the Smallville lineage and the human's role

Architecture: Stanford/Google's Generative Agents placed "25 generative agents, or unique
personas with identities and goals," into a sandbox town. "The architecture for generative
agents provides a mechanism for storing a comprehensive record of an agent's experiences,
deepening its understanding of itself and the environment through **reflection**, and
**retrieving** a compact subset of that information to inform the agent's actions … the
cognitive architecture combines **memory streams, importance scoring, and a recursive
reflection loop**."

Human participation is present but framed as *intervention*, not co-simulation: "the game's core
mechanics include inter-agent communication, **user control**, and environmental interaction."
"Each agent had its own identity and memory, made its own judgments, formed relationships with
others, and even spontaneously planned a party." Positioning: "generative agents were
demonstrated as non-player characters in a Sims-style game world, and can play roles in many
interactive applications ranging from design tools to social computing systems to immersive
environments."
([Park et al., Generative Agents (arXiv 2304.03442)](https://arxiv.org/pdf/2304.03442),
[Singularity Hub — AI agents plan parties](https://singularityhub.com/2023/04/16/the-real-world-ai-agents-plan-parties-and-ask-each-other-out-on-dates-in-16-bit-virtual-town/),
[Dazed — Inside Smallville](https://www.dazeddigital.com/life-culture/article/59633/1/smallville-inside-the-wholesome-village-populated-solely-by-ai))

## § LLM inference cost and latency at multiplayer scale

Cost: "at 2026 rates, a game with 100,000 daily active players, each having ten NPC
conversations per session, is estimated to cost between **$500,000 to $2 million per year**
using models like Gemini 3 ($0.50–$1.00 per million tokens) or GPT-5 ($0.75–$1.50 per million
tokens)." More granularly: "cloud AI creates a **per-session cost of $0.01–0.05 that scales
linearly with player engagement**."

Bottleneck: "one of the main bottlenecks is the high computational cost of real-time LLM
inference, especially in **multi-agent settings where several NPCs must perceive, reason, and
respond simultaneously**. As games scale up and NPC interactions multiply, the cost becomes a
bigger issue."

Latency: "current cloud-based NPC setups average **three to seven seconds of round-trip time**,
which results in hundreds of dead frames in a 60 FPS game."

Alternative: "edge-native approaches where inference happens locally provide **zero marginal
cost per user** and scale infinitely with the player base, though this requires moving away
from cloud LLM APIs to optimized smaller models running locally."
([Inworld — LLM inference cost at scale](https://inworld.ai/resources/llm-inference-cost-at-scale),
[Naavik — AI NPCs: the future of game characters](https://naavik.co/digest/ai-npcs-the-future-of-game-characters/),
[Veriprajna — Edge AI gaming latency whitepaper](https://veriprajna.com/technical-whitepapers/gaming-ai-edge-computing-latency))

## § World size, density, and player count

Density over extent: "the true perception of scale is dictated by **playable density** — the
amount of meaningful content packed into every unit of space — which explains why games with
smaller maps can feel vastly larger than those with expansive but empty territories."

Emptiness at scale: "to achieve real-world density in a 60,000 square mile game world, you would
need a population the size of Earth, and your game likely won't have that scale of billions of
players — no game currently does. Game design decisions must **avoid player, AI, or physics
crowding**, and mechanics like large-scale PvP battles or player-driven settlements **cannot be
handled through traditional zoning**."

Scaling as a rules problem, not only a space problem: "**scaling** in game design means a game
retains a similar experience regardless of the number of players by **changing rules, numbers,
or other design elements based on player count**."

Server-side coupling: Eco's practical player cap "depends on world size and server resources";
Minecraft's dominant lag source is chunk generation as players spread into new territory; and
in Valheim, pressure comes from ZDO count per area rather than from raw map extent.
([TV Tropes — Sliding Scale of Content Density vs. Width](https://tvtropes.org/pmwiki/pmwiki.php/Main/SlidingScaleOfContentDensityVsWidth),
[GameDev.net — calculate open-world size](https://gamedev.net/forums/topic/683510-calculate-the-size-for-an-open-world-with-no-instancing/),
[Games Precipice — Player count & scalability](https://www.gamesprecipice.com/player-count-scalability/),
[Delphi Digital — Overcoming the limits of scale](https://members.delphidigital.io/reports/overcoming-the-limits-of-scale-in-virtual-worlds))

## § promptworld's own architecture (code-grounded, from `docs/wiki/`)

Facts recorded here so the branch can reason about applicability without cross-linking to the
project wiki. Verified against `docs/wiki/` at repo commit `1de512d9`.

- **Always-on daemon + attachable clients.** "A Go daemon advances the world 24/7 whether or not
  anyone is watching, and terminal clients attach and detach without affecting it."
  (`docs/wiki/overview.md`)
- **Single-goroutine authoritative simulation.** "A single goroutine in `internal/sim` owns all
  world state and advances it in deterministic ticks (1 tick = 1 game second). All external
  input enters as **commands applied at tick boundaries** and recorded as events."
  (`docs/wiki/overview.md`)
- **Event-sourced log is the source of truth.** "`internal/store` writes every event to an
  append-only SQLite log in the world's save directory; snapshots bound recovery time. The log
  is the source of truth; state is a reducer over it." (`docs/wiki/overview.md`,
  `docs/wiki/event-log.md`)
- **Deterministic RNG.** Per-decision PCG derived from `(seed, purpose, tick, index)`; no RNG
  state is carried. (`docs/wiki/deterministic-rng.md`)
- **Transport is a Unix domain socket, JSON-lines.** "`internal/ipc` serves a JSON-lines
  protocol over a Unix domain socket inside the save directory." (`docs/wiki/overview.md`,
  `docs/wiki/ipc-protocol.md`)
- **Log-shipping to clients already exists, with gapless replay.** The server "broadcasts
  committed events to each session under a non-blocking send into a `pushBufferSize = 1024`
  channel. On overflow the subscription is canceled"; a subscription "first fills from the store
  up to the log head at subscribe time (`subscribe{since}`)." Sessions are decoupled: "a client
  can die mid-write, spam garbage, or subscribe and stall, and the loop never notices."
  (`docs/wiki/ipc-server.md`)
- **Sessions are anonymous.** "IPC sessions are anonymous (`ipc/server.go:205` — no name, no
  id)"; event provenance is `Source: "planner"/"meeting"/"metatron"` with **no operator
  identity anywhere**; the cost meter is "one global monthly ceiling with no per-operator
  attribution." (board card TASK-65)
- **One world = one save directory = at most one daemon.** "Each world run is one save directory
  and at most one daemon process; multiple worlds mean multiple daemons."
  (`docs/wiki/overview.md`)
- **Map default is 64×64 tiles**, seeded and regenerated rather than stored (`DefaultSize = 64`
  in `worldmap-generation`). (`docs/wiki/worldmap-generation.md`)
- **Cognition horizon gates model calls by staleness.** "Because model turns take real wall time
  while game time keeps flowing, the cognition horizon deterministically gates every
  model-reaching decision by how stale its answer will be when it lands," with an adaptive
  governor that "turns the player's speed setting into a ceiling, not a promise."
  (`docs/wiki/cognition.md`, `docs/wiki/cognition-governor-debt.md`)
- **Villagers are sealed; the player acts through the Guardian.** The player-facing role is a
  deity-like Guardian with an editable charter, a **charge** economy, miracles, standing orders,
  designations/directives, prophecy, and an endogenous **faith** loop where "charge regen is a
  pure faith-band function." (`docs/wiki/guardian.md`, `docs/wiki/guardian-miracles.md`,
  `docs/wiki/guardian-faith.md`, `docs/wiki/guardian-designations.md`)
- **World forking already exists.** Spec 076 provides a "fork ceremony (fresh prefix log at the
  snapshot boundary, `world.forked` lineage, seed carried, wallet inherited)" plus fork/compare
  duel doors. (`docs/wiki/world-forking.md`)
- **Prior board position.** TASK-65 records that "single-player on-laptop is the likely v1
  posture, with multiplayer (self-host / modest paid hosting) undecided," and that the shape
  decision — "parallel villages vs shared village with per-player angels" — "gates everything
  else." (board card TASK-65, status To Do, labelled `deferred`)

---

## Sources

**Dwarf Fortress**
- [Steam Discussions — Multi-threading?](https://steamcommunity.com/app/975370/discussions/0/3709306945111910300/)
- [Beehaw — DF adds experimental multithreading support](https://beehaw.org/post/853274)
- [Wikipedia — Dwarf Fortress](https://en.wikipedia.org/wiki/Dwarf_Fortress)
- [DF Wiki — Civilization](https://dwarffortresswiki.org/index.php/Civilization)
- [DF Wiki — Diplomacy](https://dwarffortresswiki.org/index.php/Diplomacy)
- [DF Wiki — World generation](https://dwarffortresswiki.org/index.php/World_generation)
- [DF Wiki — War](https://dwarffortresswiki.org/index.php/War)
- [Slyther Games — Dwarf Fortress: How to Trade](https://www.slythergames.com/2022/12/21/dwarf-fortress-how-to-trade/)

**RimWorld multiplayer**
- [DeepWiki — rwmt/Multiplayer: Determinism and Desyncs](https://deepwiki.com/rwmt/Multiplayer/7-determinism-and-desyncs)
- [Zetrith/Multiplayer Wiki — Desyncs](https://github.com/Zetrith/Multiplayer/wiki/Desyncs)
- [RimWorldHub — RimWorld Multiplayer Desyncs](https://rimworldhub.com/post/rimworld_desyncs_lets_get_back_in_sync)
- [Steam Workshop — RimWorld Together (MULTIPLAYER)](https://steamcommunity.com/sharedfiles/filedetails/?id=3005289691)

**Factorio**
- [FFF #83 — Hide the latency](https://www.factorio.com/blog/post/fff-83)
- [FFF #76 — MP inside out](https://www.factorio.com/blog/post/fff-76)
- [FFF #147 — Multiplayer rewrite](https://www.factorio.com/blog/post/fff-147)
- [FFF #302 — The multiplayer megapacket](https://www.factorio.com/blog/post/fff-302)
- [Factorio Wiki — Desynchronization](https://wiki.factorio.com/Desynchronization)
- [Alt-F4 #26 — Putting the Multi in Player](https://alt-f4.blog/ALTF4-26/)

**Minecraft**
- [ServerHeron — Minecraft TPS explained](https://serverheron.com/knowledge-base/minecraft/minecraft-tps-explained)
- [ServerTracker — Minecraft server optimization guide 2026](https://servertracker.gg/blog/minecraft-server-optimization-guide-2026)
- [Paper-chan — Little guide to Minecraft server optimization](https://paper-chan.moe/paper-optimization/)
- [ouiheberg — Ticks, TPS and main thread](https://www.ouiheberg.com/en/documentation/article/ticks-tps-and-main-thread-on-minecraft-understanding-server-lag)

**Terraria**
- [tModLoader Wiki — Basic Netcode](https://github.com/tModLoader/tModLoader/wiki/Basic-Netcode)
- [WinterNode — Terraria server lag: causes and fixes](https://winternode.com/blog/terraria/terraria-server-lag-fixes)
- [Wikipedia — Netcode](https://en.wikipedia.org/wiki/Netcode)

**Valheim**
- [Edgegap — Valheim multiplayer game backend deep dive](https://edgegap.com/blog/valheim-multiplayer-game-backend-deep-dive)
- [Valheim Wiki — Zones](https://valheim.fandom.com/wiki/Zones)
- [James A. Chambers — Fixing Valheim dedicated server lag: send/receive limits](https://jamesachambers.com/fixing-valheim-dedicated-server-lag-modify-send-receive-limits/)

**Project Zomboid**
- [The Indie Stone — Zed Clients](https://projectzomboid.com/blog/news/2020/06/zed-clients/)
- [Pine Hosting — Why Project Zomboid servers break at scale](https://pinehosting.com/blog/why-project-zomboid-servers-break-at-scale-and-how-to-fix-performance-desync-and-player-limits/)

**Space Station 13**
- [BYOND Forum — SS13 optimization (creating a lag free environment)](http://www.byond.com/forum/post/1345084)
- [Goonstation Forums — Let's talk lag!](https://forum.ss13.co/archive/index.php?thread-5866.html=)
- [Wikipedia — Space Station 13](https://en.wikipedia.org/wiki/Space_Station_13)

**Eco**
- [Wikipedia — Eco (2018 video game)](https://en.wikipedia.org/wiki/Eco_(2018_video_game))
- [Eco Wiki — Collaboration](https://wiki.play.eco/en/Collaboration)
- [XGamingServer — How to set max players on your Eco server](https://xgamingserver.com/docs/eco/max-players)
- [Massively Overpowered — Why I Play: Eco](https://massivelyop.com/2021/04/07/why-i-play-multiplayer-sandbox-eco-is-so-much-more-than-just-an-ecology-simulator/)

**Haven & Hearth / territory sandboxes**
- [Haven and Hearth Wiki — Village](https://havenandhearth.fandom.com/wiki/Village)
- [H&H Forum — Village Claims. How do they work?](https://www.havenandhearth.com/forum/viewtopic.php?f=42&t=50497)
- [Haven & Hearth on Steam](https://store.steampowered.com/app/3051280/Haven__Hearth/)

**God games**
- [Grokipedia — God game](https://grokipedia.com/page/God_game)
- [Galaxus — Populous, Black & White and beyond](https://www.galaxus.at/en/page/populous-black-white-and-beyond-the-evolution-of-the-god-simulation-39003)
- [PCGamesN — The best god games](https://www.pcgamesn.com/best-god-games-pc)
- [TechCrunch — Project Godus](https://techcrunch.com/2012/11/23/god-complex-peter-molyneux-kicks-off-first-kickstarter-with-project-godus-a-grand-plan-to-recreate-the-entire-god-game-genre)

**Persistent / asynchronous multiplayer**
- [Heathen Knowledge Base — Multiplayer](https://kb.heathen.group/game-dev-resources/game-design/multiplayer)
- [Game Developer — What I've learned about designing multiplayer games so far](https://www.gamedeveloper.com/design/what-i-ve-learned-about-designing-multiplayer-games-so-far)
- [McGill (Zhang) — Persistence in Massively Multiplayer Online Games (PDF)](https://www.cs.mcgill.ca/~adenau/pub/persistance.pdf)

**Interest management & scale**
- [Boulanger — Interest Management for Massively Multiplayer Games (McGill thesis, PDF)](https://www.cs.mcgill.ca/~jboula2/thesis.pdf)
- [GameDev.net — Interest management in an MMO](https://www.gamedev.net/forums/topic/609123-interest-management-in-an-mmo/)
- [Delphi Digital — Overcoming the limits of scale in virtual worlds](https://members.delphidigital.io/reports/overcoming-the-limits-of-scale-in-virtual-worlds)
- [TV Tropes — Sliding Scale of Content Density vs. Width](https://tvtropes.org/pmwiki/pmwiki.php/Main/SlidingScaleOfContentDensityVsWidth)
- [GameDev.net — Calculate the size for an open world with no instancing?](https://gamedev.net/forums/topic/683510-calculate-the-size-for-an-open-world-with-no-instancing/)
- [Games Precipice — Player Count & Scalability](https://www.gamesprecipice.com/player-count-scalability/)

**Determinism engineering**
- [Gaffer On Games — Floating Point Determinism](https://gafferongames.com/post/floating_point_determinism/)
- [Ashouri — The challenge of Go's map iteration in the Cosmos SDK](https://ashourics.medium.com/the-challenge-of-gos-map-iteration-in-the-cosmos-sdk-blockchain-a-dive-into-determinism-bd5a99260519)
- [Bugnet — How to debug desync in deterministic lockstep games](https://bugnet.io/blog/how-to-debug-desync-in-deterministic-lockstep-games)

**Screeps**
- [Screeps Docs — Server-side architecture overview](https://docs.screeps.com/architecture.html)
- [Screeps Docs — How does CPU limit work](https://docs.screeps.com/cpu-limit.html)
- [Screeps Blog — World Shards Launched!](https://blog.screeps.com/2017/08/shards/)
- [Screeps Blog — Optimizations roadmap](https://blog.screeps.com/2017/06/optimizations/)

**LLM-agent sims & inference economics**
- [Park et al. — Generative Agents: Interactive Simulacra of Human Behavior (arXiv 2304.03442)](https://arxiv.org/pdf/2304.03442)
- [Singularity Hub — The 'Real' World: AI agents plan parties](https://singularityhub.com/2023/04/16/the-real-world-ai-agents-plan-parties-and-ask-each-other-out-on-dates-in-16-bit-virtual-town/)
- [Dazed — Inside Smallville](https://www.dazeddigital.com/life-culture/article/59633/1/smallville-inside-the-wholesome-village-populated-solely-by-ai)
- [Inworld — LLM Inference Cost at Scale](https://inworld.ai/resources/llm-inference-cost-at-scale)
- [Naavik — AI NPCs: The Future of Game Characters](https://naavik.co/digest/ai-npcs-the-future-of-game-characters/)
- [Veriprajna — Edge AI Gaming: Eliminating the 3-Second NPC Latency Crisis](https://veriprajna.com/technical-whitepapers/gaming-ai-edge-computing-latency)

**promptworld internal (code-grounded)**
- `docs/wiki/overview.md`, `ipc-server.md`, `ipc-protocol.md`, `event-log.md`,
  `deterministic-rng.md`, `cognition.md`, `cognition-governor-debt.md`,
  `worldmap-generation.md`, `guardian.md`, `guardian-faith.md`, `guardian-miracles.md`,
  `guardian-designations.md`, `world-forking.md` — pinned at repo commit `1de512d9`
- Backlog card TASK-65 — "Operator identity and attribution groundwork (deferred pending
  multiplayer decision)"
