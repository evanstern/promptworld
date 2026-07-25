# Feature Specification: Run outcomes, the morgue file, death escalation, and graves

**Feature Branch**: `044-run-outcomes-morgue`

**Created**: 2026-07-25

**Status**: Draft

**Input**: User description: "Run outcomes, death escalation, the morgue file, and graves — permadeath made consequential at the run level (TASK-31), per the six decisions of the 2026-07-25 design session recorded on the board task (rounds 1–2)."

## Design-session decisions (given constraints)

Decided by the operator on 2026-07-25 (TASK-31 comments; learning-game synthesis,
`docs/design/learning-game-synthesis.md`). This spec elaborates them; it does not
reopen them.

1. **Run end = archive.** When the last villager dies, a `run.ended` event fires, the
   simulation halts, and the daemon keeps serving reads — chronicle, morgue, event
   log, client in postmortem mode. New run = new world directory; old runs are
   browsable archives.
2. **Death escalation = the gru can kill the already-wounded.** The survival floor
   stays for healthy villagers; a hit on an already-weakened villager can kill, so
   lethality emerges from compounding preventable spirals, never one-hit randomness.
3. **Player-attributable failure.** All-dead `run.ended` is the hard failure
   everywhere; the morgue aligns each death against the angel's charter/orders
   revision timeline — evidence, never a blame score.
4. **No standalone difficulty field.** Difficulty folds into curriculum-stage /
   scenario presets (TASK-68 / TASK-119); out of scope here.
5. **Morgue = deterministic core + narrated epilogue.** The factual record is
   event-derived and fully functional with no AI configured; a prose epilogue is
   added only when a narrator model is available, and facts never depend on it.
6. **Graves v1 = marker + memory/rumor hooks.** Mourning morale effects and
   grave-visiting behaviors are out of scope.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The run ends, and the story survives it (Priority: P1)

A player returns to their village to find the last villager has died. The world does
not keep ticking emptily and does not vanish: the run is declared over, and the player
can still open the world and read everything — the chronicle of how it happened, the
event history, and the morgue. Starting fresh means creating a new world; the fallen
one remains as an archive they can revisit and learn from.

**Why this priority**: This is the contract everything else consumes — the morgue's
run-end summary, the postmortem reading experience, and (later) scenario machinery
that treats `run.ended` as its fail signal (TASK-119). Without a defined run end,
permadeath has no run-level consequence.

**Independent Test**: On a seeded world, force the deaths of all villagers (operator
door); observe the run-end declaration, verify time no longer advances, and verify
the world remains fully readable — including across a daemon restart.

**Acceptance Scenarios**:

1. **Given** a world with one living villager, **When** that villager dies, **Then** a
   run-end record with the run's summary facts (final day, population, cause of the
   final death) is appended to the world's history and time stops advancing.
2. **Given** an ended world, **When** the player opens it (status, story feed, or the
   full-screen client), **Then** all reading surfaces work, the client clearly shows
   the world is over (postmortem posture), and no new simulation activity occurs.
3. **Given** an ended world, **When** its daemon is stopped and started again,
   **Then** it comes back in the same ended/readable state and never resumes ticking.
4. **Given** an ended world, **When** the player wants to play again, **Then** the
   path is creating a new world; the ended world is never reset in place.

---

### User Story 2 - The morgue file: a legacy document worth reading (Priority: P2)

Every death leaves a durable epitaph in a single morgue document inside the world's
save directory: how long they lived, what killed them, who mattered to them, what
they owed and were owed, and what they did that the village will remember — plus what
the player's angel was instructed to watch at the time, so the player can see the gap
between their instructions and the outcome. At run end, a village-level summary
closes the document. When a narrator model is available, a short prose epilogue
follows the facts; when none is, the facts stand complete on their own.

**Why this priority**: The morgue is the learning game's honest grader and the
retellable artifact (the "Boatmurdered lesson" — the celebrated story object is a
retelling). It delivers value at every individual death even before run ends are
common, and its charter/orders alignment is the mechanism that makes failure
player-attributable (decision 3).

