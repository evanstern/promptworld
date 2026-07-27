# Data model: faith-driven charge regeneration (spec 085)

NORMATIVE for the faith state shape, the four new event types, the faith
delta table, the prophecy entity, the claim predicate table, the regen
band table, and the wire fields. Payload-grammar conventions follow
`docs/wiki/event-types.md`; entity discipline follows
`specs/084-guardian-directives/data-model.md` (the `GuardianOrder`
clone family).

## 1. FaithState

```go
// FaithState is the village's event-sourced devotion score (spec 085).
// nil State.Faith means the genesis default — the State.Tuning
// nil-means-default precedent; pre-085 snapshots round-trip
// byte-identically (omitempty, no format bump).
type FaithState struct {
    Score int `json:"score"` // clamped 0..100 at fold time
}
```

- `State.Faith *FaithState` (`json:"faith,omitempty"`).
- **Genesis**: `FaithGenesis = 50` — deliberately inside the steady
  band (§6) so a world that has never folded a faith event regenerates
  byte-identically to pre-085. No boot seeding event (unlike
  `sim.tuning_applied`): there is nothing operator-authored to record;
  nil + accessor is the whole compat story.
- **Accessor**: `func (s *State) FaithScore() int` — nil-safe, returns
  `FaithGenesis` when `Faith == nil`. The ONLY read path (reducer,
  executor, daemon status, tests).
- Materialized (`&FaithState{...}`) by the first `faith.changed` fold.
- No tick fields → untouched by `rebaseTicks`.

## 2. Event vocabulary (four new types)

| Type | Payload | Emitter | Whitelisted | Observable | Reducer effect |
|---|---|---|---|---|---|
| `faith.changed` | `FaithChangedPayload{delta int, reason string, source_id string}` | executor faith sweep (§3) | ❌ (executor-emitted, the `charge_regenerated` class — injection refused by whitelist absence) | ❌ (v1; widening later is compatible enum-only) | validates reason ∈ §3 vocabulary and delta ≠ 0 with the reason's sign; folds `Score = clamp(Score+delta, 0, 100)`, materializing `Faith` |
| `prophecy.declared` | the full `sim.Prophecy` struct (`Status`/`PlacedSeq` IGNORED on the wire — reducer lands `active`, stamps `PlacedSeq` from `e.Seq`; the spec-054 apply-time-stamp precedent) | `prophesy` handler (`internal/guardian`), via `InjectSocial`, atomically with one companion `agent.memory_added` per living target (`OriginOmen`, dream-band salience — the vision/directive companion shape); whole batch lands or nothing | ✅ | ✅ | validate-not-clamp (§4 door table), then append + prune (active + most recent 32) |
| `prophecy.fulfilled` | `{ "id": string }` (`OrderIDPayload`) | executor verification sweep, once, when the claim's fulfil condition holds (§5) — pure over (state, tick); emits companion `OriginReport` memories per living target in the same batch | ❌ | ✅ | re-validates the fulfil condition against current state, then `active → fulfilled` |
| `prophecy.failed` | `{ "id": string }` | executor verification sweep, once, when the fail condition holds and the fulfil condition does not (fulfil checked FIRST at a shared boundary — exactly one terminal ever lands); companion `OriginReport` memories likewise | ❌ | ✅ | re-validates the fail condition, then `active → failed` |

- `metatron.charge_regenerated` is **unchanged** (type, empty payload,
  reducer arm `guardian.go:245-247`) — only its firing cadence moves (§6).
- Digest grammar: all four get in-fiction rows (`internal/tui/grammar.go`,
  `TestCatalogSweep`). Faith rows speak of devotion/doubt, never numbers
  first ("The village's faith deepens." / "Faith wavers in the village.").
- Ended world: sweeps never run (`stepEvents` emits nothing after run
  end); `prophecy.declared` refused by the ended-world narrowing.
- Shrinking/renaming any of the four after merge is BREAKING (recorded
  vocabulary, the spec-052 frozen-identifier doctrine).

## 3. Faith delta table (one home: `internal/sim/faith.go`)

All constants promoted-dial-READY (named, one place, spec-059
survival-band discipline) — deliberately NOT in `tuning.json`; dials are
earned by evidence.

