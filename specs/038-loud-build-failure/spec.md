# Feature Specification: Loud Build Failure & Occupancy Tolerance

**Feature Branch**: `038-loud-build-failure`

**Created**: 2026-07-24

**Status**: Draft

**Input**: User description: "Loud build cancellation and occupancy tolerance for wall builds (fixes TASK-91 phantom-wall belief loop). Today a build intent cancelled by mid-work re-validation (site vanished, or another agent standing on the reserved tile) resolves as a bare agent.intent_done — identical to success — so no material is spent, no wall appears, the builder gets no memory that the build failed, and conversation gists/chronicle entries about the wall stand uncorrected forever; agents plan against a phantom wall."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A failed build is loudly distinguishable from a finished one (Priority: P1)

A villager starts building a wall. Partway through the work, the site becomes
invalid (e.g. the spot is no longer buildable). Instead of the build silently
resolving exactly like a success, the world records a distinct build-failure
event, and an observer watching the event stream or the TUI digest can tell at
a glance that the build failed and why.

**Why this priority**: This is the root of the phantom-wall belief loop — a
cancelled build currently emits the same event as a completed one, so nothing
downstream (observers, tooling, the villagers themselves) can ever learn the
build didn't happen. Every other part of the fix hangs off this signal.

**Independent Test**: In a controlled simulation, invalidate a wall build's
site mid-work and assert the event log contains a distinct failure event
(with the builder, the goal, and a reason) instead of a bare completion event,
and that the TUI digest renders it as a failure, not "finished".

**Acceptance Scenarios**:

1. **Given** a villager mid-way through a wall build, **When** the build site
   becomes invalid before the work completes, **Then** the world emits a
   distinguishable build-failure event carrying the builder, the attempted
   goal, and a human-readable reason — not a bare completion event.
2. **Given** a build-failure event in the event stream, **When** the TUI
   digest renders it, **Then** the line reads as a failure (what failed and
   why), visibly distinct from "finished".
3. **Given** the documented event catalog, **When** a reader looks up the new
   event type, **Then** its payload and emission conditions are documented.

---

### User Story 2 - The builder remembers the failure (Priority: P1)

The villager whose build was cancelled walks away knowing the build did NOT
complete and why. Their next planning cycle, and any later conversation about
the wall, is grounded in a memory that contradicts the phantom-wall belief —
so beliefs formed from earlier optimistic chatter ("the walls stand") are
falsifiable against lived experience.

**Why this priority**: Even with a distinct event, the villagers' minds only
know what their memories tell them. Without a failure memory the belief loop
persists: gist memories from conversations say the wall exists, and nothing
the builder experienced says otherwise.

**Independent Test**: Cancel a build mid-work in a controlled simulation and
assert the builder gains a first-person memory (origin: their own action)
stating the build did not complete and why.

**Acceptance Scenarios**:

1. **Given** a villager whose build fails mid-work, **When** the failure
   resolves, **Then** that villager receives a situated first-person memory
   stating the build did NOT complete, naming what they were building and the
   reason it failed.
2. **Given** the failure memory exists, **When** the villager later plans or
   converses about that structure, **Then** the memory is available as
   contradicting evidence (same retrieval path as other action memories).

---

### User Story 3 - A passerby no longer kills a wall build (Priority: P2)

A villager is building a wall while friends socialize nearby. A friend walks
across the wall's reserved tile mid-build. The build is not cancelled: work
continues, and when the work is done, if someone is standing on the tile the
wall's completion simply waits until the tile is clear (nobody is ever
entombed in a wall). Only if the tile stays occupied for an unreasonably long
time does the build give up — loudly, via the failure event and memory above.

**Why this priority**: This removes the trigger that made wall builds nearly
impossible in social settings. It's P2 because even with tolerance, failures
can still happen (site vanished, permanent squatter) and those must be loud
first — loudness (P1) is what breaks the belief loop; tolerance reduces how
often failures occur at all.

**Independent Test**: In a controlled simulation, walk a second agent onto
the reserved tile mid-build, then off again; assert the wall completes. In a
second run, park the agent on the tile permanently; assert the build fails
loudly after the grace period rather than waiting forever.

**Acceptance Scenarios**:

1. **Given** a wall build in progress, **When** another agent steps onto the
   reserved tile during the work cycle and leaves before completion, **Then**
   the wall completes normally as if never interrupted.
2. **Given** a wall build whose work cycle has finished, **When** an agent is
   standing on the reserved tile at the completion moment, **Then** the wall
   does not materialize under them (never entomb) and completion is deferred
   until the tile clears.
3. **Given** completion deferred by an occupied tile, **When** the tile
   remains continuously occupied past a bounded grace period, **Then** the
   build fails loudly (failure event + failure memory), never waiting forever
   and never resolving silently.
4. **Given** a wall build in progress, **When** the site itself becomes
   invalid (independent of occupancy), **Then** the build fails loudly — site
   loss is a genuine failure, not something to wait out.

---

### Edge Cases

