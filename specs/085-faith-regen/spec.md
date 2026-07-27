# Feature Specification: Faith-driven charge regeneration — the endogenous mana loop

**Feature Branch**: `085-faith-regen` (task branch: `task-118-faith-regen`)

**Created**: 2026-07-26

**Status**: Draft

**Input**: TASK-118 (board card, 4 ACs; Wave-3 operator ratification
2026-07-25 decision 4; realignment 2026-07-26: FULFILLED DIRECTIVES are the
natural endogenous faith source; sweep runbook
`docs/design/faith-directives-sweep-runbook.md`, signed-off 2026-07-26).
Dependencies merged: TASK-67 (duel) and TASK-157 (guardian directives —
`directive.fulfilled` is the contract-named faith seam,
`specs/084-guardian-directives/contracts/events.md` §3). The spec-077 rider
rides this spec: the `first-faith-event` lesson lands here because this spec
introduces the first faith event type (FR-020 deferral,
`internal/tui/lessons.go:236-237`). The strip integration is PRE-SPECIFIED:
`docs/design/tui/panels/guardian-strip.md` §4 (the dashed faith segment
contract) — this spec satisfies it, it does not renegotiate it.

## Grounding (verified against the task worktree, 2026-07-26)

**What exists.** Charge regeneration is clock-driven: `stepEvents` emits
`metatron.charge_regenerated` at absolute 6-game-hour boundaries when the
bank is below cap (`internal/sim/executor.go:53-57`,
`chargeRegenTicks = 6*3600`, `GuardianChargeCap = 3`,
`internal/sim/guardian.go:16-31`) — a pure function of (state, tick), the
replay idiom every later sweep cites as its precedent. TASK-157 shipped the
directive layer: `directive.fulfilled{id, designation_id, targets,
issued_tick}` is executor-emitted when a directive's bound designation
fulfills (`internal/sim/executor.go:100-113`,
`internal/sim/plans.go:112` — "THE TASK-118 faith-accounting seam"), and
`directive.expired` fires on TTL lapse or all-targets-dead. Belief
provenance (spec 030) already distinguishes guardian-delivered content:
every memory carries a closed-vocabulary `Origin`
(`OriginOmen` for delivered visions/omens), `DirectPerception(origin)` is
the sole text-free classifier, and the nightly consolidation's provenance
gate coerces belief provenance from cited memory origins
(`docs/wiki/agent-memory-window.md`, `docs/wiki/nightly-consolidation.md`).
Scenario posture is boot-frozen and replay-re-armed
(`internal/sim/state.go:214-222`; the spec-054 incident sweep keys on
`s.scenario != nil` with byte-identical replay). The strip renderer today
has NO faith code path (`internal/tui/views.go:2679-2725`) and a regression
test pins its absence (`internal/tui/render_test.go:599-600`).

**What this spec adds.** Faith as event-sourced village state; a defined,
machine-checkable prophecy-verification rule (a new `prophesy` guardian
tool and a `prophecy.*` lifecycle); charge regen rewired to a pure
function of faith (band cadences replacing the single constant); the
explicit failure-spiral posture (scenario spiral vs ambient floor, with
its reversal lever); the strip's §4 faith segment; and the
`first-faith-event` lesson row.

**The one-sentence loop** (the god-game mana doctrine,
`research/Game-Gameplay-Patterns/Indirect-Control-and-Divine-Intervention.md`
§"The mana economy"): power derives from the prosperity of the flock —
better prompting → fulfilled directives and true prophecies → more faith →
faster charge regen → more capacity to help the village → more faith. This
is the ambient endgame's unscored score, kept strictly in-fiction
(overjustification caution: faith is a world fact, never a badge, streak,
or congratulatory surface).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The flock's fulfilled work feeds the guardian (Priority: P1) — AC #2

