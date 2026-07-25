---
title: Learning-Game-Design
aliases: [Learning Game Design, Teaching Games Design Patterns]
tags: [moc, learning-games, pedagogy, onboarding, retention]
type: moc
created: 2026-07-25
updated: 2026-07-25
related: []
---

# Learning-Game-Design

How games that successfully teach real skills — especially programming-shaped skills —
design their pedagogy, onboarding curves, retention, and failure-handling; plus how
observe-mostly games onboard players into the watch/act split. Researched 2026-07-25 for
TASK-120 to fill the learning-game-design gap flagged by earlier sibling research passes.

## Scope

**In:** cited facts on puzzle-as-pedagogy (Zachtronics family), tutorial/onboarding
literature (Portal, George Fan, flow theory), healthy engagement vs. dark patterns,
roguelike meta-progression and the skill-dilution debate, RimWorld's per-lesson learning
helper internals, and observe-mostly-game onboarding.
**Out:** any recommendation for promptworld — verdicts are the analysis phase's job.
Constraints and assumptions: [[Brief-and-Assumptions]].

## What is known

- **Skill-teaching puzzle games make the document the game** (TIS-100's manual, SHENZHEN
  I/O's datasheets, EXAPUNKS' zine), pose open-ended problems, and drive optimization with
  histogram comparison instead of rewards ([[Puzzle-Pedagogy-Patterns]]).
- **The tutorial literature converges on teaching inside play**: one mechanic at a time,
  doing over reading, minimal adaptive messaging, playtest-validated silent sequencing
  (Portal), with flow theory as the difficulty-curve substrate ([[Teaching-Through-Play]]).
- **Published work separates retention by coercion from retention by need satisfaction**:
  dark patterns (streaks, appointments, FOMO) vs. SDT's autonomy/competence/relatedness,
  plus session-respecting design built on natural stopping points
  ([[Healthy-Engagement-vs-Dark-Patterns]]).
- **Run-based games reframe failure as progress** via unlock ladders that double as paced
  tutorials (Slay the Spire) and stories that only move forward (Hades) — against an active
  debate that stat persistence dilutes skill expression ([[Meta-Progression-and-Failure]]).
- **RimWorld's learning helper is a scored lesson scheduler**: lessons are data defs
  (content + trigger + UI highlight + decay), selected by desire = importance ×
  knowledge-gap, with decaying knowledge and anti-spam gates ([[Learning-Helper-Anatomy]]).
- **Observe-mostly games teach through the UI itself** (single first action, progressive
  disclosure, self-paced check-in rhythms); god games proper document the observe/intervene
  split poorly, leaving community guides to fill in ([[Observe-Intervene-Onboarding]]).

## Notes

- [[Brief-and-Assumptions]] — the request's constraints, assumptions, and known thin spots
- [[Puzzle-Pedagogy-Patterns]] — Zachtronics-family teaching patterns (manual-as-artifact, histograms, open-endedness)
- [[Teaching-Through-Play]] — silent tutorials, George Fan's ten rules, flow/difficulty theory
- [[Healthy-Engagement-vs-Dark-Patterns]] — dark-pattern taxonomy vs. SDT and stopping-point research
- [[Meta-Progression-and-Failure]] — progress-across-failure mechanics and the skill-dilution debate
- [[Learning-Helper-Anatomy]] — RimWorld's ConceptDef / LessonAutoActivator lesson system internals
- [[Observe-Intervene-Onboarding]] — idle, god-game, and ambient-game onboarding into watching vs. acting

## Analyses

- [[Analysis-Pedagogy-To-UI]] — how this corpus's teaching-through-play evidence projects
  onto promptworld's TUI under the 2026-07-25 reorientation lens; operator decisions
  (lesson row, takeover ceremony/postmortem, exercise panel, guardian help section) as
  fixed constraints, sibling-branch reconciliation, ranked recommendations (2026-07-25)

## Open questions

- No first-party source on how god games teach the observe/intervene split — is there GDC
  or dev-blog material from Black & White / WorldBox / Pocket Ark developers not surfaced
  by this pass?
- Rogue Legacy's meta-progression rests on secondary sources; a Cellar Door postmortem
  would ground the archetype properly.
- "Radiant Patterns" and other ethical-engagement frameworks are described as mainly
  theoretical — has any shipped game documented adopting one?
- What do RimWorld's actual `helpText` strings look like per lesson (length, tone)? The
  def structure is grounded; the content corpus itself was not fetched.

## Grounding

- [[_grounding]] — web-search fan-out, 16 searches + 10 direct fetches, 2026-07-25
- [Game Developer: Designing Zachtronics' TIS-100](https://www.gamedeveloper.com/design/-things-we-create-tell-people-who-we-are-designing-zachtronics-i-tis-100-i-)
- [GDC Vault: How I Got My Mom to Play Through Plants vs. Zombies](https://www.gdcvault.com/play/1015541/How-I-Got-My-Mom)
- [Zagal et al.: Dark Patterns in the Design of Games (FDG 2013)](http://www.fdg2013.org/program/papers/paper06_zagal_etal.pdf)
- [Ryan, Rigby & Przybylski 2006 (SDT)](https://selfdeterminationtheory.org/SDT/documents/2006_RyanRigbyPrzybylski_MandE.pdf)
- [Inverse: Hades devs on God Mode](https://www.inverse.com/gaming/hades-god-mode-interview)
- [Chillu1/RimWorldDecompiled: LessonAutoActivator.cs](https://github.com/Chillu1/RimWorldDecompiled/blob/master/RimWorld/LessonAutoActivator.cs)
