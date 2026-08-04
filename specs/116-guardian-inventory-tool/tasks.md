# Tasks: Guardian Inventory Tool

**Spec**: `specs/116-guardian-inventory-tool/spec.md` · **Plan**: `plan.md` ·
**Contracts**: `contracts/pack-access.md` · **Board Task**: TASK-197

Every task cites the requirement it discharges. A task is done when its test asserts the
behavior, not when the code compiles.

## Phase 1: Setup

- [X] **T001** Confirm the baseline is green on the branch: `go test ./...` passes before any
  change. Record the result; a pre-existing failure is a finding to surface, not to fix here.

## Phase 2: Foundational (blocking prerequisites)

- [X] **T002** (FR-006) Add a per-villager inventory CONTENTS snapshot to the guardian mirror.
  In `internal/guardian/guardian.go`, alongside `agentNeeds`, carry `agentInv []sim.Inventory`
  (or an equivalent value copy), refreshed in `mirrorState` from
  `mt.replica.Agents[i].Inv` in the SAME loop that fills `needMirror`. Slices inside the
  inventory (`Spears`, `Axes`) MUST be copied, not aliased — the replica keeps mutating them.
  Comment it the way `needMirror`'s Bulk field is commented: why the mirror exists (the turn
  worker never touches the replica) and what it must not drift from.
- [X] **T003** (FR-006) Test: `internal/guardian` — after absorbing a batch that changes a
  villager's inventory, the mirrored contents match, and mutating the replica's `Spears` slice
  afterwards does NOT change the mirrored copy (aliasing regression).

## Phase 3: User Story 1 — the Guardian can open a pack (P1) 🎯 MVP

- [X] **T004** (FR-001) Register `inspect_pack` in `internal/tool/registry.go`: Effect `Read`,
  Gate `None`, one required string param `villager`. Follow `survey_site`'s entry exactly for
  effect/gate/param style.
- [X] **T005** (FR-001, FR-015) Append `inspect_pack` to `RosterGuardian` and
  `LoopRosterGuardian` (`internal/tool/roster.go`) and to `stage1CeilingTools`
  (`internal/guardian/charter.go`). Append LAST in each so no existing tool's registration
  position shifts. Update the roster comment blocks explaining WHY it joins the ceiling
  (read-only, widens no acting capability — the `survey_site`/`brief_myths` precedent).
- [X] **T006** (FR-002, FR-003) NEW `internal/guardian/pack.go`: `buildPackSheet` renders the
  sheet per `contracts/pack-access.md` §1 — fixed kind order, non-zero kinds only, spear/axe
  remaining-uses rendering, the mandatory food line, `sim.BulkCap` never a literal, free bulk
  floored at 0. Pure function of (name, mirrored contents); no clock reads.
- [X] **T007** (FR-001, FR-004, FR-005) In the same file, `handleInspectPack` — the
  `handleSurvey` dispatch shape: always `VerdictReadOK`; an unknown/dead/missing villager
  returns the honest miss naming the living roster. Wire it into `turnHandlers`
  (`internal/guardian/toolcalls.go`) under `d.grant.allows("inspect_pack")`.
- [X] **T008** (FR-004) Confirm — and assert in test — that `inspect_pack` consumes no charge,
  lands no event, and does not consume the turn's one act. If the loop driver's read exemption
  keys on Effect Read this is free; verify rather than assume.
