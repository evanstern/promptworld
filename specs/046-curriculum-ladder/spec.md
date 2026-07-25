# Feature Specification: The curriculum ladder — staged worlds, earned capabilities

**Feature Branch**: `046-curriculum-ladder`

**Created**: 2026-07-25

**Status**: Client-reviewed 2026-07-25 (TASK-68 AC #5 — progression alignment verified,
guardian-skin stage names approved); ready for planning

**Input**: TASK-68 (the learning game's spine) under the client's stated three-step
progression, the eight ratified synthesis decisions
(`docs/design/learning-game-synthesis.md`), the 2026-07-25 curriculum decision session,
and the TASK-121 skinnable-persona pivot.

## Ratified decisions (given constraints — this spec elaborates, never reopens)

Operator, 2026-07-25 (recorded on TASK-68/TASK-121):

1. **Four stages**: the client's three teaching stages (conversational prompting →
   instruction authoring → capability design) plus a graduation stage (full roster, the
   ambient world as endgame, canonization as its signature capability).
2. **Skin-provided names over neutral ids**: the substrate knows `stage-1..4` with
   concept-named descriptions; the active skin (TASK-121) supplies display identities.
   The default guardian skin's names are drafted here for client review.
3. **Unlock gates = scenario pass signals**: deterministic, event-derived, designer-
   paced (the Slay-the-Spire designer-controlled unlock ladder —
   `research/Learning-Game-Design/Meta-Progression-and-Failure.md`).
4. **Unlock home = per-user unlocks file + informed override**: earned stages persist
   across worlds with pointers to the proving world and events; world creation offers
   any stage via an explicit "here's what you're skipping" override. Audience posture:
   self-directed engineers. The carry-over is knowledge-shaped, not stat-shaped
   (Grid Sage "metaprogression of the mind", same vault note).