**Independent Test**: On a world with no AI configured, cause one villager's death;
read the morgue document and verify every factual field is present and correct
against the event history. Then configure a narrator and cause another death; verify
a clearly-separated epilogue appears and no fact changed.

**Acceptance Scenarios**:

1. **Given** a running world with no AI model configured, **When** a villager dies,
   **Then** the morgue document gains an epitaph with: name, days survived, cause,
   notable memories, standing relationships, debts owed and owing, and notable deeds
   — every field derived from recorded history.
2. **Given** a player who has edited the angel's charter and placed standing orders,
   **When** a villager dies, **Then** the epitaph records which charter revision was
   in force and which orders were active (and their watch subjects) at the moment of
   death — stated as evidence, with no scoring or blame language.
3. **Given** a narrator model is available, **When** an epitaph is written, **Then** a
   prose epilogue is appended after the factual record, visibly separated from it,
   and removing the narrator changes no factual content.
4. **Given** the last villager dies, **When** the run ends, **Then** the morgue gains
   a run-end summary: run length in days, the population curve, each death with its
   cause, and the run's notable events.
5. **Given** an ended world, **When** the same history is replayed, **Then** the
   morgue's factual content is reproduced identically.

---

### User Story 3 - The gru can finish the wounded (Priority: P3)

Nights get real stakes: a healthy villager still survives any single gru encounter,
but a villager already weakened — wounded and unhealed, starving, freezing — can be
killed by the gru. Deaths therefore trace back to preventable spirals (no fire, no
food, wounds untreated, nobody watching) rather than to one unlucky roll, which keeps
every death attributable to something the player's instructions could have addressed.

**Why this priority**: Without reachable death beyond pure neglect, run ends almost
never occur and the morgue stays empty — but this story deliberately rides on US1/US2
existing first, so a death that does occur lands in a defined run/record structure.

**Independent Test**: On a seeded world, arrange one healthy villager and one
weakened villager to meet the gru; verify the healthy one survives with the floor
intact and the weakened one can die, deterministically reproducible under replay.

**Acceptance Scenarios**:

1. **Given** a healthy villager, **When** the gru attacks, **Then** the villager is
   wounded but never killed by that single attack (the survival floor holds).
2. **Given** a villager already below the near-death band, **When** the gru attacks,
   **Then** the villager can die; the death carries a gru-attributed cause and flows
   through the normal death path (witnesses, belongings spilling, chronicle,
   morgue epitaph).
3. **Given** the same seed and history, **When** the run is replayed, **Then** the
   same attacks produce the same outcomes — no new nondeterminism.
4. **Given** the gru's existing safety rules (light aversion, shelter protection),
   **When** escalation lands, **Then** those rules are unchanged — only the survival
   floor's condition changed.

---

### User Story 4 - Graves on the map, grief in the village (Priority: P4)

Where a villager fell, a grave appears — visible on the map to the player, and real
to the villagers: those who saw the death or later pass the spot remember the place
as a grave, and grief becomes something villagers talk about, entering the village's
rumor life. The dead stay part of the village's story and geography.

**Why this priority**: Atmosphere and social-material payoff (pre-session decision 3:
deaths generate social material). It enriches the other stories but nothing else
depends on it.

**Independent Test**: Cause a witnessed death; verify a grave marker exists at the
spot on the map, that the witness carries a memory of the grave's place, and that
grief-related talk about the death appears in village conversations/rumors within a
game-day.

**Acceptance Scenarios**:

1. **Given** a villager dies at a location, **When** the death lands, **Then** a
   grave marker exists at that location, persists, and renders on the player's map
   and in event summaries.
2. **Given** a villager witnesses the death or later passes the grave, **Then** their
   private map knowledge gains the grave as a known place with provenance, usable in
   their situated context like other known places.
3. **Given** a witnessed death, **Then** grief-flavored material about it enters the
   rumor/conversation fabric within a game-day, riding existing social systems.
4. **Given** graves exist, **Then** no mourning morale penalty and no grave-visiting
   behavior is introduced (explicitly out of scope, deferred).

---

### Edge Cases

