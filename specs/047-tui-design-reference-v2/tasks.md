# Tasks: TUI Design Reference v2 — the Living UI Authority

**Input**: Design documents from `/specs/047-tui-design-reference-v2/`

**Prerequisites**: plan.md, spec.md, research.md (the three rulings live there — author
them, don't re-derive), data-model.md, contracts/ (control-table.md,
frontmatter-and-pins.md, check-script.md), quickstart.md

**Tests**: no test framework tasks — validation is the check script against seeded
violations (quickstart.md §2) plus the corpus acceptance sweeps (quickstart.md §4).

**Organization**: grouped by user story. All paths are repo-relative; implementation
happens in the `.worktrees/task-123` worktree on branch
`task-123-tui-design-reference-v2` (one task, one PR).

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

**Purpose**: branch/worktree and taxonomy skeleton.

- [ ] T001 Create worktree `.worktrees/task-123` on branch
      `task-123-tui-design-reference-v2` from fresh `origin/main`; create
      `docs/design/tui/overlays/` directory.

---

## Phase 2: Foundational (blocking prerequisites)

**Purpose**: conventions every page references — must exist before any page is
authored or reconciled.

- [ ] T002 Author `docs/design/tui/patterns/skin-tokens.md` per research.md R7:
      `{{skin.<domain>.<name>}}` mockup placeholder convention, control-table
      skin-token column semantics, token index table (token · default-skin/angel
      value · used-by page), explicit deferral of the runtime lookup contract to
      TASK-121. Frontmatter per contracts/frontmatter-and-pins.md
      (`status: specified`).
- [ ] T003 Rewrite `docs/design/tui/INDEX.md`: authority statement, the v2 taxonomy
      map (all files from plan.md Project Structure), the gate rules (same-PR
      amendment — extends old rule 4 to "any change"; run
      `node scripts/check-tui-design.mjs --changed` before any PR touching
      `internal/tui/`), and the FR-020 audience ruling (plain-language default, raw
      registry values behind a debug/inspector toggle) as a corpus-wide convention.
      Preserve the TASK-34 decision record as a history section.

---

## Phase 3: User Story 1 — reconcile shipped reality + taxonomy (P1)

**Goal**: every surface shipped by specs 013–046 documented where an implementer
would look; dock split per-tab; help extracted. Ground against
`docs/wiki/tui-client.md` (primary) + `internal/tui/*.go` (verification), per
research.md R1; staleness sweep set: specs 015, 018, 020, 021, 024, 028/031/033,
029, 034, 035, 037, 039, 044, 045, 046.

**Independent test**: spec.md US1 acceptance scenarios — pick any shipped element,
reach its accurate owning page in ≤2 hops from anatomy.md.

- [ ] T004 [US1] Rewrite `docs/design/tui/panels/dock.md` as tab-container chrome
      only (tab row, badges incl. `metatron •` dot pattern, tab-switch keys,
      solo-zoom seam); move all tab content out (destinations: T005–T007). Add
      frontmatter + pin (`status: shipped`), canonical control table.
- [ ] T005 [P] [US1] Author `docs/design/tui/panels/guardian.md` (fiction-layer tab
      content split from dock.md, D10): transcript/replies, ⚡ standing-order lines,
      ⏲ pause/start lines, `👁 standing orders (n)` block (spec 029), instruction
      surface (spec 021), miracle feedback (spec 016) — fiction strings as skin
      tokens (T002 conventions). Mockup + control table + pin.
- [ ] T006 [P] [US1] Author `docs/design/tui/panels/systems.md` (telemetry tab —
      never skinned, D10): provider table (spec 024) with health-condition rows
      (spec 034), horizon block (spec 037), throttle/debt/spend surfaces (specs
      028/031/033), calibration UX (spec 035). Mockup + control table + pin.
- [ ] T007 [P] [US1] Author `docs/design/tui/panels/villagers.md` (villagers tab,
      specs 015/020): roster → detail → decision-trace drill-down, verdict rows,
      states. Mockup + control table + pin.
