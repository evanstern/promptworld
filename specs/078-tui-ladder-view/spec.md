# Feature Specification: In-TUI forward-ladder view in the `?` guardian section

**Feature Branch**: `078-tui-ladder-view` (task branch: `task-152-tui-ladder-view`)

**Created**: 2026-07-26

**Status**: Draft

**Input**: TASK-152 (reorient 2026-07-26 decision 6; merged position 5 — the
WorldBox discoverability critique). The forward ladder (identity · concept ·
earned/next · unlock evidence, matching `stages --json`) renders only in the
CLI today; a TUI player can see where they are (`helpGuardianLines` shows the
CURRENT stage) but never what's next. Docs-branch invariant rider: the view is
status-derived, so `docs/design/tui/overlays/help.md`'s byte-identity table
gains a row in the same PR (board AC #2).

## Grounding (verified against main after PR #115, TASK-142's merge)

- `internal/tui/help.go` — `helpGuardianLines` (Section 5, D9/spec 063)
  renders the current stage's identity + concept + grants + granted verbs +
  example asks. It reads `world.StagesLadder[stage]` for ONE stage only; the
  three forward stages and all earned/next/unlock-evidence state are absent.
  Nil status renders the pre-ladder variant. The overlay's pager
  (`paginateHelpContent`) already handles arbitrary-length sections.
- `internal/world/world.go` — `StageOrder` (presentation order, all stages)
  and `StagesLadder` (concept/grants/unlock-evidence per stage), RELOCATED
  from `cmd/promptworld/stages.go` by spec 063 T014 **for exactly this
  consumer** ("one source, two surfaces").
- `internal/worlds/unlocks.go` — `LoadUnlocks` (load-tolerant per-user
  record; nil/missing/corrupt ⇒ nothing earned) and `Unlocks.Earned`.
  `internal/worlds` already imports `internal/world`.
- `cmd/promptworld/stages.go` — `cmdStages`: the `--json` row shape
  (`stageJSON`: id, name, line, concept, grants, unlock_evidence, earned,
  proving_world, exercise) and the earned rule (`stageEarned`: stage-1 is
  every player's unconditional floor; every other stage needs a record
  entry). `stageEarned`/`highestEarnedStage` live in package main today.
- `internal/ipc/protocol.go` — `WorldStatus.Stage` / `.StageOverridden`
  mirror the manifest (spec 046 FR-002/FR-003); the TUI reads the stage via
  `Model.currentStage()` (`views.go:250`).
- `internal/sim/curriculum.go` — `StagesUnlocked` latches the newly-unlocked
  stage id (e.g. `"stage-2"`), the SAME key vocabulary as
  `Unlocks.Entries` — a mid-session unlock is visible in the replica before
  (and regardless of whether) the per-user record write lands.
- `internal/tui/help_test.go` — the byte-identity suite:
  `TestHelpNoLLMByteIdentical` (keys/walkthrough/lessons stay nil-status
  byte-identical), `TestHelpGuardianByteIdenticalPerStatus` (guardian bytes
  constant per stage scalar), `TestHelpContentReadsNoStatusOrReplica`
  (every section renders non-empty with nil status AND nil replica).

## The TASK-151 hazard, and the parity contract's shape

TASK-151 (exercise catalog, spec 077, in flight) adds exercises and the
`skills_observed` event and may extend stage gating; it may merge before this
feature implements. Therefore this spec's parity contract is written against
`stages --json`'s **output at implementation time** — the rendered ladder and
the parity test MUST derive their expected rows at runtime from the same
substrate `cmdStages` marshals (`world.StageOrder` × `world.StagesLadder` ×
the shared earned function × the stage skin lookup), and MUST NOT hardcode a
stage inventory, stage count, or per-stage prose. Whatever `stages --json`
says when this branch cuts, the ladder block says the same — 151's merge
cannot stale this spec.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A TUI player sees the whole ladder ahead of them (Priority: P1)

A player presses `?`, tabs to the guardian section, and below the
current-stage teaching content sees the full forward ladder: every stage's
identity, the concept it teaches, its earned/next state (with the audit
pointer for record-earned stages), and the evidence that unlocks it — the
same informed identity table `promptworld stages` prints, never a difficulty
menu (spec 046 FR-003 posture). Deterministic floor, model-free: it renders
identically with no LLM configured.

**Why this priority**: board AC #1; decision 6 closes the WorldBox
discoverability critique — a TUI-only player currently has no in-game way to
learn what's next or how to earn it.

**Independent Test**: on a stage-1 world with an empty unlocks record, open
`?` → guardian section → the ladder lists every stage `StageOrder` names,
stage-1 earned (floor), the next stage marked next with its unlock evidence,
later stages visible with identity + concept + evidence stated.

**Acceptance Scenarios**:

1. **Given** a fresh player (no unlocks record) on a stage-1 world, **When**
   they open the guardian section, **Then** the ladder shows all stages in
   `StageOrder` order, stage-1 earned, stage-2 marked as next with its
   unlock evidence, and stages 3+ forward-visible — field-for-field what
   `stages --json` reports for the same record.
2. **Given** a record that earned stage-2 in world W via exercise E, **When**
   the ladder renders, **Then** stage-2 shows earned with the audit pointer
   (proving world + exercise — the FR-008/unlocks-record rule-5 pointer
   `stages --json` carries as `proving_world`/`exercise`), and stage-3 is
   marked next.
3. **Given** a stage unlock happens mid-session in THIS world (the ceremony
   fires; `replica.StagesUnlocked` gains the stage), **When** the ladder
   next renders, **Then** that stage shows earned (no client restart, no
   per-frame disk read) — audit pointer shown only once the record entry
   exists.
4. **Given** no LLM is configured and status is nil, **When** the guardian
   section renders, **Then** the ladder still renders (forward-looking
   content is per-user, not per-connection) with no you-are-here marker,
   and nothing errors or blanks.

---

### User Story 2 - The design authority stays true in the same PR (Priority: P2)

`docs/design/tui/overlays/help.md` is amended in the same PR: Section 5
documents the ladder block, the byte-identity classification table gains the
ladder's row, the control table gains its renderer row, and the page is
re-verified + re-pinned. `node scripts/check-tui-design.mjs --changed`
passes.

**Why this priority**: board AC #2 — the docs-branch invariant rider from
merged position 5; the design gate (spec 047) blocks the PR otherwise.

**Independent Test**: `node scripts/check-tui-design.mjs --changed` exits 0
on the branch; help.md's byte-identity table names the ladder row; the
control table's new row names the real renderer symbol.

**Acceptance Scenarios**:

1. **Given** the amended help.md, **When** the design gate runs `--changed`,
   **Then** it passes and the page's `verified_against` is a branch commit.
2. **Given** the byte-identity table, **When** a reader checks the ladder
   row, **Then** its classification states exactly what varies (per-user
   unlocks record + live `StagesUnlocked` + the stage/override scalars) and
   what never does (catalog text; model-free; no LLM ever).

---

### Edge Cases

- **Stage-overridden world** (`stage_overridden` manifest flag →
  `WorldStatus.StageOverridden`): the world runs at an unearned stage. The
  you-are-here marker annotates the row as running by override; the row's
  EARNED state still comes from the record — the ladder never launders an
  override into an earned claim (the `stages` command's honesty posture).
- **Pre-ladder world** (stage `""`): the guardian section's existing
  pre-ladder variant keeps its "ungated" framing; the ladder block still
  renders (it is per-user forward content, not per-world), with no
  you-are-here marker.
- **Nil status** (disconnected / no-LLM floor): same as pre-ladder — ladder
  renders, no marker, never blank (`TestHelpContentReadsNoStatusOrReplica`
  must stay green).
- **Unlocks file absent / corrupt / home unresolvable**: `LoadUnlocks`'s
  documented degrade — nothing earned beyond the stage-1 floor; render the
  honest fresh-player ladder, never an error line.
- **Ordinary vs scenario worlds**: the ladder block is identical on both
  (per-user record + static catalog); only the you-are-here/override
  annotations depend on the attached world. No exercise-tab coupling.
- **Narrow layout / small terminal**: the guardian section grows several
  lines per stage; content wraps to width and pages through
  `paginateHelpContent` (J/K) — usable at 80×24, never overflowing the
  panel budget.
- **Stage-4 / graduation**: empty `UnlockEvidence` renders the graduation
  wording (the `stages` command's "nothing — this is graduation" fact), not
  an empty field.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The `?` guardian section MUST gain a forward-ladder block
  rendering EVERY stage `world.StageOrder` names, in that order — derived by
  iteration, never a hardcoded stage list/count (the TASK-151 armor).
- **FR-002**: Each stage row MUST render: skin-resolved identity (the same
  stage skin lookup family both surfaces use — name + line), the concept
  taught, the unlock evidence (graduation wording when empty), and the
  earned/next state; a record-earned stage additionally renders the audit
  pointer (proving world + exercise) exactly when `stages --json` would
  populate `proving_world`/`exercise`.
- **FR-003**: The earned rule MUST be single-sourced: relocate
  `stageEarned`'s logic (stage-1 unconditional floor; otherwise a record
  entry) from `cmd/promptworld/stages.go` into `internal/worlds`, and
  `cmdStages`/`highestEarnedStage` MUST consume the shared function (the
  package-main copy deleted) — parity by construction, the spec 063 T014
  precedent extended from catalog content to earned state.
- **FR-004**: The "next" marker MUST mark exactly the first unearned stage
  in `StageOrder` (a pure derivation of the earned flags — `stages --json`
  carries no `next` field; this is presentation, not new state).
- **FR-005**: Parity test (the deliverable's proof): a test in
  `internal/tui` MUST derive expected rows at runtime from the same
  substrate `stages --json` marshals (`StageOrder` × `StagesLadder` × the
  shared earned function × the stage skin lookup, against a fixture unlocks
  record) and assert every field of every row surfaces in the rendered
  block — including the audit pointer and graduation wording — with zero
  hardcoded stage ids, counts, or catalog prose.
- **FR-006**: Deterministic and model-free: the unlocks record loads ONCE at
  client boot (`New()` — the `populateHelpLessons`/lesson-seen-state
  precedent), never per frame; the rendered earned state is the boot-loaded
  record UNIONED with live `replica.StagesUnlocked` (same key vocabulary —
  mid-session unlocks show without re-reading disk); no LLM call, no IPC
  command, no event emission.
- **FR-007**: You-are-here: when `currentStage()` is non-empty, its row is
  marked as the attached world's stage; when `WorldStatus.StageOverridden`
  is true the marker states the override and the row's earned state remains
  record-derived. Nil status / pre-ladder ⇒ no marker, ladder still renders.
- **FR-008**: Degradation floor: nil `Unlocks` (missing/corrupt/unresolvable
  home) renders the stage-1-floor-only ladder; nil status AND nil replica
  render non-empty content (the existing construction check must stay
  green).
- **FR-009**: Layout: rows wrap to the panel width (`wrapText`) and page
  through the shared pager; the section is fully readable at 80×24 by
  scrolling.
- **FR-010**: Byte-identity: for a fixed (stage scalar, override flag,
  unlocks snapshot, `StagesUnlocked`, world skin) the block's bytes are
  constant across renders — extend
  `TestHelpGuardianByteIdenticalPerStatus`'s guarantee to the ladder inputs.
- **FR-011** (US2): `docs/design/tui/overlays/help.md` amended same-PR:
  Section 5 documents the ladder block; the byte-identity classification
  table gains the ladder row (class: unlocks-record-derived, model-free —
  catalog text static; earned/next columns come from the per-user record +
  live `StagesUnlocked`; you-are-here from the stage/override scalars;
  never LLM-derived); the control table gains the ladder's renderer row;
  page re-verified + re-pinned; `check-tui-design.mjs --changed` passes.

### Key Entities

- **Ladder row** — one stage's rendered facts: skin identity (name, line),
  concept, unlock evidence, earned/next state, optional audit pointer
  (proving world, exercise), optional you-are-here/override annotation.
- **Shared earned function** — `internal/worlds`: stage-1 floor ∨ record
  entry; consumed by both `cmdStages` and the TUI ladder.
- **Boot-loaded unlocks snapshot** — new `Model` field populated in `New()`
  via `worlds.LoadUnlocks`; unioned with `replica.StagesUnlocked` at render.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Board AC #1 — the guardian section renders the full ladder
  with earned/next state and unlock evidence; the FR-005 parity test passes,
  and passes unchanged if the stage catalog's content changes upstream
  (proven by its zero-hardcoding construction).
- **SC-002**: Board AC #2 — help.md's byte-identity table row + Section 5 +
  control-table row land in the same PR; `check-tui-design.mjs --changed`
  exits 0.
- **SC-003**: The existing help suite stays green unmodified in intent:
  nil-status/nil-replica tolerance (`TestHelpContentReadsNoStatusOrReplica`),
  keys/walkthrough/lessons byte-identity, guardian per-stage byte-identity
  (extended per FR-010). `go test ./...` green; `gofmt -l` clean.
- **SC-004**: Wiki notes whose pinned sources this branch touches are
  re-verified + re-pinned IN THIS BRANCH (`docs/wiki/tui-input-help.md` at
  minimum — `internal/tui/help.go` is a pinned source; probe
  `cli-world-lifecycle.md` (stages.go), `curriculum-ladder.md`
  (world.go/stages.go), `curriculum-ladder-progression.md` (unlocks.go));
  `docs/player/` regenerated if `docs/wiki/` changes; the merge-drift pr
  gate exits 0.

## Assumptions

- **TASK-151 may merge first**: absorbed by construction (the runtime-derived
  parity contract above). If 151 renames or restructures the substrate
  symbols themselves, this spec's references update at implementation time —
  the contract (same substrate, two surfaces, zero hardcoding) is the stable
  part.
- **Skin divergence is presentation, not parity**: `stages --json` resolves
  identity through the default skin (`skin.Stage`); the TUI resolves through
  the boot-frozen world skin (`m.sk().Stage`), consistent with the guardian
  section's existing rendering. Substrate fields (concept, grants where
  shown, unlock evidence, earned, audit pointer) are parity byte-for-byte;
  identity is parity under the default skin, which is what the parity test
  fixtures use. A custom-skinned world intentionally shows ITS stage names.
- **Grants line**: the current-stage teaching content already renders
  `grants:` for the current stage; whether every forward row repeats its
  full grants prose or the ladder leans on concept + evidence (the
  `stages` CLI shows all fields; panel width is scarcer) is a rendering
  decision recorded in plan D3 — the parity test asserts the fields the
  block CONTRACTS to show, and the block contracts to concept + evidence +
  earned/next + identity as the board AC names, with grants available to
  the current-stage block as today.
- The unlocks record's key vocabulary (`stage-2`…) matches
  `replica.StagesUnlocked` (verified: `internal/sim/curriculum.go:145`,
  `unlocksFile`), making the FR-006 union sound.