- Two villagers die on the same tick, one of them last: exactly one run-end record
  fires, after all same-tick deaths are recorded, and the morgue receives every
  epitaph plus the summary in a stable order.
- The last villager dies while the world is paused or mid-speed-change: the run end
  lands exactly once and the postmortem posture is entered regardless of clock state.
- The narrator is configured but unavailable (provider down, budget exhausted) at
  epitaph time: the factual record is written immediately; the epilogue is skipped
  without blocking or corrupting the morgue — silence, not failure.
- The morgue file has been deleted or hand-edited between deaths: recorded history
  remains the source of truth; the factual content must be regenerable from the
  event history alone.
- A replay or migration of an ended world: the ended state is reproduced from
  history; migration tooling must not resurrect a finished run.
- The gru wounds a weakened villager who is inside shelter or beside a fire: existing
  protection rules still apply before any lethality question arises.
- A grave's location coincides with an existing structure or later construction:
  the grave persists and remains addressable; building over it is a world-design
  question deferred with the rest of grave interactions.
- An attached client at the moment of run end: the transition is pushed live; the
  client lands in postmortem posture without needing a reconnect.

## Requirements *(mandatory)*

### Functional Requirements

**Run outcomes**

- **FR-001**: The system MUST declare a run ended, exactly once, when the last living
  villager dies, recording it in the world's durable history with summary facts (at
  minimum: final tick/day, total deaths with causes, and the final death's cause).
- **FR-002**: After a run ends, simulated time MUST stop advancing and MUST NOT be
  resumable by any player, angel, or operator control short of tooling explicitly
  designed to fork/migrate worlds.
- **FR-003**: An ended world MUST remain fully readable: status, chronicle/story
  feed, event history, morgue, and the full-screen client all function, and the
  client MUST clearly present the world as over (postmortem posture).
- **FR-004**: Stopping and restarting an ended world's daemon MUST return it to the
  same ended, readable state.
- **FR-005**: The run-end fact MUST be exposed in machine-readable form (durable
  event and status surface) such that future scenario machinery (TASK-119) can
  consume it as a fail signal without new plumbing.
- **FR-006**: Ending a run MUST NOT modify the "one directory = one run, no in-place
  reset" semantics: starting over means a new world directory, and ended worlds
  remain as archives.

**The morgue file**

- **FR-007**: Each villager death MUST append a factual epitaph to a single
  accumulating morgue document in the world's save directory, containing at minimum:
  name, days survived, cause of death, notable memories, standing relationships,
  debts owed and owing, and notable deeds — all derived from recorded history, with
  no AI model required.
- **FR-008**: Each epitaph MUST record the angel-policy evidence in force at the
  moment of death: the charter revision identity and the active standing orders.
  The presentation MUST be evidential (what was instructed, what happened) and MUST
  NOT include scoring, grading, or blame language.
- **FR-009**: Run end MUST append a village-level summary to the morgue: run length,
  population over time, every death with cause, and the run's notable events.
- **FR-010**: When a narrator model is available, the system MUST append a prose
  epilogue after an epitaph's factual record, clearly separated from it; narrator
  absence or failure MUST NOT block, delay indefinitely, or alter the factual record.
- **FR-011**: The morgue's factual content MUST be a pure projection of recorded
  history: replaying the same history reproduces it identically, and it is
  regenerable from the event history alone if the file is lost.
- **FR-012**: The morgue document's structure MUST be stable and self-contained
  enough to serve as the source for a future shareable export (separate task) without
  reformatting: consistent per-death sections, a run-summary section, and clearly
  delimited narrated passages.

**Death escalation**

- **FR-013**: A gru attack on a villager in the healthy band MUST continue to wound
  and never kill (the existing survival floor holds).
- **FR-014**: A gru attack on a villager already weakened below the near-death band
  MUST be able to kill, deterministically as a function of recorded state — never a
  fresh random roll outside the world's deterministic decision discipline.
- **FR-015**: Escalated deaths MUST flow through the existing death path unchanged:
  cause attribution, witness memory formation, belongings spilling (per the shipped
  death-drop behavior), chronicle narration, and the new morgue epitaph.