- Occupant steps off during the grace period: completion resumes the next
  tick the tile is clear; the wall appears and materials are spent as normal.
- The builder is the occupant: the builder works adjacent to the reserved
  tile by design; only other agents can occupy it. If the builder somehow
  stands on it, the same wait-then-fail rules apply (never entomb anyone,
  including the builder).
- Multiple agents take turns standing on the tile with no gap: the grace
  period bounds total continuous occupancy, regardless of who occupies it.
- Site becomes invalid while completion is deferred on occupancy: the site
  failure wins — fail loudly immediately.
- Replayed worlds: a world recorded after this change replays byte-identically
  (the new event and memory are part of the deterministic event stream).
- Non-wall builds (fire, shelter, oven, chest, path): their site-vanished
  mid-work failures become loud too; they have no reserved-tile occupancy
  guard, so the tolerance mechanics don't apply to them.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST emit a distinguishable build-failure event —
  distinct from the normal completion event — whenever a build intent that
  passed initial acceptance is cancelled by mid-work re-validation. The event
  MUST carry the builder's identity, the attempted build goal, and a
  human-readable reason for the failure.
- **FR-002**: The build-failure event MUST be documented in the project's
  event-type catalog (payload shape, emitter, state effect) and MUST be
  rendered distinctly (as a failure, with reason) in the TUI digest.
- **FR-003**: When a build fails mid-work, the builder MUST receive a
  situated first-person memory (origin: own action) stating that the build
  did NOT complete, what was being built, and why it failed — retrievable by
  the same mechanisms as other action memories.
- **FR-004**: Transient occupancy of a wall build's reserved tile MUST NOT
  cancel the build: during the work cycle, only site validity is checked;
  occupancy of the reserved tile is ignored until the completion moment.
- **FR-005**: At the completion moment, if the reserved tile is occupied, the
  wall MUST NOT materialize (no agent is ever enclosed in a wall) and
  completion MUST be deferred until the tile is clear.
- **FR-006**: Deferred completion MUST be bounded: if the reserved tile
  remains continuously occupied past a fixed grace period, the build MUST
  fail loudly per FR-001 and FR-003 rather than wait indefinitely.
- **FR-007**: Site invalidity (the build spot itself becoming unusable) MUST
  remain an immediate, loud failure for all build goals — never silently
  resolved, never waited out.
- **FR-008**: A failed build MUST NOT spend the build's material inputs and
  MUST NOT produce a structure (unchanged from today), and the builder MUST
  return to normal planning afterwards (the failure clears the intent just as
  completion does).
- **FR-009**: World replay MUST remain deterministic and byte-identical with
  the new event and memory in the stream.

### Key Entities

- **Build-failure event**: A new event type in the world's event stream;
  carries builder, goal, and reason; cleared-intent state effect identical to
  completion; observability surface for tools, tests, and the TUI.
- **Failure memory**: A first-person, action-origin memory owned by the
  builder; the falsifiable counter-evidence against phantom-structure
  beliefs.
- **Grace period**: A fixed bounded number of ticks that deferred completion
  may wait on an occupied reserved tile before failing loudly.
- **Reserved tile**: The tile a wall will occupy when built; the builder
  works adjacent to it; its occupancy at completion is what defers the build.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In a simulation where a wall build's site is invalidated
  mid-work, 100% of such cancellations produce a build-failure event and a
  builder failure memory; zero resolve as bare completions.
- **SC-002**: In a simulation where a second agent crosses the reserved tile
  mid-build and leaves, the wall completes successfully — where today the
  same scenario cancels the build 100% of the time.
- **SC-003**: In a simulation where an agent squats on the reserved tile
  indefinitely, the build fails loudly within the grace period bound — it
  neither completes under the squatter, waits forever, nor resolves silently.
- **SC-004**: A recorded world containing the new failure events and memories
  replays byte-identically.
- **SC-005**: An observer reading the TUI digest can distinguish every failed
  build from finished ones (failed builds never render as "finished").

## Assumptions

- The wall's occupancy guard exists to prevent entombing agents; the fix
  preserves that invariant absolutely (never build over an occupant) while
  removing the insta-cancel.
- The grace period is a fixed constant chosen at implementation time (order
  of tens-to-hundreds of ticks — long enough for a passerby or short
  conversation, short enough that a failed build resolves within a fraction
  of the build's own work duration); tuning it later is out of scope.
- Scope is the build goals in the executor's mid-work re-validation: loud
  failure applies to all build goals' site-vanished path; occupancy
  tolerance applies to wall builds (the only builds with a reserved-tile
  occupancy guard).
- Non-goals (recorded on TASK-91): pathing avoidance / reservation index for
  build sites; loud failure for non-build goals (forage/chop/hunt/cook/
  bathe/deposit/withdraw — follow-up work); changes to the material
  pipeline or planner behavior. Existing beliefs in already-running worlds
  are corrected only prospectively (no retroactive memory surgery).
- Chronicle/gist text is generated from conversations and is not directly
  corrected by this feature; falsifiability comes from the builder's failure
  memory entering the same retrieval paths.
