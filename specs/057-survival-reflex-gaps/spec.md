# Feature Specification: Survival Reflex Gaps (fire)

**Feature Branch**: `057-survival-reflex-gaps`

**Created**: 2026-07-25

**Status**: Draft

**Input**: TASK-108 — doctrine decision (2026-07-24): self-preservation is table
stakes; err toward survival instinct wherever the reflex layer has a gap.
World-01 evidence (control-surface report §3.1): 8 fires built vs 42 burnouts
over 6 days, warmth 848→82 by day 7, Oak died of exposure, 425 intents rejected
"no warmth anywhere". Scope decided on the task: raise the refuel threshold to
~3 game-hours; ensure the cold build-fire reflex; leave fireBurnPerWood alone
(one lever at a time); thresholds ride the tuning manifest (spec 048, shipped).

**Reality check (2026-07-25, sweep grounding)**: the task's code pins predate
specs 041/043/048. The reflex ladder's night branch TODAY already contains a
build-fire rung (cold night + no reachable *known* warmth + wood ≥ 2 → build
fire) and the refuel threshold is already a tuning dial (`refuel_dying_below`,
default 3600). This spec therefore (a) changes the doctrine default, (b) proves
the existing cold reflex against the AC with tests and closes any real gap the
proof exposes, and (c) audits the rest of the reflex ladder against the
survival doctrine — it does not re-build what 041 built.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Fires get refueled long before they die (Priority: P1)

The refuel reflex arms when a known fire has less than **3 game-hours** of fuel
remaining (today: 1 game-hour). Villagers top up fires early instead of racing
burnout, cutting the burnout count that killed world-01's warmth economy.

**Why this priority**: the single evidenced lever — 42 burnouts vs 8 builds
means fires died faster than the 1-hour window let villagers react.

**Independent Test**: a fire with 2.5 game-hours of fuel and a wood-carrying
villager nearby → refuel intent fires under the new default; under the old
default it would not.

**Acceptance Scenarios**:

1. **Given** a known fire with fuel below 3 game-hours and a villager carrying
   wood, **When** the reflex evaluates, **Then** it produces a refuel intent
   (no model call).
2. **Given** a world with `tuning.json` setting `refuel_dying_below`
   explicitly, **When** it boots, **Then** the manifest value wins over the new
   default (dial semantics unchanged).

---

### User Story 2 - Doctrine-default changes don't rewrite history (Priority: P1)

Changing a promoted dial's default (as US1 does — the first default change
since spec 048 shipped) must not silently change the replay of existing
worlds. New worlds pin their birth doctrine: world creation seeds a
`sim.tuning_applied` event carrying the full default set at genesis, so a
world's effective doctrine is fixed in its own log and future default changes
never reach back into it. Worlds created before this feature carry no pin; the
US1 default change reaches them, and that hazard is documented where the
determinism scope note lives (the TASK-75 hazard class).

**Why this priority**: co-P1 because shipping US1 without it silently shifts
replay for every un-tuned world; with it, this is the last default change that
ever has that power over new worlds.

**Independent Test**: create a world, replay its log with a binary whose
compiled defaults differ → identical behavior (values come from the genesis
pin); a pre-057 world's log has no pin and follows compiled defaults.

**Acceptance Scenarios**:

1. **Given** `promptworld new`, **When** the world is created, **Then** its log
   contains one `sim.tuning_applied` event at genesis with the full current
   default set.
2. **Given** a post-057 world and a later change to any `default*` constant,
   **When** its log is replayed, **Then** behavior matches the original run
   (the genesis pin wins, not the compiled default).
3. **Given** a `tuning.json` on a post-057 world, **When** the daemon boots,
   **Then** the boot seed compares against the pinned set and appends a new
   event only on difference (spec 048 semantics preserved).

---

### User Story 3 - The cold build-fire reflex provably holds (Priority: P2)

The AC from the task — "cold villager with wood and no reachable warmth builds
a fire via reflex, no model call" — is proven by tests against the CURRENT
ladder, and any gap the proof exposes is closed. Known candidate gaps to probe
(from reading the ladder, to verify rather than assume): a villager carrying
1 wood + chop unavailable; a villager whose known-warmth belief is stale (fire
it knows is actually out); the wake/sleep boundary (does a sleeping villager
wake to cold?).

**Why this priority**: the rung exists (spec 041); what's missing is proof
against the doctrine and the survival-gap sweep. P2 because US1/US2 are the
behavior changes; this is verification plus targeted patching.

**Independent Test**: table-driven reflex tests: cold night × {wood≥2, wood=1,
wood=0} × {known warmth reachable, none known, known-but-dead} → the doctrine
outcome for each cell, each asserted with no planner involvement.

**Acceptance Scenarios**:

1. **Given** a cold night, no reachable known warmth, and ≥2 wood carried,
   **When** the reflex evaluates, **Then** a build_fire intent results.
2. **Given** the same but insufficient wood and a choppable tree known,
   **When** the reflex evaluates, **Then** a chop intent results (the ladder's
   existing wood-acquisition rung, proven not regressed).
