# Data model: Neglect detector (spec 083 / TASK-133)

All new state is reducer-derived, `omitempty`, and replay-safe by construction. No
`format_version` bump; no tuning.json entries; no injection-door whitelist change.

## 1. `NeglectState` — per-agent derived anchors (new)

One new field on `sim.Agent` (`internal/sim/agents.go`), the Journal/Hail/Map pointer
precedent — nil for every pre-083 snapshot, so canonical bytes round-trip identically:

```go
// Neglect (spec 083) is the death-by-neglect detector's derived substrate:
// per survival need, when it entered its critical band, when a class intent
// last landed, and whether this episode already fired. Written ONLY by the
// needs_changed / intent_set / sim.neglect_detected reducer arms; lazily
// allocated on first non-zero write (replay-identical — the arms are
// deterministic). omitempty pointer: pre-083 snapshots stay byte-identical.
Neglect *NeglectState `json:"neglect,omitempty"`

// NeglectState fields are flat per-need (no maps — fixed canonical JSON,
// the Needs-struct shape). All *Since / *Intent ticks are duration anchors,
// non-zero only ⇒ SHIFT under rebaseTicks (miracles.go taxonomy).
type NeglectState struct {
    // Band-entry anchors: tick the need crossed BELOW its danger band
    // (dangerFoodBelow/dangerWarmthBelow/dangerRestBelow); 0 = not in band.
    FoodSince   int64 `json:"food_since,omitempty"`
    WarmthSince int64 `json:"warmth_since,omitempty"`
    RestSince   int64 `json:"rest_since,omitempty"`
    // Class-intent stamps: tick an intent in the need's class last landed
    // (agent.intent_set, goal ∈ needClassGoals); 0 = never.
    FoodIntent   int64 `json:"food_intent,omitempty"`
    WarmthIntent int64 `json:"warmth_intent,omitempty"`
    RestIntent   int64 `json:"rest_intent,omitempty"`
    // One-per-episode latches: set by the sim.neglect_detected arm,
    // cleared with the matching *Since anchor on recovery to/above band.
    FoodFired   bool `json:"food_fired,omitempty"`
    WarmthFired bool `json:"warmth_fired,omitempty"`
    RestFired   bool `json:"rest_fired,omitempty"`
}
```

Accessor pair (`Since(need)/setSince`, `ClassIntent(need)/…`, `Fired(need)/…`) keyed by
the `recoveryNeeds` closed-set names, mirroring `needValue`'s switch — so the sweep,
the arms, and the probe predicate share one need-agnostic surface.

**Alternative rejected**: nine flat fields on `Agent` (the `LastMindIntentDone`
precedent) — correct but noisy; one pointer keeps the `Agent` struct legible and the
pre-083 byte-identity proof trivial (field absent vs nine omitempties).

## 2. Reducer arms (writers — the ONLY writers)

- **`agent.needs_changed`** (`state.go` ~1718, existing arm, extended): after folding
  the absolute values, per need in {food, warmth, rest}: value < band && Since == 0 ⇒
  Since = e.Tick; value >= band ⇒ Since = 0, Fired = false (episode over, latch
  re-armed). Uses the same band constants as the sweep (one home).
- **`agent.intent_set`** (`state.go` ~845/872, existing arm, extended): after
  `appendIntent`, if `needClassOf(p.Goal)` is non-empty ⇒ stamp that need's `*Intent` =
  e.Tick. Source-agnostic (reflex, planner, plan, meeting all count — any scheduled
  class intent proves engagement).
- **`sim.neglect_detected`** (new arm): set the payload need's Fired latch. Nothing
  else — the memory rides its own `agent.memory_added` companion (existing arm), and
  Level/Since on the payload are for consumers, not state.

## 3. `sim.neglect_detected` — event payload (new)

Beside `NeedsPayload`/`DiedPayload` (`internal/sim/agents.go` ~1262):

```go
// NeglectDetectedPayload (spec 083): a survival need has sat below its
// danger band for neglectWindowTicks with zero intents in its class over
// the same window — the shape that killed Oak (world-01 day 7).
NeglectDetectedPayload struct {
    Agent int    `json:"agent"`           // agent index (chronicle name resolution)
    Need  string `json:"need"`            // "food" | "warmth" | "rest"
    Level int    `json:"level"`           // pre-tick need value at firing
    Since int64  `json:"since"`           // tick the need entered the band
}
```

