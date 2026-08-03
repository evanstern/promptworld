# Playtest-1 findings sweep — sweep runbook (2026-07-30)

**You (the session reading this) are the ORCHESTRATOR** for the tasks below. Run each
through the host project's full PDLC — spec → link → worktree → delegated implementation →
PR → merge → re-ground — parallelizing within lanes, merging serially, treating merge
conflicts as routine. Direction is decided; do not re-litigate it: the five board cards
TASK-172..176 (each carries its playtest-1 evidence inline) win. Plan-of-record is the
board; this file carries only ordering, doctrine, and the log.

**Status:** done (2026-08-03) · operator sign-off on lanes: 2026-07-30 (lanes as drafted; TASK-173 measurement soak approved to start immediately)
<!-- Only the OPERATOR flips draft → signed-off (the author never pre-fills it). An
     executing session must refuse a runbook whose status it cannot verify. -->

## Read first (in this order)

1. The five board cards — `backlog task view TASK-172..176 --plain`. Each opens with a
   gist + use cases and carries the playtest-1 evidence (29 game-days, 1.01M events,
   world since migrated to v6 and retired) that produced it. There is no separate
   synthesis doc; the cards ARE the direction source.
2. Project gates: root `CLAUDE.md` (root-read-only + board-sync exception, wiki-in-PR
   lifecycle spec 069, merge-drift gates spec 051, claim protocol spec 065, model
   tiers `.specify/memory/constitution.md` Principle V).
3. `backlog task list --plain` — live state; other sessions move it while you work.
4. The task you're about to execute (`backlog task view TASK-<n> --plain`).

## State when this runbook was written (2026-07-30, origin/main ef115a6a)

- **Done already:** none of the scoped tasks; all five were carded 2026-07-30 from the
  playtest-1 log mining (board-sync commit ef115a6a).
- **In flight in other sessions (do not duplicate; expect their merges):**
  - **TASK-112** (guardian agentization, spec 102) — In Progress, live sibling session;
    worktree `.worktrees/task-112` is DIRTY. Never claim, rebase, or clean its branch
    or worktree. Its PR landing is Lane 2's merge gate.
  - **TASK-164** (charter-delta re-run) — In Progress, live sibling session; arm B soak
    running.
  - **TASK-158** (guardian missions) — held by the same live sibling session per the
    operator's dispatch ruling; ignore, non-blocking.
- **Paused — untouched:** no task carries the `paused` label. The operator's dispatch
  ruling (recorded verbatim): "TASK-112, TASK-164, and TASK-158 are held by a live
  sibling session — ignore, non-blocking. Never claim, rebase, or clean the task-112
  branch/worktree, the runbook-log worktree, or any world under ~/.promptworld/measure/
  (task-164-, task-112-soak-, and prior evidence worlds are preserved state), and never
  touch the playtest world."
- **Preserved state (never touch):** `.worktrees/task-112`, `.worktrees/runbook-log`
  (+ branch `runbook-log-board-sweep` — cleanup-eligible per the session gate but
  explicitly exempted by the ruling above), every world under `~/.promptworld/measure/`,
  and `~/.promptworld/worlds/playtest-1`.