- [X] **T009** (SC-001, AC#1, AC#6) Test `internal/guardian/pack_test.go`: the **world-03
  fixture** — a living villager carrying 20 wood and 4 planks and nothing else. Assert the
  sheet names `wood: 20`, `planks: 4`, `carrying 24/24`, `0 free`, and the literal "carries no
  food" line. This is AC#6's required exercise.
- [X] **T010** (FR-003, SC-001) Test: two identical `inspect_pack` calls return byte-identical
  sheets.
- [X] **T011** (FR-005) Test: an unknown name, and a dead villager's name, each return
  `read_ok` with a roster-naming miss — never an error verdict.
- [X] **T012** (FR-002) Test: empty pack renders `0/24, 24 free` and the empty-pack line; a
  pack with spears renders their remaining-uses list.

**Checkpoint**: the Guardian can see a pack. Independently shippable.

## Phase 4: User Story 2 — the Guardian looks before it speaks (P1)

- [X] **T013** (FR-007) Carry the survival origin onto the turn's dispatch state: `turnDispatch`
  (`internal/guardian/toolcalls.go`) gains a `survival bool` mirrored from `turnOrigin.survival`
  at turn construction in `internal/guardian/turn.go`, beside the existing `night` mirror.
- [X] **T014** (FR-007) `turnDispatch` gains the look-first ledger: a `map[int]bool` of villager
  indices `handleInspectPack` has successfully resolved this turn. Populated ONLY on a
  successful resolution (an honest-miss sheet records nothing).
- [X] **T015** (FR-007, FR-008) Gate `handleVision`: when `d.survival` and the resolved target
  villager is not in the ledger, return `VerdictRejectedGate` with a reason containing the
  literal `inspect_pack` and the villager's name, per contracts §2. Nothing lands, no charge
  moves. The check runs BEFORE `landVision`.
- [X] **T016** (FR-007, FR-008) Gate `handleMiracle` the same way for kinds `give_item` and
  `take_item` only. Other miracle kinds and `send_omen` stay ungated (contracts §2).
- [X] **T017** (SC-002, AC#2) Test: on a survival-origin turn against the Cedar fixture, a
  `send_vision` at Cedar with no prior `inspect_pack(Cedar)` is `rejected_gate`, the reason
  names `inspect_pack`, and no event landed.
- [X] **T018** (SC-002) Test: the same call AFTER `inspect_pack(Cedar)` lands normally — the
  gate is repairable within one turn.
- [X] **T019** (FR-008) Test: the gate is per-villager — `inspect_pack(Hazel)` does not license
  a vision at Cedar.
- [X] **T020** (FR-008) Test: on a NON-survival turn, `send_vision` is ungated and behavior is
  byte-identical to pre-feature.
- [X] **T021** (FR-007) Test: `send_omen` on a survival turn is never gated.

**Checkpoint**: the Guardian cannot speak to a dying villager without opening their pack.

## Phase 5: User Story 3 — the Guardian can empty a pack (P2)

- [X] **T022** (FR-009) Add `take_item` to `internal/tool/registry.go`'s `miracleKinds`,
  `miracleCosts` (cost 1), and `kindToEvent` (`guardian.item_taken`). Do NOT add a new
  `work_miracle` parameter — `villager`/`item`/`qty` already exist.
- [X] **T023** (FR-009, FR-018) Add `ItemTakenPayload` (contracts §3) to `internal/sim` and
  register it in `internal/sim/payloads.go`.
- [X] **T024** (FR-010, FR-011) Implement `applyItemTaken` in `internal/sim/miracles.go`
  following `applyItemGranted`'s structure exactly: full validation → charge spend → mutation,
  with the carried-quantity check rejecting WHOLE and naming the actual count. Transfer into
  `s.pileFor(a.X, a.Y)` reusing the `agent.dropped` arm's rules — spears/axes most-worn-first
  with both slices kept ascending, food preserving `FoodBatch` identity and merging same
  `(Kind, SpoilAt)`. Dispatch the arm from `internal/sim/state.go`.
- [X] **T025** (FR-012) Add the `take_item` arm to `BuildMiracleBatch`
  (`internal/guardian/miracle_batch.go`): the `guardian.item_taken` event plus one recipient
  memory, with `takeMemoryText` beside `grantMemoryText`.
- [X] **T026** (FR-012) Give the operator's IPC `miracle` door parity
  (`internal/ipc/server.go`) so `take_item` reaches the same builder — argument validation
  mirroring `give_item`'s case.
- [X] **T027** (AC#3, SC-003) Test `internal/sim`: Cedar at 24/24 with 20 wood + 4 planks;
  `take_item(wood, 20)` leaves him at 4/24, puts 20 wood in the tile's pile, spends one charge.
  Then `give_item(meals, 4)` lands — and assert it is rejected whole BEFORE the removal.
- [X] **T028** (FR-010) Test: `take_item(wood, 30)` against 20 carried is rejected whole — no
  charge spent, inventory untouched, reason names the actual count.
- [X] **T029** (FR-011, SC-005) Test the conservation invariant: total units in
  (inventory + tile pile) is unchanged by a removal, for a plain kind, for spears, and for
  food.
- [X] **T030** (FR-011) Test: removal merges into an existing pile on the tile (one pile per
  tile); spears leave most-worn-first; food keeps its spoilage batch.
- [X] **T031** (FR-010) Test: dead villager, unknown kind, zero/negative qty each rejected at
  the door with an in-fiction reason and no charge spent.
- [X] **T032** (FR-009) Extend the existing kind-vocabulary drift tests
  (`TestMiracleKindsMirrorTool` and the `GrantKinds` cross-check) so `take_item` is pinned in
  both directions rather than duplicated as a literal.

## Phase 6: User Story 4 — nothing reaches into a pack in silence (P2)

- [X] **T033** (FR-013, SC-004) Test: a landed `take_item` batch contains exactly the
  `guardian.item_taken` event and exactly one `agent.memory_added` for that villager, at
  `SalDream` / `OriginOmen`, first-person and in-fiction.
- [X] **T034** (FR-014) Test: an `inspect_pack` call emits a `cog.tool_call` record with a read
  verdict and lands NO world event; a gated refusal emits a record carrying the refusal reason.
- [X] **T035** (FR-017) `internal/tui/digest.go`: add the `guardian.item_taken` line renderer
  and the subject resolver (contracts §5), mirroring the two `guardian.item_granted` entries.
- [X] **T036** (FR-017) Test: the event renders a plain-language summary and does not fall
  through as an unknown type; the subject column attributes it to the affected villager.
- [X] **T037** (FR-018) Test: a recorded `guardian.item_taken` replays to byte-identical state.

## Phase 7: Cross-cutting polish

- [X] **T038** (FR-016) `internal/tool/derive.go`: add `guardianToolDesc` entries for
  `inspect_pack` (read heading, beside `survey_site`) and `take_item` (miracle-kind gloss).
- [X] **T039** (FR-016) Amend `give_item`'s carry-headroom hint (`derive.go:259`) to name
  `take_item` as the remedy when a villager has no free bulk — closing the world-03 loop at the
  point of guidance, not only at the point of capability.
- [X] **T040** (FR-016) `internal/skin/skin.go`: add `skin.guardian.example_ask.inspect_pack`,
  following the `survey_site` example-ask entry.
- [X] **T041** (FR-015) Update `internal/guardian/stage_test.go`'s enumerated granted rosters to
  include `inspect_pack` and NOT `take_item`, and confirm the ceiling test fails if either is
  wrong.
- [X] **T042** Run `go test ./...` and `go vet ./...`; both clean.
- [X] **T043** (SC-006) Run the TUI design gate if `internal/tui/` changed:
  `node scripts/check-tui-design.mjs --changed`, and amend `docs/design/tui/` in this PR if it
  reports a change.

## Phase 8: Grounding (rides the PR — spec 069)

- [X] **T044** Re-verify and re-pin every `docs/wiki/` note whose pinned sources this branch
  touched (`/grounding-wiki:wiki-update`). The pr gate blocks on `wiki-repin-missing`.
- [X] **T045** Regenerate `docs/player/` if `docs/wiki/` changed (the `player-docs` skill);
  the pr gate blocks on `player-docs-stale`.
- [X] **T046** Run `node scripts/check-merge-drift.mjs pr` from the worktree; exit 0 before
  opening the PR.

## Dependencies

```
T001 → T002 → T003 → T004..T012 (US1)
                       ↓
                     T013..T021 (US2, needs the handler + ledger)
T001 → T022..T032 (US3, independent of US1/US2 except T016's gate arm)
T025 → T033        (memory assertion needs the batch arm)
T024 → T037        (replay assertion needs the reducer arm)
US1+US3 → T038..T041 (guidance/roster describe what exists)
everything → T042..T046
```

## Parallel opportunities

- Phase 3 (US1) and Phase 5 (US3) touch disjoint files apart from `registry.go` and may be
  worked concurrently if two implementers are dispatched; T016 must land after both.
- Within Phase 5, the reducer arm (T024) and the TUI rendering (T035) are independent.

## Implementation strategy

Ship Phase 3 first and confirm it green — it is the MVP and it is safe (pure addition). Add
Phase 4 next, because a read tool nobody calls is exactly the failure this task exists to
correct. Phase 5 last, because it is the only piece that mutates the world and it benefits from
the read tool being available to write its tests against.

## Execution notes

- **Tier**: Opus 5 (`.claude/agents/spec-implementer-opus.md`) — doctrine-adjacent write
  access plus a new structural refusal (plan.md, Constitution Check). Record the tier and this
  justification on TASK-197.
- Do not hand-edit `backlog/`; use the `backlog` CLI.
- Never rebase; freshen by merging main INTO the branch.
- The branch pushes on its first commit and stays pushed.
