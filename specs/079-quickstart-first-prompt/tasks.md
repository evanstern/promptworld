# Tasks: Quickstart first-prompt pass (spec 079, TASK-153)

**Input**: spec.md, plan.md, research.md in this directory
**Branch**: `task-153-quickstart-first-prompt` (worktree `.worktrees/task-153`,
cut via `node scripts/check-merge-drift.mjs worktree --spec 079 --task TASK-153`)

**Scope guard (FR-008)**: the branch may change ONLY
`docs/player/getting-started.html`, the four `docs/player/stage-N-*.html`
pages, and `.claude/skills/player-docs/SKILL.md`. No Go code, no
`docs/wiki/`, no `docs/design/`, no `index.html`.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: can run in parallel (different files, no dependency)
- **[Story]**: US1 (board AC #1), US2 (board AC #2), US3 (regeneration durability)

## Phase 1: Setup & baseline

- [X] T001 Verify the baseline: from the worktree run
      `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`
      — must exit 0 (13 fresh) before any edit; if not, stop and surface
      (another lane's churn landed; re-read research.md R7).
- [X] T002 Capture current pins: record `docs/wiki/skin.md`'s
      `verified_against` value (spec-time: `31c893e0406653197e467a89b2fdb96f0bcf2ee0`)
      and confirm the `skin.guardian.example_ask.*` values in that note still
      match `internal/skin/skin.go`'s default table byte-for-byte
      (spec.md Assumptions; divergence = stop, wiki bug upstream).

## Phase 2: Foundational — the editorial contract (blocks all page edits)

**Purpose**: US3 (P1) — make the next regeneration reproduce this content.
SKILL.md defines what Phases 3–4 must write, so it lands first.

- [X] T003 [US3] Amend `.claude/skills/player-docs/SKILL.md` mapping table:
      reconcile the rows for the five touched pages to their actual declared
      sources post-change — `getting-started.html` row gains
      `docs/wiki/skin.md` AND is corrected to list the page's full current
      source set (12 sources on main + skin.md; bounded reconciliation,
      research.md R6); the four `stage-N-*.html` rows re-checked against
      those pages' tags (expected: unchanged).
- [X] T004 [US3] Add SKILL.md editorial shape notes (precedent: the TASK-114
      pair and TASK-68 quartet paragraphs): (a) getting-started carries an
      "ask your guardian one thing" step after watch-it-live — sample ask
      byte-verbatim from the `skin.guardian.example_ask.*` family as
      documented in `docs/wiki/skin.md`, verb within the pinned stage-1
      ceiling ({send_vision, send_omen, monitor_and_act, cancel_order}),
      default-skin phrasing with the custom-skin honesty note, and the `?`
      overlay guardian section named as the live per-world list; (b) each
      stage page carries a "Your first session" do-this-then-this block
      (3–5 steps, scenario-independent, exercises as when-ready follow-ons,
      stage-4 states there is none, ask linked to getting-started's step —
      never re-quoting token values). Cite TASK-153 / spec 079.

**Checkpoint**: SKILL.md now states everything Phases 3–4 must produce.

## Phase 3: User Story 1 — getting-started first-prompt step (board AC #1) 🎯

- [X] T005 [US1] In `docs/player/getting-started.html`, insert the new
      numbered step "Ask your guardian one thing" between the current
      "4. Watch it live" and "5. When you're done" (renumber the stopping
      step): where to type (the guardian message box in `promptworld ui`,
      or `G` for the full console), the recommended sample ask quoted
      byte-verbatim — `"show Ash a vision of the fire dying"`
      (`skin.guardian.example_ask.send_vision`) — what to expect back (the
      guardian replies/acts; the charge line above the message box shows the
      cost), and that this works on any world, no scenario needed (FR-001,
      FR-002, FR-006 spirit).
- [X] T006 [US1] In the same step, add the `?`-overlay pointer and the skin
      honesty note (FR-003): the overlay's guardian section lists one example
      ask per verb *your world* grants, in *your world's* phrasing; the
      phrasings printed here are the default Guardian skin's.
- [X] T007 [US1] Add the source meta tag (FR-004): one new line in `<head>` —
      `<meta name="promptworld-docs:source" content="docs/wiki/skin.md@<current verified_against>">`
      — grammar `<path>@<40-hex-lowercase>`, pin captured in T002. Touch no
      other tag.

**Checkpoint**: board AC #1 satisfiable; page self-consistent with SKILL.md.

## Phase 4: User Story 2 — stage-page first-session blocks (board AC #2)

All four tasks are parallelizable [P] (different files) and constrained by
FR-005/FR-006: 3–5 ordered steps, every claim projected from the page's
already-declared spec 046 sources, NO source-tag changes, ask linked to
getting-started's step, executable without `--scenario`.

- [X] T008 [P] [US2] `docs/player/stage-1-the-voice.html`: "Your first
      session" block — create (`promptworld new my-village`, default stage) →
      start → `ui` → ask one thing (link to getting-started's first-prompt
      step; vision/omen/watch are the stage's whole verb set) → when ready,
      re-create with `--scenario first-night` and keep everyone alive to dawn
      of day 2.
- [X] T009 [P] [US2] `docs/player/stage-2-the-written-word.html`: block
      includes opening `charter.md` and writing one durable line (in force
      from the guardian's next turn), watching it hold without re-asking;
      when ready, `the-law` — pass with your own revision in force.
- [X] T010 [P] [US2] `docs/player/stage-3-the-craft.html`: block includes
      granting one tool beyond the basics (capability settings) or composing
      a skill file, then seeing the guardian actually use it; when ready, a
      stage-3 scenario where your granted tool contributes to the pass.
- [X] T011 [P] [US2] `docs/player/stage-4-the-stewardship.html`: block
      scripts a stewardship first session (full roster incl. capstone; set
      the pace, step back, steward through the guardian) and states plainly
      there is no exercise to pass — this stage is graduation.

**Checkpoint**: board AC #2 satisfiable; all four blocks live.

## Phase 5: Gates & polish

- [X] T012 Run the player-docs freshness probe directly:
      `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`
      → MUST exit 0: 13 fresh, 0 stale, 0 missing, 0 broken-ref (SC-003). If
      it names getting-started broken-ref, the T007 tag is malformed or
      mis-pinned — fix the tag, never the probe.
- [X] T013 Assert scope byte-identity: `git diff --stat origin/main` lists
      exactly the five pages + SKILL.md; `index.html` and the other 8 topic
      pages unchanged (FR-008, SC-003).
- [X] T014 Doctrine run: `go test ./...` green (SC-004 — no Go changes; a
      failure here is pre-existing drift to surface, not to fix in this
      branch).
- [X] T015 Wiki re-pin check — expected NO-OP: this branch touches no file
      any wiki note lists as a source (research.md R7), so no
      `/grounding-wiki:wiki-update` run belongs here. If the pr gate
      disagrees (`wiki-repin-missing`), produce the re-pin on this branch —
      the gate is the authority.
- [X] T016 From the worktree, run the pr gate:
      `node scripts/check-merge-drift.mjs pr` → exit 0 (SC-005;
      `player-docs-stale` probe = T012's command). `check-tui-design.mjs` is
      NOT applicable (no `docs/design/tui/` or `internal/tui/` changes).
- [x] T017 Open the one PR for TASK-153; merge merge-commit-only
      (`gh pr merge --merge` — in-branch pins die under squash). Post-merge:
      derived state only (board move via `backlog` CLI from repo root,
      spec-bridge sync, tasks.md ticks).

## Dependencies & Execution Order

### Phase Dependencies

- Phase 1 → Phase 2 → (Phase 3 ∥ Phase 4, both depend on Phase 2's contract;
  Phase 4's ask-link depends on T005 existing, so run T005 before or
  alongside T008–T011) → Phase 5 strictly last.

### User Story Dependencies

- US3 (contract) blocks US1 and US2 — it defines what they write.
- US2's blocks link to US1's step (T005), a content dependency only.

### Parallel Opportunities

- T008–T011 are fully parallel (four independent files).
- T005–T007 are one file — serial within Phase 3.

## Implementation Strategy

Single Sonnet `spec-implementer` slice (plan.md Constitution Check V):
content-only, one surface, no concurrency. MVP = Phases 1–3 + gates
(board AC #1 alone is a coherent, honest increment); Phase 4 completes AC #2
in the same PR — one task, one PR, always.

## Notes

- The sample ask MUST stay byte-identical to the default table value — a
  paraphrase is an independently asserted fact and dies at the next
  regeneration (research.md R1/R2).
- Preserve each page's slug, links, and the shared CSS block verbatim
  (SKILL.md skeleton contract); pages stay JS-free and self-contained.
- Do not "fix" the SKILL.md mapping rows for pages this task does not touch
  (bounded reconciliation, research.md R6 — corpus-wide de-churn is a parked
  watch item).