5. **Fiction constraints**: stage identities are never easy-mode framing (TASK-68
   AC #7); the default fiction is the secular-mythic guardian (TASK-121) — folk-tale
   tone, no denominational imagery.

## The ladder (the object this spec delivers)

| Stage id | Concept taught | World grants | Pass signal (gate to next) |
|---|---|---|---|
| `stage-1` | Conversational prompting: asking well, watching outcomes, iterating | Base conversational agent + basic query/nudge tools; instruction files locked (default charter in force) | Pass a stage-1 scenario (first exercise: *first-night*) |
| `stage-2` | Instruction authoring: durable behavior lives in an authored instruction file | Stage-1 grants + charter editing unlocked | Pass a stage-2 scenario **while a player-authored charter revision is in force** (provable from the charter-fingerprint evidence, spec 044) |
| `stage-3` | Capability design: what an agent *can do* is itself authored — skill files + tool grants | Stage-2 grants + skill files + the gated tool manifest opens | Pass a stage-3 scenario in which a player-granted tool's act contributes to the pass |
| `stage-4` | Mastery: indirect influence at world scale; the ambient world as the endgame | Full tool roster incl. capstone capabilities (canonization, TASK-81) | None — graduation; the ambient world is unscored (synthesis decision 3) |

One prompt-engineering concept per stage (the tutorial literature's
one-mechanic-at-a-time convergence —
`research/Learning-Game-Design/Teaching-Through-Play.md`). Injection awareness is not a
stage: the fiction teaches it natively throughout (TASK-68 description).

### Default guardian-skin display identities (client-approved 2026-07-25, AC #5)

Skin data, not substrate: **stage-1 "The Voice"** (you speak, it acts) · **stage-2
"The Written Word"** (your law outlives the conversation) · **stage-3 "The Craft"**
(you shape what it can do) · **stage-4 "The Stewardship"** (a world in your care).
Each carries a one-line identity description at world creation.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Choosing a stage is choosing an identity (Priority: P1)

Creating a world, the player is offered the stages as named identities — each stating
plainly what it teaches, what the world grants, and what evidence unlocks the next.
Stages they have earned are offered normally; any other stage is choosable through an
explicit informed override that names what they'd be skipping. Never a difficulty menu,
never a lock icon with no explanation.

**Why this priority**: the ladder's front door; every other story assumes a world knows
its stage.

**Independent Test**: create worlds at each stage (earned and overridden); verify the
choice surface presents identities with teach/grant/unlock statements, the override
requires explicit confirmation naming the skipped concept, and the created world
records its stage durably.

**Acceptance Scenarios**:

1. **Given** a new player (no unlocks), **When** they create a world, **Then** stage-1
   is the offered default, later stages are visible with their identity, concept, and
   unlock evidence stated, and choosing one requires the informed override.
2. **Given** a player who has earned stage-2, **When** they create a world, **Then**
   stages 1–2 are offered normally and the override applies only to 3–4.
3. **Given** any created world, **Then** its stage is recorded in the world's durable
   configuration and is visible from status surfaces.
4. **Given** the active skin, **Then** stage display names come from the skin while the
   underlying stage ids and semantics are skin-independent.

---

### User Story 2 - The world grants what the stage teaches (Priority: P1)

A stage-1 world offers the base conversational agent and basic tools only: instruction
files are locked (the default charter is in force and says so). A stage-2 world unlocks
charter editing. A stage-3 world opens skill files and the gated tool manifest. A
stage-4 world grants the full roster. Ungranted capabilities are structurally absent,
not prose-forbidden (the TASK-64 substrate's rule).

**Why this priority**: co-P1 — the gate-to-feature pathway is the task's core promise.

**Independent Test**: create one world per stage; verify the agent's effective tool
roster and the editability/effect of instruction files match the stage table exactly,
and that a hand-edited instruction file in a stage-1 world has no effect (with an
honest notice, not silent ignoring).

**Acceptance Scenarios**:

1. **Given** a stage-1 world, **Then** the tool roster contains only the base set and
   player edits to instruction files do not bind the agent (with a visible notice
   explaining why and naming the stage that unlocks it).
2. **Given** a stage-2 world, **Then** charter edits bind exactly as they do today in
   an ungated world; skill files and extra tools remain absent.
3. **Given** a stage-3 world, **Then** skill files compose and the capability manifest
   honors player grants; **Given** stage-4, the full roster including capstone
   capabilities is grantable.
4. **Given** any stage, **Then** capabilities beyond it are absent from the agent's
   declared toolset (structural absence), and the stage never alters world mechanics —
   villagers, needs, the gru, and events behave identically across stages.

---

### User Story 3 - Earning the next stage, and being told so in-game (Priority: P2)

Playing a stage's scenario exercise to its pass signal unlocks the next stage — and the
game says so where the player already reads: the chronicle narrates the achievement and
status surfaces show the unlock with its evidence. The unlock persists for that player
across worlds, recorded with pointers to the world and events that prove it.

**Why this priority**: the ladder moves; failure-as-progress needs the unlock moment to
be legible (run-based games reframe failure via paced unlock ladders —
`research/Learning-Game-Design/Meta-Progression-and-Failure.md`).

**Independent Test**: on a stage-1 world, drive the first-night scenario to its pass
signal; verify the unlock appears in-game (chronicle + status), the per-user unlocks
record gains a stage-2 entry pointing at this world and its pass evidence, and a
subsequent world creation offers stage-2 normally.

**Acceptance Scenarios**:

1. **Given** a stage-1 world whose first-night scenario reaches its pass signal,
   **Then** a durable, event-derived pass record exists and the unlock is surfaced
   in-game (chronicle narration + status), not only in documentation (AC #6).
2. **Given** the stage-2 gate, **Then** it requires the pass evidence to include a
   player-authored charter revision in force at pass time (charter-fingerprint
   evidence, spec 044) — a default-charter pass does not unlock stage-3.
3. **Given** an unlock, **Then** the per-user record entry names the proving world and
   evidence such that the claim is independently auditable from that world's history
   (artifact-gated, AC #8 — never a menu toggle).
4. **Given** a deleted or absent unlocks record, **Then** the player loses convenience,
   not truth: re-earning or re-deriving from a kept world's history is possible, and
   nothing about existing worlds changes.

---

### User Story 4 - Two exercises exist and teach (Priority: P2)

At least two seeded scenario exercises ship as ladder content: *first-night* (stage 1 —
keep the village alive through night one; teaches asking for the right watch: visions
and orders) and *the-law* (stage 2 — get a norm adopted; teaches durable instruction:
the charter carries policy the conversation can't). Each is defined with its seed, its
incident framing, its event-derived rubric and pass signal, and its score narrative —
the chronicle telling of how it went (synthesis decision 3: hybrid scoring; the
chronicle is the mirror).

**Why this priority**: the gates of US3 need real exercises; two prove the pattern
(TASK-68 AC #4).

**Independent Test**: inspect the two exercise definitions; verify each names seed,
stage, taught concept, rubric terms (event-derived), pass signal, and score-narrative
framing; verify the stage-2 exercise's rubric is satisfiable only with a bound charter.

**Acceptance Scenarios**:

1. **Given** the shipped exercise definitions, **Then** *first-night* and *the-law*
   exist with: stage, seed, taught concept, incident framing, event-derived rubric,
   pass signal, and score-narrative framing.
2. **Given** a completed exercise run (pass or fail), **Then** the chronicle carries a
   readable telling of the attempt, and a failed run's morgue (spec 044) serves as its
   postmortem — failure is a story, not a scold (synthesis decision 7).
3. **Given** the exercise definitions, **Then** they are expressed as content consumed
   by the scenario machinery (TASK-119) — this feature defines exercises; it does not
   build incident scheduling.

---

### User Story 5 - The stage has a floor and a guide (Priority: P3)

A stage-1 player is oriented in-game: the tutor charter preset (the grounded feedback
layer's stage-1 orientation, AC #10) greets and guides within the fiction, and each
stage has a quickstart page in the player docs (AC #9) — the per-stage manual, in the
spirit of the document-as-game-artifact
(`research/Learning-Game-Design/Puzzle-Pedagogy-Patterns.md`).

**Why this priority**: rides the other stories' structures; the `?` overlay (spec 045)
is the charter-independent floor beneath it either way.

**Independent Test**: create a stage-1 world with the tutor preset; verify in-fiction
orientation happens without player prompting. Run the player-docs freshness check;
verify a quickstart page per stage exists and passes.

**Acceptance Scenarios**:

1. **Given** a new stage-1 world with the tutor charter preset, **Then** the agent's
   early behavior includes orientation (what to watch, how to ask) delivered in-game.
2. **Given** the player docs, **Then** each stage has a quickstart page generated via
   the player-docs skill, and the freshness gate passes.
3. **Given** a no-model world at any stage, **Then** stage gating, presets, unlock
   evidence, and quickstarts all function — only tutor/narration voices are absent.

---

### Edge Cases

- Stage recorded in a world never changes for that world's lifetime (identity, not a
  dial); playing "at a higher stage" means creating a new world — consistent with
  one-directory-one-run doctrine.
- A pre-ladder world (created before this feature): treated as stage-4/ungated —
  existing worlds lose nothing.
- The unlocks record is per-user, not per-install-per-world: concurrent worlds at
  different stages coexist; unlocks earned in any of them accrue to the user.
- Override honesty: overriding to stage-4 on day one is allowed, explicit, and recorded
  in the world's config (so its runs are comparable as overridden runs).
- A pass signal reached while the player is detached: the unlock lands on the durable
  record and is narrated on next attach — never lost, never re-prompted.
- Skin changes after creation: display names change with the active skin; stage ids,
  grants, and evidence are untouched.
- Clock-tampering or fork-based replays of a pass: unlock evidence points at a specific
  world history; fork tooling (TASK-67) creating new directories does not double-count
  or fabricate unlocks.

## Requirements *(mandatory)*

### Functional Requirements

**The ladder**

- **FR-001**: The system MUST define exactly four stages with the ids, taught concepts,
  world grants, and pass signals of the ladder table; stage semantics MUST be
  skin-independent, with display identities supplied by the active skin (default
  guardian names shipped as skin data).
- **FR-002**: Stage MUST be a per-world durable configuration fact, set at creation,
  immutable for the world's lifetime, and visible on status surfaces.
- **FR-003**: World creation MUST present stages as informed identities (concept,
  grants, unlock evidence per stage), offering earned stages normally and any other
  stage only through an explicit override that names the skipped concept(s); the
  override fact MUST be recorded in the world's configuration.

**Gating (rides the TASK-64 substrate)**

- **FR-004**: The capability manifest MUST honor the world's stage: each stage grants
  exactly the ladder table's toolset; capabilities beyond the stage are structurally
  absent from the agent's declared tools.
- **FR-005**: Instruction-surface gating MUST follow the ladder: stage-1 worlds run the
  default charter regardless of player edits (with an honest, visible notice naming the
  unlocking stage); stage-2+ binds charter edits; stage-3+ binds skill files and player
  tool grants.
- **FR-006**: Stage gating MUST NOT alter simulation mechanics: identical seeds and
  histories produce identical world behavior across stages; only the agent's granted
  surface differs.

**Unlocks & evidence**

- **FR-007**: Each stage transition MUST define an event-derived pass signal per the
  ladder table (stage-2's gate requiring player-authored charter-revision evidence in
  force at pass time; stage-3's requiring a player-granted tool's act contributing to
  the pass), recorded durably in the proving world's history.
- **FR-008**: Unlocks MUST persist per user across worlds in a durable record whose
  entries point at the proving world and evidence; the record MUST be auditable (the
  claim re-derivable from the named world's history) and its loss MUST NOT corrupt or
  alter any world.
- **FR-009**: Unlock moments MUST surface in-game: chronicle narration and a status
  surface showing current unlocks; documentation is never the only surface (AC #6).

**Exercises**

- **FR-010**: At least two seeded exercises MUST ship as content: *first-night*
  (stage 1) and *the-law* (stage 2), each defining stage, seed, taught concept,
  incident framing, an event-derived rubric, a pass signal, and score-narrative
  framing; exercise definitions MUST be consumable by the scenario machinery (TASK-119)
  without redefinition.
- **FR-011**: Exercise outcomes MUST be storied: pass or fail, the chronicle narrates
  the attempt; a failed run's postmortem is the spec-044 morgue.

**Orientation & docs**

- **FR-012**: A tutor charter preset MUST ship for stage-1 orientation, delivered
  in-game through the agent's normal channels (no new mechanics), and MUST be absent-
  safe (no-model worlds simply lack the voice, never the gating).
- **FR-013**: Each stage MUST have a quickstart page in the player docs generated via
  the player-docs skill and passing its freshness gate.
- **FR-014**: All ladder machinery (stage recording, gating, pass detection, unlock
  records) MUST function with no model configured.

### Key Entities

- **Stage**: neutral id (`stage-1..4`) + taught concept + grant set + pass signal;
  skin supplies display identity.
- **World stage fact**: the world's immutable stage + optional override marker, in
  durable world configuration.
- **Pass signal**: an event-derived predicate over a world's history declaring an
  exercise passed (with its required evidence conjuncts, e.g. charter fingerprint).
- **Unlock record**: per-user, durable; entries = (stage earned, proving world,
  evidence pointers, when).
- **Exercise definition**: content object — stage, seed, framing, rubric, pass signal,
  score-narrative framing — consumed by scenario machinery.
- **Tutor charter preset**: skin-voiced instruction-file content shipped as data.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new player can go from installing to a stage-1 world with in-game
  orientation without reading anything outside the game except (optionally) the stage-1
  quickstart.
- **SC-002**: Across one world per stage with the same seed, world-mechanics histories
  are identical while the agent's granted surface differs exactly per the ladder table.
- **SC-003**: A player who passes *first-night* sees the unlock in-game within one
  attach session and can create a stage-2 world with no manual steps; the unlock entry
  audits back to the proving world's history.
- **SC-004**: A default-charter pass of the stage-2 exercise does not unlock stage-3;
  the same exercise passed with an authored charter revision in force does — the
  difference provable from recorded evidence alone.
- **SC-005**: Both shipped exercises produce a readable chronicle telling of the
  attempt; a failed first-night yields a morgue postmortem a newcomer can learn from
  (reader test: "what should my instructions have said?").
- **SC-006**: Every stage's quickstart page exists and the player-docs freshness gate
  passes; each page names its stage's concept, grants, and unlock evidence in plain
  language.
- **SC-007**: Zero dark-pattern surface: no time-gated unlocks, no streaks, no loss-
  framed pressure — unlocks derive from demonstrated competence only (vault:
  `Healthy-Engagement-vs-Dark-Patterns.md`).

## Assumptions

- The TASK-64 substrate (per-world capability manifest, skill-file composition, charter
  hot-reload, fixed-frame invariants) is the gating mechanism; this feature adds the
  stage → manifest derivation, not new gating machinery. V1 gating stays "a per-world
  stage field the capability manifest reads" (TASK-68 description).
- Scenario incident scheduling, rubric evaluation, and pass-signal machinery are
  TASK-119's deliverable; this spec defines the exercises and gates as content and
  contracts. If 119 lands after 68's implementation begins, the ambient-evidence-free
  ladder still ships (stage recording, gating, creation UX) with unlocks activating as
  119's signals arrive.
- Charter-revision evidence rides spec 044's `metatron.charter_observed` fingerprint
  event; run-loss postmortems ride 044's morgue.
- Skin machinery (display names, tutor voice) is TASK-121's deliverable; this spec
  consumes its data shape and ships the default guardian stage names as draft content
  pending client review (AC #5). If 121 lands later, stage names ship as the
  guardian-skin strings in whatever interim string table exists — the substrate ids
  never change.
- The per-user unlocks record lives outside any world directory (user scope, exact home
  decided in plan) and is convenience-plus-evidence, never authority — worlds' histories
  are the authority.
- Curriculum research grounding: `research/Learning-Game-Design/` (branch MOC
  `Learning-Game-Design.md`; gate-verified 2026-07-25) — designer-controlled unlock
  ladders and failure-reframing (`Meta-Progression-and-Failure.md`), one-concept
  sequencing and doing-over-reading (`Teaching-Through-Play.md`), document-as-artifact
  quickstarts (`Puzzle-Pedagogy-Patterns.md`), competence-not-coercion retention
  (`Healthy-Engagement-vs-Dark-Patterns.md`), lesson anatomy for future pushed lessons
  (`Learning-Helper-Anatomy.md`), and observe/intervene onboarding pacing
  (`Observe-Intervene-Onboarding.md`).
