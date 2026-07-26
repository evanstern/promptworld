# Implementation Plan: In-TUI forward-ladder view

**Branch**: `078-tui-ladder-view` (task branch: `task-152-tui-ladder-view`) | **Date**: 2026-07-26 | **Spec**: [spec.md](spec.md)

## Summary

Append a forward-ladder block to the `?` overlay's guardian section
(`helpGuardianLines`, `internal/tui/help.go`): all stages from
`world.StageOrder`/`world.StagesLadder`, earned/next from the per-user
unlocks record via an earned rule RELOCATED to `internal/worlds` so the CLI
(`stages --json`) and the TUI render from one substrate — parity by
construction, proven by a runtime-derived parity test with zero hardcoded
stage inventory (the TASK-151 armor). Docs rider same-PR: help.md Section 5 +
byte-identity row + control-table row, re-pinned, design gate green.

## Technical Context

**Language**: Go (TUI: bubbletea Model, pure render helpers). **Testing**:
`go test ./internal/tui/ ./internal/worlds/ ./cmd/promptworld/`; full
`go test ./...`; fixture unlocks records via the `setHome(t)` precedent
(`internal/tui/tui_test.go:27`). **Scope**: `internal/tui/help.go` +
`tui.go` (one Model field + boot load) + `help_test.go`;
`internal/worlds/unlocks.go` (+test); `cmd/promptworld/stages.go` (consume
the relocated rule); `docs/design/tui/overlays/help.md`; wiki re-pins +
player-docs regen as touched. **Constraints**: no per-frame disk reads; no
IPC/LLM from render paths; no change to `stages --json`'s output shape; the
overlay's exact-height/pager discipline unchanged.

## Constitution Check

- **I. Artifact-grounded** — PASS: decision trail = reorient synthesis
  (decision 6 / position 5) → TASK-152 → this spec dir (claimed per spec
  065).
- **II. One task, one PR** — PASS: TASK-152 ↔ `task-152-tui-ladder-view` ↔
  one PR from `.worktrees/task-152`; docs rider and wiki re-pins ride the
  same PR.
- **III. Gates over assertions** — PASS: design gate (`check-tui-design.mjs
  --changed`), merge-drift worktree/pr gates (`--spec 078 --task TASK-152`),
  spec-bridge mirrors phases to the board.
- **IV. Grounding freshness** — PASS (planned): `internal/tui/help.go` is a
  pinned source of `docs/wiki/tui-input-help.md` (re-pin required);
  `cmd/promptworld/stages.go` pins `cli-world-lifecycle.md` and
  `curriculum-ladder.md`; `internal/worlds/unlocks.go` pins
  `curriculum-ladder-progression.md` — every touched-source note re-verified
  IN THIS BRANCH; `docs/player/` regenerated if `docs/wiki/` changes;
  merge-commit-only (`gh pr merge --merge` — in-branch pins are branch
  hashes).
- **V. Model tiers** — PASS: **Sonnet** (recorded on TASK-152): single-
  package deterministic TUI view on relocated-for-this substrate; no
  concurrency, no doctrine-adjacent behavior. Escalation rubric not
  triggered.

**Post-Phase-1 re-check**: PASS.

## Design

- **D1 — shared earned rule** (spec FR-003): add to
  `internal/worlds/unlocks.go`:
  `func (u *Unlocks) StageEarned(stage string) bool { return stage == world.Stage1 || u.Earned(stage) }`
  (`internal/worlds` already imports `internal/world`; nil-receiver safe
  like `Earned`). `cmd/promptworld/stages.go` deletes its local
  `stageEarned` and calls the method (both `cmdStages` and
  `highestEarnedStage`). No output change to `stages` in either form.
- **D2 — Model plumbing** (FR-006): `Model.unlocks *worlds.Unlocks`, loaded
  once in `New()` (the `populateHelpLessons` boot precedent; `LoadUnlocks`
  is already load-tolerant/nil-safe). Render-time earned predicate:
  `m.unlocks.StageEarned(id) || slices.Contains(m.replica.StagesUnlocked, id)`
  (replica union = mid-session liveness; audit pointer only from the record
  entry, which is the only place it exists).
