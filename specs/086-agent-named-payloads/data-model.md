# Data model: agent-named payloads (spec 086)

Normative companion to [spec.md](spec.md); decisions derived in
[research.md](research.md). File:line citations are against the task
worktree.

## §1 — AgentRef (the one wire shape)

```go
// internal/sim/agentref.go (new)
type AgentRef struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}
```

- **Marshal**: the plain struct marshal — `{"id":2,"name":"Cedar"}`, fixed
  field order (the canonical-JSON convention: structs, never maps). No
  custom MarshalJSON.
- **Unmarshal** (custom, dual-shape, forever): a bare JSON number `2`
  decodes as `{ID: 2, Name: ""}` (the legacy pre-086 shape); an object
  decodes field-wise. `[]AgentRef` therefore decodes both `[1,4]` and
  `[{"id":1,…},{"id":4,…}]` element-wise with no extra code.
- **Semantics**: `Name` is a denormalized copy of the roster name at
  emission (`AgentNames[ID]`, `internal/sim/agents.go:18-20` — a
  package-level constant, so emission-time and replay-time names provably
  agree). `Name == ""` means "legacy row or sentinel" and renderers fall
  back to replica lookup.
- **Sentinels**: `ID` outside `[0, agentCount)` (canonically −1: any/none/
  personal) is legal and carries `Name == ""`.

## §2 — Constructors (the only sanctioned builders)

```go
func Ref(i int) AgentRef        // {i, AgentNames[i]} in-roster; {i, ""} otherwise
func Refs(ids []int) []AgentRef // element-wise Ref, nil-safe
```

Pure functions of the index — no state, no liveness check (dead agents
keep their names; R4). Every emitter (executor `emit(...)` sites, mind,
guardian, bundle, persona) constructs refs through these.

## §3 — The payload census (NORMATIVE: every agent-referencing payload and its treatment)

Treatment codes: **REF** = field type flips int→AgentRef (or []int→
[]AgentRef) in place, same json tag; **SPLIT** = state-shared
struct-as-payload gains a payload-mirror type carrying refs while the
state entity keeps ints (§4); **ADD** = additive `omitempty` ref field;
**EXEMPT** = stays bare, allowlisted with rationale (§5).