| Reason | Source event (scanned in batch) | `source_id` | Delta | Doctrine |
|---|---|---|---|---|
| `directive_fulfilled` | `directive.fulfilled` | directive id | **+8** | the primary endogenous source (operator realignment 2026-07-26); the 157 seam payload is the binding surface |
| `directive_expired` | `directive.expired` | directive id | **−4** | the guardian's charge went unachieved (incl. all-targets-dead) — mild: half a death |
| `villager_died` | `agent.died` | agent index (decimal string) | **−6** | the flock's suffering erodes faith; the deliberate spiral feeder — one event per death |
| `prophecy_fulfilled` | `prophecy.fulfilled` | prophecy id | **+12** | the declared word came true — the strongest single faith act |
| `prophecy_failed` | `prophecy.failed` | prophecy id | **−15** | false prophet; asymmetric vs +12 so claim-spam is negative-EV |

**Sweep contract** (`faithEvents(s, batch, nextTick)`, in `stepEvents`):

- Position: AFTER every faith-source emitter (directive/prophecy sweeps,
  the needs heartbeat, gru), BEFORE `scenarioRubricEvents` and run-end
  detection — the batch-scanning idiom of both (`executor.go:401-446`).
- Pure over (pre-tick state, this batch, tick). One `faith.changed` per
  source event, in the batch's own order.
- **Emission gate**: skip any emission whose fold could not move the
  clamped score computed from the pre-tick score plus this batch's
  PRIOR faith emissions folded in order (the `charge_regenerated`
  below-cap idiom: never record a movement that moves nothing; partial
  clamps still emit the doctrine delta — the fold clamps).
