# Research: agent-named payloads — AgentRef at emission (spec 086)

Decisions verified against the task worktree (`task-17-agent-named-payloads`,
current main including the faith-directives sweep: specs 083–085's
`sim.neglect_detected`, `directive.*`, `faith.*`/`prophecy.*`, `stranger.*`
families). Every citation below is a file:line in this worktree.

## R1 — AgentRef is a struct field REPLACING the bare int, marshaling `{"id":N,"name":"S"}`; not parallel `*_name` fields

**Decision**: `type AgentRef struct { ID int `json:"id"`; Name string
`json:"name"` }` in `internal/sim`. Every agent-referencing payload field
changes type in place — `Agent int `json:"agent"`` becomes
`Agent AgentRef `json:"agent"``, `Targets []int` becomes `[]AgentRef` —
keeping its json tag. Marshal is the plain struct marshal (fixed field
order, no custom MarshalJSON needed). `UnmarshalJSON` is custom and
dual-shape: a bare JSON number decodes as `{ID: n, Name: ""}` (the legacy
shape, forever), a `{"id","name"}` object decodes field-wise.

**Rationale**: the board card ratifies this exact shape ("an AgentRef that
marshals {id, name} … used for every agent-referencing payload field" —
TASK-17 description; constitution I: derive from artifacts, don't re-ask).
A typed ref is also what makes AC #3's enforcement mechanical rather than
conventional: reflection can find every `AgentRef` in a payload value
(append validation, R5) and can find every agent-vocabulary field that is
NOT an `AgentRef` (the sweep test, R5). And the object form is
self-describing for the primary driver — out-of-sim consumers (TASK-18
webhook sinks, exported logs) that have no replica to join `{"agent":2}`
against.

**Evidence**:
- Canonical-JSON convention: payloads are "structs, never maps, so bytes
  are deterministic" (`docs/wiki/event-types.md` — the payload-struct
  convention; `docs/wiki/event-log.md`: "`Event.Payload` is canonical JSON
  (struct-marshaled, fixed field order) so histories are byte-comparable").
  A two-field struct marshal satisfies this exactly.
- No precedent conflicts: NO existing payload carries a name beside an
  index (census, data-model.md §3) — this is a green field, not a
  convention change.
- The dual-shape unmarshal has a house precedent in spirit:
  `ConversationPayload.Participants` "empty means [A, B] (pre-TASK-22
  payloads replay unchanged)" (`internal/sim/social.go:150-152`) — old
  shapes decode meaningfully forever.
- The post-hoc layer this replaces: `resolvePayloadNames`
  (`internal/tui/grammar.go:160`, regex
  `"(agent|a|b|from|to|speaker|listener|subject|owner|taker)":(-?\d+)`)
  and `m.agentNames()` (`internal/tui/tui.go:2122`) rewrite names into
  rendered lines at view time — exactly the replica-dependent lookup AC #2
  retires for new events.

**Alternatives**:
- Parallel additive fields (`agent int` + `agent_name string,omitempty`) —
  rejected: purely additive wire compat buys nothing (no external consumer
  ships yet; every in-repo consumer migrates in this PR), it doubles every
  field, and enforcement degenerates to naming convention — reflection
  cannot distinguish "agent int missing its name twin" from any other int
  mechanically without the same vocabulary list, while the typed ref makes
  the compliant state a compile-visible fact. The board card names AgentRef.
- `type AgentRef int` with a MarshalJSON that derives the name from the
  roster at marshal time — rejected: enforcement-total but couples
  `encoding/json` output to package-global state, erases the possibility of
  per-world rosters later, and hides the name from Go call sites and test
  fixtures. The constructor + validator pair (R4, R5) gets the same
  can't-forget property without the coupling.

## R2 — Hash stability: events are never hashed; state is. Therefore AgentRef must never be reachable from `sim.State`, and the four state-shared payload types split

