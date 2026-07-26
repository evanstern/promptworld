# Data model: World fork + duel v1 (spec 076)

Shapes only; behavior in spec.md, decisions in research.md.

## 1. The lineage event — `world.forked` (authoritative, AC2)

New payload struct in `internal/sim/state.go`'s shared payload block (struct, not map —
canonical JSON, deterministic bytes; the `WorldCreatedPayload` convention):

```go
// WorldForkedPayload (spec 076) records the fork's provenance as the first
// event past the carried prefix: which world it was forked from and where.
// The reducer no-ops it (recorded history, like world.created) — fork state
// at the fork tick stays byte-identical to the parent's.
WorldForkedPayload struct {
    ParentName      string `json:"parent_name"`
    ParentSeed      uint64 `json:"parent_seed"`       // == the fork's own seed (identity check)
    ParentCreatedAt string `json:"parent_created_at"` // RFC3339; disambiguates renamed/recreated parents
    ForkTick        int64  `json:"fork_tick"`         // boundary snapshot tick
    ForkSeq         int64  `json:"fork_seq"`          // boundary snapshot seq (events 1..ForkSeq carried)
}
```

Placement in the fork's log: `tick = ForkTick`, `seq = ForkSeq + 1` (assigned by
`AppendEvents`; contiguity 1..N+1 holds). Exactly one per fork; a fork of a fork appends
its own — lineage chains are read newest-first (each hop names its immediate parent).

Reducer: explicit no-op arm beside `case "world.created"` — never mutates state.
Deliberately NOT on `sim.State`: no reducer logic reads it, and keeping it off state
preserves the fork/parent byte-identity proof (research R3).

Digest (FR-009, `internal/tui/digest.go` registry + `digest_test.go` fixture, the
`TestCatalogSweep` totality gate):

```
world.forked → `forked from <parent_name> at day <D>, <HH:MM>`
```

## 2. The manifest mirror — `Manifest.Lineage` (fast offline read)

Additive `omitempty` block in `internal/world/world.go`'s `Manifest`, after `Scenario`
(the `meeting`/`scenario` optional-block precedent; no `format_version` bump — a
lineage-less `world.json` round-trips byte-identically):

```go
// Lineage records fork provenance (spec 076): the parent world this one was
// forked from and the boundary tick. The world.forked event in this world's
// own log is authoritative; this block is the offline mirror (compare's
// default window) — written once by Fork, never mutated.
Lineage *LineageConfig `json:"lineage,omitempty"`

type LineageConfig struct {
    Parent          string `json:"parent"`                      // parent manifest name
    ParentCreatedAt string `json:"parent_created_at,omitempty"` // RFC3339
    ForkTick        int64  `json:"fork_tick"`
}
```

`Open` validation: absent block → nothing (byte-identical path). Present block →
structural check only: `Parent` non-empty and `ForkTick >= 0`, else the standard
`corrupt world.json: …` error. Not a closed vocabulary — no fail-closed enum posture
needed (contrast `terrain_gen`/`memory_relevance`).

## 3. Fork identity surface — what changes vs what carries

| Surface | Fork value | Rationale |
|---|---|---|
| Manifest `name` | NEW | the identity |
| Directory / socket / pidfile | NEW (path-derived: `SockPathIn`/`PidPathIn`) | side-by-side daemons for free |
| Registry entry | none written by fork; daemon boot self-registers if outside home | registry is advisory, self-healing |
| Manifest `created_at` | fork wall time | metadata only; parent's rides in `lineage` |
| Manifest `lineage` | NEW block | §2 |
| `seed` | **CARRIED** | determinism: prefix events + `rngAt` key off it (research R2) |
| `format_version`, map dims, `terrain_gen` | carried | same world physics |
| `meeting`, `teaching`, `memory_relevance` | carried | same posture |
| `stage`, `stage_overridden`, `charter_preset`, `scenario` | carried | the duel compares the SAME exercise under different prompts |
| Event log | prefix `seq <= ForkSeq` + `world.forked` | research R1 |
| Snapshots | boundary snapshot only | recovery boundary; older/newer parent snapshots are parent history |
| Meta `seed`/`format_version` | stamped | matches `validateMeta` at first boot |
| Meta `llm_spend_*` (all keys, totals + per-provider) | **COPIED** | AC5 inherit-the-wallet (research R4) |
| Sidecar files | per research R9 table | |

## 4. `world.Fork` ceremony result (CLI summary contract)

