---
title: LLM Agent Sim Interfaces
aliases: [Smallville UI, Generative Agents UI, AI Town UI]
tags: [game-ui, llm-agents, smallville, ai-town, agent-inspection]
type: note
created: 2026-07-25
updated: 2026-07-25
related: ["[[Game-UI-UX]]", "[[Chat-and-Agent-Console-Rendering]]", "[[Recurring-Interface-Patterns]]"]
---

# LLM Agent Sim Interfaces

How LLM-agent world simulations — Stanford's Generative Agents "Smallville," AI Town, and
descendants — let viewers watch agents, inspect their internals, and intervene. Facts cited in
[[_grounding]] §LLM-agent world simulations.

## Smallville (Park et al., UIST 2023)

- A Phaser-rendered overhead pixel-art town; 25 agents as sprite avatars. Ambient action
  display: the LLM translates each agent's current action into **emoji shown in a speech
  bubble** above the avatar; clicking the avatar expands to the full natural-language action
  description ([paper](https://ar5iv.labs.arxiv.org/html/2304.03442)).
- Agent conversations render as quoted natural-language dialogue; a dashboard exposes each
  agent's memory stream, current activity, and location
  ([paper](https://ar5iv.labs.arxiv.org/html/2304.03442); [Emergent Mind](https://www.emergentmind.com/topics/generative-agents-smallville)).
- Three documented intervention modes: interview an agent under a chosen persona; speak as
  the agent's "inner voice" (treated as directives); embody an agent or rewrite object states
  in natural language (set the stove to "burning") ([paper](https://ar5iv.labs.arxiv.org/html/2304.03442)).
- The public demo was a **pre-computed replay**, not a live sim — 25 agents for two sim-days
  cost "thousands of dollars in token credits" and days of wall clock. The released code
  splits an environment server (Django, browser map view) from the simulation server, with
  Replay (debug) and Demo (sprites, playback speed 1–5 via URL) modes
  ([paper](https://ar5iv.labs.arxiv.org/html/2304.03442); [repo README](https://github.com/joonspk-research/generative_agents)).
- Evaluation-as-UI: believability was assessed by *interviewing* agents and by giving human
  evaluators one agent's replay plus full access to its memory stream — following one agent
  at a time, not all 25 ([paper](https://ar5iv.labs.arxiv.org/html/2304.03442)).

## AI Town (a16z-infra / Convex)

- Live (not replayed) hosted world, PixiJS pixel art. Users spectate or join as a character;
  navigation is click-to-move with engine pathfinding ([README](https://github.com/a16z-infra/ai-town/blob/main/README.md);
  [ARCHITECTURE.md](https://github.com/a16z-infra/ai-town/blob/main/ARCHITECTURE.md)).
- Conversations are first-class objects with invite/accept/reject states and typing
  indicators so humans and agents don't talk over each other; clicking a character opens its
  conversation view ([ARCHITECTURE.md](https://github.com/a16z-infra/ai-town/blob/main/ARCHITECTURE.md)).
- Cost/liveness controls in the UI: a manual world freeze/unfreeze button, and auto-pause
  after 5 idle minutes. The engine persists state at 1 Hz while clients replay the last
  second of positional history for smooth motion (~1.5 s input latency)
  ([README](https://github.com/a16z-infra/ai-town/blob/main/README.md); [ARCHITECTURE.md](https://github.com/a16z-infra/ai-town/blob/main/ARCHITECTURE.md)).

## Scale-up and adjacent patterns

- **AgentSims** differentiates on authoring-by-GUI: build evaluation tasks by placing agents
  and buildings interactively ([arXiv 2308.04026](https://arxiv.org/abs/2308.04026)).
- **Project Sid** (Altera; 10–1,000+ agents in Minecraft): observation shifted from per-agent
  inspection to **aggregate analytics** — specialization entropy plots, sentiment-colored
  social graphs, geographic spread maps of a religion, with LM post-processing converting
  agent conversations into memes; >1,000 agents exceeded server limits
  ([arXiv 2411.00114](https://arxiv.org/html/2411.00114v1)).
- Commercial LLM-NPC demos (NVIDIA ACE "Kairos", Inworld "Origins") use a different pattern
  entirely: diegetic first-person voice conversation with no inspector panel
  ([NVIDIA](https://www.nvidia.com/en-us/geforce/news/nvidia-ace-for-games-generative-ai-npcs/);
  [New Atlas](https://newatlas.com/games/inworld-origins-ai-npc)).

## Cross-cutting documented observations

- Progressive disclosure is the dominant inspector pattern: ambient glyph → click → full
  action text → memory stream ([paper](https://ar5iv.labs.arxiv.org/html/2304.03442)).
- Replay-vs-live is a cost decision made visible in the UI (Smallville replays; AI Town
  throttles liveness with freeze/idle-pause) ([grounding cross-cutting section](https://github.com/a16z-infra/ai-town/blob/main/README.md)).
- Per-agent inspection stops scaling somewhere between 25 and 1,000 agents; past that,
  documented practice is dashboards over aggregates ([arXiv 2411.00114](https://arxiv.org/html/2411.00114v1)).

## Grounding

- [[_grounding]] — §"UIs of LLM-agent world simulations" (all claims above).
