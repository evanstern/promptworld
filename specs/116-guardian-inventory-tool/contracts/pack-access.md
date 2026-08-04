# Contracts: Guardian Pack Access (spec 116)

The binding surfaces this feature adds or changes. Everything here is enforced by a test;
anything not written here is free implementation choice.

## 1. `inspect_pack` — the read tool

| Field | Value |
| --- | --- |
| Registry name | `inspect_pack` |
| Effect / Gate | `Read` / `None` |
| Charge | 0 |
| Consumes the turn's act | No (the loop driver's read exemption, as `survey_site`) |
| Grant-gated | Yes — handler installed only when `d.grant.allows("inspect_pack")` |
| Stage-1/2 ceiling | Included (`stage1CeilingTools`) |
| Roster | `RosterGuardian`, `LoopRosterGuardian` — appended last so no existing tool's registration position shifts |

### Parameters

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| `villager` | string | yes | a villager's name; resolution is case-folded exactly as other villager-named tools resolve |

### Verdict

Always `read_ok`. There is no error path: a missing, unknown, or dead `villager` returns an
honest in-fiction sheet naming the living roster (the `explain` unknown-topic shape).

### Sheet format

Deterministic text. Fixed kind order — the `sim.Inventory` field order, which is also
`tool.GrantKinds()`' order for the overlapping kinds:

```
Cedar's pack — carrying 24/24, 0 free to receive:
- wood: 20
- planks: 4
Cedar carries no food.
```

Rules:

- One `- kind: count` line per kind with a non-zero count, in fixed order. Kinds at zero are
  omitted (an empty pack renders the empty-pack line below instead).
- `spears` and `axes` render their count and remaining uses:
  `- spears: 2 (uses left: 3, 7)`.
- Food kinds are `food_raw`, `food_cooked`, `meals`. The closing line is exactly one of:
  - `<Name> carries no food.` — when all three are zero;
  - `<Name> carries food: <n> food_cooked, <m> meals.` — listing only the non-zero food kinds,
    in fixed order.
- An empty pack renders:
  `<Name>'s pack — carrying 0/24, 24 free to receive: the pack is empty.` followed by the
  no-food line.
- The header's cap is `sim.BulkCap`, never a literal; free bulk floors at 0 (the
  `buildTargetingDigest` defensive floor).
- Two identical calls in one turn return byte-identical bytes.

## 2. The look-first gate

Applies **only** when the turn's origin carries `survival: true` (`turnOrigin`,
`internal/guardian/orders.go`).

| Tool call | Gated on |
| --- | --- |
| `send_vision` targeting V | `inspect_pack(V)` returned earlier this turn |
| `work_miracle{kind: give_item, villager: V}` | same |
| `work_miracle{kind: take_item, villager: V}` | same |
| everything else | ungated |

- The ledger is a per-turn set of resolved villager indices, held on `turnDispatch`, populated
  by `handleInspectPack` on a successful resolution only (an honest-miss sheet for an unknown
  name records nothing).
- A gated call returns `Verdict: rejected_gate` with a reason of the form:
  `look in Cedar's pack before you speak to him — call inspect_pack first`.
  The reason MUST contain the literal `inspect_pack` and the villager's name.
- A gated refusal spends no charge, lands nothing, and does not consume the turn's act — it is
  repairable within the loop's round cap, exactly as a door refusal is.
- `send_omen` is **not** gated: it addresses a group, not a pack.

## 3. `take_item` — the removal miracle

