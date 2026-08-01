---
id: TASK-181
title: 'Research vault: multiplayer architecture and gameplay for promptworld'
status: Done
assignee: []
created_date: '2026-08-01 19:41'
updated_date: '2026-08-01 23:21'
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

PR 153 opened (draft): https://github.com/evanstern/promptworld/pull/153 — branch research-multiplayer-vault. 13 files under research/Game-Multiplayer/ (MOC + _grounding + 11 notes). Branch gate OK; check-merge-drift pr = pass, no findings; all wikilinks resolve in-branch; Home.md indexes the branch.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Delivered research/Game-Multiplayer/ — a gate-passing vault branch (MOC + _grounding + 11 neutral notes) grounding how multiplayer works in this game's family. Evidence pass: 12 web searches + 4 primary-source fetches covering DF (no MP, architectural reason, plus its civ/caravan/diplomacy model), RimWorld's two opposed MP mods (Zetrith lockstep on one colony vs RimWorld Together's separate colonies at separate pace on a shared planet), Factorio, Minecraft, Terraria, Valheim, Project Zomboid, SS13, Eco, Haven & Hearth, Screeps, god games, LLM-agent sims and inference economics, interest management, determinism hazards, and promptworld's own architecture pinned at 1de512d9. Key findings: desync is a lockstep-specific problem that vanishes under single authority; every surveyed game sacrifices speed not consistency under load (promptworld already does this via the cognition governor); published LLM-NPC round-trips of 3-7s make model latency the responsiveness floor, not network latency; player ceilings are set by the replication scheme not the design (Valheim's 10 is documented as 'a networking decision, not a game design one'); god-game MP means competing deities over separate populations, with cooperative pantheons and asynchronous divine play both unattested; and more players does not imply a bigger map — Eco's ceiling falls as world size rises, and the family adds settlements via an abstract world-map layer rather than more tactical tiles. EMBED phase only, no verdict: research:analyze-vault is the follow-on that would answer TASK-65 AC #1 (the shape decision that card says 'gates everything else'). Merged in PR 153 as merge commit c604484c.
<!-- SECTION:FINAL_SUMMARY:END -->
