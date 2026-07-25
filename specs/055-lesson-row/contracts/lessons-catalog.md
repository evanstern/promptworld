# Contract: the lesson catalog (spec 055)

The single source both surfaces read. `internal/tui/lessons.go` owns it; the help
overlay's `helpLessons` is populated from it 1:1 at init and never edited by hand.

## Minimum taxonomy (v1 ships exactly these 8; adding more is additive)

| id | tier | trigger (type + payload) | done-signal (v1) | pointer target |
|---|---|---|---|---|
| `first-suppression` | mechanics | `cog.outcome`, suppressed outcome | none (`x` only) | `?` → the screen (suppressed badge row) / speed keys |
| `first-gru-attack` | mechanics | `gru.attacked` | none | chronicle tab (`2`) |
| `first-charge-regen` | mechanics | `metatron.charge_regenerated` | none | guardian strip / guardian tab (`3`) |
| `first-order-expired` | mechanics | `metatron.order_expired` | `metatron.order_placed` (player re-places an order) | guardian tab (`3`), 👁 rows |
| `first-death` | mechanics | `agent.died` | none | chronicle tab (`2`) / villagers (`4`) |
| `first-rejected-tool-call` | prompting | `cog.tool_call`, verdict ≠ landed | none | villagers detail `d` (decision trace) |
| `first-custom-charter` | prompting | `metatron.charter_observed`, `default: false` | none | guardian tab (`3`) |
| `first-fuzzy-order` | prompting | `metatron.order_placed`, `fuzzy: true` | none | guardian tab (`3`), order row's fuzzy mark |

Exact lesson prose is implementation-time copy, bound by: one concept per lesson,
plain language, no engineering vocabulary, skin tokens for every guardian reference,
and the pointer names a real key/tab that exists on main.

## Invariants (each mechanically tested)

1. **One catalog, two surfaces**: `len(helpLessons) == len(lessonCatalog)`, id-for-id
   (SC-002).
2. **Never twice**: a surfaced id is recorded seen; a seen id never surfaces again —
   across worlds and restarts (SC-001).
3. **One active**: at most one lesson rendered at any time; concurrent triggers queue
   in arrival order; queued entries decay rather than surface stale (FR-004).
4. **Suffix**: rendered line 2 always ends `(? for more · x dismiss)` (FR-001).
5. **Tokens resolve**: no rendered string contains `{{` (SC-005); resolution per
   research.md R1 (spec 052 §2 order, default table as terminal fallback).
6. **Model-free / client-only**: no LLM call, no IPC command, no daemon change, no new
   event type (FR-002); byte-identical output for a given event sequence + skin + seen
   record.

## Keymap addition (the spec-047 gate's doc half)

`x` — dismiss the active lesson; strict no-op when no lesson is active (documented
no-op doctrine, `patterns/keymap.md` binding-selection rules). The same PR flips
`patterns/keymap.md`'s "New global keys" `x` row from specified/unbuilt into the
global-mode table, and flips `panels/lesson-row.md` `status: specified → shipped` with
real renderer symbols and a fresh `verified_against` pin.
