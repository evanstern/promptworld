# MVLS — Make Villagers Less Stoopid — sweep runbook (2026-07-25)

**You (the session reading this) are the ORCHESTRATOR** for the tasks below. Run each
through the host project's full PDLC — spec → link → worktree → delegated implementation →
PR → merge → re-ground — parallelizing within lanes, merging serially, treating merge
conflicts as routine. Direction is decided; do not re-litigate it: the TASK-101 spike
notes, `docs/design/control-surface-and-calibration.md` §3/§7, and the user decisions of
2026-07-24 recorded on each task win. Plan-of-record is the board; this file carries only
ordering, doctrine, and the log.

**Status:** signed-off · operator sign-off on lanes: 2026-07-25 (lanes as authored; TASK-89 live-world touch approved "go ahead anytime")
<!-- Only the OPERATOR flips draft → signed-off (the author never pre-fills it). An
     executing session must refuse a runbook whose status it cannot verify. -->

## Read first (in this order)

1. `backlog task view TASK-101 --plain` (the spike that produced 103/104/106) and
   `docs/design/control-surface-and-calibration.md` §3 (world-01 evidence), §7
   (decisions table that produced 107–111).
2. `specs/048-tuning-manifest/` — the shipped tuning.json promotion path (TASK-107,
   Done); TASK-108's thresholds ride it.
3. This project's CLAUDE.md (gates, tiers, worktree discipline) and
   `.specify/memory/constitution.md` (Principle V tier rubric).
4. `backlog task list --plain` — live state; other sessions move it while you work.
5. The task you're about to execute (`backlog task view TASK-<n> --plain`).

## State when this runbook was written (2026-07-25 ~18:30)