**Decision**: the determinism/hash invariant this feature must preserve is
STATE bytes, not event bytes. The design rule that preserves it: **no
`AgentRef` is reachable from `sim.State`'s type graph** (reflection-test
enforced, R5). Names are additive on NEW emissions and are never folded
into state: reducer arms read `ref.ID` only. The four types that are
simultaneously event payload and state entity — `Directive`
(`directive.issued`), `GuardianOrder` (`metatron.order_placed`),
`Prophecy` (`prophecy.declared`), and `DeathRecord` (inside
`RunEndedPayload.Deaths` and the state death ledger) — SPLIT: a
payload-mirror struct carries `[]AgentRef`/`AgentRef` on the wire with the
same json tags; the reducer arm folds `.ID`s into the unchanged int-typed
state entity.

**Rationale**: adding a name field anywhere in `State` would change
`Marshal()` bytes and therefore `Hash()`, breaking three contracts at once:
snapshot verification, `world.migrated` (whose payload embeds the entire
canonical `sim.State`), and the pre-086 replay byte-identity in AC #4.
Keeping names strictly wire-side makes pre-086 replay byte-identity
*trivially* true — the state code path is unchanged shape-wise.

**Evidence**:
- Events are never hashed: the only hashes are `State.Marshal()`/`Hash()`
  (`internal/sim/state.go:318-330`, SHA-256 over canonical state bytes) and
  the snapshot's `stateHash` over the STORED state bytes
  (`internal/store/store.go:151-165`, `SaveSnapshot`/`LatestValidSnapshot`)
  — snapshot verification hashes stored bytes, so it never re-marshals
  through new struct shapes. `docs/wiki/sim-state-reducer.md`: "`Hash()` is
  SHA-256 of that, used by snapshots verification and the determinism
  tests. Wall-clock time never appears in state."
- History is immutable: `events_no_update`/`events_no_delete` triggers
  RAISE(ABORT) in-schema (`docs/wiki/event-log.md`) — pre-086 rows keep
  their bytes forever; only NEW emissions carry the new shape. Replay reads
  stored bytes; nothing ever re-marshals a stored event.
- Determinism comparisons are same-build: replay-vs-live runs the same
  `Apply` (`docs/wiki/sim-state-reducer.md`), and two same-seed runs of the
  post-086 build both emit named payloads — event histories stay
  byte-comparable within a build, which is all the doctrine claims.
- `world.migrated` "embeds the entire canonical `sim.State`"
  (`docs/wiki/event-types.md`) — the one payload transitively containing
  agent names today (`Agent.Name`, `internal/sim/agents.go:185`); with
  State untouched its shape is untouched.
- The state-shared types (census): `Directive` struct-as-payload with
  `Targets []int` (`internal/sim/plans.go:99-103`), `GuardianOrder` with
  `Agent int` (−1 = any, `internal/sim/guardian.go:178`), `Prophecy` with
  `Targets []int` + nested `ProphecyClaim.Agent`
  (`internal/sim/prophecy.go:60,77`), `DeathRecord.Agent`
  (`internal/sim/state.go:429`) shared by the state death ledger and
  `RunEndedPayload.Deaths` (`state.go:441`).
- Old snapshots must keep decoding: `daemon.recoverState` unmarshals the
  chosen snapshot into `sim.State` then replays the tail
  (`docs/wiki/snapshots.md`) — unchanged state shapes decode unchanged.

**Alternatives**: additive `omitempty` name fields ON the shared
entity structs, zeroed by the reducer arm before folding — rejected:
works byte-wise but plants a field in state types that must always be
empty in state (a trap for every future arm), and makes the wire format
non-uniform (objects almost everywhere, parallel arrays on exactly four
types). The payload-mirror split keeps ONE wire shape everywhere and keeps
state types honest.

## R3 — Back-compat: the reducer accepts both shapes forever; name validation NEVER lives in Apply arms