Faith is village state that moves ONLY when recorded events fold: a
fulfilled directive raises it (the primary endogenous source — the
guardian's charge to the villagers, achieved), an expired directive and a
villager death lower it, a fulfilled prophecy raises it, a failed prophecy
lowers it more. Every movement is itself a recorded `faith.changed` event
emitted by the executor in the same tick batch as its source event — no
ambient accrual, no clock drip, no model judgment anywhere in the loop.

**Why this priority**: faith accounting is the feature's spine; regen
(US2) and display (US4) are functions of it.

**Independent Test**: a fixture world drives a directive to fulfillment
and asserts one `faith.changed{delta:+F, reason:"directive_fulfilled",
source_id:<directive id>}` lands in the same batch as
`directive.fulfilled`, the reducer folds it with clamping, a from-genesis
replay reproduces the score byte-identically, and a world with no faith
events reads the genesis score through the nil-safe accessor.

**Acceptance Scenarios**:

1. **Given** an active directive whose bound designation fulfills,
   **When** the executor sweep emits `directive.fulfilled`, **Then** the
   SAME tick batch carries `faith.changed{delta:+8,
   reason:"directive_fulfilled", source_id:<id>}` and the reducer moves
   `FaithScore()` up by 8 (clamped to 100).
2. **Given** a directive that expires unfulfilled (TTL or all targets
   dead), **Then** the batch carries
   `faith.changed{delta:-4, reason:"directive_expired", source_id:<id>}`.
3. **Given** a villager dies (any cause), **Then** the batch carries one
   `faith.changed{delta:-6, reason:"villager_died", source_id:<agent
   index>}` per death — the flock's suffering erodes faith, feeding the
   designed spiral.
4. **Given** faith already at 100 and a directive fulfills, **Then** NO
   `faith.changed` is emitted (the `charge_regenerated` below-cap idiom:
   never record a movement that cannot move anything); likewise at 0 for
   negative sources.
5. **Given** a pre-085 world log (no faith events), **When** replayed
   from genesis under the new code, **Then** the reconstructed state is
   byte-identical to what the old code produced — old events' semantics
   are untouched; faith derives EXCLUSIVELY from recorded `faith.changed`
   events, never retroactively from old `directive.fulfilled` rows.
6. **Given** a `designation.fulfilled` with no directive bound to it,
   **Then** NO faith moves — villager initiative is not the guardian's
   word (considered-and-excluded source, research R3).

---

### User Story 2 - Regen is a pure function of faith (Priority: P1) — AC #3, #4

The single 6-game-hour regen constant becomes a faith-band cadence:
fervent villages refill the bank faster, forsaken villages slower — or,
in a scenario world, not at all. The check keeps today's exact shape —
`nextTick % cadence == 0 && charges < cap`, absolute boundaries, the same
`metatron.charge_regenerated` event with the same empty payload — so
replay determinism is inherited, not re-proven from scratch. The
failure-spiral posture is decided HERE, explicitly: **scenario worlds get
the genre-authentic spiral (regen stops entirely in the forsaken band);
ambient worlds get a floor (regen never slower than once per game day)**,
keyed on the boot-frozen scenario presence, with a one-table reversal
lever (research R5).

**Why this priority**: this is the board card's title mechanic and the
half of the loop that makes faith matter.

**Independent Test**: table-driven cadence tests over (score, posture)
cells; a fixture drive at each band asserting boundary ticks fire and
off-boundary ticks don't (the `TestChargeRegen` shape,
`internal/sim/guardian_test.go:144-168`); a genesis-band world (no faith
events) produces a regen schedule byte-identical to pre-085.

**Acceptance Scenarios**:

1. **Given** `FaithScore()` in the steady band (40–74, where genesis 50
   sits), **Then** regen fires at absolute 6-game-hour boundaries —
   byte-identical scheduling to today for any world that has never
   folded a faith event.
2. **Given** faith ≥ 75 (fervent), **Then** the cadence is 4 game hours;
   **Given** 15–39 (wavering), **Then** 12 game hours.
3. **Given** faith < 15 (forsaken) in an AMBIENT world (no scenario
   armed), **Then** regen continues at 24 game hours — the floor: slow,
   painful, never a hard lock.
4. **Given** faith < 15 in a SCENARIO world, **Then** NO
   `charge_regenerated` fires while the band holds — the authentic
   spiral; the run can die of it, and the morgue/report card carries the
   lesson (the Hades doctrine: looping through failure is part of the
   experience).
