# Feature Specification: merge-drift pr gate — docs-stale probe fires on all pinned sources and after history moves

**Feature Branch**: `task-162-pr-gate-docs-stale-probe`

**Created**: 2026-07-29

**Status**: Draft

**Input**: User description: "merge-drift pr gate: docs-stale probe fires on all pinned sources and after history moves (TASK-162)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A branch that touches non-wiki pinned inputs is still freshness-gated (Priority: P1)

As a session opening a PR whose branch edits `README.md` or `docs/llm-providers.md`
(but nothing under `docs/wiki/`), I want the pr gate to run the player-docs freshness
checker anyway, so a stale generated page can never ride to main just because the
branch avoided the wiki directory.

**Why this priority**: This is the observed blind spot that motivated the task — the
checker's own inputs include README.md, docs/llm-providers.md, and the spec 046
quickstart sources, yet today only `docs/wiki/` changes trigger it
(scripts/check-merge-drift.mjs:1645). A keymap-doc-only change staling a pinned page
invisibly is a recorded field case.

**Independent Test**: On a synthetic branch that modifies only `README.md` while
`docs/player/` is stale relative to it, `node scripts/check-merge-drift.mjs pr` exits
1 with a `player-docs-stale` finding.

**Acceptance Scenarios**:

1. **Given** a branch whose diff vs origin/main touches `README.md` only, **When** the
   pr gate runs and the freshness checker reports stale, **Then** the gate blocks with
   `player-docs-stale` naming the stale page(s).
2. **Given** a branch touching `docs/llm-providers.md` or a spec 046 quickstart
   source, **When** the pr gate runs, **Then** the freshness checker is invoked
   (blocking on stale / env-error exactly as the wiki-triggered path does).
3. **Given** a branch touching none of the checker's pinned inputs and with no
   history move, **When** the pr gate runs, **Then** the checker is not invoked
   (current no-op behavior preserved).

---

### User Story 2 - Design-reference pins block, not warn (Priority: P2)

As a reviewer relying on the spec 047 UI authority, I want a PR that touches
`internal/tui/` (or the design-reference pages themselves) to be BLOCKED when
`docs/design/tui/` pins are stale against the branch, instead of receiving only the
non-blocking `tui-surface` warning (scripts/check-merge-drift.mjs:1704-1718), so the
design reference cannot silently drift.

**Why this priority**: The warn-level notice is advisory and routinely scrolls past;
the design reference is a gate-backed authority everywhere else in the lifecycle.

**Independent Test**: On a synthetic branch that modifies a file pinned by a
`docs/design/tui/` page without re-pinning that page, the pr gate exits 1 with a
blocking design-reference finding; after re-verifying + re-pinning the page on the
branch, the gate passes.

**Acceptance Scenarios**:

1. **Given** a branch touching a source pinned by a design-reference page without
   amending that page, **When** the pr gate runs, **Then** it blocks with a finding
   naming the page and the touched source.
2. **Given** the same branch after the page is re-verified and re-pinned on the
   branch, **When** the pr gate runs, **Then** no design-reference finding is raised
   (the existing `tui-surface` reminder may remain as a warn).

---

### User Story 3 - History moves re-trigger the probe (Priority: P2)

As a session that just merged `origin/main` into a pin-carrying branch, I want the pr
gate to re-run the freshness probe even though the branch's own diff vs origin/main
gained no pinned-source paths, so staleness introduced by the history move itself
(the recorded merge-main-into-pin-carrying-branches hazard) is caught before
`gh pr create`.

**Why this priority**: History moves change what the branch tip contains without
changing its diff vs origin/main — the trigger predicate on the diff alone provably
misses them.

**Independent Test**: On a synthetic branch whose tip is a merge of current
origin/main (diff vs origin/main touching no pinned source) where the merged-in main
side staled `docs/player/`, the pr gate invokes the checker and blocks.

**Acceptance Scenarios**:

1. **Given** a branch whose tip is (or contains, since diverging from origin/main) a
   merge commit bringing in main-side history, **When** the pr gate runs, **Then**
   the freshness probe is invoked regardless of the branch's own diff paths.
