# Data Model: Exercise catalog wave (spec 077)

Normative shapes. All new state fields are `omitempty`; zero values are honest
(pending/absent), so every pre-077 snapshot round-trips byte-identically — **no
`format_version` bump** (the spec-072 `CharterCustom` precedent).

## 1. Incident kinds and schedule parameters

`IncidentScheduleEntry` (content struct, `internal/sim/scenario.go`) gains two additive
kind-specific fields; existing entries are unchanged.

```go
type IncidentScheduleEntry struct {
    Kind   string
    Day    int64
    Time   string // "HH:MM" game time
    X, Y   int    // kind-specific position (gru/stranger entry tile; blight center)
    Radius int    // NEW — forage_blight patch radius (Manhattan); 0 elsewhere
    Hours  int    // NEW — cold_snap duration in game hours; 0 elsewhere
}
```

| Kind | Params used | Compiled window (`windowEnd`) | Emission precondition (named predicate, TASK-28-reusable) | State latch |
|---|---|---|---|---|
| `gru_emerges` (shipped) | X, Y | next dawn after tick | no gru abroad; tile passable + unprotected | `s.Gru != nil` |
| `cold_snap` | Hours | `tick + Hours*3600` (the snap's own end) | no snap active (`coldSnapActive(s, tick)` false) | `ColdSnapUntil > tick` |
| `forage_blight` | X, Y, Radius | next dawn after tick | ≥1 unharvested forage tile in patch (`blightableTiles(m, s, x, y, r)` non-empty) | stricken tiles already carry `Harvest` overlays → predicate empty → not due |
| `stranger_arrives` | X, Y | next dawn after tick | no stranger abroad; tile passable + unprotected (`strangerEntryValid`) | `s.Stranger != nil` |

Compile validation (per `compileIncident` arm): `cold_snap` requires `Hours` in [1, 24];
`forage_blight` requires `Radius` in [1, 8]; day/time rules as shipped. A compile error
is a content bug (`TestScenarioSchedulesCompile` pins the whole catalog).

## 2. New event types (all reducer-total; TestCatalogSweep rows mandatory)

| Type | Payload | Emitted by | Reducer effect | Digest family / tier |
|---|---|---|---|---|
| `sim.cold_snap` | `{night, until_tick}` | `scenarioIncidentEvents` (TASK-28 ambient later) | `s.ColdSnapUntil = until_tick` | sim / normal |
| `sim.forage_blighted` | `{x, y, radius, tiles: [{x,y}], regrow_tick}` (tiles in deterministic row-major patch order) | `scenarioIncidentEvents` (TASK-28 later) | append `Harvest{X, Y, Regrow: regrow_tick}` per tile (existing overlay; skip tiles already harvested — idempotent re-apply) | sim / normal |
| `stranger.arrived` | `{night, x, y}` | `scenarioIncidentEvents` (TASK-28 later) | `s.Stranger = &Stranger{...}` | gru-family (threat voice) / normal |
| `stranger.moved` | `{x, y}` | `strangerStep` (executor tick) | position update | gru-family / normal |
| `stranger.took` | `{x, y, kind, n}` | `strangerStep` | decrement pile/chest stock (agent-withdrawal state shapes); append `StrangerTake{Tick, X, Y, Kind, N}` ring (retain 32) | gru-family / **whole-line alert** (joins `agent.died`, `gru.attacked`, `social.chest_taken`, `norm.violated`) |
| `stranger.departed` | `{day}` | `strangerStep` at dawn | `s.Stranger = nil` | gru-family / normal |
| `metatron.skills_observed` | `{fingerprint, names: []string}` | guardian turn-time observation (the `charter_observed` pipeline; emit on fingerprint change; stage-3+ only by construction) | `s.SkillsFingerprint = fingerprint; s.SkillsObservedSeq = e.Seq; s.SkillsObservedTick = e.Tick` | guardian family / normal |

Payload rules: NO authored/scenario marker on any type (spec FR-013). `stranger.*` is a
new namespace → `familyByNamespace["stranger"]`; `metatron.*` stays the frozen namespace
(skin display-aliasing applies to the solo Type column as shipped). Every type gets a
`digestRegistry` + `catalogFixture` row, and a backticked mention in
`docs/wiki/event-types.md` (the sweep's scanned parent) with the full row in a child
note.

## 3. New / amended state (reducer is the only writer)

```go
// state.go — all omitempty, no format bump
ColdSnapUntil      int64          `json:"cold_snap_until,omitempty"`      // SHIFT (rebase)
Stranger           *Stranger      `json:"stranger,omitempty"`             // entity, gru precedent
StrangerTakes      []StrangerTake `json:"stranger_takes,omitempty"`       // ring 32; Tick KEEP
CharterObservedSeq  int64         `json:"charter_observed_seq,omitempty"` // KEEP (log coordinate)
CharterObservedTick int64         `json:"charter_observed_tick,omitempty"`// KEEP
SkillsFingerprint   string        `json:"skills_fingerprint,omitempty"`
SkillsObservedSeq   int64         `json:"skills_observed_seq,omitempty"`  // KEEP
SkillsObservedTick  int64         `json:"skills_observed_tick,omitempty"` // KEEP

type Stranger struct {
    X, Y      int   `json:"x"`,`json:"y"`
    Night     int64 `json:"night"`
    LastMove  int64 `json:"last_move,omitempty"`  // SHIFT
    LastTake  int64 `json:"last_take,omitempty"`  // SHIFT
}
type StrangerTake struct {
    Tick int64  `json:"tick"` // KEEP — historical fact
    X, Y int    `json:"x"`,`json:"y"`
    Kind string `json:"kind"`
    N    int    `json:"n"`
}
```

Rebase taxonomy (`rebaseTicks`): SHIFT = remaining-window/cooldown semantics preserved
across a time snap; KEEP = historical coordinates. Final table argued in plan D5, tested
beside the existing taxonomy suite.

Amended reducer arms: `metatron.charter_observed` additionally stamps
`CharterObservedSeq/Tick` from the envelope (`PlacedSeq` precedent).

## 4. ExerciseDefinition changes

```go
// curriculum.go — additive content fields
BoundaryDay int // N>0: evaluate at dawn of day N only (first-night: 2).
                // 0: rolling — evaluate at every dawn from day 2 until a pass lands.
```

Evidence assembly at emission (generalized `scenarioRubricEvents`): for each satisfied
rubric term whose type has a sanctioned constructor, attach —
`metatron.order_placed` → `OrderPlacedEvidence` (shipped);
`metatron.charter_observed` → `CharterEvidenceFromState` (NEW, reads
`CharterObservedSeq/Tick` + `Custom: s.CharterCustom`; omitted when Seq==0 — pre-077
honesty); `metatron.skills_observed` → `SkillsObservedEvidence` (NEW, `Custom: true` by
construction). No other constructor, no freehand refs.

## 5. The exercise inventory (normative content table)

Seeds 46101–46109; every rubric term names a cataloged event type
(`TestExerciseRubricTermsAreCatalogedEventTypes`); Met derivations are pure state facts.
Labels below are the hand-authored plain language the report cards render.

### Stage-1 — The Voice (conversational prompting) — 3 exercises

| ID | Seed | Boundary | Schedule | Rubric terms (Label / Event / Met over state) |
|---|---|---|---|---|
| `first-night` (shipped) | 46101 | day 2 | gru_emerges d1 22:00 (44,0) | unchanged (survive-to-dawn / zero deaths / watch set) |
| `cold-dawn` | 46103 | day 2 | cold_snap d1 22:00, 8h | "village survives to dawn of day 2" / `sim.day_started` / tick≥dawn2 ∧ !Ended ∧ no deaths · "no villager freezes" / `agent.died` / zero exposure-cause deaths (count = exposure deaths) · "a watch set before nightfall" / `metatron.order_placed` / firstNightWatch predicate (shared) |
| `stranger-at-the-gate` | 46104 | day 2 | stranger_arrives d1 23:00 (border tile pinned per seed) | "village survives to dawn of day 2" / `sim.day_started` · "no villager dies" / `agent.died` / len(Deaths)==0 · "nothing is taken" / `stranger.took` / len(StrangerTakes)==0 (zero-wanted — renders honestly per spec 072) |

### Stage-2 — The Written Word (durable instruction) — 2 exercises

| ID | Seed | Boundary | Schedule | Rubric terms |
|---|---|---|---|---|
| `the-law` (shipped def; emission NEW) | 46102 | rolling | — | unchanged terms (`theLawRubric`); NEW: evidence via `CharterEvidenceFromState`, pass emission at first satisfying dawn |
| `blighted-larder` | 46105 | day 4 | forage_blight d2 08:00, center on the seed's forage belt, radius 4 | "a player-authored charter in force" / `metatron.charter_observed` / CharterCustom ∧ Fingerprint≠"" · "no villager starves" / `agent.died` / zero starvation-cause deaths · "a larder banked against the blight" / `agent.deposited` / total stored food (chests+piles) ≥ threshold (content constant, pinned in code beside the definition) |

### Stage-3 — The Craft (capability design) — 2 exercises (fog visibility by stage default)

| ID | Seed | Boundary | Schedule | Rubric terms |
|---|---|---|---|---|
| `toolsmith` | 46106 | rolling | — | "your skill file guides the guardian" / `metatron.skills_observed` / SkillsFingerprint≠"" · "the guardian acts under it" / `metatron.order_placed` / ∃ player-origin order with PlacedTick ≥ SkillsObservedTick · "no villager dies" / `agent.died` / len(Deaths)==0 |
| `fog-watch` | 46107 | day 3 | cold_snap d1 22:00 6h; gru_emerges d2 22:00 (tile pinned per seed) | "village survives to dawn of day 3" / `sim.day_started` · "no villager dies" / `agent.died` · "a skill file in force before the trials" / `metatron.skills_observed` / SkillsObservedTick > 0 |

Stage-3 passes carry `SkillsObservedEvidence` (`Custom: true`) → the stage-3→4 gate's
first production satisfaction.

### Stage-4 — The Stewardship (mastery) — 2 exercises

| ID | Seed | Boundary | Schedule | Rubric terms |
|---|---|---|---|---|
| `long-winter` | 46108 | day 4 | cold_snap d1 22:00 8h; forage_blight d2 08:00 r4; stranger_arrives d2 23:00; gru_emerges d3 22:00 | "village survives to dawn of day 4" / `sim.day_started` · "no villager dies" / `agent.died` · "nothing is taken" / `stranger.took` |
| `stewards-charge` | 46109 | rolling | — | "a village law adopted" / `meeting.proposal_resolved` / len(Norms)>0 · "a player-authored charter in force" / `metatron.charter_observed` · "your skill file guides the guardian" / `metatron.skills_observed` · "no villager dies" / `agent.died` |

Stage-4 passes record but unlock nothing (`nextLadderStage` — graduation, unchanged).

`world.ValidScenarioExercise` mirrors all nine ids
(`TestScenarioVocabularyMirrorsSimCatalog`).

## 6. Lesson catalog tranche 2 (client-side content, `internal/tui/lessons.go`)

| ID | Tier | Trigger | Done | Pointer |
|---|---|---|---|---|
| `first-explain-answer` | prompting | `cog.tool_call` ∧ Tool=="explain" ∧ Verdict=="read_ok" | nil | → `?` guardian section |
| `first-report-card` | prompting | `guardian.report_card` | nil | → press 3 for the {{skin.guardian.tab_label}} tab |
| `first-skill-file` | prompting | `metatron.skills_observed` | nil | → your skill file now shapes the {{skin.guardian.epithet}} |
| `same-refusal-pattern` | prompting | FOLD: 3rd `cog.tool_call` with `rejected_*` verdict and identical non-empty Reason (session-local capped map) | nil | → press 4, then d for the decision trace — the charter may need the rule, not the retry |

Machinery delta: `lessonEntry` gains an optional fold-trigger seam
(`FoldTrigger func(*lessonFold, store.Event) bool`); `lessonTriggers` carries one bounded
`lessonFold{rejections map[string]int}` (cap 32 reasons). Per-event `Trigger` predicates
stay pure; exactly one entry uses the fold. Catalog count 8 → 12;
`populateHelpLessons` 1:1 unchanged; every string skin-tokened.

**Deferred (recorded, not stubbed)**: `first-faith-event` — rider on TASK-118.

## 7. Invariants

1. Recorded events are the only incident persistence; replay arms nothing.
2. No payload distinguishes authored from ambient emission.
3. Reducer arms are total and idempotent-on-replay (blight re-apply skips
   already-harvested tiles; stranger arms validate-at-door like `applyGru`).
4. A status/outcome never exceeds state facts (`ExerciseOutcome` unchanged).
5. Every catalog surface (CLI list, exercise tab, world vocabulary, digest, wiki
   event rows) derives from `sim.ScenarioExercises` / the event catalog — no second
   hand-maintained list anywhere new.
