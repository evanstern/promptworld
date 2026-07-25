---
name: sim-state-reducer
description: sim.State and Apply — the single event-driven mutation path used identically live and in replay; canonical JSON for hashing
kind: component
sources:
  - internal/sim/state.go
  - internal/sim/agents.go
  - internal/sim/recipes.go
  - internal/sim/miracles.go
  - internal/sim/journal.go
  - internal/sim/terrain.go
  - internal/sim/morgue.go
  - internal/sim/curriculum.go
verified_against: cea7b8f83fa07f9fcfefe4dd861aa05a78448f1b
---

# Sim state & reducer

`sim.State` is the whole world in one struct: clock state (tick, paused, speed,
degraded, effective rate — plus, since spec 028, a `RequestedSpeed
clock.Speed` `omitempty` sitting beside `Speed`: the player's ceiling, present
only while the adaptive-throttle governor holds `Speed` below it; `Speed`
itself keeps its pre-028 meaning as the loop's pacing speed, now specifically
the EFFECTIVE speed, so the router and auto-slow observer need no change)
plus the living world — agents with needs/intents/
inventories (the v2 resource set, spec 012 — [[executor]]; spec 032 US2 adds
`Axes []int`, a `Spears` clone — remaining harvest uses per carried axe, sorted
ascending, tripling chop/quarry yield)/memories (with
`IdleSince` for the reflex grace, a `NearDeath`
latch, a `Generation` interrupt counter and pending `Plan` steps for the
[[cognition]] horizon — both `omitempty` so pre-TASK-32 snapshots stay
byte-stable — plus, since spec 019 US3, a self-authored `Journal *Journal`
(`journal.go`): a durable per-agent notebook mutated ONLY by the two `journal.*`
Apply arms; an `omitempty` pointer on the Hail precedent, so a never-journaling
agent stays byte-identical to a pre-019 snapshot; each `Memory` also now carries
`omitempty` situated context `Where`/`Why`/`Conv`/`Origin` (spec 030's
closed-vocabulary provenance class — the ONLY signal `DirectPerception`,
`memory.go`, reads to classify direct perception), byte-stable when absent,
and — since spec 042 — `omitempty` `Seq`/`Vec`/`VecModel`: `Seq` is the
emitting `agent.memory_added` event's store seq, copied onto the appended
`Memory` at Apply time as its stable identity (unique where `(agent, tick)`
is not) — live, [[sim-loop]]'s `stampSeqs` pre-assigns each batch's
contiguous seqs before `Apply` runs, since `AppendEvents` otherwise only
assigns them inside its own append transaction; on replay the events already
carry their seqs from the log, so both paths land the identical value;
`Vec`/`VecModel` are a recorded embedding attached verbatim by the new
`agent.memory_embedded` arm below, nil `Vec` meaning vectorless
([[memory-retrieval]]);
plus, since spec 041, a private spatial-knowledge store, `Map *MentalMap`
(`omitempty`, the Journal/Hail pointer precedent — a never-mapped agent, i.e.
every pre-041 snapshot, round-trips byte-identically; [[mental-maps]] owns
the type and its two knowledge-event Apply arms below), plus, since spec 042,
a rolling situation (query) vector `SitVec`/`SitVecModel`/`SitVecTick`
(`omitempty`) the reducer sets verbatim from an `agent.situation_embedded`
companion — absent (nil `SitVec`) leaves selection on the legacy ranking
([[memory-retrieval]]) — plus, since spec 043 ([[decision-context]]), two
reducer-DERIVED self-knowledge surfaces maintained by existing arms with no
new event type: `IntentLog []IntentRecord` (US1, `omitempty`) — the
recent-intent ring, capacity `intentLogCap` (8), each record
`{Goal, Source, Reason, Tick, Outcome, OutcomeTick}` with `Outcome` empty
while executing then `done`/`failed`/`rejected`/`expired` — and
`NeedsAnchor *Needs`/`NeedsAnchorTick` (US2, `omitempty`; a POINTER on the
Journal/Hail precedent, deliberately deviating from the spec's value type so
a pre-043 snapshot round-trips byte-identically) — the trajectory window's
edge snapshot the decision prompt diffs current needs against to render
rising/falling/steady, `NeedsAnchorTick == 0` the unset first-window
sentinel, structures (`fire`/`shelter`/`oven`/`chest`, fires carrying a
`FuelUntil`; chests (spec 013 US3) carrying a permanent `Owner` — the builder's
agent index, zero-value round-tripping unambiguously since every chest has one —
and a `Store *Inventory` capped at `chestCap` via the same derived `bulk()` used
for agents; spec 032 US1 adds two wall kinds, `wall_plank`/`wall_stone`,
carrying `HP` — current health, 1..`wallMaxHP(kind)`, derived from kind and
never stored as a separate max, the fire lit-ness doctrine — plus a non-wall
`path` kind carrying no `HP` and never blocking passage; spec 044 US4 adds a
`grave` kind, reducer-placed only — never player-built — marking a death
tile), cleared trees,
harvested forage, den cooldowns, `Quarried` depleted
rock outcrops (spec 012, permanent, `omitempty`), `Piles []Pile` — the per-tile
ground commons of dropped/spilled goods (spec 013 US2): event-sourced overlay
state like `Quarried`, one pile per tile (a reducer invariant), non-food a flat
count, food batch-tracked (`FoodBatch{Kind, N, SpoilAt}`, same `(Kind,SpoilAt)`
merges), spears AND axes (spec 032 US2, a `Spears` clone) sorted ascending,
`omitempty` — the social
fabric — relation edges, the debt ledger, the rumor registry with per-holder
variants and the bounded conversation-record ring ([[social-fabric]]) — the
consolidated inner life: per-agent beliefs (each now carrying a `Reinforced`
decay-clock anchor since spec 030, `omitempty` — 0 is a legacy grandfather that
never decays; [[nightly-consolidation]] covers the decay curve), self-narrative,
and the
once-per-night consolidation ledger ([[nightly-consolidation]]) — the
[[gru]] (`Gru *Gru`, nil while not abroad; `omitempty` keeps pre-TASK-10
snapshots valid) — and the narrated story: the bounded `State.Chronicle`
ring ([[chronicle]], TASK-11), which rides snapshots so attaching clients
get catch-up history for free — Metatron's charge bank
(`MetatronCharges`, genesis 1, deliberately not `omitempty` so a
spent-to-zero bank round-trips as 0; [[metatron]], TASK-12) — the standing-order
substrate (`MetatronOrders []MetatronOrder`, spec 029, `omitempty` — here an
empty order set genuinely IS the zero value, unlike the charge bank, so a
pre-029 snapshot with the field absent unmarshals to nil; [[metatron-orders]])
— and the village's
law ([[governance]], TASK-13): `MeetingPlace` (set once), the `Meeting`
lifecycle (including the TASK-36 emergent-gathering watch fields
`GatherStart/GatherX/GatherY`), the `MeetingConvention` (TASK-36, nil until a
source establishes it — pre-TASK-36 snapshots load nil, a village with no
standing agreement to meet), and the `Norms` list with monotonic
`NextNormID`/`NextProposalID`, all zero-valued in pre-TASK-13 snapshots (a
lawless village) — and, since spec 044 ([[morgue]]), the run's outcome: the
`Deaths []DeathRecord` ledger (`{agent, tick, cause}`, appended by the
`agent.died` arm in application = event order, bounded by the agent count —
it exists so the run-end emission stays a pure function of (state, batch)
rather than a log scan), the `Ended bool` terminal latch and `RunEnd *RunEnd`
summary (`{tick, deaths, final_cause}`, set once by the `run.ended` arm and
never cleared by any event, so snapshot+replay restores the ended posture on
restart for free), `CharterFingerprint` (the most recent effective-charter
content hash a Metatron turn ran under — the full revision timeline lives in
the event log), and the `MorgueEpilogues []MorgueEpilogue` bounded ring
(`morgueEpilogueCap` 32, the chronicle pattern) of narrator mourning prose —
all `omitempty`, so every pre-044 snapshot stays byte-identical with no
format bump — and, since spec 046 ([[curriculum-ladder]]), the ladder's
world-visible facts: `CurriculumPasses []CurriculumPass` (a bounded ring,
`curriculumPassRetain` 32 on the standing-order prune precedent, of recorded
exercise passes each carrying `EvidenceRef{type, seq, tick, custom}` audit
pointers back into this world's log) and `StagesUnlocked []string` (the
once-per-(world,stage) unlock latch — no cap needed, at most one entry per
ladder stage), both `omitempty` so a pre-046 snapshot with neither field
round-trips byte-identically; the per-user unlocks record is a PROJECTION of
these events, this state being the replayable authority — and, since spec 048
([[world-tuning]]), the effective world-tuning dial set: `Tuning *TuningState`
(`omitempty`, the Journal/Hail/Map pointer precedent — nil means the five
promoted doctrine defaults, set only by the `sim.tuning_applied` arm below, no
`format_version` bump)
(executor types in `agents.go`; memories belong to
[[agent-mind]]). Its
`Apply(event)` method is the **only** event-driven mutation path — the live loop and
crash recovery run the exact same code, which is what makes replay provably equal to
live execution. Spec 012 bumped the save format to v2, and spec 013 (inventory &
storage — bulk cap, piles, chests, theft, rot) bumped it again to **v3**
([[world-save-directory]]); a v1 world's `Inventory` (just `wood`/`food`) cannot
decode under this build at all — [[world-migration]] is the bridge, chaining 1→2→3
in one run and landing as a single wholesale-replace event rather than incremental
`Apply` calls (below).

## How it works

`NewState(seed, m)` is genesis: tick 0 (day 1 06:00), `DefaultSpeed` (4x), eight
named agents on distinct passable tiles via `genesisPlacement` ([[deterministic-rng]]),
with deliberately imperfect needs — day 1 must demand foraging, wood, and a fire
before dark. `genesisPlacement` (spec 012 US6) is factored out so [[world-migration]]
can re-place carried souls on a regenerated v2 map byte-identically to a fresh
genesis of the same seed.

`State` also carries an unexported `m *worldmap.Map` (spec 016): the static
generated map, attached at construction and never serialized (canonical state
bytes are unchanged by it). `SetMap` attaches it to a `State` built outside
`NewState` — the loop's dry-run probe and any replica reconstructed by
unmarshalling into a bare `State` have none until called — so miracle reducer
arms can consult the terrain vocabulary (`passable`/`buildSite`/`effectiveKind`)
identically live, in the dry-run, and in replay. `world.migrated`'s wholesale
`*s = p.State` replacement preserves the receiver's existing map across the
swap (the unmarshalled payload state carries none of its own). `State.MapDims()`
(spec 041) exposes the attached map's `(W, H)` — 0,0 when unattached — so the
mind's prompt renderer ([[agent-mind]]) can size a mental-map bitmap read
without the `State` ever serializing the map itself.

`Apply` switches on event type: `clock.*` maintain pause/speed/degradation —
since spec 028, `clock.speed_set` additionally clears `RequestedSpeed` (a
player command always collapses governed state, FR-009), and two new arms,
`clock.governor_shed`/`clock.governor_recovered`, apply the daemon governor's
ceiling-preserving notch decisions: both set `Speed = to` and follow
`EffectiveRate` from it unless `Degraded`; shed additionally sets
`RequestedSpeed = requested`, recovered sets it only when `to != requested`
(clearing it once the climb reaches the ceiling) — see [[cognition]] for the
governor's decision logic and [[event-types]] for the payload shape;
`sim.night_started`/`sim.day_started` flip `Night` (waking is an explicit
`agent.woke`, never implicit); `sim.forage_regrown` clears a harvest overlay; the
`agent.*` family ([[event-types]]) drives intents (`agent.intent_set` carries a
storage goal's `Kind`/`Qty` onto the `Intent`, spec 013 R4, and also stamps
`Agent.LastGoal`/`LastGoalTick` — spec 015 R1, `omitempty`, written here and
never cleared by any event, so the [[tui-client]] villagers tab can show an
idle villager's most recent objective from any snapshot; since spec 017 the
payload carries `Job` (`omitempty`), the tool-use loop's job id when a
planner-loop landing set it, and since spec 019 the payload's LAST field,
`Reason` (`omitempty`), the planner's free-text reason — copied onto
`Intent.Reason` so it survives to completion, where the executor bakes it into
the memory's `Why`; reflex/executor-authored intents carry neither `omitempty`
tail, so those emissions marshal byte-identically to before; since spec 043
US1 the arm also appends an `IntentRecord` to the agent's ring via
`appendIntent` — source verbatim, oldest dropped past `intentLogCap` into a
fresh backing array so canonical bytes never alias dead capacity; a previous
record still open stays open, so an override reads as open-then-new),
movement, work
products (inventory + overlays + structures), eating (`agent.ate`'s `AtePayload`
sets the absolute post-eat food need and decrements each carried food form by its
consumed count — no reducer-side arithmetic), sleep, talk, needs (absolute
values; since spec 043 US2 the `agent.needs_changed` arm also rolls the
trajectory anchor — once `tick − NeedsAnchorTick ≥ trajectoryWindowTicks`
(1800, one planner cadence) it snapshots the current needs into
`NeedsAnchor`/`NeedsAnchorTick`, so direction is measured over roughly the
last window rather than instant-to-instant noise; on a fresh world the first
window carries no anchor and renders steady), and death; the v2 resource/crafting events (`agent.quarried`/
`collected_water`/`crafted`/`cooked`/`bathed`/`refueled`/`spear_broke`,
`sim.fire_burned_out`) apply inventory deltas and structure/overlay changes,
several by re-deriving the recipe from `recipes.go` (the single source for
craft/build magnitudes — a duplicated number here would drift from the contract
table), and — since spec 013 — clamp their gather yields to the taker's free
bulk (`bulkCap − bulk(Inv)`); since spec 032 US2, `agent.chopped`/
`agent.quarried`'s own yield is no longer a single flat constant — it's
`chopYieldBare`/`quarryYieldBare` (1) bare-handed or `chopYieldAxe`/
`quarryYieldAxe` (3) carrying an axe, re-derived from the SAME pre-mutation
state the emitter checked (the spear-hunt precedent), spending `Axes[0]`'s
last use; a spent-to-zero axe's removal rides its own companion
`agent.axe_broke` (an `agent.spear_broke` clone), not this payload; the v3
storage events (spec 013 US2/US3/US5)
move goods between an agent's `Inv` and a `Pile`/chest `Store`:
`agent.dropped`/`agent.picked_up` create-or-merge or drain a tile's pile
(food oldest-batch-first, spears AND axes (spec 032 US2) most-worn-first),
`agent.deposited`/
`agent.withdrew` do the same against a chest's `Store`, and `sim.food_rotted`
drains a pile's spoiled food batches (`SpoilAt ≤ tick`) — every one of these
defensively re-clamps to what's actually carried/held/available, so the reducer
stays total even against a contested or forged event, and an emptied pile is
removed in the same application; `social.chest_taken` is an effect-free record
(its consequences — the reason-`"theft"` `social.relation_changed` and the
owner/witness `agent.memory_added` events — ride the same companion batch,
[[social-fabric]]); the wall family (spec 032 US1) maintains a standing wall's
`HP`: `agent.built`'s generic arm stamps a fresh `wall_plank`/`wall_stone` at
`HP = wallMaxHP(kind)` — full health, after the same recipe-delta spend every
structure gets — and three dedicated arms carry the multi-cycle demolish/repair
loop: `agent.wall_chipped` decrements `HP` by `demolishChipHP` (clamped to
never go below 1 — a standing wall never serializes ≤ 0) and resets the
demolisher's `Intent.WorkStart` to 0, re-arming the executor's work gate for the
next cycle under the SAME intent (no new scheduling); `agent.wall_destroyed`
(the final chip) removes the structure — its tile is passable again by
construction — and clears the intent (`metatron.entity_removed` reaches the
same end through the miracle path); `agent.wall_repaired` consumes 1 unit of
`wallRepairMaterial(kind)` (planks for a plank wall, refined stone for a stone
wall — the same material each was built from) and restores `HP` by
`repairHPPerUnit`, clamped to the max, re-arming the work gate the same way when
still damaged with material in hand, otherwise clearing the intent —
`isWall`/`wallMaxHP`/`wallAt` (`terrain.go`, a `chestAt` sibling) back every one
of these arms, plus the movement `passable` check a standing wall now fails;
`agent.build_failed` (spec 038, `BuildFailedPayload{agent, goal, reason}`) is a
NEW arm with a state effect identical to `agent.intent_done` — `Intent = nil`,
`IdleSince` stamped — the executor emits it, instead of the bare completion
type, whenever a `build_*` goal's mid-work re-validation genuinely fails (site
gone, or a wall's reserved-tile occupant outlasting the grace period); the
reducer itself carries no build-specific logic, it only clears the intent the
same way completion does, so no material is spent and no structure stands
([[executor]], [[event-types]]); since spec 043 US1 both completion arms also
close the ring: `agent.intent_done` stamps the newest still-open
`IntentRecord` `"done"` and `agent.build_failed` stamps it `"failed"` (via
`stampIntentOutcome` — the newest open record IS the current intent; an older
record left open by an override stays open, the open-then-superseded shape
the alternation view preserves), while `agent.intent_rejected` — formerly a
pure telemetry no-op — now appends an ALREADY-CLOSED `"rejected"` record
(source `planner`), so the next thought can see an attempt was refused before
ever landing;
[[mental-maps]]'s two knowledge-growth arms mutate `Agent.Map` directly:
`agent.saw` upserts the perception sweep's fully-baked facts verbatim
(`Map.upsertFact`), `agent.map_corrected` removes facts the sweep found gone
(`Map.removeFact`) — both no-op on a map-less agent (a pre-041 world mid-
migration), keeping the reducer total; `social.place_told` (the talk
sidecar's directions exchange) and `metatron.place_revealed` (a vision's
optional place grant) route through the existing `applySocial`/`applyMetatron`
dispatchers below, upserting into the RECEIVER's map only where the fact is
absent or the agent's own knowledge is staler. Beside these, several
EXISTING arms gained silent DERIVED bookkeeping with no new event: `agent.moved`,
`agent.woke`, and a `villager`-class `metatron.entity_moved` each call
`markExplored`/`notePresence` — a mover's surroundings become explored
terrain and mover-and-bystanders record each other's positions — a pure
function of (state, event) with no chronicle noise, so a mind-map-populating
step never needs its own event type (research D2);
the `gru.*` family dispatches to
`applyGru` in `gru.go` ([[gru]]);
the `meeting.*`/`norm.*` families — plus `meeting.convention_established` and
the `sim.gathering_observed` watch event (TASK-36) — dispatch to
`applyGovernance` in `governance.go` ([[governance]]); the four miracle types
`metatron.time_snapped`/`metatron.item_granted`/`metatron.entity_moved`/
`metatron.entity_removed` (spec 016, [[metatron-miracles]]) dispatch to
`applyMiracle` in `miracles.go`, alongside `metatron.charge_regenerated`/
`metatron.nudged`'s `applyMetatron` — which since spec 029 also arms the
standing-order lifecycle: `metatron.order_placed` validates and appends (id
uniqueness, origin, non-empty `event_types`, a 1..7-game-day ttl, valid agent
index, condition/action length caps, and — player-origin only — the 3-order
active cap) then prunes to the active set plus the most recent 32 non-active;
`metatron.order_triggered`/`metatron.order_cancelled`/`metatron.order_expired`
each transition one order from active to a terminal status via the shared
`transitionMetatronOrder`, rejecting an unknown id or one not currently active
([[metatron-orders]]); since spec 044 (US2) `applyMetatron` also carries
`metatron.charter_observed`, which validates a non-empty fingerprint (so the
`InjectSocial` dry-run refuses a blank one at the door) then sets
`State.CharterFingerprint` — state keeps only the CURRENT fingerprint, the
full revision timeline being the log's observation sequence the [[morgue]]
aligns each death against. `morgue.epilogue` dispatches to
`applyMorgueEpilogue` in `morgue.go` (spec 044 US2): it validates the agent
index (`-1` = the run-end epilogue) and non-empty text, then appends the
bounded `State.MorgueEpilogues` ring (`morgueEpilogueCap` 32).
The `curriculum.*` pair (spec 046, [[curriculum-ladder]]) dispatches to
`applyCurriculum` in `curriculum.go` — validate-not-clamp, the metatron arm's
contract, since both types are the executor emission class (pure functions of
recorded state, so a landed event always re-applies cleanly in replay while a
malformed fixture is rejected at the door): `curriculum.exercise_passed`
checks a non-empty exercise id and the closed stage vocabulary
(`validLadderStage`, `stage-1`..`stage-4` — the reducer-side twin of
`world.ValidStage`, kept local so the deterministic core never imports the
save-directory package) then appends the bounded pass ring;
`curriculum.stage_unlocked` additionally rejects `stage-1` (the ladder's
unearned floor — only stages 2..4 ever unlock) and any stage already latched
(once per world per stage), and deliberately does NOT cross-check
`CurriculumPasses` — that ring is pruned past 32, so the gate-conjunct
evaluation (`EvaluateUnlock`) happens at emission time, never on re-apply.
`sim.tuning_applied` (spec 048, [[world-tuning]]) is a third arm in this
validate-not-clamp family: the payload is always the FULL effective five-dial
set (never a delta, never re-clamped here — clamping is `ParseTuning`'s job
on the daemon side), so the arm is a pure, idempotent `s.Tuning = &TuningState{...}`
assignment — replay re-applies it identically and the daemon boot seed never
double-counts. `State.Tuning *TuningState` (`omitempty`, no `format_version`
bump) is nil until the first such event; nil reads as the default dial set
through the nil-safe accessors (`RefuelDyingBelow()`, `FireBurnPerWood()`,
`GruEmergePerMille()`, `PlannerCadence()`, `EncounterCooldown()`) every other
promoted call site (`agents.go`'s fire-fuel arm above, [[reflex-policy]],
[[gru]], [[agent-mind]]'s cadence/encounter scheduling) reads through instead
of the retired raw constants.
`world.migrated` (spec 012 US6) is the one case that does not incrementally mutate
fields: after checking the payload's `State.Seed` matches (a mismatched payload
no-ops, keeping `Apply` total), it replaces `*s` wholesale with the embedded state —
[[world-migration]] is the only producer.
`agent.memory_added` copies the payload's situated context — `Where`/`Why`/`Conv`/
`Origin`, all `omitempty` — verbatim onto the appended `Memory` (spec 019/030:
baked at emission, never re-derived at render or replay, so live and replay
agree; `Origin` is the closed-vocabulary provenance class `DirectPerception`
reads to classify direct perception, absent classifying as secondhand, the
conservative default), and
additionally bumps `Agent.Generation` when the memory's
salience is at or above `GenerationBumpSalience` (9) — in-flight thoughts
snapshotted under the old generation are superseded at landing ([[cognition]],
[[sim-loop]]). Two spec 042 arms mutate state from the embedder driver's
companions ([[memory-retrieval]]): `agent.memory_embedded`
(`MemoryEmbeddedPayload{Agent, MemSeq, Vec, Model}`) scans the agent's
memories newest-first for the one whose `Seq` equals `MemSeq` and copies
`Vec`/`Model` onto it verbatim — a zero `MemSeq` never matches (so a pre-042,
seq-less memory can never be mistargeted) and a target that has died or
consolidated away is a deliberate no-op, never an error; `agent.situation_embedded`
(`SituationEmbeddedPayload{Agent, Tick, Text, Vec, Model}`) unconditionally
overwrites the agent's `SitVec`/`SitVecModel`/`SitVecTick` — later events
simply replace earlier ones, no history kept. The journal family (spec 019 US3) is the agent notebook's only
mutation path and, unlike the cognition telemetry types below, does mutate state:
`journal.entry_written` appends a reducer-id'd `JournalEntry` via `appendEntry`,
which enforces the per-agent `journalBudgetRunes` (4000) rune budget INSIDE
`Apply` — the budget participates in the accept/reject decision, so the
`InjectSocial` dry-run turns an over-budget append into a door rejection rather
than trusting handler courtesy (Principle III, the same door-facing gate the
miracle dry-run uses — [[agent-mind]]); `journal.entry_deleted`
removes an entry by id (ids never reused or renumbered, freed runes reclaimable),
a missing id erroring. The budget lives here as a version-stable sim constant,
not config, so a replay can never reject an event that landed live. The plan
family maintains `Agent.Plan`: `agent.plan_set`
replaces the steps, `agent.plan_step_started` pops the head, and
`agent.plan_expired` clears the whole remaining plan (a broken sequence is
not resumed) — and, since spec 043 (FR-005), also stamps the expired step
into the intent ring via `stampOrAppendExpired`: an open record matching the
step's goal closes `"expired"` (goal-matched so a concurrent non-plan intent
is never mis-stamped), otherwise a closed record is appended (the step
expired before ever firing). The hail family (TASK-47) maintains `Agent.Hail *AgentHail`
(`{By, Until}`, `omitempty` so pre-TASK-47 snapshots and un-hailed agents stay
byte-stable): `social.hailed` sets it, `social.hail_met`/`social.hail_expired`
clear it, and `agent.died`/`agent.slept` also clear it (the dead and the
sleeping shed hails). `agent.died` also spills the dying agent's entire carried
`Inv` onto a pile at its own tile (create-or-merge, food batches stamped
`tick + rotWindowTicks`), emptying `Inv` — reducer-internal, no new event (spec
013 US2, FR-006, research R7's debt-opening precedent) — and, since spec 044,
two more reducer-internal effects in the same arm: the death is appended to the
`State.Deaths` ledger (`{agent, tick, cause}` — cause now includes `"gru"`, the
[[gru]]'s escalated kill), and a `Structure{Kind: "grave"}` is placed at the
death tile (US4, FR-017, research R10). The grave placement is deliberately
unconditional — no dedup against whatever already occupies the tile: the
`Structures` slice has no per-tile uniqueness invariant outside the `buildSite`
gate governing NEW player-directed builds, and `structureAt` filters by kind,
so coexisting entries are an established pattern; appended last, the grave
also wins the TUI's last-write-wins per-tile glyph, and it blocks future
`buildSite` on the tile via the blanket any-structure check. The `run.ended`
arm (spec 044 US1) is the terminal latch: it sets `State.Ended` — which no
event ever clears, so replay/restart lands back in the ended posture and
migration tooling cannot resurrect a finished run without rewriting history —
and copies the payload verbatim onto `State.RunEnd` ([[sim-loop]] holds the
matching ended posture; the [[executor]] emits the event). The cognition telemetry types — `cog.thought`, `cog.outcome`,
`cog.recalibration_recommended`, (since spec 017)
`cog.tool_call` (the tool-use loop's call trace, [[tool-loop]]), and (since
spec 042 US2) `cog.memory_divergence` (the shadow-mode selector's rank-
divergence record, [[memory-retrieval]]) — are explicit
reducer no-ops: recorded observability with zero state effect.
`agent.intent_rejected`, formerly in that no-op list, is since spec 043 US1
split into its own STATE-MUTATING arm: the refused intent never landed, so
`Intent`/`IdleSince` stay untouched, but the ring gains the appended-closed
`"rejected"` record described above — deterministic from the event alone, so
replay-safe.
Unknown types — including `daemon.*` and `world.created` — are recorded
history but state no-ops, so new event types never break old replay.

**Tick is deliberately not event-sourced**: quiet ticks (no events) advance the clock
without a log row. The live loop mutates `state.Tick` directly; recovery sets it to
`max(snapshot tick, last event tick)` and re-lives any quiet tail deterministically.

Canonical bytes: `Marshal()` uses `encoding/json` over structs only (fixed field
order — payload shapes like `AgentMovedPayload` are structs, never maps), so equal
states produce identical bytes. `Hash()` is SHA-256 of that, used by [[snapshots]]
verification and the determinism tests. Wall-clock time never appears in state.

## Connections

[[sim-loop]] generates events via the [[executor]] and applies them here;
[[daemon-lifecycle]] replays the [[event-log]] through `Apply` at startup;
[[event-types]] lists every payload struct (the cognition-horizon payloads
live in sibling files `cognition.go`, `guard.go`, and `plan.go`; the `Journal`
type, its rune budget, and the two `journal.*` payloads live in `journal.go`;
the wall predicates `isWall`/`wallMaxHP`/`wallAt` live in `terrain.go`; the
spec 044 run-outcome types — `RunEnd`, `DeathRecord`, `RunEndedPayload` — and
`livingCount` (moved here from `governance.go`) live in `state.go`, while the
`MorgueEpilogue` ring types and `applyMorgueEpilogue` live in `morgue.go` —
[[morgue]]);
[[world-migration]]
is the sole producer of `world.migrated`; [[metatron-miracles]] covers the
miracle payload shapes, cost table, and the `rebaseTicks` shift-semantics
taxonomy `applyTimeSnapped` uses (which, since spec 029, also shifts an active
standing order's `ExpiresTick` — never its `PlacedTick` — across a time snap;
since spec 043 it likewise SHIFTs `Agent.NeedsAnchorTick` — a live-read
duration anchor, 0 staying 0 — while `IntentRecord.Tick`/`OutcomeTick` are
KEEP, history never rewritten);
[[metatron-orders]] covers the standing-order lifecycle, placement validation,
and the angel-side trigger/confirm mechanics built on top of this reducer arm.
[[mental-maps]] covers `Agent.Map`'s type, its four knowledge events'
reducer arms, and the derived explored/sighting bookkeeping several
movement-family arms now perform. [[memory-retrieval]] covers
`Memory.Seq`/`Vec`/`VecModel` and `Agent.SitVec*`'s producer (the mind-side
embedder), the `agent.memory_embedded`/`agent.situation_embedded` arms this
reducer owns, and the `cog.memory_divergence` telemetry the shadow-mode
selector records. [[decision-context]] is the consumer of the spec 043
surfaces this reducer derives — the `IntentLog` ring (types and mutators
`IntentRecord`/`appendIntent`/`stampIntentOutcome`/`stampOrAppendExpired` in
`agents.go`) renders as the prompt's self-history block and
`NeedsAnchor`/`NeedsAnchorTick` as its need-trajectory arrows.
[[curriculum-ladder]] covers the `curriculum.*` payload shapes,
`EvaluateUnlock`'s per-stage gate conjuncts, the sanctioned
`CharterObservedEvidence` constructor, and the per-user unlocks record the
daemon projects from the `StagesUnlocked` latch this reducer owns.
[[world-tuning]] covers the manifest file, the five promoted dials' defaults
and clamp bounds, and the daemon boot seed that emits `sim.tuning_applied`;
this reducer owns only the `State.Tuning` field and the one idempotent arm.

## Operational notes

`EffectiveRate`/`Degraded` are part of state (hence snapshots) but only change via
explicitly emitted transition events, so unloaded same-machine runs stay
byte-deterministic. Adding a state field means adding events that set it — direct
mutation outside `Apply` (except `Tick`) breaks the replay contract.