**Decision**: `AgentRef.UnmarshalJSON`'s legacy branch IS the back-compat
mechanism — every reducer arm decodes pre-086 (`"agent":2`) and post-086
(`"agent":{"id":2,"name":"Oak"}`) rows through the same struct, reading
only `.ID`. No arm ever validates `Name`: name-presence enforcement lives
exclusively at emission choke points (R5) that replay never traverses.
Renderers keep their post-hoc fallback for historic rows: a decoded ref
with `Name == ""` falls back to the replica-derived `names []string`
lookup (the layer SHRINKS to a fallback; it does not disappear).

**Rationale**: injected social events land in the log and are re-applied
through the same arms on every replay (`docs/wiki/sim-state-reducer.md`:
"the live loop and crash recovery run the exact same code"). A name check
inside an arm would reject every pre-086 injected row at replay time —
replay must accept unnamed shapes *forever*, so the arm is the one place
the check must never live. The door/marshal choke points, by contrast, are
live-emission-only: `mustPayload` runs when the executor constructs an
event (`internal/sim/state.go:460`), and `InjectSocial`
(`internal/sim/loop.go:382`) runs when the mind/guardian inject — neither
is on the replay path (replay streams stored bytes via `ReplayEvents`).

**Evidence**: "Unknown event types … are recorded history but state
no-ops, so new event types never break old replay"
(`docs/wiki/sim-state-reducer.md`) — the same doctrine, applied to shapes:
old shapes must never break new replay. Spec 085's US1 AS-5 is the
house formulation this spec inherits: "a pre-085 world log … replayed from
genesis under the new code … byte-identical to what the old code produced."

**Alternatives**: arm-side validation gated on "is this live or replay" —
rejected: `Apply` deliberately has no such mode bit and adding one would
fork the single mutation path the replay proof rests on.

## R4 — Names are the fixed roster constant; `Ref(i)` is a pure function of the index; dead agents and sentinels are covered by construction

**Decision**: `sim.Ref(i int) AgentRef` returns
`{ID: i, Name: AgentNames[i]}` for `0 <= i < agentCount` and
`{ID: i, Name: ""}` otherwise (sentinels: −1 = any/none/personal);
`sim.Refs([]int) []AgentRef` maps it. Every emitter — executor, mind,
guardian, bundle, persona — uses these constructors; none needs state
access, because `AgentNames` is the package-level canonical roster
(`internal/sim/agents.go:18-20`:
`var AgentNames = [agentCount]string{"Ash", "Birch", …, "Sage"}`).

**Rationale**: the board card's premise — "names are fixed per agent, so
the denormalization is replay-safe" — is even stronger in the code than
the card assumes: names are a compile-time constant, not even state. So
the "dead agent's name at emission" edge is a non-edge: a dead agent keeps
its `Name` on state (`Agent.Dead bool` beside `Name`,
`internal/sim/agents.go:185-192`; the morgue and the villager strip render
dead agents by name today), and `Ref(i)` doesn't even consult liveness.
Emission-time naming and any-later-time naming provably agree.

**Evidence**: genesis stamps `Name: AgentNames[i]`
(`internal/sim/state.go:261`); the guardian's own name resolver already
walks the same constant (`internal/sim/miracles.go:765-769`). Strangers
need nothing: no stranger payload references a villager index (census —
`StrangerArrived/Moved/Took/Departed` carry night/day/x/y/kind only,
`internal/sim/stranger.go:232-252`); the stranger's own identity is not an
agent index and is out of scope.

**Sentinel law** (validation, R5): a ref with `ID` outside `[0,
agentCount)` is legal with `Name == ""` — `GuardianOrder.Agent` −1 = any
(`internal/sim/guardian.go`), `ProposalPayload.Target` −1 = non-exile,
`MemoryAddedPayload.Subject` −1 = personal. A ref with in-roster `ID` must
carry `Name == AgentNames[ID]` exactly.

## R5 — Enforcement is three mechanical layers: the typed ref, append-time validation at both emission doors, and an exhaustive registered-payload sweep

**Decision** (AC #3 satisfied by "typed ref + append validation" AND the
"exhaustive payload test" — both, not either):

1. **The type** (R1): agent fields ARE `AgentRef`; an int can no longer
   hold an agent reference in a migrated payload.
