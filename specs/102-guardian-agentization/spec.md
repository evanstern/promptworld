# Feature Specification: Guardian agentization — the angel as a first-class autonomous agent

**Feature Branch**: `task-112-guardian-agentization`

**Created**: 2026-07-30

**Status**: Draft

**Input**: TASK-112 (direction firm, operator 2026-07-24: the angel becomes a
SINGULAR AGENT akin to the villagers — same basic agent construct, different
tool roster, extra context, god mode on) + the learning-game reframe (agentization
is "the player programs an agent", curriculum stage 3; three-lane initiative
frame) + OPERATOR RULING 2026-07-30 (in-session checkpoint): the
**deliberate-incompetence ceiling is ADOPTED** — world-acting only, never tutor
facts. Depends: TASK-111 (Done — order machinery must survive), TASK-157 (Done),
TASK-163/166/167 (Done — tool competence floor), TASK-164 (evidence pending —
dispatch gate, baseline for AC5).

## Non-negotiable guardrails (carry over UNCHANGED — card text)

No invented events; player words never reach villagers; no free miracles; no
villager removal; reducer-side charge economy. Every one demonstrably intact
post-redesign (AC#4): the existing guardrail test suites must pass unmodified
(or with mechanical-only updates), and the spec's own review checklist walks
each guardrail to its enforcing test.

## Design decisions

- **D1 — Same construct, different dress.** The guardian runs the villager
  agent shape: a mind loop with scheduled cognition, working memory + a real
  memory model (soul.md's append-only digests become structured memory with
  consolidation), a persona-equivalent (charter + skills + capabilities files
  compile into the persona surface exactly where a villager's soul/persona
  files sit), and a cadence. Its tool roster stays the guardian roster
  (stage/grant-gated, god mode); its context gains the replica digests it
  already owns.
- **D2 — An angel decision-class in the cognition registry.** Scheduled
  guardian cognition is a registry class with points/staleness budget — the
  governor and cognition horizon gate it like every villager class; under
  saturation the angel sheds before villager survival classes. Player-chat
  turns and standing-order matches remain event-driven doors (unchanged);
  agentization ADDS the scheduled lane, it does not replace the triggers.
- **D3 — The deliberate-incompetence ceiling (operator-adopted).** The DEFAULT
  compiled charter caps world-acting initiative: conservative watch thresholds,
  single-step actions only (no autonomous multi-step plans), no unprompted
  miracle spending beyond the TASK-111 genesis watches. A player-authored
  charter lifts the ceiling (multi-step initiative, proactive designations/
  directives, discretionary spending) — competence is bought with authorship.
  The ceiling is charter-data (compiled defaults), NEVER model degradation, and
  NEVER applies to the tutor voice: explain/converse quality is always full.
- **D4 — Channel split is doctrine (AC#6).** The tutor voice (converse +
  explain tool) spends no charges, lands no world events, earns no faith, and
  is excluded from every rubric; world-acting is the graded artifact. The
  split is structural: tutor-channel outputs cannot reach any world door.
- **D5 — Memory and consolidation are the guardian's own.** Guardian memories
  (survey results, mission outcomes, player exchanges, watch fires) enter a
  per-guardian store consolidated on the nightly boundary through the same
  machinery villagers use (spec 098's dream phase INCLUDED — single-store
  privacy holds trivially); soul.md remains the persona seed, not the memory
  log.
- **D6 — TASK-111 order machinery survives verbatim.** Genesis watch orders,
  standing-order triggering, and the report-card path keep their contracts;
  the scheduled lane may REVIEW watch state but never duplicates or races the
  order door (one arbiter: the existing order machinery).

## User Scenarios & Testing *(mandatory)*

### US1 - The angel thinks on its own schedule (Priority: P1)

As a player, I want the guardian to observe and act between my messages — on
its own cadence, budgeted like everyone else — so the world has a resident
caretaker rather than a vending machine.

**Acceptance Scenarios**:

1. **Given** a running world with no player input, **When** the angel class's
   cadence fires within budget, **Then** a guardian cognition turn runs
   (observe replica → optionally act through EXISTING doors), visible in the
   decision trail (AC#2).
2. **Given** governor saturation, **Then** the angel class sheds before
   villager survival classes; **Given** the horizon, **Then** angel turns are
   staleness-gated like any class.

### US2 - Default is dutiful-but-modest; charter buys brilliance (Priority: P1)

1. **Given** the default charter, **When** the angel acts autonomously,
   **Then** only ceiling-permitted actions occur (single-step, conservative
   thresholds) — test-enforced against the compiled default (D3).
2. **Given** an authored charter granting initiative, **Then** multi-step
   autonomous pursuit is permitted (the TASK-158 mission substrate consumes
   this) — and AC#5's anti-self-grading guard measures the outcome delta
   (TASK-164's design is the measurement instrument).

### US3 - The tutor never sends the bill (Priority: P1)

1. **Given** any tutor-channel exchange, **Then** zero charges, zero world
   events, zero faith, zero rubric contribution — structurally (D4), with a
   test proving tutor outputs cannot reach world doors.

### US4 - Same construct, provably (Priority: P2)

1. **Given** the implementation, **Then** the guardian's mind loop, memory
   store, consolidation, and persona compilation are the SHARED agent
   machinery (code-level reuse, not a parallel copy) — reviewed against a
   named reuse checklist in the plan (AC#3).

### Edge Cases

- Replay: scheduled cognition emits only through existing doors; any new
  telemetry types are additive (spec 094: no format bump). Existing worlds
  replay byte-identically; agentization activates via world/tuning config so
  pre-102 worlds are unchanged until opted in.
- The playtest world is NEVER touched; acceptance runs on seeded measure
  worlds.
- Budget safety: the angel class's default budget must not measurably degrade
  villager cognition latency on the reference 8-villager world (soak-checked).
- TASK-121 skins: persona compilation honors the skin layer (de-themed
  guardian data), no fiction re-hardcoded.

## Requirements *(mandatory)*

- **FR-001**: Guardian mind loop on the shared agent construct (D1) with an
  angel cognition-registry class (D2): cadence, points/staleness budget,
  governor/horizon gating; event-driven triggers unchanged.
- **FR-002**: Guardian memory store + nightly consolidation via shared
  machinery incl. spec 098 (D5); soul.md = persona seed only.
- **FR-003**: Charter/skills/capabilities compile into the persona surface;
  the DEFAULT compilation enforces the incompetence ceiling (D3) as data;
  authored charters lift it. Ceiling semantics test-enforced.
- **FR-004**: Channel split structural (D4): tutor path physically cannot
  invoke world doors, spend charges, or earn faith; rubric exclusion pinned.
- **FR-005**: All five guardrails demonstrably intact (AC#4) — enforcement
  tests enumerated and green; TASK-111 order machinery contracts unchanged
  (D6).
- **FR-006**: Anti-self-grading (AC#5): the charter-delta instrument
  (TASK-137/164 recipe) runs on the agentized guardian — default vs authored
  arms show a measurable behavior delta; outcome delta recorded (evidence
  doc).
- **FR-007**: Config/compat: agentization is opt-in per world (tuning/config);
  pre-102 worlds unchanged; replay byte-identity for existing logs.
- **FR-008**: Surfaces: decision-trail visibility for scheduled turns; digest
  entries for any new telemetry; docs (event-types, wiki notes incl. a new
  guardian-agentization note).

## Success Criteria *(mandatory)*

- **SC-001**: On a seeded measure world with no player input, the agentized
  guardian observes and acts within budget across a multi-day run; villager
  cognition latency unchanged within tolerance.
- **SC-002**: Ceiling proof: default-charter world shows ONLY ceiling-permitted
  autonomous actions across the run; authored-charter world shows lifted
  behavior (the AC#5 delta).
- **SC-003**: Guardrail suite green + tutor-channel structural-isolation test
  green.
- **SC-004**: Code-reuse review: the named agent-construct components are
  shared, not forked (plan checklist).

## Assumptions

- Tier: **Opus** (card-stated: cross-package metatron/mind/cognition/sim,
  doctrine-adjacent). Dispatch gated on TASK-164's evidence per the inherited
  checkpoint (runbook).
- TASK-158 (missions) builds directly on this; its EASY-mode default-charter
  obedience is NOT this spec's ceiling — obedience-to-direct-orders is full
  competence at any ceiling (the ceiling caps INITIATIVE, not compliance).
