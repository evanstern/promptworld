# Kithcraft — design brief (promptworld II)

**Name:** **Kithcraft** — operator-ratified 2026-08-19. *Kith* is the forgotten half of
"kith and kin": one's friends and neighbors, the people who are yours but not family —
exactly what the villagers are. *-craft* names the vehicle (Minecraft mod), the game loop
(crafting), and the project's aim: crafting company. "promptworld II" remains the working
lineage label; the project and its repo should be named Kithcraft.

**Status:** ratified by the operator, 2026-08-19 design session (no edits requested).
**Provenance:** distilled from a live operator/Claude design session in the promptworld I repo;
prior-art claims verified against web sources during the session.
**Audience:** this document is a self-contained handoff. An agent in a fresh project should be
able to start from this file alone; see "Handoff" at the end.

## Thesis

Survival-crafting games are already fun and already solved; what they have never had is
company. **promptworld II is Minecraft-shaped, and the LLM villagers exist to cure the
loneliness.** The AI layer is load-bearing for the *feeling* of a world that's alive — not for
the game loop, which the embodied mine/build/craft/explore cycle carries on its own. This
inverts promptworld I's failure: I built a fascinating simulation with nothing for the player
to do; II puts the player in a game that's already fun and gives them neighbors.

promptworld I is retroactively the R&D project where the theory of agent minds got worked
out. II throws away the world engine and keeps the minds.

## Decisions (ratified)

1. **Embodied player.** The player is a physical presence in the world, same as the agents.
   No god view, no guardian, no fiction about who the player is.
2. **Vehicle: Minecraft server mod.** The world, rendering, physics, crafting, and multiplayer
   substrate come free. A V2 with an owned engine is explicitly possible later via the
   body-protocol seam (below).
3. **Villager-shaped, smarter.** Not player-bodied bots — villagers with more brains and more
   tools, riding the existing village fiction (beds, workstations, schedules). Server-mod
   architecture family, not Mineflayer bot-clients.
4. **Small cast, deep bonds.** Start low (~3–6 named villagers). A household, not a city.
   Permadeath is real and death should sting; graves, persisting memories of the dead, and
   stories told about them are the intended texture (mechanics deferred).
5. **Minds are generated, not edited.** Personas, values, and **endogenous desires** are
   generated at birth and are not directly editable — this is load-bearing: company only
   counts if they're *others*. The moment a player can open a villager's head and edit its
   values, the villager is a puppet and the player is alone again. Weirdness dial starts
   conservative. Editing-via-unlocks is a deferred maybe.
6. **No direct control; orders land on lives in motion.** Villagers have their own wants and
   schedules; player work orders arrive on top. Reluctance, sloppy work from the resentful,
   and grumbling at the fire are *relationship*, not bugs. Micromanagement is the failure
   mode to design against — the player manages flow; drama emerges from interactions.
7. **Order interface: diegetic.** The v1 soul is the **job-board book** — orders are objects
   in the world that villagers read, argue about, and prioritize. Chat orders and
   ghost-block marking can compose in later.
8. **Real time only.** The game runs at 1x. This deletes promptworld I's most expensive
   subsystem (cognition horizon, governor, speed ladders, staleness budgets) — a villager
   taking 20 seconds to decide is a person mulling, not a degradation.
