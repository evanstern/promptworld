# Feature Specification: Guardian directives and designations — the durable plan layer

**Feature Branch**: `084-guardian-directives` (task branch: `task-157-guardian-directives`)

**Created**: 2026-07-26

**Status**: Draft

**Input**: TASK-157 (board card, 7 ACs; sweep runbook
`docs/design/faith-directives-sweep-runbook.md`, signed-off 2026-07-26).
Dependency TASK-97 merged (PR #123): the `internal/target` grammar and the
contract-named "Designation addressing (TASK-157 seam)"
(`specs/036-scriptable-agent-tools/contracts/bundle-manifest.md` §"Designation
addressing"). The DIRECTIVE-hardness ruling (operator, 2026-07-26) is FIRM and
encoded here, not re-opened: directives are HARD — a DIRECTIVE rung in
`decideIntent` between SURVIVAL and PREP; interruption is life and must not be
discouraged; interruption problems get in-game workarounds first.

## Grounding (verified against the task worktree, 2026-07-26)

**What exists.** The guardian already has watch-and-act agency
(`sim.GuardianOrder`, spec 029/059 — the entity discipline this feature
clones: deterministic ids, one-way status, validate-not-clamp reducer arms,
active+32 retention prune, executor-emitted expiry), charge-priced world
edits (miracles, spec 016/021), villager-facing influence through the
`InjectSocial` door (visions/omens with the optional spec-041 place grant),
a read-only grounded-facts tool (`explain`, spec 063, Effect `Read`,
rendered via `GuardianReadGuidance`), and a miracle targeting digest
(spec 059, `buildTargetingDigest`). What it does NOT have is the
DF/RimWorld player's **plan-making verbs**: no way to survey a site, to
stake a durable, checkable, villager-visible claim on the world ("build
here", "wall this line", "settle this zone"), or to bind villagers to that
plan. Villager-side, `decideIntent` arbitrates SURVIVAL → PREP → wander
(`internal/sim/policy.go:40`) with no rung for externally-issued goals, and
the decision context (spec 043) has no block carrying a divine command.

**What TASK-97 delivered for this spec.** `internal/target` parses point /
rect / axis-aligned line loci and enumerates their tiles deterministically
(`Address.Tiles()`, `internal/target/target.go:265`) as a stdlib-only leaf
importable from `internal/tool`, `internal/sim`, and `internal/guardian`
alike. Rect and line forms are reserved for THIS consumer.

**The three pieces** (the board card's decomposition, encoded as user
stories below): SURVEY (a free read tool for planning), DESIGNATIONS
(event-sourced world plan artifacts with structural fulfillment
predicates), DIRECTIVES (hard, TTL-bounded villager bindings to a
designation, landed through the injection door and joined to the reflex
ladder and the decision context).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The guardian surveys a site before planning (Priority: P2) — AC #2

The guardian calls `survey_site(x, y, radius)` and receives a deterministic
site fact sheet — terrain mix, nearest water/wood/rock distances,
structures present, passability — free of charge, any number of times per
turn, without consuming the turn's one act. It is the `explain` +
spec-059-targeting-digest pattern pointed at planning: exact facts instead
of guessed geography, so designations land on real, buildable ground.

**Why this priority**: survey makes the other two pieces *aimable*, but
designations and directives are independently placeable without it (the
targeting digest already gives coarse aim). P2: high value, not the spine.

**Independent Test**: a fixture world with known terrain; call
`survey_site` twice with identical args and assert byte-identical fact
sheets; assert no charge spent, no event landed, and the turn's acting
cardinality untouched (the `explain` dispatch shape).

**Acceptance Scenarios**:

1. **Given** a world with water at distance 3 and a rock face at distance
   5 from (10,10), **When** the guardian calls
   `survey_site(x:10, y:10, radius:4)`, **Then** the returned sheet lists
   the terrain-kind mix inside the radius, the nearest water/tree/rock
   distances, every structure in the area, and a passability summary — and
   the same call returns byte-identical text.
2. **Given** a survey call, **When** it completes, **Then**
   `GuardianCharges` is unchanged, no event is appended, and the guardian
   may still perform its one mediated act this turn (Effect `Read`
   exemption — "surveying is looking, not an act").
3. **Given** out-of-bounds coordinates, **When** `survey_site` is called,
   **Then** the reply is a repairable in-fiction miss naming the world
   bounds (the `explain` unknown-topic shape), never a hard error.
4. **Given** a miracle-capable turn prompt, **Then** `survey_site` renders
   under the read-tools paragraph (`GuardianReadGuidance`), never under the
   acting-tools list.

---

### User Story 2 - The guardian stakes designations on the map (Priority: P1) — AC #3

The guardian places durable plan artifacts on the world: a **settlement
zone** (rect), a **structure site** (point, naming a structure kind), a
**wall line** (axis-aligned line). Each is event-sourced (rides
`sim.State` through snapshots and replay), renders on the TUI map, is
announced to every villager as place knowledge (the spec-041 place-grant
machinery), and carries a **structural fulfillment predicate** — a pure
state check (structure-of-kind-K-at-tile) evaluated deterministically —
so the world itself can say "done". The guardian can cancel one; the
entity discipline clones `sim.GuardianOrder` (deterministic IDs, one-way
status, prune).

**Why this priority**: designations are the durable plan layer itself —
the checkable goal objects directives bind to. Without them there is
nothing for US3/US4 to point at.

**Independent Test**: place each of the three kinds through the tool door
in a fixture world; assert the reducer state, the map render (tile
registry rows on all three surfaces), the villager mental-map grant, and
that building the named structure at the site flips the designation to
`fulfilled` via a recorded, replay-reproducible event.

**Acceptance Scenarios**:

1. **Given** an empty valid tile (4,5), **When** the guardian calls
   `place_designation(kind:"structure_site", target:"4,5",
   structure_kind:"shelter")`, **Then** one `designation.placed` lands
   through `InjectSocial`, `State.Designations` holds it `active` with a
   deterministic id `dsg-<tick>-<seq>` and reducer-stamped `PlacedSeq`,
   and every living villager's mental map gains a `designation` place
   fact (provenance `revealed`).
2. **Given** an active structure-site designation for `shelter` at (4,5),
   **When** a villager completes building a shelter there, **Then** the
   executor sweep emits `designation.fulfilled` at the next tick boundary
   (the `charge_regenerated` pattern) and the reducer transitions the
   designation `active → fulfilled` — reproduced identically on replay
   with no guardian running.
3. **Given** designations of all three kinds, **When** the TUI map
   renders, **Then** designation tiles show their registry glyphs (site,
   wall-line segment, zone perimeter), a real structure/villager/pile on
   the same tile wins the tile, and the legend and `?` walkthrough carry
   the new rows with no edit outside the registry table.
4. **Given** a wall-line designation `2,2->2,6`, **When** walls stand on
   every enumerated tile, **Then** it fulfills; **When** only some tiles
   are walled, **Then** it stays `active`.
5. **Given** `cancel_designation(id)`, **Then** `designation.cancelled`
   lands and the designation transitions `active → cancelled`; a second
   cancel of the same id is rejected at the door (one-way status, the
   `transitionGuardianOrder` shape).
6. **Given** a from-genesis replay of a world whose log contains the full
   designation lifecycle, **Then** the reconstructed state is
   byte-identical to the live run's.

---

### User Story 3 - The guardian issues hard directives through the injection door (Priority: P1) — AC #4, #7

The guardian binds villagers to a designation:
`issue_directive(designation_id, targets, text, ttl_days)` targeting one
villager, a named group, or the whole village, with in-fiction framing
text and a TTL. The directive lands as a recorded `directive.issued`
event through `InjectSocial` — the prompt firewall holds exactly as for
visions: guardian prose enters the sim only as recorded event data, and
villagers only ever see it re-rendered from state. `cancel_directive`
retires one. Every `directive.*` lifecycle event joins
`observableEventTypes`, so existing standing orders can watch directives
with ZERO new trigger code.

**Why this priority**: the binding verb is the feature's namesake and the
seam TASK-118 (faith) and TASK-158 (missions) both build on.

**Independent Test**: issue a directive to two villagers in a fixture
world; assert the event payload, the resolved living targets, the
companion memories, whitelist membership, dry-run rejections (dead
target, unknown designation, bad TTL), and that a `monitor_and_act` order
watching `directive.fulfilled` triggers when the directive fulfills —
without any new matching code.

**Acceptance Scenarios**:

1. **Given** an active designation `dsg-100-0`, **When** the guardian
   calls `issue_directive(designation_id:"dsg-100-0", targets:"Rega,
   Sage", text:"Raise the shelter I have marked.", ttl_days:3)`, **Then**
   one `directive.issued` lands atomically with one
   `agent.memory_added` per target in the same batch, `State.Directives`
   holds it `active` with resolved target indices, and the villagers'
   next decision prompts carry the directive block.
2. **Given** a directive naming a dead villager, an unknown designation
   id, a non-active designation, or a TTL outside 1..7 game days,
   **When** it is injected, **Then** the dry-run rejects the whole batch
   (nothing lands) and the handler feeds the model a repairable
   `rejected_gate` naming the reason.
3. **Given** an active directive whose bound designation fulfills,
   **Then** the executor sweep emits `directive.fulfilled` carrying
   `{id, designation_id, targets, issued_tick}` — the recorded contract
   TASK-118's faith accounting will consume.
4. **Given** a standing order placed via `monitor_and_act` with
   `event_types:["directive.fulfilled"]`, **When** a directive fulfills,
   **Then** the order triggers through the existing `matchOrders` path —
   no new trigger code (AC #7).
5. **Given** a directive with TTL 2 game days and an unfulfilled
   designation, **When** the TTL elapses, **Then** the executor sweep
   emits `directive.expired` and the directive transitions one-way; the
   same holds when every targeted villager has died (an un-executable
   directive does not haunt the state).
6. **Given** the villager prompt assembly, **Then** the directive's text
   reaches the villager ONLY via the state-rendered directive block and
   the recorded companion memory — the firewall audit finds no direct
   guardian-prose channel into villager prompts.

---

### User Story 4 - Villagers execute directives between survival and free time (Priority: P1) — AC #5, #6

Villager-side, a directive is HARD: a **DIRECTIVE rung** in
`decideIntent` between SURVIVAL and PREP. A villager first makes sure it
is not dying (survival rungs run first, unconditioned), then executes its
active directives (the rung resolves a concrete intent toward the bound
designation), then free time (prep/wander — which the rung preempts).
The directive also joins the decision context as a spec-043 block, so a
planner-driven villager pursues it intelligently. Conversations, hails,
and dynamic world stimuli CAN and SHOULD interrupt directed work —
interruption is life — and the directive resumes afterward through the
existing machinery, with zero new interruption code (the
in-game-workaround-first doctrine, proven by test).

**Why this priority**: hardness is the operator's firm ruling and the
behavioral heart of the feature; without the rung a directive is just a
suggestion.

**Independent Test**: a no-planner (reflex-only) fixture world; issue a
directive; assert the villager walks to the site and works it while fed
and warm, that a survival need below its band preempts the directive,
that prep/wander never fire while the rung resolves, and that a hail
mid-walk pauses the work and the directive resumes at the next idle
decision.

**Acceptance Scenarios**:

1. **Given** a fed, warm, reflex-only villager and an active
   structure-site directive addressing it, **When** it goes idle,
   **Then** `decideIntent` resolves a directive intent (walk to /
   build at the site), never prep or wander (AC #5, "directives preempt
   prep/wander").
2. **Given** the same villager with Food below `dangerFoodBelow`,
   **Then** the survival rung decides first and the directive waits
   (AC #5, "survival always preempts") — the rung sits AFTER
   `survivalDecision` and BEFORE the `prepYields` gate.
3. **Given** a villager walking to a directive site, **When** another
   villager hails it, **Then** the hail pauses the directed intent
   exactly as it pauses any intent, the conversation runs, and at the
   next idle decision the DIRECTIVE rung re-resolves toward the same
   site — asserted with zero new interruption-handling code in the diff
   (AC #6).
4. **Given** an active directive, **Then** the villager's decision prompt
   contains the `directive` block (guardian framing text, the bound
   designation's kind and site, plain-words time remaining) in contract
   order; **Given** no directive, **Then** the block renders empty and
   the assembled prompt is byte-identical to pre-084 output.
5. **Given** a directive whose designation the reflex cannot advance
   (e.g. materials the reflex cannot craft), **Then** the rung moves the
   villager to the site when away and otherwise falls through — the
   planner, reading the block, owns the clever part; the villager never
   deadlocks idle-at-site.
6. **Given** the full 062 parity drive on a directive-free world, **Then**
   reflex decisions are byte-identical to pre-084 behavior (the rung is
   inert without directives).

---

### Edge Cases

- **Directive targeting a dead villager**: rejected at the door (dry-run;
  the `metatron.nudged` living-target precedent). A comma-list with ANY
  dead or unknown name rejects whole (atomic; the model gets a repairable
  `rejected_gate` naming the villager). `targets:"everyone"` resolves to
  all villagers living at issue.
- **Group target where some members die later**: the directive stays
  active for the living; the rung and the context block simply skip dead
  targets. When the LAST target dies, the executor sweep emits
  `directive.expired` (a pure state check — no TTL wait).
- **Designation on an occupied or invalid tile**: out-of-bounds loci are
  rejected at the door (`ErrBounds`, the target taxonomy). A
  structure-site tile already holding a structure of a DIFFERENT kind is
  rejected (structurally unfulfillable — one structure per tile is a
  buildSite invariant); the same rule per-tile for wall lines
  (non-wall structure on a line tile rejects). Zone rects may freely
  contain anything.
- **Fulfillment when the structure pre-exists**: placement is accepted
  (the guardian may consecrate what stands); the executor sweep emits
  `designation.fulfilled` at the next tick boundary. A directive bound to
  it fulfills at the following sweep — harmless, honest, and
  deterministic.
- **TTL expiry mid-execution**: `directive.expired` lands while a villager
  walks to the site; the in-flight intent runs to its own completion (the
  intent machine is never reached into), but the rung stops re-selecting
  and the block stops rendering it. No thrash: expiry is one event, one
  transition.
- **Cancel while a villager walks to the site**: identical shape —
  `directive.cancelled` (or `designation.cancelled`, which orphans the
  directive: the rung skips a directive whose designation is non-active,
  and the sweep expires it) never cancels an in-flight intent.
- **Cancel/fulfil/expiry races**: exactly one terminal status ever lands —
  the `transitionGuardianOrder` door shape: the loser finds a non-active
  entity and is refused. Executor-emitted events are ordered
  deterministically within the tick batch (designations before
  directives, slice order within each), so replay agrees.
- **Directive bound to a cancelled designation**: refused at issue
  (designation must be `active`); orphaned after issue → expired by the
  sweep (above).
- **Interruption thrash** (a villager ping-ponged between hails and a
  distant site): NOT a code concern in this spec — the operator ruling is
  in-game-workaround-first: the guardian re-issues, shortens TTLs, or
  places a standing order watching `directive.expired` to re-issue.
  Code-level anti-thrash waits for evidence no in-game workaround exists.
- **Caps**: at most 16 active designations and 3 active directives
  (validated at the door, the `GuardianPlayerOrderCap` shape); zone area
  ≤ 256 tiles, wall line length ≤ 32 tiles (door-validated bounds).
  Non-active entities prune to the most recent 32 (the
  `pruneGuardianOrders` discipline), preserving trail without unbounded
  growth.
- **Time snap across an active directive**: `Directive.ExpiresTick` is a
  future deadline → SHIFT for active directives; placed/issued ticks are
  history → KEEP; designations carry no future deadline → KEEP
  (`rebaseTicks` taxonomy entries, `internal/sim/miracles.go:271`).
- **Ended world**: `stepEvents` emits nothing, so no sweep fires; the
  injected `designation.*`/`directive.*` types are refused by the
  ended-world narrowing exactly like the order events (recorded prose
  only after run end).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001 (SURVEY)**: A `survey_site` tool MUST join `guardianTools`
  with `Effect: Read`, `Gate: None`, params `x`/`y` (required Numbers)
  and `radius` (optional, default 4, clamped 1..8). It MUST return a
  deterministic fact sheet — terrain-kind mix within the radius, nearest
  water/tree/rock distances from the center, structures present
  (kind + tile), and a passability summary — computed purely from the
  turn's state snapshot + static map, appending no event and spending no
  charge. It renders through the `GuardianReadGuidance` path (never the
  acting list) and returns a repairable in-fiction miss for out-of-bounds
  input. [AC #2]
- **FR-002 (Designation entity)**: `sim.Designation` MUST clone the
  `GuardianOrder` discipline: deterministic id `dsg-<tick>-<seq>` (the
  `nextOrderID` shape, no RNG), one-way status
  `active → fulfilled | cancelled`, reducer-stamped `PlacedSeq`, caps
  (16 active) and bounds validated in the reducer arm
  (validate-not-clamp), retention prune (active + most recent 32
  non-active). Kinds: `settlement_zone` (rect, with `MinStructures`
  1..12, default 3), `structure_site` (point + required
  `StructureKind`), `wall_line` (axis-aligned line, optional wall-kind
  narrowing). Loci are stored NORMALIZED (ints, the `target.Address`
  shape) — payloads are self-contained; replay never re-parses. [AC #3]
- **FR-003 (Designation addressing)**: designation loci MUST parse
  through `internal/target` — the one-parser law — via a new exported
  bare-locus entry point (`ParseLocus`) reusing the existing locus
  grammar, normalization, and `Tiles()` enumeration. Per-kind form
  matrix: `settlement_zone`→rect, `structure_site`→point,
  `wall_line`→line; any other (kind × form) cell is a `form` error. The
  bundle-effect matrix and the four entity classes are UNTOUCHED. [AC #3]
- **FR-004 (Designation events)**: `designation.placed` (payload = the
  entity; `Status`/`PlacedSeq` ignored, reducer-stamped) and
  `designation.cancelled` ({id}) MUST be injected via `InjectSocial`
  (whitelist + validating reducer arms) by new `place_designation` /
  `cancel_designation` guardian tools (Effect Expressive, Gate None —
  charge-free plan verbs, see research R3). `designation.fulfilled`
  ({id}) MUST be executor-emitted (`stepEvents`, the
  `charge_regenerated` pattern): a pure function of (state, tick) that
  fires once when the designation's fulfillment predicate holds, so
  replay reproduces it with no guardian running. The reducer arm
  re-validates the predicate against state before transitioning. [AC #3]
- **FR-005 (Fulfillment predicates)**: each kind's predicate MUST be a
  pure state check evaluated identically by the sweep and the reducer
  arm: structure-of-`StructureKind`-at-tile (structure_site); a wall
  structure on EVERY enumerated line tile (wall_line); at least
  `MinStructures` structures within the rect (settlement_zone). No
  clocks, no RNG, no I/O. [AC #3]
- **FR-006 (Announcement)**: the `designation.placed` reducer arm MUST
  upsert one place fact per living villager — the spec-041 place-grant
  machinery: `PlaceFact{Kind:"designation", X,Y: anchor tile, Seen:
  e.Tick, Provenance: ProvenanceRevealed, Detail: kind + label}` — into
  each agent's mental map (map-less agents skipped; reducer stays
  total). `designation` joins the closed `PlaceFact` vocabulary with its
  own freshness-horizon entry; it does NOT join `placeFactKinds`
  (send_vision reveals real world places only). `renderKnownPlaces`
  MUST render designation facts as landmarks. [AC #3]
- **FR-007 (Map rendering)**: designations MUST render on the TUI map
  through NEW tile-registry rows only (spec 068's three-surface
  discipline: renderer, compact legend, `?` walkthrough from the one
  table): a structure-site glyph, a wall-line-segment glyph, and a
  zone-PERIMETER glyph (interior tiles unmarked — no wallpaper). A real
  world entity (structure/pile/villager) on the same tile wins.
  Fulfilled/cancelled designations stop rendering (state-derived).
  `docs/design/tui/` pages amended same-PR per the spec-047 gate. [AC #3]
- **FR-008 (Directive entity)**: `sim.Directive` MUST clone the same
  entity discipline: id `dir-<tick>-<seq>`, one-way
  `active → fulfilled | cancelled | expired`, reducer-stamped
  `PlacedSeq`, cap 3 active, TTL bounds 1..7 game days (default 3 — the
  `GuardianOrderTTL*` constants, shared, not copied), framing `Text`
  ≤400 runes, `DesignationID` binding an ACTIVE designation, `Targets`
  resolved to living villager indices at issue (plus a `Village` marker
  when issued to everyone). [AC #4]
- **FR-009 (Directive events + door)**: `directive.issued` (payload =
  the entity) and `directive.cancelled` ({id}) MUST be injected via
  `InjectSocial` by new `issue_directive` / `cancel_directive` guardian
  tools (Effect Expressive, Gate None); `directive.issued` MUST ride
  atomically with one companion `agent.memory_added` per target ("The
  Guardian charges you: <text>", the vision-memory shape), so the prompt
  firewall holds exactly as for visions: guardian prose enters the sim
  only as recorded event data. `directive.fulfilled`
  ({id, designation_id, targets, issued_tick}) and `directive.expired`
  ({id}) MUST be executor-emitted pure functions of (state, tick):
  fulfilled when the bound designation is `fulfilled`; expired when
  `tick >= ExpiresTick` OR no targeted villager remains alive. [AC #4]
- **FR-010 (Observability — AC #7)**: `directive.issued`,
  `directive.fulfilled`, `directive.cancelled`, and `directive.expired`
  MUST join `observableEventTypes`
  (`internal/tool/registry.go:418`) so `monitor_and_act` standing orders
  can watch the directive lifecycle with zero new trigger code — proven
  by a test that places a watch on `directive.fulfilled` and asserts it
  triggers through the existing `matchOrders` path.
- **FR-011 (Decision-context block)**: a `directive` block MUST join the
  spec-043 assembly (`internal/mind/context.go` `fixedBlocks`) in
  contract order between `plan_echo` and `known_places`, priority
  `neverDrop` (a hard command is never shed under budget pressure),
  rendering ≤2 active directives addressing the agent, oldest first:
  the guardian's framing text, the bound designation's kind + site
  coordinates, what fulfillment requires, and plain-words time
  remaining. Empty state renders `""` — a directive-free world's
  assembled prompt stays byte-identical to pre-084 output. [AC #4, #5]
- **FR-012 (DIRECTIVE reflex rung — AC #5)**: `decideIntent` MUST gain a
  `directiveDecision` rung invoked AFTER `survivalDecision` returns
  nothing and BEFORE the `prepYields` gate is consulted
  (`internal/sim/policy.go:40-55`), unconditioned by the yield window
  and danger bands (those belong to survival's ownership, which has
  already passed). The rung selects the OLDEST active directive
  addressing the agent whose bound designation is active, and resolves a
  concrete intent per the data-model routing table (build at the site
  when reflex-expressible with materials in hand; else walk to the
  site via a new `heed_directive` goal; else fall through). When the
  rung resolves, prep and wander never run; when it cannot, the ladder
  falls through normally (the planner, reading the block, owns the
  clever part — no idle-at-site deadlock).
- **FR-013 (Interruption-friendly — AC #6)**: NO new interruption,
  pause, or resume code. Hails, conversations, and dynamic stimuli
  interrupt directed work through the existing intent-pause machinery,
  and the directive resumes because the rung re-resolves at the next
  idle decision. A test MUST prove the interrupt-and-resume cycle and
  the diff MUST contain no interruption-handling changes. Thrash
  mitigation is in-game first (re-issue, TTLs, standing-order watches on
  `directive.expired`); code changes require evidence that no in-game
  workaround exists.
- **FR-014 (Determinism & replay)**: all state mutation via reducer arms
  only; sweeps pure over (state, tick); ids deterministic (no RNG);
  executor emissions ordered deterministically within the tick;
  `rebaseTicks` gains taxonomy entries (active `Directive.ExpiresTick`
  SHIFT; placed/issued ticks KEEP; designations all-KEEP). A
  from-genesis replay of a log containing every new event type MUST
  reproduce byte-identical state; `TestCatalogSweep` MUST pass with
  digest-grammar rows for all seven new event types.
- **FR-015 (Guardian prompt truthfulness)**: the turn prompt MUST carry
  active designations and directives (id, kind, site, days-left — the
  `writeStandingOrders` shape) so the angel's counsel stays truthful to
  live state, and `place_designation`/`issue_directive` guidance renders
  through `GuardianToolGuidance` (described ≡ declared by construction).
- **FR-016 (Scope guards)**: NO mission machinery (accept / decompose /
  pursue / report — TASK-158, gated on TASK-112); NO faith accounting
  (TASK-118 consumes the `directive.fulfilled` contract named in FR-009);
  NO change to bundle-effect target grammar or matrix; NO reducer-arm
  weakening anywhere; NO new RNG purposes.

### Key Entities

- **Designation** — an event-sourced world plan artifact: kind
  (settlement_zone | structure_site | wall_line), normalized locus,
  fulfillment predicate parameters, one-way status. Normative shape in
  [data-model.md](data-model.md) §2.
- **Directive** — a hard, TTL-bounded binding of villagers to a
  designation, with guardian framing text; one-way status with three
  terminals. Normative shape in [data-model.md](data-model.md) §3.
- **Directive/designation event vocabulary** — seven new event types,
  their payloads, doors, and whitelist/observable membership:
  [contracts/events.md](contracts/events.md) (normative).
- **DIRECTIVE rung** — the new arbitration position:
  SURVIVAL → DIRECTIVE → (prepYields?) PREP → wander; routing table and
  pseudocode in [data-model.md](data-model.md) §6.
- **`directive` context block** — the eleventh spec-043 block;
  [data-model.md](data-model.md) §7.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: in a reflex-only fixture world, an issued structure-site
  directive takes a fed, warm villager to the site and the designation
  reaches `fulfilled` with no planner call; the survival-preemption and
  prep-preemption matrices pass (`go test ./internal/sim/`). [AC #5]
- **SC-002**: the hail-interruption test passes: directed work pauses for
  the conversation and resumes at the next idle decision; `git diff`
  shows zero changes to hail/pause/resume machinery. [AC #6]
- **SC-003**: a `monitor_and_act` watch on `directive.fulfilled` triggers
  through the unmodified `matchOrders` path in an integration test.
  [AC #7]
- **SC-004**: full-lifecycle replay byte-identity: a world log containing
  place/issue/fulfil/cancel/expire for both entities reconstructs
  byte-identical state from genesis; `go test ./...` green. [AC #3, #4]
- **SC-005**: `survey_site` returns byte-identical sheets for identical
  (args, state) and appends nothing; the acting-cardinality audit passes.
  [AC #2]
- **SC-006**: prompts and reflex decisions on a directive-free world are
  byte-identical to pre-084 output (block renders empty; rung inert);
  the spec-062 parity drive stays green.
- **SC-007**: `TestCatalogSweep` passes with the seven new digest rows;
  the tile-registry sweep tests pass with the three new rows;
  `node scripts/check-tui-design.mjs --changed` exits 0 with
  `docs/design/tui/` amended in-branch.
- **SC-008**: the merge-drift pr gate exits 0 from the worktree (wiki
  notes re-pinned in-branch, `docs/player/` regenerated); the PR merges
  with `gh pr merge --merge`; spec-bridge sync moves the board.

## Assumptions

- **Plan verbs are charge-free** (research R3): survey, designations, and
  directives spend no charges — charges price world EDITS (miracles);
  the plan layer is the player's basic vocabulary (the indirect-control
  doctrine) and charging re-issues would fight the in-game-workaround
  ruling. Flagged for operator review; a future economy pass can add
  pricing at the door without shape changes.
- **Stage availability**: the five new tools follow `monitor_and_act`'s
  precedent — granted at every curriculum stage (the plan loop is a
  teaching primitive, like watches). Flagged for operator review.
- **Zone fulfillment is count-based**: `MinStructures` (default 3) within
  the rect. A richer "settled" predicate (kinds, occupancy) is a future
  spec's one-table change.
- **Oldest-first execution**: the rung executes the oldest active
  directive addressing the agent; the guardian re-prioritizes by
  cancel + re-issue (in-game verb, deterministic tie-break).
- **Directive goals are designation-bound in v1**: `DesignationID` is the
  only checkable-goal binding; the entity shape (a discriminated goal
  reference) leaves room for other checkable goals later without an
  event-vocabulary change.
- **New event namespaces** (`designation.*`, `directive.*`) rather than
  `metatron.*`: these are world plan artifacts and villager-facing
  bindings, not guardian-console bookkeeping; the board card itself
  names the `directive.*` vocabulary (AC #7 wording).
- Wiki notes pinning touched sources re-pin in-branch per the pr gate
  (expected set enumerated in plan.md Phase 7); the gate is the
  authority — produce what it names.