- Excluded by decision (research R3): `designation.fulfilled`, ambient
  accrual, time decay, tutoring (TASK-112 AC #6), `metatron.nudged`.

## 4. Prophecy

```go
// Prophecy is a charge-priced, uncancellable, deadline-bounded declared
// claim (spec 085) — the GuardianOrder/Designation entity discipline.
type Prophecy struct {
    ID           string       `json:"id"`            // "pro-<tick>-<seq>", nextOrderID shape, no RNG
    Targets      []int        `json:"targets"`       // living at declaration; ascending, unique
    Village      bool         `json:"village,omitempty"` // declared to "everyone"
    Text         string       `json:"text"`          // ≤ the registry TextCapBytes (400) — NudgeTextMax's single source
    Claim        ProphecyClaim `json:"claim"`        // §5, stored NORMALIZED
    DeclaredTick int64        `json:"declared_tick"` // history → rebase KEEP
    DeadlineTick int64        `json:"deadline_tick"` // future deadline → rebase SHIFT while active
    Status       string       `json:"status"`        // active → fulfilled | failed (one-way)
    PlacedSeq    int64        `json:"placed_seq,omitempty"` // reducer-stamped from e.Seq
}

// ProphecyClaim is the discriminated, machine-checkable claim.
type ProphecyClaim struct {
    Kind          string `json:"kind"` // §5 closed vocabulary
    DesignationID string `json:"designation_id,omitempty"`
    StructureKind string `json:"structure_kind,omitempty"`
    Min           int    `json:"min,omitempty"`
    Agent         int    `json:"agent,omitempty"`
}
```

- `State.Prophecies []Prophecy` (`json:"prophecies,omitempty"`) — nil on
  every pre-085 snapshot; no format bump.
- Constants: `GuardianProphecyCap = 3` (active), retention prune active
  + most recent 32 (the `pruneGuardianOrders` discipline shape), deadline
  TTL bounds shared `GuardianOrderTTLMinDays..MaxDays` (1..7 game days,
  default 3) — shared constants, never copies.
- **No cancel verb** (research R8). **No all-targets-dead expiry**: the
  claim is a world fact; only §5's conditions terminate it.

**`prophecy.declared` door (validate-not-clamp; the dry-run is the door)**:
non-empty id unused by any past prophecy regardless of status; targets
non-empty, ascending, unique, in-range, every index living at apply;
text 1..400 runes; TTL (`DeadlineTick − DeclaredTick`) within 1..7 game
days; fewer than `GuardianProphecyCap` active prophecies; claim kind in
the closed vocabulary with kind-required fields present, valid, and
normalized (unknown/extra fields refused); **claim not already true**
(the kind's fulfil condition must NOT hold at apply — prophesying the
past is refused); **no active duplicate** (normalized-claim equality
with any `active` prophecy is refused). `Status`/`PlacedSeq` ignored and
reducer-stamped.

**`prophesy` tool** (`internal/tool/registry.go` → `guardianTools`):
`Gate: Charge` (1 — the `send_vision` price), Effect the influence
class; params `targets` (Text — name list or "everyone"), `text` (Text,
the registry text cap), `claim_kind` (Enum over §5),
`designation_id`/`structure_kind`/`min`/`agent` (kind-conditional,
handler-validated, partial-args refused — the parseReveal shape),
`deadline_days` (Number, optional, default 3, door-bounded 1..7).
Stage availability follows `send_vision`'s profile.

## 5. Claim predicate table (NORMATIVE — the verification rule)

**The rule**: a vision is 'true' exactly when its recorded claim —
declared before the fact, from this closed vocabulary — is satisfied by
recorded world state within its deadline, judged by these pure
(state, tick) conditions in the executor sweep and re-validated by the
reducer arm. Free text is never graded; no model output participates.

Sweep evaluation per active prophecy, slice order, each boundary tick:
**fulfil condition first, then fail condition** (one terminal ever).

| Kind | Required fields | Fulfil condition (pure over state, tick) | Fail condition | Door adds |
|---|---|---|---|---|
| `designation_fulfilled` | `designation_id` | the named designation's `Status == "fulfilled"` | `tick >= DeadlineTick` and fulfil doesn't hold | designation exists (any active-or-past id — the claim may name in-flight work); NOT already `fulfilled` |
| `structure_count` | `structure_kind`, `min` (1..64) | count of structures of `structure_kind` ≥ `min` | `tick >= DeadlineTick` ∧ ¬fulfil | `structure_kind` is a real buildable kind (the spec-084 mirror + drift-test discipline); current count < `min` |
| `population_at_least` | `min` (1..villager cap) | `livingCount(s) >= min` | `tick >= DeadlineTick` ∧ ¬fulfil | current living count < `min` |
| `survives` | `agent` | `tick >= DeadlineTick` ∧ agent alive | agent dead (fail-FAST, no deadline wait) — or `tick >= DeadlineTick` ∧ dead (unreachable given fulfil-first) | agent index valid and living |

Notes:

- By-deadline kinds (rows 1–3) fulfil the moment the condition first
  holds at a sweep boundary ≤ deadline; a condition that becomes true
  AFTER `failed` latched mints nothing (one-way status — the
  "verifies after the TTL" edge).
- `survives` is the at-deadline kind expressed in the same fulfil/fail
  grammar: fail-fast on death, fulfil at the first sweep tick ≥
  deadline. No second verification mode exists.
- A cancelled designation under a `designation_fulfilled` claim needs
  no special case: fulfil can never hold, fail fires at deadline.
- Normalized-claim equality (the duplicate-rejection key): (Kind,
  DesignationID, StructureKind, Min, Agent) after normalization.

## 6. Regen band table (NORMATIVE — the AC #3 curve and the AC #4 posture)

```go
// FaithRegenCadenceTicks is the pure faith → charge-regen cadence
// (spec 085): the executor's boundary check and the daemon's status
// projection both read THIS function — one home, exported so the wire
// value and the sim can never disagree. Returns 0 when no regen is
// scheduled (the scenario forsaken band).
func FaithRegenCadenceTicks(score int, scenario bool) int64
```

| Band | Score | Cadence | In fiction |
|---|---|---|---|
| fervent | ≥ 75 | 4 game hours (`4*3600`) | the village believes; power comes easily |
| steady | 40..74 | **6 game hours** (`chargeRegenTicks` — today's constant; genesis 50 lives here → pre-085-identical schedule) | the old covenant pace |
| wavering | 15..39 | 12 game hours | doubt slows the flow |
| forsaken | < 15 | **scenario: 0 (no regen — the authentic spiral)** · **ambient: 24 game hours (the floor)** | the well nearly dry |

- Executor check (replaces `executor.go:55`):
  `c := FaithRegenCadenceTicks(s.FaithScore(), s.scenario != nil);`
  `if c > 0 && nextTick%c == 0 && s.GuardianCharges < GuardianChargeCap { emit(...) }`
  — absolute boundaries, below-cap gate, same event, same empty payload.
- The posture fork keys on the boot-frozen `s.scenario != nil`
  (replay-re-armed; the spec-054 incident-sweep precedent). The
  **reversal lever** is this one table + the fork argument; the recorded
  future promotion is `tuning.json faith_floor_cadence_ticks` (0 = no
  floor) — a one-table change, no event/shape impact.
- All cadences are divisors of the game day → boundaries stay absolute
  multiples within any band; band changes take effect at the next check
  (pure function of folded score — replay-identical).
- No hysteresis in v1 (research R4 watch item).
- `GuardianChargeRegenTicks` (the exported legacy constant,
  `guardian.go:24-30`) remains exported as the steady-band value — the
  TUI's fallback against a pre-085 daemon (§7).

## 7. Wire fields (IPC status)

`internal/ipc/protocol.go` `ClockStatus` (beside `metatron_charges`):

| Field | JSON | Type | Source | Consumer |
|---|---|---|---|---|
| Faith | `faith,omitempty` | `*int` | `s.FaithScore()` (always non-nil from a spec-085 daemon) | strip faith segment: non-nil → `faith N`; nil → `faith —` (older daemon — the §4 dashed state); CLI/`metatron_status` parity (the D1 projection rule) |
| FaithRegenTicks | `faith_regen_ticks,omitempty` | `int64` | `FaithRegenCadenceTicks(score, scenario)` | strip regen forecast: next absolute boundary of the EFFECTIVE cadence; 0 (or full bank) → segment omitted (the R4.1 honesty rule generalized); absent field → legacy `sim.GuardianChargeRegenTicks` fallback |

Strip drop order under width pressure is unchanged from the shipped
contract: faith drops first, then orders, then regen, then the bank
(`joinStripSegments`, `views.go:2727-2733` — the comment already names
faith first).

## 8. Companion memories

| Moment | Recipients | Origin | Salience | Text shape |
|---|---|---|---|---|
| `prophecy.declared` (atomic batch) | living targets at declaration | `OriginOmen` (direct — a delivered omen) | dream band (`SalDream`, 8 — below `GenerationBumpSalience`) | "The Guardian foretells: <text>" (the vision-memory shape) |
| `prophecy.fulfilled` (sweep batch) | living targets | `OriginReport` (secondhand — word spreads; `DirectPerception` stays honest) | mid band (the `salPlaceTold`/talk class, ~5) | "The Guardian's foretelling came true — <text>" |
| `prophecy.failed` (sweep batch) | living targets | `OriginReport` | mid band, negative tone | "The appointed time passed; the Guardian's word did not come to pass." |

All personal (`Subject: -1`) — no gossip seeding in v1 (guardian-subject
rumors are identified, deferred; `docs/wiki/social-fabric.md` rumor
machinery keys on villager subjects). Situated constructors as always
(`situatedMemory*` — every memory carries a Where and a stamped Origin).

## 9. Rebase taxonomy (`rebaseTicks`, `internal/sim/miracles.go`)

| Field | Class | Action |
|---|---|---|
| `Prophecy.DeadlineTick` (status `active`) | future deadline | SHIFT (the `Directive.ExpiresTick` arm's clone) |
| `Prophecy.DeclaredTick` | history | KEEP (doc comment) |
| `Prophecy.*` (non-active) | settled history | KEEP |
| `FaithState` | no tick fields | untouched (doc comment) |

## 10. Test surface additions (gates)

- `TestCatalogSweep` (`internal/tui/digest_test.go:255`): +4 rows.
- Lessons taxonomy: `first-faith-event` presence asserted; the absence
  pin (`lessons_test.go:119-122`) removed. Trigger:
  `e.Type == "faith.changed"`; `Done`: none; tier mechanics;
  direction-neutral copy with skin tokens.
- Rubric hygiene (`rubric_hygiene_test.go`): `faith.` joins the banned
  prefixes.
- Strip render tests: populated / dashed / forecast-omitted states; the
  `render_test.go:599-600` absence pin flips to a presence pin.
- Regen: band × posture table; boundary/off-boundary drive per band;
  no-faith-event world schedule byte-identity vs pre-085.
- Replay: from-genesis byte-identity over a fixture log carrying every
  faith reason and the full prophecy lifecycle; pre-085 log replay
  byte-identity.
