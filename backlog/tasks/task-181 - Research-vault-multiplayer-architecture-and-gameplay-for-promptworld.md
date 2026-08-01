---
id: TASK-181
title: 'Research vault: multiplayer architecture and gameplay for promptworld'
status: In Progress
assignee: []
created_date: '2026-08-01 19:41'
updated_date: '2026-08-01 19:41'
labels:
  - research
  - multiplayer
  - vault
dependencies: []
ordinal: 163001
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Gist: a grounded research branch that surveys how multiplayer actually works in games like this one — colony sims, open-world sandboxes, persistent shared simulations, and LLM-agent villages — so the long-deferred question of what multiplayer promptworld should have can be decided from evidence instead of guesswork.

As the operator, I want to know whether two people can share a world before anyone builds for it, so the shape decision on TASK-65 stops blocking.
As a player, I want to know whether I would run my own village next to a friend's or share one with them, and what a second Guardian would even do.
As a future implementer, I want the fidelity/latency/scale tradeoffs written down with citations, so the architecture choice is not re-litigated per PR.

Delivers research/Game-Multiplayer/ — an isolated vault branch (EMBED phase only, no verdicts): a cited _grounding.md web-search pass plus eleven neutral notes covering synchronization architectures, determinism/desync, latency, scale ceilings and interest management, per-player compute budgets, multiplayer gameplay shapes, god-game multi-deity play, inter-settlement interaction verbs, world size and density, and promptworld's own code-grounded baseline. Passes the research branch gate.

Directly feeds TASK-65 AC #1 (multiplayer shape decision: parallel villages vs shared village with per-player angels), which that card states 'gates everything else'. This card is research only — it makes no recommendation; the analyze phase (research:analyze-vault) is a separate follow-on.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 research/Game-Multiplayer/ exists as an isolated vault branch and passes the research branch gate
- [x] #2 _grounding.md carries a cited research pass covering DF, RimWorld (both MP mods), Minecraft, Terraria, Valheim, Project Zomboid, SS13, Eco, Screeps, god games, and LLM-agent sims
- [x] #3 Notes cover all four questions the request named: fidelity, latency, multiplayer gameplay (Guardian-to-Guardian and village-to-village), and whether the map must grow
- [x] #4 A code-grounded baseline note records promptworld's current architecture pinned to a commit, so applicability is stated from fact not memory
- [x] #5 Home.md indexes the new branch and vault isolation holds (no cross-branch wikilinks)
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Branch research-multiplayer-vault (worktree .worktrees/research-multiplayer): research/Game-Multiplayer/ created with _grounding.md (web-search fan-out — 12 searches + 4 primary-source fetches, 2026-08-01) plus 11 neutral notes. Research branch gate passes ("OK: branch 'Game-Multiplayer' well-formed and analyzable"). EMBED phase only — no verdict authored; research:analyze-vault is the follow-on that would answer TASK-65 AC #1.
<!-- SECTION:NOTES:END -->