- [ ] T008 [P] [US1] Update `docs/design/tui/panels/chronicle.md` +
      `docs/design/tui/patterns/chronicle-grammar.md`: verify against shipped digest
      grammar (spec 018), suppression/remedy rows, verdict glossary; document the
      reserved `⏎` jump-to-source seam as `specified` for Wave 2 (D3). Canonical
      control table on the panel page; pins on both.
- [ ] T009 [P] [US1] Update `docs/design/tui/panels/map.md` and
      `docs/design/tui/panels/minibuffer.md`: verify against shipped reality (night
      dimming, priority carve-outs; minibuffer states + truncation behavior), note
      Wave-5 condition overlays as `specified` stub rows; canonical control tables +
      pins.
- [ ] T010 [P] [US1] Update `docs/design/tui/pages/home.md` and
      `docs/design/tui/pages/solo-views.md`: header segments reconciled (stage
      segment spec 046, `[llm: …]` badge spec 034, `[suppressed: …]` badge spec 037,
      `[degraded]`, ENDED token spec 044, speed posture spec 039); new chrome rows
      composed (references to T014–T016 pages); solo-views gains the new
      tabs/pages. Pins on both.
- [ ] T011 [US1] Extract the help overlay from `docs/design/tui/patterns/keymap.md`
      into `docs/design/tui/overlays/help.md`: shipped spec-045 sections (tiered
      keys, screen walkthrough, lessons seam) reconciled against
      `specs/045-tui-help-overlay/contracts/help-content.md` and
      `internal/tui/help.go`; leave keymap.md a pure printable card (parity rule
      added in T021). Canonical control table + pin. (Guardian section lands in
      T017; byte-identity classification in T023.)
- [ ] T012 [US1] Author `docs/design/tui/anatomy.md`: every visible region, strip,
      badge, chrome row → owning file (data-model.md anatomy invariants — both
      directions complete), with stage-default visibility and fold-behavior
      references. Covers all files from T004–T020.

**Checkpoint**: US1 acceptance scenarios pass; anatomy maps the shipped client
completely.

---

## Phase 4: User Story 2 — ten new-surface pages spec-before-build (P1)

**Goal**: all ten reorientation surfaces authored (`status: specified`): mockup +
canonical control table + stage defaults + linear-stream projection (D1) each.
Encode governing decisions per spec FR-012/013/014 and rulings FR-018/019 — cite
reorient decision ids in `introduced-by`.

**Independent test**: spec.md US2 acceptance scenarios; quickstart.md SC-004 sweep.

- [ ] T013 [P] [US2] Author `docs/design/tui/pages/guardian-console.md` (decisions
      1/2, D5): full-height page, document-style turns, composer, charter/skills
      read surface with binding status + `$EDITOR` handoff + "charter changed —
      next turn binds it" confirmation, report-card cards at natural stopping
      points (run end / pause / exercise resolution; badges between).
- [ ] T014 [P] [US2] Author `docs/design/tui/panels/lesson-row.md` (decision 5):
      one active lesson, ≤2 lines, dwell-until-done/dismissed, UI-pointer field,
      anti-spam (one active · spacing · opportunity decay), per-user seen state
      (D8, unlocks.json precedent), pull-path suffix on every lesson string,
      prompting-lesson tier in the trigger taxonomy (first rejected tool call,
      first custom charter observed, first fuzzy order), stage defaults (on at
      1–2; badge+overlay at 3+/pre-ladder), narrow + fold behavior per research.md
      R2/R3.
- [ ] T015 [P] [US2] Author `docs/design/tui/panels/guardian-strip.md` (decision
      7): budget line above minibuffer — charge bank, regen, standing-order count,
      faith (TASK-118, marked pending) — always visible, all stages; fold-last rule
      (relocates into dormant minibuffer line, research.md R2); carried in narrow.
- [ ] T016 [P] [US2] Author `docs/design/tui/panels/villager-strip.md` (D12):
      one-row colonist-bar strip under the header, widescreen default-on, folds to
      header count badge (research.md R2), not carried in narrow (R3).
- [ ] T017 [P] [US2] Author `docs/design/tui/panels/exercise.md` (D11, D4): scenario
      framing line, live event-derived rubric gauges (decision-trace projection),
      per-exercise incident visibility VOCABULARY (forecast at stages 1–2, fog from
      3 — a vocabulary, not a boolean), attach-time briefing, pass/fail state,
      scenario-cadence narration trigger, ceremony-trigger linkage.