9. **Multiplayer v1 = "my server, me and my villagers, a friend can drop in."** The
   open-world multi-settlement dream (wars, trade, migration between players' villages) is
   on the record and deferred to V2.
10. **Replenishment punted.** No spawning/wanderer mechanism in v1; let necessity congeal the
    design (the early-Minecraft agile instinct). Session instinct for later: wanderers with
    generated pasts who arrive and are taken in, rather than birth/spawn events.

## Architecture posture

**The anti-corner move: a world-agnostic body protocol.** The mind daemon speaks
perceive / act / remember; the Minecraft mod is the first *body vendor* implementing it.
Minds never couple to Minecraft. Consequences:

- Whether the promptworld I Go daemon survives or the mind layer is rebuilt fresh is an
  implementation detail behind the seam, not an early commitment.
- A future owned world (V2) is a second body vendor, not a rewrite.
- Minds are testable without booting Minecraft.

**What transfers from promptworld I is doctrine, not code:**

- Event-sourced memory.
- The reflex/planner split — scripted competence for *doing* (pathfind, chop, place blocks),
  LLM for *choosing and relating*. Cheap reflexes for competence, expensive thoughts for
  meaning: the single most transferable lesson of I.
- Salience + consolidation + situated memory (nightly digestion heritage).
- Epistemic hygiene — an agent knows only what it saw or was told, with provenance.
- The persona firewall.

**What dies with promptworld I:** the sim engine, executor, reducer, terrain generation, TUI,
determinism-for-replay machinery, the governor/speed ladder, and the guardian (the
player-as-intermediary concept as a whole).

## Prior art (verified during the session)

**Bot-as-player family (Mineflayer-based; agents join as fake player clients):**

- [Mineflayer](https://github.com/prismarinejs/mineflayer) — mature JS bot library: physics,
  digging, building, crafting, inventory, chat; A* pathfinder plugin.
- [Voyager](https://voyager.minedojo.org/) — LLM single agent, self-curriculum, learned skill
  library, autonomous tech-tree climbing.
- [Mindcraft](https://github.com/mindcraft-bots/mindcraft) — actively maintained multi-agent
  framework; per-bot JSON profiles (name/model/prompts), any LLM provider; blueprint
  construction and cooking-collaboration benchmarks (MineCollab).
- [Project Sid](https://github.com/altera-al/project-sid) (Altera) — 500–1000 agent
  civilization runs; agents spontaneously specialized (farmers/guards/blacksmiths), obeyed
  and amended collective rules, transmitted culture/religion. PIANO architecture
  (concurrent modules, bottlenecked decision-making) is worth reading for the mind design.

**Server-mod family (NPCs as real entities in a modded server):**

- [Citizens2](https://github.com/CitizensDev/Citizens2) — the venerable Paper/Spigot NPC
  plugin; goals/behavior-tree + navigation API, persistence traits.
- [CraftAgent](https://github.com/prskid1000/CraftAgent) — Fabric mod: LLM NPCs, SQLite
  conversation memory, world perception (blocks/entities/line-of-sight), action handlers
  (mining/building/crafting/farming/combat), web dashboard.
- [AI_NPC](https://hangar.papermc.io/NNNNTX/AI_NPC) — Paper plugin: walk-up conversational
  villagers, streaming replies, function-calling actions (hand you bread, haggle), per-player
  memory, provider-agnostic, off-main-thread calls.
- Vanilla villager brain API — Fabric exposes activity/schedule injection into real
  villager brains (points of interest, memory modules, scheduled activities): the
  Elder-Scrolls-schedule substrate already exists in-engine.

**The gap II claims:** everything above is either a research benchmark or a chat skin.
Nobody has shipped *persistent villagers with memory, relationships, desires, and mortality,
cohabiting a survival world with an embodied player over weeks*. The parts all demonstrably
work; the synthesis is unclaimed.

## v1 demo — "one real evening"

Three villagers on a survival server with the player. Names, generated personas and desires,
schedules (wake/work/socialize/sleep), persistent memory. The player posts a simple blueprint
on the job board; one villager builds it while the player builds alongside. At dusk the
villagers talk to each other — about the day, about the work, about the player. Vanilla night
danger means the player's walls and torches protect their *friends*, making base-building
emotionally load-bearing.

**Spell-breakers to design against:** tedious interactions with the player; micromanagement
required to keep villagers alive or productive; villagers taking offense at a player being a
jerk (this is not a politeness simulator).

## Open questions for the next session

- Mod stack: Fabric vs Paper/Citizens base vs hybrid.
- Entity implementation: custom entity vs augmented vanilla villager.
- Mind daemon language and the body protocol's first draft.
- LLM routing/budget for a small cast at 1x real time.
- Perception model: what a villager sees/hears; porting I's epistemic rules.
- Death mechanics detail (what kills, what remains, how the living remember).

## Handoff — bootstrapping this project elsewhere

This brief is designed to be handed to an agent in a different repo/project. Steps:

1. **Carry this file.** Copy `docs/design/promptworld-2-brief.md` into the new project (or
   paste its contents into the kickoff prompt). It is self-contained: thesis, ratified
   decisions, architecture posture, prior-art map with URLs, v1 demo definition, and open
   questions. No other promptworld I artifact is required reading to start.
2. **State the receiving agent's first job.** Recommended framing: *"You are starting
   promptworld II from this ratified design brief. Do not relitigate the ratified decisions;
   your first deliverable is resolving the 'Open questions' section — starting with the mod
   stack choice and a first draft of the body protocol — and a build plan for the 'one real
   evening' demo."*
3. **Point at promptworld I as reference, not dependency.** The doctrine that transfers is
   listed under "Architecture posture." If the new agent needs the deep version of any
   doctrine item, the promptworld I repo's `docs/wiki/` corpus is the grounded reference —
   start from `INDEX.md`, load notes just-in-time. Nothing in II should import I's code.
4. **Re-verify prior art before building on it.** The prior-art links were verified
   2026-08-19. The Minecraft modding and LLM-agent spaces move fast; the receiving agent
   should re-check versions, maintenance status, and licenses before committing to any
   dependency (especially Mineflayer/Mindcraft protocol-version support and the target
   Minecraft version).
5. **Preserve the two load-bearing constraints when in doubt.** If the new project faces a
   design fork this brief doesn't cover, decide in favor of (a) the loneliness-cure thesis —
   the AI serves the feeling of company, not the game loop — and (b) minds-are-others —
   nothing may let the player directly edit a villager's mind.
