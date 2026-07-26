# Data Model: designations, directives, and the DIRECTIVE rung (spec 084)

NORMATIVE definitions for the two entities, the seven-event vocabulary
(payload shapes are contract-bound in [contracts/events.md](contracts/events.md)),
the fulfillment predicate table, the `directive` decision-context block,
and the reflex-rung semantics. The entity discipline is
`sim.GuardianOrder`'s (research R1): deterministic ids, one-way status,
validate-not-clamp reducer arms, reducer-stamped `PlacedSeq`, retention
prune.

## 1. Designation kinds and addressing

A designation is one of three kinds, each bound to one locus form of the
spec-082 grammar (parsed via `internal/target`'s bare-locus entry point,
research R4 — same normalization, same `Tiles()` enumeration as every
other consumer):

| Kind | Locus form | Locus example | Extra parameters | Anchor tile (place-fact grant) |
|---|---|---|---|---|
| `structure_site` | point | `4,5` | `structure_kind` (required; a buildable structure kind) | the tile |
| `wall_line` | line (axis-aligned, endpoint order preserved) | `2,2->2,9` | `structure_kind` (optional; `wall_plank`\|`wall_stone` narrows, empty = any wall) | first endpoint |
| `settlement_zone` | rect (normalized min/max corners) | `1,1..8,8` | `min_structures` (optional; default 3, bounds 1..12) | normalized min corner |

Any other (kind × form) pairing is a `form` error at the door (the
082 taxonomy, consumer-side). Loci are validated in-bounds against the
world map dims (`bounds`), and stored NORMALIZED — the payload carries
ints, never the address string; replay never re-parses.

**Door-side occupancy validation** (structural fulfillability, spec.md
edge cases): a `structure_site` tile already holding a structure of a
different kind is rejected; each `wall_line` tile holding a non-wall
structure is rejected. `settlement_zone` rects are never
occupancy-checked. Size bounds: rect area ≤ 256 tiles, line length ≤ 32
tiles.

## 2. `sim.Designation`

```
Designation {
  ID            string  `json:"id"`                        // "dsg-<placedTick>-<seq>" (nextOrderID shape, no RNG)
  Kind          string  `json:"kind"`                      // settlement_zone | structure_site | wall_line
  X, Y          int     `json:"x","y"`                     // point tile / line first endpoint / rect min corner
  X2, Y2        int     `json:"x2,omitempty","y2,omitempty"` // line second endpoint / rect max corner; == X,Y for a point
  StructureKind string  `json:"structure_kind,omitempty"`  // structure_site: required; wall_line: optional narrowing
  MinStructures int     `json:"min_structures,omitempty"`  // settlement_zone only; landed value always 1..12 (default 3 applied at the tool door)
  Label         string  `json:"label,omitempty"`           // guardian's name for it, ≤80 runes
  PlacedTick    int64   `json:"placed_tick"`
  Status        string  `json:"status"`                    // active → fulfilled | cancelled (one-way)
  PlacedSeq     int64   `json:"placed_seq,omitempty"`      // reducer-stamped from e.Seq (spec 054 precedent); ignored on the wire
}
```

State: `State.Designations []Designation` (`omitempty` — a pre-084
snapshot unmarshals to nil, the spec-029 `GuardianOrders` precedent; no
format bump). Cap: at most **16 ACTIVE** designations
(`GuardianDesignationCap`, door-validated). Prune: every active + the
most recent 32 non-active (the `pruneGuardianOrders` algorithm,
generalized or cloned — deterministic, order-preserving).

## 3. `sim.Directive`

```
Directive {
  ID            string  `json:"id"`                    // "dir-<issuedTick>-<seq>"
  DesignationID string  `json:"designation_id"`        // must name an ACTIVE designation at issue (v1's only checkable-goal binding)
  Targets       []int   `json:"targets"`               // living villager indices, resolved at issue, ascending; non-empty
  Village       bool    `json:"village,omitempty"`     // issued to "everyone" (Targets = all living at issue; provenance marker)
  Text          string  `json:"text"`                  // guardian framing, 1..400 runes
  IssuedTick    int64   `json:"issued_tick"`
  ExpiresTick   int64   `json:"expires_tick"`          // issued + ttl_days game days; bounds 1..7 (GuardianOrderTTLMinDays/MaxDays, SHARED constants)
  Status        string  `json:"status"`                // active → fulfilled | cancelled | expired (one-way)
  PlacedSeq     int64   `json:"placed_seq,omitempty"`  // reducer-stamped
}
```