2. **Append validation**: `mustPayload` (`internal/sim/state.go:460`, the
   executor-side marshal choke point) gains a reflection walk over the
   payload value — every `AgentRef` found (fields, slices, nested structs,
   pointers) with in-roster `ID` must have `Name == AgentNames[ID]`, else
   panic (the exact contract `mustPayload` already has for marshal errors:
   an unnamed ref is the same class of programming bug). Out-of-sim
   emitters marshal their own bytes (guardian `mustJSON`, mind bare
   `json.Marshal` — census §4), so the injection door gets the mirror
   check: `InjectSocial` (`internal/sim/loop.go:382`) decodes each
   whitelisted event's payload into its registered type (the catalog,
   below) and runs the same walk, refusing the batch on an unnamed ref —
   live-injection-only, never on replay (R3).
3. **The exhaustive sweep**: a new `internal/sim` payload catalog —
   `var PayloadCatalog = map[string]func() any` (event type → zero payload
   value) — becomes the missing sim-side registry (census: today the only
   catalogs live in `internal/tui`: `digestRegistry` ~119 entries,
   `digest.go:121`; `catalogFixture`, `digest_test.go:29`). Three tests
   ride it:
   - `TestPayloadAgentRefSweep` (sim): for every catalog type, reflect
     over the payload struct; any `int`/`[]int` field whose json tag is in
     the frozen agent-vocabulary (`agent, a, b, from, to, speaker,
     listener, subject, owner, taker, violator, witnesses, targets,
     attendees, proposer, target, yeas, nays, agents, participants,
     source, src`) must instead be `AgentRef`/`[]AgentRef`, unless the
     field appears on the frozen, justified exemption allowlist
     (data-model.md §5). New payloads cannot dodge: the sweep cross-checks
     catalog completeness against the backticked types in
     `docs/wiki/event-types.md` — the same doc-anchoring trick
     `TestCatalogSweep` already uses (`internal/tui/digest_test.go:322-330`).
   - `TestNoAgentRefInState` (sim): reflection walk over `sim.State`'s
     full type graph asserting zero `AgentRef` — the R2 invariant as a
     unit test.
   - `TestCatalogSweep` (tui, existing gate): every `catalogFixture`
     payload for an agent-bearing type is rewritten to the named shape,
     and the sweep gains the AC #2 assertion — agent-bearing fixtures must
     digest identically with `names = nil` (proof of no replica
     dependency, R6). The tui side also asserts `catalogFixture` keys ⊆
     `sim.PayloadCatalog` keys, welding the two catalogs together.

**Rationale**: the card demands "enforced mechanically … not by
convention". Each layer catches what the others miss: the type catches
migrated fields at compile time; append validation catches an emitter that
hand-builds `AgentRef{ID: 3}` without the constructor; the sweep catches a
FUTURE payload author who declares a new `Agent int` field and never
migrates it — the failure mode types alone cannot see.

**Evidence**: `TestCatalogSweep`'s three-way sweep
(fixture→registry, registry→fixture, doc-backticks→fixture;
`internal/tui/digest_test.go:296-331`) is the proven shape for
"exhaustive, doc-anchored, cannot drift silently". The tool-registry's
described≡declared discipline (`docs/wiki/tool-registry.md` family) is the
same philosophy: the registry is the single enumerable truth, tests sweep
it.

**Alternatives**: constructors alone (rejected: convention — nothing stops
a literal); door-only validation (rejected: executor emissions bypass the
door); vocabulary sweep alone without the catalog (rejected: reflection
cannot enumerate Go types — an unregistered payload is invisible; the
doc-anchored catalog closes exactly that hole).

## R6 — The chronicle reads names FROM payloads for new events; `resolveSubject` gains a generic single-ref fallback, measurably raising the locatable-event hit rate

**Decision**: two consumer-side changes, both fallback-preserving:

