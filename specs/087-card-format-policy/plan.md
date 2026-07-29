# Implementation Plan: Card-Format Policy — Gist-First + "As a …" Use Cases

**Branch**: `task-168-card-format-policy` | **Date**: 2026-07-29 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/087-card-format-policy/spec.md`

## Summary

Make it repo policy that every Backlog.md task description opens with a
one-or-two-sentence plain-language gist, followed (where applicable) by "As a
\<role\>" use cases. Implementation is a documentation change in exactly one file:
the project `CLAUDE.md` — the policy (rule, applicability, operator's three
examples) lands inline in the **Backlog.md block** (the durable home every session
loads, FR-001), and a one-line pointer lands in the **Spec Kit block** directing the
spec phase to treat the card's opening gist as the primary statement of intent
(FR-006).

## Technical Context

**Language/Version**: Markdown (project guidance docs); no code

**Primary Dependencies**: none — the always-on loading of project `CLAUDE.md` is the
delivery mechanism

**Storage**: git-tracked files in this repo

**Testing**: manual validation per quickstart.md (read-back checks); no automated
gate (explicitly out of scope per spec Assumptions)

**Target Platform**: N/A (documentation)

**Project Type**: documentation/policy change

**Performance Goals**: N/A

**Constraints**: repo-local only — no praxisflux plugin/template changes (FR-007);
keep the CLAUDE.md addition compact (the file is always-loaded context; budget
~20 lines for the policy block)

**Scale/Scope**: 1 file edited (`CLAUDE.md`), 2 blocks touched (Backlog.md block,
Spec Kit block)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Artifact-Grounded Action — PASS.** The policy's durable home is a tracked
  file; the decision trail is TASK-168 → spec 087 → this plan → the PR.
- **II. One Task, One PR — PASS.** TASK-168 ↔ branch `task-168-card-format-policy`
  ↔ one PR; spec phases are internal breakdown.
- **III. Gates Over Assertions — PASS (no new gate).** The spec explicitly scopes
  enforcement to convention and review; no derived state is hand-edited. The
  spec-bridge gate governs the task's board status as usual.
- **IV. Grounding Freshness — PASS.** `CLAUDE.md` is not a pinned source of any
  wiki note (verified: only `docs/wiki/guardian.md` mentions it, in prose; its
  `sources:` are Go files). No wiki re-pin or player-docs regeneration is expected;
  the merge-drift pr gate probes this mechanically before the PR.
- **V. Model-Tiered Workflow — PASS.** Fable 5 authored spec/plan/tasks; the edit
  is dispatched to the spec-implementer agent on Sonnet (routine tier: doc
  reconciliation), recorded on the board task.

No violations → Complexity Tracking not required.

## Project Structure

### Documentation (this feature)

```text
specs/087-card-format-policy/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0 output (durable-home decision record)
├── quickstart.md        # Phase 1 output (validation guide)
├── checklists/
│   └── requirements.md  # Spec quality checklist (complete)
└── tasks.md             # Phase 2 output (/speckit-tasks — not created by plan)
```

data-model.md and contracts/ are intentionally omitted: the feature has no data
entities and exposes no interface — it is prose policy in guidance docs.

### Source Code (repository root)

```text
CLAUDE.md                # the ONLY file edited
├── "## Backlog.md — the board" block   ← policy: gist-first rule, applicability,
│                                          the three operator examples
└── "## Spec Kit — specs drive the work" block ← one-line spec-phase pointer:
                                           the card's gist is the primary
                                           statement of intent
```

**Structure Decision**: single-file documentation edit; both policy homes live in
`CLAUDE.md` because it is the one artifact provably loaded by every card author
(human sessions, agents) and by the spec agent — a separate referenced doc would
add a load-ordering dependency the spec's FR-001 warns about (see research.md R1).

## Complexity Tracking

Not required — no constitution violations.
