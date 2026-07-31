# Board sweep 2026-07-29 — sweep runbook

**You (the session reading this) are the ORCHESTRATOR** for the tasks below. Run each
through the host project's full PDLC — spec → link → worktree → delegated implementation →
PR → merge → re-ground — parallelizing within lanes, merging serially, treating merge
conflicts as routine. Direction is decided; do not re-litigate it: the board cards
themselves (each carries its recorded operator decisions, evidence pointers, and
re-grounding notes), the guardian-directives ideation trail (TASK-112/158 cards +
`docs/design/learning-game-synthesis.md`), the TASK-163 evidence
(`docs/design/evidence/task-163/results.md`) for 164/166/167, and the 2026-07-25 team
review rulings (TASK-134/75/76 cards) win. Plan-of-record is the board; this file
carries only ordering, doctrine, and the log.

**Status:** signed-off — executing · operator sign-off on lanes: 2026-07-29
(in-session selection "Signed off — execute"; drafted by the orchestrator, the
sign-off is the authority). Rulings recorded at sign-off: **design sessions
(23/28/30) run autonomously — orchestrator authors the specs from the cards'
recorded decisions, operator ratifies at PR review**; **TASK-148 dropped entirely**
(card untouched this sweep; scope is 17 tasks); **TASK-164 stays after 166/167**;
TASK-65 exclusion confirmed.
<!-- Only the OPERATOR flips draft → signed-off (the author never pre-fills it). An
     executing session must refuse a runbook whose status it cannot verify. -->

**Supersession note:** `docs/design/guardian-directives-runbook.md` (draft, never
signed off) scoped TASK-97 → TASK-157 → TASK-118 → TASK-112 → TASK-158 (+ TASK-81
tail). 97/157/118 are Done via the faith+directives sweep (2026-07-27); the remaining
scope — TASK-112, TASK-158, TASK-81 — transfers HERE, along with its still-live
operator checkpoints (deliberate-incompetence ceiling; default-charter eval gating;
112-dispatch-wants-outcome-evidence, now fed by TASK-164). That file is amended to
point here.

## Read first (in this order)

1. The scoped task cards (`backlog task view TASK-<n> --plain`) — every card carries
   its own decisions, evidence pointers, and drift-audit pin updates; the card is the
   direction source.
2. `docs/design/learning-game-synthesis.md` (three-lane initiative frame,
   anti-self-grading guard — binds TASK-112/158) and
   `docs/design/evidence/task-163/results.md` (binds TASK-164/166/167).
3. `docs/wiki/CAPSULES.md` for whole-corpus orientation; full notes just-in-time per
   task (guardian*, sim-state-*, executor-*, reflex-*, memory/consolidation notes).
   Never bulk-load the corpus.
4. `docs/design/tui/INDEX.md` — UI authority gate rules for any PR touching
   `internal/tui/`.
5. `backlog task list --plain` — live state; other sessions move it while you work.

## State when this runbook was written (2026-07-29, origin/main = 1833d92)