3. **Given** any gap found by the proof matrix, **When** the fix lands, **Then**
   the matrix cell's doctrine outcome passes and the fix is noted on the task
   (AC3's audit trail).

---

### User Story 4 - Reflex-layer survival audit (Priority: P3)

A written audit of the reflex ladder against the survival doctrine (eat, sleep,
warmth parity): each rung, the need it protects, the thresholds it keys on, and
any gap where a villager can die without the reflex ever acting. Gaps found are
FIXED here only if they are fire-adjacent and surgical; everything else is
carded to the board (this spec's boundary: fire; TASK-103 owns arbitration,
TASK-104 owns recovery semantics).

**Why this priority**: AC3 asks for the audit note; it seeds the next survival
tasks with evidence instead of anecdotes.

**Acceptance Scenarios**:

1. **Given** the shipped feature, **When** the audit doc is read, **Then** every
   reflex rung appears with its need, thresholds, and gap disposition (fixed
   here / carded / no gap).

---

### Edge Cases

- Refuel threshold above fuel-per-wood (10800 < 14400 today, but a tuned world
  could invert them): refuel intent still valid — the executor's fuel-cap
  truncation already bounds over-fueling.
- Genesis pin vs migrated worlds: `promptworld migrate` output (v3→v4) carries
  no pin — same class as pre-057 worlds; documented, not back-filled.
- Genesis pin and `tuning.json` present at first boot: pin (genesis, defaults)
  applies first, manifest seed compares against it — one additional event iff
  the manifest differs.
- Old snapshots/logs (no pin, no tuning event): replay under compiled defaults,
  exactly as today — including the US1 shift for pre-057 worlds, documented.
- A future default change with no other release note: the genesis pin makes it
  invisible to existing post-057 worlds — new worlds only. That is the intended
  semantic (a world's doctrine is fixed at birth unless tuned).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The refuel-reflex default threshold MUST become 10800 game-seconds
  (3 game-hours). The `refuel_dying_below` dial's clamp range, manifest
  semantics, and event discipline are unchanged; a manifest value still wins.
- **FR-002**: World creation MUST seed one `sim.tuning_applied` event at genesis
  carrying the full effective default set, so post-057 worlds replay
  independently of later compiled-default changes.
- **FR-003**: The spec-048 boot seed MUST keep its exact semantics against the
  genesis pin: no event on equality, one full-set event on difference.
- **FR-004**: The cold build-fire reflex AC MUST be proven by a test matrix over
  {wood level} × {warmth knowledge}, with every cell's doctrine outcome
  asserted reflex-only (no planner); gaps exposed by the matrix MUST be either
  fixed (fire-adjacent, surgical) or carded with evidence.
- **FR-005**: A reflex-ladder survival audit MUST be produced as a durable
  artifact (in the spec dir or docs/), covering eat/sleep/warmth parity with
  per-rung disposition.
- **FR-006**: All documentation pinned to the old default MUST move in the same
  change: spec 048's contract table (`specs/048-tuning-manifest/contracts/tuning.md`),
  the control-surface report §2.4/§6 rows, and the wiki `world-tuning` note
  (re-pin via the wiki flow post-merge).
- **FR-007**: Pre-057 and migrated worlds (no genesis pin) MUST keep loading and
  replaying; the default-shift hazard for them MUST be documented in the same
  place the determinism scope note lives.

### Key Entities

- **Genesis tuning pin**: the `sim.tuning_applied` event seeded at world
  creation; payload identical in shape to the boot-seeded event (spec 048).
- **Refuel default**: `defaultRefuelDyingBelow` 3600 → 10800 in the sim doctrine
  home (`internal/sim/tuning.go`).
- **Reflex proof matrix**: the test table proving US3's cells.
- **Survival audit**: the written per-rung audit artifact (US4).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Under the new default, a fire with 2.5 game-hours of fuel gets a
  refuel intent from a wood-carrying villager; under a manifest pinning the old
  value it does not — both proven by tests.
- **SC-002**: A post-057 world's replay is byte/hash-identical after a compiled
  default is changed out from under it (test flips a default and replays).
- **SC-003**: Every cell of the US3 proof matrix passes reflex-only; zero model
  calls involved in any asserted path.
- **SC-004**: The audit artifact exists with 100% rung coverage and explicit
  disposition per gap; anything carded has a board task id in the artifact.
- **SC-005**: Full existing suite stays green; no format_version bump; pre-057
  snapshot bytes unchanged (048's compat guarantees intact).

## Assumptions

- 10800 (not another value) per the task's "~3h"; it stays inside the dial's
  existing clamp [0, 86400]; world-01 currently has no tuning.json, so after
  this merge world-01 inherits the new default at next boot (it is a pre-057
  world; that is the intended live effect, matching the task's motivation).
- Genesis pinning is scoped to `promptworld new`; `migrate` does NOT back-fill
  pins (rewriting a migrated world's history head is worse than the documented
  hazard).
- The day-branch warmth gap named in TASK-103 ("policy.go day branch never
  considers warmth") is TASK-103's scope, not this spec's — the audit may cite
  it, the code here must not preempt the arbitration design.
- fireBurnPerWood untouched (task decision: one lever at a time).
