# Research: guardian directives and designations (spec 084)

Decisions with evidence, verified against the task worktree
(`.worktrees/task-157`, branch `task-157-guardian-directives` at current
main including TASK-97's merged `internal/target`). Format: Decision /
Rationale / Evidence / Alternatives considered.

## R1 — Clone the `sim.GuardianOrder` entity discipline verbatim

**Decision**: `Designation` and `Directive` copy the standing-order
discipline: deterministic human-readable ids minted guardian-side with no
RNG, one-way status transitions resolved at the reducer door, payload
`Status`/`PlacedSeq` ignored and reducer-stamped, caps validated in the
arm (validate-not-clamp), and an active+recent-32 retention prune.

**Rationale**: the board card names this explicitly ("entity discipline
clones sim.GuardianOrder"); the discipline is proven across snapshots,
restart, and from-genesis replay, and its race resolution (exactly one
terminal lands) is exactly what the cancel/fulfil/expiry races here need.

**Evidence**:
- Entity + lifecycle: `internal/sim/guardian.go:172-192` (`GuardianOrder`
  struct — `omitempty` `PlacedSeq` reducer-stamped per the `Memory.Seq`
  precedent, comment at 185-191).
- Validate-not-clamp arm: `internal/sim/guardian.go:335-409`
  (`metatron.order_placed`: duplicate-id-in-any-status rejection 344-349,
  origin domain 350-354, TTL bounds 374-378, cap 388-400, payload
  `Status` ignored / `PlacedSeq` stamped 401-408).
- One-way transitions + race resolution:
  `internal/sim/guardian.go:487-499` (`transitionGuardianOrder` — the
  loser of a race finds a non-active order and is refused).
- Retention prune: `internal/sim/guardian.go:505-525`
  (`pruneGuardianOrders`, deterministic, order-preserving,
  `guardianOrderRetain` 32 at :164).
- Deterministic ids: `internal/guardian/orders.go:63-88` (`nextOrderID`,
  "ord-<placedTick>-<seq>", "human-readable, deterministic, no RNG draw";
  uniqueness ultimately reducer-enforced).

**Alternatives**: UUIDs (rejected: RNG in deterministic space); separate
per-kind designation entities (rejected: one discriminated entity mirrors
`GuardianOrder.Survival`'s discriminator shape and keeps one reducer arm).

## R2 — Fulfillment/expiry are executor-emitted; placement/issue/cancel are injected

**Decision**: `designation.placed`/`designation.cancelled`/
`directive.issued`/`directive.cancelled` join `injectSocialWhitelist` and
land through `InjectSocial` (guardian tool handlers);
`designation.fulfilled`/`directive.fulfilled`/`directive.expired` are
executor-emitted from `stepEvents` as pure functions of (state, tick) and
do NOT join the whitelist.

**Rationale**: this is the exact split the standing orders use, and it is
what makes fulfillment replay-reproducible with no guardian running (the
board AC #3: "fulfillment predicates stamped in the reducer" — the sweep
emits, the arm validates the predicate and transitions). A live-only
injected fulfillment would vanish on replay.

**Evidence**:
- The split stated: `internal/sim/loop.go:207-211` +
  `docs/wiki/sim-loop-injection-doors.md` — order_placed/cancelled/
  triggered whitelisted, "`metatron.order_expired` needs no whitelist
  entry — it is executor-emitted, never injected, the `charge_regenerated`
  precedent".
- The sweep pattern: `internal/sim/executor.go:59-73` (order-expiry sweep:
  "a pure function of (state, tick), exactly the charge_regenerated
  pattern… Emitted once: the same event marks the order non-active, so
  the next tick no longer sees it active. Replay reproduces it
  deterministically without the guardian running").
- Whitelist accessor for gates: `internal/sim/loop.go:335`
  (`InjectableSocialEvent`).
- Ended-world narrowing refuses injected order events
  (`docs/wiki/guardian-orders.md` Operational notes) — the new injected
  types inherit the same behavior by joining the same whitelist.

**Alternatives**: reducer emitting follow-on events when `agent.built`
applies (rejected: reducer arms never emit — single-writer doctrine);
guardian-side live matching emitting fulfillment (rejected: live-only,
breaks replay, and needs a running model for a structural fact).

## R3 — The plan layer is charge-free (survey, designations, directives)

**Decision**: `survey_site` (Read), `place_designation`,
`cancel_designation`, `issue_directive`, `cancel_directive` are all
`Gate: None`, zero charges.

**Rationale**: the board card decides SURVEY is charge-free explicitly.
For the rest: charges price world EDITS — the miracle economy
(`work_miracle` per-kind costs, `send_vision`/`send_omen` Gate Charge) —
while designations and directives edit nothing physical: villagers still
do all the work by their own logic, which is precisely the genre doctrine
the card cites ("the player… influences the environment or general
directives, and the simulation responds" —
`research/Game-Gameplay-Patterns/Indirect-Control-and-Divine-Intervention.md`
§"The defining constraint"). Decisively: the operator's
interruption ruling names "guardian re-issue" as the FIRST workaround for
stalled directives — pricing re-issue in charges would fight the ruling's
own remedy. Standing orders set the precedent for charge-free agency
surface (`monitor_and_act`/`cancel_order` are `Gate: None`).

**Evidence**: `internal/tool/registry.go:538-543` (monitor_and_act /
cancel_order `Gate: None`), :519-537 (send_vision/send_omen
`Gate: Charge` — influence on minds via prose is priced), :573-577
(work_miracle Gate Charge); the mana-economy doctrine note
(`Indirect-Control-and-Divine-Intervention.md` §"The mana economy") prices
*intervention*, not planning. Flagged in spec.md Assumptions for operator
review — a future economy pass can price the doors without shape changes.

**Alternatives**: `issue_directive` at 1 charge (vision parity — it
reaches minds like a vision). Rejected for the re-issue-workaround
conflict above, but the door shape (Gate/Cost fields) makes this a
two-line future change.

## R4 — Designation loci parse through `internal/target` via a bare-locus entry point

**Decision**: add exported `ParseLocus(s) (Address, error)` to
`internal/target`, reusing the existing `parseLocus`/`parsePoint`
internals and `Address.Tiles()`; designation tools take a bare locus
string (`"4,5"`, `"1,1..8,8"`, `"2,2->2,9"`). The four entity classes,
the reserved-prefix rule, and the bundle form matrix are untouched.

**Rationale**: the 082 contract guarantees "designation params parse
through the SAME package bundle effects use" — one parser, one
enumeration order. But the 082 *class* token names world entity classes
(villager/structure/pile/terrain), none of which a designation addresses;
forcing `settlement_zone` into the class table would widen the
reserved-prefix set (a documented bundle-compat surface: any new reserved
word changes how bundle bare-name strings parse) for zero benefit. A
bare-locus entry point shares every line of locus grammar, normalization,
and `Tiles()` — the one-parser law holds where it matters (two consumers
can never disagree on what `1,1..8,8` means or enumerates) — while
leaving 082's contract surface byte-stable.

**Evidence**: `internal/target/target.go:104-117` (`reservedSplit` — the
class table IS the reserved-prefix set: "kept as one table so the two
cannot diverge"); :171-216 (`parseLocus` — already class-independent
internally, taking `class` only to stamp the result); :265-302 (`Tiles()`
— "a pure function of the Address, no state"); the seam contract:
`specs/036-scriptable-agent-tools/contracts/bundle-manifest.md:103-121`
("Designation addressing (TASK-157 seam)": leaf-safety, the three forms,
enumeration order — nothing there requires the class token for
designations); `specs/082-target-addressing/data-model.md` §4 ("Class
vocabulary extension (e.g. a designation-kind token) is a one-table
change" — permitted, not mandated).

**Alternatives**: extend the class table with designation kinds
(rejected: widens the bundle reserved-prefix surface and drags a bundle
matrix + fixture change into this spec); a designation-local mini-parser
(rejected outright: violates the one-parser invariant, data-model 082 §6.1).

## R5 — The DIRECTIVE rung sits after `survivalDecision`, before the `prepYields` gate, unconditioned

**Decision**: `decideIntent` becomes SURVIVAL → DIRECTIVE →
(prepYields?) PREP → wander. `directiveDecision` is NOT gated by
`prepYields` (neither the yield window nor the danger bands).

**Rationale**: the operator ruling is structural: "villagers first make
sure they are not dying, then execute active directives, then free time".
Survival preempting = the rung runs after `survivalDecision` returns
nothing. Directives preempting prep/wander = the rung runs before the
prep gate is even consulted. The yield window exists so instinct doesn't
counter-schedule the MIND (spec 062); a directive is not instinct
noise — it is the villager's current duty, and the planner sees it too
(the block), so rung and planner pull the same direction rather than
thrash. The danger bands belong to survival's ownership, which has
already had its chance one line above.

**Evidence**: `internal/sim/policy.go:40-56` (the ladder structure —
"The classification lives in this function STRUCTURE (FR-001), not in
prose"); :73-81 (`prepYields` — both clauses are prep-only doctrine;
"SURVIVAL rungs are exempt: they decide before this gate is ever
consulted" — the same exemption argument covers a hard duty layer);
`docs/wiki/reflex-prep-arbitration.md` (the yield window is "arbitration
doctrine, not scheduling" — aimed at PREP specifically).

**Alternatives**: gating the rung on the yield window (rejected: a
directive would go slack for 1800 ticks after every planner intent —
softness the ruling forbids); a directive-as-plan (injecting `set_plan`
steps — rejected: plans are mind-authored artifacts behind the firewall,
and a guardian-authored plan would bypass the villager's own agency AND
need new injection surface).

## R6 — Interruption needs zero new code; the rung re-resolves at idle

**Decision**: no pause/resume/interruption code anywhere in the diff.
The proof obligation is a test, not machinery.

**Rationale**: `decideIntent` only runs for an idle, awake agent; hails
already pause and close intents through the executor's hail machinery;
conversations run through the scene machinery. When the interruption
ends, the agent goes idle, `decideIntent` runs, survival gets first
claim, and the DIRECTIVE rung re-resolves toward the same designation —
resumption is a free consequence of statelessness (the rung re-derives
its intent from `State.Directives` every time; nothing to resume).

**Evidence**: `internal/sim/policy.go:9-12` ("deciding what an idle,
awake agent does next"); `docs/wiki/executor-social-perception.md`
summary (hails pause/close talk_to intents); the operator ruling text on
the card ("conversations, hails, and dynamic world stimuli CAN and
SHOULD interrupt directed work — interruption is life").

**Alternatives**: a directive-execution state machine on the Agent
(rejected: statefulness invents resume bugs and violates
in-game-workaround-first); suppressing hails toward directed villagers
(rejected explicitly by the ruling — "must not be discouraged").

## R7 — The `directive` block is `neverDrop`, between `plan_echo` and `known_places`

**Decision**: eleventh spec-043 block, contract position after
`plan_echo`, priority `neverDrop`, cap 2 directives × bounded rendering.

**Rationale**: a hard command a villager can forget under budget pressure
is not hard; the block is small and bounded (≤2 directives, text ≤400
runes each, a few framing lines), so `neverDrop` costs little. Position:
after the agent's own plan (the mind's commitments render first — the
directive contextualizes them, and a planner deciding between its plan
and the divine command should see both adjacent), before the droppable
world-knowledge blocks.

**Evidence**: `internal/mind/context.go:74-85` (`fixedBlocks` — the
insertion point between `plan_echo` P6 and `known_places` P5);
`docs/wiki/context-block-inventory.md` (the ten-block table, empty blocks
render `""` and are omitted entirely — the byte-identity guarantee for
directive-free worlds is free); `internal/mind/context.go:38`
(`neverDrop`).

**Alternatives**: droppable at priority 7 (rejected: hardness);
rendering directives inside `known_places` (rejected: a command is not
geography, and that block is droppable).

## R8 — Announcement rides the spec-041 place-grant machinery, reducer-side, all-villagers

**Decision**: the `designation.placed` arm upserts
`PlaceFact{Kind:"designation", Provenance: ProvenanceRevealed, Seen:
e.Tick, Detail: kind+label}` at the designation's anchor tile into every
living villager's mental map, directly in the arm (no companion events).
`designation` joins the closed `PlaceFact` kind vocabulary with its own
`factHorizon` entry; `placeFactKinds` (send_vision's enum) is NOT
extended.

**Rationale**: the card decides the channel ("announced as village
knowledge via the spec-041 place-grant machinery"). Doing it in the arm
(rather than N companion `metatron.place_revealed` events) is the
`agent.saw`-shaped choice: one event, deterministic fan-out, reducer
stays total (map-less agents skipped), and the `place_revealed` arm's
`groundFactPresent` check ("the god reveals what IS") would need a
special case anyway — a designation is not a pre-existing world thing,
it becomes real BY this event. `placeFactKinds` stays visions-only
because that enum's contract is "reveal one REAL place".

**Evidence**: `internal/sim/guardian.go:300-334` (the
`metatron.place_revealed` arm — Seen/Provenance/Detail stamped
normatively, map-less agents skipped, "the reducer stays total");
`internal/sim/mentalmap.go:49` (PlaceFact "Kind is a closed vocabulary"),
:289-292 (`upsertFact` one-fact-per-(Kind,X,Y) invariant), :102
(`factHorizon` per-kind freshness); `internal/tool/registry.go:475-484`
(`placeFactKinds` — "the domain of send_vision's optional place_kind").
Anchor tile: point = the tile; line = first endpoint (author intent,
082's preserved order); rect = the normalized min corner (deterministic,
already computed).

**Alternatives**: companion `metatron.place_revealed` per villager
(rejected: 8 events per placement, and the real-place check misfits);
no mental-map grant, block-only (rejected by the card's decision — and
`known_places` rendering gives reflex `nearestKnown`-style resolution
and prompt geography for free).

## R9 — `directive.*` joins `observableEventTypes`; designation events do not (v1)

**Decision**: add exactly `directive.issued`, `directive.fulfilled`,
`directive.cancelled`, `directive.expired` to `observableEventTypes`
(12 → 16 entries).

**Rationale**: AC #7 names the `directive.*` lifecycle. Every entry must
be a genuinely emitted type (the vocabulary's own law — the
`meeting.norm_enacted` lesson at `internal/tool/registry.go:410-417`),
which all four are. Designation events stay out in v1 to keep the watch
vocabulary tight: a watcher of plan progress watches
`directive.fulfilled` (directives bind designations, so designation
fulfillment surfaces there); widening is a one-slice change when
evidence demands it.

**Evidence**: `internal/tool/registry.go:418-423` (the current 12);
:451-473 (`monitorAndActSchema` — `event_types` items enum over the
vocabulary, so the change is enum-only: zero schema-shape or matching
code); `docs/wiki/guardian-orders.md` + `internal/guardian/orders.go`
(matching is type-keyed over `EventTypes` — any recorded event of a
listed type flows through `matchOrders` unmodified: the zero-new-trigger-
code claim is structural).

## R10 — Map rendering via three new tile-registry rows; zone renders perimeter-only

**Decision**: three new registry rows (structure-site marker, wall-line
segment, zone perimeter), rendered beneath real entities; interior zone
tiles unmarked; fulfilled/cancelled designations stop rendering.

**Rationale**: spec 068's registry IS the three-surface discipline
("a tile added (or re-skinned) here reaches all three surfaces with no
other edit" — `internal/tui/tiles.go:1-22`); new world things get new
rows, state changes get style variants. Perimeter-only for zones keeps
the map readable (no wallpaper over terrain the villagers still use) and
matches the DF designation-overlay idiom of marking extent, not filling
it.

**Evidence**: `internal/tui/tiles.go:1-22` (registry doctrine; variants
are style transforms, never new characters — so active/inactive is
"render or don't", derived from designation Status in state);
`docs/wiki/tile-registry.md` (glyph/legend/walkthrough from one table);
TUI renders from the log-shipped replica's `sim.State`, so
`State.Designations` reaches the map with no wire change
(`docs/wiki/tui-client.md` summary). Design pages to amend:
`docs/design/tui/panels/map.md` (+ legend), `docs/design/tui/pages/
guardian-console.md` (spec-047 gate: `node scripts/check-tui-design.mjs
--changed` names the authoritative set).

## R11 — Rebase taxonomy: active `Directive.ExpiresTick` SHIFTs; everything else KEEPs

**Decision**: `rebaseTicks` gains: for each ACTIVE directive, shift
`ExpiresTick`; `IssuedTick`/`PlacedTick` (both entities) and
`Designation` in full are KEEP (no future deadlines).

**Rationale/Evidence**: the exact `GuardianOrders` classification —
`internal/sim/miracles.go:327-334`: "A standing order's expiry is a
future deadline: shift only ACTIVE orders so the remaining lifetime is
preserved across the jump. PlacedTick is a historical timestamp and is
left unshifted (KEEP)." The wiki taxonomy note
(`docs/wiki/guardian-miracle-rebase-taxonomy.md`) requires every new
tick-anchored int64 to classify — the data-model carries the table.

## R12 — Survey is turn-side, over the state mirror + static map

**Decision**: `survey_site` computes guardian-side (`internal/guardian`,
a `buildTargetingDigest` sibling) from the turn's state snapshot and the
static `worldmap.Map`; the tool package carries only the declaration and
guidance prose.

**Rationale/Evidence**: the exact spec-059/063 split:
`internal/tool/derive.go:272-291` (`GuardianReadGuidance` renders Read
tools' "read freely" paragraph; `derive.go:210-212` `guardianToolDesc`
gains a `survey_site` entry beside `explain`'s);
`internal/guardian/turn.go:1077-1086` (`buildTargetingDigest` — "the
tool package has no world state to draw positions/passability from",
so the digest assembles turn-side); `internal/guardian/toolcalls.go:116`
(`handleExplain` — the Read-dispatch handler shape to clone);
`internal/tool/registry.go:589-592` (`explain`: Effect Read, Gate None,
repairable-miss doctrine for bad input). Determinism: the sheet is a
pure function of (args, state snapshot, map) — fixed iteration orders,
no clock reads.

## R13 — New goal `heed_directive` for the walk-to-site leg

**Decision**: the rung's movement leg is a new goal `heed_directive`
(walk to target tile, instant completion on arrival), registered through
the standard goal/duration mirrors; build legs reuse the EXISTING build
goals with the designation tile as the target.

**Rationale**: reusing `search` for directed movement would lie in the
intent log and the self-history block ("exploring" vs "heeding the
Guardian"); a named goal renders honestly everywhere
(`selfHistoryLine`, digests) and gives tests a crisp assertion handle.
Instant-on-arrival is the `search` completion shape — arrival IS the
outcome; the next idle decision picks the work leg.

**Evidence**: `docs/wiki/executor-goals-and-intents.md` (the Intent
state machine: walk, instant-on-arrival vs work-goal classes, duration
lookup); `docs/wiki/tool-registry.md` (derived durations/vocabulary —
new goals join the derivation, and `TestCatalogSweep`-adjacent coverage
gates catch missing rows).

## R14 — Sweep ordering inside the tick is fixed: designations, then directives

**Decision**: `stepEvents` evaluates designation fulfillment first, then
directive fulfillment/expiry, each in slice order.

**Rationale**: directive fulfillment reads designation status; sweeping
designations first means a designation fulfilled at tick T yields
`directive.fulfilled` at T+1's sweep (one-tick lag, deterministic,
documented) rather than an order-dependent same-tick race. Slice order
is append order — deterministic under replay by construction (the
`pruneGuardianOrders` order-preservation argument).

**Evidence**: `internal/sim/executor.go:55-73` (existing sweep block —
the insertion site and the once-only idiom: "the same event marks the
order non-active, so the next tick no longer sees it active").
