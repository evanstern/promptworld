# Tasks: Card-Format Policy — Gist-First + "As a …" Use Cases

**Input**: Design documents from `/specs/087-card-format-policy/`
**Prerequisites**: plan.md, spec.md, research.md, quickstart.md

**Organization**: Docs-only feature — one file (`CLAUDE.md`), two blocks. All edits
are sequential (same file); the list is deliberately minimal per the plan.

## Phase 1: Setup

No setup required — the worktree `.worktrees/task-168` on branch
`task-168-card-format-policy` already exists with spec artifacts committed.

## Phase 2: Foundational

No foundational tasks — no shared infrastructure; the delivery mechanism (always-on
CLAUDE.md loading) already exists.

## Phase 3: User Story 1 + 3 — Policy in its durable home (P1 + P2)

**Goal**: Card authors (US3) and board readers (US1) get the format rule, the
applicability rule, and the three operator examples in the always-loaded Backlog.md
block.

**Independent Test**: quickstart.md check 1 — the Backlog.md block of `CLAUDE.md`
carries the gist rule, use-case rule (any accurate role), applicability rule
(infra/bookkeeping cards may omit use cases, never the gist), and the good /
less-good / bad examples with reasons.

- [X] T001 [US1] Add a compact "Card format" subsection (~20 lines) to the
  "## Backlog.md — the board" block in `CLAUDE.md`: (a) every task description
  opens with a 1–2 sentence plain-language gist of the deliverable before any other
  content; (b) where applicable, the gist is followed by scene-setting "As a
  \<role\>" use cases, any accurate role valid ("As a player", "As a user", "As a
  villager in the game", "As the Gru"); (c) pure infra/bookkeeping cards may omit
  use cases but never the opening gist; (d) the three operator examples with
  ratings and reasons — good: "As a player, when the game starts up, I want to see
  the map on the left and the chronicle on the right."; less good: "The left side
  of the screen should be the map and the right side should be the chronicle."
  (describes the artifact, not the experience); bad: "The UI is bad, we should fix
  it" — or a wall of file/concept references a human can't easily grok; (e) cite
  TASK-168's card as the first conforming example. (FR-002, FR-003, FR-004, FR-005;
  US3 acceptance rides this same text.)

## Phase 4: User Story 2 — Spec-phase pointer (P2)

**Goal**: The spec agent treats the card's opening gist as the primary statement of
intent for tasks it didn't author.

**Independent Test**: quickstart.md check 2 — the Spec Kit block names the gist as
the primary statement of intent.

- [X] T002 [US2] Add one or two sentences to the "## Spec Kit — specs drive the
  work" block in `CLAUDE.md`: when specifying from a board task the spec author
  didn't write, the card's opening gist and its "As a …" use cases are the primary
  statement of intent — reconstruct purpose from them before mining file/concept
  references. (FR-006. Same file as T001 — sequential, not parallel.)

## Phase 5: Polish & Cross-Cutting

- [X] T003 Run the quickstart.md validation: checks 1–4 (policy present with
  examples; spec-phase pointer present; `git diff origin/main --stat` scoped to
  `CLAUDE.md` + `specs/087-card-format-policy/**` + `.specify/feature.json`;
  TASK-168 card reads as the first conforming example). Record the results in the
  implementer report. (FR-007, SC-003, SC-004.)

## Dependencies

- T001 → T002 (same file; keep diff hunks orderly) → T003 (validates both).
- No parallel opportunities — single file.

## Implementation Strategy

Single increment: T001 alone is a viable MVP (the policy exists and self-propagates);
T002 completes the spec-agent leg; T003 is read-back validation. All three land in
one commit series on the existing branch and merge in TASK-168's one PR.
