# Feature Specification: Card-Format Policy — Gist-First + "As a …" Use Cases

**Feature Branch**: `task-168-card-format-policy`

**Created**: 2026-07-29

**Status**: Draft

**Input**: User description: "Card-format policy (TASK-168): make it project policy that every Backlog.md task description opens with a one-or-two-sentence plain-language gist of the deliverable, followed where applicable by a few scene-setting use cases in 'As a <role>' form. The policy must live in a durable tracked home that card authors (human or agent) and the spec agent load when creating tasks or specs; must state when use cases apply vs may be skipped; must include the operator's good/less-good/bad format examples; and spec-phase guidance must point at the gist section as the primary statement of intent. Scope: this repo only."

Every Backlog.md task description on this board opens with a one-or-two-sentence
plain-language gist of the deliverable, followed (where applicable) by scene-setting
use cases in "As a \<role\>" form — and the policy saying so lives where card authors
and the spec agent actually load it.

As a hu-mon scanning the board, I want to catch what a card is about in a few
seconds. As the spec agent working a task I didn't author, I want the card to state
its purpose plainly up front. (This spec's own opening, like the TASK-168 card, is
written in the mandated format.)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Board reader gets the gist in seconds (Priority: P1)

A person (or agent) scanning `backlog task list` opens a card and, from its first one
or two sentences alone, understands in plain language what the deliverable is —
without decoding file paths, spec numbers, or internal jargon.

**Why this priority**: This is the core value of the policy — the board is the plan
of record, and a plan humans can't skim isn't functioning as one.

**Independent Test**: Open any card authored after the policy lands; read only its
opening sentences; a reader unfamiliar with the task can say what it delivers.

**Acceptance Scenarios**:

1. **Given** the policy is in force, **When** a new task is created (by human or
   agent), **Then** its description opens with a one-or-two-sentence plain-language
   gist of the deliverable before any other content.
2. **Given** a card whose subject serves identifiable people/roles, **When** the card
   is authored, **Then** the gist is followed by one or more "As a \<role\>" use cases
   using any accurate scene-setting role ("As a player", "As the Gru", "As a villager
   in the game", …).

---

### User Story 2 - Spec agent finds intent up front (Priority: P2)

The spec agent, writing a spec from a task card it didn't author, treats the card's
opening gist (and its use cases) as the primary statement of intent — instead of
reconstructing the ticket's purpose from file references during specify/clarify.

**Why this priority**: Reduces churn in the highest-leverage consumer of card text;
depends on Story 1's format existing.

**Independent Test**: Read the spec-phase guidance the project loads when specifying;
it explicitly directs the spec author to the card's opening gist as the primary
statement of intent.

**Acceptance Scenarios**:

1. **Given** the policy has landed, **When** the spec agent (or a human) consults the
   project's spec-phase guidance, **Then** that guidance names the card's opening
   gist section as the primary statement of intent for tasks the spec author didn't
   write.

---

### User Story 3 - Card author has a concrete format to follow (Priority: P2)

A card author — human or agent — creating a task loads the project's standing
guidance and finds the format stated concretely: gist first, then use cases, with
good / less-good / bad examples, and an explicit rule for when use cases may be
skipped.

**Why this priority**: The policy only self-propagates if the format is unambiguous
at authoring time; equal in weight to Story 2 but meaningless without Story 1.

**Independent Test**: Read the policy's durable home; it contains the format rule,
the three operator examples, and the applicability rule, in a location loaded during
task creation.

**Acceptance Scenarios**:

1. **Given** the policy's durable home, **When** an author reads it, **Then** it
   shows the operator's examples: good ("As a player, when the game starts up, I want
   to see the map on the left and the chronicle on the right."), less good ("The left
   side of the screen should be the map and the right side should be the chronicle."
   — describes the artifact, not the experience), and bad ("The UI is bad, we should
   fix it" — or a wall of file/concept references a human can't easily grok).
2. **Given** a pure infra/bookkeeping card (e.g. a flaky-test fix or a gate tweak),
   **When** it is authored under the policy, **Then** it may omit use cases but MUST
   still open with the plain-language gist.

---

### Edge Cases

- Pure infra/bookkeeping cards: use cases may be skipped; the opening gist may not.
- Cards created as residuals/follow-ups by agents mid-task (e.g. board-sync residual
  carding) are still task creations and must comply.
- Existing cards predating the policy are not retrofitted; the policy applies to
  cards created or substantially rewritten after it lands.
- A card whose natural "role" is non-human or in-fiction ("As the Gru", "As a
  villager in the game") is valid — any accurate scene-setting role counts.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The policy MUST live in a durable, git-tracked home that card authors
  (human or agent) and the spec agent load when creating tasks or specs — the
  Backlog.md block of the project CLAUDE.md, or a doc that block references and
  directs authors to read.
- **FR-002**: The policy MUST require every task description to open with a
  one-or-two-sentence plain-language gist of the deliverable, before any other
  content.
- **FR-003**: The policy MUST require, where applicable, scene-setting use cases in
  "As a \<role\>" form following the gist, and MUST state that any accurate role is
  valid (player, user, villager in the game, the Gru, …).
- **FR-004**: The policy MUST state when use cases apply versus may be skipped: pure
  infra/bookkeeping cards may omit use cases, but no card may omit the opening gist.
- **FR-005**: The policy MUST include the operator's three format examples (good /
  less good / bad) with the reason each earns its rating.
- **FR-006**: The project's spec-phase guidance MUST point at the card's opening gist
  section as the primary statement of intent for a task the spec author didn't write.
- **FR-007**: All changes MUST be scoped to this repository's tracked files — no
  praxisflux plugin or plugin-template changes.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A reader unfamiliar with a post-policy card can state what it delivers
  after reading only its opening sentences (spot-check on the next 3 cards created
  after the policy lands).
- **SC-002**: 100% of task descriptions created after the policy lands open with a
  plain-language gist; those with identifiable roles carry at least one "As a
  \<role\>" use case.
- **SC-003**: The spec-phase guidance names the gist as the primary statement of
  intent — verifiable by reading the guidance text.
- **SC-004**: The policy, its examples, and its applicability rule are readable in
  one place reachable from the always-loaded project guidance.

## Assumptions

- The TASK-168 card itself is the first conforming example and may be cited by the
  policy as such.
- "Durable home" is satisfied by the project CLAUDE.md's Backlog.md block carrying
  the policy (inline or via a referenced tracked doc); the block is always loaded in
  every session, which covers both card authors and the spec agent.
- Spec-phase guidance means guidance this repo owns (e.g. the CLAUDE.md Spec Kit
  block or repo-local spec-phase docs) — not the praxisflux plugin's skill files
  (out of scope per FR-007). Repo-local `.specify/` template edits are permitted if
  the plan chooses them, since `.specify/` is tracked in this repo, but the praxisflux
  marketplace/plugin cache is not touchable.
- Existing cards are not retrofitted; enforcement is by convention and review, not a
  new mechanical gate (no linting of card text is in scope).
