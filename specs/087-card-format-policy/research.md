# Research: Card-Format Policy (spec 087)

No NEEDS CLARIFICATION markers existed in the Technical Context; the decisions below
record the durable-home and delivery choices with alternatives considered.

## R1 — Durable home for the policy

- **Decision**: Inline in the project `CLAUDE.md`, inside the existing
  "## Backlog.md — the board" block, as a compact subsection (~20 lines including
  the three operator examples).
- **Rationale**: FR-001 requires a home that card authors AND the spec agent load
  when creating tasks or specs. `CLAUDE.md` is the only artifact loaded
  unconditionally into every session (human-driven or agent); the Backlog.md block
  is exactly where task-authoring guidance already lives (statuses, CLI usage,
  one-task-one-PR), so authors reading how to create a card meet the format rule in
  the same place.
- **Alternatives considered**:
  - *Separate doc (e.g. `docs/design/card-format-policy.md`) referenced from the
    block* — rejected: adds a load hop that agents skip in practice unless a hook
    forces it; the policy is short enough to inline, and the spec's own SC-004 asks
    for one-place readability reachable from always-loaded guidance.
  - *Backlog.md definition-of-done defaults / board config* — rejected: DoD attaches
    per-task checklist items, not authoring-time description format; it fires at
    completion, not creation.
  - *praxisflux plugin template change* — out of scope by FR-007 (the card and AC #3
    make this a repo-local policy deliberately).

## R2 — Spec-phase guidance pointer

- **Decision**: One sentence added to the existing "## Spec Kit — specs drive the
  work" block of `CLAUDE.md`: when specifying from a board task the spec author
  didn't write, the card's opening gist (and its "As a …" use cases) is the primary
  statement of intent.
- **Rationale**: FR-006 targets "spec-phase guidance the repo owns". The Spec Kit
  block is the repo-owned guidance loaded whenever specs are worked;
  `.specify/templates/spec-template.md` was considered but the template shapes the
  OUTPUT document, not how the author reads the INPUT card — the pointer belongs
  where the workflow is described.
- **Alternatives considered**:
  - *Edit `.specify/templates/spec-template.md`* — permitted (repo-local) but
    rejected: the template governs spec structure; putting reading guidance there
    hides it until after the template is opened, and template churn invites drift
    against upstream Spec Kit updates.
  - *Edit `.specify/memory/constitution.md`* — rejected: the constitution governs
    principles and gates; a card-format convention is guidance, not a principle, and
    constitution changes carry versioning ceremony this doesn't warrant.

## R3 — Enforcement mechanism

- **Decision**: Convention + review only; no lint/gate over card text.
- **Rationale**: Spec Assumptions rule out a mechanical gate; card text is prose and
  a textual gate would misfire (the operator's own examples show judgment calls like
  "less good"). The policy self-propagates through the always-loaded home.
- **Alternatives considered**: a `backlog` wrapper or hook linting new cards —
  rejected as out of scope and brittle.

## R4 — Agent-context script

- **Decision**: Skip `.specify/scripts/bash/update-agent-context.sh` for this
  feature.
- **Rationale**: the script injects tech-stack context into agent files; this
  feature has no tech stack, and this repo's `CLAUDE.md` is hand-curated PDLC
  grounding with no generated section (no prior spec has run it here). Running it
  would risk polluting the very file this feature edits.