- **D3 — the block** (FR-001/002/004/007/009): a new `helpLadderLines`
  helper appended by `helpGuardianLines` under a "The ladder" header,
  iterating `world.StageOrder`. Per stage: identity line
  (`m.sk().Stage(id)` name + line + `(id)`), `teaches:` concept,
  `unlocked by:` evidence (graduation wording when empty — the `stages`
  command's exact posture), `earned:` yes-floor / yes-with-audit-pointer /
  next / not yet, and the you-are-here marker (`◀ this world`, with
  `by override — not earned` appended when `m.status.World.StageOverridden`).
  The current-stage teaching content above keeps its `grants:` line as
  today; forward rows render identity + concept + evidence + earned/next
  (the board AC's four facets; grants prose per forward row is available in
  `stages` and would triple the section height — spec Assumptions). All
  lines `wrapText`/`clipLine` to width; the section pages via the existing
  `paginateHelpContent`.
- **D4 — parity test** (FR-005, the deliverable's proof): in
  `help_test.go`, build a fixture unlocks record under `setHome(t)` (one
  record-earned stage with world/exercise), then for each `id` in
  `world.StageOrder` compute the expected row from the SAME substrate
  `cmdStages --json` marshals — `skin.Stage(id)` (default-skin fixture
  model), `world.StagesLadder[id]`, `unlocks.StageEarned(id)`,
  `unlocks.Entries[id]` — and assert name, concept, evidence (or graduation
  wording), earned/next marker, and audit pointer all appear in the
  rendered guardian-section lines. Zero literal stage ids/counts/prose in
  the test body: everything ranges over the substrate, so a TASK-151
  catalog change flows through both surfaces and the test untouched.
- **D5 — byte-identity + degradation tests** (FR-008/010): extend
  `TestHelpGuardianByteIdenticalPerStatus` (same stage + same
  unlocks/replica ⇒ identical bytes across renders; overridden vs not are
  DIFFERENT fixed inputs, each internally constant);
  `TestHelpContentReadsNoStatusOrReplica` must stay green with the new
  content (nil status, nil replica, no unlocks file ⇒ floor ladder, no
  panic); a mid-session case asserts a replica-only unlock shows earned
  without an audit pointer. Note `TestHelpNoLLMByteIdentical` already
  excludes the guardian section from its nil-vs-status sweep (it is
  stage-keyed by design) — unchanged.
- **D6 — docs rider** (FR-011): amend
  `docs/design/tui/overlays/help.md` — Section 5 gains "The ladder"
  content-contract prose; the byte-identity classification table gains:
  `forward ladder (guardian section)` — **unlocks-record-derived,
  model-free** (catalog text static; earned/next from the per-user record +
  live `StagesUnlocked`; you-are-here from the stage/override scalars;
  never LLM-derived; absent record ⇒ stage-1 floor); the control table
  gains the row (data source `world.StageOrder`/`StagesLadder` +
  `Model.unlocks`/`replica.StagesUnlocked`, renderer `helpLadderLines`,
  introduced-by reorient D6 / spec 078). Re-verify + re-pin
  (`verified_against` → branch commit); `check-tui-design.mjs --changed`
  green.
- **D7 — grounding** (SC-004): re-verify + re-pin in-branch:
  `docs/wiki/tui-input-help.md` (help.go, tui.go — sources touched),
  `cli-world-lifecycle.md` + `curriculum-ladder.md` (stages.go touched by
  D1), `curriculum-ladder-progression.md` (unlocks.go touched by D1).
  `docs/wiki/` changes ⇒ regenerate `docs/player/` (player-docs skill) and
  pass `node .claude/skills/player-docs/scripts/check-freshness.mjs
  --check`.
- **D8 — gates & merge**: worktree cut via `node
  scripts/check-merge-drift.mjs worktree --spec 078 --task TASK-152`; PR
  from the worktree only after `node scripts/check-merge-drift.mjs pr`
  exits 0; merge with `gh pr merge --merge` (merge-commit-only); post-merge
  main commits are derived state only.

## Project Structure

### Documentation (this feature)

```
specs/078-tui-ladder-view/
├── CLAIM.md    # spec 065 claim stub (pre-existing, kept)
├── spec.md
├── plan.md
└── tasks.md
```

### Source Code (repository root)

```
internal/worlds/unlocks.go        # D1: StageEarned method (+ unlocks_test.go)
cmd/promptworld/stages.go         # D1: consume the shared rule, delete local copy
internal/tui/tui.go               # D2: Model.unlocks field + New() load
internal/tui/help.go              # D3: helpLadderLines, appended by helpGuardianLines
internal/tui/help_test.go         # D4/D5: parity + byte-identity/degradation tests
docs/design/tui/overlays/help.md  # D6: Section 5 + byte-identity row + control row, re-pin
docs/wiki/*.md                    # D7: re-pins as touched (tui-input-help at minimum)
docs/player/                      # D7: regenerated iff docs/wiki changes
```

## Complexity Tracking

Empty — no constitution violations to justify.