1. **Digest naming (AC #2)**: `digestFunc`s for agent-bearing types read
   `ref.Name` from the decoded payload first and fall back to
   `agentName(names, ref.ID)` only when `Name == ""` (historic rows). The
   `names []string` parameter stays (it IS the fallback); the AC #2 proof
   is the `names = nil` sweep assertion (R5). The regex rewriter
   `resolvePayloadNames` (`grammar.go:160`) survives untouched as the
   grammar-miss fallback for historic logs — the post-hoc layer shrinks to
   fallback duty, per the card's "the TUI lookup layer can shrink".
2. **Subject resolution (the reorient move-10 reframe)**:
   `resolveSubject` (`internal/tui/digest.go:2103`) keeps
   `subjectRegistry` (~80 hand-written per-type decoders,
   `digest.go:1535`) as the authoritative first pass, and gains a GENERIC
   second pass for registry misses: scan the raw payload JSON for
   `{"id":N,"name":…}` ref objects; if exactly ONE distinct in-roster
   agent id appears, that agent is the subject candidate (live position
   via the existing `liveAgentPos` path). Multi-ref payloads stay
   unlocatable — the registry's deliberate absences ("metatron.nudged's
   several targets … the honest hint wins over a guess",
   `digest.go:1524-1534`) remain honest, because ambiguity is now detected
   structurally instead of by hand-listing.

**Rationale**: this is the village-lens completion framing — jump-to-source
(`m.jumpToSource()`, `internal/tui/tui.go:1727`; `⏎` and click-line both
route through it) is only as good as `resolveSubject`'s hit rate. New event
families (a spec 084/085 directive or prophecy row, a `journal.*` row —
types with no registry entry today) become locatable the moment their
payloads carry exactly one ref, with zero per-type registry work.

**Measurability**: a test drives every `catalogFixture` row (rewritten
named payloads) through `resolveSubject` against a live-replica fixture
and asserts the locatable count strictly exceeds the registry-only count,
plus pins specific newly-locatable types (data-model.md §7).

**Alternatives**: rewriting `subjectRegistry` entries to read refs
(rejected as the *mechanism* for the hit-rate claim: it improves naming,
not coverage — the fallback is what adds locatable types; individual
registry entries still switch to payload names opportunistically where
they render names).

## R7 — The reverse-jump rider: strip glyph / roster row → `centerCameraOn`, via the established hit-region pattern; keyboard `J`; two mouse-parity oracle entries

**Decision**: clicking a villager-strip glyph or a villagers-tab roster
row centers the map camera on that villager (`centerCameraOn(a.X, a.Y)` —
the spec-049 camera writer, `internal/tui/tui.go:1712`, "a jump IS a
pan"); keyboard parity is a new `J` binding in villagers mode (roster +
detail views): jump to the selected villager. Mechanics:

- **Hit regions**: the renderer-records pattern (the "chronHit pointer
  pattern", `tui.go:359`; regions invalidated at frame top, `views.go:76`)
  — a `stripHit` region recorded by `villagerStripView` (glyph i's column
  is pure arithmetic: `lipgloss.Width(count) + 1 + 2*i`, `views.go:2651`;
  screen row = `headerRows` when `computeRows(…).VillagerStrip > 0`,
  widescreen only — the same arithmetic `mapHitOrigin` already encodes,
  `look.go:887-897`) and a `rosterHit` region recorded by
  `villagerRosterBody` (row height 4 wide / 1 narrow, `views.go:3195-3218`).
  `handleMouse` (`look.go:906`) gains the two branches in its fixed
  priority chain.
- **Click semantics**: roster-row click selects the row AND jumps (the
  chronicle click precedent: click-line selects + jumps in one,
  `look.go:935-937`); strip-glyph click jumps directly (strip order ==
  roster order == `replica.Agents` order, `views.go:2613` — no selection
  state needed or created). Narrow layout: the strip never renders narrow
  (no narrow strip case); a roster jump in narrow switches the active pane
  to the map (the `jumpToSource` FR-007 precedent, `tui.go:1727-1735`).
- **Dead villagers still jump**: a dead agent keeps its `X, Y` (the map
  draws the grave; the roster row shows the coordinates) — clicking a `†`
  glyph centers on the grave. No liveness gate; an empty replica is a
  no-op.
- **The standing resolution amendment**: `villagerStripView` carries
  "Display-only (standing resolution 1: no cursor, no keys, no mouse
  target)" (`views.go:2614`) and `villager-strip.md` declares "display-only
  end to end". The rider AMENDS this deliberately and visibly: the strip
  gains exactly one mouse affordance (click-to-jump) and still no cursor
  and no keys — its keyboard path is the villagers tab's `J` on the shared
  roster order. Both doc pages and the code comment are rewritten in the
  same PR.
