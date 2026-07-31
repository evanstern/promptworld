# Feature Specification: Guardian missions — accept, decompose, pursue, report

**Feature Branch**: `task-158-guardian-missions`

**Created**: 2026-07-30

**Status**: Draft

**Input**: TASK-158 — the player-facing loop on top of the TASK-157 substrate:
"Guardian, direct the villagers to settle near (x,y)" becomes a durable,
event-sourced MISSION artifact. Depends: TASK-112 (Done — steward scheduled
cognition, PR #146), TASK-157 (Done — designations/directives + observable
directive.* events). Operator decisions encoded: EASY-MODE default (2026-07-26,
firm) — the default guardian executes the player's missions without
editorializing; refusals/personality are skinned-charter data only. Eval gate
ruling (2026-07-30, in-session): the default-charter obedience edit is gated by
an IN-BRANCH obedience eval (TASK-73 precedent).

## Doctrine (from the card — not re-litigated)

A mission is durable pre-authorization — the same legal shape as a standing
order — so NO initiative-frame relaxation is needed: pursuit turns are
pre-authorized by the mission the player issued. Crucially this composes with
spec 102's incompetence ceiling: the ceiling caps INITIATIVE, and a mission is
not initiative — it is the player's explicit standing instruction, so mission
pursuit runs at FULL competence at any ceiling (exactly like console orders).

## Design decisions

- **D1 — Mission as a spec-084-discipline artifact.** `guardian.mission_*`
  events (accepted / progressed / completed / failed / cancelled) with
  deterministic IDs, one-way terminal status, prune discipline — the same
  entity shape as designations/directives/regions.
- **D2 — Decomposition through existing verbs only.** The guardian pursues a
  mission by surveying, placing designations, issuing directives, and (grant
  permitting) working miracles — all EXISTING tools through EXISTING doors. No
  new world-acting verbs; the mission artifact only records intent and derived
  progress.
- **D3 — Completion is derived, never self-graded.** Mission completion/failure
  derives from the designation/directive fulfillment predicates (spec 084) and
  recorded events; the report-card producer cites event evidence, never model
  prose.
- **D4 — Pursuit rides spec 102's scheduled lane.** Multi-turn follow-through
  uses steward cognition (mission context joins the scheduled turn's prompt);
  an active mission's pursuit tools are granted at full competence regardless
  of ceiling (doctrine above), scoped to the mission's needs. Event-driven
  triggers unchanged.
- **D5 — EASY-mode default charter, eval-gated.** The default compiled charter
  gains the obedience clause (execute missions without editorializing; counsel
  only when a mission is IMPOSSIBLE as stated — refusal names the blocking
  fact). The in-branch obedience eval (FR-008) proves the edit: old default
  reproduces the TASK-166-observed counsel-loop; new default executes directly.
  Personality/refusal-first behavior remains skinned-charter data (TASK-121).

## User Scenarios & Testing *(mandatory)*

### US1 - A standing order in plain words (Priority: P1)

As a player, I want to say "Guardian, get a second fire built near the west
huts and keep it fueled" once — and have that become a durable mission the
guardian decomposes and pursues across turns without me in the loop, reporting
back when it's done or honestly stuck.

**Acceptance Scenarios**:

1. **Given** a player mission statement in guardian chat, **When** accepted,
   **Then** guardian.mission_accepted lands (D1) with the goal expressed in
   completion-predicate vocabulary where checkable (D3).
2. **Given** an accepted mission, **When** scheduled turns run, **Then** the
   guardian pursues via existing verbs (D2) with mission context in the
   prompt; pursuit runs at full competence at any ceiling.
3. **Given** fulfillment predicates satisfied, **Then**
   guardian.mission_completed derives from recorded events; the report card
   cites the evidence trail.
4. **Given** a stalled mission (predicates unmet, pursuit blocked), **Then**
   guardian.mission_failed lands with recorded-event evidence of the blocker —
   honest failure, no prose grading (card AC#4).

### US2 - The default guardian obeys (Priority: P1)

1. **Given** the NEW default charter, **When** the player issues a direct
   mission, **Then** acceptance + pursuit begin that same turn — no
   counsel-loop (the TASK-166-observed 4-turn counseling is the before-picture
   the eval reproduces on the OLD default).
2. **Given** a skinned charter with refusal/personality clauses, **Then**
   counsel-first behavior arrives via that data — demonstrably absent from the
   default (card AC#5).

### US3 - Charter quality still changes outcomes (Priority: P2)

1. **Given** the anti-self-grading guard (card AC#6), **Then** the TASK-164
   instrument extended with a mission scenario measures default-vs-authored
   mission outcomes on seeded worlds (harness prepared in-branch; the run may
   land as the instrument's next scheduled pass post-merge).

### Edge Cases

- Missions never replace standing orders; the order door remains the single
  arbiter (spec 102 D6) — a mission may CAUSE orders through the normal door.
- Player cancel ⇒ guardian.mission_cancelled (one-way terminal); pursuit stops
  next scheduled turn.
- Impossible-as-stated missions: refusal at acceptance names the blocking fact
  — the ONE sanctioned counsel case (D5).
- Replay/versioning: additive guardian.mission_* types (spec 094, no format
  bump); pre-107 worlds byte-identical. Worlds without the steward lane
  (steward_cadence_ticks=0) accept missions but pursue only on event-driven
  turns — degraded-but-honest, documented.
- The playtest world is never touched; demos on seeded measure worlds.

## Requirements *(mandatory)*

- **FR-001**: Mission artifacts per D1 on sim.State; door-validated
  acceptance; one-way terminals; prune discipline.
- **FR-002**: Pursuit via existing verbs only (D2); mission context in
  scheduled prompts; full-competence pursuit at any ceiling, scoped to the
  active mission.
- **FR-003**: Completion/failure derived from predicates + recorded events
  (D3); report-card integration cites evidence.
- **FR-004**: EASY-mode default charter clause (D5); skinned refusals
  demonstrably separate.
- **FR-005**: Digest + event-types.md + decision-trail; TestCatalogSweep green.
- **FR-006**: Tests: acceptance/refusal, artifact discipline, derived
  completion (satisfied + stalled), ceiling-composition (default ceiling +
  active mission ⇒ pursuit granted), replay byte-identity.
- **FR-007**: Live demo on a seeded measure world: mission accepted →
  decomposed → pursued across ≥2 scheduled turns, no player in loop →
  completed with evidence; docs/design/evidence/task-158/.
- **FR-008**: In-branch obedience eval (operator ruling 2026-07-30): scripted
  direct-mission prompts, old vs new default charter, guardian route via the
  measurement proxy (small spend); pass = new default executes without
  counseling AND old default's counsel-loop reproduced; results in the
  evidence doc + PR body.

## Success Criteria *(mandatory)*

- **SC-001**: Full loop live (US1 end-to-end on a measure world).
- **SC-002**: Obedience eval passes both directions.
- **SC-003**: All spec 102 guardrail/ceiling tests still green (missions
  compose with, never bypass, the ceiling and the order door).
- **SC-004**: Replay fixtures byte-identical; additive types only.

## Assumptions

- Tier: **Opus** — initiative-frame doctrine, cross-package (the draft
  runbook's A5 tier call carried over).
- Sibling-sweep tasks 172/175 may merge mid-implementation (internal/mind) —
  routine merge-in reconciles.
