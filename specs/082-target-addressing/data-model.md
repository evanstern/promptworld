# Data Model: the target-addressing grammar

This document is the NORMATIVE definition of the address grammar, its
validation and resolution semantics, and its error taxonomy. Two consumers
bind to it: the bundle effect compiler (this spec) and TASK-157's
designation tools (future; the seam is contract-named). The grammar lives
in one leaf package — `internal/target` — so both consumers parse
identically by construction.

## 1. Address forms

A target is a single string. Grammar (EBNF-ish; `X`/`Y` are non-negative
decimal integers; the whole string is trimmed; spaces are permitted after
`,` and around `..`/`->`, nowhere else):

```
address    = bare-name | typed-name | tile-address
bare-name  = <any string NOT matching reserved-prefix>       ; villager by name (v1 compat)
typed-name = "villager:" name                                ; villager by name, explicit
tile-address = class "@" locus
class      = "villager" | "structure" | "pile" | "terrain"   ; exact, lowercase
locus      = point | rect | line
point      = X "," Y
rect       = X "," Y ".." X "," Y                            ; inclusive rectangle
line       = X "," Y "->" X "," Y                            ; inclusive, axis-aligned (v1)

reserved-prefix = ^(villager|structure|pile|terrain)[@:]
```

**Reserved-prefix rule** (disambiguation, total and deterministic): a
string matching `reserved-prefix` MUST parse as `typed-name`/`tile-address`
— if the remainder is malformed, that is a `syntax` error, never a
fallback to bare-name. Any other string is a `bare-name` villager
reference. Consequence: bare-name behavior is byte-identical to the v1
compiler for every string v1 accepted (no roster name contains `@` or
`:`; the contract states the rule so future names cannot collide).

**Examples**

| String | Parse |
|---|---|
| `Rega` | name form, class villager, name `Rega` |
| `villager:Rega` | name form (explicit) |
| `villager@5,5` | point form, class villager |
| `structure@12,7` | point form, class structure |
| `pile@{args.x},{args.y}` | (after substitution) point form, class pile |
| `terrain@9,2` | point form, class terrain |
| `structure@3,9..1,5` | rect form, normalized corners (1,5)-(3,9) |
| `structure@1,5->1,9` | line form, endpoints (1,5)→(1,9) |
| `structure@` | `syntax` error (reserved prefix, malformed locus) |
| `boulder@3,4` | bare-name (no reserved prefix) → `unresolved` villager unless a villager is literally named that |
| `structure@1,5->2,9` | `form` error (diagonal line, v1) |

Note: template substitution (`{args.x}`/`{invoker}`) happens BEFORE
parsing, in the existing `subst` layer — the parser only ever sees
resolved strings.

## 2. The Address value (parser output)

```
Address {
  Form  : name | point | rect | line
  Class : villager | structure | pile | terrain
  Name  : string        (name form only; trimmed, matched case-insensitively at resolution)
  X, Y  : int           (point: the tile; rect: min corner; line: first endpoint)
  X2, Y2: int           (rect: max corner; line: second endpoint; unset otherwise)
}
```

**Normalization (parse-time, part of the grammar):**

- rect: corners re-ordered so `(X,Y)` = componentwise min and `(X2,Y2)` =
  componentwise max. `X,Y..X,Y` (zero-area) is valid — one tile.
- line: endpoint ORDER IS PRESERVED (direction is author intent, e.g. wall
  build order); v1 requires `X==X2 || Y==Y2` (axis-aligned), else a `form`
  error. A single-point line (`X,Y->X,Y`) is valid.
- name: surrounding whitespace trimmed; case preserved in `Name`
  (resolution is case-insensitive, matching `villagerIndex` today).

