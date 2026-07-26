# Feature Specification: Target-addressing grammar for bundle effects (and the designation seam)

**Feature Branch**: `082-target-addressing` (task branch: `task-97-target-addressing`)

**Created**: 2026-07-26

**Status**: Draft

**Input**: TASK-97 (follow-up from TASK-85 / specs/036-scriptable-agent-tools;
realigned 2026-07-26 onto TASK-157's critical path). The board card carries NO
acceptance criteria — this spec's FRs/SCs are the de-facto ACs. Realignment
directive: the grammar designed here must serve BOTH bundle effects AND
TASK-157's designation tile/region addressing (settlement zone, structure
site, wall line) — designed once for both consumers.

## Grounding (verified against the task worktree, 2026-07-26)

**What exists.** The v1 bundle effect compiler
(`internal/bundle/effects.go`) addresses `move_entity`/`grant_item` targets
by **living-villager name only** (`villager()`/`villagerIndex()` — trimmed,
case-insensitive, first-match by roster index). `remove_entity` is
structurally present but **inert**: it compiles only the villager path and
the reducer rejects villager removal by doctrine
(`internal/sim/miracles.go` `applyEntityRemoved`: "a villager can never be
removed"). Meanwhile the reducer ALREADY applies structure/pile moves and
structure/pile/terrain removals — the built-in `work_miracle` reaches them
through `guardian.MiracleParams{Class, X, Y}` (class+tile,
`internal/guardian/miracle_batch.go`), and the payload structs
(`sim.EntityMovedPayload`/`sim.EntityRemovedPayload`) carry `Class`+source
tile today. **The entire gap is compiler-side grammar**: no contract
specifies how a bundle effect names a structure, pile, or terrain tile
(`specs/036-scriptable-agent-tools/contracts/bundle-manifest.md` shows only
villager-name targets), and no fixture exercises it.

**Deterministic resolution precedents.** `sim.State.VillagerAt(x,y)`
(first living by agent index — the documented deterministic choice for
shared tiles), `structureIndexAt` (at most one structure per tile — a
buildSite invariant), `pileAt` (one pile per tile — a reducer invariant).
The compiler resolves against `CompileInput.State` only; reducer
preconditions stay reducer-authoritative via `InjectSocial`'s dry run
(atomic, no charge on rejection).

**Import topology.** `internal/tool` is a leaf (cannot import `sim`,
`clock`, or `bundle` — mirrors are hand-carried with drift tests);
`internal/bundle` imports `tool` (manifest synthesizes `tool.Tool`) and
`sim`; `internal/sim` imports `tool`. So a grammar shared with TASK-157's
`internal/tool` designation params (the third consumer) cannot live in
`bundle` or `sim` — it needs a new leaf package.

**The designation consumer (TASK-157, dependency of record).** Guardian
designations address a settlement **zone** (region), a structure **site**
(point), and a **wall line** (line). This spec must guarantee the grammar
serves those shapes — parse, validate, and enumerate them deterministically
— WITHOUT building any designation tool, entity, or event (all TASK-157).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A bundle tool moves and removes structures and piles by class+tile (Priority: P1)

A bundle author writes a tool whose effects say
`{"kind": "move_entity", "target": "structure@12,7", "to_x": 4, "to_y": 4}`
or `{"kind": "remove_entity", "target": "pile@{args.x},{args.y}"}` — the
compiler resolves the class+tile address against live state and lands the
exact `metatron.entity_moved`/`metatron.entity_removed` payload the miracle
door would land. `remove_entity` stops being inert.

**Why this priority**: this is TASK-97's core deliverable — the v1
limitation named on the board and in `docs/wiki/bundle-tools.md`. Without
it the effect vocabulary claims five kinds but only reaches villagers.

**Independent Test**: a fixture world with a structure and a pile; invoke a
declarative tool moving the structure and removing the pile; assert the
landed events replay and the compiled payload bytes equal
`BuildMiracleBatch`'s for the same class+tiles.

**Acceptance Scenarios**:

1. **Given** a world with a structure at (12,7) and a valid build site at
   (4,4), **When** a bundle tool compiles
   `move_entity target=structure@12,7 to=(4,4)`, **Then** the batch holds
   one `metatron.entity_moved` with `class:"structure", x:12, y:7, to_x:4,
   to_y:4, gratis:false`, byte-identical to the miracle door's payload, and
   it lands through `InjectSocial`.
2. **Given** a world with a pile at (3,4), **When** a bundle tool compiles
   `remove_entity target=pile@3,4`, **Then** one `metatron.entity_removed`
   with `class:"pile", x:3, y:4` lands and the pile is gone after apply.
3. **Given** no structure at (9,9), **When** a tool compiles
   `move_entity target=structure@9,9`, **Then** the whole invocation is
   rejected with a T5 error naming the effect index, the field, and the
   unresolved address; nothing lands, no charge is spent.
4. **Given** an existing manifest using a bare villager name
   (`"target": "{args.target}"` → `"Rega"`), **When** it compiles, **Then**
   behavior is byte-identical to today (backward compatible — bare names
   still mean living villagers).
5. **Given** `remove_entity target=Rega` (or `villager:Rega`, or
   `villager@5,5`), **When** it compiles, **Then** the compiler rejects it
   with a form error stating villagers can never be removed (the reducer
   doctrine, mirrored for a better authoring error; the reducer arm stays
   unchanged and authoritative).

---

### User Story 2 - Terrain removal through the miracle overlay vocabulary (Priority: P2)

A bundle author removes a tree/forage/rock tile
(`remove_entity target=terrain@9,2`) and gets exactly the miracle door's
semantics: the reducer overlays via the executor's own vocabulary
(chop→Cleared, forage→Harvested+regrow, quarry→Quarried), rejecting
already-overlaid or non-removable tiles.

**Why this priority**: completes the class set `work_miracle` already
reaches; strictly additive over US1's dispatch (terrain is remove-only —
the reducer cannot move terrain).

