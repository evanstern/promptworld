# Tasks — spec 117, polish-session project skill

**Board task:** TASK-198 · **Branch:** `task-198-polish-skill`

One task, one PR. Each phase lands as a commit on this single branch.

## Phase 1 — Author the skill

- [X] T001 Create `.claude/skills/polish-session/SKILL.md` with frontmatter matching
      `.claude/skills/tui-frames/SKILL.md`'s shape: `name`, `description`,
      `metadata.author: "promptworld"`, `user-invocable: true`,
      `disable-model-invocation: false`. The `description` must name the status probe and the
      worked example, so an agent sees both without opening the file (FR-001).
- [X] T002 Body — **stub-first, stated before the loop** (FR-003): the rule, the `spec-bridge`
      filesystem-derivation mechanism that makes it necessary, why freeform sessions are the
      exposed case, and the cost of retrofitting. Cite CLAUDE.md's bullet and praxisflux
      TASK-104; do not restate them.
- [X] T003 Body — **the loop** (FR-002): the nine steps in order, imperative, with the
      decision-before-implementation rule inside the sub-loop where it applies.
- [X] T004 Body — **the ad-hoc-versus-spec test** (FR-004): the trivial-exemption bar verbatim
      from the constitution, plus its contrapositive as the escalation signal.
- [X] T005 Body — **grounding** (FR-005): routed to step 8 once, in-branch; the idempotence
      argument for why deferring is free; `wikiNotes=N` and the `wiki-footprint` warning as the
      instrument. No cadence.
- [X] T006 Body — **the two traps that look like chores** (FR-006): the stale QA binary and
      golden-test-failure-is-signal.
- [X] T007 Body — **the worked example** (FR-007): TASK-195, PR #163, spec 115, cited as the
      place to read the full run. Verify the skill nowhere re-narrates it.

## Phase 2 — Template and probe

- [X] T008 Add `.claude/skills/polish-session/templates/session-card.md` (FR-008): gist,
      "As a …" use cases per spec 087, the session contract, the decision-log section, and the
      five acceptance criteria that make the flow checkable. Derived from TASK-195's card.
- [X] T009 Add `.claude/skills/polish-session/scripts/session-status.mjs` (FR-009): binary
      freshness against the newest tracked Go source; wiki footprint read from
      `check-merge-drift.mjs session --json` by matching the current worktree. Read-only,
      advisory exit code, threshold not hardcoded.
- [X] T010 Add the CLAUDE.md pointer block (FR-010), in the shape of the existing TUI-design and
      wiki-in-PR pointers. Pointer only — the skill stays canonical.

## Phase 3 — Verify and ground

- [X] T011 Run the probe on this worktree with no binary present — confirm it reports the
      absent-binary state distinctly and explains the fix (SC-007, R4).
- [X] T012 **Positive control:** build `promptworld` in the worktree, re-run the probe, confirm
      the binary check passes and `git status` is clean afterwards (SC-007).
- [X] T013 **Negative control:** touch a tracked `.go` file so it post-dates the binary, confirm
      the probe reports the binary stale and exits nonzero, then restore (SC-007).
- [X] T014 Confirm the footprint reading matches the session gate's own `wikiNotes=N` for this
      branch — two derivations, same number.
- [X] T015 Confirm the skill is discoverable: it appears in the skills listing with its
      description, and `.claude/skills/polish-session/` matches the repo's project-skill layout.
- [X] T016 Read `SKILL.md` against SC-001..SC-006 and record the check on the board card.
- [X] T017 Wiki-in-PR (spec 069): if the branch touches any file a `docs/wiki/` note lists in
      `sources:`, re-verify and re-pin that note in-branch; if `docs/wiki/` changes, regenerate
      `docs/player/`. Probe:
      `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`.
- [X] T018 `node scripts/check-merge-drift.mjs pr` exits 0 from the worktree.