| Field | Value |
| --- | --- |
| Miracle kind | `take_item` (joins `tool.miracleKinds`) |
| Event type | `guardian.item_taken` (joins `kindToEvent`) |
| Charge | 1 (`miracleCosts`, `give_item`'s price) |
| Stage-1/2 ceiling | Excluded (world-shaping, the `work_miracle` precedent) |
| Item vocabulary | `tool.GrantKinds()` — the same set `give_item` accepts |

### Parameters

Reuses `work_miracle`'s existing flat surface: `villager`, `item`, `qty`. No new parameter is
added to the schema.

### Payload

```go
type ItemTakenPayload struct {
    Agent  Ref    `json:"agent"`
    Kind   string `json:"kind"`
    Qty    int    `json:"qty"`
    Gratis bool   `json:"gratis,omitempty"`
}
```

Registered in `internal/sim/payloads.go` so the event decodes for replay and rendering.

### Reducer arm (`applyItemTaken`)

Validation order, all of it BEFORE the charge spend and before any mutation
(`applyItemGranted`'s validate-not-clamp discipline):

1. `p.Agent.ID` in range → else `no villager at index N`
2. villager not dead → else `<Name> is beyond a working now`
3. `grantableKind(p.Kind)` → else `unknown item kind %q (takeable: …)`
4. `p.Qty > 0` → else `take quantity must be positive (got N)`
5. the villager carries at least `p.Qty` of `p.Kind` → else
   `taking N <kind> from <Name> is more than they carry (M)` — **reject whole, never clamp**
6. spend the charge
7. move exactly `p.Qty` units from `Agents[i].Inv` into `s.pileFor(a.X, a.Y)`:
   - `spears` / `axes`: take from the FRONT of the ascending remaining-uses slice (most-worn
     first), append into the pile's slice, keep both sides sorted ascending
   - food kinds: stamp the pile batch's `SpoilAt` as `e.Tick + rotWindowTicks` and merge into
     an existing batch of the same `(Kind, SpoilAt)` — verbatim the `agent.dropped` behavior.

     **Corrected 2026-08-03 during implementation.** This clause originally read "transfer
     preserving `FoodBatch` identity and spoilage". That is unsatisfiable: carried food has no
     spoilage to preserve — `sim.Inventory.FoodRaw/FoodCooked/Meals` are bare `int` counts, and
     `SpoilAt` exists only on `Pile.FoodBatch`. A pack→pile transfer therefore *mints* a rot
     window exactly as a villager's own drop does; there is no original schedule. The rule
     above is what `agent.dropped` (`internal/sim/state.go`) actually does.
   - everything else: decrement the inventory field, increment the pile's

Conservation invariant: `units(inventory) + units(tile pile)` is unchanged by the arm.

### Batch composition (`BuildMiracleBatch`)

```
take_item →
  guardian.item_taken{Agent, Kind, Qty, Gratis}
  agent.memory_added{Agent, Text: takeMemoryText(qty, item), Salience: SalDream, Subject: Ref(-1), Origin: OriginOmen}
```

`takeMemoryText` is fixed and deterministic, the `grantMemoryText` mirror:

```
"An unseen hand lifted %d %s from your pack and set it at your feet."
```

Both doors — the guardian's `landMiracle` and the IPC operator `miracle` command — compose
through this one builder (spec 016 R6).

## 4. Guidance surfaces

- `tool.guardianToolDesc` gains an `inspect_pack` entry under the read heading beside
  `survey_site`, and a `take_item` line in the miracle-kind gloss.
- `give_item`'s existing carry-headroom hint (`internal/tool/derive.go:259`) gains a clause
  naming `take_item` as the remedy for a full pack.
- `internal/skin/skin.go` gains a `skin.guardian.example_ask.inspect_pack` entry, matching the
  `survey_site` example-ask precedent.

## 5. Rendering

`internal/tui/digest.go` gains, mirroring the two `guardian.item_granted` entries it already
carries:

- a **line renderer** (`digest.go:1425`'s neighbour) producing a plain-language summary in the
  chronicle's voice — `<Name> — an unseen hand took N kind`. It must not fall through as an
  unknown event type.
- a **subject resolver** (`digest.go:2120`'s neighbour) so the event attributes to the affected
  villager in the feed's subject column, exactly as a grant does.
