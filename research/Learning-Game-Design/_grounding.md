---
title: Learning Game Design — Grounding
aliases: []
tags: [grounding]
type: source
created: 2026-07-25
updated: 2026-07-25
related: ["[[Learning-Game-Design]]"]
---

# Learning Game Design — Grounding

> Source-of-truth artifact. This is the raw, cited output of a research pass (the `deep-research`
> skill, or a direct web-search fan-out). Keep it close to verbatim — do not editorialize, prune,
> or draw conclusions here. Knowledge notes and analyses cite *into* this file.

**Research question:** How do games that successfully teach real skills (especially
programming-shaped skills) design their pedagogy, onboarding, retention, and
failure-handling — and how do observe-mostly games explain the watch/act split to new
players?
**Method:** web-search fan-out (16 searches + 10 direct page fetches) · 2026-07-25

---

## Area 1 — Puzzle-as-programming-pedagogy (Zachtronics family and adjacent)

### TIS-100: the manual as the tutorial

- TIS-100 (Zachtronics, 2015) has players write "mock assembly language code to perform
  certain tasks on a fictional, virtualized 1970s computer that has been corrupted"
  ([Wikipedia: TIS-100](https://en.wikipedia.org/wiki/TIS-100)).
- Its 14-page technical manual "functions as the game's primary teaching tool," written to
  evoke early-'80s micro-computer documentation; players are expected to print and consult
  it. Zach Barth: "having a tutorial that basically says 'Hey, read this 14-page manual,
  asshole,' seems like it would be a really bad thing for your game, but apparently it's
  actually pretty effective"
  ([Game Developer: Designing Zachtronics' TIS-100](https://www.gamedeveloper.com/design/-things-we-create-tell-people-who-we-are-designing-zachtronics-i-tis-100-i-)).
- The manual doubles as worldbuilding: "The manual has references which imply things about
  the world, the computer itself has features which imply things about the world" (Barth).
  Writer Matthew Burns on why the story exists at all: "having a story helps you confine
  your puzzle design and direction towards something that enhances the story. Otherwise
  you're just exploring this big abstract puzzle space, and you could go anywhere"
  ([Game Developer: Designing Zachtronics' TIS-100](https://www.gamedeveloper.com/design/-things-we-create-tell-people-who-we-are-designing-zachtronics-i-tis-100-i-)).
- The audience is deliberately narrow: "TIS-100 as a game almost pushes you away… it's not
  designed to be this widely-accessible game that anyone can play" (Burns, same source).

### The document-as-artifact pattern across the catalog

- SHENZHEN I/O (2016) shipped with a manual "included as a PDF alongside print-on-demand
  copies available from Lulu"; players "use the included printable set of datasheets and
  application notes to build circuit boards using microcontrollers and integrated circuits"
  ([Zachtronics Museum](https://www.zachtronics.com/zachtronics-museum/),
  [Zachtronics: SHENZHEN I/O](https://www.zachtronics.com/shenzhen-io/)).
- EXAPUNKS (2018) repeats the pattern as fiction: "Players learn to program using the
  included printable 'hacking zine' TRASH WORLD NEWS" — zines were also sold as printed
  limited-edition sets ([Zachtronics Museum](https://www.zachtronics.com/zachtronics-museum/)).
- The studio's education program, Zachademics, makes the pedagogy explicit: "All Zachtronics
  games are free for public schools and school-like non-profit organizations"; TIS-100's
  manual "teaches players how to program and introduces concepts like assembly code,
  registers, and parallel programming"; the games "reinforce iterative problem solving and
  the design feedback loop" ([Zachtronics: Zachademics](https://www.zachtronics.com/zachademics/)).

### Opus Magnum: histograms, open-endedness, intrinsic optimization

- Opus Magnum scores every solved puzzle on three axes — cost, cycles, and area — and shows
  the player's score "both on a histogram and listed against your Steam friends"
  ([Steam discussion: level leaderboards](https://steamcommunity.com/app/558990/discussions/0/2381701715716278004/),
  [PC Gamer on Opus Magnum's histograms](https://www.pcgamer.com/perfectly-solving-opus-magnums-puzzles-is-impossible-but-thats-ok/)).
- Zachtronics on open-ended solution spaces: "As long as the toolset is expressive and the
  problem is open-ended, we will see creative solutions from players, including
  optimizations we hadn't really dreamed of before"
  ([Game Developer: Road to the IGF — Opus Magnum](https://www.gamedeveloper.com/business/road-to-the-igf-zachtronics-i-opus-magnum-i-)).
- On the toolset: "The idea is to try to create a certain minimum set of tools that create
  emergent properties when used together and balance them with cost" (same source).
- On how optimization is motivated socially, not by rewards: "One way is through the
  leaderboards. You finish a puzzle and see your friend completed it in fewer cycles, so you
  see if maybe you could shave off a few cycles yourself" — and "We deliberately don't offer
  any kind of in-game rewards for optimization, so the players who do are often more
  intrinsically motivated" (same source).
- The animated-GIF export feature "helps players share their unique solutions easily"
  (same source).

### Adjacent: Human Resource Machine, Factorio

- Human Resource Machine has been analyzed against Seymour Papert's constructionist ideas as
  "a novel approach to teaching programming that merges story, graphics, and sequential
  problem-solving to teach code concepts"
  ([ResearchGate: A Look at "Human Resource Machine" According to Papert's Ideas](https://www.researchgate.net/publication/361112910_A_Look_at_Human_Resource_Machine_According_to_Papert's_Ideas);
  teacher-facing review at [Common Sense Education](https://www.commonsense.org/education/game/human-resource-machine)).
- Practitioner writing repeatedly frames Factorio as implicit software-engineering pedagogy:
  "Factorio is a sandbox where every instinct honed in debugging, refactoring, and scaling
  finds a natural outlet"; "No one builds a megabase in one go. You build something small
  that works, iterate on it, and refactor as needed"; the game "teaches monitoring and
  observability through the constant act of checking production graphs, watching for
  bottlenecks" ([Hex Shift: How a Software Engineer Plays Factorio](https://hexshift.medium.com/how-a-software-engineer-plays-factorio-b45225fa9588),
  [Iftimie: How Factorio Shapes Better Software Engineers](https://medium.com/@iftimiealexandru/from-building-factories-to-coding-solutions-how-factorio-shapes-better-software-engineers-7b3a6e932aa6),
  [Coding Blocks: Coding Lessons… from Factorio?](https://www.codingblocks.net/programming/coding-lessons-from-factorio/)).
  Note: these are engineer blog essays, not developer statements or studies.

## Area 2 — Tutorial and onboarding-curve literature

### Portal: silent tutorial via level design + relentless playtesting

- Portal "teaches players simple systems and mechanics sequentially, then adds layers of
  complexity through the combination of mechanics and level design"; e.g. test chamber 10
  teaches the "fling" from previously learned rules rather than explicit instruction
  ([battz: Level Design of Video Games — Portal](https://battzcave.wordpress.com/2016/05/14/leveldesignofvideogames06-portal/)).
- Opening design: the player begins "in a completely safe, confined room with interactive
  elements and a radio that draws attention," letting them learn controls with no threat
  (same source).
- Valve's own tutorial-design page for Portal describes the method: the team "always focused
  on some key aspect of a mechanic and on how to teach it to players while still challenging
  them" ([Valve Developer Community: Portal — Designing Test Chambers](https://developer.valvesoftware.com/wiki/Portal_Design_And_Detail)).
- Playtesting cadence: Portal was playtested "practically every week, with a playtest on
  Friday, discussion of results on Monday, application of lessons the rest of the week, and
  testing again on Friday"; playtests started "after just one week of starting to build the
  game." Gabe Newell called playtesting "our secret weapon"
  ([GMTK / Mark Brown: Valve's "Secret Weapon"](https://gmtk.substack.com/p/valves-secret-weapon),
  [Game Developer: Thinking With Portals — Creating Valve's New IP](https://www.gamedeveloper.com/design/thinking-with-portals-creating-valve-s-new-ip)).
- If a playtester found an unintended solution "and made the tester feel smart, it might be
  left in as an alternate solution"
  ([GMTK: Valve's "Secret Weapon"](https://gmtk.substack.com/p/valves-secret-weapon)).
- Generalized "no explicit tutorial" principles from design writing: teach on a
  "need to know basis" one mechanic at a time ("Every level in these games builds slowly
  upon the previous one, introducing new concepts to the player while also requiring them to
  call upon all of the previous skills that they have learned" — on Portal); reinforce
  mechanics regularly or players forget them; "show, don't tell" via obstacles that force
  discovery (Super Metroid's creatures demonstrating the wall-jump before the player needs
  it); consistent visual affordances; leverage real-world knowledge
  ([Game Developer: No More Tutorials! How to Convey Information Through Design](https://www.gamedeveloper.com/audio/no-more-tutorials-how-to-convey-information-through-design)).

### George Fan's ten tutorial rules (GDC 2012, Plants vs. Zombies)

From "How I Got My Mom to Play Through Plants vs. Zombies"
([GDC Vault](https://www.gdcvault.com/play/1015541/How-I-Got-My-Mom),
[Game Developer write-up](https://www.gamedeveloper.com/design/gdc-2012-10-tutorial-tips-from-i-plants-vs-zombies-i-creator-george-fan)):

1. **Blend the tutorial into the game** — "teach players without them ever even realizing
   they're being taught."
2. **Prioritize doing over reading** — "The best way for a player to learn is to actually
   perform actions."
3. **Spread out mechanic introductions** — "we introduced peripheral mechanics very slowly"
   (a new plant roughly every level; new peripheral systems over many levels).
4. **Get players to perform actions once** — "Once they see the results of their action,
   that's often all it takes."
5. **Use minimal text** — "There should be a maximum of eight words on the screen at any
   given moment."
6. **Use unobtrusive messaging** — passive text that never pauses or interrupts play.
7. **Adaptive messaging** — "give tips to players who [a]re doing the wrong thing" while
   skipping tips for competent players.
8. **Avoid creating noise** — "You need to be aware what your player should be focusing on";
   "If we bombard them with one irrelevant message after another, it's like being the little
   boy who cried wolf, and the player will tune out."
9. **Use visuals to teach** — each character's look encodes its function.
10. **Leverage existing knowledge** — familiar concepts (coins, plants, zombies) carry
    mechanical meaning for free.

### Flow and difficulty-curve theory as applied to games

- Csikszentmihalyi's flow model: enjoyment requires balance between challenge and skill —
  "If the challenge is higher than the ability, the activity becomes overwhelming and
  generates anxiety. If the challenge is lower than the ability, it provokes boredom,"
  with a "fuzzy safe zone" between
  ([Jenova Chen, "Flow in Games" MFA thesis, PDF](https://www.jenovachen.com/flowingames/Flow_in_games_final.pdf);
  summary at [Medium: Flow Theory in Game Design](https://medium.com/@ahmetyunusturna/flow-theory-in-game-design-7b57430667bf)).
- Chen's thesis argues games should widen the flow zone and offer player-adjustable
  difficulty because "different people have different skills and flow zones, so a game
  designed for the 'average' player might be boring to a 'hardcore' player and frustratingly
  difficult for a 'novice' player"
  ([Chen thesis](https://www.jenovachen.com/flowingames/Flow_in_games_final.pdf); companion
  article: [Flow in games (and everything else), CACM 2007](https://www.researchgate.net/publication/220421228_Flow_in_games_and_everything_else)).
- Sweetser & Wyeth's **GameFlow** model (2005) operationalizes flow into evaluable game
  criteria — concentration, challenge, player skills, control, clear goals, feedback,
  immersion, social interaction
  ([ResearchGate: Flow in games (and everything else) — citing GameFlow](https://www.researchgate.net/publication/220421228_Flow_in_games_and_everything_else)).

## Area 3 — Healthy retention vs. dark patterns

### The dark-patterns taxonomy (games)

- Founding definition (Zagal, Björk & Lewis, FDG 2013): dark game design patterns are
  design "used intentionally by a game creator to cause negative experiences for players
  which are against their best interests and likely to happen without their consent"
  ([Zagal et al., "Dark Patterns in the Design of Games," FDG 2013 PDF](http://www.fdg2013.org/program/papers/paper06_zagal_etal.pdf)).
- The paper categorizes dark patterns as **temporal, monetary, or social**. Named examples:
  "playing by appointment" (temporal — the game dictates when you may play), "grinding"
  (temporal — repeated tedious tasks that cheat players out of time), "pay to skip"
  (monetary) ([Zagal et al. 2013](http://www.fdg2013.org/program/papers/paper06_zagal_etal.pdf)).
- A follow-up CHI 2022 paper ("A Game of Dark Patterns: Designing Healthy, Highly-Engaging
  Mobile Games," incl. Sebastian Deterding) "explores the prevalence of dark patterns in
  mobile games that exploit players through temporal, monetary, social, and psychological
  means" ([ACM DL entry](https://dl.acm.org/doi/10.1145/3491101.3519837),
  [ResearchGate record](https://www.researchgate.net/publication/360409028_A_Game_of_Dark_Patterns_Designing_Healthy_Highly-Engaging_Mobile_Games)).
- Recent work frames deceptive patterns via SDT: they are "design strategies that, while
  profitable, undermine players' psychological well-being by frustrating their needs for
  autonomy, competence, and relatedness"; streaks specifically "pressure users into constant
  engagement to avoid losing progress, which can create anxiety and compulsive behavior"
  ([Veiga et al., "Dark Patterns in Games: An Empirical Study of Their Harmfulness," 2025 PDF](https://www.scitepress.org/Papers/2025/133658/133658.pdf);
  [Medium: Gamification — Dark Patterns, Light Patterns, and Psychology](https://medium.com/@neil_62402/gamification-dark-patterns-light-patterns-and-psychology-9442d49f8b56)).
- Ethical-alternative frameworks exist but are early-stage: "Radiant Patterns" are proposed
  as responses to deceptive patterns "though they have remained mainly theoretical"
  ([Springer: Beyond Dark Patterns in Games — A Radiant Framework for Sustainable Player Well-Being](https://link.springer.com/chapter/10.1007/978-3-032-30405-6_18);
  agenda paper: [From Understanding to Intervention: Countering Dark Patterns in Games](https://link.springer.com/chapter/10.1007/978-3-032-01426-9_6)).
- A dissenting position exists in the literature: ["Against 'Dark Game Design Patterns'"](https://www.researchgate.net/publication/339054289_Against_Dark_Game_Design_Patterns)
  argues the concept itself is problematic — recorded here as evidence the taxonomy is
  debated, not settled.

### The intrinsic-motivation side (self-determination theory in games)

- Ryan, Rigby & Przybylski, "The Motivational Pull of Video Games: A Self-Determination
  Theory Approach" (Motivation and Emotion, 2006): across four studies, "perceived in-game
  autonomy and competence are associated with game enjoyment, preferences, and changes in
  well-being pre- to post-play," and "autonomy, competence, and relatedness independently
  predict enjoyment and future game play"
  ([paper PDF, selfdeterminationtheory.org](https://selfdeterminationtheory.org/SDT/documents/2006_RyanRigbyPrzybylski_MandE.pdf),
  [Springer record](https://link.springer.com/article/10.1007/s11031-006-9051-8)).
- This work produced the **PENS** measure (Player Experience of Need Satisfaction): "basic
  need satisfaction [is] the pathway to enjoyable and engaging game experiences"
  ([PENS at selfdeterminationtheory.org](https://selfdeterminationtheory.org/player-experience-of-needs-satisfaction-pens/);
  follow-up model: [Przybylski, Rigby & Ryan 2010, "A Motivational Model of Video Game Engagement" PDF](https://selfdeterminationtheory.org/SDT/documents/2010_PrzybylskiRigbyRyan_ROGP.pdf)).
- The overjustification caution: "when external rewards (points, badges, streaks) are
  introduced to tasks people originally enjoyed, intrinsic motivation often decreases"
  ([Medium: Gamification — Dark Patterns, Light Patterns, and Psychology](https://medium.com/@neil_62402/gamification-dark-patterns-light-patterns-and-psychology-9442d49f8b56)).
- FOMO analyzed through SDT: "FOMO can be understood as a social concern about a comparative
  deficit in competence and relatedness… the extrinsic coercion of FOMO drives engagement
  toward extrinsic rewards rather than, and possibly with the hindrance of, the intrinsic
  reward of pure enjoyment"
  ([Kang et al., PMC](https://www.ncbi.nlm.nih.gov/pmc/articles/PMC7735608/)).

### Session-respecting design and natural stopping points

- Research definition: "A natural stopping point is a down period after a major moment in
  action… like a save room or after a chapter completion, or after a major story
  development"; players "who regarded their play session as completed when blocked from
  playing felt less frustrated than those who were cut out in the middle of a quest"
  ([UCT thesis: Evaluating the user-experience of existing strategies to limit video game session length](https://open.uct.ac.za/handle/11427/29558);
  related: [ResearchGate: Evaluating Existing Strategies to Limit Video Game Playing Time](https://www.researchgate.net/publication/297609031_Evaluating_Existing_Strategies_to_Limit_Video_Game_Playing_Time)).
- The same research calls for "features that extend beyond extrinsic control, examining how
  to make it easier for players to quit games at their own volition, and to create games
  that have natural end points" ([UCT thesis](https://open.uct.ac.za/handle/11427/29558)).
- Industry-facing advice mirrors this: "First sessions should include not only a set of core
  on-boarding milestones for the player to complete, but also a natural end point where
  players are in the best possible state to return"
  ([Game Developer: How first session length impacts game performance](https://www.gamedeveloper.com/business/how-first-session-impacts-game-performance));
  free-to-play analysis of session-length restriction at
  [GameRefinery](https://www.gamerefinery.com/3-things-to-know-about-session-length-restriction-when-designing-a-free2play-game/).
- A CHI 2026 interview study with developers documents temporal design practice ("quarters
  per minute to daily quests and seasons"), including tracking "sessions per day, minutes
  per day, and inter-session gaps"
  ([ACM DL: Developer Perspectives on Temporal Design in Video Games](https://dl.acm.org/doi/10.1145/3772318.3790636)).

## Area 4 — Roguelike meta-progression and failure

### What meta-progression is, and its tutorial function

- Definition: "A layer of persistent progression exists above the run-level systems in a
  roguelite: upgrades, unlocks, and permanent bonuses that carry over between attempts"
  ([GameBrief glossary: Meta-Progression](https://www.gamebrief.net/glossary/meta-progression);
  design guide: [Bugnet: How to Design a Roguelite Meta-Progression](https://bugnet.io/blog/how-to-design-a-roguelite-meta-progression)).
- In roguelites "death delivers two signals: you made an error, and you still made progress,
  with meta-progression reframing the punishment loop"
  ([Switchblade Gaming: Roguelike vs Roguelite](https://www.switchbladegaming.com/strategy-games/roguelike-vs-roguelite-explained/)).
- Slay the Spire's unlocks work as a designer-paced tutorial: players earn XP per run and
  "unlock three or four cards/relics each 'level'… added to the pool of potential pickups
  for future runs, not added to your starting deck"
  ([Slay the Spire Labo: unlock guide](https://slaythespire.info/en/how-to-unlock-characters-cards-and-relics-spoiler-warning/)).
  A design note on this frames it as the "Inverted pyramid of decision making": "Start with
  simple mechanics (cards, weapons, relics) to help player learn and then gradually
  introduce new mechanics," and unlocks should be "automatic and not dependent on meta
  currency or user decisions" so the learning pathway stays designer-controlled; the
  Ascension difficulty ladder (20 tiers unlocked by winning) "keeps the game loop the same
  but lengthens the enjoyment of the game drastically over time"
  ([hamatti.org: Meta progression with gradual tutorial in roguelike games](https://notes.hamatti.org/gaming/video-games/meta-progression-with-gradual-tutorial-in-roguelike-games)).
- Content gating rationale from practitioner advice: "By putting some content (especially
  the more complicated ones) behind a meta progression wall, it helps prevent overwhelming
  new players and keeps the game fresh"
  ([Entalto Studios: 5 Essential Tips to Make Your Roguelite Game Work](https://entaltostudios.com/5-essential-tips-to-make-your-roguelite-game-work/)).

### Hades: failure as forward motion

- Creative director Greg Kasavin: "It inherently feels bad to die in a game"; Hades' focus
  from the start was "how to take the sting of failure and reduce that as much as possible."
  On the genre tension: "The part where roguelikes can be brutally difficult is, ironically,
  directly at odds with the part where they're so replayable"
  ([Inverse: Hades devs on God Mode](https://www.inverse.com/gaming/hades-god-mode-interview)).
- Hades' God Mode grants damage resistance "starting with 20 percent… adding 2 percent with
  each death" — difficulty aid that still preserves the loop. Kasavin: "If you could just
  blow through it, what's interesting about the game goes away because dying in this game
  and looping through it over and over is a really important part of the experience," and
  "God Mode reinforces our belief that the way to approach difficulty settings may need to
  be proprietary to the game. It's not a one size fits all solution"
  ([Inverse](https://www.inverse.com/gaming/hades-god-mode-interview);
  origin story at [Can I Play That?](https://caniplaythat.com/2021/08/11/hades-god-mode-explained-by-supergiant-games/)).
- Kasavin traces the narrative design to "the idea of a game with a story that would only
  move forward" — death advances the story rather than resetting it
  ([Inverse](https://www.inverse.com/gaming/hades-god-mode-interview)).

### The skill-dilution debate

- The traditional-roguelike position (Josh Ge / Grid Sage Games): traditional roguelikes
  "have little to no metaprogression — players grow through knowledge and skill rather than
  explicit benefits carried over from one run to the next"
  ([Grid Sage Games: What is a Traditional Roguelike?](https://www.gridsagegames.com/blog/2020/02/traditional-roguelike/)).
- Ge distinguishes stat unlocks from "metaprogression of the mind": "We grow as players,
  expanding our understanding of the balance and mechanics driving a particular roguelike";
  "Permadeath and being unable to take back decisions serves to highlight true mastery over
  a roguelike"; "The best players will be able to do more with less"
  ([Grid Sage Games: Designing for Mastery in Roguelikes](https://www.gridsagegames.com/blog/2025/08/designing-for-mastery-in-roguelikes-w-roguelike-radio/)).
- Player-community criticism of stat-based meta-progression: it makes games "seem designed
  for you to lose until a certain point," and players "repeat the same game a hundred times
  while getting incrementally more powerful"
  ([ResetEra: stat-based meta-progression thread](https://www.resetera.com/threads/im-starting-to-feel-that-stat-based-meta-progression-is-starting-to-ruin-roguelites-generally-speaking.1509337/),
  [ResetEra: do you like meta progression](https://www.resetera.com/threads/do-you-like-meta-progression-in-your-roguelikes-roguelites.1341955/)).
- The defense: "In titles like Hades 1 & 2, you get more resources if you play well in a
  run, meaning you're still immediately rewarded by playing well," and roguelites
  "democratized the genre"; Hades' Heat system lets skilled players re-add difficulty
  ([Switchblade Gaming: Roguelike vs Roguelite](https://www.switchbladegaming.com/strategy-games/roguelike-vs-roguelite-explained/),
  [ResetEra threads above](https://www.resetera.com/threads/do-you-like-meta-progression-in-your-roguelikes-roguelites.1341955/)).
- Rogue Legacy is repeatedly named as the archetype that popularized purchased persistent
  upgrades (the manor skill tree) in run-based games
  ([Game Rant: Roguelite Games With The Best Progression Systems](https://gamerant.com/roguelite-games-with-best-progression-systems/)).
  No first-party Cellar Door design source surfaced in this pass — noted as a gap.

## Area 5 — RimWorld learning-helper per-lesson anatomy

### Player-facing behavior (wiki)

- The learning helper "sits in the top right of the screen"; "if something happens relating
  to a concept the player hasn't learned, that lesson will be activated and shown on the
  learning helper"; lessons are shown "as needed by circumstance, or on a slow timer";
  lessons are "automatically marked as learned when the player does the necessary
  interaction, and can be marked as learned manually"; the helper "can have custom lessons
  added by mods" ([RimWorld Wiki: Learning helper](https://rimworldwiki.com/wiki/Learning_helper)).

### A lesson is a `ConceptDef` (decompiled 1.x source)

From `ConceptDef.cs` in the community-decompiled source
([Chillu1/RimWorldDecompiled: ConceptDef.cs](https://github.com/Chillu1/RimWorldDecompiled/blob/master/RimWorld/ConceptDef.cs);
repo: [RimWorldDecompiled](https://github.com/Chillu1/RimWorldDecompiled)):

- Fields per lesson: `priority` (float; default MaxValue), `helpText` (the instructional
  content), `helpTextController` (variant text for controller/Steam Deck),
  `needsOpportunity` (bool — whether the lesson requires a triggering context),
  `opportunityDecays` (bool, default true — whether a teaching opportunity expires),
  `noteTeaches` (bool), `gameMode` (which program state it applies in), and
  `highlightTags` (list of UI elements to highlight while the lesson is active).
- `TriggeredDirect` is true when `priority <= 0`; `HighlightAllTags()` visually highlights
  every UI element the lesson tagged; `ConfigErrors()` validates that priority and help
  text are present.

### Lesson selection: `LessonAutoActivator` + `PlayerKnowledgeDatabase`

From `LessonAutoActivator.cs` in the same decompiled source
([Chillu1/RimWorldDecompiled: LessonAutoActivator.cs](https://github.com/Chillu1/RimWorldDecompiled/blob/master/RimWorld/LessonAutoActivator.cs)):

- Gameplay code registers teachable moments via `TeachOpportunity()`, with priority by
  opportunity type: `GoodToKnow` = 60, `Important` = 80, `Critical` = 100.
- Relevance scoring — `GetDesire()` computes roughly
  `(conc.priority + GetOpportunity(conc) / 100 * 60) * (1 - PlayerKnowledgeDatabase.GetKnowledge(conc))`
  — i.e. accumulated priority scaled by the player's remaining knowledge gap for that
  concept, "producing higher scores for important, timely lessons on unfamiliar topics."
- Decay constants: `KnowledgeDecayRate = 0.00015` (knowledge deteriorates continuously, so
  long-unused concepts can resurface) and `OpportunityDecayRate = 0.4` (opportunities fade
  when `opportunityDecays` is true).
- Anti-spam: the activator checks every 15 frames, picks `MostDesiredConcept()`, and teaches
  only if desire exceeds a `RelaxDesire` threshold, fewer than 3 concepts are active, and no
  lesson is currently running; `timeSinceLastLesson` spaces lessons out.
- Historical context: the helper and the scripted tutorial shipped together in Alpha 15
  (2016); the tutorial "walks you through the basics of managing a colony," after which
  "the adaptive learning helper starts up"
  ([Ludeon Alpha 15 announcement](https://ludeon.com/blog/2016/08/alpha-15-tutorial-and-drugs-released/)).

## Area 6 — Observe-mostly games: explaining the watch/act split

### Idle/incremental onboarding conventions

- Incremental games onboard through **progressive disclosure driven by the UI itself**:
  Candy Box 2 opens with "a line of text telling you how many candies you have, and a button
  below which lets you 'Eat all the candies'" — one action on an almost empty screen, with
  new buttons appearing as you progress; Cookie Clicker leads with "a huge cookie… just
  begging to be clicked on" and blacks out store items until affordable, when "the first
  item in the list glows." Players learn "the way the game uses UI" — "glowing icons have
  new interactions, while greyed-out ones are temporarily unavailable" — rather than reading
  a tutorial ([Justin Chong: Onboarding in Incremental Games](https://zencron.medium.com/onboarding-in-incremental-games-bdfd9ea23e7b)).
- Genre pacing advice: layering in "auto-collectors, prestige mechanics, time boosts" too
  fast "causes new players to bail — instead, use UI to unlock complexity progressively,
  visually graying out advanced tabs or using tooltip-style onboarding when new systems
  unlock" ([GameAnalytics: How to Make an Idle Game](https://www.gameanalytics.com/blog/how-to-make-an-idle-game-adjust);
  design principles: [Eric Guan: Idle Game Design Principles](https://ericguan.substack.com/p/idle-game-design-principles)).
- Idle pacing explicitly accommodates heterogeneous check-in rhythms: "Some players might
  play all day checking in every 15 minutes, while others only check their phone before
  bedtime, and heterogenous players can optimize their resource production towards their
  preferred reengagement frequency"
  ([Eric Guan: Idle Game Design Principles](https://ericguan.substack.com/p/idle-game-design-principles)).

### God-game / sandbox onboarding (WorldBox)

- WorldBox's replayable tutorial hides behind a "Tutorial Bear" in the "Other various
  powers" menu; a player who missed the initial tutorial reported "it seemed like tutorial
  info that I need to play and I don't know where to find it" and, on being told where it
  was, "I would have never thought to look there"
  ([Steam discussion: Where Is Beginning Tutorial Like Info](https://steamcommunity.com/app/1206560/discussions/0/3195866872033170139/)).
- Community-produced beginner guides fill the gap ([Game Rant: 11 Beginner Tips for
  Worldbox](https://gamerant.com/beginner-tips-worldbox-god-simulator/); Steam guide hub
  ([Steam Community guides](https://steamcommunity.com/app/1206560/guides/)). The game's
  loop is described as: "watch miniature civilizations do their thing or aid them to stay
  protected" ([superworldbox.com](https://www.superworldbox.com/)).
- No first-party writing was found on how god games *teach* the observe/intervene split —
  this sub-area is thin; the strongest evidence is the absence itself (minimal onboarding,
  community guides doing the work).

### Pure-observation framings (zero-player and ambient games)

- Progress Quest (2002) is "the most notable precursor of idle games," satirizing RPG
  grinding: "It doesn't even require clicks — you just run it and it plays itself"
  ([Wikipedia: Progress Quest](https://en.wikipedia.org/wiki/Progress_Quest);
  concept: [Wikipedia: Zero-player game](https://en.wikipedia.org/wiki/Zero-player_game);
  essay: [Molleindustria: Games Without Players](https://www.molleindustria.org/blog/games-without-players/)).
- Mountain (David OReilly, 2014) is "billed as a simulator… somewhere between a screensaver
  and a traditional video game," with "no direct control over the environment, and only
  ambient sounds and weather to observe"; it is "designed to be played in the background
  while the player uses other applications"
  ([Wikipedia: Mountain](https://en.wikipedia.org/wiki/Mountain_(video_game));
  design appreciation: [Game Developer: There is nothing to 'do' in OReilly's Mountain — and that's a good thing](https://www.gamedeveloper.com/design/there-is-nothing-to-do-in-oreilly-s-i-mountain-i---and-that-s-a-good-thing)).

## Sources

Zachtronics family:
- [Game Developer: "Things we create tell people who we are" — Designing Zachtronics' TIS-100](https://www.gamedeveloper.com/design/-things-we-create-tell-people-who-we-are-designing-zachtronics-i-tis-100-i-)
- [Wikipedia: TIS-100](https://en.wikipedia.org/wiki/TIS-100)
- [Zachtronics Museum](https://www.zachtronics.com/zachtronics-museum/) · [SHENZHEN I/O](https://www.zachtronics.com/shenzhen-io/) · [Zachademics](https://www.zachtronics.com/zachademics/)
- [Game Developer: Road to the IGF — Zachtronics' Opus Magnum](https://www.gamedeveloper.com/business/road-to-the-igf-zachtronics-i-opus-magnum-i-)
- [PC Gamer: Perfectly solving Opus Magnum's puzzles is impossible, but that's OK](https://www.pcgamer.com/perfectly-solving-opus-magnums-puzzles-is-impossible-but-thats-ok/)
- [Steam: Opus Magnum leaderboards discussion](https://steamcommunity.com/app/558990/discussions/0/2381701715716278004/)
- [ResearchGate: A Look at "Human Resource Machine" According to Papert's Ideas](https://www.researchgate.net/publication/361112910_A_Look_at_Human_Resource_Machine_According_to_Papert's_Ideas)
- [Common Sense Education: Human Resource Machine](https://www.commonsense.org/education/game/human-resource-machine)
- [Hex Shift: How a Software Engineer Plays Factorio](https://hexshift.medium.com/how-a-software-engineer-plays-factorio-b45225fa9588) · [Iftimie](https://medium.com/@iftimiealexandru/from-building-factories-to-coding-solutions-how-factorio-shapes-better-software-engineers-7b3a6e932aa6) · [Coding Blocks](https://www.codingblocks.net/programming/coding-lessons-from-factorio/)

Tutorials / onboarding curve:
- [GDC Vault: How I Got My Mom to Play Through Plants vs. Zombies](https://www.gdcvault.com/play/1015541/How-I-Got-My-Mom) · [Game Developer write-up](https://www.gamedeveloper.com/design/gdc-2012-10-tutorial-tips-from-i-plants-vs-zombies-i-creator-george-fan)
- [Valve Developer Community: Portal — Designing Test Chambers](https://developer.valvesoftware.com/wiki/Portal_Design_And_Detail)
- [GMTK / Mark Brown: Valve's "Secret Weapon"](https://gmtk.substack.com/p/valves-secret-weapon)
- [Game Developer: Thinking With Portals — Creating Valve's New IP](https://www.gamedeveloper.com/design/thinking-with-portals-creating-valve-s-new-ip)
- [Game Developer: No More Tutorials! How to Convey Information Through Design](https://www.gamedeveloper.com/audio/no-more-tutorials-how-to-convey-information-through-design)
- [battz: Level Design of Video Games — Portal](https://battzcave.wordpress.com/2016/05/14/leveldesignofvideogames06-portal/)
- [Jenova Chen: Flow in Games (MFA thesis PDF)](https://www.jenovachen.com/flowingames/Flow_in_games_final.pdf) · [Flow in games (and everything else)](https://www.researchgate.net/publication/220421228_Flow_in_games_and_everything_else)

Healthy engagement / dark patterns:
- [Zagal, Björk & Lewis: Dark Patterns in the Design of Games (FDG 2013, PDF)](http://www.fdg2013.org/program/papers/paper06_zagal_etal.pdf)
- [ACM: A Game of Dark Patterns — Designing Healthy, Highly-Engaging Mobile Games (CHI EA 2022)](https://dl.acm.org/doi/10.1145/3491101.3519837)
- [Veiga et al.: Dark Patterns in Games — An Empirical Study of Their Harmfulness (2025, PDF)](https://www.scitepress.org/Papers/2025/133658/133658.pdf)
- [Springer: Beyond Dark Patterns in Games — A Radiant Framework](https://link.springer.com/chapter/10.1007/978-3-032-30405-6_18) · [From Understanding to Intervention](https://link.springer.com/chapter/10.1007/978-3-032-01426-9_6) · [Against "Dark Game Design Patterns"](https://www.researchgate.net/publication/339054289_Against_Dark_Game_Design_Patterns)
- [Ryan, Rigby & Przybylski 2006 (PDF)](https://selfdeterminationtheory.org/SDT/documents/2006_RyanRigbyPrzybylski_MandE.pdf) · [PENS](https://selfdeterminationtheory.org/player-experience-of-needs-satisfaction-pens/) · [Przybylski, Rigby & Ryan 2010 (PDF)](https://selfdeterminationtheory.org/SDT/documents/2010_PrzybylskiRigbyRyan_ROGP.pdf)
- [Kang et al. on FOMO and SDT (PMC)](https://www.ncbi.nlm.nih.gov/pmc/articles/PMC7735608/)
- [UCT thesis: strategies to limit session length](https://open.uct.ac.za/handle/11427/29558) · [ResearchGate companion](https://www.researchgate.net/publication/297609031_Evaluating_Existing_Strategies_to_Limit_Video_Game_Playing_Time)
- [Game Developer: How first session length impacts game performance](https://www.gamedeveloper.com/business/how-first-session-impacts-game-performance) · [GameRefinery on session-length restriction](https://www.gamerefinery.com/3-things-to-know-about-session-length-restriction-when-designing-a-free2play-game/) · [CHI 2026: Developer Perspectives on Temporal Design](https://dl.acm.org/doi/10.1145/3772318.3790636)

Meta-progression:
- [Inverse: Hades devs on God Mode](https://www.inverse.com/gaming/hades-god-mode-interview) · [Can I Play That?: God Mode origins](https://caniplaythat.com/2021/08/11/hades-god-mode-explained-by-supergiant-games/)
- [Grid Sage Games: What is a Traditional Roguelike?](https://www.gridsagegames.com/blog/2020/02/traditional-roguelike/) · [Designing for Mastery in Roguelikes](https://www.gridsagegames.com/blog/2025/08/designing-for-mastery-in-roguelikes-w-roguelike-radio/)
- [hamatti.org: Meta progression with gradual tutorial in roguelike games](https://notes.hamatti.org/gaming/video-games/meta-progression-with-gradual-tutorial-in-roguelike-games)
- [Slay the Spire Labo: unlock guide](https://slaythespire.info/en/how-to-unlock-characters-cards-and-relics-spoiler-warning/) · [Bugnet: How to Design a Roguelite Meta-Progression](https://bugnet.io/blog/how-to-design-a-roguelite-meta-progression) · [GameBrief glossary](https://www.gamebrief.net/glossary/meta-progression) · [Entalto Studios: 5 Essential Tips](https://entaltostudios.com/5-essential-tips-to-make-your-roguelite-game-work/)
- [Switchblade Gaming: Roguelike vs Roguelite](https://www.switchbladegaming.com/strategy-games/roguelike-vs-roguelite-explained/) · [ResetEra thread 1](https://www.resetera.com/threads/do-you-like-meta-progression-in-your-roguelikes-roguelites.1341955/) · [ResetEra thread 2](https://www.resetera.com/threads/im-starting-to-feel-that-stat-based-meta-progression-is-starting-to-ruin-roguelites-generally-speaking.1509337/) · [Game Rant progression list](https://gamerant.com/roguelite-games-with-best-progression-systems/)

RimWorld lesson anatomy:
- [RimWorld Wiki: Learning helper](https://rimworldwiki.com/wiki/Learning_helper)
- [Chillu1/RimWorldDecompiled: ConceptDef.cs](https://github.com/Chillu1/RimWorldDecompiled/blob/master/RimWorld/ConceptDef.cs) · [LessonAutoActivator.cs](https://github.com/Chillu1/RimWorldDecompiled/blob/master/RimWorld/LessonAutoActivator.cs)
- [Ludeon: Alpha 15 — Tutorial and Drugs released](https://ludeon.com/blog/2016/08/alpha-15-tutorial-and-drugs-released/)

Observe-mostly onboarding:
- [Justin Chong: Onboarding in Incremental Games](https://zencron.medium.com/onboarding-in-incremental-games-bdfd9ea23e7b) · [Eric Guan: Idle Game Design Principles](https://ericguan.substack.com/p/idle-game-design-principles) · [GameAnalytics: How to Make an Idle Game](https://www.gameanalytics.com/blog/how-to-make-an-idle-game-adjust)
- [Steam: WorldBox tutorial discussion](https://steamcommunity.com/app/1206560/discussions/0/3195866872033170139/) · [Game Rant: WorldBox beginner tips](https://gamerant.com/beginner-tips-worldbox-god-simulator/) · [superworldbox.com](https://www.superworldbox.com/)
- [Wikipedia: Progress Quest](https://en.wikipedia.org/wiki/Progress_Quest) · [Zero-player game](https://en.wikipedia.org/wiki/Zero-player_game) · [Molleindustria: Games Without Players](https://www.molleindustria.org/blog/games-without-players/)
- [Wikipedia: Mountain](https://en.wikipedia.org/wiki/Mountain_(video_game)) · [Game Developer: There is nothing to 'do' in Mountain](https://www.gamedeveloper.com/design/there-is-nothing-to-do-in-oreilly-s-i-mountain-i---and-that-s-a-good-thing)
