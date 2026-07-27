# Feature Specification: Event payloads name their agents — AgentRef at emission

**Feature Branch**: `086-agent-named-payloads` (task branch: `task-17-agent-named-payloads`)

**Created**: 2026-07-27

**Status**: Draft

**Input**: TASK-17 (board card, 4 ACs; 2026-07-23 drift audit — payloads
carry indices only, no AgentRef exists, TUI post-hoc lookup intact;
reorient 2026-07-26 move 10 — agent-named payloads raise chronicle
jump-to-source's locatable-event hit rate via `resolveSubject`,
village-lens completion; operator-placed rider 2026-07-26 "rider is fine"
— REVERSE JUMP: strip glyph / roster row → camera center, shipping with a
mouse-parity oracle entry per the TASK-154 gate). Sweep claim: runbook
`docs/design/faith-directives-sweep-runbook.md` (signed-off 2026-07-26),
deliberate Lane C tail — this migration sweeps the sweep's own new payload
families (`directive.*`, `faith.*`/`prophecy.*`, `sim.neglect_detected`,
`stranger.*`) in one pass. Tier: Opus 4.8 (recorded on the card —
repo-wide payload migration + mechanical enforcement + back-compat
replay).

## Grounding (verified against the task worktree, 2026-07-27)

**What exists.** Every recorded event payload references agents by bare
index — `{"agent":2}`, `{"a":0,"b":3}`, `{"targets":[1,4]}` — across 69
payload struct types / ~85 event types / ~127 emission call sites (the
census, [data-model.md](data-model.md) §3). No `AgentRef` type exists
anywhere; no payload carries a name beside an index. Names are resolved
post-hoc at render time: `formatChronicleLine(e, names, sk)`
(`internal/tui/grammar.go:220`) with `names` built fresh from the replica
(`m.agentNames()`, `internal/tui/tui.go:2122`), plus the grammar-miss
regex rewriter `resolvePayloadNames` (`grammar.go:160`). Out-of-sim
consumers (webhook sinks TASK-18, exported logs, `jq`) have no replica and
cannot name anyone. Names themselves are a package constant —
`sim.AgentNames [8]string` (`internal/sim/agents.go:18-20`) — stamped onto
agents at genesis and never changed, dead or alive. Events are never
hashed; determinism hashes are over canonical STATE bytes only
(`State.Marshal()`/`Hash()`, `internal/sim/state.go:318-330`; snapshot
`stateHash` over stored bytes, `internal/store/store.go:151-165`), and the
event log is immutable in-schema (append-only triggers,
`docs/wiki/event-log.md`). Jump-to-source (spec 049) centers the camera on
a chronicle event's subject via `resolveSubject`
(`internal/tui/digest.go:2103`) over a hand-written ~80-type
`subjectRegistry`; the villager strip is display-only by standing
resolution (`views.go:2614`) and the villagers roster has keyboard
selection but no mouse target (`docs/design/tui/panels/villagers.md`).

**What this spec adds.** A typed `AgentRef` that marshals `{id, name}`,
carried by every agent-referencing payload field at emission (names fixed
per agent — replay-safe denormalization); mechanical enforcement (the
typed ref + append-time validation at both emission doors + an exhaustive
doc-anchored payload sweep); a permanent dual-shape back-compat contract
(reducer accepts named and unnamed forever; pre-086 replay byte-identical;
renderers fall back for historic rows); chronicle naming with no replica
lookup for new events plus a generic single-ref `resolveSubject` fallback
that measurably raises jump-to-source's hit rate; and the operator-placed
reverse-jump rider — strip glyph / roster row → camera center, with
keyboard parity and mouse-parity oracle entries.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The log names its people at emission (Priority: P1) — AC #1

An out-of-sim reader — a webhook sink, an exported log, a human with `jq`
— reads any NEW agent-referencing event and sees both the index and the
name: `{"agent":{"id":2,"name":"Cedar"}}`, `{"targets":[{"id":1,"name":
"Birch"},{"id":4,"name":"Fern"}]}`. The name is stamped at emission by
the emitter (executor, mind, guardian, bundle, persona — all of them),
from the fixed roster, so the log is self-describing with no replica.

**Why this priority**: this is the card's title requirement and the
format change everything else proves, enforces, or consumes.

**Independent Test**: run a fixture world across the event families
(intents, harvests, social, consolidation, governance, gru, guardian
plans/orders/prophecies, faith, neglect) and assert every agent-bearing
row in the log carries `{id,name}` objects with correct roster names —
using nothing but the log bytes.