**Independent Test**: fixture world with a tree tile; invoke
`remove_entity target=terrain@x,y`; assert `metatron.entity_removed`
`class:"terrain"` lands and the tile reads cleared; assert a grass tile is
rejected by the dry run (reducer-authoritative), whole-invocation.

**Acceptance Scenarios**:

1. **Given** a tree at (9,2), **When** `remove_entity target=terrain@9,2`
   compiles and lands, **Then** the payload is byte-identical to the
   miracle door's and the tile is cleared after apply.
2. **Given** `move_entity target=terrain@9,2`, **When** it compiles,
   **Then** the compiler rejects with a form error (terrain is not
   movable), before any dry run.
3. **Given** a tile out of the map bounds, **When** any tile-addressed
   effect compiles, **Then** a bounds error names the effect index and the
   offending coordinates.

---

### User Story 3 - The grammar serves TASK-157's designations (Priority: P3)

A designation consumer (TASK-157: settlement zone, structure site, wall
line) parses point, region, and line addresses through the SAME parser
package bundle effects use — from `internal/tool`, without importing `sim`
— and enumerates their tiles deterministically. Bundle effects reject the
region/line forms with an error that names them as reserved for
designations, so the seam is visible, tested, and contract-named.

**Why this priority**: the 2026-07-26 realignment's whole point — design
once for both consumers. Value ships as a proven, documented seam; the
designation tools themselves are OUT of scope.

**Independent Test**: unit tests in the parser package: `rect`/`line` forms
parse, normalize, and enumerate tiles in the specified deterministic order;
the package imports nothing beyond stdlib (leaf-safe for `internal/tool`);
a bundle effect using a rect address fails compile with the reserved-form
error. The contract carries a named "Designation addressing (TASK-157
seam)" section.

**Acceptance Scenarios**:

1. **Given** `structure@3,9..1,5` (corners unordered), **When** parsed,
   **Then** the address normalizes to the (1,5)-(3,9) inclusive rect and
   enumerates tiles row-major (y ascending, then x ascending) — a stable,
   documented order.
2. **Given** `structure@1,5->1,9` (axis-aligned line), **When** parsed,
   **Then** tiles enumerate from the first endpoint to the second,
   inclusive, in order; **Given** a diagonal line, **Then** parsing fails
   with a form error (axis-aligned only in v1).
3. **Given** a bundle effect whose target resolves to a rect or line form,
   **When** it compiles, **Then** the invocation is rejected with an error
   stating the form is reserved for designation consumers.
4. **Given** the parser package, **When** its imports are inspected (build
   or test assertion), **Then** it depends on stdlib only — importable by
   `internal/tool` without creating a cycle or a `sim` dependency.

---

### Edge Cases

- **Reserved prefix vs villager name**: a target string matching
  `^(villager|structure|pile|terrain)[@:]` MUST parse as a structured
  address — if malformed it is a syntax error, never silently treated as a
  villager name. Any other string is a bare villager name (v1 compat).
  Roster names (`sim.AgentNames`) contain no `@`/`:`; the rule is stated in
  the contract so future roster additions cannot collide.
