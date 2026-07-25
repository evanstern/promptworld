# Tasks: Skinnable guardian persona — de-theme the angel fiction, persona as data

**Input**: Design documents from `/specs/052-skinnable-guardian/`
**Prerequisites**: plan.md, research.md (R8 inventory is normative), data-model.md, contracts/skin-contract.md, quickstart.md
**Board**: TASK-121 · one branch (`task-121-skinnable-guardian`), one PR

**Organization**: US1 (contract) is the Lane-3 unblock and MVP; US2 (default
de-theme sweep) is co-P1; US3 (custom skins) proves persona-as-data; US4
(rename + freeze) is compiler-safe churn, last.

## Phase 1: Setup

- [x] T001 Verify baseline in the task worktree: `go build ./...`, `go test -race ./...`, `node scripts/check-tui-design.mjs --changed` green on a fresh `origin/main` fork; snapshot a pre-feature test world fixture for SC-003

## Phase 2: Foundational — the skin substrate

- [x] T002 Grow internal/skin per data-model.md: `Skin` value type (identity fields, Strings, Stages, Voice) with field-wise validation + default fallback; compiled default table (contract §3 normative rows); token resolution (override → default → token-path-itself); token-completeness test scaffold
- [x] T003 `skin.Load(worldDir)` reading `<world>/skin.json` with the capabilities.json fallback discipline (missing → default silent; malformed → default + one notice; unknown token keys ignored + notice; research R1/R4); table-driven loader tests incl. hostile identity fields (length/control-char clamps)
- [x] T004 [P] Stage identities gain the skin dimension: existing `skin.Stage`/`StageName` accessors resolve through the loaded skin with default fallback; all current call sites (cmd/promptworld/commands.go:136,143,937; internal/tui/tui.go:234, digest.go:1197; internal/metatron/charter.go:119,135) keep compiling and get world-skin awareness where a skin is in scope
- [x] T005 Daemon wiring: `SetSkin` on the guardian agent following the SetBundles/SetStage boot-frozen discipline; status surface gains additive omitempty skin fields (contract §7) in internal/ipc + the status structs; old-daemon-absent-fields → default skin in clients

## Phase 3: User Story 1 — The skin-token contract exists (P1) 🎯 Lane-3 unblock

**Goal**: published, tested lookup + table + doc twin.

**Independent Test**: `go test -race ./internal/skin -v`; contract §3 tokens resolve; skin-tokens.md carries the runtime section.

- [x] T006 [US1] Token-completeness test: every token consumed anywhere in the repo exists in the default table; a missing token renders its own path and fails the test (spec US1 AS-3)
- [x] T007 [US1] Amend docs/design/tui/patterns/skin-tokens.md: add the runtime-contract section (resolution order, fallback, skin.json format pointer, downstream obligations per contract §4), promote the interim token index into the default table's doc twin, note the page's "MUST be adopted or amended by TASK-121's PR" requirement as satisfied; re-pin `verified_against`

## Phase 4: User Story 2 — The default experience is de-themed (P1)

**Goal**: zero denylist fiction in the default experience; every literal → lookup.

**Independent Test**: denylist sweep test green; manual attach per quickstart step 1.

- [x] T008 [US2] Fiction-denylist sweep test (SC-001/002): asserts rendered TUI surfaces, CLI output, and composed prompts in the default skin contain no Metatron/angel/miracle-display/divine/heaven/scripture; explicit allowlist for frozen serialized identifiers (research R4) and history files; runs in `go test ./...`
- [x] T009 [US2] TUI sweep via status-carried skin facts (research R8 TUI list): internal/tui/help.go (6 sites), views.go (11 sites incl. tab label:376, footer:254, pane header:1511, transcript labels, busy/unreachable/exhausted, minibuffer placeholder:1744), tui.go (paneNames:48, transcript prefix:460, grant summary "workings":250-253) — all through the lookup; no bare literal survives
- [x] T010 [US2] Chronicle grammar sweep: digest.go subject lines (933-1074, 1197) render the skin name; grant/working vocabulary; Type-column family alias per FR-013 (grammar.go:173-198 leading segment via `skin.guardian.family_label`; detail pane + raw fallback stay verbatim) — tests for aliased column + raw pane
- [x] T011 [US2] CLI sweep: canonical `guardian` and `work` subcommands with hidden functional aliases `metatron`/`miracle` (main.go:33-42,102-105; miracle.go; commands.go:396-450); stagesLadder prose (stages.go:41-67) re-worded where fiction-bearing; usage/help shows canonical only; alias tests
- [x] T012 [US2] Prompt-constant sweep: persona.DefaultCharter rewritten guardian-voiced (genesis seeds skin name); fixed frame + miracle doctrine line (turn.go:869-890) de-themed with validated name substitution; digest keeper (metatron/digest.go:192-194); watch confirmer (orders.go:387-390); trigger moment + ResultForModel "working" (orders.go:640, toolcalls.go:185); tool glosses (tool/derive.go:202-222); soul genesis header (metatron.go:212); morgue line (scribe/morgue.go:342). Recorded-at-emission text NOT touched (research R8 frozen list; FR-005)
- [x] T013 [US2] README sweep (lines 30,40,62,63 + reword around frozen llm.json kinds at 98)

