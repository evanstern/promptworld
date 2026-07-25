# Tasks: Grounded feedback layer — explain tool, tutor guide, report card

**Input**: Design documents from `/specs/059-grounded-feedback/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/feedback-layer.md, quickstart.md
**Board**: TASK-115 · one branch (`task-115-grounded-feedback`), one PR

## Phase 1: Setup

- [ ] T001 Verify baseline in the task worktree on fresh `origin/main` (must include TASK-121's skin contract + TASK-125's console seam; note whether TASK-127's renderer has merged — research R5 composes it when present, ships the note standalone behind the seam otherwise): build + `go test -race ./...` + `node scripts/check-tui-design.mjs --changed` green

## Phase 2: Foundational — the shared data source

- [ ] T002 Explain fact-sheet composition in internal/tool per research R1: per-topic deterministic sections from registry declarations/guidance derivations, miracle kind/cost table, charge doctrine constants, decision classes, glyph vocabulary; effective-grant/stage-ceiling scoping; topic catalog on miss; ground-truth sweep test over every topic (SC-001)
- [ ] T003 Registry declaration + read-only class: explain declared expressive/empty-events with a read-only marker; loop driver exempts read-only tools from the one-mediated-act budget (internal/guardian turn driver); grant gating through the standard three layers; structural-absence test when ungranted

## Phase 3: User Story 1+2 — explain + tutor-lane doctrine (P1)

- [ ] T004 [US1] Handler wiring in internal/guardian: granted-subset handler returning R1 sheets; multi-round explain-then-act turn tests; unknown-topic catalog result
- [ ] T005 [US2] Neutrality suite (SC-002): charge bank untouched across explain-bearing turns; no world-mutating events; initiative-frame byte-diff test; the spec-052 adversarial battery re-run over the grown prompt surface
- [ ] T006 [US2] Rubric-hygiene sweep in internal/sim (FR-003): no exercise rubric term references tutor-lane telemetry; wired into the existing catalog sweep family

## Phase 4: User Story 3 — tutor guide (P2)

- [ ] T007 [US3] `persona.TutorGuide` compiled constant (content per contract §2) + tutor-preset-scoped composition in the guardian turn assembly (editable zone, after charter/voice/SOULs, before skills/frame — research R3); composition-order tests; non-tutor byte-identity test (SC-003)
- [ ] T008 [US3] Orientation fixture check: canned how-do-I-play/what-does-X-cost prompts compose with guide + explain grounding present (fixture-level, no live model)

## Phase 5: User Story 4 — report card (P2)

- [ ] T009 [US4] New route kind (report-card class) in internal/llm with config plumbing + cheap-chain default routing (metatron_watch precedent, research R4); budget cap; route tests
- [ ] T010 [US4] `guardian.report_card` prose event: whitelist entry, payload {fingerprint, note, citations[]}, reducer (latest-card state + validation), catalog/digest grammar entries; run-end path rides the existing epilogue channel (research R5); replay tests
- [ ] T011 [US4] Producer in internal/guardian (digest-worker pattern): stopping-point triggers (run end, exercise resolution, debounced+activity-gated pause episodes); grading inputs exactly the shared data source; citation validation against the log; silent deterministic degradation without the chain; producer tests with a stubbed chain (SC-004)
- [ ] T012 [US4] Client rendering: console card seam composition (stored note + TASK-127 checklist when available) with the unseen-badge between stopping points; postmortem embedding; render tests across {note-only, checklist-only, both, degraded, none}
- [ ] T013 [US4] Skin tokens for card labels/guide framing/example asks per contract §4 (default table + doc twin + completeness test — internal/skin)

## Phase 6: User Story 5 — the ? guardian section (P3)

- [ ] T014 [US5] D9 section in internal/tui/help.go per research R6: stage identity + effective verbs + one example ask per verb (skin nouns); byte-identity-per-status test (SC-005); spec-045 content-contract amendment recorded on overlays/help.md

## Phase 7: Polish & Cross-Cutting Concerns

- [ ] T015 [P] Design-page amendments: overlays/help.md (D9 section), pages/guardian-console.md (card production real; seam note), patterns/skin-tokens.md (token twin); re-pin every touched page to the final code commit
- [ ] T016 Run gates: `go test -race ./...`, `node scripts/check-tui-design.mjs --changed`, `node scripts/check-merge-drift.mjs pr`, gofmt/vet clean
- [ ] T017 Pre-PR: rebase onto fresh `origin/main` (expect guardian/tui conflicts with Lane siblings), re-run all gates post-rebase

## Dependencies & Execution Order

- Phase 2: T002 → T003.
- US1/2: T004 needs T003; T005 after T004; T006 [P] anytime after T002.
- US3: T007 after T003; T008 after T007.
- US4: T009 [P]; T010 after T002; T011 needs T009+T010; T012 needs T011 (+127's renderer if merged); T013 [P] with T012.
- US5: T014 [P] after T013.
- Polish: T015 → T016 → T017.

## Implementation Strategy

Data source first (explain), doctrine proven second (neutrality), then the
guide, then the push layer, then the deterministic floor. One worktree, one
PR; PR body calls out the tutor-lane invariants and the stored-never-regraded
card contract.
