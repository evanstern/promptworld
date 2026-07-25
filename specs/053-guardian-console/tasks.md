# Tasks: Guardian console page + systems-tab telemetry split

**Input**: Design documents from `/specs/053-guardian-console/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/console-and-systems.md, quickstart.md
**Board**: TASK-125 · one branch (`task-125-guardian-console`), one PR

## Phase 1: Setup

- [x] T001 Verify baseline in the task worktree on a fresh `origin/main` fork (post-Lane-1 merges): `go build ./...`, `go test -race ./...`, `node scripts/check-tui-design.mjs --changed` green; note which of TASK-121/124/126 have merged (affects fiction-string handling per contract §3 rebase note)

## Phase 2: Foundational

- [x] T002 Systems dock tab: extend the dock tab enum/paneNames (internal/tui/tui.go:48) with `systems`, key `5`, through the existing dockTabContent dispatch; per-tab state/zoom/narrow inherit; tab-grammar regression tests (2/3/4 unchanged, 5 added) in internal/tui/tui_test.go
- [x] T003 Console page state per data-model.md: `console` flag + return-target snapshot + scroll + one-shot notice on Model (internal/tui/tui.go); `G` open/toggle, `1`/`esc` close-restore; esc-release ordering gains the console layer; focus-contract tests extended (internal/tui/focus_test.go) — minibuffer-focused `G`/`e`/`5` type into the buffer

## Phase 3: User Story 2 — Telemetry moves out (P1) 🎯 merge-risk first

**Goal**: systems tab renders 100% of relocated telemetry; guardian tab fiction-only.

**Independent Test**: quickstart step 3; `go test -race ./internal/tui -run Systems`.

- [x] T004 [US2] Systems content renderer in internal/tui/views.go composing the existing llmProviderLines + spend/wallet lines + horizonLines/horizonRow (relocation, not rewrite); no-LLM world renders honest absence (systems.md structure §1–3)
- [x] T005 [US2] Remove telemetry from the guardian tab renderer (views.go:1436-1560 vicinity), keeping pane header, transcript, standing orders, provenance lines; split render tests over LLM/no-LLM fixtures asserting zero telemetry rows on guardian + full set on systems (SC-002)

## Phase 4: User Story 1 — The console page (P1)

**Goal**: `G` page with document turns + composer.

**Independent Test**: quickstart steps 1–2; `-run Console`.

- [x] T006 [US1] consoleView in internal/tui/views.go: header line (guardian pane-header data), document-style turn blocks over Model.transcript (labeled `you`/epithet + timestamp-when-present, blank-line separated, width-wrapped), special rows ⚡/👁/⏲/» inline (research R4); tail-anchored scrollback `J`/`K` reset-on-close; View() branches to the page (research R1)
- [x] T007 [US1] Composer pairing: the standard minibufferView beneath the stream (no new input widget); busy/focused/dormant states byte-identical; composer tests reusing the minibuffer state harness (SC-003)
- [x] T008 [US1] Card seam per research R6: consoleCard interface + always-empty composition slot between stream and read surface; position pinned by a test fake; seam symbol documented for TASK-127/115
- [x] T009 [US1] Footer hints: console footer per contract §2.7; global footer advertises `G`; help overlay content gains G/5 rows (internal/tui/help.go per overlays/help.md content rules)

## Phase 5: User Story 3 — Charter/skills read surface + $EDITOR (P2)

**Goal**: provenance/lock surface + $EDITOR round-trip.

**Independent Test**: quickstart steps 4–5; `-run Editor|ReadSurface`.

- [x] T010 [US3] Read surface from status fields only (charter_default/locked/preset, skills list/locked, stage — research R5): bordered sub-panel with provenance + binding + lock notices naming unlocking stages; fixture tests across {default, player-authored, tutor-locked stage-1, stage-2, stage-3+} (SC-004)
- [x] T011 [US3] $EDITOR handoff (research R2): `e` → tea.ExecProcess on the world's charter.md with pre/post content-hash; changed → one-shot "charter changed — next turn binds it"; unchanged → nothing; unset/failed $EDITOR → honest notice; scripted fake-editor round-trip tests

## Phase 6: Polish & Cross-Cutting Concerns

- [x] T012 [P] Design-page amendments (research R7): guardian-console.md + systems.md → shipped w/ real symbols (card row stays unbuilt-wave-3/4 + seam note); guardian.md, dock.md, keymap.md (G/5/e/scroll + parity gaps + footer hints), solo-views.md, overlays/help.md re-verified; ALL touched pages re-pinned
- [x] T013 Run gates: `go test -race ./...`, `node scripts/check-tui-design.mjs --changed`, gofmt/vet clean
- [x] T014 Pre-PR: rebase onto fresh `origin/main` (expect conflicts with Lane-1 siblings in views.go/tui.go/keymap.md — take main's side for anything not intentionally changed; if TASK-121 merged, swap any new fiction literals for skin-token lookups per its contract); re-run all gates post-rebase

## Dependencies & Execution Order

- Phase 2: T002 and T003 independent [P].
- US2 (T004→T005) needs T002; ships first — it's the highest-conflict slice.
- US1: T006 needs T003; T007/T008/T009 after T006.
- US3: T010 after T006; T011 after T003.
- Polish: T012 after behavior final → T013 → T014.

## Implementation Strategy

Split first (US2 — the merge-risk mover), then the page (US1), then the read
surface/$EDITOR (US3). One worktree, one PR; PR body calls out the card-seam
scope ruling (D5) for reviewers.