## Phase 5: User Story 3 — A custom skin is a per-world data bundle (P2)

**Goal**: skin.json re-themes a world; invariants unbreakable; mechanics identical.

**Independent Test**: quickstart steps 3–4; equivalence + adversarial suites.

- [x] T014 [US3] Persona-voice composition: skin Voice inserts at the SOUL-fragment seam in turnSystemPrompt (turn.go:860-900) with the bundle-SOUL cap/validation; fixed frame remains appended last on every path; composition-order unit tests
- [x] T015 [US3] Extend the adversarial battery (metatron_test.go) with hostile-skin fixtures: instruction-bearing voice, hostile name/epithet, oversized fields — invariants hold, clamps apply (SC-005)
- [x] T016 [US3] Deterministic two-skin mechanics-equivalence test (SC-004): same seed + same scripted tool calls under default and raven skins → identical event-type sequences, charge arithmetic, reducer outcomes (FR-005/006)
- [x] T017 [US3] Ship examples/skins/raven.json + examples/skins/README.md documenting the format (FR-014, original folk identity per research R7); loader round-trip test on the example

## Phase 6: User Story 4 — The internals stop lying (P3)

**Goal**: Go rename + frozen-vocabulary annotations; full compat.

**Independent Test**: quickstart step 5; SC-003 compat suite.

- [x] T018 [US4] Freeze annotations: comment every frozen serialized constant at its definition site (research R4 normative list — event types, IPC methods, JSON tags, llm.json kinds, tool ids, paths, correlation prefixes, `"omen"` origin) stating the freeze and citing spec 052 ruling 2
- [x] T019 [US4] Package rename internal/metatron → internal/guardian + unserialized Go identifier sweep (paneMetatron, familyMetatron const name w/ frozen string, metatronVerdictRow, sim Go names w/ frozen tags, etc.); serialized strings byte-identical (guarded by T018's annotations + T020's compat tests); mechanical commits separated from behavior commits
- [x] T020 [US4] Compat suite (SC-003): pre-feature world fixture opens + replays byte-identically; old capabilities.json/llm.json load; `promptworld metatron`/`miracle` aliases work; IPC `metatron_chat`/`metatron_status` respond

## Phase 7: Polish & Cross-Cutting Concerns

- [x] T021 [P] Design-corpus amendments beyond T007: any page whose token index/labels this feature's final table changes; re-pin `verified_against` on every touched page
- [x] T022 Run gates in the worktree: `go test -race ./...`, `node scripts/check-tui-design.mjs --changed`, gofmt/vet clean
- [x] T023 Pre-PR: rebase onto fresh `origin/main` (expect TUI conflicts with merged Lane-1 siblings — take main's side for anything not intentionally changed, re-run all gates post-rebase)

## Dependencies & Execution Order

- Phase 2: T002 → T003 → T005; T004 after T002 [P with T003].
- US1: T006 after T002; T007 after T006 (table stable enough to document — update again in T021 if the sweep grows it).
- US2: T008 first (test harness, initially failing), then T009/T010/T011/T012/T013 in any order [P across files] until T008 passes.
- US3: T014 → T015; T016 after T003+T014; T017 after T003.
- US4: T018 before T019 (annotations guard the rename); T020 last.
- Polish: T021 → T022 → T023.

## Implementation Strategy

Contract first (Phases 1–3) — that is the Lane-3 unblock even before the
sweep completes. Sweep test (T008) drives US2 to a mechanically-verified
finish. Rename (US4) rides late, mechanically, in separated commits. One
worktree, one PR; the PR body calls out the frozen-vocabulary ruling for
reviewers.