- **Oracle**: two new `mouseParityOracleEntry` rows
  (`internal/tui/mouseparity_test.go:190-209` — `{page, control, claim,
  check}`), checks in the `checkChronicleJumpToSourceMouseClaim` shape
  (`:309-331`): render to record the region, dispatch
  `mouseLeftRelease(x,y)` through `m.Update`, assert `panX/panY` moved
  (and, for the roster, `villSelected` changed). The TASK-154 sweep then
  enforces doc-row ↔ oracle bijection mechanically.
- **Key choice `J`**: villagers-mode keys taken are
  `esc d j k g G enter` (`tui.go:1809-1871`); `⏎` already opens detail;
  `c` would shadow global recenter. `J` is free, mnemonic (Jump), sits on
  the selection home row beside `j`/`k`, and the keymap already documents
  the shadowing-precedent pattern for mode keys (`keymap.md:152`).

**Rationale**: the rider is the operator-placed reverse direction of the
same lens ("strip glyph / roster row → camera center"); reusing
`centerCameraOn` means recenter (`c`), panned-title, and follow-suspend
semantics fall out for free — no new camera state, the spec-049 doctrine.

**Alternatives**: a strip keyboard cursor (rejected: net-new selection
state duplicating `villSelected` for zero reach — the roster IS the
keyboard surface over the same ordering); `enter` in the roster as jump
(rejected: taken — opens detail); centering via `resolveSubject`
(rejected: the villager's replica position is directly at hand; subject
resolution is for events).

## R8 — Scoped exemptions: the frozen allowlist is part of the contract, not a loophole

**Decision**: four exemption classes, each recorded in the sweep-test
allowlist with its rationale string (data-model.md §5 is normative):

1. **`PlaceFact.Source`** (`src,omitempty` — the telling agent,
   `internal/sim/mentalmap.go:65-73`): `PlaceFact` is state-resident
   (mental maps) AND rides four payloads (`SawPayload`, `PlaceToldPayload`,
   `PlaceRevealedPayload`, `MapCorrectedPayload`). It keeps the bare int
   (R2 invariant); the payloads' TOP-LEVEL actor fields
   (`agent`/`from`/`to`) carry the refs. Provenance-by-index inside a
   state fact is not a chronicle-legibility surface.
2. **`CogToolCallPayload.Args`** (`internal/sim/cognition.go:123`): opaque
   recorded tool arguments (`json.RawMessage` of the tool's own schema) —
   recorded observability of what the model SAID, never rewritten.
3. **Non-agent int fields sharing vocabulary-adjacent tags**: none exist
   today with matching tags (`tones` is per-participant sentiment,
   `rumor_id`/`conv`/`belief_id` have non-vocabulary tags); the allowlist
   holds the rationale anyway so a future collision is a deliberate
   decision, not silence.
4. **`FaithChangedPayload.SourceID`** (`internal/sim/faith.go:78`): a
   string source-EVENT identifier that encodes an agent index only for
   `reason:"villager_died"` (`faith.go:226`). It stays a string (it
   identifies directives/prophecies too), and the payload gains an
   additive `Agent *AgentRef `json:"agent,omitempty"`` set exactly when
   the reason is `villager_died` — AC #1 satisfied without overloading
   SourceID. `omitempty` keeps every other reason's emission
   byte-identical to spec 085's shape.

**Evidence**: census (data-model.md §3) — the exemptions are the complete
set of agent-adjacent fields NOT migrating; everything else migrates.
