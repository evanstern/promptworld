# Research: faith-driven charge regeneration (spec 085)

Decisions with evidence, verified against the task worktree
(`.worktrees/task-118`, branch `task-118-faith-regen` at current main
including TASK-157's merged directives layer). Format: Decision /
Rationale / Evidence / Alternatives considered.

## R1 — Faith is event-sourced state moved ONLY by recorded `faith.changed` events; old events never mint retroactively

**Decision**: `State.Faith *FaithState` (nil = genesis 50, the
nil-means-default accessor pattern) is folded exclusively from a new
executor-emitted `faith.changed` event. The fold never reads
`directive.fulfilled` (or any other pre-existing type) directly — the
executor's faith sweep OBSERVES those in the live batch and emits the
faith event beside them; the reducer folds only the faith event.

**Rationale**: changing an existing event's reducer semantics (e.g.
making the `directive.fulfilled` arm also bump faith) would make pre-085
logs replay to different state — the compatibility law spec 084 records
("no old event's semantics change",
`specs/084-guardian-directives/contracts/events.md` §4). Emitting a
sibling event keeps old logs byte-identical, gives faith its own recorded
paper trail (chronicle rows, the lesson trigger, the log-checkable AC2),
and satisfies the card's phrasing literally: every source is a recorded
event folded by the reducer.

