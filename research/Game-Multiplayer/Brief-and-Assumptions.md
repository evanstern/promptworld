---
title: Brief and Assumptions
aliases: [Multiplayer Brief]
tags: [brief, constraints, assumptions]
type: note
created: 2026-08-01
updated: 2026-08-01
related: ["[[Game-Multiplayer]]", "[[Promptworld-Baseline]]"]
---

# Brief and Assumptions

The request that produced this branch, restated, plus the assumptions made while gathering. This
note exists so the guesses are visible and correctable before anything is built on them.

## The request, verbatim

> create a new vault that dives into how multiplayer works on games like this one. DF (does not
> have) RW (may have?) but even if nothing has it, look at other games like minecraft, terraria,
> or other open world sandbox games. we need to figure out how we can ensure fidelity, latency,
> and other mp issues are worked out. also, gameplay needs to be disucssed, how do guradians
> interact, how do they work with other "villages". do we need to exapnd the size of the map
> significantly?

## What was asked for, decomposed

1. **Survey**: how multiplayer works in games of this family. Dwarf Fortress named as having
   none; RimWorld named as a maybe; explicit instruction to widen to Minecraft, Terraria, and
   other open-world sandbox games if the direct family comes up empty.
2. **Fidelity**: keeping every participant's view of the world consistent — determinism, desync,
   authority.
3. **Latency**: responsiveness under network round-trips, and "other MP issues."
4. **Gameplay**: how Guardians interact with each other; how a Guardian relates to *other
   villages*.
5. **Map size**: whether the world has to grow significantly to support more players.

## Answers the request already fixes (not open questions)

- The survey must include games outside the colony-sim family. Stated directly.
- Multiplayer gameplay is a **first-class** part of the question, not an afterthought to the
  netcode. Stated directly.
- "Guardians" (plural) and "other villages" (plural) are the framing. The request presumes more
  than one Guardian and more than one village can coexist; this branch therefore grounds both
  the shared-world and parallel-settlement families rather than assuming one.

## Assumptions made while gathering

- **A1 — "games like this one" means the tick-based, simulation-authoritative family**: colony
  sims (Dwarf Fortress, RimWorld), open-world sandboxes (Minecraft, Terraria, Valheim, Project
  Zomboid), persistent shared-simulation games (Eco, Space Station 13, Screeps), and LLM-agent
  sims (Generative Agents / Smallville). Twitch shooters and MOBAs were excluded: their netcode
  literature is dominated by sub-100 ms hit registration, which is not this game's problem.
- **A2 — promptworld's existing architecture is in scope as grounded fact.** The daemon is
  already a server-authoritative, event-sourced, log-shipping process with attachable clients,
  so the branch records those facts (§ promptworld's own architecture) rather than treating
  multiplayer as a greenfield question. Facts are pinned to `docs/wiki/` at commit `1de512d9`.
- **A3 — "fidelity" means state-consistency across participants**, not visual/simulation detail.
  Read the other way, the question would be about simulation depth, which is a different topic.
- **A4 — self-hosting is the presumed deployment**, per TASK-65's recorded position ("self-host /
  modest paid hosting"). Sources about 100k-DAU commercial scaling are included for the cost
  *shape*, not because that scale is assumed.
- **A5 — villagers stay sealed.** TASK-65 records that the "one agent per coworker" idea was
  retired and that indirect influence through the Guardian "is the entire point." The branch
  therefore treats *player = Guardian* as the operative unit of multiplayer, and does not
  research direct pawn control.
- **A6 — "expand the size of the map" is read as the 64×64 tile world grid**, the single map a
  world currently generates (`DefaultSize = 64`), not as a world-map-of-many-sites layer. Both
  readings are grounded, because the DF material shows the second layer is a distinct design
  object.

## Ambiguities flagged as open questions

- **Concurrency posture is unstated.** Whether "multiplayer" means *simultaneously attached
  operators* or *operators who visit an always-running world at different times* changes which
  half of the literature applies. The always-on daemon makes the asynchronous reading plausible;
  the request does not say.
- **Whether Guardians share a village or hold separate ones is undetermined** — this is exactly
  TASK-65's AC #1 ("parallel villages vs shared village with per-player angels"), still open on
  the board.
- **Whether Guardians are cooperative, competitive, or both** is unstated. The god-game lineage
  is overwhelmingly competitive (rival deities over rival followers); Eco's shared-world lineage
  is overwhelmingly cooperative-with-politics.
- **Player-count target is unstated.** The observed ceilings differ by two orders of magnitude
  across the surveyed games (Valheim ~10, Eco ~100, Screeps thousands across shards), and the
  ceiling is set by the architecture chosen, not by the map.
- **Whose LLM budget pays** is unstated and has no analogue in any non-LLM game surveyed; the
  closest published model is Screeps' per-player CPU allowance.

## Out of scope for this branch

- Any recommendation on which shape to adopt — that is the analyze phase, and TASK-65's AC #1.
- Implementation planning, specs, or estimates.
- Anti-cheat, account systems, matchmaking, monetisation.

## Grounding

- [[_grounding]] — all cited facts this branch rests on
- [[Promptworld-Baseline]] — the code-grounded starting position the request lands on
