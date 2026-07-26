# Implementation Plan: World fork + duel v1 — `promptworld fork` and the rubric-first scoreboard

**Branch**: `076-world-fork-duel` (task branch: `task-67-world-fork-duel`) | **Date**: 2026-07-26 | **Spec**: [spec.md](spec.md)

**Design inputs**: [research.md](research.md) (decisions R1–R9), [data-model.md](data-model.md) (shapes §1–7)

## Summary

Three moves on the shipped substrate, no refactor. (1) **sim/world vocabulary**: the
`world.forked` lineage event (reducer no-op, digest-cataloged) and the additive
`Manifest.Lineage` block. (2) **the fork ceremony** (`internal/world/fork.go`, the
`Migrate` ceremony's sibling): fresh log = parent prefix + boundary snapshot +
`world.forked` + meta (seed/format/`llm_spend_*` — AC5's inherit-the-wallet), manifest
with fresh identity (seed carried), sidecar copy per the R9 catalog; CLI `fork`
subcommand. (3) **the duel scoreboard** (`promptworld compare`): offline state per world
(extracted `OfflineSnapshot` core), facts through the spec-072 resolver made
replica-parametric and exported, the shared `reportCardView` renderer, divergence over
story events, interleaved chronicles. Design pages, wiki re-pins, and player docs ride
the PR (pr gate).

## Technical Context

**Language**: Go. **Testing**: `go test ./...`; ceremony tests on the `migrate_test.go`
model; e2e on the `manager`/`determinism` e2e models. **Scope**: `internal/sim/state.go`
(payload + no-op arm), `internal/world/world.go` (+`fork.go`, new), `internal/store/store.go`
(one meta-enumeration helper), `internal/worlds/probe.go` (extract state helper),
`internal/tui/reportcard.go`+`views.go` (exports, wrappers), `internal/tui/digest.go`
(+fixture), `cmd/promptworld/main.go`+`commands.go` (+`fork.go`, `compare.go`), tests;
`docs/design/tui/` re-pins as flagged; `docs/wiki/` + `docs/player/` per the diff.
**Constraints**: events append-only in-schema (never delete/patch — R1); replay
determinism (lineage enters as a recorded no-op event; seed carried — R2); manifest and
snapshot byte-identity for existing worlds (`omitempty`); spec-072's one-resolver law
(FR-018 — no second precedence switch); `reportCardView`/`consoleCard` contracts
unchanged; forks never merge (FR-014).

## Constitution Check (v1.2.0)

- **I. Artifact-grounded** — PASS: decision chain is 2026-07-22 review → 2026-07-25 D7 →
  reorient 2026-07-26 decision 3 → TASK-67 → this spec; the AC5 budget decision is
  derived from meter.go evidence and recorded (research R4), not re-asked.
- **II. One task, one PR** — PASS: TASK-67 ↔ `task-67-world-fork-duel` ↔ one PR; phases
  are internal breakdown; the phase-2 HTML retelling is explicitly a FOLLOW-ON deliverable
  (its own future task), not a phase of this PR.
- **III. Gates** — PASS: claim gate already run for `076-world-fork-duel`; worktree gate
  with `--spec 076 --task TASK-67`; `check-tui-design.mjs --changed`;
  `check-merge-drift.mjs pr`; spec-bridge mirror gate.
- **IV. Grounding freshness** — PASS (planned): touched sources are pinned by
  `world-save-directory.md`, `world-save-manifest-fields.md`, `world-save-path-helpers.md`
  (world.go), `event-log.md`/`snapshots.md` (store.go), `sim-state-reducer.md` + the
  event-types catalog notes (state.go, new event), `instance-manager.md` (probe.go),
  `cli-world-lifecycle.md`/`cli-promptworld.md` (main.go/commands.go, new subcommands),
  `report-card-renderer.md` (reportcard.go/views.go exports),
  `llm-budget-degraded-mode.md` (re-verify: fork-inheritance note; meter.go itself
  untouched), plus computed re-pins. All re-pinned IN the branch; player docs regenerated
  if the wiki changes; merge with `gh pr merge --merge`.
- **V. Model tiers** — PASS: this spec/plan/tasks cycle is the planning tier;
  implementation dispatches to `spec-implementer` on **Opus 4.8** (tier recorded on
  TASK-67: cross-package architectural — world lifecycle, store, sim vocabulary, TUI
  exports, two new CLI verbs; determinism-doctrine-adjacent).

**Post-Phase-1 re-check**: PASS — no new packages beyond `fork.go`/`compare.go` files in
existing homes; Complexity Tracking empty.

## Design

### D1 — Lineage vocabulary (sim)

- `internal/sim/state.go`: add `WorldForkedPayload` to the shared payload block
  (data-model §1 shape, doc comment naming spec 076 and the no-op rule); add
  `case "world.forked":` beside `case "world.created":` with the recorded-history
  comment. No state field, no whitelist change, no `format_version` bump.
- `internal/tui/digest.go` + `digest_test.go`: registry entry + `catalogFixture` row for
  `world.forked` (data-model §1 digest line) — `TestCatalogSweep` stays total (FR-009).

### D2 — Manifest lineage block (world)

- `internal/world/world.go`: `LineageConfig` + `Manifest.Lineage *LineageConfig
  json:"lineage,omitempty"` after `Scenario` (data-model §2); `Open` structural
  validation (present block: `Parent` non-empty, `ForkTick >= 0` → else the standard
  corrupt-manifest error); `world_test.go`: round-trip byte-identity for lineage-less
  manifests, validation table for present blocks.

### D3 — The fork ceremony (world + store helper)

- `internal/store/store.go`: `MetaByPrefix(prefix string) (map[string]string, error)` —
  the one new store read the spend copy needs (R4); test beside the meta tests.
- `internal/world/fork.go` (new; the `migrate.go` sibling):
  `Fork(srcDir, destDir, newName string) (*ForkResult, error)`:
  1. `Open(srcDir)` (current-format gate applies — fork never crosses formats; migrate
     first); refuse live daemon via the pidfile/`IsRunning` check the same way `Migrate`
     does (import direction: `internal/world` must not import `internal/daemon` — reuse
     `Migrate`'s own liveness mechanism, `world.go`/`migrate.go`'s existing check).
  2. `store.Open(src DBPath)` read-only use: `LatestValidSnapshot()`; nil → refuse with
     the start-and-stop remedy (FR-002). Re-verify hash (the walk already does).
  3. Destination: require empty/absent (the `Create` posture); `os.MkdirAll` + `agents/`.
     On ANY later error: best-effort `os.RemoveAll(destDir)` (R9 cleanup note).
  4. Fresh `store.Open(dest DBPath)`: stream `ReplayEvents(0, …)` filtering
     `seq <= boundary.Seq` into batched `AppendEvents` (order preserved ⇒ seqs 1..N
     reproduce); append `world.forked` (tick=boundary.Tick — lands at seq boundary.Seq+1);
     `SaveSnapshot(boundary.Tick, boundary.Seq, boundary.State)`; `SetMeta` seed +
     format_version (matching `validateMeta`); copy `MetaByPrefix("llm_spend_")` rows
     (AC5/FR-012).
  5. Write the fork manifest: parent manifest with `Name=newName`,
     `CreatedAt=now`, `Lineage` set, everything else verbatim (FR-005; seed carried — R2).
  6. Sidecar copy per the R9 table (copy: llm.json, calibration.json,
     estimator_state.json, charter.md, metatron/, bundles/, tuning.json, agents/
     contents; skip: runtime files, wal/shm, migration archives, scribe views).
  7. Return `ForkResult` (data-model §4: TruncatedTail = parent LastSeq − boundary.Seq;
     BoundaryEnded from the unmarshaled boundary state's `Ended`; SpendCarried).
- `internal/world/fork_test.go`: ceremony happy path (contiguity, snapshot hash verifies,
  lineage event payload, manifest, spend keys — SC-006's meter assertion via
  `llm.NewMeter` on the fork store); refusals (no snapshot, non-empty dest, live pid);
  determinism pair (FR-010 a+b): genesis replay of fork log → hash == boundary
  `state_hash` == parent state hash at (tick, seq).

### D4 — CLI `fork` (cmd/promptworld)

- `cmd/promptworld/fork.go` (new) + `main.go` dispatch + usage block:
  `fork <world> <new-name> [--at latest-snapshot]`. Source via `resolveWorld`;
  `<new-name>` classified name-vs-path exactly as `new` does (bare →
  `worlds.ValidateName` + worlds-home placement; path → verbatim, name = basename);
  `--at` accepts only `latest-snapshot` (default), any other value → informed refusal
  naming the follow-on (FR-002). Prints the summary from `ForkResult`: boundary as
  day/HH:MM + tick (`clock.GameTime`), events carried, truncated tail when nonzero,
  lineage line, ended-boundary warning when set, spend-carried line, and
  `start <a>` / `start <b>` next-steps.
- `cmd/promptworld/fork_test.go` / e2e (`e2e/fork_e2e_test.go`, new): SC-001 —
  new → run past a snapshot (graceful stop cuts one) → fork → start BOTH → both answer
  status, `ps --json` shows both running → stop both; FR-010(c) — run the fork pure-sim
  at max, stop, replay its log from genesis, hash == its final snapshot's `state_hash`
  (the harness pattern, `determinism_e2e_test.go` + `snapshots.md` mechanics).

### D5 — Resolver export (tui; spec-072 contract preserved)

- `internal/tui/reportcard.go`: refactor per data-model §5 — `ResolveRubricFacts(state,
  def, pass)`, `RecordedPassFor(state, exercise)`, exported aliases
  `ReportCardFact`/`ReportCardMode` (+ exported mode consts), `RenderReportCard` wrapping
  `reportCardView` (`views.go` — renderer body untouched). `Model.resolveReportCardFacts`
  and `Model.recordedPassFor` become wrappers (Model's `runEnded()` folds into the
  state-driven `s.Ended` switch — verify the two reads agree on a mid-takeover replica;
  they both derive from replica state today).
- Existing spec-072 TUI tests are the no-behavior-change proof; add the cross-surface
  identity test in D6's compare tests (duel rows == postmortem rows, SC-004).

### D6 — CLI `compare` (the duel)

- `internal/worlds/probe.go`: extract `OfflineState(w *world.World) (*sim.State, int64,
  error)` (state + lastSeq) from `OfflineSnapshot`'s body; `OfflineSnapshot` calls it
  (`ps`/`status` behavior pinned by existing tests).
- `cmd/promptworld/compare.go` (new) + dispatch/usage: `compare <a> <b> [--since TICK]`.
  Build `duelReport` (data-model §6):
  - per side: `OfflineState`; exercise from manifest `Scenario` →
    `sim.ScenarioExercises` lookup (nil for ambient — honest no-scorecard note, US2-4);
    `tui.RecordedPassFor`; `tui.ResolveRubricFacts`; outcome via `sim.ExerciseOutcome`
    mapped through the plain-language vocabulary (data-model §7 — raw enums never print,
    FR-019).
  - window: lineage fork tick when either manifest's `Lineage.Parent` names the other
    (match on name + `parent_created_at` when present), else 0; `--since` overrides
    (FR-016).
  - divergence: stream both logs' post-window story events (exclude `daemon.*`,
    `clock.*`, `cog.*`, `llm.*`; compare tick/type/payload — R7); first mismatch →
    `divergence`; none → nil (the identical-since-fork line).
  - interleave: `chronicle.entry` events `from_tick >= since`, merged stable by
    FromTick, labeled per world; divergence marker inserted in timeline position.
  - render: header (+ running-world as-of note, + different-exercises note when ids
    differ — US2-5) → side-by-side/stacked `RenderReportCard` per side → divergence
    line → interleaved chronicle. All prose in the no-blame register (FR-020).
- `cmd/promptworld/compare_test.go`: fixture pair builders (synthetic logs on the
  `scenario_test.go` fixture model): SC-004 (winner all-✓ from pass, loser ✗ with
  `agent.died: N`; enum sweep over output; duel rows byte-equal to
  `tui.ResolveRubricFacts`-fed postmortem rows), SC-005 (divergence placement;
  machinery-only difference → NO divergence; zero-divergence line; interleave labels),
  US2-3/4/5 (live markers, ambient honesty, different-exercise note).

### D7 — Design reference (spec 047 gate)

No TUI page's rendered behavior changes (exports + wrappers only), but the branch touches
`internal/tui/` pinned sources: run `node scripts/check-tui-design.mjs --changed`,
re-verify + re-pin every flagged page (reportcard.go/views.go/digest.go pin the
report-card and digest-bearing pages). Amendments only where a page states something the
diff falsifies (none expected — verify, don't assume). The fork/compare CLI verbs are not
TUI pages; no new design page is created.

### D8 — Wiki + player docs (in-branch, pr-gate enforced)

- `/grounding-wiki:wiki-update` over the branch diff. Expected review-work notes:
  `world-save-directory.md` (fork ceremony joins Create/Open/Migrate),
  `world-save-manifest-fields.md` (lineage block), `world-save-path-helpers.md`
  (copy/skip catalog reference), `cli-world-lifecycle.md` or `cli-promptworld.md` (two
  new subcommands — placement per the existing split), `event-log.md`/`snapshots.md`
  (MetaByPrefix; boundary-snapshot reuse), `sim-state-reducer.md` + the clock/world
  event-types child (`world.forked`), `instance-manager.md` (OfflineState extraction),
  `report-card-renderer.md` (fourth consumer + exported surface),
  `llm-budget-degraded-mode.md` (fork wallet-inheritance note; sources may be
  pin-unchanged — re-verify), `deterministic-rng.md` (re-verify only: the fork is a new
  consumer of the replay contract). Computed re-pins for the rest. A new note for the
  fork/duel itself if the update plan calls for one (`world-forking` concept note —
  wiki-update's judgment).
- Regenerate `docs/player/` if any wiki note changes (`player-docs` skill; probe:
  `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`).
- Gate: `node scripts/check-merge-drift.mjs pr` exits 0 from the worktree before the PR;
  merge with `gh pr merge --merge` only.

## Project Structure

### Documentation (this feature)

```text
specs/076-world-fork-duel/
├── CLAIM.md          # claim stub (spec 065) — kept
├── spec.md
├── research.md       # decisions R1–R9
├── data-model.md     # shapes §1–7
├── plan.md           # this file
└── tasks.md
```

### Source Code (repository root)

```text
internal/sim/state.go              # WorldForkedPayload + no-op arm (D1)
internal/tui/digest.go             # world.forked digest (D1)
internal/world/world.go            # Manifest.Lineage + Open validation (D2)
internal/world/fork.go             # NEW — the fork ceremony (D3)
internal/world/fork_test.go        # NEW — ceremony + determinism proofs (D3)
internal/store/store.go            # MetaByPrefix (D3)
internal/worlds/probe.go           # OfflineState extraction (D6)
internal/tui/reportcard.go         # resolver export + wrappers (D5)
internal/tui/views.go              # RenderReportCard export shim (D5)
cmd/promptworld/main.go            # dispatch + usage: fork, compare (D4/D6)
cmd/promptworld/fork.go            # NEW — cmdFork (D4)
cmd/promptworld/compare.go         # NEW — cmdCompare, duelReport (D6)
cmd/promptworld/{fork,compare}_test.go   # NEW (D4/D6)
e2e/fork_e2e_test.go               # NEW — SC-001 + FR-010(c) (D4)
internal/{sim,world,store,worlds,tui}/*_test.go  # alongside (D1–D6)
docs/design/tui/**                 # re-pins per --changed (D7)
docs/wiki/** · docs/player/**      # D8
```

**Structure Decision**: every new mechanism lives in the package that owns its precedent —
fork beside migrate (`internal/world`), the lineage payload beside the genesis payloads
(`internal/sim`), the duel beside the other CLI verbs (`cmd/promptworld`), the resolver
where spec 072 put it (`internal/tui`). No new packages; two new CLI files, one new world
file, one new e2e file.

## Complexity Tracking

Empty — no constitution violations.
