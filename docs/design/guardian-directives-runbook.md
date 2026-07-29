# Guardian directives program — sweep runbook (2026-07-26)

**You (the session reading this) are the ORCHESTRATOR** for the tasks below. Run each
through the host project's full PDLC — spec → link → worktree → delegated implementation →
PR → merge → re-ground — parallelizing within lanes, merging serially, treating merge
conflicts as routine. Direction is decided; do not re-litigate it: the 2026-07-26
guardian-directives ideation decisions (recorded on TASK-157/TASK-158 cards),
`docs/design/learning-game-synthesis.md`, and reorient 2026-07-26 decision 9 (recorded on
TASK-118) win. Plan-of-record is the board; this file carries only ordering, doctrine,
and the log.

**Status:** superseded (2026-07-29) — never signed off. TASK-97/157/118 landed via the
faith+directives sweep (`docs/design/faith-directives-sweep-runbook.md`, done
2026-07-27); the remaining scope (TASK-112, TASK-158, TASK-81 tail) and this file's
still-live operator checkpoints (2, 3, 6) transfer to
`docs/design/board-sweep-2026-07-29-runbook.md`. Do not execute from this file.
<!-- Only the OPERATOR flips draft → signed-off (the author never pre-fills it). An
     executing session must refuse a runbook whose status it cannot verify. -->

## Read first (in this order)

1. TASK-157 and TASK-158 cards — the feature's decided shape, including the two firm
   operator decisions: **hard DIRECTIVE reflex rung between SURVIVAL and PREP**
   (interruptions by convos/hails/stimuli welcome; in-game workarounds first) and
   **EASY-mode default charter obedience** (refusals/personality are skinned-guardian
   data only).
2. `docs/design/learning-game-synthesis.md` — the three-lane initiative frame and the
   anti-self-grading guard both TASK-112 and TASK-158 must encode.
3. `docs/wiki/INDEX.md` → route to `guardian`, `guardian-orders`,
   `tool-registry-guardian-tools`, `reflex-prep-arbitration`, `mental-map-propagation`
   just-in-time. Never bulk-load the corpus.
4. `docs/design/tui/INDEX.md` — the UI authority gate rules for any PR touching
   `internal/tui/`.
5. `backlog task list --plain` — live state; other sessions move it while you work.
6. The task you're about to execute (`backlog task view TASK-<n> --plain`).

## State when this runbook was written (2026-07-26, origin/main = b087a36)