- **Two villagers on one tile** (`villager@x,y`): deterministic first-by-
  agent-index via `sim.State.VillagerAt` — the exact miracle-door choice,
  so a tile-addressed bundle move and a tile-addressed miracle move can
  never name different villagers.
- **Zero-area rect** (`x,y..x,y`): valid; enumerates the single tile.
- **Negative or non-integer coordinates**: syntax error at parse (grammar
  admits non-negative decimal integers only); floats never reach payloads
  (existing `resolveInt`/`reqInt` discipline unchanged).
- **Template substitution producing a malformed address**
  (`structure@{args.x},{args.y}` with `args.x="north"`): compile-time
  syntax error naming effect index, field, and the substituted string
  (substitution runs first, parsing second — SC-005 error discipline).
- **Whitespace**: tolerated around the whole address and after commas
  (trimmed); class tokens and separators are exact and lowercase.
- **Already-overlaid terrain, occupied build site, impassable destination,
  charge shortfall**: NOT compiler concerns — the `InjectSocial` dry run
  rejects whole-batch via the unchanged reducer arms, atomically, spending
  nothing (spec 036 invariant preserved).
- **Script mode**: a `tool.star` returning
  `{"kind": "remove_entity", "target": "structure@%d,%d" % (x, y)}` flows
  through the same shared compile path — no separate script grammar.
- **narrate recipients**: unchanged — recipient vocabulary stays
  names/`all_living`/`target` (out of scope; recorded as a non-goal).
- **No new event types**: the effect vocabulary and `effectEventType` map
  are unchanged; `events` declarations, the whitelist gate, and the TUI
  digest catalog (`TestCatalogSweep`) are untouched by construction.

## Requirements *(mandatory)*

*(The board card has no ACs; these FRs and the SCs below are the de-facto
acceptance criteria for TASK-97.)*

### Functional Requirements

- **FR-001**: A new leaf package (`internal/target`) MUST implement the
  address grammar defined in [data-model.md](data-model.md): bare villager
  name (compat), `villager:<name>`, `<class>@X,Y` (point),
  `<class>@X1,Y1..X2,Y2` (rect), `<class>@X1,Y1->X2,Y2` (axis-aligned
  line), with the reserved-prefix rule, normalization, and deterministic
  tile enumeration exactly as specified there. The package MUST import
  stdlib only (leaf-safe for `internal/tool` — the TASK-157 seam).
- **FR-002**: `move_entity` MUST accept villager-designating forms (bare
  name, `villager:<name>`, `villager@X,Y`) plus `structure@X,Y` and
  `pile@X,Y`, compiling each to a `metatron.entity_moved` payload
  byte-identical to `BuildMiracleBatch`'s `move` for the same class and
  tiles (`Gratis:false`). Terrain, rect, and line forms MUST be rejected
  with a form error.
- **FR-003**: `remove_entity` MUST accept `structure@X,Y`, `pile@X,Y`, and
  `terrain@X,Y`, compiling to `metatron.entity_removed` byte-identical to
  the miracle door's payload — making the effect real. Every
  villager-designating form MUST be rejected at compile time with a form
  error naming the doctrine; the reducer's villager-removal rejection stays
  UNCHANGED and authoritative.
- **FR-004**: `grant_item` MUST accept every villager-designating form
  (bare name, `villager:<name>`, `villager@X,Y`), resolving tile form via
  `sim.State.VillagerAt`. Non-villager classes MUST be rejected with a
  form error.
- **FR-005**: Resolution MUST be deterministic and read ONLY
  `CompileInput.State` (roster walk, `VillagerAt`, one-per-tile
  structure/pile probes, map bounds) — no clock, no I/O, no ambient state.
  Same (args, state, seed, tick) ⇒ byte-identical batch; replay never
  re-resolves (landed events are self-contained data), preserving the
  spec-036 replay byte-identity guarantee.
- **FR-006**: Unresolvable or ill-formed targets MUST fail compilation with
  the data-model.md error taxonomy — `syntax`, `class`, `form`, `bounds`,
  `unresolved` — as T5 `ruleErr`s naming the effect index, the field, and
  the offending address text. Rejection is whole-invocation: nothing lands,
  no charge is spent, and the model receives the descriptive
  `ResultForModel` (existing `rejected_gate` path).
- **FR-007**: Script mode (`tool.star`) MUST accept the same grammar in
  `target` strings through the shared compile path — no script-specific
  address surface, no new `world` API.
