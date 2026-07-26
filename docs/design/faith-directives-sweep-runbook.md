# Faith + directives sweep — sweep runbook (2026-07-26)

**You (the session reading this) are the ORCHESTRATOR** for the tasks below. Run each
through the host project's full PDLC — spec → link → worktree → delegated implementation →
PR → merge → re-ground — parallelizing within lanes, merging serially, treating merge
conflicts as routine. Direction is decided; do not re-litigate it: the reorient synthesis
`docs/design/reorient-2026-07-26-ui.md` (decision 9's lane order), the learning-game
synthesis (TASK-118's Wave-3 ratification), and the guardian-directives ideation of
2026-07-26 (operator-endorsed realignments on TASK-118/97; operator-firm directive
hardness on TASK-157) win. Plan-of-record is the board; this file carries only ordering,
doctrine, and the log.

**Status:** draft · operator sign-off on lanes: pending
<!-- Only the OPERATOR flips draft → signed-off (the author never pre-fills it). An
     executing session must refuse a runbook whose status it cannot verify. -->

## Read first (in this order)

1. `docs/design/reorient-2026-07-26-ui.md` (decision 9) + the guardian-directives
   ideation trail on TASK-157/158/118/97 (board cards, 2026-07-26).
2. `docs/design/reorient-2026-07-26-sweep-runbook.md` — the completed predecessor sweep
   (its execution log is the precedent for every doctrine call below).
3. `backlog task list --plain` — live state; other sessions move it while you work.
4. The task you're about to execute (`backlog task view TASK-<n> --plain`).

## State when this runbook was written (2026-07-26 ~21:30, after the reorient sweep)

- **Done already:** the full reorient 2026-07-26 sweep (TASK-154/149/150/142/67/152/151/
  153 via PRs #112–#116, #118–#120); TASK-111 (sibling session, operator-released hold);
  TASK-146 corpus-v2 adoption. Specs 001–**081** taken; **next free spec number is 082**
  (verify with the claim gate — the sibling session takes numbers).
- **In flight in the sibling session (do not duplicate; expect their merges):**
  TASK-159 (first-person harvest memory, spec 081, worktree `.worktrees/task-159`);
  TASK-156 wiki-budget-debt WIP at root (operator-directed to migrate into a worktree).
  TASK-136/137 remain In Progress from the earlier pause — **not scoped here**; their
  In Progress status is not a blocking finding.
- **Queued (this runbook's scope):** TASK-97 → TASK-157 → TASK-118 (Lane A, the anchor
  chain); TASK-133 (Lane B, parallel); TASK-17 (Lane C, deliberate tail).

## Execution lanes (dependency-ordered; parallelize within a lane)

Rule of thumb: DEVELOP in parallel, MERGE serially — tasks below share file footprints,
so concurrent PRs will conflict; the lanes bound how bad it gets.

**Lane A — the anchor chain (strictly serial; each consumes the previous contract):**
- **TASK-97 (Opus 4.8 — cross-package: bundle effect compiler + tool registry +
  metatron miracle paths; designs a grammar TWO consumers bind to)** — target-addressing
  grammar (class+tile / region / id) for bundle effects AND designation addressing,
  designed once for both. HIGH, on 157's critical path. Board card has no ACs — the
  spec seeds them; full Spec Kit regardless (non-trivial).
- **TASK-157 (Opus 4.8 — cross-package: guardian tools + sim designation/directive
  entities + reflex-ladder arbitration (doctrine-adjacent) + map render + decision
  context)** — survey read tool, event-sourced designations with reducer-stamped
  fulfillment predicates, hard directives (DIRECTIVE rung between SURVIVAL and PREP;
  interruption-friendly, in-game-workaround-first). Operator-firm decisions encoded,
  not re-opened. `directive.*` joins observableEventTypes.
- **TASK-118 (Opus 4.8 — reducer doctrine: faith as event-sourced state, regen as pure
  function; doctrine-adjacent by definition)** — spec-first; faith earned on directive
  fulfillment (`directive.*` from 157) plus any other sources the spec identifies;
  prophecy-verification rule; failure-spiral posture (AC4) decided IN the spec and
  surfaced to the operator; guardian-strip §4 dashed faith segment is the pre-specified
  UI contract; the spec-077 `first-faith-event` lesson rider lands here.

**Lane B — parallel with Lane A (independent failure-shape work):**
- **TASK-133 (Opus 4.8 — reducer/percept event + high-salience memory injection +
  world-01 log validation; cognition-adjacent)** — neglect detector (critical need,
  zero intents in class, window T) → deterministic percept + observation memory; alert
  via the shipped severity grammar (chronicle whole-line + map overlay), never a new
  channel. HIGH. Validated against Oak's death window (fires) and healthy windows
  (silent). Composition with TASK-111 survival watches considered in the spec.

**Lane C — deliberate tail (merges LAST, sweeps the others' new payloads):**
- **TASK-17 (Opus 4.8 — repo-wide payload migration: AgentRef {id,name} across every
  agent-referencing emitter + mechanical enforcement + back-compat replay)** — merges
  last ON PURPOSE: its exhaustive payload sweep then covers the new `directive.*`,
  faith, and neglect-percept payloads in one pass instead of forcing three in-flight
  lanes to rebase over a payload-shape change. Carries the operator-placed REVERSE-JUMP
  rider (strip glyph / roster row → camera center), which ships with a mouse-parity
  oracle entry per the TASK-154 gate.

Record the model tier + rubric justification on each board task at dispatch
(one-way escalation only; escalations are operator checkpoints). All five slices are
Opus-tier under the constitution's rubric (every one is cross-package and/or
doctrine-adjacent); there is no Sonnet slice in this sweep.

## Per-PR gates this project enforces (enumerated — implementers cannot miss these)

- **Merge-drift gate: present at `scripts/check-merge-drift.mjs`.** Mandatory at every
  choke point, hook-enforced: `session` at each task start; `claim --dir NNN-slug`
  before creating any `specs/NNN-*` dir; `worktree --spec NNN --task TASK-<n>` before
  every `git worktree add`; `pr` from the worktree before every `gh pr create` AND
  after every history move — nonzero exit blocks, no bypass flag.
- **TUI authority gate (spec 047):** any PR touching `internal/tui/` runs
  `node scripts/check-tui-design.mjs --changed` and amends `docs/design/tui/` same-PR —
  including the spec-075 semantic-cells check (no `unbuilt (wave` on shipped pages) and
  the spec-073 mouse-parity sweep (new control-table mouse claims need live-dispatch
  oracle entries). Bites 157 (map designations), 133 (map overlay), 17 (chronicle,
  reverse-jump).
- **Wiki-in-PR lifecycle (spec 069):** in-branch re-pins for every note whose pinned
  sources the branch touches; `docs/player/` regenerated when wiki changes; run the
  freshness probe `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`
  DIRECTLY after every history move (player pages also pin `docs/design/tui/*` files —
  the pr gate's probe alone misses that class; predecessor-sweep lesson).
- **Tests:** `go test ./...` green, re-run after every reconcile. Every new event type
  needs its digest row (`TestCatalogSweep`) — bites ALL FIVE tasks. Corpus body budget
  (8000 chars) is FAIL-mode since v2 adoption — split notes per the child-note pattern
  rather than exempting.
- **Merge-commit-only:** `gh pr merge --merge`, never squash.
- **Reconcile doctrine (operator-RATIFIED 2026-07-26, supersedes the template's
  rebase line):** a task branch that carries in-branch pins reconciles with a moved
  origin/main by **merging origin/main into the branch** and re-pinning conflicted pins
  to the merge commit — never rebase/force-push (rewritten hashes stale every pin;
  observed across the predecessor sweep; skill-level fix carded as praxis TASK-57).
  A pin-free branch may still rebase.
- **Spec rigor:** full Spec Kit + `spec-bridge:link` BEFORE implementation for every
  task. Claim-before-work (spec 065) verbatim: first commit = card In Progress + spec
  stub, pushed immediately; rejected push = stop-the-lane signal; specs/board
  bookkeeping to main at root, task branch carries code only.

## Concurrency & conflict doctrine

- **Hotspots:** `internal/sim/agents.go` + event payload structs (17 vs everyone —
  bounded by Lane C merging last); `internal/tool`/`internal/metatron` guardian
  rosters (97, 157); `internal/sim/executor.go` reflex ladder (157's DIRECTIVE rung vs
  133's detector — different functions, same file); `internal/tui/grammar.go`+
  `digest.go` digest rows (all five); the wiki guardian/reflex/event-catalog notes and
  regenerated `docs/player/` (every PR); `backlog/tasks/*` (scoped adds only).
- The **sibling session** is live (TASK-159 now; TASK-156 wiki splits pending; the
  guardian-directives cards are its ideation): before claiming TASK-157, re-read the
  board — if the sibling session has claimed it, STOP the lane and surface to the
  operator rather than racing the claim.
- Rejected push = stop-the-lane signal (spec 065); verify merges
  (`gh api … --jq .merged`) before deleting branches; never delete+recreate a closed
  PR's head.
- Two hotspot-heavy PRs never merge within one re-ground cycle without a reconcile
  between; smaller PR merges first on sibling conflicts.

## Operator checkpoints (do not proceed silently)

- **Lane sign-off on THIS file** (status flips draft → signed-off) — the lanes above
  are the author's derivation, not operator-specified verbatim.
- **TASK-157 claim collision** — if the sibling session holds or claims it, stop and
  ask (its cards; this runbook schedules them on the operator's follow-on directive).
- **TASK-118 AC4 failure-spiral posture** — decided in the spec with the Hades
  God-Mode grounding; surfaced in the sweep report (scenario-authentic spiral vs
  ambient floor).
- **TASK-17 back-compat story** — decided in the spec (reducer accepts named and
  unnamed payloads; renderers degrade); surfaced, not re-asked.
- **Fork-duel phase 2 (HTML retelling)** — recorded follow-on with NO card; carding it
  is an operator decision, listed here so it isn't lost.
- Tier escalations (none possible — all Opus already); lane amendments (amend this
  file, note why, tell the operator).

## Done means

TASK-97, 157, 118, 133, 17 all Done on the board, each via its own merged PR (merge
commits); `go test ./...`, `check-tui-design.mjs` (bare + `--changed`),
`TestCatalogSweep`, `TestMouseParitySweep`, player-docs freshness, and
`check-merge-drift.mjs session` green on main; wiki pins current with all notes under
the v2 body budget; no sweep worktrees; this file's log complete and status flipped to
done. TASK-136/137/156/159 untouched by this sweep.

## Execution log

| date | task | PR | merge | notes |
|------|------|----|-------|-------|
