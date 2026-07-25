---
title: Game-Gameplay-Patterns
aliases: [Sim Gameplay Ideas, Colony Sim Play Patterns]
tags: [moc, gameplay, colony-sim, god-games, llm-agents]
type: moc
created: 2026-07-24
updated: 2026-07-25
related: []
---

# Game-Gameplay-Patterns

Gameplay patterns from the same family of games surveyed in the sibling player-docs branch —
colony sims (RimWorld, Dwarf Fortress), roguelikes (Rogue, Cogmind), god games (Populous,
Black & White, WorldBox), idle/observation games (Progress Quest), and LLM-agent sims
(Generative Agents/Smallville): how they pace drama, treat failure, structure the player's
mode of control, and dial intensity. Companion to the docs branch; per vault isolation the
two do not cross-link.

## Scope

**In:** drama pacing and story generation, failure-as-content, indirect/divine control,
observation-driven play, difficulty framing and dynamic depth.
**Out:** documentation/onboarding patterns (the sibling branch); recommendations for
promptworld (analysis phase). Constraints: [[Brief-and-Assumptions]].

## What is known

- **Two engines produce stories**: a director AI that reacts to player state and schedules
  pressure/release arcs (RimWorld), vs an unfettered simulation whose systems generate drama
  on their own clock (Dwarf Fortress) — with documented consequences for pacing legibility
  and player attachment ([[Simulation-vs-Director]], [[Storyteller-Driven-Pacing]]).
- **Failure is the content** in the DF/roguelike tradition: no win condition, "losing is
  fun," permadeath as irreversibility-not-pain, and the celebrated story artifacts
  (Boatmurdered) being player retellings of what the sim did
  ([[Emergent-Narrative-and-Losing-Is-Fun]]).
- **God games codify play through a divine intermediary**: no direct unit control, a
  world-level intervention vocabulary (miracles, disasters, blessings), and a mana economy
  where intervention capacity derives from the simulated population's prosperity
  ([[Indirect-Control-and-Divine-Intervention]]).
- **Watching is a viable primary verb**: idle games frame play as "set initial conditions
  and watch complexity unfold"; Progress Quest retains players on an event feed about an
  autonomous character; Smallville shows persona-based natural-language interaction and
  single-nudge cascades (the Valentine's party) as the human's role in an LLM-agent village
  ([[Observation-Driven-Play]]).
- **Intensity dials work best framed as identity, not easiness**: Cogmind's
  Easy→Explorer/Adventurer/Rogue rebrand (9.5% adoption before; first-startup choice menu
  after) and RimWorld's storyteller personas both replace an easy/hard scale with a
  fiction-level choice ([[Difficulty-Dials-and-Dynamic-Depth]]).

## Notes

- [[Brief-and-Assumptions]] — the request's constraints and assumptions
- [[Storyteller-Driven-Pacing]] — RimWorld's director AI: watchers, incident generators, personas as pacing dials
- [[Emergent-Narrative-and-Losing-Is-Fun]] — DF and roguelike failure-as-content, permadeath, retold stories
- [[Simulation-vs-Director]] — the two poles of story generation and their consequences
- [[Indirect-Control-and-Divine-Intervention]] — god-game control model and the mana economy
- [[Observation-Driven-Play]] — idle framing, Progress Quest, Generative Agents interaction modes
- [[Difficulty-Dials-and-Dynamic-Depth]] — dynamic depth and the framing of difficulty choices

## Analyses

- [[Analysis-Learning-Game-Fit]] — do these patterns serve promptworld as a prompting/agents learning game, under the operator's 2026-07-25 staged/director-lite/hybrid-scoring/faith-mana decisions; supersedes the brief's ambient-sim lens
- [[Analysis-Play-Loop-Surfaces]] — the UI/UX projection of the patterns onto concrete TUI surfaces (exercise panel, postmortem/unlock takeovers, guardian strip, fork-duel scoreboard), under the 2026-07-25 reorientation decisions and reconciled with the sibling reorientation branches

## Open questions

- WorldBox-specific loop detail (session shape, what keeps observers engaged between
  interventions) — the god-game sources found are genre-level, not WorldBox-specific.
- How Smallville-style persona interaction behaves at longer horizons — the paper's showcase
  covers a two-day cascade.
- Whether any shipped game combines a director AI *with* an LLM-agent simulation — the
  corpus shows the two separately (RimWorld's director over scripted incidents; Smallville's
  agents with no director).

## Grounding

- [[_grounding]] — the research pass this branch is built on (web-search fan-out, 2026-07-24)
- [Game Developer: How DF and RimWorld tell radically different stories](https://www.gamedeveloper.com/design/dwarf-fortress-and-rimworld-tell-very-different-stories)
- [AI Storytellers — RimWorld Wiki](https://rimworldwiki.com/wiki/AI_Storytellers)
- [Park et al. 2023, Generative Agents](https://arxiv.org/abs/2304.03442)
- [Grid Sage Games: Rebranding Difficulty Modes](https://www.gridsagegames.com/blog/2019/09/rebranding-difficulty-modes/)