- [ ] T018 [US2] Add the guardian section to `docs/design/tui/overlays/help.md`
      (D9): static-per-stage, model-free — stage identity/concept, granted verbs,
      one example ask per verb, renderable from `stagesLadder` + stage ceiling with
      nil-status fallback to the pre-ladder variant (research.md R4); note the
      deliberate spec-045 content-contract amendment; badge deep-link (`?` opens
      pre-focused on the active badge's row) as a retained layer-2 row. (Depends on
      T011.)
- [ ] T019 [P] [US2] Author `docs/design/tui/overlays/ceremony.md` (decision 6):
      takeover on `curriculum.stage_unlocked` (attached clients), dismissable,
      non-stacking with stated precedence vs postmortem, replayable from `?` and
      `stages` as an EXPLICIT acceptance criterion, unlock-attribution voice = the
      player's authorship, skin-resolved (D6), score voice = narrated skin-resolved
      chapter + rubric checklist with the instrument authoritative (FR-019 ruling),
      `q`-detach blessed-stopping-point shape (D13), interrupt-policy watch item
      recorded (reopening signal: ceremony fatigue / mid-crisis seizure evidence).
- [ ] T020 [P] [US2] Author `docs/design/tui/overlays/postmortem.md` (decision 6):
      takeover on `run.ended`, dismissable, non-stacking (precedence per T019),
      replayable from the morgue as an EXPLICIT acceptance criterion, ambient
      (unscored) contents = morgue evidence only in the no-blame register — report
      card only in scored/scenario runs (FR-018 ruling), report-card rendering
      shared with guardian-console cards (D5).
- [ ] T021 [P] [US2] Author `docs/design/tui/patterns/stage-defaults.md` (decision
      3): stage-resolved visibility DEFAULTS table for every panel/strip/row
      (stages 1–4, pre-ladder = everything), the everything-reachable-always rule
      (`?` + solo views), capability locks stay angel-only (spec 046 doctrine
      untouched), how defaults compose with the fold order (research.md R2).

**Checkpoint**: all ten surfaces authored; SC-004 sweep passes.

---

## Phase 5: User Story 5 — the Wave 0 rulings recorded (P2)

**Goal**: research.md R2/R3/R4 authored into their homes. (Runs after US1/US2 pages
exist since the rulings reference them.)

**Independent test**: quickstart.md SC-006 sweep.

- [ ] T022 [US5] Rewrite `docs/design/tui/patterns/layout.md`: keep breakpoint (112
      cols) + column budget; re-derive the row budget per research.md R2 (9-row
      stage-1–2 chrome stack, per-stage variants via stage-defaults), rule the fold
      order (legend → villager strip → lesson row → guardian strip-relocation) with
      `bodyMin = 10` and the floor layout; add the narrow-fallback chrome rules
      (R3). Preserve the style-token table and composition notes; pin.
- [ ] T023 [US5] Add the byte-identity classification to
      `docs/design/tui/overlays/help.md` per research.md R4: the section
      classification table (byte-identical / stage-keyed model-free /
      status-derived) and the restated no-LLM floor guarantee. (Depends on
      T011+T018.)

---

## Phase 6: User Story 3 — uniform control tables + parity doctrine (P2)

**Goal**: corpus-wide table conformance and the decision-8 doctrine recorded.

**Independent test**: spec.md US3 acceptance scenarios; SC-002 sweep.

- [ ] T024 [US3] Add the input-parity doctrine to
      `docs/design/tui/patterns/keymap.md` (decision 8): every action reachable by
      keyboard AND mouse, keyboard primary and complete, incremental rollout with
      honest per-page "parity rollout" notes (contracts/control-table.md rule 4);
      keep the one-page printable card format; document the mnemonic-key and
      reserved-seam binding rules; pin.