State: `State.Directives []Directive` (`omitempty`). Cap: at most **3
ACTIVE** directives (`GuardianDirectiveCap` — the player-order-cap
number, same rationale: a bounded attention economy). Prune: active +
most recent 32 non-active. TTL default 3 game days, applied at the tool
door (the `monitor_and_act` shape); the reducer validates the bound.

**Targeting resolution (tool door, `issue_directive`)**: `targets` is a
comma-separated list of living villager names, or `"everyone"` (the
`send_omen` vocabulary). Every named villager must resolve and be living
— any dead/unknown name rejects the whole call (atomic,
`rejected_gate`). `"everyone"` resolves to all living indices and sets
`Village: true`. The reducer arm re-validates: non-empty ascending
unique in-range indices, all living at apply.

## 4. Event vocabulary

Full payload shapes, door rules, and membership tables are NORMATIVE in
[contracts/events.md](contracts/events.md). Summary:

| Type | Door | Emitter | Reducer effect |
|---|---|---|---|
| `designation.placed` | injected (`InjectSocial`, whitelisted) | `place_designation` handler | validate (id/kind/form/bounds/occupancy/cap/label); land `active`, stamp `PlacedSeq`, prune; **grant place facts to all living villagers** (§5) |
| `designation.cancelled` | injected | `cancel_designation` handler | `active → cancelled` (one-way door; non-active refused) |
| `designation.fulfilled` | executor-emitted (NOT whitelisted — the `charge_regenerated` precedent) | `stepEvents` sweep | re-validate the predicate (§6) holds, then `active → fulfilled` |
| `directive.issued` | injected, atomically with companion `agent.memory_added` per target | `issue_directive` handler | validate (designation active, targets living, TTL bounds, text runes, cap); land `active`, stamp `PlacedSeq`, prune |
| `directive.cancelled` | injected | `cancel_directive` handler | `active → cancelled` |
| `directive.fulfilled` | executor-emitted | `stepEvents` sweep | validate bound designation is `fulfilled`, then `active → fulfilled`. Payload `{id, designation_id, targets, issued_tick}` — **the TASK-118 faith-accounting contract** |
| `directive.expired` | executor-emitted | `stepEvents` sweep | validate (tick ≥ ExpiresTick OR no living target), then `active → expired` |

**Whitelist delta** (`injectSocialWhitelist`, `internal/sim/loop.go:211`):
+`designation.placed`, +`designation.cancelled`, +`directive.issued`,
+`directive.cancelled`. The three executor-emitted types get NO entry.
None of the seven joins the ended-world prose whitelist.