- **Done already (deps all satisfied):** TASK-79, TASK-98, TASK-111, TASK-157,
  TASK-163 (PR #128), TASK-168; the guardian-directives program through TASK-118.
  No board task is In Progress.
- **In flight in other sessions (do not duplicate; expect their merges):** none known
  at authoring time — but the operator runs **playtest-1 (TASK-14, day 22 of 30)**
  live. See the playtest-protection doctrine below.
- **Paused — untouched:** TASK-18 (operator statement 2026-07-29: paused for re-eval;
  excluded from scope and from lane conflict analysis). TASK-14 excluded per operator
  (live proving run). Neither has branches or worktrees to protect at authoring time.
- **Excluded by its own card:** TASK-65 — "DELIBERATELY DEFERRED: do not start until
  the client picks a multiplayer direction." The sweep cannot manufacture that
  decision; the card stays on the board untouched. (Operator confirms at sign-off.)
- **Queued (this runbook's scope, 17 tasks):** 162, 165, 75, 134, 95, 166, 167,
  164, 112, 158, 80, 81, 99, 76, 23, 28, 30. (TASK-148 dropped at sign-off —
  card untouched.)
- Spec numbers: 087 is the highest claimed on origin/main; numbers move fast — always
  take the claim gate's answer, never a remembered number.

## Execution lanes (dependency-ordered; parallelize within a lane)

Rule of thumb: DEVELOP in parallel, MERGE serially — the sim-core and guardian
footprints below overlap heavily; the lanes bound how bad it gets. Spec authoring for
task N+1 pipelines while task N implements (specs are authored in the task's own
worktree after its claim — pipelining costs nothing).

**Lane 0 — process & hygiene (start immediately; merges first):**
- **TASK-162 (Sonnet — single-script tooling change with fixture tests; routine)** —
  merge-drift pr-gate probe fires on all pinned sources and after history moves.
  Lands FIRST so every later PR in this sweep runs the strengthened gate.
- **TASK-165 (Sonnet — doc reconciliation; routine)** — wiki size-budget debt
  (~24 findings). Lands immediately after 162: it churns `docs/wiki/` broadly, so it
  merges BEFORE the code lanes start carrying re-pins; any branch cut before its
  merge reconciles by merge-in. Not a lane blocker (the pr gate doesn't block on this
  class) but hygiene-first cuts sweep-long noise.
- ~~TASK-148~~ — dropped at sign-off (operator ruling 2026-07-29): the cross-repo
  praxis leg keeps it out of this sweep entirely; card untouched.

**Lane 1 — replay-doctrine spine (serial: 75 → 134 → 95):**
- **TASK-75 (Sonnet — docs/doctrine, minimal code)** — determinism scope note +
  reducer-constants hazard doctrine. Lands first: 134's machinery implements the
  doctrine this note states (the cards cross-reference each other).
- **TASK-134 (Opus — card-stated: replay/reducer doctrine, cross-package, migration
  machinery)** — event-log format_version + migration path; AC4 demonstrates the
  metatron.*→guardian.* rename end-to-end. Unblocks the TASK-121 interim shim
  removal (out of sweep scope; note on the card when it lands).
- **TASK-95 (Sonnet — extends the existing agent.build_failed pattern; escalate if
  replay expected-event churn spirals)** — loud failure for non-build goals. After
  134 (both churn event-types.md, replay expected sets, and reducer arms).

**Lane 2 — guardian program (serial spine; inherits the superseded runbook's order):**
- **TASK-166 (Sonnet — narrow door fix in `internal/sim/miracles.go` +
  `internal/guardian/turn.go`; escalate to Opus if the chosen mechanism rewrites
  reducer apply semantics)** — move-miracle target freshness. Spec-phase decision
  (AC1) leans (a) door-side name re-resolution with x/y advisory — the evidence doc's
  architectural reading; record the replay analysis on the card.
- **TASK-167 (Sonnet — guidance gloss / digest line; may close decision-only)** —
  carry-cap headroom for give_item. FR-011 constraint: never clamp at the door.
  Droppable if the recorded posture is teaching-door-only.
- **TASK-164 (no implementer — orchestrator-run eval per the measurement-run recipe;
  worlds under ~/.promptworld/measure/, never the playtest world)** — charter-delta
  outcome re-run. Ordered AFTER 166/167 merge so the re-run measures the charter
  surface, not residual tool fumbling (sequencing choice — operator may veto at
  sign-off and run it on current main instead). **OPERATOR CHECKPOINT: eval spend
  approved before dispatch (AC1).** Deliverable is a results doc + card updates via
  its own PR.
- **TASK-112 (Opus — card-stated: cross-package metatron/mind/cognition/sim,
  doctrine-adjacent)** — guardian agentization. Dispatch AFTER 164's evidence lands
  (the inherited dispatch checkpoint, now answerable with outcome-delta data).
  Guardrail set carries over unchanged; channel-split doctrine (AC6) is
  review-critical.
- **TASK-158 (Opus — initiative-frame doctrine; missions extend the ambition lane's
  pre-authorization contract)** — guardian missions, on top of 112 + 157. EASY-mode
  default charter edit is behavior-affecting → eval-gated per TASK-73 precedent
  (checkpoint at specify).

**Lane 3 — emergent lore (serial: 80 → 81; merges interleave with lane 2's gaps):**
- **TASK-80 (Opus — cross-package sim/mind epistemics; belief semantics
  doctrine-adjacent)** — perception of absence: grounded arrival observations feeding
  TASK-79's reinforcement seam. Working-window flood control (AC4) needs a live soak.
- **TASK-81 (Sonnet — narrow reducer arm + one guardian tool reusing TASK-157's
  durable-artifact entity machinery; escalate if toponymy forces worldmap/state
  restructuring)** — canonization miracle + named regions. After 80 (its
  arrival-discovery AC composes with 80's observation channel) and after 166 merges
  (shared `miracles.go` footprint).

**Lane 4 — private dreams (parallel develop; merge between lane-1/3 sim merges):**
- **TASK-99 (Opus — internal/mind consolidation orchestration; seeded-noise replay
  doctrine)** — consolidation clustering + habituation on TASK-98's recorded-vector
  infrastructure. Strictly per-agent by construction (AC1). Overlaps lane 3 on
  memory surfaces — never merge within one re-ground cycle of 80/81 without a
  reconcile.

**Lane 5 — design sessions (spec-only PRs; parallel-safe with everything):**
- **TASK-23 (planning model authors the spec)** — interactions v2. Least
  pre-decided card; posture (autonomous authoring vs interactive session) is a
  sign-off question.
- **TASK-28 (planning model)** — seasons + ambient temperature. Pre-session decisions
  recorded on the card (two seasons; diurnal curve; purpose is decision-3).
- **TASK-30 (planning model; after 28 merges — hard dependency)** — survival labor
  budget + calibration worksheet (AC2). Rewrite food-side baselines against spec-012
  pins per the card's drift audit.
- A design-session task's deliverable IS the ratified spec: one spec-only PR, linked
  via spec-bridge; the reviewer's decision is ratifying the design. Implementation is
  carded separately later (the TASK-25→spec-012→TASK-50 precedent).

**Lane 6 — tail (droppable):**
- **TASK-76 (Sonnet — mechanical accessor seam + determinism harness; routine)** —
  entity-lookup seam + store-error posture decision. Touches `state.go`/`terrain.go`
  + 7 `executor.go` call sites — merges LAST among sim-touching PRs so it rebases on
  everyone rather than everyone on it. Droppable without breaking anything.

Record the model tier + rubric justification on each board task at dispatch
(one-way escalation only; escalations are operator checkpoints).

## Per-PR gates this project enforces (enumerated — implementers cannot miss these)

- **Merge-drift gate: present at `scripts/check-merge-drift.mjs`.** Mandatory at every
  choke point, invocations verbatim: `node scripts/check-merge-drift.mjs session` at
  sweep start and before each new task (janitor + drift matrix);
  `node scripts/check-merge-drift.mjs claim --dir <NNN>-<slug>` before creating any
  new `specs/NNN-*` dir; `node scripts/check-merge-drift.mjs worktree --spec <NNN>
  --task TASK-<n>` before every `git worktree add`;
  `node scripts/check-merge-drift.mjs pr` from the worktree before every
  `gh pr create` AND after every history move — nonzero exit blocks, no bypass flag.
  (After TASK-162 lands, the pr gate's docs-stale probe is strictly stronger — expect
  it to fire on non-wiki pinned sources too.)
- **TUI design gate (spec 047):** any PR touching `internal/tui/` runs
  `node scripts/check-tui-design.mjs --changed` and amends `docs/design/tui/` in the
  same PR. Likely applies to: 80 (decision-trail visibility AC5), 81 (region names in
  chronicle/map), 112/158 (guardian surfaces), 95 (TUI digest for new failure events).
- **Wiki-in-PR (spec 069):** the branch re-pins every wiki note whose sources it
  touches; `docs/player/` regenerated when `docs/wiki/` changes
  (`node .claude/skills/player-docs/scripts/check-freshness.mjs --check` is the
  probe). Grounding rides the PR — never a post-merge main commit.
- **Merge-commit-only:** `gh pr merge --merge`; squash/rebase/force-push stale carried
  pins (observed hazards on this repo).
- **Spec rigor:** full Spec Kit per task, linked via `spec-bridge:link` BEFORE
  implementation; one task, one branch, one PR; subtasks are commits. Spec-only
  design-session tasks (lane 5) still link before their PR opens.
- **Replay determinism:** every sim-touching PR proves byte-identical replay (the
  existing harness); 134 additionally proves pre/post-migration identity; 99's noise
  (if adopted) must be rngAt-seeded.
- **Board hygiene:** board/spec-bridge/tasks.md commands from repo root, never inside
  `.worktrees/`; git-add specific task files, never `backlog/` wholesale; board-sync
  commits at root scoped to `backlog/` only (`git commit -F msgfile -- backlog/` —
  multi-line `-m` trips the root-guard parser).

## Concurrency & conflict doctrine

- **Hotspots:** `internal/sim/` core — `state.go`, `executor.go`, `agents.go`,
  `memory.go`, `miracles.go` (lanes 1/2/3/4/6 all land here); `internal/guardian/` +
  `internal/tool/` (166/167/81/112/158); `docs/wiki/` + `docs/player/` (every PR's
  re-pins, plus 165's broad churn — land 165 first); `event-types.md` + replay
  expected-event sets (134/95/80/81/99).
- **Playtest protection (iron rule for this sweep):** playtest-1 (TASK-14) is live at
  day 22/30. Nothing in this sweep restarts, migrates, probes, or writes to the
  playtest world or its daemon. Live probes (166 AC3, 80 AC4 soak, 81 AC4 demo, 164
  arms) run on separate seeded worlds (the measurement-run recipe;
  ~/.promptworld/measure/). 134's migration ships code only — no world on disk is
  migrated in-sweep except disposable test worlds.
- **Paused tasks are not live lanes:** TASK-18 (and TASK-14) are excluded and their
  state untouched — never claimed, rebased, or cleaned.
- Reconcile by what the branch carries: a **pin-carrying branch merges `origin/main`
  in** (squash/rebase/force-push all stale carried pins); a **pin-free branch
  rebases**. Take main's side for anything you didn't deliberately change.
- **Honest re-pins only — a merge-in never justifies a pin bump.** Route every pin
  the merge staled through the wiki-update classifier: read
  `git diff <old-pin>..<merge-commit> -- <sources>`, classify RE-PIN-ONLY vs
  NEEDS-REVIEW (re-verify prose BEFORE bumping). The merge commit is the re-pin
  *target*, never the *justification*.
- After every history move: re-run tests, gates, AND the player-docs freshness probe
  unconditionally — a wiki-untouched diff can still be stale.
- Two hotspot-heavy PRs never merge within one re-ground cycle without a reconcile
  between. Conflicting with a sibling session's open PR → smaller merges first.
- **Claim before work:** first commit = board card In Progress (board-sync commit at
  root, pushed immediately — the mutual-exclusion event) + spec stub dir on the task
  branch, pushed on first commit (`git push -u origin <branch>`). A rejected push is
  a stop-the-lane signal: fetch, re-read board + specs/; another session holds it →
  STOP and surface; unrelated rejection → `git pull --no-rebase` at root /
  merge-in on the branch, re-push. Never force-push a claim.
- Verify a PR is merged (`gh api repos/{owner}/{repo}/pulls/<n> --jq .merged`) before
  deleting its branch/worktree; never delete+recreate a closed PR's head.

## Operator checkpoints (do not proceed silently past)

1. ~~Sign-off on these lanes~~ — RESOLVED 2026-07-29 (see Status): lanes approved,
   TASK-65 excluded, TASK-148 dropped, design sessions autonomous with PR review,
   164 after 166/167.
2. **TASK-164 eval spend approval (AC1)** — resurfaces when lane 2 reaches it; spend
   estimate presented at that point.
3. **TASK-112 deliberate-incompetence ceiling** (inherited open question: what the
   agent must never do well without a good charter; world-acting only, never tutor
   facts) — resurfaces at 112's specify/clarify.
4. **TASK-158 default-charter edit eval-gating** (behavior-affecting; TASK-73
   precedent) — resurfaces at 158's specify.
5. **TASK-99 seeded-noise decision** (AC4: dream-distortion noise yes/no) — spec-time
   decision; checkpoint only if the spec wants it adopted (replay-doctrine surface).
6. Tier escalations; lane drops/reorders (amend this file, note why, tell the
   operator).

## Done means

All 17 queued tasks Done on the board via their own merged PRs — except:
TASK-167 optionally closed on a recorded no-change posture; lane-6 TASK-76 optionally
dropped with a note here. TASK-65/14/18 untouched. All gates green on main
(`check-merge-drift.mjs session` verdict pass; TUI design gate clean; wiki pins
current; player-docs freshness probe passing; grounding-wiki freshness exits 0 once
165 lands). No `.worktrees/` leftovers from this sweep. Execution log below complete;
status above flipped to done.

## Execution log

| date | task | PR | merge | notes |
|------|------|----|-------|-------|
| 2026-07-29 | TASK-162 | #131 | 259dd0f | lane 0; Sonnet; 31/31+10/10 tests; gate posture strengthened (probe on all pinned sources, history moves, tui-design blocking) |
| 2026-07-29 | TASK-166 | #133 | (merge sha: see PR) | lane 2; Sonnet; door-side name re-resolution; live probe landed raced move (59,20 surveyed vs 62,18 resolved); TASK-158 obedience observation recorded |
| 2026-07-29 | TASK-165 | #134 | febda4d | lane 0; Sonnet; 26 freshness findings -> 0; 14 tightened, 10 split, 0 exemptions; one RE-PIN-ONLY reconcile vs 166 |
| 2026-07-29 | TASK-167 | #135 | 4c22660 | lane 2; Sonnet; carry headroom in targeting digest + gloss; door byte-untouched (FR-011 pinned) |
| 2026-07-29 | TASK-134 | #136 | c9d30eb6 | lane 1; Opus; log format_version + translating migration + real guardian rename (13 types), TASK-121 shim deleted; byte-identity proven |
| 2026-07-29 | TASK-75 | #137 | 849e5ba6 | lane 1; Sonnet; per-log determinism corrected, reducer-constants doctrine + 13-site audit; zero behavior change |
| 2026-07-29 | TASK-95 | #138 | b9003230 | lane 1; Sonnet; agent.intent_failed for all non-build goals; reason taxonomy target-gone/contested/invalid; additive type |
| 2026-07-29 | TASK-99 | #139 | 804de10e | lane 4; Opus; private dream phase — single-store clustering (privacy perturbation-proven), recorded habituation/merge events, seeded zeroable jitter; demo evidence preserved |
| 2026-07-29 | TASK-80 | #141 | c2e8a89b | lane 3; Opus; place_observed arrival channel + belief reconciliation; soak caught flood + embed-spend defects pre-ship; follow-ups carded 169-171 |
| 2026-07-29 | TASK-76 | #142 | df015072 | lane 6 tail; Sonnet; EntityLookup accessor seam (26 sites + rot sweep), store-error fatal RATIFIED; AMENDMENT: tail merged before TASK-81 code existed (zero-conflict; 81 reconciles once) |
| 2026-07-29 | TASK-28 | #130 | 29a715517 | lane 5; design ratified by operator; seasons/ambient design (D1-D4) |
| 2026-07-29 | TASK-23 | #132 | 714f8a08e | lane 5; design ratified by operator; interactions v2 on tool substrate (OQ-1..5 recorded) |
| 2026-07-29 | TASK-30 | #140 | fc6ecf4f | lane 5; design ratified by operator; labor-budget invariants + calibration worksheet |
| 2026-07-30 | TASK-81 | #143 | 3fa33a6b | lane 3; Sonnet; canonization miracle (084 artifacts + 097 discovery, live Thornspire demo); charge shape 2-flat recorded |
| 2026-07-30 | TASK-112 | #146 | (merge on main) | lane 2; Opus; guardian agentization — steward class (de-themed per ruling), ceiling adopted + soak-proven, structural tutor split; reconciled across both sweeps |
| 2026-07-30 | TASK-158 | #149 | 4453b0fc | lane 2 final; Opus; guardian missions (084 artifacts, pre-authorization doctrine, EASY-mode default behind passing obedience eval); follow-up carded 177 |