| # | Payload type (internal/…) | Event type(s) | Agent field(s) → treatment |
|---|---|---|---|
| 1 | IntentSetPayload (sim/agents.go:1366) | agent.intent_set | Agent → REF |
| 2 | WorkStartedPayload (sim/agents.go:1403) | agent.work_started | Agent → REF |
| 3 | RecoveryStalledPayload (sim/agents.go:1421) | agent.recovery_stalled | Agent → REF |
| 4 | HarvestPayload (sim/agents.go:1426) | agent.foraged / chopped / hunted / quarried / collected_water | Agent → REF |
| 5 | BuiltPayload (sim/agents.go:1431) | agent.built | Agent → REF |
| 6 | BuildFailedPayload (sim/agents.go:1444) | agent.build_failed | Agent → REF |
| 7 | NeedsPayload (sim/agents.go:1449) | agent.needs_changed | Agent → REF |
| 8 | DiedPayload (sim/agents.go:1457) | agent.died | Agent → REF |
| 9 | NeglectDetectedPayload (sim/agents.go:1469) | sim.neglect_detected | Agent → REF |
| 10 | TalkedPayload (sim/agents.go:1475) | agent.talked | A, B → REF |
| 11 | MemoryAddedPayload (sim/agents.go:1483) | agent.memory_added | Agent, Subject (−1 personal) → REF |
| 12 | MemoryEmbeddedPayload (sim/agents.go:1500) | agent.memory_embedded | Agent → REF |
| 13 | SituationEmbeddedPayload (sim/agents.go:1512) | agent.situation_embedded | Agent → REF |
| 14 | ThoughtPayload (sim/agents.go:1519) | agent.thought | Agent → REF |
| 15 | HailedPayload (sim/agents.go:1527) | social.hailed | From, To → REF |
| 16 | HailMetPayload (sim/agents.go:1532) | social.hail_met | From, To → REF |
| 17 | HailExpiredPayload (sim/agents.go:1536) | social.hail_expired | From, To → REF |
| 18 | CraftedPayload (sim/agents.go:1547) | agent.crafted | Agent → REF |
| 19 | AtePayload (sim/agents.go:1554) | agent.ate | Agent → REF |
| 20 | CookedPayload (sim/agents.go:1563) | agent.cooked | Agent → REF |
| 21 | BathedPayload (sim/agents.go:1572) | agent.bathed | Agent → REF |
| 22 | RefueledPayload (sim/agents.go:1579) | agent.refueled | Agent → REF |
| 23 | SpearBrokePayload (sim/agents.go:1587) | agent.spear_broke | Agent → REF |
| 24 | AxeBrokePayload (sim/agents.go:1594) | agent.axe_broke | Agent → REF |
| 25 | WallWorkPayload (sim/agents.go:1609) | agent.wall_chipped / wall_destroyed / wall_repaired | Agent → REF |
| 26 | DroppedPayload (sim/agents.go:1623) | agent.dropped | Agent → REF |
| 27 | PickedUpPayload (sim/agents.go:1632) | agent.picked_up | Agent → REF |
| 28 | DepositedPayload (sim/agents.go:1641) | agent.deposited | Agent → REF |
| 29 | WithdrewPayload (sim/agents.go:1651) | agent.withdrew | Agent, Owner → REF |
| 30 | AgentMovedPayload (sim/state.go:378) | agent.moved | Agent → REF |
| 31 | AgentPayload (sim/state.go:383) | agent.intent_done / slept / woke | Agent → REF |
| 32 | RunEndedPayload (sim/state.go:441) | run.ended | Deaths[].Agent → SPLIT (DeathRecord is the state ledger type; §4) |
| 33 | RelationChangedPayload (sim/social.go:108) | social.relation_changed | A, B → REF |
| 34 | GavePayload (sim/social.go:115) | social.gave | From, To → REF |
| 35 | RumorToldPayload (sim/social.go:123) | social.rumor_told | From, To, Subject → REF |
| 36 | SecretSeededPayload (sim/social.go:133) | social.secret_seeded | Agent → REF |
| 37 | ConversationTurnPayload (sim/social.go:138) | social.conversation_turn | Speaker, Listener → REF |
| 38 | ConversationPayload (sim/social.go:144) | social.conversation | A, B, Participants → REF (Tones stays []int — per-participant sentiment, not agents) |
| 39 | ChestTakenPayload (sim/social.go:160) | social.chest_taken | Owner, Taker → REF |
| 40 | JournalWrittenPayload (sim/journal.go:269) | journal.entry_written | Agent → REF |
| 41 | JournalDeletedPayload (sim/journal.go:274) | journal.entry_deleted | Agent → REF |
| 42 | MemoryPromotedPayload (sim/consolidate.go:162) | agent.memory_promoted | Agent → REF |
| 43 | MemoryFadedPayload (sim/consolidate.go:169) | agent.memory_faded | Agent → REF |
| 44 | BeliefRevisedPayload (sim/consolidate.go:183) | agent.belief_revised | Agent, Source, Subject → REF |
| 45 | BeliefReinforcedPayload (sim/consolidate.go:204) | agent.belief_reinforced | Agent → REF (BeliefID is not an agent) |
| 46 | NarrativeSetPayload (sim/consolidate.go:209) | agent.narrative_set | Agent → REF |
| 47 | ConsolidatedPayload (sim/consolidate.go:214) | agent.consolidated | Agent → REF |
| 48 | SawPayload (sim/mentalmap.go:489) | agent.saw | Agent → REF; nested PlaceFact.Source → EXEMPT (§5.1) |
| 49 | PlaceToldPayload (sim/mentalmap.go:501) | social.place_told | From, To → REF; PlaceFact.Source → EXEMPT |
| 50 | PlaceRevealedPayload (sim/mentalmap.go:567) | metatron.place_revealed | Agent → REF; PlaceFact.Source → EXEMPT |
| 51 | MapCorrectedPayload (sim/mentalmap.go:579) | agent.map_corrected | Agent → REF; PlaceFact.Source → EXEMPT |
| 52 | CogThoughtPayload (sim/cognition.go:60) | cog.thought | Agent → REF |
| 53 | CogOutcomePayload (sim/cognition.go:85) | cog.outcome | Agent → REF |
| 54 | IntentRejectedPayload (sim/cognition.go:108) | agent.intent_rejected | Agent → REF |
| 55 | MemoryDivergencePayload (sim/cognition.go:190) | cog.memory_divergence | Agent → REF |
| 56 | PlanSetPayload (sim/plan.go:44) | agent.plan_set | Agent → REF |
| 57 | PlanStepPayload (sim/plan.go:51) | agent.plan_step_started / plan_expired | Agent → REF |
| 58 | Directive (sim/plans.go:99, struct-as-payload) | directive.issued | Targets → SPLIT (§4) |
| 59 | DirectiveFulfilledPayload (sim/plans.go:115) | directive.fulfilled | Targets → REF (payload-only type; not state-resident) |
| 60 | GuardianNudgedPayload (sim/guardian.go:53) | metatron.nudged | Targets → REF |
| 61 | GuardianOrder (sim/guardian.go:178, struct-as-payload) | metatron.order_placed | Agent (−1 = any) → SPLIT (§4) |
| 62 | Prophecy (sim/prophecy.go:60, struct-as-payload) | prophecy.declared | Targets, Claim.Agent → SPLIT (§4) |
| 63 | GruSightedPayload (sim/gru.go:289) | gru.sighted | Agent → REF |
| 64 | GruAttackedPayload (sim/gru.go:294) | gru.attacked | Agent → REF |
| 65 | MeetingOpenedPayload (sim/governance.go:235) | meeting.opened | Attendees → REF |
| 66 | TurnTakenPayload (sim/governance.go:238) | meeting.turn_taken | Agent → REF |
| 67 | ProposalPayload (sim/governance.go:244) | meeting.proposal_tabled | Target (−1 sentinel), Proposer → REF |
| 68 | ProposalResolvedPayload (sim/governance.go:253) | meeting.proposal_resolved | embedded ProposalPayload + Yeas, Nays → REF |
| 69 | NormViolatedPayload (sim/governance.go:268) | norm.violated | Violator, Witnesses → REF |
| 70 | ItemGrantedPayload (sim/miracles.go:35) | metatron.item_granted | Agent → REF |
| 71 | MorgueEpiloguePayload (sim/morgue.go:24) | morgue.epilogue | Agent → REF |
| 72 | ChronicleEntryPayload (sim/chronicle.go:18) | chronicle.entry | Agents → REF |
| 73 | FaithChangedPayload (sim/faith.go:78) | faith.changed | SourceID → EXEMPT (§5.4); gains `Agent *AgentRef json:"agent,omitempty"` → ADD (set iff reason villager_died) |
| 74 | CogToolCallPayload (sim/cognition.go:123) | cog.tool_call | Args → EXEMPT (§5.2 — opaque recorded tool args) |
| 75 | WorldMigratedPayload (sim/state.go:410) | world.migrated | embeds full sim.State → EXEMPT by the R2 invariant (State carries no refs; shape untouched) |

