# Implementation Plan: Guardian Inventory Tool

**Spec**: `specs/116-guardian-inventory-tool/spec.md`
**Board Task**: TASK-197
**Branch**: `task-197-guardian-inventory-tool`

## Summary

Three pieces on one branch, in dependency order:

1. **The eye** — a per-villager inventory CONTENTS mirror in the guardian's absorb snapshot,
   and `inspect_pack`, a charge-free Read tool that renders it as a deterministic sheet. Pure
   addition; nothing existing changes behavior.
2. **The discipline** — a per-turn look-first ledger on `turnDispatch` and a gate in the
   `send_vision` / `work_miracle` handlers that fires only on survival-origin turns. Behavior
   change, narrowly scoped, expressed as the existing repairable `rejected_gate`.
3. **The hand** — a `take_item` miracle kind, its `guardian.item_taken` event and reducer arm,
   its situated memory, and its rendering. The `give_item` mirror throughout, reusing the
   `agent.dropped` inventory→pile machinery so nothing is unmade.

Piece 1 is independently shippable and independently testable. Piece 2 depends on piece 1's
handler existing (the ledger is populated there). Piece 3 is independent of 1 and 2 except
that the gate covers it, so it lands last and the gate's `take_item` arm lands with it.

## Technical Context

**Language**: Go (module `github.com/evanstern/promptworld`).
**Packages touched**: `internal/tool` (registry, roster, derive), `internal/guardian` (mirror,
handlers, new sheet file, charter ceiling), `internal/sim` (payload, reducer arm, miracle
batch), `internal/tui` (digest rendering), `internal/skin` (example ask).
**Testing**: `go test ./...`; new tests live beside the code they cover
(`internal/guardian/pack_test.go`, `internal/sim/miracles_test.go` additions,
`internal/tool/registry_test.go` additions).

**Key existing seams reused rather than reinvented**:

| Need | Existing seam |
| --- | --- |
| Read-tool shape (free, deterministic, honest miss) | `handleSurvey` / `buildSurveySheet` (`internal/guardian/survey.go`) |
| Guardian-side state snapshot | `mirrorState` (`internal/guardian/guardian.go:454`) |
| Repairable refusal | `toolloop.VerdictRejectedGate` + `refusal()` (`internal/guardian/toolcalls.go`) |
| Survival turn identity | `turnOrigin{survival: true}` (`internal/guardian/orders.go:714`) |
| Miracle + memory atomic batch | `BuildMiracleBatch` (`internal/guardian/miracle_batch.go`) |
| Validate-then-spend-then-mutate | `applyItemGranted` (`internal/sim/miracles.go:441`) |
| Inventory → tile pile transfer | the `agent.dropped` arm (`internal/sim/state.go:1708`), `pileFor` (`state.go:569`) |
| Stage ceiling membership | `stage1CeilingTools` (`internal/guardian/charter.go:718`) |

## Constitution Check

- **Principle V (Model-Tiered Workflow)**: this slice is **doctrine-adjacent** — the Guardian
  gains a new way to reach into the world (write access to a villager's pack) and a new
  structural refusal on its own survival turn. Per the rubric that is an **Opus 5 escalation**:
  dispatch `.claude/agents/spec-implementer-opus.md`, not `spec-implementer`. The tier choice
  and this justification are recorded on TASK-197.
- **Spec rigor**: full Spec Kit; spec linked to the board via `spec-bridge:link` before
  implementation starts.
- **Root checkout read-only**: all work on `task-197-guardian-inventory-tool` in
  `.worktrees/task-197`; lands via one PR, merge-commit only.
- **One TASK, one PR**: the three pieces are internal breakdown, not PR boundaries.
- **Wiki-in-PR**: any wiki note whose pinned sources this branch touches is re-verified and
  re-pinned in-branch, and `docs/player/` is regenerated if `docs/wiki/` changes — the pr gate
  blocks otherwise.

## Project Structure

### Documentation (this feature)

```
specs/116-guardian-inventory-tool/
├── spec.md
├── plan.md
├── tasks.md
└── contracts/
    └── pack-access.md
```

### Source code

```
internal/tool/
├── registry.go      # inspect_pack Tool entry; take_item in miracleKinds/costs/kindToEvent
├── roster.go        # RosterGuardian + LoopRosterGuardian append inspect_pack
└── derive.go        # guardianToolDesc entries; give_item hint names take_item

internal/guardian/
├── guardian.go      # mirrorState: per-villager inventory contents
├── pack.go          # NEW — handleInspectPack + buildPackSheet
├── toolcalls.go     # handler wiring; look-first gate in vision/miracle handlers
├── turn.go          # turnDispatch carries survival origin + the look-first ledger
├── miracle_batch.go # take_item arm + takeMemoryText
├── charter.go       # inspect_pack joins stage1CeilingTools
└── pack_test.go     # NEW — sheet, determinism, honest miss, gate, world-03 fixture

internal/sim/
├── payloads.go      # ItemTakenPayload registration
├── miracles.go      # applyItemTaken
└── state.go         # dispatch arm for guardian.item_taken

internal/tui/digest.go   # line renderer + subject resolver for guardian.item_taken
internal/skin/skin.go    # example ask for inspect_pack
```

## Phase sequencing

| Phase | Content | Blocks |
| --- | --- | --- |
| 1 | Setup: confirm baseline green | all |
| 2 | Foundational: contents mirror (FR-006) | 3 |
| 3 | US1 — `inspect_pack` tool + sheet (FR-001…005) 🎯 MVP | 4, 5 |
| 4 | US2 — look-first ledger + gate (FR-007, FR-008) | — |
| 5 | US3 — `take_item` miracle + reducer + pile transfer (FR-009…012) | 6 |
| 6 | US4 — memory, audit, rendering (FR-013, FR-014, FR-017, FR-018) | — |
| 7 | Cross-cutting: rosters, ceiling, guidance, skin (FR-015, FR-016) | — |
| 8 | Grounding: wiki re-pin, player docs, board | — |

Phases 3 and 5 are the two independently valuable slices: 3 alone stops the Guardian speaking
blind; 5 alone unlocks the full pack. 4 is what makes 3 actually get used.

## Risks and how they are contained

- **The gate traps the Guardian in an unsatisfiable loop.** Contained by: `inspect_pack` never
  errors (always `read_ok`), never spends the act, and the refusal names the exact repair. A
  survival turn therefore always has a path to acting. Test: the gated call, then the repair,
  then the landing, all inside one turn's round cap.
- **Removal silently destroys goods.** Contained by FR-011 + SC-005's conservation invariant,
  asserted as a test over (inventory + pile) totals.
- **Spear/axe wear or food spoilage corrupted by the transfer.** Contained by reusing the
  `agent.dropped` arm's exact ordering rules, with direct tests on both.
- **Drift between the tool vocabulary and the door's accept set.** The repo already pins these
  with cross-check tests (`TestMiracleKindsMirrorTool`, the `GrantKinds` drift test); the new
  kind must extend those, not add a parallel literal.
- **A stage-1 world gaining a world-shaping verb by accident.** Contained by the explicit
  ceiling test (`stage_test.go`) which enumerates the granted roster — it will fail loudly if
  `take_item` leaks in or `inspect_pack` is left out.

## Complexity Tracking

No constitutional deviations. The one judgment call worth naming: the look-first gate is state
held on the turn (a per-turn ledger) rather than derived from the event trail. That is the
minimum sufficient mechanism — the trail cannot answer "did the model look during THIS loop"
without correlating `cog.tool_call` records mid-turn, which would couple the door to telemetry.
The ledger is in-memory, per-turn, and never persisted, so it adds no replay surface.
