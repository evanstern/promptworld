# Implementation Plan: Quickstart first-prompt pass (spec 079, TASK-153)

**Branch**: `task-153-quickstart-first-prompt` | **Spec**: specs/079-quickstart-first-prompt/spec.md

## Summary

Content-only pass over the player docs: `getting-started.html` gains an "ask
your guardian one thing" step whose sample ask is a verbatim, stage-1-legal
value from the `skin.guardian.example_ask.*` family (declared source:
`docs/wiki/skin.md`), and each of the four stage pages gains a short
"Your first session" do-this-then-this block projected from its
already-declared spec 046 sources. Durability rides the player-docs skill's
editorial contract: `.claude/skills/player-docs/SKILL.md` is amended in the
same branch (mapping rows reconciled for the five touched pages + shape
notes) so the next regeneration reproduces the content instead of erasing it
(research.md R1–R2 — the load-bearing mechanics decision).

## Technical Context

**Language/Version**: HTML (self-contained pages, no JS, shared inline CSS
block per SKILL.md skeleton); Markdown (SKILL.md)
**Primary Dependencies**: none — no code changes
**Testing**:
`node .claude/skills/player-docs/scripts/check-freshness.mjs --check`
(provenance gate, must exit 0); `go test ./...` (doctrine run, unaffected);
byte-identity of untouched pages via `git diff --stat`
**Target Platform**: static HTML read in a browser
**Project Type**: documentation content (player docs projection)
**Performance Goals**: n/a
**Constraints**: FR-008 — only `docs/player/{getting-started,stage-1..4}.html`
and `.claude/skills/player-docs/SKILL.md` change; no wiki edits, no
`index.html` change, no source-tag changes on stage pages
**Scale/Scope**: 5 HTML pages amended, 1 skill contract amended

## Constitution Check

- **I. Artifact-Grounded Action** — PASS: every content claim traces to a
  declared source at a recorded pin; the mechanics decision is recorded in
  research.md; the sample-ask constraint is pinned to stage-gating.md's
  ratified ceiling.
- **II. One Task, One PR** — PASS: TASK-153 ↔ branch
  `task-153-quickstart-first-prompt` ↔ one PR; spec phases are internal
  breakdown.
- **III. Gates Over Assertions** — PASS: the freshness probe and the
  merge-drift pr gate are the acceptance instruments; no derived state
  hand-edited.
- **IV. Grounding Freshness** — PASS: no wiki-note source files are touched
  (verified, research.md R7), so no re-pins are owed; `docs/player/` changes
  ride the task branch and the pr gate's `player-docs-stale` probe verifies
  exit 0; merge-commit-only.
- **V. Model-Tiered Workflow** — PASS: planning artifacts on Fable 5;
  implementation delegated to the `spec-implementer` agent on **Sonnet**
  (routine slice: doc reconciliation / content, single-surface, no
  concurrency or architecture) — tier + rubric justification recorded on
  TASK-153.

Re-check after design: no violations; Complexity Tracking not needed (no
deviations).

## Project Structure

### Documentation (this feature)

```
specs/079-quickstart-first-prompt/
├── CLAIM.md        # claim stub (kept)
├── spec.md         # feature specification
├── research.md     # the generated-vs-authored mechanics decision (R1–R7)
├── plan.md         # this file
└── tasks.md        # phase/task breakdown mapped to the 2 board ACs
```

### Touched files (repository root)

```
docs/player/getting-started.html        # + first-prompt step, + skin.md source tag
docs/player/stage-1-the-voice.html      # + first-session block (sources unchanged)
docs/player/stage-2-the-written-word.html
docs/player/stage-3-the-craft.html
docs/player/stage-4-the-stewardship.html
.claude/skills/player-docs/SKILL.md     # mapping rows (5 pages) + shape notes
```

No source-code directories are touched.

## Phase 0 — Research

Complete: research.md answers the one open question (where a player-docs
content change durably lives) from the artifacts — pages are
authored-projections with provenance-only gating; durability = page + meta
tag + SKILL.md editorial contract, all in-branch. R3/R4 inventory the token
family, the stage-1 ceiling, and the per-stage projectable facts; R5 resolves
the three named edge cases; R7 verifies gate expectations.

## Phase 1 — Design

No data-model.md or contracts/ needed: the feature's "contract" is SKILL.md's
existing provenance meta-tag format
(`specs/026-player-docs/contracts/provenance-and-check.md`) and mapping
table, which this feature amends rather than replaces. Design decisions all
pinned in spec.md FRs: step placement (after watch-it-live), sample-ask
constraint (verbatim, stage-1 ceiling verb, recommended `send_vision`), skin
honesty note, stage pages link-don't-quote, bounded mapping reconciliation.

## Phase 2 — Task generation approach

tasks.md phases map 1:1 onto the two board ACs, with the SKILL.md contract
amendment as the shared foundational phase (it defines what both page-content
phases must satisfy, and US3 makes it a P1 durability requirement):

1. Setup/baseline (probe green, pins captured)
2. Foundational: SKILL.md editorial contract (serves US3 + both ACs)
3. AC #1 / US1: getting-started first-prompt step + source tag
4. AC #2 / US2: four stage-page first-session blocks (parallelizable per page)
5. Gates & polish: probe exit 0, byte-identity, go test, merge-drift pr gate

## Complexity Tracking

None — no constitutional deviations.