**Census size**: 69 payload types migrate (66 REF + 3 SPLIT via new
mirror types, plus RunEnded's Deaths mirror), covering ~85 event types;
2 exemptions with additive coverage (73) or none (74, 75); ~127 emission
call sites rewritten to constructors (~95 `internal/sim`, ~22
`internal/mind`, ~10 `internal/guardian`, 3 `internal/bundle` +
`internal/persona`). Non-agent payloads (Regrown, FireBurnedOut,
FoodRotted, Promise, curriculum, stranger.*, scenario, tuning, clock/
governor/daemon/world lifecycle, meeting place/convention/rephrased/
closed, OrderTriggered/OrderID, CharterObserved, SkillsObserved,
GruEmerged/Moved/Withdrew, TimeSnapped/EntityMoved/EntityRemoved,
Recalibration, GuardianReportCard, Designation) are untouched —
**stranger payloads verified field-by-field: no villager index anywhere**
(`internal/sim/stranger.go:232-252`).

## §4 — The SPLIT types (state-shared struct-as-payload; the R2 hash invariant)

State entities keep `int`/`[]int`; a payload-mirror type owns the wire.
The mirror carries the SAME json tags, so the wire shape is uniform with
§3 and legacy rows decode through the same dual-shape unmarshal.

| Wire event | Mirror payload type (new) | State entity (unchanged) | Arm behavior |
|---|---|---|---|
| directive.issued | DirectiveIssuedPayload — Directive's fields with `Targets []AgentRef` | `Directive.Targets []int` (sim/plans.go:99) | fold `.ID`s; ignore names |
| metatron.order_placed | OrderPlacedPayload — GuardianOrder's fields with `Agent AgentRef` | `GuardianOrder.Agent int` (sim/guardian.go:178) | fold `.ID`; −1 sentinel unchanged |
| prophecy.declared | ProphecyDeclaredPayload — Prophecy's fields with `Targets []AgentRef` and a claim mirror carrying `Agent AgentRef` for the survives kind | `Prophecy` + `ProphecyClaim` (sim/prophecy.go:60,77) | fold `.ID`s; claim normalization/equality unchanged (over IDs) |
| run.ended | RunEndedPayload.Deaths becomes `[]DeathRef` — DeathRecord's fields with `Agent AgentRef` | `DeathRecord.Agent int` (state.go:429; the state death ledger) | payload is emission-side copy; ledger built by agent.died arms, untouched |

Rules: (a) the mirror is defined NEXT to its entity with a doc comment
naming the R2 invariant; (b) the injection door (`InjectSocial` dry-run)
validates the mirror exactly as it validated the entity-as-payload; (c)
`TestNoAgentRefInState` (§5) is the tripwire that keeps future authors
from re-merging them.

## §5 — Enforcement (AC #3): catalog, sweep, allowlist, append validation

**`sim.PayloadCatalog`** (new, `internal/sim/payloads.go`):
`map[string]func() any` — every event type that carries a struct payload →
its zero payload value (mirror types for §4 rows; `struct{}{}` types and
prose-only rows included for exhaustiveness). Completeness is doc-anchored:
a sim-side test parses the backticked event types in
`docs/wiki/event-types.md` (the `TestCatalogSweep` trick,
`internal/tui/digest_test.go:322-330`) and requires each in the catalog;
the tui side asserts `catalogFixture` keys ⊆ `sim.PayloadCatalog` keys.

**`TestPayloadAgentRefSweep`** (`internal/sim`): for every catalog entry,
reflect over the payload type (fields, embedded structs, slices,
pointers). Any `int` / `[]int` / `*int` field whose json tag (first
segment) is in the frozen vocabulary:

```
agent a b from to speaker listener subject owner taker violator
witnesses targets attendees proposer target yeas nays agents
participants source src
```

must instead be `AgentRef` / `[]AgentRef` / `*AgentRef` — unless
(type, field) is on the frozen allowlist below. Every allowlist entry
carries its rationale string in the test source; removing a use without
removing its entry fails the sweep (no dead exemptions).

**The allowlist** (complete):
1. `PlaceFact.Source` (`src,omitempty`) — state-resident mental-map fact
   (R2); provenance index, not a chronicle surface. Applies wherever
   PlaceFact nests (payloads 48–51).
2. `CogToolCallPayload.Args` — opaque `json.RawMessage` of the tool's own
   args; recorded observability, never rewritten.
3. `IntentSetPayload.Source` (`source,omitempty`) — a STRING
   ("reflex"|"planner"), colliding with the vocabulary by tag only; listed
   so the collision is a recorded decision. (The sweep only flags int-ish
   fields, so this entry is documentation-grade.)
4. `FaithChangedPayload.SourceID` — string source-event id (encodes an
   agent index only for villager_died); covered by the additive
   `Agent *AgentRef` (§3 row 73).

**`TestNoAgentRefInState`** (`internal/sim`): reflection walk over
`sim.State`'s reachable type graph; any `AgentRef` occurrence fails. This
is the R2 hash-stability invariant as a standing tripwire.

**Append validation** (live emission only — NEVER in Apply arms, R3):
- `mustPayload` (`internal/sim/state.go:460`): reflection walk over the
  value; every `AgentRef` with `ID` in `[0, agentCount)` must have
  `Name == AgentNames[ID]`; panic otherwise (the existing marshal-error
  contract class). Sentinel IDs must have `Name == ""`.
- `InjectSocial` (`internal/sim/loop.go:382`): before the dry-run, decode
  each event's payload via `PayloadCatalog[e.Type]()` and run the same
  walk; refuse the batch on violation (the door's existing refusal
  surface). Whitelisted types absent from the catalog are a test failure,
  not a runtime skip.
- `TestCatalogSweep` (`internal/tui/digest_test.go:296`): every
  agent-bearing `catalogFixture` payload rewritten to the named shape; NEW
  assertion — agent-bearing fixtures digest identically with `names = nil`
  (the AC #2 no-replica proof).

## §6 — Back-compat matrix (AC #4; documented behavior)

| Input | Decoder behavior | Renderer behavior |
|---|---|---|
| pre-086 row, `"agent":2` | dual-shape unmarshal → `{2, ""}`; arm reads `.ID` — identical fold | Name "" → fall back to replica `names[2]` (`agentName`, grammar.go:150); grammar-miss rows still via `resolvePayloadNames` regex |
| post-086 row, `{"id":2,"name":"Cedar"}` | object branch; arm reads `.ID` | payload name rendered directly; no replica needed |
| pre-086 snapshot | state shapes unchanged; decodes as today | n/a |
| migrated v1–v4 world | `world.migrated` embeds State — shape untouched; migration never rewrites payloads (snapshot-cut, `docs/wiki/world-migration.md`) | historic rows fall back as above |
| pre-086 log, from-genesis replay | byte-identical `State.Marshal()`/`Hash()` — no state shape or fold change anywhere | n/a |
| live pre-086 world continued under 086 | new emissions named from the next event on; old rows untouched (append-only triggers) | mixed feed: new rows payload-named, old rows fallback |

## §7 — TUI consumption (AC #2 + the hit-rate claim)

- **Digest naming**: agent-bearing `digestFunc`s read `ref.Name`,
  falling back to `agentName(names, ref.ID)` iff `Name == ""`. The
  `names []string` parameter survives as the fallback channel.
- **`resolveSubject` generic fallback** (`digest.go:2103`): registry
  first (unchanged, ~80 types); on registry miss, scan the payload JSON
  for `{"id":…,"name":…}` objects; exactly ONE distinct in-roster id →
  subject candidate (live position via `liveAgentPos`); zero or several →
  unlocatable (the honest-hint doctrine, structurally). `world.migrated`
  stays hard-excluded (embedded State is never scanned — the existing
  deliberate absence, digest.go:1524-1534).
- **Measurement**: a test drives every rewritten `catalogFixture` row
  through `resolveSubject` with a live replica fixture and asserts
  (a) locatable-count(registry+fallback) > locatable-count(registry-only),
  and (b) pins newly-locatable types (at minimum: `journal.entry_written`,
  `morgue.epilogue`, `cog.thought` — single-ref payloads with no registry
  entry today).

## §8 — The reverse-jump rider (control contract)

| Surface | Control | Effect | Keys+mouse (doc cell) | Oracle entry |
|---|---|---|---|---|
| villager strip (`villagerStripView`, views.go:2626) | villager glyph | center map camera on that villager (`centerCameraOn(a.X, a.Y)`) | `— · click glyph` (keyboard path: villagers tab `J`; noted in prose) | `panels/villager-strip.md` / `reverse-jump` / `click glyph` |
| villagers tab roster (`villagerRosterBody`, views.go:3174) | roster row | select row + center camera; narrow: active pane → map | `J · click row` | `panels/villagers.md` / `reverse-jump` / `click row` |

Mechanics (R7): renderer-recorded hit regions (`stripHit`, `rosterHit` —
the chronHit pointer pattern, invalidated at frame top); `handleMouse`
(look.go:906) gains both branches; glyph column arithmetic
`lipgloss.Width(count) + 1 + 2*i` at screen row `headerRows` (widescreen
only, `computeRows(…).VillagerStrip > 0`); roster rows 4 lines wide / 1
narrow. Dead villagers jump to their grave coordinates; empty replica =
no-op. `J` binds in villagers mode (roster and detail views), documented
in `patterns/keymap.md`'s villagers table. Standing resolution 1
(`views.go:2614`) and `villager-strip.md`'s display-only claims are
amended in the same PR; `check-tui-design --changed` re-verifies
`panels/villager-strip.md`, `panels/villagers.md`, `patterns/keymap.md`.
