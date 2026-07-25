# Implementation Plan: TUI Design Reference v2 — the Living UI Authority

**Branch**: `047-tui-design-reference-v2` | **Date**: 2026-07-25 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/047-tui-design-reference-v2/spec.md`

## Summary

Rebuild `docs/design/tui/` as the living page-by-page, control-by-control UI authority
(reorientation 2026-07-25, decision 4 / Wave 0): reconcile with shipped reality (specs
013–046), adopt the four-class taxonomy (`pages/` · `panels/` · `overlays/` · `patterns/`
+ `anatomy.md`), put one canonical control table on every panel/overlay page, author the
ten new-surface pages spec-before-build, record the three Wave 0 rulings, and mechanize
freshness with `verified_against` pins + a zero-dependency Node check script + a same-PR
amendment gate. Documentation-first: no Go code changes; the deliverable is the reference
corpus plus one script and the gate wiring (CLAUDE.md + INDEX.md rules).

## Technical Context

**Language/Version**: Markdown (reference corpus); Node.js ≥ 18 ESM for the check script
(zero npm dependencies — stdlib `fs`/`path`/`child_process` only, per the TASK-82
`check-freshness.mjs` precedent). Go code untouched.

**Primary Dependencies**: none new. Grounding inputs: `docs/wiki/tui-client.md` (387
lines, pinned, the de-facto accurate reference), `docs/wiki/curriculum-ladder.md`,
`docs/wiki/metatron.md`, `docs/wiki/chronicle.md`, specs 013–046 (esp. 044 morgue, 045
help overlay + `contracts/help-content.md`, 046 curriculum), `internal/tui/*.go` (17
files) for verification reads, `docs/design/reorient-2026-07-25-ui.md` (fixed
constraints).

**Storage**: files only — `docs/design/tui/**/*.md` + `scripts/check-tui-design.mjs`.

**Testing**: the check script validated against seeded violations (each violation class
must fail with an actionable message — SC-003); corpus audited against SC-001/002/005/006
acceptance sweeps (documented in quickstart.md). `go test ./...` must stay green
(untouched code).

**Target Platform**: repo tooling (macOS/Linux dev machines), Node ≥ 18.

**Project Type**: documentation corpus + repo tooling (single project).

**Performance Goals**: check script completes in < 5s on the full corpus (it reads ~30
Markdown files and one `git diff --name-only` range).

**Constraints**: no CI exists in this repo — enforcement follows the TASK-82 precedent:
a documented gate script run at defined moments (PR authoring, session close), named in
CLAUDE.md and INDEX.md. The script must be read-only (never rewrites pins itself).
Fiction strings in the corpus must appear only as skin tokens (D2) even though the
runtime token contract is TASK-121's deliverable.

**Scale/Scope**: ~25 reference files (9 rewrites/updates, ~15 new pages, 1 deletion by
split), 1 check script, CLAUDE.md gate wiring, board/bridge sync. Ten new-surface pages
each need mockup + control table + stage defaults + linear-stream projection.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|---|---|---|
| I. Artifact-Grounded Action | PASS | Spec 047 committed (07a11bb); operator rulings encoded in spec Clarifications; this plan + research/data-model/contracts/quickstart are the decision artifacts; board TASK-123 In Progress with plan recorded. |
| II. One Task, One PR | PASS | TASK-123 ↔ one branch (`task-123-tui-design-reference-v2` in `.worktrees/task-123`) ↔ one PR. Spec docs commit to main at root (project convention); the deliverable corpus + script ride the task branch. |
| III. Gates Over Assertions | PASS | Feature *builds* a gate rather than bypassing one; spec-bridge gate governs TASK-123 status; check script is read-only and violations are fixed by producing artifacts, never by editing derived state. |
| IV. Grounding Freshness | PASS (watch) | No `internal/` code changes, so no wiki source pins are touched. Watch item: `docs/wiki/tui-client.md` cites `docs/design/tui/` paths in prose (lines 83, 295, 297) — if the split renames a cited file, fix the prose reference in the same PR (not a pin re-verification). |
| V. Model-Tiered Workflow | PASS | Fable 5 produced spec/plan/rulings (the judgment concentrates here — the three rulings are derived in research.md, not left to the implementer). Implementation = doc reconciliation + authoring + one Node script → **Sonnet tier** per rubric ("doc reconciliation", single-surface tooling); escalation to Opus 4.8 only if a slice fails gates. Tier choice + justification recorded on TASK-123. |

Post-Phase-1 re-check: PASS — design artifacts introduce no new projects, no code
architecture, no violations. Complexity Tracking stays empty.

## Project Structure

### Documentation (this feature)

```text
specs/047-tui-design-reference-v2/
├── spec.md              # Feature spec (committed 07a11bb)
├── plan.md              # This file
├── research.md          # Phase 0: reconciliation inventory + the three rulings derived
├── data-model.md        # Phase 1: page/frontmatter/control-table/report entities
├── quickstart.md        # Phase 1: validation runbook (script + acceptance sweeps)
├── contracts/
│   ├── control-table.md         # canonical column set + per-column value grammar
│   ├── frontmatter-and-pins.md  # page frontmatter + verified_against pin format
│   └── check-script.md          # CLI contract: flags, exit codes, violation classes
├── checklists/requirements.md   # spec quality checklist (passing)
└── tasks.md             # Phase 2 (/speckit-tasks output)
```

### Source Code (repository root)

```text
docs/design/tui/                     # THE deliverable — v2 taxonomy
├── INDEX.md                         # rewrite: authority statement + gate rules + map
├── anatomy.md                       # NEW: region index — every visible element → owning file
├── pages/
│   ├── home.md                      # update: stage-shaped defaults, new chrome rows
│   ├── guardian-console.md          # NEW (decision 1/2, D5)
│   └── solo-views.md                # update: new tabs/pages reachable solo; narrow rules
├── panels/
│   ├── map.md                       # update: verify + pin (condition overlays noted as Wave 5)
│   ├── chronicle.md                 # update: verify + pin; jump-to-source seam (D3)
│   ├── dock.md                      # rewrite: tab container chrome ONLY (tabs, badges)
│   ├── guardian.md                  # NEW: fiction-layer tab content (split from dock.md, D10)
│   ├── systems.md                   # NEW: telemetry tab — provider table, horizon, spend (D10)
│   ├── villagers.md                 # NEW: villagers tab content (split from dock.md)
│   ├── exercise.md                  # NEW: scenario exercise panel (D11, D4)
│   ├── lesson-row.md                # NEW (decision 5)
│   ├── guardian-strip.md            # NEW (decision 7)
│   ├── villager-strip.md            # NEW (D12)
│   └── minibuffer.md                # update: verify + pin; guardian-strip pairing
├── overlays/
│   ├── help.md                      # NEW: extracted from keymap.md + D9 guardian section
│   │                                #   + byte-identity classification (ruling c)
│   ├── ceremony.md                  # NEW (decision 6; FR-019 ruling; replayability AC)
│   └── postmortem.md                # NEW (decision 6; FR-018 ruling; replayability AC)
└── patterns/
    ├── focus-contract.md            # update: verify + pin; new chrome rows are display-only
    ├── chronicle-grammar.md         # update: verify + pin
    ├── keymap.md                    # update: input-parity doctrine (decision 8); help
    │                                #   section moves out; stays one printable card
    ├── layout.md                    # rewrite: row budget re-derivation + fold order
    │                                #   (ruling a) + narrow chrome rules (ruling b)
    ├── skin-tokens.md               # NEW: doc token conventions + column semantics (D2)
    └── stage-defaults.md            # NEW: stage-resolved visibility defaults (decision 3)

scripts/
└── check-tui-design.mjs             # NEW: structural + same-PR gate (Node ≥18, zero-dep)

CLAUDE.md                            # gate wiring: run-the-check rule (TASK-82 precedent)
```

**Structure Decision**: single-project docs + tooling. The check script lives in
`scripts/` (repo tooling, not a skill — unlike TASK-82's checker it guards tracked
design docs, not a skill's generated output). `panels/dock.md` survives as the
tab-container chrome page only; tab *content* moves to per-tab files so the skin
boundary is a file boundary (D10).

## Complexity Tracking

No constitution violations — table intentionally empty.