```go
// ForkResult summarizes a completed fork for the CLI (the MigrateResult
// pattern): everything cmdFork prints, nothing it must recompute.
type ForkResult struct {
    Name          string // fork manifest name
    Dir           string // destination directory
    ParentName    string
    ForkTick      int64  // boundary tick (CLI renders day/HH:MM via clock)
    ForkSeq       int64  // events carried: 1..ForkSeq
    TruncatedTail int64  // parent events past the boundary NOT carried (0 = nothing lost)
    BoundaryEnded bool   // boundary state carries an ended run (warn — spec edge case)
    SpendCarried  bool   // llm_spend_* keys found and copied (AC5 line in the summary)
}
```

## 5. The exported resolver surface (spec 072 contract, replica-parametric)

In `internal/tui` (type aliases keep every existing call site and test untouched):

```go
type ReportCardFact = reportCardFact // {Term, Met, Backing string/bool/string}
type ReportCardMode = reportCardMode // ReportCardConcluded | ReportCardLive (exported consts)

// ResolveRubricFacts is spec 072's ONE precedence switch, replica-parametric:
// recorded pass → all-met concluded (evidence-backed); state.Ended → concluded
// EvaluateRubric; else live EvaluateRubric. nil state → (nil, live): no card.
func ResolveRubricFacts(s *sim.State, def sim.ExerciseDefinition, pass *sim.CurriculumPass) ([]ReportCardFact, ReportCardMode)

// RecordedPassFor finds the exercise's retained CurriculumPass on state, or nil.
func RecordedPassFor(s *sim.State, exercise string) *sim.CurriculumPass

// RenderReportCard is the exported face of reportCardView — the bordered
// ✓/✗/… card every surface (and the duel) renders through.
func RenderReportCard(title string, facts []ReportCardFact, mode ReportCardMode, width int) string
```

`Model.resolveReportCardFacts` / `Model.recordedPassFor` become one-line wrappers over
these (`m.replica` + `m.runEnded()` fold into the state-driven switch: the Model's
`runEnded()` reads the same replica facts `s.Ended` exposes). Behavior pinned by the
existing spec-072 TUI tests plus one new cross-surface identity test (duel rows ==
postmortem rows for the same state).

## 6. Compare output model (the duel report)

Assembled by `cmd/promptworld/compare.go`; also the input contract for phase 2's HTML
retelling (which is why it is a model, not just print statements):

```go
type duelReport struct {
    A, B      duelSide
    Lineage   *duelLineage // non-nil when one side's lineage names the other
    Since     int64        // comparison window start (lineage fork tick | --since | 0)
    Divergence *divergence // nil = identical since the window (the honest line)
    Entries   []duelEntry  // interleaved chronicle entries, tick order
}

type duelSide struct {
    Name     string
    State    *sim.State              // offline-reconstructed (worlds.OfflineState)
    Exercise *sim.ExerciseDefinition // nil = ambient (no scorecard — honest note)
    Pass     *sim.CurriculumPass     // recorded instrument, if retained
    Facts    []tui.ReportCardFact    // via tui.ResolveRubricFacts
    Mode     tui.ReportCardMode
    Outcome  string                  // plain language, never the raw enum (FR-019)
}

type divergence struct {
    Tick int64  // first differing story event's tick
    Seq  int64  // per-side seqs may differ; tick is the shared coordinate
    A, B string // one-line description per side ("<type>" or "— (no event)")
}

type duelEntry struct {
    World    string // side label (world name)
    Day      int64
    FromTick int64
    Text     string // chronicle.entry text
}
```

Divergence scan input: each side's post-window events filtered to story types (exclude
`daemon.*`, `clock.*`, `cog.*`, `llm.*`; compare `(tick, type, payload)`, never
`wall_time`/`seq`) — research R7. Interleave input: `chronicle.entry` events with
`from_tick >= Since`, merged by `FromTick` (ties: A before B, stable).

## 7. Rendering registers (FR-019/020)

- Scorecard: `tui.RenderReportCard(exerciseID, facts, mode, width)` per side.
- Outcome vocabulary map (the glossary discipline): `passed` → "passed", `failed` →
  "did not make it through" (postmortem register), `in_progress` → "still running".
  Sourced from `sim.ExerciseOutcome`; the raw tokens never print.
- Header: `"<B> forked from <A> at day <D>, <HH:MM> (tick <T>)"`; running-side note:
  `"<name> is running — read as of its last committed batch"`.
- Zero divergence: `"the two runs are identical since the fork — the change made no
  observable difference yet"`.
- All authored prose: facts about the village, no-blame register (the morgue rule).
