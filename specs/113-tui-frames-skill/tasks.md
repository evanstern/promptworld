# Tasks — spec 113, tui-frames project skill

**Board task:** TASK-190 · **Branch:** `task-190-tui-frames-skill`

One task, one PR. Each phase lands as a commit on this single branch.

## Phase 1 — Author the skill

- [X] T001 Create `.claude/skills/tui-frames/SKILL.md` with frontmatter matching
      `.claude/skills/player-docs/SKILL.md`'s shape: `name`, `description`,
      `metadata.author: "promptworld"`, `user-invocable: true`,
      `disable-model-invocation: false`. The `description` must name the check command, so
      an agent sees the probe without opening the file (FR-001).
- [X] T002 Body — **how to see a screen** (FR-002): read
      `docs/design/tui/frames/<fixture>__<state>__<WxH>.txt`. State the never-hand-edit rule
      and *why*: an edit is a false claim about what the client renders, erased by the next
      `--dump`.
- [X] T003 Body — **the surface** (FR-003): `--list`; live render for off-matrix sizes; the
      four committed sizes each with the boundary it straddles; `--ansi`; `--interactive`
      for feel rather than layout.
- [X] T004 Body — **the design loop** (FR-004): target frame first, implement until
      `--dump` matches, the frame diff is the review artifact.
- [X] T005 Body — **authority vs evidence** (FR-005) and the **narrow-fallback caveat**
      (FR-006, from spec 112 FR-008).
- [X] T006 Body — **the three repo traps** (FR-007): root read-only + worktree recipe; bare
      `git commit -F <file>`; merge-commit-only.
- [X] T007 Add `.claude/skills/tui-frames/scripts/check-frames.mjs` (FR-008): dump to a temp
      dir via `--out`, diff against the committed matrix ignoring `README.md`, exit 0 when
      identical and nonzero listing every differing/missing/extra file. Must not mutate the
      committed matrix and must leave the working tree clean.

## Phase 2 — Verify and ground

- [X] T008 Run the guard on an unmodified tree — exits 0, `git status` clean afterwards.
- [X] T009 **Negative control:** append a byte to one committed frame, confirm the guard
      exits nonzero and names that file, then restore. Record the observed output.
- [X] T010 Confirm the skill is discoverable: it appears in the skills listing with its
      description, and `.claude/skills/tui-frames/` matches the `player-docs` layout.
- [X] T011 Wiki-in-PR (spec 069): if the branch touches any file a `docs/wiki/` note lists in
      `sources:`, re-verify and re-pin that note in-branch; if `docs/wiki/` changes,
      regenerate `docs/player/`. Probe:
      `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`.
- [X] T012 `node scripts/check-merge-drift.mjs pr` exits 0 from the worktree.