**Deterministic tile enumeration** (`Address.Tiles()` — rect/line/point;
the designation consumer's primitive):

- point → `[(X,Y)]`.
- rect → row-major: `y` from `Y` to `Y2` ascending, inner `x` from `X` to
  `X2` ascending.
- line → from `(X,Y)` to `(X2,Y2)` inclusive, stepping ±1 along the single
  varying axis, in endpoint order.

Enumeration is a pure function of the Address — no state, no map — so any
two consumers enumerate identically.

## 3. Resolution semantics (per consumer, against `sim.State` only)

Parsing is state-free (the leaf package). Resolution binds an Address to
world entities and is the CONSUMER's job, reading only the state it was
handed (for bundles: `CompileInput.State`) — no clock, no I/O, no ambient
reads. All resolvers are deterministic:

| Address | Resolution | Determinism guarantee |
|---|---|---|
| villager name | trimmed, case-insensitive first match over living roster by agent index (existing `villagerIndex`; the `AgentIndexByName` shape) | first-by-index |
| `villager@X,Y` | `sim.State.VillagerAt(X,Y)` — first living by agent index | the miracle door's own documented choice; a tile-addressed bundle move and miracle move can never disagree |
| `structure@X,Y` | the structure on the tile (at most one — buildSite invariant; exported presence probe over today's `structureIndexAt`) | one-per-tile |
| `pile@X,Y` | the pile on the tile (one-pile-per-tile reducer invariant; probe over today's `pileAt`) | one-per-tile |
| `terrain@X,Y` | the tile itself; bounds-checked at compile, removability (base kind, not-already-overlaid) stays REDUCER-side (`removeTerrain`) via the dry run | reducer-authoritative |
| rect / line | `Tiles()` enumeration (above); per-tile binding is the consumer's business (TASK-157) | pure function |

Compile-time resolution exists for ERROR QUALITY (name the effect index and
address at the door); the `InjectSocial` dry run over the UNCHANGED reducer
arms remains the semantic authority for presence, placement rules,
passability, charge, and villager-removal doctrine. A compiled payload is
self-contained data — replay never re-resolves.

## 4. Per-consumer form matrix

**Bundle effects (this spec):**

| Effect kind | villager name / `villager:` / `villager@X,Y` | `structure@X,Y` | `pile@X,Y` | `terrain@X,Y` | rect / line (any class) |
|---|---|---|---|---|---|
| `move_entity` | ✅ (payload class `villager`, source = resolved tile) | ✅ | ✅ | ❌ `form` (reducer cannot move terrain) | ❌ `form` (reserved: designations) |
| `remove_entity` | ❌ `form` (doctrine: a villager can never be removed — reducer unchanged and still authoritative) | ✅ | ✅ | ✅ | ❌ `form` (reserved: designations) |
| `grant_item` | ✅ (resolves to agent index) | ❌ `form` | ❌ `form` | ❌ `form` | ❌ `form` |
| `snap_time` / `narrate` | no target field / recipients unchanged (out of scope) | — | — | — | — |

Compiled payloads reuse the `sim.*` structs verbatim, `Gratis:false`:
class+tile forms fill `EntityMovedPayload`/`EntityRemovedPayload`
`{Class, X, Y[, ToX, ToY]}` byte-identically to
`guardian.BuildMiracleBatch`'s `move`/`remove` for the same inputs.

**Designation consumers (TASK-157 — the guaranteed seam, not built here):**
point (structure site), rect (settlement zone), line (wall line) — parsed
by the SAME package from `internal/tool` (stdlib-only import surface makes
this legal), enumerated via `Tiles()`. Class vocabulary extension (e.g. a
designation-kind token) is a one-table change in the parser; nothing in
the grammar hard-codes "four classes forever".

## 5. Error taxonomy (compile-time target failures)

Each class is distinguishable in the message; in bundles all surface as T5
`ruleErr`s carrying the effect index, the field name, and the offending
address text, and reject the WHOLE invocation (atomic; nothing lands, no
charge; descriptive `ResultForModel` back to the model — the existing
`rejected_gate` path).

| Class | Meaning | Example message shape |
|---|---|---|
| `syntax` | reserved prefix present but the address is malformed (bad locus, negative/non-integer coordinate, trailing junk) | `effect 2 field "target": "structure@" is not a valid address (want class@X,Y, class@X1,Y1..X2,Y2, or class@X1,Y1->X2,Y2)` |
| `class` | `<word>[@:]` where word is not a known class (only reachable if the reserved-prefix set and class set ever diverge; kept distinct for message quality) | `effect 0 field "target": unknown class "boulders"` |
| `form` | a well-formed address whose (effect kind × class × form) cell is ❌ in the matrix — includes rect/line anywhere in bundles ("reserved for designation consumers, TASK-157") and villager-removal doctrine | `effect 1 (remove_entity): target "Rega" names a villager — a villager can never be removed` |
| `bounds` | point/rect/line coordinates outside the world map (checked against state's map dims) | `effect 0 field "target": (99,2) is outside the 24×24 world` |
| `unresolved` | in-bounds, allowed form, but nothing binds: no living villager by that name / no living villager, structure, or pile on that tile | `effect 0: no structure at (9,9)` |

Reducer-side rejections (already-overlaid terrain, bad destination,
charge shortfall, and the villager-removal arm) are NOT in this taxonomy —
they are the dry run's, unchanged.

## 6. Invariants

1. **One parser** — no consumer re-implements or partially copies the
   grammar; `internal/target` is the single authority (drift-proof by
   import, the `MiracleCostsByEvent` derivation discipline).
2. **Parse is state-free; resolve is state-pure** — same string ⇒ same
   Address, always; same (Address, state) ⇒ same binding, always.
3. **Backward compatibility is total** — every v1-legal target string
   compiles to byte-identical events (bare names carry no reserved prefix).
4. **The reducer is never weakened** — no arm changes; the compiler only
   gains the vocabulary to produce payloads the arms already accept, and
   mirrors (never replaces) the villager-removal doctrine for error quality.
5. **No new event types** — `effectEventType` is unchanged; declared-events
   subset gate, whitelist, and digest catalog untouched.