- **Done already:** guardian standing orders (spec 029), survival-autonomy *mechanism*
  (spec 059, PR #90) — but TASK-111 itself is still In Progress pending live evidence
  (AC3 live rejection rate = TASK-136, AC5 charter-delta = TASK-137). TASK-79 (dep of
  TASK-81) is Done.
- **In flight in other sessions (do not duplicate; expect their merges):** TASK-67
  (world-fork duel — TASK-118's ratified predecessor), TASK-151 + TASK-152 (known
  pairwise conflict across docs/wiki + docs/player surfaces — expect docs-surface
  rebases after whichever merges second), TASK-153, TASK-136, TASK-137.
- **Queued (this runbook's scope, execution order):** TASK-97 → TASK-157 → TASK-118 →
  TASK-112 → TASK-158; TASK-81 tail (droppable).

## Execution lanes (dependency-ordered; parallelize within a lane)

Every task in this program rewrites the same guardian-centric footprint
(`internal/guardian`, `internal/tool`, `internal/sim`), so the spine is **serial by
design** — the parallelism lives in pipelined spec authoring (specs live on main at
root, so authoring task N+1's spec while task N implements costs nothing).

**Lane A — the spine (one implementation in flight at a time; merges strictly serial):**

- **A1: TASK-97 (Sonnet — single-package compiler + contracts + fixtures; routine
  slice)** — target-addressing grammar (class+tile / region / id). CONTRACT-shaped:
  its published grammar unblocks TASK-157's designation addressing; TASK-157's spec
  may be authored as soon as 97's contract doc is committed, but 97 merges first.
  Escalate to Opus only if the grammar forces tool-registry schema restructuring.
- **A2: TASK-157 (Opus — cross-package: tool/sim/guardian/tui/mind; reflex-arbitration
  change is doctrine-adjacent)** — directives/designations substrate. The DIRECTIVE
  rung and the injection-door firewall are the review-critical surfaces. Touches
  `internal/tui` (map designation rendering) → TUI design gate applies.
- **A3: TASK-118 (Opus — reducer doctrine; charge economy is guardrail territory)** —
  faith-driven regen, spec-first. **Double-gated:** after TASK-157 (directive.*
  fulfillment events are the faith source) AND after TASK-67 merges (reorient
  2026-07-26 decision 9: duel before faith). If TASK-67 is unmerged when A2 merges →
  operator checkpoint 4.
- **A4: TASK-112 (Opus — stated on the card: cross-package metatron/mind/cognition/sim,
  doctrine-adjacent)** — guardian agentization. Ratified ordering: after TASK-118
  (agentization changes what earns faith). Depends on TASK-111, which is In Progress
  pending live evidence → operator checkpoint 2 before dispatch.
- **A5: TASK-158 (Opus — initiative-frame/doctrine-adjacent; missions extend the
  ambition lane's pre-authorization contract)** — missions loop on top of A2 + A4.

**Lane B — pipelined spec authoring (root-only, no worktree, no PR):** while lane A
implements task N, run specify → clarify → plan → tasks for task N+1 at root and
`spec-bridge:link` it. Spec docs commit directly to main (claim protocol applies:
`check-merge-drift.mjs claim --dir NNN-slug` before creating any `specs/NNN-*`).

**Lane C — tail (droppable):** **TASK-81 (Sonnet — narrow reducer arm + one tool;
escalate if toponymy grows into worldmap/state restructuring)** — canonization miracle,
after TASK-157 merges (reuses its durable-artifact entity discipline). May develop in
parallel with A3/A4 but merges only between spine merges, rebased, smallest-first.
Dropping it breaks nothing.

Record the model tier + rubric justification on each board task at dispatch
(one-way escalation only; escalations are operator checkpoints).

## Per-PR gates this project enforces (enumerated — implementers cannot miss these)

- **Merge-drift gate: present at `scripts/check-merge-drift.mjs`.** Mandatory at every
  choke point: `node scripts/check-merge-drift.mjs session` at sweep start and before
  each new task (janitor + drift matrix); `node scripts/check-merge-drift.mjs worktree
  --spec NNN --task TASK-<n>` before every `git worktree add`; `node
  scripts/check-merge-drift.mjs pr` from the worktree before every `gh pr create` AND
  after every rebase — nonzero exit blocks, no bypass flag.
- **Claim gate:** `node scripts/check-merge-drift.mjs claim --dir NNN-slug` before
  creating any `specs/NNN-*` directory; first commit claims card + spec number and
  pushes immediately; a rejected push is a stop-the-lane signal, never rebase-and-carry-on.
- **TUI design gate (spec 047):** any PR touching `internal/tui/` runs
  `node scripts/check-tui-design.mjs --changed` and amends `docs/design/tui/` in the
  same PR (re-verify + re-pin affected pages). Applies to A2 (map rendering) and
  likely A5 (mission surface).
- **Wiki-in-PR (spec 069):** the branch re-pins every wiki note whose sources it
  touches; `docs/player/` regenerated when `docs/wiki/` changes
  (`node .claude/skills/player-docs/scripts/check-freshness.mjs --check` is the probe).
  Grounding rides the PR — never a post-merge main commit.
- **Merge-commit-only:** `gh pr merge --merge` (a squash rewrites branch pins —
  observed hazard on this repo).
- **Spec rigor:** full Spec Kit per task, linked via `spec-bridge:link` BEFORE
  implementation; one task, one branch, one PR; subtasks are commits.
- **Board hygiene:** board/spec-bridge/tasks.md commands from repo root, never inside
  `.worktrees/`; git-add specific task files, never `backlog/` wholesale.

## Concurrency & conflict doctrine

- **Hotspots:** `internal/guardian/` (turn.go, toolcalls.go, orders.go — every spine
  task), `internal/tool/registry.go` + `derive.go` (A1/A2/C), `internal/sim/`
  (state.go, executor.go, policy.go, guardian.go — A2/A3/A4), and the grounding
  surfaces `docs/wiki/guardian*` + `docs/player/*` (every PR's re-pin rides them; the
  live TASK-151×TASK-152 conflict shows docs surfaces are the current hotspot).
- Rebase, never merge-commit into a task branch; take main's side for anything you
  didn't deliberately change; re-run tests AND gates after every rebase.
- Two hotspot-heavy PRs never merge within one re-ground cycle without a rebase between.
- Conflicting with a sibling session's open PR → the smaller PR merges first.
- Spec-number collisions: check `origin/main:specs/` (the claim gate does this
  mechanically) — numbers move fast here (080 was claimed by another session while this
  runbook was being authored); always take the claim gate's answer, never a remembered
  number.
- Verify a PR is merged (`gh api repos/{owner}/{repo}/pulls/<n> --jq .merged`) before
  deleting its branch/worktree; never delete+recreate a closed PR's head.

## Operator checkpoints (do not proceed silently)

1. **Sign-off on these lanes** (now, before any execution).
2. **TASK-112 dispatch while TASK-111 live evidence is pending** (TASK-136 rejection
   rate, TASK-137 charter delta): confirm 112 may start, or hold A4 until those land.
   Resurfaces when A3 merges.
3. **Deliberate-incompetence ceiling** (recorded open question on TASK-112: what the
   agent must never do well without a good charter; world-acting only, never tutor
   facts). Resurfaces at A4's specify/clarify.
4. **TASK-118's TASK-67 gate:** if TASK-67 (other session) hasn't merged when A2
   merges, decide — wait, resequence 118 spec-only, or relax decision 9 (operator-only).
5. **Failure-spiral floor posture** (TASK-118 AC4: scenario vs ambient floor) — spec-time
   decision at A3.
6. **Default-charter behavior change** (TASK-158 EASY-mode obedience edits the shipped
   default charter — behavior-affecting, TASK-73 precedent says eval-gated): decide the
   eval/gating approach at A5's specify.
7. Tier escalations; lane amendments (amend this file, note why, tell the operator).

## Done means

TASK-97, TASK-157, TASK-118, TASK-112, TASK-158 Done on the board, each via its own
merged PR; TASK-81 Done via merged PR or explicitly dropped with a note here; all gates
green on main (`check-merge-drift.mjs session` verdict pass, TUI design gate clean,
wiki pins current, player-docs freshness probe passing); no `.worktrees/task-*` left
from this sweep; execution log below complete; status above flipped to done.

## Execution log

| date | task | PR | merge | notes |
|------|------|----|-------|-------|