- Emitted by the `stepEvents` heartbeat sweep only (pure over pre-tick state + tick);
  executor emission class ⇒ **no** `injectSocialWhitelist` / operator-door entry.
- `agent` is already in the chronicle's `agentIndexFields` — name resolution free.
- Catalog obligations (TestCatalogSweep triple): `digestRegistry` entry +
  `catalogFixture` row + backticked mention in `docs/wiki/event-types.md`; contract
  row in `specs/018-chronicle-digest/contracts/digest-grammar.md` §3; alert-tier
  membership in `isAlertType`.

## 4. Companion memory (existing event, new emission site)

`situatedMemoryEvent(nextTick, i, salNeglect, PlaceAt(s, a.X, a.Y), "", OriginWitness,
neglectMemoryText(need))` — appended to the batch immediately after the event.

- **`salNeglect = 9`** (`internal/sim/memory.go` table) — equals
  `GenerationBumpSalience`: the reducer's existing bump arm supersedes in-flight
  thoughts (the planner beat, research §3), rate-bounded by the episode latch.
  Deliberate join of the near-death/exile interrupt band (research.md R6).
- **Fixed per-need voice-of-evidence texts** (deterministic, one home beside the
  salience const):
  - warmth: `"I am dangerously cold and I have done nothing to warm myself for hours."`
  - food: `"I am starving and I have done nothing to find food for hours."`
  - rest: `"I am exhausted and I have done nothing about resting for hours."`
- Origin `OriginWitness` (direct perception of one's own condition; explicitly not
  `OriginAction` — inaction is the subject), `Why` empty (no driving intent — that is
  the point), no Subject (personal, not gossip-seeding).

## 5. Doctrine constants (promoted-dial-READY, NOT tuning.json)

Spec-083 const block in `internal/sim/agents.go` (the spec-062/064 doctrine-block
pattern):

```go
// neglectWindowTicks (spec 083 R3): the neglect window T — a need below its
// danger band this long, with zero class intents over the same span, fires
// the detector. 7200 ticks = 2 game-hours: dozens of reflex/planner beats
// (grace 120, planner cadence 1800), and ≈5 game-hours of health runway
// before Oak's trajectory kills. Dial-ready, not dialed.
neglectWindowTicks = 7200
```

plus `salNeglect = 9` in the memory.go salience table. Critical bands are the existing
`dangerFoodBelow`/`dangerWarmthBelow`/`dangerRestBelow` — reused, never re-declared.

## 6. Class dictionary (new, beside the goal registry)

`internal/sim/policy.go`, adjacent to the resolver map (research §2's rot rule):

```go
// needClassGoals (spec 083): which intent GOALS serve which survival need.
// Lives beside the goal-resolver registry so the dictionary and the registry
// rot together or not at all (TestNeedClassGoalsResolve pins membership).
var needClassGoals = map[string]string{
    "forage": "food", "hunt": "food", "cook": "food",
    "goto_warmth": "warmth", "warm_up": "warmth",
    "build_fire": "warmth", "refuel_fire": "warmth",
    "sleep": "rest",
}
func needClassOf(goal string) string { return needClassGoals[goal] }
```

Goal-name-only granularity; kind-parameterized transfers deliberately unclassed
(research.md R4, accepted v1 false-fire mode).

## 7. Rebase taxonomy (`internal/sim/miracles.go`)

`rebaseTicks` gains the six `NeglectState` tick anchors: **SHIFT, non-zero only** (the
`NeedsAnchorTick` / `IntentRecord.Tick` / `Belief.Reinforced` row of the taxonomy).
Latches (bools) and payloads (history) are untouched. The taxonomy doc comment
(~miracles.go:218-227) is amended to list them.

## 8. State-transition summary (one episode)

```
healthy ──need < band (needs arm)──▶ in-band [Since=t₀]
in-band ──class intent (intent arm)──▶ in-band [Intent=t, zero-clock reset]
in-band ──need ≥ band (needs arm)──▶ healthy [Since=0, Fired=false]
in-band ∧ tick−Since ≥ T ∧ (Intent=0 ∨ tick−Intent ≥ T) ∧ ¬Fired ∧ awake
        ──heartbeat sweep──▶ EMIT event+memory ──reducer──▶ Fired=true
Fired ──only exit: need ≥ band──▶ healthy (re-armed)
```
