# Contract: the designation/directive event vocabulary (spec 084)

NORMATIVE for the seven new event types: payload shapes, doors, reducer
validation, whitelist/observable membership, and the downstream contracts
(TASK-118 faith, standing-order composition). Payload-grammar conventions
follow `docs/wiki/event-types.md`; entity field semantics are
[../data-model.md](../data-model.md) §2–§3.

## 1. Type-by-type

### `designation.placed` — injected

- **Emitter**: `place_designation` handler (`internal/guardian`), via
  `InjectSocial`.
- **Payload**: the full `sim.Designation` struct. `Status` and
  `PlacedSeq` are IGNORED on the wire — the reducer lands `active` and
  stamps `PlacedSeq` from `e.Seq` (the spec-054 apply-time-stamp
  precedent; identical live and in replay; the dry-run probe applies
  with Seq 0 and is discarded).
- **Reducer validation (validate-not-clamp — the dry-run is the door)**:
  non-empty id, unused by any past designation regardless of status;
  `kind` ∈ {settlement_zone, structure_site, wall_line}; locus form
  matches kind (point/line/rect per data-model §1) and is normalized
  (rect: X≤X2 ∧ Y≤Y2; line: axis-aligned; point: X2==X ∧ Y2==Y);
  in-bounds against map dims; size bounds (rect area ≤256, line length
  ≤32); occupancy fulfillability (structure_site: no different-kind
  structure at the tile; wall_line: no non-wall structure on any line
  tile); `structure_kind` required and a real buildable kind for
  structure_site, empty-or-wall-kind for wall_line, empty for
  settlement_zone; `min_structures` 1..12 for settlement_zone (0 for
  others); `label` ≤80 runes; fewer than 16 already-active designations.
- **Effect**: append (then prune: active + most recent 32 non-active) AND
  upsert one `PlaceFact{Kind:"designation", Provenance: revealed, Seen:
  e.Tick}` at the anchor tile into every living villager's mental map
  (map-less agents skipped; reducer total).

### `designation.cancelled` — injected

- **Emitter**: `cancel_designation` handler.
- **Payload**: `{ "id": string }` (the `OrderIDPayload` shape).
- **Reducer**: `active → cancelled` through the one-way transition door;
  unknown id or non-active status refused (race resolution: exactly one
  terminal ever lands).

### `designation.fulfilled` — executor-emitted (NOT whitelisted)

- **Emitter**: the `stepEvents` sweep, once, when the kind's fulfillment
  predicate (data-model §6) holds for an active designation — a pure
  function of (state, tick), the `charge_regenerated`/`order_expired`
  precedent: replay reproduces it with no guardian running; injection of
  this type is refused by whitelist absence.
- **Payload**: `{ "id": string }`.
- **Reducer**: re-validates the predicate against current state (the
  door stays authoritative), then `active → fulfilled`.

### `directive.issued` — injected, with companions

- **Emitter**: `issue_directive` handler, via `InjectSocial`,
  ATOMICALLY with one companion `agent.memory_added` per target in the
  same batch ("The Guardian charges you: <text>" — the vision-memory
  shape; `Origin: OriginOmen`-class provenance, salience at the dream
  band). The whole batch lands or nothing does.
- **Payload**: the full `sim.Directive` struct; `Status`/`PlacedSeq`
  ignored and reducer-stamped.
- **Reducer validation**: non-empty unused id; `designation_id` names an
  ACTIVE designation; `targets` non-empty, ascending, unique, in-range,
  every index living at apply; `text` 1..400 runes; TTL
  (`ExpiresTick - IssuedTick`) within 1..7 game days (the shared
  `GuardianOrderTTL*` constants); fewer than 3 already-active
  directives.

### `directive.cancelled` — injected

- **Emitter**: `cancel_directive` handler.
- **Payload**: `{ "id": string }`.
- **Reducer**: `active → cancelled`, one-way door.

### `directive.fulfilled` — executor-emitted (NOT whitelisted)

- **Emitter**: the sweep, once, when an active directive's bound
  designation has `Status == "fulfilled"`.
- **Payload** (THE TASK-118 SEAM — see §3):

```json
{
  "id": "dir-3600-0",
  "designation_id": "dsg-100-0",
  "targets": [1, 4],
  "issued_tick": 3600
}
```

- **Reducer**: validates the bound designation is `fulfilled`, then
  `active → fulfilled`.

### `directive.expired` — executor-emitted (NOT whitelisted)

- **Emitter**: the sweep, once, when for an active directive
  `nextTick >= ExpiresTick` OR no targeted villager remains alive (the
  all-dead clause is a pure state check, no TTL wait).
- **Payload**: `{ "id": string }`.
- **Reducer**: validates the same disjunction, then `active → expired`.

## 2. Membership tables

| Type | `injectSocialWhitelist` | `observableEventTypes` | ended-world prose whitelist |
|---|---|---|---|
| `designation.placed` | ✅ | ❌ (v1 — watch `directive.*` instead) | ❌ |
| `designation.cancelled` | ✅ | ❌ | ❌ |
| `designation.fulfilled` | ❌ (executor-emitted) | ❌ | ❌ |
| `directive.issued` | ✅ | ✅ | ❌ |
| `directive.cancelled` | ✅ | ✅ | ❌ |
| `directive.fulfilled` | ❌ (executor-emitted) | ✅ | ❌ |
| `directive.expired` | ❌ (executor-emitted) | ✅ | ❌ |

`observableEventTypes` grows 12 → 16 (enum-only; `monitorAndActSchema`
shape and `matchOrders` untouched — AC #7's zero-new-trigger-code
guarantee is structural). On an ENDED world nothing new lands: the
injected types are refused by the narrowed door; the sweeps never run
(`stepEvents` emits nothing after run end).

## 3. Downstream contracts

- **TASK-118 (faith accounting) reads `directive.fulfilled`**: the
  payload above is the binding surface — `id` (dedupe),
  `designation_id` (what was achieved), `targets` (who was bound; credit
  attribution), `issued_tick` (with `e.Tick`, the time-to-fulfil
  window). Faith consumes RECORDED events only; this spec adds no faith
  fields, no scoring, no state.
- **TASK-158 (missions)**: builds ON directives (accept / decompose /
  pursue / report); nothing here reserves or blocks — a mission will
  bind directives the way directives bind designations. Out of scope,
  gated on TASK-112.
- **Standing orders (spec 029)**: `monitor_and_act` may watch the four
  `directive.*` types immediately; e.g. a re-issue loop is
  `event_types:["directive.expired"]` + action "issue it again" — the
  in-game workaround the operator ruling names.
- **Chronicle/TUI digests**: all seven types get digest-grammar rows
  (`internal/tui/grammar.go`); `TestCatalogSweep`
  (`internal/tui/digest_test.go:255`) is the gate.

## 4. Compatibility

- New event types + `omitempty` state slices: **no format bump** (the
  spec-029 precedent — a pre-084 snapshot unmarshals to nil slices;
  pre-084 logs replay byte-identically since no old event's semantics
  change).
- Shrinking or renaming any of the seven after merge is BREAKING
  (recorded vocabulary, the spec-052 frozen-identifier doctrine).
- Widening `observableEventTypes` further (e.g. `designation.*`) is a
  compatible enum-only follow-up.