**Evidence**:
- The compatibility law: `specs/084-guardian-directives/contracts/events.md:143-151`
  (§4: pre-084 logs replay byte-identically "since no old event's
  semantics change").
- nil-means-default state: `docs/wiki/world-tuning.md` — `State.Tuning
  *TuningState` (`omitempty`), "nil means the default set", nil-safe
  accessors as the ONLY consumption path, no format bump (the spec-044
  `MorgueEpilogues` precedent).
- Batch-scanning sweep precedent: `internal/sim/executor.go:424-446`
  (run-end detection — "a pure function of (pre-tick state, this batch)",
  scanning the batch for `agent.died`) and `:401-411` (scenario rubric,
  same idiom, "plus THIS batch").
- The seam this consumes: `internal/sim/executor.go:100-113`
  (directive fulfillment/expiry sweep), `internal/sim/plans.go:112`
  ("THE TASK-118 faith-accounting seam. id dedupes, designation_id names
  what…"), `specs/084-guardian-directives/contracts/events.md:122-129`
  (§3: "Faith consumes RECORDED events only; this spec adds no faith
  fields, no scoring, no state").

**Alternatives**: fold faith inside existing reducer arms (rejected:
replay-divergent for old logs, above); a derived read-time computation
over the log (rejected: read paths never scan the log — the
`RunEndedPayload` carries deaths precisely "so no consumer ever scans the
log", executor.go:421-423; and derived faith would leave no event for the
lesson/digest/AC2 log-checkability); a whitelisted injected faith event
(rejected: faith must be endogenous — a console-injectable score is a
cheat surface and breaks "no model judgment").

## R2 — The regen check is REPLACED (band cadence), not paralleled; same event, same shape

**Decision**: rewrite the one boundary check
(`internal/sim/executor.go:53-57`) to
`nextTick % FaithRegenCadenceTicks(s.FaithScore(), s.scenario != nil) == 0`,
keeping `metatron.charge_regenerated` and its empty payload unchanged.
`chargeRegenTicks` (6 game hours, `internal/sim/guardian.go:17-19`)
survives as the steady band's cadence value, so a world with no faith
events (genesis score 50 → steady band) regenerates on a byte-identical
schedule.

**Rationale**: the card says regen "becomes a pure reducer function of
village faith state" — becomes, singular. Two parallel regen channels
(clock + faith bonus) would double the invariants (two boundary lattices,
two forecast rules for the strip) for no design gain; the band table
expresses "today's behavior" as one of its rows, which is the cleanest
compat story and the smallest diff. The check's SHAPE (absolute
boundaries, below-cap gate, empty payload) is what the replay guarantee
and the strip forecast both key on — preserving it inherits both.

**Evidence**:
- Today's check: `internal/sim/executor.go:53-57`; constants
  `internal/sim/guardian.go:16-31` (incl. the exported
  `GuardianChargeRegenTicks` "sim stays the single source of truth for
  the doctrine constant; the TUI never carries its own copy").
- Boundary-fire proof shape to clone: `internal/sim/guardian_test.go:144-168`
  (exported-cadence boundary fires, boundary+1 doesn't).
- The strip's forecast derives from the cadence:
  `internal/tui/views.go:2701-2709` and
  `docs/design/tui/panels/guardian-strip.md` §Structure item 2 (omitted at
  full bank — "forecasting an arrival that isn't scheduled would be a
  lie"); a variable cadence therefore requires the wire to carry the
  effective value (R7).
- Reducer arm unchanged: `internal/sim/guardian.go:245-247`
  (`charges++` below cap — cadence-agnostic already).

**Alternatives**: keep 6h clock regen + a separate faith-driven bonus
event (rejected above); regen amount (not cadence) as the faith function
(rejected: cap is 3 — amounts >1 mostly clamp away, and cadence is what
today's shape already parameterizes); continuous accrual (fractional
charge per tick, rejected: new fractional state, no event legibility, and
the absolute-boundary lattice — the replay-friendly part — disappears).

## R3 — Source vocabulary: directives primary; deaths and prophecy outcomes; designation fulfillment deliberately excluded

**Decision**: five reasons, all executor-observed:
`directive_fulfilled` +8 (primary), `directive_expired` −4,
`villager_died` −6, `prophecy_fulfilled` +12, `prophecy_failed` −15.
`designation.fulfilled` alone mints nothing. No ambient accrual, no time
decay.

**Rationale**: the operator realignment names fulfilled directives THE
endogenous source ("villager compliance with the guardian's directives
closes the god-game mana loop"). Expiry is its honest negative (the
guardian's charge went unachieved — mild: −4, half a death). Deaths are
the genre's population coupling ("power… derived from the size and
prosperity of their population of worshipers") and the deliberate feeder
of the designed spiral. Prophecies are the risk pair: the asymmetry
(+12/−15) makes claim-spam negative-EV unless the guardian shepherds
claims true — the learning loop the card wants ("better prompting →
truer prophecies"). A bare designation fulfilling without a directive is
villager initiative the guardian never staked words on — crediting it
would mint faith for watching, and double-count when a directive IS
bound (the directive companion already fires). Time decay is excluded
because it is ambient accrual with a minus sign: the card's "no ambient
accrual" cuts both ways, and the negative pressure deaths/expiries
provide is already event-shaped.

**Evidence**:
- Realignment: TASK-118 Implementation Notes ("FULFILLED DIRECTIVES are
  the natural endogenous faith source…", board card, 2026-07-26).
- Mana doctrine:
  `research/Game-Gameplay-Patterns/Indirect-Control-and-Divine-Intervention.md`
  §"The mana economy" ("power or mana derived from the size and
  prosperity of their population of worshipers… a positive feedback
  loop").
- The seam payload: `specs/084-guardian-directives/contracts/events.md:82-91`
  (`directive.fulfilled{id, designation_id, targets, issued_tick}` — id
  dedupes, targets attribute).
- `agent.died` is executor-emitted in the same batch the faith sweep
  scans: `internal/sim/executor.go:200-219`.

**Alternatives**: faith from `metatron.nudged` visions being "believed"
(rejected: unverifiable without model grading — the exact thing AC2
forbids); faith from tutoring/conversation (rejected: TASK-112 AC #6
legislates the tutor lane "earns no faith"; FR-012 enforces);
per-target faith attribution (rejected for v1: faith is village-level
state; per-villager devotion is a future coupling, recorded deferred).

## R4 — The regen curve is a four-band step function, not a smooth formula

**Decision**: bands ≥75 → 4h, 40–74 → 6h, 15–39 → 12h, <15 → posture
fork (R5). Step function, exported as one pure
`FaithRegenCadenceTicks(score int, scenario bool) int64`; no hysteresis
in v1 (watch item).

**Rationale**: absolute-boundary regen (`tick % cadence == 0`) needs a
cadence that is a stable small set of divisors of the game day — a
smooth formula would produce cadences whose boundaries drift and whose
strip forecast is unexplainable in one glance. Bands also make the
posture fork (one row) and the compat claim ("genesis row == today")
legible, and they render in fiction naturally (fervent / steady /
wavering / forsaken). Band flap on a boundary score is deterministic
and bounded by the coarse deltas (±4..15 vs band widths 25+); if play
evidence shows flap, hysteresis is a one-function change (the
spec-059 survival-band re-arm precedent shows the shape).

**Evidence**: absolute boundaries as the doctrine:
`internal/sim/executor.go:53-54` ("absolute 6-game-hour boundaries, pure
function of (state, tick)"); the survival bands as the
named-constants-one-home, promoted-dial-ready-but-not-tuning discipline:
`internal/sim/guardian.go:88-105` ("dials are earned by evidence").

**Alternatives**: linear cadence interpolation (rejected above);
faith-priced charge PURCHASES (spend faith for a charge — rejected:
makes faith a currency, which drifts toward the badge/points surface the
overjustification caution forbids, and empties the stock the regen curve
needs to read).

## R5 — Spiral posture: scenario worlds spiral (cadence 0), ambient worlds floor (24h); the lever is one table row (the AC #4 decision)

**Decision**: the forsaken band (<15) regenerates NOTHING in a scenario
world and once per 24 game hours in an ambient world, keyed on the
boot-frozen `s.scenario != nil`. The reversal lever: the entire posture
is one row of the band table plus the fork argument — flipping either
posture, or promoting `faith_floor_cadence_ticks` to `tuning.json`
(0 = no floor), is a one-table change with no event or shape impact.

**Rationale**: the Hades grounding cuts exactly along this line.
Scenario worlds are run-shaped: they have rubrics, pass/fail
boundaries, `run.ended`, and a morgue that teaches — there, "dying in
this game and looping through it over and over is a really important
part of the experience" (Kasavin), and a defanged spiral would remove
the genre's honest signal (the traditional-roguelike position: mastery
needs consequences). The ambient world is the opposite case Hades' God
Mode was built for — "take the sting of failure and reduce that as much
as possible" where the experience must keep moving forward: a
persistent homeworld has no run-reset, so a zero-regen deadlock is not
a dramatic death but a dead save. The floor keeps the spiral FELT
(24h ≈ one-quarter of baseline income) without the hard lock. Two
additional guards make the ambient floor sufficient rather than
generous: the plan verbs are charge-free (spec 084 research R3), so
directive-earned faith is reachable with an empty bank — the endogenous
exit — and deaths (the main faith drain) slow as the village stabilizes.

**Evidence**:
- `research/Learning-Game-Design/Meta-Progression-and-Failure.md` §Hades
  ("how to take the sting of failure and reduce that as much as
  possible"; God Mode preserves the loop while easing it; "difficulty
  settings may need to be proprietary to the game") and
  §"The skill-dilution debate" (both poles recorded — the scenario
  posture honors the traditionalist pole, the ambient floor the Hades
  pole).
- Posture key is boot-frozen and replay-safe:
  `internal/sim/state.go:214-222` (unexported `scenario`, re-armed at
  boot, "a world with no scenario…"), `internal/sim/executor.go:123-125`
  (the incident sweep already forks on `s.scenario != nil` with the
  byte-identical-ambient contract).
- Charge-free plan verbs as the recovery valve:
  `specs/084-guardian-directives/research.md` R3.
- The board card names the tradeoff verbatim: "genre-authentic and
  roguelike-appropriate in scenarios, but the ambient world may want a
  floor" — this decision answers it in the direction the card leans,
  with the lever recorded.

**Alternatives**: floor everywhere (rejected: scenario failure loses
teeth exactly where the curriculum wants stakes — spec 054's rubric
worlds are built to be failable); spiral everywhere (rejected: ambient
deadlock is a dead save, not a lesson — and unlike Hades there is no
"next run" to carry the lesson to); faith floor instead of regen floor
(clamping the SCORE at, say, 15 in ambient — rejected: it falsifies the
ledger to soften the consequence; the honest record with a softened
consequence keeps the artifact truthful, which is this repo's whole
doctrine); keying posture on curriculum stage (rejected: stages gate UI
and tool grants, not world shape; a stage-6 player in an ambient world
still has no run-reset).

## R6 — The verification rule: truth = a pre-declared, closed-vocabulary claim satisfied by recorded state within its deadline; free text is counsel

**Decision**: prophecy verification is structural. A `prophesy` call
records a discriminated `Claim` (four kinds v1:
`designation_fulfilled`, `structure_count`, `population_at_least`,
`survives`) plus a deadline; the executor sweep judges fulfil/fail
conditions per kind, pure over (state, tick), the
`designationFulfilled` idiom; the reducer re-validates at the door.
Free-text visions (`send_vision`) remain what they are — counsel — and
never mint faith. A claim already true at declaration is refused
(prophesying the past), as is a duplicate of an active claim.

**Rationale**: "what makes a vision 'true'" has exactly one answer that
is checkable from the log and never model-graded: the vision must carry
a machine-checkable claim, declared BEFORE the fact, judged by the same
pure-predicate machinery that already judges designation fulfillment.
Any text-matching alternative smuggles a model (or a brittle heuristic)
into deterministic space. The belief-provenance layer completes the
loop villager-side without new machinery: targets receive the declared
word as `OriginOmen` memories (direct perception — a delivered omen is
the one secondhand-looking thing spec 030 classifies as direct), and
the terminal's word-spreads memories are `OriginReport` (honestly
secondhand), so the nightly provenance gate lets villagers form
correctly-provenanced beliefs about the guardian's word from the same
recorded events faith folds — the same evidence, two readers, zero
model grading in either.

**Evidence**:
- The pure-predicate idiom to reuse: `internal/sim/executor.go:75-90`
  (designation-fulfillment sweep — "a pure function of (state, tick)…
  Emitted once… Replay reproduces it deterministically with no guardian
  running") and the fulfil-before-fail terminal ordering
  `:92-113` ("fulfillment is checked BEFORE expiry so a directive
  eligible for both at one boundary lands exactly ONE terminal").
- Origin taxonomy + the direct-perception classifier:
  `docs/wiki/agent-memory-window.md` ("`OriginOmen`… delivered
  omen/dream are direct perception; a… report… secondhand — the
  conservative default"); the provenance gate:
  `docs/wiki/nightly-consolidation.md` (spec 030 `enforceProvenance` —
  origin-only, "no text inspection, no heuristics").
- The vision-companion atomic batch shape:
  `specs/084-guardian-directives/contracts/events.md:56-63`
  (`directive.issued` + companion `agent.memory_added` per target,
  "`Origin: OriginOmen`-class provenance, salience at the dream band").
- Text cap single source: `internal/sim/guardian.go:38-41`
  (`NudgeTextMax` reads the registry's `TextCapBytes` — reuse, never a
  second literal).

**Alternatives**: model-graded truth (the guardian or a judge model
scores whether events "match" the vision text — rejected outright: AC2
forbids it, it is non-replayable, and it would put an LLM inside the
reducer's trust boundary); event-pattern claims ("an event of type T
will occur") matched by the REDUCER while folding other events
(rejected: every unrelated arm would grow prophecy checks — cross-arm
coupling; state predicates keep verification in one sweep);
villager-belief-derived faith (faith = aggregate belief confidence about
the guardian — rejected for v1: beliefs are model-authored nightly, so
faith would inherit model judgment and nondeterminism; recorded as the
long-term fiction-deepening direction the provenance machinery leaves
open).

## R7 — The strip contract is satisfied by two wire fields; the dashed state is the older-daemon case

**Decision**: `ClockStatus` gains `faith *int` (`json:"faith,omitempty"`)
and `faith_regen_ticks int64` (effective cadence, 0 = none scheduled).
The strip renders `faith N` from a non-nil pointer and `faith —` when
nil; the regen forecast uses the wire cadence (legacy exported constant
as fallback), omitted at full bank OR cadence 0.

**Rationale**: §4 pre-specifies three states — absent (pre-085: already
shipped as absence), present-dashed ("field exists on the wire but
before this strip has anything meaningful to show"), populated. Once
this spec ships, the daemon always knows the score (nil-safe accessor),
so the honest dashed case is version skew: a TUI carrying the faith
renderer against a daemon that doesn't serve the field — pointer-nil
detects exactly that, with no invented zeros (the strip's standing
honesty rule). The forecast must move to the wire because the TUI's
compiled-in constant can no longer predict a faith-dependent boundary
(the guardian-strip note's own rule: "sim stays the single source of
truth… the TUI never carries its own copy"); cadence 0 generalizes the
existing omitted-at-full-bank honesty ruling (R4.1) to
no-regen-scheduled.

**Evidence**: `docs/design/tui/panels/guardian-strip.md` §Structure
item 4 + control table row "faith segment"; the renderer + drop order:
`internal/tui/views.go:2687-2750` (faith named first in the truncation
drop order comment already); the absence pin to flip:
`internal/tui/render_test.go:599-600`; wire home:
`internal/ipc/protocol.go:213-215` (`metatron_charges` beside it); D1
projection rule: guardian-strip.md §"Linear-stream / CLI projection".

**Alternatives**: TUI computes cadence from a replicated faith score +
compiled band table (rejected: duplicates the doctrine table across
processes — the exact drift the exported-constant comment warns
about); rendering band words instead of the number (deferred to the
design page amendment — the spec fixes the data, the design page fixes
the glyphs; `faith N` is the v1 form recorded there).

## R8 — Prophecies are uncancellable, capped at 3, charge-priced, and never expire by target death

**Decision**: no `cancel_prophecy` tool; `GuardianProphecyCap` 3 active;
`prophesy` is `Gate: Charge` (1); a prophecy whose targets all die stays
active and is judged against the world.

**Rationale**: cancellation would let the guardian retract a wager the
moment it looks bad, collapsing the risk half of the economy;
"the word, once given, stands" is both the genre truth and the
anti-gaming rule. The cap and the charge bound spam independently of
the −15 penalty. Target death does not void the claim because the claim
predicates are world-state facts (a shelter either stands by dawn or it
doesn't, whoever heard the foretelling) — unlike a directive, whose
all-dead expiry exists because no executor remains
(`internal/sim/executor.go:97-99`, "the un-executable clause"); a
prophecy needs no executor.

**Evidence**: directive cap/TTL discipline to clone:
`specs/084-guardian-directives/contracts/events.md:64-70`; charge-gated
influence precedent: `specs/084-guardian-directives/research.md` R3
evidence (`send_vision`/`send_omen` `Gate: Charge` — "influence on minds
via prose is priced").

**Alternatives**: free prophecies (rejected: with +12 on fulfil and no
stake, declaring every near-certain outcome is a faith pump; the charge
makes each declaration compete with a miracle/vision use); TTL-free
prophecies (rejected: an undated claim is unverifiable — "within its
deadline" is half the truth rule); expiring on all-targets-dead
(rejected above).

## R9 — Faith and prophecy join the existing gate surfaces; the lesson and the rubric ban are in-scope obligations

**Decision**: four digest-grammar rows (`TestCatalogSweep`);
`prophecy.*` (3) join `observableEventTypes`, `faith.changed` does not
(v1); `first-faith-event` joins `lessonCatalog` with the absence-pin
test flipped; `faith.` joins the rubric-hygiene banned prefixes.

**Rationale**: each is a recorded obligation, not a choice: the digest
sweep is the standing catalog gate; observability follows the
directive precedent (enum-only, zero trigger code) and gives the
standing-order workaround loop its prophecy verbs; the lesson is the
spec-077 FR-020 rider naming this task; the rubric ban is written into
the hygiene test's own comment as "the research R2 obligation, recorded
where it will be enforced." `faith.changed` stays un-observable in v1
because the strip already surfaces it continuously and widening later
is compatible.

**Evidence**: `internal/tui/lessons.go:236-237` + 
`internal/tui/lessons_test.go:119-122` (the deliberate absence, "the
entry rides TASK-118"); `internal/sim/rubric_hygiene_test.go:22-24`
("when a faith field/event exists, its type joins the banned set
here"); observable growth precedent:
`specs/084-guardian-directives/contracts/events.md:104-121`
(§2, "enum-only… zero-new-trigger-code guarantee is structural");
lesson row shape: `internal/tui/lessons.go:146-156`
(`first-charge-regen` — the sibling this lesson sits beside).

**Alternatives**: none material; the surfaces are gates.