- [ ] T025 [US3] Corpus conformance pass: every `panels/*` and `overlays/*` page has
      exactly one canonical control table (contracts/control-table.md header,
      byte-exact), keys+mouse column filled per rule 4, `introduced-by` filled,
      zero bare fiction literals outside `patterns/skin-tokens.md`'s default-value
      column (verify: `grep -rn "Metatron" docs/design/tui/ --include="*.md"`).
      Also update `docs/design/tui/patterns/focus-contract.md`: new chrome rows are
      display-only (minibuffer stays the only text input; console composer defined
      in T013 as a page-level surface honoring the contract); pin all touched files.

---

## Phase 7: User Story 4 — freshness mechanization (P2)

**Goal**: pins everywhere, the check script, the gate wired. (Last: it gates the
finished corpus.)

**Independent test**: quickstart.md §§1–3 (clean pass + all five seeded violation
classes + doc-only range passes).

- [ ] T026 [US4] Frontmatter/pin conformance pass over every `.md` under
      `docs/design/tui/` per contracts/frontmatter-and-pins.md: `title`, `class`
      matching directory, `status`, `verified_against` = a commit containing the
      branch's reconciliation state, optional `sources` on shipped pages.
- [ ] T027 [US4] Implement `scripts/check-tui-design.mjs` per
      contracts/check-script.md: Node ≥18 ESM, zero deps, read-only; checks
      file-set, pins, control-tables, anatomy, and `--changed [range]` same-PR
      gate (default `origin/main...HEAD`); `--json` mode; exit codes 0/1/2;
      actionable violation messages naming files.
- [ ] T028 [US4] Validate the script per quickstart.md §§1–3: clean tree passes;
      each of the five seeded violation classes fails with the contracted message;
      doc-only range passes `--changed`. Record the transcript in the PR
      description / task notes.
- [ ] T029 [US4] Wire the gate: add the run-the-check rule to `CLAUDE.md` (one line
      beside the player-docs freshness rule: run
      `node scripts/check-tui-design.mjs --changed` before any PR touching
      `internal/tui/`); confirm INDEX.md's gate-rules section (T003) matches the
      shipped script exactly.

---

## Phase 8: Polish & cross-cutting

- [ ] T030 Fix `docs/wiki/tui-client.md` prose references to renamed/split design
      files (lines ~83/295/297 cite `docs/design/tui/` paths — prose fix only, no
      pin re-verification; constitution IV watch item from plan.md).
- [ ] T031 Run the full acceptance sweep per quickstart.md §4 (SC-001, SC-002,
      SC-005, SC-006 manual sweeps) and §5 regression (`go test ./...`,
      player-docs `check-freshness.mjs --check`); record results in task notes.
- [ ] T032 Open the single PR from `.worktrees/task-123`
      (`task-123-tui-design-reference-v2` → main) with the validation transcript;
      after review/merge: `spec-bridge:sync`, tick board ACs, worktree cleanup.

---

## Dependencies

- Phase 1 → Phase 2 → everything else.
- T004 (dock split) before T005–T007 conceptually (content destinations), but
  T005–T007 are parallel once the split boundaries are fixed by plan.md — safe to
  run [P] with T004.
- T011 (help extraction) → T018 (guardian section) → T023 (classification).
- T012 (anatomy) after T004–T011 and referencing T013–T021 (author last in US1 or
  amend after Phase 4 — final state must map every file).
- Phase 5 (rulings) after Phases 3–4 (references their pages).
- Phase 6 (conformance) after Phases 3–5 (sweeps the finished corpus).
- Phase 7 after Phase 6 (pins + script gate the finished corpus); T027 ∥ T026;
  T028 after T026+T027; T029 after T027.
- Phase 8 last; T032 is the terminal task.

## Parallel opportunities

- US1: T005, T006, T007, T008, T009, T010 all [P] (different files).
- US2: T013–T017, T019–T021 all [P]; T018 waits on T011.
- US4: T026 ∥ T027.

## Implementation strategy

MVP = Phase 1–3 (US1): the reference is true again — the authority claim holds for
shipped reality even before new pages land. Then US2 (the reorientation deliverable),
US5 rulings, US3 conformance, US4 mechanization, polish. Single PR delivers all
phases (one task, one PR); phases are commit boundaries on the task branch, not PR
boundaries.