2. **Given** such a branch where the probe reports fresh, **When** the pr gate runs,
   **Then** it passes — the re-trigger adds no false blocking.

---

### Edge Cases

- Branch touches a pinned input AND carries a history move: probe runs once, not
  twice; findings are not duplicated.
- Freshness checker missing or exiting 2 on a non-wiki trigger path: same
  `player-docs-env-error` blocking behavior as the existing wiki path.
- Design-reference check on a branch that touches `docs/design/tui/` pages only
  (re-pin-only PRs): must not block — a page amendment with no source drift is the
  compliant shape.
- Detached/synthetic test branches (fixture harness): triggers must be computable
  from git state alone (no network, no board access), matching the script's existing
  fixture approach.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: pr mode MUST invoke the player-docs freshness checker when the branch's
  diff vs origin/main touches ANY of the checker's pinned inputs — `docs/wiki/`,
  `README.md`, `docs/llm-providers.md`, and the spec 046 quickstart sources — not
  only `docs/wiki/`. The set of trigger paths MUST live in one named place in the
  script so future checker inputs extend it in one edit.
- **FR-002**: pr mode MUST freshness-gate design-reference pins (`docs/design/tui/*`)
  with a BLOCKING finding when the branch touches a source pinned by a
  design-reference page without carrying that page's re-verification — the same
  pin-vs-branch predicate shape as `wiki-repin-missing`. The existing warn-level
  `tui-surface` notice MAY remain as a reminder but is no longer the only signal.
- **FR-003**: pr mode MUST run the freshness probe when the branch has undergone a
  history move — its tip is or contains, since diverging from origin/main, a merge
  commit bringing in main-side history — even when the branch's own diff vs
  origin/main touches no pinned source.
- **FR-004**: A branch matching multiple triggers (pinned input + history move) MUST
  produce each finding at most once per run.
- **FR-005**: Checker-missing and checker-exit-2 conditions on the new trigger paths
  MUST block with the existing `player-docs-env-error` rule, identically to the
  wiki-triggered path.
- **FR-006**: Every new trigger MUST be covered by the script's existing
  test/fixture approach (synthetic branch cases: non-wiki pinned input, design-pin
  drift, history move, combined, and the preserved no-trigger no-op).

### Key Entities

- **Pinned input set**: the enumerated file paths/prefixes whose change triggers the
  player-docs freshness probe (wiki corpus, README, llm-providers doc, spec 046
  quickstart sources).
- **Design-reference pin**: a `docs/design/tui/` page's verified-against pin over its
  listed sources — same shape as a wiki note pin.
- **History move**: a merge of main-side history into the branch since it diverged
  from origin/main (detectable from the commit graph alone).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A stale `docs/player/` page can no longer reach a PR through any change
  to the checker's pinned inputs — every enumerated trigger path has a fixture
  proving the gate blocks (0 escapes across the fixture matrix).
- **SC-002**: Design-reference pin drift on a PR branch is a blocking finding, proven
  by fixture both ways (drift blocks; re-pinned passes).
- **SC-003**: A merge-of-main into a branch with no pinned-source diff paths triggers
  the probe, proven by fixture both ways (stale blocks; fresh passes).
- **SC-004**: All existing pr-gate fixtures still pass unchanged (no regression in
  current blocking/warning behavior).

## Assumptions

- The player-docs freshness checker's pinned-input list (README.md,
  docs/llm-providers.md, spec 046 quickstart sources) is stable enough to enumerate
  in the gate script; the single-named-place requirement (FR-001) contains the cost
  of future drift.
- Blocking severity for design-reference pin drift is scoped to pr mode only —
  session/worktree modes keep their current reporting.
- "Since the last probe" from the task card is interpreted as graph-detectable
  history moves (FR-003) rather than persisted probe timestamps — the gate stays
  stateless, consistent with its no-daemon design (spec 051).
- The TASK-165 wiki size-budget findings are a different finding class
  (grounding-wiki internal gate) and are out of scope here.
