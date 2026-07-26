# Feature Specification: Wiki corpus-spec v2 adoption

**Feature Branch**: `071-corpus-v2-adoption` (task branch: `task-146-corpus-v2`)

**Created**: 2026-07-26

**Status**: Draft

**Input**: TASK-146. Surfaced during the TASK-141 re-pin: this corpus predates
corpus-spec v2 — no `CAPSULES.md`; generating one flips the freshness gate's
size budgets from warn-only to FAILURES, and the corpus currently fails 41
ways (35 note bodies over the 8,000-char budget — `event-types` 69.5k,
`testing-strategy` 47.2k, `executor` 40.5k, `tui-client` 37.8k,
`sim-state-reducer` 36.2k, `agent-mind` 33.7k … — and 6 capsules over 500
chars: `guardian` 1286, `guardian-orders` 859, `tool-registry` 798,
`reflex-policy` 697, `guardian-miracles` 508, plus any new drift). Authority:
`~/neumo/projects/praxis/docs/corpus-spec.md` ("Note size budget and
summary-style splits", "The capsule tier and CAPSULES.md").

## Decisions

1. **Summary-style splits, parent keeps its filename.** An over-budget note
   splits into child notes; the parent RETAINS its name/path and keeps a
   one-paragraph summary + `[[wikilink]]` per child; each child links back;
   `INDEX.md` gains one line per child. Keeping parent filenames preserves
   every existing `[[link]]`, every `docs/player/*.html` source meta tag
   (pins bump, paths don't), and the merge-drift gate's note identity.
2. **Split vs tighten vs exempt, per note:** ≥2x over budget → split
   (children ≥1,500 chars of substance, never summary-duplicates); under
   ~1,500 chars over → tighten (cut redundancy) or split if a clean subtopic
   exists; genuinely unsplittable → `size_budget_exempt: <reason>` (expected
   rare — `event-types`' catalog table is the plausible candidate if its rows
   can't split by domain).
3. **Honest pins on restructure:** splitting is restructuring, NOT
   re-verification. A child inherits its parent's `verified_against` (the
   commit its claims were actually verified at); the parent keeps its own
   pin. `sources:` on each child lists exactly the files its retained claims
   depend on; the parent's `sources:` shrinks to what its remaining prose
   claims. No fresh pins without reading diffs.
4. **Capsules are routing text:** every touched note's `description:` is
   rewritten to route (what it covers, when to load it), ≤500 chars; the six
   oversized capsules are rewritten regardless of body work.
5. **CAPSULES.md generated last** (`capsules.mjs`), never hand-edited; its
   presence flips the gate to v2 failure mode — the branch lands only when
   the whole corpus passes that mode (runbook checkpoint: if the split
   stalls, the PR closes unmerged and main stays on v1 warnings).
6. **Execution shape:** orchestrator-led (constitution V: grounding docs are
   the orchestrator's hands) with Sonnet subagent fan-out for per-note
   splits; INDEX/capsule judgment and the final gate stay with the
   orchestrator.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The corpus passes v2 with capsules (Priority: P1)

A session orients capsule-first: reads `CAPSULES.md` (one ≤500-char routing
capsule per note) instead of note bodies; every note body it then loads is
≤8,000 chars or carries an explicit exemption reason.

**Independent Test**: `node <grounding-wiki>/gates/cli.mjs freshness <root>
docs/wiki` exits 0 with `CAPSULES.md` present (v2 failure mode).

**Acceptance Scenarios**:

1. **Given** the branch, **When** the freshness gate runs, **Then** exit 0:
   all pins fresh, all `[[links]]` resolve, all capsules ≤500, all bodies
   ≤8,000 or exempted, CAPSULES.md current.
2. **Given** any split parent, **When** read, **Then** it summarizes each
   child in one paragraph with a resolving `[[wikilink]]`, and the child
   links back.

---

### User Story 2 - Nothing downstream breaks (Priority: P1)

Player docs, the merge-drift gate, and spec-bridge see the same note
identities; only pins moved.

**Independent Test**: `node .claude/skills/player-docs/scripts/check-freshness.mjs
--check` passes in-branch (13/13); `node scripts/check-merge-drift.mjs pr`
passes from the worktree.

**Acceptance Scenarios**:

1. **Given** the branch, **When** the player-docs checker runs in-branch,
   **Then** 13/13 fresh (source paths unchanged; pins bumped).
2. **Given** the pr gate (spec 069, live), **When** run from the worktree,
   **Then** pass — this branch touches no Go sources (wiki-repin vacuous) and
   carries its own player-docs freshness.

---

### Edge Cases

- **Content loss**: splits move prose; a diff-based spot audit (per-note char
  accounting: parent+children ≥ ~95% of original substance net of the
  replaced-by-summary overlap) guards against silent truncation.
- **INDEX.md**: gains child lines; its own malformed-frontmatter info finding
  (pre-existing) is out of scope unless the gate escalates it.
- **Cross-note `[[links]]` into split-off sections**: links point at note
  names, not anchors — parents keep names, so no link breaks; children may
  gain incoming links opportunistically but no link rewrite pass is required.
- **Concurrent lanes**: none — runbook lane 2 owns `docs/wiki/` exclusively;
  refetch before merge anyway.

## Requirements *(mandatory)*

- **FR-001**: Every note body ≤8,000 chars or `size_budget_exempt: <reason>`;
  every capsule ≤500 chars, written for routing.
- **FR-002**: Splits are summary-style per corpus-spec v2; parents keep
  filenames; children ≥1,500 chars substance; INDEX.md updated.
- **FR-003**: Pins stay honest: children inherit the parent's pin; no pin
  advances without re-reading the diff (none expected — restructure only).
- **FR-004**: `CAPSULES.md` generated by `capsules.mjs`, committed; freshness
  gate green in v2 failure mode in-branch.
- **FR-005**: Player docs fresh in-branch (pin bumps only, paths unchanged).
- **FR-006**: The PR merges as a merge commit and only when US1+US2 pass; a
  stalled split closes the PR unmerged (runbook checkpoint).

## Success Criteria *(mandatory)*

- **SC-001**: freshness gate exit 0 with CAPSULES.md present (v2 mode).
- **SC-002**: player-docs 13/13 in-branch; merge-drift pr gate pass.
- **SC-003**: zero broken `[[links]]`; INDEX covers every note.
- **SC-004**: exemptions (if any) each carry a reason a reviewer can judge.

## Assumptions

- corpus-spec.md authority is the praxis repo copy (plugin 0.15.0's skills
  quote the same budgets); no praxisflux code changes needed (gate already
  implements v2 semantics).
- The ~35-note list from the 2026-07-26 gate run may have drifted slightly
  (tile-registry.md landed since); the branch re-derives the worklist from
  the gate's own output.