- **Done already:** TASK-107 (tuning manifest, PR #82 → 0cbf37b) — the keystone; wiki
  re-grounded (world-tuning note), player docs fresh.
- **In flight in other sessions (do not duplicate; expect their merges):** TASK-119,
  TASK-121, TASK-125, TASK-131. Specs 049–053 are claimed; next free spec number is 054 —
  re-verify against `origin/main:specs/` at each claim.
- **Queued (this runbook's scope, execution order):** TASK-108, TASK-110, TASK-109,
  TASK-111, TASK-89 (ops), TASK-106 (research) → TASK-103 → TASK-104 → TASK-122 (tail).
- **Operator instruction (2026-07-25):** other sweeps may claim one of these tasks. At
  each dispatch, re-check the task's status; if it is already In Progress or Done under
  another session, STOP that lane and surface it to the operator — do not re-implement
  or fight over it.

## Execution lanes (dependency-ordered; parallelize within a lane)

Rule of thumb: DEVELOP in parallel, MERGE serially — tasks below share file footprints,
so concurrent PRs will conflict; the lanes bound how bad it gets.

**Lane 1 — start immediately, in parallel (four worktrees + ops + research):**

- **TASK-108 (Opus 4.8 — doctrine-adjacent: survival-reflex behavior in the sim
  reducer, replay-affecting)** — build-fire cold reflex + raise refuel threshold.
  Thresholds ride tuning.json (extend `sim/tuning.go` dials only if the task's spec
  says a new dial is earned; the refuel default change is doctrine). Smallest instinct
  slice; merges FIRST in the sim hotspot so TASK-103 forks clean.
- **TASK-110 (Sonnet — single-mechanism clamps with complete file:line diagnosis on
  the task; routine)** — truncate-with-notice on expressive text fields, set_plan
  step-clamp, roster prune (collect_water, bathe). Publishes the registry text-cap
  contract BEFORE TASK-104 touches `tool/registry.go`.
- **TASK-109 (Opus 4.8 — internal/mind orchestration per rubric; diagnosis-first)** —
  find the pair-cooldown leak (planner talk_to vs encounter arming), then fix + novelty
  gate. The gate is a SHIM by decision — mark it as such in code and docs. Checkpoint
  below if diagnosis contradicts the assumed leak path.
- **TASK-111 (Opus 4.8 — doctrine-adjacent authority change in metatron turn logic)** —
  genesis watch orders, survival carve-out from the initiative frame,
  positions/passability digest in miracle guidance. Own package footprint
  (`internal/sim/metatron.go`, `metatron/turn.go`, `tool/derive.go`).
- **TASK-89 (ops, no PR — constitution trivial exemption: surgical config change,
  diagnosis pinned on task from spec 030 eval)** — point world-01 `llm.json` local tier
  at gemma4:12b-mlx. OPERATOR CHECKPOINT before touching the live world (restart).
  Evidence lands as task notes, not a PR.
- **TASK-106 (planning tier orchestrates; data-crunching delegable to Sonnet;
  research-only, no implementation)** — replay world-01 db, tune the thrash-detector
  candidate (W/K windows, need-progress clause), survey alternative metrics. Output: a
  research doc (docs/design/ or a spec-dir research.md) + a resurfaced checkpoint on
  whether/what to implement. Read-only footprint; runs alongside everything.

**Lane 1 merge order (serial, rebase + full gates between each):** TASK-110 → TASK-108 →
TASK-109 → TASK-111. (89/106 produce no PRs.)

**Lane 2 — after TASK-108 merges (sim reflex hotspot clear):**

- **TASK-103 (Opus 4.8 — doctrine-adjacent reflex/planner arbitration across sim+mind
  seams; the largest slice, gets the hotspot to itself)** — instinct yields to
  intelligence: no counter-scheduling against recent planner intents / danger-band
  needs; warmth in the day branch; prune prep behavior the planner owns.

**Lane 3 — after TASK-103 merges (intent semantics settled; TASK-110's registry
contract already on main):**

- **TASK-104 (Opus 4.8 — cross-package: sim executor + tool registry + mind prompt;
  doctrine-adjacent intent-completion semantics)** — needs-conditioned recovery:
  parameterized completion conditions (warm_up until_warmth≥N), not one-off verbs.

**Lane 4 — tail (droppable without breaking anything):**

- **TASK-122 (Sonnet — measurement + evidence doc)** — full-length (≥4 game-day)
  SC-007 flip-rate re-measure AFTER 103+104 are merged. OPERATOR CHECKPOINT to schedule
  the run (wall-clock + LLM cost on the rig).

Record the model tier + rubric justification on each board task at dispatch
(one-way escalation only; escalations are operator checkpoints).

## Per-PR gates this project enforces (enumerated — implementers cannot miss these)

- `go test ./...` AND `go vet ./...` green in the worktree — re-run BOTH after every
  rebase (sibling merges change tripwires; 048 hit the miracles_test rebase taxonomy).
- Any new tick-anchored int64 field on `sim.State` must be classified SHIFT/KEEP in the
  rebase taxonomy (`TestRebaseTaxonomyComplete`, `internal/sim/miracles_test.go`).
- `node scripts/check-tui-design.mjs --changed` before ANY PR touching `internal/tui/`;
  amend `docs/design/tui/` in the same PR if it does.
- Doctrine thresholds the task promotes to dials ride `sim/tuning.go` + the
  `sim.tuning_applied` event (spec 048 pattern) — never bare consts for anything
  world-tunable; snapshot-compat via pointer+omitempty, no format_version bump.
- Spec docs (`specs/NNN-*/`) and tasks.md ticks commit directly to main at ROOT; the
  task branch/PR carries code (+ same-PR design-doc amendments) only.
- `spec-bridge:link` (explicit spec dir + task id) BEFORE implementation;
  `spec-bridge:sync` after merge. Never rely on `.specify/feature.json` — concurrent
  sessions overwrite it; always pass explicit paths.
- Re-ground after every merge: `grounding-wiki:wiki-update` (sim/mind/metatron files
  are sources for many notes — expect NEEDS-REVIEW every merge), then
  `node .claude/skills/player-docs/scripts/check-freshness.mjs --check`.
- Board hygiene: `git add` specific `backlog/tasks/task-N*` files, never `backlog/`
  wholesale.

## Concurrency & conflict doctrine

- **Hotspots:** `internal/sim/policy.go` + `internal/sim/agents.go` (108↔103),
  `internal/tool/registry.go` (110↔104), `internal/mind/mind.go` + mind convo files
  (109↔104), `internal/sim/tuning.go` (108 + any sibling sweep touching dials),
  `.specify/feature.json` (every concurrent spec session), `backlog/` (every session).
- Rebase, never merge-commit into a task branch; take main's side for anything you
  didn't deliberately change; re-run gates after every rebase.
- Two hotspot-heavy PRs never merge within one re-ground cycle without a rebase between.
- Conflicting with a sibling session's open PR → the smaller PR merges first.
- Spec-number collisions: check `origin/main:specs/` before claiming an NNN (049–053
  already claimed at authoring time).
- Verify a PR is merged (`gh api repos/{owner}/{repo}/pulls/N --jq .merged`) before
  deleting its branch/worktree; never delete+recreate a closed PR's head.
- Wiki commits from concurrent sessions may sweep up in-progress note edits (happened
  during 048's re-ground) — always re-run the plan/freshness gates before claiming the
  corpus fresh; verify, don't assume commit messages match contents.

## Operator checkpoints (do not proceed silently)

1. **Lane sign-off** — this file's status flips to signed-off only by the operator.
2. **TASK-89 live-world touch** — editing world-01's llm.json requires a daemon restart
   of the operator's live world; confirm timing before acting.
3. **TASK-109 diagnosis review** — if the leak is NOT in mind-side arming (e.g. it's
   planner talk_to flood), the fix design changes; surface the diagnosis before
   implementing the damper.
4. **TASK-106 research review** — detector definition + whether to card/spec an
   implementation task; resurfaces when the research doc is written.
5. **TASK-122 run scheduling** — ≥4 game-days of wall-clock + LLM spend; confirm before
   launching the measurement.
6. **Any scoped task found In Progress/Done under another session at dispatch** — stop
   that lane, tell the operator, decide together (operator instruction, 2026-07-25).
7. Tier escalations; lane amendments (amend this file, note why, tell the operator).

## Done means

TASK-108, 110, 109, 111, 103, 104 all Done on the board via their own merged PRs with
all ACs checked; TASK-89 Done via ops evidence on the task (no PR); TASK-106 Done via a
committed research artifact + a recorded follow-on decision; TASK-122 Done via committed
evidence doc (post-checkpoint run) — or explicitly dropped as the tail lane with the
operator's ack. `go test ./...` green on main; wiki freshness gate 45+/45+ fresh;
player-docs freshness 13/13; `git worktree list` shows no `.worktrees/task-*` from this
sweep; this log complete and status flipped to done.

## Amendments

- 2026-07-25: spec numbers materialized as **057** (TASK-108), **058**
  (TASK-110), **059** (TASK-111), **061** (TASK-109; 060 was claimed by
  village-lens mid-sweep). **Collision note:** `059-grounded-feedback` (another
  session) and `059-metatron-survival-autonomy` (this sweep) were claimed
  near-simultaneously and both pushed — directory names are unique so all
  path-based machinery works; neither renumbered (a running implementer
  references ours). Surfaced to operator.
- 2026-07-25: TASK-109 diagnosis checkpoint cleared — leak proven to be the
  planner talk_to→hail founding path (97.8% of all world-01 scenes); operator
  chose the sim-side hail gate reusing the encounter_cooldown_ticks dial;
  novelty gate stays a marked mind-side SHIM.
- 2026-07-25: TASK-106 checkpoint cleared — operator accepted: detector
  W=4h/K=8 + need clause; **TASK-133 carded** (neglect detector, High — the
  shape that killed Oak); thrash-percept implementation deferred until after
  103/104 + TASK-122. TASK-106 Done (research-complete, no PR by design).
- 2026-07-25: TASK-89 grounding shift — world-01 llm.json already routed all
  traffic to gemma/cloud (cogito unused; stanza removed); world was
  format_version 3 → migrated v4 (backups kept); daemon restarted. AC2 (gist
  spot-check) pending scene accumulation.

## Execution log

| date | task | PR | merge | notes |
|------|------|----|-------|-------|
| 2026-07-25 | TASK-106 | — (research) | — | Done; artifact 7d79a0a; TASK-133 carded |