**Acceptance Scenarios**:

1. **Given** a live post-086 world, **When** any of the ~85 cataloged
   agent-referencing event types is emitted, **Then** every agent field
   in its payload is an `{id,name}` object whose name equals
   `AgentNames[id]` (data-model §3 is the normative census).
2. **Given** a multi-agent payload (`talked` a/b, `conversation`
   participants, `norm.violated` witnesses, `metatron.nudged` targets,
   `meeting.*` attendees/yeas/nays, `chronicle.entry` agents), **Then**
   every element of every list is a named ref.
3. **Given** an injected event from the mind or guardian (memory_added,
   rumor_told, nudged, order_placed, directive.issued, prophecy.declared,
   morgue.epilogue), **Then** it carries names exactly like
   executor-emitted events — emission-time means ALL emitters, including
   bundle/scripted-tool effects (`internal/bundle/effects.go`) and the
   persona seeder (`internal/persona/files.go`).
4. **Given** a sentinel reference (`order_placed` agent −1 = any,
   `proposal` target −1, `memory_added` subject −1 personal), **Then** it
   marshals `{"id":-1,"name":""}` — sentinels are legal and never fake a
   name.
5. **Given** an agent who died earlier, **When** any later event
   references them (a rumor subject, a morgue epilogue, a belief
   revision's source), **Then** the ref carries their roster name — death
   never blanks a name (names are the fixed roster constant).
6. **Given** `faith.changed{reason:"villager_died"}`, **Then** the
   payload carries the additive `agent:{id,name}` ref beside `source_id`
   (data-model §3 row 73) — the one payload whose agent reference hid in
   a string.

---

### User Story 2 - The format is enforced by machinery, not convention (Priority: P1) — AC #3

A future payload author who adds `Agent int` to a new event type — or an
emitter that hand-builds a ref without a name — is caught by tests or by
the emission door itself, before the row ever lands. Enforcement is three
mechanical layers: the typed ref (agent fields ARE `AgentRef`), append
validation at both emission choke points (`mustPayload` for the executor
path, the `InjectSocial` door for injected batches), and an exhaustive
doc-anchored sweep over a new sim-side payload catalog.

**Why this priority**: "carried at emission" is only true while every
FUTURE emitter also complies; the card demands mechanical enforcement,
and without it the census decays back to convention.

**Independent Test**: the sweep tests themselves — plus mutation checks:
a test payload with an unnamed in-roster ref panics `mustPayload` and is
refused by the door; a synthetic catalog entry with a vocabulary-tagged
bare int fails `TestPayloadAgentRefSweep`.

**Acceptance Scenarios**:

1. **Given** `sim.PayloadCatalog` (new; event type → zero payload value),
   **Then** its key set covers every backticked event type in
   `docs/wiki/event-types.md` (the `TestCatalogSweep` doc-anchor,
   `internal/tui/digest_test.go:322-330`) and the TUI's `catalogFixture`
   keys are a subset — a new event type cannot exist outside the catalog
   without failing tests.
2. **Given** any catalog payload type with an `int`-kind field whose json
   tag is in the frozen agent vocabulary (data-model §5) and not on the
   frozen, rationale-carrying allowlist, **Then**
   `TestPayloadAgentRefSweep` fails.
3. **Given** a payload value containing an in-roster `AgentRef` whose
   name is missing or wrong, **When** the executor marshals it, **Then**
   `mustPayload` panics (its existing programming-error contract);
   **When** it arrives at `InjectSocial`, **Then** the batch is refused
   before the dry-run.
4. **Given** `sim.State`'s reachable type graph, **Then**
   `TestNoAgentRefInState` proves it contains no `AgentRef` — the hash-
   stability invariant (research R2) as a standing tripwire.
5. **Given** the four state-shared struct-as-payload types (`Directive`,
   `GuardianOrder`, `Prophecy`, `DeathRecord` in `run.ended`), **Then**
   their wire events carry refs via payload-mirror types while the state
   entities keep bare ints (data-model §4), and the reducer folds `.ID`s
   only.

---

### User Story 3 - Old worlds replay untouched; the contract is documented (Priority: P1) — AC #4

Every pre-086 world — including v1–v4 worlds that came through
`promptworld migrate` — keeps working: old rows decode through the same
reducer arms via the dual-shape unmarshal (a bare index decodes as
`{id, ""}`), no arm ever validates names (replay must accept unnamed
shapes forever), and a from-genesis replay of a pre-086 log produces
byte-identical state, because names never enter `sim.State` at all.
Renderers fall back to the replica lookup for historic rows — the
post-hoc layer shrinks to fallback duty but never disappears.

**Why this priority**: the event log is the source of truth for every
existing world; a migration that breaks replay breaks the product's
central promise.

**Independent Test**: a checked-in pre-086 fixture log replays from
genesis under the new build to a byte-identical `State.Marshal()`; a
mixed-era feed renders correctly (old rows via fallback names, new rows
via payload names).

**Acceptance Scenarios**:

1. **Given** a pre-086 fixture log spanning the payload families, **When**
   replayed from genesis under this build, **Then** the reconstructed
   state is byte-identical (`Marshal()`/`Hash()`) to what pre-086 code
   produced — no state shape, fold, or hash changes anywhere.
2. **Given** a pre-086 snapshot, **Then** it decodes and verifies exactly
   as today (snapshot `state_hash` is over STORED bytes; state shapes are
   unchanged) and recovery replays its tail through the same arms.
3. **Given** a `world.migrated` event (payload embeds the full canonical
   `sim.State`), **Then** its shape and handling are untouched — migration
   never rewrites payloads (snapshot-cut doctrine,
   `docs/wiki/world-migration.md`), and migrated v1–v4 worlds carry no
   agent-named rows until they emit new events.
4. **Given** a live pre-086 world continued under this build, **Then**
   new emissions carry names from the next event on while old rows keep
   their bytes (append-only triggers) — a mixed log is the documented,
   permanent normal.
5. **Given** any reducer arm and a legacy-shape row, **Then** the arm
   folds identically to pre-086 (reads `.ID` only) and NEVER rejects for
   a missing name — name validation lives exclusively at live-emission
   choke points (research R3).
6. **Given** the back-compat contract, **Then** it is documented as the
   normative matrix in data-model §6 and re-grounded into the wiki's
   event-types family in this PR (AC #4's "documented" clause).

---

### User Story 4 - The chronicle names people from the payload, and jump-to-source finds more subjects (Priority: P2) — AC #2

The chronicle renders new events' names straight from the payload — no
replica, no post-hoc lookup: an agent-bearing digest produces identical
output with `names = nil`. And the village lens completes: for event
types the hand-written `subjectRegistry` never covered, `resolveSubject`
now finds the subject generically whenever a payload carries exactly one
named ref — so `⏎`/click jump-to-source works on more of the feed,
measurably.

**Why this priority**: AC #2 plus the reorient move-10 reframe — but it
renders what US1 creates, so it follows the format.

**Independent Test**: the `names = nil` sweep assertion over every
agent-bearing catalog fixture; the hit-rate test asserting
registry+fallback locates strictly more fixture types than registry-only,
pinning newly-locatable types.

**Acceptance Scenarios**:

1. **Given** any agent-bearing event emitted post-086, **When** its
   digest row renders with `names = nil`, **Then** the output is
   identical to rendering with the replica names — payload names suffice
   (the AC #2 mechanical proof, asserted inside `TestCatalogSweep`).
2. **Given** a historic (unnamed) row in the same feed, **Then** it still
   renders correctly via the replica fallback (`agentName(names, id)`),
   and grammar-miss rows still pass through `resolvePayloadNames` — the
   fallback layer shrinks, never breaks.
3. **Given** an event type with no `subjectRegistry` entry whose payload
   carries exactly one in-roster ref (e.g. `journal.entry_written`,
   `morgue.epilogue`, `cog.thought`), **Then** `resolveSubject` resolves
   it (live position, payload name) and jump-to-source centers the
   camera on it.
4. **Given** a registry-miss payload with several distinct refs
   (`metatron.nudged` multi-target, `chronicle.entry` agent list),
   **Then** it stays unlocatable — ambiguity is detected structurally,
   preserving the honest-hint doctrine (`digest.go:1524-1534`).
5. **Given** the rewritten catalog fixtures, **Then** the hit-rate test
   proves locatable(registry+fallback) > locatable(registry-only).

---

### User Story 5 - Reverse jump: the lens's other direction (Priority: P3) — operator rider

The player clicks a villager's glyph on the strip — or a roster row on
the villagers tab, or presses `J` on the selected roster entry — and the
map camera centers on that villager, exactly as jump-to-source centers on
an event's subject. Forward: event → villager on the map. Reverse:
villager → their place in the world. Dead villagers jump to their grave.

**Why this priority**: the operator-placed rider ("rider is fine") — net
new UI, homed here because this task already completes the village lens's
locate machinery; it depends on nothing above and nothing above depends
on it.

**Independent Test**: the two mouse-parity oracle entries (strip glyph,
roster row) driving real `tea.MouseMsg` dispatch and asserting the camera
pan moved; a keyboard test for `J`.

**Acceptance Scenarios**:

1. **Given** the widescreen layout with the strip rendered, **When** the
   player clicks villager i's glyph, **Then** the camera centers on that
   villager (`centerCameraOn(a.X, a.Y)`) — pan-based, so `c` recenter,
   the panned map title, and follow-suspend all behave per spec 049's
   existing camera semantics.
2. **Given** the villagers tab roster, **When** the player clicks a row,
   **Then** that row becomes selected AND the camera centers on that
   villager (the chronicle click-line select+act precedent); in narrow
   layout the active pane switches to the map (the jump-to-source FR-007
   precedent).
3. **Given** the villagers tab (roster or detail view), **When** the
   player presses `J`, **Then** the camera centers on the selected
   villager — keyboard parity, keyboard primary per the input-parity
   doctrine (`patterns/keymap.md`).
4. **Given** a dead villager (strip `†` glyph or roster row), **Then**
   the jump centers on their grave coordinates (agents keep `X, Y` after
   death); **Given** no replica yet, **Then** clicks are no-ops.
5. **Given** the TASK-154 mouse-parity sweep, **Then** both new controls
   have corpus rows with mouse claims AND matching
   `mouseParityOracle` entries, and the sweep passes both directions
   (doc→oracle, oracle→doc).
6. **Given** the strip's standing resolution 1 ("no cursor, no keys, no
   mouse target", `views.go:2614`), **Then** it is amended deliberately in
   code comment and design pages: the strip gains exactly one mouse
   affordance and still no cursor/keys — its keyboard path is the roster's
   `J` (documented equivalence, strip order == roster order).

---

### Edge Cases

- **Dead agent's name at emission**: a non-edge by construction — names
  are the package constant `AgentNames` and `Ref(i)` never consults
  liveness; dead agents keep `Name` on state and in the morgue. Posthumous
  references (rumor subjects, epilogues, belief sources) name correctly
  (US1 AS-5).
- **Migrated v1–v4 worlds**: `world.migrated` embeds the full canonical
  `sim.State` — untouched because State carries no refs (research R2);
  migration is snapshot-cut and never rewrites event payloads; the
  migrated world's historic rows are unnamed and render via fallback
  (US3 AS-3).
- **Bundles / scripted tools / persona seeding emitting agent refs**:
  `internal/bundle/effects.go` (item_granted, memory_added) and
  `internal/persona/files.go` (secret_seeded) are emitters like any other
  — constructors + door validation cover them (US1 AS-3); the census
  counts their call sites.
- **Unicode names**: the current roster is ASCII, but `AgentRef` must not
  care — `encoding/json` string escaping is deterministic; the enforcement
  tests include a unicode-named ref fixture so a future roster (or skin-
  driven rename, if ever ratified) cannot break marshal determinism.
- **Sentinels**: −1 (any/none/personal) marshals `{"id":-1,"name":""}`;
  validation requires empty name OUTSIDE the roster and exact roster name
  inside it — a fake name on a sentinel is as much a bug as a missing name
  on an agent (US1 AS-4).
- **`ProphecyClaim.Agent` with index 0**: the survives-kind claim's agent
  can legitimately be index 0; the payload mirror carries a full ref
  (never `omitempty` on the ref object), and claim
  normalization/equality keeps operating on IDs (data-model §4).
- **Same-type mixed batches**: a tick batch may contain named rows only
  (all emitters migrate in this PR — there is no partial-emitter state);
  mixed NAMED/unnamed appears only across the pre/post-086 boundary in
  the log, which the back-compat matrix covers (US3 AS-4).
- **The strip under width overflow**: glyphs shed from the end with a
  trailing `…N` marker (`views.go:2653-2673`) — the hit region covers
  only rendered glyphs; clicks on the overflow marker are no-ops (no
  guessing which hidden villager was meant).
- **`cog.tool_call` args mentioning agents**: opaque recorded model
  output (`Args json.RawMessage`) — exempt, never rewritten (allowlist
  §5.2); the recorded-observability doctrine wins over uniformity.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001 (The type)**: `internal/sim` MUST gain
  `AgentRef{ID int `json:"id"`; Name string `json:"name"`}` with plain
  struct marshal (canonical fixed field order) and a custom dual-shape
  `UnmarshalJSON` accepting a bare JSON number (legacy → `{n, ""}`) or the
  object form — permanently. Constructors `Ref(i)` / `Refs(ids)` are pure
  functions of the index over the `AgentNames` roster constant; sentinel
  IDs get empty names. [AC #1, #4]
- **FR-002 (The census migration)**: every agent-referencing payload
  field in data-model §3 MUST carry `AgentRef`/`[]AgentRef` under its
  existing json tag — 66 types migrate in place; `FaithChangedPayload`
  gains the additive `agent` ref (row 73); all ~127 emission call sites
  construct refs via FR-001's constructors. The census table is
  NORMATIVE: a payload type absent from it and from the no-agent list is
  a spec bug, not implementer discretion. [AC #1]
- **FR-003 (The split law)**: the four state-shared struct-as-payload
  types (`Directive`/directive.issued, `GuardianOrder`/order_placed,
  `Prophecy`/prophecy.declared, `DeathRecord`/run.ended) MUST split into
  payload-mirror types carrying refs on the wire (same json tags) while
  state entities keep bare ints; reducer arms fold `.ID`s only.
  [AC #1, #4]
- **FR-004 (Names never in state)**: no `AgentRef` may be reachable from
  `sim.State`'s type graph; `TestNoAgentRefInState` enforces it by
  reflection. `State.Marshal()`/`Hash()` bytes for any pre-086 event
  history MUST be byte-identical to pre-086 code — state shape, folds,
  snapshots, and `world.migrated` untouched. [AC #4]
- **FR-005 (Append validation — live only)**: `mustPayload` MUST
  reflection-validate every `AgentRef` in the value (in-roster ⇒
  `Name == AgentNames[ID]`, out-of-roster ⇒ `Name == ""`), panicking on
  violation; `InjectSocial` MUST run the same walk over each event's
  payload (decoded via `PayloadCatalog`) and refuse the batch on
  violation, before the dry-run. Name validation MUST NOT exist in any
  `Apply` arm — replay accepts unnamed shapes forever. [AC #3, #4]
- **FR-006 (The catalog + sweep)**: `sim.PayloadCatalog` (event type →
  zero payload value) MUST cover every backticked type in
  `docs/wiki/event-types.md` (doc-anchored completeness);
  `TestPayloadAgentRefSweep` MUST reflect over every catalog type and
  fail on any int-kind field with an agent-vocabulary json tag
  (data-model §5's frozen vocabulary) absent from the frozen,
  rationale-carrying allowlist (§5's four entries — the complete set).
  The TUI side MUST assert `catalogFixture` ⊆ `PayloadCatalog`. [AC #3]
- **FR-007 (Chronicle payload-first naming)**: agent-bearing
  `digestFunc`s MUST read the payload ref's name, falling back to
  `agentName(names, id)` iff the name is empty; `TestCatalogSweep`'s
  fixtures are rewritten to named shapes and gain the `names = nil`
  identical-output assertion. `resolvePayloadNames` and `m.agentNames()`
  survive as the historic-row fallback layer. [AC #2]
- **FR-008 (Subject fallback + hit rate)**: `resolveSubject` MUST gain a
  registry-miss generic pass — exactly one distinct in-roster ref object
  in the payload ⇒ subject candidate; zero or several ⇒ unlocatable;
  `world.migrated` stays excluded. A test MUST prove the locatable
  count strictly increases over registry-only and pin newly-locatable
  types. [AC #2, reorient move 10]
- **FR-009 (Reverse jump — mouse)**: clicking a villager-strip glyph or a
  villagers-tab roster row MUST center the map camera on that villager
  via `centerCameraOn` (roster click also selects the row; narrow layout
  switches the active pane to the map). Implemented via renderer-recorded
  hit regions (the chronHit pointer pattern), frame-top invalidation, and
  `handleMouse` branches. Dead villagers jump to grave coordinates; no
  replica ⇒ no-op; the strip's overflow marker is not a target. [rider]
- **FR-010 (Reverse jump — keyboard parity)**: `J` in villagers mode
  (roster and detail views) MUST center the camera on the selected
  villager — keyboard primary per the input-parity doctrine; the strip's
  keyboard path is the roster (shared ordering), stated in the design
  pages. [rider]
- **FR-011 (Oracle + design pages)**: both controls MUST ship with
  `mouseParityOracle` entries (real `tea.MouseMsg` dispatch asserting
  camera pan) and corpus control-table rows (`panels/villager-strip.md`
  `— · click glyph`; `panels/villagers.md` `J · click row`), amending
  standing resolution 1 and the pages' display-only/parity notes;
  `patterns/keymap.md` gains the villagers-mode `J` row;
  `node scripts/check-tui-design.mjs --changed` passes with all touched
  pages re-verified in-branch. [rider, TASK-154 gate, spec 047 gate]
- **FR-012 (Determinism & scope guards)**: no new RNG; no state fields;
  no event-type additions or removals; no change to any `Apply` arm's
  fold semantics; `internal/cognition` untouched;
  `internal/mind`/`internal/guardian` touched ONLY at payload
  construction sites; stranger payloads untouched (no villager refs —
  verified census). Event emission ORDER is untouched everywhere. [AC #4]

### Key Entities

- **AgentRef** — the typed `{id, name}` reference; shape, marshal
  contract, and sentinel law in [data-model.md](data-model.md) §1–§2.
- **The payload census** — the normative 75-row table of every
  agent-referencing payload and its treatment;
  [data-model.md](data-model.md) §3.
- **Payload mirrors** — the four state-shared split types;
  [data-model.md](data-model.md) §4.
- **PayloadCatalog + enforcement sweep** — the sim-side registry, frozen
  vocabulary, and allowlist; [data-model.md](data-model.md) §5.
- **Back-compat matrix** — the documented dual-shape contract;
  [data-model.md](data-model.md) §6 (normative for AC #4's documentation
  clause).
- **Reverse-jump controls** — the two-row control contract;
  [data-model.md](data-model.md) §8.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: a fixture drive across the payload families produces a log
  where every agent-bearing row carries correct `{id,name}` refs,
  verified from log bytes alone (no replica). [AC #1]
- **SC-002**: `TestPayloadAgentRefSweep`, `TestNoAgentRefInState`, the
  catalog completeness checks, and the mustPayload/door mutation tests
  all pass — and injecting/marshaling an unnamed in-roster ref
  demonstrably fails. [AC #3]
- **SC-003**: a pre-086 fixture log replays from genesis to
  byte-identical state; pre-086 snapshots verify; mixed-era feeds render
  (old rows fallback, new rows payload-named); the back-compat matrix is
  in the spec dir and re-grounded into the wiki. [AC #4]
- **SC-004**: `TestCatalogSweep` passes with rewritten named fixtures
  plus the `names = nil` assertion; the hit-rate test proves
  registry+fallback > registry-only with pinned new locatables. [AC #2]
- **SC-005**: the mouse-parity sweep passes with the two new oracle
  entries; `J` keyboard test passes;
  `node scripts/check-tui-design.mjs --changed` exits 0 with
  `panels/villager-strip.md`, `panels/villagers.md`, and
  `patterns/keymap.md` amended and re-pinned in-branch. [rider]
- **SC-006**: `go test ./...` green; the merge-drift pr gate exits 0 from
  the worktree (wiki re-pins — the event-types family especially — and
  `docs/player/` regenerated in-branch); the PR merges with
  `gh pr merge --merge`; no rebase anywhere. [gates]

## Assumptions

- **Wire-shape change is ratified by the card**: `{"agent":2}` →
  `{"agent":{"id":2,"name":"Cedar"}}` changes what same-type events look
  like across the 086 boundary. No external consumer ships yet (TASK-18
  is paired future work and is the beneficiary); every in-repo consumer
  migrates in this PR. Recorded for operator review in the PR body.
- **Names are permanent per agent**: the roster constant is the source;
  if a future spec ever makes names dynamic, the emission-time stamp
  becomes load-bearing history (the ref records what the agent was called
  WHEN it happened) — that reading is deliberate and documented here.
- **`cog.*` canonical field order**: spec 007's contract described the
  pre-086 emission shape; this spec supersedes the field TYPES at
  emission while preserving field order and tags. Historic `cog.*` rows
  are covered by the dual-shape law like every family.
- **The `J` keybinding**: chosen against the taken-key map
  (`esc d j k g G enter` in villagers mode; `c` shadows global recenter);
  flagged for operator review in the PR body alongside the strip's
  standing-resolution amendment.
- **`journal.entry_written`/`entry_deleted` and `sim.tuning_applied`**
  are real event types outside today's digest catalog (census §1); they
  join `PayloadCatalog` (and the journal types migrate to refs) but this
  spec does NOT add digest rows for them — catalog-coverage changes
  beyond the census are out of scope.
- Wiki notes pinning touched sources re-pin in-branch per the pr gate
  (expected set in plan.md); the gate is the authority.