- **Queued (this runbook's scope):** TASK-174, TASK-176 (Lane 1) · TASK-172, TASK-175
  (Lane 2) · TASK-173 (Lane 3). Spec numbers planned 103–107 in dispatch order —
  re-verify each with the claim gate at claim time; concurrent sessions take numbers.

## Execution lanes (dependency-ordered; parallelize within a lane)

Rule of thumb: DEVELOP in parallel, MERGE serially — tasks below share file footprints
(`internal/mind/` above all), so concurrent PRs will conflict; the lanes bound how bad
it gets. **Operator merge ruling (binding): TASK-172/175/173 merge only after
TASK-112's PR lands; TASK-174/176 may merge freely.**

**Lane 1 — start immediately, in parallel; free to merge (serially, smallest first):**
- **TASK-174 (Sonnet — routine single-subsystem robustness following TASK-58's
  established structured-outputs pattern)** — constrained/JSON-mode decoding for
  conversation outcomes on the local tier, or cloud fallback on parse failure.
  Footprint: `internal/mind/convo.go`, `internal/mind/parse.go`, `internal/llm/`
  request shape. Expected the smaller PR — merge first.
- **TASK-176 (Opus 4.8 — replay-determinism doctrine (spec 092/TASK-75) +
  architectural emission-shape change with cross-package consumers: sim reducer, TUI
  digest grammar, event-catalog wiki notes)** — coalesce ambient movement/needs/gru
  emission. Biggest slice; nothing else in this sweep fights for `internal/sim/`.
  Carries an operator checkpoint on the design fork (see checkpoints).

**Lane 2 — develop immediately in parallel worktrees; MERGE only after TASK-112's PR
lands (operator ruling). Merge order within lane: 172 → 175, reconcile between:**
- **TASK-172 (Opus 4.8 — cross-package: `internal/mind/consolidate.go` worker +
  `internal/llm` token-budget seam; failure-handling in async mind orchestration)** —
  truncation-aware consolidation retry + acceptance observability. Highest-priority
  fix in the sweep.
- **TASK-175 (Opus 4.8 — scheduling logic in `internal/mind` orchestration, named
  explicitly by the constitution's Opus rubric)** — gate planner scheduling on sleep
  state. Small diff, but it edits the mind driver (`internal/mind/mind.go`) that
  TASK-112's branch also touches — merging after 112 avoids a foreseeable conflict.
- After 112 lands: each Lane 2 branch merges `origin/main` IN (pin-carrying branches;
  never rebase), re-runs gates + freshness probe, then PRs serially.

**Lane 3 — tail; gated twice (measurement, then operator scope decision):**
- **TASK-173 (tier TBD at the checkpoint; Opus 4.8 if built — doctrine-adjacent
  belief/salience semantics on the spec-097 reconciliation seam)** — three steps:
  1. **Measure:** re-run the playtest-1 scenario (cold-dawn exercise, teaching,
     tutor preset, fresh seed) on current main in a NEW world — suggested name
     `task-173-measure-1` under `~/.promptworld/measure/` — and record on the task
     what spec 097 already absorbs (map-correction rate, share of chronicle entries
     narrating absence). Never reuse or touch preserved worlds or playtest-1.
  2. **Operator checkpoint:** present the measurement; operator decides build scope
     (full attribution seam / narrower narrator fix / drop).
  3. **Build** per the decision — spec → implement → PR; merges after 112 (ruling).
  The measurement (a multi-hour soak) may START as soon as Lane 1 dispatches — it
  costs no merge bandwidth and needs wall-clock time.

Record the model tier + rubric justification on each board task at dispatch
(one-way escalation only; escalations are operator checkpoints).

## Per-PR gates this project enforces (enumerated — implementers cannot miss these)

- **Merge-drift gate: present at `scripts/check-merge-drift.mjs`.** Mandatory at every
  choke point — four modes, probed at the sweep's precondition gate, invocations
  verbatim: `node scripts/check-merge-drift.mjs session` at sweep start (janitor +
  drift matrix), `node scripts/check-merge-drift.mjs claim --dir <NNN>-<slug>` before
  creating any new `specs/NNN-*` dir, `node scripts/check-merge-drift.mjs worktree
  [--spec <NNN>] [--task TASK-<n>]` before every `git worktree add`,
  `node scripts/check-merge-drift.mjs pr` from the worktree before every
  `gh pr create` AND after every history move (merge-in) — nonzero exit blocks.
  `wiki-repin-missing`, `player-docs-stale`, `player-docs-env-error` are blocking
  with no bypass flag.
- **Spec rigor:** full Spec Kit per task (specify → clarify where ambiguous → plan →
  tasks), linked to the board via `spec-bridge:link` BEFORE implementation.
- **Wiki-in-PR (spec 069):** the branch re-verifies + re-pins every wiki note whose
  pinned sources it touches; `docs/wiki/` changes regenerate `docs/player/`
  (`node .claude/skills/player-docs/scripts/check-freshness.mjs --check` is the probe).
  Likely-touched notes per task: 172 → nightly-consolidation, llm-* budget notes;
  174 → social-fabric-conversations; 175 → mind-driver-triggers; 176 → event-types-*
  catalogs, sim-state-* arms, tui-chronicle-feed; 173 → executor-perception-observation,
  nightly-consolidation/reconcile notes.
- **TUI design authority (spec 047):** any PR touching `internal/tui/` runs
  `node scripts/check-tui-design.mjs --changed` and amends `docs/design/tui/` in the
  same PR (plausible for TASK-176's digest-grammar surface).
- **Tests:** `go test ./...` green in the worktree before PR; respect existing -race
  budget conventions (TASK-169 context: chronicle -race suite is expensive).
- **Merge-commit-only:** `gh pr merge --merge` — never squash, never rebase
  (in-branch pins die otherwise; observed hazard on this repo).
- **Board hygiene:** board edits at root only via `backlog` CLI, commits scoped to the
  specific task files, pushed immediately (board-sync exception). tasks.md ticks and
  spec-bridge mirrors ride branches, never root commits.

## Concurrency & conflict doctrine

- **Hotspots:** `internal/mind/mind.go` (112, 175, 173); `internal/mind/parse.go`
  (172, 174 — cross-LANE overlap: reconcile Lane 2 branches after 174 merges);
  `internal/mind/consolidate.go` (172, 173); `internal/sim/state.go` + event
  vocabulary (176, 112); `docs/wiki/INDEX.md`, `docs/wiki/CAPSULES.md`, and the
  `event-types-*` notes (every PR's re-pins collide here — serial merges + reconcile
  absorb it); `docs/player/` regeneration; `specs/` numbering (103–107 planned).
- **Paused tasks are not live lanes:** none labeled `paused` today; the operator's
  dispatch ruling (state snapshot above) protects TASK-112/164/158's assets the same
  way — never claimed, rebased, or cleaned.
- Reconcile by what the branch carries: a **pin-carrying branch** (its own commits are
  referenced by re-pins it carries — wiki notes, design-reference pins) **merges
  `origin/main` in** — squash, rebase, and force-push all rewrite the branch's hashes
  and stale every carried pin, so its PR also lands as a merge commit, never a squash;
  a **pin-free branch rebases** — but on THIS repo rebases are forbidden everywhere
  (root CLAUDE.md), so every branch merges in. Take main's side for anything you
  didn't deliberately change.
- **Honest re-pins only — a merge-in never justifies a pin bump** (pin = merge commit
  empties the freshness probe's `git log <pin>..HEAD -- <sources>` range by
  construction). Route every pin the merge staled or conflicted through the
  wiki-update plan loop: read the main-side diff over the note's sources
  (`git diff <old-pin>..<merge-commit> -- <sources>`), classify **RE-PIN-ONLY**
  (provably prose-safe) vs **NEEDS-REVIEW** (re-verify and amend the note's prose
  against that diff BEFORE bumping). Never bump a pin without reading the diff it
  covers; the merge commit is the re-pin *target* once the note is verified, never
  the *justification*.
- After every history move (merge-in): re-run gates AND the freshness probe
  unconditionally — never gated on whether `docs/wiki/` changed; pins also reference
  design-reference files outside the wiki, so a wiki-untouched diff can still be stale.
- Two hotspot-heavy PRs never merge within one re-ground cycle without a reconcile
  between (merge-in per the pin rule).
- Conflicting with a sibling session's open PR → the smaller PR merges first.
- **Claim before work:** the FIRST commit of any task claims it — board card →
  In Progress AND the spec number's directory (a stub claims the number) — before any
  spec authoring or code. The claim rides the task branch, cut from `origin/main`
  (which does not contain the spec yet — the spec is authored on that branch, after
  the claim). Push immediately (`git push -u origin <branch>` on first commit, so
  in-flight work is auditable from any clone); never force-push a claim.
- **A rejected push means you lost the race:** fetch, re-read the board and `specs/`.
  If another session now holds that task or number, STOP the lane and surface it to
  the operator. Unrelated rejection (e.g. a board-notes push) with the task+number
  still free → fetch, merge `origin/main` into the claim branch, and re-push — a
  plain push: the merge-based remedy stays executable under the repo-wide rebase ban
  and never needs the force-push a claim forbids.
- The claim checks are mechanical: `claim --dir NNN-slug` before creating any new
  `specs/NNN-*` dir (blocks on a taken number); `worktree --spec NNN --task TASK-<n>`
  when cutting the worktree (warns if the card isn't claimed; accepts a spec dir
  already claimed by that same task).
- Verify a PR is merged (`gh api … --jq .merged`) before deleting its branch/worktree;
  never delete+recreate a closed PR's head.
- **Watching for TASK-112's landing (Lane 2's merge gate):** poll `gh pr list` /
  `git fetch` + session gate between merges. Watch only — never act on 112's branch,
  worktree, or PR.

## Operator checkpoints (do not proceed silently)

- **TASK-176 design fork:** emission-shape change vs offline compaction — surfaces at
  spec/plan time. If the chosen design changes event vocabulary or payload shapes, it
  has migration/format-version implications (spec 094 doctrine) — that is a one-way
  door; present it before tasks.md generation.
- **TASK-173 scope decision:** after the measurement run, present what spec 097
  already absorbs and get the build-scope ruling (full attribution seam / narrower
  fix / drop). Tier is assigned at this checkpoint.
- **TASK-112 PR lands** → confirm with the operator only if its merge conflicts with
  an already-open Lane 2 PR (whose-first is the smaller-PR rule; otherwise proceed
  per this runbook without asking).
- Tier escalations; lane amendments (amend this file, note why, tell the operator).

## Done means

- TASK-172, 174, 175, 176 Done on the board via their own merged PRs (spec-bridge:sync
  derived, never hand-set); TASK-173 either Done the same way or explicitly re-scoped/
  dropped at its checkpoint with the ruling recorded on the card.
- Lane 2 merges demonstrably after TASK-112's PR (execution log ordering proves it).
- All merge-drift gates green on main; wiki pins current; player-docs freshness probe
  passing; TUI design reference amended where `internal/tui/` changed.
- No sweep worktrees left (`git worktree list` clean of task-172..176 + this runbook's
  worktree after its final merge); preserved-state assets untouched.
- This file's execution log complete and status flipped to done.

## Execution log

| date | task | PR | merge | notes |
|------|------|----|-------|-------|
| 2026-07-30 | TASK-174 | — | b84c3130 | claimed: board In Progress (tier Sonnet), spec 103 stub on main, branch pushed |
| 2026-07-30 | TASK-176 | — | 4f56ae9a | claimed: board In Progress (tier Opus 4.8), spec 104 stub on main, branch pushed |
| 2026-07-30 | TASK-172 | — | 5e6a0862 | claimed: board In Progress (tier Opus 4.8), spec 105 stub on main, branch pushed |
| 2026-07-30 | TASK-175 | — | d178c905 | claimed: board In Progress (tier Opus 4.8), spec 106 stub on main, branch pushed |
| 2026-07-30 | TASK-173 | — | 36faaac1 | measurement soak started: task-173-measure-1 (cold-dawn, seed 46103, 16x, main@ef115a6a binary); scope checkpoint pending |
| 2026-07-30 | (all) | — | — | spec authoring dispatched: 103/105/106 full spec+plan+tasks, 104 spec.md-only (design-fork checkpoint) — planning tier |
| 2026-07-30 | TASK-174/172/175/176 | — | ac6ec32e | specs 103/105/106 (full) + 104 (spec.md, fork analysis) landed on main; bridge links + phase ACs planted (71e89bbc); bridge gate green |
| 2026-07-30 | TASK-176 | — | a41cb1d9 | OPERATOR CHECKPOINT RESOLVED: Arm A adopted (exact per-step sightings, interval+crossing needs, no format bump, old-world relief out of scope); plan/tasks authoring dispatched |
| 2026-07-30 | TASK-174/172/175 | — | — | implementation dispatched: 174→Sonnet, 172/175→Opus 4.8 (spec-implementer agents, parallel worktrees); Lane 2 PRs held pending TASK-112 |
| 2026-07-30 | TASK-176 | — | 6cecc7e3 | spec 104 plan/tasks landed (Arm A resolved, derived-progress engine design); phase ACs mirrored (16a1c649); implementation dispatched → Opus 4.8 |
| 2026-07-30 | TASK-175 | held | — | implementation complete @ d62fb2fe (suite+race green, sim diff empty, wiki re-pinned); open: T007 live soak (post-merge), T008 post-112 reconcile |
| 2026-07-30 | TASK-172 | held | — | implementation complete @ ada23d09 (race suite green, parse.go zero-diff, 21 notes re-pinned); open: T009 post-112 merge-in clause |
| 2026-07-30 | TASK-174 | #144 | merged | Lane 1 first merge: constrained decoding restored; T006 soak remains open (tasks.md tick corrected [~]→[ ] — bridge parser treats [~] as absent) |
| 2026-07-30 | TASK-176 | #145 | merged | Lane 1 complete: derived-progress engine, 7.7x ambient reduction measured; TASK-176 Done via sync; reconcile vs PR 144 carried honest RE-PIN-ONLY classifications |
| 2026-07-30 | (shared machine) | — | — | Sibling-session note recorded: TASK-164 arm B at 8x until ~2026-07-31 morning; arm B effective rate <8 ticks/s = first local-host starvation symptom. TASK-174 T006 soak deferred until arm B completes; 173 soak watcher extended to monitor arm B's rate and alert on degradation. Five daemons coexisting cleanly at time of note. |
| 2026-07-30 | TASK-172 | #147 | merged | Lane 2 first merge post-112: truncation ladder live; mind.go seam hand-merged (104 advance/sweep + 105 night-report); TASK-172 Done via sync |
| 2026-07-30 | TASK-175 | #148 | merged | Lane 2 complete: sleep gate + in-flight cancel live; union-restoration lesson — --ours dropped spec-102 steward/embedder wiki prose + a player paragraph, caught by audit and restored; open: T007 live soak |
| 2026-07-30 | TASK-173 | — | measurement | CHECKPOINT RESOLVED: build dropped — 097 absorbed the symptom (corrections −62%, zero absence storyline over 4.2 game-days, seed-matched); TASK-173 Done; soak daemon stopped, world preserved |
| 2026-07-30 | (handoff) | — | — | Background watchers reaped by the environment (3x) — no live watcher remains. RESUME CONDITION for the two open items: when TASK-164 arm B reaches tick 498187 (or its daemon exits), run TASK-174's T006 soak (queries: docs/design/evidence/task-174/queries.sql, ≥20 scenes, local gemma) and TASK-175's T007 soak (recipe: specs/106-sleep-gated-planning/soak.md, ≥3 game-days), record counts on the cards, tick the tasks.md boxes via a bookkeeping branch, spec-bridge sync both to Done, then flip this runbook's status to done. All five scoped tasks are otherwise merged; 172/173/176 are Done. |
| 2026-08-03 | TASK-174 | #144 | e04a69be | SWEEP COMPLETE. T006 soak closed the last open item. Soak B (qwen3.6:latest, spec 109 default): 92 founded scenes / 9.37 game-days, **0 outcome parse failures, 0 abandoned scenes** vs playtest-1's 22 / 21%. Soak A (gemma4:12b-mlx, 90 scenes) measured 10 parse failures / 3 scenes killed, which root-caused to Ollama's MLX engine silently discarding schema constraints — spec 103's code was correct and never reached the sampler. Spun out TASK-184/spec 109 (default → gguf model, merged PR #155), TASK-185 (daemon-start capability probe), TASK-186 (dead-path scheduling leak), and re-opened TASK-173 (absence storyline resurfaces past the 4-day window that justified dropping it — 969/972 corrections harvest-explained over 12 game-days). TASK-175 also closed Done on the same soak evidence (0 asleep rejections over 12.005 game-days, 192 planner calls saved). |