- **FR-008**: `specs/036-scriptable-agent-tools/contracts/bundle-manifest.md`
  MUST be amended with a "Target addressing" section (grammar, per-kind
  form matrix, error behavior, compat rule, and guidance that
  grammar-bearing args use param kind `text` — `agent_name` params stay
  villager-name-validated) AND a named **"Designation addressing (TASK-157
  seam)"** section guaranteeing: the parser package and its leaf-safety,
  the rect/line forms with their normalization and enumeration order, and
  that bundle effects reserve (reject) those forms. `docs/bundles.md` (the
  authoring guide) MUST be updated to match.
- **FR-009**: No new event types, no `effectEventType` change, no reducer
  change, no change to boot validation semantics beyond the target field's
  new value space. Existing manifests and fixtures MUST compile
  byte-identically (backward compatibility is total for bare-name targets).
- **FR-010**: Fixtures MUST cover: a declarative tool and a scripted tool
  exercising each accepted form per effect kind; each error-taxonomy class
  (table-driven); the byte-identity pins of FR-002/FR-003 against
  `BuildMiracleBatch`; and replay determinism for the new paths (the
  existing `replay_test.go`/`script_replay_test.go` pattern, including the
  delete-bundle-dir-before-replay proof).

### Key Entities

- **Address** — the parsed form of a target string: `Form`
  (name|point|rect|line), `Class` (villager|structure|pile|terrain), and
  form-dependent payload (name text, or 1–2 coordinate pairs, normalized).
  Defined normatively in data-model.md.
- **Form matrix** — the per-consumer table of which (effect kind × class ×
  form) combinations are accepted; bundle effects consume name+point,
  designations (TASK-157) consume point+rect+line.
- **Error taxonomy** — the five compile-time failure classes for targets
  (data-model.md), each with its message obligations.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A bundle tool moves a structure, moves a pile, removes a
  structure (chest contents spilling per the reducer), removes a pile, and
  removes a terrain tile end-to-end in fixture worlds — `remove_entity` is
  demonstrably no longer inert (`go test ./internal/bundle/`).
- **SC-002**: For every class+tile form, the compiled payload bytes equal
  `BuildMiracleBatch`'s for the same inputs (byte-identity test — the
  dogfood-move precedent extended to structure/pile/terrain).
- **SC-003**: Replay byte-identity holds for worlds whose logs contain the
  new addressing (bundle dir deleted before replay); `go test ./...` green.
- **SC-004**: Rect and line forms parse, normalize, and enumerate with unit
  coverage in the leaf package; a bundle effect using them fails with the
  reserved-form error; the package's stdlib-only import surface is
  asserted. The contract carries the named TASK-157 seam section.
- **SC-005**: Every error-taxonomy class produces a message naming effect
  index, field, and offending address (table test); a rejected invocation
  lands nothing and spends no charge.
- **SC-006**: Contract + authoring guide amended; `docs/wiki/bundle-tools.md`
  re-verified and re-pinned ON THE BRANCH (its "TASK-97 limitation" prose
  resolved) with `docs/player/` regenerated; the merge-drift pr gate exits
  0 from the worktree; the PR merges with `gh pr merge --merge`.

## Assumptions

- **Line = axis-aligned in v1**: a wall line (TASK-157's stated consumer)
  is horizontal or vertical; diagonal lines are a parse-time form error.
  This avoids choosing a rasterization (Bresenham variant) no consumer
  needs yet; the syntax leaves room to relax later.
- **Region = inclusive rect in v1**: TASK-157's settlement zone is a
  rectangle; arbitrary polygons are out of scope.
- **Bundle effects stay single-tile**: no rect/line-target effect (no
  "remove every structure in a region") — reserved forms are parse-only
  for bundles. Widening the matrix is a future spec's one-table change.
- **No perception memories for bundle moves** (unchanged): the miracle door
  adds `memMoved` for villager moves; the bundle path never has (narrate is
  the author's channel). Tile-addressed villager moves keep that existing
  bundle behavior — consistency with the current compiler, not the door.
- **No new tool param kind**: address-bearing args are declared `text`;
  a first-class `target` param kind (schema-validated grammar) is TASK-157's
  call when the designation tools land in `internal/tool`.
- **`sim` presence probes**: the compiler needs exported one-per-tile
  presence checks for structures/piles (today unexported
  `structureIndexAt`/`pileAt`); adding narrow exported helpers to
  `internal/sim` is in scope (plan D4) and touches no reducer arm.
- Wiki notes pinning touched sources re-pin in-branch per the pr gate
  (expected: `bundle-tools.md`, possibly `sim-state-*`/`guardian-miracle-*`
  notes if `internal/sim` helper exports land in a pinned file); the gate
  is the authority — produce what it names, don't argue.