5. **Given** any band, **When** the world is replayed from genesis,
   **Then** every `charge_regenerated` in the log reproduces — the
   cadence function is pure over (faith score, scenario presence, tick)
   and both inputs are event-sourced/boot-frozen exactly like today's.
6. **Given** faith at the floor and an empty bank (the "charges needed
   for the fix" edge), **Then** the recovery valve exists WITHOUT
   charges: the plan verbs are charge-free (spec 084 research R3), so
   the guardian can still place designations and issue directives, whose
   fulfillment mints the faith that re-speeds regen — low faith never
   removes the ability to earn faith back.

---

### User Story 3 - The guardian prophesies and the world judges the word (Priority: P2) — AC #2

The guardian stakes its credibility: `prophesy(targets, text, claim,
deadline_days)` spends one charge (the `send_vision` price — prophecy IS
a vision, with a wager attached), delivers the word to the targets as
omen-origin memories, and records a machine-checkable CLAIM with a
deadline. **The verification rule — what makes a vision 'true'**: a
vision is true exactly when its recorded claim — declared before the
fact, from a closed predicate vocabulary — is satisfied by the world's
own recorded state within its deadline, judged by a pure (state, tick)
predicate in the executor sweep and re-validated at the reducer door.
Checkable from the log end to end; never model-graded; free text without
a claim is counsel, not prophecy, and never mints faith. A prophecy
cannot be cancelled — the word, once given, stands — and a claim that
comes true after its deadline mints nothing (the terminal `failed`
status latched first and statuses are one-way).

**Why this priority**: prophecy completes the corpus loop ("better
prompting → truer prophecies → more faith") and is the risk half of the
economy — but the primary source (directives, US1) works without it.

**Independent Test**: declare each claim kind through the tool door in a
fixture world; assert the charge spend, the companion omen memories, the
door's rejection table (already-true claim, duplicate active claim, bad
deadline, dead target), fulfillment/failure sweeps firing once with
faith companions in-batch, and from-genesis replay byte-identity.

**Acceptance Scenarios**:

1. **Given** an active designation `dsg-100-0`, **When** the guardian
   calls `prophesy(targets:"everyone", text:"Before three dawns the
   shelter I have marked will stand.",
   claim:{kind:"designation_fulfilled", designation_id:"dsg-100-0"},
   deadline_days:3)`, **Then** one charge is spent, one
   `prophecy.declared` lands through `InjectSocial` atomically with one
   `agent.memory_added` per living target (`OriginOmen`, the vision
   memory shape), and `State.Prophecies` holds it `active` with id
   `pro-<tick>-<seq>`.
2. **Given** the designation fulfills before the deadline, **Then** the
   executor sweep emits `prophecy.fulfilled{id}` once, the same batch
   carries `faith.changed{delta:+12, reason:"prophecy_fulfilled",
   source_id:<id>}` plus one companion memory per living target
   (`OriginReport` — word spreads; a report is honestly secondhand, so
   the spec-030 provenance gate still refuses to launder it into
   "witnessed"), and the reducer transitions `active → fulfilled`.
3. **Given** the deadline passes with the claim unsatisfied, **Then**
   the sweep emits `prophecy.failed{id}`, the batch carries
   `faith.changed{delta:-15, reason:"prophecy_failed", source_id:<id>}`
   plus per-target `OriginReport` memories (the word did not come to
   pass), and the status latches `failed` — a later true-ing of the
   claim mints nothing (the "verifies after the TTL" edge).
4. **Given** a claim already true at declaration (prophesying the past),
   **Then** the door rejects it (dry-run, repairable `rejected_gate`
   counsel); **Given** an identical claim to an already-active prophecy,
   **Then** the door rejects the duplicate — faith cannot be farmed by
   restating one truth.
5. **Given** a `survives{agent}` claim and the villager dies before the
   deadline, **Then** the sweep fails the prophecy at that boundary
   (fail-fast — the claim is already unsatisfiable); **Given** the
   villager lives to the deadline, **Then** it fulfills at the first
   sweep tick ≥ deadline.
6. **Given** claim eligibility at both terminals on one boundary,
   **Then** fulfilled wins (checked first — the directive sweep's
   precedent) and exactly one terminal ever lands.
7. **Given** every prophecy target has died but the world lives,
   **Then** the prophecy stays active and is still judged — the word was
   spoken to the world, not contingent on its hearers (unlike a
   directive, which expires all-dead because no one can execute it);
   companion memories at the terminal simply skip the dead.

---

### User Story 4 - Faith is visible, in fiction (Priority: P2)

The strip's reserved fourth segment (guardian-strip.md §4) comes alive:
`faith N` when the wire carries a score, `faith —` (present, dashed) when
the TUI runs against a daemon that predates the field — the strip never
claims a mechanic that doesn't exist. The regen forecast segment stays
honest under variable cadence (it forecasts the next boundary of the
EFFECTIVE cadence, and is omitted when no regen is scheduled — full bank,
or the scenario forsaken band). The four new event types get in-fiction
digest rows; the `first-faith-event` lesson row joins the catalog
(closing the spec-077 rider); and the CLI/IPC status carries faith so a
non-TUI observer loses nothing (the strip's D1 projection rule). All of
it renders from state — deterministic, model-free — and all of it speaks
in-fiction: devotion and doubt, never points, streaks, or badges.

**Why this priority**: the pre-specified strip contract and the lesson
rider are binding obligations of this task, but they render what US1–US3
create.

**Independent Test**: strip render tests for the populated / dashed /
absent-forecast states (extending `render_test.go`, which today pins the
segment's absence and flips to pinning its presence); `TestCatalogSweep`
with the four new rows; the lessons taxonomy test flipped from
absence-pin to presence; a `check-tui-design --changed` pass with
`panels/guardian-strip.md` re-verified in-branch.

**Acceptance Scenarios**:

1. **Given** a status snapshot carrying `faith: 62`, **Then** the strip
   renders `faith 62` as the fourth segment (drop order under width
   pressure: faith first, per the §4/`joinStripSegments` contract).
2. **Given** a status snapshot WITHOUT the faith field (older daemon),
   **Then** the strip renders `faith —` — present, dashed, claiming
   nothing.
3. **Given** charges below cap and an effective cadence of 12 game
   hours, **Then** the regen segment forecasts the next absolute
   12-hour boundary; **Given** the scenario forsaken band (no regen
   scheduled), **Then** the regen segment is OMITTED (the R4.1 honesty
   rule generalized: never forecast an arrival that isn't scheduled).
4. **Given** the first `faith.changed` a world ever folds, **Then** the
   `first-faith-event` lesson fires once, worded in-fiction and
   direction-neutral (faith can first move DOWN — a death — and the
   copy must not congratulate).
5. **Given** any of the four new event types in the chronicle feed,
   **Then** its digest row renders in-fiction (e.g. "The village's
   faith deepens" / "wavers"), and `TestCatalogSweep` passes.

---

### Edge Cases

- **World with zero directives ever**: faith sits at genesis (50, the
  steady band) and moves only on deaths and prophecy outcomes; a
  peaceful, untouched world regenerates exactly as pre-085 forever. A
  neglected world whose villagers die drifts down-band — by design (the
  god-game reading: an unshepherded flock loses faith) — bounded by the
  ambient floor.
- **Faith at floor when charges are needed for the fix**: ambient floor
  guarantees a charge within 24 game hours, and the charge-free plan
  verbs (spec 084 R3) mean directive-earned faith is reachable with an
  empty bank — the endogenous exit from the spiral. Scenario worlds
  deliberately lack the guarantee (US2 AS-4); the run may die and the
  morgue teaches.
- **Pre-085 worlds and replay**: `State.Faith` and `State.Prophecies`
  are `omitempty` — pre-085 snapshots round-trip byte-identically, no
  format bump (the spec-029/084 precedent). Old logs contain no
  `faith.*`/`prophecy.*` events and replay byte-identically; faith is
  never derived retroactively from old events (US1 AS-5). A LIVE pre-085
  world continued under new code starts minting faith events from its
  next boundary — forward evolution, not a replay hazard.
- **Vision that verifies after the deadline**: `failed` latched first;
  one-way status; no faith. The word must come true in its appointed
  season (US3 AS-3).
- **Multiple prophecies about the same event**: distinct claims about
  one outcome each stand or fall on their own (each was a separate
  charge and a separate staked risk); an IDENTICAL claim (same kind,
  same normalized parameters) to an active prophecy is refused at the
  door (US3 AS-4). Cap 3 active prophecies (the directive-cap shape).
- **Prophecy about a designation that gets cancelled**: the claim
  `designation_fulfilled{id}` can no longer come true (one-way statuses)
  but is not yet provably false before the deadline — the sweep simply
  fails it at deadline. No special case: the predicate table already
  yields this.
- **Same-boundary source pileup**: a directive fulfills, another
  expires, and a villager dies in one batch — one `faith.changed` per
  source in the batch's fixed emission order (deterministic; deltas fold
  sequentially with clamping).
- **Time snap across an active prophecy**: `DeadlineTick` is a future
  deadline → SHIFT for active prophecies; `DeclaredTick` is history →
  KEEP; `FaithState` carries no tick → untouched (`rebaseTicks`
  taxonomy, `internal/sim/miracles.go` precedent).
- **Ended world**: `stepEvents` emits nothing — no faith movement, no
  verification, no regen (the run-end latch, spec 044); the injected
  `prophecy.declared` is refused by the ended-world narrowing like every
  non-prose type.
- **Band boundary flap**: a score oscillating across a band edge changes
  cadence each way — deterministic (pure function of folded score), and
  bounded in practice by the coarse deltas; no hysteresis in v1
  (recorded as a watch item, research R4).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001 (Faith state)**: `sim.State` MUST gain
  `Faith *FaithState` (`json:"faith,omitempty"`) where
  `FaithState{Score int}` is clamped 0..100. A nil `Faith` means the
  genesis default (`FaithGenesis = 50`) — the `State.Tuning`
  nil-means-default precedent — read ONLY through a nil-safe accessor
  `s.FaithScore()`. Pre-085 snapshots round-trip byte-identically; no
  `format_version` bump. [AC #2]
- **FR-002 (The faith event)**: ONE new event type `faith.changed`
  (`FaithChangedPayload{delta, reason, source_id}`) MUST be
  executor-emitted only (`stepEvents`; NOT whitelisted — injection
  refused by whitelist absence, the `charge_regenerated` class).
  `reason` is a closed vocabulary: `directive_fulfilled` (+8),
  `directive_expired` (−4), `villager_died` (−6), `prophecy_fulfilled`
  (+12), `prophecy_failed` (−15). The reducer arm validates the reason
  domain and a non-zero delta, then folds `Score = clamp(Score+delta,
  0, 100)`, materializing `Faith` on first fold. Deltas are named
  constants in ONE table (`faith.go`), promoted-dial-READY (named, one
  place — the spec-059 survival-band discipline: dials are earned by
  evidence, not pre-installed in tuning.json). [AC #2]
- **FR-003 (Faith accounting sweep)**: a `faithEvents(s, batch,
  nextTick)` sweep MUST run in `stepEvents` AFTER every faith-source
  emitter and BEFORE the scenario rubric and run-end detection — pure
  over (pre-tick state, this tick's batch, tick), the run-end
  detector's own idiom (`internal/sim/executor.go:424-446`). It scans
  the batch for `directive.fulfilled`, `directive.expired`,
  `agent.died`, `prophecy.fulfilled`, `prophecy.failed` and emits one
  `faith.changed` per source event, in the batch's order, SKIPPING any
  emission whose fold could not move the clamped score (US1 AS-4). No
  other faith source exists: no ambient accrual, no time decay, no
  model judgment, and `designation.fulfilled` deliberately mints
  nothing (research R3). [AC #2]
- **FR-004 (Regen curve)**: the executor's regen check MUST become
  `nextTick % FaithRegenCadenceTicks(s.FaithScore(), s.scenario != nil)
  == 0 && GuardianCharges < GuardianChargeCap`, REPLACING the fixed
  `chargeRegenTicks` (research R2 — replace, not coexist: one boundary
  check, one event type, the constant survives as the steady band's
  cadence). `FaithRegenCadenceTicks` is an exported pure function with
  the band table: score ≥ 75 → 4 game hours; 40–74 → 6 (genesis band —
  today's constant); 15–39 → 12; < 15 → 0 (no regen) when a scenario is
  armed, 24 game hours otherwise. Cadence 0 means the check never
  fires. Event type and payload (`metatron.charge_regenerated`, empty)
  are UNCHANGED. [AC #3]
- **FR-005 (Spiral posture — the AC #4 decision)**: the failure-spiral
  posture is DECIDED as FR-004's forsaken row: **scenario worlds spiral
  authentically (cadence 0); ambient worlds floor at 24 game hours** —
  grounded in the Hades God-Mode reasoning (failure must stay
  meaningful where the loop is run-shaped; the sting must be bounded
  where the world is persistent and has no run-reset —
  `research/Learning-Game-Design/Meta-Progression-and-Failure.md`
  §Hades) and in the charge-free plan verbs as the endogenous recovery
  valve (US2 AS-6). The REVERSAL LEVER: the whole posture lives in the
  one band table (four cadences + the posture fork), promoted-dial-ready;
  flipping ambient to authentic (or flooring scenarios) is a one-row
  change, and a future `tuning.json` promotion
  (`faith_floor_cadence_ticks`, 0 = no floor) needs no shape change.
  This requirement IS the operator checkpoint's artifact: the decision
  and its lever are in the spec, surfaced by the orchestrator. [AC #4]
- **FR-006 (Prophecy entity)**: `sim.Prophecy` MUST clone the
  spec-084 entity discipline: deterministic id `pro-<tick>-<seq>` (the
  `nextOrderID` shape, no RNG), one-way status
  `active → fulfilled | failed`, reducer-stamped `PlacedSeq`, cap 3
  active (`GuardianProphecyCap`, validated at the door), retention
  prune (active + most recent 32), `Text` ≤400 runes (the registry
  `TextCapBytes` cap `send_vision` already anchors —
  `internal/sim/guardian.go:38-41`), `Targets` resolved to living
  villager indices at declaration (`"everyone"` supported), `Claim` (a
  discriminated, normalized predicate — FR-007), `DeclaredTick`, and
  `DeadlineTick` with TTL bounds 1..7 game days (the shared
  `GuardianOrderTTL*` constants, not copies). There is NO cancel verb:
  the word, once given, stands. [AC #2]
- **FR-007 (The verification rule + predicate vocabulary)**: a vision
  is 'true' exactly when its recorded claim — declared before the fact
  — is satisfied by recorded world state within its deadline, judged by
  pure (state, tick) predicates. The claim vocabulary is CLOSED, each
  kind defining a fulfil condition and a fail condition (both pure; the
  normative table is [data-model.md](data-model.md) §5):
  `designation_fulfilled{designation_id}`,
  `structure_count{structure_kind, min}`, `population_at_least{min}`,
  `survives{agent}` (fail-fast on death, fulfilled at deadline if
  alive). Free text is NEVER graded; no model output ever participates
  in verification; a claim already true at declaration is refused at
  the door, as is a duplicate of an active claim (normalized equality).
  [AC #2]
- **FR-008 (Prophecy events + door)**: `prophecy.declared` (payload =
  the entity; `Status`/`PlacedSeq` reducer-stamped) MUST be injected
  via `InjectSocial` (whitelist + validating arm) by a new `prophesy`
  guardian tool (`Gate: Charge`, 1 — the `send_vision` price: prophecy
  is influence on minds, and the wager needs a stake), riding
  atomically with one companion `agent.memory_added` per living target
  (`OriginOmen`, dream-band salience — the vision/directive companion
  shape). `prophecy.fulfilled{id}` and `prophecy.failed{id}` MUST be
  executor-emitted (NOT whitelisted): the verification sweep evaluates
  active prophecies in slice order, fulfil condition BEFORE fail
  condition at the same boundary (exactly one terminal ever lands; the
  directive sweep's precedent), and the reducer arms re-validate the
  condition against state before transitioning. Terminal sweeps also
  emit one companion `OriginReport` memory per living target (word
  spreads; honestly secondhand under `DirectPerception`). [AC #2]
- **FR-009 (Wire + strip)**: the IPC status clock
  (`internal/ipc/protocol.go` `ClockStatus`, beside `metatron_charges`)
  MUST gain `faith *int` (`omitempty`) and `faith_regen_ticks int64`
  (the EFFECTIVE cadence, 0 = no regen scheduled), served from
  `FaithScore()`/`FaithRegenCadenceTicks` — the strip's D1 projection
  rule: the CLI/IPC observer sees everything the strip shows.
  `guardianStripView` MUST render the fourth segment per
  guardian-strip.md §4: `faith N` when the pointer is non-nil,
  `faith —` (dashed) when nil (older daemon), positioned last with
  first-drop truncation priority; the regen forecast MUST use the wire
  cadence when present (falling back to the legacy exported constant
  against an older daemon) and be omitted when cadence is 0 or the bank
  is full. `docs/design/tui/panels/guardian-strip.md` is amended and
  re-verified in the same PR (`check-tui-design --changed`, spec-047
  gate). [strip §4]
- **FR-010 (Digest rows + observability)**: the four new event types
  MUST get in-fiction digest-grammar rows (`internal/tui/grammar.go`;
  `TestCatalogSweep` is the gate). `prophecy.declared`,
  `prophecy.fulfilled`, `prophecy.failed` join `observableEventTypes`
  (enum-only, zero new trigger code — the directive precedent, so a
  standing order can watch "when my prophecy fails, tell me");
  `faith.changed` does NOT join in v1 (guardian-economy bookkeeping the
  strip already surfaces; widening later is a compatible enum-only
  change).
- **FR-011 (The lesson — spec-077 rider)**: `first-faith-event` joins
  `lessonCatalog` (tier mechanics, trigger `faith.changed`,
  direction-neutral in-fiction copy with skin tokens, pointer at the
  strip/guardian tab), and the taxonomy tests flip: the absence pin
  (`internal/tui/lessons_test.go:119-122`) is REMOVED and the presence
  is asserted. [rider, spec 077 FR-020]
- **FR-012 (Rubric hygiene — the recorded R2 obligation)**: `faith.*`
  joins the banned prefix set in
  `TestRubricHygieneNoTutorLaneTerms` (`internal/sim/rubric_hygiene_test.go:22-24`
  records this exact obligation): no exercise rubric may ever grade
  faith — faith is the unscored score, in-fiction only
  (overjustification caution). Prophecy events remain
  rubric-eligible world events.
- **FR-013 (Guardian prompt truthfulness)**: the guardian turn prompt
  MUST carry the faith score (in-fiction wording) and active prophecies
  (id, claim, deadline — the `writeStandingOrders` shape), and the
  `prophesy` tool's guidance renders via `GuardianToolGuidance`
  (described ≡ declared by construction).
- **FR-014 (Determinism & replay)**: all state mutation via reducer
  arms; sweeps pure over (pre-tick state, batch, tick); no new RNG
  purposes; ids deterministic; `rebaseTicks` gains the taxonomy entries
  (active `Prophecy.DeadlineTick` SHIFT; `DeclaredTick` KEEP;
  `FaithState` untouched). A from-genesis replay of a log containing
  the full prophecy lifecycle and every faith reason MUST reproduce
  byte-identical state; a pre-085 log MUST replay byte-identically; a
  no-faith-event world's regen schedule MUST be byte-identical to
  pre-085. [AC #2, #3]
- **FR-015 (Scope guards)**: NO metatron agentization (TASK-112); NO
  mission machinery (TASK-158); NO faith from tutoring — the tutor
  voice spends no charges, lands no world events, and earns no faith
  (TASK-112 AC #6, already legislated; FR-012 is its enforcement
  hook); NO faith influence on villager behavior, prompts, beliefs, or
  cognition cadence (faith reads nothing into villager decision
  contexts; `internal/cognition` and `internal/mind` untouched); NO
  rumor-driven faith spread (fulfillment memories are personal,
  `Subject: -1` — guardian-subject gossip is identified, deferred); NO
  tuning.json promotion (dial-READY only); NO badge/streak/score
  surface anywhere.

### Key Entities

- **FaithState** — the clamped 0..100 village faith score; nil means
  genesis. Normative shape in [data-model.md](data-model.md) §1.
- **`faith.changed`** — the one faith-movement event; closed reason
  vocabulary and delta table in [data-model.md](data-model.md) §2–§3.
- **Prophecy** — a charge-priced, uncancellable, deadline-bounded
  declared claim; [data-model.md](data-model.md) §4.
- **Claim predicate table** — the closed verification vocabulary with
  per-kind fulfil/fail conditions; [data-model.md](data-model.md) §5
  (normative).
- **Regen band table** — the faith → cadence pure function including
  the posture fork; [data-model.md](data-model.md) §6 (normative; the
  AC #4 artifact).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: fixture drive: directive fulfillment mints
  `faith.changed` in-batch; deltas fold with clamping; from-genesis
  replay of the full faith+prophecy lifecycle is byte-identical
  (`go test ./internal/sim/`). [AC #2]
- **SC-002**: cadence table tests pass for every (band × posture) cell;
  boundary/off-boundary drive per band (the `TestChargeRegen` shape);
  a no-faith-event world's regen schedule is byte-identical to
  pre-085. [AC #3]
- **SC-003**: the posture fork is proven both ways: a scenario fixture
  at forsaken faith regenerates nothing; an ambient fixture regenerates
  at 24h. The decision + lever are recorded (FR-005, research R5) — the
  operator checkpoint artifact exists IN the spec. [AC #4]
- **SC-004**: prophecy door rejection table passes (already-true,
  duplicate, TTL bounds, dead/unknown target, cap, text cap); terminal
  races land exactly one status; late verification mints nothing.
  [AC #2]
- **SC-005**: `TestCatalogSweep` passes with four new rows; the
  lessons taxonomy test asserts `first-faith-event` presence; the
  rubric-hygiene sweep bans `faith.*`; strip render tests cover
  populated/dashed/no-forecast states;
  `node scripts/check-tui-design.mjs --changed` exits 0 with
  `panels/guardian-strip.md` amended in-branch.
- **SC-006**: the merge-drift pr gate exits 0 from the worktree (wiki
  re-pins + `docs/player/` regenerated in-branch); the PR merges with
  `gh pr merge --merge`; spec-bridge sync moves the board. [AC #1 —
  the link itself is the orchestrator's act, recorded on the card]

## Assumptions

- **Delta and band constants are normative defaults**, expected to move
  on play evidence: +8/−4/−6/+12/−15 and the 4h/6h/12h/24h bands were
  chosen so (a) the genesis band reproduces today exactly, (b) a
  handful of fulfilled directives (~3) climbs one band, (c) a false
  prophecy costs more than a true one earns (declaring cheap prophecies
  is negative-EV unless the guardian actually shepherds them true), and
  (d) a single bad night (2 deaths) does not cross two bands. All live
  in one promoted-dial-ready table; flagged for operator review.
- **Prophecy costs 1 charge** (the `send_vision` parity): it reaches
  minds like a vision and the wager needs a stake; the plan verbs stay
  free (spec 084 R3 unchanged). Flagged for operator review.
- **No prophecy cancel**: genre-honest (a false prophet cannot unsay
  the word) and it keeps the faith wager binding; a mis-declared
  prophecy is survivable (bounded loss, cap 3). Flagged for operator
  review.
- **Posture keys on scenario presence** (`s.scenario != nil`,
  boot-frozen, replay-re-armed — the spec-054 incident-sweep
  precedent), not on curriculum stage: stages govern UI surfaces and
  tool grants, while the spiral question is about whether the world is
  run-shaped. Recorded with the reversal lever in FR-005.
- **Stage availability**: `prophesy` follows `send_vision`'s stage
  profile (granted where visions are granted) — prophecy is the same
  influence verb with a wager. Flagged for operator review.
- **No faith→villager feedback in v1**: villagers learn of prophecies
  through omen/report memories (existing provenance machinery); the
  score itself steers nothing villager-side. The obvious future
  (morale/belief coupling, guardian-subject rumors) is recorded as
  deferred, not smuggled in.
- Wiki notes pinning touched sources re-pin in-branch per the pr gate
  (expected set in plan.md); the gate is the authority.
