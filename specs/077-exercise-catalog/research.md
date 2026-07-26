# Research: Exercise catalog wave (spec 077)

Decisions with alternatives considered, grounded in the shipped code
(`internal/sim/scenario.go`, `curriculum.go`, `gru.go`, `internal/tui/lessons.go`,
`grammar.go`/`digest.go`) and the wiki corpus (`scenario-machinery`,
`curriculum-ladder-progression`, `gru`, `deterministic-rng`, `tui-chronicle-feed`,
`executor-tick-subsystems`, `report-card-renderer`).

## R1 — What "ambient-indistinguishable" means for kinds with no ambient path yet

**Decision**: indistinguishability is a property of the RECORDED ARTIFACT and the
PRECONDITIONS, not of a co-shipped dice path. Each new kind's payload carries no
authored/scenario marker, and its emission-time preconditions are named predicates an
ambient emitter can call verbatim. The ambient dice paths themselves (and their
`gruScheduledTonight`-style preemption twins) are TASK-28's recorded scope.

**Why**: for `gru_emerges` the ambient path pre-existed and the incident reused its exact
shape. For the three new kinds no ambient cause exists yet; shipping ambient rolls here
would change every existing world's ambient behavior inside a content wave — a doctrine
change the reorientation assigned to TASK-28 ("a seeded cold snap is both" — move #11).

**Alternatives rejected**:
- *Ship ambient rolls behind tuning dials defaulting to 0*: adds three tuning dials the
  spec-048 promotion discipline says must each "earn evidence"; dead config until TASK-28
  anyway.
- *Ship ambient rolls on by default (the gru's 600‰ posture)*: silently changes ambient
  worlds; outside decision 5's mandate.

## R2 — Cold snap: state latch + read-time expiry, no end event

**Decision**: `sim.cold_snap{night, until_tick}` latches `State.ColdSnapUntil int64`
(`omitempty`). The needs heartbeat reads a harsher outdoor-night warmth-loss rate while
`tick < ColdSnapUntil`; expiry is a read-time comparison (the
`sim.EffectiveConfidence` read-time-decay precedent) — no `cold_snap_ended` event.

**Why**: one event, one latch, replay-trivial; an end event would be a second thing to
keep consistent and adds nothing a comparison doesn't. The severity mechanism reuses the
EXISTING night-cold arithmetic (`warmthLossCold`, `internal/sim/agents.go:687`) — a cold
snap is "a colder night," which is precisely the ambient-indistinguishability claim at
the level villagers experience.

**Rebase**: `ColdSnapUntil` classifies SHIFT (a time snap preserves the snap's remaining
duration, matching how fire fuel and order expiries shift).

**Alternatives rejected**: paired start/end events (two latches, no gain); a multiplier
field on state (config-shaped state; the rate is doctrine, the window is the event).

## R3 — Forage blight rides the existing Harvest overlay

**Decision**: `sim.forage_blighted{x, y, radius, tiles, regrow_tick}` — one merged event
per firing (the `sim.food_rotted` one-event-per-sweep precedent), tiles enumerated in
deterministic walk order (row-major over the patch, the fixed-neighbor-order house
style); the reducer appends each tile to the EXISTING `Harvests` overlay
(`Harvest{X, Y, Regrow}`, `internal/sim/agents.go:661-665`) with the far `regrow_tick`.

**Why**: maximal indistinguishability for minimal machinery — a blighted tile IS a
harvested tile with a long regrow, a state the world already produces via heavy picking;
perception, mental-map correction (`agent.map_corrected` fires naturally when a villager
sees the barren patch), regrowth (`sim.forage_regrown`), and the miracle rebase taxonomy
(`Harvest.Regrow` already classified) all work unchanged.

**Alternatives rejected**: a new tile kind / terrain overlay (touches worldmap, tile
registry, TUI legend, migrations — massive surface for the same experience); per-tile
events (floods the log; food_rotted precedent merges).

## R4 — Stranger is an entity, not a phenomenon (the gru precedent)

**Decision**: `State.Stranger` (nil when not abroad) mirroring the `Gru` struct's shape:
`stranger.arrived{night, x, y}` / `stranger.moved{x, y}` /
`stranger.took{x, y, kind, n}` / `stranger.departed{day}`. Behavior: greedy movement
toward the nearest unattended store (ground pile or chest with no living agent adjacent),
one bounded take per visit-cooldown, absolute avoidance of fire light and shelter tiles
(the gru's `gruLightRadius`/shelter rules reused as shared predicates), departure at the
next dawn. Takes decrement pile/chest contents through the same state mutations agent
withdrawal uses and append to a ring-bounded `State.StrangerTakes` ledger (retain 32, the
`guardianOrderRetain` precedent) — the ledger is what zero-wanted rubric terms
("nothing is taken") count.

**Why**: sight-triggering, TUI rendering, rumor material, and rubric counting all need a
positioned body and a durable trace — exactly why the gru is an entity
(`docs/wiki/gru.md`). Witness memories reuse the gru's situated-memory idiom so a seen
theft becomes village rumor fuel (story, not just pressure).

**RNG**: all stranger decisions via `rngAt` with new purpose tags (`"stranger-prowl"`,
`"stranger-take"`) — per-decision PCG, no stream, per `deterministic-rng`'s standing
instruction for new systems.

**Rebase**: movement/take cooldown tick fields SHIFT; `StrangerTakes[i].Tick` KEEPs
(historical fact, the ledger precedent).

**Alternatives rejected**: an off-screen "overnight theft" event with no entity (no
sight/rumor/render; indistinguishable from a bug to the player); a ninth villager agent
(genesis is eight write-once personas — a mid-run agent violates the persona substrate).

## R5 — Severity grammar entry: existing families and tiers only

**Decision**: `familyByNamespace` gains `"stranger"` → the gru/threat family voice;
`sim.cold_snap`/`sim.forage_blighted` ride the existing `sim` family. `stranger.took`
joins the four-type whole-line alert set (beside `social.chest_taken` — theft is the
established alert semantics); every other new type renders at normal tint. No new tier,
channel, or push surface (decision 5's "small stable vocabulary extends to event
vocabulary" rider; the TASK-133 move's "no new channel" doctrine).

## R6 — Boundary generalization: a two-value vocabulary on ExerciseDefinition

**Decision**: `ExerciseDefinition` gains `BoundaryDay int` — `N > 0`: evaluate exactly at
dawn of day N (the shipped first-night semantics, generalized); `0`: rolling — evaluate
at EVERY dawn from day 2 until a pass lands. `scenarioRubricEvents` drops its
`first-night`-only guard and keeps every shipped guard (batch death scan,
`hasCurriculumPass` latch, pass→unlock same-batch order) for all exercises.

**Why**: dawn is already the boundary the emitter, `sim.day_started`, and the narrator's
chapter cadence share; two boundary shapes cover all nine exercises. A fixed-boundary
miss still never emits failure (`run.ended` stays the only fail signal) — shipped
semantics, now stated as the general contract.

**Alternatives rejected**: per-exercise closure fields on the definition (content must
stay data-shaped — closures don't serialize into docs/tests as content); arbitrary
boundary times (nothing needs sub-dawn resolution; dawn keeps the all-dead-dawn guard
meaningful).

## R7 — the-law evidence: persist the observation coordinates, keep the constructor rule

**Decision**: the `metatron.charter_observed` reducer arm additionally stamps
`State.CharterObservedSeq = e.Seq` / `CharterObservedTick = e.Tick` (the
`GuardianOrder.PlacedSeq` apply-time-stamp precedent, `internal/sim/guardian.go:393`).
A new state-sourced constructor (`CharterEvidenceFromState(s)`) builds the
`EvidenceRef{Type, Seq, Tick, Custom: s.CharterCustom}` — the honesty derivation stays
in sanctioned-constructor land; `CharterObservedEvidence` (event-sourced) remains for
fixtures.

**Why**: spec 072 FR-009 named exactly this blocker ("CharterObservedEvidence needs the
recorded event's Seq/Tick, which state does not retain"). Persisting two ints beside the
already-persisted fingerprint+custom is the minimal, precedented fix. Pre-077 snapshots
have zero Seq → the evidence entry is omitted until the next observation stamps it
(honest, self-healing; documented edge case).

## R8 — Stage-3 evidence: `metatron.skills_observed`, custom by construction

**Decision**: a new event `metatron.skills_observed{fingerprint, names}` emitted by the
guardian's turn-time observation pipeline exactly like `charter_observed` (emit on
fingerprint change of the BOUND skill set; stage-3+ only, since `stageSkills` refuses to
bind earlier); reducer persists `SkillsFingerprint/SkillsObservedSeq/SkillsObservedTick`.
`SkillsObservedEvidence` derives `Custom: true` — no game-shipped skill files exist and
stages 1–2 lock binding out, so a bound set is player-authored by construction.

**Why**: `EvaluateUnlock`'s stage-3 conjunct ("a player-granted tool's contributing act —
which tool is TASK-119's exercise design") has been an open design slot since spec 046;
skill files are stage-3's own concept ("you shape what it can do"), the observation
pipeline is precedented, and the same event powers the `first-skill-file` lesson (US3) —
one mechanism, three consumers.

**Alternatives rejected**: deriving evidence from `cog.tool_call` landings of
bundle-origin tools (the reducer cannot know a tool's bundle origin — boot-frozen config,
not state; would need a tool-provenance whitelist in state, far heavier); a
capabilities.json observation event (capabilities NARROW the grant — narrowing is not
"granting", and an empty capabilities file would assert authorship falsely).

## R9 — Exercise inventory: 3/2/2/2, seeds 46101–46109

**Decision**: nine exercises (data-model.md table is normative): stage-1 keeps
`first-night` and adds `cold-dawn` + `stranger-at-the-gate` (the entry stage gets three —
one per pressure kind a new player meets); stages 2–4 get two each (`the-law` +
`blighted-larder`; `toolsmith` + `fog-watch`; `long-winter` + `stewards-charge`).
Every rubric term is a cataloged event type; every Met derivation is a state fact
already maintained or added by this spec; boundaries dawn-shaped per R6.

**Why 2–3 not more**: the AC's own band; the rubric-gauge exposure doctrine
(reorient parked item: "all-terms-live stands until multi-exercise content exists") gets
its first real test at this size; hand-authoring quality over quantity is decision 5's
explicit v1 posture.

## R10 — The wrong-thing detector needs one bounded fold, not a trigger redesign

**Decision**: `lessonTriggers` gains a small session-local fold: a capped
`map[reason]count` over `cog.tool_call` events with `rejected_*` verdicts and non-empty
`Reason`; `same-refusal-pattern` fires when any reason's count reaches 3. Implemented as
an optional fold-trigger seam on `lessonEntry` so the existing pure per-event `Trigger`
predicates stay untouched.

**Why**: "repeated same-cause tool rejections" is inherently cross-event; the projection
already folds state for decision traces (`decisionTraces` precedent) so a bounded fold is
house style. Reason strings come from the tool loop's fixed rejection vocabulary
(`rejected_gate`/`rejected_cardinality`/... + reason text), so cardinality is naturally
small; the cap is belt.

**Alternatives rejected**: a daemon-side detector event (violates spec 055's
client-side-only lesson doctrine — no new event types for teaching); windowed/decaying
counts (three-strikes-ever is the simplest honest v1; tune on playtest evidence, the
retry-accommodation parked item's process).

## R11 — Lesson triggers for the tranche (verified against payloads)

- `first-explain-answer`: `cog.tool_call` with `Tool == "explain"` and verdict `read_ok`
  (`CogToolCallPayload`, `internal/sim/cognition.go:123-135` — explain is read-only, so
  read_ok is its landed verdict).
- `first-report-card`: `guardian.report_card` (type exists, digest row exists).
- `first-skill-file`: `metatron.skills_observed` (R8's new type).
- `same-refusal-pattern`: R10's fold.
- `first-faith-event`: no event type exists (TASK-118 unrun) — deferred with a board
  rider on TASK-118, never stubbed (spec FR-020).

## R12 — What this spec deliberately does NOT touch

No tuning dials; no IPC/status shape changes (`ScenarioExercise/Outcome` already carry
everything); no new dock tabs or takeovers; no report-card surface changes (spec 072's
resolver consumes new evaluators for free); no `format_version` bump (all new state
`omitempty`); no changes to `EvaluateUnlock`'s conjuncts (the stage-3 gate was always
"any Custom evidence" — this spec supplies the evidence, not a new gate).