- **FR-016**: The gru's existing behavioral rules (nocturnal emergence, light
  aversion, shelter protection) MUST be unchanged by this feature.

**Graves**

- **FR-017**: A villager death MUST create a persistent grave marker at the death
  location, visible on the player's map and named in event summaries.
- **FR-018**: Villagers who witness the death or later perceive the grave MUST gain
  it as known-place knowledge in their private spatial memory with normal provenance
  and freshness semantics.
- **FR-019**: A witnessed death MUST produce grief material in the social fabric
  (memories that can seed rumors/conversation) using existing mechanisms; no new
  behavioral drives (mourning morale effects, grave visiting) may be introduced.

**Cross-cutting**

- **FR-020**: All new facts (run end, graves, morgue facts, escalated deaths) MUST be
  replay-deterministic: a replay of the same history reproduces the same world state
  and the same factual artifacts.

### Key Entities

- **Run end**: the durable declaration that a world's run is over — when it happened,
  why (all villagers dead), and the run's summary facts. Consumed by the postmortem
  reading surfaces and, later, scenario machinery.
- **Morgue document**: the single accumulating legacy document of a run — an ordered
  series of per-death epitaphs (facts + angel-policy evidence + optional narrated
  epilogue) closed by a run-end summary. Lives in the world save directory beside
  the run's other artifacts.
- **Epitaph**: one death's factual record: identity, span, cause, memories,
  relationships, debts, deeds, and the charter/orders evidence in force at death.
- **Grave**: a persistent world feature at a death site, knowable by villagers as a
  place and visible to the player on the map.
- **Angel-policy evidence**: the identification of which charter revision and which
  standing orders were active at a given moment — the alignment substrate that makes
  failure attributable to the player's authored text.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: When the last villager dies, the run is declared over within one
  simulated second, and every reading surface (status, story feed, full-screen
  client, morgue) remains usable afterward — including after a daemon restart.
- **SC-002**: After any death in a world with no AI configured, the morgue contains a
  complete epitaph (all seven factual fields present) that matches the recorded
  history it summarizes.
- **SC-003**: For every death, a reader can determine from the morgue alone what the
  angel was instructed (charter revision + active orders) when it happened.
- **SC-004**: Replaying an ended world reproduces the identical run end and
  byte-identical morgue factual content (narrated epilogues excluded from the
  byte-identity requirement, facts not).
- **SC-005**: On a seeded demonstration world, a weakened villager can be killed by
  the gru while a healthy villager in the same circumstances survives with the floor
  intact — reproducibly across replays.
- **SC-006**: Within one game-day of a witnessed death, the grave is visible on the
  player's map, at least one villager holds it as a known place, and at least one
  grief-related rumor or conversation references the death.
- **SC-007**: A player reading only the morgue and chronicle after a lost run can
  answer "what killed the village, and what were my angel's instructions at each
  death" without consulting raw event data.

## Assumptions

- The existing near-death band (the health threshold already used for survival
  semantics) defines "already weakened" for escalation; the plan phase pins the
  exact predicate to current constants rather than inventing a new band.
- The narrated epilogue rides the existing narrator/chronicle model routing and
  budget discipline; it introduces no new model-call class.
- The morgue document lives beside the world's other save-directory artifacts and
  follows the "one directory = one run" doctrine already in force.
- Charter revision identity is derivable from the charter's existing hot-reload
  surface (the angel's policy files are re-read per turn today); if revisions are not
  yet individually identified, the plan phase introduces the minimal identification
  needed for evidence alignment (e.g., content fingerprint at effect time).
- The physical death-drop (carried items spilling at the death site) shipped with
  spec 013 and is consumed as-is.
- Scenario machinery (TASK-119) and curriculum stages (TASK-68) are future consumers
  of `run.ended` and preset difficulty respectively; nothing here blocks on them.
- The shareable HTML retelling export (Boatmurdered export) is a separate future
  task; this spec only guarantees the morgue's structure can feed it.
- World forking/what-if tooling (TASK-67) may later re-open ended histories by
  forking into new directories; that path is explicitly not "resuming an ended run."