**Observable delta** (`observableEventTypes`,
`internal/tool/registry.go:418` — AC #7): +`directive.issued`,
+`directive.fulfilled`, +`directive.cancelled`, +`directive.expired`
(12 → 16). Enum-only change; `monitor_and_act` matching code untouched.

**Sweep order within a tick** (research R14): designation fulfillment
first (slice order), then directive fulfillment, then directive expiry
(slice order). Each fires once — the landed event flips the entity
non-active, so the next tick's sweep skips it. A designation fulfilled
at tick T yields its bound directives' `directive.fulfilled` at T+1
(documented one-tick lag; deterministic).

## 5. The announcement grant (research R8)

The `designation.placed` arm upserts, into EVERY living villager's
mental map (map-less agents skipped — the reducer stays total):

```
PlaceFact{
  Kind:       "designation",           // NEW closed-vocabulary member (sim/mentalmap.go)
  X, Y:       <anchor tile, §1 table>,
  Seen:       e.Tick,                  // stamped normatively, the place_revealed shape
  Provenance: ProvenanceRevealed,
  Detail:     <kind phrase + label>,   // e.g. `a marked shelter site ("north shelter")`
}
```

- `factHorizon("designation")` = 7 game days (the max directive TTL —
  the announcement outlives any directive bound to it).
- `placeFactKinds` (send_vision's enum) is NOT extended.
- `renderKnownPlaces` renders designation facts as landmarks (the
  grave precedent: individually named, provenance-flavored).
- Cancellation/fulfillment does NOT retract facts — remembered history,
  the burned-out-fire precedent; the `directive` block (§7), rendered
  from world state, is the execution authority.

## 6. Fulfillment predicate table (pure state checks)

Evaluated identically by the executor sweep (to emit) and the reducer
arm (to validate before transitioning) — both read only `sim.State` (+
static map for nothing; no map needed):

| Kind | Predicate `fulfilled(s *State, d Designation) bool` |
|---|---|
| `structure_site` | a structure of kind `d.StructureKind` stands at `(d.X, d.Y)` (the one-structure-per-tile probe) |
| `wall_line` | for EVERY tile of the line's `Tiles()` enumeration: a wall structure stands there (`wall_plank` or `wall_stone`; exactly `d.StructureKind` when set) |
| `settlement_zone` | count of structures whose tile lies within the inclusive rect ≥ `d.MinStructures` |

No clock, no RNG, no I/O; deterministic iteration (Tiles() order /
`s.Structures` slice order). A predicate already true at placement is
legal — the sweep fulfills at the next tick boundary.

## 7. The `directive` decision-context block (spec 043 extension)

Inserted in `fixedBlocks` (`internal/mind/context.go:74`) between
`plan_echo` and `known_places`; the block table becomes eleven rows:

| Field | Value |
|---|---|
| Name | `directive` |
| Contract position | 6 (after `plan_echo`; `known_places` and below shift down one) |
| Priority | `neverDrop` (a hard command is never shed) |
| Content | ≤2 ACTIVE directives addressing this agent (individually, or via `Village`), oldest first (`IssuedTick`, then id). Per directive: the guardian's `Text` verbatim; the bound designation's kind, site (anchor + extent in plain words), and what fulfillment requires (predicate phrase); plain-words time remaining (the days-left shape from `writeStandingOrders`). |
| Empty condition | no active directive addresses the agent → `""` (omitted entirely; assembled prompt byte-identical to pre-084) |
| Cap | 2 directives; `Text` is already ≤400 runes at the door |

Rendering is a pure function of `(State, agentIdx)` — available in every
degraded mode, model-free.

## 8. DIRECTIVE reflex rung (normative semantics + pseudocode)

Position (`internal/sim/policy.go:40` `decideIntent`):

```go
func decideIntent(s, m, idx, tick) decision {
    a := &s.Agents[idx]
    if d, ok := survivalDecision(s, m, a, tick); ok { return d }   // unchanged: survival always preempts
    if d, ok := directiveDecision(s, m, a, idx, tick); ok { return d } // NEW: unconditioned by prepYields
    if !prepYields(s, a, tick) {                                    // unchanged
        if d, ok := prepDecision(s, m, a, tick); ok { return d }
    }
    return wanderDecision(s, m, a, idx, tick)
}
```

`directiveDecision` (deterministic; reads state only):

```
directiveDecision(s, m, a, idx, tick):
  d := oldest ACTIVE Directive in State.Directives (slice order = issue order)
       with idx ∈ d.Targets
  if none: return (_, false)
  dsg := State.Designations[d.DesignationID]
  if dsg == nil or dsg.Status != "active": return (_, false)   // orphan: sweep will expire d
  switch dsg.Kind:

  case structure_site:
    if reflexBuildable(dsg.StructureKind) and materialsInHand(a, dsg.StructureKind):
        return Intent{Goal: buildGoalFor(dsg.StructureKind), TargetX: dsg.X, TargetY: dsg.Y}, true
    if !atOrAdjacent(a, dsg.X, dsg.Y):
        return Intent{Goal: "heed_directive", TargetX: dsg.X, TargetY: dsg.Y}, true
    return (_, false)      // at site, cannot act: planner's job (block drives it); never deadlock

  case wall_line:
    t := first tile in dsg.Tiles() order lacking a qualifying wall
    if none: return (_, false)                                  // fulfilled; sweep will stamp it
    if wallMaterialsInHand(a, dsg.StructureKind):
        return Intent{Goal: "build_wall", TargetX: t.X, TargetY: t.Y (+ wall kind)}, true
    if !atOrAdjacent(a, t.X, t.Y):
        return Intent{Goal: "heed_directive", TargetX: t.X, TargetY: t.Y}, true
    return (_, false)

  case settlement_zone:
    if !inRect(a, dsg):
        return Intent{Goal: "heed_directive", TargetX/Y: nearest rect tile (BFS-deterministic)}, true
    return (_, false)      // presence achieved; what to build inside is mind work
```

Normative properties:

- **Survival always preempts** (rung position — AC #5).
- **Directives preempt prep/wander** whenever the rung resolves
  (position before the `prepYields` consult — AC #5).
- **No new interruption code** (AC #6): the rung is stateless — it
  re-derives everything from `State.Directives`/`State.Designations` at
  each idle decision, so resumption after any interruption is free.
- **Inert without directives**: first check is a state scan that finds
  nothing → byte-identical pre-084 behavior (SC-006).
- `reflexBuildable` is the CLOSED set of structure kinds the reflex
  layer already knows how to build with materials in hand (fire, chest,
  oven — the `resolveGoal` build set MINUS planner-only shelter, spec
  012); everything else routes to walk-to-site + planner.
- `heed_directive` is a NEW goal (research R13): walk to target,
  instant completion on arrival (the `search` completion shape),
  registered through the standard goal/duration derivations; Source
  stays `"reflex"` (an instinct-layer decision; the honest
  self-history line is the goal name itself).

## 9. Tool surface delta (`internal/tool/registry.go` `guardianTools`)

| Tool | Effect / Gate | Params | Events |
|---|---|---|---|
| `survey_site` | Read / None | `x` Number req; `y` Number req; `radius` Number opt (default 4, clamp 1..8) | — (Read grounds nothing) |
| `place_designation` | Expressive / None | `kind` Enum{settlement_zone, structure_site, wall_line} req; `target` Text req (bare locus, 082 grammar); `structure_kind` Enum (buildable-structure vocabulary; required for structure_site — handler-enforced, the parseReveal partial-triple shape); `min_structures` Number opt; `label` Text opt ≤80 runes | `designation.placed` |
| `cancel_designation` | Expressive / None | `id` Text req | `designation.cancelled` |
| `issue_directive` | Expressive / None | `designation_id` Text req; `targets` Text req (comma-names or `"everyone"` — the send_omen vocabulary); `text` Text req ≤400; `ttl_days` Number opt (default 3, bounds 1..7) | `directive.issued`, `agent.memory_added` |
| `cancel_directive` | Expressive / None | `id` Text req | `directive.cancelled` |

All five append AFTER the existing tools (no registration position
shifts — the `explain` precedent). Declared `Events` ⊆
`injectSocialWhitelist` is pinned by `ValidateToolCoverage` as usual.
`survey_site` renders via `GuardianReadGuidance` (a `guardianToolDesc`
entry beside `explain`'s); the four acting tools render via
`GuardianToolGuidance`. Stage availability: every stage (the
`monitor_and_act` grant precedent — spec.md Assumptions).

## 10. Rebase taxonomy entries (`rebaseTicks`, `internal/sim/miracles.go:271`)

| Field | Class | Rationale |
|---|---|---|
| `Directive.ExpiresTick` (Status == "active" only) | SHIFT | future deadline; remaining lifetime preserved (the `GuardianOrders` arm at miracles.go:327-334, verbatim) |
| `Directive.IssuedTick`, `Directive.PlacedSeq` | KEEP | history / identity |
| `Designation.*` (PlacedTick, PlacedSeq) | KEEP | no future deadlines; history |
| `PlaceFact.Seen` for designation facts | already SHIFTed | rides the existing `a.Map.Facts` loop (miracles.go:299-304) — no new code |

## 11. Invariants

1. **One-way status, one terminal**: every transition goes through a
   `transitionGuardianOrder`-shaped door; races resolve to exactly one
   landed terminal (the loser is refused on a non-active entity).
2. **Reducer-only writes; payloads self-contained**: normalized loci and
   resolved target indices land in payloads; replay never re-parses or
   re-resolves.
3. **Sweeps are pure over (state, tick)** and fire once; live and replay
   emit identically with no guardian process running.
4. **The firewall holds**: guardian prose reaches villagers only as
   recorded event data (`directive.issued` payload + companion
   memories), re-rendered from state by the block. No prompt-side
   channel.
5. **Inertness**: a world with no designations and no directives is
   byte-identical to pre-084 in reflex decisions, assembled prompts, map
   render, and every existing test.
6. **One parser**: every locus in the system parses through
   `internal/target`; no designation-side grammar copy exists.
